// Package services implements the core business logic and primary use cases of the application.
// It acts as the orchestrator bridging external delivery mechanisms (such as the Telegram API router)
// and outward persistence ports (like the SQLite database).
package services

import (
	"context"
	"fmt"
	"log/slog"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/Golang-Venezuela/adan-bot/internal/core/domain"
	"github.com/Golang-Venezuela/adan-bot/internal/core/ports"
	"github.com/Golang-Venezuela/adan-bot/internal/infra/logger"
)

// botService is the concrete implementation of the ports.BotService interface.
// It holds references to required outward-facing dependencies, such as the UserRepository.
type botService struct {
	userRepo ports.UserRepository
}

// NewBotService initializes and returns a new botService instance by injecting its required repository dependency.
func NewBotService(repo ports.UserRepository) ports.BotService {
	return &botService{
		userRepo: repo,
	}
}

// HandleStartHelp provides the standard response string for the /start and /help bot commands.
func (s *botService) HandleStartHelp(ctx context.Context) string {
	slog.Debug("Handling StartHelp context globally")
	return "/hola, /status, /micumple DD/MM, /borrarcumple and /cumples_mes."
}

// HandleHola processes the /hola command.
// It constructs the user domain entity and attempts to register or update their profile in the persistence layer.
// Whether the registration succeeds or fails, it gracefully handles the response and yields a localized greeting.
func (s *botService) HandleHola(ctx context.Context, userID int64, username, firstName, lastName string) (string, error) {
	slog.Debug("Handling Hola context", slog.String("user_id", logger.ObfuscateID(userID)))

	// 1. Save or update the user in the database (Example of interacting with the Repo port)
	user := domain.User{
		ID:        userID,
		Username:  username,
		FirstName: firstName,
		LastName:  lastName,
		CreatedAt: time.Now(),
	}

	if err := s.userRepo.SaveUser(ctx, user); err != nil {
		// Log the underlying problem but supply a friendly fallback message to the user
		return "Hola, tuve un problema interno guardando tu perfil, pero bienvenido 🤖", err
	}

	// 2. Business Logic Response
	msg := "Hola mi nombre es Adan el Bot 🤖 de la comunidad de Golang"
	msg += " Venezuela. Y como la cancion: <<naci en esta ribera del "
	msg += "arauca vibrador, soy hermano de la espuma de las garzas de "
	msg += "las rosas and del sol.>> "
	return msg, nil
}

// HandleStatus processes the /status command, yielding a static upstream health check message.
func (s *botService) HandleStatus(ctx context.Context) string {
	slog.Debug("Handling Status context globally")
	//nolint:misspell
	return "De momento todo esta bien"
}

// HandleSetBirthday parses, validates, and registers a user's birthday date.
func (s *botService) HandleSetBirthday(ctx context.Context, userID int64, dateStr string) (string, error) {
	day, month, err := parseBirthday(dateStr)
	if err != nil {
		slog.Warn("Invalid birthday format", slog.String("input", dateStr), slog.Any("error", err))
		return fmt.Sprintf("❌ Error: %s. Por favor usa el formato DD/MM (ejemplo: /micumple 15/10)", err.Error()), nil
	}

	err = s.userRepo.SetBirthday(ctx, userID, day, month)
	if err != nil {
		return "❌ Tuve un problema interno al guardar tu cumpleaños.", err
	}

	return fmt.Sprintf("🎉 ¡Guardado! Tu cumpleaños ha sido registrado para el %02d/%02d.", day, month), nil
}

// HandleRemoveBirthday deletes a user's birthday registration from persistence.
func (s *botService) HandleRemoveBirthday(ctx context.Context, userID int64) (string, error) {
	err := s.userRepo.RemoveBirthday(ctx, userID)
	if err != nil {
		return "❌ Tuve un problema interno al eliminar tu cumpleaños.", err
	}
	return "🗑️ Tu fecha de cumpleaños ha sido eliminada.", nil
}

