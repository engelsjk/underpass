package underpass

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/engelsjk/underpass/dbosm"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4/pgxpool"
)

var (
	ErrDatabase           error = errors.New("database error")
	ErrInvalidID          error = errors.New("invalid id")
	ErrInvalidType        error = errors.New("invalid type")
	ErrInvalidBoundingBox error = errors.New("invalid bbox")
	ErrInvalidTagList     error = errors.New("invalid tag list")
)

type handlers struct {
	queries *dbosm.Queries
}

func newHandlers(db *pgxpool.Pool) *handlers {
	h := &handlers{}
	h.queries = dbosm.New(db)
	return h
}

// Handler queries

func (h *handlers) QueryNodesByID(c *fiber.Ctx) error {
	id := c.Params("id")
	return listByID(c, id, "nodes", h.queries.ListByID)
}

func (h *handlers) QueryWaysByID(c *fiber.Ctx) error {
	id := c.Params("id")
	return listByID(c, id, "ways", h.queries.ListByID)
}

func (h *handlers) QueryRelationsByID(c *fiber.Ctx) error {
	id := c.Params("id")
	return listByID(c, id, "relations", h.queries.ListByID)
}

func (h *handlers) QueryBboxByBbox(c *fiber.Ctx) error {
	bbox := c.Params("bbox")
	tag := c.Query("tag")
	return listByBbox(c, bbox, tag, h.queries.ListByBoundingBox)
}

// funcs

func listByID(
	c *fiber.Ctx,
	i string,
	t string,
	f func(ctx context.Context, params dbosm.ListByIDParams) ([]pgtype.JSON, error),
) error {

	id, err := strconv.Atoi(i)
	if err != nil {
		return statusError(c, ErrInvalidID)
	}

	params := dbosm.ListByIDParams{}
	switch t {
	case "nodes":
		params.OsmID = int64(id)
		params.ElementType = "node"
	case "ways":
		params.OsmID = -int64(id)
		params.ElementType = "way"
	case "relations":
		params.OsmID = -int64(id) - 1e17
		params.ElementType = "relation"
	default:
		return statusError(c, ErrInvalidType)
	}

	rec, err := f(c.Context(), params)
	if err != nil {
		return statusError(c, ErrDatabase)
	}

	c.Append("Content-Type", "application/json")
	return c.SendString(stringifyJSONRawMessage(rec))
}

func listByBbox(
	c *fiber.Ctx,
	b string,
	t string,
	f func(ctx context.Context, arg dbosm.ListByBoundingBoxParams) ([]pgtype.JSON, error),
) error {

	// parse bbox
	bb := strings.Split(b, ",")

	lowLeftLon, err := strconv.ParseFloat(bb[0], 64)
	if err != nil {
		return statusError(c, ErrInvalidBoundingBox)
	}
	lowLeftLat, err := strconv.ParseFloat(bb[1], 64)
	if err != nil {
		return statusError(c, ErrInvalidBoundingBox)
	}
	upRightLon, err := strconv.ParseFloat(bb[2], 64)
	if err != nil {
		return statusError(c, ErrInvalidBoundingBox)
	}
	upRightLat, err := strconv.ParseFloat(bb[3], 64)
	if err != nil {
		return statusError(c, ErrInvalidBoundingBox)
	}

	// parse tag
	var key string
	var vals []string
	var isTag bool
	var isTagList bool

	if t != "" {
		tags := map[string][]string{}
		if err := json.Unmarshal([]byte(t), &tags); err != nil {
			return statusError(c, ErrInvalidTagList)
		}
		keys := make([]string, len(tags))
		i := 0
		for k := range tags {
			keys[i] = k
			i++
		}
		if len(keys) != 1 {
			return statusError(c, ErrInvalidTagList)
		}

		key = keys[0]
		vals = tags[key]

		if vals[0] == "*" {
			isTag = true
		} else {
			isTagList = true
		}
	}

	fmt.Printf("t: %s\n", t)
	fmt.Printf("key: %s\n", key)
	fmt.Printf("vals: %v\n", vals)

	args := dbosm.ListByBoundingBoxParams{
		LowLeftLon: lowLeftLon,
		LowLeftLat: lowLeftLat,
		UpRightLon: upRightLon,
		UpRightLat: upRightLat,
		Key:        key,
		Vals:       vals,
		IsTag:      isTag,
		IsTagList:  isTagList,
	}

	rec, err := f(c.Context(), args)
	if err != nil {
		return statusError(c, ErrDatabase)
	}

	c.Append("Content-Type", "application/json")
	return c.SendString(stringifyJSONRawMessage(rec))
}

func statusError(c *fiber.Ctx, err error) error {
	return c.Status(500).JSON(&fiber.Map{
		"success": false,
		"error":   err.Error(),
	})
}
