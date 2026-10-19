# S11-29 — Quote nicknames and add former surnames

**Status:** Draft
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
**Context:** [`docs/structured-name-model.md`](../../../structured-name-model.md) §4.2 roles; [`docs/seeded-vocabulary.md`](../../../seeded-vocabulary.md) §4.1, §4.3; [`docs/ideas/international-names.md`](../../../ideas/international-names.md) "Name changes as a signal"; S9-D2 title anatomy in `macos/App/Features/Conclusions/ConclusionListRow.swift`
**Planned:** 2026-10-19 against `main` @ `84a9b31`
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

A Person's name says which parts are a nickname and which a former surname. A nickname shows in curly double quotes (*James “Jake” Robins*), the genealogy convention. A new product part type, **Former surname**, records a maiden or other earlier surname; it shows in parentheses (*Mary (Smith) Jones*), and on the Persons list and Person page titles the surname inside the parentheses is italic. Today a nickname reads as one more given name (*James Jake Robins*) and a maiden name has no part type of its own.

**Why now.** User request (2026-10-19). [`international-names.md`](../../../ideas/international-names.md) names maiden names in the conclusion as the intended path for name changes.

**What stays.** Stored parts and `form` are never rewritten; the decoration is display only. The Go list sort key (`conclusionheaders`) is unchanged. Existing catalogs are not migrated (user, 2026-10-19).

## 2. User stories

**US-1.** As a researcher, when a name has a nickname part, I see it in quotes, so that I can tell it from a middle name.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | Parts `given=James`, `nick=Jake`, `surname=Robins`, Western orders, natural | `James “Jake” Robins` |
| AC-1.2 | Same, sorted | `Robins, James “Jake”` |
| AC-1.3 | Two nick parts `Jake`, `Jim` | `James “Jake” “Jim” Robins` (each part quoted) |
| AC-1.4 | The omnibar hit and the Promote row for that Person | Show the quotes (plain text) |

**US-2.** As a researcher, I can type a name part as **Former surname**, and see it in parentheses and italics on the Person's title, so that a maiden name is recorded and reads as earlier.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | The part-type picker in the name editor | Lists *Former surname* after *Surname* |
| AC-2.2 | Parts `given=Mary`, `former_surname=Smith`, `surname=Jones`, natural, new catalog | Text `Mary (Smith) Jones`; the person detail title renders *Smith* italic, the rest upright |
| AC-2.3 | Same, sorted (Persons list row) | Text `Jones, Mary (Smith)`; *Smith* italic |
| AC-2.4 | A name with no former surname | `ConclusionTitleSource` equals `.name("James Robins")` as today (no styled payload) |
| AC-2.5 | Go `namevalue.Validate` on a part typed `former_surname` | Accepted |
| AC-2.6 | A catalog created before this PR (no seed rows for the new type), natural | `Mary Jones (Smith)`: an unlisted type follows, as `NameValueDisplay` does today |
| AC-2.7 | Matching a record *Mary Smith* against a conclusion with `former_surname=Smith`, `surname=Jones` in a new catalog | *Smith* pairs as a family word (role `family`) |

## 3. Scope

| In | Covered by |
| --- | --- |
| `former_surname` in the Go part-type registry, Western seed parts and roles, match registry Western roles | STEP-1, AC-2.5, AC-2.7 |
| `NamePartType.formerSurname`, its label, picker entry | STEP-2, AC-2.1 |
| `NameValueDisplay` quotes nick parts and parenthesizes former surnames, everywhere it is used | STEP-3, AC-1.1 – AC-1.4, AC-2.6 |
| An attributed title with the former surname italic, on the Persons list row and Person detail title | STEP-4, AC-2.2 – AC-2.4 |
| Docs: seeded-vocabulary §4.1 / §4.3, structured-name-model §4.2 roles, international-names | STEP-5 |

