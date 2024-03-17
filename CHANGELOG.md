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