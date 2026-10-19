# S11-15 — Shortlist index

**Status:** Blocked — on S11-13 (planned, not landed: STEP-0 redoes §4 against its code)
**Spike / slice:** Spike 11 · Slice 4 — Segments and graph alignment
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no new chrome: an engine index; the wire, the page and its copy are unchanged)
**Board:** —
**Depends on:** S11-13 (Planned, [plan](S11-13-promote-subject-compare.md); not landed). Runs in parallel with S11-14 (D-9 covers either landing order). Branch from `main` after S11-13 merges
**Blocks:** S11-16
**Branch:** `s11-15-shortlist-index`
**Version bump:** none — Spike 11 PRs leave `VERSION` at 0.2.0, and the product bump is cut when a release ships ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-match-rule`](../../../../.agents/skills/add-match-rule/SKILL.md) ("Add a knob", "Add a layer package")
**Rules that bind:** `.cursor/rules/go-tests.mdc`, `.cursor/rules/match-rules.mdc`
**Context:** [Layer 6 procedure](../comparison-layers.md#layer-6-procedure) step 1; [Vocabulary: finding candidates](../comparison-layers.md#vocabulary-finding-candidates); [Decided on 2026-10-17](../comparison-layers.md#decided-on-2026-10-17) (misspellings: start simple); [Match registry](../comparison-layers.md#match-registry); [`docs/matching-model.md`](../../../matching-model.md) glossary and knob inventory (`alignment.Shortlist.*`); working rules "Pure means the package builds and tests with no build tags", "A normalized `Key` per value type", "One match registry", "One guard, not one per package", "Registry numbers: carry over, then estimate"; [S11-01 ship log](S11-01-matching-model.ship.md) follow-up PRF-1 (the event shortlist bucket grows with the catalog)
**Planned:** 2026-10-19 against `main` @ `6ecfd88` (S11-13's plan merged, its code not yet). Draft: §4 rows marked *(S11-13)* describe S11-13's plan, not code
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

Promote stops comparing a Source Subject against every handle of its kind. A pure index maps shortlist keys to handles, per kind: a person's name words (not initials), an event's type with the year buckets its dates cover, a place's toponym words, each built with value compare's normalization so shortlisting and comparison agree on "the same". Seeding looks a Subject's keys up and scores only the handles that come back. The index is built once per catalog state, kept in `graphcache`'s stats slot, and dropped with a handle commit, as the stats are today.

**Why now.** After S11-13, `loadCanon` profiles every handle of a kind and runs `graphalign.Candidates` over all of them on each proposal: the work grows with the catalog, not the Source (dogfood bar 3, "fast"). S11-16's layer 6 walk starts from a shortlist ("shortlist → score, keep top K"); this PR gives it one. PRF-1 asked the event bucket to be bounded: D-5 bounds it by year.

**What changes for the researcher** (the design, not regressions):
- A handle that shares no shortlist key with the Subject is not seeded, even when S11-13's full scan would have reached the weak band: a name whose only likeness is a misspelling ("Robnis" against "Robins", no shared word), an initials-only name against a full one ("J. R." against "James Robins"), a toponym that differs only by a misspelling. This is the 2026-10-17 decision "accept the gap; links catch most of it". The walk still reaches such a handle through a mapped neighbor, and the fallback still scores every canon handle (unchanged, S11-16).
- Nothing else: every handle that shares a key is scored by the same `Candidates` rule as in S11-13, so the seeded set for every golden is the same.

**What stays.** `graphalign.Candidates` and its gate (S11-13 D-4), `fallbackUnreachable` over the canon handles, the walk, refinement, bands, the wire, Swift. The stats (`Stats.FanOut`) stay where they are until S11-14 deletes them (D-9).

## 2. User stories

Conventions. `name(f)` is `propertycomparetest.Name(f)` (a western-pattern name value from a form; check the helper's exact signature at STEP-0 and use it), `text(s)` is `propertycomparetest.Text(s)`, `term(k)` is `propertycomparetest.Term(k)`, `date(y)` is `propertycomparetest.Date(datevaluetest.YMD(y, 0, 0))`. `R` is `matchregistry.Default()`. "Lookup" is `(*shortlist.Index).Lookup` on an index built with `R.Alignment.Shortlist`, `R.Value` and `subjectcomparetest.Metas()`. Alias table `bill` = one nickname set `{bill, william}` for the given role.

**US-1.** As the engine, I find the handles worth comparing by their shortlist keys, so a Subject is scored against a few handles, not the whole kind.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | Persons indexed: PER-1 `{name: name("James Robins")}`, PER-2 `{name: name("Robins")}`, PER-3 `{name: name("Mary Robins")}`, PER-4 `{name: name("James Robbins")}`, PER-5 `{name: name("Ann Ames")}`. Lookup `{name: name("J. Robins")}` | PER-1, PER-2, PER-3 (every handle with `robins`); not PER-4, not PER-5 |
| AC-1.2 | Index PER-1 `{name: name("William")}`, PER-2 `{name: name("Robert")}` under alias table `bill`. Lookup `{name: name("Bill")}` | PER-1 only (the alias set is a key) |
| AC-1.3 | Index PER-1 `{name: name("José Robins")}`. Lookup `{name: name("Jose Robins")}` | PER-1 (folded word key) |
| AC-1.4 | Places indexed: PLC-1 `{toponym: text("Saint-Jean")}`, PLC-2 `{toponym: text("York")}`. Lookup `{toponym: text("SaintJean")}`; lookup `{toponym: text("New York")}` | First: PLC-1 (the whole-value key; the words break differently). Second: PLC-2 (shared word `york`) |
| AC-1.5 | Events indexed: EVT-1 `{event_type: term(birth), date: date(1850)}`, EVT-2 `{term(birth), date(1852)}`, EVT-3 `{term(birth), date(1870)}`, EVT-4 `{term(death), date(1850)}`, EVT-5 `{term(birth)}` (undated). Lookup `{event_type: term(birth), date: date(1850)}` | EVT-1 and EVT-2; not EVT-3, EVT-4, EVT-5 |
| AC-1.6 | AC-1.5's index plus EVT-6 `{term(birth), date: BEF 1800}`. Lookup `{term(birth), date(1700)}`; lookup `{term(birth), date: AFT 1860}` | First: EVT-6 (an open date is `wide`, looked up by every dated birth). Second: EVT-1, EVT-2, EVT-3, EVT-6 (an open probe looks up `any`); never EVT-4 or EVT-5 |
| AC-1.7 | Lookup `{sex_at_birth: term(male)}` on AC-1.1's index; lookup `{event_type: term(birth)}` on AC-1.5's index; lookup on a kind with no rules (`vocabkey.SubjectType("source")`) | Nothing, each time |
| AC-1.8 | An event `{term(residence), start_date: date(1850), end_date: date(1860)}` indexed; lookup `{term(residence), date: date(1855)}` | Found (the period's years are keyed, as `subject.Periods` compares them) |

**US-2.** As a contributor, shortlisting never misses a pair value compare calls the same.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | Every ordered pair of the name corpus in T-1 (accents, initials, particles, word breaks, alias pairs, generation words), as name values and as text values, under the `bill` table | Whenever `valuecompare.Compare` says `Same` or `Equivalent`, `valuecompare.WordKeys` of the two share a key |
| AC-2.2 | Every ordered pair of the date corpus in T-3 (points, months, days, `ABT`, `BEF`, `AFT`, closed ranges, years ±0–4 apart) as two births | Whenever the date relation is neither `Different` nor `Unknown`, `shortlist.Keys`' lookup keys of one and index keys of the other share a key |
| AC-2.3 | Each kind in `R.Subject.Weights` | The most evidence a pair of that kind can earn without a shortlist key is under `DefaultConfig().WeakScore` (person without `name`: 0.6 + 0.1; event without a date: 0.5; place without `toponym`: 0) |

**US-3.** As the engine (the Promote loader), I build the index once per catalog state and seed from it.

Catalog tests in `core/database/promotealign` with the existing `fixture`.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | Sources A and B each hold person "Ann Ames". Propose B (nothing promoted: index built, empty); promote A's person; propose B again | Second proposal: B's person targets A's person handle (the handle commit dropped the index) |
| AC-3.2 | `TestProposeQueryCountDoesNotGrowWithTheLayer` | Passes unchanged |
| AC-3.3 | Every existing `core/database/promotealign` and `api/ffi/handlers` Promote test | Passes with no expected value changed |
| AC-3.4 | `graphcache` with a stored match index; `SetAliasTables`; a commit that names a handle; `Drop` | Each leaves `MatchIndex()` reporting false |

## 3. Scope

| In | Covered by |
| --- | --- |
| `valuecompare.WordKeys` and alias set keys (D-2, D-3) | STEP-1, US-2 |
| `core/shortlist` *(new)*: `Rules`, `Keys`, `Index` (D-1, D-4, D-5, D-6) | STEP-2, US-1, AC-2.2 |
| Registry `Alignment.Shortlist` section, inventory rows, guard row; `graphalign.Config.Shortlist` (D-7, D-8) | STEP-3, AC-2.3 |
| `graphcache` stats slot renamed to the match index slot; dropped by `SetAliasTables` too (D-9, D-10) | STEP-4, AC-3.4 |
| Promote seeding through the index (D-11, D-12) | STEP-5, US-3 |
| Docs the index makes false (D-13) | STEP-6 |

| Out | Home |
| --- | --- |
| Shortlisting inside the walk's fallback and the one queue; top K with ties; similarity as the gate | S11-16 |
| Fan-out and `graphalign.Stats` removal | S11-14 |
| A coarse misspelling key (surname consonant skeleton) | Not planned — [decided 2026-10-17](../comparison-layers.md#decided-on-2026-10-17): only if the sparse-data tier shows a problem |
| Updating the index in place on a commit instead of rebuilding it | S11-21 (only if the 100-Subject check fails) |
| A benchmark for seeding time | S11-21 |
| Shortlisting for search, the reconciler or merge hints | Not planned — no caller asks for it |

## 4. Current state

Rows marked *(S11-13)* are S11-13's plan, not code at the planning base. STEP-0 checks each against the merged S11-13.

| Area | Where | Today |
| --- | --- | --- |
| Value keys | `core/valuecompare/valuecompare.go` `Key(v, level, rules)`; `name.go` `name.key`, `nameWords`, `nameKey`; `text.go` `text.key`, `joinedKey`, `TextKey` | One key per value; name keys ignore alias tables because sets do not chain. No per-word key exists |
| Words | `core/valuecompare/words.go` `SplitWords`, `FoldWord`, `IsInitial` | The one word helpers |
| Aliases | `core/valuecompare/aliases.go` `AliasIndex{sets []aliasSet; byToken}`, `aliasSet{tokens, roles, kind}`, `AliasToken`, `NewAliasIndex`, `Hit` | No way to ask which sets hold a token |
| Date ends | `core/valuecompare/datewindow.go` `AboutEnds(d, units)`, `DateEnd{Open, Year, …}`; `date.go` reads dates through `AboutEnds` and calls a gap over `Tolerance` civil years `Different` | |
| Profiles | `core/subjectcompare/subjectcompare.go` `Values` (`map[vocabkey.PropertyKey][]propertycompare.Value`), `Metas`; `core/propertycompare/propertycompare.go` `Value{valuecompare.Value; Negative; Neutral}`, `Positive()` | Profile values carry no `Type`: layer 2 sets it from the meta (S11-13 D-12, `subjectprofile.Value` doc) |
| Registry | `core/matchregistry/registry.go` `Registry{Value, Subject}`, `Default()`; `subject.go`; `value.go` `Dates.Tolerance` 2, `Dates.AboutUnits` 1 | No alignment section |
| Inventory guard | `core/matchregistry/inventory_test.go` `notYetBuilt` (`"segment."`, `"alignment."`, `"evidence."`), `TestKnobInventoryMatchesRegistry`, `registryKeys` (a map of struct or map is walked by key; any other map is one leaf) | |
| Purity guard | `core/matchregistry/guard_test.go` `layerPackages` | No `core/shortlist` row |
| Knob inventory | `docs/matching-model.md` §Knob inventory rows `alignment.Shortlist.YearBucket` (2), `alignment.Shortlist.person`, `.event`, `.place`; glossary rows "shortlisting", "shortlist key", "shortlisted handle" | Written in S11-01 before the shape existed; D-8 replaces the three kind rows |
| Walk config | `core/graphalign/registry.go` `DefaultConfig`, `Config` (`WeakScore` 2); *(S11-13)* `Config.Subject` from `matchregistry.Default().Subject` | |
| Candidates | *(S11-13; (new) at the planning base)* `core/graphalign/candidates.go` `Candidates(kind, values, handles, metas, cfg, limit)`: handles whose evidence reaches `cfg.WeakScore`, sorted; doc says "Shortlisting replaces the scan (S11-15)" | |
| Fallback | *(S11-13)* `core/graphalign/align.go` `fallbackUnreachable` over `st.handlesOfKind` (the canon handles) | Unchanged by this PR |
| Seeding | *(S11-13)* `core/database/promotealign/canon.go` `loadCanon(g, layer, fixed, cfg, namePattern)`: per kind, `profiler.profile` on every `g.Kind` node, then `graphalign.Candidates(…, seedCandidateLimit)` per unfixed Subject | The scan this PR replaces |
| Profiler | *(S11-13; (new) at the planning base)* `core/database/promotealign/profiles.go` `profiler`, `profile(nodes)`, `values(id)`; `typeByKey`, `typeKey` in `canon.go` | |
| Stats slot | `core/database/graphcache/graph.go` `Graph.stats`, `statsOK`, `Stats()`, `SetStats`, `ResetStatsForTest`, `clear`; `commit.go` (a commit naming handles clears `statsOK`); `SetAliasTables` | `any`-typed slot, dropped on a handle commit and a structure drop |
| Stats use | `core/database/promotealign/stats.go` `loadStats`, `cloneStats`, `computeStats`; `export_test.go` (wraps `loadStats`); `promotealign.go` `Propose` calls `loadStats` | Only caller of the slot. S11-14 deletes fan-out (map review K-10) |
| Aliases warm | `core/database/namealiases/namealiases.go` `Warm` → `g.SetAliasTables`; `Invalidate` | Tables change only through `Warm` |
| Tests | `core/valuecompare/contract_test.go` `TestKeyContract`; `core/database/promotealign/promotealign_test.go` `TestProposeQueryCountDoesNotGrowWithTheLayer`, fixture `newFixture`, `person`, `promote`, `propose`; `core/database/graphcache/graph_test.go` | |

**Precedent.** S11-02 ([#441](https://github.com/mendahu/provenencia/pull/441), [plan](S11-02-value-compare-core.md)) added a layer package, its registry section, its guard row and inventory rows together: copy that shape for `core/shortlist`. S11-07 ([#460](https://github.com/mendahu/provenencia/pull/460), [plan](S11-07-normalization-key.md)) put one normalization rule in `valuecompare` and had every other caller read it: `WordKeys` follows it. The stats slot (S9-42, `loadStats`) is the cache pattern to keep: build on a miss, store, drop on a handle commit.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | **A new pure layer package, `core/shortlist`.** It declares `Rules`, `Dated`, `Keys`, `Index`, `NewIndex`, `Add`, `Lookup`. It imports `core/valuecompare`, `core/propertycompare`, `core/subjectcompare`, `core/vocabkey` and nothing under `core/database`. Package doc: "Package shortlist is layer 6's first step: the keys a Subject is looked up by and the index that maps them to handles. It is pure." | `matchregistry` must name the section's types, and `graphalign` already imports `matchregistry` (`DefaultConfig`), so the types cannot live in `graphalign`. The index is shared by the loader and S11-16's walk | Types in `graphalign` — an import cycle; the index in `promotealign` — S11-16's walk is pure and could not read it |
| D-2 | **Word keys are layer 1's.** `valuecompare.WordKeys(v Value, r Rules) []string` *(new, `core/valuecompare/shortlist.go` (new))*, for name and text values (`nil` for any other type and for a value with no word): `"=" + k` where `k, ok := Key(v, Equivalent, r)`; `"w:" + FoldWord(word, r.Words.Folds)` for each word that is not `IsInitial`; `"a:" + key` for each alias set key of each word (D-3). Names read their words through `nameWords` (ignored parts dropped, generation words left out); text through `SplitWords`. Sorted, no duplicates. Doc: "WordKeys are the keys shortlisting looks a value up by. Two values Compare calls Same or Equivalent share one: the whole key at Equivalent, or, for an alias, the set. Values that share a word that is not an initial share one too. Initials and generation words are never keys alone." | Layer 1 owns "the same" (working rule "A normalized `Key` per value type"); the guarantee is tested where `Compare` lives | Keys built in `core/shortlist` from `SplitWords` — a second reading of name parts that ignores the pattern's roles; the whole `Key` alone — "J. Robins" would not find "James Robins" |
| D-3 | **Alias sets as keys.** `aliasSet` gains `key string`: its tokens sorted and joined by one space, set in `NewAliasIndex`. `func (a AliasIndex) SetKeys(token string) []string` *(new)*: the keys of every set holding `token` (an `AliasToken`), roles ignored, sorted. `WordKeys` calls it with `AliasToken(word, r.Words)` | Sets do not chain, so no single word key can join them, but a set can be a key: two words that `Hit` share it. Content keys survive a rebuilt index; roles are ignored because a shortlist may over-include | Set positions as keys — they change when tables reload |
| D-4 | **Shortlist rules.** `shortlist.Rules{Words map[vocabkey.SubjectType][]vocabkey.PropertyKey; Dated map[vocabkey.SubjectType]Dated; YearBucket int}`, `Dated{Type vocabkey.PropertyKey; Dates []vocabkey.PropertyKey}`. Field docs: Words "the properties whose word keys key a Subject of the kind"; Dated "the kinds keyed by a type term and the years their dates cover"; Type "the term property keyed with each year bucket"; Dates "the date properties whose years are read"; YearBucket "years per bucket" | Typed keys (registry rule); the inventory guard reads these as `alignment.Shortlist.Words`, `alignment.Shortlist.Dated.event.Type`, `.Dates`, `alignment.Shortlist.YearBucket` | One struct per kind with every field — every kind would carry empty `Type` and `Dates` rows in the inventory |
| D-5 | **Keys.** `shortlist.Keys(kind vocabkey.SubjectType, v subjectcompare.Values, metas subjectcompare.Metas, r Rules, value valuecompare.Rules) (index, lookup []string)`. Only positive values (`propertycompare.Value.Positive`) key; each value's `Type` is set from `metas[key].ValueType` before keying. **Word kinds:** for each `r.Words[kind]` property, `string(property) + "|" + k` for each `WordKeys` key `k`; `index` and `lookup` are the same list. **Dated kinds:** for each positive `Dated.Type` value with `typeKey, ok := valuecompare.Key(t, Equivalent, value)`, and each positive date of the `Dates` properties read with `valuecompare.AboutEnds(*d, value.Dates.AboutUnits)`: when either end is open, the date is *wide*; otherwise its buckets are `floorDiv(y, r.YearBucket)` for `y` from `first.Year − value.Dates.Tolerance` to `last.Year + value.Dates.Tolerance`. A closed date adds `"t:" + typeKey + "|" + bucket` to both lists, `"t:" + typeKey + "|any"` to `index`, and `"t:" + typeKey + "|wide"` to `lookup`. A wide date adds `"t:" + typeKey + "|wide"` and `"t:" + typeKey + "|any"` to `index` and `"t:" + typeKey + "|any"` to `lookup`. A date `AboutEnds` cannot place keys nothing. An undated event keys nothing. Both lists sorted, no duplicates. `floorDiv` rounds toward minus infinity | Both sides padded by the tolerance share a bucket whenever the years are within it (`date.go`'s Near rule); `wide` / `any` cover open dates without making every probe scan its type; the event bucket is bounded by type and year (PRF-1). An undated event cannot reach the weak band (AC-2.3) | Probing neighbor buckets — the same reach, with two sides that must agree on direction; one key per year — a wide range would add a key per year |
| D-6 | **Index.** `type Index struct` *(new)*: unexported `rules Rules`, `value valuecompare.Rules`, `metas subjectcompare.Metas`, `byKey map[vocabkey.SubjectType]map[string][][]byte`. `NewIndex(r Rules, value valuecompare.Rules, metas subjectcompare.Metas) *Index`. `(*Index).Add(kind vocabkey.SubjectType, id []byte, v subjectcompare.Values)` files a copy of `id` under each `index` key. `(*Index).Lookup(kind vocabkey.SubjectType, v subjectcompare.Values) [][]byte`: the ids under any `lookup` key, deduplicated, sorted by `bytes.Compare`, each a copy; nil for a kind with no entry. A nil `*Index` looks up nothing. The index is read-only after it is built | The index keeps the rules it was built with, so a lookup can never key differently from the build | Passing rules to every call — a caller could mix two rule sets |
| D-7 | **Registry section.** `core/matchregistry/alignment.go` *(new)*: `type Alignment struct{ Shortlist shortlist.Rules }` (doc: "Alignment is layer 6's section. S11-16 adds the walk's knobs.") and `defaultAlignment()`: `Words` person → `[name]`, place → `[toponym]`; `Dated` event → `{Type: event_type, Dates: [date, start_date, end_date]}`; `YearBucket: 2` (comment `// estimate: the date tolerance, so a Near date shares a bucket`). `Registry` gains `Alignment Alignment` and `Default()` sets it. `graphalign.Config` gains `Shortlist shortlist.Rules` (after `Subject`, doc "Shortlist is the registry's shortlist keys. DefaultConfig holds the registry's."), set in `DefaultConfig` from `matchregistry.Default().Alignment.Shortlist` | One registry (working rule); the layer takes its section as an argument; `Config` already carries `Value` and `Subject` the same way | A `Registry.Shortlist` field — the design's table puts shortlist keys in the alignment section, which S11-16 grows |
| D-8 | **Inventory.** In `docs/matching-model.md`, the rows `alignment.Shortlist.person`, `.event`, `.place` are replaced by `alignment.Shortlist.Words` (start: person `name`, place `toponym`), `alignment.Shortlist.Dated.event.Type` (`event_type`), `alignment.Shortlist.Dated.event.Dates` (`date`, `start_date`, `end_date`); each source "decided (comparison-layers.md layer 6 procedure)". `alignment.Shortlist.YearBucket` stays. In `inventory_test.go`, `notYetBuilt`'s `"alignment."` becomes `"alignment.AcceptBar"`, `"alignment.Bands."`, `"alignment.TopK"`, `"alignment.TieCap"`, `"alignment.Alternatives"`, `"alignment.SettleMargin"`, `"alignment.SettleRounds"`. `guard_test.go` gains `{"core/shortlist", true}`; `.cursor/rules/match-rules.mdc` gains its glob; the `add-match-rule` Check command gains `./core/shortlist/...` | The guard needs a row per field and a prefix only for what has not landed | Keeping `"alignment."` — it would match the new field and fail the guard |
| D-9 | **One slot, either landing order.** In `graphcache`, `stats`/`statsOK`/`Stats()`/`SetStats`/`ResetStatsForTest` are renamed `matchIndex`/`matchIndexOK`/`MatchIndex()`/`SetMatchIndex`/`ResetMatchIndexForTest`; docs say "the match index: what Promote derives from the whole catalog". `promotealign` stores one value there, `catalogIndex{stats graphalign.Stats; shortlist *shortlist.Index; namePattern string}` *(new, `stats.go` renamed `index.go`)*, through `cachedIndex(g) (catalogIndex, bool)` and `storeIndex(g, catalogIndex)`. `loadStats` reads and fills `stats`; `loadShortlist` (D-11) reads and fills `shortlist`. If S11-14 has landed first and deleted `graphalign.Stats` and `loadStats`, `catalogIndex` has no `stats` field and this PR adds none | The map gives S11-14 fan-out and S11-15 the slot (review K-10); one struct holds what both need while they overlap | Two slots — a second drop rule to keep in step |
| D-10 | **Drops.** The slot is dropped where stats are today (a commit naming handles, `Drop`) and also by `SetAliasTables`, which changes the alias sets `WordKeys` reads. A stored index whose `namePattern` differs from this proposal's `valuerules.Load` pattern is rebuilt (the default name format changes word roles without a handle commit) | Every input the keys read is either in the catalog index or checked | Dropping on every vocabulary write — term labels do not change keys |
| D-11 | **Build once, then look up.** `loadShortlist(g *graphcache.Graph, p *profiler, types map[string][]byte, cfg graphalign.Config, metas subjectcompare.Metas, value valuecompare.Rules, namePattern string) (*shortlist.Index, error)` *(new, `index.go`)*: a stored index with the same `namePattern` is returned; otherwise `shortlist.NewIndex(cfg.Shortlist, value, metas)`, and for each kind in `cfg.Shortlist.Words` or `cfg.Shortlist.Dated` (sorted) with a 16-byte type id, every `g.Kind` node is profiled (`p.profile`) and `Add`ed with `p.values(n.ID)`; the result is stored. `loadCanon`'s seed loop calls `Lookup(kind, s.Values)`, loads those ids with one `g.Nodes` call per Subject, profiles them, and passes them as `[]graphalign.Handle` (`Ref`-sorted) to `graphalign.Candidates` with `seedCandidateLimit`. A looked-up id with no unmerged node is skipped | Warm proposals profile only shortlisted handles; the cold build costs what S11-13's scan costs today | Indexing from raw rows — keys must come from the displayed values the profile holds (strict profile) |
| D-12 | **Candidates keeps its signature.** Its doc's "Shortlisting replaces the scan (S11-15)" becomes "Seeding passes it shortlisted handles (S11-15); top K replaces the limit (S11-16)." `fallbackUnreachable` keeps calling it with `st.handlesOfKind` | The fallback reads only the canon set, which the seeds and the BFS already bound; S11-16 rewrites it | Shortlisting the fallback now — S11-16 deletes it |
| D-13 | **Docs.** `docs/matching-model.md` (D-8; glossary "shortlist key" gains "built with `valuecompare.WordKeys`, or an event type and year bucket (`core/shortlist`)"); `comparison-layers.md` layer 6 step 1 adds "**Built in S11-15** (`core/shortlist`): word keys for person names and toponyms, type and year bucket for events; seeding reads it, the walk from S11-16"; and "place, toponym" becomes "place, toponym words"; `docs/promote-graph-alignment.md` line 291's stats sentence names the match index; `.agents/skills/add-match-rule/SKILL.md` "Add a layer package" Check command | Working rule "One matching skill, grown with the layers" | — |

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `valuecompare.WordKeys` *(new)*, `core/valuecompare/shortlist.go` *(new)* | func, file | `valuecompare.Key`, `TextKey` | `Key` is one key per value; a shortlist needs a key per word to find `Contains` pairs. `WordKeys` includes `Key` (D-2) |
| `AliasIndex.SetKeys`, `aliasSet.key` *(new)* | method, field | `AliasIndex.Hit` | `Hit` answers "do these two share a set"; an index needs "which sets hold this" (D-3) |
| `core/shortlist` *(new)*: `Rules`, `Dated`, `Keys`, `Index`, `NewIndex`, `Add`, `Lookup` | package | `graphcache.Graph.Kind` (the scan); `promotealign.loadStats` (the cached derived value) | No key → handles map exists; `Kind` lists a whole kind (D-1) |
| `matchregistry.Alignment`, `defaultAlignment` *(new)*, `core/matchregistry/alignment.go` *(new)* | type, func, file | `subject.go` `defaultSubject` | The design's alignment file; same shape (D-7) |
| `graphalign.Config.Shortlist` *(new field)* | field | `Config.Value`, `Config.Subject` | Same pattern (D-7) |
| `graphcache` `MatchIndex`, `SetMatchIndex`, `ResetMatchIndexForTest` *(renamed)* | method | `Stats`, `SetStats`, `ResetStatsForTest` | The same slot, renamed for what it holds (D-9) |
| `promotealign` `catalogIndex`, `cachedIndex`, `storeIndex`, `loadShortlist` *(new)*; `stats.go` → `index.go` | type, func, file | `loadStats`, `cloneStats` | One slot for stats and shortlist (D-9, D-11) |
| `TestWordKeys`, `TestWordKeysShareOnSameOrEquivalent`, `TestAliasSetKeys` *(new, `core/valuecompare`)*; `TestKeys`, `TestLookup`, `TestDatedKeysNeverMissANearDate` *(new, `core/shortlist`)*; `TestShortlistSection`, `TestUnkeyedEvidenceStaysUnderTheWeakBand` *(new)*; `TestProposeSeesAHandlePromotedAfterTheIndexWasBuilt` *(new)*; `TestMatchIndexDrops` *(new)* | test | `TestKeyContract`, `TestSubjectSection`, `TestProposeQueryCountDoesNotGrowWithTheLayer` | Same table shapes |

