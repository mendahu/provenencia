# S11-23 — One home for primary kinds

**Status:** Blocked — on S11-22
**Spike / slice:** Spike 11 · Slice 1 — Graspable model and pure foundations
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no new chrome: Go constants and the Swift fake)
**Board:** —
**Depends on:** S11-22 (in progress; this plan is written against [its plan](S11-22-pure-value-types.md), not its code)
**Blocks:** S11-02
**Branch:** `s11-23-primary-kinds`
**Version bump:** none — internal refactor, no behavior change ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-seeded-vocabulary`](../../../../.agents/skills/add-seeded-vocabulary/SKILL.md), [`add-catalog-query`](../../../../.agents/skills/add-catalog-query/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md)
**Rules that bind:** `.cursor/rules/go-tests.mdc`
**Context:** working rules "Pure value types and kinds come first", "Pure means the package builds and tests with no build tags and no `core/database` import", "High tolerance for churn", "One wire vocabulary"; [`docs/seeded-vocabulary.md` §1.1](../../../seeded-vocabulary.md#11-vocabulary-origin-namespace); [`ideas/shared-list-helpers.md`](../../../ideas/shared-list-helpers.md) (its "Primary Subject kinds" note moved into the map's S11-23 scope entry)
**Planned:** 2026-10-17 against `origin/main` @ `233e87a`, plus [S11-22's plan](S11-22-pure-value-types.md) as landed code
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

The primary Subject kinds (`person`, `event`, `place`), the seeded Property keys code reads by name, and the vocabulary origins (`provenencia`, `user`, `plugin:<id>`) get one pure package, `core/vocabkey`. Today the kind list is written out seven times in Go and five times in the Swift fake, and the origin constants are declared in ten packages with seven copies of the same origin check. S11-02's match registry keys on these names; without one home it would add the next copy.

**Why now.** The working rule "Pure value types and kinds come first": S11-02 builds `core/matchregistry` with keys by kind and Property, and it must import them from a pure package.

**What stays.** Behavior is unchanged. The seed (`core/database/subjectvocab/registry.go`, `core/connectrules`) keeps its literal keys; a test ties it to the new constants. Swift's `EvidencePrimaryKind` stays where it is. The FFI wire format is untouched.

## 2. User stories

**US-1.** As a contributor building a matching layer, I import the primary kinds, the Property keys and the origins from one package with no `core/database` import.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | `go list -deps ./core/vocabkey \| grep 'provenencia/core/database'` | No output |
| AC-1.2 | `go test ./core/vocabkey/` with no build tags | Passes |
| AC-1.3 | `vocabkey.IsPrimaryType("person", "provenencia")`, `("person", "user")`, `("participation", "provenencia")` | `true`, `false`, `false` |
| AC-1.4 | `vocabkey.ValidOrigin` on `provenencia`, `user`, `plugin:dna`, `plugin:`, `Provenencia` | `true`, `true`, `true`, `false`, `false` |

**US-2.** As a contributor, I find no second copy to keep in step.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | Search S-1 below | Exactly one hit: `macos/App/Features/Workspace/OmnibarHitPresentation.swift` `case "person", "event", "place":`, a switch arm that stays (D-7) |
| AC-2.2 | Search S-2 below | Hits only in `core/vocabkey/vocabkey.go` (new): the three origin constants |
| AC-2.3 | Search S-3 below | Hits only in `core/vocabkey/vocabkey.go` (new) |
| AC-2.4 | A seed type is added with role `root`, or a constant in `vocabkey` names a key the seed lacks | `TestVocabKeysMatchSeed` fails |
| AC-2.5 | A case is added to or removed from Swift `EvidencePrimaryKind` | `TestSwiftVocabularyMatchesGo/primary_kinds_match_Go` fails |

The searches, copied exactly (bare `|` alternation; `\|\|` in S-1 is a literal `||`):

```sh
# S-1: hand-written kind lists (D-7). Non-test Go and Swift app code.
rg -n -g '!*_test.go' -e '"person", "event", "place"' -e '"person": *true' -e '== "person" \|\|' -e 'KindPerson \|\|' -e 'KindPerson: +true' core api macos/App
# S-2: origin declarations, test files included.
rg -n -i 'origin(provenencia|user|pluginpref)\s+=' --type go core api
# S-3: origin literals in non-test Go.
rg -n -e '"provenencia"' -e '"plugin:"' --type go -g '!*_test.go' core api
```

**US-3.** As a researcher, nothing behaves differently: Promote, the Sources list progress, search and delete Impact treat the same Subjects as before.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | The whole Go and Xcode suite | `scripts/check-pr.sh` prints `check-pr: PASS`; no test changes except import paths, qualifiers and the three loops D-7 names |

## 3. Scope

| In | Covered by |
| --- | --- |
| `core/vocabkey` (new): origin constants and checks, primary kinds and their checks, eight Property key constants | STEP-1, AC-1.1 – AC-1.4 |
| Every `OriginProvenencia` / `OriginUser` declaration outside `vocabkey` deleted; callers, `originOK` copies and `"plugin:"` checks use `vocabkey` | STEP-2, AC-2.2, AC-2.3 |
| Every hand-written primary kind list (D-7) uses `vocabkey`; `promote.PrimaryKind`, `match.ProfiledKinds`, `promotePrimary`, `searchindex.handleKinds`, `search.isHandleKind` deleted; `sourcegraphprogress` drops its `promote` import | STEP-3, AC-2.1 |
| `neighborTypeProperty` uses the Property key constants; a seed test ties `vocabkey` to the `subjectvocab` seed | STEP-4, AC-2.4 |
| `FakeStore` reads its kind set from `EvidencePrimaryKind`; a Go test ties that enum to `vocabkey` | STEP-5, AC-2.5 |
| Skills and domain docs name `core/vocabkey` | STEP-7 |

| Out | Home |
| --- | --- |
| Typed keys (`SubjectType`, `PropertyKey`) | S11-02 builds them in `core/matchregistry` from these constants |
| Single spellings of a kind or Property key that are not lists: seed data (`core/database/subjectvocab/registry.go`, `connectrules.go`, `hops.go`), per-kind `switch` arms (`api/ffi/handlers/conclusion_detail.go`, `api/ffi/handlers/promote_align.go`, `core/search/fts.go`, `conclusionheaders`, `FakeStore` switches, `OmnibarHitPresentation.swift`), and `conclusionheaders` / `sourceeventtitles` Property reads | not planned: each layer PR from S11-02 spells new code with `vocabkey`; `core/match`, `core/graphalign/registry.go` and `promotealign` are rewritten or deleted in S11-13 and S11-16 |
| Constants for other seeded keys (`likeness`, `role`, `relationship_type` and the rest of the seed) | not planned: a key gets a constant when code first reads it by name (D-8) |
| Swift origin literals (Swift's home is `macos/App/Features/Catalog/OriginMarkers.swift`) | not planned |
| Renaming or moving `EvidencePrimaryKind`, or adding `EvidenceBridgeKind` to the guard | not planned |
| `subjectvocab.Disambiguation*` re-exports of `connectrules` | not planned |
| Origin-filtering changes (who may create what) | not planned: each package's rule is unchanged (D-4) |

## 4. Current state

Rows marked **(after S11-22)** describe the code S11-22's plan leaves; the rest are true at `233e87a` and S11-22 does not touch them.

| Area | Where | Today |
| --- | --- | --- |
| Promote's kind list | `core/database/promote/promote.go` `primaryKinds` (L33), `PrimaryKind(key, origin)` (L36) | Callers: `promote.go` L98, `unpromote.go` L44, `batch.go` L178, `core/database/sourcegraphprogress/sourcegraphprogress.go` L60, `core/database/promotealign/layer.go` L36 |
| Progress counts | `core/database/sourcegraphprogress/sourcegraphprogress.go` `Progress.add` | Imports `core/database/promote` only for `PrimaryKind` |
| Matching's kind list | `core/match/registry.go` `ProfiledKinds()` (L258) | One caller: `core/database/promotealign/stats.go` `profilePropertyKeys` (L259) |
| Promote alignment lists | `core/database/promotealign/layer.go` `loadMetas` L93 `[]string{"person", "event", "place"}`; `neighborTypeProperty` L117 `{"event": "event_type"}`; `core/database/promotealign/canon.go` `promotePrimary` (L197) with callers L55, L165, both passing the literal `"provenencia"` | Three more copies of the list |
| Search lists | `core/database/searchindex/handles.go` `KindPerson` / `KindEvent` / `KindPlace` (L22–24), `handleKinds` (L28) read at L94 (with `typeOrigin == "provenencia"`) and L119, switches at L211, L266; `core/search/search.go` `KindPerson` / `KindEvent` / `KindPlace` (L18–20); `core/search/fts.go` `isHandleKind` (L243), one caller L252 | Two more copies; `searchindex.Kind*` are used only inside `handles.go` |
| Origin constants | `OriginProvenencia` declared in `core/database/properties/properties.go` L37, `core/database/subjecttypes/subjecttypes.go` L26, `core/database/propertyterms/propertyterms.go` L33, `core/database/sourcetypes/sourcetypes.go` L38, `core/database/metadatafields/metadatafields.go` L38, `core/database/nameformats/nameformats.go` L19, `core/database/claimconfidencegrades/grades.go` L19, `core/database/sourcecredibilitygrades/grades.go` L17, `core/connectrules/connectrules.go` L21, `core/match/registry.go` L68; `OriginUser` beside six of them; `core/database/deleteimpact/kinds.go` L33–35 `originProvenencia`, `originUser`, `originPluginPref` | About 330 qualified uses (`<pkg>.OriginProvenencia` / `OriginUser`) in 63 files (11 non-test, 52 test); 82 files name an origin constant, including `deleteimpact` |
| Origin checks | `originOK` in `properties.go` L463, `propertyterms.go` L382, `sourcetypes.go` L428, `metadatafields.go` L412, `core/database/claimconfidencegrades/grades.go` L140, `core/database/sourcecredibilitygrades/grades.go` L138: provenencia, user, or `plugin:<id>`. `subjecttypes.go` L214: provenencia or `plugin:<id>` (no user). `core/database/deleteimpact/impact.go` L134: the plugin test inline | Seven copies of one prefix test |
| Raw origin literals | `"provenencia"` in non-test Go: `core/match/profiles.go` L66, `core/graphalign/registry.go` L26, L49, L52, L55, `api/ffi/handlers/source_read.go` L160, `core/database/propertyterms/propertyterms.go` L372, `core/database/searchindex/handles.go` L94, `core/database/sourcefill/load.go` L178, `core/database/sourceeventtitles/sourceeventtitles.go` L120, `core/database/matching/matching.go` L154, `canon.go` L55, L165, L198 | Origin spelled by hand |
| Seed | `core/database/subjectvocab/registry.go` `seedTypes` (person, event, place are `RoleRoot`; bridges `RoleBridge`; source `RoleReification`), `seedProperties`, `seedBindings` | Declared data with literal keys. **(after S11-22)** value types and cardinality read `propertymeta` |
| Seed tests | `core/database/subjectvocab/profile_test.go` `TestDefaultProfileCardinality` L21 loops `[]string{"person", "event", "place"}`; `core/match/registry_test.go` L77 and `core/database/autoreconciler/autoreconciler_test.go` L82 loop the same list | Three test copies |
| Swift kinds | `macos/App/Features/Workspace/Session/SourceGraphSnapshot.swift` `EvidencePrimaryKind` (`person`, `event`, `place`; `CaseIterable`) | Swift's typed home, used by the Evidence graph and Promote |
| Swift fake | `macos/App/Platform/FakeStore.swift` `searchCatalog` local `handleKinds` L1015; `promoteSubject` guard L1346; alignment proposal L1457; apply L1549; `primaryTypeKeys` L3007 | Five literal sets of the same three keys |
| Swift guard | `api/ffi/swift_vocabulary_test.go` `TestSwiftVocabularyMatchesGo` with `swiftEnumBody` and the constants `swiftVocabulary`, `swiftValueKinds`, `swiftMarks` | Reads Swift enums and fails on drift; does not read `EvidencePrimaryKind` |
| Properties package | `core/database/properties/properties.go` const block L37–38 (Origin); **(after S11-22)** the value-type and cardinality constants are gone to `core/propertymeta`; `originOK` unchanged | — |
| Skills and docs | `.agents/skills/add-seeded-vocabulary/SKILL.md` (rules list L42 on, "Do not" list L139); `docs/seeded-vocabulary.md` §1.1 "Reserved `origin` values" (L49–57) | Name no Go home for origins |

**Precedent.** S11-22 ([plan](S11-22-pure-value-types.md)) has the same shape: a pure `core/<singular>` package, one commit per concern that repoints every caller, nothing kept for compatibility. Copy its commit rhythm and its D-6 style of "done means this search has no hits". The Swift guard copies `TestSwiftVocabularyMatchesGo`'s "seeded property keys are seeded" subtest: read a Swift enum with `swiftEnumBody`, compare with Go.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | New package `core/vocabkey`, one file `vocabkey.go`, importing only `strings`. Package doc: "Product vocabulary keys code reads by name: vocabulary origins, the primary Subject kinds, and the seeded Property keys on them. Pure: the seed and the queries live under `core/database`." | Singular `core/<noun>` like S11-22's `datevalue` / `propertymeta`; no package or common local variable is named `vocabkey` (`rg -n '\bvocabkey\b'` is empty) | `core/database/subjecttypes` (the idea file's suggestion) — because it imports `core/database`; `core/connectrules` — because it is the connect matrix, not a key list; `core/vocab` — because `vocab` is a local variable in 24 places |
| D-2 | Untyped string constants, exactly as today's values | Every caller compares against `string` fields read from SQL; S11-22 D-5 made the same call | Typed `SubjectType` / `PropertyKey` — because typed keys are S11-02's registry decision |
| D-3 | Kinds: `Person = "person"`, `Event = "event"`, `Place = "place"`; `func PrimaryKinds() []string` returns a new slice `{Person, Event, Place}` on each call; `func IsPrimaryKind(kind string) bool` (for strings that already mean a product kind: search hit kinds); `func IsPrimaryType(typeKey, origin string) bool` = `origin == OriginProvenencia && IsPrimaryKind(typeKey)` (for catalog rows) | A function, not an exported `var`, so no caller can append to the shared list; the order is today's order in every loop; two predicates because search hit kinds carry no origin | An exported `map[string]bool` — because it is mutable and unordered |
| D-4 | Origins: `OriginProvenencia = "provenencia"`, `OriginUser = "user"`, `OriginPluginPrefix = "plugin:"`; `func IsPluginOrigin(origin string) bool` (prefix and a non-empty id, exactly today's test); `func ValidOrigin(origin string) bool` (provenencia, user, or plugin). The six `originOK` that accept all three are deleted and their callers call `vocabkey.ValidOrigin`. `subjecttypes.originOK` stays, because it refuses `user`, and its body becomes `origin == vocabkey.OriginProvenencia \|\| vocabkey.IsPluginOrigin(origin)`. `deleteimpact` uses `vocabkey.IsPluginOrigin` and `vocabkey.OriginProvenencia`; its three unexported origin constants are deleted | The map gives origins one pure home; the check is part of the concept and is copied seven times | Leaving `originOK` copies — because they would spell `"plugin:"` seven more times |
| D-5 | Every exported `OriginProvenencia` / `OriginUser` outside `vocabkey` is deleted, with no alias, and every use in non-test and test Go reads `vocabkey`. Every non-test `"provenencia"` literal used as an origin (§4 Raw origin literals) reads `vocabkey.OriginProvenencia`. Raw `"provenencia"` / `"user"` strings in `_test.go` files stay: they are test data | The churn rule: no deprecated aliases | Keeping per-package constants as `= vocabkey.OriginProvenencia` — because they are aliases |
| D-6 | Property keys: `PropertyName = "name"`, `PropertySexAtBirth = "sex_at_birth"`, `PropertyGender = "gender"`, `PropertyEventType = "event_type"`, `PropertyDate = "date"`, `PropertyStartDate = "start_date"`, `PropertyEndDate = "end_date"`, `PropertyToponym = "toponym"`. Used in this PR by `core/database/promotealign/layer.go` `neighborTypeProperty` and by `TestVocabKeysMatchSeed`; S11-02's registry keys on them | These are the keys matching reads by name: the locked bindings of the three kinds that matching compares, plus `sex_at_birth` and `gender` (working rule "Sparse data first": person name and sex are core). They match Swift's `SeededPropertyKey` | Every locked key including `likeness` — because no Go code reads `likeness` by name (D-8) |
| D-7 | A **hand-written kind list** is a slice, set or map literal, or an `\|\|` chain, that names all three primary kinds. Every one in non-test Go and Swift app code is replaced: `promote.primaryKinds` + `PrimaryKind` (deleted; five callers use `vocabkey.IsPrimaryType`), `match.ProfiledKinds` (deleted; `stats.go` uses `vocabkey.PrimaryKinds()`), `layer.go` L93, `canon.go` `promotePrimary` (deleted; callers use `vocabkey.IsPrimaryKind(kind)`, since the literal origin there is always provenencia), `searchindex.handleKinds` and `searchindex.KindPerson` / `KindEvent` / `KindPlace` (deleted; `searchindex.KindSource` in `searchindex.go` stays; L94 uses `vocabkey.IsPrimaryType`, L119 loops `vocabkey.PrimaryKinds()`, the switches use `vocabkey.Person` / `Event` / `Place`), `search.isHandleKind` (deleted; L252 uses `vocabkey.IsPrimaryKind`). Test loops over exactly the three kinds use `vocabkey.PrimaryKinds()`: `profile_test.go` L21, `core/match/registry_test.go` L77, `autoreconciler_test.go` L82. Switch case lists stay: `core/search/fts.go` L432 `case KindPerson, KindEvent, KindPlace:`, the `handle(...)` calls in `core/search/registry.go` L139–141 (per-kind priorities, D-9), and `OmnibarHitPresentation.swift` L103. Test data lists that name other keys too (`subjectvocab_test.go` L523, `matching_test.go` L78, `autoreconciler_test.go` L1222, `EvidenceCitedPropertyMarksTests.swift` L118) stay. The check is search S-1 (§2) | The map's Check: "no hand-written kind list remains (grep)" | Replacing every single spelling — because it reaches seed data and dispatch switches in about 60 sites, most in code S11-13 and S11-16 rewrite |
| D-8 | A key gets a `vocabkey` constant when code reads it by name. `vocabkey` has no per-kind key lists and no `Locked` flag: the binding and its lock stay in the seed | Keeps `vocabkey` a list of names, not a second seed | A `LockedKeys(kind)` function — because nothing reads it |
| D-9 | `search.KindPerson`, `KindEvent`, `KindPlace` stay in `core/search/search.go` as hit-kind names, declared `= vocabkey.Person` / `Event` / `Place` | A handle hit's kind is its Subject type key by design (`core/database/searchindex/handles.go` sets `h.kind = typeKey`); search keeps its own hit-kind vocabulary beside `KindSource` | Deleting them — because the hit-kind list (`KindSource`, `KindSourceType`, `KindMetadataField`) would lose three members |
| D-10 | Swift: `EvidencePrimaryKind` stays in `SourceGraphSnapshot.swift`, unchanged. `FakeStore.primaryTypeKeys` becomes `Set(EvidencePrimaryKind.allCases.map(\.rawValue))` and the other four sites (L1015, L1346, L1457, L1549) read `Self.primaryTypeKeys`. `swift_vocabulary_test.go` gets `swiftGraphKinds = "../../macos/App/Features/Workspace/Session/SourceGraphSnapshot.swift"` and a subtest `primary kinds match Go` that reads the `case` names of `EvidencePrimaryKind` and compares them, as a set, with `vocabkey.PrimaryKinds()` | Swift already has a typed home; the guard already reads Swift files outside `ConclusionVocabulary.swift` (`swiftValueKinds`, `swiftMarks`) | Moving the enum into `ConclusionVocabulary.swift` — because it splits it from `EvidenceBridgeKind` for no check the test can't already do |
| D-11 | `TestVocabKeysMatchSeed` *(new)* lives in `core/database/subjectvocab/vocabkey_test.go`, internal `package subjectvocab`, reading `seedTypes` and `seedProperties` directly (no catalog): the set of `seedTypes` keys with `Role == RoleRoot` equals `vocabkey.PrimaryKinds()`; each of the eight Property constants is a `seedProperties` key | The seed is the source of truth for what exists; this ties the names to it in both directions for kinds. `TestDefaultProfileCardinality` reads the seed the same way | Generating `vocabkey` from the seed — because the seed package is impure |
| D-12 | Commits, in order, each passing the full suite: (1) `core/vocabkey` with its tests; (2) origins; (3) Go kind lists and Property keys with the seed test; (4) Swift fake and its guard; (5) docs | One concern per commit; each leaves the tree building | One commit — because a reviewer reads 80 files of origin churn mixed with the list changes |

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `core/vocabkey` *(new)* | package | `core/connectrules` (pure, declares `OriginProvenencia`); `core/database/subjecttypes` | `connectrules` is the connect matrix (D-1); `subjecttypes` imports `core/database` |
| `vocabkey.PrimaryKinds`, `IsPrimaryKind`, `IsPrimaryType` *(new)* | func | `promote.PrimaryKind`, `match.ProfiledKinds`, `promotePrimary`, `searchindex.handleKinds`, `search.isHandleKind` | Those five are deleted and their callers move here (D-7) |
| `vocabkey.ValidOrigin`, `IsPluginOrigin` *(new)* | func | the seven `originOK` copies; `deleteimpact` inline prefix test | Six copies are deleted; `subjecttypes.originOK` keeps its stricter rule on top (D-4) |
| `vocabkey.Property*` ×8 *(new)* | const | Swift `SeededPropertyKey` (`ConclusionVocabulary.swift`); `subjectvocab.seedProperties` | Swift can't be imported by Go; the seed is impure and is data, not names (D-8) |
| `TestVocabKeysMatchSeed` *(new)* | test | `core/database/subjectvocab/profile_test.go` `TestDefaultProfileCardinality` | That one checks matching Features; this one checks the names (D-11) |
| `primary kinds match Go` subtest *(new)* | test | `TestSwiftVocabularyMatchesGo` subtests | Added as one more subtest in the same function (D-10) |
| `swiftGraphKinds` *(new)* | test const | `swiftValueKinds`, `swiftMarks` | Same pattern |

**Builds on.** S11-22's pure-package convention; `swiftEnumBody`; the seed tables in `core/database/subjectvocab/registry.go`; [`docs/seeded-vocabulary.md` §1.1](../../../seeded-vocabulary.md#11-vocabulary-origin-namespace).
**Deletes.** Ten exported `OriginProvenencia` and six `OriginUser` declarations; three unexported origin constants in `deleteimpact`; six `originOK` functions; `promote.primaryKinds`, `promote.PrimaryKind`; `match.ProfiledKinds`; `promotealign.promotePrimary`; `searchindex.KindPerson` / `KindEvent` / `KindPlace` and `handleKinds`; `search.isHandleKind`; five literal kind sets in `FakeStore.swift`; the `promote` import in `sourcegraphprogress`.
**Next adds.** S11-02's `matchregistry` keys (`SubjectType`, `PropertyKey`) are built from `vocabkey` constants and its purity guard lists `core/vocabkey`. A new primary kind is one constant and one `PrimaryKinds` entry, and the two guards then point at the seed and Swift.

### Schema

None — no catalog change.

### Go core

```go
package vocabkey

