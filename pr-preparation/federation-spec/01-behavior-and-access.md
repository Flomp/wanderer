# Draft: behavior, principals, and access

**State: DRAFT — proposal for review, not an approved product policy.**
**Reviewed code:** integration commit `865251d49dc6c1d734cccb85ee51c680c62d35a6`, retained by tag `spec/federation-baseline-2026-09-13`; see [baseline retrieval](08-evidence.md#1-frozen-reference).
This document describes a proposed v1 contract and distinguishes it from observed implementation behavior.
It does not assert that all requirements are implemented, that every endpoint was audited, or that the user has approved unresolved policy choices.

## 1. Reading the requirements

- **EXISTING** means the stated behavior is supported by the reviewed code or published product documentation; the wording defines the scope of that evidence.
- **PROPOSED** means a normative requirement for the proposed design, potentially requiring implementation changes.
- **OPEN** means a decision is required; alternatives are not interchangeable implementations of an approved requirement.
- **MUST**, **MUST NOT**, and **SHOULD** apply to the proposed contract. They do not turn this draft into an approved policy.
- Requirement IDs are stable references. A future accepted revision should retain IDs when recording decisions or implementation evidence.

The current implementation uses PocketBase collection rules, request hooks, custom routes, federation conversion, and Meilisearch filters.
These mechanisms do not presently establish one universal authorization boundary for every output type.
The proposed adapter boundary complements PocketBase rules; it does not replace them with unchecked internal record reads.

## 2. Principals and v1 scope

| Principal | Evidence or credential | Authority in v1 |
| --- | --- | --- |
| Anonymous visitor | No validated local session | Public content only, subject to object and remote-status gates |
| Local authenticated user | Valid local user session or local API credential | Public content plus locally authorized ownership and grants |
| Remote actor | ActivityPub identity, optionally verified by HTTP signature | Origin identity; no local user identity or private-content entitlement |
| Share-token holder | Valid token bound to a particular local trail | Scoped local trail access; not an authenticated user or remote actor |
| Internal process | Explicit application operation such as import or indexing | Only the authority needed by that operation; not a public-output principal |
| Superuser/operator | Explicit administrative credentials | Administrative scope, separate from ordinary visitor and federation behavior |

**ACC-001 — EXISTING.** Cross-instance trail sharing is documented as public-only and view-only: it resembles a notification or mention, not collaborative editing. Private collaboration belongs to users on the same instance. [S1]

**ACC-002 — PROPOSED.** V1 MUST retain public-only, view-only cross-instance federation for trails and lists. This draft MUST NOT be interpreted as adding private federation, remotely delegated editing, or cross-instance bearer-token authorization.

**ACC-003 — PROPOSED.** A remote actor MUST NOT acquire local ownership because its `user` relation is empty. A local ownership or recipient comparison MUST first establish a nonempty, validated local user identity.

**ACC-004 — EXISTING.** The reviewed read-rule migration adds nonempty-auth guards to ownership and recipient clauses while preserving their public branches and relevant token branches. It also preserves rules locked with `nil`. This is a read-rule correction, not a claim that every write rule was reviewed. (S4)

**ACC-005 — PROPOSED.** A valid remote HTTP signature MUST establish who sent a request or activity, not whether that actor may read private local content. A signed foreign actor requesting public object data MUST pass the same public-visibility checks as an anonymous visitor.

**ACC-006 — PROPOSED.** A share token MUST remain a local capability bound to its issuing instance and trail. It MUST NOT be placed in federation activities, treated as remote actor authentication, or reused as authorization for another object or origin.

**ACC-007 — PROPOSED.** When a local session and a valid local share token are both present, the applicable local read grants MAY be combined. Neither credential grants extra remote-replica rights or bypasses the remote-status gate.

**ACC-008 — OPEN.** The administrative interface may need access to withheld data for repair or incident investigation. Its scope, audit records, and export controls require a separate administrative contract; ordinary federation and visitor endpoints MUST NOT inherit that scope.

## 3. Access decision model

The proposed ordinary read decision is the intersection of:

1. The principal's object-level authorization under the local collection policy.
2. Any required relation or parent authorization for the requested representation.
3. The `RemoteStatusGate` for every remote object whose content would be disclosed.
4. The field and output restrictions of the requested surface: record, expansion, file, search result, or aggregate.

**ACC-009 — PROPOSED.** An endpoint MUST evaluate the complete decision before publishing data. A successful internal lookup, a known record ID, a list membership, or a successful previous request MUST NOT serve as an authorization decision.

**ACC-010 — PROPOSED.** The common PocketBase adapter MUST carry the validated request principal and applicable local share token into collection-rule evaluation. It MUST preserve PocketBase's list/view distinctions and MUST NOT substitute `FindRecordById`, unrestricted expansion, or an import's authorization for a caller's read rule.

**ACC-011 — EXISTING.** The reviewed custom response helper clears preloaded expansions and invokes PocketBase enrichment with the current request. This avoids retaining unchecked relations populated by a remote payload or an indexing hook. (S5)

**ACC-012 — PROPOSED.** Expansions MUST be built from locally resolved, independently authorized relations. A nested expansion MUST NOT inherit permission solely from its root record. Unrequested preloaded expansions MUST NOT appear in a response.

**ACC-013 — PROPOSED.** Internal indexing and importing MAY use broader data than the requesting visitor can read, but MUST use distinct working objects or explicit projections. Their expansions and visibility flags MUST NOT be reused as a visitor or federation response.

**ACC-014 — PROPOSED.** Failed authorization MUST omit the restricted relation or deny the requested resource according to a documented endpoint convention. It MUST NOT fall back to unchecked cached data, a raw remote payload, or a broader internal query.

## 4. Proposed ordinary-output matrix

The following is a **PROPOSED target matrix**, not a statement of universal current implementation.
“Grant” means the particular local user has an applicable object grant; “token” means the matching local trail token.
All cells involving a remote object are additionally restricted by ACC-019 through ACC-021.

| Resource or surface | Anonymous | Local owner | Local grantee | Local token holder | Signed foreign actor |
| --- | --- | --- | --- | --- | --- |
| Public local trail, detail | Read | Read/manage within existing policy | Read; edit only if granted | Public read | Same public read as anonymous |
| Private local trail, detail | Deny | Read/manage | Read or edit within grant | Scoped read; see ACC-017 on edit | Deny |
| Public local list, detail | Read list; filter each trail | Read/manage; filter each trail | Read; edit list only if granted | No additional list grant | Same public read as anonymous |
| Private local list, detail | Deny | Read/manage | Read or edit list within grant | No additional list grant | Deny |
| Trail nested in any list | Apply the trail's own rule | Apply the trail's own rule | Apply the trail's own rule | Only if matching local trail token is deliberately applicable | Public trail only |
| Waypoint, comment, summit log | Public-parent read where its child policy permits | Parent and child permissions; author exceptions below | Parent grant plus child-specific policy | Explicit token-enabled child types only | Public-parent read only |
| Tags, category labels, public actor fields | Public view fields only | Same public fields | Same public fields | Same public fields | Same public fields |
| Share records and link-token records | Deny | Authorized management fields | Own actor-share visibility where allowed; no ownership upgrade | Token possession does not expose management records | Deny |
| GPX, photos, thumbnails, list media | Same visibility as their authorized representation | Authorized representation only | Authorized representation only | Explicit local file capability only | Publicly eligible representation only |
| Search hits, suggestions, facets, counts | Publicly eligible content | Public plus locally authorized content | Public plus applicable grants | No global token-expanded discovery | Same public discovery as anonymous |
| Profile statistics | Activities from publicly eligible trails | Activities authorized for the viewer | Activities authorized for the viewer | No global token-expanded statistics | Same public statistics eligibility as anonymous |

**ACC-015 — PROPOSED.** Publishing a list MUST NOT make its private trails public. Reading, following, announcing, or sharing the list MUST NOT bypass a contained trail's policy. A hidden trail's name, location, description, media, and child data MUST NOT appear through the list.

**ACC-016 — EXISTING.** In the reviewed changes, anonymous tag *view* access is enabled for expansion; listing the entire tag catalog remains authenticated. Summit-log token access is added explicitly, while comment token access is not added by these changes. (S6)

**ACC-017 — OPEN (D-13).** A link-share record can carry a permission value, but the v1 contract should promise token-based reading only until token-based write semantics are separately specified and verified. A UI or API MUST NOT present possession of a token alone as authenticated author or editor status.

**ACC-018 — PROPOSED.** An actor-share `view` grant MUST NOT grant mutation. An `edit` grant MUST authorize only the existing explicitly supported object mutations; it MUST NOT imply ownership, grant administration, authorship reassignment, or unrestricted child mutation.

## 5. Remote-status gate and unavailable parents

This document uses the shared replica terminology without defining synchronization transitions:
`visibility_state = unverified | public | withheld | deleted`; `materialization = stub | metadata_ready | detail_ready`.
Media readiness is a separate concern. [02-replica-lifecycle.md](02-replica-lifecycle.md) defines authority evidence, transitions, and the exact `FreshnessGate`.
A `public` state alone is insufficient: expired public verification can block output without changing the state to `withheld`.
An actor lookup, activity, local write, or list summary does not by itself establish a member trail's independent public eligibility.

**ACC-019 — PROPOSED.** Every server-controlled remote-content output MUST apply `RemoteStatusGate`. Only a replica with `visibility_state=public` that also passes the specified `FreshnessGate` may disclose ordinary content; `unverified`, `withheld`, and `deleted` MUST NOT become visible because of local ownership-like IDs, old expansions, a search index, or a local share token.

**ACC-020 — PROPOSED.** Materialization MUST NOT be treated as authorization. A fully downloaded or previously indexed object can still be withheld; an authorized stub can still lack enough data for a detail response. Output behavior MUST respect both dimensions.

**ACC-021 — PROPOSED.** When a child depends on a remote parent whose required visibility is unknown or no longer public, all ordinary server outputs MUST withhold that dependent representation. This applies to direct records, nested expansions, file delivery, search, feeds, statistics, and federation output; it is not merely a UI rule.

An allowed “unavailable” response may contain a neutral status or an authorized local reference needed to manage an existing relation.
It MUST NOT reconstruct the withheld object's title, geometry, authorship details, media, or aggregate contribution from another store.
No automatic grant restoration, replica deletion schedule, or refresh interval is specified here.

## 6. Child authorship and maintenance after parent access changes

**ACC-022 — EXISTING.** Current child rules are not a universal `canRead(parent) AND canRead(child)` formula. For example, summit-log and comment rules include authenticated author paths independent of the public-parent branch; write policies are separate again. (S4)

**ACC-023 — OPEN (D-04).** For a child attached to a known *local* parent that becomes private or revokes the author's access, choose explicitly between preserving author access to their own contribution and requiring parent readability for its normal display. Until decided, tightening all child reads MUST NOT be described as behavior-preserving.

| Situation | Current evidence | Decision required |
| --- | --- | --- |
| Author's summit log on a local trail they can no longer read | An author read branch exists | Keep author-only detail, or provide a reduced maintenance representation? |
| Author's comment on an unreadable local trail | An author read branch exists | Can its own text remain readable/editable; what parent reference is shown? |
| Waypoint created by an editor after list/trail grant revocation | Child rules use their own ownership and trail paths | Preserve, restrict, or transfer maintenance authority; verify current actor/user semantics first |
| Locally authored child of an unverified/withheld/deleted remote parent | No complete uniform policy established by this review | Ordinary output remains gated; define any author-maintenance workflow separately |

**ACC-024 — PROPOSED.** Revoking parent access MUST NOT silently delete or transfer ownership of another user's child record. Retention, moderation, removal, and author maintenance MUST be explicit operations governed by their own permissions.

**ACC-025 — OPEN (D-04).** A minimal author-maintenance endpoint may allow deleting an authored contribution without returning hidden parent data. Whether it may also return or edit the author's own text/media, especially for a gated remote parent, requires an explicit decision; it is not an implicit exception to `RemoteStatusGate`.

**ACC-026 — PROPOSED.** If an author-maintenance exception is adopted, it MUST NOT grant parent access, restore ordinary discovery, expose other authors' contributions, or re-enable federation publication. Tests MUST distinguish maintenance responses from ordinary child representations.

## 7. Files and derived media

**ACC-027 — EXISTING.** The reviewed web file proxy constructs a PocketBase file URL, accepts thumbnail parameters, and forwards the response. It does not itself implement the parent/principal decision described above. Historical file schema includes unprotected fields. This evidence does not establish uniform current byte-level protection or a completed file-security audit. (S7)

**ACC-028 — PROPOSED.** Files and derived thumbnails MUST follow the eligibility of the representation they expose. Knowing a file URL, a former public URL, a filename, or an object ID MUST NOT bypass a private-parent or remote-status restriction.

**ACC-029 — PROPOSED.** A local trail share token MAY authorize its explicitly covered local files through an intentional capability exchange or validated request. It MUST NOT become a reusable cross-origin file credential. Thumbnails, transformed media, and download variants MUST preserve the same scope.

**ACC-030 — OPEN (D-07).** Choose the implementation and compatibility treatment for previously issued file URLs: protected PocketBase fields, an authenticated proxy, short-lived capabilities, or a documented combination. Cache headers, CDN invalidation, and revocation latency need an explicit bound; the current file proxy alone is not the target guarantee.

**ACC-031 — PROPOSED.** The server MUST stop newly serving ineligible bytes once its policy requires denial. It MUST NOT promise erasure of copies already downloaded by visitors or other instances; that limitation does not permit continued local disclosure.

## 8. Search, feeds, and statistics

**ACC-032 — EXISTING.** Search-token filters currently distinguish anonymous `public=true` access from authenticated public/author/share access. Search indexes are built through internal record expansion and are a separate projection from PocketBase record responses. (S8)

**ACC-033 — PROPOSED.** Search hits, suggestions, facets, bounding boxes, counts, and feeds MUST enforce the same principal and remote-status eligibility as record detail. Filtering only the final click-through response is insufficient because the search response itself discloses data.

**ACC-034 — PROPOSED (D-07).** If clients can query an index directly, its credentials, indexed eligibility fields, and invalidation behavior MUST enforce the applicable output gate. Otherwise the query MUST pass through an enforcing service. A stale index MUST NOT be treated as evidence that a replica remains public.

**ACC-035 — EXISTING.** Local-profile statistics query summit logs and completed-trail fallbacks through the current viewer's PocketBase client. With the reviewed guards, ordinary anonymous local-profile requests do not gain private-trail activities through empty ownership/share comparisons. The endpoint does not forward an arbitrary `share` parameter into these local queries. (S9)

**ACC-036 — EXISTING.** Remote-profile statistics are fetched from the remote origin, and anonymous remote-profile resolution is currently rejected. The local origin's guards do not establish the remote origin's policy. (S9)

**ACC-037 — PROPOSED.** Statistics MUST apply eligibility before merging activities or calculating totals. Excluded content MUST NOT contribute distance, elevation, duration, dates, activity counts, categories, or other aggregates. Existing owner-log-versus-completed-trail deduplication semantics SHOULD remain intact for eligible activities. [S3]

**ACC-038 — OPEN (D-15).** Remote statistics returned only as aggregates may lack the per-object information required by the local gate. Specify an eligibility-verifiable response contract or explicitly disable that representation; silently treating arbitrary remote totals as locally authorized is not the proposed contract.

**ACC-039 — PROPOSED.** Access-controlled responses and aggregates MUST avoid shared-cache reuse across different principals. Session changes, grant revocation, and remote-status transitions MUST not leave an authorized view attached to an unrelated later request.

## 9. Profile privacy and content visibility

**ACC-040 — EXISTING.** Profile privacy and object visibility are distinct in the reviewed implementation. Private-profile handling can deny actor/profile resolution while a trail or list's own view rule remains relevant; an account-private flag is not uniformly equivalent to making every published object private. (S10)

**ACC-041 — PROPOSED.** Making a profile private MUST NOT silently rewrite the public flags or ownership of its existing trails and lists. The UI MUST explain which profile/discovery surfaces change and which explicitly public objects remain independently accessible.

**ACC-042 — OPEN (D-14).** Specify whether private profiles expose public-object author summaries, public statistics, follower collections, and profile search results. These are separate presentation/discovery decisions; do not infer them from the trail's `public` flag or automatically withdraw all object access.

**ACC-043 — PROPOSED.** A failed profile lookup MUST NOT by itself grant access to cached objects. Any permitted cached-object response MUST still apply its own collection policy and, for remote content, `RemoteStatusGate`.

## 10. Sharing lifecycle and list behavior

**SHR-001 — EXISTING.** The reviewed server hooks reject creating or updating an actor share when its object is private and its recipient is remote. A public object can be shared remotely; private objects can be shared with local users. (S11)

**SHR-002 — PROPOSED.** Share creation MUST verify that the acting local user may administer shares for the specified object, then validate the requested recipient and permission. Read access or an object's ordinary edit grant MUST NOT automatically imply share administration.

**SHR-003 — OPEN (D-06, D-10).** A remote share MUST express view/notification semantics only. Attempts to grant remote edit authority MUST be rejected or explicitly normalized with a response explaining the resulting view-only permission; choose one API behavior before implementation.

**SHR-004 — EXISTING.** The target relation is immutable on updates for `trail_share`, `list_share`, and `trail_link_share`. The reviewed target hook compares the proposed value to the stored original before continuing to persistence. (S12)

**SHR-005 — PROPOSED.** A PATCH/UPDATE that omits the target or repeats the same target MUST remain eligible for normal validation. Changing, clearing, or replacing it MUST fail even if the acting user owns both objects. Moving a share requires a separately authorized delete/create operation.

**SHR-006 — PROPOSED.** Validation MUST consider the complete proposed share: original-record authorization, target immutability, new recipient, permission, and schema constraints. Combining a target change with a recipient, permission, or token change MUST NOT evade any check.

**SHR-007 — EXISTING.** Actor-share update hooks run target validation followed by remote-recipient validation. Successful updates continue through PocketBase validation; they do not use the create-share announcement path. The current tests accept an `edit` permission value on a public remote share; the view-only product limitation therefore MUST NOT be confused with a proven server rejection of that value. SHR-003 proposes an explicit API resolution. (S11; S12)

**SHR-008 — PROPOSED.** Permission-only updates MUST preserve a share's target and recipient. They MUST NOT make a private object public, change authorship, modify nested grants, or imply that an old remote cache has received an update.

**SHR-009 — PROPOSED.** Existing private-to-remote actor shares MUST NOT be extended through ordinary updates. Repairing one by changing its recipient to an authorized local actor, or deleting it, MAY be allowed. The acting user must still have share-administration permission.

**SHR-010 — EXISTING.** Sharing a list with a local actor through the reviewed modal/store creates missing view grants for the sender's own private trails present in the loaded list expansion. Existing grants are skipped; public trails and private trails owned by somebody else are not auto-granted. (S13)

**SHR-011 — PROPOSED.** The product contract SHOULD apply that local-only auto-grant policy to all eligible current list members, not merely whichever trails happened to be loaded in a UI expansion. Establish an authoritative membership enumeration before claiming complete automatic sharing.

**SHR-012 — PROPOSED.** Auto-grants MUST be limited to private trails the sender is entitled to share. A foreign private trail MUST NOT be made public, re-shared, or granted through list membership; its visibility remains governed by its own authorization.

**SHR-013 — EXISTING.** The reviewed remote list-sharing flow creates the list share and refreshes its displayed shares without attempting private trail-share requests. The dialog rejects a private list's remote recipient before sending its request; the server independently validates the restriction. (S13)

**SHR-014 — OPEN.** Adding a private trail to an already shared list needs the same local/remote distinction, but the existing edit-page path is not proven equivalent by the modal/store tests. Specify the eligible local recipients, consent prompt, and failure handling before treating create/share/edit flows as interchangeable.

**SHR-015 — OPEN.** Decide whether changing a list share from view to edit should affect any automatically created trail grants. Recommended draft default: it changes the list grant only; trail grants remain view unless separately authorized. This avoids silently widening trail-edit rights.

**SHR-016 — EXISTING.** Product documentation says unsharing a list does not automatically unshare its trails. The reviewed UI follows this model and does not delete trail grants when deleting a list share. The documentation's broader “all trails are shared” wording needs the local/ownership qualifications above. [S2], code evidence S13.

**SHR-017 — PROPOSED.** Deleting a list share MUST remove that grant without silently removing independently existing trail grants. If revocation of derived grants is later offered, it requires explicit grant provenance and user confirmation of the affected grants, not inference from current list membership.

**SHR-018 — PROPOSED.** Share-token rotation MUST invalidate the old token for subsequent authorized requests to the issuing instance while retaining the immutable trail target. Token deletion MUST remove that local capability. Neither action promises deletion of prior external downloads.

**SHR-019 — PROPOSED.** Successful operations MUST refresh the displayed recipients and effective permissions from the server. Failure of an operation MUST NOT be presented as success; if list-share creation succeeded but a later auto-grant failed, the UI MUST report that partial state and refresh it accurately.

**SHR-020 — OPEN (D-16, API compatibility D-10).** Choose whether list sharing plus local trail auto-grants becomes one server operation or remains a sequence with explicit partial-success results and safe retries. The current client sequence is not transactional and MUST NOT be documented as all-or-nothing. (S13)

**SHR-021 — PROPOSED.** The UI MUST explain public-only cross-instance access, local-only private trail auto-grants, view-only remote recipients, and independent trail grants after list unsharing. Disabling a control is user guidance; it MUST NOT substitute for server authorization.

## 11. Acceptance and policy decisions

**ACC-044 — PROPOSED.** Contract tests MUST exercise the actual registered schema at the relevant migration boundary and current schema, plus real Records API hooks for writes. Synthetic fixtures may isolate a behavior but MUST NOT be the only evidence that production rules implement the matrix.

**ACC-045 — PROPOSED.** Negative cases MUST include anonymous/empty identity, remote actors without local users, unrelated local users, no shares, revoked grants, wrong tokens, cross-object tokens, nested relations, and unverified/withheld remote parents. Output tests MUST cover files and aggregates separately from record JSON.

Before adoption, record explicit decisions for ACC-008, ACC-017, ACC-023, ACC-025, ACC-030, ACC-038, ACC-042, SHR-003, SHR-014, SHR-015, and SHR-020.
No implementation should silently settle these questions by broadening or narrowing an existing user's rights.
The shared decision registry is [07-decisions.md](07-decisions.md); D-04 covers child-author maintenance, D-07 covers file/search enforcement, and D-10 covers API contract choices.
Acceptance scenarios belong in [05-acceptance-tests.md](05-acceptance-tests.md); this document defines the behavior they must distinguish.
The replica document owns state transitions and refresh scheduling; this document owns how their eligibility decisions constrain access.

## 12. Sources and evidence boundaries

All code entries below refer to the reviewed local integration commit `865251d49dc6c1d734cccb85ee51c680c62d35a6` (abbreviated `865251d49`).
That integration commit is not assumed to be published on GitHub. Paths and line numbers are repository-relative evidence locators, not links to a moving IDE checkout.
[08-evidence.md](08-evidence.md) records reproducible `git show` evidence; product documentation links below are relative to this draft's repository location.

[S1]: ../../docs/src/content/docs/use/share-trails.md "Published sharing limitations: public-only and view-only across instances"
[S2]: ../../docs/src/content/docs/use/lists.md "List sharing and independent trail grants"
[S3]: ../../docs/src/content/docs/use/statistics.md "Statistics and completed-trail fallback semantics"

- **S4:** `db/migrations/1789200002_guard_anonymous_read_rules.go:29` at `865251d49`: current read guards; `:50` and `:57` show child author paths.
- **S5:** `db/routes/remote_expand.go:12` at `865251d49`: custom-route response enrichment; `db/util/meilisearch.go:416` shows internal list expansion.
- **S6:** `db/migrations/1789200001_updated_tags_view.go:10` and `db/migrations/1789200003_updated_summit_logs_link_share.go:10` at `865251d49`: anonymous tag views and explicit summit-log token access.
- **S7:** `web/src/routes/api/v1/files/[collection]/[record]/[file]/+server.ts:52` and `db/migrations/1742167087_collections_snapshot.go:168` at `865251d49`: existing file proxy and historical unprotected file fields; current byte-level guarantees require separate verification.
- **S8:** `db/routes/search_token.go:11` and `db/util/meilisearch.go:320` at `865251d49`: principal-dependent index filters and internal indexing expansion.
- **S9:** `web/src/routes/api/v1/profile/[handle]/stats/+server.ts:94`, `web/src/hooks.server.ts:55`, `web/src/lib/models/api/base_schema.ts:4`, and `web/src/lib/util/activitypub_server_util.ts:17` at `865251d49`: local viewer context, parsed query scope, remote delegation, and anonymous remote-profile restriction.
- **S10:** `db/routes/remote_list.go:109` and `db/federation/actor.go:197` at `865251d49`: private-profile handling distinct from object view rules; these references do not settle all discovery policy.
- **S11:** `db/hooks/share_guard.go:18`, `db/main.go:129`, and `db/hooks/share_guard_update_test.go:19` at `865251d49`: remote-recipient validation, hook ordering, and currently accepted public remote permission updates.
- **S12:** `db/hooks/share_target.go:9` and `db/main.go:129` at `865251d49`: immutable targets across all three share collections.
- **S13:** `web/src/lib/stores/list_share_store.ts:46` and `web/src/lib/components/list/list_share_modal.svelte:64` / `:86` at `865251d49`: local-only auto-grants, remote flow, independent trail grants after deletion, and UI refresh.

The existing [federation documentation](../../docs/src/content/docs/develop/federation.md) describes exchanged objects and on-demand loading; this draft supplements it with proposed access decisions.
The migration-based evidence harness is `db/migrations/migration_test_helpers_test.go:11` at `865251d49`; its existence does not establish coverage of every proposed file, search, or aggregate guarantee.
