// Package domain contains the core business models for the application.
package domain

import "time"

// Warning represents a formal moderation warning issued to a user within a chat.
type Warning struct {
	// ID is the unique identifier for the warning record.
	ID      string
	// ChatID is the Telegram ID of the chat where the warning was issued.
	ChatID  int64
	// UserID is the Telegram ID of the user receiving the warning.
	UserID  int64
	// AdminID is the Telegram ID of the administrator who issued the warning.
	AdminID int64
	// Reason is the textual justification for the warning.
	Reason  string
	// Date is the timestamp when the warning was recorded.
	Date    time.Time
}

// ChatConfig defines moderation settings and state for a specific Telegram chat.
type ChatConfig struct {
	// ChatID is the unique Telegram ID of the chat.
	ChatID int64
	// RoMode (Read-Only mode) indicates if the chat is currently restricted to admin-only messages.
	RoMode bool
}

