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

Spike 13 is done: the user has checked the dogfood bar on their own people and accepts what shipped as it is (user, 2026-11-08). This PR closes the spike the way Spikes 11 and 12 closed. `docs/deployment-plan/spike-13/` becomes a short decisions note under `archive/spike-13/`, kept beside the one design doc that a domain doc still links. Everything operational is deleted: the map, the completed log, the briefs, the PR plans, and the review and ship logs. Every inbound link is retargeted, so no doc points into a deleted file.

**Why now.** Every row in the PR map is Landed or Descoped, and the bar has passed. Spike 14 is still open and three of its plans link into `spike-13/`, so the next branch that touches them would fail `check-plan-docs.py`.

**What stays.** All code. The domain docs (`conclusion-layer-data-model.md`, `matching-model.md`, `macos-client-patterns.md`, `design-system-layers.md`) except the one link in STEP-4. Spikes 14 and 15, apart from the link lines STEP-4 names.

## 2. User stories

**US-1.** As a future planner, when I open `docs/deployment-plan/archive/spike-13/README.md` (new), I read Spike 13's lasting decisions, what it refused, and where its leftovers went, without reading any PR history.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | Open `archive/spike-13/README.md` (new) | It has 40 – 80 lines and the headings `## Decisions`, `## Refused` and `## Leftovers`, with the content in §6 **Decisions note**. It lists no PR map and pastes no dogfood bar |
| AC-1.2 | Follow the link in `docs/conclusion-layer-data-model.md` §5.1 ("Spike 13's design") | It opens `archive/spike-13/evidence-graph-comparison.md` (new) at the heading "First version (2026-11-07)" |

**US-2.** As a contributor, no link in `docs/`, `.agents/` or `.cursor/` points into the deleted `spike-13/` folder.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | `rg -n '\]\([^)]*spike-13/' docs .agents .cursor` | Every hit's target is under `archive/spike-13/`, and that file exists |
| AC-2.2 | `python3 scripts/check-plan-docs.py` with no arguments (every plan and open brief of every open spike) | No errors |
| AC-2.3 | Read `docs/deployment-plan/README.md` | **Current** does not name Spike 13; **Completed** has a Spike 13 row at the top that links `archive/spike-13/` |

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
| `docs/deployment-plan/README.md`: the Current paragraphs and a Completed row | STEP-5, AC-2.3 |
| Last commit deletes this plan, its review log and its ship log | STEP-6, AC-3.1, AC-3.2 |

| Out | Home |
| --- | --- |
| Running the dogfood bar or any PR's unrun §10 manual check | Not planned: the user has checked the bar (D-1) |
| Ship-log and review follow-ups (26 substantive items, among them 7 flaky tests) and about 135 logged nits, including S13-23 UX-5's "A stronger match is %@" copy | Dropped (user, 2026-11-08, D-3); the decisions note's Leftovers table says so in one row |
| Event and Place pages adopting the sections, the provenance sheet and record links; deleting their Why accordion and its own record table; Event and Place pictures still opening the Citation; the Event `TermFacts` wire; record-link scoring after the first version | The user carries these into later spikes (user, 2026-11-08, D-4); one Leftovers row names them. No idea file is created |
| Fixing `docs/conclusion-layer-data-model.md` line 332 ("Observation pins" means Reconciliation pins, S13-25 CON-1) | Dropped with the nits (D-3) |
| Rewording what a landed Spike 14 plan says beyond its link targets | Not planned: Spike 14's own close deletes those plans |

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

