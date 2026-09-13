# Acceptance and regression specification

Status: **DRAFT / PROPOSED**. This is a catalog of 52 possible acceptance scenarios, not a requirement to implement 52 tests before the next refactor. The recommended next scope is P0–P2 only; its bounded acceptance selection is in section 2 below. P3–P8 and their tests require a separate maintainer decision. No new test suite has been implemented or executed as part of this documentation task. Existing tests are evidence for particular fixes, not proof of the complete proposed contract.

Requirements are defined in [access](01-behavior-and-access.md), [replica lifecycle](02-replica-lifecycle.md), [events](03-events-and-delivery.md), and [architecture](04-architecture-and-data.md). Behavior-changing expected outcomes depend on acceptance of the corresponding [decisions](07-decisions.md).

## 1. Test layers and responsibilities

| Layer | Environment | Purpose | What it cannot establish alone |
| --- | --- | --- | --- |
| L0: policy and transitions | Pure functions, fixed clock, generated event sequences | Exhaustive principal/state combinations, completeness and fencing invariants | Actual PocketBase rule/hook/HTTP behavior |
| L1: real application API | Disposable PocketBase DB with real migrations, actual router and registered hooks | Records API, custom route and file access; DB state after rejection | Cross-instance delivery and restart behavior |
| L2: coordinator and workers | Real DB, injected transport/clock/scheduler; restartable worker when introduced | In-process concurrent fetches, failures, restart recovery and transaction/effect ordering; leases only if implemented | Complete deployed proxy and third-party interoperability |
| L3: two-instance federation | Two isolated app stacks, real actor identities/signatures, controlled connectivity | End-to-end publish/share/read/offline/revoke/delete behavior | Compliance of arbitrary third-party peers |
| L4: interoperability fixtures | Recorded/reproducible ActivityPub and legacy-wanderer fixtures | Protocol variants, unsupported capabilities and negotiated extensions | A guarantee for every server version in the fediverse |

L0/L1 and the small existing response/background-sync seam suffice for P0–P2. A two-instance harness, a general scheduler, multi-process workers and real Meilisearch integration are not prerequisites for this bounded work. L3/L4 become necessary only for later claims that depend on actual peer interactions or protocol compatibility.

**Requirements for the harness used by the affected implementation:**

- Use the registered application migrations through the target head. Do not maintain a stale handpicked list of rule migrations.
- Any replaced external-index migration must be on an explicit reviewed list and contain only external index setup; no schema or rule migration may be skipped because it errors.
- Set origin and proxy credentials explicitly in isolated test configuration. Tests must not depend on a developer's environment or contact live community instances.
- Where a test claims hook wiring, invoke production hook registration in L1/L3, or add a focused registration assertion if that wiring is not independently callable. A test that manually binds one helper remains useful but does not prove that `main.go` actually registers it. A generic app-startup framework is not required for P1.
- Use barriers/channels and a controllable clock for concurrency and retry timing. Wall-clock sleeps are not an ordering assertion.
- Production HTTP protection remains active. Tests can inject a resolver/transport restricted to the test origins; they must not add a production “allow arbitrary private destinations” escape hatch.
- Worker dependencies are scoped per app/test. Avoid package-global mutable transport overrides in parallel tests unless the test is explicitly serialized and restores the value.
- Run real Meilisearch integration for later search/facet access claims. A fake projector suffices for retry/intent tests but not for the security behavior of a directly queryable index. P0–P2 make no new search/facet access claim.

## 2. Bounded acceptance selection for P0–P2

P0 records D-01/D-02/D-03 as accepted, amended, or deferred, and explicitly separates existing behavior from later policy proposals. A deferred policy decision does not prevent an extraction that preserves the existing policy. P1 builds only the fixtures needed by the rows below, preferably by reusing current regression tests. P2 introduces explicit import/output boundaries with no replica-state migration or new access policy.

