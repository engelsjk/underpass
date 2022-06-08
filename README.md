# Underpass

An API built on top of a live, local OpenStreetMap (OSM) instance that provides custom and fast queries of selected map data. Intended as a quicker alternative to the [Overpass API](https://wiki.openstreetmap.org/wiki/Overpass_API).

## Queries

By default, Underpass includes only two SQL queries to the Docker-OSM database:

* ListByID: a list of features by OSM node/way/relation ID
* ListByBoundingBox: a list of features by a lower-left / upper-right coordinate pair bounding box

## Setup

The Underpass setup consists of two phases:

* First, spin up a live instance of Docker-OSM for your area of interest using an Underpass-specific database schema
* Then start an Underpass server that connects to the running Docker-OSM instance

### Docker-OSM Setup

Underpass requires a running instance of Docker-OSM ([kartoza/docker-osm](https://github.com/kartoza/docker-osm)), which is an OSM PostGIS database with automatic updates from the global OSM data feed.

To start, follow the Docker-OSM [set up instructions](https://github.com/kartoza/docker-osm#quick-setup) to spin up an OSM instance for your geographic area of interest.

Note that Docker-OSM uses the [omniscale/imposm3](https://github.com/omniscale/imposm3) tool to import OSM data from a PBF file into a PostGIS database. Generally, you can create a customized database schema for this importing process by editing a ```mapping.yml``` configuration file when setting up Docker-OSM. However, the built-in Underpass queries require a specific PostGIS schema using the following ```mapping.yml``` configuration:

```yml
tags:
  load_all: true
use_single_id_space: true
areas:
  area_tags: [building, landuse, leisure, natural, aeroway]
  linear_tags: [highway, barrier]
tables:
  all:
    type: geometry
    columns:
      - name: osm_id
        type: id
      - name: tags
        type: hstore_tags
      - name: geometry
        type: geometry
    type_mappings:
      points:
        __any__: [__any__]
      linestrings:
        __any__: [__any__]
      polygons:
        __any__: [__any__]
```

Depending on the size of your geographic area of interest, the sequence of downloading a PBF file, spinning up a Docker-OSM instance and waiting for the OSM feature import process to complete can take a bit of time.

Once that process completes and you have a running instance (that both has data in a database and is importing live data from the OSM global feed), then you can spin up an Underpass API server.

### Underpass Setup

First, clone the Underpass repository:

```bash
git clone git@github.com:engelsjk/underpass.git
```

Then build an Underpass binary:

```bash
go build -o bin/underpass cmd/underpass/main.go
```

This Underpass binary requires a ```.env``` file that includes the connection parameters to a Docker-OSM instance. It must include the following variables:

```
DB_HOST=localhost
DB_PORT=35432
DB_USER=docker
DB_PASSWORD=docker
DB_NAME=gis
UNDERPASS_LOG=underpass.log
UNDERPASS_HOST=
UNDERPASS_PORT=3000
```

Copy the  ```.example.env``` file in this repository, rename it to ```.env``` and the edit the variables to match the existing Docker-OSM instance that you've set up in the previous section.

Note that the Underpass binary has a single flag option ```-save_logs``` which will output the Underpass server logs to the filename specified in the ```.env``` file.

## Run

After building the binary, start up the Underpass server:

```bash
bin/underpass -save_logs
```

## PostgreSQL Interface

The PostgreSQL driver from [jackc/pgx](https://github.com/jackc/pgx) is used to interface with the Docker-OSM PostGIS instance. A database schema file is used to map to the Docker-OSM schema (```sqlc/schema.sql```), and OSM feature queries are defined in SQL files. Type-safe Go code is then generated from those SQL files using [kyleconroy/sqlc](https://github.com/kyleconroy/sqlc).

## Customization

By default, Underpass is configured to match a PostgreSQL database schema for a specific Docker-OSM configuration (shown above in ```Docker-OSM Setup```). A limited set of SQL queries is also configured by default.

However, a customized Docker-OSM schema can be used and more queries can be added to your unique Underpass instance.

Underpass uses the [kyleconroy/sqlc](https://github.com/kyleconroy/sqlc) tool to generate type-safe  code from SQL. To customize your Underpass SQL configuration, first edit the ```sqlc.yaml``` file as needed. Then edit the database schema and SQL query files.

WARNING: Customization of schemas and queries in Underpass can be quite involved. It requires new SQL code and new type-safe Go code generated from it, and then extensive modification to add new route handlers for the new queries. Your mileage may vary.

### Schema

This is the default SQL schema for the Docker-OSM mapping shown above:

```sql
CREATE TABLE public.osm_all (
    id integer NOT NULL,
    osm_id bigint NOT NULL,
    tags public.hstore,
    geometry public.geometry(Geometry,4326)
);
```

This schema can me edited in ```sqlc/schema.sql``` to match whatever mapping you have configured in your Docker-OSM setup.

### Queries

You can add your own SQL code to include other queries by editing ```sqlc/query.sql```.

### Type-Safe Go from SQL

Next, generate new type-safe Go code from your customized schema and queries:

```sqlc generate```

This new type-safe Go code can be then used to add these custom queries as new routes and handlers to the Underpass API.

## Dependencies

* [gofiber/fiber](https://github.com/gofiber/fiber)
* [jackc/pgx](https://github.com/jackc/pgx)
* [jackc/pgtype](https://github.com/jackc/pgtype)
* [paulmach/go.geojson](https://github.com/paulmach/go.geojson)
