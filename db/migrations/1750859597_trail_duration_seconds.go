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
			duration := t.GetFloat("duration")
			t.Set("duration", duration*60)
			err = app.Save(t)
			if err != nil {
				return err
			}
		}

		return nil
	}, func(app core.App) error {
		trails, err := app.FindAllRecords("trails")
		if err != nil {
			return err
		}

		for _, t := range trails {
			duration := t.GetFloat("duration")
			t.Set("duration", duration/60)
			err = app.Save(t)
			if err != nil {
				return err
			}
		}

		return nil
	})
}
