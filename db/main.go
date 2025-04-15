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
	"time"

	"github.com/meilisearch/meilisearch-go"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/plugins/migratecmd"
	"github.com/pocketbase/pocketbase/tools/filesystem"
	"github.com/pocketbase/pocketbase/tools/security"
	"github.com/spf13/cast"

	"pocketbase/integrations/komoot"
	"pocketbase/integrations/strava"

	_ "pocketbase/migrations"
	"pocketbase/util"
)

const defaultMeiliMasterKey = "vODkljPcfFANYNepCHyDyGjzAMPcdHnrb6X5KyXQPWo"

// verifySettings checks if the required environment variables are set.
// If they are not set, it logs a warning.
func verifySettings(app core.App) {
	encryptionKey := os.Getenv("POCKETBASE_ENCRYPTION_KEY")

	if len(encryptionKey) == 0 || len(encryptionKey) < 32 {
		app.Logger().Warn("POCKETBASE_ENCRYPTION_KEY not set or is shorter than 32 bytes")
	}

	meiliMasterKey := os.Getenv("MEILI_MASTER_KEY")

	if len(meiliMasterKey) == 0 || len(meiliMasterKey) < 32 {
		app.Logger().Warn("MEILI_MASTER_KEY not set or is shorter than 32 bytes")
	}

	if meiliMasterKey == defaultMeiliMasterKey {
		app.Logger().Warn("MEILI_MASTER_KEY is still set to the default value. Please change it to a secure value.")
	}
}

func main() {

	app := pocketbase.New()
	client := initializeMeiliSearch()

	verifySettings(app)

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
	app.OnRecordAfterCreateSuccess("users").BindFunc(createUserHandler(client))

	app.OnRecordAfterCreateSuccess("trails").BindFunc(createTrailHandler(client))
	app.OnRecordAfterUpdateSuccess("trails").BindFunc(updateTrailHandler(client))
	app.OnRecordAfterDeleteSuccess("trails").BindFunc(deleteTrailHandler(client))

	app.OnRecordAfterCreateSuccess("trail_share").BindFunc(createTrailShareHandler(client))
	app.OnRecordAfterDeleteSuccess("trail_share").BindFunc(deleteTrailShareHandler(client))

	app.OnRecordAfterCreateSuccess("lists").BindFunc(createListHandler(client))
	app.OnRecordAfterUpdateSuccess("lists").BindFunc(updateListHandler(client))
	app.OnRecordAfterDeleteSuccess("lists").BindFunc(deleteListHandler(client))

	app.OnRecordAfterCreateSuccess("list_share").BindFunc(createListShareHandler(client))
	app.OnRecordAfterDeleteSuccess("list_share").BindFunc(deleteListShareHandler(client))

	app.OnRecordAfterCreateSuccess("follows").BindFunc(createFollowHandler())
	app.OnRecordAfterCreateSuccess("comments").BindFunc(createCommentHandler())

	app.OnRecordsListRequest("integrations").BindFunc(listIntegrationHandler())
	app.OnRecordCreate("integrations").BindFunc(createIntegrationHandler())
	app.OnRecordAfterCreateSuccess("integrations").BindFunc(createUpdateIntegrationSuccessHandler())
	app.OnRecordUpdate("integrations").BindFunc(updateIntegrationHandler())
	app.OnRecordAfterUpdateSuccess("integrations").BindFunc(createUpdateIntegrationSuccessHandler())

	app.OnRecordRequestEmailChangeRequest("users").BindFunc(changeUserEmailHandler())
	app.OnServe().BindFunc(onBeforeServeHandler(client))

	app.OnBootstrap().BindFunc(onBootstrapHandler())
}

func createUserHandler(client meilisearch.ServiceManager) func(e *core.RecordEvent) error {
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
		if err := e.App.Save(e.Record); err != nil {
			return err
		}

		err = createDefaultUserSettings(e.App, e.Record.Id)
		if err != nil {
			return err
		}
		return e.Next()
	}
}

