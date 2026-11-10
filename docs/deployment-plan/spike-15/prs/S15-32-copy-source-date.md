# S15-32 — Copy Source can carry the date

**Status:** Blocked — on board S15-D13
**Spike / slice:** Spike 15 · Slice 1 — Source date, privacy, credibility
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** [S15-D13](../design/S15-D13-copy-source-dialog.md), board not agreed yet
**Board:** — (not drawn yet). When agreed: the S15-D1 Source page board, turn "S15-D13 · Copy Source", read with the Artifact tool's `read` on its link. §6 macOS describes each frame in words from the brief; the board's decisions, where they differ, are copied here when it is agreed
**Depends on:** S15-02 (landed [#762](https://github.com/mendahu/provenencia/pull/762)), S15-D13
**Blocks:** —
**Branch:** `s15-32-copy-source-date`
**Version bump:** minor, 0.7.0 → **0.8.0** (if `main`'s `VERSION` is no longer 0.7.0 when you rebase, the next minor after it): `VERSION`, `core/version.go`, and all four `MARKETING_VERSION` lines in `macos/Provenencia.xcodeproj/project.pbxproj` — a new choice on the Copy form and a new field on `CopySourceRequest` that old clients ignore ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-date-value`](../../../../.agents/skills/add-date-value/SKILL.md), [`add-ffi-handler`](../../../../.agents/skills/add-ffi-handler/SKILL.md), [`add-audit-scope`](../../../../.agents/skills/add-audit-scope/SKILL.md), [`add-ui-component`](../../../../.agents/skills/add-ui-component/SKILL.md), [`add-localized-string`](../../../../.agents/skills/add-localized-string/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md)
**Rules that bind:** `.cursor/rules/go-tests.mdc`, `.cursor/rules/date-values.mdc`, `.cursor/rules/macos-client.mdc`, `.cursor/rules/versioning.mdc`
**Context:** working rules "The Source date is one optional DateValue on `sources`" and "Not public is a stored switch and nothing else" ([`../deployment-plan.md`](../deployment-plan.md#working-rules)); [`../source-provenance-date.md`](../source-provenance-date.md); [`../../../source-layer-data-model.md` § 4.2](../../../source-layer-data-model.md#42-copy-source); S15-02's [plan](S15-02-source-date-and-public-flag.md) D-7, D-14 and **Next adds**; S15-04's [plan](S15-04-sources-list-date.md) D-14 (the date's one name is "Published")
**Planned:** 2026-11-10 against `main` @ `5a10ccb`
**Review:** [`S15-32-copy-source-date.review.md`](S15-32-copy-source-date.review.md) — Fits with changes / Not ready (on the board only), 1 round; every finding Fixed or answered; nothing open but the board
**Runs on:** macOS. STEP-2's `./scripts/generate-proto.sh` needs `protoc`, `protoc-gen-go` and `protoc-gen-swift`; STEP-3 – STEP-5 and `scripts/check-pr.sh` need Xcode ([`README.md`](../../../../README.md)). Without them, stop: never hand-edit the generated `engine.pb.go` or `engine.pb.swift`

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

A researcher copying a Source chooses whether the copy keeps the original's published date. The Copy form gains a fourth part tick, **Published date**, whose hint is the date itself; it starts ticked when the original is dated, and is disabled with "None on this Source" when it is not. Ticked, the copy gets a date equal to the original's in a `date_values` row of its own, so editing one Source's date later never touches the other. This closes the slice-1 Check line "a copied Source keeps the date when its tick is on" (user, 2026-11-09).

**Why now.** S15-02 to S15-04 landed: researchers date their Sources, and every copy today starts undated, so a sibling page of a census loses the date it shares.

**What stays.** The not-public flag is never carried: a copy always starts public (working rule "Not public is a stored switch and nothing else"). The other three parts, the Artifact checklist, the form's states, its width and its buttons are unchanged. Notes, Subjects, Citations, Observations, graphs and the cover are still never copied.

## 2. User stories

**US-1.** As a researcher, when I copy a dated Source, I keep its published date on the copy unless I untick it, so that a sibling page does not need its date typed again.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | Open Copy Source on a Source dated `4 Jul 1881` | Among the parts, a ticked "Published date" with the hint "4 Jul 1881" (CS-1, CS-2) |
| AC-1.2 | Copy with the tick on | The copy opens on its page dated `4 Jul 1881`; its `date_value_id` differs from the original's; the Sources list shows "Published 4 Jul 1881" on the copy's row |
| AC-1.3 | Copy with the tick off | The copy is undated |
| AC-1.4 | After AC-1.2, change the copy's date to `11 Jul 1881` on its page | The original still reads `4 Jul 1881` |
| AC-1.5 | The original's date is the phrase "in the reign of Victoria" | The hint reads the phrase; copied, the copy's date is that phrase, and it sorts with the undated as the original does |
| AC-1.6 | The original is undated | The tick is unticked and disabled, hint "None on this Source"; the copy is undated (CS-3) |
| AC-1.7 | Untick, then tick Published date again | The form reads changed after the untick and unchanged after the re-tick (CS-5) |
| AC-1.8 | A dated original that is not public, copied with every tick on | The copy is dated and public |

**US-2.** As the engine, a copy's date is one audited part of the one `copy_source` write.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | `CopySource` with `copy_date: true` on a dated original | One revision, action `copy_source`, holds the copy's `source` create, a `date_value` create for the new row, and a `source` update setting `date_value_id`; the response's `Source` carries `date` and `date_sort_key` |
| AC-2.2 | `copy_date: true` on an undated original, or on one whose `date_value_id` points at a missing row | The copy is undated; no error; no `date_value` change |
| AC-2.3 | `copy_date: false` on a dated original | No `date_value` row is created; the response's `Source` has no date |

## 3. Scope

| In | Covered by |
| --- | --- |
| `sourcecopy` reads the original's date into the draft and, when chosen, gives the copy its own row through `sources.SetDate` | STEP-1, AC-2.1 – AC-2.3, AC-1.4 |
| `CopySourceRequest.copy_date` and the handler passing it | STEP-2, AC-2.1 |
| `GenealogyStore.copySource(copyDate:)` in the protocol, `GoStore`, `FakeStore` and test stubs | STEP-3 |
| `SourceCopyModel`: the date choice, its availability and its hint text | STEP-4, AC-1.1, AC-1.6, AC-1.7 |
| The Published date tick on the form (CS-1 – CS-6, CS-W1) | STEP-5, AC-1.1 – AC-1.6 |
| Docs: § 4.2 of the Source layer model, the owner-check idea's trigger line | STEP-6 |

| Out | Home |
| --- | --- |
| Carrying the not-public flag | Refused (working rule "Not public is a stored switch and nothing else") |
| A date field or picker on the Copy form | Not planned (S15-D13 § 2.1: the copy's date is changed on its page) |
| A date on the Add Source dialog | Not planned |
| The value-row owner check in `datevalues.PutTx` | [`docs/ideas/value-row-owner-check.md`](../../../ideas/value-row-owner-check.md) (D-6) |
| A date on `sources.CreateInput` | Not planned (D-2) |
| `SourceCopyModel` routing `sources.not_found` through `SourcePageContext.fail` | [`docs/ideas/structural-debt.md`](../../../ideas/structural-debt.md) row "S15-03 plan": the model is also hosted from the Sources list, which has no `SourcePageContext`, so the route needs its own design; that row gains "S15-32 left it: two hosts" (STEP-6) |
| Renaming `SourcesRowDate` (its callers are now a row, the page header and the Copy form) | [`docs/ideas/structural-debt.md`](../../../ideas/structural-debt.md), a new row (STEP-6) |
| A `SourceCopyParts` value in place of the Bool parameters of `copySource` | Not planned (no further Copy part is named in the map) |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Copy engine | `core/database/sourcecopy/sourcecopy.go` `Draft`, `CopyInput`, `ReadDraft`, `CreateFromDraft`, `Copy`, `apply` | `ReadDraft` reads title, type, description, metadata and layout, credibility and Artifacts; never the date. `CreateFromDraft` returns the `sources.Create` result, so the copy is undated and `IsPublic: true` |
| Source date write | `core/database/sources/sources.go` `SetDate` | Given a Source with no date and a non-nil value: `datevalues.PutTx(tx.Tx, nil, v)` inserts a new row (create change), `UPDATE sources SET date_value_id`, a `source` update change, then re-reads the Source with `getTx` and returns it with `Date` filled. An invalid value is `datevalue.ErrInvalid` |
| Source read | `core/database/sources/sources.go` `getOne` | A `date_value_id` whose row is gone reads as undated (`Date` nil) |
| New date row | `core/database/datevalues/datevalues.go` `PutTx`, `insert` | With no `prevID` it mints a fresh UUIDv7 and ignores `v.ID` |
| Audit scope | `core/database/effects/registry.go` `"date_values"` entry | `ownedValue("date_value", "value_date_id", inbound("sources", "date_value_id"))`: a `date_value` change scopes to its owning Source. Nothing to add |
| FFI | `api/proto/engine.proto` `CopySourceRequest` (fields 1–9); `api/ffi/handlers/sources.go` `CopySource` | No date choice; builds `sourcecopy.CopyInput` from the request |
| Swift store | `macos/App/Platform/GenealogyStore.swift` `copySource(…)`; `macos/App/Platform/GoStore.swift` `copySource`; `macos/App/Platform/FakeStore.swift` `copySource` | Eight parameters after `projectDir`, no date. The fake builds a `CatalogSource` with no `date` |
| Test stubs | `macos/ProvenenciaTests/OnboardingModelTests.swift` `copySource(…)`; `macos/ProvenenciaTests/EvidenceGraphStoreTests.swift` three `copySource` calls | Conform to and call the eight-parameter signature |
| Copy model | `macos/App/Features/Sources/SourceCopyModel.swift` `SourceCopyChoices`, `SourceCopyModel` `descriptionAvailable`, `metadataAvailable`, `credibilityAvailable`, `copy()` | Three part flags, each sent as `include… && …Available` |
| Copy form | `macos/App/Features/Sources/SourceCopyDialog.swift` `SourceCopyForm` `partCheckboxes`, `partBinding` | Three `PVCheckbox` (LocalizedStringResource initializer) for Description, Metadata, Credibility; an unavailable part reads unticked and is disabled |
| Date row text | `macos/App/Features/Sources/SourcesRowDate.swift` `SourcesRowDate.line` | The one place a Source's date becomes its row text; nil when undated or its line is empty |
| Checkbox | `macos/App/DesignSystem/Components/Checkbox/PVCheckbox.swift` `init(verbatim:description:isChecked:isIndeterminate:isDisabled:isLocked:)` | Takes the label and description as Strings; `PVChecklistRows` uses it for computed text |
| Copy | `macos/App/Platform/L10n.swift` `L10n.Sources.copySource…`, `copySourcePartAbsent` = "None on this Source" | No date key |
| Tests | `core/database/sourcecopy/sourcecopy_test.go` `TestCopy`, helpers `seedObituary`, `runCopy`; `api/ffi/handlers/sources_test.go` `TestCopySource` ("copy is undated and public"); `macos/ProvenenciaTests/SourceCopyModelTests.swift` | No date-carry case; the handler case asserts the copy is undated |
| Domain doc | `docs/source-layer-data-model.md` § 4.2 | "Never copied: the date and the not-public flag (… until Copy Source carries the date, S15-32)" |

**Precedent.** The credibility part in `sourcecopy` (Spike 12): the draft holds an optional value read from the original (`Draft.Credibility *…`), `apply` drops it when unchosen, and `CreateFromDraft` writes it after `sources.Create` with the owner package's own write (`sourcecredibility.Upsert`), appending its changes. Copy that shape for the date with `sources.SetDate`. On the Swift side, copy the Credibility tick: `includeCredibility` in `SourceCopyChoices`, `credibilityAvailable`, `partBinding`.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | `sourcecopy.Draft` gains `Date *datevalue.Value` (nil when undated). `ReadDraft` copies the struct (`d := *src.Date; d.ID = nil; draft.Date = &d`), never clearing `ID` through `src.Date`. `CopyInput` gains `Date bool`; `apply` sets `d.Date = nil` when `!in.Date` | The credibility shape (precedent); a draft holds facts, not the original's row id | Reading the date inside `CreateFromDraft` — `ReadDraft` is the one read of the original |
| D-2 | `CreateFromDraft` calls `sources.SetDate(tx, userID, src.ID, d.Date)` right after `sources.Create` when `d.Date != nil`, appends its changes after the Create changes (so the order is Source, its Default graph, the date row and the Source's `date_value_id` update, then metadata, layout, credibility, Artifacts), and returns the Source `SetDate` returned. The package doc and the `CreateFromDraft` doc list the date | `SetDate` is the one owner write of a Source date (S15-02 D-7), mints the copy's own row through `PutTx`, and audits it; returning its Source puts `Date` and `date_sort_key` on the response so the list row shows the date without a reread | Adding `Date` to `sources.CreateInput` — a second path that writes a Source date beside `SetDate`, for one caller |
| D-3 | `CopySourceRequest` gains `bool copy_date = 10;` with the comment `// the original's date, as the copy's own`. The handler passes `req.GetCopyDate()` to `CopyInput.Date`. `copy_date: true` on an undated original is not an error | The same contract as `copy_description` on an original with no description | A separate RPC to date the copy after copying — two writes, two revisions |
| D-4 | Swift: `SourceCopyChoices.includeDate = true`; `SourceCopyModel.dateLine: String?` *(new)* = `SourcesRowDate.line(workspace?.source.date)`; `dateAvailable: Bool` *(new)* = `dateLine != nil`; `copy()` sends `copyDate: choices.includeDate && dateAvailable` | Availability reads the freshly loaded workspace, as the other three parts do; one helper decides "dated" for the list, the page and this form, so a date the researcher cannot see is not carried silently | Reading `original.date` — the snapshot the dialog opened with, stale after an edit elsewhere |
| D-5 | `SourceCopyForm` gains one private `partTick(_ label: LocalizedStringResource, hint: String, keyPath: WritableKeyPath<SourceCopyChoices, Bool>, available: Bool, id: String) -> some View` *(new)* that builds `PVCheckbox(verbatim: L10n.string(label), description: available ? hint : L10n.string(L10n.Sources.copySourcePartAbsent), isChecked: partBinding(keyPath, available: available), isDisabled: !available \|\| model.isSaving).accessibilityIdentifier(id)`. All four parts go through it: Published date (`L10n.Sources.copySourceDate`, hint `model.dateLine ?? ""`, `\.includeDate`, `model.dateAvailable`, `sources.copy.date`) first, then Description, Metadata, Credibility with their existing hints (`L10n.string(…Hint)`) and identifiers, unless the agreed board places the date elsewhere (CS-1). The identifier and the L10n key both read `sources.copy.date`; they are separate namespaces | The rule "an unavailable part reads unticked, disabled, with None on this Source" lives once instead of four times (review M-1); the date hint is a value, so every part takes the `verbatim:` initializer, as `PVChecklistRows` does | A fourth hand-copied block; a `PVCheckbox` initializer mixing a resource label and a String description — a kit change for one caller |
| D-6 | The value-row owner check stays in its idea file. Its trigger line is reworded so it no longer names S15-32 | The copy mints its own row (`PutTx` with no `prevID`) and never edits a row in place, so the crafted shared-row case the idea guards is not reached by this PR; a Source is not a new kind of date owner | Adding the check here — a change to `datevalues.PutTx` and `namevalues.PutTx` that this PR does not exercise |
| D-7 | The Swift `FakeStore.copySource` copies `original.date` and `original.dateSortKey` onto the copy when `copyDate` is true | The fake mirrors what the wire returns, as it does for description and credibility; it does not compute a sort key | Leaving the fake undated — the model tests could not see the carried date |

**Rules this plan bends.** None.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `sourcecopy.Draft.Date`, `CopyInput.Date` *(new)* | field | `Draft.Credibility`, `CopyInput.Credibility` | Extends the same structs in the same shape |
| `CopySourceRequest.copy_date` *(new)* | proto field | `copy_credibility` | Same message, next field |
| `copySource(…, copyDate:)` *(new parameter)* | Swift store parameter | `copyCredibility:` | Same method |
| `SourceCopyChoices.includeDate`, `SourceCopyModel.dateLine`, `dateAvailable` *(new)* | property | `includeCredibility`, `credibilityAvailable`; `SourcesRowDate.line` | Reuses `SourcesRowDate.line`; no second date-to-text rule |
| `SourceCopyForm.partTick` *(new)* | private func | the three hand-copied blocks in `partCheckboxes` | Replaces them (D-5) |
| `L10n.Sources.copySourceDate` *(new)* | L10n key | `copySourceCredibility`; `L10n.Sources.publishedLabel` ("Published") | The tick names a part ("Published date"), not a row label; one key in the copy namespace beside its siblings |

Searched: `SetDate` callers (`api/ffi/handlers/sources.go` `SetSourceDate` only), every `PVCheckbox(verbatim:` caller, `SourcesRowDate` callers, `copy_` fields in `engine.proto`.

**Builds on.** `sources.SetDate` and `datevalues.PutTx` (S15-02 D-7) for the owned row and its audit; the `date_values` effect's inbound `sources.date_value_id` for the scope; the Copy form's part-tick pattern; `SourcesRowDate` (S15-04).
**Deletes.** The three hand-copied `PVCheckbox` blocks in `SourceCopyForm.partCheckboxes` (D-5). The handler test case "copy is undated and public" becomes "copy carries the date when chosen and stays public" (T-4). The sentence in `docs/source-layer-data-model.md` § 4.2 that names the date as never copied. Nothing else becomes dead.
**Next adds.** S15-08 adds Artifact metadata to `sourcecopy.DraftArtifact`, not a new part or tick. S15-07 re-keys the `sourcemetadata` calls in `ReadDraft` and `CreateFromDraft`, beside the date lines this PR adds; whichever lands second rebases over the other.

### Reach

| New path or state | Destination | Checked: resets on the new id · handles a missing referent · shows the state |
| --- | --- | --- |
| A copy that is dated at birth | Source page header (`SourcePageIdentityHeader`), Sources list row, sort by published date | All three already render any dated Source from `CatalogSource.date` and `dateSortKey` (S15-03, S15-04); the copy reaches the list through `session.apply(effects, .createdSource(copy))` with its date |
| An original whose date row is missing | `ReadDraft` → `sources.GetTx` → `getOne` | Reads as undated; nothing carried (AC-2.2, T-2) |
| The original's date changed or cleared elsewhere while the dialog is open | `ReadDraft` at write time | The engine copies the date as it is when Copy runs; a cleared date copies nothing (AC-2.2). The hint may show the earlier date until the workspace refetches, as the other parts' availability does today |

### Scale

At the map's Scale: Copy Source copies one Source. The date adds one `date_values` lookup (inside `GetTx`, already made), one insert, one `UPDATE sources` and one re-read of the copy (`getTx`) inside the one write. No list or loop changes. Invalidation is unchanged: `copy_source` already reports the new Source's scope and Swift applies `.createdSource(copy)`.

### Edges

- **Assumptions.** `SetDate` assumes the Source exists in the transaction: it does (created on the line before). `PutTx` with no `prevID` mints a new id and ignores `v.ID` (`insert`); D-1 still clears `ID` so the draft never holds the original's row id. Two changes for one `source` entity in one revision (create, then update): `audit_changes` has no per-entity uniqueness (`core/database/migrations/000003.sql`), and the S15-02 page write already records a `source` update and a `date_value` create together. T-1 asserts the revision holds both.
- **Neighbors.** Ticked on an undated original → undated, no error (T-2). Unticked on a dated original → undated, no `date_value` row created (T-3). A phrase-only date → copied verbatim, no sort key (T-1 case, AC-1.5). A range or qualified date (`bet 1834 and 1840`, `abt 1895`) → copied with every field (T-1 compares the whole `datevalue.Value` less `ID`).
- **Removed inputs.** The original deleted before Copy runs → `sources.ErrNotFound` from `ReadDraft` as today. The original's date cleared before Copy runs → nothing carried (AC-2.2). After the copy, the two rows are independent: editing or clearing one Source's date leaves the other's row (T-1's last case).
- **Real data.** None — no table- or seed-driven rule.
- **Untrusted size.** None — one date value per Source; `datevalue.Validate` (inside `PutTx`) bounds its fields.

### Schema

None — no catalog change. `sources.date_value_id` and `date_values` exist since S15-02.

### Go core

| Package | Change |
| --- | --- |
| `core/database/sourcecopy` | `Draft.Date *datevalue.Value` *(new)*; `CopyInput.Date bool` *(new)*; `ReadDraft` fills `Date` from `src.Date` with `ID` nil (D-1); `apply` clears it when unchosen; `CreateFromDraft` calls `sources.SetDate` after `sources.Create` and returns its Source (D-2); the package doc's list gains "the date" |

Errors: none new. A stored date that fails `datevalue.Validate` (a crafted catalog) returns `datevalue.ErrInvalid` and the copy is not written.

### FFI

`api/proto/engine.proto`: `CopySourceRequest` gains `bool copy_date = 10;` (D-3); the message comment's "never copied" list is unchanged (the date was never in it). Handler `api/ffi/handlers/sources.go` `CopySource` sets `Date: req.GetCopyDate()`. Regenerate with `./scripts/generate-proto.sh`.

### macOS

From the S15-D13 inventory. The board is not agreed yet; the frames, in words from the brief:

- **Frame 1a / 1b** (Regular, Wide): the dialog over the Source page, 520pt, centered at both widths. Under the title and type fields and the divider, the "COPY FROM THE ORIGINAL" heading, then four ticks: Published date (hint "4 Jul 1881"), Description, Metadata, Credibility, all ticked; then the Artifact checklist.
- **Frame 2:** an undated original: Published date unticked, disabled, hint "None on this Source".
- **Frame 3:** a long phrase date: the hint wraps; the dialog keeps its width.
- **Frame 4:** frame 1a with Published date unticked; Copy enabled.

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Copy Source dialog | `pvFormDialog` | `macos/App/Features/Sources/SourceCopyDialog.swift` `SourceCopyDialogModifier` | Unchanged; width 520 at both widths (CS-W1) |
| Copy form body | Snowflake | `SourceCopyForm` `partCheckboxes`, `partTick` *(new)* | Four one-line `partTick` calls (D-5) |
| Published date tick | `PVCheckbox` `init(verbatim:…)` | `macos/App/DesignSystem/Components/Checkbox/PVCheckbox.swift` | Label "Published date", hint the date line or "None on this Source" |
| Date text | `SourcesRowDate.line` | `macos/App/Features/Sources/SourcesRowDate.swift` | Through `SourceCopyModel.dateLine` (D-4) |
| Parts load area, Artifact checklist, save Callout | `SourceFormLoadArea`, `PVArtifactChecklist`, `PVCallout` | as today | Unchanged |

States: loading and read-failed hide all four ticks (inside `SourceFormLoadArea`); saving disables all four; the dirty-close confirm counts the date tick through `isDirty` (`choices != prefill`). Accessibility: `PVCheckbox` speaks label, description and state: "Published date, 4 Jul 1881, checkbox, checked" (CS-6); Tab order follows the visual order.

### Strings

L10n keys *(new)*: `L10n.Sources.copySourceDate` = "Published date", key `sources.copy.date`, comment "Copy Source checkbox: give the copy the original's published date". Reused: `L10n.Sources.copySourcePartAbsent` ("None on this Source"). Existing copy that names the old scope: `docs/source-layer-data-model.md` § 4.2 "Never copied: the date and the not-public flag" (STEP-6). `GenealogyStore.copySource`'s doc comment lists the parts and gains "the date".

## 7. Steps

**STEP-1 — The copy engine carries the date when chosen**
Files: `core/database/sourcecopy/sourcecopy.go`, `core/database/sourcecopy/sourcecopy_test.go`. Skill: `add-date-value`, `add-audit-scope`.
Do: D-1 and D-2. Add T-1 – T-3 as cases of `TestCopy`. Read a revision's changes with a test helper `revisionChanges(t *testing.T, c *database.Catalog) []rowchange.Change` *(new)* in `sourcecopy_test.go` that selects `entity_type, entity_id, action, changes_json` from `audit_changes` for the latest `audit_transactions` row, as `core/database/audit/scopes_effects_test.go` `loadRevisionChanges` does; `runCopy` keeps its signature. For T-2's missing row, turn foreign keys off on one connection and delete the `date_values` row, as `core/database/sources/sources_test.go` does near `PRAGMA foreign_keys=OFF`. No existing `TestCopy` case asserts the full change list.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/sourcecopy/ ./core/database/sources/` passes.

**STEP-2 — `CopySourceRequest` takes `copy_date`**
Files: `api/proto/engine.proto`, generated `api/proto/engine/engine.pb.go` and `macos/App/Platform/Generated/engine.pb.swift`, `api/ffi/handlers/sources.go`, `api/ffi/handlers/sources_test.go`. Skill: `add-ffi-handler`.
Do: D-3; `./scripts/generate-proto.sh`; T-4, T-5 in `TestCopySource`.
Check: `CGO_ENABLED=1 go test -tags fts5 ./api/...` passes.

**STEP-3 — The Swift store passes the choice**
Files: `macos/App/Platform/GenealogyStore.swift`, `macos/App/Platform/GoStore.swift`, `macos/App/Platform/FakeStore.swift`, `macos/ProvenenciaTests/OnboardingModelTests.swift`, `macos/ProvenenciaTests/EvidenceGraphStoreTests.swift`. Skill: `add-swift-test`.
Do: add `copyDate: Bool` after `copyCredibility:` in the protocol, `GoStore` (`req.copyDate = copyDate`), `FakeStore` (D-7) and the stub; pass `copyDate: false` at the three `EvidenceGraphStoreTests` call sites; add "the date" to the protocol's doc comment.
Check: the app and test targets build (`xcodebuild build-for-testing` as `scripts/check-pr.sh` runs it).

**STEP-4 — The Copy model knows the date part**
Files: `macos/App/Features/Sources/SourceCopyModel.swift`, `macos/ProvenenciaTests/SourceCopyModelTests.swift`. Skill: `add-swift-test`.
Do: D-4. Add T-6 – T-9; give the test fixture `original()` a `date: CatalogDateValueInput? = nil` parameter. The dated fixture is `CatalogDateValueInput(kind: "point", startYear: 1881, startMonth: 7, startDay: 4)`, as `SourcesModelTests` builds dates. `FakeStore.getSourceWorkspace` returns the stored `CatalogSource` (through `enrichCoverFields`, which keeps `date`).
Check: `xcodebuild test -only-testing:ProvenenciaTests/SourceCopyModelTests` passes (scheme and destination as `scripts/check-pr.sh` uses).

**STEP-5 — The Published date tick on the form**
Files: `macos/App/Features/Sources/SourceCopyDialog.swift`, `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`. Skill: `add-ui-component`, `add-localized-string`.
Do: D-5 (the `partTick` helper and all four parts on it) and the key under **Strings**. Match the agreed board's frames 1a – 4.
Check: `python3 scripts/check-localizable-xcstrings.py` passes; the manual check (§ 10) steps 1–3.

**STEP-6 — Docs**
Files: `docs/source-layer-data-model.md` § 4.2 (the date moves from "Never copied" to "Copied, each only when chosen": "the date, as a `date_values` row of the copy's own"; "Never copied" keeps "the not-public flag (the copy always starts public)"), `docs/ideas/value-row-owner-check.md` (D-6: the trigger line reads "Touch it when `PutTx` next changes, or with the next new kind of owner of a date row."), `docs/ideas/structural-debt.md` (the "S15-03 plan" row gains "S15-32 left `SourceCopyModel` as it was: it has two hosts and the Sources list has no `SourcePageContext`"; a new row `| — | S15-32 plan | SourcesRowDate now serves a row, the Source page header and the Copy form; rename it for what it does (a Source's date as text). |`), `docs/deployment-plan/spike-15/completed.md` (a paragraph at the top in the shape of S15-04's: what the researcher sees, the engine change, "Product 0.8.0."), `docs/deployment-plan/spike-15/deployment-plan.md` (map row Status `Landed #N`), this plan (Status, a Landed note), and the brief archived as S15-D2 was: `git mv` it to `design/archive/`, add one `../` to each of its relative links, move its row from **Scheduled** to **Archived** in `design/README.md` (the S15-D2 row is the model: brief link into `archive/`, view "Copy Source dialog: a Published date tick", the agreed board link and date, built by "S15-32 (#N)"), and repoint its links in `deployment-plan.md` and this plan to `design/archive/`. Version files per the header.
Do: per the Definition of done.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-15/prs/S15-32-copy-source-date.md` reports no errors; `scripts/check-pr.sh` prints `check-pr: PASS`.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period ("Let Copy Source carry the original's published date.").

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/database/sourcecopy/sourcecopy_test.go` `TestCopy` case "carries the date as its own row" *(new)* | Original dated `bet 1834 and 1840` (and a sub-case for the phrase "in the reign of Victoria"), `Date: true` → copy's `Date` equals the original's in every field but `ID`; `DateValueID` differs; the revision holds a `date_value` create and a `source` update with `date_value_id`; then `sources.SetDate` on the copy to `1841` leaves the original at `bet 1834 and 1840` | AC-1.2, AC-1.4, AC-1.5, AC-2.1, D-1, D-2 |
| T-2 | Go | `TestCopy` case "date chosen on an undated original" *(new)* | Undated original, and an original whose `date_value_id` row is deleted with foreign keys off, `Date: true` → copy undated, no error, no `date_value` change | AC-2.2 |
| T-3 | Go | `TestCopy` case "date not chosen" *(new)* | Dated original, `Date: false` → copy undated; `date_values` row count unchanged | AC-1.3, AC-2.3 |
| T-4 | Go | `api/ffi/handlers/sources_test.go` `TestCopySource` case "copy carries the date when chosen and stays public" (replaces "copy is undated and public") | Original dated `12 Mar 1880` and not public, `CopyDate: true` → response `Source.Date` equals the date, `DateSortKey` set, `IsPublic` true; audit action `copy_source` | AC-1.8, AC-2.1, D-3 |
| T-5 | Go | `TestCopySource` case "copy without the date is undated" *(new)* | Same original, `CopyDate: false` → `Date` nil, `DateSortKey` nil, `IsPublic` true | AC-2.3 |
| T-6 | Swift | `macos/ProvenenciaTests/SourceCopyModelTests.swift` `dateTickIsOnWhenDatedAndCarried` *(new)* | Original dated 4 Jul 1881 → `dateAvailable`, `dateLine == SourcesRowDate.line(date)` (not a literal: the line follows the test machine's locale), `choices.includeDate`; `copy()` → the copy's `date` equals the original's | AC-1.1, AC-1.2, D-4, D-7 |
| T-7 | Swift | `SourceCopyModelTests` `untickedDateIsNotCarried` *(new)* | Dated original, `includeDate = false` → the copy's `date` is nil | AC-1.3 |
| T-8 | Swift | `SourceCopyModelTests` `absentPartsAreUnavailableAndNotSent` (extended) | Undated original → `!dateAvailable`, `dateLine == nil`; the copy's `date` is nil | AC-1.6 |
| T-9 | Swift | `SourceCopyModelTests` `isDirtyTracksChanges` (extended) | `includeDate = false` → dirty; back to true → clean | AC-1.7 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `TestCopySource`; `EvidenceGraphStoreTests`; `SourceCopyModelTests` (the `partTick` move).

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left
- [ ] Sweep: each renamed or deleted name `rg`'d by its old, qualified name across `.agents/`, `.cursor/rules/`, `docs/*.md` and `L10n.swift` with no hit left; `scripts/check-guidance-names.py` passes (check-pr runs it); `rg "until Copy Source carries the date"` has no hit
- [ ] Every row of **Rules this plan bends** has the user's answer, and its rule is updated
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together)
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated
- [ ] Domain doc updated: docs/source-layer-data-model.md § 4.2; docs/ideas/value-row-owner-check.md trigger line
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

1. Open a project with a Newspaper Source "Ottawa County Herald, 4 July issue" dated `4 Jul 1881`, with a description, credibility and two Artifacts. On its page, choose Copy Source. See Published date ticked with the hint "4 Jul 1881", at Regular and with the window stretched past 1600pt of pane (the dialog stays 520pt and centered).
2. Copy. The copy's page header shows the date; the Sources list row reads "Published 4 Jul 1881". Change the copy's date to `11 Jul 1881`; reopen the original: still `4 Jul 1881`.
3. Copy the original again, untick Published date, Copy: the copy is undated. Copy an undated Source: the tick is disabled with "None on this Source".
4. Mark the original Private, copy it with every tick on: the copy's Private switch is off.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| The board moves the tick or changes its copy | D-5 follows the board's place; copy changes go through **Strings** after the board is agreed |

**Stop and report if:**

- `sources.SetDate` no longer mints a new row for a Source with no date (it edits or shares a row instead).
- `audit.Record` or `effects.Resolve` refuses a create and an update of one `source` entity in one revision.
- The agreed S15-D13 board draws anything other than a fourth `PVCheckbox` (a date field, a new control, a Private tick).
- `SourceCopyModel` no longer reads its parts from the `sourceWorkspace` query.

## 12. Open questions

None. The plan is blocked only on the S15-D13 board.

## 13. PR

**Title:** S15-32 — Copy Source can carry the date
**Body:** objective in one sentence; what shipped against US-1 and US-2; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-15/prs/S15-32-copy-source-date.md`.
