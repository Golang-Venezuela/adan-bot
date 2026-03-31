// Package services_test contains unit tests for the core business logic implementations.
package services

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/Golang-Venezuela/adan-bot/internal/adapters/repository"
)

// TestModerationService_GetWelcomeMessage verifies that the welcome message 
// generated for new users contains their first name and community branding.
func TestModerationService_GetWelcomeMessage(t *testing.T) {
	repo := repository.NewMemoryModerationRepo()
	svc := NewModerationService(repo)
	name := "Juan"

	msg := svc.GetWelcomeMessage(context.Background(), name)
	if !strings.Contains(msg, name) {
		t.Errorf("expected welcome message to contain %q, but got %q", name, msg)
	}
	if !strings.Contains(msg, "Golang Venezuela") {
		t.Error("expected welcome message to contain community name")
	}
}

// TestModerationService_CheckAntiSpam runs a suite of table-driven tests to verify 
// that links are correctly identified and blocked for new members while allowed 
// for established community members.
func TestModerationService_CheckAntiSpam(t *testing.T) {
	repo := repository.NewMemoryModerationRepo()
	svc := NewModerationService(repo)
	ctx := context.Background()
	chatID := int64(123)
	userID := int64(456)

	tests := []struct {
		name        string
		message     string
		memberSince time.Time
		wantSpam    bool
	}{
		{
			name:        "New member (<24h) with link http",
			message:     "Check this http://spam.com",
			memberSince: time.Now().Add(-1 * time.Hour),
			wantSpam:    true,
		},
		{
			name:        "New member (<24h) with link https",
			message:     "Check this https://spam.com",
			memberSince: time.Now().Add(-23 * time.Hour),
			wantSpam:    true,
		},
		{
			name:        "New member (<24h) with link t.me",
			message:     "Join my channel t.me/spam",
			memberSince: time.Now().Add(-10 * time.Hour),
			wantSpam:    true,
		},
		{
			name:        "New member (<24h) with normal text",
			message:     "Hello group!",
			memberSince: time.Now().Add(-5 * time.Hour),
			wantSpam:    false,
		},
		{
			name:        "Old member (>24h) with link",
			message:     "Useful link: https://golang.org",
			memberSince: time.Now().Add(-25 * time.Hour),
			wantSpam:    false,
		},
		{
			name:        "Unknown join date (Zero time) with link",
			message:     "Spam link: www.spam.com",
			memberSince: time.Time{}, // Zero time
			wantSpam:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := svc.CheckAntiSpam(ctx, chatID, userID, tt.message, tt.memberSince)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.wantSpam {
				t.Errorf("CheckAntiSpam() = %v, want %v", got, tt.wantSpam)
			}
		})
	}
}

// TestModerationService_IssueWarning_Lifecycle verifies the full flow of issuing warnings,
// including the automated reset of the counter once the threshold (3 warnings) is reached.
func TestModerationService_IssueWarning_Lifecycle(t *testing.T) {
	repo := repository.NewMemoryModerationRepo()
	svc := NewModerationService(repo)
	ctx := context.Background()
	chatID := int64(1234)
	userID := int64(5678)
	adminID := int64(9999)

	// 1st warning
	count, _ := svc.IssueWarning(ctx, chatID, userID, adminID, "Rule 1")
	if count != 1 {
		t.Errorf("expected 1 warning, got %d", count)
	}

	// 2nd warning
	count, _ = svc.IssueWarning(ctx, chatID, userID, adminID, "Rule 2")
	if count != 2 {
		t.Errorf("expected 2 warnings, got %d", count)
	}

	// 3rd warning (should return 3, but the repo should be reset after)
	count, _ = svc.IssueWarning(ctx, chatID, userID, adminID, "Rule 3")
	if count != 3 {
		t.Errorf("expected 3 warnings, got %d", count)
	}

	// Verify repo is reset (4th warning will be 1)
	count, _ = svc.IssueWarning(ctx, chatID, userID, adminID, "Rule 4")
	if count != 1 {
		t.Errorf("expected count to reset to 1 after 3 warnings, got %d", count)
	}
}

// TestModerationService_RoMode verifies that the service correctly interacts with 
// the persistence layer to manage and check the Read-Only mode status.
func TestModerationService_RoMode(t *testing.T) {
	repo := repository.NewMemoryModerationRepo()
	svc := NewModerationService(repo)
	ctx := context.Background()
	chatID := int64(100)
	adminID := int64(200)

	// Initial
	if svc.IsRoModeActive(ctx, chatID) {
		t.Error("expected RoMode to be inactive")
	}

	// Enable
	_ = svc.ToggleRoMode(ctx, chatID, adminID, true)
	if !svc.IsRoModeActive(ctx, chatID) {
		t.Error("expected RoMode to be active")
	}

	// Disable
	_ = svc.ToggleRoMode(ctx, chatID, adminID, false)
	if svc.IsRoModeActive(ctx, chatID) {
		t.Error("expected RoMode to be inactive")
	}
}

