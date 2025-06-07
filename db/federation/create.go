package federation

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"pocketbase/util"

	pub "github.com/go-ap/activitypub"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/filesystem"
	"github.com/pocketbase/pocketbase/tools/security"
)

func CreateTrailActivity(app core.App, trail *core.Record, typ pub.ActivityVocabularyType) error {
	if !trail.GetBool("public") {
		// only broadcast the trail if it is public
		return nil
	}
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

	category := ""
	categoryRecord := trail.ExpandedOne("category")
	if categoryRecord != nil {
		category = categoryRecord.GetString("name")
	}

	tagRecords := trail.ExpandedAll("tags")
	tags := pub.ItemCollection{
		pub.Object{
			Type:    pub.NoteType,
			Name:    pub.NaturalLanguageValuesNew(pub.LangRefValueNew(pub.NilLangRef, "category")),
			Content: pub.NaturalLanguageValuesNew(pub.LangRefValueNew(pub.NilLangRef, category)),
		},
		pub.Object{
			Type:    pub.NoteType,
			Name:    pub.NaturalLanguageValuesNew(pub.LangRefValueNew(pub.NilLangRef, "difficulty")),
			Content: pub.NaturalLanguageValuesNew(pub.LangRefValueNew(pub.NilLangRef, trail.GetString("difficulty"))),
		},
		pub.Object{
			Type:    pub.NoteType,
			Name:    pub.NaturalLanguageValuesNew(pub.LangRefValueNew(pub.NilLangRef, "elevation_gain")),
			Content: pub.NaturalLanguageValuesNew(pub.LangRefValueNew(pub.NilLangRef, fmt.Sprintf("%fm", trail.GetFloat("elevation_gain")))),
		},
		pub.Object{
			Type:    pub.NoteType,
			Name:    pub.NaturalLanguageValuesNew(pub.LangRefValueNew(pub.NilLangRef, "elevation_loss")),
			Content: pub.NaturalLanguageValuesNew(pub.LangRefValueNew(pub.NilLangRef, fmt.Sprintf("%fm", trail.GetFloat("elevation_loss")))),
		},
		pub.Object{
			Type:    pub.NoteType,
			Name:    pub.NaturalLanguageValuesNew(pub.LangRefValueNew(pub.NilLangRef, "distance")),
			Content: pub.NaturalLanguageValuesNew(pub.LangRefValueNew(pub.NilLangRef, fmt.Sprintf("%fm", trail.GetFloat("distance")))),
		},
		pub.Object{
			Type:    pub.NoteType,
			Name:    pub.NaturalLanguageValuesNew(pub.LangRefValueNew(pub.NilLangRef, "duration")),
			Content: pub.NaturalLanguageValuesNew(pub.LangRefValueNew(pub.NilLangRef, fmt.Sprintf("%fm", trail.GetFloat("duration")))),
		},
	}

	for _, v := range tagRecords {
		tags.Append(pub.IRI(v.GetString("name")))
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

	gpx := ""
	if trail.GetString("gpx") != "" {
		gpx = fmt.Sprintf("%s/api/v1/files/trails/%s/%s", origin, trail.Id, trail.GetString("gpx"))
	}

	attachments := make(pub.ItemCollection, max(len(photos), 2))
	for i := range min(len(photos), 3) {
		iri := fmt.Sprintf("%s/api/v1/files/trails/%s/%s", origin, trail.Id, photos[i])

		attachments[i] = pub.Image{
			Type:      pub.ImageType,
			MediaType: "image/jpeg",
			URL:       pub.IRI(iri),
		}
	}
	if gpx != "" {
		attachments.Append(pub.Document{
			Type:      pub.DocumentType,
			MediaType: "application/xml+gpx",
			URL:       pub.IRI(gpx),
		})
	}

	trailObject := pub.ObjectNew(pub.ArticleType)

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

	trailObject.StartTime = trail.GetDateTime("date").Time()
	trailObject.Attachment = attachments

	trailObject.Tag = tags

	activity := pub.ActivityNew(pub.IRI(id), typ, trailObject)
	activity.Actor = pub.IRI(trailAuthor.GetString("iri"))
	activity.To = pub.ItemCollection{pub.IRI(to)}
	activity.CC = pub.ItemCollection{pub.IRI(cc)}
	activity.Published = time.Now()

	record := core.NewRecord(collection)
	record.Set("id", recordId)
	record.Set("iri", id)
	record.Set("to", []string{to})
	record.Set("cc", []string{cc})
	record.Set("type", string(typ))
	record.Set("object", trailObject)
	record.Set("actor", trailAuthor.GetString("iri"))
	record.Set("published", time.Now())

	err = app.Save(record)
	if err != nil {
		return err
	}

	follows, err := app.FindRecordsByFilter("follows", "followee={:followee}&&status='accepted'", "", -1, 0, dbx.Params{"followee": trailAuthor.Id})
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

func CreateCommentActivity(app core.App, comment *core.Record, typ pub.ActivityVocabularyType) error {

	origin := os.Getenv("ORIGIN")
	if origin == "" {
		return fmt.Errorf("ORIGIN not set")
	}

	// author of the comment
	commentAuthor, err := app.FindRecordById("activitypub_actors", comment.GetString("author"))
	if err != nil {
		return err
	}

	commentTrail, err := app.FindRecordById("trails", comment.GetString("trail"))
	if err != nil {
		return err
	}
	commentTrailAuthor, err := app.FindRecordById("activitypub_actors", commentTrail.GetString("author"))
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

	trailURL := ""
	if commentTrailAuthor.GetBool("isLocal") {
		trailURL = fmt.Sprintf("https://%s/api/v1/%s", commentTrailAuthor.GetString("domain"), comment.GetString("trail"))
	} else {
		trailURL = commentTrail.GetString("iri")
	}

	author := commentAuthor.GetString("iri")

	commentObject := pub.ObjectNew(pub.NoteType)
	commentObject.ID = pub.IRI(fmt.Sprintf("%s/api/v1/comment/%s", origin, commentRecordId))
	commentObject.Content = pub.NaturalLanguageValuesNew(pub.LangRefValueNew(pub.NilLangRef, comment.GetString("text")))
	commentObject.Published = comment.GetDateTime("created").Time()
	commentObject.AttributedTo = pub.IRI(author)
	commentObject.InReplyTo = pub.IRI(trailURL)

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
	record.Set("to", []string{to})
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

func CreateSummitLogActivity(app core.App, summitLog *core.Record, typ pub.ActivityVocabularyType) error {

	origin := os.Getenv("ORIGIN")
	if origin == "" {
		return fmt.Errorf("ORIGIN not set")
	}

	summitLogAuthor, err := app.FindRecordById("activitypub_actors", summitLog.GetString("author"))
	if err != nil {
		return err
	}

	var summitLogAuthorId string
	// first check if we find the trail locally
	summitLogTrail, err := app.FindRecordById("trails", summitLog.GetString("trail"))
	if err != nil {
		return err
	}
	if !summitLogTrail.GetBool("public") {
		// only broadcast the log if the trail it belongs to is public
		return nil
	}
	summitLogAuthorId = summitLogTrail.GetString("author")

	summitLogTrailAuthor, err := app.FindRecordById("activitypub_actors", summitLogAuthorId)
	if err != nil {
		return err
	}

	collection, err := app.FindCollectionByNameOrId("activitypub_activities")
	if err != nil {
		return err
	}

	var trailIRI pub.IRI
	if summitLogTrailAuthor.GetBool("isLocal") {
		trailId := summitLog.GetString("trail")
		trailIRI = pub.IRI(fmt.Sprintf("%s/api/v1/trail/%s", origin, trailId))
	} else {
		trailIRI = pub.IRI(summitLogTrail.GetString("iri"))
	}

	recordId := security.RandomStringWithAlphabet(core.DefaultIdLength, core.DefaultIdAlphabet)

	id := fmt.Sprintf("%s/api/v1/activitypub/activity/%s", origin, recordId)
	to := pub.ItemCollection{pub.IRI("https://www.w3.org/ns/activitystreams#Public")}
	cc := pub.ItemCollection{pub.IRI(summitLogAuthor.GetString("iri") + "/followers")}

	// someone else created the summit log on the trail -> inform the trail's author
	if summitLogAuthor.Id != summitLogTrailAuthor.Id {
		to.Append(pub.IRI(summitLogTrailAuthor.GetString("iri")))
	}

	photos := summitLog.GetStringSlice("photos")

	gpx := ""
	if summitLog.GetString("gpx") != "" {
		gpx = fmt.Sprintf("%s/api/v1/files/summit_logs/%s/%s", origin, summitLog.Id, summitLog.GetString("gpx"))
	}

	attachments := make(pub.ItemCollection, max(len(photos), 2))
	for i := range len(photos) {
		iri := fmt.Sprintf("%s/api/v1/files/summit_logs/%s/%s", origin, summitLog.Id, photos[i])

		attachments[i] = pub.Document{
			Type:      pub.ImageType,
			MediaType: "image/jpeg",
			URL:       pub.IRI(iri),
		}
	}
	if gpx != "" {
		attachments.Append(pub.Document{
			Type:      pub.DocumentType,
			MediaType: "application/xml+gpx",
			URL:       pub.IRI(gpx),
		})
	}

	tags := pub.ItemCollection{
		pub.Object{
			Type:    pub.NoteType,
			Name:    pub.NaturalLanguageValuesNew(pub.LangRefValueNew(pub.NilLangRef, "elevation_gain")),
			Content: pub.NaturalLanguageValuesNew(pub.LangRefValueNew(pub.NilLangRef, fmt.Sprintf("%fm", summitLog.GetFloat("elevation_gain")))),
		},
		pub.Object{
			Type:    pub.NoteType,
			Name:    pub.NaturalLanguageValuesNew(pub.LangRefValueNew(pub.NilLangRef, "elevation_loss")),
			Content: pub.NaturalLanguageValuesNew(pub.LangRefValueNew(pub.NilLangRef, fmt.Sprintf("%fm", summitLog.GetFloat("elevation_loss")))),
		},
		pub.Object{
			Type:    pub.NoteType,
			Name:    pub.NaturalLanguageValuesNew(pub.LangRefValueNew(pub.NilLangRef, "distance")),
			Content: pub.NaturalLanguageValuesNew(pub.LangRefValueNew(pub.NilLangRef, fmt.Sprintf("%fm", summitLog.GetFloat("distance")))),
		},
		pub.Object{
			Type:    pub.NoteType,
			Name:    pub.NaturalLanguageValuesNew(pub.LangRefValueNew(pub.NilLangRef, "duration")),
			Content: pub.NaturalLanguageValuesNew(pub.LangRefValueNew(pub.NilLangRef, fmt.Sprintf("%fm", summitLog.GetFloat("duration")))),
		},
	}

	logObject := pub.ObjectNew(pub.NoteType)

	logObject.Content = pub.NaturalLanguageValuesNew(pub.LangRefValueNew(pub.NilLangRef, summitLog.GetString("text")))
	logObject.AttributedTo = pub.IRI(summitLogAuthor.GetString("iri"))
	logObject.Published = summitLog.GetDateTime("created").Time()
	logObject.ID = pub.IRI(fmt.Sprintf("%s/api/v1/summit-log/%s", origin, summitLog.Id))
	logObject.URL = pub.IRI(fmt.Sprintf("%s/trail/view/@%s/%s", origin, summitLogTrailAuthor.GetString("username"), summitLog.GetString("trail")))
	logObject.InReplyTo = trailIRI
	logObject.Tag = tags

	logObject.StartTime = summitLog.GetDateTime("date").Time()
	logObject.Attachment = attachments

	activity := pub.ActivityNew(pub.IRI(id), typ, logObject)
	activity.Actor = pub.IRI(summitLogAuthor.GetString("iri"))
	activity.To = to
	activity.CC = cc
	activity.Published = time.Now()

	follows, err := app.FindRecordsByFilter("follows", "followee={:followee}&&status='accepted'", "", -1, 0, dbx.Params{"followee": summitLogAuthor.Id})
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

	if summitLogAuthor.Id != summitLogTrailAuthor.Id {
		recipients = append(recipients, summitLogTrailAuthor.GetString("iri")+"/inbox")
	}

	err = PostActivity(app, summitLogAuthor, activity, recipients)
	if err != nil {
		return err
	}

	record := core.NewRecord(collection)
	record.Set("id", recordId)
	record.Set("iri", id)
	record.Set("to", to)
	record.Set("cc", cc)
	record.Set("type", string(typ))
	record.Set("object", logObject)
	record.Set("actor", summitLogAuthor.GetString("iri"))
	record.Set("published", time.Now())

	return app.Save(record)
}

func CreateListActivity(app core.App, list *core.Record, typ pub.ActivityVocabularyType) error {
	if !list.GetBool("public") {
		// only broadcast the list if it is public
		return nil
	}
	origin := os.Getenv("ORIGIN")
	if origin == "" {
		return fmt.Errorf("ORIGIN not set")
	}

	// author of the list
	listAuthor, err := app.FindRecordById("activitypub_actors", list.GetString("author"))
	if err != nil {
		return err
	}

	activityRecordId := security.RandomStringWithAlphabet(core.DefaultIdLength, core.DefaultIdAlphabet)

	id := fmt.Sprintf("%s/api/v1/activitypub/activity/%s", origin, activityRecordId)
	to := "https://www.w3.org/ns/activitystreams#Public"
	cc := listAuthor.GetString("iri") + "/followers"

	avatar := ""
	if list.GetString("avatar") != "" {
		avatar = fmt.Sprintf("%s/api/v1/files/lists/%s/%s", origin, list.Id, list.GetString("avatar"))
	}

	attachments := make(pub.ItemCollection, 2)
	if avatar != "" {
		attachments[0] = pub.Image{
			Type:      pub.ImageType,
			MediaType: "image/jpeg",
			URL:       pub.IRI(avatar),
		}
	}

	author := listAuthor.GetString("iri")

	listObject := pub.ObjectNew(pub.ArticleType)
	listObject.Name = pub.NaturalLanguageValuesNew(pub.LangRefValueNew(pub.NilLangRef, list.GetString("name")))
	listObject.Content = pub.NaturalLanguageValuesNew(pub.LangRefValueNew(pub.NilLangRef, list.GetString("description")))

	listObject.AttributedTo = pub.IRI(listAuthor.GetString("iri"))
	listObject.Published = list.GetDateTime("created").Time()
	listObject.ID = pub.IRI(fmt.Sprintf("%s/api/v1/list/%s", origin, list.Id))
	listObject.URL = pub.IRI(fmt.Sprintf("%s/lists/@%s/%s", origin, listAuthor.GetString("username"), list.Id))
	listObject.Attachment = attachments

	activity := pub.ActivityNew(pub.IRI(id), typ, listObject)
	activity.Actor = pub.IRI(author)
	activity.To = pub.ItemCollection{pub.IRI(to)}
	activity.To = pub.ItemCollection{pub.IRI(cc)}
	activity.Published = time.Now()
	activity.Object = listObject

	collection, err := app.FindCollectionByNameOrId("activitypub_activities")
	if err != nil {
		return err
	}

	follows, err := app.FindRecordsByFilter("follows", "followee={:followee}&&status='accepted'", "", -1, 0, dbx.Params{"followee": listAuthor.Id})
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

	err = PostActivity(app, listAuthor, activity, recipients)
	if err != nil {
		return err
	}

	record := core.NewRecord(collection)
	record.Set("id", activityRecordId)
	record.Set("iri", id)
	record.Set("to", []string{to})
	record.Set("type", string(typ))
	record.Set("object", listObject)
	record.Set("actor", author)
	record.Set("published", time.Now())

	return app.Save(record)
}

func ProcessCreateOrUpdateActivity(app core.App, actor *core.Record, activity pub.Activity) error {

	var err error
	if strings.Contains(activity.Object.GetID().String(), "/api/v1/trail") {
		err = processCreateOrUpdateTrailActivity(activity, app, actor)
	} else if strings.Contains(activity.Object.GetID().String(), "/api/v1/comment") {
		err = processCreateOrUpdateCommentActivity(activity, app, actor)
	} else if strings.Contains(activity.Object.GetID().String(), "/api/v1/summit-log") {
		err = processCreateOrUpdateSummitLogActivity(activity, app, actor)
	} else if strings.Contains(activity.Object.GetID().String(), "/api/v1/list") {
		err = processCreateOrUpdateListActivity(activity, app, actor)
	}

	if err != nil {
		return err
	}

	return nil

}

func processCreateOrUpdateTrailActivity(activity pub.Activity, app core.App, actor *core.Record) error {

	// no need to do anything if the actor is local
	if actor.GetBool("isLocal") {
		return nil
	}

	return util.TrailFromActivity(activity, app, actor)
}

func processCreateOrUpdateCommentActivity(activity pub.Activity, app core.App, actor *core.Record) error {

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

	trail, err := app.FindRecordById("trails", trailId)

	// if the trail is not present on this instance just ignore it
	if err != nil {
		return nil
	}

	trailAuthor, err := app.FindRecordById("activitypub_actors", trail.GetString("author"))
	if err != nil {
		return err
	}

	// no need to do anything else if the actor is local
	if actor.GetBool("isLocal") {
		return nil
	}

	record, err := app.FindFirstRecordByData("comments", "iri", commentObject.ID.String())
	if err != nil {
		if err == sql.ErrNoRows {
			collection, err := app.FindCollectionByNameOrId("comments")
			if err != nil {
				return err
			}

			record = core.NewRecord(collection)
		} else {
			return err
		}
	}

	record.Set("iri", commentObject.ID.String())
	record.Set("text", commentObject.Content.First().Value)
	record.Set("author", actor.Id)
	record.Set("trail", trailId)

	err = app.Save(record)
	if err != nil {
		return err
	}

	if activity.Type == pub.CreateType {
		// send a notification to the trail author
		notification := util.Notification{
			Type: util.TrailComment,
			Metadata: map[string]string{
				"comment":      commentObject.Content.First().Value.String(),
				"trail_id":     trailId,
				"trail_name":   trail.GetString("name"),
				"trail_author": fmt.Sprintf("@%s", trailAuthor.GetString("username")),
			},
			Seen:   false,
			Author: actor.Id,
		}
		return util.SendNotification(app, notification, trailAuthor)

	}

	return nil
}

func processCreateOrUpdateSummitLogActivity(activity pub.Activity, app core.App, actor *core.Record) error {
	logObject, err := pub.ToObject(activity.Object)
	if err != nil {
		return err
	}

	trailIRI, err := url.Parse(logObject.InReplyTo.GetID().String())
	if err != nil {
		return err
	}
	trailId := path.Base(trailIRI.Path)

	trail, err := app.FindFirstRecordByFilter("trails", "iri={:iri} || id={:id}", dbx.Params{"id": trailId, "iri": logObject.InReplyTo.GetID().String()})
	// if the trail is not present on this instance just ignore it
	if err != nil {
		return nil
	}

	trailAuthor, err := app.FindRecordById("activitypub_actors", trail.GetString("author"))
	if err != nil {
		return err
	}

	newSummitLog := false
	record, err := app.FindFirstRecordByData("summit_logs", "iri", logObject.ID.String())
	if err != nil {
		if err == sql.ErrNoRows {
			collection, err := app.FindCollectionByNameOrId("summit_logs")
			if err != nil {
				return err
			}

			record = core.NewRecord(collection)
			newSummitLog = true
		} else {
			return err
		}
	}
	// no need to do anything else if the actor is local
	if actor.GetBool("isLocal") {
		return nil
	}

	var distance, duration, elevation_gain, elevation_loss float64
	tags, err := pub.ToItemCollection(logObject.Tag)
	if err != nil {
		return err
	}

	for _, tag := range tags.Collection() {
		tagObj, err := pub.ToObject(tag)
		if err != nil {
			continue
		}
		content := tagObj.Content.First().Value.String()
		switch tagObj.Name.First().Value.String() {
		case "elevation_gain":
			elevation_gain, err = strconv.ParseFloat(content[:len(content)-1], 64)
		case "elevation_loss":
			elevation_loss, err = strconv.ParseFloat(content[:len(content)-1], 64)
		case "duration":
			duration, err = strconv.ParseFloat(content[:len(content)-1], 64)
		case "distance":
			distance, err = strconv.ParseFloat(content[:len(content)-1], 64)
		}
		if err != nil {
			continue
		}
	}

	record.Set("date", logObject.StartTime)
	record.Set("text", logObject.Content.First().Value)
	record.Set("distance", distance)
	record.Set("duration", duration)
	record.Set("elevation_gain", elevation_gain)
	record.Set("elevation_loss", elevation_loss)
	record.Set("author", actor.Id)
	record.Set("trail", trail.Id)
	record.Set("iri", logObject.ID.String())

	if logObject.Attachment != nil {
		attachments, err := pub.ToItemCollection(logObject.Attachment)
		if err != nil {
			return err
		}

		photoURLs := []string{}
		gpxURL := ""
		for _, a := range attachments.Collection() {
			attachment, err := pub.ToObject(a)
			if err != nil {
				continue
			}
			if attachment.Type == pub.DocumentType && attachment.MediaType == "application/xml+gpx" {
				gpxURL = attachment.URL.GetLink().String()
			} else if attachment.Type == pub.ImageType {
				photoURLs = append(photoURLs, attachment.URL.GetLink().String())
			}
		}

		if len(photoURLs) > 0 {
			photos := make([]*filesystem.File, len(photoURLs))
			for i, purl := range photoURLs {
				photo, err := filesystem.NewFileFromURL(context.Background(), purl)
				if err != nil {
					continue
				}
				photos[i] = photo
			}

			record.Set("photos", photos)
		}

		if gpxURL != "" {
			gpx, err := filesystem.NewFileFromURL(context.Background(), gpxURL)
			if err != nil {
				return err
			}

			record.Set("gpx", gpx)
		}
	}

	err = app.Save(record)
	if err != nil {
		return err
	}

	if newSummitLog {
		// send a notification to the trail author
		notification := util.Notification{
			Type: util.SummitLogCreate,
			Metadata: map[string]string{
				"trail_id":     trail.Id,
				"trail_name":   trail.GetString("name"),
				"trail_author": fmt.Sprintf("@%s", trailAuthor.GetString("username")),
			},
			Seen:   false,
			Author: actor.Id,
		}
		return util.SendNotification(app, notification, trailAuthor)
	}

	return nil
}

func processCreateOrUpdateListActivity(activity pub.Activity, app core.App, actor *core.Record) error {

	// no need to do anything if the actor is local
	if actor.GetBool("isLocal") {
		return nil
	}

	return util.ListFromActivity(activity, app, actor)
}
