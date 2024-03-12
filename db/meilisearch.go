package main

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"os"

	"github.com/meilisearch/meilisearch-go"
	"github.com/pocketbase/pocketbase/models"
)

func hasIndex(index string, client *meilisearch.Client) (resp bool, err error) {

	if _, err := client.GetIndex(index); err != nil {
		if err.(*meilisearch.Error).StatusCode == 404 {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func bootstrapMeilisearch(client *meilisearch.Client) error {
	indexExists, err := hasIndex("cities500", client)
	if err != nil {
		return err
	}
	if !indexExists {

		jsonFile, err := os.Open("migrations/initial_data/cities500.json")

		if err != nil {
			return err
		}
		defer jsonFile.Close()

		byteValue, err := io.ReadAll(jsonFile)

		if err != nil {
			return err
		}
		var cities []map[string]interface{}
		json.Unmarshal(byteValue, &cities)

		_, err = client.CreateIndex(&meilisearch.IndexConfig{
			Uid:        "cities500",
			PrimaryKey: "id",
		})
		if err != nil {
			return err
		}
		settings := meilisearch.Settings{
			SortableAttributes: []string{
				"_geo",
			},
			FilterableAttributes: []string{
				"_geo",
			},
		}
		_, err = client.Index("cities500").UpdateSettings(&settings)
		if err != nil {
			return err
		}
		_, err = client.Index("cities500").AddDocuments(cities)
		if err != nil {
			return (err)
		}
	}

	indexExists, err = hasIndex("trails", client)
	if err != nil {
		return err
	}
	if !indexExists {
		_, err = client.CreateIndex(&meilisearch.IndexConfig{
			Uid:        "trails",
			PrimaryKey: "id",
		})
		if err != nil {
			return err
		}
		settings := meilisearch.Settings{
			SortableAttributes: []string{
				"name", "distance", "elevation_gain", "difficulty", "created",
			},
			FilterableAttributes: []string{
				"category", "difficulty", "distance", "elevation_gain", "completed", "_geo", "public", "author",
			},
		}
		_, err = client.Index("trails").UpdateSettings(&settings)
		if err != nil {
			return err
		}
	}
	return nil
}

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
