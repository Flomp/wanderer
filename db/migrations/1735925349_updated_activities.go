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

		collection.ListRule = types.Pointer("(\n    @request.auth.id = author || \n    (@collection.users_anonymous.id ?= author && @collection.users_anonymous.private ?= false)\n)\n&&\n(\n    @collection.trails.id ?= trail_id && \n    (\n        @collection.trails.author = @request.auth.id ||\n        @collection.trails.public = true || \n        (@request.auth.id != \"\" && @collection.trails.trail_share_via_trail.user ?= @request.auth.id)\n    )\n)")

		collection.ViewRule = types.Pointer("(\n    @request.auth.id = author || \n    (@collection.users_anonymous.id ?= author && @collection.users_anonymous.private ?= false)\n)\n&&\n(\n    @collection.trails.id ?= trail_id && \n    (\n        @collection.trails.author = @request.auth.id ||\n        @collection.trails.public = true || \n        (@request.auth.id != \"\" && @collection.trails.trail_share_via_trail.user ?= @request.auth.id)\n    )\n)")

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

		// add
		new_trail_id := &core.JSONField{}
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
		}`), new_trail_id); err != nil {
			return err
		}
		collection.Fields.Add(new_trail_id)

		// add
		new_date := &core.JSONField{}
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
		}`), new_date); err != nil {
			return err
		}
		collection.Fields.Add(new_date)

		// add
		new_name := &core.JSONField{}
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
		}`), new_name); err != nil {
			return err
		}
		collection.Fields.Add(new_name)

		// add
		new_description := &core.JSONField{}
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
		}`), new_description); err != nil {
			return err
		}
		collection.Fields.Add(new_description)

		// add
		new_gpx := &core.JSONField{}
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
		}`), new_gpx); err != nil {
			return err
		}
		collection.Fields.Add(new_gpx)

		// add
		new_author := &core.JSONField{}
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
		}`), new_author); err != nil {
			return err
		}
		collection.Fields.Add(new_author)

		// add
		new_photos := &core.JSONField{}
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
		}`), new_photos); err != nil {
			return err
		}
		collection.Fields.Add(new_photos)

		// add
		new_distance := &core.JSONField{}
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
		}`), new_distance); err != nil {
			return err
		}
		collection.Fields.Add(new_distance)

		// add
		new_duration := &core.JSONField{}
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
		}`), new_duration); err != nil {
			return err
		}
		collection.Fields.Add(new_duration)

		// add
		new_elevation_gain := &core.JSONField{}
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
		}`), new_elevation_gain); err != nil {
			return err
		}
		collection.Fields.Add(new_elevation_gain)

		// add
		new_elevation_loss := &core.JSONField{}
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
		}`), new_elevation_loss); err != nil {
			return err
		}
		collection.Fields.Add(new_elevation_loss)

		// add
		new_type := &core.JSONField{}
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

		collection.ListRule = types.Pointer("@request.auth.id = author || (@collection.users_anonymous.id ?= author && @collection.users_anonymous.private ?= false)")

		collection.ViewRule = types.Pointer("@request.auth.id = author || (@collection.users_anonymous.id ?= author && @collection.users_anonymous.private ?= false)")

		// add
		del_trail_id := &core.JSONField{}
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
		}`), del_trail_id); err != nil {
			return err
		}
		collection.Fields.Add(del_trail_id)

		// add
		del_date := &core.JSONField{}
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
		}`), del_date); err != nil {
			return err
		}
		collection.Fields.Add(del_date)

		// add
		del_name := &core.JSONField{}
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
		}`), del_name); err != nil {
			return err
		}
		collection.Fields.Add(del_name)

		// add
		del_description := &core.JSONField{}
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
		}`), del_description); err != nil {
			return err
		}
		collection.Fields.Add(del_description)

		// add
		del_gpx := &core.JSONField{}
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
		}`), del_gpx); err != nil {
			return err
		}
		collection.Fields.Add(del_gpx)

		// add
		del_author := &core.JSONField{}
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
		}`), del_author); err != nil {
			return err
		}
		collection.Fields.Add(del_author)

		// add
		del_photos := &core.JSONField{}
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
		}`), del_photos); err != nil {
			return err
		}
		collection.Fields.Add(del_photos)

		// add
		del_distance := &core.JSONField{}
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
		}`), del_distance); err != nil {
			return err
		}
		collection.Fields.Add(del_distance)

		// add
		del_duration := &core.JSONField{}
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
		}`), del_duration); err != nil {
			return err
		}
		collection.Fields.Add(del_duration)

		// add
		del_elevation_gain := &core.JSONField{}
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
		}`), del_elevation_gain); err != nil {
			return err
		}
		collection.Fields.Add(del_elevation_gain)

		// add
		del_elevation_loss := &core.JSONField{}
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
		}`), del_elevation_loss); err != nil {
			return err
		}
		collection.Fields.Add(del_elevation_loss)

		// add
		del_type := &core.JSONField{}
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
		}`), del_type); err != nil {
			return err
		}
		collection.Fields.Add(del_type)

		// remove
		collection.Fields.RemoveById("1gmbao3t")

		// remove
		collection.Fields.RemoveById("pvlr9fv1")

		// remove
		collection.Fields.RemoveById("ueqomhey")

		// remove
		collection.Fields.RemoveById("chjcbrid")

		// remove
		collection.Fields.RemoveById("2wwngiu1")

		// remove
		collection.Fields.RemoveById("vfbs4wij")

		// remove
		collection.Fields.RemoveById("kpitkmnj")

		// remove
		collection.Fields.RemoveById("90almjd5")

		// remove
		collection.Fields.RemoveById("49io1roa")

		// remove
		collection.Fields.RemoveById("kngt1gs1")

		// remove
		collection.Fields.RemoveById("faei1oos")

		// remove
		collection.Fields.RemoveById("fyeket06")

		return app.Save(collection)
	})
}