const (
	OriginProvenencia  = "provenencia"
	OriginUser         = "user"
	OriginPluginPrefix = "plugin:"
)

const (
	Person = "person"
	Event  = "event"
	Place  = "place"
)

const (
	PropertyName       = "name"
	PropertySexAtBirth = "sex_at_birth"
	PropertyGender     = "gender"
	PropertyEventType  = "event_type"
	PropertyDate       = "date"
	PropertyStartDate  = "start_date"
	PropertyEndDate    = "end_date"
	PropertyToponym    = "toponym"
)

func PrimaryKinds() []string
func IsPrimaryKind(kind string) bool
func IsPrimaryType(typeKey, origin string) bool
func IsPluginOrigin(origin string) bool
func ValidOrigin(origin string) bool
```

| Package | Change |
| --- | --- |
| `core/vocabkey` *(new)* | As above |
| `core/database/{properties,propertyterms,sourcetypes,metadatafields,claimconfidencegrades,sourcecredibilitygrades}` | Origin constants and `originOK` deleted; validation calls `vocabkey.ValidOrigin`; internal uses read `vocabkey` |
| `core/database/subjecttypes`, `core/database/nameformats`, `core/connectrules`, `core/match` | Origin constants deleted; `subjecttypes.originOK` per D-4 |
| `core/database/deleteimpact` | `originProvenencia`, `originUser`, `originPluginPref` deleted; `impact.go` L134 and L137 read `vocabkey` |
| `core/database/promote` | `primaryKinds`, `PrimaryKind` deleted; `Save`, `unpromote.go`, `batch.go` call `vocabkey.IsPrimaryType` |
| `core/database/sourcegraphprogress` | `vocabkey.IsPrimaryType`; `promote` import removed |
| `core/database/promotealign` | `layer.go` L36 `vocabkey.IsPrimaryType`, L93 `vocabkey.PrimaryKinds()`, L117 `{vocabkey.Event: vocabkey.PropertyEventType}`; `stats.go` `vocabkey.PrimaryKinds()`; `canon.go` `promotePrimary` deleted |
| `core/match` | `ProfiledKinds`, `OriginProvenencia` deleted; `profiles.go` `product` uses `vocabkey.OriginProvenencia` |
| `core/database/searchindex`, `core/search` | Per D-7 and D-9 |
| Raw literal sites in §4 | `vocabkey.OriginProvenencia` |

Errors: unchanged; no new codes.

### FFI

None — `api/ffi/handlers/conclusion_detail.go`, `source_read.go` and the other handlers change only origin qualifiers; no proto change.

### macOS

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Primary kind set in the fake | `EvidencePrimaryKind` | `macos/App/Platform/FakeStore.swift` | D-10; no view changes |

States: none — no UI change. Accessibility: none.

### Strings

None — no copy.

## 7. Steps

**STEP-1 — Create `core/vocabkey`**
Files: `core/vocabkey/vocabkey.go` (new), `core/vocabkey/vocabkey_test.go` (new).
Do: the constants and five functions in §6 (D-1 – D-4, D-6); tests T-1 – T-3.
Check: `go test ./core/vocabkey/` passes with no build tags; `go list -deps ./core/vocabkey | grep 'provenencia/'` prints only `github.com/mendahu/provenencia/core/vocabkey`. Commit (D-12).

**STEP-2 — Move the origins**
Files: every file from `rg -l '\bOrigin(Provenencia|User)\b|originProvenencia|originUser|originPluginPref' --type go core api` (82 at the planning base, about 330 qualified uses), and the raw-literal files in §4. Skill: `add-seeded-vocabulary`.
Do: D-4, D-5. Delete each declaration; replace each `<pkg>.OriginProvenencia` / `<pkg>.OriginUser` and each unqualified use with the same name in `vocabkey`; delete the six `originOK`; rewrite `subjecttypes.originOK`; replace the non-test `"provenencia"` origin literals and the `"plugin:"` checks. Remove imports left unused (a test file that imported a vocabulary package only for its origin constant drops that import).
Check: `gofmt -l core api` prints nothing; `CGO_ENABLED=1 go test -tags fts5 ./api/... ./core/...` passes; searches S-2 and S-3 show only `core/vocabkey/vocabkey.go` (new). Commit.

**STEP-3 — Replace the Go kind lists**
Files: `core/database/promote/promote.go`, `core/database/promote/unpromote.go`, `core/database/promote/batch.go`, `core/database/sourcegraphprogress/sourcegraphprogress.go`, `core/match/registry.go`, `core/database/promotealign/layer.go`, `core/database/promotealign/stats.go`, `core/database/promotealign/canon.go`, `core/database/searchindex/handles.go`, `core/search/search.go`, `core/search/fts.go`, `core/database/subjectvocab/profile_test.go`, `core/match/registry_test.go`, `core/database/autoreconciler/autoreconciler_test.go`.
Do: D-7, D-9. `ErrUnsupportedType` and its call sites stay as they are, now behind `vocabkey.IsPrimaryType`.
Check: `CGO_ENABLED=1 go test -tags fts5 ./api/... ./core/...` passes; `go list -deps ./core/database/sourcegraphprogress | grep 'core/database/promote$'` prints nothing; search S-1 shows only the `OmnibarHitPresentation.swift` arm.

**STEP-4 — Property keys and the seed test**
Files: `core/database/promotealign/layer.go` (`neighborTypeProperty`), `core/database/subjectvocab/vocabkey_test.go` (new).
Do: D-6, D-11; T-4.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/subjectvocab/ ./core/database/promotealign/` passes. Commit STEP-3 and STEP-4 together (D-12).

