# Draft: federation events and delivery

**Status: DRAFT — proposed contract, not approved policy or implemented behavior.**
**Reviewed code:** local integration commit `865251d49dc6c1d734cccb85ee51c680c62d35a6`.

**Scope:** New durable delivery, receipt processing and the negotiated withdrawal extension are deferred options. P0–P2 retain current event behavior while improving import/output boundaries. Later workers may run within the existing single backend process; a distributed worker system is not a prerequisite.
The requirements below describe the intended event model for review.
They do not claim that the existing PRs implement durable delivery, withdrawal, or every recipient transition.
Normative words apply to this proposed Wanderer contract; they do not assert additional ActivityPub standards requirements.

Related chapters: [access](01-behavior-and-access.md), [replicas](02-replica-lifecycle.md), [architecture](04-architecture-and-data.md), [acceptance tests](05-acceptance-tests.md), and [open decisions](07-decisions.md).

## 1. Scope, identities, and source authority

**EVT-001 — Public-only federation.** V1 MUST federate only public trail/list content and authorized public-parent interactions.
A targeted remote share MUST mean a notification about a public object, not a private grant or remote editing capability.
Local account shares and local trail-link tokens MUST remain local authorization mechanisms.
No event may transfer a local user session, API credential, or share token to another instance.

**EVT-002 — Separate identities.** An object IRI identifies the object; an activity IRI identifies one logical action; a delivery job identifies intended notification of one recipient across its attempts.
Retries MUST retain the original activity IRI and immutable activity body.
A later business action MUST receive a new activity IRI, even when it references the same object.
Local PocketBase IDs MUST NOT substitute for these cross-instance identities.

**EVT-003 — Ownership binding.** A replica MUST retain the established origin and owner actor of each object independently of the latest received payload.
An event claiming a different owner MUST NOT silently replace that binding.
An authenticated sender may create an interaction of its own; that does not authorize it to update, withdraw, or delete the referenced trail.
Unknown ownership MUST be resolved or quarantined before applying an owner-only effect.

**EVT-004 — Positive visibility evidence.** `Create`, `Update`, `Announce`, and public addressing MUST NOT by themselves establish that a stored replica is currently public.
They MAY create an `unverified` stub, record a candidate revision, or request canonical revalidation.
Only a validated origin public snapshot or qualifying canonical `304` may establish or renew `last_public_verified_at`, as defined in chapter 02.
Local writes, actor fetches, list summaries, activity receipt, and successful inbox delivery MUST NOT renew it.

**EVT-005 — Shared replica terms.** This chapter uses `visibility_state = unverified | public | withheld | deleted` and `materialization = stub | metadata_ready | detail_ready`.
Materialization does not establish visibility; visibility does not establish detail completeness.
Exceeding the permitted verification age suppresses serving through the freshness gate without inventing a visibility-state transition.
All feed, notification, search, expansion, and media presentation MUST respect the same applicable read gate.

## 2. Event table

The table is **PROPOSED**, including rows that extend current behavior.
“Known recipients” means actors recorded in the origin's object-distribution ledger, including recipients of uncertain delivery attempts.
It does not mean every server that may have independently fetched or copied a public object.
“Verify” means canonical public revalidation under chapter 02, not trusting an activity payload.
Recipient sets MUST be deduplicated by actor while preserving separate local-recipient effects.

