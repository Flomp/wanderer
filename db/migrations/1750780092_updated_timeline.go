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
			"viewQuery": "SELECT\n    id,\n    trail_id,\n    trail_author_username,\n    trail_author_domain,\n    trail_iri,\n    date,\n    name,\n    description,\n    gpx,\n    author,\n    photos,\n    distance,\n    duration,\n    elevation_gain,\n    elevation_loss,\n    created,\n    type\nFROM\n    (\n        SELECT\n            summit_logs.id,\n            summit_logs.trail as trail_id,\n            tapa.preferred_username as trail_author_username,\n            tapa.domain as trail_author_domain,\n            trails.iri as trail_iri,\n            summit_logs.date,\n            trails.name,\n            text as description,\n            summit_logs.gpx,\n            sapa.iri as author,\n            summit_logs.photos,\n            summit_logs.distance,\n            summit_logs.duration,\n            summit_logs.elevation_gain,\n            summit_logs.elevation_loss,\n            summit_logs.created,\n            \"summit_log\" as type\n        FROM\n            summit_logs\n            JOIN trails ON summit_logs.trail = trails.id\n            JOIN activitypub_actors sapa ON sapa.id = summit_logs.author\n              JOIN activitypub_actors tapa ON tapa.id = trails.author\n        UNION\n        SELECT\n            trails.id,\n            trails.id as trail_id,\n            activitypub_actors.preferred_username as trail_author_username,\n            activitypub_actors.domain as trail_author_domain,\n            trails.iri as trail_iri,\n            date,\n            trails.name,\n            description,\n            gpx,\n            activitypub_actors.iri as author,\n            photos,\n            distance,\n            duration,\n            elevation_gain,\n            elevation_loss,\n            trails.created,\n            \"trail\" as type\n        FROM\n            trails\n            JOIN activitypub_actors ON activitypub_actors.id = trails.author\n    )\nORDER BY\n    created DESC;\n"
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
			"viewQuery": "SELECT\n    id,\n    trail_id,\n    trail_author_username,\n    trail_author_domain,\n    trail_iri,\n    date,\n    name,\n    description,\n    gpx,\n    author,\n    photos,\n    distance,\n    duration,\n    elevation_gain,\n    elevation_loss,\n    created,\n    type\nFROM\n    (\n        SELECT\n            summit_logs.id,\n            summit_logs.trail as trail_id,\n            tapa.username as trail_author_username,\n            tapa.domain as trail_author_domain,\n            trails.iri as trail_iri,\n            summit_logs.date,\n            trails.name,\n            text as description,\n            summit_logs.gpx,\n            sapa.iri as author,\n            summit_logs.photos,\n            summit_logs.distance,\n            summit_logs.duration,\n            summit_logs.elevation_gain,\n            summit_logs.elevation_loss,\n            summit_logs.created,\n            \"summit_log\" as type\n        FROM\n            summit_logs\n            JOIN trails ON summit_logs.trail = trails.id\n            JOIN activitypub_actors sapa ON sapa.id = summit_logs.author\n              JOIN activitypub_actors tapa ON tapa.id = trails.author\n        UNION\n        SELECT\n            trails.id,\n            trails.id as trail_id,\n            activitypub_actors.username as trail_author_username,\n            activitypub_actors.domain as trail_author_domain,\n            trails.iri as trail_iri,\n            date,\n            trails.name,\n            description,\n            gpx,\n            activitypub_actors.iri as author,\n            photos,\n            distance,\n            duration,\n            elevation_gain,\n            elevation_loss,\n            trails.created,\n            \"trail\" as type\n        FROM\n            trails\n            JOIN activitypub_actors ON activitypub_actors.id = trails.author\n    )\nORDER BY\n    created DESC;\n"
		}`), &collection); err != nil {
			return err
		}

		return app.Save(collection)
	})
}
