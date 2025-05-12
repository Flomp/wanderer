package migrations

import (
	"pocketbase/federation"

	pub "github.com/go-ap/activitypub"
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		trails, err := app.FindAllRecords("trails")

		if err != nil {
			return err
		}

		for _, t := range trails {
			err = federation.CreateTrailActivity(app, t, pub.CreateType)
			if err != nil {
				continue
			}
		}

		return nil
	}, func(app core.App) error {
		// add down queries...

		return nil
	})
}
