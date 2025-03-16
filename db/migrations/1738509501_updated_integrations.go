package migrations

import (
	"encoding/json"

	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {

		collection, err := app.FindCollectionByNameOrId("iz4sezoehde64wp")
		if err != nil {
			return err
		}

		if err := json.Unmarshal([]byte(`[
			"CREATE UNIQUE INDEX `+"`"+`idx_qMhw0Em`+"`"+` ON `+"`"+`integrations`+"`"+` (`+"`"+`user`+"`"+`)"
		]`), &collection.Indexes); err != nil {
			return err
		}

		return app.Save(collection)
	}, func(app core.App) error {

		collection, err := app.FindCollectionByNameOrId("iz4sezoehde64wp")
		if err != nil {
			return err
		}

		if err := json.Unmarshal([]byte(`[]`), &collection.Indexes); err != nil {
			return err
		}

		return app.Save(collection)
	})
}
