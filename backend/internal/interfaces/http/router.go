package http

import (
	"github.com/gofiber/fiber/v2"

	"jurnal-be/internal/interfaces/http/handlers"
	"jurnal-be/internal/interfaces/http/middleware"
)

type Handlers struct {
	Auth      *handlers.AuthHandler
	Entry     *handlers.EntryHandler
	Workspace *handlers.WorkspaceHandler
}

func RegisterRoutes(app *fiber.App, h Handlers, jwtSecret string) {
	app.Get("/health", handlers.HealthCheck)

	auth := app.Group("/auth")
	auth.Post("/google", h.Auth.GoogleLogin)
	auth.Post("/register", h.Auth.Register)
	auth.Post("/login", h.Auth.PasswordLogin)
	auth.Post("/refresh", h.Auth.Refresh)
	auth.Post("/logout", h.Auth.Logout)

	api := app.Group("/api", middleware.RequireAuth(jwtSecret))

	entries := api.Group("/entries")
	entries.Post("/", h.Entry.Create)
	entries.Get("/", h.Entry.List)
	entries.Get("/:id", h.Entry.Get)
	entries.Delete("/:id", h.Entry.Delete)

	workspaces := api.Group("/workspaces")
	workspaces.Get("/", h.Workspace.List)
	workspaces.Post("/", h.Workspace.Create)
	workspaces.Delete("/:id", h.Workspace.Delete)
}
