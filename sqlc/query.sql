-- name: ListByID :many
SELECT ST_AsGeoJSON(t.*)
FROM osm_all AS t
WHERE (
    t.osm_id = @osm_id::bigint AND
    ST_GeometryType(t.geometry) IN (@geom1::text, @geom2::text) AND
    t.tags ? @tag::text
);

-- name: ListByBoundingBox :many
SELECT ST_AsGeoJSON(t.*)
FROM osm_all AS t
WHERE (
    ST_Intersects(
        t.geometry,
        ST_SetSRID(
            ST_MakeBox2D(
                ST_Point(@low_left_lon::float,@low_left_lat::float),
                ST_Point(@up_right_lon::float,@up_right_lat::float)
            ),
            4326
        )
    ) AND
    ST_GeometryType(t.geometry) IN (@geom1::text, @geom2::text) AND
    t.tags ? @tag::text
);