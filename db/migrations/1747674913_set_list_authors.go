package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		lists, err := app.FindAllRecords("lists")
		if err != nil {
			return err
		}

		for _, l := range lists {
			actor, err := app.FindFirstRecordByData("activitypub_actors", "user", l.GetString("user"))
			if err != nil {
				continue
			}
			l.Set("author", actor.Id)
			err = app.UnsafeWithoutHooks().Save(l)
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
