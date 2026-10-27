# S11-44 — Des as a nickname of Desiree and Desirae

**Status:** Landed in #594 (2026-10-27)
**Landed:** No differences from the plan.
**Spike / slice:** Spike 11 · Slice 10 — Name candidates
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no new chrome: a seed edit; Promote's target menu already shows any alias hit)
**Board:** —
**Depends on:** S11-28 (landed [#589](https://github.com/mendahu/provenencia/pull/589))
**Blocks:** —
**Branch:** `s11-44-des-nickname`
**Version bump:** none — Spike 11 PRs leave `VERSION` at 0.2.0, and the product bump is cut when a release ships ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-seeded-vocabulary`](../../../../.agents/skills/add-seeded-vocabulary/SKILL.md), [`add-match-rule`](../../../../.agents/skills/add-match-rule/SKILL.md), [`add-name-value`](../../../../.agents/skills/add-name-value/SKILL.md)
**Rules that bind:** `.cursor/rules/go-tests.mdc`, `.cursor/rules/seeded-vocabulary.mdc`
**Context:** [Case 2](../name-candidates.md#case-2--desiree-against-des) in [`name-candidates.md`](../name-candidates.md); [`docs/seeded-vocabulary.md`](../../../seeded-vocabulary.md) §4.5; [S11-28 plan](S11-28-alias-fold.md) (Landed note); working rules: "Des is a nickname of Desiree and of Desirae"
**Planned:** 2026-10-27 against `main` @ `5312337`
**Review:** [`S11-44-des-nickname.review.md`](S11-44-des-nickname.review.md) — round 1 Fits with changes / Ready with changes; descoped by the user's answer; round 2 Fits / Ready

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

A Source person named "Des" is offered the record "Desiree" (or "Desirae") in Promote's target menu, and the other way round. Two nickname sets, `{desiree, des}` and `{desirae, des}`, join the product nickname table, so every catalog created after this PR has them, and the reconciler folds "Des" into the formal name within one entity.

**Why now.** Dogfood on 2026-10-24: "Desiree" was not offered "Des" ([case 2](../name-candidates.md#case-2--desiree-against-des)).

**What stays.** Alias rows are catalog data written once at create (`namealiases.Install`), and nothing heals on open. The user's existing project gets the new sets, and S11-48's Religion and Languages spoken, from a one-off command the user runs by hand after this PR merges, with the app closed (user 2026-10-27: "We don't need to build anything for this. It's a one-off command."). That command is not part of this PR.

## 2. User stories

**US-1.** As a researcher, when a Source names "Des" and my tree has "Desiree", Promote offers Desiree as a target, so that a nickname record finds its formal-name person.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | A new catalog; rules from its alias rows; compare given "Des" with given "Desiree", and "Des" with "Desirae" | Each scores `matchregistry.Default().Value.Words.AliasCredit` (0.98) |
| AC-1.2 | Same catalog; compare "Desiree" with "Desirae" | Scores below `AliasCredit`: the two sets share only `des` and stay separate |
| AC-1.3 | The reconciler's rules (`autoreconciler.ValueRules`): given "Des" and "Desiree" (surname Robins) in one entity; then "Des" and "Desirae"; then "Desiree" and "Desirae" | One value, given "Desiree"; one value, given "Desirae"; both given names kept |

## 3. Scope

| In | Covered by |
| --- | --- |
| `{desirae, des}` and `{desiree, des}` in `nicknameSets` | STEP-1, AC-1.1, AC-1.2 |
| `autoreconciler.CacheVersion` 19 with its changelog line; `productAliasPin` re-pinned; the set-count guard at 1229 | STEP-1, AC-1.3 |
| Docs: `completed.md`, map row status, this plan | STEP-2 |

| Out | Home |
| --- | --- |
| Adding the new sets, and S11-48's `religion` / `languages_spoken`, to an existing catalog | not planned in code: a one-off command the user runs by hand after merge (user 2026-10-27) |
| A committed command, FFI handler or menu item for that | not planned (user 2026-10-27) |
| Removing seed pairs an older catalog still has (audit L-4) | not planned (`docs/deployment-plan/spike-12/app-health-audit.md` L-4) |
| Prefix names ("Des" for "Desmond" with no set) | S11-46, S11-45 |
| An editor for alias tables | not planned |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Nickname seed | `core/database/namealiases/nicknames_seed.go` `nicknameSets` | 1227 sets, head first, sorted by head. No `des`, `desiree` or `desirae` token (`rg '"des"\|"desiree"\|"desirae"' core/database/namealiases` → no hit). `{"derrick", "eric", "rick", "ricky"}` is followed by `{"deuteronomy", "duty"}` |
| Set count guard | `core/database/namealiases/registry_test.go` `TestNicknameDirectionsCoverTwoWayPairs` | Asserts `len(nicknameSets) == 1227`, 108 directions, 35 kept, 143 two-way pairs |
| Cache version | `core/database/autoreconciler/autoreconciler.go` `CacheVersion` | 18; the changelog comment above it ends with line 18 (S11-28) |
| Seed pin | `core/database/autoreconciler/autoreconciler_test.go` `productAliasPin`, `TestProductAliasTablesPinCacheVersion` | `{CacheVersion: 18, Hash: "dd2d088826d1caf666c0013ff8599ea787d4b0b9903d4819b8f5b397c5e74be0"}`; on a seed change the failure message prints the new hash |
| Fold check | `autoreconciler_test.go` `TestValueRulesFoldNicknamesOnly` | Table of `{a, b, want}`; `want == ""` means both given names are kept |
| Alias scoring | `core/database/namealiases/namealiases_test.go` `TestWarmHoldsAliasTablesUntilInvalidated`; helpers `newCatalog`, `mustDB`, `compareGiven`, `assertScore` | Builds rules with `Rules(c.Graph(), db, matchregistry.Default().Value)` after `Install` and `Warm`, and asserts a score |

**Precedent.** S11-28 ([#589](https://github.com/mendahu/provenencia/pull/589)): a product seed change plus a `CacheVersion` line and a re-pinned `productAliasPin`. Copy its changelog style and its test rows.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | Insert `{"desirae", "des"},` then `{"desiree", "des"},` after `{"derrick", "eric", "rick", "ricky"},` and before `{"deuteronomy", "duty"},` | The map's call (user 2026-10-24); head first so S11-28's fold keeps the formal name; sorted by head like the rest | One set `{desiree, desirae, des}` — joins two formal names |
| D-2 | `CacheVersion = 19`, changelog line `//	19: Des is a nickname of Desiree and of Desirae (S11-44).`; `productAliasPin` `{CacheVersion: 19, Hash: <the hash the failing test prints>}`; `TestNicknameDirectionsCoverTwoWayPairs` expects 1229 sets. Directions (108), kept (35) and two-way (143) are unchanged: neither new pair is listed both ways | S11-28's recipe; the pin fails until both move | — |
| D-3 | No code reaches existing catalogs | User 2026-10-27 | A command in the repo |

**Rules this plan bends.** None.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `TestDesIsANicknameOfDesireeAndDesirae` *(new)* | test | `TestWarmHoldsAliasTablesUntilInvalidated` | Same harness, one named case for this PR's sets |

**Builds on.** The product nickname seed and S11-28's seed pin.
**Deletes.** None — nothing becomes dead.
**Next adds.** Another product nickname set is one seed line, a `CacheVersion` line, a re-pin and the set count.

### Reach

None — no new path into existing code; a new set is one more row the alias index already reads.

### Scale

None — two sets in a list of 1227.

### Schema

None — no catalog change.

### Go core

| Package | Change |
| --- | --- |
| `core/database/namealiases` | `nicknames_seed.go`: D-1. `registry_test.go`: 1229. `namealiases_test.go`: T-1 |
| `core/database/autoreconciler` | `autoreconciler.go`: D-2. `autoreconciler_test.go`: `productAliasPin`, T-2 rows |

### FFI

None — no wire change.

### macOS

None — no Swift change.

### Strings

None — no copy.

## 7. Steps

**STEP-1 — Des sets in the seed**
Files: `core/database/namealiases/nicknames_seed.go`, `core/database/namealiases/registry_test.go`, `core/database/namealiases/namealiases_test.go`, `core/database/autoreconciler/autoreconciler.go`, `core/database/autoreconciler/autoreconciler_test.go`. Skill: `add-match-rule`, `add-name-value`.
Do: D-1, D-2. Run T-3 once and copy the hash it prints into `productAliasPin`. Add T-1 and T-2.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/namealiases/... ./core/database/autoreconciler/... ./core/autoreconcile/...` passes.

**STEP-2 — Docs**
Files: `docs/deployment-plan/spike-11/completed.md`, `docs/deployment-plan/spike-11/deployment-plan.md` (map row status), this plan (status, Landed note).
Do: per the Definition of done.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-11/prs/S11-44-des-nickname.md` passes.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/database/namealiases/namealiases_test.go` `TestDesIsANicknameOfDesireeAndDesirae` *(new)* | `newCatalog`, `Install`, `Warm`, rules from `Rules(c.Graph(), mustDB(t, c), matchregistry.Default().Value)`: `assertScore` "Des" / "Desiree" and "Des" / "Desirae" at `AliasCredit`; `compareGiven` "Desiree" / "Desirae" below `AliasCredit` | AC-1.1, AC-1.2 |
| T-2 | Go | `core/database/autoreconciler/autoreconciler_test.go` `TestValueRulesFoldNicknamesOnly`, three rows | `{"Des", "Desiree", "Desiree"}`, `{"Des", "Desirae", "Desirae"}`, `{"Desiree", "Desirae", ""}` | AC-1.3 |
| T-3 | Go | `autoreconciler_test.go` `TestProductAliasTablesPinCacheVersion` | Re-pinned at 19 → passes | D-2 |
| T-4 | Go | `core/database/namealiases/registry_test.go` `TestNicknameDirectionsCoverTwoWayPairs` | 1229 sets, 108 directions, 35 kept, 143 two-way | D-2 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: reconciler and search tests that would fold a "Des" (none at the planning base: `rg -il '"des"' core --type go` lists no file).

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks); off macOS, report each Go, guidance and plan-doc check run and name xcodebuild as not run (no Swift change)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left — none: no copy
- [ ] Sweep: each renamed or deleted name `rg`'d by its old, qualified name across `.agents/`, `.cursor/rules/`, `docs/*.md` and `L10n.swift` with no hit left; `scripts/check-guidance-names.py` passes (check-pr runs it) — nothing renamed
- [ ] Every row of **Rules this plan bends** has the user's answer, and its rule is updated — none
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New — no UI
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together) — none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status `Landed #N`; completed.md: a section at the top, `## S11-44 — Des as a nickname of Desiree and Desirae`, `**Landed YYYY-MM-DD** in [#N](https://github.com/mendahu/provenencia/pull/N).` and one sentence: new projects treat Des as a nickname of Desiree and of Desirae, so Promote offers one for the other and the reconciler shows the formal name
- [ ] Brief moved to design/archive/ and design/README.md updated (if gated) — not gated
- [ ] Domain doc updated if a rule in docs/*.md changed — none changed
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

1. Build the app from this branch and create a new project.
2. File a person "Desiree Long". On a Source, record a person "Des Long" and open Promote: the row's target menu offers Desiree Long.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| An entity with members named Des, Desiree and Desirae | S11-28's fold picks the best supported head; out of this PR's AC |

**Stop and report if:**

- `nicknameSets` already holds `des`, `desiree` or `desirae` at your base.
- `TestNicknameDirectionsCoverTwoWayPairs` reports directions, kept or two-way counts other than 108, 35, 143 after D-1.
- A test outside §8 changes result because of the new sets.

## 12. Open questions

None.

## 13. PR

**Title:** S11-44 — Des as a nickname of Desiree and Desirae
**Body:** objective in one sentence; what shipped against US-1; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-11/prs/S11-44-des-nickname.md`.
