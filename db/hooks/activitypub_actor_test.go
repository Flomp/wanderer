package hooks

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/pocketbase/pocketbase/core"
	pbtests "github.com/pocketbase/pocketbase/tests"
	"github.com/pocketbase/pocketbase/tools/security"
)

// An account deletion is announced to other instances. PocketBase runs the
// whole cascade in one transaction, and the actor is only one branch of it: a
// sibling branch can still fail afterwards and roll everything back. The
// announcement must therefore wait for the commit, or followers are told an
// account is gone that still exists.
//
// That is why these tests delete a parent record rather than the actor itself,
// with the actor cascading first and a blocking sibling second, which is the
// shape users → activitypub_actors → api_tokens had. And why the observable is
// a real HTTP delivery to a follower's inbox rather than the activity row: the
// row is written inside the transaction and rolled back with it, so only the
// network call escapes.

const testEncryptionKey = "0123456789abcdef0123456789abcdef"

type inboxCounter struct {
	server *httptest.Server
	hits   atomic.Int32
}

func newInboxCounter(t *testing.T) *inboxCounter {
	t.Helper()

	c := &inboxCounter{}
	c.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			c.hits.Add(1)
		}
		w.WriteHeader(http.StatusAccepted)
	}))
	t.Cleanup(c.server.Close)

	return c
}

func (c *inboxCounter) url() string {
	return c.server.URL + "/inbox"
}

