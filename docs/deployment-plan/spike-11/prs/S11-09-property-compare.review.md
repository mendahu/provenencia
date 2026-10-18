# Review — S11-09 Property compare

Plan: [`S11-09-property-compare.md`](S11-09-property-compare.md). Lint (`scripts/check-plan-docs.py`) clean before round 1.

## Round 1 (2026-10-18)

Maintainability: **Fits with changes**. Executability: **Ready with changes**.

| ID | Severity | Finding | Answer |
| --- | --- | --- | --- |
| M-1 | blocker | D-6 added a second relation ranking beside `valuecompare` `relationOrder` (`name.go`) and a class table that copied D-8's outcome mapping | **Fixed.** D-16 moves `relationOrder` to `valuecompare.go`, lists all seven relations, and exports `Stronger`; D-8 writes the mapping once (`outcomeOf`); D-6 ranks by `speakingOrder` over outcomes, then similarity, then `Stronger`. STEP-1, T-7 |
| M-2 | major | Date fixtures would be a sixth private copy; add `datevaluetest` beside `namevaluetest` | **Fixed** for this PR: D-12 adds `core/datevalue/datevaluetest` (`YMD`, `About`, `Phrase`) with a guard row; STEP-2. Moving the other copies (including `valuecompare`'s) is an Out line, "not planned — each moves when its file is next edited": rewriting other packages' tests widens a layer PR for no behavior |
| M-3 | major | Skill, editor rule and glossary **cause** row still describe layer 1 only; `propertymeta` package doc goes stale | **Fixed.** D-2 (package doc), D-13 (`match-rules.mdc` text and Check line), D-14 and STEP-8 (skill description, codes per layer, full layer package recipe, Check; glossary **cause** row) |
| M-4 | major | `Pair` takes the glossary's word for S11-10's pairs compared together | **Fixed.** D-1: `SpeakingPair`, field `Speaking`, after the glossary's **speaking pair** |
| M-5 | minor | Out and Next adds cite `possible-values.md` for a non-evidence reason, but that idea rules matching out | **Fixed.** Out row re-homed "not planned" with the strict profile as the reason; the extension-point sentences removed from Next adds and the skill bullet |
| X-1 | blocker | S11-08 has not landed; the plan does not say what to do then | **Fixed.** Header **Depends on**: branch after S11-08 merges, or from `main` if it has not, then merge `main` keeping both sides; a real conflict in a STEP-8 paragraph is a stop condition |
| X-2 | blocker | No builder named for date, text, name and term fixtures | **Fixed.** D-12 names `datevaluetest` and the helpers in `helpers_test.go`, with the fields each sets; §2 intro maps `1817`, `ABT 1817`, `spring` to builders |
| X-3 | major | STEP-3 hid five changes | **Fixed.** Split into STEP-4 (positive pairs, outcome), STEP-5 (negatives), STEP-6 (neutral rows) |
| X-4 | major | The doc's step list puts negatives after choosing the pair, against D-7 | **Fixed.** STEP-8 rewrites "What it does, in order" |
| X-5 | minor | `match-rules.mdc` says layers import no `matchregistry`, but tests do; Check line misses the package | **Fixed.** D-13 rewords the rule to non-test files and extends the Check line |
| X-6 | minor | Generator: neutral marking, failure message, subtest / cardinality mapping | **Fixed.** D-15 spells out each |
| X-7 | minor | `check-pr.sh` Xcode leg cannot run on Linux | **Fixed.** DoD line: on a host without macOS every check but `xcode-test` passes and CI runs the Xcode tests; this PR changes no Swift |

No blocker open; no To user item. Round 2 skipped: the round 1 blockers were mechanical (a shared order, fixture builders, a dependency rule) and every fix is checked by the lint and the plan's own tests.

**Verdict:** Fits with changes / Ready with changes → fixed, 1 round.
