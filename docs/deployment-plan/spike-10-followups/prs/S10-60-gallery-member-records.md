# S10-60 — The gallery reads only gallery records

**Status:** Planned
**Spike / slice:** Spike 10 follow-ups · likeness hardening
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no new chrome: engine read only)
**Board:** —
**Depends on:** —
**Blocks:** —
**Branch:** `s10-60-gallery-member-records`
**Version bump:** none — follow-up PRs leave `VERSION` at 0.2.0, and the product bump is cut when a release ships, as Spike 10 and Spike 11 do ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** —
**Rules that bind:** `.cursor/rules/go-tests.mdc` — the new cases follow each file's existing shape: narrative `t.Run` subtests in `TestForEntities` and `TestForEntitiesPreferred`, one more row in the `TestSetPreferredLikeness` "refused" table
**Context:** [conclusion layer §6](../../../conclusion-layer-data-model.md#6-canonical_entities), [conclusion reconciliation, depiction row](../../../conclusion-reconciliation.md), [Spike 10 decisions](../../archive/spike-10/README.md) ("A gallery is the reconciled likeness values; a preferred portrait is a display pointer"), working rules: "Spike 10 decisions hold"
**Planned:** 2026-10-18 against `main` @ `aa60b63`
**Review:** [`S10-60-gallery-member-records.review.md`](S10-60-gallery-member-records.review.md) — Fits / Ready with changes, 2 rounds

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

A Person's gallery lists the likenesses the researcher accepted as that Person's, earliest recorded first. Today the gallery query reads every reconciler outcome that points at a likeness value, so a **negative** record ("this photo is not Ada"), a provisional member's record, or a denied record of the same picture becomes one of that likeness's records. In the 2026-10-18 audit repro (CON-1), a negative saved first decides the earliest likeness, leads the Person header and the Persons list, opens from the gallery tile, and is accepted as the chosen portrait. In the second repro (CON-2), a provisional member who saves the same picture as an accepted member's weak likeness hides that weak likeness from the gallery. After this PR the gallery is built from **gallery records** only: each record whose own outcome reason `Reason.InGallery` admits (`kept`, `weak`). "Member" keeps its handle meaning: a Subject linked to the handle by an Identity Claim ([conclusion layer §2.2](../../../conclusion-layer-data-model.md)), accepted or provisional, as `autoreconcile.Candidate.Provisional` and the reconciliation doc use it ("a provisional member"); this plan never uses it for a record.

**Why now.** S11-19 and S11-20 read likenesses through the gallery for the claim evidence sheet; they must not inherit records that are not gallery records. The fix is one read in `core/database/likenesses`, and every consumer (Person detail tiles, header and list lead, the portrait check) follows from it.

**What stays.** The reconciler, its cached tables, and `CacheVersion` (16) do not change. Earliest is still the smallest gallery-record Observation id (first recorded), ties by value rank, and a redraw keeps its place. A preferred portrait is still honoured only when it is in the gallery. No Swift, FFI, or schema change.

## 2. User stories

**US-1.** As a researcher, when I record that a photo is **not** Ada (a negative likeness) and the same photo is also recorded as Ada elsewhere, the negative record never leads, never opens from a gallery tile, and cannot be chosen as the portrait.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | Negative N (shade 100) saved first on Album, positive B (shade 200) on Census, positive A (shade 100, same pixels as N) on Register; all three Subjects promoted onto handle H | `ForEntities(H)` has two likenesses. `[0].ObservationIDs == [B]`, `[1].ObservationIDs == [A]`; `[0].Leads` is true. N is in no likeness's `ObservationIDs` |
| AC-1.2 | The same catalog, Person detail of H | Two tiles. Tile 0's `Record.ObservationID` is B and its `Record.Reason` is `kept`; no tile's record is N |
| AC-1.3 | One person Subject on one Source (the `conclusionheaders` fixture): negative N (shade 100), then B (shade 200), then A (shade 100) | The Persons list header's `LikenessImageFileID` is B's image, not A's. One Subject on one Source is deliberate: it is the header fixture's shape, and N is no stronger than A there, so A is kept and N counts against it |
| AC-1.4 | The same catalog, `SetPreferredLikeness(H, N)` | Refused with `canonicalentities.ErrInvalid`; the pointer row is unchanged |
| AC-1.5 | A catalog written before this PR with `preferred_likeness_observation_id = N` | `ForEntities(H)` leads with B (the earliest gallery record); no likeness has `Preferred`; the pointer row still holds N |

**US-2.** As a researcher, when a provisional member's record carries a picture, only that record is left out; the gallery still shows the accepted members' likenesses of the same picture, weak ones included.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | Accepted A (shade 100, Register, standard trust), accepted W (shade 200, Gossip graded `low_trust`), provisional P (shade 200, Diary, standard trust) on handle H | `ForEntities(H)` has two likenesses; the second's `ObservationIDs == [W]` |
| AC-2.2 | Provisional P (shade 100, Diary) saved first, accepted A (shade 100, Register) saved second, both on handle H | One likeness with `ObservationIDs == [A]`, so the Person detail tile's record is A |

## 3. Scope

| In | Covered by |
| --- | --- |
| `likenesses.ForEntities` admits an outcome by its own reason through `autoreconcile.Reason.InGallery`, not by the value's reason | STEP-2, AC-1.1, AC-2.1, AC-2.2 |
| `Reason.InGallery` doc comment states it is read on an outcome's reason | STEP-1 (same commit as STEP-2) |
| `observationstest.SaveNegativeLikeness` beside `SaveLikeness`, and the tests for the negative, provisional, and weak cases, the portrait refusal, the stale pointer, the Person detail tile, and the header lead | STEP-2, STEP-3, T-1 – T-7 |
| The gallery rule (gallery records are `kept` or `weak`, first recorded, ties by rank, a redraw keeps its place) written in the conclusion layer doc §6 and linked from the reconciliation doc (audit CON-3) | STEP-4 |

| Out | Home |
| --- | --- |
| Joining the gallery to the record's own `depiction_values.image_file_id` instead of trusting `auto_reconciler_values.value_image_file_id` (audit SEC-4) | [depictions and likenesses idea](../../../ideas/depictions-and-likenesses.md) — added as a Still-open bullet in STEP-4 |
| A hidden value's cached `reason` and `Support` taking a provisional or denied entry's reason and Source (`core/autoreconcile/pipeline.go` hidden rows: `reason: e.reason`, `row.support()` over `r.members`) | [depictions and likenesses idea](../../../ideas/depictions-and-likenesses.md) — added as a Still-open bullet in STEP-4 |
| Dropping the unused `Likeness.Preferred` flag (audit MNT-3) | not planned in this folder; the audit action list keeps it |
| One shared Source-credibility test helper for the five inline copies (`autoreconciler_test.go` `credibility`, `conclusiondetails_test.go`, `conclusionheaders_test.go` `lowTrust`, `likenesses_test.go`) | not planned — logged in the review log (M-4) |
| Any Swift, FFI, or proto change | not planned — the read change reaches every surface through `ForEntities` |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Gallery query | `core/database/likenesses/likenesses.go` `sqlGallery` | Selects `v.entity_id, v.rank, v.reason, v.value_image_file_id, f.media_type, ao.observation_id, e.preferred_likeness_observation_id`. Joins `auto_reconciler_outcomes ao` on `entity_id`, `property_id`, and `ao.value_rank = v.rank` only, so every outcome with that value rank comes back: the value's voters, attached denied and provisional records, and negatives (`against`) |
| Gallery admit | `core/database/likenesses/likenesses.go` `ForEntities` | Skips a row when `autoreconcile.Reason(reason).InGallery()` is false, where `reason` is the **value's** reason. Appends every remaining row's `ao.observation_id` to `Likeness.ObservationIDs`, sorts them, orders likenesses by `ObservationIDs[0]` then `Rank`, and sets `Preferred` / `Leads` from the pointer through `ByObservation` |
| Reason rule | `core/autoreconcile/autoreconcile.go` `Reason.InGallery` | `kept` or `weak`. Its doc comment says it is read on "a value with this reason". Only caller: `likenesses.ForEntities` |
| Outcome reasons | `core/autoreconcile/pipeline.go` `reconcile` | A voter's outcome is `kept`, `folded`, `outvoted`, or `weak`. A denied or provisional record of a displayed value's picture is **attached** to that row (`r.attached`) and keeps its own reason with that row's index. A hidden row takes its first entry's reason (`reason: e.reason`), so a provisional record ranked above a weak one gives the row reason `provisional`. A negative's outcome is `against` with the index of the first row carrying its picture |
| Depiction module | `core/autoreconcile/modules.go` `moduleFor` (`ValueTypeDepiction` → `keyModule`), `keyModule.fold` | Groups by image File id; `fold` always returns false, so no depiction outcome is ever `folded`. `likeness` is `CardinalityMultiple` (`core/database/subjectvocab/registry.go`), so none is `outvoted` |
| Cache | `core/database/autoreconciler/autoreconciler.go` `CacheVersion` (16), `auto_reconciler_outcomes.reason` | Each outcome's reason is already stored; this PR only reads it |
| Detail tiles | `core/database/conclusiondetails/conclusiondetails.go` `ForEntity` (the `gallery` / `leads` / `leadOutcomes` block) | Takes each likeness's `ObservationIDs[0]` as the tile's record. Follows `ForEntities` |
| Header lead | `core/database/conclusionheaders/compose.go` (`likenesses.ForEntities(g, need)` → `likenesses.Lead`) | Follows `ForEntities` |
| Portrait write | `core/database/canonicalentities/canonicalentities.go` `SetPreferredLikeness` | Refuses an Observation `likenesses.ByObservation` cannot find in the handle's gallery. Follows `ForEntities` |
| Tests | `core/database/likenesses/likenesses_test.go` `TestForEntities`, `TestForEntitiesPreferred`, `TestSetPreferredLikeness`; `core/database/conclusiondetails/conclusiondetails_test.go` `TestForEntityLikenesses`; `core/database/conclusionheaders/conclusionheaders_test.go` `TestPersonHeaderLeadsWithEarliestLikeness`; `core/autoreconcile/autoreconcile_test.go` `TestReasonInGallery` | Cover order, redraw, pointer, refusal, a stronger negative that denies, and PNG checks. None saves a same-picture negative of equal strength, a provisional record, or a weak likeness with a provisional twin |
| Fixtures | `core/database/observations/observationstest/observationstest.go` `SaveLikeness`, `SaveCitation`, `FileArtifact`, `Crop`, `LikenessRegion`, `ImageOf` | Same shade → same pixels → one image File. `SaveLikeness` saves a positive only; the one negative-likeness save is inline in `likenesses_test.go` `TestForEntitiesPreferred` "stronger negative denying the pointed image falls back" (`SaveCitation` with `Polarity: observations.PolarityNegative`). `conclusiondetails_test.go` and `conclusionheaders_test.go` have no negative helper |
| Test fixture | `core/database/likenesses/likenesses_test.go` `fixture` (`source`, `personOn`, `promote`, `twoLikenesses`, `setPreferred`, `gallery`, `pointer`, `preferredFlags`, `leadsFlags`) | `twoLikenesses` saves A (Register, 100) then B (Census, 200). The "stronger negative denying the pointed image falls back" subtest sets `high_trust` inline with `sourcecredibilitygrades.Lookup` + `sourcecredibility.Upsert`. No `identityclaims` import |
| Doc | `docs/conclusion-layer-data-model.md` §6 (the `preferred_likeness_observation_id` paragraph); `docs/conclusion-reconciliation.md` (the depiction paragraph under the value-type table) | Say the earliest likeness leads and the gallery "shows `kept` and `weak` values". Neither says what "earliest" means or which records count |

**Precedent.** `fixture.credibility(src, key)` copies `core/database/autoreconciler/autoreconciler_test.go` `fixture.credibility` (Lookup the seeded grade, `sourcecredibility.Upsert` under `writes.Call` Action `create_source_credibility_assessment`); likenesses only ever creates, so it drops the update branch. A provisional claim is made the way `autoreconciler_test.go` `TestDepictionsReconcileByImage` "a provisional member" does: `writes.Call(f.c, writes.Op{Action: "create_identity_claim", UserID: userID}, …)` returning `identityclaims.Create(tx, userID, identityclaims.CreateInput{SubjectID: s.ID, EntityID: h, Status: identityclaims.StatusProvisional})`; the provisional Subject is never promoted. `SaveNegativeLikeness` is `SaveLikeness` with `Polarity: observations.PolarityNegative`.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | The gallery admits an **outcome** when `autoreconcile.Reason(ao.reason).InGallery()` is true. The value's reason is no longer read: `sqlGallery` selects `ao.reason` in place of `v.reason`. A record so admitted is a **gallery record** | An outcome with reason `kept` or `weak` voted for the value at its rank; `against`, `denied`, `provisional`, and `no_evidence` did not. One rule, already named, decides both whether a value is listed and which records it has | A second method (`Reason.Member`) beside `InGallery` — because for depictions the set of voting outcomes and the gallery set are the same two reasons, and two methods would be two rules for one question. A SQL `ao.reason IN ('kept','weak')` — because the rule then lives in a string literal beside a Go method that says the same thing |
| D-2 | `InGallery` keeps returning false for `folded` and `outvoted`, and `TestReasonInGallery` is unchanged | A depiction never folds (`keyModule.fold` is false) and `likeness` is `multiple`, so neither reason occurs on a likeness outcome. Widening the rule for a case that cannot happen changes a tested contract for nothing | Adding `folded` — because no depiction can reach it; Stop condition below if one does |
| D-3 | No `CacheVersion` bump and no migration | Only the read changes; `auto_reconciler_outcomes.reason` is already stored for every outcome | Bumping to force a rebuild — because the cached rows are already right |
| D-4 | A pointer stored before this PR at a record that is not a gallery record (AC-1.5) is left in the row; readers fall back to the earliest because `ByObservation` no longer finds it | The working rule already says a reader honours the pointer only when it is in the gallery; the existing "unpromoted likeness falls back and keeps the pointer row" test does the same | Clearing such pointers in a migration — because the read rule already handles them and a migration is a format step for a display pointer |
| D-5 | The pointer is still captured from every scanned row before the reason check | `pointers[entity]` must be known for a handle whose rows include records the check skips; moving the capture after the check would lose it when the first row is skipped | Selecting the pointer in a separate query — because the existing scan already carries it |
| D-6 | The rule is written in `docs/conclusion-layer-data-model.md` §6 as one paragraph after the pointer paragraph, and `docs/conclusion-reconciliation.md`'s depiction paragraph links to it | §6 is where the pointer and the earliest lead are described; the reconciliation doc describes values, not galleries | A new section — because the rule is three sentences |
| D-7 | The negative-likeness save is one exported helper, `observationstest.SaveNegativeLikeness(t testing.TB, c *database.Catalog, userID, sourceID, subjectID []byte, shade uint8) observations.Observation`, beside `SaveLikeness`. Both call one private `saveLikeness(t testing.TB, c *database.Catalog, userID, sourceID, subjectID []byte, shade uint8, polarity string) observations.Observation`. The AC-1.1 catalog is one builder, `fixture.negativeTwin() (h []byte, n, a, b observations.Observation)`, in `likenesses_test.go` beside `twoLikenesses` | Three packages save a negative likeness (T-1, T-4, T-5, T-6, T-7 and the existing "stronger negative denying the pointed image falls back" subtest); `observationstest` is the package that owns likeness fixtures. T-1 and T-5 share one catalog | A private `fixture.saveNegative` per package — because it would be three copies of one save |

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| Subtests in `TestForEntities` *(new)*: "a negative of the same picture is not one of its records", "a provisional record of the same picture is not one of its records", "a provisional record does not hide a weak likeness" | test | `TestForEntities` "earliest first, and a redraw keeps its place" | Extends the existing table of subtests and its fixture |
| Subtest "a stored pointer at a negative falls back to the earliest" *(new)* in `TestForEntitiesPreferred` | test | "unpromoted likeness falls back and keeps the pointer row" | Same shape: write the pointer, read the gallery, check `Leads` and the pointer row |
| Case "a negative of the same picture" *(new)* in `TestSetPreferredLikeness` "refused" | test | the `cases` table in "refused" | One more row in the existing table |
| `observationstest.SaveNegativeLikeness` *(new)* and private `saveLikeness` *(new)* | helper | `observationstest.SaveLikeness` | Extends it: `SaveLikeness` becomes `saveLikeness` with `observations.PolarityPositive`, the new function passes `PolarityNegative` (D-7) |
| `fixture.negativeTwin` *(new)* in `likenesses_test.go` | helper | `fixture.twoLikenesses` | Same kind of named-catalog builder; `twoLikenesses` saves A before B and has no negative, and AC-1.1 needs N, B, A in that order |
| `fixture.credibility(src sources.Source, key string)` *(new)* in `likenesses_test.go` | helper | `autoreconciler_test.go` `fixture.credibility` | Same name as the autoreconciler fixture (a test package cannot import another's fixture). It calls `must(f.t, err)` itself and returns nothing, like the file's other fixture methods (`source`, `personOn`, `promote`). Not `grade`, which in that fixture means a claim-confidence grade id |
| `fixture.provisional(s subjects.Subject, onto []byte)` *(new)* in `likenesses_test.go` | helper | `autoreconciler_test.go` `TestDepictionsReconcileByImage` "a provisional member" (inline `identityclaims.Create` under `writes.Call`) | T-2 and T-3 each need one provisional claim; one method instead of two inline copies. Calls `must(f.t, err)`, returns nothing; adds the `identityclaims` import |

Searched for: `InGallery`, `Reason(`, `ao.reason`, `against`, `provisional` under `core/database/likenesses`, `core/database/conclusiondetails`, `core/database/conclusionheaders`, `core/database/canonicalentities`; `sourcecredibility.Upsert`, `credibility`, `PolarityNegative`, `SaveLikeness` under `core/database`.

**Builds on.** `autoreconcile.Reason.InGallery` stays the one gallery rule. `likenesses.ForEntities` stays the one gallery read; `ByObservation` and `Lead` are untouched.
**Deletes.** The `v.reason` column from `sqlGallery`. The inline grade and negative-save code in "stronger negative denying the pointed image falls back" (moved onto `f.credibility` and `observationstest.SaveNegativeLikeness`).
**Next adds.** S11-19 / S11-20 read the gallery and get gallery records only; S11-20's likeness tests get a negative case from `SaveNegativeLikeness`. The SEC-4 join (Out) is one more `JOIN` in the same query.

### Schema

None — no catalog change.

### Go core

| Package | Change |
| --- | --- |
| `core/database/likenesses` | `sqlGallery` selects `ao.reason` instead of `v.reason`. `ForEntities` scans it into `reason` and keeps the existing `InGallery` check, which now reads the outcome's reason. Package doc comment becomes: "Package likenesses reads the likenesses a handle's gallery shows: its cached likeness values with their gallery records, each record whose own outcome reason a gallery lists (autoreconcile.Reason.InGallery), earliest recorded first. Lead picks the likeness a handle leads with." The `Likeness.ObservationIDs` comment stays a trailing `//` comment and becomes: "the value's gallery records, ascending; [0] is the first recorded; a negative, denied, or provisional record of the same picture is not one". |
| `core/autoreconcile` | `Reason.InGallery` doc comment becomes, in ASCII quotes like the current one: `// InGallery reports whether a gallery lists a record whose outcome has this reason:` / `// a vote for a displayed value, or weak evidence; a weak photo does not contradict` / `// another photo (working rule "A gallery is the reconciled likeness values").` No code change |
| `core/database/observations/observationstest` | `func SaveNegativeLikeness(t testing.TB, c *database.Catalog, userID, sourceID, subjectID []byte, shade uint8) observations.Observation` *(new)*, doc comment "SaveNegativeLikeness is SaveLikeness for a negative: the photo is not the Subject." `SaveLikeness` keeps its signature and calls the new private `saveLikeness(t, c, userID, sourceID, subjectID, shade, polarity string)` |

Errors: none new.

### FFI

None — the response shape is unchanged.

### macOS

None — no Swift change.

### Strings

None — no copy.

## 7. Steps

**STEP-1 — Negative-likeness fixture**
Files: `core/database/observations/observationstest/observationstest.go`, `core/database/likenesses/likenesses_test.go`.
Do: Add `saveLikeness` and `SaveNegativeLikeness` (D-7); `SaveLikeness` calls `saveLikeness` with `observations.PolarityPositive`. In `likenesses_test.go`, add `fixture.credibility(src, key)`, `fixture.provisional(s, onto)`, and `fixture.negativeTwin()`, and move the inline negative save and `high_trust` assessment of "stronger negative denying the pointed image falls back" onto `SaveNegativeLikeness` and `f.credibility(album, "high_trust")`. No product code change.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/likenesses/ ./core/database/observations/...` passes.

**STEP-2 — Admit gallery records by their own outcome**
Files: `core/database/likenesses/likenesses.go`, `core/database/likenesses/likenesses_test.go`, `core/autoreconcile/autoreconcile.go`.
Do: Add T-1, T-2, T-3 and run them: all three fail before the query change. Then change `sqlGallery` to select `ao.reason` (D-1), keep the pointer capture before the check (D-5), set the package and `ObservationIDs` comments, and reword the `Reason.InGallery` doc comment, all as in §6 Go core. STEP-1, STEP-2 and STEP-3 are one commit, "Leave a negative or provisional record of a picture out of the Person's gallery.": the fixture has no behavior of its own, the comments are true only with the query change, and STEP-3 only adds tests of that behavior.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/likenesses/ ./core/autoreconcile/` passes, T-1 – T-3 included.

**STEP-3 — Portrait and surface tests**
Files: `core/database/likenesses/likenesses_test.go`, `core/database/conclusiondetails/conclusiondetails_test.go`, `core/database/conclusionheaders/conclusionheaders_test.go`.
Do: Add T-4 (refusal case), T-5 (stale pointer: after every promote, run `UPDATE canonical_entities SET preferred_likeness_observation_id = ? WHERE id = ?` on `f.c.DB()`, then read with `f.gallery(h)`), T-6 (detail tile), T-7 (header lead).
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/likenesses/ ./core/database/conclusiondetails/ ./core/database/conclusionheaders/ ./core/database/canonicalentities/ ./core/database/autoreconciler/` passes. Commit STEP-1 – STEP-3 together (STEP-2).

**STEP-4 — Docs**
Files: `docs/conclusion-layer-data-model.md`, `docs/conclusion-reconciliation.md`, `docs/ideas/depictions-and-likenesses.md`.
Do:
- `conclusion-layer-data-model.md` §6, after the `preferred_likeness_observation_id` paragraph, add: "A Person's **gallery** is its likeness values with their gallery records: each record whose own outcome is `kept` or `weak`. A negative, denied, or provisional record of the same picture is not a gallery record. The gallery is ordered by each likeness's first recorded gallery record (the smallest Observation id), then by value rank; a redraw keeps its Observation, so it keeps its place. The earliest likeness is the first in that order."
- `conclusion-reconciliation.md`: the depiction paragraph's second sentence becomes exactly: "`likeness` is `multiple`, so a person's images form a gallery, built from the records whose own outcome is `kept` or `weak` (conclusion layer §6)." with "conclusion layer §6" made a Markdown link to `conclusion-layer-data-model.md#6-canonical_entities` (both files are in `docs/`)
- `docs/ideas/depictions-and-likenesses.md` Still open, two bullets: "**Gallery image from the record.** The gallery reads the image File id from the reconciler cache; joining it to the record's own `depiction_values.image_file_id` keeps an edited cache row from showing another PNG (audit SEC-4, 2026-10-18)." and "**A hidden value's reason and support.** A hidden reconciler row takes its first entry's reason and counts every entry's Source, so a value carried by an accepted weak photo reads `provisional` when a provisional member's copy of the same picture ranks above it. The gallery reads each record's own outcome instead (S10-60); other readers of `auto_reconciler_values.reason` still see the row's."
- The folder has no `completed.md`, and the PR number does not exist yet: the map row and this plan's Status are set to `Landed #N` in the post-PR commit, `implement-pr` Round 0 step 8 ("Record S10-60 as landed in pull request N."). This step makes only the three doc edits above, in one commit: "Document which records make a Person's likeness gallery."
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-10-followups/prs/S10-60-gallery-member-records.md` reports no errors; `grep -c '^# 6\. ' docs/conclusion-layer-data-model.md` prints 1 (the §6 heading the anchor names) (the anchor `#6-canonical_entities` resolves).

Commits: two, as STEP-2 and STEP-4 say, then the post-PR commit. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/database/likenesses/likenesses_test.go` `TestForEntities` "a negative of the same picture is not one of its records" *(new)* | `f.negativeTwin()`: Sources Album, Census, Register, one person Subject on each; N = `SaveNegativeLikeness` shade 100 on Album, then B = `SaveLikeness` shade 200 on Census, then A = `SaveLikeness` shade 100 on Register; promote A's Subject (h), then B's and N's onto h → two likenesses, `[0].ObservationIDs == [B]`, `[1].ObservationIDs == [A]`, `[0].Leads`; `ByObservation(g, N)` is not ok | AC-1.1, D-1 |
| T-2 | Go | same file, `TestForEntities` "a provisional record of the same picture is not one of its records" *(new)* | P = `SaveLikeness` shade 100 on Diary saved first, A = `SaveLikeness` shade 100 on Register second; promote A's Subject (h); `f.provisional(P's Subject, h)`, P's Subject never promoted → one likeness, `ObservationIDs == [A]` | AC-2.2, D-1 |
| T-3 | Go | same file, `TestForEntities` "a provisional record does not hide a weak likeness" *(new)* | A shade 100 on Register, W shade 200 on Gossip, P shade 200 on Diary, saved in that order; then `f.credibility(gossip, "low_trust")`; then promote A's Subject (h) and W's onto h; then `f.provisional(P's Subject, h)` → two likenesses; `ByObservation(g, W)` ok; `ByObservation(g, P)` not ok | AC-2.1, D-1 |
| T-4 | Go | same file, `TestSetPreferredLikeness` "refused" table, case "a negative of the same picture" *(new)* | In the subtest's setup, after `twoLikenesses()`: Source Album, a person Subject on it, n = `SaveNegativeLikeness` shade 100 (A's pixels), promote that Subject onto h. Case `{"a negative of the same picture", h, n.ID, userID}` → `canonicalentities.ErrInvalid`, pointer still B | AC-1.4 |
| T-5 | Go | same file, `TestForEntitiesPreferred` "a stored pointer at a negative falls back to the earliest" *(new)* | `f.negativeTwin()`; SQL sets the pointer to N; `f.gallery(h)` → `Lead` is B; `preferredFlags` all false; `f.pointer(h)` still N | AC-1.5, D-4, D-5 |
| T-6 | Go | `core/database/conclusiondetails/conclusiondetails_test.go` `TestForEntityLikenessesSkipNegatives` *(new)* | With that package's fixture (`f.source`, `f.personOn`, `f.promote`, `f.detail`): N = `observationstest.SaveNegativeLikeness` shade 100 on Album, B = `SaveLikeness` shade 200 on Census, A = `SaveLikeness` shade 100 on Register; promote A's Subject (h), then B's and N's onto h → two tiles; tile 0 `Record.ObservationID` B, `Record.Reason` `kept`; no tile record is N | AC-1.2 |
| T-7 | Go | `core/database/conclusionheaders/conclusionheaders_test.go` `TestPersonHeaderSkipsNegativeLikeness` *(new)* | `s, _, h := f.person("Ann Lee")`; on `f.source`, N = `observationstest.SaveNegativeLikeness` shade 100, then B shade 200 and A shade 100 through `observationstest.SaveLikeness`, all on `s` → the `f.list()` row for h has `LikenessImageFileID` equal to `observationstest.ImageOf(t, f.c, B.ID)`; `f.c.Graph().Verify()` passes | AC-1.3 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `TestForEntitiesPreferred`, `TestForEntityLikenesses`, `TestDepictionsReconcileByImage` (`core/database/autoreconciler/autoreconciler_test.go`, which reads the gallery).

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests, catalog and plan-doc checks) — on the user's Mac, as ship-pr "Before you start" requires
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left — none in this PR
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New — no UI in this PR
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together) — none
- [ ] Post-PR commit (implement-pr Round 0 step 8): this plan's Status and the S10-60 row in docs/deployment-plan/spike-10-followups/deployment-plan.md → Landed #N; a Landed note lists every difference from the plan (this folder has no completed.md)
- [ ] Brief moved to design/archive/ and design/README.md updated (if gated) — not gated
- [ ] Domain doc updated: conclusion-layer-data-model.md §6 gallery paragraph; conclusion-reconciliation.md depiction sentence links it; ideas/depictions-and-likenesses.md SEC-4 and hidden-row bullets
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