**Builds on.** `valuecompare.Key`, `SplitWords`, `FoldWord`, `IsInitial`, `AliasIndex`, `AboutEnds`; the registry and its two guards; S11-13's `profiler` and `Candidates`; the graph's stats slot.
**Deletes.** `graphcache` `Stats`, `SetStats`, `ResetStatsForTest` (renamed); `core/database/promotealign/stats.go` (renamed `index.go`); the per-kind `g.Kind` profiling loop in `loadCanon`'s seeding; the `notYetBuilt` prefix `"alignment."`; the inventory rows `alignment.Shortlist.person`, `.event`, `.place`.
**Next adds.** S11-16's walk calls `Lookup` for every Subject and keeps the top K; the index then moves into `graphalign.Canon` or an `Align` argument (S11-16 decides). A new shortlist property is one `Words` entry or one `Dated` entry, and an inventory row. A coarse misspelling key is one more prefix inside `WordKeys`. S11-14 deletes `catalogIndex.stats`. S11-21 replaces the rebuild with an in-place update only if its benchmark fails.

### Schema

None — no catalog change.

### Go core

| Package | Change |
| --- | --- |
| `core/valuecompare` | `shortlist.go` *(new)*: `WordKeys`; `aliases.go`: `aliasSet.key`, `SetKeys` (D-2, D-3) |
| `core/shortlist` *(new)* | `shortlist.go` *(new)*: `Rules`, `Dated`, `Keys`, `floorDiv`; `index.go`: `Index`, `NewIndex`, `Add`, `Lookup` (D-1, D-4 – D-6) |
| `core/matchregistry` | `alignment.go` *(new)*; `registry.go` `Registry.Alignment`; `inventory_test.go`, `guard_test.go`, `registry_test.go` (D-7, D-8) |
| `core/graphalign` | `registry.go` `Config.Shortlist`; `candidates.go` doc (D-7, D-12) |
| `core/database/graphcache` | `graph.go`, `commit.go`: the slot rename; `SetAliasTables` drops it (D-9, D-10) |
| `core/database/promotealign` | `stats.go` → `index.go` *(new)*; `canon.go` seeding; `promotealign.go`; `export_test.go` (D-9, D-11) |

