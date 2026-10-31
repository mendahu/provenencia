# Ship log — S13-11 Provenance sheet

**Plan:** [`S13-11-provenance-sheet.md`](S13-11-provenance-sheet.md) · **Branch:** `s13-11-provenance-sheet` · **PR:** #650
**State:** fixing 3 — user scope change

## Round 0 — implement
Base `main` @ `474ca934`. Head `5ec67292`. Gate PASS (orchestrator run). Report: §10 manual check not run (Escape focus return, Back after each link, VoiceOver wording, 1600pt resize, visual match to frames 6c – 6n; the Escape-focus stop condition is unverified). Deviations: (1) STEP-3 – STEP-7 in one commit; (2) `ReconciliationRecordGroup` gained `init(owner:records:values:)` with `values` defaulting to one headless run; (3) `actionColumnWidth` moved to `ReconciliationRecordRows`; (4) the sheet's Read as, Outcome and Record columns have fixed widths (140/190/112pt Regular, 220/280pt Wide) so the headings line up; (5) a record is not one VoiceOver element: the citation line and its links are separate controls; (6) `S13-D2-provenance-sheet.review.md` stays in `design/`; (7) PR body attribution added with `gh pr edit`.

## Round 1 — review
Reviewers: CON (Sonnet), MNT, UX, SEC+PRF. Skipped: none. UX for `macos/App/**` and strings; SEC+PRF for `api/ffi/**`, `core/locator` locator code, and `macos/App/Features/**` models.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| CON-1 | CON | major | Plan Status says Landed but no Landed note lists the deviations | Fix now → F-1 | verified: `rg "Landed"` on the plan matches only Status and the DoD line; the implementer's report said it was written |
| CON-2, UX-3 | CON, UX | major | VoiceOver: a record is several elements, "Source" link names no Source, sheet title is not a heading | Fix now → F-2 | verified: plan §6 Accessibility line "each record is one element with its links separate", "the sheet title … are headings"; `sheetRow` combines only the Record cell; deviation (5) has no Decision behind it |
| UX-1 | UX | major | Selectable text inside the citation-line Button may swallow the click (AC-2.2's only mouse route) | To user → pending design | User 2026-10-31: selection only served copy; wants a copy button to the right of the line instead, board turn 7 (frames 7a – 7k) gives the proposed set CPY-1 – CPY-3 → F-10. D-18 asks for both "opens the Citation" and "text stays selectable"; `CitationLineButton` puts `.textSelection(.enabled)` inside `Button(.plain)`; not observed in the app |
| UX-2, MNT-2 | UX, MNT | major | Citation line: hand-rolled hover, no `PVHoverEffect`, no kit focus ring on the record's first Tab stop | Fix now → F-3 | verified `ReconciliationRecordRows.swift:248-271`: `@State hovering` + `.onHover`, no `.focused`/`.pvFocusRing`; precedent `SourcesGraphButton.swift:245`, gallery tile `ConclusionDetailPage.swift:620-622` |
| UX-4, MNT-4 | UX, MNT | major | Sheet header shows one value on a multi-valued row; header and value heading copy the row's value rendering | Fix now → F-4 | verified: `ProvenanceSheet.header` branches only on `lead`/`emptyText`; PRV-4 "a multi-valued row lists its values" |
| MNT-1 | MNT | major | `ReconciliationRecordGroup` stores `records` and `values` independently | Fix now → F-5 | verified `ReconciliationRecord.swift:112-120`; D-13 "stores `values`" is kept, `records` becomes computed from it |
| MNT-3 | MNT | major | Kept-or-folded rule copied inline in `conclusiondetails.finish` | Fix now → F-6 | same-pass cleanup: this PR adds the copy beside `pipeline.go:198`; `Reason.Displayed()`/`InGallery()` precedent |
| SEC-1 | SEC | minor | Locator page above int32 range is narrowed at the FFI and shows a wrong page | Fix now → F-7 | verified: `validatePage` checks only `>= 1`; `int32(o.ArtifactPage)` in `conclusion_detail.go`; code this PR adds |
| MNT-6 | MNT | minor | `locator.Page` adds a fourth decode-then-`validateDoc` copy | Fix now → F-7 | same-pass cleanup: this PR adds the fourth copy |
| UX-5 | UX | minor | Return does not close the sheet; precedent `PVDeleteImpact` Done has `.defaultAction` | Fix now → F-8 | minor in a file this PR adds; S10-08 Return rule |
| MNT-5 | MNT | minor | `isWide: Bool` threaded through three types where `PVWidthClass` exists | Fix now → F-8 | minor in files this PR adds; no Decision names `isWide` |
| UX-7 | UX | minor | `.single` row's visible summary reads "single, 1 Source · …" (VoiceOver shorthand) | To user → Fix now → F-9 | User 2026-10-31 chose a visible state phrase "from %@" (supersedes D-14's `spokenState` in the visible summary). Follows D-14 (`spokenState` in the summary) but reads as screen-reader text |
| UX-6 | UX | minor | Nothing in code backs Escape returning focus to the control | To user (manual check 1) | §11 stop condition; settled only by running the app |

