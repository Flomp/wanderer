package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand/v2"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/meilisearch/meilisearch-go"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/plugins/migratecmd"
	"github.com/pocketbase/pocketbase/tools/filesystem"
	"github.com/pocketbase/pocketbase/tools/security"

	"pocketbase/integrations/komoot"
	"pocketbase/integrations/strava"

	_ "pocketbase/migrations"
	"pocketbase/util"
)

func main() {
	app := pocketbase.New()
	client := initializeMeiliSearch()

	registerMigrations(app)
	setupEventHandlers(app, client)

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}

func initializeMeiliSearch() meilisearch.ServiceManager {
	return meilisearch.New(
		os.Getenv("MEILI_URL"),
		meilisearch.WithAPIKey(os.Getenv("MEILI_MASTER_KEY")),
	)
}

func registerMigrations(app *pocketbase.PocketBase) {
	migratecmd.MustRegister(app, app.RootCmd, migratecmd.Config{
		Dir:         "migrations",
		Automigrate: true,
	})
}

func setupEventHandlers(app *pocketbase.PocketBase, client meilisearch.ServiceManager) {
	app.OnRecordAfterCreateSuccess("users").BindFunc(createUserHandler(app, client))

	app.OnRecordAfterCreateSuccess("trails").BindFunc(createTrailHandler(app, client))
	app.OnRecordAfterUpdateSuccess("trails").BindFunc(updateTrailHandler(app, client))
	app.OnRecordAfterDeleteSuccess("trails").BindFunc(deleteTrailHandler(client))

	app.OnRecordAfterCreateSuccess("trail_share").BindFunc(createTrailShareHandler(app, client))
	app.OnRecordAfterDeleteSuccess("trail_share").BindFunc(deleteTrailShareHandler(client))

	app.OnRecordAfterCreateSuccess("lists").BindFunc(createListHandler(app, client))
	app.OnRecordAfterUpdateSuccess("lists").BindFunc(updateListHandler(client))
	app.OnRecordAfterDeleteSuccess("lists").BindFunc(deleteListHandler(client))

	app.OnRecordAfterCreateSuccess("list_share").BindFunc(createListShareHandler(app, client))
	app.OnRecordAfterDeleteSuccess("list_share").BindFunc(deleteListShareHandler(client))

	app.OnRecordAfterCreateSuccess("follows").BindFunc(createFollowHandler(app))
	app.OnRecordAfterCreateSuccess("comments").BindFunc(createCommentHandler(app))

	app.OnRecordsListRequest("integrations").BindFunc(listIntegrationHandler())
	app.OnRecordCreateRequest("integrations").BindFunc(createIntegrationHandler(app))
	app.OnRecordUpdateRequest("integrations").BindFunc(updateIntegrationHandler(app))

	app.OnRecordRequestEmailChangeRequest("users").BindFunc(changeUserEmailHandler(app))
	app.OnServe().BindFunc(onBeforeServeHandler(app, client))
}

func createUserHandler(app *pocketbase.PocketBase, client meilisearch.ServiceManager) func(e *core.RecordEvent) error {
	return func(e *core.RecordEvent) error {
		userId := e.Record.Id

		searchRules := map[string]interface{}{
			"lists": map[string]string{
				"filter": "public = true OR author = " + userId + " OR shares = " + userId,
			},
			"trails": map[string]string{
				"filter": "public = true OR author = " + userId + " OR shares = " + userId,
			},
		}

		token, err := util.GenerateMeilisearchToken(searchRules, client)
		if err != nil {
			return err
		}
		e.Record.Set("token", token)
		if err := app.Save(e.Record); err != nil {
			return err
		}

		return createDefaultUserSettings(app, e.Record.Id)
	}
}

func createDefaultUserSettings(app *pocketbase.PocketBase, userId string) error {
	collection, err := app.FindCollectionByNameOrId("settings")
	if err != nil {
		return err
	}
	settings := core.NewRecord(collection)
	settings.Set("language", "en")
	settings.Set("unit", "metric")
	settings.Set("mapFocus", "trails")
	settings.Set("user", userId)
	return app.Save(settings)
}

