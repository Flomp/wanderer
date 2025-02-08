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

	"github.com/labstack/echo/v5"
	"github.com/meilisearch/meilisearch-go"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/forms"
	"github.com/pocketbase/pocketbase/models"
	"github.com/pocketbase/pocketbase/plugins/migratecmd"
	"github.com/pocketbase/pocketbase/tools/cron"
	"github.com/pocketbase/pocketbase/tools/filesystem"
	"github.com/pocketbase/pocketbase/tools/hook"
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
	app.OnModelAfterCreate("users").Add(createUserHandler(app, client))

	app.OnModelAfterCreate("trails").Add(createTrailIndexHandler(client))

	app.OnRecordAfterCreateRequest("trails").Add(createTrailHandler(app))
	app.OnRecordAfterUpdateRequest("trails").Add(updateTrailHandler(client))
	app.OnRecordAfterDeleteRequest("trails").Add(deleteTrailHandler(client))

	app.OnRecordAfterCreateRequest("trail_share").Add(createTrailShareHandler(app, client))
	app.OnRecordAfterDeleteRequest("trail_share").Add(deleteTrailShareHandler(client))

	app.OnRecordAfterCreateRequest("lists").Add(createListHandler(app, client))
	app.OnRecordAfterUpdateRequest("lists").Add(updateListHandler(client))
	app.OnRecordAfterDeleteRequest("lists").Add(deleteListHandler(client))

	app.OnRecordAfterCreateRequest("list_share").Add(createListShareHandler(app, client))
	app.OnRecordAfterDeleteRequest("list_share").Add(deleteListShareHandler(client))

	app.OnRecordAfterCreateRequest("follows").Add(createFollowHandler(app))
	app.OnRecordAfterCreateRequest("comments").Add(createCommentHandler(app))

	app.OnRecordsListRequest("integrations").Add(listIntegrationHandler())
	app.OnRecordAfterCreateRequest("integrations").Add(createIntegrationAfterHandler())
	app.OnRecordAfterUpdateRequest("integrations").Add(updateIntegrationAfterHandler())
	app.OnRecordBeforeCreateRequest("integrations").Add(createIntegrationBeforeHandler(app))
	app.OnRecordBeforeUpdateRequest("integrations").Add(updateIntegrationBeforeHandler(app))

	app.OnRecordBeforeRequestEmailChangeRequest("users").Add(changeUserEmailHandler(app))
	app.OnBeforeServe().Add(onBeforeServeHandler(app, client))
}

func createUserHandler(app *pocketbase.PocketBase, client meilisearch.ServiceManager) func(e *core.ModelEvent) error {
	return func(e *core.ModelEvent) error {
		record := e.Model.(*models.Record)
		userId := record.GetId()

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
		record.Set("token", token)
		if err := app.Dao().SaveRecord(record); err != nil {
			return err
		}

		return createDefaultUserSettings(app, record.Id)
	}
}

func createDefaultUserSettings(app *pocketbase.PocketBase, userId string) error {
	collection, err := app.Dao().FindCollectionByNameOrId("settings")
	if err != nil {
		return err
	}
	settings := models.NewRecord(collection)
	settings.Set("language", "en")
	settings.Set("unit", "metric")
	settings.Set("mapFocus", "trails")
	settings.Set("user", userId)
	return app.Dao().SaveRecord(settings)
}

func createTrailIndexHandler(client meilisearch.ServiceManager) func(e *core.ModelEvent) error {
	return func(e *core.ModelEvent) error {
		record := e.Model.(*models.Record)
		if err := util.IndexTrail(record, client); err != nil {
			return err
		}
		return nil
	}
}

func createTrailHandler(app *pocketbase.PocketBase) func(e *core.RecordCreateEvent) error {
	return func(e *core.RecordCreateEvent) error {
		if e.Record.GetBool("public") {
			notification := util.Notification{
				Type: util.TrailCreate,
				Metadata: map[string]string{
					"id":    e.Record.Id,
					"trail": e.Record.GetString("name"),
				},
				Seen:   false,
				Author: e.Record.GetString("author"),
			}
			return util.SendNotificationToFollowers(app, notification)
		}
		return nil
	}
}

