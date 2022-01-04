CREATE TABLE public.osm_all (
    id integer NOT NULL,
    osm_id bigint NOT NULL,
    tags public.hstore,
    geometry public.geometry(Geometry,4326)
);