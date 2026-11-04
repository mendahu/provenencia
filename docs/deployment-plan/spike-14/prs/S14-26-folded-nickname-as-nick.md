# S14-26 — A folded nickname shows as a nickname

**Status:** Planned
**Spike / slice:** Spike 14 · Slice 3 — Name display
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no chrome change: the reconciler stores a `nick` part; S14-09 already quotes `nick` on every surface that shows a name)
**Board:** —
**Depends on:** S14-09 (landed [#649](https://github.com/mendahu/provenencia/pull/649)), S14-18 (landed [#677](https://github.com/mendahu/provenencia/pull/677))
**Blocks:** —
**Branch:** `s14-26-folded-nickname-as-nick`
**Version bump:** minor — researchers see nicknames the fold used to drop ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-match-rule`](../../../../.agents/skills/add-match-rule/SKILL.md) (reconciler section and `CacheVersion`), [`add-name-value`](../../../../.agents/skills/add-name-value/SKILL.md) (part types; `nick` already exists)
**Rules that bind:** `.cursor/rules/match-rules.mdc`, `.cursor/rules/go-tests.mdc`, `.cursor/rules/versioning.mdc`
**Context:** [`docs/conclusion-reconciliation.md`](../../../conclusion-reconciliation.md) §5 step 3 and §7.2 Fold / One name; working rules: "A nickname folds into its set's formal name, as today", "(Batch 3) S14-09 is Spike 11's S11-29, refreshed", "(Batch 3) A nickname-class fold reappears as a nick part"; dogfood decision Jake 2026-10-28 (nickname-class re-tag; note deleted 2026-11-04, already scoped); S14-18 **Next adds**; bar outcome 22
**Planned:** 2026-11-04 against `main` @ `62b94f03`
**Review:** [`S14-26-folded-nickname-as-nick.review.md`](S14-26-folded-nickname-as-nick.review.md) — Fits with changes — Ready with changes, 2 rounds

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

When one Person's records hold a nickname written as a given name beside the formal name it stands for (*Bill* and *William*, *Jim* and *James*), the reconciler still keeps the formal given name, but the assembled name also carries the folded word as a `nick` part. Every surface S14-09 decorates then reads *William “Bill” Robins* / *James “Jim” Robins* instead of dropping the short form. Initials (*J.*) and abbreviations (*Wm*) still fold away with no nick.

**Why now.** Bar outcome 22 and the 2026-10-27 dogfood ask (decided 2026-10-28: nickname-class only). S14-09 shipped quotes for recorded `nick` parts; S14-18 made every fold's root one hop with a `total`, which this PR reads.

**What stays.** Which pairs fold, fold directions, which value wins, majority, weak drop, and Swift display decoration are unchanged. A researcher-typed `nick` part still shows as today. Promote's proposal UI is unchanged.

## 2. User stories

**US-1.** As a researcher, when a Person's records hold a nickname-class short form recorded as a given name beside its formal name, I see the short form in nickname quotes on the Persons list and the Person page, so the name they went by is not lost from the headline.

Every example: one candidate per name in the order listed, ObservationIDs `{0,0,0,1}`, `{0,0,0,2}`, …, no SourceID, default Provenance, `surname=Robins` on each, product rules (`autoreconciler.ValueRules`), cardinality single. One displayed value, reason `kept`, holding every candidate; folded candidates have outcome `folded`.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | Given `Bill`, `William` | Parts `given=William\|nick=Bill\|surname=Robins` |
| AC-1.2 | Given `Jim`, `James` | Parts `given=James\|nick=Jim\|surname=Robins` |
| AC-1.3 | Given `J.`, `James` | Parts `given=James\|surname=Robins` (no nick) |
| AC-1.4 | Given `Wm`, `William` | Parts `given=William\|surname=Robins` (no nick) |
| AC-1.5 | Given `B.`, `Bill`, `William` | Parts `given=William\|nick=Bill\|surname=Robins` (initial dropped; nickname kept) |
| AC-1.6 | Given `Ann`, `Anna`, `Hannah` | Parts `given=Hannah\|nick=Ann\|nick=Anna\|surname=Robins` (chain: both nickname hops reappear) |
| AC-1.7 | Given `Bill` once and `William` three times (four candidates) | Parts `given=William\|nick=Bill\|surname=Robins` (once is enough) |
| AC-1.8 | Given `Bill`, `Billy`, `William` | Parts `given=William\|nick=Bill\|nick=Billy\|surname=Robins` (distinct nicknames both kept; order is best-supported folded value first, then best ranked) |
| AC-1.9 | Candidates `nick=Bill\|surname=Robins` and `given=William\|surname=Robins` | Parts `given=William\|nick=Bill\|surname=Robins` (researcher-typed nick unchanged; no second Bill) |
| AC-1.10 | Candidates `given=Bill\|surname=Robins` and `given=William\|nick=Bill\|surname=Robins` | Parts `given=William\|nick=Bill\|surname=Robins` (folded Bill dedupes with the typed nick) |

**US-2.** As a contributor, every catalog rebuilds once for the new assembled names, and the reconciliation doc says what a folded nickname becomes.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | A catalog stamped `CacheVersion` 22 | `NeedsRebuild` is true; `autoreconciler.CacheVersion` is 23; `productAliasPin` holds version 23 with the hash unchanged |
| AC-2.2 | `docs/conclusion-reconciliation.md` §7.2 | States that a nickname-class fold reappears as a `nick` part; initials and abbreviations do not |
| AC-2.3 | Pairs that already fold under product rules without a nickname hop (`J.`+`James`, `Wm`+`William`) | Same assembled parts as at the planning base (no nick) |

## 3. Scope

| In | Covered by |
| --- | --- |
| When a given (or other AliasRoles) unit folds into a kept root through nickname-class hops, `assemble` keeps the folded recorded parts D-7 names as `nick` parts | STEP-1, STEP-2, AC-1.1 – AC-1.10, D-1 – D-7 |
| Initials, abbreviations, and pure prefix folds that are not nickname-class leave no nick | STEP-2, AC-1.3, AC-1.4, AC-2.3, D-1, D-7 |
| Pipeline passes each folded attach into `assemble`, with `nickEdge` / `chainNick` from `foldValues` (S14-18 flattened `into`; this PR classifies hops beside that matrix) | STEP-1, STEP-2, D-2, D-7, D-11 |
| `CacheVersion` 23 and its pin | STEP-3, AC-2.1, D-8 |
| `conclusion-reconciliation.md` §7.2 updated; map Check examples use product-seed pairs | STEP-4, AC-2.2, D-9 |
| Patch-free minor version bump | STEP-5 |

| Out | Home |
| --- | --- |
| Folds of surnames | not planned (product alias tables never cover family) |
| Changing which pairs fold, `nicknameDirections`, or the seed | not planned; user adds Directions as dogfood finds them |
| A new outcome reason (`ReasonRetyped`) or Why chrome for the re-tag | not planned (D-6); outcomes stay `folded` / `kept` |
| Swift display changes | S14-09 already quotes `nick` |
| Promote proposal UI | not planned |
| Search document alias lines (`NameFoldsInto`) | S14-21; still required for spellings no record used as a nick |
| Reconciliation Claims / concluded-name override | not built; `applyConcluded` keeps the concluded value as written (D-10) |
| Collapsing a fold cycle into one member | not planned (S14-18 D-2); cycle members do not fold, so none re-tag |

## 4. Current state

Verified at planning base `main` @ `62b94f03`.

| Area | Where | Today |
| --- | --- | --- |
| Fold rule | `core/autoreconcile/names.go` `nameModule.fold` | A unit folds when Compare says Contains/Equivalent and every paired word is Same, Initial, Prefix (role-allowed), or Alias/Abbreviation that `Expands`. Does not record which cause produced the fold |
| Assemble | `core/autoreconcile/names.go` `nameModule.assemble` | Builds parts only from **settled** (kept) units' raw parts. A folded given is gone from the name. Order comes from members' recorded part-type orders; `nick` appears only when a member already had a nick part |
| Module API | `core/autoreconcile/modules.go` `module.assemble` | `assemble(members []Candidate, settled map[string][]unit) Value`. `dateModule` and `keyModule` implement it; only names need folded pairs |
| Pipeline fold | `core/autoreconcile/pipeline.go` `foldValues` | Sets `into` one hop to the sink (S14-18). Path and fold kind are not kept |
| Pipeline roots | `pipeline.go` reconcile step 6 | When `v.into != nil`, rewrites `e.roots[name] = root(v)` and sets outcome `folded`. The folded unit is discarded from the entry |
| Settled input | `pipeline.go` `row.settledUnits` | Distinct root units only. Folded units never reach `assemble` |
| Alias kinds | `core/valuecompare/aliases.go` `AliasKind`, `Hit`, `Expands` | `AliasNickname` vs `AliasAbbreviation`. `Expands` is true for both; `Hit` returns the kind. Initials are `CauseInitial`, not an alias hit |
| Product seed | `core/database/namealiases/nicknames_seed.go`, `registry.go` | `james` set holds `jim` / `jamie` / …; `jake` folds into `jacob` / `jakob`, **not** `james`. `william` holds `bill`. Abbreviations table holds `wm`→`william` |
| Display | Swift `NamePartType.nick.decoration` (S14-09) | Quotes every `nick` part. No app change needed once Go stores the part |
| Cache | `core/database/autoreconciler/autoreconciler.go` `CacheVersion` | 22; pin `{22, "0dee1d25…"}` in `autoreconciler_test.go` |
| Tests | `names_test.go` `TestReconcileNamesAliases` AC-1.1 | `Chris`+`Christopher` → `given=Christopher` only (the row this PR changes). `TestReconcileNamesCombine` "nickname recorded as a given name" uses `testRules` (no alias table), so Jake+James stay two givens |
| Docs | `conclusion-reconciliation.md` §7.2 | Fold drops the short form from the assembled name; "One name" example still shows typed `nick Jake` + `given James` only |

**Precedent.** S14-18 (`#677`, plan `S14-18-name-folds-never-crash.md`): reconciler fold change + pure/product tests + `CacheVersion` bump + doc line. Copy that commit shape. S14-18 **Next adds** says this PR asks the module about the `(value, root)` pair and adds the path if needed — this plan keeps a nickname-class edge matrix beside `foldValues`' fold matrix (D-7) so chain hops use the same Compare classifier as `nameModule.fold`. Do not copy S14-18's sink-resolution rewrite.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | Only **nickname-class** hops reappear as `nick`. `module.fold` returns `(folds, nickClass bool)`: `folds` is today's fold result; `nickClass` is true only when `folds` and every non-`Same` Compare part is `CauseAlias` with `Expands` (`nameModule`); `dateModule` and `keyModule` always return `nickClass false`. Initials, abbreviations, and prefix hops are not nickname-class. Emission rules are in D-7 | Jake 2026-10-28; dogfood named Cause as the discriminator; one Compare in `fold` (M-5); probe: bill→william Alias, wm→william Abbreviation, j→james Initial, al→alexandra Prefix | A second `nicknameHop` that Compares again; alias-index token BFS; treating abbreviations as nicks |
| D-2 | The change lives in Go: pipeline records each folded `(from, root)` unit pair on the entry before root rewrite; `assemble` receives those pairs and emits `nick` parts into the stored `NameValue`. Swift only displays what Go stored | Map recommendation; Why outcomes already say `folded`; one place for every surface | Inferring nicknames in Swift from fold outcomes; rewriting Observations |
| D-3 | A nickname recorded once beside many records of the formal name still reappears (AC-1.7) | Bar outcome 22; dogfood: the short form is a real fact even as a minority spelling. Fold already ran; support count does not gate the nick | Require MinMajoritySupport on the folded value before re-tagging |
| D-4 | Distinct nickname-class folded parts all reappear (`Bill` and `Billy`); `assemble`'s existing `containsAll` per type dedupes identical words, including against a researcher-typed `nick` | Dogfood open question on cardinality; matches "One name" keeping every surviving value per type | Keep only the best-supported nick; drop typed nick when a fold re-tags the same word |
| D-5 | Emitted parts keep the folded record's **spelling** and `Idx` order within the nick slot; `Type` becomes `namevalue.PartTypeNick`. No new Reason | Display and form read recorded text; S14-09 quotes by type | Invent `ReasonRetyped`; normalize spelling to the alias token |
| D-6 | No Why / outcome chrome change. Folded candidates stay `ReasonFolded`; the displayed value stays `ReasonKept` with the extra nick parts | Claims and Why already name fold; Out | A new reason or Why line for "re-tagged as nick" |
| D-7 | In the existing `foldValues` n² loop, set `edge[a][b]` from `folds` and `nickEdge[a][b]` from `nickClass` of one `m.fold` call (D-1). Each `foldAttach` carries `from`, `root`, `sources`, `rank`, and `chainNick` (nickEdge-BFS from `from` reaches `root`). `nicknameParts`: (1) `Compare(from, root)` — emit each from raw part that has a `CauseAlias`+`Expands` match and is not already in the root (same predicate as `nickClass`); (2) else if `chainNick` and step 1 emitted nothing, emit each from raw part not `containsAll` in the root. Ann+Anna+Hannah → both nicks; B.+Bill+William → Bill only; Al+Alexandra → no nick; Chris+R + Christopher+Robert → nick Chris only | S14-18 Next adds; M-1; M-5 (one fold result fills both matrices) | Separate `nicknameHop`; alias-index BFS; direct Compare alone |
| D-8 | `CacheVersion` 23, history line `23: a nickname-class fold reappears as a nick part of the assembled name (S14-26).` Pin version 23; hash unchanged | Assembled cached names change | No bump |
| D-9 | Manual check and map Check examples use product-seed pairs: **Bill+William** and/or **Jim+James**. The map's narrative *James*+*Jake* is wrong for the seed (`jake` folds into `jacob`, not `james`); the scope entry and bar wording keep the idea, the Check names Jim | Probe at planning base | Add a Jake→James Direction in this PR (Out: seed edits) |
| D-10 | A concluded name (`applyConcluded`) is unchanged: claims are not built, and when they are the concluded value is authoritative as written | `conclusion-reconciliation.md` §3; Out | Attach folded nicks onto a concluded value |
| D-11 | `module.assemble` gains a third argument `folded map[string][]foldAttach`. `foldAttach` holds `from unit`, `root unit`, `sources int`, `rank int`, `chainNick bool` (D-7). `row.foldedUnits()` sorts each type's attaches by `sources` descending, then `rank` ascending (the sink tie-break in `foldValues`, not `total` — a folded `unitValue` never has `total`). `dateModule` and `keyModule` ignore the map. `nameModule` is the only consumer | Folded units never reach assemble today; `unit` has no support fields (M-2 / X-1); names are the only module that re-tags | Stash on `nameModule` fields; `foldAttach` as bare `{from,root unit}` without support; sort by `total` |

**Rules this plan bends.** None. Jake's 2026-10-28 nickname-class decision is already the map working rule "(Batch 3) A nickname-class fold reappears as a nick part" (planner added it with this plan); STEP-4 only adds the §7.2 domain doc line if that rule is already present.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `foldAttach` *(new)* | type | `unitValue.into` | `into` is rewritten away before assemble; need endpoints, support, `chainNick` |
| `entry.folded` *(new)* | field | `entry.roots` | Roots are post-fold; folded keeps the pre-rewrite unit and attach fields per name |
| `row.foldedUnits` *(new)* | method | `row.settledUnits` | Same shape for the third assemble arg; sorts by `sources` then `rank` (D-11) |
| `nicknameParts` *(new)* unexported on `nameModule` | func | Assemble's settled-part loop | Emits nick parts from a `foldAttach` (D-7); hop class comes from `fold`, not a second classifier |
| `fold` returns `(folds, nickClass bool)` | signature | Today's `fold(...) bool` | One Compare; `foldValues` sets `edge` and `nickEdge` from the same call (M-5) |
| `nickEdge` matrix inside `foldValues` | local | Existing `edge` matrix in `foldValues` | Filled beside `edge` from `nickClass`; no extra Compare |
| Fold-nickname expectation updates | test rows | `TestReconcileNamesAliases`, `TestValueRulesFoldChains`, `TestAutoReconciledNames` | Same tables; expectations gain `nick=` |

No new package, FFI method, table, L10n key, or Swift type.

**Builds on.** S14-18 fold roots and edge matrix; S14-09 `nick` decoration; `nameModule.fold`'s CauseAlias branch; `assemble`'s `containsAll` dedupe; `add-match-rule` CacheVersion rule.
**Deletes.** Nothing becomes dead. Alias-fold test expectations that assert "formal given only" are updated to include the nick (they are the behavior change, not dead code).
**Next adds.** A new nickname Direction is one seed/Directions edit + CacheVersion bump; nickEdge follows through `fold`. A concluded-name claim that should keep auto nicks is a later claims PR (D-10).

### Reach

| New path or state | Destination | Checked: resets on the new id · handles a missing referent · shows the state |
| --- | --- | --- |
| Assembled `NameValue` with auto `nick` parts | Conclusion headers, Persons list, Person page title, Why title (S14-09 quotes); search / profile that read the reconciled name | Rebuild on CacheVersion 23; empty nick when no nickname fold; typed+folded Bill dedupes (AC-1.9, AC-1.10) |
| Fold cycle (Dick, Ricky, Rick) | `assemble` with empty folded map for those units | No fold → no re-tag; all three givens stay (S14-18 AC-1.7 unchanged) |

### Scale

At the map's Scale (heaviest person in `BenchmarkReconcileNames`: 200 name candidates): `foldValues` still runs n² `m.fold` calls (each Compare once); `nickEdge` is a second n² bool from the same return; `chainNick` BFS is over at most n value indexes. No alias-index walk. Bound: `BenchmarkReconcileNamesAliases` *(new)* — same 200-candidate shape as `BenchmarkReconcileNames`, but `Reconcile(aliasRules, …)` with given tokens that alias-fold (Bill/William, Jim/James, Chris/Christopher rotating); median at most 25% slower than the same bench with nick emission stubbed no-op. `BenchmarkReconcileNames` under `testRules` stays green but does **not** gate this PR (M-4). CacheVersion bump rebuilds every catalog once on open.

### Edges

- **Assumptions.** Folded pairs are those with `into != nil` after `foldValues` (cycle members excluded). `nickEdge` ⊆ fold `edge`. `chainNick` BFS visits each value index at most once. Distinct nicks sort by `sources` then `rank` (D-11). Direct Alias emits require `CauseAlias` and `Expands`, matching `fold`.
- **Neighbors.** Bill+William → William “Bill”. Wm+William → William. J.+James → James. B.+Bill+William → William “Bill”. Ann+Anna+Hannah → Hannah “Ann” “Anna”. Chris+Christopher → Christopher “Chris”. Al+Alexandra (no Alex) → Alexandra only (prefix). Ann+Annette (no Anna) → Annette only (prefix). Chris R + Christopher Robert → Christopher Robert “Chris”. Dick+Ricky+Rick → three givens, no nick. Add Richard → Richard with nicks Dick, Ricky, Rick (each nickEdge-reaches Richard). Chris+Kit → unchanged apart. Researcher nick=Bill + given=William → William “Bill”. Folded given=Bill + typed nick=Bill → one Bill. Fred+Bill+William → William “Fred” “Bill” (Fred reaches William only on nickEdge through Bill). Bill beside Robert and William → nicks attach to the sink `foldValues` picks; the other formal stays a given.
- **Removed inputs.** Drop William from Bill+William → given Bill only. Drop Bill from B.+Bill+William → given B.|William (no nick). Drop Anna from Ann+Anna+Hannah → Ann may still nickEdge-reach Hannah only if a nickEdge path remains; with only `{hannah,anna}` and `{anna,ann}` edges, Ann→Hannah needs Anna — without Anna, Ann and Hannah stay apart or prefix-fold if Compare says so (product: Ann+Hannah without Anna do not fold; both givens).
- **Real data.** Product: 2,627 expand edges; 28 prefix pairs that nickname-index-reach without direct Expands (M-1 probe) must **not** grow a nick when the recorded hop is prefix-only (Al+Alexandra). `TestValueRulesFoldChains` already runs B./Bill/William, Ann/Anna/Hannah, Mo/Molly/Mary, Caroline/Chick/Charles, Peggy/Margaret, D./Dick/Richard under product rules — update its expected parts (M-3). `jake`→`james` Expands false (D-9).
- **Untrusted size.** n is distinct values of one part type on one person (same bound as today's fold matrix). `chainNick` BFS is O(n²) bool steps on that n. No alias-index walk.

### Schema

None — no catalog change.

### Go core

| Package | Change |
| --- | --- |
| `core/autoreconcile` | `module.fold` → `(folds, nickClass bool)` (D-1); update `nameModule.fold`, `dateModule.fold`, `keyModule.fold`, and date fold tests. `foldValues` sets `edge`/`nickEdge` from that return and `unitValue.chainNick` via nickEdge-BFS to the chosen `into` (D-7). `foldAttach` / `entry.folded` / `row.foldedUnits` / `assemble` third arg per D-11. `nameModule.assemble` emits via `nicknameParts` (D-7); insert `nick` into `order` after `given` (after `prefix` if no given); dedupe with `containsAll`. Doc comments on Fold/Assemble |
| `core/database/autoreconciler` | `CacheVersion = 23` with history line (D-8); pin update |

Sketch (not the implementation):

```go
type foldAttach struct {
	from, root unit
	sources    int
	rank       int
	chainNick  bool
}

func (m nameModule) fold(typ string, from, into unit) (folds, nickClass bool) {
	// today's fold body; nickClass when every non-Same part is CauseAlias with Expands.
}

func (m nameModule) nicknameParts(typ string, a foldAttach) []namevalue.Part {
	// D-7: (1) Alias+Expands parts from Compare(from,root); (2) else if chainNick, parts not in root.
}
```

Errors: none.

### FFI

None — no wire change. Cached name blobs change shape after rebuild; clients already accept `nick` parts (S14-09).

### macOS

None — no app change. S14-09 quotes `nick` on the Persons list and Person page title.

### Strings

None — no copy changes. Sweep still runs (DoD).

## 7. Steps

**STEP-1 — Pass folded pairs into assemble; fold returns nickClass**
Files: `core/autoreconcile/modules.go`, `core/autoreconcile/pipeline.go`, `core/autoreconcile/dates.go`, `core/autoreconcile/dates_test.go`, `core/autoreconcile/names.go`. Skill: `add-match-rule`.
Do: change `module.fold` to `(folds, nickClass bool)` (D-1); `nameModule.fold` returns `nickClass` per D-1 (still always computes folds as today); dates/keys return `false` for nickClass; update `foldValues` to read both returns (ignore nickClass this step, or store it unused). Add `foldAttach` (D-11; `chainNick` false); extend `assemble` with `folded map[string][]foldAttach`; dates/key ignore it; add `unitValue.chainNick bool` (false); `entry.folded map[string]foldAttach`; step 6 stores the attach before root rewrite; `row.foldedUnits()` sorted per D-11; call `assemble(..., r.foldedUnits())`. `nameModule.assemble` ignores folded this step.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/autoreconcile/ ./core/database/autoreconciler/ ./core/subjectprofile/ ./core/database/promotealign/` passes with no expectation changes.

**STEP-2 — Emit nickname-class folded parts as nick**
Files: `core/autoreconcile/names.go`, `core/autoreconcile/pipeline.go`, `core/autoreconcile/names_test.go`, `core/database/autoreconciler/autoreconciler_test.go`. Skill: `add-match-rule`, `add-name-value` (PartTypeNick only). Rule: `go-tests.mdc`.
Do: wire `nickEdge`/`chainNick` in `foldValues` from `fold`'s `nickClass` (D-7); implement `nicknameParts` / assemble emission. Update names.go Fold/Assemble comments. Add `{william, billy}` to `aliasRules`. Update **every** `TestReconcileNamesAliases` fold-table row that nickname-class-folds so `want` gains `nick=` (minimum: Chris→Christopher; Chris+R→Christopher+Robert; Ebbie→Abel; B.+Bill+William; Ann+Anna+Hannah; and every other nickname-class row). Leave abbreviation (Wm), initial-only, cycle-without-sink, and `sameResult` rows formal-only / unchanged. Change `TestValueRulesFoldChains` to assert full `nameSpec` (not `givenOf`) including nicks: Bill on William; Ann+Anna on Hannah; Peggy on Margaret; Molly on Mary (Mo has no nickEdge path → no Mo nick); Chick+Caroline on Charles; Dick on Richard (not D.). Update `TestReconciledNames` subtests `"a nickname folds into the given name it stands for"` and `"a chain of folds files and rebuilds"` `nameSpec` expectations. Add T-3 with `nameSpec`. Add Al+Alexandra and Ann+Annette formal-only rows under product rules. Add `BenchmarkReconcileNamesAliases` (Scale corpus).
Check: `CGO_ENABLED=1 go test -tags fts5 -run 'TestReconcileNames|TestValueRulesFold|TestReconciledNames|TestPrefix' ./core/autoreconcile/ ./core/database/autoreconciler/` passes. With emission reverted, Bill+William nick assertion fails. `CGO_ENABLED=1 go test -tags fts5 -run '^$' -bench BenchmarkReconcileNamesAliases -count 5 ./core/autoreconcile/`: median at most 25% slower than the same bench with nick emission stubbed no-op.

**STEP-3 — CacheVersion 23**
Files: `core/database/autoreconciler/autoreconciler.go`, `core/database/autoreconciler/autoreconciler_test.go`. Skill: `add-match-rule`.
Do: bump per D-8; update `productAliasPin.CacheVersion` to 23.
Check: `CGO_ENABLED=1 go test -tags fts5 -run 'TestProductAliasTablesPinCacheVersion|TestNeedsRebuild|TestValueRulesFold' ./core/database/autoreconciler/` passes.

**STEP-4 — Docs and map**
Files: `docs/conclusion-reconciliation.md` §7.2; `docs/deployment-plan/spike-14/deployment-plan.md` only if the Batch 3 nick working rule or S14-26 Check still disagree with D-9.
Do: after the Fold bullet's alias sentence, state that a nickname-class fold reappears as a `nick` part of the assembled name, and that initials, abbreviations, and prefix-only hops do not. In "One name", extend the example: *given Bill* + *given William* → `given: [William], nick: [Bill]`. Confirm the map already has working rule "(Batch 3) A nickname-class fold reappears as a nick part" and Bill/Jim Check examples; edit only if drifted.
Check: `rg -n "reappears as a \`nick\`|nickname-class fold" docs/conclusion-reconciliation.md` hits the new §7.2 lines.

**STEP-5 — Version bump**
Files: `VERSION`, `core/version.go`, Xcode `MARKETING_VERSION` (both configurations). Rule: `versioning.mdc`.
Do: bump minor per header (0.2.13 → 0.3.0 at planning base; if `VERSION` moved, take the next minor from whatever is on the branch).
Check: the three sources match; `rg` for the old marketing version in those files has no stale hit.

**STEP-6 — Docs (landed paperwork)**
Files: this plan (status), `deployment-plan.md` map row, `completed.md`.
Do: per Definition of done when the PR lands (implementer). While planning, the planner sets status Planned and map Plan link.
Check: links resolve; `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-14/prs/S14-26-folded-nickname-as-nick.md` reports no errors.

Commits: one per behavior. (1) "Pass each folded name unit into assemble so a nickname can reappear." (STEP-1). (2) "Show a folded nickname as a nick part of the reconciled name." (STEP-2–STEP-3). (3) "Say that a nickname-class fold reappears as a nick part." (STEP-4). (4) "Bump the product version for folded nicknames as nick parts." (STEP-5). (5) Docs landed lines (STEP-6).

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/autoreconcile/names_test.go` `TestReconcileNamesAliases` rows *(update/add)* | Under `aliasRules` (add `{william,billy}`): Bill+William → `given=William\|nick=Bill\|surname=Robins`; Jim+James → `given=James\|nick=Jim\|surname=Robins`; B.+Bill+William → `given=William\|nick=Bill\|surname=Robins`; Ann+Anna+Hannah → `given=Hannah\|nick=Ann\|nick=Anna\|surname=Robins`; Bill+Billy+William → `given=William\|nick=Bill\|nick=Billy\|surname=Robins`; Chris+Christopher → `given=Christopher\|nick=Chris\|surname=Robins`; every other nickname-class fold row in that table gains its `nick=` the same way | AC-1.1, AC-1.2, AC-1.5, AC-1.6, AC-1.8, D-1, D-7 |
| T-2 | Go | `names_test.go` `TestReconcileNamesAliases` *(new rows)* | J.+James → `given=James\|surname=Robins`; Wm+William → `given=William\|surname=Robins`; `nick=Bill`+`given=William` → one Bill nick; `given=Bill`+`given=William\|nick=Bill` → one Bill nick | AC-1.3, AC-1.4, AC-1.9, AC-1.10, D-4 |
| T-3 | Go | `core/database/autoreconciler/autoreconciler_test.go` `TestValueRulesFoldedNicknameParts` *(new)* | Assert with `nameSpec` (full `type=value\|…`, not `givenOf`): Bill+William; Jim+James; J.+James; Wm+William; one Bill + three Williams; Al+Alexandra → no nick; Ann+Annette → no nick | AC-1.1, AC-1.2, AC-1.3, AC-1.4, AC-1.7, D-3, D-7, D-9 |
| T-4 | Go | `autoreconciler_test.go` `TestProductAliasTablesPinCacheVersion` | Pin version 23, same hash | AC-2.1, D-8 |
| T-5 | Go | `TestReconcileNamesAliases` unchanged rows + `TestPrefixNeverFoldsASurnameOrAcrossTypes` + updated `TestValueRulesFoldChains` (`nameSpec`) + updated `TestReconciledNames` nick rows | Still green; chain/DB goldens carry nicks | AC-2.3, D-7 |
| T-6 | Docs | STEP-4 check | `rg` hits the §7.2 "reappears as a nick" sentence | AC-2.2 |
| T-7 | Go | `names_test.go` `BenchmarkReconcileNamesAliases` *(new)* | Alias-folding corpus; gates Scale 25% bar | D-7, Scale |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `TestReconcileNamesAliases` formal-only expectations, `TestValueRulesFoldNicknamesOnly`, any golden that shows Chris/Christopher after reconcile.

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
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together) — or none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated (if gated)
- [ ] Domain doc updated if a rule in docs/*.md changed: conclusion-reconciliation.md §7.2
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

Makes slice 3 Check / bar outcome 22 true.

1. Create a new project. Add a Source with person given *William*, surname *Robins*; add another Source (or second Observation) with given *Bill*, surname *Robins*; promote and join into one Person (or promote so both land on one handle). Quit and reopen so CacheVersion 23 rebuilds.
2. Persons list and Person page title read *William “Bill” Robins*.
3. Repeat with *James* + *Jim* → *James “Jim” Robins*.
4. A Person with only *J.* and *James* still reads *James Robins* (no quotes).
5. A Person with only *Wm* and *William* still reads *William Robins*.
6. Open `docs/conclusion-reconciliation.md` §7.2 and confirm it states that a nickname-class fold reappears as a `nick` part and that initials and abbreviations do not (AC-2.2).

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| Prefix folds that index-reach a formal name get a nick | D-7 nickEdge (not alias-index walk); T-3 Al+Alexandra / Ann+Annette |
| Benchmark under `testRules` looks fine while alias path regresses | T-7 `BenchmarkReconcileNamesAliases` gates Scale |
| Map still says Jake+James | Planner already corrected Check examples (D-9); STEP-4 confirms |

**Stop and report if:**

- `foldValues` / step-6 root rewrite no longer matches §4 (cannot record `entry.folded` / `chainNick`).
- `nameModule.fold`'s Alias branch no longer uses `CauseAlias` + `Expands` (cannot set `nickClass`).
- Emitting nicks requires changing Observations, Swift decoration, or a new Reason (Out).
- `BenchmarkReconcileNamesAliases` median is more than 25% slower than the no-emission stub on the same machine after STEP-2.
- A test outside §7 fails for a reason other than an assembled name that now carries a `nick` part (update that golden and list it in the PR).

## 12. Open questions

None.

## 13. PR

**Title:** S14-26 — A folded nickname shows as a nickname
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-14/prs/S14-26-folded-nickname-as-nick.md`.
