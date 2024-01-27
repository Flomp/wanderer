/// <reference path="../pb_data/types.d.ts" />
migrate((db) => {
  const dao = new Dao(db)
  const collection = dao.findCollectionByNameOrId("e864strfxo14pm4")

  // add
  collection.schema.addField(new SchemaField({
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
  }))

  // add
  collection.schema.addField(new SchemaField({
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
  }))

  return dao.saveCollection(collection)
}, (db) => {
  const dao = new Dao(db)
  const collection = dao.findCollectionByNameOrId("e864strfxo14pm4")

  // remove
  collection.schema.removeField("e1lwowvd")

  // remove
  collection.schema.removeField("ppq2sist")

  return dao.saveCollection(collection)
})
