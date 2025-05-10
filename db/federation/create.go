package federation

import (
	"fmt"
	"net/url"
	"os"
	"path"
	"time"

	"pocketbase/models"
	"pocketbase/util"

	pub "github.com/go-ap/activitypub"
	"github.com/meilisearch/meilisearch-go"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/security"
)

func CreateTrailActivity(app core.App, trail *core.Record, typ pub.ActivityVocabularyType) error {
	origin := os.Getenv("ORIGIN")
	if origin == "" {
		return fmt.Errorf("ORIGIN not set")
	}

	trailAuthor, err := app.FindRecordById("activitypub_actors", trail.GetString("author"))
	if err != nil {
		return err
	}
	errs := app.ExpandRecord(trail, []string{"tags"}, nil)
	if len(errs) > 0 {
		return fmt.Errorf("failed to expand tags: %v", errs)
	}
	errs = app.ExpandRecord(trail, []string{"category"}, nil)
	if len(errs) > 0 {
		return fmt.Errorf("failed to expand category: %v", errs)
	}

	tagRecords := trail.ExpandedAll("tags")
	tags := pub.ItemCollection{}

	for _, v := range tagRecords {
		tags.Append(pub.IRI(v.GetString("name")))
	}

	category := ""
	categoryRecord := trail.ExpandedOne("category")
	if categoryRecord != nil {
		category = categoryRecord.GetString("name")
	}

	collection, err := app.FindCollectionByNameOrId("activitypub_activities")
	if err != nil {
		return err
	}

	recordId := security.RandomStringWithAlphabet(core.DefaultIdLength, core.DefaultIdAlphabet)

	id := fmt.Sprintf("%s/api/v1/activitypub/activity/%s", origin, recordId)
	to := "https://www.w3.org/ns/activitystreams#Public"
	cc := trailAuthor.GetString("iri") + "/followers"

	photos := trail.GetStringSlice("photos")
	thumbnail := ""
	if len(photos) > 0 {
		thumbnail = fmt.Sprintf("%s/api/v1/files/trails/%s/%s", origin, trail.Id, photos[trail.GetInt("thumbnail")])
	}
	gpx := ""
	if trail.GetString("gpx") != "" {
		gpx = fmt.Sprintf("%s/api/v1/files/trails/%s/%s", origin, trail.Id, trail.GetString("gpx"))
	}

	attachments := pub.ItemCollection{}
	if thumbnail != "" {
		attachments.Append(pub.Document{
			Type:      pub.DocumentType,
			MediaType: "image/jpeg",
			URL:       pub.IRI(thumbnail),
		})
	}
	if gpx != "" {
		attachments.Append(pub.Document{
			Type:      pub.DocumentType,
			MediaType: "application/xml+gpx",
			URL:       pub.IRI(gpx),
		})
	}

	trailObject := models.TrailNew()
	trailObject.Name = pub.NaturalLanguageValuesNew(pub.LangRefValueNew(pub.NilLangRef, trail.GetString("name")))
	trailObject.Content = pub.NaturalLanguageValuesNew(pub.LangRefValueNew(pub.NilLangRef, trail.GetString("description")))
	trailObject.Location = pub.Place{
		Type:      pub.PlaceType,
		Name:      pub.NaturalLanguageValuesNew(pub.LangRefValueNew(pub.NilLangRef, trail.GetString("location"))),
		Latitude:  trail.GetFloat("lat"),
		Longitude: trail.GetFloat("lon"),
	}
	trailObject.AttributedTo = pub.IRI(trailAuthor.GetString("iri"))
	trailObject.Published = trail.GetDateTime("created").Time()
	trailObject.ID = pub.IRI(fmt.Sprintf("%s/api/v1/trail/%s", origin, trail.Id))
	trailObject.URL = pub.IRI(fmt.Sprintf("%s/trail/view/@%s/%s", origin, trailAuthor.GetString("username"), trail.Id))
	trailObject.TrailId = trail.Id

	trailObject.Distance = trail.GetFloat("distance")
	trailObject.ElevationGain = trail.GetFloat("elevation_gain")
	trailObject.ElevationLoss = trail.GetFloat("elevation_loss")
	trailObject.Duration = trail.GetFloat("duration")
	trailObject.Difficulty = trail.GetString("difficulty")
	trailObject.Category = category
	trailObject.Date = trail.GetDateTime("date").Time()
	trailObject.Attachment = attachments
	trailObject.Gpx = gpx
	trailObject.Thumbnail = thumbnail
	trailObject.Tag = tags

	activity := pub.ActivityNew(pub.IRI(id), typ, trailObject)
	activity.Actor = pub.IRI(trailAuthor.GetString("iri"))
	activity.To = pub.ItemCollection{pub.IRI(to)}
	activity.CC = pub.ItemCollection{pub.IRI(cc)}
	activity.Published = time.Now()

	record := core.NewRecord(collection)
	record.Set("id", recordId)
	record.Set("iri", id)
	record.Set("to", to)
	record.Set("cc", cc)
	record.Set("type", string(typ))
	record.Set("object", trailObject)
	record.Set("actor", trailAuthor.GetString("iri"))
	record.Set("published", time.Now())

	err = app.Save(record)
	if err != nil {
		return err
	}

	follows, err := app.FindRecordsByFilter("follows", "followee={:followee}", "", -1, 0, dbx.Params{"followee": trailAuthor.Id})
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

	return PostActivity(app, trailAuthor, activity, recipients)
}

