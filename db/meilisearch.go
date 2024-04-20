package main

import (
	"errors"
	"log"

	"github.com/meilisearch/meilisearch-go"
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
			"difficulty":     r.Get("difficulty"),
			"category":       r.Get("category"),
			"completed":      len(r.GetStringSlice("summit_logs")) > 0,
			"date":           r.GetDateTime("date").Time().Unix(),
			"created":        r.GetDateTime("created").Time().Unix(),
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

func generateMeilisearchToken(rules map[string]interface{}, client *meilisearch.Client) (resp string, err error) {
	apiKeyUid := ""
	apiKey := ""

	if keys, err := client.GetKeys(nil); err != nil {
		log.Fatal(err)
	} else {
		for _, k := range keys.Results {
			if k.Name == "Default Search API Key" {
				apiKeyUid = k.UID
				apiKey = k.Key
			}
		}
	}

	if len(apiKey) == 0 || len(apiKeyUid) == 0 {
		return "", errors.New("unable to locate meilisearch API key")
	}

	options := &meilisearch.TenantTokenOptions{
		APIKey: apiKey,
	}

	return client.GenerateTenantToken(apiKeyUid, rules, options)
}
