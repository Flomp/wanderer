package routes

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"pocketbase/pluginsystem"

	"github.com/pocketbase/pocketbase/core"
	pbtests "github.com/pocketbase/pocketbase/tests"
	"github.com/pocketbase/pocketbase/tools/types"
)

func TestPluginAssetThumbnailSizeClassifiesLongestEdge(t *testing.T) {
	for thumb, want := range map[string]string{
		"100x100":  pluginAssetThumbnailSizeSmall,
		"300x300":  pluginAssetThumbnailSizeSmall,
		"0x300":    pluginAssetThumbnailSizeSmall,
		"300x300f": pluginAssetThumbnailSizeSmall,
		"301x100":  pluginAssetThumbnailSizeLarge,
		"600x0":    pluginAssetThumbnailSizeLarge,
		"0x0":      pluginAssetThumbnailSizeLarge,
		"small":    pluginAssetThumbnailSizeLarge,
		"":         pluginAssetThumbnailSizeLarge,
	} {
		if got := pluginAssetThumbnailSize(thumb); got != want {
			t.Errorf("pluginAssetThumbnailSize(%q) = %q, want %q", thumb, got, want)
		}
	}
}

// A tile must not get a cached preview, nor a preview a cached tile, from the
// provider endpoint or from a linked photo.
func TestPluginAssetThumbnailsServeTheRequestedSizeClass(t *testing.T) {
	app, instance, asset := newPluginAssetThumbnailTestApp(t)
	keys := []string{}
	for _, size := range []string{pluginAssetThumbnailSizeSmall, pluginAssetThumbnailSizeLarge} {
		key := pluginAssetThumbnailCacheKey("user", "photos", instance.Id, "photo-1", size)
		keys = append(keys, key)
		putPluginAssetThumbnailCache(key, pluginAssetThumbnailCacheEntry{
			ContentType: "image/webp",
			Body:        []byte(size),
			ETag:        pluginAssetThumbnailETag([]byte(size)),
			ExpiresAt:   time.Now().Add(time.Hour),
			LastAccess:  time.Now(),
		})
	}
	t.Cleanup(func() {
		pluginAssetThumbnailCache.Lock()
		defer pluginAssetThumbnailCache.Unlock()
		for _, key := range keys {
			pluginAssetThumbnailCache.totalBytes -= int64(len(pluginAssetThumbnailCache.items[key].Body))
			delete(pluginAssetThumbnailCache.items, key)
		}
	})

	auth := core.NewRecord(core.NewAuthCollection("test_users"))
	auth.Id = "user"
	for _, test := range []struct {
		target  string
		handler func(*core.RequestEvent) error
		want    string
	}{
		{"/plugins/assets/photos/thumbnail/photo-1?thumb=300x300", PluginSystemAssetThumbnail, pluginAssetThumbnailSizeSmall},
		{"/plugins/assets/photos/thumbnail/photo-1", PluginSystemAssetThumbnail, pluginAssetThumbnailSizeLarge},
		{"/assets/" + asset.Id + "/file?thumb=100x100", AssetFile, pluginAssetThumbnailSizeSmall},
		{"/assets/" + asset.Id + "/file?thumb=600x0", AssetFile, pluginAssetThumbnailSizeLarge},
	} {
		recorder := httptest.NewRecorder()
		e := &core.RequestEvent{App: app, Auth: auth}
		e.Request = httptest.NewRequest(http.MethodGet, test.target, nil)
		e.Request.SetPathValue("plugin", "photos")
		e.Request.SetPathValue("asset", "photo-1")
		e.Request.SetPathValue("id", asset.Id)
		e.Response = recorder
		if err := test.handler(e); err != nil {
			t.Fatalf("%s: %v", test.target, err)
		}
		if recorder.Code != http.StatusOK || recorder.Body.String() != test.want {
			t.Errorf("%s: HTTP %d %q, want the %s thumbnail", test.target, recorder.Code, recorder.Body.String(), test.want)
		}
	}
}

func newPluginAssetThumbnailTestApp(t *testing.T) (*pbtests.TestApp, *core.Record, *core.Record) {
	t.Helper()
	app, err := pbtests.NewTestApp(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(app.Cleanup)
	collections := map[string][]core.Field{
		"activitypub_actors": {&core.TextField{Name: "user"}},
		"plugin_instances":   {&core.TextField{Name: "user"}, &core.TextField{Name: "plugin_id"}, &core.BoolField{Name: "enabled"}, &core.JSONField{Name: "config"}, &core.JSONField{Name: "auth"}},
		"installed_plugins":  {&core.TextField{Name: "plugin_id"}, &core.TextField{Name: "path"}, &core.JSONField{Name: "manifest"}, &core.JSONField{Name: "config"}},
		"assets":             {&core.TextField{Name: "author"}, &core.TextField{Name: "storage_mode"}, &core.TextField{Name: "external_provider"}, &core.TextField{Name: "external_id"}, &core.JSONField{Name: "metadata"}, &core.FileField{Name: "file", MaxSelect: 1}},
	}
	for name, fields := range collections {
		collection := core.NewBaseCollection(name)
		collection.Fields.Add(fields...)
		if name == "assets" {
			collection.ViewRule = types.Pointer("")
		}
		if err := app.Save(collection); err != nil {
			t.Fatal(err)
		}
	}
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "plugin.wasm"), nil, 0o600); err != nil {
		t.Fatal(err)
	}
	save := func(collectionName string, values map[string]any) *core.Record {
		t.Helper()
		collection, err := app.FindCollectionByNameOrId(collectionName)
		if err != nil {
			t.Fatal(err)
		}
		record := core.NewRecord(collection)
		record.Load(values)
		if err := app.Save(record); err != nil {
			t.Fatalf("save %s: %v", collectionName, err)
		}
		return record
	}
	save("installed_plugins", map[string]any{"plugin_id": "photos", "path": dir, "manifest": pluginsystem.Manifest{
		ManifestVersion: pluginsystem.ManifestVersion, ID: "photos", Type: pluginsystem.PluginTypeAssets, Name: "Photos", Version: "1.0.0",
		Runtime:      pluginsystem.RuntimeManifest{Type: pluginsystem.RuntimeWASM, Entrypoint: "plugin.wasm"},
		Capabilities: []pluginsystem.CapabilityManifest{{Name: "asset_library", Version: "v1", Export: "asset_library_v1"}},
	}})
	instance := save("plugin_instances", map[string]any{"user": "user", "plugin_id": "photos", "enabled": true})
	actor := save("activitypub_actors", map[string]any{"user": "user"})
	asset := save("assets", map[string]any{
		"author": actor.Id, "storage_mode": "link_private", "external_provider": "photos", "external_id": "photo-1",
		"metadata": map[string]any{"remote": pluginsystem.RemotePhotoAsset{PluginID: "photos", Source: pluginsystem.MediaSource{
			Type: "connector", MediaRef: &pluginsystem.MediaRef{Connector: "api", Path: "/photos/photo-1"},
		}}},
	})
	return app, instance, asset
}
