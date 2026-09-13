# Evidence register and analysis limits

Status: **DRAFT / EVIDENCE**. This chapter distinguishes observed source behavior from architectural inference and new policy. It is not a complete security audit or a report of a newly executed integration suite.

## 1. Frozen reference

The primary reference is the combined integration commit:

```text
865251d49dc6c1d734cccb85ee51c680c62d35a6
```

It combines the then-prepared anonymous read-rule correction, immutable share-target correction, #1222 sharing changes, #1224 synchronization corrections, and #1223 checked expansions on dev `555924d8bc3ce7ca00c2fccfc76d6669e7fc5cb2`. It is an analysis fixture, not a published release, a claim about current PR heads, or a commit assumed accessible through GitHub's blob URLs.

The tag **`spec/federation-baseline-2026-09-13`** retains this integration fixture. It must be published alongside the documentation branch: pushing the documentation branch alone does not transfer this separate integration history. Publish only this named evidence tag, not unrelated local tags. The documentation branch itself contains documentation changes on `dev`; it does not merge the integration fixture's runtime changes.

After the tag has been published, readers can retrieve it from the repository hosting the proposal (called `origin` in this example):

```bash
git fetch origin tag spec/federation-baseline-2026-09-13
```

Source locators in these chapters refer to the frozen commit unless explicitly identified otherwise. The documentation branch and current `dev` are not substitutes for that combined tree, and may lack helpers added by the reviewed fixes. Relative links to repository documentation are convenient reading links; use `git show` to reproduce baseline text without changing the checkout.

Examples, each read-only:

```bash
git rev-parse spec/federation-baseline-2026-09-13
git show spec/federation-baseline-2026-09-13:db/routes/remote_list.go
git show 865251d49:db/routes/remote_trail.go
git show 865251d49:db/routes/remote_list.go
git show 865251d49:db/hooks/share_target.go
git show 865251d49:docs/src/content/docs/develop/federation.md
git show 865251d49:docs/src/content/docs/use/share-trails.md
```

Line numbers are navigation aids at that commit. Function names and requirement descriptions identify the relevant behavior if subsequent edits shift lines.

## 2. Findings and their architectural relevance

