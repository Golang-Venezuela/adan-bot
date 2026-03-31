// Package services implements the core business logic and use cases of the application.
package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Golang-Venezuela/adan-bot/internal/core/domain"
	"github.com/Golang-Venezuela/adan-bot/internal/core/ports"
)

// moderationService implements the ports.ModerationService interface.
// It handles community management, including greeting new members, 
// issuing warnings to enforce rules, and managing restrictive modes like Read-Only.
type moderationService struct {
	repo ports.ModerationRepository
}

// NewModerationService creates a new instance of a moderation service using the provided repository.
func NewModerationService(repo ports.ModerationRepository) ports.ModerationService {
	return &moderationService{repo: repo}
}

// GetWelcomeMessage returns a localized HTML-formatted greeting message for a new user.
func (s *moderationService) GetWelcomeMessage(ctx context.Context, name string) string {
	msg := fmt.Sprintf("¡Hola <b>%s</b>! Mi nombre es Adan, el Bot 🤖 oficial de la comunidad de <b>Golang Venezuela</b>.\n\n", name)
	msg += "Recuerda leer las <b>reglas</b> en la descripción del grupo para mantener la convivencia sana.\n"
	msg += "¡Siéntete libre de presentarte al grupo y hacer tus preguntas sobre Go!"
	return msg
}

// CheckAntiSpam evaluates a message for potential spam or unauthorized links.
// It applies stricter rules to recent members (joined within the last 24 hours).
func (s *moderationService) CheckAntiSpam(ctx context.Context, chatID, userID int64, message string, memberSince time.Time) (bool, error) {
	// If the user joined within the last 24 hours (or join date is unknown), block links.
	if memberSince.IsZero() || time.Since(memberSince) < 24*time.Hour {
		messageLower := strings.ToLower(message)
		if strings.Contains(messageLower, "http://") ||
			strings.Contains(messageLower, "https://") ||
			strings.Contains(messageLower, "t.me/") ||
			strings.Contains(messageLower, "www.") {
			return true, nil // Detected restricted link.
		}
	}
	return false, nil
}

// IssueWarning records a new warning for a user and returns their total accumulated warnings.
// If a user reaches 3 warnings, the count is reset for future interactions after enforcement.
func (s *moderationService) IssueWarning(ctx context.Context, chatID, userID, adminID int64, reason string) (int, error) {
	idStr := fmt.Sprintf("%d-%d-%d", chatID, userID, time.Now().UnixNano())
	warning := domain.Warning{
		ID:      idStr,
		ChatID:  chatID,
		UserID:  userID,
		AdminID: adminID,
		Reason:  reason,
		Date:    time.Now(),
	}

	if err := s.repo.SaveWarning(ctx, warning); err != nil {
		return 0, fmt.Errorf("could not save warning: %w", err)
	}

	count, err := s.repo.GetWarningsCount(ctx, chatID, userID)
	if err != nil {
		return 0, fmt.Errorf("could not get warnings count: %w", err)
	}

	if count >= 3 {
		// Reset warnings after reaching the threshold.
		_ = s.repo.ResetWarnings(ctx, chatID, userID)
	}

	return count, nil
}

// ToggleRoMode enables or disables Read-Only mode for a specific chat.
func (s *moderationService) ToggleRoMode(ctx context.Context, chatID, adminID int64, enabled bool) error {
	return s.repo.SetRoMode(ctx, chatID, enabled)
}

// IsRoModeActive checks if Read-Only mode is currently active for a given chat.
func (s *moderationService) IsRoModeActive(ctx context.Context, chatID int64) bool {
	active, _ := s.repo.GetRoMode(ctx, chatID)
	return active
}

