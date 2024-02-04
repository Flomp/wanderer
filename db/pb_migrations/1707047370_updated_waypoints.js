/// <reference path="../pb_data/types.d.ts" />
migrate((db) => {
  const dao = new Dao(db)
  const collection = dao.findCollectionByNameOrId("goeo2ubp103rzp9")

  collection.updateRule = "@request.auth.id != \"\" && @collection.trails.waypoints.id ?= id && (@collection.trails.author = @request.auth.id)"
  collection.deleteRule = "@request.auth.id != \"\" && @collection.trails.waypoints.id ?= id && (@collection.trails.author = @request.auth.id)"

  return dao.saveCollection(collection)
}, (db) => {
  const dao = new Dao(db)
  const collection = dao.findCollectionByNameOrId("goeo2ubp103rzp9")

  collection.updateRule = "@request.auth.id != \"\" && @collection.trails.waypoints ?= id && (@collection.trails.author = @request.auth.id)"
  collection.deleteRule = "@request.auth.id != \"\" && @collection.trails.waypoints ?= id && (@collection.trails.author = @request.auth.id)"

  return dao.saveCollection(collection)
})
