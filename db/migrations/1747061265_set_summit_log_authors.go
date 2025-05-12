package migrations

import (
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		logs, err := app.FindAllRecords("summit_logs")
		if err != nil {
			return err
		}

		for _, l := range logs {
			if l.GetString("user") == "" {
				trail, err := app.FindFirstRecordByFilter("trails", "summit_logs ?~ {:id}", dbx.Params{"id": l.Id})
				if err != nil {
					// orphaned
					err = app.Delete(l)
					if err != nil {
						return err
					}
				}
				l.Set("user", trail.GetString("user"))
			}
			actor, err := app.FindFirstRecordByData("activitypub_actors", "user", l.GetString("user"))
			if err != nil {
				return err
			}
			l.Set("author", actor.Id)
			err = app.Save(l)
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
