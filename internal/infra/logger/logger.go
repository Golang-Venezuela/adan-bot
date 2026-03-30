// Package logger provides a centralized, standard structured logger for the Adan Bot application.
// It integrates automatic scrubbing mechanisms to mask sensitive data, such as private API tokens, prior to IO boundary logging.
package logger

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
)

// Logger is the application-wide logging instance properly configured with the standard JSON handler.
var Logger *slog.Logger

// init initializes the global slog default logger substituting sensitive environmental values (like tokens)
// with obfuscated string permutations globally through a robust ReplaceAttr hook format.
func init() {
	// Initialize logger with JSON format and level based on environment
	logLevel := slog.LevelInfo
	if os.Getenv("DEBUG") == "true" {
		logLevel = slog.LevelDebug
	}

	token := strings.Trim(os.Getenv("TELEGRAM_API_TOKEN"), `"`)

	opts := &slog.HandlerOptions{
		Level: logLevel,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// If we have a token and the value is a string, scrub it
			if token != "" && a.Value.Kind() == slog.KindString {
				s := a.Value.String()
				if strings.Contains(s, token) {
					a.Value = slog.StringValue(strings.ReplaceAll(s, token, Obfuscate(token)))
				}
			} else if token != "" && a.Value.Kind() == slog.KindAny {
				// Handle error type specifically, as some slog calls pass err inside 'Any'
				if err, ok := a.Value.Any().(error); ok {
					s := err.Error()
					if strings.Contains(s, token) {
						a.Value = slog.StringValue(strings.ReplaceAll(s, token, Obfuscate(token)))
					}
				}
			}
			return a
		},
	}

	Logger = slog.New(slog.NewJSONHandler(os.Stdout, opts))
	slog.SetDefault(Logger)
}

// Obfuscate masks all characters of a given string with asterisks, substituting everything
// except for the last character, inherently preventing log exposure of secrets. It gracefully returns an empty string if it evaluates an empty entity.
func Obfuscate(s string) string {
	if len(s) == 0 {
		return ""
	}
	return strings.Repeat("*", len(s)-1) + s[len(s)-1:]
}

// ObfuscateID securely transforms an integer ID metric to a string and masks its values adhering to the Obfuscate substitution rules.
func ObfuscateID(id int64) string {
	return Obfuscate(fmt.Sprintf("%d", id))
}