```go
// core/shortlist

type Rules struct {
	Words      map[vocabkey.SubjectType][]vocabkey.PropertyKey
	Dated      map[vocabkey.SubjectType]Dated
	YearBucket int
}

type Dated struct {
	Type  vocabkey.PropertyKey
	Dates []vocabkey.PropertyKey
}

func Keys(kind vocabkey.SubjectType, v subjectcompare.Values, metas subjectcompare.Metas, r Rules, value valuecompare.Rules) (index, lookup []string)

type Index struct{ /* rules, value, metas, byKey */ }

func NewIndex(r Rules, value valuecompare.Rules, metas subjectcompare.Metas) *Index
func (x *Index) Add(kind vocabkey.SubjectType, id []byte, v subjectcompare.Values)
func (x *Index) Lookup(kind vocabkey.SubjectType, v subjectcompare.Values) [][]byte

// core/valuecompare

func WordKeys(v Value, r Rules) []string
func (a AliasIndex) SetKeys(token string) []string

// core/database/promotealign (unexported)

type catalogIndex struct {
	stats       graphalign.Stats // until S11-14
	shortlist   *shortlist.Index
	namePattern string
}
```

`core/shortlist` holds no numeric literal other than 0 and 1 (the guard). `floorDiv(y, w int) int` is `q := y / w; if y%w != 0 && y < 0 { q-- }; return q`.

