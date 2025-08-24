package migrations

import (
	"encoding/json"

	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		collection, err := app.FindCollectionByNameOrId("pbc_1973704172")
		if err != nil {
			return err
		}

		// update collection data
		if err := json.Unmarshal([]byte(`{
			"viewQuery": "SELECT\n    (ROW_NUMBER() OVER ()) as id,\n    actor,\n    item,\n    type,\n    created\nFROM\n    (\n        SELECT\n            author as actor,\n            id as item,\n            'list' as type,\n            created\n        FROM\n            lists\n        WHERE lists.public = TRUE\n        UNION\n        SELECT\n            author as actor,\n            id as item,\n            'trail' as type,\n            created\n        FROM trails\n        WHERE trails.public = TRUE\n    )\nORDER BY created desc;"
		}`), &collection); err != nil {
			return err
		}

		return app.Save(collection)
	}, func(app core.App) error {
		collection, err := app.FindCollectionByNameOrId("pbc_1973704172")
		if err != nil {
			return err
		}

		// update collection data
		if err := json.Unmarshal([]byte(`{
			"viewQuery": "SELECT\n    (ROW_NUMBER() OVER ()) as id,\n    actor,\n    item,\n    type,\n    created\nFROM\n    (\n        SELECT\n            author as actor,\n            id as item,\n            'list' as type,\n            created\n        FROM\n            lists\n        UNION\n        SELECT\n            author as actor,\n            id as item,\n            'trail' as type,\n            created\n        FROM trails\n    )\nORDER BY created desc;"
		}`), &collection); err != nil {
			return err
		}

		return app.Save(collection)
	})
}