| Out | Home |
| --- | --- |
| Migrating existing catalogs' name-format rows | not planned (user, 2026-10-19: new catalogs only) |
| Italics anywhere but the Persons list row and Person detail title (value rows, omnibar, Promote, Evidence) | not planned; those surfaces are plain text and keep the parentheses |
| Quotes in Go `nameformats.Format` (list sort key) | not planned; a sort key carries no punctuation |
| The reconciler treating a former surname and a surname as related | not planned |
| Nickname alias folding | S11-28 |
| GEDCOM import / export of former surnames | not planned; no GEDCOM path exists |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Part-type registry (Go) | `core/namevalue/namevalue.go` consts; `core/namevalue/registry.go` `partTypeRegistry` | Seven types: prefix 0, given 1, nick 2, surname_prefix 3, surname 4, suffix 5, undetermined 6 |
| Western seed | `core/database/nameformats/registry.go` `seedProfiles` | Natural: prefix, given, nick, surname_prefix, surname, suffix. Sorted: surname_prefix, surname (`", "`), prefix, given, nick, suffix. Roles for all seven types. `Install` fails when the default profile lacks a role for a registry type |
| Match registry | `core/matchregistry/value.go` `Names.Western.PartRoles` | Roles for the seven types and `""` |
| Swift part types | `macos/App/Features/Names/NamePartType.swift` `NamePartType` | Seven cases; `selectOptions` builds the picker from `allCases` |
| Strings | `macos/App/Platform/L10n.swift` `L10n.NameValue.partType*`; `macos/App/Resources/Localizable.xcstrings` | `nameValue.part.type.<key>` labels |
| Display | `macos/App/Features/Names/NameValueDisplay.swift` `NameValueDisplay.join` | Groups parts by type, joins words with spaces, groups in the profile's order with trailing separators; unlisted types follow in stored order. Returns `String`; no per-type decoration |
| Title source | `macos/App/Features/Conclusions/ConclusionListRow.swift` `ConclusionTitleSource` | `.name(String)`, `.label(String)` (rendered italic), `.ref(String)`; `text` |
| Person title | `macos/App/Features/Conclusions/PersonHeaderDisplay.swift` `titleSource(name:entity:style:orders:)` | The one place a Person's title is made; used by `PersonsList.titleSource` (sorted) and `PersonDetailContent` (natural) |
| Title rendering | `ConclusionListRow.titleText(_:)`; `ConclusionDetailPage.swift` `title` (around line 430) | `Text(verbatim: text)` for `.name` |
| Tests | `macos/ProvenenciaTests/PersonHeaderDisplayTests.swift`, `PersonsListTests.swift`, `PersonDetailTests.swift`, `NameValueDraftTests.swift`; `core/namevalue/registry_test.go` `TestPartTypesRegistry` (asserts `len == 7`); `core/matchregistry/registry_test.go` `TestWesternPatternCoversEveryPartType`; `core/database/nameformats/nameformats_test.go` | Parity tests force a role for every registry type |

