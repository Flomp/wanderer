---
title: Configuration
description: How to configure wanderer with environment variables
---

Global settings for wanderer can be adjusted via environment variables. If you depoloyed wanderer with docker you can change the environment variables directly in the `docker-compose.yaml`. If you deployed wanderer on bare-metal you can change the environment variables in the launch script.

## Common
These variables are shared between all three services.

| Environment Variable | Description                                                      | Default                                     |
| -------------------- | ---------------------------------------------------------------- | ------------------------------------------- |
| MEILI_URL            | IP or hostname (including the port) of your meilisearch instance | http://search:7700                          |
| MEILI_MASTER_KEY     | Master API key for your meilisearch instance                     | vODkljPcfFANYNepCHyDyGjzAMPcdHnrb6X5KyXQPWo |

## Meilisearch
Since we use an unmodified installation of meilisearch you can use all variables listed in meilisearch's documentation. You can find a full list over [here](https://www.meilisearch.com/docs/learn/configuration/instance_options).

| Environment Variable | Description                   | Default |
| -------------------- | ----------------------------- | ------- |
| MEILI_NO_ANALYTICS   | Disable meilisearch telemetry | true    |


## Frontend

| Environment Variable  | Description                                                          | Default                            |
| --------------------- | -------------------------------------------------------------------- | ---------------------------------- |
| ORIGIN                | Public IP or hostname (including the port) of your wanderer instance | http://localhost:3000              |
| BODY_SIZE_LIMIT       | Maximum allowed upload size                                          | Infinity                           |
| PUBLIC_POCKETBASE_URL | IP or hostname (including the port) of your wanderer instance        | http://db:8090                     |
| PUBLIC_DISABLE_SIGNUP | Disables signup option for new users                                 | false                              |
| PUBLIC_VALHALLA_URL   | Public IP or hostname (including the port) of a valhalla instance    | https://valhalla1.openstreetmap.de |
| UPLOAD_FOLDER         | Folder from which wanderer auto-uploads trails                       | /app/uploads                       |
| UPLOAD_USER           | Username for the account with which wanderer auto-uploads trails     |                                    |
| UPLOAD_PASSWORD       | Password for the account with which wanderer auto-uploads trails     |                                    |
