# Ship log — S11-01 One-page matching model

**Plan:** [`S11-01-matching-model.md`](S11-01-matching-model.md) · **Branch:** `claude/modest-gates-tunm8o` (session-designated; plan header names `s11-01-matching-model`) · **PR:** #432
**State:** ready

Gate note: this ship runs in a Linux cloud session. `scripts/check-pr.sh` runs every check except `xcode-test` (macOS only); the PR is docs only, and CI runs the Xcode tests on the non-draft PR.

## Round 0 — implement
Head `5408b6b`. Gate PASS (all checks; `xcode-test` skipped, not macOS). Report: used the session branch; base moved `af50bd6` → `233e87a` with no Current state row invalidated; extra whitespace-fix commit `5408b6b`; landing docs (Landed #N, map row, `completed.md`) wait on the PR number and land in the finish commit.

## Round 1 — review
Base `233e87a`, head `5408b6b`. Verdicts: CON changes needed · MNT changes needed · UX clean · SEC clean · PRF clean.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| CON-1 | CON | major | Layer table gives layer 2 the `Near` minimums | Fix now → F-1 | verified: `comparison-layers.md` "Near has a minimum similarity per type. It lives in layer 1's Rules"; the parenthetical is implementer wording, not §6 |
| CON-2, MNT-1 | CON, MNT | major | `lifespan` is a sixth top-level registry section | To user → Q1 → Fix now F-2 | verified: map rule "one file per section (value, subject, segment, alignment, evidence)" and `comparison-layers.md` `subject.go` row "lifespan rules and constants"; keys come verbatim from the plan's §6 table |
| MNT-2 | MNT | major | Key grammar unstated; `sex` is not a property key | To user → Q1 → Fix now F-2 | verified: rows use 2–4 segments with mixed case; `subjectvocab` has `sex_at_birth` and `gender`, no `sex`; keys are §6 verbatim |
| UX-1 | UX | minor | Worked example drops the 0.1 lifespan agree though both sides have a birth | To user → Q2 → Fix now F-3 | verified: `comparison-layers.md` lifespan "subject of a birth at D: born at D … Overlaps or Contains is consistent"; sentence is the plan's §6 item 4 |
| UX-3, MNT-5 | UX, MNT | minor | Banners overstate what is replaced / don't say the page describes running code | To user → Q3 → Fix now F-4 | text prescribed verbatim by STEP-6; D-5 keeps §1–§3, §9 of `promote-graph-alignment.md` |
| UX-2 | UX | minor | Glossary says a researcher can tune knobs | Fix now → F-5 | README "Not this spike: researcher-facing tuning"; implementer wording, file already in the fix |
| MNT-4 | MNT | minor | "evidence" names points and the `evidence` registry section | Fix now → F-6 | glossary evidence = points; `evidence.*` rows are the claim-evidence sheet's line policy; few lines |
| MNT-3 | MNT | minor | Shortlisting terms defined twice, already drifting | Fix now → F-7 | `comparison-layers.md` table vs model glossary ("year" vs "year bucket"); file already touched; D-8 keeps decision history, only the definitions move |
| PRF-1 | PRF | minor | Event shortlist bucket grows with the catalog | Follow-up | design question for the S11-15 / S11-16 plans; nothing in this docs PR |

### To user
- **Q1** (CON-2, MNT-1, MNT-2): re-key the inventory. Recommendation: put the five `lifespan.*` constants at `subject.person.lifespan.*` beside Agree/Conflict, drop `lifespan` from the layer table, and state the key grammar (Capitalized = registry field; lowercase = typed map key: kind, property key, pair name, or bridge key), keeping `sex` as the declared pair name. Mirror in the plan's §6.
- **Q2** (UX-1): count the lifespan 0.1 in the worked example (2.6 / 2.75 / 5.35, still strong) and remove the "not counted" sentence; mirror in the plan's §6 item 4.
- **Q3** (UX-3, MNT-5): banners say "until then this page describes the running code"; `promote-graph-alignment.md`'s banner names §4.1, §5, §6 and §12 only.

### User answers (2026-10-17)
- **Q1:** lifespan constants go under `subject` (the design's `subject.go` holds "lifespan rules and constants"; `evidence` is the claim-evidence sheet's policy). The sex rows key on `sex_at_birth`, the pair's lead property, covering `gender` through the pair. Add the key grammar. Mirror in the plan's §6.
- **Q2:** count the 0.1 lifespan agree (2.6 / 2.75 / 5.35, strong). Mirror in the plan's §6 item 4.
- **Q3:** narrow the banners. Mirror in the plan's STEP-6.

### Fix plan 1
Docs only; no failing test applies. Every edit to `docs/matching-model.md` that copies plan text is mirrored in `docs/deployment-plan/spike-11/prs/S11-01-matching-model.md` (§6 table, §6 item 4, STEP-6) so the plan and the model agree; add a dated line under the plan's header, "**Changed in ship (2026-10-17):** lifespan keys under `subject`, `sex_at_birth` keys, key grammar, worked example counts the lifespan agree, narrower banners — user answers Q1–Q3 in the ship log."

**F-1 — Layer 2 reads `value`, not the `Near` minimums** (from CON-1)
Files: `docs/matching-model.md` layer table.
Do: the layer 2 "Registry section" cell becomes `value` (passed through to layer 1). Do not add Near minimums to layer 2 anywhere.
Check: `rg -n "Near. minimums" docs/matching-model.md` finds no layer-2 row.

**F-2 — Lifespan constants under `subject`; `sex_at_birth` keys; key grammar** (from CON-2, MNT-1, MNT-2; Q1)
Files: `docs/matching-model.md` (layer table, inventory, key description), plan §6 table.
Do: rename `lifespan.MaxYears`, `lifespan.ParentMinYears`, `lifespan.ParentMaxYears`, `lifespan.MotherMaxYears`, `lifespan.MarriageMinYears` to `subject.person.lifespan.<same Field>` and move the rows next to `subject.person.lifespan.Agree/Conflict`. Layer 3 and 4 cells read `subject` only. Rename `subject.person.sex.Agree/Conflict` to `subject.person.sex_at_birth.Agree/Conflict`, Changes column "… (the `sex_at_birth` / `gender` pair)"; rename `subject.Pairs.person.sex` to `subject.Pairs.person.sex_at_birth`. After the `section.Group.Field` sentence add: "A Capitalized part is a registry field. A lowercase part is a typed map key: a Subject kind, a property key, or a bridge or participation key; a pair compared together is keyed by its lead property (`sex_at_birth` covers `gender`)." Search the whole model for any other `lifespan.` top-level or `.sex.` key and fix it the same way.
Check: `rg -n '^\| .lifespan\.|\.sex\.' docs/matching-model.md docs/deployment-plan/spike-11/prs/S11-01-matching-model.md` prints nothing; no section other than value, subject, segment, alignment, evidence begins a key.

**F-3 — Worked example counts the lifespan agree** (from UX-1; Q2)
Files: `docs/matching-model.md` worked example, plan §6 item 4.
Do: add a step after step 3 (or fold into it): both sides derive a lifespan "born 1 May 1901" from their birth; the windows are consistent, so `subject.person.lifespan.Agree` 0.1 counts and Gracie's own evidence is 2.5 + 0.1 = 2.6. Segment stays 2.75 (credit uses the birth's own evidence 2.5). Total 2.6 + 2.75 = 5.35 ≥ 4 → strong; not contested. Delete the "With only the birth on one side…" sentence.
Check: the numbers 2.6, 2.75, 5.35 appear; "not counted" does not.

