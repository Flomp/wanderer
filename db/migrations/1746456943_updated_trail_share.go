package migrations

import (
	"encoding/json"

	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		collection, err := app.FindCollectionByNameOrId("1mns8mlal6uf9ku")
		if err != nil {
			return err
		}

		// update collection data
		if err := json.Unmarshal([]byte(`{
			"createRule": "trail.author.user = @request.auth.id",
			"deleteRule": "trail.author.user = @request.auth.id",
			"listRule": "trail.author.user = @request.auth.id || user = @request.auth.id",
			"updateRule": "trail.author.user = @request.auth.id",
			"viewRule": "trail.author.user = @request.auth.id || user = @request.auth.id"
		}`), &collection); err != nil {
			return err
		}

		return app.Save(collection)
	}, func(app core.App) error {
		collection, err := app.FindCollectionByNameOrId("1mns8mlal6uf9ku")
		if err != nil {
			return err
		}

		// update collection data
		if err := json.Unmarshal([]byte(`{
			"createRule": "trail.author = @request.auth.id",
			"deleteRule": "trail.author = @request.auth.id",
			"listRule": "trail.author = @request.auth.id || user = @request.auth.id",
			"updateRule": "trail.author = @request.auth.id",
			"viewRule": "trail.author = @request.auth.id || user = @request.auth.id"
		}`), &collection); err != nil {
			return err
		}

		return app.Save(collection)
	})
}