| Slice | Small concrete acceptance selection | What is deliberately not asserted yet |
| --- | --- | --- |
| P1: existing collection rules | T-001 for anonymous root reads with an empty local-user relation and one private child; T-007 for immutable share targets through the actual Records API. Reuse the complete existing security suites when present | A new cross-surface replica eligibility gate, new token scope or revised child-author rights |
| P1: current checked expansion | T-003 for a public mixed list through the real list route; T-004 for preloaded/nested expansions with and without requested `expand`; T-005 for anonymous tag expansion | A new wire snapshot contract, concealment of raw membership IDs, or a two-instance deployment |
| P1: sharing compatibility | Existing-policy portion of T-009: create/update of private remote shares is rejected; a public remote `permission=edit` input retains the baseline result while effective remote edit rights are not inferred | The proposed rejection of public remote `edit` input, new recipient-change activities or atomic list-sharing commands |
| P2: import allowlist | Two focused cases below for each changed trail/list import entry point; explicit decoder/persistence assertions with the current supported payload shape | New fixed fetch scope, relation completeness/reconciliation, author/visibility policy, or a new remote protocol |
| P2: response isolation and compatibility | T-030 against actual trail/list handlers with deterministic barriers; existing-policy portion of T-052 compares successful root/expand shapes and documented current errors | New stale/withheld/unsupported statuses, media access gates or UI readiness states |

The two P2 import cases are:

1. **Protected-field input:** start with a record containing known local identity and sync state. Import representative allowed remote content together with forged local IDs, collection fields, sync flags and broader `expand` data. Assert that the expected content changes, local-owned values remain intact, and the actual route output still rebuilds authorized expansions. Existing relation decoding must continue to work; removing the whole `expand` input before decoding legitimate relations is not an acceptable shortcut.
2. **New schema field is not automatically importable:** in a disposable real-schema DB, add an ordinary test-only local field unknown to the decoder, store a sentinel, and include an attacker-chosen value for it in the remote payload. Assert that import leaves the sentinel unchanged while an allowed content field updates. This detects a fallback to wholesale `Record.Load` even when today's known protected fields are manually stripped. The additional field is a test fixture, not a production migration. Public output retains its approved existing contract; any new field's exposure needs an explicit presenter decision.

These are focused extensions of the current regression suite, not a requirement to replace it. Parameterize trail/list cases where the boundary is shared. Record any discovered behavior change as a separate fix or explicit policy proposal; do not change the expected result merely to call P2 behavior-preserving.

**P0–P2 do not require** the later age/withholding/deletion model, durable jobs/delivery, lease expiry, fixed full-detail snapshots, search/file/aggregate enforcement, a withdrawal extension, migration/backfill/restore tooling, or L3 stories A–D. Existing bugs in those areas remain visible in the evidence and deferred scenario catalog; completion of P2 makes no claim that they are solved.

## 3. Canonical fixtures

Instance A is the authority for Alice; instance B is the authority for Bob. Charlie and Dana are separate local users on A. Mallory is a valid remote actor from B who owns none of Alice's objects. An anonymous client has no local user or actor grant.

Create these objects with deterministic logical names and generated valid storage IDs:

| Fixture | Properties and purpose |
| --- | --- |
| `T-public` | Alice's public trail, tags/category, geometry, GPX manifest, optional photos and waypoints |
| `T-private` | Alice's private trail with no shares |
| `T-foreign` | Dana's private trail; Charlie/Alice do not own it |
| `L-mixed` | Alice's public list containing `T-public`, `T-private` and `T-foreign` in a known order |
| `S-view`, `S-edit` | Local grants to Charlie for Alice's appropriate objects |
| `S-remote` | Targeted recommendation to Bob of a public object |
| `K-local` | Local trail token for `T-private`, plus wrong/revoked/rotated token variants |
| `C-own`, `SL-own` | Charlie's contribution to Alice's trail for child-author maintenance decisions |
| `R-stub` | Remote identity known, no public verification or complete data |
| `R-detail` | Verified public remote copy with a fixed complete snapshot; media readiness can vary |
| `R-withheld`, `R-deleted` | Previously public copies suppressed by authoritative evidence |
| Private-profile author | Public object whose author profile is private, to separate profile and object policy |

