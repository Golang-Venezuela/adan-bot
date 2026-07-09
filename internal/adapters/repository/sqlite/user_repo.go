// Package sqlite provides a database repository implementation for the Adan Bot platform.
// It utilizes Turso/libsql to interface with SQLite-compatible environments cleanly.
package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/Golang-Venezuela/adan-bot/internal/core/domain"
	"github.com/Golang-Venezuela/adan-bot/internal/infra/logger"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	_ "github.com/tursodatabase/libsql-client-go/libsql" // Driver compatible with local SQLite and Turso
)

// userRepo represents the SQLite-backed user repository implementing the domain ports.
type userRepo struct {
	db *sqlx.DB
}

// NewUserRepository establishes a database connection and creates the target schema if it does not exist.
func NewUserRepository(dbUrl string) (*userRepo, error) {
	db, err := sql.Open("libsql", dbUrl)
	if err != nil {
		slog.Error("error opening db", slog.Any("error", err))
		return nil, fmt.Errorf("error opening db: %w", err)
	}

	dbx := sqlx.NewDb(db, "libsql")

	// Create base table if it doesn't already exist.
	query := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY,
		username TEXT,
		first_name TEXT,
		last_name TEXT,
		birthday_day INTEGER,
		birthday_month INTEGER,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	if _, err := dbx.ExecContext(context.Background(), query); err != nil {
		slog.Error("error creating users table", slog.Any("error", err))
		db.Close()
		return nil, fmt.Errorf("error creating users table: %w", err)
	}

	// Run migration steps to add birthday columns if they don't exist in an existing database
	_, _ = dbx.ExecContext(context.Background(), "ALTER TABLE users ADD COLUMN birthday_day INTEGER;")
	_, _ = dbx.ExecContext(context.Background(), "ALTER TABLE users ADD COLUMN birthday_month INTEGER;")

	return &userRepo{db: dbx}, nil
}

// SaveUser inserts a newly authenticated user into the database or updates their current profile
// if their Telegram ID already exists, enabling deterministic user synchronization.
func (r *userRepo) SaveUser(ctx context.Context, u domain.User) error {
	query := `
		INSERT INTO users (id, username, first_name, last_name, birthday_day, birthday_month, created_at)
		VALUES (:id, :username, :first_name, :last_name, :birthday_day, :birthday_month, :created_at)
		ON CONFLICT(id) DO UPDATE SET 
			username = excluded.username,
			first_name = excluded.first_name,
			last_name = excluded.last_name,
			birthday_day = COALESCE(excluded.birthday_day, users.birthday_day),
			birthday_month = COALESCE(excluded.birthday_month, users.birthday_month);
	`
	_, err := r.db.NamedExecContext(ctx, query, u)
	if err != nil {
		slog.Error("error saving user", slog.Any("error", err), slog.String("user_id", logger.ObfuscateID(u.ID)))
		return fmt.Errorf("error saving user: %w", err)
	}

	slog.Debug("User saved successfully", slog.String("user_id", logger.ObfuscateID(u.ID)))
	return nil
}

// GetUserByID fetches a user's record from the database based on their unique Telegram ID.
// It securely maps the row scan to a domain entity.
// It returns nil, nil when the user is not found to prevent masking valid non-existences as SQL errors.
func (r *userRepo) GetUserByID(ctx context.Context, id int64) (*domain.User, error) {
	var u domain.User
	query := `SELECT id, username, first_name, last_name, birthday_day, birthday_month, created_at FROM users WHERE id = ?`

	err := r.db.GetContext(ctx, &u, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			slog.Debug("User not found", slog.String("user_id", logger.ObfuscateID(id)))
			return nil, nil // Not found
		}
		slog.Error("error getting user", slog.Any("error", err), slog.String("user_id", logger.ObfuscateID(id)))
		return nil, fmt.Errorf("error getting user: %w", err)
	}

	slog.Debug("User retrieved successfully", slog.String("user_id", logger.ObfuscateID(id)))
	return &u, nil
}

// SetBirthday sets the birthday (day and month) for a specific user ID.
func (r *userRepo) SetBirthday(ctx context.Context, userID int64, day, month int) error {
	query := `UPDATE users SET birthday_day = ?, birthday_month = ? WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, day, month, userID)
	if err != nil {
		slog.Error("error setting birthday", slog.Any("error", err), slog.String("user_id", logger.ObfuscateID(userID)))
		return fmt.Errorf("error setting birthday: %w", err)
	}

	slog.Debug("Birthday updated successfully", slog.String("user_id", logger.ObfuscateID(userID)))
	return nil
}

// RemoveBirthday clears the birthday fields for a specific user ID.
func (r *userRepo) RemoveBirthday(ctx context.Context, userID int64) error {
	query := `UPDATE users SET birthday_day = NULL, birthday_month = NULL WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		slog.Error("error removing birthday", slog.Any("error", err), slog.String("user_id", logger.ObfuscateID(userID)))
		return fmt.Errorf("error removing birthday: %w", err)
	}

	slog.Debug("Birthday removed successfully", slog.String("user_id", logger.ObfuscateID(userID)))
	return nil
}

// GetBirthdaysByDayAndMonth retrieves all users who celebrate their birthday on a specific day and month.
func (r *userRepo) GetBirthdaysByDayAndMonth(ctx context.Context, day, month int) ([]domain.User, error) {
	var users []domain.User
	query := `SELECT id, username, first_name, last_name, birthday_day, birthday_month, created_at FROM users WHERE birthday_day = ? AND birthday_month = ?`
	err := r.db.SelectContext(ctx, &users, query, day, month)
	if err != nil {
		slog.Error("error getting birthdays by day and month", slog.Any("error", err), slog.Int("day", day), slog.Int("month", month))
		return nil, fmt.Errorf("error getting birthdays by day and month: %w", err)
	}
	return users, nil
}

// GetBirthdaysByMonth retrieves all users who celebrate their birthday in a specific month.
func (r *userRepo) GetBirthdaysByMonth(ctx context.Context, month int) ([]domain.User, error) {
	var users []domain.User
	query := `SELECT id, username, first_name, last_name, birthday_day, birthday_month, created_at FROM users WHERE birthday_month = ?`
	err := r.db.SelectContext(ctx, &users, query, month)
	if err != nil {
		slog.Error("error getting birthdays by month", slog.Any("error", err), slog.Int("month", month))
		return nil, fmt.Errorf("error getting birthdays by month: %w", err)
	}
	return users, nil
}
