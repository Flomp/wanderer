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
					"id": "json1148540665",
					"maxSize": 1,
					"name": "actor",
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
					"id": "json521872670",
					"maxSize": 1,
					"name": "item",
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
				}
			],
			"id": "pbc_1973704172",
			"indexes": [],
			"listRule": "",
			"name": "profile_feed",
			"system": false,
			"type": "view",
			"updateRule": null,
			"viewQuery": "SELECT\n    (ROW_NUMBER() OVER ()) as id,\n    actor,\n    author,\n    item,\n    type,\n    created\nFROM\n    (\n        SELECT\n            author as actor,\n            author,\n            id as item,\n            \"list\" as type,\n            created\n        FROM\n            lists\n        UNION\n        SELECT\n            author as actor,\n            author,\n            id as item,\n            \"trail\" as type,\n            created\n        FROM trails\n    )\nORDER BY created desc;",
			"viewRule": null
		}`

		collection := &core.Collection{}
		if err := json.Unmarshal([]byte(jsonData), &collection); err != nil {
			return err
		}

		return app.Save(collection)
	}, func(app core.App) error {
		collection, err := app.FindCollectionByNameOrId("pbc_1973704172")
		if err != nil {
			return err
		}

		return app.Delete(collection)
	})
}
