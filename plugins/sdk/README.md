# Wanderer Plugin SDK for Go

TinyGo-compatible helpers for Wanderer WASM plugins.

```go
import "github.com/open-wanderer/wanderer/plugins/sdk"
```

The SDK contains only plugin-side protocol types and host-function helpers. It
does not depend on Wanderer core or PocketBase.

Common protocol types:

- `ListInput`, `ListOutput`, `TrailImport`, `Track`, `Waypoint`, `Photo`
- `RefreshSessionInput`, `RefreshSessionOutput`
- `TrailSendInput`, `TrailSendPlan`
- `HostRequestSpec`, `HostResponse`, `PluginError`

Provider HTTP requests use connector targets. Plugins provide a connector name,
a relative path, and ordered query parameters; the host owns the final base URL,
path scope, redirects, TLS, and private-network policy. Public external media
URLs remain available only through `MediaSource{Type: "url"}`.

Host HTTP request bodies support JSON, `application/x-www-form-urlencoded`, and
multipart. Use `PostJSON` for JSON and `PostForm` for ordered form fields. Any
request body, including a login form POST, is governed by manifest
`permissions.uploads.maxBytes` and `permissions.uploads.contentTypes`; in this
contract "uploads" means plugin-to-provider request bodies, not only media/file
uploads.

Set `HostRequestSpec.FollowRedirects` to `sdk.Bool(false)` when a plugin needs
to inspect a redirect response itself, for example to collect `Location` and
`Set-Cookie` during a provider login flow. `HostResponse.HeaderValues` is the
only response-header representation and preserves all values. Prefer
`FirstHeader` for scalar headers and `HeaderValuesFor` for headers that can
appear more than once.

Plugins can emit host-visible logs with `LogDebug`, `LogInfo`, `LogWarn`, and
`LogError`. Log levels are strict (`debug`, `info`, `warn`, `error`) and
messages must be non-empty. Arbitrary guest text is redacted because it may
contain provider secrets. Use `LogEvent(level, event, err)` for allowlisted
request/sync start, completion and failure events. Typed errors retain their
code, class, operation, HTTP status and any retry delay (including zero) in host logs; free-form messages, URLs,
response bodies and panic values never do. The host adds bounded cause labels
at local boundaries such as database saves, policy checks and auto-merges.
Logs are best-effort and should not be part of plugin control flow.

Provider failures use `PluginFailure`, which carries a safe `PluginError`.
`ResponseError(operation, response)` classifies a failed HTTP response, and
`SafeError(err, operation)` preserves that classification across wrapped errors.
Serialize this error with `pdk.SetErrorString` before returning a nonzero export
code. Arbitrary error strings and provider bodies are discarded.

`PluginError` retains the v1 `code` and `message` fields and adds optional
`operation`, `httpStatus`, and `retryAfterSeconds` metadata. These additions do
not require a new manifest version. Messages are generated from allowlisted
codes and provide a safe static English fallback for older hosts/UI. Current
hosts and UI localize by code; provider messages are never used. The v0.21.0
worker can still turn a guest rejection into a fatal runtime error because it
checks Extism errors before nonzero export codes. The fallback preserves the
wire message for consumers that receive it; it does not fix that older worker.
`Retry-After` accepts seconds or an HTTP date,
including zero, with a maximum delay of 24 hours. A missing delay lets the host
choose its fallback.

Use `login`, `oauth_exchange`, `oauth_refresh`, `list`, `detail`, or `send` for
the failing operation. A token refresh rejected with `invalid_grant` requires
reconnection; an unavailable provider does not. A detail request rejected with
401 or 403 does not by itself prove that the connection is invalid.

Small sync helpers are included for the repeated mechanics that every provider
needs:

- `StringField` / `StringOption`
- `IntState`
- `IntOption`
- `KnownIDs`
- `SyncLimit`
- `NextPageState`

Additional TinyGo-compatible helper packages:

```go
import sdkgpx "github.com/open-wanderer/wanderer/plugins/sdk/gpx"
import "github.com/open-wanderer/wanderer/plugins/sdk/polyline"
```

- `gpx` writes simple GPX 1.1 track documents from provider track points.
- `polyline` decodes Google-style encoded polylines and provides small helpers
  for coordinate scale normalization, coordinate swap detection, and mapping
  shorter elevation arrays onto track points.
