# S15-03 — The Source page shows the date and the Private switch

**Status:** Blocked — on board S15-D1
**Spike / slice:** Spike 15 · Slice 1 — Source date, privacy, credibility
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** [S15-D1](../design/S15-D1-source-page.md), board not drawn yet
**Board:** — (ask the user how the agreed board reaches the implementer when it is agreed; until then **Design → macOS** describes each frame in words from the brief)
**Depends on:** S15-02 (landed [#762](https://github.com/mendahu/provenencia/pull/762)), S15-D1
**Blocks:** S15-09, S15-23, S15-25 (they build on the page layout this PR adds)
**Branch:** `s15-03-source-page-date-and-private-switch`
**Version bump:** minor, 0.5.0 → 0.6.0 — new Source page controls and a new error code ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-ui-component`](../../../../.agents/skills/add-ui-component/SKILL.md), [`add-localized-string`](../../../../.agents/skills/add-localized-string/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md), [`add-date-value`](../../../../.agents/skills/add-date-value/SKILL.md), [`add-ffi-handler`](../../../../.agents/skills/add-ffi-handler/SKILL.md), [`add-catalog-query`](../../../../.agents/skills/add-catalog-query/SKILL.md)
**Rules that bind:** `.cursor/rules/macos-client.mdc` (Go wire vocabulary as a typed enum), `.cursor/rules/go-tests.mdc`, `.cursor/rules/ffi-handlers.mdc`, `.cursor/rules/catalog-queries.mdc`, `.cursor/rules/versioning.mdc`
**Context:** working rules "The Source date is one optional DateValue on `sources`", "Not public is a stored switch and nothing else" ([`../deployment-plan.md`](../deployment-plan.md#working-rules)); [`../source-provenance-date.md`](../source-provenance-date.md); [privacy note](../2026-10-27-source-privacy-toggle.md); S15-02's [plan](S15-02-source-date-and-public-flag.md) and [ship log](S15-02-source-date-and-public-flag.ship.md) (UX-1, UX-2); [`macos-client-patterns.md` § 9](../../../macos-client-patterns.md#9-window-widths-regular-and-wide); [`../../../ideas/wide-layout.md`](../../../ideas/wide-layout.md) "Source page"
**Planned:** 2026-11-09 against `main` @ `f439a14`
**Review:** [`S15-03-source-page-date-and-private-switch.review.md`](S15-03-source-page-date-and-private-switch.review.md) — round 1: Fits with changes / Not ready → all findings Fixed; round 2 runs when the board is agreed

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

On the Source page, a researcher sets, changes and clears the Source's own date (when the record was made or first published) from its header, with the app's structured date form in a dialog, and marks the Source private with a switch that saves at once. The page moves onto Page columns, so at Wide its Metadata board sits in a side column instead of leaving most of the pane empty. A write to, or a visit of, a Source deleted elsewhere now says so and goes back to the Sources list instead of showing "Invalid source." This is the Source page half of dogfood bars 1 and 2.

**Why now.** S15-02 stores both facts and nothing shows them; S15-09, S15-23 and S15-25 add to the page layout this PR sets.

**What stays.** Every section's content and controls, and the Regular arrangement (Description and Credibility beside Metadata, then graphs, Artifacts, Notes). Title, type and description keep their shared `updateSource` write. The Sources list (S15-04) and Copy Source (S15-32) are untouched. The Private flag changes nothing else in the app (working rule).

## 2. User stories

**US-1.** As a researcher, I give a Source its date, change it and clear it from the Source page, so the record carries when it was made.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | An undated Source; the header | Shows "Add date" in the date's place (SD-2) |
| AC-1.2 | Add date → the dialog; kind Point, year 1880; Save | The dialog closes; the header shows "1880"; the store's Source has `date.kind == "point"`, `startYear == 1880` |
| AC-1.3 | That Source; the pencil → the dialog | The form is prefilled with point 1880 (`DateValueDraft(from:)`); Clear date is offered |
| AC-1.4 | Change to About, March, 12; Save | The header shows `DateRowDisplay.line(date:)`'s "abt 12 Mar 1880" |
| AC-1.5 | The pencil → Clear date | The form empties and a caption reads "Saving removes this Source's date."; Save is enabled; Save → the store's Source has no date, the header shows "Add date" |
| AC-1.6 | The dialog opened from Add date, blank form | Save is disabled |
| AC-1.7 | Save fails with `datevalues.invalid` | The dialog stays open with its draft; a danger Callout shows the shared `error.datevalues.invalid` copy, now "This date isn't valid. Check the year, month and day." (UX-2) |
| AC-1.8 | The dialog is dirty; Cancel | The FormDialog asks before discarding, as every FormDialog |

**US-2.** As a researcher, I mark a Source private, so later sharing can leave it out.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | A Source with `isPublic: true` | The Private switch is off |
| AC-2.2 | Flip it on | While the write runs the switch shows on and is disabled; after it the store's Source has `isPublic == false`; reopening the page shows it on |
| AC-2.3 | Flip it on; the write fails with `internal.unknown` | The switch is back off; the page's error Callout shows the error message |
| AC-2.4 | The description is in edit with unsaved text; flip Private | The description edit stays open with its text |

**US-3.** As a researcher, when another window has deleted this Source, the page tells me and takes me back to the Sources list (UX-1).

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | Go: `sources.SetDate`, `SetPublic`, `SetCover`, `Update`, `Delete`, `Get`, `GetTx`, and `sourcecopy.ReadDraft` on an unused 16-byte id | `sources.ErrNotFound` (code `sources.not_found`, `KindNotFound`) |
| AC-3.2 | Go: the `GetSourceWorkspace` and `SetSourceDate` RPCs on an unused id | error code `sources.not_found` |
| AC-3.3 | Go: a 3-byte id or a missing user | Still `sources.ErrInvalid` |
| AC-3.4 | Swift: any Source page section write (date, switch, title, type, description, cover, note, metadata, credibility, Artifact) throws `sources.not_found` | The page leaves for the Sources list; `session.noticeToast` reads "“1880 census” was deleted." with the title last shown |
| AC-3.5 | Swift: the page's workspace read fails with `sources.not_found` | The page leaves (as today) and `session.noticeToast` reads `error.sources.not_found`: "This Source no longer exists." |
| AC-3.6 | Swift: the Person page's read fails with `conclusiondetails.not_found` | Unchanged: it leaves with its notice (moved onto the shared modifier) |

**US-4.** As a researcher on a wide monitor, the Source page uses the width.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-4.1 | Pane 2296pt | The header pinned in Page columns' band; a main column at `widthContentMax` holding Description, Credibility, graphs, Artifacts, Notes; Metadata in a side column beside it (SD-W1) |
| AC-4.2 | Pane 1102pt | As today: Description and Credibility beside Metadata, Credibility at the bottom of its column, then graphs, Artifacts, Notes (SD-W2) |
| AC-4.3 | Resize across 1600pt with the date dialog open, a metadata value in edit, and the description in edit | The dialog stays open with its draft; both edits keep their text; the scroll position is kept |

## 3. Scope

| In | Covered by |
| --- | --- |
| SD-1, SD-2: the date and Add date in the header's meta row | STEP-6, AC-1.1, AC-1.4 |
| SD-3, SD-4, SD-5: the date dialog, Clear date inside it, the refused-date Callout and the shared copy fix (UX-2) | STEP-2, STEP-4, STEP-6, AC-1.2 – AC-1.8 |
| SD-6: the Private switch | STEP-4, STEP-6, AC-2.1 – AC-2.4 |
| SD-7: a Source deleted elsewhere — Go `sources.ErrNotFound` from one place (UX-1); Swift `EngineErrorCode.sourceNotFound`, `SourcePageContext.fail`, the shared leave-with-notice | STEP-1, STEP-2, STEP-5, AC-3.1 – AC-3.6 |
| SD-8: tab order, VoiceOver, focus back after the dialog | STEP-6, manual check 5 |
| SD-W1, SD-W2: the page on `PVPageColumns` with a feature layout | STEP-7, AC-4.1 – AC-4.3 |
| `CatalogSource.date`, `.isPublic`; `setSourceDate`, `setSourcePublic` on `GenealogyStore`, `GoStore`, `FakeStore` | STEP-3 |

| Out | Home |
| --- | --- |
| The date on Sources list rows and the date sort; `CatalogSource.dateSortKey` | S15-04 |
| An Artifact's metadata board (SD-9, SD-10) | S15-09 |
| The citation preview (SD-11 – SD-13) | S15-23 |
| Move on Artifact and graph cards (SD-14) | S15-25 |
| Copy Source carrying the date | S15-32 |
| Any use of the Private flag (badge, filter, sharing) | not planned in this spike (working rule) |
| Not found from the note, metadata, credibility and Artifact packages and from the Copy and graph form models (`SourceCopyModel`, `SourceGraphFormModel`) when their Source is gone | [`docs/ideas/structural-debt.md`](../../../ideas/structural-debt.md) gains a line (STEP-8) |
| Moving the shared date form's `L10n.Sources.date*` keys to `L10n.Dates` | [`docs/ideas/structural-debt.md`](../../../ideas/structural-debt.md) gains a line (STEP-8) |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Header | `macos/App/Features/Sources/SourcePageIdentityHeader.swift` `SourcePageIdentityHeader`, `metaRow` | Cover, title inline edit, meta row: ref · type pill + pencil; Copy and Delete. Draws its own `widthContentMax` frame, `gutterPage` padding, `surfaceCard` background and `PVDivider`; renders nothing inside when `model.workspace == nil` |
| Page shell | `macos/App/Features/Sources/SourcePageView.swift` `SourcePageContent`, `overviewColumns`, `SourcePageOverviewHeightKey`, `SourcePageLoadingShell` | `VStack(header, ScrollView { … })` with its own content-width frame, left-aligned. `overviewColumns` pins Credibility to the bottom through a PreferenceKey height loop. The loading shell draws the header and its own ScrollView. `.leavesForSectionRoot(when: workspaceHandle.read.isFailed)` leaves silently |
| Section state | `SourcePageDescriptionView.descriptionDraft`, `SourcePageCredibilityView.argumentLocal`, `SourcePageMetadataView` drafts (`@State`) | Local view state: lost if the view changes parent |
| Header model | `macos/App/Features/Sources/SourceIdentitySection.swift` `SourceIdentitySection`, `save` | Title, type, description drafts sharing one `updateSource` save; errors through `L10n.Errors.message(for:)` |
| Page model | `macos/App/Features/Sources/SourcePageModel.swift` `apply`, `sync` | `apply` runs `identity.reset()` whenever `previous?.source != workspace.source`, so any `CatalogSource` change (cover, and with this PR date and flag) drops title, type and description edits |
| Page context | `macos/App/Features/Sources/SourcePageContext.swift` `applySource`, `setSourceCover`, `refreshCoverFromStore` | Write failures set `pageError = L10n.Errors.message(for: error)`; 16 such sites across `SourceArtifactsSection` (4), `SourceMetadataSection` (5), `SourceNotesSection` (3), `SourceCredibilitySection` (1), `SourceIdentitySection` (1), `SourcePageContext` (2) |
| Dialog pattern | `SourceGraphFormDialog.swift` `sourceGraphFormDialog(_:)` + `SourceGraphFormModel`; `SourceCopyDialog.swift` + `SourceCopyModel`; `PVFormDialog.swift` `pvFormDialog(item:…onConfirm: (Model) async -> Bool)` | Form model + one-line modifier |
| Date form and display | `macos/App/Features/Dates/DateValueEditorForm.swift`; `macos/App/Platform/DateValueDraft.swift` (`isBlank`, `isValid`, `toInput()`, `init(from:)`); `macos/App/Platform/DateRowDisplay.swift` `DateRowDisplay.line(date:)` ("abt 12 Mar 1880", "bet 1834 and 1840"; the page and list form) | `DateValueDisplay` is the editors' formatter, not used here |
| Switch | `macos/App/DesignSystem/Components/Switch/PVSwitch.swift` `PVSwitch` | No Features caller |
| Page columns | `macos/App/DesignSystem/Components/PageColumns/PVPageColumns.swift` | Draws the pinned header's rule and band, no background. `hasSecondary` comes from the type, so an empty secondary still stacks at Regular. `primaryFill` lets the primary fill the pane at Wide. `PVPageColumnsMetrics.split(available:sideMax:)` sizes a main column beside a side column. `PVPageColumnsLayout` keeps subview identity across widths |
| Side column precedent | `macos/App/Features/Conclusions/ConclusionPinnedSideColumn.swift`; `PersonSourcesSection.swift` (`PVPageColumns(primaryFill: true)` + side column) | Main and side over the same two subviews, so crossing 1600 keeps identity |
| Gone precedent | `macos/App/Features/Conclusions/ConclusionDetailPage.swift` (`.onChange(of: presentation == .gone, initial: true)` sets `session.noticeToast`, then `.leavesForSectionRoot(when:)`); `macos/App/Features/Workspace/WorkspaceNavigation.swift` `LeavesForSectionRoot`; `WorkspaceSession.noticeToast` | The notice is set at each call site |
| Typed codes | `macos/App/Platform/ConclusionVocabulary.swift` `EngineErrorCode`; `macos/App/Platform/CoreInvoke.swift` `CoreInvokeError.engineCode`; `api/ffi/swift_vocabulary_test.go` row "engine error codes" | `.conclusionNotFound`, `.promoteStale`, `.sourceMetadataInvalid` |
| Swift model | `macos/App/Platform/GenealogyStore.swift` `CatalogSource`; `GoStore.swift` `mapSource`, `setSourceCover`, `mapDateValueToProto`, `mapDateValue`; `FakeStore.swift` `updateSource`, `setSourceCover`, `getSourceWorkspace`, `updateSourceError`; `CoreInvoke.swift` `CoreMethod.setSourceCover` | `CatalogSource` has no date or flag; no store method for the two RPCs. Generated `Provenencia_Engine_V1_Source` has `date`/`hasDate`, `isPublic`; `Method.setSourceDate` (117), `.setSourcePublic` (118). FakeStore throws `StoreBoom.boom` for an unknown id on writes and returns a blank Source on read |
| Go | `core/database/sources/sources.go` `getOne`, `getTx`, `GetTx`, `Get`, `getByRef`, `loadForWrite`, `Delete` (`deleteimpact.Codes{NotFound: ErrInvalid}`); `core/database/sourcecopy/sourcecopy.go` `ReadDraft`; `core/database/sourcemetadata/sourcemetadata.go` `ListWorkspace` | `getOne` returns raw `sql.ErrNoRows` (internal on the wire); `loadForWrite`, `ReadDraft` and `ListWorkspace` each turn it into `ErrInvalid` (`sources.invalid`) |
| Error copy | `core/apperr/apperr.go`; `macos/App/Platform/L10n.swift` `L10n.Errors.dateValuesInvalid` ("Invalid date value."), `message(for:)` | No `sources.not_found` |
| Tests | `core/database/sources/sources_test.go` (missing id ~836, `getByRef` ~521); `api/ffi/handlers/sources_test.go` (`DeleteSource` missing ~792, `TestSetSourceDate` "unknown source" ~1089, `TestSetSourcePublic`, `TestGetSourceWorkspace`); `macos/ProvenenciaTests/SourcePageModelTests.swift`; `PersonDetailTests.swift` `presentationFollowsTheHandle` | Missing-Source cases expect `ErrInvalid` |

**Precedent.** `SourceGraphFormModel` + `sourceGraphFormDialog` for the date dialog; `SourcePageContext.setSourceCover` for a one-attribute Source write; `ConclusionPinnedSideColumn` over `PVPageColumnsLayout` for a layout that keeps identity across widths; `ConclusionDetailPage`'s gone notice for US-3; `EngineErrorCode.conclusionNotFound` for the typed code.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | The date is edited in a FormDialog opened from the header (user, 2026-11-09) | The shared form is tall | Inline expanding editor |
| D-2 | The switch is labeled "Private", off by default; on writes `isPublic: false` (user, 2026-11-09) | The user's wording; the stored flag stays `is_public` | "Public", on by default |
| D-3 | Go: `apperr.CodeSourcesNotFound = "sources.not_found"` and `sources.ErrNotFound = apperr.New(…, apperr.KindNotFound)` *(new)*. `getOne` maps `sql.ErrNoRows` to `ErrNotFound`, once, so `Get`, `GetTx`, `getTx` and `getByRef` all return it. `loadForWrite` passes it through (its `ErrNoRows` branch goes). `Delete`'s `deleteimpact.Codes{NotFound: …}` becomes `ErrNotFound`. `sourcecopy.ReadDraft` and `sourcemetadata.ListWorkspace` drop their `ErrNoRows` → `ErrInvalid` branches and pass `ErrNotFound` through. Malformed ids and users stay `ErrInvalid` (user, 2026-11-09: "not-found code, then leave") | One contract for the package; S15-24 and S15-32 inherit it | Mapping in `loadForWrite` and `Get` only — `GetTx` and `ReadDraft` would still answer `sources.invalid` |
| D-4 | Swift: `EngineErrorCode.sourceNotFound` *(new)* for `sources.not_found`, with `apperr.CodeSourcesNotFound` added to the "engine error codes" row of `api/ffi/swift_vocabulary_test.go`. Code compares `error.engineCode == .sourceNotFound`, never the string | `macos-client.mdc`: wire vocabulary is one typed enum | A raw string compare |
| D-5 | `SourcePageContext.fail(_ error: Error) -> String?` *(new)*: for `.sourceNotFound` it sets `gone = L10n.Sources.sourceDeletedNotice(title: source?.title ?? "")` and returns nil; otherwise it returns `L10n.Errors.message(for: error)`. All 16 page-section catch sites in §4 use it (`pageError = context.fail(error)`, field errors likewise); `SourceIdentitySection.save` returns `context.fail(error)` so a leaving page shows no field error | One door for every Source page write; later packages that return `sources.not_found` leave for free | An opt-in call at a few sites |
| D-6 | `SourcePageContext.gone: String?` *(new, observable)*, exposed as `SourcePageModel.gone`. `SourcePageModel.sync` sets `gone = L10n.string(L10n.Errors.sourcesNotFound)` when `handle.read` is `.failed(let error)` with `error.engineCode == .sourceNotFound` | Write path and read path feed one value | Two toast paths |
| D-7 | `leavesForSectionRoot(when gone: Bool, notice: String? = nil)` *(extended)* in `WorkspaceNavigation.swift`: on `gone` it sets `session.noticeToast = VocabularyToast(title: notice, body: "", tone: .info)` when `notice` is non-nil, then leaves; `initial: true` as today. `ConclusionDetailPage` moves onto it (`notice: L10n.string(L10n.Errors.conclusionDetailsNotFound)` when gone) and drops its own `.onChange`. The Source page: `.leavesForSectionRoot(when: workspaceHandle.read.isFailed || model.gone != nil, notice: model.gone)` | One "say it's gone and leave" | Each page sets the toast itself |
| D-8 | `SourceDateFormModel` *(new)* in `macos/App/Features/Sources/SourceDateFormModel.swift`, shaped as `SourceGraphFormModel`: `draft: DateValueDraft`, `initial: CatalogDateValueInput?`, `clears: Bool`, `isSaving`, `saveError: String?`, `isDirty`, `canSave` (`clears || (!draft.isBlank && draft.isValid)`), `clear()` (empties the draft, `clears = true`), `save() async -> Bool`. A one-line modifier `.sourceDateDialog($model.identity.dateForm)` *(new)* in `SourceDateDialog.swift` *(new)* over `pvFormDialog(item:)`. `SourceIdentitySection` gains `dateForm: SourceDateFormModel?`, `openDate()` | Third dialog in the page's pattern | State spread across `SourceIdentitySection` |
| D-9 | Clear date empties the form inside the dialog; Save then writes no date. No separate Confirm: the dialog's Save is the confirmation, Cancel keeps the date, and the caption says what Save will do | No Confirm sheet over the FormDialog sheet (one sheet per host) | A Confirm over the dialog |
| D-10 | Each write calls the store, then `context.applySource(updated, effects:)`; the header reads `context.source?.date` and `.isPublic` | The RPCs return the Source and effects, as `setSourceCover` | A workspace re-read |
| D-11 | The switch binds to `Binding(get: { identity.pendingPrivate ?? !(source.isPublic) }, set: { v in Task { await identity.setPrivate(v) } })`; `pendingPrivate: Bool?` *(new)* is set while the write runs and cleared after; disabled while non-nil. On failure the binding reads the saved value again and `pageError = context.fail(error)` | No local copy to drift | Optimistic `@State` |
| D-12 | `SourcePageModel.apply` resets `identity` only when `id`, `title`, `sourceTypeID` or `description` differ, through `SourceIdentitySection.resetIfChanged(from: CatalogSource?, to: CatalogSource)` *(new)* | Flipping Private keeps a description edit (AC-2.4) | Exempting the date dialog alone |
| D-13 | `CatalogSource` gains `date: CatalogDateValueInput?` and `isPublic: Bool = true`; `GoStore.mapSource` sets `date` only when `s.hasDate` | Smallest model the page needs | Adding `dateSortKey` now — S15-04's |
| D-14 | Store methods *(new)*: `setSourceDate(projectDir:userID:sourceID:date: CatalogDateValueInput?) async throws -> (CatalogSource, CatalogWriteEffects)`, `setSourcePublic(projectDir:userID:sourceID:isPublic: Bool) async throws -> (CatalogSource, CatalogWriteEffects)`; `CoreMethod.setSourceDate`, `.setSourcePublic`. `FakeStore` stores them, bumps the revision, and has `setSourceDateError`, `setSourcePublicError`, `setSourceCoverError`, `getSourceWorkspaceError` *(new)* hooks. `FakeStore.sourceNotFound` *(new)* = `CoreInvokeError.coded(status: 1, code: "sources.not_found", kind: .notFound, params: [])`, thrown by its Source writes for an unknown id in place of `StoreBoom.boom` | Mirrors `setSourceCover`; the gone path is testable | — |
| D-15 | The refused date shows the shared `L10n.Errors.dateValuesInvalid`, whose copy changes to "This date isn't valid. Check the year, month and day." (UX-2); the composer shows the same | Fix the shared message, not one page | A page-only key |
| D-16 | Page shell: `PVPageColumns(headerPinned: true, scrollResetKey: sourceID, primaryFill: true, header: { SourcePageIdentityHeader(model:) }, primary: { … })`, no secondary. The header always passes its view (it renders nothing while there is no workspace), drops its frame, gutter, `surfaceCard` background and `PVDivider`: the kit draws the band's rule. Loading, the load-error Callout and `pageError` go at the top of `primary`. `SourcePageLoadingShell` becomes `PVPageColumns(headerPinned: true, header:, primary: { ProgressView })` | The kit owns page chrome | Keeping the header's own card and rule — a double rule |
| D-17 | `SourcePageOverviewLayout` *(new)*, a feature-local `Layout` in `SourcePageView.swift` over three fixed subviews: A (Description, then Credibility), M (Metadata), R (graphs, Artifacts, Notes). It reads `pvWidthClass` and the proposed width (the primary's, full pane less gutters at Wide through `primaryFill`). Regular: a group of `min(widthContentMax, width)`, left-aligned as today; A and M side by side with `PVSpacing.space11` between, each half the group; both proposed the taller one's height so Credibility sits at the bottom of A (A keeps its `Spacer`); R below at the group width after `space9`. Wide: `PVPageColumnsMetrics.split(available: width, sideMax: PVSpacing.widthSecondaryMax)`; A then R stacked in the main column, M in the side column, top-aligned; the group centered. `SourcePageOverviewHeightKey` and `metadataColumnHeight` go | No section changes parent, so drafts survive 1600 by construction; S15-23 adds the preview as a fourth subview in M's column | Moving sections between primary and secondary by `if` — loses `@State`; an always-present empty secondary — stacks a gap |
| D-18 | Copy keys use `L10n.Sources.sourceDate*` (keys `sources.page.sourceDate.*`) and identifiers `sources.page.sourceDate.*`; the shared form keeps its `sources.page.date*` keys | One prefix, one meaning | Sharing the form's prefix |

**Rules this plan bends.** None.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `sources.ErrNotFound`, `apperr.CodeSourcesNotFound` | error | `identity.ErrNotFound` (`KindNotFound`) | Same shape, Source domain |
| `EngineErrorCode.sourceNotFound` | enum case | `.conclusionNotFound` | Extends the enum |
| `CatalogSource.date`, `.isPublic` | fields | `coverMode` | Extends the struct |
| `setSourceDate`, `setSourcePublic`, `CoreMethod` cases, FakeStore hooks, `FakeStore.sourceNotFound` | func | `setSourceCover`, `updateSourceError` | Mirrors them |
| `SourceDateFormModel`, `sourceDateDialog(_:)` | model, modifier | `SourceGraphFormModel`, `sourceGraphFormDialog(_:)` | Same pattern, new form |
| `SourcePageContext.fail`, `.gone` | func, state | the 16 `L10n.Errors.message(for:)` sites | Replaces them |
| `leavesForSectionRoot(when:notice:)` | modifier param | `leavesForSectionRoot(when:)`, `ConclusionDetailPage`'s `.onChange` | Extends the shared leave; absorbs the Person page's copy |
| `SourceIdentitySection.resetIfChanged`, `pendingPrivate`, `setPrivate`, `openDate`, `dateForm` | members | `reset()`, the title/type flows | Extends the header's model |
| `SourcePageOverviewLayout` | Layout | `ConclusionPinnedSideColumn`, `PVPageColumnsLayout` | Three subviews, two arrangements; the pinned column is two subviews and a pin offset |
| `L10n.Sources.sourceDate*`, `privateSwitch*`, `sourceDeletedNotice`; `L10n.Errors.sourcesNotFound` | keys | `L10n.Errors.conclusionDetailsNotFound` | New copy |

**Builds on.** `DateValueEditorForm`, `DateValueDraft`, `DateRowDisplay`; `PVFormDialog`, `PVSwitch`, `PVPageColumns`, `PVPageColumnsMetrics.split`; `applySource`; `WorkspaceSession.noticeToast`; the `apperr` registry and the Swift vocabulary test.
**Deletes.** `SourcePageOverviewHeightKey`, `metadataColumnHeight` and `overviewColumns`; the header's own frame, gutter, background and `PVDivider`; `SourcePageContent`'s and `SourcePageLoadingShell`'s own ScrollViews; `ConclusionDetailPage`'s gone `.onChange`; the `ErrNoRows` → `ErrInvalid` branches in `loadForWrite`, `ReadDraft` and `ListWorkspace`; `FakeStore`'s `StoreBoom.boom` for an unknown Source in `updateSource` and `setSourceCover`.
**Next adds.** S15-04 adds `dateSortKey` and calls `DateRowDisplay.line(date:)` on list rows; S15-05 the same on the Person page; S15-09 adds a board in each Artifact card inside R; S15-23 adds the preview under Metadata in M's column; S15-25 adds a menu item; S15-32 reads `CatalogSource.date`.

### Reach

| New path or state | Destination | Checked: resets on the new id · handles a missing referent · shows the state |
| --- | --- | --- |
| A Source deleted elsewhere, then any page write or a Back visit | Sources list | `scrollResetKey: sourceID`; leaves through the shared modifier; notice (US-3) |
| A Source whose date row is missing | header | Go reads it as undated (S15-02 F-2); Add date |
| The page at Wide | `SourcePageOverviewLayout` | AC-4.1 – AC-4.3 |

### Scale

Each write touches one Source and applies `.updatedSource` (the page workspace and the Sources list row patch in place, as `setSourceCover`). The layout places three subviews; a Source with 60 Artifacts renders today's cards inside R.

### Edges

- **Assumptions.** `apply` reset only on the four identity fields (D-12, T-6). The layout's Regular branch must reproduce today's: A's height is the taller of A and M (AC-4.2).
- **Neighbors.** A phrase-only date shows its phrase; a range shows "bet 1834 and 1840"; an unchanged save writes nothing (S15-02 AC-1.6). A metadata edit at Wide sits in a column of `widthInspector` to `widthSecondaryMax` (340–480pt), narrower than today's half column: manual check 6.
- **Removed inputs.** A Source deleted while the dialog is open: Save gets `sources.not_found` → leave and notice (AC-3.4).
- **Untrusted size.** None.

### Schema

None — S15-02 added the columns.

### Go core

| Package | Change |
| --- | --- |
| `core/apperr` | `CodeSourcesNotFound` *(new)* |
| `core/database/sources` | `ErrNotFound` *(new)*; `getOne` maps `sql.ErrNoRows`; `loadForWrite`, `Delete`, doc comments (`getByRef`'s "or sql.ErrNoRows") per D-3 |
| `core/database/sourcecopy` | `ReadDraft`: drop the mapping; doc "A missing Source is sources.ErrNotFound" |
| `core/database/sourcemetadata` | `ListWorkspace`: drop the `ErrNoRows` branch |

### FFI

None — no new RPC. Every Source RPC returns `sources.not_found` for a missing Source through D-3. `api/ffi/swift_vocabulary_test.go` gains the code in "engine error codes".

### macOS

The board is not agreed. Until it is, the arrangement is the brief's recommendation; the plan is updated to the board before it is Planned.

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Page shell | `PVPageColumns` | `SourcePageView.swift` | D-16 |
| Overview layout | `SourcePageOverviewLayout` | `SourcePageView.swift` | D-17 |
| Meta row: ref · type · date · Private | header snowflake | `SourcePageIdentityHeader.swift` `metaRow` | SD-1; Private at the row's end, not with Copy and Delete; the row wraps under its ref when it does not fit (`PVFlowLayout`) |
| Date text | `Text` + `PVIcon(.calendar)` through `DateRowDisplay.line(date:)` | header | accessibility label `L10n.Sources.sourceDateAccessibility(date:)` |
| Edit date / Add date | `PVIconButton(.penLine, size: .sm)` / `PVButton(variant: .ghost, size: .sm, icon: .calendar)` | header | identifiers `sources.page.sourceDate.edit`, `.add`; focus returns through `FocusClaim.claim` after the dialog closes, as the type pencil |
| Date dialog | `.sourceDateDialog` over `pvFormDialog(item:)` with `DateValueEditorForm(draft:, accessibilityIdentifierPrefix: "sources.page.sourceDate")` | `SourceDateDialog.swift` | `PVCallout(tone: .danger)` above the form for `saveError` |
| Clear date | `PVButton(variant: .ghost, tone: .danger, size: .sm)` under the form, only when `initial != nil`; caption under it when `clears` | dialog | D-9 |
| Private switch | `PVSwitch(L10n.Sources.privateSwitch, isOn:, isDisabled:)` + `.help(L10n.Sources.privateSwitchHelp)` | header | identifier `sources.page.private` |
| Deleted notice | `leavesForSectionRoot(when:notice:)` | `WorkspaceNavigation.swift` | D-7 |

States: undated, dated, dialog saving, dialog clearing, refused date, switch saving, switch failed, gone. Accessibility per SD-8.

### Strings

New, `L10n.Sources`:

- `sourceDateAdd` = "Add date"; `sourceDateEdit` = "Edit date"; `sourceDateAccessibility(date:)` = "Date, %@"
- `sourceDateDialogTitle` = "Source date"; `sourceDateDialogSubtitle` = "When the record was made or first published."; `sourceDateSave` = "Save"
- `sourceDateClear` = "Clear date"; `sourceDateClearsCaption` = "Saving removes this Source's date."
- `privateSwitch` = "Private"; `privateSwitchHelp` = "Private Sources will be kept out of sharing. Nothing else changes yet."
- `sourceDeletedNotice(title:)` = "“%@” was deleted."

New, `L10n.Errors.sourcesNotFound` = "This Source no longer exists.", with `case "sources.not_found":` in `message(for:)`.
Changed: `error.datevalues.invalid` "Invalid date value." → "This date isn't valid. Check the year, month and day." (the composer shows it too).
The board's wording replaces any of these.

## 7. Steps

**STEP-1 — Go: a missing Source is not found**
Files: `core/apperr/apperr.go`, `core/database/sources/sources.go`, `core/database/sourcecopy/sourcecopy.go`, `core/database/sourcemetadata/sourcemetadata.go`, their tests, `api/ffi/handlers/sources_test.go`. Skills: `add-catalog-query`, `add-ffi-handler`.
Do: D-3; update the missing-id tests; T-1, T-2.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/... ./api/...` passes.

**STEP-2 — Typed code and copy**
Files: `ConclusionVocabulary.swift`, `api/ffi/swift_vocabulary_test.go`, `L10n.swift`, `Localizable.xcstrings`. Skill: `add-localized-string`.
Do: D-4, D-15, D-18 and every key in **Strings**.
Check: `CGO_ENABLED=1 go test -tags fts5 ./api/ffi/ -run TestSwiftVocabulary` passes; `python3 scripts/check-localizable-xcstrings.py` passes.

**STEP-3 — Swift model and store**
Files: `GenealogyStore.swift`, `GoStore.swift`, `FakeStore.swift`, `CoreInvoke.swift`.
Do: D-13, D-14.
Check: T-3 passes (`xcodebuild test`, scheme as in `scripts/check-pr.sh`).

**STEP-4 — Header model and date form model**
Files: `SourceIdentitySection.swift`, `SourceDateFormModel.swift` *(new)*, `SourcePageModel.swift`, `SourcePageContext.swift`. Skill: `add-swift-test`.
Do: D-8, D-9, D-10, D-11, D-12.
Check: T-4, T-5, T-6 pass.

**STEP-5 — Gone: one door, one leave**
Files: `SourcePageContext.swift`, `SourcePageModel.swift`, the five section files in §4, `WorkspaceNavigation.swift`, `ConclusionDetailPage.swift`, `SourcePageView.swift`.
Do: D-5, D-6, D-7.
Check: T-7, T-8 and `PersonDetailTests` pass.

**STEP-6 — Header view and dialog**
Files: `SourcePageIdentityHeader.swift`, `SourceDateDialog.swift` *(new)*, `SourcePageView.swift`. Skill: `add-ui-component`.
Do: SD-1 – SD-6, SD-8 per **macOS** and the board.
Check: `xcodebuild test` passes; manual checks 1–3, 5.

**STEP-7 — Page columns and the overview layout**
Files: `SourcePageView.swift`, `SourcePageIdentityHeader.swift`, `macos/ProvenenciaTests/SourcePageOverviewLayoutTests.swift` *(new)*.
Do: D-16, D-17.
Check: T-9 passes; manual check 6.

**STEP-8 — Version and docs**
Files: `VERSION`, `core/version.go`, `MARKETING_VERSION` in the Xcode project (0.6.0); `../completed.md`; `../deployment-plan.md` (map row); this plan (status, Landed note); `../design/README.md` (S15-D1 board line); `docs/source-layer-data-model.md` § 4 (the switch reads "Private"); `docs/macos-client-patterns.md` § 9 (the Source page is on Page columns, not left-aligned); `docs/ideas/wide-layout.md` "Source page" row; `docs/ideas/structural-debt.md` (the two Out lines).
Do: per the Definition of done. The brief stays in `design/` (it gates S15-09, S15-23, S15-25).
Check: links resolve; `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-15/prs/S15-03-source-page-date-and-private-switch.md` passes.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log`. No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/database/sources/sources_test.go` `TestMissingSourceIsNotFound` *(new)* | table over `SetCover`, `SetDate`, `SetPublic`, `Update`, `Delete`, `Get`, `GetTx` with an unused 16-byte id → `ErrNotFound`; 3-byte id → `ErrInvalid`; `sourcecopy.ReadDraft` unused id → `ErrNotFound` (in `sourcecopy`'s tests) | AC-3.1, AC-3.3, D-3 |
| T-2 | Go | `api/ffi/handlers/sources_test.go` `TestSetSourceDate` "unknown source", `TestSetSourcePublic` and `TestGetSourceWorkspace` *(new cases)*, `DeleteSource` missing id | → code `sources.not_found` | AC-3.2, D-3 |
| T-3 | Swift | `macos/ProvenenciaTests/SourcePageModelTests.swift` `fakeStoreSetsDateAndPublic` *(new)* | FakeStore set, clear, flip → the Source reads back | D-13, D-14 |
| T-4 | Swift | same file, `dateFormSetsChangesAndClears` *(new)* | `openDate` → blank draft, `canSave == false`; year 1880 → `save` → `source.date?.startYear == 1880`, `dateForm == nil`; `openDate` → prefilled; `clear()` → `canSave`, `save` → `source.date == nil` | AC-1.1, AC-1.2, AC-1.3, AC-1.5, AC-1.6, D-8, D-9 |
| T-5 | Swift | same file, `refusedDateKeepsTheDialog` *(new)* | `setSourceDateError = coded("datevalues.invalid")` → `dateForm` non-nil, draft kept, `saveError == L10n.string(L10n.Errors.dateValuesInvalid)` | AC-1.7, D-15 |
| T-6 | Swift | same file, `privateSwitchSavesRevertsAndKeepsEdits` *(new)* | `setPrivate(true)` → `isPublic == false`; with an internal error → `isPublic == true`, `pageError` set; with `identity.editingDescription` and a draft, `setPrivate(true)` → still editing, draft kept | AC-2.2, AC-2.3, AC-2.4, D-11, D-12 |
| T-7 | Swift | same file, `writeToADeletedSourceLeaves` *(new)* | `setSourcePublicError`, `updateSourceError`, `setSourceCoverError` each `FakeStore.sourceNotFound` → `model.gone == "“…” was deleted."`, no field error | AC-3.4, D-5 |
| T-8 | Swift | same file, `readOfADeletedSourceIsGone` *(new)* | `getSourceWorkspaceError = FakeStore.sourceNotFound` → `warmFromSession` → `model.gone == L10n.string(L10n.Errors.sourcesNotFound)`; an internal error → `gone == nil` | AC-3.5, D-6 |
| T-9 | Swift | `macos/ProvenenciaTests/SourcePageOverviewLayoutTests.swift` *(new)* | the layout's frames for three fixed-size subviews: Regular 1102 → A and M side by side, equal heights, R below; Wide 2296 → A over R in a `widthContentMax` main column, M beside at `split`'s side width, group centered | AC-4.1, AC-4.2, D-17 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `SourcePageModelTests`, `PersonDetailTests`, handler tests expecting `sources.ErrInvalid` for a missing Source, the Swift vocabulary test.

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left
- [ ] Sweep: each renamed or deleted name `rg`'d by its old, qualified name across `.agents/`, `.cursor/rules/`, `docs/*.md` and `L10n.swift` with no hit left; `scripts/check-guidance-names.py` passes (check-pr runs it)
- [ ] Every row of **Rules this plan bends** has the user's answer, and its rule is updated
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New
- [ ] Version bump 0.5.0 → 0.6.0 (VERSION, core/version.go, MARKETING_VERSION together)
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] design/README.md: S15-D1's board line; the brief stays in design/ until S15-25 lands
- [ ] Domain docs updated: source-layer-data-model.md § 4, macos-client-patterns.md § 9, ideas/wide-layout.md, ideas/structural-debt.md
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

On a project with a census Source with metadata, three graphs and four Artifacts:

1. Undated Source: the header shows Add date (AC-1.1). Add with an empty form: Save is disabled (AC-1.6). Add 1880 → "1880" (AC-1.2). Reopen with the pencil: prefilled (AC-1.3). Edit to About 12 March 1880 → "abt 12 Mar 1880" (AC-1.4). Change the month, press Cancel: it asks before discarding (AC-1.8). Reopen the page: the date is there (bar 1, Source page half).
2. Pencil → Clear date: the caption shows; Save → Add date again (AC-1.5).
3. A Source never flipped shows Private off (AC-2.1). Start editing the description, flip Private on: the edit stays (AC-2.4). Leave and reopen the Source: still on (AC-2.2, bar 2).
4. Open the Source in a second window, delete it there, then flip Private in the first: the Sources list with "“…” was deleted." (AC-3.4). Press Back to the deleted Source: the Sources list with "This Source no longer exists." (AC-3.5). Open a Person page, merge it away in another window, return: its notice as before (AC-3.6).
5. Tab through the meta row: type pencil, date pencil, Private. VoiceOver reads "Date, abt 12 Mar 1880" and "Private, off". Close the dialog with Escape: focus is on the pencil (SD-8).
6. At 1102pt the page looks as before (AC-4.2). Widen past 1600pt: Metadata in the side column, the header pinned (AC-4.1); its rows, suggestions and Add fit the column. With the date dialog open, then a metadata value in edit, then the description in edit, cross 1600pt both ways: the dialog, both edits' text and the scroll position are kept (AC-4.3).

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| Metadata rows are cramped in the 340–480pt side column | Manual check 6; the board draws it |
| `primaryFill`'s header band keeps `widthContentMax` while the overview group at Wide is wider | The board decides; Stop condition if it asks for a band the kit cannot give |
| The layout's Regular branch differs from today's two-column look | T-9 and manual check 6 against a screenshot of `main` |

**Stop and report if:**

- The agreed board differs from SD-1 – SD-8, SD-W1 or SD-W2 as written here (update the plan first).
- The board asks for the header band to span the overview group at Wide (`PVPageColumns` would need a new option).
- A Go caller outside §4 branches on `sql.ErrNoRows` from `sources` (`rg "sources\.(Get|GetTx)\(" --type go`).
- `DateValueEditorForm` needs a change to work in a dialog.
- `CoreErrorKind` has no `.notFound` for `FakeStore.sourceNotFound`.

## 12. Open questions

1. **The board.** S15-D1 is not drawn. When it is agreed: how does it reach the implementer (a link the implementer opens with the Artifact tool's `read`, or exported frames)? The plan's macOS section and Strings are then updated to the board, round 2 of the review runs, and this question closes.

## 13. PR

**Title:** S15-03 — The Source page shows the date and the Private switch
**Body:** objective in one sentence; what shipped against US-1 – US-4; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-15/prs/S15-03-source-page-date-and-private-switch.md`.
