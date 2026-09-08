package hooks

import (
	"pocketbase/federation"

	"github.com/pocketbase/pocketbase/core"
)

// Search projection runs on successful record mutations, including internal
// writes. The request hook handles only the explicit federation announcement.
func CreateListShareHandler() func(e *core.RecordRequestEvent) error {
	return func(e *core.RecordRequestEvent) error {
		if err := e.Next(); err != nil {
			return err
		}
		return federation.CreateAnnounceActivity(e.App, e.Record, federation.ListAnnounceType)
	}
}
