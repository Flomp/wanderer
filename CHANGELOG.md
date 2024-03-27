# v0.2.1
## Bug fixes

- summit book dates now show in the correct format for the current locale
- fixed a bug that would overwrite trail names and descriptions when editing a trail

## Translations

- adds Dutch translation (thanks to [yves-bonami](https://github.com/yves-bonami))
- adds Polish translation (thanks to [ludrol](https://github.com/ludrol))
- adds Portugese translation (thanks to [poVoq](https://github.com/poVoq))

# v0.2.0
## Features

- when creating/editing a trail you can now drag & drop photos into the photo section to upload them
- you can now attach photos to waypoints
- you can now edit the "Distance", "Elevation Gain" and "Est. Duration" fields
- waypoint markers can now be moved with drag & drop
- lists can now be displayed as a map showing all trails contained in the list
- you can now prevent users from signing up by setting the `DISABLE_SIGNUP` environment variable to `true`
- you can now upload GPX files via the API to create trails. Check the [documentation](https://github.com/Flomp/wanderer/wiki/API#upload-trails) for more info.
- the city index now includes states

> Note: for city states to show up in your search you have to delete your data.ms folder if you already have a previous installation of wanderer. The indices will then be rebuilt on startup.

## Docs

-  added complete API documentation

# v0.1.1
## Bug fixes

- fixed a bug that would prevent trails longer than 20km from being displayed
- added BODY_SIZE_LIMIT env variable to docker compose to allow for bigger file uploads
- fixed a bug that caused only 5 trails to be shown at a time
- fixed a bug that would cause waypoints not to be deleted from the backend
- updated the default docker-compose.yml to include a secure MEILI_MASTER_KEY
- the default location field now sets the value correctly after clicking on a search result
  
## Docs

- updated the docs to include BODY_SIZE_LIMIT

# v0.1.0 
- Initial release