**Precedent.** `nick` is the nearest part type: copy every place it appears (Go const and registry row, seed part rows in both styles, seed role, match registry role, Swift case, L10n key, xcstrings entry). The `.label` title case shows how a title goes italic in both the list row and the detail page.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | Nickname display: each `nick` part wrapped in L10n `nameValue.display.nick` = `“%@”` | User call (2026-10-19); GEDCOM tools, Gramps and RootsMagic quote nicknames; localizable for « » or „ “ | Parentheses — conventionally a maiden name |
| D-2 | Former surname display: each `former_surname` part wrapped in L10n `nameValue.display.formerSurname` = `(%@)`; italic only on the word inside, only on the Persons list row and Person detail title | User call (2026-10-19): parentheses + italic. Plain-text surfaces keep the parentheses, so the meaning survives without italics | Italic only — lost on plain-text surfaces and close to the italic working-label title (S9-D2); *née* — gendered and language-bound |
| D-3 | Key `former_surname`; label "Former surname"; Sort 5 (after surname); suffix 6, undetermined 7 | Generic for maiden, earlier married or changed names (user) | `maiden` — narrower |
| D-4 | Western order: natural prefix, given, nick, **former_surname**, surname_prefix, surname, suffix; sorted surname_prefix, surname (`", "`), prefix, given, nick, **former_surname**, suffix | *Mary (Smith) Jones* / *Jones, Mary (Smith)* | After the surname — reads as a second surname |
| D-5 | Role `family` in the seed roles and the match registry | A former surname is a family name for matching; the record *Mary Smith* should pair it ([international-names](../../../ideas/international-names.md)) | `ignored` — throws away the maiden-name signal; a new role — needs a comparer change |
| D-6 | Decoration lives in Swift on `NamePartType`: `displayFormat: LocalizedStringResource?` *(new)* and `isEmphasized: Bool` *(new)*. Go has no display decoration | Display is Swift's job; Go `Format` is a sort key | A catalog column for decoration — schema for two constants |
| D-7 | `NameValueDisplay` builds runs `[(text, emphasized)]` once; `string(...)` concatenates them; `attributed(for:style:orders:) -> AttributedString` *(new)* sets `inlinePresentationIntent = .emphasized` on emphasized runs | One join for both outputs; every existing caller keeps `String` | A second join for attributed text — a near-copy |
| D-8 | `ConclusionTitleSource.name(String, styled: AttributedString? = nil)`. `PersonHeaderDisplay.titleSource` sets `styled` only when a run is emphasized | Existing `.name("…")` constructions and test equalities keep working; one payload for both views | A new `.styledName` case — every switch grows a case; always setting `styled` — breaks AC-2.4's equality |
| D-9 | Both title views render `Text(styled)` when `styled` is set, else `Text(verbatim: text)`, keeping today's modifiers | The detail page's display font carries through; the run adds italic | `.italic()` on the whole title — that is the label treatment |
| D-10 | No migration; old catalogs show the type after the listed ones (AC-2.6) and match it untyped | User call (2026-10-19) | Backfill migration |

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `namevalue.PartTypeFormerSurname` *(new)* | const | `PartTypeNick` | Extends the registry |
| `NamePartType.formerSurname`, `displayFormat`, `isEmphasized` *(new)* | enum case, properties | `NamePartType.label` | Extends the one Swift part-type registry |
| `NameValueDisplay.attributed(for:style:orders:)` *(new)* | func | `NameValueDisplay.string(for:style:orders:)` | Same runs, different output (D-7) |
| `L10n.NameValue.partTypeFormerSurname`, `displayNick(_:)`, `displayFormerSurname(_:)` *(new)* | keys | `L10n.NameValue.partTypeNick`; `L10n.Workspace.personRowAccessibility` (format pattern) | New copy |
| `ConclusionTitleSource.name` `styled` payload *(new)* | enum payload | `.label` italic | D-8 |

**Builds on.** `namevalue` registry, `nameformats` seed, `matchregistry` Western pattern, `NamePartType`, `NameValueDisplay`, `PersonHeaderDisplay`.
**Deletes.** None — nothing becomes dead.
**Next adds.** Another decorated part type is one `displayFormat` / `isEmphasized` entry on `NamePartType`.

### Schema

None — part types are a compiled registry; `name_format_profile_parts.part_type` and `name_format_profile_roles.part_type` are open TEXT. New catalogs get the rows from the seed.

### Go core

| Package | Change |
| --- | --- |
| `core/namevalue` | `PartTypeFormerSurname = "former_surname"` *(new)*; registry row `{Key: PartTypeFormerSurname, L10nKey: "nameValue.part.type.former_surname", Sort: 5}`; suffix Sort 6, undetermined 7. `registry_test.go`: `len == 8` |
| `core/database/nameformats` (`registry.go`) | Part rows per D-4; role `{PartTypeFormerSurname, RoleFamily}` |
| `core/matchregistry` (`value.go`) | `namevalue.PartTypeFormerSurname: valuecompare.RoleFamily` in `Western.PartRoles` |

### FFI

None — part types and name-format rows already cross as strings.

