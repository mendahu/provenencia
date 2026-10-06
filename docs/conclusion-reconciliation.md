# Provenencia — Conclusion Reconciliation

## Status

Design, agreed 2026-10-05. This document is authoritative for how a canonical entity's **displayed values** are derived from its members' evidence: the auto-reconcilers, their shared pipeline, their output, and how a Reconciliation Claim overrides them. Schema for claims stays in [`conclusion-layer-data-model.md`](conclusion-layer-data-model.md); judgment vocabulary in [`research-judgment-model.md`](research-judgment-model.md).

Implementation lags this design; see [§12](#12-implementation-status).

---

# 1. The problem

Everywhere the Conclusion layer is shown needs one value per Property, or an honest set of them:

- the Person list and Person detail
- Events and Places
- the family tree
- later visualizations

A canonical entity has many member Subjects, each with many Observations, and any Property may be observed more than once. So every Property on an entity is a list of candidates from several Sources. Each one has to be reconciled into what we display. **Every value type needs an auto-reconciler.**

Names and dates have the most interesting logic because they are structured, but text, integers, terms and subject references need reconciling too. Their candidates can still be weighed by their **evidence**: which Source, how well transcribed, how confident the identity, how many Sources agree.

**Vocabulary.** These names are used everywhere: code, tables, docs.

| Term | Meaning | In code |
| --- | --- | --- |
| **auto-reconciler** | the shared pipeline plus its value-type modules | `core/autoreconcile` (`Reconcile`) |
| **auto-reconciled value** | one distinct value it produced, displayed or not, with its reason | `ReconciledValue`; table `auto_reconciler_values` |
| **outcome** | what it did with one Observation (kept, folded, outvoted, …) | `Outcome`; table `auto_reconciler_outcomes` |
| **the auto-reconciler cache** | both tables, derived and rebuildable | `core/database/autoreconciler` |

"Resolve", "resolver", "resolved value" and "cluster" are retired for this concept.

---

# 2. Principles

1. **Display policy, never truth.** Auto-reconciliation is stateless projection. It never writes a claim, an Observation, a DateValue or a NameValue row ([`conclusion-layer-data-model.md`](conclusion-layer-data-model.md) §2.5, §7). Its output lives only in the derived auto-reconciler cache, rebuildable from truth tables.
2. **No stored scores.** Reconciliation weighs credibility, certainty and confidence but never multiplies them into a stored rollup ([`research-judgment-model.md`](research-judgment-model.md) §1.1). The cache stores order, counts and reasons, not a likelihood.
3. **A Reconciliation Claim always wins.** An accepted Reconciliation Claim on (entity, Property) is the displayed value, whatever the reconciler would say. A provisional claim is a working choice the UI shows differently. A rejected claim is history only. This holds for **every** value type.
4. **Be bold, and show why.** Because a researcher can always override with a claim, the reconciler may choose decisively. It must say what it considered and why each candidate won or lost, so the UI can explain it.
5. **The Observation normalizes; the reconciler doesn't interpret.** Turning what a Source said into a comparable value is the Observation's job. A cited age becomes a birth DateValue on the Observation, with UI help for first-class events such as birth and death. The reconciler compares values of one type and does not reach into another Property to reinterpret them.

---

# 3. Precedence

```text
accepted Reconciliation Claim(s) on (entity, Property) → the concluded value(s) (state: concluded)
otherwise                                                → the auto-reconciler's output
```

A claim's value is **never written into the auto-reconciler cache**: those tables hold only the auto-reconciler's own output. When claims ship, the composers that build lists and pages lay an accepted claim over the auto-reconciled value for that (entity, Property). (`autoreconcile.Reconcile` still accepts a concluded input, which joins the value it equals or leads alone, for callers that want one combined result.)

---

# 4. Inputs

A **candidate** is one Observation of the Property on a member Subject of the entity, with what it needs to be weighed:

| Input | From | Use |
| --- | --- | --- |
| Value | the Observation's typed value | compared by the value-type module (§7) |
| Polarity | `observations.polarity` | negatives deny or count against, never display |
| Source | Citation → Artifact → Source | majority counts **Sources**, not Observations |
| Source credibility | `source_credibility_assessments` (none = `standard`) | evidence strength |
| Transcription certainty | `citations.transcription_uncertain` | evidence strength |
| Identity Claim confidence | the member's claim (none = `moderate`) | evidence strength |
| Identity Claim status | the member's claim | membership (below) |

Credibility and confidence are read relative to their vocabulary's default grade by **sort order**, not by key. A candidate is **weak** if any of the three is below default. One candidate is **stronger** than another by comparing credibility, then certainty, then confidence.

**Membership status:**

- **Accepted** members are the evidence.
- **Provisional** members are passed in but **always eliminated**: they never influence the answer and appear only in the reasoning. That is stricter than weak evidence, which can still win when nothing stronger disagrees. Showing a provisional member's value as *possibly X* when nothing else survives is an idea for later: [`ideas/possible-values.md`](ideas/possible-values.md).
- **Rejected** members do not count and are not passed in.

---

# 5. The shared pipeline

One pipeline for every value type. Each value type supplies a small module (§7); everything else is shared. Each step records an outcome for the candidates it touches (§6).

1. **Admit.**
   - Provisional-member candidates are eliminated (`provisional`).
   - Candidates with no usable value are eliminated (`no_evidence`): an empty value, a name with no parts, a term the module calls no evidence (such as an *unknown* sex term).
2. **Deny.** A negative candidate eliminates the positive candidates with the **same value** whose evidence it is stronger than (`denied`). Those candidates cast no vote. A negative at equal or lower strength only counts against.
3. **Group.** The module's *same value* test groups equal values. Its *fold* test folds a less specific value into a more specific one it fits (`folded`): `[J]` into `[James]`, `MAY 1985` into `14 MAY 1985`. When a value fits several, it folds into the best supported, then the strongest.
4. **Majority.** Support is counted in **distinct Sources** (a derivative Source counts as its own Source until derivation can be recorded). A value with support from at least two Sources and more than half the support crowds out the rest (`outvoted`); a module may limit what it outvotes (names: spelling variants only, §7.2). Two of three wins; one to one keeps both.
5. **Confidence.** Among survivors, a value carried only by weak candidates is eliminated when a value with a non-weak carrier survived (`weak`).
6. **Merge.** The module merges each surviving group into one displayed value. Single-valued Properties have one or more surviving values (§8). When several survive, the state is *mixed* and all are shown.

Steps 2 to 5 run per comparable unit. For most types that unit is the value. For names it is each part type (§7.2).

---

# 6. Output

The reconciler returns the **answer and its reasoning**, so the UI can be rich about what happened:

- **State:** `single`, `merged`, `mixed`, `concluded`, or empty, derived from the shape, never stored beside it.
- **Values:** the surviving value or values in display order, each with its support (distinct Sources), its member Observations, and how many negatives count against it.
- **Every candidate considered,** with an **outcome** and a machine-readable **reason**:

| Reason | Meaning |
| --- | --- |
| `kept` | supports a displayed value as given |
| `folded` | merged into a fuller value (names the value) |
| `outvoted` | crowded out by a majority of Sources |
| `weak` | dropped: only weak evidence, and stronger evidence disagreed |
| `denied` | eliminated by a stronger negative Observation (names it) |
| `provisional` | from a provisional member; never counts |
| `no_evidence` | nothing usable to compare |
| `against` | a negative that counted against a value without eliminating it |

**Eliminated means not displayed, never dropped.** Every distinct value is cached, with a reason on each value (`kept` when displayed, else `outvoted`, `weak`, `denied` or `provisional`). The state and list counts read the displayed values; search and matching read them all, so an outvoted spelling still finds its Person.

The cache stores the reasoning with the values: state, support, the against count, and every candidate's outcome and reason. List and detail pages render from it without re-running the reconciler.

---

# 7. Value-type modules

Each module answers three questions:

- **Same value?** When are two values equal?
- **Fold?** When does a less specific value fold into a more specific one?
- **Merge?** How does a group become one displayed value?

| Value type | Same value | Fold | Merge | No evidence |
| --- | --- | --- | --- | --- |
| **text** | equal after trimming whitespace, ignoring case (*York* = *york*) | none | the best-ranked member's text | empty |
| **integer** | equal | none | the value | — |
| **term** | same term | none | the term | terms the module marks neutral (*unknown*, *indeterminate*) |
| **subject** | candidate Subjects map to the **same handle**; unpromoted Subjects drop out | none | the handle (an edge of the canonical graph) | unpromoted |
| **date** | §7.1 | precision containment | §7.1 | empty |
| **name** | §7.2 | subsumption per part type | §7.2 | no parts |

The same *same value* test (`autoreconcile.Compatible`) is what Promote alignment counts as an agreement ([`promote-alignment.md`](promote-alignment.md) §6), and what matching should use for agreement, so a Promote comparison and a Person page can never disagree about whether two values match.

## 7.1 Dates

From [`structured-date-model.md`](structured-date-model.md): missing components are unknown, never zero.

- A less precise point containing a more precise one folds into it (`MAY 1985` + `14 MAY 1985` → `14 MAY 1985`).
- Shared components are kept. Disagreeing finer components either widen to a range (`3 MAY 1985` + `14 JUN 1985` → `BET 3 MAY 1985 AND 14 JUN 1985`) or drop to the shared precision (`1985`). Which one is a test-case decision.
- `ABT`, `BEF`, `AFT` and ranges are the same value when their windows overlap.
- Years that disagree with no overlap stay separate, so the state reads *mixed*.

## 7.2 Names

From [`structured-name-model.md`](structured-name-model.md). Built in S9-13b (see §12):

- **Parts only.** `form` is a transcription and is never compared. A name with no parts is `no_evidence`.
- **Format-agnostic.** A part type is only an identifier: parts are compared only with parts of the same type, and no type behaves differently from another. A prefix, a surname and a given name are all just types. Culture lives in name format profiles, not here.
- **The comparable unit is the part type.** A candidate's value for a type is its ordered list of **words** of that type: each part normalized (case, punctuation and whitespace ignored; accents are not folded) and split into words. A hyphen or a space inside a part is the same as separate parts: *Smith-Jones* = *Smith Jones* = *Smith* + *Jones*. Words are compared; the best-ranked record's recorded parts are displayed.
- **Fold (subsumption):** a list folds into a fuller one when its parts map, in order, onto a subsequence, each equal or an initial of it: `[J]` → `[James]`, `[James]` → `[James, Kenneth]`, `[J, K]` → `[James, Kenneth]`. The `initial` part type is retired; an initial is the part it stands for.
- **Deny** compares whole names: a negative denies positives with the same parts by type.
- **Majority only outvotes misspellings** (decided 2026-10-05). Within a type, a majority winner outvotes a minority value only when it is a **spelling variant** of the winner: the same number of parts, each equal or similar (`SpellingSimilarity`, shared with matching: at least 0.8 similarity by edit distance with adjacent swaps, or one added or dropped letter from 3 letters up). *Robbins* beside *Robins* is outvoted. *Jake* beside *James*, or a married *Smith* beside *Robins*, never is: sources often record a nickname as a given name and seldom say so.
- **Weak evidence still drops** within a type, and denial still applies.
- **One name, never mixed.** Every displayed candidate contributes to a single NameValue. Each type keeps every surviving value, best supported first, a recorded part skipped when all its words are already there: *given Jake* + *given James* → `given: [Jake, James]`; a maiden *Smith* folds into a married *Smith-Jones*; *nick Jake* + *given James* → `given: [James], nick: [Jake]`. The state is single or merged. The name is assembled in the type order of the member with the most types, or is a member's own value when one carries exactly those parts. Outvoted, weak, denied and provisional names keep their own cached rows.
- **Single-valued at the structure level.** A Person has one concluded NameValue, and the auto-reconciler produces one too. Multiple given names, surnames and so on live inside it as multiple parts. There is no second independent name structure.
- **The form reads the parts in order** (*Jake James Robins*). There is no separate middle-name type, so the form doesn't distinguish alternatives from middle names; name display styles and the *Why* view do that work.

---

# 8. Cardinality

Each Property needs a **cardinality** in its configuration:

Most facts that repeat or change over time are better modelled another way: occupations, residences, religion and military service as **events** with spans (`residence` is already an event type), several given names as **parts** of one name, family links as **associations**. Cardinality is for what is left: concurrent place names, and Properties researchers create. External identifiers stay single-valued with one Property per identifier system (two values in one system usually means a duplicate, so *mixed* is the right signal).

- **Single** (sex at birth, a date, name): the pipeline aims for one value. When the evidence can't decide, several survive and the state is *mixed*.
- **Multiple** (`toponym`: Montréal and Montreal, Tkaronto and Toronto; researcher-created Properties such as languages spoken): several distinct values are all true. The pipeline merges duplicates and variants (same value, fold, deny) and does not crowd out a distinct value by majority. The confidence step still applies: a distinct value carried only by weak evidence is dropped when stronger evidence supports a different value. Promoting weak evidence to a concluded value is an explicit researcher action (a Reconciliation Claim), so the auto-reconciler stays opinionated.

Model consequences:

- The Property needs a configuration setting. Seeded Properties get a default.
- **Reconciliation Claims for a multi-valued Property are one per value:** at most one claim per (entity, Property, value), each with its own evidence, confidence and argument. A single-valued Property keeps one claim per (entity, Property).
- **Concluded values replace the list.** Once any value of a multi-valued Property is concluded, only concluded values are displayed and the auto-reconciler steps aside for that Property. The UI offers to conclude a second or third value. This is the same rule as single-valued: a claim always wins.
- Changing a Property's cardinality changes resolution for every entity carrying it, so the cache must recompute (§10).

---

# 9. Values reached through other entities

Many displayed values come from another entity. A Person's birth date is their birth Event's date, and a place in the tree is an Event's Location. Order:

1. Each entity reconciles its own Properties, association ends included (the `subject` module).
2. Derived values read the reconciled values of the entities they point at.

The model allows a Person with two birth Events, and the reconciler must handle it: the derived value reconciles across the Events' reconciled values. The **UI should push the researcher toward a cleaner state**, merging duplicate Events into one canonical Event so reconciliation happens inside it.

---

# 10. Cache and upkeep

The reconciler's output is stored in the auto-reconciler cache (Spike 9 R3, `core/database/autoreconciler`) and rewritten in the transaction of every write that can change it. A change to reconciliation rules is a cache version bump, which rebuilds on open.

Writes that recompute affected entities:

- Observation save or delete, Subject delete, Promote (claim create)
- Source credibility change, Citation certainty change
- Identity Claim status or confidence change (no edit path exists yet)
- **Property configuration change** (cardinality, and any per-Property reconciliation setting)
- Reconciliation Claim write (when claims ship)

The rebuild-equals-upkeep tests hold upkeep equal to a full rebuild. Every new trigger joins them.

---

# 11. Open questions

1. **Places.** Decided in §11.1. Labels for the seeded types are a UI decision for later.

**Decided 2026-10-05:**

- **Text ignores case.** This changes Spike 9 S9-05's exact-after-trimming rule.
- **Weak evidence drops in multi-valued Properties too** (§8). A researcher who wants a weak value concluded makes a Reconciliation Claim.
- **Claims for multi-valued Properties:** one claim per value; concluded values replace the auto-reconciled list (§8).
- **Derivative Sources count as separate votes, for now.** There is no way to record that an index derives from a register, so the app treats them as two Sources. Counting them once waits for source-to-source relationships ([`ideas/source-to-source-relationships.md`](ideas/source-to-source-relationships.md)).
- **Reason vocabulary:** the §6 list, extended as modules need.

## 11.1 Places (decided 2026-10-05)

- **Separate places with relationships, not one composite place.** "Toronto, Ontario, Canada" is three Places linked upward, so queries like "everyone born in Ontario" work.
- **A place is whatever the research needs.** A township, a county, a region (the Lower Mainland), a family farm. No rules about what can contain what. A place **kind** (township, county, province) is left for later: the hierarchy already says what a place is part of.
- **Places relate through a place relationship**, an association kind alongside Location and Participation, with three Properties:

  | Property | Value | Meaning |
  | --- | --- | --- |
  | `from` | place | the part, or the predecessor |
  | `to` | place | the whole, or the successor |
  | `place_relationship_type` | term | what the relationship is |

  Each relationship is cited and reconciled like any other evidence; the ends use the subject-valued module (§7). It is a dedicated kind, not the Person Relationship, because its ends are two directional places. It can carry more Properties later without remodelling.
- **Types are an open, researcher-extensible vocabulary, and each type has a category:**
  - **hierarchical:** `administrative` (Guam in the United States), `geographic` (the Lower Mainland in British Columbia), `ecclesiastical` (a parish in a diocese), and any a researcher adds. These follow a chain upward ("everyone in Ontario"), hold where the two places' periods overlap, and build display chains ("Toronto, Ontario, Canada"). A place may have several parents, so places form a graph, not a tree.
  - **temporal:** `succeeded_by` (York succeeded by Toronto), and any a researcher adds. These link a lineage that search may follow. They never build a display chain, and containment is not inherited across them.

  The category is data on the vocabulary term (a new field), so a researcher-added type tells the app how to behave.
- **No loops.** Hierarchical relationships must not form a cycle, and temporal ones are directional. Both are checked in the app layer, not the schema.
- **Periods live on Places, not on links.** A Place has a **period** when it existed or was meaningful (a city from incorporation, a country from independence, a farm until it was sold). Start and end are both optional structured dates; a place with no period is always valid.
- **A link holds where the two places' periods overlap.** Toronto (1834–) is part of Upper Canada (1791–1841) until 1841, the Province of Canada (1841–1867) until 1867, and Ontario (1867–) after. So "what is Toronto part of?" needs a date, and the periods answer it.
  - **Accepted imprecision:** when both places persist and the link changes (Guam, Spanish until 1898 and American after), the link reads as holding for the whole overlap, and Guam is part of both. Where that matters, model the jurisdictions as distinct Places with their own periods (the Kingdom of Spain, the American colonies, the United States).
- **A rename is a new Place.** York (1793–1834) and Toronto (1834–) are two Places. Merge is only for two handles that turn out to be the same place.
- **Succession may branch.** A split is one Place `succeeded_by` several (a county divided in two); an amalgamation is several Places `succeeded_by` one (townships joined into a city). A lineage has any number of predecessors and successors.
- **Concurrent names are one Place.** `toponym` is multi-valued (§8): Montréal and Montreal, Köln and Cologne, Tkaronto and Toronto are all names of one Place at the same time. Spelling duplicates still merge.
- **Succession.** A second place-to-place relationship records that one place became another (York → Toronto), so history and search can treat them as one lineage while they stay distinct Places. A query like "born in Toronto" may follow it.
- **Display uses the connected event's date** to pick the parent chain. When the date can't decide (an approximate or ranged date straddling a change), show every candidate: *born about 1841 in Toronto, Upper Canada or Province of Canada*.
- **Evidence, the hard way, for now.** Every place and every "part of" link is created by hand with a cited Source, even for common knowledge ("Alberta is part of Canada"). Researcher knowledge uses a researcher-knowledge Source with an argument. The model does not change for convenience.
- **The easy way later.** A Provenencia places service, likely paid, would write the same evidence automatically from gazetteers: [`ideas/place-gazetteer-service.md`](ideas/place-gazetteer-service.md).

**Out of scope for now:** values that change over time (a married name from 1885, an occupation by decade). Cardinality is where they would hook in later.

---

# 12. Implementation status

| Piece | Where | State |
| --- | --- | --- |
| Auto-reconciler core: exact clusters, support then id, concluded input | `core/autoreconcile` (S9-05) | on `main` |
| Auto-reconciler cache and upkeep | `core/database/autoreconciler` (S9-06) | on `main` |
| `initial` part type retired (migration 000037); `against` column (000038) | S9-13a | on `main` |
| Shared pipeline and simple modules; every value cached with its reason (migration 000039) | S9-13 | open PR #259 |
| Name module: parts by type, subsumption, one name per Person, majority outvotes misspellings only, parts compared as words; cache version 8 | S9-13b | open PR #260, stacked on S9-13 |
| Evidence loaded (Sources, provenance, negatives, provisional members); outcomes cached in `auto_reconciler_outcomes`; credibility and certainty upkeep; the "auto-reconciler" naming; cache version 9 | S9-14 | stacked on S9-13b |
| Detail read: each field's state, values and outcomes with their evidence (`conclusiondetails`, `GetConclusionDetail`); Swift wording of states and outcomes; off-screen detail pages evicted | S9-15 | stacked on S9-14 |
| An outvoted outcome keeps the vote that beat it (migration 000041, cache version 10); the Person page shows every field's state, values and Why | S9-16 | stacked on S9-15 |

PRs #255 and #256 built names-first versions of S9-13b and S9-14 to the first plan and were closed. Their migrations landed unchanged in S9-13a; their name logic, provenance logic, fixtures and test tables (75+ cases) are lifted into S9-13, S9-13b and S9-14 (see the Spike 9 plan, slice 4).