func createDefaultUserSettings(app core.App, userId string) error {
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

func createTrailHandler(client meilisearch.ServiceManager) func(e *core.RecordEvent) error {
	return func(e *core.RecordEvent) error {
		record := e.Record
		author, err := e.App.FindRecordById("users", record.GetString(("author")))
		if err != nil {
			return err
		}
		if err := util.IndexTrail(e.App, record, author, client); err != nil {
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
			err = util.SendNotificationToFollowers(e.App, notification)
			if err != nil {
				return err
			}
		}
		return e.Next()
	}
}

func updateTrailHandler(client meilisearch.ServiceManager) func(e *core.RecordEvent) error {
	return func(e *core.RecordEvent) error {
		record := e.Record
		author, err := e.App.FindRecordById("users", record.GetString(("author")))
		if err != nil {
			return err
		}
		err = util.UpdateTrail(e.App, record, author, client)
		if err != nil {
			return err
		}
		return e.Next()
	}
}

func deleteTrailHandler(client meilisearch.ServiceManager) func(e *core.RecordEvent) error {
	return func(e *core.RecordEvent) error {
		record := e.Record
		task, err := client.Index("trails").DeleteDocument(record.Id)
		if err != nil {
			return err
		}

		interval := 500 * time.Millisecond
		_, err = client.WaitForTask(task.TaskUID, interval)
		if err != nil {
			log.Fatalf("Error waiting for task completion: %v", err)
		}

		return e.Next()
	}
}

func createTrailShareHandler(client meilisearch.ServiceManager) func(e *core.RecordEvent) error {
	return func(e *core.RecordEvent) error {
		record := e.Record

		trailId := record.GetString("trail")
		shares, err := e.App.FindAllRecords("trail_share",
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

		if errs := e.App.ExpandRecord(record, []string{"trail", "trail.author"}, nil); len(errs) > 0 {
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
		err = util.SendNotification(e.App, notification, record.GetString("user"))
		if err != nil {
			return err
		}
		return e.Next()
	}
}

func deleteTrailShareHandler(client meilisearch.ServiceManager) func(e *core.RecordEvent) error {
	return func(e *core.RecordEvent) error {
		record := e.Record

		trailId := record.GetString("trail")
		err := util.UpdateTrailShares(trailId, []string{}, client)
		if err != nil {
			return err
		}
		return e.Next()
	}
}

func createListHandler(client meilisearch.ServiceManager) func(e *core.RecordEvent) error {
	return func(e *core.RecordEvent) error {
		record := e.Record

		if err := util.IndexList(record, client); err != nil {
			return err
		}
		if !record.GetBool("public") {
			return e.Next()
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
		err := util.SendNotificationToFollowers(e.App, notification)
		if err != nil {
			return err
		}
		return e.Next()
	}
}

func updateListHandler(client meilisearch.ServiceManager) func(e *core.RecordEvent) error {
	return func(e *core.RecordEvent) error {
		record := e.Record
		err := util.UpdateList(record, client)
		if err != nil {
			return err
		}

		return e.Next()
	}
}

func deleteListHandler(client meilisearch.ServiceManager) func(e *core.RecordEvent) error {
	return func(e *core.RecordEvent) error {
		record := e.Record
		_, err := client.Index("lists").DeleteDocument(record.Id)
		if err != nil {
			return err
		}

		return e.Next()
	}
}

func createListShareHandler(client meilisearch.ServiceManager) func(e *core.RecordEvent) error {
	return func(e *core.RecordEvent) error {
		record := e.Record
		listId := record.GetString("list")
		shares, err := e.App.FindAllRecords("list_share",
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

		if errs := e.App.ExpandRecord(record, []string{"list", "list.author"}, nil); len(errs) > 0 {
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
		err = util.SendNotification(e.App, notification, record.GetString("user"))
		if err != nil {
			return err
		}
		return e.Next()
	}
}

func deleteListShareHandler(client meilisearch.ServiceManager) func(e *core.RecordEvent) error {
	return func(e *core.RecordEvent) error {
		record := e.Record
		listId := record.GetString("list")
		err := util.UpdateListShares(listId, []string{}, client)
		if err != nil {
			return err
		}
		return e.Next()
	}
}

func createFollowHandler() func(e *core.RecordEvent) error {
	return func(e *core.RecordEvent) error {
		record := e.Record
		if errs := e.App.ExpandRecord(record, []string{"follower"}, nil); len(errs) > 0 {
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
		err := util.SendNotification(e.App, notification, record.GetString("followee"))
		if err != nil {
			return err
		}
		return e.Next()
	}
}

func createCommentHandler() func(e *core.RecordEvent) error {
	return func(e *core.RecordEvent) error {
		record := e.Record

		if errs := e.App.ExpandRecord(record, []string{"trail", "author"}, nil); len(errs) > 0 {
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
		err := util.SendNotification(e.App, notification, commentTrail.GetString("author"))
		if err != nil {
			return err
		}
		return e.Next()
	}
}

func listIntegrationHandler() func(e *core.RecordsListRequestEvent) error {
	return func(e *core.RecordsListRequestEvent) error {
		if e.HasSuperuserAuth() {
			return e.Next()
		}
		for _, r := range e.Records {

			err := censorIntegrationSecrets(r)
			if err != nil {
				return err
			}
		}

		return e.Next()
	}
}

func createIntegrationHandler() func(e *core.RecordEvent) error {
	return func(e *core.RecordEvent) error {
		err := encryptIntegrationSecrets(e.App, e.Record)
		if err != nil {
			return err
		}

		return e.Next()
	}
}

func createUpdateIntegrationSuccessHandler() func(e *core.RecordEvent) error {
	return func(e *core.RecordEvent) error {
		err := censorIntegrationSecrets(e.Record)
		if err != nil {
			return err
		}
		return e.Next()
	}
}

func updateIntegrationHandler() func(e *core.RecordEvent) error {
	return func(e *core.RecordEvent) error {
		err := encryptIntegrationSecrets(e.App, e.Record)
		if err != nil {
			return err
		}

		return e.Next()
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
			if integration == nil {
				continue
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

func encryptIntegrationSecrets(app core.App, r *core.Record) error {
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
				// If the secret is already encrypted, we don't re-encrypt it.
				// TODO: This is a bit of a hack, we should handle this in a more robust way (e.g.
				// storing flag on the record or prefixing encrypted strings with enc: or smilar).
				// Doing that would also potentially allow us to support key rotation in the future.
				if secret, ok := integration[secretKey].(string); ok && len(secret) > 0 && !util.CanDecryptSecret(secret) {
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
					if integration == nil {
						continue
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

func changeUserEmailHandler() func(e *core.RecordRequestEmailChangeRequestEvent) error {
	return func(e *core.RecordRequestEmailChangeRequestEvent) error {

		e.Record.Set("email", e.NewEmail)
		if err := e.App.Save(e.Record); err != nil {
			return err
		}
		return nil
	}
}

func onBeforeServeHandler(client meilisearch.ServiceManager) func(se *core.ServeEvent) error {
	return func(se *core.ServeEvent) error {
		registerRoutes(se, client)
		registerCronJobs(se.App)
		bootstrapData(se.App, client)

		return se.Next()
	}

}

func onBootstrapHandler() func(se *core.BootstrapEvent) error {
	return func(e *core.BootstrapEvent) error {
		if err := e.Next(); err != nil {
			return err
		}

		if v := os.Getenv("POCKETBASE_SMTP_SENDER_ADRESS"); v != "" {
			e.App.Settings().Meta.SenderAddress = v
		}
		if v := os.Getenv("POCKETBASE_SMTP_SENDER_NAME"); v != "" {
			e.App.Settings().Meta.SenderName = v
		}
		if v := os.Getenv("POCKETBASE_SMTP_ENABLED"); v != "" {
			e.App.Settings().SMTP.Enabled = cast.ToBool(v)
		}
		if v := os.Getenv("POCKETBASE_SMTP_HOST"); v != "" {
			e.App.Settings().SMTP.Host = v
		}
		if v := os.Getenv("POCKETBASE_SMTP_PORT"); v != "" {
			e.App.Settings().SMTP.Port = cast.ToInt(v)
		}
		if v := os.Getenv("POCKETBASE_SMTP_USERNAME"); v != "" {
			e.App.Settings().SMTP.Username = v
		}
		if v := os.Getenv("POCKETBASE_SMTP_PASSWORD"); v != "" {
			e.App.Settings().SMTP.Password = v
		}

		return e.App.Save(e.App.Settings())
	}
}

func registerRoutes(se *core.ServeEvent, client meilisearch.ServiceManager) {
	se.Router.GET("/public/search/token", func(e *core.RequestEvent) error {
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
	se.Router.GET("/trail/recommend", func(e *core.RequestEvent) error {
		qSize := e.Request.URL.Query().Get("size")
		size, err := strconv.Atoi(qSize)
		if err != nil {
			size = 4
		}

		userId := ""
		if e.Auth != nil {
			userId = e.Auth.Id
		}

		trails, err := e.App.FindRecordsByFilter(
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

	se.Router.POST("/integration/strava/token", func(e *core.RequestEvent) error {
		encryptionKey := os.Getenv("POCKETBASE_ENCRYPTION_KEY")
		if len(encryptionKey) == 0 {
			return apis.NewBadRequestError("POCKETBASE_ENCRYPTION_KEY not set", nil)
		}

		var data strava.TokenRequest
		if err := e.BindBody(&data); err != nil {
			return apis.NewBadRequestError("Failed to read request data", err)
		}

		userId := ""
		if e.Auth != nil {
			userId = e.Auth.Id
		}

		integrations, err := e.App.FindAllRecords("integrations", dbx.NewExp("user = {:id}", dbx.Params{"id": userId}))
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
		err = e.App.Save(integration)
		if err != nil {
			return err
		}
		return e.JSON(http.StatusOK, nil)
	})

	se.Router.GET("/integration/komoot/login", func(e *core.RequestEvent) error {
		encryptionKey := os.Getenv("POCKETBASE_ENCRYPTION_KEY")
		if len(encryptionKey) == 0 {
			return apis.NewBadRequestError("POCKETBASE_ENCRYPTION_KEY not set", nil)
		}

		userId := ""
		if e.Auth != nil {
			userId = e.Auth.Id
		}

		integrations, err := e.App.FindAllRecords("integrations", dbx.NewExp("user = {:id}", dbx.Params{"id": userId}))
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

func registerCronJobs(app core.App) {
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

func bootstrapData(app core.App, client meilisearch.ServiceManager) error {
	bootstrapCategories(app)
	bootstrapMeilisearchTrails(app, client)
	return nil
}

func bootstrapCategories(app core.App) error {
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

func bootstrapMeilisearchTrails(app core.App, client meilisearch.ServiceManager) error {
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
		if err := util.IndexTrail(app, trail, author, client); err != nil {
			return err
		}

		shares, err := app.FindAllRecords("trail_share",
			dbx.NewExp("trail = {:trailId}", dbx.Params{"trailId": trail.Id}),
		)
		if err != nil {
			return err
		}
		userIds := make([]string, len(shares))
		for i, r := range shares {
			userIds[i] = r.GetString("user")
		}
		err = util.UpdateTrailShares(trail.Id, userIds, client)

		if err != nil {
			return err
		}
	}
	return nil
}
