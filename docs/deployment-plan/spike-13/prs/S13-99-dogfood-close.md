# S13-99 — Dogfood close and docs

**Status:** Draft
**Spike / slice:** Spike 13 · Close (no slice)
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no chrome change: the PR touches only `docs/`)
**Board:** —
**Depends on:** every other Spike 13 row, each Landed or Descoped (last: S13-28, landed [#744](https://github.com/mendahu/provenencia/pull/744))
**Blocks:** —
**Branch:** `s13-99-dogfood-close`
**Version bump:** none — docs only ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`archive-docs`](../../../../.agents/skills/archive-docs/SKILL.md)
**Rules that bind:** —
**Context:** [`archive-docs` § Spike close — checklist](../../../../.agents/skills/archive-docs/SKILL.md#spike-close--checklist), precedent [`archive/spike-12/README.md`](../../archive/spike-12/README.md) and [`archive/spike-11/README.md`](../../archive/spike-11/README.md); the user's close answers (2026-11-08) in **Decisions**
**Planned:** 2026-11-08 against `main` @ `2ac041a4`
**Review:** [`S13-99-dogfood-close.review.md`](S13-99-dogfood-close.review.md) — round 2 pending

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

Spike 13 is done: the user has checked the dogfood bar on their own people and accepts what shipped as it is (user, 2026-11-08). This PR closes the spike the way Spikes 11 and 12 closed. `docs/deployment-plan/spike-13/` becomes a short decisions note under `archive/spike-13/`, kept beside the one design doc that a domain doc still links. Everything operational is deleted: the map, the completed log, the briefs, the PR plans, and the review and ship logs. Every inbound link is retargeted, so no doc points into a deleted file.

**Why now.** Every row in the PR map is Landed or Descoped, and the bar has passed. Spike 14 is still open and three of its plans link into `spike-13/`, so the next branch that touches them would fail `check-plan-docs.py`.

**What stays.** All code. The domain docs (`conclusion-layer-data-model.md`, `matching-model.md`, `macos-client-patterns.md`, `design-system-layers.md`) except the one link in STEP-4. Spikes 14 and 15, apart from the link lines STEP-4 names.

## 2. User stories

**US-1.** As a future planner, when I open `docs/deployment-plan/archive/spike-13/README.md` (new), I read Spike 13's lasting decisions, what it refused, and where its leftovers went, without reading any PR history.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | Open `archive/spike-13/README.md` (new) | It has 30 – 80 lines and the headings `## Decisions`, `## Refused` and `## Leftovers`, with the content in §6 **Decisions note**. It lists no PR map and pastes no dogfood bar |
| AC-1.2 | Follow the link in `docs/conclusion-layer-data-model.md` §5.1 ("Spike 13's design") | It opens `archive/spike-13/evidence-graph-comparison.md` (new) at the heading "First version (2026-11-07)" |

**US-2.** As a contributor, no link in `docs/`, `.agents/` or `.cursor/` points into the deleted `spike-13/` folder.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | `rg -n '\]\([^)]*spike-13/' docs .agents .cursor` | Every hit's target, resolved from its file's folder, is one of the two files the PR creates in the archive folder (new) |
| AC-2.2 | `python3 scripts/check-plan-docs.py` with no arguments (every plan and open brief of every open spike) | No errors |
| AC-2.4 | Read `docs/conclusion-layer-data-model.md` line 332 | It reads `\| **Inferred** \| no accepted Identity Claim; at least one accepted Reconciliation Claim with Observation pins \|` |
| AC-2.5 | Read the last row of `docs/ideas/structural-debt.md` | It is the row in §6 **Debt row** |
| AC-2.3 | Read `docs/deployment-plan/README.md` | **Current** does not name Spike 13 (line 11's "From Spike 13, odd-numbered spikes" sentence stays); **Completed** has a Spike 13 row at the top that links `archive/spike-13/` |

**US-3.** As a contributor, nothing operational from Spike 13 is left in `docs/`.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | `ls docs/deployment-plan/spike-13` after the last commit | No such directory |
| AC-3.2 | `ls docs/deployment-plan/archive/spike-13` | Exactly `README.md` and `evidence-graph-comparison.md` |

## 3. Scope

| In | Covered by |
| --- | --- |
| Decisions note `archive/spike-13/README.md` (new), text in §6 | STEP-1, AC-1.1 |
| Move `evidence-graph-comparison.md` into the archive, with a new Status line and fixed links | STEP-2, AC-1.2 |
| Delete the rest of `spike-13/` except this PR's own plan, review log and ship log | STEP-3, AC-3.1 |
| Retarget the 13 inbound links in §4 **Inbound links** | STEP-4, AC-2.1, AC-2.2 |
| `docs/conclusion-layer-data-model.md` line 332: "Reconciliation" → "Reconciliation Claim" (user, 2026-11-08) | STEP-4, AC-2.4 |
| One row in `docs/ideas/structural-debt.md`: a link check over all of `docs/` (user, 2026-11-08) | STEP-4, AC-2.5 |
| `docs/deployment-plan/README.md`: the Current paragraphs and a Completed row | STEP-5, AC-2.3 |
| Last commit deletes this plan, its review log and its ship log | STEP-6, AC-3.1, AC-3.2 |

| Out | Home |
| --- | --- |
| Running the dogfood bar or any PR's unrun §10 manual check | Not planned: the user has checked the bar (D-1) |
| Ship-log and review follow-ups (26 substantive items, among them 7 flaky tests) and about 135 logged nits | Dropped (user, 2026-11-08, D-3); the decisions note's Leftovers table names them in one row |
| Event and Place pages adopting the sections, the provenance sheet or record links | Not planned and not a leftover: those pages are correct as built (user, 2026-11-08, D-4) |
| Building the all-docs link check that the new `structural-debt.md` row describes | The row itself (STEP-4) |
| Rewording what a landed Spike 14 plan says beyond its link targets, including the backticked `spike-13` paths at lines 332 and 368 of `S14-02-pictures-in-the-app.md` | Not planned: Spike 14's own close deletes those plans |
| Prose that names a deleted plan, ship log or review id without linking it (`docs/ideas/structural-debt.md` rows cite "S13-37 plan", "S13-16 PRF-2" and others) | Stays as written: an id with no link breaks nothing, as `archive-docs` § Retarget allows |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Map | `docs/deployment-plan/spike-13/deployment-plan.md` | Every row Landed or Descoped; S13-99 `Scoped`. Holds the working rules the decisions note distills (§ Working rules) |
| Feature docs | `docs/deployment-plan/spike-13/person-detail.md`, `provenance-sheet.md`, `evidence-graph-comparison.md` | Requirements, the kit boundary, the record-link design |
| Live contract check | `provenance-sheet.md` | Graduated: `docs/design-system-layers.md` line 148, `docs/macos-client-patterns.md` line 106 and `macos/App/DesignSystem/README.md` (ProvenanceSheet row) hold its boundary. Deleted, not kept |
| Live contract check | `person-detail.md` | Requirements R-1 … R-40; no skill or code links it. Its **Not this spike** items are in D-4 and the Leftovers table. Deleted |
| Live contract check | `evidence-graph-comparison.md` | Linked by `docs/conclusion-layer-data-model.md` line 213 (anchor `#first-version-2026-11-07`) and `docs/ideas/promote-ranking.md` line 3. Kept (D-2) |
| Operational | `completed.md`, `deployment-plan.review.md`, `README.md`, `design/` (5 review logs, `design/README.md`, 10 briefs in `design/archive/`), `prs/` (30 plans, 30 review logs, 30 ship logs) | Deleted per `archive-docs` |
| Index | `docs/deployment-plan/README.md` § Current, § Completed | Says Spike 13 is open; Spike 14's paragraph ends "while Spike 13 continues"; Completed starts at Spike 12 |
| Archive precedent | `docs/deployment-plan/archive/spike-12/README.md` (37 lines), `archive/spike-11/README.md` (42 lines) | Shape to copy: **Done** line with live models, Decisions, Refused, Leftovers table |
| Tests | `scripts/check-plan-docs.py` | Lints relative links in every plan of every open spike; `scripts/check-pr.sh` runs it only on the plans a branch changes |

### Inbound links

Found with `rg -n 'spike-13/' --glob '!docs/deployment-plan/spike-13/**' .` at the planning base. Code, skills and `.cursor/` have none.

| # | File:line |
| --- | --- |
| L-1 | `docs/conclusion-layer-data-model.md:213` |
| L-2 | `docs/ideas/promote-ranking.md:3` |
| L-3 | `docs/ideas/kinship-view.md:5` |
| L-4 | `docs/ideas/kinship-view.md:9` |
| L-5 | `docs/ideas/provisional-identity-claims.md:40` |
| L-6 | `docs/ideas/provenance.md:25` |
| L-7 | `docs/ideas/provenance.md:48` |
| L-8 | `docs/deployment-plan/archive/spike-11/README.md:32` |
| L-9 | `docs/deployment-plan/archive/spike-12/README.md:32` |
| L-10 | `docs/deployment-plan/spike-14/deployment-plan.md:69` |
| L-11 | `docs/deployment-plan/spike-14/prs/S14-02-pictures-in-the-app.md:15` |
| L-12 | `docs/deployment-plan/spike-14/prs/S14-19-merge-leaves-no-stale-link.md:14` |
| L-13 | `docs/deployment-plan/spike-14/prs/S14-22-promote-wire-says-what-go-decided.md:15` |

Each pair below is an exact string replacement on that line: the first line of a pair is the text today, the second is what it becomes. Nothing else on the line changes.

```text
L-1   (Spike 13's design](deployment-plan/spike-13/evidence-graph-comparison.md#first-version-2026-11-07)
      (Spike 13's design](deployment-plan/archive/spike-13/evidence-graph-comparison.md#first-version-2026-11-07)
L-2   ](../deployment-plan/spike-13/evidence-graph-comparison.md)
      ](../deployment-plan/archive/spike-13/evidence-graph-comparison.md)
L-3   ([person detail](../deployment-plan/spike-13/person-detail.md#2-family))
      ([decisions](../deployment-plan/archive/spike-13/README.md#decisions))
L-4   ([working rule](../deployment-plan/spike-13/deployment-plan.md#engine))
      ([decision](../deployment-plan/archive/spike-13/README.md#decisions))
L-5   from brief [S13-D11](../deployment-plan/spike-13/design/archive/S13-D11-promote-without-assessment.md) §4.1.
      from brief S13-D11 §4.1 (in git with #633).
L-6   Related: [S13-36 ship log](../deployment-plan/spike-13/prs/S13-36-provenance-sheet-in-the-kit.ship.md) (SEC-1, F-11).
      Related: the S13-36 ship log (in git with #702), SEC-1 and F-11.
L-7   Related: [S13-06 ship log](../deployment-plan/spike-13/prs/S13-06-derived-rows-carry-their-records.ship.md) (PRF-1, PRF-2), [S13-11 ship log](../deployment-plan/spike-13/prs/S13-11-provenance-sheet.ship.md).
      Related: the S13-06 ship log (in git with #640), PRF-1 and PRF-2; the S13-11 ship log (in git with #650).
L-8   [Spike 13 slice 0](../../spike-13/deployment-plan.md)
      [Spike 13](../spike-13/README.md#decisions)
L-9   [Spike 13 slice 0](../../spike-13/deployment-plan.md)
      [Spike 13](../spike-13/README.md#decisions)
L-10  Spike 13's [S13-23](../spike-13/deployment-plan.md#s13-23--promote-stops-asking-for-pins).
      Spike 13's S13-23 ([decisions](../archive/spike-13/README.md#decisions)).
L-11  Spike 13's [S13-10 plan](../../spike-13/prs/S13-10-header-and-sections.md) D-3
      Spike 13's S13-10 plan (in git with #634) D-3
L-12  ([`../../spike-13/prs/S13-03-siblings-from-a-shared-parent.ship.md`](../../spike-13/prs/S13-03-siblings-from-a-shared-parent.ship.md), "fails
      (in git with #666, "fails
L-13  ([`S13-23-promote-stops-asking-for-pins.md`](../../spike-13/prs/S13-23-promote-stops-asking-for-pins.md))
      (in git with #633)
```

Lines 332 and 368 of `S14-02-pictures-in-the-app.md` name `spike-13` in backticks inside a landed plan's steps. They are not links and stay (Out: Spike 14's close deletes them).

**Precedent.** `878b45ea` (Close Spike 11 as a decisions note) and `35ec7443` (Close Spike 12) did the same shape. Copy their split: one README of decisions, one kept reasoning doc whose Status line says what it is the reasoning behind, the rest deleted, inbound links retargeted, and the index row slimmed. Neither close had a plan file, so STEP-6 follows `archive-docs` alone.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | No bar pass is run in this PR. The decisions note says the bar was checked by the user on 2026-11-08 and that what shipped is accepted as built, bar line 6 included (Source rows with a line per Evidence graph record, not Subject cards, as the agreed Sources board drew them) | The user's answer (2026-11-08): "Everything's been checked. We can just close it."; "Everything as shipped is fine." | Re-running each PR's unrun §10 — because the user has done the check |
| D-2 | Keep `evidence-graph-comparison.md` in `archive/spike-13/` as the reasoning behind the record-link count; delete `person-detail.md` and `provenance-sheet.md` | A domain doc links its anchor (L-1), and Spike 11 kept `comparison-layers.md` the same way. `provenance-sheet.md`'s boundary has graduated to three docs (§4); `person-detail.md` is requirements nobody links | Keeping `person-detail.md` — because no skill or code reads it, and `archive-docs` keeps only live contracts |
| D-3 | Every unhomed ship-log or review follow-up and every logged nit is dropped, not re-homed, flaky suites included. One Leftovers row names them so git can find them once the logs are deleted | User, 2026-11-08: "You can drop all these", then "Name them, drop flakes" after the review (M-2). `archive-docs` asks that each follow-up be homed or dropped with a reason, and the Spike 11 and 12 Dropped rows name their items | A `structural-debt.md` row for the flaky suites — because the user dropped it |
| D-4 | The Event and Place pages are not a leftover: they are correct as built, and a later redesign owes nothing now. The decisions note has no row for them and no idea file changes. The record-link scoring questions after the first version are homed in `docs/conclusion-layer-data-model.md` §13 item 4 (a domain doc, which already says "conflicts, reasons and Source quality are open there"), and Promote ranking with that scoring in `docs/ideas/promote-ranking.md` | User, 2026-11-08, after the review (M-1): "The existing events and places detail pages are built as planned. They're just going to change later, further down the line, and I don't need to write that down anywhere. They work right now. They're correct. There's nothing to fix." `archive-docs` homes open leftovers only, so nothing bends | A row in `wide-layout.md` or a new idea file — because the user said there is nothing to record |
| D-5 | A link to a doc this PR deletes takes one of two forms. A link to the map or a feature doc (`deployment-plan.md`, `person-detail.md`) becomes a link to `README.md#decisions` in the archive (L-3, L-4, L-8 – L-10). A link to a plan, a brief or a ship log becomes plain text that names the PR number, with no link (L-5 – L-7, L-11 – L-13) | The decisions note is where a map or feature-doc reader now finds the decision. A plan or ship log has no archive copy, so a README link would land somewhere that does not hold what it cited. This departs from Spike 11, which linked such files to its README (`878b45ea`) | Linking every deleted file to the README — because it lands the reader on a page that does not hold the cited text |
| D-6 | Two commits. Commit 1, "Close Spike 13 as a decisions note.", does STEP-1 – STEP-5. Commit 2 records the ship log and deletes this plan, its review log and its ship log (STEP-6) | `archive-docs`: "The close PR (S#-99) deletes its own plan, review log, and ship log in its last commit, the one that records the ship log." | One commit — because `ship-pr` writes the ship log during the run |

**Rules this plan bends.** None.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `docs/deployment-plan/archive/spike-13/README.md` *(new)* | file | `docs/deployment-plan/archive/spike-12/README.md` | Same shape, one per closed spike |
| `docs/deployment-plan/archive/spike-13/evidence-graph-comparison.md` *(new)*, moved | file | `docs/deployment-plan/archive/spike-11/comparison-layers.md` | A `git mv` of the existing file |
| One `docs/ideas/structural-debt.md` row *(new)* | doc row | The S13-16 PRF-2 row in the same table | Same table, one more row |

**Builds on.** `archive-docs`; the Spike 11 and 12 closes.
**Deletes.** `docs/deployment-plan/spike-13/` whole: `README.md`, `completed.md`, `deployment-plan.md`, `deployment-plan.review.md`, `person-detail.md`, `provenance-sheet.md`, `design/` with `design/archive/`, `prs/` (and, in STEP-6, this plan, `S13-99-dogfood-close.review.md` and `S13-99-dogfood-close.ship.md`).
**Next adds.** Spike 14's close copies this shape. A later Event and Place redesign reads the Decisions note's lines on the sectioned page and the provenance sheet.

### Reach

None — no code changes, and no new path into existing code.

### Scale

None — no write, list or loop changes.

### Edges

None — no rule, comparison or algorithm changes.

### Decisions note

STEP-1 writes this text into `docs/deployment-plan/archive/spike-13/README.md` (new). The implementer may fix a fact that `main` has changed since the planning base, and reports each fix. They add no line.

```markdown
# Spike 13 — Person detail page

**Done** (closed 2026-11-08). The Person page is where a researcher understands one canonical person: a fixed header over Vitals, Family, Timeline, Gallery and Sources, a provenance sheet beside every Vitals value, the claims that make the person up, and how the person's records connect. Before it, Identity Claims lost their Observation pins and Promote lost its evidence sheet. The user checked the dogfood bar on their own people and accepted what shipped as built (2026-11-08). Live models: [`conclusion-layer-data-model.md`](../../../conclusion-layer-data-model.md), [`matching-model.md`](../../../matching-model.md), [`macos-client-patterns.md`](../../../macos-client-patterns.md), [`design-system-layers.md`](../../../design-system-layers.md), [`catalog-deletes.md`](../../../catalog-deletes.md). Design reasoning for record links: [`evidence-graph-comparison.md`](evidence-graph-comparison.md).

## Decisions

- **An Identity Claim is membership only:** subject, entity, status. No Observation pins, no confidence grade, no argument; the tables and columns are dropped and audit history keeps its old rows. Promote is a quick action that files accepted claims and their connections, and its score ranks suggestions unseen.
- **One sectioned Conclusion page, not a Person fork.** `ConclusionDetailPage` on Page columns serves Person, Event and Place; the Person page passes five sections. A section is a history place (`WorkspaceLocation.pageSection`), so Back, Forward and restore return to it.
- **Each section reads its own query** (claims, family, timeline, record links, beside the detail read), served from the graph cache and refreshed by one per-person invalidation spec.
- **Canonical only, filtered in Go.** Nothing unpromoted appears on the page.
- **The provenance sheet is a kit recipe** (a record row, a record table, the sheet), filled by a pure assembler from the wire records, and it reports a link for the caller to map to a place. It replaces the Why on the Person page only; Event and Place keep the Why accordion, drawing the recipe's parts, until they adopt sections. The sheet leaves room for manual Reconciliation Claims ([conclusion §13](../../../conclusion-layer-data-model.md) item 5), so they slot in without a redesign.
- **A derived row carries its owner's records:** birth and death date and place show every surviving event's or Place's records, grouped by owner.
- **Family is grouped in Go** and shaped as a family tree's first step: units over a list of ids, each link saying what it rests on. Siblings from a shared parent are page-only; matching and Promote are unchanged.
- **The Timeline is direct participation, any role,** on a lifeline drawn to scale: birth and death always have a marker (a placeholder when unrecorded), undated events sit before death, and a date known only by its end ("bef. 1880") is dated and ordered at that end everywhere.
- **Display is decoupled from content.** The person card (`PVTileCard`, `PVPersonCard`) and gender treatment are design-system pieces; name, date, place and title formatting live in `Platform/`; a built-in term reads with the app's name everywhere.
- **One Unpromote flow** for the Evidence graph and the Person page. Removing a handle's last member deletes the handle in the same revision through `deleteimpact`, and a write's effects name the ids it deleted.
- **How a person's records connect is computed, never stored.** The first version counts what two Evidence graphs share about each other person they both name, weak or strong at two facts (the match registry's `RecordLinks.StrongAt`), drawn as a circular link diagram beside the Sources rows. It is not Promote's alignment score.

## Refused

Provisional claims at Promote ([`provisional-identity-claims.md`](../../../ideas/provisional-identity-claims.md)); a claim review sheet (replaced by record links); Subject cards in the Sources section (Source rows with a line per record instead); a compact header on scroll; stored, researcher-chosen links between records and a research log; unpromoted Subjects anywhere on the page.

## Leftovers

| Item | Home |
| --- | --- |
| The Source date and standard citations | Spike 15: [`source-provenance-date.md`](../../spike-15/source-provenance-date.md), [`citation-formats.md`](../../spike-15/citation-formats.md) |
| Record links after the first version: conflicts, reasons, a richer weight, Source quality and independence, events and places as lines | [conclusion §13](../../../conclusion-layer-data-model.md) item 4; the questions are in [`evidence-graph-comparison.md`](evidence-graph-comparison.md#open) |
| Promote ranking with the record-link scoring | [`promote-ranking.md`](../../../ideas/promote-ranking.md) |
| Kinship beyond the nuclear family; sibling evidence for matching | [`kinship-view.md`](../../../ideas/kinship-view.md) |
| Remember researcher decisions | [`matching-test-suite.md`](../../../ideas/matching-test-suite.md) |
| Ship-log and review follow-ups not already in an idea file or `structural-debt.md`: flaky macOS suites (`SourceGraphFormModelTests`, `SourcesModelTests`, `SourcePageModelTests`, `PVWidthClassTests`, `PVSegmentedControlTests`, the graph-progress `waitUntil` tests); gallery and likeness viewer focus and VoiceOver details; Family's "Show N more" focus; the Timeline card id namespace; `edgeKey`'s missing separator; the gallery's unbounded tile count; `DateRowDisplay` letting a date with no year win; a test seam for kit Return and focus; a wider `observationstest.Cite`; `productKey(of:)`, the summary-state rule and hand-rolled dividers as copies; the via term derived in Swift; the "A stronger match is %@" copy; and every logged nit | Dropped (user, 2026-11-08) |
```

### Kept doc

`git mv docs/deployment-plan/spike-13/evidence-graph-comparison.md docs/deployment-plan/archive/spike-13/evidence-graph-comparison.md`, then:

- Line 3, the Status line, is replaced by:

  ```markdown
  **Status:** the design behind how a person's records connect, which Spike 13 built (2026-10-29 – 2026-11-08). Kept as the reasoning behind [`conclusion-layer-data-model.md`](../../../conclusion-layer-data-model.md) §5.1 and [`matching-model.md`](../../../matching-model.md); where they differ, the model and the code win. Spike 13's decisions: [README](README.md).
  ```

- Line 7: the link target `../../conclusion-layer-data-model.md` gains one `../`.
- The link target `../../matching-model.md` gains one `../`.
- Line 30: this exact text

  ```markdown
  The decisions are working rules in the [map](deployment-plan.md#slice-7--how-a-persons-records-connect): "Slice 7 starts with a simple count" and "The view is a circular link diagram". In short:
  ```

  becomes

  ```markdown
  The [decisions](README.md#decisions) say it in two lines, the record links and Refused. In short:
  ```

- Every other line stays, including **Open** and **Later** (D-4).

### Debt row

STEP-4 appends this row to the table in `docs/ideas/structural-debt.md`:

```markdown
| — | S13-99 review M-5 | `scripts/check-plan-docs.py` `check_links` resolves relative links only in plans and open briefs (`default_targets`), so nothing checks links or anchors in `docs/*.md`, `docs/ideas/` or `docs/deployment-plan/archive/`, and each spike close sweeps them by hand with `rg`. Let `check_links` run over every `docs/**/*.md`, anchors included, from `scripts/check-pr.sh`. |
```

### Schema

None — no catalog change.

### Go core

None — no Go change.

### FFI

None — no wire change.

### macOS

None — no app change.

### Strings

None — no copy change.

## 7. Steps

**STEP-1 — Decisions note**
Files: `docs/deployment-plan/archive/spike-13/README.md` (new). Skill: `archive-docs`.
Do: write the text in §6 **Decisions note**. Before writing, check each code name it uses on `main`: `rg -n 'pageSection|StrongAt|PVTileCard|PVPersonCard|PVPageColumns|ConclusionDetailPage' core macos` and `ls core/database/deleteimpact`, and that `identity_claim_evidence` appears in no `CREATE TABLE` after `core/database/migrations/000058.sql`. A name or fact that fails is fixed in the note and reported.
Check: `wc -l` gives 30 – 80; each relative link resolves (`ls` each target from the archive folder).

**STEP-2 — Move the kept doc**
Files: `docs/deployment-plan/spike-13/evidence-graph-comparison.md` → `docs/deployment-plan/archive/spike-13/evidence-graph-comparison.md` (new).
Do: `git mv`, then the edits in §6 **Kept doc**.
Check: `rg -o '\]\([^)#]+' docs/deployment-plan/archive/spike-13/evidence-graph-comparison.md` lists only targets that resolve from that folder.

**STEP-3 — Delete the operational docs**
Files: `docs/deployment-plan/spike-13/` (everything except `prs/S13-99-dogfood-close.md`, `prs/S13-99-dogfood-close.review.md` and `prs/S13-99-dogfood-close.ship.md`).
Do: `git rm` the files listed under **Deletes**.
Check: `git ls-files docs/deployment-plan/spike-13` lists only the three S13-99 files.

**STEP-4 — Retarget inbound links, and two small doc fixes**
Files: the 11 files in §4 **Inbound links**; `docs/ideas/structural-debt.md`.
Do: apply L-1 – L-13 as written; change nothing else on those lines. In `docs/conclusion-layer-data-model.md` line 332, replace `accepted Reconciliation with Observation pins` with `accepted Reconciliation Claim with Observation pins`. Append the §6 **Debt row** to the end of the `structural-debt.md` table.
Check: `rg -n '\]\([^)]*spike-13/' docs .agents .cursor` lists only links whose target, resolved from the file's own folder, is `docs/deployment-plan/archive/spike-13/README.md` (new) or `docs/deployment-plan/archive/spike-13/evidence-graph-comparison.md` (new), plus links inside the three S13-99 files. `python3 scripts/check-plan-docs.py` with no arguments reports no errors (warnings are allowed).

**STEP-5 — The index**
Files: `docs/deployment-plan/README.md`.
Do: delete the Spike 13 paragraph under **Current**. In the Spike 14 paragraph, delete the clause " while Spike 13 continues". Add this row at the top of **Completed**:

```markdown
| [Spike 13](archive/spike-13/) | The Person page is one sectioned Conclusion page: provenance sheet, family, timeline, gallery, sources. Identity Claims are membership only. A person's records connect by a computed count. |
```

Check: `rg -n 'Spike 13' docs/deployment-plan/README.md` hits two lines: line 11 ("From Spike 13, odd-numbered spikes", which stays) and the new Completed row. Commit 1 (D-6).

**STEP-6 — Delete this plan**
Files: `docs/deployment-plan/spike-13/prs/S13-99-dogfood-close.md`, `S13-99-dogfood-close.review.md`, `S13-99-dogfood-close.ship.md`.
Do: the review log exists from planning (the **Review:** line links it); `ship-pr` writes the ship log on this branch. In the commit that records the ship log, `git rm` all three; if the ship log is not there because the PR was built without `ship-pr`, `git rm` the other two. `docs/deployment-plan/spike-13/` is then gone. Any deviation goes in the PR body, not a Landed note.
Check: `ls docs/deployment-plan/spike-13` fails; `ls docs/deployment-plan/archive/spike-13` lists `README.md` and `evidence-graph-comparison.md`; `scripts/check-pr.sh` prints `check-pr: PASS` (off macOS it reports `SKIP` for the two Xcode steps, lines 120 – 125, which is a pass for this docs-only PR).

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Docs | `scripts/check-plan-docs.py` (no arguments) | Every plan and open brief of Spikes 14 and 15 after STEP-4 → no errors | AC-2.2, D-5 |
| T-2 | Docs | `rg -n '\]\([^)]*spike-13/' docs .agents .cursor` | After STEP-6 → every hit's target, resolved from its file's folder, is one of the two archive files | AC-2.1, AC-1.2 |
| T-3 | Docs | `wc -l` and `rg -n '^## ' docs/deployment-plan/archive/spike-13/README.md` | 30 – 80 lines; headings Decisions, Refused, Leftovers | AC-1.1, D-3, D-4 |
| T-4 | Docs | `ls docs/deployment-plan/spike-13 docs/deployment-plan/archive/spike-13` | First fails; second lists two files | AC-3.1, AC-3.2, D-2 |
| T-6 | Docs | `sed -n 332p docs/conclusion-layer-data-model.md`; `tail -1 docs/ideas/structural-debt.md` | Line 332 says "Reconciliation Claim with Observation pins"; the last row starts `\| — \| S13-99 review M-5 \|` | AC-2.4, AC-2.5 |
| T-5 | Docs | `rg -n 'Spike 13' docs/deployment-plan/README.md` | Two hits: line 11 (unchanged) and the Completed row | AC-2.3 |

Must stay green: everything `scripts/check-pr.sh` runs. The suite most likely to break is `check-plan-docs.py` on a Spike 14 plan whose link STEP-4 missed.

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] Sweep: T-2's command passes; `rg -n 'spike-13' .agents .cursor core api macos scripts` has no hit
- [ ] No VERSION bump (docs only)
- [ ] No Landed note, no completed.md paragraph, no map status for S13-99: the map and this plan are deleted (archive-docs)
- [ ] PR description: what shipped, AC evidence, deviations from the plan, manual check
```

## 10. Manual check

1. On the PR's branch on GitHub, open `docs/deployment-plan/archive/spike-13/README.md` (new). Click every link: each opens a page (AC-1.1).
2. Open `docs/conclusion-layer-data-model.md`, find "Spike 13's design" in §5.1 and click it: it lands on "First version (2026-11-07)" in the archived `evidence-graph-comparison.md` (AC-1.2).
3. Open `docs/deployment-plan/README.md`: Spike 13 appears only in Completed, and its link opens the archive folder (AC-2.3).

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| A link into `spike-13/` lands on `main` after the planning base | STEP-4's check runs `rg` on the branch, not the list in §4; each new hit gets the D-5 treatment |
| The decisions note states a fact `main` has since changed | STEP-1 checks each code name with `rg` first |

**Stop and report if:**

- A Spike 13 map row is not Landed or Descoped on `main`.
- `rg -n 'spike-13' .agents .cursor core api macos scripts` finds a hit: a skill or code reads a file this PR deletes.
- `check-plan-docs.py` with no arguments reports an error that is not a link into `spike-13/`. That error is not this PR's to fix.
- A link into `spike-13/` that is not L-1 – L-13 points at a requirement or rule (an `R-` id or a working rule) that the decisions note does not state.

## 12. Open questions

None.

## 13. PR

**Title:** S13-99 — Close Spike 13 as a decisions note
**Body:** objective in one sentence; the archive's two files; the user's close answers (D-1, D-3, D-4); AC evidence (T-1 – T-5 output); deviations from the plan; the manual check; `Plan: docs/deployment-plan/spike-13/prs/S13-99-dogfood-close.md` (deleted in the last commit, in git).
