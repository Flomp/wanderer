package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
	"github.com/pocketbase/pocketbase/tools/types"
)

func init() {
	m.Register(func(app core.App) error {

		collection, err := app.FindCollectionByNameOrId("goeo2ubp103rzp9")
		if err != nil {
			return err
		}

		collection.ListRule = types.Pointer("@request.auth.id != \"\" && (@collection.trails.waypoints.id ?= id && (@collection.trails.author = @request.auth.id || @collection.trails.public = true) || author = @request.auth.id)")

		return app.Save(collection)
	}, func(app core.App) error {

		collection, err := app.FindCollectionByNameOrId("goeo2ubp103rzp9")
		if err != nil {
			return err
		}

		collection.ListRule = types.Pointer("@request.auth.id != \"\" && @collection.trails.waypoints.id ?= id && (@collection.trails.author = @request.auth.id || @collection.trails.public = true)")

		return app.Save(collection)
	})
}
