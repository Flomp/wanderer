package migrations

import (
	"encoding/json"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/daos"
	m "github.com/pocketbase/pocketbase/migrations"
	"github.com/pocketbase/pocketbase/models/schema"
)

func init() {
	m.Register(func(db dbx.Builder) error {
		dao := daos.New(db);

		collection, err := dao.FindCollectionByNameOrId("t9lphichi5xwyeu")
		if err != nil {
			return err
		}

		options := map[string]any{}
		if err := json.Unmarshal([]byte(`{
			"query": "SELECT id,trail_id,date,name,description,gpx,author,photos,distance,duration,elevation_gain,elevation_loss,created,type \nFROM (\n    SELECT summit_logs.id,trails.id as trail_id,summit_logs.date,trails.name,text as description,summit_logs.gpx,summit_logs.author,summit_logs.photos,summit_logs.distance,summit_logs.duration,summit_logs.elevation_gain,summit_logs.elevation_loss,summit_logs.created,\"summit_log\" as type \n    FROM summit_logs\n    JOIN trails ON summit_logs.id IN (\n    SELECT value\n    FROM json_each(trails.summit_logs)\n    )\n    UNION\n    SELECT id,id as trail_id,date,name,description,gpx,author,photos,distance,duration,elevation_gain,elevation_loss,created,\"trail\" as type \n    FROM trails\n)\nORDER BY date DESC"
		}`), &options); err != nil {
			return err
		}
		collection.SetOptions(options)

		// remove
		collection.Schema.RemoveField("mpcwfh4x")

		// remove
		collection.Schema.RemoveField("xsxx8b8a")

		// remove
		collection.Schema.RemoveField("4ybgblqj")

		// remove
		collection.Schema.RemoveField("f7y3unyx")

		// remove
		collection.Schema.RemoveField("cowkmcyy")

		// remove
		collection.Schema.RemoveField("dgewwftr")

		// remove
		collection.Schema.RemoveField("ugtikydq")

		// remove
		collection.Schema.RemoveField("gal6bffx")

		// remove
		collection.Schema.RemoveField("kimwtlps")

		// remove
		collection.Schema.RemoveField("80vmdusd")

		// remove
		collection.Schema.RemoveField("mrq3rozi")

		// remove
		collection.Schema.RemoveField("i4dy54g8")

		// add
		new_trail_id := &schema.SchemaField{}
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
		collection.Schema.AddField(new_trail_id)

		// add
		new_date := &schema.SchemaField{}
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
		collection.Schema.AddField(new_date)

		// add
		new_name := &schema.SchemaField{}
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
		collection.Schema.AddField(new_name)

		// add
		new_description := &schema.SchemaField{}
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
		collection.Schema.AddField(new_description)

		// add
		new_gpx := &schema.SchemaField{}
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
		collection.Schema.AddField(new_gpx)

		// add
		new_author := &schema.SchemaField{}
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
		collection.Schema.AddField(new_author)

		// add
		new_photos := &schema.SchemaField{}
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
		collection.Schema.AddField(new_photos)

		// add
		new_distance := &schema.SchemaField{}
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
		collection.Schema.AddField(new_distance)

		// add
		new_duration := &schema.SchemaField{}
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
		collection.Schema.AddField(new_duration)

		// add
		new_elevation_gain := &schema.SchemaField{}
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
		collection.Schema.AddField(new_elevation_gain)

		// add
		new_elevation_loss := &schema.SchemaField{}
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
		collection.Schema.AddField(new_elevation_loss)

		// add
		new_type := &schema.SchemaField{}
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
		collection.Schema.AddField(new_type)

		return dao.SaveCollection(collection)
	}, func(db dbx.Builder) error {
		dao := daos.New(db);

		collection, err := dao.FindCollectionByNameOrId("t9lphichi5xwyeu")
		if err != nil {
			return err
		}

		options := map[string]any{}
		if err := json.Unmarshal([]byte(`{
			"query": "SELECT id,trail_id,date,name,description,gpx,author,photos,distance,duration,elevation_gain,elevation_loss,created,type \nFROM (\n    SELECT summit_logs.id,trails.id as trail_id,summit_logs.date,trails.name,text as description,summit_logs.gpx,summit_logs.author,summit_logs.photos,summit_logs.distance,summit_logs.duration,summit_logs.elevation_gain,summit_logs.elevation_loss,summit_logs.created,\"summit_log\" as type \n    FROM summit_logs\n    LEFT JOIN trails ON summit_logs.id IN (\n    SELECT value\n    FROM json_each(trails.summit_logs)\n    )\n    UNION\n    SELECT id,id as trail_id,date,name,description,gpx,author,photos,distance,duration,elevation_gain,elevation_loss,created,\"trail\" as type \n    FROM trails\n)\nORDER BY date DESC"
		}`), &options); err != nil {
			return err
		}
		collection.SetOptions(options)

		// add
		del_trail_id := &schema.SchemaField{}
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
		collection.Schema.AddField(del_trail_id)

		// add
		del_date := &schema.SchemaField{}
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
		collection.Schema.AddField(del_date)

		// add
		del_name := &schema.SchemaField{}
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
		collection.Schema.AddField(del_name)

		// add
		del_description := &schema.SchemaField{}
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
		collection.Schema.AddField(del_description)

		// add
		del_gpx := &schema.SchemaField{}
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
		collection.Schema.AddField(del_gpx)

		// add
		del_author := &schema.SchemaField{}
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
		collection.Schema.AddField(del_author)

		// add
		del_photos := &schema.SchemaField{}
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
		collection.Schema.AddField(del_photos)

		// add
		del_distance := &schema.SchemaField{}
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
		collection.Schema.AddField(del_distance)

		// add
		del_duration := &schema.SchemaField{}
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
		collection.Schema.AddField(del_duration)

		// add
		del_elevation_gain := &schema.SchemaField{}
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
		collection.Schema.AddField(del_elevation_gain)

		// add
		del_elevation_loss := &schema.SchemaField{}
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
		collection.Schema.AddField(del_elevation_loss)

		// add
		del_type := &schema.SchemaField{}
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
		collection.Schema.AddField(del_type)

		// remove
		collection.Schema.RemoveField("nsakta5t")

		// remove
		collection.Schema.RemoveField("oahimwtf")

		// remove
		collection.Schema.RemoveField("t9nlksmj")

		// remove
		collection.Schema.RemoveField("fzfiex8c")

		// remove
		collection.Schema.RemoveField("1vk5125z")

		// remove
		collection.Schema.RemoveField("skmgodvs")

		// remove
		collection.Schema.RemoveField("5rkwvgcy")

		// remove
		collection.Schema.RemoveField("xsuyuifg")

		// remove
		collection.Schema.RemoveField("0hsfxjcm")

		// remove
		collection.Schema.RemoveField("izyphasx")

		// remove
		collection.Schema.RemoveField("bffz6a6t")

		// remove
		collection.Schema.RemoveField("2xlfylnh")

		return dao.SaveCollection(collection)
	})
}
