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

		collection, err := dao.FindCollectionByNameOrId("1mns8mlal6uf9ku")
		if err != nil {
			return err
		}

		collection.ListRule = types.Pointer("trail.author = @request.auth.id || user = @request.auth.id")

		collection.ViewRule = types.Pointer("trail.author = @request.auth.id || user = @request.auth.id")

		return dao.SaveCollection(collection)
	}, func(db dbx.Builder) error {
		dao := daos.New(db);

		collection, err := dao.FindCollectionByNameOrId("1mns8mlal6uf9ku")
		if err != nil {
			return err
		}

		collection.ListRule = types.Pointer("trail.author = @request.auth.id")

		collection.ViewRule = types.Pointer("trail.author = @request.auth.id")

		return dao.SaveCollection(collection)
	})
}
