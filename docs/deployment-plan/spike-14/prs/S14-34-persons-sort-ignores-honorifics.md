# S14-34 — Persons sort ignores honorifics

**Status:** Draft
**Spike / slice:** Spike 14 · Slice 7d — List and nav polish
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no chrome change: the Persons list rows look the same; only their order changes)
**Board:** —
**Depends on:** —
**Blocks:** —
**Branch:** `s14-34-persons-sort-ignores-honorifics`
**Version bump:** patch — a researcher-visible bug fix in list order ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-name-value`](../../../../.agents/skills/add-name-value/SKILL.md)
**Rules that bind:** `.cursor/rules/go-tests.mdc`, `.cursor/rules/versioning.mdc`
**Context:** working rule "(Batch 7) Persons list sort omits honorific prefixes"; bar outcome 29; dogfood note [`ux/2026-10-15-persons-sort-ignores-honorifics.md`](../ux/2026-10-15-persons-sort-ignores-honorifics.md) (deleted on land); [`structured-name-model.md`](../../../structured-name-model.md) § name display, requirements 2 and 3
**Planned:** 2026-11-06 against `main` @ `91e58c9`
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

The Persons list files a person by their name without its honorific. Today a name typed as "Mrs. Smith" sorts under **M**, because the app's name splitter stores "Mrs." as a `prefix` part and "Smith" as `given`, and the sort key starts with the prefix. Within one surname, rows order by title ("Smith, Dr. …" before "Smith, Mr. …") instead of by given name. After this PR the sort key leaves `prefix` parts out, so "Mrs. Smith" files with Smith and "Smith, Mr. John" files as "Smith, John".

**Why now.** Dogfood note 2026-10-15, pulled into Batch 7 (user, 2026-11-04); bar outcome 29.

**What stays.** How each row's title displays (the title keeps "Mrs." / "Mr."). The name splitter in Swift. The working-label fallback: it is not stripped of titles (user, 2026-11-06: descoped — no code path writes a Person label today).

## 2. User stories

**US-1.** As a researcher, when I scan the alphabetical Persons list, people recorded with an honorific sit beside others of the same name, so that I find them where I look.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | Persons named [prefix "Mrs.", given "Smith"], [given "Zachary", surname "Adams"], [given "Pat", surname "Nolan"]; Western profile installed | Order is Adams, Nolan, Mrs. Smith (keys `Adams, Zachary`, `Nolan, Pat`, `Smith`). Today: Adams, Mrs. Smith, Nolan (`Mrs. Smith` < `Nolan, Pat`) |
| AC-1.2 | Persons [prefix "Dr.", given "Zed", surname "Smith"], [given "Jane", surname "Smith"], [prefix "Mr.", given "John", surname "Smith"]; Western profile installed | Order is Jane Smith, Mr. John Smith, Dr. Zed Smith (keys `Smith, Jane`, `Smith, John`, `Smith, Zed`). Today: Dr. Zed, Jane, Mr. John (`Smith, Dr. Zed` < `Smith, Jane` < `Smith, Mr. John`) |
| AC-1.3 | Catalog with no default name format (stored order); Persons [prefix "Dr.", given "Ann", surname "Bell"], [given "Carl", surname "Ames"] | Order is Ann Bell, Carl Ames (keys `Ann Bell`, `Carl Ames`; no `Dr.` lead) |
| AC-1.4 | A Person whose only name part is [prefix "Rev."], beside [given "Zoe", surname "Quinn"] | The prefix-only row keeps today's key `Rev.`: order is Quinn, Rev. (`Quinn, Zoe` < `Rev.`). It does not drop to the ref-only tail |
| AC-1.5 | [surname_prefix "van", surname "Gogh", given "Vincent"] and [given "Ada", surname "Young"] | Unchanged: key `van Gogh, Vincent` — `surname_prefix` is not `prefix` and stays in the key |

## 3. Scope

| In | Covered by |
| --- | --- |
| The Persons list sort key omits `namevalue.PartTypePrefix` parts | STEP-1, AC-1.1, AC-1.2, AC-1.3 |
| A name with only prefix parts sorts by its full sorted text, as today | STEP-1, AC-1.4 |
| Domain doc says the key omits prefixes | STEP-3 |

| Out | Home |
| --- | --- |
| Stripping leading titles from the working-label fallback (needs a Go title list) | Not planned — descoped by the user 2026-11-06 |
| How a title displays (the row keeps "Mrs." / "Smith, Mr. John") | Not planned (map scope "Not this PR") |
| The splitter reading "Mrs. Smith" as prefix + surname (`NameValueDraft.classify`) | Not planned (map scope "Not this PR") |
| Changing the seeded Western sorted order or `nameformats.Format` | Not planned — D-2 |
| Making `SortByTitle` compute each key once | Not planned — the cost is already per comparison and stays cheap (§6 Scale) |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Persons list order | `core/database/conclusionsummaries/order.go` `personSortKey` | A name with parts sorts by `nameformats.Format(*h.Name, order)`; otherwise by `strings.TrimSpace(h.Entity.Label)`. Every part type is in the key, `prefix` included |
| Caller | `core/database/conclusionsummaries/compose.go` `persons` | Loads `nameformats.LoadFor`, takes `loaded.Profile.Sorted`, calls `SortByTitle(out, personSortKey…, ref)`. Only caller of `personSortKey` |
| Title collation | `core/database/conclusionsummaries/order.go` `SortByTitle` | Root collation, case- and diacritic-insensitive; empty title sorts last by ref; ties by ref. Calls the title func on every comparison |
| Name formatting | `core/database/nameformats/nameformats.go` `Format` | Joins parts by the order; part types the order does not list follow in stored order; empty order = stored order. Mirrors Swift `NameValueDisplay` for display |
| Western sorted order | `core/database/nameformats/registry.go` `seedProfiles` | `surname_prefix`, `surname` + `", "`, `prefix`, `given`, `nick`, `former_surname`, `suffix` |
| Part type | `core/namevalue/namevalue.go` `PartTypePrefix` = `"prefix"`; `PartTypeSurnamePrefix` = `"surname_prefix"` | — |
| Labels | `core/database/promote/promote.go` `fileSubject`, `core/database/identityclaims/bridgefile.go` | Create entities with no `Label`; only tests write a Person label |
| Tests | `core/database/conclusionsummaries/conclusionsummaries_test.go` `TestListPersons`, `TestListPersonsSortsBySortedPartsNotForm`, `TestListOrderIsByShownTitle` | Cover list content, sorted-style order, and collation; no prefix case |
| Test seam | `core/database/conclusionsummaries/export_test.go` `SortPlaces` | Applies the Places order to hand-built rows |

**Precedent.** `TestListOrderIsByShownTitle` + `SortPlaces` in `export_test.go`: hand-built summary rows through a test-only export, table-driven. Copy that shape for persons (D-4). `TestListPersonsSortsBySortedPartsNotForm` shows the end-to-end fixture (`nameformats.Install`, cite, promote, `f.list()`).

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | `personSortKey` formats a copy of the name with every `prefix` part removed. When that copy has no parts left, or formats to empty, the key is `nameformats.Format(*h.Name, order)` of the full name; when that is empty too, the label, as today | The list order lives in `conclusionsummaries` ("Go owns list order"). A prefix-only name keeps a visible title, so it sorts by it instead of falling to the ref-only tail (AC-1.4) | Falling straight to the label — a "Rev." row would jump to the end, sorted by ref |
| D-2 | Do not change `nameformats.Format`, the seeded sorted order, or the profile tables | `Format` mirrors Swift display; dropping `prefix` from the order would not work anyway: `Format` appends unlisted types, giving `Smith, John Mr.`. Seeds reach new catalogs only | Removing `prefix` rows from `seedProfiles` |
| D-3 | Copy the parts slice only when the name holds a `prefix` part; otherwise pass the name through unchanged | The key runs on every comparison of `SortByTitle`; no allocation for the common case | Always filtering into a new slice |
| D-4 | Unit test through a new test-only export `SortPersons(rows []PersonSummary, order []nameformats.PartOrder)` in `export_test.go` that calls the same `SortByTitle(…personSortKey…)` line; one end-to-end fixture case through `ListPersons` | Table-driven cases with exact parts, no Promote needed per case; the fixture case proves `persons` uses it | Only fixture tests — slow and indirect for 5 cases |
| D-5 | No Swift change and no FFI change | The app shows rows in Go's order (`ConclusionListPage`) | — |

**Rules this plan bends.** None. The working rule "(Batch 7) Persons list sort omits honorific prefixes" named label-fallback stripping; the user descoped that half on 2026-11-06 ("You can just descope it"), and the rule and the S14-34 scope entry were rewritten with this plan.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `withoutPrefixes(name namevalue.Value) namevalue.Value` *(new, unexported, `order.go`)* | func | Searched `rg "PartTypePrefix" core`: `core/autoreconcile/names.go` reads prefix order, `core/matchregistry/value.go` sets its role ignored — neither filters parts | Matching ignores prefixes through roles, a different layer; a five-line filter next to its one caller |
| `SortPersons` *(new, `export_test.go`)* | test helper | `SortPlaces` in the same file | Same shape, the Persons key |
| `TestPersonSortOmitsPrefixParts` *(new)* | test | `TestListOrderIsByShownTitle` | New cases; same table shape |

**Builds on.** `SortByTitle` (one owner of title order), `nameformats.Format`, `namevalue.PartTypePrefix`.
**Deletes.** The dogfood note `ux/2026-10-15-persons-sort-ignores-honorifics.md` on land. No code becomes dead.
**Next adds.** None named.

### Reach

None — no new path into existing code; the same rows reach the same list in a different order.

### Scale

At the map's Scale (37,000 Subjects; the Persons list is at most that many rows): `SortByTitle` calls `personSortKey` about 2·n·log₂n ≈ 1.1 M times for 37,000 rows, as today. D-3 adds no allocation for a name without a prefix part, and one small slice copy for a name with one. No query is added (`TestListPersonsQueryCountIsConstant` holds it).

### Edges

- **Assumptions.** `SortByTitle` needs a key that is a pure function of the row (stable sort, consistent comparisons). `withoutPrefixes` reads only the row. Holds by T-1.
- **Neighbors.** `surname_prefix` ("van") stays in the key (AC-1.5, T-1). A `prefix` in the middle of stored parts is removed wherever it sits (T-1 "stored order" case). Two rows differing only by prefix ("Mr. John Smith", "Dr. John Smith") tie on `Smith, John` and break by ref (T-1).
- **Removed inputs.** A name whose parts are all `prefix` keeps its full text (D-1, AC-1.4, T-1).
- **Real data.** None — no table or seed drives the rule; it reads one part type.
- **Untrusted size.** None — no new list or loop over imported data.

### Schema

None — no catalog change.

### Go core

| Package | Change |
| --- | --- |
| `core/database/conclusionsummaries` `order.go` | `personSortKey` uses `withoutPrefixes` per D-1; doc comment says prefixes are not a sort input. `withoutPrefixes` *(new)* per D-3 |
| `core/database/conclusionsummaries` `export_test.go` | `SortPersons` *(new)* |

Sketch:

```go
func personSortKey(h PersonSummary, order []nameformats.PartOrder) string {
	if h.Name != nil && len(h.Name.Parts) > 0 {
		if key := strings.TrimSpace(nameformats.Format(withoutPrefixes(*h.Name), order)); key != "" {
			return key
		}
		if key := strings.TrimSpace(nameformats.Format(*h.Name, order)); key != "" {
			return key
		}
	}
	return strings.TrimSpace(h.Entity.Label)
}
```

`Format` of a value with no parts returns its `form`. `withoutPrefixes` therefore returns the copy with `Form` cleared whenever it removes a part, so a prefix-only name formats to `""` on the first line and takes the second. A form never becomes a sort key ([`structured-name-model.md`](../../../structured-name-model.md) requirement 3).

Errors: none.

### FFI

None — the wire is unchanged.

### macOS

None — D-5.

### Strings

None — no copy.

## 7. Steps

**STEP-1 — Leave prefix parts out of the Persons key**
Files: `core/database/conclusionsummaries/order.go`. Skill: `add-name-value`.
Do: add `withoutPrefixes` (D-3; clears `Form` on the copy, see §6 note) and use it in `personSortKey` (D-1). Update the doc comment: "prefix parts (honorifics) are not a sort input; a name of only prefix parts sorts by them."
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/conclusionsummaries/` passes.

