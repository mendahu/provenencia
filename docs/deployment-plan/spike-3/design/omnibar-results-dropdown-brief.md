# S3 — Omnibar search results dropdown (rich hit rows)

Kind: Claude Design board
Spike: Provenencia Spike 3 (workspace layout + catalog search)
Depends on: App Layout toolbar (omnibar field already top-right of main column, Cmd-K); navigation history Back/Forward + breadcrumbs in the same toolbar
Authoritative behavior: docs/deployment-plan/spike-3/omnibar-search.md
Out of this board: search engine / FTS ranking internals; FFI; redesigning the toolbar field placement (already designed)

Paste this entire document into Claude Design as the requirements for one board/flow.

---

## 1. Objective

Design the omnibar results dropdown that opens from the existing toolbar search field while the researcher types (or after Cmd-K focus + query).

This is a project-scoped command-palette / Spotlight results surface — not web search, not Settings search, not a second window. Choosing a hit navigates via workspace history (go(to:)) and dismisses the dropdown.

Locked presentation direction to explore and refine: one shared rich row skeleton with kind-specific slots; results are a flat ranked list (engine order). Do not invent a unique card layout per entity kind. Grouped sections / facets are optional explorations only if the flat list stays confusing.

Toolbar:

  [ <- ] [ -> ]  Breadcrumbs ...     [ search  query........  Cmd-K ]
                                       |
                                       +-- dropdown ------------------+
                                       |  rich hit rows               |
                                       |  (keyboard + click)          |
                                       |  dismiss on select / Esc     |
                                       +------------------------------+

Do not redraw the full app shell, sidebar, or omnibar field chrome except as needed to show the dropdown anchored correctly under/near the field.

---

## 2. Product context (must respect)

1. Offline-first macOS genealogy app; search is scoped to the open project only.
2. Visual language: modern-but-archival — parchment surfaces, serif display for page titles elsewhere, Spectral body, IBM Plex Mono for refs. Match shipped workspace / onboarding tokens. No purple dashboard aesthetic; no emoji as primary icons.
3. Composition: dropdown is a transient raised surface (menu/popover energy), not a page of cards.
4. Heterogeneous hits must still feel like one list. Kind disambiguation is a chip/label + leading visual, not ten different row templates.
5. Notes / metadata / filenames under a Source are not separate hits — matching them returns the Source, with optional "why matched" context. Vocabulary types/fields are their own hits (sidebar destinations).

---

## 3. Shared hit-row skeleton (design this hard)

Every result row uses the same chrome. Slots fill differently by kind.

  +--------------------------------------------------------------------+
  | [ leading visual ]   PRIMARY TITLE                      KIND CHIP  |
  |                      secondary line                       REF mono |
  |                      tertiary / match context (optional)           |
  +--------------------------------------------------------------------+

Slot map:

- Leading visual: thumb, type icon, or file-type glyph. Fixed size (align with PVThumbnail energy). Never collapse the slot when missing — use placeholder / glyph.
- Primary title: best human name. One line, ellipsis; highest emphasis in the row.
- Kind chip: short disambiguator (Source, Field, Type, File, Artifact, Person, ...). Not a long sentence.
- Secondary: supporting identity (type name, data_type, media type, origin, parent Source title, etc.). Muted.
- Ref: trailing short catalog id when present (SRC-..., ART-..., etc.). IBM Plex Mono. Omit slot content when the entity has no ref.
- Tertiary / match context: optional one muted line explaining why this hit matched when the primary did not contain the query (e.g. "Note: ...", "Metadata · Author: ...", "Filename: ..."). Omit when the match is obviously the title/ref.

Interaction:

- Hover / selected keyboard focus: one clear selected state (arrow keys move; Return activates).
- Click / Return: navigate + dismiss.
- Esc / click outside: dismiss without navigating.
- Show a calm loading state (typing debounce) and empty state ("No matches for ...").
- Exact ref paste (SRC-F4N2P): that hit should be visually obvious at the top (optional subtle "exact ref" emphasis without breaking the skeleton).

Density: richer than a single-line menu item; still denser than Source page cards. Aim for scannable 8–12 rows before scroll.

---

## 4. Per-kind slot content (rich rows)

Design example rows for each kind below (Spike 3 ships Sources + types + fields first; still design later kinds so the skeleton flexes). Use realistic family-research copy.

### 4.1 Source (SRC-...) — navigable root

- Leading: raster thumbnail of primary evidence when available; else file-type glyph; else Source-type icon / placeholder
- Primary: Source title
- Kind: Source
- Secondary: Source type label (e.g. Parish register)
- Ref: SRC-3K9M2
- Tertiary: when match was in rolled-up child text: Note / Metadata · field label / Filename

### 4.2 Source type (vocabulary)

