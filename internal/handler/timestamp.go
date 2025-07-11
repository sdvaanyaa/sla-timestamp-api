package handler

import (
	"errors"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/sdvaanyaa/sla-timestamp-api/internal/entity"
	"github.com/sdvaanyaa/sla-timestamp-api/internal/repository"
	"github.com/sdvaanyaa/sla-timestamp-api/internal/service"
	"strconv"
)

type TimestampHandler struct {
	svc service.TimestampService
}

func New(app *fiber.App, svc service.TimestampService) {
	h := &TimestampHandler{svc: svc}
	app.Post("/timestamps", h.Create)
	app.Get("timestamps/:id", h.GetByID)
	app.Get("/timestamps", h.List)
	app.Delete("/timestamps/:id", h.Delete)
}

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

// List godoc
// List lists timestamps with pagination.
//
//	@Summary		List timestamps
//	@Description	Retrieve a list of timestamps with optional pagination
//	@Tags			timestamps
//	@Produce		json
//	@Param			limit	query		int	false	"Limit"		default(10)
//	@Param			offset	query		int	false	"Offset"	default(0)
//	@Success		200		{array}		entity.Timestamp
//	@Failure		500		{object}	map[string]string	"Internal error"
//	@Router			/timestamps [get]
func (h *TimestampHandler) List(c *fiber.Ctx) error {
	limitStr := c.Query("limit", "10")
	offsetStr := c.Query("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid limit"})
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid offset"})
	}

	list, err := h.svc.List(c.Context(), limit, offset)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err})
	}

	return c.Status(fiber.StatusOK).JSON(list)
}

// Delete godoc
// Delete deletes a timestamp by ID.
//
//	@Summary		Delete timestamp
//	@Description	Delete a timestamp entry by its ID
//	@Tags			timestamps
//	@Param			id	path		string				true	"Timestamp ID"
//	@Success		204	{string}	string				"No content"
//	@Failure		400	{object}	map[string]string	"Invalid ID"
//	@Failure		500	{object}	map[string]string	"Internal error"
//	@Router			/timestamps/{id} [delete]
func (h *TimestampHandler) Delete(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid ID"})
	}

	err = h.svc.Delete(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err})
	}

	return c.SendStatus(fiber.StatusNoContent)
}