**STEP-2 — Tests**
Files: `core/database/conclusionsummaries/export_test.go`, `core/database/conclusionsummaries/conclusionsummaries_test.go`.
Do: add `SortPersons` (D-4); add T-1 and T-2 (§8).
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/conclusionsummaries/ -run 'TestPersonSortOmitsPrefixParts|TestListPersonsSortIgnoresHonorifics' -v` passes. With STEP-1 reverted, T-1 cases (a), (b), (c), (f) and T-2 fail; (d), (e), (g) are guards and pass either way.

**STEP-3 — Docs and version**
Files: `docs/structured-name-model.md`, `docs/deployment-plan/spike-14/deployment-plan.md`, `docs/deployment-plan/spike-14/completed.md`, this plan, `docs/deployment-plan/spike-14/ux/2026-10-15-persons-sort-ignores-honorifics.md` (delete), `VERSION`, `core/version.go`, `MARKETING_VERSION`.
Do: per the Definition of done. In `structured-name-model.md` requirement 2 add "`prefix` parts (honorifics) are left out of the key; a name of only prefix parts keeps them." In the map: row status; the "Where the notes went" row → "Dogfood 2026-10-15: Persons sort ignores honorifics (note deleted YYYY-MM-DD; landed #N)". Patch bump per `.cursor/rules/versioning.mdc`: the next patch after `VERSION` at implement time (0.3.7 if nothing lands first) in `VERSION`, `core/version.go` and every `MARKETING_VERSION` line in `macos/Provenencia.xcodeproj/project.pbxproj`. `completed.md`: a `## S14-34 — Persons sort ignores honorifics` section at the top, under the intro, in the S14-33 entry's shape. Map row status: `Landed #N`.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-14/prs/S14-34-persons-sort-ignores-honorifics.md` reports no errors; `scripts/check-pr.sh` prints `check-pr: PASS`.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Sort the Persons list without honorifics."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/database/conclusionsummaries/conclusionsummaries_test.go` `TestPersonSortOmitsPrefixParts` *(new)*, table-driven through `SortPersons` with the Western sorted order written out in the test | (a) AC-1.1 rows → Adams, Nolan, Mrs. Smith; (b) AC-1.2 rows → Jane Smith, Mr. John Smith, Dr. Zed Smith; (c) nil order (stored order) AC-1.3 → Ann Bell, Carl Ames; (d) prefix-only "Rev." + Zoe Quinn → Quinn, Rev.; (e) AC-1.5 van Gogh + Ada Young → van Gogh, Young; (f) [prefix "Mr.", John Smith] ref P1 and [prefix "Dr.", John Smith] ref P2 → P1, P2 (a tie on `Smith, John`, broken by ref; today P2 first); (g) prefix-only name with `Form` "Rev. Brown" sorts by `Rev.`, not by form | AC-1.1, AC-1.2, AC-1.3, AC-1.4, AC-1.5, D-1, D-3 |
| T-2 | Go | same file `TestListPersonsSortIgnoresHonorifics` *(new)*. Build each person inline as `TestListPersonsSortsBySortedPartsNotForm` does: `must(t, nameformats.Install(f.c))` first; per person `subjectstest.OfType(t, f.c, userID, f.graph, f.typeID("person"), "")`, `observationstest.Cite(… observations.Input{SubjectID, PropertyID: f.name.ID, Name: &namevalue.Value{Form: "<words joined by spaces>", Parts: …}})` with `Idx` 0..n in typed order, then `promotetest.Save`. A closure local to the test holds those three calls; no new `namevaluetest` helper | AC-1.1 rows → `ListPersons` order Adams, Nolan, Mrs. Smith; each listed `Name` still holds its `prefix` part | AC-1.1, D-1 through `persons` |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `TestListPersons`, `TestListPersonsSortsBySortedPartsNotForm`, `TestListPersonsQueryCountIsConstant`.

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks). It needs macOS; on a machine without `xcodebuild`, report the Go steps' results and "Xcode steps not run" (this PR changes no Swift) instead of claiming PASS
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left (none expected)
- [ ] Sweep: each renamed or deleted name `rg`'d by its old, qualified name across `.agents/`, `.cursor/rules/`, `docs/*.md` and `L10n.swift` with no hit left; `scripts/check-guidance-names.py` passes (check-pr runs it)
- [ ] Every row of **Rules this plan bends** has the user's answer, and its rule is updated
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New (no UI)
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together)
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status, "Where the notes went" row; completed.md: one short paragraph of what shipped
- [ ] Dogfood note ux/2026-10-15-persons-sort-ignores-honorifics.md deleted
- [ ] Domain doc updated: structured-name-model.md requirement 2
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

