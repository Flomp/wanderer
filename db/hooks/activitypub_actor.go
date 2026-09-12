package hooks

import (
	"log"
	"pocketbase/federation"
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

		return util.UpdateActor(e.Record, client)
	}
}

const actorDeleteRecipientsKey = "__delete_recipients"

func CollectActorDeleteRecipientsHandler() func(e *core.RecordEvent) error {
	return func(e *core.RecordEvent) error {
		actor := e.Record

		recipients, err := federation.ActorDeleteRecipients(e.App, actor)
		if err != nil {
			e.App.Logger().Error(
				"could not collect recipients to announce actor deletion to",
				"actor", actor.Id, "error", err,
			)
			recipients = nil
		}

		actor.Set(actorDeleteRecipientsKey, recipients)

		return e.Next()
	}
}

func AnnounceActorDeleteHandler() func(e *core.RecordEvent) error {
	return func(e *core.RecordEvent) error {
		actor := e.Record

		recipients, _ := actor.GetRaw(actorDeleteRecipientsKey).([]string)

		if err := federation.CreateActorDeleteActivity(e.App, actor, recipients); err != nil {
			e.App.Logger().Error(
				"could not announce actor deletion",
				"actor", actor.Id, "error", err,
			)
		}

		return e.Next()
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
