package strava

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/filesystem"
	"github.com/pocketbase/pocketbase/tools/security"
	"github.com/twpayne/go-gpx"
	"github.com/twpayne/go-polyline"
)

type StravaApi struct {
	AceessToken string
}

func SyncStrava(app core.App) error {
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
		stravaString := i.GetString("strava")
		var stravaIntegration StravaIntegration
		err := json.Unmarshal([]byte(stravaString), &stravaIntegration)
		if err != nil {
			return err
		}

		if !stravaIntegration.Active || stravaIntegration.RefreshToken == "" {
			continue
		}

		decryptedSecret, err := security.Decrypt(stravaIntegration.ClientSecret, encryptionKey)
		if err != nil {
			return err
		}

		decryptedRefreshToken, err := security.Decrypt(stravaIntegration.RefreshToken, encryptionKey)
		if err != nil {
			return err
		}

		request := RefreshTokenRequest{
			ClientID:     stravaIntegration.ClientID,
			ClientSecret: string(decryptedSecret),
			RefreshToken: string(decryptedRefreshToken),
			GrantType:    "refresh_token",
		}
		r, err := GetStravaToken(request)
		if err != nil {
			warning := fmt.Sprintf("error refreshing strava access token: %v\n", err)
			fmt.Print(warning)
			app.Logger().Warn(warning)
			continue
		}
		if r.AccessToken != "" {
			stravaIntegration.AccessToken = r.AccessToken
		}
		if r.RefreshToken != "" {
			stravaIntegration.RefreshToken = r.RefreshToken
		}
		if r.AccessToken != "" {
			stravaIntegration.ExpiresAt = r.ExpiresAt
		}

		b, err := json.Marshal(stravaIntegration)
		if err != nil {
			return err
		}
		i.Set("strava", string(b))
		err = app.Save(i)
		if err != nil {
			return err
		}

		if stravaIntegration.Routes {
			page := 1
			hasNewRoutes := true
			for hasNewRoutes {

				routes, err := fetchStravaRoutes(r.AccessToken, page)
				page += 1
				if err != nil {
					warning := fmt.Sprintf("error fetching routes from strava: %v\n", err)
					fmt.Print(warning)
					app.Logger().Warn(warning)
					continue
				}
				hasNewRoutes, err = syncTrailsWithRoutes(app, r.AccessToken, userId, routes)
				if err != nil {
					warning := fmt.Sprintf("error syncing strava routes with trails: %v\n", err)
					fmt.Print(warning)
					app.Logger().Warn(warning)
					continue
				}
			}
		}
		if stravaIntegration.Activities {
			page := 1
			hasNewActivities := true
			for hasNewActivities {
				activities, err := fetchStravaActivities(r.AccessToken, page)
				page += 1
				if err != nil {
					warning := fmt.Sprintf("error fetching activities from strava: %v", err)
					fmt.Print(warning)
					app.Logger().Warn(warning)
					continue
				}
				hasNewActivities, err = syncTrailsWithActivities(app, r.AccessToken, userId, activities)
				if err != nil {
					warning := fmt.Sprintf("error syncing strava activities with trails: %v", err)
					fmt.Print(warning)
					app.Logger().Warn(warning)
					continue
				}
			}
		}
	}

	return nil
}

func GetStravaToken(request any) (*RefreshTokenResponse, error) {
	const stravaTokenURL = "https://www.strava.com/oauth/token"

	requestBody, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", stravaTokenURL, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get token: received status %d", resp.StatusCode)
	}

	var tokenResponse RefreshTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResponse); err != nil {
		return nil, err
	}

	return &tokenResponse, nil
}

func fetchStravaRoutes(accessToken string, page int) ([]StravaRoute, error) {
	stravaRoutesURL := fmt.Sprintf("https://www.strava.com/api/v3/athlete/routes?page=%d", page)

	req, err := http.NewRequest("GET", stravaRoutesURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch routes: received status %d", resp.StatusCode)
	}

	var routes []StravaRoute
	if err := json.NewDecoder(resp.Body).Decode(&routes); err != nil {
		return nil, err
	}

	return routes, nil
}

