package util

import (
	"context"
	"testing"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
	pbtests "github.com/pocketbase/pocketbase/tests"
)

func TestReconcileFederatedPhotoAssetsIsIdempotent(t *testing.T) {
	app := setupFederatedAssetsTestApp(t)
	defer app.Cleanup()

	asset := createFederatedTestAsset(t, app, "actor1", "https://origin/api/v1/assets/a1")
	createFederatedTestLink(t, app, "summit_log_assets", "summit_log", "sl1", asset.Id)

	photos := []FederatedPhoto{{CanonicalID: "https://origin/api/v1/assets/a1", FileURL: "https://origin/api/v1/assets/a1/file"}}
	for range 3 {
		if err := ReconcileFederatedPhotoAssets(app, context.Background(), "summit_log", "sl1", "actor1", photos); err != nil {
			t.Fatal(err)
		}
	}

	assertFederatedAssetCount(t, app, 1)
	assertFederatedLinkCount(t, app, "summit_log_assets", "summit_log", "sl1", 1)
}

func TestReconcileFederatedPhotoAssetsLinksExistingAssetWithoutFetch(t *testing.T) {
	app := setupFederatedAssetsTestApp(t)
	defer app.Cleanup()

	// same photo already materialized for this author on another target
	asset := createFederatedTestAsset(t, app, "actor1", "https://origin/api/v1/assets/a1")
	createFederatedTestLink(t, app, "trail_assets", "trail", "trail1", asset.Id)

	// unreachable FileURL: reaching the fetch would drop the photo silently
	photos := []FederatedPhoto{{CanonicalID: "https://origin/api/v1/assets/a1", FileURL: "https://invalid.invalid/file"}}
	if err := ReconcileFederatedPhotoAssets(app, context.Background(), "summit_log", "sl1", "actor1", photos); err != nil {
		t.Fatal(err)
	}

	assertFederatedAssetCount(t, app, 1)
	assertFederatedLinkCount(t, app, "summit_log_assets", "summit_log", "sl1", 1)
	assertFederatedLinkCount(t, app, "trail_assets", "trail", "trail1", 1)
}

func TestReconcileFederatedPhotoAssetsRemovesUnlistedPhotos(t *testing.T) {
	app := setupFederatedAssetsTestApp(t)
	defer app.Cleanup()

	kept := createFederatedTestAsset(t, app, "actor1", "https://origin/api/v1/assets/a1")
	removed := createFederatedTestAsset(t, app, "actor1", "https://origin/api/v1/assets/a2")
	shared := createFederatedTestAsset(t, app, "actor1", "https://origin/api/v1/assets/a3")
	createFederatedTestLink(t, app, "summit_log_assets", "summit_log", "sl1", kept.Id)
	createFederatedTestLink(t, app, "summit_log_assets", "summit_log", "sl1", removed.Id)
	createFederatedTestLink(t, app, "summit_log_assets", "summit_log", "sl1", shared.Id)
	createFederatedTestLink(t, app, "trail_assets", "trail", "trail1", shared.Id)

	photos := []FederatedPhoto{{CanonicalID: "https://origin/api/v1/assets/a1", FileURL: "https://origin/api/v1/assets/a1/file"}}
	if err := ReconcileFederatedPhotoAssets(app, context.Background(), "summit_log", "sl1", "actor1", photos); err != nil {
		t.Fatal(err)
	}

	assertFederatedLinkCount(t, app, "summit_log_assets", "summit_log", "sl1", 1)
	if _, err := app.FindRecordById("assets", removed.Id); err == nil {
		t.Fatal("orphaned asset was not deleted")
	}
	// still linked elsewhere -> unlinked here but kept
	if _, err := app.FindRecordById("assets", shared.Id); err != nil {
		t.Fatal("asset linked to another target was deleted")
	}
	assertFederatedLinkCount(t, app, "trail_assets", "trail", "trail1", 1)
}

func TestReconcileFederatedPhotoAssetsKeepsLegacyPhotosUntouched(t *testing.T) {
	app := setupFederatedAssetsTestApp(t)
	defer app.Cleanup()

	// photo materialized before assets carried a canonical identity
	legacy := createFederatedTestAsset(t, app, "actor1", "")
	createFederatedTestLink(t, app, "summit_log_assets", "summit_log", "sl1", legacy.Id)

	photos := []FederatedPhoto{{CanonicalID: "https://origin/api/v1/assets/a1", FileURL: "https://origin/api/v1/assets/a1/file"}}
	if err := ReconcileFederatedPhotoAssets(app, context.Background(), "summit_log", "sl1", "actor1", photos); err != nil {
		t.Fatal(err)
	}

	assertFederatedAssetCount(t, app, 1)
	assertFederatedLinkCount(t, app, "summit_log_assets", "summit_log", "sl1", 1)
	if _, err := app.FindRecordById("assets", legacy.Id); err != nil {
		t.Fatal("legacy asset was removed")
	}
}

