# S11-29 — Quote nicknames and add former surnames

**Status:** Planned
**Spike / slice:** Spike 11 · Slice 2 — Value compare, everywhere (follow-up to S11-06; display only)
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no new chrome: existing name titles gain punctuation and one italic run; no new component, layout or state. The user chose the treatment on 2026-10-19)
**Board:** —
**Depends on:** —
**Blocks:** —
**Branch:** `s11-29-name-display`
**Version bump:** none — Spike 11 PRs leave `VERSION` at 0.2.0, and the product bump is cut when a release ships ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-name-value`](../../../../.agents/skills/add-name-value/SKILL.md), [`add-match-rule`](../../../../.agents/skills/add-match-rule/SKILL.md), [`add-localized-string`](../../../../.agents/skills/add-localized-string/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md), [`add-ui-component`](../../../../.agents/skills/add-ui-component/SKILL.md)
**Rules that bind:** `.cursor/rules/name-values.mdc`, `.cursor/rules/match-rules.mdc`, `.cursor/rules/go-tests.mdc`, `.cursor/rules/macos-l10n.mdc`, `.cursor/rules/swift-tests.mdc`, `.cursor/rules/macos-client.mdc`
**Context:** [`docs/structured-name-model.md`](../../../structured-name-model.md) §4.2a roles, §4.5 display styles; [`docs/seeded-vocabulary.md`](../../../seeded-vocabulary.md) §4.1, §4.3; [`docs/ideas/international-names.md`](../../../ideas/international-names.md) "Name changes as a signal"; S9-D2 title anatomy in `macos/App/Features/Conclusions/ConclusionListRow.swift`
**Planned:** 2026-10-19 against `main` @ `84a9b31`
**Review:** [`S11-29-name-display.review.md`](S11-29-name-display.review.md) — Fits with changes / Not ready → fixed, 1 round

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

A Person's name says which parts are a nickname and which a former surname. A nickname shows in curly double quotes (*James “Jake” Robins*), the genealogy convention. A new product part type, **Former surname**, records a maiden or other earlier surname; it shows in parentheses (*Mary (Smith) Jones*), and on the Persons list and Person page titles the surname inside the parentheses is italic. Today a nickname reads as one more given name (*James Jake Robins*) and a maiden name has no part type of its own.

**Why now.** User request (2026-10-19). [`international-names.md`](../../../ideas/international-names.md) names maiden names in the conclusion as the intended path for name changes.

**What stays.** Stored parts and `form` are never rewritten; the decoration is display only. The Go list sort key (`nameformats.Format`, read by `conclusionheaders`) stays undecorated. Existing catalogs are not migrated (user, 2026-10-19).

## 2. User stories

**US-1.** As a researcher, when a name has a nickname part, I see it in quotes, so that I can tell it from a middle name.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | Parts `given=James`, `nick=Jake`, `surname=Robins`, the `western` fixture (T-4), natural | `James “Jake” Robins` |
| AC-1.2 | Same, sorted | `Robins, James “Jake”` |
| AC-1.3 | Two nick parts `Jake`, `Jim` | `James “Jake” “Jim” Robins` (each part quoted) |
| AC-1.4 | The omnibar hit (with a person header) and the Promote row for that Person | Show the quotes; both reach `NameValueDisplay` |

**US-2.** As a researcher, I can type a name part as **Former surname**, and see it in parentheses and italics on the Person's title, so that a maiden name is recorded and reads as earlier.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | The part-type picker in the name editor | Lists *Former surname* right after *Surname* |
| AC-2.2 | Parts `given=Mary`, `former_surname=Smith`, `surname=Jones`, natural | Text `Mary (Smith) Jones`; the attributed title has one emphasized run, `Smith` |
| AC-2.3 | Same, sorted (Persons list row) | Text `Jones, Mary (Smith)`; one emphasized run, `Smith` |
| AC-2.4 | A name with no former surname | `ConclusionTitleSource` equals `.name("James Robins")` as today (`styled` nil) |
| AC-2.5 | Go `namevalue.Validate` on a part typed `former_surname` | Accepted |
| AC-2.6 | Orders that do not list `former_surname` (a catalog created before this PR), natural | `Mary Jones (Smith)`: an unlisted type follows, as `NameValueDisplay` does today |
| AC-2.7 | `valuecompare.Compare` of *Mary Smith* (`given=Mary`, `surname=Smith`) with `given=Mary`, `former_surname=Smith`, `surname=Jones`, shipped rules | `Contains`; one pair `smith`/`smith`, relation `Same` |

## 3. Scope

| In | Covered by |
| --- | --- |
| `former_surname` in the Go part-type registry, Western seed parts and roles, match registry Western roles | STEP-1, AC-2.5, AC-2.7 |
| `NamePartType.formerSurname`, its label, picker entry, and `NamePartType.decoration` | STEP-2, AC-2.1 |
| `NameValueDisplay` builds runs; nick parts quoted, former surnames parenthesized, for every caller | STEP-3, AC-1.1, AC-1.2, AC-1.3, AC-1.4, AC-2.6 |
| An attributed Person title with the former surname italic, on the Persons list row and Person detail title | STEP-4, AC-2.2, AC-2.3, AC-2.4 |
| Docs: structured-name-model §4.2a and §4.5, seeded-vocabulary §4.1 / §4.3, international-names; comments on `nameformats.Format` and `NameValueDisplay` | STEP-5 |

| Out | Home |
| --- | --- |
| Migrating existing catalogs' name-format rows | not planned (user, 2026-10-19: new catalogs only) |
| Italics anywhere but the Persons list row and Person detail title (value rows, omnibar, Promote, Evidence) | not planned; those surfaces are plain text and keep the parentheses |
| Quotes in Go `nameformats.Format` (list sort key) and the Go-indexed omnibar title used when a hit has no person header | not planned; a sort or index key carries no punctuation (§4.5 rewritten in STEP-5) |
| `NameFormSplit` recognising quoted nicknames and parenthesized former surnames when splitting a typed form | not planned; logged in [`docs/ideas/international-names.md`](../../../ideas/international-names.md) in STEP-5, pointing at `NamePartType.decoration` |
| A Go↔Swift part-type parity check; dropping the unused `Sort` field | not planned; logged in `docs/ideas/international-names.md` in STEP-5 |
| The reconciler treating a former surname and a surname as related | not planned |
| Nickname alias folding | S11-28 |
| GEDCOM import / export of former surnames | not planned; no GEDCOM path exists |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Part-type registry (Go) | `core/namevalue/namevalue.go` consts; `core/namevalue/registry.go` `partTypeRegistry` | Seven types: prefix 0, given 1, nick 2, surname_prefix 3, surname 4, suffix 5, undetermined 6 |
| Western seed | `core/database/nameformats/registry.go` `seedProfiles` | Natural: prefix, given, nick, surname_prefix, surname, suffix. Sorted: surname_prefix, surname (`", "`), prefix, given, nick, suffix. Roles for all seven types. `Install` fails when the default profile lacks a role for a registry type |
| Match registry | `core/matchregistry/value.go` `Names.Western.PartRoles` | Roles for the seven types and `""` |
| Go display join | `core/database/nameformats/nameformats.go` `Format` | The same order-and-separator join as Swift, used as the Persons list sort key (`core/database/conclusionheaders/order.go:43`) |
| Swift part types | `macos/App/Features/Names/NamePartType.swift` `NamePartType` | Seven cases; `label` switch; `selectOptions` builds the picker from `allCases` |
| Strings | `macos/App/Platform/L10n.swift` `L10n.NameValue.partType*` (static `LocalizedStringResource` constants); `macos/App/Resources/Localizable.xcstrings`; `L10n.format(_:_:)` | `nameValue.part.type.<key>` labels |
| Display | `macos/App/Features/Names/NameValueDisplay.swift` `NameValueDisplay.join` | Groups parts by type, joins words with spaces, groups in the profile's order with trailing separators; unlisted types follow in stored order. Returns `String`; no per-type decoration |
| Display callers | `PersonHeaderDisplay`, `ReconciledValueDisplay`, `EventTitleDisplay`, `ObservationValueDisplay`, `EvidenceSubjectCard`, `CitationObservationValue`, `SourceGraphSnapshot.displayName` (Promote, Evidence graph) | All call `NameValueDisplay.string`; they gain the quotes and parentheses with no change |
| Title source | `macos/App/Features/Conclusions/ConclusionListRow.swift` `ConclusionTitleSource` | `.name(String)`, `.label(String)` (rendered italic), `.ref(String)`; `text` matches `case .name(let text), .label(let text), .ref(let text)` in one arm |
| Person title | `macos/App/Features/Conclusions/PersonHeaderDisplay.swift` `titleSource(name:entity:style:orders:)` | The one place a Person's title is made; calls `NameValueDisplay.string`; used by `PersonsList.titleSource` (sorted), `PersonDetailContent` (natural), omnibar and Promote (`title`) |
| Title rendering | `ConclusionListRow.titleText(_:)`; `ConclusionDetailPage.swift` `title` (line 430) | `Text(verbatim: text)` for `.name` |
| Tests | Go: `core/namevalue/registry_test.go` `TestPartTypesRegistry` (`len == 7`); `core/matchregistry/registry_test.go` `TestWesternPatternCoversEveryPartType`; `core/database/nameformats/nameformats_test.go` `TestInstallLoadsWesternAndFormats`, `TestSeedMatchesRegistryWestern` (seed roles equal registry roles). Swift: `PersonHeaderDisplayTests.swift`, `PersonsListTests.swift`, `PersonDetailTests.swift`, `NameValueDraftTests.swift`, `ObservationValueDisplayTests.swift`, `CitationComposerModelTests.swift` | Parity tests force a role for every registry type |

**Precedent.** `nick` is the nearest part type: copy every place it appears (Go const and registry row, seed part rows in both styles, seed role, match registry role, Swift case, L10n key, xcstrings entry). `core/valuecompare/name_test.go` `partCodes` is a test alphabet; it does not change. The `.label` title case shows how a title goes italic in both the list row and the detail page.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | Nickname display: each `nick` part wrapped by `L10n.NameValue.displayNick` (key `nameValue.display.nick`, value `“%@”`) | User call (2026-10-19); GEDCOM tools, Gramps and RootsMagic quote nicknames; localizable for « » or „ “ | Parentheses — conventionally a maiden name |
| D-2 | Former surname display: each `former_surname` part wrapped by `L10n.NameValue.displayFormerSurname` (key `nameValue.display.former_surname`, value `(%@)`); the word inside is emphasized, the parentheses are not; italics show only on the Persons list row and Person detail title | User call (2026-10-19): parentheses + italic. Plain-text surfaces keep the parentheses | Italic only — lost on plain-text surfaces and close to the italic working-label title (S9-D2); *née* — gendered and language-bound |
| D-3 | Key `former_surname`; label "Former surname"; Sort 5 (after surname); suffix 6, undetermined 7 | Generic for maiden, earlier married or changed names (user) | `maiden` — narrower |
| D-4 | Western order: natural prefix, given, nick, **former_surname**, surname_prefix, surname, suffix; sorted surname_prefix, surname (`", "`), prefix, given, nick, **former_surname**, suffix | *Mary (Smith) Jones* / *Jones, Mary (Smith)* | After the surname — reads as a second surname |
| D-5 | Role `family` in the seed roles and the match registry (both, so `TestSeedMatchesRegistryWestern` holds) | A former surname is a family name for matching; the record *Mary Smith* should pair it | `ignored` — throws away the maiden-name signal; a new role — needs a comparer change |
| D-6 | Decoration lives in Swift: `struct NamePartDecoration { let format: LocalizedStringResource; let emphasized: Bool }` *(new)* and `NamePartType.decoration: NamePartDecoration?` *(new)*: `.nick` → `(displayNick, false)`, `.formerSurname` → `(displayFormerSurname, true)`, every other case `nil`. Go has no display decoration | One property per decorated type; display is Swift's job; Go `Format` is a sort key | Two parallel properties (`displayFormat`, `isEmphasized`) — two switches per type (review M-3); a catalog column — schema for two constants |
| D-7 | `NamePartDecoration.runs(_ word: String) -> [NameDisplayRun]` *(new)* formats `L10n.format(format, NamePartDecoration.slot)` where `slot = "\u{E000}"` (a private-use character), splits it with `formatted.components(separatedBy: NamePartDecoration.slot)` (keeps empty pieces; not `split(separator:)`), counts the pieces, then drops empty ones; exactly two pieces → `[prefix (plain), word (emphasized when emphasized), suffix (plain)]`, empty pieces dropped; any other count (a translation without `%@`, or with it twice) → `[word]` undecorated | `L10n.format` stays the one formatting call; the split has one home; a bad translation degrades to the bare word | Formatting the word directly and searching for it — breaks when the word appears in the format |
| D-8 | `struct NameDisplayRun: Equatable { var text: String; var emphasized: Bool }` *(new, internal)*. `NameValueDisplay.runs(for:order:) -> [NameDisplayRun]` *(new, internal)* replaces `join`: each part of a decorated type contributes its decoration's runs, an undecorated part a plain run; the space between words and the trailing separators between groups are plain runs. `string(for:order:)` = the runs' text concatenated (the form fallback unchanged). `attributed(for:style:orders:) -> AttributedString` *(new)* builds from the same runs, setting `inlinePresentationIntent = .emphasized` on emphasized runs; when the runs are empty (no parts, or all-blank parts) it is `AttributedString(string(form:))`, the same fallback as `string` | One join for both outputs (review M-2) | A second join for attributed text — a near-copy |
| D-9 | `ConclusionTitleSource.name(String, styled: AttributedString? = nil)`, built for names by `static func name(_ styled: AttributedString) -> ConclusionTitleSource` *(new)*: `text` is `String(styled.characters)`, and `styled` is kept only when a run has `inlinePresentationIntent == .emphasized`. `text` splits its arm: `case .name(let text, _): text` and `case .label(let text), .ref(let text): text`. The enum's doc comment adds: "a name may carry a styled copy with emphasized runs" | One source for the characters; `.name("James Robins")` equality holds (AC-2.4) | A new `.styledName` case — every switch grows a case; always setting `styled` — breaks AC-2.4 |
| D-10 | `PersonHeaderDisplay.titleSource(name:…)` calls `NameValueDisplay.attributed` once and returns `.name(styled)` (D-9 constructor) when its text is not empty; `title(...)` still returns `titleSource(...).text` | One join per title | Calling `string` and `attributed` — two joins |
| D-11 | Both title views render `Text(styled)` when `styled` is set, else `Text(verbatim: text)`, keeping today's modifiers. If the emphasized run renders upright in manual check step 3, the fallback is binding: before rendering, set `.font` on each emphasized run's range to the view's own font with `.italic()` (`PVFont.display(size: PVTypeScale.display3, weight: PVFontWeight.medium).italic()` on the detail page; `PVFont.body(size: PVTypeScale.body).italic()` in the list row, the font `PVList` gives titles) | The display font already has an italic (the `.label` title uses it) | `.italic()` on the whole title — the label treatment |
| D-12 | No migration; old catalogs show the type after the listed ones (AC-2.6) and match it untyped | User call (2026-10-19) | Backfill migration |
| D-13 | The omnibar hit without a person header keeps the Go-indexed title, unquoted | That title is a search index projection (Out) | Re-formatting index titles in Swift |

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `namevalue.PartTypeFormerSurname` *(new)* | const | `PartTypeNick` | Extends the registry |
| `NamePartType.formerSurname`, `NamePartType.decoration` *(new)* | enum case, property | `NamePartType.label` | Extends the one Swift part-type registry |
| `NamePartDecoration`, `NamePartDecoration.runs(_:)` *(new)*, in `NamePartType.swift` | type, func | `L10n.format` | Wraps a word with a localized pattern and reports the emphasized piece |
| `NameDisplayRun`, `NameValueDisplay.runs(for:order:)` *(new)* | type, func | `NameValueDisplay.join` (replaced) | `join` becomes `runs` |
| `NameValueDisplay.attributed(for:style:orders:)` *(new)* | func | `NameValueDisplay.string(for:style:orders:)` | Same runs, different output |
| `ConclusionTitleSource.name(_ styled: AttributedString)` *(new)* and the `styled` payload | constructor, enum payload | `.label` italic | D-9 |
| `L10n.NameValue.partTypeFormerSurname`, `displayNick`, `displayFormerSurname` *(new)* | keys | `L10n.NameValue.partTypeNick` | New copy |
| `TestWesternSeedOrder` *(new)*, `TestFormerSurnamePairsAsFamily` *(new)* | Go tests | `TestInstallLoadsWesternAndFormats` | No existing test reads the seeded order |

**Builds on.** `namevalue` registry, `nameformats` seed, `matchregistry` Western pattern, `NamePartType`, `NameValueDisplay`, `PersonHeaderDisplay`, `L10n.format`.
**Deletes.** `NameValueDisplay.join` (replaced by `runs`). The sentence in structured-name-model §4.5 requirement 2 that the sort key is "the same string the row shows".
**Next adds.** `patronymic` or `maternal_surname` ([international-names](../../../ideas/international-names.md)) is one enum case and, if decorated, one `decoration` arm.

### Schema

None — part types are a compiled registry; `name_format_profile_parts.part_type` and `name_format_profile_roles.part_type` are open TEXT. New catalogs get the rows from the seed.

### Go core

| Package | Change |
| --- | --- |
| `core/namevalue` | `PartTypeFormerSurname = "former_surname"` *(new)*; registry row `{Key: PartTypeFormerSurname, L10nKey: "nameValue.part.type.former_surname", Sort: 5}`; suffix Sort 6, undetermined 7 |
| `core/database/nameformats` (`registry.go`) | Part rows per D-4; role `{PartTypeFormerSurname, RoleFamily}`. `nameformats.go` `Format` doc comment adds: "The same order and separators as Swift `NameValueDisplay`; only Swift decorates (quotes, parentheses). This string is a sort key." |
| `core/matchregistry` (`value.go`) | `namevalue.PartTypeFormerSurname: valuecompare.RoleFamily` in `Western.PartRoles` |

### FFI

None — part types and name-format rows already cross as strings.

### macOS

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Part type | `NamePartType.formerSurname = "former_surname"` | `macos/App/Features/Names/NamePartType.swift` | Case right after `surname`; `label` → `L10n.NameValue.partTypeFormerSurname`; `decoration` per D-6 |
| Decoration | `NamePartDecoration` | same file | D-6, D-7 |
| Runs | `NameDisplayRun`, `NameValueDisplay.runs` | `macos/App/Features/Names/NameValueDisplay.swift` | D-8. A comment on `runs`: "The same order and separators as Go `nameformats.Format`; only this side decorates." |
| Title payload | `ConclusionTitleSource` | `ConclusionListRow.swift` | D-9; the three matches on `.name` are `text`, `titleText` (this file) and `ConclusionDetailPage.title` |
| Person title | `PersonHeaderDisplay.titleSource(name:…)` | `PersonHeaderDisplay.swift` | D-10 |
| List row | `ConclusionListRow.titleText` | `ConclusionListRow.swift` | `.name(let text, let styled)`: `if let styled { Text(styled) } else { Text(verbatim: text) }` |
| Detail title | `ConclusionDetailPage` `title` | `ConclusionDetailPage.swift` | Same, with today's font, tracking, colour and header trait |

States: a name with only a former surname → `(Smith)`, italic inside. Accessibility: VoiceOver reads `text` (with quotes and parentheses); `ConclusionListRow.accessibilityLabel` is unchanged.

### Strings

Static `LocalizedStringResource` constants *(new)* in `L10n.NameValue`, each with an entry in `Localizable.xcstrings`:

- `partTypeFormerSurname`: `nameValue.part.type.former_surname` = "Former surname" — comment "NameValue part type: former_surname"
- `displayNick`: `nameValue.display.nick` = "“%@”" — comment "A nickname inside a displayed name; the argument is the nickname. Keep exactly one %@"
- `displayFormerSurname`: `nameValue.display.former_surname` = "(%@)" — comment "A former surname inside a displayed name; the argument is the surname. Keep exactly one %@"

## 7. Steps

Steps 2–4 need macOS with Xcode for their checks (`scripts/check-pr.sh` runs `xcodebuild test`). On a host without Xcode, finish STEP-1 and the code of STEP-2 – STEP-4, then stop and report that the Swift checks and the manual check need a macOS runner.

**STEP-1 — Go: the former surname part type**
Files: `core/database/conclusionheaders/order.go` (`personSortKey` comment: "the profile's sorted part order, undecorated; the row may add quotes and parentheses"), `core/namevalue/namevalue.go`, `core/namevalue/registry.go`, `core/namevalue/registry_test.go`, `core/database/nameformats/registry.go`, `core/database/nameformats/nameformats.go` (comment), `core/database/nameformats/nameformats_test.go`, `core/matchregistry/value.go`, `core/matchregistry/registry_test.go`. Skills: `add-name-value`, `add-match-rule`.
Do: D-3, D-4, D-5. Add T-1, T-2, T-6.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/namevalue/ ./core/database/nameformats/ ./core/matchregistry/ ./core/valuecompare/` passes.

