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

---

# 2. Principles

1. **Display policy, never truth.** Auto-reconciliation is stateless projection. It never writes a claim, an Observation, a DateValue or a NameValue row ([`conclusion-layer-data-model.md`](conclusion-layer-data-model.md) §2.5, §7). Its output lives only in the derived resolved-values cache, rebuildable from truth tables.
2. **No stored scores.** Reconciliation weighs credibility, certainty and confidence but never multiplies them into a stored rollup ([`research-judgment-model.md`](research-judgment-model.md) §1.1). The cache stores order, counts and reasons, not a likelihood.
3. **A Reconciliation Claim always wins.** An accepted Reconciliation Claim on (entity, Property) is the displayed value, whatever the reconciler would say. A provisional claim is a working choice the UI shows differently. A rejected claim is history only. This holds for **every** value type.
4. **Be bold, and show why.** Because a researcher can always override with a claim, the reconciler may choose decisively. It must say what it considered and why each candidate won or lost, so the UI can explain it.
5. **The Observation normalizes; the reconciler doesn't interpret.** Turning what a Source said into a comparable value is the Observation's job. A cited age becomes a birth DateValue on the Observation, with UI help for first-class events such as birth and death. The reconciler compares values of one type and does not reach into another Property to reinterpret them.

---

# 3. Precedence

```text
accepted Reconciliation Claim on (entity, Property)   → that value (state: concluded)
otherwise                                              → the auto-reconciler's output
```

The resolver takes the concluded value as an input so the claim plugs in without changing callers. The concluded value joins the cluster it equals, or leads alone with no support.

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
- **Provisional** members are passed in but **always eliminated**: they never influence the answer and appear only in the reasoning. That is stricter than weak evidence, which can still win when nothing stronger disagrees.
- **Rejected** members do not count and are not passed in.

---

# 5. The shared pipeline

One pipeline for every value type. Each value type supplies a small module (§7); everything else is shared. Each step records an outcome for the candidates it touches (§6).

1. **Admit.**
   - Provisional-member candidates are eliminated (`provisional`).
   - Candidates with no usable value are eliminated (`no_evidence`): an empty value, a name with no parts, a term the module calls no evidence (such as an *unknown* sex term).
2. **Deny.** A negative candidate eliminates the positive candidates with the **same value** whose evidence it is stronger than (`denied`). Those candidates cast no vote. A negative at equal or lower strength only counts against.
3. **Group.** The module's *same value* test groups equal values. Its *fold* test folds a less specific value into a more specific one it fits (`folded`): `[J]` into `[James]`, `MAY 1985` into `14 MAY 1985`. When a value fits several, it folds into the best supported, then the strongest.
4. **Majority.** Support is counted in **distinct Sources**. A value with support from at least two Sources and more than half the support crowds out the rest (`outvoted`). Two of three wins; one to one keeps both.
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

The cache stores enough of this for list and detail pages to render without re-running the reconciler: values, state, support, the against count, and the per-candidate outcomes.

---

# 7. Value-type modules

Each module answers three questions:

- **Same value?** When are two values equal?
- **Fold?** When does a less specific value fold into a more specific one?
- **Merge?** How does a group become one displayed value?

| Value type | Same value | Fold | Merge | No evidence |
| --- | --- | --- | --- | --- |
| **text** | equal after trimming (case folding is open, §11) | none | the best-ranked member's text | empty |
| **integer** | equal | none | the value | — |
| **term** | same term | none | the term | terms the module marks neutral (*unknown*, *indeterminate*) |
| **subject** | candidate Subjects map to the **same handle**; unpromoted Subjects drop out | none | the handle (an edge of the canonical graph) | unpromoted |
| **date** | §7.1 | precision containment | §7.1 | empty |
| **name** | §7.2 | subsumption per part type | §7.2 | no parts |

The same *same value* test is what Promote uses to start compatible pairs checked (Spike 9 R7), and what matching should use for agreement, so a Promote comparison and a Person page can never disagree about whether two values match.

## 7.1 Dates

From [`structured-date-model.md`](structured-date-model.md): missing components are unknown, never zero.

- A less precise point containing a more precise one folds into it (`MAY 1985` + `14 MAY 1985` → `14 MAY 1985`).
- Shared components are kept. Disagreeing finer components either widen to a range (`3 MAY 1985` + `14 JUN 1985` → `BET 3 MAY 1985 AND 14 JUN 1985`) or drop to the shared precision (`1985`). Which one is a test-case decision.
- `ABT`, `BEF`, `AFT` and ranges are the same value when their windows overlap.
- Years that disagree with no overlap stay separate, so the state reads *mixed*.

## 7.2 Names

From [`structured-name-model.md`](structured-name-model.md). Designed and built in S9-13 / S9-14 (see §12):

