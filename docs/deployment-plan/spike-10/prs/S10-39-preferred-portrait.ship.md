# Ship log — S10-39 Preferred portrait

**Plan:** [`S10-39-preferred-portrait.md`](S10-39-preferred-portrait.md) · **Branch:** `s10-39-preferred-portrait` · **PR:** #450
**State:** ready

Base `a439ae24` (main).

## Round 0 — implement
Head `b544f736`. Gate PASS (orchestrator run). Report deviations:
- `WorkspaceDestinationHost` adds `.id(navigation.currentLocation.entityId)` to `PersonDetailView` so the `@State` portrait model does not keep a stale entity id.
- `ThrowingStore` (`OnboardingModelTests.swift`) gained a `setPreferredLikeness` stub; the test target would not build without it.
- STEP-2 and STEP-3 done together; T-1 sets the pointer through `SetPreferredLikeness`.
- D-15 evaluation done by reading the two files; no shell extracted.
- Claude Design boards not opened; UI follows plan §6 in words.
- Not run: AC-2.1 menu, AC-2.4 VoiceOver, AC-1.1/1.3 pixels, §10 manual check.

## Round 1 — review
Base `a439ae24`, head `b544f736`. Verdicts: CON changes needed · MNT changes needed · UX clean · SEC clean · PRF clean.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| CON-1 | CON | major | UI-only ACs (AC-2.1, 2.3, 2.4) and the D-16 menu-keyboard regression never run | Manual check (user) | Not reachable by a unit test (popup window, NSEvent monitor); this is plan §10 steps 2, 4, 6, which the user runs before merge. Listed under Needs you |
| CON-2 | CON | minor → major | AC-3.2's "stronger negative denies B's image" trigger has no portrait test | Fix now → F-1 | Re-rated: §9 asks for evidence for every AC; this half of AC-3.2 has none, and `ForEntities` fills `pointers` before the `InGallery`/`ReadableImage` filter, so a reorder would pass every test |
| CON-3 | CON | minor | Claude Design boards not opened | Done by orchestrator | Read *Person Detail.dc.html* frame 2l. It matches §6 except: toast bottom-trailing (§6: top trailing); toast body "The header and the Persons list now lead with the likeness from …" (§6: "Using the likeness from {Source} as the portrait."); no menu panel title (§6: "Person portrait"). The code follows the Decisions; differences go in the PR body. No stop (placement of Badge, button, menu matches) |
| CON-4 | CON | nit | `TestForEntitiesPreferred` is subtests, not a table | Logged | Same cases covered |
| MNT-1 | MNT | major | Fifth private `nullBlob` copy | Follow-up (M-4) | The plan's Codebase fit row chooses a private copy and its Out line queues `database.NullBlob` as M-4; the plan's Decisions bind. Raised with the user as a possible M-4 pull-forward |
| MNT-2 | MNT | minor | Handler test shadows package `contains`; fourth inline detail decode | Follow-up | `conclusion_detail_test.go` not otherwise touched by the fix plan |
| MNT-3 | MNT | minor | `sameBools` reimplements `slices.Equal` | Fix now → F-1 (rides along) | Same file as F-1, a few lines |
| MNT-4 | MNT | minor | `set_preferred_likeness` test wrapper written three times | Follow-up (M-3) | Add the three sites to M-3's task text |
| MNT-5 | MNT | minor | `coverMenuItemCount` Int only tested `> 0`; dead `[]` branch | Follow-up | File not in the fix plan |
| MNT-6 | MNT | minor | `preferred` vs `isPortrait` naming | Rejected | D-11 names `isPortrait` |
| MNT-out | MNT→UX | nit | Portrait `.pvContextMenu` attached to Event/Place headers too | Logged | Inert: trigger gated on `showsPortraitMenu` |
| UX-1 | UX | minor | Refused write toasts "Invalid person, event, or place record." and leaves the stale tile | Follow-up | Meets AC-1.5 as written; reachable only from a stale page. Refresh-on-failure plus a `canonicalentities.not_in_gallery` code |
| UX-2 | UX | minor | Focus lost when the focused "Use as portrait" button becomes the Badge | Follow-up | Same gap as the Source artifact card; fix both together |
| UX-3 | UX | minor | "Use earliest likeness" not reachable by keyboard alone | Follow-up | Shares the gap with the Source cover; D-13 rejects a visible button; kit-level change |
| UX-4 | UX | nit | Two likenesses from one Source read the same | Logged | Inherited from S10-35 `likenessTileLabel` |
| UX-5 | UX | nit | "as portrait" vs "as the portrait"; right-click hint for VoiceOver | Logged | Plan §6 copy |
| SEC-out | SEC→CON | nit | `SetPreferredLikeness` does not refuse a merged handle or check subject type | Logged | Gallery check makes both harmless (a merged or non-Person handle has no gallery to match) |
| CON-out | CON→UX | nit | Composer clear menu reset one update later; first frame may show old highlight | Logged | Covered by §10 step 6 |
| PRF-1 | PRF | minor | Hidden Persons list reloads in full after any handle write | Follow-up | Existing routing in `CatalogQueryRegistry.swift`; idea: patch rows via `PersonsByIDs` |

