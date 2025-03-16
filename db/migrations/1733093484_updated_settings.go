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
		new_terrain := &core.URLField{}
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

		// remove
		collection.Fields.RemoveById("mvxf9llv")

		return app.Save(collection)
	})
}