- **Parts only.** `form` is a transcription and is never compared. A name with no parts is `no_evidence`.
- **Format-agnostic.** A part type is only an identifier: parts are compared only with parts of the same type, and no type behaves differently from another. A prefix, a surname and a given name are all just types. Culture lives in name format profiles, not here.
- **The comparable unit is the part type.** A candidate's value for a type is its ordered list of parts of that type, each part one normalized unit (case, punctuation and whitespace ignored; accents are not folded).
- **Fold (subsumption):** a list folds into a fuller one when its parts map, in order, onto a subsequence, each equal or an initial of it: `[J]` → `[James]`, `[James]` → `[James, Kenneth]`, `[J, K]` → `[James, Kenneth]`. The `initial` part type is retired; an initial is the part it stands for.
- **Deny** compares whole names: a negative denies positives with the same parts by type.
- **Survivors to names.** A candidate survives when all its part values survived. Survivors that agree on every type they share form one displayed name. A candidate missing a type joins the agreeing name with the most members. The displayed name is assembled from each type's fullest value, or is a member's own value when one carries exactly those parts.
- **Single-valued at the structure level.** A Person has one concluded NameValue. Multiple given names, surnames and so on live inside it as multiple parts. There is no second independent name structure.

---

# 8. Cardinality

Each Property needs a **cardinality** in its configuration:

- **Single** (sex at birth, birth date, name): the pipeline aims for one value. When the evidence can't decide, several survive and the state is *mixed*.
- **Multiple** (occupation, residence, religion): several distinct values are all true. The pipeline merges duplicates and variants (same value, fold, deny) but must not crowd out a distinct value as an outlier.

Model consequences:

- The Property needs a configuration setting. Seeded Properties get a default.
- A Reconciliation Claim is "one per (entity, Property)" today. A multi-valued Property needs a claim shape that can conclude several values. Open (§11).
- Changing a Property's cardinality changes resolution for every entity carrying it, so the cache must recompute (§10).

---

# 9. Values reached through other entities

Many displayed values come from another entity. A Person's birth date is their birth Event's date, and a place in the tree is an Event's Location. Order:

1. Each entity reconciles its own Properties, association ends included (the `subject` module).
2. Derived values read the reconciled values of the entities they point at.

The model allows a Person with two birth Events, and the reconciler must handle it: the derived value reconciles across the Events' reconciled values. The **UI should push the researcher toward a cleaner state**, merging duplicate Events into one canonical Event so reconciliation happens inside it.

---

# 10. Cache and upkeep

The reconciler's output is stored in the resolved-values cache (Spike 9 R3, `core/database/resolvedvalues`) and rewritten in the transaction of every write that can change it. A change to reconciliation rules is a cache version bump, which rebuilds on open.

Writes that recompute affected entities:

- Observation save or delete, Subject delete, Promote (claim create)
- Source credibility change, Citation certainty change
- Identity Claim status or confidence change (no edit path exists yet)
- **Property configuration change** (cardinality, and any per-Property reconciliation setting)
- Reconciliation Claim write (when claims ship)

The rebuild-equals-upkeep tests hold upkeep equal to a full rebuild. Every new trigger joins them.

---

# 11. Open questions

1. **Places.** A Place's toponyms are text, but "Toronto", "Ontario" and "Canada" are not alternatives to reconcile: they are different places at different grains. Place reconciliation needs a place hierarchy (`contained_in`, gazetteer parents; [`conclusion-layer-data-model.md`](conclusion-layer-data-model.md) §13) before a module makes sense. Descoped from Spike 9 slice 4; next to design.
2. **Claims for multi-valued Properties.** What a Reconciliation Claim concludes when a Property has several true values.
3. **Text normalization.** Whether text equality folds case (*York* vs *york*). Spike 9 S9-05 decided exact after trimming.
4. **Confidence in multi-valued Properties.** Whether a weak distinct value is dropped when stronger evidence supports a *different* value, given both may be true.
5. **Derivative Sources.** Majority counts Sources. An index copied from a register is arguably one Source's evidence twice. That needs source-to-source relationships ([`ideas/source-to-source-relationships.md`](ideas/source-to-source-relationships.md)).

**Out of scope for now:** values that change over time (a married name from 1885, an occupation by decade). Cardinality is where they would hook in later.

---

# 12. Implementation status

| Piece | Where | State |
| --- | --- | --- |
| Resolver core: exact clusters, support then id, concluded input | `core/resolve` (S9-05) | on `main` |
| Resolved-values cache and upkeep | `core/database/resolvedvalues` (S9-06) | on `main` |
| Name module: parts only, format-agnostic, subsumption, majority; `initial` retired | S9-13, PR #255 | open, unmerged |
| Provenance, negatives, confidence pass for names; `against`; credibility and certainty upkeep | S9-14, PR #256 (stacked) | open, unmerged |

#255 and #256 match this design's direction but not its shape:

- The passes live inside the name code instead of a shared pipeline.
- Majority counts Observations, not Sources.
- There is no per-candidate reasoning output.
- Provisional members are not passed in.

Their test tables (75+ cases) carry over as the spec for the name module.
