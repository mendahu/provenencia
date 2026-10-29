# S14-08 — Raise other-link support so a thin name plus kinship clears Weak

**Status:** Landed in [#613](https://github.com/mendahu/provenencia/pull/613) (2026-10-28)
**Landed:** `TestProposeKinshipShapes/AC-2.2` in `promotealign_test.go` now expects **Medium** for marriage-only on both sides (3.375 under raised Other support); not named in the plan, required for `check-pr`. Dogfood manual check (§10) not run in the app by the implementer.
**Spike / slice:** Spike 14 · Slice 2 — Matching calibration
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no new chrome: registry knob and tests only; Promote copy and sheet unchanged)
**Board:** —
**Depends on:** —
**Blocks:** —
**Branch:** `s14-08-other-support-calibration`
**Version bump:** patch — Promote bands move for rows that earn other-link support ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-match-rule`](../../../../.agents/skills/add-match-rule/SKILL.md)
**Rules that bind:** `.cursor/rules/match-rules.mdc`, `.cursor/rules/go-tests.mdc`, `.cursor/rules/versioning.mdc`
**Context:** [`docs/matching-model.md`](../../../matching-model.md) knob inventory (`segment.OtherSupport`); [`docs/promote-graph-alignment.md`](../../../promote-graph-alignment.md) §6 edge comparisons; dogfood Promote sheet (2026-10-28): "Jake" into Jake James Kenneth Robins with an agreeing grandparent read Little evidence (band None) at score ≈ 1.93 under Weak 2; user chose option A (blanket `OtherSupport` 0.75, `OtherCredit` unchanged) 2026-10-28
**Planned:** 2026-10-28 against `main` @ `03c97090`
**Review:** [`S14-08-other-support-calibration.review.md`](S14-08-other-support-calibration.review.md) — Fits — Ready, 2 rounds

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

A Source person recorded only as a short given name, with a matching kinship link (parent, spouse, grandparent, and every other non-birth/death link), reaches at least the Weak Promote band when that link's neighbor also lines up. Dogfood on 2026-10-28 showed "Jake" against the profile "Jake James Kenneth Robins" plus an agreeing grandparent at ≈ 1.93 points — under `alignment.Bands.Weak` (2) — so a researcher-chosen target read "Little evidence" even though the sheet showed a Same name among members and an AGREES grandparent. This PR raises only `segment.OtherSupport` so that shape clears Weak, without touching the name curve, Weak bar, or kinship-only registries.

**Why now.** The user asked to move slow on matching calibration and chose this single lever (option A) after the dogfood case.

**What stays.** `OtherCredit` 0.25; birth/death `One*` weights; `UnplacedAlike` 2; accept bar and bands; name similarity curve; claim-evidence sheet policy; no new chrome.

## 2. User stories

**US-1.** As a researcher, when I promote a Source person whose name only partly matches the record and who shares a lined-up grandparent (or other non-birth/death link) with a neighbor that itself matches, I see at least Weak evidence on that row so the assessment is not "Little evidence" / "No match" for that shape.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | Shipped registry. Subject name: one given part `Jake`. Profile name: given `Jake`, given `James`, given `Kenneth`, surname `Robins` (Western). Own evidence from `subjectcompare.Compare` on those Values | Own evidence is `2.5 × (2/4.5)` = `10/9` ≈ 1.111… (Partial Contains); under Weak alone |
| AC-1.2 | AC-1.1 plus one corresponding `relationship` segment (term `grandchild` or `grandparent` or canonical `child` key — any corresponding pair), neighbors Carol Jane Breakell vs Carol Breakell (Western), placement `Together` or `Unplaced` (neighbor evidence ≥ `UnplacedAlike`) | `graphalign.ScoreCandidate(own, segment.Evidence, 1).Score` is ≥ `alignment.Bands.Weak` (2) and &lt; `alignment.Bands.Medium` (3) |
| AC-1.3 | Shipped `matchregistry.Default().Segment.OtherSupport` | Equals `0.75` |

**US-2.** As a contributor, when I change other-link support, the layer table that hardcodes the old Other totals is updated, and docs that quote the shipped number agree with the registry.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | `TestCompare` row AC-2.5 (residence together, event_type agree only) | Expected segment evidence is `0.75 + 0.25 × 0.5` = `0.875` |
| AC-2.2 | `TestCompare` row AC-2.8 (`part_of` from, own 3, place agree) | Expected segment evidence is `0.75 + 0.25 × 3` = `1.5` |
| AC-2.3 | Inventory row `segment.OtherSupport` in `docs/matching-model.md` | Start column `0.75`; Source names the 2026-10-28 dogfood decision (S14-08) |
| AC-2.4 | `docs/promote-graph-alignment.md` §6 sentence that quotes other-link support | Reads `0.75 + 0.25 ×`, not `0.25 + 0.25 ×` |

## 3. Scope

| In | Covered by |
| --- | --- |
| Set `segment.OtherSupport` to `0.75` in the registry with an updated source comment | STEP-1, AC-1.3, D-1 |
| Update hardcoded Other totals in `segmentcompare` `TestCompare` (AC-2.5, AC-2.8) | STEP-2, AC-2.1, AC-2.2 |
| Add a Go calibration test for the Jake + grandparent score band | STEP-3, AC-1.1, AC-1.2, T-1 |
| Update knob inventory and `promote-graph-alignment.md` quoted numbers | STEP-4, AC-2.3, AC-2.4 |
| Patch VERSION bump | STEP-5, D-4 |

| Out | Home |
| --- | --- |
| Relax the partial-name similarity curve | dogfood / later calibration PR; [`name-candidates.md`](../../archive/spike-11/README.md) "What stays later" |
| Lower `alignment.Bands.Weak` | not planned for this case |
| Raise `OtherCredit` or restore old parent-scale `1.0` / `0.5` | not planned (user rejected option C, 2026-10-28) |
| Kinship-only support map / middle tier for parent·spouse·grandparent | not planned (user rejected option B, 2026-10-28) |
| Multi-Source corroboration bonus | not planned |
| Claim-evidence sheet or L10n changes | not planned |
| Changing `OneSupport` / birth·death weights | not planned |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Other-link support | `core/matchregistry/segment.go` `defaultSegment` `OtherSupport` | `0.25` with comment `carried from graphalign EdgeSupportHigh`; `OtherCredit` `0.25` |
| Segment math | `core/segmentcompare/segmentcompare.go` `CompareNeighbor` | `out.Evidence = support + credit*out.Neighbor.Evidence` with `OtherSupport` / `OtherCredit` when the key is not in `One` |
| Pair score | `core/graphalign/score.go` `ScoreCandidate` | `(r.Evidence + edge) * provenance` — additive; no corroboration term |
| Bands | `core/matchregistry/alignment.go` `Bands` | Weak `2`, Medium `3`, Strong `4`; "Little evidence" is Swift wording for chosen band None (`L10n.Promote.assessmentChosenNone`) |
| Hardcoded Other expects | `core/segmentcompare/segmentcompare_test.go` `compareCases` AC-2.5, AC-2.8 | Evidence `0.375` and `1` (from `0.25` support) |
| Dynamic expects | `core/graphalign/align_test.go` (e.g. `TestNeighborCreditStopsAtOneHop`, `TestSettlingMovesAPairAndSaysWhy`) | Read `cfg.Segment.OtherSupport` — stay green without edits when the knob moves |
| Inventory | `docs/matching-model.md` row `segment.OtherSupport` | Start `0.25`, source `graphalign EdgeSupportHigh` |
| Domain prose | `docs/promote-graph-alignment.md` §6 | Quotes `0.25 + 0.25 ×` for every other link |
| Name partial | `docs/deployment-plan/spike-11/name-candidates.md` Case 1 | "Jake"/full-name shape is the Amy case: evidence `2.5 × 0.44` ≈ 1.1; curve held on purpose |
| S11-14 history | `docs/deployment-plan/spike-11/prs/S11-14-segment-compare.md` D-5, risk table | Other links dropped from ~1.5+0.5 (fan-out≈1) to 0.25+0.25; dogfood was the intended follow-up |

**Precedent.** [#493](https://github.com/mendahu/provenencia/pull/493) / S11-14 set `OtherSupport` in `core/matchregistry/segment.go` and inventory rows the same way. Copy that one-number edit + inventory Source rewrite (as S11-43 did for `alignment.MinSimilarity` in `docs/matching-model.md`). Do not add a new `Rules` field.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | Set `OtherSupport` to **0.75**; leave `OtherCredit` at **0.25** | User option A (2026-10-28). Jake+grandparent probe: own ≈ 1.111 + segment `0.75 + 0.25×≈2.273` ≈ **2.43** → Weak, under Medium. Sex-only via parent stays `0.75 + 0.25×2.5` = 1.375 &lt; AcceptBar 2 (still not auto-placed on kinship alone) | Option B kinship-only tier — more structure than this slow pass needs. Option C `1.0`/`0.5` — lands Medium on this case; larger band move |
| D-2 | No new registry map or `Rules` field — only the existing `OtherSupport` literal | `add-match-rule` "Add a knob" is for new fields; tuning an existing number is one registry edit + inventory Source | Per-term support map — Out (option B) |
| D-3 | Add one calibration test in `core/graphalign` that asserts `OtherSupport == 0.75`, the Jake+grandparent score clears Weak and stays under Medium, using shipped `AlignConfig` / `Default()` numbers and existing `givenProp` / `nameProp` helpers (no `Patterns` map fill) | Locks the dogfood bar and the knob in CI; empty `NamePattern` falls back to `Names.Western`; graphalign already owns `ScoreCandidate` and band constants | Indexing nil `Names.Patterns` — panics on `Default()`. Putting the case only in `segmentcompare` — would not assert the band |
| D-4 | Product **patch** bump | Researcher-visible Promote bands change | Docs-only / no bump — wrong for a released behavior change |
| D-5 | Update inventory Start to `0.75` and Source to the dogfood decision (same pattern as `alignment.MinSimilarity` after S11-43) | Drift guard compares keys only; Start tracks the shipped value in practice | Leaving Start at `0.25` while code says `0.75` — invents a second "current" column |

**Rules this plan bends.** None.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `TestPartialNameWithGrandparentClearsWeak` *(new)* | test | `core/graphalign/align_test.go` `TestSexAgreementAloneStaysUnderWeak`, `TestSettlingMovesAPairAndSaysWhy` | Those assert other shapes; this locks the dogfood calibration. Same helpers (`nameMetas`, propertycomparetest, `ScoreCandidate`) |

**Builds on.** `add-match-rule` (numbers only in `matchregistry`); `segmentcompare.Rules.OtherSupport`; `graphalign.ScoreCandidate`.
**Deletes.** None — nothing becomes dead. Hardcoded expected floats in two `TestCompare` rows are rewritten, not removed.
**Next adds.** A later PR can raise `OtherCredit`, add a kinship tier, or relax the name curve without fighting a second support mechanism.

### Reach

None — no new path into existing code, no new state reachable. Band outcomes change for existing rows that already score other-link support.

### Scale

None — no write, list or loop changes.

### Schema

None — no catalog change.

### Go core

| Package | Change |
| --- | --- |
| `core/matchregistry` | `defaultSegment`: `OtherSupport: 0.75` with comment naming dogfood 2026-10-28 / S14-08 (was EdgeSupportHigh carry) |
| `core/segmentcompare` | `compareCases` AC-2.5 expect `0.875`; AC-2.8 expect `1.5` |
| `core/graphalign` | `TestPartialNameWithGrandparentClearsWeak` *(new)* — see T-1 |

Errors: none.

### FFI

None — no wire change.

### macOS

None — no app change. "Little evidence" continues to mean chosen band None; more rows clear Weak so fewer chosen targets show that label for this shape.

### Strings

None — no L10n change. Existing assessment copy stays.

## 7. Steps

**STEP-1 — Registry value**
Files: `core/matchregistry/segment.go`. Skill: `add-match-rule`.
Do: set `OtherSupport` to `0.75`. Put the decision on the line **above** the field (same form as `alignment.go` `MinSimilarity`): `// decided: user, 2026-10-28 (S14-08): dogfood Jake + grandparent under Weak at 0.25; 0.75 clears Weak (~2.43) without auto-placing kinship-alone pairs (still under AcceptBar 2)`. Leave `OtherCredit` at `0.25` with its short trailing carry comment.
Check: `go test ./core/matchregistry/ -run TestSegmentSection` passes.

