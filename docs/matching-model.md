# Matching model

The matching model Spike 11 built; all six layers run, and [`matching.md`](matching.md) says where. The reasoning behind each choice is in [`deployment-plan/spike-11/comparison-layers.md`](deployment-plan/spike-11/comparison-layers.md); this page is the "what".

**Revised 2026-10-29:** these layers propose Promote's targets. They are not the measure of whether a member belongs to its handle: that comparison is to be designed new ([`conclusion-layer-data-model.md`](conclusion-layer-data-model.md) §5.1), not built from layer 6's band. The band and the registry's `evidence` section are deprecated.

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
| 3 | Subject compare | `core/subjectcompare` | Two Subjects of one type, with every property each holds | A similarity and evidence in points | `subject`, and `value` passed through to layer 2 |
| 4 | Entity compare | `core/subjectprofile` (`Build`) with `core/subjectcompare` | A Subject and an entity's reconciled conclusion, built into a Subject-shaped profile | A similarity and evidence in points | `subject` |
| 5 | Segment compare | `core/segmentcompare` | A segment on the layer and a segment in the research graph | An outcome, a similarity, evidence, and a detail | `segment`, `subject` |
| 6 | Graph alignment | `core/graphalign` | The layer's Subjects and the research graph's handles | Rows: target, band, contested mark, reasons, alternatives | `alignment`, `segment`, `subject` |

The subject profile builder sits beside layer 4. It turns the reconciler's conclusion into a Subject-shaped profile and compares nothing: displayed values are positive, denied values negative, and outvoted, weak and provisional values are left out. It also builds a plain Subject's side from its own Observations.
Before layer 6, each side's links are read through the kinship patterns (`core/kinship`): a shared marriage implies spouses, a birth's father or mother and subject imply a parent, and a parent of a parent implies a grandparent. The Source side is derived when its layer loads; the canon side is read through `canonlinks.Reader.For`, which answers from the graph cache's results. Both go through one `kinship.LinkSet` (implied links enter through `AddImplied`, or `PlaceImplied` in `loadCanon`, each recorded by `Implied.Recording`), so a fact recorded two ways on one side is one link. An implied link weighs as the relationship it stands for. A fact both sides record through the same pattern counts once when the walk has paired the event or middle parent it passes through.
Layer 6 decides what to compare (shortlisting), in what order (one best-first queue), and what to place (the accept bar); a row's band comes from its evidence. See the reasoning in [`comparison-layers.md`](deployment-plan/spike-11/comparison-layers.md#layer-6-procedure).

## Worked example

