package util

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"path"

	"github.com/meilisearch/meilisearch-go"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
	"github.com/tkrajina/gpxgo/gpx"
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

	category := ""
	trailCategory := r.ExpandedOne("category")
	if trailCategory != nil {
		category = trailCategory.GetString("name")
	}

	polyline, err := getPolyline(app, r)
	if err != nil {
		polyline = ""
	}

	domain := ""
	if !author.GetBool("isLocal") {
		domain = author.GetString("domain")
	}

	logCount, err := app.CountRecords("summit_logs", dbx.NewExp("trail={:id}", dbx.Params{"id": r.Id}))
	if err != nil {
		return nil, err
	}

	document := map[string]any{
		"id":             r.Id,
		"author":         author.Id,
		"author_name":    author.GetString("preferred_username"),
		"author_avatar":  author.GetString("icon"),
		"name":           r.GetString("name"),
		"description":    r.GetString("description"),
		"location":       r.GetString("location"),
		"distance":       r.GetFloat("distance"),
		"elevation_gain": r.GetFloat("elevation_gain"),
		"elevation_loss": r.GetFloat("elevation_loss"),
		"duration":       r.GetFloat("duration"),
		"difficulty":     r.Get("difficulty"),
		"category":       category,
		"completed":      logCount > 0,
		"date":           r.GetDateTime("date").Time().Unix(),
		"created":        r.GetDateTime("created").Time().Unix(),
		"public":         r.GetBool("public"),
		"thumbnail":      thumbnail,
		"gpx":            r.GetString("gpx"),
		"tags":           tags,
		"polyline":       polyline,
		"domain":         domain,
		"iri":            r.GetString("iri"),
		"_geo": map[string]float64{
			"lat": r.GetFloat("lat"),
			"lng": r.GetFloat("lon"),
		},
	}

	if includeShares {
		document["shares"] = []string{}
		document["likes"] = []string{}
		document["like_count"] = 0

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
	gpxData, err := gpx.Parse(content)
	if err != nil {
		return "", err
	}

	gpxData.SimplifyTracks(50)
	coordinates := make([][]float64, 4)
	for _, trk := range gpxData.Tracks {
		for _, seg := range trk.Segments {
			for _, pt := range seg.Points {
				coordinates = append(coordinates, []float64{pt.Latitude, pt.Longitude})
			}
		}
	}
	return string(polyline.EncodeCoords(coordinates)), nil
}

func documentFromListRecord(r *core.Record, author *core.Record, includeShares bool) (map[string]interface{}, error) {

	totalElevationGain := 0.0
	totalElevationLoss := 0.0
	totalDistance := 0.0
	totalDuration := 0.0
	trails := len(r.GetStringSlice("trails"))

	if r.GetString("iri") != "" {
		doc, err := documentFromRemoteRecord(r, "lists")
		if err == nil {
			totalElevationGain = doc["elevation_gain"].(float64)
			totalElevationLoss = doc["elevation_loss"].(float64)
			totalDistance = doc["distance"].(float64)
			totalDuration = doc["duration"].(float64)

			trails = int(doc["trails"].(float64))
		}

	} else {
		allTrails := r.ExpandedAll("trails")

		for _, t := range allTrails {
			totalElevationGain += t.GetFloat("elevation_gain")
			totalElevationLoss += t.GetFloat("elevation_loss")
			totalDistance += t.GetFloat("distance")
			totalDuration += t.GetFloat("duration")

		}
	}

	domain := ""
	if !author.GetBool("isLocal") {
		domain = author.GetString("domain")
	}

	document := map[string]interface{}{
		"id":             r.Id,
		"author":         author.Id,
		"author_name":    author.GetString("preferred_username"),
		"author_avatar":  author.GetString("icon"),
		"avatar":         r.GetString("avatar"),
		"name":           r.GetString("name"),
		"description":    r.GetString("description"),
		"elevation_gain": totalElevationGain,
		"elevation_loss": totalElevationLoss,
		"distance":       totalDistance,
		"duration":       totalDuration,
		"domain":         domain,
		"public":         r.GetBool("public"),
		"created":        r.GetDateTime("created").Time().Unix(),
		"trails":         trails,
		"iri":            r.GetString("iri"),
	}

	if includeShares {
		document["shares"] = []string{}
	}

	return document, nil
}

func documentFromRemoteRecord(r *core.Record, index string) (map[string]interface{}, error) {
	client := &http.Client{}

	if r.GetString("iri") == "" {
		return nil, fmt.Errorf("record has no iri")
	}

	iri := r.GetString("iri")

	url, err := url.Parse(iri)
	if err != nil {
		return nil, err
	}

	remoteRecordId := path.Base(url.Path)

	searchURL := fmt.Sprintf("%s://%s/api/v1/search/%s", url.Scheme, url.Host, index)
	body := []byte(fmt.Sprintf(`{"q": "%s"}`, remoteRecordId))

	req, err := http.NewRequest("POST", searchURL, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch remote record: received status %d", resp.StatusCode)
	}

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var searchResponse meilisearch.SearchResponse
	json.Unmarshal(respBytes, &searchResponse)

	if len(searchResponse.Hits) == 0 {
		return nil, fmt.Errorf("no documents in result set")
	}

	document, ok := searchResponse.Hits[0].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected document format")
	}
	return document, nil
}

func IndexTrail(app core.App, r *core.Record, author *core.Record, client meilisearch.ServiceManager) error {
	errs := app.ExpandRecord(r, []string{"tags"}, nil)
	if len(errs) > 0 {
		return fmt.Errorf("failed to expand tags: %v", errs)
	}
	errs = app.ExpandRecord(r, []string{"category"}, nil)
	if len(errs) > 0 {
		return fmt.Errorf("failed to expand category: %v", errs)
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
		return fmt.Errorf("failed to expand tags: %v", errs)
	}
	errs = app.ExpandRecord(r, []string{"category"}, nil)
	if len(errs) > 0 {
		return fmt.Errorf("failed to expand category: %v", errs)
	}

	doc, err := documentFromTrailRecord(app, r, author, false)
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

func UpdateTrailLikes(trailId string, likes []string, client meilisearch.ServiceManager) error {
	documents := []map[string]interface{}{
		{
			"id":         trailId,
			"like_count": len(likes),
			"likes":      likes,
		},
	}
	if _, err := client.Index("trails").UpdateDocuments(documents); err != nil {
		return err
	}
	return nil
}

func IndexList(app core.App, r *core.Record, author *core.Record, client meilisearch.ServiceManager) error {
	errs := app.ExpandRecord(r, []string{"trails"}, nil)
	if len(errs) > 0 {
		return fmt.Errorf("failed to expand trails: %v", errs)
	}

	documents, err := documentFromListRecord(r, author, true)
	if err != nil {
		return err
	}
	if _, err = client.Index("lists").AddDocuments(documents); err != nil {
		return err
	}

	return nil
}

func UpdateList(app core.App, r *core.Record, author *core.Record, client meilisearch.ServiceManager) error {
	errs := app.ExpandRecord(r, []string{"trails"}, nil)
	if len(errs) > 0 {
		return fmt.Errorf("failed to expand trails: %v", errs)
	}

	documents, err := documentFromListRecord(r, author, false)
	if err != nil {
		return err
	}

	if _, err = client.Index("lists").UpdateDocuments(documents); err != nil {
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
