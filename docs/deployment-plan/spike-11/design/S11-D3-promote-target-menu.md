# S11-D3 — Promote page target menu: find another record

**Kind:** Claude Design board (**enhancement** of the Promote page board, S9-D16, as extended by S11-D1)
**Spike:** Provenencia Spike 11 (Promote matching layers)
**View:** Promote page — a row's target menu and the picker it opens
**Implements later as:** PR **S11-36** (find another record in the target menu)
**Depends on:** S11-35 (landed [#543](https://github.com/mendahu/provenencia/pull/543): `PromoteFlow.hold` holds a record Promote did not offer); S11-37 (landed [#551](https://github.com/mendahu/provenencia/pull/551): `CatalogSearchModel` and `CatalogSearchHitRow`, the omnibar's search and hit row in a shared home)
**Related:** [`../promote-find-another-record.md`](../promote-find-another-record.md) (the feature and the [decided](../promote-find-another-record.md#decided-user-2026-10-24) product calls), [`archive/S11-D1-promote-rows.md`](archive/S11-D1-promote-rows.md) (the rows this board extends), [`../../../promote-graph-alignment.md`](../../../promote-graph-alignment.md) §3
**Board:** [Claude Design — Promote flow](https://claude.ai/design/p/bc84685e-bbc3-4053-a5c9-0f5ac7a13ccd?via=share), agreed 2026-10-25: frames 20 – 26b (§8)
**Design system layers:** [`docs/design-system-layers.md`](../../../design-system-layers.md)
**Skill:** [`add-design-brief`](../../../../.agents/skills/add-design-brief/SKILL.md); [`add-ui-component`](../../../../.agents/skills/add-ui-component/SKILL.md)

Paste this entire document into Claude Design as the requirements for one board. This file is the whole prompt.

### Claude Design — do this first (in order)

Work **in place** on this board. Do not fork a parallel copy of the surface.
- **Rethink** (this brief says replace): throw away the old frames. Do not keep a before/after to ship.
- **Enhancement**: add to the existing frames. Do not start a second composer / graph / page.

1. **Clear this board’s local design-system cache.** Claude Design keeps a stale pack; drawing against it invents local copies of kit controls.
2. **Delete this board’s reference** to the design-system bundle.
3. **Pull a fresh copy** of the Provenencia design system from the main project. Do not continue until the fetched kit lists current components. If the kit looks stale or empty, delete the cache and refetch. Do **not** draw a replacement kit locally.
4. **Compose from that kit.** Instance existing components. Reach for a **bespoke / local** control only when the use is truly this domain. One call site is not a new design-system primitive.

**Reach for (kit).** Instance these first. The **UI building-block inventory** later in this brief names the snowflakes and which kit piece each situation should use.

| Situation | Use |
| --- | --- |
| Labeled value, textarea, or trailing control | Field + TextArea / Input |
| Primary / secondary / ghost action | Button; icon-only → IconButton |
| Choose one from a short list | Select |
| Searchable pick | ComboBox |
| Warning, error, or inline hint | Callout |
| Page- or pane-level empty | EmptyState |
| Confirm replace or destroy | Confirm (`item:` snapshot, not a Bool) |
| Resource delete with inbound check | DeleteImpact recipe (S8-D9 / **S8-13**) — confirm if allowed, notice if blocked |
| Short create / edit form | FormDialog |
| Status / count / polarity mark | Badge; compact token → Chip |
| Cover or file thumb | Thumbnail |
| Grouping / raised or sunken row | Card |
| Section title | SectionHeader |
| Transient after-save notice | Toast |
| Native menu of actions | ContextMenu |

Do **not** invent a local Field, Button, Card, Select, Callout, or Confirm.

### This brief is an enhancement

Extend the Promote page frames that exist (S9-D16, S11-D1 frames 14 – 19, S11-D2's sheet title). Do not redraw the header, the section headers, the row's name column, its assessment column layout, its Updated / Decided column, its notes, its connection subrows, the footer, or the evidence sheet. Three things change: the target menu gains one last entry, choosing that entry turns the row's target cell into a search field for a moment, and a record the researcher found that way reads its weak evidence in words that do not contradict the choice.

---

## 1. Objective

A researcher on the Promote page who knows a Subject is a record Promote did not offer (a nickname the aliases don't hold, a recorded variant, a married name) can find that record from the row and file the Subject on it.

**What changes.**

1. Every working row's target menu ends with **Find another record…**.
2. Choosing it turns the row's target cell into a search field. Typing searches every record of the row's kind in the catalog, ranked by how well it matches what was typed, in the list the toolbar search draws.
3. Picking a hit makes it the row's target, like a menu pick. The record joins the row's menu, the row is Decided, and the assessment column shows the record's own band, Contested when it conflicts, and "Your choice".
4. A chosen record the evidence rates below Weak no longer reads "No match" beside "Your choice".

**What stays.** The menu's other items and their order (records, then New, then Skip), how a record is drawn in the menu, the band and Contested Badges, the Decided and Updated Badges, the conflict and duplicate callouts, the evidence sheet. Promote's own proposals do not change.

---

## 2. Domain facts

| Fact | What it means on this board |
| --- | --- |
| A row's target menu | Today: the records Promote offers for the Subject (the proposed record first, then alternatives), then "New Person" / "New Event" / "New Place", then "Skip — not filed yet". A row Promote finds nothing for has only New and Skip. Each record item shows the kind mark, the record's title, and "ref · life line" under it. |
| Which rows have a menu | Every working row. An already-filed row (anchor) has no menu and gets no entry. |
| A row's kind | Person, event or place. The search finds records of that kind only. A person row never finds a place. |
| What the search finds | Every record of the kind in the catalog, merged records left out, ranked by the catalog search's own match to the typed text. Records already in the menu, and records another row on the page has taken, are listed too. Nothing is filtered or warned about. |
| The search needs two letters | Under two typed characters there is no search. The search waits a fifth of a second after the last keystroke, then answers well under a second on a real catalog. |
| What a hit shows | The toolbar search's hit row: a lead mark tile, the record's title (a person sorted: "Jones, Polly"), a kind Badge, the life line ("1871 – 1950 · Bristol") or the event's or place's secondary line, the record's ref (accented when the match was on the ref), and a match line when the match was in a field other than the title ("Name: Polly Ann Jones"). |
| Picking is a decision | A pick is the researcher's choice, like choosing a record or New from the menu: the row shows Decided, and the page asks Promote again with that choice held. Promote never moves it. |
| The band comes after | The page re-asks Promote on every decision; the answer takes well under a second. Until then the assessment column is empty, as it is today right after any menu change. Then it shows the record's band, Contested when a property or link conflicts, and the reason "Your choice". |
| A found record's band can be None | A record whose first name or surname differs from the Subject's (the reason it was not offered) often rates below Weak: band None. Today band None reads "No match", so the row reads "No match · Your choice". The researcher has said they are one; "No match" reads as an error. |
| No warning | A differently named pick shows no Callout and no extra mark. The band and Contested are enough (user, 2026-10-24). |
| The found record stays | It stays in the row's menu while the page is open, so the researcher can go back to an offered record, New or Skip, and back to it. |
| Opening the search is not a decision | Choosing "Find another record…" and then cancelling leaves the row exactly as it was: same target, not Decided. |
| One search at a time | Opening the search on another row closes the first one. |
| Filing in progress | While Done files, every target menu is disabled; so are the entry and the search. |

### 2.1 What this board is not

- Searching from the evidence sheet, or comparing two records side by side before picking.
- Records of another kind.
- Merging two records.
- A warning or Callout for a differently named pick.
- A band, score or Contested mark per hit in the search list. The search does not compare; the row does after the pick.
- The toolbar search. Its panel and rows are drawn already and do not change.

---

## 3. Implementation gate

| Ships in **S11-36** | Stays out |
| --- | --- |
| FR-1 – FR-9: the menu entry, the search field in the target cell and its list, picking and cancelling, the found record in the menu, the below-Weak wording for a chosen record, VoiceOver. | Searching from the evidence sheet, other kinds, merging ([what stays later](../promote-find-another-record.md#what-stays-later)). A Callout for a `Different` pick (refused, user 2026-10-24). Any change to the toolbar search's panel or ranking. |

---

## 4. Requirements

| ID | Requirement |
| --- | --- |
| FR-1 | Every working row's target menu ends with one more item after Skip: a **search** icon and "Find another record…" in secondary text. It is never the field's selection and never shows the checkmark. No divider and no section title: the menu has none today. Anchor rows have no menu and no entry. |
| FR-2 | Choosing the entry replaces the row's target menu, in the same column and the same width, with a small **ComboBox** search field holding the focus, its placeholder "Search people", "Search events" or "Search places" by the row's kind, and a small **IconButton** (dismiss **×**) after it labelled "Cancel search". The field and the button together fill the target column. The row's other columns do not move: the assessment column and the Badges show what they showed before the entry was chosen. |
| FR-3 | Typing opens the ComboBox's own list under the field, at the field's width: one hit row per record, the toolbar search's hit row as is (inventory). With text typed, the top hit is highlighted so Return takes it. The list scrolls past about five hits. |
| FR-4 | The list's line when it has no hits: under two characters typed (and on an empty field), "Type at least two letters of a name or a ref"; while the first results load, the ComboBox's own searching line ("Searching…"); hits already shown stay while a new search runs; no hits, "No people match “Polly”" (by kind: "No events match …", "No places match …"); a failed search, its error message in the same line. |
| FR-5 | Picking a hit (click, or Return on the highlighted hit) closes the search, and the target menu comes back with the record as its selection. The row shows the Decided Badge. When Promote answers, the assessment column shows the record's band Badge, the Contested Badge when the pair conflicts, and "Your choice". Opening the menu shows the record as an item after the offered records and before New, drawn like every record item (mark, title, "ref · life line"), checked. A hit that is already in the menu is selected there, not added twice. |
| FR-6 | **Cancel search**, or Escape with nothing typed, closes the search: the target menu comes back unchanged and the row is not Decided unless it was before. Escape is the ComboBox's: the first press closes the list, the next clears what was typed, the one after closes the search. A click outside closes the list and leaves the field, as every ComboBox does; the search stays in the target cell, with its Cancel button, until a pick or a cancel. Choosing "Find another record…" on another row closes this one. |
| FR-7 | A record the researcher chose, by search or from the menu, whose band is None reads **"Little evidence"** in the neutral band Badge instead of "No match", on the row and on the evidence sheet's title. Contested shows after it when the pair conflicts. A suggested record Promote rates None, not yet chosen, keeps "No match". No Callout. |
| FR-8 | VoiceOver: the entry reads "Find another record"; the field and its list read "Find another record for Ada Smith"; each hit reads as the toolbar search's does: title, kind, life line, ref, match line ("Jones, Polly, Person, 1871 – 1950, PER-7QH3D"); the button reads "Cancel search". A FR-7 row's evidence button reads "Evidence for Ada Smith: Little evidence, Your choice". |
| FR-9 | While Done files, the entry is disabled with the rest of the menu, and an open search field and its Cancel button are disabled. |

---

## 5. Suggested frames

Number them after S11-D2's frames on the board.

1. **The entry**: a person row Promote found nothing for (target "Select where this files", menu open: New Person, Skip — not filed yet, Find another record…) and a person row with two offered records, menu open, the entry last (FR-1).
2. **The search, empty**: the first row with the search field focused in the target cell, placeholder "Search people", Cancel button, the list open on the "Type at least two letters…" line (FR-2, FR-4). The assessment column is empty and the row is not Decided.
3. **Hits**: "Polly" typed, four person hits, the first highlighted; one hit has a match line, one hit's ref is accented (FR-3). Variant: the searching line before the first results (FR-4).
4. **No hits and a failure**: "Pollx" with "No people match “Pollx”", and the error line (FR-4).
5. **Picked**: the row with "Jones, Polly · PER-7QH3D" selected, Decided, a Weak band Badge, Contested and "Your choice"; the menu open on it with the record checked between the offered records and New (FR-5).
6. **Little evidence**: a row holding a found record at band None ("Little evidence · Your choice"), one at None and Contested, and the evidence sheet's title for the first (FR-7).
7. **Two rows**: a search open on one row, with an event row below showing its menu untouched (FR-6), and a row while Done files with everything disabled (FR-9).

---

## 6. UI building-block inventory

This table is **binding**. Instance the Ship kit rows. Paths from `macos/App/` unless noted.

| Building block | Layer | Status | Home | Notes |
| --- | --- | --- | --- | --- |
| Promote row | Snowflake | **Extend** | `macos/App/Features/Promote/PromoteView.swift` `PromoteAlignmentRow` | Its target cell shows either the target menu or the search (FR-2). No other column changes. |
| Target menu | Component | Ship | `macos/App/DesignSystem/Components/Select/PVSelect.swift` | The entry is one more option. No action item, divider or section in Select: the kit gains nothing. |
| Menu entry | Snowflake | **Extend** | `macos/App/Features/Promote/PromoteView.swift` `PromoteTargetMenuRow` | Draws the search icon and the text for the entry, as it already draws New and Skip. |
| Found record in the menu | Snowflake | Ship | `macos/App/Features/Promote/PromoteView.swift` `PromoteTargetMenuRow` | Drawn like every record item. No "found" mark. |
| Search field and its list | Component | **Extend** | `macos/App/DesignSystem/Components/ComboBox/PVComboBox.swift` | Remote results: the page searches and the box shows them, with its own searching line, empty line, list window, keyboard, Escape and click-away. Not a second popup list. S11-36 gives it a public init with a caller-drawn empty line, tells a remote caller when the box resets the field, and carries the field's environment into its list (user, 2026-10-25). No visual change. |
| Hit row | Recipe | **Extend** | `macos/App/DesignSystem/Recipes/OmnibarHitRow/PVOmnibarHitRow.swift`, filled by `macos/App/Features/Catalog/Search/CatalogSearchHitRow.swift` | Inside the ComboBox list as each option's row. The ComboBox draws the highlight and the checkmark gutter. S11-36 removes the row's unused `selected` fill and puts the spoken label on the row; no visual change. The kind Badge stays. |
| Cancel search | Component | Ship | `macos/App/DesignSystem/Components/IconButton/PVIconButton.swift` (`.dismiss`, `.sm`) | Same as the Source page's type editor cancel. |
| Band, Contested, Decided marks | Component | Ship | `macos/App/DesignSystem/Components/Badge/PVBadge.swift` | As today. |
| Row readout | Snowflake | **Extend** | `macos/App/Features/Promote/PromoteReadout.swift` `PromoteReadoutBadges` | The badge text and the spoken band both come from one L10n band wording (S11-36 D-9): "Little evidence" for a chosen record at band None (FR-7), on the row and the sheet title. Same neutral tone. |
| Toolbar search panel | Snowflake | Ship | `macos/App/Features/Workspace/OmnibarResultsHost.swift` | Not used on this page and not redrawn. |

**Non-goals.** No new kit component. No action item, divider or header in Select. No popup or overlay beside the ComboBox's own list. No Callout, Chip or extra mark for a found or differently named record. No band or score per hit. No visual change to the hit row recipe.

---

## 7. Handoff

When the board is agreed: add a **Board notes** section here listing the frames and anything the board changed, record the board link in the S11-36 plan's **Board** line, move S11-36 from `Blocked — on board S11-D3` to `Planned` with `plan-pr`, and move this brief to the spike's `design/archive/` when S11-36 lands, updating [`README.md`](README.md).

---

## 8. Board notes (agreed 2026-10-25)

Checked against `Promote flow.dc.html`. Where the board differs from this brief, the board wins and the S11-36 plan copies it.

| Frame | Requirement | The board |
| --- | --- | --- |
| 20 | FR-1 | As written. The entry is a 14 pt search icon and the text in secondary ink, 8 pt apart. |
| 21 | FR-2, FR-4 | As written. The Cancel search button is the ghost × IconButton. Specimens draw the field unfocused; the live field holds the focus. |
| 22, 22b, 22c | FR-3, FR-4 | The fixtures are redrawn on the board's obituary ("Marion"), not "Polly". A typed ref accents the matched refs. The list is capped at 290 pt (the kit's 288). |
| 23, 23b | FR-4 | No hits reads "No people match “Mairon”". A failed search shows the search's own message; the brief named none. |
| 24 | FR-5 | As written: Weak, Contested, "Your choice"; the found record after the offered one and before New, checked; no "found" mark. |
| 25, 25s | FR-7, FR-8 | "Little evidence" for every record the researcher chose at band None. A suggested None (Ken's PER-7KG40) keeps "No match". S11-D2's E03, a menu choice at None, now reads "Little evidence" too. |
| 26 | FR-6 | A click outside closes the list and leaves the field, dropping the typed text; the search stays until a pick or Cancel. Escape closes the list, then clears the text, then closes the search. |
| 26b | FR-9 | As written. |

