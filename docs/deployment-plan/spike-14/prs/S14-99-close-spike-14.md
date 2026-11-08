# S14-99 — Close Spike 14 as a decisions note

**Status:** Draft
**Spike / slice:** Spike 14 · Spike close
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no chrome change: docs only, no view changes)
**Board:** —
**Depends on:** every other Spike 14 row, all Landed or Descoped at the planning base (last: S14-42, landed [#748](https://github.com/mendahu/provenencia/pull/748))
**Blocks:** —
**Branch:** `s14-99-close-spike-14`
**Version bump:** none — docs only ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`archive-docs`](../../../../.agents/skills/archive-docs/SKILL.md)
**Rules that bind:** —
**Context:** [`archive-docs`](../../../../.agents/skills/archive-docs/SKILL.md) "Spike close — checklist"; precedent commits `878b45ea` (Spike 11 close) and `35ec7443` (Spike 12 close); [`archive/spike-12/README.md`](../../archive/spike-12/README.md) (shape to match); working rule "Even spikes are odds and ends"
**Planned:** 2026-11-08 against `main` @ `2ac041a4`
**Review:** — (not reviewed yet)

### Implementer — read this first

1. **This file is the whole task.** Read it top to bottom before you change code. Then read every file under **Current state** and load every skill under **Skills to load**.
2. **Check the base.** Compare `main` with the **Planned** commit. If a file under **Current state** has changed so a row is no longer true, stop and report which rows.
3. **Decisions are made.** Do not reopen a **Decision** or a **Scope** line, and do not swap in a design you like better. If one cannot work, stop: say which ID it is, what you found (path, line, output), and what you would need. Do not improvise around it.
4. **Do only In.** Out stays out, even when it is one line away. Do not refactor, rename, or tidy code the plan does not name.
5. **Work the Steps in order.** Run each step's check before you start the next. A failing check is fixed in that step.
6. **Hit a Stop condition → stop.** Report it. Do not work around it.
7. **Done is the Definition of done.** Report each box with its evidence: the command and its result, the test name, or the manual check you ran.
8. **Docs-only lines in the DoD are part of the PR.** Do not skip them.

---

## 1. Objective

Spike 14 is finished: every PR in its map has landed or been descoped. This PR closes it the way Spikes 11 and 12 were closed. `docs/deployment-plan/spike-14/` becomes `docs/deployment-plan/archive/spike-14/` *(new)*, holding a short decisions note and the one design doc that code and skills still build on (`summary-inputs.md`). Every open follow-up from the ship logs moves to an ideas file or is dropped with a reason. Every link into the deleted folder is retargeted.

**Why now.** S14-42 (#748) was the last open row. Spike 16, the next odds-and-ends spike, is being planned and takes one-off work from here on (user, 2026-11-08).

**What stays.** No code, test, string or schema changes. Code comments that name a brief id (`S14-D1` in `PVPageColumns.swift`, `S14-D9` in `PVColor.swift`, `S14-D1` in `macos/App/DesignSystem/README.md`) stay: they name an id and link nothing ([`archive-docs`](../../../../.agents/skills/archive-docs/SKILL.md) "Retarget"). The dogfood bar pass is skipped (user, 2026-11-08).

## 2. User stories

**US-1.** As a contributor, when I open `docs/deployment-plan/`, I find Spike 14 among the completed spikes as one decisions note, so I learn what it decided without reading 40 plans.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | `ls docs/deployment-plan/archive/spike-14/` *(new)* | Exactly `README.md` and `summary-inputs.md` |
| AC-1.2 | `ls docs/deployment-plan/` | No `spike-14` entry |
| AC-1.3 | `docs/deployment-plan/README.md` § Completed | Spike 14 is the first row, linking `archive/spike-14/`; § Current names no Spike 14 and says Spike 16 is being planned |
| AC-1.4 | `wc -l docs/deployment-plan/archive/spike-14/README.md` | 40 – 90 lines, with sections Decisions, Refused, Leftovers (D-2) |

**US-2.** As a contributor, when I follow any link in the repo, it does not land on a deleted Spike 14 file.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | The link script (T-1) over the branch, against its run on the planning base | No broken link the base did not already have |
| AC-2.2 | `rg -n "spike-14/" --glob '!docs/deployment-plan/archive/spike-14/**' .` | Only lines of the form `archive/spike-14` |
| AC-2.3 | `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-13/prs/S13-10-header-and-sections.md docs/deployment-plan/spike-13/prs/S13-14-gallery-section.md` | 0 errors |

**US-3.** As a contributor, no follow-up a Spike 14 review raised disappears with the ship logs.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | Each "Re-home" row of §6 Leftovers | Its text is in the named ideas file and section |
| AC-3.2 | Each "Drop" row of §6 Leftovers | It is a line in the decisions note's Leftovers table with its reason |

## 3. Scope

| In | Covered by |
| --- | --- |
| Decisions note `archive/spike-14/README.md` *(new)* | STEP-1, AC-1.1, AC-1.4 |
| `summary-inputs.md` moved to the archive, PR sequence stripped, links fixed | STEP-2, AC-1.1, AC-2.1 |
| Open ship-log follow-ups re-homed in ideas files | STEP-3, AC-3.1 |
| Inbound links retargeted (18 links in 10 files) | STEP-4, AC-2.1 – AC-2.3 |
| `docs/deployment-plan/README.md` Current and Completed | STEP-5, AC-1.3 |
| Delete the rest of `spike-14/`: map, map review log, `completed.md`, `README.md`, both audits, `event-place-pictures.md`, `oversized-jpeg-thumbnails.md`, `design/`, `ux/`, every plan, review log and ship log | STEP-6, STEP-8, AC-1.2 |

| Out | Home |
| --- | --- |
| The dogfood bar pass | Not planned (user, 2026-11-08: skipped) |
| Triage of the audits' unscheduled items (audit 1 L-1 – L-17 and nits, audit 2 B-25 – B-39) | Not planned (user, 2026-11-08: deleted without triage; git history keeps them) |
| The two pre-existing broken links in S13-23 and S13-25 (`../../../../ideas/matching-test-suite.md`, one level too deep) | Spike 13's close; not this spike's files |
| Opening Spike 16's folder or map | Spike 16's `plan-spike` (being planned, user 2026-11-08) |
| Fixing any re-homed follow-up | The ideas file that holds it |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Map | `docs/deployment-plan/spike-14/deployment-plan.md` | Every row Landed or Descoped except S14-99 (`Scoped`). Working rules from line 108 are the source for the Decisions list in §6 |
| Spike folder | `docs/deployment-plan/spike-14/` | `README.md`, `deployment-plan.md`, `deployment-plan.review.md`, `completed.md`, `app-health-audit.md`, `app-health-audit-2.md`, `event-place-pictures.md`, `oversized-jpeg-thumbnails.md`, `summary-inputs.md`, `design/` (README + 10 archived briefs and one review log), `ux/` (README + 3 notes, all shipped), `prs/` (38 plans with their review and ship logs) |
| Live design doc | `docs/deployment-plan/spike-14/summary-inputs.md` | The only full description of the input index (`graphcache.InputScope`, the `searchindex` table, `ReprojectChanged`). Line 3 links the map and tells the PR story; lines 44 and 123 link `../../ideas/structural-debt.md` |
| Graduated docs | `event-place-pictures.md`, `oversized-jpeg-thumbnails.md` | Already in `docs/conclusion-layer-data-model.md` (gallery keys, preferred pointer), `docs/seeded-vocabulary.md` (`event_picture`, `place_picture`) and `core/raster/jpegdc.go`. Their "Stays later" / "Not this design" items have no other home except signature and the Timeline picture |
| Audits | `app-health-audit.md`, `app-health-audit-2.md` | Linked for provenance only by `docs/ideas/structural-debt.md:3`, `docs/ideas/promote-ranking.md:5`, `docs/ideas/share-packages.md:49`, `docs/deployment-plan/archive/spike-12/source-structure.md:68`. The ideas rows describe themselves |
| Inbound links | 10 files, listed in §6 Retarget | Break when `spike-14/` goes (measured by deleting it on a scratch worktree) |
| Index | `docs/deployment-plan/README.md:54` | "Spike 14 — Odds and ends is open" paragraph with four links into `spike-14/`; Completed table starts at Spike 12 |
| Plan-doc lint | `scripts/check-plan-docs.py`, run by `scripts/check-pr.sh` on changed plans only | Checks every relative link in a plan. Two plans this PR edits (S13-10, S13-14) must stay at 0 errors |
| Guidance names | `scripts/check-guidance-names.py` | Run by `check-pr.sh`; no skill or rule names a Spike 14 path today |

**Precedent.** `878b45ea` "Close Spike 11 as a decisions note." and `35ec7443` "Close Spike 12: …" Copy their shape: one `archive/spike-N/README.md` with Decisions, Refused, Leftovers; one kept design doc beside it (`comparison-layers.md`, `source-structure.md`); everything else deleted; the README's Completed row is one line. Do not copy Spike 12 moving its audits to another spike: here they are deleted (D-4).

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | The archive keeps `README.md` and `summary-inputs.md` only. | `summary-inputs.md` is the one doc code and skills build on (`add-searchable-kind` "records its inputs through `graphcache.InputScope`", `catalog-deletes.md` "the input index") and no domain doc describes the index whole. The pictures and JPEG docs have graduated (§4). | Keeping the pictures or JPEG docs — their contract is in domain docs and code; graduating `summary-inputs.md` into a new `docs/*.md` — larger than a close, and Spike 11/12 kept their design doc in the archive the same way |
| D-2 | The decisions note is the text in §6 "Decisions note", copied as written, with the close date filled in. | The decisions are product and engine calls already made in the map's working rules; the implementer transcribes, it does not distill. | Letting the implementer distill from the map |
| D-3 | Every open ship-log follow-up is re-homed or dropped exactly as §6 Leftovers says. | Triaged at planning against the code and ideas files (2026-11-08). | Re-triaging during the PR |
| D-4 | Both audits are deleted with no triage of their unscheduled items. The three ideas intros and `source-structure.md:68` name the audit by date and commit instead of linking it. | User, 2026-11-08. | Moving the audits to the archive; re-homing each low item |
| D-5 | No bar pass. | User, 2026-11-08. | Walking bar outcomes 1 – 34 |
| D-6 | § Current says Spike 16 (odds and ends) is being planned. If a `spike-16/` folder exists under `docs/deployment-plan/` on the branch's base, the line links it; otherwise it has no link. | User, 2026-11-08: Spike 16 is being planned and arrives soon. | Opening a Spike 16 skeleton here |
| D-7 | Retargets follow the table in §6 Retarget exactly. A link to the map or a plan goes to the decisions note; a link to a brief becomes the plain brief id, or the decisions note where no live doc is linked beside it. | Briefs are deleted at close and the boards live in Claude Design; the gallery slot decisions S13-14 cites (S14-02 D-1 – D-5) are in the decisions note. | Linking git commits for deleted files |
| D-8 | `docs/dogfood/README.md:15` (the Batch 7 notes line) is deleted, not retargeted. | The notes shipped (S14-28, S14-36, S14-37, S14-39) and the folder goes. | Pointing it at the archive |
| D-9 | Two commits: "Close Spike 14 as a decisions note." holds STEP-1 to STEP-6 (a STEP-7 fix adds its own commit); the last commit, made after `ship-pr` writes the ship log, deletes `prs/S14-99-close-spike-14.md`, its `.review.md` and `.ship.md`, leaving no `spike-14/` folder. | `archive-docs`: "The close PR (S#-99) deletes its own plan, review log, and ship log in its last commit, the one that records the ship log." | One commit — the ship log does not exist yet |

**Rules this plan bends.** None.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `docs/deployment-plan/archive/spike-14/README.md` *(new)* | file | `docs/deployment-plan/archive/spike-12/README.md` | It is the same shape, for another spike |

**Builds on.** [`archive-docs`](../../../../.agents/skills/archive-docs/SKILL.md) spike-close checklist; the Spike 11 and 12 close commits.
**Deletes.** Everything in `docs/deployment-plan/spike-14/` except `summary-inputs.md` (moved). The § Current Spike 14 paragraph in `docs/deployment-plan/README.md`. `docs/dogfood/README.md:15`.
**Next adds.** Spike 16's close copies this plan's shape.

### Reach

None — docs only; no new path into code.

### Scale

None — no write, list or loop changes.

### Edges

None — no rule, comparison or algorithm changes.

### Schema

None — no catalog change.

### Go core

None — docs only.

### FFI

None — docs only.

### macOS

None — docs only.

### Strings

None — no copy changes.

### Decisions note

Write `docs/deployment-plan/archive/spike-14/README.md` *(new)* with this text. Replace `YYYY-MM-DD` with the date of the close commit.

```markdown
# Spike 14 — Odds and ends

**Done** (closed YYYY-MM-DD). The first odds-and-ends spike: eight batches of small features, audit fixes and dogfood UX, each landing on its own, ending with summary rows that record their inputs so a write finds what it moved by lookup. Live models: [`conclusion-layer-data-model.md`](../../../conclusion-layer-data-model.md), [`seeded-vocabulary.md`](../../../seeded-vocabulary.md), [`interpretation-layer-data-model.md`](../../../interpretation-layer-data-model.md), [`catalog-deletes.md`](../../../catalog-deletes.md), [`macos-client-patterns.md`](../../../macos-client-patterns.md), [`structured-name-model.md`](../../../structured-name-model.md), [`matching-model.md`](../../../matching-model.md). Design reasoning: [`summary-inputs.md`](summary-inputs.md).

## Decisions

- **Even spikes are odds and ends.** Small features, refinements and fixes land in batches, each a slice with its own Check; large features go to odd spikes.
- **Events and Places have pictures.** One locked depiction Property per kind, `event_picture` and `place_picture`, both labelled Picture; `likeness` stays the Person key. New seeds reach new catalogs only.
- **The gallery layer is named gallery.** "Likeness" stays only for the Person key, its copy and the portrait; `canonical_entities.preferred_likeness_observation_id` keeps its name.
- **One gallery for every kind.** The gallery reads a list of gallery keys in `core/vocabkey`; one preferred pointer, one lead rule and one gallery slot on `ConclusionDetailBody` serve Person, Event and Place. An Event or Place chooses its thumbnail as a Person chooses its portrait.
- **A depiction is drawn as a rectangle.** A Citation's own region keeps all seven tools; an old polygon depiction loads and saves as it is.
- **Oversized JPEGs get a reduced decode of our own.** Over the 64-megapixel budget, a pure-Go decoder reads only DC coefficients at 1/8 scale, bounded and fuzzed, so a Windows build reuses it. Other formats over the budget stay refused.
- **A Citation belongs to one Evidence graph** (`citations.graph_id`), never moves, and is deleted with its graph.
- **A hidden graph or Source page goes stale; it does not reload.**
- **File work happens before the catalog session.** One staged-object type in `ingest` streams, hashes, sniffs and decodes outside `catalogsession.Do`.
- **Go decides what the app may hand macOS to open.** Swift keeps no second allowlist.
- **Kit components announce themselves.** A danger confirm's default is Cancel; a dirty form asks before a link inside it navigates away.
- **Every audited write returns its effects**, held by a test in `api/ffi`. Sidebar counts are an effects list, and every list reaches its keys through its descriptor.
- **FakeStore is a seeded fake, not a second engine.**
- **Proposals are liberal; the researcher decides.** Other-link support is 0.75, so a thin name plus one lined-up link clears Weak.
- **A nickname folds into its set's formal name and reappears as a nick part:** *James “Jim” Robins*. Former surnames show in parentheses, italic on the titles.
- **One width breakpoint: 1600pt of pane.** Pages adopt Page columns in their own redesigns.
- **Boolean is a first-class value type.** Single-only, a Yes / No segmented control with nothing chosen at first, no Negative checkbox.
- **Smaller calls.** Conflict brackets honour cardinality; Tab reaches a following text area through `PVTabPair`; place relationship rows read the inverse label at the far end; creating a bridge between two handles files it; the Persons sort ignores honorific prefixes.
- **The structural fix comes before the patch.** A bug that needs a patch at each call site waits for the mechanism to be replaced.
- **Summary rows record their inputs; invalidation is a lookup.** An input is a handle's values, one bridge's links at one end, or an association's values; a row is its own input; isolated composition is the oracle. One input index, owned by `searchindex`, with no foreign key; every write reports what it moved through `summary_handles`. The hand-written inverse (`Dependents`, `AssociationEndpoints`, `linkEnds`, `SetDisplayDeps`) is gone.

## Refused

A transparent-polygon depiction crop. A graph name in the canvas header. A per-kind thumbnail column or RPC. A Property named `depiction`. A separate Page columns PR (built with Spike 13's Person page). Promote picture wording (the evidence sheet went in Spike 13).

## Leftovers

(the table from the plan's §6 Leftovers, as below)
```

The Leftovers table in the note has columns `| Item | Home |` and these rows, in this order:

| Item | Home |
| --- | --- |
| A loading caption while the viewer opens; a drawing hint that says "tools above" for one tool; EXIF orientation; a larger cover derivative; pictures as matching evidence for an Event or Place | [`depictions-and-likenesses.md`](../../../ideas/depictions-and-likenesses.md) |
| A reloading sign on a revisited canvas | [`evidence-graph-followups.md`](../../../ideas/evidence-graph-followups.md) |
| A failed sidebar recount is silent | [`shared-list-helpers.md`](../../../ideas/shared-list-helpers.md) |
| A failed bind after a Property create | [`properties-and-terms.md`](../../../ideas/properties-and-terms.md) |
| "Soby-Kirke" misses "Søby Kirke" | [`international-names.md`](../../../ideas/international-names.md) |
| A name part length bound on Validate and Insert | [`share-packages.md`](../../../ideas/share-packages.md) "Harden before sharing" |
| A field relabel reaching Source pages beside its effects; a new-spelling join reprojecting the city; a two-term filed-link set and its repeated link key | [`structural-debt.md`](../../../ideas/structural-debt.md) |
| The two app-health audits' unscheduled items | Dropped with the audits (user, 2026-11-08); in git history |
| Polygon vertex add, keyboard nudge, VoiceOver per vertex | Dropped (user, 2026-11-05) |
| Dropped: a thumbnail choice reloading three lists, a term create reloading Properties (2 ms), a broken oversized JPEG re-decoded per response (off the session since S14-12), copy nits ("A smaller copy of the file will open", a generic "Discard changes?", a race with another window's unlink), a test for the localizable script, a split of `check-guidance-names.py`, code nits in `evidencegraphs`, `FakeStore`, `L10n.Promote` | Not worth a PR; no measured cost (2026-11-08) |

### Leftovers

Re-home (STEP-3). Each row adds the text to the named file and section. Write it in the file's own style (a bullet or a section with a short paragraph; a table row in `structural-debt.md`, with `S14-NN ship log` in its source column as rows 24 – 28 do).

| From | Item | File · section | Text to add (substance) |
| --- | --- | --- | --- |
| S14-03 UX-4, UX-5 | Composer depiction caption | `docs/ideas/depictions-and-likenesses.md` § Still open | While the viewer loads, the depiction caption falls through to the empty case ("use a region button"); it needs a loading string. `depictionDrawingHint` says "tools above" though a depiction has one tool. Both wait for the composer board's next pass |
| `oversized-jpeg-thumbnails.md` "Not this design" | EXIF orientation | same § Still open | No JPEG honours EXIF orientation; fix it for every JPEG, not only oversized ones |
| same | Cover derivative | same § Still open | Whether a 256 px thumbnail is enough for a Source cover from a very large scan |
| `event-place-pictures.md` "Stays later" | Pictures as matching evidence | same § Still open | `subjectcompare` gives Event and Place pictures no weight |
| S14-11 UX-1 | Reloading sign | `docs/ideas/evidence-graph-followups.md`, new § "No reloading sign on a revisited canvas" | `isRereadingGraph` is drawn only in the error callout (`EvidenceGraphView.swift`); a revisited stale canvas shows old cards with no sign it is rereading. Needs a brief |
| S14-16 UX-1 | Silent recount failure | `docs/ideas/shared-list-helpers.md`, new § after "Source page section counts differ" | `WorkspaceView` reads `lastRefreshError` only in its first `.task`; a sidebar recount that fails after a write says nothing. Shape: observe it with `.onChange` |
| S14-17 UX-2 | Failed bind after create | `docs/ideas/properties-and-terms.md`, new § "A failed bind after a Property create" | When the bind after a Property create fails, the dialog stays open and Retry calls `createProperty` again, which fails on the duplicate key (`PropertiesModel.swift`). Retry should bind the Property already made |
| S14-21 CON-2 | Hyphenated folded place | `docs/ideas/international-names.md` § Scripts, sound, and dates | "Soby-Kirke" (hyphen, no ø) does not find "Søby Kirke register": S14-21's D-4 splits phrases on spaces only |
| S14-26 | Name part length | `docs/ideas/share-packages.md` § Harden before sharing (its table) | `namevalue` bounds a part's length on assemble only; `Insert` and `Validate` are not bounded (`core/namevalue/namevalue.go` doc comment) |
| S14-17 MNT-7 | Relabel beside effects | `docs/ideas/structural-debt.md`, the existing row sourced "S13-37 plan" (`.deletedSource` vs `WriteEffects.deleted`) | Append: a field relabel or delete also reaches Source pages through a Swift mutation beside its effects (`macos-client-patterns.md` rule 2) |
| S14-21 PRF-1 | New-spelling join | `docs/ideas/structural-debt.md`, the existing row sourced "S14-33 ship log" (first `part_of` from a city) | Append: a join that adds a new spelling to the 6,000-event city also reprojects every dependent (2.1 s at S14-21, 1.33 s after S14-40) |
| S14-22 two-term set, MNT-9 | Filed-link set | `docs/ideas/structural-debt.md`, new row, source "S14-22 ship log" | A two-term bridge filed mid-Done is missing from an already-loaded filed-link set for its other term (low confidence, not reproduced); `bridgefile.go` repeats `graphalign`'s link key. One keyed set in `connectrules` would serve both |

Drop (no file change; listed in the note's Leftovers): S14-02 PRF-1, S14-10 UX-2 / MNT-5 / MNT-6, S14-13 UX-4, S14-14 CON-2 / UX-3, S14-15 MNT-2, S14-17 PRF-1, S14-20 MNT-3, S14-22 MNT-10, S14-23 PRF-1, S14-29's vertex items.

Already resolved (no action): S14-01 UX-1 and UX-2 (S14-02), S14-03 UX-6 (S14-14 D-16), S14-13's flaky tests (S14-03's CI round), S14-33 FU-1 (S14-42), S14-33 PRF-3 / PRF-6 / PRF-4 / FU-3 / FU-4 and S14-37's two items and S14-09's four lines (already in ideas files), the three `ux/` notes (shipped).

### Retarget

"The note" below means `docs/deployment-plan/archive/spike-14/README.md` *(new)*, linked by a path relative to the file being edited. "Link L → X" means the Markdown link whose target is L. Targets are written without brackets so this plan's own link check does not read them.

| File:line | Today | Becomes |
| --- | --- | --- |
| `docs/deployment-plan/README.md:54` | Spike 14 open paragraph | Removed; see STEP-5 |
| `docs/deployment-plan/archive/spike-12/source-structure.md:68` | Link "audit" → `../../spike-14/app-health-audit.md` | Plain text "the 2026-10-27 app-health audit", no link |
| `docs/deployment-plan/spike-13/deployment-plan.md:61` | "brief" + link "S14-D1" → `../spike-14/design/archive/S14-D1-wide-layout.md` | Plain "brief S14-D1" (§ 9 of `macos-client-patterns.md` is linked beside it) |
| `docs/deployment-plan/spike-13/deployment-plan.md:355` | Link "S14-D1" → same brief | "S14-D1 (" + link "Spike 14 decisions" → `../archive/spike-14/README.md` + ")" |
| `docs/deployment-plan/spike-13/deployment-plan.md:474` | Link "S14-02" → `../spike-14/deployment-plan.md#s14-02--pictures-in-the-app` | "S14-02 (" + link "Spike 14 decisions" → `../archive/spike-14/README.md` + ")" |
| `docs/deployment-plan/spike-13/person-detail.md:102` | Link "Spike 14" → `../spike-14/deployment-plan.md` | Link "Spike 14" → `../archive/spike-14/README.md` |
| `docs/deployment-plan/spike-13/design/archive/S13-D5-gallery-and-viewer.md:8` | "the S14-D2 Pictures section, " + link "archived brief" → `../../../spike-14/design/archive/S14-D2-event-page-pictures.md` | "the S14-D2 Pictures section" (link and comma removed) |
| `docs/deployment-plan/spike-13/design/archive/S13-D5-gallery-and-viewer.md:9` | Link with code label `../../spike-14/deployment-plan.md` → `../../../spike-14/deployment-plan.md#working-rules` | Link "Spike 14 decisions" → `../../../archive/spike-14/README.md` |
| `docs/deployment-plan/spike-13/prs/S13-10-header-and-sections.md:6` | Link "S14-D1" → `../../spike-14/design/archive/S14-D1-wide-layout.md` | Link "S14-D1" → `../../archive/spike-14/README.md` |
| `docs/deployment-plan/spike-13/prs/S13-14-gallery-section.md:15` | "the S14-02 plan (" + link → `../../spike-14/prs/S14-02-pictures-in-the-app.md` + ") D-1 – D-5" | "the S14-02 plan's D-1 – D-5 (now in " + link "Spike 14 decisions" → `../../archive/spike-14/README.md` + ")" |
| `docs/dogfood/README.md:15` | Batch 7 notes line | Line deleted (D-8) |
| `docs/ideas/depictions-and-likenesses.md:3` | Link "Spike 14" → `../deployment-plan/spike-14/` | Link "Spike 14" → `../deployment-plan/archive/spike-14/` |
| `docs/ideas/structural-debt.md:3` | "now at " + link to `app-health-audit.md` | "(2026-10-27, at `731e423`; deleted when Spike 14 closed, in git history)" |
| `docs/ideas/promote-ranking.md:5` | "now at " + link to `app-health-audit-2.md` | "(2026-10-28, at `3cc609c`; deleted when Spike 14 closed, in git history)" |
| `docs/ideas/share-packages.md:49` | "now at " + link to `app-health-audit-2.md` | Same replacement as `promote-ranking.md` |

## 7. Steps

**STEP-1 — Write the decisions note**
Files: `docs/deployment-plan/archive/spike-14/README.md` *(new)*. Skill: `archive-docs`.
Do: write the text in §6 "Decisions note", with the Leftovers table in place of its placeholder line.
Check: `wc -l` is 40 – 90; every link in the file resolves (T-1 run on the file's folder).

**STEP-2 — Move the input-index design doc**
Files: `docs/deployment-plan/spike-14/summary-inputs.md` → `docs/deployment-plan/archive/spike-14/summary-inputs.md` *(new)* (`git mv`).
Do: replace line 3 with one sentence that says what the doc is: "Design of the input index: summary rows and search documents record their inputs, so a write finds what it moved by lookup. Decided in Spike 14 (user, 2026-11-07)." Change `../../ideas/` to `../../../ideas/` (lines 44 and 123). Remove any other link into the deleted map. Leave every other section as it is, "Measured at planning" included.
Check: `rg -n "deployment-plan.md|S14-38"` over the moved file *(new)* finds no link to the map; T-1 finds no broken link in the file.

**STEP-3 — Re-home the follow-ups**
Files: `docs/ideas/depictions-and-likenesses.md`, `evidence-graph-followups.md`, `shared-list-helpers.md`, `properties-and-terms.md`, `international-names.md`, `share-packages.md`, `structural-debt.md`.
Do: each Re-home row of §6 Leftovers. Before writing a row, `rg` the file for the item; if it is already there, leave it and note it in the PR body.
Check: AC-3.1 — each item is found by `rg` in its file.

**STEP-4 — Retarget inbound links**
Files: the ten files in §6 Retarget other than `docs/deployment-plan/README.md`.
Do: each row, exactly.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-13/prs/S13-10-header-and-sections.md docs/deployment-plan/spike-13/prs/S13-14-gallery-section.md` reports 0 errors.

**STEP-5 — The deployment-plan index**
Files: `docs/deployment-plan/README.md`.
Do: in § Current, delete the Spike 14 paragraph and add after the Spike 13 paragraph: "Spike 16 — Odds and ends is being planned. Until its map is open, a one-off PR waits for it." (linked per D-6). In § Completed, add as the first row (its link targets `archive/spike-14/`): `| Spike 14 (linked) | Odds and ends in batches: Event and Place pictures, effects on every write, a boolean value type. Summary rows record their inputs. |`
Check: AC-1.3.

**STEP-6 — Delete the spike folder**
Files: `docs/deployment-plan/spike-14/`.
Do: `git rm -r` everything left in it except `prs/S14-99-close-spike-14.md` and `prs/S14-99-close-spike-14.review.md`. Commit STEP-1 to STEP-6 as one commit: "Close Spike 14 as a decisions note." with a body in the shape of `878b45ea`'s (what the archive keeps, what is deleted, where follow-ups went, that inbound links point at the archive).
Check: `ls docs/deployment-plan/spike-14/prs` lists only the S14-99 files.

**STEP-7 — Sweep**
Files: none changed unless the sweep finds a miss.
Do: run T-1, T-2 and T-3.
Check: all pass. A miss found here is fixed in a follow-up commit on this branch with the same rules as STEP-4.

**STEP-8 — Delete this plan (last commit)**
Files: `docs/deployment-plan/spike-14/prs/S14-99-close-spike-14.md`, `.review.md`, `.ship.md`.
Do: after `ship-pr` records the ship log, `git rm` all three in one commit: "Delete the Spike 14 close plan and its logs." (D-9). No Landed note and no `completed.md` paragraph: the PR body carries any deviation.
Check: `test ! -e docs/deployment-plan/spike-14` succeeds; AC-1.2.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Script | Inline link check: for every `*.md` outside `.git`, each relative Markdown link target with the `#fragment` stripped exists. Run it on `main` @ `2ac041a4` and on the branch; `comm -13` the sorted outputs | Branch has no line the base lacks | AC-2.1, AC-1.1, D-7 |
| T-2 | Script | `rg -n "spike-14/" --glob '!docs/deployment-plan/archive/spike-14/**' .` | Every hit contains `archive/spike-14` | AC-2.2 |
| T-3 | Script | `scripts/check-pr.sh` | `check-pr: PASS` (plan-docs over S13-10 and S13-14, guidance-names, diff-check) | AC-2.3, D-1 |
| T-4 | Script | `ls docs/deployment-plan/archive/spike-14/ docs/deployment-plan/` | AC-1.1 and AC-1.2 hold | AC-1.1, AC-1.2 |
| T-5 | Script | `rg` each Re-home item in its file | Found once | AC-3.1, D-3 |
| T-6 | Script | `sed -n '/^## Completed/,$p'` and `sed -n '/^## Current/,/^## Completed/p'` on `docs/deployment-plan/README.md` | Spike 14 is the first Completed row; Current has no "Spike 14" and has the Spike 16 line | AC-1.3, D-6 |
| T-7 | Script | `wc -l` and `grep -n '^## '` on the note *(new)* | 40 – 90 lines; headings Decisions, Refused, Leftovers | AC-1.4, D-2 |
| T-8 | Script | `rg` each Drop item's words in the note's Leftovers table *(new)* | Each found | AC-3.2, D-3 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: plan-docs on the two Spike 13 plans this PR edits.

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] Sweep: each deleted path `rg`'d across `.agents/`, `.cursor/rules/`, `docs/` and `macos/` with no link left (T-1, T-2); `scripts/check-guidance-names.py` passes (check-pr runs it)
- [ ] Version bump: none
- [ ] Decisions note matches §6 text; Leftovers table matches §6
- [ ] Every Re-home row is in its ideas file; every Drop row is in the note
- [ ] docs/deployment-plan/README.md: Spike 14 in Completed, Spike 16 in Current (D-6)
- [ ] Last commit deletes this plan, its review log and its ship log; no spike-14/ folder remains; no Landed note, no completed.md paragraph, no map status change (archive-docs)
- [ ] PR description: what shipped, AC evidence, deviations, the Drop list, and that the bar pass was skipped (user, 2026-11-08)
```

## 10. Manual check

1. Open `docs/deployment-plan/README.md` on the branch in GitHub's file view. Spike 14 is the first Completed row and its link opens the decisions note.
2. In the decisions note, open each Live models link and the `summary-inputs.md` link. Each opens.
3. Open `docs/deployment-plan/spike-13/prs/S13-14-gallery-section.md` and follow the Spike 14 decisions link. It opens the note, which states the gallery slot decision.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| A link into `spike-14/` added on `main` after the planning base | T-1 and T-2 run on the branch, not on this plan's list |
| A Spike 14 PR re-opened or a new S14 row added after planning | Stop condition below |
| Spike 16's map links a Spike 14 file | T-1 catches it; retarget it as D-7 says |

**Stop and report if:**

- Any row of the map's PR table other than S14-99 is not `Landed …` or `Descoped …` on the branch's base.
- A file outside `docs/` (a skill, a `.cursor/rules/*.mdc`, Swift or Go source) links a Spike 14 path with a Markdown link: that is guidance, and it needs a retarget decision this plan does not make.
- `summary-inputs.md` has gained sections that describe PR order (not the mechanism) beyond line 3.
- T-1 shows a broken link the §6 Retarget table does not cover and D-7 does not decide.
- A Re-home item turns out fixed in code at the branch's base: drop it, list it in the PR body, and do not add it to the ideas file.

## 12. Open questions

None.

## 13. PR

**Title:** S14-99 — Close Spike 14 as a decisions note
**Body:** objective in one sentence; what shipped against each US; AC evidence (T-1 – T-5 output); deviations from the plan; the Drop list; "Bar pass skipped (user, 2026-11-08)"; `Plan: docs/deployment-plan/spike-14/prs/S14-99-close-spike-14.md (deleted in the last commit)`.
