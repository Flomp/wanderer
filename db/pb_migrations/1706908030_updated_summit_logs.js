/// <reference path="../pb_data/types.d.ts" />
migrate((db) => {
  const dao = new Dao(db)
  const collection = dao.findCollectionByNameOrId("dd2l9a4vxpy2ni8")

  collection.listRule = "@request.auth.id != \"\" && @collection.trails.summit_logs ?= id && (@collection.trails.author = @request.auth.id || @collection.trails.public = true)"
  collection.viewRule = "@request.auth.id != \"\" && @collection.trails.summit_logs ?= id && (@collection.trails.author = @request.auth.id || @collection.trails.public = true)"
  collection.updateRule = "@request.auth.id != \"\" && @collection.trails.waypoints ?= id && (@collection.trails.author = @request.auth.id)"
  collection.deleteRule = "@request.auth.id != \"\" && @collection.trails.waypoints ?= id && (@collection.trails.author = @request.auth.id)"

  return dao.saveCollection(collection)
}, (db) => {
  const dao = new Dao(db)
  const collection = dao.findCollectionByNameOrId("dd2l9a4vxpy2ni8")

  collection.listRule = "@request.auth.id != \"\" && @collection.trails.waypoints = id && (@collection.trails.author = @request.auth.id || @collection.trails.public = true)"
  collection.viewRule = "@request.auth.id != \"\" && @collection.trails.waypoints = id && (@collection.trails.author = @request.auth.id || @collection.trails.public = true)"
  collection.updateRule = "@request.auth.id != \"\" && @collection.trails.waypoints = id && (@collection.trails.author = @request.auth.id)"
  collection.deleteRule = "@request.auth.id != \"\" && @collection.trails.waypoints = id && (@collection.trails.author = @request.auth.id)"

  return dao.saveCollection(collection)
})
