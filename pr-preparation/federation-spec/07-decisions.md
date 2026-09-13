# Decision register

Status: **DRAFT**. These are recommendations for maintainer review, not decisions already accepted by the project. `EXISTING` in another chapter describes documented or observed baseline behavior; it does not mean this complete proposal has been approved.

Use this register to record an outcome, rationale and responsible reviewer before implementing a behavior change. The current value of every new decision is **pending**. Pure extraction/refactoring can proceed independently when it preserves existing behavior and tests.

The recommended next scope is P0–P2: resolve D-01–03 as the direction for the work and preserve current behavior during extraction. Choosing a future freshness/completeness policy does not activate it in P2. Other decisions remain deferred unless an explicitly selected change touches them; this register is not a requirement to resolve all 16 before useful refactoring can start.

| ID | Topic | Recommended direction | Acceptance needed before |
| --- | --- | --- | --- |
| D-01 | Scope of federation | Public remote content and read-only targeted sharing remain v1 | Treating any new remote grant as supported |
| D-02 | Meaning of synchronized/offline | Fixed snapshots; media readiness separate | Changing completeness flags or UI claims |
| D-03 | Cache freshness and access evidence | Refresh after 1h; suppress unverified public access after 7d | Enforcing the new stale limit |
| D-04 | Child authors after parent restriction | Preserve narrow access to one's own contribution; no inherited parent access | Changing comment/log/waypoint rules |
| D-05 | Reversible withdrawal signaling | Negotiated Wanderer extension; legacy revalidation fallback | Publishing a wire extension |
| D-06 | Targeted share lifecycle | Explicit update/Undo behavior for recipients; no content permission implied | Changing outbound effects on share updates/deletes |
| D-07 | Files/search/aggregates | One authoritative access and replica gate on every output | Claiming end-to-end withholding guarantees |
| D-08 | Delivery durability and retries | Durable recipient jobs; bounded retries | Replacing current delivery scheduling |
| D-09 | Legacy replica backfill | Do not invent verification timestamps from old flags | Migrating stored replicas |
| D-10 | HTTP and payload compatibility | Separate refactor from response-code/shape changes | Shipping new API-visible behavior |
| D-11 | Retention and deletion | Hide immediately, purge bytes separately, retain minimal tombstones | Scheduling destructive cleanup |
| D-12 | Resource budgets | Explicit bounded fetch/pagination/media/queue work | Enabling new snapshot/worker paths |
| D-13 | Token write semantics | Promise scoped reading; separately decide any token-authorized mutation | Changing link-share behavior or UI promises |
| D-14 | Profile privacy and attribution | Separate profile discovery from independently public objects | Changing actor/profile exposure |
| D-15 | Remote aggregate statistics | Require eligibility-verifiable data or withhold unsupported totals | Applying local access guarantees to remote totals |
| D-16 | Local list auto-grants | Explicit server command with authoritative membership and atomic local changes | Replacing the current client request sequence |

## D-01 — Public federation scope

**Existing evidence:** The sharing guide explicitly restricts cross-instance trail sharing to public content and view-only interaction. Local private shares and link tokens are different access mechanisms.

**Recommendation:** Keep that scope for this series. Authenticate ActivityPub authors and actions, but do not interpret a signature or a targeted Announce as authorization to fetch a private trail. A local edit grant does not authorize modifying the original on another instance.

**Alternative:** Private federation would require recipient-bound fetch authorization, recipient isolation in storage, update and revoke semantics, file access, and compatibility negotiation. It is a separate design effort, not a relaxation of the current guard.

**Consequence:** The architecture can use a shared anonymous/public replica without mixing viewer-specific content into it. Existing local private sharing remains supported.

## D-02 — Snapshot and offline semantics

**Recommendation:** `detail_ready` means a complete fixed data scope, not every file and historical discussion. Trail detail includes core metadata, taxonomy/tags, the waypoint scope and media manifest. List detail includes ordered public member identities and summaries. Comments and summit logs are independent paginated resources. A list does not recursively hydrate every trail.

Primary route-file readiness and photo readiness are tracked separately. “Available offline” must identify the supported action: reading a cached description, displaying route geometry, opening the GPX or viewing all media are not equivalent promises.

**Alternative:** Require every media file for `detail_ready`. This makes a missing optional photo block useful metadata and route access, and can prevent completion indefinitely.

**Review question:** Which exact media are mandatory for the product's “offline route” indicator? Recommended minimum: current primary route/geometry and required navigation data; optional photographs do not block it.

## D-03 — Public verification and stale cache

**Recommendation:** Default refresh eligibility begins at **1 hour**. A previously verified public replica can serve stale content after transient failure while its last public verification is less than **7 days** old. At or beyond 7 days it is suppressed until verification succeeds. These are proposed product defaults, not current guarantees or ActivityPub requirements.