**F-4 — Narrow the banners** (from UX-3, MNT-5; Q3)
Files: `docs/matching.md`, `docs/promote-graph-alignment.md` (top banner only; section markers unchanged), plan STEP-6.
Do: `matching.md`: "**Spike 11 replaces this page.** The target model is `[matching-model.md](matching-model.md)`; until then this page describes the running code." `promote-graph-alignment.md`: "**Spike 11 replaces §4.1, §5, §6 and §12.** The target model is `[matching-model.md](matching-model.md)`; until then those sections describe the running code." Keep the existing link form used there.

**F-5 — Glossary: a developer tunes knobs** (from UX-2)
Files: `docs/matching-model.md` glossary.
Do: knob → "One registry field a developer can tune; changing it is one edit in the registry".

**F-6 — Say what the `evidence` section is** (from MNT-4)
Files: `docs/matching-model.md`.
Do: one sentence above the inventory (or in the key description): "The `evidence` section is the claim evidence sheet's line policy (link to `claim-evidence.md`), not evidence points; the Promote evidence sheet reads it, no comparison layer does."

**F-7 — One home for the shortlisting terms** (from MNT-3)
Files: `docs/deployment-plan/spike-11/comparison-layers.md` lines ~494–506.
Do: replace the five-row shortlisting definitions table with one line linking to the model's glossary (`../../matching-model.md#glossary`); keep the "Agreed 2026-10-17 … we don't use that word" decision sentence. Check that the model's glossary has all five terms.

Check for the round: `scripts/check-pr.sh` PASS bar `xcode-test`; every link in changed files resolves.

## Round 1 — fix
Head `8fd64d4`. Gate PASS (bar `xcode-test`, not macOS); links in changed docs resolve. F-1 – F-7 done. Deviations: plan STEP-6 writes the banner text without a link (plan-docs cannot resolve it from the plan's folder); worked example step 5 says "birth's own evidence".

## Round 2 — review
Scope `5408b6b...8fd64d4`. CON clean · MNT clean · UX clean. Every round-1 Fix-now finding resolved; no new blocker.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| CON-R2-a | CON | minor | Plan STEP-2 check still says sums 2.5 / 2.75 / 5.25 | Fix now (rides the landing commit, which edits the plan) | plan line ~271 vs §6 item 4 and the model (2.6 / 2.75 / 5.35) |
| CON-R2-b | CON → MNT | minor | Key grammar's lowercase list omits `lifespan` | Fix now (rides the landing commit) | `matching-model.md` key description; 7 keys use `lifespan` |

No Fix plan 2 (no round-2 blocker).

## Landing
Do, in one commit "Record S11-01 as landed in pull request 432.": the plan's Status → Landed 2026-10-17 in #432; the map row and scope entry for S11-01 in `deployment-plan.md` → Landed #432; the `completed.md` paragraph from the implementer's round-0 report (state that the model also fixes lifespan under `subject` and `sex_at_birth` keys); plus the two minors above: plan STEP-2 check → "the sums are 2.6, 2.75 and 5.35"; key grammar → "a Subject kind, a property key, a derived property (`lifespan`), or a bridge or participation key".
Head `9df157c`. Gate PASS (bar `xcode-test`). Plan Status Landed #432; map row and scope entry Landed; `completed.md` paragraph written by the implementer from the plan (no paragraph in the round-0 report); both round-2 minors applied.

## Result
Gate PASS (Linux; `xcode-test` left to CI) · CI: pending on ready · open: none · follow-ups: PRF-1 (S11-15 / S11-16 plans: cap or nominate the event shortlist bucket) · to user: Q1–Q3 answered 2026-10-17
