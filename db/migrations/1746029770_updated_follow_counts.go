package migrations

import (
	"encoding/json"

	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		collection, err := app.FindCollectionByNameOrId("j6w72f0kb5ivd7x")
		if err != nil {
			return err
		}

		// update collection data
		if err := json.Unmarshal([]byte(`{
			"viewQuery": "SELECT \n    users.id,\n    COALESCE(followers.count, 0) AS followers,\n    COALESCE(following.count, 0) AS following\nFROM users\nLEFT JOIN (\n    SELECT followee AS user_id, COUNT(*) AS count\n    FROM follows\n    GROUP BY followee\n) AS followers ON users.id = followers.user_id\nLEFT JOIN (\n    SELECT follower AS user_id, COUNT(*) AS count\n    FROM follows\n    GROUP BY follower\n) AS following ON users.id = following.user_id"
		}`), &collection); err != nil {
			return err
		}

		// add field
		if err := collection.Fields.AddMarshaledJSONAt(1, []byte(`{
			"hidden": false,
			"id": "json2215181735",
			"maxSize": 1,
			"name": "followers",
			"presentable": false,
			"required": false,
			"system": false,
			"type": "json"
		}`)); err != nil {
			return err
		}

		// add field
		if err := collection.Fields.AddMarshaledJSONAt(2, []byte(`{
			"hidden": false,
			"id": "json1908379107",
			"maxSize": 1,
			"name": "following",
			"presentable": false,
			"required": false,
			"system": false,
			"type": "json"
		}`)); err != nil {
			return err
		}

		return app.Save(collection)
	}, func(app core.App) error {
		collection, err := app.FindCollectionByNameOrId("j6w72f0kb5ivd7x")
		if err != nil {
			return err
		}

		// update collection data
		if err := json.Unmarshal([]byte(`{
			"viewQuery": "SELECT \n    users.id\n    \nFROM users\n"
		}`), &collection); err != nil {
			return err
		}

		// remove field
		collection.Fields.RemoveById("json2215181735")

		// remove field
		collection.Fields.RemoveById("json1908379107")

		return app.Save(collection)
	})
}
