package federation

import (
	"fmt"
	"os"
	"time"

	"pocketbase/models"
	"pocketbase/util"

	pub "github.com/go-ap/activitypub"
	"github.com/meilisearch/meilisearch-go"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/security"
)

// create outgoing follow activity
func CreateTrailCreateActivity(app core.App, trail *core.Record) error {
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

	activity := pub.CreateNew(pub.IRI(id), trailObject)
	activity.Actor = pub.IRI(trailAuthor.GetString("iri"))
	activity.To = pub.ItemCollection{pub.IRI(to)}
	activity.CC = pub.ItemCollection{pub.IRI(cc)}

	record := core.NewRecord(collection)
	record.Set("id", recordId)
	record.Set("iri", id)
	record.Set("to", to)
	record.Set("cc", cc)
	record.Set("type", string(pub.CreateType))
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

	// for now we only process "Trail" activities
	if activity.Object.GetType() != "Trail" {
		return nil
	}

	// no need to do anything if the actor is local
	// if actor.GetBool("isLocal") {
	// 	return nil
	// }

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