func createTrailHandler(app *pocketbase.PocketBase, client meilisearch.ServiceManager) func(e *core.RecordEvent) error {
	return func(e *core.RecordEvent) error {
		record := e.Record
		author, err := app.FindRecordById("users", record.GetString(("author")))
		if err != nil {
			return err
		}
		if err := util.IndexTrail(record, author, client); err != nil {
			return err
		}

		if record.GetBool("public") {
			notification := util.Notification{
				Type: util.TrailCreate,
				Metadata: map[string]string{
					"id":    record.Id,
					"trail": record.GetString("name"),
				},
				Seen:   false,
				Author: record.GetString("author"),
			}
			return util.SendNotificationToFollowers(app, notification)
		}
		return nil
	}
}

func updateTrailHandler(app *pocketbase.PocketBase, client meilisearch.ServiceManager) func(e *core.RecordEvent) error {
	return func(e *core.RecordEvent) error {
		record := e.Record
		author, err := app.FindRecordById("users", record.GetString(("author")))
		if err != nil {
			return err
		}
		return util.UpdateTrail(record, author, client)
	}
}

func deleteTrailHandler(client meilisearch.ServiceManager) func(e *core.RecordEvent) error {
	return func(e *core.RecordEvent) error {
		record := e.Record
		_, err := client.Index("trails").DeleteDocument(record.Id)
		return err
	}
}

func createTrailShareHandler(app *pocketbase.PocketBase, client meilisearch.ServiceManager) func(e *core.RecordEvent) error {
	return func(e *core.RecordEvent) error {
		record := e.Record

		trailId := record.GetString("trail")
		shares, err := app.FindAllRecords("trail_share",
			dbx.NewExp("trail = {:trailId}", dbx.Params{"trailId": trailId}),
		)
		if err != nil {
			return err
		}
		userIds := make([]string, len(shares))
		for i, r := range shares {
			userIds[i] = r.GetString("user")
		}
		err = util.UpdateTrailShares(trailId, userIds, client)

		if err != nil {
			return err
		}

		if errs := app.ExpandRecord(record, []string{"trail", "trail.author"}, nil); len(errs) > 0 {
			return fmt.Errorf("failed to expand: %v", errs)
		}
		shareTrail := record.ExpandedOne("trail")
		shareTrailAuthor := shareTrail.ExpandedOne("author")

		notification := util.Notification{
			Type: util.TrailShare,
			Metadata: map[string]string{
				"id":     shareTrail.Id,
				"trail":  shareTrail.GetString("name"),
				"author": shareTrailAuthor.GetString("username"),
			},
			Seen:   false,
			Author: shareTrailAuthor.Id,
		}
		return util.SendNotification(app, notification, record.GetString("user"))
	}
}

func deleteTrailShareHandler(client meilisearch.ServiceManager) func(e *core.RecordEvent) error {
	return func(e *core.RecordEvent) error {
		record := e.Record

		trailId := record.GetString("trail")
		return util.UpdateTrailShares(trailId, []string{}, client)
	}
}

func createListHandler(app *pocketbase.PocketBase, client meilisearch.ServiceManager) func(e *core.RecordEvent) error {
	return func(e *core.RecordEvent) error {
		record := e.Record

		if err := util.IndexList(record, client); err != nil {
			return err
		}
		if !record.GetBool("public") {
			return nil
		}
		notification := util.Notification{
			Type: util.ListCreate,
			Metadata: map[string]string{
				"id":   record.Id,
				"list": record.GetString("name"),
			},
			Seen:   false,
			Author: record.GetString("author"),
		}
		return util.SendNotificationToFollowers(app, notification)
	}
}

func updateListHandler(client meilisearch.ServiceManager) func(e *core.RecordEvent) error {
	return func(e *core.RecordEvent) error {
		record := e.Record
		return util.UpdateList(record, client)
	}
}

func deleteListHandler(client meilisearch.ServiceManager) func(e *core.RecordEvent) error {
	return func(e *core.RecordEvent) error {
		record := e.Record
		_, err := client.Index("lists").DeleteDocument(record.Id)
		return err
	}
}

func createListShareHandler(app *pocketbase.PocketBase, client meilisearch.ServiceManager) func(e *core.RecordEvent) error {
	return func(e *core.RecordEvent) error {
		record := e.Record
		listId := record.GetString("list")
		shares, err := app.FindAllRecords("list_share",
			dbx.NewExp("list = {:listId}", dbx.Params{"listId": listId}),
		)
		if err != nil {
			return err
		}
		userIds := make([]string, len(shares))
		for i, r := range shares {
			userIds[i] = r.GetString("user")
		}
		err = util.UpdateListShares(listId, userIds, client)

		if err != nil {
			return err
		}

		if errs := app.ExpandRecord(record, []string{"list", "list.author"}, nil); len(errs) > 0 {
			return fmt.Errorf("failed to expand: %v", errs)
		}
		shareList := record.ExpandedOne("list")
		shareListAuthor := shareList.ExpandedOne("author")

		notification := util.Notification{
			Type: util.ListShare,
			Metadata: map[string]string{
				"id":     shareList.Id,
				"list":   shareList.GetString("name"),
				"author": shareListAuthor.GetString("username"),
			},
			Seen:   false,
			Author: shareListAuthor.Id,
		}
		return util.SendNotification(app, notification, record.GetString("user"))
	}
}