| Evidence | Observed behavior at the frozen baseline | Inference or proposed follow-up |
| --- | --- | --- |
| `db/routes/remote_trail.go:217`, `db/routes/remote_list.go:163` | Full sync derives outgoing query parameters from the triggering URL, removing the handle. Imported relations depend on returned expansion sections. | Shared public hydration needs its own fixed scope, independent of browser query/credentials; RPL-008–017. This is not a claim that every possible query currently produces an exploitable disclosure. |
| `db/routes/remote_trail.go:259`, `:482`, `db/routes/remote_list.go:210`, `:249` | File synchronization occurs inside the record transaction; failed downloads can leave missing files while sync completion is recorded. | Define data and media completion separately; move bounded network operations outside short commit transactions. |
| `db/routes/remote_list.go:207`, `:222`, `:276`, `:278`, `:281`, `db/routes/remote_trail.go:346` | List-member import runs inside the list transaction. Actor resolution there may perform network I/O. `syncTrailMetadata` imports a member summary's `public` value; absent expanded author data falls back to the list author. | List membership cannot establish the member's independent public verification or ownership. Resolve identities before the transaction and retain per-object provenance; RPL-012/033 and EVT-003. |
| `db/util/activitypub.go:560`, `db/federation/create.go:757` | For an existing list, `ListFromActivity` sets `needs_full_sync` in memory and returns. The create/update activity handler inserts a feed effect but does not save that list record. | Persist invalidation; this specific bug can be fixed and regression-tested independently of a new coordinator. RPL-027 describes the broader invariant. |
| `db/routes/remote_trail.go:304`, `db/routes/remote_list.go:239`, `db/routes/remote_sync.go` | Import removes selected protected keys, then loads a general map; shared stripping now protects local completion flags. | Use an allowlist whose field authority is explicit; new DB fields should not silently become remote-writable. |
| `db/routes/remote_trail.go:362` and its caller | Tags are resolved from `expand.tags`; the stored tag relation is assigned only when the result is nonempty. | Distinguish missing, failed and verified empty data. The current behavior does not provide complete-empty reconciliation. |
| `db/routes/remote_expand.go:12` | Checked response enrichment clears pre-existing expansion state and uses the caller's context. | Preserve this fix; separate presentation from import/index representations so its precondition remains stable. |
| `db/routes/remote_trail.go:135`, `db/routes/remote_list.go:80` | Background refresh receives a fresh record copy. | Preserve response isolation and extend generation fencing to stale persistent results. Copying alone does not define deletion/revalidation ordering. |
| `db/routes/remote_trail.go:40` onward | Cache fallback distinguishes a previously completed copy from a placeholder. | Preserve the useful distinction; define verification clocks and explicit materialization instead of expanding boolean meanings. |
| `db/hooks/share_guard.go`, `db/hooks/share_target.go`, `db/main.go:129` | Request hooks protect remote recipients and fix share targets for actor and link shares. | Local request authorization and internal federation import are separate trust boundaries. These security fixes do not by themselves require a replica redesign. |
| `db/migrations/1789200002_guard_anonymous_read_rules.go` | Nonempty-auth guards close anonymous empty-relation comparisons in collection reads. | Audit the full set of entry points, including views; a custom-route fix alone cannot protect the Records API. |
| `db/util/meilisearch.go:320`, `:416`, `db/routes/search_token.go:11` | Search is a separately populated projection with principal-dependent token filters and internal expansions. | A current authoritative output gate is required before claiming immediate suppression while indexing lags. The draft does not claim existing tokens implement that new replica gate. |
| `web/src/lib/stores/search_store.ts:84`, `web/src/routes/api/v1/search/[index]/+server.ts:42`, `web/src/routes/api/v1/search/multi/+server.ts`, `web/src/routes/api/v1/search/trails/cluster/+server.ts`, `web/src/hooks.server.ts:103`, `:149` | Browser search uses existing SvelteKit routes; the server-side Meilisearch client uses a tenant token. The same token is in a cookie with `httpOnly: false`. The reviewed source does not establish whether every deployment exposes Meilisearch directly. | Extend existing proxy enforcement rather than assume no proxy exists. Check ordinary, multi and map search including counts/facets, and close direct-token bypass only where reachable. This refines the cost assessment without making an immediate-withholding claim. |
| `web/src/routes/api/v1/files/[collection]/[record]/[file]/+server.ts:52` | The file proxy builds a PocketBase file URL and forwards bytes, with thumbnail handling. | Inventory direct PB URLs, proxy, thumbnails and caches before claiming complete file-level parent/replica enforcement. Historical file fields alone do not prove all current paths' behavior. |
| `db/federation/activity.go:55` | `PostActivity` schedules goroutines, deduplicates destination URLs and logs delivery results. | URL deduplication is valuable but is not a durable job/receipt model; propose restartable recipient work. |
| `db/federation/create.go`, `db/federation/delete.go:16`, `db/hooks/trail_share.go` | Public-only publication checks, create-share Announce and private-record deletion checks cover particular events. | Specify previous-recipient cleanup, share changes, true deletion and reversible withdrawal separately. |

Chapter 01 includes more precise S4–S13 references for child rights, profile statistics, share behavior and UI sequencing. Chapter 03 includes event-specific references. Those locators support bounded observations rather than a claim that the entire system was exhaustively reviewed.

## 3. Historical PR context

The following primary PR records were considered in the preceding analysis. Their descriptions and historical diffs explain problem families; this draft does not assert their live merge state or equate them with the frozen local tree.

