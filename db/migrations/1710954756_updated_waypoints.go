package migrations

import (
	"encoding/json"

	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {

		collection, err := app.FindCollectionByNameOrId("goeo2ubp103rzp9")
		if err != nil {
			return err
		}

		// add
		new_author := &core.RelationField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "8qbxrsd8",
			"name": "author",
			"type": "relation",
			"required": true,
			"presentable": false,
			"unique": false,
			"options": {
				"collectionId": "_pb_users_auth_",
				"cascadeDelete": false,
				"minSelect": null,
				"maxSelect": 1,
				"displayFields": null
			}
		}`), new_author)
		collection.Fields.Add(new_author)

		return app.Save(collection)
	}, func(app core.App) error {

		collection, err := app.FindCollectionByNameOrId("goeo2ubp103rzp9")
		if err != nil {
			return err
		}

		// remove
		collection.Fields.RemoveById("8qbxrsd8")

		return app.Save(collection)
	})
}
