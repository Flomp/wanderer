package migrations

import (
	"encoding/json"

	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {

		collection, err := app.FindCollectionByNameOrId("iz4sezoehde64wp")
		if err != nil {
			return err
		}

		// add
		new_komoot := &core.JSONField{}
		if err := json.Unmarshal([]byte(`{
			"system": false,
			"id": "6s0oxqgp",
			"name": "komoot",
			"type": "json",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"maxSize": 2000000
			}
		}`), new_komoot); err != nil {
			return err
		}
		collection.Fields.Add(new_komoot)

		return app.Save(collection)
	}, func(app core.App) error {

		collection, err := app.FindCollectionByNameOrId("iz4sezoehde64wp")
		if err != nil {
			return err
		}

		// remove
		collection.Fields.RemoveById("6s0oxqgp")

		return app.Save(collection)
	})
}