An authenticated origin refusal or deletion is handled immediately, regardless of age. The 7-day bound covers unknown changes during lack of successful verification, not permission to ignore a known withdrawal. Jobs, Activities, actor refreshes, list summaries and local writes cannot restart this clock.

**Alternatives:** A shorter bound reduces unknown-privacy exposure but makes outages more disruptive. An indefinite stale cache preserves availability but cannot offer bounded cooperation with origin visibility changes.

**Consequences:** Permanently abandoned remote servers eventually disappear from normal remote-content views. Users needing an independent durable copy require a separate explicitly attributed local-copy/export feature; this proposal does not silently convert replicas into local originals.

**Configuration:** Refresh and maximum-age values must have clear operator documentation and UI semantics. No automatic fallback to “infinite” on parse errors. Disabling the upper bound would be an explicit policy override outside the default conformance profile.

## D-04 — Contributions on a now-inaccessible parent

**Existing complication:** Some child read rules let a local author access their own comment or summit log independently of the parent's public state. Blanket parent-gating would silently remove those rights.

**Recommendation:** Preserve a narrow contribution-management view for the authenticated local author, with only their content and minimal reference to the parent. It must not expand hidden parent metadata, siblings, other authors' files or activity. Editing/deletion of one's own contribution must remain tied to its original parent and the established write policy.

**Alternative:** Gate all child reads by parent access, including author management. This is simpler, but needs an explicit product decision and a recovery/export path for authors.

**Still open:** Exact permitted edits after parent restriction, retention after true parent deletion, and administrator moderation workflows. The draft must not broaden existing edit privileges while resolving this question.

## D-05 — Withdrawal is not deletion

**Recommendation:** A local public-to-private transition immediately stops local public serving. For cooperating Wanderer peers, design a negotiated owner-authenticated withdrawal signal carrying identity/version evidence without private content. Chapter 03 proposes the protocol boundary; namespace and capability syntax need a dedicated interoperability review.

Generic ActivityPub peers continue to rely on origin revalidation and their own behavior. The application must not claim it can recall information already received or copied by another server.

**Rejected shortcut:** Do not emit an actual `Delete` merely to make an object private and later reuse that identity. Do not label a sparse partial `Update` as a standards-defined withdrawal. The ActivityPub server-to-server Update model is discussed in [03-events-and-delivery.md](03-events-and-delivery.md).

**Alternative to evaluate:** A complete redacted Update representation might work with specific peers. It needs explicit field and compatibility tests; it is not assumed safe or universally understood in this draft.

**Release consequence:** Reliable local/remote-read gating can ship before a withdrawal extension. The absence of that extension must remain a documented limit, not be hidden by a misleading success notification.

**Scope/cost:** Defer the extension beyond P0–P2. Its additional benefit is limited to capability-aware Wanderer peers and must justify specification, version negotiation and interoperability maintenance. Adopting the small refactor does not commit the project to developing this vocabulary.

## D-06 — Targeted sharing and recipient changes

**Recommendation:** Keep local grants and remote recommendations separate in semantics, even if they initially share a collection. Creating a remote share produces one targeted Announce. Changing the remote recipient withdraws the previous recommendation and creates one for the new recipient. Deleting that share cancels the recommendation, not the public source object. Repeating an unchanged update must not repeat notifications.

**Existing behavior difference:** Current update guards validate requests but do not announce updates. This proposal would add effects and therefore needs separate acceptance and tests.

For a remote recipient, recommend requiring `permission=view` on new or changed shares and rejecting `edit` explicitly. Older stored `edit` values must not be represented as effective remote editing rights; repair/migration treatment needs a release note. A legacy private remote share repaired to a local recipient can retain whatever local permission is actually authorized. This input-validation change is not already supplied by the current remote guard.

For public lists, local auto-grants apply only to the sharer's own private trails. Remote recipients receive no private auto-grants. Removing a list share does not silently remove independent trail grants, matching the existing guide.

**Open product detail:** Whether later-added private members receive automatic grants to existing local list recipients. Recommended first scope: preserve current explicit sharing behavior; treat continuous propagation as a separate feature with grant provenance and revocation rules.

## D-07 — Complete output coverage

**Recommendation:** Evaluate object and replica eligibility at ordinary reads, Records API expansions, file downloads, thumbnails, search, feeds, statistics and aggregate views. A successful index removal must not be required before access is denied.

**Consequences:** Directly exposed file or search paths may need authenticated gateways or revised cache behavior. This is more work than updating trail/list handlers. Acceptance requires an actual route inventory and tests, not only a new helper.

