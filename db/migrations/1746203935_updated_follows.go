package migrations

import (
	"encoding/json"

	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		collection, err := app.FindCollectionByNameOrId("8obn1ukumze565i")
		if err != nil {
			return err
		}

		// update collection data
		if err := json.Unmarshal([]byte(`{
			"deleteRule": "@request.auth.id = follower.user.id",
			"listRule": "@request.auth.id = follower.user.id || @request.auth.id = followee.user.id",
			"updateRule": "@request.auth.id = follower.user.id",
			"viewRule": "@request.auth.id = follower.user.id || @request.auth.id = followee.user.id"
		}`), &collection); err != nil {
			return err
		}

		return app.Save(collection)
	}, func(app core.App) error {
		collection, err := app.FindCollectionByNameOrId("8obn1ukumze565i")
		if err != nil {
			return err
		}

		// update collection data
		if err := json.Unmarshal([]byte(`{
			"deleteRule": "@request.auth.id = follower.id",
			"listRule": "@request.auth.id = follower.id || @request.auth.id = followee.id",
			"updateRule": "@request.auth.id = follower.id",
			"viewRule": "@request.auth.id = follower.id || @request.auth.id = followee.id"
		}`), &collection); err != nil {
			return err
		}

		return app.Save(collection)
	})
}
