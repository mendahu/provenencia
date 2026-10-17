# Matching model

The matching model Spike 11 builds. Layers land in S11-02 – S11-16; until then [`matching.md`](matching.md) describes the running code. The reasoning behind each choice is in [`deployment-plan/spike-11/comparison-layers.md`](deployment-plan/spike-11/comparison-layers.md); this page is the "what", kept true as the layers land.

## The stack

Six pure layers. Each layer reads only the layers under it and its own section of the match registry. Nothing in layers 1 to 5 reads the catalog, the clock or a database.

```text
matchregistry ─────────────────────────────────────────────┐ (each layer reads its section)
1 value compare      value × value       → relation, similarity, detail
2 property compare   values × values     → outcome, speaking pair
3 Subject compare    Subject × Subject   → similarity, evidence
4 entity compare     Subject × profile   → similarity, evidence   ◀── subject profile builder
5 segment compare    segment × segment   → outcome, similarity, evidence
6 graph alignment    layer × canon       → rows: target, band, contested, reasons
```

| # | Layer | Package | Input | Output | Registry section |
| --- | --- | --- | --- | --- | --- |
| 1 | Value compare | `core/valuecompare` | Two values of one value type | A relation, a similarity from 0 to 1, and a detail saying which parts paired | `value` |
| 2 | Property compare | `core/propertycompare` | Every value two Subjects hold for one property, with the property's value type and cardinality | An outcome (agree, partial, conflict, unknown) and the speaking pair | `value` (passed through to layer 1) |
| 3 | Subject compare | `core/subjectcompare` | Two Subjects of one type, with every property each holds | A similarity and evidence in points | `subject` |
| 4 | Entity compare | `core/subjectprofile` (`Build`) with `core/subjectcompare` | A Subject and an entity's reconciled conclusion, built into a Subject-shaped profile | A similarity and evidence in points | `subject` |
| 5 | Segment compare | `core/segmentcompare` | A segment on the layer and a segment in the research graph | An outcome, a similarity, evidence, and a detail | `segment`, `subject` |
| 6 | Graph alignment | `core/graphalign` | The layer's Subjects and the research graph's handles | Rows: target, band, contested mark, reasons, alternatives | `alignment`, `segment`, `subject` |

The subject profile builder sits beside layer 4. It turns the reconciler's conclusion into a Subject-shaped profile and compares nothing.
Layer 6 decides what to compare (shortlisting), in what order (one best-first queue), and what to place (the accept bar); a row's band comes from its evidence. See the reasoning in [`comparison-layers.md`](deployment-plan/spike-11/comparison-layers.md#layer-6-procedure).

## Worked example

