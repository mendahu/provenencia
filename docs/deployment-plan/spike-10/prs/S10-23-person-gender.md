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
**Rules that bind:** `.cursor/rules/design-system-layers.mdc`, `.cursor/rules/go-tests.mdc`
**Context:** working rules "Sex / gender styling", gotchas "Gender and sex at birth are different Properties", "Gallery and gender share a page"; [`docs/seeded-vocabulary.md`](../../../seeded-vocabulary.md) §3.7 (neutral terms) and §3.9 (gender); [`docs/design-system-layers.md`](../../../design-system-layers.md)
**Planned:** 2026-10-15 against `main` @ `daf4e08` (draft; redo §4 when the board is agreed)
**Review:** [`S10-23-person-gender.review.md`](S10-23-person-gender.review.md) — round 1 on the draft: Fits with changes / Not ready, all findings answered; round 2 runs when the board is agreed

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

**What stays.** Every Details row, including Sex at birth and Gender with their badges and Why; the name, the b. and d. lines, the ref and member count; the Event and Place pages; the engine. No schema, seed, FFI, or engine change. The only Go change is one drift subtest (D-10).

## 2. User stories

**US-1.** As a researcher, when I open a Person whose gender is recorded, I see it under the name, so I know who I am looking at without reading the rows.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | Gender `woman` (one displayed value), sex at birth `female` | Badge text “Woman” (drawn uppercase by `PVBadge`); header thumb has the woman colours (PDP-1, PDP-3). |
| AC-1.2 | Gender `man`, sex at birth `female` | Badge “Man”, man colours. Gender wins (working rule). |
| AC-1.3 | Gender `non_binary`, no sex at birth | Badge “Non-binary”, non-binary colours. |
| AC-1.4 | Gender is a researcher term, key `two_spirit`, label “Two-spirit” | Badge “Two-spirit”, other colours. |
| AC-1.5 | Gender has two displayed values (state `mixed`), sex at birth `female` | No Badge, plain thumb. Sex at birth is not used (PDP-4). |

**US-2.** As a researcher, when only sex at birth is recorded, the page shows it and says it is sex at birth, so I never read it as gender identity.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | No gender field, sex at birth `male` | Badge “Male · at birth”; man colours (PDP-2, PDP-3). |
| AC-2.2 | No gender field, sex at birth `female` | Badge “Female · at birth”; woman colours. |
| AC-2.3 | No gender field, sex at birth `intersex` | Badge “Intersex · at birth”; other colours. |
| AC-2.4 | Sex at birth recorded only as `unknown` or `indeterminate` (neutral terms; the reconciler marks them `no_evidence`, so nothing is displayed) | No Badge, plain thumb. |
| AC-2.5 | Sex at birth `mixed`, no gender | No Badge, plain thumb. |
| AC-2.6 | Neither field | The page as it is today: no Badge, plain thumb, same rows. |

**US-3.** As a VoiceOver user, I hear what the Badge means.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | AC-1.1 | The Badge reads “Gender, Woman”: the field's label, then the value, the way its Details row speaks (PDP-5). |
| AC-3.2 | AC-2.1 | The Badge reads “Sex at birth, Male”. |

## 3. Scope

| In | Covered by |
| --- | --- |
| Read the treatment from the detail's fields: precedence, family, text (PDP-1, PDP-2, PDP-3, PDP-4) | STEP-1, STEP-2, AC-1.1, AC-2.1 |
| Typed seeded term keys in `ConclusionVocabulary.swift` and their Go drift subtest | STEP-1, T-7 |
| `PVThumbnail` optional colours on the mark and symbol states (PDP-3, inventory Extend) | STEP-3 |
| Header thumb reads one `thumbnail` slot from the kind's content (inventory Extend) | STEP-4, T-6 |
| Badge under the name on the Person page, with its spoken label (PDP-1, PDP-2, PDP-5) | STEP-5, AC-3.1, AC-3.2 |
| Copy: “{value} · at birth” | STEP-2 |
| Details rows unchanged (PDP-6) | AC-2.6, T-6 |

