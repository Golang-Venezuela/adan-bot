// Package helpers contains utility variables and logic
// utilized across various application boundaries to standardize general behaviors and error handling.
package helpers

import "errors"

// Global error variables that can be utilized and referenced across all structural layers
// to cleanly validate, compare, and contrast specific error types (e.g., using errors.Is(err, ErrUserNotFound)).
var (
	// ErrUserNotFound indicates that the requested user is absent from the datastore.
	ErrUserNotFound = errors.New("user no encontrado en la base de datos")
	// ErrDatabaseConnection indicates a failure when establishing a session with the database.
	ErrDatabaseConnection = errors.New("no se pudo establecer conexión con la base de datos")
	// ErrInvalidCommand denotes that the processed generic payload does not match any recognized bot instruction.
	ErrInvalidCommand = errors.New("comando introducido no es válido")
	// ErrMissingAPIKey is returned when the external API token is omitted from the application's environment payload.
	ErrMissingAPIKey = errors.New("TELEGRAM_API_TOKEN no está definido en el entorno")
	// ErrBotInitialization denotes an underlying transport or structural failure when instantiating the Telegram Bot API client.
	ErrBotInitialization = errors.New("no se pudo inicializar la conexión con el API de Telegram")
)
