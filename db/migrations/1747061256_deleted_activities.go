package migrations

import (
	"encoding/json"

	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		collection, err := app.FindCollectionByNameOrId("t9lphichi5xwyeu")
		if err != nil {
			return nil
		}

		return app.Delete(collection)
	}, func(app core.App) error {
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
			"id": "t9lphichi5xwyeu",
			"indexes": [],
			"listRule": "(@collection.users_anonymous.id ?= author && @collection.users_anonymous.private ?= false\n&&\n@collection.trails.id ?= trail_id && \n    (\n        @collection.trails.author ?= @request.auth.id ||\n        @collection.trails.public ?= true || \n        (@request.auth.id != \"\" && @collection.trails.trail_share_via_trail.user ?= @request.auth.id)\n    )) && @request.auth.id = author",
			"name": "activities",
			"system": false,
			"type": "view",
			"updateRule": null,
			"viewQuery": "SELECT id,trail_id,date,name,description,gpx,author,photos,distance,duration,elevation_gain,elevation_loss,created,type \nFROM (\n    SELECT summit_logs.id,trails.id as trail_id,summit_logs.date,trails.name,text as description,summit_logs.gpx,summit_logs.author,summit_logs.photos,summit_logs.distance,summit_logs.duration,summit_logs.elevation_gain,summit_logs.elevation_loss,summit_logs.created,\"summit_log\" as type \n    FROM summit_logs\n    JOIN trails ON summit_logs.id IN (\n    SELECT value\n    FROM json_each(trails.summit_logs)\n    )\n    UNION\n    SELECT id,id as trail_id,date,name,description,gpx,author,photos,distance,duration,elevation_gain,elevation_loss,created,\"trail\" as type \n    FROM trails\n)\nORDER BY created DESC",
			"viewRule": "(@collection.users_anonymous.id ?= author && @collection.users_anonymous.private ?= false\n&&\n@collection.trails.id ?= trail_id && \n    (\n        @collection.trails.author ?= @request.auth.id ||\n        @collection.trails.public ?= true || \n        (@request.auth.id != \"\" && @collection.trails.trail_share_via_trail.user ?= @request.auth.id)\n    )) && @request.auth.id = author"
		}`

		collection := &core.Collection{}
		if err := json.Unmarshal([]byte(jsonData), &collection); err != nil {
			return err
		}

		return app.Save(collection)
	})
}
