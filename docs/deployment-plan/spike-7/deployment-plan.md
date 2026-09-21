# Deployment Plan — Spike 7

Citations, Observations, NameValue, Subject **fields** editor, citation composer place, and durable connect macros. Authoritative design: [`interpretation-graph-ui.md`](../../ideas/interpretation-graph-ui.md) §4–§6 / slices 3–7. Schema: [`interpretation-layer-data-model.md`](../../interpretation-layer-data-model.md), [`structured-name-model.md`](../../structured-name-model.md). Canvas: [Spike 6 archive](../archive/spike-6/). Navigation skills: [`add-workspace-location`](../../../.cursor/skills/add-workspace-location/SKILL.md), [`add-workspace-place`](../../../.cursor/skills/add-workspace-place/SKILL.md).

## Status

**Open.** Landings go in [`completed.md`](completed.md).

> **Goal of this spike:** prove the full evidence path — cite an Artifact portion, assert typed Observations on a subject, see them on the card, and make connect write real Citation-backed edges — without layering composer a11y onto the canvas.

> **Subject types stay product-seeded.** person / event / place / relationship / participation / location / source are first-class app kinds (palette, cards, macros), not a researcher-extensible CatalogVocabulary. **S7-D1 / S7-04 are descoped.**

> **Behavior lives in one registry.** Classification (root / bridge / reification), canvas placeability, required/locked bindings, and connect endpoint rules are declared next to the create-time seed — not hard-coded across graph, composer, and Subject fields UI. Future `plugin:<id>` types extend that same registry shape.

## Goal (dogfood bar)