| Requirement / trigger | Origin effect | Activity | Intended remote recipients | Receiving Wanderer effect | Idempotency / ordering rule |
| --- | --- | --- | --- | --- | --- |
| EVT-006: create a public trail or list | Commit public object and delivery intent | `Create` with its public representation | Accepted followers; trail mentions where supported | Record candidate, verify, then expose eligible representation and feed entry | One activity; one object identity; one effect per local recipient |
| EVT-007: edit a public trail or list | Commit new revision and delivery intent | `Update` with complete new public representation | Current followers and applicable mentions; known recipients needing revision notice | Schedule verify; replace only validated sections; preserve local annotations | Replayed or older revision cannot overwrite a newer committed snapshot |
| EVT-008: create/edit a private trail or list | Commit authorized local change | No public content activity | None | No new remote entitlement | No event created accidentally by an indexing/import hook |
| EVT-009: first publish a previously private object | Commit public revision | `Create` for its first federation publication | Current followers and applicable mentions | Verify before public activation | First publication is tracked independently of local creation time |
| EVT-010: public object becomes private | Close origin public gates; advance visibility generation | Negotiated withdrawal proposal; see D-05 below | Known recipients supporting the extension | Authenticate owner, mark withheld, invalidate serving and schedule verification | Prior positive jobs/results cannot reopen the object |
| EVT-011: republish a withdrawn object | Commit a new public revision of the still-existing IRI | `Update` for a previously published object | Current public audience and relevant known recipients | Verify current public state before leaving withheld | Same IRI allowed only because no real deletion occurred |
| EVT-012: targeted remote trail/list share | Commit a view-only notification relationship | `Announce` referring to the public object | Selected remote actor | Verify referenced object; create recipient-scoped share/notification effect | Unique active relationship; duplicate delivery creates no duplicate alert |
| EVT-013: remove a targeted remote share | Remove local relationship | `Undo` of its original `Announce` | Original recipient | Remove that share/notification effect; retain independently public object | Undo recorded even if its Announce has not arrived |
| EVT-014: change a targeted share recipient | Commit removal of old relationship and creation of new one | `Undo` old `Announce`, then a new `Announce` | Old recipient and new recipient respectively | Reconcile only each addressed recipient's relationship | Separate activity IDs; stale old jobs fenced; D-06 |
| EVT-015: update local share permission or rotate local link token | Commit authorized local change | None | None | No federation effect | Share target remains immutable; no remote edit grant introduced |
| EVT-016: really delete an object | Close origin gates; retain minimal deletion identity | `Delete` identifying the object | Known recipients, including former followers/share recipients where known | Authenticate owner; set deleted; remove visible projections | Deletion terminal for this IRI; repeated Delete harmless |
| EVT-017: follow an actor | Commit pending follow and intent | `Follow` | Followed actor | Validate local target; create/resolve follow according to actor policy | One active relation per actor pair; retain activity identity |
| EVT-018: accept/reject a follow | Commit explicit result of that follow | `Accept` / `Reject` referring to original `Follow` | Original follower | Change only matching pending relation | Late response cannot restore a follow already undone |
| EVT-019: unfollow an actor | End local follow | `Undo` of original `Follow` | Followed actor | End subscription/future fanout; preserve public cached objects | Undo binds to that follow generation, not a later follow |
| EVT-020: like/unlike a public trail | Commit/remove actor's interaction | `Like` / `Undo` of original `Like` | Trail origin/author | Validate actor and target; maintain one active like and correct count | Duplicate Like/Undo does not multiply or underflow counts |
| EVT-021: create/edit a comment | Commit comment authored by the actor, under public-parent policy | `Create` / `Update` of comment | Trail author and explicit mentions | Validate comment authority, parent, and recipient; verify eligible parent before exposure | Object revision update does not create a new comment or repeated notification |
| EVT-022: create/edit a summit log | Commit log authored by the actor, under public-parent policy | `Create` / `Update` of summit log | Log author's accepted followers, trail author, explicit mentions | Validate log authority and public-parent policy; materialize independently | Distinct author-owned object; no authority to edit the trail |
| EVT-023: delete a comment or summit log | Commit real deletion of that child | `Delete` of child IRI | Known recipients of that child | Validate child owner; delete child projection, not parent trail | Replayed delete cannot erase sibling or parent objects |
| EVT-024: add/remove/reorder list membership | Commit list revision; preserve each trail's own visibility | Public list `Update`, if list is public | List audience and relevant known recipients | Verify list projection; update edges in authoritative order | Removal deletes an edge, not its referenced trail or independent share |

## 3. Recipient and interaction policy

