package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
	"github.com/pocketbase/pocketbase/tools/types"
)

func init() {
	m.Register(func(app core.App) error {

		collection, err := app.FindCollectionByNameOrId("_pb_users_auth_")
		if err != nil {
			return err
		}

		collection.ListRule = types.Pointer("@request.auth.id != \"\"")

		return app.Save(collection)
	}, func(app core.App) error {

		collection, err := app.FindCollectionByNameOrId("_pb_users_auth_")
		if err != nil {
			return err
		}

		collection.ListRule = types.Pointer("id = @request.auth.id")

		return app.Save(collection)
	})
}
