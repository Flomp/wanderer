package main

import "github.com/open-wanderer/wanderer/plugins/sdk"

const (
	immichThumbnailSizeThumbnail = "thumbnail"
	immichThumbnailSizePreview   = "preview"
)

// immichThumbnailSize picks the Immich thumbnail for the host's size class:
// the small thumbnail (250 px by default) for tiles, otherwise the preview
// (1440 px by default), which hosts without size classes also get.
func immichThumbnailSize(size string) string {
	if size == sdk.AssetThumbnailSizeSmall {
		return immichThumbnailSizeThumbnail
	}
	return immichThumbnailSizePreview
}
