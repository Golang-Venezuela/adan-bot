// Package main provides the AWS Lambda entry point for the Adan Bot application.
// It initializes the backend dependencies, such as the Turso Database and the Telegram Bot API client,
// and sets up an AWS API Gateway proxy handler that synchronously processes incoming Telegram webhook updates.
package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"github.com/Golang-Venezuela/adan-bot/internal/adapters/delivery/telegram"
	"github.com/Golang-Venezuela/adan-bot/internal/adapters/repository/sqlite"
	"github.com/Golang-Venezuela/adan-bot/internal/core/services"
	"github.com/Golang-Venezuela/adan-bot/internal/infra/config"
	_ "github.com/Golang-Venezuela/adan-bot/internal/infra/logger"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"

	tele "gopkg.in/telebot.v4"
)

// botHandler is a globalized HTTP handler instance.
// It is kept in the global scope to leverage AWS Lambda execution model,
// allowing the handler to be reused across warm starts.
var botHandler http.Handler

// init acts as the cold-start initialization mechanism for the Lambda function.
// It provisions the bot token, establishes the database connection,
// configures the Telegram bot instance synchronously, and maps the HTTP routes.
func init() {
	apiToken := strings.Trim(config.Getenv("TELEGRAM_API_TOKEN", ""), `"`)
	if apiToken == "" {
		slog.Error("Missing TELEGRAM_API_TOKEN in environment")
		return
	}

	// Retrieve the Turso DSN environment variable for serverless deployment.
	dbUrl := config.Getenv("TURSO_DB_URL", "file:/tmp/local.db")
	userRepo, err := sqlite.NewUserRepository(dbUrl)
	if err != nil {
		slog.Error("Could not initialize Turso DB on cold start", slog.Any("error", err))
		return
	}

	botSvc := services.NewBotService(userRepo)

	// Configure the Telegram bot settings for a serverless execution environment.
	pref := tele.Settings{
		Token:       apiToken,
		// Synchronous must be true because serverless functions freeze their execution context
		// immediately after the handler returns, potentially killing background goroutines abruptly.
		Synchronous: true,
	}

	bot, err := tele.NewBot(pref)
	if err != nil {
		slog.Error("Could not initialize Telegram bot", slog.Any("error", err))
		return
	}

	telegram.RegisterHandlers(bot, botSvc)

	// Configure a custom HTTP handler that decodes the incoming Telegram webhook payload
	// and triggers synchronous processing. We avoid using bot.Start() or tele.Webhook
	// because the standard Poller initializes channels and goroutines that are not well-suited for Lambda.
	botHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var update tele.Update
		if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
			slog.Error("Could not decode update", slog.Any("error", err))
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}
		bot.ProcessUpdate(update)
		w.WriteHeader(http.StatusOK)
	})
}

// HandleRequest serves as the primary AWS Lambda handler.
// It wraps our custom HTTP handler using an API Gateway adapter to process incoming webhook events.
func HandleRequest(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return httpadapter.New(botHandler).ProxyWithContext(ctx, req)
}

// main registers the HandleRequest function to the AWS Lambda runtime execution engine.
func main() {
	slog.Info("Starting AWS Lambda Handler")
	lambda.Start(HandleRequest)
}