func fetchStravaActivities(accessToken string, page int) ([]StravaActivity, error) {
	stravaRoutesURL := fmt.Sprintf("https://www.strava.com/api/v3/athlete/activities?page=%d", page)
	req, err := http.NewRequest("GET", stravaRoutesURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch activities: received status %d", resp.StatusCode)
	}

	var activities []StravaActivity
	if err := json.NewDecoder(resp.Body).Decode(&activities); err != nil {
		return nil, err
	}

	return activities, nil
}

func syncTrailsWithRoutes(app core.App, accessToken string, user string, routes []StravaRoute) (bool, error) {
	hasNewRoutes := false
	for _, route := range routes {
		trails, err := app.FindRecordsByFilter("trails", "external_id = {:id}", "", 1, 0, dbx.Params{"id": route.IDStr})
		if err != nil {
			return hasNewRoutes, err
		}
		if len(trails) != 0 {
			continue
		}
		hasNewRoutes = true
		gpx, err := fetchRouteGPX(route, accessToken)
		if err != nil {
			app.Logger().Warn(fmt.Sprintf("Unable to fetch GPX for route '%s': %v", route.Name, err))
			continue
		}
		wpIds, err := createWaypointsFromRoute(app, route, user)
		if err != nil {
			app.Logger().Warn(fmt.Sprintf("Unable to create waypoints for route '%s': %v", route.Name, err))
			continue
		}
		err = createTrailFromRoute(app, route, gpx, user, wpIds)
		if err != nil {
			app.Logger().Warn(fmt.Sprintf("Unable to create trail for route '%s': %v", route.Name, err))
			continue
		}

	}

	return hasNewRoutes, nil
}

func fetchRouteGPX(route StravaRoute, accessToken string) (*filesystem.File, error) {
	url := fmt.Sprintf("https://www.strava.com/api/v3/routes/%s/export_gpx", route.IDStr)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		if resp.Body != nil {
			resp.Body.Close()
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch GPX: received status %d", resp.StatusCode)
	}

	var buf bytes.Buffer
	_, err = io.Copy(&buf, resp.Body)
	if err != nil {
		return nil, err
	}

	gpxFile, err := filesystem.NewFileFromBytes(buf.Bytes(), route.Name+".gpx")
	if err != nil {
		return nil, err
	}

	return gpxFile, nil
}

func createTrailFromRoute(app core.App, route StravaRoute, gpx *filesystem.File, user string, wpIds []string) error {
	collection, err := app.FindCollectionByNameOrId("trails")
	if err != nil {
		return err
	}

	record := core.NewRecord(collection)

	buf := []byte(route.Map.SummaryPolyline)
	coords, _, _ := polyline.DecodeCoords(buf)

	var lat, lon float64
	if len(coords) > 0 && len(coords[0]) >= 2 {
		lat = coords[0][0]
		lon = coords[0][1]
	} else {
		app.Logger().Warn("Warning: No coordinates available, setting lat/lon to 0")
		lat, lon = 0, 0
	}

	bikeCategory, _ := app.FindFirstRecordByData("categories", "name", "Biking")
	hikeCategory, _ := app.FindFirstRecordByData("categories", "name", "Walking")

	category := ""

	if route.Type == 1 && bikeCategory != nil {
		category = bikeCategory.Id
	} else if route.Type == 2 && hikeCategory != nil {
		category = hikeCategory.Id
	}

	record.Load(map[string]any{
		"name":              route.Name,
		"description":       route.Description,
		"public":            !route.Private,
		"distance":          route.Distance,
		"elevation_gain":    route.ElevationGain,
		"duration":          route.EstimatedMovingTime / 60,
		"date":              time.Unix(int64(route.Timestamp), 0),
		"external_provider": "strava",
		"external_id":       route.IDStr,
		"lat":               lat,
		"lon":               lon,
		"waypoints":         wpIds,
		"difficulty":        "easy",
		"category":          category,
		"author":            user,
	})

	if gpx != nil {
		record.Set("gpx", gpx)
	}

	if err := app.Save(record); err != nil {
		return err
	}

	return nil
}