**STEP-2 — Swift: the part type, its label and decoration**
Files: `NamePartType.swift`, `L10n.swift`, `Localizable.xcstrings`, `NameValueDraftTests.swift`. Skill: `add-localized-string`.
Do: the case, label, `NamePartDecoration`, `decoration`, and the three keys (D-6, D-7). Add T-3 and T-7.
Check: `python3 scripts/check-localizable-xcstrings.py` passes; T-3 and T-7 pass.

**STEP-3 — Swift: quotes and parentheses in every name display**
Files: `NameValueDisplay.swift`, `PersonHeaderDisplayTests.swift`. Skill: `add-swift-test`.
Do: D-1, D-2, D-8. Add T-4.
Check: Xcode tests pass T-4 and the existing `PersonHeaderDisplayTests`, `ObservationValueDisplayTests`, `CitationComposerModelTests`.

**STEP-4 — Swift: italic former surname on the Person titles**
Files: `ConclusionListRow.swift`, `ConclusionDetailPage.swift`, `PersonHeaderDisplay.swift`, `PersonHeaderDisplayTests.swift`. Skill: `add-ui-component` (no new component; the title stays a snowflake of `ConclusionListRow`).
Do: D-9, D-10, D-11. Add T-5.
Check: Xcode tests pass, including `PersonsListTests` and `PersonDetailTests` unchanged; manual check steps 2–3.