The complete matrix also includes a remote actor with an empty local-user relation, an empty share back-relation, a local administrative principal, and an unauthenticated request containing an arbitrary actor identifier. They must not collapse to the same authority. Build fixtures incrementally as an accepted package needs them; P1 does not need every replica lifecycle state in this table.

## 4. Access and sharing scenario catalog

Each row is a named acceptance scenario for the corresponding accepted behavior. Rows that combine existing and proposed behavior are not indivisible early merge gates: use the explicitly scoped selection in section 2 for P0–P2. Parameterize equivalent trail/list or direct/nested variants instead of duplicating entire test setups.

| ID | Setup and action | Required observable outcome | Layers / trace |
| --- | --- | --- | --- |
| T-001 | Anonymous GET/LIST of private local or remote-authored records; ownership/share user relations empty | No private root, child, share, or aggregate data becomes visible through empty-string equality | L1; ACC-003/004, existing read-rule tests |
| T-002 | Owner, local viewer, local editor, unrelated local user and signed remote actor read the same private object | Only explicit local permissions apply; signature and remote identity confer no private access | L0/L1; ACC-005/009/018 |
| T-003 | Public mixed list expanded anonymously, as owner, as local grantee, and via federation fetch | Each trail evaluated separately; no hidden trail metadata or unintended grant | L1/L3; ACC-012/015, SHR contract |
| T-004 | Preload a record with broader nested expansions, both with and without request `expand` | Response contains only authorized requested relations; stored/index payload cannot leak through | L1; ACC-011/013/014 |
| T-005 | Public trail has tags; anonymous viewer and public remote fetch request it | Allowed tag views survive; authenticated catalog-listing policy remains distinct | L1/L3; ACC-016 |
| T-006 | Valid/wrong/revoked local token reads root, waypoints, summit logs, comments and media | Scope matches the accepted token matrix exactly; unsupported child scope is denied | L1; ACC-006/016/017/029 |
| T-007 | PATCH an owned trail/list/link share to a foreign private target, optionally changing recipient/token too | Rejected; stored target and all other fields unchanged; formerly unauthorized GET remains denied | L1; immutable-target SHR requirements |
| T-008 | PATCH permissions, a legitimate recipient, an explicitly unchanged target, or link token rotation | Accepted within existing policy; prior token no longer works after rotation; no ownership upgrade | L1; ACC-018 and SHR requirements |
| T-009 | Create/update private remote share through direct Records API and web flow; separately create/PATCH a public remote share with `permission=edit` | Private remote request rejected without persistence or Announce; public remote edit input follows accepted D-06/D-10 behavior explicitly, without suggesting effective edit rights | L1/L3; SHR create/update guard, D-06/10 |
| T-010 | Share `L-mixed` with Charlie, then Bob; provide an incomplete browser member expansion, fail the second auto-grant, and retry | Server command enumerates authoritative members and atomically commits eligible local grants or none; retry is idempotent. Until D-16 ships, browser sequence reports partial success accurately and repairs it on retry. Bob receives no private grants; recipient list refreshes | L1/L3; SHR-011/019/020, D-16 |
| T-011 | Delete list share; independently existing trail shares remain | Exactly documented list-edge and grant behavior; no silent bulk revocation | L1; SHR unshare semantics |
| T-012 | Private-profile author has a public trail/list | Object visibility is independent under accepted profile policy; no accidental profile/private-field disclosure | L1; ACC-040–042, pending profile decisions |
| T-013 | Parent becomes unreadable after Charlie authored a comment/log; separately seed comments/logs/waypoints/files with missing or unverified remote parent | Ordinary Records API, expansions, bytes and aggregates cannot expose children through unknown parent authority; accepted author-maintenance exception separately exposes only permitted own content | L1/L3; ACC-021–026/045, D-04 |
| T-014 | Read existing file URLs, thumbnails and download variants after local revoke or remote withholding | New byte serving is denied through every route; URL knowledge and warm caches do not bypass access | L1/L3; ACC-028–031, D-07 |
| T-015 | Search index deliberately retains an ineligible object | No hit/snippet/facet/count/map-bound leak; direct index credentials cannot bypass the chosen gate | L1/L3 with real index; ACC-033/034, ARC-018 |
| T-016 | Statistics contain eligible and ineligible contributions with duplicate completion/log entries | Eligibility before aggregation; existing deduplication preserved; no hidden contribution in totals | L1; ACC-035–039 |
| T-017 | Reuse an HTTP/app cache across logout, different users and grant revocation | No cross-principal response reuse; token-based access does not populate a shared public cache | L1/L3; ACC-039, ARC-012 |