| Record | Relevance to this proposal |
| --- | --- |
| [#1103](https://github.com/open-wanderer/wanderer/pull/1103) | Cached actor/object availability during remote failures; motivates explicit fallback and authority semantics. |
| [#1222](https://github.com/open-wanderer/wanderer/pull/1222) | Private remote sharing and list dialog behavior; subsequently prepared with separate target-immutability protection. |
| [#1223](https://github.com/open-wanderer/wanderer/pull/1223) | Applying relation view rules to custom-route output and preserving required public/tag/token behavior. |
| [#1224](https://github.com/open-wanderer/wanderer/pull/1224) | Local completion flags and isolation of response records from background refresh. |
| [#1232](https://github.com/open-wanderer/wanderer/pull/1232) | Anonymous collection-read guards, independently useful for a security patch release. |
| [#1052](https://github.com/open-wanderer/wanderer/pull/1052) | Duplicate feed effects across repeated object updates; motivates logical effect identity. |
| [#1056](https://github.com/open-wanderer/wanderer/pull/1056) | Multiple federation paths needed related fixes; motivates common typed boundaries while retaining entity-specific rules. |
| [#1205](https://github.com/open-wanderer/wanderer/pull/1205) | Actor/object ownership, trusted ingress and outbound-request protections; motivates explicit authority at each boundary. |

## 4. Existing documentation and remaining gaps

| Existing document | Established subject | Additional contract needed |
| --- | --- | --- |
| [Federation](../../docs/src/content/docs/develop/federation.md) | Exchanged entities/message examples and on-demand loading | State transitions, fixed snapshot scope, authority, duplicate/reordered messages, outage and withdrawal semantics |
| [Sharing trails](../../docs/src/content/docs/use/share-trails.md) | Public-only, view-only cross-instance sharing and local collaboration | Match every API path to that promise; distinguish accepted permission input from effective remote rights |
| [Lists](../../docs/src/content/docs/use/lists.md) | List sharing and independently retained trail shares | Qualify local auto-grants by ownership; explain remote private-member exclusion and partial success |
| [Statistics](../../docs/src/content/docs/use/statistics.md) | Activity sources and completed-trail fallback | Viewer/replica eligibility and trustworthy remote aggregate scope |

Documentation absence is scoped to these reviewed sources and repository paths. It is not proof that no maintainer has discussed the intended behavior elsewhere.

The protocol discussion in chapter 03 refers to the [W3C ActivityPub Recommendation](https://www.w3.org/TR/activitypub/) and [ActivityStreams vocabulary](https://www.w3.org/TR/activitystreams-vocabulary/). The proposed withdrawal extension, cache durations, retry limits, internal snapshot names and storage design are application proposals, not guarantees supplied by those standards.

## 5. Existing test evidence and later supplement

The preceding PR preparation recorded successful Go tests, vet and builds for the prepared heads and combined tree. The obsolete local preparation files and test logs were subsequently removed after the PRs had been created. Those historical results concern those fixes; they do not validate all proposed lifecycle, file, index or delivery behavior. Re-run the relevant checks against the intended implementation head when validating future changes.

Later inspected commit:

```text
ddd1929d8b36c6b115e2fd8f06884b3ddcc9a3f3
```

It adds `db/routes/remote_background_sync_test.go` and an injectable `newRemoteSyncHTTPClient` used by trail/list full sync. The test coordinates the actual route's background refresh with response access, expecting the cached response to remain stable while storage receives refreshed content. This is stronger response-isolation evidence than checking `Fresh()` in isolation and should be retained during future extraction. It does not itself supply persistent generation fencing, media revocation or a durable scheduler.

Read it without changing branches:

```bash
git show ddd1929d8:db/routes/remote_background_sync_test.go
git show ddd1929d8:db/routes/remote_sync.go
```

No application test, build, live federation exchange, benchmark, deployment or migration was newly executed for this specification-writing task. The new documents receive structural and consistency review; T-001…T-052 describe future acceptance work. Current branch heads and registered schema must be checked again before implementing a work package.

## 6. Limits and confidence

- **High confidence, bounded source findings:** query-dependent hydration, mixed map/record/output roles, independent authorization paths and the specific regression mechanisms documented above.
- **Architectural recommendation:** explicit data ownership and lifecycle contracts reduce the opportunity for the same error classes; they do not eliminate the need for careful PocketBase rules and hook registration.
- **Pending product policy:** every D-* choice, including cache/retention limits, token write scope, private profile attribution and legacy migration availability.
- **Pending engineering validation:** actual file/index bypass coverage, performance with large lists, queue/storage costs, precise peer capability and revision contracts, and mixed-version interoperability.

These limits define the next review and implementation work. They must not be silently converted into statements that the current PRs already provide the proposed guarantees.
