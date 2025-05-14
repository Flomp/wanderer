package migrations

import (
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
			actor, err := app.FindFirstRecordByData("activitypub_actors", "user", t.GetString("user"))
			if err != nil {
				return err
			}
			t.Set("author", actor.Id)
			err = app.UnsafeWithoutHooks().Save(t)
			if err != nil {
				return err
			}
		}

		return nil
	}, func(app core.App) error {
		// add down queries...

		return nil
	})
}
