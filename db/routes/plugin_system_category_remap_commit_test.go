package routes

import (
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"github.com/pocketbase/pocketbase/core"
	"pocketbase/pluginsystem"
)

func pluginCategoryRemapCommitBody(t *testing.T, instanceID, categoryID string) string {
	t.Helper()
	payload, err := json.Marshal(map[string]any{"instanceId": instanceID, "config": map[string]any{"host": map[string]any{"categoryMapping": map[string]any{"Ride": categoryID}}}})
	if err != nil {
		t.Fatal(err)
	}
	return string(payload)
}

func TestCategoryRemapSQLCommitFailureDoesNotReportSuccess(t *testing.T) {
	f, trail, category := newPluginCategoryRemapFixture(t)
	// The INSERT is valid until COMMIT: a real deferred foreign-key failure
	// proves that a completed Save is insufficient evidence of a committed
	// remap. The production PocketBase SQLite connection enables foreign keys.
	for _, statement := range []string{
		`CREATE TABLE remap_commit_parent (id TEXT PRIMARY KEY)`,
		`CREATE TABLE remap_commit_gate (parent TEXT REFERENCES remap_commit_parent(id) DEFERRABLE INITIALLY DEFERRED)`,
		`CREATE TRIGGER remap_commit_gate_trigger AFTER UPDATE ON trails BEGIN INSERT INTO remap_commit_gate(parent) VALUES ('route-commit-private-secret'); END`,
	} {
		if _, err := f.app.DB().NewQuery(statement).Execute(); err != nil {
			t.Fatal(err)
		}
	}
	saveSucceeded := false
	afterCommitCalls := 0
	f.app.OnRecordUpdate("trails").BindFunc(func(e *core.RecordEvent) error {
		err := e.Next()
		saveSucceeded = err == nil
		return err
	})
	f.app.OnRecordAfterUpdateSuccess("trails").BindFunc(func(e *core.RecordEvent) error {
		afterCommitCalls++
		return e.Next()
	})
	enablePluginRequestLogs(t, f.app)
	const path = "/api/test/remap/apply"
	status, body := pluginHostRouteRequest(t, f.app, f.owner, http.MethodPost, path, pluginCategoryRemapCommitBody(t, f.instance.Id, category.Id), PluginSystemCategoryRemapApply)
	if status != 400 {
		t.Fatalf("failed SQL commit returned %d, want safe 400: %#v", status, body)
	}
	if !saveSucceeded || afterCommitCalls != 0 {
		t.Errorf("expected successful UPDATE and rejected COMMIT, got save=%v afterCommitCalls=%d", saveSucceeded, afterCommitCalls)
	}
	stored, err := f.app.FindRecordById("trails", trail.Id)
	if err != nil {
		t.Fatal(err)
	}
	if stored.GetString("category") != "" {
		t.Errorf("SQL commit failure did not roll back category: %q", stored.GetString("category"))
	}
	logs := assertPluginHostRouteLogs(t, f.app, http.MethodPost, path, status, body, "category_remap_apply", "database_save", "route-commit-private-secret")
	assertPluginCategoryRemapDiagnosticRefs(t, logs, f.owner.Id, f.instance.Id)
}

func TestCategoryRemapPostCommitHookFailureReturnsCompleteResult(t *testing.T) {
	f, first, category := newPluginCategoryRemapFixture(t)
	trails, err := f.app.FindCollectionByNameOrId("trails")
	if err != nil {
		t.Fatal(err)
	}
	second := core.NewRecord(trails)
	if err := f.app.Save(second); err != nil {
		t.Fatal(err)
	}
	refs, err := f.app.FindCollectionByNameOrId("trail_external_reference")
	if err != nil {
		t.Fatal(err)
	}
	ref := core.NewRecord(refs)
	ref.Set("user", f.owner.Id)
	ref.Set("plugin_id", f.manifest.ID)
	ref.Set("trail", second.Id)
	ref.Set("external_id", "second-provider-secret")
	ref.Set("provider_category", "Ride")
	if err := f.app.Save(ref); err != nil {
		t.Fatal(err)
	}
	hookCalls := 0
	f.app.OnRecordAfterUpdateSuccess("trails").BindFunc(func(e *core.RecordEvent) error {
		hookCalls++
		// A hook may already carry a different safe cause. The confirmed
		// post-commit failure must still be classified as post_commit_hook.
		return pluginsystem.WithDiagnosticCause(errors.New(pluginHostRoutePrivateError), "database_lookup")
	})
	enablePluginRequestLogs(t, f.app)
	const path = "/api/test/remap/apply"
	status, body := pluginHostRouteRequest(t, f.app, f.owner, http.MethodPost, path, pluginCategoryRemapCommitBody(t, f.instance.Id, category.Id), PluginSystemCategoryRemapApply)
	if status != 200 || body["count"] != float64(2) || body["remapped"] != float64(2) {
		t.Fatalf("confirmed commit returned incomplete result: status=%d body=%#v", status, body)
	}
	if hookCalls != 2 {
		t.Errorf("got %d actual post-commit hook calls, want 2", hookCalls)
	}
	for _, id := range []string{first.Id, second.Id} {
		stored, err := f.app.FindRecordById("trails", id)
		if err != nil {
			t.Fatal(err)
		}
		if stored.GetString("category") != category.Id {
			t.Errorf("successful remap did not commit trail %s: %q", id, stored.GetString("category"))
		}
	}
	logs := assertPluginHostRouteLogs(t, f.app, http.MethodPost, path, status, body, "category_remap_apply", "post_commit_hook", "second-provider-secret")
	assertPluginCategoryRemapDiagnosticRefs(t, logs, f.owner.Id, f.instance.Id)
}
