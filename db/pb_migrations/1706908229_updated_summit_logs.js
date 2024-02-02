/// <reference path="../pb_data/types.d.ts" />
migrate((db) => {
  const dao = new Dao(db)
  const collection = dao.findCollectionByNameOrId("dd2l9a4vxpy2ni8")

  collection.viewRule = ""

  return dao.saveCollection(collection)
}, (db) => {
  const dao = new Dao(db)
  const collection = dao.findCollectionByNameOrId("dd2l9a4vxpy2ni8")

  collection.viewRule = "@request.auth.id != \"\" && @collection.trails.summit_logs ?= id && (@collection.trails.author = @request.auth.id || @collection.trails.public = true)"

  return dao.saveCollection(collection)
})
