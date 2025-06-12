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
					"autogeneratePattern": "[a-z0-9]{15}",
					"hidden": false,
					"id": "text3208210256",
					"max": 15,
					"min": 15,
					"name": "id",
					"pattern": "^[a-z0-9]+$",
					"presentable": false,
					"primaryKey": true,
					"required": true,
					"system": true,
					"type": "text"
				},
				{
					"cascadeDelete": true,
					"collectionId": "e864strfxo14pm4",
					"hidden": false,
					"id": "relation2993194383",
					"maxSelect": 1,
					"minSelect": 0,
					"name": "trail",
					"presentable": false,
					"required": true,
					"system": false,
					"type": "relation"
				},
				{
					"cascadeDelete": true,
					"collectionId": "pbc_1295301207",
					"hidden": false,
					"id": "relation1148540665",
					"maxSelect": 1,
					"minSelect": 0,
					"name": "actor",
					"presentable": false,
					"required": true,
					"system": false,
					"type": "relation"
				},
				{
					"hidden": false,
					"id": "autodate2990389176",
					"name": "created",
					"onCreate": true,
					"onUpdate": false,
					"presentable": false,
					"system": false,
					"type": "autodate"
				},
				{
					"hidden": false,
					"id": "autodate3332085495",
					"name": "updated",
					"onCreate": true,
					"onUpdate": true,
					"presentable": false,
					"system": false,
					"type": "autodate"
				}
			],
			"id": "pbc_1995454416",
			"indexes": [
				"CREATE UNIQUE INDEX ` + "`" + `idx_ywIkmeaSFo` + "`" + ` ON ` + "`" + `trail_like` + "`" + ` (\n  ` + "`" + `trail` + "`" + `,\n  ` + "`" + `actor` + "`" + `\n)"
			],
			"listRule": null,
			"name": "trail_like",
			"system": false,
			"type": "base",
			"updateRule": null,
			"viewRule": null
		}`

		collection := &core.Collection{}
		if err := json.Unmarshal([]byte(jsonData), &collection); err != nil {
			return err
		}

		return app.Save(collection)
	}, func(app core.App) error {
		collection, err := app.FindCollectionByNameOrId("pbc_1995454416")
		if err != nil {
			return err
		}

		return app.Delete(collection)
	})
}
