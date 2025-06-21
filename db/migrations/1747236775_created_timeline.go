package migrations

import (
	"encoding/json"

	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		jsonData := `{
			"createRule": null,
			"deleteRule": null,
			"fields": [
				{
					"autogeneratePattern": "",
					"hidden": false,
					"id": "text3208210256",
					"max": 0,
					"min": 0,
					"name": "id",
					"pattern": "^[a-z0-9]+$",
					"presentable": false,
					"primaryKey": true,
					"required": true,
					"system": true,
					"type": "text"
				},
				{
					"hidden": false,
					"id": "json2310347867",
					"maxSize": 1,
					"name": "trail_id",
					"presentable": false,
					"required": false,
					"system": false,
					"type": "json"
				},
				{
					"hidden": false,
					"id": "json2862495610",
					"maxSize": 1,
					"name": "date",
					"presentable": false,
					"required": false,
					"system": false,
					"type": "json"
				},
				{
					"hidden": false,
					"id": "json1579384326",
					"maxSize": 1,
					"name": "name",
					"presentable": false,
					"required": false,
					"system": false,
					"type": "json"
				},
				{
					"hidden": false,
					"id": "json1843675174",
					"maxSize": 1,
					"name": "description",
					"presentable": false,
					"required": false,
					"system": false,
					"type": "json"
				},
				{
					"hidden": false,
					"id": "json3275261007",
					"maxSize": 1,
					"name": "gpx",
					"presentable": false,
					"required": false,
					"system": false,
					"type": "json"
				},
				{
					"hidden": false,
					"id": "json3182418120",
					"maxSize": 1,
					"name": "author",
					"presentable": false,
					"required": false,
					"system": false,
					"type": "json"
				},
				{
					"hidden": false,
					"id": "json142008537",
					"maxSize": 1,
					"name": "photos",
					"presentable": false,
					"required": false,
					"system": false,
					"type": "json"
				},
				{
					"hidden": false,
					"id": "json479369857",
					"maxSize": 1,
					"name": "distance",
					"presentable": false,
					"required": false,
					"system": false,
					"type": "json"
				},
				{
					"hidden": false,
					"id": "json2254405824",
					"maxSize": 1,
					"name": "duration",
					"presentable": false,
					"required": false,
					"system": false,
					"type": "json"
				},
				{
					"hidden": false,
					"id": "json3015100073",
					"maxSize": 1,
					"name": "elevation_gain",
					"presentable": false,
					"required": false,
					"system": false,
					"type": "json"
				},
				{
					"hidden": false,
					"id": "json3171089056",
					"maxSize": 1,
					"name": "elevation_loss",
					"presentable": false,
					"required": false,
					"system": false,
					"type": "json"
				},
				{
					"hidden": false,
					"id": "json2990389176",
					"maxSize": 1,
					"name": "created",
					"presentable": false,
					"required": false,
					"system": false,
					"type": "json"
				},
				{
					"hidden": false,
					"id": "json2363381545",
					"maxSize": 1,
					"name": "type",
					"presentable": false,
					"required": false,
					"system": false,
					"type": "json"
				}
			],
			"id": "pbc_468398817",
			"indexes": [],
			"listRule": "@collection.trails.id ?= trail_id && @collection.trails.public ?= true",
			"name": "timeline",
			"system": false,
			"type": "view",
			"updateRule": null,
			"viewQuery": "-- database: /Users/christianbeutel/Documents/svelte/wanderer/db/pb_data/data.db\nSELECT\n    (ROW_NUMBER() OVER ()) as id,\n    trail_id,\n    date,\n    name,\n    description,\n    gpx,\n    author,\n    photos,\n    distance,\n    duration,\n    elevation_gain,\n    elevation_loss,\n    created,\n    type\nFROM\n    (\n        SELECT\n            summit_logs.trail as trail_id,\n            summit_logs.date,\n            trails.name,\n            text as description,\n            summit_logs.gpx,\n            activitypub_actors.iri as author,\n            summit_logs.photos,\n            summit_logs.distance,\n            summit_logs.duration,\n            summit_logs.elevation_gain,\n            summit_logs.elevation_loss,\n            summit_logs.created,\n            \"summit_log\" as type\n        FROM\n            summit_logs\n            JOIN trails ON summit_logs.trail = trails.id\n            JOIN activitypub_actors ON activitypub_actors.id = summit_logs.author\n        UNION\n        SELECT\n            trails.id as trail_id,\n            date,\n            trails.name,\n            description,\n            gpx,\n            activitypub_actors.iri as author,\n            photos,\n            distance,\n            duration,\n            elevation_gain,\n            elevation_loss,\n            trails.created,\n            \"trail\" as type\n        FROM\n            trails\n            JOIN activitypub_actors ON activitypub_actors.id = trails.author\n    )\nORDER BY\n    created DESC;\n",
			"viewRule": "@collection.trails.id ?= trail_id && @collection.trails.public ?= true"
		}`

		collection := &core.Collection{}
		if err := json.Unmarshal([]byte(jsonData), &collection); err != nil {
			return err
		}

		return app.Save(collection)
	}, func(app core.App) error {
		collection, err := app.FindCollectionByNameOrId("pbc_468398817")
		if err != nil {
			return err
		}

		return app.Delete(collection)
	})
}