**STEP-2 — Segment table expects**
Files: `core/segmentcompare/segmentcompare_test.go`. Skill: —.
Do: AC-2.5 expected evidence `0.875`; AC-2.8 expected evidence `1.5`. Do not change AC-2.6 (`0.25` is One-path arithmetic, not OtherSupport).
Check: `go test ./core/segmentcompare/ -run TestCompare` passes.

**STEP-3 — Calibration test**
Files: `core/graphalign/align_test.go` (append after `TestSexAgreementAloneStaysUnderWeak`). Skill: —. Rule: `.cursor/rules/go-tests.mdc`.
Do: implement T-1. Use `matchregistry.Default().AlignConfig()` and `nameMetas()`. Build name Values with the package helpers already used in this file (`givenProp` / `nameProp` / `nameVals` shape — empty `NamePattern` falls back to `Names.Western` in `core/valuecompare/name.go`; do **not** index `cfg.Value.Names.Patterns`, which is nil on `Default()`). Own: Jake given-only vs Jake James Kenneth Robins. Neighbors: Carol Jane Breakell vs Carol Breakell. `segmentcompare.Compare` with the same key on both sides (`vocabkey.BridgeKey{Type: vocabkey.BridgeRelationship, Term: vocabkey.BridgeTerm("grandchild")}`), `Together`, `own.Evidence`. `graphalign.ScoreCandidate(own, seg.Evidence, 1)`. Assert `cfg.Segment.OtherSupport == 0.75`; own alone &lt; Weak; total score ≥ Weak and &lt; Medium.
Check: `go test ./core/graphalign/ -run TestPartialNameWithGrandparentClearsWeak` passes.

