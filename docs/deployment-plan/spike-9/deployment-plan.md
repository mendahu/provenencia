# Deployment Plan — Spike 9

MVP for the **Conclusion layer**: assemble canonical Persons, Events, and Places from Interpretation Subjects through Identity Claims, and show them on list and detail pages. Authoritative model: [`conclusion-layer-data-model.md`](../../conclusion-layer-data-model.md). Reconciliation: [`conclusion-reconciliation.md`](../../conclusion-reconciliation.md). Values: [`structured-name-model.md`](../../structured-name-model.md), [`structured-date-model.md`](../../structured-date-model.md). Vocabulary: [`seeded-vocabulary.md`](../../seeded-vocabulary.md) §3, §5.

## Status

**Open.** Slices 1–3 landed. **Replanned 2026-10-05** from slice 4 on: the reconciliation design ([`conclusion-reconciliation.md`](../../conclusion-reconciliation.md)) replaces R2's name / date / ranking plan with one reconciler pipeline for every value type, adds per-Property cardinality, and brings place hierarchy into the spike. Briefs S9-D1…D14 in [`design/`](design/) (D5 and D7 need revising). Landings go in [`completed.md`](completed.md).

> **Goal of this spike:** a researcher can promote Subjects off an Evidence graph into Persons, Events, and Places, and open a page for each that shows who or what it is — name and life dates, event and date, place names and where the place sits — reconciled from every member Subject, with the reasoning shown.

> **Foundation, grown in vertical slices.** The resolved-values cache (R3) is core infrastructure every Conclusion surface reads — lists, details, Promote, search, and later the tree. Build it properly, but grow it slice by slice so each layer is exercised in the app as soon as it lands. Stubs and temporarily incomplete UI along the way are fine; no band-aid caches per screen.

## Goal (dogfood bar)

**By spike close**, each of these must be true in the app on real research:

1. **Promote a person from a birth record.** From a person card on the Evidence graph, Promote mints a new Person. The walk then offers the birth event, its participation, its place, and its location. Each can join an existing handle or mint a new one. Each step saves on Next; **Done** stops the walk with everything so far kept.
2. **Promote a second record onto the same Person.** A census person joins the existing Person. The comparison lines up name (and anything else comparable) against the current members. Confirmed pairs are pinned on both claims.
3. **Browse.** Persons, Events, and Places appear in the sidebar. Each list shows every handle of that kind with the row fields in R5: Persons with name, birth–death dates, birth and death places; Events with a derived name, date, and place; Places with a name and parent chain. Every row has a thumbnail slot and its ref.
4. **Person detail** shows the thumbnail slot, the reconciled name, the birth and death dates, and the birth and death places. Two members saying `14 MAY 1985` and `MAY 1985` show one date. Two members saying `MAY 1985` and `APR 1985` show that they disagree. Every value can explain itself: each record considered and what happened to it (kept, folded, outvoted, weak, denied).
5. **Event detail** shows the thumbnail slot, the event title, the event date, and the event place(s).
6. **Place detail** shows the thumbnail slot, the Place's names (concurrent names all kept; spellings and case merged), its period, the places it is part of and the places part of it, and what it succeeded or was succeeded by.
7. **Places sit in a hierarchy, the hard way.** On the Evidence graph a researcher draws "part of" (administrative, geographic, ecclesiastical) and "succeeded by" between place subjects, each cited like any evidence. A birth place reads with its chain at the birth date: *Toronto, Province of Canada* in 1850, *Toronto, Ontario, Canada* in 1950; a date that straddles a change shows both (*Toronto, Upper Canada or Province of Canada*).
8. **The graph knows.** A promoted subject's card shows the handle it belongs to and links to its page. Promote is not offered twice for a subject that is already a member.
9. **Search finds them.** Typing `PER-7KD45`, *James Robins*, *Jim Robins* (an alternate), *Birth of James*, or a toponym in the omnibar lists the handle with a row that reads like its list row; selecting it opens the detail page. Editing a member's name Observation updates the hit.
10. **The cache is honest.** A full rebuild of the resolved-values cache produces exactly what incremental upkeep produced, after any sequence of writes (R3 test). Timings on the deep fixture are recorded in the [performance ledger](performance-ledger.md).

---

## Architecture

```text
Interpretation (truth)          Conclusion (truth)
 subjects, observations,         canonical_entities,
 citations, sources, …           identity_claims (+ evidence)
            │                              │
            └────── reconciler (R2) ───────┘
                          │   written in the same transaction as the change (R3)
                          ▼
            conclusion_resolved_values   ← the one derived cache; rebuildable
              scalar Properties  (name, date, toponym, event_type, role, …)
              entity-valued ends (Participation.person → PER-1, …) = the canonical graph
            conclusion_resolved_candidates  ← the reasoning: every candidate's outcome
                          │
                          ▼
            composers in Go (R4)  — entity headers, walks, naming matrix
              │         │          │            │
           lists (R5) details (R6) Promote (R7) search docs (R8)   place chains (R9)   … tree later
```

- **Truth tables stay truth.** The cache is derived, rebuildable, and never referenced by truth data.
- **One cache, many compositions.** A new screen adds a composer, not a cache.
- **Stop at resolved values.** Below them, indexed joins over bounded sets; above them, a few indexed lookups. A second derived tier or an in-memory copy is added only if the ledger's timings demand it.
- **Go deals in structures; Swift makes text.** Payloads carry DateValues, NameValues, term ids, and title *parts* — never display strings. Formatting a date, a name, or an event title is a UI job ([`L10n`](../../macos-client-patterns.md), naming-matrix templates). The one exception is search: FTS needs text, so Go writes match text into the index (see R8).

---

## Requirements

### R1 — Conclusion schema and Go core

The Conclusion tables do not exist yet. Ship the subset this spike writes.

| | |
| --- | --- |
| **In** | `claim_confidence_grades` (+ seed §5.5); `canonical_entities`; `identity_claims` (composite FKs on `subject_type_id`, `UNIQUE (subject_id, entity_id)`, partial unique index on one accepted claim per subject); `identity_claim_evidence`. Canonical refs minted from `subject_types.ref_prefix`. Seed Property `event_name` (`text`) bound to `event` (Q13) — in the Install seed and via migration for existing projects. Audit on every write. Go packages + FFI handlers for create / read / list. |
| **Out** | `reconciliation_claims` and its evidence table; `canonical_entity_notes`; merge re-pointing; `project_settings` / `name_format_profiles`. |
| **Delete Impact** | Non-blocking `cascades`: a Subject with an Identity Claim names the handle it would leave; a pinned Observation names the handles whose claims lose it. Neither delete is refused. Claims and pins are removed **explicitly** in the delete transaction and audited; the schema `CASCADE`s are backstops. Weakened claims go to the review alert (model §5.2, Spike 10). Replace the stale `ViaSamenessEvidence` reservation. |
| **Derived** | The resolved-values cache (R3) and search documents (R8). |

Skills: [`add-catalog-migration`](../../../.cursor/skills/add-catalog-migration/SKILL.md), [`add-catalog-ref`](../../../.cursor/skills/add-catalog-ref/SKILL.md), [`add-seeded-vocabulary`](../../../.cursor/skills/add-seeded-vocabulary/SKILL.md), [`add-ffi-handler`](../../../.cursor/skills/add-ffi-handler/SKILL.md), [`add-catalog-delete`](../../../.cursor/skills/add-catalog-delete/SKILL.md).

### R2 — Value reconciliation

**Authoritative design: [`conclusion-reconciliation.md`](../../conclusion-reconciliation.md).** This section is the spike's slice of it.

**Every Property on a canonical entity is a list** — many members, each with many Observations — and every value type needs an **auto-reconciler**. The reconciler is pure Go: candidates in, the answer and its reasoning out. Its output lives only in the derived cache (R3); it never creates a claim or truth row.

- **One shared pipeline** (design §5): admit (provisional members and empty values eliminated) → deny (stronger negatives) → group (same value, fold) → majority (counted in **distinct Sources**) → confidence (weak values drop when stronger evidence disagrees) → merge.
- **A small module per value type** (design §7) answers *same value*, *fold* and *merge*. This spike builds all six: **text** (trimmed, case-insensitive), **integer**, **term** (neutral terms are no evidence), **name** (§7.2: structured parts only, format-agnostic, subsumption), **date** (§7.1: containment, overlapping windows), **subject** (ends map to handles; the canonical graph).
- **Evidence per candidate** (design §4): Source credibility, transcription certainty, Identity Claim confidence (each against its default grade), claim status, polarity, and the Source it came from. Promote writes `accepted` only this spike (Q7), so provisional members never occur yet; the pipeline handles them.
- **Reasoning is output and is cached** (design §6): every candidate's outcome and reason (`kept`, `folded`, `outvoted`, `weak`, `denied`, `provisional`, `no_evidence`, `against`).
- **Cardinality** (design §8): each Property is single- or multi-valued. Seeded Properties are single except `toponym`. Multi-valued Properties keep every distinct surviving value.
- **States:** single, merged, mixed, empty, read off the result's shape. **Concluded** needs Reconciliation Claims, which stay out of scope (design §8: one claim per value for multi-valued Properties); the reconciler takes a concluded input so they plug in later.

The modules' *same value* test also drives the Promote comparison's "compatible pairs start checked" (R7 step 3).

### R3 — Resolved-values cache (foundation)

One derived table holds the resolver's output for every Property on every handle. Everything above it is composition.

```text
conclusion_resolved_values
  entity_id        BLOB  → canonical_entities (CASCADE)
  property_id      BLOB  → properties         -- vocabulary as a row, never a column
  rank             INT   -- 1 = displayed value; 2..n = other distinct clusters
  value_text / value_integer / value_term_id / value_entity_id
  value_date       BLOB  -- protobuf DateValue (may be synthesized; never a date_values row)
  value_name       BLOB  -- protobuf NameValue (same)
  date_lo          INT   -- earliest day the resolved date could fall on (ordinal; NULL = open)
  date_hi          INT   -- latest day (ordinal; NULL = open)
  sort_key         TEXT  -- normalized text, or date ordinal
  support          INT   -- distinct Sources backing this value
  against          INT   -- negative candidates that match it
  reason           TEXT  -- 'kept' when displayed, else why not (outvoted, weak, denied, provisional)
  PRIMARY KEY (entity_id, property_id, rank)
  INDEX (property_id, value_entity_id)   -- reverse edges: who points at PER-1?
  INDEX (property_id, sort_key)          -- sort, resemblance
  INDEX (property_id, date_lo, date_hi)  -- date windows: "born 1800–1850", timelines
```

- **Useful for sorting, querying, and resolving (Q12).** Structured values are stored whole as the protobuf messages FFI already uses (`DateValueInput` / `NameValueInput` in [`engine.proto`](../../../api/proto/engine.proto); renaming them to plain value messages is optional cleanup). SQL never decodes them: ordering uses `sort_key`, date range queries use `date_lo` / `date_hi` (the window the date resolver already computes to merge), and the resolver and composers decode in Go. Resolver output is never written to `date_values` / `name_values`.