### Fix plan 1

**F-1 — Prove a denied portrait image falls back to the earliest** (from CON-2, MNT-3)
Files: `core/database/likenesses/likenesses_test.go`.
Test first: a new `TestForEntitiesPreferred` case, "stronger negative denying the pointed image falls back". On a fresh fixture with likenesses A (earliest) and B, set the pointer to B with the existing `(*fixture).setPreferred`, then file a stronger negative on B's image the way S10-35's `TestDepictionsReconcileByImage` does (reuse its helper/setup; do not add a new mechanism). Expect: gallery is [A]; `likenesses.Lead` returns A; no likeness has `Preferred` set; the stored `preferred_likeness_observation_id` is still B. The production code is expected to pass it already; report if it does not.
Do: also replace the hand-written `sameBools` with `slices.Equal` and delete `sameBools`. No production code changes.
Check: `go test -tags fts5 ./core/database/likenesses/` passes; scripts/check-pr.sh PASS.

## Round 1 — fix
Head `79beb466`. Gate PASS (orchestrator run). F-1 done with no production change: the deny case passed on its first run, as the fix plan expected (`go test -count=1 -tags fts5 -run TestForEntitiesPreferred ./core/database/likenesses/` PASS, 6 cases). The test asserts gallery [A], which only holds if B is denied. `sameBools` replaced by `slices.Equal`.

## Round 2 — review
Scope `b544f736...79beb466`, lenses CON and MNT (the two with Fix-now findings).

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| CON-2 | CON | major | (round 1) deny case untested | Resolved | Reviewer mutation-checked with `go test -overlay`: the case fails when the negative misses B's image, when it is not stronger, when it is positive, and when the `InGallery()` filter is disabled (only this case fails then) |
| MNT-3 | MNT | minor | (round 1) `sameBools` | Resolved | `rg sameBools core/` empty |

No new findings. Both verdicts Clean.

