package main

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	pub "github.com/go-ap/activitypub"
	"github.com/meilisearch/meilisearch-go"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"

	"pocketbase/federation"
)

// A remote trail deleted at its origin takes its photos along, those of its
// waypoints and summit logs included. They belong to the remote author, so
// nobody here could find or remove them afterwards.
func TestInboundTrailDeleteRemovesFederatedAssets(t *testing.T) {
	f := newAssetCleanupFixture(t)
	bob := f.save("activitypub_actors", map[string]any{"preferred_username": "bob",
		"iri": "https://remote.example/api/v1/activitypub/user/bob", "is_local": false})
	trail := f.save("trails", map[string]any{"name": "Ridge", "author": bob.Id, "public": true,
		"iri": "https://remote.example/api/v1/trail/ridge"})
	other := f.save("trails", map[string]any{"name": "Valley", "author": bob.Id, "public": true,
		"iri": "https://remote.example/api/v1/trail/valley"})
	waypoint := f.save("waypoints", map[string]any{"name": "Hut", "trail": trail.Id, "author": bob.Id,
		"iri": "https://remote.example/api/v1/waypoint/hut"})
	log := f.save("summit_logs", map[string]any{"trail": trail.Id, "author": bob.Id, "date": "2026-01-01 00:00:00.000Z",
		"iri": "https://remote.example/api/v1/summit-log/ridge"})
	trailPhoto := f.federatedPhoto(bob, "trail")
	waypointPhoto := f.federatedPhoto(bob, "waypoint")
	logPhoto := f.federatedPhoto(bob, "log")
	shared := f.federatedPhoto(bob, "shared")
	f.save("trail_assets", map[string]any{"asset": trailPhoto.Id, "trail": trail.Id})
	f.save("waypoint_assets", map[string]any{"asset": waypointPhoto.Id, "waypoint": waypoint.Id})
	f.save("summit_log_assets", map[string]any{"asset": logPhoto.Id, "summit_log": log.Id})
	f.save("trail_assets", map[string]any{"asset": shared.Id, "trail": trail.Id})
	f.save("trail_assets", map[string]any{"asset": shared.Id, "trail": other.Id})
	f.bindHooks()

	activity := pub.DeleteNew(pub.IRI("https://remote.example/api/v1/activitypub/activity/delete"), pub.IRI(trail.GetString("iri")))
	activity.Actor = pub.IRI(bob.GetString("iri"))
	if err := federation.ProcessDeleteActivity(f.app, bob, *activity); err != nil {
		t.Fatal(err)
	}

	if f.exists(trail) {
		t.Fatal("trail was not deleted")
	}
	for name, photo := range map[string]*core.Record{"trail": trailPhoto, "waypoint": waypointPhoto, "summit log": logPhoto} {
		if f.exists(photo) {
			t.Errorf("photo of the %s survived the trail's deletion", name)
		}
	}
	if !f.exists(shared) {
		t.Error("photo still linked to another trail was deleted")
	}
}