| # | File:line | Link today | Becomes |
| --- | --- | --- | --- |
| L-1 | `docs/conclusion-layer-data-model.md:213` | `deployment-plan/spike-13/evidence-graph-comparison.md#first-version-2026-11-07` | `deployment-plan/archive/spike-13/evidence-graph-comparison.md#first-version-2026-11-07` |
| L-2 | `docs/ideas/promote-ranking.md:3` | `../deployment-plan/spike-13/evidence-graph-comparison.md` | `../deployment-plan/archive/spike-13/evidence-graph-comparison.md` |
| L-3 | `docs/ideas/kinship-view.md:5` | `../deployment-plan/spike-13/person-detail.md#2-family` | `../deployment-plan/archive/spike-13/README.md` |
| L-4 | `docs/ideas/kinship-view.md:9` | `../deployment-plan/spike-13/deployment-plan.md#engine` | `../deployment-plan/archive/spike-13/README.md` |
| L-5 | `docs/ideas/provisional-identity-claims.md:40` | link to brief `S13-D11-promote-without-assessment.md` | Plain text: `brief S13-D11 (in git with #633)`, no link |
| L-6 | `docs/ideas/provenance.md:25` | link to the S13-36 ship log | Plain text: `S13-36 review (#702) SEC-1, F-11`, no link |
| L-7 | `docs/ideas/provenance.md:48` | links to the S13-06 and S13-11 ship logs | Plain text: `S13-06 review (#640) PRF-1, PRF-2; S13-11 (#650)`, no links |
| L-8 | `docs/deployment-plan/archive/spike-11/README.md:32` | `../../spike-13/deployment-plan.md` | `../spike-13/README.md` |
| L-9 | `docs/deployment-plan/archive/spike-12/README.md:32` | `../../spike-13/deployment-plan.md` | `../spike-13/README.md` |
| L-10 | `docs/deployment-plan/spike-14/deployment-plan.md:69` | `../spike-13/deployment-plan.md#s13-23--promote-stops-asking-for-pins` | `../archive/spike-13/README.md` |
| L-11 | `docs/deployment-plan/spike-14/prs/S14-02-pictures-in-the-app.md:15` | link to the S13-10 plan | Plain text: `Spike 13's S13-10 plan (in git with #634) D-3, D-4, D-5, D-10, D-16` |
| L-12 | `docs/deployment-plan/spike-14/prs/S14-19-merge-leaves-no-stale-link.md:14` | link to the S13-03 ship log | Plain text: `the S13-03 ship log follow-up (in git with #666)`, keeping the quoted test name |
| L-13 | `docs/deployment-plan/spike-14/prs/S14-22-promote-wire-says-what-go-decided.md:15` | link to the S13-23 plan | Plain text: `S13-23 plan **Next adds** (in git with #633)` |

Lines 332 and 368 of `S14-02-pictures-in-the-app.md` name `spike-13` in backticks inside a landed plan's steps. They are not links and stay (Out: Spike 14's close deletes them).

**Precedent.** `878b45ea` (Close Spike 11 as a decisions note) and `35ec7443` (Close Spike 12) did the same shape. Copy their split: one README of decisions, one kept reasoning doc whose Status line says what it is the reasoning behind, the rest deleted, inbound links retargeted, and the index row slimmed. Neither close had a plan file, so STEP-6 follows `archive-docs` alone.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | No bar pass is run in this PR. The decisions note says the bar was checked by the user on 2026-11-08 and that what shipped is accepted as built, bar line 6 included (Source rows with a line per Evidence graph record, not Subject cards; board S13-D6 rev 2) | The user's answer (2026-11-08): "Everything's been checked. We can just close it."; "Everything as shipped is fine." | Re-running each PR's unrun §10 — because the user has done the check |
| D-2 | Keep `evidence-graph-comparison.md` in `archive/spike-13/` as the reasoning behind the record-link count; delete `person-detail.md` and `provenance-sheet.md` | A domain doc links its anchor (L-1), and Spike 11 kept `comparison-layers.md` the same way. `provenance-sheet.md`'s boundary has graduated to three docs (§4); `person-detail.md` is requirements nobody links | Keeping `person-detail.md` — because no skill or code reads it, and `archive-docs` keeps only live contracts |
| D-3 | Every unhomed ship-log or review follow-up and every logged nit is dropped, not re-homed; one Leftovers row records it | User, 2026-11-08: "You can drop all these." `archive-docs` asks that each follow-up be homed or dropped with a reason; the reason is the user's answer | Homing them in idea files — because the user declined |
| D-4 | The Event and Place redesign cluster and the record-link scoring questions get no new file; one Leftovers row names them as the user's for later spikes, and the kept `evidence-graph-comparison.md` keeps its **Open** and **Later** lists unchanged | User, 2026-11-08: "I'll handle those in later spikes." | A new `ideas/event-and-place-pages.md` — because the user declined |
| D-5 | A link to a file this PR deletes (a plan, a brief, a ship log) becomes plain text that names the PR number, never a link to git history | The archive policy keeps that history in git; a GitHub URL to a deleted path breaks on renames, and Spike 11 and 12 retargeted the same way | Links to blob URLs at a commit — because nothing else in `docs/` does it |
| D-6 | Two commits. Commit 1, "Close Spike 13 as a decisions note.", does STEP-1 – STEP-5. Commit 2 records the ship log and deletes this plan, its review log and its ship log (STEP-6) | `archive-docs`: "The close PR (S#-99) deletes its own plan, review log, and ship log in its last commit, the one that records the ship log." | One commit — because `ship-pr` writes the ship log during the run |

