---
title: Integrations
description: How to set up third-party integrations with wanderer.
---

You can automatically sync trails to wanderer at regular intervals using the third-party integration feature. Currently, we support two providers: **strava** and **komoot**.

It is important to note that synchronization only works from the provider to wanderer and not the other way around. Additionally, if a trail has already been synced to wanderer, subsequent changes made in the provider will not be transferred unless the trail is deleted in wanderer.

## strava Integration

### Creating an App in strava

Before integrating strava with wanderer, you need to create an API application in strava. Visit [strava's API settings](https://www.strava.com/settings/api) and follow the steps to create a new API application. Your setup should resemble the following:

![strava API Application](../../../assets/guides/strava_api_app.png)

### Setting Up the Integration

1. Copy the **Client ID** and **Client Secret**.
2. Go to the integrations page in wanderer's settings.
3. Click the settings button for the strava integration.
4. Enter your **Client ID** and **Client Secret**.
5. Choose whether you want to sync routes, activities, or both.

![wanderer strava Integration](../../../assets/guides/wanderer_integration_strava.png)

6. Save the settings and toggle the integration on.
7. You will be redirected to strava's authorization page. Keep all checkboxes selected and click **Authorize**.
8. You will then be redirected back to wanderer. The strava integration is now active.

## komoot Integration

The komoot integration requires only your komoot username and password:

1. Open the komoot settings from the integrations menu.
2. Enter your komoot credentials.
3. Save the settings.
4. Toggle the integration on. It will become active immediately.

Your planned and completed trails will now sync with wanderer.

## Sync Interval

By default, trails are synced every night at **02:00 AM**. You can modify this schedule using the `POCKETBASE_CRON_SYNC_SCHEDULE` [environment variable](/run/environment-configuration#pocketbase).

:::note
Please set a reasonable sync interval. Both strava and komoot impose usage limits on their APIs. Exceeding these limits may result in rejected requests or account suspension.
:::

