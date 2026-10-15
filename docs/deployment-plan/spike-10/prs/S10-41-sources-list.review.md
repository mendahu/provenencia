# Review — S10-41 Sources list

**Documents:** [`S10-41-sources-list.md`](S10-41-sources-list.md) (plan), [`../design/S10-D8-sources-list.md`](../design/S10-D8-sources-list.md) (brief)
**Lint:** `python3 scripts/check-plan-docs.py` — 0 errors, 0 warnings on both after round 1.
**Reviewers, round 1:** plan maintainability (Opus) — Fits with changes; plan executability (Sonnet) — Ready with changes; brief maintainability (Opus) — Fits with changes; brief executability (Sonnet) — Ready with changes.

## Plan — maintainability

| ID | Sev | Finding | Answer |
| --- | --- | --- | --- |
| M-1 | major | The Sources row copies S10-22's `LastUpdatedDisplay` → `PVListDetail` mapping and reuses `L10n.Conclusions.a11yList`, a Conclusions string | **Fixed.** D-10: `LastUpdatedDisplay.listDetail` and `spokenLabel(after:)` in `Platform`, with `L10n.LastUpdated.a11yJoin`; `ConclusionListRow` and `SourcesSplitRow` both call them. **To user:** move this into the S10-22 plan instead (S10-22 is in progress). |
| M-2 | major | The primaries query re-reads the subject join beside `sqlCountSubjectsBySource` | **Fixed.** D-5 replaces both subject queries with one grouped-by-type query; one loop fills subject, primary, and filed counts; `canvasTypeSQL` deleted. |
| M-3 | major | `UserDefaults` contradicts the documented per-project home (JSON under Application Support, `navigation-history.md`, `add-workspace-location`) | **Fixed.** D-12: `sources-list/{projectKey}.json` keyed by user id; `InstallPaths.projectKey` extracted from `navigationFileName`. Option (a); the documented rule settles it, so not sent to the user. |
| M-4 | minor | Four typed properties and a String struct; four identical `.onChange` handlers | **Fixed.** D-12, D-14, D-18: one typed `settings` value with field-by-field decoding, one save, one `.onChange(of: model.settings)`. |
| M-5 | minor | `SourceEvidenceStatus` split across files; “missing means empty” written twice | **Fixed.** D-7: nested `SourcesModel.EvidenceStatus` with `label` and `init(_ progress:)`. |
| M-6 | minor | FakeStore gets a fifth copy of the primary-kind list | **Fixed** for the new site (D-20, `primaryTypeKeys`). The four older sites stay; editing them is outside this PR. |
| M-7 | minor | `refreshSourceGraphProgress(for:)` repeats a mutation list the registry has | **Declined for this PR.** The registry's `.sourceGraph` list also includes `.deletedSource`, which must not patch progress, so one shared list is not a drop-in. Logged for a later cleanup. |
| M-8 | minor | A third text-matching rule | **Logged.** D-15 stays; unifying list filters and fixing Properties' accent miss is a later idea. |
| M-9 | nit | Version field; speculative Next adds; duplicated FROM/JOIN | **Fixed.** No version field (D-12); Next adds trimmed; `sqlScopeRevisions` shared (D-1). |
| M-10 | nit | Extract the cell in S10-22, not here | **To user** with M-1. D-10 extracts it in this PR and uses S10-22's version if it already exists. |

## Plan — executability

| ID | Sev | Finding | Answer |
| --- | --- | --- | --- |
| X-1 | blocker | `LatestByScope` while `List`'s rows are open deadlocks (`SetMaxOpenConns(1)`) | **Fixed.** D-2 orders the reads and names the trap; D-5 the same; Current state row and Risk added. |
| X-2 | major | Proto generation and Xcode checks cannot run on Linux | **Fixed.** Risk and Stop condition: run on macOS; never hand-edit generated code. |
| X-3 | major | `d` used in `pageZone` but built in `body` | **Fixed.** D-9 `private var updated`; D-11 keeps `Text(verbatim:)`. |
| X-4 | major | Restored type before types load flashes the empty state | **Fixed.** D-17 `effectiveTypeFilterID` and `isAwaitingFilterData`. |
| X-5 | major | Status filter when progress fails | **Fixed.** D-17 `statusCountsError`; D-18 shows a danger callout; brief SRL-8 matches. A nil progress handle counts as loading. |
| X-6 | minor | FakeStore `updatedAt`; origin in the fake's primary test | **Fixed.** D-9 leaves it empty; D-20 checks key only. |
| X-7 | minor | Settings key was pseudo-code | **Fixed.** D-12 `InstallPaths.projectKey`; whitespace or `{}` uuid → no store. |
| X-8 | minor | Status binding, where the store lives, pbxproj for both files, row init | **Fixed.** D-19 binding, D-13 store, §6 macOS Settings row, D-9 no init change. |
| X-9 | minor | Row time stale after Promote | **Declined.** Promote writes `identity_claims` and `identity_claim_evidence`, whose effects have `CacheSource` and no `Source` path (`core/database/effects/registry.go`), so `audit.ScopeSource` does not move and the time is not stale. |

