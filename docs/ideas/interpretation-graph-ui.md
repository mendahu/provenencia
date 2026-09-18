# Interpretation graph (brainstorm)

## Status

**Brainstorm, with working decisions — now scheduled through the canvas prototype.** Not a UI spec, but no longer open-ended: the design questions below were worked through and settled; slice 1 shipped as [Spike 5](../deployment-plan/archive/spike-5/); slice 2 (canvas UI risk) is [Spike 6](../deployment-plan/spike-6/).

Authoritative schema for everything described here is [`interpretation-layer-data-model.md`](../interpretation-layer-data-model.md). Client rules are [`macos-client-patterns.md`](../macos-client-patterns.md). Nothing in this file overrides those — where this note reaches a conclusion that would change a model doc, that edit has to be made there deliberately.

## Decisions so far

| # | Decision | Where |
| --- | --- | --- |
| 1 | The primary surface is a Source-scoped spatial canvas called the **Evidence graph**, not a list/detail pair per table | §1.2, §12 |
| 2 | Root bubbles (person / event / place) are placed freely; **connect** writes a bridge subject plus its edges as one atomic macro, with a required disambiguation step | §3.1, §3.2 |
| 3 | Draft status needs **no schema flag** — "uncited" is a query for Subjects with no Observations; `subjects.label` is the free working handle | §4.2 |
| 4 | Citation → Observation stays **one-to-many**, with one-to-one as the default *behavior* and pinning opt-in | §4.5 |
| 5 | Cross-source Observations stay legal in the model (removing them costs a trigger), but the canvas scopes to `source_id = ?` | §4.6 |
| 6 | **`source` subjects never render on the canvas.** Source-to-source commentary lives on the Source page. UI decision, no schema change | §4.6 |
| 7 | Layout is **one unaudited table** keyed by `subject_id` with integer grid cells; positions travel with the project, camera does not | §5.1, §5.2 |
| 8 | No auto-layout engine — unplaced subjects go in a **tray** | §5 |
| 9 | Pan/zoom needs **AppKit `NSScrollView`** regardless of deployment target; raising past macOS 14 is a separate decision on support-matrix grounds | §7.2, §7.3 |
| 10 | Build **canvas first**, behind only the load-bearing minimum; the property vocabulary is *not* load-bearing | §11.1, §11.2 |
| 11 | Persons, events, and places ship together — three seeded rows, not three features | §11.1.1 |
| 12 | Entry is from the **Sources** list (dual action: open Source page *or* open Evidence graph) and later from the Source page. There is **no** Interpretation sidebar section — filing and the source-scoped graph are one product layer | §1.4, §11.1.2 |
| 13 | The foundation is its **own spike** (Spike 5) so the canvas spike opens with no schema work. Spike 5 ships **no Subject UI at all** — the data path is proven by Go tests, and the first visible subject is a bubble | §11.4 |
| 14 | The canvas is expected to be reused for a family tree, so **geometry goes in a neutral module** — but the position table stays concrete per layer rather than polymorphic | §13 |
| 15 | **One ref format, two prefixes.** Candidate subjects get their own three-letter prefix (`CPR-7KD45`) rather than a marker segment, so `subject_types` carries both `ref_prefix` and `candidate_ref_prefix`. Costs a second vocabulary column; saves a parallel mint/validate/search path everywhere else | [`catalog-refs.md`](../catalog-refs.md) §2 |
| 16 | **The graph is the only surface** for Subjects, Citations, and Observations — no co-equal list view, no view toggle, no subject list destination. Accessibility is met by the canvas's own accessibility representation, built from slice 2 rather than deferred to polish | §1.3, §7.4 |
| 17 | Product vocabulary for the graph's config mirrors Source types / Source fields: **Subject types** (`subject_types`) and **Subject fields** (`properties` + bindings). Avoid "claim" — Conclusion already owns that word | §1.4 |
| 18 | Sidebar: **Sources** is the primary work item; Source types / Source fields / Subject types / Subject fields are **nested config** under it (de-emphasized), not peer destinations | §1.4 |
| 19 | **Schema rename: Node → Subject.** Catalog tables are `subjects`, `subject_types`, `subject_positions`, `subject_type_fields`; product and engine share one word. Bridge rows are subjects too. Do not use "claim." | §1.4, interpretation model §4 |

Two items found along the way that were on nobody's list: **candidate ref support** did not exist in `core/ref` (§11.1, resolved in S5-01), and **NameValue does not exist in either language** (§4.4).

---

# 1. The premise

The Interpretation layer is Citation → Observation → subject. Subjects are candidate persons, events, places, and the bridge-like associations between them (`participation`, `location`, `relationship`), plus reified `source` subjects — which exist in the layer but are deliberately kept off this canvas (§4.6).

The default way to build a client for that is one destination per table: a Citations list and detail, an Observations list and detail, a Subjects list and detail. That is how the Source layer is built today (Sources, Source types, Source fields), and it works there because a researcher genuinely does think about one Source at a time.

That will **not** work here, because the unit of thought in interpretation is not one row. Reading a single census line produces a person, a household, several relationships, a residence, an occupation, and a birth-year inference — all at once, all from one locator, all held in mind together. A UI that makes the researcher navigate between three list/detail pairs to record that is charging a navigation tax on every fact.

## 1.1 The workflow the UI should match

Interpretation is *mostly* source-scoped: `subjects.source_id` is the subject's home Source, and a work session is normally "sit with one Source and map what it appears to say."

```text
open a Source
  → look at its Artifacts
      → mark what a portion of an Artifact says      (Citation)
      → say what that means                          (Observations)
          → about candidate things                   (Subjects)
      → repeat until the Source is exhausted
```

So the natural container is **one Source**, and the natural artifact of a work session is **a small graph** — typically tens of Subjects, not thousands.

## 1.2 The proposed shape

A single spatial workspace, scoped to a Source, with a small tool palette:

| Tool | Gesture | Result |
| --- | --- | --- |
| Add Person / Add Event / Add Place | click an empty grid cell | a bubble appears; a subject row is written |
| Connect | click subject A, then subject B | a line with a **bubble in the middle** — the bridge subject — plus the edges that bind it |
| Add Property | select any bubble (root or bridge) | citation composer opens: pick the locator, transcribe, fill the Observation, save; the property appears on the bubble |

Bridge bubbles are visually differentiated from root bubbles, so the association is a *thing* on the canvas the way it is a row in the catalog — but it reads as "the relationship between these two" rather than as a mysterious extra entity.

Because the graph is Source-scoped, it is reached from the **Sources** list (open Source *or* open Evidence graph) and later also from a control on the Source page (§11.1.2). There is no separate Interpretation sidebar section.

Two supporting destinations sit in the **same Sources family** in the sidebar:

- **Subject types** and **Subject fields** — the product names for `subject_types`, `properties`, and `subject_type_fields` bindings. Direct mirror of Source types / Source fields; reuse [`Features/CatalogVocabulary/`](../../macos/App/Features/CatalogVocabulary/). Not needed for the first slice (§11.2);
- a **"what other Sources say about this one"** section on the Source page, which is where source-to-source commentary lives instead of on the canvas (§4.6).

## 1.3 One surface, not two

Worth stating plainly, because an earlier draft of this note assumed otherwise. **The graph is the primary and only way to interact with Subjects, Citations, and Observations.** There is no co-equal list or table view of them, no view toggle, and no per-Source node list destination.

The researcher's path into the graph is: sidebar → **Sources** → open **Evidence graph** on a Source → that Source's graph. Everything past that click is spatial.

The reasoning is that a second, visible editing surface over the same rows costs more than it returns. It doubles the UI that every later slice has to update — Citations, Observations, value types, bridge macros would all have to land twice — and it invites the researcher to form a tabular mental model of a property graph, which is the model the canvas exists to replace.

This puts real weight on the canvas being good, and it removes the fallback: if the canvas turns out to be unpleasant, the answer is to fix the canvas, not to route around it. It also means accessibility cannot be satisfied by a sibling view and has to be solved inside the canvas — see §7.4, which is where the cost of this decision actually lands.

## 1.4 One product layer for Source + source-scoped graph

The data model still separates Source filing from Citations / Observations / Subjects (different rows, different epistemic weight, candidate refs vs concluded people). The **product** should not.

Filing a Source and mapping what it appears to say are one job: *work this Source*. A sidebar group called Interpretation that only contained "pick a Source → graph" (+ later vocabulary) felt hollow — not a second world, just the deep work inside a Source. Conclusion (cross-source belief) is the interesting nav boundary, not Source vs "Interpretation."