func createWaypointsFromRoute(app core.App, route StravaRoute, user string) ([]string, error) {
	collection, err := app.FindCollectionByNameOrId("waypoints")
	if err != nil {
		return nil, err
	}

	wpIds := make([]string, len(route.Waypoints))

	for i, wp := range route.Waypoints {
		record := core.NewRecord(collection)

		record.Set("name", strconv.Itoa(i))
		record.Set("description", wp.Description)
		record.Set("lat", wp.Latlng[0])
		record.Set("lon", wp.Latlng[1])
		record.Set("icon", "circle")
		record.Set("author", user)
		record.Set("distance_from_start", wp.DistanceIntoRoute)

		app.Save(record)

		wpIds[i] = record.Id
	}

	return wpIds, nil
}

func syncTrailsWithActivities(app core.App, accessToken string, user string, activities []StravaActivity) (bool, error) {
	hasNewActivites := false
	for _, activity := range activities {
		trails, err := app.FindRecordsByFilter("trails", "external_id = {:id}", "", 1, 0, dbx.Params{"id": strconv.Itoa(int(activity.ID))})
		if err != nil {
			return hasNewActivites, err
		}
		if len(trails) != 0 {
			continue
		}
		hasNewActivites = true
		detailedActivity, err := fetchDetailedActivity(activity, accessToken)
		if err != nil {
			app.Logger().Warn(fmt.Sprintf("Unable to fetch detailed activity '%s': %v", activity.Name, err))
			continue
		}
		gpx, err := generateActivityGPX(detailedActivity, accessToken)
		if err != nil {
			app.Logger().Warn(fmt.Sprintf("Unable to fetch GPX for activity '%s': %v", activity.Name, err))
			continue
		}
		err = createTrailFromActivity(app, detailedActivity, gpx, user)
		if err != nil {
			app.Logger().Warn(fmt.Sprintf("Unable to create trail from activity '%s': %v", activity.Name, err))
			continue
		}

	}

	return hasNewActivites, nil
}

func fetchDetailedActivity(activity StravaActivity, accessToken string) (*DetailedStravaActivity, error) {
	url := fmt.Sprintf("https://www.strava.com/api/v3/activities/%d", activity.ID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch activity: received status %d", resp.StatusCode)
	}

	var detailedActivity DetailedStravaActivity
	if err := json.NewDecoder(resp.Body).Decode(&detailedActivity); err != nil {
		return nil, err
	}

	return &detailedActivity, nil
}

