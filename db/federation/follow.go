package federation

import (
	"fmt"
	"os"
	"time"

	pub "github.com/go-ap/activitypub"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/security"
)

// create outgoing follow activity
func CreateFollowActivity(app core.App, follow *core.Record) error {
	origin := os.Getenv("ORIGIN")
	if origin == "" {
		return fmt.Errorf("ORIGIN not set")
	}

	follower := follow.GetString("follower")
	followee := follow.GetString("followee")

	followerActor, err := app.FindRecordById("activitypub_actors", follower)
	if err != nil {
		return err
	}

	followeeActor, err := app.FindRecordById("activitypub_actors", followee)
	if err != nil {
		return err
	}

	// this was a local follow
	// we can stop here since we auto accept all incoming follow requests
	if followeeActor.GetBool("isLocal") {
		return nil
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
	record.Set("type", string(pub.FollowType))
	record.Set("object", followeeActor.GetString("iri"))
	record.Set("actor", followerActor.GetString("iri"))
	record.Set("published", time.Now())

	err = app.Save(record)
	if err != nil {
		return err
	}

	activity := pub.FollowNew(pub.IRI(id), pub.IRI(followeeActor.GetString("iri")))
	activity.Actor = pub.IRI(followerActor.GetString("iri"))

	return PostActivity(app, followerActor, activity, []string{followeeActor.GetString("inbox")})
}

// process incoming follow activity
func ProcessFollowActivity(app core.App, actor *core.Record, activity pub.Activity) error {
	origin := os.Getenv("ORIGIN")
	if origin == "" {
		return fmt.Errorf("ORIGIN not set")
	}

	// find the followee in our db
	object, err := app.FindFirstRecordByData("activitypub_actors", "iri", activity.Object)
	if err != nil {
		return err
	}

	// a remote actor has requested the follow
	// this means we have not yet created a follow entry in our db
	// we accept it immediately
	if !actor.GetBool("isLocal") {
		followCollection, err := app.FindCollectionByNameOrId("follows")
		if err != nil {
			return err
		}
		followRecord := core.NewRecord(followCollection)
		followRecord.Set("follower", actor.Id)
		followRecord.Set("followee", object.Id)
		followRecord.Set("status", "accepted")
		err = app.Save(followRecord)
		if err != nil {
			return err
		}
	}

	// create record of the accept activity in our db
	collection, err := app.FindCollectionByNameOrId("activitypub_activities")
	if err != nil {
		return err
	}
	recordId := security.RandomStringWithAlphabet(core.DefaultIdLength, core.DefaultIdAlphabet)
	id := fmt.Sprintf("%s/api/v1/activitypub/activity/%s", origin, recordId)

	record := core.NewRecord(collection)
	record.Set("id", recordId)
	record.Set("iri", id)
	record.Set("type", string(pub.AcceptType))
	record.Set("object", activity)
	record.Set("actor", object.GetString("iri"))
	record.Set("published", time.Now())

	err = app.Save(record)
	if err != nil {
		return err
	}

	// send the accept activity back to the actor's inbox
	acceptActivity := pub.AcceptNew(pub.IRI(id), activity.Actor.GetLink())
	acceptActivity.Actor = activity.Object
	acceptActivity.Object = activity

	return PostActivity(app, object, acceptActivity, []string{actor.GetString("inbox")})
}

func ProcessAcceptActivity(app core.App, actor *core.Record, activity pub.Activity) error {

	followActivity := activity.Object.(*pub.Activity)

	follower, err := app.FindFirstRecordByData("activitypub_actors", "iri", followActivity.Actor)
	if err != nil {
		return err
	}

	follow, err := app.FindFirstRecordByFilter("follows", "follower={:follower} && followee={:followee}", dbx.Params{"follower": follower.Id, "followee": actor.Id})
	if err != nil {
		return err
	}
	follow.Set("status", "accepted")
	err = app.Save(follow)
	if err != nil {
		return err
	}
	return nil
}

// create outgoing follow activity
func CreateUnfollowActivity(app core.App, follow *core.Record) error {
	origin := os.Getenv("ORIGIN")
	if origin == "" {
		return fmt.Errorf("ORIGIN not set")
	}

	follower := follow.GetString("follower")
	followee := follow.GetString("followee")

	followerActor, err := app.FindRecordById("activitypub_actors", follower)
	if err != nil {
		return err
	}

	followeeActor, err := app.FindRecordById("activitypub_actors", followee)
	if err != nil {
		return err
	}

	// this was a local follow
	// we don't need to notify anyone
	if followeeActor.GetBool("isLocal") {
		return nil
	}

	// find the original follow activity
	followActivityRecord, err := app.FindFirstRecordByFilter("activitypub_activities", "actor={:actor}&&object={:object}&&type={:type}", dbx.Params{"actor": followerActor.GetString("iri"), "object": followeeActor.GetString("iri"), "type": string(pub.FollowType)})
	if err != nil {
		return err
	}
	followActivity := pub.FollowNew(pub.IRI(followActivityRecord.GetString("iri")), pub.IRI(followeeActor.GetString("iri")))
	followActivity.Actor = pub.IRI(followActivityRecord.GetString("actor"))

	collection, err := app.FindCollectionByNameOrId("activitypub_activities")
	if err != nil {
		return err
	}

	recordId := security.RandomStringWithAlphabet(core.DefaultIdLength, core.DefaultIdAlphabet)

	id := fmt.Sprintf("%s/api/v1/activitypub/activity/%s", origin, recordId)

	record := core.NewRecord(collection)
	record.Set("id", recordId)
	record.Set("iri", id)
	record.Set("type", string(pub.UndoType))
	record.Set("object", followActivity)
	record.Set("actor", followerActor.GetString("iri"))
	record.Set("published", time.Now())

	err = app.Save(record)
	if err != nil {
		return err
	}

	activity := pub.UndoNew(pub.IRI(id), followActivity)
	activity.Actor = pub.IRI(followerActor.GetString("iri"))

	return PostActivity(app, followerActor, activity, []string{followeeActor.GetString("inbox")})
}

func ProcessUnfollowActivity(app core.App, actor *core.Record, activity pub.Activity) error {

	followActivity := activity.Object.(*pub.Activity)

	followee, err := app.FindFirstRecordByData("activitypub_actors", "iri", followActivity.Object)
	if err != nil {
		return err
	}

	follow, err := app.FindFirstRecordByFilter("follows", "follower={:follower} && followee={:followee}", dbx.Params{"follower": actor.Id, "followee": followee.Id})
	if err != nil {
		return err
	}

	err = app.Delete(follow)
	if err != nil {
		return err
	}
	return nil
}
