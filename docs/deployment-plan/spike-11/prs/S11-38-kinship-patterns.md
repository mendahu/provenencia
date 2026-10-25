# S11-38 — Kinship patterns

**Status:** Planned
**Spike / slice:** Spike 11 · Slice 9 — Kinship shapes
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no new chrome: a pure engine package with no caller yet)
**Board:** —
**Depends on:** S11-41 (landed [#561](https://github.com/mendahu/provenencia/pull/561))
**Blocks:** S11-39
**Branch:** `s11-38-kinship-patterns`
**Version bump:** none — a new package with no caller, so nothing changes for researchers ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-match-rule`](../../../../.agents/skills/add-match-rule/SKILL.md) ("Add a layer package"; STEP-5 edits it)
**Rules that bind:** `.cursor/rules/go-tests.mdc`, `.cursor/rules/match-rules.mdc`
**Context:** [Kinship shapes § Shape](../kinship-shapes.md#shape), [§ Counting once](../kinship-shapes.md#counting-once), [§ Weight](../kinship-shapes.md#weight); [segments](../comparison-layers.md#segments-layer-5-and-graph-alignment-layer-6); working rules: "Implied links, not a new comparison", "Three patterns, product terms only", "One pure package, `core/kinship`", "One fact, one link per side", "No new knob", "Pure means the package builds and tests with no build tags and no `core/database` import", "One guard, not one per package"
**Planned:** 2026-10-25 against `main` @ `846fa65`
**Review:** [`S11-38-kinship-patterns.review.md`](S11-38-kinship-patterns.review.md) — Fits with changes / Ready with changes, 1 round (findings fixed or logged)

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

A new pure package, `core/kinship`, reads one side's links through three kinship patterns and returns the relationship links they imply: two `spouse`s in one `marriage` imply a `spouse` relationship, a `father` or `mother` and the `subject` of one `birth` imply a `parent` relationship, and a parent of a parent implies a `grandparent` relationship. Each implied link carries the key a recorded relationship between the same two people would have, read through `segmentcompare.Keys`, and the links, bridge Subjects and event it was made from. Nothing calls it yet; S11-39 wires it into Promote's loaders.

**Why now.** It is the first build PR of slice 9 ([kinship shapes](../kinship-shapes.md)); S11-41 gave it the one link reading it needs. Build, then migrate: S11-39 adds the callers.

**What stays.** Every proposal, band and evidence line: no caller reads the package in this PR. `segmentcompare`, `graphalign` and the loaders are unchanged.

## 2. User stories

**US-1.** As the engine, given one side's links, I get the relationship links its marriages, births and parent chains imply, keyed as recorded relationships are, so a later PR can match them against the other side's recorded relationships.

Fixtures in every AC: one-byte Subject ids, people `P1`, `P2`, … = `0x11`, `0x12`, …; events `E1`, … = `0x21`, …; bridge Subjects `b1`, … = `0x31`, … (so `P1` sorts before `P2`); terms `segmentcomparetest.Terms()`; a participation edge has the person at `A` and the event at `B` unless the AC says otherwise; a relationship edge has the `person` endpoint at `A`. `rel(term)` below is `segmentcompare.Keys(relationship bridge, term, "", terms)`.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | `P1` and `P2` each `spouse` in `E1` (`marriage`), bridges `b1`, `b2` | One implied link: `A: P1, B: P2`, keys `rel(spouse)` (both `relationship·spouse`), `Pattern: marriage_spouses`, `Event: E1`, `Associations: [b1, b2]`, `From`: the two participation edges |
| AC-1.2 | AC-1.1 with both participation edges stored event first (`A: E1, B: P1`) | The same implied link as AC-1.1 |
| AC-1.3 | `P1` `father`, `P2` `mother`, `P3` `subject` in `E1` (`birth`), bridges `b1`, `b2`, `b3` | Two implied links: `P1 → P3` and `P2 → P3`, each keys `rel(parent)` (`KeyA` = `relationship·child·related_to`, `KeyB` = `relationship·child·person`), `Pattern: birth_parents`, `Event: E1`, `Associations` `[b1, b3]` and `[b2, b3]` |
| AC-1.4 | `P1` `parent` `P2` (relationship, `b1`), `P2` `parent` `P3` (`b2`) | One implied link `P1 → P3`, keys `rel(grandparent)`, `Pattern: parent_of_parent`, `Through: P2`, `Event` nil, `Associations: [b1, b2]`, `From`: both relationship edges |
| AC-1.5 | `P1` `parent` `P2` (`b1`); `P2` `father` and `P3` `subject` in `E1` (`birth`, `b2`, `b3`) | Two implied links: the birth's `P2 → P3` parent link, then `P1 → P3` grandparent with `Through: P2`, `Associations: [b1, b2, b3]`, `From`: `[P1–P2 edge, the implied P2 → P3 link's Edge]` |
| AC-1.6 | `P2` recorded as `child` of `P1` (relationship, `person` endpoint `P2`, term `child`), `P2` `parent` `P3` | `P1 → P3` grandparent, as AC-1.4: a `child` link reads as the parent link with the ends swapped |
| AC-1.7 | `P1` `grandparent` `P3` (relationship), no other links | Nothing: a grandparent implies no parent |
| AC-1.8 | `P1 → P2 → P3 → P4`, each a `parent` relationship | Two implied links, `P1 → P3` and `P2 → P4`; no link `P1 → P4` |
| AC-1.9 | A `spouse` relationship `P1`–`P2`; a `parent` relationship `P1 → P2` | Nothing: direct links are not returned |
| AC-1.10 | `P1`, `P2` each `groom` (a catalog's own role) in a `marriage`; `P1`, `P2` each `spouse` in a `wedding` (a catalog's own event type) | Nothing |
| AC-1.11 | `P1` `father` and `P2` `child` (role, not `subject`) in a `birth`; `P1` `witness` and `P2` `spouse` in a `marriage`; `P1` `father` and `P2` `subject` in a `marriage` | Nothing |
| AC-1.12 | `P1` `father` and `subject` of one `birth`; `P1` `parent` `P2` and `P2` `parent` `P1` | Nothing: no link joins a Subject to itself |
| AC-1.13 | `P1` `parent` `P2` (`b1`) and `P1` `father` of `P2`'s `birth` `E1` (`b2`, `b3`), `P2` `parent` `P3` (`b4`) | The birth's implied `P1 → P2`, then one grandparent `P1 → P3` with `From: [b1 edge, implied P1 → P2 Edge, b4 edge]` and `Associations: [b1, b2, b3, b4]` |
| AC-1.14 | No edges, or `nil` | `nil` |
| AC-1.15 | AC-1.1 with `P2`'s participation edge before `P1`'s in the input | The same implied link as AC-1.1: `A: P1`, `From` `[P1's edge, P2's edge]`, `Associations: [b1, b2]` |
| AC-1.16 | `P1` and `P2` each `spouse` in `E1` (`b1`, `b2`) and in `E2` (`b3`, `b4`), both `marriage` | Two implied spouse links `P1`–`P2`, the first with `Event: E1`, `Associations: [b1, b2]`, the second with `Event: E2`, `Associations: [b3, b4]` |

**US-2.** As a contributor, I add the next event pattern (an event type, the roles at each end, the implied term) as one table row, with its terms as `vocabkey` constants pinned to the seed.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | `CGO_ENABLED=1 go test -tags fts5 ./core/database/subjectvocab/` | `TestVocabKeysMatchSeed` passes and checks `marriage`, `father`, `mother`, `spouse` (role), and `spouse`, `parent`, `grandparent` (relationship type) |
| AC-2.2 | `go test ./core/kinship/... ./core/matchregistry/...` with no build tags | Passes; `TestLayerPackages` has a `core/kinship/pure` and a `core/kinship/numbers` subtest |
| AC-2.3 | `rg -n 'kinship' .agents/skills/add-match-rule/SKILL.md .cursor/rules/match-rules.mdc` | The skill's description lists the package, its "Rules that hold everywhere" list has the kinship bullet, it has an "Add a kinship pattern" section and its Check command lists `./core/kinship/...`; the rule's globs include `core/kinship/**` |

## 3. Scope

| In | Covered by |
| --- | --- |
| `vocabkey` constants `EventMarriage`, `RoleFather`, `RoleMother`, `RoleSpouse`, `TermSpouse`, `TermParent`, `TermGrandparent` *(new)*, with rows in `TestVocabKeysMatchSeed` (D-9) | STEP-1, AC-2.1 |
| `core/kinship` *(new)*: `Edge`, `Implied`, `Pattern` and its three codes, `Derive`, and the unexported event pattern table (D-1 – D-8, D-10) | STEP-2, STEP-3, AC-1.1 – AC-1.14 |
| Exact table test and a fixed-seed invariants test (D-10) | STEP-2, STEP-3, T-1, T-2 |
| `core/kinship` row in `layerPackages`, with `numbers` true | STEP-4, T-3, AC-2.2 |
| `.cursor/rules/match-rules.mdc` glob; `add-match-rule` description, rule bullet, "Add a kinship pattern" section and Check command | STEP-5, AC-2.3 |
| Docs owed (map row, `completed.md`, this plan) | STEP-6 |

| Out | Home |
| --- | --- |
| Any caller: the layer loader, `loadCanon`, merging implied links through `edgeSeenKey` | S11-39 |
| `canonLinks`, `claimevidence.Relationship` shapes, bundles and pin ids | S11-40 |
| The sheet's sentences | S11-42 |
| Patterns beyond the three (siblings, pibling, baptism parents, a `child` birth role, great-grandparent), conflicts between shapes, a project's own patterns | [kinship view idea](../../../ideas/kinship-view.md) |
| Family trees and person pages reading `core/kinship` | [kinship view idea](../../../ideas/kinship-view.md) |
| A registry knob or section for kinship | not planned — working rule "No new knob" |
| Merging an implied link with a direct one of the same ends and key | S11-39 — working rule "One fact, one link per side" puts it in the loaders' merge |
| Replacing the `"parent"` / `"spouse"` string literals in other packages' tests and in `segmentcomparetest.Terms` | not planned — tests and the seed-pinned fixture name terms by hand on purpose (S11-41 Out) |
| `docs/matching-model.md` and `docs/promote-graph-alignment.md` saying links are read through the patterns | S11-39, which makes it true |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Link reading | `core/segmentcompare/key.go` `Keys(b connectrules.Bridge, term vocabkey.BridgeTerm, eventType vocabkey.EventType, terms Terms) (first, second vocabkey.BridgeKey)` | A bridge's keys from its first endpoint's end and its second, through `LinkOf` and `Key`. Only a participation keeps the event type. Called by `loadLayerBridges` and `canonKeys` (S11-41) |
| Key canon | `core/segmentcompare/key.go` `Key` | A term and its inverse read as the one that sorts first, with the ends swapped: `parent` reads as `child`, `grandparent` as `grandchild`. A same-type link whose term is not directed has an empty `End` (`spouse`) |
| Bridges | `core/connectrules/connectrules.go` `LookupBridge`, `productBridges` | Participation: endpoints `person` (person), `event` (event). Relationship: `person`, `related_to` (both person). Pure: imports `core/vocabkey` only |
| Relationship meaning | `core/database/subjectvocab/registry.go` `relationship_type` description | "person is this type of related_to": a relationship with term `parent`, `person` `P1`, `related_to` `P2` says `P1` is `P2`'s parent |
| Seeded terms | `core/database/subjectvocab/registry.go` `seedTerms` | `event_type` `marriage`; `role` `father`, `mother`, `spouse`, `child`; `relationship_type` `spouse`, `parent` (inverse `child`), `grandparent` (inverse `grandchild`). All exist |
| vocabkey terms | `core/vocabkey/vocabkey.go` | Has `EventBirth`, `EventDeath`, `EventBurial`, `RoleSubject`, `TermPartOf`. No constant for `marriage`, `father`, `mother`, `spouse`, `parent` or `grandparent` |
| Seed check | `core/database/subjectvocab/vocabkey_test.go` `TestVocabKeysMatchSeed` | Term rows `{property, key}` checked against `seedTerms`; the role rows use the literal `"role"` property, the place one `"place_relationship_type"` |
| Term fixture | `core/segmentcompare/segmentcomparetest/terms.go` `Terms` | Every seeded role, relationship and place relationship term with the seed's directed and inverse marks |
| Loader shapes the input mirrors | `core/graphalign/types.go` `Bridge` (`A, B`, `KeyA, KeyB`, `Associations`), `CanonEdge` (`From, To`, `KeyFrom, KeyTo`, no bridge Subjects) | The two shapes S11-39 converts to and from `kinship.Edge` |
| Loader merge | `core/database/promotealign/terms.go` `edgeSeenKey`; `layer.go` `loadLayerBridges` | Same ends and same keys merge into one edge whose `Associations` gather every bridge. S11-39 sends implied links through it |
| Guard | `core/matchregistry/guard_test.go` `layerPackages`, `TestLayerPackages` | No `core/kinship` row. A `numbers` row may hold no numeric literal other than 0 and 1 |
| Rule | `.cursor/rules/match-rules.mdc` `globs` | Lists every layer package, not `core/kinship` |
| Skill | `.agents/skills/add-match-rule/SKILL.md` | Description, rule bullets and the Check command list the layer packages; no kinship |

**Precedent.** `e8046e37` ("Add the shortlist keys and index a Subject is looked up by…", S11-15) added a new pure layer-side package, `core/shortlist`, with its `layerPackages` row, its `.cursor/rules/match-rules.mdc` glob and an external-package table test. Copy that shape: package doc, the guard row, the glob, the `_test` package. Do not copy its registry section: this package has no knob. For the fixed-seed invariants test, copy `core/segmentcompare/segmentcompare_test.go` `TestInvariants` (`rand.New(rand.NewSource(1))`, generated cases, a property checked on each).

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | The input is `kinship.Edge{A, B []byte; KeyA, KeyB vocabkey.BridgeKey; Associations [][]byte}` *(new)*, one link between two primary Subjects read from each end, with its bridge Subjects. Its fields are `graphalign.Bridge`'s, in the same order with the same names, so S11-39 converts with `kinship.Edge(b)` and `graphalign.Bridge(imp.Edge)` The output is `[]kinship.Implied` *(new)*, each embedding the implied link as an `Edge` | The map scope: "takes a side's links (ends, `vocabkey.BridgeKey`, bridge Subjects)"; the event type is already in `BridgeKey.EventType` (map review K-11). The package is what trees and person pages read later, so it must not depend on Promote's alignment types | Take `[]graphalign.Bridge` — `core/kinship` would depend on layer 6, and `loadCanon`'s `CanonEdge` has a different shape anyway; take `segmentcompare.Link` — it is one end's reading, with no Subject ids |
| D-2 | `func Derive(edges []Edge, terms segmentcompare.Terms) []Implied` *(new)* is the one entry point. It reads the participation and relationship bridges with `connectrules.LookupBridge`; when either is missing it returns `nil` | One function per the map; `LookupBridge` is the registry every link reading already uses | A `Deriver` holding pre-built keys — no caller derives twice on one side, and S11-39 measures cost |
| D-3 | Every edge is recognized by its keys only, from either end, never by which end is `A`. A participation in role `r` of an event of type `e`: `pk, ek := segmentcompare.Keys(participation, r, e, terms)`; an edge with `(KeyA, KeyB) == (pk, ek)` has the person at `A`, `== (ek, pk)` at `B`. A relationship step of term `t`: `fk, sk := segmentcompare.Keys(relationship, t, "", terms)`; `(KeyA, KeyB) == (fk, sk)` reads `A t B`, `== (sk, fk)` reads `B t A` | The keys are the one reading (S11-41). `Key` folds `child` into `parent` with the ends swapped, so a recorded `child` link counts as a parent step with no extra rule (AC-1.6), and stored end order cannot matter (AC-1.2) | Read `Link` terms back out of the keys — duplicates `Key`'s inverse rule |
| D-4 | An implied link's `A` is the relationship's `person` end: `A` is `B`'s parent or grandparent. Its keys are `segmentcompare.Keys(relationship, term, "", terms)`. For `spouse`, whose two keys are equal, `A` is the pair's lesser id (`bytes.Compare`) and each unordered pair is made once per event | "person is this type of related_to" (seed); the same keys a recorded relationship has, as the working rule asks. A fixed order makes the output exact and testable | Keep input order for spouses — `P1, P2` and `P2, P1` in one marriage would both be made |
| D-5 | Event patterns, one implied link per event, `A` participation edge and `B` participation edge: the marriage pattern links every two different people who are each `spouse` in one `marriage`; the birth pattern links each `father` or `mother` to each `subject` of one `birth`. `Event` is the event's id, `From` is `[A's participation edge, B's participation edge]`, `Associations` their bridges in that order, for spouses too after D-4 puts the lesser id at `A` (AC-1.15). One implied link is made per event, `A` edge and `B` edge: the loaders already merge one person's same-role edges to one event, and the same pair in two marriages gives two links, each with its own `Event` (AC-1.16); the loaders' merge joins them | The working rule's patterns, exactly; a `child` birth role waits ([open question 3](../kinship-shapes.md#open-questions), proposed "birth with `subject` only") | Accept `child` as a birth role — not in the working rule |
| D-6 | The chain pattern runs after the event patterns, over every parent step: each input relationship edge that reads as `parent` (D-3) and each implied `parent` link from D-5. For each grandparent `G`, middle `M`, grandchild `C` with `G ≠ C`, it makes one link `G → C` with `Through: M`, `Event` nil, `From`: every parent step `G → M` then every parent step `M → C`, each list in step order (input edges first, in input order, then implied ones, in output order), and `Associations` the union of their bridges in `From` order, each once (`bytes.Equal`). A grandparent link, direct or implied, is never a step, so nothing deeper than grandparent is made | The working rule: "a parent of a parent (each direct or implied) implies grandparent; nothing deeper". One link per triple keeps a fact recorded in two parent shapes from doubling the chain (AC-1.13); it is not the loaders' merge, which joins links of the same ends and keys across different middles and shapes | One link per pair of steps — a parent recorded twice makes two identical grandparents, and S11-40's bundle would split them |
| D-7 | No link joins a Subject to itself: a pair with `A` equal to `B` is skipped in every pattern, and an input edge joining a Subject to itself is never a parent step | A person cannot be their own parent or spouse; such data is an entry error that matching should not reward | Return it — S11-39 would credit nonsense |
| D-8 | `Derive` returns only implied links, never an input edge, and does not merge an implied link with a direct one of the same ends and keys. It never modifies `edges`; returned slices share the input's byte slices, and a caller that keeps them clones them, as the loaders already do | Working rule "One fact, one link per side": the merge is the loaders' (`edgeSeenKey`, map review K-5). The loaders already copy ids into their own structs | Merge in `core/kinship` — a second merge beside the loaders' |
| D-9 | The terms are `vocabkey` constants in `core/vocabkey/vocabkey.go`: `EventMarriage EventType = "marriage"` beside `EventBirth`; `RoleFather`, `RoleMother`, `RoleSpouse Role` beside `RoleSubject`; `TermSpouse`, `TermParent`, `TermGrandparent BridgeTerm`: the single `const TermPartOf` becomes a `const (…)` block, `TermPartOf` under its place relationship comment and the three new ones under "Relationship types of a person to related_to." The edited blocks are re-aligned by `gofmt`. `TestVocabKeysMatchSeed` gains one row per constant | The working rule names `vocabkey` constants; the test is the existing seed pin. All seven terms are seeded | Constants in `core/kinship` — every other product term key lives in `vocabkey`, and S11-39 / S11-40 read them too |
| D-10 | The event patterns are unexported data in `core/kinship/patterns.go` *(new)*: `eventPatterns` (code, event type, `A` roles, `B` roles, implied term), which `Derive` loops. The chain is one rule in `Derive`, named by `vocabkey.TermParent` and `vocabkey.TermGrandparent`, with no table. `Pattern` codes: `PatternMarriageSpouses = "marriage_spouses"`, `PatternBirthParents = "birth_parents"`, `PatternParentOfParent = "parent_of_parent"` | A new event pattern is one row (US-2); the codes tell S11-40 how a side recorded the fact. Product terms only: the table names `vocabkey` constants, and a catalog cannot add a row. The one chain is parent of parent; siblings (two steps down from one middle) and great-grandparents (two different step terms) do not fit a `{step, term}` row, and nothing deeper than grandparent is derived (working rule), so a chain table would promise a slot it cannot keep (review M-2) | A `chainPatterns` table of `{step, term}` — fits only this one chain; a two-step row with middle ends — generality no roadmapped pattern needs yet ([`kinship-view.md`](../../../ideas/kinship-view.md)); exported or registered tables — a project's own patterns are an idea |
| D-11 | Output order: event patterns in table order, then events in the order their first participation edge appears in `edges`, then `A` edge order, then `B` edge order; chain links after them: a nested loop over the parent steps in D-6's step order (outer `G → M`, inner `M → C`) makes each triple where it is first reached. `Derive` returns `nil` when nothing is implied; `Event` and `Through` are `nil` when unset, never empty slices; a chain's `Associations` is never empty | Exact tests need one order; a permuted input gives the same set (T-2) | Sort by id — loses the input order the loaders' tie-breaks keep |
| D-12 | Tests: one external test file, `core/kinship/kinship_test.go` *(new)* (package `kinship_test`), with `TestDerive` (exact table, every AC of US-1) and `TestDeriveInvariants` (fixed seed). Edges are built in the test with `segmentcompare.Keys` over `connectrules.LookupBridge` and `segmentcomparetest.Terms()`; two helpers in the file, `part` and `rel`, build a participation and a relationship edge | The layer-package test rule in `add-match-rule` ("an exact table and a fixed-seed invariants test"); keys built through the same reading the loaders use | Hand-written `BridgeKey` literals — they would pin `Key`'s canon in a second place |

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `core/kinship` *(new)* | package | `core/segmentcompare` (reads links), `core/hops` (named graph steps) | `segmentcompare` scores one link of one Subject; kinship reads a whole side and returns new links. The working rule names its own package, the home trees and person pages read later. `hops` names SQL and graph walks, not facts |
| `kinship.Edge` *(new)* | type | `core/graphalign/types.go` `Bridge`, `CanonEdge` | The same fields as `Bridge`, by name and order, so the two convert directly; reusing the type makes kinship depend on layer 6 (D-1). `CanonEdge` has no bridge Subjects |
| `kinship.Implied`, `Pattern`, three `Pattern*` codes *(new)* | type, consts | `graphalign.Reason` codes | Codes of a different concept (how a fact was recorded); no existing type names a kinship shape |
| `kinship.Derive` *(new)* | func | none: no code derives a link from others (searched `rg -n 'grandparent\|implied' core --type go`) | — |
| `eventPatterns` *(new)* | table | `core/connectrules` `productBridges` (a product table looped by consumers) | Same idea, different facts; kept unexported in the one package that reads them |
| `vocabkey.EventMarriage`, `RoleFather`, `RoleMother`, `RoleSpouse`, `TermSpouse`, `TermParent`, `TermGrandparent` *(new)* | consts | `vocabkey.EventBirth`, `RoleSubject`, `TermPartOf` | Extends the same blocks |
| `TestDerive`, `TestDeriveInvariants`, helpers `part`, `rel` *(new)* | test | `core/segmentcompare/segmentcompare_test.go` `TestKeys`, `TestInvariants`; the `rel` closure inside `TestKey` | Same table shape; `TestInvariants` is the model for the fixed seed and the loop only. That `rel` builds a `Link`, this one an `Edge` with ids, so it is not shared |

**Builds on.** `segmentcompare.Keys` and `Key` (S11-41, S11-14); `connectrules.LookupBridge`; the `vocabkey` term blocks and `TestVocabKeysMatchSeed`; the one purity guard; the `core/shortlist` package precedent (`e8046e37`).
**Deletes.** None — nothing becomes dead; there is no earlier kinship code.
**Next adds.** S11-39 converts `graphalign.Bridge` and `CanonEdge` to `kinship.Edge`, calls `Derive`, and appends each `graphalign.Bridge(imp.Edge)` through `edgeSeenKey`. S11-40 reads `Pattern`, `Event`, `Through` and `From` for the line's shape and bundle. A new event pattern from [`kinship-view.md`](../../../ideas/kinship-view.md) (a baptism's parents) is a row in `eventPatterns` plus its `vocabkey` constants; a chain of another shape (siblings, great-grandparents) changes `Derive`.

### Schema

None — no catalog change.

### Go core

| Package | Change |
| --- | --- |
| `core/vocabkey` | Seven constants per D-9 |
| `core/kinship` *(new)* | `kinship.go`: package doc, `Edge`, `Pattern` and codes, `Implied`, `Derive`. `patterns.go`: `eventPatterns` |

Sketch (field comments in the code say the same as the decisions):

```go
// Package kinship reads one side's links through the kinship patterns: a
// shared marriage implies spouses, a birth's father or mother and subject
// imply a parent, and a parent of a parent implies a grandparent. Pure: no
// catalog, no I/O, no numbers; the catalog's bridge terms come in as an
// argument. It returns only implied links; callers merge them with direct ones.
package kinship

// Edge is one link between two primary Subjects: its two ends, the link read
// from each end, and the bridge Subjects behind it.
type Edge struct {
	A, B       []byte
	KeyA, KeyB vocabkey.BridgeKey // the link read from A's end and from B's
	// Associations are the bridge Subjects behind the link.
	Associations [][]byte
}

// Pattern is how an implied link was made.
type Pattern string

const (
	PatternMarriageSpouses Pattern = "marriage_spouses"
	PatternBirthParents    Pattern = "birth_parents"
	PatternParentOfParent  Pattern = "parent_of_parent"
)

// Implied is a relationship link a pattern found. A is the relationship's
// person end: the parent or grandparent; for spouses, the lesser id.
type Implied struct {
	Edge
	Pattern Pattern
	Event   []byte // the marriage or birth; nil for a chain
	Through []byte // the parent between, for a chain; nil otherwise
	From    []Edge // the links it was made from
}

func Derive(edges []Edge, terms segmentcompare.Terms) []Implied

// patterns.go
type eventPattern struct {
	code   Pattern
	event  vocabkey.EventType
	a, b   []vocabkey.Role
	term   vocabkey.BridgeTerm
}

var eventPatterns = []eventPattern{
	{PatternMarriageSpouses, vocabkey.EventMarriage, []vocabkey.Role{vocabkey.RoleSpouse}, []vocabkey.Role{vocabkey.RoleSpouse}, vocabkey.TermSpouse},
	{PatternBirthParents, vocabkey.EventBirth, []vocabkey.Role{vocabkey.RoleFather, vocabkey.RoleMother}, []vocabkey.Role{vocabkey.RoleSubject}, vocabkey.TermParent},
}
```

`core/kinship` imports `bytes`, `core/connectrules`, `core/segmentcompare` and `core/vocabkey` only. The chain reads a parent step through `segmentcompare.Keys(relationship, vocabkey.TermParent, "", terms)` for both input edges and the event patterns' implied `parent` links (D-3, D-6).

Errors: none. `Derive` cannot fail.

### FFI

None — no wire change.

### macOS

None — no Swift change.

### Strings

None — no copy.

## 7. Steps

**STEP-1 — Add the term constants**
Files: `core/vocabkey/vocabkey.go`, `core/database/subjectvocab/vocabkey_test.go`. Skill: `add-match-rule` ("Add a Subject property or pair", step 1, for the seed row).
Do: add the seven constants per D-9. In `TestVocabKeysMatchSeed`'s term table add `{vocabkey.PropertyEventType, string(vocabkey.EventMarriage)}`, `{"role", string(vocabkey.RoleFather)}`, `{"role", string(vocabkey.RoleMother)}`, `{"role", string(vocabkey.RoleSpouse)}`, `{"relationship_type", string(vocabkey.TermSpouse)}`, `{"relationship_type", string(vocabkey.TermParent)}`, `{"relationship_type", string(vocabkey.TermGrandparent)}`.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/subjectvocab/ ./core/vocabkey/` passes.

**STEP-2 — Add the package with the event patterns**
Files: `core/kinship/kinship.go` *(new)*, `core/kinship/patterns.go` *(new)*, `core/kinship/kinship_test.go` *(new)*. Skill: `add-match-rule` ("Add a layer package").
Do: write the package doc, `Edge`, `Pattern` and its three codes, `Implied` and `eventPatterns` per §6, and `Derive` with the event patterns only (D-2 – D-5, D-7, D-8, D-11). In the test file, the helpers per D-12: `part(person, event byte, role vocabkey.Role, eventType vocabkey.EventType, bridge byte) kinship.Edge` (person at `A`) and `rel(person, relatedTo byte, term vocabkey.BridgeTerm, bridge byte) kinship.Edge`; a test flips an edge for AC-1.2 by swapping `A`/`B` and `KeyA`/`KeyB`. Add `TestDerive` rows for AC-1.1, AC-1.2, AC-1.3, AC-1.15, AC-1.16, AC-1.9 (spouse relationship only), AC-1.10, AC-1.11, AC-1.12 (the birth half), AC-1.14. Each row compares the whole `[]kinship.Implied` with `reflect.DeepEqual`.
Check: `go test ./core/kinship/...` passes with no build tags.

**STEP-3 — Add the chain pattern and the invariants**
Files: `core/kinship/kinship.go`, `core/kinship/kinship_test.go` (new), both added in STEP-2.
Do: add the chain pass per D-6, D-7 and D-11. Add `TestDerive` rows for AC-1.4 – AC-1.8, AC-1.9 (parent relationship only), AC-1.12 (the cycle half), AC-1.13. Add `TestDeriveInvariants` per T-2.
Check: `go test ./core/kinship/...` passes with no build tags; `go vet ./core/kinship/...` is clean.

**STEP-4 — Join the purity guard**
Files: `core/matchregistry/guard_test.go`.
Do: add `{"core/kinship", true},` to `layerPackages` after the `core/segmentcomparetest` row (`core/kinship` builds on segment compare).
Check: `go test ./core/matchregistry/...` passes and `go test ./core/matchregistry/ -run 'TestLayerPackages/core/kinship' -v` lists `core/kinship/pure` and `core/kinship/numbers` as passing.

**STEP-5 — Teach the matching skill and rule**
Files: `.agents/skills/add-match-rule/SKILL.md`, `.cursor/rules/match-rules.mdc`.
Do: in the skill's `description`, add `core/kinship` to the package list after `core/segmentcompare`, and "a kinship pattern" to the "Use when adding" list. In the skill's "Rules that hold everywhere" list, add a bullet after the Layer 5 (`core/segmentcompare`) bullet: "**Kinship patterns (`core/kinship`) add implied links before alignment.** `Derive` reads one side's links (`Edge`: ends, keys from each end, bridge Subjects as `Associations`, the fields of `graphalign.Bridge`) through `eventPatterns` and one parent-of-parent chain rule and returns only the implied relationship links, keyed through `segmentcompare.Keys`; it recognizes a link by its keys alone. Callers merge implied links with direct ones through their own same-ends, same-key merge. An implied link weighs as the direct link it stands for: no knob." Add a section after "Add a link rule", headed "Add a kinship pattern", with four numbered items: (1) "Add a `vocabkey` constant for each term the pattern names, and its row in `TestVocabKeysMatchSeed`. Product terms only: a catalog's own terms imply nothing." (2) "An event pattern is a row in `eventPatterns` (an event type, the roles at each end, the implied relationship term) in `core/kinship/patterns.go` (new), with a new `Pattern` code. The chain is one rule in `Derive`, parent of parent only; a pattern of another shape (siblings, great-grandparents) changes `Derive` and is planned as its own PR." (3) "Add `TestDerive` rows: the pattern holds, a near miss (another role or event type) does not, and the implied link's direction." (4) "Add no knob: the implied link weighs as the direct link it stands for (the segment registry's Other pair)." Add `./core/kinship/...` to the Check command after `./core/segmentcompare/...`. In `.cursor/rules/match-rules.mdc`, add `core/kinship/**` to `globs` after `core/segmentcompare/**`; nothing else in that file changes.
Check: `rg -n 'kinship' .agents/skills/add-match-rule/SKILL.md .cursor/rules/match-rules.mdc` prints the description, the bullet, the section, the Check line and the glob (AC-2.3).

**STEP-6 — Docs**
Files: `docs/deployment-plan/spike-11/completed.md`, `docs/deployment-plan/spike-11/deployment-plan.md` (map row), this plan (status).
Do: map row S11-38 in `deployment-plan.md`: Status `Landed #N`. `completed.md`: a new section at the top (newest first), `## S11-38 — Kinship patterns`, then `**Landed YYYY-MM-DD** in [#N](https://github.com/mendahu/provenencia/pull/N).` and two or three sentences on what the package derives and that nothing calls it until S11-39, ending "Nothing changes for researchers." This plan: Status `Landed in [#N](https://github.com/mendahu/provenencia/pull/N) (YYYY-MM-DD)`, and a `**Landed:**` header line after **Planned** listing every difference from the plan ("None." when there is none).
Open the draft PR after the code commit, then make this docs commit with its real number.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-11/prs/S11-38-kinship-patterns.md` reports no errors; `scripts/check-pr.sh` prints `check-pr: PASS`. On a host with no Xcode, report the result of each Go step of `check-pr.sh` and name the Xcode step as not run.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits. For this PR: one code commit ("Derive the spouse, parent and grandparent links a side's marriages, births and parent chains imply.") and one docs commit.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/kinship/kinship_test.go` `TestDerive` *(new)* | One row per AC of US-1, input and exact output as the AC states; rows named by AC id | AC-1.1, AC-1.2, AC-1.3, AC-1.4, AC-1.5, AC-1.6, AC-1.7, AC-1.8, AC-1.9, AC-1.10, AC-1.11, AC-1.12, AC-1.13, AC-1.14, AC-1.15, AC-1.16, D-3, D-4, D-5, D-6, D-7, D-8, D-11 |
| T-2 | Go | `core/kinship/kinship_test.go` `TestDeriveInvariants` *(new)* | `rand.New(rand.NewSource(1))`, 200 cases; each draws 2 – 6 people, 1 – 3 events of type `birth`, `marriage` or a catalog `wedding`, and 0 – 10 edges: participations in roles `subject`, `father`, `mother`, `spouse`, `witness`, catalog `groom`, and relationships `parent`, `child`, `spouse`, `grandparent`, each with a fresh bridge id and a random stored end order. Each case checks: (a) every output's `(KeyA, KeyB)` is `rel(spouse)`, `rel(parent)` or `rel(grandparent)` and its `A ≠ B`; (b) every output association is an association of some input edge; (c) `Derive` of the edges in reverse order returns the same links (compared as a sorted list of `A, B, KeyA, KeyB, Pattern, Event, Through`); (d) appending edges whose roles and terms are only catalog ones (`groom`, a relationship `godparent`) leaves the output unchanged; (e) appending a direct `grandparent` relationship between two people adds no output | D-3, D-4, D-6, D-7, D-8, D-10, D-11 |
| T-3 | Go | `core/matchregistry/guard_test.go` `TestLayerPackages` (existing, new row) | `core/kinship/pure` and `core/kinship/numbers` pass | AC-2.2 |
| T-4 | Go | `core/database/subjectvocab/vocabkey_test.go` `TestVocabKeysMatchSeed` (existing, new rows) | The seven new constants are seeded terms of their properties | AC-2.1, D-9 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `TestLayerPackages` (a numeric literal such as a capacity hint in `make`, which the numbers guard refuses) and `TestVocabKeysMatchSeed` (a wrong property in a new row).

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left — none in this PR
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New — no UI in this PR
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together) — none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated (if gated) — not gated
- [ ] Domain doc updated if a rule in docs/*.md changed: none changes (S11-39 updates matching-model.md); add-match-rule and match-rules.mdc updated (STEP-5)
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

1. Run `go test ./core/kinship/... -run TestDerive -v` and read the row names: one per AC of US-1, all passing.
2. Run `rg -n 'kinship' .agents/skills/add-match-rule/SKILL.md .cursor/rules/match-rules.mdc` and see the description, bullet, section, Check line and glob (AC-2.3).
3. Run `go list -deps ./core/kinship | rg 'provenencia/core/database'` — it prints nothing.
4. No app check: nothing calls the package until S11-39, whose manual check covers slice 9's Check.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| A key mismatch makes a pattern silently never hold (a wrong role constant, a wrong event type) | T-1's exact rows fail on an empty output; D-3 builds the expected keys through `segmentcompare.Keys`, never by hand |
| The chain multiplies when a parent is recorded in two shapes | D-6 makes one link per triple; AC-1.13 pins it |
| The numbers guard refuses a literal in `core/kinship` | Write no literal other than 0 and 1 in non-test files: no `make(…, 0, 2)`, no `[2]`, no `len(x) < 2`; pair with nested `range` loops and append. Test files are not checked |

**Stop and report if:**

- `segmentcompare.Keys` or `connectrules.LookupBridge` no longer has the signature in §4.
- `core/kinship` importing `core/segmentcompare` gives an import cycle, or `TestLayerPackages` reports a path to `core/database`.
- `segmentcompare.Keys(relationship, vocabkey.TermParent, "", segmentcomparetest.Terms())` does not return `relationship·child·related_to`, `relationship·child·person` (AC-1.3's keys): the key canon moved and D-3 / D-4 need rechecking.
- A seeded term in D-9 is missing from `seedTerms`, or `TestVocabKeysMatchSeed` fails on one.
- An AC's expected output can only pass by changing `segmentcompare`, `graphalign` or `promotealign` (Out).

## 12. Open questions

None.

## 13. PR

**Title:** S11-38 — Kinship patterns
**Body:** objective in one sentence; what shipped against US-1 and US-2; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-11/prs/S11-38-kinship-patterns.md`.