So the shipped Sources family becomes:

```text
Sources                         ← primary work: list → Source page *or* Evidence graph
  ├─ Source types               ← config: how documents are classified
  ├─ Source fields              ← config: metadata filed *about* the document
  ├─ Subject types              ← config: what kinds of subject a document can talk about
  └─ Subject fields             ← config: what can be asserted about a subject
```

**Sources** is the only primary work destination in this family. The four config items are **nested / de-emphasized** under it (smaller type, indent, disclosure — board decides), so they read as settings for how Sources work, not peer places you "go do research." Flat peer rows for all five would over-promote vocabulary admin.

**Subject** (not "claim") is deliberate: Conclusion already owns *claim* (`sameness_claims`, …). **Subject types / Subject fields** mirror Source types / Source fields so the two vocabularies stay parallel and distinct.

The catalog matches that language: `subject_types`, `subjects`, `subject_positions`, `subject_type_fields` (bindings). Bridge rows (relationship / participation / location) are subjects too. Engine and product share one word.

The spatial surface is the **Evidence graph** (§12) — easier once entry is already Source-scoped from the Sources list. Engine docs may still say Interpretation *layer*; the sidebar and the graph chrome do not have to.

---

# 2. What is genuinely new here

The codebase has **no prior art for either half of this**. Worth stating plainly before estimating:

| Piece | Current state |
| --- | --- |
| Spatial canvas, drag, connect, pan/zoom | Nothing. No `Canvas`, no custom gesture surface. `PVReorderableList` drag is the most complex gesture shipped. |
| Artifact rendering | Nothing. No PDFKit, no QuickLook, no AVKit. Files/Artifacts have thumbnails ([`core/derivatives`](../../core/derivatives/)) and a list row; nobody has ever *opened* an Artifact in the app. |
| Locator capture | Nothing. The selector vocabulary (`page`, `region`, `text_quote`, `time_range`) is fully specified and entirely unimplemented. |
| Structured NameValue | **Nothing, in either language.** `name_values` has no migration, no Go package, and no Swift editor — see §4.4. |
| Structured DateValue | **Done.** [`core/database/datevalues`](../../core/database/datevalues/) plus `DateValueDraft` / `DateValueEditorForm` in [`Features/Dates/`](../../macos/App/Features/Dates/), already wired into `SourcePageMetadataView`. The template for NameValue. |
| Typed value dispatch | Nothing. Source metadata is `value_text` + optional `date_value_id`; no `value_type` enum exists in the product — see §11.2. |
| Interpretation / Subject vocabulary UI | Reusable shell exists (`CatalogVocabulary`, `PVTable`, origin markers). Product names: Subject types / Subject fields (§1.4). |
| Candidate refs (`CPR-…`) | **Done** (S5-01), and smaller than expected. Candidates use their own prefix, so `ref.Mint` already produces them; S5-01 reduced to `ValidatePrefix` plus the reserved-prefix guard. |
| Interpretation schema / Go / FFI | Nothing. Migrations stop at `000020.sql`; no `subjects`, `citations`, `observations` packages. |

The vocabulary browser is a known quantity. The canvas and the artifact viewer are both greenfield, and the artifact viewer is probably the *larger* of the two.

---

# 3. The graph interaction model

## 3.1 A drawn line is a macro, not a row

The visual model is `[Wm Robins] ──father of──▶ [John Robins]`. The data model deliberately does not store that. It stores a birth Event, two `participation` bridge subjects, and six Observations, and expects the app to *project* "father of" from them (interpretation layer §1.4, §6).

So the connect tool writes a subgraph. The good news is that the "bubble in the middle of the line" design maps onto the schema almost exactly:

```text
  [Person A]  ──────  ( relationship )  ──────  [Person B]
       │                     │                       │
   free subject            free subject               free subject
                             │
              participant edges = Observations
                   └── each needs a Citation
```

**The bubble is free; the line segments cost a Citation.** A subject — including a bridge subject — needs nothing but a Source and a Subject type (§4.1). The *edges* radiating from it are Observations, and every Observation requires a Citation. That single asymmetry drives most of §4.

## 3.2 Connect is a macro with a required disambiguation step

Type-pair inference works, and the pairs match the seeded `subject_type_fields` bindings — but it is not silent, because every one of these macros needs a value the app cannot guess:

| From → To | Bridge subject | Edges written | What the app must ask |
| --- | --- | --- | --- |
| person → event | `participation` | `person`, `event` | `role` — subject, father, witness, … |
| person → person | `relationship` | `participant` ×2 | `relationship_type` — and *whether* it is a direct relationship at all (see below) |
| event → place | `location` | `event`, `place` | nothing; clean |
| person → place | *none seeded* | — | residence is `person → event(residence) → place`; offer a two-hop macro or refuse |
| event → event, place → place | *none seeded* | — | refuse |

The person → person case is genuinely ambiguous: two people in one record might be in a direct `relationship` ("cousins"), or might simply both participate in one Event ("both at this wedding"), which is a pair of `participation` Subjects instead. The connect tool has to offer that choice rather than pick.

This is fine, because the gesture already opens a form (§6). The form carries the role/type picker. **The rule is: no macro silently invents vocabulary.**

## 3.3 Reverse rendering is the harder direction

Draw → subgraph is a choice the app makes. Subgraph → draw cannot always be clean. Real, legal data with no tidy visual form:

- a `participation` with a `person` but no `event` yet — valid, because Subjects are sparse and Observations accumulate;
- two competing `birth_date` Observations on one subject from two Citations — expected, not an error;
- a **negative-polarity** edge — "this source denies he was there" — which is emphatically not the absence of a line;
- the same association asserted by three Citations with different transcriptions.

A canvas that can only draw the tidy cases will silently hide the interesting ones, which is the opposite of what an evidence-first tool is for. So every drawn element needs a visual state for **conflicted**, **negated**, and **incomplete**, and the canvas must never be the only way to see a subject's Observations.

---

# 4. Creation order, lifecycle, and deletion

This section answers the "do we have to create one thing before another, and are we trapping ourselves" question.

## 4.1 What the foreign keys actually require

```text
sources ──▶ artifacts ──▶ citations ──▶ observations
                                            ▲
subject_types ──▶ subjects ───────────────────────┤
properties ─────────────────────────────────┘
```

- **`subjects` is free.** It needs `source_id` (the Source whose graph is open) and `subject_type_id` (seeded vocabulary). `label` and `description` are nullable. Nothing else. Click the grid, get a row.
- **`citations` needs an Artifact.** `artifact_id` is `NOT NULL`.
- **`observations` needs a Citation, a subject, and a Property**, plus exactly one typed value.

So the only hard ordering is **Artifact → Citation → Observation**, and **Subject → Observation**. Subjects and Citations are independent of each other; either can come first.

**The gate worth noticing:** a Source with no Artifacts can hold Subjects but cannot hold a single Observation. The graph would be a board of uncited bubbles. That needs a deliberate empty state ("add a file to this Source before you can cite it"), not a mystery.

## 4.2 Draft status needs no schema flag

A subject with zero Observations is **already valid data** — the data model explicitly allows sparse Subjects ("creating a `person` subject does not require knowing a name"). So "uncommitted" is not a column; it is a query: *Subjects with no Observations where they are the subject.*

That gives a clean lifecycle rule:

| subject kind | Stands alone? | Uncited state |
| --- | --- | --- |
| Root — `person`, `event`, `place` | **Yes.** Created freely, one at a time. | Normal and expected. Persist immediately. |
| Bridge — `relationship`, `participation`, `location` | Legal, but semantically empty | **Never persist alone.** Commit the bridge subject, its edges, and the Citation in one transaction, or write nothing. |
| `source` | Yes, but not via this canvas | Created by the Source-page commentary flow, never placed on the graph (§4.6). |

The reason bridges differ: if we persisted a bridge subject and the researcher then cancelled the citation dialog, the canvas would be showing line segments that do not exist in the catalog. The connect gesture must be atomic. Root placement does not have that problem, because a lone person bubble is truthful — it says "this source seems to mention somebody," which is exactly what it means.

**`subjects.label` is the canvas's free working handle** — a nullable, non-evidentiary string. So a researcher can drop a bubble and type "head of household, line 14" with no citation at all. The UI has to keep that visually distinct from an asserted `name` Observation, or the distinction the whole layer exists to preserve gets blurred at the point of entry.

