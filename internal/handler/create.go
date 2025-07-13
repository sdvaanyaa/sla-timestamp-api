package handler

import (
	"errors"
	"github.com/gofiber/fiber/v2"
	"github.com/sdvaanyaa/sla-timestamp-api/internal/entity"
	"github.com/sdvaanyaa/sla-timestamp-api/internal/service"
)

// Create godoc
// Create creates a new timestamp.
//
//	@Summary		Create a timestamp
//	@Description	Create a new timestamp entry
//	@Tags			timestamps
//	@Accept			json
//	@Produce		json
//	@Param			body	body		entity.CreateTimestampRequest	true	"Timestamp body"
//	@Success		201		{object}	map[string]uuid.UUID
//	@Failure		400		{object}	map[string]string	"Invalid input"
//	@Failure		500		{object}	map[string]string	"Internal error"
//	@Router			/timestamps [post]
func (h *TimestampHandler) Create(c *fiber.Ctx) error {
	var req entity.CreateTimestampRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid JSON"})
	}

	ts := req.ToTimestamp()
	id, err := h.svc.Create(c.Context(), ts)
	if err != nil {
		status := fiber.StatusInternalServerError
		if errors.Is(err, service.ErrInvalidInput) {
			status = fiber.StatusBadRequest
		}
		return c.Status(status).JSON(fiber.Map{"error": err})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"id": id.String()})
}
