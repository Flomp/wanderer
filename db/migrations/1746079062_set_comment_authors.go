package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		comments, err := app.FindAllRecords("comments")
		if err != nil {
			return err
		}

		for _, c := range comments {
			actor, err := app.FindFirstRecordByData("activitypub_actors", "user", c.GetString("author"))
			if err != nil {
				continue
			}
			c.Set("author", actor.Id)
			err = app.Save(c)
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