- **State is the shape of the clusters, never stored.** The resolver (S9-05) returns a list of clusters plus a *concluded* flag; readers derive `single` / `merged` / `mixed` from the rows (rank 2 exists → mixed; else rank 1's `support` 1 → single, more → merged). A `concluded` flag column arrives with Reconciliation.
- **Edges are resolved values.** A Participation's `person` end is its resolved `person` Property with `value_entity_id`. The canonical graph is this table plus its reverse index; there is no separate edges table.
- **Nothing is dropped.** Every distinct value is a row, displayed or not; `reason` says which (S9-13). Displayed rows come first, so rank 1 is the first displayed value. State and *+N* read displayed rows; search and matching read every row.
- **The per-candidate reasoning is kept beside the values** in `conclusion_resolved_candidates` (S9-14): one row per candidate (entity, Property, Observation) with its outcome, reason and the rank of the value it supports or folded into. Detail pages explain a value from it without re-running the reconciler. Rewritten with its handle's rows.
- **Every value is kept**, not just the displayed one: search gets alternates (*Jim*), Promote compares against every value, details render values without re-resolving.
- **No vocabulary in the schema.** No `birth_date` columns, no per-kind tables. Genealogical concepts live in Go (resolver, composers, registry of recognized keys) and in `property_id` rows. Labels are not stored (term **ids** are), so relabels need no recompute.
- **Maintained in the write transaction.** Every trigger is one hop from the write:

  | Write | Recompute |
  | --- | --- |
  | Observation save / delete on subject S, Property P | S's handle (every Property — see below) |
  | Identity Claim S → E created (Promote) or removed (Subject delete CASCADE) | every Property of E; plus (E′, P′) for handles whose members' Observations point at S (their end now maps differently) |
  | Identity Claim confidence or status change | every Property of that handle (no edit path this spike) |
  | Citation certainty / Source credibility change | handles with member Observations under it |
  | Property cardinality change | every handle carrying that Property |
  | Reconciliation Claim (later) | (E, P) → `concluded` |
  | Term relabel | nothing |

  Recompute is **per handle** for now: a trigger rewrites all of the handle's Properties through the same batched loader as rebuild. Narrowing to (E, P) waits for timings that call for it (S9-33).

- **Rebuildable.** A version row (`conclusion_resolved_meta.cache_version` against `resolvedvalues.CacheVersion`, like `catalog_search_meta`); a mismatch on Open rebuilds from truth tables. Bumping the version is the development rebuild — there is no separate command.
- **Proven.** Rebuild-equals-upkeep tests apply write sequences through the real write APIs, then check incremental upkeep equals a full rebuild: long sequences generated from **fixed** seeds (deterministic; a failure names its seed and step) plus named hand-written scenarios. A failure a seed finds is shrunk into a scenario. Ships in the same PR as the cache.
- **Batched loading.** The resolver's inputs (members, candidates, provenance) load in a fixed number of set-based queries per batch of handles — no per-Property or per-member loops. The same loader serves rebuild and incremental upkeep.

### R4 — Composers (headers, walks, derived values)

Screen-shaped values are **composed in Go** from R3 at read time; they are not stored. One **entity header composer** per kind serves every surface that shows a handle as a row: lists (R5), Promote's target picker (R7), omnibar hits and documents (R8), and later tree nodes.

Walks follow **canonical** edges only (Q3). Birth / death use `role = subject` Participations only (Q4).

- **Person birth / death:** Person ← Participation (`person` end, resolved `role = subject`) → Event with resolved `event_type = birth` / `death` → its resolved `date` (else `start_date`).
- **Person birth / death place:** that Event ← Location (`event` end) → Place (`place` end) → its names and its **chain at the Event's date** (R9).
- **Event date:** resolved `date`, else `start_date`–`end_date`.
- **Event place:** Event ← Locations → Places → names and chain at the Event's date. Several Locations (York *and* Upper Canada) are all returned.
- **Event name:** composed from `event_type` and the people on it through the **naming matrix** (below).
- **Place:** its names, period, and hierarchy (R9).

Each hop is an indexed lookup on R3. List composers run set-based over the whole list, not per row.

#### Event naming (Q10)

**Convention: `role = subject` means the event's principal(s), and there may be several.** Every other role describes someone's connection *to* the principals. A marriage has **two `subject` Participations**, not `spouse`; `spouse` is for a principal's spouse named on another event (the widow on a death record). This is also the rule every Person walk uses (Q4).

**Precedence.** An event's title is chosen in this order; the first that yields a value wins:

1. **Recorded name** — the resolved `event_name` Property (text; seeded this spike, Q13). For historically named events (*The Great Fire of 1849*), cited like any Observation or, later, concluded by a Reconciliation Claim.
2. **Composed from subjects** — the matrix below, when the event has at least one `subject`.
3. **`label`** — the researcher's working handle, for events with no recorded name and no subject (*Grandpa's house fire*).
4. **Composed without subjects** — type and place (*Fire at York*), or *Unspecified fire*.
5. **`ref`**.

**Matrix.** Go returns the parts (event type key + label, resolved names of the `subject` Persons in a stable order, resolved place); Swift formats them through L10n templates keyed by `event_type`, so the matrix is localizable and lives in one table.

| Case | Template | Example |
| --- | --- | --- |
| One subject | *{Type} of {subject}* | *Birth of James Robins*, *Census of James Robins* |
| Two subjects, `marriage` | *Marriage of {a} and {b}* | *Marriage of James Robins and Mary Smith* |
| More subjects (any other type, or 3+ on a marriage) | *{Type} of {first} et al.* | *Census of James Robins et al.* |
| No subject, has place | *{Type} at {toponym}* | *Fire at York* |
| No subject, no place | *Unspecified {type}* | *Unspecified fire* |
| No `event_type` | *Event* (then the rules above) | *Event at York* |
| Researcher-added type | the same generic templates | *Emigration of James Robins* |

- A subject Person with no resolved name reads as *unnamed person* (*Birth of unnamed person*); the ref stays in the row.
- Subject order is stable (a resolver test-table decision), so *et al.* always names the same person.
- Per-term custom templates are later, with term editing.
- The title does not carry a *mixed* indicator; the Person's or Event's page does.

### R5 — List pages (Persons, Events, Places)

Three sidebar destinations under a new **Conclude** section (sidebar sections: Source, Conclude, Narrate later, Configure — S9-D1), with counts ([`CatalogCounts`](../../../macos/App/Features/Catalog/CatalogCounts.swift)). Rows come from the R4 header composers (Q6).

