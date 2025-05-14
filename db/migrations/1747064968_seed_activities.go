package migrations

import (
	"os"
	"pocketbase/federation"

	pub "github.com/go-ap/activitypub"
	"github.com/meilisearch/meilisearch-go"
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	client := meilisearch.New(os.Getenv("MEILI_URL"), meilisearch.WithAPIKey(os.Getenv("MEILI_MASTER_KEY")))
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

		logs, err := app.FindAllRecords("summit_logs")
		if err != nil {
			return err
		}

		for _, t := range logs {
			err = federation.CreateSummitLogActivity(app, client, t, pub.CreateType)
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
