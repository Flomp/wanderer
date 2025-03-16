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

		// update
		edit_language := &core.SelectField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "t1wlsqyp",
			"name": "language",
			"type": "select",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"maxSelect": 1,
				"values": [
					"en",
					"de",
					"fr",
					"nl",
					"pl",
					"pt"
				]
			}
		}`), edit_language)
		collection.Fields.Add(edit_language)

		return app.Save(collection)
	}, func(app core.App) error {

		collection, err := app.FindCollectionByNameOrId("_pb_users_auth_")
		if err != nil {
			return err
		}

		// update
		edit_language := &core.SelectField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "t1wlsqyp",
			"name": "language",
			"type": "select",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"maxSelect": 1,
				"values": [
					"en",
					"de",
					"nl",
					"pl",
					"pt"
				]
			}
		}`), edit_language)
		collection.Fields.Add(edit_language)

		return app.Save(collection)
	})
}
