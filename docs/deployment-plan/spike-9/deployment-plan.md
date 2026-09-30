# Deployment Plan — Spike 9

MVP for the **Conclusion layer**: assemble canonical Persons, Events, and Places from Interpretation Subjects through Identity Claims, and show them on list and detail pages. Authoritative model: [`conclusion-layer-data-model.md`](../../conclusion-layer-data-model.md). Values: [`structured-name-model.md`](../../structured-name-model.md), [`structured-date-model.md`](../../structured-date-model.md). Vocabulary: [`seeded-vocabulary.md`](../../seeded-vocabulary.md) §3, §5.

## Status

**Open.** Requirements, [design track](#design-track), and [PR sequence](#pr-sequence) drafted; briefs S9-D1…D13 written in [`design/`](design/). Landings go in [`completed.md`](completed.md).

> **Goal of this spike:** a researcher can promote Subjects off an Evidence graph into Persons, Events, and Places, and open a page for each that shows who or what it is — name and life dates, event and date, toponym — resolved from every member Subject.

> **Foundation first.** The resolved-values cache (R3) is core infrastructure every Conclusion surface reads — lists, details, Promote, search, and later the tree. Build it properly. Work may land piecemeal, with stubs and temporarily broken UI along the way; no band-aid caches per screen.

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
| **Delete Impact** | An Observation pinned by `identity_claim_evidence` names that claim. A Subject with an Identity Claim names the handle it would leave (the claim itself CASCADEs per the model). Replace the stale `ViaSamenessEvidence` reservation with the identity-evidence via. |
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
  state            TEXT  -- single | merged | mixed   (later: concluded)
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

- **Edges are resolved values.** A Participation's `person` end is its resolved `person` Property with `value_entity_id`. The canonical graph is this table plus its reverse index; there is no separate edges table.
- **Every cluster is kept**, not just the winner: lists get *+N*, search gets alternates (*Jim*), Promote compares against every value, details render clusters without re-resolving.
- **No vocabulary in the schema.** No `birth_date` columns, no per-kind tables. Genealogical concepts live in Go (resolver, composers, registry of recognized keys) and in `property_id` rows. Labels are not stored (term **ids** are), so relabels need no recompute.
- **Maintained in the write transaction.** Every trigger is one hop from the write:

  | Write | Recompute |
  | --- | --- |
  | Observation save / delete on subject S, Property P | (S's handle, P) |
  | Identity Claim S → E created (Promote) or removed (Subject delete CASCADE) | every Property of E; plus (E′, P′) for handles whose members' Observations point at S (their end now maps differently) |
  | Identity Claim confidence change | every Property of that handle |
  | Citation certainty / Source credibility change | handles with member Observations under it |
  | Reconciliation Claim (later) | (E, P) → `concluded` |
  | Term relabel | nothing |

- **Rebuildable.** A version row (like `catalog_search_meta`); a mismatch on Open rebuilds from truth tables. A Go rebuild command for development.
- **Proven.** A randomized test applies write sequences, then checks incremental upkeep equals a full rebuild. Ships in the same PR as the cache.
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

Three sidebar destinations under a new Conclusions group, with counts ([`CatalogCounts`](../../../macos/App/Features/Catalog/CatalogCounts.swift)). Rows come from the R4 header composers (Q6).

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
  - **Triggers:** Observation save / delete, Subject delete, bridge create / delete, Source delete, credibility / certainty changes, and each Promote step.
  - **Evict, don't revalidate.** Detail keys are dropped and reload when next visited; the visible page revalidates. A deliberate exception to cache contract rule 2; note it in [`macos-client-patterns.md`](../../macos-client-patterns.md).
  - Narrowing later can reuse R3's affected-handles set; not needed while reloads are cheap.
- **Measure.** Build the deep fixture (a Person on ~10 Sources; birth, death, marriage events each multi-member, with Locations; a few relationships; scaled to a few hundred handles) and time: rebuild, a single write's upkeep, list composition, detail composition. Record in the [performance ledger](performance-ledger.md).

---

## Design track

**Every UI PR is gated by Claude Design briefs — one brief per view.** A PR that touches two views waits on two briefs. Briefs follow [`add-design-brief`](../../../.cursor/skills/add-design-brief/SKILL.md) and live in [`design/`](design/). **Each brief is designed alongside its feature**, just before the PR it gates — see the [PR sequence](#pr-sequence).

| Brief | View | Covers | Gates |
| --- | --- | --- | --- |
| **S9-D1** | Workspace sidebar | Conclusions group; Persons / Events / Places destinations; counts | **S9-18** |
| **S9-D2** | Persons list | Row anatomy (thumbnail, name, life dates, birth / death place, ref); *mixed* and *+N* markers; sort; empty state pointing at Promote | **S9-19** |
| **S9-D3** | Events list | Extends D2's row anatomy: title, date, place, ref | **S9-20** |
| **S9-D4** | Places list | Extends D2's row anatomy: toponym, ref | **S9-21** |
| **S9-D5** | Person detail | Header: thumbnail, name, life dates, places; value states (single / merged / mixed / empty); cluster list per field | **S9-22** |
| **S9-D6** | Event detail | Extends D5: title, date, place(s) | **S9-23** |
| **S9-D7** | Place detail | Extends D5: every toponym cluster | **S9-24** |
| **S9-D8** | Evidence graph (subject card) | Promote control on primary-kind cards; member badge + link to handle on promoted cards | **S9-25** |
| **S9-D9** | Promote — choose target | Promote shell (place vs sheet decided here); mint vs existing; header rows; suggestions; same-type filter | **S9-26** |
| **S9-D10** | Promote — claim fields | Status dropdown (one option, laid out for three); confidence grade; argument; Next / Done | **S9-27** |
| **S9-D11** | Promote — compare | Existing handle only: Property rows × members; pre-checked compatible pairs; accept-all per Property; skip | **S9-28** |
| **S9-D12** | Promote — walk | Connected-subjects list after each save; bridge confirm; progress; Done off-ramp; leave guard | **S9-29** |
| **S9-D13** | Omnibar results | Person / Event / Place hit rows from structured headers | **S9-30** |

---

## PR sequence

Suggested order, top to bottom. **✎ = design brief**, run through Claude Design just before the PR it points at. Solid dependencies are shown in parentheses where they come from outside the arrow.

```text
FOUNDATION
  S9-01  Conclusion schema + stores
    ├──▶ S9-02  Seed event_name
    └──▶ S9-03  Delete Impact for Identity Claims
  S9-04  Date auto-reconciler ──┐
  S9-05  Name auto-reconciler ──┤
                                ▼
  S9-06  Resolver core
                                ▼                               (+ S9-01)
  S9-07  Resolved-values cache + rebuild
                                ▼
  S9-08  Cache upkeep + rebuild-equals-upkeep test
    └──▶ S9-09  Deep fixture + timings

READS
  S9-10  Header + detail composers                              (← S9-07, S9-02)
    ├──▶ S9-11  Conclusion reads: FFI + Swift store + query keys
    └──▶ S9-12  Swift Conclusion formatters

PAGES                                                            (all ← S9-11, S9-12)
  ✎ S9-D1 ──▶ S9-18  Sidebar Conclusions group
                ├──▶ ✎ S9-D2 ──▶ S9-19  Persons list ──▶ ✎ S9-D5 ──▶ S9-22  Person detail
                ├──▶ ✎ S9-D3 ──▶ S9-20  Events list  ──▶ ✎ S9-D6 ──▶ S9-23  Event detail
                └──▶ ✎ S9-D4 ──▶ S9-21  Places list  ──▶ ✎ S9-D7 ──▶ S9-24  Place detail

PROMOTE
  S9-13  Source graph carries membership                        (← S9-01)
    └──▶ ✎ S9-D8 ──▶ S9-25  Graph card: Promote entry + membership
  S9-15  Promote reads: target suggestions + comparison         (← S9-06, S9-10)
    └──▶ ✎ S9-D9 ──▶ S9-26  Promote shell + choose target       (← S9-25)
  S9-14  Promote step write                                     (← S9-01, S9-08)
    └──▶ ✎ S9-D10 ─▶ S9-27  Promote claim fields + save         (← S9-26)
                       ├──▶ ✎ S9-D11 ─▶ S9-28  Promote compare  (← S9-15)
                       └──▶ S9-16  Promote read: neighborhood    (← S9-01)
                              └──▶ ✎ S9-D12 ─▶ S9-29  Promote walk + off-ramp

SEARCH
  S9-17  Search kinds + reprojection                            (← S9-08, S9-10)
    └──▶ ✎ S9-D13 ─▶ S9-30  Omnibar Conclusion hits             (← S9-12)

S9-99  Dogfood close / docs
```

- **Why pages before Promote.** S9-09's fixture seeds realistic handles, so the lists and details prove the cache and composers visually before the write flow is built on them. Promote can move ahead of Pages if real data matters more; nothing in Pages blocks it.
- **Persons first within Pages.** S9-D3 / D4 extend S9-D2's row, and S9-D6 / D7 extend S9-D5's page, so design the Persons pair before the others.
- **Can run side by side:** S9-02, S9-03, S9-04, S9-05 with S9-01; S9-09 with S9-10; S9-13 and S9-15 with the Pages work; S9-17 any time after S9-10.
- **Churn is expected.** Stub destinations, placeholder rows, and a Promote shell with one working step are fine between PRs.

### Dependencies at a glance

| PR | Brief | Depends on |
| --- | --- | --- |
| S9-01 Conclusion schema + stores | — | — |
| S9-02 Seed `event_name` | — | — |
| S9-03 Delete Impact for Identity Claims | — | S9-01 |
| S9-04 Date auto-reconciler | — | — |
| S9-05 Name auto-reconciler | — | — |
| S9-06 Resolver core | — | S9-04, S9-05 |
| S9-07 Resolved-values cache + rebuild | — | S9-01, S9-06 |
| S9-08 Cache upkeep | — | S9-07 |
| S9-09 Deep fixture + timings | — | S9-08 |
| S9-10 Header + detail composers | — | S9-07, S9-02 |
| S9-11 Conclusion reads | — | S9-10 |
| S9-12 Swift formatters | — | S9-10 |
| S9-13 Source graph carries membership | — | S9-01 |
| S9-14 Promote step write | — | S9-01, S9-08 |
| S9-15 Promote reads: targets + comparison | — | S9-06, S9-10 |
| S9-16 Promote read: neighborhood | — | S9-01 |
| S9-17 Search kinds + reprojection | — | S9-08, S9-10 |
| S9-18 Sidebar | **S9-D1** | S9-11 |
| S9-19 Persons list | **S9-D2** | S9-11, S9-12, S9-18 |
| S9-20 Events list | **S9-D3** (extends D2) | S9-11, S9-12, S9-18 |
| S9-21 Places list | **S9-D4** (extends D2) | S9-11, S9-12, S9-18 |
| S9-22 Person detail | **S9-D5** | S9-11, S9-12 |
| S9-23 Event detail | **S9-D6** (extends D5) | S9-11, S9-12 |
| S9-24 Place detail | **S9-D7** (extends D5) | S9-11, S9-12 |
| S9-25 Graph card | **S9-D8** | S9-13 |
| S9-26 Promote shell + target | **S9-D9** | S9-15, S9-25 |
| S9-27 Promote claim fields + save | **S9-D10** (extends D9) | S9-14, S9-26 |
| S9-28 Promote compare | **S9-D11** (extends D9) | S9-15, S9-27 |
| S9-29 Promote walk | **S9-D12** (extends D9) | S9-16, S9-27 |
| S9-30 Omnibar hits | **S9-D13** | S9-12, S9-17 |
| S9-99 Dogfood close | — | all |

### Milestones

| After | The app can… |
| --- | --- |
| S9-09 | Rebuild the cache on a deep fixture; timings in the ledger (bar 9) |
| S9-19…24 | Browse fixture Persons / Events / Places and open their pages (bar 3–6) |
| S9-27 | Mint a new handle from a graph card — first real data (bar 1, partial; bar 7) |
| S9-28 | Join a second record with pinned comparisons and backfill (bar 2) |
| S9-29 | Walk the neighborhood with a Done off-ramp (bar 1) |
| S9-30 | Find handles in the omnibar (bar 8) |

---

## Checklist

In suggested order; each brief sits just above the PR it gates.

- [ ] S9-01 — Conclusion schema + stores
- [ ] S9-02 — Seed `event_name`
- [ ] S9-03 — Delete Impact for Identity Claims
- [ ] S9-04 — Date auto-reconciler + windows
- [ ] S9-05 — Name auto-reconciler
- [ ] S9-06 — Resolver core
- [ ] S9-07 — Resolved-values cache + rebuild
- [ ] S9-08 — Cache upkeep + rebuild-equals-upkeep test
- [ ] S9-09 — Deep fixture + timings
- [ ] S9-10 — Header and detail composers
- [ ] S9-11 — Conclusion reads: FFI + Swift store + query keys
- [ ] S9-12 — Swift Conclusion formatters
- [ ] ✎ S9-D1 — Design: workspace sidebar
- [ ] S9-18 — Sidebar Conclusions group
- [ ] ✎ S9-D2 — Design: Persons list
- [ ] S9-19 — Persons list
- [ ] ✎ S9-D5 — Design: Person detail
- [ ] S9-22 — Person detail
- [ ] ✎ S9-D3 — Design: Events list
- [ ] S9-20 — Events list
- [ ] ✎ S9-D6 — Design: Event detail
- [ ] S9-23 — Event detail
- [ ] ✎ S9-D4 — Design: Places list
- [ ] S9-21 — Places list
- [ ] ✎ S9-D7 — Design: Place detail
- [ ] S9-24 — Place detail
- [ ] S9-13 — Source graph carries membership
- [ ] ✎ S9-D8 — Design: graph subject card
- [ ] S9-25 — Graph card: Promote entry + membership
- [ ] S9-15 — Promote reads: target suggestions + comparison
- [ ] ✎ S9-D9 — Design: Promote shell + choose target
- [ ] S9-26 — Promote shell + choose target
- [ ] S9-14 — Promote step write
- [ ] ✎ S9-D10 — Design: Promote claim fields
- [ ] S9-27 — Promote claim fields + save step
- [ ] ✎ S9-D11 — Design: Promote compare
- [ ] S9-28 — Promote compare
- [ ] S9-16 — Promote read: neighborhood
- [ ] ✎ S9-D12 — Design: Promote walk
- [ ] S9-29 — Promote walk + off-ramp
- [ ] S9-17 — Search kinds + reprojection
- [ ] ✎ S9-D13 — Design: omnibar hits
- [ ] S9-30 — Omnibar Conclusion hits
- [ ] S9-99 — Dogfood close / docs

---

## PRs

### S9-01 — Conclusion schema + stores

| | |
| --- | --- |
| **In** | Migration: `claim_confidence_grades` (+ seed §5.5), `canonical_entities`, `identity_claims` (composite FKs, `UNIQUE (subject_id, entity_id)`, one-accepted partial index), `identity_claim_evidence`. Canonical ref minting from `subject_types.ref_prefix`. Go stores: create entity, create claim (+ pins), get / list by kind, membership lookups both ways. Audit on every write. |
| **Out** | FFI; Promote orchestration (S9-14); cache (S9-07); Reconciliation tables. |
| **Testable** | Type mismatch rejected by FK; second accepted claim for a subject rejected; ref format; audit rows written; membership queries. |
| **Depends on** | — |

### S9-02 — Seed `event_name`

| | |
| --- | --- |
| **In** | `event_name` (`text`) Property bound to `event`: Install seed + migration for existing projects. Seeded-vocabulary doc. |
| **Out** | Any resolver or title use (S9-10). |
| **Testable** | New and migrated projects both have the Property and binding; Subject Fields shows it on events. |
| **Depends on** | — |

### S9-03 — Delete Impact for Identity Claims

| | |
| --- | --- |
| **In** | Register `identity_claim_evidence.observation_id` as an inbound via (replaces stale `ViaSamenessEvidence`); Observation delete names the pinning claim(s); Subject delete Impact names the handle it would leave. L10n kind strings. |
| **Out** | Cache dependents on delete (S9-08). |
| **Testable** | Pinned Observation blocked with the claim named; Subject delete lists the handle; unpinned paths unchanged. |
| **Depends on** | S9-01 |

### S9-04 — Date auto-reconciler + windows

| | |
| --- | --- |
| **In** | Pure Go: n DateValues (ranked) → one DateValue + state (exact / merged / mixed) + contributing inputs. `date_lo` / `date_hi` window per DateValue (qualifiers, ranges, partial components). Table-driven tests as the spec (~50+). |
| **Out** | Candidate loading; ranking (inputs arrive ranked). |
| **Testable** | Precision containment; overlap merge; range widening vs shared-precision drop; `ABT` / `BEF` / `AFT`; disjoint → mixed; missing components never zero. |
| **Depends on** | — |

### S9-05 — Name auto-reconciler

| | |
| --- | --- |
| **In** | Pure Go: n NameValues (ranked) → one NameValue + state + contributing inputs. Surname merge, initial expansion, case / punctuation normalization, disagreeing-given-name stream, `form` fallback for untyped parts. Table-driven tests (~50+). |
| **Out** | Name format profiles. |
| **Testable** | Each rule plus combinations; untyped-only inputs; single input passes through. |
| **Depends on** | — |

### S9-06 — Resolver core

| | |
| --- | --- |
| **In** | Candidate model with provenance (Source credibility, Citation certainty, member claim confidence, polarity). Ranking (R2 order), clustering, three states, optional concluded input. Dispatch: date → S9-04, name → S9-05, subject-valued → map to handle / drop unpromoted, everything else → distinct ranked. Output: ranked clusters. |
| **Out** | Database access (candidates are passed in). |
| **Testable** | Ranking order and defaults; negative polarity excluded and counted against; subject-valued mapping; concluded input wins; stable tiebreak. |
| **Depends on** | S9-04, S9-05 |

### S9-07 — Resolved-values cache + rebuild

| | |
| --- | --- |
| **In** | Migration: `conclusion_resolved_values` (R3 shape, indexes) + version row. Batched candidate loader (members, Observations, provenance in a fixed number of set-based queries per batch). Full rebuild; version check on Open. Protobuf value columns and `date_lo` / `date_hi`. |
| **Out** | Incremental upkeep (S9-08). |
| **Testable** | Rebuild on a small fixture matches hand-computed rows; version mismatch triggers rebuild; loader query count is constant per batch. |
| **Depends on** | S9-01, S9-06 |

### S9-08 — Cache upkeep + rebuild-equals-upkeep test

| | |
| --- | --- |
| **In** | Affected-handles function. Triggers (R3 table) wired into Observation save / delete, Identity Claim create, Subject delete (computed **before** CASCADE), confidence / credibility / certainty changes — each in the write transaction. Randomized test: write sequences, then incremental == full rebuild. |
| **Out** | Search reprojection (S9-17); Swift invalidation (S9-11). |
| **Testable** | Each trigger individually; randomized equivalence; delete paths leave no stale rows. |
| **Depends on** | S9-07 |

### S9-09 — Deep fixture + timings

| | |
| --- | --- |
| **In** | Seeded project generator: a Person on ~10 Sources, multi-member birth / death / marriage events with Locations, relationships; scaled to a few hundred handles. Benchmarks: rebuild, one-write upkeep, one Promote step (once S9-14 lands), list and detail composition (once S9-10 lands). Ledger rows. |
| **Out** | Optimization. |
| **Testable** | Fixture is deterministic; benchmarks run in CI-less `go test -bench`. |
| **Depends on** | S9-08 (extends as S9-10 / S9-14 land) |

### S9-10 — Header and detail composers

| | |
| --- | --- |
| **In** | Go composers per kind: headers (list rows, pickers, search) and details (fields with states and clusters). Canonical walks via R3 reverse index (`role = subject`, birth / death, Locations → Places). Event title parts per R4 precedence (`event_name` → subjects → label → type / place → ref). Set-based list composition. Protobuf messages for headers, details, and title parts — structures only, no display strings. |
| **Out** | FFI handlers (S9-11); text formatting (Swift, S9-12). |
| **Testable** | Each walk; `et al.` ordering; unnamed person; no-subject titles; list composition query count constant in list length. |
| **Depends on** | S9-07 (S9-02 for `event_name`) |

### S9-11 — Conclusion reads: FFI + Swift store + query keys

| | |
| --- | --- |
| **In** | FFI: list by kind, detail by id, counts. `GenealogyStore` + FakeStore. `CatalogQueryKey`s and `PlaceRegistry` entries for the six places (stub views are fine). Invalidation: bust all Conclusion keys on the R3 trigger mutations; evict details, revalidate visible. `macos-client-patterns.md` note. |
| **Out** | Any real view (S9-18+). |
| **Testable** | Handler round-trips; FakeStore parity; a mutation marks Conclusion keys stale and evicts details. |
| **Depends on** | S9-10 |

### S9-12 — Swift Conclusion formatters

| | |
| --- | --- |
| **In** | Extend `DateValueDisplay` / `NameValueDisplay` for resolved values; event-title formatter from title parts via L10n templates (matrix in R4); value-state model (single / merged / mixed / empty, *+N*). Unit tests only. |
| **Out** | Views. |
| **Testable** | Every matrix row; `et al.`; *Unspecified {type}*; *unnamed person*; localization keys present. |
| **Depends on** | S9-10 (proto shapes) |

### S9-13 — Source graph carries membership

| | |
| --- | --- |
| **In** | The source-graph read includes each subject's accepted handle (id, ref, kind) or none. Swift model field. |
| **Out** | Card UI (S9-25). |
| **Testable** | Promoted / unpromoted subjects reported correctly; graph cache invalidated on claim create. |
| **Depends on** | S9-01 |

### S9-14 — Promote step write

| | |
| --- | --- |
| **In** | One FFI call per step: mint (or target existing) handle + accepted Identity Claim + pins + backfill onto compared members' claims + R3 recompute, one transaction. Rejects type mismatch and already-member subjects. Bridge subjects file onto an existing association or mint one. |
| **Out** | Status other than `accepted` (Q7). |
| **Testable** | Mint path (zero pins); join path with backfill on both claims; race on one-accepted index fails only that step; cache rows updated in the same transaction. |
| **Depends on** | S9-01, S9-08 |

### S9-15 — Promote reads: target suggestions + comparison

| | |
| --- | --- |
| **In** | Target suggestions for a subject: related handles (during a walk) first, then resemblance via R3 `sort_key` and headers. Comparison rows: incoming Observations × each member's Observations per Property, with resolver-driven compatibility pre-check. |
| **Out** | UI. |
| **Testable** | Same-type filter; related-first ordering; compatible pairs flagged; members with several Observations per Property give one row each. |
| **Depends on** | S9-06, S9-10 |

### S9-16 — Promote read: neighborhood

| | |
| --- | --- |
| **In** | For a just-filed subject: unpromoted Interpretation neighbors; queue rules (not handled, not members of the target, not neighbors of other members); bridge readiness (both ends have handles) and the association the ends already share, if any. |
| **Out** | Queue persistence (none, by design). |
| **Testable** | Birth-record walk shape from the model (§5.4); skip does not reach through; bridges wait for both ends. |
| **Depends on** | S9-01 |

### S9-17 — Search kinds + reprojection

| | |
| --- | --- |
| **In** | Registry kinds `person` / `event` / `place`; location mapping; documents built from S9-10 headers (match text only); header-dependent reprojection in the write transaction; `SearchHit` carries the structured header; `ProjectionVersion` bump; FakeStore. |
| **Out** | Hit row UI (S9-30). |
| **Testable** | Ref, name, alternate (*Jim*), event title, toponym hits; a member name edit updates the Person and its subject Events' documents; rebuild matches incremental. |
| **Depends on** | S9-08, S9-10 |

### S9-18 — Sidebar Conclusions group

| | |
| --- | --- |
| **In** | Per **S9-D1**: Conclusions group in `WorkspaceSidebar` with Persons / Events / Places and counts. Destinations open stub pages until S9-19…21. |
| **Depends on** | S9-D1, S9-11 |

### S9-19 / S9-20 / S9-21 — Persons, Events, Places lists

| | |
| --- | --- |
| **In** | One PR per list, per **S9-D2 / D3 / D4**. Rows from headers via S9-12 formatters; default sort; *mixed* / *+N*; empty state. |
| **Depends on** | Its brief, S9-11, S9-12, S9-18 |

### S9-22 / S9-23 / S9-24 — Person, Event, Place detail

| | |
| --- | --- |
| **In** | One PR per detail, per **S9-D5 / D6 / D7**. Header fields with states; cluster list per field; superset of the list row (Q11). Read-only. |
| **Depends on** | Its brief, S9-11, S9-12 |

### S9-25 — Graph card: Promote entry + membership

| | |
| --- | --- |
| **In** | Per **S9-D8**: Promote control on unpromoted primary-kind cards (opens the S9-26 shell, stubbed until then); member badge + link on promoted cards. |
| **Depends on** | S9-D8, S9-13 |

### S9-26 — Promote shell + choose target

| | |
| --- | --- |
| **In** | Per **S9-D9**: the Promote place (or sheet, per the board) with step navigation; choose target (mint vs existing, header rows, suggestions). Leave guard on the step in progress. |
| **Depends on** | S9-D9, S9-15, S9-25 |

### S9-27 — Promote claim fields + save step

| | |
| --- | --- |
| **In** | Per **S9-D10**: Status dropdown (one option), confidence, argument; Next saves via S9-14; Done exits. First end-to-end: mint from a graph card. |
| **Depends on** | S9-D10, S9-14, S9-26 |

### S9-28 — Promote compare

| | |
| --- | --- |
| **In** | Per **S9-D11**: comparison step for existing handles, inserted between target and claim fields; pins and backfill flow into S9-14. |
| **Depends on** | S9-D11, S9-15, S9-27 |

### S9-29 — Promote walk + off-ramp

| | |
| --- | --- |
| **In** | Per **S9-D12**: after each save, connected subjects list; picking one loops the steps; bridge confirm; Done off-ramp. |
| **Depends on** | S9-D12, S9-16, S9-27 |

### S9-30 — Omnibar Conclusion hits

| | |
| --- | --- |
| **In** | Per **S9-D13**: `PVOmnibarHitRow` for the three kinds, formatted from structured headers (S9-12). |
| **Depends on** | S9-D13, S9-12, S9-17 |

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
| Delete Impact naming Identity Claims | Review queue for claims whose comparison member left (§5.2) |
| Omnibar search for Persons / Events / Places (ref + full text on resolved values) | Subjects as hits; cross-root association search |

---

## Gotchas

1. **Persons have no date Properties.** Birth and death dates come through Participation → Event (R4). Do not seed `birth_date` on person to shortcut it.
2. **`event_name`, not `name`.** Historical event names use the seeded `event_name` text Property (Q13). `name` is a personal NameValue — never bind it to `event`.
3. **The cache is not truth.** Nothing references `conclusion_resolved_values`; synthesized dates and merged names live only there, serialized. No claim, DateValue, or NameValue row is written by resolution ([`seeded-vocabulary.md`](../../seeded-vocabulary.md) §5.3).
4. **No vocabulary-named columns** in any derived table. If a screen needs a new concept, it is a composer change or a `property_id` row — never a column.
5. **Upkeep misses are silent.** A trigger the affected-handles function forgets leaves a stale row nobody notices. The rebuild-equals-upkeep test is the guard; every new write path adds its trigger and a test sequence.
6. **Cascades bypass Go.** `ON DELETE CASCADE` (Subject → Identity Claims) removes rows without a Go write path. The Go delete must compute affected handles **before** deleting.
7. **Promoting a subject can change other handles.** Its Observations' targets and inbound subject-valued Observations re-map; R3's claim trigger covers the second hop.
8. **One accepted claim per Subject** is a partial unique index. Promote must hide or refuse subjects that are already members, and a step that loses a race fails alone — earlier steps stay saved.
9. **Composite FKs** carry `subject_type_id` on the claim. A person Subject cannot be claimed onto a Place; the target picker filters by type so the researcher never sees that error.
10. **Backfill is symmetric.** A confirmed pair pins both Observations on the incoming claim **and** the existing member's claim. The older claim's `argument` is not rewritten.
11. **Pins are Observations only.** Not Citations, Subjects, or Sources. Confirming a match creates no Observation.
12. **Stale reservation.** [`deleteimpact/reserved.go`](../../../core/database/deleteimpact/reserved.go) still names `sameness_claim_evidence` — superseded by `identity_claim_evidence`.
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
- [`catalog-deletes.md`](../../catalog-deletes.md): Identity Claim / evidence Impact; compute cache dependents before CASCADE.
- [`macos-client-patterns.md`](../../macos-client-patterns.md): Conclusion query keys and their invalidation rule; Go returns structures (DateValue, NameValue, title parts), Swift formats text.
- [`omnibar-search.md`](../archive/spike-3/omnibar-search.md): new kinds, location mapping, composed documents and header-dependent reprojection.

---

## Definition of done

- [ ] Dogfood bar items 1–9 met on real research
- [ ] Checklist complete (or items explicitly descoped)
- [ ] Design briefs archived under `design/archive/`
- [ ] Open questions answered and folded into the model docs
- [ ] [`docs/deployment-plan/README.md`](../README.md) points at the archive when closed