| List | Row |
| --- | --- |
| **Persons** | thumbnail slot · name · birth date – death date · birth place · death place · ref (`PER-…`) |
| **Events** | thumbnail slot · event title (R4 precedence) · event date · event place · ref (`EVT-…`) |
| **Places** | thumbnail slot · name · parent chain (today's) · ref (`PLC-…`) |

- **One value per cell.** The rank-1 value, with a *+N* count when there are more clusters (toponyms, Locations). **No *mixed* marker in rows** (S9-D2 decision): a mixed value shows its top-ranked value, and disagreement is surfaced on the detail page.
- Name / toponym fall back to `label` → `ref`. Missing dates and places are empty, not "Unknown."
- Sort: Persons by name, Events by date, Places by toponym (R3 `sort_key`; default only, no sort controls this spike). Persons sort by the normalized text of the reconciled name this spike (*James Robins* files under J); surname-first sorting and the *Robins, James* list style come with name format profiles ([`structured-name-model.md`](../../structured-name-model.md) §4.5).
- Only `person`, `event`, and `place` get pages. Association handles exist but are not listed.
- Skills: [`add-workspace-place`](../../../.cursor/skills/add-workspace-place/SKILL.md), [`add-catalog-query`](../../../.cursor/skills/add-catalog-query/SKILL.md).

### R6 — Detail pages (Person, Event, Place)

- **Person:** thumbnail slot, name, birth date, death date, birth place, death place.
- **Event:** thumbnail slot, event title, event date, event place(s).
- A detail page is always a superset of its list row (Q11).
- **Place:** thumbnail slot, names, period, parents (by type and period), parts, succession.
- Every value shows its state — single, merged, mixed, or empty — its ranked values, and **why**: the candidates and their outcomes from R3's reasoning. Drilling into the Observations behind a cluster loads live for that one handle. A member list (Subjects with their Source) is a stretch goal.
- Read-only this spike.

### R7 — Promote (Identity Claim workflow)

The hard part. Model: [`conclusion-layer-data-model.md`](../../conclusion-layer-data-model.md) §5.1–§5.4.

**Promote only creates claims.** It is one-way: nobody comes back into it to edit, re-pin, or remove a claim. Those are a separate workflow with its own place (Spike 10), even though they write the same tables. No resume-midstream affordances.

1. **Entry.** A **Promote** control at the bottom of a subject card on the Evidence graph, for primary kinds (`person`, `event`, `place`) that have no accepted Identity Claim. Bridge cards are not entry points; they join through the walk.
2. **Choose target.** Mint a new handle, or pick an existing one of the same Subject type. The picker shows R4 headers. Suggestions: handles already related to the one just filed (during the walk), then resemblance (resolved name, date, toponym, event type — scored by per-kind match profiles over R3, see [`docs/matching.md`](../../matching.md); the R8 index can narrow candidates).
3. **Compare (existing handle only).** Line up the incoming Subject's Observations against each accepted member's Observations for the same Property. Compatible pairs start checked (the value-type module's *same value* or *fold*, R2); differing pairs start unchecked. Accept all checked for a Property in one gesture, clear a pair, or skip and accept with no pins. Checked pairs are pinned; unchecked pairs are simply not pinned.
   - **Minting a new handle skips this step.** The grounding claim normally has **zero pins**. Its Observations get pinned later, by backfill, when a second Subject joins and confirms against it (§5.1).
4. **Claim fields.** A **Status** dropdown is in the layout now, with one option — `accepted` (Q7). Later spikes add `provisional` / `rejected` to it without a relayout. Optional confidence grade. `argument` drafted from the confirmed rows when there are any; editable; empty is fine on grounding.
5. **Next = save this step.** One transaction per promoted Subject: the new handle (if minted), the Identity Claim, its pins, **backfill** pins onto the compared members' claims (§5.1), and the R3 recompute. A person and its birth event are two transactions. A failure (for example the one-accepted-claim index losing a race) fails that step only.
6. **Walk.** After a step saves, show the Subjects connected to the one just filed and offer each for promotion. Picking one repeats steps 2–5 for it; its unpromoted Interpretation neighbors are appended to the queue if they are not already queued, not handled in this pass, not members of the target, not neighbors of other members. Declining a subject does not reach through it. Bridge subjects (`participation`, `location`, `relationship`) wait until both ends have a handle, then become a short confirm: file onto the association the two ends already share, or mint one.
7. **Off-ramp.** **Done** ends the walk at any point. Earlier steps are already saved; the unvisited queue is not persisted and those Subjects stay unpromoted on the graph. The leave guard (`WorkspaceLeaveGuard`) covers only the unsaved step in progress. There is no rollback; corrections are Spike 10.
8. **After.** The graph card shows membership (handle ref, link to its page). Conclusion caches refresh (see *Data fetching*).

### R8 — Omnibar search

Persons, Events, and Places become omnibar hit kinds (Q8). Contract: [`omnibar-search.md`](../archive/spike-3/omnibar-search.md). Skill: [`add-searchable-kind`](../../../.cursor/skills/add-searchable-kind/SKILL.md).

- **Kinds:** `person`, `event`, `place` in the Go registry, each mapping to its detail place (`WorkspaceLocation` section + entity id). Context boost when the current section is that kind's list or detail.
- **Refs:** exact and prefix match on `PER-…` / `EVT-…` / `PLC-…` through the existing ref fast path.
- **Documents from the header composer.** One FTS document per handle, built by R4 so a hit can never disagree with its list row:

  | Kind | `title` | `secondary` | `body` |
  | --- | --- | --- | --- |
  | Person | resolved name | birth–death years, birth / death place | every name cluster (so *Jim* finds James), `label` |
  | Event | R4 title (precedence) | date, place | recorded-name clusters, event type label, subject Person names, `label` |
  | Place | first name | parent chain (today's) | every name, `label` |

  Both FTS indexes (unicode61 + trigram) stay in sync; bump `ProjectionVersion`.
- **Hit rows.** `PVOmnibarHitRow` per kind: thumbnail placeholder, title, kind, ref, subtitle from the header (Person: birth–death; Event: date · place). Conclusion hits carry the **structured header** and Swift formats it, like the list row; Go's FTS text is for matching only. This extends the `SearchHit` DTO (today's kinds send Go-built `display_title` strings).
- **Reprojection.** Search documents are stored (FTS needs them) and composed across handles, so they have one more level of dependency than R3. When R3 rows for handle E change in a transaction, reproject E and its **header dependents** — handles whose header composer reads E, found by the fixed reverse walk over R3's reverse index (Place → Locations → Events → subject Persons; Person → subject Participations → Events). The dependents function lives next to the composers and is covered by the same rebuild-and-compare test. `ProjectionVersion` bump → full rebuild, as today.
- **Out:** Interpretation Subjects (`CPR-…`) as hits; cross-root association queries ("John Smith *with* a birth certificate", omnibar-search *Associated entities*).

### R9 — Place hierarchy (the hard way)

Design: [`conclusion-reconciliation.md`](../../conclusion-reconciliation.md) §11.1. A gazetteer that writes this evidence automatically is a later, optional service ([`ideas/place-gazetteer-service.md`](../../ideas/place-gazetteer-service.md)); this spike does it by hand, cited.

- **Places are separate entities** at any grain the research needs (township, county, region, farm). No place kinds this spike: the hierarchy says what a place is part of, and a kind label can come later.
- **Period:** a Place's `start_date` / `end_date` (bound to `place`) say when it existed or mattered; both optional; none = always.
- **Place relationship**, a new association kind (bridge) with `from` (place), `to` (place), `place_relationship_type` (term). Types are researcher-extensible; each term has a **category**:
  - **hierarchical** — seeded `administrative`, `geographic`, `ecclesiastical`. A link holds where the two places' periods overlap.
  - **temporal** — seeded `succeeded_by` (York → Toronto). A lineage for history and search; never a display chain.
- **Drawn on the Evidence graph** like any bridge (a `connectrules` entry, disambiguated by type), cited, and promoted like Location.
- **Chains at a date.** The composer walks hierarchical parents whose periods hold at the date (an Event's reconciled date, or today). When the date can't decide, every candidate is returned (*Upper Canada or Province of Canada*). A place may have several parents; the display follows `administrative` first.
- **No loops**, checked in Go when a place relationship is filed: a hierarchical link that would close a cycle, or a succession that loops, is refused. Composers guard anyway.
- **Names:** `toponym` is multi-valued (concurrent names). A rename is a new Place linked by `succeeded_by`.

---

## Data fetching

The Conclusion pages read **across** Sources — every earlier place scoped to one Source or one vocabulary. R3 absorbs that cost at write time.

- **Go composes; Swift renders.** Swift asks for "Persons list" or "Person X" and displays the payload. It never assembles members or walks ([`macos-client-patterns.md`](../../macos-client-patterns.md) *One cache owns each list*).
- **Query keys:** one list key per kind, one detail key per handle, and the Promote reads (target suggestions, comparison rows for a subject × entity, the neighborhood of a subject). The Promote draft is interaction state, not a cache key.
- **Swift invalidation: bust all Conclusion keys (Q5).** Any Interpretation or Conclusion write that R3 recomputes marks every cached Conclusion key stale — cheap, because reloading reads R3.
  - **Triggers:** Observation save / delete, Subject delete, bridge create / delete, Source delete, credibility / certainty changes, Property configuration changes, and each Promote step. Shipped in S9-07 as one set, `CatalogQueryRegistry.conclusionTriggers` (`savedCitation`, `deletedSubject`, `promotedSubject`, `deletedSource`, `mutatedSourceWorkspace` — the last carries credibility and over-busts on notes / artifacts / metadata). Certainty already rides `savedCitation`; a cardinality change (S9-37) joins via `updatedProperty`.
  - **Evict, don't revalidate.** Detail keys are dropped and reload when next visited; the visible page revalidates. A deliberate exception to cache contract rule 2, noted in [`macos-client-patterns.md`](../../macos-client-patterns.md). **Built in S9-15** with the first detail key: the session can't evict or tell which place is visible yet, and list keys (S9-07) follow rule 2 as usual.
  - Narrowing later can reuse R3's affected-handles set; not needed while reloads are cheap.
- **Measure.** Build the deep fixture (a Person on ~10 Sources; birth, death, marriage events each multi-member, with Locations; a few relationships; scaled to a few hundred handles) and time: rebuild, a single write's upkeep, list composition, detail composition. Record in the [performance ledger](performance-ledger.md).

---

## Design track

**Every UI PR is gated by Claude Design briefs — one brief per view.** A PR that touches two views waits on two briefs. Briefs follow [`add-design-brief`](../../../.cursor/skills/add-design-brief/SKILL.md) and live in [`design/`](design/). **Each brief is designed alongside its feature**, just before the PR it gates — see the [PR sequence](#pr-sequence).

| Brief | View | Gates | Later PRs on the same view |
| --- | --- | --- | --- |
| **S9-D8** | Evidence graph subject card | **S9-04** | S9-09 (name on the membership row), S9-11 (Promote opens the flow) |
| **S9-D1** | Workspace sidebar | **S9-08** | S9-23, S9-26 (Events / Places go live) |
| **S9-D2** | Persons list | **S9-09** | S9-32 (life dates and places) |
| **S9-D9** | Promote — choose target (+ shell) | **S9-11** | — |
| **S9-D10** | Promote — claim fields | **S9-12** | — |
| **S9-D5** | Person detail (**revise:** reasoning) | **S9-16** | S9-32 (life dates and places) |
| **S9-D11** | Promote — compare | **S9-19** | — |
| **S9-D3** | Events list | **S9-23** | S9-32 (subject titles, places) |
| **S9-D6** | Event detail | **S9-24** | S9-32 (subject titles, places) |
| **S9-D14** | Properties — cardinality | **S9-37** | — |
| **S9-D4** | Places list (**revise:** names, parent chain) | **S9-26** | S9-40 (chain cell) |
| **S9-D7** | Place detail (**revise:** names, period, hierarchy, succession) | **S9-27** | S9-40 (hierarchy rows) |
| **S9-D15** | Custom term dialog — category | **S9-38b** | — |
| **S9-D12** | Promote — walk | **S9-30** | — |
| **S9-D13** | Omnibar results | **S9-35** | — |

A brief covers its **whole** view, including cells a first PR leaves empty. Later PRs on the same view fill in designed frames and need no new brief.

---

## PR sequence

**Vertical slices.** Every slice ends in something visible in the app, so a mistake in a lower layer (schema, resolver, cache) shows up in the slice that introduces it — not weeks later. The foundation is grown, not front-loaded: the resolver, cache, composers, and formatters each start small in the first slice that needs them and gain capability in later slices. Nothing is throwaway; each PR extends the last.

**✎ = design brief**, run through Claude Design just before the PR it points at. **Check** = what you can verify in the app when the slice lands.

```text
SLICE 1 — Mint from the graph
  S9-01  Conclusion schema + stores
  S9-02  Promote write v1: mint + claim; Subject delete names its handle
  S9-03  Source graph carries membership
  ✎ S9-D8 ──▶ S9-04  Graph card: Promote (confirm-and-mint) + membership (ref)
  Check: Promote a person card → it shows PER-…; delete that subject → Impact names PER-….

SLICE 2 — Persons list
  S9-05  Resolver core v1: clusters, states, stable order (pure Go)
  S9-06  Resolved-values cache: table, loader, rebuild, upkeep, rebuild-equals-upkeep test
  S9-07  Person header composer (name) + list read + Swift store / keys / name formatting
  S9-07b Rename configuration: Source fields → Metadata, Subject fields → Properties (app, Go, FFI, tables)
  ✎ S9-D1 ──▶ S9-08  Sidebar: Source / Conclude sections (Events / Places stubbed); Configure bottom-aligned
  ✎ S9-D2 ──▶ S9-09  Persons list (name + ref) + name on the card's membership row
  Check: promoted Persons listed by name; edit a name Observation → the row updates.

SLICE 3 — Join an existing Person
  S9-10  Promote write + reads: existing target, target suggestions
  S9-34a Handle search from cached values + kinds filter (pulled forward from Slice 10 for the picker)
  ✎ S9-D9  ──▶ S9-11  Promote shell + choose target (card Promote now opens it)
  ✎ S9-D10 ──▶ S9-12  Claim fields + save
  Check: two records on one Person → one list row; the second card shows the same PER-….

SLICE 4 — Reconciler + Person detail
  S9-13a Land migrations 000037 / 000038: retire `initial`; add `against` (from the closed #255 / #256)
  S9-13  Reconciler pipeline + text / integer / term modules + reasoning (pure Go, table-driven)
  S9-13b Name module (parts only, format-agnostic)
  S9-14  Evidence + reasoning in the cache: provenance, polarity, Sources; candidates table; credibility / certainty upkeep
  S9-15  Detail composer + detail read + value-state formatting (with reasoning)
  ✎ S9-D5 ──▶ S9-16  Person detail (name with states, values and why)
  Check: J. Robins + James Robins → one name, merged; James / Jim → mixed; a low-trust
         Source's spelling drops and the page says why; two Observations from one Source are one vote.

SLICE 5 — Compare, pins, backfill
  S9-17  Comparison read (module same-value) + pins + backfill in the Promote write
  S9-18  Pinned-Observation delete end to end (composer confirm names the evidence it leaves)
  ✎ S9-D11 ──▶ S9-19  Promote compare
  Check: join with confirmed pairs → both claims pinned; delete a pinned Observation → allowed, confirm names the handle, pins audited away.

SLICE 6 — Events
  S9-20  Seed event_name
  S9-21  Date module + date windows (pure Go, table-driven)
  S9-22  Event composer (title precedence without subjects, date) + reads + title formatting
  ✎ S9-D3 ──▶ S9-23  Events list (sidebar Events goes live)
  ✎ S9-D6 ──▶ S9-24  Event detail
  Check: promoted event cards list with titles; MAY 1985 + 14 MAY 1985 merge; APR vs MAY → mixed.

SLICE 7 — Places: names and cardinality
  S9-36  Property cardinality: column, seeds (toponym multi-valued), pipeline, config-change upkeep
  ✎ S9-D14 ──▶ S9-37  Properties page: cardinality control
  S9-25  Place composer (names) + reads
  ✎ S9-D4 ──▶ S9-26  Places list (sidebar Places goes live)
  ✎ S9-D7 ──▶ S9-27  Place detail (names; hierarchy rows empty until S9-40)
  Check: Montréal + Montreal → two names on one Place; "york" + "York" → one; a low-trust spelling drops.

SLICE 8 — Walk + bridges
  S9-28  Edges: subject module, bridge filing in the Promote write, inbound-end upkeep
  S9-29  Neighborhood read
  ✎ S9-D12 ──▶ S9-30  Promote walk + off-ramp
  Check: walk a birth record → Participation + Location handles filed; bridges wait for both ends; Done keeps saved steps.

SLICE 9 — Place hierarchy (the hard way)
  S9-38  Place model: place relationship bridge, typed terms with categories, place period, connect rule, loop refusal
  ✎ S9-D15 ──▶ S9-38b Custom term dialog: category for new place relationship types
  S9-39  Place chain composer: parents at a date, candidates when undecided, parts, succession; header dependents
  S9-40  Place hierarchy in Places list and Place detail (designed in D4 / D7)
  Check: draw Toronto part of Upper Canada / Province of Canada / Ontario with periods → Toronto's page shows each by period;
         a cycle is refused; York succeeded by Toronto shows on both.

SLICE 10 — Derived values across the graph
  S9-31  Composer walks: life dates + places with chains at the event date, subject titles, event places; header dependents
  S9-32  Fill derived cells in Persons / Events lists and Person / Event detail (designed in D2 / D3 / D5 / D6)
  S9-33  Deep fixture + timings
  Check: "Birth of James Robins"; James shows 1817 – 1880 · York, Upper Canada → Toronto, Ontario; timings in the ledger.

SLICE 11 — Search
  S9-34  Search documents from headers + dependents (S9-34a shipped kinds, cached-value documents, upkeep)
  ✎ S9-D13 ──▶ S9-35  Omnibar Conclusion hits
  Check: PER-7KD45, Jim Robins, Birth of James, a place name → each finds its handle; a name edit updates the hit.

CLOSE
  S9-99  Dogfood close / docs
```

- **Slices run in order.** Within a slice, PRs run top to bottom; the Go PRs at the top of a slice can usually go side by side (S9-13b / S9-14; S9-20 / S9-21; S9-36 / S9-25).
- **Slice 4 is the foundation for the rest.** S9-13 / S9-14 put every value type on the pipeline; slices 5–7 can then swap or run side by side. Slice 9 needs slice 8 (bridges and edges).
- **Replanned 2026-10-05.** IDs of PRs that keep their purpose stay; new work takes new IDs (S9-13a, S9-13b, S9-36 – S9-40), so handoff notes in [`completed.md`](completed.md) and code comments stay right.
- **Migrations 000037 / 000038 are fixed.** They first shipped on the closed PRs #255 / #256, and the researcher's local projects already carry them. S9-13a lands them on `main` byte-for-byte so those projects open again; nothing else may take those numbers, and later changes are new migrations (000039 on), never edits. **000039** is S9-13's `reason` column; S9-14's candidates table is **000040**. **Cache versions start at 5** after S9-13a: projects may hold a cache stamped 3 or 4 by the closed PRs, and a new meaning must never reuse a stamp.
- **The cache is honest from slice 2.** S9-06 ships the rebuild-equals-upkeep test; every later PR that adds a write path or trigger adds to it.
- **Churn is expected.** A confirm-and-mint Promote button (slice 1), stubbed sidebar items, and empty life-date cells are fine between slices.

### Dependencies at a glance

| PR | Brief | Depends on |
| --- | --- | --- |
| S9-01 Conclusion schema + stores | — | — |
| S9-02 Promote write v1 | — | S9-01 |
| S9-03 Source graph carries membership | — | S9-01 |
| S9-04 Graph card: Promote + membership | **S9-D8** | S9-02, S9-03 |
| S9-05 Resolver core v1 | — | — |
| S9-06 Resolved-values cache | — | S9-01, S9-05 |
| S9-07 Person header + list read | — | S9-06 |
| S9-07b Rename configuration views | — | — |
| S9-08 Sidebar | **S9-D1** | S9-07, S9-07b |
| S9-09 Persons list | **S9-D2** | S9-07, S9-08 |
| S9-10 Promote write + reads: existing target | — | S9-06 |
| S9-34a Handle search from cached values | — | S9-06 |
| S9-11 Promote shell + choose target | **S9-D9** | S9-04, S9-10, S9-34a |
| S9-12 Claim fields + save | **S9-D10** | S9-11 |
| S9-13a Land migrations 000037 / 000038 | — | — |
| S9-13 Reconciler pipeline + simple modules | — | S9-05 |
| S9-13b Name module | — | S9-13, S9-13a |
| S9-14 Evidence + reasoning in the cache | — | S9-06, S9-13, S9-13a |
| S9-15 Detail composer + read | — | S9-07, S9-14 |
| S9-16 Person detail | **S9-D5** | S9-13b, S9-15 |
| S9-17 Compare read + pins + backfill | — | S9-12, S9-13b |
| S9-18 Pinned-Observation delete end to end | — | S9-17 |
| S9-19 Promote compare | **S9-D11** | S9-17 |
| S9-20 Seed `event_name` | — | — |
| S9-21 Date module | — | S9-13 |
| S9-22 Event composer + reads | — | S9-15, S9-20, S9-21 |
| S9-23 Events list | **S9-D3** (extends D2) | S9-09, S9-22 |
| S9-24 Event detail | **S9-D6** (extends D5) | S9-16, S9-22 |
| S9-36 Property cardinality | — | S9-14 |
| S9-37 Properties page: cardinality | **S9-D14** | S9-36 |
| S9-25 Place composer + reads | — | S9-15, S9-36 |
| S9-26 Places list | **S9-D4** (extends D2) | S9-09, S9-25 |
| S9-27 Place detail | **S9-D7** (extends D5) | S9-16, S9-25 |
| S9-28 Edges + bridge filing | — | S9-12, S9-13 |
| S9-29 Neighborhood read | — | S9-28 |
| S9-30 Promote walk | **S9-D12** (extends D9) | S9-29 |
| S9-38 Place model | — | S9-28, S9-36 |
| S9-38b Custom term dialog: category | **S9-D15** | S9-38 |
| S9-39 Place chain composer | — | S9-38, S9-21, S9-25 |
| S9-40 Place hierarchy in list and detail | (D4 / D7) | S9-39, S9-26, S9-27 |
| S9-31 Composer walks | — | S9-22, S9-28, S9-39 |
| S9-32 Fill derived cells | (D2 / D3 / D5 / D6) | S9-31, S9-23, S9-24 |
| S9-33 Deep fixture + timings | — | S9-31 |
| S9-34 Search documents from headers | — | S9-31, S9-34a |
| S9-35 Omnibar hits | **S9-D13** | S9-34 |
| S9-99 Dogfood close | — | all |

---

## Checklist

In order; each brief sits just above the PR it gates.

- [x] S9-01 — Conclusion schema + stores → [`completed.md`](completed.md)
- [x] S9-02 — Promote write v1 → [`completed.md`](completed.md)
- [x] S9-03 — Source graph carries membership → [`completed.md`](completed.md)
- [x] ✎ S9-D8 — Design: graph subject card → [`completed.md`](completed.md)
- [x] S9-04 — Graph card: Promote + membership → [`completed.md`](completed.md)
- [x] S9-05 — Resolver core v1 → [`completed.md`](completed.md)
- [x] S9-06 — Resolved-values cache → [`completed.md`](completed.md)
- [x] S9-07 — Person header composer + list read → [`completed.md`](completed.md)
- [x] ✎ S9-D1 — Design: workspace sidebar → [`completed.md`](completed.md)
- [x] S9-07b — Rename configuration: Metadata, Properties → [`completed.md`](completed.md)
- [x] S9-08 — Sidebar sections: Source, Conclude, Configure → [`completed.md`](completed.md)
- [x] ✎ S9-D2 — Design: Persons list → [`completed.md`](completed.md)
- [x] S9-09 — Persons list → [`completed.md`](completed.md)
- [x] S9-10 — Promote write + reads: existing target → [`completed.md`](completed.md)
- [x] S9-34a — Handle search from cached values + kinds filter → [`completed.md`](completed.md)
- [x] ✎ S9-D9 — Design: Promote shell + choose target → [`completed.md`](completed.md)
- [x] S9-11 — Promote shell + choose target → [`completed.md`](completed.md)
- [x] ✎ S9-D10 — Design: Promote claim fields → [`completed.md`](completed.md)
- [x] S9-12 — Promote claim fields + save → [`completed.md`](completed.md)
- [x] S9-13a — Land migrations 000037 / 000038 → [`completed.md`](completed.md)
- [x] S9-13 — Reconciler pipeline + text / integer / term modules → [`completed.md`](completed.md)
- [x] S9-13b — Name module → [`completed.md`](completed.md)
- [ ] S9-14 — Evidence + reasoning in the cache
- [ ] S9-15 — Detail composer + detail read
- [ ] ✎ S9-D5 — Design: Person detail (revise for reasoning)
- [ ] S9-16 — Person detail
- [ ] S9-17 — Compare read + pins + backfill
- [ ] S9-18 — Pinned-Observation delete end to end
- [ ] ✎ S9-D11 — Design: Promote compare
- [ ] S9-19 — Promote compare
- [ ] S9-20 — Seed `event_name`
- [ ] S9-21 — Date module + windows
- [ ] S9-22 — Event composer + reads
- [ ] ✎ S9-D3 — Design: Events list
- [ ] S9-23 — Events list
- [ ] ✎ S9-D6 — Design: Event detail
- [ ] S9-24 — Event detail
- [ ] S9-36 — Property cardinality
- [ ] ✎ S9-D14 — Design: Properties cardinality
- [ ] S9-37 — Properties page: cardinality
- [ ] S9-25 — Place composer + reads
- [ ] ✎ S9-D4 — Design: Places list (revise for names and chain)
- [ ] S9-26 — Places list
- [ ] ✎ S9-D7 — Design: Place detail (revise for hierarchy)
- [ ] S9-27 — Place detail
- [ ] S9-28 — Edges + bridge filing
- [ ] S9-29 — Neighborhood read
- [ ] ✎ S9-D12 — Design: Promote walk
- [ ] S9-30 — Promote walk + off-ramp
- [ ] S9-38 — Place model: relationships, periods
- [ ] ✎ S9-D15 — Design: custom term category
- [ ] S9-38b — Custom term dialog: category
- [ ] S9-39 — Place chain composer
- [ ] S9-40 — Place hierarchy in list and detail
- [ ] S9-31 — Composer walks + header dependents
- [ ] S9-32 — Fill derived cells in lists and details
- [ ] S9-33 — Deep fixture + timings
- [ ] S9-34 — Search documents from headers + dependents
- [ ] ✎ S9-D13 — Design: omnibar hits
- [ ] S9-35 — Omnibar Conclusion hits
- [ ] S9-99 — Dogfood close / docs

---

## PRs

### Slice 1 — Mint from the graph

#### S9-01 — Conclusion schema + stores

**Done.** See [`completed.md`](completed.md#s9-01--conclusion-schema--stores).

| | |
| --- | --- |
| **In** | Migration: `claim_confidence_grades` (+ seed §5.5), `canonical_entities`, `identity_claims` (composite FKs, `UNIQUE (subject_id, entity_id)`, one-accepted partial index), `identity_claim_evidence`. Canonical ref minting from `subject_types.ref_prefix`. Go stores: create entity, create claim, get / list by kind, membership lookups both ways. Audit on every write. |
| **Out** | FFI (S9-02); cache (S9-06); Reconciliation tables. |
| **Testable** | Type mismatch rejected by FK; second accepted claim rejected; ref format; audit rows. |
| **Depends on** | — |

#### S9-02 — Promote write v1

**Done.** See [`completed.md`](completed.md#s9-02--promote-write-v1). v1 refuses non-primary kinds (`promote.unsupported_type`); **S9-28** relaxes that when bridge filing lands.

| | |
| --- | --- |
| **In** | FFI + Go: mint a handle of the subject's type and write an accepted Identity Claim, one transaction. Refuses already-member subjects. Swift `GenealogyStore` + FakeStore. Subject delete Impact names the handle the subject would leave (claim CASCADEs). |
| **Out** | Existing targets (S9-10); pins (S9-17); cache (S9-06). |
| **Testable** | Mint writes both rows; second promote of the same subject refused; Subject delete Impact lists the handle. |
| **Depends on** | S9-01 |

#### S9-03 — Source graph carries membership

**Done.** See [`completed.md`](completed.md#s9-03--source-graph-carries-membership). Nothing dispatches `.promotedSubject(sourceId:)` yet — **S9-04** must, after `promoteSubject` succeeds.

| | |
| --- | --- |
| **In** | Source-graph read includes each subject's accepted handle (id, ref, kind) or none. Swift model field. Promote mutation invalidates that Source's graph. |
| **Testable** | Promoted / unpromoted subjects reported correctly. |
| **Depends on** | S9-01 |

#### S9-04 — Graph card: Promote + membership

**Done.** See [`completed.md`](completed.md#s9-04--graph-card-promote--membership). Brief archived: [`design/archive/S9-D8-graph-subject-card.md`](design/archive/S9-D8-graph-subject-card.md). The membership row's `.openHandle` target is live but routes nowhere until a handle page exists (S9-16 / S9-24 / S9-27).

| | |
| --- | --- |
| **In** | Per **S9-D8**: Promote control on unpromoted person / event / place cards (`SourceGraphPlacedSubject.membership == nil`). **v1:** a Confirm (*Create a new Person from CPR-…?*) then mint via S9-02 and apply `.promotedSubject(sourceId:)`. Membership row shows the handle ref from `membership.entity.ref` (name arrives in S9-09). Canvas action targets + VoiceOver actions. |
| **Out** | The Promote flow (S9-11 reroutes the button to it). |
| **Check** | Promote a card; it shows PER-… and no longer offers Promote. |
| **Depends on** | **S9-D8**, S9-02, S9-03 |

### Slice 2 — Persons list

#### S9-05 — Resolver core v1

**Done.** See [`completed.md`](completed.md#s9-05--resolver-core-v1). `core/resolve.Resolve(valueType, candidates, concluded)` → `Result{Clusters, Concluded}`; state is `Result.State()`. After the replan, S9-13 replaces the per-type cluster key (`key` in `resolve.go`) with the reconciler pipeline and its modules; S9-14 loads the evidence the pipeline weighs.

| | |
| --- | --- |
| **In** | Pure Go. Candidates in → ranked clusters out with states (single / merged / mixed). v1 clustering: exact-equal values (names by normalized `form`, text, terms); stable order by support then Observation id. Optional concluded input. |
| **Out** | Name / date reconcilers (S9-13, S9-21); provenance ranking (S9-14); subject-valued mapping (S9-28). |
| **Testable** | Clustering, states, stable order, concluded input wins. |
| **Depends on** | — |

#### S9-06 — Resolved-values cache

**Done.** See [`completed.md`](completed.md#s9-06--resolved-values-cache). Package `core/database/resolvedvalues`: `RecomputeTx` / `RecomputeSubjectsTx` are the upkeep calls a new write path adds (and an operation in `TestRebuildEqualsUpkeep_SeededSequences`, plus a scenario in `TestRebuildEqualsUpkeep_Scenarios` where it has a characteristic sequence); a resolution change bumps `CacheVersion`. Still NULL: `date_lo` / `date_hi` and date `sort_key` (**S9-21**), `value_entity_id` — subject-valued Properties are not cached yet (**S9-28**). The loader reads no provenance until **S9-14**. Dates and names are protobuf via `core/valuecodec`.

| | |
| --- | --- |
| **In** | Migration: `conclusion_resolved_values` (R3 shape, all columns and indexes up front) + version row. Batched loader (members, Observations, provenance in a fixed number of queries per batch). Full rebuild; version check on Open. Upkeep in the write transaction for the write paths that exist now: Promote (claim create), Observation save / delete, Subject delete (affected handles from `deleteimpact.ReleaseFacets` → `Released.Handles`, computed before anything is gone; a member Observation's save / delete recomputes its Subject's handle directly — `Released.Handles` only covers membership and evidence changes). **Rebuild-equals-upkeep test** (fixed-seed sequences + named scenarios). |
| **Out** | Credibility / certainty triggers (S9-14); inbound-end trigger (S9-28). |
| **Testable** | Rebuild matches hand-computed rows; version mismatch rebuilds; rebuild-equals-upkeep equivalence; loader query count constant. |
| **Depends on** | S9-01, S9-05 |

#### S9-07 — Person header composer + list read

**Done.** See [`completed.md`](completed.md#s9-07--person-header-composer--list-read). For **S9-08**: `WorkspaceSection.persons` and its place exist but the sidebar doesn't list it; `CatalogCounts.persons` is filled by `refreshAll()` (nav counts carry `persons`) — re-publishing it after a Promote is S9-08's job. For **S9-09**: the place's `queryKeys` is empty and `PersonsListView` is an EmptyState; add `.personsList` to the place and render `CatalogPersonHeader` rows titled by `PersonHeaderDisplay.title`.

| | |
| --- | --- |
| **In** | Go Person header (resolved name, fallback label → ref) from the cache; set-based list composition; list + count FFI; Swift store, FakeStore, `CatalogQueryKey` + `PlaceRegistry` for Persons (stub view); invalidation (bust all Conclusion keys; evict details). Name formatting from NameValue structures. |
| **Testable** | Header fallbacks; list query count constant; a trigger mutation stales the key. |
| **Depends on** | S9-06 |

#### S9-07b — Rename configuration: Metadata, Properties

**Done.** See [`completed.md`](completed.md#s9-07b--rename-configuration-metadata-properties). Section ids are now `metadata` / `properties`; parse any section id with `WorkspaceSection(id:)`, which still maps `files`, `source-fields` and `subject-fields`. S9-08 builds the sidebar on `.metadata` / `.properties`.

Researcher's decision while revising **S9-D1**: two configuration views get plain names, renamed **all the way down** so code, wire, and tables say what the UI says. **Source types** keeps its name (a table or page called just "types" would not say what it holds).

| Today | Becomes |
| --- | --- |
| Source fields | **Metadata** (one row: a *metadata field*) |
| Subject fields | **Properties** (Properties and their bindings to Subject types) |

| | |
| --- | --- |
| **In** | **App:** labels, page titles, VoiceOver, L10n keys and values; `WorkspaceSection` cases and history ids (`metadata`, `properties`; the old `source-fields` / `subject-fields` ids still decode, like the retired `files` id); `Features/SourceFields` → `Features/Metadata`, `Features/SubjectFields` → `Features/Properties`, and their types (`SourceFieldsView` → `MetadataView`, `SubjectFieldsModel` → `PropertiesModel`, …); FakeStore and tests. **Go / FFI:** `sourcefields` package → `metadatafields`, `SourceField` types, delete-Impact kinds, FFI methods and proto messages (`SourceField` → `MetadataField`), search kinds (with a `ProjectionVersion` bump). **Tables:** a migration renaming the subject-type binding table `subject_type_fields` → `subject_type_properties` (and any other table whose name still says *source field* / *subject field*; `source_metadata_fields` already reads right), with the delete-Impact register and its honesty tests following. **Docs and skills** that name these views. |
| **Decided** | Rename everything, the table included. Migration `000035` renames `subject_type_fields` → `subject_type_properties` (and its indexes) and rewrites the two stored audit strings (`source_field` → `metadata_field`, `delete_source_field` → `delete_metadata_field`); bindings were never audited. `apperr` codes `sourcefields.*` → `metadatafields.*` outright (never stored; app and engine ship together). Search / Impact kind `source_field` → `metadata_field`, with search `ProjectionVersion` 5 rebuilding stored docs. Old section ids decode everywhere a section id is parsed (history, engine locations, sidebar). |
| **Out** | The sidebar layout (S9-08). Renaming Source types. Any change to what the pages do. |
| **Testable** | Old navigation history decodes to the renamed sections; an existing catalog migrates and still opens (schema hash, Impact honesty); search finds metadata fields and Properties after the rebuild; no remaining `SourceField` / `SubjectField` identifiers outside migrations and legacy decoders (a grep check in the PR). |
| **Depends on** | — (independent; lands before S9-08 so the sidebar is built on the new names) |

#### S9-08 — Sidebar sections: Source, Conclude, Configure

**Done.** See [`completed.md`](completed.md#s9-08--sidebar-sections-source-conclude-configure). Brief archived: [`design/archive/S9-D1-sidebar.md`](design/archive/S9-D1-sidebar.md). The sidebar is built from `WorkspaceSidebarSections`; Narrate's slot is marked there. For **S9-09 / S9-23 / S9-26**: the `persons` / `events` / `places` places exist with `queryKeys: []` and `ConclusionStubView`, and their counts come from nav counts; give each place its list key and replace the stub.

| | |
| --- | --- |
| **In** | Per **S9-D1**: titled sections. **Source** (Sources) and **Conclude** (Persons live; Events / Places present but stubbed until S9-23 / S9-26) top-aligned, with counts. **Configure** — **Source types**, **Metadata**, **Properties** (names from S9-07b) — bottom-aligned directly above the session footer, separated by space; on a short window the column scrolls as one (W-5b). Rail icons. **Narrate** is reserved after Conclude but hidden (Narrative layer, later spike). Sections don't collapse; the configuration destinations stop being children of Sources and become Configure's own rows. |
| **Check** | Persons count matches promoted Persons; configuration sits above the footer on a tall window and follows the research section on a short one. |
| **Depends on** | **S9-D1**, S9-07, S9-07b |

#### S9-09 — Persons list

**Done.** See [`completed.md`](completed.md#s9-09--persons-list). Brief archived: [`design/archive/S9-D2-persons-list.md`](design/archive/S9-D2-persons-list.md). For **S9-23 / S9-26**: build on kit `PVList` and `ConclusionListRow` (swap the mark; no mixed marker). For **S9-32**: the secondary line slot is empty — fill it with the b. / d. event groups. For **S9-16**: the Person-detail place exists as a stub (`WorkspaceLocation.entityId`, `PlaceID.personDetail`); rows and the card's membership row already open it.

| | |
| --- | --- |
| **In** | Per **S9-D2**: rows with thumbnail placeholder, name, ref; empty state. Life-date and place cells render empty until S9-32. The graph card's membership row shows the resolved name. The graph card's membership row swaps *Open person page* for the handle's resolved name — same slot, no relayout (S9-D8). |
| **Check** | Promoted Persons listed by name; edit a name Observation on a member → row updates. |
| **Depends on** | **S9-D2**, S9-07, S9-08 |

### Slice 3 — Join an existing Person

#### S9-10 — Promote write + reads: existing target

**Done.** See [`completed.md`](completed.md#s9-10--promote-write--reads-existing-target). Matching is its own module: [`docs/matching.md`](../../matching.md).
- **S9-11:** `listPromoteTargetSuggestions` gives the *Suggested* rows: `CatalogPromoteTargetSuggestion` (score, reasons, `person` header), best first. Search the existing `.personsList` for the picker, and join by passing `entityID` to `promoteSubject`.
- **S9-12:** grades come from `listClaimConfidenceGrades`; pass `confidenceGradeID` / `argument` on the same `promoteSubject` call.
- **S9-22 / S9-25:** add an `event` / `place` header to `PromoteTargetSuggestion`; matching for those kinds already ships.
- **S9-28 / S9-29:** new signals (life dates through edges, related-first ordering) are Features in `core/match` profiles.

| | |
| --- | --- |
| **In** | Promote write accepts an existing handle of the same type (claim only; cache upkeep). Target-suggestion read: same-type handles as headers, resemblance via cache `sort_key`. Confidence grade + argument on the claim. |
| **Out** | Pins / backfill (S9-17); related-first suggestions during a walk (S9-29). |
| **Testable** | Join writes a claim onto the existing handle; cache recomputed; suggestions filtered by type. |
| **Depends on** | S9-06 |

#### S9-34a — Handle search from cached values + kinds filter

**Done.** See [`completed.md`](completed.md#s9-34a--handle-search-from-cached-values--kinds-filter). This was pulled forward from Slice 10 so the Promote picker (S9-11) can search Persons, Events and Places through catalog search. S9-34 as written couldn't go first: S9-34 → S9-31 → S9-28 → S9-12 → S9-11.
- **S9-11:** search with `searchCatalog(…, kinds: [kind])`; hits carry `memberCount` and a location with `entityId`.
- **S9-34:** replace the cached-value documents with header-built ones and reproject header dependents; kinds, upkeep, the filter and member counts already exist.
- **S9-35:** add the kinds to the omnibar's default set (`DefaultInEverything`) once their rows are designed.

| | |
| --- | --- |
| **In** | Search index kinds `person` / `event` / `place`. Documents come from `conclusion_resolved_values` (rank-1 name, toponym or event type as the title; other values as secondary; label, then ref, as fallbacks). They are reprojected by `resolvedvalues.RecomputeTx` in the same transaction. Merged handles drop out. `ProjectionVersion` 6. `SearchCatalogRequest.kinds` filter; the omnibar default excludes handle kinds. `SearchHit.member_count`; `WorkspaceLocation.entity_id`. |
| **Out** | Header-built documents and dependents (S9-34); omnibar rows (S9-35). |
| **Testable** | Kinds filter; omnibar default excludes; a name edit, join or merge updates the index; rebuild equals upkeep for handle documents. |
| **Depends on** | S9-06 |

#### S9-11 — Promote shell + choose target

**Done.** See [`completed.md`](completed.md#s9-11--promote-shell--choose-target). Brief archived: [`design/archive/S9-D9-promote-target.md`](design/archive/S9-D9-promote-target.md). Promote is a workspace place (`SourceSurface.promote`, `PromoteView` / `PromoteModel` in `Features/Promote/`).
- **The flow is a state machine:** `PromoteFlow` (`Features/Promote/PromoteFlow.swift`) is the single source of truth for steps, the draft, writes in flight, the leave guard and the walk. `PromoteModel` sends it events and runs the effects it returns. A new step is a case on `PromoteStep`, its screen, and its draft fields; it isn't new plumbing.
- **S9-12:** done; see [S9-12](#s9-12--promote-claim-fields--save).
- **S9-19:** add `.compare` to `PromoteStep.built` and its screen; confirmed pairs go on the draft. Restore the board's hint ("James Robins will be compared with the 2 members of PER-…").
- **S9-29 / S9-30:** the walk is `PromoteFlow.enqueue` (a queue of subjects; after each save the flow moves to the next subject and `savedCount` counts them). Add the "Related to {ref}" group (frame 06), the "N saved" badge and the walk sentence in the leave guard.
- **S9-32:** life years and places in the candidate rows (`PromoteCandidateRow`) and the search rows (`PromoteSearchRow`), both in `Features/Promote`.

| | |
| --- | --- |
| **In** | Per **S9-D9**: the Promote place (or sheet) with step navigation and leave guard; choose target (new vs existing, suggestions). The card's Promote now opens it (replaces the S9-04 confirm). |
| **Depends on** | **S9-D9**, S9-04, S9-10, S9-34a |

#### S9-12 — Promote claim fields + save

**Done.** See [`completed.md`](completed.md#s9-12--promote-claim-fields--save). Brief archived: [`design/archive/S9-D10-promote-claim-fields.md`](design/archive/S9-D10-promote-claim-fields.md). The claim step is `PromoteClaimStep`; the footer reads only `PromoteFlow.controls`.
- **S9-19:**
  - add `.compare` to `PromoteStep.built`, plus its screen and its `Edit` cases; Back, Next and the step count follow on their own.
  - Draft the join argument from confirmed pairs, with the hint "Draft from your N confirmed matches · edit before saving". A changed target must clear what was drafted from it (the hook is named in `PromoteFlow`'s doc comment).
  - Restore the summary line's Compare clause and the board's choose-target hint ("James Robins will be compared with the 2 members of PER-…").
- **S9-17:** the summary's pins badge counts the pins on the claim (it reads "No pins" today).
- **S9-30:** after a save, the walk moves to the next subject (`PromoteFlow` already does). Add the toast body "{next} is next in this walk", "earlier steps stay saved" in the refusal callout, and the "N saved" badge.
- **Later spike:** Provisional and Rejected are `PromoteFlow.ClaimStatus` cases, plus engine support for writing them.

| | |
| --- | --- |
| **In** | Per **S9-D10**: Status dropdown (one option), confidence, argument; Save writes the step; Done exits. |
| **Check** | Promote a second census person onto an existing Person → one list row, both cards show the same PER-…. |
| **Depends on** | **S9-D10**, S9-11 |

### Slice 4 — Reconciler + Person detail

Design: [`conclusion-reconciliation.md`](../../conclusion-reconciliation.md). PRs #255 (name reconciler) and #256 (name confidence + polarity) were built to the first plan and are **closed**. Their pieces are lifted on purpose: the `initial` retirement and both migrations into S9-13a; provenance, deny and confidence logic and its cases into S9-13; the name logic, `namevaluestest` fixtures and the 55 name cases into S9-13b; the loader query, upkeep hooks, second-Source fixtures and seeded steps into S9-14.

#### S9-13a — Land migrations 000037 / 000038

**Done.** See [`completed.md`](completed.md#s9-13a--land-migrations-000037--000038). `against` exists but is always 0 until S9-14 writes it. The next migration is **000039**; the next cache version is **5**.

| | |
| --- | --- |
| **In** | Cherry-pick #255's retire-`initial` commit as-is: migration **000037** (`UPDATE name_value_parts SET type = 'given' WHERE type = 'initial';`), `PartTypeInitial` out of the Go registry and the Western name pattern, the Swift enum, L10n key and string-catalog entry, tests and seeded-vocabulary lists. Migration **000038** byte-for-byte from #256 (`ALTER TABLE conclusion_resolved_values ADD COLUMN against INTEGER NOT NULL DEFAULT 0;`); nothing writes it until S9-14. No cache version change (the reconciler hasn't changed). |
| **Why first** | The researcher's projects already ran both migrations from the closed PRs, so `main` refuses to open them (a newer `user_version`). Landing them unchanged makes those projects open, and the open-time schema check passes because the schema text is identical. |
| **Testable** | `go test` for core (schema hash, Delete Impact honesty), macOS unit tests; an existing project from the closed-PR builds opens. |
| **Depends on** | — |

#### S9-13 — Reconciler pipeline + text / integer / term modules

**Done.** See [`completed.md`](completed.md#s9-13--reconciler-pipeline--text--integer--term-modules). For **S9-13b**: implement `module` (`core/resolve/modules.go`) for names: `split` into one unit per part type, `fold` for subsumption, `assemble` from the settled units; swap it into `moduleFor`. For **S9-14**: fill `Candidate.SourceID`, `Provenance`, `Negative` and `Provisional` in the loader; the pipeline already uses them. Store `Result.Candidates` in migration **000040**.

| | |
| --- | --- |
| **In** | Pure Go in `core/resolve`. The shared pipeline (design §5): admit → deny → group (same value, fold) → majority in distinct Sources → confidence → merge; per-candidate outcome and reason (design §6); `against`. A module interface (*same value*, *fold*, *merge*, *no evidence*) that also lets a module **split a candidate into comparable units and reassemble survivors** (names run the passes per part type, then rebuild one name; the simple modules use one unit per candidate). Design the split in now so S9-13b doesn't reshape the interface. The simple modules: **text** (trimmed, case-insensitive; changes S9-05's case-sensitive rule), **integer**, **term** (neutral terms are no evidence). Candidate inputs carry provenance, claim status, polarity and Source id. Cardinality is an input (single only until S9-36). Concluded input kept. Table-driven: the shared passes once, each module's cases separately. **Nothing is dropped** (decided while planning): every value keeps a row with its `reason`; migration **000039** adds the column; state and *+N* read displayed rows. |
| **Out** | Name, date and subject modules (S9-13b, S9-21, S9-28); loading evidence from the catalog (S9-14). |
| **Testable** | Every pass and reason; majority counts Sources not Observations; provisional always eliminated; a stronger negative denies, an equal one only counts against; order independence; non-name results for unchanged inputs match S9-05 except case folding. |
| **Depends on** | S9-05 |

#### S9-13b — Name module

**Done.** See [`completed.md`](completed.md#s9-13b--name-module). `nameModule` in `core/resolve/names.go`: one name per Person (`oneValue`), majority outvotes spelling variants only (`outvotes`, `resolve.SpellingSimilarity`); `resolve.IsInitial`; `namevaluestest.Western` builds parts for fixtures. For **S9-17**: a pair is compatible when `nameModule` (through `moduleFor(properties.ValueTypeName)`) gives both the same units or one folds into the other, the same test the Person page uses.

| | |
| --- | --- |
| **In** | The name module on the pipeline (design §7.2): structured parts only (`form` never read; no parts = `no_evidence`); format-agnostic, part types as identifiers; split into one unit per part type, subsumption (`[J]` → `[James]`, `[James]` → `[James, Kenneth]`); **one name per Person, never mixed**: each type keeps every surviving value; majority outvotes **spelling variants only** (decided while building it; a nickname recorded as a given name is never thrown out). Test fixtures that wrote form-only names build parts (`namevaluestest`). Cache version ≥ 5. (`initial` is already retired by S9-13a.) |
| **Testable** | 50+ table cases (the spec), seeded invariants (input order, forms never matter, renaming types renames nothing else). |
| **Depends on** | S9-13, S9-13a |

#### S9-14 — Evidence + reasoning in the cache

| | |
| --- | --- |
| **In** | Loader reads each candidate's polarity, Source, credibility, certainty, claim confidence and claim status in the existing candidate query (accepted and provisional members; rejected excluded); query count unchanged. `against` (column from S9-13a's migration 000038) is written. New migration **000040**: `conclusion_resolved_candidates` (entity, Property, Observation, outcome, reason, value rank) rewritten with the handle. Upkeep on Source credibility and Citation certainty changes (`RecomputeSourceTx`, `RecomputeCitationTx`). Rebuild-equals-upkeep gains credibility, certainty, negative and graded-promote steps across two Sources, asserted right after each provenance edit. Cache version bump (≥ 5). **Swift:** nothing to wire (certainty rides `savedCitation`, credibility `mutatedSourceWorkspace`). |
| **Testable** | Reasons stored; removing either hook fails the sequences; loader query count constant. |
| **Depends on** | S9-06, S9-13, S9-13a |

#### S9-15 — Detail composer + detail read

| | |
| --- | --- |
| **In** | Go detail composer: fields with states, every value with support and `against`, and the reasoning (candidates with their Source, outcome and reason) from the cache. Detail FFI; Swift store + detail key (stub view); value-state formatting (single / merged / mixed / empty, *+N*). **Eviction:** `WorkspaceSession` learns the visible place and evicts non-visible Conclusion detail keys on a Conclusion trigger (the rule-2 exception, deferred from S9-07). |
| **Depends on** | S9-07, S9-14 |

#### S9-16 — Person detail

| | |
| --- | --- |
| **In** | Per **S9-D5** (revised for the reasoning): header with name, states, values and why; life-date and place rows render empty until S9-32. Replace the Person-detail stub place from S9-09 (`ConclusionStubView` on `PlaceID.personDetail`) and give it the detail key; list rows and the graph card's `.openHandle` already route there. |
| **Check** | *J. Robins* + *James Robins* → merged; *James* / *Jim* → one name *James Jim Robins*, merged; two of three *Robins* outvote a *Robbins*; lowering one Source to low trust drops its spelling and the page says *weak*. |
| **Depends on** | **S9-D5**, S9-13b, S9-15 |

### Slice 5 — Compare, pins, backfill

#### S9-17 — Compare read + pins + backfill

| | |
| --- | --- |
| **In** | Comparison read: incoming Observations × each member's per Property; a pair is compatible when the value-type module says *same value* or *fold* (`moduleFor` in `core/resolve/modules.go`; names compare unit by unit, S9-13b). Promote write takes confirmed pairs: pins on the new claim **and** backfill onto the member's claim, same transaction. |
| **Testable** | Pins on both claims; older `argument` untouched; compatible pairs flagged by the same test the Person page uses. |
| **Depends on** | S9-12, S9-13b |

#### S9-18 — Pinned-Observation delete end to end

| | |
| --- | --- |
| **Already shipped** | S9-02 made pins a non-blocking, **Named** facet release: migration `000033` (`observation_id … ON DELETE CASCADE` as backstop), audited removal through `deleteimpact.ReleaseFacets`, and the confirm sentence naming the handles. |
| **In** | With real pins from S9-17: composer row delete and Subject delete through the shipped confirm; FakeStore models pins so Swift tests cover it; drop the Swift `sameness_claim` leftovers. Confirm the audit trail reads back per claim. |
| **Check** | Delete a pinned Observation from the composer → allowed; confirm names PER-…; both claims keep their other pins; audit shows the removed pins. |
| **Depends on** | S9-17 |

#### S9-19 — Promote compare

| | |
| --- | --- |
| **In** | Per **S9-D11**: compare step for existing handles with members, between target and claim fields. |
| **Check** | Join with two confirmed pairs → both claims pinned. |
| **Depends on** | **S9-D11**, S9-17 |

### Slice 6 — Events

#### S9-20 — Seed `event_name`

| | |
| --- | --- |
| **In** | `event_name` (`text`) bound to `event`: Install seed + migration. |
| **Depends on** | — |

#### S9-21 — Date module + windows

| | |
| --- | --- |
| **In** | The date module on the pipeline (design §7.1), table-driven (~50+ cases): containment folds, overlapping windows are the same value, range widening vs shared precision (a test-case decision), qualifiers, disjoint → mixed. `date_lo` / `date_hi` and the date `sort_key` computed and stored. Cache version bump. Period dates on Places (S9-38) use the same windows. |
| **Depends on** | S9-13 |

#### S9-22 — Event composer + reads

| | |
| --- | --- |
| **In** | Event header + detail: title precedence without subjects (`event_name` → label → *Unspecified {type}* → ref), date (else span). List / detail / count FFI and keys. Swift event-title formatter (L10n templates, full matrix so S9-31 only supplies parts). |
| **Depends on** | S9-15, S9-20, S9-21 |
| **Note** | Promote names an Event subject by `SourceGraphPlacedSubject.displayName` (S9-12), which uses the label until now. Point its Event case at this formatter, fed from the subject's own Observations on the graph (type term, `event_name`, date), so Promote's header reads like the Event's title. |

#### S9-23 — Events list

| | |
| --- | --- |
| **In** | Per **S9-D3**; sidebar Events goes live. Place cell empty until S9-32. |
| **Check** | Promoted event cards list with titles; `MAY 1985` + `14 MAY 1985` merge; `APR` vs `MAY` → mixed. |
| **Depends on** | **S9-D3**, S9-09, S9-22 |

#### S9-24 — Event detail

| | |
| --- | --- |
| **In** | Per **S9-D6**. Places empty until S9-32. Route the graph card's `.openHandle` target to this page for events. |
| **Depends on** | **S9-D6**, S9-16, S9-22 |

### Slice 7 — Places: names and cardinality

#### S9-36 — Property cardinality

| | |
| --- | --- |
| **In** | Migration: `properties.cardinality` (`single` / `multiple`, default `single`); `toponym` seeded `multiple` (Install + migration). The pipeline honors it (design §8): multi-valued keeps every distinct surviving value; majority never crowds out a distinct value; confidence still drops weak ones. Upkeep: a cardinality change recomputes every handle carrying the Property; joins rebuild-equals-upkeep. FFI on the Property read / update. |
| **Depends on** | S9-14 |

#### S9-37 — Properties page: cardinality

| | |
| --- | --- |
| **In** | Per **S9-D14**: a cardinality control in the Property inspector and create form (seeded Properties read-only). The mutation joins `conclusionTriggers` (via `updatedProperty`). |
| **Depends on** | **S9-D14**, S9-36 |

#### S9-25 — Place composer + reads

| | |
| --- | --- |
| **In** | Place header + detail: every reconciled name (multi-valued), with the reasoning. List / detail / count FFI and keys. Period, kind and hierarchy fields present but empty until S9-38 / S9-39. |
| **Depends on** | S9-15, S9-36 |

#### S9-26 — Places list

| | |
| --- | --- |
| **In** | Per **S9-D4** (revised); sidebar Places goes live. The chain cell is empty until S9-40. |
| **Depends on** | **S9-D4**, S9-09, S9-25 |

#### S9-27 — Place detail

| | |
| --- | --- |
| **In** | Per **S9-D7** (revised). Names with their reasoning; hierarchy, period and succession rows render empty until S9-40. Route the graph card's `.openHandle` target to this page for places. |
| **Check** | *Montréal* + *Montreal* → two names on one Place; *york* + *York* → one; a low-trust spelling drops with its reason. |
| **Depends on** | **S9-D7**, S9-16, S9-25 |

### Slice 8 — Walk + bridges

#### S9-28 — Edges + bridge filing

| | |
| --- | --- |
| **In** | The **subject module**: subject-valued Properties map to the target's handle (unpromoted drop out as `no_evidence`) — the canonical graph, in `value_entity_id`. Upkeep: a claim create / remove recomputes handles whose members' Observations point at that subject. Promote write files bridge subjects onto the association the ends already share, or mints one (lift the S9-02 primary-kinds guard in `promote.Save`). Extend the rebuild-equals-upkeep tests. Once bridge edge Observations can be pinned, test that deleting a bridge Subject releases those pins audited: its connection-facet release erases each edge Observation through `ReleaseFacets(KindObservation)`, which takes the pins first. |
| **Depends on** | S9-12, S9-13 |

#### S9-29 — Neighborhood read

| | |
| --- | --- |
| **In** | Unpromoted neighbors of a just-filed subject; queue rules; bridge readiness and shared association. Related-first target suggestions during a walk. |
| **Depends on** | S9-28 |

#### S9-30 — Promote walk + off-ramp

| | |
| --- | --- |
| **In** | Per **S9-D12**: connected subjects after each save; loop; bridge confirm; Done. |
| **Check** | Walk a birth record: person → event → place → participation → location; bridges wait for both ends; Done keeps saved steps. |
| **Depends on** | **S9-D12**, S9-29 |

### Slice 9 — Place hierarchy (the hard way)

Design: [`conclusion-reconciliation.md`](../../conclusion-reconciliation.md) §11.1 and R9.

#### S9-38 — Place model: relationships, periods

| | |
| --- | --- |
| **In** | Seed (Install + migration): subject type `place_relationship` (bridge) with Properties `from` (place), `to` (place), `place_relationship_type` (term); term **category** on `property_terms` (`hierarchical` / `temporal`, nullable for terms that don't use it); seeded types `administrative`, `geographic`, `ecclesiastical` (hierarchical) and `succeeded_by` (temporal); `start_date` / `end_date` bound to `place` (its period). A `connectrules` bridge place ↔ place disambiguated by `place_relationship_type`, so it is drawn and cited on the Evidence graph like any bridge; Promote files it through S9-28. **Loop refusal** in the filing write: a hierarchical link that would close a cycle among canonical places, or a looping succession, is refused with a clear error. Delete Impact and search registries for the new kind. |
| **Testable** | Seeds; connect rule; cycles refused (direct and transitive); a researcher-added type with a category behaves like the seeded one. |
| **Depends on** | S9-28, S9-36 |

#### S9-38b — Custom term dialog: category

| | |
| --- | --- |
| **In** | Per **S9-D15**: when a researcher adds a term to a Property whose terms carry a category (`place_relationship_type`), the composer's custom term dialog asks for it (hierarchical / temporal); FFI `createPropertyTerm` takes the category. |
| **Depends on** | **S9-D15**, S9-38 |

#### S9-39 — Place chain composer

| | |
| --- | --- |
| **In** | Go composer over the cache: a Place's hierarchical parents at a date (links hold where the two places' periods overlap; no period = always), following `administrative` first for display; every candidate when the date can't decide; its parts; its succession both ways. A chain reader for other composers (Persons, Events, search). Header dependents for places (a parent's rename reprojects its children's headers). Cycle guard. |
| **Testable** | Toronto in 1820 / 1850 / 1950; *about 1841* → both; several parents; succession never builds a chain; undated places always hold. |
| **Depends on** | S9-38, S9-21, S9-25 |

#### S9-40 — Place hierarchy in list and detail

| | |
| --- | --- |
| **In** | Places list chain cell (today's chain); Place detail period, parents by type with their periods, parts, succession — designed in **S9-D4 / S9-D7**; no new brief. |
| **Check** | Draw Toronto part of Upper Canada, Province of Canada and Ontario with periods → Toronto's page shows each by period; York succeeded by Toronto shows on both pages. |
| **Depends on** | S9-39, S9-26, S9-27 |

### Slice 10 — Derived values across the graph

#### S9-31 — Composer walks + header dependents

| | |
| --- | --- |
| **In** | Person birth / death date and place, the place with its **chain at the event's date** (S9-39); Event subject titles (*Birth of …*, marriage, *et al.*, *unnamed person*, *{Type} at {name}*) and places with chains. Header-dependents function (reverse walk) for later reprojection. |
| **Depends on** | S9-22, S9-28, S9-39 |
| **Note** | Give Promote's Event name (`displayName`) the same subject parts from the graph's participation and location bridges, so a Baptism card promotes as *Baptism of James Robins* rather than its label. People are named by their own `displayName` rule (name form, then label). |

#### S9-32 — Fill derived cells

| | |
| --- | --- |
| **In** | Persons / Events lists and Person / Event detail render the new cells — already designed in **S9-D2 / D3 / D5 / D6**; no new brief. |
| **Check** | *Birth of James Robins*; James shows *1817 – 1880 · York, Upper Canada → Toronto, Ontario*. |
| **Depends on** | S9-31, S9-23, S9-24 |

#### S9-33 — Deep fixture + timings

| | |
| --- | --- |
| **In** | Seeded project generator (a Person on ~10 Sources, multi-member events with Locations, relationships, a place hierarchy several levels deep; a few hundred handles). Benchmarks: rebuild, one-write upkeep, one Promote step, list and detail composition, chain composition. Ledger rows. |
| **Depends on** | S9-31 |

### Slice 11 — Search

#### S9-34 — Search documents from headers + dependents

S9-34a shipped the kinds, cached-value documents, upkeep through `RecomputeTx`, the kinds filter and member counts. What remains here:

| | |
| --- | --- |
| **In** | Documents from the header composers (match text only): "Birth of James Robins", a Person's life dates and places, a Place's names and today's chain. Reproject a handle's header dependents in the write transaction (deletes feed it `Released.Handles`). `SearchHit` carries the structured header. `ProjectionVersion` bump; FakeStore. Existing hit rows render a fallback until S9-35. |
| **Depends on** | S9-31, S9-34a |

#### S9-35 — Omnibar Conclusion hits

| | |
| --- | --- |
| **In** | Per **S9-D13**. |
| **Check** | `PER-7KD45`, *Jim Robins*, *Birth of James*, a place name → each finds its handle; a member name edit updates the hit. |
| **Depends on** | **S9-D13**, S9-34 |

### S9-99 — Dogfood close / docs

Honesty pass against the [goal bar](#goal-dogfood-bar); ledger timings recorded; docs in *Docs to update* folded in; briefs archived; spike archived via [`archive-docs`](../../../.cursor/skills/archive-docs/SKILL.md).

---

## Scope boundary

| In | Out |
| --- | --- |
| Canonical entities, Identity Claims, evidence pins | Reconciliation Claims (designed: one per value for multi-valued Properties), `name_format` |
| Names display as the reconciled name (a member's own `form` when one carries exactly its parts, else its parts in order); lists sort by that, normalized | Name display styles (natural / sorted), surname-first sort, profiles — decided in [`structured-name-model.md`](../../structured-name-model.md) §4.5 for a later spike |
| Reconciler: one pipeline, modules for every value type, reasoning cached; per-Property cardinality | Persisted "auto" claims; scores stored as catalog truth; values that change over time |
| Place hierarchy the hard way: place relationships (typed, categorized), periods, chains at a date, loop refusal | Gazetteer lookup ([`ideas/place-gazetteer-service.md`](../../ideas/place-gazetteer-service.md)); geometry; place reconciliation across grains |
| Resolved-values cache with upkeep, rebuild, and rebuild-equals-upkeep test | Per-screen caches; stored derived values (life dates, event names); resident in-memory graph (only if timings demand) |
| Header composers shared by lists, Promote, search | — |
| Promote from the Evidence graph: per-step saves, walk, Done off-ramp | Stub handles with no Subject; canonical merge |
| Promote creates claims (one-way) | Editing, re-pinning, or removing Identity Claims; removing a member (Spike 10) |
| Promote writes `accepted` claims; the reconciler already treats provisional members as reasoning-only and ignores rejected ones | `provisional` / `rejected` in Promote (later spike) |
| Person / Event / Place lists and details | Association-kind pages; tree / timeline / map |
| Thumbnail **slot** with placeholder | Likeness / depiction value type |
| Delete Impact naming Identity Claims (non-blocking cascades, audited removal) | Review queue / weak-claim alert for claims whose evidence or comparison member left (§5.2, Spike 10) |
| Omnibar search for Persons / Events / Places (ref + full text on resolved values) | Subjects as hits; cross-root association search |

---

## Gotchas

1. **Persons have no date Properties.** Birth and death dates come through Participation → Event (R4). Do not seed `birth_date` on person to shortcut it.
2. **`event_name`, not `name`.** Historical event names use the seeded `event_name` text Property (Q13). `name` is a personal NameValue — never bind it to `event`.
3. **The cache is not truth.** Nothing references `conclusion_resolved_values`; synthesized dates and merged names live only there, serialized. No claim, DateValue, or NameValue row is written by resolution ([`seeded-vocabulary.md`](../../seeded-vocabulary.md) §5.3).
4. **No vocabulary-named columns** in any derived table. If a screen needs a new concept, it is a composer change or a `property_id` row — never a column.
5. **Upkeep misses are silent.** A trigger the affected-handles function forgets leaves a stale row nobody notices. The rebuild-equals-upkeep test is the guard; every new write path adds its trigger and a test sequence.
6. **Cascades bypass Go — so Go doesn't rely on them.** Every official delete calls `deleteimpact.ReleaseFacets`, which removes and audits claims, pins, notes, and connection facets before the parent `DELETE` and fails if anything is left for the `CASCADE` backstop. Its `Released.Handles` is how cache upkeep (S9-06) — and with it handle search reprojection (S9-34a) — learns which handles a delete touched — computed before anything is gone. A new delete path gets this by calling `ReleaseFacets`, not by hand-wiring helpers.
7. **Promoting a subject can change other handles.** Its Observations' targets and inbound subject-valued Observations re-map; R3's claim trigger covers the second hop.
8. **One accepted claim per Subject** is a partial unique index. Promote must hide or refuse subjects that are already members, and a step that loses a race fails alone — earlier steps stay saved.
9. **Composite FKs** carry `subject_type_id` on the claim. A person Subject cannot be claimed onto a Place; the target picker filters by type so the researcher never sees that error.
10. **Backfill is symmetric.** A confirmed pair pins both Observations on the incoming claim **and** the existing member's claim. The older claim's `argument` is not rewritten.
11. **Pins are Observations only.** Not Citations, Subjects, or Sources. Confirming a match creates no Observation.
12. ~~**Stale reservation.**~~ Go side retired in S9-01. Swift still carries `sameness_claim` L10n / preview until S9-18.
15. **Pins never block.** A pinned Observation (or a Subject whose Observations are pinned on other members' claims) deletes freely; the claims stay with a weaker exhibit. Do not reintroduce a blocking evidence probe — weak claims are the §5.2 review alert's job.
13. **Candidate vs canonical refs.** Subjects are `CPR-…`; handles are `PER-…`. Both prefixes already exist on `subject_types`.
14. **Cross-Source reads** run on the serialized catalog session. Keep rebuild off the open path's critical section if it gets long.
16. **`form` is a transcription.** The reconciler never reads a NameValue's `form`; a name with no parts is no evidence. Fixtures and seeds that write names must write parts.
17. **Majority counts Sources.** Two Observations under one Source are one vote; a test that wants two votes needs two Sources.
18. **Places don't reconcile across grains.** Toronto, Ontario and Canada are three Places linked by relationships, never names of one Place. A rename is a new Place with `succeeded_by`.

---

## Open questions

| # | Question | Leaning |
| --- | --- | --- |
| Q1 | ~~Submit at the end vs commit each step.~~ | **Decided:** one transaction per promoted Subject, saved on Next, with a **Done** off-ramp after every step. Matches §5.4. |
| Q2 | ~~Correcting a mistake.~~ | **Descoped to Spike 10:** remove a member, remove its Identity Claim, add evidence to an existing claim. Promote stays create-only. |
| Q3 | ~~Walk canonical or Interpretation neighbors for derived values?~~ | **Decided:** canonical only. Detail pages show promoted handles and claims; Interpretation neighbors live on the Evidence graph. An unpromoted birth event gives the Person no birth date. |
| Q4 | ~~Which Participation roles count for a Person's birth / death?~~ | **Decided:** `role = subject` only — that is what the role is for. A father on a birth is not his birth. |
| Q5 | ~~Invalidation scope for Conclusion keys after Interpretation writes.~~ | **Decided:** bust all Swift Conclusion keys (evict details, revalidate the visible page). Cheap because R3 makes reloads cheap. |
| Q6 | ~~List description line.~~ | **Decided:** row fields per list in R5. |
| Q7 | ~~Provisional claims in Promote.~~ | **Decided:** Promote writes `accepted` only (optional confidence grade). The UI still ships the Status dropdown with that single option, so the design reserves its place. The `status` column and CHECK ship in R1. A later spike adds a status toggle to Promote along with its supporting work (candidate display, resolver exclusion, graph card state, re-promote rules). |
| Q8 | ~~Handles in the omnibar.~~ | **Decided — in scope (R8):** refs, full text on resolved values and clusters, hit rows per kind. |
| Q9 | ~~Event with two dates.~~ | **Decided:** no different from any single-valued Property: the date module reconciles them (R2). |
| Q10 | ~~Event naming matrix.~~ | **Decided (R4):** `subject` = principal(s), marriage is two subjects; *et al.* for several; *{Type} at {toponym}* / *Unspecified {type}* with no subject; *unnamed person*; generic templates for researcher-added types; precedence recorded name → composed → label → type/place → ref. |
| Q11 | ~~Details vs lists.~~ | **Decided:** a detail page is a superset of its list row. Person adds birth / death place; Event adds place(s) (R6). |
| Q12 | ~~Serialized value format in R3.~~ | **Decided:** protobuf DateValue / NameValue in `value_date` / `value_name`, plus `date_lo` / `date_hi` ordinals for range queries; never `date_values` / `name_values` rows. Go returns structures; Swift formats text. |
| Q13 | ~~Event recorded-name Property.~~ | **Decided:** seed `event_name` (`value_type = text`) and bind it to `event` this spike (R1). |
| Q14 | ~~Label vs composed-from-subjects.~~ | **Decided:** label stays below subject titles; it wins only when there is no subject. Revisit on dogfood. |
| Q15 | ~~Reconciling values other than names and dates.~~ | **Decided (replan):** every value type gets a module on one pipeline ([`conclusion-reconciliation.md`](../../conclusion-reconciliation.md)). Text is case-insensitive; majority counts Sources; reasoning is cached. |
| Q16 | ~~Places and their hierarchy.~~ | **Decided (replan):** separate Places linked by typed place relationships (hierarchical or temporal), periods on Places, chains at a date, evidence by hand this spike (R9). |
| Q17 | ~~Multi-valued Properties.~~ | **Decided (replan):** per-Property cardinality; `toponym` is the seeded multi-valued one. Most repeating facts are Events. |

---

## Docs to update as work lands

- [`conclusion-layer-data-model.md`](../../conclusion-layer-data-model.md): §5.3–§5.4 "future UI" → shipped behavior (per-step saves; grounding with zero pins; create-only). New section: the resolved-values cache as a derived, rebuildable projection; edges as resolved subject-valued Properties.
- [`research-judgment-model.md`](../../research-judgment-model.md) §1.1: cached order and reasoning are derived, not stored judgment.
- [`conclusion-reconciliation.md`](../../conclusion-reconciliation.md) §12: implementation status as each module lands.
- [`conclusion-layer-data-model.md`](../../conclusion-layer-data-model.md) §13: Place `contained_in` question answered by place relationships (R9); new `place_relationship` kind in the kinds list.
- [`seeded-vocabulary.md`](../../seeded-vocabulary.md): `place_relationship` kind and its Properties, `place_relationship_type` terms with categories, place `start_date` / `end_date`, Property cardinality (S9-36, S9-38).
- [`seeded-vocabulary.md`](../../seeded-vocabulary.md) §5.5: mark claim confidence grades as seeded. §3.5: `subject` = the event's principal(s), possibly several; marriage uses two `subject` Participations; `spouse` is a principal's spouse on another event. §3.2 / §3.3: `event_name` (text) bound to `event`.
- [`catalog-refs.md`](../../catalog-refs.md): canonical ref minting.
- [`catalog-deletes.md`](../../catalog-deletes.md): Identity Claim / evidence Impact (done in S9-02: non-blocking cascades, explicit audited release); hook cache dependents into the release calls (S9-06).
- [`macos-client-patterns.md`](../../macos-client-patterns.md): Conclusion query keys and their invalidation rule; Go returns structures (DateValue, NameValue, title parts), Swift formats text.
- [`omnibar-search.md`](../archive/spike-3/omnibar-search.md): new kinds, location mapping, composed documents and header-dependent reprojection.

---

## Definition of done

- [ ] Dogfood bar items 1–10 met on real research
- [ ] Checklist complete (or items explicitly descoped)
- [ ] Design briefs archived under `design/archive/`
- [ ] Open questions answered and folded into the model docs
- [ ] [`docs/deployment-plan/README.md`](../README.md) points at the archive when closed
