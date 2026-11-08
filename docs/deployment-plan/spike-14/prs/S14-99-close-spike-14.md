# S14-99 — Close Spike 14 as a decisions note

**Status:** Draft
**Spike / slice:** Spike 14 · Spike close
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no chrome change: docs and one lint script, no view changes)
**Board:** —
**Depends on:** every other Spike 14 row, all Landed or Descoped at the planning base (last: S14-42, landed [#748](https://github.com/mendahu/provenencia/pull/748))
**Blocks:** —
**Branch:** `s14-99-close-spike-14`
**Version bump:** none — docs and a lint script only; no product change ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`archive-docs`](../../../../.agents/skills/archive-docs/SKILL.md)
**Rules that bind:** —
**Context:** [`archive-docs`](../../../../.agents/skills/archive-docs/SKILL.md) "Spike close — checklist"; precedent commits `878b45ea` (Spike 11 close) and `35ec7443` (Spike 12 close); [`archive/spike-12/README.md`](../../archive/spike-12/README.md) (shape to match); working rule "Even spikes are odds and ends"
**Planned:** 2026-11-08 against `main` @ `2ac041a4`
**Review:** [`S14-99-close-spike-14.review.md`](S14-99-close-spike-14.review.md) — round 1 answered, round 2 pending

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

Spike 14 is finished: every PR in its map has landed or been descoped. This PR closes it the way Spikes 11 and 12 were closed. `docs/deployment-plan/spike-14/` becomes `docs/deployment-plan/archive/spike-14/` *(new)*, holding a short decisions note and the one design doc that code and skills still build on (`summary-inputs.md`), now linked from the docs that teach its mechanism. Every open follow-up from the ship logs moves to an ideas file or is dropped with a reason; ideas lines that Spike 14 finished are removed or put in the present tense. Every link into the deleted folder is retargeted. `archive-docs` learns the parts of this close that the next close will repeat, and `check-plan-docs.py` gains a link check for any Markdown file so the next close does not write its own.

**Why now.** S14-42 (#748) was the last open row. Spike 16, the next odds-and-ends spike, is being planned and takes one-off work from here on (user, 2026-11-08).

**What stays.** No product code, test, string or schema changes. Code comments that name a brief id (`S14-D1` in `macos/App/DesignSystem/Components/PageColumns/PVPageColumns.swift` and `macos/App/DesignSystem/README.md`, `S14-D9` in `macos/App/DesignSystem/Tokens/PVColor.swift`) stay: they name an id and link nothing. Mentions of `spike-14/` paths in code spans inside dated Spike 13 logs and landed plans stay for the same reason (§6 Retarget, last paragraph). The dogfood bar pass is skipped (user, 2026-11-08).

## 2. User stories

**US-1.** As a contributor, when I open `docs/deployment-plan/`, I find Spike 14 among the completed spikes as one decisions note, so I learn what it decided without reading 35 plans.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | `ls docs/deployment-plan/archive/spike-14/` *(new)* after STEP-10 | Exactly `README.md` and `summary-inputs.md` |
| AC-1.2 | `ls docs/deployment-plan/` after STEP-10 | No `spike-14` entry |
| AC-1.3 | `docs/deployment-plan/README.md` § Current and § Completed | Spike 14 is the first Completed row, linking `archive/spike-14/`; Current holds the Spike 13, Spike 15 and Spike 16 paragraphs of §6 Index text and no Spike 14 |
| AC-1.4 | The note (`archive/spike-14/README.md`, new) | 40 – 80 lines; headings Decisions, Refused, Leftovers |

**US-2.** As a contributor, when I follow any link in the repo, it does not land on a deleted Spike 14 file, and from the docs that teach the input index I reach its design.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | T-1 after STEP-10 | No broken link the base did not already have |
| AC-2.2 | T-2 after STEP-10 | No Markdown link target contains `spike-14/` without `archive/spike-14` |
| AC-2.3 | `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-13/prs/S13-10-header-and-sections.md docs/deployment-plan/spike-13/prs/S13-14-gallery-section.md` | 0 errors |
| AC-2.4 | `rg -l "archive/spike-14/summary-inputs.md" core/database/effects/README.md docs/catalog-deletes.md .agents/skills/add-searchable-kind/SKILL.md docs/ideas/structural-debt.md` | All four files listed |

**US-3.** As a contributor, no follow-up a Spike 14 review raised disappears with the ship logs, and the ideas files list only open work.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | Each Re-home row of §6 Leftovers | Its text is in the named ideas file and section |
| AC-3.2 | Each phrase in §6 "Drop phrases" | Found in the note's Leftovers table |
| AC-3.3 | Each row of §6 "Ideas made stale" | Applied: no struck-through Spike 14 row left in `structural-debt.md`; no future tense about S14-16, S14-22 or S14-42 in `docs/ideas/` |

**US-4.** As the next planner closing a spike, I read how to do it in `archive-docs` and check links with one command.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-4.1 | `python3 scripts/check-plan-docs.py --links docs/ideas/README.md` | Prints no error line and exits 0 |
| AC-4.2 | `python3 scripts/check-plan-docs.py --links` on a file holding one link to a missing file | Exits 1 and prints `<path>:<line>: error: broken link: <target>` |
| AC-4.3 | `.agents/skills/archive-docs/SKILL.md` | Holds the four lines of §6 "archive-docs additions" |

## 3. Scope

| In | Covered by |
| --- | --- |
| `check-plan-docs.py --links` mode | STEP-1, AC-4.1, AC-4.2 |
| Decisions note `archive/spike-14/README.md` *(new)* | STEP-2, AC-1.1, AC-1.4 |
| `summary-inputs.md` moved, stripped of PR history, linked from four docs | STEP-3, STEP-5, AC-1.1, AC-2.4 |
| Open ship-log follow-ups re-homed in ideas files | STEP-4, AC-3.1 |
| Ideas lines Spike 14 finished, removed or put in the present tense; `docs/ideas/README.md` summaries | STEP-4, AC-3.3 |
| Inbound links retargeted | STEP-5, AC-2.1 – AC-2.3 |
| `docs/deployment-plan/README.md` Current and Completed | STEP-6, AC-1.3 |
| `archive-docs` additions | STEP-7, AC-4.3 |
| Delete the rest of `spike-14/`: map, map review log, `completed.md`, `README.md`, both audits, `event-place-pictures.md`, `oversized-jpeg-thumbnails.md`, `design/`, `ux/`, all 35 other plans with their review and ship logs | STEP-8, STEP-10, AC-1.2 |

| Out | Home |
| --- | --- |
| The dogfood bar pass | Not planned (user, 2026-11-08: skipped) |
| Triage of the audits' unscheduled items (audit 1 L-1 – L-17 and nits, audit 2 B-25 – B-39) | Not planned (user, 2026-11-08: deleted without triage; git history keeps them) |
| Links that were broken before this PR and point outside `spike-14/` (S13-23 and S13-25 to `matching-test-suite.md`, and the rest of the base's broken links) | Spike 13's close and later sweeps; T-1 compares against the base so they do not fail this PR |
| Opening Spike 16's folder or map | Spike 16's `plan-spike` (being planned, user 2026-11-08) |
| Fixing any re-homed follow-up | The ideas file that holds it |
| Running `--links` inside `check-pr.sh` | Not planned: the base already holds about 110 broken links, so a repo-wide gate would fail every PR |
| A link to `summary-inputs.md` from `.cursor/rules/catalog-queries.mdc` | Not planned: the rule names `graphcache.InputScope`, and the skill it pairs with (`add-searchable-kind`) gets the link |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Map | `docs/deployment-plan/spike-14/deployment-plan.md` | Every PR row Landed or Descoped except S14-99 (`Scoped`). Working rules from line 108 are the source of the note's Decisions |
| Spike folder | `docs/deployment-plan/spike-14/` | `README.md`, `deployment-plan.md`, `deployment-plan.review.md`, `completed.md`, `app-health-audit.md`, `app-health-audit-2.md`, `event-place-pictures.md`, `oversized-jpeg-thumbnails.md`, `summary-inputs.md`, `design/` (`README.md`, and `archive/` with briefs S14-D1 – S14-D9 and `S14-D1-wide-layout.review.md`), `ux/` (`README.md` and three shipped notes), `prs/` (35 plans besides this one, each with a review and a ship log) |
| Live design doc | `docs/deployment-plan/spike-14/summary-inputs.md` | The only full description of the input index. Line 3 links the map and tells the PR story; lines 44 and 123 link `../../ideas/structural-debt.md`; lines 11, 32, 75, 78, 106, 114, 123 and 125 cite S14 plan, test and decision ids. Linked only from files inside `spike-14/` |
| Docs that teach the index | `core/database/effects/README.md:15` ("from the input index"), `docs/catalog-deletes.md:46` ("from the input index"), `.agents/skills/add-searchable-kind/SKILL.md:31` (`graphcache.InputScope`), `docs/ideas/structural-debt.md:35` (source cell "Spike 14 Batch 8 design") | Name the mechanism; link nothing |
| Graduated docs | `event-place-pictures.md`, `oversized-jpeg-thumbnails.md` | Their contract is in `docs/conclusion-layer-data-model.md`, `docs/seeded-vocabulary.md` and `core/raster/jpegdc.go`. Their "Stays later" / "Not this design" items have no other home except signature and the Timeline picture |
| Audits | `app-health-audit.md`, `app-health-audit-2.md` | Linked for provenance from `docs/ideas/structural-debt.md:3`, `promote-ranking.md:5`, `share-packages.md:49` and `docs/deployment-plan/archive/spike-12/source-structure.md:68`; named in code spans in the S13-12 ship log and S13-25 review log. The ideas rows describe themselves |
| Inbound links | 15 files, listed in §6 Retarget | Two of them (S13-D1:14, S13-D4:19) are already broken at the base (one `../` too few); the rest break when `spike-14/` goes (measured by deleting it on a scratch worktree) |
| Stale ideas | `docs/ideas/structural-debt.md:18,25,35,36`, `docs/ideas/provisional-identity-claims.md:34,51`, `docs/ideas/depictions-and-likenesses.md:3`, `docs/ideas/README.md:21,50,57` | Rows struck through as done, or future tense about S14-16, S14-22, S14-42 (all landed) |
| Index | `docs/deployment-plan/README.md:50-58` | § Current: Spike 13 (open), Spike 14 (open, four links into `spike-14/`), Spike 15 (not yet planned). § Completed starts at Spike 12 |
| Lint | `scripts/check-plan-docs.py` `LINK_RE` (line 99), `prose_lines`, `check_links` (line 197), `main` (line 333) | `check_links` checks relative links of plans and briefs only; `main` skips any other file with a warning. No other link check exists in `scripts/`, `.agents/` or `.cursor/` |
| PR gate | `scripts/check-pr.sh` | Runs diff-check, gofmt, go vet, go test, localizable, guidance-names, plan-docs (changed plans and briefs only), xcode-test and swift-warnings. xcode-test fails off macOS by design |
| Skill | `.agents/skills/archive-docs/SKILL.md` (`.claude/skills` is a symlink to `.agents/skills`) | Delete list names no audits and no `ux/`; Retarget names no command; note target "~40–80 lines" |

**Precedent.** `878b45ea` "Close Spike 11 as a decisions note." and `35ec7443` "Close Spike 12: …". Copy their shape: one `archive/spike-N/README.md` with Decisions, Refused, Leftovers; one kept design doc beside it, whose line 3 says the domain docs and code win (`archive/spike-12/source-structure.md:3`); everything else deleted; the index's Completed row is one line. Do not copy Spike 12 moving its audits to another spike: here they are deleted (D-4).

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | The archive keeps `README.md` and `summary-inputs.md` only, and four docs link the kept doc (STEP-5). | `summary-inputs.md` is the one full description of the input index; without links from where the mechanism is taught it has no reader once the spike folder goes. The pictures and JPEG docs have graduated (§4). | Keeping the pictures or JPEG docs; graduating `summary-inputs.md` into a new `docs/*.md` (larger than a close; Spike 11 and 12 kept theirs in the archive the same way) |
| D-2 | The note is the text in §6 "Decisions note", copied as written. `YYYY-MM-DD` is the date the implementer writes the file. | The decisions were made in the map's working rules; the implementer transcribes. Spike 12's decisions already in its note are not repeated. | Letting the implementer distill from the map |
| D-3 | Follow-ups are re-homed, dropped and resolved exactly as §6 Leftovers says, with the text given there. | Triaged at planning against the code and ideas files (2026-11-08). | Re-triaging or rewording during the PR |
| D-4 | Both audits are deleted with no triage of their unscheduled items. Their provenance links become date-and-commit text. | User, 2026-11-08. | Moving the audits to the archive; re-homing each low item |
| D-5 | No bar pass. | User, 2026-11-08. | Walking bar outcomes 1 – 34 |
| D-6 | § Current gets the Spike 16 paragraph of §6 Index text, unlinked; if a `spike-16/` folder exists under `docs/deployment-plan/` on the branch's base, "Spike 16 — Odds and ends" is linked to it. | User, 2026-11-08: Spike 16 is being planned and arrives soon. | Opening a Spike 16 skeleton here |
| D-7 | Retargets follow §6 Retarget exactly. A link to the map or a plan goes to the note; a link to a brief becomes the plain brief id, or the note where no live doc is linked beside it. Code-span mentions in dated logs and landed plans are not edited. | Briefs are deleted at close; the boards live in Claude Design. Logs and landed plans are records, and a code span links nothing. | Linking git commits for deleted files; rewriting dated logs |
| D-8 | `docs/dogfood/README.md:15` is deleted, not retargeted. | The notes shipped (S14-28, S14-36, S14-37, S14-39) and the folder goes. | Pointing it at the archive |
| D-9 | `check-plan-docs.py` gets a `--links PATH...` mode that runs the existing `check_links` on any Markdown file. It is a manual tool, not a `check-pr.sh` step. | One link rule for plans and closes; the close needs it and the next close repeats it. The repo has about 110 broken links at the base, so a gate would fail every PR. | An inline script per close (rewritten each time); a repo-wide gate |
| D-10 | `archive-docs` gains the lines of §6 "archive-docs additions" and nothing else. | The parts of this close the next one repeats belong in the skill, not in a plan that deletes itself. | Leaving the pattern only in git history |
| D-11 | Commits: STEP-1 "Check the links of any Markdown file with check-plan-docs." STEP-2 to STEP-8 "Close Spike 14 as a decisions note." A STEP-9 fix is its own commit. STEP-10, after `ship-pr` records the ship log, "Delete the Spike 14 close plan and its logs." | `archive-docs`: the close PR deletes its own plan, review log and ship log in its last commit, the one that records the ship log. | One commit (the ship log does not exist yet) |
| D-12 | PR-gate evidence: on macOS, `scripts/check-pr.sh` prints `check-pr: PASS`. Elsewhere, every step except xcode-test and swift-warnings passes, and CI's macOS job is green on the PR. | No Swift changes; xcode-test fails off macOS by design. | Requiring a Mac for a docs PR |

**Rules this plan bends.** None.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `docs/deployment-plan/archive/spike-14/README.md` *(new)* | file | `docs/deployment-plan/archive/spike-12/README.md` | Same shape, another spike |
| `--links` *(new)* flag in `scripts/check-plan-docs.py` | func | `check_links` in the same file | Extends it: the flag routes any path to the existing function |

**Builds on.** [`archive-docs`](../../../../.agents/skills/archive-docs/SKILL.md) spike-close checklist; the Spike 11 and 12 close commits; `check_links`.
**Deletes.** Everything in `docs/deployment-plan/spike-14/` except `summary-inputs.md` (moved). The Spike 14 paragraph in `docs/deployment-plan/README.md`. `docs/dogfood/README.md:15`. Rows 18 and 25 of `docs/ideas/structural-debt.md`.
**Next adds.** Spike 13's and Spike 16's closes follow `archive-docs` and run `--links`.

### Reach

None — docs and a manual lint mode; no new path into product code.

### Scale

None — no write, list or loop changes. `--links` over every Markdown file in the repo runs in under a few seconds.

### Edges

None — no product rule changes. `--links` reuses `check_links` unchanged, so it skips the same targets (http, mailto, placeholders) and the same fenced code as plan checking.

### Schema

None — no catalog change.

### Go core

None — no Go change.

### FFI

None — no wire change.

### macOS

None — no Swift change.

### Strings

None — no copy changes.

### Lint script

`scripts/check-plan-docs.py`:

- `main` accepts `--links` as its first argument. With it, each following path is read and passed to `check_links` with a `Report` for that path; no plan or brief check runs; output and exit status are as today (`<path>:<line>: error: broken link: <target>`, the summary line, exit 1 on any error).
- The module docstring's Usage gains: `python3 scripts/check-plan-docs.py --links FILE ...   relative links of any Markdown file`.

T-1 compares the base and the branch with it:

```bash
git worktree add /tmp/base-2ac041a4 2ac041a4
python3 scripts/check-plan-docs.py --links $(cd /tmp/base-2ac041a4 && git ls-files '*.md' | sed 's|^|/tmp/base-2ac041a4/|') \
  | grep 'broken link' | sed -E 's|^/tmp/base-2ac041a4/||; s|:[0-9]+: |\t|' | sort -u > /tmp/links-base.txt
python3 scripts/check-plan-docs.py --links $(git ls-files '*.md') \
  | grep 'broken link' | sed -E "s|^$PWD/||; s|:[0-9]+: |\t|" | sort -u > /tmp/links-branch.txt
comm -13 /tmp/links-base.txt /tmp/links-branch.txt
git worktree remove /tmp/base-2ac041a4
```

Expected: `comm` prints nothing. Lines are `file<TAB>error: broken link: target`, without line numbers, so edits that shift lines do not show as new. A file the PR moves or deletes is not on the branch side, so its old broken links drop out.

### Decisions note

Write `docs/deployment-plan/archive/spike-14/README.md` *(new)* with this text, the Leftovers table below in place of its placeholder line.

```markdown
# Spike 14 — Odds and ends

**Done** (closed YYYY-MM-DD). The first odds-and-ends spike: eight batches of small features, audit fixes and dogfood UX, each landing on its own, ending with summary rows that record their inputs so a write finds what it moved by lookup. It also built Spike 12's one-graph Citation, seeded FakeStore and Go-owned open allowlist ([Spike 12](../spike-12/README.md)). Live models: [`conclusion-layer-data-model.md`](../../../conclusion-layer-data-model.md), [`seeded-vocabulary.md`](../../../seeded-vocabulary.md), [`interpretation-layer-data-model.md`](../../../interpretation-layer-data-model.md), [`catalog-deletes.md`](../../../catalog-deletes.md), [`macos-client-patterns.md`](../../../macos-client-patterns.md), [`structured-name-model.md`](../../../structured-name-model.md), [`matching-model.md`](../../../matching-model.md). Design reasoning: [`summary-inputs.md`](summary-inputs.md).

## Decisions

- **Even spikes are odds and ends.** Small features, refinements and fixes land in batches, each a slice with its own Check; large features go to odd spikes.
- **Events and Places have pictures.** One locked depiction Property per kind, `event_picture` and `place_picture`, both labelled Picture; `likeness` stays the Person key. New seeds reach new catalogs only.
- **The gallery layer is named gallery.** "Likeness" stays only for the Person key, its copy and the portrait; `canonical_entities.preferred_likeness_observation_id` keeps its name.
- **One gallery for every kind.** The gallery reads a list of gallery keys in `core/vocabkey`; one preferred pointer, one lead rule and one gallery slot on `ConclusionDetailBody` serve Person, Event and Place. An Event or Place chooses its thumbnail as a Person chooses its portrait.
- **A depiction is drawn as a rectangle.** A Citation's own region keeps all seven tools; an old polygon depiction loads and saves as it is.
- **Oversized JPEGs get a reduced decode of our own.** Over the 64-megapixel budget, a pure-Go decoder reads only DC coefficients at 1/8 scale, bounded and fuzzed, so a Windows build reuses it. Other formats over the budget stay refused.
- **A hidden graph or Source page goes stale; it does not reload.**
- **File work happens before the catalog session.** One staged-object type in `ingest` streams, hashes, sniffs and decodes outside `catalogsession.Do`.
- **Kit components announce themselves.** A danger confirm's default is Cancel; a dirty form asks before a link inside it navigates away.
- **Every audited write returns its effects**, held by a test in `api/ffi`. Sidebar counts are an effects list, and every list reaches its keys through its descriptor.
- **Proposals are liberal; the researcher decides.** Other-link support is 0.75, so a thin name plus one lined-up link clears Weak.
- **A nickname folds into its set's formal name and reappears as a nick part:** *James “Jim” Robins*. Former surnames show in parentheses, italic on the titles.
- **One width breakpoint: 1600pt of pane.** Pages adopt Page columns in their own redesigns.
- **Boolean is a first-class value type.** Single-only, a Yes / No segmented control with nothing chosen at first, no Negative checkbox.
- **Smaller calls.** Conflict brackets honour cardinality; Tab reaches a following text area through `PVTabPair`; place relationship rows read the inverse label at the far end; creating a bridge between two handles files it; the Persons sort ignores honorific prefixes.
- **The structural fix comes before the patch.** A bug that needs a patch at each call site waits for the mechanism to be replaced.
- **Summary rows record their inputs; invalidation is a lookup.** An input is a handle's values, one bridge's links at one end, or an association's values; a row is its own input; isolated composition is the oracle. One input index, owned by `searchindex`, with no foreign key; every write reports what it moved through `summary_handles`. The hand-written inverse (`Dependents`, `AssociationEndpoints`, `linkEnds`, `SetDisplayDeps`) is gone.

## Refused

A transparent-polygon depiction crop. A per-kind thumbnail column or RPC. A Property named `depiction`. A separate Page columns PR (built with Spike 13's Person page). Promote picture wording (the evidence sheet went in Spike 13).

## Leftovers

LEFTOVERS TABLE
```

The Leftovers table, columns `| Item | Home |`, these rows in this order:

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
| Dropped: a thumbnail choice reloading three lists; a term create reloading Properties (2 ms); a broken oversized JPEG re-decoded per response (off the session since S14-12); copy nits ("A smaller copy of the file will open", a generic "Discard changes?", a race with another window's unlink); a test for the localizable script; a split of `check-guidance-names.py`; code nits in `evidencegraphs`, `FakeStore` and `L10n.Promote` | Not worth a PR; no measured cost (2026-11-08) |

**Drop phrases** (T-8 greps each in the note): `thumbnail choice reloading three lists`, `term create reloading Properties`, `re-decoded per response`, `A smaller copy of the file will open`, `Discard changes?`, `another window's unlink`, `localizable script`, `check-guidance-names.py`, `code nits`, `Polygon vertex add`, `unscheduled items`.

### Moved design doc

`git mv docs/deployment-plan/spike-14/summary-inputs.md docs/deployment-plan/archive/spike-14/summary-inputs.md` *(new)*, then these edits and no others:

| Line | Today | Becomes |
| --- | --- | --- |
| 3 | "Design for Spike 14 Batch 8 (user, 2026-11-07). The PRs are in the map …half-fixed." | "The design Spike 14 built for summary rows that record their inputs (closed with the spike; link "decisions" → `README.md`). Kept as the reasoning behind the input index; the rules are in `docs/catalog-deletes.md`, `core/database/effects/README.md`, `add-searchable-kind` and the code, and where they differ, those win. It came out of a stale search document after Unpromote (2026-11-07), where three new pieces of invalidation code still left bridge delete half-fixed." |
| 11 | Column head "How a write found what to rebuild before S14-42" | "How a write found what to rebuild before the index" |
| 32 | "which S14-38 could not" (the end of the sentence that names S14-38) | "which a patch could not" |
| 75 | "which is the S14-38 failure again" | "which is the stale-document failure again" |
| 78 | "(S14-40's plan, D-7)" | removed, with its parentheses |
| 106 | "Final counts from S14-40 (T-9, the recorder-checked index on the same fixture, every row):" | "Final counts (the recorder-checked index on the same fixture, every row):" |
| 114 | "T-2 and T-3", "T-10", "(D-13)" | "the scenario-world tests", "the rebuild test", removed |
| 44, 123 | `../../ideas/structural-debt.md` | `../../../ideas/structural-debt.md` |
| 123 | "Moving them onto the scope API is a later step after S14-42" | "Moving them onto the scope API is a later step" |
| 125 | "are held by a `graphcachetest` merge test in S14-42" | "are held by a `graphcachetest` merge test" |

The "Measured at planning (2026-11-07)" heading and its numbers stay.

### Leftovers

Re-home (STEP-4). Add each text as written. "Bullet" means a `- **Bold lead.** rest` line at the end of the named section; "Section" means a new `## Heading` with the text as its paragraph, at the end of the file; "Row" means a table row.

| From | File · place | Form | Text |
| --- | --- | --- | --- |
| S14-03 UX-4, UX-5 | `docs/ideas/depictions-and-likenesses.md` § Still open | Bullet | **Depiction captions.** While the viewer loads, the composer's depiction caption falls through to the empty case ("use a region button") and needs a loading string; `depictionDrawingHint` says "tools above" though a depiction has one tool. Both wait for the composer board's next pass. |
| `oversized-jpeg-thumbnails.md` | same | Bullet | **EXIF orientation.** No JPEG honours EXIF orientation; a fix covers every JPEG, not only the oversized ones the reduced decoder reads. |
| same | same | Bullet | **A larger cover derivative.** Whether a 256 px thumbnail is enough for a Source cover cut from a very large scan. |
| `event-place-pictures.md` | same | Bullet | **Pictures as matching evidence.** `subjectcompare` gives an Event's or Place's pictures no weight. |
| S14-11 UX-1 | `docs/ideas/evidence-graph-followups.md` | Section `## No reloading sign on a revisited canvas` | `isRereadingGraph` is drawn only in the error callout (`EvidenceGraphView.swift`), so a revisited stale canvas shows its old cards with no sign that it is rereading. Needs a brief. |
| S14-16 UX-1 | `docs/ideas/shared-list-helpers.md` | Section `## A failed sidebar recount is silent` | `WorkspaceView` reads `lastRefreshError` only in its first `.task`, so a sidebar recount that fails after a write says nothing. Observe it with `.onChange` and show it where the counts are. |
| S14-17 UX-2 | `docs/ideas/properties-and-terms.md` | Section `## A failed bind after a Property create` | When the bind after a Property create fails, the dialog stays open and Retry calls `createProperty` again, which fails on the duplicate key (`PropertiesModel.swift`). Retry binds the Property already made. |
| S14-21 CON-2 | `docs/ideas/international-names.md` § Scripts, sound, and dates | Bullet | **Hyphenated folded spellings.** "Soby-Kirke" (hyphen, no ø) does not find "Søby Kirke register": the search document's folded phrases split on spaces only. |
| S14-26 | `docs/ideas/share-packages.md` § Harden before sharing table | Row | `\| — \| S14-26 \| A name part's length is bounded when names are assembled, not in `namevalue` `Insert` or `Validate`, so a crafted package can store an unbounded part. \|` |
| S14-17 MNT-7 | `docs/ideas/structural-debt.md`, the row whose source cell is "S13-37 plan" (`.deletedSource`) | Append to the row's last cell | " A field relabel or delete also reaches Source pages through a Swift mutation beside its effects (S14-17)." |
| S14-21 PRF-1 | same file, the row whose source cell is "S14-33 ship log" and starts "The first `part_of` filed from a city" | Append to the row's last cell | " A join that adds a new spelling to the 6,000-event city also reprojects every dependent: 2.1 s at S14-21, 1.33 s after the input index." |
| S14-22 two-term set, MNT-9 | same file, after the last row | Row | `\| — \| S14-22 ship log \| A two-term bridge filed mid-Done can be missing from an already-loaded filed-link set for its other term (not reproduced), and `bridgefile.go` repeats `graphalign`'s link key. One keyed set in `connectrules` would serve both. \|` |

Drop (no ideas change; in the note's Leftovers): S14-02 PRF-1, S14-10 UX-2 / MNT-5 / MNT-6, S14-13 UX-4, S14-14 CON-2 / UX-3, S14-15 MNT-2, S14-17 PRF-1, S14-20 MNT-3, S14-22 MNT-10, S14-23 PRF-1, S14-29's vertex items.

Already resolved (no action): S14-01 UX-1 and UX-2 (S14-02), S14-03 UX-6 (S14-14 D-16), S14-13's flaky tests (S14-03's CI round), S14-33 FU-1 (S14-42), S14-33 PRF-3 / PRF-6 / PRF-4 / FU-3 and S14-37's two items and S14-09's four lines (already in ideas files), the three `ux/` notes (shipped).

### Ideas made stale

| File:line | Today | Becomes |
| --- | --- | --- |
| `docs/ideas/structural-debt.md:18` | Struck-through `SourceSurface` row, "Done in S14-36" | Row deleted |
| `docs/ideas/structural-debt.md:25` | Struck-through `subjects.Delete` row, "→ Spike 14 S14-42" | Row deleted |
| `docs/ideas/structural-debt.md:35` | Source cell "Spike 14 Batch 8 design"; "once the input index is used for invalidation (S14-42) they can record" | Source cell: link "Spike 14 input index" → `../deployment-plan/archive/spike-14/summary-inputs.md`; "now that the input index drives invalidation, they can record" |
| `docs/ideas/structural-debt.md:36` | "A term key could become a fourth input kind once S14-42 lands." | "A term key could become a fourth input kind." |
| `docs/ideas/provisional-identity-claims.md:34` | "S14-16 rewrites … Whichever lands second carries the `status` field through the rewrite." | "S14-16 (landed) rewrote `ApplyPromoteGraphAlignment`'s handler and Swift mapping; this work carries the `status` field through the rewritten code." |
| `docs/ideas/provisional-identity-claims.md:51` | "If that slot is reclaimed first (a candidate for S14-22), the board's placement needs redrawing." | "S14-22 (landed) left the slot as it was; if it is reclaimed first, the board's placement needs redrawing." |
| `docs/ideas/depictions-and-likenesses.md:3` | "… are [Spike 14](link to `../deployment-plan/spike-14/`) batch 1." | "… shipped in [Spike 14](link to `../deployment-plan/archive/spike-14/`)." |
| `docs/ideas/README.md:21` | "card image memory, upward shadows, P3 color drift, keyboard multi-select" | append ", reloading sign" |
| `docs/ideas/README.md:50` | "… term management (direction / inverses)" | append ", failed bind after create" |
| `docs/ideas/README.md:57` | "… Source section counts" | append ", silent sidebar recount" |

Before editing a "Today" line, check it still reads as shown; if not, stop (§11).

### Retarget

"The note" means `docs/deployment-plan/archive/spike-14/README.md` *(new)*, linked by a path relative to the edited file. "Link L → X" means the Markdown link whose label is L and target X. Targets are written without brackets so this plan's own link check does not read them.

| File:line | Today | Becomes |
| --- | --- | --- |
| `docs/deployment-plan/README.md:54` | Spike 14 open paragraph | Removed; see STEP-6 |
| `docs/deployment-plan/archive/spike-12/source-structure.md:68` | Link "audit" → `../../spike-14/app-health-audit.md` | Plain "the 2026-10-27 app-health audit", no link |
| `docs/deployment-plan/spike-13/deployment-plan.md:61` | "brief" + link "S14-D1" → `../spike-14/design/archive/S14-D1-wide-layout.md` | Plain "brief S14-D1" (`macos-client-patterns.md` § 9 is linked beside it) |
| `docs/deployment-plan/spike-13/deployment-plan.md:355` | Link "S14-D1" → same brief | "S14-D1 (" + link "Spike 14 decisions" → `../archive/spike-14/README.md` + ")" |
| `docs/deployment-plan/spike-13/deployment-plan.md:474` | Link "S14-02" → `../spike-14/deployment-plan.md#s14-02--pictures-in-the-app` | "S14-02 (" + link "Spike 14 decisions" → `../archive/spike-14/README.md` + ")" |
| `docs/deployment-plan/spike-13/person-detail.md:102` | Link "Spike 14" → `../spike-14/deployment-plan.md` | Link "Spike 14" → `../archive/spike-14/README.md` |
| `docs/deployment-plan/spike-13/design/archive/S13-D1-person-page-frame.md:14` | A link to `../../spike-14/design/archive/S14-D1-wide-layout.md` (broken at the base) | Its label as plain text, no link |
| `docs/deployment-plan/spike-13/design/archive/S13-D4-timeline-section.md:19` | A link to `../../spike-14/design/archive/S14-D4-events-list-thumbnail.md` (broken at the base) | Its label as plain text, no link |
| `docs/deployment-plan/spike-13/design/archive/S13-D5-gallery-and-viewer.md:8` | "the S14-D2 Pictures section, " + link "archived brief" → `../../../spike-14/design/archive/S14-D2-event-page-pictures.md` | "the S14-D2 Pictures section" (link and comma removed) |
| `docs/deployment-plan/spike-13/design/archive/S13-D5-gallery-and-viewer.md:9` | Link with code label `../../spike-14/deployment-plan.md` → `../../../spike-14/deployment-plan.md#working-rules` | Link "Spike 14 decisions" → `../../../archive/spike-14/README.md` |
| `docs/deployment-plan/spike-13/prs/S13-10-header-and-sections.md:6` | Link "S14-D1" → `../../spike-14/design/archive/S14-D1-wide-layout.md` | Link "S14-D1" → `../../archive/spike-14/README.md` |
| `docs/deployment-plan/spike-13/prs/S13-14-gallery-section.md:15` | "the S14-02 plan (" + link → `../../spike-14/prs/S14-02-pictures-in-the-app.md` + ") D-1 – D-5" | "the S14-02 plan's D-1 – D-5 (now in " + link "Spike 14 decisions" → `../../archive/spike-14/README.md` + ")" |
| `docs/dogfood/README.md:15` | Batch 7 notes line | Line deleted (D-8) |
| `docs/ideas/structural-debt.md:3` | "now at " + link to `app-health-audit.md` | "(2026-10-27, at `731e423`; deleted when Spike 14 closed, in git history)" |
| `docs/ideas/promote-ranking.md:5` | "now at " + link to `app-health-audit-2.md` | "(2026-10-28, at `3cc609c`; deleted when Spike 14 closed, in git history)" |
| `docs/ideas/share-packages.md:49` | "now at " + link to `app-health-audit-2.md` | Same as `promote-ranking.md` |
| `core/database/effects/README.md:15` | "from the input index" | "from the input index (" + link "design" → `../../../docs/deployment-plan/archive/spike-14/summary-inputs.md` + ")" |
| `docs/catalog-deletes.md:46` | "come from the input index." | "come from the " + link "input index" → `deployment-plan/archive/spike-14/summary-inputs.md` + "." |
| `.agents/skills/add-searchable-kind/SKILL.md:31` | "…records its inputs through `graphcache.InputScope`." | "…records its inputs through `graphcache.InputScope` (" + link "design" → `../../../docs/deployment-plan/archive/spike-14/summary-inputs.md` + ")." |

`docs/ideas/depictions-and-likenesses.md:3` and `structural-debt.md:35` are in "Ideas made stale". Not edited (D-7): code-span paths in `spike-13/prs/S13-10-header-and-sections.md:169`, `S13-12-family-section.ship.md:16`, `S13-25-the-pin-table-goes.review.md:40`, `S13-31-header-data-becomes-summaries.md:136` (landed).

### Index text

`docs/deployment-plan/README.md` § Current keeps the Spike 13 and Spike 15 paragraphs as they are, deletes the Spike 14 paragraph, and adds after Spike 15:

> Spike 16 — Odds and ends is being planned. Until its map is open, a one-off PR waits for it.

§ Completed gets, as its first row (with "Spike 14" linked to `archive/spike-14/`):

`| Spike 14 | Odds and ends in batches: Event and Place pictures, effects on every write, a boolean value type. Summary rows record their inputs. |`

### archive-docs additions

In `.agents/skills/archive-docs/SKILL.md`:

1. "Spike close — checklist", the Delete line: append ", app-health audits (once their items are triaged or the user drops them), shipped dogfood notes (`ux/`)".
2. "Spike close — checklist", new last line: `- [ ] The S#-99 plan carries the note's text and the Leftovers triage, so the implementer transcribes them`.
3. "Retarget", new last paragraph: "Find the links a delete breaks with `python3 scripts/check-plan-docs.py --links` over every `*.md`, on the base and on the branch, without line numbers, and compare; the base already has broken links of its own. A link to a deleted brief becomes the plain brief id; a link to the map or a plan goes to the decisions note."
4. "Keep (live contract)", new last paragraph: "A kept file is linked from the domain doc, skill or README that teaches its mechanism, or nothing will lead a reader to it once the spike folder goes. Its first lines say the domain docs and code win where they differ, and it cites no plan, test or decision id from the deleted plans."

## 7. Steps

**STEP-1 — Link check for any Markdown file**
Files: `scripts/check-plan-docs.py`.
Do: §6 Lint script. Commit (D-11).
Check: AC-4.1, AC-4.2 (make the broken-link file in a temp folder outside the repo); `python3 scripts/check-plan-docs.py` with no argument prints the same summary line as before.

**STEP-2 — Write the decisions note**
Files: `docs/deployment-plan/archive/spike-14/README.md` *(new)*. Skill: `archive-docs`.
Do: §6 Decisions note.
Check: T-7; `python3 scripts/check-plan-docs.py --links` on the note *(new)* exits 0.

**STEP-3 — Move the input-index design doc**
Files: `docs/deployment-plan/spike-14/summary-inputs.md` → `docs/deployment-plan/archive/spike-14/summary-inputs.md` *(new)*.
Do: §6 Moved design doc.
Check: `rg -n 'S14-|\bT-[0-9]|\bD-[0-9]'` over the moved file *(new)* prints nothing; `--links` on it exits 0.

**STEP-4 — Ideas files**
Files: `docs/ideas/depictions-and-likenesses.md`, `evidence-graph-followups.md`, `shared-list-helpers.md`, `properties-and-terms.md`, `international-names.md`, `share-packages.md`, `structural-debt.md`, `provisional-identity-claims.md`, `README.md`.
Do: §6 Leftovers Re-home rows and §6 Ideas made stale.
Check: T-5; AC-3.3 (`rg -n '~~' docs/ideas/structural-debt.md` prints no Spike 14 row; `rg -n 'once S14|lands second|candidate for S14' docs/ideas` prints nothing).

**STEP-5 — Retarget inbound links**
Files: every file in §6 Retarget except `docs/deployment-plan/README.md`.
Do: each row, exactly.
Check: AC-2.3; AC-2.4.

**STEP-6 — The deployment-plan index**
Files: `docs/deployment-plan/README.md`.
Do: §6 Index text, with D-6.
Check: T-6.

**STEP-7 — Teach the close in archive-docs**
Files: `.agents/skills/archive-docs/SKILL.md`. Skill: `archive-docs`.
Do: §6 archive-docs additions.
Check: AC-4.3; `python3 scripts/check-guidance-names.py` passes.

**STEP-8 — Delete the spike folder**
Files: `docs/deployment-plan/spike-14/`.
Do: `git rm -r` everything left in it except `prs/S14-99-close-spike-14.md` and `prs/S14-99-close-spike-14.review.md`. Commit STEP-2 to STEP-8 (D-11) with a body in the shape of `878b45ea`'s: what the archive keeps, what is deleted, where follow-ups went, that inbound links and the docs that teach the index point at the archive.
Check: `ls docs/deployment-plan/spike-14/prs` lists only the S14-99 files.

**STEP-9 — Sweep**
Files: none, unless the sweep finds a miss.
Do: run T-1, T-2 and T-3 with `docs/deployment-plan/spike-14/` excluded (this plan and its logs still live there and link the deleted map by design).
Check: all pass. A miss is fixed in its own commit by the rules of D-7, if D-7 decides it; otherwise stop (§11).

**STEP-10 — Delete this plan (last commit)**
Files: `docs/deployment-plan/spike-14/prs/S14-99-close-spike-14.md`, `.review.md`, `.ship.md`.
Do: after `ship-pr` records the ship log, `git rm` all three in one commit (D-11). No Landed note, no `completed.md` paragraph, no map status: the PR body carries any deviation. Then run T-1, T-2 and T-4 again with nothing excluded.
Check: AC-1.1, AC-1.2, AC-2.1, AC-2.2.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Script | §6 Lint script comparison | `comm -13` prints nothing (at STEP-9 drop lines whose file starts `docs/deployment-plan/spike-14/`) | AC-2.1, D-7, D-9 |
| T-2 | Script | `rg -n '\]\([^)]*spike-14/' --glob '!docs/deployment-plan/archive/spike-14/**' .` (at STEP-9 also `--glob '!docs/deployment-plan/spike-14/**'`) | Every hit's target contains `archive/spike-14` | AC-2.2 |
| T-3 | Script | `scripts/check-pr.sh` | D-12 | AC-2.3, D-12 |
| T-4 | Script | `ls docs/deployment-plan/archive/spike-14/ docs/deployment-plan/` | AC-1.1 and AC-1.2 | AC-1.1, AC-1.2 |
| T-5 | Script | `rg -F` each Re-home text's bold lead or heading in its file | Found once | AC-3.1, D-3 |
| T-6 | Script | `sed -n '/^## Current/,/^## Completed/p' docs/deployment-plan/README.md` and the first row under `## Completed` | Current: Spike 13, Spike 15, Spike 16 lines, no "Spike 14"; first Completed row is Spike 14 | AC-1.3, D-6 |
| T-7 | Script | `wc -l` and `grep '^## '` on the note *(new)* | 40 – 80 lines; Decisions, Refused, Leftovers | AC-1.4, D-2 |
| T-8 | Script | `rg -F` each §6 Drop phrase in the note *(new)* | Each found | AC-3.2, D-3 |
| T-9 | Script | `check-plan-docs.py --links` on `docs/ideas/README.md`, and on a temp file outside the repo holding one link to a missing file | Exit 0 with no error line; exit 1 with `broken link` | AC-4.1, AC-4.2, D-9 |
| T-11 | Script | `rg -n '~~' docs/ideas/structural-debt.md`; `rg -n 'once S14|lands second|candidate for S14' docs/ideas` | No Spike 14 row struck through; no hit | AC-3.3 |
| T-10 | Script | `rg -n 'S14-99 plan carries|--links|live contract' .agents/skills/archive-docs/SKILL.md`; `rg -l archive/spike-14/summary-inputs.md` on the four files of AC-2.4 | Four skill additions present; four files listed | AC-4.3, AC-2.4, D-1, D-10 |

Must stay green: everything `scripts/check-pr.sh` runs (D-12). Most likely to break: plan-docs on S13-10 and S13-14, and guidance-names on the `archive-docs` edit.

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] Gate per D-12: check-pr.sh PASS on macOS, or every step but xcode-test and swift-warnings passing elsewhere plus CI's macOS job green
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] Sweep: T-1 and T-2 after STEP-10 with nothing excluded; scripts/check-guidance-names.py passes (check-pr runs it)
- [ ] Version bump: none
- [ ] Note matches §6 text; Leftovers table matches §6
- [ ] Every Re-home and Ideas made stale row applied; every Drop phrase in the note
- [ ] archive-docs has the four additions; check-plan-docs.py has --links
- [ ] Last commit deletes this plan, its review log and its ship log; no spike-14/ folder remains; no Landed note, no completed.md paragraph, no map status change (archive-docs)
- [ ] PR description: what shipped, AC evidence, deviations, the Drop list, and that the bar pass was skipped (user, 2026-11-08)
```

## 10. Manual check

1. Open `docs/deployment-plan/README.md` on the branch in GitHub's file view. Spike 14 is the first Completed row and its link opens the decisions note.
2. In the note, open each Live models link and the `summary-inputs.md` link. Each opens.
3. Open `docs/catalog-deletes.md` and follow the "input index" link. It opens `summary-inputs.md`, whose line 3 says the domain docs and code win.
4. Open `docs/deployment-plan/spike-13/prs/S13-14-gallery-section.md` and follow the Spike 14 decisions link. It opens the note, which states the gallery slot decision.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| A link into `spike-14/` added on `main` after the planning base | T-1 and T-2 run on the branch, not on this plan's list |
| A Spike 14 row reopened or added after planning | Stop condition below |
| Spike 16's map links a Spike 14 file | T-1 catches it; retarget it by D-7 |

**Stop and report if:**

- Any PR-map row other than S14-99 is not `Landed …` or `Descoped …` on the branch's base.
- A file outside `docs/` other than `core/database/effects/README.md` and `.agents/skills/add-searchable-kind/SKILL.md` links a Spike 14 path with a Markdown link.
- A "Today" cell of §6 Moved design doc, Ideas made stale or Retarget does not match its file at the base.
- T-1 shows a broken link that §6 Retarget does not cover and D-7 does not decide.
- A Re-home item is fixed in code at the branch's base: drop it, list it in the PR body, do not add it to the ideas file.

## 12. Open questions

None.

## 13. PR

**Title:** S14-99 — Close Spike 14 as a decisions note
**Body:** objective in one sentence; what shipped against each US; AC evidence (T-1 – T-10 output); deviations from the plan; the Drop list; "Bar pass skipped (user, 2026-11-08)"; `Plan: docs/deployment-plan/spike-14/prs/S14-99-close-spike-14.md (deleted in the last commit)`.
