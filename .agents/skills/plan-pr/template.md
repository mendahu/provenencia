# PR plan template

Copy everything below the line into
`docs/deployment-plan/<open-spike>/prs/S#-NN-<slug>.md`. Keep every heading.
Write `None —` and the reason when a section does not apply. Copy the
**Implementer** block verbatim. Guidance in `<!-- -->` comments is for the
planner; delete it from the finished plan.

---

# S#-NN — Title in plain words

**Status:** Draft | Planned | Blocked — on … | In progress | Landed in [#N](https://github.com/mendahu/provenencia/pull/N) (YYYY-MM-DD) | Descoped
**Spike / slice:** Spike N · Slice K — slice name
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** [S#-D#](../design/S#-D#-slug.md), board agreed YYYY-MM-DD | — (no new chrome: one clause why)
**Board:** where the implementer sees the agreed frames | —
**Depends on:** S#-NN (landed [#N](…)), S#-D# | —
**Blocks:** S#-NN | —
**Branch:** `s#-nn-slug`
**Version bump:** none | patch | minor — one clause why ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-…`](../../../../.agents/skills/add-…/SKILL.md), …
**Rules that bind:** `.cursor/rules/….mdc`, … | —
**Context:** [feature doc §](../feature.md#section), [domain model §](../../../model.md#section), working rules: "rule name", …
**Planned:** YYYY-MM-DD against `main` @ `abc1234`
**Review:** [`S#-NN-slug.review.md`](S#-NN-slug.review.md) — verdict, rounds | — (not reviewed yet)

<!-- Status Planned only after the self-review gate, check-plan-docs.py, and review-plan pass, and Open questions is empty. -->

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

<!-- 2–4 sentences. The outcome for the researcher, not the code change. Where it came from (dogfood note, idea, working rule). What stays as it is. -->

**Why now.** …

**What stays.** …

## 2. User stories

<!-- 1–5 stories. Engine-only PRs may use "As the engine" or "As a contributor". Each AC is a concrete input → result. Every AC is covered by a test (T-#) or the manual check. -->

**US-1.** As a researcher, when …, I … so that ….

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | … | … |
| AC-1.2 | … | … |

## 3. Scope

| In | Covered by |
| --- | --- |
| … (brief requirement ID if any) | STEP-#, AC-# |

| Out | Home |
| --- | --- |
| … | S#-NN / [`ideas/…`](../../../ideas/….md) / dogfood log / not planned |

## 4. Current state

<!-- Verified at the planning base. Exact paths and symbols. What it does today and what is wrong with it for this PR. Include the tests that cover it today. -->

| Area | Where | Today |
| --- | --- | --- |
| … | `core/…/file.go` `FuncName` | … |
| … | `macos/App/Features/…/View.swift` `TypeName` | … |
| Tests | `…_test.go` `TestName` | … |

**Precedent.** [#N](…) / `abc1234` did the same shape for …. Copy …; do not copy ….

## 5. Decisions

<!-- Every choice the implementer would otherwise face. Product calls the user answered, and technical calls made here. Binding. -->

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | … | … | … — because … |
| D-2 | … | … | … |

## 6. Design

<!-- Only the layers this PR touches; write "None —" for the rest. Short sketches for signatures, DDL, proto. Mark new names (new). -->

### Codebase fit

<!-- The reuse survey. Every new file, type, function, table, FFI method, component, L10n namespace, and test helper this PR adds is a row. Search before you write the row; name what you searched for. review-plan reads this table first. -->

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `…` *(new)* | type / func / file / table / FFI / component / key / helper | `path` `Symbol` | … |

**Builds on.** The patterns and registries this PR extends instead of adding its own: skill, registry, precedent.
**Deletes.** Code, strings, tests, stubs, and docs this PR makes dead, and removes. | None — nothing becomes dead.
**Next adds.** Later PRs in the map or ideas this shape makes easier, and how (a registry entry, a new case). | None named.

### Schema

None — no catalog change. | Migration `core/database/migrations/NNNNNN.sql` *(new)*:

```sql
-- sketch
```

Registered in `catalogmodel` as …. Delete class: …. Audit: ….

### Go core

| Package | Change |
| --- | --- |
| `core/…` | `func Name(…) (…, error)` *(new)* — … |

Errors: `domain.code` → when ….

### FFI

None — | `api/proto/engine.proto`: `Method` *(new)*, request `…`, response `…`. Handler `api/ffi/handlers/….go`. Regenerate with `./scripts/generate-proto.sh`.

### macOS

<!-- From the brief's inventory when there is one: building block → kit component → home. Name the view types, session/cache keys, navigation places, states (empty, loading, error, huge). If the implementer cannot see the board, describe each frame in words. -->

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| … | `PV…` | `macos/App/…` | … |

States: empty …, error …, large …. Accessibility: labels / rotor / keyboard ….

### Strings

None — | L10n keys *(new)*: `L10n.….…` = "…". Removed: ….

## 7. Steps

<!-- Ordered. Each step leaves the tree building and its check passing. About 10 steps is the limit for one PR. Last step is docs. -->

**STEP-1 — …**
Files: `…`. Skill: `add-…`.
Do: ….
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/…` passes; ….

**STEP-2 — …**
Files: …. Skill: ….
Do: ….
Check: ….

**STEP-N — Docs**
<!-- A spike close (S#-99) drops this step and the DoD lines for this plan's status and completed.md: archive-docs "Spike close — checklist". -->
Files: `completed.md`, `deployment-plan.md` (map row), this plan (status), `design/` (brief → `design/archive/`), `docs/*.md` if a domain rule changed.
Do: per the Definition of done.
Check: links resolve.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

<!-- Named, with input → expected. Go: table-driven (.cursor/rules/go-tests.mdc). Swift: Swift Testing (add-swift-test). Migrations are not tested directly (add-catalog-migration); test the behavior they enable. -->

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/…/x_test.go` `TestX` *(new)* | … → … | AC-1.1 |
| T-2 | Swift | `macos/ProvenenciaTests/…Tests.swift` *(new)* | … → … | AC-1.2, D-2 |

Must stay green: everything `scripts/check-pr.sh` runs. Name any suite this PR is most likely to break: ….

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together) — or none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated (if gated)
- [ ] Domain doc updated if a rule in docs/*.md changed: …
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

<!-- What the reviewer does in the app. Makes the slice Check true for this PR. Exact data to use. -->

1. Open a project with …. Do …. See ….
2. ….

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| … | … |

**Stop and report if:**

- `…` no longer does what §4 says.
- … cannot pass without changing … (Out).
- The brief's frame and AC-… disagree.
- ….

## 12. Open questions

None. <!-- Must be "None." for status Planned. Otherwise list each with a recommended answer. -->

## 13. PR

**Title:** S#-NN — …
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/<spike>/prs/S#-NN-slug.md`.
