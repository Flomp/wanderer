package migrations

import (
	"encoding/json"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {

		// remove bio from users_anonymous to prevent ambiguity
		uaCollection, err := app.FindCollectionByNameOrId("xku110v5a5xbufa")
		if err != nil {
			return err
		}

		uaCollection.ViewQuery = "SELECT users.id, username, avatar, users.created, CAST(COALESCE(json_extract(privacy, '$.account') = 'private', false) as BOOL) as private FROM users LEFT JOIN settings ON settings.user = users.id"

		err = app.Save(uaCollection)
		if err != nil {
			return err
		}

		// add bio field to settings
		collection, err := app.FindCollectionByNameOrId("settings")
		if err != nil {
			return err
		}

		new_bio := &core.TextField{}
		if err := json.Unmarshal([]byte(`{
			"system": false,
			"id": "pd2cq8sq",
			"name": "bio",
			"type": "text",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"min": null,
				"max": 10000,
				"pattern": ""
			}
		}`), new_bio); err != nil {
			return err
		}
		collection.Fields.Add(new_bio)

		err = app.Save(collection)
		if err != nil {
			return err
		}

		// migrate existing bios
		users, err := app.FindRecordsByFilter("_pb_users_auth_", "bio != null", "", -1, 0)
		if err != nil {
			return nil
		}

		for _, user := range users {
			bio := user.GetString(("bio"))
			userSettings, err := app.FindRecordsByFilter("settings", "user = {:userId}", "", 1, 0, dbx.Params{"userId": user.Id})
			if err != nil {
				return err
			}

			userSettings[0].Set("bio", bio)

			if err := app.Save(userSettings[0]); err != nil {
				return err
			}
		}

		// remove bio from users
		uCollection, err := app.FindCollectionByNameOrId("_pb_users_auth_")
		if err != nil {
			return err
		}

		uCollection.Fields.RemoveById("pd2cq8sq")
		err = app.Save(uCollection)
		if err != nil {
			return err
		}

		// add bio back to users_anaonymous
		uaCollection.ViewQuery = "SELECT users.id, username, avatar, bio, users.created, CAST(COALESCE(json_extract(privacy, '$.account') = 'private', false) as BOOL) as private FROM users LEFT JOIN settings ON settings.user = users.id"

		err = app.Save(uaCollection)
		if err != nil {
			return err
		}

		return nil
	}, func(app core.App) error {

		uaCollection, err := app.FindCollectionByNameOrId("xku110v5a5xbufa")
		if err != nil {
			return err
		}

		uaCollection.ViewQuery = "SELECT users.id, username, avatar, users.created, CAST(COALESCE(json_extract(privacy, '$.account') = 'private', false) as BOOL) as private FROM users LEFT JOIN settings ON settings.user = users.id"

		err = app.Save(uaCollection)
		if err != nil {
			return err
		}

		collection, err := app.FindCollectionByNameOrId("settings")
		if err != nil {
			return err
		}

		// remove
		collection.Fields.RemoveById("pd2cq8sq")

		err = app.Save(collection)
		if err != nil {
			return err
		}

		uCollection, err := app.FindCollectionByNameOrId("_pb_users_auth_")
		if err != nil {
			return err
		}

		// add
		new_bio := &core.TextField{}
		if err := json.Unmarshal([]byte(`{
			"system": false,
			"id": "pd2cq8sq",
			"name": "bio",
			"type": "text",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"min": null,
				"max": 10000,
				"pattern": ""
			}
		}`), new_bio); err != nil {
			return err
		}
		uCollection.Fields.Add(new_bio)

		err = app.Save(uCollection)
		if err != nil {
			return err
		}

		uaCollection.ViewQuery = "SELECT users.id, username, avatar, bio, users.created, CAST(COALESCE(json_extract(privacy, '$.account') = 'private', false) as BOOL) as private FROM users LEFT JOIN settings ON settings.user = users.id"

		return app.Save(uaCollection)

	})
}
