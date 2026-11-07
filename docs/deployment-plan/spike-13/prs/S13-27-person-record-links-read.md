# S13-27 — Person record links read

**Status:** Planned
**Spike / slice:** Spike 13 · Slice 7 — How a person's records connect
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no chrome change: an engine read, an RPC, a store method and a cache key; no view reads them until S13-28)
**Board:** —
**Depends on:** S13-01 (landed [#655](https://github.com/mendahu/provenencia/pull/655)), S13-26 (landed [#736](https://github.com/mendahu/provenencia/pull/736)), S13-29 (landed [#741](https://github.com/mendahu/provenencia/pull/741))
**Blocks:** S13-28
**Branch:** `s13-27-person-record-links-read`
**Runs on:** macOS: STEP-6 needs `protoc` and its plugins, and STEP-8, STEP-9 and `scripts/check-pr.sh` need Xcode
**Version bump:** none — Spike 11–14 PRs leave `VERSION` alone until a release is cut, as S13-01, S13-04 and S13-05 did ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-match-rule`](../../../../.agents/skills/add-match-rule/SKILL.md), [`add-catalog-query`](../../../../.agents/skills/add-catalog-query/SKILL.md), [`add-ffi-handler`](../../../../.agents/skills/add-ffi-handler/SKILL.md), [`use-catalog-session`](../../../../.agents/skills/use-catalog-session/SKILL.md), [`add-workspace-place`](../../../../.agents/skills/add-workspace-place/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md)
**Rules that bind:** `.cursor/rules/go-tests.mdc`
**Context:** [`evidence-graph-comparison.md`](../evidence-graph-comparison.md) (Requirement 1, § First version, § Scale), working rules: "Links between Evidence graphs are computed, never stored", "Scoring and the view are two concerns", "Slice 7 starts with a simple count", "Each section reads its own query", "One invalidation rule for person-section keys", "Canonical only, filtered in Go", "Engine PRs run ahead of their slice"; [`docs/ideas/structural-debt.md`](../../../ideas/structural-debt.md) rows A-34 and S13-16 PRF-2
**Planned:** 2026-11-07 against `main` @ `26ada55f`
**Review:** [`S13-27-person-record-links-read.review.md`](S13-27-person-record-links-read.review.md) — round 1 Fits with changes / Not ready, round 2 Fits with changes / Ready with changes; every finding Fixed but M-2 and M-6 (declined); none open

### Implementer — read this first

1. **This file is the whole task.** Read it top to bottom before you change code. Then read every file under **Current state** and load every skill under **Skills to load**.
2. **Check the base.** Compare `main` with the **Planned** commit. If a file under **Current state** has changed so a row is no longer true, stop and report which rows.
3. **Decisions are made.** Do not reopen a **Decision** or a **Scope** line, and do not swap in a design you like better. If one cannot work, stop: say which ID it is, what you found (path, line, output), and what you would need. Do not improvise around it.
4. **Do only In.** Out stays out, even when it is one line away. Do not refactor, rename, or tidy code the plan does not name.
5. **Work the Steps in order.** Run each step's check before you start the next. A failing check is fixed in that step.
6. **Hit a Stop condition → stop.** Report it. Do not work around it.
7. **Done is the Definition of done.** Report each box with its evidence: the command and its result, the test name, or the manual check you ran.
8. **Docs-only lines in the DoD are part of the PR.** Do not skip them.

---

## 1. Objective

The Person page will show how a person's records connect (S13-28): each Evidence graph that holds an accepted member of the person, and between two graphs one thin or thick line for each other person both name. This PR is the read behind that view. `GetPersonRecordLinks` returns the person's Evidence graphs, in the order the Sources section lists them, and for each pair of graphs the lines S13-26's `core/recordlinks` counts (the shared person's entity and the weight). Nothing is stored; the read is computed each time from the graph cache and the shared Evidence layer (S13-29).

**Why now.** S13-28 draws only this read's output (working rule "Scoring and the view are two concerns"). S13-26 left the read its own bound to set: `recordlinks.MaxPairWork` caps one pair of graphs, so a person in 50 crafted graphs is 1,225 pairs at the cap, about 12 minutes by arithmetic (S13-26 ship log, round 3).

**What stays.** One pair is counted and weighed as S13-26 shipped it. Promote, the claims, family and timeline reads, and every Swift view are unchanged. The Evidence graph cache keeps its cap of 64 graphs (`maxEvidenceGraphs`).

## 2. User stories

**US-1.** As a researcher, when I open a person, the app can read which of their records name the same other people, and how much the records agree about each, so that S13-28 can draw a thin or thick line between two records for each shared person.

AC-1.1's records. Each graph is the sole graph of its own Source; every Subject is cited only with what the table says; no `relationship` is cited.

| Graph (Source title) | Subjects and links | Promoted onto |
| --- | --- | --- |
| A ("A register") | son "Thomas Smith"; father "William Smith"; father's birth (`Birth`) with the father's `participation` "subject"; mother "Mary Ann Jones" | son → S (new), father → F (new), birth → E (new), mother → M (new) |
| B ("B register") | son "Thomas Smith"; father "William Smith"; father's birth with the father's `participation` "subject" | son → S, father → F, birth → E |
| C ("C register") | son "Thomas Smith"; mother "Mary Jones" | son → S, mother → M |
| D ("D register") | son "Thomas Smith" | son → S |

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | The records above; read S | `graphs` lists A, B, C, D. `pairs` holds exactly two: A–B with one line {F, 2 facts, strong} (the name agrees; the birth participations correspond, far ends both on E); A–C with one line {M, 1 fact, weak} (the name partly agrees: "Ann" is left over). D is in no pair |
| AC-1.2 | AC-1.1's records | No line names S, though every graph holds the son |
| AC-1.3 | Add graph E1 (Source "E register"): son → S and a father "William Smith" with a provisional claim on F; and graph E2 (Source "F register"): son → S and a father "William Smith" never promoted | No pair A–E1, B–E1, A–E2 or B–E2: a member with no accepted claim has no entity and draws no line |
| AC-1.4 | Sources "Census 1851", "Baptisms", "Census 1841", each with the son promoted onto S; "Baptisms" holds two graphs, its sole graph renamed "Register" and a second one "Index" (`evidencegraphstest.Create`), the son on each | `graphs` order: "Baptisms" / "Index", "Baptisms" / "Register", "Census 1841", "Census 1851": Source title, then graph name, each with its ref as the tie-break — the order `persondetail.Claims` gives |
| AC-1.5 | One graph holds two Subjects both promoted onto S | The graph is listed once |
| AC-1.6 | The id of an Event handle (E); a random 16-byte id; a 3-byte id | `conclusiondetails.ErrNotFound` for each |
| AC-1.7 | A person with one accepted member; a person whose two graphs name no one else | One graph and no pair; two graphs and no pair |

**US-2.** As a researcher with a well-documented person, the record-links read stays fast at the spike's Scale, and a crafted catalog cannot hang it.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | A person in 50 household graphs (D-13: 1,225 pairs, three shared people per pair) | Cold read (graph cache dropped, then `catalogsession.WarmOnOpen`) within 1.5 s; repeat read within 300 ms; both times logged |
| AC-2.2 | The same households at 65 graphs, past `maxEvidenceGraphs` (64) | The read equals the read after `Graph().Drop()`, and its repeat time is within the 1.5 s cold budget, logged |
| AC-2.3 | A read whose work passes `recordlinks.MaxReadWork` but whose number of lines does not | Every line is listed with 0 facts and weak; no property or link is compared |
| AC-2.4 | A read whose number of lines alone passes `recordlinks.MaxReadWork` | No pair is returned; every graph is still listed |
| AC-2.5 | A read within `MaxReadWork` | Each pair's lines equal the one-pair count over the same two graphs, a pair over `MaxPairWork` included (its lines at 0 facts, weak) |
| AC-2.6 | A warm repeat read | Issues no graph-cache query |

**US-3.** As the app, the record-links key reloads whenever what it shows may have changed, and not otherwise.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | `.personRecordLinks(e1)` warmed, its value holding graphs `g1`, `g2`; a write names Evidence graph `g1` (promoting or unpromoting the father there, a relationship added, the graph renamed, its Source retitled) | The key revalidates |
| AC-3.2 | The same key; a write touches handle `e1` (the person promoted from a new graph, or unpromoted from one) | The key revalidates |
| AC-3.3 | The same key; a write touches handle `e7` and names only graph `g9` (the father promoted in a record the person is not in) | The key does not revalidate |
| AC-3.4 | A property-vocabulary write | The key revalidates |
| AC-3.5 | The Person page is off screen | The key is evicted, not revalidated (`whenHidden: .evict`) |
| AC-3.6 | `FakeStore` seeded with a value for `e1` | The key loads it; an unseeded handle throws `conclusiondetails.not_found` |

## 3. Scope

| In | Covered by |
| --- | --- |
| `recordlinks.Pairs` and `recordlinks.Pair` *(new)*: every pair of a read's graphs, its lines, one read-wide bound; the one-pair `Lines` becomes unexported | STEP-1, AC-2.3 – AC-2.5 |
| `recordlinks.Rules.MaxReadWork` *(new)* knob, its registry value and inventory row | STEP-1, D-3 |
| `persondetail.RecordLinks` *(new)*; one Person check shared with `PersonEvents` | STEP-3, AC-1.1 – AC-1.7 |
| `persondetail` row in `matchregistry`'s `readerPackages` guard (S13-29 D-11) | STEP-3 |
| Budget test at 50 and 65 graphs | STEP-4, AC-2.1, AC-2.2, AC-2.6 |
| Shared timing helpers `databasetest.Fastest`, `Since`, `WithinBudget` *(new)*, moved out of `promotealign/perf_test.go` | STEP-4, D-14 |
| `evidencelayertest.Open` *(new)*: Records on an existing catalog | STEP-2, D-12 |
| `GetPersonRecordLinks` RPC, method 116 | STEP-6, STEP-7, AC-1.1, AC-1.6 |
| Swift store method, types, `FakeStore` seed | STEP-8, AC-3.6 |
| `.personRecordLinks` key; two descriptor builders for the per-person keys and the derived kind count (structural-debt A-34) | STEP-9, AC-3.1 – AC-3.5 |
| Docs and skills: map, `completed.md`, `matching-model.md`, `evidence-graph-comparison.md` § Scale, `add-match-rule`, `add-workspace-place`, structural-debt A-34 row removed | STEP-10 |

| Out | Home |
| --- | --- |
| The circular link diagram, warming the key from the Sources section, any chrome | S13-28 |
| Facts per line, conflicts, or what a line counted on the wire | not planned in Spike 13: [`evidence-graph-comparison.md` § Open](../evidence-graph-comparison.md#open) |
| Event and Place record links | [`evidence-graph-comparison.md` § Later](../evidence-graph-comparison.md#later) |
| Raising `maxEvidenceGraphs`; a batch Evidence graph read past 64 graphs; a multi-graph layer load | [`docs/ideas/structural-debt.md`](../../../ideas/structural-debt.md) row S13-16 PRF-2 (already there) |
| Memoizing the read's result in Go | not planned: the repeat read is within budget (AC-2.1) and the Swift key caches it |
| Following a merged handle's `merged_into_id` for a member's entity | not planned: the app writes no merge (`merged_into_id` is set only by `graphcachetest`) |
| Refreshing on a name-alias or default-name-format write | not planned here: those tables report no effect (`core/database/effects/registry.go` `"name_alias_tables"`, `"project_settings"`: `None: true`), so no reconciled key refreshes on them today; the next page open reads the new rules |
| A `graphcache.EvidenceGraph` lookup of a Subject's entity shared with `promotealign.mergeFixed` | not planned: `RecordLinks` builds a map once per graph, `mergeFixed` walks the members in order; no shared access pattern |
| Changing `PairWork`'s doc or the one-pair count's behavior | not planned |

## 4. Current state

Verified at `26ada55f`.

| Area | Where | Today |
| --- | --- | --- |
| Pair comparison | `core/recordlinks/recordlinks.go` `Lines`, `NewGraph`, `Graph`, `Member`, `Line`, `Rules`, `Config` | `Lines(page, a, b, metas, cfg)` returns one `Line{Entity, Facts, Strength}` per person entity both graphs hold, the page excluded, sorted by entity. It loops `slices.Sorted(maps.Keys(a.persons))`. `overWork` → `pairWork(page, a, b, ha, hb, cfg, limit)` counts the pair's work before any compare and stops past `limit`; over `MaxPairWork` every line has 0 facts. `holdings` (by member id) caches each member's `tally` once per call. No production caller outside the package (`rg "recordlinks\."` finds only `matchregistry`) |
| Test exports | `core/recordlinks/export_test.go` `PairWork` | Exposes `pairWork` with no limit, for the cap's measurement |
| Record-links tests | `core/recordlinks/recordlinks_test.go` `TestLines`, `TestLinesBounds`, `TestLinesWorkCap`, `TestLinesValueSize`, `BenchmarkLinesAtCap`; helpers `person`, `event`, `part`, `rel`, `build`, `lines`, `idA`, `idB`, `ent`, `solo`, `pair`, `axes`, `underCap`, `pt` (property test values); `core/recordlinks/invariants_test.go` `TestLinesInvariants` (`genGraph`, `materialize`); `core/recordlinks/fuzz_test.go` `FuzzLines` | The pair rules, bounds and invariants (symmetry, input order). `go test` runs `FuzzLines`' seed corpus |
| Registry section | `core/matchregistry/recordlinks.go` `defaultRecordLinks`; `core/matchregistry/registry.go` `RecordLinksConfig` | Five knobs, each with a `// decided:` comment |
| Knob inventory | `docs/matching-model.md` line 33 (paragraph), rows 216 – 220; `core/matchregistry/inventory_test.go` `TestKnobInventoryMatchesRegistry` | Every `recordlinks.*` knob has a row; the test fails on a knob with none |
| Skill text | `.agents/skills/add-match-rule/SKILL.md` line 34 | "Its five knobs are …", `Lines` named as the entry |
| Import guard | `core/matchregistry/guard_test.go` `readerPackages`, `TestReaderPackages` | Rows for `core/database/evidencelayer` and `core/recordlinks`; S13-29 D-11 says S13-27 adds `persondetail` |
| Evidence layer | `core/database/evidencelayer/evidencelayer.go` `Load(g, graphID, namePattern) (Layer, error)`, `Layer{Subjects []subjectcompare.Subject; Links []kinship.Link}`, `ErrNoGraph` | Memoized per graph under `memoKey + namePattern`; `Subjects` are the graph's primary Subjects with values; `Links` recorded and implied |
| Value rules | `core/database/valuerules/valuerules.go` `Load(g, q) (valuecompare.Rules, string, error)`, `Metas(g)`; `core/database/promotealign/promotealign.go` `loadConfig` | The catalog's value rules and name pattern key; `loadConfig` sets `cfg.Value = rules` on a registry config |
| Claims read | `core/database/persondetail/claims.go` `Claims(c, entityID) ([]ClaimGroup, error)`, `memberGraphs` | `conclusiondetails.Handle`, then groups the person's accepted claims by Source title, claims by graph name (refs as tie-breaks); each `Claim.Graph` is the loaded `graphcache.EvidenceGraph`. Skips a member whose graph is not found or whose Subject is gone. Checks no kind |
| Claims test fixture | `core/database/persondetail/claims_test.go` `fixture` (embeds `*evidencelayertest.Records`), `newFixture`, `source`, `person(graphID, label)`, `cite`, `promote(s, onto)`, `provisional(s, e)`, `burialsAndBaptisms` (renames a sole graph with `evidencegraphs.Update`, adds graphs with `evidencegraphstest.Create`) | Builds Subjects on any graph of a Source, and provisional claims |
| Timeline read | `core/database/persondetail/timeline.go` `PersonEvents` lines 90 – 105 | `conclusiondetails.Handle`, `g.Vocabulary()`, `v.TypeIDsByKey()`, `!n.Is(personType)` → `conclusiondetails.ErrNotFound`; it also reads `eventType` from the same map |
| Evidence graph | `core/database/graphcache/evidence_graph.go` `EvidenceGraph` (`GraphID`, `GraphRef`, `GraphName`, `SourceID`, `SourceRef`, `SourceTitle`, `SourceTypeID`, `Members`), `EvidenceMember{SubjectID, ClaimID, EntityID, Kind}`, `maxEvidenceGraphs` = 64 | `Members` are the graph's accepted claims (`graphfill.loadGraphMembers`: `status = 'accepted'`), each with its entity. Past 64 loaded graphs the least recently used is evicted with its memos |
| Test records | `core/database/evidencelayer/evidencelayertest/evidencelayertest.go` `New`, `Records`, `Person`, `Birth`, `Participation`, `NewSource`, `Graph`, `Layer`, `UserID` | `New` creates a catalog in a temp dir, registers its close, upserts `UserID`, installs the subject vocabulary and creates Source "Register" with Artifact "Scan". Every builder places Subjects on the Source's sole graph. No way to build on a catalog that already exists |
| Promote in tests | `core/database/promote/promotetest` `Save(t, c, userID, promote.Input{SubjectID, EntityID})` | Files an accepted claim on a new entity, or on `EntityID` |
| Budgets | `core/database/promotealign/perf_test.go` `fastest`, `since` (lines 127 – 145), the `over` closure in `TestLargeSourceBudgets` | Best of n runs; skip the check under `graphcache.VerifyMode()`; log each time |
| Test helper package | `core/database/databasetest/selects.go` package doc | "holds test helpers that watch a catalog's SQLite connection" |
| Wire | `api/proto/engine.proto` `Method` (last: `METHOD_GET_UNPROMOTE_IMPACT = 115`), `ListPersonEventsRequest`, `PersonClaimGroup`, `PersonClaim` | No record-links message |
| Handlers | `api/ffi/handlers/person_detail.go` `ListPersonEvents`, `kinshipPatternProto`; `api/ffi/handlers/person_detail_test.go` `TestListPersonEvents` (an `onboarding.Complete` project and `catalogsession.Do`), `TestKinshipPatternProto`; `api/ffi/dispatch.go` lines 106, 320 | Thin handlers through `withProjectCatalog`; ids go out through `database.UUIDString` |
| Swift store | `macos/App/Platform/GenealogyStore.swift` line 1544 `listPersonEvents`; `GoStore.swift` line 1236; `FakeStore.swift` line 1712 (`seededPersonEvents`); `CoreInvoke.swift` line 142; `macos/ProvenenciaTests/OnboardingModelTests.swift` line 617 (`ThrowingStore`) | One method per person read, in each conformer. `GoStore` drops an unspecified enum value with `compactMap` (`mapFamilyUnit`) |
| Swift key | `macos/App/Features/Workspace/Session/CatalogQueryKey.swift` `.personEvents` (line 40), `Kind`, `kind`, `project`; `CatalogQueryRegistry.swift` lines 532 – 600 (`.personClaims`, `.personFamily`, `.personEvents`), `CatalogQueryDescriptor.make` (line 52) | `.personFamily` and `.personEvents` are one descriptor but for `load`: `.sessionFresh`, `.evict`, refreshes `[.applyHiddenPolicyOnAny(follows: .touchedHandles), .followsPropertyLabels]` (A-34). `.personClaims`: `.sessionFresh`, `.evict`, refreshes `[.applyHiddenPolicy(follows: .touchedHandles) { [entityId] }, .applyHiddenPolicy(followsValue: .evidenceGraphs) { graph ids }, .followsPropertyLabels]` |
| Swift tests | `macos/ProvenenciaTests/CatalogQueryRegistryTests.swift` `everyKindHasOneDescriptor` (`#expect(kinds.count == 21)`, line 451), `warmedVocabularySession`, `familyCases`, `noMutationReachesAConclusionKey`, `personEventsRevalidateOnTouchedHandle`, `personEventsLoadThroughTheStore`; `WorkspaceSessionTests.swift` `personClaimsFollowEveryGraphUntilLoaded` | The per-key tests S13-01 and S13-05 added |
| Workspace skill | `.agents/skills/add-workspace-place/SKILL.md` line 63 (Refresh row) | Teaches per-key `refreshes`; no person-section builder |
| Effects | `core/database/effects/registry.go` `"identity_claims"` (`EvidenceGraph: claimGraph`, line 443), `"sources"` (line 311), `"subjects"`, `"observations"` | A claim, Subject, Observation, or Source title or type write names the Evidence graphs it lands on |

**Precedent.** S13-05 ([#680](https://github.com/mendahu/provenencia/pull/680), plan [`S13-05-person-timeline-read.md`](S13-05-person-timeline-read.md)) added a person-section read of the same shape: `persondetail` function, RPC, store method, `FakeStore` seed, key, registry and session tests. Copy its STEP-5 to STEP-8 and the test `personEventsLoadThroughTheStore`. S13-01's `.personClaims` descriptor is the refresh shape to copy, and `personClaimsFollowEveryGraphUntilLoaded` its session test. S13-26 ([#736](https://github.com/mendahu/provenencia/pull/736), ship log F-8) added `MaxPairWork`: copy how its value was set from measurement and recorded (`// decided:` comment, inventory row, benchmark). Do not copy S13-05's `invalidation:` / `invalidateOn:` arguments: its landed descriptor has none.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | Layers. `core/recordlinks` (pure) owns every rule about lines: the pair loop over a read's graphs and the read-wide bound (`Pairs`, `MaxReadWork`). `core/database/persondetail/recordlinks.go` *(new)* loads graphs, layers, rules and metas, builds each `recordlinks.Graph`, and calls `Pairs`. The handler only maps | Working rules "Scoring and the view are two concerns" and "Each section reads its own query". The work count is `recordlinks`' own (`pairWork`) | The pair loop and bound in `persondetail` — a second place that knows what a line costs |
| D-2 | `type Pair struct { A, B int; Lines []Line }` *(new)*, doc "the lines between two graphs of a read; A < B index the graphs Pairs was given". `func Pairs(page []byte, graphs []Graph, metas subjectcompare.Metas, cfg Config) []Pair` *(new)*. Behavior: (1) an index from each person entity other than `page` to the graph indexes that hold it gives each pair's shared entities in ascending byte order, only for pairs that share one; (2) the number of lines is counted first, stopping once past `MaxReadWork`, and past it `Pairs` returns nil: for each entity, walking its holders in index order, the i-th holder (from 0) adds i, the lines it makes with the holders before it (no halving: `core/recordlinks` may hold no numeric literal but 0 and 1, `matchregistry` `TestLayerPackages`); (3) the read work is that number plus, for each pair, its pair work (as `pairWork` counts it over the pair's shared list, which already excludes the page, with limit `MaxPairWork`; an entity over `MaxEntityMembers` keeps its line and adds no work, as in `pairWork` today; each member's cost is read per pair, as `pairWork` does) capped at `MaxPairWork`, counted before any compare and stopping once past `MaxReadWork`; (4) past it, every pair is listed and every line has 0 facts and is weak; (5) otherwise each pair's lines are exactly what the one-pair count gives over `graphs[A]`, `graphs[B]` (a pair over `MaxPairWork` at 0 facts, weak). Pairs are ordered by `A`, then `B`; a pair with no line is absent. One `holdings` serves the whole call (member ids are Subject ids, unique across graphs). `pairWork` and the one-pair count take the shared entities as a sorted list; the one-pair count builds it from `a.persons` when called alone. Counting stops past the cap, so no count overflows | The index makes the work proportional to the lines, not to n² pairs or to each graph's persons; counting lines first bounds the output itself. Charging a capped pair `MaxPairWork` keeps the sum independent of order: each term is fixed before the sum, and the sum only grows. One holdings per read reads each member's tally once, not once per pair | Calling the one-pair count per pair from `persondetail` — no read bound, and each tally re-read 49 times at Scale. A `[][]Line` matrix — n² entries, mostly empty. Charging a capped pair 0 — the result then depends on counting order |
| D-3 | `Rules.MaxReadWork int` *(new)* after `MaxValueBytes`, doc "the most work one Pairs call may do before it compares nothing: every line keeps 0 facts, and when the number of lines alone passes it, no line is listed. Work is the number of lines plus each pair's work as MaxPairWork counts it, capped at MaxPairWork". Value set from measurement in STEP-1: at least 20 times `ReadWork(entP, householdGraphs(50), cfg)` (D-11, D-13), rounded up to two significant digits. Comment: `// decided: S13-27 plan D-3, the rule the user set for MaxPairWork (2026-11-07, S13-26 SEC-1): a high cap a researcher never reaches, so a crafted catalog cannot hang a read`. Inventory row after `recordlinks.MaxValueBytes` in `docs/matching-model.md`, with the 50-graph work, the value and `BenchmarkPairsAtCap`'s time | The map's Check: the read needs its own bound. One high cap on the total, set from measurement above real data at the map's Scale, counted before the work (S13-26 F-8) | A cap on the number of graphs — real people reach 50 and more, and it bounds neither lines nor words. A time deadline — the result would depend on the machine |
| D-4 | The one-pair count becomes unexported: `Lines` is renamed `pairLines`, its doc unchanged but for the name, and `export_test.go` gains `var Lines = pairLines` so every test calls `recordlinks.Lines` unchanged. The package doc names `Pairs` as the entry. `Pairs` is the only exported comparison | After this PR the only production caller goes through `Pairs`; a later reader (Event and Place pages) cannot loop the one-pair count and skip `MaxReadWork` (review M-3) | Keeping `Lines` exported — an unbounded entry beside the bounded one |
| D-5 | `persondetail.RecordLinks(c *database.Catalog, entityID []byte) (PersonRecordLinks, error)` *(new)* in `recordlinks.go` *(new)*, with `type PersonRecordLinks struct { Graphs []graphcache.EvidenceGraph; Pairs []recordlinks.Pair }` *(new)*. (1) `n, err := personNode(c, entityID)` (D-6); (2) `claims(c, n)` (D-6); the graphs are each claim's `Graph`, walking groups then claims in that order and keeping each `GraphID` once, at its first occurrence; (3) `linksBetween(c, entityID, graphs)` *(new, unexported)* returns the pairs: `db, err := c.DB()`; `rules, namePattern, err := valuerules.Load(g, db)`; `metas, err := valuerules.Metas(g)`; `cfg := matchregistry.Default().RecordLinksConfig(); cfg.Value = rules`; per graph `evidencelayer.Load(g, graph.GraphID, namePattern)`, any error returned (`ErrNoGraph` cannot occur: `claims` already skips a graph that is not found, the catalog session runs nothing else during the read, and `Load` reloads an evicted graph); each layer Subject becomes `recordlinks.Member{Subject: s, Entity: entityOf[string(s.ID)]}`, `entityOf` built once per graph from `graph.Members` (`SubjectID` → `EntityID`), nil when absent; `recordlinks.NewGraph(members, layer.Links)`; then `recordlinks.Pairs(entityID, rgraphs, metas, cfg)`. `linksBetween` takes no Person assumption, so an Event or Place page adds only its graph list (review M-7) | `claims` gives the Sources section's order and skips members dropped mid-read; reusing it keeps one order. `graph.Members` holds only accepted claims, so a provisional or unpromoted member has no entity (working rule "Canonical only") and draws no line. `valuerules.Load` gives the catalog's alias tables and name pattern, as `promotealign.loadConfig` does | Calling `memberGraphs` and sorting again — a second copy of the order. `Default().Value` — loses the catalog's alias tables and name format. Reading neighbors' handles from `g.Node` — `EvidenceMember.EntityID` already is the handle. Sharing `loadConfig` — it builds graph alignment's config, which `persondetail` must not import (D-10) |
| D-6 | In `persondetail`: `personNode(c *database.Catalog, entityID []byte) (*graphcache.Node, error)` *(new, unexported)* in `claims.go`: `conclusiondetails.Handle`, `g.Vocabulary()`, `n.Is(v.TypeIDsByKey()[vocabkey.Person])` or `conclusiondetails.ErrNotFound`. `PersonEvents` calls it and reads `eventType` from its own `g.Vocabulary()` call (cached). `Claims(c, id)` becomes `Handle` then `claims(c, n)` *(new, unexported)*, the current body after `Handle`; `Claims`' behavior does not change (it still accepts any kind) | One place decides "is this a Person page id" (review M-5); `RecordLinks` reads the handle once | A fourth copy of the check |
| D-7 | Wire (method `METHOD_GET_PERSON_RECORD_LINKS = 116`). `GetPersonRecordLinksRequest { string project_dir = 1; string entity_id = 2; }`. `GetPersonRecordLinksResponse { repeated RecordLinksGraph graphs = 1; repeated RecordLinksPair pairs = 2; }`. `RecordLinksGraph { string graph_id = 1; string graph_ref = 2; string graph_name = 3; string source_id = 4; string source_title = 5; string source_type_id = 6; }`. `RecordLinksPair { string graph_a_id = 1; string graph_b_id = 2; repeated RecordLink lines = 3; }`. `RecordLink { string entity_id = 1; RecordLinkStrength strength = 2; }`. `enum RecordLinkStrength { RECORD_LINK_STRENGTH_UNSPECIFIED = 0; RECORD_LINK_STRENGTH_WEAK = 1; RECORD_LINK_STRENGTH_STRONG = 2; }`. Every id is a UUID string (`database.UUIDString`); `graph_ref` is empty for a graph with no ref. Comments state the order (graphs as the Sources section; pairs by graph order, `graph_a_id` first; lines by entity id bytes) and that facts are not sent | The map row names these fields, and the S13-28 scope entry says "The view reads only S13-27's output" (review M-2 declined, see the log). Graph ids, not indexes, so Swift needs no index arithmetic. Facts stay off the wire: the view shows no line data | Reusing `EvidenceGraph` — its `description` and `artifact_ids` would be sent empty. Sending `facts` — invites a label the working rule refuses. Graph ids only, labels from `.personClaims` — contradicts the S13-28 scope entry |
| D-8 | Handler `GetPersonRecordLinks` *(new)* in `api/ffi/handlers/person_detail.go`, after `ListPersonEvents`; `parseID` failure → `conclusiondetails.ErrNotFound`; maps through `recordLinkStrengthProto(recordlinks.Strength) engine.RecordLinkStrength` *(new)*, with `TestRecordLinkStrengthProto` *(new)* failing on any `Strength` that maps to unspecified, as `TestKinshipPatternProto` does. Dispatch `MethodGetPersonRecordLinks` in `api/ffi/dispatch.go` | Same shape as the three person reads | — |
| D-9 | Swift types in `GenealogyStore.swift` after `CatalogPersonTimeline`: `CatalogPersonRecordLinks { graphs: [CatalogRecordLinksGraph]; pairs: [CatalogRecordLinksPair] }` (`init(graphs: [] , pairs: [])` defaults), `CatalogRecordLinksGraph { graphID, graphRef, graphName, sourceID, sourceTitle, sourceTypeID: String }` (`Identifiable`, `id` is `graphID`), `CatalogRecordLinksPair { graphAID: String; graphBID: String; lines: [CatalogRecordLink] }`, `CatalogRecordLink { entityID: String; strength: CatalogRecordLinkStrength }`, `enum CatalogRecordLinkStrength { case weak, strong }`; all `Sendable, Equatable`. Store: `getPersonRecordLinks(projectDir:entityID:) async throws -> CatalogPersonRecordLinks`. `GoStore` drops a line whose strength is unspecified or unrecognized (`compactMap`, as `mapFamilyUnit`). `FakeStore.seededPersonRecordLinks: [String: CatalogPersonRecordLinks]`, unseeded → the same `conclusiondetails.not_found` error `listPersonEvents` throws. `ThrowingStore` throws `StoreBoom.boom` | S13-04 / S13-05's `Catalog*` shape and store conventions | A Swift-side pair index — Go already pairs |
| D-10 | `core/matchregistry/guard_test.go` `readerPackages` gains `{"core/database/persondetail", []string{"core/graphalign", "core/database/promotealign"}}` | S13-29 D-11 names this row | — |
| D-11 | Test exports in `core/recordlinks/export_test.go`: `Lines` (D-4) and `ReadWork(page []byte, graphs []Graph, cfg Config) int` *(new)*, which calls `readWork(page, graphs, h, cfg, math.MaxInt)` *(new, unexported)*, the count of D-2 (2) and (3) that `Pairs` calls with limit `MaxReadWork`. `PairWork` keeps its signature and derives the shared entities as `pairLines` does when called alone | The cap is set from a number the tests can print, as `PairWork` does for `MaxPairWork` | Exporting the work in production — no production caller |
| D-12 | `evidencelayertest.Open(t testing.TB, c *database.Catalog) *Records` *(new)*: upserts the `UserID` user (beside any user the catalog has), creates Source "Register" with Artifact "Scan", and returns Records on `c`. It registers no close: the caller owns `c`. `New` becomes: create the catalog, register its close, `subjectvocab.Install(c)`, then `Open(t, c)`. Doc: "Open is Records on a catalog that already exists, such as a project's; the catalog must hold the seeded vocabulary." The handler test builds on a project from `onboarding.Complete` inside one `catalogsession.Do` | The handler reads a project dir; the scenario needs Evidence graphs and promotes, which `evidencelayertest` already builds. One constructor body, not two | Building AC-1.1's records through RPCs — about 40 calls. A second fixture package for handlers |
| D-13 | Scale fixture. Each household graph holds the son "Thomas Smith", the father "William Smith", the mother "Mary Ann Jones" and the sister "Ann Smith", each cited with its name only, and four `relationship` "parent" links: father → son, mother → son, father → sister, mother → sister. Every graph's four Subjects are promoted onto the same four entities. Each pair of graphs then has three lines (father, mother, sister), each strong. `households(f *fixture, n int) (son []byte)` *(new)* in `core/database/persondetail/recordlinks_perf_test.go` *(new)* builds n Sources "Household 1" to "Household n" with `Records.Person`, `Relationship` and `promotetest.Save`. `householdGraphs(n int) []recordlinks.Graph` *(new)* in `core/recordlinks/recordlinks_test.go` builds the same in memory: member ids `[]byte{byte(i), k}` per graph i (k 1 to 4 for son, father, mother, sister), entities `entP`, `entF`, `entM`, `entS`, names through `pt.Name`, links through `rel(…, "parent", …)` with a distinct association byte per link. Each helper's doc names the other and the numbers both assert (1,225 pairs at 50, three strong lines each) | One shape at the map's Scale for the budget and for the cap's measurement; three shared people per pair is a household, the commonest graph (review M-8) | An obituary-sized fixture — the map's Scale names 50 graphs of mostly 5 – 20, a household's few people |
| D-14 | `core/database/databasetest/budget.go` *(new)*, with the same `//go:build cgo` line as `selects.go`: `Fastest(n int, run func() time.Duration) time.Duration`, `Since(call func()) time.Duration` (bodies of `promotealign`'s `fastest` and `since`), and `WithinBudget(t testing.TB, what string, got, budget time.Duration)`, the body of `TestLargeSourceBudgets`' `over` closure (log, skip under `graphcache.VerifyMode()`, `t.Errorf` over budget). The package doc stays in `selects.go` and becomes "Package databasetest holds catalog test helpers: watching a catalog's SQLite connection and timing work against a budget." `promotealign/perf_test.go` deletes `fastest`, `since` and `over` and calls these | A second budget test would copy all three (memory "same-pass cleanups") | A copy in `persondetail` |
| D-15 | Budgets for `TestRecordLinksBudgets` *(new)*: cold read 1.5 s, repeat read 300 ms, 65-graph repeat read 1.5 s; cold is the best of 2 (each after `Graph().Drop()` and `catalogsession.WarmOnOpen`), repeat the best of 5 with `runtime.GC()` before each, as `TestLargeSourceBudgets` does | Within the order of Promote's budgets on a smaller catalog; the 65-graph read reloads every graph each time, so it is held to the cold budget. Times go in the ship log (map Check) | A budget only logged — a slow change would land unseen |
| D-16 | Swift key `.personRecordLinks(project:, entityId:)` *(new)* after `.personEvents` in the cases, in `Kind`, and in the `kind` and `project` switches; doc "One Person's Evidence graphs and the lines between them (S13-27), read for the record links view. Evicted, not revalidated, while off screen." Its refreshes are `.personClaims`' set, keyed: its own handle through `.touchedHandles`, the graphs its value holds through `.evidenceGraphs`, and `.followsPropertyLabels` | Working rule "One invalidation rule for person-section keys": "unless the plan finds a keyed effect". Every write that can change a line names a graph the value holds (a neighbor's claim names its graph through `claimGraph`; a Subject, Observation or Source write names its graph) or the person's own handle (a graph gained or lost). A write elsewhere does not reload 1,225 pairs (review M-1). Property vocabulary changes the metas the count reads | `applyHiddenPolicyOnAny(follows: .touchedHandles)` as on Family and Timeline — reloads on every handle write anywhere while the page is shown |
| D-17 | Structural-debt A-34, on this PR. Two `static func`s beside `make` in `CatalogQueryRegistry.swift`. `CatalogQueryDescriptor.personSection<Value>(kind: CatalogQueryKey.Kind, load: @escaping @Sendable (CatalogQueryKey, any GenealogyStore) async throws -> Value) -> CatalogQueryDescriptor` *(new)*: `.make(kind:, stalePolicy: .sessionFresh, whenHidden: .evict, refreshes: [.applyHiddenPolicyOnAny(follows: .touchedHandles), .followsPropertyLabels], load:)`, doc "A per-person section key that a relative's or co-participant's change reaches: that change touches their id, not the Person's, so any touched handle reloads it." `.personFamily` and `.personEvents` move onto it, keeping their comments. `CatalogQueryDescriptor.personGraphsSection<Value>(kind:, entityId: @escaping @Sendable (CatalogQueryKey) -> String?, graphIDs: @escaping @Sendable (Value) -> Set<String>, load:)` *(new)*: `.make(kind:, stalePolicy: .sessionFresh, whenHidden: .evict, refreshes: [.applyHiddenPolicy(follows: .touchedHandles) { _, key in entityId(key).map { [$0] } ?? [] }, .applyHiddenPolicy(followsValue: .evidenceGraphs, ids: graphIDs), .followsPropertyLabels], load:)`, doc "A per-person section key read from the Person's Evidence graphs: it reloads when its own handle or one of the graphs its value holds is written." `.personClaims` moves onto it unchanged in behavior; `.personRecordLinks` uses it with `{ (v: CatalogPersonRecordLinks) in Set(v.graphs.map(\.graphID)) }`. In `everyKindHasOneDescriptor`, `#expect(kinds.count == 21)` is deleted: `Set(described) == Set(kinds)` over `Kind.allCases` already proves one descriptor per kind | The A-34 row says "touch it with S13-27"; two shapes, each used twice or more. The hard-coded count fails only when a kind is added, which the set equality already covers | One builder with optional arguments — two refresh rules behind flags. Leaving `.personClaims` as its own copy — the record-links key would be its second copy |
| D-18 | `BenchmarkPairsAtCap` *(new)* measures the slowest crafted read under the shipped cap on the "names" axis, the slowest in `BenchmarkLinesAtCap` (S13-26 ship log: 651 ms). `solo`'s per-graph closure is factored out as `soloGraph(key, text, wordsEach, n int) func(tag string, id, ev func(byte) []byte) recordlinks.Graph` *(new)*, which `solo` passes to `pair`. The benchmark builds m graphs, graph i through `soloGraph(name, pt.Name, cfg.Value.Words.MaxExactPairing, n)(fmt.Sprintf("g%d", i), id, nil)` with `id = func(k byte) []byte { return []byte{byte(i), k} }` (a distinct tag per graph, so every graph's words differ, as `pair`'s "a" and "b" do), n from `underCap` on the names axis, and m found by starting at 2 and adding one graph while `ReadWork` over m + 1 graphs stays within `MaxReadWork`. It reports ns/op and m | Times the cap's real worst case instead of multiplying a pair benchmark; reuses the axis builders | A new crafted generator — `solo` already builds the worst names |

**Rules this plan bends.** None.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `recordlinks.Pairs`, `recordlinks.Pair` *(new)* | func, type | `recordlinks.Lines` (becomes `pairLines`) | `Pairs` is the one-pair count over every pair plus the read bound; it calls the same internals (D-2, D-4). `rg "func .*Pairs\|type Pair " core` finds none in the matching layers |
| `recordlinks.Rules.MaxReadWork` *(new)* | registry knob | `recordlinks.Rules.MaxPairWork` | Extends the same section |
| `readWork`, `ReadWork` *(new)* | func, test export | `pairWork`, `PairWork` | The same count one level up |
| `persondetail.RecordLinks`, `PersonRecordLinks`, `linksBetween`, `recordlinks.go` *(new)* | func, type, file | `persondetail.PersonEvents`, `persondetail.Claims` | One read per section (working rule); reuses the claims read for its graphs (D-5) |
| `personNode`, `claims` *(new, unexported)* | func | `PersonEvents`' inline check, `Claims` | Extracted, not added (D-6) |
| `GetPersonRecordLinks` RPC and messages *(new)* | FFI | `ListPersonEvents`; `EvidenceGraph`, `PersonClaim` messages | One RPC per section; `EvidenceGraph` has fields this read would send empty; `PersonClaim` carries a Subject and Observations (D-7) |
| `recordLinkStrengthProto` *(new)* | func | `kinshipPatternProto` | The same mapper pattern, its own enum |
| `getPersonRecordLinks`, `CatalogPersonRecordLinks` and its parts *(new)* | store method, types | `listPersonEvents`, `CatalogPersonTimeline` | One method and value type per person read |
| `.personRecordLinks` *(new)* | query key | `.personClaims` | One key per section (working rule); shares `.personClaims`' builder |
| `personSection`, `personGraphsSection` *(new)* | helper | `CatalogQueryDescriptor.make` | Fold the copies (A-34, D-17) |
| `evidencelayertest.Open` *(new)* | test helper | `evidencelayertest.New` | `New` now calls it (D-12) |
| `databasetest.Fastest`, `Since`, `WithinBudget`, `budget.go` *(new)* | test helpers | `promotealign` `fastest`, `since`, `over` | Moved, not copied (D-14) |
| `soloGraph`, `householdGraphs` *(new)* | test helpers | `solo`, `obituary` | `soloGraph` is `solo`'s closure, extracted (D-18) |
| `households` *(new)* | test helper | `claims_test.go` `burialsAndBaptisms` | A different shape at Scale (D-13) |
| Tests in §8 *(new)* | tests | — | — |

**Builds on.** `add-match-rule` "Add a knob"; the person-section read shape of S13-05; `evidencelayer.Load`'s memo (S13-29); the `readerPackages` guard (S13-29 D-11); `.personClaims`' value-keyed refresh (S13-01).
**Deletes.** `core/database/promotealign/perf_test.go` `fastest`, `since` and the `over` closure (D-14); the exported `recordlinks.Lines` (D-4); the policy arguments of the `.personClaims`, `.personFamily` and `.personEvents` descriptors (D-17); `everyKindHasOneDescriptor`'s hard-coded count (D-17); `PersonEvents`' inline Person check (D-6); `docs/ideas/structural-debt.md` row A-34.
**Next adds.** S13-28 reads `.personRecordLinks` and warms it from the Sources section in `PlaceRegistry`. Event and Place record links call `linksBetween` with their own graph list. The S13-16 PRF-2 batch read replaces the per-graph `evidencelayer.Load` loop in `linksBetween`. A new per-person section key picks `personSection` or `personGraphsSection` (`add-workspace-place`).

### Reach

| New path or state | Destination | Checked: resets on the new id · handles a missing referent · shows the state |
| --- | --- | --- |
| An Event, Place, merged, unknown or malformed id | `persondetail.RecordLinks` via `personNode` | `conclusiondetails.ErrNotFound` (AC-1.6) |
| A person with no accepted member | `RecordLinks` | No graph, no pair; `claims` returns no group |
| A person in more than 64 graphs | `graphcache` LRU | Every read reloads graphs and layers; the result is unchanged (AC-2.2) |

No view reads the key in this PR; no navigation path changes.

### Scale

At the map's Scale (50 graphs, 1,225 pairs, a household's few shared people per pair): `RecordLinks` reads 50 graphs (cached after the first read), 50 layers (memoized), the value rules and metas once, then `Pairs`: the entity index is linear in the graphs' persons, and each of the 1,225 pairs compares three shared people. Measured by `TestRecordLinksBudgets` (AC-2.1) and grown past the cache cap at 65 graphs (AC-2.2). Writes: none. The key reloads only on a write that names its own handle or one of its graphs (D-16), and is evicted while hidden.

### Edges

- **Assumptions.** The one-pair count is symmetric and order-free (`TestLinesInvariants` (a), (b)); `Pairs` passes the lower index first and relies on that. One `holdings` across pairs assumes member ids are unique across graphs: they are Subject ids (one graph per Subject, `graphcache.Member` doc). `NewGraph` ignores a repeated member id. `Pairs` equals the one-pair count per pair under the read cap: `TestPairsInvariants` (T-3).
- **Neighbors.** A shared person with no matching fact: a weak line at 0 facts (S13-26 rule, unchanged). A neighbor with a provisional claim, or never promoted: no entity, no line (AC-1.3). Two members of the son in one graph: the graph once (AC-1.5), both excluded as the page. A pair over `MaxPairWork` with the read under `MaxReadWork`: that pair's lines at 0 facts, the others counted (AC-2.5). A read exactly at `MaxReadWork`: compared (the cap is "passes"); one over: not (T-2 rows).
- **Removed inputs.** Unpromoting the father in B removes F from B's members: the A–B line goes; the key reloads because the claim write names B (AC-3.1). Unpromoting the son from C removes C from `Graphs`; the write touches the son's handle (AC-3.2). A deleted graph leaves `claims`, so it leaves the read. A Source retitled: its graphs are named (AC-3.1).
- **Real data.** Not driven by a shipped table. `MaxReadWork` is set against the household fixture at Scale; the dogfood projects' people are S13-28's manual check.
- **Untrusted size.** Graphs per person: unbounded by the catalog; each costs one cached load and one memoized layer (structural-debt PRF-2 for > 64). Lines per read: counted from the entity index before any pair is built, bounded by `MaxReadWork` (AC-2.4). Work per pair: `MaxPairWork`; per read: `MaxReadWork`, counted before any compare (AC-2.3). Members per entity, values per property, bytes per value: S13-26's bounds, unchanged. Edges per graph walk: `maxWalkEdges` (S13-29).

### Schema

None — no catalog change.

### Go core

| Package | Change |
| --- | --- |
| `core/recordlinks` | `Pair`, `Pairs`, `readWork` *(new)* (D-2, D-11); `Rules.MaxReadWork` *(new)* (D-3); `Lines` → `pairLines` (D-4); `pairWork` and the one-pair loop read a sorted list of shared entities (D-2); `export_test.go` `Lines`, `ReadWork` *(new)* (D-4, D-11) |
| `core/matchregistry` | `defaultRecordLinks` sets `MaxReadWork` (D-3); `guard_test.go` row (D-10) |
| `core/database/persondetail` | `recordlinks.go` *(new)*: `PersonRecordLinks`, `RecordLinks`, `linksBetween` (D-5); `claims.go`: `personNode`, `claims` (D-6); `timeline.go`: `PersonEvents` calls `personNode` (D-6) |
| `core/database/evidencelayer/evidencelayertest` | `Open` *(new)*; `New` calls it (D-12) |
| `core/database/databasetest` | `budget.go` *(new)*; package doc (D-14) |
| `core/database/promotealign` | `perf_test.go` uses `databasetest` (D-14) |

```go
// Pair is the lines between two graphs of a read; A < B index the graphs
// Pairs was given.
type Pair struct {
	A, B  int
	Lines []Line
}

func Pairs(page []byte, graphs []Graph, metas subjectcompare.Metas, cfg Config) []Pair

// PersonRecordLinks is a Person's Evidence graphs, in the Sources section's
// order, and the lines between each pair; Pairs index Graphs.
type PersonRecordLinks struct {
	Graphs []graphcache.EvidenceGraph
	Pairs  []recordlinks.Pair
}

func RecordLinks(c *database.Catalog, entityID []byte) (PersonRecordLinks, error)
```

Errors: `conclusiondetails.not_found` → an unknown, merged, non-Person or malformed id.

### FFI

`api/proto/engine.proto`: `METHOD_GET_PERSON_RECORD_LINKS = 116` after `METHOD_GET_UNPROMOTE_IMPACT`; the messages and enum of D-7 after `PersonEvent`. Handler `GetPersonRecordLinks` (D-8) in `api/ffi/handlers/person_detail.go`; dispatch in `api/ffi/dispatch.go`. Regenerate with `./scripts/generate-proto.sh`.

### macOS

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Store method | `getPersonRecordLinks` | `GenealogyStore.swift`, `GoStore.swift`, `FakeStore.swift`, `CoreInvoke.swift` (`getPersonRecordLinks` method id), `OnboardingModelTests.swift` `ThrowingStore` | D-9 |
| Value types | `CatalogPersonRecordLinks` and parts | `GenealogyStore.swift` | D-9 |
| Query key | `.personRecordLinks` | `CatalogQueryKey.swift`, `CatalogQueryRegistry.swift` | D-16, D-17 |

States: none drawn; no view reads the key. Accessibility: none.

### Strings

None — no user-facing copy.

## 7. Steps

**STEP-1 — Every pair of a read, with its bound**
Files: `core/recordlinks/recordlinks.go`, `core/recordlinks/export_test.go`, `core/recordlinks/recordlinks_test.go`, `core/recordlinks/invariants_test.go`, `core/matchregistry/recordlinks.go`, `docs/matching-model.md`. Skill: `add-match-rule` ("Add a knob").
Do: D-2, D-4, D-11, D-13's `householdGraphs`, D-18. Reword the `Rules` field docs that say "one Lines call" (`MaxEntityMembers`, `MaxPairWork`, `MaxValueBytes`) to "one pair of graphs", and `docs/matching-model.md`'s `recordlinks.MaxPairWork` row's "a `Lines` call" to "a pair of graphs". New numeric literals in `core/recordlinks` are only 0 or 1. Write T-1 – T-3 first, with `MaxReadWork` set to `math.MaxInt` in the registry; `TestPairsAtScale` (T-4) logs `ReadWork` at 50 graphs; then set D-3's value, its comment, the inventory row, and one sentence in the line-33 paragraph: "`Pairs` compares every pair of a read; a read whose number of lines plus pair work passes `recordlinks.MaxReadWork` compares nothing: every line keeps 0 facts, and when the lines alone pass it, no line is listed." Replace `Lines` with `Pairs` where line 33 names the entry. Add `BenchmarkPairsAtCap` (T-5) and run it.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/recordlinks/ ./core/matchregistry/` passes (it runs `FuzzLines`' seeds), with every `TestLines*` row unchanged; `CGO_ENABLED=1 go test -tags fts5 -run '^$' -bench 'BenchmarkPairsAtCap' ./core/recordlinks/` prints ns/op and m.

**STEP-2 — Records on an existing catalog**
Files: `core/database/evidencelayer/evidencelayertest/evidencelayertest.go`.
Do: D-12.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/evidencelayer/... ./core/database/promotealign/ ./core/database/persondetail/` passes unchanged.

**STEP-3 — The record-links read**
Files: `core/database/persondetail/recordlinks.go` *(new)*, `core/database/persondetail/claims.go`, `core/database/persondetail/timeline.go`, `core/database/persondetail/recordlinks_test.go` *(new)*, `core/matchregistry/guard_test.go`. Skills: `add-catalog-query`, `use-catalog-session`.
Do: D-5, D-6, D-10. Add T-6 on the `claims_test.go` `fixture` (`newFixture`). Builders: each Source with `src := f.source(title)` and `art := f.arts[string(src.ID)]`; on a Source's sole graph, `f.Person(src, art, name)`, `f.Birth(src, art)` and `f.Participation(src, art, father, birth, "subject")` (AC-1.1, AC-1.3, AC-1.5, AC-1.7); accepted claims with `f.promote(s, onto)`, AC-1.3's provisional claim with `f.provisional(s, entity)`. AC-1.4: rename "Baptisms"' sole graph to "Register" with `evidencegraphs.Update` and add "Index" with `evidencegraphstest.Create`, as `burialsAndBaptisms` does; the son on each graph with `f.person(graphID, "Thomas")` and `f.promote`.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/persondetail/ ./core/matchregistry/` passes, `TestReaderPackages/core/database/persondetail` and every existing timeline and claims test included.

**STEP-4 — Budgets at Scale**
Files: `core/database/databasetest/budget.go` *(new)*, `core/database/databasetest/selects.go` (package doc), `core/database/promotealign/perf_test.go`, `core/database/persondetail/recordlinks_perf_test.go` *(new)*.
Do: D-14 first (the promotealign test passes on the moved helpers), then D-13's `households` and D-15's `TestRecordLinksBudgets` (T-7).
Check: `CGO_ENABLED=1 go test -tags fts5 -run 'TestRecordLinksBudgets|TestLargeSourceBudgets' -v ./core/database/persondetail/ ./core/database/promotealign/` passes and logs the cold, repeat and 65-graph times.

**STEP-5 — Verify mode**
Files: none.
Do: run the read's tests under the cache's verify mode.
Check: `PROVENENCIA_GRAPH_VERIFY=1 CGO_ENABLED=1 go test -tags fts5 ./core/database/persondetail/` passes.

**STEP-6 — Wire messages**
Files: `api/proto/engine.proto`, generated `api/proto/engine/engine.pb.go`, `macos/App/Platform/Generated/engine.pb.swift`. Skill: `add-ffi-handler`.
Do: D-7. Run `./scripts/generate-proto.sh`.
Check: `go build ./...` passes; the generated files hold `METHOD_GET_PERSON_RECORD_LINKS`.

**STEP-7 — The handler**
Files: `api/ffi/handlers/person_detail.go`, `api/ffi/dispatch.go`, `api/ffi/handlers/person_detail_test.go`. Skills: `add-ffi-handler`, `use-catalog-session`.
Do: D-8. Add T-8, T-9.
Check: `CGO_ENABLED=1 go test -tags fts5 ./api/ffi/...` passes.

**STEP-8 — Swift store**
Files: `macos/App/Platform/GenealogyStore.swift`, `GoStore.swift`, `FakeStore.swift`, `CoreInvoke.swift`, `macos/ProvenenciaTests/OnboardingModelTests.swift`. Skill: `add-swift-test`.
Do: D-9.
Check: `xcodebuild build` for the app scheme succeeds with no new warning.

**STEP-9 — The cache key and the two builders**
Files: `macos/App/Features/Workspace/Session/CatalogQueryKey.swift`, `CatalogQueryRegistry.swift`, `macos/ProvenenciaTests/CatalogQueryRegistryTests.swift`, `macos/ProvenenciaTests/WorkspaceSessionTests.swift`. Skills: `add-workspace-place`, `add-swift-test`.
Do: D-16, D-17. Add T-10 – T-12. In `warmedVocabularySession`, add `"personRecordLinks": .personRecordLinks(project: project, entityId: "e1")` with `session.setQueryValue(keys["personRecordLinks"]!, value: CatalogPersonRecordLinks())`; add `"personRecordLinks"` to `familyCases`' "property vocabulary" `reached`, and `.personRecordLinks` to `noMutationReachesAConclusionKey`'s `conclusionKinds`.
Check: `xcodebuild test` passes `CatalogQueryRegistryTests` and `WorkspaceSessionTests`; the existing `personClaims…`, `personFamily…` and `personEvents…` tests pass unchanged.

**STEP-10 — Docs and skills**
Files: `docs/deployment-plan/spike-13/deployment-plan.md` (map row Status → `Landed #N`; the S13-27 scope entry: in **Ships**, the key refreshes on its own handle and the Evidence graphs it holds (the keyed effect the working rule allows) and neighbors' handles come from each graph's accepted members; in **Check**, name `recordlinks.MaxReadWork` as the read's bound), `docs/deployment-plan/spike-13/evidence-graph-comparison.md` (§ Scale: "A read of one person is bounded as a whole by `recordlinks.MaxReadWork`, beside each pair's `MaxPairWork`."), `.agents/skills/add-match-rule/SKILL.md` line 34 (six knobs with `MaxReadWork`; `Pairs` as the entry, the one-pair count unexported; the read-wide bound in the parenthesis), `.agents/skills/add-workspace-place/SKILL.md` Refresh row (one sentence: a per-person section key uses `CatalogQueryDescriptor.personSection` when a relative's change reaches it, or `personGraphsSection` when it reads the Person's Evidence graphs), `docs/ideas/structural-debt.md` (row A-34 removed), `docs/deployment-plan/spike-13/completed.md`, this plan (status, Landed note).
Do: per the Definition of done.
Check: links resolve; `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-13/prs/S13-27-person-record-links-read.md` reports no errors.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/recordlinks/recordlinks_test.go` `TestPairs` *(new)*, table | graphs built with `build`, `person`, `pt.Name`: F named "John Smith" in graphs 0 and 1 → `[{0, 1, [{entF, 1, weak}]}]`; F in 0 and 2, M in 1 and 2 → `[{0, 2, [F]}, {1, 2, [M]}]`; the page person in every graph → no line names `entP`; one graph → nil; no graph → nil; a graph sharing no one → in no pair | AC-1.1, AC-1.2, AC-1.7, D-2 |
| T-2 | Go | `core/recordlinks/recordlinks_test.go` `TestPairsReadCap` *(new)*, table, `cfg.RecordLinks.MaxReadWork` set per row from `ReadWork` | cap = `ReadWork` → lines counted; cap = `ReadWork` − 1 → every pair listed, every line 0 facts, weak; three lines and cap 2 → nil; a pair over `MaxPairWork` (the names axis) beside a small pair, read under the cap → that pair 0 facts, the other counted; each row also run with the graphs reversed: pair (A, B) becomes (n−1−B, n−1−A) with equal lines | AC-2.3, AC-2.4, AC-2.5, D-2 |
| T-3 | Go | `core/recordlinks/invariants_test.go` `TestPairsInvariants` *(new)* | 100 rounds of four `genGraph` graphs: each pair of `Pairs` equals `Lines` over the same two graphs; a pair is present exactly when `Lines` is non-empty | AC-2.5, D-2 |
| T-4 | Go | `core/recordlinks/recordlinks_test.go` `TestPairsAtScale` *(new)* | `householdGraphs(50)` → 1,225 pairs, each with lines for `entF`, `entM`, `entS`, each strong; logs `ReadWork` | AC-2.1, D-3, D-13 |
| T-5 | Go | `core/recordlinks/recordlinks_test.go` `BenchmarkPairsAtCap` *(new)* | D-18's read → ns/op and m, in the ship log | D-3, D-18 |
| T-6 | Go | `core/database/persondetail/recordlinks_test.go` `TestRecordLinks` *(new)*, table | AC-1.1's records → graphs A, B, C, D; pairs `{A, B, [{F, 2, strong}]}`, `{A, C, [{M, 1, weak}]}`; AC-1.3 → no pair with E1 or E2; AC-1.4 → that order; AC-1.5 → the graph once; AC-1.6 → `conclusiondetails.ErrNotFound` each; AC-1.7 → one graph no pair, two graphs no pair; a warm repeat read with `SetQueryCounterForTest` → 0 queries | AC-1.1 – AC-1.7, AC-2.6, D-5, D-6 |
| T-7 | Go | `core/database/persondetail/recordlinks_perf_test.go` `TestRecordLinksBudgets` *(new)* | subtests "fixture" (50 graphs → 1,225 pairs, 3 strong lines each), "cold", "repeat", "past the cache" (65 graphs: equals the read after `Graph().Drop()`; repeat within 1.5 s) | AC-2.1, AC-2.2, D-15 |
| T-8 | Go | `api/ffi/handlers/person_detail_test.go` `TestGetPersonRecordLinks` *(new)* | AC-1.1's records on an `onboarding.Complete` project through `evidencelayertest.Open` inside one `catalogsession.Do` → four graphs with their ids, refs, names, Source ids, titles and type ids; pairs (A, B) with F strong and (A, C) with M weak, by graph id; an Event id → `conclusiondetails.not_found` | AC-1.1, AC-1.6, D-7, D-8, D-12 |
| T-9 | Go | `api/ffi/handlers/person_detail_test.go` `TestRecordLinkStrengthProto` *(new)* | each `recordlinks.Strength` → a value other than unspecified | D-8 |
| T-10 | Swift | `macos/ProvenenciaTests/CatalogQueryRegistryTests.swift` `personRecordLinksRevalidate` *(new)* | warmed `.personRecordLinks(e1)` holding g1, g2: `evidenceGraphs: ["g1"]` → revalidated; `handles: ["e1"]` → revalidated; `handles: ["e7"], evidenceGraphs: ["g9"]` → not; hidden policy `.evict`; stale policy `.sessionFresh`; refreshes' follows `[.touchedHandles, .evidenceGraphs, .list(.propertyVocabulary)]` | AC-3.1, AC-3.2, AC-3.3, AC-3.5, D-16 |
| T-11 | Swift | `macos/ProvenenciaTests/CatalogQueryRegistryTests.swift` `personRecordLinksLoadThroughTheStore` *(new)* | `FakeStore` seeded for e1 with two graphs and one pair of one weak and one strong line → the handle's value equals the seed; e9 → `CoreInvokeError.coded(status: 1, code: "conclusiondetails.not_found", kind: .user, params: [])` | AC-3.6, D-9 |
| T-12 | Swift | `CatalogQueryRegistryTests.swift` `listFamiliesReachTheirKeys`, `noMutationReachesAConclusionKey`, `everyKindHasOneDescriptor` (updated); `WorkspaceSessionTests.swift` `personRecordLinksFollowEveryGraphUntilLoaded` *(new)*, a copy of `personClaimsFollowEveryGraphUntilLoaded` | the property-vocabulary family reaches `personRecordLinks`; no mutation reaches it; every kind has one descriptor, with no hard-coded count; before its first value the key follows every graph, after it only its own | AC-3.4, D-16, D-17 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `TestLines*`, `FuzzLines` (the `pairWork` refactor and rename), `TestKnobInventoryMatchesRegistry`, `TestLargeSourceBudgets` (the moved helpers), the `persondetail` timeline tests (`personNode`), the `personClaims…` / `personFamily…` / `personEvents…` Swift tests (the builders).

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] PROVENENCIA_GRAPH_VERIFY=1 go test -tags fts5 ./core/database/persondetail/ passes
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left (none expected)
- [ ] Sweep: `rg -w 'fastest|since' core/database/promotealign` finds no definition or call left (only `time.Since`); `rg 'recordlinks\.Lines' --glob '!*_test.go'` and `rg 'five knobs'` find no hit across core/, .agents/, docs/*.md; `rg -w 'Lines' core/recordlinks/recordlinks.go docs/matching-model.md` finds no reference to the old function; "A-34" only in completed.md and this plan; `scripts/check-guidance-names.py` passes (check-pr runs it)
- [ ] Every row of **Rules this plan bends** has the user's answer, and its rule is updated (none)
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New (no UI)
- [ ] Version bump as the header says (none)
- [ ] Ship log: the 50-graph read work, `MaxReadWork`'s value, `BenchmarkPairsAtCap` ns/op and m, and the cold, repeat and 65-graph times
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status and the S13-27 Check line; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated (not gated: none)
- [ ] Domain docs and skills updated: docs/matching-model.md, evidence-graph-comparison.md § Scale, add-match-rule, add-workspace-place, docs/ideas/structural-debt.md (A-34 removed)
- [ ] PR description: what shipped, AC evidence, deviations, the measured numbers
```

## 10. Manual check

None in the app — no view reads `.personRecordLinks` until S13-28, whose manual check opens dogfood people with it. For this PR, run `CGO_ENABLED=1 go test -tags fts5 -run 'TestRecordLinks|TestGetPersonRecordLinks' -v ./core/database/persondetail/ ./api/ffi/handlers/` and read the logged times.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| The `pairWork` refactor changes the one-pair result | `TestLines*`, `TestLinesInvariants`, `FuzzLines` seeds unchanged in STEP-1; T-3 compares `Pairs` with it |
| 20× the Scale work makes a crafted read slow | `BenchmarkPairsAtCap` measures it; the stop condition below caps how slow |
| The household fixture takes long to build | 50 Sources × 4 people; `TestLargeSourceBudgets` builds a 4,000-person catalog in the same suite |

**Stop and report if:**

- `persondetail.Claims` no longer orders graphs as §4 says, or skips members differently.
- Any existing `TestLines`, `TestLinesBounds`, `TestLinesWorkCap` or `TestLinesValueSize` row changes after the `pairWork` refactor.
- 20 times the 50-graph work makes `BenchmarkPairsAtCap` take more than 60 s per op on the implementer's Mac: report the work, the value, m and the time.
- AC-1.1's A–B line is not {F, 2, strong} or its A–C line is not {M, 1, weak} with the records of §2: report the facts counted and which property or link differs.
- A household pair's lines are not three strong lines (T-4, T-7): report the facts per line.
- `promotetest.Save` refuses a second Subject of one graph onto S (AC-1.5).
- The cold read at 50 graphs measures over 1.5 s, or the repeat over 300 ms, on the implementer's Mac.
- `evidencelayertest.Open` cannot run inside `catalogsession.Do` on an `onboarding.Complete` project (a missing vocabulary row, a nested write, a user conflict).

## 12. Open questions

None.

## 13. PR

**Title:** S13-27 — Person record links read
**Body:** objective in one sentence; what shipped against US-1 – US-3; AC evidence; the measured read work, cap and times; deviations from the plan; `Plan: docs/deployment-plan/spike-13/prs/S13-27-person-record-links-read.md`.