1. Open a project. In a Citation, type the name "Mrs. Smith" for one person, "Dr. Zed Smith" for a second, "Jane Smith" for a third and "Pat Nolan" for a fourth; Promote all four.
2. Open the Persons list, sorted by name. See Nolan first, then the three Smith rows together: "Mrs. Smith", "Smith, Jane", "Smith, Dr. Zed" in that order (keys `Smith`, `Smith, Jane`, `Smith, Zed`). Before this PR the order was Mrs. Smith, Nolan, Dr. Zed Smith, Jane Smith.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| A prefix-only name falls back to `form` | `withoutPrefixes` clears `Form` on the copy; T-1 (g) |
| Within one surname, "Mrs. Smith" (no surname part) sorts apart from "Smith, Jane" | Accepted: key `Smith` sorts first among the Smiths; the splitter fix that would make it a surname is Out |

**Stop and report if:**

- `personSortKey` is no longer the only Persons list key, or `persons` in `compose.go` no longer calls it.
- `nameformats.Format` of a value with no parts no longer returns `form` (the §6 note would be wrong).
- T-2's promoted name loses its `prefix` part through reconciliation (the kept name has no prefix) — then AC-1.1 is not reachable end to end and the reconciler's handling needs a decision.
- Any Swift test sorts Persons locally and fails.

## 12. Open questions

None.

## 13. PR

**Title:** S14-34 — Persons sort ignores honorifics
**Body:** objective in one sentence; what shipped against US-1; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-14/prs/S14-34-persons-sort-ignores-honorifics.md`.
