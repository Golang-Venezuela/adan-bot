// Package main provides the entry point for the Adan Bot application.
// It initializes the bot in long-polling mode, configures logging and middlewares,
// and defines the core handlers for user commands.
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/Golang-Venezuela/adan-bot/internal/adapters/delivery/telegram"
	"github.com/Golang-Venezuela/adan-bot/internal/adapters/repository"
	sqlite "github.com/Golang-Venezuela/adan-bot/internal/adapters/repository/sqlite"
	"github.com/Golang-Venezuela/adan-bot/internal/core/ports"
	"github.com/Golang-Venezuela/adan-bot/internal/core/services"
	"github.com/Golang-Venezuela/adan-bot/internal/infra/config"
	"github.com/Golang-Venezuela/adan-bot/internal/infra/logger"
	"github.com/Golang-Venezuela/adan-bot/internal/infra/profiling"

	tele "gopkg.in/telebot.v4"
)

var (
	// startTime is used to calculate the bot's uptime.
	startTime = time.Now()

	// ErrMissingToken is returned when the TELEGRAM_API_TOKEN environment variable is absent or empty.
	ErrMissingToken = errors.New("missing Telegram API token")
)

// Main handles the core initialization and execution loop of the bot.
// It retrieves the API token, configures bot preferences, sets up observability middlewares,
// registers all primary command handlers, and starts the long-polling process.
//
//nolint:cyclop,gocognit
func Main() error {
	APITOKEN := strings.Trim(config.Getenv("TELEGRAM_API_TOKEN", ""), `"`)
	if APITOKEN == "" {
		return ErrMissingToken
	}

	slog.Info("Bot configuration loaded", slog.String("token", logger.Obfuscate(APITOKEN)))

	// Configure fundamental bot settings including the polling timeout and error callbacks.
	pref := tele.Settings{
		Token:  APITOKEN,
		Poller: &tele.LongPoller{Timeout: 60 * time.Second},
		OnError: func(err error, c tele.Context) {
			if c != nil && c.Chat() != nil {
				slog.Error("Cannot send message", slog.String("username", logger.Obfuscate(c.Chat().Username)))
			} else {
				slog.Error("Telegram error", slog.Any("error", err))
			}
		},
	}

	bot, err := tele.NewBot(pref)
	if err != nil {
		return fmt.Errorf("cannot connect to Telegram API: %w", err)
	}

	slog.Info("Authorized", slog.String("account", logger.Obfuscate(bot.Me.Username)))

	// Initialize Database dependencies
	dbUrl := config.Getenv("TURSO_DB_URL", "file:local.db")
	userRepo, err := sqlite.NewUserRepository(dbUrl)
	if err != nil {
		return fmt.Errorf("cannot initialize database: %w", err)
	}
	botSvc := services.NewBotService(userRepo)
	telegram.RegisterHandlers(bot, botSvc)

	// Initialize Moderation dependencies
	modRepo := repository.NewMemoryModerationRepo()
	modSvc := services.NewModerationService(modRepo)
	telegram.RegisterModerationHandlers(bot, modSvc)

	// Start background birthday scheduler
	startBirthdayScheduler(bot, botSvc)

	// Inject a global middleware to intercept and log incoming messages for tracing and observability.
	bot.Use(func(next tele.HandlerFunc) tele.HandlerFunc {
		return func(c tele.Context) error {
			if c.Message() != nil {
				slog.Info("Received message",
					slog.String("user_id", logger.ObfuscateID(c.Sender().ID)),
					slog.String("username", logger.Obfuscate(c.Sender().Username)),
					slog.String("text", c.Text()),
				)
			}
			return next(c)
		}
	})

	// Setup basic handlers for standard bot commands (e.g., /ayuda, /hola, /estatus).
	bot.Handle("/ayuda", func(c tele.Context) error {
		slog.Info("Executing /ayuda command", slog.String("user_id", logger.ObfuscateID(c.Sender().ID)))

		menu := &tele.ReplyMarkup{ResizeKeyboard: true}
		btnHola := menu.Text("Bienvenida")
		btnStatus := menu.Text("Estatus")

		menu.Reply(
			menu.Row(btnHola, btnStatus),
		)

		return c.Send("¡Hola! Soy Adan Bot. Selecciona una opción del menú de abajo:", menu)
	})

	bot.Handle("Bienvenida", func(c tele.Context) error {
		slog.Info("Executing /hola command", slog.String("user_id", logger.ObfuscateID(c.Sender().ID)))

		displayName := c.Sender().FirstName
		if displayName == "" {
			displayName = c.Sender().Username
			if displayName != "" {
				displayName = "@" + displayName
			}
		}

		// send sticker with welcome greeting
		sticker := &tele.Sticker{File: tele.FromURL("https://raw.githubusercontent.com/TelegramBots/book/master/src/docs/sticker-fred.webp")}
		_ = c.Send(sticker)

		ctx := context.Background()
		msg := modSvc.GetWelcomeMessage(ctx, displayName)
		return c.Send(msg, tele.ModeHTML)
	})

	bot.Handle("Estatus", func(c tele.Context) error {
		slog.Info("Executing /estatus command", slog.String("user_id", logger.ObfuscateID(c.Sender().ID)))

		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		uptime := time.Since(startTime).Round(time.Second)
		goroutines := runtime.NumGoroutine()
		memUsedMB := m.Alloc / 1024 / 1024

		msg := fmt.Sprintf("✅ <b>Servicio Online</b>\n\n"+
			"⏱ <b>Uptime:</b> %s\n"+
			"🧵 <b>Goroutines:</b> %d\n"+
			"💾 <b>Memoria:</b> %d MB", uptime, goroutines, memUsedMB)

		return c.Send(msg, tele.ModeHTML)
	})

	// Fallback handler for any text payload that resembles an unsupported command.
	bot.Handle(tele.OnText, func(c tele.Context) error {
		if strings.HasPrefix(c.Text(), "/") {
			slog.Warn("Unknown command received", slog.String("text", c.Text()), slog.String("user_id", logger.ObfuscateID(c.Sender().ID)))
			return c.Send("I don't know that command")
		}
		return nil
	})

	slog.Info("Bot starting", slog.String("username", bot.Me.Username))
	bot.Start()

	return nil
}