Nits (logged): `CON-3 · ReconciledValueRowModel.swift · blank line before listedValues doc` · `MNT-7 · ReconciliationRecord.swift:95,106 · ValueGroup Identifiable id unused` · `MNT-8 · ReconciliationRecord.swift:93-94 · doc says only the sheet reads values` · `UX-8 · ReconciledValueRowModel.swift:210-224 · shown owner with no records still counted` · `UX-9 · ProvenanceSheet.swift:163 · 560pt cap vs window height (D-12)`

### Fix plan 1

**F-1 — Write the plan's Landed note** (from CON-1)
Files: `docs/deployment-plan/spike-13/prs/S13-11-provenance-sheet.md`.
Do: add a **Landed** note under the header listing every difference from the plan: round-0 deviations (1)–(4) and (6) from this ship log's Round 0, and every decision this fix plan supersedes (F-9 over D-14's visible summary; F-2 restores §6's one-element record, so deviation (5) is gone). Fix-round-2 additions get appended later.
Check: `rg -n "Landed" <plan>` shows the note; `scripts/check-plan-docs.py` 0 errors.

**F-2 — VoiceOver: one element per record, named Source link, heading title** (from CON-2, UX-3)
Files: `macos/App/Features/Conclusions/ReconciliationRecordRows.swift`, `ProvenanceSheet.swift`, `L10n.swift`, `Localizable.xcstrings`. Skill: `add-localized-string`.
Failing test first: none feasible for VoiceOver tree; add a model test if you add a spoken-label builder on `ReconciliationRecord` (e.g. `ProvenanceSheetTests/aRecordSpeaksAsOneLine`: Read as, citation line, outcome phrase, Subject ref, in that order).
Do: in `.sheet` mode give each record one combined, labelled accessibility element (Read as, citation line, outcome phrase, Record refs) with the citation control and the "Source" / "Observation" links staying separate controls after it (plan §6: "each record reads as one element … and its links as separate controls"). Give the Source link a spoken label via new key `L10n.Conclusions.provenanceSourceLinkSpoken(_ title:)` "Source %@" using the record's Source title. Make the sheet title a heading: pass `Text(...).accessibilityHeading(.h1)` (or equivalent) as `PVPanel`'s title from `ProvenanceSheet`; do not change `PVPanel`.
Check: tests pass; scripts/check-pr.sh PASS.

**F-3 — Citation line on `PVHoverEffect` with the kit focus ring** (from UX-2, MNT-2)
Files: `macos/App/Features/Conclusions/ReconciliationRecordRows.swift`.
Do: replace `CitationLineButton`'s `@State hovering` + `.onHover` with a private `ButtonStyle` whose `makeBody` wraps the look in `PVHoverEffect(isPressed:hoverAnimation:)` (precedent `Features/Sources/SourcesGraphButton.swift:245`), keeping D-18's ink rule (primary when counted, else secondary; `textLinkHover` on hover; underline at 35%). Add `@FocusState` + `.focused` + `.pvFocusRing(focused, cornerRadius: PVRadius.xs)` as the gallery tile does (`ConclusionDetailPage.swift:620-622`). Leave `.textSelection` as is for now (UX-1 pending a design).
Check: scripts/check-pr.sh PASS.

**F-4 — Sheet header lists a multi-valued row's values; share the row's value rendering** (from UX-4, MNT-4)
Files: `ReconciledValueRow.swift`, `ProvenanceSheet.swift`, `ReconciliationRecordRows.swift`.
Failing test first: if the header choice lives in a model helper, a `ProvenanceSheetTests` case that a `.multiple` row's header shows its `listedValues`.
Do: add static helpers on `ReconciledValueRow` beside `font(for:)` / `badge(_:)` — one for the lead/empty value, one for a listed value line (`OtherValue` text in `font(for:)` + mono caption support). Use them in the row's own `value`, the sheet header (rule: `listedValues` stack when non-empty, else `lead`, else `emptyText`, same as the row), and `valueHeading(.value)`.
Check: tests pass; scripts/check-pr.sh PASS.

