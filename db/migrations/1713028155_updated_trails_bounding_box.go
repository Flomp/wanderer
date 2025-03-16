package migrations

import (
	"encoding/json"

	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
	"github.com/pocketbase/pocketbase/tools/types"
)

func init() {
	m.Register(func(app core.App) error {

		collection, err := app.FindCollectionByNameOrId("urytyc428mwlbqq")
		if err != nil {
			return err
		}

		collection.ViewRule = types.Pointer("@request.auth.id = id")

		// remove
		collection.Fields.RemoveById("iyhsoisl")

		// remove
		collection.Fields.RemoveById("kx2qfztr")

		// remove
		collection.Fields.RemoveById("z4qsnjeb")

		// remove
		collection.Fields.RemoveById("p66xomdb")

		// add
		new_max_lat := &core.JSONField{}
		if err := json.Unmarshal([]byte(`{
			"system": false,
			"id": "osfey1yx",
			"name": "max_lat",
			"type": "json",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"maxSize": 1
			}
		}`), new_max_lat); err != nil {
			return err
		}
		collection.Fields.Add(new_max_lat)

		// add
		new_max_lon := &core.JSONField{}
		if err := json.Unmarshal([]byte(`{
			"system": false,
			"id": "eohslzky",
			"name": "max_lon",
			"type": "json",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"maxSize": 1
			}
		}`), new_max_lon); err != nil {
			return err
		}
		collection.Fields.Add(new_max_lon)

		// add
		new_min_lat := &core.JSONField{}
		if err := json.Unmarshal([]byte(`{
			"system": false,
			"id": "kigvankd",
			"name": "min_lat",
			"type": "json",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"maxSize": 1
			}
		}`), new_min_lat); err != nil {
			return err
		}
		collection.Fields.Add(new_min_lat)

		// add
		new_min_lon := &core.JSONField{}
		if err := json.Unmarshal([]byte(`{
			"system": false,
			"id": "ifnks9mg",
			"name": "min_lon",
			"type": "json",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"maxSize": 1
			}
		}`), new_min_lon); err != nil {
			return err
		}
		collection.Fields.Add(new_min_lon)

		return app.Save(collection)
	}, func(app core.App) error {

		collection, err := app.FindCollectionByNameOrId("urytyc428mwlbqq")
		if err != nil {
			return err
		}

		collection.ViewRule = nil

		// add
		del_max_lat := &core.JSONField{}
		if err := json.Unmarshal([]byte(`{
			"system": false,
			"id": "iyhsoisl",
			"name": "max_lat",
			"type": "json",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"maxSize": 1
			}
		}`), del_max_lat); err != nil {
			return err
		}
		collection.Fields.Add(del_max_lat)

		// add
		del_max_lon := &core.JSONField{}
		if err := json.Unmarshal([]byte(`{
			"system": false,
			"id": "kx2qfztr",
			"name": "max_lon",
			"type": "json",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"maxSize": 1
			}
		}`), del_max_lon); err != nil {
			return err
		}
		collection.Fields.Add(del_max_lon)

		// add
		del_min_lat := &core.JSONField{}
		if err := json.Unmarshal([]byte(`{
			"system": false,
			"id": "z4qsnjeb",
			"name": "min_lat",
			"type": "json",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"maxSize": 1
			}
		}`), del_min_lat); err != nil {
			return err
		}
		collection.Fields.Add(del_min_lat)

		// add
		del_min_lon := &core.JSONField{}
		if err := json.Unmarshal([]byte(`{
			"system": false,
			"id": "p66xomdb",
			"name": "min_lon",
			"type": "json",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"maxSize": 1
			}
		}`), del_min_lon); err != nil {
			return err
		}
		collection.Fields.Add(del_min_lon)

		// remove
		collection.Fields.RemoveById("osfey1yx")

		// remove
		collection.Fields.RemoveById("eohslzky")

		// remove
		collection.Fields.RemoveById("kigvankd")

		// remove
		collection.Fields.RemoveById("ifnks9mg")

		return app.Save(collection)
	})
}
