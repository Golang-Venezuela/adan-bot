// Package main provides the entry point for the Adan Bot application.
// It initializes the bot in long-polling mode, configures logging and middlewares,
// and defines the core handlers for user commands.
package main

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/Golang-Venezuela/adan-bot/internal/infra/config"
	"github.com/Golang-Venezuela/adan-bot/internal/infra/logger"
	"github.com/Golang-Venezuela/adan-bot/internal/infra/profiling"

	tele "gopkg.in/telebot.v4"
)

// Predefined errors used within the bot initialization.
var (
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

	// Setup basic handlers for standard bot commands (e.g., /help, /hola, /status).
	bot.Handle("/help", func(c tele.Context) error {
		slog.Info("Executing /help command", slog.String("user_id", logger.ObfuscateID(c.Sender().ID)))
		return c.Send("/hola and /status.")
	})

	bot.Handle("/hola", func(c tele.Context) error {
		slog.Info("Executing /hola command", slog.String("user_id", logger.ObfuscateID(c.Sender().ID)))
		msg := "Hola mi nombre es Adan el Bot 🤖 de la comunidad de Golang"
		msg += " Venezuela. Y como la cancion: <<naci en esta ribera del "
		msg += "arauca vibrador, soy hermano de la espuma de las garzas de "
		msg += "las rosas y del sol.>> "
		return c.Send(msg)
	})

	bot.Handle("/status", func(c tele.Context) error {
		slog.Info("Executing /status command", slog.String("user_id", logger.ObfuscateID(c.Sender().ID)))
		//nolint:misspell
		return c.Send("De momento todo esta bien")
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
