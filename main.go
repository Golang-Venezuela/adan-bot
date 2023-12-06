package main

import (
	"log"

	"github.com/Golang-Venezuela/adan-bot/envs"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {
	APITOKEN := envs.Get("TELEGRAM_API_TOKEN", "Token-default")

	bot, err := tgbotapi.NewBotAPI(APITOKEN)
	if err != nil {
		panic(err)
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil { // ignore any non-Message updates
			continue
		}

		if !update.Message.IsCommand() { // ignore any non-command Messages
			continue
		}

		// Create a new MessageConfig. We don't have text yet,
		// so we leave it empty.
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")

		// Extract the command from the Message.
		switch update.Message.Command() {
		case "help":
			msg.Text = "/hola and /status."
		case "hola":
			msg.Text = "Hola mi nombre es Adan el Bot 🤖 de la comunidad de Golang"
			msg.Text += " Venezuela. Y como la cancion: <<naci en esta ribera del "
			msg.Text += "arauca vibrador, soy hermano de la espuma de las garzas de "
			msg.Text += "las rosas y del sol.>> "
		case "status":
			//nolint:misspell
			msg.Text = "De momento todo esta bien"
		default:
			msg.Text = "I don't know that command"
		}

		if _, err := bot.Send(msg); err != nil {
			log.Panic(err)
		}
	}
}
