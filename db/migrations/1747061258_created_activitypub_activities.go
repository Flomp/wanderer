package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		jsonData := `[
    {
        "id": "pbc_3752774184",
        "listRule": "",
        "viewRule": "",
        "createRule": null,
        "updateRule": null,
        "deleteRule": null,
        "name": "activitypub_activities",
        "type": "base",
        "fields": [
            {
                "autogeneratePattern": "[a-z0-9]{15}",
                "hidden": false,
                "id": "text3208210256",
                "max": 15,
                "min": 15,
                "name": "id",
                "pattern": "^[a-z0-9]+$",
                "presentable": false,
                "primaryKey": true,
                "required": true,
                "system": true,
                "type": "text"
            },
            {
                "exceptDomains": [],
                "hidden": false,
                "id": "url2434853685",
                "name": "iri",
                "onlyDomains": [],
                "presentable": false,
                "required": true,
                "system": false,
                "type": "url"
            },
            {
                "autogeneratePattern": "",
                "hidden": false,
                "id": "text2363381545",
                "max": 0,
                "min": 0,
                "name": "type",
                "pattern": "",
                "presentable": false,
                "primaryKey": false,
                "required": true,
                "system": false,
                "type": "text"
            },
            {
                "hidden": false,
                "id": "json3616002756",
                "maxSize": 0,
                "name": "to",
                "presentable": false,
                "required": false,
                "system": false,
                "type": "json"
            },
            {
                "hidden": false,
                "id": "json3685882489",
                "maxSize": 0,
                "name": "cc",
                "presentable": false,
                "required": false,
                "system": false,
                "type": "json"
            },
            {
                "hidden": false,
                "id": "json2893285722",
                "maxSize": 0,
                "name": "object",
                "presentable": false,
                "required": true,
                "system": false,
                "type": "json"
            },
            {
                "exceptDomains": [],
                "hidden": false,
                "id": "url1148540665",
                "name": "actor",
                "onlyDomains": [],
                "presentable": false,
                "required": true,
                "system": false,
                "type": "url"
            },
            {
                "hidden": false,
                "id": "date1748787223",
                "max": "",
                "min": "",
                "name": "published",
                "presentable": false,
                "required": true,
                "system": false,
                "type": "date"
            },
            {
                "autogeneratePattern": "",
                "hidden": false,
                "id": "text1653163849",
                "max": 15,
                "min": 15,
                "name": "relation",
                "pattern": "^[a-z0-9]+$",
                "presentable": false,
                "primaryKey": false,
                "required": false,
                "system": false,
                "type": "text"
            },
            {
                "hidden": false,
                "id": "autodate2990389176",
                "name": "created",
                "onCreate": true,
                "onUpdate": false,
                "presentable": false,
                "system": false,
                "type": "autodate"
            },
            {
                "hidden": false,
                "id": "autodate3332085495",
                "name": "updated",
                "onCreate": true,
                "onUpdate": true,
                "presentable": false,
                "system": false,
                "type": "autodate"
            }
        ],
        "indexes": [],
        "system": false
    }
]`

		return app.ImportCollectionsByMarshaledJSON([]byte(jsonData), false)
	}, func(app core.App) error {
		return nil
	})
}
