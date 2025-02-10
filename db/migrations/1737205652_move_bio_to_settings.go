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
		dao := daos.New(db)

		// remove bio from users_anonymous to prevent ambiguity
		uaCollection, err := dao.FindCollectionByNameOrId("xku110v5a5xbufa")
		if err != nil {
			return err
		}

		options := map[string]any{}
		if err := json.Unmarshal([]byte(`{
			"query": "SELECT users.id, username, avatar, users.created, CAST(COALESCE(json_extract(privacy, '$.account') = 'private', false) as BOOL) as private FROM users LEFT JOIN settings ON settings.user = users.id"
		}`), &options); err != nil {
			return err
		}
		uaCollection.SetOptions(options)

		err = dao.SaveCollection(uaCollection)
		if err != nil {
			return err
		}

		// add bio field to settings
		collection, err := dao.FindCollectionByNameOrId("settings")
		if err != nil {
			return err
		}

		new_bio := &schema.SchemaField{}
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
		collection.Schema.AddField(new_bio)

		err = dao.SaveCollection(collection)
		if err != nil {
			return err
		}

		// migrate existing bios
		users, err := dao.FindRecordsByFilter("_pb_users_auth_", "bio != null", "", -1, 0)
		if err != nil {
			return nil
		}

		for _, user := range users {
			bio := user.GetString(("bio"))
			userSettings, err := dao.FindRecordsByFilter("settings", "user = {:userId}", "", 1, 0, dbx.Params{"userId": user.Id})
			if err != nil {
				return err
			}

			userSettings[0].Set("bio", bio)

			if err := dao.SaveRecord(userSettings[0]); err != nil {
				return err
			}
		}

		// remove bio from users
		uCollection, err := dao.FindCollectionByNameOrId("_pb_users_auth_")
		if err != nil {
			return err
		}

		uCollection.Schema.RemoveField("pd2cq8sq")
		err = dao.SaveCollection(uCollection)
		if err != nil {
			return err
		}

		// add bio back to users_anaonymous
		if err := json.Unmarshal([]byte(`{
			"query": "SELECT users.id, username, avatar, bio, users.created, CAST(COALESCE(json_extract(privacy, '$.account') = 'private', false) as BOOL) as private FROM users LEFT JOIN settings ON settings.user = users.id"
		}`), &options); err != nil {
			return err
		}
		uaCollection.SetOptions(options)

		err = dao.SaveCollection(uaCollection)
		if err != nil {
			return err
		}

		return nil
	}, func(db dbx.Builder) error {
		dao := daos.New(db)

		uaCollection, err := dao.FindCollectionByNameOrId("xku110v5a5xbufa")
		if err != nil {
			return err
		}

		options := map[string]any{}
		if err := json.Unmarshal([]byte(`{
			"query": "SELECT users.id, username, avatar, users.created, CAST(COALESCE(json_extract(privacy, '$.account') = 'private', false) as BOOL) as private FROM users LEFT JOIN settings ON settings.user = users.id"
		}`), &options); err != nil {
			return err
		}
		uaCollection.SetOptions(options)

		err = dao.SaveCollection(uaCollection)
		if err != nil {
			return err
		}

		collection, err := dao.FindCollectionByNameOrId("settings")
		if err != nil {
			return err
		}

		// remove
		collection.Schema.RemoveField("pd2cq8sq")

		err = dao.SaveCollection(collection)
		if err != nil {
			return err
		}

		uCollection, err := dao.FindCollectionByNameOrId("_pb_users_auth_")
		if err != nil {
			return err
		}

		// add
		new_bio := &schema.SchemaField{}
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
		uCollection.Schema.AddField(new_bio)

		err = dao.SaveCollection(uCollection)
		if err != nil {
			return err
		}

		if err := json.Unmarshal([]byte(`{
			"query": "SELECT users.id, username, avatar, bio, users.created, CAST(COALESCE(json_extract(privacy, '$.account') = 'private', false) as BOOL) as private FROM users LEFT JOIN settings ON settings.user = users.id"
		}`), &options); err != nil {
			return err
		}
		uaCollection.SetOptions(options)

		return dao.SaveCollection(uaCollection)

	})
}