func updateTrailHandler(client meilisearch.ServiceManager) func(e *core.RecordUpdateEvent) error {
	return func(e *core.RecordUpdateEvent) error {
		return util.UpdateTrail(e.Record, client)
	}
}

func deleteTrailHandler(client meilisearch.ServiceManager) func(e *core.RecordDeleteEvent) error {
	return func(e *core.RecordDeleteEvent) error {
		_, err := client.Index("trails").DeleteDocument(e.Record.Id)
		return err
	}
}

func createTrailShareHandler(app *pocketbase.PocketBase, client meilisearch.ServiceManager) func(e *core.RecordCreateEvent) error {
	return func(e *core.RecordCreateEvent) error {
		trailId := e.Record.GetString("trail")
		shares, err := app.Dao().FindRecordsByExpr("trail_share",
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

		if errs := app.Dao().ExpandRecord(e.Record, []string{"trail", "trail.author"}, nil); len(errs) > 0 {
			return fmt.Errorf("failed to expand: %v", errs)
		}
		shareTrail := e.Record.ExpandedOne("trail")
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
		return util.SendNotification(app, notification, e.Record.GetString("user"))
	}
}

func deleteTrailShareHandler(client meilisearch.ServiceManager) func(e *core.RecordDeleteEvent) error {
	return func(e *core.RecordDeleteEvent) error {
		trailId := e.Record.GetString("trail")
		return util.UpdateTrailShares(trailId, []string{}, client)
	}
}

func createListHandler(app *pocketbase.PocketBase, client meilisearch.ServiceManager) func(e *core.RecordCreateEvent) error {
	return func(e *core.RecordCreateEvent) error {
		if err := util.IndexList(e.Record, client); err != nil {
			return err
		}
		if !e.Record.GetBool("public") {
			return nil
		}
		notification := util.Notification{
			Type: util.ListCreate,
			Metadata: map[string]string{
				"id":   e.Record.Id,
				"list": e.Record.GetString("name"),
			},
			Seen:   false,
			Author: e.Record.GetString("author"),
		}
		return util.SendNotificationToFollowers(app, notification)
	}
}

func updateListHandler(client meilisearch.ServiceManager) func(e *core.RecordUpdateEvent) error {
	return func(e *core.RecordUpdateEvent) error {
		return util.UpdateList(e.Record, client)
	}
}

func deleteListHandler(client meilisearch.ServiceManager) func(e *core.RecordDeleteEvent) error {
	return func(e *core.RecordDeleteEvent) error {
		_, err := client.Index("lists").DeleteDocument(e.Record.Id)
		return err
	}
}

func createListShareHandler(app *pocketbase.PocketBase, client meilisearch.ServiceManager) func(e *core.RecordCreateEvent) error {
	return func(e *core.RecordCreateEvent) error {
		listId := e.Record.GetString("list")
		shares, err := app.Dao().FindRecordsByExpr("list_share",
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

		if errs := app.Dao().ExpandRecord(e.Record, []string{"list", "list.author"}, nil); len(errs) > 0 {
			return fmt.Errorf("failed to expand: %v", errs)
		}
		shareList := e.Record.ExpandedOne("list")
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
		return util.SendNotification(app, notification, e.Record.GetString("user"))
	}
}

func deleteListShareHandler(client meilisearch.ServiceManager) func(e *core.RecordDeleteEvent) error {
	return func(e *core.RecordDeleteEvent) error {
		listId := e.Record.GetString("list")
		return util.UpdateListShares(listId, []string{}, client)
	}
}

func createFollowHandler(app *pocketbase.PocketBase) func(e *core.RecordCreateEvent) error {
	return func(e *core.RecordCreateEvent) error {
		if errs := app.Dao().ExpandRecord(e.Record, []string{"follower"}, nil); len(errs) > 0 {
			return fmt.Errorf("failed to expand: %v", errs)
		}
		follower := e.Record.ExpandedOne("follower")

		notification := util.Notification{
			Type: util.NewFollower,
			Metadata: map[string]string{
				"follower": follower.GetString("username"),
			},
			Seen:   false,
			Author: e.Record.GetString("follower"),
		}
		return util.SendNotification(app, notification, e.Record.GetString("followee"))
	}
}

func createCommentHandler(app *pocketbase.PocketBase) func(e *core.RecordCreateEvent) error {
	return func(e *core.RecordCreateEvent) error {

		if errs := app.Dao().ExpandRecord(e.Record, []string{"trail", "author"}, nil); len(errs) > 0 {
			return fmt.Errorf("failed to expand: %v", errs)
		}
		commentAuthor := e.Record.ExpandedOne("author")
		commentTrail := e.Record.ExpandedOne("trail")

		notification := util.Notification{
			Type: util.TrailComment,
			Metadata: map[string]string{
				"id":      commentTrail.Id,
				"author":  commentAuthor.GetString("username"),
				"trail":   commentTrail.GetString("name"),
				"comment": e.Record.GetString("text"),
			},
			Seen:   false,
			Author: e.Record.GetString("author"),
		}
		return util.SendNotification(app, notification, commentTrail.GetString("author"))
	}
}

func listIntegrationHandler() func(e *core.RecordsListEvent) error {
	return func(e *core.RecordsListEvent) error {
		info := apis.RequestInfo(e.HttpContext)
		if info.Admin != nil {
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

func createIntegrationBeforeHandler(app *pocketbase.PocketBase) func(e *core.RecordCreateEvent) error {
	return func(e *core.RecordCreateEvent) error {
		err := encryptIntegrationSecrets(app, e.Record)
		if err != nil {
			return err
		}
		return nil
	}
}

func createIntegrationAfterHandler() func(e *core.RecordCreateEvent) error {
	return func(e *core.RecordCreateEvent) error {
		err := censorIntegrationSecrets(e.Record)
		if err != nil {
			return err
		}
		return nil
	}
}

func updateIntegrationBeforeHandler(app *pocketbase.PocketBase) func(e *core.RecordUpdateEvent) error {
	return func(e *core.RecordUpdateEvent) error {
		err := encryptIntegrationSecrets(app, e.Record)
		if err != nil {
			return err
		}

		return nil
	}
}

func updateIntegrationAfterHandler() func(e *core.RecordUpdateEvent) error {
	return func(e *core.RecordUpdateEvent) error {
		err := censorIntegrationSecrets(e.Record)
		if err != nil {
			return err
		}
		return nil
	}
}

func censorIntegrationSecrets(r *models.Record) error {
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

func encryptIntegrationSecrets(app *pocketbase.PocketBase, r *models.Record) error {
	encryptionKey := os.Getenv("POCKETBASE_ENCRYPTION_KEY")
	if len(encryptionKey) == 0 {
		return apis.NewBadRequestError("POCKETBASE_ENCRYPTION_KEY not set", nil)
	}

	secrets := map[string][]string{
		"strava": {"clientSecret", "refreshToken", "accessToken", "expiresAt"},
		"komoot": {"password"},
	}

	original, _ := app.Dao().FindRecordById("integrations", r.Id)

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

func changeUserEmailHandler(app *pocketbase.PocketBase) func(e *core.RecordRequestEmailChangeEvent) error {
	return func(e *core.RecordRequestEmailChangeEvent) error {
		form := forms.NewRecordEmailChangeRequest(app, e.Record)
		if err := e.HttpContext.Bind(form); err != nil {
			return err
		}
		e.Record.Set("email", form.NewEmail)
		if err := app.Dao().SaveRecord(e.Record); err != nil {
			return err
		}
		return hook.StopPropagation
	}
}

func onBeforeServeHandler(app *pocketbase.PocketBase, client meilisearch.ServiceManager) func(e *core.ServeEvent) error {
	return func(e *core.ServeEvent) error {
		registerRoutes(e, app, client)
		registerCronJobs(app)
		return bootstrapData(app, client)
	}

}

func registerRoutes(e *core.ServeEvent, app *pocketbase.PocketBase, client meilisearch.ServiceManager) {
	e.Router.GET("/public/search/token", func(c echo.Context) error {
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
		return c.JSON(http.StatusOK, map[string]string{"token": token})
	})
	e.Router.GET("/trail/recommend", func(c echo.Context) error {
		qSize := c.QueryParam("size")
		size, err := strconv.Atoi(qSize)
		if err != nil {
			size = 4
		}
		user, success := c.Get(apis.ContextAuthRecordKey).(*models.Record)

		userId := ""
		if success {
			userId = user.Id
		}
		trails, err := app.Dao().FindRecordsByFilter(
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
		return c.JSON(http.StatusOK, randomTrails)

	})

	e.Router.POST("/integration/strava/token", func(c echo.Context) error {
		encryptionKey := os.Getenv("POCKETBASE_ENCRYPTION_KEY")
		if len(encryptionKey) == 0 {
			return apis.NewBadRequestError("POCKETBASE_ENCRYPTION_KEY not set", nil)
		}

		var data strava.TokenRequest
		if err := c.Bind(&data); err != nil {
			return apis.NewBadRequestError("Failed to read request data", err)
		}

		user, success := c.Get(apis.ContextAuthRecordKey).(*models.Record)
		userId := ""
		if success {
			userId = user.Id
		}

		integrations, err := app.Dao().FindRecordsByExpr("integrations", dbx.NewExp("user = {:id}", dbx.Params{"id": userId}))
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
		err = app.Dao().SaveRecord(integration)
		if err != nil {
			return err
		}
		return c.JSON(http.StatusOK, nil)
	})

	e.Router.GET("/integration/komoot/login", func(c echo.Context) error {
		encryptionKey := os.Getenv("POCKETBASE_ENCRYPTION_KEY")
		if len(encryptionKey) == 0 {
			return apis.NewBadRequestError("POCKETBASE_ENCRYPTION_KEY not set", nil)
		}

		user, success := c.Get(apis.ContextAuthRecordKey).(*models.Record)
		userId := ""
		if success {
			userId = user.Id
		}

		integrations, err := app.Dao().FindRecordsByExpr("integrations", dbx.NewExp("user = {:id}", dbx.Params{"id": userId}))
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

		return c.JSON(http.StatusOK, nil)
	})
}

func registerCronJobs(app *pocketbase.PocketBase) {
	scheduler := cron.New()

	schedule := os.Getenv("POCKETBASE_CRON_SYNC_SCHEDULE")
	if len(schedule) == 0 {
		schedule = "0 2 * * *"
	}

	scheduler.MustAdd("integrations", schedule, func() {
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

	scheduler.Start()
}

func bootstrapData(app *pocketbase.PocketBase, client meilisearch.ServiceManager) error {
	bootstrapCategories(app)
	bootstrapMeilisearchTrails(app, client)
	return nil
}

func bootstrapCategories(app *pocketbase.PocketBase) error {
	query := app.Dao().RecordQuery("categories")
	records := []*models.Record{}

	if err := query.All(&records); err != nil {
		return err
	}
	if len(records) == 0 {
		collection, _ := app.Dao().FindCollectionByNameOrId("categories")

		categories := []string{"Hiking", "Walking", "Climbing", "Skiing", "Canoeing", "Biking"}
		for _, element := range categories {
			record := models.NewRecord(collection)
			form := forms.NewRecordUpsert(app, record)
			form.LoadData(map[string]any{
				"name": element,
			})
			f, _ := filesystem.NewFileFromPath("migrations/initial_data/" + strings.ToLower(element) + ".jpg")
			form.AddFiles("img", f)
			form.Submit()
		}
	}
	return nil
}

func bootstrapMeilisearchTrails(app *pocketbase.PocketBase, client meilisearch.ServiceManager) error {
	query := app.Dao().RecordQuery("trails")
	trails := []*models.Record{}

	if err := query.All(&trails); err != nil {
		return err
	}

	for _, trail := range trails {
		if err := util.UpdateTrail(trail, client); err != nil {
			return err
		}
	}
	return nil
}
