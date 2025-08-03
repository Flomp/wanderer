package komoot

import (
	"bytes"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/filesystem"
	"github.com/pocketbase/pocketbase/tools/security"
	"github.com/twpayne/go-gpx"
)

func SyncKomoot(app core.App) error {
	integrations, err := app.FindAllRecords("integrations", dbx.NewExp("true"))
	if err != nil {
		return err
	}

	for _, i := range integrations {
		encryptionKey := os.Getenv("POCKETBASE_ENCRYPTION_KEY")
		if len(encryptionKey) == 0 {
			return errors.New("POCKETBASE_ENCRYPTION_KEY not set")
		}

		userId := i.GetString("user")
		actor, err := app.FindFirstRecordByData("activitypub_actors", "user", userId)
		if err != nil {
			warning := fmt.Sprintf("no actor found for user: %s\n", userId)
			fmt.Print(warning)
			app.Logger().Warn(warning)
			continue
		}
		actorId := actor.Id

		komootString := i.GetString("komoot")
		komootIntegration := KomootIntegration{
			Planned:   true,
			Completed: true,
		}
		json.Unmarshal([]byte(komootString), &komootIntegration)

		if !komootIntegration.Active || komootIntegration.Email == "" || komootIntegration.Password == "" {
			continue
		}
		k := &KomootApi{}

		decryptedPassword, err := security.Decrypt(komootIntegration.Password, encryptionKey)
		if err != nil {
			warning := fmt.Sprintf("unable to decrypt password: %v\n", err)
			fmt.Print(warning)
			app.Logger().Warn(warning)
			continue
		}

		err = k.Login(komootIntegration.Email, string(decryptedPassword))
		if err != nil {
			warning := fmt.Sprintf("komoot login failed: %v\n", err)
			fmt.Print(warning)
			app.Logger().Warn(warning)
			continue
		}
		hasNewTours := true
		page := 0
		for hasNewTours {
			tours, err := k.fetchTours(page)
			if err != nil {
				warning := fmt.Sprintf("error fetching tours from komoot: %v\n", err)
				fmt.Print(warning)
				app.Logger().Warn(warning)
				continue
			}

			hasNewTours, err = syncTrailWithTours(app, k, komootIntegration, userId, actorId, tours)
			if err != nil {
				warning := fmt.Sprintf("error syncing komoot tours with trails: %v\n", err)
				fmt.Print(warning)
				app.Logger().Warn(warning)
				continue
			}
			page += 1
		}
	}

	return nil
}

type BasicAuthToken struct {
	Key   string
	Value string
}

func (b BasicAuthToken) Apply(req *http.Request) {
	authStr := "Basic " + base64.StdEncoding.EncodeToString([]byte(b.Key+":"+b.Value))
	req.Header.Set("Authorization", authStr)
}

type KomootApi struct {
	UserID string
	Token  string
}

func (k *KomootApi) buildHeader() *BasicAuthToken {
	if k.UserID != "" && k.Token != "" {
		return &BasicAuthToken{k.UserID, k.Token}
	}
	return nil
}

func sendRequest(url string, auth *BasicAuthToken) ([]byte, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	if auth != nil {
		auth.Apply(req)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("error sending request to komoot (%d): %s", resp.StatusCode, string(body))
	}

	return io.ReadAll(resp.Body)
}

func (k *KomootApi) Login(email, password string) error {
	url := fmt.Sprintf("https://api.komoot.de/v006/account/email/%s/", email)

	body, err := sendRequest(url, &BasicAuthToken{email, password})
	if err != nil {
		return err
	}

	var data LoginResponse
	json.Unmarshal(body, &data)

	k.UserID = data.Username
	k.Token = data.Password

	return nil
}
func (k *KomootApi) fetchTours(page int) ([]KomootTour, error) {
	currentUri := fmt.Sprintf("https://api.komoot.de/v007/users/%s/tours/?page=%d&sort_field=date&sort_direction=desc&limit=30", k.UserID, page)

	body, err := sendRequest(currentUri, k.buildHeader())
	if err != nil {
		return nil, err
	}

	var data KomootToursResponse
	json.Unmarshal(body, &data)

	tours := data.Embedded.Tours

	return tours, nil
}

