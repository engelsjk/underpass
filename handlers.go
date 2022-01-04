package underpass

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"github.com/engelsjk/underpass/dbosm"
	"github.com/gofiber/fiber/v2"
)

type Handlers struct {
	queries *dbosm.Queries
}

func NewHandlers(db *sql.DB) *Handlers {
	h := &Handlers{}
	h.queries = dbosm.New(db)
	return h
}

// Highways

func (h *Handlers) QueryHighwaysByID(c *fiber.Ctx) error {
	id := c.Params("id")
	return listByID(c, id, "highways", h.queries.ListByID)
}

func (h *Handlers) QueryHighwaysByBoundingBox(c *fiber.Ctx) error {
	bbox := c.Params("bbox")
	return listByBbox(c, bbox, "highways", h.queries.ListByBoundingBox)
}

// Buildings

func (h *Handlers) QueryBuildingsByID(c *fiber.Ctx) error {
	id := c.Params("id")
	return listByID(c, id, "highways", h.queries.ListByID)
}

func (h *Handlers) QueryBuildingsByBoundingBox(c *fiber.Ctx) error {
	bbox := c.Params("bbox")
	return listByBbox(c, bbox, "highways", h.queries.ListByBoundingBox)
}

// funcs

func listByID(
	c *fiber.Ctx,
	i string,
	t string,
	f func(ctx context.Context, arg dbosm.ListByIDParams) ([]interface{}, error),
) error {

	id, err := strconv.Atoi(i)
	if err != nil {
		return statusError(c, err)
	}

	var args dbosm.ListByIDParams
	switch t {
	case "highways":
		args = dbosm.ListByIDParams{
			OsmID: int64(id),
			Geom1: "ST_LineString",
			Geom2: "ST_MultiLineString",
			Tag:   "highway",
		}
	case "buildings":
		args = dbosm.ListByIDParams{
			OsmID: int64(id),
			Geom1: "ST_Polygon",
			Geom2: "ST_MultiPolygon",
			Tag:   "building",
		}
	default:
		return statusError(c, fmt.Errorf("type %s not recognized", t))
	}

	rec, err := f(c.Context(), args)
	if err != nil {
		return statusError(c, err)
	}

	c.Append("Content-Type", "application/json")
	return c.SendString(stringify(rec))
}

func listByBbox(
	c *fiber.Ctx,
	b string,
	t string,
	f func(ctx context.Context, arg dbosm.ListByBoundingBoxParams) ([]interface{}, error),
) error {

	bb := strings.Split(b, ",")

	lowLeftLon, err := strconv.ParseFloat(bb[0], 64)
	if err != nil {
		return statusError(c, err)
	}
	lowLeftLat, err := strconv.ParseFloat(bb[1], 64)
	if err != nil {
		return statusError(c, err)
	}
	upRightLon, err := strconv.ParseFloat(bb[2], 64)
	if err != nil {
		return statusError(c, err)
	}
	upRightLat, err := strconv.ParseFloat(bb[3], 64)
	if err != nil {
		return statusError(c, err)
	}

	args := dbosm.ListByBoundingBoxParams{
		LowLeftLon: lowLeftLon,
		LowLeftLat: lowLeftLat,
		UpRightLon: upRightLon,
		UpRightLat: upRightLat,
	}

	switch t {
	case "highways":
		args.Geom1 = "ST_LineString"
		args.Geom2 = "ST_MultiLineString"
		args.Tag = "highway"
	case "buildings":
		args.Geom1 = "ST_Polygon"
		args.Geom2 = "ST_MultiPolygon"
		args.Tag = "building"
	default:
		return statusError(c, fmt.Errorf("type %s not recognized", t))
	}

	rec, err := f(c.Context(), args)
	if err != nil {
		return statusError(c, err)
	}

	c.Append("Content-Type", "application/json")
	return c.SendString(stringify(rec))
}

func statusError(c *fiber.Ctx, err error) error {
	return c.Status(500).JSON(&fiber.Map{
		"success": false,
		"error":   err,
	})
}
