package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
	"github.com/pocketbase/pocketbase/tools/types"
)

func init() {
	m.Register(func(app core.App) error {

		collection, err := app.FindCollectionByNameOrId("dd2l9a4vxpy2ni8")
		if err != nil {
			return err
		}

		collection.ListRule = types.Pointer("(@request.auth.id != \"\" && @collection.trails.summit_logs.id ?= id && @collection.trails.author ?= @request.auth.id) || (@collection.trails.summit_logs.id ?= id && @collection.trails.public = true)")

		collection.ViewRule = types.Pointer("(@request.auth.id != \"\" && @collection.trails.summit_logs.id ?= id && @collection.trails.author ?= @request.auth.id) || (@collection.trails.summit_logs.id ?= id && @collection.trails.public = true)")

		return app.Save(collection)
	}, func(app core.App) error {

		collection, err := app.FindCollectionByNameOrId("dd2l9a4vxpy2ni8")
		if err != nil {
			return err
		}

		collection.ListRule = types.Pointer("(@request.auth.id != \"\" && @collection.trails.summit_logs.id ?= id && @collection.trails.author ?= @request.auth.id) || @collection.trails.summit_logs.id ?= id && @collection.trails.public = true")

		collection.ViewRule = types.Pointer("(@request.auth.id != \"\" && @collection.trails.summit_logs.id ?= id && @collection.trails.author ?= @request.auth.id) || @collection.trails.summit_logs.id ?= id && @collection.trails.public = true")

		return app.Save(collection)
	})
}
