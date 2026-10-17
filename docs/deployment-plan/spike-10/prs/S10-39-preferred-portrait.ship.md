# Ship log — S10-39 Preferred portrait

**Plan:** [`S10-39-preferred-portrait.md`](S10-39-preferred-portrait.md) · **Branch:** `s10-39-preferred-portrait` · **PR:** #450
**State:** fixing 1

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
