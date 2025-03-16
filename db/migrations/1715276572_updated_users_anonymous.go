package migrations

import (
	"encoding/json"

	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
	"github.com/pocketbase/pocketbase/tools/types"
)

func init() {
	m.Register(func(app core.App) error {

		collection, err := app.FindCollectionByNameOrId("xku110v5a5xbufa")
		if err != nil {
			return err
		}

		collection.ListRule = types.Pointer("@request.auth.id != \"\"")

		collection.ViewRule = types.Pointer("@request.auth.id != \"\"")

		// remove
		collection.Fields.RemoveById("cknini6y")

		// add
		new_username := &core.TextField{}
		if err := json.Unmarshal([]byte(`{
			"system": false,
			"id": "qzkmy7nv",
			"name": "username",
			"type": "text",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"min": null,
				"max": null,
				"pattern": ""
			}
		}`), new_username); err != nil {
			return err
		}
		collection.Fields.Add(new_username)

		return app.Save(collection)
	}, func(app core.App) error {

		collection, err := app.FindCollectionByNameOrId("xku110v5a5xbufa")
		if err != nil {
			return err
		}

		collection.ListRule = nil

		collection.ViewRule = nil

		// add
		del_username := &core.TextField{}
		if err := json.Unmarshal([]byte(`{
			"system": false,
			"id": "cknini6y",
			"name": "username",
			"type": "text",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"min": null,
				"max": null,
				"pattern": ""
			}
		}`), del_username); err != nil {
			return err
		}
		collection.Fields.Add(del_username)

		// remove
		collection.Fields.RemoveById("qzkmy7nv")

		return app.Save(collection)
	})
}
