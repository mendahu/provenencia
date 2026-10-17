# S11-01 — One-page matching model

**Status:** Planned
**Spike / slice:** Spike 11 · Slice 1 — Graspable model and pure foundations
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no new chrome: docs only)
**Board:** —
**Depends on:** —
**Blocks:** S11-02
**Branch:** `s11-01-matching-model`
**Version bump:** none — docs only ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** —
**Rules that bind:** —
**Context:** [`comparison-layers.md`](../comparison-layers.md) (whole), [`claim-evidence.md`](../claim-evidence.md), working rules "Registry numbers: carry over, then estimate", "One match registry", "Matching is a stack of pure layers"; [`ideas/matching-test-suite.md`](../../../ideas/matching-test-suite.md) "First: make the algorithm graspable"
**Planned:** 2026-10-17 against `main` @ `af50bd6`
**Changed in ship (2026-10-17):** lifespan keys under `subject`, `sex_at_birth` keys, key grammar, worked example counts the lifespan agree, narrower banners — user answers Q1–Q3 in the ship log.
**Review:** [`S11-01-matching-model.review.md`](S11-01-matching-model.review.md) — Fits with changes / Not ready → fixed, 1 round

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

One domain doc explains the matching algorithm Spike 11 builds: the six layers, the words we use for it, and every number it can be tuned by, each with its starting value. It is the "make the algorithm graspable" step of [`ideas/matching-test-suite.md`](../../../ideas/matching-test-suite.md), and the reference every later Spike 11 PR builds against and keeps true.

**Why now.** S11-02 creates the match registry and its inventory-drift guard, which reads the knob inventory this PR writes. Without one agreed list of knobs and starting values, each layer PR would invent its own.

**What stays.** No code changes. [`docs/matching.md`](../../../matching.md) and [`docs/promote-graph-alignment.md`](../../../promote-graph-alignment.md) still describe the code that runs today; they get a banner and section markers, and the layer PRs fold or delete them as the code moves (D-5).

## 2. User stories

**US-1.** As a contributor, when I open the matching model, I can follow one Promote row from value comparison to its band on one page, so that I know which layer owns a behavior before I change it.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | The model's diagram | It shows the six layers in order (value, property, Subject, entity, segment, graph alignment), the subject profile builder beside layer 4, the match registry feeding all of them, and what each layer passes up |
| AC-1.2 | The model's layer table | Each layer has one sentence each for its input, its output, and the registry section it reads, matching [`comparison-layers.md`](../comparison-layers.md) |
| AC-1.3 | The worked example | "Gracie Gates, name + birth 1 May 1901" is traced from value compare to band, with the numbers from the knob inventory |

**US-2.** As a contributor, when I meet a matching word (relation, evidence, segment, shortlist, contested…), the glossary says what it means, and which old word it replaces.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | The glossary | Every term in D-3's list has a one- or two-sentence definition |
| AC-2.2 | The retired-words table | Each old word in D-4 maps to its new term or to "removed" |

**US-3.** As a contributor tuning matching, every knob is listed once with its starting value and where the value came from.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | The knob inventory | It has one row per knob in §6's inventory, with key, starting value, source (today's symbol or "estimate" with a reason), and what it changes |
| AC-3.2 | A carried-over knob | Its source names today's symbol: name agree 2.5 → `core/graphalign/registry.go` `DefaultConfig` name `PropertyScale` |
| AC-3.3 | The removed-knobs and not-a-knob tables | Every exported var, const and `Config` field in today's `core/match/registry.go` and `core/graphalign/registry.go`, and every value in `DefaultConfig`, is an inventory Source, a removed row, or a not-a-knob row |

## 3. Scope

| In | Covered by |
| --- | --- |
| `docs/matching-model.md` (new): diagram, layer table, worked example, glossary, retired words, knob inventory, removed knobs, not-a-knob list | STEP-1 – STEP-5, AC-1.1 – AC-3.3 |
| Banner on `docs/matching.md` and `docs/promote-graph-alignment.md`; section markers on the superseded parts of the latter | STEP-6 |
| Links from the Spike 11 README and `comparison-layers.md` to the model | STEP-7 |

