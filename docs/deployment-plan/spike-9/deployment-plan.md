# Deployment Plan — Spike 9

MVP for the **Conclusion layer**: assemble canonical Persons, Events, and Places from Interpretation Subjects through Identity Claims, and show them on list and detail pages. Authoritative model: [`conclusion-layer-data-model.md`](../../conclusion-layer-data-model.md). Values: [`structured-name-model.md`](../../structured-name-model.md), [`structured-date-model.md`](../../structured-date-model.md). Vocabulary: [`seeded-vocabulary.md`](../../seeded-vocabulary.md) §3, §5.

## Status

**Open.** Requirements, [design track](#design-track), and [PR sequence](#pr-sequence) drafted; briefs S9-D1…D13 written in [`design/`](design/). Landings go in [`completed.md`](completed.md).

> **Goal of this spike:** a researcher can promote Subjects off an Evidence graph into Persons, Events, and Places, and open a page for each that shows who or what it is — name and life dates, event and date, toponym — resolved from every member Subject.

> **Foundation, grown in vertical slices.** The resolved-values cache (R3) is core infrastructure every Conclusion surface reads — lists, details, Promote, search, and later the tree. Build it properly, but grow it slice by slice so each layer is exercised in the app as soon as it lands. Stubs and temporarily incomplete UI along the way are fine; no band-aid caches per screen.

## Goal (dogfood bar)

**By spike close**, each of these must be true in the app on real research:

1. **Promote a person from a birth record.** From a person card on the Evidence graph, Promote mints a new Person. The walk then offers the birth event, its participation, its place, and its location. Each can join an existing handle or mint a new one. Each step saves on Next; **Done** stops the walk with everything so far kept.
2. **Promote a second record onto the same Person.** A census person joins the existing Person. The comparison lines up name (and anything else comparable) against the current members. Confirmed pairs are pinned on both claims.
3. **Browse.** Persons, Events, and Places appear in the sidebar. Each list shows every handle of that kind with the row fields in R5: Persons with name, birth–death dates, birth and death places; Events with a derived name, date, and place; Places with a toponym. Every row has a thumbnail slot and its ref.
4. **Person detail** shows the thumbnail slot, the resolved name, the birth and death dates, and the birth and death places. Two members saying `14 MAY 1985` and `MAY 1985` show one date. Two members saying `MAY 1985` and `APR 1985` show that they disagree.
5. **Event detail** shows the thumbnail slot, the event title, the event date, and the event place(s).
6. **Place detail** shows the thumbnail slot and **every** toponym its members carry (no reconciliation).
7. **The graph knows.** A promoted subject's card shows the handle it belongs to and links to its page. Promote is not offered twice for a subject that is already a member.
8. **Search finds them.** Typing `PER-7KD45`, *James Robins*, *Jim Robins* (an alternate), *Birth of James*, or a toponym in the omnibar lists the handle with a row that reads like its list row; selecting it opens the detail page. Editing a member's name Observation updates the hit.
9. **The cache is honest.** A full rebuild of the resolved-values cache produces exactly what incremental upkeep produced, after any sequence of writes (R3 test). Timings on the deep fixture are recorded in the [performance ledger](performance-ledger.md).

---

## Architecture

```text
Interpretation (truth)          Conclusion (truth)
 subjects, observations,         canonical_entities,
 citations, sources, …           identity_claims (+ evidence)
            │                              │
            └──────── resolver (R2) ───────┘
                          │   written in the same transaction as the change (R3)
                          ▼
            conclusion_resolved_values   ← the one derived cache; rebuildable
              scalar Properties  (name, date, toponym, event_type, role, …)
              entity-valued ends (Participation.person → PER-1, …) = the canonical graph
                          │
                          ▼
            composers in Go (R4)  — entity headers, walks, naming matrix
              │         │          │            │
           lists (R5) details (R6) Promote (R7) search docs (R8)   … tree, timeline later
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

### R2 — Value resolution (multi-value Properties)

**Multiple values are first-class.** Any member Subject may carry any Property more than once, and a handle has many members, so every Property on a canonical entity is a list. Displaying one needs a rule. The **resolver** is a set of pure Go functions: candidates in, ranked clusters out. Its output is stored only in the derived cache (R3); it never creates a claim or truth row ([`conclusion-layer-data-model.md`](../../conclusion-layer-data-model.md) §2.5, §7; [`research-judgment-model.md`](../../research-judgment-model.md) §1.1).

#### Three states

| State | When | Display |
| --- | --- | --- |
| **Single** | One candidate value | That value |
| **Resolved / mixed** | Several candidates, no accepted Reconciliation Claim | Names and dates: the **auto-reconciler** output. Everything else: the top-**ranked** value. Either way, a *mixed* indicator when the candidates did not fully agree. |
| **Concluded** | An accepted Reconciliation Claim exists | The claim's value, always. The researcher made the call. |

Reconciliation Claims are out of scope, so **Concluded** never occurs this spike. The resolver still takes an optional concluded value as input, so Reconciliation plugs in later without changing callers.

#### Ranking (provenance)

Each candidate carries its provenance from the evidence trail. Rank by, in order:

1. **Source credibility** of the Citation's root Source (`high_trust` > `standard` > `low_trust`; no assessment = `standard`)
2. **Citation transcription certainty** (`transcription_uncertain = 0` beats `1`)
3. **Identity Claim confidence** of the member the Observation sits on (`high_confidence` > `moderate` > `low_confidence`; none = `moderate`)
4. **Agreement** — how many candidates support the value (after auto-reconcile clustering)
5. Stable tiebreak (Observation id)

Negative-polarity Observations are never displayed as values; they count *against* a matching positive candidate's agreement. The ranking is display policy: the cache stores its order, never a score as catalog truth (research-judgment §1.1). The weighting is expected to change as dogfood shows where it misleads — a change is a cache version bump.

#### Auto-reconcilers (names and dates)

Two pure functions: **n DateValues → one DateValue** and **n NameValues → one NameValue**, each also returning whether the result is exact, merged, or mixed, and which inputs fed it. Inputs arrive ranked, so ties go to the better-provenanced value. Refined over time; expect ~50+ table-driven test cases each, and treat that table as the spec.

- **Dates** ([`structured-date-model.md`](../../structured-date-model.md) — missing components are unknown, never zero):
  - A less precise point containing a more precise one merges to the precise one (`MAY 1985` + `14 MAY 1985` → `14 MAY 1985`).
  - Shared components are kept; disagreeing finer components widen to a `range` (`3 MAY 1985` + `14 JUN 1985` → `BET 3 MAY 1985 AND 14 JUN 1985`), or drop to the shared precision when that reads better (`1985`) — a test-case decision.
  - `ABT` / `BEF` / `AFT` / `range` inputs merge when their windows overlap.
  - Disagreeing years with no overlap → **mixed**: top-ranked value shown with the indicator.
- **Names** ([`structured-name-model.md`](../../structured-name-model.md) — typed parts first, `form` fallback):
  - Surnames that match merge into one.
  - Initials expand against a full part that starts with the same letter (`J. Robins` + `James Robins` → `James Robins`).
  - Case, punctuation, and whitespace differences are ignored.
  - Given names that disagree are laid out as a stream (`James / Jim Robins`) rather than picking one — exact shape is a test-case decision.
  - Untyped parts fall back to comparing `form`. Display uses `form`; there is no name format profile yet.
- **Subject-valued Properties** (association ends: `person`, `event`, `place`, `related_to`): each candidate subject maps to its accepted handle. Candidates whose subject is unpromoted drop out (Q3). The result is an entity id — an edge of the canonical graph.
- **Everything else** (`toponym`, `event_type`, `role`, text, terms): no merge. Distinct values ranked.

The same reconcilers drive the Promote comparison's "compatible pairs start checked" (R7 step 3).

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
  support          INT   -- candidates backing this cluster
  PRIMARY KEY (entity_id, property_id, rank)
  INDEX (property_id, value_entity_id)   -- reverse edges: who points at PER-1?
  INDEX (property_id, sort_key)          -- sort, resemblance
  INDEX (property_id, date_lo, date_hi)  -- date windows: "born 1800–1850", timelines
```

- **Useful for sorting, querying, and resolving (Q12).** Structured values are stored whole as the protobuf messages FFI already uses (`DateValueInput` / `NameValueInput` in [`engine.proto`](../../../api/proto/engine.proto); renaming them to plain value messages is optional cleanup). SQL never decodes them: ordering uses `sort_key`, date range queries use `date_lo` / `date_hi` (the window the date resolver already computes to merge), and the resolver and composers decode in Go. Resolver output is never written to `date_values` / `name_values`.

- **State is the shape of the clusters, never stored.** The resolver (S9-05) returns a list of clusters plus a *concluded* flag; readers derive `single` / `merged` / `mixed` from the rows (rank 2 exists → mixed; else rank 1's `support` 1 → single, more → merged). A `concluded` flag column arrives with Reconciliation.
- **Edges are resolved values.** A Participation's `person` end is its resolved `person` Property with `value_entity_id`. The canonical graph is this table plus its reverse index; there is no separate edges table.
- **Every cluster is kept**, not just the winner: lists get *+N*, search gets alternates (*Jim*), Promote compares against every value, details render clusters without re-resolving.
- **No vocabulary in the schema.** No `birth_date` columns, no per-kind tables. Genealogical concepts live in Go (resolver, composers, registry of recognized keys) and in `property_id` rows. Labels are not stored (term **ids** are), so relabels need no recompute.
- **Maintained in the write transaction.** Every trigger is one hop from the write:

  | Write | Recompute |
  | --- | --- |
  | Observation save / delete on subject S, Property P | S's handle (every Property — see below) |
  | Identity Claim S → E created (Promote) or removed (Subject delete CASCADE) | every Property of E; plus (E′, P′) for handles whose members' Observations point at S (their end now maps differently) |
  | Identity Claim confidence change | every Property of that handle |
  | Citation certainty / Source credibility change | handles with member Observations under it |
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
- **Person birth / death place:** that Event ← Location (`event` end) → Place (`place` end) → resolved `toponym`.
- **Event date:** resolved `date`, else `start_date`–`end_date`.
- **Event place:** Event ← Locations → Places → resolved `toponym`. Several Locations (York *and* Upper Canada) are all returned.
- **Event name:** composed from `event_type` and the people on it through the **naming matrix** (below).
- **Place:** resolved `toponym` clusters.

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
| **Places** | thumbnail slot · toponym · ref (`PLC-…`) |

- **One value per cell.** The rank-1 value. A *mixed* indicator when its state is mixed; a *+N* count when there are more clusters (toponyms, Locations).
- Name / toponym fall back to `label` → `ref`. Missing dates and places are empty, not "Unknown."
- Sort: Persons by name, Events by date, Places by toponym (R3 `sort_key`; default only, no sort controls this spike).
- Only `person`, `event`, and `place` get pages. Association handles exist but are not listed.
- Skills: [`add-workspace-place`](../../../.cursor/skills/add-workspace-place/SKILL.md), [`add-catalog-query`](../../../.cursor/skills/add-catalog-query/SKILL.md).

### R6 — Detail pages (Person, Event, Place)

- **Person:** thumbnail slot, name, birth date, death date, birth place, death place.
- **Event:** thumbnail slot, event title, event date, event place(s).
- A detail page is always a superset of its list row (Q11).
- **Place:** thumbnail slot, toponym(s) — every cluster.
- Every value shows its state — single, merged, mixed, or empty — and can list its ranked clusters from R3. Drilling into the Observations behind a cluster loads live for that one handle. A member list (Subjects with their Source) is a stretch goal.
- Read-only this spike.

### R7 — Promote (Identity Claim workflow)

The hard part. Model: [`conclusion-layer-data-model.md`](../../conclusion-layer-data-model.md) §5.1–§5.4.

**Promote only creates claims.** It is one-way: nobody comes back into it to edit, re-pin, or remove a claim. Those are a separate workflow with its own place (Spike 10), even though they write the same tables. No resume-midstream affordances.

1. **Entry.** A **Promote** control at the bottom of a subject card on the Evidence graph, for primary kinds (`person`, `event`, `place`) that have no accepted Identity Claim. Bridge cards are not entry points; they join through the walk.
2. **Choose target.** Mint a new handle, or pick an existing one of the same Subject type. The picker shows R4 headers. Suggestions: handles already related to the one just filed (during the walk), then resemblance (resolved name, date, toponym, event type — via R3 `sort_key` and the R8 index).
3. **Compare (existing handle only).** Line up the incoming Subject's Observations against each accepted member's Observations for the same Property. Compatible pairs start checked (R2 rules); differing pairs start unchecked. Accept all checked for a Property in one gesture, clear a pair, or skip and accept with no pins. Checked pairs are pinned; unchecked pairs are simply not pinned.
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
  | Place | resolved toponym | — | every toponym cluster, `label` |

  Both FTS indexes (unicode61 + trigram) stay in sync; bump `ProjectionVersion`.
- **Hit rows.** `PVOmnibarHitRow` per kind: thumbnail placeholder, title, kind, ref, subtitle from the header (Person: birth–death; Event: date · place). Conclusion hits carry the **structured header** and Swift formats it, like the list row; Go's FTS text is for matching only. This extends the `SearchHit` DTO (today's kinds send Go-built `display_title` strings).
- **Reprojection.** Search documents are stored (FTS needs them) and composed across handles, so they have one more level of dependency than R3. When R3 rows for handle E change in a transaction, reproject E and its **header dependents** — handles whose header composer reads E, found by the fixed reverse walk over R3's reverse index (Place → Locations → Events → subject Persons; Person → subject Participations → Events). The dependents function lives next to the composers and is covered by the same rebuild-and-compare test. `ProjectionVersion` bump → full rebuild, as today.
- **Out:** Interpretation Subjects (`CPR-…`) as hits; cross-root association queries ("John Smith *with* a birth certificate", omnibar-search *Associated entities*).

---

## Data fetching

The Conclusion pages read **across** Sources — every earlier place scoped to one Source or one vocabulary. R3 absorbs that cost at write time.

- **Go composes; Swift renders.** Swift asks for "Persons list" or "Person X" and displays the payload. It never assembles members or walks ([`macos-client-patterns.md`](../../macos-client-patterns.md) *One cache owns each list*).
- **Query keys:** one list key per kind, one detail key per handle, and the Promote reads (target suggestions, comparison rows for a subject × entity, the neighborhood of a subject). The Promote draft is interaction state, not a cache key.
- **Swift invalidation: bust all Conclusion keys (Q5).** Any Interpretation or Conclusion write that R3 recomputes marks every cached Conclusion key stale — cheap, because reloading reads R3.
  - **Triggers:** Observation save / delete, Subject delete, bridge create / delete, Source delete, credibility / certainty changes, and each Promote step. Shipped in S9-07 as one set, `CatalogQueryRegistry.conclusionTriggers` (`savedCitation`, `deletedSubject`, `promotedSubject`, `deletedSource`, `mutatedSourceWorkspace` — the last carries credibility and over-busts on notes / artifacts / metadata). Certainty joins in S9-14.
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
| **S9-D5** | Person detail | **S9-16** | S9-32 (life dates and places) |
| **S9-D11** | Promote — compare | **S9-19** | — |
| **S9-D3** | Events list | **S9-23** | S9-32 (subject titles, places) |
| **S9-D6** | Event detail | **S9-24** | S9-32 (subject titles, places) |
| **S9-D4** | Places list | **S9-26** | — |
| **S9-D7** | Place detail | **S9-27** | — |
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
  ✎ S9-D9  ──▶ S9-11  Promote shell + choose target (card Promote now opens it)
  ✎ S9-D10 ──▶ S9-12  Claim fields + save
  Check: two records on one Person → one list row; the second card shows the same PER-….

SLICE 4 — Person detail + name resolution
  S9-13  Name auto-reconciler (pure Go, table-driven)
  S9-14  Provenance ranking + polarity; upkeep on credibility / certainty changes
  S9-15  Detail composer + detail read + value-state formatting
  ✎ S9-D5 ──▶ S9-16  Person detail (name with states and clusters)
  Check: J. Robins + James Robins → merged; James / Jim → mixed with alternates;
         raising a Source's credibility reorders them.

SLICE 5 — Compare, pins, backfill
  S9-17  Comparison read + pins + backfill in the Promote write
  S9-18  Pinned-Observation delete end to end (composer confirm names the evidence it leaves)
  ✎ S9-D11 ──▶ S9-19  Promote compare
  Check: join with confirmed pairs → both claims pinned; delete a pinned Observation → allowed, confirm names the handle, pins audited away.

SLICE 6 — Events
  S9-20  Seed event_name
  S9-21  Date auto-reconciler + date windows (pure Go, table-driven)
  S9-22  Event composer (title precedence without subjects, date) + reads + title formatting
  ✎ S9-D3 ──▶ S9-23  Events list (sidebar Events goes live)
  ✎ S9-D6 ──▶ S9-24  Event detail
  Check: promoted event cards list with titles; MAY 1985 + 14 MAY 1985 merge; APR vs MAY → mixed.

SLICE 7 — Places
  S9-25  Place composer + reads
  ✎ S9-D4 ──▶ S9-26  Places list (sidebar Places goes live)
  ✎ S9-D7 ──▶ S9-27  Place detail
  Check: a Place with Upper Canada / U.C. shows both, ranked.

SLICE 8 — Walk + bridges
  S9-28  Edges: subject-valued resolution, bridge filing in the Promote write, inbound-end upkeep
  S9-29  Neighborhood read
  ✎ S9-D12 ──▶ S9-30  Promote walk + off-ramp
  Check: walk a birth record → Participation + Location handles filed; bridges wait for both ends; Done keeps saved steps.

SLICE 9 — Derived values across the graph
  S9-31  Composer walks: life dates + places, subject titles (et al., unnamed person), event places; header dependents
  S9-32  Fill derived cells in Persons / Events lists and Person / Event detail (designed in D2 / D3 / D5 / D6)
  S9-33  Deep fixture + timings
  Check: "Birth of James Robins"; James shows 1817 – 1880 · York → Toronto; timings in the ledger.

SLICE 10 — Search
  S9-34  Search kinds + documents + reprojection
  ✎ S9-D13 ──▶ S9-35  Omnibar Conclusion hits
  Check: PER-7KD45, Jim Robins, Birth of James, a toponym → each finds its handle; a name edit updates the hit.

CLOSE
  S9-99  Dogfood close / docs
```

- **Slices run in order.** Within a slice, PRs run top to bottom; the Go PRs at the top of a slice can usually go side by side (S9-13 / S9-14; S9-20 / S9-21).
- **Slices 6 and 7 can swap or run beside slices 4–5**; they need only slices 1–3.
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
| S9-11 Promote shell + choose target | **S9-D9** | S9-04, S9-10 |
| S9-12 Claim fields + save | **S9-D10** | S9-11 |
| S9-13 Name auto-reconciler | — | S9-05 |
| S9-14 Provenance ranking | — | S9-06 |
| S9-15 Detail composer + read | — | S9-07 |
| S9-16 Person detail | **S9-D5** | S9-13, S9-14, S9-15 |
| S9-17 Compare read + pins + backfill | — | S9-12, S9-13 |
| S9-18 Pinned-Observation delete end to end | — | S9-17 |
| S9-19 Promote compare | **S9-D11** | S9-17 |
| S9-20 Seed `event_name` | — | — |
| S9-21 Date auto-reconciler | — | S9-05 |
| S9-22 Event composer + reads | — | S9-15, S9-20, S9-21 |
| S9-23 Events list | **S9-D3** (extends D2) | S9-09, S9-22 |
| S9-24 Event detail | **S9-D6** (extends D5) | S9-16, S9-22 |
| S9-25 Place composer + reads | — | S9-15 |
| S9-26 Places list | **S9-D4** (extends D2) | S9-09, S9-25 |
| S9-27 Place detail | **S9-D7** (extends D5) | S9-16, S9-25 |
| S9-28 Edges + bridge filing | — | S9-12 |
| S9-29 Neighborhood read | — | S9-28 |
| S9-30 Promote walk | **S9-D12** (extends D9) | S9-29 |
| S9-31 Composer walks | — | S9-22, S9-25, S9-28 |
| S9-32 Fill derived cells | (D2 / D3 / D5 / D6) | S9-31, S9-23, S9-24 |
| S9-33 Deep fixture + timings | — | S9-31 |
| S9-34 Search kinds + reprojection | — | S9-31 |
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
- [ ] ✎ S9-D1 — Design: workspace sidebar
- [x] S9-07b — Rename configuration: Metadata, Properties → [`completed.md`](completed.md)
- [ ] S9-08 — Sidebar sections: Source, Conclude, Configure
- [ ] ✎ S9-D2 — Design: Persons list
- [ ] S9-09 — Persons list
- [ ] S9-10 — Promote write + reads: existing target
- [ ] ✎ S9-D9 — Design: Promote shell + choose target
- [ ] S9-11 — Promote shell + choose target
- [ ] ✎ S9-D10 — Design: Promote claim fields
- [ ] S9-12 — Promote claim fields + save
- [ ] S9-13 — Name auto-reconciler
- [ ] S9-14 — Provenance ranking + polarity
- [ ] S9-15 — Detail composer + detail read
- [ ] ✎ S9-D5 — Design: Person detail
- [ ] S9-16 — Person detail
- [ ] S9-17 — Compare read + pins + backfill
- [ ] S9-18 — Pinned-Observation delete end to end
- [ ] ✎ S9-D11 — Design: Promote compare
- [ ] S9-19 — Promote compare
- [ ] S9-20 — Seed `event_name`
- [ ] S9-21 — Date auto-reconciler + windows
- [ ] S9-22 — Event composer + reads
- [ ] ✎ S9-D3 — Design: Events list
- [ ] S9-23 — Events list
- [ ] ✎ S9-D6 — Design: Event detail
- [ ] S9-24 — Event detail
- [ ] S9-25 — Place composer + reads
- [ ] ✎ S9-D4 — Design: Places list
- [ ] S9-26 — Places list
- [ ] ✎ S9-D7 — Design: Place detail
- [ ] S9-27 — Place detail
- [ ] S9-28 — Edges + bridge filing
- [ ] S9-29 — Neighborhood read
- [ ] ✎ S9-D12 — Design: Promote walk
- [ ] S9-30 — Promote walk + off-ramp
- [ ] S9-31 — Composer walks + header dependents
- [ ] S9-32 — Fill derived cells in lists and details
- [ ] S9-33 — Deep fixture + timings
- [ ] S9-34 — Search kinds + reprojection
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

**Done.** See [`completed.md`](completed.md#s9-05--resolver-core-v1). `core/resolve.Resolve(valueType, candidates, concluded)` → `Result{Clusters, Concluded}`; state is `Result.State()`. S9-13 / S9-21 replace the per-type cluster key (`key` in `resolve.go`); S9-14 puts provenance ahead of support in the order.

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

| | |
| --- | --- |
| **In** | Per **S9-D1**: titled sections. **Source** (Sources) and **Conclude** (Persons live; Events / Places present but stubbed until S9-23 / S9-26) top-aligned, with counts. **Configure** — **Source types**, **Metadata**, **Properties** (names from S9-07b) — bottom-aligned directly above the session footer, separated by space; on a short window the column scrolls as one (W-5b). Rail icons. **Narrate** is reserved after Conclude but hidden (Narrative layer, later spike). Sections don't collapse; the configuration destinations stop being children of Sources and become Configure's own rows. |
| **Check** | Persons count matches promoted Persons; configuration sits above the footer on a tall window and follows the research section on a short one. |
| **Depends on** | **S9-D1**, S9-07, S9-07b |

#### S9-09 — Persons list

| | |
| --- | --- |
| **In** | Per **S9-D2**: rows with thumbnail placeholder, name, ref; empty state. Life-date and place cells render empty until S9-32. The graph card's membership row shows the resolved name. The graph card's membership row swaps *Open person page* for the handle's resolved name — same slot, no relayout (S9-D8). |
| **Check** | Promoted Persons listed by name; edit a name Observation on a member → row updates. |
| **Depends on** | **S9-D2**, S9-07, S9-08 |

### Slice 3 — Join an existing Person

#### S9-10 — Promote write + reads: existing target

| | |
| --- | --- |
| **In** | Promote write accepts an existing handle of the same type (claim only; cache upkeep). Target-suggestion read: same-type handles as headers, resemblance via cache `sort_key`. Confidence grade + argument on the claim. |
| **Out** | Pins / backfill (S9-17); related-first suggestions during a walk (S9-29). |
| **Testable** | Join writes a claim onto the existing handle; cache recomputed; suggestions filtered by type. |
| **Depends on** | S9-06 |

#### S9-11 — Promote shell + choose target

| | |
| --- | --- |
| **In** | Per **S9-D9**: the Promote place (or sheet) with step navigation and leave guard; choose target (new vs existing, suggestions). The card's Promote now opens it (replaces the S9-04 confirm). |
| **Depends on** | **S9-D9**, S9-04, S9-10 |

#### S9-12 — Promote claim fields + save

| | |
| --- | --- |
| **In** | Per **S9-D10**: Status dropdown (one option), confidence, argument; Save writes the step; Done exits. |
| **Check** | Promote a second census person onto an existing Person → one list row, both cards show the same PER-…. |
| **Depends on** | **S9-D10**, S9-11 |

### Slice 4 — Person detail + name resolution

#### S9-13 — Name auto-reconciler

| | |
| --- | --- |
| **In** | Pure Go, table-driven (~50+ cases): surname merge, initial expansion, normalization, given-name stream, `form` fallback. Plugged into the resolver for `name` values; cache version bump. |
| **Depends on** | S9-05 |

#### S9-14 — Provenance ranking + polarity

| | |
| --- | --- |
| **In** | Resolver ranking by Source credibility → Citation certainty → member claim confidence → agreement → id; negative polarity excluded and counted against. Upkeep triggers for credibility, transcription-certainty, and claim-confidence changes; extend the rebuild-equals-upkeep tests. Cache version bump. **Swift:** a certainty mutation joins `conclusionTriggers`; split a dedicated credibility mutation out of `mutatedSourceWorkspace` if its over-bust shows up. |
| **Depends on** | S9-06 |

#### S9-15 — Detail composer + detail read

| | |
| --- | --- |
| **In** | Go detail composer (fields with states and all clusters, support counts); detail FFI; Swift store + detail key (stub view); value-state formatting (single / merged / mixed / empty, *+N*). **Eviction:** `WorkspaceSession` learns the visible place and evicts non-visible Conclusion detail keys on a Conclusion trigger (the rule-2 exception, deferred from S9-07). |
| **Depends on** | S9-07 |

#### S9-16 — Person detail

| | |
| --- | --- |
| **In** | Per **S9-D5**: header with name, states, clusters; life-date and place rows render empty until S9-32. Route the graph card's `.openHandle` target (`EvidenceGraphModel.openHandle`) to this page for persons. |
| **Check** | *J. Robins* + *James Robins* → merged; *James* / *Jim* → mixed with alternates; raising one Source's credibility reorders them. |
| **Depends on** | **S9-D5**, S9-13, S9-14, S9-15 |

### Slice 5 — Compare, pins, backfill

#### S9-17 — Compare read + pins + backfill

| | |
| --- | --- |
| **In** | Comparison read: incoming Observations × each member's per Property, compatibility from the resolver. Promote write takes confirmed pairs: pins on the new claim **and** backfill onto the member's claim, same transaction. |
| **Testable** | Pins on both claims; older `argument` untouched; compatible pairs flagged. |
| **Depends on** | S9-12, S9-13 |

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

#### S9-21 — Date auto-reconciler + windows

| | |
| --- | --- |
| **In** | Pure Go, table-driven (~50+ cases): containment, overlap, range widening vs shared precision, qualifiers, disjoint → mixed. `date_lo` / `date_hi` computed and stored. Cache version bump. |
| **Depends on** | S9-05 |

#### S9-22 — Event composer + reads

| | |
| --- | --- |
| **In** | Event header + detail: title precedence without subjects (`event_name` → label → *Unspecified {type}* → ref), date (else span). List / detail / count FFI and keys. Swift event-title formatter (L10n templates, full matrix so S9-31 only supplies parts). |
| **Depends on** | S9-15, S9-20, S9-21 |

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

### Slice 7 — Places

#### S9-25 — Place composer + reads

| | |
| --- | --- |
| **In** | Place header + detail (toponym clusters); list / detail / count FFI and keys. |
| **Depends on** | S9-15 |

#### S9-26 — Places list

| | |
| --- | --- |
| **In** | Per **S9-D4**; sidebar Places goes live. |
| **Depends on** | **S9-D4**, S9-09, S9-25 |

#### S9-27 — Place detail

| | |
| --- | --- |
| **In** | Per **S9-D7**. Route the graph card's `.openHandle` target to this page for places. |
| **Check** | A Place with *Upper Canada* / *U.C.* shows both, ranked, mixed. |
| **Depends on** | **S9-D7**, S9-16, S9-25 |

### Slice 8 — Walk + bridges

#### S9-28 — Edges + bridge filing

| | |
| --- | --- |
| **In** | Resolver: subject-valued Properties map to the target's handle (unpromoted drop out) — the canonical graph. Upkeep: a claim create / remove recomputes handles whose members' Observations point at that subject. Promote write files bridge subjects onto the association the ends already share, or mints one (lift the S9-02 primary-kinds guard in `promote.Save`). Extend the rebuild-equals-upkeep tests. Once bridge edge Observations can be pinned, test that deleting a bridge Subject releases those pins audited: its connection-facet release erases each edge Observation through `ReleaseFacets(KindObservation)`, which takes the pins first. |
| **Depends on** | S9-12 |

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

### Slice 9 — Derived values across the graph

#### S9-31 — Composer walks + header dependents

| | |
| --- | --- |
| **In** | Person birth / death date and place; Event subject titles (*Birth of …*, marriage, *et al.*, *unnamed person*, *{Type} at {toponym}*) and places. Header-dependents function (reverse walk) for later reprojection. |
| **Depends on** | S9-22, S9-25, S9-28 |

#### S9-32 — Fill derived cells

| | |
| --- | --- |
| **In** | Persons / Events lists and Person / Event detail render the new cells — already designed in **S9-D2 / D3 / D5 / D6**; no new brief. |
| **Check** | *Birth of James Robins*; James shows *1817 – 1880 · York → Toronto*. |
| **Depends on** | S9-31, S9-23, S9-24 |

#### S9-33 — Deep fixture + timings

| | |
| --- | --- |
| **In** | Seeded project generator (a Person on ~10 Sources, multi-member events with Locations, relationships; a few hundred handles). Benchmarks: rebuild, one-write upkeep, one Promote step, list and detail composition. Ledger rows. |
| **Depends on** | S9-31 |

### Slice 10 — Search

#### S9-34 — Search kinds + reprojection

| | |
| --- | --- |
| **In** | Registry kinds `person` / `event` / `place`; location mapping; documents from headers (match text only); reprojection of the handle and its header dependents in the write transaction (deletes feed it `Released.Handles`); `SearchHit` carries the structured header; `ProjectionVersion` bump; FakeStore. Existing hit rows render a fallback until S9-35. |
| **Depends on** | S9-31 |

#### S9-35 — Omnibar Conclusion hits

| | |
| --- | --- |
| **In** | Per **S9-D13**. |
| **Check** | `PER-7KD45`, *Jim Robins*, *Birth of James*, a toponym → each finds its handle; a member name edit updates the hit. |
| **Depends on** | **S9-D13**, S9-34 |

### S9-99 — Dogfood close / docs

Honesty pass against the [goal bar](#goal-dogfood-bar); ledger timings recorded; docs in *Docs to update* folded in; briefs archived; spike archived via [`archive-docs`](../../../.cursor/skills/archive-docs/SKILL.md).

---

## Scope boundary

| In | Out |
| --- | --- |
| Canonical entities, Identity Claims, evidence pins | Reconciliation Claims, `name_format` |
| Resolver: states, provenance ranking, name / date auto-reconcilers | Persisted "auto" claims; ranking stored as catalog truth |
| Resolved-values cache with upkeep, rebuild, and rebuild-equals-upkeep test | Per-screen caches; stored derived values (life dates, event names); resident in-memory graph (only if timings demand) |
| Header composers shared by lists, Promote, search | — |
| Promote from the Evidence graph: per-step saves, walk, Done off-ramp | Stub handles with no Subject; canonical merge |
| Promote creates claims (one-way) | Editing, re-pinning, or removing Identity Claims; removing a member (Spike 10) |
| Promote writes `accepted` claims | `provisional` / `rejected` in Promote (later spike) |
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
6. **Cascades bypass Go — so Go doesn't rely on them.** Every official delete calls `deleteimpact.ReleaseFacets`, which removes and audits claims, pins, notes, and connection facets before the parent `DELETE` and fails if anything is left for the `CASCADE` backstop. Its `Released.Handles` is how cache upkeep (S9-06) and handle search reprojection (S9-34) learn which handles a delete touched — computed before anything is gone. A new delete path gets this by calling `ReleaseFacets`, not by hand-wiring helpers.
7. **Promoting a subject can change other handles.** Its Observations' targets and inbound subject-valued Observations re-map; R3's claim trigger covers the second hop.
8. **One accepted claim per Subject** is a partial unique index. Promote must hide or refuse subjects that are already members, and a step that loses a race fails alone — earlier steps stay saved.
9. **Composite FKs** carry `subject_type_id` on the claim. A person Subject cannot be claimed onto a Place; the target picker filters by type so the researcher never sees that error.
10. **Backfill is symmetric.** A confirmed pair pins both Observations on the incoming claim **and** the existing member's claim. The older claim's `argument` is not rewritten.
11. **Pins are Observations only.** Not Citations, Subjects, or Sources. Confirming a match creates no Observation.
12. ~~**Stale reservation.**~~ Go side retired in S9-01. Swift still carries `sameness_claim` L10n / preview until S9-18.
15. **Pins never block.** A pinned Observation (or a Subject whose Observations are pinned on other members' claims) deletes freely; the claims stay with a weaker exhibit. Do not reintroduce a blocking evidence probe — weak claims are the §5.2 review alert's job.
13. **Candidate vs canonical refs.** Subjects are `CPR-…`; handles are `PER-…`. Both prefixes already exist on `subject_types`.
14. **Cross-Source reads** run on the serialized catalog session. Keep rebuild off the open path's critical section if it gets long.

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
| Q9 | ~~Event with two dates.~~ | **Decided:** no different from any multi-value Property. R2's three states and provenance ranking apply everywhere. |
| Q10 | ~~Event naming matrix.~~ | **Decided (R4):** `subject` = principal(s), marriage is two subjects; *et al.* for several; *{Type} at {toponym}* / *Unspecified {type}* with no subject; *unnamed person*; generic templates for researcher-added types; precedence recorded name → composed → label → type/place → ref. |
| Q11 | ~~Details vs lists.~~ | **Decided:** a detail page is a superset of its list row. Person adds birth / death place; Event adds place(s) (R6). |
| Q12 | ~~Serialized value format in R3.~~ | **Decided:** protobuf DateValue / NameValue in `value_date` / `value_name`, plus `date_lo` / `date_hi` ordinals for range queries; never `date_values` / `name_values` rows. Go returns structures; Swift formats text. |
| Q13 | ~~Event recorded-name Property.~~ | **Decided:** seed `event_name` (`value_type = text`) and bind it to `event` this spike (R1). |
| Q14 | ~~Label vs composed-from-subjects.~~ | **Decided:** label stays below subject titles; it wins only when there is no subject. Revisit on dogfood. |

---

## Docs to update as work lands

- [`conclusion-layer-data-model.md`](../../conclusion-layer-data-model.md): §5.3–§5.4 "future UI" → shipped behavior (per-step saves; grounding with zero pins; create-only). New section: the resolved-values cache as a derived, rebuildable projection; edges as resolved subject-valued Properties.
- [`research-judgment-model.md`](../../research-judgment-model.md) §1.1: cached rank order is derived, not stored judgment.
- [`seeded-vocabulary.md`](../../seeded-vocabulary.md) §5.5: mark claim confidence grades as seeded. §3.5: `subject` = the event's principal(s), possibly several; marriage uses two `subject` Participations; `spouse` is a principal's spouse on another event. §3.2 / §3.3: `event_name` (text) bound to `event`.
- [`catalog-refs.md`](../../catalog-refs.md): canonical ref minting.
- [`catalog-deletes.md`](../../catalog-deletes.md): Identity Claim / evidence Impact (done in S9-02: non-blocking cascades, explicit audited release); hook cache dependents into the release calls (S9-06).
- [`macos-client-patterns.md`](../../macos-client-patterns.md): Conclusion query keys and their invalidation rule; Go returns structures (DateValue, NameValue, title parts), Swift formats text.
- [`omnibar-search.md`](../archive/spike-3/omnibar-search.md): new kinds, location mapping, composed documents and header-dependent reprojection.

---

## Definition of done

- [ ] Dogfood bar items 1–9 met on real research
- [ ] Checklist complete (or items explicitly descoped)
- [ ] Design briefs archived under `design/archive/`
- [ ] Open questions answered and folded into the model docs
- [ ] [`docs/deployment-plan/README.md`](../README.md) points at the archive when closed