## Brief — executability

| ID | Sev | Finding | Answer |
| --- | --- | --- | --- |
| X-1 | blocker | The cell has no shareable view | **Fixed.** Inventory names `PVListDetailCell`, extracted in S10-41 (plan D-10). |
| X-2 | blocker | The S10-D9 file disagrees with the agreed board (icon, 136pt, help prefix, faint dash) | **Fixed.** S10-D8 restates the cell and names the working rule and S10-22 D-5, D-6, D-11 as the source. S10-D9 itself is S10-22's brief and is archived with it. |
| X-3 | major | SRL-10 both in and out | **Fixed.** §3 splits it. |
| X-4 | major | Clear filters and scroll | **Fixed.** SRL-10 and the domain fact include it. |
| X-5 | major | Counts failure | **Fixed.** SRL-8. |
| X-6 | major | “Promoted” vs accepted claim | **Fixed.** Defined; provisional is not promoted. |
| X-7 | major | Search rules | **Fixed.** Trimmed, one phrase, substring, no hyphen folding. |
| X-8 | major | Search width | **Fixed.** 160–280pt; selects never wrap; controls leading. |
| X-9 | major | Invalid restored values | **Fixed.** Default, silently. |
| X-10 | minor | Nothing matches is Extend | **Fixed.** |
| X-11 | minor | Counting state undefined | **Fixed.** Spinner below the toolbar, no caption band. |
| X-12 | minor | Spoken wording | **Fixed.** Example string in SRL-2. |
| X-13 | minor | Cell click target | **Fixed.** Not a separate target. |
| X-14 | minor | No recorded change = revision zero | **Fixed.** |
| X-15 | minor | Date sample vs locale | **Fixed.** |

## Brief — maintainability

| ID | Sev | Finding | Answer |
| --- | --- | --- | --- |
| M-1 | major | Build the cell as its own view in S10-22 | **To user** (same as plan M-1 and M-10). |
| M-2 | minor | Status counts are the Evidence graph counts | **Fixed.** |
| M-3 | minor | Name the loading view; no skeleton | **Fixed.** |
| M-4 | nit | Search size vs the Properties precedent | **Fixed.** Regular size, stated. |
| M-5 | nit | `LastUpdatedDisplay` layer | **Fixed.** “— (Platform value)”. |
| M-6 | nit | Three matchers; Properties empty state | **Logged.** |

## Round 2 — executability (Sonnet): Ready with changes, no blockers

| ID | Sev | Finding | Answer |
| --- | --- | --- | --- |
| X-1 | major | STEP-1 removes `UpdatedRevision` but `sourceProto` changes in STEP-2, so `./api/...` does not compile | **Fixed.** STEP-1 changes the one `sourceProto` line. |
| X-2 | major | `GoStore.mapSourceGraphProgress` not assigned; Filed would never match in the app | **Fixed.** D-7 and STEP-4 name it. |
| X-3 | major | The ref has no width rule | **Fixed.** D-10 `.fixedSize` on the ref; STEP-6 narrow check. |
| X-4 | minor | Count line while counting or failed | **Fixed.** D-16, T-12, brief SRL-6. |
| X-5 | minor | `source`-only subjects would create an all-zero entry | **Fixed.** D-5 loop rule; T-4 case. |

## To user

1. **Move the cell view and the `LastUpdatedDisplay` mapping and join into S10-22** (plan M-1, M-10; brief M-1). Recommended if S10-22 has not built its cell yet: one sentence in S10-22 D-7 and D-8. The S10-41 plan already handles both cases (D-10 uses what S10-22 lands).

## Verdict

Round 1 had blockers, all fixed. Round 2 is pending: fresh reviewers check the answers above and raise only new blockers. The plan stays **Blocked — on S10-22 and board S10-D8**.