func deleteListShareHandler(client meilisearch.ServiceManager) func(e *core.RecordEvent) error {
	return func(e *core.RecordEvent) error {
		record := e.Record
		listId := record.GetString("list")
		return util.UpdateListShares(listId, []string{}, client)
	}
}

func createFollowHandler(app *pocketbase.PocketBase) func(e *core.RecordEvent) error {
	return func(e *core.RecordEvent) error {
		record := e.Record
		if errs := app.ExpandRecord(record, []string{"follower"}, nil); len(errs) > 0 {
			return fmt.Errorf("failed to expand: %v", errs)
		}
		follower := record.ExpandedOne("follower")

		notification := util.Notification{
			Type: util.NewFollower,
			Metadata: map[string]string{
				"follower": follower.GetString("username"),
			},
			Seen:   false,
			Author: record.GetString("follower"),
		}
		return util.SendNotification(app, notification, record.GetString("followee"))
	}
}

func createCommentHandler(app *pocketbase.PocketBase) func(e *core.RecordEvent) error {
	return func(e *core.RecordEvent) error {
		record := e.Record

		if errs := app.ExpandRecord(record, []string{"trail", "author"}, nil); len(errs) > 0 {
			return fmt.Errorf("failed to expand: %v", errs)
		}
		commentAuthor := record.ExpandedOne("author")
		commentTrail := record.ExpandedOne("trail")

		notification := util.Notification{
			Type: util.TrailComment,
			Metadata: map[string]string{
				"id":      commentTrail.Id,
				"author":  commentAuthor.GetString("username"),
				"trail":   commentTrail.GetString("name"),
				"comment": record.GetString("text"),
			},
			Seen:   false,
			Author: record.GetString("author"),
		}
		return util.SendNotification(app, notification, commentTrail.GetString("author"))
	}
}

func listIntegrationHandler() func(e *core.RecordsListRequestEvent) error {
	return func(e *core.RecordsListRequestEvent) error {
		if e.HasSuperuserAuth() {
			return nil
		}
		for _, r := range e.Records {

			err := censorIntegrationSecrets(r)
			if err != nil {
				return err
			}
		}

		return nil
	}
}

func createIntegrationHandler(app *pocketbase.PocketBase) func(e *core.RecordRequestEvent) error {
	return func(e *core.RecordRequestEvent) error {
		err := encryptIntegrationSecrets(app, e.Record)
		if err != nil {
			return err
		}

		if err := e.Next(); err != nil {
			return err
		}

		err = censorIntegrationSecrets(e.Record)
		if err != nil {
			return err
		}
		return nil
	}
}

func updateIntegrationHandler(app *pocketbase.PocketBase) func(e *core.RecordRequestEvent) error {
	return func(e *core.RecordRequestEvent) error {
		err := encryptIntegrationSecrets(app, e.Record)
		if err != nil {
			return err
		}

		if err := e.Next(); err != nil {
			return err
		}

		err = censorIntegrationSecrets(e.Record)
		if err != nil {
			return err
		}
		return nil
	}
}
func censorIntegrationSecrets(r *core.Record) error {
	secrets := map[string][]string{
		"strava": {"clientSecret", "refreshToken", "accessToken", "expiresAt"},
		"komoot": {"password"},
	}
	for key, secretKeys := range secrets {
		if integrationString := r.GetString(key); integrationString != "" {
			var integration map[string]interface{}
			if err := json.Unmarshal([]byte(integrationString), &integration); err != nil {
				return err
			}
			for _, secretKey := range secretKeys {
				integration[secretKey] = ""
			}
			b, err := json.Marshal(integration)
			if err != nil {
				return err
			}
			r.Set(key, string(b))
		}
	}

	return nil
}

