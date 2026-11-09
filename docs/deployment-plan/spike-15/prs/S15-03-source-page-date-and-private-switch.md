# S15-03 — The Source page shows the date and the Private switch

**Status:** Blocked — on board S15-D1
**Spike / slice:** Spike 15 · Slice 1 — Source date, privacy, credibility
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** [S15-D1](../design/S15-D1-source-page.md), board not drawn yet
**Board:** — (ask the user how the agreed board reaches the implementer when it is agreed; until then **Design → macOS** describes each frame in words from the brief)
**Depends on:** S15-02 (landed [#762](https://github.com/mendahu/provenencia/pull/762)), S15-D1
**Blocks:** S15-09, S15-23, S15-25 (they build on the Page columns arrangement this PR adds)
**Branch:** `s15-03-source-page-date-and-private-switch`
**Version bump:** minor — new Source page controls and a new error code old clients map to a generic message ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-ui-component`](../../../../.agents/skills/add-ui-component/SKILL.md), [`add-localized-string`](../../../../.agents/skills/add-localized-string/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md), [`add-date-value`](../../../../.agents/skills/add-date-value/SKILL.md), [`add-ffi-handler`](../../../../.agents/skills/add-ffi-handler/SKILL.md), [`add-catalog-query`](../../../../.agents/skills/add-catalog-query/SKILL.md)
**Rules that bind:** `.cursor/rules/go-tests.mdc`, `.cursor/rules/ffi-handlers.mdc`, `.cursor/rules/catalog-queries.mdc`, `.cursor/rules/versioning.mdc`
**Context:** working rules "The Source date is one optional DateValue on `sources`", "Not public is a stored switch and nothing else" ([`../deployment-plan.md`](../deployment-plan.md#working-rules)); [`../source-provenance-date.md`](../source-provenance-date.md); [privacy note](../2026-10-27-source-privacy-toggle.md); S15-02's [plan](S15-02-source-date-and-public-flag.md) and [ship log](S15-02-source-date-and-public-flag.ship.md) (UX-1, UX-2); [`macos-client-patterns.md` § 9](../../../macos-client-patterns.md#9-window-widths-regular-and-wide); [`../../../ideas/wide-layout.md`](../../../ideas/wide-layout.md) "Source page"
**Planned:** 2026-11-09 against `main` @ `f439a14`
**Review:** — (not reviewed yet; runs when the board is agreed and the macOS section is final)

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

On the Source page, a researcher sets, changes and clears the Source's own date (when the record was made or first published) from its header, with the app's structured date form in a dialog, and marks the Source private with a switch that saves at once. The page moves onto Page columns, so at Wide its Metadata board sits in a second column instead of leaving most of the pane empty. A write to a Source deleted in another window now says so and goes back to the Sources list instead of showing "Invalid source." This is the Source page half of dogfood bars 1 and 2.

**Why now.** S15-02 stores both facts and nothing shows them; S15-09, S15-23 and S15-25 add to the page arrangement this PR sets.

**What stays.** Every section's content and controls, and the Regular arrangement (Description and Credibility beside Metadata, then graphs, Artifacts, Notes). Title, type and description keep their shared `updateSource` write. The Sources list (S15-04) and Copy Source (S15-32) are untouched. The Private flag changes nothing else in the app (working rule).

## 2. User stories

**US-1.** As a researcher, I give a Source its date, change it and clear it from the Source page, so the record carries when it was made.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | An undated Source; the header | Shows "Add date" in the date's place (SD-2) |
| AC-1.2 | Add date → the dialog; kind Point, year 1880; Save | The dialog closes; the header shows "1880"; `setSourceDate` was called once with `kind: "point", startYear: 1880` |
| AC-1.3 | That Source; the pencil → the dialog | The form is prefilled with point 1880 (`DateValueDraft(from:)`); Clear date is offered |
| AC-1.4 | Change to About, March, 12; Save | The header shows "abt. 12 Mar 1880" |
| AC-1.5 | Clear date → Confirm "Clear date" | `setSourceDate` with no date; the header shows "Add date"; the dialog is closed |
| AC-1.6 | The dialog with a blank form | Save is disabled |
| AC-1.7 | Save fails with `datevalues.invalid` | The dialog stays open with its draft; a danger Callout reads "This date isn't valid. Check the year, month and day." (UX-2) |
| AC-1.8 | The dialog is dirty; Cancel | The FormDialog asks before discarding, as every FormDialog |

**US-2.** As a researcher, I mark a Source private, so later sharing can leave it out.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | A Source with `isPublic: true` | The Private switch is off |
| AC-2.2 | Flip it on | `setSourcePublic(isPublic: false)` is called once; while it runs the switch is disabled; after it the switch is on and reopening the page shows it on |
| AC-2.3 | Flip it on; the write fails with `internal` | The switch is back off; the page's error Callout shows the error message |

**US-3.** As a researcher, when another window has deleted this Source, a write from this page tells me and takes me back to the Sources list (UX-1).

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | Go: `sources.SetDate`, `SetPublic`, `SetCover`, `Update`, `Delete` on an id with no row | `sources.ErrNotFound` (code `sources.not_found`, kind not found) |
| AC-3.2 | Go: `sources.Get` on an id with no row; `GetSourceWorkspace` RPC on it | `sources.ErrNotFound` |
| AC-3.3 | Swift: a date, switch, title, type, description or cover write throws `sources.not_found` | The page leaves for the Sources list; `session.noticeToast` reads "“1880 census” was deleted." with the title last shown |
| AC-3.4 | Swift: the page's workspace read fails with `sources.not_found` | The page leaves (as today) and `session.noticeToast` reads "This Source was deleted." |
| AC-3.5 | Go: a malformed id (not 16 bytes) or a missing user | Still `sources.ErrInvalid` |

**US-4.** As a researcher on a wide monitor, the Source page uses the width.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-4.1 | Pane 2296pt | Page columns: the header pinned across the group; primary holds Description, Credibility, graphs, Artifacts, Notes; secondary holds Metadata (SD-W1, as the board agrees) |
| AC-4.2 | Pane 1102pt | As today: Description and Credibility beside Metadata, then graphs, Artifacts, Notes (SD-W2) |
| AC-4.3 | Resize across 1600pt with the date dialog open, and with the description in edit | The dialog stays open with its draft; the description edit keeps its text; scroll position is kept |

## 3. Scope

| In | Covered by |
| --- | --- |
| SD-1, SD-2: the date and Add date in the header's meta row | STEP-4, AC-1.1, AC-1.4 |
| SD-3, SD-4, SD-5: the date dialog, Clear date with Confirm, the refused-date Callout (UX-2) | STEP-3, STEP-4, AC-1.2 – AC-1.8 |
| SD-6: the Private switch | STEP-3, STEP-4, AC-2.1 – AC-2.3 |
| SD-7: a Source deleted elsewhere — Go `sources.ErrNotFound` for every Source write and the Source read (UX-1); Swift notice and leave | STEP-1, STEP-5, AC-3.1 – AC-3.5 |
| SD-8: tab order, VoiceOver, focus back after the dialog | STEP-4, manual check 5 |
| SD-W1, SD-W2: the page on `PVPageColumns` | STEP-6, AC-4.1 – AC-4.3 |
| `CatalogSource.date`, `.isPublic`; `setSourceDate`, `setSourcePublic` on `GenealogyStore`, `GoStore`, `FakeStore` | STEP-2 |

| Out | Home |
| --- | --- |
| The date on Sources list rows and the date sort; `CatalogSource.dateSortKey` | S15-04 |
| An Artifact's metadata board (SD-9, SD-10) | S15-09 |
| The citation preview (SD-11 – SD-13) | S15-23 |
| Move on Artifact and graph cards (SD-14) | S15-25 |
| Copy Source carrying the date | S15-32 |
| Any use of the Private flag (badge, filter, sharing) | not planned in this spike (working rule) |
| Not-found handling for note, metadata, credibility and Artifact writes to a deleted Source (they fail in their own packages) | [`docs/ideas/value-row-owner-check.md`](../../../ideas/value-row-owner-check.md) gains a line (STEP-7) |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Header | `macos/App/Features/Sources/SourcePageIdentityHeader.swift` `SourcePageIdentityHeader`, `metaRow` | Cover, title inline edit, meta row: ref · type pill + pencil; Copy and Delete. Frames itself with `PVSpacing.widthContentMax` and `gutterPage`; no date, no switch |
| Page shell | `macos/App/Features/Sources/SourcePageView.swift` `SourcePageContent`, `overviewColumns` | `VStack(header, ScrollView { … })` with its own `widthContentMax` frame, left-aligned; no Wide arrangement. `.leavesForSectionRoot(when: workspaceHandle.read.isFailed)` leaves silently on any failed read |
| Header model | `macos/App/Features/Sources/SourceIdentitySection.swift` `SourceIdentitySection` | Title, type, description drafts sharing one `updateSource` save; errors through `L10n.Errors.message(for:)` |
| Page context | `macos/App/Features/Sources/SourcePageContext.swift` `applySource`, `setSourceCover` | `applySource` patches the workspace and applies `.updatedSource`; cover writes set `pageError` on failure |
| Page model | `macos/App/Features/Sources/SourcePageModel.swift` | Composes the sections; `confirmResourceDelete(leaving:)` shows how a delete leaves the page |
| Date form | `macos/App/Features/Dates/DateValueEditorForm.swift` `DateValueEditorForm`; `macos/App/Platform/DateValueDraft.swift` `DateValueDraft` (`isBlank`, `isValid`, `toInput()`, `init(from:)`); `macos/App/Platform/DateValueDisplay.swift` `DateValueDisplay.string(for: CatalogDateValueInput)` | The shared date form; only the composer uses it (`CitationComposerAuthoringBox.swift`). Its identifier prefix already defaults to `sources.page.date` |
| Switch | `macos/App/DesignSystem/Components/Switch/PVSwitch.swift` `PVSwitch` | Exists; no Features caller |
| Page columns | `macos/App/DesignSystem/Components/PageColumns/PVPageColumns.swift` `PVPageColumns(headerPinned:header:primary:secondary:)`; `hasSecondary` is decided by the `Secondary` type | `ConclusionDetailPage.swift` is the caller with a pinned header |
| Gone notice precedent | `macos/App/Features/Conclusions/ConclusionDetailPage.swift` (`presentation == .gone` → `session.noticeToast`, `.leavesForSectionRoot`); `WorkspaceSession.noticeToast` | The Person page's pattern for a record that no longer exists |
| Swift model | `macos/App/Platform/GenealogyStore.swift` `CatalogSource`; `GoStore.swift` `mapSource`, `setSourceCover`, `mapDateValueToProto`, `mapDateValue`; `FakeStore.swift` `setSourceCover`; `CoreInvoke.swift` `CoreMethod.setSourceCover` | `CatalogSource` has no date or flag; no store method for the two RPCs. The generated `Provenencia_Engine_V1_Source` has `date`/`hasDate`, `isPublic`; `Method.setSourceDate` (117), `.setSourcePublic` (118) |
| Go writes | `core/database/sources/sources.go` `loadForWrite` | `sql.ErrNoRows` → `ErrInvalid` (`sources.invalid`), for `SetCover`, `SetDate`, `SetPublic`, `Update`, `Delete` |
| Go read | `sources.Get` → `getOne` | A missing row returns raw `sql.ErrNoRows`, which reaches Swift as an internal error. `sourcemetadata.ListWorkspace` (line ~504) and the `UpdateSource` handler call `sources.Get` |
| Error codes | `core/apperr/apperr.go` (`CodeSourcesInvalid`, `KindNotFound`); `macos/App/Platform/L10n.swift` `L10n.Errors` (`case "sources.invalid":`) | No `sources.not_found` |
| Tests | `core/database/sources/sources_test.go` (line ~836, `runDelete` on a missing id expects `ErrInvalid`); `api/ffi/handlers/sources_test.go` (DeleteSource missing id ~792, `TestSetSourceDate` "unknown source" ~1089); `macos/ProvenenciaTests/SourcePageModelTests.swift` | The missing-Source cases expect `ErrInvalid` |

**Precedent.** The Person page's gone handling (`ConclusionDetailPage.swift`, `conclusiondetails.not_found` → `noticeToast` + `leavesForSectionRoot`) for US-3; `SourcePageContext.setSourceCover` for a one-attribute Source write and `applySource`; `ConclusionDetailPage`'s `PVPageColumns(headerPinned:…)` for the shell; the composer's `DateValueEditorForm` for the form.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | The date is edited in a FormDialog opened from the header (user, 2026-11-09) | The shared form is tall; a dialog keeps the header compact | Inline expanding editor — pushes the page down under a pinned header |
| D-2 | The switch is labeled "Private", off by default; on writes `isPublic: false` (user, 2026-11-09) | The user's wording; the stored flag stays `is_public` | "Public", on by default |
| D-3 | Go: `sources.ErrNotFound` *(new)*, `apperr.CodeSourcesNotFound = "sources.not_found"` *(new)*, `KindNotFound`. `loadForWrite` and `Get` return it for `sql.ErrNoRows`; malformed ids and users stay `ErrInvalid`. `GetTx` and `getByRef` are unchanged. `sourcemetadata.ListWorkspace` checks `errors.Is(err, sources.ErrNotFound)` in place of `sql.ErrNoRows` and keeps returning its own `ErrInvalid` (user, 2026-11-09: "not-found code, then leave") | Fixes the code where it is made, so every Source write gains it | Swift re-reads after any failure — silent, and `Get`'s raw `ErrNoRows` still reads as internal |
| D-4 | Swift: one `SourcePageContext.handleGone(_ error: Error) -> Bool` *(new)*: when `(error as? CoreInvokeError)?.code == "sources.not_found"`, it sets `session.noticeToast` (`L10n.Sources.sourceDeletedNotice(title:)`), sets `sourceGone = true` *(new, observable)* and returns true. Every Source-row write on the page calls it first in its `catch`: `SourceIdentitySection.save`, the date and switch writes, `setSourceCover`. The view leaves through `.leavesForSectionRoot(when: workspaceHandle.read.isFailed || model.sourceGone)` | One place; the leave stays the view's, as today | A `leave` closure passed into each section — the sections do not know navigation |
| D-5 | A failed workspace read with `sources.not_found` sets `noticeToast` to `L10n.Sources.sourceGoneNotice` ("This Source was deleted.") in `SourcePageContent`'s `.onChange(of: workspaceHandle.read.isFailed)`, the Person page's pattern; any other failed read leaves silently, as today | Matches `ConclusionDetailPage` | A toast for every failed read — other failures are not deletions |
| D-6 | The date and switch state live in `SourceIdentitySection` (the header's model): `dateDraft: DateValueDraft?` *(new; non-nil while the dialog is open)*, `dateError: String?`, `pendingClearDate: SourceDateClear?` *(new; item snapshot for the Confirm, `struct SourceDateClear: Identifiable { let id: String }`)*, `isSavingDate`, `isSavingPublic`; functions `openDate()`, `cancelDate()`, `saveDate()`, `askClearDate()`, `confirmClearDate()`, `setPrivate(_:)` *(new)* | The header owns these controls; title, type and description already live here | A new section model — one more object for two controls on the same header |
| D-7 | Each write calls the store, then `context.applySource(updated, effects:)`; the date and switch read their saved values from `context.source` (`date`, `isPublic`) | The S15-02 RPCs return the Source and effects, as `setSourceCover` does; `.updatedSource` patches the Sources list row too | A workspace re-read |
| D-8 | The switch binds to a computed `Binding` whose `get` is `!(source.isPublic)` while not saving, and the requested value while saving; `set` starts `setPrivate`. On failure the binding reads the saved value again (it flips back) and `pageError` takes the message | No local `@State` copy to drift | Optimistic local state |
| D-9 | `CatalogSource` gains `date: CatalogDateValueInput?` and `isPublic: Bool = true`; `GoStore.mapSource` fills `date` only when `s.hasDate`. `dateSortKey` is S15-04's | Smallest model the page needs | Adding the sort key now — S15-04 owns its reader |
| D-10 | Store methods *(new)* on `GenealogyStore`: `setSourceDate(projectDir:userID:sourceID:date: CatalogDateValueInput?) async throws -> (CatalogSource, CatalogWriteEffects)` and `setSourcePublic(projectDir:userID:sourceID:isPublic:) async throws -> (CatalogSource, CatalogWriteEffects)`; `CoreMethod.setSourceDate`, `.setSourcePublic` *(new)*. `FakeStore` stores them on its Source list, bumps the activity revision, and has `setSourceDateError` / `setSourcePublicError` *(new)* hooks; it throws `sources.not_found` for an unknown id | Mirrors `setSourceCover` | One combined RPC — S15-02 shipped two |
| D-11 | Refused-date copy: `datevalues.invalid` inside the date dialog shows `L10n.Sources.dateInvalid` *(new)*; every other error shows `L10n.Errors.message(for:)` (UX-2) | Names the field without changing the shared error copy | Changing `error.datevalues.invalid` — the composer also uses it |
| D-12 | Page columns: one `PVPageColumns(headerPinned: true, header:, primary:, secondary:)` always, so crossing 1600 keeps one scroll view. `SourcePageContent` reads `@Environment(\.pvWidthClass)`: at Wide the primary holds Description and Credibility stacked, then graphs, Artifacts, Notes, and the secondary holds `SourcePageMetadataView`; at Regular the primary holds today's `overviewColumns` and the secondary is empty. The header drops its own `widthContentMax` frame and gutter (Page columns gives the band) and keeps its background and divider as the pinned band | The page keeps one identity across widths (rule 5) | Two `PVPageColumns` branches by width — a resize rebuilds the scroll view |
| D-13 | The file-drop overlay, dialogs and confirms stay on the outer view, outside `PVPageColumns` | They cover the pane, not a column | — |

**Rules this plan bends.** None.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `sources.ErrNotFound`, `apperr.CodeSourcesNotFound` *(new)* | error | `identity.ErrNotFound`, `CodeConclusionDetailsNotFound` | Same shape, Source domain |
| `CatalogSource.date`, `.isPublic` *(new)* | fields | `coverMode`, `primaryArtifactID` | Extends the struct |
| `setSourceDate`, `setSourcePublic` store methods, `CoreMethod` cases *(new)* | func | `setSourceCover` | Mirrors it |
| Date and switch state in `SourceIdentitySection` *(new members)* | model | the title/type/description flows | Extends the header's model (D-6) |
| `SourceDateClear` *(new)* | type | `SourceMetadataSection.pendingDelete` item | Item snapshot for `pvConfirm` |
| `SourcePageContext.handleGone`, `sourceGone` *(new)* | func | `ConclusionDetailPage` gone handling | Same pattern, Source page |
| `L10n.Sources` keys (Strings) *(new)* | key | `L10n.Sources.date*` form keys | New copy |

**Builds on.** `DateValueEditorForm`, `DateValueDraft`, `DateValueDisplay`; `PVFormDialog`, `PVConfirm`, `PVSwitch`, `PVPageColumns`; `applySource`; `WorkspaceSession.noticeToast`; `apperr` registry.
**Deletes.** `SourcePageIdentityHeader`'s `.frame(maxWidth: PVSpacing.widthContentMax)` and horizontal gutter, and `SourcePageContent`'s ScrollView with its own content-width frame (Page columns replaces both).
**Next adds.** S15-04 adds `dateSortKey` to `CatalogSource`; S15-09 puts a board on each Artifact card in the primary; S15-23 adds the Citation preview to the secondary at Wide; S15-25 adds a menu item; S15-32 reads `CatalogSource.date` in the Copy form.

### Reach

| New path or state | Destination | Checked: resets on the new id · handles a missing referent · shows the state |
| --- | --- | --- |
| A Source deleted elsewhere, then a write | `SourcePageContent` → Sources list | Leaves through `.leavesForSectionRoot`; notice toast (US-3) |
| A Source whose date row is missing | header | Go reads it as undated (S15-02 F-2); shows Add date |
| The page at Wide | `PVPageColumns` | First Wide arrangement; AC-4.1 – AC-4.3 |

### Scale

Each write touches one Source and applies `.updatedSource` (the Source page workspace and the Sources list row patch in place, as `setSourceCover` does today). No list or loop changes. A Source with 60 Artifacts renders the same cards as today, in the primary column.

### Edges

- **Assumptions.** `SourcePageModel.apply` resets `identity` when `previous?.source != workspace.source`; with `date` and `isPublic` on `CatalogSource`, a date or switch write also resets the title, type and description edit state. `identity.reset()` must leave `dateDraft` alone while the dialog is open (T-4).
- **Neighbors.** A phrase-only date ("in the reign of Victoria") shows as its phrase; a range shows "1834–1840". A date the researcher saves unchanged writes nothing (Go returns no effects; AC-1.6 of S15-02).
- **Removed inputs.** A Source deleted while the dialog is open: Save gets `sources.not_found` → leave and notice (AC-3.3).
- **Untrusted size.** None — no list read from data.

### Schema

None — S15-02 added the columns.

### Go core

| Package | Change |
| --- | --- |
| `core/apperr` | `CodeSourcesNotFound = "sources.not_found"` *(new)* |
| `core/database/sources` | `var ErrNotFound = apperr.New(apperr.CodeSourcesNotFound, apperr.KindNotFound)` *(new)*; `loadForWrite` and `Get` map `sql.ErrNoRows` to it; doc comments updated |
| `core/database/sourcemetadata` | `ListWorkspace`: `errors.Is(err, sources.ErrNotFound)` in place of `sql.ErrNoRows` |

### FFI

None — no new RPC. `GetSourceWorkspace`, `UpdateSource`, `DeleteSource`, `SetSourceCover`, `SetSourceDate`, `SetSourcePublic` return `sources.not_found` for a missing Source through the package change.

### macOS

The board is not agreed. Until it is, the arrangement is the brief's recommendation; the implementer builds what the agreed board shows (the plan is updated when it is).

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Page shell | `PVPageColumns` | `SourcePageView.swift` `SourcePageContent` | D-12 |
| Meta row: ref · type · date · Private | snowflake in `SourcePageIdentityHeader` | `SourcePageIdentityHeader.swift` `metaRow` | SD-1 |
| Date text | `Text` + `PVIcon(.calendar)` (or the icon the board names) through `DateValueDisplay.string(for:)` | header | Accessibility label `L10n.Sources.dateAccessibility(date:)` |
| Edit date / Add date | `PVIconButton(.penLine, size: .sm)` / `PVButton(variant: .ghost, size: .sm, icon: .calendar)` | header | identifiers `sources.page.date.edit`, `sources.page.date.add` |
| Date dialog | `.pvFormDialog(isPresented:)` with `DateValueEditorForm(draft:, accessibilityIdentifierPrefix: "sources.page.date")` | `SourcePageView.swift` beside the other dialogs | confirmDisabled `draft.isBlank || !draft.isValid`; isDirty draft ≠ the saved date's draft; prefix `sources.page.dateDialog` |
| Clear date | `PVButton(variant: .ghost, tone: .danger)` in the dialog's form; `.pvConfirm(item:)` on `pendingClearDate` | dialog + `SourcePageView.swift` | Only when the Source has a date |
| Refused date | `PVCallout(tone: .danger)` above the form | dialog | D-11 |
| Private switch | `PVSwitch(L10n.Sources.privateSwitch, isOn:, isDisabled:)` with `.help(L10n.Sources.privateSwitchHelp)` | header | identifier `sources.page.private` |
| Deleted notice | `session.noticeToast` | `SourcePageContext.handleGone`, `SourcePageContent` | D-4, D-5 |

States: undated (Add date), dated, dialog saving (`isRunning`), refused date, switch saving (disabled), switch failed (page Callout), gone (leave + toast). Accessibility: SD-8 tab order follows view order; after the dialog closes, focus returns to the pencil or Add date through `FocusClaim.claim` as the type pencil does.

### Strings

L10n keys *(new)*, `L10n.Sources`:

- `dateAdd` = "Add date"; `dateEdit` = "Edit date"
- `dateAccessibility(date:)` = "Date, %@"
- `dateDialogTitle` = "Source date"; `dateDialogSubtitle` = "When the record was made or first published."; `dateDialogSave` = "Save"
- `dateClear` = "Clear date"; `dateClearConfirmTitle` = "Clear this Source's date?"; `dateClearConfirmMessage` = "The Source will be undated and sort last in the Sources list."; `dateClearConfirm` = "Clear date"; `dateClearKeep` = "Keep"
- `dateInvalid` = "This date isn't valid. Check the year, month and day."
- `privateSwitch` = "Private"; `privateSwitchHelp` = "Private Sources will be kept out of sharing. Nothing else changes yet."
- `sourceDeletedNotice(title:)` = "“%@” was deleted."; `sourceGoneNotice` = "This Source was deleted."

`L10n.Errors.sourcesNotFound` *(new)* = "This Source no longer exists." with `case "sources.not_found":` in `message(for:)`. The board's wording replaces any of these.

## 7. Steps

**STEP-1 — Go: a missing Source is not found**
Files: `core/apperr/apperr.go`, `core/database/sources/sources.go`, `core/database/sourcemetadata/sourcemetadata.go`, `core/database/sources/sources_test.go`, `api/ffi/handlers/sources_test.go`. Skill: `add-catalog-query`, `add-ffi-handler`.
Do: D-3. Update the missing-id tests to `ErrNotFound` and add T-1, T-2.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/... ./api/...` passes.

**STEP-2 — Swift model and store**
Files: `GenealogyStore.swift`, `GoStore.swift`, `FakeStore.swift`, `CoreInvoke.swift`, `L10n.swift`, `Localizable.xcstrings`. Skill: `add-localized-string`.
Do: D-9, D-10; `L10n.Errors.sourcesNotFound`.
Check: `xcodebuild test` builds; T-3 passes.

**STEP-3 — Header model**
Files: `SourceIdentitySection.swift`, `SourcePageContext.swift`, `SourcePageModel.swift`. Skill: `add-swift-test`.
Do: D-4, D-6, D-7, D-8, D-11; `handleGone` in `SourceIdentitySection.save` and `setSourceCover`; expose `sourceGone` on `SourcePageModel`.
Check: T-4 – T-7 pass.

**STEP-4 — Header view and dialog**
Files: `SourcePageIdentityHeader.swift`, `SourcePageView.swift`, `L10n.swift`, `Localizable.xcstrings`. Skills: `add-ui-component`, `add-localized-string`.
Do: SD-1 – SD-6, SD-8 per **macOS** and the board; the Strings.
Check: `xcodebuild test` passes; manual checks 1–3.

**STEP-5 — Leave when the Source is gone**
Files: `SourcePageView.swift`.
Do: D-4's `.leavesForSectionRoot` condition; D-5's read-failure notice.
Check: T-6, T-7; manual check 4.

**STEP-6 — Page columns**
Files: `SourcePageView.swift`, `SourcePageIdentityHeader.swift`.
Do: D-12, D-13 per the board's SD-W1 and SD-W2.
Check: manual check 6 at 1102pt and 2296pt; AC-4.3.

**STEP-7 — Docs**
Files: `../completed.md`, `../deployment-plan.md` (map row; S15-03 scope entry says "Private"), this plan (status, Landed note), `../design/README.md` (S15-D1 board line), `docs/source-layer-data-model.md` § 4 (the switch reads "Private"), `docs/ideas/value-row-owner-check.md` (the Out line on other writes to a deleted Source).
Do: per the Definition of done. The brief stays in `design/` (it gates S15-09, S15-23, S15-25).
Check: links resolve; `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-15/prs/S15-03-source-page-date-and-private-switch.md` passes.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log`. No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/database/sources/sources_test.go` `TestMissingSourceIsNotFound` *(new)* | table over `SetCover`, `SetDate`, `SetPublic`, `Update`, `Delete`, `Get` with an unused 16-byte id → `ErrNotFound`; with a 3-byte id → `ErrInvalid` | AC-3.1, AC-3.2, AC-3.5, D-3 |
| T-2 | Go | `api/ffi/handlers/sources_test.go` `TestSetSourceDate` "unknown source", `TestSetSourcePublic` *(new case)*, `TestGetSourceWorkspace` *(new case)* | unknown id → error code `sources.not_found` | AC-3.1, AC-3.2 |
| T-3 | Swift | `macos/ProvenenciaTests/SourcePageModelTests.swift` `fakeStoreSetsDateAndPublic` *(new)* | FakeStore set, clear, flip → the Source reads back | D-9, D-10 |
| T-4 | Swift | same file, `dateDialogSetsChangesAndClears` *(new)* | `openDate` → draft empty; year 1880 → `saveDate` → `source.date.startYear == 1880`, dialog closed; `openDate` prefilled; `askClearDate` + `confirmClearDate` → `source.date == nil` | AC-1.1 – AC-1.6, D-6 |
| T-5 | Swift | same file, `refusedDateKeepsTheDialog` *(new)* | `setSourceDateError = coded("datevalues.invalid")` → dialog open, draft kept, `dateError == L10n.string(L10n.Sources.dateInvalid)` | AC-1.7, D-11 |
| T-6 | Swift | same file, `privateSwitchSavesAndRevertsOnFailure` *(new)* | `setPrivate(true)` → `isPublic == false`; with an internal error → `isPublic == true`, `pageError` set | AC-2.2, AC-2.3, D-8 |
| T-7 | Swift | same file, `writeToADeletedSourceLeaves` *(new)* | `setSourcePublicError` and `updateSourceError` = `coded("sources.not_found")` → `sourceGone`, `session.noticeToast?.title` names the title | AC-3.3, D-4 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `SourcePageModelTests` (identity reset on Source change), handler tests expecting `sources.ErrInvalid` for a missing Source.

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
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together) — minor
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] design/README.md: S15-D1's board line; the brief stays in design/ until S15-25 lands
- [ ] Domain doc updated: docs/source-layer-data-model.md § 4 names the switch "Private"; docs/ideas/value-row-owner-check.md has the Out line
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

