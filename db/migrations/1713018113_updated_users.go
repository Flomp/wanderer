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

		// remove
		collection.Fields.RemoveById("wjofulpg")

		// remove
		collection.Fields.RemoveById("t1wlsqyp")

		// remove
		collection.Fields.RemoveById("fhxhln9g")

		// remove
		collection.Fields.RemoveById("wosrk4ue")

		query := app.RecordQuery("_pb_users_auth_")

		users := []*core.Record{}
		if err := query.All(&users); err != nil {
			return err
		}
		settingsCollection, err := app.FindCollectionByNameOrId("settings")

		if err != nil {
			return err
		}

		for _, user := range users {
			settings := core.NewRecord(settingsCollection)
			language := user.Get("language")
			unit := user.Get("unit")
			location := user.Get("location")
			settings.Set("language", language)
			settings.Set("unit", unit)
			settings.Set("location", location)
			settings.Set("user", user.Id)
			settings.Set("mapFocus", "location")
			if err := app.Save(settings); err != nil {
				return err
			}
		}

		return app.Save(collection)
	}, func(app core.App) error {

		collection, err := app.FindCollectionByNameOrId("_pb_users_auth_")
		if err != nil {
			return err
		}

		// add
		del_unit := &core.SelectField{}
		if err := json.Unmarshal([]byte(`{
			"system": false,
			"id": "wjofulpg",
			"name": "unit",
			"type": "select",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"maxSelect": 1,
				"values": [
					"metric",
					"imperial"
				]
			}
		}`), del_unit); err != nil {
			return err
		}
		collection.Fields.Add(del_unit)

		// add
		del_language := &core.SelectField{}
		if err := json.Unmarshal([]byte(`{
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
					"hu",
					"nl",
					"pl",
					"pt",
					"zh"
				]
			}
		}`), del_language); err != nil {
			return err
		}
		collection.Fields.Add(del_language)

		// add
		del_mapView := &core.SelectField{}
		if err := json.Unmarshal([]byte(`{
			"system": false,
			"id": "fhxhln9g",
			"name": "mapView",
			"type": "select",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"maxSelect": 1,
				"values": [
					"location",
					"trails"
				]
			}
		}`), del_mapView); err != nil {
			return err
		}
		collection.Fields.Add(del_mapView)

		// add
		del_location := &core.JSONField{}
		if err := json.Unmarshal([]byte(`{
			"system": false,
			"id": "wosrk4ue",
			"name": "location",
			"type": "json",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"maxSize": 2000000
			}
		}`), del_location); err != nil {
			return err
		}
		collection.Fields.Add(del_location)

		return app.Save(collection)
	})
}
