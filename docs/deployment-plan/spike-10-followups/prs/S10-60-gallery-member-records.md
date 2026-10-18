# S10-60 — Gallery reads member records

**Status:** Draft
**Spike / slice:** Spike 10 follow-ups · likeness hardening
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no new chrome: engine read only)
**Board:** —
**Depends on:** —
**Blocks:** —
**Branch:** `s10-60-gallery-member-records`
**Version bump:** none — follow-up PRs leave `VERSION` at 0.2.0, and the product bump is cut when a release ships, as Spike 10 and Spike 11 do ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** —
**Rules that bind:** `.cursor/rules/go-tests.mdc`
**Context:** [conclusion layer §6](../../../conclusion-layer-data-model.md#6-canonical_entities), [conclusion reconciliation, depiction row](../../../conclusion-reconciliation.md), [Spike 10 decisions](../../archive/spike-10/README.md) ("A gallery is the reconciled likeness values; a preferred portrait is a display pointer"), working rules: "Spike 10 decisions hold"
**Planned:** 2026-10-18 against `main` @ `aa60b63`
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

A Person's gallery lists the likenesses the researcher accepted as that Person's, earliest recorded first. Today the gallery query reads every reconciler outcome that points at a likeness value, so a **negative** record ("this photo is not Ada"), a provisional member's record, or a denied record of the same picture becomes one of that likeness's records. In the 2026-10-18 audit repro (CON-1), a negative saved first decides the earliest likeness, leads the Person header and the Persons list, opens from the gallery tile, and is accepted as the chosen portrait. In the second repro (CON-2), a provisional member who saves the same picture as an accepted member's weak likeness hides that weak likeness from the gallery. After this PR the gallery is built from **member records only**: each outcome whose own reason a gallery lists (`kept`, `weak`).

**Why now.** S11-19 and S11-20 read likenesses through the gallery for the claim evidence sheet; they must not inherit records that are not members. The fix is one read in `core/database/likenesses`, and every consumer (Person detail tiles, header and list lead, the portrait check) follows from it.

**What stays.** The reconciler, its cached tables, and `CacheVersion` (16) do not change. Earliest is still the smallest member Observation id (first recorded), ties by value rank, and a redraw keeps its place. A preferred portrait is still honoured only when it is in the gallery. No Swift, FFI, or schema change.

## 2. User stories

**US-1.** As a researcher, when I record that a photo is **not** Ada (a negative likeness) and the same photo is also recorded as Ada elsewhere, the negative record never leads, never opens from a gallery tile, and cannot be chosen as the portrait.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | Negative N (shade 100) saved first on Album, positive B (shade 200) on Census, positive A (shade 100, same pixels as N) on Register; all three Subjects promoted onto handle H | `ForEntities(H)` has two likenesses. `[0].ObservationIDs == [B]`, `[1].ObservationIDs == [A]`; `[0].Leads` is true. N is in no likeness's `ObservationIDs` |
| AC-1.2 | The same catalog, Person detail of H | Two tiles. Tile 0's `Record.ObservationID` is B and its `Record.Reason` is `kept`; no tile's record is N |
| AC-1.3 | The same catalog, Persons list header of H | `LikenessImageFileID` is B's image, not A's |
| AC-1.4 | The same catalog, `SetPreferredLikeness(H, N)` | Refused with `canonicalentities.ErrInvalid`; the pointer row is unchanged |
| AC-1.5 | A catalog written before this PR with `preferred_likeness_observation_id = N` | `ForEntities(H)` leads with B (the earliest member); no likeness has `Preferred`; the pointer row still holds N |

**US-2.** As a researcher, when a provisional member's record carries a picture, only that record is left out; the gallery still shows the accepted members' likenesses of the same picture, weak ones included.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | Accepted A (shade 100, Register, standard trust), accepted W (shade 200, Gossip graded `low_trust`), provisional P (shade 200, Diary, standard trust) on handle H | `ForEntities(H)` has two likenesses; the second's `ObservationIDs == [W]` |
| AC-2.2 | Provisional P (shade 100, Diary) saved first, accepted A (shade 100, Register) saved second, both on handle H | One likeness with `ObservationIDs == [A]`, so the Person detail tile's record is A |

## 3. Scope

| In | Covered by |
| --- | --- |
| `likenesses.ForEntities` admits an outcome by its own reason through `autoreconcile.Reason.InGallery`, not by the value's reason | STEP-2, AC-1.1, AC-2.1, AC-2.2 |
| `Reason.InGallery` doc comment states it is read on an outcome's reason | STEP-1 |
| Tests for the negative, provisional, and weak cases, the portrait refusal, the stale pointer, the Person detail tile, and the header lead | STEP-2, STEP-3, T-1 – T-7 |
| The gallery rule (members kept or weak, first recorded, ties by rank, redraw keeps its place) written in the conclusion layer doc §6 and linked from the reconciliation doc (audit CON-3) | STEP-4 |

| Out | Home |
| --- | --- |
| Joining the gallery to the record's own `depiction_values.image_file_id` instead of trusting `auto_reconciler_values.value_image_file_id` (audit SEC-4) | [depictions and likenesses idea](../../../ideas/depictions-and-likenesses.md) — added as a Still-open bullet in STEP-4 |
| A hidden value's cached `reason` taking its best-ranked member's reason even when that member does not vote (`core/autoreconcile/pipeline.go` hidden rows) | not planned — the gallery no longer reads the value's reason; other readers show it as the value's reason, which is the documented behaviour |
| Dropping the unused `Likeness.Preferred` flag (audit MNT-3) | not planned in this folder; the audit action list keeps it |
| Any Swift, FFI, or proto change | not planned — the read change reaches every surface through `ForEntities` |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Gallery query | `core/database/likenesses/likenesses.go` `sqlGallery` | Selects `v.entity_id, v.rank, v.reason, v.value_image_file_id, f.media_type, ao.observation_id, e.preferred_likeness_observation_id`. Joins `auto_reconciler_outcomes ao` on `entity_id`, `property_id`, and `ao.value_rank = v.rank` only, so every outcome with that value rank comes back: members, attached denied and provisional records, and negatives (`against`) |
| Gallery admit | `core/database/likenesses/likenesses.go` `ForEntities` | Skips a row when `autoreconcile.Reason(reason).InGallery()` is false, where `reason` is the **value's** reason. Appends every remaining row's `ao.observation_id` to `Likeness.ObservationIDs`, sorts them, orders likenesses by `ObservationIDs[0]` then `Rank`, and sets `Preferred` / `Leads` from the pointer through `ByObservation` |
| Reason rule | `core/autoreconcile/autoreconcile.go` `Reason.InGallery` | `kept` or `weak`. Its doc comment says it is read on "a value with this reason". Only caller: `likenesses.ForEntities` |
| Outcome reasons | `core/autoreconcile/pipeline.go` `reconcile` | A voter's outcome is `kept`, `folded`, `outvoted`, or `weak`. A denied or provisional record of a displayed value's picture is **attached** to that row (`r.attached`) and keeps its own reason with that row's index. A hidden row takes its first member's reason (`reason: e.reason`), so a provisional record ranked above a weak one gives the row reason `provisional`. A negative's outcome is `against` with the index of the first row carrying its picture |
| Depiction module | `core/autoreconcile/modules.go` `moduleFor` (`ValueTypeDepiction` → `keyModule`), `keyModule.fold` | Groups by image File id; `fold` always returns false, so no depiction outcome is ever `folded`. `likeness` is `CardinalityMultiple` (`core/database/subjectvocab/registry.go`), so none is `outvoted` |
| Cache | `core/database/autoreconciler/autoreconciler.go` `CacheVersion` (16), `auto_reconciler_outcomes.reason` | Each outcome's reason is already stored; this PR only reads it |
| Detail tiles | `core/database/conclusiondetails/conclusiondetails.go` `ForEntity` (the `gallery` / `leads` / `leadOutcomes` block) | Takes each likeness's `ObservationIDs[0]` as the tile's record. Follows `ForEntities` |
| Header lead | `core/database/conclusionheaders/compose.go` (`likenesses.ForEntities(g, need)` → `likenesses.Lead`) | Follows `ForEntities` |
| Portrait write | `core/database/canonicalentities/canonicalentities.go` `SetPreferredLikeness` | Refuses an Observation `likenesses.ByObservation` cannot find in the handle's gallery. Follows `ForEntities` |
| Tests | `core/database/likenesses/likenesses_test.go` `TestForEntities`, `TestForEntitiesPreferred`, `TestSetPreferredLikeness`; `core/database/conclusiondetails/conclusiondetails_test.go` `TestForEntityLikenesses`; `core/database/conclusionheaders/conclusionheaders_test.go` `TestPersonHeaderLeadsWithEarliestLikeness`; `core/autoreconcile/autoreconcile_test.go` `TestReasonInGallery` | Cover order, redraw, pointer, refusal, a stronger negative that denies, and PNG checks. None saves a same-picture negative of equal strength, a provisional record, or a weak likeness with a provisional twin |
| Fixtures | `core/database/observations/observationstest/observationstest.go` `SaveLikeness`, `SaveCitation`, `FileArtifact`, `Crop`, `LikenessRegion`, `ImageOf` | Same shade → same pixels → one image File; a negative is saved through `SaveCitation` with `Polarity: observations.PolarityNegative` (precedent: `TestForEntitiesPreferred` "stronger negative denying the pointed image falls back") |
| Doc | `docs/conclusion-layer-data-model.md` §6 (the `preferred_likeness_observation_id` paragraph); `docs/conclusion-reconciliation.md` (the depiction paragraph under the value-type table) | Say the earliest likeness leads and the gallery "shows `kept` and `weak` values". Neither says what "earliest" means or that only member records count |

**Precedent.** The audit repro in the session scratchpad (`gallery_test.go`, `provisional_test.go`, `weak_test.go`) builds each AC's catalog with the `likenesses_test.go` fixture's helpers; T-1 – T-3 are those repros moved into the package tests. The low-trust grade is set the way `TestForEntitiesPreferred` "stronger negative denying the pointed image falls back" sets `high_trust` (`sourcecredibilitygrades.Lookup` + `sourcecredibility.Upsert`). A provisional member is made with `identityclaims.Create(... Status: identityclaims.StatusProvisional)`.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | The gallery admits an **outcome** when `autoreconcile.Reason(ao.reason).InGallery()` is true. The value's reason is no longer read: `sqlGallery` selects `ao.reason` in place of `v.reason` | An outcome with reason `kept` or `weak` is a member of the value at its rank; `against`, `denied`, `provisional`, and `no_evidence` are not. One rule, already named, decides both whether a value is listed and which records it has | A second method (`Reason.Member`) beside `InGallery` — because for depictions the member set and the gallery set are the same two reasons, and two methods would be two rules for one question. A SQL `ao.reason IN ('kept','weak')` — because the rule then lives in a string literal beside a Go method that says the same thing |
| D-2 | `InGallery` keeps returning false for `folded` and `outvoted`, and `TestReasonInGallery` is unchanged | A depiction never folds (`keyModule.fold` is false) and `likeness` is `multiple`, so neither reason occurs on a likeness outcome. Widening the rule for a case that cannot happen changes a tested contract for nothing | Adding `folded` — because no depiction can reach it; Stop condition below if one does |
| D-3 | No `CacheVersion` bump and no migration | Only the read changes; `auto_reconciler_outcomes.reason` is already stored for every outcome | Bumping to force a rebuild — because the cached rows are already right |
| D-4 | A pointer stored before this PR at a non-member (AC-1.5) is left in the row; readers fall back to the earliest because `ByObservation` no longer finds it | The working rule already says a reader honours the pointer only when it is in the gallery; the existing "unpromoted likeness falls back and keeps the pointer row" test does the same | Clearing such pointers in a migration — because the read rule already handles them and a migration is a format step for a display pointer |
| D-5 | The pointer is still captured from every scanned row before the reason check | `pointers[entity]` must be known for a handle whose rows include non-members; moving the capture after the check would lose it when the first row is a non-member | Selecting the pointer in a separate query — because the existing scan already carries it |
| D-6 | The rule is written in `docs/conclusion-layer-data-model.md` §6 as one paragraph after the pointer paragraph, and `docs/conclusion-reconciliation.md`'s depiction paragraph links to it | §6 is where the pointer and the earliest lead are described; the reconciliation doc describes values, not galleries | A new section — because the rule is three sentences |

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| Subtests in `TestForEntities` *(new)*: "a negative of the same picture is not one of its records", "a provisional record of the same picture is not one of its records", "a provisional record does not hide a weak likeness" | test | `TestForEntities` "earliest first, and a redraw keeps its place" | Extends the existing table of subtests and its fixture |
| Subtest "a stored pointer at a negative falls back to the earliest" *(new)* in `TestForEntitiesPreferred` | test | "unpromoted likeness falls back and keeps the pointer row" | Same shape: write the pointer, read the gallery, check `Leads` and the pointer row |
| Case "a negative of the same picture" *(new)* in `TestSetPreferredLikeness` "refused" | test | the `cases` table in "refused" | One more row in the existing table |
| `fixture.grade` helper *(new)* in `likenesses_test.go` | helper | the inline `sourcecredibilitygrades.Lookup` + `sourcecredibility.Upsert` in "stronger negative denying the pointed image falls back" | Two subtests now set a grade; the inline copy in the existing subtest moves into the helper |
| `fixture.saveNegative` helper *(new)* in `likenesses_test.go` | helper | the inline negative `SaveCitation` in the same subtest | Two subtests now save a negative likeness; the inline copy moves into the helper |

Searched for: `InGallery`, `Reason(`, `ao.reason`, `against`, `provisional` under `core/database/likenesses`, `core/database/conclusiondetails`, `core/database/conclusionheaders`, `core/database/canonicalentities`; `sourcecredibility.Upsert` and `PolarityNegative` in `likenesses_test.go`.

**Builds on.** `autoreconcile.Reason.InGallery` stays the one gallery rule. `likenesses.ForEntities` stays the one gallery read; `ByObservation` and `Lead` are untouched.
**Deletes.** The `v.reason` column from `sqlGallery`. The inline grade and negative-save code in "stronger negative denying the pointed image falls back" (moved into the two helpers).
**Next adds.** S11-19 / S11-20 read gallery records and get members only. The SEC-4 join (Out) is one more `JOIN` in the same query.

### Schema

None — no catalog change.

### Go core

| Package | Change |
| --- | --- |
| `core/database/likenesses` | `sqlGallery` selects `ao.reason` instead of `v.reason`. `ForEntities` scans it into `reason` and keeps the existing `InGallery` check, which now reads the outcome's reason. The package doc comment and the `Likeness.ObservationIDs` comment say "member records" |
| `core/autoreconcile` | `Reason.InGallery` doc comment: "reports whether a gallery lists a record whose outcome has this reason: a member of a displayed value, or weak evidence; a weak photo does not contradict another photo (working rule “A gallery is the reconciled likeness values”)". No code change |

Errors: none new.

### FFI

None — the response shape is unchanged.

### macOS

None — no Swift change.

### Strings

None — no copy.

## 7. Steps

**STEP-1 — Say what InGallery reads**
Files: `core/autoreconcile/autoreconcile.go`.
Do: Reword the `Reason.InGallery` doc comment as in §6 Go core. No code change.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/autoreconcile/` passes.

**STEP-2 — Admit gallery records by their own outcome**
Files: `core/database/likenesses/likenesses.go`, `core/database/likenesses/likenesses_test.go`.
Do: Add the `grade` and `saveNegative` helpers to the fixture and move the inline code of "stronger negative denying the pointed image falls back" onto them. Add T-1, T-2, T-3 and run them: T-1 and T-3 fail. Change `sqlGallery` to select `ao.reason` (D-1), keep the pointer capture before the check (D-5), and update the package and `ObservationIDs` comments.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/likenesses/` passes, T-1 – T-3 included.

**STEP-3 — Portrait and surface tests**
Files: `core/database/likenesses/likenesses_test.go`, `core/database/conclusiondetails/conclusiondetails_test.go`, `core/database/conclusionheaders/conclusionheaders_test.go`.
Do: Add T-4 (refusal case), T-5 (stale pointer, written with `UPDATE canonical_entities SET preferred_likeness_observation_id = ? WHERE id = ?` on `f.c.DB()`), T-6 (detail tile), T-7 (header lead).
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/likenesses/ ./core/database/conclusiondetails/ ./core/database/conclusionheaders/ ./core/database/canonicalentities/ ./core/database/autoreconciler/` passes.

**STEP-4 — Docs**
Files: `docs/conclusion-layer-data-model.md`, `docs/conclusion-reconciliation.md`, `docs/ideas/depictions-and-likenesses.md`, `docs/deployment-plan/spike-10-followups/deployment-plan.md`, this plan.
Do:
- `conclusion-layer-data-model.md` §6, after the `preferred_likeness_observation_id` paragraph, add: "A Person's **gallery** is its likeness values with their member records: each record whose own outcome is `kept` or `weak`. A negative, denied, or provisional record of the same picture is not one of them. The gallery is ordered by each likeness's first recorded member (the smallest Observation id), then by value rank; a redraw keeps its Observation, so it keeps its place. The earliest likeness is the first in that order."
- `conclusion-reconciliation.md`: in the depiction paragraph, replace "which shows `kept` and `weak` values" with "built from the member records whose outcome is `kept` or `weak`", followed by a Markdown link with text "conclusion layer §6" and target `conclusion-layer-data-model.md#6-canonical_entities` (relative to `docs/`).
- `docs/ideas/depictions-and-likenesses.md` Still open: add "**Gallery image from the record.** The gallery reads the image File id from the reconciler cache; joining it to the record's own `depiction_values.image_file_id` keeps an edited cache row from showing another PNG (audit SEC-4, 2026-10-18)."
- The folder has no `completed.md`: set this plan's Status and the S10-60 row Status in `docs/deployment-plan/spike-10-followups/deployment-plan.md` instead.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-10-followups/prs/S10-60-gallery-member-records.md` reports no errors; the two doc links resolve.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/database/likenesses/likenesses_test.go` `TestForEntities` "a negative of the same picture is not one of its records" *(new)* | N negative shade 100 (Album), B shade 200 (Census), A shade 100 (Register); promote A, then B and N onto it → two likenesses, `[0].ObservationIDs == [B]`, `[1].ObservationIDs == [A]`, `[0].Leads`; `ByObservation(g, N)` is not ok | AC-1.1, D-1 |
| T-2 | Go | same file, `TestForEntities` "a provisional record of the same picture is not one of its records" *(new)* | P shade 100 (Diary) saved first, A shade 100 (Register) second; promote A; provisional claim of P's Subject onto H → one likeness, `ObservationIDs == [A]` | AC-2.2, D-1 |
| T-3 | Go | same file, `TestForEntities` "a provisional record does not hide a weak likeness" *(new)* | A shade 100 (Register), W shade 200 (Gossip, `low_trust`), P shade 200 (Diary, provisional claim) → two likenesses; `ByObservation(g, W)` ok; `ByObservation(g, P)` not ok | AC-2.1, D-1 |
| T-4 | Go | same file, `TestSetPreferredLikeness` "refused" table, case "a negative of the same picture" *(new)* | The T-1 catalog's N on H → `canonicalentities.ErrInvalid`, pointer unchanged | AC-1.4 |
| T-5 | Go | same file, `TestForEntitiesPreferred` "a stored pointer at a negative falls back to the earliest" *(new)* | T-1 catalog; SQL sets the pointer to N → `Lead` is B; `preferredFlags` all false; `f.pointer(h)` still N | AC-1.5, D-4, D-5 |
| T-6 | Go | `core/database/conclusiondetails/conclusiondetails_test.go` `TestForEntityLikenessesSkipNegatives` *(new)* | T-1 catalog shape → two tiles; tile 0 `Record.ObservationID` B, `Record.Reason` `kept`; no tile record is N | AC-1.2 |
| T-7 | Go | `core/database/conclusionheaders/conclusionheaders_test.go` `TestPersonHeaderSkipsNegativeLikeness` *(new)* | `f.person("Ann Lee")`; on `f.source`, negative N shade 100 through `observationstest.SaveCitation` + `FileArtifact` with `Polarity: observations.PolarityNegative`, then B shade 200 and A shade 100 through `observationstest.SaveLikeness` on the same Subject → the `f.list()` row for the handle has `LikenessImageFileID` equal to `observationstest.ImageOf(B)`; `f.c.Graph().Verify()` passes | AC-1.3 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `TestForEntitiesPreferred`, `TestForEntityLikenesses`, `TestDepictionsReconcileByImage` (`core/database/autoreconciler/autoreconciler_test.go`, which reads the gallery).

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left — none in this PR
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New — no UI in this PR
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together) — none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] docs/deployment-plan/spike-10-followups/deployment-plan.md: S10-60 row status (this folder has no completed.md)
- [ ] Brief moved to design/archive/ and design/README.md updated (if gated) — not gated
- [ ] Domain doc updated: conclusion-layer-data-model.md §6 gallery paragraph; conclusion-reconciliation.md depiction paragraph links it; ideas/depictions-and-likenesses.md SEC-4 bullet
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

