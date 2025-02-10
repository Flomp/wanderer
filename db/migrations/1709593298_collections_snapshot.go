package migrations

import (
	"encoding/json"
	"os"

	"github.com/meilisearch/meilisearch-go"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/daos"
	m "github.com/pocketbase/pocketbase/migrations"
	"github.com/pocketbase/pocketbase/models"
)

func init() {
	client := meilisearch.New(os.Getenv("MEILI_URL"), meilisearch.WithAPIKey(os.Getenv("MEILI_MASTER_KEY")))

	m.Register(func(db dbx.Builder) error {
		_, err := client.CreateIndex(&meilisearch.IndexConfig{
			Uid:        "trails",
			PrimaryKey: "id",
		})
		if err != nil {
			return err
		}

		jsonData := `[
			{
				"id": "_pb_users_auth_",
				"created": "2024-01-26 15:05:23.262Z",
				"updated": "2024-02-29 16:52:40.743Z",
				"name": "users",
				"type": "auth",
				"system": false,
				"schema": [
					{
						"system": false,
						"id": "dlzhxcn2",
						"name": "token",
						"type": "text",
						"required": false,
						"presentable": false,
						"unique": false,
						"options": {
							"min": null,
							"max": null,
							"pattern": ""
						}
					},
					{
						"system": false,
						"id": "users_avatar",
						"name": "avatar",
						"type": "file",
						"required": false,
						"presentable": false,
						"unique": false,
						"options": {
							"mimeTypes": [
								"image/jpeg",
								"image/png",
								"image/svg+xml",
								"image/gif",
								"image/webp"
							],
							"thumbs": null,
							"maxSelect": 1,
							"maxSize": 5242880,
							"protected": false
						}
					},
					{
						"system": false,
						"id": "wjofulpg",
						"name": "unit",
						"type": "select",
						"required": false,
						"presentable": false,
						"unique": false,
						"options": {
							"maxSelect": 1,
							"values": [
								"metric",
								"imperial"
							]
						}
					},
					{
						"system": false,
						"id": "t1wlsqyp",
						"name": "language",
						"type": "select",
						"required": false,
						"presentable": false,
						"unique": false,
						"options": {
							"maxSelect": 1,
							"values": [
								"en",
								"de"
							]
						}
					},
					{
						"system": false,
						"id": "wosrk4ue",
						"name": "location",
						"type": "json",
						"required": false,
						"presentable": false,
						"unique": false,
						"options": {
							"maxSize": 2000000
						}
					}
				],
				"indexes": [],
				"listRule": "id = @request.auth.id",
				"viewRule": "id = @request.auth.id",
				"createRule": "",
				"updateRule": "id = @request.auth.id",
				"deleteRule": "id = @request.auth.id",
				"options": {
					"allowEmailAuth": true,
					"allowOAuth2Auth": true,
					"allowUsernameAuth": true,
					"exceptEmailDomains": null,
					"manageRule": null,
					"minPasswordLength": 8,
					"onlyEmailDomains": null,
					"onlyVerified": false,
					"requireEmail": false
				}
			},
			{
				"id": "e864strfxo14pm4",
				"created": "2024-01-26 18:46:35.843Z",
				"updated": "2024-02-23 12:47:30.327Z",
				"name": "trails",
				"type": "base",
				"system": false,
				"schema": [
					{
						"system": false,
						"id": "wquvuytd",
						"name": "name",
						"type": "text",
						"required": true,
						"presentable": false,
						"unique": false,
						"options": {
							"min": null,
							"max": null,
							"pattern": ""
						}
					},
					{
						"system": false,
						"id": "6kkucam1",
						"name": "description",
						"type": "text",
						"required": false,
						"presentable": false,
						"unique": false,
						"options": {
							"min": null,
							"max": null,
							"pattern": ""
						}
					},
					{
						"system": false,
						"id": "8x74ba26",
						"name": "location",
						"type": "text",
						"required": false,
						"presentable": false,
						"unique": false,
						"options": {
							"min": null,
							"max": null,
							"pattern": ""
						}
					},
					{
						"system": false,
						"id": "ehrmydva",
						"name": "public",
						"type": "bool",
						"required": false,
						"presentable": false,
						"unique": false,
						"options": {}
					},
					{
						"system": false,
						"id": "epgmtyxy",
						"name": "distance",
						"type": "number",
						"required": false,
						"presentable": false,
						"unique": false,
						"options": {
							"min": 0,
							"max": null,
							"noDecimal": false
						}
					},
					{
						"system": false,
						"id": "5wxdt3aj",
						"name": "elevation_gain",
						"type": "number",
						"required": false,
						"presentable": false,
						"unique": false,
						"options": {
							"min": null,
							"max": null,
							"noDecimal": false
						}
					},
					{
						"system": false,
						"id": "ukr9rqz4",
						"name": "duration",
						"type": "number",
						"required": false,
						"presentable": false,
						"unique": false,
						"options": {
							"min": 0,
							"max": null,
							"noDecimal": false
						}
					},
					{
						"system": false,
						"id": "eqeqja1s",
						"name": "lat",
						"type": "number",
						"required": false,
						"presentable": false,
						"unique": false,
						"options": {
							"min": null,
							"max": null,
							"noDecimal": false
						}
					},
					{
						"system": false,
						"id": "y6dbfyw6",
						"name": "lon",
						"type": "number",
						"required": false,
						"presentable": false,
						"unique": false,
						"options": {
							"min": null,
							"max": null,
							"noDecimal": false
						}
					},
					{
						"system": false,
						"id": "mcqce8l9",
						"name": "thumbnail",
						"type": "text",
						"required": false,
						"presentable": false,
						"unique": false,
						"options": {
							"min": null,
							"max": null,
							"pattern": ""
						}
					},
					{
						"system": false,
						"id": "aqbpyawe",
						"name": "photos",
						"type": "file",
						"required": false,
						"presentable": false,
						"unique": false,
						"options": {
							"mimeTypes": [
								"image/jpeg",
								"image/vnd.mozilla.apng",
								"image/png",
								"image/webp",
								"image/svg+xml"
							],
							"thumbs": [],
							"maxSelect": 99,
							"maxSize": 5242880,
							"protected": false
						}
					},
					{
						"system": false,
						"id": "k8xdrsyv",
						"name": "gpx",
						"type": "file",
						"required": false,
						"presentable": false,
						"unique": false,
						"options": {
							"mimeTypes": [],
							"thumbs": [],
							"maxSelect": 1,
							"maxSize": 5242880,
							"protected": false
						}
					},
					{
						"system": false,
						"id": "b49obm5u",
						"name": "category",
						"type": "relation",
						"required": false,
						"presentable": false,
						"unique": false,
						"options": {
							"collectionId": "kjxvi8asj2igqwf",
							"cascadeDelete": false,
							"minSelect": null,
							"maxSelect": 1,
							"displayFields": null
						}
					},
					{
						"system": false,
						"id": "1utgul91",
						"name": "author",
						"type": "relation",
						"required": true,
						"presentable": false,
						"unique": false,
						"options": {
							"collectionId": "_pb_users_auth_",
							"cascadeDelete": true,
							"minSelect": null,
							"maxSelect": 1,
							"displayFields": null
						}
					},
					{
						"system": false,
						"id": "e1lwowvd",
						"name": "summit_logs",
						"type": "relation",
						"required": false,
						"presentable": false,
						"unique": false,
						"options": {
							"collectionId": "dd2l9a4vxpy2ni8",
							"cascadeDelete": false,
							"minSelect": null,
							"maxSelect": null,
							"displayFields": null
						}
					},
					{
						"system": false,
						"id": "ppq2sist",
						"name": "waypoints",
						"type": "relation",
						"required": false,
						"presentable": false,
						"unique": false,
						"options": {
							"collectionId": "goeo2ubp103rzp9",
							"cascadeDelete": false,
							"minSelect": null,
							"maxSelect": null,
							"displayFields": null
						}
					}
				],
				"indexes": [],
				"listRule": "author = @request.auth.id || public = true",
				"viewRule": "author = @request.auth.id || public = true",
				"createRule": "@request.auth.id != \"\" && (@request.data.author = @request.auth.id)",
				"updateRule": "author = @request.auth.id ",
				"deleteRule": "author = @request.auth.id ",
				"options": {}
			},
			{
				"id": "kjxvi8asj2igqwf",
				"created": "2024-01-26 18:47:30.797Z",
				"updated": "2024-01-27 09:38:36.240Z",
				"name": "categories",
				"type": "base",
				"system": false,
				"schema": [
					{
						"system": false,
						"id": "38tbd8u4",
						"name": "name",
						"type": "text",
						"required": false,
						"presentable": false,
						"unique": false,
						"options": {
							"min": null,
							"max": null,
							"pattern": ""
						}
					},
					{
						"system": false,
						"id": "64dsnxtb",
						"name": "img",
						"type": "file",
						"required": false,
						"presentable": false,
						"unique": false,
						"options": {
							"mimeTypes": [],
							"thumbs": [],
							"maxSelect": 1,
							"maxSize": 5242880,
							"protected": false
						}
					}
				],
				"indexes": [],
				"listRule": "",
				"viewRule": "",
				"createRule": null,
				"updateRule": null,
				"deleteRule": null,
				"options": {}
			},
			{
				"id": "goeo2ubp103rzp9",
				"created": "2024-01-26 18:50:13.695Z",
				"updated": "2024-02-04 11:49:30.732Z",
				"name": "waypoints",
				"type": "base",
				"system": false,
				"schema": [
					{
						"system": false,
						"id": "2yegzjtk",
						"name": "name",
						"type": "text",
						"required": false,
						"presentable": false,
						"unique": false,
						"options": {
							"min": null,
							"max": null,
							"pattern": ""
						}
					},
					{
						"system": false,
						"id": "3xtcjtxv",
						"name": "description",
						"type": "text",
						"required": false,
						"presentable": false,
						"unique": false,
						"options": {
							"min": null,
							"max": null,
							"pattern": ""
						}
					},
					{
						"system": false,
						"id": "ygotgxzy",
						"name": "lat",
						"type": "number",
						"required": false,
						"presentable": false,
						"unique": false,
						"options": {
							"min": null,
							"max": null,
							"noDecimal": false
						}
					},
					{
						"system": false,
						"id": "q0ygnxd2",
						"name": "lon",
						"type": "number",
						"required": false,
						"presentable": false,
						"unique": false,
						"options": {
							"min": null,
							"max": null,
							"noDecimal": false
						}
					},
					{
						"system": false,
						"id": "rnjgm2tk",
						"name": "icon",
						"type": "text",
						"required": false,
						"presentable": false,
						"unique": false,
						"options": {
							"min": null,
							"max": null,
							"pattern": ""
						}
					}
				],
				"indexes": [],
				"listRule": "@request.auth.id != \"\" && @collection.trails.waypoints.id ?= id && (@collection.trails.author = @request.auth.id || @collection.trails.public = true)",
				"viewRule": "@request.auth.id != \"\" && @collection.trails.waypoints.id ?= id && (@collection.trails.author = @request.auth.id || @collection.trails.public = true)",
				"createRule": "@request.auth.id != \"\"",
				"updateRule": "@request.auth.id != \"\" && @collection.trails.waypoints.id ?= id && (@collection.trails.author = @request.auth.id)",
				"deleteRule": "@request.auth.id != \"\" && @collection.trails.waypoints.id ?= id && (@collection.trails.author = @request.auth.id)",
				"options": {}
			},
			{
				"id": "dd2l9a4vxpy2ni8",
				"created": "2024-01-27 12:23:48.069Z",
				"updated": "2024-02-02 21:24:15.600Z",
				"name": "summit_logs",
				"type": "base",
				"system": false,
				"schema": [
					{
						"system": false,
						"id": "gxq1yeld",
						"name": "date",
						"type": "date",
						"required": false,
						"presentable": false,
						"unique": false,
						"options": {
							"min": "",
							"max": ""
						}
					},
					{
						"system": false,
						"id": "0ykzwuia",
						"name": "text",
						"type": "text",
						"required": false,
						"presentable": false,
						"unique": false,
						"options": {
							"min": null,
							"max": null,
							"pattern": ""
						}
					}
				],
				"indexes": [],
				"listRule": "@request.auth.id != \"\" && @collection.trails.summit_logs.id ?= id && (@collection.trails.author = @request.auth.id || @collection.trails.public = true)",
				"viewRule": "@request.auth.id != \"\" && @collection.trails.summit_logs.id ?= id && (@collection.trails.author = @request.auth.id || @collection.trails.public = true)",
				"createRule": "@request.auth.id != \"\"",
				"updateRule": "@request.auth.id != \"\" && @collection.trails.summit_logs.id ?= id && (@collection.trails.author = @request.auth.id)",
				"deleteRule": "@request.auth.id != \"\" && @collection.trails.summit_logs.id ?= id && (@collection.trails.author = @request.auth.id)",
				"options": {}
			},
			{
				"id": "r6gu2ajyidy1x69",
				"created": "2024-02-27 11:46:07.367Z",
				"updated": "2024-02-27 13:22:52.195Z",
				"name": "lists",
				"type": "base",
				"system": false,
				"schema": [
					{
						"system": false,
						"id": "0goi1ipa",
						"name": "name",
						"type": "text",
						"required": true,
						"presentable": false,
						"unique": false,
						"options": {
							"min": 1,
							"max": null,
							"pattern": ""
						}
					},
					{
						"system": false,
						"id": "3v3pwnqn",
						"name": "description",
						"type": "text",
						"required": false,
						"presentable": false,
						"unique": false,
						"options": {
							"min": null,
							"max": null,
							"pattern": ""
						}
					},
					{
						"system": false,
						"id": "decatj0h",
						"name": "avatar",
						"type": "file",
						"required": false,
						"presentable": false,
						"unique": false,
						"options": {
							"mimeTypes": [
								"image/png",
								"image/vnd.mozilla.apng",
								"image/jpeg",
								"image/webp",
								"image/svg+xml"
							],
							"thumbs": [],
							"maxSelect": 1,
							"maxSize": 5242880,
							"protected": false
						}
					},
					{
						"system": false,
						"id": "jqtcrcnq",
						"name": "trails",
						"type": "relation",
						"required": false,
						"presentable": false,
						"unique": false,
						"options": {
							"collectionId": "e864strfxo14pm4",
							"cascadeDelete": false,
							"minSelect": null,
							"maxSelect": null,
							"displayFields": null
						}
					},
					{
						"system": false,
						"id": "kwm6zdet",
						"name": "author",
						"type": "relation",
						"required": false,
						"presentable": false,
						"unique": false,
						"options": {
							"collectionId": "_pb_users_auth_",
							"cascadeDelete": false,
							"minSelect": null,
							"maxSelect": 1,
							"displayFields": null
						}
					}
				],
				"indexes": [],
				"listRule": "author = @request.auth.id ",
				"viewRule": "author = @request.auth.id ",
				"createRule": "author = @request.auth.id ",
				"updateRule": "author = @request.auth.id ",
				"deleteRule": "author = @request.auth.id ",
				"options": {}
			}
		]`

		collections := []*models.Collection{}
		if err := json.Unmarshal([]byte(jsonData), &collections); err != nil {
			return err
		}

		return daos.New(db).ImportCollections(collections, true, nil)
	}, func(db dbx.Builder) error {
		return nil
	})
}
