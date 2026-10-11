package hooks

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"pocketbase/federation"
	"pocketbase/util"

	"github.com/meilisearch/meilisearch-go"
	"github.com/pocketbase/pocketbase/core"
)

func CreateSummitLogHandler(client meilisearch.ServiceManager, updates *FederationUpdates) func(e *core.RecordRequestEvent) error {
	return func(e *core.RecordRequestEvent) error {
		err := e.Next()
		if err != nil {
			return err
		}

		// add local iri
		origin := os.Getenv("ORIGIN")
		if origin == "" {
			return fmt.Errorf("ORIGIN not set")
		}
		if e.Record.GetString("iri") == "" {
			e.Record.Set("iri", fmt.Sprintf("%s/api/v1/summit-log/%s", origin, e.Record.Id))
		}
		err = e.App.UnsafeWithoutHooks().Save(e.Record)
		if err != nil {
			return err
		}

		trail, err := e.App.FindRecordById("trails", e.Record.GetString("trail"))
		if err != nil {
			return err
		}

		if err := util.IndexTrails(e.App, []*core.Record{trail}, client); err != nil {
			return err
		}

		// The photos are linked after the log is saved, so the Create waits
		// for them and carries them.
		updates.EnqueueCreate(e.Record)
		return nil
	}
}

func UpdateSummitLogHandler(updates *FederationUpdates) func(e *core.RecordRequestEvent) error {
	return func(e *core.RecordRequestEvent) error {

		err := e.Next()
		if err != nil {
			return err
		}

		// The photos are linked after the log is saved, so the Update waits
		// for them and goes out once for the whole save.
		updates.Enqueue(e.Record)
		return nil
	}
}

// DeleteSummitLogHandler runs on OnRecordAfterDeleteSuccess rather than on the
// delete request, so that summit logs removed by a cascade — when their trail is
// deleted, or when their author's account is — also retract their federated
// copies. Inside a transaction these hooks are deferred until after it commits,
// so the record is already gone by the time this runs and an error here cannot
// roll the deletion back.
func DeleteSummitLogHandler(client meilisearch.ServiceManager) func(e *core.RecordEvent) error {
	return func(e *core.RecordEvent) error {
		trail, err := e.App.FindRecordById("trails", e.Record.GetString("trail"))
		switch {
		case err == nil:
			if err := util.IndexTrails(e.App, []*core.Record{trail}, client); err != nil {
				return err
			}
		case errors.Is(err, sql.ErrNoRows):
			// The trail went first and took this log with it. Nothing is left
			// to reindex, but the log's own followers still have to be told;
			// CreateSummitLogDeleteActivity handles the missing trail.
		default:
			return err
		}

		err = federation.CreateSummitLogDeleteActivity(e.App, e.Record)
		if err != nil {
			return err
		}
		return e.Next()
	}
}
