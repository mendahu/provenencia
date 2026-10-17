# S10-99 — Close Spike 10

**Status:** Draft
**Spike / slice:** Spike 10 · S10-99 — Dogfood close / docs
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#s10-99--dogfood-close--docs)
**Brief:** — (no new chrome: docs only)
**Board:** —
**Depends on:** S10-01…S10-57, each landed or descoped (last: S10-39, landed [#450](https://github.com/mendahu/provenencia/pull/450))
**Blocks:** —
**Branch:** `s10-99-dogfood-close`
**Version bump:** none — docs only ([`docs/versioning.md`](../../../versioning.md): no bump for docs-only changes)
**Skills to load:** [`archive-docs`](../../../../.agents/skills/archive-docs/SKILL.md)
**Rules that bind:** —
**Context:** map row S10-99 and "Docs to update as work lands" (S10-99 line) in [`../deployment-plan.md`](../deployment-plan.md); precedent [#326](https://github.com/mendahu/provenencia/pull/326) (Spike 9 close); [`archive/spike-9/README.md`](../../archive/spike-9/README.md) for shape
**Planned:** 2026-10-17 against `main` @ `bdce4cd`
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

Spike 10 is closed. Its folder `docs/deployment-plan/spike-10/` (the PR map, `completed.md`, design briefs, PR plans, review and ship logs, and the three design notes) is replaced by one short decisions note, `docs/deployment-plan/archive/spike-10/README.md` (new). Every link that pointed into the old folder points at that note or at a live domain doc. Spike 11 is the only open spike.

**Why now.** Every Spike 10 PR has landed or been descoped; S10-39 ([#450](https://github.com/mendahu/provenencia/pull/450)) was the last. The user checked the dogfood bar on a real project (2026-10-17). The `likeness` Property on existing dogfood projects was handled by the user outside this repo (2026-10-17).

**What stays.** No code, no schema, no domain doc rule changes. The domain docs already carry Spike 10's contracts (each PR updated them as it landed). The dogfood log keeps its open notes. The ideas folder keeps its leftover notes. Spike 11's docs are not edited.

## 2. User stories

**US-1.** As a contributor, when I open `docs/deployment-plan/`, I see Spike 11 as the only open spike and Spike 10's decisions in one short note, so that I read decisions and not PR history.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | The PR is merged; `ls docs/deployment-plan` | Lists `README.md`, `archive`, `spike-11`. No `spike-10` |
| AC-1.2 | `docs/deployment-plan/archive/spike-10/README.md` (new) | Exists, is the text in §6 "Archive README", has the headings `## Decisions`, `## Refused`, `## Leftovers`, and is at most 80 lines |
| AC-1.3 | `docs/deployment-plan/README.md` | **Current** names Spike 11 only. **Completed** has a Spike 10 row above the Spike 9 row |

**US-2.** As a reader following a link in `docs/` or `.agents/`, every link that pointed into Spike 10's folder still resolves, so that I do not land on a 404.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | T-1's `rg` command | Prints only the two Spike 11 plan lines named in T-1 |
| AC-2.2 | T-2's link check over every edited file | Exit 0, no output |

**US-3.** As a planner of a later spike, every Spike 10 leftover has a home, so that nothing descoped is lost.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | The **Leftovers** table in the archive README | Every row links a file that exists (T-2) |
| AC-3.2 | `docs/ideas/depictions-and-likenesses.md` | Its Status line no longer says Spike 10 takes a crop cache or Promote's side-by-side crops; it says Promote comparing two likenesses is Spike 11 (S11-20) |

## 3. Scope

| In | Covered by |
| --- | --- |
| Write `archive/spike-10/README.md` as a decisions note (map row: "Distill this spike with archive-docs") | STEP-1, AC-1.2 |
| The three design notes' decisions go into that README; the notes are deleted (map row: "the close distill keeps their decisions and drops the PR sequence") | STEP-1, STEP-4, D-2 |
| Retarget every inbound link into `docs/deployment-plan/spike-10/` | STEP-2, AC-2.1, AC-2.2 |
| Fix the stale Status line and Windows bullet in `docs/ideas/depictions-and-likenesses.md` | STEP-2, AC-3.2 |
| `docs/deployment-plan/README.md`: Current and Completed | STEP-3, AC-1.3 |
| Delete `docs/deployment-plan/spike-10/` | STEP-4, STEP-6, AC-1.1 |

| Out | Home |
| --- | --- |
| Pulling the parked density, extract, transliteration, or Reconciliation-Claim items in at the close (map row Out) | Their homes in the archive README **Leftovers** table; not planned for this PR |
| Editing any rule in a `docs/*.md` domain doc | Not planned: each Spike 10 PR already updated its doc (D-1) |
| Editing Spike 11 docs, including the `--glob '!docs/deployment-plan/spike-10/**'` text in two landed Spike 11 plans | Not planned (D-7) |
| Removing `S10-…` ids from code comments, skills, or review lenses | Not planned: `archive-docs` "Retarget" allows ids that do not link to a deleted file (D-8) |
| Rewording `docs/ideas/*.md` beyond the lines in §6 "Retarget edits" | Not planned |
| Moving the three design notes to `docs/ideas/archive/` | Not planned (D-2) |
| A catalog upgrade tool for `likeness`, `engagement`, or `interests` on existing projects | Not planned: the user handled the dogfood projects (2026-10-17) |
| The sentence "Spike 10 and earlier cut each PR as an In / Out / Testable table" in `docs/deployment-plan/README.md` | Not planned: still true of the archived spikes (D-9) |

## 4. Current state

Verified at `bdce4cd`.

| Area | Where | Today |
| --- | --- | --- |
| Spike folder | `docs/deployment-plan/spike-10/` | `README.md`, `deployment-plan.md`, `deployment-plan.review.md`, `completed.md`, `depictions-and-likenesses.md`, `international-names.md`, `name-matching-enhancements.md`, `design/README.md`, `design/archive/` (13 briefs, one brief review log), `prs/` (plans, review logs, ship logs). All operational; nothing in `.agents/`, `.cursor/`, `core/`, `api/`, or `macos/` links into it |
| Deployment-plan index | `docs/deployment-plan/README.md` | **Current** says Spike 10 is open and links `spike-10/` and `spike-10/deployment-plan.md`. **Completed** starts at Spike 9 |
| Archive | `docs/deployment-plan/archive/` | `spike-1.md`, `spike-2/` … `spike-9/`. No `spike-10/` |
| Inbound link | `docs/matching.md` line 163 | Links `deployment-plan/spike-10/name-matching-enhancements.md` and says event and place affinity is Spike 10; that was descoped (S10-31) |
| Inbound link | `docs/dogfood/ux.md` line 49 | Links `../deployment-plan/spike-10/` |
| Inbound link | `docs/ideas/source-structure.md` line 58 | Links `../deployment-plan/spike-10/`; says box-select "is still S10-20" (it landed, [#396](https://github.com/mendahu/provenencia/pull/396)) |
| Inbound link | `docs/ideas/international-names.md` line 3 | Links `../deployment-plan/spike-10/international-names.md` |
| Inbound link | `docs/ideas/depictions-and-likenesses.md` line 3 | Links `../deployment-plan/spike-10/depictions-and-likenesses.md`; says Spike 10 takes a crop cache (descoped, S10-33) and Promote's side-by-side crops (moved to S11-20). Line 11 ("Windows.") says "The cache row is the same" |
| Inbound link | `docs/ideas/archive/name-matching-enhancements.md` line 3 | Links `../../deployment-plan/spike-10/name-matching-enhancements.md` |
| Inbound link | `docs/ideas/archive/glyph-consolidation.md` line 3 | Links `../../deployment-plan/spike-10/prs/S10-19-event-mark-sidebar.review.md` |
| Text only, not links | `docs/deployment-plan/spike-11/prs/S11-02-value-compare-core.md` line 292, `docs/deployment-plan/spike-11/prs/S11-03-value-compare-names.md` line 264 | An `rg --glob '!docs/deployment-plan/spike-10/**'` inside a landed plan's Check. Stays (D-7) |
| Domain docs already updated | `docs/matching.md`, `docs/interpretation-layer-data-model.md` (`depiction`), `docs/source-layer-data-model.md`, `docs/conclusion-reconciliation.md`, `docs/conclusion-layer-data-model.md` (Unpromote §5, `preferred_likeness_observation_id`), `docs/audit-revision-history.md` (`canonical_entity`), `docs/seeded-vocabulary.md` (`gender`, `engagement`, `interests`, `likeness`), `docs/structured-name-model.md`, `docs/deployment-plan/archive/spike-3/navigation-history.md` (list scroll) | Carry the contracts the archive README points at |
| Tests | — | None: docs only |

**Precedent.** [#326](https://github.com/mendahu/provenencia/pull/326) (`0b5d99e`) closed Spike 9: it added `archive/spike-9/README.md`, deleted the spike folder (map, `completed.md`, briefs), slimmed the index's Completed row, and fixed domain-doc pointers. Copy its file moves and its README shape. Do not copy its domain-doc edits: this PR changes no rule.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | No Spike 10 file is kept as a live contract. Everything in `docs/deployment-plan/spike-10/` is deleted. | `archive-docs` keeps a file only when skills or code implement against it. Each contract graduated: depiction → `interpretation-layer-data-model.md`, `source-layer-data-model.md`, `conclusion-reconciliation.md`; preferred portrait and Unpromote → `conclusion-layer-data-model.md`; matching → `matching.md`; last updated → `audit-revision-history.md`; list scroll → `navigation-history.md`. No skill links the folder. | Keeping `depictions-and-likenesses.md` like Spike 3's contracts — because its rules are in the domain docs, and a second copy drifts |
| D-2 | The three design notes are deleted, not moved to `docs/ideas/archive/`. Their decisions are lines in the archive README. | Their leftovers already live in `docs/ideas/depictions-and-likenesses.md`, `docs/ideas/international-names.md`, `docs/ideas/daitch-mokotoff.md`, and `docs/ideas/archive/name-matching-enhancements.md`. The map row says to archive them "as decisions". | Moving them to `ideas/archive/` — because they are spike design notes, not ideas, and would be a third copy |
| D-3 | The archive README is the text in §6 "Archive README", copied verbatim. | The README is the deliverable; its content is the decision. | Letting the implementer distill — because that is a judgment call this plan makes |
| D-4 | Each inbound link is fixed by the exact replacement in §6 "Retarget edits". Links to Spike 10 point at the directory `archive/spike-10/`, the same way links point at `archive/spike-9/`. | Exact edits leave nothing to choose. | Linking `archive/spike-10/README.md` — the index already links spike directories |
| D-5 | STEP-4 deletes everything in `docs/deployment-plan/spike-10/` except `prs/S10-99-dogfood-close.md`, `prs/S10-99-dogfood-close.review.md`, and `prs/S10-99-dogfood-close.ship.md`. STEP-6 deletes those three last, in the commit that records the ship log, so the folder is gone at the PR's head. | `ship-pr` writes and commits the ship log beside this plan until the end; deleting the plan first would break that loop. | Deleting the whole folder in STEP-4 — because the orchestrator would recreate `prs/` for its ship log |
| D-6 | No Landed note and no `completed.md` paragraph. The PR body carries any deviation. | Both files are deleted by this PR. | Writing them and deleting them in the same PR |
| D-7 | The two landed Spike 11 plans that name `docs/deployment-plan/spike-10/**` in an `rg --glob` are left as they are. | They are landed history, not links; an exclude glob for a missing folder is harmless. | Editing landed Spike 11 plans |
| D-8 | `S10-…` ids in code comments, `.agents/skills/`, and review lenses stay. | `archive-docs` "Retarget": ids are fine when they do not link to a deleted file. T-1 proves none link. | A sweep of ids |
| D-9 | `docs/deployment-plan/README.md` changes only the **Current** paragraph and adds one **Completed** row. | The rest of that file is still true. | Rewriting "How a spike is planned" |
| D-10 | Gate: `scripts/check-pr.sh`. On macOS it passes. Elsewhere the one accepted failure is `xcode-test (needs macOS)`, because this PR changes no Swift or Xcode file; CI runs Xcode on the PR. | The gate's other checks answer for docs. | Skipping the gate |

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `docs/deployment-plan/archive/spike-10/README.md` *(new)* | file | `docs/deployment-plan/archive/spike-9/README.md` | Same shape, one per closed spike (`archive-docs`) |

**Builds on.** `archive-docs` "Spike close — checklist"; the Spike 9 close ([#326](https://github.com/mendahu/provenencia/pull/326)).
**Deletes.** `docs/deployment-plan/spike-10/` and everything in it (D-1, D-5).
**Next adds.** The Spike 11 close follows the same shape.

### Schema

None — no catalog change.

### Go core

None — docs only.

### FFI

None — docs only.

### macOS

None — docs only.

### Strings

None — no user-facing copy.

### Archive README

`docs/deployment-plan/archive/spike-10/README.md` *(new)*, verbatim:

```markdown
# Spike 10 — Dogfood UX

**Done.** The annoyances logged while entering real research are gone, names match across accents and abbreviations, and a face on a photo is evidence. Live models: [`matching.md`](../../../matching.md), [`interpretation-layer-data-model.md`](../../../interpretation-layer-data-model.md), [`source-layer-data-model.md`](../../../source-layer-data-model.md), [`conclusion-layer-data-model.md`](../../../conclusion-layer-data-model.md), [`conclusion-reconciliation.md`](../../../conclusion-reconciliation.md), [`audit-revision-history.md`](../../../audit-revision-history.md), [`seeded-vocabulary.md`](../../../seeded-vocabulary.md), [`navigation-history.md`](../spike-3/navigation-history.md).

## Decisions

- **The Evidence graph draws the full card at every zoom.** Cards draw in AppKit, so the camera stays 0.25 to 1.25 with no simplified face, every card on the Source is mounted, and no cover hides the first layout. The grid is one tile. The document grows by a step before a card reaches its edge. Hit boxes are the painted controls.
- **The graph camera is session state.** Magnification and offset come back after the composer, Back, or a breadcrumb for the rest of the session. No saved camera table.
- **Box-select is a marquee of subject cards.** Hold Shift, or arm Select, and drag. A Shift rectangle adds; Shift-click toggles one card. Select disarms when its rectangle is released.
- **The composer adds a bridge and a plain Observation the same way.** The new row is where you are looking, with keyboard focus in its first field. A bridge between two subjects on the Source starts in the composer. A saved row duplicates into an unsaved form. Choosing a Citation does not start Auto Transcribe.
- **Entry defaults are silent.** The date modal opens on Year, filled from the first date in the transcription. A full name line fills an empty part list on blur; Split asks before replacing hand edits. Name part types stay a fixed-width select.
- **Shared controls, not one-off fixes.** Clicking outside a sheet closes it and asks first when it is dirty. Enter saves a single-line edit when Save is enabled and leaves it otherwise. A long select label truncates, with the full string as help. Same-letter keypresses cycle in every select. A combo box ranks a label prefix, then a later word, then any other match.
- **Opening an existing project selects the oldest user**, the earliest `users.id`.
- **Last updated is an audit scope, not a column.** Handles get a `canonical_entity` scope that moves with their claims, pins, and member evidence, not with a property or term edit. Lists show it in one fixed column.
- **List scroll is a field on the history entry**, written when you leave the list, ignored for identity.
- **The Sources list is one page.** Sort, type, evidence status, and text search persist per project and user. Status counts primaries in one grouped read. Search filters the loaded rows by title and ref; it is not the omnibar.
- **One mark per subject kind.** The Event mark is a calendar, drawn everywhere through `PVMark`.
- **Gender wins over sex at birth on the Person page.** `gender` is seeded `man`, `woman`, `non_binary`, and researchers add terms. Sex at birth borrows a tint and says “· at birth”.
- **Names match the way records are written.** Accent folding is matching only (0.95); display, `NormalizeForm`, and `sort_key` never fold, so José and Jose stay separate groupings. Abbreviation and nickname tables are catalog vocabulary (0.98). Spaced particles rejoin. A value counts by its support, and a negative Observation is a contradiction. Date windows compare in days. The Western name-format profile is a catalog row. Project name frequency shipped and Spike 11 removes it ([`comparison-layers.md`](../../spike-11/comparison-layers.md#frequency)).
- **A depiction is an owned value row: a region plus an image File.** The app renders the region; Go checks, decodes, and re-encodes its own PNG and never stores the app's bytes. The image is not an Artifact and not a thumbnail. A File is released when no declared pointer remains, and one purge removes its object.
- **A gallery is the reconciled likeness values; a preferred portrait is a display pointer.** With no choice, the earliest likeness leads the Person page and the People list.
- **Unpromote deletes a filing.** It deletes the Subject's claim and the bridge claims on that Source that use it. It does not reject, and the handle stays.
- **New seeds are for new catalogs.** `engagement`, `interests`, `gender`, and `likeness` come from Install. No migration adds them to an existing catalog.

## Refused

An adaptive choice field. A default role on a new connection. Omnibar whole-word rank (already right). Simplified cards, culling cards to the window, and a canvas cover. A crop cache. Daitch–Mokotoff. Related event types and place chains in matching. A collation locale. Census-scale layers. Rejecting a claim. A handle delete.

## Leftovers

| Item | Home |
| --- | --- |
| Promote compares two likenesses as pictures | Spike 11, S11-20 |
| Other depiction keys, signature, transparent polygon, Windows PDF draw | [`depictions-and-likenesses.md`](../../../ideas/depictions-and-likenesses.md) |
| Other name patterns, transliteration, calendars, collation locale | [`international-names.md`](../../../ideas/international-names.md) |
| Daitch–Mokotoff | [`daitch-mokotoff.md`](../../../ideas/daitch-mokotoff.md) |
| Matcher leftovers | [`name-matching-enhancements.md`](../../../ideas/archive/name-matching-enhancements.md) |
| Place a common cluster in one action | [`evidence-graph-structure-shortcuts.md`](../../../ideas/evidence-graph-structure-shortcuts.md) |
| Reject a claim, remove a member, weak-claim review | [`identity-claim-review.md`](../../../ideas/identity-claim-review.md) |
| Census-scale layers, guided extract, a page-change guard for depictions, keyboard multi-select | [`docs/dogfood/ux.md`](../../../dogfood/ux.md) |
```

### Retarget edits

Each edit replaces the whole OLD text with the NEW text. Match on the OLD text, not a line number. The text is written in fenced blocks so it can be copied exactly.

**R-1** — `docs/matching.md`. OLD is the start of the line; the rest of the line (from ` [`ideas/archive/…`) stays.

```text
OLD: - **Abbreviations, event and place affinity:** Spike 10, [`deployment-plan/spike-10/name-matching-enhancements.md`](deployment-plan/spike-10/name-matching-enhancements.md). Known-as, concluded values, calibration, and rejected-claim memory:
NEW: - **Abbreviations:** shipped in [Spike 10](deployment-plan/archive/spike-10/). Event and place affinity, known-as, concluded values, calibration, and rejected-claim memory:
```

**R-2** — `docs/dogfood/ux.md`.

```text
OLD: Descoped from [Spike 10](../deployment-plan/spike-10/) on 2026-10-08. Canvas growth and box-select are in that spike; hiding part of a Source is not.
NEW: Descoped from [Spike 10](../deployment-plan/archive/spike-10/) on 2026-10-08. Canvas growth and box-select shipped in that spike; hiding part of a Source did not.
```

**R-3** — `docs/ideas/source-structure.md`.

```text
OLD: - Canvas growth landed in [Spike 10](../deployment-plan/spike-10/) (S10-03). Box-select is still S10-20.
NEW: - Canvas growth and box-select landed in [Spike 10](../deployment-plan/archive/spike-10/).
```

**R-4** — `docs/ideas/international-names.md`, the end of line 3.

```text
OLD: Design: [`spike-10/international-names.md`](../deployment-plan/spike-10/international-names.md). Abbreviations, project name frequency, and the rest of the matcher are in [`name-matching-enhancements.md`](archive/name-matching-enhancements.md) for the leftovers, and in that same spike for what it builds.
NEW: What shipped: [Spike 10](../deployment-plan/archive/spike-10/). Matcher leftovers: [`name-matching-enhancements.md`](archive/name-matching-enhancements.md).
```

**R-5** — `docs/ideas/depictions-and-likenesses.md`, the whole line 3.

```text
OLD: **Status:** idea, not scheduled. Spike 10 takes the value type, the `likeness` property, the bounding-box crop cache, the composer row, the Person gallery, a preferred portrait, and Promote's side-by-side crops. Design for that work: [`spike-10/depictions-and-likenesses.md`](../deployment-plan/spike-10/depictions-and-likenesses.md).
NEW: **Status:** idea, not scheduled. Spike 10 shipped the `depiction` value type, the `likeness` property, the composer row, the Person gallery, and a preferred portrait ([decisions](../deployment-plan/archive/spike-10/)). Promote comparing two likenesses is Spike 11 (S11-20).
```

**R-6** — `docs/ideas/depictions-and-likenesses.md`.

```text
OLD: - **Windows.** The cache row is the same. That client draws the PDF locator with its own stack at save time. Spike 10 draws it with PDFKit on the Mac.
NEW: - **Windows.** That client draws the PDF region with its own stack before it sends the image. Spike 10 draws it with PDFKit on the Mac.
```

**R-7** — `docs/ideas/archive/name-matching-enhancements.md`.

```text
OLD: Spike 10's matcher leftovers ([`spike-10/name-matching-enhancements.md`](../../deployment-plan/spike-10/name-matching-enhancements.md)).
NEW: Spike 10's matcher leftovers ([Spike 10](../../deployment-plan/archive/spike-10/)).
```

**R-8** — `docs/ideas/archive/glyph-consolidation.md`.

```text
OLD: Raised in the S10-19 plan review ([`S10-19-event-mark-sidebar.review.md`](../../deployment-plan/spike-10/prs/S10-19-event-mark-sidebar.review.md), M-3 and M-4).
NEW: Raised in the S10-19 plan review (M-3 and M-4).
```

**R-9** — `docs/deployment-plan/README.md`, under `## Current`. OLD is two paragraphs separated by one blank line; the second is indented below only to line up, and has no indent in the file. NEW is one paragraph.

```text
OLD: [Spike 10 — Dogfood UX](spike-10/) is open. Plan: [`spike-10/deployment-plan.md`](spike-10/deployment-plan.md). Briefs are authored just before the PR they gate.

     [Spike 11 — Promote matching refinement](spike-11/) is open alongside it. Plan: [`spike-11/deployment-plan.md`](spike-11/deployment-plan.md).
NEW: [Spike 11 — Promote matching refinement](spike-11/) is open. Plan: [`spike-11/deployment-plan.md`](spike-11/deployment-plan.md). Briefs are authored just before the PR they gate.
```

**R-10** — `docs/deployment-plan/README.md`, under `## Completed`. Insert this row directly above the Spike 9 row (the row that starts `| [Spike 9]`).

```text
| [Spike 10](archive/spike-10/) | Dogfood fixes. Names match across accents and abbreviations. A face on a photo is a depiction value. |
```

## 7. Steps

**STEP-1 — Write the archive README**
Files: `docs/deployment-plan/archive/spike-10/README.md` *(new)*. Skill: `archive-docs`.
Do: create the file with the §6 "Archive README" text, without the surrounding code fence.
Check: `wc -l docs/deployment-plan/archive/spike-10/README.md` prints at most 80; `grep -c '^## ' docs/deployment-plan/archive/spike-10/README.md` prints 3.

**STEP-2 — Retarget inbound links**
Files: `docs/matching.md`, `docs/dogfood/ux.md`, `docs/ideas/source-structure.md`, `docs/ideas/international-names.md`, `docs/ideas/depictions-and-likenesses.md`, `docs/ideas/archive/name-matching-enhancements.md`, `docs/ideas/archive/glyph-consolidation.md`.
Do: apply R-1 to R-8.
Check: `rg -n 'spike-10/' docs/matching.md docs/dogfood/ux.md docs/ideas` prints only lines containing `archive/spike-10/`.

**STEP-3 — Update the deployment-plan index**
Files: `docs/deployment-plan/README.md`.
Do: apply R-9 and R-10.
Check: `rg -n 'spike-10' docs/deployment-plan/README.md` prints only the new Completed row.

**STEP-4 — Delete the spike folder**
Files: `docs/deployment-plan/spike-10/`.
Do: `git rm -r` every file in `docs/deployment-plan/spike-10/` except `prs/S10-99-dogfood-close.md`, `prs/S10-99-dogfood-close.review.md`, and `prs/S10-99-dogfood-close.ship.md` (D-5).
Check: `git ls-files docs/deployment-plan/spike-10` prints only `S10-99-dogfood-close*` files under `prs/`.

**STEP-5 — Prove the links**
Files: none changed.
Do: run T-1 and T-2.
Check: both print what §8 says; `scripts/check-pr.sh` passes per D-10.

**STEP-6 — Remove this plan with the ship log**
Files: `docs/deployment-plan/spike-10/prs/S10-99-dogfood-close.md`, `.review.md`, `.ship.md`.
Do: after the ship log is final, `git rm -r docs/deployment-plan/spike-10` in the commit that records it. This is the branch's last commit. Without `ship-pr`, the implementer makes this commit before marking the PR ready.
Check: `ls docs/deployment-plan` prints `README.md`, `archive`, `spike-11`.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Docs (shell) | `rg -n 'deployment-plan/spike-10\|\.\./spike-10/\|\(spike-10/' docs .agents .cursor --glob '!docs/deployment-plan/spike-10/**'` | Prints exactly two lines: `docs/deployment-plan/spike-11/prs/S11-02-value-compare-core.md:292` and `docs/deployment-plan/spike-11/prs/S11-03-value-compare-names.md:264` | AC-2.1, D-7 |
| T-2 | Docs (shell) | `python3 -c 'import re,sys,pathlib as P;b=[f"{f}: {m[0]}" for f in sys.argv[1:] for m in re.findall(r"\]\(([^)#\s]+)(#[^)]*)?\)",P.Path(f).read_text()) if not m[0].startswith(("http:","https:","mailto:")) and not (P.Path(f).parent/m[0]).exists()];print(*b,sep="\n");sys.exit(bool(b))' docs/deployment-plan/archive/spike-10/README.md docs/deployment-plan/README.md docs/matching.md docs/dogfood/ux.md docs/ideas/*.md docs/ideas/archive/*.md` | Exit 0, no output | AC-2.2, AC-3.1 |
| T-3 | Docs (shell) | `wc -l docs/deployment-plan/archive/spike-10/README.md`; `grep '^## ' docs/deployment-plan/archive/spike-10/README.md` | ≤ 80 lines; `## Decisions`, `## Refused`, `## Leftovers` | AC-1.2, D-3 |
| T-4 | Docs (shell) | `ls docs/deployment-plan` at the PR head | `README.md`, `archive`, `spike-11` | AC-1.1, D-5 |
| T-5 | Docs (shell) | `rg -n 'Spike 1[01]' docs/deployment-plan/README.md` | The Spike 11 Current line and the Spike 10 Completed row, Spike 10 above Spike 9 | AC-1.3, D-9 |
| T-6 | Docs (shell) | `rg -n 'crop cache\|side-by-side\|cache row' docs/ideas/depictions-and-likenesses.md` | No output | AC-3.2 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `plan-docs` if a Spike 11 plan or brief links a deleted file (T-1 catches it first).

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS, or on non-macOS fails only xcode-test (needs macOS) (D-10)
- [ ] No Out line was implemented; nothing outside §4 / §6 was edited
- [ ] No user-facing copy changed
- [ ] No VERSION bump
- [ ] docs/deployment-plan/spike-10/ is gone at the PR head, this plan with it (D-5); no Landed note or completed.md entry (D-6)
- [ ] docs/deployment-plan/README.md: Current names Spike 11 only; Completed has the Spike 10 row
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

1. On the PR's Files tab, open `docs/deployment-plan/archive/spike-10/README.md` (new) rendered. Click every link in it: each opens a page.
2. Open `docs/deployment-plan/README.md`: Current names only Spike 11; the Spike 10 row links the new README.
3. Open `docs/ideas/depictions-and-likenesses.md` and `docs/ideas/source-structure.md`: the Spike 10 links open the archive README.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| A PR merged after `bdce4cd` adds a link into `spike-10/` (for example a Spike 11 plan citing a Spike 10 plan) | T-1 runs at the PR head and lists it |
| A Spike 10 follow-up lands while this PR is open and edits `spike-10/` | Merge `main` in; the deletion wins; its decision, if new, is a README line only when it changes a decision above |

**Stop and report if:**

- T-1 prints a line other than the two Spike 11 plan lines, and that line is in a file outside §6 "Retarget edits".
- A skill under `.agents/skills/` or a file under `core/`, `api/`, or `macos/` links into `docs/deployment-plan/spike-10/`.
- `main` has a Spike 10 PR still open or a map row in `docs/deployment-plan/spike-10/deployment-plan.md` whose status is not landed or descoped (other than S10-99).
- A §6 "Retarget edits" old text is not in its file.

## 12. Open questions

None.

## 13. PR

**Title:** S10-99 — Close Spike 10
**Body:** objective in one sentence; the archive README and the deleted folder (US-1); the retargeted links (US-2); the leftovers table (US-3); T-1 to T-6 output; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-10/prs/S10-99-dogfood-close.md` (deleted in this PR; see the branch history).
