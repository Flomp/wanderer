package migrations

import (
	"encoding/json"

	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		collection, err := app.FindCollectionByNameOrId("goeo2ubp103rzp9")
		if err != nil {
			return err
		}

		// update collection data
		if err := json.Unmarshal([]byte(`{
			"listRule": "author = @request.auth.id || trails_via_waypoints.author ?= @request.auth.id || trails_via_waypoints.public ?= true || \n(@collection.trail_share.trail.id ?= trails_via_waypoints.id && @collection.trail_share.user ?= @request.auth.id)",
			"viewRule": "author = @request.auth.id || trails_via_waypoints.author ?= @request.auth.id || trails_via_waypoints.public ?= true || \n(@collection.trail_share.trail.id ?= trails_via_waypoints.id && @collection.trail_share.user ?= @request.auth.id)"
		}`), &collection); err != nil {
			return err
		}

		return app.Save(collection)
	}, func(app core.App) error {
		collection, err := app.FindCollectionByNameOrId("goeo2ubp103rzp9")
		if err != nil {
			return err
		}

		// update collection data
		if err := json.Unmarshal([]byte(`{
			"viewRule": "(@request.auth.id != \"\" && @collection.trails.waypoints.id ?= id && @collection.trails.author ?= @request.auth.id) || (@collection.trails.waypoints.id ?= id && @collection.trails.public ?= true) || \n(@collection.trail_share.trail.waypoints.id ?= id && @collection.trail_share.user ?= @request.auth.id) ||\n(author = @request.auth.id)"
		}`), &collection); err != nil {
			return err
		}

		return app.Save(collection)
	})
}