**Search architecture decision, deferred from P0–P2:** The existing browser flow uses SvelteKit's `/api/v1/search/...` routes, whose server-side client queries Meilisearch with a tenant token. The token is also stored in a JavaScript-readable cookie; whether it permits a direct bypass depends on Meilisearch's deployment reachability. This remains a substantial independent work item, but a proxy routing layer already exists. Estimate query, facet/count, latency and operational effects before choosing an approach:

- Extend the existing server routes to validate candidates against current authority, prevent leaks in facets/counts/map results, and close any directly usable bypass credentials or network paths.
- Relying on index/token policy, whether accessed through the proxy or directly, requires propagating every relevant status transition into searchable state and addressing age expiry, old tokens, index failure and update lag. Asynchronous re-indexing alone cannot satisfy immediate withholding. A bounded inconsistency policy is a different, explicitly accepted contract and needs a demonstrated bound and failure behavior.

The strict target in ARC-018 assumes the first kind of authoritative enforcement. If maintainers choose the second approach, revise that target and its acceptance tests explicitly; do not imply both offer the same guarantee. P2 does not change the search proxy or make a new index-consistency promise. Frozen source locators for the current flow are in chapter 08.

Recommended default for server-controlled delivery is checking current eligibility before each new response, including cached bytes. A short-lived URL alone cannot deliver immediate revocation unless its consumer also checks the current grant/generation. Public CDN reuse that bypasses that check requires a separately accepted nonzero revocation bound and must not be described as immediate withholding. Previously downloaded third-party copies remain outside this guarantee.

**Open compatibility detail:** Raw relation IDs, total counts, map bounds and facets can reveal hidden membership. Decide which public shapes change and which adapters filter them. An owner may retain full membership while anonymous/public snapshots contain only visible members.

## D-08 — Delivery guarantees

**Recommendation:** Durably record per-recipient delivery work with the local mutation. Retry transient failures, survive restart and make repeated receipt harmless. Proposed default bounds: **8 attempts within 72 hours**, honoring recipient retry timing without exceeding the job deadline. Exact schedule and failure classification are in chapter 03.

Local command success means the change and intent are committed. It does not mean every remote inbox accepted the activity. Operators can inspect failed deliveries and retry deliberately without generating a new logical activity.

**Trade-off:** Durable jobs add schema, storage, cleanup and operational work. Introduce them after the access/replica contracts, except where a rollout explicitly relies on durable invalidation or cleanup.

**Execution model:** Start from one backend process and SQLite. If durable jobs are selected later, prioritize persisted intent, startup recovery, bounded in-process concurrency and stale-result protection. Multiple processes, distributed coordination and lease renewal are not prerequisites. RPL-028 permits this simpler execution model; cross-process claims become mandatory only if that deployment is deliberately supported.

## D-09 — Existing database records

**Recommendation:** Keep valid local originals unchanged. Map remote legacy records conservatively: old completion flags are evidence of an earlier code path, not proof of the new snapshot scope or recent public visibility. Store “verification unknown”, enqueue bounded checks and preserve local annotations/grants.

**Availability trade-off:** Enforcing the final eligibility gate immediately can hide many legacy records until checks finish. Stage backfill and report coverage before activation. Do not claim a record was verified “now” merely because a migration ran.

**Open rollout choice:** An explicitly bounded compatibility phase may temporarily retain the old read policy while verification progresses. It must have a deadline, observable progress and a documented weaker guarantee. Final strict mode cannot inherit fabricated timestamps from that phase.

## D-10 — API compatibility

**Recommendation:** Pure refactors preserve existing response shapes and codes. A later versioned or explicitly documented behavior release normalizes: denied/withheld public reads to generic 404, temporarily unavailable known replicas to 503 with retry guidance, and usable cached reads to 200 with suitable freshness information.

Unsupported detail capability is a separate permanent outcome: proposed `422` with `code=remote_detail_unsupported` and no `Retry-After`, plus an explicitly supported summary representation when eligible. Do not retry an ActivityPub-only peer indefinitely or pretend its summary satisfies a full trail detail contract. The exact response code remains an API compatibility decision.

Anonymous responses should not distinguish private existence from absence. Owners/administrators may receive richer diagnostics through an authorized channel. New UI strings should describe availability, not internal state fields.

**Open details:** Header versus JSON freshness metadata; compatibility for PocketBase-generated errors; empty versus omitted filtered expansions; membership-ID filtering. Resolve these before updating frontend expectations and API reference examples.

## D-11 — Retention and irreversible actions

**Recommendation:** Separate access suppression from physical deletion. Withholding/deletion denies access immediately. Suggested initial operations policy: purge downloaded bytes for confirmed deletion within 24 hours; retain withheld replica bytes for at most 30 days behind the gate to allow efficient revalidation, subject to storage/privacy policy. Both intervals are proposed and need review.