**STEP-5 — Swift fake and its guard**
Files: `macos/App/Platform/FakeStore.swift`, `api/ffi/swift_vocabulary_test.go`. Skill: `add-swift-test` (for running the suite; no new Swift test file).
Do: D-10; T-5.
Check: `CGO_ENABLED=1 go test -tags fts5 ./api/ffi/ -run TestSwiftVocabularyMatchesGo` passes; search S-1 shows only the `OmnibarHitPresentation.swift` arm; on macOS, `xcodebuild test` per `scripts/check-pr.sh` passes. Commit.

**STEP-6 — Verify**
Files: none.
Do: run AC-1.1, AC-1.2, AC-2.1 – AC-2.3.
Check: each prints what §2 says; `scripts/check-pr.sh` prints `check-pr: PASS`.

**STEP-7 — Skills and docs**
Files: `.agents/skills/add-seeded-vocabulary/SKILL.md`, `docs/seeded-vocabulary.md`, `docs/deployment-plan/spike-11/deployment-plan.md` (map row; S11-02 scope entry), `docs/deployment-plan/spike-11/completed.md`, this plan.
Do: in `add-seeded-vocabulary`, append to rule 3 (the one that starts "All registry rows are"): "Origins, primary kinds and the Property keys code reads by name come from `core/vocabkey`. A vocabulary package does not declare its own `OriginProvenencia` or origin check; it calls `vocabkey.ValidOrigin`." Add as the last bullet of its "Do not" list: "Write `\"provenencia\"`, `\"plugin:\"` or a list of the primary kinds by hand". In `docs/seeded-vocabulary.md` §1.1 "Reserved `origin` values", add one sentence after the paragraph that ends "without updating this catalog.": "Go code names these origins through `core/vocabkey` (`OriginProvenencia`, `OriginUser`, `OriginPluginPrefix`, `ValidOrigin`)." In the map's S11-02 scope entry, "Registry keys use S11-23's kinds and property keys" becomes "Registry keys are built from `core/vocabkey` (S11-23), and the purity guard lists it". Then the DoD lines.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-11/prs/S11-23-primary-kinds.md` reports no errors; links resolve. Commit.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/vocabkey/vocabkey_test.go` `TestIsPrimaryType` *(new)*, table | `person`/`provenencia` → true; `event`/`provenencia` → true; `place`/`provenencia` → true; `person`/`user` → false; `person`/`plugin:x` → false; `participation`/`provenencia` → false; `source`/`provenencia` → false; `Person`/`provenencia` → false | AC-1.3, D-3 |
| T-2 | Go | same file `TestPrimaryKinds` *(new)* | returns `[person event place]`; setting element 0 of one result to `x` leaves the next call's element 0 `person`; `IsPrimaryKind` true for each, false for `location` | D-3 |
| T-3 | Go | same file `TestValidOrigin` *(new)*, table, with an `IsPluginOrigin` column | `provenencia` → valid, not plugin; `user` → valid, not plugin; `plugin:dna` → valid, plugin; `plugin:` → invalid, not plugin; `""` → invalid; `Provenencia` → invalid; `plugins:x` → invalid | AC-1.4, D-4 |
| T-4 | Go | `core/database/subjectvocab/vocabkey_test.go` `TestVocabKeysMatchSeed` *(new)* | root seed type keys = `{person, event, place}`; each of the eight `Property*` constants is a `seedProperties` key | AC-2.4, D-11 |
| T-5 | Go | `api/ffi/swift_vocabulary_test.go` `TestSwiftVocabularyMatchesGo/primary_kinds_match_Go` *(new)* | `EvidencePrimaryKind` cases `{person, event, place}` = `vocabkey.PrimaryKinds()`; read with the existing `rawCase` regexp; zero cases read → `t.Fatal("no cases read")`; the file imports `core/vocabkey` | AC-2.5, D-10 |
| T-6 | Go | existing origin-validation and gate tests: `core/database/deleteimpact` tests, `core/database/promote` tests (unsupported type), `core/database/sourcegraphprogress/sourcegraphprogress_test.go`, `core/search/handles_test.go` | unchanged except qualifiers | AC-3.1 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: tests in the six vocabulary packages that build rows with `<pkg>.OriginUser`, `core/search/handles_test.go`, and the Promote Swift tests that run on `FakeStore`.

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left — none in this PR
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New — no UI in this PR
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together) — or none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status and the S11-02 scope entry (STEP-7); completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated (if gated) — not gated
- [ ] Domain doc updated if a rule in docs/*.md changed: docs/seeded-vocabulary.md §1.1; skill add-seeded-vocabulary per STEP-7
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

1. `go list -deps ./core/vocabkey | grep 'provenencia/core/database'`: no output (AC-1.1).
2. `go test ./core/vocabkey/` without build tags: passes (AC-1.2, AC-1.3, AC-1.4).
3. Searches S-1, S-2 and S-3 from §2: hits only where §2 says (AC-2.1, AC-2.2, AC-2.3).
4. Add `case source` to `EvidencePrimaryKind` locally and run `go test -tags fts5 ./api/ffi/ -run TestSwiftVocabularyMatchesGo`: it fails; revert (AC-2.5). Add a `RoleRoot` type to a local copy of `seedTypes` and run `TestVocabKeysMatchSeed`: it fails; revert (AC-2.4).
5. `scripts/check-pr.sh`: `check-pr: PASS` (AC-3.1). On Linux it prints `SKIP xcode-test "not macOS"`; then the Swift change is checked on macOS before the PR leaves draft.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| A file imported a vocabulary package only for its origin constant and now has an unused import | STEP-2 removes it; `go vet` in `check-pr.sh` catches the rest |
| A local variable named `vocabkey` | None exist at the planning base (D-1); rename the local if one appears |
| `canon.go` callers lose the origin check | They pass the literal `"provenencia"` today, so `IsPrimaryKind` is the same test (D-7) |

**Stop and report if:**

- S11-22 has landed with a Landed note that moves the Origin constants, or adds a package named `vocabkey`, or touches `originOK`.
- An `originOK` in §4 accepts a set other than the one §4 says (D-4 would change behavior).
- A `PrimaryKind` / `promotePrimary` / `handleKinds` caller passes an origin other than provenencia or a key read from the catalog.
- `vocabkey` needs an import other than `strings`.
- Search S-2 or S-3 has a hit that is not an origin (a different meaning of `"provenencia"`).
- More than 95 files name `OriginProvenencia` / `OriginUser` at your base, or a new kind list appears that §4 does not name.

## 12. Open questions

None.

## 13. PR

**Title:** S11-23 — One home for primary kinds
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-11/prs/S11-23-primary-kinds.md`.
