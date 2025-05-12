package federation

import (
	"fmt"
	"net/url"
	"os"
	"path"
	"strings"
	"time"

	pub "github.com/go-ap/activitypub"
	"github.com/meilisearch/meilisearch-go"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/security"
)

func CreateTrailDeleteActivity(app core.App, r *core.Record) error {
	// prevents running this hook during migrations
	if app.IsTransactional() {
		return nil
	}

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

func CreateCommentDeleteActivity(app core.App, client meilisearch.ServiceManager, r *core.Record) error {
	// prevents running this hook during migrations
	if app.IsTransactional() {
		return nil
	}

	origin := os.Getenv("ORIGIN")
	if origin == "" {
		return fmt.Errorf("ORIGIN not set")
	}

	author, err := app.FindRecordById("activitypub_actors", r.GetString("author"))
	if err != nil {
		return err
	}

	commentTrail := struct {
		Domain string `json:"domain"`
		Author string `json:"author"`
		URL    string `json:"url"`
	}{}
	err = client.Index("trails").GetDocument(r.GetString("trail"), &meilisearch.DocumentQuery{
		Fields: []string{"domain", "author", "url"},
	}, &commentTrail)
	if err != nil {
		return err
	}

	commentTrailAuthor, err := app.FindRecordById("activitypub_actors", r.GetString("author"))
	if err != nil {
		return err
	}

	collection, err := app.FindCollectionByNameOrId("activitypub_activities")
	if err != nil {
		return err
	}

	recordId := security.RandomStringWithAlphabet(core.DefaultIdLength, core.DefaultIdAlphabet)

	id := fmt.Sprintf("%s/api/v1/activitypub/activity/%s", origin, recordId)
	to := commentTrailAuthor.GetString("iri")
	object := fmt.Sprintf("%s#comment-%s", to, r.Id)

	record := core.NewRecord(collection)
	record.Set("id", recordId)
	record.Set("iri", id)
	record.Set("type", string(pub.DeleteType))
	record.Set("to", to)
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
	activity.Published = time.Now()

	return PostActivity(app, author, activity, []string{to + "/inbox"})
}

func CreateSummitLogDeleteActivity(app core.App, client meilisearch.ServiceManager, r *core.Record) error {
	// prevents running this hook during migrations
	if app.IsTransactional() {
		return nil
	}

	origin := os.Getenv("ORIGIN")
	if origin == "" {
		return fmt.Errorf("ORIGIN not set")
	}

	author, err := app.FindRecordById("activitypub_actors", r.GetString("author"))
	if err != nil {
		return err
	}

	summitLogTrail := struct {
		Domain string `json:"domain"`
		Author string `json:"author"`
		URL    string `json:"url"`
	}{}
	err = client.Index("trails").GetDocument(r.GetString("trail"), &meilisearch.DocumentQuery{
		Fields: []string{"domain", "author", "url"},
	}, &summitLogTrail)
	if err != nil {
		return err
	}

	commentTrailAuthor, err := app.FindRecordById("activitypub_actors", r.GetString("author"))
	if err != nil {
		return err
	}

	collection, err := app.FindCollectionByNameOrId("activitypub_activities")
	if err != nil {
		return err
	}

	recordId := security.RandomStringWithAlphabet(core.DefaultIdLength, core.DefaultIdAlphabet)

	id := fmt.Sprintf("%s/api/v1/activitypub/activity/%s", origin, recordId)
	to := commentTrailAuthor.GetString("iri")
	object := fmt.Sprintf("%s#summit-log-%s", to, r.Id)

	record := core.NewRecord(collection)
	record.Set("id", recordId)
	record.Set("iri", id)
	record.Set("type", string(pub.DeleteType))
	record.Set("to", to)
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
	activity.Published = time.Now()

	return PostActivity(app, author, activity, []string{to + "/inbox"})
}

func ProcessDeleteActivity(app core.App, actor *core.Record, activity pub.Activity) error {
	// no need to do anything if the actor is local
	if actor.GetBool("isLocal") {
		return nil
	}

	object := activity.Object.GetID().String()

	var err error
	switch {
	case strings.Contains(object, "trail"):
		err = processDeleteTrailActivity(activity)
	case strings.Contains(object, "comment"):
		err = processDeleteCommentActivity(app, actor, activity)
	case strings.Contains(object, "summit-log"):
		err = processDeleteSummitLogActivity(app, actor, activity)
	}

	if err != nil {
		return err
	}

	return nil
}

func processDeleteTrailActivity(activity pub.Activity) error {
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

func processDeleteCommentActivity(app core.App, actor *core.Record, activity pub.Activity) error {
	object := activity.Object.GetID().String()

	commentId := object[len(object)-15:]
	comment, err := app.FindRecordById("comments", commentId)
	if err != nil {
		return err
	}

	if comment.GetString("author") != actor.Id {
		return fmt.Errorf("actor is not comment author")
	}

	err = app.Delete(comment)
	if err != nil {
		return err
	}
	return nil
}

func processDeleteSummitLogActivity(app core.App, actor *core.Record, activity pub.Activity) error {
	object := activity.Object.GetID().String()

	summitLogId := object[len(object)-15:]
	summitLog, err := app.FindRecordById("summit_logs", summitLogId)
	if err != nil {
		return err
	}

	if summitLog.GetString("author") != actor.Id {
		return fmt.Errorf("actor is not summit log author")
	}

	err = app.Delete(summitLog)
	if err != nil {
		return err
	}
	return nil
}
