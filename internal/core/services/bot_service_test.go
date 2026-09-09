package services

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/Golang-Venezuela/adan-bot/internal/core/domain"
)

type mockUserRepo struct {
	users map[int64]domain.User
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{
		users: make(map[int64]domain.User),
	}
}

func (m *mockUserRepo) SaveUser(ctx context.Context, u domain.User) error {
	m.users[u.ID] = u
	return nil
}

func (m *mockUserRepo) GetUserByID(ctx context.Context, id int64) (*domain.User, error) {
	u, ok := m.users[id]
	if !ok {
		return nil, nil
	}
	return &u, nil
}

func (m *mockUserRepo) SetBirthday(ctx context.Context, userID int64, day, month int) error {
	u, ok := m.users[userID]
	if !ok {
		u = domain.User{ID: userID}
	}
	u.BirthdayDay = &day
	u.BirthdayMonth = &month
	m.users[userID] = u
	return nil
}

func (m *mockUserRepo) RemoveBirthday(ctx context.Context, userID int64) error {
	u, ok := m.users[userID]
	if !ok {
		return nil
	}
	u.BirthdayDay = nil
	u.BirthdayMonth = nil
	m.users[userID] = u
	return nil
}

func (m *mockUserRepo) GetBirthdaysByDayAndMonth(ctx context.Context, day, month int) ([]domain.User, error) {
	var result []domain.User
	for _, u := range m.users {
		if u.BirthdayDay != nil && *u.BirthdayDay == day && u.BirthdayMonth != nil && *u.BirthdayMonth == month {
			result = append(result, u)
		}
	}
	return result, nil
}

func (m *mockUserRepo) GetBirthdaysByMonth(ctx context.Context, month int) ([]domain.User, error) {
	var result []domain.User
	for _, u := range m.users {
		if u.BirthdayMonth != nil && *u.BirthdayMonth == month {
			result = append(result, u)
		}
	}
	return result, nil
}

func TestBotService_HandleSetBirthday(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewBotService(repo)
	ctx := context.Background()
	userID := int64(12345)

	tests := []struct {
		name     string
		input    string
		wantOk   bool
		contains string
	}{
		{
			name:     "Valid birthday Oct 15",
			input:    "15/10",
			wantOk:   true,
			contains: "¡Guardado!",
		},
		{
			name:     "Valid birthday leap day Feb 29",
			input:    "29/02",
			wantOk:   true,
			contains: "¡Guardado!",
		},
		{
			name:     "Invalid day 32",
			input:    "32/01",
			wantOk:   false,
			contains: "Error",
		},
		{
			name:     "Invalid month 13",
			input:    "15/13",
			wantOk:   false,
			contains: "Error",
		},
		{
			name:     "Invalid separator",
			input:    "15-10",
			wantOk:   false,
			contains: "Error",
		},
		{
			name:     "Empty input",
			input:    "",
			wantOk:   false,
			contains: "Error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg, err := svc.HandleSetBirthday(ctx, userID, tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !strings.Contains(msg, tt.contains) {
				t.Errorf("expected message to contain %q, but got %q", tt.contains, msg)
			}

			if tt.wantOk {
				user, err := repo.GetUserByID(ctx, userID)
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if user.BirthdayDay == nil || user.BirthdayMonth == nil {
					t.Errorf("expected birthday to be saved in repository")
				}
			}
		})
	}
}

func TestBotService_HandleRemoveBirthday(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewBotService(repo)
	ctx := context.Background()
	userID := int64(12345)

	// Set initial birthday
	_ = repo.SetBirthday(ctx, userID, 15, 10)

	msg, err := svc.HandleRemoveBirthday(ctx, userID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(msg, "eliminada") {
		t.Errorf("expected confirmation message, got %q", msg)
	}

	user, _ := repo.GetUserByID(ctx, userID)
	if user.BirthdayDay != nil || user.BirthdayMonth != nil {
		t.Errorf("expected birthday to be nil after deletion")
	}
}

func TestBotService_HandleGetBirthdaysOfMonth(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewBotService(repo)
	ctx := context.Background()

	loc, err := time.LoadLocation("America/Caracas")
	if err != nil {
		loc = time.UTC
	}
	now := time.Now().In(loc)
	currentMonth := int(now.Month())

	// Add users
	user1 := domain.User{ID: 1, Username: "user_one", FirstName: "One"}
	user2 := domain.User{ID: 2, FirstName: "Two"} // No username

	_ = repo.SaveUser(ctx, user1)
	_ = repo.SaveUser(ctx, user2)

	// Set birthdays
	_ = repo.SetBirthday(ctx, 1, 20, currentMonth) // user1
	_ = repo.SetBirthday(ctx, 2, 5, currentMonth)  // user2

	msg, err := svc.HandleGetBirthdaysOfMonth(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify header contains current month name (translated)
	monthName := translateMonth(currentMonth)
	if !strings.Contains(msg, monthName) {
		t.Errorf("expected message to contain %q, but got %q", monthName, msg)
	}

	// Verify user1 is mentioned using @username
	if !strings.Contains(msg, "@user_one") {
		t.Errorf("expected message to contain username @user_one, got %q", msg)
	}

	// Verify user2 (no username) is mentioned using FirstName html link
	if !strings.Contains(msg, "Two") || !strings.Contains(msg, "tg://user?id=2") {
		t.Errorf("expected message to contain FirstName link for user 2, got %q", msg)
	}

	// Verify chronological sorting (Day 05 should appear before Day 20)
	idx5 := strings.Index(msg, "Día 05")
	idx2 := strings.Index(msg, "Día 20")
	if idx5 == -1 || idx2 == -1 || idx5 > idx2 {
		t.Errorf("expected Day 05 to be listed before Day 20, got indexing: 05=%d, 20=%d", idx5, idx2)
	}
}

func TestBotService_HandleTodayBirthdays(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewBotService(repo)
	ctx := context.Background()

	loc, err := time.LoadLocation("America/Caracas")
	if err != nil {
		loc = time.UTC
	}
	now := time.Now().In(loc)
	day := now.Day()
	month := int(now.Month())

	// No birthdays today
	msg, err := svc.HandleTodayBirthdays(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msg != "" {
		t.Errorf("expected empty message when there are no birthdays today, got %q", msg)
	}

	// Add single birthday today
	u1 := domain.User{ID: 10, Username: "celebrant1", FirstName: "C1"}
	_ = repo.SaveUser(ctx, u1)
	_ = repo.SetBirthday(ctx, 10, day, month)

	msg, err = svc.HandleTodayBirthdays(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(msg, "@celebrant1") || !strings.Contains(msg, "cumpleaños") {
		t.Errorf("expected birthday greeting for @celebrant1, got %q", msg)
	}

	// Add second birthday today
	u2 := domain.User{ID: 11, FirstName: "C2"} // No username
	_ = repo.SaveUser(ctx, u2)
	_ = repo.SetBirthday(ctx, 11, day, month)

	msg, err = svc.HandleTodayBirthdays(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(msg, "@celebrant1") || !strings.Contains(msg, "C2") || !strings.Contains(msg, "y") {
		t.Errorf("expected joint birthday greeting for both users, got %q", msg)
	}
}
