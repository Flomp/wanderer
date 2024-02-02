/// <reference path="../pb_data/types.d.ts" />
migrate((db) => {
  const dao = new Dao(db)
  const collection = dao.findCollectionByNameOrId("goeo2ubp103rzp9")

  collection.listRule = "@request.auth.id != \"\" && @collection.trails.waypoints.id ?= id && (@collection.trails.author = @request.auth.id || @collection.trails.public = true)"
  collection.viewRule = "@request.auth.id != \"\" && @collection.trails.waypoints.id ?= id && (@collection.trails.author = @request.auth.id || @collection.trails.public = true)"

  return dao.saveCollection(collection)
}, (db) => {
  const dao = new Dao(db)
  const collection = dao.findCollectionByNameOrId("goeo2ubp103rzp9")

  collection.listRule = "@request.auth.id != \"\" && @collection.trails.waypoints.id ?= id "
  collection.viewRule = "@request.auth.id != \"\" && @collection.trails.waypoints.id ?= id "

  return dao.saveCollection(collection)
})
