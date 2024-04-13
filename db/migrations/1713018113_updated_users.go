package migrations

import (
	"encoding/json"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/daos"
	m "github.com/pocketbase/pocketbase/migrations"
	"github.com/pocketbase/pocketbase/models"
	"github.com/pocketbase/pocketbase/models/schema"
)

func init() {
	m.Register(func(db dbx.Builder) error {
		dao := daos.New(db)

		collection, err := dao.FindCollectionByNameOrId("_pb_users_auth_")
		if err != nil {
			return err
		}

		// remove
		collection.Schema.RemoveField("wjofulpg")

		// remove
		collection.Schema.RemoveField("t1wlsqyp")

		// remove
		collection.Schema.RemoveField("fhxhln9g")

		// remove
		collection.Schema.RemoveField("wosrk4ue")

		query := dao.RecordQuery("_pb_users_auth_")

		users := []*models.Record{}
		if err := query.All(&users); err != nil {
			return err
		}
		settingsCollection, err := dao.FindCollectionByNameOrId("settings")

		if err != nil {
			return err
		}

		for _, user := range users {
			settings := models.NewRecord(settingsCollection)
			language := user.Get("language")
			unit := user.Get("unit")
			location := user.Get("location")
			settings.Set("language", language)
			settings.Set("unit", unit)
			settings.Set("location", location)
			settings.Set("user", user.Id)
			settings.Set("mapFocus", "location")
			if err := dao.SaveRecord(settings); err != nil {
				return err
			}
		}

		return dao.SaveCollection(collection)
	}, func(db dbx.Builder) error {
		dao := daos.New(db)

		collection, err := dao.FindCollectionByNameOrId("_pb_users_auth_")
		if err != nil {
			return err
		}

		// add
		del_unit := &schema.SchemaField{}
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
		collection.Schema.AddField(del_unit)

		// add
		del_language := &schema.SchemaField{}
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
		collection.Schema.AddField(del_language)

		// add
		del_mapView := &schema.SchemaField{}
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
		collection.Schema.AddField(del_mapView)

		// add
		del_location := &schema.SchemaField{}
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
		collection.Schema.AddField(del_location)

		return dao.SaveCollection(collection)
	})
}
