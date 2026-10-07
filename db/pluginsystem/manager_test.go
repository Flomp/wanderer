package pluginsystem

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/pocketbase/pocketbase/core"
)

func TestPluginCacheDiagnosticIdentityCannotClaimHealthyManifestID(t *testing.T) {
	app := core.NewBaseApp(core.BaseAppConfig{DataDir: t.TempDir()})
	if err := app.Bootstrap(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = app.ResetBootstrapState() })
	collection := core.NewBaseCollection("installed_plugins")
	for _, name := range []string{"plugin_id", "name", "type", "version", "runtime", "path", "status", "error"} {
		collection.Fields.Add(&core.TextField{Name: name})
	}
	collection.Fields.Add(&core.JSONField{Name: "manifest"}, &core.JSONField{Name: "config"})
	collection.AddIndex("idx_diagnostic_identity", true, "plugin_id", "")
	if err := app.Save(collection); err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	bundle := func(folder, id string) {
		t.Helper()
		path := filepath.Join(dir, folder)
		if err := os.MkdirAll(path, 0700); err != nil {
			t.Fatal(err)
		}
		manifest := Manifest{ManifestVersion: "1.0", ID: id, Type: PluginTypeTrails, Name: id, Version: "1.0", Runtime: RuntimeManifest{Type: RuntimeWASM, Entrypoint: "plugin.wasm"}, Capabilities: []CapabilityManifest{{Name: "list", Version: "1.0", Export: "list"}}}
		data, err := json.Marshal(manifest)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(path, "plugin.json"), data, 0600); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(path, "plugin.wasm"), []byte("\x00asm\x01\x00\x00\x00"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	healthyID := collisionPluginIssueID(LocalPluginIssue{ID: "a", Dir: filepath.Join(dir, "a")})
	bundle("a", "first")
	bundle("z", healthyID)
	manager := NewManager(app, dir)
	if err := manager.SyncInstalledPlugins(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(dir, "a", "plugin.wasm")); err != nil {
		t.Fatal(err)
	}
	bundle("b", "first")
	for range 2 {
		if err := manager.SyncInstalledPlugins(context.Background()); err != nil {
			t.Fatal(err)
		}
		if count, err := app.CountRecords(collection.Id); err != nil || count != 3 {
			t.Fatalf("diagnostic collision lost cache entries: %d/%v", count, err)
		}
		for id, folder := range map[string]string{"first": "b", healthyID: "z"} {
			row, err := app.FindFirstRecordByData(collection.Id, "plugin_id", id)
			if err != nil || row.GetString("path") != filepath.Join(dir, folder) || row.GetString("status") != "available" {
				t.Fatalf("diagnostic identity claimed healthy manifest: %v/%v", row, err)
			}
		}
		broken, err := app.FindFirstRecordByData(collection.Id, "path", filepath.Join(dir, "a"))
		if err != nil || broken.GetString("plugin_id") != healthyID+"-1" || broken.GetString("status") != "error" {
			t.Fatalf("diagnostic identity not reserved deterministically: %v/%v", broken, err)
		}
	}
}

