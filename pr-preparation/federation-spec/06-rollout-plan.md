# Incremental implementation and rollout proposal

Status: **DRAFT / PROPOSED**. The recommended next project is **P0–P2 only**: agree the initial decisions, reuse real-schema fixtures, and introduce allowlisted import and stable response boundaries. P3–P8 are future options requiring separate effort and maintenance review, not an agreed roadmap or an obligation created by finishing P2. This document creates no issues, PRs, deployments or migration approvals. Immediate fixes remain separate from architecture work.

## 1. Starting position

The reference integration tree is `865251d49`. It contains the then-prepared read-rule fix (#1232), independent share-target fix, #1222 sharing changes, #1224 sync-state/response-isolation changes, and #1223 expansion changes. This is a reproducible local baseline, not a claim about their current GitHub merge status.

On 2026-09-13 the IDE checkout was observed at `ddd1929d8` on the #1224 branch, adding a deterministic background-response regression and an injectable full-sync HTTP client. Preserve that additional work. Also re-check current branch heads before implementing this plan; the specification task intentionally does not rebase or edit them.

**PLAN-001 — Keep the next commitment bounded.** Independently understood bug/security fixes proceed on their own merits; this proposal is not a prerequisite for patch releases. Recommend only P0–P2 for the next maintainer commitment. After P2, review its value and outstanding risks before accepting any further package. No acceptance of P0–P2 implies acceptance of P3–P8, all 52 catalog scenarios, or a two-instance harness.

## 2. Recommended next work: P0–P2

The aim is to reduce the opportunity for another import/response-boundary mistake within the current application. It does not require a new replica state machine, durable queue, search proxy, file-access redesign, withdrawal protocol or distributed worker framework.

| Package | Purpose and concrete output | Dependencies | Changes behavior/schema? | Exit evidence |
| --- | --- | --- | --- | --- |
| P0 | Discuss D-01/D-02/D-03; record agreed direction, explicit deferrals and the existing policy P2 must preserve; correct documentation only where current behavior is established | Maintainer review | Documentation only | Decision status is explicit; no proposed age/default/error behavior is presented as already implemented |
| P1 | Reuse real-schema fixtures and existing router regressions; inventory only the trail/list import and response entry points to change | P0's explicit existing-policy baseline | Tests/tooling only | Bounded selection in chapter 05 section 2, with current hook/API regressions retained |
| P2 | Extract allowlisted trail/list input mapping and checked response projection; make import working copies and response objects explicit | P1 | No intended policy/API change; no required production schema | Forged protected fields and a newly added fixture field cannot enter local state; actual route isolation and successful response shapes remain covered |

P0 need not resolve D-04–D-16 to permit P1/P2. If D-02/D-03 remain unsettled, record them as deferred and keep current completeness/cache behavior in P2; do not invent an implicit decision in implementation. Adopting a future direction in a decision record is not approval to implement all its dependent packages now.

**P2 stop point:** current import and output boundaries are explicit, their field ownership is reviewed, and relevant existing tests plus the small new allowlist/isolation cases pass. Runtime behavior that belongs to a later policy package remains documented as a gap. In particular, query-dependent full-sync scope, network work inside transactions, replica eligibility, file/search exposure and non-durable delivery do not become solved merely because helpers have been extracted. A narrow independently useful fix discovered during P1/P2 can be proposed separately with its own regression.

## 3. Deferred options: P3–P8

These packages describe a substantially larger program. Before starting one, maintainers select an owner, acceptable review/maintenance cost, behavior decisions, concrete surface and test budget. Its dependency list is conditional on that selection; it does not impose work on P0–P2.

| Package | Purpose and concrete output | Dependencies if selected | Changes behavior/schema? | Exit evidence |
| --- | --- | --- | --- | --- |
| P3 | Centralize replica coordinator and classify outcomes; separate clocks and completion scopes | P0 D-02/03, P2 | Additive state/schema; initially observe alongside existing path | State comparison, deterministic concurrency and upgrade tests |
| P4 | Cover files/search/aggregates and all record routes with authoritative eligibility | P0 D-04/07/10/14/15, P3 | Access/API behavior; possibly file/search configuration | Route inventory closed; no bypass through old URLs/index paths |
| P5 | Fixed snapshot provider/adapter and capability handling for mixed peer versions | P0 D-02/10/12, P2/P3 | New adapter/provider capability; optional versioned wire additions | Completeness and pagination interoperability tests |
| P6 | Durable effects, inbox processing and delivery jobs with retry/idempotency | P0 D-06/08, P2/P3 | Additive tables/workers and share effects | Restart/crash/retry stories; idempotent real receipt |
| P7 | Enforce strict replica read/age policy, perform controlled backfill and cleanup | P3/P4/P5; P6 where required for durable effects | User-visible availability and migrations | All critical upgrade/read tests and observable coverage |
| P8 | Negotiated withdrawal extension and publication lifecycle improvements | P0 D-05/11, P6/P7 | New interoperability feature | Capability negotiation + legacy fixtures; explicit limitations |

For P3, the initial deployment model is one Wanderer process on SQLite. Specify coalescing between goroutines, commit fencing and restart recovery before considering lease machinery. Expiring multi-owner claims are optional only if an actual worker/deployment design requires them. No multi-process scheduler or distributed locking layer is a prerequisite.

P4 is a major product and infrastructure choice: the existing SvelteKit search proxy would need eligibility checks covering hits, facets/counts, multi-search and map results. Browser-readable tenant tokens also require checking whether a reachable Meilisearch endpoint permits bypassing that proxy. D-07 must select and cost compatible enforcement and index-lag behavior; creating a new proxy from scratch is not a prerequisite because the routing layer already exists. Likewise, P8's withdrawal extension benefits capable Wanderer peers; it cannot establish general revocation on legacy or arbitrary ActivityPub servers. These tradeoffs justify separate commitment rather than hiding both inside a refactor.

If the later program is accepted, P4 and P5 can proceed in parallel after their dependencies. P6 can be developed independently of optional withdrawal signaling. P7 requires durable invalidation/index intent handling, whether supplied by P6 or a smaller P3-specific worker. Do not enable strict guarantees while a required output path is still unguarded.

## 4. Suggested PR boundaries

Only the first four items belong to the recommended P0–P2 scope. Each item is a proposed PR topic, not an already created PR. Split further when review size warrants it.

1. **P0 — Bounded contract:** D-01/D-02/D-03 status, current access/import/output behavior, explicitly deferred changes, and corrected user documentation where current behavior is verified. No implementation claims before corresponding features ship.
2. **P1 — Focused regression foundation:** reuse real-schema fixtures and current hook/router tests for chapter 05's bounded selection. Assert affected hook registration where needed; retain the existing deterministic transport/response-isolation seam. No general scheduler, whole-app fixture framework or two-instance harness is required.
3. **P2 — Allowlisted import:** explicit trail/list decoders and local-field ownership, with protected-field and newly added schema-field tests. Preserve public API and current supported payload/synchronization behavior; record absent relation inputs without silently changing reconciliation semantics. Classify any newly rejected legitimate payload as a compatibility change.
4. **P2 — Response projection:** shared checked-output adapter, explicit working copies, tests for preloaded/nested expansions and response isolation through actual routes. Retain the PocketBase enforcement path and current response contract.

The remaining items are options after a separate maintainer decision:

5. **Replica state and coordinator:** additive schema, transition service, generation checks, restart recovery, typed failures and side-by-side diagnostics. Add leases only when the chosen worker model needs them. No aggressive backfill or mass fetching in a migration.
6. **Fixed snapshot and media readiness:** source capability/provider, legacy adapters, completeness and pagination, media manifests/version checks. Comments/logs remain separately paginated.
7. **Output enforcement:** files and thumbnails, Records API/custom routes, search/facets/feed/statistics. Separate PRs per surface are appropriate, but strict mode waits for the complete set.
8. **Durable publication/delivery:** transactional intents, restartable workers, receipt/effect deduplication, recipient retries and inspection.
9. **Sharing command/lifecycle:** accepted D-06/D-16 behavior, server-side authoritative member enumeration, no accidental partial success, recipient-change and Undo semantics.
10. **Strict activation and operator tooling:** conservative backfill, workload budgets, index rebuild, availability metrics and safe restore procedure.
11. **Optional withdrawal extension:** wire namespace/capabilities, owned-object signaling, old-peer fallback and republish semantics. Keep this separately reviewable from ordinary ActivityPub compatibility.

**PLAN-002 — Separate refactors from policy changes.** A PR described as behavior-preserving must not silently change anonymous status codes, token write rights, child-author rights, remote share permissions, membership-ID exposure, stale-cache age or file access. Such changes need an accepted decision, release note and acceptance scenario.

**PLAN-003 — Preserve provenance and authorship.** Maintain separate commit history for existing contributions and new extraction/refactor work where practical. Do not treat this local proposal as permission to rewrite published branches or combine unrelated security changes into a single release commit.

## 5. Migration design for separately accepted later packages

P0–P2 need no replica-state migration or backfill. The following requirements apply only when a selected later package introduces those changes.

### 5.1 Pre-migration inventory

Before writing an executable migration, collect aggregate counts locally:

- Local originals versus remote records by canonical identity and actor origin.
- Records with conflicting/missing identity, old flag combinations and media presence.
- Known usable snapshots versus missing independent public verification.
- Existing private-to-remote shares, remote `edit` values and orphaned imported records.
- File delivery paths, direct search access and external cache configuration.
- Expected fetch backlog by origin; available worker/storage capacity.

Use counts and bounded samples for review; do not export private payloads or tokens. Inventory is not authorization to delete questionable rows.

### 5.2 Additive schema and local backfill

**PLAN-004 — No network in migrations.** Add state/job/intent structures and initialize local facts using bounded resumable batches. Keep verification unknown when evidence is missing. Schedule optional validation after the migration commits.

Keep compatibility flags while old paths are being removed, but define their single derivation from the new state. Never maintain two independent writers that can disagree about serving authority. Enforce uniqueness constraints only after a documented duplicate-resolution step.

Unknown or conflicting identities are publicly suppressed and queued for investigation; do not invent an origin or actor. Preserve local originals, their grants and user data during replica backfill.

### 5.3 Observe before strict activation

Run the proposed decision in observation mode alongside existing reads without widening current permissions. Measure cases that would become unverified, incomplete, expired or withheld. Observation mode is not fulfillment of the new security/availability guarantee.

Recommended default: activate strict mode only when the required gates are installed and the operator has reviewed coverage and impact. Unverified legacy objects then remain hidden until revalidated. An exceptional compatibility period needs separate D-09 approval, a deadline and an explicit statement of the weaker policy; it must not fabricate verification times.

### 5.4 Bounded revalidation

Prioritize recently requested public objects and subscribed feeds. Use per-origin fairness, coalescing and D-12 limits. Do not fetch every historic replica at process startup. Persist progress and backoff so restarts cannot create a retry storm.

An unreachable origin is an expected outcome. Show the operator the number of affected records and leave them unverified/expired as appropriate. Never mark failure as success to make migration progress reach 100%.

### 5.5 Cleanup and restore

After eligibility is authoritative, separately propose removal of orphaned cached bytes and confirmed deleted content under D-11. Retain minimal tombstone/fence information. Shared blobs require reference checks; removing one list member must not delete a trail used elsewhere.

Backups predating tombstones or withholding events cannot safely restore public serving by themselves. Restore with remote serving disabled and first reconcile a durable suppression ledger newer than the backup, preserving all accepted terminal deletions. Then revalidate potentially restorable identities before enabling access. A positive origin response cannot replace a lost terminal tombstone: if newer suppression history cannot be recovered, keep affected remote serving disabled and report that the strict restore guarantee cannot yet be established. The implementation must therefore define how this ledger survives ordinary database restore. Document that deleted bytes and already sent messages are not undone by schema rollback.

## 6. Release and operator communication

For P0–P2, describe the import/output boundary changes and the compatibility evidence. Do not announce new offline, revocation, retry or search guarantees. The following release requirements apply only to later packages that actually change the stated behavior.

**PLAN-005 — Describe consequences, not internal flags.** Relevant release notes should explain:

- Which formerly visible stale remote items may become temporarily unavailable.
- The distinction between cached metadata and a downloaded route file.
- What happens after origin refusal, deletion, prolonged outage or republishing.
- Which existing URLs/API status codes change, and how legacy clients behave.
- Whether private-to-remote shares or ignored remote edit values need explicit repair.
- The limits of withdrawal for older peers, copies already downloaded and backups.
- New worker configuration, queue inspection, disk limits and restore procedure.

**PLAN-006 — Version claims follow the accepted test scope.** P0–P2 require the bounded tests in chapter 05 section 2, not the full catalog. Claim complete files/search/aggregate coverage only after the corresponding later route inventory and acceptance layers pass. Do not describe a new ActivityPub extension as generally interoperable because both test peers run the same new branch.

## 7. Observability and launch gates for later behavior changes

These signals are introduced with the feature that needs them. Building this dashboard/worker program is not a P0–P2 deliverable.

| Signal | Useful breakdown | Gate or response |
| --- | --- | --- |
| Replica inventory | unverified/public/withheld/deleted, materialization, age class | Review backfill impact before strict activation |
| Read outcomes | served fresh/stale, denied, incomplete, expired, unsupported | Catch accidental changes across entry paths |
| Fetch work | queued age, per-origin rate, retry class, recovered abandoned work; lease expiry only if leases exist | Ensure bounded load and forward progress |
| Generation rejection | obsolete snapshot/media/index/delivery work | Confirm stale work is discarded; investigate abnormal rates |
| Projection lag | committed intent age, failing index operations | Fail output safely while restoring indexing |
| Delivery | accepted/retrying/exhausted/permanent rejection, queue age | Separate local success from remote receipt |
| Cache storage | active/unreferenced/withheld/deleted bytes | Enforce accepted retention without deleting local originals |

Before enabling an accepted later behavior phase:

1. Record accepted decisions, intended endpoints and supported peer versions.
2. Run the relevant acceptance scenarios against the exact migration/head combination.
3. Confirm queued work survives restart and remains within configured budgets.
4. Check that old code paths cannot bypass the new authority.
5. Prepare operator diagnostics and a rollback procedure that preserves suppression.

## 8. Non-goals and further deferred topics

- Private cross-instance collaboration, including authorization to edit the origin.
- Universal recall of data from arbitrary third-party servers.
- Replacing PocketBase, ActivityPub or Meilisearch as a prerequisite.
- Multi-process workers or distributed coordination without a demonstrated deployment need.
- A guaranteed full offline archive of every discussion and optional attachment.
- Unrelated profile, moderation or plugin redesign.

These topics may deserve separate proposals. Their absence must not be filled by accidental behavior in the shared public replica path.

## 9. Definition of the proposal being ready

This document set is ready for maintainer discussion when the known baseline, normative targets, pending decisions and acceptance scenarios are internally consistent. It is ready for implementation of a particular package only after that package's behavior decisions and compatibility requirements are settled.

The next concrete discussion is D-01/D-02/D-03 and approval of the P0–P2 boundary. After P2, assess whether the reduced import/output risk and resulting code are sufficient for the next release. The output, state, delivery and withdrawal packages remain separate choices; the project may stop after P2 without leaving an agreed architecture rollout unfinished.
