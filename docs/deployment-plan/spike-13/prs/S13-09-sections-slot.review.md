# Review — S13-09 Sections slot on the Conclusion page

**Verdict:** Fits / Ready
**Rounds:** 2. **Reviewers:** maintainability (cursor-grok-4.6-xhigh), executability (composer-2.5)

## To user
None.

## Writer's own calls
- Slot API name `pageSections` (not `sections`) to keep Place hierarchy `ConclusionDetailBody.sections` distinct.
- Default fill via convenience `where` init (kit pattern `PVSelect` / `PVSelectPlainRow`), so Event/Place need no `pageSections:` argument this PR.
- `ConclusionDetailDefaultPageSections` is **internal** (not `private`) so the constrained init is visible to `*DetailView.swift`.
- New types stay in `ConclusionDetailPage.swift` (no pbxproj).

## Round 1
### M-1 · major · Default `pageSections` the way `portrait` is defaulted; keep the default fill private
**Answer:** Fixed — convenience init supplies the default; then superseded by M-4 (private cannot back the `where` constraint). Final shape: internal default + convenience init (D-3, D-5).

### M-2 · major · Rewrite comments that still teach a Person page fork and a Body sections slot
**Answer:** Fixed — D-7 rewrites the `ConclusionDetailBody` doc comment; D-6 / map Ships name only `ConclusionDetailPage` for the slot.

### M-3 · minor · Drop `Equatable` on `ConclusionDetailPageSectionsContext`
**Answer:** Fixed — D-1 is a plain struct.

### X-1 · major · `ConclusionDetailPageSectionsContext: Equatable` will not synthesize
**Answer:** Fixed — same as M-3.

### X-2 · major · STEP-1 check and commit 1 fight compile order
**Answer:** Fixed — STEP-1 is one compile unit (slot + default + convenience init + call-site inference); docs are STEP-2.

### X-3 · minor · STEP-1 omits page → Loaded plumbing
**Answer:** Fixed — STEP-1 names passing `pageSections` into `ConclusionDetailLoaded`.

### X-4 · nit · Checks not runnable in this dry-run environment
**Answer:** Logged — Xcode / `check-pr.sh` Swift legs run on macOS or CI.

## Round 2
### M-4 · blocker · Private default `PageSections` cannot back the convenience init the three detail views call
**Plan (after R1):** D-3 kept the default `private`; D-5 said do not edit call sites.
**Codebase:** A `private` type cannot appear in a `where PageSections == …` constraint used from `PersonDetailView` / `EventDetailView` / `PlaceDetailView`. Kit precedent: `PVSelectPlainRow` is internal. Two-arg `ConclusionDetailPage<A, B>(` specializations cannot name a third generic.
**Answer:** Fixed — D-3 makes `ConclusionDetailDefaultPageSections` internal; D-5 drops the two-arg specializations so inference picks the convenience init (or names all three types if inference fails); stop conditions forbid a second defaulting mechanism.

### X-5 · minor · §4 Map Ships row no longer matches `main`
**Answer:** Fixed — §4 Map Ships cell refreshed to the live Ships text; STEP-2 verifies AC-3.3 and owns status/`completed.md` at ship.
