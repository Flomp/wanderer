package util

import (
	"errors"
	"log"

	"github.com/meilisearch/meilisearch-go"
	"github.com/pocketbase/pocketbase/models"
)

func documentFromTrailRecord(r *models.Record, author *models.Record, includeShares bool) map[string]interface{} {
	photos := r.GetStringSlice("photos")
	thumbnail := ""
	if len(photos) > 0 {
		thumbnail = r.GetStringSlice("photos")[r.GetInt("thumbnail")]
	}

	document := map[string]interface{}{
		"id":             r.Id,
		"author":         r.GetString("author"),
		"author_name":    author.GetString("username"),
		"author_avatar":  author.GetString("avatar"),
		"name":           r.GetString("name"),
		"description":    r.GetString("description"),
		"location":       r.GetString("location"),
		"distance":       r.GetFloat("distance"),
		"elevation_gain": r.GetFloat("elevation_gain"),
		"elevation_loss": r.GetFloat("elevation_loss"),
		"duration":       r.GetFloat("duration"),
		"difficulty":     r.Get("difficulty"),
		"category":       r.Get("category"),
		"completed":      len(r.GetStringSlice("summit_logs")) > 0,
		"date":           r.GetDateTime("date").Time().Unix(),
		"created":        r.GetDateTime("created").Time().Unix(),
		"public":         r.GetBool("public"),
		"thumbnail":      thumbnail,
		"gpx":            r.GetString("gpx"),
		"_geo": map[string]float64{
			"lat": r.GetFloat("lat"),
			"lng": r.GetFloat("lon"),
		},
	}

	if includeShares {
		document["shares"] = []string{}
	}

	return document
}

func documentFromListRecord(r *models.Record, includeShares bool) map[string]interface{} {
	document := map[string]interface{}{
		"id":          r.Id,
		"author":      r.GetString("author"),
		"name":        r.GetString("name"),
		"description": r.GetString("description"),
		"public":      r.GetBool("public"),
		"created":     r.GetDateTime("created").Time().Unix(),
		"trails":      r.GetStringSlice("trails"),
	}

	if includeShares {
		document["shares"] = []string{}
	}

	return document
}

func IndexTrail(r *models.Record, author *models.Record, client meilisearch.ServiceManager) error {
	documents := []map[string]interface{}{documentFromTrailRecord(r, author, true)}

	if _, err := client.Index("trails").AddDocuments(documents); err != nil {
		return err
	}

	return nil
}

func UpdateTrail(r *models.Record, author *models.Record, client meilisearch.ServiceManager) error {
	documents := documentFromTrailRecord(r, author, false)

	if _, err := client.Index("trails").UpdateDocuments(documents); err != nil {
		return err
	}

	return nil
}

func UpdateTrailShares(trailId string, shares []string, client meilisearch.ServiceManager) error {
	documents := []map[string]interface{}{
		{
			"id":     trailId,
			"shares": shares,
		},
	}
	if _, err := client.Index("trails").UpdateDocuments(documents); err != nil {
		return err
	}
	return nil
}

func IndexList(r *models.Record, client meilisearch.ServiceManager) error {
	documents := []map[string]interface{}{documentFromListRecord(r, true)}

	if _, err := client.Index("lists").AddDocuments(documents); err != nil {
		return err
	}

	return nil
}

func UpdateList(r *models.Record, client meilisearch.ServiceManager) error {
	documents := documentFromListRecord(r, false)

	if _, err := client.Index("lists").UpdateDocuments(documents); err != nil {
		return err
	}

	return nil
}

func UpdateListShares(listId string, shares []string, client meilisearch.ServiceManager) error {
	documents := []map[string]interface{}{
		{
			"id":     listId,
			"shares": shares,
		},
	}
	if _, err := client.Index("lists").UpdateDocuments(documents); err != nil {
		return err
	}
	return nil
}

func GenerateMeilisearchToken(rules map[string]interface{}, client meilisearch.ServiceManager) (resp string, err error) {
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
