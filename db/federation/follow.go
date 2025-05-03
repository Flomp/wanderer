package federation

import (
	"fmt"
	"os"
	"time"

	pub "github.com/go-ap/activitypub"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/security"
)

func FollowCreate(app core.App, follow *core.Record) error {
	origin := os.Getenv("ORIGIN")
	if origin == "" {
		return fmt.Errorf("ORIGIN not set")
	}

	follower := follow.GetString("follower")
	followee := follow.GetString("followee")
	status := follow.GetString("status")

	if status != "pending" {
		return fmt.Errorf("follow status != 'pending'")
	}

	followerActor, err := app.FindRecordById("activitypub_actors", follower)
	if err != nil {
		return err
	}

	followeeActor, err := app.FindRecordById("activitypub_actors", followee)
	if err != nil {
		return err
	}

	collection, err := app.FindCollectionByNameOrId("activitypub_activities")
	if err != nil {
		return err
	}

	recordId := security.RandomStringWithAlphabet(core.DefaultIdLength, core.DefaultIdAlphabet)

	id := fmt.Sprintf("%s/api/v1/activitypub/activity/%s", origin, recordId)

	record := core.NewRecord(collection)
	record.Set("id", recordId)
	record.Set("iri", id)
	record.Set("type", "Follow")
	record.Set("object", followeeActor.GetString("iri"))
	record.Set("actor", followerActor.GetString("iri"))
	record.Set("published", time.Now())

	err = app.Save(record)
	if err != nil {
		return err
	}

	activity := pub.ActivityNew(pub.IRI(id), pub.FollowType, pub.IRI(followeeActor.GetString("iri")))
	activity.Actor = pub.IRI(followerActor.GetString("iri"))

	return PostActivity(app, followerActor, activity, []string{followeeActor.GetString("inbox")})
}
