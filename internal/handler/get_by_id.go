package handler

import (
	"errors"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/sdvaanyaa/sla-timestamp-api/internal/repository"
)

// GetByID godoc
// GetByID gets a timestamp by ID.
//
//	@Summary		Get timestamp by ID
//	@Description	Retrieve a timestamp entry by its ID
//	@Tags			timestamps
//	@Produce		json
//	@Param			id	path		string	true	"Timestamp ID"
//	@Success		200	{object}	entity.Timestamp
//	@Failure		400	{object}	map[string]string	"Invalid ID"
//	@Failure		404	{object}	map[string]string	"Not found"
//	@Failure		500	{object}	map[string]string	"Internal error"
//	@Router			/timestamps/{id} [get]
func (h *TimestampHandler) GetByID(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid ID"})
	}

	ts, err := h.svc.GetByID(c.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "timestamp not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err})
	}

	return c.Status(fiber.StatusOK).JSON(ts)
}
