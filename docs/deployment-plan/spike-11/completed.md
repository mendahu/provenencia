# Spike 11 — Completed

One short paragraph per landed PR, newest first. Written by the PR that lands it.

## S11-15 — Shortlist index

**Landed 2026-10-19** in [#497](https://github.com/mendahu/provenencia/pull/497). Promote no longer scores a Source Subject against every handle of its kind. A pure index (`core/shortlist`) maps a person's name words, an event's type with the year buckets its date can still be near, and a place's toponym words to handles; value compare makes every key, so shortlisting never misses a pair it calls the same. The index is built once per catalog state, holds no life events, and is dropped with a handle commit or new alias tables. A name or toponym that differs only by a misspelling, with no shared word, is no longer offered; the walk still reaches it through a mapped neighbor.

## S11-14 — Segment compare

**Landed 2026-10-19** in [#493](https://github.com/mendahu/provenencia/pull/493). Promote judges each link with segment compare. A link is a typed key, so a spouse never stands for a sibling and a parent link reads the same from either record. Subject of a birth or a death is one per person and one per event: a birth placed on a different event, or on a different person, subtracts 3. Every other link earns the same small support whatever the catalog holds; the fan-out scan is gone.

## S11-13 — Migrate Promote scoring onto subject compare

**Landed 2026-10-18** in [#487](https://github.com/mendahu/provenencia/pull/487). Promote now judges a Source Subject against a handle with entity compare: each side is built once per proposal (the handle from its conclusion, displayed values positive and denied values negative; a person with the dated events they take part in), and every property scores its subject section weight. A birth is never filed on a death, a person who died before the Subject was born is never proposed, and sex at birth and gender compare as one. A handle is offered only when its own properties reach the weak band. Event types now earn 0.5 and cost 20. The lifespan counts but is not shown until S11-20. `core/match` and `core/database/matching` are gone.

## S11-24 — Delete frequency weighting

**Landed 2026-10-18** in [#483](https://github.com/mendahu/provenencia/pull/483). Promote no longer weighs a value by how common it is in the catalog. Every property scores a plain weight from the walk's registry: 2 for each value type and 3 for a toponym, so an exact toponym alone is still medium and a sparse name and birth date still files strong. `Stats.ValueFreq`, the priors, the lift to the accept bar and the value-frequency scan are gone; the fan-out scan stays until S11-14.

## S11-12 — Lifespan for people

**Landed 2026-10-18** in [#480](https://github.com/mendahu/provenencia/pull/480). People now have a lifespan. The profile builder lays every dated event a person takes part in on one timeline; birth starts it, death or burial ends it, and an open end is padded by 110 years. Subject compare checks each side's dated events against the other side's window, so a person who died before a Subject was born, a 140-year life, or a residence after a death is a large conflict. The four age-rule knobs were dropped. No caller yet; nothing changes for researchers.

## S11-11 — Subject profile and entity compare

**Landed 2026-10-18** in [#476](https://github.com/mendahu/provenencia/pull/476). The profile builder exists: `subjectprofile.Build` turns an entity's reconciled values into a Subject-shaped profile, with displayed values as positives, denied values as negatives and nothing else, so entity compare is `Build` then `subjectcompare.Compare`. `Options.Excluding` reconciles again without some members' Observations, for leave-one-out checks, and `ForSubject` builds a plain Subject's side. No caller yet; nothing changes for researchers.

## S11-10 — Subject compare

**Landed 2026-10-18** in [#471](https://github.com/mendahu/provenencia/pull/471). Layer 3 of matching exists: `subjectcompare.Compare` says how alike two Subjects of one type are as a similarity (0 to 1) and evidence (points), from layer 2's outcome on each property the registry's new `subject` section lists. Sex at birth and gender are pooled, and an event with no date compares its start and end as one period. The section keys by the new `vocabkey.SubjectType` and `PropertyKey`, and a seed test ties it to the product vocabulary. `valuecompare.EndsDate` turns date ends back into a date, and `propertycomparetest` and `subjectcomparetest` hold the layers' test builders. No caller yet; nothing changes for researchers.

## S11-09 — Property compare

**Landed 2026-10-18** in [#468](https://github.com/mendahu/provenencia/pull/468). Layer 2 of matching exists as `core/propertycompare`: given every value two Subjects record for one property, it drops neutral values, lets a negative value contradict a value it matches (`Same`, `Equivalent`, `Contains`), compares every pair of positive values with value compare, picks the pair that speaks (outcome, then similarity, then the stronger relation, then each side's `Same` key; a tie keeps side A's input order, contradictions included), treats a difference on a multiple-value property as no conflict, and returns agree, partial, conflict or unknown with a cause, the speaking pair and the excluded values. It holds no number and reads no refs, support or standing. `valuecompare.Stronger` is now the one relation order (`valuecompare.Relations()` lists it, and a test fails until a new relation has an outcome), `propertymeta.Meta` carries a property's value type and cardinality, and `datevaluetest` builds date fixtures. No caller uses the layer yet; nothing changes for researchers.

## S11-08 — One normalization helper in Swift

**Landed 2026-10-18** in [#466](https://github.com/mendahu/provenencia/pull/466). Every filter-as-you-type list ignores case, accents, punctuation and word breaks, the way search does: `jose` finds `José` in the Sources list, Properties, a combo box, a table's type-select and the graph canvas rotor, and `godfather` finds `God-father`. The combo box marks the span that matched, and the name splitter peels a title however it is punctuated (`Dr.`, `Jr.,`), while dotted initials such as `S.R.` stay names. `TextMatch` (`macos/App/Platform/TextMatch.swift`) is the one Swift rule, a mirror of `valuecompare.TextKey`; `TestSwiftTextMatchMirrorsTextKey` reads its fixtures and fold table so a change on either side fails in Go. `FakeStore` groups a Person's names by parts at `Same`, as the reconciler does, and searches like Go's rescoring.

## S11-07 — One normalization key across the core

**Landed 2026-10-18** in [#460](https://github.com/mendahu/provenencia/pull/460). Search, the search index, the term duplicate check and the alias tables use value compare's text key. Text ignores where words break, so `god-father` and `godfather`, and `York Town` and `Yorktown`, are the same text (the reconciler groups them; `CacheVersion` 17). Searching `Montreal` ranks `Montréal` like an exact hit, and a Person or Place search document lists each alternate name once (`ProjectionVersion` 8). Creating or renaming a term to one the Property already has, of any origin, is refused with the existing term's label. Alias tokens are stored and indexed as the comparison tokens names produce, through `valuecompare.AliasToken`.

## S11-06 — Migrate the reconciler onto value compare

**Landed 2026-10-17** in [#456](https://github.com/mendahu/provenencia/pull/456). The reconciler decides "the same value" with value compare: values group when it says `Same` and fold when it says `Contains`, and `Equivalent` (an accent fold, an alias) never groups. Text that differs only by case, punctuation or spacing is one value (`Sainte-Anne` / `Sainte Anne`) while `Montréal` and `Montreal` stay two, `Mc Donald` and `McDonald` are one surname, word order inside a name part type no longer splits a value, and `ABT 1985` reads as 1984 – 1986 and folds into `1986`. Place membership picks the later start and earlier end by day. `Reconcile` and `Compatible` take `valuecompare.Rules`, Promote passes the rules it holds, the cache reads the shipped registry (`CacheVersion` 16), and `valuecompare.DateWindow` replaces `DateBounds` and `WindowOfDate`; `PeriodWindow` and `LinkMembership` moved to `conclusionheaders`. The reconciler's own key and fold code is deleted.

## S11-05 — Migrate matching onto value compare

**Landed 2026-10-17** in [#451](https://github.com/mendahu/provenencia/pull/451). Promote and `Rank` now compare values with value compare under the registry's value rules, with the catalog's name pattern and alias tables passed in as `Rules` by one loader, `valuerules.Load`. A name with words on one side only (`Robins` / `James Robins`) is a resemblance below a full match, a different first given name is no match at all, `ABT 1817` reads as 1816 – 1818, toponyms compare in word order, and a common surname costs nothing extra. `match`'s comparers, the name frequency discount (`Stats.Words` included) and Promote's per-type floors are deleted; one outcome rule, `match.OutcomeOf`, serves `Evaluate` and `ApplyScale`. A Subject whose only look-alike has another first name now proposes New (reason `no_match`, no alternatives) where it proposed Skip with alternatives, so such a record is no longer offered in the Promote target menu. The matching skill, `matching.md` and the other docs that named the deleted code are corrected.

## S11-03 — Value compare: names

**Landed 2026-10-17** in [#447](https://github.com/mendahu/provenencia/pull/447). Value compare now relates two names by their parts: `Same`, `Equivalent`, `Contains`, `Near`, `Different` or `Unknown`, a similarity, and a detail that lists which word paired with which, in which role, and why. Initials and alias tables joined value compare's word similarity, and a name `Key` agrees with `Compare`. The name types moved to `core/valuecompare`, the product alias and nickname seeds with their license to `core/database/namealiases`, and the name knobs into the registry; `core/match` reads them from there and shares the pairing and alias lookup. A name of more than `value.Names.MaxWords` (32) words is `Unknown` (`too_long`), and the alias index is built once per catalog load. No caller uses the new comparer yet; nothing changes for researchers.

## S11-04 — Value compare: dates

**Landed 2026-10-17** in [#445](https://github.com/mendahu/provenencia/pull/445). Two dates can now be related without the database. `valuecompare.Compare` reads each date as a span of days and returns `Same`, `Contains`, `Overlaps`, `Near`, `Different` or `Unknown`, a similarity, and the broader side and the gap in days. `ABT` reads as a range of one unit of its most precise part, so `ABT 1817` contains `1818`, replacing "ABT doubles the tolerance". The date ends and the window algebra (`DateEnds`, `Window` with `Overlaps`, `Contains`, `Intersect`) moved from `core/autoreconcile` into value compare, and the civil calendar arithmetic into `core/datevalue/calendar.go`; `HoldKind`, `Relate` and `HoldsAt` are gone, since `HoldsAt` was `Overlaps`. The date knobs joined the registry's value section and `core/match` reads them. `match.DateComparer` still scores Promote until S11-05, so nothing changes for researchers.

## S11-02 — Value compare core and match registry

**Landed 2026-10-17** in [#441](https://github.com/mendahu/provenencia/pull/441). Two values of one type can now be related without the database. `core/valuecompare` takes text, term and integer values and returns a relation, a similarity and a structured detail (what paired with what, what is on one side only, why), and gives each value a normalized `Key` that agrees with `Compare`. Text compares words in order, so `Upper Canada, York` and `York, Upper Canada` are `Different`. Text longer than `value.Text.MaxWords` words that is not the same or equivalent is `Unknown` (`too_long`), and a word over `value.Words.MaxSpellingLength` runes pairs only when equal or equal after folding, so a long remark cannot exhaust memory. `core/matchregistry` opens with `Default()` and the value section's text, term and integer knobs; `core/match` reads the same numbers from it. The word helpers moved out of `core/autoreconcile` into value compare. `TestLayerPackages` keeps the layer packages free of catalog imports and stray numbers, `TestKnobInventoryMatchesRegistry` keeps the knob inventory in step with the registry, and the `add-match-rule` skill writes down the steps each layer PR repeats. Nothing changes for researchers.

## S11-23 — One home for primary kinds

**Landed 2026-10-17** in [#438](https://github.com/mendahu/provenencia/pull/438). The primary Subject kinds (`person`, `event`, `place`), the eight seeded Property keys matching reads by name, and the vocabulary origins now live in `core/vocabkey`, which imports only `strings`. The ten per-package `OriginProvenencia` declarations, six `originOK` copies, and the hand-written kind lists in Promote, progress, alignment, search and the Swift fake are gone; `search.KindSpec` gains `Handle` to mark handle kinds. `TestVocabKeysMatchSeed` ties the constants to the seed and `TestSwiftVocabularyMatchesGo/primary_kinds` ties them to `EvidencePrimaryKind`. Nothing changes for researchers.

## S11-22 — Pure value types

**Landed 2026-10-17** in [#434](https://github.com/mendahu/provenencia/pull/434). Dates, names and property types can now be used without the database. The date value, the name value with its part-type registry, and the property value-type and cardinality constants moved into `core/datevalue`, `core/namevalue` and `core/propertymeta`, which import nothing from `core/database`, so they build and test with no build tags. Validation is exported (`datevalue.Validate`, `namevalue.Validate`) and each package owns its `ErrInvalid`. The database packages keep only their queries; nothing changes for researchers.

## S11-01 — One-page matching model

**Landed 2026-10-17** in [#432](https://github.com/mendahu/provenencia/pull/432). The matching algorithm now has one page to read. [`matching-model.md`](../../matching-model.md) draws the six layers in one diagram, defines the terms (relation, similarity, evidence, segment, shortlist, candidate, band, contested), walks one Promote row from value compare to band with real numbers, and lists every registry knob with its section, starting value or estimate, and what it changes. The model also fixes lifespan under `subject` (`subject.person.lifespan.*`) and keys the sex rows on `sex_at_birth`, which covers `gender` through the pair. `matching.md` and `promote-graph-alignment.md` carry banners saying which sections Spike 11 replaces and that they describe the running code until then. Docs only; no code changed.
