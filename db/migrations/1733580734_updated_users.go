package migrations

import (
	"encoding/json"

	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {

		collection, err := app.FindCollectionByNameOrId("_pb_users_auth_")
		if err != nil {
			return err
		}

		// add
		new_bio := &core.TextField{}
		if err := json.Unmarshal([]byte(`{
			"system": false,
			"id": "pd2cq8sq",
			"name": "bio",
			"type": "text",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"min": null,
				"max": 10000,
				"pattern": ""
			}
		}`), new_bio); err != nil {
			return err
		}
		collection.Fields.Add(new_bio)

		return app.Save(collection)
	}, func(app core.App) error {

		collection, err := app.FindCollectionByNameOrId("_pb_users_auth_")
		if err != nil {
			return err
		}

		// remove
		collection.Fields.RemoveById("pd2cq8sq")

		return app.Save(collection)
	})
}