**STEP-5 — Docs**
Files: `docs/seeded-vocabulary.md` §4.1 (inside the code fence, after `surname`: `former_surname      -- no GEDCOM part tag (GEDCOM 7: a NAME with TYPE MAIDEN or BIRTH)`) and §4.3 (both orders per D-4); `docs/structured-name-model.md` §4.2a roles sentence (`former_surname` → `family`) and §4.5 (requirement 2 becomes "The key is that part order, undecorated: the row may add quotes and parentheses, the key never does"; the table gains a row "**Decoration** — nickname quotes, former-surname parentheses | The **part type**, in the client, localized by UI language; never in the sort key or the profile | *James “Jake” Robins*"); `docs/ideas/international-names.md` under "Name changes as a signal" ("The `former_surname` part type (S11-29) records a maiden or earlier surname; using it as a matching signal is still open.") and two parked lines (Out: `NameFormSplit` wrappers; Go↔Swift part-type parity and `Sort`); `docs/deployment-plan/spike-11/deployment-plan.md` map row; `docs/deployment-plan/spike-11/completed.md`; this plan.
Do: per the Definition of done.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-11/prs/S11-29-name-display.md` reports no errors; links resolve.

Commits: one per behavior ("Show nicknames in quotes.", "Add a former surname name part, shown in parentheses and italics.").

## 8. Tests

The Swift tests share a fixture `western` *(new, private in `PersonHeaderDisplayTests`)*: `NameFormatOrders` built by hand with D-4's two orders, `", "` after `surname` in sorted.

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/namevalue/registry_test.go` `TestPartTypesRegistry` | The keys in order equal `[prefix given nick surname_prefix surname former_surname suffix undetermined]` (replaces `len == 7`); `Validate` accepts a `former_surname` part | AC-2.5, D-3 |
| T-2 | Go | `core/database/nameformats/nameformats_test.go` `TestWesternSeedOrder` *(new)*; `TestSeedMatchesRegistryWestern` and `core/matchregistry` `TestWesternPatternCoversEveryPartType` (existing) | After `Install`, the default profile's `Natural` and `Sorted` `PartOrder.Type` lists equal D-4's; the existing tests pass with role `family` | D-4, D-5 |
| T-6 | Go | `core/matchregistry/registry_test.go` `TestFormerSurnamePairsAsFamily` *(new)* | AC-2.7's compare under `Default().Value` → `Contains`, a pair `smith`/`smith` with relation `Same` | AC-2.7 |
| T-3 | Swift | `NameValueDraftTests.swift` `formerSurnameIsAllowed` *(new)* | `NamePartType.isAllowed("former_surname")`; in `selectOptions`, `former_surname` comes right after `surname` | AC-2.1 |
| T-7 | Swift | `NameValueDraftTests.swift` `decorationRuns` *(new)* | `NamePartType.formerSurname.decoration!.runs("Smith")` → `[("(", false), ("Smith", true), (")", false)]`; `.nick` → `[("“", false), ("Jake", false), ("”", false)]`; a decoration whose format is `(%@` (nothing after the slot) → `[("(", false), ("Smith", true)]`; `.given.decoration == nil` | D-6, D-7 |
| T-4 | Swift | `PersonHeaderDisplayTests.swift` `nicknamesAreQuoted`, `formerSurnameIsParenthesized`, `unlistedFormerSurnameFollows` *(new)* | With `western`: AC-1.1, AC-1.2, AC-1.3 strings; `Mary (Smith) Jones` natural and `Jones, Mary (Smith)` sorted; with `western` minus `former_surname` → `Mary Jones (Smith)` | AC-1.1, AC-1.2, AC-1.3, AC-1.4, AC-2.6, D-1, D-2 |
| T-5 | Swift | `PersonHeaderDisplayTests.swift` `formerSurnameRunIsEmphasized` *(new)* | `attributed` for Mary, natural and sorted: the runs with `inlinePresentationIntent == .emphasized` are exactly one, text `Smith`; `titleSource` for Mary is `.name` with `styled` set and `text == String(styled.characters)`; for James Robins `== .name("James Robins")` | AC-2.2, AC-2.3, AC-2.4, D-9, D-10 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `TestPartTypesRegistry`, `nameformats_test.go`, any Swift test or golden that shows a name with a `nick` part (it gains quotes). Such a golden is updated to the quoted string and listed in the PR.

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left
- [ ] UI matches the brief inventory; no new kit primitive — not gated; no new component
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together) — none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; docs/deployment-plan/spike-11/completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated (if gated) — not gated
- [ ] Domain doc updated: seeded-vocabulary §4.1 / §4.3, structured-name-model §4.2a / §4.5, international-names
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