### FFI

None — no proto or handler change.

### macOS

None — no Swift change.

### Strings

None — no user-facing copy.

## 7. Steps

**STEP-0 — Rebase the draft on S11-13**
Files: this plan.
Do: after S11-13 merges, read its **Landed** note and redo every §4 row marked *(S11-13)* against the code (`rg -n 'func Candidates|func \(p \*profiler\)|seedCandidateLimit|handlesOfKind' core/graphalign core/database/promotealign`). Update the **Planned** line and drop "Draft" from it. A row that is no longer true is a stop condition until the planner refreshes the plan.
Check: every *(S11-13)* symbol resolves with `rg`; status is set to `Planned` by the planner, not the implementer.

**STEP-1 — Word keys in value compare**
Files: `core/valuecompare/shortlist.go` *(new)*, `core/valuecompare/aliases.go`, `core/valuecompare/shortlist_test.go` *(new)*. Skill: `add-match-rule`.
Do: D-2, D-3. T-1, T-2.
Check: `go test ./core/valuecompare/...` passes with no tags; `TestKeyContract` is unchanged.

**STEP-2 — The shortlist package**
Files: `core/shortlist/shortlist.go`, `index.go`, `shortlist_test.go` *(new)* (all three). Skill: `add-match-rule` ("Add a layer package").
Do: D-1, D-4, D-5, D-6. Tests from an external `shortlist_test` package with `propertycomparetest`, `datevaluetest`, `subjectcomparetest.Metas()` and `matchregistry.Default()`. T-3, T-4.
Check: `go test ./core/shortlist/...` passes with no tags.