func (k *KomootApi) fetchDetailedTour(tour KomootTour) (*DetailedKomootTour, error) {
	url := fmt.Sprintf("https://api.komoot.de/v007/tours/%d?_embedded=coordinates,way_types,surfaces,directions,participants,timeline,cover_images&directions=v2&fields=timeline&format=coordinate_array&timeline_highlights_fields=tips,recommenders&page=2", tour.ID)
	body, err := sendRequest(url, k.buildHeader())
	if err != nil {
		return nil, err
	}

	var data *DetailedKomootTour
	json.Unmarshal(body, &data)
	return data, nil
}

func syncTrailWithTours(app core.App, k *KomootApi, i KomootIntegration, user string, actor string, tours []KomootTour) (bool, error) {
	hasNewTours := false
	for _, tour := range tours {
		trails, err := app.FindRecordsByFilter("trails", "external_id = {:id}", "", 1, 0, dbx.Params{"id": strconv.Itoa(int(tour.ID))})
		if err != nil {
			return hasNewTours, err
		}
		if len(trails) != 0 || (tour.Type == "tour_planned" && !i.Planned) || (tour.Type == "tour_recorded" && !i.Completed) {
			continue
		}
		hasNewTours = true
		detailedTour, err := k.fetchDetailedTour(tour)
		if err != nil {
			app.Logger().Warn(fmt.Sprintf("Unable to fetch details for tour '%s': %v", tour.Name, err))
			continue
		}
		gpx, err := generateTourGPX(detailedTour)
		if err != nil {
			app.Logger().Warn(fmt.Sprintf("Unable to generate GPX for tour '%s': %v", tour.Name, err))
			continue
		}
		wpIds, err := createWaypointsFromTour(app, detailedTour, user)
		if err != nil {
			app.Logger().Warn(fmt.Sprintf("Unable to create waypoints for tour '%s': %v", tour.Name, err))
			continue
		}
		err = createTrailFromTour(app, k, detailedTour, gpx, actor, wpIds)
		if err != nil {
			app.Logger().Warn(fmt.Sprintf("Unable to create trail for tour '%s': %v", tour.Name, err))
			continue
		}

	}
	return hasNewTours, nil
}

// mapCategory returns the category ID for the given Komoot category.
// A two-step lookup will be done. First a lookup for a exact category will be performed. If the
// user has manually defined the categories of the exactMap, these will be used.
// If there is no result, a mapping to the default Wanderer categories is done.
func mapCategory(app core.App, komootCategory string) string {
	// List of sports retrieved from here: https://static.komoot.de/doc/external-api/v007/sports.html
	exactMap := map[string]string{
		// Main sport types
		"hike":             "Hiking",
		"mountaineering":   "Hiking",
		"racebike":         "Racebiking",
		"e_racebike":       "Racebiking",
		"touringbicycle":   "Biking",
		"e_touringbicycle": "Biking",
		"mtb":              "Mountainbiking",
		"e_mtb":            "Mountainbiking",
		"mtb_easy":         "Gravelbiking", // Gravel riding
		"e_mtb_easy":       "Gravelbiking", // E-Gravel riding
		"mtb_advanced":     "Endurobiking", // Enduro mountain biking
		"e_mtb_advanced":   "Endurobiking", // E-enduro mountain biking
		"jogging":          "Running",
		// Tracking sport types
		"climbing":      "Climbing",
		"downhillbike":  "Downhillbiking",
		"nordic":        "Skiing", // Cross-Country Skiing
		"nordicwalking": "Nordicwalking",
		"skaten":        "Skating",
		"skialpin":      "Skiing",     // Alpine Skiing,
		"skitour":       "Skitouring", // Alpine Ski Touring
		"sled":          "Sledding",   // Sledding
		"snowboard":     "Snowboarding",
		"snowshoe":      "Snowshoeing", // Snow Shoe
		"unicycle":      "Unicycling",  // Unicycling
		"citybike":      "Biking",
		"other":         "Hiking",
	}

	defaultMap := map[string]string{
		// Main sport types
		"hike":             "Hiking",
		"mountaineering":   "Hiking",
		"racebike":         "Biking",
		"e_racebike":       "Biking",
		"touringbicycle":   "Biking",
		"e_touringbicycle": "Biking",
		"mtb":              "Biking",
		"e_mtb":            "Biking",
		"mtb_easy":         "Biking", // Gravel riding
		"e_mtb_easy":       "Biking", // E-Gravel riding
		"mtb_advanced":     "Biking", // Enduro mountain biking
		"e_mtb_advanced":   "Biking", // E-enduro mountain biking
		"jogging":          "Walking",
		// Tracking sport types
		"climbing":      "Climbing",
		"downhillbike":  "Biking",
		"nordic":        "Skiing", // Cross-Country Skiing
		"nordicwalking": "Walking",
		"skaten":        "Walking",
		"skialpin":      "Skiing", // Alpine Skiing,
		"skitour":       "Skiing", // Alpine Ski Touring
		"sled":          "Skiing", // Sledding
		"snowboard":     "Skiing",
		"snowshoe":      "Skiing", // Snow Shoe
		"unicycle":      "Biking", // Unicycling
		"citybike":      "Biking",
		"other":         "Walking",
	}

	// Check if there is a user defined exact category mapping found
	category, err := app.FindFirstRecordByData("categories", "name", exactMap[komootCategory])
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			app.Logger().Warn("Komoot import: error getting exact category mapping", "error", err)
		}
	} else if category != nil {
		return category.Id
	}

	// Map the Komoot categories to the default Wanderer categories
	category, err = app.FindFirstRecordByData("categories", "name", defaultMap[komootCategory])
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			app.Logger().Warn("Komoot import: error getting default category mapping", "error", err)
		}
	} else if category != nil {
		return category.Id
	}

	// If no mapping found
	if errors.Is(err, sql.ErrNoRows) {
		app.Logger().Warn("Komoot import: category mapping is missing. Probably Komoot introduced a new category.", "Komoot category", komootCategory)
	}
	return ""
}

