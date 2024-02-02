/// <reference path="../pb_data/types.d.ts" />
migrate((db) => {
  const dao = new Dao(db)
  const collection = dao.findCollectionByNameOrId("goeo2ubp103rzp9")

  collection.listRule = "@request.auth.id != \"\" && @collection.trails.waypoints = id && (@collection.trails.author = @request.auth.id || @collection.trails.public = true)"
  collection.viewRule = "@request.auth.id != \"\" && @collection.trails.waypoints = id && (@collection.trails.author = @request.auth.id || @collection.trails.public = true)"
  collection.createRule = "@request.auth.id != \"\""
  collection.updateRule = "@request.auth.id != \"\" && @collection.trails.waypoints = id && (@collection.trails.author = @request.auth.id)"
  collection.deleteRule = "@request.auth.id != \"\" && @collection.trails.waypoints = id && (@collection.trails.author = @request.auth.id)"

  return dao.saveCollection(collection)
}, (db) => {
  const dao = new Dao(db)
  const collection = dao.findCollectionByNameOrId("goeo2ubp103rzp9")

  collection.listRule = ""
  collection.viewRule = ""
  collection.createRule = ""
  collection.updateRule = ""
  collection.deleteRule = ""

  return dao.saveCollection(collection)
})