func TestDeleteAssetIfOrphanedByAuthorKeepsForeignOrphan(t *testing.T) {
	app := setupFederatedAssetsTestApp(t)
	defer app.Cleanup()

	asset := createFederatedTestAsset(t, app, "actor2", "")
	deleted, err := DeleteAssetIfOrphanedByAuthor(app, asset.Id, "actor1")
	if err != nil {
		t.Fatal(err)
	}
	if deleted {
		t.Fatal("foreign orphaned asset was deleted")
	}
	if _, err := app.FindRecordById("assets", asset.Id); err != nil {
		t.Fatal("foreign orphaned asset was removed")
	}
}

func TestDeleteAssetIfOrphanedByAuthorDeletesOwnedOrphan(t *testing.T) {
	app := setupFederatedAssetsTestApp(t)
	defer app.Cleanup()

	asset := createFederatedTestAsset(t, app, "actor1", "")
	deleted, err := DeleteAssetIfOrphanedByAuthor(app, asset.Id, "actor1")
	if err != nil {
		t.Fatal(err)
	}
	if !deleted {
		t.Fatal("owned orphaned asset was not deleted")
	}
	if _, err := app.FindRecordById("assets", asset.Id); err == nil {
		t.Fatal("owned orphaned asset still exists")
	}
}

// A trail's thumbnail is its marked photo, else its first one. The route
// preview the editor generates stands in only for a trail without a real
// photo, as it did when it was stored as an ordinary trail photo.
func TestTrailThumbnailURLsPrefersRealPhotosOverGeneratedPreview(t *testing.T) {
	type photo struct {
		file     string
		metadata map[string]any
		marked   bool
	}
	preview := map[string]any{"generated": map[string]any{"kind": "route-preview"}}
	migratedPreview := map[string]any{"source_file": "route_abcdef1234.webp"}
	for _, tc := range []struct {
		name   string
		photos []photo
		want   int
	}{
		{
			name:   "marked photo",
			photos: []photo{{file: "first.jpg"}, {file: "marked.jpg", marked: true}},
			want:   1,
		},
		{
			name:   "photo over marked preview",
			photos: []photo{{file: "route.webp", metadata: migratedPreview, marked: true}, {file: "photo.jpg"}},
			want:   1,
		},
		{
			name:   "preview without photo",
			photos: []photo{{file: "preview.webp", metadata: preview}},
			want:   0,
		},
		{
			name:   "marked preview",
			photos: []photo{{file: "old.webp", metadata: preview}, {file: "new.webp", metadata: preview, marked: true}},
			want:   1,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			app := setupFederatedAssetsTestApp(t)
			defer app.Cleanup()

			assetIDs := make([]string, len(tc.photos))
			for i, p := range tc.photos {
				asset := createFederatedTestAsset(t, app, "actor1", "")
				asset.Set("file", p.file)
				asset.Set("metadata", p.metadata)
				if err := app.Save(asset); err != nil {
					t.Fatal(err)
				}
				link := createFederatedTestLink(t, app, "trail_assets", "trail", "trail1", asset.Id)
				link.Set("is_thumbnail", p.marked)
				if err := app.Save(link); err != nil {
					t.Fatal(err)
				}
				assetIDs[i] = asset.Id
			}
			trail := core.NewRecord(core.NewBaseCollection("trails"))
			trail.Id = "trail1"

			urls, err := TrailThumbnailURLs(app, []*core.Record{trail}, "")
			if err != nil {
				t.Fatal(err)
			}
			collection, err := app.FindCollectionByNameOrId("assets")
			if err != nil {
				t.Fatal(err)
			}
			want := "/api/v1/files/" + collection.Id + "/" + assetIDs[tc.want] + "/" + tc.photos[tc.want].file
			if urls["trail1"] != want {
				t.Fatalf("thumbnail URL = %q, want %q", urls["trail1"], want)
			}
		})
	}
}

