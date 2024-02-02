/// <reference path="../pb_data/types.d.ts" />
migrate((db) => {
  const dao = new Dao(db)
  const collection = dao.findCollectionByNameOrId("e864strfxo14pm4")

  collection.listRule = "author = @request.auth.id || public = true"
  collection.viewRule = "author = @request.auth.id || public = true"
  collection.createRule = "@request.auth.id != \"\" && (@request.data.author = @request.auth.id)"
  collection.updateRule = "author = @request.auth.id "
  collection.deleteRule = "author = @request.auth.id "

  return dao.saveCollection(collection)
}, (db) => {
  const dao = new Dao(db)
  const collection = dao.findCollectionByNameOrId("e864strfxo14pm4")

  collection.listRule = ""
  collection.viewRule = ""
  collection.createRule = ""
  collection.updateRule = ""
  collection.deleteRule = ""

  return dao.saveCollection(collection)
})
