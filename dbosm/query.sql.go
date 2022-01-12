// Code generated by sqlc. DO NOT EDIT.
// source: query.sql

package dbosm

import (
	"context"

	"github.com/jackc/pgtype"
)

const listByBoundingBox = `-- name: ListByBoundingBox :many
SELECT json_agg(json_build_object(
    'type',       'Feature',
    'id',         T3.element_id,
    'geometry',   ST_AsGeoJSON(T3.geometry)::json,
    'properties', (to_jsonb(T3.tags))::json
)) FROM (
    SELECT DISTINCT ON (T2.osm_id) t2.osm_id, t2.geometry, t2.tags, t2.key, t2.element_id
    FROM (
        SELECT 
        T1.osm_id,T1.geometry,T1.tags, 
        jsonb_object_keys(to_jsonb(T1.tags)) as key,
        CASE
             WHEN osm_id > 0 THEN CONCAT('node/', osm_id)
            WHEN osm_id < 0 AND osm_id > -1e17 THEN CONCAT('way/',-osm_id)
            WHEN osm_id < -1e17 THEN CONCAT('relation/',-(osm_id+1e17))
            ELSE 'unknown'
        END AS element_id
        FROM osm_all AS T1
        WHERE (
            ST_Intersects(
                geometry,
                ST_SetSRID(
                    ST_MakeBox2D(
                        ST_Point($1::float,$2::float),
                        ST_Point($3::float,$4::float)
                    ),4326
                )
            ) AND
            CASE
                WHEN ST_GeometryType(geometry) = 'ST_LineString' THEN ST_IsClosed(geometry) IS NOT TRUE
                ELSE TRUE
            END
        )
    ) AS T2 
    WHERE (
        CASE 
            WHEN $5::boolean THEN tags ? $6::text
            WHEN $7::boolean THEN tags -> $6::text = ANY($8::text[])
        ELSE TRUE
        END
    )
) AS T3
`

type ListByBoundingBoxParams struct {
	LowLeftLon float64
	LowLeftLat float64
	UpRightLon float64
	UpRightLat float64
	IsTag      bool
	Key        string
	IsTagList  bool
	Vals       []string
}

func (q *Queries) ListByBoundingBox(ctx context.Context, arg ListByBoundingBoxParams) ([]pgtype.JSON, error) {
	rows, err := q.db.Query(ctx, listByBoundingBox,
		arg.LowLeftLon,
		arg.LowLeftLat,
		arg.UpRightLon,
		arg.UpRightLat,
		arg.IsTag,
		arg.Key,
		arg.IsTagList,
		arg.Vals,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []pgtype.JSON
	for rows.Next() {
		var json_agg pgtype.JSON
		if err := rows.Scan(&json_agg); err != nil {
			return nil, err
		}
		items = append(items, json_agg)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const listByID = `-- name: ListByID :many
SELECT json_build_object(
    'type',       'Feature',
    'id',         T3.element_id,
    'geometry',   ST_AsGeoJSON(T3.geometry)::json,
    'properties', (to_jsonb(T3.tags))::json
) FROM (
    SELECT t2.osm_id, t2.geometry, t2.tags, t2.element_id, t2.element_type
    FROM (
        SELECT 
        T1.osm_id,T1.geometry,T1.tags, 
        CASE
            WHEN osm_id > 0 THEN CONCAT('node/', osm_id)
            WHEN osm_id < 0 AND osm_id > -1e17 THEN CONCAT('way/',-osm_id)
            WHEN osm_id < -1e17 THEN CONCAT('relation/',-(osm_id+1e17))
            ELSE 'unknown'
        END AS element_id,
        CASE
            WHEN osm_id > 0 THEN 'node'
            WHEN osm_id < 0 AND osm_id > -1e17 THEN 'way'
            WHEN osm_id < -1e17 THEN 'relation'
            ELSE 'unknown'
        END AS element_type
        FROM osm_all AS T1
        WHERE (
            osm_id = $1::bigint
        )
    ) AS T2 
    WHERE (
        element_type = $2::text AND
        CASE
            WHEN ST_GeometryType(geometry) = 'ST_LineString' THEN ST_IsClosed(geometry) IS NOT TRUE
            ELSE TRUE
        END
    )
) AS T3
`

type ListByIDParams struct {
	OsmID       int64
	ElementType string
}

func (q *Queries) ListByID(ctx context.Context, arg ListByIDParams) ([]pgtype.JSON, error) {
	rows, err := q.db.Query(ctx, listByID, arg.OsmID, arg.ElementType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []pgtype.JSON
	for rows.Next() {
		var json_build_object pgtype.JSON
		if err := rows.Scan(&json_build_object); err != nil {
			return nil, err
		}
		items = append(items, json_build_object)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
