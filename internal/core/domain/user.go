package domain

import "time"

// User representa a un usuario que interactúa con el bot
type User struct {
	ID            int64     `db:"id"`
	Username      string    `db:"username"`
	FirstName     string    `db:"first_name"`
	LastName      string    `db:"last_name"`
	BirthdayDay   *int      `db:"birthday_day"`
	BirthdayMonth *int      `db:"birthday_month"`
	CreatedAt     time.Time `db:"created_at"`
}
