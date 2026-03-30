// Package telegram provides the delivery layer for the Telegram Bot API.
// It maps incoming Telegram commands to the underlying core services.
package telegram

import (
	"context"
	"log/slog"
	"strings"

	"github.com/Golang-Venezuela/adan-bot/internal/core/ports"
	"github.com/Golang-Venezuela/adan-bot/internal/infra/logger"
	tele "gopkg.in/telebot.v4"
)

// RegisterHandlers configures the bot routing by injecting the core BotService.
// It sets up the primary endpoints (/help, /hola, /status) and maps them to their respective
// service handlers, enabling loose coupling between the delivery mechanism and business logic.
func RegisterHandlers(bot *tele.Bot, svc ports.BotService) {
	bot.Handle("/help", func(c tele.Context) error {
		ctx := context.Background()
		return c.Send(svc.HandleStartHelp(ctx))
	})

	bot.Handle("/hola", func(c tele.Context) error {
		ctx := context.Background()
		userID := c.Sender().ID
		username := c.Sender().Username
		firstName := c.Sender().FirstName
		lastName := c.Sender().LastName

		slog.Debug("Executing /hola", slog.String("user_id", logger.ObfuscateID(userID)))

		msg, err := svc.HandleHola(ctx, userID, username, firstName, lastName)
		if err != nil {
			slog.Error("Error handled in HandleHola", slog.Any("error", err))
		}

		return c.Send(msg)
	})

	bot.Handle("/status", func(c tele.Context) error {
		ctx := context.Background()
		return c.Send(svc.HandleStatus(ctx))
	})

	// Fallback handler for unmapped textual commands.
	bot.Handle(tele.OnText, func(c tele.Context) error {
		if strings.HasPrefix(c.Text(), "/") {
			slog.Warn("Unknown command received", slog.String("text", c.Text()))
			return c.Send("I don't know that command")
		}
		return nil
	})
}
