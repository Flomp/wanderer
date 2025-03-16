package migrations

import (
	"encoding/json"

	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {

		collection, err := app.FindCollectionByNameOrId("uavt73rsqcn1n13")
		if err != nil {
			return err
		}

		// add
		new_tilesets := &core.JSONField{}
		if err := json.Unmarshal([]byte(`{
			"system": false,
			"id": "xdbayoqg",
			"name": "tilesets",
			"type": "json",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"maxSize": 2000000
			}
		}`), new_tilesets); err != nil {
			return err
		}
		collection.Fields.Add(new_tilesets)

		return app.Save(collection)
	}, func(app core.App) error {

		collection, err := app.FindCollectionByNameOrId("uavt73rsqcn1n13")
		if err != nil {
			return err
		}

		// remove
		collection.Fields.RemoveById("xdbayoqg")

		return app.Save(collection)
	})
}
