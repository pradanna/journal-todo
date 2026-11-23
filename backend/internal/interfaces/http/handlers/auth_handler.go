package handlers

import (
	"github.com/gofiber/fiber/v2"

	"jurnal-be/internal/application"
)

type AuthHandler struct {
	auth *application.AuthService
}

func NewAuthHandler(auth *application.AuthService) *AuthHandler {
	return &AuthHandler{auth: auth}
}

type googleLoginRequest struct {
	IDToken    string `json:"id_token"`
	DeviceInfo string `json:"device_info"`
}

func (h *AuthHandler) GoogleLogin(c *fiber.Ctx) error {
	var req googleLoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "payload tidak valid"})
	}
	pair, err := h.auth.LoginWithGoogle(c.Context(), req.IDToken, req.DeviceInfo)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"access_token": pair.AccessToken, "refresh_token": pair.RefreshToken})
}

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req registerRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "payload tidak valid"})
	}
	user, err := h.auth.RegisterWithPassword(req.Email, req.Password, req.Name)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"id": user.ID, "email": user.Email})
}

type passwordLoginRequest struct {
	Email      string `json:"email"`
	Password   string `json:"password"`
	DeviceInfo string `json:"device_info"`
}

func (h *AuthHandler) PasswordLogin(c *fiber.Ctx) error {
	var req passwordLoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "payload tidak valid"})
	}
	pair, err := h.auth.LoginWithPassword(req.Email, req.Password, req.DeviceInfo)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"access_token": pair.AccessToken, "refresh_token": pair.RefreshToken})
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

func (h *AuthHandler) Refresh(c *fiber.Ctx) error {
	var req refreshRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "payload tidak valid"})
	}
	access, err := h.auth.RefreshAccessToken(req.RefreshToken)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"access_token": access})
}

func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	var req refreshRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "payload tidak valid"})
	}
	_ = h.auth.Logout(req.RefreshToken)
	return c.SendStatus(fiber.StatusNoContent)
}