- Leading: curated Source-type icon (book, microfilm, certificate, ...) if that system exists; else generic type glyph
- Primary: type label
- Kind: Type or Source type
- Secondary: machine key (mono or muted) and optional origin cue (provenencia / custom)
- Ref: none
- Tertiary: match in description only, if needed

### 4.3 Source field (vocabulary)

- Leading: field / list glyph (not a photo thumb)
- Primary: field label
- Kind: Field
- Secondary: key · data_type (text / date / ...) · origin cue
- Ref: none
- Tertiary: match in description only, if needed

### 4.4 Artifact (ART-...) — optional first-class hit (design even if v1 rolls into Source)

- Leading: Artifact/File thumb or MIME glyph or fileless type icon
- Primary: Artifact label (or description fallback)
- Kind: Artifact
- Secondary: parent Source title (critical — Artifacts live under Sources)
- Ref: ART-...
- Tertiary: filename if that was the match

### 4.5 File (when searchable)

- Leading: MIME / media glyph or thumb if image
- Primary: original_filename
- Kind: File
- Secondary: media type · parent Source title (and Artifact label if useful)
- Ref: none
- Tertiary: omit unless needed

### 4.6 Later layers (skeleton must not break)

Show at least one example each so the system scales:

- Citation: primary = short quote / page label / description; secondary = parent Source or locator context; ref = CIT-...
- Observation: primary = property label + value summary; secondary = Node context; ref = OBS-...
- Node (candidate): primary = label / name form; secondary = node type; ref = PER-C-... etc.
- Canonical entity: primary = working label / projected name; secondary = kind (Person, Place, ...); ref = PER-... etc.

Do not design Relationship as a top-level hit row unless you have a clear identity; prefer noting out of scope / weak identity.

---

## 5. Dropdown chrome (around the list)

Design the container, not only rows:

1. Anchoring — under the toolbar omnibar field; width related to the field (may widen slightly for rich rows); max height with internal scroll.
2. Elevation — raised surface, subtle border, large shadow (same family as Back/Forward history menu on App Layout).
3. Header (optional) — query echo or "Results"; keep minimal; no heavy toolbar chrome inside the dropdown.
4. Footer (optional) — hint for arrow keys / Return / Esc only if it stays calm.
5. Empty — No matches for "..." with short guidance (check spelling / try a ref).
6. Loading — inline, not a blocking app modal.
7. Mixed-kind list — flat ranked order from the engine (Sources may sit above Types for the same query). Kind chip is the disambiguator. Optional alternate frame: grouped by kind (Sources / Types / Fields) with sticky mini-headers — present as a variant only if you believe flat fails.

---

## 6. States and scenarios (frame inventory)

Produce frames for at least:

1. Omnibar focused, empty query (blank vs "recent" placeholder — pick one recommendation; recent is optional later).
2. Typing query — loading.
3. Mixed results: Sources + Source types + Source fields for "birth" or "John Smith birth certificate".
4. Exact ref hit on top: query SRC-3K9M2.
5. Source hit where match is only in a note/metadata (tertiary match context visible).
6. Source hit with thumbnail vs PDF glyph vs fileless type icon (three leading-visual variants).
7. Keyboard selection mid-list.
8. No matches.
9. Long primary title + long secondary (ellipsis / wrapping rules).
10. Optional: grouped-by-kind variant of frame 3 for comparison.

Annotate which slots are required vs optional on each example row.

---

## 7. Design-system guidance

- Leading tile: reuse / extend PVThumbnail patterns (image, placeholder, future glyph modes).
- Ref: mono token treatment already used for SRC-... / USR-...
- Kind chip: prefer existing chip language if any; else a quiet caption/pill that is not a loud badge sticker.
- Dropdown surface: same family as workspace history jump menu (raised, border, shadow).
- Do not: per-kind card layouts; dashboard tiles; purple accents; emoji icons.
- Call out anything that should become a reusable OmnibarHitRow / PVSearchResultRow component.

---

## 8. Out of scope

- Redesigning sidebar or Back/Forward
- Search algorithm, FTS, scoring UI ("sort by relevance" controls)
- Faceted query syntax UI (type:person) unless a light optional exploration
- Settings / audit / help search
- Cross-project search
- AI / natural-language "ask the catalog"

---

## 9. Acceptance checks

- One shared row skeleton; kinds differ by slot content only
- Every Spike 3 kind (Source, type, field) has a concrete example row with all slots filled or explicitly empty
- Source rows show how thumb / MIME glyph / type icon / match-context appear
- Refs are mono and trailing when present
- Dropdown anchoring, scroll, empty, loading, keyboard selection are shown
- Flat ranked list is the default recommendation; grouping only as a clear alternate
- Visual language matches Provenencia workspace (archival, calm, not a third product)
