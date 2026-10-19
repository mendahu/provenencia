# S11-14 — Segment compare

**Status:** Draft
**Spike / slice:** Spike 11 · Slice 4 — Graph alignment
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no new chrome: the wire fields, the page and its copy are unchanged)
**Board:** —
**Depends on:** S11-13 (landed [#487](https://github.com/mendahu/provenencia/pull/487); [plan](S11-13-promote-subject-compare.md) and its **Landed** note: ported fixtures were reshaped to keep their intent, `TestSeedFromNewlyFixedIntoUnpromoted` passes fan-out stats on a birth link, and STEP-2 to STEP-4 were one commit). Branch from `main`
**Blocks:** S11-16
**Branch:** `s11-14-segment-compare`
**Version bump:** none — Spike 11 PRs leave `VERSION` at 0.2.0, and the product bump is cut when a release ships ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-match-rule`](../../../../.agents/skills/add-match-rule/SKILL.md) (a new layer package, a registry section, knobs and inventory rows; this PR extends the skill)
**Rules that bind:** `.cursor/rules/go-tests.mdc`, `.cursor/rules/match-rules.mdc`
**Context:** [Segments](../comparison-layers.md#segments-layer-5-and-graph-alignment-layer-6), [Match registry](../comparison-layers.md#match-registry) (its "Typed keys" rule and `segment.go`; D-2 replaces its `BridgeKey{Kind, Role, NeighborType, End}` sketch), [Rules for every layer](../comparison-layers.md#rules-for-every-layer); [`docs/matching-model.md`](../../../matching-model.md) knob inventory (`segment.*` rows) and worked example step 5; working rules "Segment compare is one hop and pure", "Build, then migrate", "High tolerance for churn", "Sparse data first", "Registry numbers: carry over, then estimate", "One matching skill, grown with the layers", "One wire vocabulary"; user calls of 2026-10-19 written as D-5 and D-10
**Planned:** 2026-10-19 against `main` @ `6191fc8` (S11-13 merged); drafted at `6ecfd88` and refreshed
**Review:** — (round 1 answered in [`S11-14-segment-compare.review.md`](S11-14-segment-compare.review.md); round 2 pending)

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

Promote judges a link between two Subjects with layer 5, segment compare, instead of matching edge signatures and measuring fan-out from the catalog. A segment is primary – bridge – primary. Its bridge reads as a typed `vocabkey.BridgeKey`, so a spouse never corresponds to a sibling, and "Mary parent of John" corresponds to "John child of Mary" with the ends swapped. Whether a link is one-cardinality is declared in the registry: subject of a birth and subject of a death, from both ends. A differing neighbor there subtracts; on every other link it is unknown. `core/segmentcompare` is pure. The walk in `core/graphalign` calls it, and the fan-out scan goes.

**Why now.** Slice 4's Check ("a spouse never corresponds to a sibling; parent ↔ child corresponds with ends swapped; differing births subtract; differing residences are unknown") is S11-14's scope entry. S11-16's layer 6 walk compares segments around accepted pairs, so it needs this layer first. S11-13 left edge support, credit, fan-out and signatures in place for this PR.

**What changes for the researcher** (each is the design, not a regression; tests are updated to it, D-11):
- A link's support no longer depends on how many such links the catalog holds. Subject of a birth or a death earns 1.5 plus 0.5 of the neighbor's own evidence. Every other link (parents, spouses, residences, witnesses, locations, place parts) earns 0.25 plus 0.25. On a real catalog, a parent or spouse link with fan-out near 1 used to earn 1.5 + 0.5.
- A person whose birth (or death) is already placed on a different event than the handle's birth loses 3 points. A birth whose subject is already placed on a different person than the handle's subject loses 3 points too (user, 2026-10-19). Before, both were silent.
- A differing neighbor on any other link stays silent, as today. `part_of` keeps its gates: support needs 2 points of own evidence, and a different parent cancels the part's own evidence.
- The loader walks every bridge `connectrules` registers, with any term. Place relationships were walked only as `part_of` and `succeeded_by`; the seeded `place_relationship_type` vocabulary is product-locked to those two, so no link appears or disappears today.

**What stays.** The walk, the fallback pass, `refineOneHop`, `ConflictWithFixed`, the bands and the accept bar (S11-16). Which neighbors are compared stays the walk's choice: only neighbors already placed are compared, as today. The wire (`engine.proto`), Swift and copy. A bridge's own properties (dates on a place relationship) are not compared (D-10).

## 2. User stories

Conventions. `name(f)` is `propertycomparetest.Name(f)`, `term(k)` is `propertycomparetest.Term(k)`, `date(y)` is `propertycomparetest.Date(datevaluetest.YMD(y, 0, 0))`. Metas are `subjectcomparetest.Metas()` unless a row says otherwise; terms are `segmentcomparetest.Terms()` (D-3). Rules are `matchregistry.Default()`. Scores are within 1e-9. A key is written as its `String()` (D-2): `participation·subject·person·birth`. "Together" and "apart" are `segmentcompare.Together` and `segmentcompare.Apart` (D-4).

**US-1.** As the engine, I read a bridge as one typed key, so links correspond only when they mean the same thing.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | Relationship `spouse` from either end; relationship `sibling` | Both spouse keys are `relationship·spouse`; the sibling key is `relationship·sibling`; `Corresponds` is false between them |
| AC-1.2 | "Mary parent of John" (`person` Mary, `related_to` John, term `parent`) and "John child of Mary" (`person` John, `related_to` Mary, term `child`) | Both read from Mary's end as `relationship·child·related_to` and from John's end as `relationship·child·person` (`child` sorts before its inverse `parent`, so it is the canonical term, as `kinship.canonical` chose) |
| AC-1.3 | Participation, role `subject`, event type `birth`, from the person's end and from the event's end | `participation·subject·person·birth` and `participation·subject·event·birth` |
| AC-1.4 | Place relationship `part_of` from the part (`from`) and from the whole (`to`) | `place_relationship·part_of·from` and `place_relationship·part_of·to` |
| AC-1.5 | `segmentcompare.Compare` on two segments whose keys differ (spouse against sibling), together | Outcome unknown, cause `not_corresponding`, evidence 0 |

**US-2.** As the engine, I compare one segment with one research segment, and the cardinality table decides what a differing neighbor means.

`segmentcompare.Compare(a, b, placed, own, metas, reg.Segment, reg.Subject, reg.Value)` (D-4), where `a` and `b` share a key, `own` is the pair's own evidence, and `reg` is `matchregistry.Default()`.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | Key `participation·subject·person·birth`; neighbors event {`term(birth)`, `date(1817)`} on both sides; together; own 2.5 | Outcome agree; evidence `OneSupport + OneCredit × (event_type.Agree + date.Agree)` = 1.5 + 0.5 × 2.5 = 2.75; `Neighbor.Evidence` 2.5 |
| AC-2.2 | AC-2.1, apart | Outcome conflict; cause `placed_apart`; evidence −`OneConflict` = −3 |
| AC-2.3 | Key `participation·subject·event·birth`; apart | Outcome conflict; evidence −3 (the `One` entry `participation·subject·birth` holds for both ends) |
| AC-2.4 | Key `participation·subject·person·residence`; apart | Outcome unknown; evidence 0 |
| AC-2.5 | AC-2.4, together, neighbors {`term(residence)`} both sides | Outcome agree; evidence `OtherSupport + OtherCredit × 0.5` = 0.25 + 0.125 = 0.375 |
| AC-2.6 | Key `participation·subject·person·birth`; together; neighbor events {`term(birth)`, `date(1817)`} against {`term(birth)`, `date(1830)`} | Outcome partial; evidence 1.5 + 0.5 × (0.5 − 3) = 0.25 |
| AC-2.7 | Key `place_relationship·part_of·from`; together; neighbors {toponym "York"} both sides; own 1.9 | Outcome unknown; cause `gated`; evidence 0; `Neighbor.Evidence` 3 |
| AC-2.8 | AC-2.7 with own 3 | Outcome agree; evidence 0.25 + 0.25 × 3 = 1 |
| AC-2.9 | Key `place_relationship·part_of·from`; apart; own 3 | Outcome conflict; evidence −`Gates[…from].ConflictFactor` × 3 = −3 |
| AC-2.10 | Key `place_relationship·part_of·to`; apart; own 3 | Outcome unknown; evidence 0 |

**US-3.** As the engine, the walk scores links through segment compare, with no fan-out.

`graphalign` tests. A person's values are `{name: name("Ann Ames")}` with person metas; an event's are `{event_type: term(t), date: date(y)}`.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | Layer person P "Ann Ames", subject of birth E {birth, 1817}. Canon PER-1 "Ann Ames" subject of EVT-1 {birth, 1817}; PER-2 "Ann Ames" subject of EVT-2 {birth, 1817}. Fixed P → PER-1, E → EVT-2 | P's row score: 2.5 (name) − 3 = −0.5. E's row score: 2.5 (event) − 3 = −0.5 |
| AC-3.2 | AC-3.1 with both events `residence` instead of `birth` | P's row score 2.5; E's row score 2.5 |
| AC-3.3 | Layer Ann fixed to H-Ann; layer "Ann spouse Bob"; canon "H-Ann sibling PER-B", PER-B "Bob Ames" | Bob's row reason is not `via`; its score is the name weight 2.5 |
| AC-3.4 | `TestSparseNameAndBirthDate` | Unchanged: person and birth each strong at 5.25 |
| AC-3.5 | Every other `core/graphalign` test, ported (D-11) | Passes under D-11 |

**US-4.** As the engine (the Promote loader), I list a Source's links from the connect rules and key both sides through one builder.

Catalog tests in `core/database/promotealign` with the existing `fixture` (`newFixture`, `person`, `bare`, `cite`, `birth`, `place`, `participation`, `locate`, `relationship`, `placeRel`, `promote`, `propose`, `newSource`). A "dated birth Y" is `f.bare(src, "event")` cited with `event_type` `birth` and `date` Y, as in `TestProposeOneHopDateExhibit`.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-4.1 | A Source with one participation (`f.person` subject of `f.birth`), one location (`f.locate` that birth at `f.place`), one relationship (`f.relationship` person parent of person) and one place relationship (`f.placeRel` place `part_of` place) | `promotealign.LayerBridges` returns four bridges; their `KeyA.Type` values, sorted, equal `connectrules.BridgeTypeKeys()` sorted |
| AC-4.2 | `TestProposeRelationshipFromEitherEnd` (Source A records "Mary parent of John", Source B records "John child of Mary") | Passes: John's row on Source B proposes John's handle through Mary; its `Via.Key.Type` is `relationship` and `Via.Key.Term` is `child` |
| AC-4.3 | Source A: person "Ann Ames" subject of a dated birth 1817, both promoted. Source C: person "Mary Ames" subject of a dated birth 1817, both promoted. Source B: person "Ann Ames" subject of a dated birth 1817. Propose B with B's birth fixed to C's birth handle | B's person row's Target is not A's person handle (before this PR it filed weak at 2.6: name 2.5 + lifespan 0.1, the differing birth silent; now 2.6 − 3 = −0.4, under the accept bar) |
| AC-4.4 | `go vet ./...`; `rg -n 'FanOut|loadStats|LoadStatsForTest|computeStats|EdgeSignature|SignatureScale|EdgePoints|EdgeFact' core api` | Builds; the grep prints nothing |
| AC-4.5 | Source A: a birth located at place "York" (`f.birth`, `f.place`, `f.locate`), both promoted. Source B: the same shape. Propose B with B's birth fixed to A's birth handle | B's place row: Target A's place handle, reason `via` (a layer location key equals the canon one: no event type on a location) |

**US-5.** As a contributor, the segment section is one registry entry per rule and the skill says how to add one.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-5.1 | `TestKnobInventoryMatchesRegistry` | Passes with `segment.` removed from `notYetBuilt` |
| AC-5.2 | `TestSegmentSectionMatchesSeed` *(new)* | Every key in `One` and `Gates` names a registered bridge type, a seeded term of its disambiguation property as `Term`, and a seeded `event_type` term as `EventType` when set; a `Gates` key names one of its endpoint properties as `End`; every `One` key has an empty `End`; `segmentcomparetest.Terms()` equals the seed's directed and inverse marks for every term of every bridge's disambiguation property |
| AC-5.3 | `TestLayerPackages` | `core/segmentcompare` (`numbers` true) and `core/segmentcompare/segmentcomparetest` (`numbers` false) are listed and pass |

## 3. Scope

| In | Covered by |
| --- | --- |
| `vocabkey.BridgeKey`, `BridgeTerm`, bridge type, endpoint and `part_of` constants; `connectrules` and `propertyterms.KeyPartOf` read them (D-2) | STEP-1, US-1 |
| `core/segmentcompare`: `Link`, `Terms`, `Key`, `Corresponds`, `Segment`, `Placement`, `Compare`, `Result`, `Rules`, `Gate`, causes; `segmentcomparetest.Terms` (D-1, D-3, D-4) | STEP-2, US-1, US-2 |
| Registry segment section, inventory rows, seed check (D-5) | STEP-3, US-5 |
| `graphalign` on segment compare; edge signatures, edge facts, signature scales, fan-out config and `Stats` deleted (D-6) | STEP-4, US-3 |
| Promote loader: link kinds from `connectrules`, one key builder for both sides, terms from the vocabulary; `stats.go` deleted; `graphcache.Link`'s unread term marks deleted (D-7) | STEP-5, US-4 |
| FFI `Via` fields from the key (D-8) | STEP-5 |
| Catalog tests for the slice Check (D-11) | STEP-6 |
| Skill and docs the change makes false (D-12) | STEP-7 |

| Out | Home |
| --- | --- |
| Comparing a bridge's own properties (place relationship dates) through subject compare (user, 2026-10-19) | S11-16 (scope entry updated at planning) |
| Comparing neighbors that are not yet placed, by entity compare alone (a third `Placement`) | S11-16 |
| The layer 6 walk, top K, dirty-pair settling, `ConflictWithFixed` | S11-16 |
| `graphcache`'s stats slot (`Stats`, `SetStats`, `statsOK`, `ResetStatsForTest`) | S11-15 renames and reuses it; this PR only corrects its doc comment |
| Moving `conclusionheaders.RelPartOf` and the `"subject"` / `"birth"` / `"death"` literals in `core/hops/hops.go`, `core/database/promotealign/layer.go` and `core/database/conclusionheaders/compose.go` onto `vocabkey` constants | Not planned — each moves when next edited (S11-12 Out) |
| A type alias making `vocabkey.Role` a `BridgeTerm` | Not planned — `Role` types participation roles for lifespans, and an alias would let a relationship term pass as a role |
| Evidence lines for a corresponding segment (relationship bundles) | S11-19 |
| Plugin bridges in the match registry (a plugin's link is an "other" link with no entry) | Not planned in Spike 11 |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Edge signature | `core/graphalign/types.go` `EdgeSignature{BridgeType, RoleOrType, NeighborKind, NeighborTypeTerm, Directed}`, `Key()`, `ScaleKey(fromEnd)`; `Bridge{A, B, Signature}`; `CanonEdge{From, To, Signature}`; `Via{NeighborSubjectID, Signature}`; `Stats{FanOut}`; package doc "S9-42 loads Layer / Canon / Stats" | Untyped pipe-joined strings; a directed term carries `Directed` and ends through `fromEnd` |
| Walk links | `core/graphalign/align.go` `layerLink`, `canonLink` (`neighbor`, `sig`, `fromEnd`), `corresponds` (callers in `pushFromAnchor`, `supports`, `edgeFacts`, `scoreFixed`), `newState` (builds `layerAdj` / `canonAdj`), `scoreFixed` (`fromNeighbor` view), `scoredCand.viaSig`, `enqueue(…, via EdgeSignature, …)`, `recordBest` | Correspondence is `sig.Key()` equality plus same end for a directed term |
| Edge scoring | `core/graphalign/align.go` `edgeFacts`, `nodeScore`, `scoreCandidate` (passes `r.Evidence` to `EdgePoints`); `core/graphalign/score.go` `EdgeFact`, `EdgePoints`, `ScoreCandidate(r subjectcompare.Result, edge, provenance float64) Scored` (its doc says "edge is EdgePoints over the corresponding links") | Support and credit by fan-out (`FanOutLowMax`); a mismatch subtracts only through `SignatureScale.ConflictFactor` |
| New-row duplicates | `core/graphalign/align.go` `alikeNew`, `newPairFacts`, `pairGroup`, `neighborsMatch` (one caller: `newPairFacts`), `pairNode` | Groups neighbors by signature and end; each pair is evaluated twice (`neighborsMatch`, then `pairNode`) |
| Walk config | `core/graphalign/registry.go` `Config{EdgeSupportLow, EdgeSupportHigh, NeighborCreditLow, NeighborCreditHigh, FanOutLowMax, FanOutUnknown, SignatureScale}`, `SignatureScale{MinNode, ConflictFactor}`, `Config.Subject`, `DefaultConfig` | The `part_of` scale is keyed `place_relationship|part_of|place||from` / `…|to` |
| Entry | `core/graphalign/align.go` `Align(layer, canon, stats, fixed, cfg)`, `newState(layer, canon, stats, fixed, cfg)`, `state.stats` | Takes fan-out stats |
| Walk tests | `core/graphalign/align_test.go` helpers `partSig`, `relSig` (test-only terms `parent_of`, `kin` beside seeded ones), `partOfSig`; 41 `graphalign.Stats{…}` arguments; `TestHighFanOutNameCreditStaysUnderTheBar`, `TestConflictingNeighborLowersTheScore`, `TestEdgePointsHonorsTheSignatureScale`, `TestDefaultPartOfScale`, `TestPartOfDoesNotFileADifferentToponym`, `TestPartOfBreaksASharedToponym`, `TestDirectedRelationshipsMatchFromTheSameEnd` (toggles `Directed` on `parent_of`), `TestNewDuplicateCheckComparesDirectedNeighborsByEnd`, `TestNewEventsDisagreeWhenTheirParticipantsDo`, `TestSeedFromNewlyFixedIntoUnpromoted` (fan-out 1 on a birth link); `core/graphalign/pick_internal_test.go` calls `newState(layer, canon, Stats{}, nil, cfg)` | |
| Layer links | `core/database/promotealign/layer.go` `hopPersonToEvent`, `hopEventToPlace`, `hopPersonRelated`, `hopPlaceParent`, `hopPlaceSuccessor`, `loadLayerBridges` (four hand-written walks; it also returns `[]participation` for primary person and event pairs, which `loadLayer` reads for lifespans), `bridgeKey`, `neighborTypeProperty`; `core/database/promotealign/profiles.go` `profiler.profile` and `profiler.build` follow `hopPersonToEvent` | Link kinds listed in code; the relationship walk swaps ends through `edgeStamp`'s `flip` |
| Canon links | `core/database/promotealign/canon.go` `canonSignature(self, selfKind, n, l, neighbor, props map[string]graphcache.VocabProperty, kin kinship)` (event type from `keptTerm` only when `l.BridgeTypeKey == "participation"`), `keptTerm(n, props, key)`, the edge loop in `loadCanon` (`kin` from `loadKinship`; `ek` dedupe by `sig.Key()`) | Signs with `edgeStamp` |
| Terms | `core/database/promotealign/terms.go` `kinship`, `loadKinship`, `edgeStamp`, `kinship.canonical`, `observedTerms`, `termsBy` | Directed and inverse marks, keyed `propertyKey|termKey`; the canonical term is the one that sorts first |
| Fan-out | `core/database/promotealign/stats.go` `loadStats`, `cloneStats`, `computeStats`, `sqlFanOut`, `fanOutSum`, `addFanOut`; `export_test.go` `LoadStatsForTest` (its only export); `promotealign.go` `Propose` calls `loadStats` and passes `stats` to `Align`; tests `TestStatsCacheInvalidatesOnWrite`, `TestStatsFanOut` (calls `LoadStatsForTest`) | One SQL scan per bridge type, cached on `graphcache` |
| Graph cache | `core/database/graphcache/graph.go` `Stats`, `SetStats`, `ResetStatsForTest`, fields `stats`, `statsOK`; `Link.Directed`, `Link.InverseKey` (doc: "copied off the term row at load; the edge signature is built later, at the feature edge"); `load.go` `termMark{…, inverseKey, directed}`, filled by `disambiguation` (columns `t.inverse_key`, `t.directed`) and copied in `finishLinks` | Only `promotealign` reads the stats slot; nothing outside `graphcache` reads `Link.Directed` or `Link.InverseKey` |
| Connect rules | `core/connectrules/connectrules.go` `productBridges` (literals `"participation"`, `"relationship"`, `"location"`, `"place_relationship"`, `"person"`, `"event"`, `"place"`, `"related_to"`, `"from"`, `"to"`; already imports `vocabkey`), `Bridges()`, `Bridge{Origin, BridgeTypeKey, Endpoints, Disambiguation}`, `Endpoint{PropertyKey, TypeKey}`, `HasDisambiguation`, `LookupBridge`, `BridgeTypeKeys` | The one product table of bridges |
| Part-of constants | `core/database/propertyterms/propertyterms.go` `KeyPartOf = "part_of"`; `core/database/conclusionheaders/places.go` `RelPartOf = "part_of"` | Two copies of the term key |
| Hops | `core/hops/hops.go` `NewHop(bridgeTypeKey, fromProperty, toProperty, filter)`; `core/database/canonicalgraph/canonicalgraph.go` `WalkSource` | `NewHop` checks a hop against `connectrules` |
| FFI | `api/ffi/handlers/promote_align.go` `ViaBridgeType: r.Via.Signature.BridgeType`, `ViaRole: r.Via.Signature.RoleOrType` | No test asserts either field |
| Keys | `core/vocabkey/vocabkey.go` `SubjectType`, `PropertyKey`, `EventType`, `Role`, `EventBirth`, `EventDeath`, `RoleSubject`; `vocabkey_test.go`; `core/database/subjectvocab/vocabkey_test.go` `TestVocabKeysMatchSeed` | No bridge key |
| Registry | `core/matchregistry/registry.go` `Registry{Value, Subject}`; `inventory_test.go` `notYetBuilt` (`segment.`, `alignment.`, `evidence.`); `guard_test.go` `layerPackages` | No segment section |
| Seed | `core/database/subjectvocab/registry.go` `seedTerms` (`role` terms; `relationship_type` with `Directed` and `Inverse`; `place_relationship_type` `part_of`, `succeeded_by`, directed, no inverse), `seedProperties`, `seedTypes`; `subject_section_test.go` `TestSubjectSectionMatchesSeed` (the shape to copy) | |
| Fixtures | `core/subjectcompare/subjectcomparetest` `Metas()` (the shared fixture shape) | |
| Docs | `docs/matching-model.md` knob inventory rows `segment.*`, worked example step 5 (`segment.Cardinality.birth`), "Removed knobs"; `docs/promote-graph-alignment.md` (signature, fan-out and support text: `rg -n 'fan-out|FanOut|signature|EdgeSupport|NeighborCredit' docs/promote-graph-alignment.md`); `docs/deployment-plan/spike-11/comparison-layers.md` §Segments (its bridge bullet "by Subject comparison on the bridge's own properties") and §Match registry (the `BridgeKey{Kind, Role, NeighborType, End}` sketch); `.agents/skills/add-match-rule/SKILL.md` | |

**Precedent.** S11-10 ([#471](https://github.com/mendahu/provenencia/pull/471), [plan](S11-10-subject-compare.md)) built a layer package with its registry section, inventory rows, guard line, seed check, fixture package and skill section: copy its shape for STEP-2 and STEP-3 (`TestSubjectSection`, `TestSubjectSectionMatchesSeed` and `subjectcomparetest` are the models). S11-13 ([#487](https://github.com/mendahu/provenencia/pull/487), [plan](S11-13-promote-subject-compare.md)) moved the walk onto a layer and ported `align_test.go` in place, reading expected numbers from the config and reshaping fixtures to keep their intent: copy that for STEP-4. S11-12's **Landed** note: a signature change that crosses packages is one commit (D-13).

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | **Package.** `core/segmentcompare` *(new)*, package doc "Package segmentcompare is layer 5 of matching: whether a segment, two primary Subjects joined by a bridge, lines up with a research segment. Pure: no I/O, no catalog, no numbers; the segment section, layer 3's rules and the bridge terms come in as arguments." Files: `segmentcompare.go` (`Rules`, `Gate`, `Segment`, `Placement`, `Result`, `Cause`, `Compare`), `key.go` (`Link`, `Terms`, `Term`, `Key`, `Corresponds`). External tests in `segmentcompare_test`. Fixture package `core/segmentcompare/segmentcomparetest` *(new)*, package doc "Package segmentcomparetest holds segment compare fixtures.", with `Terms() segmentcompare.Terms`: a fresh map of every seeded term of `role`, `relationship_type` and `place_relationship_type` under its bridge type, with the seed's directed and inverse marks | Layer 5's package per the working rule "Packages"; the fixture follows `subjectcomparetest.Metas()`, checked against the seed (AC-5.2) | Folding it into `graphalign` — layer 6 must compose layer 5, not contain it; terms written in each test — unguarded copies of seed facts |
| D-2 | **Typed keys in `vocabkey`.** `type BridgeTerm string` ("a bridge's disambiguation term key: a participation role, a relationship type or a place relationship type, typed for registry rules"). `type BridgeKey struct { Type SubjectType; Term BridgeTerm; End PropertyKey; EventType EventType }` with doc "BridgeKey is one bridge read from one end: its bridge type, its term in canonical direction, the endpoint Property of the end it is read from (empty for a symmetric link between two Subjects of one type), and the type of the event on a participation". `func (k BridgeKey) String() string`: the non-empty fields joined with `·` in field order. Constants in a block "Bridge types", untyped like `Person`: `BridgeParticipation = "participation"`, `BridgeRelationship = "relationship"`, `BridgeLocation = "location"`, `BridgePlaceRelationship = "place_relationship"`; added to the "Seeded Property keys" block: `PropertyPerson = "person"`, `PropertyEvent = "event"`, `PropertyPlace = "place"`, `PropertyRelatedTo = "related_to"`, `PropertyFrom = "from"`, `PropertyTo = "to"`; with the term keys: `TermPartOf BridgeTerm = "part_of"`. The role term is written `BridgeTerm(RoleSubject)`. `connectrules.productBridges` writes its bridge types, endpoint property keys and endpoint types with these constants (`Person`, `Event`, `Place` for types). `propertyterms.KeyPartOf` becomes `string(vocabkey.TermPartOf)`. `TestVocabKeysMatchSeed` checks the four bridge types against `seedTypes`, the six endpoint properties against `seedProperties`, and `TermPartOf` under `place_relationship_type` | The registry rule "Typed keys, not pipe-joined strings"; S11-12 named `BridgeKey` as the next user of its typed term keys; the skill: "no other package declares a key type". `Type` and the `Bridge` prefix follow `connectrules`' "bridge type" and vocabkey's family prefixes, so `Location` never reads as a fourth kind. `connectrules` reading the constants gives each a reader and one declaration | `Kind` for the bridge type — vocabkey's "kind" is a primary Subject type; `End` as a bool (`fromEnd`) — it cannot say which end of a participation; the §Match registry sketch `{Kind, Role, NeighborType, End}` — `Role` is too narrow for relationship terms and `NeighborType` is the event type from either end |
| D-3 | **Key rule.** `segmentcompare.Link{Type vocabkey.SubjectType; Term vocabkey.BridgeTerm; End, Other vocabkey.PropertyKey; EventType vocabkey.EventType; SameType bool}`: a bridge as recorded, read from the end whose endpoint Property is `End`; `SameType` is true when both endpoints hold the same Subject type. `Terms map[vocabkey.SubjectType]map[vocabkey.BridgeTerm]Term`, `Term{Directed bool; Inverse vocabkey.BridgeTerm}`: the catalog's term facts per bridge type; a term or type with no entry is the zero `Term`. `Key(l Link, terms Terms) vocabkey.BridgeKey`: start from `l.Term`, `l.End`, `l.Other`; when `terms[l.Type][term].Inverse` is non-empty and sorts before `term`, take the inverse and swap `End` and `Other`; then when `l.SameType` and the resulting term is not `Directed`, `End` is empty. `Type` and `EventType` are copied. `Corresponds(a, b vocabkey.BridgeKey) bool` is `a == b`, doc "Corresponds reports whether two links mean the same thing read from the same end. It is the one correspondence rule; the walk and Compare call it." | "A term and its inverse correspond with the ends swapped" is layer 5's rule ([segments](../comparison-layers.md#segments-layer-5-and-graph-alignment-layer-6)); directed and inverse marks are data-model facts passed in. Same canonical choice as `kinship.canonical` (`child` over `parent`), so `via_role` keeps its values. One named function for correspondence, so S11-16 or a bridge-property rule changes one place | Normalizing in the loader (today's `edgeStamp`) — layer 5's rule outside layer 5; inline `==` at each walk site — five copies of the rule |
| D-4 | **Compare.** `Segment{Key vocabkey.BridgeKey; NeighborKind vocabkey.SubjectType; Neighbor subjectcompare.Values}`. `type Placement string`: `Together = "together"` (the layer neighbor is placed on this research neighbor), `Apart = "apart"` (placed on another handle). `Rules{One []vocabkey.BridgeKey; OneSupport, OneCredit, OneConflict, OtherSupport, OtherCredit float64; Gates map[vocabkey.BridgeKey]Gate}`, `Gate{MinOwnEvidence, ConflictFactor float64}`; doc on `One`: "links where a differing neighbor conflicts, written with an empty End: each holds from both ends". `Result{Outcome propertycompare.Outcome; Cause Cause; Evidence float64; Neighbor subjectcompare.Result}`. Causes: `CauseNotCorresponding = "not_corresponding"`, `CauseGated = "gated"`, `CausePlacedApart = "placed_apart"`; doc on `Cause`: "why a segment adds nothing or subtracts; empty when it supports". `Compare(a, b Segment, placed Placement, own float64, metas subjectcompare.Metas, rules Rules, subject subjectcompare.Rules, value valuecompare.Rules) Result`, in order: (1) `!Corresponds(a.Key, b.Key)` → unknown, `not_corresponding`, 0. (2) `g := rules.Gates[a.Key]`; `one` is true when `rules.One` holds `a.Key` with its `End` cleared. (3) Apart: evidence `−(OneConflict if one) − g.ConflictFactor × own`; outcome conflict when `one` or `g.ConflictFactor > 0`, else unknown; cause `placed_apart`. (4) Together: `Neighbor = subjectcompare.Compare(a.NeighborKind, a.Neighbor, b.Neighbor, metas, subject, value)`. (5) `g.MinOwnEvidence > 0` and `own < g.MinOwnEvidence` → unknown, `gated`, 0, `Neighbor` kept. (6) Evidence `support + credit × Neighbor.Evidence` with the One pair when `one`, else the Other pair; outcome partial when any `Neighbor.Properties` entry's outcome is conflict, else agree; cause empty | The cardinality table in [segments](../comparison-layers.md#segments-layer-5-and-graph-alignment-layer-6), with today's credit and `part_of` gates carried over. Placement comes from the walk because only layer 6 knows where a neighbor is placed. `Result.Neighbor` carries layer 3's speaking values for S11-19's bundle ("output carries value refs"), and the walk's duplicate check reads it (D-6). The cause says what `Corresponds` would, so no separate flag | Entity compare alone deciding "differs" in this PR — the walk compares placed neighbors only, and changing that is S11-16's walk; a `Corresponds` field — the same fact as `Cause == not_corresponding`; a `Similarity` field — `Neighbor.Similarity` holds it |
| D-5 | **Registry section (user, 2026-10-19: both ends).** `matchregistry.Registry` gains `Segment segmentcompare.Rules` ("Segment is layer 5: what a matching or differing link counts."), set by `defaultSegment()` in `core/matchregistry/segment.go` *(new)*: `One` = `participation·subject·birth` and `participation·subject·death` (comment: decided: a person has one birth and one death, and each has one subject); `OneSupport` 1.5 and `OneCredit` 0.5 (`graphalign` `EdgeSupportLow`, `NeighborCreditLow`); `OtherSupport` 0.25 and `OtherCredit` 0.25 (`EdgeSupportHigh`, `NeighborCreditHigh`); `OneConflict` 3 (estimate: a differing birth or death is as strong as a differing name or date, `subject.Weights.person.name.Conflict`); `Gates[place_relationship·part_of·from]` = {2, 1} and `Gates[place_relationship·part_of·to]` = {2, 0} (`SignatureScale` `part_of`; the whole's end had no conflict factor). Inventory rows replace `segment.Cardinality.*` and `segment.PartOf.*` (STEP-3) | Numbers are the inventory's starting values; `OneConflict` is the one new knob, estimated under the working rule "Registry numbers: carry over, then estimate". One entry per fact, as §Segments says ("adding another axiomatic case later is one registry entry") | One key per end — four entries for two facts, and a new case needs two; a cardinality enum per key — `One` is the only declared value; `ConflictFactor × own` for births — a sparse person's own evidence is small, so a differing birth would barely count |
| D-6 | **Walk on segments.** `graphalign.Bridge{A, B []byte; KeyA, KeyB vocabkey.BridgeKey}` (KeyA is the bridge read from A's end); `CanonEdge{From, To []byte; KeyFrom, KeyTo vocabkey.BridgeKey}`; `Via{NeighborSubjectID []byte; Key vocabkey.BridgeKey}`. `Config` loses `EdgeSupportLow`, `EdgeSupportHigh`, `NeighborCreditLow`, `NeighborCreditHigh`, `FanOutLowMax`, `FanOutUnknown`, `SignatureScale` and gains `Segment segmentcompare.Rules` (after `Subject`, doc "Segment is the registry's segment section: what a link counts. DefaultConfig holds the registry's."); `DefaultConfig` sets `Segment: reg.Segment`. `Align(layer, canon, fixed, cfg)` and `newState(layer, canon, fixed, cfg)` lose `stats`; the package doc reads "S9-42 loads Layer / Canon and calls Align". Deleted: `Stats`, `state.stats`, `EdgeSignature`, `SignatureScale`, `EdgeFact`, `EdgePoints`, `corresponds`, `pairGroup`, `nodeScore`, `pairNode`, `neighborsMatch`. `layerLink` and `canonLink` become `{neighbor []byte; key, back vocabkey.BridgeKey}` (`key` read from this end, `back` from the neighbor's). Every correspondence check calls `segmentcompare.Corresponds`: `pushFromAnchor` and `supports` on `link.key, ce.key`; `scoreFixed`'s neighbor view on `link.back, ce.key`. `edgeFacts` becomes `segments(subjectID []byte, h *Handle, own float64) float64`: for each layer link of the Subject (each link on its own, as today) whose neighbor is placed on handle `hid`, the first canon link of `h` in `canonAdj` order that corresponds and lands on `hid` is compared together; failing that, the first that corresponds is compared apart; none adds nothing. The layer side is `Segment{link.key, kind of st.subjects[neighbor], its Values}`, the canon side `Segment{ce.key, kind of st.handles[ce.neighbor], its Values}`; it returns the sum of `Evidence`. `scoreCandidate` passes the Subject's `subjectcompare.Result.Evidence` as `own`; `ScoreCandidate`'s doc says `edge` is the summed segment evidence. `newPairFacts` becomes `newPairSegments(a, b *Subject, own float64) (float64, bool)`: each Subject's neighbors grouped by `link.key`; for a key both have, every neighbor pair is compared together; a pair matches when it is the same Subject or `Result.Neighbor.Evidence` is at least `AcceptScore` (today's `neighborsMatch` bar); no match → disagree; otherwise it adds the best `Evidence` among matching pairs. `alikeNew` computes the pair's own result first. `scoredCand.viaSig` becomes `viaKey`, set from `link.key` wherever it took `link.sig` | Layer 6 composes layer 5 and decides placement (rule 7); one key comparison replaces signature equality plus `Directed` plus `fromEnd`; one subject compare per neighbor pair | Passing `Terms` into `graphalign` — the loader has the vocabulary, and keys travel ready-made; keeping `neighborsMatch` — a second evaluation of the pair `Compare` already scored |
| D-7 | **Loader: one key builder.** `linkOf(b connectrules.Bridge, term, eventType string, firstEnd bool) segmentcompare.Link` *(new, `terms.go`)*: `Type` = `b.BridgeTypeKey`; `End` / `Other` = the first / second endpoint's `PropertyKey` (swapped when `!firstEnd`); `SameType` = the two endpoints' `TypeKey`s are equal; `EventType` = `eventType` when `b.BridgeTypeKey` is `vocabkey.BridgeParticipation`, else empty (today's rule on both sides). `loadTerms(g *graphcache.Graph) (segmentcompare.Terms, error)` *(new, `terms.go`)*: for each `connectrules.Bridges()` entry whose disambiguation `HasDisambiguation`, every `g.Vocabulary()` term with that `PropertyKey`, under the bridge type, with `Directed` and `InverseKey`. *Layer:* `loadLayerBridges` walks each `connectrules.Bridges()` entry with two endpoints whose first endpoint's type has Subjects on the layer (`byKind`): `hops.NewHop(b.BridgeTypeKey, ep[0].PropertyKey, ep[1].PropertyKey, nil)` (its error is returned), `canonicalgraph.WalkSource`; the term through `observedTerms(q, assocs, b.Disambiguation)` when `connectrules.HasDisambiguation`, else empty; for a participation, the event type through `observedTerms(q, toIDs, neighborTypeProperty[vocabkey.Event])`. A bridge with a disambiguation but no observed term keys with an empty `Term`. For a participation whose ends are both primary, the edge also adds `participation{person, event, role}` as today. Each edge becomes a `Bridge` with `KeyA = segmentcompare.Key(linkOf(b, term, eventType, true), terms)` and `KeyB` the same with `false`. `bridgeKey` is renamed `edgeSeenKey(a, b []byte, keyA, keyB vocabkey.BridgeKey) string`: when `a` sorts after `b`, swap both the ids and the keys; then `a|b|keyA.String()|keyB.String()`. *Canon:* `canonSignature` becomes `canonKeys(self []byte, selfKind string, n *graphcache.Node, l graphcache.Link, neighbor *graphcache.Node, props map[string]graphcache.VocabProperty, terms segmentcompare.Terms) (from, to []byte, keyFrom, keyTo vocabkey.BridgeKey, ok bool)`: `b, ok := connectrules.LookupBridge(l.BridgeTypeKey)` (not ok → the loop skips the link); `from, to` = self and neighbor ordered by `l.FromEnd` as recorded (no swap); the event type from `keptTerm` as today; keys through `linkOf(b, l.TermKey, eventType, true / false)` and `Key`. `loadCanon` calls `loadTerms` where it called `loadKinship`; `ek` uses `keyFrom.String()`. *Deleted:* `hopEventToPlace`, `hopPersonRelated`, `hopPlaceParent`, `hopPlaceSuccessor` (`hopPersonToEvent` stays: `profiles.go` reads it), `kinship`, `loadKinship`, `edgeStamp`, `canonical`, `canonSignature`, `stats.go`, `export_test.go`'s `LoadStatsForTest`, `TestStatsCacheInvalidatesOnWrite`, `TestStatsFanOut`; `Propose` drops `loadStats`. *Graph cache:* `Link.Directed` and `Link.InverseKey`, their copies in `finishLinks`, and `termMark.inverseKey` / `directed` with their two columns in `disambiguation` are deleted (`loadTerms` is the one source); the `Link` doc loses "Directed and InverseKey are copied off the term row at load; the edge signature is built later, at the feature edge."; the `Stats` / `SetStats` doc says the slot holds a catalog-wide scan a feature caches, filled by nothing until S11-15 | The map: "The layer loader reads link kinds from `connectrules` instead of a list in code"; a plugin's registered bridge reaches the walk with no edit here. One builder decides `End`, `Other`, `SameType` and the event-type rule for both sides, so a layer key and a canon key cannot drift (a location has no event type on either side, as today) | The four named walks — the list in code the map retires; a second key builder in `canonKeys` — two copies of the event-type rule; keeping `Link.Directed` / `InverseKey` — a second, unread path for the same facts; deleting `graphcache`'s stats slot — S11-15 reuses it |
| D-8 | **Wire.** The FFI handler sends `ViaBridgeType: string(r.Via.Key.Type)`, `ViaRole: string(r.Via.Key.Term)`. No proto change | Same values as `BridgeType` and `RoleOrType` today (the canonical term), so Swift is untouched ("One wire vocabulary"); AC-4.2 pins the term on a parent / child row | A new `via_end` field — nothing on the page reads it |
| D-9 | **No reordering of the walk.** Which pairs are scored, when `refineOneHop` runs, the fallback, `ConflictWithFixed` and the bands stay as S11-13 left them. Only the points a link adds change | S11-16 rewrites the walk; this PR replaces the link scorer under it | Comparing unplaced neighbors now — S11-16's step 4 |
| D-10 | **Bridge properties are not compared (user, 2026-10-19).** `Segment` holds no bridge values, and neither loader reads them. The key carries the role, relationship type and event type. The map's S11-14 and S11-16 scope entries already say so (edited at planning); STEP-7 corrects `comparison-layers.md` §Segments | No inventory weight exists for any bridge property, so the comparison would score nothing, and both loaders would read rows for it | Plumbing bridge values with no weights — loader work and a code path no test can show working |
| D-11 | **Ported tests.** In `align_test.go`, `partSig(role, neighborKind, typeTerm)` becomes `partKeys(role vocabkey.BridgeTerm, eventType vocabkey.EventType) (person, event vocabkey.BridgeKey)`, `relSig(typ)` becomes `relKeys(term vocabkey.BridgeTerm) (a, b vocabkey.BridgeKey)` and `partOfSig()` becomes `partOfKeys() (part, whole vocabkey.BridgeKey)`; each builds its keys with `segmentcompare.Key` on a `Link` over `segmentcomparetest.Terms()`, never by hand. Test-only terms move to seeded ones: `parent_of` → `parent`, `kin` → `cousin`; `TestDirectedRelationshipsMatchFromTheSameEnd` becomes two rows, `spouse` (symmetric, wants `h-parent`) and `parent` (directed, wants `h-child`), with the edges recorded as "PER-0 parent of Ann" and "Ann parent of PER-9". Every `graphalign.Stats{…}` argument goes; `pick_internal_test.go` calls `newState(layer, canon, nil, cfg)`. A test reading `cfg.EdgeSupportLow` / `NeighborCreditLow` reads `cfg.Segment.OneSupport` / `OneCredit` when its key is a `One` fact, else `OtherSupport` / `OtherCredit`; `SignatureScale` reads `cfg.Segment.Gates`. `TestHighFanOutNameCreditStaysUnderTheBar` is renamed `TestOtherLinkCreditStaysUnderTheBar` (residence: the Other pair). `TestEdgePointsHonorsTheSignatureScale` and `TestDefaultPartOfScale` are deleted; their cases are T-3's rows AC-2.7 to AC-2.10. **Expected changes, and only these:** (a) a link's support and credit come from `One` / Other, not fan-out (a test with fan-out 1 on a non-birth link drops to the Other pair); (b) a birth or death link placed apart subtracts 3 from both ends. A ported test outside (c) whose asserted Target flips under (a) or (b) has its fixture reshaped to keep its intent (the S11-13 precedent), and each reshape is a line of the Landed note. (c) A test whose name states the behavior (`TestSparseNameAndBirthDate`, `TestGracieObituaryGolden`, `TestPartOfChainIsStrongWithoutADecision`, `TestPartOfChainIsAStrongMatch`, `TestPartOfDoesNotFileADifferentToponym`, `TestPartOfBreaksASharedToponym`, `TestDirectedRelationshipsMatchFromTheSameEnd`, `TestProposeRelationshipFromEitherEnd`, `TestProposePartOfMatchesTheSameEnd`) keeps every Target and handle it asserts. Any other change is a stop condition | The S11-13 precedent; the goldens are the acceptance tests ("Sparse data first"); keys built through `Key` keep the direction rule in one place | Hand-built keys in helpers — a second copy of D-3's rule; rewriting the suite on new fixtures — numbers would move for no behavior |
| D-12 | **Skill and docs.** `add-match-rule`: the description and intro list `core/segmentcompare`; a rule bullet "Layer 5 (`core/segmentcompare`) scores one link": keys come from `Key` with the catalog's `Terms`; correspondence is `Corresponds` and nothing else; `One` (empty `End`, both ends) and `Gates` (per end) are keyed by `vocabkey.BridgeKey`; placement comes from layer 6; neighbor evidence through layer 3; tests build terms with `segmentcomparetest.Terms`. A section "Add a link rule": a `One` key or a `Gates` entry in `core/matchregistry/segment.go` (new) with a source comment, its inventory row named by the key's `String()`, `TestSegmentSectionMatchesSeed`, and a `TestCompare` row. "Add a relation, cause or unit code" gains: "a segment cause is layer 5's: add it in `core/segmentcompare` and to the layer 5 output list in `comparison-layers.md` §Segments." The Check command gains `./core/segmentcompare/...`. `.cursor/rules/match-rules.mdc` globs gain `core/segmentcompare/**`. `matching-model.md`: the segment rows (STEP-3); worked example step 5 reads `segment.One` (`participation·subject·birth`); "Removed knobs" gains `graphalign` `SignatureScale` (→ `segment.Gates`) and the fan-out scan (`Stats.FanOut`, `sqlFanOut`). `comparison-layers.md` §Segments: the bridge bullet becomes "The bridge as a typed key: its role or relationship term in canonical direction, the end it is read from, and the event type on a participation. A term and its inverse correspond with the ends swapped. Its own properties (dates) are compared in S11-16 if a golden needs them."; a "Built in S11-14" paragraph names `segmentcompare.Compare`, `Key`, `Corresponds`, the two placements and the three causes. §Match registry's sketch becomes `BridgeKey{Type, Term, End, EventType}`. `promote-graph-alignment.md`: its signature, fan-out and support text describes keys and the segment section | Working rule "One matching skill, grown with the layers"; each migration PR fixes the docs it makes false; a layer 5 cause gets a home | — |
| D-13 | **Commits.** STEP-4 and STEP-5 are one commit: `graphalign`'s types change breaks `promotealign` and the FFI handler until both move. Every other step is its own commit | S11-12's Landed note | Two commits that do not build |

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `core/segmentcompare` *(new)*: `Compare`, `Result`, `Rules`, `Gate`, `Segment`, `Placement`, `Cause` | package | `core/graphalign/score.go` `EdgeFact`, `EdgePoints`; `registry.go` `SignatureScale` | Layer 5 is its own pure layer; these move into it and are deleted (D-1, D-4) |
| `segmentcompare.Key`, `Corresponds`, `Link`, `Terms`, `Term` *(new)* | func, type | `promotealign` `edgeStamp`, `kinship.canonical`; `graphalign.corresponds`; `graphcache.Link.Directed` / `InverseKey` | The same rules, moved to the layer that owns them; the unread cache fields go (D-3, D-7) |
| `core/segmentcompare/segmentcomparetest` *(new)* `Terms` | fixture | `subjectcomparetest.Metas` | Same shape for bridge terms (D-1) |
| `vocabkey.BridgeKey`, `BridgeTerm`, `BridgeKey.String` *(new)*; constants `BridgeParticipation`, `BridgeRelationship`, `BridgeLocation`, `BridgePlaceRelationship`, `PropertyPerson`, `PropertyEvent`, `PropertyPlace`, `PropertyRelatedTo`, `PropertyFrom`, `PropertyTo`, `TermPartOf` *(new)* | type, const | `graphalign.EdgeSignature`; `vocabkey.Role`, `EventType`; literals in `connectrules.productBridges`; `propertyterms.KeyPartOf`, `conclusionheaders.RelPartOf` | The key home (D-2); `connectrules` and `KeyPartOf` read the constants; `RelPartOf` moves when next edited (Out) |
| `matchregistry.Registry.Segment`, `core/matchregistry/segment.go` *(new)* `defaultSegment` | field, file | `subject.go` `defaultSubject` | Same shape, one file per section ([match registry](../comparison-layers.md#match-registry)) |
| `graphalign.Config.Segment` *(new field)*; `Bridge.KeyA/KeyB`, `CanonEdge.KeyFrom/KeyTo`, `Via.Key` *(new fields)* | field | `Config.SignatureScale`, `EdgeSupport*`; `Signature` fields | Replace them (D-6) |
| `state.segments`, `state.newPairSegments` *(new)* | method | `edgeFacts`, `newPairFacts`, `neighborsMatch`, `pairNode` | Rewrites on the layer; the pair is scored once (D-6) |
| `promotealign` `loadTerms`, `linkOf`, `canonKeys`, `edgeSeenKey` *(new)* | func | `loadKinship`, `edgeStamp`, `canonSignature`, `bridgeKey` | Same jobs on the typed key, one builder for both sides (D-7) |
| `TestKey`, `TestCompare`, `TestInvariants` in `core/segmentcompare/segmentcompare_test.go` *(new)*; `TestBridgeKeyString` in `core/vocabkey/vocabkey_test.go`; `TestSegmentSection` in `core/matchregistry/segment_test.go` *(new)*; `TestSegmentSectionMatchesSeed` in `core/database/subjectvocab/segment_section_test.go` *(new)*; `TestBirthPlacedApartSubtracts`, `TestResidencePlacedApartIsUnknown`, `TestSpouseNeverCorrespondsToSibling`, `TestOtherLinkCreditStaysUnderTheBar` (renamed) in `align_test.go`; `TestLayerBridgesCoverEveryConnectRule`, `TestProposeBirthOnAnotherPersonSubtracts`, `TestProposePlaceThroughALocation` in `promotealign_test.go`; `LayerBridges` in `export_test.go` | test, helper | `TestSubjectSection`, `TestSubjectSectionMatchesSeed`, `subjectcompare` `TestCompare` / invariants, `LoadStatsForTest` | Same table shapes |

**Builds on.** `subjectcompare.Compare` and the subject section (S11-10), the handle profiles S11-13 builds, `connectrules` and `hops.NewHop`, the `add-match-rule` skill, the guard and inventory tests.
**Deletes.** `graphalign` `EdgeSignature`, `Stats`, `SignatureScale`, `EdgeFact`, `EdgePoints`, `corresponds`, `pairGroup`, `nodeScore`, `pairNode`, `neighborsMatch`, the seven edge `Config` fields; `promotealign` `stats.go`, `LoadStatsForTest`, `kinship`, `loadKinship`, `edgeStamp`, `canonical`, four of five hop vars, `canonSignature`, `bridgeKey`; `graphcache.Link.Directed`, `Link.InverseKey` and their term-mark columns; the `connectrules` literals; tests `TestEdgePointsHonorsTheSignatureScale`, `TestDefaultPartOfScale`, `TestStatsCacheInvalidatesOnWrite`, `TestStatsFanOut`; inventory rows `segment.Cardinality.*`, `segment.PartOf.*`; `segment.` from `notYetBuilt`.
**Next adds.** S11-16 adds a third `Placement` for unplaced neighbors (entity compare decides) and calls `Compare` around accepted pairs; bridge-property comparison is a `Segment` field and subject section weights for a bridge type, if a golden needs it (D-10). A new axiomatic one-cardinality link is one `One` entry; a new gate one `Gates` entry; a change to what corresponds is `Corresponds`. S11-17 sends the cause codes to Swift. S11-19 bundles a corresponding segment's evidence from `Result.Neighbor`. S11-15's shortlist index takes `graphcache`'s stats slot.

### Schema

None — no catalog change.

### Go core

| Package | Change |
| --- | --- |
| `core/vocabkey` | `BridgeKey`, `BridgeTerm`, `String`, constants (D-2) |
| `core/connectrules` | `productBridges` reads the `vocabkey` constants (D-2) |
| `core/database/propertyterms` | `KeyPartOf = string(vocabkey.TermPartOf)` (D-2) |
| `core/segmentcompare` *(new)* | `segmentcompare.go` (new), `key.go` (new), tests; `segmentcomparetest/terms.go` (new) (D-1, D-3, D-4) |
| `core/matchregistry` | `registry.go` `Segment` field; `segment.go` (new); `segment_test.go` (new); `inventory_test.go` `notYetBuilt`; `guard_test.go` `layerPackages` (D-5) |
| `core/graphalign` | `types.go`, `registry.go`, `score.go`, `align.go`; tests (D-6, D-11) |
| `core/database/promotealign` | `layer.go`, `canon.go`, `terms.go`, `promotealign.go`; `stats.go` deleted; tests (D-7) |
| `core/database/graphcache` | `graph.go` `Link` fields and docs, `load.go` term marks (D-7) |
| `core/database/subjectvocab` | `segment_section_test.go` (new); `TestVocabKeysMatchSeed` rows (D-2, D-5) |
| `api/ffi/handlers` | `promote_align.go` via fields (D-8) |

```go
// vocabkey

type BridgeTerm string

type BridgeKey struct {
	Type      SubjectType // the bridge type
	Term      BridgeTerm  // canonical direction; empty when the bridge has no term
	End       PropertyKey // the endpoint read from; empty for a symmetric link between two of one type
	EventType EventType   // the event on a participation; empty otherwise
}

func (k BridgeKey) String() string

// segmentcompare

type Link struct {
	Type       vocabkey.SubjectType
	Term       vocabkey.BridgeTerm
	End, Other vocabkey.PropertyKey
	EventType  vocabkey.EventType
	SameType   bool
}
type Term struct {
	Directed bool
	Inverse  vocabkey.BridgeTerm
}
type Terms map[vocabkey.SubjectType]map[vocabkey.BridgeTerm]Term

func Key(l Link, terms Terms) vocabkey.BridgeKey
func Corresponds(a, b vocabkey.BridgeKey) bool

type Segment struct {
	Key          vocabkey.BridgeKey
	NeighborKind vocabkey.SubjectType
	Neighbor     subjectcompare.Values
}
type Placement string
type Cause string
type Gate struct{ MinOwnEvidence, ConflictFactor float64 }
type Rules struct {
	One                                []vocabkey.BridgeKey
	OneSupport, OneCredit, OneConflict float64
	OtherSupport, OtherCredit          float64
	Gates                              map[vocabkey.BridgeKey]Gate
}
type Result struct {
	Outcome  propertycompare.Outcome
	Cause    Cause
	Evidence float64
	Neighbor subjectcompare.Result
}

func Compare(a, b Segment, placed Placement, own float64, metas subjectcompare.Metas, rules Rules, subject subjectcompare.Rules, value valuecompare.Rules) Result

// graphalign

func Align(layer Layer, canon Canon, fixed []Fixed, cfg *Config) Proposal

// promotealign (unexported)

func linkOf(b connectrules.Bridge, term, eventType string, firstEnd bool) segmentcompare.Link
func loadTerms(g *graphcache.Graph) (segmentcompare.Terms, error)
func canonKeys(self []byte, selfKind string, n *graphcache.Node, l graphcache.Link, neighbor *graphcache.Node, props map[string]graphcache.VocabProperty, terms segmentcompare.Terms) (from, to []byte, keyFrom, keyTo vocabkey.BridgeKey, ok bool)
func edgeSeenKey(a, b []byte, keyA, keyB vocabkey.BridgeKey) string
```

`Rules` fields carry the doc comments the inventory's "Changes" column states. `Gates` doc: "per key, a minimum own evidence before support counts and a factor of own evidence lost when placed apart; a missing key is the zero gate".

### FFI

None — `api/proto/engine.proto` is unchanged; the handler fills the same fields from the key (D-8).

### macOS

None — no Swift change.

### Strings

None — no copy.

## 7. Steps

**STEP-1 — Typed bridge keys**
Files: `core/vocabkey/vocabkey.go`, `core/vocabkey/vocabkey_test.go`, `core/database/subjectvocab/vocabkey_test.go`, `core/connectrules/connectrules.go`, `core/database/propertyterms/propertyterms.go`. Skill: `add-match-rule` ("Add a Subject property or pair", step 1).
Do: D-2. `vocabkey_test.go` gains `TestBridgeKeyString` *(new)* (T-1).
Check: `go test ./core/vocabkey/... ./core/connectrules/...` and `CGO_ENABLED=1 go test -tags fts5 ./core/database/subjectvocab/ ./core/database/propertyterms/...` pass.

**STEP-2 — Segment compare**
Files: `core/segmentcompare/segmentcompare.go`, `key.go`, `segmentcompare_test.go`, `segmentcomparetest/terms.go` (all (new)); `core/matchregistry/guard_test.go`. Skill: `add-match-rule` ("Add a layer package").
Do: D-1, D-3, D-4. Add `{"core/segmentcompare", true}` and `{"core/segmentcompare/segmentcomparetest", false}` to `layerPackages` after `core/subjectprofile`. T-2 to T-4 build their segment rules as a literal `segmentcompare.Rules` in the test file with the D-5 numbers; STEP-3 moves them onto the registry.
Check: `go test ./core/segmentcompare/... ./core/matchregistry/...` passes.

**STEP-3 — Registry segment section**
Files: `core/matchregistry/registry.go`, `segment.go` (new), `segment_test.go` (new), `inventory_test.go`; `core/database/subjectvocab/segment_section_test.go` (new); `docs/matching-model.md`; `core/segmentcompare/segmentcompare_test.go`. Skill: `add-match-rule` ("Add a knob").
Do: D-5. Delete `"segment."` from `notYetBuilt`. In the knob inventory, replace `segment.Cardinality.birth`, `segment.Cardinality.death` and the three `segment.PartOf.*` rows with: `segment.One` (`participation·subject·birth`, `participation·subject·death`; "decided 2026-10-17, both ends 2026-10-19"; "Links where a differing neighbor conflicts, from either end"), `segment.OneConflict` (3; the D-5 estimate; "Points lost when a one-cardinality link's neighbor is placed elsewhere"), `segment.Gates.place_relationship·part_of·from.MinOwnEvidence` (2) and `.ConflictFactor` (1), `segment.Gates.place_relationship·part_of·to.MinOwnEvidence` (2) and `.ConflictFactor` (0; source "decided: the whole's end had no conflict factor"; "a different whole is silent"), each other row with today's source column. Replace the literal rules in `segmentcompare_test.go` with `matchregistry.Default().Segment`.
Check: `go test ./core/matchregistry/... ./core/segmentcompare/...` and `CGO_ENABLED=1 go test -tags fts5 ./core/database/subjectvocab/` pass.

**STEP-4 — Walk on segments**
Files: `core/graphalign/types.go`, `registry.go`, `score.go`, `align.go`, `align_test.go`, `pick_internal_test.go`. Skill: `add-match-rule`.
Do: D-6, D-9, D-11. Add T-5 to T-7.
Check: `go test ./core/graphalign/...` passes (the tree does not build outside `core/graphalign` until STEP-5; D-13).

**STEP-5 — Loader, cache and wire**
Files: `core/database/promotealign/layer.go`, `canon.go`, `terms.go`, `promotealign.go`, `export_test.go`, `promotealign_test.go`; delete `stats.go`; `core/database/graphcache/graph.go`, `load.go`; `api/ffi/handlers/promote_align.go`.
Do: D-7, D-8. Before deleting `stats.go`, run `rg -n 'loadStats|LoadStatsForTest|computeStats|graphalign.Stats' core api`; before deleting the `Link` fields, run `rg -n '\.Directed\b|InverseKey' core api --glob '!core/database/graphcache/vocab.go'`; a hit outside the files this step names, other than `VocabTerm` and `propertyterms` fields, is a stop condition.
Check: `go build ./... && go vet ./... && CGO_ENABLED=1 go test -tags fts5 ./core/database/promotealign/... ./core/database/graphcache/... ./api/ffi/...` passes; AC-4.4's grep prints nothing.

**STEP-6 — Slice Check in the catalog**
Files: `core/database/promotealign/promotealign_test.go`, `export_test.go`.
Do: T-8, T-9 and T-10, and the `Via.Key` assertion in `TestProposeRelationshipFromEitherEnd` (AC-4.2). `export_test.go` gains `LayerBridges(c *database.Catalog, sourceID []byte) ([]graphalign.Bridge, error)`: `db, _ := c.DB()`; `rules, namePattern, err := valuerules.Load(c.Graph(), db)`; `cfg := graphalign.DefaultConfig()`; `cfg.Value = rules`; `layer, _, err := loadLayer(c.Graph(), db, sourceID, cfg, namePattern)`; return `layer.Bridges` (errors returned at each call).
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/promotealign/...` passes.

**STEP-7 — Skill and docs**
Files: `.agents/skills/add-match-rule/SKILL.md`, `.cursor/rules/match-rules.mdc`, `docs/matching-model.md`, `docs/promote-graph-alignment.md`, `docs/deployment-plan/spike-11/comparison-layers.md`, `completed.md`, `deployment-plan.md`, this plan.
Do: D-12; the Definition of done's docs lines (the map row's Status is the only map edit).
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-11/prs/S11-14-segment-compare.md` reports no errors; `go test ./core/matchregistry/...` passes (the inventory test reads `matching-model.md`); `scripts/check-pr.sh` prints `check-pr: PASS` on macOS.

Commits: one per behavior (STEP-4 and STEP-5 together, D-13). The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log`. No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/vocabkey/vocabkey_test.go` `TestBridgeKeyString` *(new)* | `{participation, subject, person, birth}` → `participation·subject·person·birth`; `{relationship, spouse, "", ""}` → `relationship·spouse` | D-2 |
| T-2 | Go | `core/segmentcompare/segmentcompare_test.go` `TestKey` *(new)* | AC-1.1 to AC-1.4, one row each, over `segmentcomparetest.Terms()` | AC-1.1, AC-1.2, AC-1.3, AC-1.4, D-3 |
| T-3 | Go | same file, `TestCompare` *(new)*, rows named by AC | AC-1.5, AC-2.1 to AC-2.10, one row each | AC-1.5, AC-2.1, AC-2.2, AC-2.3, AC-2.4, AC-2.5, AC-2.6, AC-2.7, AC-2.8, AC-2.9, AC-2.10, D-4 |
| T-4 | Go | same file, `TestInvariants` *(new)*, fixed seed, 500 random segment pairs from the fixtures | Non-corresponding keys → evidence 0; together and ungated, evidence equals the key's support + credit × `Neighbor.Evidence`; apart on a key in neither `One` nor `Gates` → 0; a `One` fact read from either end gives the same apart evidence | D-4, D-5 |
| T-5 | Go | `core/graphalign/align_test.go` `TestBirthPlacedApartSubtracts` *(new)* | AC-3.1 | AC-3.1, D-5 |
| T-6 | Go | same file, `TestResidencePlacedApartIsUnknown` *(new)* | AC-3.2 | AC-3.2 |
| T-7 | Go | same file, `TestSpouseNeverCorrespondsToSibling` *(new)* | AC-3.3 | AC-3.3 |
| T-8 | Go | `core/database/promotealign/promotealign_test.go` `TestLayerBridgesCoverEveryConnectRule` *(new)* | AC-4.1 | AC-4.1, D-7 |
| T-9 | Go | same file, `TestProposeBirthOnAnotherPersonSubtracts` *(new)* | AC-4.3 | AC-4.3 |
| T-10 | Go | same file, `TestProposePlaceThroughALocation` *(new)* | AC-4.5 | AC-4.5, D-7 |
| T-11 | Go | `core/matchregistry/segment_test.go` `TestSegmentSection` *(new)* | `Default().Segment`: every `One` key is a participation key with `Term` `subject` and an empty `End`; every `Gates` key's `Type` is a `connectrules` bridge type; all support and credit numbers are positive | AC-5.1, D-5 |
| T-12 | Go | `core/database/subjectvocab/segment_section_test.go` `TestSegmentSectionMatchesSeed` *(new)* | AC-5.2 | AC-5.2 |
| T-13 | Go | `core/matchregistry/guard_test.go` `TestLayerPackages`; `inventory_test.go` `TestKnobInventoryMatchesRegistry` | Pass with the new rows and section | AC-5.1, AC-5.3 |
| T-14 | Go | Ported: `TestSparseNameAndBirthDate`, `TestProposeRelationshipFromEitherEnd` (with its `Via.Key` assertion), the rest of `core/graphalign` and `core/database/promotealign`, `api/ffi/handlers` `TestProposePromoteGraphAlignment` | Pass under D-11 | AC-3.4, AC-3.5, AC-4.2, D-8, D-11 |

AC-4.4 is the STEP-5 build and grep. Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `TestGracieObituaryGolden`, the `part_of` tests in both packages, `TestSeedFromNewlyFixedIntoUnpromoted` (keeps its numbers: a birth link is a `One` fact).

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left — none in this PR
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New — no UI in this PR
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together) — or none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan and every reshaped fixture (D-11)
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated (if gated) — not gated
- [ ] Domain doc updated if a rule in docs/*.md changed: matching-model.md, promote-graph-alignment.md, comparison-layers.md §Segments and §Match registry (D-12)
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

`completed.md` gets, above the newest entry: "## S11-14 — Segment compare" and "**Landed YYYY-MM-DD** in [#N](https://github.com/mendahu/provenencia/pull/N). Promote judges each link with segment compare. A link is a typed key, so a spouse never stands for a sibling and a parent link reads the same from either record. Subject of a birth or a death is one per person and one per event: a birth placed on a different event, or on a different person, subtracts 3. Every other link earns the same small support whatever the catalog holds; the fan-out scan is gone."

## 10. Manual check

On macOS, in the app:

1. Open a project with two Sources about one family: Source A records "Mary Smith parent of John Smith", both promoted. Source B records "John Smith child of Mary Smith". Open Promote on B and decide Mary onto Mary's handle. John's row proposes John's handle, read "through Mary".
2. On Source B, add a spouse link from Mary to a new person "Robert Smith" while A records Robert as Mary's sibling. Robert's row is not proposed through Mary.
3. Source C records Ann Ames, subject of a birth on 1 May 1817. The research graph holds two Ann Ames, each the subject of a birth on that date. Decide C's birth onto the second Ann's birth. Ann's row on C does not propose the first Ann.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| Spouse and parent links drop from 1.5 + 0.5 to 0.25 + 0.25 on real catalogs, so a family placed by its links files weaker | The inventory's starting values (S11-01); D-11 (c) holds the goldens' Targets; dogfood notes go to the knob inventory, not into this PR |
| Both-end birth conflicts push a decided row negative where the old walk was silent | AC-3.1 pins the number; `ConflictWithFixed` is unchanged (S11-16 replaces it with the contested mark) |
| A plugin bridge walked generically trips `hops.NewHop` | `NewHop` returns its error and `Propose` fails loudly; T-8 covers every product bridge |
| Layer and canon keys drift for one bridge type | One builder (`linkOf`) for both sides; T-10 covers locations, AC-4.2 relationships, AC-3.4 participations |

**Stop and report if:**

- A test named in D-11 (c) changes a Target or handle it asserts.
- A ported number changes for a reason not in D-11 (a) or (b).
- `connectrules.Bridges()` returns a bridge with other than two endpoints whose first endpoint's type has Subjects on the layer.
- `TestProposeRelationshipFromEitherEnd`'s `Via.Key.Term` is not `child` (D-3's canonical choice differs from `kinship.canonical`).
- A STEP-5 `rg` finds a caller outside the files it names.

## 12. Open questions

None.

## 13. PR

**Title:** S11-14 — Segment compare
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-11/prs/S11-14-segment-compare.md`.
