package migrations

import (
	"encoding/json"

	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		jsonData := `{
			"id": "urytyc428mwlbqq",
			"created": "2024-04-13 17:00:41.541Z",
			"updated": "2024-04-13 17:00:41.541Z",
			"name": "trails_bounding_box",
			"type": "view",
			"system": false,
			"schema": [
				{
					"system": false,
					"id": "iyhsoisl",
					"name": "max_lat",
					"type": "json",
					"required": false,
					"presentable": false,
					"unique": false,
					"options": {
						"maxSize": 1
					}
				},
				{
					"system": false,
					"id": "kx2qfztr",
					"name": "max_lon",
					"type": "json",
					"required": false,
					"presentable": false,
					"unique": false,
					"options": {
						"maxSize": 1
					}
				},
				{
					"system": false,
					"id": "z4qsnjeb",
					"name": "min_lat",
					"type": "json",
					"required": false,
					"presentable": false,
					"unique": false,
					"options": {
						"maxSize": 1
					}
				},
				{
					"system": false,
					"id": "p66xomdb",
					"name": "min_lon",
					"type": "json",
					"required": false,
					"presentable": false,
					"unique": false,
					"options": {
						"maxSize": 1
					}
				}
			],
			"indexes": [],
			"listRule": null,
			"viewRule": null,
			"createRule": null,
			"updateRule": null,
			"deleteRule": null,
			"options": {
				"query": "SELECT users.id, COALESCE(MAX(trails.lat), 0) AS max_lat, COALESCE(MAX(trails.lon), 0) AS max_lon, COALESCE(MIN(trails.lat), 0) AS min_lat, COALESCE(MIN(trails.lon), 0) AS min_lon FROM users LEFT JOIN trails ON users.id = trails.author GROUP BY users.id;"
			}
		}`

		collection := &core.Collection{}
		if err := json.Unmarshal([]byte(jsonData), &collection); err != nil {
			return err
		}

		return app.Save(collection)
	}, func(app core.App) error {

		collection, err := app.FindCollectionByNameOrId("urytyc428mwlbqq")
		if err != nil {
			return err
		}

		return app.Delete(collection)
	})
}
