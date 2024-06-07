package migrations

import (
	"os"

	"github.com/meilisearch/meilisearch-go"
	"github.com/pocketbase/dbx"

	"github.com/pocketbase/pocketbase/daos"
	m "github.com/pocketbase/pocketbase/migrations"

	"pocketbase/util"
)

func init() {
	client := meilisearch.NewClient(meilisearch.ClientConfig{
		Host:   os.Getenv("MEILI_URL"),
		APIKey: os.Getenv("MEILI_MASTER_KEY"),
	})

	m.Register(func(db dbx.Builder) error {
		_, err := client.Index("trails").UpdateFilterableAttributes(&[]string{
			"_geo", "author", "category", "completed", "date", "difficulty", "distance", "elevation_gain", "public", "shares",
		})

		if err != nil {
			return err
		}

		dao := daos.New(db)

		var usernames []string
		err = db.NewQuery("SELECT username FROM users").Column(&usernames)

		if err != nil {
			return err
		}

		for _, username := range usernames {

			record, err := dao.FindAuthRecordByUsername("users", username)
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

				if err := dao.SaveRecord(record); err != nil {
					return err
				}
			}

		}

		return nil
	}, func(db dbx.Builder) error {
		// add down queries...

		return nil
	})
}