1. Open a project made by this build. Create Sources Album, Census, Register. Add the same photo file as an Artifact on Album and on Register, and a different photo on Census. On each Source, cite its photo and add a person Subject.
2. On Album, save a **negative** Likeness of a face. On Census, save a Likeness of a different face. On Register, save a Likeness of the same face region as Album's (draw the same box on the same photo, so the pixels match).
3. On Register, Promote its Subject to a new Person; on Census and Album, Promote each Subject onto that Person (the Promote page's existing-Person choice). On the Person page, the header thumb and the first gallery tile are the Census face. The Album record opens from no tile. In the Persons list, the thumb is the Census face.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| A likeness value with no gallery record disappears from the gallery | Intended: a value no accepted record voted for is not the Person's likeness (AC-2.2 drops P's record; a value carried only by P would drop) |
| A test that counted a skipped record changes | T-1 – T-3 replace that expectation; no existing test asserts a negative, denied, or provisional gallery record (checked: §4 Tests row) |

**Stop and report if:**

- `rg -n 'ReasonFolded|ReasonOutvoted|"folded"|"outvoted"' core/database/likenesses core/database/autoreconciler/autoreconciler_test.go` shows a hit inside `core/database/likenesses` or inside `TestDepictionsReconcileByImage` (D-2 assumed neither reason occurs for depictions). The three hits at the base, `autoreconciler_test.go` lines 527, 579 and 1018, are name-property tests and are expected.
- `auto_reconciler_outcomes` has no `reason` column at the base: `rg -n -A8 'CREATE TABLE auto_reconciler_outcomes' core/database/migrations/*.sql | rg reason` prints nothing (at the base it prints `000040.sql-25-	reason         TEXT NOT NULL,`).
- Any of T-1 – T-3 already passes before the STEP-2 query change.
- In T-7, N denies A (A's outcome reason is `denied`): the single-Source shape behaves unlike AC-1.3 says.
- `TestDepictionsReconcileByImage` fails after STEP-2 in a way that needs a reconciler change (Out).

## 12. Open questions

None.

## 13. PR

**Title:** S10-60 — The gallery reads only gallery records
**Body:** objective in one sentence; what shipped against US-1 and US-2; AC evidence (T-1 – T-7); deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-10-followups/prs/S10-60-gallery-member-records.md`.
