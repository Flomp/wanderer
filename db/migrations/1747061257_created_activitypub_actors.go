package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		jsonData := `[
    {
        "id": "pbc_1295301207",
        "listRule": "",
        "viewRule": "",
        "createRule": null,
        "updateRule": null,
        "deleteRule": null,
        "name": "activitypub_actors",
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
                "autogeneratePattern": "",
                "hidden": false,
                "id": "text4166911607",
                "max": 0,
                "min": 0,
                "name": "username",
                "pattern": "",
                "presentable": false,
                "primaryKey": false,
                "required": true,
                "system": false,
                "type": "text"
            },
            {
                "autogeneratePattern": "",
                "hidden": false,
                "id": "text4002953752",
                "max": 0,
                "min": 0,
                "name": "preferred_username",
                "pattern": "",
                "presentable": false,
                "primaryKey": false,
                "required": false,
                "system": false,
                "type": "text"
            },
            {
                "autogeneratePattern": "",
                "hidden": false,
                "id": "text2812878347",
                "max": 0,
                "min": 0,
                "name": "domain",
                "pattern": "",
                "presentable": false,
                "primaryKey": false,
                "required": false,
                "system": false,
                "type": "text"
            },
            {
                "autogeneratePattern": "",
                "hidden": false,
                "id": "text3458754147",
                "max": 0,
                "min": 0,
                "name": "summary",
                "pattern": "",
                "presentable": false,
                "primaryKey": false,
                "required": false,
                "system": false,
                "type": "text"
            },
            {
                "hidden": false,
                "id": "number1386272118",
                "max": null,
                "min": null,
                "name": "followerCount",
                "onlyInt": true,
                "presentable": false,
                "required": false,
                "system": false,
                "type": "number"
            },
            {
                "hidden": false,
                "id": "number3430500629",
                "max": null,
                "min": null,
                "name": "followingCount",
                "onlyInt": true,
                "presentable": false,
                "required": false,
                "system": false,
                "type": "number"
            },
            {
                "hidden": false,
                "id": "date1748787223",
                "max": "",
                "min": "",
                "name": "published",
                "presentable": false,
                "required": false,
                "system": false,
                "type": "date"
            },
            {
                "exceptDomains": null,
                "hidden": false,
                "id": "url126331327",
                "name": "iri",
                "onlyDomains": null,
                "presentable": false,
                "required": true,
                "system": false,
                "type": "url"
            },
            {
                "exceptDomains": null,
                "hidden": false,
                "id": "url2115105593",
                "name": "inbox",
                "onlyDomains": null,
                "presentable": false,
                "required": true,
                "system": false,
                "type": "url"
            },
            {
                "exceptDomains": null,
                "hidden": false,
                "id": "url1793578352",
                "name": "outbox",
                "onlyDomains": null,
                "presentable": false,
                "required": false,
                "system": false,
                "type": "url"
            },
            {
                "exceptDomains": null,
                "hidden": false,
                "id": "url1704208859",
                "name": "icon",
                "onlyDomains": null,
                "presentable": false,
                "required": false,
                "system": false,
                "type": "url"
            },
            {
                "exceptDomains": null,
                "hidden": false,
                "id": "url2215181735",
                "name": "followers",
                "onlyDomains": null,
                "presentable": false,
                "required": false,
                "system": false,
                "type": "url"
            },
            {
                "exceptDomains": null,
                "hidden": false,
                "id": "url1908379107",
                "name": "following",
                "onlyDomains": null,
                "presentable": false,
                "required": false,
                "system": false,
                "type": "url"
            },
            {
                "hidden": false,
                "id": "bool2193750486",
                "name": "isLocal",
                "presentable": false,
                "required": false,
                "system": false,
                "type": "bool"
            },
            {
                "autogeneratePattern": "",
                "hidden": false,
                "id": "text1727648867",
                "max": 0,
                "min": 0,
                "name": "public_key",
                "pattern": "",
                "presentable": false,
                "primaryKey": false,
                "required": true,
                "system": false,
                "type": "text"
            },
            {
                "autogeneratePattern": "",
                "hidden": true,
                "id": "text4160324774",
                "max": 0,
                "min": 0,
                "name": "private_key",
                "pattern": "",
                "presentable": false,
                "primaryKey": false,
                "required": false,
                "system": false,
                "type": "text"
            },
            {
                "cascadeDelete": true,
                "collectionId": "_pb_users_auth_",
                "hidden": false,
                "id": "relation2375276105",
                "maxSelect": 1,
                "minSelect": 0,
                "name": "user",
                "presentable": false,
                "required": false,
                "system": false,
                "type": "relation"
            },
            {
                "hidden": false,
                "id": "date2062531289",
                "max": "",
                "min": "",
                "name": "last_fetched",
                "presentable": false,
                "required": false,
                "system": false,
                "type": "date"
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
        "indexes": [
            "CREATE UNIQUE INDEX ` + "`idx_rpT7QJwWTm` ON `activitypub_actors` (`iri`)" + `"
        ],
        "system": false
    }
]`

		return app.ImportCollectionsByMarshaledJSON([]byte(jsonData), false)
	}, func(app core.App) error {
		return nil
	})
}
