package main

import (
	"encoding/json"
	"testing"
)

func TestThumbnailRequestPicksImmichSize(t *testing.T) {
	for request, want := range map[string]string{
		`{"action":"thumbnail","assetIds":["a"],"thumbnailSize":"small"}`: "thumbnail",
		`{"action":"thumbnail","assetIds":["a"],"thumbnailSize":"large"}`: "preview",
		// Hosts without size classes keep getting the preview.
		`{"action":"thumbnail","assetIds":["a"]}`: "preview",
	} {
		var decoded assetLibraryRequest
		if err := json.Unmarshal([]byte(request), &decoded); err != nil {
			t.Fatal(err)
		}
		if got := immichThumbnailSize(decoded.ThumbnailSize); got != want {
			t.Errorf("%s: Immich size = %q, want %q", request, got, want)
		}
	}
}