**STEP-4 — Docs numbers**
Files: `docs/matching-model.md`, `docs/promote-graph-alignment.md`. Skill: —.
Do: inventory row Start `0.75`, Source `decided: user, 2026-10-28 (S14-08): dogfood Jake + grandparent under Weak at 0.25`. In `promote-graph-alignment.md` §6, replace the quoted `0.25 + 0.25 ×` for every other link with `0.75 + 0.25 ×`.
Check: `rg 'OtherSupport|0\\.25 \\+ 0\\.25' docs/matching-model.md docs/promote-graph-alignment.md` shows inventory 0.75 and prose 0.75+0.25; no stale other-link `0.25 + 0.25` in `promote-graph-alignment.md`.

**STEP-5 — Version bump**
Files: `VERSION`, `core/version.go`, `macos/Provenencia.xcodeproj/project.pbxproj` (every `MARKETING_VERSION` line — App Debug/Release and Tests Debug/Release, four lines). Skill: —; rule: `versioning.mdc`.
Do: bump product PATCH together per `docs/versioning.md`.
Check: `VERSION`, `core.Version`, and all four `MARKETING_VERSION` lines equal the new PATCH.

**STEP-6 — Docs / map**
Files: `docs/deployment-plan/spike-14/completed.md`, `docs/deployment-plan/spike-14/deployment-plan.md` (map row), this plan (status).
Do: per the Definition of done.
Check: links resolve.

