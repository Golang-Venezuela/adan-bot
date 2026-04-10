// Package repository_test contains integration and unit tests for repository implementations.
package repository

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/Golang-Venezuela/adan-bot/internal/core/domain"
)

// TestMemoryModerationRepo_Warnings verifies the basic CRUD operations for user warnings
// in the in-memory repository, including counting and resetting.
func TestMemoryModerationRepo_Warnings(t *testing.T) {
	repo := NewMemoryModerationRepo()
	ctx := context.Background()

	chatID := int64(123)
	userID := int64(456)
	adminID := int64(789)

	// 1. Initially 0 warnings
	count, err := repo.GetWarningsCount(ctx, chatID, userID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 warnings, got %d", count)
	}

	// 2. Save a warning
	warning := domain.Warning{
		ChatID:  chatID,
		UserID:  userID,
		AdminID: adminID,
		Reason:  "Spamming",
		Date:    time.Now(),
	}
	if err := repo.SaveWarning(ctx, warning); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 3. Count should be 1
	count, _ = repo.GetWarningsCount(ctx, chatID, userID)
	if count != 1 {
		t.Errorf("expected 1 warning, got %d", count)
	}

	// 4. Save another warning
	if err := repo.SaveWarning(ctx, warning); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	count, _ = repo.GetWarningsCount(ctx, chatID, userID)
	if count != 2 {
		t.Errorf("expected 2 warnings, got %d", count)
	}

	// 5. Reset warnings
	if err := repo.ResetWarnings(ctx, chatID, userID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	count, _ = repo.GetWarningsCount(ctx, chatID, userID)
	if count != 0 {
		t.Errorf("expected 0 warnings after reset, got %d", count)
	}
}

// TestMemoryModerationRepo_RoMode ensures that Read-Only mode can be correctly
// toggled and its state is persisted per chat.
func TestMemoryModerationRepo_RoMode(t *testing.T) {
	repo := NewMemoryModerationRepo()
	ctx := context.Background()
	chatID := int64(123)

	// 1. Default should be false
	active, err := repo.GetRoMode(ctx, chatID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if active {
		t.Error("expected RoMode to be inactive by default")
	}

	// 2. Enable RoMode
	if err := repo.SetRoMode(ctx, chatID, true); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	active, _ = repo.GetRoMode(ctx, chatID)
	if !active {
		t.Error("expected RoMode to be active")
	}

	// 3. Disable RoMode
	if err := repo.SetRoMode(ctx, chatID, false); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	active, _ = repo.GetRoMode(ctx, chatID)
	if active {
		t.Error("expected RoMode to be inactive")
	}
}

// TestMemoryModerationRepo_Concurrency stresses the repository with multiple
// simultaneous reads and writes to verify thread safety and race-free operation.
func TestMemoryModerationRepo_Concurrency(t *testing.T) {
	repo := NewMemoryModerationRepo()
	ctx := context.Background()
	chatID := int64(123)
	userID := int64(456)

	var wg sync.WaitGroup
	iterations := 100

	// Concurrent writes
	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			_ = repo.SaveWarning(ctx, domain.Warning{
				ChatID: chatID,
				UserID: userID,
				Reason: "test",
			})
			_, _ = repo.GetRoMode(ctx, chatID)
		}(i)
	}

	// Concurrent reads
	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = repo.GetWarningsCount(ctx, chatID, userID)
			_ = repo.SetRoMode(ctx, chatID, true)
		}()
	}

	wg.Wait()

	count, _ := repo.GetWarningsCount(ctx, chatID, userID)
	if count != iterations {
		t.Errorf("expected %d warnings, got %d", iterations, count)
	}
}
