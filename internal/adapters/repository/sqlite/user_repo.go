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
	_ "github.com/tursodatabase/libsql-client-go/libsql" // Driver compatible with local SQLite and Turso
)

// userRepo represents the SQLite-backed user repository implementing the domain ports.
type userRepo struct {
	db *sql.DB
}

// NewUserRepository establishes a database connection and creates the target schema if it does not exist.
func NewUserRepository(dbUrl string) (*userRepo, error) {
	db, err := sql.Open("libsql", dbUrl)
	if err != nil {
		slog.Error("error opening db", slog.Any("error", err))
		return nil, fmt.Errorf("error opening db: %w", err)
	}

	// Create base table if it doesn't already exist.
	query := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY,
		username TEXT,
		first_name TEXT,
		last_name TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	if _, err := db.ExecContext(context.Background(), query); err != nil {
		slog.Error("error creating users table", slog.Any("error", err))
		return nil, fmt.Errorf("error creating users table: %w", err)
	}

	return &userRepo{db: db}, nil
}

// SaveUser inserts a newly authenticated user into the database or updates their current profile
// if their Telegram ID already exists, enabling deterministic user synchronization.
func (r *userRepo) SaveUser(ctx context.Context, u domain.User) error {
	query := `
		INSERT INTO users (id, username, first_name, last_name, created_at)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET 
			username = excluded.username,
			first_name = excluded.first_name,
			last_name = excluded.last_name;
	`
	_, err := r.db.ExecContext(ctx, query, u.ID, u.Username, u.FirstName, u.LastName, u.CreatedAt)
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
	query := `SELECT id, username, first_name, last_name, created_at FROM users WHERE id = ?`

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&u.ID, &u.Username, &u.FirstName, &u.LastName, &u.CreatedAt,
	)
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
