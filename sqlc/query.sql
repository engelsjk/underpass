-- name: ListByID :many
SELECT json_agg(json_build_object(
    'type',       'Feature',
    'id',         t3.element_id,
    'geometry',   ST_AsGeoJSON(t3.geometry)::json,
    'properties', (to_jsonb(t3.tags))::json
)) FROM (
    SELECT t2.*, 
    CASE
        WHEN t2.geometry_type='ST_Point' THEN CONCAT('node/',t2.osm_id)
        WHEN t2.geometry_type='ST_LineString' AND id > 0 THEN CONCAT('way/',t2.osm_id)
        WHEN t2.geometry_type='ST_Polygon' AND id > 0 THEN CONCAT('way/',t2.osm_id)
        WHEN t2.geometry_type='ST_MultiPolygon' AND id > 0 THEN CONCAT('way/',t2.osm_id)
        WHEN t2.geometry_type='ST_MultiLineString' AND id > 0 THEN CONCAT('way/',t2.osm_id)
        WHEN t2.geometry_type='ST_LineString' AND id < 0 THEN CONCAT('relation/',t2.osm_id)
        WHEN t2.geometry_type='ST_Polygon' AND id < 0 THEN CONCAT('relation/',t2.osm_id)
        WHEN t2.geometry_type='ST_MultiPolygon' AND id < 0 THEN CONCAT('relation/',t2.osm_id)
        WHEN t2.geometry_type='ST_MultiLineString' AND id < 0 THEN CONCAT('relation/',t2.osm_id)
        ELSE CONCAT('unknown/',t2.osm_id)
    END AS element_id
    FROM (
        SELECT 
        osm_id,id,geometry,tags, 
        ST_GeometryType(geometry) AS geometry_type
        FROM osm_all AS t1
        WHERE (
            t1.osm_id = @osm_id::bigint AND
            ST_GeometryType(t1.geometry) IN (@geom1::text, @geom2::text) AND
            t1.tags ? @tag::text
        )
    ) AS t2
) AS t3;

-- name: ListByBoundingBox :many
SELECT json_agg(json_build_object(
    'type',       'Feature',
    'id',         t3.element_id,
    'geometry',   ST_AsGeoJSON(t3.geometry)::json,
    'properties', (to_jsonb(t3.tags))::json
)) FROM (
    SELECT t2.*, 
    CASE
        WHEN t2.geometry_type='ST_Point' THEN CONCAT('node/',t2.osm_id)
        WHEN t2.geometry_type='ST_LineString' AND id > 0 THEN CONCAT('way/',t2.osm_id)
        WHEN t2.geometry_type='ST_Polygon' AND id > 0 THEN CONCAT('way/',t2.osm_id)
        WHEN t2.geometry_type='ST_MultiPolygon' AND id > 0 THEN CONCAT('way/',t2.osm_id)
        WHEN t2.geometry_type='ST_MultiLineString' AND id > 0 THEN CONCAT('way/',t2.osm_id)
        WHEN t2.geometry_type='ST_LineString' AND id < 0 THEN CONCAT('relation/',t2.osm_id)
        WHEN t2.geometry_type='ST_Polygon' AND id < 0 THEN CONCAT('relation/',t2.osm_id)
        WHEN t2.geometry_type='ST_MultiPolygon' AND id < 0 THEN CONCAT('relation/',t2.osm_id)
        WHEN t2.geometry_type='ST_MultiLineString' AND id < 0 THEN CONCAT('relation/',t2.osm_id)
        ELSE CONCAT('unknown/',t2.osm_id)
    END AS element_id
    FROM (
        SELECT 
        osm_id,id,geometry,tags, 
        ST_GeometryType(geometry) AS geometry_type
        FROM osm_all AS t1
        WHERE (
            ST_Intersects(
                t1.geometry,
                ST_SetSRID(
                    ST_MakeBox2D(
                        ST_Point(@low_left_lon::float,@low_left_lat::float),
                        ST_Point(@up_right_lon::float,@up_right_lat::float)
                    ),
                    4326
                )
            ) AND
            ST_GeometryType(t1.geometry) IN (@geom1::text, @geom2::text) AND
            t1.tags ? @tag::text
        )
    ) AS t2
) AS t3;



-- name: DepracatedListByID :many
SELECT ST_AsGeoJSON(t2.*)
FROM osm_all AS t
WHERE (
    t2.osm_id = @osm_id::bigint AND
    ST_GeometryType(t2.geometry) IN (@geom1::text, @geom2::text) AND
    t2.tags ? @tag::text
);

-- name: DeprecatedListByBoundingBox :many
SELECT ST_AsGeoJSON(t2.*)
FROM osm_all AS t
WHERE (
    ST_Intersects(
        t2.geometry,
        ST_SetSRID(
            ST_MakeBox2D(
                ST_Point(@low_left_lon::float,@low_left_lat::float),
                ST_Point(@up_right_lon::float,@up_right_lat::float)
            ),
            4326
        )
    ) AND
    ST_GeometryType(t2.geometry) IN (@geom1::text, @geom2::text) AND
    t2.tags ? @tag::text
);