**EVT-025 — Notification versus object ownership.** A targeted `Announce` MUST NOT grant the announcer ownership of its object.
Receiving an Announce MUST NOT force `public = true`, copy private child payloads into a public placeholder, or bypass canonical validation.
A missing or private source object may leave an unresolved notification receipt, but MUST NOT yield a visible private object or payload-bearing notification.
An unsupported third-party resharing format MUST NOT be silently interpreted as a Wanderer private grant.

**EVT-026 — Sharing a list.** Remote list sharing MUST announce only the list's public projection.
It MUST NOT synthesize private trail shares or include hidden trail metadata, membership counts, media, or expansion payloads.
Automatic local grants for the owner's private trails are a local command with its own authorization; they are not federation events.
Removing an Announce or list membership MUST NOT erase a public trail that has another reason to remain cached.

**EVT-027 — Recipient change is an explicit delta (D-06).** The proposed remote-recipient update behavior is `Undo` to the old recipient and a fresh `Announce` to the new recipient.
These intents MUST commit with the relationship update; recipients MUST NOT be inferred later from the mutated share record.
An existing private remote share MAY be repaired into a local share without announcing private content.
Its old remote notification may receive a minimal Undo, which is cleanup of an earlier action rather than a new private grant.
This is a proposed extension: the reviewed baseline does not send an Announce on share updates.

**EVT-028 — Public interactions need parent eligibility.** Remote comments, likes, and summit logs MUST be accepted only under the v1 public-parent interaction policy.
A private parent at the origin MUST NOT become public because a foreign interaction references it.
Hiding a parent MUST hide dependent public presentations of its children even where the child's own record remains stored.
The independent authorship and retention questions for children remain in D-04; a trail owner does not thereby acquire the log author's deletion authority.

**EVT-029 — Follow acceptance.** The existing documented automatic acceptance of Follow is the default compatibility target.
The receiver MUST nevertheless validate that the named local actor exists and permits that interaction.
Private-profile discovery/acceptance rules MUST follow chapter 01, not an assumption that every addressed path is a valid actor.
An Accept without a matching authenticated pending Follow MUST NOT create a subscription.

**EVT-030 — Recipient ledger.** The origin MUST retain the object IRI, logical recipient actor, reason for distribution, activity identity, and relevant generation needed for later updates, withdrawal, and deletion.
The ledger MUST include a recipient before a delivery may have escaped the process; an ambiguous network result cannot prove non-delivery.
Leaving a follower set or removing a share MUST NOT erase the cleanup information immediately.
Full payload retention is unnecessary for this ledger; retention and privacy limits belong to D-11.

## 4. Withdrawal and actual deletion

### D-05 remains OPEN: reversible withdrawal protocol

The recommendation is an owner-authenticated, negotiated Wanderer extension provisionally named `wdr:Withdraw`.
Its namespace, version, capability advertisement, required fields, signature binding, and interoperability tests remain **OPEN**.
The name is illustrative, not a registered or implemented vocabulary term.
It MUST NOT be sent as an assumed universal ActivityPub operation.

