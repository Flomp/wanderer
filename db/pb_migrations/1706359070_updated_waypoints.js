/// <reference path="../pb_data/types.d.ts" />
migrate((db) => {
  const dao = new Dao(db)
  const collection = dao.findCollectionByNameOrId("goeo2ubp103rzp9")

  // add
  collection.schema.addField(new SchemaField({
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
  }))

  return dao.saveCollection(collection)
}, (db) => {
  const dao = new Dao(db)
  const collection = dao.findCollectionByNameOrId("goeo2ubp103rzp9")

  // remove
  collection.schema.removeField("rnjgm2tk")

  return dao.saveCollection(collection)
})
