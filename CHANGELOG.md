# v0.5.1
## Features
- You can now export trails as GPX or GEOJson files. Optionally you can include photos and summit book entries of the trail. This replaces the "Download GPX" function in previous versions.

## Bug fixes
- Fixed a bug that wouldprevent users from creating multiple summit log entries with out reloading the page
- wanderer now takes the `<rte>` tag into account when displaying a trail on the map
- A trails date attribute is now the current date by default

# v0.5.0

## ⚠️ Breaking changes ⚠️
- This version updates the index pattern of the meilisearch index. Please delete your `data.ms` folder before launching wanderer. The indices will be rebuilt on launch. Otherwise trail filtering will no longer work.

## Features
- Trails can now be filtered by date
- Elevation, slope and speed graphs are now also visible when creating a new trail
- When creating a new trail you now have the option to create a new route from scratch without uploading a GPX file. Press the "Draw a route" button and plan your new route directly in wanderer. We use [valhalla](https://github.com/valhalla/valhalla) and their associated free [hosted service](https://gis-ops.com/global-open-valhalla-server-online/) to calculate the routes. To activate the feature make sure to set the PUBLIC_VALHALLA_URL environment variable on you wanderer-web service. See the current [docker-compose.yml](https://github.com/Flomp/wanderer/blob/main/docker-compose.yml) for a working configuration.

## Bug fixes
- Uploaded trails will now have a date if it can be parsed from the file

# v0.4.0
## Features

- Trails can now be printed to PDF. Select "Print" from the menu when viewing a trail.
- Trails can now have a date.
- You can now comment on trails
- Adds a setting to focus the map on all trails instead of a specific location

## Bug fixes
- fixed a bug that would show a wrong date for summit logs for certain time zones (now really)
- fixed a bug that prevented waypoints from showing up in public trails

# v0.3.2
## Bug fixes

- Fixed a bug that caused a 500 Internal Error to appear when viewing trails without an account

## Translations

- improves Dutch translation (thanks to [Vistaus](https://github.com/Vistaus))


# v0.3.1
## Features

- Max values for elevation gain and distance filters are now dynamically calculated based on your longest trail
- Disabling username/email & password auth in PocketBase is now reflected in wanderer's login UI

## Bug fixes

- Fixed a bug that prevented import trails from appearing in map view

## Translations

- adds Simplified Chinese translation (thanks to [icyleaf](https://github.com/icyleaf))

## Dependencies

- updates PocketBase to v0.22.7

# v0.3.0
## Features

- Trails can now be added to a list while editing or creating a trail. The trail must be saved at least once to add it to a list.
- wanderer now has an auto-upload folder. GPX files in this folder will be autmatically uploaded and converted to a trail. Read the [docs](https://github.com/Flomp/wanderer/wiki/API#auto-upload-folder) for more information.
- addded support for TCX and KML files. Note that this feature is still experimental. Please report any issues you encounter.
- added OAuth support. Read [here](https://github.com/Flomp/wanderer/wiki/OAuth) how to enable providers.
  
## Bug fixes

- fixed a bug that would show a wrong date for summit logs for certain time zones
- added client side validation for usernames

## Translations
- adds French translation (thanks to [seb2020](https://github.com/seb2020))
- adds Hungarian translation (thanks to [sszemtelen](https://github.com/sszemtelen))


# v0.2.1
## Bug fixes

- summit book dates now show in the correct format for the current locale
- fixed a bug that would overwrite trail names and descriptions when editing a trail

## Translations

- adds Dutch translation (thanks to [yves-bonami](https://github.com/yves-bonami))
- adds Polish translation (thanks to [ludrol](https://github.com/ludrol))
- adds Portuguese translation (thanks to [poVoq](https://github.com/poVoq))

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