All of the following must be true in the app **by spike close**. Build them as incremental dogfood slices (see [back half](#incremental-ui-dogfood-back-half)) — do not wait for the whole bar before testing in the app.

1. **Subject fields** replaces its stub with a real editor for Properties + bindings to **seeded** Subject types — **new IA for a long property list**, not a Source fields clone. Subject types destination stays stub / non-editable.
2. On the Evidence graph, a card has **Add property** → navigate to the **citation composer place** (Artifact pick as needed inside that place or as a short prelude).
3. Composer supports **images** (zoom/pan + region polygon) and **PDFs** (page nav + zoom/pan + region); audio/video deferred — *after* a thin text-only cite path already works.
4. One submit writes **one Citation + N Observations**; **Back** returns to the graph; card **grows** with cited property rows — *shippable with text Observations before viewers/locators*.
5. **NameValue** works end-to-end (schema → Go → reusable Swift editor per **S7-D5**, DateValue-shaped) — nested modal/sheet *inside* the composer place is fine; *late fill-in*.
6. **Connect** is durable: disambiguation on the graph → navigate to composer with two edge Observations pre-filled → submit → back to graph with a real bridge (replaces Spike 6 provisional links).
7. Empty Artifact gate is honest (cannot cite without an Artifact).
8. Composer has its **own accessibility tree** — not layered on the canvas.

## Observation value types (locked)

Product value types: **`text`**, **`integer`**, **`date`**, **`name`**, **`subject`**, **`term`**.

**Dropped:** `real` and `boolean` — no editors, not offered when creating Properties, no seed Properties use them. Align [`interpretation-layer-data-model.md`](../../interpretation-layer-data-model.md) in S7-01 / S7-01b / S7-03 so the schema does not invent unused columns.

| Type | Why |
| --- | --- |
| **text** | Prose Properties (`toponym`, `remark`, researcher-defined notes) |
| **term** | Kind/edge identity via **registry-only** Properties (`event_type`, `role`, `relationship_type`) → `property_terms` + `value_term_id` (**S7-01b**). Not offered when researchers create Properties. |
| **subject** | Bridge edges (`person`, `event`, `place`, `related_to`) |
| **date** | Event `date` / `start_date` / `end_date` (locked on event for Conclusion ordering); reuse DateValue |
| **name** | Primary person assertion; NameValue (`name_values` + `name_value_parts` only) |
| **integer** | Researcher-defined counts / ages (no seed Property yet) |

**Seed:** [`seeded-vocabulary.md`](../../seeded-vocabulary.md) §3.2–3.3 for person / event / place / participation / location / relationship. Defer `source` / `mentions` / `remark` UI. Kind/edge Properties use **Property terms** (large product sets + optional user terms via picker) — not free-text pickers.

## Interpretation subject registry (central source of truth)

Spike 5 already seeds Subject types from [`core/database/subjecttypes/registry.go`](../../../core/database/subjecttypes/registry.go). Spike 7 **must not** sprinkle “person is placeable,” “participation needs person+event,” or connect pair rules across Swift views and handlers.

**In S7-01**, grow a single declarative Interpretation vocabulary registry (prefer one package parallel to [`sourcevocab`](../../../core/database/sourcevocab/) — e.g. expand `subjecttypes` into / introduce `subjectvocab` that owns Install for types + properties + bindings). That registry is the **only** product definition of:

| Concern | Declared on | Used by |
| --- | --- | --- |
| Type identity | `key`, labels, ref prefixes, `origin=provenencia` | DB seed rows (as today) |
| **Role / capabilities** | e.g. root vs bridge vs reification; placeable on Evidence graph; create requires Citation | Palette, card chrome, connect entry, empty states |
| **Graph presentation** | Everything type-keyed for Evidence graph chrome: display name / `L10n` key, icon symbol token, card colors (ink / tint / chip / line tokens), edge/gradient tokens for roots **and** bridges | Palette, cards, ghosts, bridge cards, relationship **lines** |
| **Properties** | key, value_type, labels | `properties` seed + Subject fields + Observation editors |
| **Property terms** | term keys per Property (+ capabilities on recognized terms) | `property_terms` seed (**S7-01b**); composer term picker; birthday / tree / connect |
| **Bindings** | type_key → property_key (+ sort) | `subject_type_fields` seed; Add-property menus |
| **Required / locked bindings** | which seeded bindings macros need and UI must not unbind | Subject fields delete/unbind; connect pre-fill |
| **Connect matrix** | allowed endpoint pairs + which bridge type + which edge Properties / disambiguation fields | S7-10 macros (read registry; do not re-encode §3.2 in the view) |

**Type-keyed chrome:** Today’s `EvidencePrimaryKind` / `EvidenceSubjectKindStyle` (and edge gradients that sample primary ink) are the anti-pattern. The graph loads presentation from the registry for **every** Subject type it can show, then maps tokens to design-system colors/symbols locally. If it is configured *because of the type* (card name, palette label, wash, chip, line ink, gradient endpoints), it belongs in the registry. Platform tokens (`PVColor.…`, curated mark keys after **S7-12**, `L10n` keys) stay in the design system / catalogs; the registry stores **which token for which type**.

**Palette example (today’s Add Person / Event / Place):** those three tools are not a hard-coded Swift enum forever. The Evidence graph asks the registry (via Go/FFI): *which Subject types are placeable on this canvas, in what order?* For each, it renders a toggle from that type’s **presentation**. Adding or dropping a placeable type changes palette **and** card/line chrome via the registry, not via style `switch`es. Connect stays a separate chrome tool (not a Subject type), unless the registry later declares other non-type tools.

**Rules:**

1. **DB stays structural** — no requirement to add `kind` columns in Spike 7; capabilities + presentation tokens live in the registry (exported read-only to the client as needed).
2. **UI and FFI handlers look up by `(key, origin)` / registry helpers** — no scattered `if typeKey == "participation"` (or `EvidencePrimaryKind` exhaustive switches that invent the placeable set) across card chrome, composer, and connect sheet.
3. **Create-time Install only** — same `add-seeded-vocabulary` semantics as Source; open does not heal.
4. **Plugin seam** — later `plugin:<id>` contributes additional registry modules (or merged Install entries) with the **same capability + presentation fields**; researcher UI still does not invent Subject types. Document the registry shape in code comments / package README so the first plugin spike knows where to plug in.

Dogfood check for the registry: changing a capability, type presentation (name/icon/colors/line tokens), or locked binding in **one** place changes seed + app behavior without hunting call sites. After S7-01 lands the API, graph PRs (**S7-09** / earlier if needed) **migrate** Spike 6 palette, card kind style, and edge gradient styling off hard-coded primary kinds onto registry-driven presentation.

### Proposed registry shape (lock in S7-01)

Decide the **Go struct layout** in S7-01 (same spirit as `sourcevocab/registry.go`). This is the plugin contract later; tweak field names in implementation if needed, but keep the partitions below. **Not** persisted as JSON in SQLite in Spike 7 — Install writes only the structural catalog rows; the rest is read from the compiled registry (and exposed read-only over FFI as needed).

```text
subjectvocab/   (name flexible — may absorb today’s subjecttypes.Install)
  registry.go
    seedTypes[]       → subject_types rows + capabilities + presentation
    seedProperties[]  → properties rows
    seedBindings[]    → subject_type_fields rows (+ locked)
    seedTerms[]       → property_terms rows (+ term capabilities)   # S7-01b
    seedConnect[]     → connect macros (app-only; no table)
```

**`seedType`**

| Field | Purpose |
| --- | --- |
| `Key`, `Label`, `Description` | Catalog identity (seeded into `subject_types`) |
| `RefPrefix`, `CandidateRefPrefix` | Minting |
| `Role` | `root` \| `bridge` \| `reification` |
| `Placeable` | Appears on Evidence graph palette / click-to-place |
| `PaletteSort` | Order among placeables (`0…`; ignored if not placeable) |
| `RequiresCitationAtCreate` | Bridges: true; roots: false |
| `Presentation` | See below |

**`Presentation`** (string **tokens**, not raw colors)

| Field | Purpose |
| --- | --- |
| `L10nKey` | Client display name (palette + card); DB `Label` remains English seed fallback |
| `IconSymbol` | Curated mark key after **S7-12** (subject family in the shared pack); not SF Symbol forever |
| `InkToken`, `TintToken`, `ChipToken`, `LineToken` | Map to `PVColor` (or equivalent) in Swift — card wash, chip, stroke |
| `EdgeFromToken` / `EdgeToToken` | Optional; default edge gradients can derive from endpoint `InkToken` / `LineToken` if unset |

**`seedProperty`**

| Field | Purpose |
| --- | --- |
| `Key`, `Label`, `Description` | Catalog identity |
| `ValueType` | `text` \| `integer` \| `date` \| `name` \| `subject` \| `term` |

**`seedTerm`** (S7-01b)

| Field | Purpose |
| --- | --- |
| `PropertyKey`, `Key`, `Label`, `Description` | Catalog identity under that Property |
| `Capabilities` (optional) | e.g. birthday facet, tree-edge role — compiled registry, not SQL columns |

**`seedBinding`**

| Field | Purpose |
| --- | --- |
| `TypeKey`, `PropertyKey`, `SortOrder` | Seed `subject_type_fields` |
| `Locked` | Required for macros / product integrity (incl. event dates for Conclusion ordering) — Subject fields UI must not unbind or delete while locked |

**`seedConnect`** (one row per allowed endpoint pair)

| Field | Purpose |
| --- | --- |
| `FromTypeKey`, `ToTypeKey` | Ordered pair (or undirected flag if needed) |
| `BridgeTypeKey` | e.g. `participation`, `relationship`, `location` |
| `EdgePropertyKeys[]` | Observations to pre-fill (usually two) |
| `Disambiguation` | e.g. `none` \| `role` \| `relationship_type` (person→person is always relationship — no shared-event fork) |
| `Refuse` | If true, pair is explicitly illegal (optional; omit row = refuse by default) |

**Lookup API (package + FFI as needed):** `PlaceableTypes()`, `Presentation(key)`, `BindingsForType(key)`, `LockedBinding(type, property)`, `Connect(from, to)`. Graph / Subject fields / connect call these — never re-list the three primaries in Swift.

**Out of registry for Spike 7:** researcher-authored types; raw hex colors; full plugin manifest file format (same structs, different Install source, later).

## Composer presentation (locked: Option B)

**Navigate away** to a first-class workspace place. Canvas unloads; full-window side-by-side viewer|form; **Back** returns to the Evidence graph.

Rationale: keep graph a11y/keyboard from growing; composer gets a clean focus model and room for PDF/locator tools.

In-window modal over the graph and companion `NSWindow` are **out**.

### Navigation / breadcrumbs (defaults — confirm in S7-D4)

Composer is a **deep place under Sources**. Extend `WorkspaceLocation` (Spike 5 `sourceSurface` page|graph) with a composer discriminant plus context (`subjectId`, optional connect draft ids). Follow `add-workspace-location`.

**Breadcrumb shape (default — lock / revise in S7-D4):**

```text
Sources › {Source title} › Evidence graph › Citation for {scope}
```

Same crumb for Add property and Connect (`{scope}` = subject label, or bridge endpoints + edge phrase). No separate `Connect ›` segment.

**History policy (defaults — lock / revise in S7-D4):**

- Composer is a **first-class** history place (`go(to:)` / Back / Forward). No special stack surgery for Cancel or submit.
- **Cancel / Back** without submit writes nothing; the composer entry may remain in history.
- **Successful submit** navigates to the graph normally — do not pop/replace solely to scrub drafts.
- Persist across relaunch if subject still exists; else fallback to Evidence graph for that Source (same as other deep places).
- Form field values are UI ephemera (not restored from history JSON).
- Camera/selection stay out of the composer location.
- Connect **disambiguation** is a small sheet on the graph (not its own history entry). Connect-prefilled Observations must be reconstructible from the composer location payload.

Pinning a Citation across successive graph edits is **out** (one Citation + N Observations per submit).

## Design track (briefs)

**All UI is designed in Claude Design before the matching UI PR** (S7-D6 / S7-D7 / S7-D8 may be light handoffs rather than full surface boards). Briefs: [`design/`](design/).

| Step | Brief | Covers | Gates |
| --- | --- | --- | --- |
| **S7-D2** | Subject fields | Properties + bindings; create Property offers **five** types (**not** `term`); **few types (7) / many Properties** — creative IA, not Source fields; explore type cards etc. **No** Event types / Roles admin destinations | S7-05 |
| **S7-D6** | Curated marks | Unify into `Recipes/Marks/` (move evidence + add subject, incl. **source** from Subject fields); tint / size API; migrate graph **and** Subject fields | S7-12 |
| **S7-D8** | PVCallout actions | Optional actions slot on Callout (recovery CTA); no graph gate wiring | S7-14 |
| **S7-D3** | Evidence graph updates | Add-property; cited-data rows; subject refs; bridge edge summaries; Artifact gate; connect disambiguation → composer handoff | S7-09, S7-10 |
| **S7-D4** | Citation composer place | Full-window viewer\|form; Artifact pick; locators; observation list; DateValue reuse; breadcrumbs; composer-only a11y — **hosts** NameValue modal, does not design it | S7-08 |
| **S7-D5** | NameValue editor | Reusable NameValue modal (DateValue twin); form + optional parts | S7-02b |
| **S7-D7** | Card component | Claude Design **Card** reference from shipped `PVCard`; tones / border / elevation; not graph snowflake cards | S7-13 |

~~**S7-D1** Subject types editor~~ — **descoped** (see [Descoped](#descoped) below).

**Scheduling:** **S7-D2** early (unblocks S7-05). **S7-D6** before mark consolidation (**S7-12**). **S7-D8** → **S7-14** after **S7-12**, before Add-property (**S7-09**). **S7-D3** before **S7-09**. **S7-D4** before thin composer (**S7-08**). **S7-D5** before NameValue Swift (**S7-02b**), late in the composer fill-in. Prefer **D6 → 12 → D8 → 14 → D3 → 09** so kit Callout actions exist before the graph message center. **S7-D7 → S7-13** late (after **S7-10**, before dogfood close).

## Incremental UI dogfood (back half)

Do **not** stack viewers → locators → NameValue → composer → Add property. That forces building five PRs before anything is clickable in the app.

**Invert:** ship a thin vertical path you can click immediately, then thicken the composer.

```text
S7-12  → Recipes/Marks/ (move evidence + subject_*; tint API; graph + Subject fields)
         dogfood: cards / palette / fields strip at zoom; source mark present; type_* hosts still work
S7-14  PVCallout actions slot (kit Extend)
         dogfood: preview / any demo callout with a button; existing callouts unchanged
S7-09  Add property on cards → navigate to composer place (stub/shell OK)
         dogfood: button, place, breadcrumbs, Back; No-Artifact callout can use actions
S7-08  Thin composer: Artifact pick + citation + text Observations + submit
         dogfood: cite a line of text, card grows (no fancy viewer yet)
S7-06  Image + PDF viewers in the composer
S7-07  Locator tools (page + region)
S7-02b NameValue editor hosted in composer
S7-10  Durable connect macros
S7-D7  Card component (design-system reference from shipped PVCard)
S7-13  Migrate manual card cousins → PVCard
S7-11  Full dogfood bar / close
```

Each step is independently testable in the running app against **S7-03** (and S7-01 / S7-01b / S7-02 / **S7-12** / **S7-14** as needed).

## PR sequence

```text
design                              build
─────────────                       ─────────────────────────────────────────

S7-D2 Subject fields                S7-01  properties + subject_type_fields
  │                                 │      + **central subject registry**
  │                                 ▼
  │                                 S7-01b Property terms (`property_terms`,
  │                                 │      value_type=term, seed terms,
  │                                 │      migrate kind/edge Properties)
  │                                 ▼
  └────── D2 gates ───────────────▶ S7-05  Subject fields UI

                                    S7-02  NameValue schema + Go
                                    │      (parallel; before Observations)
                                    ▼
                                    S7-03  citations + observations + locator
                                    │      validation (Go) + FFI macros
                                    │      (includes value_term_id)
                                    │
S7-D6 Curated marks                 │
  │                                 ▼
  └────── D6 gates ───────────────▶ S7-12 → Recipes/Marks/
                                    │      (move EvidenceIcon + subject_*;
                                    │       graph + Subject fields migrated)
                                    │
S7-D8 PVCallout actions             │
  │                                 ▼
  └────── D8 gates ───────────────▶ S7-14  PVCallout actions slot
                                    │      (kit Extend; no graph gate yet)
                                    │
S7-D3 Graph updates                 │
  │                                 ▼
  └────── D3 gates ───────────────▶ S7-09  Add property + composer navigation
                                    │      (+ card chrome / cited-row slots;
                                    │       composer may be a thin stub)
                                    │      dogfood: click Add property → place
                                    │
S7-D4 Composer place                │
  │                                 ▼
  └────── D4 gates ───────────────▶ S7-08  Thin composer (submit path)
                                    │      Artifact pick, citation fields,
                                    │      text + term Observations, submit → card grows
                                    │      Viewer pane: placeholder OK
                                    │
                                    S7-06  Image + PDF viewers → composer
                                    │
                                    ▼
                                    S7-07  Locator tools (page + region)
                                    │
S7-D5 NameValue editor              │
  │                                 ▼
  └────── D5 gates ───────────────▶ S7-02b NameValue Swift editor → composer
                                    │
                                    ▼
                                  S7-10  Durable connect macros
                                    │
S7-D7 Card component                │
  │                                 ▼
  └────── D7 gates ───────────────▶ S7-13  PVCard call-site cleanup
                                    │      (four manual cousins → PVCard)
                                    │
                                    ▼
                                  S7-11  Dogfood close / docs
```

**Dependency notes:**

- **S7-01b** → **S7-01**. Lands **before** S7-05 and **before** S7-03; introduces kind/edge Properties as `term` (they are not seeded as text in S7-01).
- **S7-05** → **S7-01** + **S7-01b** + **S7-D2** only (no NameValue UI).
- **S7-12** → **S7-D6** only (design-system consolidation). **Not** related to Citations/Observations work. Schedule after schema/Go (**S7-03**) and **before** S7-09 so graph chrome thickens on one mark pipeline.
- **S7-14** → **S7-D8**. Kit-only Callout **Extend** (actions slot). After **S7-12**, **before** **S7-09** so the No-Artifact message center can compose a recovery CTA without a hand-rolled banner.
- **S7-09** → **S7-03** (graph payload can show Observations) + **S7-12** + **S7-14** + **S7-D3**. Registers composer `WorkspaceLocation`; destination may stub until S7-08.
- **S7-08** → **S7-09** + **S7-D4** + **S7-03**. **Does not** require S7-06/07/02b — text/term Observations and a placeholder viewer are enough to dogfood submit + card growth.
- **S7-06 / S7-07 / S7-02b** fill the composer in place; each is dogfoodable on top of S7-08.
- **S7-10** → working composer submit (S7-08+) + **S7-D3**.
- **S7-13** → **S7-D7**. Late hygiene after connect; **before** S7-11. Does **not** rewrite Evidence graph snowflake cards.

Schema/Go (01–03, 01b) may start before design finishes; **UI PRs gate on the matching brief.**

---

## Checklist

- [x] S7-D2 — Design: Subject fields → [`completed.md`](completed.md)
- [x] S7-D3 — Design: Evidence graph updates → [`completed.md`](completed.md)
- [ ] S7-D4 — Design: Citation composer place → [`completed.md`](completed.md)
- [ ] S7-D5 — Design: NameValue editor → [`completed.md`](completed.md)
- [x] S7-01 — `properties` + `subject_type_fields` + **central Interpretation subject registry** + Go/FFI → [`completed.md`](completed.md)
- [x] S7-01b — Property terms (`property_terms`, `value_type=term`, kind/edge seed) → [`completed.md`](completed.md)
- [x] S7-05 — Subject fields UI → [`completed.md`](completed.md)
- [x] S7-02 — NameValue schema + Go → [`completed.md`](completed.md)
- [x] S7-03 — Citations + Observations + locator validation + FFI → [`completed.md`](completed.md)
- [x] S7-D6 — Design: Curated marks (subject → evidence icon pack) → [`completed.md`](completed.md)
- [x] S7-12 — `Recipes/Marks/` consolidation + graph / Subject fields migration → [`completed.md`](completed.md)
- [x] S7-D8 — Design: PVCallout actions slot → [`completed.md`](completed.md)
- [x] S7-14 — `PVCallout` actions slot (kit Extend) → [`completed.md`](completed.md)
- [x] S7-09 — Add property + composer navigation (stub OK) → [`completed.md`](completed.md)
- [ ] S7-08 — Thin composer (submit + card growth; viewer placeholder OK) → [`completed.md`](completed.md)
- [ ] S7-06 — Image + PDF viewers in composer → [`completed.md`](completed.md)
- [ ] S7-07 — Locator tools (page + region) → [`completed.md`](completed.md)
- [ ] S7-02b — NameValue Swift editor → [`completed.md`](completed.md)
- [ ] S7-10 — Durable connect macros → [`completed.md`](completed.md)
- [ ] S7-D7 — Design: Card component (design-system reference) → [`completed.md`](completed.md)
- [ ] S7-13 — `PVCard` call-site cleanup (manual cousins) → [`completed.md`](completed.md)
- [ ] S7-11 — Dogfood close / docs → [`completed.md`](completed.md)

## Descoped

| Step | Notes |
| --- | --- |
| **S7-D1** / **S7-04** | Subject types CatalogVocabulary editor. Types remain Spike 5 seeded rows + first-class graph behavior. Brief archived: [`design/archive/S7-D1-subject-types.md`](design/archive/S7-D1-subject-types.md). |

---

## S7-D2 — Design: Subject fields

Claude Design board for Subject fields. Brief archived: [`design/archive/S7-D2-subject-fields.md`](design/archive/S7-D2-subject-fields.md). Gates **S7-05**. Design around **seven fixed types** (non-list type chrome welcome) and **many Properties**. Source fields layout explicitly out. Board must treat **`term` as registry-only** (visible on seeded Properties, not in create Property) and must **not** invent Event types / Roles destinations. Addendum archived: [`design/archive/S7-D2-subject-fields-addendum-property-terms.md`](design/archive/S7-D2-subject-fields-addendum-property-terms.md).

---

## S7-D6 — Design: Curated marks

Handoff / light board for the unified **Marks** recipe (`DesignSystem/Recipes/Marks/` — move evidence icons + add subject marks). Brief: [`design/archive/S7-D6-curated-marks.md`](design/archive/S7-D6-curated-marks.md) (Claude Design board links for Evidence graph + Subject fields artwork). Gates **S7-12**. Pulls the **source** folio from Subject fields (S7-D2) into the same set as S6 graph marks. Locks tint/size API (template assets + call-site `.foregroundStyle`). Includes **UI building-block inventory** (§7: New/Extend **Marks**, Extend graph + Subject fields call sites, Retire `EvidenceIcon` + `SubjectIcon`). Does **not** redesign graph cards (**S7-D3**) or Subject fields IA.

---

## S7-D3 — Design: Evidence graph updates

Claude Design board for Add property, cited rows, subject refs, bridge edge summaries, connect disambiguation handoff. Brief archived: [`design/archive/S7-D3-evidence-graph-updates.md`](design/archive/S7-D3-evidence-graph-updates.md). Gates **S7-09**, **S7-10**. Does **not** design the composer place (S7-D4). Prefer **S7-12** and **S7-14** (`PVCallout` actions) already landed. **Implement S7-09 before the thick composer** so Add property is dogfoodable early.

---

## S7-D4 — Design: Citation composer place

Claude Design board for the navigable composer. Brief: [`design/S7-D4-citation-composer.md`](design/S7-D4-citation-composer.md). Gates **S7-08**. Confirms breadcrumbs and history policy. **Phase the board:** shell + form first (text Observations); viewer/locator/NameValue as later fill-ins. Does **not** design the NameValue editor (S7-D5).

---

## S7-D5 — Design: NameValue editor

Claude Design board for the reusable NameValue modal (DateValue twin). Brief: [`design/S7-D5-name-value-editor.md`](design/S7-D5-name-value-editor.md). Gates **S7-02b**. Schedule late — after thin composer works.

---

## S7-D8 — Design: PVCallout actions

Light Claude Design board / kit handoff for an optional **actions** slot on Callout. Brief archived: [`design/archive/S7-D8-pvcallout-actions.md`](design/archive/S7-D8-pvcallout-actions.md). Gates **S7-14**. Unblocks the Evidence graph No-Artifact message center recovery CTA without a parallel banner. Schedule after **S7-12**, before **S7-09**. Does **not** wire the graph gate (that’s **S7-D3** / **S7-09**).

---

## S7-D7 — Design: Card component

Claude Design **Card** kit page / board, referenced from the already-shipped macOS [`PVCard`](../../../macos/App/DesignSystem/Components/Card/PVCard.swift). Brief: [`design/S7-D7-card-component.md`](design/S7-D7-card-component.md). Gates **S7-13**. Document tones, solid/dashed border, elevation, and radius; use implemented Artifacts / Subject fields / metadata cards as the visual source of truth. Does **not** redesign Evidence graph subject/bridge cards (those stay snowflakes under **S7-D3**). Schedule after **S7-10**, before dogfood close.

---

## S7-01 — Properties + subject_type_fields + seed

Migration(s) for `properties`, `subject_type_fields`; `value_type` limited to text / integer / date / name / subject; create-time seed via `add-seeded-vocabulary`; audited CRUD + FFI; list bindings for Add-property menus. Update interpretation-layer model docs to drop `real` / `boolean`. Bindings reference existing Spike 5 `subject_types` rows only.

**Registry (required):** Centralize Interpretation subject vocabulary in one declarative registry (expand `subjecttypes` and/or add `subjectvocab` parallel to `sourcevocab`) that:

- Seeds Subject types (move/keep today’s seven rows here), Properties, and `subject_type_fields`
- Declares **capabilities** per type (root / bridge / reification, canvas placeable, citation-required-at-create, …)
- Declares **graph presentation** per Subject type (placeable roots **and** bridges): display name/`L10n` key, icon token, card color tokens, line/gradient tokens — anything type-keyed on the Evidence graph
- Declares **required/locked** bindings for bridge macros
- Declares the **connect matrix** (endpoint pairs → bridge type → edge Properties / disambiguation fields)
- Exposes lookup helpers / FFI so graph, Subject fields, composer, and connect **do not hard-code type keys, the placeable set, or kind→chrome maps**

Wire Install into `onboarding.createCatalog` only. Follow [`.cursor/skills/add-seeded-vocabulary`](../../../.cursor/skills/add-seeded-vocabulary/SKILL.md); update that skill’s domain table when the package lands.

| | |
| --- | --- |
| **In** | Tables; central registry + Install; Go packages; FFI list/create/update/delete (unused Properties); bindings query; capability/presentation/connect lookup API (incl. “list placeable types for palette” and “presentation for type key”). |
| **Out** | Subject fields UI (S7-05); Observations; Subject types user CRUD; scattering capability checks, palette membership, or kind→color/gradient maps in Swift views. Property terms (**S7-01b**). |
| **Testable** | Create project seeds §3.2–3.3 bindings; registry tests for capabilities, full presentation tokens, locked bindings, placeable set; Go tests; FakeStore round-trip. |
| **Depends on** | Spike 5 `subject_types` table (may fold Install into the new registry package). **Not** gated on design. |

---

## S7-01b — Property terms

Land **immediately after S7-01**, before Subject fields UI and before Observations write kind/edge values as free text.

**Problem:** Free-text `event_type` / `role` / `relationship_type` lets researchers invent synonyms (`birth` vs `birthday` vs `DOB`); UI facets (birthday, family tree, connect disambiguation) and Conclusion type-identity then miss. Tiny enum + `other` + free-text name fails the same way for sameness. Graph `subjects.label` stays working identity only — not type identity.

**Model:** New `value_type = term` + `property_terms` table (origin-namespaced vocabulary rows). Observations (S7-03) store `value_term_id`. **Introduce** kind/edge Properties (`event_type`, `role`, `relationship_type`) here as `term` — they are **not** seeded as `text` in S7-01. Install seeds those Properties + bindings + **large** product term sets ([`seeded-vocabulary.md`](../../seeded-vocabulary.md) §3.4–3.6). Term capabilities (birthday, tree-edge, …) are **deferred** to a later discussion/PR. **`term` Properties are registry-only** — Create Property for `origin=user` refuses `value_type = term` (Subject fields offers the other five types only). Researchers may mint `origin=user` **term rows** under those registry Properties via composer picker **Add custom…** (rename/delete when unused) — **no** Event types / Roles CatalogVocabulary destinations. Plugins that add bridge or kind subjects contribute term-typed Properties + term sets through the same `subjectvocab` registry shape.

Authoritative schema notes: [`interpretation-layer-data-model.md`](../../interpretation-layer-data-model.md) §5.1.1. Design decision: [`interpretation-graph-ui.md`](../../ideas/interpretation-graph-ui.md) #23.

| | |
| --- | --- |
| **In** | Migration: `property_terms`; extend `properties.value_type` CHECK with `term`; Go package + audited CRUD for terms; registry `seedProperties`/`seedBindings`/`seedTerms` for kind/edge (term capabilities deferred); FFI list/create/update/delete (user **term rows**; product terms locked); Create Property refuses `term` for user origin; docs/skill updates. |
| **Out** | Observation writers (S7-03); composer term picker UI (S7-08 / S7-D4); Subject fields layout (S7-05); Event types / Roles sidebar destinations; researcher-created term-typed Properties. |
| **Testable** | New project seeds term-typed kind/edge Properties + term rows; Lookup by `(property, key, origin)`; refuse delete while in use; product terms not user-editable; user Create Property with `term` refused; registry capability helpers. |
| **Depends on** | **S7-01**. **Not** gated on design (S7-D2 must not offer `term` in create Property). |

---

## S7-02 — NameValue schema + Go

`name_values` / `name_value_parts` per structured-name-model §2–3 (Interpretation only — not Conclusion `name_format`); `core/database/namevalues`. **No Swift UI** in this PR.

Lives on the **Observations branch**, not the Subject fields branch. Needed so **S7-03** can FK `value_name_id`. Does **not** block **S7-05**.

| | |
| --- | --- |
| **In** | Schema, Go package, tests. |
| **Out** | Swift editor (S7-02b); composer wiring (S7-08); name_format profiles; Subject fields UI. |
| **Testable** | Go round-trip create/read parts. |
| **Depends on** | — (parallel with S7-05 after S7-01). **Not** gated on S7-D5. |

---

## S7-03 — Citations + Observations + locator validation

`citations`, `citation_notes`, `observations`, `observation_notes`; value columns for the six types (including **`value_term_id`**); Go locator validate (`page`, `region`; optional `text_quote`); atomic “create Citation + Observations” RPC; graph payload includes Observations for card rows.

| | |
| --- | --- |
| **In** | Migrations, Go packages, locator invariants, FFI macros, graph query enrichment. |
| **Out** | Composer UI; card chrome. |
| **Testable** | Go locator reject/accept; atomic create (text + term); FakeStore. |
| **Depends on** | S7-01, **S7-01b**, S7-02 (name column / FK). |

---

## S7-12 — Curated marks consolidation

Move subject type marks (`PVSubjectIcon` Canvas paths) **and** today’s evidence icons into one asset-backed recipe: **`DesignSystem/Recipes/Marks/`** (`file_*` + `type_*` + `subject_*`). Rename/move `Recipes/EvidenceIcon/` + `Assets.xcassets/EvidenceIcons/` into Marks in the same PR. Template-rendered SVGs; document tint API (call-site `.foregroundStyle` / ambient foreground — colors not baked into assets).

**Artwork:** export S6 graph marks **and** the **source** folio from Subject fields (S7-D2 / `SourceMark`) — board links in [`design/archive/S7-D6-curated-marks.md`](design/archive/S7-D6-curated-marks.md).

**Call sites:** migrate **every** `PVSubjectIcon` / `PVSubjectIconKind` use — Evidence graph (cards, palette, bridges) **and** Subject fields type strip / binding chrome — onto Marks; update existing `PVEvidenceIcon` hosts to the new home; then retire `EvidenceIcon` + `SubjectIcon`. Align registry presentation icon tokens with curated mark keys so **S7-09** does not invent a second icon channel.

| | |
| --- | --- |
| **In** | `Recipes/Marks/` (+ asset catalog); seven `subject_*` glyphs incl. **source**; key enum / View API; tint + size docs (`MARKS.md`); graph **+ Subject fields** + existing evidence-icon call-site migration; retire `EvidenceIcon` / `SubjectIcon`; DesignSystem README update. |
| **Out** | Card Add-property / cited-row UX (**S7-09**); Subject fields IA changes; researcher-editable subject icons; new mark metaphors beyond today’s seven kinds; keeping a permanent `EvidenceIcon` folder. |
| **Testable** | Palette + cards show subject marks at zoom; Subject fields type strip still shows all kinds incl. source; Sources/type icons still tint; `rg` clean of old recipe paths (or only deprecated shim). |
| **Depends on** | **S7-D6**. Schedule after **S7-03**, before **S7-14** / **S7-09** — not gated on Citations/Observations code. |

---

## S7-14 — PVCallout actions slot

Extend [`PVCallout`](../../../macos/App/DesignSystem/Components/Callout/PVCallout.swift) with an optional actions slot (match **S7-D8** / web Callout `actions`). Existing text-only call sites must keep compiling and looking the same.

| | |
| --- | --- |
| **In** | Optional `@ViewBuilder` (or equivalent) actions on `PVCallout`; preview with button(s); DesignSystem README Callout row; follow `add-ui-component` Extend rules. |
| **Out** | Evidence graph No-Artifact gate / disable-controls (**S7-09**); new banner/message-center component; dismiss/`detail`/`plain` unless D8 locked them. |
| **Testable** | Preview shows callout + action; an existing call site (e.g. Subject fields locked note) unchanged; SwiftUI build green. |
| **Depends on** | **S7-D8**. Schedule after **S7-12**, before **S7-09**. |

---

## S7-05 — Subject fields UI

Properties + `subject_type_fields` bindings; researcher create offers **five** value_types (`text` / `integer` / `date` / `name` / `subject`). Seeded **`term`** Properties from the registry appear in the list/bindings like any other Property but are not creatable here. Implement the **S7-D2** IA. Do **not** invent Event types / Roles destinations — user **term rows** under registry term Properties are composer-local (S7-D4).

| | |
| --- | --- |
| **In** | Browse/create/edit at scale; search/filter; bind to seeded Subject types; respect registry **locked** bindings; create Property with five researcher value_types only. |
| **Out** | Composer; NameValue tables/editor; Observation editors; Subject types CRUD; Source-fields layout reuse; Event types / Roles admin; offering `term` in create Property. |
| **Testable** | Find a Property in a long list; create a `name`-typed Property; bind a field; locked binding cannot be removed; create UI has no `term` option. |
| **Depends on** | S7-01, **S7-01b**, **S7-D2** — **not** S7-02 / S7-02b. |

---

## S7-09 — Add property + composer navigation

Grow `EvidenceSubjectCardChrome`; Add property → `go(to: composer)`; cited-row **slots** / growth chrome; subject **refs**; uncited → cited shell when Observations exist; edge layout heights; registry-driven palette/presentation migration; graph a11y for new controls.

**Composer destination may be a stub** (“form next”) as long as navigation, breadcrumbs, and Back work. Real submit lands in S7-08.

| | |
| --- | --- |
| **In** | Add property control; composer `WorkspaceLocation`; card chrome for cited rows + **refs**; Artifact gate messaging (graph-wide `PVCallout` **with actions** from **S7-14**); slight card widen if D3 locks it; registry presentation for kinds (icons via **S7-12** pack). |
| **Out** | Full composer form (S7-08); connect (S7-10); durable bridge edge-summary phrases (S7-10). |
| **Testable** | Select a Person → Add property → land on composer place → Back to graph. Ref visible on card. No submit required yet. |
| **Depends on** | S7-03 (optional empty cited rows), **S7-12**, **S7-14**, **S7-D3**. |

---

## S7-08 — Thin citation composer (submit path)

New workspace place + location discriminant; breadcrumb per S7-D4; Artifact pick; citation fields; N Observations with **text** (and other simple types if cheap); single submit; cancel/Back; composer-only a11y.

**Intentionally thin:** left pane may be a **placeholder** (“viewer in S7-06”) or Artifact title/metadata only. **No** requirement for PDF/image viewers, locators, or NameValue in this PR — those land as follow-ons so this PR is dogfoodable as soon as Add property exists.

| | |
| --- | --- |
| **In** | Navigable place; form; submit Citation + Observations via S7-03; card growth when returning to graph; DateValue reuse if needed. |
| **Out** | Image/PDF viewers (S7-06); locators (S7-07); NameValue modal (S7-02b); connect macros (S7-10). |
| **Testable** | Add property → composer → pick Artifact → add text Observation → submit → Back → card shows cited row. |
| **Depends on** | S7-09, S7-03, **S7-D4**. |

---

## S7-06 — Artifact viewers (image + PDF)

PDFKit + image overlay; resolve files via `ProjectFiles`; no QuickLook for composer (no locator API). **Plug into the existing composer** from S7-08 (replace placeholder).

| | |
| --- | --- |
| **In** | Zoom/pan image; PDF page nav + zoom/pan; left pane of composer. |
| **Out** | Locator drawing (S7-07); audio/video. |
| **Testable** | Open composer on an image/PDF Artifact — see viewer; submit still works. |
| **Depends on** | S7-08. |

---

## S7-07 — Locator tools (page + region)

Interactive tools feeding nested selectors; Go is source of truth for invariants (≥3 points, non-self-intersecting, etc.).

| | |
| --- | --- |
| **In** | Page selector UI; polygon region draw on image/PDF page; produce `locator_json`. |
| **Out** | `time_range`; required `text_quote` (optional if cheap). |
| **Testable** | Draw region in composer → submit → Citation stores validated locator. |
| **Depends on** | S7-06, S7-03 validation. |

---

## S7-02b — NameValue Swift editor

Swift `NameValueDraft` / editor modal under `Features/Names/` (mirror `Features/Dates/`). Host in the composer for `value_type = name`.

**Late fill-in** after thin composer works. Not required for S7-05 or S7-08/09.

| | |
| --- | --- |
| **In** | Reusable modal UI per S7-D5; unit tests for draft validation; wire into S7-08 observation rows. |
| **Out** | Schema (S7-02); name_format profiles. |
| **Testable** | In composer, add a name Observation → edit NameValue → submit. |
| **Depends on** | S7-02, S7-08, **S7-D5**. |

---

## S7-10 — Durable connect macros

Disambiguation sheet on graph (§3.2 matrix); navigate to composer with pre-filled edge Observations; replace provisional Spike 6 links; bridge cards show **edge summary** phrases (S7-D3 §3.1).

| | |
| --- | --- |
| **In** | person→event `role`; person→person always `relationship` + `relationship_type` (no shared-event fork — shared events = Event bubble + person→event lines); event→place clean; refuse unsupported; atomic bridge + Citation + edges — **driven by the S7-01 registry connect matrix**; bridge body summaries (“is the {type} of”, “took place in”, role-aware participation); **L10n** for product property terms + phrase templates (DB `label` is English seed / fallback only). |
| **Out** | Pinning; full conflicted/negated chrome; inventing connect rules outside the registry; translating user-minted term labels. |
| **Testable** | Connect two people → disambiguate → cite (thin or thick composer) → bridge persists with readable **localized** summary; relaunch keeps edges. |
| **Depends on** | S7-08 (submit path), S7-09, **S7-D3**, S7-01 registry. |

---

## S7-13 — PVCard call-site cleanup

After **S7-D7**, migrate the remaining **manual card cousins** onto shared `PVCard` so fill + clip + hairline (+ dashed / sunken / elevation where needed) stop drifting.

**In scope (four call sites):**

| Call site | Today | Target |
| --- | --- | --- |
| [`OnboardingIdentifyView`](../../../macos/App/Features/Onboarding/OnboardingIdentifyView.swift) | Hand-built `surfaceCard` + `md` + `borderSubtle` | `PVCard` |
| [`OnboardingProjectMetaLines`](../../../macos/App/Features/Onboarding/OnboardingProjectMetaLines.swift) | Same pattern | `PVCard` |
| [`SourceTypesDetailPane`](../../../macos/App/Features/SourceTypes/SourceTypesDetailPane.swift) (icon / tile chrome ~line 411) | Same with `sm` radius | `PVCard(cornerRadius: .sm)` (or equivalent) |
| [`OnboardingOpenPicker`](../../../macos/App/Features/Onboarding/OnboardingOpenPicker.swift) (empty-folder block) | Sunken + dashed | `PVCard(tone: .sunken, border: .dashed)` |

| | |
| --- | --- |
| **In** | Those four migrations; any small `PVCard` API tweak required by **S7-D7**; keep DesignSystem README Card row accurate. |
| **Out** | Evidence graph subject/bridge/palette chrome; Subject fields type-strip tiles; file-choice / icon-picker selected cells; full-bleed pane fills; omnibar `overlay` elevation (unless D7 explicitly unifies it). |
| **Testable** | Onboarding identify + open-picker + project meta, and Source types detail tile, still look correct; `rg` for the old hand-built patterns at those sites is gone. |
| **Depends on** | **S7-D7**. Schedule after **S7-10**, before **S7-11**. |

---

## S7-11 — Dogfood close / docs

Honesty pass against the [goal bar](#goal-dogfood-bar), including viewers/locators/NameValue/connect. Record in [`completed.md`](completed.md); update [`README.md`](README.md); archive design briefs; point [`docs/deployment-plan/README.md`](../README.md) at archive when closed; SemVer only if cutting a product release.

---

## Scope boundary

| In | Out |
| --- | --- |
| Subject **fields** editor | Source-page `mentions` / `remark` |
| Composer as **navigable place** (Option B) | In-window modal (A); companion window (C) |
| Image + PDF + page/region | Audio / video / QuickLook-as-composer |
| NameValue schema + reusable editor (S7-D5 stream) | Conclusion name_format |
| Durable connect via composer | Citation pinning across graph edits |
| Card cited-property rows | Unplaced tray, minimap, auto-layout |
| Five→six value-type editors (incl. term picker) | `real` / `boolean`; full conflicted/negated visual language |
| Seeded Subject types (Spike 5) | **Subject types** CatalogVocabulary / user-defined types |
| Property terms (S7-01b) | Event types / Roles admin destinations; free-text kind/edge values |

---

## Gotchas (read before coding)

1. **Composer a11y is its own place** — do not bolt VoiceOver for the form onto the canvas representation (§7.4 stays about the graph).
2. **Locator JSON validates in Go** — UI must not write malformed selectors.
3. **Bridge subjects never persist alone** — connect is atomic with Citation + edge Observations (design note §4.2).
4. **`subject_type_id` is immutable** — type pickers at create only for Subjects; Properties' `value_type` likewise immutable after create.
5. **No Artifact → no Observation** — empty state must route to add a file, not a broken composer.
6. **Provisional Spike 6 links** must be replaced or clearly migrated; do not leave honesty labels as permanent UI.
7. **NameValue ≠ transcription** — transcription stays on the Citation; NameValue is the Observation normalization.
8. **Drop `real` / `boolean` in model docs** when shipping S7-01/S7-03 so product and schema stay aligned.
9. **No sprinkled type keys** — placeability, palette membership, and **all type-keyed chrome** (card names, icons, colors, line/gradient tokens), bridge vs root, locked bindings, and connect pairs come from the S7-01 Interpretation subject registry. Views resolve tokens; they do not own `EvidencePrimaryKind` / style switch maps.
10. **Plugin path is the registry** — when plugins arrive, they extend Install/registry modules with the same capability fields (including **Property term** sets); do not invent a second configuration channel.
11. **Kind/edge values are Property terms** — do not write free-text Observations for `event_type` / `role` / `relationship_type` after S7-01b; do not use `subjects.label` as type identity; do not let researchers create Properties with `value_type = term`.

---

## Definition of done

- [ ] Checklist above complete
- [ ] Dogfood bar items 1–8 met
- [ ] Design briefs archived under `design/archive/`
- [ ] [`docs/deployment-plan/README.md`](../README.md) points at archive when closed
- [ ] Design note status line points at Spike 7 archive (slices 3–7 advanced)

## What the next spike inherits

On success: a cited Evidence graph — Subject fields configurable against seeded types, properties visible on cards, connect writes real evidence, NameValue available for person names. Ready for honesty/polish (conflicted/negated), Source-page commentary, and media types beyond image/PDF.
