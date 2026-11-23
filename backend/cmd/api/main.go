package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"

	"jurnal-be/internal/application"
	"jurnal-be/internal/config"
	httpif "jurnal-be/internal/interfaces/http"
	"jurnal-be/internal/interfaces/http/handlers"
	"jurnal-be/internal/infrastructure/postgres"
)

func main() {
	_ = godotenv.Load() // abaikan kalau .env gak ada (misal di production pakai env asli)
	cfg := config.Load()

	db, err := postgres.Connect(cfg)
	if err != nil {
		log.Fatalf("gagal konek database: %v", err)
	}

	userRepo := postgres.NewUserRepository(db)
	workspaceRepo := postgres.NewWorkspaceRepository(db)
	entryRepo := postgres.NewJournalEntryRepository(db)
	refreshTokenRepo := postgres.NewRefreshTokenRepository(db)

	authService := application.NewAuthService(userRepo, refreshTokenRepo, cfg.GoogleClientID, cfg.JWTAccessSecret, cfg.JWTRefreshSecret)
	entryService := application.NewEntryService(entryRepo, workspaceRepo)

	h := httpif.Handlers{
		Auth:      handlers.NewAuthHandler(authService),
		Entry:     handlers.NewEntryHandler(entryService),
		Workspace: handlers.NewWorkspaceHandler(workspaceRepo),
	}

	app := fiber.New()
	httpif.RegisterRoutes(app, h, cfg.JWTAccessSecret)

	log.Printf("jurnal-api jalan di port %s", cfg.AppPort)
	if err := app.Listen(":" + cfg.AppPort); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
