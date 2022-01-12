-- name: ListByID :many
SELECT json_build_object(
    'type',       'Feature',
    'id',         T3.element_id,
    'geometry',   ST_AsGeoJSON(T3.geometry)::json,
    'properties', (to_jsonb(T3.tags))::json
) FROM (
    SELECT T2.*
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
            osm_id = @osm_id::bigint
        )
    ) AS T2 
    WHERE (
        element_type = @element_type::text AND
        CASE
            WHEN ST_GeometryType(geometry) = 'ST_LineString' THEN ST_IsClosed(geometry) IS NOT TRUE
            ELSE TRUE
        END
    )
) AS T3;

-- name: ListByBoundingBox :many
SELECT json_agg(json_build_object(
    'type',       'Feature',
    'id',         T3.element_id,
    'geometry',   ST_AsGeoJSON(T3.geometry)::json,
    'properties', (to_jsonb(T3.tags))::json
)) FROM (
    SELECT DISTINCT ON (T2.osm_id) T2.*
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
                        ST_Point(@low_left_lon::float,@low_left_lat::float),
                        ST_Point(@up_right_lon::float,@up_right_lat::float)
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
            WHEN @is_tag::boolean THEN tags ? @key::text
            WHEN @is_tag_list::boolean THEN tags -> @key::text = ANY(@vals::text[])
        ELSE TRUE
        END
    )
) AS T3;