**Rules this plan bends.** None.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `docs/deployment-plan/archive/spike-13/README.md` *(new)* | file | `docs/deployment-plan/archive/spike-12/README.md` | Same shape, one per closed spike |
| `docs/deployment-plan/archive/spike-13/evidence-graph-comparison.md` *(new)*, moved | file | `docs/deployment-plan/archive/spike-11/comparison-layers.md` | A `git mv` of the existing file |

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
- **The provenance sheet is a kit recipe** (a record row, a record table, the sheet), filled by a pure assembler from the wire records, and it reports a link for the caller to map to a place. It replaces the Why on the Person page only; Event and Place keep the Why accordion, drawing the recipe's parts, until they adopt sections.
- **A derived row carries its owner's records:** birth and death date and place show every surviving event's or Place's records, grouped by owner.
- **Family is grouped in Go** and shaped as a family tree's first step: units over a list of ids, each link saying what it rests on. Siblings from a shared parent are page-only; matching and Promote are unchanged.
- **The Timeline is direct participation, any role,** on a lifeline drawn to scale: birth and death always have a marker (a placeholder when unrecorded), undated events sit before death, and a date known only by its end ("bef. 1880") is dated and ordered at that end everywhere.
- **Display is decoupled from content.** The person card (`PVTileCard`, `PVPersonCard`) and gender treatment are design-system pieces; name, date, place and title formatting live in `Platform/`; a built-in term reads with the app's name everywhere.
- **One Unpromote flow** for the Evidence graph and the Person page. Removing a handle's last member deletes the handle in the same revision through `deleteimpact`, and a write's effects name the ids it deleted.
- **How a person's records connect is computed, never stored.** The first version counts what two Evidence graphs share about each other person they both name, weak or strong at two facts (`recordlinks.StrongAt`), drawn as a circular link diagram beside the Sources rows. It is not Promote's alignment score.

## Refused

Provisional claims at Promote ([`provisional-identity-claims.md`](../../../ideas/provisional-identity-claims.md)); a claim review sheet (replaced by record links); Subject cards in the Sources section (Source rows with a line per record instead, board S13-D6 rev 2); a compact header on scroll; stored, researcher-chosen links between records and a research log; unpromoted Subjects anywhere on the page.

## Leftovers

