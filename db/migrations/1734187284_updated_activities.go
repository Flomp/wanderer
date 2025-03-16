package migrations

import (
	"encoding/json"

	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
	"github.com/pocketbase/pocketbase/tools/types"
)

func init() {
	m.Register(func(app core.App) error {

		collection, err := app.FindCollectionByNameOrId("t9lphichi5xwyeu")
		if err != nil {
			return err
		}

		collection.ListRule = types.Pointer("")

		collection.ViewRule = types.Pointer("")

		// remove
		collection.Fields.RemoveById("zegqj8bt")

		// remove
		collection.Fields.RemoveById("mztv5od5")

		// remove
		collection.Fields.RemoveById("4hq14mc5")

		// remove
		collection.Fields.RemoveById("hbyuxehw")

		// remove
		collection.Fields.RemoveById("syp8ya96")

		// remove
		collection.Fields.RemoveById("rrfp5omm")

		// remove
		collection.Fields.RemoveById("tacveduk")

		// remove
		collection.Fields.RemoveById("ydhldat1")

		// remove
		collection.Fields.RemoveById("9gg24ge8")

		// remove
		collection.Fields.RemoveById("ujsn6lqc")

		// remove
		collection.Fields.RemoveById("nvvrulvj")

		// add
		new_date := &core.JSONField{}
		if err := json.Unmarshal([]byte(`{
			"system": false,
			"id": "qejtfjom",
			"name": "date",
			"type": "json",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"maxSize": 1
			}
		}`), new_date); err != nil {
			return err
		}
		collection.Fields.Add(new_date)

		// add
		new_name := &core.JSONField{}
		if err := json.Unmarshal([]byte(`{
			"system": false,
			"id": "nskezkjx",
			"name": "name",
			"type": "json",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"maxSize": 1
			}
		}`), new_name); err != nil {
			return err
		}
		collection.Fields.Add(new_name)

		// add
		new_description := &core.JSONField{}
		if err := json.Unmarshal([]byte(`{
			"system": false,
			"id": "viedej4h",
			"name": "description",
			"type": "json",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"maxSize": 1
			}
		}`), new_description); err != nil {
			return err
		}
		collection.Fields.Add(new_description)

		// add
		new_gpx := &core.JSONField{}
		if err := json.Unmarshal([]byte(`{
			"system": false,
			"id": "vljhhvdy",
			"name": "gpx",
			"type": "json",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"maxSize": 1
			}
		}`), new_gpx); err != nil {
			return err
		}
		collection.Fields.Add(new_gpx)

		// add
		new_author := &core.JSONField{}
		if err := json.Unmarshal([]byte(`{
			"system": false,
			"id": "z1uaeqzu",
			"name": "author",
			"type": "json",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"maxSize": 1
			}
		}`), new_author); err != nil {
			return err
		}
		collection.Fields.Add(new_author)

		// add
		new_photos := &core.JSONField{}
		if err := json.Unmarshal([]byte(`{
			"system": false,
			"id": "9nnpyhlq",
			"name": "photos",
			"type": "json",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"maxSize": 1
			}
		}`), new_photos); err != nil {
			return err
		}
		collection.Fields.Add(new_photos)

		// add
		new_distance := &core.JSONField{}
		if err := json.Unmarshal([]byte(`{
			"system": false,
			"id": "7kzcclzu",
			"name": "distance",
			"type": "json",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"maxSize": 1
			}
		}`), new_distance); err != nil {
			return err
		}
		collection.Fields.Add(new_distance)

		// add
		new_duration := &core.JSONField{}
		if err := json.Unmarshal([]byte(`{
			"system": false,
			"id": "gwpsktbu",
			"name": "duration",
			"type": "json",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"maxSize": 1
			}
		}`), new_duration); err != nil {
			return err
		}
		collection.Fields.Add(new_duration)

		// add
		new_elevation_gain := &core.JSONField{}
		if err := json.Unmarshal([]byte(`{
			"system": false,
			"id": "5tze9ybp",
			"name": "elevation_gain",
			"type": "json",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"maxSize": 1
			}
		}`), new_elevation_gain); err != nil {
			return err
		}
		collection.Fields.Add(new_elevation_gain)

		// add
		new_elevation_loss := &core.JSONField{}
		if err := json.Unmarshal([]byte(`{
			"system": false,
			"id": "xkuz8ixp",
			"name": "elevation_loss",
			"type": "json",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"maxSize": 1
			}
		}`), new_elevation_loss); err != nil {
			return err
		}
		collection.Fields.Add(new_elevation_loss)

		// add
		new_type := &core.JSONField{}
		if err := json.Unmarshal([]byte(`{
			"system": false,
			"id": "soylsusn",
			"name": "type",
			"type": "json",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"maxSize": 1
			}
		}`), new_type); err != nil {
			return err
		}
		collection.Fields.Add(new_type)

		return app.Save(collection)
	}, func(app core.App) error {

		collection, err := app.FindCollectionByNameOrId("t9lphichi5xwyeu")
		if err != nil {
			return err
		}

		collection.ListRule = nil

		collection.ViewRule = nil

		// add
		del_date := &core.JSONField{}
		if err := json.Unmarshal([]byte(`{
			"system": false,
			"id": "zegqj8bt",
			"name": "date",
			"type": "json",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"maxSize": 1
			}
		}`), del_date); err != nil {
			return err
		}
		collection.Fields.Add(del_date)

		// add
		del_name := &core.JSONField{}
		if err := json.Unmarshal([]byte(`{
			"system": false,
			"id": "mztv5od5",
			"name": "name",
			"type": "json",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"maxSize": 1
			}
		}`), del_name); err != nil {
			return err
		}
		collection.Fields.Add(del_name)

		// add
		del_description := &core.JSONField{}
		if err := json.Unmarshal([]byte(`{
			"system": false,
			"id": "4hq14mc5",
			"name": "description",
			"type": "json",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"maxSize": 1
			}
		}`), del_description); err != nil {
			return err
		}
		collection.Fields.Add(del_description)

		// add
		del_gpx := &core.JSONField{}
		if err := json.Unmarshal([]byte(`{
			"system": false,
			"id": "hbyuxehw",
			"name": "gpx",
			"type": "json",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"maxSize": 1
			}
		}`), del_gpx); err != nil {
			return err
		}
		collection.Fields.Add(del_gpx)

		// add
		del_author := &core.JSONField{}
		if err := json.Unmarshal([]byte(`{
			"system": false,
			"id": "syp8ya96",
			"name": "author",
			"type": "json",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"maxSize": 1
			}
		}`), del_author); err != nil {
			return err
		}
		collection.Fields.Add(del_author)

		// add
		del_photos := &core.JSONField{}
		if err := json.Unmarshal([]byte(`{
			"system": false,
			"id": "rrfp5omm",
			"name": "photos",
			"type": "json",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"maxSize": 1
			}
		}`), del_photos); err != nil {
			return err
		}
		collection.Fields.Add(del_photos)

		// add
		del_distance := &core.JSONField{}
		if err := json.Unmarshal([]byte(`{
			"system": false,
			"id": "tacveduk",
			"name": "distance",
			"type": "json",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"maxSize": 1
			}
		}`), del_distance); err != nil {
			return err
		}
		collection.Fields.Add(del_distance)

		// add
		del_duration := &core.JSONField{}
		if err := json.Unmarshal([]byte(`{
			"system": false,
			"id": "ydhldat1",
			"name": "duration",
			"type": "json",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"maxSize": 1
			}
		}`), del_duration); err != nil {
			return err
		}
		collection.Fields.Add(del_duration)

		// add
		del_elevation_gain := &core.JSONField{}
		if err := json.Unmarshal([]byte(`{
			"system": false,
			"id": "9gg24ge8",
			"name": "elevation_gain",
			"type": "json",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"maxSize": 1
			}
		}`), del_elevation_gain); err != nil {
			return err
		}
		collection.Fields.Add(del_elevation_gain)

		// add
		del_elevation_loss := &core.JSONField{}
		if err := json.Unmarshal([]byte(`{
			"system": false,
			"id": "ujsn6lqc",
			"name": "elevation_loss",
			"type": "json",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"maxSize": 1
			}
		}`), del_elevation_loss); err != nil {
			return err
		}
		collection.Fields.Add(del_elevation_loss)

		// add
		del_type := &core.JSONField{}
		if err := json.Unmarshal([]byte(`{
			"system": false,
			"id": "nvvrulvj",
			"name": "type",
			"type": "json",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"maxSize": 1
			}
		}`), del_type); err != nil {
			return err
		}
		collection.Fields.Add(del_type)

		// remove
		collection.Fields.RemoveById("qejtfjom")

		// remove
		collection.Fields.RemoveById("nskezkjx")

		// remove
		collection.Fields.RemoveById("viedej4h")

		// remove
		collection.Fields.RemoveById("vljhhvdy")

		// remove
		collection.Fields.RemoveById("z1uaeqzu")

		// remove
		collection.Fields.RemoveById("9nnpyhlq")

		// remove
		collection.Fields.RemoveById("7kzcclzu")

		// remove
		collection.Fields.RemoveById("gwpsktbu")

		// remove
		collection.Fields.RemoveById("5tze9ybp")

		// remove
		collection.Fields.RemoveById("xkuz8ixp")

		// remove
		collection.Fields.RemoveById("soylsusn")

		return app.Save(collection)
	})
}