// waitForHits polls until the inbox has seen at least want deliveries or the
// deadline passes, then returns the count. Delivery is asynchronous, so a
// "nothing arrived" assertion has to give it time to be wrong.
func (c *inboxCounter) waitForHits(want int32, deadline time.Duration) int32 {
	stop := time.Now().Add(deadline)
	for time.Now().Before(stop) {
		if c.hits.Load() >= want {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	return c.hits.Load()
}

func setupActorDeleteHooksTestApp(t *testing.T) *pbtests.TestApp {
	t.Helper()

	t.Setenv("ORIGIN", "https://local.example")
	t.Setenv("POCKETBASE_ENCRYPTION_KEY", testEncryptionKey)

	app, err := pbtests.NewTestApp(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(app.Cleanup)

	// Stands in for users: the record whose deletion fans out to the rest.
	owners := core.NewBaseCollection("owners")
	if err := app.Save(owners); err != nil {
		t.Fatal(err)
	}

	actors := core.NewBaseCollection("activitypub_actors")
	actors.Fields.Add(
		&core.TextField{Name: "iri"},
		&core.TextField{Name: "inbox"},
		&core.TextField{Name: "private_key"},
		&core.BoolField{Name: "is_local"},
		&core.RelationField{Name: "owner", CollectionId: owners.Id, MaxSelect: 1, CascadeDelete: true},
	)
	if err := app.Save(actors); err != nil {
		t.Fatal(err)
	}

	// Read by followerInboxes through a raw join.
	follows := core.NewBaseCollection("follows")
	follows.Fields.Add(
		&core.RelationField{Name: "follower", CollectionId: actors.Id, MaxSelect: 1, CascadeDelete: true},
		&core.RelationField{Name: "followee", CollectionId: actors.Id, MaxSelect: 1, CascadeDelete: true},
		&core.TextField{Name: "status"},
	)
	if err := app.Save(follows); err != nil {
		t.Fatal(err)
	}

	activities := core.NewBaseCollection("activitypub_activities")
	activities.Fields.Add(
		&core.TextField{Name: "iri"},
		&core.TextField{Name: "type"},
		&core.TextField{Name: "to"},
		&core.TextField{Name: "cc"},
		&core.TextField{Name: "object"},
		&core.TextField{Name: "actor"},
		&core.TextField{Name: "published"},
	)
	if err := app.Save(activities); err != nil {
		t.Fatal(err)
	}

	// The sibling that blocks: a required relation to the owner with no
	// cascade, which is what api_tokens.user was. PocketBase cascades in
	// collection-name order, so "activitypub_actors" is processed, and would
	// have announced, before "blockers" fails.
	blockers := core.NewBaseCollection("blockers")
	blockers.Fields.Add(
		&core.RelationField{Name: "owner", CollectionId: owners.Id, MaxSelect: 1, Required: true, CascadeDelete: false},
	)
	if err := app.Save(blockers); err != nil {
		t.Fatal(err)
	}

	app.OnRecordDelete("activitypub_actors").BindFunc(CollectActorDeleteRecipientsHandler())
	app.OnRecordAfterDeleteSuccess("activitypub_actors").BindFunc(AnnounceActorDeleteHandler())

	return app
}

// newDepartingActor creates an owner, a local actor able to sign (with a real
// encrypted key, since PostActivity signs before it sends), and one remote
// follower whose inbox is the counting server.
func newDepartingActor(t *testing.T, app *pbtests.TestApp, inbox *inboxCounter) (owner, actor *core.Record) {
	t.Helper()

	ownersCollection, err := app.FindCollectionByNameOrId("owners")
	if err != nil {
		t.Fatal(err)
	}
	owner = core.NewRecord(ownersCollection)
	if err := app.Save(owner); err != nil {
		t.Fatal(err)
	}

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	encryptedKey, err := security.Encrypt(x509.MarshalPKCS1PrivateKey(key), testEncryptionKey)
	if err != nil {
		t.Fatal(err)
	}

	actorsCollection, err := app.FindCollectionByNameOrId("activitypub_actors")
	if err != nil {
		t.Fatal(err)
	}

	iri := "https://local.example/api/v1/activitypub/user/alice"
	actor = core.NewRecord(actorsCollection)
	actor.Set("iri", iri)
	actor.Set("inbox", iri+"/inbox")
	actor.Set("private_key", encryptedKey)
	actor.Set("is_local", true)
	actor.Set("owner", owner.Id)
	if err := app.Save(actor); err != nil {
		t.Fatal(err)
	}

	follower := core.NewRecord(actorsCollection)
	follower.Set("iri", "https://remote.example/api/v1/activitypub/user/bob")
	follower.Set("inbox", inbox.url())
	follower.Set("is_local", false)
	if err := app.Save(follower); err != nil {
		t.Fatal(err)
	}

	followsCollection, err := app.FindCollectionByNameOrId("follows")
	if err != nil {
		t.Fatal(err)
	}
	follow := core.NewRecord(followsCollection)
	follow.Set("follower", follower.Id)
	follow.Set("followee", actor.Id)
	follow.Set("status", "accepted")
	if err := app.Save(follow); err != nil {
		t.Fatal(err)
	}

	return owner, actor
}

func TestActorDeleteAnnouncementWaitsForCommit(t *testing.T) {
	t.Run("RolledBackDeletionReachesNoInbox", func(t *testing.T) {
		app := setupActorDeleteHooksTestApp(t)
		inbox := newInboxCounter(t)

		owner, actor := newDepartingActor(t, app, inbox)

		blockers, err := app.FindCollectionByNameOrId("blockers")
		if err != nil {
			t.Fatal(err)
		}
		blocker := core.NewRecord(blockers)
		blocker.Set("owner", owner.Id)
		if err := app.Save(blocker); err != nil {
			t.Fatal(err)
		}

		if err := app.Delete(owner); err == nil {
			t.Fatal("expected the delete to be blocked by the sibling's required reference, got nil")
		}

		if _, err := app.FindRecordById("activitypub_actors", actor.Id); err != nil {
			t.Fatalf("expected the actor to survive the rolled-back delete: %v", err)
		}

		if got := inbox.waitForHits(1, 500*time.Millisecond); got != 0 {
			t.Fatalf("a rolled-back deletion must reach no inbox, but the follower received %d delivery(ies)", got)
		}
	})

	t.Run("CommittedDeletionReachesFollower", func(t *testing.T) {
		app := setupActorDeleteHooksTestApp(t)
		inbox := newInboxCounter(t)

		owner, actor := newDepartingActor(t, app, inbox)
		iri := actor.GetString("iri")

		if err := app.Delete(owner); err != nil {
			t.Fatalf("expected the delete to succeed, got %v", err)
		}

		if _, err := app.FindRecordById("activitypub_actors", actor.Id); err == nil {
			t.Fatal("expected the actor to be gone")
		}

		if got := inbox.waitForHits(1, 5*time.Second); got != 1 {
			t.Fatalf("expected exactly one delivery to the follower, got %d", got)
		}

		announcement, err := app.FindFirstRecordByData("activitypub_activities", "object", iri)
		if err != nil {
			t.Fatalf("expected a persisted announcement naming the deleted actor: %v", err)
		}
		if got := announcement.GetString("type"); got != "Delete" {
			t.Fatalf("announcement type = %q, want Delete", got)
		}
		if got := announcement.GetString("actor"); got != iri {
			t.Fatalf("announcement signed as %q, want the deleted actor %q", got, iri)
		}
	})
}
