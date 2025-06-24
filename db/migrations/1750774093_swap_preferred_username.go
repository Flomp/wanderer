package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		actors, err := app.FindAllRecords("activitypub_actors")
		if err != nil {
			return err
		}

		for _, a := range actors {
			if !a.GetBool("isLocal") {
				continue
			}

			username := a.GetString("username")
			preferredUsername := a.GetString("preferred_username")

			a.Set("username", preferredUsername)
			a.Set("preferred_username", username)

			err = app.Save(a)
			if err != nil {
				return err
			}
		}

		return nil
	}, func(app core.App) error {
		actors, err := app.FindAllRecords("activitypub_actors")
		if err != nil {
			return err
		}

		for _, a := range actors {
			if !a.GetBool("isLocal") {
				continue
			}

			username := a.GetString("username")
			preferredUsername := a.GetString("preferred_username")

			a.Set("username", preferredUsername)
			a.Set("preferred_username", username)

			err = app.Save(a)
			if err != nil {
				return err
			}
		}

		return nil
	})
}
