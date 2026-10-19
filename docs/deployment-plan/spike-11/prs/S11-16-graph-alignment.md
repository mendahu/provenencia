# S11-16 — Rewrite graph alignment

**Status:** Planned
**Spike / slice:** Spike 11 · Slice 4 — Segments and graph alignment
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no new chrome: an engine rewrite; the wire, the page and its copy are unchanged, and the contested marker's chrome is S11-18 behind S11-D1)
**Board:** —
**Depends on:** S11-14 (landed [#493](https://github.com/mendahu/provenencia/pull/493); [plan](S11-14-segment-compare.md) and its **Landed** note, which changes nothing this plan reads), S11-15 (landed [#497](https://github.com/mendahu/provenencia/pull/497); [plan](S11-15-shortlist-index.md) and its **Landed** note: the stats slot was renamed and `matchIndex` holds only the shortlist; `shortlist.Rules` gained `MaxBuckets`; a vocabulary commit also drops the match index; `loadCanon` skips the index when every Subject is fixed). Branch from `origin/main` (fetch first)
**Blocks:** S11-17, S11-19, S11-21
**Branch:** `s11-16-graph-alignment`
**Version bump:** none — Spike 11 PRs leave `VERSION` at 0.2.0, and the product bump is cut when a release ships ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-match-rule`](../../../../.agents/skills/add-match-rule/SKILL.md) ("Add a knob", "Add a link rule", "Add a relation, cause or unit code", "Add a layer package"; this PR extends the skill)
**Rules that bind:** `.cursor/rules/go-tests.mdc`, `.cursor/rules/match-rules.mdc`
**Context:** [Layer 6 procedure](../comparison-layers.md#layer-6-procedure), [Bands](../comparison-layers.md#bands), [Segments](../comparison-layers.md#segments-layer-5-and-graph-alignment-layer-6), [Match registry](../comparison-layers.md#match-registry), [`docs/matching-model.md`](../../../matching-model.md) (glossary: nominated handle, candidate, queue, dirty pair, settle, contested; knob inventory `alignment.*`); working rules "Graph alignment follows the layer 6 procedure", "Bands come from evidence", "One match registry", "One guard, not one per package", "One matching skill, grown with the layers", "One wire vocabulary", "High tolerance for churn", "Registry numbers: carry over, then estimate"; logged follow-ups: [S11-11 review](S11-11-subject-profile.review.md) M-9 (import guard), [S11-01 plan](S11-01-matching-model.md) D-5 (rewrite `promote-graph-alignment.md` §4.1, §5, §6, §12 and re-point links into them), [S11-05 ship log](S11-05-migrate-matching.ship.md) MNT-3 (registry comments naming deleted identifiers)
**Planned:** 2026-10-19 against `main` @ `8915340` (S11-15 merged); drafted the same day against `219a3b8` and refreshed
**Review:** [`S11-16-graph-alignment.review.md`](S11-16-graph-alignment.review.md) — Fits with changes / Ready with changes, 2 rounds (findings fixed)

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

Promote's graph alignment becomes the layer 6 procedure: each Subject's handles come from its shortlist or from a link to an accepted pair; the best K by similarity become candidates; every candidate enters one best-first queue; an accepted pair re-scores only its neighbors; and when the queue empties, only the pairs whose neighbors moved are settled again. A link whose neighbor is not placed yet is judged by entity compare, so the answer no longer depends on which pair the walk happened to place first. Each row gets a band from its evidence and a contested flag when any of its comparisons conflict. The walk's numbers move into the registry's alignment section, and the three walk-like passes (walk, fallback re-walk, one-hop refinement) and `core/graphalign/registry.go` are deleted.

**Why now.** It is slice 4's last engine PR: S11-14 gave the walk segment compare, S11-15 gave it a shortlist. Dogfood bar 3 ("a decision re-proposes only what moved, each with updated because…") and 4 ("a conflicting row shows contested") need it, and S11-17 (contested over FFI), S11-19 (claim evidence) and S11-21 (performance) build on it.

**What changes for the researcher** (the design, not regressions):
- A look-alike whose linked birth or death differs is never placed first and corrected later: an unplaced birth that differs counts against the pair from the start (D-3).
- A handle that is less than half alike (similarity under 0.5) is no longer offered on its own properties; a handle reached through a link is offered at any score, as today (D-4, the user's call of 2026-10-19).
- A common name keeps every handle tied at the K cutoff, up to 20, so the right one is still there when its links are compared (D-4).
- A handle at least half alike but under the weak band (a same-name look-alike whose sex conflicts, 1.5) is now a candidate, so a row whose only match is such a handle comes back Skip with it as an alternative (band none), not New: [`promote-graph-alignment.md`](../../../promote-graph-alignment.md) principle 5, "New only when it has no candidate at all", unchanged (D-8).
- A decided row's "your neighbors point elsewhere" note appears when a rival leads by 1 point (the settle margin), not 3 (D-8).

**What stays.** The wire (`engine.proto`), the FFI handler, Swift and copy. Row targets, reason codes and their meaning, `Via`, `Alternatives`, the possible-duplicate flags (`flagSharedHandles`, `flagAlikeNewRows`, `alikeNew`; `newPairSegments` reads layer 5's `Unplaced`, D-3), `ScoreCandidate`, provenance at 1, the loader's layer and canon reads, exhibits and pins. Segment and subject numbers.

## 2. User stories

Conventions. `R` is `matchregistry.Default()`; `cfg` is `R.AlignConfig()` (D-1). `canon` in an `Align` test has `Candidates` filled by `withCandidates(layer, canon, cfg)` *(new test helper in `align_test.go`)*, which runs `graphalign.Candidates` over every fixture handle of the Subject's kind, as the loader does over its shortlist. `nameVals`, `vals`, `termProp`, `partLink`, `relLink`, `bridge`, `edge`, `rowBySubject` are the existing helpers in `align_test.go`. "Ann Ames" is `nameVals("Ann", "Ames")`; "male" adds `sex_at_birth: term(male)`. A birth in year Y is `vals(event_type: birth, date: Y)`. Name agree is 2.5, sex agree 0.6 / conflict 1, an agreeing birth's own evidence 2.5 (type 0.5 + date 2), a differing birth's −2.5 (0.5 − 3); a One link earns 1.5 + 0.5 × neighbor evidence and loses 3 apart; an Other link earns 0.25 + 0.25 × neighbor evidence.

**US-1.** As a researcher, a sparse Subject files on the handle whose links line up, from the first pass, and not on a look-alike.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | `TestSparseNameAndBirthDate`'s fixture: person PER-S "Ann Ames" — birth EVT-S 1817; canon PER-1 "Ann Ames" — EVT-1 1817, PER-2 "Ann Ames" — EVT-2 1830 | PER-S → PER-1 strong at 5.25; EVT-S → EVT-1 strong at 5.25; PER-2 is an alternative of PER-S at −0.5 (2.5 − 3: the unplaced EVT-S differs from EVT-2) |
| AC-1.2 | AC-1.1 run with EVT-S fixed on EVT-1, then with PER-S fixed on PER-1 | Every other row's target, handle and band equal AC-1.1's (T-11 (c)) |
| AC-1.3 | Person PER-S "Ann Ames" male — birth EVT-S 1817; canon PER-1 "Ann Ames" female — EVT-1 1817 | PER-S → PER-1 at 4.25 (2.5 − 1 + 1.5 + 0.5 × 2.5), strong, `Flags.Contested` true; EVT-S → EVT-1 at 4.75, strong, `Flags.Contested` false (a neighbor's conflict makes the link partial, not a conflict) |

**US-2.** As the engine, I compare a pair only when something cheap put it forward.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | Layer person "Robnis" with no bridge; canon PER-1 "Robins"; `Canon.Candidates` has no entry for the Subject (the loader's shortlist found none) | Row New, `no_match`: PER-1 is never scored |
| AC-2.2 | Layer person "Robnis" — birth EVT-S 1817, EVT-S fixed on EVT-1; canon PER-1 "Robins" — EVT-1 birth 1817; no `Canon.Candidates` entry for "Robnis" | "Robnis" → PER-1 (nominated through EVT-1), reason `via` the birth row |
| AC-2.3 | `graphalign.Candidates` with three handles at similarity 1 and four at a lower equal similarity (probe "Ann Ames" male; handles "Ann Ames" male ×3 at 1, "Ann Ames" female ×4 at 2.5 / 3.5) | 7 handles: the three, then all four tied at the K-th place |
| AC-2.4 | `Candidates` with 30 "John Smith" handles against "John Smith" | 20 (the tie cap), the first 20 by `Ref` |
| AC-2.6 | Layer person "Ann Ames" male, no bridge; canon PER-1 "Ann Ames" female (similarity 0.71, evidence 1.5) | Row Skip, reason `weak`, assessment none, `Alternatives[0]` PER-1 |
| AC-2.5 | `Candidates` for an event birth 1850 against handles death 1850 (similarity under 0.5) and birth 1850 | The birth only |

**US-3.** As the engine, a decision moves only what it touches, and each move says why.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | Layer P "Ann Ames" male —spouse— S "John Ames"; canon PER-1 "Ann Ames" male (no links), PER-2 "Ann Ames" —spouse— PER-3 "John Ames". `cfg` with `Segment.UnplacedAlike` 100 and `Alignment.SettleMargin` 0.1 | P → PER-2 at 3.375 (2.5 + 0.25 + 0.25 × 2.5) with `Move{Reason: neighbor_placed, NeighborSubjectID: S}`; S → PER-3, `Move` nil |
| AC-3.2 | AC-3.1 with `Alignment.SettleRounds` 0 | P → PER-1 at 3.1, `Move` nil (no settle round) |
| AC-3.3 | AC-3.1's fixture with the shipped `cfg` | P → PER-2 at 3.375, `Move` nil (the unplaced spouse already lined up) |
| AC-3.4 | The invariants (T-11) over the obituary golden, AC-1.1's fixture (this is also AC-1.2's assertion) and the `part_of` chain of `TestPartOfChainIsStrongWithoutADecision` | Same proposal under any input order; same on a second run; fixing any proposed handle row on its handle leaves every other row unchanged; a fixed row never moves |
| AC-3.5 | `TestDecidedRowTheNeighborsContradict` | Still flags `ConflictWithFixed` with `Alternatives[0]` the handle the neighbors point at |

**US-4.** As a contributor, every walk number is one registry field and no layer imports the registry.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-4.1 | `TestKnobInventoryMatchesRegistry` | Passes with no `alignment.` prefix in `notYetBuilt`, rows for every `alignment.*` field and `segment.UnplacedAlike` |
| AC-4.2 | `TestLayerPackages` | `core/graphalign` passes `pure` and `numbers` |
| AC-4.3 | `TestLayerPackages` `pure` subtests | Fail when a `numbers` layer package reaches `core/matchregistry`, or a package other than `core/subjectprofile` reaches `core/autoreconcile`, through any chain of non-test imports; pass on the tree |
| AC-4.4 | `rg -n 'DefaultConfig|seedCandidateLimit|refineOneHop|fallbackUnreachable|AcceptScore|AlternativeLimit|ConflictPenalty' --type go --glob '!core/matchregistry/**'` | Prints nothing (registry source comments keep "carried from …" names) |

## 3. Scope

| In | Covered by |
| --- | --- |
| Segment compare's third placement, `Unplaced`, and `segment.UnplacedAlike` (D-3) | STEP-1, AC-1.1, AC-1.3 |
| The alignment section as `graphalign.Rules`; `registry.go` and `DefaultConfig` deleted; `Config` holds the four sections; guard rows and the import guard (D-1, D-2, D-11) | STEP-2, US-4 |
| Candidates: similarity gate, top K with ties, tie cap (D-4) | STEP-3, AC-2.3 – AC-2.5 |
| The walk: shortlist and nominate, one queue, segments around accepted pairs, settling dirty pairs with move reasons (D-5, D-6, D-7, D-9) | STEP-4, US-1, US-2, AC-3.1 – AC-3.3 |
| Rows: bands from `Alignment.Bands`, `Flags.Contested`, `ConflictWithFixed` on the settle margin, `Row.Move` (D-8, D-9) | STEP-5, AC-1.3, AC-3.5 |
| The invariants test (D-12) | STEP-6, AC-3.4 |
| Promote's loader on the new `Candidates` and `Config` (D-10) | STEP-7 |
| Docs and the skill the rewrite makes false (D-13, D-14) | STEP-8 |

| Out | Home |
| --- | --- |
| Contested and move reasons on the wire, Swift parsing | S11-17 |
| The contested marker and "updated because…" copy | S11-18 |
| Comparing a bridge's own properties (place relationship dates) | Not planned — the scope entry's condition is "if a golden needs them"; no golden does (D-15) |
| A contested flag per alternative | S11-18, if S11-D1 draws one |
| Memoizing layers 1–4 across re-runs; timing the index build over the canon | S11-21 |
| Provenance below 1 (credibility, transcription, claim confidence) | Not planned — unchanged since Spike 9 ([`promote-graph-alignment.md`](../../../promote-graph-alignment.md) §12) |
| Relationships implied by shared events | [`docs/ideas/relationships-from-shared-events.md`](../../../ideas/relationships-from-shared-events.md) |
| A coarse misspelling shortlist key | Not planned — [decided 2026-10-17](../comparison-layers.md#decided-on-2026-10-17) |

## 4. Current state

Verified at `8915340`.

| Area | Where | Today |
| --- | --- | --- |
| Entry | `core/graphalign/align.go` `Align(layer, canon, fixed, cfg *Config)`: `newState`, `seed`, `walk`, `assignedIDs`, `fallbackUnreachable`, `propagateNew`, `walk`, `refineOneHop`, `scoreFixed`, `proposal` | Three walk-like passes; nil cfg reads `DefaultConfig()` |
| Walk state | `align.go` `state` (`assigned`, `fixedSet`, `handleOwner`, `held`, `lostTo`, `best`, `queue`, `seen`, `kindCands`), `adjLink`, `scoredCand`, `queueItem`, `candHeap` | `seen` blocks a pair from being popped twice; `best` only grows during a walk (`recordBest`) |
| Nominating | `align.go` `pushFromAnchor`, `enqueue` | Neighbors of an anchor through corresponding canon links |
| Link scoring | `align.go` `segments`, `compareSegments`, `compare`, `supports` | Only placed neighbors count: `Together` when on the canon neighbor, else `Apart` on the first corresponding link; an unplaced neighbor adds nothing |
| Fallback and refinement | `align.go` `fallbackUnreachable` (`Candidates` over `handlesOfKind`), `refineOneHop`, `rescoreBest`, `reassignUnfixed`, `lostContest` | Re-score every recorded candidate each round, at most one round per Subject |
| Fixed rows | `align.go` `scoreFixed`, `proposal` (`best.score > sc.score + ConflictPenalty` → `ConflictWithFixed`) | Margin 3 |
| Rows | `align.go` `proposal`, `flagSharedHandles`, `flagAlikeNewRows`, `alikeNew`, `newPairSegments`, `strongestAgreement`, `comparisonsFrom`, `pickScore`, `bestNonAssigned`, `alts` (`limit = 3` fallback literal), `alternative`, `subjectHasValues`, `band` | |
| Types | `core/graphalign/types.go` `Subject`, `Bridge`, `Handle`, `CanonEdge`, `Layer`, `Canon`, `Fixed`, `Target`, `Assessment`, `Comparison`, `Via`, `Exhibit`, `Reason` (seven codes), `Alternative`, `RowFlags{ConflictWithFixed, PossibleDuplicate, DuplicateOf}`, `Row`, `Proposal` | No contested flag, no move |
| Config | `core/graphalign/registry.go` `DefaultConfig`, `Config{AcceptScore, StrongScore, MediumScore, WeakScore, ConflictPenalty, AlternativeLimit, Value, Subject, Segment, Shortlist}` (`Shortlist` from `reg.Alignment.Shortlist`); `align.go` `Align` reads `DefaultConfig()` on a nil cfg; the package's only import of `core/matchregistry` | |
| Candidates | `core/graphalign/candidates.go` `Candidates(kind, values, handles, metas, cfg, limit)`: evidence ≥ `WeakScore`, sorted by evidence, similarity, `Ref`, `ID`; doc "Seeding passes it shortlisted handles (S11-15); top K replaces the limit (S11-16)." | S11-13 D-4 gate |
| Score | `core/graphalign/score.go` `Scored`, `ScoreCandidate` | Stays |
| Segment compare | `core/segmentcompare/segmentcompare.go` `Placement` (`Together`, `Apart`), `Cause` (three), `Rules`, `Result`, `Compare`; `key.go` `Corresponds` | No unplaced reading |
| Registry | `core/matchregistry/registry.go` `Registry{Value, Subject, Segment, Alignment}`, `Default()`; `alignment.go` `type Alignment struct{ Shortlist shortlist.Rules }` (its doc says S11-16 replaces it), `defaultAlignment` (`Words`, `Dated`, `YearBucket` 2, `MaxBuckets` 100); `segment.go` `defaultSegment`; `value.go` comments naming deleted code (`// match.DefaultNames affinities.`, `// DefaultDates.Tolerance…`) | |
| Guards | `core/matchregistry/guard_test.go` `layerPackages` (`core/shortlist` row, no `core/graphalign` row), `checkPure`, `checkNumbers`; `inventory_test.go` `notYetBuilt` (`alignment.AcceptBar`, `alignment.Bands.`, `alignment.TopK`, `alignment.TieCap`, `alignment.Alternatives`, `alignment.SettleMargin`, `alignment.SettleRounds`, `evidence.`) | |
| Shortlist | `core/shortlist` `Rules`, `Dated`, `Keys`, `NewIndex(r, subject, value, metas) *Index`, `(*Index).Add(kind, id, v)`, `(*Index).Lookup(kind, v) [][]byte`; `core/database/promotealign/index.go` `matchIndex{shortlist}`, `matchIndexOf`, `loadShortlist(g, p, cfg)` (reads `cfg.Shortlist`) | Read by the loader only |
| Loader | `core/database/promotealign/promotealign.go` `Propose` → `propose` (`cfg := graphalign.DefaultConfig(); cfg.Value = rules`; `graphalign.Align(layer, canon, anchors, &cfg)`); `canon.go` `seedCandidateLimit`, `loadCanon(g, prof, layer, fixed, cfg, terms)` (for the unfixed Subjects with weights: `loadShortlist`, `seedHandles`, then `graphalign.Candidates(…, seedCandidateLimit)` per Subject, then the BFS); `exhibits.go` `attachExhibits(…, cfg graphalign.Config, …)` reads `cfg.Value`; `export_test.go` and `exhibits_pin_test.go` call `graphalign.DefaultConfig()` | |
| FFI | `api/ffi/handlers/promote_align.go` maps `Reason`, `Flags.ConflictWithFixed`, `PossibleDuplicate`, `DuplicateOf`, `Via` | Unchanged by this PR |
| Swift | `macos/App/Features/Promote/PromoteFlow.swift` `Reason(code:…)` (unknown → `.empty`), `conflictNote` from `conflictWithFixed`; `updated` is computed by diffing proposals | Unchanged |
| Tests | `core/graphalign/align_test.go` (40 tests, among them `TestGracieObituaryGolden`, `TestSparseNameAndBirthDate`, `TestContestedHandleIsDecidedTheSameEveryRun`, `TestDecidedRowTheNeighborsContradict`, `TestRefinementDropsSupportFromAMovedNeighbor`, `TestUnreachableFallbackScoresOnTheWalkScale`, `TestOrphanFallsBackToPropertyOnlyOrSkip`, `TestConfigOverrideWithoutEditingRegistry`); `candidates_test.go` `TestCandidates`; `pick_internal_test.go` `TestPickScoreScoresAMissingPick`; `core/segmentcompare/segmentcompare_test.go`; `core/database/promotealign/promotealign_test.go` (34 tests); `api/ffi/handlers/promote_align_test.go` | |
| Docs | `docs/promote-graph-alignment.md` §4.1, §5, §6, §12 (each under a "Superseded in Spike 11" line), §11; `docs/matching-model.md` stack row 6, worked example step 6, knob inventory `alignment.*`, "Removed knobs"; `docs/matching.md` (the `graphalign.Candidates` sentence); `docs/conclusion-reconciliation.md` L146 (`promote-graph-alignment.md` §6); `docs/conclusion-layer-data-model.md` L256; `comparison-layers.md` §Layer 6 procedure, §Segments; `.agents/skills/add-match-rule/SKILL.md` | |

**Precedent.** S11-14 ([#493](https://github.com/mendahu/provenencia/pull/493), [plan](S11-14-segment-compare.md)) rewrote part of the walk under a new layer, moved its numbers into a registry section, extended the skill and fixed the docs it made false in one PR: copy its step shape and its docs step. S11-02 ([#441](https://github.com/mendahu/provenencia/pull/441)) is the shape for a section whose type the layer declares and the registry fills (`segmentcompare.Rules` in `core/matchregistry/segment.go`). S11-13 ([#487](https://github.com/mendahu/provenencia/pull/487)) ported `align_test.go` across a scoring change: its Landed note says reshaped fixtures keep their intent; copy that, not new assertions.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | **The layer owns its section.** `core/graphalign/rules.go` *(new)*: `type Rules struct{ Shortlist shortlist.Rules; MinSimilarity float64; TopK, TieCap int; AcceptBar float64; Bands Bands; Alternatives int; SettleMargin float64; SettleRounds int }`, `type Bands struct{ Strong, Medium, Weak float64 }`, a doc comment per field (Shortlist "the keys the loader shortlists by"; MinSimilarity "the least similarity for a shortlisted handle to be a candidate"; TopK "candidates kept per Subject from its shortlist"; TieCap "most candidates kept when ties at the K-th place pass K"; AcceptBar "the least evidence to place a pair"; Bands "the evidence a row's band starts at"; Alternatives "alternatives listed per row"; SettleMargin "the lead a free rival needs over a placed pair to release it, and over a decided row's handle to flag it"; SettleRounds "most settle rounds"). `Config` *(rewritten)*: `{Value valuecompare.Rules; Subject subjectcompare.Rules; Segment segmentcompare.Rules; Alignment Rules}`, each doc "the registry's value / subject / segment / alignment section". `registry.go` and `DefaultConfig` are deleted. `Align(layer Layer, canon Canon, fixed []Fixed, cfg Config) Proposal` takes the config by value; there is no default. In `core/matchregistry/alignment.go` the `Alignment` type is deleted, `Registry.Alignment` becomes `graphalign.Rules` (doc "Alignment is layer 6: candidates, the accept bar, bands and settling"), and `defaultAlignment() graphalign.Rules` sets every field (D-2). **One conversion:** `func (r Registry) AlignConfig() graphalign.Config` *(new, `core/matchregistry/registry.go`)*, doc "AlignConfig is the registry as graph alignment reads it.", returns the four sections; the loader and every test (`promotealign_test.go`, `exhibits_pin_test.go`, `export_test.go`, `align_test.go`, `candidates_test.go`) call it, and the loader then sets `Value` to the catalog's rules | Working rule "Types live in the layer, values in the registry"; S11-15 D-10 left the type in the registry only until `registry.go` goes; with it gone `graphalign` no longer imports `matchregistry`, so the registry can name `graphalign.Rules` as it names `segmentcompare.Rules`, and owns the one place that lists the sections (review M-3) | A `DefaultConfig` kept in `graphalign` — it must import the registry, which then cannot name the layer's type (cycle); a `config` builder in `promotealign` plus one per test package — four copies of the field list |
| D-2 | **Numbers.** `defaultAlignment`: `MinSimilarity: 0.5` (`// estimate: at least half alike; a name with a sex conflict is 0.71, a birth against a death 0.09`), `TopK: 5` (`// carried from promotealign seedCandidateLimit`), `TieCap: 20` (`// estimate: room for a common name's ties without unbounded work`), `AcceptBar: 2` (`// carried from graphalign AcceptScore`), `Bands{Strong: 4, Medium: 3, Weak: 2}` (`// carried from graphalign StrongScore, MediumScore, WeakScore`), `Alternatives: 3` (`// carried from graphalign AlternativeLimit`), `SettleMargin: 1` (`// estimate: one point, a third of a band`), `SettleRounds: 10` (`// estimate: a safety cap; dirty pairs normally settle in two or three`). Inventory rows (`docs/matching-model.md`): the existing `alignment.AcceptBar`, `alignment.Bands.*`, `alignment.TopK`, `alignment.TieCap`, `alignment.Alternatives`, `alignment.SettleMargin`, `alignment.SettleRounds` keep their starts; `alignment.MinSimilarity` *(new row)* 0.5 with the estimate text; `segment.UnplacedAlike` *(new row)* 2 (D-3). `notYetBuilt` loses every `alignment.` prefix | Working rule "Registry numbers: carry over, then estimate"; the inventory already holds every number except the similarity floor (the user's call of 2026-10-19: gate on similarity) and the unplaced bar | Gating on evidence ≥ the weak band (S11-13 D-4) — superseded by the user's 2026-10-19 call |
| D-3 | **Unplaced neighbors (layer 5).** `segmentcompare.Unplaced Placement = "unplaced"` *(new)*, doc "the layer neighbor is placed nowhere yet: entity compare decides". `Rules.UnplacedAlike float64` *(new)*, doc "the least neighbor evidence for an unplaced neighbor to count as lined up"; registry `UnplacedAlike: 2` (`// estimate: the accept bar, a neighbor that could be placed on its own properties`). Causes *(new)*: `CauseUnplacedDiffers = "unplaced_differs"`, `CauseUnplacedUnsure = "unplaced_unsure"`. `Compare` with `Unplaced`, after the correspondence check: `Neighbor = subjectcompare.Compare(…)`; when `Neighbor.Evidence >= rules.UnplacedAlike` it continues exactly as `Together` (gate, support, credit, outcome); when `Neighbor.Evidence < 0` it scores as `Apart` (One loses `OneConflict`, a gate loses `ConflictFactor × own`, outcome conflict when either applies, else unknown) with cause `unplaced_differs` and `Neighbor` kept; otherwise outcome unknown, cause `unplaced_unsure`, evidence 0. `Compare`'s doc gains one sentence for it. The apart scoring becomes one local helper, `apart(one bool, g Gate, own float64, rules Rules) Result`, that `Apart` and unplaced-differs both call, the caller setting `Cause` and `Neighbor`; `Result.Neighbor`'s doc becomes "layer 3's verdict on the neighbors; zero when placed apart". **The one rule for an unplaced neighbor:** `graphalign` `newPairSegments` compares a different neighbor pair with `Unplaced` (the same Subject stays `Together`) and counts it as lined up when the outcome is agree or partial; its `res.Neighbor.Evidence < AcceptScore` gate goes, so `alikeNew` reads `UnplacedAlike` through layer 5 | The scope entry ("segment compare learns unplaced neighbors, entity compare decides"); S11-14 Next adds. Once the neighbor is placed on that research neighbor, `Together` gives the same number, so a score does not depend on placement order (layer 6 procedure step 3) | Scoring an unplaced neighbor as `Together` — a differing birth would add 1.5 + 0.5 × −2.5 = 0.25 instead of subtracting; a third placement decided in `graphalign` — layer 5's rule outside layer 5 |
| D-4 | **Candidates.** `Candidates(kind string, values subjectcompare.Values, handles []Handle, metas subjectcompare.Metas, cfg Config) []Handle` (the `limit` parameter goes): entity compare each handle of `kind`; keep those with `Similarity >= cfg.Alignment.MinSimilarity`; sort by similarity (high first), evidence (high first), `Ref`, `ID` bytes; keep the first `TopK`, then every following handle whose similarity equals the `TopK`-th's, at most `TieCap` in all. Doc: "Candidates are the shortlisted handles a Subject is compared with: similarity at least MinSimilarity, the best TopK by similarity, and every handle tied with the TopK-th, up to TieCap. Evidence decides everything after." Nominated handles never pass through it | [Bands](../comparison-layers.md#bands): "similarity gates and roughly sorts; it never decides"; ties at the cutoff are kept so a sparse match isn't cut before its links are compared. One function for the loader's seeds and the walk, as S11-13 D-4 set up | Ties by evidence — similarity is the design's first sort; no cap — 30 John Smiths × their neighborhoods |
| D-5 | **The walk.** `Align` builds `state` (D-6), then: (1) **Anchors.** Each fixed handle pair, in `sortedSubjects` order, is placed (`place`, D-7), and its held handle is scored as its first candidate. (2) **Candidates.** `Canon.Candidates map[string][][]byte` *(new field, `types.go`)*, doc "Candidates are each Subject's shortlisted candidates, by Subject id, in `Candidates` order; the loader fills them": every Subject neither placed nor held whose kind has `cfg.Subject.Weights` gets the ids listed for it that are canon handles; each is scored and pushed. Align runs no shortlist lookup and no `Candidates` call of its own. (3) **Queue.** `walk` pops best first (score, then handle `Ref`, handle id, subject id, as `candHeap` orders today). An entry whose version is stale, or whose Subject is placed or held, is skipped. An entry under `AcceptBar` is skipped. An entry whose handle another Subject holds records that owner in `lostTo` (first one only, as `lostContest`) and is skipped. Otherwise the pair is placed. (4) **Settle** (D-7) when the queue is empty. (5) **Rows** (D-8). There is no fallback pass, no second walk and no refinement | [Layer 6 procedure](../comparison-layers.md#layer-6-procedure) steps 1–5; an isolated Subject is a candidate that never gets link support. One shortlist and one `Candidates` run per Subject, in the loader, which needs the top K anyway to choose what to load (review M-1) | Align building its own index over the canon — a second shortlisting path whose answer can differ from the loader's, rebuilt every run (review M-1, X-4) |
| D-6 | **State and scoring.** `state` keeps `subjects`, `handles`, `layerAdj`, `canonAdj`, `held`, `lostTo`, `fixedSet` as today; `assigned` / `handleOwner` are renamed `placed` / `owner`; `best` and `seen` are replaced by `cands map[string]map[string]*cand` (subject → handle → its latest score) with `cand{handleID []byte; ref string; score, edge float64; result subjectcompare.Result; links []segmentcompare.Result; via []byte; viaKey vocabkey.BridgeKey; version int}`; a queue item carries the `version` it was pushed at. `score(s, h)` is `evaluate`, then `links := st.segments(s, h, own, st.placedOf)`, `edge` their summed evidence, then `ScoreCandidate`; it bumps the version. `segments(subjectID []byte, h *Handle, own float64, placedOf func(subjectID []byte) ([]byte, bool)) []segmentcompare.Result` is the one placement rule: for each layer link of the Subject to neighbor `n` (skipping a held `n`), over the canon links of `h` that `Corresponds`: when `placedOf(n)` is the canon neighbor → `Together`; else when `n` is placed elsewhere → `Apart` on the first corresponding link; else (n unplaced) → `Unplaced` on each corresponding link whose canon neighbor no other Subject holds, keeping the one result with the most evidence (an alike link beats a differing one); none → no result for that link. One result per layer link that has one | One record per pair, replaced on every re-score, so a stale score never outlives its neighbors (today's `recordBest` keeps the higher); "skip stale queue entries by version" (procedure step 4); a differing unplaced birth conflicts only when no corresponding research neighbor lines up. The placement rule takes its lookup as an argument and the per-link results stay on the candidate, so contested (D-8) and S11-19's relationship lines read the same results (review M-5) | Keeping `recordBest`'s "never lower" rule — it is why refinement had to replace scores outright; keeping only the summed edge — S11-19 would copy the placement rule |
| D-7 | **Place, release, settle.** `place(s, h)`: `placed[s] = h`, `owner[h] = s`; for each layer neighbor `n` of `s` that is not held: every canon neighbor `g` of `h` through a corresponding link, of `n`'s kind and not yet a candidate of `n`, becomes one (nominated: `via = s`, `viaKey` = the layer link's key, any score, no K); one that is already a candidate with no `via` gets `via = s` and `viaKey`, and an existing `via` is kept — for a fixed `n` too, so a decided row has the handles its neighbors point at (D-8); then, unless `n` is fixed, if `n` is placed it is marked dirty with reason `neighbor_placed` (or `neighbor_moved` when `s` was placed on another handle earlier in this run), else every candidate of `n` is re-scored and pushed. A Subject's dirty mark keeps the first reason and neighbor it got; later marks before the round clears it change nothing. `release(s)`: deletes `placed[s]` and `owner`, sets `releasedFrom[s] = h` and `move[s] = Move{its dirty reason, its dirty neighbor}` (each overwrites an earlier release in this run), marks each placed neighbor dirty `neighbor_released`, re-scores and pushes every candidate of each unplaced neighbor and of `s`. `settle`: up to `SettleRounds` rounds while the dirty set is not empty: take the dirty Subjects in `sortedSubjects` order with their marks and clear the set; for each still placed and not fixed, re-score all its candidates and release it when its placed pair scores under `AcceptBar`, or when a candidate whose handle no Subject holds scores more than its placed pair plus `SettleMargin`; every Subject of the round is checked before `walk` runs once. After the last round every candidate of every Subject that is not held, fixed ones included, is re-scored once more, without moving anything, so rows read the mapping they end on. A released Subject may be placed again on the handle it was released from; `Row.Move` is then nil. A fixed Subject is never released | Procedure step 5: only dirty pairs are re-scored, changes ripple as far as they reach, the round cap is the safety net, and each move has a reason | Re-scoring every candidate each round (`refineOneHop`) — the cost the design removes; displacing a rival's placed pair — one-to-one contests go through the queue |
| D-8 | **Rows.** `proposal` reads `cands` instead of `best` / `pickScore`; `alts` and `bestNonAssigned` read the candidate list sorted by score, `Ref`, `ID`. A fixed handle row's candidates are its held handle and the handles `place` nominated for it (D-5, D-7), as `scoreFixed` gathers them today. Bands: `band(score)` against `cfg.Alignment.Bands` (strong ≥ Strong, medium ≥ Medium, weak ≥ Weak, else none). Alternatives: at most `cfg.Alignment.Alternatives`, the row's handle excluded; the `limit = 3` fallback goes. `via` and `Row.Via` stay while the nominating neighbor still `supports` the pair, as `rescoreBest` checks today. **`subjectcompare.Result.Conflicts() bool`** *(new, `core/subjectcompare/subjectcompare.go`)*, doc "Conflicts reports whether any property comparison conflicts."; `segmentcompare.Compare`'s partial check calls it. **`RowFlags.Contested bool`** *(new)*, doc "Contested: a comparison behind the pair this row reports conflicts, whatever its band": the reported pair is the row's handle (decided rows included), or `Alternatives[0]` on a Skip row the walk proposed (reason `weak` or `taken`); it is true when the pair's `result.Conflicts()` or any of its `links` has outcome conflict. A New row, a row the researcher decided New or Skip, and an empty row are never contested. **`ConflictWithFixed`** stays with its meaning and wire field: a fixed handle row whose best other candidate leads the held handle's score by more than `SettleMargin`; `ConflictPenalty` goes. Reasons are unchanged (`via`, `decided`, `agrees`, `weak`, `taken`, `no_match`, `empty`) | [Bands](../comparison-layers.md#bands): bands are a pure evidence cutoff, contested is a separate mark; the matching model already names the settle margin as `ConflictPenalty`'s replacement ("Removed knobs"); a decided row's note and a conflict are two facts, and the page already shows the first. One definition of "this comparison conflicts" (review M-2) | Folding `ConflictWithFixed` into contested — a held row its spouse contradicts has no conflict (an Other link apart is unknown), so the note would vanish until S11-18; a contested flag on every alternative — S11-18's board decides; a second per-property conflict loop in `graphalign` |
| D-9 | **Move reasons.** `type Move struct{ Reason Reason; NeighborSubjectID []byte }` *(new)*, doc "Move is why settling released this row's earlier pair: a neighbor was placed, moved or released"; `Row.Move *Move` *(new)*, set when `releasedFrom[s]` exists and the row's final target is not that handle. Reason codes *(new)*: `ReasonNeighborPlaced = "neighbor_placed"`, `ReasonNeighborMoved = "neighbor_moved"`, `ReasonNeighborReleased = "neighbor_released"`, documented as move reasons. `Row.Reason` keeps its meaning. The handler does not send `Move` (S11-17) | Working rule "One wire vocabulary": move reasons extend `Reason`; procedure step 5 "each move has a reason" | Overwriting `Row.Reason` with the move — loses `via` / `agrees`, which the page words today |
| D-10 | **The loader.** `Propose` does `cfg := matchregistry.Default().AlignConfig(); cfg.Value = rules` and passes it by value to `Align`. `loadCanon` keeps its one `graphalign.Candidates(s.Kind, s.Values, handles, layer.Metas, cfg)` call per unfixed Subject over its looked-up handles, and stores each result's ids, in order, in `Canon.Candidates[string(s.ID)]`; `seedCandidateLimit` is deleted. The BFS, profiles and exhibits are unchanged | The loader seeds the canon from the candidates the walk scores, so every candidate is in the canon with its neighborhood, and both read one list | Passing the cached index to Align — Align would run `Candidates` a second time |
| D-11 | **Guards.** `layerPackages` gains `{"core/graphalign", true}`. `checkPure` *(extended)* takes the row's `numbers` flag; the walk keeps today's rule for `core/database` (every hop walks `Imports`, `TestImports` and `XTestImports`), and adds two rules that look only at non-test `Imports`, at the root and at every hop reached through non-test imports: `core/matchregistry` is reported when `numbers` is true, and `core/autoreconcile` when the row's dir is not `core/subjectprofile`, each with the import chain. Layer tests that import the registry stay legal. `flagSharedHandles`' `if len(idx) < 2` becomes `if len(idx) <= 1` (the guard allows only 0 and 1). The skill's rule "never imports `core/matchregistry`" names `TestLayerPackages`. `.cursor/rules/match-rules.mdc` globs gain `core/graphalign/**` | S11-11 M-9 logged it for this PR, which removes the one exception; one walker (working rule "One guard", review M-4) | A second test with its own direct-import scan — two walkers to keep in step |
| D-12 | **Invariants.** `TestAlignInvariants` *(new)*, in `core/graphalign/invariants_test.go`, table over three fixtures — the obituary golden (moved out of `TestGracieObituaryGolden` into `gracieObituary() (graphalign.Layer, graphalign.Canon)`), AC-1.1's, and `TestPartOfChainIsStrongWithoutADecision`'s (moved into `partOfChain()`) — asserts: (a) reversing `Subjects`, `Bridges`, `Handles` and `Edges` gives an identical `Proposal` (`reflect.DeepEqual` after sorting rows by `SubjectID`); (b) a second `Align` is identical; (c) for each row of the undecided proposal with target handle, fixing that row on its handle gives every other row the same target, handle and band; (d) every fixed row's target and handle are what was fixed | Procedure "Invariants"; the slice Check ("the same proposal whether or not a row starts decided") | A random-graph generator — deferred with the test suite (working rule "The test suite is deferred") |
| D-13 | **Ported tests keep their intent.** A test of a deleted pass is rewritten to the behavior it pinned, under a name that says it: `TestRefinementDropsSupportFromAMovedNeighbor` → `TestSettlingDropsSupportFromAMovedNeighbor`, `TestUnreachableFallbackScoresOnTheWalkScale` → `TestIsolatedSubjectScoresOnTheWalkScale`, `TestOrphanFallsBackToPropertyOnlyOrSkip` → `TestOrphanMatchesOnItsPropertiesOrSkips`, `TestContestedHandleIsDecidedTheSameEveryRun` → `TestSharedHandleIsDecidedTheSameEveryRun` ("contested" now names the conflict mark); `TestConfigOverrideWithoutEditingRegistry` overrides `cfg.Alignment.AcceptBar`; `TestCandidates` is rewritten for D-4; `pick_internal_test.go` is deleted with `pickScore`. Every `graphalign.Align(…, nil)` and `DefaultConfig()` reads `R.AlignConfig()`, and every `Align` test fills `Canon.Candidates` with `withCandidates(layer graphalign.Layer, canon graphalign.Canon, cfg graphalign.Config) graphalign.Canon`, which returns `canon` with `Candidates` set. An expected value that changes in STEP-3 is legitimate only when the handle's similarity is under 0.5 or it was cut or kept by the K / tie rule. An expected score, band or target may change only where D-3, D-4, D-7 or D-8 makes it change; each change is listed in the PR body and the Landed note with its D-#. `TestSparseNameAndBirthDate`, `TestBirthIsNeverProposedOnADeath`, `TestLifespanRulesOutAPerson`, `TestGracieObituaryGolden`, `TestPartOfChainIsStrongWithoutADecision` and every `core/database/promotealign` and `api/ffi/handlers` Promote test keep their expected targets, bands, reasons and `Via` | Working rule "goldens are updated, not preserved", bounded so a sparse-data regression cannot slip in as an update | Rewriting `align_test.go` from scratch — loses 39 pinned behaviors |
| D-14 | **Docs and skill.** `docs/promote-graph-alignment.md`: the top banner and the four "Superseded" lines go; §4.1 shrinks to a paragraph saying `graphalign` composes the layers and reads the registry's alignment section (link [`matching-model.md`](../../../matching-model.md)); §5 is rewritten to the layer 6 procedure as built (shortlist or nominate, candidates, one queue, unplaced neighbors, settle with move reasons, the round cap); §6 to the segment rules with `Unplaced`, linking the model for weights; §11's thresholds bullet names `core/matchregistry/alignment.go`; §12 lists what differs from §5 (provenance 1). `docs/conclusion-reconciliation.md` L146: "§6" becomes a link to [`matching-model.md`](../../../matching-model.md); `docs/conclusion-layer-data-model.md` L256 also links the model for scoring. `docs/matching.md`: the `Candidates` sentence becomes "A handle is a candidate when it is shortlisted and at least half alike (`graphalign.Candidates`), or reached through a link to an accepted pair." `docs/matching-model.md`: D-2's inventory rows; worked example step 6 adds "Its linked birth was not placed yet when the pair was first scored; entity compare found it alike, so the link counted from the start."; "Removed knobs" gains `graphalign` `DefaultConfig`, `seedCandidateLimit` (→ `alignment.TopK`), `AlternativeLimit` (→ `alignment.Alternatives`), and the rows for `ConflictPenalty` keep their text. `comparison-layers.md` layer 6 procedure: a "**Built in S11-16**" paragraph naming `Align`, `Candidates`, `place`, `settle`, `Move` and the three move reasons; §Segments output list gains the two causes and `Unplaced`; the bridge bullet's "compared in S11-16 if a golden needs them" becomes "not compared: no golden needs them". Skill: description and Check list `core/graphalign`; a rule bullet "**Layer 6 (`core/graphalign`) walks and decides.** Candidates come from the shortlist through `Candidates` or from a link to a placed pair; scores are replaced, never kept higher; placement is `Together`, `Apart` or `Unplaced`; bands, contested and moves are row output. Its section is `graphalign.Rules`"; "Add a knob" step 1 names the layer's own `Rules` file (`core/graphalign/rules.go` *(new)* for a walk knob); "Add a relation, cause or unit code" gains "a move reason is layer 6's: a `Reason` constant in `core/graphalign/types.go`". `core/matchregistry/value.go` (MNT-3): every comment that `rg -n 'match\.|Default(Dates|Names|Text)|scale' core/matchregistry/value.go` finds naming an identifier of `core/match` or `graphalign` that no longer exists reads "carried from <identifier>", the form `segment.go` uses; in `core/matchregistry/subject.go` the two "; ConflictPenalty" endings read "; carried from graphalign ConflictPenalty" | Working rule "One matching skill, grown with the layers"; S11-01 D-5; MNT-3 | — |
| D-15 | **Bridge properties are not compared.** `Segment` gains no bridge values | The scope entry moved them here "if a golden needs them"; no golden or AC compares a bridge's dates, and no inventory weight exists for one (S11-14 D-10) | Plumbing dates with no weight — a path no test can show working |

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `graphalign.Rules`, `Bands`, `core/graphalign/rules.go` *(new)* | type, file | `segmentcompare.Rules`; `matchregistry.Alignment` (S11-15, deleted here); `graphalign.Config` fields | The layer declares its section (D-1); `Config` keeps only the four sections |
| `segmentcompare.Unplaced`, `Rules.UnplacedAlike`, `CauseUnplacedDiffers`, `CauseUnplacedUnsure` *(new)* | const, field | `Together`, `Apart`, `CausePlacedApart` | A third reading of the same link; its rule is layer 5's (D-3) |
| `cand`, `place`, `release`, `settle`, `score` in `align.go` *(rewritten)* | func, type | `scoredCand`, `accept`, `refineOneHop`, `scoreCandidate` | The walk is rewritten in place (working rule "rewritten in place in S11-16"); names follow the glossary |
| `RowFlags.Contested`, `Move`, `Row.Move`, `ReasonNeighborPlaced`, `ReasonNeighborMoved`, `ReasonNeighborReleased` *(new)* | field, type, const | `RowFlags`, `Reason`, `Via` | Working rule "One wire vocabulary" (D-8, D-9) |
| `core/graphalign/rows.go` *(new)* | file | `align.go` | `proposal` and its helpers move out of the 983-line file unchanged except D-8; `align.go` keeps the walk |
| `matchregistry.Registry.AlignConfig` *(new)* | method | `graphalign.DefaultConfig` (deleted) | The one place that turns the registry into a `Config`; the registry can name `graphalign` once `registry.go` is gone (D-1) |
| `Canon.Candidates` *(new)* | field | `promotealign` `loadCanon`'s `Candidates` call; S11-15's index | The loader's candidate lists travel to Align instead of Align shortlisting again (D-5, D-10) |
| `subjectcompare.Result.Conflicts` *(new)* | method | `segmentcompare.Compare`'s partial loop | One definition of a conflicting comparison; that loop calls it (D-8) |
| `checkPure`'s two forbidden imports *(extended)* | test | `TestLayerPackages` `checkPure` | Same walker, two more rules (D-11) |
| `TestAlignInvariants`, `core/graphalign/invariants_test.go`, `gracieObituary`, `partOfChain` *(new)* | test, helper | `TestContestedHandleIsDecidedTheSameEveryRun` | It checks one fixture's determinism; D-12 checks four invariants on three |
| `withCandidates` helper in `align_test.go` *(new)* | test helper | `graphalign.Candidates` | Fills `Canon.Candidates` the way the loader does, over all fixture handles (D-5) |

**Builds on.** `segmentcompare.Compare` and `Corresponds`; `subjectcompare.Compare`; S11-15's `shortlist.Index`; `ScoreCandidate`; `candHeap`; the registry and its two guards; `flagSharedHandles` / `flagAlikeNewRows`.
**Deletes.** `core/graphalign/registry.go` (`DefaultConfig`, `AcceptScore`, `StrongScore`, `MediumScore`, `WeakScore`, `ConflictPenalty`, `AlternativeLimit`); `align.go` `seed`, `fallbackUnreachable`, `assignedIDs`, `propagateNew`, `refineOneHop`, `rescoreBest`, `reassignUnfixed`, `scoreFixed` (folded into D-8), `handlesOfKind`, `kindCands`, `recordBest`, `pickScore`, `seen`, `pushFromAnchor`, `enqueue` (both become `place`), `accept` (becomes `walk`'s pop rule), `scoredCand` and `sortCands` (become `cand` and the candidate sort), `scoreCandidate` (renamed `score`); `lostContest` stays as the helper that records `lostTo`; `queueItem` gains `version`; `pick_internal_test.go`; `matchregistry.Alignment` (S11-15's type); `promotealign` `seedCandidateLimit`; the `alignment.` prefixes in `notYetBuilt`; the "Superseded" lines and superseded text in `promote-graph-alignment.md`.
**Next adds.** S11-17 sends `Flags.Contested` and `Row.Move` over `engine.proto` and parses the three move reasons in `ConclusionVocabulary.swift`. S11-19 calls `segments` with a lookup over row targets (exporting it then) and reads the same per-link results. S11-21 memoizes `score` per pair if its budget fails. A new walk knob is one `graphalign.Rules` field, one registry value and one inventory row. Bridge-property comparison, if a golden needs it, is a `Segment` field and subject weights for the bridge type.

### Schema

None — no catalog change.

### Go core

| Package | Change |
| --- | --- |
| `core/segmentcompare` | D-3 |
| `core/graphalign` | `rules.go` *(new)* (D-1); `types.go`: `RowFlags.Contested`, `Move`, `Row.Move`, three `Reason` constants (D-8, D-9); `candidates.go` (D-4); `align.go` rewritten (D-5 – D-7); `rows.go` *(new)* (D-8); `registry.go`, `pick_internal_test.go` deleted |
| `core/matchregistry` | `alignment.go` (D-1, D-2); `segment.go` `UnplacedAlike`; `registry.go` field type; `value.go` comments (D-14); `guard_test.go`, `inventory_test.go` (D-11, D-2) |
| `core/database/promotealign` | `promotealign.go` `config` (D-10); `canon.go` (D-10) |

```go
// core/graphalign/rules.go (new)
type Rules struct {
	Shortlist     shortlist.Rules
	MinSimilarity float64
	TopK, TieCap  int
	AcceptBar     float64
	Bands         Bands
	Alternatives  int
	SettleMargin  float64
	SettleRounds  int
}

func Align(layer Layer, canon Canon, fixed []Fixed, cfg Config) Proposal
func Candidates(kind string, values subjectcompare.Values, handles []Handle, metas subjectcompare.Metas, cfg Config) []Handle
```

Errors: none — `Align` is total.

### FFI

None — the handler reads the same fields; `Flags.Contested` and `Row.Move` reach the wire in S11-17.

### macOS

None — no Swift change.

### Strings

None — no copy.

## 7. Steps

**STEP-1 — Unplaced neighbors in segment compare**
Files: `core/segmentcompare/segmentcompare.go`, `segmentcompare_test.go`, `core/subjectcompare/subjectcompare.go`, `core/matchregistry/segment.go`, `docs/matching-model.md` (one row). Skill: `add-match-rule` "Add a knob", "Add a relation, cause or unit code".
Do: D-3, and D-8's `subjectcompare.Result.Conflicts`, which `Compare`'s partial check calls. T-1.
Check: `go test ./core/segmentcompare/... ./core/matchregistry/...` passes.

**STEP-2 — The alignment section belongs to the layer**
Files: `core/graphalign/rules.go` *(new)*, `registry.go` (deleted), `types.go` (`Canon.Candidates`), `align.go`, `candidates.go`, `align_test.go`, `candidates_test.go`, `pick_internal_test.go`, `core/matchregistry/alignment.go`, `registry.go`, `guard_test.go`, `inventory_test.go`, `core/database/promotealign/promotealign.go`, `canon.go`, `layer.go`, `exhibits.go`, `index.go` (`loadShortlist` reads `cfg.Alignment.Shortlist`), `export_test.go`, `exhibits_pin_test.go`, `promotealign_test.go`, `docs/matching-model.md`, `.cursor/rules/match-rules.mdc`. Skill: `add-match-rule` "Add a layer package", "Add a knob".
Do: D-1, D-2, D-10's config line, D-11. Until STEP-4 the old walk reads `Canon.Candidates` nowhere and keeps running on renamed fields (`cfg.Alignment.AcceptBar` for `AcceptScore`, `Bands.*` for the three scores, `Alternatives` for `AlternativeLimit`, `SettleMargin` for `ConflictPenalty`; `alikeNew` and `newPairSegments` read `cfg.Alignment.AcceptBar` until STEP-4). Delete `pick_internal_test.go` here: an internal test cannot import `core/matchregistry` (cycle), and STEP-4 deletes `pickScore`. T-2, T-3.
Check: `go test ./core/graphalign/... ./core/matchregistry/...` and `CGO_ENABLED=1 go test -tags fts5 ./core/database/promotealign/ ./api/ffi/handlers/` pass; `rg -n 'matchregistry' core/graphalign --glob '!*_test.go'` prints nothing.

**STEP-3 — Candidates gate on similarity and keep ties**
Files: `core/graphalign/candidates.go`, `candidates_test.go`, `align.go` (`fallbackUnreachable` calls `Candidates(s.Kind, s.Values, st.handlesOfKind(s.Kind), st.metas, st.cfg)`), `core/database/promotealign/canon.go`.
Do: D-4; the loader stores `Canon.Candidates` and `seedCandidateLimit` goes (D-10). T-4.
Check: `go test ./core/graphalign/...` passes with `TestCandidates` rewritten; `CGO_ENABLED=1 go test -tags fts5 ./core/database/promotealign/` passes; a changed expectation follows D-13's STEP-3 rule and is listed in the PR body.

**STEP-4 — One queue and settling**
Files: `core/graphalign/align.go`, `types.go` (`Move`, reasons), `rows.go` *(new)* (moved code), `align_test.go`, Skill: `add-match-rule`.
Do: D-5, D-6, D-7, D-9, D-3's `newPairSegments` change, and D-13's ports. Move `proposal` and its helpers to `rows.go` first, in their own commit, with no change. T-5 – T-8.
Check: `go test ./core/graphalign/...` passes; `rg -n 'fallbackUnreachable|refineOneHop|rescoreBest|reassignUnfixed|propagateNew|handlesOfKind|recordBest|pickScore|pushFromAnchor|enqueue|scoredCand|sortCands|scoreCandidate|func \(st \*state\) accept' core/graphalign` prints nothing.

**STEP-5 — Bands, contested and the decided-row note**
Files: `core/graphalign/rows.go` *(new)* in STEP-4, `types.go`, `align_test.go`.
Do: D-8. T-9, T-10.
Check: `go test ./core/graphalign/...` passes; AC-1.3 and AC-3.5 hold.

**STEP-6 — Invariants**
Files: `core/graphalign/invariants_test.go` *(new)*, `align_test.go` (fixtures moved into `gracieObituary`, `partOfChain`).
Do: D-12. T-11.
Check: `go test -count=5 ./core/graphalign/ -run TestAlignInvariants` passes.

**STEP-7 — Promote end to end**
Files: `core/database/promotealign/*_test.go`, `api/ffi/handlers/promote_align_test.go`.
Do: run the Promote suites; any changed expectation follows D-13. Add T-13. T-12, T-13.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/promotealign/ ./api/ffi/handlers/` passes; AC-4.4's `rg` prints nothing; `CGO_ENABLED=1 go test -tags fts5 -run '^$' -bench BenchmarkProposeColdHub -benchmem ./core/database/promotealign/` on `main` and on the branch, both in the PR body.

**STEP-8 — Docs and skill**
Files: `docs/promote-graph-alignment.md`, `docs/matching-model.md`, `docs/matching.md`, `docs/conclusion-reconciliation.md`, `docs/conclusion-layer-data-model.md`, `docs/deployment-plan/spike-11/comparison-layers.md`, `.agents/skills/add-match-rule/SKILL.md`, `core/matchregistry/value.go`, `core/matchregistry/subject.go`, `docs/deployment-plan/spike-11/completed.md`, `docs/deployment-plan/spike-11/deployment-plan.md` (map row), this plan (status).
Do: D-14; the Definition of done's docs lines.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-11/prs/S11-16-graph-alignment.md` passes; `rg -n 'Superseded in Spike 11' docs/promote-graph-alignment.md` prints nothing; `scripts/check-pr.sh` prints `check-pr: PASS`.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Judge a link whose neighbor is not placed yet by entity compare, so a differing birth counts from the start."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/segmentcompare/segmentcompare_test.go` `TestCompare` (rows added) | One birth link, `Unplaced`: neighbor birth 1817 vs 1817 (2.5) → agree, 1.5 + 0.5 × 2.5 = 2.75; vs 1830 (−2.5) → conflict, −3, `unplaced_differs`; vs undated birth (0.5) → unknown, 0, `unplaced_unsure`; spouse link vs a different name (−3) → unknown, 0, `unplaced_differs`; `part_of` from end, own 1, neighbor alike → unknown, `gated` | D-3, AC-1.1 |
| T-2 | Go | `core/matchregistry/inventory_test.go` `TestKnobInventoryMatchesRegistry`; `registry_test.go` `TestAlignmentSection` *(new)* | Inventory and registry agree; `TopK ≥ 1`, `TieCap ≥ TopK`, `Bands.Weak ≤ Medium ≤ Strong`, `AcceptBar > 0`, `SettleRounds ≥ 1` | AC-4.1, D-2 |
| T-3 | Go | `core/matchregistry/guard_test.go` `TestLayerPackages` (extended `checkPure`) | `core/graphalign` pure and literal-free; no layer reaches the registry; only `core/subjectprofile` reaches `core/autoreconcile` | AC-4.2, AC-4.3 |
| T-4 | Go | `core/graphalign/candidates_test.go` `TestCandidates` (rewritten) | AC-2.3 → 7; AC-2.4 → 20 by `Ref`; AC-2.5 → the birth; a name with a sex conflict (0.71) is kept | AC-2.3 – AC-2.5, D-4 |
| T-5 | Go | `core/graphalign/align_test.go` `TestSparseNameAndBirthDate` (extended) | AC-1.1 numbers, PER-2 alternative at −0.5 | AC-1.1 |
| T-6 | Go | `align_test.go` `TestOnlyCandidatesOrNominatedHandlesAreCompared` *(new)* | AC-2.1 → New `no_match`; AC-2.2 → PER-1, `via`; AC-2.6 → Skip `weak` | AC-2.1, AC-2.2, AC-2.6 |
| T-7 | Go | `align_test.go` `TestSettlingMovesAPairAndSaysWhy` *(new)* | AC-3.1, AC-3.2, AC-3.3 | AC-3.1 – AC-3.3, D-7, D-9 |
| T-8 | Go | `align_test.go` ported tests (D-13) | Each keeps its intent; renamed ones per D-13 | D-13 |
| T-9 | Go | `align_test.go` `TestContestedRow` *(new)* | AC-1.3; a New row and a held Skip row are not contested | AC-1.3, D-8 |
| T-10 | Go | `align_test.go` `TestDecidedRowTheNeighborsContradict` (fixture unchanged) | AC-3.5 | AC-3.5, D-8 |
| T-11 | Go | `core/graphalign/invariants_test.go` `TestAlignInvariants` *(new)* | D-12 (a)–(d) on three fixtures | AC-3.4, D-12 |
| T-12 | Go | `core/database/promotealign/promotealign_test.go`, `api/ffi/handlers/promote_align_test.go` (all) | Pass; targets and bands unchanged | AC-1.2, D-10, D-13 |
| T-13 | Go | `core/database/promotealign/promotealign_test.go` `TestProposeMarksAConflictingRowContested` *(new)*, `TestProposeDoesNotCompareAnUnshortlistedHandle` *(new)* | AC-1.3's people and births filed in the catalog → the person row `Flags.Contested` true, strong; a promoted \"Robins\" and a Source person \"Robnis\" with no bridge → New | AC-1.3, AC-2.1, D-5 |

AC-1.2 is asserted by T-11 (c) on AC-1.1's fixture.
AC-4.4 is the `rg` in STEP-7's check.

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `TestGracieObituaryGolden`, `TestPartOfChainIsStrongWithoutADecision`, `TestPartOfBreaksASharedToponym`, `TestHubPlaceProposalStaysOnTheLayer`, `TestProposeQueryCountDoesNotGrowWithTheLayer`, `TestKnobInventoryMatchesRegistry`.

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left — none in this PR
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New — no UI in this PR
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together) — or none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan, and every expected value D-13 let change
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated (if gated) — not gated
- [ ] Domain doc updated if a rule in docs/*.md changed: promote-graph-alignment.md, matching-model.md, matching.md, conclusion-reconciliation.md, conclusion-layer-data-model.md, comparison-layers.md, the add-match-rule skill (D-14)
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

1. Open a project with a promoted person "Ann Ames" whose birth is filed as 1817, and a second promoted "Ann Ames" born 1830. Add a Source with person "Ann Ames" and her birth in 1817. Open Promote on the person. See: the person on the 1817 Ann Ames, strong; the birth on her birth, strong; the 1830 Ann Ames in the menu below the bar.
2. On the same page, set the birth row to the 1817 birth (decided). See: no other row changes target or band.
3. Add a Source whose person "Ann Ames" is recorded male and born 1817, against a filed female Ann Ames born 1817. Open Promote. See: the row proposes her, strong. The contested mark arrives in S11-18; until then `TestProposeMarksAConflictingRowContested` (T-13) is the evidence, named in the PR body.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| Unplaced compares and the larger candidate sets (ties to 20) cost more per Propose | The PR body reports `BenchmarkProposeColdHub` on `main` and on the branch (STEP-7); S11-21 sets the budget |
| Settling oscillates between two pairs | `SettleRounds` caps it; AC-3.2 (T-7) pins the cap |
| A ported golden changes band | D-13: goldens keep their targets and bands, or stop |

**Stop and report if:**

- `TestSparseNameAndBirthDate`, `TestBirthIsNeverProposedOnADeath`, `TestLifespanRulesOutAPerson`, `TestGracieObituaryGolden` or `TestPartOfChainIsStrongWithoutADecision` needs a different target or band to pass.
- `TestAlignInvariants` (c) fails on a fixture after D-7 is implemented as written: report the fixture, the row, and the two proposals.
- `core/graphalign` cannot drop its `core/matchregistry` import without a cycle (D-1).
- A Promote handler test or a Swift test needs a wire change to pass (Out: S11-17).
- AC-3.1's numbers differ from 3.375 / 3.1 with the shipped segment section.

## 12. Open questions

None.

## 13. PR

**Title:** S11-16 — Rewrite graph alignment
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan and every expected value D-13 let change; manual check; `Plan: docs/deployment-plan/spike-11/prs/S11-16-graph-alignment.md`.
