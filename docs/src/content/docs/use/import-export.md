---
title: Import/Export
description: How to import and export trails in wanderer
---

## Import

<span class="-tracking-[0.075em]">wanderer</span> supports bulk uploading of trails via an auto-upload folder. A cronjob fetches all files from this folder and uploads them automatically every 15 minutes. This feature is currently only available for docker installations. If you want to replicate it in a bare metal installation you will need to create your own cronjob using the `web/cron.sh` script.

:::caution
Successfully uploaded files will be deleted from the auto-upload folder.
:::

:::note
Currently only GPX files are supported.
:::
### Configuration

The following environment variables must be present in the `<span class="-tracking-[0.075em]">wanderer</span>-web` docker container and set to valid values.

| Environment Variable | Description                                                            | Default      |
|----------------------|------------------------------------------------------------------------|--------------|
| UPLOAD_FOLDER        | Path to the auto-upload folder                                         | /app/uploads |
| UPLOAD_USER          | Username of the account that will be the author of the uploaded trails |              |
| UPLOAD_PASSWORD      | Password of the account that will be the author of the uploaded trails |              |

### Volume
Make sure to mount the upload folder as a volume to your host system. The default `docker-compose.yml` already includes this volume. Ensure that the mapped value matches the one in the `UPLOAD_FOLDER` environment variable.

### Manually run the upload job
In case you do not want to wait until the next scheduled execution you can also run the job manually:

```bash
docker exec -it wanderer-web run-parts /etc/periodic/15min
```

## Export

To export a single trail head over to `/trails` and select the trail you want to export. From the <span class="inline-block w-8 h-8 bg-primary rounded-full text-center text-white">⋮</span> menu select "Export". You can export the route data either in GPX or in GeoJSON format. Furthermore, you can choose whether you want to include the photos and the summit book of the trail. In any case, <span class="-tracking-[0.075em]">wanderer</span> will create a ZIP archive with all the data that is then downloaded.

You can also export all of your trails at once. To do so, head over to `/settings/export` and click "Export all trails". The other steps remain analogous to exporting a single trail.
