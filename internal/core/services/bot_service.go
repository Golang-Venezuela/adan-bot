// Package services implements the core business logic and primary use cases of the application.
// It acts as the orchestrator bridging external delivery mechanisms (such as the Telegram API router) 
// and outward persistence ports (like the SQLite database).
package services

import (
	"context"
	"log/slog"
	"time"

	"github.com/Golang-Venezuela/adan-bot/internal/core/domain"
	"github.com/Golang-Venezuela/adan-bot/internal/core/ports"
	"github.com/Golang-Venezuela/adan-bot/internal/infra/logger"
)

// botService is the concrete implementation of the ports.BotService interface.
// It holds references to required outward-facing dependencies, such as the UserRepository.
type botService struct {
	userRepo ports.UserRepository
}

// NewBotService initializes and returns a new botService instance by injecting its required repository dependency.
func NewBotService(repo ports.UserRepository) ports.BotService {
	return &botService{
		userRepo: repo,
	}
}

// HandleStartHelp provides the standard response string for the /start and /help bot commands.
func (s *botService) HandleStartHelp(ctx context.Context) string {
	slog.Debug("Handling StartHelp context globally")
	return "/hola and /status."
}

// HandleHola processes the /hola command.
// It constructs the user domain entity and attempts to register or update their profile in the persistence layer.
// Whether the registration succeeds or fails, it gracefully handles the response and yields a localized greeting.
func (s *botService) HandleHola(ctx context.Context, userID int64, username, firstName, lastName string) (string, error) {
	slog.Debug("Handling Hola context", slog.String("user_id", logger.ObfuscateID(userID)))

	// 1. Save or update the user in the database (Example of interacting with the Repo port)
	user := domain.User{
		ID:        userID,
		Username:  username,
		FirstName: firstName,
		LastName:  lastName,
		CreatedAt: time.Now(),
	}

	if err := s.userRepo.SaveUser(ctx, user); err != nil {
		// Log the underlying problem but supply a friendly fallback message to the user
		return "Hola, tuve un problema interno guardando tu perfil, pero bienvenido 🤖", err
	}

	// 2. Business Logic Response
	msg := "Hola mi nombre es Adan el Bot 🤖 de la comunidad de Golang"
	msg += " Venezuela. Y como la cancion: <<naci en esta ribera del "
	msg += "arauca vibrador, soy hermano de la espuma de las garzas de "
	msg += "las rosas y del sol.>> "
	return msg, nil
}

// HandleStatus processes the /status command, yielding a static upstream health check message.
func (s *botService) HandleStatus(ctx context.Context) string {
	slog.Debug("Handling Status context globally")
	//nolint:misspell
	return "De momento todo esta bien"
}
