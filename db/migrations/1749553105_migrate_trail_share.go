package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		shares, err := app.FindAllRecords("trail_share")
		if err != nil {
			return err
		}

		for _, s := range shares {
			actor, err := app.FindFirstRecordByData("activitypub_actors", "user", s.GetString("user"))
			if err != nil {
				continue
			}
			s.Set("actor", actor.Id)
			err = app.Save(s)
			if err != nil {
				return err
			}
		}

		return nil
	}, func(app core.App) error {
		return nil
	})
}
