package migrations

import (
	"encoding/json"

	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {

		collection, err := app.FindCollectionByNameOrId("t9lphichi5xwyeu")
		if err != nil {
			return err
		}

		collection.ViewQuery = "SELECT id,trail_id,date,name,description,gpx,author,photos,distance,duration,elevation_gain,elevation_loss,created,type \nFROM (\n    SELECT summit_logs.id,trails.id as trail_id,summit_logs.date,trails.name,text as description,summit_logs.gpx,summit_logs.author,summit_logs.photos,summit_logs.distance,summit_logs.duration,summit_logs.elevation_gain,summit_logs.elevation_loss,summit_logs.created,\"summit_log\" as type \n    FROM summit_logs\n    JOIN trails ON summit_logs.id IN (\n    SELECT value\n    FROM json_each(trails.summit_logs)\n    )\n    UNION\n    SELECT id,id as trail_id,date,name,description,gpx,author,photos,distance,duration,elevation_gain,elevation_loss,created,\"trail\" as type \n    FROM trails\n)\nORDER BY date DESC"

		// remove
		collection.Fields.RemoveById("mpcwfh4x")

		// remove
		collection.Fields.RemoveById("xsxx8b8a")

		// remove
		collection.Fields.RemoveById("4ybgblqj")

		// remove
		collection.Fields.RemoveById("f7y3unyx")

		// remove
		collection.Fields.RemoveById("cowkmcyy")

		// remove
		collection.Fields.RemoveById("dgewwftr")

		// remove
		collection.Fields.RemoveById("ugtikydq")

		// remove
		collection.Fields.RemoveById("gal6bffx")

		// remove
		collection.Fields.RemoveById("kimwtlps")

		// remove
		collection.Fields.RemoveById("80vmdusd")

		// remove
		collection.Fields.RemoveById("mrq3rozi")

		// remove
		collection.Fields.RemoveById("i4dy54g8")

		// add
		new_trail_id := &core.JSONField{}
		if err := json.Unmarshal([]byte(`{
			"system": false,
			"id": "nsakta5t",
			"name": "trail_id",
			"type": "json",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"maxSize": 1
			}
		}`), new_trail_id); err != nil {
			return err
		}
		collection.Fields.Add(new_trail_id)

		// add
		new_date := &core.JSONField{}
		if err := json.Unmarshal([]byte(`{
			"system": false,
			"id": "oahimwtf",
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
			"id": "t9nlksmj",
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
			"id": "fzfiex8c",
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
			"id": "1vk5125z",
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
			"id": "skmgodvs",
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
			"id": "5rkwvgcy",
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
			"id": "xsuyuifg",
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
			"id": "0hsfxjcm",
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
			"id": "izyphasx",
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
			"id": "bffz6a6t",
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
			"id": "2xlfylnh",
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

		collection.ViewQuery = "SELECT id,trail_id,date,name,description,gpx,author,photos,distance,duration,elevation_gain,elevation_loss,created,type \nFROM (\n    SELECT summit_logs.id,trails.id as trail_id,summit_logs.date,trails.name,text as description,summit_logs.gpx,summit_logs.author,summit_logs.photos,summit_logs.distance,summit_logs.duration,summit_logs.elevation_gain,summit_logs.elevation_loss,summit_logs.created,\"summit_log\" as type \n    FROM summit_logs\n    LEFT JOIN trails ON summit_logs.id IN (\n    SELECT value\n    FROM json_each(trails.summit_logs)\n    )\n    UNION\n    SELECT id,id as trail_id,date,name,description,gpx,author,photos,distance,duration,elevation_gain,elevation_loss,created,\"trail\" as type \n    FROM trails\n)\nORDER BY date DESC"

		// add
		del_trail_id := &core.JSONField{}
		if err := json.Unmarshal([]byte(`{
			"system": false,
			"id": "mpcwfh4x",
			"name": "trail_id",
			"type": "json",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"maxSize": 1
			}
		}`), del_trail_id); err != nil {
			return err
		}
		collection.Fields.Add(del_trail_id)

		// add
		del_date := &core.JSONField{}
		if err := json.Unmarshal([]byte(`{
			"system": false,
			"id": "xsxx8b8a",
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
			"id": "4ybgblqj",
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
			"id": "f7y3unyx",
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
			"id": "cowkmcyy",
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
			"id": "dgewwftr",
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
			"id": "ugtikydq",
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
			"id": "gal6bffx",
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
			"id": "kimwtlps",
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
			"id": "80vmdusd",
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
			"id": "mrq3rozi",
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
			"id": "i4dy54g8",
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
		collection.Fields.RemoveById("nsakta5t")

		// remove
		collection.Fields.RemoveById("oahimwtf")

		// remove
		collection.Fields.RemoveById("t9nlksmj")

		// remove
		collection.Fields.RemoveById("fzfiex8c")

		// remove
		collection.Fields.RemoveById("1vk5125z")

		// remove
		collection.Fields.RemoveById("skmgodvs")

		// remove
		collection.Fields.RemoveById("5rkwvgcy")

		// remove
		collection.Fields.RemoveById("xsuyuifg")

		// remove
		collection.Fields.RemoveById("0hsfxjcm")

		// remove
		collection.Fields.RemoveById("izyphasx")

		// remove
		collection.Fields.RemoveById("bffz6a6t")

		// remove
		collection.Fields.RemoveById("2xlfylnh")

		return app.Save(collection)
	})
}