Gracie Gates appears on a Source with a name and a birth on 1 May 1901. The research graph has a Person handle with the same name and a birth on 1 May 1901. Every number below is a starting value from the [knob inventory](#knob-inventory).

1. **Value compare.** The two names are `Same`; the two dates are `Same`; the two event types are `Same`.
2. **Property compare.** Name: relation `Same` on a single-value property, so the outcome is agree.
3. **Subject compare, Gracie.** `subject.person.name.Agree` is 2.5. Both sides also derive a lifespan, "born 1 May 1901", from their birth; the windows are consistent, so `subject.person.lifespan.Agree` 0.1 counts. Her own evidence is 2.5 + 0.1 = 2.6.
4. **Subject compare, the birth.** Event type agrees: `subject.event.event_type.Agree` 0.5. The date agrees: `subject.event.date.Agree` 2. The birth's own evidence is 0.5 + 2 = 2.5.
5. **Segment compare.** Gracie is the subject of the birth, and `segment.Cardinality.birth` is `one`, so a matching neighbor earns `segment.OneSupport` 1.5 plus `segment.OneCredit` 0.5 of the birth's own evidence: 1.5 + 0.5 × 2.5 = 2.75.
6. **Graph alignment.** Total evidence is 2.6 + 2.75 = 5.35. That clears `alignment.AcceptBar` (2), so the walk places the pair, and it is at least `alignment.Bands.Strong` (4), so the band is strong. Nothing conflicted, so the row is not contested.

## Glossary

| Term | Means |
| --- | --- |
| **value** | One recorded fact of one value type: a name, a date, a piece of text, a term, an integer |
| **value type** | The kind of a value; each has its own value comparison |
| **speaking pair** | The one pair of values, one from each side, that decides a property comparison |
| **relation** | What kind of match two values are: `Same`, `Equivalent`, `Contains`, `Overlaps`, `Near`, `Different` or `Unknown` |
| **`Same`** | Identical after the normalization every comparer does (case, whitespace, punctuation) |
| **`Equivalent`** | The same value written another way: accent fold, alias or nickname, abbreviation |
| **`Contains`** | One side is more specific and fits inside the other (`1817` and `3 Mar 1817`, `Robins` and `James Robins`); the detail says which side is broader |
| **`Overlaps`** | Each side has something the other lacks, but they share some |
| **`Near`** | Separate values that are close, at least the type's `Near` minimum similarity; below it the pair is `Different` |
| **`Different`** | No match |
| **`Unknown`** | Nothing to compare |
| **similarity** | How alike two things are, from 0 to 1; 1 is `Same` and 0 is `Different` |
| **detail** | The structured reasoning behind a result: which parts paired, which were left over, the cause |
| **cause** | A code from a shared vocabulary (`exact`, `fold`, `alias`, `initial`, `spelling`, `precision`, `missing_part`, `too_long`, and so on) naming the rule that decided a relation or outcome |
| **value ref** | An opaque reference a caller attaches to a value (an Observation id); the layers hand it back and never read it |
| **key** | A normalized form of a value; two values have the same key exactly when they are `Same` (or `Equivalent`, at that level) |
| **property comparison** | Layer 2: whether two Subjects agree on one property, from every value each holds |
| **outcome** | Layer 2's verdict for one property: agree, partial, conflict or unknown |
| **Subject comparison** | Layer 3: how alike two Subjects are over the properties both have |
| **evidence** | Points of connection: agree weights minus conflict penalties, plus segment support; the main currency from layer 3 up |
| **agree weight** | Points a property earns when it agrees |
| **conflict penalty** | Points a property costs when it conflicts; also counts as 0 in similarity |
| **pair (compared together)** | Two properties compared as one: sex with gender, `date` with `start_date` and `end_date` |
| **profile** | A Subject-shaped set of values built from an entity's reconciled conclusion |
| **entity comparison** | Layer 4: a Subject against an entity's profile |
| **lifespan** | A derived window for a person, built from the dates of direct neighbors, compared as a registry-declared special case |
| **segment** | Two primary Subjects joined by a bridge: primary, bridge, primary |
| **bridge** | The Subject that joins the two primary Subjects of a segment (a birth, a relationship) |
| **cardinality** | How many a Subject can have through one bridge; only birth and death are declared `one` |
| **support** | Points a matching neighbor adds to a pair through a segment |
| **credit** | The share of a neighbor's own evidence passed on to the pair |
| **shortlisting** | The cheap index lookup that decides which handles get compared at all |
| **shortlist key** | A normalized value used for that lookup: a name word, an event type and year bucket, a toponym |
| **shortlisted handle** | A handle the lookup returned; nothing has been compared yet |
| **nominated handle** | A handle reached through a link from an accepted pair; no lookup needed |
| **candidate** | A shortlisted or nominated handle that has been scored and kept (top K) |
| **anchor** | A pair already placed (a decided row or an accepted pair) whose segments add evidence to its neighbors |
| **accept bar** | The least evidence for the walk to place a pair |
| **queue** | The one best-first list of candidate pairs the walk pops |
| **dirty pair** | A pair scored before a neighbor was placed, moved or released; it is re-scored when the queue empties |
| **settle** | Re-score dirty pairs until none changes or the round cap is reached |
| **band** | Strong, medium, weak or no match, for a whole row, from its evidence |
| **contested** | A visible mark on a row whose comparisons include a conflict, whatever its band |
| **match registry** | The one package holding every matching tunable, one section per layer group |
| **knob** | One registry field a developer can tune; changing it is one edit in the registry |

## Retired words

Both vocabularies are in the code for several PRs. These old words go.

| Old word | Now |
| --- | --- |
| score | similarity (0 to 1) or evidence (points) |
| points | evidence |
| weight | agree weight |
| contradiction | conflict penalty |
| scale and floor | relation outcome and the `Near` minimum |
| fan-out | cardinality |
| signature | bridge key |
| support for a link | support; the reconciler's support count keeps its own name |
| `Rank` | shortlisting and entity comparison |
| profile in `core/match` | the subject registry section |
| blocking | shortlisting |
| frequency, m, u, discount | removed |

## Knob inventory

Keys are dotted paths of the registry, `core/matchregistry`: `section.Group.Field`. A Capitalized part is a registry field. A lowercase part is a typed map key: a Subject kind, a property key, a derived property (`lifespan`), or a bridge or participation key; a pair compared together is keyed by its lead property (`sex_at_birth` covers `gender`). The `evidence` section is the claim evidence sheet's line policy ([`claim-evidence.md`](deployment-plan/spike-11/claim-evidence.md)), not evidence points; the Promote evidence sheet reads it, no comparison layer does. The Go registry is the only source of a current value. Start is the value a knob started with. The registry's drift guard compares keys only: every registry field has a row, and every row is a field. `TestKnobInventoryMatchesRegistry` enforces it; rows whose layer has not landed are listed in its `notYetBuilt`. Tuning a knob is one edit, in the registry.

| Key | Start | Source | Changes |
| --- | --- | --- | --- |
| `value.Words.FuzzyFloor` | 0.8 | `autoreconcile.SpellingFloor` | Least edit ratio for a spelling variant |
| `value.Words.ShortVariantMinLength` | 3 | `autoreconcile.SpellingShortMin` | Words this long may differ by one added or dropped letter |
| `value.Words.InitialCredit` | 0.5 | `DefaultWordRules.InitialCredit` | Initial against the word it begins |
| `value.Words.AliasCredit` | 0.98 | `match.AliasCredit` | Nickname or abbreviation hit |
| `value.Words.FoldCredit` | 0.95 | `match` `foldOnly` | Pair equal only after folding accents |
| `value.Words.MaxExactPairing` | 12 | `match.MaxExactPairing` | Words above this pair greedily |
| `value.Words.MaxSpellingLength` | 32 | estimate | Longest word, in runes, that may pair as a spelling variant |
| `value.Words.Particles` | `o`, `mc`, `mac`, `st`, `d`, `fitz` | `match.SpacedParticles` | Words that rejoin the next word |
| `value.Words.Folds` | today's map | `match.LetterFolds` | Letters replaced when folding |
| `value.Text.Partial` | 0.7 | `DefaultText.Partial` | Cap for text that shares only some words |
| `value.Text.MaxWords` | 64 | estimate | Most words either side may have for the alignment to run; beyond it `too_long` |
| `value.Names.NearMinimum` | 0.75 | `graphalign` name `PropertyScale.Floor` | Below this a name pair is `Different` |
| `value.Names.UntypedAffinity` | 0.8 | `DefaultNames.UntypedAffinity` | Typed word against an untyped one |
| `value.Names.NickAffinity` | 0.9 | `DefaultNames.NickAffinity` | Nickname against a given name |
| `value.Names.CrossRole` | 0.5 | `DefaultNames.CrossRole` | Any other role mismatch |
| `value.Names.Generation` | 0.5 | estimate | Weight of a generation word, Jr. or III, on either side |
| `value.Names.Western.Weights.Family` | 1.5 | `WesternNamePattern.Weights` | Surname word weight |
| `value.Names.Western.Weights.FirstGiven` | 1 | `WesternNamePattern.Weights` | First given name weight |
| `value.Names.Western.Weights.OtherGiven` | 0.5 | `WesternNamePattern.Weights` | Middle names and initials |
| `value.Names.Western.Weights.Nick` | 0.3 | `WesternNamePattern.Weights` | Nickname part weight |
| `value.Names.Western.Weights.Untyped` | 1 | `WesternNamePattern.Weights` | Untyped part weight |
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
| `subject.Pairs.person.sex_at_birth` | `sex_at_birth` with `gender` | decided ([`comparison-layers.md`](deployment-plan/spike-11/comparison-layers.md#layer-3-subject-comparison)) | Compared as one pair |
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
| `alignment.Shortlist.person` | name words, not initials | decided ([`comparison-layers.md`](deployment-plan/spike-11/comparison-layers.md#layer-6-procedure)) | Shortlist keys for a person |
| `alignment.Shortlist.event` | event type + year bucket | decided ([`comparison-layers.md`](deployment-plan/spike-11/comparison-layers.md#layer-6-procedure)) | Shortlist keys for an event |
| `alignment.Shortlist.place` | toponym | decided ([`comparison-layers.md`](deployment-plan/spike-11/comparison-layers.md#layer-6-procedure)) | Shortlist keys for a place |
| `evidence.Checked` | `Same`, `Equivalent` | decided ([`claim-evidence.md`](deployment-plan/spike-11/claim-evidence.md#lines)) | Relations whose lines start checked |
| `evidence.Pinnable` | `Contains`, `Overlaps`, `Near` | decided ([`claim-evidence.md`](deployment-plan/spike-11/claim-evidence.md#lines)) | Pinnable but unchecked |
| `evidence.Defining.event` | `event_type` | decided ([`claim-evidence.md`](deployment-plan/spike-11/claim-evidence.md#evidence-through-relationships)) | Neighbor properties bundled into a relationship line |

## Not knobs

These are identifiers or catalog data, not tunables, and have no row: the vocabulary origins and primary kinds (`core/vocabkey`), `AliasAbbreviation`, `AliasNickname`, `valuecompare.WesternPattern` (a key), `BuiltinNamePatterns` (the set holding `WesternNamePattern`), `DefaultNames.Patterns` / `Pattern`, `value.Names.Aliases` and `value.Names.Patterns` (catalog rows, passed in at run time), and `graphalign` `Config.Names` (a test hook).

## Removed knobs

| Today | Why removed |
| --- | --- |
| `DefaultNames.GivenOnlyFactor`, `SuffixConflict` | A surname or first given name left over on both sides, and a generation conflict, are `Different` (S11-03) |
| `DefaultNames.Frequency`, `FrequencyDiscount`, `FrequencyMinPeople`; `graphalign` `MPrior`, `UPrior`, frequency `ValueTypeScale` entries | Frequency weighting is descoped |
| `graphalign` `PropertyAgreement` | Frequency weighting is descoped; the toponym bonus is folded into `subject.place.toponym.Agree` |
| `graphalign` `FanOutLowMax`, `FanOutUnknown` | Replaced by declared cardinality |
| `graphalign` `ValueTypeScale` floors other than date, `PropertyScale` sex / gender `Floor` 1 | Replaced by relation outcomes and the `Near` minimum |
| `graphalign` `DefaultConfig.ConflictPenalty` | Value carried into `subject.event.date.Conflict` and `subject.place.toponym.Conflict`; its second role, the conflict-with-decision margin, is replaced by the contested marker and the settle margin |
| `graphalign` `mFor` fallback 0.85, `uOr` fallback 0.1 | Frequency weighting is descoped |
| `DefaultProfile` `start_date` / `end_date` features | Replaced by the `date` pair in `subject.Pairs.event.date` |
| `DefaultDates.ApproxToleranceFactor` | Replaced by `ABT` as a range |
| `DefaultProfile` weights and `MinScore` (`Rank`'s 10-point scale), `DefaultSuggestionLimit` | `Rank` is replaced by shortlisting and entity comparison; one weight set |
