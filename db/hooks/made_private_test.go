package hooks

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/meilisearch/meilisearch-go"
	"github.com/pocketbase/pocketbase/core"
	pbtests "github.com/pocketbase/pocketbase/tests"
)

// A public trail or list that is switched back to private was already handed
// out. The update has to retract it with the same Delete a real deletion
// sends, or the copies on other servers stay up. An update that does not
// change privacy sends nothing. An actor the record is still shared with
// keeps their copy. processDeleteTrailActivity and processDeleteListActivity
// delete that copy for the whole server, so every inbox on the share
// holder's host is left out of the Delete. An inbox with no parseable host
// is compared as the raw string.

func TestTrailMadePrivateReachesFollower(t *testing.T) {
	app, author, inbox := newMadePrivateTest(t)
	bindMadePrivate(app, "trails", announceTrailMadePrivate)

	trail := newAuthoredRecord(t, app, "trails", "trail", author, true)
	resavePublic(t, app, trail, false)

	assertOneDeleteDelivered(t, app, inbox, trail.GetString("iri"))
	if _, err := app.FindRecordById("trails", trail.Id); err != nil {
		t.Fatalf("making a trail private should keep the local record: %v", err)
	}
}

func TestTrailUpdateWithoutPrivacyChangeSendsNoDelete(t *testing.T) {
	for _, public := range []bool{true, false} {
		t.Run(map[bool]string{true: "public", false: "private"}[public], func(t *testing.T) {
			app, author, inbox := newMadePrivateTest(t)
			bindMadePrivate(app, "trails", announceTrailMadePrivate)

			trail := newAuthoredRecord(t, app, "trails", "trail", author, public)
			resavePublic(t, app, trail, public)

			assertNoDelete(t, app, inbox)
		})
	}
}

func TestListMadePrivateReachesFollower(t *testing.T) {
	app, author, inbox := newMadePrivateTest(t)
	bindMadePrivate(app, "lists", announceListMadePrivate)

	list := newAuthoredRecord(t, app, "lists", "list", author, true)
	resavePublic(t, app, list, false)

	assertOneDeleteDelivered(t, app, inbox, list.GetString("iri"))
	if _, err := app.FindRecordById("lists", list.Id); err != nil {
		t.Fatalf("making a list private should keep the local record: %v", err)
	}
}

func TestListUpdateWithoutPrivacyChangeSendsNoDelete(t *testing.T) {
	for _, public := range []bool{true, false} {
		t.Run(map[bool]string{true: "public", false: "private"}[public], func(t *testing.T) {
			app, author, inbox := newMadePrivateTest(t)
			bindMadePrivate(app, "lists", announceListMadePrivate)

			list := newAuthoredRecord(t, app, "lists", "list", author, public)
			resavePublic(t, app, list, public)

			assertNoDelete(t, app, inbox)
		})
	}
}

// The share row is what still grants access. A follower who is also a current
// share holder shows up in both halves of the audience, and neither half may
// deliver the Delete. Someone who holds a copy for another reason still has
// to be told: a commenter for a trail, a revoked share for a list.
func TestTrailMadePrivateSkipsCurrentShareHolder(t *testing.T) {
	app, author, followerInbox := newMadePrivateTest(t)
	bindMadePrivate(app, "trails", announceTrailMadePrivate)

	commenterInbox := newInboxCounter(t)
	shareeInbox := newInboxCounter(t)
	commenter := newRemoteActor(t, app, "commenter", commenterInbox)
	sharee := newRemoteActor(t, app, "sharee", shareeInbox)
	acceptFollow(t, app, sharee, author)

	trail := newAuthoredRecord(t, app, "trails", "trail", author, true)
	commentOn(t, app, commenter, trail)
	shareWith(t, app, "trail_share", "trail", sharee, trail)

	resavePublic(t, app, trail, false)

	assertOneDeleteDelivered(t, app, followerInbox, trail.GetString("iri"))
	assertOneDeleteDelivered(t, app, commenterInbox, trail.GetString("iri"))
	assertNoDeleteDeliveries(t, shareeInbox)
	if shares, err := app.FindAllRecords("trail_share"); err != nil || len(shares) != 1 {
		t.Fatalf("the share should outlive the trail going private: %d, err %v", len(shares), err)
	}
}