Retain minimal identity/authority/tombstone data needed to reject stale activities; do not retain full private payloads in tombstones or delivery logs. The default proposal never reuses a truly deleted object IRI. Republishing a merely withdrawn object can retain its identity after new authoritative public verification.

Backups, already completed downloads and third-party copies have separate retention limits. Restoring a database requires reconciling newer durable suppression history before revalidating restorable objects; revalidation alone cannot reconstruct a lost terminal tombstone. The storage/backup design must retain that history independently of an older restored snapshot, or keep affected serving disabled when it cannot establish the guarantee. The product documentation must describe these practical boundaries. Migration rollback cannot restore intentionally purged bytes or undo messages already delivered.

## D-12 — Initial resource budgets

These are **starting values for measurement**, not discovered current limits. Validate representative large trails/lists before adoption.

| Resource | Proposed initial bound | When exceeded |
| --- | --- | --- |
| Single metadata JSON response | 4 MiB decoded body | Reject attempt; keep previous eligible snapshot |
| Total structured snapshot across pages | 32 MiB / 100 pages | Stop with capacity error; never mark partial data complete |
| Waypoints or list memberships per snapshot | 20,000 entries | Report unsupported size; do not truncate and reconcile deletions |
| JSON nesting / redirect hops | 32 levels / 5 redirects | Reject malformed or excessive indirection |
| Active fetches per origin / process | 2 / 16 | Queue and fairly schedule |
| Active deliveries per destination / process | 2 / 16 | Queue; inbound request volume cannot create unlimited goroutines |
| Metadata request timeout / snapshot deadline | 15 seconds / 60 seconds | Cancel attempt and classify transient result |
| Single media file / media per object | 50 MiB / 500 MiB | Leave that media unavailable; metadata readiness remains separate |

Queue capacity, per-account discovery quotas and disk high-water behavior must be selected for the expected deployment size before enabling broad prefetch. Existing safety controls must remain effective; these suggestions do not authorize fetching private networks or relaxing certificate/redirect validation.

## D-13 — Link-token mutation

**Recommendation:** Keep the documented v1 promise to scoped local reading. Record any existing accepted write behavior before restricting it; a stored link-share permission value alone is not proof that every child mutation is supported or that the caller has an authenticated author identity.

**If editing is desired:** Specify permitted operations, contribution attribution, file upload scope, target immutability, expiry/revocation and abuse limits as a separate capability contract. Do not simply treat token possession as a local user session. Final enforcement and UI wording must match the accepted outcome.

## D-14 — Private profiles with public objects

**Recommendation:** Independently public trails/lists remain readable when profile discovery is private. Use only already-public object attribution or a minimal actor reference if richer profile data is unavailable. Do not expose a private biography, follower list, settings or private statistics merely to render an author label.

**Still open:** Exact public actor fields, how private profiles appear in search, whether existing follows are affected, and administrator visibility. These need a profile-specific matrix before changing actor endpoint behavior; object public flags must not be silently rewritten.

## D-15 — Remote statistics

**Recommendation:** Do not present opaque remote totals as if the local per-object access gate had validated them. For the strict profile, require item-level eligible contributions or a separately specified aggregate contract whose scope can be checked. Otherwise hide the unsupported aggregate and optionally link to the origin's public profile.

**Alternative:** Display origin-reported public totals with explicit provenance and a different guarantee. This cannot claim equivalence to locally recomputed, per-object filtered statistics. Select the product contract before implementing a broad statistics gateway.

## D-16 — List sharing as an explicit command

**Recommendation:** Introduce an explicit local `ShareListWithActor` operation that enumerates authoritative current membership, checks all intended grants and atomically commits the list share plus eligible local view grants. Remote sharing creates no private grants and only persists its permitted recommendation/intent. Repeated requests are idempotent for the same target/recipient.

This is a proposed replacement for the current browser sequence, not current transactional behavior. Until it exists, retain explicit partial-success handling and safe retries in the UI. Avoid implicitly triggering bulk auto-grants on every raw `list_share` save or federation import.

Changing the list permission does not upgrade trail permissions. Adding members later and revoking derived grants remain separate explicit product choices; the first scope preserves independent trail grants after unsharing. If future cascading revocation is wanted, record grant provenance rather than inferring it from present membership.

## Acceptance record template

For each accepted decision, append:

```text
Decision ID:
Outcome and chosen values:
Rationale / alternatives rejected:
Compatibility and migration consequences:
Acceptance scenarios:
Maintainer / date / tracking issue:
```

No acceptance records have been filled in by this preparation task.
