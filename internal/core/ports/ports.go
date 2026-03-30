// Package ports defines the primary interfaces (ports in Hexagonal Architecture)
// that isolate the application's business logic from external frameworks, databases, and delivery mechanisms.
package ports

import (
	"context"
	"github.com/Golang-Venezuela/adan-bot/internal/core/domain"
)

// UserRepository defines the persistence contract for user data.
// It allows the business logic to securely save and retrieve user records 
// without being coupled to a specific underlying database technology.
type UserRepository interface {
	SaveUser(ctx context.Context, user domain.User) error
	GetUserByID(ctx context.Context, id int64) (*domain.User, error)
}

// BotService defines the core business operations and use cases for the Adan Bot.
// It encapsulates the logic that is executed when users trigger specific conversational commands.
type BotService interface {
	HandleStartHelp(ctx context.Context) string
	HandleHola(ctx context.Context, userID int64, username, firstName, lastName string) (string, error)
	HandleStatus(ctx context.Context) string
}
