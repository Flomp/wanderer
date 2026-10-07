---
title: wanderer API
description: How to use wanderer's API
---

wanderer comes with a fully functional RESTful API out of the box. It largely follows the CRUD conventions implemented by the [PocketBase backend](https://pocketbase.io/docs/api-records/#crud-actions). All endpoints are available at `http://<wanderer_host>/api/v1`. The full technical API documentation can be found at `http://<wanderer_host>/docs/api/index.html` and is also provided in the [API Reference](/api-reference).

## Authentication
wanderer's API uses cookie-based authentication. To receive an auth cookie send a request to the `/auth/login` endpoint. The request must contain a JSON body in the following form:
```
{
    username: string,
    password: string
}
```
The cookie from the response can be sent in a subsequent request to authenticate it.
### Example
```bash
curl --header "Content-Type: application/json" --request POST \
--cookie-jar ./wanderer-credentials \
--data '{"username":"MyUser","password":"mysecretpassword"}' \
http://localhost:3000/api/v1/auth/login
```


## Upload trails
One common use case for wanderer's API is bulk uploading GPX files to create new trails. For that, the API provides a separate endpoint: `/trail/upload`. You must first log in to use the endpoint. Afterwards you can send a GPX file to the endpoint to let <span class="-tracking-[0.075em]">wanderer</span> parse it an create a new trail in your collection. <span class="-tracking-[0.075em]">wanderer</span> will try to infer as much information as possible from the file itself. All additional information can be added to the trail via the UPDATE [endpoint](/api-reference/operations/updatetrail).

### Example
```bash
curl --location --request PUT 'http://localhost:3000/api/v1/trail/upload' \
--header 'Content-Type: application/gpx+xml' \
--cookie './wanderer-credentials' \
--data-binary '@my_trail.gpx'
```


## Plugin connections

Create and settings writes to PocketBase's `plugin_instances` collection now
return **403**, including superuser and batch REST writes. Use the host service:
`POST /plugins/instances` and `PATCH /plugins/instances/{id}/settings` on the
PocketBase backend. The Web API retains `PUT /api/v1/plugin-instance` and
`POST /api/v1/plugin-instance/{id}` and forwards these requests to the host.
Owner reads and `DELETE /api/v1/plugin-instance/{id}` remain available.

The create body contains `plugin_id` and optional `enabled`, `config`, `auth`,
and `authContext`. The authenticated identity supplies the owner. Settings
accept the same editable fields except `plugin_id`; `user`, `status`, `state`,
read-only revisions and OAuth-internal fields cannot be written. Send a JSON
object containing only the intended changes.

```json
{
  "plugin_id": "komoot",
  "authContext": "provider_session",
  "auth": { "email": "user@example.org", "password": "new-password" },
  "enabled": false
}
```

An existing connection read exposes `auth_context`, opaque
`connection_generation`, `enabled`, connection `status`, and temporary `busy`.
Secrets are empty placeholders only when stored material exists; an absent
secret has no key. A placeholder proves material presence, not validity. Token
revisions and pending OAuth flow fields are never returned.

Active credential or context changes require the previously read generation:

```json
{
  "auth": { "password": "replacement-password" },
  "expected_connection_generation": "generation-from-the-latest-read"
}
```

Omitted credentials and empty secret strings preserve the latest stored value.
An explicit `null` removes an editable secret credential. Nonempty secret inputs are
plaintext and encrypted by the host, including a value re-entered to repair an
unreadable credential. Send `{"enabled": false}` alone to disable a plugin;
this action works even if its stored configuration is invalid.

Credential editability follows the chosen authentication context. For example,
`accessToken` is host-owned in an OAuth context even if an inactive API-key
context uses that same input name. Select the API-key context in the guarded
settings patch before replacing that shared value; returning to OAuth then
requires fresh authorization.

The host fully validates configuration on create, actual provider/plugin config
changes and a transition from disabled to enabled. Credential-only repair,
unchanged config and category-remap dismissal do not trigger that full check.
Session validation always checks effective configuration and accepts an optional
unsaved `config` draft. Existing-instance validation, OAuth start and revoke also
require `expected_connection_generation`. OAuth start accepts
`activate_on_success: true` to enable the plugin after successful authorization;
a later explicit enable/disable action overrides that activation request.

Conflict responses retain the PocketBase envelope:

```json
{
  "status": 409,
  "message": "Safe public message",
  "data": { "operation": { "code": "plugin_connection_conflict", "message": "Safe operation message" } }
}
```

Use `data.operation.code` to distinguish `plugin_instance_busy` (try deliberately
later), `plugin_connection_conflict` (reload and review),
`plugin_oauth_flow_conflict` (begin a new authorization), `plugin_instance_exists`
(open the existing instance), and `plugin_instance_disabled` (reload activation
and eligibility). A missing required generation is HTTP400 with
`connection_generation_required`. Do not automatically repeat a send or an OAuth
code exchange after a conflict.

Revoke returns `{"ok": true, "connection_generation": "committed-revision"}`.
For repair followed by a new OAuth start, pin this committed revision. A later
read may already describe another connection, so its revision must not replace
the revoke result's guard for that continuation.
