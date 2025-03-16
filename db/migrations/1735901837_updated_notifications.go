package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
	"github.com/pocketbase/pocketbase/tools/types"
)

func init() {
	m.Register(func(app core.App) error {

		collection, err := app.FindCollectionByNameOrId("khrcci2uqknny8h")
		if err != nil {
			return err
		}

		collection.UpdateRule = types.Pointer("@request.auth.id = recipient && (@request.data.type = null||@request.data.type = type) && (@request.data.metadata = null||@request.data.metadata = metadata) && (@request.data.recipient = null||@request.data.recipient = recipient) && (@request.data.author = null||@request.data.author = author) && @request.data.seen = true")

		return app.Save(collection)
	}, func(app core.App) error {

		collection, err := app.FindCollectionByNameOrId("khrcci2uqknny8h")
		if err != nil {
			return err
		}

		collection.UpdateRule = types.Pointer("@request.auth.id = recipient && @request.data.type = type && @request.data.metadata = metadata && @request.data.recipient = recipient && @request.data.author = author && @request.data.seen = true")

		return app.Save(collection)
	})
}