**F-5 — `ReconciliationRecordGroup` stores `values` only** (from MNT-1)
Files: `ReconciliationRecord.swift`, `ReconciliationRecordRows.swift`, `ReconciledValueRowModel.swift` and any caller.
Do: make `records` a computed `values.flatMap(\.records)`; keep `init(owner:records:)` building one headless run. `.why` keeps drawing from `records`. Also fix nits MNT-7 (drop unused `Identifiable` or use its id in the `ForEach`) and MNT-8 (doc comment).
Check: existing T-3/T-5/T-12 tests pass; scripts/check-pr.sh PASS.

**F-6 — Kept-or-folded rule on `autoreconcile.Reason`** (from MNT-3)
Files: `core/autoreconcile/autoreconcile.go`, `core/autoreconcile/pipeline.go`, `core/database/conclusiondetails/conclusiondetails.go`.
Do: add `func (r Reason) IntoValue() bool` (kept or folded) beside `Displayed()` / `InGallery()`, with a table test in the package's existing test file; call it from `pipeline.go:198` and `finish`.
Check: `grep -rn "ReasonFolded" core api | grep -v _test` prints only the definition and `IntoValue`; go tests pass.

**F-7 — `locator.Page` bounds the page; one `parse` helper** (from SEC-1, MNT-6)
Files: `core/locator/depiction.go`, `core/locator/validate.go`, `core/locator/validate_test.go`.
Failing test first: `TestPage` row `{"type":"page","artifact_page":4294967333}` → `0, ""`.
Do: add unexported `parse(json string) (document, error)` running `decode` then `validateDoc`; `Validate`, `ValidateDepictionRegion`, `regionDocPages`, `Page` call it. In `citationPage` (so `ValidateDepictionRegion` shares it) treat a page above `math.MaxInt32` as no page.
Check: `grep -n "decode(" core/locator/*.go | grep -v _test` prints only `parse` and `decode`'s definition; go tests pass.

**F-8 — Done is the default action; pass `PVWidthClass`** (from UX-5, MNT-5)
Files: `ProvenanceSheet.swift`, `ReconciliationRecordRows.swift`, `ConclusionDetailPage.swift`.
Do: add `.keyboardShortcut(.defaultAction)` to the sheet's Done (precedent `PVDeleteImpact.swift:217-218`). Replace `isWide: Bool` with `widthClass: PVWidthClass` in `ProvenanceSheet`, `ReconciliationRecordRows`, `SheetHeadings` and their width functions. Also nit CON-3 (blank line before `listedValues` doc).
Check: scripts/check-pr.sh PASS.

**F-9 — Visible summary uses a visible state phrase** (from UX-7; user 2026-10-31)
Files: `ReconciledValueRowModel.swift`, `L10n.swift`, `Localizable.xcstrings`. Skill: `add-localized-string`.
Failing test first: extend `theSummarySaysOnlyWhatTheRowKnows` — a `.single` row's summary starts "from 1 Source", not "single,".
Do: add `L10n.Conclusions.provenanceStateSingle(_ sources:)` "from %@" and use it for `.single` in `provenanceSummary` instead of `spokenState`; other states keep their phrase. VoiceOver's `spokenState` is unchanged.
Check: tests pass; scripts/check-pr.sh PASS.

