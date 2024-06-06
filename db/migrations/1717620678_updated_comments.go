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

		collection, err := dao.FindCollectionByNameOrId("lf06qip3f4d11yk")
		if err != nil {
			return err
		}

		collection.ListRule = types.Pointer("((@request.auth.id != \"\" && trail.author = @request.auth.id) || trail.public = true) || author = @request.auth.id || trail.trail_share_via_trail.user ?= @request.auth.id")

		collection.ViewRule = types.Pointer("((@request.auth.id != \"\" && trail.author = @request.auth.id) || trail.public = true) || author = @request.auth.id || trail.trail_share_via_trail.user ?= @request.auth.id")

		collection.CreateRule = types.Pointer("@request.auth.id != \"\" && (trail.author = @request.auth.id || trail.public = true || trail.trail_share_via_trail.user ?= @request.auth.id)")

		return dao.SaveCollection(collection)
	}, func(db dbx.Builder) error {
		dao := daos.New(db);

		collection, err := dao.FindCollectionByNameOrId("lf06qip3f4d11yk")
		if err != nil {
			return err
		}

		collection.ListRule = types.Pointer("((@request.auth.id != \"\" && trail.author = @request.auth.id) || trail.public = true) || author = @request.auth.id")

		collection.ViewRule = types.Pointer("((@request.auth.id != \"\" && trail.author = @request.auth.id) || trail.public = true) || author = @request.auth.id")

		collection.CreateRule = types.Pointer("@request.auth.id != \"\" && (trail.author = @request.auth.id || trail.public = true)")

		return dao.SaveCollection(collection)
	})
}
