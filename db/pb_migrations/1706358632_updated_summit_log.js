/// <reference path="../pb_data/types.d.ts" />
migrate((db) => {
  const dao = new Dao(db)
  const collection = dao.findCollectionByNameOrId("dd2l9a4vxpy2ni8")

  // add
  collection.schema.addField(new SchemaField({
    "system": false,
    "id": "xwqncqng",
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
  }))

  return dao.saveCollection(collection)
}, (db) => {
  const dao = new Dao(db)
  const collection = dao.findCollectionByNameOrId("dd2l9a4vxpy2ni8")

  // remove
  collection.schema.removeField("xwqncqng")

  return dao.saveCollection(collection)
})