**STEP-3 — Registry section and config**
Files: `core/matchregistry/alignment.go` *(new)*, `registry.go`, `registry_test.go`, `inventory_test.go`, `guard_test.go`, `core/graphalign/registry.go`, `core/graphalign/candidates.go`, `docs/matching-model.md` (inventory rows only), `.cursor/rules/match-rules.mdc`. Skill: `add-match-rule` ("Add a knob").
Do: D-7, D-8, D-12. T-5, T-6.
Check: `go test ./core/matchregistry/... ./core/graphalign/...` passes; `TestKnobInventoryMatchesRegistry` and the purity guard pass with the `core/shortlist` row.

**STEP-4 — The match index slot**
Files: `core/database/graphcache/graph.go`, `commit.go`, `graph_test.go`, `core/database/promotealign/stats.go` → `index.go`, `export_test.go`, `promotealign.go`.
Do: D-9, D-10 (the slot and its drops; `catalogIndex` with `stats` only so far). Rename every caller (`rg -n 'ResetStatsForTest|SetStats|\.Stats\(\)' --type go` prints nothing after). T-7.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/graphcache/ ./core/database/promotealign/` passes.

**STEP-5 — Seed from the index**
Files: `core/database/promotealign/index.go` *(new)* (from STEP-4), `canon.go`, `promotealign_test.go`.
Do: D-10 (the pattern check), D-11. `catalogIndex` gains `shortlist` and `namePattern`. T-8; the existing suites stay green (AC-3.2, AC-3.3).
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/promotealign/ ./api/ffi/handlers/` passes with no expected value changed; `rg -n 'g\.Kind\(' core/database/promotealign` shows only `loadShortlist` and the code S11-13 left outside seeding.