func TestListMadePrivateSkipsCurrentShareHolder(t *testing.T) {
	app, author, followerInbox := newMadePrivateTest(t)
	bindMadePrivate(app, "lists", announceListMadePrivate)

	revokedInbox := newInboxCounter(t)
	shareeInbox := newInboxCounter(t)
	revoked := newRemoteActor(t, app, "revoked", revokedInbox)
	sharee := newRemoteActor(t, app, "sharee", shareeInbox)
	acceptFollow(t, app, sharee, author)

	list := newAuthoredRecord(t, app, "lists", "list", author, true)
	recordAnnounce(t, app, list, revoked)
	shareWith(t, app, "list_share", "list", sharee, list)

	resavePublic(t, app, list, false)

	assertOneDeleteDelivered(t, app, followerInbox, list.GetString("iri"))
	assertOneDeleteDelivered(t, app, revokedInbox, list.GetString("iri"))
	assertNoDeleteDeliveries(t, shareeInbox)
	if shares, err := app.FindAllRecords("list_share"); err != nil || len(shares) != 1 {
		t.Fatalf("the share should outlive the list going private: %d, err %v", len(shares), err)
	}
}

// A follower who is not the share holder still loses the record if their
// inbox is on the share holder's host: the Delete is applied server-wide.
func TestMadePrivateSkipsFollowerOnShareHolderHost(t *testing.T) {
	for _, kind := range []struct {
		collection, shareCollection, shareField, iriPath string
		announce                                         func(*core.RecordEvent) error
	}{
		{"trails", "trail_share", "trail", "trail", announceTrailMadePrivate},
		{"lists", "list_share", "list", "list", announceListMadePrivate},
	} {
		t.Run(kind.collection, func(t *testing.T) {
			app, author, followerInbox := newMadePrivateTest(t)
			bindMadePrivate(app, kind.collection, kind.announce)

			shared := newInboxCounter(t)
			sharee := newRemoteActorURL(t, app, "sharee", shared.pathURL("/sharee/inbox"))
			neighbor := newRemoteActorURL(t, app, "neighbor", shared.pathURL("/neighbor/inbox"))
			acceptFollow(t, app, sharee, author)
			acceptFollow(t, app, neighbor, author)

			record := newAuthoredRecord(t, app, kind.collection, kind.iriPath, author, true)
			shareWith(t, app, kind.shareCollection, kind.shareField, sharee, record)
			resavePublic(t, app, record, false)

			assertOneDeleteDelivered(t, app, followerInbox, record.GetString("iri"))
			assertNoDeleteDeliveries(t, shared)
		})
	}
}

// Production binds UpdateListHandler, which also writes the search document.
// Saving a public list to private through that handler has to deliver the
// Delete; if the handler stops announcing, this assertion is what fails.
func TestUpdateListHandlerAnnouncesMadePrivate(t *testing.T) {
	app, author, inbox := newMadePrivateTest(t)
	addListTrailsRelation(t, app)
	list := newAuthoredRecord(t, app, "lists", "list", author, true)

	app.OnRecordAfterUpdateSuccess("lists").BindFunc(UpdateListHandler(newDocumentUpdateServer(t, "lists")))
	resavePublic(t, app, list, false)

	assertOneDeleteDelivered(t, app, inbox, list.GetString("iri"))
	if _, err := app.FindRecordById("lists", list.Id); err != nil {
		t.Fatalf("making a list private should keep the local record: %v", err)
	}
}

// UpdateTrailHandler has the same announcement after its search write.
func TestUpdateTrailHandlerAnnouncesMadePrivate(t *testing.T) {
	app, author, inbox := newMadePrivateTest(t)
	addTrailSearchRelations(t, app)
	trail := newAuthoredRecord(t, app, "trails", "trail", author, true)

	app.OnRecordAfterUpdateSuccess("trails").BindFunc(UpdateTrailHandler(newDocumentUpdateServer(t, "trails")))
	resavePublic(t, app, trail, false)

	assertOneDeleteDelivered(t, app, inbox, trail.GetString("iri"))
	if _, err := app.FindRecordById("trails", trail.Id); err != nil {
		t.Fatalf("making a trail private should keep the local record: %v", err)
	}
}

// newDocumentUpdateServer accepts the document PUT UpdateList and UpdateTrail
// send before they announce. The body is the task Meilisearch returns.
func newDocumentUpdateServer(t *testing.T, index string) meilisearch.ServiceManager {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPut && r.URL.Path == "/indexes/"+index+"/documents" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusAccepted)
			fmt.Fprintf(w, `{"taskUid":1,"status":"enqueued","indexUid":%q,"type":"documentAdditionOrUpdate"}`, index)
			return
		}
		http.Error(w, "unexpected request "+r.Method+" "+r.URL.RequestURI(), http.StatusNotFound)
	}))
	t.Cleanup(server.Close)
	return meilisearch.New(server.URL)
}

