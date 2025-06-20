---
title: Backend configuration
description: How to access the PocketBase backend
---
## Setup

For many configuration options, it is necessary that you are able to access the PocketBase backend. PocketBase comes with a handy dashboard that allows you to configure basically everything in the backend.

If you are using docker make sure to forward the internal port 8090 to a public port. With the default configuration, the PocketBase admin panel is available at `http://localhost:8090/_/`. If this is your first time visiting the panel you will need to create an admin account. 
To create backend access navigate to the location of your `docker-compose.yaml` file on the server and type:
```
docker compose exec -it db /pocketbase superuser upsert email@example.com myverysecurepassword
```
Via the online dashboard, you will now have access with the user "email@example.com" and the password "myverysecurepassword" to all tables in the backend and can modify the underlying data directly.

## Configure SMTP settings 

wanderer can send email notifications to users (e.g. when a user gains a new follower). This is also relevant to send password reset notifications. To enable sending email, you need to configure your SMPT settings in PocketBase. 

![Pocketbase Mail Settings](../../../assets/guides/pocketbase_mail_settings.png)

In the pocketbase admin panel go to Settings -> Mail settings an enable "Use SMTP mail server". Enter the details of your SMTP server and send a test email to ensure your configuration is correct. On the same page you can also adjust the email template of the password reset email.

Alternatively, you can set these options via the respective [environment variables](/run/environment-configuration/#pocketbase).

## OAuth

### Create an OAuth app

This step will vary wildly from provider to provider. Please refer to your provider's documentation for the specific steps. 

No matter your provider, you will need a redirect URL. This redirect URL must have the following format: `$ORIGIN/login/redirect`. `$ORIGIN` refers to the `ORIGIN` environment variable that defines the public host at which your wanderer instance can be reached. So for the default installation, the redirect URL is `http://localhost:3000/login/redirect`. 

In any case, once you have successfully created your OAuth app you will receive a Client ID and a Client Secret.

### Enable a provider in PocketBase
![Pocketbase OAuth](../../../assets/guides/pocketbase_oauth.png)

In the PocketBase admin panel navigate to the `users` table. Click the gear icon at the top to open the table's settings and navigate to `Options`. In the tab `OAuth2`, add your provider and fill in the Client ID and Client Secret from the step before and save your changes.

## More options

To learn more about what you can do in the admin dashboard please refer to PocketBase's [documentation](https://pocketbase.io/docs/).