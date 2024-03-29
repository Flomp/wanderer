package main

import (
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/labstack/echo/v5"
	"github.com/meilisearch/meilisearch-go"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/forms"
	"github.com/pocketbase/pocketbase/models"
	"github.com/pocketbase/pocketbase/plugins/migratecmd"
	"github.com/pocketbase/pocketbase/tools/filesystem"

	_ "pocketbase/migrations"
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
				"filter": "public = true OR author = " + userId,
			},
		}

		if token, err := generateMeilisearchToken(searchRules, client); err != nil {
			return err
		} else {
			record.Set("token", token)

			if err := app.Dao().SaveRecord(record); err != nil {
				return err
			}
		}

		return nil
	})

	app.OnRecordAfterCreateRequest("trails").Add(func(e *core.RecordCreateEvent) error {
		if err := indexRecord(e.Record, client); err != nil {
			return err
		}
		return nil
	})

	app.OnRecordAfterUpdateRequest("trails").Add(func(e *core.RecordUpdateEvent) error {
		if err := indexRecord(e.Record, client); err != nil {
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

	app.OnBeforeServe().Add(func(e *core.ServeEvent) error {
		e.Router.GET("/public/search/token", func(c echo.Context) error {
			searchRules := map[string]interface{}{
				"cities500": map[string]string{},
				"trails": map[string]string{
					"filter": "public = true",
				},
			}

			if token, err := generateMeilisearchToken(searchRules, client); err != nil {
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

			if err := indexRecord(trail, client); err != nil {
				return err
			}
		}

		return nil
	})

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