### macOS

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Part type | `NamePartType.formerSurname = "former_surname"` | `macos/App/Features/Names/NamePartType.swift` | Case after `surname`; `label` → `L10n.NameValue.partTypeFormerSurname`; `displayFormat`: `.nick` → `L10n.NameValue.displayNick`, `.formerSurname` → `L10n.NameValue.displayFormerSurname`, others nil; `isEmphasized` true only for `.formerSurname` |
| Runs | `NameValueDisplay` | `macos/App/Features/Names/NameValueDisplay.swift` | `join` returns runs. A group of a decorated type is each part's words formatted with `L10n.format(displayFormat, word)`, joined by spaces. For an emphasized type the run splits into the literal prefix, the word (emphasized) and the suffix: format with a placeholder, split on it. Unknown types are plain |
| Title payload | `ConclusionTitleSource.name(String, styled: AttributedString? = nil)` | `ConclusionListRow.swift` | `text` returns the String. Update every `case .name(let text)` match to `case .name(let text, _)` or bind `styled` |
| Person title | `PersonHeaderDisplay.titleSource(name:…)` | `PersonHeaderDisplay.swift` | Builds `attributed`; passes it as `styled` when any run is emphasized |
| List row | `ConclusionListRow.titleText` | same file | `.name(let text, let styled)`: `styled.map { Text($0) } ?? Text(verbatim: text)` |
| Detail title | `ConclusionDetailPage` `title` | `ConclusionDetailPage.swift` | Same, with today's font, tracking, colour and header trait |

States: a name with only a former surname → `(Smith)`, italic inside. Accessibility: VoiceOver reads `text` (with quotes and parentheses); `ConclusionListRow.accessibilityLabel` is unchanged.

### Strings

L10n keys *(new)*, each in `L10n.NameValue` and `Localizable.xcstrings`:

- `nameValue.part.type.former_surname` = "Former surname" — comment "NameValue part type: former_surname"
- `nameValue.display.nick` = "“%@”" — comment "A nickname inside a displayed name; the argument is the nickname"
- `nameValue.display.formerSurname` = "(%@)" — comment "A former surname inside a displayed name; the argument is the surname"

## 7. Steps

**STEP-1 — Go: the former surname part type**
Files: `core/namevalue/namevalue.go`, `core/namevalue/registry.go`, `core/namevalue/registry_test.go`, `core/database/nameformats/registry.go`, `core/database/nameformats/nameformats_test.go`, `core/matchregistry/value.go`. Skills: `add-name-value`, `add-match-rule`.
Do: D-3, D-4, D-5. Add T-1 and T-2.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/namevalue/ ./core/database/nameformats/ ./core/matchregistry/ ./core/valuecompare/` passes.

**STEP-2 — Swift: the part type and its label**
Files: `NamePartType.swift`, `L10n.swift`, `Localizable.xcstrings`, `NameValueDraftTests.swift`. Skill: `add-localized-string`.
Do: the case, label, `displayFormat`, `isEmphasized`, and the three keys. Add T-3.
Check: `python3 scripts/check-localizable-xcstrings.py` passes; T-3 passes.

**STEP-3 — Swift: quotes and parentheses in every name display**
Files: `NameValueDisplay.swift`, `PersonHeaderDisplayTests.swift`. Skill: `add-swift-test`.
Do: D-1, D-2, D-7 (runs, `string`, `attributed`). Add T-4.
Check: the Xcode test run in `scripts/check-pr.sh` passes T-4 and the existing `PersonHeaderDisplayTests`, `ObservationValueDisplayTests`, `CitationComposerModelTests`.

**STEP-4 — Swift: italic former surname on the Person titles**
Files: `ConclusionListRow.swift`, `ConclusionDetailPage.swift`, `PersonHeaderDisplay.swift`, `PersonsListTests.swift`, `PersonDetailTests.swift`. Skill: `add-ui-component` (no new component; confirms the title stays a snowflake of `ConclusionListRow`).
Do: D-8, D-9. Add T-5.
Check: Xcode tests pass; manual check steps 2–3.

**STEP-5 — Docs**
Files: `docs/seeded-vocabulary.md` §4.1 (add `former_surname -- no GEDCOM part tag (GEDCOM 7 records a NAME with TYPE MAIDEN or BIRTH)`) and §4.3 (both orders); `docs/structured-name-model.md` §4.2 roles sentence (`former_surname` → `family`); `docs/ideas/international-names.md` (the former surname part type exists; the matching signal is still open); `deployment-plan.md` map row; `completed.md`; this plan.
Do: per the Definition of done.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-11/prs/S11-29-name-display.md` reports no errors; links resolve.