## 4.3 The real traps

1. **`subject_type_id` is immutable after insert.** Dropped a Person and meant an Event? There is no UPDATE — correcting a type means a new subject with a new `ref`. So either pick the type before placing (the tool palette already does this), or offer "change type" *only* while the subject has no Observations, implemented as delete-and-recreate. Refs are cheap to burn: [`core/ref`](../../core/ref/) mints random Crockford tokens from `crypto/rand`, not a sequence, so a discarded ref leaves no visible gap.
2. **Deletion is blocked by the database, and foreign keys are enforced.** [`core/database/catalog.go`](../../core/database/catalog.go) opens with `_foreign_keys=1`, and the spec'd `observations` → `subjects` and `observations` → `citations` references carry no `ON DELETE` clause — so the default `NO ACTION` makes deleting a referenced subject *fail*. Worse, a subject can be referenced as the **subject** (`subject_id`) or as the **object** of an edge (`value_subject_id`), so "what dies with this subject" is two queries, not one. The canvas needs an app-side cascade in a transaction plus a confirmation that counts the damage ("removes 7 Observations across 3 Citations"). Whether to write explicit `RESTRICT` or `CASCADE` is a decision to make *at migration time*, not after the UI exists.
3. **Every Property value type needs an editor.** `subject_type_fields` drives the Add-Property list, and the value editor depends on `properties.value_type`: `text`, `integer`, `real`, `boolean`, `date`, `name`, `node`. That is seven editors, and `node` means a subject picker scoped to the graph.

## 4.4 The hidden dependency: NameValue

`name` is the single most common Observation a genealogist will ever record, and **the structured NameValue model does not exist in any layer** — no migration, no `core/database/namevalues`, no Swift editor. [`structured-name-model.md`](../structured-name-model.md) specifies it (a required full `form` plus optional ordered parts with an open part-type vocabulary), but nothing is built.

DateValue is in much better shape, which makes it easy to assume names are too. They are not. Building NameValue end to end is its own chunk of work sitting directly on the critical path of "record a person's name," and it should be planned explicitly rather than discovered in the middle of a slice.

## 4.5 Should Citation and Observation be one-to-one?

Considered and **rejected** — but the reasoning is worth recording, because the simplicity argument is real and the decision is hard to reverse.

The proposal: enforce one Citation per Observation, rather than letting one Citation support many. Taken far enough, the two tables could merge into one row carrying artifact, locator, transcription, subject, Property, polarity, and value.

**What it would genuinely buy.** No "reuse this Citation or make a new one?" picker. Trivial deletion, with no shared-Citation refcounting and no orphaned Citations. No Citation-with-zero-Observations state to explain. One `ref` per assertion instead of a `CIT-…` and an `OBS-…`. If merged, one fewer table and one notes table instead of two. None of that is nothing.

**What it would cost: duplicated evidence, which is the one thing this layer exists to prevent.** Take a census line — *"William Robins, carpenter, aged 43, b. Somerset"* — yielding four Observations:

- **Four copies of the same `transcription`.** Correcting a misreading becomes four edits, and the four copies can drift. Divergent transcriptions of one line is corruption *in the evidence record*, which is the worst possible place for it.
- **Four copies of `locator_json`**, including four hand-drawn region polygons that will not be identical. Four Observations would then claim slightly different evidence for what was one act of reading.
- **Four settings of `transcription_uncertain` / `transcription_note`.** "Surname damaged" is a property of the reading, not of each fact derived from it.
- **`citation_notes` loses its subject.** "This line is in a different hand" is commentary about the evidence, not about the name assertion.
- **Audit noise.** Fixing a reading becomes N revisions across N rows instead of one audited act of correcting one transcription.

**Cases already written into the model docs that 1:1 cannot express without duplication:**

- [`interpretation-layer-data-model.md`](../interpretation-layer-data-model.md) §7 — one Citation `C1` supporting `name`, `occupation`, and `birth_date` on one person.
- §6 — one letter Citation `L1` supporting a `remark` about a **source** subject *and* a positive/negative `birth_date` pair about a **person** subject. That is one reading yielding four Observations across two different subject Subjects, where the positive/negative pair is only meaningful *as a pair*.
- A census household block: one region selector, six person Subjects.