## 5. Replica, completeness and concurrency scenario catalog

| ID | Setup and action | Required observable outcome | Layers / trace |
| --- | --- | --- | --- |
| T-018 | Same new remote trail requested via local ID/handle with several different expand strings | One fixed snapshot scope and one eligibility outcome; query-dependent completeness is impossible | L1/L2; RPL-026/027, ARC-012/013 |
| T-019 | Incoming Activity or list summary references previously unknown trail | Stable identity and queued verification; no invented public clock or arbitrary public placeholder | L2/L3; RPL-041, ACC-019 |
| T-020 | Snapshot has complete empty tags/waypoints, then another omits them; a removed waypoint later appears under another parent | Complete emptiness clears the relevant scope; omission does not erase data or count as complete; scope removal creates no terminal identity tombstone | L0/L2; ARC-008, RPL-015–017 |
| T-021 | Multi-page relation fails halfway or changes source revision between pages | No destructive partial reconciliation; previous coherent detail snapshot retained | L2; RPL-035/037 |
| T-022 | List snapshot completes while a member's origin is offline | List scope can complete; member independently unavailable; no recursive hydration or borrowed public proof | L2/L3; ARC-009, ListSnapshotV1 |
| T-023 | Metadata and manifest complete, optional photo download fails; repeat with mandatory route data missing | Detail readiness established; file readiness remains accurate; optional photo failure does not block the accepted offline-route scope, while missing mandatory route data does | L2/L3; RPL-038/039, D-02 |
| T-024 | GPX/photo source version changes while an older file is stored | Old filename is not proof of new content; versioned replacement and explicit readiness | L2; RPL-039 |
| T-025 | Timeout, 429 and 5xx at age 2h, just below 7d, and exactly 7d | Eligible stale response only before maximum age; no object content at/after 7d without fresh proof | L0/L2/L3; D-03, RPL read gate |
| T-026 | Trusted origin returns 401/403/404 after a usable public snapshot | Withheld immediately; details, children, media and projections stop serving content | L2/L3; RPL HTTP mapping, ACC-021 |
| T-027 | Origin returns 410 or verified owner Delete arrives; later deliver an older valid withdrawal | Tombstone and generation fence; stale Create/Update/Announce or 200 cannot resurrect that identity; delayed withdrawal cannot change deleted to withheld | L2/L3; RPL-042/043, EVT-032/036 |
| T-028 | Actor fetch, repeated Activity, local index write, or list refresh occurs at age limit | None renews the object's `last_public_verified_at` | L0/L2; RPL evidence rules |
| T-029 | 304 response with matching versus absent/wrong representation validator | Only matching eligible public representation may renew verification; no fabricated relation completeness | L2; RPL HTTP/validator rules |
| T-030 | Background fetch pauses; response authorized/serialized; fetch then commits changed values | Response is a coherent captured version; stored refresh succeeds independently | L1/L2; ARC-010/017, existing background regression |
| T-031 | Refresh/media/index job pauses; withholding or deletion commits; paused job resumes | Superseded generation cannot publish, restore public visibility or reattach media | L2/L3; RPL-031/036/040 |
| T-032 | For a later P3 coordinator, two goroutines request the same scope and the process restarts while work is pending. Only if multiple worker owners or expiring claims are introduced, additionally simulate an expired claim and an old worker resuming | Required for P3: coalesced in-process work, abandoned work becomes eligible after restart, and obsolete results cannot overwrite newer state. Conditional variant: a reclaimed claim rejects the previous owner's commit; no multi-process framework is required | L2; RPL-028/030; deferred beyond P2 |
| T-033 | An older fetch finishes after a newer accepted snapshot or newer invalidation | Older result discarded; origin clock/opaque ETag not incorrectly compared as an integer revision | L2; RPL-031/032 |
| T-034 | SQL save fails, projector is down, or network stalls during fetch | Atomic DB rollback as appropriate; no network held inside write transaction; failed index does not bypass gate | L2/L3; RPL-033–036, ARC-014/015 |
| T-035 | Valid activity/signature from an unrelated actor or unsafe redirect response claims deletion | No authority transition for victim object; rejected activity does not invalidate its cache clock | L1/L2; identity/authority requirements |
| T-036 | Resource cap exceeded in JSON, relation pages, media or queue | Bounded error and retained previous eligible state; no truncation promoted to full detail | L2; ARC-025, D-12 |

