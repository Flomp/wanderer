package migrations

import (
	"encoding/json"

	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
	"github.com/pocketbase/pocketbase/tools/types"
)

func init() {
	m.Register(func(app core.App) error {

		collection, err := app.FindCollectionByNameOrId("4wbv9tz5zjdrjh1")
		if err != nil {
			return err
		}

		collection.ListRule = nil

		collection.ViewRule = types.Pointer("@request.auth.id = id")

		// remove
		collection.Fields.RemoveById("5udo4avx")

		// remove
		collection.Fields.RemoveById("omuxiatj")

		// remove
		collection.Fields.RemoveById("yq4snqnf")

		// remove
		collection.Fields.RemoveById("q8d41loj")

		// remove
		collection.Fields.RemoveById("xsetw0o6")

		// remove
		collection.Fields.RemoveById("n4pq80mu")

		// add
		new_max_distance := &core.JSONField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "nqiuah7b",
			"name": "max_distance",
			"type": "json",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"maxSize": 1
			}
		}`), new_max_distance)
		collection.Fields.Add(new_max_distance)

		// add
		new_max_elevation_gain := &core.JSONField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "rpzvilry",
			"name": "max_elevation_gain",
			"type": "json",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"maxSize": 1
			}
		}`), new_max_elevation_gain)
		collection.Fields.Add(new_max_elevation_gain)

		// add
		new_max_duration := &core.JSONField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "pfppgcfx",
			"name": "max_duration",
			"type": "json",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"maxSize": 1
			}
		}`), new_max_duration)
		collection.Fields.Add(new_max_duration)

		// add
		new_min_distance := &core.JSONField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "jrqp7c8d",
			"name": "min_distance",
			"type": "json",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"maxSize": 1
			}
		}`), new_min_distance)
		collection.Fields.Add(new_min_distance)

		// add
		new_min_elevation_gain := &core.JSONField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "xjcpg4su",
			"name": "min_elevation_gain",
			"type": "json",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"maxSize": 1
			}
		}`), new_min_elevation_gain)
		collection.Fields.Add(new_min_elevation_gain)

		// add
		new_min_duration := &core.JSONField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "cuj9awlq",
			"name": "min_duration",
			"type": "json",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"maxSize": 1
			}
		}`), new_min_duration)
		collection.Fields.Add(new_min_duration)

		return app.Save(collection)
	}, func(app core.App) error {

		collection, err := app.FindCollectionByNameOrId("4wbv9tz5zjdrjh1")
		if err != nil {
			return err
		}

		collection.ListRule = types.Pointer("@request.auth.id != \"\"")

		collection.ViewRule = nil

		// add
		del_max_distance := &core.JSONField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "5udo4avx",
			"name": "max_distance",
			"type": "json",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"maxSize": 1
			}
		}`), del_max_distance)
		collection.Fields.Add(del_max_distance)

		// add
		del_max_elevation_gain := &core.JSONField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "omuxiatj",
			"name": "max_elevation_gain",
			"type": "json",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"maxSize": 1
			}
		}`), del_max_elevation_gain)
		collection.Fields.Add(del_max_elevation_gain)

		// add
		del_max_duration := &core.JSONField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "yq4snqnf",
			"name": "max_duration",
			"type": "json",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"maxSize": 1
			}
		}`), del_max_duration)
		collection.Fields.Add(del_max_duration)

		// add
		del_min_distance := &core.JSONField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "q8d41loj",
			"name": "min_distance",
			"type": "json",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"maxSize": 1
			}
		}`), del_min_distance)
		collection.Fields.Add(del_min_distance)

		// add
		del_min_elevation_gain := &core.JSONField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "xsetw0o6",
			"name": "min_elevation_gain",
			"type": "json",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"maxSize": 1
			}
		}`), del_min_elevation_gain)
		collection.Fields.Add(del_min_elevation_gain)

		// add
		del_min_duration := &core.JSONField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "n4pq80mu",
			"name": "min_duration",
			"type": "json",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"maxSize": 1
			}
		}`), del_min_duration)
		collection.Fields.Add(del_min_duration)

		// remove
		collection.Fields.RemoveById("nqiuah7b")

		// remove
		collection.Fields.RemoveById("rpzvilry")

		// remove
		collection.Fields.RemoveById("pfppgcfx")

		// remove
		collection.Fields.RemoveById("jrqp7c8d")

		// remove
		collection.Fields.RemoveById("xjcpg4su")

		// remove
		collection.Fields.RemoveById("cuj9awlq")

		return app.Save(collection)
	})
}
