package migrations

import (
	"encoding/json"

	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		collection, err := app.FindCollectionByNameOrId("1kot7t9na3hi0gl")
		if err != nil {
			return err
		}

		// update collection data
		if err := json.Unmarshal([]byte(`{
			"createRule": "list.author.user = @request.auth.id",
			"deleteRule": "list.author.user = @request.auth.id",
			"listRule": "list.author.user = @request.auth.id || actor.user = @request.auth.id",
			"updateRule": "list.author.user = @request.auth.id",
			"viewRule": "list.author.user = @request.auth.id || actor.user = @request.auth.id"
		}`), &collection); err != nil {
			return err
		}

		return app.Save(collection)
	}, func(app core.App) error {
		collection, err := app.FindCollectionByNameOrId("1kot7t9na3hi0gl")
		if err != nil {
			return err
		}

		// update collection data
		if err := json.Unmarshal([]byte(`{
			"createRule": "list.author = @request.auth.id",
			"deleteRule": "list.author = @request.auth.id",
			"listRule": "list.author = @request.auth.id || actor.user = @request.auth.id",
			"updateRule": "list.author = @request.auth.id",
			"viewRule": "list.author = @request.auth.id || actor.user = @request.auth.id"
		}`), &collection); err != nil {
			return err
		}

		return app.Save(collection)
	})
}
