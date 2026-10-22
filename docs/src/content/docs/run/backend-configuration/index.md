---
title: Backend configuration
description: How to access the PocketBase backend
---

For many configuration options, it is necessary that you are able to access the PocketBase backend. 
PocketBase comes with a handy dashboard that allows you to configure basically everything in the backend.

If you are using docker make sure to forward the internal port 8090 to a public port. 
With the default configuration, the PocketBase admin panel is available at `http://localhost:8090/_/`. 
If this is your first time visiting the panel you will need to create an admin account.
To create backend access navigate to the location of your `docker-compose.yaml` file on the server and type:

```sh
docker compose exec -it db /pocketbase superuser upsert email@example.com myverysecurepassword
```

Via the online dashboard, you will now have access with the user "email@example.com" and the password "myverysecurepassword" to all tables in the backend and can modify the underlying data directly.

## Visitor IP addresses

The web service forwards the visitor address resolved by SvelteKit to PocketBase
in `X-Forwarded-For`. In the PocketBase dashboard, open **Settings → Application**
and set **User IP proxy headers** to `X-Forwarded-For` so IP rate limits use the
visitor address instead of the web service address. See PocketBase's
[reverse proxy guidance](https://pocketbase.io/docs/going-to-production/#using-reverse-proxy).

Where PocketBase trusts this header, restrict backend access to the web service
and other trusted proxies using private networking or firewall rules. The
default `docker-compose.yml` publishes `8090:8090`; remove or firewall that
port before enabling **User IP proxy headers**. Any proxy that accepts public
requests must overwrite untrusted forwarding headers with an authoritative
client address. A publicly reachable backend that trusts caller-supplied
headers allows callers to spoof their rate-limit identity.

For the web service's own reverse proxy configuration, see
[Visitor IP addresses](/run/environment-configuration#visitor-ip-addresses).
`POCKETBASE_PROXY_SECRET` authenticates inbound federation delivery; it does not
authenticate IP headers or make a proxy trustworthy.

For specific configuration guides see:

- [SMTP settings](./smtp/)
- [Auth providers](./auth-providers/)
- [Backup server](./backup-server/)
- [Custom categories](./custom-categories/)
- [Adjust Filesize Limits](./adjust-filesize-limits/)
- [Mobile app support](./mobile-app/)
- [Region catalogue](./region-catalogue/)

To learn more about what you can do in the admin dashboard please refer to PocketBase's [documentation](https://pocketbase.io/docs/).
