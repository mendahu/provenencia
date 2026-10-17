# S10-39 — Preferred portrait

**Status:** Landed in [#450](https://github.com/mendahu/provenencia/pull/450) (2026-10-17)
**Spike / slice:** Spike 10 · Slice 8 — Likeness
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#s10-39--preferred-portrait)
**Brief:** [S10-D12](../design/archive/S10-D12-person-detail.md) (PDP-10, PDP-11 ship here), board agreed 2026-10-16; [S10-D9](../design/archive/S10-D9-persons-list.md) (PSL-8 ships here), board agreed before S10-22
**Board:** [Person Detail](https://claude.ai/design/p/c8660cdb-2c61-4c30-bc39-b2411036e6c0?via=share), file *Person Detail.dc.html*, the Portrait frame (brief §5 frame 10); [Persons list](https://claude.ai/design/p/5c00cb55-8261-4579-b0e6-bc571ab23f78?via=share), file *Persons List.dc.html*, the Preferred portrait frame (brief §5 frame 8). The implementer opens these links (user, 2026-10-17). §6 macOS also describes each frame in words. Where the board and a Decision differ, follow the Decision and report the difference in the PR body.
**Depends on:** S10-35 (landed [#446](https://github.com/mendahu/provenencia/pull/446); [`S10-35-person-gallery.md`](S10-35-person-gallery.md) and its Landed note), S10-D12, S10-D9
**Blocks:** —
**Branch:** `s10-39-preferred-portrait`
**Version bump:** none — Spike 10 PRs leave `VERSION` at 0.2.0, and the product bump is cut when a release ships ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-catalog-migration`](../../../../.agents/skills/add-catalog-migration/SKILL.md), [`add-catalog-model`](../../../../.agents/skills/add-catalog-model/SKILL.md), [`add-catalog-query`](../../../../.agents/skills/add-catalog-query/SKILL.md), [`add-audit-scope`](../../../../.agents/skills/add-audit-scope/SKILL.md), [`add-ffi-handler`](../../../../.agents/skills/add-ffi-handler/SKILL.md), [`use-catalog-session`](../../../../.agents/skills/use-catalog-session/SKILL.md), [`add-ui-component`](../../../../.agents/skills/add-ui-component/SKILL.md), [`evaluate-ui-component`](../../../../.agents/skills/evaluate-ui-component/SKILL.md), [`add-localized-string`](../../../../.agents/skills/add-localized-string/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md)
**Rules that bind:** `.cursor/rules/catalog-queries.mdc`, `.cursor/rules/ffi-handlers.mdc`, `.cursor/rules/go-tests.mdc`, `.cursor/rules/design-system-layers.mdc`, `.cursor/rules/macos-l10n.mdc`, `.cursor/rules/swift-tests.mdc`
**Context:** working rules "A preferred portrait is a display pointer, not a claim", "A gallery is the reconciled likeness values", "Last updated"; [depictions and likenesses](../depictions-and-likenesses.md) (the preferred-portrait paragraph); [S10-35 plan](S10-35-person-gallery.md) D-3, D-7, D-8, D-10, D-15, D-16 and Next adds; [S10-44 plan](S10-44-unpromote.md) (the last handle write over the FFI); [`conclusion-layer-data-model.md`](../../../conclusion-layer-data-model.md) §6
**Planned:** 2026-10-17 against `main` @ `ced9508` (S10-35 landed in #446); first drafted against `79f7f04` from the S10-35 plan. The refresh found every (S10-35) row true; the one change is the test fixture: S10-35 shares its likeness helper as `observationstest.SaveLikeness`, and `likenesses_test.go` is already `package likenesses_test` (D-14, T-1, T-2)
**Review:** [`S10-39-preferred-portrait.review.md`](S10-39-preferred-portrait.review.md) — Fits with changes (applied); Ready with changes (applied); 2 rounds, plus a refresh check after S10-35 landed; M-1 answered by the user (in this PR)

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

**Machines.** STEP-1 to STEP-4 and STEP-5's Go tests run on Linux. STEP-5's `./scripts/generate-proto.sh` needs `protoc`, `protoc-gen-go`, and `protoc-gen-swift`. STEP-6 to STEP-10's Swift checks, `scripts/check-pr.sh`'s Xcode stage, and §10 need a Mac. On a machine without them, run every check it can, and report each one it could not run as not run.

## Landed

Differences from this plan: `WorkspaceDestinationHost` also puts `.id(entityId)` on `PersonDetailView`, so the `@State` portrait model, which holds the entity id, is rebuilt when one Person page replaces another; `ThrowingStore` in `OnboardingModelTests.swift` gained a `setPreferredLikeness` stub (a `GenealogyStore` conformer the file list did not name). The `evaluate-ui-component` verdict (D-15): the tile and the artifact card share only the Badge and ghost-button pairing, so no shell is extracted and none is queued. The manual check in section 10 (VoiceOver, the 120pt column at the default size, the menu keyboard regression) needs a person at the Mac and was not run by the implementer; the Go and Swift tests in section 8 cover every AC except those. The Claude Design boards were not opened. `canonicalentities` binds the preferred id with the shared `database.NullBlob` instead of a private copy (user, 2026-10-17); the other packages' copies move in M-4.

## 1. Objective

The researcher marks one likeness in a Person's gallery as the portrait. That crop then leads the Person page header and is the Persons list thumb; the other likenesses stay in the gallery, in the same order. The header thumb's menu goes back to the earliest likeness. Deleting the chosen likeness, or that likeness leaving the gallery, puts the earliest back.

**Why now.** S10-35 shows every likeness and leads with the earliest. The earliest is often a poor face (a group photo, a blurred crop), and the map's Definition of done says a person can mark one likeness to lead the page and the People list.

**Also in this PR (user, 2026-10-17).** The kit's context menu resets its keyboard state when it opens, so the portrait menu does not add a fourth hand-written reset, and the Source cover menu and the composer's clear-region menu drop theirs (review M-1).

**What stays.** The gallery order (earliest first), the admit rule (`Reason.InGallery`), the Details rows, the gender Badge and tint, the Persons list layout (no portrait badge, no control: PSL-8), Events and Places, Promote. The portrait is not evidence: no Reconciliation Claim, no rank among values.

## 2. User stories

**US-1.** As a researcher, on a Person page with two or more likenesses, I choose one as the portrait so that the face I recognise leads the page and the list.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | Person with likeness A (earliest, *Register*) and B (*Census*), no portrait | Each tile shows a ghost “Use as portrait” button under its caption; no tile has a Badge; the header thumb is A |
| AC-1.2 | Click “Use as portrait” under B | `canonical_entities.preferred_likeness_observation_id` is B's record Observation id; B's tile shows the accent “Portrait” Badge with the photo icon and no button; A's tile keeps its button; tile order stays [A, B]; a success toast “Portrait updated” / “Using the likeness from Census as the portrait.” |
| AC-1.3 | After AC-1.2 | The header thumb shows B, spoken “Likeness from Census”; the Persons list row thumb shows B, with no badge or mark on the row (PSL-8) |
| AC-1.4 | After AC-1.2, click “Use as portrait” under A | The pointer is A's record; A has the Badge, B the button; one portrait at most |
| AC-1.5 | The write fails (engine error) | A danger toast whose title is the mapped error message; the page and tiles are unchanged |

**US-2.** As a researcher, I go back to the earliest likeness from the header thumb.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | Portrait B chosen; right-click the header thumb | A menu titled “Person portrait” with one item “Use earliest likeness” |
| AC-2.2 | Choose “Use earliest likeness” | The pointer is NULL; A leads the header and the list; every tile shows its button; a success toast “Portrait updated” / “Using the earliest likeness as the portrait.” |
| AC-2.3 | No portrait chosen; right-click the header thumb | No menu opens; the thumb has no menu help or hint |
| AC-2.4 | VoiceOver on the header thumb while a portrait is chosen | The thumb offers the action “Use earliest likeness” |

**US-3.** As a researcher, the portrait never shows a face the gallery does not.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | Portrait B; delete B's Observation in the composer | The pointer is NULL (`ON DELETE SET NULL`); A leads; the gallery shows only A |
| AC-3.2 | Portrait B; B's Subject is unpromoted, or a stronger negative denies B's image | B is not in the gallery; A leads; no tile shows the Badge. The pointer row is unchanged |
| AC-3.3 | Set the portrait to an Observation that is not one of this handle's gallery likenesses (another Person's likeness, an id that names no Observation) | Refused with `canonicalentities.invalid`; nothing written |
| AC-3.4 | Portrait B already chosen; choose B again through the FFI | No revision is written; effects are empty |

## 3. Scope

| In | Covered by |
| --- | --- |
| Migration: `canonical_entities.preferred_likeness_observation_id` → `observations(id)` `ON DELETE SET NULL`, with its index; `catalogmodel` FK `BucketOptional` | STEP-1, T-1 |
| `likenesses.Likeness.Preferred` and `likenesses.Lead` (the one lead rule) | STEP-2, T-1 |
| `canonicalentities.SetPreferredLikeness`, audited, refused outside the gallery | STEP-3, T-2 |
| `conclusiondetails.LikenessTile.Preferred`; the Person header lead through `likenesses.Lead` | STEP-4, T-3, T-4 |
| FFI: `SetPreferredLikeness` method 101, `LikenessDetail.preferred` | STEP-5, T-5 |
| Swift store method and mapping; `FakeStore` stub | STEP-6, T-7 |
| `ConclusionPortraitModel`: write, apply effects, toast (PDP-10 toast) | STEP-7, T-7 |
| Tile Badge and “Use as portrait” button (PDP-10) | STEP-9, T-6 |
| Kit: `pvContextMenu(…, keyboardTitles:)` resets the keyboard on open; `PVContextMenuKeyboard.opened(titles:)`; the Source cover and composer clear menus move onto it (review M-1) | STEP-8, T-8 |
| Header thumb menu “Use earliest likeness” (PDP-11) | STEP-9, T-6 |
| Persons list thumb shows the portrait (PSL-8) | STEP-4 (no Swift change), T-3, §10 step 3 |
| Docs: domain doc §6, brief archive, map, completed | STEP-10 |

| Out | Home |
| --- | --- |
| A portrait for a place or an event | Not planned (map row S10-39 Out) |
| Picking a portrait from Promote or from the Persons list | Not planned (working rule; S10-D12 non-goals) |
| A second Person-detail or Persons-list brief | Not planned (map row S10-39 Out) |
| Reordering the gallery by the portrait | Not planned (PDP-11: order does not change) |
| An audit row for the FK's own `SET NULL` on delete | Not planned: same as `sources.primary_artifact_id` (D-3) |
| Clearing a pointer whose likeness left the gallery | Not planned: the read falls back (D-5); the pointer comes back into force if the likeness returns |
| Keeping the pointer through a handle merge | Not planned: there is no merge path yet ([`conclusion-layer-data-model.md`](../../../conclusion-layer-data-model.md) §5) |
| Extracting a shared image-card shell from the artifact card and the likeness tile | Not planned in this PR (D-15) |
| Moving the handler tests' likeness helpers onto `observationstest` | Not planned in this PR (S10-52 Next adds, still open) |
| Moving `PVSelect` onto the open-time reset | Not planned: it seeds the highlight from its session's `highlightIndex`, not -1, so it passes no `keyboardTitles` (D-16) |
| One promoted-Person-with-likenesses fixture shared by `likenesses`, `canonicalentities`, `conclusionheaders`, and S11-20 tests | Queued follow-up task (review M-3) |
| One `database.NullBlob` and audit UUID encoder instead of per-package copies | Queued follow-up task (review M-4) |

## 4. Current state

Rows marked **(S10-35)** are code S10-35 added in #446, verified at the planning base.

| Area | Where | Today |
| --- | --- | --- |
| Handle table | `core/database/migrations/000032.sql` `canonical_entities` | `id`, `subject_type_id`, `ref`, `argument`, `label`, `merged_into_id`. No display pointer. Latest migration is `000053.sql` |
| Back-pointer precedent | `core/database/migrations/000015.sql` (`sources.primary_artifact_id … ON DELETE SET NULL`), `000031.sql` (`sources_primary_artifact_id_idx`); `core/database/catalogmodel/model.go` FK row `sources.primary_artifact_id` `BucketOptional` | The optional back-pointer shape. `deleteimpact` has no code for `BucketOptional` (`TestResourceAndOwnedHonesty` probes resource FKs only). Artifact delete clears the pointer silently, with no audit row |
| Setter precedent | `core/database/sources/sources.go` `SetCover` | Reads the previous row, returns no changes when unchanged, writes one `rowchange.Change` (`ActionUpdate`, `uuidJSON` old/new) |
| Handle package | `core/database/canonicalentities/canonicalentities.go` `Create`, `InsertTx`, `Get`, `GetManyTx`, `Entity`, `ErrInvalid`, `uuidJSON` | No update of any column. `Entity` and `sqlColumns` have no pointer |
| Effects | `core/database/effects/registry.go` `"canonical_entities": {Entity: "canonical_entity", Handles: self(), HandleScope: self()}` | A `canonical_entity` change recomputes that handle, moves its `canonical_entity` audit scope, and reports it in `Handles` and `HandleScopes`. `TestResolversCoverEveryEntityType` already accepts `canonical_entity` |
| Write path | `core/writes/run.go` `Run`; `core/database/graphcache/commit.go` `OnCommit` | `Run` recomputes `set.Handles` before commit; `OnCommit` drops the stored display rows of every handle in `set.Handles`, so a pointer write drops that Person's stored header |
| Gallery read **(S10-35)** | `core/database/likenesses/likenesses.go` `Likeness{Rank, ImageFileID, ObservationIDs}`, `ForEntities(q sqlbatch.Querier, entityIDs [][]byte)`, `sqlGallery` | One `sqlbatch.Query` over `auto_reconciler_values` ⨝ `properties` ⨝ `auto_reconciler_outcomes` ⟕ `files`; the scan callback drops a row whose reason is not `InGallery` or whose image `depictionvalues.ReadableImage` empties, groups the rest by (entity, rank) into `byValue`, then sorts each handle's list by `ObservationIDs[0]`, then `Rank`. No pointer |
| Detail **(S10-35)** | `core/database/conclusiondetails/conclusiondetails.go` `LikenessTile{ImageFileID, Record}`, `Detail.Likenesses` | One tile per gallery likeness, in `ForEntities` order |
| Header **(S10-35)** | `core/database/conclusionheaders/conclusionheaders.go` `PersonHeader.LikenessImageFileID`; `compose.go` `composePersons`, `onePerson(…, lead []byte)` | `composePersons` calls `ForEntities(g, need)` once; per id, `if ls := gallery[string(id)]; len(ls) > 0 { lead = ls[0].ImageFileID }`, then `onePerson(v, byID[string(id)], g, pg, lead)`. S10-35's Next adds: S10-39 changes only this choice |
| Proto **(S10-35)** | `api/proto/engine.proto` `LikenessDetail{image_rel_path = 1, record = 2}`, `PersonHeader.likeness_image_rel_path = 8`, `ConclusionDetail.likenesses = 7`; `Method` ends at `METHOD_UNPROMOTE_SUBJECT = 100` | No pointer field, no set method |
| FFI write precedent | `api/ffi/handlers/promote.go` `UnpromoteSubject`; `api/ffi/handlers/write_effects.go` `writeEffectsProto`; `api/ffi/dispatch.go` `MethodUnpromoteSubject`; `api/libprovenencia/call_test.go` "unknown method carries Error protobuf" (`method: 101`) | `writes.Run` inside `withProjectCatalog`, response `WriteEffects`. The unknown-method test uses 101, which this PR takes |
| Test fixtures | `core/database/observations/observationstest/observationstest.go` `SaveLikeness(t, c, userID, sourceID, subjectID, shade)` (returns the `observations.Observation`); `core/database/likenesses/likenesses_test.go` (`package likenesses_test`) `newFixture`, `source`, `personOn`, `promote`, `TestForEntities`; `core/database/conclusionheaders/conclusionheaders_test.go` `TestPersonHeaderLeadsWithEarliestLikeness`; `core/database/conclusiondetails/conclusiondetails_test.go` `TestForEntityLikenesses` | S10-35 shares the likeness save in `observationstest` (its Landed note) |
| Handler tests | `api/ffi/handlers/citations_test.go` `likenessFixture`, `saveLikeness` (returns `CreateCitationWithObservationsResponse`, whose `observations[0].id` is the record), `api/ffi/handlers/thumbnails_test.go` `writeCrop`; `api/ffi/handlers/conclusion_detail_test.go` `TestGetConclusionDetailListsLikenesses` | Fixture with one person and the `likeness` Property; `saveLikeness` returns the created Observations |
| Swift store | `macos/App/Platform/GenealogyStore.swift` `unpromoteSubject`, `CatalogLikeness`; `macos/App/Platform/GoStore.swift` `unpromoteSubject`, `mapWriteEffects`, `getConclusionDetail`; `macos/App/Platform/CoreInvoke.swift` `CoreMethod.unpromoteSubject`; `macos/App/Platform/FakeStore.swift` `unpromoteSubject`, `unpromoteSubjectError`, `recordedCalls` | The write-returns-effects shape to copy |
| Session | `macos/App/Features/Workspace/Session/WorkspaceSession.swift` `apply(_:_:)`; `macos/App/Features/Workspace/Session/CatalogQueryRegistry.swift` `invalidations(for:project:)` | `effects.handles` invalidates the Persons, Events, and Places lists and each handle's `conclusionDetail`; `handleScopes` patch Last updated |
| Page **(S10-35)** | `macos/App/Features/Conclusions/ConclusionDetailPage.swift` `ConclusionDetailPage`, `ConclusionDetailLoaded`, `ConclusionDetailHeader`, `ConclusionDetailThumbnail`, `ConclusionDetailLikeness`, `ConclusionDetailLikenessSection` | Read-only page: no store, no user id, no toast. Header thumb is `CachedThumbnail` with no menu. A tile is `VStack { thumb button; caption }` in a 120pt column |
| Person view | `macos/App/Features/Conclusions/PersonDetailView.swift` `PersonDetailView(session:entityId:)`; `PersonDetailContent.swift` `likenesses`, `thumbnail`, `leadLabel` (S10-35 D-11, D-15) | No user id; `WorkspaceDestinationHost` mounts it with `session` and `entityId` only |
| Host | `macos/App/Features/Workspace/WorkspaceDestinationHost.swift` `.personDetail` case | Has `userID` and `store` in scope (passes both to `PropertiesView`) |
| Cover UI precedent | `macos/App/Features/Sources/SourcePageArtifactsView.swift` `metaRow` (Cover Badge, “Set cover” ghost button); `SourceArtifactsSection.swift` `useAsThumbnail`, `revertCoverToTypeIcon`; `SourcePageIdentityHeader.swift` `coverThumbnail`, `coverMenuPanel`, `coverMenu`, `coverKeyboard` | Badge `PVBadge(…, tone: .accent, icon: .photo)`; `PVButton(…, variant: .ghost, size: .sm, icon: .imageUp)`; the thumb carries `.pvContextMenuTrigger`, the header row `.pvContextMenu`; success `VocabularyToast` |
| Menu keyboard | `macos/App/DesignSystem/Components/ContextMenu/PVContextMenu.swift` `PVContextMenuKeyboard`, `pvContextMenu(_:keyboard:stealKeys:dismissOnMouseUp:dismissOnClickAway:content:)`, private `PVContextMenuPresenter` (`onChange(of: state.isPresented)`); call sites `macos/App/Features/Sources/SourcePageIdentityHeader.swift` (`coverThumbnail`'s `onChange` sets `coverKeyboard = PVContextMenuKeyboard(itemCount: coverMenuItemCount, activeIndex: -1)`, no `itemTitles`), `macos/App/Features/CitationComposer/CitationComposerViewerPane.swift` `toggleClearMenu` (sets `clearKeyboard` with two titles before `present`), `macos/App/DesignSystem/Components/Select/PVSelect.swift` `presentMenu` (sets `activeIndex: session.highlightIndex`); tests `macos/ProvenenciaTests/PVContextMenuTests.swift` | The presenter takes the keyboard binding but never resets it. Each right-click menu resets it by hand, and the Source cover's copy leaves type-to-select off |
| Toast host | `macos/App/Features/CatalogVocabulary/VocabularyChrome.swift` `vocabularyToastOverlay(_:identifier:)`; `SourcePageView.swift` (`.vocabularyToastOverlay($model.toast, …)` on an `@State` model) | Top-trailing toast; a title-only toast has no body line |
| Swift tests | `macos/ProvenenciaTests/PersonDetailTests.swift` (S10-35 adds `headerThumbIsTheLeadLikeness`, `likenessesBecomeTilesEarliestFirst`); `macos/ProvenenciaTests/EvidenceGraphModelTests.swift` `confirmUnpromoteRefreshesToPromote`, `unpromoteFailureStaysInConfirm` | The model-over-`FakeStore` test shape; sessions are `WorkspaceSession(projectKey:store:)` |

**Precedent.** Engine: `sources.primary_artifact_id` (`000015.sql`, `000031.sql`, `sources.SetCover`) for the column, its index, its `BucketOptional` registration, and the setter that writes no change when unchanged. FFI and Swift store: S10-44's `UnpromoteSubject` (`2af1f54` engine call; `0f70c8e` Swift store) for a method that returns `WriteEffects` and a Swift call that hands them to `session.apply`. UI: the Source page cover (S10-18): Cover Badge, ghost button, toast, and the cover thumb's revert menu. Copy those; do not copy the Source cover's always-present menu with a disabled “Default icon is in use” row (PDP-11 shows the menu only while a portrait is chosen).

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | Migration `core/database/migrations/000054.sql` *(new)*: `ALTER TABLE canonical_entities ADD COLUMN preferred_likeness_observation_id BLOB REFERENCES observations(id) ON DELETE SET NULL;` and `CREATE INDEX canonical_entities_preferred_likeness_observation_id_idx ON canonical_entities(preferred_likeness_observation_id);`. `catalogmodel.FKs` gains `{From: "canonical_entities", Column: "preferred_likeness_observation_id", To: "observations", OnDelete: "SET NULL", Bucket: BucketOptional}` beside the two `canonical_entities` rows. No backfill. | The map row names this column and shape. The honesty tests require a covering index on every FK child column (`add-catalog-migration`). `SET NULL` clears the pointer when its Observation is deleted, with no `deleteimpact` code (AC-3.1). | A side table `canonical_entity_portraits` — a second row per handle for one optional id. A pointer to the image File — a region redraw replaces the File and would drop the choice. |
| D-2 | `likenesses.Likeness` gains `Preferred bool` *(new)*: true when `ByObservation` finds it by the handle's `preferred_likeness_observation_id`. `sqlGallery` adds `JOIN canonical_entities e ON e.id = v.entity_id` and selects `e.preferred_likeness_observation_id` as its last column; The scan callback stores each row's pointer in a local `map[string][]byte` keyed by `string(entity id)` before its `InGallery` and `ReadableImage` checks; after sorting each handle's list, `ForEntities` sets `ls[i].Preferred = true` where `i, ok := ByObservation(ls, pointer)` and `ok`. Order is unchanged (earliest first). `func ByObservation(gallery []Likeness, observationID []byte) (i int, ok bool)` *(new)*: the index of the entry whose `ObservationIDs` holds `observationID` (`bytes.Equal`), `-1, false` when none; doc "the gallery likeness an Observation went into; the one rule for which likeness a record names". `func Lead(gallery []Likeness) (Likeness, bool)` *(new)*: the first `Preferred` entry, else `gallery[0]`; `false` when the gallery is empty. Doc: "the likeness a handle leads with: its preferred portrait when that likeness is in the gallery, else the earliest (working rule \"A preferred portrait is a display pointer, not a claim\")". | The map row: the pointer is read through S10-35's admit rule. Reading it in the gallery query means a pointer to a denied, provisional, moved, or non-PNG likeness matches nothing, so the fallback is free and there is one rule. Still one query per call. | A second query for the pointer in `composePersons` — a second place that decides whether the pointer counts. Adding the column to `canonicalentities.Entity` and `graphcache.Node` — every handle read and `Verify` grows for one Person-only display field. |
| D-3 | `func SetPreferredLikeness(tx *database.Tx, userID, entityID, observationID []byte) ([]rowchange.Change, error)` *(new)* in `canonicalentities`. Empty `observationID` clears. Checks in order: `tx` non-nil and `len(entityID) == 16`, else `ErrInvalid`; `database.RequireUserID(userID, ErrInvalid)`; the previous pointer from `sqlPreferred = SELECT preferred_likeness_observation_id FROM canonical_entities WHERE id = ?` *(new)* (`sql.ErrNoRows` → `ErrInvalid`); a non-empty `observationID` must have `len == 16` and `likenesses.ByObservation(likenesses.ForEntities(tx.Tx, [][]byte{entityID})[string(entityID)], observationID)` must report `ok`, else `ErrInvalid`; a `ForEntities` error is returned as is. Equal to the previous pointer (`bytes.Equal`) → `nil, nil`. Else `UPDATE canonical_entities SET preferred_likeness_observation_id = ? WHERE id = ?` with `nullBlob` *(new, private)* and one change: `EntityType: "canonical_entity"`, `EntityID: entityID`, `Action: rowchange.ActionUpdate`, `Fields: {"preferred_likeness_observation_id": {Old: uuidJSON(prev), New: uuidJSON(observationID)}}`. | The `sources.SetCover` shape. The table's package owns its writes (`add-catalog-query`). Go refuses a pointer the read would ignore, so the stored value is always a gallery likeness at write time (AC-3.3). The existing `canonical_entities` effect recomputes the handle, drops its stored header, and moves its Last updated, which the working rule "Last updated" names ("its own row"). An unchanged choice writes no revision (AC-3.4). | Checking in the FFI handler — validation outside core. A dedicated effect entry with no `Handles` — the stored Person header would not be dropped and would keep the old lead. |
| D-4 | `conclusiondetails.LikenessTile` gains `Preferred bool` *(new)*, copied from `likenesses.Likeness.Preferred` where `ForEntity` builds `Detail.Likenesses` (S10-35 D-7). A gallery likeness whose lead record has no outcome row gives no tile (S10-35 keeps that skip); S10-35's stop conditions name that case as impossible, so the header lead always has a tile. | The tile Badge needs it; the same read already decided it. | Comparing the tile's image with the header lead in Swift — a rule in two places. |
| D-5 | In `composePersons`, the `lead` passed to `onePerson` becomes `l.ImageFileID` where `l, ok := likenesses.Lead(gallery[string(id)])`, nil when `!ok`, replacing S10-35's `[0].ImageFileID`. `PersonHeader.LikenessImageFileID`'s doc line "S10-35: the earliest gallery likeness." becomes "The preferred portrait when its likeness is in the gallery, else the earliest (likenesses.Lead)." | S10-35 Next adds: S10-39 changes only this choice; the page header, the list thumb, the proto path, and the Swift thumbs all read `LikenessImageFileID` already (PSL-8 needs no Swift change). | Choosing the lead in the handler or in Swift — the stored header and `checkDisplay` would disagree with what is shown. |
| D-6 | Proto: `LikenessDetail.preferred = 3` *(new)*, doc "true for the likeness chosen as the Person's portrait (S10-39); at most one". `PersonHeader.likeness_image_rel_path`'s doc "(S10-35: the earliest gallery likeness)" becomes "(the preferred portrait, else the earliest gallery likeness)". `METHOD_SET_PREFERRED_LIKENESS = 101` *(new)*; `message SetPreferredLikenessRequest { string project_dir = 1; string user_id = 2; string entity_id = 3; string observation_id = 4; }` *(new)*, doc "observation_id is a gallery likeness's record Observation; empty clears the choice"; `message SetPreferredLikenessResponse { WriteEffects effects = 1; }` *(new)*. The `call_test.go` unknown-method case moves from 101 to 102 in both places: `method: 101` and the params assertion `"101"`. | The S10-44 shape: the session refreshes from `WriteEffects`, so the response carries no detail. One method sets and clears, like `SetSourceCover`. | Returning the new `ConclusionDetail` — the session already reloads it from `handles`. Two methods (set, clear) — the same write twice. |
| D-7 | Handler `SetPreferredLikeness(in []byte) ([]byte, error)` *(new)* in `api/ffi/handlers/conclusion_detail.go`, after `GetConclusionDetail`: unmarshal (`unmarshalErr("set_preferred_likeness", err)`), `parseUserID`, `parseID(entity_id)`, `parseID(observation_id)` only when non-empty, then `withProjectCatalog` → `writes.Run(c, writes.Op{Action: "set_preferred_likeness", UserID: userID}, …canonicalentities.SetPreferredLikeness…)` → `SetPreferredLikenessResponse{Effects: writeEffectsProto(wrote)}`. Dispatch: `MethodSetPreferredLikeness` *(new)* in `api/ffi/dispatch.go`. S10-35's `attachLikenesses` sets `Preferred: l.Preferred` on each `LikenessDetail` it appends. | `UnpromoteSubject` line for line. The Person page's reads live in this file. | A new handler file for one method. |
| D-8 | Swift store: `func setPreferredLikeness(projectDir: String, userID: String, entityID: String, observationID: String) async throws -> CatalogWriteEffects` *(new)* on `GenealogyStore` (doc: "Choose a gallery likeness as the Person's portrait, or clear it with an empty observationID. A display choice, not evidence."), `GoStore` (`CoreMethod.setPreferredLikeness` *(new)* in `CoreInvoke.swift`; `Self.mapWriteEffects(resp.effects)`), and `FakeStore`: records `"setPreferredLikeness entity=\(entityID) observation=\(observationID)"`, throws `setPreferredLikenessError` *(new)* when set, else returns `CatalogWriteEffects(handles: [entityID], handleScopes: [entityID])`; it does not change seeded details. `CatalogLikeness.preferred: Bool = false` *(new)*; `GoStore` maps it. | The S10-44 store shape. Go owns the read, so the fake reports effects only (S10-35 D-19). | Mutating `FakeStore`'s seeded detail — a second copy of the lead rule. |
| D-9 | `@MainActor @Observable final class ConclusionPortraitModel` *(new)* in `macos/App/Features/Conclusions/ConclusionPortraitModel.swift` *(new)*: `init(session: WorkspaceSession, userID: String, entityId: String)`; `var toast: VocabularyToast?`; `private(set) var isSaving = false`; `func choose(_ likeness: ConclusionDetailLikeness) async` sends `likeness.id`; `func useEarliest() async` sends `""`. Each: return when `isSaving`; set it, `defer` clear; `let effects = try await session.store.setPreferredLikeness(projectDir: session.projectKey.projectDir, userID:, entityID:, observationID:)`; `session.apply(effects)`; `toast = VocabularyToast(title: L10n.Conclusions.toastPortraitUpdatedTitle, body: <chosen or earliest body>, tone: .success)` (§6 Strings). A thrown error sets `toast = VocabularyToast(title: L10n.Errors.message(for: error), body: "", tone: .danger)`. The file is added to the `Provenencia` target in `project.pbxproj` the way `PersonGenderTreatment.swift` was (`41fc9db`), with the next two free ids in that series, `A90A00000000000000000005` (file reference) and `A90A00000000000000000006` (build file). | The page is read-only today and holds no state. A model owns the write and the toast, like `SourceArtifactsSection`; `session.apply` refreshes the page and the list (AC-1.3, AC-2.2). An error toast keeps the page usable (AC-1.5). | Writing from the view — no test seam. A `PVCallout` on the page — a page-level error for a one-click choice. |
| D-10 | `PersonDetailView` gains `let userID: String` and `@State private var portrait: ConclusionPortraitModel`, built in `init(session:entityId:userID:)` with `_portrait = State(initialValue: …)`, as `SourcePageView` does. It passes `portrait: portrait` to `ConclusionDetailPage` and applies `.vocabularyToastOverlay($portrait.toast, identifier: "persons.detail.toast")` to it. `WorkspaceDestinationHost` passes `userID: userID`. `ConclusionDetailPage` gains `var portrait: ConclusionPortraitModel? = nil` (after `errorIdentifier`), passed to `ConclusionDetailLoaded` and from there to `ConclusionDetailHeader` and `ConclusionDetailLikenessSection` (each `let portrait: ConclusionPortraitModel?`). Event and Place pass nothing. | Only the Person page chooses a portrait; the shared page draws the controls only when a kind hands it a model. The model's name says what it holds, not the kind, so the shared page does not name Person. | A protocol member on `ConclusionDetailBody` — content is pure data and holds no actions. |
| D-11 | `ConclusionDetailLikeness.isPortrait: Bool = false` *(new)*, set in `PersonDetailContent.init` from `record.preferred`. `ConclusionDetailThumbnail.isPortrait: Bool = false` *(new)*, doc "the thumb shows a chosen portrait, so its menu offers the earliest"; `PersonDetailContent.thumbnail` sets it to `likenesses.contains(where: \.isPortrait)`. | One Go read decides both; Swift copies flags. | Comparing `likenessImageRelPath` with the earliest tile in Swift. |
| D-12 | Tile controls (PDP-10), in `ConclusionDetailLikenessSection`, only when `portrait` is non-nil: under the caption, in the same `VStack(alignment: .leading, spacing: PVSpacing.space3)`: when `tile.isPortrait`, `PVBadge(L10n.Conclusions.portraitBadge, tone: .accent, icon: .photo)` with `.accessibilityIdentifier("persons.detail.likeness.\(tile.id).portraitBadge")`; otherwise `PVButton(L10n.Conclusions.useAsPortrait, variant: .ghost, size: .sm) { Task { await portrait.choose(tile) } }`, `.disabled(portrait.isSaving)`, `.fixedSize(horizontal: true, vertical: false)`, `.accessibilityLabel(L10n.Conclusions.useAsPortraitLabel(tile.sourceTitle))`, `.accessibilityIdentifier("persons.detail.likeness.\(tile.id).usePortrait")`. The button has no icon. | The brief's inventory rows (ghost small button, accent Badge with the photo icon) and the Source cover's shapes. Without an icon the label fits the 120pt column at the `sm` metrics (10pt padding each side, caption font); with `.imageUp` it would pass 120pt and push into the next column. The spoken label names the tile because several buttons read alike. | The Source button's `.imageUp` icon — wider than the column. |
| D-13 | Header thumb menu (PDP-11), in `ConclusionDetailHeader`: `@State private var portraitMenu = PVContextMenuState()` and `@State private var portraitKeyboard = PVContextMenuKeyboard.inactive`. `showsPortraitMenu = portrait != nil && content.thumbnail.isPortrait`. When true, the `CachedThumbnail` gets `.pvContextMenuTrigger($portraitMenu)`, `.help(L10n.string(L10n.Conclusions.portraitMenuHint))`, `.accessibilityHint(L10n.string(L10n.Conclusions.portraitMenuHint))`, `.accessibilityAction(named: Text(L10n.Conclusions.useEarliestLikeness)) { Task { await portrait?.useEarliest() } }`, `.accessibilityIdentifier("persons.detail.portrait")`, and no `onChange` reset: the header `HStack`'s `.pvContextMenu` passes `keyboardTitles: [L10n.string(L10n.Conclusions.useEarliestLikeness)]` (D-16). When false, the thumb draws as S10-35 left it. The header `HStack` carries `.pvContextMenu($portraitMenu, keyboard: $portraitKeyboard) { PVContextMenuPanel(title: L10n.Conclusions.portraitMenuTitle, accessibilityIdentifier: "persons.detail.portrait.menu") { PVContextMenuItem(L10n.Conclusions.useEarliestLikeness, index: 0, accessibilityIdentifier: "persons.detail.portrait.useEarliest") { Task { await portrait?.useEarliest() } } } }`, the Source header's placement (menu on the row, trigger on the thumb). | PDP-11: the menu is shown only while a portrait is chosen, with one item. VoiceOver needs a way to the item without a right-click (AC-2.4). | The Source cover's always-present menu with a disabled row — PDP-11 hides it. A visible “Use earliest” button — the brief puts it on the menu. |
| D-14 | Tests for the write live in `core/database/likenesses/likenesses_test.go` (`package likenesses_test`, so importing `promote`, which imports `canonicalentities`, is no cycle) beside S10-35's `TestForEntities`, using its `newFixture`, `source`, `personOn`, `promote`, and `observationstest.SaveLikeness`. T-1's pointer setup uses `writes.Call`; T-2 uses `writes.Run`, because it asserts the result's effects and revision. | That file already builds promoted Persons with likenesses; S10-35 put the shared save in `observationstest`. A new fixture in `canonicalentities_test.go` would copy `newFixture` and `promote` again. | A fixture in `canonicalentities_test.go`; an internal `package likenesses` test (an import cycle once `canonicalentities` imports `likenesses`). |
| D-16 | Kit (review M-1, user 2026-10-17): `PVContextMenuKeyboard` gains `static func opened(titles: [String]) -> PVContextMenuKeyboard` *(new)* = `PVContextMenuKeyboard(itemCount: titles.count, activeIndex: -1, itemTitles: titles)`, doc "the state of a menu that just opened: nothing highlighted, type-to-select over `titles`". `pvContextMenu` gains `keyboardTitles: [String]? = nil` *(new)* after `keyboard`, passed to `PVContextMenuPresenter` (`var keyboardTitles: [String]?`). In the presenter's existing `onChange(of: state.isPresented)`, when `open` and both `keyboard` and `keyboardTitles` are non-nil, it sets `keyboard.wrappedValue = .opened(titles: keyboardTitles)` before deferring the dismiss monitor. Doc on `keyboardTitles`: "row titles of a right-click menu; when set, the keyboard is reset on each open. Leave nil when the caller seeds its own highlight (`PVSelect`)." Call sites: `SourcePageIdentityHeader` deletes its `onChange(of: coverMenu.isPresented)` block and passes `keyboardTitles: coverMenuItemCount > 0 ? [L10n.string(L10n.Sources.thumbnailRevertToDefault)] : []` (gains type-to-select); `CitationComposerViewerPane.toggleClearMenu` deletes its `clearKeyboard = PVContextMenuKeyboard(…)` assignment and the `.pvContextMenu($clearMenu, …)` call passes `keyboardTitles:` with the same two titles; the portrait menu passes its one title (D-13). `PVSelect` is unchanged. | One open-time rule in the kit; the next right-click menu gets the reset and type-to-select by passing its titles. `PVSelect`'s highlight comes from its session, so the nil default keeps it. | A `keyboard` binding the kit owns internally — every caller already holds `@State` for it, and `PVSelect` reads it. Resetting on every open even without titles — would clobber `PVSelect`'s seeded highlight. |
| D-15 | No shared card shell is extracted in this PR. S10-35's Next adds asked S10-39 to run `evaluate-ui-component` on the image-tile shell first. Run it on `ConclusionDetailLikenessSection`'s tile and `SourcePageArtifactsView`'s artifact card; record its verdict in the PR body. When it proposes an extraction, queue it as a follow-up task and do not do it here. | The tile gains two controls, the artifact card already has them; the two differ in layout (96pt crop in a 120pt column vs a full-width well with a meta row and edit/delete). An extraction is its own reviewable change. | Extracting in this PR — a refactor of the Source page inside a Person feature. |

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `000054.sql` *(new)*, `canonical_entities.preferred_likeness_observation_id` *(new)* | migration / column | `sources.primary_artifact_id` (`000015.sql`, `000031.sql`) | Same shape on another table (D-1) |
| `canonicalentities.SetPreferredLikeness` *(new)*, `sqlPreferred` *(new)*, `nullBlob` *(new)* | func / const | `sources.SetCover`, `sources.nullBlob` | First update in the package; the private helper is two lines and every package keeps its own (`observations.nullBlob`, `sources.nullBlob`) |
| `likenesses.Likeness.Preferred` *(new)*, `likenesses.ByObservation` *(new)*, `likenesses.Lead` *(new)* | field / func | S10-35's `ForEntities` | Extends the one gallery read (D-2) |
| `conclusiondetails.LikenessTile.Preferred` *(new)* | field | S10-35's `LikenessTile` | Extends it (D-4) |
| `LikenessDetail.preferred` *(new)*, `SetPreferredLikenessRequest` *(new)*, `SetPreferredLikenessResponse` *(new)*, `METHOD_SET_PREFERRED_LIKENESS` *(new)* | proto | `UnpromoteSubjectRequest` / `UnpromoteSubjectResponse`, `SetSourceCoverRequest` | Same shape; no existing method writes a handle row |
| `handlers.SetPreferredLikeness` *(new)*, `MethodSetPreferredLikeness` *(new)* | FFI | `handlers.UnpromoteSubject` | Same shape (D-7) |
| `GenealogyStore.setPreferredLikeness` *(new)*, `CoreMethod.setPreferredLikeness` *(new)*, `FakeStore.setPreferredLikenessError` *(new)*, `CatalogLikeness.preferred` *(new)* | store API | `unpromoteSubject`, `unpromoteSubjectError` | Same shape (D-8) |
| `ConclusionPortraitModel` *(new)* | type / file | `SourceArtifactsSection.useAsThumbnail` / `revertCoverToTypeIcon` | Those live on the Source page model; the Person page has none (D-9) |
| `ConclusionDetailLikeness.isPortrait` *(new)*, `ConclusionDetailThumbnail.isPortrait` *(new)*, `ConclusionDetailPage.portrait` *(new)* | field / param | S10-35's `ConclusionDetailLikeness`, `ConclusionDetailThumbnail` | Extends the one slot each (D-10, D-11) |
| `PVContextMenuKeyboard.opened(titles:)` *(new)*, `pvContextMenu(…, keyboardTitles:)` *(new)* | kit API | three hand-written resets (Source cover, composer clear menu, `PVSelect`) | Extends the one presenter; `PVSelect` keeps its own seeded highlight (D-16) |
| Portrait menu and tile controls in `ConclusionDetailHeader` / `ConclusionDetailLikenessSection` | view code | `SourcePageIdentityHeader.coverMenuPanel`; `SourcePageArtifactsView.metaRow` | Kit components instanced the same way; no new component (D-12, D-13, D-15) |
| `L10n.Conclusions.portraitBadge`, `useAsPortrait`, `useAsPortraitLabel(_:)`, `useEarliestLikeness`, `portraitMenuTitle`, `portraitMenuHint`, `toastPortraitUpdatedTitle`, `toastPortraitChosenBody(_:)`, `toastPortraitEarliestBody` *(new)* | key | `L10n.Sources.coverBadge`, `useAsThumbnail`, `thumbnailMenuTitle`, `thumbnailMenuHint`, `toastThumbnailUpdated*` | Person copy; the Source keys name a Source thumbnail |
| `TestSetPreferredLikeness` *(new)* (Go, two files), `portrait*` Swift tests *(new)* | test | `TestDepictionsReconcileByImage`, `confirmUnpromoteRefreshesToPromote` | — |

Searched for: "portrait", "preferred", "primary_artifact", "SetCover", "cover" in `core/`, `api/`, `macos/App`; writers of `canonical_entities`; `WriteEffects` responses; `pvContextMenuTrigger` and `vocabularyToastOverlay` call sites; free FFI method numbers in `engine.proto` and the Spike 11 plans.

**Builds on.** `catalogmodel` `BucketOptional` (no `deleteimpact` code), the `canonical_entities` effect entry (recompute, header drop, Last updated), `writes.Run`, `writeEffectsProto` and `session.apply`, S10-35's `likenesses.ForEntities` and lead slot, the Source cover UI.
**Deletes.** The keyboard-reset `onChange` in `SourcePageIdentityHeader.coverThumbnail` and the keyboard assignment in `CitationComposerViewerPane.toggleClearMenu` (D-16). S10-35's `[0].ImageFileID` lead expression in `composePersons` (replaced by `likenesses.Lead`), and the "S10-35: the earliest" doc wording on `PersonHeader.LikenessImageFileID` and `likeness_image_rel_path`. S10-D12 leaves `design/` for `design/archive/` (its last PR).
**Next adds.** A place or event portrait is the same column read by `Lead` once those kinds have galleries (S10-35 Next adds). A handle merge (not planned) copies or clears the pointer in the merge transaction.

### Schema

Migration `core/database/migrations/000054.sql` *(new)*:

```sql
-- S10-39: a Person's preferred portrait is a display pointer to one likeness
-- Observation. Not evidence. Deleting the Observation clears it.
ALTER TABLE canonical_entities ADD COLUMN preferred_likeness_observation_id BLOB
	REFERENCES observations(id) ON DELETE SET NULL;
CREATE INDEX canonical_entities_preferred_likeness_observation_id_idx
	ON canonical_entities(preferred_likeness_observation_id);
```

Registered in `catalogmodel.FKs` as `BucketOptional` (D-1). Delete class: optional back-pointer, no probe, no release. Audit: the set and clear are audited as `canonical_entity` updates (D-3); the FK's own `SET NULL` on an Observation delete writes no audit row, as for `sources.primary_artifact_id`.

### Go core

| Package | Change |
| --- | --- |
| `core/database/catalogmodel` | One FK row (D-1) |
| `core/database/likenesses` | `Likeness.Preferred`, `sqlGallery` join and column, `Lead` (D-2); package doc gains "Lead picks the likeness a handle leads with." |
| `core/database/canonicalentities` | `SetPreferredLikeness`, `sqlPreferred`, `nullBlob` (D-3) |
| `core/database/conclusiondetails` | `LikenessTile.Preferred` (D-4) |
| `core/database/conclusionheaders` | `composePersons` lead through `Lead`; `PersonHeader.LikenessImageFileID` doc (D-5) |

```go
// Lead returns the likeness a handle leads with: its preferred portrait when
// that likeness is in the gallery, else the earliest. ok is false when the
// gallery is empty.
func Lead(gallery []Likeness) (l Likeness, ok bool)

// SetPreferredLikeness chooses one of the handle's gallery likenesses as its
// portrait by the likeness's record Observation, or clears the choice when
// observationID is empty. A display pointer, not evidence. An unchanged
// choice returns no changes.
func SetPreferredLikeness(tx *database.Tx, userID, entityID, observationID []byte) ([]rowchange.Change, error)
```

Errors: `canonicalentities.invalid` → bad ids, unknown handle, or an Observation that is not one of this handle's gallery likenesses.

### FFI

`api/proto/engine.proto` per D-6; handler per D-7; regenerate with `./scripts/generate-proto.sh`.

```proto
  METHOD_SET_PREFERRED_LIKENESS = 101;

// SetPreferredLikeness chooses a Person's portrait (S10-39): observation_id is
// one of the Person's gallery likenesses' record Observation; empty clears the
// choice. A display pointer, not evidence. An unchanged choice writes nothing.
message SetPreferredLikenessRequest {
  string project_dir = 1;
  string user_id = 2;
  string entity_id = 3;
  string observation_id = 4;
}

message SetPreferredLikenessResponse {
  WriteEffects effects = 1;
}

// LikenessDetail
  // preferred is true for the likeness chosen as the Person's portrait
  // (S10-39); at most one.
  bool preferred = 3;
```

### macOS

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Use as portrait | `PVButton` ghost, `sm`, no icon | `ConclusionDetailPage.swift` `ConclusionDetailLikenessSection` | PDP-10; D-12. Under the caption, leading |
| Portrait mark | `PVBadge(…, tone: .accent, icon: .photo)` | same | PDP-10; D-12. Replaces the button on the chosen tile |
| Chosen toast | `PVToast` through `vocabularyToastOverlay` | `PersonDetailView.swift` | PDP-10; D-9, D-10. Identifier `persons.detail.toast` |
| Header thumb menu | `PVContextMenu` (`pvContextMenuTrigger`, `pvContextMenu`, `PVContextMenuPanel`, `PVContextMenuItem`) | `ConclusionDetailPage.swift` `ConclusionDetailHeader` | PDP-11; D-13. Only while a portrait is chosen |
| Persons list thumb | unchanged `CachedThumbnail` in `PVList` | `ConclusionListPage.swift` | PSL-8; the header's `likenessImageRelPath` already carries the portrait (D-5) |

Frames in words. **Portrait** (brief frame 10): the Gallery frame (header thumb, “Likenesses 3”, three 96pt tiles each in a 120pt column with its Source caption) with the second tile chosen. Under the second tile's caption, an accent Badge “Portrait” with the photo icon; under the first and third, a ghost small “Use as portrait” button, leading-aligned. A success toast at the top trailing corner of the page: title “Portrait updated”, body “Using the likeness from {Source} as the portrait.” The header thumb shows the second crop (with the gender line from PDP-9) and its menu is open below the pointer: panel title “Person portrait”, one item “Use earliest likeness”. The tiles keep their order. **Persons list, Preferred portrait** (D9 frame 8): the same row as the likeness frame, now with the chosen crop in the 44pt thumb; nothing on the row says it is preferred. **Dark mode**: kit tokens only; no new colour.

States: no likeness → no section, no menu; one likeness → its tile shows the button, and choosing it changes the header only in that the menu now exists; saving → every “Use as portrait” is disabled until the write returns; failure → danger toast, nothing else changes; a chosen likeness that leaves the gallery → no Badge, earliest leads, no menu. Accessibility: each button is spoken “Use the likeness from {Source} as portrait”; the Badge reads “Portrait” inside the section container; the header thumb keeps “Likeness from {Source}” and, while a portrait is chosen, the hint “Right-click for portrait options” and the action “Use earliest likeness”.

### Strings

L10n keys *(new)* in `L10n.Conclusions`, with `Localizable.xcstrings` entries:

- `portraitBadge` = `conclusions.person.portraitBadge` "Portrait" — Badge on the gallery likeness chosen as the Person's portrait
- `useAsPortrait` = `conclusions.person.useAsPortrait` "Use as portrait" — button under a gallery likeness that makes it the Person's portrait
- `useAsPortraitLabel(_ source: String) -> String` = `conclusions.person.useAsPortraitLabel` "Use the likeness from %@ as portrait" — accessibility label of that button; the argument is the Source title (or the Observation ref when the title is empty)
- `useEarliestLikeness` = `conclusions.person.useEarliestLikeness` "Use earliest likeness" — header thumb menu item and accessibility action that clears the portrait
- `portraitMenuTitle` = `conclusions.person.portraitMenuTitle` "Person portrait" — title of the header thumb menu
- `portraitMenuHint` = `conclusions.person.portraitMenuHint` "Right-click for portrait options" — help and accessibility hint on the header thumb while a portrait is chosen
- `toastPortraitUpdatedTitle` = `conclusions.person.toastPortraitUpdatedTitle` "Portrait updated" — toast title after choosing or clearing the portrait (a `String`, like `L10n.Sources.toastThumbnailUpdatedTitle`)
- `toastPortraitChosenBody(_ source: String) -> String` = `conclusions.person.toastPortraitChosenBody` "Using the likeness from %@ as the portrait." — toast body after choosing; the argument is the tile's Source title
- `toastPortraitEarliestBody` = `conclusions.person.toastPortraitEarliestBody` "Using the earliest likeness as the portrait." — toast body after clearing

Removed: none.

## 7. Steps

**STEP-1 — Engine: the column**
Files: `core/database/migrations/000054.sql` *(new)*, `core/database/catalogmodel/model.go`. Skills: `add-catalog-migration`, `add-catalog-model`.
Do: D-1.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/ ./core/database/catalogmodel/ ./core/database/deleteimpact/` passes (schema hash, `TestPragmaHonesty`, honesty tests).

**STEP-2 — Engine: the gallery knows the portrait**
Files: `core/database/likenesses/likenesses.go`, `core/database/likenesses/likenesses_test.go` (both). Skill: `add-catalog-query`.
Do: D-2. Add the read cases of T-1; they set the pointer with raw `UPDATE` SQL until STEP-3 exists, then STEP-3 switches them to `SetPreferredLikeness`.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/likenesses/ ./core/database/autoreconciler/` passes.

**STEP-3 — Engine: choose and clear**
Files: `core/database/canonicalentities/canonicalentities.go`, `core/database/likenesses/likenesses_test.go`. Skills: `add-catalog-query`, `add-audit-scope`.
Do: D-3, D-14. Add T-2; switch T-1's pointer setup to `writes.Call` + `SetPreferredLikeness`; T-2 runs through `writes.Run`.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/canonicalentities/ ./core/database/likenesses/ ./core/database/audit/ ./core/database/effects/` passes.

**STEP-4 — Engine: detail and header lead with the portrait**
Files: `core/database/conclusiondetails/conclusiondetails.go`, `core/database/conclusiondetails/conclusiondetails_test.go`, `core/database/conclusionheaders/conclusionheaders.go`, `core/database/conclusionheaders/compose.go`, `core/database/conclusionheaders/conclusionheaders_test.go`.
Do: D-4, D-5. Add T-3, T-4.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/conclusiondetails/ ./core/database/conclusionheaders/ ./core/database/graphcache/` passes, including the query-count tests.

**STEP-5 — FFI**
Files: `api/proto/engine.proto`, generated files from `./scripts/generate-proto.sh`, `api/ffi/dispatch.go`, `api/ffi/handlers/conclusion_detail.go`, `api/ffi/handlers/conclusion_detail_test.go`, `api/libprovenencia/call_test.go`. Skills: `add-ffi-handler`, `use-catalog-session`.
Do: D-6, D-7. Add T-5; move the unknown-method case to 102.
Check: `./scripts/generate-proto.sh` leaves no further diff on a second run; `CGO_ENABLED=1 go test -tags fts5 ./api/...` passes.

**STEP-6 — Swift store**
Files: `macos/App/Platform/GenealogyStore.swift`, `macos/App/Platform/GoStore.swift`, `macos/App/Platform/CoreInvoke.swift`, `macos/App/Platform/FakeStore.swift`.
Do: D-8.
Check: `xcodebuild test -project macos/Provenencia.xcodeproj -scheme Provenencia -destination 'platform=macOS' -only-testing:ProvenenciaTests CODE_SIGNING_ALLOWED=NO` (written `xcodebuild test …` below) builds and passes.

**STEP-7 — Portrait model**
Files: `macos/App/Features/Conclusions/ConclusionPortraitModel.swift` *(new)*, `macos/Provenencia.xcodeproj/project.pbxproj`, `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`, `macos/ProvenenciaTests/PersonDetailTests.swift`. Skills: `add-localized-string`, `add-swift-test`.
Do: D-9; keys `toastPortraitUpdatedTitle`, `toastPortraitChosenBody`, `toastPortraitEarliestBody`. Add T-7.
Check: `xcodebuild test …` passes; `python3 scripts/check-localizable-xcstrings.py` passes.

**STEP-8 — Kit: the menu resets its keyboard on open**
Files: `macos/App/DesignSystem/Components/ContextMenu/PVContextMenu.swift`, `macos/App/Features/Sources/SourcePageIdentityHeader.swift`, `macos/App/Features/CitationComposer/CitationComposerViewerPane.swift`, `macos/ProvenenciaTests/PVContextMenuTests.swift`. Skill: `add-ui-component`.
Do: D-16. Add T-8.
Check: `xcodebuild test …` passes; `rg -n "PVContextMenuKeyboard\(" macos/App` lists only `PVContextMenu.swift` and `PVSelect.swift`.

**STEP-9 — Tile controls and header menu**
Files: `macos/App/Features/Conclusions/ConclusionDetailPage.swift`, `PersonDetailView.swift`, `PersonDetailContent.swift`, `macos/App/Features/Workspace/WorkspaceDestinationHost.swift`, `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`, `macos/ProvenenciaTests/PersonDetailTests.swift`. Skills: `add-ui-component`, `add-localized-string`, `evaluate-ui-component` (D-15).
Do: D-10 to D-13, D-15 (the menu uses D-16's `keyboardTitles`); keys `portraitBadge`, `useAsPortrait`, `useAsPortraitLabel`, `useEarliestLikeness`, `portraitMenuTitle`, `portraitMenuHint`. Add T-6.
Check: `xcodebuild test …` passes; `python3 scripts/check-localizable-xcstrings.py` passes; the Person page previews build.

**STEP-10 — Docs**
Files: `docs/conclusion-layer-data-model.md` §6, `docs/deployment-plan/spike-10/completed.md`, `docs/deployment-plan/spike-10/deployment-plan.md` (S10-39 scope entry status, Status paragraph), this plan (status, Landed note), `docs/deployment-plan/spike-10/design/S10-D12-person-detail.md` → `docs/deployment-plan/spike-10/design/archive/S10-D12-person-detail.md` *(new)*, `docs/deployment-plan/spike-10/design/README.md`, and every link to the brief's old path (`rg -l "design/S10-D12-person-detail.md" docs .agents`), landed plans included (`S10-23-person-gender.md`, `S10-35-person-gallery.md`: link fixes only). The moved brief's own relative links each gain one `../`.
Do: per the Definition of done. In §6 of the domain doc, add `preferred_likeness_observation_id BLOB REFERENCES observations(id) ON DELETE SET NULL` to the DDL and one paragraph: "`preferred_likeness_observation_id` is a display pointer for a Person: the likeness Observation the Person page and the Persons list lead with. It is not evidence and not a Reconciliation Claim. A reader honours it only when that likeness is in the handle's gallery; otherwise the earliest likeness leads. Deleting the Observation clears it."
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-10/prs/S10-39-preferred-portrait.md` reports no errors; `rg "design/S10-D12-person-detail.md" docs .agents` finds no link outside `design/archive/`; `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-10/design/archive/S10-D12-person-detail.md docs/deployment-plan/spike-10/prs/S10-23-person-gender.md docs/deployment-plan/spike-10/prs/S10-35-person-gallery.md` reports no broken link.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/database/likenesses/likenesses_test.go` `TestForEntitiesPreferred` *(new)* | Table, each case on a fresh `newFixture` with likeness A (shade 100, *Register*) then B (shade 200, *Census*) saved by `observationstest.SaveLikeness` on two Subjects promoted onto one Person. No pointer → no `Preferred`, `Lead` A. Pointer B → order [A, B], only B `Preferred`, `Lead` B. Pointer B, then B's Subject unpromoted (`promote.Unpromote` through `writes.Call`) → gallery [A], `Lead` A, pointer row unchanged. Pointer B, then `writes.Call` with `observations.Delete(tx, userID, B's record id)` → `SELECT preferred_likeness_observation_id` is NULL, `Lead` A. Empty gallery → `Lead` returns `ok == false` | AC-1.2, AC-3.1, AC-3.2, D-2 |
| T-2 | Go | `likenesses_test.go` `TestSetPreferredLikeness` *(new)* | Table over `canonicalentities.SetPreferredLikeness` through `writes.Run`: B's record → one change, `Fields["preferred_likeness_observation_id"]` Old nil, New B's UUID; result `Effects.Handles` and `Effects.HandleScopes` hold the handle. B again → no changes, revision 0. A → pointer A. Empty → pointer NULL, change New nil. A second promoted Person's likeness (`observationstest.SaveLikeness` on another Subject, shade 150) → `ErrInvalid`, pointer unchanged. A fresh UUIDv7 that names no Observation → `ErrInvalid`. Unknown entity id → `ErrInvalid`. Nil user → `ErrInvalid` | AC-1.2, AC-1.4, AC-2.2, AC-3.3, AC-3.4, D-3 |
| T-3 | Go | `core/database/conclusionheaders/conclusionheaders_test.go` `TestPersonHeaderLeadsWithPortrait` *(new)* | Person with A then B (`observationstest.SaveLikeness`, as `TestPersonHeaderLeadsWithEarliestLikeness`); set pointer B through `writes.Call` + `canonicalentities.SetPreferredLikeness` (the commit drops the stored row; no `Graph().Drop()`) → `ListPersons` and `PersonsByIDs` `LikenessImageFileID` = B's image; `c.Graph().Verify()` → nil; clear → A's image | AC-1.3, AC-2.2, D-5 |
| T-4 | Go | `core/database/conclusiondetails/conclusiondetails_test.go` `TestForEntityLikenesses` (extended) | After the existing case, set pointer B through `writes.Call` + `canonicalentities.SetPreferredLikeness` → `Likenesses` [A, B], `Likenesses[1].Preferred` true, `[0].Preferred` false | AC-1.2, D-4 |
| T-5 | Go | `api/ffi/handlers/conclusion_detail_test.go` `TestSetPreferredLikeness` *(new)* | `likenessFixture`, `saveLikeness` A (`writeCrop(t, 300, 200, 100)`) then B (`writeCrop(t, 300, 200, 200)`), `PromoteSubject` on the person (empty `entity_id`); the record ids are the `observations[0].id` of each response. Set B → `effects.handles` and `effects.handle_scopes` contain the entity; `GetConclusionDetail` → `likenesses[1].preferred`, `person.likeness_image_rel_path == likenesses[1].image_rel_path`. Set B again → `effects.revision == 0`. Clear (empty `observation_id`) → lead is `likenesses[0]`. Set to a fresh UUIDv7 string → error code `canonicalentities.invalid`. Set B, `DeleteObservation` B → one likeness, lead A | AC-1.2, AC-1.3, AC-2.2, AC-3.1, AC-3.3, AC-3.4, D-6, D-7 |
| T-6 | Swift | `macos/ProvenenciaTests/PersonDetailTests.swift` `portraitFlagsFollowThePreferredLikeness` *(new)* | Detail with likenesses [A (*Register*), B (*Census*, `preferred`)] and header path B → `content.likenesses.map(\.isPortrait)` [false, true]; `thumbnail.isPortrait` true; `thumbnail.imageLabel` "Likeness from Census". Neither preferred → both false, `thumbnail.isPortrait` false. `EventDetailContent(...).thumbnail.isPortrait` false | AC-1.1, AC-1.3, AC-2.3, D-11 |
| T-7 | Swift | `PersonDetailTests.swift` `choosingAPortraitWritesAndToasts`, `useEarliestClearsAndToasts`, `portraitFailureToastsTheError` *(new)* | `FakeStore` + `WorkspaceSession(projectKey:store:)` + `ConclusionPortraitModel(session:userID: "user-1", entityId: "per-1")`. `choose(B)` → last `recordedCalls` "setPreferredLikeness entity=per-1 observation=\(B.id)"; `toast` success, title "Portrait updated", body "Using the likeness from Census as the portrait."; `isSaving` false. `useEarliest()` → "… observation=" and body "Using the earliest likeness as the portrait.". With `setPreferredLikenessError` set → toast tone `.danger`, title `L10n.Errors.message(for:)` of that error, body "" | AC-1.2, AC-1.5, AC-2.2, D-8, D-9 |
| T-8 | Swift | `macos/ProvenenciaTests/PVContextMenuTests.swift` `openedHighlightsNothingAndTypeSelectsTitles` *(new)* | `PVContextMenuKeyboard.opened(titles: ["Clear region", "Reset to entire artifact"])` → `itemCount` 2, `activeIndex` -1, `itemTitles` those two; `applyTypeSelect("r")` → `activeIndex` 1. `opened(titles: [])` → `itemCount` 0, `applyTypeSelect("r")` false | AC-2.1, D-16 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `catalogmodel.TestPragmaHonesty` and the schema hash (STEP-1), `TestListPersonsQueryCountIsConstant` and `TestForEntityQueryCountIsConstant` (the query count must not change: the pointer rides the existing gallery query), `api/libprovenencia` `TestProvenenciaCall` (method 101 is taken), every `GenealogyStore` conformer (`GoStore`, `FakeStore`).

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together) — or none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated (if gated) — S10-D12 moves now (its last PR); links to the old path repointed
- [ ] Domain doc updated if a rule in docs/*.md changed: conclusion-layer-data-model.md §6 (the column and its paragraph)
- [ ] PR description: what shipped, AC evidence, deviations, manual check, and the evaluate-ui-component verdict (D-15)
```

## 10. Manual check

Run on a Mac, on a test project.

1. In the S10-34 composer, give *Ann* two likenesses on two Sources (*Register* first, then *Census*); promote both onto one Person. Give her gender Woman.
2. Open *Ann*: the header shows the *Register* crop. Both tiles show “Use as portrait”; right-clicking the header thumb opens nothing (AC-1.1, AC-2.3).
3. Click “Use as portrait” under *Census*: the toast reads “Portrait updated” / “Using the likeness from Census as the portrait.”; *Census* shows the “Portrait” Badge and no button; the tiles stay in order; the header shows the *Census* crop with the plum outline. Open the Persons list: *Ann*'s thumb is the *Census* crop and the row shows nothing else new (AC-1.2, AC-1.3).
4. Back on *Ann*, right-click the header thumb: “Person portrait” / “Use earliest likeness”. Choose it: the *Register* crop leads, both tiles show their buttons, the toast says the earliest is used (AC-2.1, AC-2.2). With VoiceOver, after choosing *Census* again, the header thumb offers “Use earliest likeness” (AC-2.4).
5. With *Census* chosen, open its Citation from the tile and delete that likeness Observation. Back on *Ann*: one tile, *Register* leads, no Badge, no menu (AC-3.1).
6. Regression for D-16: on a Source page with a cover, right-click the cover, press R: “Revert to default” highlights; Enter reverts. In the composer, open the clear-region menu twice: nothing is highlighted on each open; ↓ highlights “Clear region”. On *Ann* with a portrait, right-click the header thumb, press ↓ then Enter: the earliest likeness leads (AC-2.1).
7. Choose the remaining likeness, then unpromote its Subject on the Evidence graph: *Ann* has no gallery and the person mark (AC-3.2).

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| A pointer write leaves a stale stored Person header | The `canonical_entities` effect puts the handle in `set.Handles`; `OnCommit` drops its row; T-3 runs `Verify` |
| The pointer outlives its likeness's place in the gallery | `Lead` only honours a gallery likeness (D-2); T-1 covers unpromote and delete |
| The `Use as portrait` button overflows the 120pt column at larger text sizes | No icon (D-12); §10 step 3 checks it at the default size |
| The open-time reset changes how the Source cover or the composer menu behaves | Same values they set by hand, plus titles; `PVSelect` passes no titles and is untouched; §10 step 6 |
| Recomputing the handle on every pointer write | One handle per click, the same cost as a label edit |

**Stop and report if:**

- `000054.sql` is taken by another PR when you start STEP-1.
- `METHOD_SET_PREFERRED_LIKENESS = 101` collides with a method another PR added.
- Adding the `canonical_entities` join changes the result of S10-35's `TestForEntities` or `TestDepictionsReconcileByImage` cases (a handle row missing for a cached value).
- `TestListPersonsQueryCountIsConstant` or `TestForEntityQueryCountIsConstant` needs a higher count.
- `TestResolversCoverEveryEntityType`, `audit.Record`, or `effects.Resolve` rejects a `canonical_entity` `ActionUpdate` change.
- The board's Portrait frame puts the Badge, the button, or the menu somewhere §6 does not say, or with other copy (report the difference; do not restyle).

## 12. Open questions

None.

## 13. PR

**Title:** S10-39 — Preferred portrait
**Body:** objective in one sentence; what shipped against US-1 to US-3; AC evidence; deviations from the plan; the `evaluate-ui-component` verdict; manual check; `Plan: docs/deployment-plan/spike-10/prs/S10-39-preferred-portrait.md`.
