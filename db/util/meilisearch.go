package util

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"

	"github.com/meilisearch/meilisearch-go"
	"github.com/pocketbase/pocketbase/core"
	"github.com/twpayne/go-gpx"
	"github.com/twpayne/go-polyline"
)

func documentFromTrailRecord(app core.App, r *core.Record, author *core.Record, includeShares bool) (map[string]interface{}, error) {
	photos := r.GetStringSlice("photos")
	thumbnail := ""
	if len(photos) > 0 {
		thumbnailIndex := r.GetInt("thumbnail")
		if thumbnailIndex >= len(photos) {
			thumbnailIndex = 0
		}
		thumbnail = photos[thumbnailIndex]
	}

	tagRecords := r.ExpandedAll("tags")
	tags := make([]string, len(tagRecords))

	for i, v := range tagRecords {
		tags[i] = v.GetString("name")
	}

	polyline, err := getPolyline(app, r)
	if err != nil {
		return nil, err
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
		"tags":           tags,
		"polyline":       polyline,
		"_geo": map[string]float64{
			"lat": r.GetFloat("lat"),
			"lng": r.GetFloat("lon"),
		},
	}

	if includeShares {
		document["shares"] = []string{}
	}

	return document, nil
}

func getPolyline(app core.App, r *core.Record) (string, error) {
	gpxPath := r.GetString("gpx")
	if len(gpxPath) == 0 {
		return "", nil
	}
	avatarKey := r.BaseFilesPath() + "/" + gpxPath
	fsys, err := app.NewFilesystem()
	if err != nil {
		return "", err
	}
	defer fsys.Close()

	gpxFile, err := fsys.GetFile(avatarKey)
	if err != nil {
		return "", err
	}
	defer gpxFile.Close()

	content := new(bytes.Buffer)
	_, err = io.Copy(content, gpxFile)
	if err != nil {
		return "", err
	}
	gpxData, err := gpx.Read(content)
	if err != nil {
		return "", err
	}
	coordinates := make([][]float64, 4)
	for _, trk := range gpxData.Trk {
		for _, seg := range trk.TrkSeg {
			for _, pt := range seg.TrkPt {
				coordinates = append(coordinates, []float64{pt.Lat, pt.Lon})
			}
		}
	}
	return string(polyline.EncodeCoords(coordinates)), nil
}

func documentFromListRecord(r *core.Record, includeShares bool) map[string]interface{} {
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

func IndexTrail(app core.App, r *core.Record, author *core.Record, client meilisearch.ServiceManager) error {
	errs := app.ExpandRecord(r, []string{"tags"}, nil)
	if len(errs) > 0 {
		return fmt.Errorf("failed to expand: %v", errs)
	}
	doc, err := documentFromTrailRecord(app, r, author, true)
	if err != nil {
		return err
	}
	documents := []map[string]interface{}{doc}

	if _, err := client.Index("trails").AddDocuments(documents); err != nil {
		return err
	}

	return nil
}

func UpdateTrail(app core.App, r *core.Record, author *core.Record, client meilisearch.ServiceManager) error {
	errs := app.ExpandRecord(r, []string{"tags"}, nil)
	if len(errs) > 0 {
		return fmt.Errorf("failed to expand: %v", errs)
	}

	doc, err := documentFromTrailRecord(app, r, author, true)
	if err != nil {
		return err
	}
	documents := []map[string]interface{}{doc}

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

func IndexList(r *core.Record, client meilisearch.ServiceManager) error {
	documents := []map[string]interface{}{documentFromListRecord(r, true)}

	if _, err := client.Index("lists").AddDocuments(documents); err != nil {
		return err
	}

	return nil
}

func UpdateList(r *core.Record, client meilisearch.ServiceManager) error {
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
