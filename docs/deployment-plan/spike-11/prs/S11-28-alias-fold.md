# S11-28 — Fold a nickname into the given name it stands for

**Status:** Draft
**Spike / slice:** Spike 11 · Slice 2 — Value compare, everywhere (follow-up to S11-06)
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no new chrome: the reconciler's fold rule changes in Go; pages render the cache as they do today)
**Board:** —
**Depends on:** S11-06 (landed [#456](https://github.com/mendahu/provenencia/pull/456))
**Blocks:** —
**Branch:** `s11-28-alias-fold`
**Version bump:** none — Spike 11 PRs leave `VERSION` at 0.2.0, and the product bump is cut when a release ships ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-match-rule`](../../../../.agents/skills/add-match-rule/SKILL.md), [`add-name-value`](../../../../.agents/skills/add-name-value/SKILL.md)
**Rules that bind:** `.cursor/rules/go-tests.mdc`, `.cursor/rules/match-rules.mdc`, `.cursor/rules/name-values.mdc`
**Context:** [`docs/conclusion-reconciliation.md`](../../../conclusion-reconciliation.md) §7 (grouping policy), §7.2; [S11-06 plan](S11-06-migrate-reconciler.md); working rules "The reconciler keeps its own decisions", "High tolerance for churn"
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

A Person whose records say *Chris Robins* and *Christopher Robins* shows one given name, *Christopher*, on the Persons list and the Person page. Today the reconciler keeps both (`given: [Christopher, Chris]`) because S11-06 decided that an alias never folds. The user asked (2026-10-19) that the product nickname and abbreviation tables fold a shorter given name into the longer name it stands for, the way an initial folds into the name it begins.

**Why now.** Dogfood: every Person recorded once by a nickname and once by a formal name reads as two given names.

**What stays.** Accent folds never fold (*Jose* stays beside *José*). Majority still only outvotes spelling variants. Surnames never alias-fold. Promote scoring is unchanged: aliases already score there.

## 2. User stories

**US-1.** As a researcher, when one record calls a person *Chris* and another *Christopher*, I see one given name, *Christopher*, so that the title reads as one name.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | Candidates `given=Chris\|surname=Robins` and `given=Christopher\|surname=Robins` | One value `given=Christopher\|surname=Robins`, both members, state merged |
| AC-1.2 | `given=Wm\|surname=Robins` and `given=William\|surname=Robins` | One value `given=William\|surname=Robins` |
| AC-1.3 | `given=Chris\|given=R\|surname=Robins` and `given=Christopher\|given=Robert\|surname=Robins` | One value `given=Christopher\|given=Robert\|surname=Robins` |
| AC-1.4 | `given=Chris` beside `given=Christopher` (1 record) and `given=Christina` (2 records) | Chris folds into Christina, the best supported; the name is `given=Christina\|given=Christopher` (the ambiguous-initial rule, applied to an alias) |
| AC-1.5 | `given=Jose` and `given=José\|given=Kenneth` | Unchanged: `given=Jose\|given=José\|given=Kenneth` (an accent fold never folds) |
| AC-1.6 | `nick=Chris` and `given=Christopher` | Unchanged: `given=Christopher\|nick=Chris` (a fold never crosses part types) |
| AC-1.7 | `surname=Bill` and `surname=William` (with an alias set holding both) | Unchanged: both stay; a surname never alias-folds |
| AC-1.8 | Alias pair of equal length: `given=Chris` and `given=Kris` (one set) | Both stay; no fold without a longer side |

**US-2.** As the engine, the reconciler cache reads the product alias tables, so that the fold happens on every catalog and the cache changes only when `CacheVersion` does.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | A catalog with two Observations naming the same person *Chris Robins* and *Christopher Robins*, joined in one entity | `autoreconciler` writes one displayed name `Christopher Robins` |
| AC-2.2 | `autoreconciler.CacheVersion` | `18`; a cache written at 17 rebuilds on open |

## 3. Scope

| In | Covered by |
| --- | --- |
| `nameModule.fold` accepts an alias or abbreviation pair whose `from` word is shorter than its `into` word, for part types whose Western role is given or nick | STEP-1, AC-1.1 – AC-1.8 |
| `autoreconciler.valueRules` carries the product alias tables; `CacheVersion` 18 | STEP-2, AC-2.1, AC-2.2 |
| `Compatible` follows the new fold (it calls `fold`); its alias test flips | STEP-1, T-3 |
| Docs: grouping policy and §7.2 fold rule, `names.go` header | STEP-3 |

| Out | Home |
| --- | --- |
| A canonical marker on alias sets, so *Abel* / *Ebbie* folds the right way | not planned; [`docs/ideas/international-names.md`](../../../ideas/international-names.md) gets a line in STEP-3 |
| The reconciler reading a catalog's own alias rows (an alias editor) | not planned; no alias editor exists |
| Nickname quotes, former surnames, italics on the title | S11-29 |
| Folding across part types (`nick=Chris` into `given=Christopher`) | not planned |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Fold rule | `core/autoreconcile/names.go` `nameModule.fold` | Folds only when `valuecompare.Compare(from, into)` is `Contains` with `Broader == SideA` and every pair is `Same` or cause `initial`. `Chris` vs `Christopher` is `Equivalent` with one pair `Equivalent`/`alias`, so it never folds. The type argument is ignored (`_ string`) |
| Unit compare value | `core/autoreconcile/names.go` `nameUnit.compareValue` | Strips part types, so every word is untyped (role `untyped`); the product alias tables allow `untyped`, so alias pairs are found for every part type including surnames |
| Header comment | `core/autoreconcile/names.go` package comment, "Fold" bullet | "An accent fold or an alias never folds." |
| Compatible | `core/autoreconcile/modules.go` `Compatible` | Two values are compatible when every shared unit has the same key or one folds into the other. Called from `core/database/promotealign/exhibits.go` |
| Reconciler rules | `core/database/autoreconciler/autoreconciler.go` `valueRules` | `matchregistry.Default().Value`: no alias tables, by design, "so the cache changes only when CacheVersion does" |
| Cache version | `core/database/autoreconciler/autoreconciler.go` `CacheVersion` | 17, with a changelog comment above it |
| Product alias seed | `core/database/namealiases/registry.go` `AliasTables`; `namealiases.go` `withAliases` (unexported) | The abbreviations and nicknames tables, roles given / nick / untyped |
| Western roles | `core/matchregistry/value.go` `Names.Western.PartRoles` | `given` → given, `nick` → nick, `surname` → family, `undetermined` and `""` → untyped |
| Tests | `core/autoreconcile/names_test.go` `TestReconcileNames` (group "fold"), `TestReconcileNamesWords` ("an accent-folded pair never folds"); `modules_test.go` `TestCompatibleIgnoresAliases`; `core/database/autoreconciler/autoreconciler_test.go` `TestReconciledNames` | `testRules` (in `autoreconcile_test.go`) has no aliases. `TestCompatibleIgnoresAliases` asserts Jim Robins is not compatible with James Kenneth Robins |

Value compare's output for the cases, checked at the planning base with the product rules plus a nickname table: `Chris` / `Christopher` → `equivalent`, one pair `equivalent alias`; `Chris, R` / `Christopher, Robert` → `contains`, `Broader = a`, pairs `alias` and `initial`; `Wm` / `William` → `equivalent`, pair cause `abbreviation`; `Jose` / `José` → `equivalent`, pair cause `fold`.

**Precedent.** S11-06 ([#456](https://github.com/mendahu/provenencia/pull/456), commits `fc01031a`, `1b407323`) changed the same fold and outvote rules: copy its table-driven cases in `names_test.go` and its `CacheVersion` changelog line. Do not copy its `outvotes` change; outvoting is not touched.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | A unit folds into another when `Compare(from, into)` is `Contains` with `Broader == SideA`, or `Equivalent`; every pair is `Same`, cause `initial`, or an alias pair (cause `alias` or `abbreviation`) whose `from` word has fewer runes than its `into` word; and at least one pair is not `Same` | The shorter word is the nickname in both product tables (*Chris*, *Wm*, *Bill*); `Equivalent` covers a unit with no unpaired words, `Contains` one with extra words | Folding into the set's canonical name — the catalog does not store which token is canonical (`name_alias_members` has no order), and the user accepted "shorter folds into longer" (2026-10-19) |
| D-2 | Equal length (*Chris* / *Kris*) or a longer `from` word does not fold | A fold needs a direction; with none both names stay, as today | Alphabetical tie-break — arbitrary |
| D-3 | Alias folding applies only when the unit's part type maps to role `given` or `nick` in `rules.Names.Western.PartRoles`; other types fold only on `Same` and initials, as today | The alias tables are given-name tables (roles given / nick / untyped); the reconciler strips roles before comparing, so without this a surname *Bill* would fold into *William* | Leaving it to the alias roles — the reconciler passes untyped words, which every table allows; hard-coding the strings `"given"` / `"nick"` in the module — the pattern already maps types to roles |
| D-4 | Both nickname and abbreviation tables fold | User call (2026-10-19): "Just Christopher", nicknames and abbreviations both fold | Nicknames only |
| D-5 | The folded name disappears from the displayed value; it is not turned into a `nick` part | User call (2026-10-19); matches how an initial folds | Showing *Christopher "Chris"* — a name no source wrote |
| D-6 | `autoreconciler.valueRules` = the shipped registry plus the **product** alias tables (`namealiases.AliasTables()`), through `namealiases.ProductRules(base)` *(new)* | Keeps the cache's rule that it changes only when `CacheVersion` does; no catalog edits alias rows today | Reading the catalog's alias rows — the cache would go stale on an alias change with no rebuild hook |
| D-7 | `Compatible` changes with the fold; `TestCompatibleIgnoresAliases` becomes `TestCompatibleFoldsAliases` and asserts Jim Robins is compatible with James Kenneth Robins | `Compatible` is defined as "the same value as the pipeline would see it"; Promote's exhibits (deleted in S11-19) agree with the reconciler | Keeping `Compatible` alias-blind through a flag — a second fold rule |
| D-8 | `CacheVersion` 18 with the changelog line `18: a shorter alias of a given name folds into the longer (S11-28).` | Cached values change | — |
| D-9 | No new registry knob | The rule has no number: rune length and the cause codes decide | A minimum length — nothing needs tuning |

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `namealiases.ProductRules(base valuecompare.Rules) valuecompare.Rules` *(new)* | func | `namealiases.Rules` (catalog / graph tables), `withAliases` | `Rules` reads catalog rows; this one reads `AliasTables()`. It reuses `withAliases` |
| `aliasFold(from, into string) bool` *(new, unexported, `names.go`)* | func | `valuecompare.SpellingSimilarity` use in `outvotes` | A rune-length test on one pair; nothing similar exists |
| `aliasRules` *(new, test var in `names_test.go`)* | helper | the inline index in `TestCompatibleIgnoresAliases` | Moves that table to one var both tests use |

**Builds on.** S11-06's fold in `nameModule`; value compare's `Detail.Parts` causes; the Western pattern in `matchregistry`; `namealiases.withAliases`.
**Deletes.** The "an accent fold or an alias never folds" sentence in `names.go` and in §7 and §7.2 of `conclusion-reconciliation.md`; replaced in STEP-3. The map's working rule was amended at planning. The inline alias table in `TestCompatibleIgnoresAliases`.
**Next adds.** A canonical token per alias set would replace D-1's length test with one check in `aliasFold`.

### Schema

None — no catalog change.

### Go core

| Package | Change |
| --- | --- |
| `core/autoreconcile` (`names.go`) | `fold(typ string, from, into unit) bool` reads `typ`. Sketch below. Header comment "Fold" bullet rewritten (STEP-3 text) |
| `core/database/namealiases` (`namealiases.go`) | `func ProductRules(base valuecompare.Rules) valuecompare.Rules { return withAliases(base, AliasTables()) }` *(new)* with a doc comment saying the reconciler cache reads it |
| `core/database/autoreconciler` (`autoreconciler.go`) | `var valueRules = namealiases.ProductRules(matchregistry.Default().Value)`; comment says "the shipped registry and the product alias tables, never a catalog's own rows or name formats". `CacheVersion = 18` with D-8's line |

```go
func (m nameModule) fold(typ string, from, into unit) bool {
	res, det := valuecompare.Compare(from.data.(nameUnit).cmp, into.data.(nameUnit).cmp, m.rules)
	switch {
	case res.Relation == valuecompare.Contains && det.Broader == valuecompare.SideA:
	case res.Relation == valuecompare.Equivalent:
	default:
		return false
	}
	role := m.rules.Names.Western.PartRoles[typ]
	aliases := role == valuecompare.RoleGiven || role == valuecompare.RoleNick
	changed := false
	for _, p := range det.Parts {
		switch {
		case p.Relation == valuecompare.Same:
		case p.Cause == valuecompare.CauseInitial:
			changed = true
		case aliases && (p.Cause == valuecompare.CauseAlias || p.Cause == valuecompare.CauseAbbreviation) && aliasFold(p.A.Text, p.B.Text):
			changed = true
		default:
			return false
		}
	}
	return changed || res.Relation == valuecompare.Contains
}

// aliasFold: the from word is the shorter, by runes.
func aliasFold(from, into string) bool { return utf8.RuneCountInString(from) < utf8.RuneCountInString(into) }
```

`typ` is the trimmed part type key (`readName` trims it), which is the `PartRoles` key. `p.A` is the `from` side because `from` is the first argument to `Compare`.

### FFI

None — the cache shape is unchanged.

### macOS

None — pages render the cache.

### Strings

None —.

## 7. Steps

**STEP-1 — Fold aliases of given names**
Files: `core/autoreconcile/names.go`, `core/autoreconcile/names_test.go`, `core/autoreconcile/modules_test.go`. Skill: `add-match-rule` (no knob; read it for the layer rules).
Do: implement D-1 – D-3 per the sketch. Add `aliasRules` (product `matchregistry.Default().Value` with a `NewAliasIndex` over two tables: nickname `{christopher, chris, kit}`, `{christina, chris}`, `{chris, kris}`, `{william, bill}`, `{james, jim}`; abbreviation `{william, wm}`; roles given / nick / untyped). Add `TestReconcileNamesAliases` *(new)* with T-1's rows. Rename `TestCompatibleIgnoresAliases` to `TestCompatibleFoldsAliases` and flip it per D-7, reading `aliasRules`.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/autoreconcile/` passes, including the untouched `TestReconcileNamesWords` row "an accent-folded pair never folds".

**STEP-2 — The reconciler cache reads the product alias tables**
Files: `core/database/namealiases/namealiases.go`, `core/database/autoreconciler/autoreconciler.go`, `core/database/autoreconciler/autoreconciler_test.go`.
Do: add `ProductRules`; wire `valueRules`; `CacheVersion` 18 (D-6, D-8). Add T-4's subtest to `TestReconciledNames`.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/autoreconciler/ ./core/database/namealiases/` passes; `go build ./...` has no import cycle.

**STEP-3 — Docs**
Files: `core/autoreconcile/names.go` (header), `docs/conclusion-reconciliation.md` §7 grouping policy and §7.2 Fold bullet, `docs/deployment-plan/spike-11/deployment-plan.md` (map row; the working rule was amended at planning), `docs/ideas/international-names.md` (one line under name changes: canonical alias tokens), `completed.md`, this plan.
Do: replace "an alias never folds" with: "An alias or abbreviation folds a shorter given name or nickname into the longer one it stands for (*Chris* → *Christopher*, *Wm* → *William*; decided 2026-10-19, S11-28). An accent fold never folds, and a surname never alias-folds." In §7 grouping policy: "`Equivalent` never groups; it folds only through an alias of a given name (§7.2)". Then the Definition of done.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-11/prs/S11-28-alias-fold.md` reports no errors; links resolve.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period ("Fold a nickname into the given name it stands for.").

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/autoreconcile/names_test.go` `TestReconcileNamesAliases` *(new)* | One row per AC-1.1, AC-1.2, AC-1.3, AC-1.4, AC-1.5, AC-1.6, AC-1.7, AC-1.8, with `aliasRules`; AC-1.7 uses the `{william, bill}` set on surnames | AC-1.1, AC-1.2, AC-1.3, AC-1.4, AC-1.5, AC-1.6, AC-1.7, AC-1.8, D-1 – D-5 |
| T-2 | Go | same, row "the folded alias is in the value's members" | AC-1.1's result has members `{1, 2}` and reasons `kept` | AC-1.1 |
| T-3 | Go | `core/autoreconcile/modules_test.go` `TestCompatibleFoldsAliases` *(renamed)* | Jim Robins vs James Kenneth Robins → compatible; Jose Robins vs José Robins → not compatible | D-7 |
| T-4 | Go | `core/database/autoreconciler/autoreconciler_test.go` `TestReconciledNames` (new subtest "a nickname folds into the given name it stands for", shaped like "an initial expands into the full name") | `nameIn(a, …, "given=Chris\|surname=Robins")` and `nameIn(b, …, "given=Christopher\|surname=Robins")`, joined → one name row, `Support` 2, parts `given=Christopher\|surname=Robins` | AC-2.1, D-6 |
| T-5 | Go | `core/database/autoreconciler/autoreconciler_test.go` (existing check at the `StoredVersion(db)` call, line 1113 at the base) | Unchanged test; passes with `CacheVersion` 18 | AC-2.2 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `TestReconcileNamesInvariants` (order independence: the fold must not depend on input order), `core/database/promotealign` tests that count exhibits through `Compatible`.

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left — none in this PR
- [ ] UI matches the brief inventory — no UI in this PR
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together) — none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated (if gated) — not gated
- [ ] Domain doc updated: conclusion-reconciliation.md §7 and §7.2
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

1. Open a project with a Source that names a person *Chris Robins* and another Source that names *Christopher Robins*; join both into one Person in Promote.
2. Open Persons. The row reads *Robins, Christopher*, not *Robins, Christopher Chris*.
3. Open the Person. The title reads *Christopher Robins*; Why on the name shows both records counted toward it.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| A longer nickname than its formal name (*Abel* / *Ebbie*) folds the wrong way | Needs both on one person; Out line records the canonical-token fix |
| Order dependence when one short name aliases two longer ones | AC-1.4 uses the pipeline's existing best-supported rule; `TestReconcileNamesInvariants` guards order |

**Stop and report if:**

- `valuecompare.Compare` on `Chris` / `Christopher` does not return `Equivalent` with one pair of cause `alias` at your base.
- `TestReconcileNamesInvariants` fails after STEP-1 and the failure is not one of T-1's new expectations.
- Importing `namealiases` from `autoreconciler` creates an import cycle.
- A test outside `core/autoreconcile`, `core/database/autoreconciler` and `core/database/promotealign` changes result.

## 12. Open questions

None.

## 13. PR

**Title:** S11-28 — Fold a nickname into the given name it stands for
**Body:** objective in one sentence; what shipped against US-1 and US-2; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-11/prs/S11-28-alias-fold.md`.
