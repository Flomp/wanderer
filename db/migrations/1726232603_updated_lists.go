package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
	"github.com/pocketbase/pocketbase/tools/types"
)

func init() {
	m.Register(func(app core.App) error {

		collection, err := app.FindCollectionByNameOrId("r6gu2ajyidy1x69")
		if err != nil {
			return err
		}

		collection.ListRule = types.Pointer("author = @request.auth.id || (@request.auth.id != \"\" && list_share_via_list.user ?= @request.auth.id)")

		collection.ViewRule = types.Pointer("author = @request.auth.id || (@request.auth.id != \"\" && list_share_via_list.user ?= @request.auth.id)")

		collection.CreateRule = types.Pointer("@request.auth.id != \"\" && (@request.data.author = @request.auth.id)")

		collection.UpdateRule = types.Pointer("author = @request.auth.id || (@request.auth.id != \"\" && list_share_via_list.list = id && list_share_via_list.user ?= @request.auth.id && list_share_via_list.permission = \"edit\")")

		return app.Save(collection)
	}, func(app core.App) error {

		collection, err := app.FindCollectionByNameOrId("r6gu2ajyidy1x69")
		if err != nil {
			return err
		}

		collection.ListRule = types.Pointer("author = @request.auth.id ")

		collection.ViewRule = types.Pointer("author = @request.auth.id ")

		collection.CreateRule = types.Pointer("author = @request.auth.id ")

		collection.UpdateRule = types.Pointer("author = @request.auth.id ")

		return app.Save(collection)
	})
}
