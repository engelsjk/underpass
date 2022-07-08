# Underpass

An API built on top of a live OpenStreetMap (OSM) instance that provides custom and fast queries of selected OSM features. An alternative to the [Overpass API](https://wiki.openstreetmap.org/wiki/Overpass_API).

## Setup

The Underpass setup consists of two phases:

* Spin up a live instance of ```docker-osm``` for your area of interest using an Underpass-specific database schema
* Start an Underpass server that connects to the running ```docker-osm``` instance

### Docker-OSM Setup

Underpass requires a running instance of ```docker-osm``` ([kartoza/docker-osm](https://github.com/kartoza/docker-osm)), which is an OSM PostGIS database with automatic updates from the global OSM data feed.

To start, follow the ```docker-osm``` [set up instructions](https://github.com/kartoza/docker-osm#quick-setup) to spin up an OSM instance for your geographic area of interest.

Note that ```docker-osm``` uses the [omniscale/imposm3](https://github.com/omniscale/imposm3) tool to import OSM data from a PBF file into a PostGIS database. Generally, you can create a customized database schema for this importing process by editing a ```mapping.yml``` configuration file when setting up ```docker-osm```. However, the built-in Underpass queries require a specific PostGIS schema using the following ```mapping.yml``` configuration:

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

Depending on the size of your geographic area of interest, the sequence of downloading a PBF file, spinning up a ```docker-osm``` instance and waiting for the OSM feature import process to complete can take a bit of time.

Once that process completes and you have a running instance (which both has data in a database and is importing live data from the OSM global feed), then you can spin up an Underpass API server.

### Underpass Setup

First, clone the Underpass repository:

```bash
git clone git@github.com:engelsjk/underpass.git
```

Then build an Underpass binary:

```bash
go build -o bin/underpass cmd/underpass/main.go
```

This Underpass binary requires a ```.env``` file that includes the connection parameters to the ```docker-osm``` instance. Copy the  ```.example.env``` file in this repository, rename it to ```.env``` and then edit the variables to match the existing ```docker-osm``` instance that you've set up in the previous section.

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

Note that the Underpass binary has a single flag option ```-save_logs``` which will output the Underpass server logs to the filename specified by the ```UNDERPASS_LOG``` variable in the  ```.env``` file.

## Run

After building the binary, start up the Underpass server:

```bash
bin/underpass -save_logs
```

## Queries

By default, Underpass includes only two SQL queries to the ```docker-osm``` database:

* ListByID: a list of features by OSM node/way/relation ID
* ListByBoundingBox: a list of features by a lower-left / upper-right coordinate pair bounding box

### List By ID

This query is used to get a single OSM feature (either a node, way or relation) by its OSM ID.

```
/api/v1/features/{osm_id}/{node|way|relation}
```

If that OSM ID exists in the database, the response will be a single GeoJSON feature.

For example:

```
/api/v1/features/48985299/way
```

```json
{
    "type": "Feature",
    "id": "way/48985299",
    "geometry": {
      "type": "Polygon",
      "coordinates": [...]
    },
    "properties": {
      "name": "AT&T Bedminster",
      "building": "commercial",
      ...
    }
}
```

### List By Bounding Box

This query is used to get all OSM features within a bounding box, given by its lower-left and upper-right coordinate pair.

```
/api/v1/features?bbox={ll_lat},{ll_lon},{ur_lat},{ur_lon}
```

The response will be an array of GeoJSON features.

Additionally, a query parameter ```tag``` can be included to filter the OSM features in the bounding box by a single OSM tag key.

```
?tag={"highway": ["*"]}
```

The tag filter can be either a wildcard set (eg ```["*"]```) to include all OSM features that have that tag, regardless of the tag value. Or it can be an inclusive set  of multiple tag values (eg ```["service", "cycleway"])``` to include any OSM features that have those tag key:value properties.

### List By Key

This query can be used to get features by specific OSM tag key values. Currently, only the tags 'wikidata' and 'wikipedia' are available by default.

```
/api/v1/features?wikidata={wikidata_id}
```

```
/api/v1/features?wikipedia={wikipedia_name}
```

These queries will work ok with a default ```docker-osm```, but response times may be slow depending on the size of your database. For faster queries, you can build an index within the ```docker-osm``` instance for each tag key that you'd like to query.

```
CREATE INDEX osm_all_tags_key_wikidata_idx ON osm_all ((tags -> 'wikidata'));
```

## PostgreSQL Interface

The PostgreSQL driver from [jackc/pgx](https://github.com/jackc/pgx) is used to interface with the ```docker-osm``` PostGIS instance. A database schema file is used to map to the ```docker-osm``` schema (```sqlc/schema.sql```) and OSM feature queries are defined in a query file (```sqlc/query.sql```). Type-safe Go code is then generated from those SQL files using [kyleconroy/sqlc](https://github.com/kyleconroy/sqlc).

## Customization

By default, Underpass is configured to match a PostgreSQL database schema for a specific ```docker-osm``` configuration (shown above in ```Docker-OSM Setup```). A limited set of SQL queries is also configured by default, as noted above.

However, a customized ```docker-osm``` schema can be used that match a desired ```docker-osm``` configuration, and more queries can be added to your unique Underpass instance.

Customization of schemas and queries in Underpass can be quite involved however. It requires new SQL code and new type-safe Go code to be generated from the SQL, as well as modification to add new route handlers for the new queries. Your mileage may vary.

Underpass uses the [kyleconroy/sqlc](https://github.com/kyleconroy/sqlc) tool to generate type-safe  code from SQL. To customize your Underpass SQL configuration, first edit the ```sqlc.yaml``` file as needed. Then edit the database schema and SQL query files.

### Schema

This is the default SQL schema for the ```docker-osm``` mapping shown above:

```sql
CREATE TABLE public.osm_all (
    id integer NOT NULL,
    osm_id bigint NOT NULL,
    tags public.hstore,
    geometry public.geometry(Geometry,4326)
);
```

This schema can me edited in ```sqlc/schema.sql``` to match whatever mapping you have configured in your ```docker-osm``` setup.

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
