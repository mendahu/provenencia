# S11-15 — Shortlist index

**Status:** Blocked — on review-plan round 2 (refreshed on S11-13's code)
**Spike / slice:** Spike 11 · Slice 4 — Segments and graph alignment
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no new chrome: an engine index; the wire, the page and its copy are unchanged)
**Board:** —
**Depends on:** S11-13 (landed [#487](https://github.com/mendahu/provenencia/pull/487); [plan](S11-13-promote-subject-compare.md) and its **Landed** note, which changes nothing this plan reads; its [ship log](S11-13-promote-subject-compare.ship.md) PRF-1, the user's call that this PR loads life events only for shortlisted handles). S11-14 is not landed at the planning base (`loadStats` and `graphalign.Stats` exist); STEP-0 checks again and D-9 says what each order means. Branch from `main`
**Blocks:** S11-16
**Branch:** `s11-15-shortlist-index`
**Version bump:** none — Spike 11 PRs leave `VERSION` at 0.2.0, and the product bump is cut when a release ships ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-match-rule`](../../../../.agents/skills/add-match-rule/SKILL.md) ("Add a knob", "Add a layer package", "Add a comparer")
**Rules that bind:** `.cursor/rules/go-tests.mdc`, `.cursor/rules/match-rules.mdc`
**Context:** [Layer 6 procedure](../comparison-layers.md#layer-6-procedure) step 1; [Vocabulary: finding candidates](../comparison-layers.md#vocabulary-finding-candidates); [Decided on 2026-10-17](../comparison-layers.md#decided-on-2026-10-17) (misspellings: start simple); [Match registry](../comparison-layers.md#match-registry); [`docs/matching-model.md`](../../../matching-model.md) glossary and knob inventory (`alignment.Shortlist.*`); working rules "Pure means the package builds and tests with no build tags", "A normalized `Key` per value type", "One match registry", "One guard, not one per package", "One matching skill, grown with the layers", "Registry numbers: carry over, then estimate"; [S11-01 ship log](S11-01-matching-model.ship.md) follow-up PRF-1 (the event shortlist bucket grows with the catalog)
**Planned:** 2026-10-19 against `main` @ `6191fc8` (S11-13 merged); refreshed from the draft of the same day against `6ecfd88`
**Review:** [`S11-15-shortlist-index.review.md`](S11-15-shortlist-index.review.md) — round 1 (draft): Fits with changes / Not ready, findings answered; round 2: pending

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

Promote stops comparing a Source Subject against every handle of its kind. A pure index maps shortlist keys to handles, per kind: a person's name words (not initials), an event's type with the year buckets its date can still be near, a place's toponym words. Layer 1 makes every key, so shortlisting and comparison agree on "the same", and layer 3's reader picks the values, so a period counts as the `date` it stands for. Seeding looks a Subject's keys up and scores only the handles that come back. The index is built once per catalog state, kept in `graphcache`'s stats slot (renamed the match index), and dropped with a handle commit, as the stats are today.

**Why now.** Since S11-13, `loadCanon` profiles every handle of a kind, every person with the events they take part in, and runs `graphalign.Candidates` over all of them on each proposal: the work grows with the catalog, not the Source (dogfood bar 3, "fast"). S11-13's review measured it on the 20k-person `bulkCanon` fixture: a warm repropose went from 40 to 160 ms and 49 to 181 MB (ship log PRF-1), and the user deferred the fix to this PR with the note that life events load only for shortlisted handles, plus the fixed and BFS-reached ones. S11-16's layer 6 walk starts from a shortlist ("shortlist → score, keep top K"); this PR gives it one. PRF-1 asked the event bucket to be bounded: D-5 bounds it by year.

**What changes for the researcher** (the design, not regressions):
- A handle that shares no shortlist key with the Subject is not seeded, even when S11-13's full scan would have reached the weak band: a name whose only likeness is a misspelling ("Robnis" against "Robins", no shared word), an initials-only name against a full one ("J. R." against "James Robins"), a toponym that differs only by a misspelling. This is the 2026-10-17 decision "accept the gap; links catch most of it". The walk still reaches such a handle through a mapped neighbor, and the fallback still scores every canon handle (unchanged, S11-16).
- Nothing else: every handle that shares a key is scored by the same `Candidates` rule as in S11-13.

**What stays.** `graphalign.Candidates` and its gate (S11-13 D-4), `fallbackUnreachable` over the canon handles, the walk, refinement, bands, the wire, Swift. The stats (`Stats.FanOut`) stay until S11-14 deletes them (D-9).

## 2. User stories

Conventions. `name(f)` is `propertycomparetest.Name(f)` (`namevaluetest.Western(f)`), `text(s)` is `propertycomparetest.Text(s)`, `term(k)` is `propertycomparetest.Term(k)`, `date(y)` is `propertycomparetest.Date(datevaluetest.YMD(y, 0, 0))`. `bef(y)` and `aft(y)` are test-local helpers *(new)* (in `core/shortlist/shortlist_test.go` and `core/valuecompare/shortlist_test.go`): `v := datevaluetest.YMD(y, 0, 0); v.Qualifier = datevalue.QualifierBEF` (or `QualifierAFT`); `return v` — the shape of `datevaluetest.About`. `R` is `matchregistry.Default()`. "Lookup" is `(*shortlist.Index).Lookup` on an index built with `shortlist.NewIndex(R.Alignment.Shortlist, R.Subject, R.Value, subjectcomparetest.Metas())`. Alias table `bill`: `valuecompare.AliasTable{Key: "bill", Kind: valuecompare.AliasNickname, Roles: []valuecompare.NameRole{valuecompare.RoleGiven}, Sets: [][]string{{"bill", "william"}}}`, put in rules with `rules.Names.Aliases = valuecompare.NewAliasIndex([]valuecompare.AliasTable{bill}, rules.Words)` on a copy of `R.Value`.

**US-1.** As the engine, I find the handles worth comparing by their shortlist keys, so a Subject is scored against a few handles, not the whole kind.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | Persons indexed: PER-1 `{name: name("James Robins")}`, PER-2 `{name: name("Robins")}`, PER-3 `{name: name("Mary Robins")}`, PER-4 `{name: name("James Robbins")}`, PER-5 `{name: name("Ann Ames")}`. Lookup `{name: name("J. Robins")}` | PER-1, PER-2, PER-3 (every handle with `robins`); not PER-4, not PER-5 |
| AC-1.2 | Under alias table `bill`: index PER-1 `{name: name("William")}`, PER-2 `{name: name("Robert")}`. Lookup `{name: name("Bill")}` | PER-1 only (the alias set is a key) |
| AC-1.3 | Index PER-1 `{name: name("José Robins")}`. Lookup `{name: name("Jose Robins")}` | PER-1 (folded word key) |
| AC-1.4 | Places indexed: PLC-1 `{toponym: text("Saint-Jean")}`, PLC-2 `{toponym: text("York")}`. Lookup `{toponym: text("SaintJean")}`; lookup `{toponym: text("New York")}` | First: PLC-1 (the whole-value key; the words break differently). Second: PLC-2 (shared word `york`) |
| AC-1.5 | Events indexed: EVT-1 `{event_type: term(birth), date: date(1850)}`, EVT-2 `{term(birth), date(1852)}`, EVT-3 `{term(birth), date(1870)}`, EVT-4 `{term(death), date(1850)}`, EVT-5 `{term(birth)}` (undated). Lookup `{event_type: term(birth), date: date(1850)}` | EVT-1 and EVT-2; not EVT-3, EVT-4, EVT-5 |
| AC-1.6 | AC-1.5's index plus EVT-6 `{term(birth), date: bef(1800)}`. Lookup `{term(birth), date(1700)}`; lookup `{term(birth), date: aft(1860)}` | First: EVT-6 only (an open date is `wide`, looked up by every dated birth). Second: EVT-1, EVT-2, EVT-3, EVT-6 (an open probe looks up `any`); never EVT-4 or EVT-5 |
| AC-1.7 | Lookup `{sex_at_birth: term(male)}` on AC-1.1's index; lookup `{event_type: term(birth)}` on AC-1.5's index; lookup on a kind with no rules (`vocabkey.SubjectType("source")`); lookup on a nil `*Index` | Nothing, each time |
| AC-1.8 | Events indexed: EVT-7 `{term(residence), start_date: date(1840), end_date: date(1870)}`, EVT-8 `{term(residence), start_date: date(1850)}` (no end). Lookup `{term(residence), date: date(1855)}`; lookup `{term(residence), date: date(1900)}` | First: EVT-7 and EVT-8 (the period 1840–1870 is one range, as `subjectcompare.Side` reads it; a period with no end is open). Second: EVT-8 only |
| AC-1.9 | `{event_type: term(residence), date: date(1855), start_date: date(1700), end_date: date(1710)}` indexed; lookup `{term(residence), date: date(1705)}` | Not found (a positive `date` hides the period, as in `Side`) |

**US-2.** As a contributor, shortlisting never misses a pair value compare relates.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | Every ordered pair of T-2's name corpus, as name values and as text values, under alias table `bill` | Whenever `valuecompare.Compare` says `Same` or `Equivalent`, `valuecompare.WordKeys` of the two share a key |
| AC-2.2 | Every ordered pair of T-4's date corpus | Whenever the date relation is neither `Different` nor `Unknown`, the two `valuecompare.YearReach` results overlap, or one is open |
| AC-2.3 | Each kind in `R.Subject.Weights` | The most evidence a pair of that kind can earn with no shortlist key is under `graphalign.DefaultConfig().WeakScore` (person without `name`: 0.6 + 0.1; event without `date`: 0.5; place without `toponym`: 0) |

**US-3.** As the engine (the Promote loader), I build the index once per catalog state and seed from it.

Catalog tests in `core/database/promotealign` with the existing `fixture`.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | Sources A and B each hold person "Ann Ames". Propose B (nothing promoted: index built, empty); promote A's person; propose B again | Second proposal: B's person targets A's person handle (the handle commit dropped the index) |
| AC-3.2 | `TestProposeQueryCountDoesNotGrowWithTheLayer`, and the same test run a second time on the warm catalog (T-9) | 2 and 6 people take the same number of queries, cold and warm |
| AC-3.3 | Every existing `core/database/promotealign` and `api/ffi/handlers` Promote test | Passes with no expected value changed |
| AC-3.5 | Cold proposal of a Source whose one person shares no key with the catalog's people | The proposal reads no participation of those people: the query count does not grow with them (T-12) |
| AC-3.6 | Warm repropose on the 20k-person `bulkCanon` fixture | The PR reports ns/op and heap on `main` and on the branch (T-13); S11-21 sets the budget |
| AC-3.4 | `graphcache` with a stored match index; `SetAliasTables`; a commit that names a handle; `Drop`; a vocabulary-only commit | The first three leave `MatchIndex()` reporting false; the vocabulary-only commit leaves it true |

## 3. Scope

| In | Covered by |
| --- | --- |
| `valuecompare.WordKeys` through a `wordKeyer` on the name and text comparers; alias set keys; `valuecompare.YearReach` (D-2, D-3, D-4) | STEP-1, AC-2.1, AC-2.2 |
| `subjectcompare.Side` and `propertycompare.Typed` exported (D-5) | STEP-2 |
| `core/shortlist` *(new)*: `Rules`, `Dated`, `Keys`, `Index` (D-1, D-6, D-7, D-8) | STEP-3, US-1 |
| Registry `Alignment.Shortlist` section, inventory rows, guard row; `graphalign.Config.Shortlist` (D-10, D-11) | STEP-4, AC-2.3 |
| `graphcache` stats slot renamed the match index slot; dropped by `SetAliasTables` too (D-9, D-12) | STEP-5, AC-3.4 |
| Promote seeding through the index (D-13, D-14) | STEP-6, US-3 |
| Docs and the skill the index makes false or leaves silent (D-15) | STEP-7 |

| Out | Home |
| --- | --- |
| Shortlisting inside the walk's fallback and the one queue; top K with ties; similarity as the gate | S11-16 |
| Fan-out and `graphalign.Stats` removal | S11-14 |
| A coarse misspelling key (surname consonant skeleton) | Not planned — [decided 2026-10-17](../comparison-layers.md#decided-on-2026-10-17): only if the sparse-data tier shows a problem |
| Updating the index in place on a commit instead of rebuilding it (`Index.Remove` plus re-`Add` of dirty handles) | S11-21 (only if the 100-Subject check fails) |
| A benchmark for seeding time | S11-21 |
| A writer of the catalog's default name format, and its index drop | Not planned — `nameformats.Install` at onboarding is the only writer; D-12 leaves a doc line for the first one |
| Shortlisting for search, the reconciler or merge hints | Not planned — no caller asks for it |
| One shared `floorDiv` (`core/database/connect/connect.go` has one; a pure package cannot import it) | Not planned — logged in the review (M-11) |

## 4. Current state

Verified at `6191fc8`.

| Area | Where | Today |
| --- | --- | --- |
| Comparers | `core/valuecompare/valuecompare.go` `comparer` interface (`compare`, `key`), `comparers` map, `Key(v, level, rules)` (dispatches through `comparers`) | One key per value |
| Value keys | `core/valuecompare/name.go` `name.key`, `nameWords` (ignored parts dropped, generation words returned apart), `nameKey`; `text.go` `text.key`, `joinedKey`, `TextKey` | Name keys ignore alias tables because sets do not chain. No per-word key exists |
| Words | `core/valuecompare/words.go` `SplitWords`, `FoldWord`, `IsInitial` (one letter) | The one word helpers |
| Aliases | `core/valuecompare/aliases.go` `AliasIndex{sets []aliasSet; byToken}`, `aliasSet{tokens, roles, kind}`, `AliasToken(token, w) (string, bool)`, `NewAliasIndex`, `Hit`; `rules.go` `NameRules.Aliases` (`r.Names.Aliases`) | No way to ask which sets hold a token. Text compare never reads aliases |
| Dates | `core/valuecompare/date.go` `readDate` (through `AboutEnds`), `compare` (a gap of more than `Tolerance` civil years between `later.first.Year` and `earlier.last.Year` is `Different`), `date.key`; `datewindow.go` `DateEnd{Open, Year, …}`, `ShiftYears` | |
| Subject sides | `core/subjectcompare/pairs.go` `side(kind, lead, values, rules)` (the lead's values, its pooled values, or the period from `start` / `end` when no positive lead value), `period` | Unexported |
| Value typing | `core/propertycompare/propertycompare.go` `split` (`v.Type = meta.ValueType`), `Value{valuecompare.Value; Negative; Neutral}`, `Positive()` | Profile values carry no `Type` (S11-13 D-12 `subjectprofile.Value` doc) |
| Registry | `core/matchregistry/registry.go` `Registry{Value, Subject}`, `Default()`, package doc "Layers declare their section's types"; `subject.go` `Periods[event][date] = {start_date, end_date}`; `value.go` `Dates.Tolerance` 2, `Dates.AboutUnits` 1 | No alignment section |
| Inventory guard | `core/matchregistry/inventory_test.go` `notYetBuilt` (`"segment."`, `"alignment."`, `"evidence."`), `TestKnobInventoryMatchesRegistry`, `registryKeys` | |
| Purity guard | `core/matchregistry/guard_test.go` `layerPackages` | No `core/shortlist` row |
| Knob inventory | `docs/matching-model.md` §Knob inventory rows `alignment.Shortlist.YearBucket` (2), `alignment.Shortlist.person`, `.event`, `.place`; glossary rows "shortlisting", "shortlist key", "shortlisted handle" | Written in S11-01 before the shape existed; D-11 replaces the three kind rows |
| Walk config | `core/graphalign/registry.go` `DefaultConfig` (the package's only import of `matchregistry`), `Config` (`WeakScore` 2); `Config.Subject` | |
| Candidates | `core/graphalign/candidates.go` `Candidates(kind, values, handles, metas, cfg, limit)`; its doc says "Shortlisting replaces the scan (S11-15)" | |
| Fallback | `core/graphalign/align.go` `fallbackUnreachable` over `st.handlesOfKind` | Unchanged by this PR |
| Seeding | `core/database/promotealign/canon.go` `loadCanon(g, layer, fixed, cfg, namePattern)`: `newProfiler`; for each unfixed Subject of a kind in `cfg.Subject.Weights`, `kindHandles[s.Kind]` is built once per kind from `g.Kind(types[s.Kind])`, `prof.profile(nodes)`, and a `graphalign.Handle{ID, Ref, Kind, Values: prof.values(n.ID)}` per unmerged node; then `graphalign.Candidates(s.Kind, s.Values, hs, layer.Metas, cfg, seedCandidateLimit)` and `add(h.ID)` | The scan this PR replaces |
| Profiler | `core/database/promotealign/profiles.go` `profiler{g, props, terms, metas, types, rules, reconcile, namePattern, done}`, `newProfiler`, `kind(n)`, `profile(nodes)` (for a kind with a lifespan: `g.Follow(n.ID, hopPersonToEvent)` per node, one `g.Nodes` for the events, events built first), `build(n)` (reads `n.Values` into `subjectprofile.Property` per product property, adds `Life` from built events, then `subjectprofile.Build(e, p.metas, p.reconcile, p.rules.Lifespans[kind], subjectprofile.Options{})`), `values(id)`; `canon.go` `typeByKey`, `typeKey` | A person's profile always loads and builds their events |
| Stats slot | `core/database/graphcache/graph.go` `Graph.stats`, `statsOK`, `Stats()`, `SetStats`, `ResetStatsForTest`, `clear`; `commit.go` (a commit naming handles clears `statsOK` and `stats`); `SetAliasTables` (does not drop it) | `any`-typed slot |
| Benchmark | `core/database/promotealign/bench_test.go` `BenchmarkProposeColdHub`, `bulkCanon` (20k persons, 30k events, each person the subject of a birth) | Cold only; S11-13's reviewer timed warm reproposes on the same fixture |
| Stats use | `core/database/promotealign/stats.go` `loadStats` (`cached.(graphalign.Stats)`), `cloneStats`, `computeStats`; `export_test.go` (wraps `loadStats`); `promotealign.go` `Propose` calls `loadStats` | Only caller of the slot. S11-14 deletes fan-out (map review K-10) |
| Kind index | `core/database/graphcache/graph.go` `kinds`, `kindSet`, `Kind`; `load.go` patches them in place on a commit | The in-place precedent S11-21 would follow |
| Search | `core/database/searchindex` (`catalog_search_fts`, `catalog_search_fts_trigram`); `core/search/fts.go` `mergeFuzzyShortlist` | A token → entity map in SQL with its own normalization, kept apart on purpose ([reuse inventory](../comparison-layers.md#reuse-across-the-app), "Search's fuzzy ranker… stays") |
| Aliases warm | `core/database/namealiases/namealiases.go` `Warm` → `g.SetAliasTables`; `Invalidate` | Tables change only through `Warm` |
| Name format | `core/database/nameformats` package doc; `registry.go` `Install` (onboarding, the only writer of the default format); `api/ffi/handlers/nameformat.go` (`GetNameFormat` only) | |
| Tests | `core/valuecompare/contract_test.go` `TestKeyContract`; `core/database/promotealign/promotealign_test.go` `TestProposeQueryCountDoesNotGrowWithTheLayer`, fixture `newFixture`, `person`, `promote`, `propose`; `core/database/graphcache/graph_test.go` | |

**Precedent.** S11-02 ([#441](https://github.com/mendahu/provenencia/pull/441), [plan](S11-02-value-compare-core.md)) added a layer package, its registry section, its guard row and inventory rows together: copy that shape for `core/shortlist`. S11-07 ([#460](https://github.com/mendahu/provenencia/pull/460), [plan](S11-07-normalization-key.md)) put one normalization rule in `valuecompare` and had every other caller read it: `WordKeys` and `YearReach` follow it. The stats slot (S9-42, `loadStats`) is the cache pattern to keep: build on a miss, store, drop on a handle commit.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | **A new pure layer package, `core/shortlist`.** It declares `Rules`, `Dated`, `Keys`, `Index`, `NewIndex`, `Add`, `Lookup`. It imports `core/valuecompare`, `core/propertycompare`, `core/propertymeta`, `core/subjectcompare`, `core/vocabkey` and nothing under `core/database` or `core/matchregistry`. Package doc: "Package shortlist is layer 6's first step: the keys a Subject is looked up by and the index that maps them to handles. Layer 1 makes every key and layer 3 picks the values, so shortlisting reads a Subject as comparison does. It is pure." | `matchregistry` must name the section's types, and `graphalign` imports `matchregistry` (`DefaultConfig`), so the types cannot live in `graphalign` yet. The index is shared by the loader and S11-16's walk | Types in `graphalign` — an import cycle until S11-16 deletes `registry.go`; the index in `promotealign` — S11-16's walk is pure and could not read it |
| D-2 | **Word keys are layer 1's, per comparer.** An optional interface `wordKeyer { wordKeys(v Value, r Rules) []string }` *(new, `core/valuecompare/shortlist.go` (new))*, implemented by `name` and `text`. `WordKeys(v Value, r Rules) []string` *(new)* looks `v.Type` up in `comparers` and calls `wordKeys` when the comparer implements it; otherwise nil. Every implementation returns `"=" + k` for `k, ok := Key(v, Equivalent, r)` when ok, and `"w:" + FoldWord(word, r.Words.Folds)` for each word that is not `IsInitial`. **Name** reads words through `nameWords` (ignored parts dropped; generation words never keyed) and adds `"a:" + s` for each `s` in `r.Names.Aliases.SetKeys(t)`, where `t, ok := AliasToken(word, r.Words)`; a word with no token adds no alias key. **Text** adds no alias key (text compare never reads aliases). Sorted, no duplicates; a value with no word returns nil. Doc on `WordKeys`: "WordKeys are the keys shortlisting looks a value up by. Two values Compare calls Same or Equivalent share one: the whole key at Equivalent, or, for an alias, the set. Values that share a word that is not an initial share one too. Initials and generation words are never keys alone." | Layer 1 owns "the same" (working rule "A normalized `Key` per value type"); a value type lists its behaviors in one place, beside `key`; the guarantee is tested where `Compare` lives | A free function with a type switch — a second per-type list beside `comparers`; keys built in `core/shortlist` from `SplitWords` — ignores the pattern's roles; the whole `Key` alone — "J. Robins" would not find "James Robins" |
| D-3 | **Alias sets as keys.** `aliasSet` gains `key string`: its tokens sorted and joined by one space, set in `NewAliasIndex`. `func (a AliasIndex) SetKeys(token string) []string` *(new)*: the `key` of every set holding `token` (an `AliasToken`), roles ignored, sorted, no duplicates; nil for the zero index | Sets do not chain, so no word key can join them, but a set can be a key: two words that `Hit` share it. Content keys survive a reloaded table; roles are ignored because a shortlist may over-include | Set positions as keys — they change when tables reload |
| D-4 | **The year reach is layer 1's.** `valuecompare.YearReach(v Value, r Rules) (lo, hi int, open, ok bool)` *(new, `date.go`)*: ok is false when `v.Date` is nil or `readDate` cannot place it; open when either end `readDate` returns is open; otherwise `lo` = `first.Year − r.Dates.Tolerance`, `hi` = `last.Year + r.Dates.Tolerance`. Doc: "YearReach is the civil years a date can still be Near to. Two dates Compare relates other than Different or Unknown have overlapping reaches, or one is open." | The Near rule lives in `date.go` (`readDate`, the tolerance gap); a change to it changes the reach in the same file; the skill says not to move years by hand outside it | Padding by `Dates.Tolerance` in `core/shortlist` — a second copy of the Near rule |
| D-5 | **Two exports for one reading.** `subjectcompare.side` is renamed `Side` (doc: "Side is the values Compare reads for lead on one Subject: the lead's own, its pooled properties', or, when the lead has a period and no positive value, the period built from its start and end. It copies.") and every caller in `subjectcompare` is updated. `propertycompare.Typed(v Value, meta propertymeta.Meta) Value` *(new)*: `v` with `Type` = `meta.ValueType`; `split` uses it | Shortlisting must read a Subject as `Compare` does: a period is one range (the 1840–1870 residence contains 1855); a pooled property counts with its lead. One reader, one typing rule | Listing `start_date` and `end_date` in the shortlist rules — a third hand-written copy of `Periods`, and it misses a period's middle |
| D-6 | **Shortlist rules.** `shortlist.Rules{Words map[vocabkey.SubjectType][]vocabkey.PropertyKey; Dated map[vocabkey.SubjectType]Dated; YearBucket int}`, `Dated{Type, Date vocabkey.PropertyKey}`. Field docs: Words "the leads whose word keys key a Subject of the kind"; Dated "the kinds keyed by a type term with the year buckets of a date"; Type "the term lead keyed with each bucket"; Date "the date lead whose reach is bucketed; its period comes from the subject section"; YearBucket "years per bucket". Every named property is a lead with weights in the subject section (T-5) | Typed keys (registry rule); periods and pools stay in one place, the subject section | One struct per kind with every field — empty fields on every kind in the inventory |
| D-7 | **Keys.** `shortlist.Keys(kind vocabkey.SubjectType, v subjectcompare.Values, metas subjectcompare.Metas, r Rules, subject subjectcompare.Rules, value valuecompare.Rules) (index, lookup []string)`. Every value is read with `subjectcompare.Side(kind, lead, v, subject)`, kept when `Positive()`, and typed with `propertycompare.Typed(x, metas[lead])`. **Word leads:** for each `r.Words[kind]` lead, `string(lead) + "|" + k` for each `valuecompare.WordKeys` key `k`; `index` and `lookup` are the same list. **Dated kinds:** for each `Dated.Type` value with `typeKey, ok := valuecompare.Key(t.Value, valuecompare.Equivalent, value)` and each `Dated.Date` value with `lo, hi, open, ok := valuecompare.YearReach(d.Value, value)`: a closed reach adds `"t:" + typeKey + "|" + strconv.Itoa(b)` for every `b` from `floorDiv(lo, r.YearBucket)` to `floorDiv(hi, r.YearBucket)` to both lists, `"t:" + typeKey + "|any"` to `index`, and `"t:" + typeKey + "|wide"` to `lookup`; an open reach adds `"t:" + typeKey + "|wide"` and `"t:" + typeKey + "|any"` to `index` and `"t:" + typeKey + "|any"` to `lookup`. A date with no reach, an event with no date, and `YearBucket < 1` key nothing. Both lists sorted, no duplicates. `floorDiv(y, w int) int` rounds toward minus infinity: `q := y / w; if y%w != 0 && y < 0 { q-- }; return q` | Overlapping reaches share a bucket; `wide` / `any` cover open dates without making every probe scan its type; the event bucket is bounded by type and year (PRF-1). An undated event cannot reach the weak band (AC-2.3) | Probing neighbor buckets — the same reach, with two sides that must agree on a direction; one key per year — a wide range adds a key per year |
| D-8 | **Index.** `type Index struct` *(new)*: unexported `rules Rules`, `subject subjectcompare.Rules`, `value valuecompare.Rules`, `metas subjectcompare.Metas`, `byKey map[vocabkey.SubjectType]map[string][][]byte`. `NewIndex(r Rules, subject subjectcompare.Rules, value valuecompare.Rules, metas subjectcompare.Metas) *Index`. `(*Index).Add(kind vocabkey.SubjectType, id []byte, v subjectcompare.Values)` files a copy of `id` under each `index` key. `(*Index).Lookup(kind vocabkey.SubjectType, v subjectcompare.Values) [][]byte`: the ids under any `lookup` key, deduplicated, sorted by `bytes.Compare`, each a copy; nil for a kind with no entry and for a nil `*Index`. Read-only once built | The index keeps the rules it was built with, so a lookup never keys differently from the build | Passing rules to every call — a caller could mix two rule sets |
| D-9 | **One slot, either landing order.** In `graphcache`, `stats` / `statsOK` / `Stats()` / `SetStats` / `ResetStatsForTest` are renamed `matchIndex` / `matchIndexOK` / `MatchIndex()` / `SetMatchIndex` / `ResetMatchIndexForTest`, same drop points; docs say "the match index: what Promote derives from the whole catalog". `promotealign` stores one value there: `matchIndex{stats *graphalign.Stats; shortlist *shortlist.Index}` *(new, in `index.go`, which `stats.go` is renamed to)*. A nil field is not built yet. `matchIndexOf(g *graphcache.Graph) matchIndex` *(new)* returns the stored value or the zero value; a loader reads it, fills its own field, and stores the whole value back with `g.SetMatchIndex` (catalogsession serializes, so read-modify-write is safe). `loadStats` fills `stats`; `loadShortlist` fills `shortlist`. **If S11-14 landed first** (STEP-0): with `graphalign.Stats` and `loadStats` gone, `matchIndex` has only `shortlist`; if S11-14 also deleted the slot, STEP-5 adds `MatchIndex` / `SetMatchIndex` / `ResetMatchIndexForTest` with the drop points the stats slot had (handle commit, `Drop`, `clear`) | The map gives S11-14 fan-out and S11-15 the slot (review K-10); one value holds what both need while they overlap; pointers tell "not built" from zero | Two slots — a second drop rule to keep in step; a value `Stats` field — zero fan-out reads as built |
| D-10 | **Registry section.** `core/matchregistry/alignment.go` *(new)*: `type Alignment struct{ Shortlist shortlist.Rules }`, doc "Alignment is layer 6's section. It is declared here only until S11-16 deletes `core/graphalign/registry.go`, which is what keeps `graphalign` from declaring it; S11-16 replaces it with a layer-owned type and adds the walk's knobs." `defaultAlignment()`: `Words` person → `[name]`, place → `[toponym]` (comment `// decided: comparison-layers layer 6 procedure`); `Dated` event → `{Type: event_type, Date: date}` (same comment); `YearBucket: 2` (comment `// estimate: the date tolerance, so a Near date shares a bucket`). `Registry` gains `Alignment Alignment`, set in `Default()`. `graphalign.Config` gains `Shortlist shortlist.Rules` (after `Subject`, doc "Shortlist is the registry's shortlist keys. DefaultConfig holds the registry's."), set in `DefaultConfig` from `matchregistry.Default().Alignment.Shortlist` | One registry (working rule); `Config` already carries `Value` and `Subject` the same way; S11-16 inherits a named slot | A `Registry.Shortlist` field — the design puts shortlist keys in the alignment section, which S11-16 grows |
| D-11 | **Inventory.** In `docs/matching-model.md`, the rows `alignment.Shortlist.person`, `.event`, `.place` become `alignment.Shortlist.Words` (Start: person `name`, place `toponym`; Source: decided ([`comparison-layers.md`](../comparison-layers.md#layer-6-procedure)); Changes: "Leads whose words key a Subject"), `alignment.Shortlist.Dated.event.Type` (`event_type`; same source; "The term keyed with each year bucket"), `alignment.Shortlist.Dated.event.Date` (`date`; same source; "The date whose reach is bucketed; a period stands for it"); `alignment.Shortlist.YearBucket` stays. In `inventory_test.go`, `notYetBuilt`'s `"alignment."` becomes `"alignment.AcceptBar"`, `"alignment.Bands."`, `"alignment.TopK"`, `"alignment.TieCap"`, `"alignment.Alternatives"`, `"alignment.SettleMargin"`, `"alignment.SettleRounds"`. `guard_test.go` gains `{"core/shortlist", true}`; `.cursor/rules/match-rules.mdc` gains `core/shortlist/**` beside the other layer globs | The guard needs a row per field and a prefix only for what has not landed | Keeping `"alignment."` — it matches the new field and fails the guard |
| D-12 | **One staleness rule: drop on change.** The match index slot is dropped where stats are today (a commit naming handles, `Drop`) and also by `SetAliasTables`, which changes the alias sets `WordKeys` reads. Nothing else is compared. The `nameformats` package doc gains one sentence: "A future writer of the catalog's default name format must drop the graph's match index: shortlist keys read its part roles." | Every input the keys read changes only through a path that drops the slot; the one input with no runtime writer gets a line where its first writer will look | A stored name pattern compared on each proposal — guards a write that does not exist and misses a role change under the same key |
| D-13 | **Build once, then look up in one pass.** `loadShortlist(g *graphcache.Graph, p *profiler, cfg graphalign.Config) (*shortlist.Index, error)` *(new, `index.go`)*: `matchIndexOf(g).shortlist` when non-nil; otherwise `shortlist.NewIndex(cfg.Shortlist, cfg.Subject, cfg.Value, p.metas)`, and for each kind in `cfg.Shortlist.Words` or `cfg.Shortlist.Dated`, in sorted order, whose `p.types` entry is 16 bytes, every unmerged `g.Kind` node is `Add`ed with `p.keyValues(n)`; the result is stored (D-9). Metas and types come from the profiler only. **Keys need no life** (no shortlist lead is derived, T-6), so the build loads and builds no life event: `profiler.entity(n *graphcache.Node) (subjectprofile.Entity, error)` *(new)* is `build`'s property loop moved out (no `Life`); `build` calls it and adds `Life` as today; `profiler.keyValues(n *graphcache.Node) (subjectcompare.Values, error)` *(new)* is `subjectprofile.Build(entity, p.metas, p.reconcile, nil, subjectprofile.Options{})`, not stored in `done` (a `done` entry is a full profile). `loadCanon`'s seed step: one pass collects `Lookup(kind, s.Values)` for every unfixed Subject of a kind with `cfg.Subject.Weights`; one `g.Nodes` call loads the union; one `p.profile` call profiles the present, unmerged nodes; then each Subject gets `graphalign.Candidates(s.Kind, s.Values, handles, layer.Metas, cfg, seedCandidateLimit)` over its own looked-up ids, each a `graphalign.Handle{ID, Ref: n.Ref, Kind: s.Kind, Values: p.values(id)}`, sorted by `Ref` then `ID` bytes. An id with no node, or a merged node, is skipped | Full profiles, life events included, are built only for shortlisted, fixed and BFS-reached handles (the user's PRF-1 note); warm proposals do one node load and one profile batch however many Subjects; the cold build reads every node of the indexed kinds once per handle commit, with no life | Indexing from raw rows — keys must come from the displayed values the profile holds; a load per Subject — grows with the layer |
| D-14 | **Candidates keeps its signature.** Its doc's "Shortlisting replaces the scan (S11-15)" becomes "Seeding passes it shortlisted handles (S11-15); top K replaces the limit (S11-16)." `fallbackUnreachable` keeps calling it with `st.handlesOfKind` | The fallback reads only the canon set, which the seeds and the BFS bound; S11-16 rewrites it | Shortlisting the fallback now — S11-16 deletes it |
| D-15 | **Docs and skill.** `docs/matching-model.md`: D-11's rows; glossary "shortlist key" gains "made by layer 1 (`valuecompare.WordKeys`, `YearReach`) and indexed by `core/shortlist`". `comparison-layers.md` layer 6 step 1: append "**Built in S11-15** (`core/shortlist`): word keys for person names and toponyms, type and year bucket for events; seeding reads it, the walk from S11-16." and change "place, toponym" to "place, toponym words". `docs/promote-graph-alignment.md`: the sentence "Stats are cached on the catalog's graph and dropped by any commit that names a handle." becomes "Stats and the shortlist index are cached on the catalog's graph as its match index and dropped by any commit that names a handle." `.agents/skills/add-match-rule/SKILL.md`: `core/shortlist` added to the description's package list and to the Check command; under "Rules that hold everywhere", "**Shortlist keys come from layer 1.** `WordKeys` and `YearReach` make them; `core/shortlist` only buckets and indexes. A lead whose agree weight alone can reach `WeakScore` needs a shortlist entry (`TestUnkeyedEvidenceStaysUnderTheWeakBand`)."; a new section "Add a shortlist key": "1. Add the lead to `Words` or a kind to `Dated` in `core/matchregistry/alignment.go` *(new)*. 2. Add its inventory row (`alignment.Shortlist.Words`, `alignment.Shortlist.Dated.<kind>.Type` / `.Date`). 3. A value type the lead uses implements `wordKeys` or has a `YearReach`; extend T-2's or T-4's corpus."; "Add a comparer" gains step 6: "A type that shortlists by words implements `wordKeys`; add forms to `TestWordKeysShareOnSameOrEquivalent`." Map: the S11-16 scope entry's "Registry alignment section." becomes "Grows the registry alignment section S11-15 opened, and moves its type into `graphalign`." | Working rule "One matching skill, grown with the layers"; the next agent adding a lead finds the rule before T-6 fails | — |

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `wordKeyer`, `WordKeys`, `name.wordKeys`, `text.wordKeys` *(new)*, `core/valuecompare/shortlist.go` *(new)* | interface, func, method, file | `comparer.key`, `Key`, `TextKey` | `Key` is one key per value; a shortlist needs a key per word to find `Contains` pairs. `WordKeys` includes `Key` and dispatches like it (D-2) |
| `AliasIndex.SetKeys`, `aliasSet.key` *(new)* | method, field | `AliasIndex.Hit` | `Hit` answers "do these two share a set"; an index needs "which sets hold this" (D-3) |
| `valuecompare.YearReach` *(new)* | func | `readDate`, `date.key`, `ShiftYears` | Built on `readDate`; the reach is the Near rule read as years (D-4) |
| `subjectcompare.Side` *(renamed export)*, `propertycompare.Typed` *(new)* | func | `side`, `split`'s typing line | Same code, now shared (D-5) |
| `core/shortlist` *(new)*: `Rules`, `Dated`, `Keys`, `floorDiv`, `Index`, `NewIndex`, `Add`, `Lookup` | package | `graphcache.Graph.Kind` (the scan); `core/database/searchindex` + `core/search/fts.go` `mergeFuzzyShortlist` (token → entity in SQL) | `Kind` lists a whole kind; search keeps its own normalization on purpose and lives in SQL, outside a pure layer (D-1) |
| `matchregistry.Alignment`, `defaultAlignment` *(new)*, `core/matchregistry/alignment.go` *(new)* | type, func, file | `subject.go` `defaultSubject` | The design's alignment file; temporary home for the type (D-10) |
| `graphalign.Config.Shortlist` *(new field)* | field | `Config.Value`, `Config.Subject` | Same pattern (D-10) |
| `graphcache` `MatchIndex`, `SetMatchIndex`, `ResetMatchIndexForTest` *(renamed)* | method | `Stats`, `SetStats`, `ResetStatsForTest` | The same slot, named for what it holds (D-9) |
| `promotealign` `matchIndex`, `matchIndexOf`, `loadShortlist` *(new)*; `stats.go` → `index.go` | type, func, file | `loadStats`, `cloneStats` | One slot for stats and shortlist (D-9, D-13) |
| `TestWordKeys`, `TestWordKeysShareOnSameOrEquivalent`, `TestAliasSetKeys`, `TestYearReach`, `TestYearReachCoversEveryRelatedPair` *(new, `core/valuecompare`)*; `TestKeys`, `TestLookup` *(new, `core/shortlist`)*; `TestShortlistSection` *(new)*; `TestUnkeyedEvidenceStaysUnderTheWeakBand` *(new)*; `TestMatchIndexDrops` *(new)*; `TestProposeSeesAHandlePromotedAfterTheIndexWasBuilt`, `TestProposeQueryCountStaysFlatWhenWarm` *(new)* | test | `TestKeyContract`, `TestSubjectSection`, `TestProposeQueryCountDoesNotGrowWithTheLayer` | Same table shapes |

**Builds on.** The `comparers` dispatch, `Key`, `SplitWords`, `FoldWord`, `IsInitial`, `AliasIndex`, `readDate`; `subjectcompare`'s side reader and the subject section's periods and pools; the registry and its two guards; S11-13's `profiler` and `Candidates`; the graph's stats slot.
**Deletes.** `graphcache` `Stats`, `SetStats`, `ResetStatsForTest` (renamed); `core/database/promotealign/stats.go` (renamed `index.go`); the per-kind `g.Kind` profiling loop in `loadCanon`'s seeding; `split`'s inline typing line (now `Typed`); the `notYetBuilt` prefix `"alignment."`; the inventory rows `alignment.Shortlist.person`, `.event`, `.place`.
**Next adds.** S11-16's walk calls `Lookup` for every Subject and keeps the top K, and moves `Alignment` into `graphalign` once `registry.go` is gone (D-10). A new shortlist lead is one `Words` or `Dated` entry plus an inventory row (skill "Add a shortlist key"). A coarse misspelling key is one more prefix in `name.wordKeys`. S11-14 deletes `matchIndex.stats`. If S11-21's check fails, the additive fix is to record dirty handle ids on a commit instead of dropping, and `Index.Remove` plus re-`Add` only those (ids are already what `byKey` holds).

### Schema

None — no catalog change.

### Go core

| Package | Change |
| --- | --- |
| `core/valuecompare` | `shortlist.go` *(new)*: `wordKeyer`, `WordKeys`, `name.wordKeys`, `text.wordKeys`; `aliases.go`: `aliasSet.key`, `SetKeys`; `date.go`: `YearReach` (D-2 – D-4) |
| `core/subjectcompare`, `core/propertycompare` | `Side` export; `Typed` (D-5) |
| `core/shortlist` *(new)* | `shortlist.go` *(new)*: `Rules`, `Dated`, `Keys`, `floorDiv`; `index.go` *(new)*: `Index`, `NewIndex`, `Add`, `Lookup` (D-1, D-6 – D-8) |
| `core/matchregistry` | `alignment.go` *(new)*; `registry.go` `Registry.Alignment`; `inventory_test.go`, `guard_test.go`, `registry_test.go` (D-10, D-11) |
| `core/graphalign` | `registry.go` `Config.Shortlist`; `candidates.go` doc; `candidates_test.go` T-6 (D-10, D-14) |
| `core/database/graphcache` | `graph.go`, `commit.go`: the slot rename; `SetAliasTables` drops it (D-9, D-12) |
| `core/database/nameformats` | package doc line (D-12) |
| `core/database/promotealign` | `stats.go` → `index.go` *(new)*; `canon.go` seeding; `promotealign.go`; `export_test.go` (D-9, D-13) |

```go
// core/shortlist

type Rules struct {
	Words      map[vocabkey.SubjectType][]vocabkey.PropertyKey
	Dated      map[vocabkey.SubjectType]Dated
	YearBucket int
}

type Dated struct {
	Type, Date vocabkey.PropertyKey
}

func Keys(kind vocabkey.SubjectType, v subjectcompare.Values, metas subjectcompare.Metas, r Rules, subject subjectcompare.Rules, value valuecompare.Rules) (index, lookup []string)

type Index struct{ /* rules, subject, value, metas, byKey */ }

func NewIndex(r Rules, subject subjectcompare.Rules, value valuecompare.Rules, metas subjectcompare.Metas) *Index
func (x *Index) Add(kind vocabkey.SubjectType, id []byte, v subjectcompare.Values)
func (x *Index) Lookup(kind vocabkey.SubjectType, v subjectcompare.Values) [][]byte

// core/valuecompare

type wordKeyer interface{ wordKeys(v Value, r Rules) []string }

func WordKeys(v Value, r Rules) []string
func (a AliasIndex) SetKeys(token string) []string
func YearReach(v Value, r Rules) (lo, hi int, open, ok bool)

// core/subjectcompare, core/propertycompare

func Side(kind vocabkey.SubjectType, lead vocabkey.PropertyKey, values Values, rules Rules) []propertycompare.Value
func Typed(v Value, meta propertymeta.Meta) Value

// core/database/promotealign (unexported)

type matchIndex struct {
	stats     *graphalign.Stats // until S11-14
	shortlist *shortlist.Index
}
```

`core/shortlist` holds no numeric literal other than 0 and 1 (the guard).

### FFI

None — no proto or handler change.

### macOS

None — no Swift change.

### Strings

None — no user-facing copy.

## 7. Steps

**STEP-0 — Which order S11-14 landed in**
Files: none.
Do: run `rg -n 'func loadStats|type Stats struct|func \(g \*Graph\) Stats' core`. Hits mean S11-14 has not landed: use D-9's first shape. No hits mean it has: use D-9's "If S11-14 landed first" shape.
Check: the PR description names the shape used.

**STEP-1 — Keys in value compare**
Files: `core/valuecompare/shortlist.go` *(new)*, `aliases.go`, `date.go`, `core/valuecompare/shortlist_test.go` *(new)*. Skill: `add-match-rule` ("Add a comparer" for the shape).
Do: D-2, D-3, D-4. T-1 – T-4.
Check: `go test ./core/valuecompare/...` passes with no tags; `TestKeyContract` is unchanged.

**STEP-2 — One reader, one typing rule**
Files: `core/subjectcompare/pairs.go`, `core/subjectcompare/subjectcompare.go` (callers of `side`), `core/propertycompare/propertycompare.go`.
Do: D-5. No behavior change.
Check: `go test ./core/subjectcompare/... ./core/propertycompare/...` passes with no test edited.

**STEP-3 — The shortlist package**
Files: `core/shortlist/shortlist.go`, `index.go`, `shortlist_test.go` *(new)* (all three). Skill: `add-match-rule` ("Add a layer package").
Do: D-1, D-6, D-7, D-8. Tests from an external `shortlist_test` package with `propertycomparetest`, `datevaluetest`, `subjectcomparetest.Metas()` and `matchregistry.Default()`. T-5 (the `TestKeys` / `TestLookup` part).
Check: `go test ./core/shortlist/...` passes with no tags.

**STEP-4 — Registry section and config**
Files: `core/matchregistry/alignment.go` *(new)*, `registry.go`, `registry_test.go`, `inventory_test.go`, `guard_test.go`, `core/graphalign/registry.go`, `core/graphalign/candidates.go`, `core/graphalign/candidates_test.go`, `docs/matching-model.md` (inventory rows only), `.cursor/rules/match-rules.mdc`. Skill: `add-match-rule` ("Add a knob").
Do: D-10, D-11, D-14. T-6, T-7.
Check: `go test ./core/matchregistry/... ./core/graphalign/...` passes; `TestKnobInventoryMatchesRegistry` and the purity guard pass with the `core/shortlist` row.

**STEP-5 — The match index slot**
Files: `core/database/graphcache/graph.go`, `commit.go`, `graph_test.go`, `core/database/nameformats` (package doc file), `core/database/promotealign/stats.go` → `index.go`, `export_test.go`, `promotealign.go`.
Do: D-9 (the shape STEP-0 found), D-12. `matchIndex` with `stats` only so far. Rename every caller; afterwards `rg -n 'ResetStatsForTest|SetStats|\.Stats\(\)' --type go` prints nothing. T-8.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/graphcache/ ./core/database/promotealign/` passes.

**STEP-6 — Seed from the index**
Files: `core/database/promotealign/index.go` *(new)* (from STEP-5), `canon.go`, `profiles.go`, `promotealign_test.go`, `bench_test.go`.
Do: D-13 (`entity`, `keyValues`, `loadShortlist`, the seed pass). `matchIndex` gains `shortlist`. T-9, T-10, T-12, T-13; the existing suites stay green (AC-3.3).
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/promotealign/ ./api/ffi/handlers/` passes with no expected value changed; `rg -n 'g\.Kind\(' core/database/promotealign` lists `loadShortlist` and no line in `loadCanon`.

**STEP-7 — Docs**
Files: `docs/matching-model.md` (glossary), `docs/deployment-plan/spike-11/comparison-layers.md`, `docs/promote-graph-alignment.md`, `.agents/skills/add-match-rule/SKILL.md`, `docs/deployment-plan/spike-11/completed.md`, `docs/deployment-plan/spike-11/deployment-plan.md` (map row; S11-16 scope entry), this plan (status).
Do: D-15 and the Definition of done.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-11/prs/S11-15-shortlist-index.md` reports no errors; links resolve.

Commits: one per step from STEP-1, each subject one full imperative sentence ("Look Promote's seeds up in a shortlist index instead of scanning every handle of the kind.").

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/valuecompare/shortlist_test.go` `TestWordKeys` *(new)* | Table (name values `Value{Type: name, Name: namevaluetest.Western(f), NamePattern: valuecompare.WesternPattern}`): "J. Robins" → `"=" + Key(…, Equivalent)` and `"w:robins"` only; "José Robins" → contains `"w:jose"`; "John Smith Jr" → no `"w:jr"`; "Bill" under table `bill` → contains `"a:bill william"`; text "New York" → `"w:new"`, `"w:york"`, no `"a:"` key; integer value → nil; empty name → nil | D-2, AC-1.2, AC-1.3 |
| T-2 | Go | `core/valuecompare/shortlist_test.go` `TestWordKeysShareOnSameOrEquivalent` *(new)* | Corpus "James Robins", "J. Robins", "Robins", "José", "Jose", "JOSE", "Mc Donald", "McDonald", "Saint-Jean", "SaintJean", "Bill", "William", "Robert Jr", "Robert", "Ann-Marie", "Ann Marie", "Ó Briain", "O Briain", "J. R.", "J R", "Mary Robins", "Robnis", "Müller", "Muller" × itself, as name values (`namevaluetest.Western`) and as text, under table `bill`: when `Compare` is `Same` or `Equivalent`, the `WordKeys` share a key; the test asserts it met at least one `Equivalent` alias pair and one `Same` word-break pair | AC-2.1 |
| T-3 | Go | `core/valuecompare/shortlist_test.go` `TestAliasSetKeys` *(new)* | Two tables sharing `chris` (`{chris, christine}`, `{chris, christopher}`): `SetKeys("chris")` → both keys, sorted; `SetKeys("christine")` → one; zero index → nil | D-3 |
| T-4 | Go | `core/valuecompare/shortlist_test.go` `TestYearReach`, `TestYearReachCoversEveryRelatedPair` *(new)* | `TestYearReach`: `YMD(1850, 0, 0)` → 1848, 1852, closed; `About(1850, 0, 0)` → widened by `AboutUnits`; `bef(1850)` → open; nil date → not ok. `TestYearReachCoversEveryRelatedPair`: corpus points 1846–1854, `1850-03`, `1850-03-04`, `ABT 1850`, `bef(1850)`, `aft(1850)`, `Between(1840, 0, 0, 1845, 0, 0)`, `Between(1840, 0, 0, 1870, 0, 0)` × itself: when `Compare` is neither `Different` nor `Unknown`, the reaches overlap or one is open | D-4, AC-2.2 |
| T-5 | Go | `core/shortlist/shortlist_test.go` `TestKeys`, `TestLookup` *(new)* | AC-1.1 – AC-1.9, each a table row; a `Negative` value → no key from it; `YearBucket: 0` → no dated key | AC-1.1, AC-1.2, AC-1.3, AC-1.4, AC-1.5, AC-1.6, AC-1.7, AC-1.8, AC-1.9 |
| T-6 | Go | `core/matchregistry/registry_test.go` `TestShortlistSection` *(new)* | `YearBucket` > 0; every `Words` lead and every `Dated` `Type` and `Date` has `Default().Subject.Weights[kind][lead]` | D-6, D-10 |
| T-7 | Go | `core/graphalign/candidates_test.go` `TestUnkeyedEvidenceStaysUnderTheWeakBand` *(new)* | For each kind in `cfg.Subject.Weights`: the sum of `Agree` over leads that are neither a `Words` lead nor the kind's `Dated.Date` lead is under `cfg.WeakScore` | AC-2.3 |
| T-8 | Go | `core/database/graphcache/graph_test.go` `TestMatchIndexDrops` *(new)* | Store a value; `SetAliasTables(nil)` → `MatchIndex()` false; store; commit naming a handle → false; store; `Drop` → false; store; vocabulary-only commit → true | AC-3.4, D-12 |
| T-9 | Go | `core/database/promotealign/promotealign_test.go` `TestProposeSeesAHandlePromotedAfterTheIndexWasBuilt` *(new)* | AC-3.1 | AC-3.1, D-13 |
| T-10 | Go | `core/database/promotealign/promotealign_test.go` `TestProposeQueryCountStaysFlatWhenWarm` *(new)*, `TestProposeQueryCountDoesNotGrowWithTheLayer` (unchanged) | The existing test's two Sources (2 and 6 people), each proposed twice with no commit between; the second proposals' counts are equal to each other | AC-3.2, D-13 |
| T-12 | Go | `core/database/promotealign/promotealign_test.go` `TestProposeColdIndexBuildsNoUnshortlistedLife` *(new)* | Two catalogs, one with 2 and one with 6 promoted persons named "Unrelated 1" … each the subject of a dated birth; a Source with one person "Zed Zane"; `Graph().Drop()` then one counted `Propose` on each → equal query counts (the index build follows no participation; S11-13's scan followed one per person) | AC-3.5, D-13 |
| T-13 | Go | `core/database/promotealign/bench_test.go` `BenchmarkProposeWarmHub` *(new)* | `BenchmarkProposeColdHub`'s fixture; one untimed `Propose`, then timed `Propose` calls with no `Drop` → reported, not gated: ns/op and `heap-B` on `main` and on the branch, in the PR description | AC-3.6, D-13 |
| T-11 | Go | Every existing test in `core/database/promotealign`, `api/ffi/handlers` and `core/graphalign` | Unchanged → pass with no expected value edited | AC-3.3 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `TestProposeQueryCountDoesNotGrowWithTheLayer`, `TestKnobInventoryMatchesRegistry`, the purity guard.

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
- [ ] deployment-plan.md: map row status; S11-16 scope entry (D-15); completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated — not gated
- [ ] Domain docs and skill updated (D-15): matching-model.md, comparison-layers.md, promote-graph-alignment.md, add-match-rule
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

1. Open a dogfood project with at least two promoted Sources. Promote a third Source that shares a person by name with one of them. The person proposes the shared handle as before.
2. Decide that row, Apply, and promote a fourth Source naming the same person. It proposes the handle Apply just filed (the index was rebuilt after the commit).

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| A key misses a pair the scan found, and a test moves | T-2 and T-4 hold the guarantee at layer 1; T-7 holds that a keyless pair cannot reach the weak band; AC-3.3 keeps every catalog test unchanged |
| The cold build reads every person, event and place node after each handle commit | No life is built (D-13); cached until the next commit; S11-21 sets the budget with T-13's numbers |
| A wide date's `any` key is the whole dated type | Only open-ended probes read it; closed dates read their own buckets |
| A wide closed range (`BET 1700 AND 1900`) adds about a hundred bucket keys | Rare; bounded by the range; S11-21 measures it |

**Stop and report if:**

- STEP-1 is reached on a plan whose status is not `Planned`.
- T-12's counts differ for a reason other than life loading (name the query that grows).
- Any expected value in `core/database/promotealign`, `api/ffi/handlers` or `core/graphalign` tests changes.
- T-2 or T-4 finds a related pair with no shared key or overlapping reach under D-2 / D-4 as written.
- T-7 fails under `matchregistry.Default()` (a kind could reach the weak band with no key).
- STEP-2 needs a test edited (`Side` and `Typed` change no behavior).
- `core/shortlist` needs an import under `core/database` or `core/matchregistry`.
- A `g.Kind` read is needed in `loadCanon` itself.

## 12. Open questions

None.

## 13. PR

**Title:** S11-15 — Shortlist index
**Body:** objective in one sentence; what shipped against US-1 – US-3; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-11/prs/S11-15-shortlist-index.md`.