**STEP-6 — Docs**
Files: `docs/matching-model.md` (glossary), `docs/deployment-plan/spike-11/comparison-layers.md`, `docs/promote-graph-alignment.md`, `.agents/skills/add-match-rule/SKILL.md`, `docs/deployment-plan/spike-11/completed.md`, `docs/deployment-plan/spike-11/deployment-plan.md` (map row), this plan (status).
Do: D-13 and the Definition of done.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-11/prs/S11-15-shortlist-index.md` reports no errors; links resolve.

Commits: one per step from STEP-1, each subject one full imperative sentence ("Look Promote's seeds up in a shortlist index instead of scanning every handle of the kind.").

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/valuecompare/shortlist_test.go` `TestWordKeys` *(new)* | Table: name "J. Robins" → `["=…", "w:robins"]` (the `=` key is `Key` at Equivalent); "José Robins" → contains `w:jose`; "John Smith Jr" → no `w:jr`; text "New York" → `w:new`, `w:york`; "Bill" under table `bill` → contains `a:bill william`; integer value → nil; empty name → nil | D-2, D-3, AC-1.2, AC-1.3 |
| T-2 | Go | `core/valuecompare/shortlist_test.go` `TestWordKeysShareOnSameOrEquivalent` *(new)* | Corpus (at least 24 forms: "James Robins", "J. Robins", "Robins", "José", "Jose", "JOSE", "Mc Donald", "McDonald", "Saint-Jean", "SaintJean", "Bill", "William", "Robert Jr", "Robert", "Ann-Marie", "Ann Marie", "Ó Briain", "O Briain", "J. R.", "J R", "Mary Robins", "Robnis", "Müller", "Muller") × itself, as name values (family = last word, given = the rest) and as text, under table `bill`: when `Compare` is `Same` or `Equivalent`, the `WordKeys` share a key; the test also asserts it checked at least one `Equivalent` alias pair and one word-break pair | AC-2.1 |
| T-3 | Go | `core/shortlist/shortlist_test.go` `TestKeys`, `TestLookup` *(new)* | AC-1.1 – AC-1.8 rows, each a table row; `Keys` on a value with `Negative` set → no key from it | AC-1.1, AC-1.2, AC-1.3, AC-1.4, AC-1.5, AC-1.6, AC-1.7, AC-1.8 |
| T-4 | Go | `core/shortlist/shortlist_test.go` `TestDatedKeysNeverMissANearDate` *(new)* | Date corpus (points 1846–1854, `1850-03`, `1850-03-04`, `ABT 1850`, `BEF 1850`, `AFT 1850`, `BET 1840 AND 1845`) × itself as births: when the date relation under `R.Value` is neither `Different` nor `Unknown`, one's `lookup` and the other's `index` share a key | AC-2.2 |
| T-5 | Go | `core/matchregistry/registry_test.go` `TestShortlistSection` *(new)* | `YearBucket` > 0; every `Words` and `Dated` property is weighted in `Default().Subject.Weights[kind]` or is a `Periods[kind]` start / end | D-7 |
| T-6 | Go | `core/graphalign/candidates_test.go` `TestUnkeyedEvidenceStaysUnderTheWeakBand` *(new)* | For each kind in `cfg.Subject.Weights`: the sum of `Agree` over leads that are not a shortlist `Words` property and not the dated kind's date lead is under `cfg.WeakScore` | AC-2.3 |
| T-7 | Go | `core/database/graphcache/graph_test.go` `TestMatchIndexDrops` *(new)* | Store a value; `SetAliasTables(nil)` → `MatchIndex()` false; store; commit naming a handle → false; store; `Drop` → false; store; vocabulary-only commit → still true | AC-3.4, D-10 |
| T-8 | Go | `core/database/promotealign/promotealign_test.go` `TestProposeSeesAHandlePromotedAfterTheIndexWasBuilt` *(new)* | AC-3.1 | AC-3.1, D-11 |
| T-9 | Go | `core/database/promotealign/promotealign_test.go` `TestProposeQueryCountDoesNotGrowWithTheLayer` and every existing test in `core/database/promotealign` and `api/ffi/handlers` | Unchanged → pass with no expected value edited | AC-3.2, AC-3.3 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `core/graphalign` goldens (they do not touch seeding, so a change there is a stop condition), `TestProposeQueryCountDoesNotGrowWithTheLayer`, `TestKnobInventoryMatchesRegistry`.

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left — none in this PR
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New — no UI in this PR
- [ ] Version bump as the header says — none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated — not gated
- [ ] Domain docs updated (D-13): matching-model.md, comparison-layers.md, promote-graph-alignment.md, add-match-rule skill
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