// HandleGetBirthdaysOfMonth retrieves and formats the list of birthdays for the current month.
func (s *botService) HandleGetBirthdaysOfMonth(ctx context.Context) (string, error) {
	loc, err := time.LoadLocation("America/Caracas")
	if err != nil {
		loc = time.UTC
	}
	now := time.Now().In(loc)
	month := int(now.Month())

	users, err := s.userRepo.GetBirthdaysByMonth(ctx, month)
	if err != nil {
		return "❌ Tuve un problema interno al buscar los cumpleaños.", err
	}

	if len(users) == 0 {
		return fmt.Sprintf("📅 No hay cumpleaños registrados para el mes de %s.", translateMonth(month)), nil
	}

	// Sort users chronologically by birthday_day
	sort.Slice(users, func(i, j int) bool {
		di := 0
		dj := 0
		if users[i].BirthdayDay != nil {
			di = *users[i].BirthdayDay
		}
		if users[j].BirthdayDay != nil {
			dj = *users[j].BirthdayDay
		}
		return di < dj
	})

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("📅 <b>Cumpleaños de %s:</b>\n\n", translateMonth(month)))
	for _, u := range users {
		dayVal := 0
		if u.BirthdayDay != nil {
			dayVal = *u.BirthdayDay
		}

		var name string
		if u.Username != "" {
			name = "@" + u.Username
		} else {
			name = fmt.Sprintf("<a href=\"tg://user?id=%d\">%s</a>", u.ID, u.FirstName)
		}
		sb.WriteString(fmt.Sprintf("🎂 Día %02d: %s\n", dayVal, name))
	}

	return sb.String(), nil
}

// HandleTodayBirthdays checks for birthdays today and builds a celebration message if any.
func (s *botService) HandleTodayBirthdays(ctx context.Context) (string, error) {
	loc, err := time.LoadLocation("America/Caracas")
	if err != nil {
		loc = time.UTC
	}
	now := time.Now().In(loc)
	day := now.Day()
	month := int(now.Month())

	users, err := s.userRepo.GetBirthdaysByDayAndMonth(ctx, day, month)
	if err != nil {
		return "", err
	}

	if len(users) == 0 {
		return "", nil // No birthdays today
	}

	var mentions []string
	for _, u := range users {
		if u.Username != "" {
			mentions = append(mentions, "@"+u.Username)
		} else {
			mentions = append(mentions, fmt.Sprintf("<a href=\"tg://user?id=%d\">%s</a>", u.ID, u.FirstName))
		}
	}

	var msg string
	if len(users) == 1 {
		msg = fmt.Sprintf("🎂 ¡Hoy es un día especial en Golang Venezuela! Queremos desearle un muy feliz cumpleaños a %s. ¡Que tengas un excelente día programando y celebrando! 🥳🎉", mentions[0])
	} else {
		listStr := strings.Join(mentions[:len(mentions)-1], ", ") + " y " + mentions[len(mentions)-1]
		msg = fmt.Sprintf("🎂 ¡Hoy es un día especial en Golang Venezuela! Queremos desearle un muy feliz cumpleaños a %s. ¡Que tengan un excelente día programando y celebrando! 🥳🎉", listStr)
	}

	return msg, nil
}

// parseBirthday validates and extracts day and month from a "DD/MM" formatted string.
func parseBirthday(dateStr string) (int, int, error) {
	dateStr = strings.TrimSpace(dateStr)
	parts := strings.Split(dateStr, "/")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("formato incorrecto")
	}

	day, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, fmt.Errorf("día inválido")
	}

	month, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, fmt.Errorf("mes inválido")
	}

	if month < 1 || month > 12 {
		return 0, 0, fmt.Errorf("el mes debe estar entre 1 y 12")
	}

	daysInMonth := []int{0, 31, 29, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31}
	if day < 1 || day > daysInMonth[month] {
		return 0, 0, fmt.Errorf("el día %d no es válido para el mes %d", day, month)
	}

	return day, month, nil
}

// translateMonth converts a month number into its Spanish name.
func translateMonth(month int) string {
	months := []string{"", "Enero", "Febrero", "Marzo", "Abril", "Mayo", "Junio", "Julio", "Agosto", "Septiembre", "Octubre", "Noviembre", "Diciembre"}
	if month < 1 || month > 12 {
		return ""
	}
	return months[month]
}
