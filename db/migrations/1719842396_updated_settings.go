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

		// update
		edit_language := &core.SelectField{}
		if err := json.Unmarshal([]byte(`{
			"system": false,
			"id": "0sepzvkh",
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
					"hu",
					"it",
					"nl",
					"pl",
					"pt",
					"zh"
				]
			}
		}`), edit_language); err != nil {
			return err
		}
		collection.Fields.Add(edit_language)

		return app.Save(collection)
	}, func(app core.App) error {

		collection, err := app.FindCollectionByNameOrId("uavt73rsqcn1n13")
		if err != nil {
			return err
		}

		// update
		edit_language := &core.SelectField{}
		if err := json.Unmarshal([]byte(`{
			"system": false,
			"id": "0sepzvkh",
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
					"hu",
					"nl",
					"pl",
					"pt",
					"zh"
				]
			}
		}`), edit_language); err != nil {
			return err
		}
		collection.Fields.Add(edit_language)

		return app.Save(collection)
	})
}
