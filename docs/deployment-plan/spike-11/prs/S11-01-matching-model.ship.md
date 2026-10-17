# Ship log — S11-01 One-page matching model

**Plan:** [`S11-01-matching-model.md`](S11-01-matching-model.md) · **Branch:** `claude/modest-gates-tunm8o` (session-designated; plan header names `s11-01-matching-model`) · **PR:** #432
**State:** review 1 — paused on To user (Q1–Q3)

Gate note: this ship runs in a Linux cloud session. `scripts/check-pr.sh` runs every check except `xcode-test` (macOS only); the PR is docs only, and CI runs the Xcode tests on the non-draft PR.

## Round 0 — implement
Head `5408b6b`. Gate PASS (all checks; `xcode-test` skipped, not macOS). Report: used the session branch; base moved `af50bd6` → `233e87a` with no Current state row invalidated; extra whitespace-fix commit `5408b6b`; landing docs (Landed #N, map row, `completed.md`) wait on the PR number and land in the finish commit.

## Round 1 — review
Base `233e87a`, head `5408b6b`. Verdicts: CON changes needed · MNT changes needed · UX clean · SEC clean · PRF clean.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| CON-1 | CON | major | Layer table gives layer 2 the `Near` minimums | Fix now → F-1 | verified: `comparison-layers.md` "Near has a minimum similarity per type. It lives in layer 1's Rules"; the parenthetical is implementer wording, not §6 |
| CON-2, MNT-1 | CON, MNT | major | `lifespan` is a sixth top-level registry section | To user → Q1 | verified: map rule "one file per section (value, subject, segment, alignment, evidence)" and `comparison-layers.md` `subject.go` row "lifespan rules and constants"; keys come verbatim from the plan's §6 table |
| MNT-2 | MNT | major | Key grammar unstated; `sex` is not a property key | To user → Q1 | verified: rows use 2–4 segments with mixed case; `subjectvocab` has `sex_at_birth` and `gender`, no `sex`; keys are §6 verbatim |
| UX-1 | UX | minor | Worked example drops the 0.1 lifespan agree though both sides have a birth | To user → Q2 | verified: `comparison-layers.md` lifespan "subject of a birth at D: born at D … Overlaps or Contains is consistent"; sentence is the plan's §6 item 4 |
| UX-3, MNT-5 | UX, MNT | minor | Banners overstate what is replaced / don't say the page describes running code | To user → Q3 | text prescribed verbatim by STEP-6; D-5 keeps §1–§3, §9 of `promote-graph-alignment.md` |
| UX-2 | UX | minor | Glossary says a researcher can tune knobs | Fix now → F-2 | README "Not this spike: researcher-facing tuning"; implementer wording, file already in the fix |
| MNT-4 | MNT | minor | "evidence" names points and the `evidence` registry section | Fix now → F-3 | glossary evidence = points; `evidence.*` rows are the claim-evidence sheet's line policy; few lines |
| MNT-3 | MNT | minor | Shortlisting terms defined twice, already drifting | Fix now → F-4 | `comparison-layers.md` table vs model glossary ("year" vs "year bucket"); file already touched; D-8 keeps decision history, only the definitions move |
| PRF-1 | PRF | minor | Event shortlist bucket grows with the catalog | Follow-up | design question for the S11-15 / S11-16 plans; nothing in this docs PR |

### To user
- **Q1** (CON-2, MNT-1, MNT-2): re-key the inventory. Recommendation: put the five `lifespan.*` constants at `subject.person.lifespan.*` beside Agree/Conflict, drop `lifespan` from the layer table, and state the key grammar (Capitalized = registry field; lowercase = typed map key: kind, property key, pair name, or bridge key), keeping `sex` as the declared pair name. Mirror in the plan's §6.
- **Q2** (UX-1): count the lifespan 0.1 in the worked example (2.6 / 2.75 / 5.35, still strong) and remove the "not counted" sentence; mirror in the plan's §6 item 4.
- **Q3** (UX-3, MNT-5): banners say "until then this page describes the running code"; `promote-graph-alignment.md`'s banner names §4.1, §5, §6 and §12 only.