func addListTrailsRelation(t *testing.T, app *pbtests.TestApp) {
	t.Helper()

	lists, err := app.FindCollectionByNameOrId("lists")
	if err != nil {
		t.Fatal(err)
	}
	trails, err := app.FindCollectionByNameOrId("trails")
	if err != nil {
		t.Fatal(err)
	}
	lists.Fields.Add(&core.RelationField{Name: "trails", CollectionId: trails.Id, MaxSelect: 100})
	if err := app.Save(lists); err != nil {
		t.Fatal(err)
	}
}

func addTrailSearchRelations(t *testing.T, app *pbtests.TestApp) {
	t.Helper()

	tags := core.NewBaseCollection("tags")
	categories := core.NewBaseCollection("categories")
	for _, collection := range []*core.Collection{tags, categories} {
		if err := app.Save(collection); err != nil {
			t.Fatal(err)
		}
	}
	trails, err := app.FindCollectionByNameOrId("trails")
	if err != nil {
		t.Fatal(err)
	}
	trails.Fields.Add(
		&core.RelationField{Name: "tags", CollectionId: tags.Id, MaxSelect: 100},
		&core.RelationField{Name: "category", CollectionId: categories.Id, MaxSelect: 1},
	)
	if err := app.Save(trails); err != nil {
		t.Fatal(err)
	}
}

func newMadePrivateTest(t *testing.T) (*pbtests.TestApp, *core.Record, *inboxCounter) {
	t.Helper()

	app := setupActorDeleteHooksTestApp(t)
	lists, err := app.FindCollectionByNameOrId("lists")
	if err != nil {
		t.Fatal(err)
	}
	lists.Fields.Add(&core.TextField{Name: "iri"}, &core.BoolField{Name: "public"})
	if err := app.Save(lists); err != nil {
		t.Fatal(err)
	}

	inbox := newInboxCounter(t)
	_, author := newDepartingActor(t, app, inbox)
	return app, author, inbox
}

// bindMadePrivate wires the retraction the update handler calls, without the
// Meilisearch write that handler also does.
func bindMadePrivate(app *pbtests.TestApp, collection string, announce func(*core.RecordEvent) error) {
	app.OnRecordAfterUpdateSuccess(collection).BindFunc(func(e *core.RecordEvent) error {
		if err := announce(e); err != nil {
			return err
		}
		return e.Next()
	})
}

func newAuthoredRecord(t *testing.T, app *pbtests.TestApp, collection, iriPath string, author *core.Record, public bool) *core.Record {
	t.Helper()

	coll, err := app.FindCollectionByNameOrId(collection)
	if err != nil {
		t.Fatal(err)
	}
	record := core.NewRecord(coll)
	record.Set("author", author.Id)
	record.Set("public", public)
	if err := app.Save(record); err != nil {
		t.Fatal(err)
	}
	record.Set("iri", "https://local.example/api/v1/"+iriPath+"/"+record.Id)
	if err := app.Save(record); err != nil {
		t.Fatal(err)
	}
	return record
}

// resavePublic reloads the record first. Original() is the row as loaded, and
// a record that was inserted in this test still carries its blank defaults
// until then, so a public-to-private change would be invisible.
func resavePublic(t *testing.T, app *pbtests.TestApp, record *core.Record, public bool) {
	t.Helper()

	fresh, err := app.FindRecordById(record.Collection().Name, record.Id)
	if err != nil {
		t.Fatal(err)
	}
	fresh.Set("public", public)
	if err := app.Save(fresh); err != nil {
		t.Fatal(err)
	}
}

func newRemoteActor(t *testing.T, app *pbtests.TestApp, name string, inbox *inboxCounter) *core.Record {
	t.Helper()
	return newRemoteActorURL(t, app, name, inbox.url())
}

func newRemoteActorURL(t *testing.T, app *pbtests.TestApp, name, inboxURL string) *core.Record {
	t.Helper()

	actors, err := app.FindCollectionByNameOrId("activitypub_actors")
	if err != nil {
		t.Fatal(err)
	}
	actor := core.NewRecord(actors)
	actor.Set("iri", "https://remote.example/api/v1/activitypub/user/"+name)
	actor.Set("inbox", inboxURL)
	actor.Set("is_local", false)
	if err := app.Save(actor); err != nil {
		t.Fatal(err)
	}
	return actor
}

