---
title: Authentication
description: Authentication with email/password and OAuth
---

For the majority of wanderer's features you need an account to interact with them.

## Email/Username & Password

The quickest way to create an account is by heading over to `/register` and entering a username, a valid email address, and a password of your choice.
After registering you will be redirected to the homepage and can start creating your first trail.

:::note
The username must be at least 3 characters long, the password at least 8.
:::

## OAuth2

Alternatively, wanderer supports authenticating via OAuth2. The following providers are supported:

- GitHub
- Apple
- Google
- Microsoft
- Yandex
- Facebook
- Instagram
- GitLab
- Bitbucket
- Gitee
- Gitea
- Discord
- Twitter
- Kakao
- VK
- Spotify
- Twitch
- Patreon (v2)
- Strava
- LiveChat
- mailcow
- OpenID Connect

### Prerequisites

To set up OAuth support you will need to access the PocketBase backend. Make sure to forward port 8090 of the `wanderer-db` container. Access the PocketBase admin panel in your browser at `http://<your_pocketbase_url>:8090/_/` and create an admin account.

### Create an OAuth app

This step will vary wildly from provider to provider. Please refer to your provider's documentation for the specific steps. 

No matter your provider, you will need a redirect URL. This redirect URL must have the following format: `$ORIGIN/login/redirect`. `$ORIGIN` refers to the `ORIGIN` environment variable that defines the public host at which your wanderer instance can be reached. So for the default installation, the redirect URL is `http://localhost:3000/login/redirect`. 

In any case, once you have successfully created your OAuth app you will receive a Client ID and a Client Secret.

### Enable a provider in PocketBase
![Pocketbase OAuth](../../../assets/guides/pocketbase_oauth.png)

In the PocketBase admin panel navigate to `Settings -> Auth providers`. Click the gear icon next to your provider, fill in the Client ID and Client Secret from the step before and save your changes.

### Login using OAuth
![wanderer OAuth](../../../assets/guides/wanderer_oauth.png)

That's it! You should now see your OAuth provider appear in wanderer's login form. Click the button, authorize wanderer, and wait for the authentication to finish. You are now logged in and can use wanderer like any other user.