1. Open a project made by this build. Create Sources Album, Census, Register; on each, a person Subject on a photo Citation.
2. On Album, save a **negative** Likeness of a face. On Census, save a Likeness of a different face. On Register, save a Likeness of the same face region as Album's (draw the same box on the same photo, so the pixels match).
3. Promote the three Subjects onto one Person. On the Person page, the header thumb and the first gallery tile are the Census face. The Album record opens from no tile. In the Persons list, the thumb is the Census face.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| A likeness value whose only member outcomes are non-members disappears from the gallery | Intended: a value with no member record is not the Person's likeness (AC-2.2) |
| A test that counted a non-member as a record changes | T-1 – T-3 replace that expectation; no existing test asserts a non-member record (checked: §4 Tests row) |

**Stop and report if:**

- A likeness outcome in any test carries reason `folded` or `outvoted` (D-2 assumed neither occurs for depictions).
- `auto_reconciler_outcomes` does not store a `reason` column at the base, or a likeness member's stored reason is not `kept` or `weak`.
- T-1 or T-3 already passes before the STEP-2 query change.
- `TestDepictionsReconcileByImage` fails after STEP-2 in a way that needs a reconciler change (Out).

## 12. Open questions

None.

## 13. PR

**Title:** S10-60 — Gallery reads member records
**Body:** objective in one sentence; what shipped against US-1 and US-2; AC evidence (T-1 – T-7); deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-10-followups/prs/S10-60-gallery-member-records.md`.