| Out | Home |
| --- | --- |
| Likenesses section, the crop in the header thumb, the thumb colours on an image, and the `CachedThumbnail` colour pass-through (PDP-7, PDP-8, PDP-9) | S10-35 |
| Marking a portrait (PDP-10, PDP-11) | S10-39 |
| Seeding `gender` or its terms | Already landed in `3d70139`; not planned |
| Adding `gender` to catalogs created before `3d70139` | Not planned (Install data; add by hand, same as `engagement`) |
| Gender or sex on Event or Place pages, the Persons list, Promote | Not planned (brief §2.1) |
| Editing a value from the header | Not planned |
| A colour per researcher-added term, a new palette colour | Not planned (working rule) |
| Pronouns or text inferred from the value | Not planned (brief §2.1) |
| Moving `PromoteKindTile` and the Promote header thumb onto `PVThumbnail` | Not planned (review log M-8) |

## 4. Current state

Verified at `daf4e08`.

| Area | Where | Today |
| --- | --- | --- |
| Seed | `core/database/subjectvocab/registry.go` `seedProperties`, `seedBindings`, `seedTerms` | `gender` (term, unlocked on person, sort 2) with `man` / `woman` / `non_binary` already seeded. `sex_at_birth` terms: `female`, `male`, `intersex`, and neutral `indeterminate`, `unknown`. Nothing to change. |
| Neutral terms | `core/autoreconcile/modules.go` `termKey`; covered by `core/autoreconcile/modules_test.go` `TestModules` (`neutralTerm`) | A neutral term is no evidence, so its outcome is `no_evidence` and it is never a displayed value. |
| Detail payload | `macos/App/Platform/GenealogyStore.swift` `CatalogConclusionDetail`, `CatalogConclusionField` (`propertyKey`, `label`, `state`, `displayedValues`), `CatalogConclusionValue.term(id:key:label:)` | Carries the gender and sex-at-birth fields with term key, term label, and Property label. |
| Wire vocabulary | `macos/App/Platform/ConclusionVocabulary.swift` `SeededPropertyKey`; drift test `api/ffi/swift_vocabulary_test.go` `TestSwiftVocabularyMatchesGo`, subtest "seeded Property keys are seeded" | Has `sexAtBirth`, no `gender`, no term keys. The file header: "Spell these values only here". |
| Person content | `macos/App/Features/Conclusions/PersonDetailContent.swift` `PersonDetailContent` | Title, ref, members, vitals, rows. A Gender field with outcomes already shows as an ordinary row through the trailing `for field in fields` loop. |
| Person view | `macos/App/Features/Conclusions/PersonDetailView.swift` `PersonDetailView`, private `PersonDetailVitals` | The summary slot shows the b. and d. lines only. |
| Shared page | `macos/App/Features/Conclusions/ConclusionDetailPage.swift` `ConclusionDetailBody`, private `ConclusionDetailHeader` | Header builds `PVThumbnail(.init(glyph: .mark(mark), label: thumbnailLabel), size: 80)`. `sections` is the precedent for a protocol member with a default. |
| Thumb | `macos/App/DesignSystem/Components/Thumbnail/PVThumbnail.swift` `PVThumbnail.init(_:size:cornerRadius:)` | Filled states use `PVColor.surfaceInset` and `PVColor.borderDefault`; the mark uses `PVColor.textSecondary`. Framing (`size`, `cornerRadius`) is an init parameter; `Content` is state. `macos/App/Platform/CachedThumbnail.swift` passes `size` and `cornerRadius` through and replaces its `Content` whole on load. |
| Badge | `macos/App/DesignSystem/Components/Badge/PVBadge.swift` `PVBadge.init(text:tone:icon:subtle:foreground:)` | Data text with an ink override; draws its text uppercase. Precedent: `macos/App/Features/EvidenceGraph/EvidenceSubjectCard.swift` (ref badge with the kind's ink). |
| Value text and speech | `macos/App/Features/Conclusions/ReconciledValueDisplay.swift` `ReconciledValueDisplay.string(for:)`; `L10n.Conclusions.a11yList(_:rest:)` | A term reads as its label, else its key. Rows speak “{field label}, {value}” through `a11yList`. |
| Colour roles | `macos/App/Features/EvidenceGraph/EvidenceSubjectKindStyle.swift` `EvidenceSubjectKindStyle` (`ink`, `tint`, `line`, `chip`); `PVColor.subjectPersonTint` and siblings | The role names this PR reuses: tint (fill), line, ink. |
| Tests | `macos/ProvenenciaTests/PersonDetailTests.swift` `detail(...)`, `rowsAreNameThenLifeThenRecordedFields`; helpers `ReconciledValueDisplayTests.field/value/outcome` | Build fields by state and values. The new tests reuse them. |
| Xcode project | `macos/Provenencia.xcodeproj/project.pbxproj` | Every file has explicit entries (for example `PersonHeaderDisplay.swift`). New files must be registered. |

**Precedent.** The S10-19 kind marks: one feature rule decides the look, the kit components stay content-agnostic. Copy `EvidenceSubjectCard`'s `PVBadge(text:tone:icon:subtle:foreground:)` for an inked badge, and `ConclusionDetailBody.sections` for a protocol member with a default. Do not copy `SubjectTypeChrome`'s use of `PVPalette`; the colours are `PVColor` tokens (D-5).

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | Swift reads the treatment from `CatalogConclusionDetail.fields`. No engine or FFI change. | Go already sends both fields with term key and labels; the treatment is presentation of reconciled values (gotcha "Gender and sex at birth are different Properties"). | A Go `PersonHeader.gender` field — a second read of data the page already holds. |
| D-2 | Precedence counts `displayedValues`, not `state`. Gender field with exactly one displayed value → use it. Two or more → neutral, no fallback. Gender absent or zero displayed → sex at birth: exactly one displayed → use it; otherwise neutral. A displayed value that is not `.term`, or whose text from `ReconciledValueDisplay.string(for:)` is empty, is neutral (no fallback when it is the gender field). | Counts survive `concluded` and `unrecognized` states; zero displayed covers neutral-only records (§4 Neutral terms). | Switching on `ReconciledState` — misses `unrecognized`, and `concluded` would need its own case. |
| D-3 | Family by key. Gender: `man` → `.man`, `woman` → `.woman`, `non_binary` → `.nonBinary`, any other key → `.other`. Sex at birth: `female` → `.woman`, `male` → `.man`, any other key → `.other`. | User answer 2026-10-15 (map working rule). | A colour per researcher term — forbidden by the working rule. |
| D-4 | New feature type `PersonGenderTreatment` in `macos/App/Features/Conclusions/PersonGenderTreatment.swift` *(new)*, pure, `Equatable`. `PersonDetailContent` holds `treatment: PersonGenderTreatment?` (nil = neutral). | Same shape as `PersonHeaderDisplay`: one place decides, views lay out. Testable without views. | Logic in the view — untestable; in `ReconciledValueDisplay` — shared by Event and Place, which have none of this. |
| D-5 | Signatures: `struct PersonGenderTreatment: Equatable { enum Family: Equatable { case man, woman, nonBinary, other }; enum Basis: Equatable { case gender, sexAtBirth }; var family: Family; var basis: Basis; var text: String; var fieldLabel: String; static func read(_ fields: [CatalogConclusionField]) -> PersonGenderTreatment? }`. `Family` has computed `tint`, `line`, `ink: Color`, each an existing `PVColor` token named on the agreed board. `PersonGenderTreatment` exposes `badgeText: String`, `accessibilityLabel: String`, and `thumbnailColors: PVThumbnail.Colors`. No new `PVColor` or `PVPalette` entry. | User answer 2026-10-15: existing tokens, no new palette colour. Role names follow `EvidenceSubjectKindStyle` (review M-7). | Status tones (meaning collision); the Event / Place kind inks (brief §2). |
| D-6 | `PVThumbnail` gains `struct Colors: Equatable { var tint: Color; var line: Color }` *(new)* and an init parameter `colors: Colors? = nil`, after `cornerRadius`. In the mark and symbol states, `colors.tint` replaces the background and `colors.line` the border. Every other state ignores it in this PR. `Content` does not change. | Framing is an init parameter, like `size` and `cornerRadius`, so `CachedThumbnail`'s whole-`Content` reloads in S10-35 cannot drop it (review M-1). | A field on `Content` — wiped by `CachedThumbnail.load()`; an overlay — the thumb paints an opaque background. |
| D-7 | `ConclusionDetailThumbnail` *(new)*, `struct ConclusionDetailThumbnail: Equatable { var colors: PVThumbnail.Colors? = nil }`, in `ConclusionDetailPage.swift`. `ConclusionDetailBody` gains `var thumbnail: ConclusionDetailThumbnail { get }`, default `.init()` in the extension. `ConclusionDetailHeader` passes `content.thumbnail.colors` to `PVThumbnail(_:size:cornerRadius:colors:)`. | One slot for the header thumb; S10-35 adds the crop to this struct instead of a second member (review M-2). | A bare `thumbnailTint` member — S10-35 and S10-39 would each add another. |
| D-8 | The Badge is `PVBadge(text: badgeText, tone: .neutral, subtle: false, foreground: family.ink)`, no icon, in `PersonDetailVitals` above the b. and d. lines, followed by `.accessibilityLabel(accessibilityLabel)`. `PVBadge` already makes itself one accessibility element. Not a `PVChip`. If the agreed board draws `subtle`, an icon, or another tone, stop. | Brief inventory: Badge, display only. | `PVChip` — a selectable control with an action. |
| D-9 | `text` is `ReconciledValueDisplay.string(for:)` of the displayed value. `badgeText`: gender → `text`; sex at birth → `L10n.Conclusions.sexAtBirthBadge(text)` *(new)* = “%@ · at birth”. `accessibilityLabel` = `L10n.Conclusions.a11yList(fieldLabel, rest: text)`, where `fieldLabel` is the field's `label`. | One source of the value and of the Property name, the same ones the Details row uses (review M-4). | New “Gender: %@” keys — a second way to speak a field. |
| D-10 | `SeededPropertyKey.gender = "gender"` *(new)*. Term keys go in `ConclusionVocabulary.swift` as `enum SeededGenderTerm { static let man, woman, nonBinary }` and `enum SeededSexAtBirthTerm { static let female, male }` *(new)*. `TestSwiftVocabularyMatchesGo` gains subtest "seeded term keys are seeded" *(new)*: it installs the seed, looks up each Property by key, then each term with `propertyterms.Lookup(c, propertyID, key, properties.OriginProvenencia)`. | The file's rule and drift test cover every wire value (review M-3). | Private constants in the treatment — a renamed seed term would silently fall to `.other`. |
| D-11 | `xcodebuild`, the Swift tests, and the manual check run on a Mac. On a machine without them, the implementer runs the Go checks, stops before STEP-3, and reports which checks did not run. | Precedent: S10-41 §11. | Claiming a Swift check that was not run. |

## 6. Design

### Codebase fit

Searched: `gender`, `sexAtBirth`, `Tint`, `tint`, `Colors` under `macos/App`; `PVThumbnail(` call sites; `foreground:` on `PVBadge`; `EvidenceSubjectKindStyle`; `CachedThumbnail`.

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `PersonGenderTreatment` *(new)* | type / file | `PersonHeaderDisplay` | Different job (title text). Same pure-helper shape, beside it. |
| `PVThumbnail.Colors`, init `colors:` *(new)* | component extend | `PVThumbnail.init(_:size:cornerRadius:)`; `EvidenceSubjectKindStyle` (role names) | Extended. Role names reused. |
| `ConclusionDetailThumbnail`, `ConclusionDetailBody.thumbnail` *(new)* | type / protocol member | `ConclusionDetailBody.sections` | Extended the same way. |
| `SeededPropertyKey.gender`, `SeededGenderTerm`, `SeededSexAtBirthTerm` *(new)* | key | `SeededPropertyKey.sexAtBirth` | Extended in the wire-vocabulary file. |
| Subtest "seeded term keys are seeded" *(new)* | Go test | "seeded Property keys are seeded" | Same shape, one level down. |
| `L10n.Conclusions.sexAtBirthBadge` *(new)* | key | `L10n.Conclusions.a11yRef` | Same namespace and format shape. |
| `PersonGenderTreatmentTests` *(new)* | test file | `PersonDetailTests` | New subject; reuses `ReconciledValueDisplayTests` helpers. |

**Builds on.** `ConclusionDetailBody` defaults, `PVBadge` ink override, `ReconciledValueDisplay.string(for:)`, `a11yList`, the wire-vocabulary drift test, the reconciler's neutral-term rule.
**Deletes.** None — nothing becomes dead.
**Next adds.** S10-35 adds the crop to `ConclusionDetailThumbnail`, renders the header thumb through `CachedThumbnail` with a `colors` pass-through, and draws `colors.line` on the image state. S10-39 changes which crop `ConclusionDetailThumbnail` carries. Neither adds a protocol member for the thumb.

### Schema

None — no catalog change.

### Go core

None — D-1. Test only: D-10.

### FFI

None — D-1.

### macOS

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Value → family | `PersonGenderTreatment` *(new)* | `macos/App/Features/Conclusions/PersonGenderTreatment.swift` *(new)* | D-2, D-3, D-4, D-5, D-9. |
| Person content | `PersonDetailContent` | `PersonDetailContent.swift` | `var treatment: PersonGenderTreatment?` set in `init`; `var thumbnail: ConclusionDetailThumbnail` returns `.init(colors: treatment?.thumbnailColors)`. |
| Header thumb | `PVThumbnail` | `DesignSystem/Components/Thumbnail/PVThumbnail.swift` | D-6. Default `nil` keeps every existing call site unchanged. |
| Shared header | `ConclusionDetailHeader` | `ConclusionDetailPage.swift` | D-7. |
| Badge | `PVBadge` | `PersonDetailView.swift` `PersonDetailVitals` | D-8. `PersonDetailVitals` takes the content's `treatment` alongside `vitals`. The Badge sits in the existing `PVSpacing.space1` stack, first. |

Frames this PR ships, in words (brief §5 frames 1 to 8, 8b, and 11): the Badge sits directly under the display name and above the b. line, left-aligned with them. The 80pt thumb behind the person mark is filled with the family's tint and outlined with its line. Frames 7, 8, and 8b are today's page. Frame 11 checks dark mode.

States: loading and error unchanged; neutral = today's header. Accessibility: the Badge is one element with the D-9 spoken label; the thumb's label stays `L10n.Conclusions.personNoLikeness`.

### Strings

L10n key *(new)* in `L10n.Conclusions`: `sexAtBirthBadge(_:)` → `conclusions.sexAtBirthBadge` = “%@ · at birth”, comment “Person page badge when only sex at birth is recorded; argument is the term label”. Removed: none.

## 7. Steps

**STEP-1 — The rule and its keys**
Files: `macos/App/Platform/ConclusionVocabulary.swift`, `api/ffi/swift_vocabulary_test.go`, `macos/App/Features/Conclusions/PersonGenderTreatment.swift` *(new)*, `macos/ProvenenciaTests/PersonGenderTreatmentTests.swift` *(new)*, `macos/Provenencia.xcodeproj/project.pbxproj`. Skill: `add-swift-test`.
Do: D-10 keys and subtest; `PersonGenderTreatment` per D-2, D-3, D-5 with the colours from the brief's Board line; register both new files in the project (app target and test target, per `add-swift-test`).
Check: `CGO_ENABLED=1 go test -tags fts5 ./api/ffi/...` passes; T-1, T-2, T-3, T-4 pass in `xcodebuild test`.

**STEP-2 — Copy**
Files: `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`, `PersonGenderTreatment.swift`. Skill: `add-localized-string`.
Do: the §6 Strings key; `badgeText` and `accessibilityLabel` per D-9.
Check: T-5 passes.

**STEP-3 — Thumb colours**
Files: `macos/App/DesignSystem/Components/Thumbnail/PVThumbnail.swift`. Skill: `add-ui-component`.
Do: D-6. Add a `#Preview` case with `colors:` on a mark if the file has a preview block.
Check: the app builds; `git diff --stat -- macos/App` lists only files named in STEP-1 to STEP-5.

**STEP-4 — Shared header**
Files: `ConclusionDetailPage.swift`, `PersonDetailContent.swift`.
Do: D-7; `PersonDetailContent.treatment` and `thumbnail`.
Check: T-6 passes; Event and Place pages build unchanged.

**STEP-5 — Badge**
Files: `PersonDetailView.swift`.
Do: D-8.
Check: manual check §10 steps 1 to 4.

**STEP-6 — Docs**
Files: `completed.md` (one paragraph), `deployment-plan.md` (S10-23 entry: **Status:** `Landed #N`, and the PR sequence line "— done, #N"), this plan (status + Landed note), `design/README.md` (S10-D12 row in **Scheduled**: add “PDP-1 to PDP-6 shipped in S10-23 (#N)”). The brief stays in `design/`; S10-39 archives it.
Do: per the Definition of done.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-10/prs/S10-23-person-gender.md` reports no errors.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Show a Person's gender under their name on the Person page."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Swift | `PersonGenderTreatmentTests` `genderWinsOverSexAtBirth` *(new)* | gender `man` single + sex `female` single → `.man`, `.gender`; `woman` → `.woman`; `non_binary` → `.nonBinary`; key `two_spirit` → `.other` | AC-1.1, AC-1.2, AC-1.3, AC-1.4, D-3 |
| T-2 | Swift | `PersonGenderTreatmentTests` `mixedGenderIsNeutralAndDoesNotFallBack` *(new)* | gender two displayed values (`mixed`) + sex `female` → nil | AC-1.5, D-2 |
| T-3 | Swift | `PersonGenderTreatmentTests` `sexAtBirthBorrowsAFamily` *(new)* | no gender; sex `female` → `.woman`/`.sexAtBirth`; `male` → `.man`; `intersex` → `.other`; gender field with zero displayed + sex `male` → `.man` | AC-2.1, AC-2.2, AC-2.3, D-2, D-3 |
| T-4 | Swift | `PersonGenderTreatmentTests` `nothingUsableIsNeutral` *(new)* | sex only `unknown` and `indeterminate` outcomes with `no_evidence`, nothing displayed → nil; sex `mixed` → nil; neither field → nil; gender displayed `.text("x")` → nil; gender term with empty label and key → nil | AC-2.4, AC-2.5, AC-2.6, D-2 |
| T-5 | Swift | `PersonGenderTreatmentTests` `badgeAndSpokenText` *(new)* | gender “Woman”, field label “Gender” → badge “Woman”, spoken “Gender, Woman”; sex “Male”, label “Sex at birth” → badge “Male · at birth”, spoken “Sex at birth, Male” | AC-3.1, AC-3.2, D-9 |
| T-6 | Swift | `PersonDetailTests` `treatmentRidesOnContentAndRowsStay` *(new)* | `detail(extra: [gender woman with outcome])` → `treatment?.family == .woman`, `thumbnail.colors != nil`, rows include “Gender”; `detail()` → `treatment == nil`, `thumbnail == .init()`, rows as today; an `EventDetailContent` and a `PlaceDetailContent` → `thumbnail == .init()` | PDP-6, D-7 |
| T-7 | Go | `api/ffi/swift_vocabulary_test.go` `TestSwiftVocabularyMatchesGo` / "seeded term keys are seeded" *(new)* | every key in `SeededGenderTerm` and `SeededSexAtBirthTerm` → found under its Property after `subjectvocab.Install` | D-10 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `PersonDetailTests`, `ReconciledValueDisplayTests` (shared helpers), `TestSwiftVocabularyMatchesGo`, and any snapshot of `PVThumbnail`.

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests, catalog and plan-doc checks); any check not run on this machine is reported as not run (D-11)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together) — or none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated (if gated) — here: brief stays in design/ until S10-39; README row notes PDP-1 to PDP-6 shipped
- [ ] Domain doc updated if a rule in docs/*.md changed: none (seeded-vocabulary.md §3.9 already describes gender)
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

On a Mac, with a project created after `3d70139` (so `gender` is bound on person). Record values in the citation composer: open a Source, choose a Citation, add an Observation on a person Subject with Property Gender (or Sex at birth) and a term, save. Promote the person Subject on the Promote page so it files to a Person.

1. Person A: gender Woman and sex at birth Female on one Source. Open A: Badge “WOMAN” (`PVBadge` draws uppercase), woman colours on the thumb; Details shows Gender and Sex at birth rows. VoiceOver on the Badge: “Gender, Woman”.
2. On a second Source, record gender Man for a Subject filed to A (now mixed). Reopen A: no Badge, plain thumb; the Gender row shows Mixed.
3. Person B: only sex at birth Male. Badge “MALE · AT BIRTH”, man colours. VoiceOver on the Badge: “Sex at birth, Male”.
4. Person C: neither. The page matches `main`. Switch to dark mode and repeat step 1: the colours read on the dark page.

This makes the Slice 6 Check line true: “A Person with gender shows gender; one with only sex at birth shows that; unknown is neutral.”

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| A colour reads as a status | D-5 bars status tones; the board names the tokens. |
| Extending `PVThumbnail` changes list thumbs | `colors` defaults to nil; STEP-3 check. |
| Old catalogs have no `gender` | Sex at birth still drives the treatment; the manual check uses a new catalog. |

**Stop and report if:**

- `core/autoreconcile/modules_test.go` `TestModules` no longer shows a neutral term as no evidence (D-2 relies on it).
- The agreed board places the Badge somewhere other than the Person summary slot, or draws a control, an icon, `subtle`, or a tone other than neutral (D-8).
- The board's colours are not existing `PVColor` tokens, or it does not name a tint, line, and ink token for each family (D-5).
- `xcodebuild` cannot run on this machine (D-11): finish the Go checks, then stop and report.

## 12. Open questions

1. **Colour tokens.** Which existing `PVColor` tokens are the tint, line, and ink of the man, woman, non-binary, and other families? Answered by the agreed S10-D12 board (PDP-12); record on the brief's Board line and copy into D-5.
2. **Board link.** The Claude Design link for the agreed board, for the **Board** line.

## 13. PR

**Title:** S10-23 — Sex / gender on the Person page
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-10/prs/S10-23-person-gender.md`.
