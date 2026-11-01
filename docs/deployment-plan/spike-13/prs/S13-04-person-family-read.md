# S13-04 — Person family read

**Status:** Planned
**Spike / slice:** Spike 13 · Slice 3 — Family
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no chrome change: an engine read, an RPC, a store method and a cache key; no view reads them until S13-12)
**Board:** —
**Depends on:** S13-01 (landed [#655](https://github.com/mendahu/provenencia/pull/655)), S13-02 (landed [#659](https://github.com/mendahu/provenencia/pull/659)), S13-31 (landed [#660](https://github.com/mendahu/provenencia/pull/660)), S13-03 (landed [#666](https://github.com/mendahu/provenencia/pull/666)), S11-51 (landed: `canonlinks.Reader.Relationships`, `kinship.LinkSet`)
**Blocks:** S13-12
**Branch:** `s13-04-person-family-read`
**Runs on:** macOS: STEP-4 needs `protoc` and its plugins, and STEP-6, STEP-7 and `scripts/check-pr.sh` need Xcode
**Version bump:** none — Spike 11–14 PRs leave `VERSION` alone until a release is cut, as S13-01 and S13-03 did ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-match-rule`](../../../../.agents/skills/add-match-rule/SKILL.md) (the `core/kinship` and `canonlinks` lines), [`add-ffi-handler`](../../../../.agents/skills/add-ffi-handler/SKILL.md), [`use-catalog-session`](../../../../.agents/skills/use-catalog-session/SKILL.md), [`add-catalog-query`](../../../../.agents/skills/add-catalog-query/SKILL.md), [`add-workspace-place`](../../../../.agents/skills/add-workspace-place/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md)
**Rules that bind:** `.cursor/rules/go-tests.mdc`, `.cursor/rules/match-rules.mdc`; `core/matchregistry/guard_test.go` lists `core/kinship` as a numbers package, so `families.go` holds no numeric literal but 0 and 1 (`iota + 1` passes)
**Context:** [person-detail.md § 2. Family](../person-detail.md#2-family) (R-2, R-16 – R-22), [`docs/ideas/kinship-view.md`](../../../ideas/kinship-view.md), working rules: "Each section reads its own query", "One invalidation rule for person-section keys", "One person summary for every tile", "Canonical only, filtered in Go", "Siblings are page-only", "One person's links, direct and implied, from one function", "Family groups in Go", "The family read is served from the graph cache", "The family read is the family tree's first step", "Engine PRs run ahead of their slice"
**Planned:** 2026-11-01 against `main` @ `60a7e12`
**Review:** [`S13-04-person-family-read.review.md`](S13-04-person-family-read.review.md) — round 1 Fits with changes / Not ready, round 2 Fits with changes / Ready with changes; every finding Fixed, none open

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

The engine can answer "who is this person's family": a canonical person's parents, siblings, spouses and children, grouped into family units (each set of parents with the siblings they share, each partner with the children the two have, the children with no other known parent, and the siblings whose parents are not known), each relative as a `PersonSummary` with the link that says why they are there. The app can fetch it under its own per-person cache key. A second read of the same family issues no SQL, and a first read loads in a fixed number of batches however large the family.

**Why now.** S13-12 (Family section, brief S13-D3) draws this read. Every engine piece it composes has landed: S11-51's `Relationships`, S13-03's siblings, S13-01's accepted-members helper and per-person key pattern, S13-02's gender terms on the summary, and S13-31's names.

**What stays.** `GetConclusionDetail`, `ListPersonClaims` and every view are unchanged. Matching is unchanged: nothing here is in `kinship.Derive`, `Patterns()`, `Reader.For`'s answer or `promotealign` (working rule "Siblings are page-only"). `Reader.Relationships` returns what it returns today.

## 2. User stories

Names below are people built with `graphcachetest.Canon`: each is a canonical Person with one accepted member unless the AC says otherwise. "Birth B: John father, Mary mother, Pat subject" is a birth event with three participation links. Units are written `Kind{partners; children}`. Partners and children within a unit are compared as sets; the order of units is asserted only where D-3 and D-4 fix it whatever the link order (kinds in D-3's order, a person's own parent sets before half-sibling sets, recorded links before implied ones).

**US-1.** As the engine, when asked for a canonical person's family, I return their family units so that the Family section can draw each family with its lines.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | Two marriages. Marriage M1: Pat and Sam subjects; marriage M2: Pat and Tess subjects. Recorded parent links Pat→Uma and Sam→Uma. Birth B2: Pat father, Tess mother, Vic subject | Two units, `Partner{Sam; Uma}` and `Partner{Tess; Vic}`. Sam's link is implied only, one shape `{marriage_spouses, M1}`; Vic's link is implied only, `{birth_parents, B2}`; Uma's link is recorded with Pat→Uma's association |
| AC-1.2 | Conflicting parents. Birth B1: John father, Mary mother, Pat subject. Recorded parent link James→Pat | `Parents{James;}` then `Parents{John, Mary;}`: James's recorded link comes before the birth's implied ones, and a lone parent is emitted at its place (D-4) |
| AC-1.3 | Recorded parent links John→Pat and Mary→Pat, and a recorded spouse link John–Mary | `Parents{John, Mary;}`. The same without the spouse link: two units, `Parents{John;}` and `Parents{Mary;}` (D-5) |
| AC-1.4 | A recorded spouse link Pat–Sam and marriage M1 with both as subjects | One `Partner{Sam;}`; Sam's link is recorded (not only implied) with the shape `{marriage_spouses, M1}` and the spouse link's association |
| AC-1.5 | Birth B1 (John, Mary, Pat). Birth B3: John father, Mary mother, Quin subject. Recorded parent link Mary→Rae | `Parents{John, Mary; Quin}`, then `Parents{Mary; Rae}`. Quin's link has two `shared_parent` shapes, through John and through Mary; Rae's has one, through Mary |
| AC-1.6 | A recorded sibling link Pat–Zed, and Zed shares no parent with Pat. A recorded sibling link Pat–Quin where Quin is also Pat's sibling through B1's parents (as in AC-1.5) | `Parents{John, Mary; Quin}` with Quin's link recorded and implied; `Siblings{; Zed}`. Quin is not in the `Siblings` unit |
| AC-1.7 | Recorded parent links Pat→Wren and Xan→Wren; no spouse link or marriage between Pat and Xan | `Partner{Xan (no link); Wren}` (D-6) |
| AC-1.8 | Recorded parent link Pat→Yul, and Yul has no other parent. A spouse link Pat–Sam and no child of both | `Partner{Sam;}` (no children), then `OneParent{; Yul}` |
| AC-1.9 | Recorded parent links John→Pat and Pat→Uma. John's only claim is provisional. Uma was merged into another Person | No unit names John or Uma; the person has no units |
| AC-1.10 | Recorded parent link John→Pat and a recorded spouse link Pat–John | John is in `Parents{John;}` and in `Partner{John;}` |
| AC-1.11 | Recorded grandparent link Gus→Pat, cousin link Pat–Cy, guardian link Gil→Pat, and a chain John→Mid→Pat | No unit names Gus, Cy or Gil; John (Pat's grandparent through Mid) is in no unit; Mid is in `Parents{Mid;}` |
| AC-1.12 | Entity id unknown, of an Event, or merged; a malformed id | `conclusiondetails.ErrNotFound` from `PersonFamily`; `Families` skips the id |
| AC-1.13 | An unmerged Person with no relationship link and no event | Zero units, no error |
| AC-1.14 | `Families(c, [Pat, Sam, Uma])` over AC-1.1's family | Three `Family` values in that order; each equals `PersonFamily`'s for that id |

**US-2.** As the engine, I serve the family from the graph cache, so that the page reads it cheaply and a tree can read a generation at once.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | `PersonFamily(Pat)` read twice on AC-1.1 + AC-1.5's family | The second read issues zero graph queries (`SetQueryCounterForTest`) and returns an equal result |
| AC-2.2 | A cold `PersonFamily` on (a) Pat with one parent pair from a birth, one sibling, one marriage and one child with a birth; (b) the same with seven siblings and seven children, each with a birth, and a second marriage | Both issue the same number of graph queries, and it is more than zero |
| AC-2.3 | A cold `Families` on one of (a)'s people, and on three people of three copies of (a) | Both issue the same number of graph queries |
| AC-2.4 | After `Reader.LoadRelationships(ids, Options{Siblings: true})`, `Relationships(id, Options{Siblings: true})` for each id | Zero queries; a cold `LoadRelationships` issues the same number for one child and for seven children with births, and for one id and three |
| AC-2.5 | After AC-2.1's warm read, a parent link Pat→Ned is filed and committed (`Canon.Commit(Pat, Ned, link)` with the link's association id, as `canonlinks` tests commit `w.k.Commit(id)`) | The next read has Ned in `OneParent{; Ned}`; after the link is removed and committed, the read after has no Ned |

**US-3.** As the app, I fetch a person's family over FFI.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | `GetPersonFamily` on AC-1.1 + AC-1.7's family | Three units, each `FAMILY_UNIT_KIND_PARTNER`; each member's `person` is that relative's `PersonSummary` (entity id, ref, name); Sam's `link.recorded` is false with one `implied` shape `{KINSHIP_PATTERN_MARRIAGE_SPOUSES, M1}`; Uma's `link.recorded` is true with Pat→Uma's association id in `association_ids`; Xan has no `link` |
| AC-3.2 | Unknown or malformed `entity_id` | Error code `conclusiondetails.not_found` |

**US-4.** As the app, I keep a person's family in the session cache and refresh it when it may be stale.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-4.1 | A seeded family for e1 in `FakeStore`; a query of `.personFamily(project:, entityId: "e1")` | The handle's value is the seed; e9 (unseeded) throws `conclusiondetails.not_found` |
| AC-4.2 | A promote (`.promotedSubject`), a Source delete, a Property update | Each invalidates `.allCached(.personFamily)`; no other mutation does |
| AC-4.3 | A warmed `.personFamily` key for e1; write effects touching only `e7` (a relative's handle) | The key is revalidated; with no touched handle it is not; the key's hidden policy is `.evict` and its stale policy `.sessionFresh` |

## 3. Scope

| In | Covered by |
| --- | --- |
| `canonlinks.Reader.LoadRelationships(ids, opts)`: one batched load for what `Relationships` reads for many ids, sharing `forLinks`' step-people walk | STEP-1, AC-2.4 |
| `kinship.KinOf` and `kinship.FamilyUnits`: pure nuclear roles and family units with their links | STEP-2, AC-1.1 – AC-1.11 |
| `persondetail.Families(c, ids)` and `persondetail.PersonFamily(c, id)`: load, filter to canonical relatives, group, attach summaries | STEP-3, AC-1.9, AC-1.12 – AC-1.14, AC-2.1 – AC-2.3, AC-2.5 |
| `GetPersonFamily` RPC (proto, regenerate, handler, dispatch) | STEP-4, STEP-5, AC-3.1, AC-3.2 |
| Swift: wire mapping, `GenealogyStore.getPersonFamily`, `FakeStore` seed, `.personFamily` key and descriptor | STEP-6, STEP-7, AC-4.1 – AC-4.3 |
| Docs the PR makes false | STEP-8 |

| Out | Home |
| --- | --- |
| The diagram, tiles, lines, ordering tiles by birth (placement is Swift's, working rule "Engine PRs run ahead of their slice") | S13-12 |
| Grandparents, aunts and uncles, cousins, guardians and deeper generations | [`docs/ideas/kinship-view.md`](../../../ideas/kinship-view.md) |
| A tree RPC, a layout engine, saved tree views | [`docs/ideas/kinship-view.md`](../../../ideas/kinship-view.md) |
| A half-sibling's other parent (the parent Pat does not share) | S13-12 follow-up if the board draws it (working rule "If a board asks for data a landed read lacks, a follow-up PR adds it"); the sibling's own family read gives it |
| Flagging conflicts between shapes (a sibling who is also a spouse) | [`docs/ideas/kinship-view.md`](../../../ideas/kinship-view.md) "Conflicts between shapes" |
| Moving link reading out of `segmentcompare` (now that `persondetail` imports `core/kinship` directly, K-12's "only through `canonlinks`" no longer holds; the weight is still already paid, since `persondetail` imports `canonlinks`, which imports `segmentcompare`; STEP-8 adds this to K-12) | Declined in the map review (K-12); stays in [`docs/ideas/kinship-view.md`](../../../ideas/kinship-view.md) |
| A memo slot for the grouped family | Not planned (D-9) |
| Sibling evidence in matching | [`docs/ideas/kinship-view.md`](../../../ideas/kinship-view.md) (R-18a) |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| One person's links | `core/database/canonlinks/reader.go` `Reader.Relationships(id, opts)` | Returns `id`'s relationship links, recorded then `For`'s then (with `Options{Siblings: true}`) the shared-parent siblings, merged through one `kinship.LinkSet`. Calls `d.around(id)`, `r.For(id)` and `r.LoadSiblings([][]byte{id})`. Nothing filtered |
| `For`'s loads | `reader.go` `Reader.For`, `deriver.forLinks`, `deriver.around` | `forLinks` builds `people` (each person one parent step from `id`, from `kinEdges(n, out)`) inline, then calls `d.around(m)` per person: a cold `For` loads one node and one neighbor batch per birth-implied parent or child. Nothing batches it for many ids (S13-03 plan, Out row "`For` loading around many people in one batch — S13-04") |
| Siblings | `reader.go` `Reader.LoadSiblings`, `deriver.siblingsOf`, `siblingSlot` | Four batched loads (ids, neighbors, parents, parents' neighbors) for many ids; stored per person; dropped on any handle commit |
| Reader construction | `reader.go` `NewReader(g, rules)`; its non-test caller `core/database/promotealign/canon.go` passes `matchregistry.Default().Kinship` | Loads terms and the event type property once |
| Parent steps | `core/kinship/kinship.go` `ParentStep`, `Edge.From` | The one rule for a parent step, from either end; `From` reads an edge from one end |
| Link merge | `core/kinship/links.go` `Link`, `LinkSet` | One link per pair of ends and keys; `Recording` says recorded or only implied, with its `Shapes` (`Pattern`, `Pivot`) |
| Patterns | `core/kinship/kinship.go` `PatternMarriageSpouses`, `PatternBirthParents`, `PatternParentOfParent`; `siblings.go` `PatternSharedParent` | The codes a link's shape carries |
| Kin terms | `core/vocabkey/vocabkey.go` `TermSpouse`, `TermParent`, `TermGrandparent`, `TermSibling`; seed `core/database/subjectvocab/registry.go` | `spouse` and `sibling` undirected; `parent` directed with inverse `child`; also `cousin`, `grandparent`, `pibling`, `guardian` and inverses |
| Page-handle guard | `core/database/conclusiondetails/conclusiondetails.go` `Handle`, `ErrNotFound` | nil catalog or graph, a malformed id, unknown or merged → `ErrNotFound` |
| Person section reads | `core/database/persondetail/claims.go` `Claims` | The package's one read today; uses `Handle` and `Node.AcceptedMembers` |
| Accepted members | `core/database/graphcache/graph.go` `Node.AcceptedMembers`, `Node.Merged`, `Node.SubjectTypeID` | S13-01's helper |
| Person type id | `core/database/graphcache/vocab.go` `Vocabulary.TypeIDsByKey` | Provenencia type key → id; read by `core/database/promotealign/profiles.go` |
| Summaries | `core/database/conclusionsummaries/compose.go` `ByIDs`, `Want.Add`, `Summaries.Person`, `Summaries.PictureImageFileIDs` | Summaries by id for unmerged Persons; warm reads come from the display slot |
| Summary wire | `api/ffi/handlers/conclusion_summaries.go` `personSummaryProto`; `api/ffi/handlers/search_catalog.go` `fillHitSummaries` | Map a summary with its picture path from one `files.RelPaths` call |
| Claims handler | `api/ffi/handlers/person_detail.go` `ListPersonClaims` | The precedent handler: `parseID` failure → `ErrNotFound`, `withProjectCatalog`, one read |
| Dispatch | `api/ffi/dispatch.go` `MethodListPersonClaims`, `case MethodListPersonClaims` | One const and one case per method |
| Proto | `api/proto/engine.proto` `Method` (last: `METHOD_LIST_PERSON_CLAIMS = 112`), `ListPersonClaimsRequest`, `PersonSummary`, `DerivedRowKind` | `DerivedRowKind` is the precedent for a closed proto enum; numbers 113 and up are unused |
| Swift store | `macos/App/Platform/GenealogyStore.swift` `CatalogPersonSummary`, `CatalogPersonClaimGroup`, protocol `listPersonClaims`; `GoStore.swift` `listPersonClaims`, `mapPersonSummary` (private), `mapDerivedRow`; `FakeStore.swift` `seededPersonClaims`, `listPersonClaims`; `CoreInvoke.swift` `CoreMethod.listPersonClaims` | S13-01's shape; `mapDerivedRow` returns nil for `.unspecified, .UNRECOGNIZED` |
| Cache key | `macos/App/Features/Workspace/Session/CatalogQueryKey.swift` `.personClaims`, `Kind.personClaims`; `CatalogQueryRegistry.swift` descriptor `case .personClaims`, `conclusionTriggers`; `CatalogLastUpdated.swift` `CatalogQueryRefresh.applyHiddenPolicyOnAny(follows:)` | `.personsList` uses `applyHiddenPolicyOnAny(follows: .touchedHandles)`; `.personClaims` follows its own handle and its graphs |
| Tests | `core/database/canonlinks/reader_test.go` `world`, `TestLoadSiblingsQueryCount`; `core/kinship/kinship_test.go` `terms`, `rules`, `p`, `e`, `rel`, `part`; `core/database/persondetail/claims_test.go`; `api/ffi/handlers/person_detail_test.go` `TestListPersonClaims`; `core/catalogsession/warm_test.go` `newProject` (Canon inside `catalogsession.Do`); `macos/ProvenenciaTests/CatalogQueryRegistryTests.swift` `invalidationsComeFromRegistrySpecs`, `personClaimsInvalidateOnPromote`, `personClaimsLoadThroughTheStore`; `macos/ProvenenciaTests/WorkspaceSessionTests.swift` `personClaimsFollowHandlesAndGraphs` | `invalidationsComeFromRegistrySpecs` lists the exact invalidations of `.deletedSource` and `.promotedSubject`, and asserts every Conclusion trigger stales `.allCached(.personClaims)` |

**Precedent.** S13-01 ([#655](https://github.com/mendahu/provenencia/pull/655), merge `fdd37f19`) added a per-person read in `persondetail`, its handler, proto, `CoreMethod`, store method, `FakeStore` seed, `CatalogQueryKey` case and descriptor, and their tests: copy that file list and each file's shape. S13-03 ([#666](https://github.com/mendahu/provenencia/pull/666)) added `LoadSiblings` and `TestLoadSiblingsQueryCount`: copy its batched-load shape and its "same count for 1 and 7, for 1 and 3 ids" test. Do not copy S13-01's `.evidenceGraphs` refresh: the family key follows handles only (D-12).

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | Three layers. `core/kinship/families.go` *(new)* holds the pure rules: `KinOf` reads a person's links as nuclear roles, `FamilyUnits` groups roles into units. `core/database/persondetail/family.go` *(new)* loads, filters to canonical relatives and calls them. The handler maps to the wire | "Family groups in Go" and "data, not layout"; each case tested at the lowest layer that can see it (S13-03 D-12); `kinship` is already the pure home of kin rules (`Siblings`, `ParentStep`) | All grouping in `persondetail` — only catalog tests could reach the cases; grouping in Swift — the working rule |
| D-2 | One reader of a relationship link, one table of roles. `readsTerm(e Edge, term vocabkey.BridgeTerm, terms segmentcompare.Terms) (first, second []byte, ok bool)` *(new, package-private)* reads an edge as one relationship term from either end: `relationshipKeys(term, terms)` *(new, package-private: the one `LookupBridge(BridgeRelationship)` + `segmentcompare.Keys` lookup)*, then `e.reads(first, second)`, false for a link whose ends are one id. `ParentStep` becomes `readsTerm(e, TermParent, terms)` (its doc keeps "the one rule for what a parent step is"); `chains`, `Siblings` and `eventLinks` take their keys from `relationshipKeys`, so `Derive` no longer looks up the relationship bridge and `eventLinks` and `chains` lose their `rel` parameter (`Derive` keeps its participation lookup). Roles come from a declared table in `families.go`, `nuclearRoles` *(new)*: `{RoleParent, TermParent, second}`, `{RoleChild, TermParent, first}`, `{RoleSpouse, TermSpouse, first}`, `{RoleSibling, TermSibling, first}`, in that order, where the end is where the person stands (`parent`'s first end is the parent, as `ParentStep` reads it; for the undirected terms both ends read alike). `KinOf(person, links, terms)` puts a link in each row where `readsTerm` succeeds and the row's end is `person`. Any other term (grandparent, cousin, pibling, guardian, their inverses) matches no row. `Kin` is `type Kin map[Role][]Link`, read as `kin[RoleParent]`; a link keeps its input order within a role; a relative may be in several roles (AC-1.10). A grandparent row is one more table entry through the same reader | One rule for "which end is which" across `Derive`, `Siblings`, `Parents`, `stepPeople` and `KinOf` (review M-10); one key lookup in the package, `eventLinks` included (M-11, M-8); the package's table style, as `eventPatterns` (M-3) | A per-role `Kin` field and branch — each named add repeats both; matching one end's key through `Edge.From` — a second parent-step rule beside `ParentStep`; classifying by `TermKey` strings — misses the inverse filing (`child`) that `Key` folds |
| D-3 | `FamilyUnits(person, kin, kinOf)` returns, in this order: (1) parent units, (2) partner units, (3) one `OneParent` unit when any child has no other parent, (4) one `Siblings` unit when any sibling has no shown shared parent. Units with no members are not returned. Every `Relative.Link` is `person`'s own link to that relative, so it says why they are there (recorded associations, implied shapes); a partner known only through a child has `Link == nil` | "Units, not four lists"; "Every link says why"; one link per relative and role, from the `LinkSet` merge | A link per relative pair (partner→child) — the next generation's read gives it |
| D-4 | Parent units. Parents are `kin[RoleParent]` in order. Two parents are one set when (a) either's `kinOf` has a spouse link to the other, or (b) `person`'s links to both carry a `birth_parents` shape with the same pivot. Each paired couple (i < j) is a set `{qi, qj}`; a parent in no pair is a set alone; sets are emitted walking i, each pair at i in j order, a lone parent at its i. A birth naming three parents (two fathers and a mother) pairs all three ways: `{F1, F2}`, `{F1, M}`, `{F2, M}` (malformed data is shown, not resolved). Each sibling with `shared_parent` shapes has shared parents S = the shapes' pivots that are shown parents. It joins every set wholly inside S. A sibling that joins no set joins a unit whose partners are exactly S (in parent order), made after the person's own sets in first-seen sibling order, one per distinct S. A sibling with no shown shared parent (recorded only, or implied only through a parent not shown) goes to the `Siblings` unit | R-19 "a person with conflicting parents shows each set"; user, 2026-11-01: parents pair by evidence, a parent paired with nobody stands alone. A set holds the siblings of every parent in it, so a half-sibling sits with the parent they share (AC-1.5) | One set of all parents — AC-1.2 shows three parents as one couple; "one set unless 3+" — offered, the user chose evidence |
| D-5 | No fallback pairing: two parents with no spouse link and no shared birth are two one-parent sets, even when they are the person's only two (AC-1.3) | User, 2026-11-01 (pair by evidence); a pairing the data does not hold would hide a conflict | Pairing the only two parents — offered and declined |
| D-6 | Partner units. Spouses are `kin[RoleSpouse]` in order. For each child c, its other parents are c's `kinOf` `RoleParent` links minus `person`, in order. Partners are the spouses, then each other parent of a child who is not a spouse, first-seen over children in order. Each partner's unit holds the children who have them as an other parent; a spouse's unit may have no children (AC-1.8). A child with no other parent goes to the `OneParent` unit; a child with two other parents is in both units. A co-parent who is not a spouse is a partner with `Link == nil` | User, 2026-11-01: a co-parent gets a unit, marked as having no link to the person. The working rule's "one family per spouse (with the children who have both as parents)" and "children with one known parent in their own group" | Co-parents' children in `OneParent` — offered and declined |
| D-7 | Canonical only: `persondetail` drops, before the second `KinOf`, every link whose other end is not shown, from the person's links and from each relative's. Shown is one named function, `shownPerson(n *graphcache.Node, personTypeID []byte) bool` *(new)* in `family.go`: not nil, not merged, of the provenencia Person type, with at least one accepted member (`Node.AcceptedMembers`). Its doc names it the package's one canonical-relative rule (S13-05's participants call it). `conclusionsummaries.nodesOf` checks only the first three, because a summary is shown for any unmerged Person, page or list; the accepted-member half is the page's R-2. `shownPerson` composes `isPerson(n, personTypeID)` *(new)*: not nil, not merged, Person type, and adds the accepted-member check. The person themself is checked by `isPerson` only: in `Families` (an id that fails is skipped) and in `PersonFamily` after `Handle` (a handle that fails, an Event or Place, is `ErrNotFound`), so a person with only provisional claims has a family read | R-2; working rule "Canonical only, filtered in Go" binds S13-01, S13-04, S13-05 (review M-5); S13-03 D-6 left the filter to this PR; a provisional-only page still opens today (`Handle`) | Inline checks — S13-05 would copy them; filtering after grouping — a dropped parent would still pair or hold siblings |
| D-8 | One batched walk. `func (d *deriver) loadSteps(ids [][]byte, event func([]byte) ([]kinship.Implied, error)) ([]stepped, error)` *(new)* loads, in four `Nodes` batches however many ids: the ids; their neighbors; the people one parent step from each; those people's neighbors. It returns, per id that is a node (an id that is no node is skipped), `stepped{node, edges, people}` *(new type)*: the node, its `d.parentEdges(n, event)`, and the distinct people one parent step away over those edges (`stepPeople(n.ID, edges, terms) [][]byte` *(new)*: first-seen, never `n.ID`, each through `kinship.ParentStep`). `forLinks` builds its people with `stepPeople` instead of its inline loop; `siblingsOf` takes its first four loads and each id's edges and parents (`kinship.Parents` over `edges`) from `loadSteps` instead of its own walk. `canonlinks.Reader.LoadRelationships(ids [][]byte, opts Options) error` *(new)* calls `loadSteps`, then derives and stores each id's `For` answer (`r.For(id)`, which reads only cached nodes now), and with `opts.Siblings` calls `LoadSiblings(ids)`, which derives and stores the siblings from cached nodes. So both `Load*` methods mean derive and store (review M-7). After it, `Relationships(id, opts)` for each id issues no query (AC-2.4) | Working rule "The family read is the family tree's first step", point 3: "a tree can load one generation in a fixed number of queries"; S13-03's Out row gave For's batching to this PR. One walk holds the batching, so a later load rule (a size bound, a skip) is made once (review M-1) | A second four-batch walk beside `siblingsOf`'s — two copies of one batching over the same edges; a loop of `Relationships` calls cold — one batch per birth-implied relative; loading in `persondetail` — it would copy walks `canonlinks` owns |
| D-9 | `persondetail.Families(c, ids)` per call, in this order: (1) `canonlinks.NewReader(g, matchregistry.Default().Kinship)`, `terms` from `canonlinks.LoadTerms(g)` and `personTypeID` from `g.Vocabulary()`'s `TypeIDsByKey()[vocabkey.Person]`, each once per call (both read the cached vocabulary, no query when warm); (2) `LoadRelationships(ids, {Siblings: true})`; (3) per distinct id (first occurrence order; duplicates give one `Family`; an id of other than 16 bytes is skipped) that is an unmerged Person by `g.Cached`, `Relationships(id, {Siblings: true})` and an unfiltered `KinOf`; (4) one `LoadRelationships(R, {})` where R is every relative those unfiltered roles name, all four roles, over all ids: this loads sibling and spouse nodes too (a sibling known only through a parent's birth is not the parent's neighbor and no earlier batch loads it) and the other parents of each child (one step from the child); (5) filter each person's links with D-7, reading nodes with `g.Cached` only (never `g.Node`, which would query per relative), and `KinOf` again; (6) per parent and child r, `Relationships(r, {})`, filtered, `KinOf`, into `kinOf[string(r)]`; (7) `kinship.FamilyUnits`. No memo slot holds the grouped family: every input is already stored (`forSlot`, `siblingSlot`, `kinshipSlot`, nodes), so a warm read recomputes linear work with no SQL | Working rule "Ids inside, summaries at the edge"; the warm budget is met without a new slot, which would need its own `Check` and drop rules | A `canonlinks.family` memo slot — a fourth slot and check for no query saved |
| D-10 | `persondetail.PersonFamily(c, entityID) (Family, conclusionsummaries.Summaries, error)`: `conclusiondetails.Handle` (→ `ErrNotFound`), then `isPerson` (an Event or Place handle → `ErrNotFound`, AC-1.12), then `Families(c, [][]byte{entityID})`, then one `conclusionsummaries.ByIDs` with every person the units name (`Want.Add(vocabkey.Person, id)`); then it drops each relative whose summary was not read (a node dropped mid-read: the next read is right, as `memberGraphs` skips) and then each unit that is left empty: a `UnitParents` or `UnitPartner` unit with no partner, and a `UnitOneParent` or `UnitSiblings` unit with no child. The returned family and summaries agree: every relative has a summary (review M-2). Types: `type Family struct{ Person []byte; Units []kinship.Unit }` | The page path is one guarded id; the summaries are the edge (working rules "One person summary for every tile", "Ids inside, summaries at the edge") | Summaries inside `Families` — a tree reads ids per generation and attaches summaries once |
| D-11 | Wire: `GetPersonFamily` (method 113) returns `repeated FamilyUnit units`; a unit has `FamilyUnitKind kind`, `repeated FamilyMember partners`, `repeated FamilyMember children`; a member has `PersonSummary person` and an optional `FamilyLink link`; a link has `bool recorded` (not only implied), `repeated string association_ids` (whatever `Link.Associations` holds, for a recorded link and an implied-only one alike: an implied link carries the associations of the links it was made from), `repeated KinshipShape implied` (`KinshipPattern pattern`, `string pivot_id`). The handler only maps: `PersonFamily` has already dropped what has no summary (D-10). Picture paths come from one `files.RelPaths(c, summaries.PictureImageFileIDs())`. `FamilyUnitKind` and `KinshipPattern` are proto enums; every pattern code has a value, `parent_of_parent` included (no nuclear link carries it today; a grandparent row will); a test loops over `kinship.Patterns()` and `PatternSharedParent` and fails on any code that maps to UNSPECIFIED, so a new pattern cannot vanish from the wire (review M-6, M-12) | One summary per member, as the map's scope says; enums give Swift a generated type and no parity row (`DerivedRowKind`, S13-06) | A string kind or pattern — a Swift parity row in `swift_vocabulary_test.go` for codes Swift reads; a separate `people` list — Swift would re-key; the map says "each as a `PersonSummary`" |
| D-12 | Swift key `.personFamily(project:, entityId:)`, placed after `.personClaims` in the `CatalogQueryKey` cases, in `Kind`, in the `kind` and `project` switches of `CatalogQueryKey.swift`, and in the descriptor switch, so its invalidation follows `.allCached(.personClaims)` in every list: `stalePolicy: .sessionFresh`, `invalidateOn: conclusionTriggers`, `whenHidden: .evict`, `invalidation: .allCached(.personFamily)`, `refreshes: [.applyHiddenPolicyOnAny(follows: .touchedHandles)]` | Working rule "One invalidation rule for person-section keys": a relative's change touches their id, not this person's, and the family's relatives are up to two hops away; the rule allows widening to any handle commit because the warm read is cheap (AC-2.1) | Following only the ids the cached value names — a new relative (a newly filed child) is not in the cached value |
| D-13 | Swift types in `GenealogyStore.swift`: `CatalogPersonFamily { units }`, `CatalogFamilyUnit { kind: Kind; partners; children }` with `enum Kind { parents, partner, oneParent, siblings }`, `CatalogFamilyMember { person: CatalogPersonSummary; link: CatalogFamilyLink? }`, `CatalogFamilyLink { recorded; associationIDs; implied: [CatalogKinshipShape] }`, `CatalogKinshipShape { pattern: Pattern; pivotID }` with `enum Pattern { marriageSpouses, birthParents, sharedParent, parentOfParent }`. `GoStore` maps `.unspecified, .UNRECOGNIZED` to nil and drops that unit or shape, as `mapDerivedRow` does; `mapPersonSummary` stays private and is reused | S13-01's `Catalog*` shape; one summary mapper | A Swift copy of the grouping — the working rule |
| D-14 | Tests at each layer: pure cases in `core/kinship/families_test.go` *(new)*; the load in `reader_test.go`; canonical filtering, ordering of a real read, query counts and commits in `core/database/persondetail/family_test.go` *(new)*; the wire in `api/ffi/handlers/person_detail_test.go` with a family built inside `catalogsession.Do` (as `core/catalogsession/warm_test.go` does); Swift in the two registry test files. The family builders live once in `graphcachetest.Canon` *(new methods)*: `Person()`, `Event(eventType)`, `Join(person, event, role)`, `Relationship(a, b, term)`, `Parent(parent, child)`, `Marriage(people...)`, `Birth(subject, father, mother)` (nil skips a role), `Merge(id, into)`, `Provisional(id)`, each returning the ids it made: the first six copy `world`'s bodies (`relationship(a, b, term)` files the term from `a`'s end, `person` key to `a`, `related_to` to `b`), and `Merge` and `Provisional` copy the raw updates in `core/database/persondetail/claims_test.go`; `canonlinks`' `world` methods `person`, `event`, `join`, `relationship`, `parent`, `marriage` call them and keep only their name bookkeeping | Each case at the lowest layer that sees it (S13-03 D-12); one set of family builders instead of three (review M-4) | A `familyWorld` copy in `persondetail` — a third copy of `world`'s builders; building families through RPCs — dozens of calls |

**Rules this plan bends.** None.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `Reader.LoadRelationships` *(new)* | func | `Reader.LoadSiblings`, `deriver.siblingsOf` | Both now sit on `loadSteps`; `LoadRelationships` also stores `For`'s answer and calls `LoadSiblings` for the sibling half (D-8) |
| `core/kinship/families.go` *(new)*: `Kin`, `KinOf`, `UnitKind` (`UnitParents`, `UnitPartner`, `UnitOneParent`, `UnitSiblings`), `Relative`, `Unit`, `FamilyUnits` | file, types, funcs | `siblings.go` `Siblings`, `Parents`; `ParentStep` | A grouping over links, not a pattern; built on `ParentStep`, `Edge.From` and `Keys` (D-1, D-2) |
| `core/database/persondetail/family.go` *(new)*: `Family`, `Families`, `PersonFamily` | file, type, funcs | `claims.go` `Claims` | The second section read in the package the working rule names |
| `GetPersonFamily`, `FamilyUnit`, `FamilyMember`, `FamilyLink`, `KinshipShape`, `FamilyUnitKind`, `KinshipPattern` *(new)* | FFI, proto | `ListPersonClaims*`, `DerivedRowKind`, `PersonSummary` | Reuses `PersonSummary` for every member (D-11) |
| handler `GetPersonFamily`, `familyMemberProto`, `kinshipPatternProto` *(new)* | funcs | `ListPersonClaims`, `personSummaryProto`, `fillHitSummaries` | Reuses `personSummaryProto` and the `RelPaths` call shape |
| `CatalogPersonFamily`, `CatalogFamilyUnit`, `CatalogFamilyMember`, `CatalogFamilyLink`, `CatalogKinshipShape` *(new)* | Swift types | `CatalogPersonClaimGroup`, `CatalogPersonSummary` | Members reuse `CatalogPersonSummary` (D-13) |
| `getPersonFamily` (protocol, `GoStore`, `FakeStore`), `seededPersonFamilies`, `CoreMethod.getPersonFamily`, `GoStore.mapFamilyUnit` *(new)* | Swift funcs, var | `listPersonClaims`, `seededPersonClaims`, `mapDerivedRow` | One per method, as S13-01 |
| `.personFamily` key, `Kind.personFamily`, descriptor *(new)* | key | `.personClaims`, `.personsList` | Lists' `applyHiddenPolicyOnAny` refresh (D-12) |
| `graphcachetest.Canon` `Person`, `Event`, `Join`, `Relationship`, `Parent`, `Marriage`, `Birth`, `Merge`, `Provisional` *(new)* | test helpers | `canonlinks` `world.person`, `world.event`, `world.join`, `world.relationship`, `world.parent`, `world.marriage`; the raw merge and provisional updates in `core/database/persondetail/claims_test.go` | One home for family builders; `world` delegates to them (D-14) |
| `nuclearRoles`, `Role`, `relationshipKeys`, `readsTerm` *(new)* | table, type, funcs | `eventPatterns`; the lookup in `ParentStep`, `chains`, `Siblings`, `eventLinks`; `ParentStep`'s `e.reads` | The package's table style; one key lookup and one reader of a relationship link (D-2) |
| `deriver.loadSteps`, `stepped`, `stepPeople` (package func), `isPerson`, `shownPerson` *(new)* | funcs, type | `siblingsOf`'s first half, `forLinks`' people loop; `nodesOf`'s checks | One walk for both loads (D-8); one canonical-relative rule (D-7) |
| `families_test.go` helper `linksFor(t, person, edges)` *(new)* | test helper | `world.relationships` | Builds the same `LinkSet` `Relationships` builds (recorded, `Derive`, `Siblings`) from pure edges |

Searched for an existing family grouping by name (`rg -n -i 'famil|spouses|parentset' core api --type go`: no grouping; `familyFixture`-style names only in tests) and by behavior (callers of `ParentStep`: `Derive`, `Parents`, `Siblings`, `forLinks`, `kinEdges`): none groups units.

**Builds on.** S11-51's `Relationships` and `LinkSet`; S13-03's `LoadSiblings`, shapes and `PatternSharedParent`; S13-01's `Handle`, `AcceptedMembers`, handler and key shape; S13-02 / S13-31's `PersonSummary` with gender terms; `conclusionsummaries.ByIDs`; `files.RelPaths`; `graphcachetest.Canon`; `applyHiddenPolicyOnAny`.
**Deletes.** `forLinks`' inline step-people loop (→ `stepPeople`); `siblingsOf`'s own four-load walk (→ `loadSteps`); the `LookupBridge` + `Keys` pairs in `ParentStep`, `chains`, `Siblings` and `eventLinks`, `Derive`'s relationship-bridge lookup and the `rel` parameters of `eventLinks` and `chains` (→ `relationshipKeys`); `ParentStep`'s body (→ `readsTerm`); `canonlinks` `world`'s builder bodies (→ `Canon` methods). The other raw `merged_into_id` updates in other packages' tests stay (not this PR's files). Nothing else becomes dead: `Relationships` gains its first non-test caller.
**Next adds.** S13-12 reads `.personFamily` and draws units; a tree calls `persondetail.Families` per generation with the previous generation's relatives (each call is a fixed number of batches). A half-sibling's other parent is `kinOf[string(sibling)][RoleParent]` with siblings added to D-9's second load. A grandparents row is one `nuclearRoles` entry (`TermGrandparent`) and a fifth `UnitKind`.

### Reach

| New path or state | Destination | Checked: resets on the new id · handles a missing referent · shows the state |
| --- | --- | --- |
| `Relationships` reached from a section read on any page open | `reader.go` `Relationships`, `For`, `LoadSiblings` | Their only caller was a test; the warm path issues no query (AC-2.1); a missing id returns nil (AC-1.12, `around`) |
| A relative merged, deleted or provisional meanwhile | `persondetail` filter (D-7), handler member drop (D-11) | Not shown (AC-1.9); a node dropped mid-read is skipped |
| A person with no relatives | `Families`, handler | Zero units, no error (AC-1.13) |
| `.personFamily` revalidated by any touched handle | `WorkspaceSession.revalidatedKeys` | Hidden keys evict, the visible one reloads once per `apply` (AC-4.3; S13-01's "revalidated once") |

### Scale

A family has no Scale line in the map; S13-03 measured at a large family (2 parents, 15 children, each with a birth) and a generation (three people). Cold `PersonFamily`: `Handle` (one node), `LoadRelationships` for the person (four batches, `LoadSiblings` then hits cache), `LoadRelationships` for every relative the person's roles name (four batches; siblings and spouses included, D-9), then `ByIDs`' fixed batches; AC-2.2 proves the count equal for a family of 1 and of 7 siblings and 7 children with births, AC-2.3 for one id and three. Warm: zero graph queries (AC-2.1); the handler adds one `files.RelPaths` query when any member leads with a picture, as `ListPersonClaims` does. Grouping is linear in the links except parent pairing (D-4), which checks each pair of the person's parents: parents are at most what the person's own links name, and a birth event pairs at most `MaxPivotEnd` (10) at each end (`eventLinks`), so the output is bounded by the input. Writes: none. Swift: every write that touches any handle reloads a visible Family section once (D-12); the Go read after a commit rebuilds from cached nodes and reloads only the committed nodes.

### Edges

- **Assumptions.** `LinkSet` keys a link by its ends and keys, so a relative filed twice in one role is one `Link` with both associations (AC-1.4). `spouse` and `sibling` are undirected in the seed and in `segmentcomparetest.Terms()`, so `Keys` returns one key twice and `KinOf` compares one key (Stop condition). `parent`'s inverse `child` folds to one key through `Key`, and `ParentStep` reads either filing (AC-1.1's recorded links are filed parent→child; a `child` filing is a T-1 row). Graph loads exclude merged neighbors (`load.go`, `merged_into_id IS NULL`), and D-7 filters again so the rule does not rest on the SQL. `Relationships` returns links in a stable order (node links by association ref, then `For`'s, then siblings), so unit order is stable across reads (AC-2.1's equality).
- **Neighbors.** Not in the family: a grandparent, cousin, guardian, or a grandparent through a chain (AC-1.11); a parent's spouse who is not the person's parent (not in `kin[RoleParent]`). In the family: a parent from a birth only (AC-1.2), a sibling through one parent (AC-1.5), a co-parent with no link (AC-1.7), a relative in two roles (AC-1.10), a recorded sibling who also shares parents (AC-1.6).
- **Removed inputs.** A relative not shown drops every link to them before grouping: a parent not shown pairs with no one and holds no sibling, so a sibling through only that parent moves to `Siblings`, and a child whose only other parent is not shown moves to `OneParent` (T-1 rows). A removed link drops `forSlot` and `siblingSlot` on commit; the next read loses it (AC-2.5).
- **Real data.** The seed has eleven `relationship_type` terms. `KinOf` reads four (`spouse`, `sibling`, `parent`, and `child`, which `Key` folds into `parent`) and puts the other seven (`cousin`, `grandparent`, `grandchild`, `pibling`, `nibling`, `guardian`, `ward`) in no role; T-1 runs one link of each of the eleven through `KinOf` with `segmentcomparetest.Terms()`.
- **Untrusted size.** Links per person: linear in what the loads already read. Parent pairs: quadratic in the person's distinct parents, each pair a map lookup; pairs that become sets are bounded by spouse links and birth events in the input (Scale). Siblings × sets: each sibling checks each set, both bounded by the person's links. No dimension multiplies past the input.

### Schema

None — no catalog change.

### Go core

| Package | Change |
| --- | --- |
| `core/database/canonlinks` | `reader.go`: `deriver.loadSteps`, `stepped`, `stepPeople` *(new)*; `siblingsOf` on `loadSteps`, `forLinks` on `stepPeople`; `Reader.LoadRelationships` *(new)* (D-8). `canonlinks.go`: package doc names `LoadRelationships` |
| `core/kinship` | `families.go` *(new)*: `Role`, `nuclearRoles`, `readsTerm`, `relationshipKeys`, `Kin`, `KinOf`, `UnitKind` and its four values, `Relative`, `Unit`, `FamilyUnits` (D-2 – D-6). Package doc: one sentence that `KinOf` and `FamilyUnits` read a person's nuclear family for the page |
| `core/database/persondetail` | `family.go` *(new)*: `Family`, `Families`, `PersonFamily` (D-7, D-9, D-10). Package doc unchanged |

```go
// LoadRelationships loads what Relationships(id, opts) reads for each id in a
// fixed number of batches, however many ids and relatives: the ids, their
// neighbors, the people one parent step from each, and those people's
// neighbors; with opts.Siblings, the siblings' loads. After it, Relationships
// for each id issues no query.
func (r *Reader) LoadRelationships(ids [][]byte, opts Options) error

// Role is a nuclear-family role of a relative, read from the person's end.
type Role int

const (
	RoleParent Role = iota + 1
	RoleChild
	RoleSpouse
	RoleSibling
)

// Kin is a person's links by role, each in link order.
type Kin map[Role][]Link

// KinOf reads person's links through nuclearRoles: each link is in every role
// whose key it carries from person's end. Any other link is in no role.
func KinOf(person []byte, links []Link, terms segmentcompare.Terms) Kin

type UnitKind int

const (
	UnitParents   UnitKind = iota + 1 // Partners: a set of person's parents; Children: siblings through each of them
	UnitPartner                       // Partners: one partner; Children: person's children with them
	UnitOneParent                     // Children: person's children with no other known parent
	UnitSiblings                      // Children: siblings with no shown shared parent
)

// Relative is one person in a unit and person's own link to them; Link is nil
// for a co-parent known only through a child.
type Relative struct {
	ID   []byte
	Link *Link
}

type Unit struct {
	Kind     UnitKind
	Partners []Relative
	Children []Relative
}

// FamilyUnits groups person's kin into units (D-3 – D-6). kinOf holds the kin of
// each of person's parents and children, keyed by string(id); a missing entry
// has none.
func FamilyUnits(person []byte, kin Kin, kinOf map[string]Kin) []Unit
```

```go
// Family is one person's family units.
type Family struct {
	Person []byte
	Units  []kinship.Unit
}

// Families reads the family of each id that is an unmerged Person, in ids
// order, from the graph cache: canonical relatives only, a fixed number of
// batched loads cold, no query warm.
func Families(c *database.Catalog, ids [][]byte) ([]Family, error)

// PersonFamily is the page read: one handle's family and the summary of every
// person its units name. An unknown or merged handle is ErrNotFound.
func PersonFamily(c *database.Catalog, entityID []byte) (Family, conclusionsummaries.Summaries, error)
```

Errors: `conclusiondetails.ErrNotFound` (`conclusiondetails.not_found`) from `PersonFamily`; otherwise the loads' errors.

### FFI

`api/proto/engine.proto`: `METHOD_GET_PERSON_FAMILY = 113` *(new)*. Handler `api/ffi/handlers/person_detail.go` `GetPersonFamily` *(new)*; dispatch `MethodGetPersonFamily` *(new)*. Regenerate with `./scripts/generate-proto.sh`.

```proto
// GetPersonFamilyRequest asks for a canonical person's nuclear family. An
// unknown or merged handle is conclusiondetails.not_found.
message GetPersonFamilyRequest {
  string project_dir = 1;
  string entity_id = 2;
}

// GetPersonFamilyResponse is the person's family units, parent sets first,
// then partners, then children with no other known parent, then siblings
// whose parents are not shown. Canonical people only.
message GetPersonFamilyResponse {
  repeated FamilyUnit units = 1;
}

// FamilyUnitKind is what a unit is to the person.
enum FamilyUnitKind {
  FAMILY_UNIT_KIND_UNSPECIFIED = 0;
  FAMILY_UNIT_KIND_PARENTS = 1;     // partners: a set of the person's parents; children: siblings through each
  FAMILY_UNIT_KIND_PARTNER = 2;     // partners: one partner; children: the person's children with them
  FAMILY_UNIT_KIND_ONE_PARENT = 3;  // children: the person's children with no other known parent
  FAMILY_UNIT_KIND_SIBLINGS = 4;    // children: siblings with no shown shared parent
}

message FamilyUnit {
  FamilyUnitKind kind = 1;
  repeated FamilyMember partners = 2;
  repeated FamilyMember children = 3;
}

// FamilyMember is one relative and the person's link to them; link is unset
// for a co-parent known only through a child.
message FamilyMember {
  PersonSummary person = 1;
  FamilyLink link = 2;
}

// FamilyLink says what a link rests on: recorded is true when a link is filed
// (not only implied); association_ids are the relationship and participation
// entities behind it; implied are the patterns that imply it, each through its
// pivot (the marriage or birth, or the shared parent).
message FamilyLink {
  bool recorded = 1;
  repeated string association_ids = 2;
  repeated KinshipShape implied = 3;
}

enum KinshipPattern {
  KINSHIP_PATTERN_UNSPECIFIED = 0;
  KINSHIP_PATTERN_MARRIAGE_SPOUSES = 1;
  KINSHIP_PATTERN_BIRTH_PARENTS = 2;
  KINSHIP_PATTERN_SHARED_PARENT = 3;
  KINSHIP_PATTERN_PARENT_OF_PARENT = 4;
}

message KinshipShape {
  KinshipPattern pattern = 1;
  string pivot_id = 2;
}
```

### macOS

None — no view. The session and store layer only:

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Wire types | `CatalogPersonFamily`, `CatalogFamilyUnit`, `CatalogFamilyMember`, `CatalogFamilyLink`, `CatalogKinshipShape` *(new)* | `macos/App/Platform/GenealogyStore.swift` | D-13 |
| Store | `getPersonFamily(projectDir:entityID:) async throws -> CatalogPersonFamily` *(new)* | `GenealogyStore.swift` protocol; `GoStore.swift`; `FakeStore.swift` (`seededPersonFamilies: [String: CatalogPersonFamily]`, unseeded → `conclusiondetails.not_found`, no `recordedCalls`) | As `listPersonClaims` |
| Method | `CoreMethod.getPersonFamily` *(new)* | `macos/App/Platform/CoreInvoke.swift` | |
| Key | `.personFamily(project:, entityId:)`, `Kind.personFamily` *(new)* | `CatalogQueryKey.swift`; descriptor in `CatalogQueryRegistry.swift` | D-12 |

### Strings

None — no copy.

## 7. Steps

**STEP-1 — Load a generation's relationships in batches**
Files: `core/database/canonlinks/reader.go`, `canonlinks.go`, `reader_test.go`, `core/database/graphcache/graphcachetest/graphcachetest.go`. Skill: `add-match-rule`.
Do: D-8. First move `siblingsOf`'s walk and `forLinks`' people loop onto `loadSteps` and `stepPeople` with no behavior change (`TestLoadSiblingsQueryCount` and `TestFor` unchanged), then add `LoadRelationships`. Also D-14's `Canon` builders in `core/database/graphcache/graphcachetest/graphcachetest.go`, with `world` delegating to them. Add T-2.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/canonlinks/ ./core/database/promotealign/` passes, existing tests unchanged.

**STEP-2 — Nuclear roles and family units**
Files: `core/kinship/families.go` *(new)*, `core/kinship/kinship.go` (package doc; `ParentStep` onto `readsTerm`; `chains` and `eventLinks` onto `relationshipKeys`, dropping their `rel` parameters and `Derive`'s relationship-bridge lookup), `core/kinship/siblings.go` (`Siblings` onto `relationshipKeys`), `core/kinship/families_test.go` *(new)*. Skill: `add-match-rule`; `.cursor/rules/go-tests.mdc`.
Do: D-2 – D-6. Add T-1.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/kinship/` passes.

**STEP-3 — The family read**
Files: `core/database/persondetail/family.go` *(new)*, `core/database/persondetail/family_test.go` *(new)*. Skill: `add-catalog-query`.
Do: D-7, D-9, D-10. Add T-3, T-4.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/persondetail/` passes; `go list -deps ./core/database/persondetail | grep api/ffi/handlers` prints nothing.

**STEP-4 — Wire messages**
Files: `api/proto/engine.proto`, generated `api/proto/engine/engine.pb.go`, `macos/App/Platform/Generated/engine.pb.swift`. Skill: `add-ffi-handler`.
Do: §6 FFI. Run `./scripts/generate-proto.sh`.
Check: `go build ./...` passes; the generated files hold `METHOD_GET_PERSON_FAMILY`.

**STEP-5 — The handler**
Files: `api/ffi/handlers/person_detail.go`, `api/ffi/dispatch.go`, `api/ffi/handlers/person_detail_test.go`. Skills: `add-ffi-handler`, `use-catalog-session`.
Do: D-11. Add T-5.
Check: `CGO_ENABLED=1 go test -tags fts5 ./api/ffi/...` passes.

**STEP-6 — Swift store**
Files: `macos/App/Platform/GenealogyStore.swift`, `GoStore.swift`, `FakeStore.swift`, `CoreInvoke.swift`, `macos/ProvenenciaTests/OnboardingModelTests.swift` (its `ThrowingStore` conformer gains `getPersonFamily`, throwing `StoreBoom.boom` as its `listPersonClaims` does). Skill: `add-swift-test`.
Do: D-13 and §6 macOS rows 1 – 3.
Check: `xcodebuild build` for the app scheme succeeds with no new warning.

**STEP-7 — The cache key**
Files: `macos/App/Features/Workspace/Session/CatalogQueryKey.swift`, `CatalogQueryRegistry.swift`, `macos/ProvenenciaTests/CatalogQueryRegistryTests.swift`, `macos/ProvenenciaTests/WorkspaceSessionTests.swift`. Skills: `add-workspace-place`, `add-swift-test`.
Do: D-12. Add T-6 – T-8 and update `invalidationsComeFromRegistrySpecs` (T-9).
Check: `xcodebuild test` passes the two suites.

**STEP-8 — Docs**
Files: `.agents/skills/add-match-rule/SKILL.md` (line 33: after "`LoadSiblings` loads them for many people", add "; `LoadRelationships` loads what `Relationships` reads for many people in one fixed set of batches"; and one sentence that `kinship.KinOf` and `kinship.FamilyUnits` group a person's nuclear family for the page and matching never calls them), `docs/ideas/kinship-view.md` (line 5: S13-04 shipped the family read as units over a list of ids), `docs/deployment-plan/spike-13/deployment-plan.review.md` (K-12: one sentence that `persondetail` imports `core/kinship` from S13-04 on, and the weight is still paid through `canonlinks`), `docs/deployment-plan/spike-13/completed.md`, `deployment-plan.md` (map row), this plan (status, Landed note).
Do: per the Definition of done.
Check: links resolve; `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-13/prs/S13-04-person-family-read.md` reports no errors.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/kinship/families_test.go` `TestKinOf`, `TestFamilies` *(new)* | `TestKinOf`: one link of each of the eleven seeded `relationship_type` terms (`parent` and `child` filings both; terms with no `vocabkey` constant as `vocabkey.BridgeTerm("cousin")` and so on) → `parent`/`child` in `kin[RoleParent]` or `kin[RoleChild]` by end, `spouse` in `kin[RoleSpouse]`, `sibling` in `kin[RoleSibling]`, the rest in none; a link to self in none. `TestFamilies`: one table row per AC-1.1, AC-1.2, AC-1.3 (both halves), AC-1.4, AC-1.5, AC-1.6, AC-1.7, AC-1.8, AC-1.10, AC-1.11 built with `linksFor`, plus: a parent not in kin (filtered) with a sibling only through them → `Siblings`; a child whose only other parent is filtered → `OneParent`; a child with two other parents → in both partner units; a birth with fathers F1, F2 and mother M → three parent sets (D-4); ten recorded parents with no pairing evidence → ten one-parent sets; empty kin → no units | AC-1.1, AC-1.2, AC-1.3, AC-1.4, AC-1.5, AC-1.6, AC-1.7, AC-1.8, AC-1.10, AC-1.11, D-2 – D-6 |
| T-2 | Go | `core/database/canonlinks/reader_test.go` `TestLoadRelationshipsQueryCount` *(new)* | AC-2.4's two parts, built as `TestLoadSiblingsQueryCount` builds them, with each child also a birth's subject and the person a birth's father | AC-2.4, D-8 |
| T-3 | Go | `core/database/persondetail/family_test.go` (package `persondetail_test`) `TestPersonFamily` *(new)* | AC-1.1, AC-1.2, AC-1.5, AC-1.7, AC-1.9 (provisional via `Canon.Provisional`; merged via `Canon.Merge`, then `Canon.Commit`); a relative whose summary is missing is dropped with its emptied unit (D-10), AC-1.12, AC-1.13, AC-1.14, AC-2.5; summaries returned for every member | AC-1.1, AC-1.2, AC-1.5, AC-1.7, AC-1.9, AC-1.12 – AC-1.14, AC-2.5, D-7, D-10 |
| T-4 | Go | `family_test.go` `TestPersonFamilyQueryCount` *(new)* | AC-2.1, AC-2.2 (a) vs (b), AC-2.3 one id vs three | AC-2.1 – AC-2.3, D-9 |
| T-5 | Go | `api/ffi/handlers/person_detail_test.go` `TestGetPersonFamily`, `TestKinshipPatternProto` *(new)* | AC-3.1 on a project from `onboarding.Complete`, the family written with `graphcachetest.NewCanon` builders inside `catalogsession.Do`; AC-3.2 for an unknown and a malformed id; `TestKinshipPatternProto`: every code in `kinship.Patterns()` plus `PatternSharedParent` maps to a non-UNSPECIFIED value, with no exception | AC-3.1, AC-3.2, D-11 |
| T-6 | Swift | `macos/ProvenenciaTests/CatalogQueryRegistryTests.swift` `personFamilyInvalidateOnPromote` *(new)* | `.promotedSubject` → contains `.allCached(.personFamily)`; key kind, project, `.evict`, `.sessionFresh`; `Kind.personFamily.descriptor.refreshes.map(\.follows) == [.touchedHandles]` | AC-4.2, AC-4.3, D-12 |
| T-7 | Swift | `CatalogQueryRegistryTests.swift` `personFamilyLoadsThroughTheStore` *(new)* | AC-4.1 with one seeded unit | AC-4.1, D-13 |
| T-8 | Swift | `macos/ProvenenciaTests/WorkspaceSessionTests.swift` `personFamilyFollowsEveryTouchedHandle` *(new)* | A warmed key for e1; `revalidatedKeys(for: CatalogWriteEffects(handles: ["e7"]))` names it; `CatalogWriteEffects.empty` names nothing | AC-4.3 |
| T-9 | Swift | `CatalogQueryRegistryTests.swift` `invalidationsComeFromRegistrySpecs` (changed) | `.deletedSource` and `.promotedSubject` lists end with `.allCached(.personFamily)`; the loop asserts `.allCached(.personFamily)` exactly on Conclusion triggers | AC-4.2 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `core/database/canonlinks` (STEP-1's extraction), `TestSiblingsLeaveMatchingAlone`, `promotealign` goldens, and `invalidationsComeFromRegistrySpecs`.

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left (none expected: no copy)
- [ ] Sweep: each renamed or deleted name `rg`'d by its old, qualified name across `.agents/`, `.cursor/rules/`, `docs/*.md` and `L10n.swift` with no hit left; `scripts/check-guidance-names.py` passes (check-pr runs it)
- [ ] Every row of **Rules this plan bends** has the user's answer, and its rule is updated (none)
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New (no UI)
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together) — or none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated (if gated) — not gated
- [ ] Domain doc updated if a rule in docs/*.md changed: `docs/ideas/kinship-view.md` line 5; `add-match-rule` line 33 (STEP-8)
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

No view reads the family until S13-12; the reviewer checks the read and that matching did not move:

1. `CGO_ENABLED=1 go test -tags fts5 ./core/kinship/ ./core/database/canonlinks/ ./core/database/persondetail/ ./api/ffi/handlers/ -run 'Famil|KinOf|LoadRelationships' -v` and read T-1's and T-3's case names against §2.
2. `CGO_ENABLED=1 go test -tags fts5 ./core/database/promotealign/ ./core/graphalign/...` passes with no golden changed (`git diff --stat` shows no `testdata` change).
3. Open the app on the product owner's project and a Person page with two marriages: the page looks and behaves as before (nothing calls `getPersonFamily` yet).

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| `LoadRelationships` misses a node `For` reads, so a warm read issues a query | AC-2.4 and AC-2.1 count queries with births on both sides |
| `ByIDs` issues a query on a warm read | AC-2.1 counts the whole `PersonFamily`, summaries included |
| Unit order drifts between reads | AC-2.1 compares two reads for equality |
| Parent pairing quadratic on malformed data | Bounded by the input (Scale, Edges); T-1 has a ten-parent row |

**Stop and report if:**

- `segmentcompare.Keys(rel, vocabkey.TermSpouse, "", terms)` or the same for `TermSibling` returns two different keys with the seeded terms (D-2 assumes undirected).
- A second `PersonFamily` read issues a graph query after `LoadRelationships` (AC-2.1), and the query comes from code outside `canonlinks` and `conclusionsummaries`.
- `graphcachetest.NewCanon` fails inside `catalogsession.Do` on an onboarded project (T-5, D-14).
- Extracting `stepPeople` changes any existing `canonlinks` or `promotealign` test result (STEP-1).
- A relative's `Relationships` link order differs between two reads with no commit between them.
- `go list -deps ./core/database/persondetail` lists a handler package.
- `protoc`, `protoc-gen-go` or `protoc-gen-swift` is missing (STEP-4), or `xcodebuild` is not available (STEP-6, STEP-7, check-pr): the Swift and proto steps run on macOS; report and do not hand-edit generated files.

## 12. Open questions

None.

## 13. PR

**Title:** S13-04 — Person family read
**Body:** objective in one sentence; what shipped against US-1 – US-4; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-13/prs/S13-04-person-family-read.md`.