| Out | Home |
| --- | --- |
| The match registry package and its drift guard | S11-02 |
| Rewriting or deleting `docs/matching.md` | S11-05, S11-13 (D-5) |
| Rewriting the superseded sections of `docs/promote-graph-alignment.md` | S11-16 (D-5) |
| Tuning any number beyond carry-over and estimates | not planned in Spike 11 (working rule "Registry numbers: carry over, then estimate") |
| The rest of the test suite | [`ideas/matching-test-suite.md`](../../../ideas/matching-test-suite.md) |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Matching domain doc | `docs/matching.md` | Describes `core/match` today: comparers, profiles with `Rank`'s 10-point weights, name frequency, support weighting. Becomes wrong as Spike 11 lands |
| Graph alignment doc | `docs/promote-graph-alignment.md` §4.1, §5, §6, §12 | Describes today's registry, walk, fan-out scoring and implementation status. §1–§3 and §9 (the problem, principles, the page, writing) stay true |
| Spike design | `docs/deployment-plan/spike-11/comparison-layers.md` | The decided design, written as requirements with decision history. A spike doc: archived at spike close |
| Today's comparer tunables | `core/match/registry.go` `DefaultWordRules`, `AliasCredit`, `MaxExactPairing`, `DefaultNames`, `WesternNamePattern`, `DefaultDates`, `DefaultText`, `DefaultIntegers`, `DefaultSuggestionLimit`, `DefaultProfile`; `core/match/words.go` `foldOnly`; `core/autoreconcile/spelling.go` `SpellingFloor`, `SpellingShortMin` | Sources for the value section's carried-over values |
| Today's alignment tunables | `core/graphalign/registry.go` `DefaultConfig`; `core/database/promotealign/canon.go` `seedCandidateLimit` | Sources for the subject, segment and alignment sections |
| Tests | — | None for docs; links are checked by hand (§10) |

