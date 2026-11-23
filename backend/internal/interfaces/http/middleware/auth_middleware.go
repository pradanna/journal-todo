package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"

	"jurnal-be/internal/infrastructure/authutil"
)

const ContextUserIDKey = "user_id"

func RequireAuth(secret string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		header := c.Get("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "token tidak ada"})
		}
		raw := strings.TrimPrefix(header, "Bearer ")

		claims, err := authutil.ParseAccessToken(raw, secret)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "token tidak valid"})
		}

		c.Locals(ContextUserIDKey, claims.UserID)
		return c.Next()
	}
}
