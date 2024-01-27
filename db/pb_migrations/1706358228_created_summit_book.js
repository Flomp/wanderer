/// <reference path="../pb_data/types.d.ts" />
migrate((db) => {
  const collection = new Collection({
    "id": "dd2l9a4vxpy2ni8",
    "created": "2024-01-27 12:23:48.069Z",
    "updated": "2024-01-27 12:23:48.069Z",
    "name": "summit_book",
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
    "listRule": null,
    "viewRule": null,
    "createRule": null,
    "updateRule": null,
    "deleteRule": null,
    "options": {}
  });

  return Dao(db).saveCollection(collection);
}, (db) => {
  const dao = new Dao(db);
  const collection = dao.findCollectionByNameOrId("dd2l9a4vxpy2ni8");

  return dao.deleteCollection(collection);
})