## Result
Gate PASS · CI see PR · open blockers/majors: none (CON-1 is the §10 manual check, the user's) · follow-ups: MNT-1 (M-4), MNT-2, MNT-4 (M-3), MNT-5, UX-1, UX-2, UX-3, PRF-1 · to user: CON-1 manual check; MNT-1 pull-forward option

## User fix (after round 2)
The user, 2026-10-17, on MNT-1: "Pull M-4 forward into this PR: add database.NullBlob and use it in SetPreferredLikeness." This overrides the plan's Codebase fit row for `nullBlob`. It covers only the new copy; moving the existing `artifacts`, `sources` and `observations` copies and `uuidJSON` stays in M-4. No third review round: the gate covers it.

**F-2 — One shared `database.NullBlob` instead of a fifth private copy** (from MNT-1)
Files: `core/database/nullblob.go` (new, beside `userid.go`), `core/database/canonicalentities/canonicalentities.go`, and a test in `core/database` (the package's existing test style).
Failing test first: `TestNullBlob` (table): nil → nil, `[]byte{}` → nil, a 16-byte id → the same bytes.
Do: add `func NullBlob(b []byte) any` with a doc comment in `userid.go`'s style ("binds an empty BLOB as SQL NULL"). In `SetPreferredLikeness`, call `database.NullBlob(observationID)` and delete the private `nullBlob`. Do not touch the other packages' copies. Update the plan's Landed note (or the PR's deviation line) to say `canonicalentities` uses `database.NullBlob` (user, 2026-10-17).
Check: `go test -tags fts5 ./core/database/ ./core/database/canonicalentities/` passes; scripts/check-pr.sh PASS.

## User fix — result
Head `98cbce64`. Gate PASS (orchestrator run). F-2 done: `TestNullBlob` red first (`undefined: NullBlob`), then green. MNT-1 → Fixed for this PR's copy; the remaining copies and `uuidJSON` stay in M-4.

## User finding (manual check, 2026-10-17)
On the test project, a Person with likenesses and no chosen portrait leads its header with the earliest crop, but that tile still shows "Use as portrait". Clicking it changes nothing in the header and only moves the Badge. This is AC-1.1 as written (no Badge until a portrait is chosen); the user finds it misleading. User's call, chosen from two options: **the tile that leads the header always shows the Badge**. That is the earliest when nothing is chosen, the chosen one otherwise. The header menu appears only while the lead is not the earliest. This replaces AC-1.1, AC-2.3 and §10 step 2 (user, 2026-10-17).

**F-3 — The tile that leads the header always shows the Portrait Badge** (user)
Files: `core/database/likenesses/likenesses.go` (+ test), `api/proto/engine.proto` + regenerated Go/Swift (`./scripts/generate-proto.sh`), `api/ffi/handlers/conclusion_detail.go` (+ test), `macos/App/Platform/GenealogyStore.swift` (`CatalogLikeness`), `macos/App/Platform/GoStore.swift`, `macos/App/Features/Conclusions/PersonDetailContent.swift`, `macos/App/Features/Conclusions/ConclusionDetailPage.swift`, `macos/ProvenenciaTests/PersonDetailTests.swift`, the plan.
Failing tests first:
- `TestForEntitiesPreferred`: every case also asserts a `Leads` flag per likeness. With no pointer it is `[true, false]`. With pointer B it is `[false, true]`. Unpromoted, denied and deleted all fall back to `[true]`. An empty gallery has no flags.
- `portraitFlagsFollowThePreferredLikeness` (Swift): with neither preferred, `content.likenesses.map(\.isPortrait)` is `[true, false]` and the header menu is not shown. With B preferred it is `[false, true]` and the menu is shown. With A (the earliest) preferred it is `[true, false]` and the menu is not shown.
Do:
- Go: add `Leads bool` to `likenesses.Likeness`, documented as "the likeness the handle leads with: the preferred one when it is in the gallery, else the earliest; exactly one in a non-empty gallery". In `ForEntities`, after the sort and the `Preferred` marking, set `Leads` on the preferred index, or on index 0 when nothing is preferred. `Lead` returns the likeness with `Leads` set (the one rule). Keep `Preferred`.
- Proto: add `bool leads = 4` to the likeness detail message beside `preferred`, with a comment, and set it in the handler.
- Swift: add `CatalogLikeness.leads` and map it in `GoStore`. The tile's `isPortrait` becomes `$0.leads`.
- Header menu: replace the header's "any likeness is portrait" flag with `likenesses.first.map { !$0.leads } ?? false`. The gallery is earliest first, so the menu shows only while a chosen portrait is not the earliest. Name the field for what it means, e.g. `ConclusionDetailThumbnail.offersEarliest`, and update `showsPortraitMenu`.
- Do not add a new string.
- Plan: rewrite AC-1.1 ("the earliest tile shows the Portrait Badge and no button; the other tiles show the button; the header thumb is A"), AC-2.3 ("No portrait chosen, or the earliest chosen: no menu opens") and §10 step 2 to match, each marked "(user, 2026-10-17)". Add a line to the Landed note.
Check: the Go tests in `likenesses`, `conclusionheaders`, `conclusiondetails` and `api/ffi/handlers` pass; `xcodebuild` tests pass; scripts/check-pr.sh PASS.

## User finding — result
Head `e402e1fe`. Gate PASS (orchestrator run). F-3 done; tests were red first: `TestForEntitiesPreferred` (Leads flags), handler `TestSetPreferredLikeness`, `portraitFlagsFollowThePreferredLikeness`. `likenesses.Lead` now returns the one likeness with `Leads` set; the tile Badge follows `leads`; the header menu shows only while the first (earliest) tile does not lead (`offersEarliest`). The plan's AC-1.1, AC-2.2, AC-2.3, AC-3.2, T-6 and §10 steps 2, 4 and 5 were updated (user, 2026-10-17). D-2, D-11 and PDP-10 still describe the earlier model and were left as written.