On a project with a census Source with metadata, three graphs and four Artifacts:

1. Undated Source: the header shows Add date (AC-1.1). Add with an empty form: Save is disabled (AC-1.6). Add 1880 → "1880" (AC-1.2). Reopen with the pencil: prefilled (AC-1.3). Edit to About 12 March 1880 → "abt. 12 Mar 1880" (AC-1.4). Change the month, press Cancel: it asks before discarding (AC-1.8). Reopen the page: the date is there (bar 1, Source page half).
2. Clear date → Confirm → Add date again (AC-1.5).
3. A Source never flipped shows Private off (AC-2.1). Flip Private on, leave and reopen the Source: still on (AC-2.2, bar 2).
4. Open the Source in a second window, delete it there, then flip Private in the first: the Sources list with "“…” was deleted." (AC-3.3). Open a Back entry to the deleted Source: the Sources list with "This Source was deleted." (AC-3.4).
5. Tab through the meta row: type pencil, date pencil, Private. VoiceOver reads "Date, about 12 March 1880" and "Private, off". Close the dialog with Escape: focus is on the pencil.
6. At 1102pt the page looks as before (AC-4.2); widen past 1600pt: Metadata moves to the second column and the header stays pinned (AC-4.1); with the date dialog open, then with the description in edit, cross 1600pt both ways: the dialog, the edit's text and the scroll position are kept (AC-4.3).

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| `identity.reset()` on a Source change closes the date dialog mid-edit (Edges) | T-4; `reset()` does not touch `dateDraft` |
| An always-present empty secondary at Regular leaves a visible gap under the primary | Measure at 1102pt in STEP-6 |
| The metadata board's local `@State` drafts are lost when it moves between primary and secondary across 1600 | AC-4.3 with an open metadata value edit |

**Stop and report if:**

- The agreed board differs from SD-1 – SD-8, SD-W1 or SD-W2 as written here (update the plan first).
- `PVPageColumns` with an empty secondary at Regular draws a gap or changes the Regular layout (D-12).
- A metadata edit or the description edit loses its text when the window crosses 1600pt (AC-4.3).
- Any caller of `sources.Get` other than those in §4 depends on `sql.ErrNoRows` (`rg "sources.Get\(" --type go`).
- `DateValueEditorForm` needs a change to work in a dialog.

## 12. Open questions

1. **The board.** S15-D1 is not drawn. When it is agreed: how does it reach the implementer (link the implementer opens with the Artifact tool's `read`, or exported frames)? The plan's macOS section and Strings are then updated to the board and this question closes.

## 13. PR

**Title:** S15-03 — The Source page shows the date and the Private switch
**Body:** objective in one sentence; what shipped against US-1 – US-4; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-15/prs/S15-03-source-page-date-and-private-switch.md`.
