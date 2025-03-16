package migrations

import (
	"os"
	"pocketbase/util"

	"github.com/meilisearch/meilisearch-go"

	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	client := meilisearch.New(os.Getenv("MEILI_URL"), meilisearch.WithAPIKey(os.Getenv("MEILI_MASTER_KEY")))

	m.Register(func(app core.App) error {
		_, err := client.Index("trails").UpdateFilterableAttributes(&[]string{
			"_geo", "author", "category", "completed", "date", "difficulty", "distance", "elevation_gain", "public", "shares",
		})

		if err != nil {
			return err
		}

		var usernames []string
		err = app.DB().NewQuery("SELECT username FROM users").Column(&usernames)

		if err != nil {
			return err
		}

		for _, username := range usernames {
			record, err := app.FindFirstRecordByData("users", "username", username)
			if err != nil {
				return err
			}

			searchRules := map[string]interface{}{
				"cities500": map[string]string{},
				"trails": map[string]string{
					"filter": "public = true OR author = " + record.Id + " OR shares = " + record.Id,
				},
			}

			if token, err := util.GenerateMeilisearchToken(searchRules, client); err != nil {
				return err
			} else {
				record.Set("token", token)

				if err := app.Save(record); err != nil {
					return err
				}
			}

		}

		return nil
	}, func(app core.App) error {
		// add down queries...

		return nil
	})
}
