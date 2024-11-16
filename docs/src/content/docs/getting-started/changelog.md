---
title: Changelog
description: What changed in the last patch?
---

## v0.11.0
### Features
- Other user's profiles can now be viewed
- The summit log author is now listed in the summit log table
- The trail author is now listed for every trail
- Trails can now be filtered by author
- Users will now receive an error message when they try to upload a photo that is too large
- Updated 3D Model on front page
  
### Bug fixes
- Waypoints and summit logs of shared trails are now properly displayed
- Fixes missing translation for trail categories
- Fixes events in profile calendar

## v0.10.1
### Features
- Adds elevation loss to trails. Please note that trails created before this version will have a default elevation loss of 0. Edit & save to update.
  
### Bug fixes
- Fixes bug that caused auto-added summit logs to not have distance, durtaion etc.
- Fixes error in the auto-upload feature
- Fixes access permissions for profile page
- A summit log is now also created automatically when you upload a trail directly
- Fixes link to the API documentation in footer

## v0.10.0
### Features
- A new summit log entry is now added automatically when uploading a new GPX file for a new or existing trail
- GPX files can now be attached to summit logs
- Adds a new profile page with filterable statistics derived from summit log GPX data
- When importing a GPX file the filename will be used as a fallback

## v0.9.0
### Features
- Complete visual overhaul of the list page
- Lists can now be shared with other users
- Updates visual style for waypoints

### Bug fixes
- Fixes a bug that caused categories not to load properly on page reload
- Fixes icon picker suggestions for waypoints
- Fixes a bug that would prevent public and shared trails from showing up in the overview

## v0.8.2
### Features
- The pagination is now available at top and bottom
- HEIC image format is now supported
- Exporting only a trail without summit logs or photos will create a single file instead of a ZIP folder
- The current page is now remembered when navigating back to the trail overview

### Bug fixes
- Fixed map height when viewing a trail in detail view

## v0.8.1
### Features
- Public and shared trails can now be exported and printed

### Bug fixes
- Correctly adds the xmlns header to exported GPX files
- Fixes detail view for shared and public trails
- Fixes bug that caused the default category to be re-applied when editing a trail


## v0.8.0
### Features
- The settings page layout got a complete visual overhaul
- You can now change your email and password from the web UI
- You can import GPX, KML, TCX, and FIT files directly through the UI or using the API
- Exporting all trails at once is now possible on the settings page
- Additional custom map tile sets can now be added via the settings page

### Bug fixes
- Added multiple missing translations
- The hotline metric selector icons are now radio buttons instead of checkboxes


## v0.7.3
### Features
- Adds support for FIT files
  
### Bug fixes
- Fixes display issue for the filter panel in the map view