## 6. Event and delivery scenario catalog

| ID | Setup and action | Required observable outcome | Layers / trace |
| --- | --- | --- | --- |
| T-037 | Local public create/update commits while destination is offline | Local command succeeds with durable intent; retry survives worker/process restart | L2/L3; chapter 03 durable delivery, ARC-015 |
| T-038 | Crash after send before delivery acknowledgement is recorded | Same logical activity can be resent; receiver applies effects once | L2/L3; chapter 03 idempotency |
| T-039 | Same Activity delivered repeatedly to same recipient and to two different local recipients | One effect per intended recipient; deduplication does not suppress the second recipient | L1/L3; chapter 03 receipt keys |
| T-040 | Two different Update activities refer to same object | Content state updates without duplicate `(recipient, object)` feed entry; deliberate notifications follow event policy | L1/L3; feed/effect uniqueness |
| T-041 | Local object becomes private before queued old Create/Update is sent | Obsolete public payload is cancelled or replaced by accepted lifecycle intent; worker rechecks generation before send | L2/L3; chapter 03 and ARC-016 |
| T-042 | Share remote recipient changes or share is removed under accepted D-06 | Correct old-recipient Undo/new-recipient Announce once; source public object remains available | L1/L3; D-06 and chapter 03 |
| T-043 | Sender retry receives 429/Retry-After, transient 5xx, permanent rejection, then deadline exhaustion | No retry earlier than allowed; bounded attempts; durable inspectable terminal state | L2/L3; D-08 |
| T-044 | Negotiated withdrawal reaches capable peer; same action targets legacy peer | Capable peer suppresses via accepted extension; legacy limitation explicit and verification-age gate respected | L3/L4; D-05 |
| T-045 | Withdrawal, republication and delayed earlier messages arrive in different orders | Withheld object restored only by current authoritative public verification; truly deleted IRI stays terminal | L0/L2/L3; chapters 02/03 |
| T-046 | Forged/replayed Follow, Accept, Undo, Like or child activity; for Follow/Like/Announce receive Undo before original, restart, then deliver the original and a new legitimate action to two local recipients | Author/recipient checked; durable cancellation suppresses only the delayed original's effects, including counts; new authorized action remains possible; one recipient's processing never consumes another's work | L1/L2/L3; EVT-049/051/053/054 |
| T-047 | Inbox application fails after a durable receipt but before all effects commit | Receiver retries internally or returns retriable failure consistently; no acknowledged lost work | L2/L3; chapter 03 receipt lifecycle |

## 7. Upgrade, restore and compatibility scenario catalog

| ID | Setup and action | Required observable outcome | Layers / trace |
| --- | --- | --- | --- |
| T-048 | Upgrade rows containing every combination of old sync flags and missing verification evidence | No migration-time public proof invented; local originals preserved; remote copies conservatively classified | L1/L2; RPL-044–047, D-09 |
| T-049 | Interrupt and resume backfill; enable strict gate before all origins reachable | Deterministic progress; uninitialized rows cannot bypass gate; operator sees availability impact | L2/L3; D-09 |
| T-050 | Restore backup predating an accepted Delete; origin now returns 200; repeat with newer suppression ledger unavailable and with old-binary rollback | Reconciled tombstone remains terminal despite 200; revalidation is not a replacement for missing deletion history; affected serving stays disabled until safe reconciliation or equivalent suppression exists | L2/L3; RPL-048, EVT-036, D-11 |
| T-051 | Legacy peer omits completeness metadata or supports ActivityPub summaries only | No false full-detail promotion and no endless unsupported fetch retries; summary capability explicit | L2/L4; ARC-013 |
| T-052 | Legacy browser/API client consumes compatible successes and newly proposed unavailable errors | No accidental response-shape break in pure refactors; accepted API changes and UI states explicitly tested | L1/L3; D-10, ARC-021–023 |

