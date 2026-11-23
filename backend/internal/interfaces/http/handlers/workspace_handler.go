package handlers

import (
	"github.com/gofiber/fiber/v2"

	"jurnal-be/internal/domain"
	"jurnal-be/internal/interfaces/http/middleware"
)

type WorkspaceHandler struct {
	workspaces domain.WorkspaceRepository
}

func NewWorkspaceHandler(workspaces domain.WorkspaceRepository) *WorkspaceHandler {
	return &WorkspaceHandler{workspaces: workspaces}
}

func (h *WorkspaceHandler) List(c *fiber.Ctx) error {
	userID, _ := c.Locals(middleware.ContextUserIDKey).(uint)
	ws, err := h.workspaces.FindAllByUser(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(ws)
}

type createWorkspaceRequest struct {
	Name string `json:"name"`
}

func (h *WorkspaceHandler) Create(c *fiber.Ctx) error {
	userID, _ := c.Locals(middleware.ContextUserIDKey).(uint)
	var req createWorkspaceRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "payload tidak valid"})
	}
	ws := &domain.Workspace{UserID: userID, Name: req.Name}
	if err := h.workspaces.Create(ws); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(ws)
}

func (h *WorkspaceHandler) Delete(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "id tidak valid"})
	}
	if err := h.workspaces.Delete(uint(id)); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}
	return c.SendStatus(fiber.StatusNoContent)
}
