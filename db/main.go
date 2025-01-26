package main

import (
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
	"github.com/pocketbase/pocketbase/tools/filesystem"
	"github.com/pocketbase/pocketbase/tools/hook"

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

	app.OnRecordAfterCreateRequest("trails").Add(createTrailHandler(app, client))
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

func createTrailHandler(app *pocketbase.PocketBase, client meilisearch.ServiceManager) func(e *core.RecordCreateEvent) error {
	return func(e *core.RecordCreateEvent) error {
		if err := util.IndexTrail(e.Record, client); err != nil {
			return err
		}
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
		log.Println(trail)

		if err := util.UpdateTrail(trail, client); err != nil {
			return err
		}
	}
	return nil
}