// main is the application entry point. It wraps Main with execution profiling.
func main() {
	if err := profiling.Profile(Main); err != nil {
		slog.Error("Fatal runtime error exiting", slog.Any("error", err))
		os.Exit(1)
	}
}

// startBirthdayScheduler runs a background worker that checks birthdays daily.
func startBirthdayScheduler(bot *tele.Bot, svc ports.BotService) {
	chatIDStr := config.Getenv("TELEGRAM_GROUP_CHAT_ID", "")
	if chatIDStr == "" {
		slog.Warn("TELEGRAM_GROUP_CHAT_ID is not configured; birthday announcements will not be sent")
		return
	}

	var chatID int64
	if _, err := fmt.Sscan(chatIDStr, &chatID); err != nil {
		slog.Error("Invalid TELEGRAM_GROUP_CHAT_ID format", slog.String("value", chatIDStr), slog.Any("error", err))
		return
	}

	go func() {
		loc, err := time.LoadLocation("America/Caracas")
		if err != nil {
			loc = time.UTC
		}

		for {
			now := time.Now().In(loc)
			nextRun := time.Date(now.Year(), now.Month(), now.Day(), 8, 0, 0, 0, loc)
			if now.After(nextRun) {
				nextRun = nextRun.Add(24 * time.Hour)
			}

			duration := time.Until(nextRun)
			slog.Info("Birthday scheduler scheduled next check", slog.Time("next_run", nextRun), slog.Duration("wait_duration", duration))

			time.Sleep(duration)

			ctx := context.Background()
			msg, err := svc.HandleTodayBirthdays(ctx)
			if err != nil {
				slog.Error("Error checking today's birthdays in scheduler", slog.Any("error", err))
				continue
			}

			if msg != "" {
				slog.Info("Sending daily birthday congratulations message")
				_, err = bot.Send(&tele.Chat{ID: chatID}, msg, tele.ModeHTML)
				if err != nil {
					slog.Error("Failed to send birthday greetings message", slog.Any("error", err))
				}
			}
		}
	}()
}