Commits: one per behavior ("Show nicknames in quotes.", "Add a former surname name part, shown in parentheses and italics.").

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/namevalue/registry_test.go` `TestPartTypesRegistry` | `len == 8`; `KnownPartType("former_surname")`; sort order holds; `Validate` accepts a `former_surname` part | AC-2.5, D-3 |
| T-2 | Go | `core/database/nameformats/nameformats_test.go` (existing Install/Load test) + `core/matchregistry/registry_test.go` `TestWesternPatternCoversEveryPartType` | Installed natural order lists `former_surname` between `nick` and `surname_prefix`; sorted between `nick` and `suffix`; role `family` in both seed and registry. One `valuecompare` name compare: *Mary Smith* vs `given=Mary\|former_surname=Smith\|surname=Jones` under the Western pattern pairs *Smith* | AC-2.7, D-4, D-5 |
| T-3 | Swift | `NameValueDraftTests.swift` `formerSurnameIsAllowed` *(new)* | `NamePartType.isAllowed("former_surname")`; `selectOptions` lists it right after Surname | AC-2.1 |
| T-4 | Swift | `PersonHeaderDisplayTests.swift` `nicknamesAreQuoted`, `formerSurnameIsParenthesized`, `unlistedFormerSurnameFollows` *(new)* | AC-1.1, AC-1.2, AC-1.3 strings; `Mary (Smith) Jones` / `Jones, Mary (Smith)` with Western-shaped orders; with orders lacking the type → `Mary Jones (Smith)` | AC-1.1, AC-1.2, AC-1.3, AC-2.3, AC-2.6, D-1, D-2 |
| T-5 | Swift | `PersonHeaderDisplayTests.swift` `formerSurnameRunIsEmphasized` *(new)*; `PersonsListTests.swift` existing `.name("James Robins")` equality | `attributed(...)` has exactly one run with `inlinePresentationIntent == .emphasized`, text `Smith`; `titleSource` for Mary carries `styled`; for James Robins `styled == nil` | AC-2.2, AC-2.3, AC-2.4, D-7, D-8 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `TestPartTypesRegistry`, `nameformats_test.go` role counts, Swift tests matching `case .name(let text)`.

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
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated (if gated) — not gated
- [ ] Domain doc updated: seeded-vocabulary §4.1 / §4.3, structured-name-model §4.2, international-names
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

1. Create a new project (the seed carries the new rows). Add a Source with a person named given *James*, nick *Jake*, surname *Robins*; promote it.
2. Persons list: *Robins, James “Jake”*. Person page title: *James “Jake” Robins*. Omnibar search for *Jake* shows the same text (AC-1.4).
3. Add a person given *Mary*, Former surname *Smith*, surname *Jones*; the picker lists Former surname after Surname. Persons list: *Jones, Mary (Smith)* with *Smith* italic. Person page title: *Mary (Smith) Jones* with *Smith* italic in the display font.
4. Open an older project with a former-surname part (or set one): the title reads *Mary Jones (Smith)* (AC-2.6).

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| The display font has no italic face and the run renders upright | The `.label` title already renders italic in the same font (`ConclusionDetailPage`); manual check step 3 |
| `family` role makes a former surname cost points against records that carry only the married surname | Accepted (D-5); matching numbers are reviewed in Spike 11's goldens |

**Stop and report if:**

- `Text(AttributedString)` with `inlinePresentationIntent = .emphasized` renders upright in manual check step 3.
- A test outside the files in §7 fails because of the quotes (a golden string with a nick part); list it instead of changing it.
- `nameformats.Install` or `TestWesternPatternCoversEveryPartType` fails after STEP-1 for a reason other than the new type.

## 12. Open questions

None.

## 13. PR

**Title:** S11-29 — Quote nicknames and add former surnames
**Body:** objective in one sentence; what shipped against US-1 and US-2; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-11/prs/S11-29-name-display.md`.