func createTrailFromTour(app core.App, k *KomootApi, detailedTour *DetailedKomootTour, gpx *filesystem.File, actor string, wpIds []string) error {
	trailid := security.RandomStringWithAlphabet(core.DefaultIdLength, core.DefaultIdAlphabet)

	collection, err := app.FindCollectionByNameOrId("trails")
	if err != nil {
		return err
	}

	record := core.NewRecord(collection)

	categoryId := mapCategory(app, detailedTour.Sport)

	var photos []*filesystem.File
	if len(detailedTour.Embedded.CoverImages.Embedded.Items) > 0 {
		photos, err = fetchRoutePhotos(k, detailedTour)
		if err != nil {
			return err
		}
	} else {
		photo, err := fetchPhoto(detailedTour.MapImage.Src, "", "")
		if err != nil {
			return err
		}
		photos = append(photos, photo)
	}

	diffculty := detailedTour.Difficulty.Grade
	if diffculty == "" {
		diffculty = "easy"
	}

	record.Load(map[string]any{
		"id":                trailid,
		"name":              detailedTour.Name,
		"public":            detailedTour.Status == "public",
		"distance":          detailedTour.Distance,
		"elevation_gain":    detailedTour.ElevationUp,
		"elevation_loss":    detailedTour.ElevationDown,
		"duration":          detailedTour.Duration,
		"date":              detailedTour.Date,
		"external_provider": "komoot",
		"external_id":       strconv.Itoa(detailedTour.ID),
		"lat":               detailedTour.StartPoint.Lat,
		"lon":               detailedTour.StartPoint.Lng,
		"difficulty":        diffculty,
		"category":          categoryId,
		"waypoints":         wpIds,
		"author":            actor,
	})

	if photos != nil {
		record.Set("photos", photos)
	}
	if gpx != nil {
		record.Set("gpx", gpx)
	}

	if err := app.Save(record); err != nil {
		return err
	}

	if detailedTour.Type == "tour_recorded" {
		collection, err := app.FindCollectionByNameOrId("summit_logs")
		if err != nil {
			return err
		}

		summitLogRecord := core.NewRecord(collection)
		summitLogRecord.Load(map[string]any{
			"distance":       detailedTour.Distance,
			"elevation_gain": detailedTour.ElevationUp,
			"elevation_loss": detailedTour.ElevationDown,
			"duration":       detailedTour.Duration,
			"date":           detailedTour.Date,
			"author":         actor,
			"trail":          trailid,
		})
		if err := app.Save(summitLogRecord); err != nil {
			return err
		}
	}

	return nil
}