## 8. End-to-end stories for later accepted behavior

These stories compose the scenario IDs above. They are a proposed later validation program, not P0–P2 prerequisites or a request to build L3 infrastructure now. When maintainers commit to the corresponding cross-instance behavior, the selected story must run through real application ingress, including the trusted frontend/backend inbox hop where used. A partial story proves only its explicitly implemented steps.

### Story A: ordinary public federation

1. Alice publishes a trail on A; Bob follows Alice from B.
2. B receives the activity once and as a retry; verify one logical feed entry.
3. Bob opens the trail; B obtains the fixed public snapshot and displays permitted data.
4. Bob downloads the route, leaving optional photos uncached; the UI distinguishes those readiness levels.
5. Alice updates tags/waypoints, including an explicit empty set; B applies a coherent new snapshot.

Pass: correct identity, permissions, no duplicate effects, accurate completeness and media status.

### Story B: outage and changed visibility

1. Establish a verified cached public trail at B.
2. Make A unreachable; at 2h serve a clearly stale permitted copy.
3. Move the clock to 7d; B stops normal content serving without successful revalidation.
4. Restore A with an authoritative denial; B remains withheld across detail, file, search and aggregate paths.
5. Re-publicize the same non-deleted object; a fresh verified public snapshot can restore it.

Pass: no fabricated clock renewal and no inferred authority from restored actor connectivity alone.

### Story C: concurrent deletion

1. Start a slow refresh and slow media download on B.
2. A deletes the object; B commits the verified deletion.
3. Release the old fetch/download; also let a delayed index job and Create retry finish.
4. Restart B and repeat a direct old file URL request.

Pass: no stale job resurrects public content; minimal tombstone survives restart; bytes already sent are outside the revocation guarantee.

### Story D: mixed list and share updates

1. Alice shares `L-mixed` with Charlie locally and Bob remotely.
2. Verify exact local auto-grants, public-only remote membership and no grants for Dana's private trail.
3. Use direct Records API updates to try changing each existing share's target.
4. Apply legitimate permission/recipient changes and token rotation.
5. Remove list sharing and verify independent trail grants remain according to policy.

Pass: browser behavior and server authority match; remote notifications do not become private object rights.

## 9. Completion evidence and regression handling

For each accepted implementation package, record commit, migration head, test layer, scenario IDs or scoped variants, command/result and any skipped dependency. An out-of-scope L3/L4 scenario is not a missing P2 requirement. If a PR makes a claim that requires L3/L4 and skips it, record that as a limitation rather than a passing result.

The existing migration harness, checked-expansion tests, share-guard API tests and `ddd1929d8` background-sync test should be reused. The latter already provides a deterministic response-versus-background-mutation regression; its existence corrects any earlier impression that this particular path lacks testing.

A merge gate for P0–P2 should include:

- The section 2 selection relevant to changed entry points and retained existing regressions; no requirement to newly implement the entire catalog.
- Full Go suite, vet and build for backend changes; focused frontend tests and type checks for changed UI contracts.
- An explicit check that the refactor has not silently adopted a deferred behavior or protocol change.

For separately accepted later work, add only the layers required by its claims: the affected L3 story for cross-instance behavior, L4 fixtures before publishing a withdrawal/capability extension or changing legacy response interpretation, and a requirement-to-test trace review for changed requirements. All stories A–D would be necessary before claiming the *complete* proposed federation contract is implemented; that comprehensive claim is not the next project's objective.

Performance acceptance for a later package introducing resource limits is measured against its accepted D-12 budgets using representative large data. Passing a tiny fixture is not evidence that pagination, per-origin fairness, index filtering or media cancellation scales correctly. Establishing that later performance program is not a prerequisite for the initial boundary extraction.
