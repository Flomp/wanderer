package migrations

import (
	"encoding/json"

	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {

		collection, err := app.FindCollectionByNameOrId("dd2l9a4vxpy2ni8")
		if err != nil {
			return err
		}

		// add
		new_distance := &core.NumberField{}
		if err := json.Unmarshal([]byte(`{
			"system": false,
			"id": "jovws28m",
			"name": "distance",
			"type": "number",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"min": 0,
				"max": null,
				"noDecimal": false
			}
		}`), new_distance); err != nil {
			return err
		}
		collection.Fields.Add(new_distance)

		// add
		new_elevation_gain := &core.NumberField{}
		if err := json.Unmarshal([]byte(`{
			"system": false,
			"id": "m2kndtwn",
			"name": "elevation_gain",
			"type": "number",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"min": 0,
				"max": null,
				"noDecimal": false
			}
		}`), new_elevation_gain); err != nil {
			return err
		}
		collection.Fields.Add(new_elevation_gain)

		// add
		new_elevation_loss := &core.NumberField{}
		if err := json.Unmarshal([]byte(`{
			"system": false,
			"id": "uqqo9cws",
			"name": "elevation_loss",
			"type": "number",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"min": 0,
				"max": null,
				"noDecimal": false
			}
		}`), new_elevation_loss); err != nil {
			return err
		}
		collection.Fields.Add(new_elevation_loss)

		// add
		new_duration := &core.NumberField{}
		if err := json.Unmarshal([]byte(`{
			"system": false,
			"id": "vwxjsrae",
			"name": "duration",
			"type": "number",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"min": 0,
				"max": null,
				"noDecimal": false
			}
		}`), new_duration); err != nil {
			return err
		}
		collection.Fields.Add(new_duration)

		return app.Save(collection)
	}, func(app core.App) error {

		collection, err := app.FindCollectionByNameOrId("dd2l9a4vxpy2ni8")
		if err != nil {
			return err
		}

		// remove
		collection.Fields.RemoveById("jovws28m")

		// remove
		collection.Fields.RemoveById("m2kndtwn")

		// remove
		collection.Fields.RemoveById("uqqo9cws")

		// remove
		collection.Fields.RemoveById("vwxjsrae")

		return app.Save(collection)
	})
}
