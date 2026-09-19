# Deployment Plan — Spike 7

Citations, Observations, NameValue, Subject **fields** editor, citation composer place, and durable connect macros. Authoritative design: [`interpretation-graph-ui.md`](../../ideas/interpretation-graph-ui.md) §4–§6 / slices 3–7. Schema: [`interpretation-layer-data-model.md`](../../interpretation-layer-data-model.md), [`structured-name-model.md`](../../structured-name-model.md). Canvas: [Spike 6 archive](../archive/spike-6/). Navigation skills: [`add-workspace-location`](../../../.cursor/skills/add-workspace-location/SKILL.md), [`add-workspace-place`](../../../.cursor/skills/add-workspace-place/SKILL.md).

## Status

**Open.** Landings go in [`completed.md`](completed.md).

> **Goal of this spike:** prove the full evidence path — cite an Artifact portion, assert typed Observations on a subject, see them on the card, and make connect write real Citation-backed edges — without layering composer a11y onto the canvas.

> **Subject types stay product-seeded.** person / event / place / relationship / participation / location / source are first-class app kinds (palette, cards, macros), not a researcher-extensible CatalogVocabulary. **S7-D1 / S7-04 are descoped.**

> **Behavior lives in one registry.** Classification (root / bridge / reification), canvas placeability, required/locked bindings, and connect endpoint rules are declared next to the create-time seed — not hard-coded across graph, composer, and Subject fields UI. Future `plugin:<id>` types extend that same registry shape.

## Goal (dogfood bar)

All of the following must be true in the app:

1. **Subject fields** replaces its stub with a real CatalogVocabulary editor (Properties + bindings to the **seeded** Subject types). Subject types destination stays stub / non-editable.
2. On the Evidence graph, a card has **Add property** → navigate to the **citation composer place** (Artifact pick as needed inside that place or as a short prelude).
3. Composer supports **images** (zoom/pan + region polygon) and **PDFs** (page nav + zoom/pan + region); audio/video deferred.
4. One submit writes **one Citation + N Observations**; **Back** returns to the graph; card **grows** with cited property rows.
5. **NameValue** works end-to-end (schema → Go → reusable Swift editor per **S7-D5**, DateValue-shaped) — nested modal/sheet *inside* the composer place is fine.
6. **Connect** is durable: disambiguation on the graph → navigate to composer with two edge Observations pre-filled → submit → back to graph with a real bridge (replaces Spike 6 provisional links).
7. Empty Artifact gate is honest (cannot cite without an Artifact).
8. Composer has its **own accessibility tree** — not layered on the canvas.

## Observation value types (locked)

Product value types: **`text`**, **`integer`**, **`date`**, **`name`**, **`subject`**.

**Dropped:** `real` and `boolean` — no editors, not offered when creating Properties, no seed Properties use them. Align [`interpretation-layer-data-model.md`](../../interpretation-layer-data-model.md) in S7-01 / S7-03 so the schema does not invent unused columns.

| Type | Why |
| --- | --- |
| **text** | Most seeded Properties (`occupation`, `event_type`, `role`, `relationship_type`, `toponym`, …) |
| **subject** | Bridge edges (`person`, `event`, `place`, `participant`) |
| **date** | `birth_date`, `date`; reuse DateValue |
| **name** | Primary person assertion; NameValue (`name_values` + `name_value_parts` only) |
| **integer** | `age_at_event` |

**Seed:** [`seeded-vocabulary.md`](../../seeded-vocabulary.md) §3.2–3.3 for person / event / place / participation / location / relationship. Defer `source` / `mentions` / `remark` UI. Open pickers for `event_type` / `role`; `relationship_type` open text (± short starter).

## Interpretation subject registry (central source of truth)

Spike 5 already seeds Subject types from [`core/database/subjecttypes/registry.go`](../../../core/database/subjecttypes/registry.go). Spike 7 **must not** sprinkle “person is placeable,” “participation needs person+event,” or connect pair rules across Swift views and handlers.

**In S7-01**, grow a single declarative Interpretation vocabulary registry (prefer one package parallel to [`sourcevocab`](../../../core/database/sourcevocab/) — e.g. expand `subjecttypes` into / introduce `subjectvocab` that owns Install for types + properties + bindings). That registry is the **only** product definition of:

| Concern | Declared on | Used by |
| --- | --- | --- |
| Type identity | `key`, labels, ref prefixes, `origin=provenencia` | DB seed rows (as today) |
| **Role / capabilities** | e.g. root vs bridge vs reification; placeable on Evidence graph; create requires Citation | Palette, card chrome, connect entry, empty states |
| **Graph presentation** | Everything type-keyed for Evidence graph chrome: display name / `L10n` key, icon symbol token, card colors (ink / tint / chip / line tokens), edge/gradient tokens for roots **and** bridges | Palette, cards, ghosts, bridge cards, relationship **lines** |
| **Properties** | key, value_type, labels | `properties` seed + Subject fields + Observation editors |
| **Bindings** | type_key → property_key (+ sort) | `subject_type_fields` seed; Add-property menus |
| **Required / locked bindings** | which seeded bindings macros need and UI must not unbind | Subject fields delete/unbind; connect pre-fill |
| **Connect matrix** | allowed endpoint pairs + which bridge type + which edge Properties / disambiguation fields | S7-10 macros (read registry; do not re-encode §3.2 in the view) |

**Type-keyed chrome:** Today’s `EvidencePrimaryKind` / `EvidenceSubjectKindStyle` (and edge gradients that sample primary ink) are the anti-pattern. The graph loads presentation from the registry for **every** Subject type it can show, then maps tokens to design-system colors/symbols locally. If it is configured *because of the type* (card name, palette label, wash, chip, line ink, gradient endpoints), it belongs in the registry. Platform tokens (`PVColor.…`, SF Symbols, `L10n` keys) stay in the design system / catalogs; the registry stores **which token for which type**.

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
| `IconSymbol` | SF Symbol / `PVSymbol` token |
| `InkToken`, `TintToken`, `ChipToken`, `LineToken` | Map to `PVColor` (or equivalent) in Swift — card wash, chip, stroke |
| `EdgeFromToken` / `EdgeToToken` | Optional; default edge gradients can derive from endpoint `InkToken` / `LineToken` if unset |

**`seedProperty`**

| Field | Purpose |
| --- | --- |
| `Key`, `Label`, `Description` | Catalog identity |
| `ValueType` | `text` \| `integer` \| `date` \| `name` \| `subject` |

**`seedBinding`**

| Field | Purpose |
| --- | --- |
| `TypeKey`, `PropertyKey`, `SortOrder` | Seed `subject_type_fields` |
| `Locked` | Required for macros / product integrity — Subject fields UI must not unbind or delete while locked |

**`seedConnect`** (one row per allowed endpoint pair)

| Field | Purpose |
| --- | --- |
| `FromTypeKey`, `ToTypeKey` | Ordered pair (or undirected flag if needed) |
| `BridgeTypeKey` | e.g. `participation`, `relationship`, `location` |
| `EdgePropertyKeys[]` | Observations to pre-fill (usually two) |
| `Disambiguation` | e.g. `none` \| `role` \| `relationship_type` \| `person_person_choice` (shared-event vs relationship) |
| `Refuse` | If true, pair is explicitly illegal (optional; omit row = refuse by default) |

**Lookup API (package + FFI as needed):** `PlaceableTypes()`, `Presentation(key)`, `BindingsForType(key)`, `LockedBinding(type, property)`, `Connect(from, to)`. Graph / Subject fields / connect call these — never re-list the three primaries in Swift.

**Out of registry for Spike 7:** researcher-authored types; raw hex colors; full plugin manifest file format (same structs, different Install source, later).

## Composer presentation (locked: Option B)

**Navigate away** to a first-class workspace place. Canvas unloads; full-window side-by-side viewer|form; **Back** returns to the Evidence graph.

Rationale: keep graph a11y/keyboard from growing; composer gets a clean focus model and room for PDF/locator tools.

In-window modal over the graph and companion `NSWindow` are **out**.

### Navigation / breadcrumbs (defaults — confirm in S7-D4)

Composer is a **deep place under Sources**. Extend `WorkspaceLocation` (Spike 5 `sourceSurface` page|graph) with a composer discriminant plus context (`subjectId`, optional connect draft ids). Follow `add-workspace-location`.

**Breadcrumb shape (default):**

```text
Sources › {Source title} › Evidence graph › Cite {subject label}
```

Connect entry:

```text
Sources › {Source title} › Evidence graph › Connect › Cite
```

**History policy (defaults):**

- Entering the composer **pushes** history (Back → graph, same `sourceId`).
- **Successful submit** pops or replaces back to the graph (prefer single Back, not stacked drafts).
- **Cancel / Back** without submit writes nothing.
- Persist across relaunch only if subject still exists; else fallback to Evidence graph for that Source.
- Camera/selection stay out of the composer location.
- Connect **disambiguation** is a small sheet on the graph (not its own history entry).

