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
		new_distance_from_start := &core.NumberField{}
		if err := json.Unmarshal([]byte(`{
			"system": false,
			"id": "s1prb3fx",
			"name": "distance_from_start",
			"type": "number",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"min": 0,
				"max": null,
				"noDecimal": false
			}
		}`), new_distance_from_start); err != nil {
			return err
		}
		collection.Fields.Add(new_distance_from_start)

		return app.Save(collection)
	}, func(app core.App) error {

		collection, err := app.FindCollectionByNameOrId("goeo2ubp103rzp9")
		if err != nil {
			return err
		}

		// remove
		collection.Fields.RemoveById("s1prb3fx")

		return app.Save(collection)
	})
}