### Translations
- adds Italian translation (thanks to [lukasitaly](https://github.com/lukasitaly))
  

## v0.7.2
### Bug fixes
- Icons in dropdown menus are now displaying properly again
- The trail id is now returned when using the upload API endpoint

## v0.7.1
### Features
- A warning is now displayed if the ORIGIN environment variable is misconfigured
  
### Bug fixes
- Fixes login for http connections
- Trails are now properly sorted across all pages

## v0.7.0
### Features
- Trail sort and sort direction are now remembered through a page reload
- You can now pick between OpenStreetMaps and OpenTopoMaps tiles
- The gradient track line can now display altitude, slope, and speed
- A waypoint's coordinates can now be inferred from a photo's EXIF data

### Docs

- Added a guide for custom categories

## v0.6.1
### Bug fixes
- Fixes a bug that would show wrong comments under a trail
- Fixes an overflow issue in list views
- Settings are now created properly when signing up with OAuth
- Fixes a bug that caused properly named trails to get a generic name

## v0.6.0
### Features
- You can now share trails with other users by selecting "Share" from the trail contextmenu. You can set the permission level as "View" or "Edit".
- Adds missing translations

### Bug fixes
- Fixed a bug that prevented comments from showing up
- GPX files without a name in the metadata section will now receive a generic name when uploaded through the API

## v0.5.1
### Features
- You can now export trails as GPX or GEOJson files. Optionally you can include photos and summit book entries of the trail. This replaces the "Download GPX" function in previous versions.

### Bug fixes
- Fixed a bug that would prevent users from creating multiple summit log entries without reloading the page
- wanderer now takes the `<rte>` tag into account when displaying a trail on the map
- A trail's date attribute is now the current date by default

## v0.5.0

### ⚠️ Breaking changes ⚠️
- This version updates the index pattern of the meilisearch index. Please delete your `data.ms` folder before launching wanderer. The indices will be rebuilt on launch. Otherwise trail filtering will no longer work.

### Features
- Trails can now be filtered by date
- Elevation, slope and speed graphs are now also visible when creating a new trail
- When creating a new trail you now have the option to create a new route from scratch without uploading a GPX file. Press the "Draw a route" button and plan your new route directly in wanderer. We use [valhalla](https://github.com/valhalla/valhalla) and their associated free [hosted service](https://gis-ops.com/global-open-valhalla-server-online/) to calculate the routes. To activate the feature make sure to set the PUBLIC_VALHALLA_URL environment variable on you wanderer-web service. See the current [docker-compose.yml](https://github.com/Flomp/wanderer/blob/main/docker-compose.yml) for a working configuration.

### Bug fixes
- Uploaded trails will now have a date if it can be parsed from the file

## v0.4.0
### Features

- Trails can now be printed to PDF. Select "Print" from the menu when viewing a trail.
- Trails can now have a date.
- You can now comment on trails
- Adds a setting to focus the map on all trails instead of a specific location

### Bug fixes
- fixed a bug that would show a wrong date for summit logs for certain time zones (now really)
- fixed a bug that prevented waypoints from showing up in public trails

## v0.3.2
### Bug fixes

- Fixed a bug that caused a 500 Internal Error to appear when viewing trails without an account

### Translations

- improves Dutch translation (thanks to [Vistaus](https://github.com/Vistaus))


## v0.3.1
### Features

- Max values for elevation gain and distance filters are now dynamically calculated based on your longest trail
- Disabling username/email & password auth in PocketBase is now reflected in wanderer's login UI

### Bug fixes

- Fixed a bug that prevented import trails from appearing in map view

### Translations

- adds Simplified Chinese translation (thanks to [icyleaf](https://github.com/icyleaf))

### Dependencies

- updates PocketBase to v0.22.7

## v0.3.0
### Features

- Trails can now be added to a list while editing or creating a trail. The trail must be saved at least once to add it to a list.
- wanderer now has an auto-upload folder. GPX files in this folder will be autmatically uploaded and converted to a trail. Read the [docs](https://github.com/Flomp/wanderer/wiki/API#auto-upload-folder) for more information.
- addded support for TCX and KML files. Note that this feature is still experimental. Please report any issues you encounter.
- added OAuth support. Read [here](https://github.com/Flomp/wanderer/wiki/OAuth) how to enable providers.
  
### Bug fixes

- fixed a bug that would show a wrong date for summit logs for certain time zones
- added client side validation for usernames

### Translations
- adds French translation (thanks to [seb2020](https://github.com/seb2020))
- adds Hungarian translation (thanks to [sszemtelen](https://github.com/sszemtelen))


## v0.2.1
### Bug fixes

- summit book dates now show in the correct format for the current locale
- fixed a bug that would overwrite trail names and descriptions when editing a trail

### Translations

- adds Dutch translation (thanks to [yves-bonami](https://github.com/yves-bonami))
- adds Polish translation (thanks to [ludrol](https://github.com/ludrol))
- adds Portuguese translation (thanks to [poVoq](https://github.com/poVoq))

## v0.2.0
### Features

- when creating/editing a trail you can now drag & drop photos into the photo section to upload them
- you can now attach photos to waypoints
- you can now edit the "Distance", "Elevation Gain" and "Est. Duration" fields
- waypoint markers can now be moved with drag & drop
- lists can now be displayed as a map showing all trails contained in the list
- you can now prevent users from signing up by setting the `DISABLE_SIGNUP` environment variable to `true`
- you can now upload GPX files via the API to create trails. Check the [documentation](https://github.com/Flomp/wanderer/wiki/API#upload-trails) for more info.
- the city index now includes states

> Note: for city states to show up in your search you have to delete your data.ms folder if you already have a previous installation of wanderer. The indices will then be rebuilt on startup.

### Docs

-  added complete API documentation

## v0.1.1
### Bug fixes

- fixed a bug that would prevent trails longer than 20km from being displayed
- added BODY_SIZE_LIMIT env variable to docker compose to allow for bigger file uploads
- fixed a bug that caused only 5 trails to be shown at a time
- fixed a bug that would cause waypoints not to be deleted from the backend
- updated the default docker-compose.yml to include a secure MEILI_MASTER_KEY
- the default location field now sets the value correctly after clicking on a search result
  
### Docs

- updated the docs to include BODY_SIZE_LIMIT

## v0.1.0 
- Initial release