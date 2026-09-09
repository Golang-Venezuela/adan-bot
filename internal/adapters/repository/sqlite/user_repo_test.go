package sqlite

import (
	"context"
	"testing"
	"time"

	"github.com/Golang-Venezuela/adan-bot/internal/core/domain"
	_ "github.com/mattn/go-sqlite3"
)

func TestUserRepository_SQLite(t *testing.T) {
	ctx := context.Background()

	// Use an in-memory sqlite database for testing
	repo, err := NewUserRepository("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("failed to initialize repository: %v", err)
	}
	defer repo.db.DB.Close()

	// Test 1: SaveUser and GetUserByID
	user := domain.User{
		ID:        999,
		Username:  "gopher_venezuela",
		FirstName: "Gopher",
		LastName:  "VZLA",
		CreatedAt: time.Now().Round(time.Second),
	}

	err = repo.SaveUser(ctx, user)
	if err != nil {
		t.Fatalf("failed to save user: %v", err)
	}

	got, err := repo.GetUserByID(ctx, 999)
	if err != nil {
		t.Fatalf("failed to get user by id: %v", err)
	}

	if got == nil {
		t.Fatal("expected user to be found, got nil")
	}

	if got.Username != user.Username || got.FirstName != user.FirstName || got.LastName != user.LastName {
		t.Errorf("GetUserByID() = %+v, want %+v", got, user)
	}

	// Test 2: GetUserByID not found
	gotNotFound, err := repo.GetUserByID(ctx, 9999)
	if err != nil {
		t.Fatalf("unexpected error getting non-existent user: %v", err)
	}
	if gotNotFound != nil {
		t.Errorf("expected nil for non-existent user, got %+v", gotNotFound)
	}

	// Test 3: SaveUser updates profile on conflict (keeps birthday nil)
	user.FirstName = "NewName"
	err = repo.SaveUser(ctx, user)
	if err != nil {
		t.Fatalf("failed to update user: %v", err)
	}
	gotUpdated, _ := repo.GetUserByID(ctx, 999)
	if gotUpdated.FirstName != "NewName" {
		t.Errorf("expected first name to be updated, got %q", gotUpdated.FirstName)
	}
	if gotUpdated.BirthdayDay != nil || gotUpdated.BirthdayMonth != nil {
		t.Errorf("expected birthday to remain nil")
	}

	// Test 4: SetBirthday and GetBirthdaysByDayAndMonth
	err = repo.SetBirthday(ctx, 999, 24, 7)
	if err != nil {
		t.Fatalf("failed to set birthday: %v", err)
	}

	gotBday, _ := repo.GetUserByID(ctx, 999)
	if gotBday.BirthdayDay == nil || *gotBday.BirthdayDay != 24 || gotBday.BirthdayMonth == nil || *gotBday.BirthdayMonth != 7 {
		t.Errorf("birthday not set correctly in user record: %+v", gotBday)
	}

	// Test 5: GetBirthdaysByDayAndMonth filter
	bdayUsers, err := repo.GetBirthdaysByDayAndMonth(ctx, 24, 7)
	if err != nil {
		t.Fatalf("failed to get birthdays by day and month: %v", err)
	}
	if len(bdayUsers) != 1 || bdayUsers[0].ID != 999 {
		t.Errorf("expected 1 birthday user with ID 999, got %v", bdayUsers)
	}

	// Non-matching day
	emptyBdayUsers, _ := repo.GetBirthdaysByDayAndMonth(ctx, 25, 7)
	if len(emptyBdayUsers) != 0 {
		t.Errorf("expected 0 birthday users, got %v", emptyBdayUsers)
	}

	// Test 6: GetBirthdaysByMonth
	monthUsers, err := repo.GetBirthdaysByMonth(ctx, 7)
	if err != nil {
		t.Fatalf("failed to get birthdays by month: %v", err)
	}
	if len(monthUsers) != 1 || monthUsers[0].ID != 999 {
		t.Errorf("expected 1 user in month 7, got %v", monthUsers)
	}

	// Test 7: SaveUser on conflict does not wipe out birthday (COALESCE check)
	user.LastName = "UpdatedLastName"
	err = repo.SaveUser(ctx, user)
	if err != nil {
		t.Fatalf("failed to save user: %v", err)
	}
	gotAfterSave, _ := repo.GetUserByID(ctx, 999)
	if gotAfterSave.LastName != "UpdatedLastName" {
		t.Errorf("expected last name to update, got %q", gotAfterSave.LastName)
	}
	if gotAfterSave.BirthdayDay == nil || *gotAfterSave.BirthdayDay != 24 {
		t.Errorf("expected birthday to be preserved, got nil or wrong day")
	}

	// Test 8: RemoveBirthday
	err = repo.RemoveBirthday(ctx, 999)
	if err != nil {
		t.Fatalf("failed to remove birthday: %v", err)
	}
	gotRemoved, _ := repo.GetUserByID(ctx, 999)
	if gotRemoved.BirthdayDay != nil || gotRemoved.BirthdayMonth != nil {
		t.Errorf("expected birthday fields to be nil after removal, got %v / %v", gotRemoved.BirthdayDay, gotRemoved.BirthdayMonth)
	}
}