**It also damages the canvas design specifically.** The connect tool writes two or three Observations per gesture (the bridge's `participant` edges plus its type). Under 1:1, "these two are cousins" becomes three Citations — three locators and three transcriptions for a single reading — and the macro stops being a macro. The pinned Citation in §6.1 would have nothing to pin, so the hot path gets *more* expensive, not less: a six-person household with five facts each is thirty Observations, meaning thirty locator selections and thirty transcriptions instead of roughly six.

**The asymmetry that settles it.** One-to-many can *behave* as one-to-one. One-to-one cannot later become one-to-many without retroactively guessing which duplicate Citations were "really" the same Citation, by comparing locators and transcriptions — lossy and ambiguous. For a product whose data is meant to outlive the application, take the reversible option.

**And the complexity being avoided is already small.** Invariant 7 already states that every Observation is supported by exactly one Citation, so the Observation side *is* one-to-one. The only multiplicity is on the read side, and it is a plain foreign key. The UI cost is not a schema problem; it is a single affordance — a visible "current Citation."

**Recommendation: one-to-one by default, sharing opt-in.** Every Add Property starts a fresh Citation unless one is pinned. A researcher who never pins never encounters Citations as shared objects, and gets exactly the simple model the question was after. A researcher working a census pins one and gets the speedup. No schema change, no lost capability.

## 4.6 Cross-source subjects: what they are actually for

Worth pinning down, because the obvious guess about why they exist is wrong, and that changes what the canvas should offer.

### It is not a feature — it is a constraint we did not add

There is no cross-source machinery in the schema. `subjects.source_id` is an ordinary column, and nothing anywhere confines an Observation's Citation to the subject's home Source. The interpretation model says so directly: `source_id` "exists so the application can efficiently surface Subjects that belong with a given Source during common same-source workflows. It does not restrict which Citations or Observations may reference the subject." Invariant 3 repeats it.

So "removing it" does not mean deleting code. It means **adding** enforcement — and SQLite cannot express it as a `CHECK`, because the rule spans `observations` → `citations` → `artifacts` → `sources` and back to `subjects.source_id`. It would be a trigger or an application invariant, and it would run against the grain of the interpretation model's own §1.2 ("the database should enforce generic graph integrity… it should not attempt to encode the entire genealogy ontology into rigid table structure") and of the existing posture that even target Subject type constraints are application invariants rather than SQL allow-lists.

### Consolidating a person across Sources is *not* the reason

This is the guess worth killing. If two Sources describe the same man, the architected answer is **one subject per Source plus a Sameness Claim** — not one shared subject. Both worked examples in [`conclusion-layer-data-model.md`](../conclusion-layer-data-model.md) do exactly that:

```text
§12.1   subject N1 (home = photograph), subject N2 (home = testimony)
        sameness_claim: N1 same_as N2 (accepted) → members(E1) = {N1, N2}

§12.2   subject NC (home = certificate), subject NL (home = letter)
        sameness_claim: NC same_as NL (accepted)
        reconciliation_claim on E, birth_date → 2 JAN 1800
```

And the Conclusion layer already solves the follow-on problem too: separation 4 in its §3 states that a Reconciliation Claim's exhibit "Observations may be about **other** Subjects; they are not retargeted." So evidence recorded against one Source's subject can support a conclusion about the shared entity without anything being rewritten or shared.

Using a cross-source Observation to consolidate people would actively **damage** the model: it attaches Source B's assertion to a subject whose identity came from Source A, destroying the per-Source "what does this one source appear to say" guarantee that the entire layer exists to protect.

### The one thing that genuinely requires it

**Source-to-source commentary.** A `source` subject reifies exactly one Source row, and the interpretation model's §4.2 says the application "should maintain at most one `source` subject per Source." So when a book mentions a certificate:

```text
Book Citation B1:   SB1 -- mentions --> SC1
Letter Citation L1:  SC1 -- remark --> "date of birth on certificate mistyped"
```

`SC1` is homed to the certificate while the Citation sits under the book or the letter. That is cross-source by construction and there is no same-source way to say it. The only workaround would be minting a second `source` subject for the certificate under each citing Source — which breaks the at-most-one rule and would mean reconciling duplicate reifications of a row whose `id` is sitting right there.

Note the shape: in `mentions` the foreign subject is in the **object** position (`value_subject_id`); in `remark` it is the **subject**. Both occur.

### Why only one `source` subject per Source

The rule looks arbitrary next to person Subjects, where duplicates are not only allowed but expected. The difference is that **duplicates are only safe where there is a mechanism to resolve them, and for `source` subjects there deliberately is not.**

1. **Source identity is already machine-known.** A person subject is a *candidate* — that is what the `CPR` prefix in `CPR-7KD45` means — whose real identity is uncertain and source-local. A `source` subject reifies a row in your own catalog. There is nothing uncertain to resolve: `sources.id` is the answer, and the Source already has its own canonical `SRC-…` ref.
2. **So `source` subjects get no canonical rows.** [`conclusion-layer-data-model.md`](../conclusion-layer-data-model.md) §6 states it directly: "Reification `source` subjects are not typically given canonical rows." That is the crux. Duplicate person Subjects are fine because `canonical_entities` + `sameness_claims` + `reconciliation_claims` exist to unify them. That entire apparatus is intentionally out of play here.
3. **Which leaves duplicates unresolvable.** You *could* write `sameness_claims` between two `source` subjects — both endpoints share a Subject type, which the composite foreign key requires, and `source` is a Subject type. But membership closure runs from an `identity_anchor_id` on a canonical entity (§2.2), so with no canonical row the cluster has no handle and the unification is inert. You would be asserting, by hand, something the database already knows from `source_id` — and getting nothing back for it.
4. **And duplicates would buy nothing.** Subjects carry no evidentiary content; a `source` subject has only `label` and `description`. Each citing Source's independent view of the certificate already lives in *its own* Observations under *its own* Citations. Sharing the subject does not merge those testimonies, so splitting it does not protect them.

So the single-subject rule is not a restriction that costs expressiveness. It is **uniqueness by construction standing in for the reconciliation layer that `source` subjects deliberately do not have.**

Note it cannot be a SQL constraint. A partial unique index would need `WHERE subject_type_id = <the source type>`, but `subject_type_id` is a runtime UUID looked up from data rather than an enum, so there is no static predicate to index on. Hence the doc's "should" — it is a Go-side application invariant, like the value-population and target-type rules.

### The coherent alternative, for the record

There is one design that would make `source` subjects uniform with every other subject *and* eliminate the only hard requirement for cross-source Observations: **split home from reified.** Let `source_id` mean "the Source that encountered this subject," exactly as it does for persons, and add a second reference — `reifies_source_id` — for the Source being talked about. Then each citing Source mints its own `source` subject homed to itself, every Observation becomes same-Source, and duplicates auto-cluster on `reifies_source_id` with no researcher effort.

It is not a silly idea. The reasons not to take it: it puts a type-specific column on a table explicitly designed to hold almost no domain data; the "reifies" relationship cannot be expressed as an Observation instead, because there is no `source` value type among the seven primitives; and it requires edits to interpretation invariant 4, §4.2, §5.1, and conclusion §12.2. That trades a small application invariant for a permanent schema special case, which is the worse deal — but it should be a decision, not an oversight.

### A real gap worth noting

`subjects.source_id` is `NOT NULL REFERENCES sources(id)`, so a `source` subject must point at a Source that exists in the catalog. That means "this book mentions a marriage certificate I have never seen" forces a placeholder `sources` row for a document you do not hold.

That is arguably fine and even useful — an unheld Source is a legitimate research to-do, and the Source layer already permits Sources with no Artifacts. But it has a consequence: if two Sources each mention what might be the same unheld document, resolving that means **merging Sources**, which is a Source-layer operation that does not exist. Not a blocker for the graph slices, but it belongs on the list before `mentions` ships.

### Keep Sources off the canvas — but as a UI decision, not a schema change

The thematic objection is correct: the Interpretation graph is about historical persons, events, and places. A bubble representing "the marriage certificate in my filing cabinet" is a category error on that surface — it is a research object, not a historical entity. The `source` Subject type is already slightly embarrassed about itself in the model, which warns that it "must not use `SRC`; a distinct prefix such as `SRN` keeps Source catalog refs distinguishable from source-subjects in speech."

But two decisions are easy to conflate here, and only one of them is cheap:

| Decision | Cost |
| --- | --- |
| **(a)** Don't render `source`-type Subjects on the canvas | **Free.** The graph query is ours to write: filter to `node_type != source AND source_id = <this Source>`. |
| **(b)** Stop modelling source commentary as Observations, and have Citations point at `sources` directly | **Expensive**, and it breaks documented behavior — see below. |

**Take (a). Do not take (b).** They are independent: the data model can keep `source` subjects and `remark` / `mentions` Observations exactly as designed while the canvas simply never draws them. Source-to-source commentary then lives where a researcher would actually look for it — a "what other Sources say about this one" section on the Source page — and the canvas stays about people, events, and places.

### Why (b) costs more than it looks

**It breaks Conclusion exhibit.** Both `sameness_claim_evidence` and `reconciliation_claim_evidence` are `observation_id … REFERENCES observations(id)`, and the Conclusion doc is explicit: "Exhibit pins are **Observations only** for now — not Citations, Sources, or other Claims." Conclusion §12.2's worked example pins "the birth_date Observations (and related letter Observations as needed)" — and the related letter Observation *is* `SC -- remark --> "date of birth on certificate mistyped"`. Move remarks out of `observations` and that exhibit path stops existing. "The letter says this certificate is a forgery" is exactly the kind of thing a researcher needs to pin to a Reconciliation Claim.

**It closes an open vocabulary.** `remark` and `mentions` are ordinary Properties today, so a researcher can add `supersedes`, `is_transcription_of`, or `derived_from` with no migration — precisely the §1.2 extensibility promise. A bespoke `source_remarks` table freezes that vocabulary at whatever columns it ships with.

**It duplicates the evidence machinery.** A second Citation-backed assertion table means a second ref prefix, its own audit wiring, its own notes table, its own FFI surface, its own search projection, and a second thing that has to stay consistent with Citations forever.

**And the polymorphic alternative is worse.** Adding `subject_source_id` alongside `subject_id` on `observations` breaks invariant 9 ("exactly one subject") and makes the subject a sparse polymorphic pair that every reader must handle — importing the value-column problem into the subject position, on the layer's most important table.

### On making a canonical entity for a Source

Also considered: bridge the gap by giving each Source a `canonical_entities` row, so `source` subjects could resolve through the ordinary Conclusion machinery.

This is the wrong direction. A Source **already has canonical identity** — a row in `sources` and a public `SRC-…` ref. Adding a canonical entity would create a *second* canonical handle for one document, with two refs, two merge stories, and a synchronization rule between them. The Conclusion doc's "Reification `source` subjects are not typically given canonical rows" is guarding exactly that. The catalog is the canonical layer for Sources; Conclusion is the canonical layer for historical things that only exist as inference.

### What this means for the canvas

| Decision | Call |
| --- | --- |
| Enforce same-Source Observations in the schema | **No.** It costs a trigger to add, breaks `mentions` / `remark`, and contradicts the model's stated posture. The permissive model is a superset that costs nothing to keep. |
| Show foreign Subjects on the canvas in early slices | **No.** Scope the graph query to `subjects.source_id = ?`. Free, and it removes the two-payload cache hazard in §5.3. |
| Offer "attach this to a person from another Source" | **Never.** The canvas should not make the epistemically wrong thing easy. Create a candidate here; claim sameness later. |
| Render `source`-type Subjects on the canvas | **No, ever.** Source-to-source commentary belongs on the Source page, not the graph. That removes the last reason the canvas would ever need cross-source display, and it is a query filter, not a schema change. |

The honest cost of this posture: a researcher working twenty census years on one family creates twenty person Subjects for the same man and reconciles them in the Conclusion layer. That friction is *designed* — it is what keeps each Source's testimony independent — but it is real, and it is the strongest argument for making Sameness Claims a pleasant workflow when that layer arrives. The canvas should not route around it.

**Net effect on the graph slices:** with `source` subjects off the canvas, every subject the canvas draws is homed to the open Source, so the graph query is `source_id = ?` with no exceptions, the two-payload cache hazard in §5.3 never arises, and cross-source Observations remain legal in the model without ever being the canvas's problem.

---

# 5. Layout state

The Interpretation schema has no `x`/`y` and should not get them — position is not evidence. But it is not throwaway either: twenty minutes arranging a census household is real intellectual work that must survive relaunch and ideally project sharing ([`ideas/share-packages.md`](share-packages.md)).

A separate catalog table, explicitly excluded from audit, is the leaning: it travels with the project, shares, and syncs, at the cost of one migration and some UI state in the catalog.

Two simplifications worth taking:

- **Snap-to-grid means integers.** If the grid is the interaction model, make it the *storage* model: `grid_x` / `grid_y` as `INTEGER` cell coordinates rather than floats. No float drift, trivial equality, cheap conflict resolution, and snapping stops being a separate feature.
- **A tray beats an auto-layout engine.** Something has to handle Subjects with no stored position — future imports, anything created outside the canvas, and Subjects whose position row was never written. Instead of building auto-layout for v1, put unplaced subjects in a **tray along the edge of the canvas** and let the researcher drag them onto the grid. That removes an entire algorithmic dependency and is arguably better behavior: the app never guesses at an arrangement that means something.

## 5.1 One table is enough

```sql
CREATE TABLE subject_positions (
    subject_id BLOB PRIMARY KEY REFERENCES subjects(id) ON DELETE CASCADE,
    grid_x     INTEGER NOT NULL,
    grid_y     INTEGER NOT NULL
) STRICT;
```

Design notes, each of which is a decision worth making deliberately:

- **The key is `subject_id` alone.** One position per subject; the open graph is still Source-scoped (§4.6), and layout for that graph is subjects with `subjects.source_id = ?` joined to this table. There is no plan for the same subject on multiple Source graphs, so denormalizing `source_id` onto the layout row is unnecessary. (If that ever changes, this becomes a migration — acceptable because layout is unaudited UI state.)
- **No `graphs` table**, and the table name says so. One graph per Source is the premise of the whole design, so the Source *is* the graph identity. If named views or multiple layouts per Source are ever wanted, that becomes a `graph_id` and a migration — acceptable precisely because this is unaudited UI state and therefore cheap to migrate. A future family tree gets its own concrete table rather than a shared polymorphic one; the reasoning is in §13.3.
- **Absence of a row means unplaced**, which is exactly what feeds the tray. No nullable coordinates and no sentinel values. Coordinates stay off `subjects` so layout remains unaudited and the tray stays “no row” rather than null columns on an audited entity.
- **`subject_id` cascades**, which is a deliberate *contrast* with `observations` (§4.3). Deleting a subject should silently drop its position. Evidence must never be swept away that quietly, but layout should be. Source-level layout wipe is indirect: `subjects.source_id` is `NO ACTION`, so a Source with Subjects cannot be deleted until those subjects (and thus their positions) are gone.
- **No `UNIQUE (grid_x, grid_y)`.** Tempting, but it would make overlaps a hard error — and a drag that swaps two bubbles transiently collides, which a database constraint cannot accommodate without temporary values. Let bubbles overlap and let the UI nudge.
- **Coordinates are signed.** The canvas is an unbounded plane around an origin, so do not add a `CHECK (grid_x >= 0)`.
- **Not audited.** Arranging bubbles is not a research assertion, and audit rows for every drag would drown the revision history that matters.

## 5.2 Positions travel; the camera does not

Worth splitting explicitly, because they look like the same kind of state and are not:

| State | Where it belongs | Why |
| --- | --- | --- |
| Bubble positions | The catalog, in the table above | Intellectual work. Must survive relaunch, travel in a share package, and sync. |
| Pan offset and zoom | App-local (`UserDefaults`), or nowhere | Per-machine view state. It is not research, should not sync, and should not appear in a shared project. |

Selection is likewise transient. Neither camera nor selection belongs in navigation history (§10.12).

## 5.3 Cache shape

This fits the existing session model cleanly. One key — something like `.sourceGraph(project:sourceId:)` alongside the cases in `CatalogQueryKey` — owns **both** the subjects and their positions for one Source, because nothing else owns positions and the subject set is genuinely Source-scoped. That satisfies the "one cache owns each list" rule in [`macos-client-patterns.md`](../macos-client-patterns.md) §1 without duplicating another key's data.

Drag is then a **patch, not an invalidation**: the move response names exactly the one row that changed, which is the case `setQueryValue` exists for. Creating or deleting a bubble invalidates the key.

One caveat for later: *if* foreign Subjects are ever displayed, the same subject would appear in two graph payloads and renaming its label would have to invalidate both — the duplicated-versus-derived hazard from the same doc section. Keeping the canvas Source-scoped (§4.6) means this never arises in the planned slices, and `source` subjects never raise it at all because they are never drawn.

Dragging must not write per frame. Positions batch and debounce, flushing on gesture end; the catalog session serializes FFI ([`use-catalog-session`](../../.cursor/skills/use-catalog-session/SKILL.md)), so a chatty canvas would queue behind badge refreshes and list loads.

---

# 6. The citation composer

Every property edit opens this, so it is on the hot path of the whole layer.

```text
Artifact viewer            Locator                  Interpretation
─────────────────          ─────────────────        ─────────────────
PDF page / image /         page + page_label        subject
audio / video              region polygon           Property
                           text_quote               polarity
                           time_range               typed value
transcription                                       (text/int/real/bool/
transcription_uncertain                              date/name/node)
description
```

macOS building blocks, per media type:

- **PDF** — `PDFKit.PDFView` wrapped in `NSViewRepresentable`. Gives page navigation (feeding `artifact_page`), and `PDFSelection` gives selected text plus surrounding context, which is exactly `text_quote.exact` / `prefix` / `suffix`. `PDFPage` coordinate conversion supports normalized region points.
- **Image** — `NSImage` plus a custom overlay for polygon drawing.
- **Audio/video** — AVKit `AVPlayer` with a time observer for `time_range`.
- **QuickLook** (`QLPreviewView`) is tempting because it handles every format nearly free — but it exposes no selection or coordinate API, so it cannot produce locators. Useful as a read-only preview, not as the composer.

Region polygons need a custom drag overlay producing normalized points, and the invariants are strict (≥3 distinct points, non-self-intersecting, non-zero area, in-bounds). **Validate locator JSON in Go**, so a future Windows client inherits it and `locator_json` can never be written malformed.

## 6.1 A pinned Citation collapses the map-first / citation-first argument

One Citation may support many Observations — the data model says so explicitly. So the composer should be able to **stay pinned** to the current Citation while the researcher keeps working: transcribe the census line once, then attach eight Observations, place three bubbles, and draw two connections against that same locator.

That matters because it dissolves a workflow question this note originally treated as either/or. Map-first ("place things, then cite them") and citation-first ("read one line, then record what it contains") become the same UI operated in a different order. Nothing has to be bet on which one researchers actually prefer.

It also argues strongly for **side-by-side, not modal**. A sheet that covers the graph you are annotating is the wrong default; the composer wants to be an inspector or a companion window that persists across successive edits.

---

# 7. macOS feasibility, and the deployment-target question

## 7.1 Rendering

The standard pattern is a hybrid: SwiftUI `Canvas` for the grid and edges (immediate-mode, GPU-backed, cheap for hundreds of lines), with real SwiftUI views for the subjects in a `ZStack`, positioned with `.position(_:)`. `Canvas` cannot host subviews, take per-shape gestures, or expose accessibility children, so Subjects must be real views to get hit-testing, text fields, focus, and VoiceOver. That comfortably handles the tens-to-low-hundreds of Subjects one Source produces — and would not handle a whole-project graph, which is a further reason to keep this Source-scoped.

## 7.2 Pan and zoom needs AppKit regardless of target

Checked, because it drives the target question: **SwiftUI's `ScrollView` has no zoom on any current version.** Apple's own documentation states it "does not provide zooming functionality," and macOS 15's `ScrollPosition` adds programmatic scrolling to a view id, offset, or edge — *not* magnification. So the answer is `NSScrollView` via `NSViewRepresentable`, using `allowsMagnification`, `magnification`, `minMagnification` / `maxMagnification`, `setMagnification(_:centeredAt:)`, and `magnify(toFit:)`. That is also the more Mac-native result (elastic bounds, native scroll momentum) and is squarely the "escape hatch, not the default" case in [`macos-client-patterns.md`](../macos-client-patterns.md) §4.

**Concrete gotcha to plan for:** a magnified `NSScrollView` does not correctly translate points into the coordinate space of its SwiftUI children. So hit-testing and drag-to-connect cannot naively trust SwiftUI gesture locations — coordinates need resolving in the scroll view's content space. Build **one** coordinate-conversion seam, keep it a pure function over value types, and unit-test it. Getting this wrong produces bugs that feel like "the canvas is haunted."

## 7.3 Should we drop macOS 14?

Current state: local Xcode 26.6 on macOS 26.6; deployment target 14.0 in both configurations; CI runs the `macos-15` runner. Raising to 15 is cheap and CI-compatible today.

But **do not justify it with the canvas.** Pan/zoom needs AppKit either way (§7.2), which was the main hoped-for win. What raising to 15 actually buys is modest and general: `ScrollPosition` for "scroll to this subject," `onScrollGeometryChange` for viewport tracking or a future minimap, the `@Entry` macro, and possibly `LocalizedStringResource` inits that would retire the `String(localized:)` boilerplate documented in [`macos-client-patterns.md`](../macos-client-patterns.md) §6.

macOS 26 adds rich-text `TextEditor` bound to `AttributedString`. That is deliberately **not** wanted for `transcription`, which must stay faithful plain text — formatting in evidence text is a liability, not a feature. It might be interesting for notes later, but `citation_notes.body` / `observation_notes.body` are `TEXT`, so it would need a serialization decision first.

**Recommendation: decide this on support-matrix grounds, not capability grounds.** It is a defensible call — macOS 14 is two releases behind — but it should be its own decision, not a rider on this spike. If we do raise it, bump the CI runner and both Xcode configurations together.

## 7.4 Accessibility and testing

A free-form spatial canvas is genuinely hostile to VoiceOver and keyboard-only use. It needs full keyboard parity for every canvas action, plus a structured representation of the same graph for assistive technology.

**That representation lives inside the canvas, not beside it.** The graph is the only surface for Subjects, Citations, and Observations (§1.3), so there is no second destination to fall back to — which means the accessible structure is not an alternate view the researcher can choose, it is the canvas's own accessibility tree. SwiftUI supports this directly: `accessibilityRepresentation(_:)` substitutes an entirely different view hierarchy for assistive technology while the canvas is what gets drawn, `accessibilityChildren(_:)` supplies children for a view that has none of its own, and `accessibilityRotor(_:)` gives VoiceOver a way to move through Subjects without spatial navigation. Because Subjects are real SwiftUI views in a `ZStack` rather than shapes inside `Canvas` (§7.1), they can carry their own accessibility identity; the representation is there to impose *order and relationships* on what is otherwise an unordered plane.

**This cannot be deferred to polish.** With no list destination, the representation is the only path for a VoiceOver or keyboard-only researcher — so if it is missing, the layer is not merely awkward for them, it is unusable. It ships with the first bubbles in slice 2 (§11.4), not in the honesty-and-polish slice. Retrofitting an accessibility tree onto a finished spatial editor is the standard way software ends up permanently inaccessible: the structure has to exist while there is still almost nothing to structure.

**It is also what makes the canvas testable at all.** XCUITest finds controls by accessibility rather than by pixels, so a canvas without a representation cannot be driven by UI tests — but a canvas *with* one can. That inverts the earlier conclusion here: the canvas does not have to be resigned to hand verification. The representation is the test handle, which is a second, self-interested reason to build it early rather than late. The rest of the coverage split is unchanged: Go tests for schema, macros, and locator validation; Swift unit tests for pure geometry and coordinate conversion.

---

# 8. Pros

- **It matches the unit of thought.** One Source, one graph, one session. No bouncing between three list/detail pairs to record one line of a census.
- **It makes bridge subjects tolerable.** `participation`, `location`, and `relationship` are the price of a generic property graph and are miserable to fill in through forms. As a connect gesture with a labelled bubble, they become almost pleasant — and the bubble keeps them honest, visible as the entity they really are.
- **It is naturally bounded.** Source-scoped graphs stay small. This is why it can work where "show me my whole family tree as a graph" does not.
- **Completeness becomes visible.** A sparse subject or a dangling bridge is obvious on a canvas and invisible in a list. The graph doubles as a QA view for the interpretation of that Source.
- **Citation stays on the hot path.** If every property edit routes through the composer, "cited by construction" becomes the path of least resistance rather than a discipline to maintain.
- **It is the product differentiator.** Evidence-first genealogy tools are mostly form-driven. This is the part of Provenencia that would not be a reimplementation of something that exists.

---

# 9. Cons and risks

- **Highest-risk UI in the app, on zero prior art**, landing alongside the layer's schema, Go, and FFI. Mitigated by slicing (§11), not eliminated.
- **The canvas cannot express everything.** Polarity, competing Observations, transcription uncertainty, and cross-source Observations all resist spatial representation. Some list/detail surface is still required, so this is "graph *plus* fewer pages," not "graph instead of pages."
- **Round-trip lossiness** (§3.3), in a tool whose entire purpose is not misrepresenting evidence. This is the one risk with no clean mitigation yet.
- **Position becomes state we did not want** — a UI-state table in the catalog, with sharing and migration consequences. Accepted deliberately, shape settled in §5.1.
- **Accessibility and keyboard parity are non-optional work**, not a polish pass.
- **It strains the current session cache.** `WorkspaceSession` patch/invalidate is tuned for lists and detail payloads, while a graph is one large payload mutated constantly from inside the view. §5.3 proposes the shape (one key owning Subjects plus positions, patch on drag, invalidate on create/delete), but it is unproven until the canvas slice runs it.
- **Deletion is a genuine UX problem, not a button** (§4.3).
- **Scope creep is the default outcome.** "Snap to grid, then auto-layout, then edge routing, then minimap, then multi-select, then alignment guides" is an infinite backlog that produces no genealogical capability. The tray (§5) is one deliberate refusal; there will need to be more.

---

# 10. Hiccups checklist

1. **Edge macros** — the matrix in §3.2, including the refusals and the person→person disambiguation.
2. **Reverse rendering rules** — visual states for negated, conflicted, and incomplete; collapse/expand of bridge bubbles.
3. **Deletion semantics** — `ON DELETE` choice at migration time; subject *and* object references; the confirmation that counts the damage.
4. **Type correction** — delete-and-recreate, gated on having no Observations.
5. **Layout table** — the `subject_id` primary key shape in §5.1, unaudited, plus the unplaced tray and the positions-travel/camera-does-not split (§5.2).
6. **Artifact gate** — the empty state when a Source has no Artifact.
7. **NameValue end to end** (§4.4) — schema, Go, and editor, on the critical path.
8. **Locator validation in Go**, with the full invariant set, before any UI writes `locator_json`.
9. **Coordinate conversion** under magnification (§7.2) — one seam, unit-tested.
10. **Cross-source subjects** — scoped out of the canvas entirely: `source` subjects never render, and foreign person/event/place Subjects are deferred. See §4.6 for why they exist and why the canvas must never offer person reuse across Sources.
11. **Cache strategy** for a large, constantly mutated graph payload — shape proposed in §5.3 (one `sourceGraph` key, patch on drag, invalidate on create/delete).
12. **Navigation history** — the graph is a place ([`add-workspace-location`](../../.cursor/skills/add-workspace-location/SKILL.md)); camera and selection are not (§5.2). `WorkspaceLocation` needs **no** new field — a new section case plus the existing `sourceId` is already a distinct identity (§11.1.2).
13. **Undo** — one connect gesture writes several rows and people will hit ⌘Z. The audit model already says undo is a forward revision, not a deletion ([`audit-revision-history.md`](../audit-revision-history.md) §9), so undo is a Go concern, not an `UndoManager` concern.
14. **Density and filtering** — a census page yields dozens of Subjects and hundreds of Observations; needs layers/filters before it is usable on a real source.
15. **Design system** — canvas chrome, bubbles, edges, and selection from existing tokens, not a parallel visual language.
16. **Localization** — fewer strings than a form-heavy screen, but bubbles, macro menus, and every warning state still go through `L10n`.

---

# 11. Vertical slices

Not backend-then-frontend. Each slice cuts through migration → Go → FFI → UI. **The canvas goes first**, preceded by the smallest foundation that will physically support it, because the canvas is the only thing in this document that could invalidate the plan. Everything else — vocabulary, value types, NameValue, citations — is durable work on known patterns that will be needed whatever the canvas turns out to be, so none of it should stand in front of the risk.

## 11.1 The load-bearing minimum

The test is narrow: **what does the first `subjects` INSERT actually require?** The table has exactly two foreign keys, `sources` and `subject_types`, and `sources` already exists. That is the whole foundation.

| Item | Why it is load-bearing | Size |
| --- | --- | --- |
| **Candidate ref support in `core/ref`** | ~~`subjects.ref` needs a form `Mint` could not produce.~~ **Resolved in S5-01**, and mostly by deleting the problem: giving candidates their own `candidate_ref_prefix` means `ref.Mint` already works, leaving only `ValidatePrefix` and the reserved-prefix guard to add. Rationale in [`catalog-refs.md`](../catalog-refs.md) §2. | Small |
| **`subject_types` table + seeds** | `subject_types` is the other FK. Needs the table, plus rows via the existing idempotent `Install` registry pattern (`sourcecredibilitygrades` is the closest template, `add-seeded-vocabulary` the skill), plus `ref_prefix` values. **Table and seed only — not the browser UI.** Seed all seven types from [`seeded-vocabulary.md`](../seeded-vocabulary.md) §3.1; only three are placeable, and they cost the same as one (§11.1.1). | Small |
| **`subjects` table + `core/database/subjects`** | Create, list, rename, delete — **with audit wiring.** Every domain write in this product goes through `audit.Record(tx, …)` (see `sources/notes.go`); that is not optional, and it is the bulk of the work here. | Medium |
| **Layout table** | Integer grid cells, unaudited (§5). Persistence across relaunch is part of what the canvas validates, so it cannot be held in memory — and it ships in slice 1 so the canvas slice opens with no schema work in front of it. | Small |
| **FFI handlers** | `add-ffi-handler` skill. | Small |
| **Graph workspace place** | A **deep place under Sources**, not a new sidebar section. Needs a `WorkspaceLocation` discriminator so Source page and graph are distinct at the same `sourceId` (§11.1.2). Stub destination in slice 1; canvas in slice 2. `CatalogQueryKey` for the graph payload. `add-workspace-place` / `add-workspace-location`. | Medium |
| **Subject types / Subject fields (nav)** | Two more sidebar destinations beside Source types / Source fields (§1.4). **Nav chrome + stubs** can ship with the foundation; the real vocabulary editors are not load-bearing for bubbles (§11.2). | Small (nav) / Medium (editors later) |

### 11.1.1 Three Subject types cost the same as one

Persons, Events, and Places are **three seeded rows**, not three features. `subjects.subject_type_id` is a generic foreign key, so the migration, the Go package, the FFI handlers, and the layout table are all type-agnostic — a `Create` that takes a Subject type id serves all three.

What does scale with the count is small and entirely presentational: three palette buttons instead of one, per-type bubble styling so a Person does not look like a Place (icons already exist under `EvidenceIcons`), and the `L10n` strings for each. That is view code, not architecture.

So the earlier split of "person first, then events and places" was a false economy — the second slice would have been almost empty. They collapse into one slice (§11.4), and the canvas gets to look like the real product from the first demo, which also makes it far easier to judge whether the idea works.

### 11.1.2 Entry points: dual action on Sources, not a second section

The graph is Source-scoped, so the researcher has to pick a Source before there is anything to draw. The primary path is the **existing Sources list**, with two actions per row: open the Source page (filing) and open the graph (structured reading). A control on the Source page is a convenient second path, deferred until the graph exists to jump to (slice 2).

**There is no Interpretation `WorkspaceSection`.** Filing and mapping share the Sources family in the sidebar (§1.4). That avoids an empty layer group whose only job was a second Source picker.

**`WorkspaceLocation` does need a new field.** Identity today is `section` + `sourceId` + `fieldId` + `typeId`. Source page and graph both want `section: .sources` and the same `sourceId` — without a discriminator they are the same place, which breaks history, Back/Forward, and restore:

```text
WorkspaceLocation(section: .sources, sourceId: X, sourceSurface: .page)   → Source detail
WorkspaceLocation(section: .sources, sourceId: X, sourceSurface: .graph)  → that Source's graph
```

Name and encoding are an implementation choice (`sourceSurface`, `sourceView`, …); the requirement is that page and graph are unequal for `==` and for persistence. Absent the new field, treat legacy entries as `.page` so old history keeps opening the Source page.

Fallback when a Source is deleted while its graph is in history: `fallbackToSectionRoot()` → Sources list. No separate Interpretation root.

**Defer graph-specific columns on the Sources list.** "12 subjects, 3 uncited" needs derived counts and would stale the list when the canvas writes — out of scope for the first slices.

**No-Artifact Sources** stay in the list; the Evidence graph action is disabled and a shortcut opens the Source page to add an Artifact (Citations require an Artifact).

## 11.2 What is not load-bearing

The structural fact that shrinks the prework dramatically: **`subjects` has no foreign key to `properties`.** Only `observations` does. So the entire property vocabulary — the piece most naturally listed first — is not needed to put bubbles on a grid.

Deferred, blocking nothing:

- `properties` and `subject_type_fields` — Observation concerns, not subject concerns.
- The vocabulary browser UI — **Subject types / Subject fields** in the product (§1.4). Subject types get seeded directly from the registry in [`seeded-vocabulary.md`](../seeded-vocabulary.md) §3.1; browsing and user extension can wait. Seed all seven rather than only the three placeable ones: they are rows, not features, and a partial seed just means editing the registry again. **Sidebar stubs** for those destinations can still ship early so the Sources family looks complete.
- `ref_prefix` user-facing validation (global uniqueness across origins, reserved `SRC` / `ART` / `CIT` / `OBS` / `C`) — only needed when *users* define Subject types, which is the browser UI.
- NameValue, the typed value dispatch, and the seven value editors — all Observation concerns. Scope notes for when they land: Interpretation needs only `name_values` and `name_value_parts`, because interpretation model §5.1 puts `name_format` in the Conclusion layer — two tables, not the four in [`structured-name-model.md`](../structured-name-model.md) §4. `date` is already done and is the template. The seven-way sparse-column dispatch is the genuinely new part, since Source metadata is only `value_text` plus an optional `date_value_id` and no `value_type` enum exists anywhere in the product yet.
- Citations, the artifact viewer, locator validation.
- Open value vocabulary (`event_type`, `role`, `relationship_type`) — and note this one is not even a build task yet: [`seeded-vocabulary.md`](../seeded-vocabulary.md) §1.5 is explicit that these are open free-text values rather than vocabulary-definition tables, with no `origin`, so the seeded picker defaults have no home in the schema. Where they live is a design question to settle in planning.

## 11.3 What the canvas slice proves, and what it does not

**Proves:** the `NSScrollView` bridge, coordinate conversion under magnification (§7.2), bubble hit-testing and drag, snap-to-grid persistence, session-cache behavior under canvas mutation, whether the canvas accessibility representation is workable at all (§7.4) — and above all whether the thing feels right to use.

**Does not prove:** the citation composer, the property flow, or the connect macros. Those are real risks, but they are *different* risks with their own slices. The one retired here is the only one with no prior art and no fallback: whether a spatial editing surface is buildable and pleasant on this stack.

## 11.4 Slice order

1. **Foundation** (§11.1) — candidate refs, `subject_types` and its seed, audited `subjects` CRUD, the layout table, FFI, `WorkspaceLocation` discriminator for page vs graph, Sources-list dual action, Evidence graph stub, and nested Subject types / Subject fields **nav stubs** (§1.4). **No Subject UI**: Evidence graph opens a stub until the canvas exists.
2. **Bubbles** — the canvas proper: `NSScrollView` bridge, persons / events / places, working labels, drag / snap / select / persist, the tray for unplaced subjects. Replaces the stub, and ships the accessibility representation and keyboard parity alongside the first bubbles (§7.4).
3. **Subject vocabulary** — `properties`, `subject_type_fields`, the Subject types / Subject fields editors (replacing nav stubs), `ref_prefix` validation.
4. **Artifact viewer + Citations** — PDF and image; `page`, `region`, `text_quote`; Go-side locator validation.
5. **First Observation** — Add Property on a bubble, `text` value type only. The whole vertical path proven end to end.
6. **Remaining value types** — including NameValue end to end; reuse the existing DateValue editor.
7. **Connect tool** — bridge macros, the disambiguation form, the pinned Citation (§6.1).
8. **Honesty and polish** — negated / conflicted / uncited states, filtering, undo. Accessibility is *not* here; it moved to slice 2 (§7.4).

**Spike boundary: slice 1 is [Spike 5](../deployment-plan/archive/spike-5/), slice 2 is [Spike 6](../deployment-plan/spike-6/).** Splitting foundation from canvas keeps the canvas spike pure — Spike 6's first PR draws a bubble rather than writing a migration. Spike 6 also prototypes **connect lines** as UI risk (not Observation macros); see that plan's S6-04 scope note.

**The cost of the split is that slice 1 ships a layer with no visible capability, and that cost is now accepted rather than bought off.** An earlier draft of this note put a plain list of a Source's Subjects at the graph destination to prove the data path, justified as the structured non-canvas editing path §7.4 was committed to. Since the graph is now the only surface (§1.3), that list would be throwaway UI — built, designed, and then deleted by slice 2 — so it is dropped.

What slice 1 ships instead is dual action on the Sources list plus an Evidence graph stub, because the navigation contract already has a Sources root (§11.1.2) and inventing a second Interpretation list was the wrong product shape (§1.4). Opening Evidence graph can land on a stub / "coming soon" destination — early development; Spike 6 replaces the view with the canvas.

The data path — migration → Go → FFI → store → session cache — is therefore proven by Go tests and a `FakeStore` round-trip rather than by a screen. That is a real reduction in confidence compared to seeing a row render, and it is the deliberate price of the first visible subject being a bubble. The mitigation is that slice 2 opens against a fully tested rail, so a canvas bug in slice 2 is a canvas bug and not an ambiguity about which layer is broken.

---

# 12. Naming

**Decided: Evidence graph** for the spatial surface.

Earlier drafts used **Interpretation graph** to match the data-model layer name. That still works in engine docs, but once entry is dual action on the Sources list the researcher already knows they are inside one Source — so a product name can emphasize *evidence* (what this document shows) without repeating "interpretation." **Evidence graph** is the working UI name for chrome, actions, stubs, and breadcrumbs.

Layer vocabulary *inside* the graph stays Citations, Observations, and Subjects (engine / schema words).

**Product sidebar vocabulary** (§1.4):

| Product | Engine |
| --- | --- |
| Evidence graph | Interpretation-layer canvas (Citations / Observations / Subjects) |
| Subject types | `subject_types` |
| Subject fields | `properties` (+ `subject_type_fields`) |
| Open Evidence graph (action) | open the graph place |

Do not use **claim** in this family — Conclusion owns that word.

Considered and set aside for the canvas name:

| Candidate | Why not (as the *product* canvas name) |
| --- | --- |
| Interpretation graph | Accurate to the model layer; heavier once the user is already Source-scoped. Keep as the brainstorm / engine alias. |
| Interpretation map | "Map" reads well but adds a term the data model does not use. |
| Source map / Source graph | Emphasizes scope, but "map" sounds geographic beside `place` Subjects; "source graph" is fine technically, less friendly. |
| Worksheet | Plays down the visual; too humble for the primary surface. |

---

# 13. Reuse: the eventual family tree view

A family tree is the same idea — subjects, edges, spatial layout, click a thing to see what supports it — so the canvas built here is very likely the canvas built there. That is worth planning for, but the reuse is not where it first appears to be, and one piece of it is a trap.

## 13.1 What already shares a spine

Two things are shared *by design* in the model docs, before any canvas exists:

- **Subject type vocabulary.** `canonical_entities.subject_type_id REFERENCES subject_types(id)` ([`conclusion-layer-data-model.md`](../conclusion-layer-data-model.md) §2.1) — the same seven rows Spike 5 seeds. A Person in the tree and a candidate person subject share one type row.
- **Ref prefixes.** `subject_types` is shared vocabulary for both layers (interpretation model §4.1), carrying `ref_prefix` for the canonical entity (`PER-7KD45`) and `candidate_ref_prefix` for the subject (`CPR-7KD45`). A family tree view reads canonical refs and an interpretation canvas reads candidate refs off the same type row, so `Mint` and the reserved-prefix guard already serve both.

So the foundation spike is already doing family-tree work without trying to.

## 13.2 The reuse is in the geometry, not the storage

The expensive, risky part of the canvas is not persistence — it is the interaction surface: the `NSScrollView` bridge, coordinate conversion under magnification (§7.2), hit-testing, snap-to-grid, drag, selection, edge routing, keyboard parity, and the accessibility representation (§7.4). Storage is roughly forty lines of SQL and a small Go package.

That is the right split, and §7.2 and §7.4 already argue for it on testability grounds: keep the geometry as **pure functions over value types**, with no view and no store. What makes those functions unit-testable is exactly what makes them reusable for a second graph over a different table.

**The consequence is for the canvas spike, not this one: put the canvas primitives in a neutral module from day one.** If they are born inside `Features/Interpretation/` and reach for `CatalogSubject` directly, extracting them later is a refactor across the riskiest code in the app. A generic layer parameterized over "a thing with an id, a type, a label, and a grid position" costs nothing extra to write first and is very expensive to retrofit.

## 13.3 Do not generalize the position table

The tempting move — one `graph_positions` table with a `graph_id` and a polymorphic subject — is the trap. Interpretation positions `subjects.id`; a tree positions `canonical_entities.id`. They are different tables, so a shared subject column cannot carry a foreign key.

That forfeits the single best property of the design in §5.1: **`subject_id` cascades**, so deleting a subject silently drops its position. A polymorphic subject replaces that with application-side cleanup and tolerated orphan rows — in exchange for avoiding a second forty-line table.

There is also no shared graph scope to key on. Interpretation's scope is a Source; a tree's is the whole project, or a named view, or a starting person. `source_id` is not a column a tree can fill.

So: **two concrete tables, `subject_positions` and a future `canonical_entity_positions`.** They will have near-identical Go packages, and that duplication is the cheap kind — no polymorphism, no orphans, full referential integrity, and each one deletable without touching the other. (This is also why the Spike 5 table is *not* called `graph_subject_positions`: §5.1 explicitly declines to introduce a `graphs` entity, so the name should not imply one.)

## 13.4 Two things that will not transfer

Worth knowing now, because both are places where a decision taken here is right for Interpretation and wrong for a tree:

- **The tray instead of auto-layout** (§5). A researcher hand-placing tens of Subjects per Source is the premise. Nobody hand-places four hundred ancestors, and generational rows and sibling ordering are not aesthetic preferences in a tree — they carry meaning. **A family tree forces the auto-layout engine this note deliberately refused.** That is a genuine cost the tree pays, not a shortfall in the tray decision.
- **The rendering strategy** (§7.1). The `Canvas`-plus-`ZStack`-of-real-views hybrid works precisely *because* Source-scoped graphs are small; the same section says outright that it "would not handle a whole-project graph." A tree is a whole-project graph. It will need viewport culling or virtualization, which means the node-rendering layer is the one part of the canvas most likely to be rewritten rather than reused.

The pattern in both: what transfers is the **plumbing** — coordinates, gestures, hit-testing, persistence shape, accessibility. What does not is anything whose design was justified by Interpretation's bounded scope.

---

# 14. Open questions

Answered during this brainstorm, recorded so they are not reopened by accident: whether the canvas shows `source` subjects (no, §4.6), whether Citation→Observation should be 1:1 (no, §4.5), whether layout lives in the catalog (yes, unaudited, §5.1), whether raising the deployment target unlocks the canvas (no, §7.2), and whether the property vocabulary blocks the canvas (no, §11.2).

Still open:

- Does the canvas *create* root subjects only, or also adopt Subjects created elsewhere (imports)? Cross-Source adoption is answered in §4.6: no.
- Can one graph span Sources (a "case view")? Source scope should be hard for editing; a read-only multi-Source view is a different feature and would need the Conclusion layer to be meaningful.
- What does the Source-page commentary surface actually look like? `mentions` and `remark` Observations stay in the data model (§4.6) but now need a home that is not the canvas, and it is unspecified.
- Before `mentions` ships: does mentioning an unheld document create a placeholder Source, and what resolves two placeholders that turn out to be the same document? Source merge does not exist (§4.6).
- Does the person → person disambiguation (§3.2) earn its complexity, or should person → person simply always mean `relationship` and let shared-event modelling go through the event bubble?
- Is a Citation with zero Observations a legal, useful state — "I transcribed this line, I have not interpreted it yet" — or should the composer refuse to save a Citation that asserts nothing? This is the one real loose end left by keeping one-to-many (§4.5), and it is a UI policy question rather than a schema one.
- How does Conclusion-layer work ([`conclusion-layer-data-model.md`](../conclusion-layer-data-model.md)) surface here later — same canvas with a layer toggle, or a separate reconciliation view? "Not now" is fine; "never" would be a mistake.
