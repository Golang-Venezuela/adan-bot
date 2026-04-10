// Package telegram provides the delivery layer for the Telegram Bot API.
// It maps incoming Telegram commands to the underlying core services.
package telegram

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/Golang-Venezuela/adan-bot/internal/core/ports"
	"github.com/Golang-Venezuela/adan-bot/internal/infra/logger"
	tele "gopkg.in/telebot.v4"
)

// RegisterModerationHandlers configures the bot routing for community moderation features.
// It handles events such as users joining, issuance of warnings, and toggle of read-only mode.
func RegisterModerationHandlers(bot *tele.Bot, svc ports.ModerationService) {
	// 1. Welcome Message Handler: Triggered when a new user joins a chat.
	bot.Handle(tele.OnUserJoined, func(c tele.Context) error {
		ctx := context.Background()
		if c.Chat() == nil || c.Sender() == nil {
			return nil
		}

		displayName := c.Sender().FirstName
		if displayName == "" {
			displayName = c.Sender().Username
			if displayName != "" {
				displayName = "@" + displayName
			}
		}

		slog.Info("New user joined", slog.String("user_id", logger.ObfuscateID(c.Sender().ID)))

		// Send a generic greeting sticker (owners can replace the URL with a specific FileID).
		sticker := &tele.Sticker{File: tele.FromURL("https://raw.githubusercontent.com/TelegramBots/book/master/src/docs/sticker-fred.webp")}
		_ = c.Send(sticker)

		welcomeMsg := svc.GetWelcomeMessage(ctx, displayName)
		return c.Send(welcomeMsg, tele.ModeHTML)
	})

	// 2. Moderation Command: /warn
	// Issues a warning to a user by replying to their message.
	// At 3 cumulative warnings, the user is automatically banned (kicked) from the chat.
	bot.Handle("/warn", func(c tele.Context) error {
		ctx := context.Background()
		if c.Chat() == nil || c.Sender() == nil {
			return nil
		}

		// Requirement: The command must be a reply to the target user's message.
		if !c.Message().IsReply() {
			return c.Reply("You must reply to the message of the user you wish to warn.")
		}

		targetUserID := c.Message().ReplyTo.Sender.ID
		adminID := c.Sender().ID
		chatID := c.Chat().ID

		// Parse the reason from the command arguments.
		reason := strings.TrimSpace(strings.TrimPrefix(c.Text(), "/warn"))
		if reason == "" {
			reason = "Violation of community rules."
		}

		count, err := svc.IssueWarning(ctx, chatID, targetUserID, adminID, reason)
		if err != nil {
			slog.Error("Failed to issue warning", slog.Any("error", err))
			return c.Reply("An error occurred while processing the warning.")
		}

		msg := "The user has received a warning."
		if count >= 3 {
			// Reach threshold (3 warns) -> Execute Ban/Kick.
			member := &tele.ChatMember{User: c.Message().ReplyTo.Sender}
			if err := bot.Ban(c.Chat(), member); err != nil {
				slog.Error("Failed to kick user", slog.Any("error", err))
				msg += "\n(Could not kick the user. Do I have administrator permissions?)"
			} else {
				msg += "\nUser kicked after reaching 3 warnings."
			}
		} else {
			msg += fmt.Sprintf("\nAccumulated warnings: %d/3", count)
		}

		return c.Send(msg)
	})

	// 3. Moderation Command: /romode
	// Toggles Read-Only mode for the chat. When enabled, all non-admin messages are deleted.
	bot.Handle("/romode", func(c tele.Context) error {
		ctx := context.Background()
		if c.Chat() == nil || c.Sender() == nil {
			return nil
		}

		args := strings.TrimSpace(strings.TrimPrefix(c.Text(), "/romode"))
		chatID := c.Chat().ID
		adminID := c.Sender().ID

		var enable bool
		switch args {
		case "on":
			enable = true
		case "off":
			enable = false
		default:
			return c.Reply("Usage: /romode on | off")
		}

		if err := svc.ToggleRoMode(ctx, chatID, adminID, enable); err != nil {
			return c.Reply("Could not change the RoMode state.")
		}

		var msg string
		if enable {
			msg = "Read-Only mode (RoMode) has been ACTIVATED."
		} else {
			msg = "Read-Only mode (RoMode) has been DEACTIVATED."
		}

		return c.Send(msg)
	})

	// 4. Moderation Middleware: Intercepts all textual updates to enforce RoMode and Anti-Spam.
	bot.Use(func(next tele.HandlerFunc) tele.HandlerFunc {
		return func(c tele.Context) error {
			if c.Message() == nil || c.Chat() == nil || c.Sender() == nil || c.Text() == "" {
				return next(c)
			}

			// Bypass moderation for private (1-on-1) chats.
			if c.Chat().Type == tele.ChatPrivate {
				return next(c)
			}

			ctx := context.Background()

			// RoMode Enforcement: Delete any message if mode is active.
			if svc.IsRoModeActive(ctx, c.Chat().ID) {
				_ = c.Delete()
				return nil
			}

			// Anti-Spam Enforcement: Block links from recent members (e.g., joined < 24h ago).
			// Note: Retrieve member join date when supported:
			// member, err := bot.ChatMemberOf(c.Chat(), c.Sender())
			// if err == nil && member != nil {
			// 	// Note: Precise join time might vary based on chat member scope.
			// }
			var joinDate time.Time

			isSpam, _ := svc.CheckAntiSpam(ctx, c.Chat().ID, c.Sender().ID, c.Text(), joinDate)
			if isSpam {
				_ = c.Delete()
				return c.Send("Message deleted: URLs are not allowed from recent members.")
			}

			return next(c)
		}
	})
}
