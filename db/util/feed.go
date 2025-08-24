package util

import "github.com/pocketbase/pocketbase/core"

type FeedType string

const (
	ListFeed      FeedType = "list"
	TrailFeed     FeedType = "trail"
	SummitLogFeed FeedType = "summit_log"
)

func InsertIntoFeed(app core.App, actorId string, authorId string, itemId string, feedType FeedType) (*core.Record, error) {
	collection, err := app.FindCollectionByNameOrId("feed")
	if err != nil {
		return nil, err
	}

	record := core.NewRecord(collection)

	record.Set("actor", actorId)
	record.Set("author", authorId)
	record.Set("item", itemId)
	record.Set("type", string(feedType))

	return record, app.Save(record)
}

func DeleteFromFeed(app core.App, itemId string) error {

	record, err := app.FindFirstRecordByData("feed", "item", itemId)
	if err != nil {
		return err
	}

	return app.Delete(record)
}
