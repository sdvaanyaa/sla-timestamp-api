package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

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
