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

		collection, err := dao.FindCollectionByNameOrId("khrcci2uqknny8h")
		if err != nil {
			return err
		}

		collection.UpdateRule = types.Pointer("@request.auth.id = recipient && @request.data.type = type && @request.data.metadata = metadata && @request.data.recipient = recipient && @request.data.author = author && @request.data.seen = true")

		return dao.SaveCollection(collection)
	}, func(db dbx.Builder) error {
		dao := daos.New(db);

		collection, err := dao.FindCollectionByNameOrId("khrcci2uqknny8h")
		if err != nil {
			return err
		}

		collection.UpdateRule = types.Pointer("@request.auth.id = recipient && @request.data.type = type && @request.data.metadata = metadata && @request.data.recipient = recipient && @request.data.author = author")

		return dao.SaveCollection(collection)
	})
}
