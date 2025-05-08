package federation

import (
	"fmt"
	"net/url"
	"os"
	"path"
	"time"

	pub "github.com/go-ap/activitypub"
	"github.com/meilisearch/meilisearch-go"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/security"
)

func CreateDeleteActivity(app core.App, r *core.Record) error {
	origin := os.Getenv("ORIGIN")
	if origin == "" {
		return fmt.Errorf("ORIGIN not set")
	}

	author, err := app.FindRecordById("activitypub_actors", r.GetString("author"))
	if err != nil {
		return err
	}

	collection, err := app.FindCollectionByNameOrId("activitypub_activities")
	if err != nil {
		return err
	}

	recordId := security.RandomStringWithAlphabet(core.DefaultIdLength, core.DefaultIdAlphabet)

	id := fmt.Sprintf("%s/api/v1/activitypub/activity/%s", origin, recordId)
	to := "https://www.w3.org/ns/activitystreams#Public"
	cc := author.GetString("iri") + "/followers"
	object := fmt.Sprintf("%s/api/v1/trail/%s", origin, r.Id)

	record := core.NewRecord(collection)
	record.Set("id", recordId)
	record.Set("iri", id)
	record.Set("type", string(pub.DeleteType))
	record.Set("to", to)
	record.Set("cc", cc)
	record.Set("object", object)
	record.Set("actor", author.GetString("iri"))
	record.Set("published", time.Now())

	err = app.Save(record)
	if err != nil {
		return err
	}

	activity := pub.DeleteNew(pub.IRI(id), pub.IRI(object))
	activity.Actor = pub.IRI(author.GetString("iri"))
	activity.To = pub.ItemCollection{pub.IRI(to)}
	activity.CC = pub.ItemCollection{pub.IRI(cc)}
	activity.Published = time.Now()

	follows, err := app.FindRecordsByFilter("follows", "followee={:followee}", "", -1, 0, dbx.Params{"followee": author.Id})
	if err != nil {
		return err
	}

	recipients := []string{}
	for _, f := range follows {
		follower, err := app.FindRecordById("activitypub_actors", f.GetString("follower"))
		if err != nil {
			return err
		}
		recipients = append(recipients, follower.GetString("iri")+"/inbox")
	}

	return PostActivity(app, author, activity, recipients)
}

func ProcessDeleteActivity(app core.App, actor *core.Record, activity pub.Activity) error {
	// no need to do anything if the actor is local
	// if actor.GetBool("isLocal") {
	// 	return nil
	// }

	client := meilisearch.New(
		os.Getenv("MEILI_URL"),
		meilisearch.WithAPIKey(os.Getenv("MEILI_MASTER_KEY")),
	)
	trailUrl, err := url.Parse(activity.Object.GetID().String())
	if err != nil {
		return err
	}
	recordId := path.Base(trailUrl.Path)

	_, err = client.Index("trails").Delete(recordId)
	if err != nil {
		return err
	}

	return nil
}