func encryptIntegrationSecrets(app *pocketbase.PocketBase, r *core.Record) error {
	encryptionKey := os.Getenv("POCKETBASE_ENCRYPTION_KEY")
	if len(encryptionKey) == 0 {
		return apis.NewBadRequestError("POCKETBASE_ENCRYPTION_KEY not set", nil)
	}

	secrets := map[string][]string{
		"strava": {"clientSecret", "refreshToken", "accessToken", "expiresAt"},
		"komoot": {"password"},
	}

	original, _ := app.FindRecordById("integrations", r.Id)

	for key, secretKeys := range secrets {
		if integrationString := r.GetString(key); integrationString != "" {
			var integration map[string]interface{}
			if err := json.Unmarshal([]byte(integrationString), &integration); err != nil {
				return err
			}

			for _, secretKey := range secretKeys {
				if secret, ok := integration[secretKey].(string); ok && len(secret) > 0 {
					encryptedSecret, err := security.Encrypt([]byte(secret), encryptionKey)
					if err != nil {
						return err
					}
					integration[secretKey] = encryptedSecret
				} else if original != nil {

					originalString := original.GetString(key)
					var originalIntegration map[string]interface{}
					if err := json.Unmarshal([]byte(originalString), &originalIntegration); err != nil {
						return err
					}
					integration[secretKey] = originalIntegration[secretKey]
				}
			}

			b, err := json.Marshal(integration)
			if err != nil {
				return err
			}
			r.Set(key, string(b))
		}
	}

	return nil
}

func changeUserEmailHandler(app *pocketbase.PocketBase) func(e *core.RecordRequestEmailChangeRequestEvent) error {
	return func(e *core.RecordRequestEmailChangeRequestEvent) error {

		e.Record.Set("email", e.Record.Email())
		if err := app.Save(e.Record); err != nil {
			return err
		}
		return nil
	}
}

func onBeforeServeHandler(app *pocketbase.PocketBase, client meilisearch.ServiceManager) func(e *core.ServeEvent) error {
	return func(e *core.ServeEvent) error {
		registerRoutes(e, app, client)
		registerCronJobs(app)
		bootstrapData(app, client)

		return e.Next()
	}

}

