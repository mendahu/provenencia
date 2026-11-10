# Review — S16-03 One open section and one list of section reads

**Verdict:** pending round 2
**Rounds:** 1 of 2 so far. **Reviewers:** maintainability (Opus), executability (Sonnet)

## To user
None.

## Writer's own calls
- D-12: no version bump. The PR changes nothing a researcher sees; the eviction P-13 describes cannot happen while Vitals, which reads nothing, is first.
- D-6: the context field and the page's private property are renamed `openPageSection`; `ConclusionDetailHeaderBand`'s `pageSection` parameter keeps its name.
- M-1 answered by building every section identifier from one root rather than by dropping the page-identifier threading, so the working rule's "identifier root comes from the page identifier" holds and the scope grows by about 18 one-line string edits.
- M-9 accepted: `ConclusionPageSection` moves to its own file in STEP-2, which the map's scope entry did not name.

## Round 1

### M-1 · blocker · Section identifiers would be built two ways in the same views
**Answer:** Fixed. D-8, D-9: each section computes its root once with `identifierRoot(pageIdentifier:inColumn:)` (renamed from `hostIdentifierRoot`) and builds every child identifier from it; `PersonRecordLinksColumn` takes `identifierRoot`. The Out row is gone; In row, §4 "Section identifiers", STEP-6, AC-3.4 and manual check 6 added.

### M-2 · blocker · `showsRefreshing` and the place's `map(\.key)` will silently misfile S16-29's `prefetch`
**Answer:** Fixed. D-4: `Role.showsRefreshing` and `Role.warmsWithPlace` are exhaustive switches; the place filters on `warmsWithPlace`. AC-1.2, AC-1.4, sketch, T-1, T-2 and Next adds updated.

### M-3 · major · The gallery model moves into a view file, and D-10's reason is false
**Answer:** Fixed. D-10: the gallery models go to `ConclusionDetailGallery.swift` (new); the reason now cites the synchronized `App` group. §4 and STEP-1 updated. This departs from the audit's text, which named `ConclusionGallerySection.swift`.

### M-4 · major · "Read only through `openPageSection`" is false for an existing reader
**Answer:** Fixed. §4 row "Crumbs" (`WorkspaceLocationLabels.labels(for:)` stays on the stored value); STEP-7's skill text gives each property its job.

### M-5 · major · Stale doc comments the sweep doesn't reach
**Answer:** Fixed. §4 row "Section doc comments"; STEP-6 rewrites them; its check adds `ConclusionPageSection\.queryKey`; the DoD sweep covers `macos/App`.

### M-6 · minor · `pageSection` would mean two things
**Answer:** Fixed. D-6, STEP-5: the context field and the page's private property are renamed `openPageSection`; added to the sweep.

### M-7 · minor · "An Event or Place section is a case in `offered(by:)` plus its entries in `reads`" is not true
**Answer:** Fixed. Next adds and STEP-7: a kind also gives its place spec the open-section line and its own fill.

### M-8 · minor · The supporting key is still written down twice
**Answer:** Fixed. D-11 now reads the Source types through `read(.supporting, …)`.

### M-9 · minor · The cache-read type lands in the navigation file
**Answer:** Fixed. D-1, STEP-2: `ConclusionPageSection` and `ConclusionSectionRead` move to `macos/App/Features/Workspace/ConclusionPageSection.swift` (new).

### M-10 · nit · `inColumn: Bool` sits beside the `Placement` enum
**Answer:** Logged as an Out row (not planned).

### X-1 · major · A `let project` becomes unused in STEP-6
**Answer:** Fixed. STEP-6 deletes it.

### X-2 · major · T-4 has no fixture or setup; the status is Draft
**Answer:** Fixed for T-4 (setup as `applyTriggerEvictsHiddenDetailPagesAndRevalidatesTheVisibleOne`, `makeSession(store: FakeStore())`). Declined for the status: it stays Draft until this review closes, per `plan-pr` step 8.

### X-3 · minor · Checks that need macOS
**Answer:** Fixed in part. §10 opens with "On macOS with Xcode". Declined for the Implementer block, which `plan-pr` requires word for word.

### X-4 · minor · STEP-1 placement is under-specified
**Answer:** Fixed. STEP-1: each type keeps its doc comment and order; the gallery models start a new file.

### X-5 · minor · Small choices the plan leaves open
**Answer:** No change. The reviewer found each one settled.

### X-6 · nit · `PersonRecordLinksColumn`'s hard-coded root
**Answer:** Fixed with M-1 (D-9).
