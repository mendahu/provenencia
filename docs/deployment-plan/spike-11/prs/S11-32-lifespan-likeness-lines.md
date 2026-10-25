# S11-32 — Lifespan and likeness evidence lines

**Status:** Landed in [#523](https://github.com/mendahu/provenencia/pull/523) (2026-10-22)

**Landed 2026-10-22.** Differences from the plan: AC-2.5's dates did not conflict (the row's span 1817–1890 overlaps a record window ending 1838), so `TestBuildLifespan` and `TestEvidenceLifespan` use a burial on 9 Jan 1810 to make the lives disjoint. D-20 named no facet for `subject_type_id`, but `TestFacetReleaseHonesty` counts `subject_types` as deletable (it is an Impact lookup kind), so `deleteimpact/facets.go` has a second `rowFacet` on `KindSubjectType` for it, and `docs/catalog-deletes.md` says so. `datedEvent` lost its `polarity` parameter (every caller passed ""), and calls `fixture.datedOn`. `Vocabulary.BindingOrder` has a value receiver, as `TermsByID` does. T-10 asserts `lifespan` unset through the handler; the lifespan's dates on the wire are `TestClaimEvidenceProtoCarriesTheLifespan` (a unit test on `claimEvidenceProto`), and the loader's dates are `TestEvidenceLifespan`. The STEP-6 commit passes an empty `claimevidence.Inputs{}` from the loader until STEP-7 fills it. The names S11-20 reads: `claimevidence.Evidence.Lifespan` (`claimevidence.Lifespan`), `subjectcompare.LifeSide`, `subjectcompare.LifespanMissing`, `claimevidence.Inputs`, `claimevidence.Rules.ByEye`; Swift `CatalogClaimEvidenceLifespan`, `CatalogClaimEvidenceLifeSide`, `LifespanMissing`, `ValueRelation.none`, `ValueCause.unjudged`, `CatalogConclusionValue.depiction`, `L10n.Conclusions.likenessValue`. No stop condition fired.
**Spike / slice:** Spike 11 · Slice 5 — Claim evidence
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no new chrome: today's sheet words a likeness value as "Likeness" in its existing columns and does not draw the lifespan; S11-20 draws both from S11-D2)
**Board:** —
**Depends on:** S11-25 (landed [#517](https://github.com/mendahu/provenencia/pull/517)), S10-54 (landed [#433](https://github.com/mendahu/provenencia/pull/433))
**Blocks:** S11-20
**Branch:** `s11-32-lifespan-likeness-lines`
**Version bump:** none — Spike 11 PRs leave `VERSION` at 0.2.0, and the product bump is cut when a release ships ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-match-rule`](../../../../.agents/skills/add-match-rule/SKILL.md), [`add-audit-scope`](../../../../.agents/skills/add-audit-scope/SKILL.md), [`add-catalog-delete`](../../../../.agents/skills/add-catalog-delete/SKILL.md), [`add-catalog-model`](../../../../.agents/skills/add-catalog-model/SKILL.md), [`add-depiction-value`](../../../../.agents/skills/add-depiction-value/SKILL.md), [`add-seeded-vocabulary`](../../../../.agents/skills/add-seeded-vocabulary/SKILL.md), [`add-ffi-handler`](../../../../.agents/skills/add-ffi-handler/SKILL.md), [`add-localized-string`](../../../../.agents/skills/add-localized-string/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md)
**Rules that bind:** `.cursor/rules/match-rules.mdc`, `.cursor/rules/depiction-values.mdc`, `.cursor/rules/go-tests.mdc`, `.cursor/rules/ffi-handlers.mdc`
**Context:** map scope entry [S11-32](../deployment-plan.md#s11-32--lifespan-and-likeness-evidence-lines) (the contract S11-20 builds on); [`claim-evidence.md`](../claim-evidence.md) (Lines); [`comparison-layers.md`](../comparison-layers.md#lifespan-a-derived-property-for-people) (lifespan), §Detail (causes); [S11-12 plan](S11-12-lifespan.md) D-12 (when a side has no lifespan); [S11-25 plan](S11-25-claim-evidence-rpc.md) D-2 – D-4 and its Landed note; [S11-D2](../design/archive/S11-D2-evidence-sheet.md) §5 (the example project these lines feed); working rules "The pin policy has one owner, in Go", "One wire vocabulary", "Layers and packages", "High tolerance for churn"
**Planned:** 2026-10-22 against `main` @ `d81808f`. **Refreshed:** 2026-10-22 after the ship loop stopped at STEP-4 ([ship log](S11-32-lifespan-likeness-lines.ship.md)): D-7 rewritten, D-19 – D-22 added, STEP-5 added (user, 2026-10-22: route binding writes through `writes.Run`, audit them, release them on a Property delete, and add a reorder RPC). STEP-1 – STEP-3 and STEP-4's seed and cache halves are already committed on the branch (`bd81258`, `391dd1f`, `eeff78e`, `948864c`)
**Review:** [`S11-32-lifespan-likeness-lines.review.md`](S11-32-lifespan-likeness-lines.review.md) — round 1: Fits with changes / Ready with changes; refresh (2026-10-22): round 2 Fits / Ready with changes, no open blocker; product calls answered by the user

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

**Resuming (refresh, 2026-10-22).** The branch already holds STEP-1 – STEP-3 and STEP-4's seed order and `Vocabulary.Bindings`. Check them against this plan, then start at STEP-4's remaining check (T-3's `TestVocabularyBindings`) and continue with STEP-5. Do not redo committed work.

The claim evidence S11-19 and S11-25 built leaves out two things the evidence sheet must show. A person's **lifespan** is compared in matching but never reaches the sheet, and a side whose own dates contradict each other has no lifespan with no reason given. **Likenesses** never reach the evidence at all: the loader drops the depiction, and value compare has no comparer for it. This PR carries both through the engine and the wire, and orders the sheet's sections by the catalog's own property order, so S11-20 can draw S11-D2 with no engine work.

**Why now.** S11-20 (the sheet) is planned against this PR's contract and waits on it (map, user 2026-10-22).

**What stays.** Matching: likeness has no weight, earns no points and is never a conflict; bands, candidates, contested and moves do not change. The lifespan's place in matching (S11-12, S11-13). Apply and the pin write (S11-25): likeness pins come from `PinOffers` like any line. Today's sheet layout (S11-25 D-14). The reconciler's grouping of likenesses: the same image File is the same value. The Person page (it shows likenesses as a gallery, not a field, so the person binding order change does not move it).

## 2. User stories

Conventions. Go fixtures: `promotealign_test.go`'s `newFixture`, `newSource`, `person`, `participation`, `promote`, `join`, `lifeFixture`; `evidence_test.go`'s `filedBirth`; `datedOn` *(new, D-14)*; `observationstest.SaveLikeness(t, c, userID, sourceID, subjectID, shade)`, which saves a crop whose PNG is decided by `shade` (the same shade on two Sources gives one image File: `ingest.StagedImage.InsertTx` finds it by checksum). "Image A" is shade 40, "image B" shade 200.

**US-1.** As a researcher, the evidence lists each record's likenesses beside this Source's, so I can judge them by eye.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | A Source person with image A; a handle with two members, one with image A, one with image B | One `likeness` section with two lines: (A, A) relation `same`, cause `exact`, outcome `agree`, checked, pinnable; (A, B) relation `unknown`, cause `unjudged`, outcome `unknown`, pinnable, not checked |
| AC-1.2 | AC-1.1, `PinOffers` | Two offers for the likeness lines: (A, A) checked, (A, B) not checked |
| AC-1.3 | AC-1.1, Done with the row never opened | The new claim pins the incoming likeness and the image-A member's likeness; the image-B member's is not pinned |
| AC-1.4 | Three members with image A | One collapsed line naming three pins |
| AC-1.5 | A member whose likeness image is refused (its File set to `image/jpeg`, so `Readable` empties it) | No line for that member |
| AC-1.6 | The same Source person and handle proposed with and without the likenesses | The same band, score and contested flag |

**US-2.** As a researcher, the evidence carries the lifespan check, and says when a side has none and why.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | A Source person, subject of a birth on 14 May 1817 and a death on 2 Jun 1890; a handle whose member is subject of a birth on 14 May 1817 and a census on 2 Apr 1871 | `Evidence.Lifespan`: outcome `agree`; row side first 14 May 1817, last 2 Jun 1890, no missing reason; record side first 14 May 1817, last 2 Apr 1871 |
| AC-2.2 | The Source person's death on 2 Jun 1809 (before the birth) | Row side: no dates, missing `contradicts`; outcome `unknown`, relation and cause empty |
| AC-2.3 | The handle with no dated events | Record side: missing `undated`; outcome `unknown` |
| AC-2.4 | Neither side has a dated event | `Evidence.Lifespan` is nil |
| AC-2.5 | The record's member buried 9 Jan 1838 after a baptism on 3 Mar 1801, against AC-2.1's Source person | Outcome `conflict` |
| AC-2.6 | An event or place row | `Evidence.Lifespan` is nil |
| AC-2.7 | A side with one dated event (married in 1843) | First and last are both that date |
| AC-2.8 | The RPC for AC-2.1 | `lifespan` with both sides' `first` and `last` as dates and `missing` empty; it adds no offer |

**US-3.** As the sheet, sections arrive in the catalog's property order for the kind.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | A new catalog; a person with `name`, `likeness`, `sex_at_birth` and `interests` sections | Order `name`, `likeness`, `sex_at_birth`, `interests` (the seeded binding order) |
| AC-3.2 | An event with `date` and `event_type` sections | Order `event_type`, `date` |
| AC-3.3 | A warm catalog session: an evidence open, then `ReorderSubjectTypeProperties` for person with `interests` before `name`, then a second evidence open | The second open lists `interests` first, with no restart |
| AC-3.5 | `ReorderSubjectTypeProperties` with a list that is not exactly the type's bound properties (one missing, one extra, or one repeated) | `subject_vocab.invalid`; no order changes |
| AC-3.6 | Assign, remove and reorder a binding | Each records one revision whose changes have entity type `subject_type_property` (a reorder: one change per row whose order moved); a reorder that moves nothing, and an assign of a bound pair, record no revision |
| AC-3.7 | Delete a user Property bound to the person type | One revision: the Property's delete change and the binding's release change (`subject_type_property`, action delete, `subject_type_id` and `property_id` in its fields) |
| AC-3.4 | A section whose property has no binding for the kind | After every bound section, by key |

**US-4.** As a contributor, the new codes cross the wire once.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-4.1 | `TestSwiftVocabularyMatchesGo` | `ValueCause` lists `unjudged`; `ValueRelation` lists `""`; `LifespanMissing` lists `""`, `undated`, `contradicts` |
| AC-4.2 | A likeness line's value on the wire, in Swift | `.depiction(CatalogDepictionValue)` with the image's relative path; today's sheet words it "Likeness" |

## 3. Scope

| In | Covered by |
| --- | --- |
| Depiction comparer and cause `unjudged` (layer 1) | STEP-1, D-1, D-2 |
| The reconciler keys depictions through `valuecompare.Key` | STEP-2, D-3 |
| `subjectprofile` carries the depiction and returns each side's life with its values | STEP-3, D-4, D-5 |
| Seeded person binding order (likeness second); bindings on the cached vocabulary, dropped on a binding edit | STEP-4, D-6, D-7 |
| Binding writes through `writes.Run`, audited as `subject_type_property`; the binding FKs `Audited`; a Property delete releases its bindings; `ReorderSubjectTypeProperties` RPC | STEP-5, D-19, D-20, D-21, D-22 |
| `claimevidence`: section order from the catalog, `Rules.ByEye`, the lifespan result, registry default and inventory row | STEP-6, D-8, D-9, D-10, D-11 |
| Loader: decode the depiction, keep each side's life, pass order and lives, resolve image paths | STEP-7, D-12, D-13, D-14, D-15 |
| Wire: `ConclusionValue.depiction`, `PromoteClaimEvidenceLifespan`, handler mapping | STEP-8, D-16, D-17 |
| Swift: vocabulary, store types and mapping, `ReconciledValueDisplay` | STEP-9, D-18 |
| Docs | STEP-10 |

| Out | Home |
| --- | --- |
| Drawing the lifespan and the crops, the new copy, section order on screen | S11-20 |
| Face comparison, or any likeness weight in matching | not planned |
| Sending a depiction's region on the evidence wire | not planned (D-16: the sheet draws the stored crop) |
| A lifespan on members (each member's own) | not planned (the line compares the row with the record's profile) |
| Lifespans for events or places | not planned (working rule "Lifespan is for people only") |
| Reordering existing catalogs' person bindings | not planned ("New seeds are for new catalogs", Spike 10 decisions); a researcher reorders through `ReorderSubjectTypeProperties` once a gesture calls it |
| A reorder gesture for a type's bindings, and the `GenealogyStore` method it calls | `docs/ideas/binding-order-editing.md` *(new)*, written in STEP-10: the Properties page is property-centric today (a type's bindings are toggles on a property, `PropertiesView.swift` L501–517), so an ordered per-type list is new chrome and needs a brief |
| A delete for subject types | not planned: `subjecttypes` has no delete, so `subject_type_properties.subject_type_id` gets no facet release (`TestFacetReleaseHonesty` allows an audited CASCADE whose parent is not deletable) |
| `source_type_metadata_fields` (source type ↔ metadata field suggestions) on the audited shape | not planned (user, 2026-10-22: they are UI hints, not research, and stay silent; review M-23) |
| `subjectvocab.EnsureBinding` through `writes.Run` | not planned: `Install` writes the seed at create time with no audit, as `properties.Upsert` does |
| Retiring `L10n.Conclusions.likenessValue` once crops draw | S11-20 (its Deletes) |
| Recording S11-32's names in S11-20's plan | S11-20 STEP-0 |
| Sharing a `Verdict` type between `Line` and `Lifespan` | not planned (review M-6, declined: two structs of four fields, filled by one helper, D-11) |

## 4. Current state

Verified at `d81808f`.

| Area | Where | Today |
| --- | --- | --- |
| Layer 1 | `core/valuecompare/valuecompare.go` `Value` (L151–162: `Type, Text, Term, Integer, Name *namevalue.Value, NamePattern, Date *datevalue.Value, Ref`), `comparers` (L171–177: text, term, integer, date, name), `Compare` (L181), `Key` (L201), cause block (L54–83, last `CauseUnsupported`); `term.go` (the 20-line comparer shape) | No depiction field and no comparer: a depiction is `Unknown` / `unsupported` |
| Layer 1 tests | `contract_test.go` `samples` (L32–89), `generators` (L206–233), `TestComparerContract` (L318, `allTypes` includes `ValueTypeDepiction`), `TestKeyContract` (L402); `valuecompare_test.go` `TestCompareUnsupported` (L13–39) | — |
| Layer 2 | `core/propertycompare/propertycompare.go` `Compare` (L127–168), `outcomeOf` (L97–107) | An `Unknown` relation gives `OutcomeUnknown` / `CauseRelation` **with** `Speaking` set (L162–163); a multiple property's conflict becomes `OutcomeUnknown` / `CauseCardinality` (L164–166) |
| Test fixtures | `core/subjectcompare/subjectcomparetest/subjectcomparetest.go` `Metas()` (no `likeness`, no `interests`); `core/propertycompare/propertycomparetest/propertycomparetest.go` (`Term`, `Text`, `Name`, `Date`, `Ref`; no depiction); `core/database/subjectvocab/subject_section_test.go` (pins `Metas()` against the seed) | — |
| Reconciler | `core/autoreconcile/modules.go` `moduleFor` (L48–74: depiction keyed by `idKey(v.Depiction.ImageFileID)`), `compareValue` (L78–80: text, integer, date only), `idKey` (L117); `autoreconcile.go` `Value.Depiction *depictionvalue.Value` (L87–89) | A private key rule beside value compare's |
| Profile | `core/subjectprofile/subjectprofile.go` `Entity` (L22), `LifeEvent` (L30), `Subject` (L38), `Build` (L74), `ForSubject` (L126), `addLifespans` (L141–149: silent on failure), `lifespan` (L158–215: not ok at L184 no dated event, L191 and L197 a year shift fails, L201–205 a date outside the window, L206 and L210 `EndsDate` refuses), `Value` (L220–236: shares `Name` and `Date`; not `Depiction`); `lifespan_test.go` `TestLifespanFromEvents` (AC-2.1 – AC-2.18, through `ForSubject`) | No reason kept; the span's first and last dates not exposed |
| Subject compare | `core/subjectcompare/subjectcompare.go` `Lifespan` (L45–50), `Rules` (L53), `Compare` (L103), `PropertyResult` (L91–95), `Derived` (L154); `lifespan.go` | `Compare` returns the lifespan lead's `PropertyResult` |
| Seed (landed on the branch in `948864c`; rows below describe `d81808f`) | `core/database/subjectvocab/registry.go` `seedBindings` (L185–189: person `name` 0, `sex_at_birth` 1, `gender` 2, `interests` 3, `likeness` 4); `subjectvocab_test.go` (L435 asserts gender 2 and interests 3; L490 asserts likeness 4); `docs/seeded-vocabulary.md` (L355–360 person list) | — |
| Vocabulary cache | `core/database/graphcache/vocab.go` `Vocabulary{Properties, Terms, Types, Bindings}` and `VocabBinding` (landed on the branch in `948864c`), `loadVocabulary`; `core/database/effects/registry.go` `"subject_type_properties": {None: true}` (L321); `core/database/graphcache/commit.go` drops the vocabulary on `effects.ListVocabulary` (L23) | A binding edit drops nothing |
| Binding writes | `core/database/subjectvocab/subjectvocab.go` `EnsureBinding` (L102, Install only, L285), `AppendBinding` (L119, `ON CONFLICT DO NOTHING`), `DeleteBinding` (L196, refuses `LockedBinding` on a provenencia Property with `ErrLocked`), all raw `db.Exec` on `*database.Catalog`; `ErrInvalid` (`subject_vocab.invalid`) L28; `api/ffi/handlers/subject_defs.go` `AssignSubjectTypeProperty` (L158), `RemoveSubjectTypeProperty` (L188): no `writes.Run`, so no audit and no effects | A binding edit records nothing and drops no cache |
| Binding FKs and deletes | `core/database/catalogmodel/model.go` L178–179: both `subject_type_properties` FKs CASCADE, `BucketFacet`, not `Audited`; `core/database/effects/registry_completeness_test.go` `TestCascadeIntoNonEmptyEffectIsAudited` (L44); `core/database/deleteimpact/pragma_test.go` `TestFacetReleaseHonesty` (L105); `core/database/deleteimpact/facets.go` `facetReleases` (`source_metadata_layout` rows L210–214), `rowFacet` (L446); `core/database/properties/properties.go` `Delete` (L380: `Impact`, `Refuse`, then `sqlDelete`, no `ReleaseFacets`); `core/database/subjecttypes` has no delete | A Property delete drops its bindings silently |
| Reorder precedent | `core/database/sourcemetadata/sourcemetadata.go` `Reorder` (L381–439: validates ids and duplicates, one change per moved row, entity id the field id); `api/ffi/handlers/sources.go` `ReorderSourceMetadata` (L561, `writes.Run`); `core/database/metadatafields/metadatafields.go` `Delete` (L357–370: `Impact`, `Refuse`, `ReleaseFacets`); `api/proto/engine.proto` `ReorderSourceMetadataRequest` (L656), `METHOD_ASSIGN_SUBJECT_TYPE_PROPERTY = 58`, last method `METHOD_GET_PROMOTE_CLAIM_EVIDENCE = 110`; `api/ffi/dispatch.go`; `api/ffi/handlers/subject_defs_test.go` `TestSubjectTypePropertiesAndRegistry` (L111), `TestCreateUpdateDeleteProperty` (L53) | — |
| Binding order reader | `core/database/conclusiondetails/conclusiondetails.go` `sqlPropertyIDs` (L124, `LEFT JOIN subject_type_properties`), `loadProperties` (L257–310: calls `g.Vocabulary()` at L258; sorts bound first by `sort_order`, then `PropertyKey`, L296–302) | A second order read beside the cache |
| Evidence | `core/claimevidence/claimevidence.go` `Rules` (L20–29), `Line` (L47–58), `Evidence` (L92–97); `sections.go` `Sections` (L61–100: skips no meta L66, no ref L71, `OutcomeUnknown` L84; sections sorted by key L96–98), `line` (L104–121); `build.go` `Build` (L24–42; L28 keeps only `subjectcompare.Compare(...).Evidence`; `bundles` calls `Sections` for neighbor sections) | Unknown lines dropped; ascending key order; no lifespan |
| Evidence tests | `sections_test.go` `TestSections` (helper `sections(kind, in, ms...)`), `build_test.go` `TestBuild` (helper `scene.build`), `invariants_test.go` `TestSectionsInvariants` | — |
| Registry | `core/matchregistry/evidence.go` `defaultEvidence` (L11–22), `evidence_test.go` `TestEvidenceSection`, `inventory_test.go` `TestKnobInventoryMatchesRegistry`; `docs/matching-model.md` L200–202 | — |
| Loader | `core/database/promotealign/profiles.go` `decode` (L54–77, no depiction), `reconciled` (L80, ignores `graphcache.Value.ImageFileID`), `layerValues` (L93–120, ignores `EvidenceObservation.DepictionImageFileID`), `lifeEvent` (L124), profiler `done map[string]subjectcompare.Values` shared with `index.go` `matchIndex.profiles` (L24) through `newProfiler` (L173–177) and `keep()`, read at L212, L250, L308, and by `export_test.go` (L50, L106), `build` (L296–319); `layer.go` `primarySubject{id, ref, kind}` (L19–23), `layerMemo{layer, primary}` (L36), `forSubject` with `life` (L101–127); `evidence.go` `evidence` (L56–183, `claimevidence.Build` at L176–178), `memberValues` (L275–334, `ForSubject(..., nil)` at L311); `values.go` `EvidenceValue{Type, Value, TermLabel}`, `NewEvidenceValues`, `Incoming`, `Member`, `value` (L85–100); `pins.go` `PinOffers` (L38–70) | Likeness dropped at decode; life not kept after building values |
| Loader fixtures | `promotealign_test.go` `datedEvent(f, src, art, eventType, year, polarity)` (L888), `lifeFixture` (L901), `datedBirth(src, art, y int)` (L957); `evidence_test.go` `filedBirth` (L176) | Year-only dates |
| Cache | `core/database/graphcache/evidence_graph.go` `EvidenceObservation.DepictionImageFileID` (L66, already through `Readable`); `graph.go` `Value.ImageFileID` (L94) | Both in hand |
| Image refusal precedent | `core/database/conclusiondetails/conclusiondetails_test.go` L324, `core/database/autoreconciler/autoreconciler_test.go` L863: `UPDATE files SET media_type = 'image/jpeg' WHERE id = ?` on `observationstest.ImageOf` | — |
| Wire | `api/proto/engine.proto` `ConclusionValue` (L1404–1412, next field 6), `DepictionValue{region_json = 1, image_rel_path = 2}` (L1962), `GetPromoteClaimEvidenceResponse` (L1597–1602, next field 5); `api/ffi/handlers/promote_evidence.go` `GetPromoteClaimEvidence` (L21–62), `claimEvidenceProto(e, values)` (L64), `evidenceSectionsProto` (L104), `evidenceValueProto` (L140–155); `conclusion_detail.go` `conclusionValueProto` (L140–158); `conclusion_detail_test.go` `TestConclusionValueProtoCoversEveryKind` (L113), `likenessFixture` (L235); `core/database/files/files.go` `RelPaths` (L92) | A depiction arrives as an empty `ConclusionValue` |
| Wire tests | `api/ffi/handlers/promote_evidence_test.go` `TestGetPromoteClaimEvidence`, `newEvidenceFixture`; `api/ffi/swift_vocabulary_test.go` `TestSwiftVocabularyMatchesGo` (value relations and value causes rows, L284–291) | — |
| Swift | `macos/App/Platform/GenealogyStore.swift` `CatalogConclusionValue` (L638–645), `CatalogDepictionValue` (L487), `CatalogClaimEvidence` (L406–418); `GoStore.swift` `getPromoteClaimEvidence` (L1034–1074), `mapConclusionValue` (L1429–1438, exhaustive), the inline `DepictionValue` mapping (L2333); `ConclusionVocabulary.swift` `ValueRelation` (no `""` case), `ValueCause` (L164–218, `.none` for `""`); `macos/App/Features/Conclusions/ReconciledValueDisplay.swift` `string(for:locale:orders:)` (L78–97, the one exhaustive switch over `CatalogConclusionValue`) | — |

**Precedent.** `source_metadata_layout` is the audited join-table shape this PR copies for bindings: its Audited FKs, its effects entry, its `rowFacet`, `sourcemetadata.Reorder` and `ReorderSourceMetadata`. S11-25 ([#517](https://github.com/mendahu/provenencia/pull/517)) carried evidence over the wire: copy its proto, handler, `GoStore` mapping and vocabulary-row shape. S11-12 ([#480](https://github.com/mendahu/provenencia/pull/480)) built the lifespan in `subjectprofile`: extend `lifespan` in place. S10-54 ([#433](https://github.com/mendahu/provenencia/pull/433)) sent `DepictionValue` with paths through `files.RelPaths` (`observations.go` `depictionPaths`: collect ids, one query).

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | **A depiction comparer.** `valuecompare.Value.Depiction *depictionvalue.Value` *(new)*, as `Name` and `Date` hold their pure types. `core/valuecompare/depiction.go` *(new)* `type depiction struct{}`: a nil `Depiction` or an empty `ImageFileID` on either side → `Unknown` / `CauseEmpty`; equal `ImageFileID` bytes → `Same`, similarity 1, `CauseExact`; otherwise `Unknown`, similarity 0, `CauseUnjudged` *(new)* `"unjudged"` ("two values the layer cannot relate: the researcher judges them"). `key`: `string(ImageFileID)`, ok when non-empty, at `Same` and `Equivalent`. Registered under `propertymeta.ValueTypeDepiction` | The same File is the same pixels; two crops can show one face and the app cannot tell. `Unknown` keeps matching untouched. One pointer per structured type (review M-2) | A bare `Image []byte` — a second shape and a synonym of `ImageFileID`; `Near` or `Different` — judgements the layer cannot make |
| D-2 | **The relation set does not change.** `CauseUnjudged` is a cause, not a relation; `outcomeOf` and `contradicts` are untouched | `TestEveryRelationHasAnOutcome` unchanged | A new relation |
| D-3 | **The reconciler keys depictions through layer 1.** `Value.compareValue` sets `Depiction: v.Depiction`; `moduleFor`'s depiction case joins text and integer's `valuecompare.Key(..., Same, ...)` case and its closure goes. Keys are byte-identical to `idKey`'s, so grouping is unchanged and `autoreconciler.CacheVersion` is not bumped | One key rule per value type; "Build, then migrate" | Two rules for one type |
| D-4 | **Profiles carry the depiction.** `subjectprofile.Value` sets `Depiction: v.Depiction`, as it shares `Name` and `Date` | One conversion path | A copy step per field |
| D-5 | **One lifespan pass returns the life.** In `core/subjectcompare/subjectcompare.go`, beside `Lifespan`: `type LifespanMissing string` with `LifespanUndated = "undated"`, `LifespanContradicts = "contradicts"`, and `type LifeSide struct{ First, Last *datevalue.Value; Missing LifespanMissing }` *(new)*. `subjectprofile.lifespan` returns `(window, span propertycompare.Value, side subjectcompare.LifeSide)`; it has a lifespan exactly when `side.Missing == ""`. No dated event (L184) or a year shift that fails (L191, L197: no window can be built from those dates) → `undated`; a date outside the window (L201–205) or an `EndsDate` refusal (L206, L210) → `contradicts`. It keeps a `[]datevalue.Value` parallel to `early` and `late`: `First` is the first date whose early end equals the span's first end, `Last` the first whose late end equals its last (`DateEnd` compares with `==`). `addLifespans` returns the `LifeSide` of the first lead in key order (the zero value with no lead). `Build` and `ForSubject` return `(subjectcompare.Values, subjectcompare.LifeSide)` — `Build` keeps its `error` as a third result. Callers that need only values ignore the side: `profiles.go` L291 (`keyValues`), `evidence.go` L311 (`memberValues`) | One computation per side: the window and the reason cannot drift (review M-5); `claimevidence` may not import `subjectprofile` (`TestLayerPackages`); original dates keep "abt 1815" readable (S11-20 D-13). Names match the wire (`LifeSide`, `missing`) and avoid `gap`, a value cause (review M-4) | A second `LifeOf` pass; ends instead of dates |
| D-6 | **Likeness is seeded second.** `seedBindings` person: `name` 0, `likeness` 1, `sex_at_birth` 2, `gender` 3, `interests` 4. New catalogs only (Install); no migration | User, 2026-10-22: the sheet uses the catalog's order, with Likeness after Name as S11-D2 draws | A sheet-only order |
| D-7 | **The cache knows binding order.** `graphcache.Vocabulary.Bindings []VocabBinding` *(new)*, `VocabBinding{TypeID, PropertyID []byte; SortOrder int}`, loaded in `loadVocabulary` from `subject_type_properties`. `(v *Vocabulary) BindingOrder() map[vocabkey.SubjectType][]vocabkey.PropertyKey` *(new)* resolves ids to keys once: per type key, its bound property keys by `SortOrder`, then key. `vocabkey.CompareByOrder(order []PropertyKey, a, b PropertyKey) int` *(new)* is the one ranking: a key in `order` before one that is not, bound keys by index, the rest by key. `conclusiondetails.loadProperties` sorts its fields with it from `BindingOrder()` (the handle's type key read from `v.Types` by `subjectTypeID`), and `sqlPropertyIDs` drops its `subject_type_properties` join and `sort_order` column (M-20). The drop on a binding edit is D-19 – D-21: binding writes commit through `writes.Run`, and the `subject_type_properties` effect lists `ListVocabulary` | A warm evidence open issues no catalog query (`TestEvidenceWarmIssuesNoQuery`); a reorder shows on the next open (AC-3.3) | A per-request SQL read; the conclusion page's own order read beside the cache; caching without a drop; binding writers that drop the cache by hand (a second invalidation path beside effects) |
| D-8 | **Section order comes from the catalog.** `claimevidence.Inputs{Order map[vocabkey.SubjectType][]vocabkey.PropertyKey; Row, Record subjectcompare.LifeSide}` *(new)*: what the loader read beside the values. `Sections` gains `order []vocabkey.PropertyKey` (the kind's) and sorts with `vocabkey.CompareByOrder` (D-7). `Build` gains a last parameter `in Inputs` and passes `in.Order[kind]` to its own sections and `in.Order[n.Kind]` to each neighbor's | One owner of property order, shared with the conclusion page (review M-1, user 2026-10-22); `claimevidence` stays pure | `Rules.Order` — a second per-kind order beside the catalog's |
| D-9 | **Lines the researcher judges.** `claimevidence.Rules.ByEye []valuecompare.Cause` *(new)*: an `OutcomeUnknown` line whose speaking cause is listed is kept, pinnable and never checked. `Sections` L84 keeps such lines; `line` sets `Pinnable` for them. Default `CauseUnjudged`; inventory row `evidence.ByEye`. A line with no speaking pair or no member ref is still dropped | Policy in the evidence section, not a likeness special case; user, 2026-10-22 | `if key == "likeness"` |
| D-10 | **The lifespan result.** `claimevidence.Lifespan{Outcome propertycompare.Outcome; PropertyCause propertycompare.Cause; Relation valuecompare.Relation; Cause valuecompare.Cause; Row, Record subjectcompare.LifeSide}` *(new)*; `Evidence.Lifespan *Lifespan` *(new)*. `Build` keeps the whole `subjectcompare.Compare` result; when `subject.Lifespans[kind]` has a lead, it takes that lead's `PropertyResult` (the first lead in key order). Nil for a kind with no lead, and when neither side has `First` and neither side is `contradicts` | Reuses matching's comparison | A second compare in the loader |
| D-11 | **One verdict helper.** `verdictOf(res propertycompare.Result) (outcome, propertyCause, relation, cause)` *(new, private in `sections.go`)* reads the speaking pair (relation and cause empty when `Speaking` is nil); `line` and D-10 both call it | One reading of a result (review M-6) | A `Verdict` struct embedded in both — churn in every `Line` literal |
| D-12 | **The loader decodes depictions.** `decode` gains a last parameter `image []byte`; a non-empty image sets `Value.Depiction = &depictionvalue.Value{ImageFileID: <a copy>}`. `layerValues` passes `o.DepictionImageFileID`, `reconciled` passes `row.ImageFileID` | Both are already `Readable`; one decoder | A second read of `depiction_values` |
| D-13 | **The loader keeps each side's life.** `primarySubject` gains `life subjectcompare.LifeSide`, from `forSubject`'s second result in `layer.go`, memoized with the layer. The profile map holds `profile{values subjectcompare.Values; life subjectcompare.LifeSide}` *(new)*: `matchIndex.profiles` (`index.go`) and `profiler.done` change type together; `values(handle)` keeps its signature and `life(handle)` *(new)* reads the side; `lifeEvent` and the reads at L212, L250, L308 and `export_test.go` read `.values` | Memoized with what they describe (S11-19 D-21, D-22); no second map beside the index (review X-2) | A side map |
| D-14 | **One dated-event fixture.** `promotealign_test.go` `datedOn(src, art, eventType string, y, m, d int) subjects.Subject` *(new; `m` and `d` 0 for a year)*; `datedEvent` and `datedBirth` call it. Record-side events are filed with `lifeFixture`'s shape | Day-precision ACs with no near-copy helper (review X-4) | `datedDeath` beside `datedBirth` |
| D-15 | **Image paths with the other display values.** `EvidenceValue.ImageRelPath string` *(new)*. `NewEvidenceValues` stays; `(v *EvidenceValues) ResolvePaths(c *database.Catalog, e claimevidence.Evidence) error` *(new)* collects every `Depiction.ImageFileID` on lines, pins and relationship sections' lines and pins, calls `files.RelPaths` once, and `Incoming` / `Member` fill `ImageRelPath` from it | `EvidenceValues` already decorates values (term labels); one walk, in `promotealign` (review M-3) | Paths in the handler; a field `conclusiondetails` never sets |
| D-16 | **The wire value.** `ConclusionValue` gains `DepictionValue depiction = 6`, filled with `image_rel_path` only (region empty). `conclusionValueProto` gains a second parameter `depiction *engine.DepictionValue`: when not nil it is the result, before any other case; every existing caller passes nil. `evidenceValueProto` passes `&engine.DepictionValue{ImageRelPath: v.ImageRelPath}` for `propertymeta.ValueTypeDepiction`. `GetPromoteClaimEvidence` calls `values.ResolvePaths(c, e)` inside `withProjectCatalog` before `claimEvidenceProto` | The one encoder takes a depiction input (review M-3) | `conclusiondetails.Value.ImageRelPath` |
| D-17 | **The lifespan on the wire.** `PromoteClaimEvidenceLifeSide{DateValueInput first = 1; DateValueInput last = 2; string missing = 3;}`, `PromoteClaimEvidenceLifespan{string outcome = 1; string property_cause = 2; string relation = 3; string cause = 4; PromoteClaimEvidenceLifeSide row = 5; PromoteClaimEvidenceLifeSide record = 6;}` *(new)*; `GetPromoteClaimEvidenceResponse.lifespan = 5`, unset when nil. Dates through `valuecodec.DateToProto`. No offer | Not a section line: no member, no pin | — |
| D-18 | **Swift.** `ConclusionVocabulary.swift`: `ValueCause.unjudged`; `ValueRelation.none` matched by `case "": self = .none` (listed first, as `ValueCause.none`); `enum LifespanMissing: Sendable, Equatable { case none, undated, contradicts, unrecognized(String) }` *(new)* with `init(wire:)`. `GenealogyStore.swift`: `CatalogConclusionValue.depiction(CatalogDepictionValue)`; `CatalogClaimEvidenceLifeSide{first: CatalogDateValueInput?, last: CatalogDateValueInput?, missing: LifespanMissing}`, `CatalogClaimEvidenceLifespan{outcome: EvidenceOutcome, propertyCause: EvidencePropertyCause, relation: ValueRelation, cause: ValueCause, row, record: CatalogClaimEvidenceLifeSide}` *(new)*, both `Sendable, Equatable`; `CatalogClaimEvidence.lifespan: CatalogClaimEvidenceLifespan?`. `GoStore`: `mapDepictionValue(_:) -> CatalogDepictionValue` *(new)*, extracted from the inline mapping at L2333 and used there and in `mapConclusionValue`. `ReconciledValueDisplay.string(for:)` words `.depiction` as `L10n.string(L10n.Conclusions.likenessValue)` *(new)* "Likeness" | One wire vocabulary; an empty relation is named (review X-1); today's sheet stays readable | `relation: ValueRelation?` — every reader unwraps |
| D-19 | **Binding writes are audited.** In `subjectvocab`: `AppendBinding(tx *database.Tx, userID, typeID, propertyID []byte) ([]rowchange.Change, error)` and `DeleteBinding(tx *database.Tx, userID, typeID, propertyID []byte) ([]rowchange.Change, error)` replace the `*database.Catalog` versions (same checks, read through `subjecttypes.GetByIDTx(tx.Tx, …)` and `properties.GetByIDTx(tx.Tx, …)`; `database.RequireUserID`). Each returns one change: `EntityType: "subject_type_property"`, `EntityID: propertyID`, fields from a private `bindingFields(typeID, propertyID []byte, sortOrder int) map[string]any` *(new)* (`subject_type_id`, `property_id` as `database.UUIDJSON`, `sort_order`), wrapped in `rowchange.FullRow` on append and `rowchange.DeletedRow` on remove, as `evidencegraphs` `linkFields` is (M-24). Append runs `sqlAppendBinding`; `RowsAffected` 0 (already bound) returns no change and records no revision; otherwise it reads the new row's `sort_order` with a `SELECT` before building the change. Remove reads the row's `sort_order` first; an unbound pair returns no change and no error, as today. Callers move to the tx versions: `subject_defs.go` (L173, L203) and the tests `subjectvocab_test.go` (L152, L189), `core/database/subjects/delete_test.go` (L197), `core/database/deleteimpact/impact_test.go` (L411, L469, L518), each through `writes.Call(c, writes.Op{Action: "assign_subject_type_property", UserID: userID}, func(tx *database.Tx) (struct{}, []rowchange.Change, error) {...})` (or `remove_…`), with the fixture's user id. `properties.Delete`'s doc comment ("Type bindings CASCADE") says they are released and audited. `EnsureBinding` stays as it is, for `Install` | Every catalog edit is audited and drives effects; entity id is the property id, as `source_metadata_layout` uses the field id | Composite ids; auditing under the type id |
| D-20 | **The binding effect and FKs.** `effects` registry: `"subject_type_properties": {Entity: "subject_type_property", Lists: []List{ListVocabulary}}`, with no `SourceScope` and no `Handles` (vocabulary layer, per `add-audit-scope`). `catalogmodel` L178–179: both FKs gain `Audited: true`. `core/database/deleteimpact/facets.go` gains `rowFacet(catalogmodel.KindProperty, "subject_type_properties", "property_id", "subject_type_property", "property_id", col{"subject_type_id", colUUID}, col{"property_id", colUUID}, col{"sort_order", colInt})`; no facet for `subject_type_id` (no subject type delete) | `TestCascadeIntoNonEmptyEffectIsAudited` and `TestFacetReleaseHonesty` require exactly this | A `None` effect with a hand drop |
| D-21 | **A Property delete releases its bindings.** `properties.Delete` calls `deleteimpact.ReleaseFacets(tx, catalogmodel.KindProperty, id)` after `Refuse` and before `sqlDelete`, and appends the released changes before its own, as `metadatafields.Delete` does | The cascade backstop stays; the released rows are audited (AC-3.7) | Leaving the cascade silent (the effects test refuses it) |
| D-22 | **The reorder RPC.** `METHOD_REORDER_SUBJECT_TYPE_PROPERTIES = 111` *(new)*; `ReorderSubjectTypePropertiesRequest{string project_dir = 1; string user_id = 2; string subject_type_id = 3; repeated string property_ids = 4;}`, `ReorderSubjectTypePropertiesResponse{repeated SubjectTypeProperty properties = 1; WriteEffects effects = 2;}` *(new)*, as `ReorderSourceMetadataResponse`. `subjectvocab.ReorderBindings(tx *database.Tx, userID, typeID []byte, propertyIDs [][]byte) ([]rowchange.Change, error)` *(new)*: the type must exist (`subjecttypes.GetByIDTx`, else `ErrInvalid`); `propertyIDs` must be non-empty and exactly the type's bound property ids, each once (else `ErrInvalid`, nothing written; a type with no bindings always gets `ErrInvalid`); reads the type's bound ids and orders on the tx with a private `sqlBindingOrders` (`SELECT property_id, sort_order FROM subject_type_properties WHERE subject_type_id = ?`), then sets `sort_order` to each id's index; one update change (`sort_order` old and new) per row whose order moved, in the shape of `sourcemetadata.Reorder`. Locked bindings reorder like any other. Handler `ReorderSubjectTypeProperties` in `subject_defs.go` runs it in `writes.Run` (action `reorder_subject_type_properties`) and returns `ListBindings` after the write with `writeEffectsProto(wrote)`. `subjectTypePropertyProto(b subjectvocab.Binding) *engine.SubjectTypeProperty` *(new)* in `subject_defs.go` builds each row; `ListSubjectTypeProperties` (L141) and the registry handler (L295) move onto it (M-22). `AssignSubjectTypeProperty` and `RemoveSubjectTypeProperty` run D-19's writers in `writes.Run` (actions `assign_subject_type_property`, `remove_subject_type_property`); their wire is unchanged. No `GenealogyStore` method (Out) | A full permutation cannot leave two rows on one order; one shape with the Source metadata reorder | A move-one-row RPC; partial lists |

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `valuecompare` `depiction` comparer, `Value.Depiction`, `CauseUnjudged` *(new)* | comparer, field, code | `term.go`; `autoreconcile` `idKey` | Layer 1 owns comparers; the reconciler's rule moves onto it (D-3) |
| `subjectcompare.LifespanMissing`, `LifeSide` *(new)* | types | `subjectcompare.Lifespan` | Beside the rule they report on; reachable from `claimevidence` |
| `claimevidence.Inputs`, `Rules.ByEye`, `Lifespan`, `Evidence.Lifespan`, `verdictOf` *(new)* | types, field, func | `Rules.Checked`, `line` | The evidence section's policy and inputs |
| `graphcache.VocabBinding`, `Vocabulary.Bindings`, `Vocabulary.BindingOrder`, `vocabkey.CompareByOrder` *(new)* | type, field, funcs | `conclusiondetails` `sqlPropertyIDs` join and `loadProperties` sort | Cached with the vocabulary; the conclusion page moves onto them, so binding order has one reader and one ranking |
| Loader `primarySubject.life`, `profile`, `life(handle)`, `EvidenceValue.ImageRelPath`, `ResolvePaths` *(new)* | fields, type, funcs | `layerMemo`, `matchIndex.profiles`, `EvidenceValues` | Beside the values they belong to |
| Proto `ConclusionValue.depiction`, `PromoteClaimEvidenceLifespan`, `PromoteClaimEvidenceLifeSide` *(new)* | FFI | `DepictionValue`, `PromoteClaimEvidenceLine` | Reuse `DepictionValue` |
| Swift `LifespanMissing`, `ValueRelation.none`, `CatalogClaimEvidenceLifespan`, `CatalogClaimEvidenceLifeSide`, `CatalogConclusionValue.depiction`, `mapDepictionValue`, `L10n.Conclusions.likenessValue` *(new)* | enum, case, types, func, key | `ValueCause`, `CatalogClaimEvidenceLine` | Same pattern |
| `subjectvocab.ReorderBindings`, tx-based `AppendBinding` / `DeleteBinding`, entity `subject_type_property` *(new)* | funcs, audit entity | `sourcemetadata.Reorder`, `properties.Create` / `Delete` | The binding table's own package; the shapes are copied |
| `deleteimpact` `subject_type_properties.property_id` facet *(new)* | facet release | the `source_metadata_layout` facets | One more row in the registry |
| RPC `ReorderSubjectTypeProperties` *(new)* | FFI | `ReorderSourceMetadata`, `AssignSubjectTypeProperty` | Same shape; returns the list as `ListSubjectTypeProperties` does |
| Test helpers `datedOn`, `propertycomparetest.Depiction(id byte)` *(new)* | helpers | `datedBirth`, `propertycomparetest.Ref` | One dated-event helper; one depiction value builder |

**Builds on.** `graphcache`'s `Readable` depiction fields; `files.RelPaths`; `valuecodec.DateToProto`; S11-25's `PinOffers`, `EvidenceValues`, `evidenceValueProto`, vocabulary rows; S11-12's `lifespan`; the seeded binding order.
**Deletes.** `autoreconcile`'s depiction key closure (D-3). The `subject_type_properties` join in `conclusiondetails.sqlPropertyIDs` and `loadProperties`' own sort (D-7). The `*database.Catalog` `AppendBinding` and `DeleteBinding` (D-19). The silent cascade of bindings on a Property delete (D-21). The "depiction unsupported" assertion in `TestCompareUnsupported`. The ascending-key-only section sort.
**Next adds.** A binding-order gesture calls `ReorderSubjectTypeProperties` through one new store method. S11-20 draws `lifespan` and `.depiction` with no engine work and drops `likenessValue`. A depiction property on another kind is one seeded property and binding; a new judged cause is one `ByEye` entry; a user property gets its place on the sheet from its binding.

### Schema

None — no catalog change. The seed's binding order changes for new catalogs (D-6).

### Go core

| Package | Change |
| --- | --- |
| `core/valuecompare` | D-1, D-2 |
| `core/autoreconcile` | D-3 |
| `core/subjectcompare` | D-5's types |
| `core/subjectprofile` | D-4, D-5 |
| `core/database/subjectvocab` | D-6 |
| `core/database/graphcache`, `core/vocabkey`, `core/database/conclusiondetails` | D-7 |
| `core/database/subjectvocab`, `core/database/effects`, `core/database/catalogmodel`, `core/database/deleteimpact`, `core/database/properties` | D-19 – D-22 |
| `core/claimevidence` | D-8 – D-11 |
| `core/matchregistry` | `defaultEvidence`: `ByEye: {valuecompare.CauseUnjudged}` with a `// decided:` comment |
| `core/database/promotealign` | D-12 – D-15 |

```go
// package subjectcompare
type LifespanMissing string
const (
	LifespanUndated     LifespanMissing = "undated"     // no dated event to build one from
	LifespanContradicts LifespanMissing = "contradicts" // its own dated events cannot all be true
)
type LifeSide struct {
	First, Last *datevalue.Value // the earliest and latest dated events; nil when missing
	Missing     LifespanMissing
}

// package claimevidence
type Inputs struct {
	Order       map[vocabkey.SubjectType][]vocabkey.PropertyKey // each kind's bound properties, in binding order
	Row, Record subjectcompare.LifeSide
}
type Lifespan struct {
	Outcome       propertycompare.Outcome
	PropertyCause propertycompare.Cause
	Relation      valuecompare.Relation
	Cause         valuecompare.Cause
	Row, Record   subjectcompare.LifeSide
}
```

The loader sets `Inputs.Order` to `Vocabulary.BindingOrder()` (D-7).

Errors: none new.

### FFI

`api/proto/engine.proto`, regenerated with `./scripts/generate-proto.sh` (needs `protoc` and `protoc-gen-swift`: a macOS host with the toolchain, as for S11-25):

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
  string missing = 3; // subjectcompare.LifespanMissing
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

Proto for D-22 (STEP-5):

```proto
// Method: METHOD_REORDER_SUBJECT_TYPE_PROPERTIES = 111;

// ReorderSubjectTypePropertiesRequest sets a type's binding order. property_ids
// is every bound property of the type, each once, in the new order.
message ReorderSubjectTypePropertiesRequest {
  string project_dir = 1;
  string user_id = 2;
  string subject_type_id = 3;
  repeated string property_ids = 4;
}

message ReorderSubjectTypePropertiesResponse {
  repeated SubjectTypeProperty properties = 1;
  WriteEffects effects = 2;
}
```

The enum value goes last in `Method`, after `METHOD_GET_PROMOTE_CLAIM_EVIDENCE = 110`; the messages go after `RemoveSubjectTypePropertyResponse`. `api/ffi/dispatch.go` gains `MethodReorderSubjectTypeProperties` and its `case`, as `MethodReorderSourceMetadata`.

### macOS

None — no chrome. D-18's types and display only.

### Strings

`L10n.Conclusions.likenessValue` *(new)* = "Likeness" (`conclusions.value.likeness`, comment "A likeness value worded as text where no crop is drawn").

## 7. Steps

**STEP-1 — Depiction comparer**
Files: `core/valuecompare/valuecompare.go`, `depiction.go` *(new)*, `contract_test.go`, `valuecompare_test.go`, `depiction_test.go` *(new)*, `core/propertycompare/propertycomparetest/propertycomparetest.go` (`Depiction(id byte)`: a depiction value whose 16-byte image id is `id` repeated), `docs/deployment-plan/spike-11/comparison-layers.md` (§Detail cause list). Skill: `add-match-rule` ("Add a comparer", "Add a relation, cause or unit code").
Do: D-1, D-2. Contract sample and generator for depictions (equal and unequal ids).
Check: `go test ./core/valuecompare/... ./core/propertycompare/... ./core/matchregistry/...` passes; T-1.

**STEP-2 — The reconciler keys depictions through layer 1**
Files: `core/autoreconcile/modules.go`.
Do: D-3.
Check: `go test ./core/autoreconcile/...` passes unchanged; `CGO_ENABLED=1 go test -tags fts5 ./core/database/likenesses/ ./core/database/autoreconciler/...` passes.

**STEP-3 — Profiles: depiction and life**
Files: `core/subjectcompare/subjectcompare.go`, `core/subjectprofile/subjectprofile.go`, `lifespan_test.go`, `subjectprofile_test.go`, and every caller of `Build` / `ForSubject` (`core/database/promotealign/profiles.go`, `layer.go`, `evidence.go`, and tests: `rg -n 'subjectprofile\.(Build|ForSubject)\(' core`). Skill: `add-match-rule`.
Do: D-4, D-5. Callers take the new result or `_`; `layer.go` keeps the side for STEP-7. `TestLifespanFromEvents` rows gain a `side subjectcompare.LifeSide` read from `ForSubject`'s second result.
Check: `go test ./core/subjectcompare/... ./core/subjectprofile/... ./core/matchregistry/...` passes; `CGO_ENABLED=1 go test -tags fts5 ./core/database/promotealign/...` passes unchanged; T-2.

**STEP-4 — Binding order: seed and cache**
Files: `core/database/subjectvocab/registry.go`, `subjectvocab_test.go`, `core/database/graphcache/vocab.go`, `graphcache` tests, `docs/seeded-vocabulary.md` (person list order). Skill: `add-seeded-vocabulary`.
Also: `core/vocabkey/vocabkey.go`, `vocabkey_test.go`, `core/database/conclusiondetails/conclusiondetails.go`, `core/database/graphcache/vocab_test.go` *(new)*, package `graphcache_test`.
Do: D-6 and `Vocabulary.Bindings` are committed (`948864c`). Add `BindingOrder` and `CompareByOrder` and move `conclusiondetails` onto them (D-7). Add `TestVocabularyBindings` and `TestCompareByOrder`. The effects entry is STEP-5's.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/vocabkey/... ./core/database/subjectvocab/... ./core/database/graphcache/... ./core/database/conclusiondetails/...` passes, conclusion tests unchanged; T-3's seed and cache rows.

**STEP-5 — Audited binding writes and the reorder RPC**
Files: `core/database/subjectvocab/subjectvocab.go`, `subjectvocab_test.go`, `core/database/effects/registry.go`, `core/database/catalogmodel/model.go`, `core/database/deleteimpact/facets.go`, `core/database/properties/properties.go`, `properties_delete_test.go`, `api/proto/engine.proto`, regenerated `api/proto/engine/engine.pb.go` and `macos/App/Platform/Generated/engine.pb.swift`, `api/ffi/dispatch.go`, `api/ffi/handlers/subject_defs.go`, `subject_defs_test.go`, `core/database/graphcache/vocab_test.go` *(new)* in STEP-4, `core/database/subjects/delete_test.go`, `core/database/deleteimpact/impact_test.go`. Skills: `add-audit-scope`, `add-catalog-delete`, `add-catalog-model`, `add-ffi-handler`.
Do: D-19 – D-22, D-7's drop. Copy `sourcemetadata.Reorder` for `ReorderBindings` and `ReorderSourceMetadata` for the handler. At the end `rg -n '(AppendBinding|DeleteBinding)\(c,' core api` prints nothing. Update the `ListVocabulary` comment in `core/database/effects/registry.go` (L35–36) to say a binding edit also drives it.
Check: `./scripts/generate-proto.sh` then a second run leaves no diff; `CGO_ENABLED=1 go test -tags fts5 ./core/database/... ./api/...` passes (`TestCascadeIntoNonEmptyEffectIsAudited`, `TestFacetReleaseHonesty`, `TestResolversCoverEveryEntityType` included); `xcodebuild build` still succeeds (generated Swift only); T-3's drop row, T-13, T-14.

**STEP-6 — The evidence section**
Files: `core/claimevidence/claimevidence.go`, `sections.go`, `build.go`, `sections_test.go`, `build_test.go`, `core/subjectcompare/subjectcomparetest/subjectcomparetest.go` (`Metas()` gains `likeness` depiction / multiple and `interests` text / multiple, as seeded), `core/matchregistry/evidence.go`, `evidence_test.go` (`TestEvidenceSection` asserts `ByEye == {CauseUnjudged}`), `docs/matching-model.md` (row `evidence.ByEye`). Skill: `add-match-rule`.
Do: D-8 – D-11. `sections(...)` and `scene.build` pass an `Inputs` whose `Order` is the seed's (person `name`, `likeness`, `sex_at_birth`, `gender`, `interests`; event `event_type`, `event_name`, `date`).
Check: `go test ./core/claimevidence/... ./core/subjectcompare/... ./core/matchregistry/...` passes; `CGO_ENABLED=1 go test -tags fts5 ./core/database/subjectvocab/` passes (the fixture metas match the seed); T-4, T-5.

**STEP-7 — The loader**
Files: `core/database/promotealign/profiles.go`, `index.go`, `export_test.go`, `layer.go`, `evidence.go`, `values.go`, `promotealign_test.go` (`datedOn`), `evidence_test.go`, `pins_test.go`, `values_test.go`. Skill: `use-catalog-session`, `add-depiction-value`.
Do: D-12 – D-15; `evidence` builds `Inputs` (D-8) and passes it to `Build`.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/promotealign/... ./core/database/promote/...` passes, S11-19 and S11-25 tests unchanged except new fields; `TestEvidenceWarmIssuesNoQuery` passes; T-6, T-7, T-8.

**STEP-8 — The wire**
Files: `api/proto/engine.proto`, regenerated `api/proto/engine/engine.pb.go` and `macos/App/Platform/Generated/engine.pb.swift`, `api/ffi/handlers/conclusion_detail.go`, `conclusion_detail_test.go`, every `conclusionValueProto` caller (`rg -n 'conclusionValueProto\(' api`), `promote_evidence.go`, `promote_evidence_test.go`, `api/ffi/swift_vocabulary_test.go`, `macos/App/Platform/ConclusionVocabulary.swift`. Skill: `add-ffi-handler`.
Do: D-16, D-17; D-18's vocabulary (Go and Swift together, so the test passes).
Check: a second `./scripts/generate-proto.sh` leaves no diff; `CGO_ENABLED=1 go test -tags fts5 ./api/...` passes; T-9, T-10, T-11.

**STEP-9 — Swift**
Files: `macos/App/Platform/GenealogyStore.swift`, `GoStore.swift`, `macos/App/Features/Conclusions/ReconciledValueDisplay.swift`, `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`, `macos/ProvenenciaTests/ReconciledValueDisplayTests.swift`. Skill: `add-swift-test`, `add-localized-string`.
Do: D-18.
Check: `xcodebuild test -project macos/Provenencia.xcodeproj -scheme Provenencia -destination 'platform=macOS'` passes; `python3 scripts/check-localizable-xcstrings.py` passes; T-12.

**STEP-10 — Docs**
Files: `docs/deployment-plan/spike-11/claim-evidence.md` (Lines: likenesses judged by eye, `evidence.ByEye`; sections in binding order; the lifespan carried, never pinned), `docs/deployment-plan/spike-11/comparison-layers.md` (lifespan section: a side's missing reason is reported; reuse table: depiction keys in layer 1), `docs/matching-model.md` (row), `docs/catalog-deletes.md` (L50: `subject_type_properties` moves to the audited column, `property_id` released, `subject_type_id` audited with no delete path, worded as `identity_claims.entity_id`; L66: bindings are released and audited), `core/writes/README.md` (L15: only `EnsureBinding`, for `Install`, stays off `Run`), `docs/interpretation-layer-data-model.md` (L693: `sort_order` also orders the conclusion page and the evidence sections), `.agents/skills/add-catalog-delete/SKILL.md` (L28: the `subjectvocab` join-row remove is an audited `writes.Run` write), `.agents/skills/add-match-rule/SKILL.md` (the evidence-section rule names `ByEye` and that order comes from bindings), `docs/ideas/binding-order-editing.md` *(new)* (a per-type ordered bindings list on the Properties page calling `ReorderSubjectTypeProperties`; needs a brief) and its line in `docs/ideas/README.md`, `docs/deployment-plan/spike-11/deployment-plan.md` (map row), `docs/deployment-plan/spike-11/completed.md`, this plan (status, Landed note).
Do: per the Definition of done.
Check: links resolve; `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-11/prs/S11-32-lifespan-likeness-lines.md` reports no errors.

Commits: one per step (STEP-4's remaining test may land as its own commit), except STEP-8 and STEP-9, which land as one commit: `generate-proto.sh` writes the Go and Swift code together, and Swift builds only once STEP-9 is done (as S11-25 did). The subject is one full imperative sentence that says what changed for the researcher, ending in a period ("Show each record's likenesses beside this Source's on the evidence."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/valuecompare/depiction_test.go` `TestCompareDepiction` *(new)*, table-driven; `contract_test.go` rows; `TestCompareUnsupported` updated | Same id → `same`/`exact`/1; two ids → `unknown`/`unjudged`/0; nil or empty side → `unknown`/`empty`; `Key` equal exactly for one id; `ComparerTypes()` includes depiction | AC-1.1, D-1, D-2 |
| T-2 | Go | `core/subjectprofile/lifespan_test.go` `TestLifespanFromEvents` (rows gain `side`) | AC-2.11 → `undated`; AC-2.12, AC-2.16, AC-2.17 → `contradicts`; AC-2.13 (no rules) → zero side; AC-2.18 → `First` is the event with the earliest early end (the baptism "1850"); a two-event case with "abt 1815" → `First` kept as written; one event → `First == Last`; a likeness candidate → `Depiction` set on the value | AC-2.2, AC-2.3, AC-2.7, D-4, D-5 |
| T-3 | Go | `core/database/subjectvocab/subjectvocab_test.go` (updated orders); `core/database/graphcache/vocab_test.go` `TestVocabularyBindings` *(new)*; `core/vocabkey/vocabkey_test.go` `TestCompareByOrder` *(new)*; `core/database/effects` registry test (the `subject_type_properties` entry) | Install → person `likeness` 1, `gender` 3, `interests` 4; `BindingOrder()[person]` is `name`, `likeness`, `sex_at_birth`, `gender`, `interests`; `CompareByOrder` puts bound before unbound and unbound by key; (STEP-5) a `ReorderBindings` committed through `writes.Run` drops the vocabulary and the next read sees the new order | AC-3.1, AC-3.3, D-6, D-7 |
| T-4 | Go | `core/claimevidence/sections_test.go` `TestSections` (new rows) | AC-1.1's two lines with D-9's flags (`propertycomparetest.Depiction`); AC-1.4's collapse; an `unknown` line with cause `empty` dropped; AC-3.1, AC-3.2 and AC-3.4's orders | AC-1.1, AC-1.4, AC-3.1, AC-3.2, AC-3.4, D-8, D-9 |
| T-5 | Go | `core/claimevidence/build_test.go` `TestBuildLifespan` *(new)* | Each side built with `subjectprofile.ForSubject` (a test import) from life events, its `LifeSide` passed in `Inputs`: agreeing → result with both sides; one side `contradicts` → result, `unknown`, relation and cause empty; both `undated` → nil; an event kind → nil; AC-2.5's spans → `conflict` | AC-2.1, AC-2.2, AC-2.3, AC-2.4, AC-2.5, AC-2.6, D-10, D-11 |
| T-6 | Go | `core/database/promotealign/evidence_test.go` `TestEvidenceLikenesses` *(new)* | `SaveLikeness` shade 40 on the Source person and on member 1, shade 200 on member 2, and a third member whose image is set to `image/jpeg` as `conclusiondetails_test.go` L324 does (then `c.Graph().Drop()` or a fresh catalog open so the cache rereads) → AC-1.1's lines and no line for the third | AC-1.1, AC-1.4, AC-1.5, D-12 |
| T-7 | Go | same file `TestEvidenceLifespan` *(new)*, with `datedOn` | AC-2.1, AC-2.2, AC-2.3 and AC-2.5; AC-3.1's order through the loader | AC-2.1, AC-2.2, AC-2.3, AC-2.5, AC-3.1, D-13, D-14 |
| T-8 | Go | `pins_test.go` `TestResolvePinsLikeness` *(new)*; `promotealign_test.go` `TestProposeIgnoresLikenesses` *(new)*; `values_test.go` `TestEvidenceValues` (a depiction row) | Default pins → the same-image pair only; a proposal with and without likenesses → equal rows; `ResolvePaths` → `ImageRelPath` set for both images, different for A and B | AC-1.2, AC-1.3, AC-1.6, D-15 |
| T-9 | Go | `api/ffi/handlers/conclusion_detail_test.go` `TestConclusionValueProtoCoversEveryKind` | A depiction row: `conclusionValueProto(d, &DepictionValue{ImageRelPath: "objects/x.png"})` → the depiction case | D-16 |
| T-10 | Go | `api/ffi/handlers/promote_evidence_test.go` `TestGetPromoteClaimEvidenceLikeness` *(new)* | `newEvidenceFixture` plus `observationstest.SaveLikeness` on the incoming and one member (inside `withProjectCatalog`) → the likeness line's `incoming.depiction.image_rel_path` and its pin's path non-empty; `lifespan` unset (no dated events). The lifespan's dates are T-7's | AC-2.8, AC-4.2, D-16, D-17 |
| T-11 | Go | `api/ffi/swift_vocabulary_test.go` `TestSwiftVocabularyMatchesGo` | `unjudged` in value causes; `""` in value relations; a `LifespanMissing` row (`""`, `undated`, `contradicts`) and in the unrecognized subtest | AC-4.1, D-18 |
| T-13 | Go | `core/database/subjectvocab/subjectvocab_test.go` `TestReorderBindings` *(new)*, `TestBindingChanges` *(new)*; `core/database/properties/properties_delete_test.go` (a bound user Property) | Reorder: full permutation → orders 0..n-1 and one change per moved row; same order → no changes; missing, extra, repeated id → `ErrInvalid` and orders unchanged; locked binding reorders. Append → one create change, appending again → none; delete → one delete change carrying both ids; locked → `ErrLocked`. Property delete → the binding's release change before the Property's | AC-3.5, AC-3.6, AC-3.7, D-19, D-21, D-22 |
| T-14 | Go | `api/ffi/handlers/subject_defs_test.go` `TestReorderSubjectTypeProperties` *(new)*; `promote_evidence_test.go` `TestGetPromoteClaimEvidenceFollowsBindingOrder` *(new)* | RPC → response lists the new order; a bad list → `subject_vocab.invalid`; each of assign, remove, reorder adds one revision (`assertLatestAuditAction` in `harness_test.go`, actions `assign_subject_type_property`, `remove_subject_type_property`, `reorder_subject_type_properties`). Evidence open, reorder `interests` before `name`, evidence open in the same session → `interests` section first | AC-3.3, AC-3.5, AC-3.6, D-22 |
| T-12 | Swift | `macos/ProvenenciaTests/ReconciledValueDisplayTests.swift` `aDepictionReadsAsLikeness` *(new)* | `.depiction(CatalogDepictionValue(regionJSON: "", imageRelPath: "objects/x.png"))` → "Likeness" | AC-4.2, D-18 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `valuecompare` contract tests (D-1), `TestLayerPackages` (D-5), `TestKnobInventoryMatchesRegistry` (D-9), `subject_section_test.go` (fixture metas), S11-19's `TestEvidence*` (D-12, D-8), `TestEvidenceWarmIssuesNoQuery` (D-7), `autoreconcile` and `likenesses` tests (D-3), `TestSubjectTypePropertiesAndRegistry` and `TestCreateUpdateDeleteProperty` (D-19, D-21), `deleteimpact` and `effects` registry tests (D-20).

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests, catalog and plan-doc checks); on a host without Xcode, report the Go and plan-doc steps and say Xcode was not run
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings: conclusions.value.likeness
- [ ] UI matches the brief inventory — none: no chrome
- [ ] Version bump as the header says — none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan, and the exact names S11-20 reads (Lifespan, LifeSide, LifespanMissing, Inputs, the Swift types)
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ — none: not gated
- [ ] Docs: claim-evidence.md, comparison-layers.md, interpretation-layer-data-model.md, matching-model.md row, seeded-vocabulary.md, add-match-rule skill, docs/catalog-deletes.md, core/writes/README.md, add-catalog-delete skill row, docs/ideas/binding-order-editing.md (new) and its docs/ideas/README.md line
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

Data: a new catalog with S11-D2 §5.2's project: the obituary's John Smith with likeness photo A; the funeral card's member saved from the same image file; the album's member with photo B.

1. Open Promote on the obituary; open John's sheet. The likeness lines show in today's columns, worded "Likeness": the funeral card's checked, the album's unchecked. The "This subject" card lists Name, then Likeness, then Sex at birth. (AC-1.1, AC-3.1, AC-4.2)
2. Press Done on a fresh copy without opening the sheet: on the person's conclusion page, the new claim pins the funeral card's likeness and not the album's. (AC-1.3)
3. Proposal bands and contested marks are the same as on `main` for every row. (AC-1.6)
4. Run T-14 and read its output: the second evidence open lists Interests first. No gesture calls the RPC yet (Out). (AC-3.3)
5. Run T-7 and T-10 and read their output for the lifespan and the paths. (AC-2.8)

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| A likeness changes a proposal | Likeness has no weight; T-8 compares proposals with and without |
| The reconciler groups depictions differently | Keys are byte-identical (D-3); its existing tests run unchanged |
| A binding edit now drops the vocabulary and its memos | Binding edits are rare researcher actions on the Properties page; the drop is what makes AC-3.3 true |
| Order is ranked by key, not id: a user Property whose key equals a seeded one's shares its rank on the conclusion page | Order only, never which row shows; the evidence path already keys values by `PropertyKey` |
| Auditing bindings changes what a Property delete records | The release is one change per bound type, in the same revision (D-21); the cascade stays as a backstop |
| A member with two images shows only one line | The speaking pair is the strongest: a same-image pair wins; between two different images, the first in the member's value order speaks. Accepted: each member is judged by eye once |
| Existing catalogs keep likeness last | User decision (D-6): new seeds are for new catalogs; the Properties page reorders |

**Stop and report if:**

- `in.layer.Metas` has no `likeness` meta at your base, so `Sections` would skip the property.
- Two `SaveLikeness` calls with one shade on two Sources give two image File ids.
- `claimevidence` or `subjectcompare` needs an import of `core/subjectprofile`, `core/autoreconcile` or `core/database` to build.
- A proposal's band, score or contested flag changes in an existing `promotealign` test.
- `autoreconcile`'s depiction tests change value after D-3.
- After D-20, `TestCascadeIntoNonEmptyEffectIsAudited`, `TestFacetReleaseHonesty` or `TestResolversCoverEveryEntityType` fails for a table other than `subject_type_properties`.
- `audit.Record` refuses the `subject_type_property` change after D-20's effects entry.
- `ReorderSubjectTypeProperties` would need a Swift caller to compile or to pass `TestSwiftVocabularyMatchesGo`.
- `./scripts/generate-proto.sh` cannot run on your host: regenerated code is never written by hand.

## 12. Open questions

None. Two product calls were answered by the user on 2026-10-22: sections follow the catalog's binding order, and likeness is seeded second (D-6, D-8); binding writes are audited through `writes.Run` with a reorder RPC, after the ship loop's stop (D-19 – D-22).

## 13. PR

**Title:** S11-32 — Lifespan and likeness evidence lines
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-11/prs/S11-32-lifespan-likeness-lines.md`.