func registerRoutes(e *core.ServeEvent, app *pocketbase.PocketBase, client meilisearch.ServiceManager) {
	e.Router.GET("/public/search/token", func(e *core.RequestEvent) error {
		searchRules := map[string]interface{}{
			"lists": map[string]string{
				"filter": "public = true",
			},
			"trails": map[string]string{
				"filter": "public = true",
			},
		}
		token, err := util.GenerateMeilisearchToken(searchRules, client)
		if err != nil {
			return err
		}
		return e.JSON(http.StatusOK, map[string]string{"token": token})
	})
	e.Router.GET("/trail/recommend", func(e *core.RequestEvent) error {
		qSize := e.Request.URL.Query().Get("size")
		size, err := strconv.Atoi(qSize)
		if err != nil {
			size = 4
		}
		userId := e.Auth.Id

		trails, err := app.FindRecordsByFilter(
			"trails",
			"author = {:userId} || public = true || ({:userId} != '' && trail_share_via_trail.user ?= {:userId})",
			"",
			-1,
			0,
			dbx.Params{"userId": userId},
		)
		if err != nil {
			return err
		}
		if len(trails) < size {
			size = len(trails)
		}
		rand.Shuffle(len(trails), func(i, j int) {
			trails[i], trails[j] = trails[j], trails[i]
		})
		randomTrails := trails[:size]
		return e.JSON(http.StatusOK, randomTrails)

	})

	e.Router.POST("/integration/strava/token", func(e *core.RequestEvent) error {
		encryptionKey := os.Getenv("POCKETBASE_ENCRYPTION_KEY")
		if len(encryptionKey) == 0 {
			return apis.NewBadRequestError("POCKETBASE_ENCRYPTION_KEY not set", nil)
		}

		var data strava.TokenRequest
		if err := e.BindBody(&data); err != nil {
			return apis.NewBadRequestError("Failed to read request data", err)
		}

		userId := e.Auth.Id

		integrations, err := app.FindAllRecords("integrations", dbx.NewExp("user = {:id}", dbx.Params{"id": userId}))
		if err != nil {
			return err
		}
		if len(integrations) == 0 {
			return apis.NewBadRequestError("user has no integration", nil)
		}
		integration := integrations[0]
		stravaString := integration.GetString("strava")
		if len(stravaString) == 0 {
			return apis.NewBadRequestError("strava integration missing", nil)
		}
		var stravaIntegration strava.StravaIntegration
		err = json.Unmarshal([]byte(stravaString), &stravaIntegration)
		if err != nil {
			return err
		}
		decryptedSecret, err := security.Decrypt(stravaIntegration.ClientSecret, encryptionKey)
		if err != nil {
			return err
		}

		request := strava.TokenRequest{
			ClientID:     stravaIntegration.ClientID,
			ClientSecret: string(decryptedSecret),
			Code:         data.Code,
			GrantType:    "authorization_code",
		}
		r, err := strava.GetStravaToken(request)
		if err != nil {
			return err
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

		stravaIntegration.Active = true

		b, err := json.Marshal(stravaIntegration)
		if err != nil {
			return err
		}
		integration.Set("strava", string(b))
		err = app.Save(integration)
		if err != nil {
			return err
		}
		return e.JSON(http.StatusOK, nil)
	})

	e.Router.GET("/integration/komoot/login", func(e *core.RequestEvent) error {
		encryptionKey := os.Getenv("POCKETBASE_ENCRYPTION_KEY")
		if len(encryptionKey) == 0 {
			return apis.NewBadRequestError("POCKETBASE_ENCRYPTION_KEY not set", nil)
		}

		userId := e.Auth.Id

		integrations, err := app.FindAllRecords("integrations", dbx.NewExp("user = {:id}", dbx.Params{"id": userId}))
		if err != nil {
			return err
		}
		if len(integrations) == 0 {
			return apis.NewBadRequestError("user has no integration", nil)
		}
		integration := integrations[0]
		komootString := integration.GetString("komoot")
		if len(komootString) == 0 {
			return apis.NewBadRequestError("komoot integration missing", nil)
		}
		var komootIntegration komoot.KomootIntegration
		err = json.Unmarshal([]byte(komootString), &komootIntegration)
		if err != nil {
			return err
		}
		decryptedPassword, err := security.Decrypt(komootIntegration.Password, encryptionKey)
		if err != nil {
			return err
		}

		k := &komoot.KomootApi{}

		err = k.Login(komootIntegration.Email, string(decryptedPassword))
		if err != nil {
			return apis.NewUnauthorizedError("invalid credentials", nil)
		}

		return e.JSON(http.StatusOK, nil)
	})
}

func registerCronJobs(app *pocketbase.PocketBase) {
	schedule := os.Getenv("POCKETBASE_CRON_SYNC_SCHEDULE")
	if len(schedule) == 0 {
		schedule = "0 2 * * *"
	}

	app.Cron().MustAdd("integrations", schedule, func() {
		err := strava.SyncStrava(app)
		if err != nil {
			warning := fmt.Sprintf("Error syncing with strava: %v", err)
			fmt.Println(warning)
			app.Logger().Error(warning)
		}
		err = komoot.SyncKomoot(app)
		if err != nil {
			warning := fmt.Sprintf("Error syncing with komoot: %v", err)
			fmt.Println(warning)
			app.Logger().Error(warning)
		}
	})
}

func bootstrapData(app *pocketbase.PocketBase, client meilisearch.ServiceManager) error {
	bootstrapCategories(app)
	bootstrapMeilisearchTrails(app, client)
	return nil
}

func bootstrapCategories(app *pocketbase.PocketBase) error {
	query := app.RecordQuery("categories")
	records := []*core.Record{}

	if err := query.All(&records); err != nil {
		return err
	}
	if len(records) == 0 {
		collection, _ := app.FindCollectionByNameOrId("categories")

		categories := []string{"Hiking", "Walking", "Climbing", "Skiing", "Canoeing", "Biking"}
		for _, element := range categories {
			record := core.NewRecord(collection)
			record.Set("name", element)
			f, _ := filesystem.NewFileFromPath("migrations/initial_data/" + strings.ToLower(element) + ".jpg")
			record.Set("img", f)
			err := app.Save(record)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func bootstrapMeilisearchTrails(app *pocketbase.PocketBase, client meilisearch.ServiceManager) error {
	query := app.RecordQuery("trails")
	trails := []*core.Record{}

	if err := query.All(&trails); err != nil {
		return err
	}

	_, err := client.Index("trails").DeleteAllDocuments()
	if err != nil {
		return err
	}
	for _, trail := range trails {
		author, err := app.FindRecordById("users", trail.GetString(("author")))
		if err != nil {
			return err
		}
		if err := util.IndexTrail(trail, author, client); err != nil {
			return err
		}
	}
	return nil
}
