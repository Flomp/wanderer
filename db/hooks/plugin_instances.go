package hooks

import (
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"pocketbase/pluginsystem"
)

// Direct collection writes are forbidden, including superuser and batch REST.
// Trusted internal model saves do not pass through request hooks.
func DenyPluginInstanceWriteHandler() func(*core.RecordRequestEvent) error {
	return func(e *core.RecordRequestEvent) error {
		return apis.NewForbiddenError("Use the plugin instance host routes to create or update an instance.", nil)
	}
}

func EnrichPluginInstanceHandler() func(*core.RecordEnrichEvent) error {
	return func(e *core.RecordEnrichEvent) error {
		if err := e.Next(); err != nil {
			return err
		}
		owner := e.RequestInfo != nil && e.RequestInfo.Auth != nil && e.RequestInfo.Auth.Id == e.Record.GetString("user")
		pluginsystem.ProjectConnectionRecord(e.App, e.Record, owner)
		return nil
	}
}
