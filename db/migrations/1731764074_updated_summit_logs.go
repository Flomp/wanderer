package migrations

import (
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/daos"
	m "github.com/pocketbase/pocketbase/migrations"
	"github.com/pocketbase/pocketbase/tools/types"
)

func init() {
	m.Register(func(db dbx.Builder) error {
		dao := daos.New(db);

		collection, err := dao.FindCollectionByNameOrId("dd2l9a4vxpy2ni8")
		if err != nil {
			return err
		}

		collection.ListRule = types.Pointer("(@request.auth.id != \"\" && @collection.trails.summit_logs.id ?= id && @collection.trails.author ?= @request.auth.id) || (@collection.trails.summit_logs.id ?= id && @collection.trails.public ?= true) ||\n(@collection.trail_share.trail.summit_logs.id ?= id && @collection.trail_share.user ?= @request.auth.id) ||\n(author = @request.auth.id)")

		collection.ViewRule = types.Pointer("(@request.auth.id != \"\" && @collection.trails.summit_logs.id ?= id && @collection.trails.author ?= @request.auth.id) || (@collection.trails.summit_logs.id ?= id && @collection.trails.public ?= true) || \n(@collection.trail_share.trail.summit_logs.id ?= id && @collection.trail_share.user ?= @request.auth.id) ||\n(author = @request.auth.id)")

		return dao.SaveCollection(collection)
	}, func(db dbx.Builder) error {
		dao := daos.New(db);

		collection, err := dao.FindCollectionByNameOrId("dd2l9a4vxpy2ni8")
		if err != nil {
			return err
		}

		collection.ListRule = types.Pointer("(@request.auth.id != \"\" && @collection.trails.summit_logs.id ?= id && @collection.trails.author ?= @request.auth.id) || (@collection.trails.summit_logs.id ?= id && @collection.trails.public ?= true) || (author = @request.auth.id)")

		collection.ViewRule = types.Pointer("(@request.auth.id != \"\" && @collection.trails.summit_logs.id ?= id && @collection.trails.author ?= @request.auth.id) || (@collection.trails.summit_logs.id ?= id && @collection.trails.public ?= true) || (author = @request.auth.id)")

		return dao.SaveCollection(collection)
	})
}