**Precedent.** `docs/design-system-layers.md` graduated from [`ideas/archive/design-system-hardening.md`](../../../ideas/archive/design-system-hardening.md): a spike's design became the lasting domain doc, and the idea was archived. Copy that move: the domain doc holds the model and rules; the spike doc keeps its decision history until spike close.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | The model is a new `docs/matching-model.md` (new), not a rewrite of `docs/matching.md` | `docs/matching.md` must keep describing the running code until the layer PRs replace it; a domain rule that outlives the spike belongs in a domain doc, not in the spike folder | Rewriting `docs/matching.md` now — because it would describe code that does not exist for several PRs |
| D-2 | The diagram is a `text` code block, like `docs/promote-graph-alignment.md` and the spike docs | Repo style; renders everywhere | Mermaid or an image — because no domain doc uses them |
| D-3 | Glossary terms: value, value type, speaking pair, relation (each of `Same`, `Equivalent`, `Contains`, `Overlaps`, `Near`, `Different`, `Unknown`), similarity, detail, cause, value ref, key, property comparison, outcome (agree, partial, conflict, unknown), Subject comparison, evidence, agree weight, conflict penalty, pair (compared together), profile, entity comparison, lifespan, segment, bridge, cardinality, support, credit, shortlisting, shortlist key, shortlisted handle, nominated handle, candidate, anchor, accept bar, queue, dirty pair, settle, band, contested, match registry, knob | These are the words the layer PRs and the evidence sheet use | A longer list — because a term no PR uses is noise |
| D-4 | Retired words, each with its replacement: score (→ similarity or evidence), points (→ evidence), weight (→ agree weight), contradiction (→ conflict penalty), scale and floor (→ relation outcome and `Near` minimum), fan-out (→ cardinality), signature (→ bridge key), support for a link (→ support; the reconciler's support count keeps its own name), `Rank` (→ shortlisting + entity comparison), profile in `core/match` (→ subject registry section), blocking (→ shortlisting), frequency / m / u / discount (→ removed) | The vocabulary problems found while scoping ([`comparison-layers.md`](../comparison-layers.md#why)) | Leaving old words undocumented — because both vocabularies will be in the code for several PRs |
| D-5 | Who updates the old docs: S11-05 points `docs/matching.md` at the model for comparers; S11-13 deletes `docs/matching.md` (its content is in the model and in code comments); S11-16 rewrites `docs/promote-graph-alignment.md` §4.1, §5, §6 and §12 against the model, and re-points the links into those sections from `docs/conclusion-layer-data-model.md` (L256) and `docs/conclusion-reconciliation.md` (L141). This PR only adds the banner and markers | Each doc changes in the PR that makes it false (working rule "One matching skill, grown with the layers") | Deleting the old docs now — because they still describe the running code |
| D-6 | Inventory keys are dotted paths of the planned registry: `section.Group.Field` (`value.Names.UntypedAffinity`, `subject.person.name.Agree`, `alignment.AcceptBar`). The table is a plain Markdown table with exactly the columns Key, Start, Source, Changes, and every registry field is a row, numeric or not (data such as particles and cardinality included). **The Go registry is the only source of a current value.** Start is the value a knob started with; S11-02's drift guard compares **keys only** (every registry field has a row, every row is a field), never values, so tuning a knob is one edit | One table the guard can read, and no number kept equal in two places (`core/match/registry.go`: "no number is written twice") | A guard that compares values — because every tuning edit would also edit the doc; prose lists — because a guard can't read them |
| D-7 | Starting values: carried over from today's symbol where the knob exists and has a fixed value; otherwise an estimate written as "estimate: <one-sentence reason>", including values today's code derives at run time (frequency weights). A symbol whose value is carried into a new knob while the symbol itself goes is listed in the inventory as the Source and in the removed table as "value carried into <key>". The values are the ones in §6's inventory | Working rule "Registry numbers: carry over, then estimate" | Picking new values by feel — because nothing measures them yet |
| D-8 | The comparison-layers doc stays as the spike's design record; the model links to it for the reasoning and does not copy the decision history | The model is the "what"; the spike doc is the "why", archived at spike close with `archive-docs` | Merging both — because the model would carry dated decision notes forever |

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `docs/matching-model.md` *(new)* | file | `docs/matching.md`, `docs/promote-graph-alignment.md` | They describe today's code, which runs until the layer PRs land (D-1, D-5) |

**Builds on.** The domain-doc pattern of `docs/design-system-layers.md`; the spike design in `comparison-layers.md`.
**Deletes.** None — nothing becomes dead in this PR; D-5 names the PRs that delete the old docs.
**Next adds.** S11-02 reads the knob inventory for its drift guard; each layer PR edits its own inventory rows and glossary terms when it lands.

### Schema

None — no catalog change.

### Go core

None — docs only.

### FFI

None — docs only.

### macOS

None — docs only.

### Strings

None — docs only.

### Document outline

`docs/matching-model.md` (new), in this order:

1. **Status line:** "The matching model Spike 11 builds. Layers land in S11-02 – S11-16; until then `docs/matching.md` describes the running code." Links to [`comparison-layers.md`](../comparison-layers.md) for the reasoning.
2. **The stack** (diagram, D-2):

   ```text
   matchregistry ─────────────────────────────────────────────┐ (each layer reads its section)
   1 value compare      value × value       → relation, similarity, detail
   2 property compare   values × values     → outcome, speaking pair
   3 Subject compare    Subject × Subject   → similarity, evidence
   4 entity compare     Subject × profile   → similarity, evidence   ◀── subject profile builder
   5 segment compare    segment × segment   → outcome, similarity, evidence
   6 graph alignment    layer × canon       → rows: target, band, contested, reasons
   ```
3. **Layer table:** one row per layer: package, input, output, registry section (AC-1.2).
4. **Worked example:** Gracie (AC-1.3). Name `Same` → property agree; both sides derive the lifespan "born 1 May 1901" from their birth, the windows are consistent, and `subject.person.lifespan.Agree` 0.1 counts → her own evidence 2.5 + 0.1 = 2.6. Her birth on both sides: event type agrees (0.5) and date `Same` (2), so the birth's own evidence is 2.5. Birth segment: one-cardinality support 1.5 + credit 0.5 × 2.5 = 2.75. Total evidence 2.6 + 2.75 = 5.35 ≥ strong 4 → strong; no conflict → not contested.
5. **Glossary** (D-3) and **retired words** (D-4).
6. **Knob inventory** (D-6, D-7): the table below.
7. **Not knobs** and **removed knobs** (AC-3.3): the two lists below.

### Knob inventory (starting values)

| Key | Start | Source | Changes |
| --- | --- | --- | --- |
| `value.Words.FuzzyFloor` | 0.8 | `autoreconcile.SpellingFloor` | Least edit ratio for a spelling variant |
| `value.Words.ShortVariantMinLength` | 3 | `autoreconcile.SpellingShortMin` | Words this long may differ by one added or dropped letter |
| `value.Words.InitialCredit` | 0.5 | `DefaultWordRules.InitialCredit` | Initial against the word it begins |
| `value.Words.AliasCredit` | 0.98 | `match.AliasCredit` | Nickname or abbreviation hit |
| `value.Words.FoldCredit` | 0.95 | `match` `foldOnly` | Pair equal only after folding accents |
| `value.Words.MaxExactPairing` | 12 | `match.MaxExactPairing` | Words above this pair greedily |
| `value.Words.Particles` | `o`, `mc`, `mac`, `st`, `d`, `fitz` | `match.SpacedParticles` | Words that rejoin the next word |
| `value.Words.Folds` | today's map | `match.LetterFolds` | Letters replaced when folding |
| `value.Text.Partial` | 0.7 | `DefaultText.Partial` | Cap for text that shares only some words |
| `value.Names.NearMinimum` | 0.75 | `graphalign` name `PropertyScale.Floor` | Below this a name pair is `Different` |
| `value.Names.UntypedAffinity` | 0.8 | `DefaultNames.UntypedAffinity` | Typed word against an untyped one |
| `value.Names.NickAffinity` | 0.9 | `DefaultNames.NickAffinity` | Nickname against a given name |
| `value.Names.CrossRole` | 0.5 | `DefaultNames.CrossRole` | Any other role mismatch |
| `value.Names.GivenOnlyFactor` | 0.5 | `DefaultNames.GivenOnlyFactor` | Surnames share nothing |
| `value.Names.SuffixConflict` | 0.3 | `DefaultNames.SuffixConflict` | Generations differ (Jr. / Sr.) |
| `value.Names.Western.Family` | 1.5 | `WesternNamePattern.Weights` | Surname word weight |
| `value.Names.Western.FirstGiven` | 1 | `WesternNamePattern.Weights` | First given name weight |
| `value.Names.Western.OtherGiven` | 0.5 | `WesternNamePattern.Weights` | Middle names and initials |
| `value.Names.Western.Nick` | 0.3 | `WesternNamePattern.Weights` | Nickname part weight |
| `value.Names.Western.Untyped` | 1 | `WesternNamePattern.Weights` | Untyped part weight |
| `value.Names.Western.PartRoles` | today's map | `WesternNamePattern.PartRoles` | Which part type plays which role |
| `value.Dates.Tolerance` | 2 | `DefaultDates.Tolerance` | Most years apart that can still be `Near` |
| `value.Dates.NearMinimum` | 0 | `graphalign` `ValueTypeScale["date"].Floor` | Below this a date pair is `Different` |
| `value.Dates.MissingDay` | 0.9 | `DefaultDates.MissingDay` | Same month, a day missing on one side |
| `value.Dates.MissingMonth` | 0.8 | `DefaultDates.MissingMonth` | A month missing on one side |
| `value.Dates.OtherDay` | 0.7 | `DefaultDates.OtherDay` | Same month, different known days |
| `value.Dates.OtherMonth` | 0.6 | `DefaultDates.OtherMonth` | Different known months |
| `value.Dates.YearsApartFrom` | 0.8 | `DefaultDates.YearsApartFrom` | Start of the fall-off across years |
| `value.Dates.SpanOneYear` | 0.8 | `DefaultDates.SpanOneYear` | A span of a year or less |
| `value.Dates.SpanPerExtraYear` | 0.05 | `DefaultDates.SpanPerExtraYear` | Less for each extra year of span |
| `value.Dates.SpanFloor` | 0.2 | `DefaultDates.SpanFloor` | Lowest span similarity; open bounds |
| `value.Dates.AboutUnits` | 1 | estimate: the design's `ABT` rule, one unit of the most precise part each side | Width of an `ABT` range |
| `value.Integers.Tolerance` | 0 | `DefaultIntegers.Tolerance` | Most an integer pair may differ and be `Near` |
| `subject.person.name.Agree` | 2.5 | `graphalign` name `PropertyScale.Weight` | Points for a matching name |
| `subject.person.name.Conflict` | 3 | `graphalign` name `PropertyScale.Contradiction` | Points lost for a different name |
| `subject.person.sex_at_birth.Agree` | 0.6 | `graphalign` sex / gender `PropertyScale.Weight` | Points for matching sex or gender (the `sex_at_birth` / `gender` pair) |
| `subject.person.sex_at_birth.Conflict` | 1 | `graphalign` sex / gender `PropertyScale.Contradiction` | Points lost when they differ (the `sex_at_birth` / `gender` pair) |
| `subject.person.lifespan.Agree` | 0.1 | estimate: near zero, the dates are already counted through their segments | Points for overlapping lifespans |
| `subject.person.lifespan.Conflict` | 20 | estimate: impossible together, must never recover | Points lost for impossible lifespans |
| `subject.person.lifespan.MaxYears` | 110 | estimate: longest plausible human life | Born to died, at most |
| `subject.person.lifespan.ParentMinYears` | 12 | estimate: the design's "about 12–60 years" | Youngest age at a child's birth |
| `subject.person.lifespan.ParentMaxYears` | 60 | estimate: the design's "about 12–60 years" | Oldest age at a child's birth |
| `subject.person.lifespan.MotherMaxYears` | 50 | estimate: "narrower for mothers"; used only when the parent's sex is recorded female | Oldest mother at a child's birth |
| `subject.person.lifespan.MarriageMinYears` | 14 | estimate: the design's "roughly 14 or older" | Youngest age at a marriage |
| `subject.event.event_type.Agree` | 0.5 | estimate: generic, says little alone (decided low) | Points for the same event type |
| `subject.event.event_type.Conflict` | 20 | estimate: a birth is never a death | Points lost for different event types |
| `subject.event.date.Agree` | 2 | estimate: today's cold-start frequency weight ln(0.85 / 0.1) ≈ 2.1, rounded | Points for matching dates (one pair with start / end) |
| `subject.event.date.Conflict` | 3 | `graphalign` `DefaultConfig.ConflictPenalty` | Points lost for different dates |
| `subject.place.toponym.Agree` | 3 | estimate: today's cold-start ln(0.9 / 0.1) + toponym `PropertyAgreement` 1 ≈ 3.2, rounded, so a toponym alone reads medium | A matching toponym alone reads medium |
| `subject.place.toponym.Conflict` | 3 | `graphalign` `DefaultConfig.ConflictPenalty` | Points lost for different toponyms |
| `subject.Pairs.event.date` | `date` with `start_date` / `end_date` | `match.DatePeriods` | Compared as one period |
| `subject.Pairs.person.sex_at_birth` | `sex_at_birth` with `gender` | decided ([`comparison-layers.md`](../comparison-layers.md#layer-3-subject-comparison)) | Compared as one pair |
| `segment.OneSupport` | 1.5 | `graphalign` `EdgeSupportLow` | Points for a matching one-cardinality link (birth, death) |
| `segment.OneCredit` | 0.5 | `graphalign` `NeighborCreditLow` | Share of the neighbor's own evidence passed on, one-cardinality |
| `segment.OtherSupport` | 0.25 | `graphalign` `EdgeSupportHigh` | Points for any other matching link |
| `segment.OtherCredit` | 0.25 | `graphalign` `NeighborCreditHigh` | Share of the neighbor's evidence passed on, other links |
| `segment.PartOf.From.MinOwnEvidence` | 2 | `graphalign` `SignatureScale` `part_of` from-end `MinNode` | The part's own evidence needed before `part_of` support counts |
| `segment.PartOf.To.MinOwnEvidence` | 2 | `graphalign` `SignatureScale` `part_of` to-end `MinNode` | The whole's own evidence needed before `part_of` support counts |
| `segment.PartOf.From.ConflictFactor` | 1 | `graphalign` `SignatureScale` `part_of` from-end `ConflictFactor` | A different parent cancels the part's own evidence |
| `segment.Cardinality.birth` | `one` | decided (working rule on segment compare) | Participation · subject · birth is one per person |
| `segment.Cardinality.death` | `one` | decided (working rule on segment compare) | Participation · subject · death is one per person |
| `alignment.AcceptBar` | 2 | `graphalign` `AcceptScore` | Least evidence for the walk to place a pair |
| `alignment.Bands.Strong` | 4 | `graphalign` `StrongScore` | Strong band from this evidence |
| `alignment.Bands.Medium` | 3 | `graphalign` `MediumScore` | Medium band |
| `alignment.Bands.Weak` | 2 | `graphalign` `WeakScore` | Weak band; below is no match |
| `alignment.TopK` | 5 | `promotealign` `seedCandidateLimit` | Candidates kept per Subject after shortlisting |
| `alignment.TieCap` | 20 | estimate: room for a common name's ties without unbounded work | Most candidates kept when ties pass K |
| `alignment.Alternatives` | 3 | `graphalign` `AlternativeLimit` | Alternatives shown per row |
| `alignment.SettleMargin` | 1 | estimate: one point, a third of a band | Lead a rival needs to displace a placed pair |
| `alignment.SettleRounds` | 10 | estimate: a safety cap; dirty pairs normally settle in two or three | Most settle rounds |
| `alignment.Shortlist.YearBucket` | 2 | estimate: the date tolerance, so a `Near` date shares a bucket | Width of an event's year bucket |
| `alignment.Shortlist.person` | name words, not initials | decided ([`comparison-layers.md`](../comparison-layers.md#layer-6-procedure)) | Shortlist keys for a person |
| `alignment.Shortlist.event` | event type + year bucket | decided ([`comparison-layers.md`](../comparison-layers.md#layer-6-procedure)) | Shortlist keys for an event |
| `alignment.Shortlist.place` | toponym | decided ([`comparison-layers.md`](../comparison-layers.md#layer-6-procedure)) | Shortlist keys for a place |
| `evidence.Checked` | `Same`, `Equivalent` | decided ([`claim-evidence.md`](../claim-evidence.md#lines)) | Relations whose lines start checked |
| `evidence.Pinnable` | `Contains`, `Overlaps`, `Near` | decided ([`claim-evidence.md`](../claim-evidence.md#lines)) | Pinnable but unchecked |
| `evidence.Defining.event` | `event_type` | decided ([`claim-evidence.md`](../claim-evidence.md#evidence-through-relationships)) | Neighbor properties bundled into a relationship line |

### Not knobs

These are identifiers or catalog data, not tunables, and have no row: `OriginProvenencia`, `AliasAbbreviation`, `AliasNickname`, `WesternPattern` (a key), `BuiltinNamePatterns` (the set holding `WesternNamePattern`), `DefaultNames.Patterns` / `Pattern`, `DefaultWordRules.Aliases` (catalog tables passed in at run time), `ProfiledKinds` (S11-23), and `graphalign` `Config.Names` (a test hook).

### Removed knobs

| Today | Why removed |
| --- | --- |
| `DefaultNames.Frequency`, `FrequencyDiscount`, `FrequencyMinPeople`; `graphalign` `MPrior`, `UPrior`, frequency `ValueTypeScale` entries | Frequency weighting is descoped |
| `graphalign` `PropertyAgreement` | Frequency weighting is descoped; the toponym bonus is folded into `subject.place.toponym.Agree` |
| `graphalign` `FanOutLowMax`, `FanOutUnknown` | Replaced by declared cardinality |
| `graphalign` `ValueTypeScale` floors other than date, `PropertyScale` sex / gender `Floor` 1 | Replaced by relation outcomes and the `Near` minimum |
| `graphalign` `DefaultConfig.ConflictPenalty` | Value carried into `subject.event.date.Conflict` and `subject.place.toponym.Conflict`; its second role, the conflict-with-decision margin, is replaced by the contested marker and the settle margin |
| `graphalign` `mFor` fallback 0.85, `uOr` fallback 0.1 | Frequency weighting is descoped |
| `DefaultProfile` `start_date` / `end_date` features | Replaced by the `date` pair in `subject.Pairs.event.date` |
| `DefaultDates.ApproxToleranceFactor` | Replaced by `ABT` as a range |
| `DefaultProfile` weights and `MinScore` (`Rank`'s 10-point scale), `DefaultSuggestionLimit` | `Rank` is replaced by shortlisting and entity comparison; one weight set |

## 7. Steps

**STEP-1 — Skeleton and the stack**
Files: `docs/matching-model.md` (new).
Do: write the status line, the diagram (§6 outline item 2) and the layer table (AC-1.1, AC-1.2) from [`comparison-layers.md`](../comparison-layers.md) §The layers, §Layer 1 – §Layer 3, §Entities, §Segments, §Layer 6 procedure, §Bands.
Check: every layer in the diagram has a table row; package names match the map's working rule "Packages".

**STEP-2 — Worked example**
Files: `docs/matching-model.md` (new).
Do: write the Gracie trace (§6 outline item 4) using only inventory numbers (AC-1.3).
Check: each number in the example appears in the inventory table, and the sums are 2.5, 2.75 and 5.25.

**STEP-3 — Glossary and retired words**
Files: `docs/matching-model.md` (new).
Do: define every D-3 term; write the D-4 table (AC-2.1, AC-2.2). Definitions follow `comparison-layers.md` §Relations, §Detail, §Vocabulary, §Bands.
Check: every D-3 term has a row; `rg -n "e\.g\.|etc\." docs/matching-model.md` prints nothing.

**STEP-4 — Knob inventory**
Files: `docs/matching-model.md` (new).
Do: copy §6's knob inventory table (AC-3.1, AC-3.2, D-6, D-7). Verify each carried-over value against its source symbol in `core/match/registry.go`, `core/match/words.go`, `core/autoreconcile/spelling.go`, `core/graphalign/registry.go` and `core/database/promotealign/canon.go`.
Check: for each row whose Source names a symbol (not "estimate" or "decided"), the value equals that symbol's value at the planning base.

**STEP-5 — Removed knobs**
Files: `docs/matching-model.md` (new).
Do: copy §6's not-a-knob list and removed-knobs table (AC-3.3).
Check: every exported var, const and `Config` field in `core/match/registry.go` and `core/graphalign/registry.go`, and every value set in `DefaultConfig`, is an inventory Source, a removed row, or a not-a-knob entry.

**STEP-6 — Banners and markers**
Files: `docs/matching.md`, `docs/promote-graph-alignment.md`.
Do: add one blockquote line directly under each file's `# ` title: `> **Spike 11 replaces this page.** The target model is [matching-model.md](matching-model.md); until then this page describes the running code.` in `docs/matching.md`, and `> **Spike 11 replaces §4.1, §5, §6 and §12.** The target model is [matching-model.md](matching-model.md); until then those sections describe the running code.` in `docs/promote-graph-alignment.md`, each with matching-model.md as a relative Markdown link. In `docs/promote-graph-alignment.md`, add one blockquote line directly under each of the `## 4.1`, `# 5`, `# 6` and `# 12` headings: `> Superseded in Spike 11; see matching-model.md.`, with the same link (D-5). Heading text is unchanged.
Check: both files link to `matching-model.md`.

**STEP-7 — Wire and docs**
Files: `docs/deployment-plan/spike-11/README.md` (documents table), `docs/deployment-plan/spike-11/comparison-layers.md` (status line), `docs/deployment-plan/spike-11/deployment-plan.md` (map row), `docs/deployment-plan/spike-11/completed.md`, this plan.
Do: link the model from the README's documents table and from the comparison-layers status line. In the map row, set Status to `Landed [#N](https://github.com/mendahu/provenencia/pull/N)`. Add to `completed.md`: "## S11-01 — One-page matching model / **Landed** <date> in #N. `docs/matching-model.md` (new) holds the six-layer model, the glossary and the knob inventory with starting values; `docs/matching.md` and `docs/promote-graph-alignment.md` point at it." Set this plan's status to Landed.
Check: `python3 scripts/check-plan-docs.py` reports no new errors; every link added resolves (`ls` each target).

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

None — docs only. The ACs are checked by §10, and S11-02's drift guard later tests the inventory against the registry (D-6).

Must stay green: everything `scripts/check-pr.sh` runs (unchanged by a docs PR).

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together) — or none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated (if gated)
- [ ] Domain doc updated if a rule in docs/*.md changed: docs/matching-model.md (new) added; banners on docs/matching.md and docs/promote-graph-alignment.md
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

1. Open `docs/matching-model.md` (new). The diagram lists the six layers, the profile builder and the registry (AC-1.1); the layer table has a row per layer with input, output and section (AC-1.2).
2. Follow the Gracie example; each number is in the inventory and the band is strong (AC-1.3).
3. Pick five glossary terms from D-3 and five old words from D-4; each is defined or mapped (AC-2.1, AC-2.2).
4. Pick three carried-over rows; open the source symbol and confirm the value (AC-3.1, AC-3.2).
5. Read `core/graphalign/registry.go` `Config`: every field is an inventory source or a removed row (AC-3.3).

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| The model drifts from `comparison-layers.md` | D-8: the model links to it and copies no decision history; the layer PRs update both |
| A carried-over value is misread | STEP-4's check compares each row with its source symbol |

**Stop and report if:**

- A value in §6's inventory differs from its source symbol at the planning base.
- `comparison-layers.md` and this plan disagree on a layer's input or output.
- An exported var, const or `Config` field in `core/match/registry.go` or `core/graphalign/registry.go` fits none of the three lists.

## 12. Open questions

None.

## 13. PR

**Title:** S11-01 — One-page matching model
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-11/prs/S11-01-matching-model.md`.
