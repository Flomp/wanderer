/// <reference path="../pb_data/types.d.ts" />
migrate((db) => {
  const dao = new Dao(db)
  const collection = dao.findCollectionByNameOrId("e864strfxo14pm4")

  collection.viewRule = ""

  return dao.saveCollection(collection)
}, (db) => {
  const dao = new Dao(db)
  const collection = dao.findCollectionByNameOrId("e864strfxo14pm4")

  collection.viewRule = "author = @request.auth.id || public = true"

  return dao.saveCollection(collection)
})
