# S14-34 — Persons sort ignores honorifics

**Status:** Landed in #727 (2026-11-06)
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
**Context:** working rule "(Batch 7) Persons list sort omits honorific prefixes"; bar outcome 29; dogfood note `ux/2026-10-15-persons-sort-ignores-honorifics.md` (deleted on land); [`structured-name-model.md`](../../../structured-name-model.md) § name display, requirements 2 and 3
**Landed:** [#727](https://github.com/mendahu/provenencia/pull/727), 2026-11-06. Differences from the plan: (1) The Status line reads "Landed in #727 (2026-11-06)" and the dogfood note's link in Context is plain text, because `check-plan-docs` rejects the shorter form and a dead link. (2) `SortKey` is the method `nameformats.Profile.SortKey`, superseding the D-1 and D-5 signatures, and the free `Format` is private (`format`), with comment-only Swift edits against D-6; both by the user, 2026-11-06. Decisions and Scope text stand as written.
**Planned:** 2026-11-06 against `main` @ `91e58c9`
**Review:** [`S14-34-persons-sort-ignores-honorifics.review.md`](S14-34-persons-sort-ignores-honorifics.review.md) — Fits with changes — Ready with changes, 2 rounds; no open blockers; To user none

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

**What stays.** How each row's title displays, and `nameformats.Format`'s output (the title keeps "Mrs." / "Mr."). The name splitter in Swift. The working-label fallback: it is not stripped of titles (user, 2026-11-06: descoped — no code path writes a Person label today).

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
| `nameformats.SortKey` *(new)* owns a name's list sort key: the order's join minus `prefix` parts, keeping them when nothing else remains, never `form` | STEP-1, AC-1.1, AC-1.3, AC-1.4, AC-1.5 |
| The Persons list key uses `SortKey`, then the label | STEP-3, AC-1.1, AC-1.2, AC-1.3 |
| `SortByTitle` computes each row's title once (map review M-1, measured at Scale) | STEP-2 |
| One function owns each list's sort line (`sortPersons`, `sortPlaces`), and the test seam exposes them | STEP-2 |
| Docs that say `Format` or the sorted order is the key | STEP-4 |

| Out | Home |
| --- | --- |
| Stripping leading titles from the working-label fallback (needs a Go title list) | Not planned — descoped by the user 2026-11-06 |
| How a title displays (the row keeps "Mrs." / "Smith, Mr. John") | Not planned (map scope "Not this PR") |
| The splitter reading "Mrs. Smith" as prefix + surname (`NameValueDraft.classify`) | Not planned (map scope "Not this PR") |
| Changing the seeded Western sorted order, the profile tables, or `nameformats.Format`'s output | Not planned — D-2 |
| A per-profile list of part types left out of the sort key (stored in the profile) | [`docs/ideas/international-names.md`](../../../ideas/international-names.md) "Particle sort" — it becomes a profile field there |
| Precomputing collation keys (`titleCollator.KeyFromString`) in `SortByTitle` | Not planned — keys once already brings 37,000 rows to 0.37 s (§6 Scale) |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Persons list order | `core/database/conclusionsummaries/order.go` `personSortKey` | A name with parts sorts by `nameformats.Format(*h.Name, order)`; otherwise by `strings.TrimSpace(h.Entity.Label)`. Every part type is in the key, `prefix` included |
| Persons caller | `core/database/conclusionsummaries/compose.go` `persons` | Loads `nameformats.LoadFor`, takes `loaded.Profile.Sorted`, calls `SortByTitle(out, func(h) personSortKey(h, order), ref)`. Only caller of `personSortKey` |
| Places caller | `core/database/conclusionsummaries/compose.go` `places` | `SortByTitle(out, placeTitle, ref)` inline |
| Title sort | `core/database/conclusionsummaries/order.go` `SortByTitle` | Root collation, case- and diacritic-insensitive; empty title last by ref; ties by lowercased ref. Calls `title()` twice per comparison (`sort.SliceStable` closure). Other callers: `core/database/persondetail/claims.go` (two), `core/database/persondetail/participants.go` |
| Keyed sort precedent | `core/database/conclusionsummaries/order.go` `sortByDateKey` | Computes each key once into a `keyed` slice, sorts that stably, writes rows back |
| Name formatting | `core/database/nameformats/nameformats.go` `Format` | Joins parts by the order; unlisted part types follow in stored order; empty order = stored order; no parts returns `form`. Doc says "This string is a sort key". Mirrors Swift `NameValueDisplay` |
| Western sorted order | `core/database/nameformats/registry.go` `seedProfiles` | `surname_prefix`, `surname` + `", "`, `prefix`, `given`, `nick`, `former_surname`, `suffix` |
| Part types | `core/namevalue/namevalue.go` `PartTypePrefix` = `"prefix"`; `PartTypeSurnamePrefix` = `"surname_prefix"` | — |
| Labels | `core/database/promote/promote.go` `fileSubject`, `core/database/identityclaims/bridgefile.go` | Create entities with no `Label`; only tests write a Person label |
| Tests | `core/database/conclusionsummaries/conclusionsummaries_test.go` `TestListPersons`, `TestListPersonsSortsBySortedPartsNotForm`, `TestListOrderIsByShownTitle`; `core/database/nameformats/nameformats_test.go` `TestInstallLoadsWesternAndFormats` | List content, sorted-style order, collation, Format; no prefix case |
| Test seam | `core/database/conclusionsummaries/export_test.go` `SortPlaces` | Restates `compose.go`'s Places sort line |
| Domain doc | `docs/structured-name-model.md` §4.5 requirements 2 and 5, and the closing paragraph ("The People list files by the project's sorted part order") | Say the key is the sorted order; req 5 leaves where `prefix` sits open |

**Precedent.** `sortByDateKey` in the same file for computing keys once. `TestListOrderIsByShownTitle` for a table-driven order test over hand-built rows. `TestListPersonsSortsBySortedPartsNotForm` for the end-to-end fixture (`nameformats.Install`, cite, promote, `f.list()`).

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | `nameformats.SortKey(name namevalue.Value, order []PartOrder) string` *(new)* is the one name sort key. It drops parts whose type is in `sortKeyOmits` *(new, unexported, `[]string{namevalue.PartTypePrefix}`)*; when that leaves no part, it keeps every part; it joins with `Format` and trims. A name with no parts returns `""`: `form` is never a sort input | `nameformats` owns name orders and `Format` already claims to be the sort key; one home for the rule, and the "Particle sort" idea becomes an entry or a profile field there. A prefix-only name keeps a visible title, so it sorts by it (AC-1.4) | A private filter in `conclusionsummaries` (map review M-2): a second definition of the key; falling to the label for a prefix-only name — it would drop to the ref tail |
| D-2 | `Format` output, the seeded orders and the profile tables do not change; `Format`'s doc drops "This string is a sort key" and points to `SortKey` | `Format` is display and mirrors Swift. Dropping `prefix` from the order cannot work: `Format` appends unlisted types (`Smith, John Mr.`) | Editing `seedProfiles` |
| D-3 | `personSortKey` returns `SortKey(*h.Name, order)` when `h.Name != nil` and that is non-empty, else the trimmed label | Same fallback as today, through the new key | — |
| D-4 | One generic `sortByKey[T, K any](rows []T, key func(T) K, less func(a, b K) bool)` *(new, `order.go`)* holds "compute each key once, sort stably, write rows back", lifted out of `sortByDateKey`. `sortByDateKey` becomes `sortByKey(rows, key, dateKeyLess)`. `SortByTitle` calls `sortByKey` with `titleKey{title, ref string}` *(new, unexported: trimmed title, lowercased ref)* and `titleKeyLess` *(new)* holding today's rule unchanged: empty title last, then `titleCollator.CompareString`, then ref. `SortByTitle`'s signature does not change | Measured at Scale (§6): 1.25 s → 0.37 s for 37,000 Persons; all five callers gain; one keyed-sort mechanism in the package (map review M-1, M-6) | Per-comparison keys (draft D-3); a second copy of the keyed loop beside `sortByDateKey` (round-1 D-4) |
| D-5 | `sortPersons(rows []PersonSummary, order []nameformats.PartOrder)` and `sortPlaces(rows []PlaceSummary)` *(new, `order.go`)* hold the two list sort lines; `persons` and `places` call them; `export_test.go` exposes `var SortPersons = sortPersons` and `var SortPlaces = sortPlaces`, replacing the restated `SortPlaces` func | Tests exercise the production line, not a copy (map review M-3) | A test-only `SortPersons` copy |
| D-6 | No Swift change and no FFI change | The app shows rows in Go's order (`ConclusionListPage`) | — |

**Rules this plan bends.** None. The working rule "(Batch 7) Persons list sort omits honorific prefixes" named label-fallback stripping; the user descoped that half on 2026-11-06 ("You can just descope it"), and the rule and the S14-34 scope entry were rewritten with this plan.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `nameformats.SortKey` *(new)* | func | `nameformats.Format` | Extends it: calls `Format` on the filtered parts. `Format` stays display, as Swift mirrors it |
| `sortKeyOmits` *(new)* | var | `seedProfiles` Roles in `core/database/nameformats/registry.go` (loaded as `Profile.Pattern`) mark `prefix` `RoleIgnored`; `core/matchregistry/value.go` the same | `RoleIgnored` also covers `surname_prefix`, which stays in the key (AC-1.5); roles are matching weights |
| `sortByKey`, `titleKey`, `titleKeyLess` *(new)* | func / type | `sortByDateKey` `keyed` | Lifted out of it; `sortByDateKey` and `SortByTitle` both call it |
| `sortPersons`, `sortPlaces` *(new)* | func | `compose.go` inline lines; `export_test.go` `SortPlaces` | Moves the lines; the test copy goes |
| `TestSortKey`, `TestPersonListOrder`, `TestListPersonsSortIgnoresHonorifics`, `TestSortByTitleKeysOnce` *(new)* | test | `TestInstallLoadsWesternAndFormats`, `TestListOrderIsByShownTitle`, `TestListPersonsSortsBySortedPartsNotForm` | New cases in the same shapes |

**Builds on.** `nameformats` (owner of name orders), `SortByTitle` (owner of title order), `sortByDateKey`'s keyed pattern.
**Deletes.** `sortByDateKey`'s own `keyed` loop (now `sortByKey`); the test-only `SortPlaces` func body in `export_test.go` (becomes a var to `sortPlaces`); the "This string is a sort key" sentence on `Format`; the dogfood note `ux/2026-10-15-persons-sort-ignores-honorifics.md` on land.
**Next adds.** "Particle sort" ([`docs/ideas/international-names.md`](../../../ideas/international-names.md)) needs a per-profile rule: it replaces `sortKeyOmits` with a profile field that `SortKey` reads, and `SortKey` stays the one place that reads it.

### Reach

No new path into existing code. One existing consumer reorders with the key:

| New path or state | Destination | Checked: resets on the new id · handles a missing referent · shows the state |
| --- | --- | --- |
| `PersonsByIDs` (through `persons`) orders by the new key | `core/database/persondetail/timeline.go` builds its participant index from that order; `core/database/persondetail/participants.go` `sortParticipants` breaks rank ties by it | Wanted: an event's participants of equal rank now order by name without honorifics. Covered by `persondetail` tests in STEP-2 and STEP-3 checks; no new state |

### Scale

At the map's Scale (37,000 Subjects; at most that many Persons rows), measured by the map reviewer with the real `SortByTitle` and `Format` on 37,000 Western-named rows: per-comparison keys 1.25 s (1,415,792 key calls); keys once 0.37 s. D-4 makes every `SortByTitle` caller compute n keys. `SortKey` allocates one filtered slice per call. No query is added (`TestListPersonsQueryCountIsConstant`).

### Edges

- **Assumptions.** `SortByTitle` needs a key that is a pure function of the row; `SortKey` reads only the name and order. Keys once must keep stability and the tie rule: T-4 and `TestListOrderIsByShownTitle` hold both.
- **Neighbors.** `surname_prefix` ("van") stays in the key (AC-1.5). A `prefix` anywhere in stored order is dropped (T-1 stored-order case). Two rows differing only by prefix tie on `Smith, John` and break by ref (T-2 f). `suffix` ("Jr.") stays.
- **Removed inputs.** A name of only prefix parts keeps them (AC-1.4). A name with no parts gives `""`, never `form` (T-1).
- **Real data.** None — no table or seed drives the rule; it reads one part type.
- **Untrusted size.** None — no new list or loop over imported data.

### Schema

None — no catalog change.

### Go core

| Package | Change |
| --- | --- |
| `core/database/nameformats` `nameformats.go` | `SortKey` *(new)*, `sortKeyOmits` *(new)* per D-1; `Format` doc per D-2 |
| `core/database/conclusionsummaries` `order.go` | `SortByTitle` keys once (D-4); `personSortKey` per D-3, doc says prefixes are not a sort input; `sortPersons`, `sortPlaces` *(new)* (D-5) |
| `core/database/conclusionsummaries` `compose.go` | `persons` calls `sortPersons(out, order)`; `places` calls `sortPlaces(out)` |
| `core/database/conclusionsummaries` `export_test.go` | `SortPersons`, `SortPlaces` as vars (D-5) |

Sketch:

```go
// SortKey is a name's list sort key: its parts joined in order, without
// honorific prefixes unless nothing else remains. form is never a sort key.
func SortKey(name namevalue.Value, order []PartOrder) string {
	if len(name.Parts) == 0 {
		return ""
	}
	var kept []namevalue.Part // a new slice; name.Parts is never changed
	for _, part := range name.Parts {
		if !slices.Contains(sortKeyOmits, part.Type) {
			kept = append(kept, part)
		}
	}
	if len(kept) == 0 {
		kept = name.Parts
	}
	return strings.TrimSpace(Format(namevalue.Value{Parts: kept}, order))
}
```

Errors: none.

### FFI

None — the wire is unchanged.

### macOS

None — D-6.

### Strings

None — no copy.

## 7. Steps

**STEP-1 — The name sort key**
Files: `core/database/nameformats/nameformats.go`, `core/database/nameformats/nameformats_test.go`. Skill: `add-name-value`.
Do: add `SortKey` and `sortKeyOmits` (D-1); change `Format`'s doc (D-2); add T-1.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/nameformats/` passes.

**STEP-2 — Title sort computes each key once; one line per list**
Files: `core/database/conclusionsummaries/order.go`, `core/database/conclusionsummaries/compose.go`, `core/database/conclusionsummaries/export_test.go`, `core/database/conclusionsummaries/conclusionsummaries_test.go`.
Do: D-4 (`sortByKey` first, `sortByDateKey` onto it, then `SortByTitle`) and D-5 (with `sortPersons` still calling today's `personSortKey`); add T-4.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/conclusionsummaries/ ./core/database/persondetail/` passes, `TestListOrderIsByShownTitle` included.

**STEP-3 — Persons key without honorifics**
Files: `core/database/conclusionsummaries/order.go`, `core/database/conclusionsummaries/conclusionsummaries_test.go`.
Do: D-3; add T-2 and T-3.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/conclusionsummaries/ -run 'TestPersonListOrder|TestListPersonsSortIgnoresHonorifics' -v` passes. With D-3 reverted, T-2 cases (a), (b), (c), (f) and T-3 fail; (d), (e) pass either way.

**STEP-4 — Docs and version**
Files: `docs/structured-name-model.md`, `docs/deployment-plan/spike-14/deployment-plan.md`, `docs/deployment-plan/spike-14/completed.md`, this plan, `docs/deployment-plan/spike-14/ux/2026-10-15-persons-sort-ignores-honorifics.md` (delete), `VERSION`, `core/version.go`, `macos/Provenencia.xcodeproj/project.pbxproj`.
Do: `structured-name-model.md` §4.5: requirement 2 adds "`prefix` parts (honorifics) are left out of the key unless the name has nothing else (`nameformats.SortKey`)"; requirement 5 says the key's `prefix` placement is decided and display's is not; the closing paragraph says "by the project's sorted part order, without honorifics". Map: row status `Landed #N`; the "Where the notes went" row → "Dogfood 2026-10-15: Persons sort ignores honorifics (note deleted YYYY-MM-DD; landed #N)"; the S14-34 scope entry's **Context** link to the note → "dogfood note 2026-10-15 (deleted on land)", as S14-33's entry reads. `#N` is the draft PR's number (`implement-pr` opens it; write this commit after); YYYY-MM-DD is the date of this commit. `completed.md`: a `## S14-34 — Persons sort ignores honorifics` section at the top, under the intro, in the S14-33 entry's shape. Patch bump per `.cursor/rules/versioning.mdc`: the next patch after `VERSION` at implement time (0.3.7 if nothing lands first) in `VERSION`, `core/version.go` and every `MARKETING_VERSION` line in the project file.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-14/prs/S14-34-persons-sort-ignores-honorifics.md` reports no errors; `scripts/check-pr.sh` prints `check-pr: PASS`.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Sort the Persons list without honorifics."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/database/nameformats/nameformats_test.go` `TestSortKey` *(new)*, table-driven. Western order: `Install` on a test catalog then `Load(db)`'s `Profile.Sorted`, as `TestInstallLoadsWesternAndFormats` does; no order literal. Nil order for the stored-order case | [Mrs. / given Smith] → `Smith`; [Mr. / John / Smith] → `Smith, John`; nil order [Dr. / Ann / Bell] → `Ann Bell`; [Rev.] only → `Rev.`; [van / Gogh / Vincent] → `van Gogh, Vincent`; [John / Smith / Jr.] → `Smith, John Jr.`; no parts, `Form` "Rev. Brown" → `""` | AC-1.1, AC-1.3, AC-1.4, AC-1.5, D-1 |
| T-2 | Go | `core/database/conclusionsummaries/conclusionsummaries_test.go` `TestPersonListOrder` *(new)*, table-driven through `SortPersons` over hand-built `PersonSummary` rows. Western order: `newFixture(t)`, `nameformats.Install(f.c)`, then `nameformats.LoadFor(f.c.Graph(), db)`'s `Profile.Sorted`; no order literal. Case (c) passes `nil` | (a) AC-1.1 rows → Adams, Nolan, Mrs. Smith; (b) AC-1.2 rows → Jane Smith, Mr. John Smith, Dr. Zed Smith; (c) nil order, AC-1.3 → Ann Bell, Carl Ames; (d) [Rev.] only + Zoe Quinn → Quinn, Rev.; (e) van Gogh + Ada Young → van Gogh, Young; (f) [Mr. John Smith] ref P1, [Dr. John Smith] ref P2 → P1, P2 (tie on `Smith, John`; today P2 first) | AC-1.1, AC-1.2, AC-1.3, AC-1.4, AC-1.5, D-3 |
| T-3 | Go | same file `TestListPersonsSortIgnoresHonorifics` *(new)*. Build each person inline as `TestListPersonsSortsBySortedPartsNotForm` does: `must(t, nameformats.Install(f.c))` first; per person `subjectstest.OfType(t, f.c, userID, f.graph, f.typeID("person"), "")`, `observationstest.Cite(… observations.Input{SubjectID, PropertyID: f.name.ID, Name: &namevalue.Value{Form: "<words joined by spaces>", Parts: …}})` with `Idx` 0..n in typed order, then `promotetest.Save`. A closure local to the test holds those calls; no new `namevaluetest` helper | AC-1.1 rows → `ListPersons` order Adams, Nolan, Mrs. Smith; the listed "Mrs. Smith" `Name` still holds its `prefix` part | AC-1.1, D-3, D-5 |
| T-4 | Go | same file `TestSortByTitleKeysOnce` *(new)* | 50 Place rows through `SortByTitle` with a title func that counts calls → 50 calls; order equals sorting by the same titles | D-4 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `TestListPersons`, `TestListPersonsSortsBySortedPartsNotForm`, `TestListOrderIsByShownTitle`, `TestListPersonsQueryCountIsConstant`, `core/database/persondetail` tests.

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
- [ ] Domain doc updated: structured-name-model.md §4.5 requirements 2 and 5 and the closing paragraph
- [ ] PR description: what shipped, AC evidence, deviations, manual check (or "manual check not run: no macOS" when the app cannot be launched)
```

## 10. Manual check

1. Open a project. In a Citation, type the name "Mrs. Smith" for one person (the splitter stores prefix "Mrs." + given "Smith"), "Dr. Zed Smith" for a second (prefix + given + surname), "Jane Smith" for a third and "Pat Nolan" for a fourth; Promote all four.
2. Open the Persons list, sorted by name. See Nolan first, then the three Smith rows together: "Mrs. Smith", "Smith, Jane", "Smith, Dr. Zed" in that order (keys `Smith`, `Smith, Jane`, `Smith, Zed`). Before this PR the order was Mrs. Smith, Nolan, Dr. Zed Smith, Jane Smith.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| A prefix-only name falls back to `form` | `SortKey` never reads `form`; T-1 |
| Within one surname, "Mrs. Smith" (no surname part) sorts apart from "Smith, Jane" | Accepted: key `Smith` sorts first among the Smiths; the splitter fix that would make it a surname is Out |
| Keys once changes an order a `persondetail` list relied on | Same comparison, still stable; `persondetail` tests in STEP-2's check |

**Stop and report if:**

- `personSortKey` is no longer the only Persons list key, or `persons` in `compose.go` no longer calls it.
- A caller of `SortByTitle` relies on its title func running per comparison (side effects, a counter).
- T-3's promoted name loses its `prefix` part through reconciliation (the kept name has no prefix) — then AC-1.1 is not reachable end to end and the reconciler's handling needs a decision.
- Any Swift test sorts Persons locally and fails.

## 12. Open questions

None.

## 13. PR

**Title:** S14-34 — Persons sort ignores honorifics
**Body:** objective in one sentence; what shipped against US-1; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-14/prs/S14-34-persons-sort-ignores-honorifics.md`.
