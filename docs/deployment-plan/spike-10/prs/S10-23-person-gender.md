# S10-23 — Sex / gender on the Person page

**Status:** Blocked — on board S10-D12
**Spike / slice:** Spike 10 · Slice 6 — Lists and the Person page
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#s10-23--sex--gender-on-the-person-page)
**Brief:** [S10-D12](../design/S10-D12-person-detail.md) (PDP-1 to PDP-6 ship here), board not agreed yet
**Board:** a Claude Design link, recorded on the brief's **Board** line when the board is agreed (answered 2026-10-15). Until then §6 macOS describes each frame in words from the brief.
**Depends on:** S10-D12 (board not agreed)
**Blocks:** S10-35 (fills the gallery frames on the same page)
**Branch:** `s10-23-person-gender`
**Version bump:** none — Spike 10 PRs leave `VERSION` at 0.2.0, and the product bump is cut when a release ships ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-ui-component`](../../../../.agents/skills/add-ui-component/SKILL.md), [`add-localized-string`](../../../../.agents/skills/add-localized-string/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md)
**Rules that bind:** `.cursor/rules/design-system-layers.mdc`
**Context:** working rules "Sex / gender styling", gotchas "Gender and sex at birth are different Properties", "Gallery and gender share a page"; [`docs/seeded-vocabulary.md`](../../../seeded-vocabulary.md) §3.7 (neutral terms) and §3.9 (gender); [`docs/design-system-layers.md`](../../../design-system-layers.md)
**Planned:** 2026-10-15 against `main` @ `daf4e08` (draft; redo §4 when the board is agreed)
**Review:** — (not reviewed yet; runs when the board is agreed)

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

The researcher can see a person's gender at a glance on the Person page, or, when no gender is recorded, their sex at birth, clearly labeled as such. A Badge under the name names the value, and the header thumb takes a soft tint for it. Unknown, empty, and disagreeing values leave the page as it is today.

**Why now.** Dogfood note "Show sex / gender on the Person detail page" (map, Slice 6). The `gender` Property and its terms `man`, `woman`, `non_binary` already ship from Install (commit `3d70139`, with S10-38 and S10-45), so this PR is the page treatment only.

**What stays.** Every Details row, including Sex at birth and Gender with their badges and Why; the name, the b. and d. lines, the ref and member count; the Event and Place pages; the engine. No Go, FFI, schema, or seed change.

## 2. User stories

**US-1.** As a researcher, when I open a Person whose gender is recorded, I see it under the name, so I know who I am looking at without reading the rows.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | Gender `woman` (one kept value), sex at birth `female` | Badge reads “Woman”; header thumb has the Woman tint (PDP-1, PDP-3). |
| AC-1.2 | Gender `man`, sex at birth `female` | Badge “Man”, Man tint. Gender wins (working rule). |
| AC-1.3 | Gender `non_binary`, no sex at birth | Badge “Non-binary”, Non-binary tint. |
| AC-1.4 | Gender is a researcher term, key `two_spirit`, label “Two-spirit” | Badge “Two-spirit”, other tint. |
| AC-1.5 | Gender has two kept values (state `mixed`), sex at birth `female` | No Badge, plain thumb. Sex at birth is not used (PDP-4). |

**US-2.** As a researcher, when only sex at birth is recorded, the page shows it and says it is sex at birth, so I never read it as gender identity.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | No gender field, sex at birth `male` | Badge “Male · at birth”; Man tint (PDP-2, PDP-3). |
| AC-2.2 | No gender field, sex at birth `female` | Badge “Female · at birth”; Woman tint. |
| AC-2.3 | No gender field, sex at birth `intersex` | Badge “Intersex · at birth”; other tint. |
| AC-2.4 | Sex at birth recorded only as `unknown` (a neutral term; the reconciler marks it `no_evidence`, so nothing is displayed) | No Badge, plain thumb. |
| AC-2.5 | Sex at birth `mixed`, no gender | No Badge, plain thumb. |
| AC-2.6 | Neither field | The page as it is today: no Badge, plain thumb, same rows. |

**US-3.** As a VoiceOver user, I hear what the Badge means.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | AC-1.1 | Badge reads “Gender: Woman” (PDP-5). |
| AC-3.2 | AC-2.1 | Badge reads “Sex at birth: Male”. |

## 3. Scope

| In | Covered by |
| --- | --- |
| Read the treatment from the detail's fields: precedence, tint, label (PDP-1, PDP-2, PDP-3, PDP-4) | STEP-1, STEP-2, AC-1.1, AC-2.1 |
| `PVThumbnail` optional tint on the filled state (PDP-3, inventory Extend) | STEP-3 |
| Header thumb takes the kind's tint through `ConclusionDetailBody` (inventory Extend) | STEP-4 |
| Badge under the name on the Person page, with its spoken label (PDP-1, PDP-2, PDP-5) | STEP-5, AC-3.1, AC-3.2 |
| Copy: “{label} · at birth”, “Gender: {label}”, “Sex at birth: {label}” | STEP-2 |
| Details rows unchanged (PDP-6) | AC-2.6, T-6 |

| Out | Home |
| --- | --- |
| Likenesses section and the earliest crop in the header (PDP-7, PDP-8, PDP-9) | S10-35 |
| Marking a portrait (PDP-10, PDP-11) | S10-39 |
| Seeding `gender` or its terms | Already landed in `3d70139`; not planned |
| Adding `gender` to catalogs created before `3d70139` | Not planned (Install data; add by hand, same as `engagement`) |
| Gender or sex on Event or Place pages, the Persons list, Promote | Not planned (brief §2.1) |
| Editing a value from the header | Not planned |
| A colour per researcher-added term, a new palette colour | Not planned (working rule) |
| Pronouns or text inferred from the value | Not planned (brief §2.1) |

## 4. Current state

Verified at `daf4e08`.

| Area | Where | Today |
| --- | --- | --- |
| Seed | `core/database/subjectvocab/registry.go` `seedProperties`, `seedBindings`, `seedTerms` | `gender` (term, unlocked on person, sort 2) with `man` / `woman` / `non_binary` already seeded. `sex_at_birth` terms: `female`, `male`, `intersex`, and neutral `indeterminate`, `unknown`. Nothing to change. |
| Neutral terms | `core/autoreconcile/modules.go` `termKey` | A neutral term is no evidence, so its outcome is `no_evidence` and it is never a displayed value. The treatment never sees `unknown` as a lead. |
| Detail payload | `macos/App/Platform/GenealogyStore.swift` `CatalogConclusionDetail`, `CatalogConclusionField` (`propertyKey`, `state`, `displayedValues`), `CatalogConclusionValue.term(id:key:label:)` | Carries the gender and sex-at-birth fields with term key and catalog label. No engine change is needed. |
| Property keys | `macos/App/Platform/ConclusionVocabulary.swift` `SeededPropertyKey` | Has `sexAtBirth`; no `gender`. |
| Person content | `macos/App/Features/Conclusions/PersonDetailContent.swift` `PersonDetailContent` | Title, ref, members, vitals, rows. No treatment. A Gender field with outcomes already shows as an ordinary row through the trailing `for field in fields` loop. |
| Person view | `macos/App/Features/Conclusions/PersonDetailView.swift` `PersonDetailView`, private `PersonDetailVitals` | The summary slot shows the b. and d. lines only. |
| Shared page | `macos/App/Features/Conclusions/ConclusionDetailPage.swift` `ConclusionDetailBody`, private `ConclusionDetailHeader` | Header builds `PVThumbnail(.init(glyph: .mark(mark), label: thumbnailLabel), size: 80)` with no tint. `sections` is the precedent for a protocol member with a default in an extension. |
| Thumb | `macos/App/DesignSystem/Components/Thumbnail/PVThumbnail.swift` `PVThumbnail.Content` | Filled states use `PVColor.surfaceInset` and `PVColor.borderDefault`; no tint. |
| Badge | `macos/App/DesignSystem/Components/Badge/PVBadge.swift` `PVBadge.init(text:tone:icon:subtle:foreground:)` | Data text with an ink override. Precedent: `EvidenceSubjectCard.swift` (ref badge with the kind's ink). |
| Value text | `macos/App/Features/Conclusions/ReconciledValueDisplay.swift` `ReconciledValueDisplay.string(for:)` | A term reads as its label, else its key. The Badge uses the same text the Details row shows. |
| Tests | `macos/ProvenenciaTests/PersonDetailTests.swift` `detail(...)`, `rowsAreNameThenLifeThenRecordedFields`; helpers `ReconciledValueDisplayTests.field/value/outcome` | Build fields by state and values. The new tests reuse them. |

**Precedent.** The S10-19 kind marks: one feature rule decides the look, the kit components stay content-agnostic. Copy `EvidenceSubjectCard`'s `PVBadge(text:tone:icon:subtle:foreground:)` for an inked badge, and `ConclusionDetailBody.sections` for a protocol member with a default. Do not copy `SubjectTypeChrome`'s use of `PVPalette`; the tints are `PVColor` tokens (D-5).

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | Swift only. The treatment is read from `CatalogConclusionDetail.fields`. | Go already sends both fields with term key and label; the treatment is presentation of reconciled values (gotcha "Gender and sex at birth are different Properties"). | A Go `PersonHeader.gender` field — a second read of data the page already holds. |
| D-2 | Precedence counts `displayedValues`, not `state`. Gender with exactly one displayed term → use it. Gender with two or more → neutral, no fallback. Gender absent or zero displayed → sex at birth: one displayed → use it; otherwise neutral. | Counts survive `concluded` and `unrecognized` states; zero displayed covers neutral-only records (§4 Neutral terms). | Switching on `ReconciledState` — misses `unrecognized`, and `concluded` would need its own case. |
| D-3 | Tint by key. Gender: `man` → man, `woman` → woman, `non_binary` → nonBinary, any other key → other. Sex at birth: `female` → woman, `male` → man, any other key → other. | User answer 2026-10-15 (map working rule). | A tint per researcher term — forbidden by the working rule. |
| D-4 | New feature snowflake `PersonGenderTreatment` in `macos/App/Features/Conclusions/PersonGenderTreatment.swift` *(new)*, pure, `Equatable`; `PersonDetailContent` holds it as `treatment: PersonGenderTreatment?` (nil = neutral). | Same shape as `PersonHeaderDisplay`: one place decides, views lay out. Testable without views. | Logic in the view — untestable; in `ReconciledValueDisplay` — shared by Event and Place, which have none of this. |
| D-5 | The four tints are existing `PVColor` tokens, named on the agreed board (fill, line, ink per tint). They live in `PersonGenderTreatment.Tint` as computed colours. No new `PVColor` or `PVPalette` entry. | User answer 2026-10-15: existing tokens, no new palette colour. | Status tones (meaning collision); the Event / Place kind inks (brief §2). |
| D-6 | `PVThumbnail.Content` gains `tint: PVThumbnail.Tint?` *(new)*, `struct Tint: Equatable { var fill: Color; var line: Color }`. In the mark and symbol states it replaces the background and border colours; in the image state only the border (for S10-35). Empty and loading ignore it. | Inventory row "Header thumb — Extend". Content-agnostic: names colours, not gender. | An overlay in `ConclusionDetailHeader` — the thumb paints its own opaque background, so a wrapper cannot tint behind the mark. |
| D-7 | `ConclusionDetailBody` gains `var thumbnailTint: PVThumbnail.Tint? { get }` with a default `nil` in the extension. `ConclusionDetailHeader` passes it into the thumb's `Content`. Event and Place change nothing. | Inventory "Shared Conclusion page — Extend"; same shape as `sections`. | A new init argument on `ConclusionDetailPage` — the value comes from content, not from the call site. |
| D-8 | The Badge is `PVBadge(text:tone: .neutral, foreground: tint ink)` in `PersonDetailVitals`, above the b. and d. lines, with `.accessibilityLabel` set to the spoken label. Not a `PVChip`. | Brief inventory: Badge, display only. | `PVChip` — a selectable control with an action. |
| D-9 | Badge text is the term text from `ReconciledValueDisplay.string(for:)`: gender → that text; sex at birth → `L10n.Conclusions.sexAtBirthBadge(text)` = “%@ · at birth”. Spoken: `L10n.Conclusions.a11yGender(text)` = “Gender: %@”, `L10n.Conclusions.a11ySexAtBirth(text)` = “Sex at birth: %@”. | One source of the label, the same one the Details row uses. | Localizing seeded gender terms in `L10n.PropertyTerm` — the Details row does not, and two texts for one value would drift. |
| D-10 | `SeededPropertyKey.gender = "gender"` *(new)*. Term keys live as private constants inside `PersonGenderTreatment`. | Only this file reads them. | A shared term-key enum — one reader. |

## 6. Design

### Codebase fit

Searched: `gender`, `sexAtBirth`, `Tint`, `tint:` under `macos/App`; `PVThumbnail(` call sites; `foreground:` on `PVBadge`.

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `PersonGenderTreatment` *(new)* | type / file | `PersonHeaderDisplay` | Different job (title text). Same pure-helper shape, beside it. |
| `PVThumbnail.Tint`, `Content.tint` *(new)* | component extend | `PVThumbnail.Content` | Extended, not copied. |
| `ConclusionDetailBody.thumbnailTint` *(new)* | protocol member | `ConclusionDetailBody.sections` | Extended the same way. |
| `SeededPropertyKey.gender` *(new)* | key | `SeededPropertyKey.sexAtBirth` | Extended. |
| `L10n.Conclusions.sexAtBirthBadge`, `a11yGender`, `a11ySexAtBirth` *(new)* | key | `L10n.Conclusions.a11yRef` | Same namespace and format shape. |
| `PersonGenderTreatmentTests` *(new)* | test file | `PersonDetailTests` | New subject; reuses `ReconciledValueDisplayTests` helpers. |

**Builds on.** `ConclusionDetailBody` defaults, `PVBadge` ink override, `ReconciledValueDisplay.string(for:)`, the reconciler's neutral-term rule.
**Deletes.** None — nothing becomes dead.
**Next adds.** S10-35 sets `PVThumbnail.Content.image` on the same header thumb and keeps `tint` as the outline (D-6), and adds the Likenesses section through `ConclusionDetailBody.sections` or a sibling slot. S10-39 needs no change here.

### Schema

None — no catalog change.

### Go core

None — D-1.

### FFI

None — D-1.

### macOS

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Value → tint | `PersonGenderTreatment` *(new)* | `macos/App/Features/Conclusions/PersonGenderTreatment.swift` *(new)* | `static func read(_ fields: [CatalogConclusionField]) -> PersonGenderTreatment?`; stores `tint`, `basis` (`.gender` / `.sexAtBirth`), `text`; exposes `badgeText`, `accessibilityLabel`, `thumbnailTint`, `ink`. D-2, D-3, D-4, D-5, D-9. |
| Person content | `PersonDetailContent` | `PersonDetailContent.swift` | `var treatment: PersonGenderTreatment?` set in `init`; `var thumbnailTint` returns `treatment?.thumbnailTint`. |
| Header thumb | `PVThumbnail` | `DesignSystem/Components/Thumbnail/PVThumbnail.swift` | D-6. Default `nil` keeps every existing call site unchanged. |
| Shared header | `ConclusionDetailHeader` | `ConclusionDetailPage.swift` | D-7. |
| Badge | `PVBadge` | `PersonDetailView.swift` `PersonDetailVitals` | D-8. `PersonDetailVitals` takes the content's `treatment` alongside `vitals`. Spacing between Badge and the b. line is the existing `PVSpacing.space1` stack unless the board shows otherwise. |

Frames this PR ships, in words (brief §5 frames 1–8): the Badge sits directly under the display name and above the b. line, left-aligned with them. The 80pt thumb behind the person mark is filled with the tint's fill and outlined with its line. Frames 7 and 8 are today's page. Frame 11 checks dark mode.

States: loading and error unchanged; neutral = today's header. Accessibility: the Badge is one element with the D-9 spoken label; the thumb's label stays `L10n.Conclusions.personNoLikeness`.

### Strings

L10n keys *(new)* in `L10n.Conclusions`: `sexAtBirthBadge(_:)` → `conclusions.sexAtBirthBadge` = “%@ · at birth”; `a11yGender(_:)` → `conclusions.a11y.gender` = “Gender: %@”; `a11ySexAtBirth(_:)` → `conclusions.a11y.sexAtBirth` = “Sex at birth: %@”. Removed: none.

## 7. Steps

**STEP-1 — The rule**
Files: `macos/App/Platform/ConclusionVocabulary.swift`, `macos/App/Features/Conclusions/PersonGenderTreatment.swift` *(new)*, `macos/ProvenenciaTests/PersonGenderTreatmentTests.swift` *(new)*. Skill: `add-swift-test`.
Do: add `SeededPropertyKey.gender`; write `PersonGenderTreatment.read` per D-2, D-3 with `tint` and `basis`; colours per D-5 from the board's Board line.
Check: T-1, T-2, T-3, T-4 pass in `xcodebuild test`.

**STEP-2 — Copy**
Files: `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`, `PersonGenderTreatment.swift`. Skill: `add-localized-string`.
Do: the three keys in §6 Strings; `badgeText` and `accessibilityLabel` per D-9.
Check: T-5 passes.

**STEP-3 — Thumb tint**
Files: `macos/App/DesignSystem/Components/Thumbnail/PVThumbnail.swift`. Skill: `add-ui-component`.
Do: D-6.
Check: the app builds; every existing `PVThumbnail(` call site compiles unchanged (`rg "PVThumbnail\(" macos/App` shows no edits outside this file and STEP-4).

**STEP-4 — Shared header**
Files: `ConclusionDetailPage.swift`, `PersonDetailContent.swift`.
Do: D-7; `PersonDetailContent.treatment` and `thumbnailTint`.
Check: T-6 passes; Event and Place pages build unchanged.

**STEP-5 — Badge**
Files: `PersonDetailView.swift`.
Do: D-8.
Check: manual check §10 steps 1–4.

**STEP-6 — Docs**
Files: `completed.md`, `deployment-plan.md` (S10-23 status), this plan (status + Landed note), `design/S10-D12-person-detail.md` stays in `design/` (it still gates S10-35 and S10-39), `design/README.md` note that PDP-1 to PDP-6 shipped.
Do: per the Definition of done.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-10/prs/S10-23-person-gender.md` reports no errors.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Show a Person's gender under their name on the Person page."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Swift | `PersonGenderTreatmentTests` `genderWinsOverSexAtBirth` *(new)* | gender `man` single + sex `female` single → tint `.man`, basis `.gender`; `woman` → `.woman`; `non_binary` → `.nonBinary`; key `two_spirit` → `.other` | AC-1.1, AC-1.2, AC-1.3, AC-1.4, D-3 |
| T-2 | Swift | `PersonGenderTreatmentTests` `mixedGenderIsNeutralAndDoesNotFallBack` *(new)* | gender two kept values (`mixed`) + sex `female` → nil | AC-1.5, D-2 |
| T-3 | Swift | `PersonGenderTreatmentTests` `sexAtBirthBorrowsATint` *(new)* | no gender; sex `female` → `.woman`/`.sexAtBirth`; `male` → `.man`; `intersex` → `.other`; gender field with zero displayed + sex `male` → `.man` | AC-2.1, AC-2.2, AC-2.3, D-2, D-3 |
| T-4 | Swift | `PersonGenderTreatmentTests` `nothingUsableIsNeutral` *(new)* | sex only an `unknown` outcome with `no_evidence`, no displayed → nil; sex `mixed` → nil; neither field → nil | AC-2.4, AC-2.5, AC-2.6 |
| T-5 | Swift | `PersonGenderTreatmentTests` `badgeAndSpokenText` *(new)* | gender “Woman” → badge “Woman”, spoken “Gender: Woman”; sex “Male” → badge “Male · at birth”, spoken “Sex at birth: Male” | AC-3.1, AC-3.2, D-9 |
| T-6 | Swift | `PersonDetailTests` `treatmentRidesOnContentAndRowsStay` *(new)* | `detail(extra: [gender woman with outcome])` → `content.treatment?.tint == .woman`, `thumbnailTint != nil`, rows include “Gender”; `detail()` → `treatment == nil`, `thumbnailTint == nil`, rows equal today's | PDP-6, D-7 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `PersonDetailTests`, `ReconciledValueDisplayTests` (shared helpers), and any snapshot of `PVThumbnail`.

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
- [ ] Brief moved to design/archive/ and design/README.md updated (if gated) — here: brief stays in design/ until S10-39; README notes PDP-1 to PDP-6 shipped
- [ ] Domain doc updated if a rule in docs/*.md changed: none (seeded-vocabulary.md §3.9 already describes gender)
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

Use a project created after `3d70139` (so `gender` is bound on person).

1. Person A: an Observation gender Woman and one sex at birth Female. Open A: Badge “WOMAN”, Woman tint on the thumb; Details shows Gender and Sex at birth rows.
2. Add an Observation gender Man on a second Source for A (now mixed). Reopen: no Badge, plain thumb; Gender row shows Mixed.
3. Person B: only sex at birth Male. Badge “MALE · AT BIRTH”, Man tint. VoiceOver on the Badge: “Sex at birth: Male”.
4. Person C: neither. The page matches `main`. Switch to dark mode and repeat step 1: the tint reads on the dark page.

This makes the Slice 6 Check line true: “A Person with gender shows gender; one with only sex at birth shows that; unknown is neutral.”

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| A tint reads as a status colour | D-5 bars status tones; the board names the tokens. |
| Extending `PVThumbnail` changes list thumbs | `tint` defaults to nil; STEP-3 check. |
| Old catalogs have no `gender` | Sex at birth still drives the treatment; manual check uses a new catalog. |

**Stop and report if:**

- A neutral sex-at-birth term (`unknown`, `indeterminate`) appears in `displayedValues` for a field in a test fixture built from the engine (D-2 assumes it never does).
- The agreed board places the Badge somewhere other than the Person summary slot, or draws a control instead of a Badge (D-8).
- The board's tints are not existing `PVColor` tokens (D-5).
- `PVThumbnail` has a call site that sets a background or border of its own that the tint would fight.

## 12. Open questions

1. **Tint tokens.** Which existing `PVColor` tokens are the fill, line, and ink of the Man, Woman, Non-binary, and other tints? Answered by the agreed S10-D12 board; record on the brief's Board line and copy into D-5.
2. **Board link.** The Claude Design link for the agreed board, for the **Board** line.

## 13. PR

**Title:** S10-23 — Sex / gender on the Person page
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-10/prs/S10-23-person-gender.md`.
