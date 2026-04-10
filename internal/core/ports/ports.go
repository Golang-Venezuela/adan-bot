// Package ports defines the primary interfaces (ports in Hexagonal Architecture)
// that isolate the application's business logic from external frameworks, databases, and delivery mechanisms.
package ports

import (
	"context"
	"time"

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

// ModerationRepository handles persistence of moderation records.
type ModerationRepository interface {
	SaveWarning(ctx context.Context, warning domain.Warning) error
	GetWarningsCount(ctx context.Context, chatID, userID int64) (int, error)
	ResetWarnings(ctx context.Context, chatID, userID int64) error
	SetRoMode(ctx context.Context, chatID int64, enabled bool) error
	GetRoMode(ctx context.Context, chatID int64) (bool, error)
}

// ModerationService defines the core business operations for moderation features.
type ModerationService interface {
	GetWelcomeMessage(ctx context.Context, username string) string
	CheckAntiSpam(ctx context.Context, chatID, userID int64, message string, memberSince time.Time) (bool, error)
	IssueWarning(ctx context.Context, chatID, userID, adminID int64, reason string) (int, error)
	ToggleRoMode(ctx context.Context, chatID, adminID int64, enabled bool) error
	IsRoModeActive(ctx context.Context, chatID int64) bool
}
