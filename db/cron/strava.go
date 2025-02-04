package cron

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/forms"
	"github.com/pocketbase/pocketbase/models"
	"github.com/pocketbase/pocketbase/tools/filesystem"
	"github.com/twpayne/go-gpx"
	"github.com/twpayne/go-polyline"
)

func SyncStrava(app *pocketbase.PocketBase) error {
	integrations, err := app.Dao().FindRecordsByExpr("integrations", dbx.NewExp("true"))
	if err != nil {
		return err
	}

	for _, i := range integrations {
		userId := i.GetString("user")
		stravaString := i.GetString("strava")
		var stravaIntegration StravaIntegration
		json.Unmarshal([]byte(stravaString), &stravaIntegration)

		if !stravaIntegration.Active || stravaIntegration.RefreshToken == nil {
			continue
		}

		r, err := refreshStravaToken(stravaIntegration.ClientID, stravaIntegration.ClientSecret, *stravaIntegration.RefreshToken)
		if err != nil {
			return err
		}
		stravaIntegration.AccessToken = &r.AccessToken
		stravaIntegration.RefreshToken = &r.RefreshToken
		stravaIntegration.ExpiresAt = r.ExpiresAt

		b, err := json.Marshal(stravaIntegration)
		if err != nil {
			return err
		}
		i.Set("strava", string(b))
		app.Dao().SaveRecord(i)

		if stravaIntegration.Routes {
			page := 1
			hasNewRoutes := true
			for hasNewRoutes {

				routes, err := fetchStravaRoutes(r.AccessToken, page)
				if err != nil {
					return err
				}
				hasNewRoutes, err = syncTrailsWithRoutes(app, r.AccessToken, userId, routes)
				if err != nil {
					return err
				}
				page += 1
			}
		}
		if stravaIntegration.Activities {
			page := 1
			hasNewActivities := true
			for hasNewActivities {
				activities, err := fetchStravaActivities(r.AccessToken, page)
				if err != nil {
					return err
				}
				hasNewActivities, err = syncTrailsWithActivities(app, r.AccessToken, userId, activities)
				if err != nil {
					return err
				}
				page += 1
			}
		}
	}

	return nil
}

func refreshStravaToken(clientID int32, clientSecret, refreshToken string) (*RefreshTokenResponse, error) {
	const stravaTokenURL = "https://www.strava.com/oauth/token"

	requestBody, err := json.Marshal(RefreshTokenRequest{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RefreshToken: refreshToken,
		GrantType:    "refresh_token",
	})
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
		return nil, fmt.Errorf("failed to refresh token: received status %d", resp.StatusCode)
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

func syncTrailsWithRoutes(app *pocketbase.PocketBase, accessToken string, user string, routes []StravaRoute) (bool, error) {
	hasNewRoutes := false
	for _, route := range routes {
		trails, err := app.Dao().FindRecordsByFilter("trails", "external_id = {:id}", "", 1, 0, dbx.Params{"id": route.IDStr})
		if err != nil {
			return hasNewRoutes, err
		}
		if len(trails) != 0 {
			continue
		}
		hasNewRoutes = true
		gpx, err := fetchRouteGPX(route, accessToken)
		if err != nil {
			return hasNewRoutes, err
		}
		wpIds, err := createWaypointsFromRoute(app, route, user)
		if err != nil {
			return hasNewRoutes, err
		}
		err = createTrailFromRoute(app, route, gpx, user, wpIds)
		if err != nil {
			return hasNewRoutes, err
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

func createTrailFromRoute(app *pocketbase.PocketBase, route StravaRoute, gpx *filesystem.File, user string, wpIds []string) error {
	collection, err := app.Dao().FindCollectionByNameOrId("trails")
	if err != nil {
		return err
	}

	record := models.NewRecord(collection)

	form := forms.NewRecordUpsert(app, record)

	buf := []byte(route.Map.SummaryPolyline)
	coords, _, _ := polyline.DecodeCoords(buf)

	var lat, lon float64
	if len(coords) > 0 && len(coords[0]) >= 2 {
		lat = coords[0][0]
		lon = coords[0][1]
	} else {
		fmt.Println("Warning: No coordinates available, setting lat/lon to 0")
		lat, lon = 0, 0
	}

	bikeCategory, _ := app.Dao().FindFirstRecordByData("categories", "name", "Biking")
	hikeCategory, _ := app.Dao().FindFirstRecordByData("categories", "name", "Walking")

	category := ""

	if route.Type == 1 && bikeCategory != nil {
		category = bikeCategory.Id
	} else if route.Type == 2 && hikeCategory != nil {
		category = hikeCategory.Id
	}

	form.LoadData(map[string]any{
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

	form.AddFiles("gpx", gpx)

	if err := form.Submit(); err != nil {
		return err
	}

	return nil
}

func createWaypointsFromRoute(app *pocketbase.PocketBase, route StravaRoute, user string) ([]string, error) {
	collection, err := app.Dao().FindCollectionByNameOrId("waypoints")
	if err != nil {
		return nil, err
	}

	wpIds := make([]string, len(route.Waypoints))

	for i, wp := range route.Waypoints {
		record := models.NewRecord(collection)

		record.Set("name", string(i))
		record.Set("description", wp.Description)
		record.Set("lat", wp.Latlng[0])
		record.Set("lon", wp.Latlng[1])
		record.Set("icon", "circle")
		record.Set("author", user)
		record.Set("distance_from_start", wp.DistanceIntoRoute)

		app.Dao().SaveRecord(record)

		wpIds[i] = record.Id
	}

	return wpIds, nil
}

func syncTrailsWithActivities(app *pocketbase.PocketBase, accessToken string, user string, activities []StravaActivity) (bool, error) {
	hasNewActivites := false
	for _, activity := range activities {
		trails, err := app.Dao().FindRecordsByFilter("trails", "external_id = {:id}", "", 1, 0, dbx.Params{"id": strconv.Itoa(int(activity.ID))})
		if err != nil {
			return hasNewActivites, err
		}
		if len(trails) != 0 {
			continue
		}
		hasNewActivites = true
		detailedActivity, err := fetchDetailedActivity(activity, accessToken)
		if err != nil {
			return hasNewActivites, err
		}
		gpx, err := generateActivityGPX(detailedActivity, accessToken)
		if err != nil {
			return hasNewActivites, err
		}
		err = createTrailFromActivity(app, detailedActivity, gpx, user)
		if err != nil {
			return hasNewActivites, err
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

func createTrailFromActivity(app *pocketbase.PocketBase, activity *DetailedStravaActivity, gpx *filesystem.File, user string) error {
	collection, err := app.Dao().FindCollectionByNameOrId("trails")
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

	record := models.NewRecord(collection)

	form := forms.NewRecordUpsert(app, record)

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

	category, _ := app.Dao().FindFirstRecordByData("categories", "name", activityMap[activity.Type])
	categoryId := ""
	if category != nil {
		categoryId = category.Id
	}

	form.LoadData(map[string]any{
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
		form.AddFiles("photos", photo)
	}
	form.AddFiles("gpx", gpx)

	if err := form.Submit(); err != nil {
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
