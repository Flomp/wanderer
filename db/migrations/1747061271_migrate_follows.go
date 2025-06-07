package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		follows, err := app.FindAllRecords("follows")
		if err != nil {
			return err
		}

		seen := map[string]string{}
		for _, f := range follows {
			oldFollower := f.GetString("old_follower")
			oldFollowee := f.GetString("old_followee")

			if oldFollowee == "" || oldFollower == "" {
				continue
			}

			seenFollowee, ok := seen[oldFollower]
			if ok && seenFollowee == oldFollowee {
				err = app.Delete(f)
				if err != nil {
					return err
				}
				continue
			}
			seen[oldFollower] = oldFollowee

			followerActor, err := app.FindFirstRecordByData("activitypub_actors", "user", oldFollower)
			if err != nil {
				return err
			}

			followeeActor, err := app.FindFirstRecordByData("activitypub_actors", "user", oldFollowee)
			if err != nil {
				return err
			}

			f.Set("follower", followerActor.Id)
			f.Set("followee", followeeActor.Id)
			f.Set("status", "accepted")

			err = app.UnsafeWithoutHooks().Save(f)
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
