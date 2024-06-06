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

		collection, err := dao.FindCollectionByNameOrId("e864strfxo14pm4")
		if err != nil {
			return err
		}

		collection.ListRule = types.Pointer("author = @request.auth.id || public = true || (trail_share_via_trail.trail = id && trail_share_via_trail.user = @request.auth.id)")

		collection.ViewRule = types.Pointer("author = @request.auth.id || public = true || (trail_share_via_trail.trail = id && trail_share_via_trail.user = @request.auth.id)")

		return dao.SaveCollection(collection)
	}, func(db dbx.Builder) error {
		dao := daos.New(db);

		collection, err := dao.FindCollectionByNameOrId("e864strfxo14pm4")
		if err != nil {
			return err
		}

		collection.ListRule = types.Pointer("author = @request.auth.id || public = true")

		collection.ViewRule = types.Pointer("author = @request.auth.id || public = true")

		return dao.SaveCollection(collection)
	})
}
