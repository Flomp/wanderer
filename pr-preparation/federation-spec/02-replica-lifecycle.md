# 02 — Replica lifecycle and synchronization

**Status: DRAFT — proposed behavior, not a description of a released implementation.**
Prepared on 2026-09-13 against integration commit `865251d49dc6c1d734cccb85ee51c680c62d35a6`, retained by tag `spec/federation-baseline-2026-09-13`; see [baseline retrieval](08-evidence.md#1-frozen-reference).

**Delivery scope:** The complete state machine below is a deferred P3+ option. The recommended next work is P0–P2: decisions, targeted fixtures, typed import and response isolation. These packages do not require new job tables, leases, fixed-snapshot wire support or strict cache-age enforcement.
Requirements marked MUST, MUST NOT, and SHOULD are normative within this draft, subject to maintainer acceptance.
The decision IDs refer to the specification's shared decision register; an explicit proposed default does not imply that the decision is already accepted.

This chapter defines replicas of public remote trails and lists, their metadata, associated media, and fetch jobs.
It does not introduce private cross-instance sharing, an offline guarantee, or authority to change another instance's objects.
[Chapter 03 — Events and delivery](03-events-and-delivery.md) defines authenticated owner events, withdrawal, deletion, delivery, and tombstone semantics; this chapter specifies their effect on local state.

## Observed baseline and reason for this proposal

The following are code observations, not additional requirements or claims that the current PR fixes are ineffective.

- #1103 separates a previously synchronized copy from a placeholder before permitting fallback after a temporary failure. Its two flags represent useful, distinct facts. See [fallback eligibility at `555924d8b`](https://github.com/open-wanderer/wanderer/blob/555924d8bc3ce7ca00c2fccfc76d6669e7fc5cb2/db/routes/remote_trail.go#L40-L65).
- A full sync currently forwards the incoming query, downloads files inside a transaction, and marks completion even when requested relationships are absent. See [the trail sync path](https://github.com/open-wanderer/wanderer/blob/555924d8bc3ce7ca00c2fccfc76d6669e7fc5cb2/db/routes/remote_trail.go#L234-L297).
- File imports currently ignore individual download failures and mostly populate missing local files. See [file synchronization](https://github.com/open-wanderer/wanderer/blob/555924d8bc3ce7ca00c2fccfc76d6669e7fc5cb2/db/routes/remote_trail.go#L481-L518).
- #1224 protects local completion flags from values in imported metadata. The final original PR diff contains [the explicit exclusion helper at `28e3a4474`](https://github.com/open-wanderer/wanderer/blob/28e3a44742e3021c83654ca40f87674462abedc8/db/routes/remote_sync.go). The reviewed integration also isolates background mutation from the response record.
- Activity ingestion and HTTP hydration implement separate transitions: [existing trails persist an invalidation](https://github.com/open-wanderer/wanderer/blob/555924d8bc3ce7ca00c2fccfc76d6669e7fc5cb2/db/util/activitypub.go#L174-L192), whereas [the existing-list branch returns after setting its flag](https://github.com/open-wanderer/wanderer/blob/555924d8bc3ce7ca00c2fccfc76d6669e7fc5cb2/db/util/activitypub.go#L560-L566).
- The earlier migration inferred completion from `iri != '' AND needs_full_sync = FALSE`; that is not a verified-public timestamp or proof of a fixed snapshot contract. See [the completion backfill](https://github.com/open-wanderer/wanderer/blob/555924d8bc3ce7ca00c2fccfc76d6669e7fc5cb2/db/migrations/1789200000_add_full_sync_completed.go#L10-L16).

These source links pin historical revisions underlying the reviewed integration. The local integration SHA above is the authoritative combined review baseline; the links do not imply that this draft has been implemented.
Exact combined-tree evidence is `865251d49:db/routes/remote_trail.go:40-65,125-145,234-299,482-518`, `865251d49:db/routes/remote_list.go:177-246,257-300`, `865251d49:db/routes/remote_sync.go:3-9`, `865251d49:db/util/activitypub.go:174-192,560-566`, and `865251d49:db/migrations/1789200000_add_full_sync_completed.go:10-16`.

## State model and authority

**RPL-001 — Replica identity.** A replica MUST be keyed by canonical object IRI and object type, with its owning actor IRI and authoritative origin recorded separately from its local database ID.
A list containing an object, a sharing actor, a local ID collision, or a redirect target MUST NOT silently replace that authority.
Local originals MUST remain outside the replica state machine; uncertain local/remote classification MUST be quarantined for operator review.

**RPL-002 — Independent dimensions.** Visibility evidence, materialization, and job execution MUST be stored as independent dimensions.
No combination of downloaded files, an existing database row, or a successful job alone grants read access.

| Dimension | Values | Meaning |
| --- | --- | --- |
| `visibility_state` | `unverified`, `public`, `withheld`, `deleted` | What the receiver has established about authoritative public availability. |
| `materialization` | `stub`, `metadata_ready`, `detail_ready` | Which validated data contract is stored, independently of whether it may be served. |
| Fetch job state | `idle`, `queued`, `running`, `backoff`, `failed` | Work scheduling and execution; never an access-control decision. |
| Per-media readiness | `missing`, `queued`, `fetching`, `present`, `failed`, `withheld` | Availability of one manifest entry's bytes; separate from object detail readiness. |

**RPL-003 — Visibility meanings.** `unverified` means that no accepted proof of public availability exists for the current object incarnation.
`public` records accepted origin evidence; a separate age gate can still make the replica unavailable.
`withheld` means the receiver has accepted evidence that it must stop serving the copy, without concluding permanent deletion.
`deleted` means an accepted deletion or trusted-origin 410 has created a tombstone under chapter 03.

**RPL-004 — Materialization meanings.** A `stub` contains identity and scheduling data, with no promise of usable content.
`metadata_ready` contains a validated metadata subset with provenance and explicitly recorded missing sections.
`detail_ready` means the applicable complete `TrailSnapshotV1` or `ListSnapshotV1` contract has been committed.
`detail_ready` MUST NOT mean that every media file, comment, or summit log is available offline.

**RPL-005 — Non-public storage.** `withheld` or `unverified` may coexist with `detail_ready` because historical bytes may remain in storage.
Those bytes MUST NOT be served through detail APIs, search, feeds, list expansions, thumbnails, or file URLs.
Deletion MUST retain the minimum tombstone and generation information needed to prevent resurrection; content retention and erasure follow D-11 and chapter 03.

**RPL-006 — Required local facts.** Implementations MUST persist equivalents of the following facts; exact table and column names are not prescribed.

| Fact | Owner and purpose |
| --- | --- |
| Canonical IRI, type, owner IRI, authoritative origin | Verified identity and provenance, protected against ordinary payload overwrite. |
| `visibility_state`, `materialization`, missing sections | Receiver decisions and validated coverage. |
| `local_generation` | Monotonic receiver fence for invalidation, withdrawal, deletion, and replacement of accepted state. |
| Source revision/validator and snapshot contract version | Accepted content identity; source revision ordering only where explicitly supported. |
| `last_public_verified_at` | Receiver time of accepted origin evidence of public availability. |
| `last_successful_snapshot_at` | Receiver time of the last committed or qualifying-304-reconfirmed complete snapshot. |
| `last_attempt_at`, last outcome, next eligible attempt | Operational history; none of these renew public availability. |
| Current attempt identity, captured generation, attempt count, recovery state; lease expiry only if needed | Single-process execution and restart recovery; cross-process claims are optional. |
| Media manifest revision, per-entry readiness and local blob reference | File availability and the snapshot to which the bytes belong. |

**RPL-007 — Source and local clocks.** Source `created`/`updated` values MAY be retained as content metadata but MUST NOT determine cache age, job deadlines, or local verification time. This also applies to lease expiry if the optional lease model in RPL-028 is adopted.
An actor-profile refresh, received Activity, local edit, list summary, media download, or retry MUST NOT advance `last_public_verified_at`.
A source's completion flags MUST never be interpreted as receiver completion state.

## Fixed snapshot contracts

**RPL-008 — Public synchronization scope.** The replica fetch contract MUST represent the public view of its authoritative object.
Browser authentication, user-supplied `expand`, arbitrary query parameters, and link-share tokens MUST NOT change the shared replica's hydration scope.
Any future authorized private fetch requires an independently specified, partitioned storage and authorization model; it cannot populate this public replica implicitly.

**RPL-009 — Snapshot envelope.** A normalized internal V1 decoder result MUST identify its contract version, canonical object IRI, owning actor, public availability, source validator/revision if supported, and the completeness of every required section.
This does not require legacy peers to already emit that envelope. An adapter MUST establish coverage from a documented peer contract, explicit completeness evidence or equivalent bounded consistent enumeration; a present `expand` alone is not proof. If a peer cannot supply sufficient evidence, retain its supported metadata capability rather than inventing full-detail completeness.
Validation MUST establish provenance as well as shape; a `public: true` field copied from an arbitrary sender is not sufficient evidence.
The implementation MUST use a typed, allowlisted import representation instead of bulk loading arbitrary response keys into a database record.

**RPL-010 — TrailSnapshotV1.** A complete trail snapshot MUST contain the following fixed sections, regardless of what the triggering client requested.

| Required section | Contract |
| --- | --- |
| Public core metadata | Identity, name, description, route/date/location fields, route statistics, difficulty, and public state; optional values are explicitly absent/null under the versioned schema. |
| Author | Author actor IRI and the public display fields required to render attribution; no local user ID, credentials, or private keys. |
| Taxonomy and tags | Complete category/subcategory assignment and complete tag-name set, with explicit empty values where appropriate. |
| Waypoints | Complete ordered waypoint dataset for this trail, including each waypoint's stable identity, public content, position/geometry, and media references. |
| Media manifest | Complete manifest for trail and waypoint media, including GPX, photos, and thumbnails or their derivation rules. |

The schema MUST enumerate the actual fields before implementation; this table fixes the section boundaries, not permission to import additional database fields.
Bounds and route geometry needed for the normal detail view belong to public core metadata or a declared manifest entry, never an undocumented extra expand.

**RPL-011 — Bounded trail completion.** Comments and summit logs MUST be independent paginated collections, not unbounded sections of `TrailSnapshotV1`.
Their absence or incomplete pagination MUST NOT prevent the parent trail from reaching `detail_ready`.
Each collection MUST have its own access checks, fetch progress, provenance, and update/deletion rules; permission to read the parent alone does not replace the child rules specified under D-04.

**RPL-012 — ListSnapshotV1.** A complete list snapshot MUST contain public core metadata, author identity, its media manifest, and a complete ordered sequence of publicly visible member IRIs with their public summaries.
It MUST NOT recursively fetch every member's detail snapshot or wait for member media before completing the list.
A member summary MAY seed `metadata_ready` data with list provenance, but MUST NOT independently set the member's `visibility_state` to `public` or renew its verification time.
Serving a member projection MUST also honor any receiver-known withdrawal/deletion and the member-access policy in chapter 01; output enforcement is D-07.

**RPL-013 — Visible membership.** Completeness of a list's member section means completeness of the public projection, not disclosure of private members or private counts.
Removing an IRI from that complete projection removes the list edge; it does not prove that the referenced trail was deleted.
Reordering MUST preserve the authoritative order and MUST NOT depend on local record IDs or arrival order.

**RPL-014 — Media manifest.** Each manifest entry MUST identify its parent object, stable source locator, media role, and a source version or validator where available.
MIME type, byte size, and digest SHOULD be included when the origin can supply them; downloaded content MUST still undergo receiver validation.
The manifest MUST distinguish “no media” from “manifest not supplied”; remote filenames alone MUST NOT become local file identities.

**RPL-015 — Missing is not empty.** An omitted relation, truncated page, failed nested request, unsupported field, or permission-filtered unknown section MUST NOT be treated as an authoritative empty set.
An explicit empty array with verified completeness MAY clear that section; an omitted section MUST preserve stored data while recording incomplete or unknown coverage.
A generic PocketBase response with a missing expand does not satisfy the complete V1 contract.

**RPL-016 — Complete replacement and diffs.** Destructive reconciliation MUST occur only against a validated complete snapshot for the specific section and source revision being replaced.
For a complete waypoint section, absent entries MAY be removed from that parent's projection or evicted as local cached representations; unrelated local records MUST remain untouched. Absence from this scope is not evidence of permanent deletion of the waypoint identity: reparenting or changed public visibility can produce the same result. A terminal tombstone requires independently verified deletion evidence under chapter 03.
For complete tags, explicit empty input MUST clear prior tags; for list members, complete membership replaces only that list's edges.
An incomplete snapshot MUST NOT erase complete old sections or be labeled a newer complete snapshot.

**RPL-017 — Consistent pagination.** If a required section is transported across pages, completion MUST require all pages for the same revision or a documented equivalent consistency token.
A failed page or revision change MUST leave the prior committed complete section intact and schedule another attempt.
Independent comments/summit-log pages MUST NOT cause whole-collection deletion diffs until an explicitly complete collection traversal has been established.

## Freshness and public serving

**RPL-018 — Proposed freshness defaults (D-03, acceptance required).** Set `fresh_for = 1 hour` and `max_public_revalidation_age = 7 days`.
For a known public replica, age is receiver `now - last_public_verified_at`; a missing verification timestamp is not fresh.
At age exactly one hour the replica is stale; at age exactly seven days it exceeds the public serving window.
Content refresh MUST also consider `last_successful_snapshot_at`: fresh proof of public visibility alone does not make an old incomplete or unrefreshed detail representation fresh.

**RPL-019 — Single read gate.** Every public read path MUST apply the same eligibility predicate: accepted `public` state, verification age below the configured maximum, required materialization, and any narrower per-record or child authorization.
Local-ID routes, handle/IRI routes, search, feeds, expansions, exports, and file delivery MUST NOT bypass this predicate.
Search index membership or a stored thumbnail MUST NOT be treated as independent permission to serve content; see D-07.

**RPL-020 — Fresh and stale reads.** Within the fresh interval, the required committed data SHOULD be served without synchronous revalidation.
From one hour up to seven days, eligible cached content MAY be served immediately while a single revalidation job is scheduled.
An accepted owner invalidation MUST schedule work even if the freshness interval has not elapsed; an accepted withdrawal MUST immediately stop serving regardless of age.

**RPL-021 — Maximum public age.** At or beyond seven days without renewed public verification, the known copy MUST be hidden from public results until revalidated.
This age gate MUST NOT change `public` to `withheld`, fabricate a deletion, discard the snapshot, or advance its timestamp.
A later valid public response may restore availability; repeated temporary failures MUST NOT extend the seven-day deadline.

**RPL-022 — Renewing verification.** A valid origin 200 that proves public availability MAY advance `last_public_verified_at`; `detail_ready` additionally requires the complete relevant contract.
A trusted 304 MAY renew verification only for the exact public resource, stored validator, contract, and security context previously verified, with no intervening withdrawal or deletion.
An unexplained 304, actor fetch, redirect response, incomplete unvalidated body, or Activity receipt MUST NOT renew verification.

## Remote HTTP outcomes and local responses

**RPL-023 — Trusted-origin outcomes.** The following table applies only to a request bound to the verified authoritative object endpoint after the required transport and origin checks.
Responses from an unrelated host, invalid sender, unapproved redirect, or network-policy rejection MUST NOT change visibility.

| Remote outcome | Visibility transition | Data and job treatment |
| --- | --- | --- |
| Valid public 200, complete applicable V1 snapshot | `unverified`/`public` → `public`; restoration from `withheld` subject to chapter 03 fencing | Atomically commit snapshot, set `detail_ready`, renew verification, finish job. |
| Valid public 200, only verified metadata subset | May establish `public`; does not satisfy detail completion | Commit explicitly validated subset as `metadata_ready` if no complete snapshot exists; otherwise stage it separately, preserve the complete snapshot, and queue missing work. |
| Qualifying 304 | Remain `public` | Renew only the verification covered by that validator; do not invent missing sections. |
| Timeout, DNS/connection failure, trusted 429, or 5xx | No visibility change | Keep committed data; enter `backoff`; stale serving only while RPL-019 permits it. |
| Trusted 401, 403, or 404 | `unverified`/`public` → `withheld`; never weaken `deleted` | Fence prior work, suppress public data/media, record reason; no stale fallback. |
| Trusted 410 | → `deleted` | Apply the tombstone transition from chapter 03; no stale fallback or automatic resurrection. |
| Malformed 200, wrong identity/type, unsupported contract, invalid provenance | No visibility change | Reject candidate; preserve prior data; record protocol failure and bounded retry/`failed`. |
| Other status or unapproved redirect | No visibility change unless explicitly specified | Record unsupported/transport outcome; never infer public status or deletion. |

**RPL-024 — Failure classification.** Temporary origin failures, receiver network-policy failures, validation failures, authorization withdrawal, and deletion MUST have distinct internal outcomes.
An actor lookup MUST be able to return cached identity plus a typed refresh outcome without forcing callers to discard usable identity or mistake it for fresh content authorization.
Actor resolution failures MUST NOT invalidate an already verified object; they also MUST NOT renew that object's availability window.

**RPL-025 — Local API mapping (D-10, proposed default).** The public detail API MUST return 200 only when its documented detail contract is available and the read gate passes.
For `withheld` or `deleted`, return a generic 404 without cached fields; an operator diagnostic MAY expose the internal distinction to an authorized operator.
For an incomplete or expired replica whose permitted validation attempt cannot finish, return 503 with a bounded `Retry-After` and a stable machine-readable pending/unavailable reason, without object metadata.
An existing metadata-only replica MUST NOT be returned as an empty successful trail/list detail; an explicitly documented summary endpoint may use `metadata_ready`.
Permanently unsupported detail capability is not a temporary failure: D-10 proposes `422` with `code=remote_detail_unsupported`, no `Retry-After`, and an eligible summary alternative. The coordinator MUST NOT keep retrying an unsupported operation. Adopting that code/representation requires the explicit compatibility decision.

**RPL-026 — Client independence.** A request MAY wait for a queued shared fetch within a bounded request budget, or receive the pending response defined above.
Different `expand` choices and access through a local ID versus a handle MUST NOT change synchronization scope, cache eligibility, or visibility outcomes.
Response expansion MUST be built from an immutable committed view using the caller's applicable read rules; a background job MUST NOT mutate the response record.

## Fetch jobs and concurrency

**RPL-027 — One coordinator.** Initial lookup, Activity invalidation, stale reads, operator retry, and scheduled revalidation MUST enter the same coordinator for an object's snapshot contract.
A trigger MUST persist its required transition and scheduling intent together; setting an in-memory flag without saving it is insufficient.
Triggers MUST be idempotent and MUST NOT require every incoming Activity recipient to create a separate fetch.

**RPL-028 — Single-process coalescing and restart recovery.** The proposed initial coordinator assumes one backend process with SQLite. Concurrent requests and goroutines MUST coalesce by canonical IRI, object type, contract, and public security context; only the current authorized attempt may commit. A process-local map or per-key lock is sufficient for coalescing within that process.
If durable synchronization is adopted in P3+, required work and progress MUST also survive restart. A minimal implementation may persist pending work, reclaim interrupted attempts at exclusive process startup, and protect commits with local generation/attempt checks. The in-memory map itself need not survive restart; loss of the persisted work intent is the failure to prevent.
Metadata/detail requests MAY raise the required coverage of existing work rather than launch competing writes; media work is keyed per manifest entry. Cross-process expiring leases and renewal are required only if multiple processes are deliberately supported, or if concurrent reassignment otherwise permits an older attempt to keep running. They are not an initial infrastructure requirement. Later references to an execution claim mean the single-process attempt identity unless the lease variant is chosen.

| Job event | State transition | Required action |
| --- | --- | --- |
| First trigger | `idle` → `queued` | Persist required coverage, generation, and next eligible attempt. |
| Duplicate trigger | `queued`/`running`/`backoff` → same state | Coalesce work; preserve backoff and record any newer invalidation. |
| Worker starts eligible job | `queued`/`backoff` → `running` | Establish current attempt ownership and capture generation and expected source revision. |
| Accepted success | `running` → `idle`, or `queued` if additional coverage is needed | Commit under fences and clear/reset the completed attempt. |
| Temporary failure | `running` → `backoff` | Persist failure class, attempt, retry time, and release execution claim. |
| Non-retryable or repeatedly invalid response | `running` → `failed` | Preserve data; record actionable diagnostic; no busy retry loop. |
| Exclusive process restart after interrupted work | `running` → `queued`/`backoff` | Recover persisted intent before accepting new work; no previous process remains a writer. |
| Optional lease expires in a deployment supporting reassignment | `running` → `queued`/`backoff` | Reclaim with a new claim; the older attempt can no longer commit. |
| Accepted withdrawal/deletion | any → `idle` | Invalidate/cancel obsolete work; future revalidation follows chapter 03. |
| Authorized operator retry or accepted newer authoritative evidence | `failed` → `queued` | Record reason and a fresh scheduling decision without bypassing visibility fences. |

**RPL-029 — Bounded retries (D-08).** Retry policy MUST use exponential backoff with jitter, an origin-level concurrency limit, and bounded per-attempt network time and response size.
Valid `Retry-After` values on 429 SHOULD be honored within operator safety limits; repeated reads MUST NOT reset backoff.
Exact retry counts, durations, concurrency, and any optional lease intervals remain D-08 choices; the proposed default is bounded background retry, never immediate unbounded retry.

**RPL-030 — Cancellation and attempt ownership.** Workers MUST have a job deadline and MUST terminate or lose commit authority after cancellation, superseding reassignment, or expiry of a lease where that variant is used.
Cancellation of one browser request SHOULD detach that waiter rather than cancel useful shared work for other readers.
A background context without a deadline MUST NOT permit a stalled peer to hold a job or connection indefinitely.

**RPL-031 — Local generation fence.** Every candidate result MUST carry the `local_generation` captured for its job.
The receiver MUST advance the generation on accepted invalidation that supersedes in-flight data, withdrawal, deletion, a local protection change, or acceptance of a new snapshot version. Reconfirmed unchanged data need not create a new content generation. Existing object authority remains immutable; identity replacement requires a separately authorized new object under chapter 03.
Commit MUST check current generation and execution ownership under the coordinator's synchronization and database transaction; a mismatch discards the candidate instead of overwriting current state. If cross-process leases are introduced, their token must also be checked atomically in that transaction.

**RPL-032 — Revision fence.** If the source provides an ordered revision, the receiver MUST reject revisions older than the accepted or required source revision.
Opaque ETags MUST be used as equality validators, not lexically ordered version numbers; source wall-clock timestamps MUST NOT be assumed to provide reliable ordering.
When a known newer authoritative revision cannot be reconciled with the returned candidate, the job MUST preserve current state and retry or fail explicitly.
Local generation fencing remains mandatory even where source revision ordering exists.

## Fetch, validate, and commit

**RPL-033 — No network I/O in database transactions.** Object requests, actor resolution, pagination, and media downloads MUST occur outside write transactions.
Workers MUST first stage a candidate and resolve required identities; the commit transaction performs only local validation, mapping, writes, and durable scheduling of follow-up effects.
Media bytes SHOULD be fetched separately from metadata so a slow photo host cannot hold an object transaction open.

**RPL-034 — Protected local fields.** Remote input MUST NOT set local IDs, local actor-user associations, access-control grants, verification timestamps, generations, job state, local file references, completion flags, or index bookkeeping.
Payload fields outside the versioned allowlist MUST be ignored or rejected according to schema validation, not implicitly imported when a new database column appears.
Remote content metadata and local operational metadata MUST be separate at the import boundary even if stored in the same physical database.

**RPL-035 — Atomic snapshot commit.** A commit MUST recheck identity, authority, execution ownership, generation, revision, completeness, and current visibility restrictions.
It MUST atomically replace the accepted snapshot sections, reconcile permitted relation diffs, update materialization/verification facts, and persist index/media work intents.
A failed transaction MUST leave the previously committed snapshot and public-verification timestamp intact.
Public responses MUST observe either the prior committed version or the new one, never a mixture assembled from an in-flight mutable record.

**RPL-036 — Derived effects.** Search updates, notifications, and storage cleanup MUST be retryable, idempotent effects of a committed transition.
Failure of a search service MUST NOT falsely mark the origin snapshot incomplete; failure to remove an index entry MUST NOT bypass the read gate.
A stale fetch or stale index job MUST NOT republish data after a newer withdrawal, tombstone, or accepted generation.

**RPL-037 — Partial success.** A valid metadata subset MAY be retained with explicit section coverage, but MUST NOT advance `last_successful_snapshot_at` for a complete snapshot.
If a complete snapshot already exists, partial newer data MUST be staged separately until it can replace a coherent snapshot; it MUST NOT silently mix new core fields with old relationships in a supposedly complete version.
Independent failure of comments, summit-log pagination, or media MUST update those resources' own state rather than relabel the parent snapshot as complete or deleted.
A transient failed refresh MUST preserve the last complete snapshot's materialization even when its visibility or age makes it temporarily unservable.

## Media and offline behavior

**RPL-038 — Proposed media default (D-02).** Cache media lazily: fetching a detail snapshot MUST fetch the manifest, not necessarily all bytes.
UI and API contracts MUST distinguish detail availability, individual file availability, and an explicitly requested offline download.
Any offline action MUST report completion only for its declared manifest revision and required entries; parent `detail_ready` is not enough.

**RPL-039 — Media transitions.** A manifest entry begins `missing`, becomes `queued`/`fetching` when requested, and becomes `present` only after verified bytes are durably stored.
A failed download becomes `failed` with retry metadata; it MUST NOT remove previously valid bytes for a different committed manifest revision before replacement is safe.
A changed source version requires a new validated file; “a filename is already present” is not evidence that its bytes match the new manifest.

**RPL-040 — Media read gate and fencing.** Ordinary media delivery MUST apply the current parent's and child's authorization/visibility gates even when the blob is already cached. Any accepted D-04 author-maintenance exception must use a separately scoped authorized representation and cannot disclose hidden parent or third-party media.
After withdrawal, readiness becomes effectively `withheld` and active downloads MUST lose publication authority under the same generation fence.
A late successful download MUST NOT recreate a public URL, mark an expired replica current, or undo deletion.
Media commit MUST also match the current manifest and entry version; a normal content/media update can supersede a download even without withdrawal.
Physical cleanup and limits on already distributed URLs or downloaded files are covered by D-07, D-11, and chapter 03.

## Owner events, withdrawal, and tombstones

**RPL-041 — Event effects.** Authenticated owner Create/Update events MAY seed validated metadata and MUST durably queue appropriate revalidation.
They MUST NOT alone renew `last_public_verified_at`; the accepted event's authority and revision constrain subsequent fetches.
Unknown or invalid Activities MUST NOT withdraw, delete, or reset an existing replica's generation; sender-controlled retry noise is not a revocation mechanism.

**RPL-042 — Public-to-private transition (D-05).** For a non-deleted object, a verified negotiated `wdr:Withdraw` extension, if accepted in chapter 03, MUST immediately set `withheld`, advance the generation, and suppress object and derived media access. If the identity is already `deleted`, record the accepted receipt without weakening its tombstone or scheduling restoration; deletion remains terminal.
This draft MUST NOT encode withdrawal as an undocumented sparse Update; extension negotiation and the complete event contract belong to chapter 03.
For legacy peers without that extension, trusted-origin 401/403/404 revalidation and the seven-day maximum age are the proposed receiver fallback, not a claim of immediate revocation.

**RPL-043 — Deletion wins.** A trusted-origin 410 or authenticated owner Delete MUST create/update a tombstone and fence any earlier snapshot, media, or index job.
An ordinary 200, cached Activity replay, or Announce MUST NOT automatically resurrect a deleted incarnation.
The proposed default permits re-publication of a merely withheld object after fresh authoritative public verification, but a truly deleted IRI is terminal and must not be reused; a new publication requires a new IRI. An unrelated list retaining an old IRI is never restoration evidence.

## Legacy migration and rollback

**RPL-044 — No invented historical proof (D-09).** Migration MUST NOT infer current public verification from `public`, `updated`, local file presence, `needs_full_sync`, or `full_sync_completed` alone.
In particular, older imports may have accepted foreign completion flags, and the earlier completion migration was a heuristic backfill.
Migration MUST NOT set `last_public_verified_at` to the migration time for convenience.

**RPL-045 — Proposed conservative mapping (D-09, acceptance required).** Preserve useful legacy bytes while separating them from authorization and completeness evidence.

| Legacy row | Proposed initial state | Required follow-up |
| --- | --- | --- |
| Confirmed local original | Remains local; no replica conversion | Preserve existing local identity and rules. |
| Remote placeholder, either legacy completion flag value, no usable metadata | `unverified`, `stub`, no verification time | Queue bounded initial validation on demand or migration budget. |
| Remote row with usable metadata, `full_sync_completed = false` | `unverified`, at most `metadata_ready` | Retain bytes; validate origin and fixed sections. |
| Remote row with `full_sync_completed = true`, regardless of `needs_full_sync` | `unverified`, at most `metadata_ready` | Do not trust the flag; independently establish visibility and V1 completion. |
| Independently auditable origin verification and section coverage | Preserve only facts actually supported, using original timestamps | Apply age gate; promote detail only if the complete V1 contract is demonstrably satisfied. |
| Existing trustworthy withdrawal/deletion evidence | `withheld`/`deleted` with a generation fence | Preserve suppression and tombstone policy. |
| Conflicting origin, owner, or local/remote identity | Quarantined, publicly unserved | Operator resolution; never guess authority from an ID match. |

**RPL-046 — Migration rollout.** State backfill MUST be local and resumable, with no outbound fetch in the migration transaction.
Activation MUST ensure that every exposed remote row has initialized state and is covered by the new gate; half-migrated rows MUST NOT fall through to old public behavior.
Operators MUST receive counts of quarantined, unverified, stale, and ready rows and an estimated revalidation workload, without dumping private content into logs.
At strict activation the proposed default has no automatic grandfather period: unverifiable legacy replicas remain hidden until validation, including during an origin outage. A separately approved, time-limited observation/compatibility phase may precede strict activation under D-09, but cannot claim the strict guarantee or invent verification timestamps.

**RPL-047 — Compatibility flags.** During a transition, legacy flags MAY remain as derived compatibility fields, but MUST NOT remain independent authorities for serving or synchronization.
A compatibility value equivalent to “completed” MAY be produced only after local acceptance of the fixed snapshot contract; incoming values remain ignored.
No mapping of the two old booleans can fully encode withdrawal, deletion, verification age, materialization, and job status.

**RPL-048 — Rollback limitation.** Rolling back to a binary that lacks visibility and generation gates MUST NOT be advertised as a safe automatic schema downgrade.
Before such a rollback, remote serving MUST be disabled or equivalent suppression retained, especially for withheld, deleted, expired, and unverified replicas.
Replaying a pre-withdrawal database backup MUST NOT erase newer tombstones or silently restore public serving; operators need a documented restore/reconciliation procedure under D-11.

## Acceptance scenarios

**RPL-049 — Required lifecycle coverage.** Implementation acceptance MUST exercise the following scenarios through the coordinator and public read paths, with durable state assertions.
Tests SHOULD use a controllable clock, origin responses, worker barriers, and restarted workers; arbitrary sleeps are not evidence of correct ordering.

| Scenario | Required result |
| --- | --- |
| First Activity or list-summary discovery | Stable replica identity; no unverified public detail; work queued once. |
| Complete trail snapshot with media downloads failing | `public` and `detail_ready`; failed/missing media visible as independent state; no offline-success claim for a scope whose mandatory files are missing. Optional photos do not block the accepted offline-route scope. |
| Same trail requested with different expands | Identical fixed snapshot coverage and verification semantics. |
| Missing tags/waypoints versus complete empty sections | Missing preserves old data; complete empty clears only the corresponding authoritative section. |
| Multi-page snapshot interrupted or revision changes | No destructive partial replacement and no false detail completion. |
| Complete list with unavailable member origins | List completion does not recursively hydrate members; member visibility is not invented. |
| Timeout/429/5xx at two hours, then seven days | Eligible stale copy at two hours; no public fields at seven days without renewed proof. |
| 304 without a matching verified representation | No freshness renewal and no fabricated sections. |
| Trusted-origin 403 or 404 after prior success | Immediate `withheld`, including search/file gates; cached public body is not returned. |
| Withdrawal/Delete while snapshot or file fetch is running | Generation changes; late results cannot restore public state or bytes. |
| Forged event or unrelated-host response | No visibility transition or tombstone creation for the existing object. |
| Concurrent readers/goroutines and process restart | One current attempt per key; persisted work recovers after restart; superseded attempts cannot commit. Cross-process lease expiry is tested only if that variant is supported. |
| Newer source revision followed by a slower older response | Older candidate rejected; accepted revision and generation remain intact. |
| Database commit or search indexing fails | No partial snapshot; indexing failure cannot bypass the read gate or rewrite visibility evidence. |
| Legacy flags claim completion but no verification exists | Remains unverified after migration; no timestamp invented and no silent public fallback. |
| Detail by local ID, handle, search result, expansion, or file URL | Same applicable visibility/age gate on every route. |

## Decisions requiring maintainer acceptance

**RPL-050 — Decisions are explicit configuration boundaries.** Operators MUST NOT be allowed to bypass authority, generation, completeness, or visibility rules by tuning ordinary cache settings.
Where a setting changes the exposure policy, its effective value and consequences MUST be documented and testable.

| Decision | Proposed default in this chapter | Still requires agreement |
| --- | --- | --- |
| D-02 detail/media | Fixed snapshots; lazy media; independent paginated comments/summit logs | Exact V1 field schemas and offline-download contract. |
| D-03 freshness | Fresh for one hour; public revalidation required within seven days; no unbounded stale serving | Whether operators may shorten only or also lengthen the maximum; proposed safe default is shortening only until approved. |
| D-04 child rights | Ordinary child output remains gated; any narrow author-maintenance representation is explicit | Which own contributions remain readable/editable after parent restriction without exposing hidden parent content. |
| D-05 withdrawal | Immediate suppression for a negotiated verified extension; legacy origin revalidation otherwise | Extension adoption, restoration proof, and peer compatibility in chapter 03. |
| D-07 files/search | Apply the same live visibility gate to every derived serving surface | Signed/cache URL strategy and practical revocation limits. |
| D-08 work limits | Persisted work with single-process recovery, bounded retries/backoff, no network inside transactions | Concrete timeouts, retry limits, fairness and concurrency; optional leases only for a justified execution model. |
| D-09 legacy backfill | No blind trust in old flags; unverifiable copies hidden until validated | Operational rollout budget and availability impact; any exception needs explicit provenance rules. |
| D-10 API compatibility | 200 only for available contract; generic 404 for withheld/deleted; 503 for pending/expired validation | Response envelope, compatibility period, and frontend handling. |
| D-11 retention/deletion | Preserve suppression fences across cleanup and restore | Tombstone retention, physical erasure, offline copies, and rollback procedure. |
| D-12 resource budgets | Bound response sizes, pagination, media and concurrent jobs | Initial values in the decision register require measurement on representative data. |

Approval of these defaults should precede a migration that changes existing users' offline availability.
This draft specifies the reviewable target; implementing it is separate work from the immediate PR security and synchronization fixes.
