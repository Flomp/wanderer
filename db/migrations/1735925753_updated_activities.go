package migrations

import (
	"encoding/json"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/daos"
	m "github.com/pocketbase/pocketbase/migrations"
	"github.com/pocketbase/pocketbase/models/schema"
	"github.com/pocketbase/pocketbase/tools/types"
)

func init() {
	m.Register(func(db dbx.Builder) error {
		dao := daos.New(db);

		collection, err := dao.FindCollectionByNameOrId("t9lphichi5xwyeu")
		if err != nil {
			return err
		}

		collection.ListRule = types.Pointer("(\n    @request.auth.id = author || \n    (@collection.users_anonymous.id ?= author && @collection.users_anonymous.private ?= false)\n)\n&&\n(\n    @collection.trails.id ?= trail_id && \n    (\n        @collection.trails.author ?= @request.auth.id ||\n        @collection.trails.public ?= true || \n        (@request.auth.id != \"\" && @collection.trails.trail_share_via_trail.user ?= @request.auth.id)\n    )\n)")

		collection.ViewRule = types.Pointer("(\n    @request.auth.id = author || \n    (@collection.users_anonymous.id ?= author && @collection.users_anonymous.private ?= false)\n)\n&&\n(\n    @collection.trails.id ?= trail_id && \n    (\n        @collection.trails.author ?= @request.auth.id ||\n        @collection.trails.public ?= true || \n        (@request.auth.id != \"\" && @collection.trails.trail_share_via_trail.user ?= @request.auth.id)\n    )\n)")

		options := map[string]any{}
		if err := json.Unmarshal([]byte(`{
			"query": "SELECT id,trail_id,date,name,description,gpx,author,photos,distance,duration,elevation_gain,elevation_loss,created,type \nFROM (\n    SELECT summit_logs.id,trails.id as trail_id,summit_logs.date,trails.name,text as description,summit_logs.gpx,summit_logs.author,summit_logs.photos,summit_logs.distance,summit_logs.duration,summit_logs.elevation_gain,summit_logs.elevation_loss,summit_logs.created,\"summit_log\" as type \n    FROM summit_logs\n    JOIN trails ON summit_logs.id IN (\n    SELECT value\n    FROM json_each(trails.summit_logs)\n    )\n    UNION\n    SELECT id,id as trail_id,date,name,description,gpx,author,photos,distance,duration,elevation_gain,elevation_loss,created,\"trail\" as type \n    FROM trails\n)\nORDER BY created DESC"
		}`), &options); err != nil {
			return err
		}
		collection.SetOptions(options)

		// remove
		collection.Schema.RemoveField("1gmbao3t")

		// remove
		collection.Schema.RemoveField("pvlr9fv1")

		// remove
		collection.Schema.RemoveField("ueqomhey")

		// remove
		collection.Schema.RemoveField("chjcbrid")

		// remove
		collection.Schema.RemoveField("2wwngiu1")

		// remove
		collection.Schema.RemoveField("vfbs4wij")

		// remove
		collection.Schema.RemoveField("kpitkmnj")

		// remove
		collection.Schema.RemoveField("90almjd5")

		// remove
		collection.Schema.RemoveField("49io1roa")

		// remove
		collection.Schema.RemoveField("kngt1gs1")

		// remove
		collection.Schema.RemoveField("faei1oos")

		// remove
		collection.Schema.RemoveField("fyeket06")

		// add
		new_trail_id := &schema.SchemaField{}
		if err := json.Unmarshal([]byte(`{
			"system": false,
			"id": "c9vvjnme",
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
			"id": "flg0ahfn",
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
			"id": "e6lik1pd",
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
			"id": "hxfpjwhr",
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
			"id": "t4vcxvux",
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
			"id": "cpnvh0me",
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
			"id": "wsh8ppyd",
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
			"id": "fkd0ua94",
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
			"id": "cj05f9if",
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
			"id": "kjvv3j2w",
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
			"id": "pob1he5w",
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
			"id": "rd0cyqkb",
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

		collection.ListRule = types.Pointer("(\n    @request.auth.id = author || \n    (@collection.users_anonymous.id ?= author && @collection.users_anonymous.private ?= false)\n)\n&&\n(\n    @collection.trails.id ?= trail_id && \n    (\n        @collection.trails.author = @request.auth.id ||\n        @collection.trails.public = true || \n        (@request.auth.id != \"\" && @collection.trails.trail_share_via_trail.user ?= @request.auth.id)\n    )\n)")

		collection.ViewRule = types.Pointer("(\n    @request.auth.id = author || \n    (@collection.users_anonymous.id ?= author && @collection.users_anonymous.private ?= false)\n)\n&&\n(\n    @collection.trails.id ?= trail_id && \n    (\n        @collection.trails.author = @request.auth.id ||\n        @collection.trails.public = true || \n        (@request.auth.id != \"\" && @collection.trails.trail_share_via_trail.user ?= @request.auth.id)\n    )\n)")

		options := map[string]any{}
		if err := json.Unmarshal([]byte(`{
			"query": "SELECT id,trail_id,date,name,description,gpx,author,photos,distance,duration,elevation_gain,elevation_loss,created,type \nFROM (\n    SELECT summit_logs.id,trails.id as trail_id,summit_logs.date,trails.name,text as description,summit_logs.gpx,summit_logs.author,summit_logs.photos,summit_logs.distance,summit_logs.duration,summit_logs.elevation_gain,summit_logs.elevation_loss,summit_logs.created,\"summit_log\" as type \n    FROM summit_logs\n    JOIN trails ON summit_logs.id IN (\n    SELECT value\n    FROM json_each(trails.summit_logs)\n    )\n    UNION\n    SELECT id,id as trail_id,date,name,description,gpx,author,photos,distance,duration,elevation_gain,elevation_loss,created,\"trail\" as type \n    FROM trails\n)\nORDER BY date DESC"
		}`), &options); err != nil {
			return err
		}
		collection.SetOptions(options)

		// add
		del_trail_id := &schema.SchemaField{}
		if err := json.Unmarshal([]byte(`{
			"system": false,
			"id": "1gmbao3t",
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
			"id": "pvlr9fv1",
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
			"id": "ueqomhey",
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
			"id": "chjcbrid",
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
			"id": "2wwngiu1",
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
			"id": "vfbs4wij",
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
			"id": "kpitkmnj",
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
			"id": "90almjd5",
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
			"id": "49io1roa",
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
			"id": "kngt1gs1",
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
			"id": "faei1oos",
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
			"id": "fyeket06",
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
		collection.Schema.RemoveField("c9vvjnme")

		// remove
		collection.Schema.RemoveField("flg0ahfn")

		// remove
		collection.Schema.RemoveField("e6lik1pd")

		// remove
		collection.Schema.RemoveField("hxfpjwhr")

		// remove
		collection.Schema.RemoveField("t4vcxvux")

		// remove
		collection.Schema.RemoveField("cpnvh0me")

		// remove
		collection.Schema.RemoveField("wsh8ppyd")

		// remove
		collection.Schema.RemoveField("fkd0ua94")

		// remove
		collection.Schema.RemoveField("cj05f9if")

		// remove
		collection.Schema.RemoveField("kjvv3j2w")

		// remove
		collection.Schema.RemoveField("pob1he5w")

		// remove
		collection.Schema.RemoveField("rd0cyqkb")

		return dao.SaveCollection(collection)
	})
}
