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

		// remove
		collection.Fields.RemoveById("mvxf9llv")

		// add
		new_terrain := &core.JSONField{}
		if err := json.Unmarshal([]byte(`{
			"system": false,
			"id": "unsh0qsp",
			"name": "terrain",
			"type": "json",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"maxSize": 2000000
			}
		}`), new_terrain); err != nil {
			return err
		}
		collection.Fields.Add(new_terrain)

		return app.Save(collection)
	}, func(app core.App) error {

		collection, err := app.FindCollectionByNameOrId("uavt73rsqcn1n13")
		if err != nil {
			return err
		}

		// add
		del_terrain := &core.URLField{}
		if err := json.Unmarshal([]byte(`{
			"system": false,
			"id": "mvxf9llv",
			"name": "terrain",
			"type": "url",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"exceptDomains": [],
				"onlyDomains": []
			}
		}`), del_terrain); err != nil {
			return err
		}
		collection.Fields.Add(del_terrain)

		// remove
		collection.Fields.RemoveById("unsh0qsp")

		return app.Save(collection)
	})
}