**F-10 — `PVClipboardButton` kit component; the citation line gets one and stops being selectable** (from UX-1; user 2026-10-31: copy button to the right of the line, board turn 7 frames 7a – 7k, proposed set CPY-1 – CPY-3; user 2026-10-31: promote the copy-then-check behaviour to the kit so the app reuses it)
Files: `macos/App/DesignSystem/Components/ClipboardButton/PVClipboardButton.swift` *(new)*, `macos/App/DesignSystem/Components/Icon/PVIcon.swift`, `macos/App/DesignSystem/README.md` (built-table row), `ReconciliationRecordRows.swift`, `L10n.swift`, `Localizable.xcstrings`, `project.pbxproj`, a new test file under `macos/ProvenenciaTests`. Skills: `add-ui-component` (component layer: domain-agnostic, composes `PVIconButton`), `add-localized-string`, `add-swift-test`.
Do:
- **Kit.** `PVIcon` has `.check` but no copy glyph: add `case copy = "doc.on.doc"` (Lucide `copy`, as the board's kit names it) with a one-line doc comment as its neighbours have. Add `PVClipboardButton(text: String, label: PVCopy, copiedAnnouncement: PVCopy, size: PVIconButton.Size = .sm)` — name chosen because `PVCopy` already means a component's text. It is a `PVIconButton` (neutral/ghost tone, the kit default) showing `.copy`; a click writes `text` to `NSPasteboard.general` (clear, then set string), swaps the glyph to `.check` in `PVColor.success` for 1.6s then back (a click during the 1.6s re-copies and restarts the timer; the task is cancelled on disappear), and posts a polite VoiceOver announcement of `copiedAnnouncement`. No tooltip swap (CPY-3). Keep the pasteboard write and the timing in a small testable seam (e.g. an injectable writer / a pure state type) so a test can prove it without the real pasteboard. Add the DesignSystem README built-table row (`check-guidance-names.py` requires it).
- **Use.** In `.sheet` mode, drop `.textSelection(.enabled)` from the citation line and place `PVClipboardButton(text: record.citationLine, label: L10n.Conclusions.provenanceCopyCitation /* "Copy citation" */, copiedAnnouncement: L10n.Conclusions.provenanceCitationCopied /* "Citation copied" */)` in a 26pt track 6pt to the right of the line inside the Citation column (the track comes out of that column's width); "Source" / "Observation" stay under the line, unchanged. **CPY-1:** top-aligned, centre on the first line's centre, at one, two or three lines. **CPY-2:** always shown, every record. Tab order per record: citation line → Copy citation → Source → Observation (D-19 extended). `.why` mode unchanged.
Failing test first: `PVClipboardButtonTests` — copying writes exactly the given text to the injected writer and enters the copied state; the state returns to idle after the interval; a second copy restarts it.
Plan: add to the plan's Landed note that F-10 supersedes D-18's / PRV-15's "the text stays selectable" with the copy button (board turn 7, CPY-1 – CPY-3) and adds `PVClipboardButton` to the kit at the user's request (one caller today).
Check: tests pass; `scripts/check-guidance-names.py` passes; scripts/check-pr.sh PASS.

## Round 1 — fix
Head `cef59611` (F-1 – F-9), then `710f979a` (F-10, a fresh implementer). Gate PASS (orchestrator run at `710f979a`). F-1 – F-10 done. Report deviations: F-2, F-3, F-4, F-5, F-8, F-9 in one commit; F-2 puts the combined record element on the Read-as cell (a `GridRow` cannot be one element), so the citation line is spoken in the record line and again as its control; F-4 shares one `valueContent` helper; F-10's test was written with the type, not run red first; `copiedAnnouncement` is a `String` (`AccessibilityAnnouncement.post` needs one); `size` is `PVControlSize`; added `PVIconButtonTone.success` for the green check; CPY-1 first-line centre approximated as baseline − bodySmall/3. An implementer gate run failed `promotealign` `TestLargeSourceBudgets` once under parallel load (2.68s vs 2.5s), passed on rerun. App not run.
SEC-1 (minor, SEC+PRF not sent back): **Resolved** — read `362721fb`: `citationPage` returns no page above `math.MaxInt32`; `TestPage/page_above_int32`. MNT-6 also in `362721fb` (`parse` is `decode`'s only caller).

## Round 2 — review
Reviewers: CON (Sonnet), MNT, UX — each had a Fix-now major. Not sent back: SEC+PRF (one minor, resolved above). Scope `5ec67292...710f979a`.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| CON-1, CON-2 | CON | — | Round-1 majors | Resolved | Landed note at plan line 17; `record.spoken` on the Read-as cell, Outcome/Record hidden, `aRecordSpeaksAsOneLine`; title `.accessibilityHeading(.h1)`; Source link "Source %@" |
| UX-1 – UX-4 | UX | — | Round-1 majors | Resolved | F-10 copy button, no `.textSelection` in the Button; `CitationLineButtonStyle` on `PVHoverEffect` + `pvFocusRing`; one element per record; header via `valueContent` |
| MNT-1 – MNT-4 | MNT | — | Round-1 majors | Resolved | `records` computed from `values`; `PVHoverEffect` style; `Reason.IntoValue()`; shared `valueContent` / `valueLine` |
| UX-5, MNT-5, MNT-6 | UX, MNT | minor | Round-1 minors | Resolved | `.keyboardShortcut(.defaultAction)` on Done; no `isWide` in `Features/Conclusions`; `parse` sole `decode` caller |
| CON-4 | CON | minor | Clamping in `citationPage` flips `ValidateDepictionRegion` for a page > MaxInt32 | Follow-up | Round 2 adds blockers only; needs a ~4-billion page; fix: clamp in `Page` only, or reject in `validatePage` |
| MNT-9 | MNT | minor | `ReconciliationRecord.spoken` is the third copy of the `a11yList` reduce | Follow-up | Round 2 adds blockers only; same-pass shape (`L10n.Conclusions.a11yList(_ parts:)`) — offered to the user to take on this PR |
| MNT-10 | MNT | minor | Visible-summary state rule inline in `ReconciledValueRowModel.swift:88-90` | Follow-up | belongs beside `spokenState` as `summaryState(_:)` |
| UX-6 | UX | minor | Escape focus return unbacked by code | To user | manual check 1 |

Nits (logged): `UX-10 · ReconciliationRecordRows.swift:201 · value heading gap now space5 (was space4) via shared valueLine`

## Result
Gate PASS at `710f979a` (`main` unmoved) · CI green · open: none · follow-ups: CON-4, MNT-9, MNT-10 · to user: UX-6 (manual check 1), §10 manual check not run · CON miss: none

## Scope change — user, 2026-10-31
After round 2 the user asked, for this PR, that the birth and death place rows read the event–place `location` link instead of the Place's Toponym, with each record read as its own toponym for the place. The plan records it as D-20 – D-22 and AC-1.5 (header, "Scope change"). The user asked for it on this PR, which authorizes a third review round scoped to it. Separately, the user parked loading records on demand as [`docs/ideas/provenance-on-demand.md`](../../../ideas/provenance-on-demand.md), with S13-06 PRF-1.

### Fix plan 3

**F-11 — Place rows read the event–place link** (plan D-20 – D-22, AC-1.5)
Files: `core/database/conclusionheaders/compose.go`, `core/database/conclusionheaders/placechains_test.go`, `core/database/conclusiondetails/conclusiondetails.go` and its test, `api/proto/engine.proto` (+ `./scripts/generate-proto.sh`), `api/ffi/handlers/conclusion_detail.go` and its test, `macos/App/Platform/GoStore.swift`, `FakeStore.swift`, `macos/App/Features/Conclusions/ReconciledValueRowModel.swift`, `PersonDetailContent.swift`, Swift tests. Skill: `add-ffi-handler`.
Failing tests first: `TestPersonLifeOwners` / `TestPersonDerivedRows` — a birth event linked to a Place through a `location` link: `birth.Places` owner is the link's association with the place end property and `Place` = that Place, not the Place with `PropertyToponym`. A `conclusiondetails` case: an outcome on the link's place end reads its place Subject's first positive toponym text.
Do:
1. **Go owners (D-20).** In `lifeOwners`, replace the Place owners: for each pick, `follow(g, pick.event.ID, hops.PlacesOfEvent)` and, in `eventPlaces` order, add `FieldOwner{Entity: entityOf(association node), PropertyID: l.EndpointPropertyID, Place: &HeaderPlace{…the end's Place, unfolded, toponyms(n, …)}}`. Keep one owner per association. If `Link.Association` is empty or `EndpointPropertyID` is not the place end's Property, **stop and report**. `PersonDerivedRows` loads the association nodes with the rest and copies `Place` onto `OwnerField`. Update the package doc and the `HeaderPlace.entity` proto comment ("carries its toponym records") to say place rows carry the link's records.
2. **Read as (D-21).** In `sqlOutcomes`, add the record's toponym for a Subject-valued Observation: a correlated subquery on `observations` for `subject_id = o.value_subject_id`, the toponym Property (pass the id as a bound parameter from the vocabulary, not a key string in SQL), `polarity = 'positive'`, non-empty `value_text`, ordered by `ref`, `LIMIT 1`. When the outcome's recorded value is otherwise empty, use it as the recorded text. No other value changes.
3. **Wire.** `ConclusionOwnerField.place = 4` (`HeaderPlace`), mapped in the handler, `GoStore` and `FakeStore` (the FakeStore person fixture's place owners become links with a place).
4. **Swift (D-22).** Give `CatalogOwnerField` an optional `place`. In `ReconciledValueRowModel.derived`, compare, order and ink owners by `owner.place?.entity.id ?? owner.entity.id`. `ownerLine` titles a place owner from `owner.place` through the existing `DerivedPlace` / `PlaceTitleDisplay` helpers. Its id and ref are the Place's, so the group line reads as today. The "from N Places" summary counts distinct place ids. `PersonDetailContent` passes the shown Place's id as `shownOwnerID`, as it already does.
Tests: update `lifeRowsCarryTheirOwnersRecords` and `placeOwnersEachHoldOneValueWithoutHeadings`. Add `ProvenanceSheetTests` `twoLinksToOnePlaceCountOnePlace`: the summary says one Place, and both group lines title the Place.
Check: `rg -n "PropertyToponym" core/database/conclusionheaders/compose.go` shows no life-owner use; go tests and Swift tests pass; scripts/check-pr.sh PASS. Add F-11 to the plan's Landed note.
