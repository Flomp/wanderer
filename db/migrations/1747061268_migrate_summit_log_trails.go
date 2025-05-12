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
			trail, err := app.FindFirstRecordByFilter("trails", "summit_logs ?~ {:id}", dbx.Params{"id": l.Id})
			if err != nil {
				continue
			}
			l.Set("trail", trail.Id)
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
