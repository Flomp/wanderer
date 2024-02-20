package main

import (
	"log"

	"github.com/meilisearch/meilisearch-go"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/models"
)

func indexRecord(r *models.Record, client *meilisearch.Client) error {
	documents := []map[string]interface{}{
		{
			"id":             r.Id,
			"author":         r.GetString("author"),
			"name":           r.GetString("name"),
			"description":    r.GetString("description"),
			"location":       r.GetString("location"),
			"distance":       r.GetFloat("distance"),
			"elevation_gain": r.GetFloat("elevation_gain"),
			"duration":       r.GetFloat("duration"),
			"category":       r.Get("category"),
			"completed":      len(r.GetStringSlice("summit_logs")) > 0,
			"created":        r.GetDateTime("created"),
			"public":         r.GetBool("public"),
			"_geo": map[string]float64{
				"lat": r.GetFloat("lat"),
				"lng": r.GetFloat("lon"),
			},
		},
	}

	if _, err := client.Index("trails").AddDocuments(documents); err != nil {
		return err
	}

	return nil
}

func main() {
	app := pocketbase.New()

	client := meilisearch.NewClient(meilisearch.ClientConfig{
		Host:   "http://localhost:7700",
		APIKey: "p2gYZAWODOrwTPr4AYoahCZ9CI8y9bUd0yQLGk-E3m8",
	})

	app.OnRecordAfterCreateRequest("users").Add(func(e *core.RecordCreateEvent) error {
		userId := e.Record.GetId()
		apiKeyUid := "f44e1493-9e9b-4013-ae0f-b6e2c366e93c"
		apiKey := "adce85854ce63f0e7507e40d2280d374d7202b41890b0734d70fe4c711c793f0"

		searchRules := map[string]interface{}{
			"trails": map[string]string{
				"filter": "public = true OR author = " + userId,
			},
		}
		options := &meilisearch.TenantTokenOptions{
			APIKey: apiKey,
		}

		if token, err := client.GenerateTenantToken(apiKeyUid, searchRules, options); err != nil {
			return err
		} else {
			e.Record.Set("token", token)

			if err := app.Dao().SaveRecord(e.Record); err != nil {
				return err
			}
		}

		return nil
	})

	app.OnRecordAfterCreateRequest("trails").Add(func(e *core.RecordCreateEvent) error {
		if err := indexRecord(e.Record, client); err != nil {
			return err
		}
		return nil
	})

	app.OnRecordAfterUpdateRequest("trails").Add(func(e *core.RecordUpdateEvent) error {
		if err := indexRecord(e.Record, client); err != nil {
			return err
		}
		return nil
	})

	app.OnRecordAfterDeleteRequest("trails").Add(func(e *core.RecordDeleteEvent) error {
		if _, err := client.Index("trails").DeleteDocument(e.Record.Id); err != nil {
			return err
		}
		return nil
	})

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