| Item | Home |
| --- | --- |
| The Source date and standard citations | Spike 15: [`source-provenance-date.md`](../../spike-15/source-provenance-date.md), [`citation-formats.md`](../../spike-15/citation-formats.md) |
| Event and Place pages adopting the sections, the provenance sheet and record links; deleting their Why accordion and its own record table; their pictures still opening the Citation; the Event summary's term facts on the wire | Later spikes (user, 2026-11-08) |
| Record links after the first version: conflicts, reasons, a richer weight, Source quality and independence, events and places as lines, Promote ranking with this scoring | Later spikes (user, 2026-11-08); the questions are in [`evidence-graph-comparison.md`](evidence-graph-comparison.md#open) |
| Kinship beyond the nuclear family; sibling evidence for matching | [`kinship-view.md`](../../../ideas/kinship-view.md) |
| Remember researcher decisions | [`matching-test-suite.md`](../../../ideas/matching-test-suite.md) |
| Ship-log and review follow-ups not already in an idea file or `structural-debt.md`, and every logged nit | Dropped (user, 2026-11-08) |
```

### Kept doc

`git mv docs/deployment-plan/spike-13/evidence-graph-comparison.md docs/deployment-plan/archive/spike-13/evidence-graph-comparison.md`, then:

- Line 3, the Status line, is replaced by:

  ```markdown
  **Status:** the design behind how a person's records connect, which Spike 13 built (2026-10-29 – 2026-11-08). Kept as the reasoning behind [`conclusion-layer-data-model.md`](../../../conclusion-layer-data-model.md) §5.1 and [`matching-model.md`](../../../matching-model.md); where they differ, the model and the code win. Spike 13's decisions: [README](README.md).
  ```

- Line 7: the link target `../../conclusion-layer-data-model.md` gains one `../`.
- The link target `../../matching-model.md` gains one `../`.
- Line 30: the link to the map's slice 7 rules, with its two quoted rule names, is replaced by the text below; the rest of the sentence stays.

  ```markdown
  [decisions](README.md#decisions)
  ```

- Every other line stays, including **Open** and **Later** (D-4).

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
Do: write the text in §6 **Decisions note**. Before writing, check each code name it uses on `main` (`rg -n 'pageSection|StrongAt|PVTileCard|PVPersonCard' core macos`).
Check: `wc -l` gives 40 – 80; each relative link resolves (`ls` each target from the archive folder).

**STEP-2 — Move the kept doc**
Files: `docs/deployment-plan/spike-13/evidence-graph-comparison.md` → `docs/deployment-plan/archive/spike-13/evidence-graph-comparison.md` (new).
Do: `git mv`, then the edits in §6 **Kept doc**.
Check: `rg -o '\]\([^)#]+' docs/deployment-plan/archive/spike-13/evidence-graph-comparison.md` lists only targets that resolve from that folder.

**STEP-3 — Delete the operational docs**
Files: `docs/deployment-plan/spike-13/` (everything except `prs/S13-99-dogfood-close.md`, `prs/S13-99-dogfood-close.review.md` and `prs/S13-99-dogfood-close.ship.md`).
Do: `git rm` the files listed under **Deletes**.
Check: `git ls-files docs/deployment-plan/spike-13` lists only the three S13-99 files.

**STEP-4 — Retarget inbound links**
Files: the 11 files in §4 **Inbound links**.
Do: apply L-1 – L-13 as written; change nothing else on those lines.
Check: `rg -n '\]\([^)]*spike-13/' docs .agents .cursor` shows only targets under `archive/spike-13/` plus the three S13-99 files' own links. `python3 scripts/check-plan-docs.py` with no arguments reports no errors (warnings are allowed).

**STEP-5 — The index**
Files: `docs/deployment-plan/README.md`.
Do: delete the Spike 13 paragraph under **Current**. In the Spike 14 paragraph, delete the clause " while Spike 13 continues". Add this row at the top of **Completed**:

```markdown
| [Spike 13](archive/spike-13/) | The Person page is one sectioned Conclusion page: provenance sheet, family, timeline, gallery, sources. Identity Claims are membership only. A person's records connect by a computed count. |
```

Check: `rg -n 'Spike 13' docs/deployment-plan/README.md` hits only the Completed row. Commit 1 (D-6).

**STEP-6 — Delete this plan**
Files: `docs/deployment-plan/spike-13/prs/S13-99-dogfood-close.md`, `S13-99-dogfood-close.review.md`, `S13-99-dogfood-close.ship.md`.
Do: in the commit that records the ship log, `git rm` all three; `docs/deployment-plan/spike-13/` is then gone. Any deviation goes in the PR body, not a Landed note.
Check: `ls docs/deployment-plan/spike-13` fails; `ls docs/deployment-plan/archive/spike-13` lists `README.md` and `evidence-graph-comparison.md`; `scripts/check-pr.sh` prints `check-pr: PASS`.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Docs | `scripts/check-plan-docs.py` (no arguments) | Every plan and open brief of Spikes 14 and 15 after STEP-4 → no errors | AC-2.2, D-5 |
| T-2 | Docs | `rg -n '\]\([^)]*spike-13/' docs .agents .cursor` | After STEP-6 → every hit targets `archive/spike-13/README.md` or `archive/spike-13/evidence-graph-comparison.md` | AC-2.1, AC-1.2 |
| T-3 | Docs | `wc -l` and `rg -n '^## ' docs/deployment-plan/archive/spike-13/README.md` | 40 – 80 lines; headings Decisions, Refused, Leftovers | AC-1.1, D-3, D-4 |
| T-4 | Docs | `ls docs/deployment-plan/spike-13 docs/deployment-plan/archive/spike-13` | First fails; second lists two files | AC-3.1, AC-3.2, D-2 |
| T-5 | Docs | `rg -n 'Spike 13' docs/deployment-plan/README.md` | One hit, the Completed row | AC-2.3 |

Must stay green: everything `scripts/check-pr.sh` runs. The suite most likely to break is `check-plan-docs.py` on a Spike 14 plan whose link STEP-4 missed.

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] Sweep: `rg -n 'spike-13/' docs .agents .cursor` hits only archive/spike-13 links and the two landed S14-02 backtick lines named in §4
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
- A new link into `spike-13/` points at a requirement or rule (an `R-` id or a working rule) that the decisions note does not state.

## 12. Open questions

None.

## 13. PR

**Title:** S13-99 — Close Spike 13 as a decisions note
**Body:** objective in one sentence; the archive's two files; the user's close answers (D-1, D-3, D-4); AC evidence (T-1 – T-5 output); deviations from the plan; the manual check; `Plan: docs/deployment-plan/spike-13/prs/S13-99-dogfood-close.md` (deleted in the last commit, in git).