func TestReconcileFederatedPhotoAssetsSetsTrailThumbnail(t *testing.T) {
	app := setupFederatedAssetsTestApp(t)
	defer app.Cleanup()

	first := createFederatedTestAsset(t, app, "actor1", "https://origin/api/v1/assets/a1")
	marked := createFederatedTestAsset(t, app, "actor1", "https://origin/api/v1/assets/a2")

	photos := []FederatedPhoto{
		{CanonicalID: "https://origin/api/v1/assets/a1", FileURL: "https://invalid.invalid/a1"},
		{CanonicalID: "https://origin/api/v1/assets/a2", FileURL: "https://invalid.invalid/a2", IsThumbnail: true},
	}
	if err := ReconcileFederatedPhotoAssets(app, context.Background(), "trail", "trail1", "actor1", photos); err != nil {
		t.Fatal(err)
	}

	assertFederatedLinkCount(t, app, "trail_assets", "trail", "trail1", 2)
	if first.Id == marked.Id {
		t.Fatal("test assets unexpectedly share an id")
	}
	for _, asset := range []*core.Record{first, marked} {
		link, err := app.FindFirstRecordByFilter("trail_assets", "trail={:trail} && asset={:asset}", dbx.Params{"trail": "trail1", "asset": asset.Id})
		if err != nil {
			t.Fatal(err)
		}
		if got, want := link.GetBool("is_thumbnail"), asset.Id == marked.Id; got != want {
			t.Fatalf("asset %s is_thumbnail = %t, want %t", asset.Id, got, want)
		}
	}
}

func setupFederatedAssetsTestApp(t *testing.T) *pbtests.TestApp {
	t.Helper()

	app, err := pbtests.NewTestApp(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}

	assets := core.NewBaseCollection("assets")
	assets.Fields.Add(
		&core.TextField{Name: "type"},
		&core.TextField{Name: "storage_mode"},
		&core.TextField{Name: "remote_status"},
		&core.TextField{Name: "author"},
		&core.TextField{Name: "file"},
		&core.TextField{Name: "external_provider"},
		&core.TextField{Name: "external_id"},
		&core.DateField{Name: "taken_at"},
		&core.NumberField{Name: "lat"},
		&core.NumberField{Name: "lon"},
		&core.JSONField{Name: "metadata"},
		&core.AutodateField{Name: "created", OnCreate: true},
		&core.AutodateField{Name: "updated", OnCreate: true, OnUpdate: true},
	)
	if err := app.Save(assets); err != nil {
		app.Cleanup()
		t.Fatal(err)
	}

	for _, config := range []struct {
		name  string
		field string
	}{
		{name: "trail_assets", field: "trail"},
		{name: "waypoint_assets", field: "waypoint"},
		{name: "summit_log_assets", field: "summit_log"},
	} {
		collection := core.NewBaseCollection(config.name)
		collection.Fields.Add(
			&core.TextField{Name: "asset"},
			&core.TextField{Name: config.field},
		)
		if config.name == "trail_assets" {
			collection.Fields.Add(&core.BoolField{Name: "is_thumbnail"})
		}
		collection.Fields.Add(
			&core.AutodateField{Name: "created", OnCreate: true},
			&core.AutodateField{Name: "updated", OnCreate: true, OnUpdate: true},
		)
		if err := app.Save(collection); err != nil {
			app.Cleanup()
			t.Fatal(err)
		}
	}

	return app
}

func createFederatedTestAsset(t *testing.T, app core.App, author string, externalID string) *core.Record {
	t.Helper()

	collection, err := app.FindCollectionByNameOrId("assets")
	if err != nil {
		t.Fatal(err)
	}
	record := core.NewRecord(collection)
	record.Set("type", "photo")
	record.Set("storage_mode", "copy")
	record.Set("author", author)
	if externalID != "" {
		record.Set("external_provider", FederationAssetProvider)
		record.Set("external_id", externalID)
	}
	if err := app.Save(record); err != nil {
		t.Fatal(err)
	}
	return record
}

func createFederatedTestLink(t *testing.T, app core.App, collectionName string, targetField string, targetID string, assetID string) *core.Record {
	t.Helper()

	record, err := EnsureAssetLink(app, collectionName, targetField, targetID, assetID)
	if err != nil {
		t.Fatal(err)
	}
	return record
}

func assertFederatedAssetCount(t *testing.T, app core.App, want int) {
	t.Helper()

	records, err := app.FindRecordsByFilter("assets", "type='photo'", "", -1, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != want {
		t.Fatalf("got %d assets, want %d", len(records), want)
	}
}

func assertFederatedLinkCount(t *testing.T, app core.App, collectionName string, targetField string, targetID string, want int) {
	t.Helper()

	links, err := app.FindRecordsByFilter(collectionName, targetField+"={:target}", "", -1, 0, dbx.Params{"target": targetID})
	if err != nil {
		t.Fatal(err)
	}
	if len(links) != want {
		t.Fatalf("got %d links for %s, want %d", len(links), targetID, want)
	}
}