Pinning a Citation across successive graph edits is **out** (one Citation + N Observations per submit).

## Design track (four briefs — one view / surface each)

**All UI is designed in Claude Design before the matching UI PR.** Briefs: [`design/`](design/).

| Step | Brief | Covers | Gates |
| --- | --- | --- | --- |
| **S7-D2** | Subject fields | Properties + bindings to **seeded** Subject types; five value_types; mirror Source fields | S7-05 |
| **S7-D3** | Evidence graph updates | Add-property; cited-data rows; Artifact gate; connect disambiguation → composer handoff; bridge honesty once cited | S7-09, S7-10 |
| **S7-D4** | Citation composer place | Full-window viewer\|form; Artifact pick; locators; observation list; DateValue reuse; breadcrumbs; composer-only a11y — **hosts** NameValue modal, does not design it | S7-08 |
| **S7-D5** | NameValue editor | Reusable NameValue modal (DateValue twin); form + optional parts | S7-02b |

~~**S7-D1** Subject types editor~~ — **descoped** (see [Descoped](#descoped) below).

Run **S7-D2 / D5** early (parallel with schema). **S7-D3 before card/connect UI.** **S7-D4 before composer place UI.** **S7-D5 before NameValue Swift UI.** D3, D4, and D5 may run in parallel once boundaries are clear.

## PR sequence

```text
design                              build
─────────────                       ─────────────────────────────────────────

S7-D2 Subject fields                S7-01  properties + subject_type_fields
  │                                 │      + **central subject registry**
  │                                 │      (capabilities, locked bindings,
  │                                 │       connect matrix, Install)
  │                                 ▼
S7-D5 NameValue editor              S7-02  NameValue schema + Go
  │                                 │      (ungated; tables + package)
  │                                 ▼
  │                                 S7-03  citations + observations + locator
  │                                 │      validation (Go) + FFI macros
  │                                 ▼
  └────── gates ──────────────────▶ S7-02b NameValue Swift editor
                                    │      (reusable modal; DateValue twin)
                                    │
  └────── D2 gates ───────────────▶ S7-05  Subject fields UI
                                    │
S7-D3 Graph updates                 │
S7-D4 Composer place                │
  │                                 S7-06  Artifact viewers: image + PDF
  │                                 │
  │                                 ▼
  │                                 S7-07  Locator tools: page + region
  │                                 │
  │                                 ▼
  └────── D4 gates ───────────────▶ S7-08  Citation composer place (B)
  │                                 │      (+ host NameValue from S7-02b)
  │                                 │      WorkspaceLocation + breadcrumbs
  │                                 │      Citation + N Observations submit
  │                                 ▼
  └────── D3 gates ───────────────▶ S7-09  Card growth + Add property
                                    │      (navigates to composer)
                                    ▼
                                  S7-10  Durable connect macros
                                    │      (disambiguation → composer)
                                    ▼
                                  S7-11  Dogfood close / docs
```

Schema/Go PRs (01–03, 02) may start before design finishes; **UI PRs gate on the matching brief.** S7-03 needs S7-02 (schema) only — not S7-02b. S7-08 needs **S7-02b** + **S7-D4**.

---

## Checklist

- [ ] S7-D2 — Design: Subject fields → [`completed.md`](completed.md)
- [ ] S7-D3 — Design: Evidence graph updates → [`completed.md`](completed.md)
- [ ] S7-D4 — Design: Citation composer place → [`completed.md`](completed.md)
- [ ] S7-D5 — Design: NameValue editor → [`completed.md`](completed.md)
- [ ] S7-01 — `properties` + `subject_type_fields` + **central Interpretation subject registry** + Go/FFI → [`completed.md`](completed.md)
- [ ] S7-02 — NameValue schema + Go → [`completed.md`](completed.md)
- [ ] S7-02b — NameValue Swift editor → [`completed.md`](completed.md)
- [ ] S7-03 — Citations + Observations + locator validation + FFI → [`completed.md`](completed.md)
- [ ] S7-05 — Subject fields UI → [`completed.md`](completed.md)
- [ ] S7-06 — Artifact viewers (image + PDF) → [`completed.md`](completed.md)
- [ ] S7-07 — Locator tools (page + region) → [`completed.md`](completed.md)
- [ ] S7-08 — Citation composer place → [`completed.md`](completed.md)
- [ ] S7-09 — Card growth + Add property → [`completed.md`](completed.md)
- [ ] S7-10 — Durable connect macros → [`completed.md`](completed.md)
- [ ] S7-11 — Dogfood close / docs → [`completed.md`](completed.md)

## Descoped

| Step | Notes |
| --- | --- |
| **S7-D1** / **S7-04** | Subject types CatalogVocabulary editor. Types remain Spike 5 seeded rows + first-class graph behavior. Brief archived: [`design/archive/S7-D1-subject-types.md`](design/archive/S7-D1-subject-types.md). |

---

## S7-D2 — Design: Subject fields

Claude Design board for Subject fields (Properties + bindings). Brief: [`design/S7-D2-subject-fields.md`](design/S7-D2-subject-fields.md). Gates **S7-05**. Bindings target the **fixed seeded** Subject types — not a user type browser.

---

## S7-D3 — Design: Evidence graph updates

Claude Design board for Add property, cited rows, connect disambiguation handoff. Brief: [`design/S7-D3-evidence-graph-updates.md`](design/S7-D3-evidence-graph-updates.md). Gates **S7-09**, **S7-10**. Does **not** design the composer place (S7-D4).

---

## S7-D4 — Design: Citation composer place

Claude Design board for the navigable composer. Brief: [`design/S7-D4-citation-composer.md`](design/S7-D4-citation-composer.md). Gates **S7-08**. Confirms breadcrumbs and history policy. **Does not** design the NameValue editor (S7-D5) — only the host affordance that opens it.

---

## S7-D5 — Design: NameValue editor

Claude Design board for the reusable NameValue modal (DateValue twin). Brief: [`design/S7-D5-name-value-editor.md`](design/S7-D5-name-value-editor.md). Gates **S7-02b**.

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
| **Out** | Subject fields UI (S7-05); Observations; Subject types user CRUD; scattering capability checks, palette membership, or kind→color/gradient maps in Swift views. |
| **Testable** | Create project seeds §3.2–3.3 bindings; registry tests for capabilities, full presentation tokens, locked bindings, placeable set; Go tests; FakeStore round-trip. |
| **Depends on** | Spike 5 `subject_types` table (may fold Install into the new registry package). **Not** gated on design. |

---

## S7-02 — NameValue schema + Go

`name_values` / `name_value_parts` per structured-name-model §2–3 (Interpretation only — not Conclusion `name_format`); `core/database/namevalues`. **No Swift UI** in this PR.

| | |
| --- | --- |
| **In** | Schema, Go package, tests. |
| **Out** | Swift editor (S7-02b); composer wiring (S7-08); name_format profiles. |
| **Testable** | Go round-trip create/read parts. |
| **Depends on** | — (can parallel S7-01). **Not** gated on S7-D5. |

---

## S7-02b — NameValue Swift editor

Swift `NameValueDraft` / editor modal under `Features/Names/` (mirror `Features/Dates/`). Reusable from the composer and later hosts.

| | |
| --- | --- |
| **In** | Reusable modal UI per S7-D5; unit tests for draft validation. |
| **Out** | Composer host wiring (S7-08 opens the modal); schema (S7-02). |
| **Testable** | Swift tests for form-required / parts ordering; preview of modal. |
| **Depends on** | S7-02, **S7-D5**. |

---

## S7-03 — Citations + Observations + locator validation

`citations`, `citation_notes`, `observations`, `observation_notes`; value columns for the five types only; Go locator validate (`page`, `region`; optional `text_quote`); atomic “create Citation + Observations” RPC; graph payload includes Observations for card rows.

| | |
| --- | --- |
| **In** | Migrations, Go packages, locator invariants, FFI macros, graph query enrichment. |
| **Out** | Composer UI; card chrome. |
| **Testable** | Go locator reject/accept; atomic create; FakeStore. |
| **Depends on** | S7-01, S7-02 (name column / FK). |

---

## S7-05 — Subject fields UI

Properties + `subject_type_fields` bindings; five value_types only; mirror Source fields. Bindings pick among **seeded** Subject types only (no Subject types admin destination).

| | |
| --- | --- |
| **In** | Vocabulary browser; bind Properties to seeded Subject types; value_type at create (immutable); respect registry **locked** bindings. |
| **Out** | Composer; Observation editors beyond type pickers; Subject types CRUD. |
| **Testable** | Bind a field; see it available for that type. |
| **Depends on** | S7-01, **S7-D2**. |

---

## S7-06 — Artifact viewers (image + PDF)

PDFKit + image overlay; resolve files via `ProjectFiles`; no QuickLook for composer (no locator API).

| | |
| --- | --- |
| **In** | Zoom/pan image; PDF page nav + zoom/pan; hostable in composer. |
| **Out** | Locator drawing (S7-07); audio/video. |
| **Testable** | Open a PDF Artifact and an image Artifact in a host preview/harness. |
| **Depends on** | Existing Artifacts / Files. **Not** gated on S7-D4 (chrome waits on D4). |

---

## S7-07 — Locator tools (page + region)

Interactive tools feeding nested selectors; Go is source of truth for invariants (≥3 points, non-self-intersecting, etc.).

| | |
| --- | --- |
| **In** | Page selector UI; polygon region draw on image/PDF page; produce `locator_json`. |
| **Out** | `time_range`; required `text_quote` (optional if cheap). |
| **Testable** | Drawn region validates in Go; invalid polygons rejected. |
| **Depends on** | S7-06, S7-03 validation. |

---

## S7-08 — Citation composer place (Option B)

New workspace place + location discriminant; breadcrumb per S7-D4; Artifact pick; form (citation fields + N Observations); typed value dispatch; single submit; cancel/Back policy; composer-only a11y.

| | |
| --- | --- |
| **In** | Navigable place; viewer\|form layout; submit Citation + Observations; host NameValue modal from S7-02b; reuse DateValue. |
| **Out** | Card growth wiring (S7-09); connect macros (S7-10); pinning; designing NameValue itself (S7-D5 / S7-02b). |
| **Testable** | From a temporary entry point or FakeStore-driven nav: cite a page/region, add two Observations (including a name), submit, see rows in catalog; Back returns to graph. |
| **Depends on** | S7-03, S7-06, S7-07, S7-02b, **S7-D4**. |

---

## S7-09 — Card growth + Add property

Grow `EvidenceSubjectCardChrome`; Add property → `go(to: composer)`; edge layout heights; graph a11y only for new card controls.

| | |
| --- | --- |
| **In** | Cited-data rows; Add property control; navigation to composer; uncited → cited shell when Observations exist; **palette, card, and edge chrome driven by registry presentation** (retire hard-coded Add Person/Event/Place set and kind→style switches). |
| **Out** | Connect durability (S7-10). |
| **Testable** | Add property on a Person → composer → submit → card shows row and grows. |
| **Depends on** | S7-08, **S7-D3**. |

---

## S7-10 — Durable connect macros

Disambiguation sheet on graph (§3.2 matrix); navigate to composer with pre-filled edge Observations; replace provisional Spike 6 links.

| | |
| --- | --- |
| **In** | person→event role; person→person relationship vs shared-event choice; event→place clean; refuse unsupported; atomic bridge + Citation + edges — **driven by the S7-01 registry connect matrix**, not a second hard-coded table in the view. |
| **Out** | Pinning; full conflicted/negated chrome; inventing connect rules outside the registry. |
| **Testable** | Connect two people → disambiguate → cite → bridge persists with Observations; relaunch keeps edges; unsupported pairs refuse per registry. |
| **Depends on** | S7-09, **S7-D3**, S7-01 registry. |

---

## S7-11 — Dogfood close / docs

Honesty pass against the [goal bar](#goal-dogfood-bar). Record in [`completed.md`](completed.md); update [`README.md`](README.md); archive design briefs; point [`docs/deployment-plan/README.md`](../README.md) at archive when closed; SemVer only if cutting a product release.

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
| Five value-type editors | `real` / `boolean`; full conflicted/negated visual language |
| Seeded Subject types (Spike 5) | **Subject types** CatalogVocabulary / user-defined types |

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
10. **Plugin path is the registry** — when plugins arrive, they extend Install/registry modules with the same capability fields; do not invent a second configuration channel.

---

## Definition of done

- [ ] Checklist above complete
- [ ] Dogfood bar items 1–8 met
- [ ] Design briefs archived under `design/archive/`
- [ ] [`docs/deployment-plan/README.md`](../README.md) points at archive when closed
- [ ] Design note status line points at Spike 7 archive (slices 3–7 advanced)

## What the next spike inherits

On success: a cited Evidence graph — Subject fields configurable against seeded types, properties visible on cards, connect writes real evidence, NameValue available for person names. Ready for honesty/polish (conflicted/negated), Source-page commentary, and media types beyond image/PDF.
