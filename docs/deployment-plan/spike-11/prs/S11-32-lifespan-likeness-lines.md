# S11-32 — Lifespan and likeness evidence lines

**Status:** Draft
**Spike / slice:** Spike 11 · Slice 5 — Claim evidence
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no new chrome: today's sheet words a likeness value as "Likeness" in its existing columns and does not draw the lifespan; S11-20 draws both from S11-D2)
**Board:** —
**Depends on:** S11-25 (landed [#517](https://github.com/mendahu/provenencia/pull/517)), S10-54 (landed [#433](https://github.com/mendahu/provenencia/pull/433))
**Blocks:** S11-20
**Branch:** `s11-32-lifespan-likeness-lines`
**Version bump:** none — Spike 11 PRs leave `VERSION` at 0.2.0, and the product bump is cut when a release ships ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-match-rule`](../../../../.agents/skills/add-match-rule/SKILL.md), [`add-depiction-value`](../../../../.agents/skills/add-depiction-value/SKILL.md), [`add-ffi-handler`](../../../../.agents/skills/add-ffi-handler/SKILL.md), [`add-localized-string`](../../../../.agents/skills/add-localized-string/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md)
**Rules that bind:** `.cursor/rules/match-rules.mdc`, `.cursor/rules/depiction-values.mdc`, `.cursor/rules/go-tests.mdc`, `.cursor/rules/ffi-handlers.mdc`
**Context:** map scope entry [S11-32](../deployment-plan.md#s11-32--lifespan-and-likeness-evidence-lines) (the contract S11-20 builds on); [`claim-evidence.md`](../claim-evidence.md) (Lines); [`comparison-layers.md`](../comparison-layers.md#lifespan-a-derived-property-for-people) (lifespan), §Detail (causes); [S11-12 plan](S11-12-lifespan.md) D-12 (when a side has no lifespan); [S11-25 plan](S11-25-claim-evidence-rpc.md) D-2 – D-4 and its Landed note; [S11-D2](../design/S11-D2-evidence-sheet.md) §5 (the example project these lines feed); working rules "The pin policy has one owner, in Go", "One wire vocabulary", "Layers and packages", "High tolerance for churn"
**Planned:** 2026-10-22 against `main` @ `d81808f`
**Review:** — (not reviewed yet)

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

The claim evidence S11-19 and S11-25 built leaves out two things the evidence sheet must show. A person's **lifespan** is compared in matching but never reaches the sheet, and a side whose own dates contradict each other has no lifespan with no reason given. **Likenesses** never reach the evidence at all: the loader drops the depiction, and value compare has no comparer for it. This PR carries both through the engine and the wire, and lets Go order the sheet's sections, so S11-20 can draw S11-D2 with no engine work.

**Why now.** S11-20 (the sheet) is planned against this PR's contract and waits on it (map, user 2026-10-22).

**What stays.** Matching: likeness has no weight, earns no points and is never a conflict; bands, candidates, contested and moves do not change. The lifespan's place in matching (S11-12, S11-13). Apply and the pin write (S11-25): likeness pins come from `PinOffers` like any line. Today's sheet layout (S11-25 D-14). The reconciler's grouping of likenesses: the same image File is the same value.

## 2. User stories

Conventions. Go fixtures: `promotealign_test.go`'s `newFixture`, `newSource`, `person`, `datedBirth`, `participation`, `promote`, `join`; `evidence_test.go`'s `filedBirth`; `observationstest.SaveLikeness(t, c, userID, sourceID, subjectID, shade)`, which saves a crop whose PNG is decided by `shade` (the same shade on two Sources gives one image File: `ingest.StagedImage.InsertTx` finds it by checksum). "Image A" is shade 40, "image B" shade 200. A death fixture is `datedDeath` *(new, beside `datedBirth`)*.

**US-1.** As a researcher, the evidence lists each record's likenesses beside this Source's, so I can judge them by eye.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | A Source person with image A; a handle with two members, one with image A, one with image B | One `likeness` section with two lines: (A, A) relation `same`, cause `exact`, outcome `agree`, checked, pinnable; (A, B) relation `unknown`, cause `unjudged`, outcome `unknown`, pinnable, not checked |
| AC-1.2 | AC-1.1, `PinOffers` | Two offers for the likeness lines: (A, A) checked, (A, B) not checked |
| AC-1.3 | AC-1.1, Done with the row never opened | The new claim pins the incoming likeness and the image-A member's likeness; the image-B member's is not pinned |
| AC-1.4 | Three members with image A | One collapsed line naming three pins |
| AC-1.5 | A member whose likeness has a refused image (not PNG, so `Readable` empties it) | No line for that member |
| AC-1.6 | The same Source person and handle proposed with and without the likenesses | The same band, score and contested flag |

**US-2.** As a researcher, the evidence carries the lifespan check, and says when a side has none and why.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | A Source person born 14 May 1817 and dead 2 Jun 1890; a handle whose member is subject of a birth on 14 May 1817 and a census on 2 Apr 1871 | `Evidence.Lifespan`: outcome `agree`; row side first date 14 May 1817, last 2 Jun 1890, no reason; record side first 14 May 1817, last 2 Apr 1871 |
| AC-2.2 | The Source person dead "2 Jun 1809" (before the birth) | Row side: no dates, reason `contradicts`; outcome `unknown` |
| AC-2.3 | The handle with no dated events | Record side: reason `undated`; outcome `unknown` |
| AC-2.4 | Neither side has a dated event | `Evidence.Lifespan` is nil |
| AC-2.5 | The record buried 9 Jan 1838 and the Source person dead 1890 | Outcome `conflict` |
| AC-2.6 | An event or place row | `Evidence.Lifespan` is nil |
| AC-2.7 | A side with one dated event (married 1843) | First and last are both that date |
| AC-2.8 | The RPC for AC-2.1 | `lifespan` with both sides' `first` and `last` as dates, and `none` empty; it adds no offer |

**US-3.** As the sheet, sections arrive in the order the evidence section declares.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | A person with `name`, `likeness`, `sex_at_birth` and `interests` sections | Order `name`, `likeness`, `interests`, `sex_at_birth` |
| AC-3.2 | An event with `date` and `event_type` sections | Order `event_type`, `date` |

**US-4.** As a contributor, the new codes cross the wire once.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-4.1 | `TestSwiftVocabularyMatchesGo` | `ValueCause` lists `unjudged`; `LifespanGap` lists `""`, `undated`, `contradicts` |
| AC-4.2 | A likeness line's value on the wire, in Swift | `.depiction(CatalogDepictionValue)` with the image's relative path; today's sheet words it "Likeness" |

## 3. Scope

| In | Covered by |
| --- | --- |
| Depiction comparer, `Value.Image`, cause `unjudged` (layer 1) | STEP-1, D-1, D-2 |
| The reconciler keys depictions through `valuecompare.Key` | STEP-2, D-3 |
| `subjectprofile` carries the image and reports each side's lifespan gap and first and last dated events | STEP-3, D-4, D-5 |
| `claimevidence`: `Rules.Order`, `Rules.ByEye`, the lifespan result, registry defaults and inventory rows | STEP-4, D-6, D-7, D-8 |
| Loader: decode the depiction, the row's and the record's life, `Evidence` sets the lifespan | STEP-5, D-9, D-10 |
| Wire: `ConclusionValue.depiction`, `PromoteClaimEvidenceLifespan`, handler mapping and paths | STEP-6, D-11, D-12 |
| Swift: vocabulary, store types and mapping, `ReconciledValueDisplay` | STEP-7, D-13 |
| Docs | STEP-8 |

| Out | Home |
| --- | --- |
| Drawing the lifespan and the crops, the new copy, section order on screen | S11-20 |
| Face comparison, or any likeness weight in matching | not planned |
| Sending a depiction's region on the evidence wire | not planned (D-11: the sheet draws the stored crop) |
| A lifespan on members (each member's own) | not planned (the line compares the row with the record's profile) |
| Lifespans for events or places | not planned (working rule "Lifespan is for people only") |
| Recording S11-32's names in S11-20's plan | S11-20 STEP-0 |

## 4. Current state

Verified at `d81808f`.

| Area | Where | Today |
| --- | --- | --- |
| Layer 1 | `core/valuecompare/valuecompare.go` `Value` (L151–162: `Type, Text, Term, Integer, Name, NamePattern, Date, Ref`), `comparers` (L171–177: text, term, integer, date, name), `Compare` (L181), `Key` (L201), cause block (L54–83, last `CauseUnsupported`); `term.go` (the 20-line comparer shape) | No depiction field and no comparer: a depiction is `Unknown` / `unsupported` |
| Layer 1 tests | `contract_test.go` `samples` (L32–89), `generators` (L206–233), `TestComparerContract` (L318, `allTypes` includes `ValueTypeDepiction`), `TestKeyContract` (L402); `valuecompare_test.go` `TestCompareUnsupported` (L13–39: `ComparerTypes()` is date, integer, name, term, text; depiction and subject unsupported) | — |
| Layer 2 | `core/propertycompare/propertycompare.go` `Compare` (L127–168), `outcomeOf` (L97–107: `Unknown` → `OutcomeUnknown`); a multiple property's conflict becomes `OutcomeUnknown` / `CauseCardinality` (L164–166) | An `Unknown` relation gives `OutcomeUnknown` / `CauseRelation` **with** `Speaking` set |
| Reconciler | `core/autoreconcile/modules.go` `moduleFor` (L49–75: depiction keyed by `idKey(v.Depiction.ImageFileID)`), `compareValue` (L77–79: text, integer, date only), `idKey` (L117); `autoreconcile.go` `Value.Depiction *depictionvalue.Value` (L87–89) | A private key rule beside value compare's |
| Profile | `core/subjectprofile/subjectprofile.go` `LifeEvent` (L30), `Subject` (L38), `Build` (L74), `ForSubject` (L126), `addLifespans` (L141–149: silent on failure), `lifespan` (L158–215: not ok at L184 no dated event, L191 and L197 a year shift fails, L201–205 a date outside the window, L206 and L210 `EndsDate` refuses), `Value` (L220–236: copies text, term, integer, name, date; not `Depiction`); `lifespan_test.go` `TestLifespanFromEvents` (AC-2.1 – AC-2.18) | No reason kept; the span's first and last dates are not exposed |
| Subject compare | `core/subjectcompare/subjectcompare.go` `Lifespan` (L45–50), `Rules` (L53), `Compare` (L103), `PropertyResult` (L91–95), `Derived` (L154); `lifespan.go` | `Compare` returns the lifespan lead's `PropertyResult` |
| Evidence | `core/claimevidence/claimevidence.go` `Rules` (L20–29: `Checked`, `Pinnable`, `Defining`), `Evidence` (L92–97: `Kind`, `Sections`, `Relationships`, `Sources`); `sections.go` `Sections` (L61–100: skips no meta L66, no ref L71, `OutcomeUnknown` L84; sections sorted by key L96–98), `line` (L104–121); `build.go` `Build` (L24–42; L28 keeps only `subjectcompare.Compare(...).Evidence`) | Unknown lines dropped; no order rule; no lifespan |
| Evidence tests | `sections_test.go` `TestSections` (AC-1.1 – AC-1.12), `build_test.go` `TestBuild` (AC-2.1 – AC-2.6, helper `scene.build`), `invariants_test.go` `TestSectionsInvariants` | — |
| Registry | `core/matchregistry/evidence.go` `defaultEvidence` (L11–22), `evidence_test.go` `TestEvidenceSection`, `inventory_test.go` `TestKnobInventoryMatchesRegistry`; `docs/matching-model.md` L200–202 | — |
| Loader | `core/database/promotealign/profiles.go` `decode` (L54–77, no depiction), `reconciled` (L80, ignores `graphcache.Value.ImageFileID`), `layerValues` (L93–120, ignores `EvidenceObservation.DepictionImageFileID`), profiler `build` (L296–319, `e.Life` from `hopPersonToEvent`), `done map[string]subjectcompare.Values`; `layer.go` `primarySubject{id, ref, kind}` (L19–23), `layerMemo{layer, primary}` (L36), `forSubject` with `life` (L101–127); `evidence.go` `evidence` (L56–183, `claimevidence.Build` at L176–178), `memberValues` (L275–334, `ForSubject(..., nil)`); `values.go` `EvidenceValue`, `value` (L85–100); `pins.go` `PinOffers` (L38–70) | Likeness dropped at decode; life events not kept after building values |
| Cache | `core/database/graphcache/evidence_graph.go` `EvidenceObservation.DepictionImageFileID` (L66, already through `Readable`); `graph.go` `Value.ImageFileID` (L94) | Both in hand |
| Wire | `api/proto/engine.proto` `ConclusionValue` (L1404–1412, next field 6), `DepictionValue{region_json = 1, image_rel_path = 2}` (L1962), `GetPromoteClaimEvidenceResponse` (L1597–1602, next field 5); `api/ffi/handlers/promote_evidence.go` `claimEvidenceProto` (L64), `evidenceSectionsProto` (L104), `evidenceValueProto` (L140–155); `conclusion_detail.go` `conclusionValueProto` (L140–158); `core/database/conclusiondetails/conclusiondetails.go` `Value` (L44–54); `observations.go` `depictionPaths` (L400–408) over `files.RelPaths` (`core/database/files/files.go:92`) | A depiction arrives as an empty `ConclusionValue` |
| Wire tests | `api/ffi/handlers/promote_evidence_test.go` `TestGetPromoteClaimEvidence`, `newEvidenceFixture`; `api/ffi/swift_vocabulary_test.go` `TestSwiftVocabularyMatchesGo` (value causes row L284–291) | — |
| Swift | `macos/App/Platform/GenealogyStore.swift` `CatalogConclusionValue` (L638–645), `CatalogDepictionValue` (L487), `CatalogClaimEvidence` (L406–418); `GoStore.swift` `getPromoteClaimEvidence` (L1034–1074), `mapConclusionValue` (L1429–1438, exhaustive); `ConclusionVocabulary.swift` `ValueCause` (L164–218); `macos/App/Features/Conclusions/ReconciledValueDisplay.swift` `string(for:locale:orders:)` (L78–97, the one exhaustive switch over `CatalogConclusionValue`) | — |

**Precedent.** S11-25 ([#517](https://github.com/mendahu/provenencia/pull/517)) carried evidence over the wire: copy its proto, handler, `GoStore` mapping and vocabulary-row shape. S11-12 ([#480](https://github.com/mendahu/provenencia/pull/480)) built the lifespan in `subjectprofile`: extend `lifespan` in place, its tests table-driven as `TestLifespanFromEvents` is. S10-54 ([#433](https://github.com/mendahu/provenencia/pull/433)) sent `DepictionValue` with paths through `files.RelPaths`.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | **A depiction comparer.** `valuecompare.Value.Image []byte` *(new)*: the depiction's image File id, opaque, like `Ref`. `core/valuecompare/depiction.go` *(new)* `type depiction struct{}`: an empty `Image` on either side → `Unknown` / `CauseEmpty`; equal bytes → `Same`, similarity 1, `CauseExact`; otherwise `Unknown`, similarity 0, `CauseUnjudged` *(new)* `"unjudged"` ("two values the layer cannot relate: the researcher judges them"). `key`: `string(v.Image)`, ok when non-empty, at `Same` and `Equivalent`. Registered under `propertymeta.ValueTypeDepiction` | The same File is the same pixels (the reconciler's rule); two crops can show one face and the app cannot tell. `Unknown` keeps matching untouched | `Near` — a judgement the layer cannot make; `Different` — two photos of one person are not a disagreement |
| D-2 | **The relation set does not change.** `CauseUnjudged` is a cause, not a relation; `outcomeOf` and `contradicts` are untouched; a negative likeness never contradicts by a different image | `TestEveryRelationHasAnOutcome` unchanged; layer 2 stays relation-driven | A new relation — every switch over relations grows |
| D-3 | **The reconciler keys depictions through layer 1.** `Value.compareValue` sets `Image` from `v.Depiction.ImageFileID` when `Depiction` is not nil; `moduleFor`'s depiction case joins text and integer's `valuecompare.Key(..., Same, ...)` case; the depiction-specific closure goes. Keys are byte-identical to `idKey`'s, so grouping is unchanged and `autoreconciler.CacheVersion` is not bumped | One key rule per value type (reuse table in `comparison-layers.md`); "Build, then migrate" | Leaving `idKey` for depictions — a second rule beside the new comparer |
| D-4 | **Profiles carry the image.** `subjectprofile.Value` copies `v.Depiction.ImageFileID` into `Image` when `v.Depiction` is not nil | One conversion path into layer 2 | A likeness field outside `valuecompare.Value` |
| D-5 | **A side says why it has no lifespan.** In `core/subjectcompare/subjectcompare.go`, beside `Lifespan`: `type LifespanGap string` with `LifespanUndated = "undated"` and `LifespanContradicts = "contradicts"` *(new)*, and `type LifeSpan struct{ First, Last *datevalue.Value; Gap LifespanGap }` *(new)*: the dates of the earliest and the latest dated event the span runs between, or the gap. `subjectprofile.lifespan` returns `(window, span propertycompare.Value, life subjectcompare.LifeSpan)`: no dated event (L184) → `undated`; a year shift that fails (L191, L197) → `undated` (no window can be built from those dates); a date outside the window (L201–205) or an `EndsDate` refusal (L206, L210) → `contradicts`. `First` is the date whose early end is the span's first end, `Last` the date whose late end is its last, each the first such in event order; with one dated event they are the same date. `subjectprofile.LifeOf(events []LifeEvent, lifespans map[vocabkey.PropertyKey]subjectcompare.Lifespan) subjectcompare.LifeSpan` *(new)* returns the life of the kind's one lifespan lead (the zero value when `lifespans` is empty); `addLifespans` and `LifeOf` share `lifespan` | `claimevidence` cannot import `subjectprofile` (only `subjectprofile` may reach `autoreconcile`, `TestLayerPackages`); `subjectcompare` already owns `Lifespan`. Original dates, not ends, keep "abt 1815" readable on the sheet (S11-20 D-13) | A reason inside `Values` — a marker value every reader must skip; ends — lose the qualifier |
| D-6 | **Section order is a registry rule.** `claimevidence.Rules.Order map[vocabkey.SubjectType][]vocabkey.PropertyKey` *(new)*: keys first in this order, then the rest ascending. `Sections` sorts its section keys by it. Default: person `name`, `likeness`; event `event_type`. Inventory row `evidence.Order` | One owner of order (S11-20 review M-2); relationships' sections use the neighbor kind's order | Order in Swift |
| D-7 | **Lines the researcher judges.** `claimevidence.Rules.ByEye []valuecompare.Cause` *(new)*: an `OutcomeUnknown` line whose speaking cause is listed is kept, pinnable and never checked. `Sections` L84 keeps such lines; `line` sets `Pinnable` for them. Default: `CauseUnjudged`. Inventory row `evidence.ByEye`. A line with no speaking pair or no member ref is still dropped | Policy in the evidence section, not a likeness special case; user, 2026-10-22: two images are shown, pinnable and unchecked | `if key == "likeness"` in `Sections` |
| D-8 | **The lifespan result.** `claimevidence.Lifespan{Outcome propertycompare.Outcome; PropertyCause propertycompare.Cause; Relation valuecompare.Relation; Cause valuecompare.Cause; Row, Record subjectcompare.LifeSpan}` *(new)*; `Evidence.Lifespan *Lifespan` *(new)*. `Build` gains a last parameter `life Lives` *(new type `Lives{Row, Record subjectcompare.LifeSpan}`)*, keeps the whole `subjectcompare.Compare` result, and when `subject.Lifespans[kind]` has a lead takes that lead's `PropertyResult`: outcome, property cause, and the speaking pair's relation and cause (empty when none). `Evidence.Lifespan` is nil for a kind with no lifespan lead, and when neither side has `First` and neither side's gap is `contradicts` | The lifespan line belongs with the rest of the evidence and needs no member; matching's own comparison is reused, not recomputed | A second lifespan compare in the loader |
| D-9 | **The loader decodes depictions.** `decode` gains a last parameter `image []byte`; a non-empty image sets `Value.Depiction = &depictionvalue.Value{ImageFileID: <a copy>}`. `layerValues` passes `o.DepictionImageFileID`, `reconciled` passes `row.ImageFileID` | Both are already `Readable` (graphfill `FromJoin`); one decoder | A second read of `depiction_values` |
| D-10 | **The loader keeps each side's life.** `primarySubject` gains `life subjectcompare.LifeSpan`, set in `layer.go` from `subjectprofile.LifeOf(life[id], cfg.Subject.Lifespans[kind])` where the row's values are built, so it is memoized with the layer. The profiler stores `LifeSpan` per handle beside its values (`done` becomes `map[string]profile` with `profile{values subjectcompare.Values; life subjectcompare.LifeSpan}`, and `values(handle)` keeps its signature; `life(handle)` *(new)* reads it). `evidence` passes `claimevidence.Lives{Row: primary[row].life, Record: q.prof.life(row.HandleID)}` | Read where the events already are, memoized with what they describe (S11-19 D-21, D-22) | Re-walking life events per request |
| D-11 | **The wire.** `ConclusionValue` gains `DepictionValue depiction = 6`; the evidence path fills `image_rel_path` only (region empty: the sheet draws the stored crop). `conclusiondetails.Value` gains `ImageRelPath string` *(new)*; `conclusionValueProto` sets the depiction case when it is not empty, checked before text. `evidenceValueProto` gains `paths map[string]string` and, for `propertymeta.ValueTypeDepiction`, sets `ImageRelPath: paths[string(v.Value.Image)]`. `claimEvidenceProto` collects every `Image` on lines and pins first and calls `files.RelPaths` once | The one value encoder (S11-25 review round 1); one path query per response (S10-54's `depictionPaths` shape) | A per-value path query; the region over the wire |
| D-12 | **The lifespan on the wire.** `PromoteClaimEvidenceLifeSide{DateValueInput first = 1; DateValueInput last = 2; string gap = 3;}` and `PromoteClaimEvidenceLifespan{string outcome = 1; string property_cause = 2; string relation = 3; string cause = 4; PromoteClaimEvidenceLifeSide row = 5; PromoteClaimEvidenceLifeSide record = 6;}` *(new)*; `GetPromoteClaimEvidenceResponse.lifespan = 5`, unset when `Evidence.Lifespan` is nil. Dates through `valuecodec.DateToProto`. No offer | One message, not a section line: it has no member and no pin (scope entry) | A section line with no pins |
| D-13 | **Swift.** `ConclusionVocabulary.swift`: `ValueCause.unjudged` (`"unjudged"`), and `enum LifespanGap { case none, undated, contradicts, unrecognized(String) }` *(new)* with `init(wire:)`. `GenealogyStore.swift`: `CatalogConclusionValue.depiction(CatalogDepictionValue)`; `CatalogClaimEvidenceLifeSide{first: CatalogDateValueInput?, last: CatalogDateValueInput?, gap: LifespanGap}` and `CatalogClaimEvidenceLifespan{outcome: EvidenceOutcome, propertyCause: EvidencePropertyCause, relation: ValueRelation, cause: ValueCause, row, record: CatalogClaimEvidenceLifeSide}` *(new)*; `CatalogClaimEvidence.lifespan: CatalogClaimEvidenceLifespan?`. `GoStore` maps both; `mapConclusionValue` maps `.depiction` with the existing `DepictionValue` mapping. `ReconciledValueDisplay.string(for:)` words `.depiction` as `L10n.Conclusions.likenessValue` *(new)* "Likeness" | One wire vocabulary; today's sheet stays readable with no new chrome | Drawing a crop now — S11-20 |

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `valuecompare` `depiction` comparer, `Value.Image`, `CauseUnjudged` *(new)* | comparer, field, code | `term.go`; `autoreconcile` `idKey` | Layer 1 owns comparers; the reconciler's rule moves onto it (D-3) |
| `subjectcompare.LifespanGap`, `LifeSpan` *(new)* | types | `subjectcompare.Lifespan` | Beside the rule they report on; reachable from `claimevidence` |
| `subjectprofile.LifeOf` *(new)* | func | `addLifespans` | Shares `lifespan`; the loader needs the life without the values |
| `claimevidence.Rules.Order`, `ByEye`, `Lifespan`, `Lives`, `Evidence.Lifespan` *(new)* | fields, types | `Rules.Checked`, `Pinnable` | The evidence section's own policy |
| Loader `primarySubject.life`, profiler `profile`, `life(handle)` *(new)* | fields, func | `layerMemo`, `done` | Kept beside the values they belong to |
| Proto `ConclusionValue.depiction`, `PromoteClaimEvidenceLifespan`, `PromoteClaimEvidenceLifeSide` *(new)* | FFI | `DepictionValue`, `PromoteClaimEvidenceLine` | Reuse `DepictionValue`; the lifespan is not a line |
| `conclusiondetails.Value.ImageRelPath` *(new)* | field | `conclusionValueProto` | The one encoder takes the path |
| Swift `LifespanGap`, `CatalogClaimEvidenceLifespan`, `CatalogClaimEvidenceLifeSide`, `CatalogConclusionValue.depiction`, `L10n.Conclusions.likenessValue` *(new)* | enum, types, case, key | `ValueCause`, `CatalogClaimEvidenceLine` | Same pattern |
| Test fixture `datedDeath` *(new)* | helper | `datedBirth` | A death beside a birth |

**Builds on.** `graphcache`'s `Readable` depiction fields; `files.RelPaths`; `valuecodec.DateToProto`; S11-25's `PinOffers`, `evidenceValueProto`, vocabulary rows; S11-12's `lifespan`.
**Deletes.** `autoreconcile`'s depiction-specific key closure in `moduleFor` (D-3). The "depiction unsupported" assertion in `TestCompareUnsupported`.
**Next adds.** S11-20 draws `lifespan` and `.depiction` with no engine work. A depiction property on another kind is one seeded property; a new judged cause is one `ByEye` entry; a new kind's order is one `Order` entry.

### Schema

None — no catalog change.

### Go core

| Package | Change |
| --- | --- |
| `core/valuecompare` | D-1, D-2 |
| `core/autoreconcile` | D-3 |
| `core/subjectcompare` | D-5's types |
| `core/subjectprofile` | D-4, D-5 |
| `core/claimevidence` | D-6, D-7, D-8 |
| `core/matchregistry` | `defaultEvidence`: `Order`, `ByEye` with `// decided:` comments |
| `core/database/promotealign` | D-9, D-10; `values.go` needs no change (a depiction's type comes from `types`) |
| `core/database/conclusiondetails` | D-11's field |

```go
// package valuecompare
// CauseUnjudged: two values the layer cannot relate (two different images);
// the researcher judges them.
CauseUnjudged Cause = "unjudged"

// package subjectcompare
type LifespanGap string
const (
	LifespanUndated     LifespanGap = "undated"     // no dated event to build one from
	LifespanContradicts LifespanGap = "contradicts" // its own dated events cannot all be true
)
type LifeSpan struct {
	First, Last *datevalue.Value // the earliest and latest dated events; nil with a gap
	Gap         LifespanGap
}

// package claimevidence
type Lives struct{ Row, Record subjectcompare.LifeSpan }
type Lifespan struct {
	Outcome       propertycompare.Outcome
	PropertyCause propertycompare.Cause
	Relation      valuecompare.Relation
	Cause         valuecompare.Cause
	Row, Record   subjectcompare.LifeSpan
}
```

Errors: none new.

### FFI

`api/proto/engine.proto`, regenerated with `./scripts/generate-proto.sh`:

```proto
message ConclusionValue {
  oneof kind {
    // … 1 – 5 unchanged
    DepictionValue depiction = 6;
  }
}

message PromoteClaimEvidenceLifeSide {
  DateValueInput first = 1;
  DateValueInput last = 2;
  string gap = 3; // subjectcompare.LifespanGap
}

message PromoteClaimEvidenceLifespan {
  string outcome = 1;
  string property_cause = 2;
  string relation = 3;
  string cause = 4;
  PromoteClaimEvidenceLifeSide row = 5;
  PromoteClaimEvidenceLifeSide record = 6;
}

message GetPromoteClaimEvidenceResponse {
  // … 1 – 4 unchanged
  PromoteClaimEvidenceLifespan lifespan = 5;
}
```

### macOS

None — no chrome. D-13's types and display only.

### Strings

`L10n.Conclusions.likenessValue` *(new)* = "Likeness" (`conclusions.value.likeness`, comment "A likeness value worded as text where no crop is drawn").

## 7. Steps

**STEP-1 — Depiction comparer**
Files: `core/valuecompare/valuecompare.go`, `depiction.go` *(new)*, `contract_test.go`, `valuecompare_test.go`, `depiction_test.go` *(new)*, `docs/deployment-plan/spike-11/comparison-layers.md` (§Detail cause list). Skill: `add-match-rule` ("Add a comparer", "Add a relation, cause or unit code").
Do: D-1, D-2. Contract sample and generator for depictions (equal and unequal 16-byte ids).
Check: `go test ./core/valuecompare/... ./core/propertycompare/... ./core/matchregistry/...` passes; T-1.

**STEP-2 — The reconciler keys depictions through layer 1**
Files: `core/autoreconcile/modules.go`.
Do: D-3.
Check: `go test ./core/autoreconcile/...` passes unchanged (its depiction cases in `modules_test.go`); `CGO_ENABLED=1 go test -tags fts5 ./core/database/likenesses/ ./core/database/autoreconciler/...` passes.

**STEP-3 — Profiles: image and life**
Files: `core/subjectcompare/subjectcompare.go`, `core/subjectprofile/subjectprofile.go`, `lifespan_test.go`, `subjectprofile_test.go`. Skill: `add-match-rule`.
Do: D-4, D-5. Extend `TestLifespanFromEvents`' table with each case's `LifeSpan`.
Check: `go test ./core/subjectcompare/... ./core/subjectprofile/... ./core/matchregistry/...` passes (the purity guard included); T-2.

**STEP-4 — The evidence section**
Files: `core/claimevidence/claimevidence.go`, `sections.go`, `build.go`, `sections_test.go`, `build_test.go`, `invariants_test.go`, `core/matchregistry/evidence.go`, `evidence_test.go`, `docs/matching-model.md` (rows `evidence.Order`, `evidence.ByEye`). Skill: `add-match-rule`.
Do: D-6, D-7, D-8. `scene.build` passes `claimevidence.Lives{}`.
Check: `go test ./core/claimevidence/... ./core/matchregistry/...` passes; `TestKnobInventoryMatchesRegistry` passes; T-3, T-4.

**STEP-5 — The loader**
Files: `core/database/promotealign/profiles.go`, `layer.go`, `evidence.go`, `promotealign_test.go` (`datedDeath`), `evidence_test.go`, `pins_test.go`. Skill: `use-catalog-session` (read-only paths), `add-depiction-value`.
Do: D-9, D-10.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/promotealign/... ./core/database/promote/...` passes, S11-19 and S11-25 tests unchanged except new fields; T-5, T-6, T-7.

**STEP-6 — The wire**
Files: `api/proto/engine.proto`, regenerated `api/proto/engine/engine.pb.go` and `macos/App/Platform/Generated/engine.pb.swift`, `core/database/conclusiondetails/conclusiondetails.go`, `api/ffi/handlers/conclusion_detail.go`, `promote_evidence.go`, `promote_evidence_test.go`, `api/ffi/swift_vocabulary_test.go`, `macos/App/Platform/ConclusionVocabulary.swift`. Skill: `add-ffi-handler`.
Do: D-11, D-12; the vocabulary rows of D-13 (Go and Swift together, so the test passes).
Check: a second `./scripts/generate-proto.sh` leaves no diff; `CGO_ENABLED=1 go test -tags fts5 ./api/...` passes; T-8, T-9.

**STEP-7 — Swift**
Files: `macos/App/Platform/GenealogyStore.swift`, `GoStore.swift`, `macos/App/Features/Conclusions/ReconciledValueDisplay.swift`, `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`, `macos/ProvenenciaTests/ReconciledValueDisplayTests.swift`. Skill: `add-swift-test`, `add-localized-string`.
Do: D-13.
Check: `xcodebuild test -project macos/Provenencia.xcodeproj -scheme Provenencia -destination 'platform=macOS'` passes; `python3 scripts/check-localizable-xcstrings.py` passes; T-10.

**STEP-8 — Docs**
Files: `docs/deployment-plan/spike-11/claim-evidence.md` (Lines: likenesses are judged by eye, `evidence.ByEye`; the lifespan is carried, never pinned), `docs/deployment-plan/spike-11/comparison-layers.md` (lifespan section: a side's gap is reported; the reuse table: depiction keys in layer 1), `docs/matching-model.md` (rows), `.agents/skills/add-match-rule/SKILL.md` (the evidence-section rule names `Order` and `ByEye`), `docs/deployment-plan/spike-11/deployment-plan.md` (map row), `docs/deployment-plan/spike-11/completed.md`, this plan (status, Landed note).
Do: per the Definition of done.
Check: links resolve; `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-11/prs/S11-32-lifespan-likeness-lines.md` reports no errors.

Commits: one per step, except STEP-6 and STEP-7, which land as one commit: `generate-proto.sh` writes the Go and Swift code together, and Swift builds only once STEP-7 is done (as S11-25 did). The subject is one full imperative sentence that says what changed for the researcher, ending in a period ("Show each record's likenesses beside this Source's on the evidence."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/valuecompare/depiction_test.go` `TestCompareDepiction` *(new)*, table-driven; `contract_test.go` rows; `TestCompareUnsupported` updated | Same id → `same`/`exact`/1; two ids → `unknown`/`unjudged`/0; an empty side → `unknown`/`empty`; `Key` equal exactly for one id; `ComparerTypes()` includes depiction | AC-1.1, D-1, D-2 |
| T-2 | Go | `core/subjectprofile/lifespan_test.go` `TestLifespanFromEvents` (extended); `TestLifeOf` *(new)* | Each existing case's `LifeSpan`: AC-2.11 → `undated`; AC-2.12, AC-2.16, AC-2.17 → `contradicts`; AC-2.18 → no gap; a two-event case → `First`/`Last` the original dates (an "abt 1815" birth kept as written); one event → `First == Last`; `ForSubject` of a likeness candidate → `Image` set | AC-2.2, AC-2.3, AC-2.7, D-4, D-5 |
| T-3 | Go | `core/claimevidence/sections_test.go` `TestSections` (new rows) | AC-1.1's two lines with D-7's flags; AC-1.4's collapse; an `unknown` line with cause `empty` dropped; AC-3.1 and AC-3.2's orders | AC-1.1, AC-1.4, AC-3.1, AC-3.2, D-6, D-7 |
| T-4 | Go | `core/claimevidence/build_test.go` `TestBuildLifespan` *(new)* | Agreeing lifespans → result with both sides; one side `contradicts` → result, `unknown`; both `undated` → nil; an event kind → nil; a conflict → `conflict` | AC-2.1, AC-2.2, AC-2.3, AC-2.4, AC-2.5, AC-2.6, D-8 |
| T-5 | Go | `core/database/promotealign/evidence_test.go` `TestEvidenceLikenesses` *(new)* | `SaveLikeness` shade 40 on the Source person, on member 1 (shade 40) and member 2 (shade 200), plus a refused-image member (stop if the fixture cannot refuse one: drop that case and report) → AC-1.1's lines; AC-1.5 | AC-1.1, AC-1.4, AC-1.5, D-9 |
| T-6 | Go | same file `TestEvidenceLifespan` *(new)*, with `datedDeath` | AC-2.1 (birth and death on the Source; birth and census on the record), AC-2.2 (death 1809), AC-2.3, AC-2.5 | AC-2.1, AC-2.2, AC-2.3, AC-2.5, D-10 |
| T-7 | Go | `pins_test.go` `TestResolvePinsLikeness` *(new)*; `promotealign_test.go` `TestProposeIgnoresLikenesses` *(new)* | Default pins → the same-image pair only; a proposal with and without likenesses → equal rows | AC-1.2, AC-1.3, AC-1.6 |
| T-8 | Go | `api/ffi/handlers/promote_evidence_test.go` `TestGetPromoteClaimEvidenceLikenessAndLifespan` *(new)* | A likeness line's `incoming.depiction.image_rel_path` and pins' paths non-empty and different for two images; `lifespan` set with dates for AC-2.1; unset for an event row | AC-2.8, AC-4.2, D-11, D-12 |
| T-9 | Go | `api/ffi/swift_vocabulary_test.go` `TestSwiftVocabularyMatchesGo` | `unjudged` in value causes; a `LifespanGap` row (`""`, `undated`, `contradicts`) and in the unrecognized subtest | AC-4.1, D-13 |
| T-10 | Swift | `macos/ProvenenciaTests/ReconciledValueDisplayTests.swift` `aDepictionReadsAsLikeness` *(new)* | `.depiction(CatalogDepictionValue(regionJSON: "", imageRelPath: "objects/x.png"))` → "Likeness" | AC-4.2, D-13 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `valuecompare` contract tests (D-1), `TestLayerPackages` (D-5), `TestKnobInventoryMatchesRegistry` (D-6, D-7), S11-19's `TestEvidence*` (D-9), `autoreconcile` and `likenesses` tests (D-3).

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests, catalog and plan-doc checks); on a host without Xcode, report the Go and plan-doc steps and say Xcode was not run
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings: conclusions.value.likeness
- [ ] UI matches the brief inventory — none: no chrome
- [ ] Version bump as the header says — none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan, and the exact names S11-20 reads (Lifespan, LifeSpan, LifespanGap, the Swift types)
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ — none: not gated
- [ ] Docs: claim-evidence.md, comparison-layers.md, matching-model.md rows, add-match-rule skill
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

Data: S11-D2 §5.2's project in a dogfood catalog: the obituary's John Smith with likeness photo A; the funeral card's member saved from the same image file; the album's member with photo B.

1. Open Promote on the obituary; open John's sheet. The likeness lines show in today's columns, worded "Likeness": the funeral card's checked, the album's unchecked. Name, sex and relationships are as before. (AC-1.1, AC-4.2)
2. Close the sheet and press Done on a fresh copy without opening it: on the person's conclusion page, the new claim pins the funeral card's likeness and not the album's. (AC-1.3)
3. Proposal bands and contested marks are the same as before this PR for every row. (AC-1.6)
4. In a debugger or a test log, the RPC response for John carries `lifespan` with the obituary's 14 May 1817 – 2 Jun 1890 and the record's 14 May 1817 – 2 Apr 1871. (AC-2.8)

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| A likeness changes a proposal | Likeness has no weight; T-7 compares proposals with and without |
| The reconciler groups depictions differently | Keys are byte-identical (D-3); its existing tests run unchanged |
| A member with two images shows only one line | The speaking pair is the strongest (layer 2): a same-image pair wins; between two different images, the first in the member's value order speaks. Accepted: each member is judged by eye once |

**Stop and report if:**

- `in.layer.Metas` has no `likeness` meta at your base, so `Sections` would skip the property (the loader reads every property today through `propertiesByID`).
- Two `SaveLikeness` calls with one shade on two Sources give two image File ids.
- `propertycompare.Compare` returns no `Speaking` for an `Unknown` relation.
- `claimevidence` or `subjectcompare` needs an import of `core/subjectprofile`, `core/autoreconcile` or `core/database` to build.
- A proposal's band, score or contested flag changes in an existing `promotealign` test.
- `autoreconcile`'s depiction tests change value after D-3.

## 12. Open questions

None.

## 13. PR

**Title:** S11-32 — Lifespan and likeness evidence lines
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-11/prs/S11-32-lifespan-likeness-lines.md`.
