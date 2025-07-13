package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sdvaanyaa/sla-timestamp-api/internal/service"
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