// A deleted trail takes the photos along that nothing else links to, whoever
// added them, and its waypoints and summit logs take theirs. The hooks of that
// cascade run after the commit and report their errors through the trail's
// delete, which then skips the trail's own after-delete hooks; its photos have
// to go all the same. A trail merge deletes its source inside a transaction.
func TestTrailDeleteRemovesOrphanedAssets(t *testing.T) {
	for _, tc := range []struct {
		name         string
		cascadeFails bool
		inTx         bool
	}{
		{name: "hooks succeed"},
		{name: "cascade hook fails", cascadeFails: true},
		{name: "inside a transaction", inTx: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := newAssetCleanupFixture(t)
			alice := f.localActor("alice")
			carol := f.localActor("carol")
			bob := f.save("activitypub_actors", map[string]any{"preferred_username": "bob",
				"iri": "https://remote.example/api/v1/activitypub/user/bob", "is_local": false})
			trail := f.save("trails", map[string]any{"name": "Ridge", "author": alice.Id})
			other := f.save("trails", map[string]any{"name": "Valley", "author": alice.Id})
			waypoint := f.save("waypoints", map[string]any{"name": "Hut", "trail": trail.Id, "author": alice.Id})
			carolLog := f.save("summit_logs", map[string]any{"trail": trail.Id, "author": carol.Id, "date": "2026-01-01 00:00:00.000Z"})
			bobLog := f.save("summit_logs", map[string]any{"trail": trail.Id, "author": bob.Id, "date": "2026-01-02 00:00:00.000Z",
				"iri": "https://remote.example/api/v1/summit-log/ridge"})
			otherLog := f.save("summit_logs", map[string]any{"trail": other.Id, "author": carol.Id, "date": "2026-01-03 00:00:00.000Z"})

			// Added by Carol, who may edit the trail.
			trailPhoto := f.photo(carol)
			waypointPhoto := f.photo(alice)
			carolPhoto := f.photo(carol)
			bobPhoto := f.federatedPhoto(bob, "log")
			sharedTrailPhoto := f.photo(alice)
			sharedLogPhoto := f.photo(carol)
			f.save("trail_assets", map[string]any{"asset": trailPhoto.Id, "trail": trail.Id})
			f.save("waypoint_assets", map[string]any{"asset": waypointPhoto.Id, "waypoint": waypoint.Id})
			f.save("summit_log_assets", map[string]any{"asset": carolPhoto.Id, "summit_log": carolLog.Id})
			f.save("summit_log_assets", map[string]any{"asset": bobPhoto.Id, "summit_log": bobLog.Id})
			f.save("trail_assets", map[string]any{"asset": sharedTrailPhoto.Id, "trail": trail.Id})
			f.save("trail_assets", map[string]any{"asset": sharedTrailPhoto.Id, "trail": other.Id})
			f.save("waypoint_assets", map[string]any{"asset": sharedLogPhoto.Id, "waypoint": waypoint.Id})
			f.save("summit_log_assets", map[string]any{"asset": sharedLogPhoto.Id, "summit_log": otherLog.Id})
			f.bindHooks()
			if tc.cascadeFails {
				f.app.OnRecordAfterDeleteSuccess("waypoints").BindFunc(func(e *core.RecordEvent) error {
					return errors.New("waypoint hook failed")
				})
			}

			var err error
			if tc.inTx {
				err = f.app.RunInTransaction(func(txApp core.App) error {
					return txApp.Delete(trail)
				})
			} else {
				err = f.app.Delete(trail)
			}
			if tc.cascadeFails != (err != nil) {
				t.Fatalf("delete error = %v", err)
			}

			if f.exists(trail) {
				t.Fatal("trail was not deleted")
			}
			for name, photo := range map[string]*core.Record{
				"trail":                trailPhoto,
				"waypoint":             waypointPhoto,
				"local summit log":     carolPhoto,
				"federated summit log": bobPhoto,
			} {
				if f.exists(photo) {
					t.Errorf("photo of the %s survived the trail's deletion", name)
				}
			}
			if !f.exists(sharedTrailPhoto) {
				t.Error("photo still linked to another trail was deleted")
			}
			if !f.exists(sharedLogPhoto) {
				t.Error("photo still linked to another trail's summit log was deleted")
			}
		})
	}
}

// assetCleanupFixture stores its records before the hooks are bound, against
// a search stub that accepts every request.
type assetCleanupFixture struct {
	t   *testing.T
	app *core.BaseApp
}

func newAssetCleanupFixture(t *testing.T) *assetCleanupFixture {
	return &assetCleanupFixture{t: t, app: newRealSchemaApp(t)}
}

func (f *assetCleanupFixture) bindHooks() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodGet {
			w.Write([]byte(`{"uid":1,"status":"succeeded"}`))
			return
		}
		w.WriteHeader(http.StatusAccepted)
		w.Write([]byte(`{"taskUid":1,"status":"enqueued"}`))
	}))
	f.t.Cleanup(server.Close)
	setupEventHandlers(&pocketbase.PocketBase{App: f.app}, meilisearch.New(server.URL))
	// The link hooks reindex a trail half a second after its links change, in
	// the background, which must not find the app torn down: the reindex
	// recovers from the closed database, but the race detector still reports it.
	f.t.Cleanup(func() { time.Sleep(time.Second) })
}

func (f *assetCleanupFixture) save(name string, values map[string]any) *core.Record {
	f.t.Helper()
	collection, err := f.app.FindCollectionByNameOrId(name)
	if err != nil {
		f.t.Fatal(err)
	}
	record := core.NewRecord(collection)
	record.Load(values)
	if err := f.app.SaveNoValidate(record); err != nil {
		f.t.Fatal(name, err)
	}
	return record
}

func (f *assetCleanupFixture) localActor(name string) *core.Record {
	user := f.save("users", map[string]any{"email": name + "@example.com", "username": name})
	return f.save("activitypub_actors", map[string]any{"user": user.Id, "preferred_username": name,
		"iri": "https://example.com/api/v1/activitypub/user/" + name, "is_local": true})
}

func (f *assetCleanupFixture) photo(author *core.Record) *core.Record {
	return f.save("assets", map[string]any{"type": "photo", "storage_mode": "copy", "author": author.Id})
}

// federatedPhoto is a copy of a remote author's photo, as federation stores it.
func (f *assetCleanupFixture) federatedPhoto(author *core.Record, name string) *core.Record {
	return f.save("assets", map[string]any{"type": "photo", "storage_mode": "copy", "author": author.Id,
		"external_provider": "federation", "external_id": "https://remote.example/api/v1/assets/" + name})
}

func (f *assetCleanupFixture) exists(record *core.Record) bool {
	_, err := f.app.FindRecordById(record.Collection(), record.Id)
	return err == nil
}
