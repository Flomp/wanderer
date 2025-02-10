package migrations

import (
	"encoding/json"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/daos"
	m "github.com/pocketbase/pocketbase/migrations"
	"github.com/pocketbase/pocketbase/models"
)

func init() {
	m.Register(func(db dbx.Builder) error {
		jsonData := `[
			{
				"id": "e864strfxo14pm4",
				"created": "2024-01-26 18:46:35.843Z",
				"updated": "2024-11-09 13:21:52.381Z",
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
						"id": "xutbwpq4",
						"name": "elevation_loss",
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
								"image/svg+xml",
								"image/heic"
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
					},
					{
						"system": false,
						"id": "k2giqyjq",
						"name": "thumbnail",
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
						"id": "dywtnynw",
						"name": "difficulty",
						"type": "select",
						"required": false,
						"presentable": false,
						"unique": false,
						"options": {
							"maxSelect": 1,
							"values": [
								"easy",
								"moderate",
								"difficult"
							]
						}
					},
					{
						"system": false,
						"id": "hovyvbtt",
						"name": "date",
						"type": "date",
						"required": false,
						"presentable": false,
						"unique": false,
						"options": {
							"min": "",
							"max": ""
						}
					}
				],
				"indexes": [],
				"listRule": "author = @request.auth.id || public = true || (@request.auth.id != \"\" && trail_share_via_trail.user ?= @request.auth.id)",
				"viewRule": "author = @request.auth.id || public = true || (@request.auth.id != \"\" && trail_share_via_trail.user ?= @request.auth.id)",
				"createRule": "@request.auth.id != \"\" && (@request.data.author = @request.auth.id)",
				"updateRule": "author = @request.auth.id || (@request.auth.id != \"\" && trail_share_via_trail.trail = id && trail_share_via_trail.user ?= @request.auth.id && trail_share_via_trail.permission = \"edit\")",
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
				"updated": "2024-11-16 13:33:40.896Z",
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
					},
					{
						"system": false,
						"id": "tfhs3juh",
						"name": "photos",
						"type": "file",
						"required": false,
						"presentable": false,
						"unique": false,
						"options": {
							"mimeTypes": [
								"image/jpeg",
								"image/png",
								"image/vnd.mozilla.apng",
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
						"id": "8qbxrsd8",
						"name": "author",
						"type": "relation",
						"required": true,
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
				"listRule": "(@request.auth.id != \"\" && @collection.trails.waypoints.id ?= id && @collection.trails.author ?= @request.auth.id) || (@collection.trails.waypoints.id ?= id && @collection.trails.public ?= true) || \n(@collection.trail_share.trail.waypoints.id ?= id && @collection.trail_share.user ?= @request.auth.id) ||\n(author = @request.auth.id)",
				"viewRule": "(@request.auth.id != \"\" && @collection.trails.waypoints.id ?= id && @collection.trails.author ?= @request.auth.id) || (@collection.trails.waypoints.id ?= id && @collection.trails.public ?= true) || \n(@collection.trail_share.trail.waypoints.id ?= id && @collection.trail_share.user ?= @request.auth.id) ||\n(author = @request.auth.id)",
				"createRule": "@request.auth.id != \"\"",
				"updateRule": "@request.auth.id != \"\" && ((@collection.trails.waypoints.id ?= id && @collection.trails.author = @request.auth.id) || author = @request.auth.id)",
				"deleteRule": "@request.auth.id != \"\" && ((@collection.trails.waypoints.id ?= id && @collection.trails.author = @request.auth.id) || author = @request.auth.id)",
				"options": {}
			},
			{
				"id": "dd2l9a4vxpy2ni8",
				"created": "2024-01-27 12:23:48.069Z",
				"updated": "2024-11-30 16:45:46.575Z",
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
					},
					{
						"system": false,
						"id": "rfwmdcpt",
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
						"id": "ixnksbkt",
						"name": "photos",
						"type": "file",
						"required": false,
						"presentable": false,
						"unique": false,
						"options": {
							"mimeTypes": [
								"image/jpeg",
								"image/png",
								"image/vnd.mozilla.apng",
								"image/webp",
								"image/svg+xml",
								"image/heic"
							],
							"thumbs": [],
							"maxSelect": 99,
							"maxSize": 5242880,
							"protected": false
						}
					},
					{
						"system": false,
						"id": "jovws28m",
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
						"id": "m2kndtwn",
						"name": "elevation_gain",
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
						"id": "uqqo9cws",
						"name": "elevation_loss",
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
						"id": "vwxjsrae",
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
						"id": "r0mj3tkr",
						"name": "author",
						"type": "relation",
						"required": true,
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
				"listRule": "(@request.auth.id != \"\" && @collection.trails.summit_logs.id ?= id && @collection.trails.author ?= @request.auth.id) || (@collection.trails.summit_logs.id ?= id && @collection.trails.public ?= true) ||\n(@collection.trail_share.trail.summit_logs.id ?= id && @collection.trail_share.user ?= @request.auth.id) ||\n(author = @request.auth.id)",
				"viewRule": "(@request.auth.id != \"\" && @collection.trails.summit_logs.id ?= id && @collection.trails.author ?= @request.auth.id) || (@collection.trails.summit_logs.id ?= id && @collection.trails.public ?= true) || \n(@collection.trail_share.trail.summit_logs.id ?= id && @collection.trail_share.user ?= @request.auth.id) ||\n(author = @request.auth.id)",
				"createRule": "@request.auth.id != \"\"",
				"updateRule": "@request.auth.id != \"\" && @collection.trails.summit_logs.id ?= id && (@collection.trails.author ?= @request.auth.id) || (author = @request.auth.id)",
				"deleteRule": "@request.auth.id != \"\" && @collection.trails.summit_logs.id ?= id && (@collection.trails.author ?= @request.auth.id) || (author = @request.auth.id)",
				"options": {}
			},
			{
				"id": "r6gu2ajyidy1x69",
				"created": "2024-02-27 11:46:07.367Z",
				"updated": "2024-12-14 13:04:30.337Z",
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
					},
					{
						"system": false,
						"id": "rolk3q3j",
						"name": "public",
						"type": "bool",
						"required": false,
						"presentable": false,
						"unique": false,
						"options": {}
					}
				],
				"indexes": [],
				"listRule": "author = @request.auth.id || public = true || (@request.auth.id != \"\" && list_share_via_list.user ?= @request.auth.id)",
				"viewRule": "author = @request.auth.id || public = true || (@request.auth.id != \"\" && list_share_via_list.user ?= @request.auth.id)",
				"createRule": "@request.auth.id != \"\" && (@request.data.author = @request.auth.id)",
				"updateRule": "author = @request.auth.id || (@request.auth.id != \"\" && list_share_via_list.list = id && list_share_via_list.user ?= @request.auth.id && list_share_via_list.permission = \"edit\")",
				"deleteRule": "author = @request.auth.id ",
				"options": {}
			},
			{
				"id": "4wbv9tz5zjdrjh1",
				"created": "2024-04-01 17:47:10.015Z",
				"updated": "2024-11-09 13:06:05.825Z",
				"name": "trails_filter",
				"type": "view",
				"system": false,
				"schema": [
					{
						"system": false,
						"id": "oyvx8oaq",
						"name": "max_distance",
						"type": "json",
						"required": false,
						"presentable": false,
						"unique": false,
						"options": {
							"maxSize": 1
						}
					},
					{
						"system": false,
						"id": "4xu7voeh",
						"name": "max_elevation_gain",
						"type": "json",
						"required": false,
						"presentable": false,
						"unique": false,
						"options": {
							"maxSize": 1
						}
					},
					{
						"system": false,
						"id": "sqzilovv",
						"name": "max_elevation_loss",
						"type": "json",
						"required": false,
						"presentable": false,
						"unique": false,
						"options": {
							"maxSize": 1
						}
					},
					{
						"system": false,
						"id": "rupidtew",
						"name": "max_duration",
						"type": "json",
						"required": false,
						"presentable": false,
						"unique": false,
						"options": {
							"maxSize": 1
						}
					},
					{
						"system": false,
						"id": "2y7dbhf9",
						"name": "min_distance",
						"type": "json",
						"required": false,
						"presentable": false,
						"unique": false,
						"options": {
							"maxSize": 1
						}
					},
					{
						"system": false,
						"id": "ooiwexpc",
						"name": "min_elevation_gain",
						"type": "json",
						"required": false,
						"presentable": false,
						"unique": false,
						"options": {
							"maxSize": 1
						}
					},
					{
						"system": false,
						"id": "ondmfwce",
						"name": "min_elevation_loss",
						"type": "json",
						"required": false,
						"presentable": false,
						"unique": false,
						"options": {
							"maxSize": 1
						}
					},
					{
						"system": false,
						"id": "kcrhzmaw",
						"name": "min_duration",
						"type": "json",
						"required": false,
						"presentable": false,
						"unique": false,
						"options": {
							"maxSize": 1
						}
					}
				],
				"indexes": [],
				"listRule": null,
				"viewRule": "@request.auth.id = id",
				"createRule": null,
				"updateRule": null,
				"deleteRule": null,
				"options": {
					"query": "SELECT users.id, COALESCE(MAX(trails.distance), 0) AS max_distance,\n  COALESCE(MAX(trails.elevation_gain), 0) AS max_elevation_gain, \n  COALESCE(MAX(trails.elevation_loss), 0) AS max_elevation_loss, \n  COALESCE(MAX(trails.duration), 0) AS max_duration, \n  COALESCE(MIN(trails.distance), 0) AS min_distance,   \n  COALESCE(MIN(trails.elevation_gain), 0) AS min_elevation_gain, \n  COALESCE(MIN(trails.elevation_loss), 0) AS min_elevation_loss, \n  COALESCE(MIN(trails.duration), 0) AS min_duration \nFROM users \n  LEFT JOIN trails ON \n  users.id = trails.author OR \n  trails.public = 1 OR \n  EXISTS (\n    SELECT 1 \n    FROM trail_share \n    WHERE trail_share.trail = trails.id \n    AND trail_share.user = users.id\n  ) GROUP BY users.id;"
				}
			},
			{
				"id": "lf06qip3f4d11yk",
				"created": "2024-04-03 18:44:47.201Z",
				"updated": "2024-06-29 19:23:12.607Z",
				"name": "comments",
				"type": "base",
				"system": false,
				"schema": [
					{
						"system": false,
						"id": "0udwb0kl",
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
					},
					{
						"system": false,
						"id": "fhgxdiam",
						"name": "rating",
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
						"id": "7lwo1mxx",
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
					},
					{
						"system": false,
						"id": "snrlpxar",
						"name": "trail",
						"type": "relation",
						"required": false,
						"presentable": false,
						"unique": false,
						"options": {
							"collectionId": "e864strfxo14pm4",
							"cascadeDelete": true,
							"minSelect": null,
							"maxSelect": 1,
							"displayFields": null
						}
					}
				],
				"indexes": [],
				"listRule": "((@request.auth.id != \"\" && trail.author = @request.auth.id) || trail.public = true) || author = @request.auth.id || trail.trail_share_via_trail.user ?= @request.auth.id",
				"viewRule": "((@request.auth.id != \"\" && trail.author = @request.auth.id) || trail.public = true) || author = @request.auth.id || trail.trail_share_via_trail.user ?= @request.auth.id",
				"createRule": "@request.auth.id != \"\" && (trail.author = @request.auth.id || trail.public = true || trail.trail_share_via_trail.user ?= @request.auth.id)",
				"updateRule": "@request.auth.id = author",
				"deleteRule": "@request.auth.id = author",
				"options": {}
			},
			{
				"id": "uavt73rsqcn1n13",
				"created": "2024-04-13 13:54:26.023Z",
				"updated": "2024-12-16 21:39:04.787Z",
				"name": "settings",
				"type": "base",
				"system": false,
				"schema": [
					{
						"system": false,
						"id": "0sepzvkh",
						"name": "language",
						"type": "select",
						"required": false,
						"presentable": false,
						"unique": false,
						"options": {
							"maxSelect": 1,
							"values": [
								"en",
								"de",
								"fr",
								"hu",
								"it",
								"nl",
								"pl",
								"pt",
								"zh"
							]
						}
					},
					{
						"system": false,
						"id": "zwg1jl0d",
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
						"id": "jo1zcsbu",
						"name": "mapFocus",
						"type": "select",
						"required": false,
						"presentable": false,
						"unique": false,
						"options": {
							"maxSelect": 1,
							"values": [
								"trails",
								"location"
							]
						}
					},
					{
						"system": false,
						"id": "ufhepjxo",
						"name": "location",
						"type": "json",
						"required": false,
						"presentable": false,
						"unique": false,
						"options": {
							"maxSize": 2000000
						}
					},
					{
						"system": false,
						"id": "owlyzl1x",
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
						"id": "xdbayoqg",
						"name": "tilesets",
						"type": "json",
						"required": false,
						"presentable": false,
						"unique": false,
						"options": {
							"maxSize": 2000000
						}
					},
					{
						"system": false,
						"id": "unsh0qsp",
						"name": "terrain",
						"type": "json",
						"required": false,
						"presentable": false,
						"unique": false,
						"options": {
							"maxSize": 2000000
						}
					},
					{
						"system": false,
						"id": "slurqh9q",
						"name": "privacy",
						"type": "json",
						"required": false,
						"presentable": false,
						"unique": false,
						"options": {
							"maxSize": 2000000
						}
					},
					{
						"system": false,
						"id": "5uip7a4p",
						"name": "user",
						"type": "relation",
						"required": false,
						"presentable": false,
						"unique": false,
						"options": {
							"collectionId": "_pb_users_auth_",
							"cascadeDelete": true,
							"minSelect": null,
							"maxSelect": 1,
							"displayFields": null
						}
					}
				],
				"indexes": [],
				"listRule": "user = @request.auth.id",
				"viewRule": "user = @request.auth.id",
				"createRule": "",
				"updateRule": "user = @request.auth.id",
				"deleteRule": null,
				"options": {}
			},
			{
				"id": "urytyc428mwlbqq",
				"created": "2024-04-13 17:00:41.541Z",
				"updated": "2024-06-29 19:23:12.555Z",
				"name": "trails_bounding_box",
				"type": "view",
				"system": false,
				"schema": [
					{
						"system": false,
						"id": "qtzda9la",
						"name": "max_lat",
						"type": "json",
						"required": false,
						"presentable": false,
						"unique": false,
						"options": {
							"maxSize": 1
						}
					},
					{
						"system": false,
						"id": "vfwrnqmf",
						"name": "max_lon",
						"type": "json",
						"required": false,
						"presentable": false,
						"unique": false,
						"options": {
							"maxSize": 1
						}
					},
					{
						"system": false,
						"id": "fyg5ixlu",
						"name": "min_lat",
						"type": "json",
						"required": false,
						"presentable": false,
						"unique": false,
						"options": {
							"maxSize": 1
						}
					},
					{
						"system": false,
						"id": "quugrvou",
						"name": "min_lon",
						"type": "json",
						"required": false,
						"presentable": false,
						"unique": false,
						"options": {
							"maxSize": 1
						}
					}
				],
				"indexes": [],
				"listRule": null,
				"viewRule": "@request.auth.id = id",
				"createRule": null,
				"updateRule": null,
				"deleteRule": null,
				"options": {
					"query": "SELECT users.id, COALESCE(MAX(trails.lat), 0) AS max_lat, COALESCE(MAX(trails.lon), 0) AS max_lon, COALESCE(MIN(trails.lat), 0) AS min_lat, COALESCE(MIN(trails.lon), 0) AS min_lon FROM users LEFT JOIN trails ON users.id = trails.author GROUP BY users.id;"
				}
			},
			{
				"id": "xku110v5a5xbufa",
				"created": "2024-05-09 17:39:57.419Z",
				"updated": "2024-12-16 23:17:35.212Z",
				"name": "users_anonymous",
				"type": "view",
				"system": false,
				"schema": [
					{
						"system": false,
						"id": "hkdk6yjz",
						"name": "username",
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
						"id": "jaienekb",
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
						"id": "e6pktmqs",
						"name": "bio",
						"type": "text",
						"required": false,
						"presentable": false,
						"unique": false,
						"options": {
							"min": null,
							"max": 10000,
							"pattern": ""
						}
					},
					{
						"system": false,
						"id": "b2lrxk9w",
						"name": "private",
						"type": "bool",
						"required": false,
						"presentable": false,
						"unique": false,
						"options": {}
					}
				],
				"indexes": [],
				"listRule": "",
				"viewRule": "",
				"createRule": null,
				"updateRule": null,
				"deleteRule": null,
				"options": {
					"query": "SELECT users.id, username, avatar, bio, users.created, CAST(COALESCE(json_extract(privacy, '$.account') = 'private', false) as BOOL) as private FROM users LEFT JOIN settings ON settings.user = users.id"
				}
			},
			{
				"id": "1mns8mlal6uf9ku",
				"created": "2024-05-09 17:48:01.382Z",
				"updated": "2024-06-29 19:23:12.587Z",
				"name": "trail_share",
				"type": "base",
				"system": false,
				"schema": [
					{
						"system": false,
						"id": "eskurfx6",
						"name": "trail",
						"type": "relation",
						"required": true,
						"presentable": false,
						"unique": false,
						"options": {
							"collectionId": "e864strfxo14pm4",
							"cascadeDelete": true,
							"minSelect": null,
							"maxSelect": 1,
							"displayFields": null
						}
					},
					{
						"system": false,
						"id": "yyzimwee",
						"name": "user",
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
						"id": "zr7aaqxl",
						"name": "permission",
						"type": "select",
						"required": true,
						"presentable": false,
						"unique": false,
						"options": {
							"maxSelect": 1,
							"values": [
								"view",
								"edit"
							]
						}
					}
				],
				"indexes": [],
				"listRule": "trail.author = @request.auth.id || user = @request.auth.id",
				"viewRule": "trail.author = @request.auth.id || user = @request.auth.id",
				"createRule": "trail.author = @request.auth.id",
				"updateRule": "trail.author = @request.auth.id",
				"deleteRule": "trail.author = @request.auth.id",
				"options": {}
			},
			{
				"id": "_pb_users_auth_",
				"created": "2024-06-29 19:23:12.388Z",
				"updated": "2024-12-07 14:12:14.202Z",
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
						"id": "pd2cq8sq",
						"name": "bio",
						"type": "text",
						"required": false,
						"presentable": false,
						"unique": false,
						"options": {
							"min": null,
							"max": 10000,
							"pattern": ""
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
				"id": "1kot7t9na3hi0gl",
				"created": "2024-09-13 12:40:20.308Z",
				"updated": "2024-09-13 12:59:40.556Z",
				"name": "list_share",
				"type": "base",
				"system": false,
				"schema": [
					{
						"system": false,
						"id": "luqrtipy",
						"name": "list",
						"type": "relation",
						"required": true,
						"presentable": false,
						"unique": false,
						"options": {
							"collectionId": "r6gu2ajyidy1x69",
							"cascadeDelete": false,
							"minSelect": null,
							"maxSelect": 1,
							"displayFields": null
						}
					},
					{
						"system": false,
						"id": "mix12kkh",
						"name": "user",
						"type": "relation",
						"required": true,
						"presentable": false,
						"unique": false,
						"options": {
							"collectionId": "_pb_users_auth_",
							"cascadeDelete": false,
							"minSelect": null,
							"maxSelect": 1,
							"displayFields": null
						}
					},
					{
						"system": false,
						"id": "n9rjdx5g",
						"name": "permission",
						"type": "select",
						"required": true,
						"presentable": false,
						"unique": false,
						"options": {
							"maxSelect": 1,
							"values": [
								"view",
								"edit"
							]
						}
					}
				],
				"indexes": [],
				"listRule": "list.author = @request.auth.id || user = @request.auth.id",
				"viewRule": "list.author = @request.auth.id || user = @request.auth.id",
				"createRule": "list.author = @request.auth.id",
				"updateRule": "list.author = @request.auth.id",
				"deleteRule": "list.author = @request.auth.id",
				"options": {}
			},
			{
				"id": "t9lphichi5xwyeu",
				"created": "2024-12-14 14:25:55.563Z",
				"updated": "2024-12-16 23:26:17.884Z",
				"name": "activities",
				"type": "view",
				"system": false,
				"schema": [
					{
						"system": false,
						"id": "7xrdkudh",
						"name": "trail_id",
						"type": "json",
						"required": false,
						"presentable": false,
						"unique": false,
						"options": {
							"maxSize": 1
						}
					},
					{
						"system": false,
						"id": "l1cqbctd",
						"name": "date",
						"type": "json",
						"required": false,
						"presentable": false,
						"unique": false,
						"options": {
							"maxSize": 1
						}
					},
					{
						"system": false,
						"id": "fe8b0uvy",
						"name": "name",
						"type": "json",
						"required": false,
						"presentable": false,
						"unique": false,
						"options": {
							"maxSize": 1
						}
					},
					{
						"system": false,
						"id": "zbieb39e",
						"name": "description",
						"type": "json",
						"required": false,
						"presentable": false,
						"unique": false,
						"options": {
							"maxSize": 1
						}
					},
					{
						"system": false,
						"id": "e8edjmsl",
						"name": "gpx",
						"type": "json",
						"required": false,
						"presentable": false,
						"unique": false,
						"options": {
							"maxSize": 1
						}
					},
					{
						"system": false,
						"id": "htr3h7jb",
						"name": "author",
						"type": "json",
						"required": false,
						"presentable": false,
						"unique": false,
						"options": {
							"maxSize": 1
						}
					},
					{
						"system": false,
						"id": "yhf5zw13",
						"name": "photos",
						"type": "json",
						"required": false,
						"presentable": false,
						"unique": false,
						"options": {
							"maxSize": 1
						}
					},
					{
						"system": false,
						"id": "digs0pen",
						"name": "distance",
						"type": "json",
						"required": false,
						"presentable": false,
						"unique": false,
						"options": {
							"maxSize": 1
						}
					},
					{
						"system": false,
						"id": "ropo4j37",
						"name": "duration",
						"type": "json",
						"required": false,
						"presentable": false,
						"unique": false,
						"options": {
							"maxSize": 1
						}
					},
					{
						"system": false,
						"id": "bawfccve",
						"name": "elevation_gain",
						"type": "json",
						"required": false,
						"presentable": false,
						"unique": false,
						"options": {
							"maxSize": 1
						}
					},
					{
						"system": false,
						"id": "5lbx0uvm",
						"name": "elevation_loss",
						"type": "json",
						"required": false,
						"presentable": false,
						"unique": false,
						"options": {
							"maxSize": 1
						}
					},
					{
						"system": false,
						"id": "sbhzsvjc",
						"name": "type",
						"type": "json",
						"required": false,
						"presentable": false,
						"unique": false,
						"options": {
							"maxSize": 1
						}
					}
				],
				"indexes": [],
				"listRule": "@request.auth.id = author || (@collection.users_anonymous.id ?= author && @collection.users_anonymous.private ?= false)",
				"viewRule": "@request.auth.id = author || (@collection.users_anonymous.id ?= author && @collection.users_anonymous.private ?= false)",
				"createRule": null,
				"updateRule": null,
				"deleteRule": null,
				"options": {
					"query": "SELECT id,trail_id,date,name,description,gpx,author,photos,distance,duration,elevation_gain,elevation_loss, type \nFROM (\n    SELECT summit_logs.id,trails.id as trail_id,summit_logs.date,trails.name,text as description,summit_logs.gpx,summit_logs.author,summit_logs.photos,summit_logs.distance,summit_logs.duration,summit_logs.elevation_gain,summit_logs.elevation_loss,\"summit_log\" as type \n    FROM summit_logs\n    LEFT JOIN trails ON summit_logs.id IN (\n    SELECT value\n    FROM json_each(trails.summit_logs)\n    )\n    UNION\n    SELECT id,id as trail_id,date,name,description,gpx,author,photos,distance,duration,elevation_gain,elevation_loss, \"trail\" as type \n    FROM trails\n)\nORDER BY date DESC"
				}
			},
			{
				"id": "8obn1ukumze565i",
				"created": "2024-12-14 17:17:00.381Z",
				"updated": "2024-12-16 23:26:53.312Z",
				"name": "follows",
				"type": "base",
				"system": false,
				"schema": [
					{
						"system": false,
						"id": "in1traur",
						"name": "follower",
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
					},
					{
						"system": false,
						"id": "wxwomfd5",
						"name": "followee",
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
				"listRule": "@request.auth.id = follower.id || @request.auth.id = followee.id",
				"viewRule": "@request.auth.id = follower.id || @request.auth.id = followee.id",
				"createRule": "@request.auth.id = follower.id && (@collection.users_anonymous.id ?= followee && @collection.users_anonymous.private ?= false)",
				"updateRule": "@request.auth.id = follower.id",
				"deleteRule": "@request.auth.id = follower.id",
				"options": {}
			},
			{
				"id": "j6w72f0kb5ivd7x",
				"created": "2024-12-14 18:20:18.920Z",
				"updated": "2024-12-16 23:25:12.103Z",
				"name": "follow_counts",
				"type": "view",
				"system": false,
				"schema": [
					{
						"system": false,
						"id": "6axn8l2f",
						"name": "followers",
						"type": "json",
						"required": false,
						"presentable": false,
						"unique": false,
						"options": {
							"maxSize": 1
						}
					},
					{
						"system": false,
						"id": "mjrckc16",
						"name": "following",
						"type": "json",
						"required": false,
						"presentable": false,
						"unique": false,
						"options": {
							"maxSize": 1
						}
					}
				],
				"indexes": [],
				"listRule": "@request.auth.id = id || (@collection.users_anonymous.id ?= id && @collection.users_anonymous.private ?= false)",
				"viewRule": "@request.auth.id = id || (@collection.users_anonymous.id ?= id && @collection.users_anonymous.private ?= false)",
				"createRule": null,
				"updateRule": null,
				"deleteRule": null,
				"options": {
					"query": "SELECT \n    users.id,\n    COALESCE(followers.count, 0) AS followers,\n    COALESCE(following.count, 0) AS following\nFROM users\nLEFT JOIN (\n    SELECT followee AS user_id, COUNT(*) AS count\n    FROM follows\n    GROUP BY followee\n) AS followers ON users.id = followers.user_id\nLEFT JOIN (\n    SELECT follower AS user_id, COUNT(*) AS count\n    FROM follows\n    GROUP BY follower\n) AS following ON users.id = following.user_id"
				}
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
