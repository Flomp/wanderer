package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
	"github.com/pocketbase/pocketbase/tools/types"
)

func init() {
	m.Register(func(app core.App) error {

		collection, err := app.FindCollectionByNameOrId("1mns8mlal6uf9ku")
		if err != nil {
			return err
		}

		collection.ListRule = types.Pointer("trail.author = @request.auth.id || user = @request.auth.id")

		collection.ViewRule = types.Pointer("trail.author = @request.auth.id || user = @request.auth.id")

		return app.Save(collection)
	}, func(app core.App) error {

		collection, err := app.FindCollectionByNameOrId("1mns8mlal6uf9ku")
		if err != nil {
			return err
		}

		collection.ListRule = types.Pointer("trail.author = @request.auth.id")

		collection.ViewRule = types.Pointer("trail.author = @request.auth.id")

		return app.Save(collection)
	})
}
