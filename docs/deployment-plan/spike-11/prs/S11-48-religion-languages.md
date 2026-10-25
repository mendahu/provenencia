# S11-48 — Religion and languages spoken on a person

**Status:** Planned
**Spike / slice:** Spike 11 · Person vocabulary (one-off, user 2026-10-25)
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no new chrome: the Citation composer, the Properties page and the Person page already show any term Property bound to a person)
**Board:** —
**Depends on:** —
**Blocks:** S11-44 (its command copies these into catalogs made before this PR)
**Branch:** `s11-48-religion-languages`
**Version bump:** none — Spike 11 PRs leave `VERSION` at 0.2.0, and the product bump is cut when a release ships ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-seeded-vocabulary`](../../../../.agents/skills/add-seeded-vocabulary/SKILL.md)
**Rules that bind:** `.cursor/rules/go-tests.mdc`
**Context:** [`seeded-vocabulary.md`](../../../seeded-vocabulary.md) §3.2, §3.3, §3.9; [`conclusion-reconciliation.md`](../../../conclusion-reconciliation.md) (Property cardinality); [`interpretation-layer-data-model.md`](../../../interpretation-layer-data-model.md) (Property terms); working rules: "High tolerance for churn"
**Planned:** 2026-10-25 against `main` @ `bc673dd6`
**Review:** [`S11-48-religion-languages.review.md`](S11-48-religion-languages.review.md) — Fits with changes / Ready with changes, 1 round (findings fixed or logged)

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

A new project's person has two more Properties: **Religion** and **Languages spoken**. Both are term Properties, so "Catholic" on one census and "Catholic" on the next are the same value and can be counted together, and both keep every distinct value. Each starts with one product term, Catholic and English. The researcher adds the rest with the composer's **Add custom…** as records turn them up.

**Why now.** Records name a person's religion and languages (census columns, church registers, naturalization papers), and there is nowhere to transcribe them except free-text `remark`s or a researcher Property that can't aggregate. The user chose term Properties over an Organization Subject type on 2026-10-25. The Organization type is parked in [`ideas/organizations.md`](../../../ideas/organizations.md).

**What stays.** Matching: neither Property is weighed in Promote (`matchregistry` lists only the Properties it reads). Catalogs made before this PR: Install runs only at create. S11-44's command adds the two Properties to an existing catalog. Every other seeded Property, binding and term.

## 2. User stories

**US-1.** As a researcher transcribing a census, when a person's religion column reads "R.C.", I pick **Religion → Catholic** on the person, so every Catholic record of that person agrees.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | A catalog made by `onboarding.createCatalog` | Property `religion` (origin `provenencia`) exists: `value_type` `term`, cardinality `multiple`, label "Religion" |
| AC-1.2 | Same catalog | `religion` has exactly one product term: key `catholic`, label "Catholic" |
| AC-1.3 | Same catalog, person bindings | `religion` is bound to `person`, unlocked, sort order 5 |

**US-2.** As a researcher, when a person is recorded as speaking English and French, I record both under **Languages spoken**, so the person's Conclusion shows both.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | A catalog made by `createCatalog` | Property `languages_spoken` (origin `provenencia`) exists: `value_type` `term`, cardinality `multiple`, label "Languages spoken" |
| AC-2.2 | Same catalog | `languages_spoken` has exactly one product term: key `english`, label "English" |
| AC-2.3 | Same catalog, person bindings | `languages_spoken` is bound to `person`, unlocked, sort order 6 |

**US-3.** As a researcher, when two Sources give one person different values of a multiple term Property (Catholic in 1881, Anglican in 1901), the Conclusion keeps both, as it does for `toponym` and `interests`.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | `autoreconcile.Reconcile` with value type `term`, cardinality `multiple`, candidates `catholic` (source a), `catholic` (source b), `anglican` (source c) | Two kept rows (the two `catholic`, then `anglican`); state `StateMultiple` |
| AC-3.2 | Same, candidates `catholic` and a low-trust `anglican` | `catholic` kept, `anglican` `ReasonWeak`; state `StateSingle` (the confidence step still applies) |

## 3. Scope

| In | Covered by |
| --- | --- |
| `religion` and `languages_spoken` in `seedProperties`, `seedBindings`, `seedTerms` | STEP-1, AC-1.1 – AC-2.3 |
| One registry-driven Install test; the tests that restate the person binding order or the Property count read the registry or stop asserting it | STEP-1, STEP-2 |
| A term case in the reconciler's multiple-cardinality test | STEP-3, AC-3.1, AC-3.2 |
| `seeded-vocabulary.md` §3.2, §3.3, new §3.10 and §3.11; `conclusion-reconciliation.md` cardinality paragraph and **Multiple** bullet; the `add-seeded-vocabulary` Swift-step sentence | STEP-4 |

| Out | Home |
| --- | --- |
| Adding the two Properties to a catalog made before this PR | S11-44 (its command widens to seeded subject vocabulary, user 2026-10-25) |
| More product terms (other denominations, other languages) | not planned: the researcher adds `origin=user` terms; the product list grows when dogfood asks |
| Grouping terms (Wesleyan Methodist under Methodist) | not planned: [`ideas/term-management.md`](../../../ideas/term-management.md) is where a term hierarchy would go |
| Weighing religion or languages in Promote matching | not planned: unlisted Properties are never read (`subjectcompare_test.go` "AC-1.15 an unlisted property is never read") |
| Swift constants or L10n wording for the new terms | not planned: D-3 |
| Organizations as a Subject type (congregations, employers, schools, regiments, lodges) | [`ideas/organizations.md`](../../../ideas/organizations.md) |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Seed | `core/database/subjectvocab/registry.go` `seedProperties` | 22 Properties; the last person ones are `interests` (text, multiple) and `likeness` (depiction, multiple). No `religion` or `languages_spoken` |
| Seed | same file `seedBindings` | Person: `name` 0 (locked), `likeness` 1 (locked), `sex_at_birth` 2, `gender` 3, `interests` 4 |
| Seed | same file `seedTerms` | Terms for `sex_at_birth`, `gender`, `event_type`, `role`, `relationship_type`, `place_relationship_type` |
| Install | `core/database/subjectvocab/subjectvocab.go` `Install` | Upserts every registry list; nothing to change |
| Term values | `core/autoreconcile/modules.go` `moduleFor` | `ValueTypeTerm` → `keyModule{keyOf: termKey}`; cardinality is handled apart from value type in `autoreconcile.go` and `pipeline.go`. No seeded Property is term **and** multiple yet |
| Person page | `macos/App/Features/Conclusions/PersonDetailContent.swift` `rows` | One row per field with outcomes, except `name`; a term value shows its catalog label (`ReconciledValueDisplay.swift` `.term` case). No per-Property code |
| Composer | `macos/App/Features/CitationComposer/CitationComposerAuthoringBox.swift` | Term picker with **Add custom…** for any term Property |
| Swift pins | `api/ffi/swift_vocabulary_test.go` `TestSwiftVocabularyMatchesGo` | `seededTermEnums` lists only the Properties the app names terms of; a Property not in that list needs no Swift constant |
| Tests | `core/database/subjectvocab/subjectvocab_test.go` `TestInstallSeedsEngagementTravelInterestsAndGender`, `TestInstallSeedsLikeness`, `TestInstallSeedsFuneralAndPallbearer` | Each copies registry values (labels, value types, sort orders, locks) into literals and checks Install copied them. `TestSubjectVocab` already compares with `len(seedProperties)` |
| Registry read | `core/database/subjectvocab/subjectvocab.go` `BindingsForType` | Exported; returns a type key's seed bindings (registry plus `connectrules`). No caller today |
| Tests | `core/database/graphcache/vocab_test.go` `TestVocabularyBindings`, `TestVocabularyBindingsPluginTypeSeededKey` | `want` = `name, likeness, sex_at_birth, gender, interests` (twice), written by hand |
| Tests | `api/ffi/handlers/subject_defs_test.go` "lists seeded properties" | `len(resp.Properties) != 22`, then one `found…` boolean for each of five Properties. `gender` (term / single) and `interests` (text / multiple) already prove the wire carries value type and cardinality |
| Tests | `core/claimevidence/sections_test.go` `seedOrder` | A matching fixture nothing pins to the seed; no case puts values on the new Properties. Not touched |
| Tests | `core/autoreconcile/pipeline_test.go` `TestPipelineMultipleCardinality` | Two text cases; value type fixed to `ValueTypeText` in the loop |
| Docs | `docs/seeded-vocabulary.md` §3.2, §3.3, §3.9 | Lists the 22 Properties and person bindings; last term section is §3.9 `gender` |
| Docs | `docs/conclusion-reconciliation.md` cardinality paragraph ("Most facts that repeat or change over time…") | Says religion is better modelled as an event with a span. The **Multiple** bullet after it gives "researcher-created Properties such as languages spoken" as its example |
| Skill | `.agents/skills/add-seeded-vocabulary/SKILL.md`, paragraph "A seeded term the Mac app names has a Swift step too" | Ends "A new term-bearing Property needs the same two steps and an entry in that test's list", which reads as covering these two Properties; D-3 says they need neither |

**Precedent.** #382 (`3d701390`) seeded `gender` (term) and `interests` (text, multiple) on person: rows in `seedProperties`, `seedBindings` and `seedTerms`, an Install test, `seeded-vocabulary.md`, and the property count in `subject_defs_test.go` (`f06250ae`). Copy that shape. Do not copy its Swift and L10n step: that PR also seeded event types the app titles, and this one seeds no term the app names (D-3).

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | Keys `religion` and `languages_spoken`; labels "Religion" and "Languages spoken"; terms `catholic` "Catholic" and `english` "English" | The user's names; snake_case keys like `sex_at_birth` | `language` — the Property holds several, and the label says which |
| D-2 | Both cardinality `multiple` | The user's call for religion, 2026-10-25: a conversion keeps both values, with no conflict. Languages are several at once | `single` for religion — a conversion would read as mixed |
| D-3 | No Swift constant, no `L10n.PropertyTerm.wording`, no entry in `seededTermEnums` | The app names no term of either Property; the Person page and composer show the catalog label, as for researcher terms. `add-seeded-vocabulary` asks for the Swift step only for a term the Mac app names | Wording for Catholic and English — nothing reads it |
| D-4 | No `vocabkey` constant for either key | `vocabkey` holds the keys code reads by name; nothing does. Tests use string literals, as for `interests` | A constant with no reader |
| D-5 | Bindings unlocked, sort 5 and 6, after `interests` | Unlocked like `gender` and `interests`: nothing keys into the slot | Locked |
| D-6 | The descriptions say these are product term vocabularies the researcher extends: religion "Religion or denomination as recorded. More than one value is allowed. Product term vocabulary; add your own terms." and languages_spoken "Languages the person spoke. More than one value is allowed. Product term vocabulary; add your own terms." | Same voice as `gender` and `interests` | — |
| D-7 | `TestInstallMatchesRegistry` *(new)* checks every installed Property, term and binding against its registry row. `TestSeedReligionAndLanguagesSpoken` *(new)* checks the registry holds the D-1, D-2 and D-5 values. The three older Install tests stay as they are | `Install` is a loop over the registry, so one round-trip test covers every later seed row with no new test code. S11-44's command gets the same "catalog matches the registry" check (review M-1) | A fourth hand-copied Install test. Folding the older three in: they also check user-term create and `Neutral`, which stay where they are |
| D-8 | Tests outside `subjectvocab` stop restating the seed. `graphcache` builds its expected person order from `subjectvocab.BindingsForType("person")` sorted by `SortOrder`. The handler test drops its `!= 22` count and adds no `found…` booleans. `claimevidence`'s `seedOrder` is left alone | A seeded Property then costs no test edits outside the registry (review M-2), and `BindingsForType` gains its first caller | Updating four hand lists again, as #382 did |

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `religion`, `languages_spoken` rows *(new)* | registry rows | `gender`, `interests` rows in `registry.go` | They are registry rows of that list |
| `TestInstallMatchesRegistry` *(new)* | test | `TestInstallSeedsEngagementTravelInterestsAndGender` and two siblings; `TestSubjectVocab`'s `len(seedProperties)` | D-7: one generic check instead of a fourth copy |
| `TestSeedReligionAndLanguagesSpoken` *(new)* | test | — | Pins this PR's product values in the registry (D-1, D-2, D-5) |
| `seeded-vocabulary.md` §3.10, §3.11 *(new)* | doc sections | §3.9 `gender` | Same shape |

**Builds on.** `add-seeded-vocabulary`'s registry and create-time Install; the generic term picker, Person rows and reconciler term module.
**Deletes.** None — nothing becomes dead. The `conclusion-reconciliation.md` sentence on religion as an event is rewritten.
**Next adds.** S11-44's command adds these to old catalogs, and its test reuses `TestInstallMatchesRegistry`'s check. A third person term Property is three registry rows and a doc section, with no test edits outside `subjectvocab`.

### Schema

None — no catalog change. Properties and terms are rows in existing tables.

### Go core

| Package | Change |
| --- | --- |
| `core/database/subjectvocab` | `seedProperties`: two rows appended after `likeness` (D-1, D-2, D-6), `ValueType: propertymeta.ValueTypeTerm, Cardinality: propertymeta.CardinalityMultiple`. `seedBindings`: `{TypeKey: "person", PropertyKey: "religion", SortOrder: 5}` and `{TypeKey: "person", PropertyKey: "languages_spoken", SortOrder: 6}` after the `interests` row. `seedTerms`: a blank line after the last `place_relationship_type` term, then `{PropertyKey: "religion", Key: "catholic", Label: "Catholic"}`, a blank line, `{PropertyKey: "languages_spoken", Key: "english", Label: "English"}`; no new comment. The comment above `seedTerms` changes from "§3.4–3.9" to "§3.4–3.11" |

### FFI

None — `ListProperties` and the term RPCs already carry any Property.

### macOS

None — no Swift change (D-3).

### Strings

None — catalog labels, not L10n.

## 7. Steps

**STEP-1 — Seed the two Properties**
Files: `core/database/subjectvocab/registry.go`, `core/database/subjectvocab/subjectvocab_test.go`. Skill: `add-seeded-vocabulary`.
Do: add the rows in §6 Go core. Add `TestInstallMatchesRegistry` and `TestSeedReligionAndLanguagesSpoken` (T-1, T-2) after `TestInstallSeedsFuneralAndPallbearer`. `TestInstallMatchesRegistry` uses the setup of the tests above it (`database.Create(t.TempDir(), "t.provenencia")`, then `Install`) and checks three things:
- each `seedProperties` row: `properties.Lookup(c, key, vocabkey.OriginProvenencia)` has the registry's label, value type and cardinality (an empty registry cardinality reads as `propertymeta.CardinalitySingle`);
- each `seedTerms` row: `propertyterms.Lookup(c, propertyID, key, vocabkey.OriginProvenencia)` has the registry's label, `Neutral`, `Directed` and inverse key;
- each `seedTypes` row: `ListBindings(c, typeID)` holds the same property keys, sort orders and `Locked` as `BindingsForType(typeKey)`.

The stored twins are `properties.Property` `Label` / `ValueType` / `Cardinality`, `propertyterms.Term` `Label` / `Neutral` / `Directed` / `InverseKey`, and `subjectvocab.Binding` `Property.Key` / `SortOrder` / `Locked`.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/subjectvocab/...` passes.

