package migrations

import (
	"encoding/json"

	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		collection, err := app.FindCollectionByNameOrId("pbc_468398817")
		if err != nil {
			return err
		}

		// update collection data
		if err := json.Unmarshal([]byte(`{
			"viewQuery": "SELECT\n    id,\n    trail_id,\n    date,\n    name,\n    description,\n    gpx,\n    author,\n    photos,\n    distance,\n    duration,\n    elevation_gain,\n    elevation_loss,\n    created,\n    type\nFROM\n    (\n        SELECT\n            summit_logs.id,\n            summit_logs.trail as trail_id,\n            summit_logs.date,\n            trails.name,\n            text as description,\n            summit_logs.gpx,\n            activitypub_actors.iri as author,\n            summit_logs.photos,\n            summit_logs.distance,\n            summit_logs.duration,\n            summit_logs.elevation_gain,\n            summit_logs.elevation_loss,\n            summit_logs.created,\n            \"summit_log\" as type\n        FROM\n            summit_logs\n            JOIN trails ON summit_logs.trail = trails.id\n            JOIN activitypub_actors ON activitypub_actors.id = summit_logs.author\n        UNION\n        SELECT\n            trails.id,\n            trails.id as trail_id,\n            date,\n            trails.name,\n            description,\n            gpx,\n            activitypub_actors.iri as author,\n            photos,\n            distance,\n            duration,\n            elevation_gain,\n            elevation_loss,\n            trails.created,\n            \"trail\" as type\n        FROM\n            trails\n            JOIN activitypub_actors ON activitypub_actors.id = trails.author\n    )\nORDER BY\n    created DESC;\n"
		}`), &collection); err != nil {
			return err
		}

		return app.Save(collection)
	}, func(app core.App) error {
		collection, err := app.FindCollectionByNameOrId("pbc_468398817")
		if err != nil {
			return err
		}

		// update collection data
		if err := json.Unmarshal([]byte(`{
			"viewQuery": "SELECT\n    (ROW_NUMBER() OVER ()) as id,\n    trail_id,\n    date,\n    name,\n    description,\n    gpx,\n    author,\n    photos,\n    distance,\n    duration,\n    elevation_gain,\n    elevation_loss,\n    created,\n    type\nFROM\n    (\n        SELECT\n            summit_logs.trail as trail_id,\n            summit_logs.date,\n            trails.name,\n            text as description,\n            summit_logs.gpx,\n            activitypub_actors.iri as author,\n            summit_logs.photos,\n            summit_logs.distance,\n            summit_logs.duration,\n            summit_logs.elevation_gain,\n            summit_logs.elevation_loss,\n            summit_logs.created,\n            \"summit_log\" as type\n        FROM\n            summit_logs\n            JOIN trails ON summit_logs.trail = trails.id\n            JOIN activitypub_actors ON activitypub_actors.id = summit_logs.author\n        UNION\n        SELECT\n            trails.id as trail_id,\n            date,\n            trails.name,\n            description,\n            gpx,\n            activitypub_actors.iri as author,\n            photos,\n            distance,\n            duration,\n            elevation_gain,\n            elevation_loss,\n            trails.created,\n            \"trail\" as type\n        FROM\n            trails\n            JOIN activitypub_actors ON activitypub_actors.id = trails.author\n    )\nORDER BY\n    created DESC;\n"
		}`), &collection); err != nil {
			return err
		}

		return app.Save(collection)
	})
}
