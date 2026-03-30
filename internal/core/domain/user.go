package domain

import "time"

// User representa a un usuario que interactúa con el bot
type User struct {
	ID        int64
	Username  string
	FirstName string
	LastName  string
	CreatedAt time.Time
}
