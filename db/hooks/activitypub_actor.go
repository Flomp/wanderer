package hooks

import (
	"log"
	"pocketbase/util"
	"time"

	"github.com/meilisearch/meilisearch-go"
	"github.com/pocketbase/pocketbase/core"
)

func CreateActorHandler(client meilisearch.ServiceManager) func(e *core.RecordEvent) error {
	return func(e *core.RecordEvent) error {
		err := e.Next()
		if err != nil {
			return err
		}

		return util.IndexActors([]*core.Record{e.Record}, client)
	}
}

func UpdateActorHandler(client meilisearch.ServiceManager) func(e *core.RecordEvent) error {
	return func(e *core.RecordEvent) error {
		err := e.Next()
		if err != nil {
			return err
		}

		if err := util.UpdateActor(e.Record, client); err != nil {
			return err
		}
		original := e.Record.Original()
		for _, field := range []string{"preferred_username", "icon", "domain", "is_local"} {
			if e.Record.GetString(field) != original.GetString(field) {
				return util.UpdateActorReferences(e.App, e.Record.Id, client)
			}
		}
		return nil
	}
}

func DeleteActorHandler(client meilisearch.ServiceManager) func(e *core.RecordEvent) error {
	return func(e *core.RecordEvent) error {
		task, err := client.Index("actors").DeleteDocument(e.Record.Id, nil)
		if err != nil {
			return err
		}

		interval := 500 * time.Millisecond
		_, err = client.WaitForTask(task.TaskUID, interval)
		if err != nil {
			log.Fatalf("Error waiting for task completion: %v", err)
		}
		return e.Next()
	}
}