func createTrailFromActivity(app core.App, activity *DetailedStravaActivity, gpx *filesystem.File, user string) error {
	if len(activity.StartLatlng) < 2 {
		return nil
	}

	collection, err := app.FindCollectionByNameOrId("trails")
	if err != nil {
		return err
	}

	var photo *filesystem.File
	if len(activity.Photos.Primary.Urls.Num600) > 0 {
		photo, err = fetchActivityPhoto(activity)
		if err != nil {
			return err
		}
	}

	record := core.NewRecord(collection)

	activityMap := map[string]string{
		"AlpineSki":       "Skiing",
		"BackcountrySki":  "Skiing",
		"Canoeing":        "Canoeing",
		"Crossfit":        "Workout",
		"EBikeRide":       "Biking",
		"Elliptical":      "Workout",
		"Golf":            "Walking",
		"Handcycle":       "Biking",
		"Hike":            "Hiking",
		"IceSkate":        "Skiing",
		"InlineSkate":     "Biking",
		"Kayaking":        "Canoeing",
		"Kitesurf":        "Canoeing",
		"NordicSki":       "Skiing",
		"Ride":            "Biking",
		"RockClimbing":    "Climbing",
		"RollerSki":       "Skiing",
		"Rowing":          "Canoeing",
		"Run":             "Walking",
		"Sail":            "Canoeing",
		"Skateboard":      "Walking",
		"Snowboard":       "Skiing",
		"Snowshoe":        "Hiking",
		"Soccer":          "Workout",
		"StairStepper":    "Workout",
		"StandUpPaddling": "Canoeing",
		"Surfing":         "Canoeing",
		"Swim":            "Workout",
		"Velomobile":      "Biking",
		"VirtualRide":     "Biking",
		"VirtualRun":      "Walking",
		"Walk":            "Walking",
		"WeightTraining":  "Workout",
		"Wheelchair":      "Walking",
		"Windsurf":        "Canoeing",
		"Workout":         "Workout",
		"Yoga":            "Workout",
	}

	category, _ := app.FindFirstRecordByData("categories", "name", activityMap[activity.Type])
	categoryId := ""
	if category != nil {
		categoryId = category.Id
	}

	record.Load(map[string]any{
		"name":              activity.Name,
		"description":       activity.Description,
		"public":            !activity.Private,
		"distance":          activity.Distance,
		"elevation_gain":    activity.TotalElevationGain,
		"duration":          activity.ElapsedTime / 60,
		"date":              activity.StartDate,
		"external_provider": "strava",
		"external_id":       activity.ID,
		"lat":               activity.StartLatlng[0],
		"lon":               activity.StartLatlng[1],
		"difficulty":        "easy",
		"category":          categoryId,
		"author":            user,
	})

	if photo != nil {
		record.Set("photos", photo)
	}

	if gpx != nil {
		record.Set("gpx", gpx)
	}

	if err := app.Save(record); err != nil {
		return err
	}

	return nil
}

func fetchActivityPhoto(activity *DetailedStravaActivity) (*filesystem.File, error) {
	req, err := http.NewRequest("GET", activity.Photos.Primary.Urls.Num600, nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch photo: received status %d", resp.StatusCode)
	}

	var buf bytes.Buffer
	_, err = io.Copy(&buf, resp.Body)
	if err != nil {
		return nil, err
	}

	photo, err := filesystem.NewFileFromBytes(buf.Bytes(), "photo")
	if err != nil {
		return nil, err
	}

	return photo, nil
}

func generateActivityGPX(activity *DetailedStravaActivity, accessToken string) (*filesystem.File, error) {
	url := fmt.Sprintf("https://www.strava.com/api/v3/activities/%d/streams?keys=latlng,time,altitude&key_by_type=true", activity.ID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch activity: %s", resp.Status)
	}

	var streamResponse ActivityStreamResponse
	if err := json.NewDecoder(resp.Body).Decode(&streamResponse); err != nil {
		return nil, err
	}

	latLngStream := streamResponse.LatLng
	timeStream := streamResponse.Time
	altitudeStream := streamResponse.Altitude

	var points []*gpx.WptType

	for i, latlng := range latLngStream.Data {
		lat := latlng[0]
		lon := latlng[1]
		alt := altitudeStream.Data[i]
		t := activity.StartDate.Unix() + int64(timeStream.Data[i])

		points = append(points, &gpx.WptType{Lat: lat, Lon: lon, Ele: alt, Time: time.Unix(t, 0)})
	}

	gpx := &gpx.GPX{
		Version: "1.1",
		Creator: "Strava GPX Exporter",
		Trk: []*gpx.TrkType{
			{
				Name: activity.Name,
				TrkSeg: []*gpx.TrkSegType{
					{
						TrkPt: points,
					},
				},
			},
		},
	}
	var buf bytes.Buffer
	err = gpx.Write(&buf)
	if err != nil {
		return nil, err
	}

	gpxFile, err := filesystem.NewFileFromBytes(buf.Bytes(), activity.Name+".gpx")
	if err != nil {
		return nil, err
	}

	return gpxFile, nil
}
