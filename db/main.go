package main

import (
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/labstack/echo/v5"
	"github.com/meilisearch/meilisearch-go"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase"
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

	migratecmd.MustRegister(app, app.RootCmd, migratecmd.Config{
		Dir:         "migrations",
		Automigrate: true,
	})

	client := meilisearch.NewClient(meilisearch.ClientConfig{
		Host:   os.Getenv("MEILI_URL"),
		APIKey: os.Getenv("MEILI_MASTER_KEY"),
	})

	app.OnModelAfterCreate("users").Add(func(e *core.ModelEvent) error {
		record := e.Model.(*models.Record)
		userId := record.GetId()

		searchRules := map[string]interface{}{
			"cities500": map[string]string{},
			"trails": map[string]string{
				"filter": "public = true OR author = " + userId + " OR shares = " + userId,
			},
		}

		if token, err := util.GenerateMeilisearchToken(searchRules, client); err != nil {
			return err
		} else {
			record.Set("token", token)

			if err := app.Dao().SaveRecord(record); err != nil {
				return err
			}
		}

		collection, err := app.Dao().FindCollectionByNameOrId("settings")
		if err != nil {
			return err
		}

		settings := models.NewRecord(collection)

		settings.Set("language", "en")
		settings.Set("unit", "metric")
		settings.Set("mapFocus", "trails")
		settings.Set("user", record.Id)

		if err := app.Dao().SaveRecord(settings); err != nil {
			return err
		}

		return nil
	})

	app.OnRecordAfterCreateRequest("trails").Add(func(e *core.RecordCreateEvent) error {
		if err := util.IndexTrail(e.Record, client); err != nil {
			return err
		}
		return nil
	})

	app.OnRecordAfterUpdateRequest("trails").Add(func(e *core.RecordUpdateEvent) error {
		if err := util.UpdateTrail(e.Record, client); err != nil {
			return err
		}
		return nil
	})

	app.OnRecordAfterDeleteRequest("trails").Add(func(e *core.RecordDeleteEvent) error {
		if _, err := client.Index("trails").DeleteDocument(e.Record.Id); err != nil {
			return err
		}
		return nil
	})

	app.OnRecordAfterCreateRequest("trail_share").Add(func(e *core.RecordCreateEvent) error {
		trailId := e.Record.GetString("trail")
		shares, err := app.Dao().FindRecordsByExpr("trail_share",
			dbx.NewExp("trail = {:trailId}", dbx.Params{"trailId": trailId}),
		)
		if err != nil {
			return err
		}
		userIds := []string{}
		for _, r := range shares {
			userIds = append(userIds, r.GetString("user"))
		}

		if err := util.UpdateTrailShares(trailId, userIds, client); err != nil {
			return err
		}
		return nil
	})

	app.OnRecordAfterDeleteRequest("trail_share").Add(func(e *core.RecordDeleteEvent) error {
		trailId := e.Record.GetString("trail")

		if err := util.UpdateTrailShares(trailId, []string{}, client); err != nil {
			return err
		}
		return nil
	})

	app.OnRecordBeforeRequestEmailChangeRequest("users").Add(func(e *core.RecordRequestEmailChangeEvent) error {
		form := forms.NewRecordEmailChangeRequest(app, e.Record)
		if err := e.HttpContext.Bind(form); err != nil {
			return err
		}
		e.Record.Set("email", form.NewEmail)

		if err := app.Dao().SaveRecord(e.Record); err != nil {
			return err
		}

		return hook.StopPropagation
	})

	app.OnBeforeServe().Add(func(e *core.ServeEvent) error {
		e.Router.GET("/public/search/token", func(c echo.Context) error {
			searchRules := map[string]interface{}{
				"cities500": map[string]string{},
				"trails": map[string]string{
					"filter": "public = true",
				},
			}

			if token, err := util.GenerateMeilisearchToken(searchRules, client); err != nil {
				return err
			} else {
				return c.JSON(http.StatusOK, map[string]string{"token": token})
			}

		})

		// bootstrap pocketbase
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

		// bootstrap meilisearch
		query = app.Dao().RecordQuery("trails")
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
	})

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
