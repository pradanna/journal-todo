package handlers

import (
	"time"

	"github.com/gofiber/fiber/v2"

	"jurnal-be/internal/application"
	"jurnal-be/internal/domain"
	"jurnal-be/internal/interfaces/http/middleware"
)

type EntryHandler struct {
	entries *application.EntryService
}

func NewEntryHandler(entries *application.EntryService) *EntryHandler {
	return &EntryHandler{entries: entries}
}

func userIDFromCtx(c *fiber.Ctx) uint {
	v, _ := c.Locals(middleware.ContextUserIDKey).(uint)
	return v
}

type createEntryItemRequest struct {
	Type     string `json:"type"`
	Content  string `json:"content"`
	Metadata string `json:"metadata"`
}

type createEntryRequest struct {
	WorkspaceID *uint                    `json:"workspace_id"`
	EntryDate   string                   `json:"entry_date"` // format YYYY-MM-DD
	HoursSpent  float64                  `json:"hours_spent"`
	Mood        *string                  `json:"mood"`
	Items       []createEntryItemRequest `json:"items"`
}

func (h *EntryHandler) Create(c *fiber.Ctx) error {
	var req createEntryRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "payload tidak valid"})
	}

	entryDate, err := time.Parse("2006-01-02", req.EntryDate)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "entry_date harus format YYYY-MM-DD"})
	}

	items := make([]application.CreateEntryItemInput, 0, len(req.Items))
	for _, it := range req.Items {
		items = append(items, application.CreateEntryItemInput{
			Type:     domain.EntryItemType(it.Type),
			Content:  it.Content,
			Metadata: it.Metadata,
		})
	}

	entry, err := h.entries.CreateEntry(application.CreateEntryInput{
		UserID:      userIDFromCtx(c),
		WorkspaceID: req.WorkspaceID,
		EntryDate:   entryDate,
		HoursSpent:  req.HoursSpent,
		Mood:        req.Mood,
		Items:       items,
	})
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(entry)
}

func (h *EntryHandler) List(c *fiber.Ctx) error {
	var filter domain.JournalEntryFilter

	if from := c.Query("from"); from != "" {
		if t, err := time.Parse("2006-01-02", from); err == nil {
			filter.From = &t
		}
	}
	if to := c.Query("to"); to != "" {
		if t, err := time.Parse("2006-01-02", to); err == nil {
			filter.To = &t
		}
	}

	entries, err := h.entries.ListEntries(userIDFromCtx(c), filter)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(entries)
}

func (h *EntryHandler) Get(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "id tidak valid"})
	}
	entry, err := h.entries.GetEntry(uint(id), userIDFromCtx(c))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(entry)
}

func (h *EntryHandler) Delete(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "id tidak valid"})
	}
	if err := h.entries.DeleteEntry(uint(id), userIDFromCtx(c)); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}
	return c.SendStatus(fiber.StatusNoContent)
}
