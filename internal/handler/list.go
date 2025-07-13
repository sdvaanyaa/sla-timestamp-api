package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/sdvaanyaa/sla-timestamp-api/internal/entity"
	"github.com/sdvaanyaa/sla-timestamp-api/internal/service"
	"strconv"
	"time"
)

// List godoc
// List lists timestamps with pagination.
//
//	@Summary		List timestamps
//	@Description	Retrieve a list of timestamps with optional pagination
//	@Tags			timestamps
//	@Produce		json
//	@Param			limit			query		int		false	"Limit"		default(10)
//	@Param			offset			query		int		false	"Offset"	default(0)
//	@Param			external_id		query		string	false	"External ID"
//	@Param			tag				query		string	false	"Tag"						Enums(incident, sla, deployment, maintenance, alert)
//	@Param			stage			query		string	false	"Stage"						Enums(created, acknowledged, in_progress, resolved, closed)
//	@Param			timestamp_from	query		string	false	"Timestamp from (RFC3339)"	example(2025-07-10T00:00:00Z)
//	@Param			timestamp_to	query		string	false	"Timestamp to (RFC3339)"	example(2025-07-13T00:00:00Z)
//	@Param			meta_filter		query		string	false	"Meta filter as JSON"		example({"source":"email"})
//	@Success		200				{array}		entity.Timestamp
//	@Failure		400				{object}	map[string]string	"Invalid input"
//	@Failure		500				{object}	map[string]string	"Internal error"
//	@Router			/timestamps [get]
func (h *TimestampHandler) List(c *fiber.Ctx) error {
	params, err := parseListQueryParams(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err})
	}

	list, err := h.svc.List(
		c.Context(),
		params.Limit,
		params.Offset,
		params.ExternalID,
		params.Tag,
		params.Stage,
		params.TimestampFrom,
		params.TimestampTo,
		params.MetaFilter,
	)

	if err != nil {
		if errors.Is(err, service.ErrInvalidInput) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err})
	}

	return c.Status(fiber.StatusOK).JSON(list)
}

func parseListQueryParams(c *fiber.Ctx) (*entity.ListQueryParams, error) {
	limit, err := parseIntQuery(c, "limit", "10", 1)
	if err != nil {
		return nil, err
	}

	offset, err := parseIntQuery(c, "offset", "0", 0)
	if err != nil {
		return nil, err
	}

	timestampFrom, err := parseTimeQuery(c, "timestamp_from")
	if err != nil {
		return nil, err
	}

	timestampTo, err := parseTimeQuery(c, "timestamp_to")
	if err != nil {
		return nil, err
	}

	metaFilterStr := c.Query("meta_filter")
	var metaFilter map[string]any
	if metaFilterStr != "" {
		if err = json.Unmarshal([]byte(metaFilterStr), &metaFilter); err != nil {
			return nil, fmt.Errorf("invalid meta_filter")
		}
	}

	return &entity.ListQueryParams{
		Limit:         limit,
		Offset:        offset,
		ExternalID:    c.Query("external_id"),
		Tag:           c.Query("tag"),
		Stage:         c.Query("stage"),
		TimestampFrom: timestampFrom,
		TimestampTo:   timestampTo,
		MetaFilter:    metaFilter,
	}, nil
}

func parseIntQuery(c *fiber.Ctx, key, defaultVal string, min int) (int, error) {
	str := c.Query(key, defaultVal)
	val, err := strconv.Atoi(str)
	if err != nil || val < min {
		return 0, fmt.Errorf("invalid %s", key)
	}
	return val, nil
}

func parseTimeQuery(c *fiber.Ctx, key string) (*time.Time, error) {
	str := c.Query(key)
	if str == "" {
		return nil, nil
	}
	t, err := time.Parse(time.RFC3339, str)
	if err != nil {
		return nil, fmt.Errorf("invalid %s", key)
	}
	return &t, nil
}