Gracie Gates appears on a Source with a name and a birth on 1 May 1901. The research graph has a Person handle with the same name and a birth on 1 May 1901. Every number below is a starting value from the [knob inventory](#knob-inventory).

1. **Value compare.** The two names are `Same`; the two dates are `Same`; the two event types are `Same`.
2. **Property compare.** Name: relation `Same` on a single-value property, so the outcome is agree.
3. **Subject compare, Gracie.** `subject.Weights.person.name.Agree` is 2.5. Both sides also derive a lifespan from their birth: window 1 May 1891 to 1 May 2021, span 1 May 1901; each span lies in the other's window, so `subject.Weights.person.lifespan.Agree` 0.1 counts. Her own evidence is 2.5 + 0.1 = 2.6.
4. **Subject compare, the birth.** Event type agrees: `subject.Weights.event.event_type.Agree` 0.5. The date agrees: `subject.Weights.event.date.Agree` 2. The birth's own evidence is 0.5 + 2 = 2.5.
5. **Segment compare.** Gracie is the subject of the birth, and `segment.One` holds `participation·subject·birth`, so a matching neighbor earns `segment.OneSupport` 1.5 plus `segment.OneCredit` 0.5 of the birth's own evidence: 1.5 + 0.5 × 2.5 = 2.75.
6. **Graph alignment.** Total evidence is 2.6 + 2.75 = 5.35. That clears `alignment.AcceptBar` (2), so the walk places the pair, and it is at least `alignment.Bands.Strong` (4), so the band is strong. Its linked birth was not placed yet when the pair was first scored; entity compare found it alike, so the link counted from the start. Nothing conflicted, so the row is not contested.

## Glossary

| Term | Means |
| --- | --- |
| **value** | One recorded fact of one value type: a name, a date, a piece of text, a term, an integer |
| **value type** | The kind of a value; each has its own value comparison |
| **speaking pair** | The one pair of values, one from each side, that decides a property comparison |
| **relation** | What kind of match two values are: `Same`, `Equivalent`, `Contains`, `Overlaps`, `Near`, `Different` or `Unknown` |
| **`Same`** | Identical after the normalization every comparer does (case, whitespace, punctuation, and where words break, for text) |
| **`Equivalent`** | The same value written another way: accent fold, alias or nickname, abbreviation |
| **`Contains`** | One side is more specific and fits inside the other (`1817` and `3 Mar 1817`, `Robins` and `James Robins`); the detail says which side is broader |
| **`Overlaps`** | Each side has something the other lacks, but they share some |
| **`Near`** | Separate values that are close, at least the type's `Near` minimum similarity; below it the pair is `Different` |
| **`Different`** | No match |
| **`Unknown`** | Nothing to compare |
| **similarity** | How alike two things are, from 0 to 1; 1 is `Same` and 0 is `Different` |
| **detail** | The structured reasoning behind a result: which parts paired, which were left over, the cause |
| **cause** | A code naming the rule behind a result. Layer 1's relation causes (`exact`, `fold`, `alias`, `initial`, `spelling`, `precision`, `missing_part`, `too_long`, and so on) say why two values relate as they do; layer 2's outcome causes (`relation`, `negative`, `no_values`, `cardinality`) say why a property comparison came out as it did |
| **value ref** | An opaque reference a caller attaches to a value (an Observation id); the layers hand it back and never read it |
| **key** | A normalized form of a value; two values have the same key exactly when they are `Same` (or `Equivalent`, at that level) |
| **property comparison** | Layer 2: whether two Subjects agree on one property, from every value each holds |
| **outcome** | Layer 2's verdict for one property: agree, partial, conflict or unknown |
| **negative value** | A value a Subject says the property does not have; it contradicts a value on the other side that is `Same`, `Equivalent` or `Contains` |
| **neutral value** | A term that is no evidence (sex `unknown`); layer 2 drops it before comparing |
| **Subject comparison** | Layer 3: how alike two Subjects are over the properties both have |
| **evidence** | Points of connection: agree weights minus conflict penalties, plus segment support; the main currency from layer 3 up |
| **agree weight** | Points a property earns when it agrees |
| **conflict penalty** | Points a property costs when it conflicts; also counts as 0 in similarity |
| **pair (compared together)** | Two or more properties compared as one, keyed by the lead: a pool adds properties' values to the lead's on each side (sex at birth with gender); a period reads `start_date` and `end_date` as one range for `date` on a side with no date |
| **profile** | A Subject-shaped set of values built from an entity's reconciled conclusion: displayed values positive, denied values negative, nothing else |
| **entity comparison** | Layer 4: a Subject against an entity's profile |
| **lifespan** | A person's derived window and span, built from the exact dates (BEF, AFT and half-open ones are skipped) of every event they take part in, widened by a slack of years each way; each side's span is checked against the other's window |
| **segment** | Two primary Subjects joined by a bridge: primary, bridge, primary |
| **bridge** | The Subject that joins the two primary Subjects of a segment (a birth, a relationship) |
| **cardinality** | How many a Subject can have through one bridge; only birth and death are declared `one` |
| **support** | Points a matching neighbor adds to a pair through a segment |
| **credit** | The share of a neighbor's own evidence passed on to the pair |
| **shortlisting** | The cheap index lookup that decides which handles get compared at all |
| **shortlist key** | A normalized value used for that lookup: a name word, an event type and year bucket, a toponym word; made by layer 1 as index and lookup lists (`valuecompare.WordKeys`, `YearReach`) and indexed by `core/shortlist` |
| **shortlisted handle** | A handle the lookup returned; nothing has been compared yet |
| **nominated handle** | A handle reached through a link from an accepted pair; no lookup needed |
| **candidate** | A shortlisted or nominated handle that has been scored and kept (top K of a shortlist, or of one placed pair's nominations) |
| **anchor** | A pair already placed (a decided row or an accepted pair) whose segments add evidence to its neighbors |
| **accept bar** | The least evidence for the walk to place a pair |
| **queue** | The one best-first list of candidate pairs the walk pops |
| **dirty pair** | A pair scored before a neighbor was placed, moved or released; it is re-scored when the queue empties |
| **settle** | Re-score dirty pairs until none changes or the round cap is reached |
| **band** | Strong, medium, weak or no match, for a whole row, from its evidence. Deprecated: a signal about the proposal only |
| **contested** | A visible mark on a row whose comparisons include a conflict, whatever its band |
| **kinship pattern** | A row of `core/kinship`'s table that says a fact recorded one way implies a relationship link: a marriage's two subjects imply spouses, a birth's father or mother and subject imply a parent, and a parent of a parent implies a grandparent |
| **implied link** | A relationship link a kinship pattern adds to one side, keyed as a recorded relationship; never stored or filed |
| **shape** | One way a side records a link: a kinship pattern and the event or middle parent (its pivot) it passes through |
| **pivot** | The one node a kinship pattern reads: an event (a marriage or a birth), or the person between in a chain; the graph cache keeps one result per pivot, derived on first read (`canonlinks.Reader.For`) |
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

Keys are dotted paths of the registry, `core/matchregistry`: `section.Group.Field`. A Capitalized part is a registry field. A lowercase part is a typed map key: a Subject kind, a property key, a derived property (`lifespan`), or a bridge or participation key; a pair compared together is keyed by its lead property (`sex_at_birth` covers `gender`). The subject section's weights are `subject.Weights.<kind>.<property>`, a derived property's (`lifespan`) included; a pool is keyed by its lead in `subject.Pools`, a period by its lead in `subject.Periods`, a lifespan by its lead (the window) in `subject.Lifespans`. The `evidence` section is the claim evidence sheet's line policy ([`claim-evidence.md`](deployment-plan/spike-11/claim-evidence.md)), not evidence points; `core/claimevidence` reads it, no comparison layer does. It is deprecated and goes with the sheet and the pins ([`conclusion-layer-data-model.md`](conclusion-layer-data-model.md) §5). The Go registry is the only source of a current value. Start is the value a knob started with. The registry's drift guard compares keys only: every registry field has a row, and every row is a field. `TestKnobInventoryMatchesRegistry` enforces it; rows whose layer has not landed are listed in its `notYetBuilt`. Tuning a knob is one edit, in the registry.

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
| `value.Names.MaxWords` | 32 | estimate | Most words, generation words included, either name may have for the pairing to run; beyond it `too_long` |
| `value.Names.Generation` | 0.5 | estimate | Weight of a generation word, Jr. or III, on either side |
| `value.Names.PrefixMinLength` | 2 | decided: user, 2026-10-24 | Fewest letters of a given-name prefix the shortlist keys and the comparer counts; 0 turns the prefix rule off |
| `value.Names.PrefixCredit` | 0.7 | estimate | A given, nickname or untyped word that is the start of the other's word |
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
| `subject.Weights.person.name.Agree` | 2.5 | `graphalign` name `PropertyScale.Weight` | Points for a matching name |
| `subject.Weights.person.name.Conflict` | 3 | `graphalign` name `PropertyScale.Contradiction` | Points lost for a different name |
| `subject.Weights.person.sex_at_birth.Agree` | 0.6 | `graphalign` sex / gender `PropertyScale.Weight` | Points for matching sex or gender (the `sex_at_birth` / `gender` pair) |
| `subject.Weights.person.sex_at_birth.Conflict` | 1 | `graphalign` sex / gender `PropertyScale.Contradiction` | Points lost when they differ (the `sex_at_birth` / `gender` pair) |
| `subject.Weights.person.lifespan.Agree` | 0.1 | estimate: near zero, the dates are already counted through their segments | Points for consistent lifespans |
| `subject.Weights.person.lifespan.Conflict` | 20 | estimate: impossible together, must never recover | Points lost for impossible lifespans |
| `subject.Lifespans.person.lifespan.Span` | `lifespan_span` | derived (decided 2026-10-18) | The value holding a life's first to last dated event; the lead `lifespan` holds its window |
| `subject.Lifespans.person.lifespan.MaxYears` | 110 | estimate: longest plausible human life | Years an unfixed start or end is padded from the first or last exact dated event; open-ended dates (BEF, AFT) are skipped |
| `subject.Lifespans.person.lifespan.SlackYears` | 10 | estimate: forgiving margin for misdated records | Years the window is widened each way, so only egregious lifespans conflict |
| `subject.Lifespans.person.lifespan.Starts` | `birth` · `subject` | decided 2026-10-18 | Events that fix the start of a life |
| `subject.Lifespans.person.lifespan.Ends` | `death` · `subject`, `burial` · `subject` | decided 2026-10-18 | Events that fix the end of a life |
| `subject.Weights.event.event_type.Agree` | 0.5 | estimate: generic, says little alone (decided low) | Points for the same event type |
| `subject.Weights.event.event_type.Conflict` | 20 | estimate: a birth is never a death | Points lost for different event types |
| `subject.Weights.event.date.Agree` | 2 | estimate: the cold-start frequency weight ln(0.85 / 0.1) ≈ 2.1 that S11-24 removed, rounded | Points for matching dates (one pair with start / end) |
| `subject.Weights.event.date.Conflict` | 3 | `graphalign` `DefaultConfig.ConflictPenalty` | Points lost for different dates |
| `subject.Weights.place.toponym.Agree` | 3 | estimate: the cold-start ln(0.9 / 0.1) + toponym `PropertyAgreement` 1 ≈ 3.2 that S11-24 removed, rounded, so a toponym alone reads medium | A matching toponym alone reads medium |
| `subject.Weights.place.toponym.Conflict` | 3 | `graphalign` `DefaultConfig.ConflictPenalty` | Points lost for different toponyms |
| `subject.Periods.event.date.Start` | `start_date` | `match.DatePeriods` | Start of the period that stands for `date` on a side with no date |
| `subject.Periods.event.date.End` | `end_date` | `match.DatePeriods` | End of that period |
| `subject.Pools.person.sex_at_birth.With` | `gender` | decided ([`comparison-layers.md`](deployment-plan/spike-11/comparison-layers.md#layer-3-subject-comparison)) | Properties whose values join `sex_at_birth`'s on each side |
| `subject.Pools.person.sex_at_birth.Terms.gender` | `man` → `male`, `woman` → `female` | decided (user, 2026-10-21) | Reads a gender term as the sex at birth term it names before comparing; a gender with no entry (`non_binary`) is left out, never a conflict |
| `segment.OneSupport` | 1.5 | `graphalign` `EdgeSupportLow` | Points for a matching one-cardinality link (birth, death) |
| `segment.OneCredit` | 0.5 | `graphalign` `NeighborCreditLow` | Share of the neighbor's own evidence passed on, one-cardinality |
| `segment.OtherSupport` | 0.75 | decided: user, 2026-10-28 (S14-08): dogfood Jake + grandparent under Weak at 0.25 | Points for any other matching link |
| `segment.OtherCredit` | 0.25 | `graphalign` `NeighborCreditHigh` | Share of the neighbor's evidence passed on, other links |
| `segment.One` | `participation·subject·birth`, `participation·subject·death` | decided 2026-10-17, both ends 2026-10-19 | Links where a differing neighbor conflicts, from either end |
| `segment.OneConflict` | 3 | estimate: a differing birth or death is as strong as a differing name or date (`subject.Weights.person.name.Conflict`) | Points lost when a one-cardinality link's neighbor is placed elsewhere |
| `segment.UnplacedAlike` | 2 | estimate: the accept bar, a neighbor that could be placed on its own properties | Least neighbor evidence for an unplaced neighbor to count as lined up |
| `segment.Gates.place_relationship·part_of·from.MinOwnEvidence` | 2 | `graphalign` `SignatureScale` `part_of` from-end `MinNode` | The part's own evidence needed before `part_of` support counts |
| `segment.Gates.place_relationship·part_of·from.ConflictFactor` | 1 | `graphalign` `SignatureScale` `part_of` from-end `ConflictFactor` | A different parent cancels the part's own evidence |
| `segment.Gates.place_relationship·part_of·to.MinOwnEvidence` | 2 | `graphalign` `SignatureScale` `part_of` to-end `MinNode` | The whole's own evidence needed before `part_of` support counts |
| `segment.Gates.place_relationship·part_of·to.ConflictFactor` | 0 | decided: the whole's end had no conflict factor | A different whole is silent |
| `alignment.AcceptBar` | 2 | `graphalign` `AcceptScore` | Least evidence for the walk to place a pair |
| `alignment.Bands.Strong` | 4 | `graphalign` `StrongScore` | Strong band from this evidence |
| `alignment.Bands.Medium` | 3 | `graphalign` `MediumScore` | Medium band |
| `alignment.Bands.Weak` | 2 | `graphalign` `WeakScore` | Weak band; below is no match |
| `alignment.MinSimilarity` | 0.4 | decided: user, 2026-10-24 (S11-43): one given name reaches a record with three more names ("Amy" / "Amy Lee Anne Long" 0.44); a name with a sex conflict is 0.71, a birth against a death 0.09 | Least similarity for a shortlisted handle to be a candidate |
| `alignment.TopK` | 5 | `promotealign` `seedCandidateLimit` | Candidates kept per Subject after shortlisting, and per layer neighbor of a placed pair |
| `alignment.TieCap` | 20 | estimate: room for a common name's ties without unbounded work | Most candidates kept when ties pass K, and per layer neighbor of a placed pair |
| `alignment.Alternatives` | 3 | `graphalign` `AlternativeLimit` | Alternatives shown per row |
| `alignment.SettleMargin` | 1 | estimate: one point, a third of a band | Lead a rival needs to displace a placed pair |
| `alignment.SettleRounds` | 10 | estimate: a safety cap; dirty pairs normally settle in two or three | Most settle rounds |
| `alignment.Shortlist.YearBucket` | 2 | estimate: the date tolerance, so a `Near` date shares a bucket | Width of an event's year bucket |
| `alignment.Shortlist.MaxBuckets` | 100 | estimate: two centuries at `YearBucket` 2; a wider reach keys as open | Most year buckets a date may key |
| `alignment.Shortlist.Words` | person `name`, place `toponym` | decided ([`comparison-layers.md`](deployment-plan/spike-11/comparison-layers.md#layer-6-procedure)) | Leads whose words key a Subject |
| `alignment.Shortlist.Dated.event.Type` | `event_type` | decided ([`comparison-layers.md`](deployment-plan/spike-11/comparison-layers.md#layer-6-procedure)) | The term keyed with each year bucket |
| `alignment.Shortlist.Dated.event.Date` | `date` | decided ([`comparison-layers.md`](deployment-plan/spike-11/comparison-layers.md#layer-6-procedure)) | The date whose reach is bucketed; a period stands for it |
| `kinship.MaxPivotEnd` | 10 | user, 2026-10-26 (S11-39): more than two, for unusual families, with a bound so one malformed event stays cheap | Most participants at either end of a pivot (an event, or the parent between in a chain) before it reads as malformed and implies nothing |
| `evidence.Checked` | `Same`, `Equivalent` | decided ([`claim-evidence.md`](deployment-plan/spike-11/claim-evidence.md#lines)) | Relations whose lines start checked |
| `evidence.ByEye` | `unjudged` | decided ([`claim-evidence.md`](deployment-plan/spike-11/claim-evidence.md#lines)) | Causes of an unknown line the researcher judges: kept, never checked |
| `evidence.Defining` | event `event_type` | decided ([`claim-evidence.md`](deployment-plan/spike-11/claim-evidence.md#evidence-through-relationships)) | Neighbor properties bundled into a relationship line |

## Not knobs

These are identifiers or catalog data, not tunables, and have no row: the vocabulary origins and primary kinds (`core/vocabkey`), `AliasAbbreviation`, `AliasNickname`, `valuecompare.WesternPattern` (a key), `value.Names.Aliases` (catalog rows for matching, the product seed for the reconciler through `valuerules.Product`) and `value.Names.Patterns` (catalog rows, passed in at run time), and `graphalign` `Config.Value` (the value section with catalog data, filled by `valuerules.Load`).

## Removed knobs

| Today | Why removed |
| --- | --- |
| `DefaultNames.GivenOnlyFactor`, `SuffixConflict` | A surname or first given name left over on both sides, and a generation conflict, are `Different` (S11-03) |
| `DefaultNames.Frequency`, `FrequencyDiscount`, `FrequencyMinPeople`; `graphalign` `MPrior`, `UPrior`, frequency `ValueTypeScale` entries | Frequency weighting is descoped |
| `graphalign` `PropertyAgreement` | Frequency weighting is descoped; the toponym bonus is folded into `subject.Weights.place.toponym.Agree` |
| `graphalign` `PropertyScale`, `ValueTypeScale` (weight and contradiction per property and per value type) | Promote scores properties with `subject.Weights` (S11-13); a property with no weight there scores nothing |
| `graphalign` `FanOutLowMax`, `FanOutUnknown`, the fan-out scan (`Stats.FanOut`, `sqlFanOut`) | Replaced by declared cardinality (`segment.One`) |
| `graphalign` `SignatureScale` | Became `segment.Gates` |
| `graphalign` `ValueTypeScale` floors other than date, `PropertyScale` sex / gender `Floor` 1 | Replaced by relation outcomes and the `Near` minimum |
| `graphalign` `DefaultConfig` (S11-16: the registry's `AlignConfig` replaces it) | The walk's numbers are `graphalign.Rules`, filled in the registry's alignment section |
| `promotealign` `seedCandidateLimit` | Became `alignment.TopK` |
| `graphalign` `AlternativeLimit` | Became `alignment.Alternatives` |
| `graphalign` `DefaultConfig.ConflictPenalty` | Value carried into `subject.Weights.event.date.Conflict` and `subject.Weights.place.toponym.Conflict`; its second role, the conflict-with-decision margin, is replaced by the contested marker and the settle margin |
| `graphalign` `mFor` fallback 0.85, `uOr` fallback 0.1 | Frequency weighting is descoped |
| `DefaultProfile` `start_date` / `end_date` features | Replaced by the `date` period in `subject.Periods.event.date` |
| `DefaultDates.ApproxToleranceFactor` | Replaced by `ABT` as a range |
| `DefaultProfile` weights and `MinScore` (`Rank`'s 10-point scale), `DefaultSuggestionLimit` | `Rank` is replaced by shortlisting and entity comparison; one weight set |
