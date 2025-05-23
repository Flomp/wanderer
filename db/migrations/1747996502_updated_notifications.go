package migrations

import (
	"encoding/json"

	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		collection, err := app.FindCollectionByNameOrId("khrcci2uqknny8h")
		if err != nil {
			return err
		}

		// update collection data
		if err := json.Unmarshal([]byte(`{
			"listRule": "@request.auth.id = recipient.user",
			"updateRule": "@request.auth.id = recipient.user && (@request.body.type = null||@request.body.type = type) && (@request.body.metadata = null||@request.body.metadata = metadata) && (@request.body.recipient = null||@request.body.recipient = recipient) && (@request.body.author = null||@request.body.author = author) && @request.body.seen = true",
			"viewRule": "@request.auth.id = recipient.user"
		}`), &collection); err != nil {
			return err
		}

		return app.Save(collection)
	}, func(app core.App) error {
		collection, err := app.FindCollectionByNameOrId("khrcci2uqknny8h")
		if err != nil {
			return err
		}

		// update collection data
		if err := json.Unmarshal([]byte(`{
			"listRule": "@request.auth.id = recipient",
			"updateRule": "@request.auth.id = recipient && (@request.body.type = null||@request.body.type = type) && (@request.body.metadata = null||@request.body.metadata = metadata) && (@request.body.recipient = null||@request.body.recipient = recipient) && (@request.body.author = null||@request.body.author = author) && @request.body.seen = true",
			"viewRule": "@request.auth.id = recipient"
		}`), &collection); err != nil {
			return err
		}

		return app.Save(collection)
	})
}
