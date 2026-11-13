# Review — S16-19 Summary composers read the place entries and keep what readers re-derive

**Verdict:** round 1: Fits with changes — Not ready; round 2 pending
**Rounds:** 1 so far. **Reviewers:** maintainability (Opus), executability (Sonnet)

## To user

- Answered 2026-11-13. **A-33's shape (M-1):** one participation type, `EventParticipant{ID, Roles}` on the event row from one `Follow`, with no participant node reads; `EventSubject` stays the title's name summary. The map's "one shape" line is reworded (D-7).
- Answered 2026-11-13. **The event row's date Property:** dropped from scope; no reader once the person row keeps its date owners (D-6).
- Answered 2026-11-13. **Life owners on a crafted over-cap catalog:** the stored row's owners may go stale as its places already may; accepted, as S14-40's miss (D-11).

## Writer's own calls

- `FieldOwner` holds ids (`ID`), not entities, so a Persons list fill loads no association node; `PersonDerivedRows` builds entities from the nodes it already loads (D-4, D-5).
- The Timeline labels the stored roles itself (`TermLabeler`) into new slices; the row keeps them unfinished (D-10).
- `BenchmarkCacheSizes`' `displays` measure is fixed to fill stored rows, compared before and after only (D-12).
- Stop thresholds: search inputs 1.5 ×, times 1.25 ×, `displays-B` +20 MB at 16,000 persons (§11).

## Round 1

### M-1 · blocker · The one full `EventParticipant` carries fields neither use reads and widens the event row's inputs
Participant names have no reader (the Timeline uses person rows, `personEventProto`); subject roles have no reader (`eventSummaryProto`, `eventTitle`); reading every participant's node records `NodeKey(person)` on every event row and persists a `catalog_search_inputs` row per participant; P-8 names ids and roles. Better: `EventParticipant{ID, Roles}` from one `Follow`, `EventSubject` stays.
**Answer:** To user → answered (user, 2026-11-13: one participation type). D-7 rewritten; D-8's relabel on every read goes (D-10 labels in the Timeline); map and design reworded.

### M-2 · blocker · `Subjects` rebuilt the hop's term filter in the composer
**Answer:** Fixed — D-7 keeps `follow(in, n.ID, hops.SubjectsOfEvent)`; the Edges assumption is gone.

### M-3 · major · A fourth caller of a duplicated node-to-entity helper (three copies)
**Answer:** Fixed — D-3 adds `canonicalentities.FromNode`; both private `entityOf`s and the inline copy in `source_read.go` go (STEP-3).

### M-4 · major · `addChains` follows the locations `lifeFacts` just read
**Answer:** Fixed — D-4: `lifeFacts` returns the seeds it followed; `composePersons` calls `rec.addClosure` as `composeEvents` does; `addChains` goes.

### M-5 · major · Scale left out the persisted search inputs and the reindex path
**Answer:** Fixed — §6 Scale, STEP-1, STEP-8 and §11 measure the `catalog_search_inputs` count, a reproject of every handle and a name edit on the busiest participant, with stops. M-1's answer removes most of the growth.

### M-6 · major · D-6 said the map and design were amended in the planning commit; they were not
**Answer:** Fixed — the map row, the scope entry and `cache-architecture.md` are amended in this commit, with the user's answers.

### M-7 · minor · `conclusionsummaries/participants.go` named like `persondetail/participants.go`
**Answer:** Fixed — D-9 puts `RolesByNeighbor` in `events.go`, its tests in `conclusionsummaries_test.go`.

### M-8 · minor · The `known` closure repeated `TermLabeler`'s check
**Answer:** Fixed — D-8 `vocab.term`, which `TermLabeler` calls.

### X-1 · blocker · STEP-2 deleted `eventPlaceGraph` while `PersonLifeOwners` still called it
**Answer:** Fixed — D-2: `PersonLifeOwners` builds its graph inline until STEP-4; `lifeOwners` keeps its association reads until then.

### X-2 · blocker · The STEP-5 check could never pass (`engine.EventSubject`)
**Answer:** Fixed — moot: `EventSubject` stays (M-1). The DoD sweep is one `rg` over names that leave the code entirely.

### X-3 · major · The planning commit did not make D-6's amendments
**Answer:** Fixed — same as M-6.

### X-4 · major · AC-2.3's harness and the over-cap staleness were unclear
**Answer:** Fixed — AC-2.3 names `newFixture`, `f.lifePerson` / `f.lifeEvent` / `f.placeSubject` and `f.Cite`, whose write reports `SummaryDeps` and drops the stored row; §4 "Stored rows" says how. The over-cap case is D-11 (user's answer); §11's risk row points at it.

### X-5 · minor · `lifeFacts` sorted the caller's slice
**Answer:** Fixed — D-4 sorts a copy and names the key func. (Recorded keys are sorted by `InputScope.Keys`, so the order never moved them.)

### X-6 · minor · Choices left open
**Answer:** Fixed — T-1/T-2 in `walks_test.go`; `keep` is a literal in `oneEvent` (D-7); the moved tests read `.ID` / `.Roles` (STEP-5); T-8 now checks the stored row through `EventsByIDs` (AC-3.7), no display assertion; T-4's ids compare with `FieldOwner.ID` (STEP-4); the ship log path and `hostname` (STEP-1); `displays-ms` taken as `nodes-ms` is, compared before and after only (D-12).

### X-7 · minor · Base reference ambiguous
**Answer:** Fixed — the header and §4 say `origin/main` @ `7455758`.

### X-8 · nit · Smaller ambiguities
**Answer:** Fixed — D-5 guards a nil catalog and graph; D-10 labels roles through `label(r.ID, r.Key)` (`TermLabeler`); AC-3.4 names the full key `{event, "participation", hops.ParticipantsOfEvent.From()}`.