The standards distinction matters: server-to-server `Update` carries the complete new representation; partial-update behavior belongs to client-to-server interaction.
A sparse Update is therefore not a generic standards-defined withdrawal message. [ActivityPub §7.3](https://www.w3.org/TR/activitypub/#update-activity-inbox)
ActivityPub also requires authority checks for Delete and ties Undo to the actor of the original action. [§7.4](https://www.w3.org/TR/activitypub/#delete-activity-inbox), [§7.12](https://www.w3.org/TR/activitypub/#undo-activity-inbox)
The withdrawal extension and all timing defaults here are Wanderer design proposals.

**EVT-031 — Minimal withdrawal intent.** If D-05 is accepted, withdrawal MUST identify the object, authenticated owner, action identity, and agreed ordering/generation evidence.
It MUST NOT carry the newly private route, name, description, media, children, token, or local recipient grants.
It MUST address only known prior recipients, without the Public address.
An origin MUST stop all public representations locally before exposing successful completion of the privacy change.

**EVT-032 — Receiver withdrawal.** For a non-deleted object, an accepted withdrawal from the established owner MUST immediately set `visibility_state = withheld` and invalidate cached public projections.
It MUST advance the local fencing generation and schedule canonical revalidation, without waiting for a network round trip to hide content.
An unavailable source MUST leave the object withheld.
A later activity alone MUST NOT republish it; restoration needs fresh positive canonical verification.
For an already deleted identity, record the accepted receipt while retaining `deleted`; do not schedule restoration or weaken its tombstone because a delayed withdrawal arrived.

**EVT-033 — No fake deletion.** A public-to-private switch MUST NOT emit `Delete` as if the object had ceased to exist.
A withdrawn object can legitimately become public again under the same IRI; a truly deleted object cannot under this proposal.
Foreign implementations may treat deletion as irreversible, making that shortcut incompatible with reversible privacy changes.
A fully redacted replacement Update is an alternative requiring separate semantics and interoperability tests, not an approved substitute here.

**EVT-034 — Legacy peers and bounded verification.** Without negotiated withdrawal, the origin MUST still close its public gates and fence pending public jobs.
A conforming upgraded Wanderer receiver limits stale public serving through D-03: proposed maximum age is seven days since `last_public_verified_at`.
That bound does not apply to unupgraded peers, unrelated ActivityPub software, backups, or completed downloads.
No UI or documentation may promise universal erasure or immediate concealment of copies on servers outside this contract.

**EVT-035 — Actual Delete.** A real deletion MUST create durable cleanup intent even if the object was already private or withdrawn when deleted.
Its activity SHOULD contain only the identifiers and authority information necessary to authenticate the deletion, not a copy of deleted content.
The origin MUST retain enough identity and owner evidence to reject future reuse and service legitimate deletion verification.
Receiving Wanderer MUST establish authority before setting `visibility_state = deleted`.

**EVT-036 — Irreversible identity.** Once a real Delete is accepted, the same object IRI MUST NOT return to public or withheld through later Create/Update/Announce, fetch completion, restore, or administrative retry.
Recreating deleted content requires a new object IRI.
An origin `404` alone MUST NOT be interpreted as an authenticated irreversible deletion; chapter 02 defines its conservative result.
Deleting a local replica's bytes for storage cleanup is not a federated Delete and MUST NOT impersonate the origin.

**EVT-037 — Stale work fencing.** Every job capable of publishing content or committing a snapshot MUST capture a relevant generation and compare it before its effect.
Privacy changes, real deletion, superseding revisions, and relationship withdrawal MUST invalidate incompatible pending work.
A superseded attempt, including one with an expired lease where leases are used, MUST NOT commit using a newer attempt's authority.
Bytes already in flight cannot be recalled; cleanup events and receiver verification address that residual race, rather than claiming atomic control across servers.

## 5. Durable outbound delivery

**EVT-038 — Transactional job outbox.** A local mutation and its durable delivery intent MUST commit in one local transaction.
A rolled-back command MUST leave neither an externally visible activity nor dispatchable work.
An accepted command MUST survive a process crash before the first network send.
Network delivery MUST occur after commit and MUST NOT keep a PocketBase write transaction open across remote I/O.

The **delivery job outbox** is an internal queue, not the ActivityPub actor's `/outbox` collection.
The actor outbox is a protocol-visible collection with its own access and representation rules.
Storing an `activitypub_activities` row alone does not demonstrate a durable retry queue.
Protocol collection retention MUST NOT control whether a committed delivery can still be retried.

**EVT-039 — Durable recipient expansion.** Transactional intent MUST preserve an immutable logical audience or a deterministic, versioned expansion task.
Large follower lists MAY be expanded in bounded batches after commit, provided membership changes cannot silently rewrite the intended event audience.
Per-recipient jobs MUST be uniquely keyed by logical activity and recipient actor.
If several actors use one shared inbox, transport batching MAY combine sends while retaining each logical recipient and outcome.

**EVT-040 — Delivery record.** A delivery MUST track activity IRI, actor, recipient, validated destination, object/relationship generation, state, attempt count, next attempt time, current execution claim, deadline, and bounded last-error classification. The initial single-process worker may recover interrupted work at exclusive startup; an expiring lease is needed only for a chosen reassignment/multiple-process model, as in RPL-028.
Proposed states are `queued`, `in_flight`, `retry_wait`, `delivered`, `superseded`, and `failed`.
Claim and completion MUST be atomic; expired claims MUST become recoverable.
The activity body MUST remain immutable across attempts; newer content requires a newer activity.

**EVT-041 — Dispatch authorization.** Before each attempt, the worker MUST confirm that the activity remains permitted by current origin state and its generation.
Stale public content jobs MUST become superseded rather than sending an earlier public payload after withdrawal.
Delete, Undo, and agreed withdrawal jobs MUST remain dispatchable when their underlying content/share row no longer exists.
Signing material MUST be loaded through the actor's authorized credential mechanism and MUST NOT be stored in job payloads.

**EVT-042 — Recipient address safety.** Inbox resolution and redirects MUST pass the same outbound-network and actor-origin validation as other federation fetches.
A destination change MUST be validated against the logical recipient before use.
Delivery MUST be bounded by per-destination and global concurrency, request size, response size, and timeout budgets.
An activity-supplied URL MUST NOT directly become an unrestricted delivery target or trusted key endpoint.

**EVT-043 — Success semantics.** Supported successful inbox responses MUST mark transport acceptance, not proof of rendering, notification, full synchronization, or retention.
Local command success means local state and intent are durable, not that every remote instance is online.
Partial recipient failure MUST NOT undo the original local change or resend to already delivered recipients unnecessarily.
Operator status MUST distinguish pending, transport-accepted, superseded, and permanently failed work.

### D-08 proposed operational defaults

These are tunable starting values, **not ActivityPub requirements** or measurements of existing production limits.
The recommendation is at most **8 attempts total, including the first, within 72 hours of enqueue**.
For retry number `n >= 1`, the base delay is `min(24 hours, 5 minutes * 2^(n-1))`, measured from completion of the preceding attempt.
Apply bounded jitter, for example a multiplier uniformly sampled between `0.5` and `1.5`, with the final computed delay capped at 24 hours.
Only actual network attempts consume the attempt budget; a worker crash before sending consumes no fictitious attempt, while an uncertain sent request counts.

**EVT-044 — Failure classification.** Transport failures, temporary name-resolution failures, request timeouts, `408`, `429`, and `5xx` SHOULD schedule a bounded retry.
Other `4xx` responses SHOULD be terminal, subject to an explicit, bounded endpoint/key refresh policy rather than blind repeated retries.
A response body MUST NOT be treated as an executable instruction or a trusted recipient reconfiguration.
The final classification and deadline MUST be observable for operator review.

**EVT-045 — Retry-After.** A valid recipient `Retry-After` MUST act as a not-before time when the response is classified retryable.
The sender MUST choose the later of its own backoff and that time; it MUST NOT cap a recipient's longer delay by retrying earlier.
If the recipient delay falls beyond the 72-hour job deadline, fail the bounded job with that reason instead of ignoring the advice.
Invalid retry timing falls back to local backoff; arithmetic and accepted date ranges MUST be bounded.

**EVT-046 — Expiry and manual recovery.** Attempt exhaustion or the 72-hour deadline MUST produce terminal failure, not silent disappearance.
Terminal job records SHOULD remain inspectable for seven days after completion under the initial D-08 proposal, with bounded metadata and without unnecessary payload copies.
Deletion/withdrawal identity records and inbox replay protection have separate retention requirements; seven-day job cleanup MUST NOT erase them.
An authorized manual retry MAY reopen eligible delivery work while retaining the original activity ID, but MUST repeat generation and authority checks.

## 6. Inbound authentication and durable processing

**EVT-047 — Authenticate before effects.** The ingress boundary MUST validate request integrity, sender/key ownership, freshness/replay constraints, payload limits, and the binding between authenticated actor and activity actor.
An internal forwarded request MUST authenticate the forwarding boundary as well as preserve verified sender evidence.
A valid HTTP signature is sender identity evidence, not arbitrary permission over objects or local users.
No notification, feed entry, object visibility change, or fetch fanout may occur before the request passes the relevant validation.

**EVT-048 — Validate recipients.** The receiver MUST resolve the intended local actor from a validated inbox route or verified shared-inbox audience expansion.
It MUST check that the activity is admissible for that recipient under the event policy, addressing, subscription, and referenced object relationships.
Public content can be discoverable without entitling a sender to manufacture arbitrary local-user notifications.
An activity addressed to actor A MUST NOT acquire actor B's effects through an untrusted query, forwarding header, or payload field.

**EVT-049 — Validate action authority.** Create/Update/Delete/withdrawal of an object MUST use the object's established owner authority.
Undo MUST reference an action attributable to the authenticated actor; a foreign Undo cannot cancel another actor's Like, Follow, or Announce.
Accept/Reject MUST come from the actor targeted by the original Follow and name that relationship generation.
Ownership checks MUST cover child objects independently of their parent trails.

**EVT-050 — Durable receipt before acknowledgement.** A successful inbox acknowledgement MUST mean the accepted envelope and required recipient processing work are durably recorded, or that a matching accepted duplicate is already known.
Canonical fetching MAY happen after acknowledgement; its failure MUST leave retryable processing state rather than lose the accepted action.
A process crash after acknowledgement MUST not discard the action.
Malformed, unauthorized, and unsupported actions MUST have explicit handling and MUST NOT be acknowledged as completed business effects.

**EVT-051 — Two levels of idempotency.** Store an authenticated envelope once by `(authenticated_actor_iri, activity_iri)` and store processing outcome separately by `(activity_iri, local_recipient_actor_iri, effect_kind)`.
The activity identity MUST remain bound to its original authenticated actor and immutable digest; reuse with a conflicting actor or body MUST be rejected or quarantined.
Global envelope deduplication MUST NOT skip the second legitimate local recipient on the same instance.
A duplicate network request MUST be able to resume incomplete recipient work without repeating completed effects.

**EVT-052 — Transactional recipient effects.** Updating a relationship, creating its feed/notification effect, and recording that effect's receipt MUST commit atomically where they share the local database.
Email or other external notification work MUST use a separate durable, idempotent intent instead of sending within that transaction.
There SHOULD be one visible feed row per `(recipient, object)`, backed by separate unique reason edges such as an accepted follow or a particular Announce. Removing one reason MUST retain the row when another eligible reason remains. Notification identity SHOULD distinguish recipient, logical action, and notification kind.
Repeated Updates MUST NOT create repeated “new object” notices; distinct legitimate actions may create distinct notices according to user preferences.

**EVT-053 — Relationship uniqueness.** One active Like per actor/object and one active Follow per actor pair MUST be enforced independently of activity-ID deduplication.
Share/Announce effects MUST retain the originating action and recipient, so Undo removes only its own contribution.
Counts MUST derive from unique active relationships or transactional changes to them, not blindly increment for every delivery.
Multiple deliveries with newly minted activity IDs MUST not defeat relationship uniqueness.

**EVT-054 — Undo before original.** An authenticated Undo received before its referenced action MUST retain a recipient-scoped cancellation marker sufficient to suppress that delayed original.
Receiving the original later MUST NOT create its feed, notification, count, or relationship effect.
A later genuinely new Follow, Like, or Announce has a new logical action and may be handled independently if authorized.
Neither arrival order nor an untrusted remote timestamp alone may decide whether a relation has been restored.

## 7. Out-of-order data and reconciliation

**EVT-055 — Events trigger verification, not blind replacement.** The receiver MUST validate event-object identity and origin before scheduling canonical work.
An event payload MAY provide candidate metadata but MUST NOT overwrite local ownership, local grants, sync state, completion flags, or verified visibility.
Out-of-order public Create/Update MUST converge through the current canonical representation rather than repeatedly importing older embedded objects.
Only negotiated trustworthy revision evidence may support direct revision comparisons; foreign timestamps alone are insufficient.

**EVT-056 — Canonical result fencing.** Fetch jobs MUST capture the replica's relevant generation and verify it again before committing.
A public fetch started before an accepted withdrawal or Delete MUST NOT complete afterward and reopen visibility.
An older complete snapshot MUST NOT overwrite newer committed sections merely because its HTTP request finished later.
When ordering cannot be established safely, discard the conflicting result and revalidate rather than guessing.

**EVT-057 — Negative and positive evidence differ.** Accepted owner withdrawal/Delete can reduce visibility immediately.
An ordinary public Update is only a reason to revalidate and MUST NOT reverse negative state or reset verification age by itself.
A withheld object can return to public only through fresh qualifying canonical evidence after the latest invalidation; a deleted IRI cannot return.
Network failure alone does not prove private/deleted status, and age-based suppression remains separate from `visibility_state`.

**EVT-058 — Shared-instance isolation.** Object synchronization MAY be deduplicated across recipients because an object's origin state is shared.
Recipient-specific shares, follows, feed entries, notifications, receipts, and Undo effects MUST remain independently addressable.
The presence of a verified public replica for recipient A does not authorize fabricating an addressed interaction for recipient B.
The absence of a share for A does not require deleting a public replica still eligible for other uses.

**EVT-059 — No echo federation.** Importing a remote object or processing a remote activity MUST NOT make the local instance its author or emit a new origin Create/Update automatically.
Local reactions or deliberate announcements may create new locally authored interactions under their own authorization.
Any supported forwarding path MUST preserve original provenance and apply explicit loop, audience, and resource limits.
Index updates and cache reconciliation MUST NOT implicitly broadcast content.

## 8. Baseline evidence and implementation boundary

The following observations are read-only findings from the reviewed local integration commit, not a complete protocol audit.
References use `repository-file:line @ 865251d49`; this local merge is not assumed to exist on GitHub.

| Baseline observation | Evidence | Proposed change or retained rule |
| --- | --- | --- |
| Public trail/list activity creation skips currently private objects | `db/federation/create.go:21`, `:349` | Retain public-only scope; add explicit publication/withdrawal generations |
| Current PostActivity starts asynchronous goroutines and reports recipient failures through logs | `db/federation/activity.go:55`, `:105`, `:136` | Durable per-recipient jobs, restart recovery, retry, and observable outcomes |
| Current create-share hook sends Announce after local share processing | `db/hooks/trail_share.go:12`, `:41`; `db/federation/announce.go:25` | Preserve public notification meaning; commit intent with the command |
| Current delete-share hook updates local state without Undo Announce | `db/hooks/trail_share.go:50` | Proposed recipient cleanup and recipient-change delta, D-06 |
| Incoming create/update uses the local recipient when inserting feed entries | `db/federation/create.go:433`, `:757` | Retain recipient scope; add explicit durable receipt/effect uniqueness |
| Current trail Delete producer exits when the record is private | `db/federation/delete.go:16` | Real deletion cleanup must include previously distributed, now-private objects |
| Existing public cache fallback distinguishes completed copies from placeholders | `db/routes/remote_trail.go:40` | Preserve the distinction; introduce explicit states and verification age from chapter 02 |

The existing user documentation already distinguishes remote public notification from local collaboration: [Share trails](../../docs/src/content/docs/use/share-trails.md).
The technical documentation describes message examples and on-demand loading: [Federation](../../docs/src/content/docs/develop/federation.md).
Those documents do not establish the new delivery, withdrawal, or freshness guarantees proposed here.

## 9. Review exit criteria

**EVT-060 — Acceptance boundary.** Before this chapter becomes an accepted contract, maintainers MUST resolve D-05, D-06, D-08, and the related freshness/retention choices in the decision register.
Acceptance tests MUST exercise restarts, duplicate delivery, multiple local recipients, reordered actions, uncertain network results, withdrawal during a fetch/send, and Delete followed by stale public work.
Rolling-version tests MUST distinguish extension-aware Wanderer peers, legacy Wanderer peers, and supported foreign ActivityPub software.
Passing the current PR suites alone MUST NOT be represented as satisfying these future lifecycle and delivery guarantees.