func acceptFollow(t *testing.T, app *pbtests.TestApp, follower, followee *core.Record) {
	t.Helper()

	follows, err := app.FindCollectionByNameOrId("follows")
	if err != nil {
		t.Fatal(err)
	}
	follow := core.NewRecord(follows)
	follow.Set("follower", follower.Id)
	follow.Set("followee", followee.Id)
	follow.Set("status", "accepted")
	if err := app.Save(follow); err != nil {
		t.Fatal(err)
	}
}

func shareWith(t *testing.T, app *pbtests.TestApp, collection, field string, actor, item *core.Record) {
	t.Helper()

	coll, err := app.FindCollectionByNameOrId(collection)
	if err != nil {
		t.Fatal(err)
	}
	share := core.NewRecord(coll)
	share.Set("actor", actor.Id)
	share.Set(field, item.Id)
	if err := app.Save(share); err != nil {
		t.Fatal(err)
	}
}

func commentOn(t *testing.T, app *pbtests.TestApp, actor, trail *core.Record) {
	t.Helper()

	comments, err := app.FindCollectionByNameOrId("comments")
	if err != nil {
		t.Fatal(err)
	}
	comment := core.NewRecord(comments)
	comment.Set("author", actor.Id)
	comment.Set("trail", trail.Id)
	if err := app.Save(comment); err != nil {
		t.Fatal(err)
	}
}

// recordAnnounce files the Announce a share sent, as the holder query reads
// it back: the actor's IRI in to, the object's id inside object.
func recordAnnounce(t *testing.T, app *pbtests.TestApp, object, actor *core.Record) {
	t.Helper()

	activities, err := app.FindCollectionByNameOrId("activitypub_activities")
	if err != nil {
		t.Fatal(err)
	}
	record := core.NewRecord(activities)
	record.Set("iri", "https://local.example/api/v1/activitypub/activity/announce-"+actor.Id)
	record.Set("type", "Announce")
	record.Set("object", map[string]any{"id": object.GetString("iri"), "type": "Note"})
	record.Set("to", []string{actor.GetString("iri")})
	if err := app.Save(record); err != nil {
		t.Fatal(err)
	}
}

func assertOneDeleteDelivered(t *testing.T, app *pbtests.TestApp, inbox *inboxCounter, objectIRI string) {
	t.Helper()

	if got := inbox.waitForHits(1, 5*time.Second); got != 1 {
		t.Fatalf("expected exactly one delivery to %s, got %d", inbox.url(), got)
	}
	// A second copy can leave just after the first is counted.
	time.Sleep(200 * time.Millisecond)
	if got := inbox.hits.Load(); got != 1 {
		t.Fatalf("expected exactly one delivery to %s, got %d", inbox.url(), got)
	}

	deletes := deleteActivities(t, app)
	if len(deletes) != 1 {
		t.Fatalf("expected one Delete activity, found %d", len(deletes))
	}
	var object string
	if err := deletes[0].UnmarshalJSONField("object", &object); err != nil {
		t.Fatal(err)
	}
	if object != objectIRI {
		t.Fatalf("Delete object = %q, want %q", object, objectIRI)
	}
}

func assertNoDelete(t *testing.T, app *pbtests.TestApp, inbox *inboxCounter) {
	t.Helper()

	if deletes := deleteActivities(t, app); len(deletes) != 0 {
		t.Fatalf("expected no Delete activity, found %d", len(deletes))
	}
	assertNoDeleteDeliveries(t, inbox)
}

func assertNoDeleteDeliveries(t *testing.T, inbox *inboxCounter) {
	t.Helper()

	if got := inbox.waitForHits(1, 500*time.Millisecond); got != 0 {
		t.Fatalf("expected no delivery to %s, got %d", inbox.url(), got)
	}
}

func deleteActivities(t *testing.T, app *pbtests.TestApp) []*core.Record {
	t.Helper()

	records, err := app.FindAllRecords("activitypub_activities")
	if err != nil {
		t.Fatal(err)
	}
	deletes := make([]*core.Record, 0)
	for _, record := range records {
		if record.GetString("type") == "Delete" {
			deletes = append(deletes, record)
		}
	}
	return deletes
}