1. Create a new project (the seed carries the new rows). Add a Source with a person named given *James*, nick *Jake*, surname *Robins*; promote it.
2. Persons list: *Robins, James “Jake”*. Person page title: *James “Jake” Robins*. The Promote row and an omnibar search for *Jake* show the quotes (AC-1.4).
3. Add a person given *Mary*, Former surname *Smith*, surname *Jones*; the picker lists Former surname after Surname. Persons list: *Jones, Mary (Smith)* with *Smith* italic. Person page title: *Mary (Smith) Jones* with *Smith* italic in the display font. If *Smith* is upright, apply D-11's fallback and recheck.
4. Open an older project with a former-surname part (or set one): the title reads *Mary Jones (Smith)* (AC-2.6).

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| The emphasized run renders upright in the display font | D-11's binding fallback |
| `family` role makes a former surname cost points against records that carry only the married surname | Accepted (D-5); matching numbers are reviewed in Spike 11's goldens |
| The list row shows quotes the sort key does not have | §4.5 says so (STEP-5); the key sorts by the same words |

**Stop and report if:**

- D-11's fallback also renders upright.
- A test outside the files in §7 fails for a reason other than a name with a `nick` part now showing quotes.
- `nameformats.Install` or `TestWesternPatternCoversEveryPartType` fails after STEP-1 for a reason other than the new type.
- The host has no Xcode (see §7): stop after STEP-4's code and report.

## 12. Open questions

None.

## 13. PR

**Title:** S11-29 — Quote nicknames and add former surnames
**Body:** objective in one sentence; what shipped against US-1 and US-2; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-11/prs/S11-29-name-display.md`.