func ProcessCreateActivity(app core.App, actor *core.Record, activity pub.Activity) error {

	// no need to do anything if the actor is local
	// if actor.GetBool("isLocal") {
	// 	return nil
	// }

	var err error
	switch activity.Object.GetType() {
	case models.TrailType:
		err = processCreateTrailActivity(activity, app, actor)
	case pub.NoteType:
		err = processCreateCommentActivity(activity, app, actor)
	}

	if err != nil {
		return err
	}

	return nil

}

func processCreateTrailActivity(activity pub.Activity, app core.App, actor *core.Record) error {
	client := meilisearch.New(
		os.Getenv("MEILI_URL"),
		meilisearch.WithAPIKey(os.Getenv("MEILI_MASTER_KEY")),
	)

	trailObject, err := models.ToTrail(activity.Object)
	if err != nil {
		return err
	}

	doc, err := util.DocumentFromActivity(app, trailObject, actor)
	if err != nil {
		return err
	}
	documents := []map[string]interface{}{doc}

	if _, err := client.Index("trails").AddDocuments(documents); err != nil {
		return err
	}
	return nil
}

func processCreateCommentActivity(activity pub.Activity, app core.App, actor *core.Record) error {

	commentObject, err := pub.ToObject(activity.Object)
	if err != nil {
		return err
	}

	if commentObject.InReplyTo == nil {
		return fmt.Errorf("error processing comment: InReplyTo empty")
	}

	trailUrl, err := url.Parse(commentObject.InReplyTo.GetLink().String())
	if err != nil {
		return err
	}
	trailId := path.Base(trailUrl.Path)

	recordId := commentObject.ID.String()[len(commentObject.ID.String())-15:]

	var record *core.Record
	if activity.Type == pub.CreateType {
		collection, err := app.FindCollectionByNameOrId("comments")
		if err != nil {
			return err
		}

		record := core.NewRecord(collection)
		record.Set("id", recordId)
	} else if activity.Type == pub.UpdateType {
		record, err = app.FindRecordById("comments", recordId)
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("activity must be of type 'Create' or 'Update")
	}

	record.Set("text", commentObject.Content.First().Value)
	record.Set("author", actor.Id)
	record.Set("trail", trailId)

	err = app.Save(record)
	if err != nil {
		return err
	}

	return nil
}

func CreateCommentActivity(app core.App, client meilisearch.ServiceManager, comment *core.Record, typ pub.ActivityVocabularyType) error {
	origin := os.Getenv("ORIGIN")
	if origin == "" {
		return fmt.Errorf("ORIGIN not set")
	}

	// author of the comment
	commentAuthor, err := app.FindRecordById("activitypub_actors", comment.GetString("author"))
	if err != nil {
		return err
	}

	// trail that the comment was left on
	commentTrail := struct {
		Domain string `json:"domain"`
		Author string `json:"author"`
		URL    string `json:"url"`
	}{}
	err = client.Index("trails").GetDocument(comment.GetString("trail"), &meilisearch.DocumentQuery{
		Fields: []string{"domain", "author", "url"},
	}, &commentTrail)
	if err != nil {
		return err
	}

	// author of the trail that the comment was left on
	commentTrailAuthor, err := app.FindRecordById("activitypub_actors", commentTrail.Author)
	if err != nil {
		return err
	}

	commentRecordId := comment.Id
	if commentRecordId == "" {
		commentRecordId = security.RandomStringWithAlphabet(core.DefaultIdLength, core.DefaultIdAlphabet)
	}
	activityRecordId := security.RandomStringWithAlphabet(core.DefaultIdLength, core.DefaultIdAlphabet)

	id := fmt.Sprintf("%s/api/v1/activitypub/activity/%s", origin, activityRecordId)
	to := commentTrailAuthor.GetString("iri")

	author := commentAuthor.GetString("iri")

	commentObject := pub.ObjectNew(pub.NoteType)
	commentObject.ID = pub.IRI(fmt.Sprintf("%s#comment-%s", author, commentRecordId))
	commentObject.Content = pub.NaturalLanguageValuesNew(pub.LangRefValueNew(pub.NilLangRef, comment.GetString("text")))
	commentObject.Published = comment.GetDateTime("created").Time()
	commentObject.AttributedTo = pub.IRI(author)
	commentObject.InReplyTo = pub.IRI(commentTrail.URL)

	activity := pub.ActivityNew(pub.IRI(id), typ, commentObject)
	activity.Actor = pub.IRI(author)
	activity.To = pub.ItemCollection{pub.IRI(to)}
	activity.Published = time.Now()
	activity.Object = commentObject

	collection, err := app.FindCollectionByNameOrId("activitypub_activities")
	if err != nil {
		return err
	}

	record := core.NewRecord(collection)
	record.Set("id", activityRecordId)
	record.Set("iri", id)
	record.Set("to", to)
	record.Set("type", string(typ))
	record.Set("object", commentObject)
	record.Set("actor", author)
	record.Set("published", time.Now())

	err = app.Save(record)
	if err != nil {
		return err
	}

	return PostActivity(app, commentAuthor, activity, []string{to + "/inbox"})

}