**STEP-2 — Tests outside the registry stop restating it**
Files: `core/database/graphcache/vocab_test.go`, `api/ffi/handlers/subject_defs_test.go`.
Do: in both `vocab_test.go` tests, `want` is built from `subjectvocab.BindingsForType("person")` sorted by `SortOrder`, with the keys converted to `vocabkey.PropertyKey` (T-3). In `subject_defs_test.go` "lists seeded properties", delete the `len(resp.Properties) != 22` check. The five `found…` checks stay, and none is added (T-4). `core/claimevidence/sections_test.go` and `core/subjectcompare/subjectcomparetest` `Metas()` are not touched (matching is Out).
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/graphcache/... ./api/ffi/...` passes, `TestSwiftVocabularyMatchesGo` included.

**STEP-3 — Multiple term values reconcile apart**
Files: `core/autoreconcile/pipeline_test.go`.
STEP-1 does not change this test's result; this step is test-only.
Do: `TestPipelineMultipleCardinality`'s case struct gains `valueType string`; the two text cases set `propertymeta.ValueTypeText`; two term cases (T-5) use the `term` helper in `modules_test.go`; the loop passes `tc.valueType`.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/autoreconcile/...` passes.

**STEP-4 — Docs**
Files: `docs/seeded-vocabulary.md`, `docs/conclusion-reconciliation.md`, `.agents/skills/add-seeded-vocabulary/SKILL.md`, `docs/deployment-plan/spike-11/completed.md`, `docs/deployment-plan/spike-11/deployment-plan.md` (map row), this plan (status).
Do:
- `seeded-vocabulary.md` §3.2: after the `likeness` line, two lines in the block's column layout: `religion            term          religion or denomination as recorded; cardinality `multiple`; unlocked on person; terms §3.10` and `languages_spoken    term          languages the person spoke; cardinality `multiple`; unlocked on person; terms §3.11`. In the sentence after the block, "(§3.4–3.7 and §3.9)" becomes "(§3.4–3.7 and §3.9–3.11)".
- §3.3: after `person          interests           # text; multiple; unlocked`, the lines `person          religion            # term; multiple; unlocked` and `person          languages_spoken    # term; multiple; unlocked`.
- New §3.10 and §3.11 after §3.9, before the `---`, each in §3.9's shape: heading `## 3.10 Property terms: \`religion\`` with "Product-seeded **term keys** for a person's religion or denomination as a record states it. More than one value is allowed, so a change over time keeps both." and a text block `catholic       Catholic`; heading `## 3.11 Property terms: \`languages_spoken\`` with "Product-seeded **term keys** for the languages a person spoke. More than one value is allowed." and a text block `english        English`. Each ends with §3.9's last paragraph, word for word.
- `conclusion-reconciliation.md`, the paragraph "Most facts that repeat or change over time…": "occupations, residences, religion and military service as **events** with spans" becomes "occupations, residences and military service as **events** with spans"; after "Cardinality is for what is left: concurrent place names, and Properties researchers create." insert "Religion and languages spoken are term Properties with cardinality *multiple* (S11-48), so a person recorded Catholic and later Anglican keeps both." Nothing else in the paragraph changes. In the **Multiple** bullet, "(`toponym`: Montréal and Montreal, Tkaronto and Toronto; researcher-created Properties such as languages spoken)" becomes "(`toponym`: Montréal and Montreal, Tkaronto and Toronto; `religion` and `languages_spoken`; researcher-created Properties such as hobbies)".
- `add-seeded-vocabulary/SKILL.md`: the sentence "A new term-bearing Property needs the same two steps and an entry in that test's list." becomes "A new term-bearing Property whose terms the app names needs the same two steps and an entry in that test's list. One the app only shows by catalog label (`religion`, `languages_spoken`) needs neither."
- Map row S11-48 in `deployment-plan.md`: Status `Landed #N` (the Plan cell already links this file). `completed.md`: a new section at the top (newest first), `## S11-48 — Religion and languages spoken on a person`, then `**Landed YYYY-MM-DD** in [#N](https://github.com/mendahu/provenencia/pull/N).` and two sentences: new projects' persons take Religion and Languages spoken as term Properties, seeded with Catholic and English, and catalogs made earlier get them from S11-44's command. This plan: Status `Landed in [#N](https://github.com/mendahu/provenencia/pull/N) (YYYY-MM-DD)`, and a `**Landed:**` header line after **Planned** listing every difference from the plan ("None." when there is none).
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-11/prs/S11-48-religion-languages.md` reports no errors (warnings that `ideas/…` link text is not a repo-root path are expected; the links resolve).

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/database/subjectvocab/subjectvocab_test.go` `TestInstallMatchesRegistry` *(new)* | Installed catalog → every `seedProperties`, `seedTerms` and seeded-type binding row is in the catalog with the registry's values | AC-1.1, AC-1.2, AC-1.3, AC-2.1, AC-2.2, AC-2.3, D-7 |
| T-2 | Go | same file `TestSeedReligionAndLanguagesSpoken` *(new)* | Registry → `religion` and `languages_spoken` are term / multiple with D-1 labels; their only terms are `catholic` "Catholic" and `english` "English"; person bindings at 5 and 6, `LockedBinding` false | AC-1.1, AC-1.2, AC-1.3, AC-2.1, AC-2.2, AC-2.3, D-1, D-2, D-5 |
| T-3 | Go | `core/database/graphcache/vocab_test.go` `TestVocabularyBindings`, `TestVocabularyBindingsPluginTypeSeededKey` | Person binding order equals the registry's, now ending `religion`, `languages_spoken` | AC-1.3, AC-2.3, D-8 |
| T-4 | Go | `api/ffi/handlers/subject_defs_test.go` "lists seeded properties" | No count; the five `found…` checks pass | D-8 |
| T-5 | Go | `core/autoreconcile/pipeline_test.go` `TestPipelineMultipleCardinality` | "distinct terms both kept": `from("a", term(1, "t1", "catholic"))`, `from("b", term(2, "t1", "catholic"))`, `from("c", term(3, "t2", "anglican"))` → rows `{1,2}` kept, `{3}` kept, `StateMultiple`; "a weak distinct term still drops": `term(1, "t1", "catholic")`, `lowTrust(term(2, "t2", "anglican"))` → `{1}` kept, `{2}` `ReasonWeak`, `StateSingle` | AC-3.1, AC-3.2, D-2 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `TestSwiftVocabularyMatchesGo` (it must pass unchanged, D-3) and any test that counts seeded Properties or person bindings.

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left — none: no Swift change
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New — none: no brief
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together) — none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated (if gated) — not gated
- [ ] Domain doc updated if a rule in docs/*.md changed: seeded-vocabulary.md §3.2, §3.3, §3.10, §3.11; conclusion-reconciliation.md cardinality paragraph and Multiple bullet; add-seeded-vocabulary skill sentence
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

Run it if your session can build and launch the Mac app. If it can't, report §10 as "not run" in the DoD with the reason; the reviewer runs it before merge. The Go tests cover every AC without it.

1. Create a new project. Open the Properties page: Person lists Religion and Languages spoken after Interests, both unlocked.
2. In a Source's Citation composer, add an Observation on a person: Religion → Catholic. Add a second: Religion → **Add custom…** "Anglican". Add Languages spoken → English.
3. Promote the person. The Person page shows Religion with Catholic and Anglican, and Languages spoken with English.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| A test outside §4 counts seeded Properties or person bindings | `scripts/check-pr.sh` finds it; update its expected list in STEP-2 and name it in the Landed note |
| The Person page or composer special-cases term Properties by key | §4 checked both read any term Property; the manual check proves it |

**Stop and report if:**

- `TestInstallMatchesRegistry` fails on a seed row that predates this PR (Install does not store a registry value as written): report the row and field; do not change Install.

- `TestSwiftVocabularyMatchesGo` fails because of the new Properties (D-3 is wrong).
- The composer offers no term picker or no **Add custom…** for `religion` in the manual check.
- `Reconcile` does not keep two distinct term values under cardinality `multiple` (T-4 fails for a reason other than the test's own setup).
- A failing test needs a Swift or `matchregistry` change to pass.

## 12. Open questions

None.

## 13. PR

**Title:** S11-48 — Religion and languages spoken on a person
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-11/prs/S11-48-religion-languages.md`.
