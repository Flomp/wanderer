package migrations

import (
	"os"
	"pocketbase/util"

	"github.com/meilisearch/meilisearch-go"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	client := meilisearch.New(os.Getenv("MEILI_URL"), meilisearch.WithAPIKey(os.Getenv("MEILI_MASTER_KEY")))

	m.Register(func(app core.App) error {

		_, err := client.CreateIndex(&meilisearch.IndexConfig{
			Uid:        "lists",
			PrimaryKey: "id",
		})
		if err != nil {
			return err
		}

		_, err = client.Index("lists").UpdateSortableAttributes(&[]string{
			"created", "name",
		})
		if err != nil {
			return err
		}

		_, err = client.Index("lists").UpdateFilterableAttributes(&[]string{
			"author", "public", "shares",
		})
		if err != nil {
			return err
		}

		lists, err := app.FindAllRecords("lists", dbx.NewExp("true"))
		if err != nil {
			return err
		}

		for _, l := range lists {
			err = util.IndexList(l, client)
			if err != nil {
				return err
			}
			shares, err := app.FindAllRecords("list_share",
				dbx.NewExp("list = {:listId}", dbx.Params{"listId": l.Id}),
			)
			if err != nil {
				return err
			}
			userIds := make([]string, len(shares))
			for i, r := range shares {
				userIds[i] = r.GetString("user")
			}
			err = util.UpdateListShares(l.Id, userIds, client)
			if err != nil {
				return err
			}
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
				"lists": map[string]string{
					"filter": "public = true OR author = " + record.Id + " OR shares = " + record.Id,
				},
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

		_, err = client.DeleteIndex("cities500")
		if err != nil {
			return err
		}

		return nil
	}, func(app core.App) error {
		_, err := client.DeleteIndex("lists")
		if err != nil {
			return err
		}
		return nil
	})
}
