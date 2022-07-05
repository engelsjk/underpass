package underpass

import (
	"context"
	"encoding/json"
	"errors"
	"net/url"
	"strconv"
	"strings"

	"github.com/engelsjk/underpass/dbosm"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4/pgxpool"
)

var (
	ErrDatabase                 error = errors.New("database error")
	ErrInvalidID                error = errors.New("invalid id")
	ErrInvalidQuery             error = errors.New("invalid query")
	ErrMissingElementQueryParam error = errors.New("missing element query parameter")
	ErrInvalidType              error = errors.New("invalid type")
	ErrInvalidBoundingBox       error = errors.New("invalid bbox")
	ErrInvalidTagList           error = errors.New("invalid tag list")
)

type handlers struct {
	queries *dbosm.Queries
}

func newHandlers(db *pgxpool.Pool) *handlers {
	h := &handlers{}
	h.queries = dbosm.New(db)
	return h
}

/////////////////////////////////////////////////////////////
/////////////////////////////////////////////////////////////
// legacy handlers

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
	bboxParam := c.Params("bbox")
	tagQueryParam := c.Query("tag")
	return listByBbox(c, bboxParam, tagQueryParam, h.queries.ListByBoundingBox)
}

func (h *handlers) QueryFeaturesByWikidataID(c *fiber.Ctx) error {
	idParam := c.Params("id")
	return listByKey(c, "wikidata", idParam, h.queries.ListByKey)
}

func (h *handlers) QueryFeaturesByWikipediaName(c *fiber.Ctx) error {
	nameParam := c.Params("name")
	return listByKey(c, "wikipedia", nameParam, h.queries.ListByKey)
}

/////////////////////////////////////////////////////////////
/////////////////////////////////////////////////////////////
// v1 handlers

func (h *handlers) QueryFeatures(c *fiber.Ctx) error {
	bbox := c.Query("bbox")
	if bbox != "" {
		tag := c.Query("tag")
		return listByBbox(c, bbox, tag, h.queries.ListByBoundingBox)
	}

	wikidata := c.Query("wikidata")
	if wikidata != "" {
		return listByKey(c, "wikidata", wikidata, h.queries.ListByKey)
	}

	wikipedia := c.Query("wikipedia")
	if wikipedia != "" {
		return listByKey(c, "wikipedia", wikipedia, h.queries.ListByKey)
	}

	return statusError(c, ErrInvalidQuery)
}

func (h *handlers) QueryFeatureByID(c *fiber.Ctx) error {
	id := c.Params("id")
	element := c.Query("element")
	switch element {
	case "":
		return statusError(c, ErrMissingElementQueryParam)
	case "node":
		return listByID(c, id, "nodes", h.queries.ListByID)
	case "way":
		return listByID(c, id, "ways", h.queries.ListByID)
	case "relation":
		return listByID(c, id, "relations", h.queries.ListByID)
	default:
		return statusError(c, ErrInvalidQuery)
	}
}

func (h *handlers) InvalidQuery(c *fiber.Ctx) error {
	return statusError(c, ErrInvalidQuery)
}

/////////////////////////////////////////////////////////////
/////////////////////////////////////////////////////////////
// funcs

func listByID(
	c *fiber.Ctx,
	idParam string,
	featureType string,
	f func(ctx context.Context, params dbosm.ListByIDParams) ([]pgtype.JSON, error),
) error {

	id, err := strconv.Atoi(idParam)
	if err != nil {
		return statusError(c, ErrInvalidID)
	}

	params := dbosm.ListByIDParams{}
	switch featureType {
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
	bbox string,
	tag string,
	f func(ctx context.Context, arg dbosm.ListByBoundingBoxParams) ([]pgtype.JSON, error),
) error {

	// parse bbox
	bb := strings.Split(bbox, ",")

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

	if tag != "" {
		tags := map[string][]string{}
		if err := json.Unmarshal([]byte(tag), &tags); err != nil {
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

func listByKey(
	c *fiber.Ctx,
	key string,
	val string,
	f func(ctx context.Context, params dbosm.ListByKeyParams) ([]pgtype.JSON, error),
) error {

	k := key

	v, err := url.QueryUnescape(val)
	if err != nil {
		return statusError(c, ErrInvalidQuery)
	}

	args := dbosm.ListByKeyParams{
		Key: k,
		Val: v,
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