Commits: one per behavior. Subjects: (1) "Raise other-link support so a thin name plus kinship clears Weak." (registry, segment expects, calibration test). (2) "Record the other-support calibration in matching docs." (matching-model, promote-graph-alignment, spike-14 map/completed). (3) "Bump product version for other-link support calibration."

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/graphalign/align_test.go` `TestPartialNameWithGrandparentClearsWeak` *(new)* | Jake given-only vs Jake James Kenneth Robins; Carol Jane Breakell vs Carol Breakell on a corresponding relationship segment Together; shipped registry → `OtherSupport == 0.75`; own &lt; Weak; `ScoreCandidate` ≥ Weak and &lt; Medium | AC-1.1, AC-1.2, AC-1.3, D-1, D-3 |
| T-2 | Go | `core/segmentcompare/segmentcompare_test.go` `TestCompare` AC-2.5, AC-2.8 | Updated expects `0.875`, `1.5` | AC-2.1, AC-2.2 |
Must stay green: everything `scripts/check-pr.sh` runs. Suites most likely to break: `./core/segmentcompare/`, `./core/graphalign/`, `./core/matchregistry/` (hardcoded Other totals only in segmentcompare AC-2.5/2.8; graphalign mostly reads the knob).

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left
- [ ] Sweep: each renamed or deleted name `rg`'d by its old, qualified name across `.agents/`, `.cursor/rules/`, `docs/*.md` and `L10n.swift` with no hit left; `scripts/check-guidance-names.py` passes (check-pr runs it)
- [ ] Every row of **Rules this plan bends** has the user's answer, and its rule is updated
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together) — or none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated (if gated)
- [ ] Domain doc updated if a rule in docs/*.md changed: matching-model.md inventory; promote-graph-alignment.md §6
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

1. On the dogfood project that showed Jake → Jake James Kenneth Robins with the grandparent line AGREES: reopen Promote for that Source (or re-propose). The row for Jake targeting that person reads **Weak** (or stronger), not "Little evidence" / "No match", when the target is selected. (AC-1.2)
2. Spot-check a sex-only child reached only via a parent link from a named anchor: the child still does **not** auto-file on the handle (score stays under AcceptBar); the handle may still appear as an alternative.
3. `rg 'segment\\.OtherSupport' docs/matching-model.md` shows Start `0.75` and Source naming S14-08. (AC-2.3)
4. `rg '0\\.75 \\+ 0\\.25' docs/promote-graph-alignment.md` hits the other-link sentence; `rg '0\\.25 \\+ 0\\.25' docs/promote-graph-alignment.md` has no hit for that sentence. (AC-2.4)

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| Residences, witnesses, and place `part_of` (after gates) also earn more support | Accepted under option A; still well under OneSupport 1.5. Manual check 2 watches kinship-alone placement |
| A golden hardcodes `0.25` for Other evidence | STEP-2 + `go test ./core/graphalign/... ./core/segmentcompare/...`; stop if a failure needs fixture redesign beyond expect floats |
| Score lands ≥ Medium on the Jake case | Probe said ~2.43; if a NamePattern/alias change pushes ≥ 3, stop and report (D-1 target is Weak) |

**Stop and report if:**

- `defaultSegment` or `ScoreCandidate` no longer matches §4 (e.g. OtherSupport moved behind a map).
- Clearing Weak on AC-1.2 requires changing `OtherCredit`, the name curve, or Weak (Out).
- Any test failure demands a kinship-only tier or fixture rewrite beyond expect-number updates.
- The calibration test's own evidence is not under Weak (name curve already raised) — then the PR's dogfood premise is gone; stop.

## 12. Open questions

None.

## 13. PR

**Title:** S14-08 — Raise other-link support so thin name plus kinship clears Weak
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-14/prs/S14-08-other-support-calibration.md`.