func TestPluginCacheIdentityTransitionsAreAtomicAndOrderIndependent(t *testing.T) {
	for _, targets := range [][]string{{"second", "third"}, {"second", "first"}} {
		t.Run(strings.Join(targets, "/"), func(t *testing.T) {
			app := core.NewBaseApp(core.BaseAppConfig{DataDir: t.TempDir()})
			if err := app.Bootstrap(); err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = app.ResetBootstrapState() })
			collection := core.NewBaseCollection("installed_plugins")
			collection.Fields.Add(&core.TextField{Name: "plugin_id", Required: true}, &core.TextField{Name: "name"}, &core.JSONField{Name: "config"})
			collection.AddIndex("idx_transition_plugin_id", true, "plugin_id", "")
			if err := app.Save(collection); err != nil {
				t.Fatal(err)
			}
			ids := []string{"aaaaaaaaaaaaaaa", "bbbbbbbbbbbbbbb"}
			originals := []string{"first", "second"}
			for i, id := range ids {
				record := core.NewRecord(collection)
				record.Id = id
				record.Set("plugin_id", originals[i])
				record.Set("name", "preserved-"+id)
				record.Set("config", map[string]any{"admin": id})
				if err := app.Save(record); err != nil {
					t.Fatal(err)
				}
			}
			if err := app.RunInTransaction(func(tx core.App) error {
				rows, err := tx.FindRecordsByFilter(collection.Id, "", "id", -1, 0)
				if err != nil {
					return err
				}
				plan := pluginCachePlan{}
				for i, row := range rows {
					plan = append(plan, &pluginCacheEntry{record: row, values: map[string]any{"plugin_id": targets[i]}})
				}
				return NewManager(tx, "").applyInstalledPluginCache(context.Background(), plan)
			}); err != nil {
				t.Fatal(err)
			}
			for i, id := range ids {
				record, err := app.FindRecordById(collection.Id, id)
				if err != nil {
					t.Fatal(err)
				}
				if record.GetString("plugin_id") != targets[i] || record.GetString("name") != "preserved-"+id || !reflect.DeepEqual(JSONMapFromRecord(record, "config"), map[string]any{"admin": id}) {
					t.Fatalf("identity transition changed fields or left a temporary ID: %#v", record.FieldsData())
				}
			}
			// Failure after the temporary writes must roll the whole transition back.
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			app.OnRecordUpdate("installed_plugins").BindFunc(func(e *core.RecordEvent) error {
				if strings.HasPrefix(e.Record.GetString("plugin_id"), "cache-transition-") {
					cancel()
				}
				return e.Next()
			})
			err := app.RunInTransaction(func(tx core.App) error {
				rows, err := tx.FindRecordsByFilter(collection.Id, "", "id", -1, 0)
				if err != nil {
					return err
				}
				plan := pluginCachePlan{}
				for i, row := range rows {
					plan = append(plan, &pluginCacheEntry{record: row, values: map[string]any{"plugin_id": originals[i]}})
				}
				return NewManager(tx, "").applyInstalledPluginCache(ctx, plan)
			})
			if err != context.Canceled {
				t.Fatalf("canceled transition error=%v", err)
			}
			for i, id := range ids {
				row, err := app.FindRecordById(collection.Id, id)
				if err != nil || row.GetString("plugin_id") != targets[i] {
					t.Fatalf("partial transition survived rollback: %v/%v", row, err)
				}
			}
		})
	}
}

func TestPluginIssueRecordID(t *testing.T) {
	valid := pluginIssueRecordID(LocalPluginIssue{ID: "komoot", Dir: "/plugins/komoot"})
	if valid != "komoot" {
		t.Fatalf("pluginIssueRecordID(valid) = %q, want komoot", valid)
	}

	first := pluginIssueRecordID(LocalPluginIssue{ID: "@@@", Dir: "/plugins/@@@"})
	second := pluginIssueRecordID(LocalPluginIssue{ID: "***", Dir: "/plugins/***"})
	if first == second {
		t.Fatalf("invalid plugin issue ids collided: %q", first)
	}
	for _, got := range []string{first, second} {
		if !pluginIDPattern.MatchString(got) {
			t.Fatalf("pluginIssueRecordID() = %q, not a valid plugin id", got)
		}
	}
}

func TestMergePluginIssuesPreservesSetupErrorCodes(t *testing.T) {
	infos := []PluginInfo{{
		ID:     "cached",
		Path:   "/plugins/cached",
		Status: "available",
	}}
	issues := []LocalPluginIssue{
		{
			ID:             "cached",
			Name:           "Cached",
			Dir:            "/plugins/cached",
			SetupErrorCode: SetupErrorCodeManifestInvalid,
			Error:          "private cached diagnostic",
		},
		{
			ID:             "new",
			Name:           "New",
			Dir:            "/plugins/new",
			SetupErrorCode: SetupErrorCodeRuntimeEntrypointMissing,
			Error:          "private new diagnostic",
		},
	}

	got := mergePluginIssues(infos, map[string]int{"/plugins/cached": 0}, issues)
	if len(got) != 2 {
		t.Fatalf("got %d plugin infos, want 2", len(got))
	}
	if got[0].Status != "error" ||
		got[0].SetupErrorCode != SetupErrorCodeManifestInvalid ||
		got[0].Error != "private cached diagnostic" {
		t.Fatalf("cached plugin issue was not overlaid: %#v", got[0])
	}
	if got[1].ID != "new" ||
		got[1].SetupErrorCode != SetupErrorCodeRuntimeEntrypointMissing ||
		got[1].Error != "private new diagnostic" {
		t.Fatalf("new plugin issue was not appended: %#v", got[1])
	}
}
