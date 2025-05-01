package federation

import (
	"fmt"
	"net/http"
	"strconv"

	"bytes"
	"io"

	pub "github.com/go-ap/activitypub"
	"github.com/go-ap/jsonld"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"

	"github.com/go-fed/httpsig"
)

func get_outbox(e *core.RequestEvent) error {
	username := e.Request.PathValue("username")
	page := e.Request.URL.Query().Get("page")

	user, err := e.App.FindFirstRecordByData("users", "username", username)
	if err != nil {
		return err
	}
	actor, err := e.App.FindFirstRecordByData("activitypub_actors", "user", user.Id)
	if err != nil {
		return err
	}

	iri := actor.GetString("IRI")
	orderedCollection := pub.OrderedCollectionNew(pub.IRI(iri + "/outbox"))
	orderedCollection.First = pub.IRI(iri + "/outbox?page=1")

	outbox := pub.OrderedCollectionPageNew(orderedCollection)
	outbox.First = pub.IRI(iri + "/outbox?page=1")

	totalActivities, err := e.App.CountRecords("activitypub_activities", dbx.HashExp{"user": user.Id})
	if err != nil {
		return err
	}
	outbox.TotalItems = uint(totalActivities)

	if page != "" {
		intPage, err := strconv.Atoi(page)
		if err != nil {
			return err
		}
		limit := 10
		activities, err := e.App.FindRecordsByFilter(
			"activitypub_activities",
			"user = '{:user}'",          // filter
			"-created",                  // sort
			limit,                       // limit
			intPage*limit,               // offset
			dbx.Params{"user": user.Id}, // optional filter params
		)
		if err != nil {
			return err
		}

		if intPage > 1 {
			outbox.Prev = pub.IRI(fmt.Sprintf("%s/outbox?page=%d", iri, intPage-1))
		}
		if intPage*limit+limit < int(totalActivities) {
			outbox.Next = pub.IRI(fmt.Sprintf("%s/outbox?page=%d", iri, intPage+1))
		}
		for _, a := range activities {
			pubActivity, err := pub.UnmarshalJSON([]byte(a.GetString("activity")))
			if err != nil {
				return err
			}
			err = outbox.OrderedItems.Append(pubActivity)
			if err != nil {
				return err
			}
		}
	}
	e.Response.Header().Add("Content-Type", jsonld.ContentType)
	e.Response.WriteHeader(http.StatusOK)

	binary, err := jsonld.WithContext(
		jsonld.IRI(pub.ActivityBaseURI),
		jsonld.IRI(pub.SecurityContextURI),
	).Marshal(outbox)

	e.Response.Write(binary)
	return err
}

func post_outbox(e *core.RequestEvent) error {
	body, err := io.ReadAll(e.Request.Body)
	if err != nil {
		return err
	}
	defer e.Request.Body.Close()
	var activity pub.Activity
	err = activity.UnmarshalJSON(body)
	if err != nil {
		return err
	}
	username := e.Request.PathValue("username")

	user, err := e.App.FindFirstRecordByData("users", "username", username)
	if err != nil {
		return err
	}
	actor, err := e.App.FindFirstRecordByData("activitypub_actors", "user", user.Id)
	if err != nil {
		return err
	}

	PostActivity(e.App, actor, &activity)

	return nil
}

func PostActivity(app core.App, actor *core.Record, activity *pub.Activity) error {

	algs := []httpsig.Algorithm{httpsig.RSA_SHA512, httpsig.RSA_SHA256}
	postHeaders := []string{"(request-target)", "Date", "Digest"}
	expiresIn := 60

	// create HTTP signer
	signer, _, err := httpsig.NewSigner(algs, httpsig.DigestSha256, postHeaders, httpsig.Signature, int64(expiresIn))
	if err != nil {
		return err
	}

	// marshal activity
	body, err := jsonld.WithContext(
		jsonld.IRI(pub.ActivityBaseURI),
		jsonld.IRI(pub.SecurityContextURI),
	).Marshal(activity)
	if err != nil {
		return err
	}

	client := &http.Client{}

	// for each recipient
	for _, v := range activity.To {
		buf := bytes.NewBuffer(body)
		req, err := http.NewRequest(http.MethodPost, string(v.GetLink()), buf)
		if err != nil {
			return err
		}
		req.Header.Add("Content-Type", `application/ld+json; profile="https://www.w3.org/ns/activitystreams"`)

		err = signer.SignRequest(actor.GetString("public_key"), actor.GetString("private_key"), req, body)
		if err != nil {
			return err
		}
		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return fmt.Errorf("error sending request to inbox %s (%d): %s", v, resp.StatusCode, string(body))
		}
	}
	// store activity in local db
	collection, err := app.FindCollectionByNameOrId("activitypub_activities")
	if err != nil {
		return err
	}
	record := core.NewRecord(collection)
	record.Set("activity", string(body))
	record.Set("user", actor.GetString("user"))

	return app.Save(record)

}
