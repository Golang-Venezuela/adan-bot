// Package repository provides implementations for the system's data access layers.
package repository

import (
	"context"
	"sync"

	"github.com/Golang-Venezuela/adan-bot/internal/core/domain"
	"github.com/Golang-Venezuela/adan-bot/internal/core/ports"
)

// memoryModerationRepo is an in-memory implementation of the ports.ModerationRepository.
// This implementation is thread-safe and suitable for testing or small-scale deployments.
type memoryModerationRepo struct {
	mu       sync.RWMutex
	warnings map[int64]map[int64][]domain.Warning // chatID -> userID -> []Warning
	roMode   map[int64]bool                       // chatID -> enabled
}

// NewMemoryModerationRepo creates a new instance of an in-memory moderation repository.
func NewMemoryModerationRepo() ports.ModerationRepository {
	return &memoryModerationRepo{
		warnings: make(map[int64]map[int64][]domain.Warning),
		roMode:   make(map[int64]bool),
	}
}

// SaveWarning persists a new warning record for a specific user in a chat.
func (r *memoryModerationRepo) SaveWarning(ctx context.Context, warning domain.Warning) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.warnings[warning.ChatID] == nil {
		r.warnings[warning.ChatID] = make(map[int64][]domain.Warning)
	}

	r.warnings[warning.ChatID][warning.UserID] = append(
		r.warnings[warning.ChatID][warning.UserID],
		warning,
	)

	return nil
}

// GetWarningsCount returns the total number of warnings accumulated by a user in a specific chat.
func (r *memoryModerationRepo) GetWarningsCount(ctx context.Context, chatID, userID int64) (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.warnings[chatID] == nil {
		return 0, nil
	}

	return len(r.warnings[chatID][userID]), nil
}

// ResetWarnings clears all warning records for a specific user within a chat.
func (r *memoryModerationRepo) ResetWarnings(ctx context.Context, chatID, userID int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.warnings[chatID] != nil {
		delete(r.warnings[chatID], userID)
	}

	return nil
}

// SetRoMode updates the Read-Only mode status for a specific chat.
func (r *memoryModerationRepo) SetRoMode(ctx context.Context, chatID int64, enabled bool) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.roMode[chatID] = enabled
	return nil
}

// GetRoMode retrieves the current Read-Only mode status for a specific chat.
func (r *memoryModerationRepo) GetRoMode(ctx context.Context, chatID int64) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.roMode[chatID], nil
}
