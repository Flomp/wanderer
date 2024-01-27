/// <reference path="../pb_data/types.d.ts" />
migrate((db) => {
  const dao = new Dao(db)
  const collection = dao.findCollectionByNameOrId("goeo2ubp103rzp9")

  // remove
  collection.schema.removeField("0fz9yp7c")

  return dao.saveCollection(collection)
}, (db) => {
  const dao = new Dao(db)
  const collection = dao.findCollectionByNameOrId("goeo2ubp103rzp9")

  // add
  collection.schema.addField(new SchemaField({
    "system": false,
    "id": "0fz9yp7c",
    "name": "trail",
    "type": "relation",
    "required": false,
    "presentable": false,
    "unique": false,
    "options": {
      "collectionId": "e864strfxo14pm4",
      "cascadeDelete": false,
      "minSelect": null,
      "maxSelect": 1,
      "displayFields": null
    }
  }))

  return dao.saveCollection(collection)
})