func createWaypointsFromTour(app core.App, tour *DetailedKomootTour, user string) ([]string, error) {
	collection, err := app.FindCollectionByNameOrId("waypoints")
	if err != nil {
		return nil, err
	}

	wpIds := make([]string, len(tour.Embedded.Timeline.Embedded.Items))

	for i, wp := range tour.Embedded.Timeline.Embedded.Items {
		photos, err := fetchWaypointPhotos(wp)
		if err != nil {
			return nil, err
		}
		record := core.NewRecord(collection)

		wpDescription := ""
		if len(wp.Embedded.Reference.Embedded.Tips.Embedded.Items) > 0 {
			wpDescription = wp.Embedded.Reference.Embedded.Tips.Embedded.Items[0].Text
		}

		wpLat := wp.Embedded.Reference.StartPoint.Lat
		if wpLat == 0 {
			wpLat = tour.StartPoint.Lat
		}

		wpLon := wp.Embedded.Reference.StartPoint.Lng
		if wpLon == 0 {
			wpLon = tour.StartPoint.Lng
		}

		record.Load(map[string]any{
			"name":                wp.Embedded.Reference.Name,
			"description":         wpDescription,
			"lat":                 wpLat,
			"lon":                 wpLon,
			"icon":                "circle",
			"author":              user,
			"distance_from_start": 0,
		})

		if photos != nil {
			record.Set("photos", photos)
		}

		if err := app.Save(record); err != nil {
			return nil, err
		}

		wpIds[i] = record.Id
	}

	return wpIds, nil
}

func fetchRoutePhotos(k *KomootApi, tour *DetailedKomootTour) ([]*filesystem.File, error) {
	url := fmt.Sprintf("https://api.komoot.de/v007/tours/%d/cover_images/", tour.ID)
	body, err := sendRequest(url, k.buildHeader())
	if err != nil {
		return nil, err
	}

	var data *CoverImages
	err = json.Unmarshal(body, &data)
	if err != nil {
		return nil, err
	}

	photos := make([]*filesystem.File, data.Page.TotalElements)

	for i, img := range data.Embedded.Items {
		photo, err := fetchPhoto(img.Src, "", "")
		if err != nil {
			return nil, err
		}
		if strings.HasSuffix(photo.Name, ".gif") {
			continue
		}
		photos[i] = photo

		//TODO: komoot photos can have location data. Maybe we should create a waypoint for those photos?
	}

	return photos, nil
}

func fetchWaypointPhotos(wp Item) ([]*filesystem.File, error) {

	photos := make([]*filesystem.File, len(wp.Embedded.Reference.Embedded.Images.Embedded.Items))

	for i, img := range wp.Embedded.Reference.Embedded.Images.Embedded.Items {
		photo, err := fetchPhoto(img.Src, "", "")
		if err != nil {
			return nil, err
		}
		if strings.HasSuffix(photo.Name, ".gif") {
			continue
		}
		photos[i] = photo
	}

	return photos, nil
}

func fetchPhoto(url string, width string, height string) (*filesystem.File, error) {
	url = strings.Replace(url, "{crop}", "false", 1)
	url = strings.Replace(url, "{width}", width, 1)
	url = strings.Replace(url, "{height}", height, 1)

	bytes, err := sendRequest(url, nil)
	if err != nil {
		return nil, err
	}

	return filesystem.NewFileFromBytes(bytes, "photo")
}

func generateTourGPX(detailedTour *DetailedKomootTour) (*filesystem.File, error) {
	var points []*gpx.WptType

	for _, item := range detailedTour.Embedded.Coordinates.Items {
		t := detailedTour.Date.Unix() + int64(item.T/1000)

		points = append(points, &gpx.WptType{Lat: item.Lat, Lon: item.Lng, Ele: item.Alt, Time: time.Unix(t, 0)})
	}

	gpx := &gpx.GPX{
		Version: "1.1",
		Creator: "komoot GPX Exporter",
		Trk: []*gpx.TrkType{
			{
				Name: detailedTour.Name,
				TrkSeg: []*gpx.TrkSegType{
					{
						TrkPt: points,
					},
				},
			},
		},
	}
	var buf bytes.Buffer
	err := gpx.Write(&buf)
	if err != nil {
		return nil, err
	}

	gpxFile, err := filesystem.NewFileFromBytes(buf.Bytes(), detailedTour.Name+".gpx")
	if err != nil {
		return nil, err
	}

	return gpxFile, nil
}