1. Open a dogfood project with at least two promoted Sources. Promote a third Source that shares a person by name with one of them. The person proposes the shared handle as before.
2. Decide that row, Apply, and promote a fourth Source naming the same person. It proposes the handle Apply just filed (the index was rebuilt after the commit).

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| A key misses a pair the scan found, and a golden moves | Goldens do not seed from the catalog; the catalog suites must pass unchanged (AC-3.3); T-2 and T-4 hold the guarantee |
| The cold build profiles every handle, as the S11-13 scan does on each proposal | Built once per handle commit; S11-21 measures it |
| A wide date's `any` key is the whole dated type | Only open-ended probes read it; closed dates read their own buckets |
| S11-14 lands first and changes `stats.go` | D-9 covers both orders |

**Stop and report if:**

- STEP-0 finds an *(S11-13)* row false (`Candidates` signature, `profiler.profile` / `values`, the seed loop calling `Candidates` per Subject).
- Any expected value in `core/database/promotealign`, `api/ffi/handlers` or `core/graphalign` tests changes.
- T-2 or T-4 finds a `Same` / `Equivalent` (or near-date) pair with no shared key under D-2 / D-5 as written.
- T-6 fails under `matchregistry.Default()` (a kind could reach the weak band with no key, so the index would drop real candidates).
- `core/shortlist` needs an import under `core/database` or `core/matchregistry`.
- A `g.Kind` read is needed in seeding other than in `loadShortlist`.

## 12. Open questions

None. The status is Blocked only on S11-13 landing.

## 13. PR

**Title:** S11-15 — Shortlist index
**Body:** objective in one sentence; what shipped against US-1 – US-3; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-11/prs/S11-15-shortlist-index.md`.
