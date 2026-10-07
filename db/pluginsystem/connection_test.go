package pluginsystem

import (
	"errors"
	"testing"

	"github.com/pocketbase/pocketbase/core"
)

func connectionTestRecord(t *testing.T) (core.App, *core.Record) {
	t.Helper()
	app := core.NewBaseApp(core.BaseAppConfig{DataDir: t.TempDir()})
	if err := app.Bootstrap(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = app.ResetBootstrapState() })
	col := core.NewBaseCollection("plugin_instances")
	for _, name := range []string{"user", "plugin_id", "auth_context", "connection_generation", "oauth_token_revision", "status"} {
		col.Fields.Add(&core.TextField{Name: name})
	}
	col.Fields.Add(&core.BoolField{Name: "enabled"}, &core.JSONField{Name: "auth"}, &core.JSONField{Name: "config"}, &core.JSONField{Name: "state"})
	if err := app.Save(col); err != nil {
		t.Fatal(err)
	}
	r := core.NewRecord(col)
	r.Set("user", "owner")
	r.Set("plugin_id", "test")
	r.Set("auth_context", "api")
	r.Set("connection_generation", NewConnectionRevision())
	r.Set("oauth_token_revision", NewConnectionRevision())
	r.Set("enabled", true)
	if err := app.Save(r); err != nil {
		t.Fatal(err)
	}
	return app, r
}

func TestConnectionSnapshotAndFreshPatch(t *testing.T) {
	app, r := connectionTestRecord(t)
	snapshot := CaptureConnection(r)
	auth := map[string]any{"oauthFlowContext": "next", "accessToken": "fresh-token"}
	r.Set("auth", auth)
	r.Set("oauth_token_revision", NewConnectionRevision())
	r.Set("state", map[string]any{"cursor": "new"})
	if err := app.Save(r); err != nil {
		t.Fatal(err)
	}
	saved, err := PatchConnection(app, snapshot, true, func(fresh *core.Record) error { fresh.Set("status", "configured"); return nil })
	if err != nil {
		t.Fatal(err)
	}
	if JSONMapFromRecord(saved, "auth")["accessToken"] != "fresh-token" || JSONMapFromRecord(saved, "state")["cursor"] != "new" {
		t.Fatal("fresh patch overwrote concurrent auth/state")
	}
	if !snapshot.Matches(saved, true) {
		t.Fatal("flow and token revisions changed connection identity")
	}
	saved.Set("enabled", false)
	if err := app.Save(saved); err != nil {
		t.Fatal(err)
	}
	if _, err := CheckConnection(app, snapshot, true); !errors.Is(err, ErrInstanceDisabled) || !errors.Is(err, ErrConnectionObsolete) {
		t.Fatalf("disabled guard: %v", err)
	}
	if _, err := CheckConnection(app, snapshot, false); err != nil {
		t.Fatal(err)
	}
	saved.Set("connection_generation", NewConnectionRevision())
	if err := app.Save(saved); err != nil {
		t.Fatal(err)
	}
	if _, err := PatchConnection(app, snapshot, false, func(*core.Record) error { t.Fatal("obsolete patch executed"); return nil }); !errors.Is(err, ErrConnectionConflict) {
		t.Fatalf("changed generation: %v", err)
	}
	if err := app.Delete(saved); err != nil {
		t.Fatal(err)
	}
	if _, err := CheckConnection(app, snapshot, false); !errors.Is(err, ErrConnectionObsolete) {
		t.Fatalf("deleted instance: %v", err)
	}
}

func TestConnectionRequiresInitializedRevisions(t *testing.T) {
	_, r := connectionTestRecord(t)
	c := CaptureConnection(r)
	r.Set("oauth_token_revision", "")
	if c.Matches(r, false) {
		t.Fatal("empty token revision accepted")
	}
	r.Set("oauth_token_revision", "token")
	r.Set("connection_generation", "")
	if CaptureConnection(r).Matches(r, false) {
		t.Fatal("empty generation accepted")
	}
}

func TestInstanceTryLockHasNoWaitersAndReleaseIsIdempotent(t *testing.T) {
	release, ok := TryLockInstance("connection-lock-test")
	if !ok || !InstanceBusy("connection-lock-test") {
		t.Fatal("lock not held")
	}
	if _, ok := TryLockInstance("connection-lock-test"); ok {
		t.Fatal("busy lock acquired twice")
	}
	release()
	release()
	if InstanceBusy("connection-lock-test") {
		t.Fatal("release retained busy")
	}
	next, ok := TryLockInstance("connection-lock-test")
	if !ok {
		t.Fatal("lock not reusable")
	}
	next()
}

func TestConnectionSnapshotPinsKnownProviderAccount(t *testing.T) {
	_, record := connectionTestRecord(t)
	unknown := CaptureConnection(record)
	record.Set("auth", map[string]any{"providerUserId": "first-account"})
	if !unknown.Matches(record, true) {
		t.Fatal("first identity learned on same token chain invalidated unknown snapshot")
	}
	pinned := CaptureConnection(record)
	record.Set("auth", map[string]any{"providerUserId": "other-account"})
	if pinned.Matches(record, true) {
		t.Fatal("known provider account was rebound")
	}
	record.Set("auth", map[string]any{})
	if pinned.Matches(record, true) {
		t.Fatal("known provider account silently disappeared")
	}
}
