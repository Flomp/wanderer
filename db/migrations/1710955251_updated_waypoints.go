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

		collection, err := dao.FindCollectionByNameOrId("goeo2ubp103rzp9")
		if err != nil {
			return err
		}

		collection.UpdateRule = types.Pointer("@request.auth.id != \"\" && ((@collection.trails.waypoints.id ?= id && @collection.trails.author = @request.auth.id) || @collection.waypoints.author = @request.auth.id)")

		collection.DeleteRule = types.Pointer("@request.auth.id != \"\" && ((@collection.trails.waypoints.id ?= id && @collection.trails.author = @request.auth.id) || @collection.waypoints.author = @request.auth.id)")

		return dao.SaveCollection(collection)
	}, func(db dbx.Builder) error {
		dao := daos.New(db);

		collection, err := dao.FindCollectionByNameOrId("goeo2ubp103rzp9")
		if err != nil {
			return err
		}

		collection.UpdateRule = types.Pointer("@request.auth.id != \"\" && ((@collection.trails.waypoints.id ?= id && @collection.trails.author = @request.auth.id) || @request.data.author = @request.auth.id)")

		collection.DeleteRule = types.Pointer("@request.auth.id != \"\" && ((@collection.trails.waypoints.id ?= id && @collection.trails.author = @request.auth.id) || @request.data.author = @request.auth.id)")

		return dao.SaveCollection(collection)
	})
}
