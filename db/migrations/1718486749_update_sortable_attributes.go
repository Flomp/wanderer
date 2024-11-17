package migrations

import (
	"os"

	"github.com/meilisearch/meilisearch-go"
	"github.com/pocketbase/dbx"

	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	client := meilisearch.New(os.Getenv("MEILI_URL"), meilisearch.WithAPIKey(os.Getenv("MEILI_MASTER_KEY")))

	m.Register(func(db dbx.Builder) error {
		_, err := client.Index("trails").UpdateSortableAttributes(&[]string{
			"created", "date", "difficulty", "distance", "elevation_gain", "name",
		})

		if err != nil {
			return err
		}

		return nil
	}, func(db dbx.Builder) error {
		_, err := client.Index("trails").UpdateSortableAttributes(&[]string{
			"created", "difficulty", "distance", "elevation_gain", "name",
		})

		if err != nil {
			return err
		}

		return nil
	})
}
