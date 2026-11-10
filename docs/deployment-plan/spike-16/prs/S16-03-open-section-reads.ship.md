# Ship log — S16-03 One open section and one list of section reads

**Plan:** [`S16-03-open-section-reads.md`](S16-03-open-section-reads.md) · **Branch:** `s16-03-open-section-reads` · **PR:** #769
**State:** ready

Base: `main` @ `16efd08f`.

## Round 0 — implement
Head `b77f5b68` (PR #769, draft). Gate PASS (orchestrator run). Report deviations: host init takes `accessibilityIdentifierRoot:` per D-7/D-8 (the §6 sketch said `pageIdentifier:`); four code changes grouped in one commit `c1c9a8dc` because they compile only together; the six code/doc commits before `b77f5b68` lack the Co-Authored-By trailer (already pushed, not amended); AC-3.2/AC-3.3 are manual checks, not run.

## Round 1 — review
Reviewers: CON, MNT, UX, SEC+PRF. UX: diff touches `macos/App/**` (section hosts, accessibility identifier roots). SEC+PRF: performance paths — `macos/App/Features/**` models, `PlaceRegistry` warm keys and `WorkspaceSession` cache eviction; security has no path of its own here but shares the reviewer.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| MNT-1 | MNT | minor | `ConclusionDetailHeaderBand` still names the open section `pageSection` | Fix now → F-1 | verified: `ConclusionDetailPage.swift:299` passes `pageSection: openPageSection`; `:337`, `:342` (`pinned` reads `pageSection != nil` as "has sections"). A minor in a file this PR changes; D-6's own rationale (one word per concept) covers it. Supersedes STEP-5's "keeps its name" Out note for this one private struct with one caller; recorded in the plan's Landed note. No behavior change |
| CON-1 | CON | minor | No test tells `openPageSection` from `pageSection` in the place or the session | Fix now → F-2 | verified: Vitals reads `[]`, so T-2's nil case and T-4's nil case pass under either resolution. No seam gives a first section a read today (D-12). Fix is a doc comment on T-2 and T-4 naming the gap and that the first PR giving the first section a read adds the failing nil case |

Nits (logged): `MNT-2 · macos/ProvenenciaTests/PersonDetailTests.swift:569 · read(role) loop mirrors its implementation; assert direct values` · `MNT-3 · macos/App/Features/Workspace/ConclusionPageSection.swift:348 · inline comments on Sources reads repeat Role docs` · `MNT-4 · macos/App/Features/Conclusions/PersonSourcesSection.swift:249 · wrap the long PersonSourcesRowView call` · `CON-2 · S16-03 plan manual check list numbers "6." between 3 and 4`

SEC+PRF: no findings. UX: no findings.

### Fix plan 1

**F-1 — Name the header band's open section `openPageSection`** (from MNT-1)
Files: `macos/App/Features/Conclusions/ConclusionDetailPage.swift`; the plan's Landed note (`S16-03-open-section-reads.md`).
Failing test first: none — a rename with no behavior; the build is the check.
Do: rename `ConclusionDetailHeaderBand`'s `pageSection` property to `openPageSection` (declaration, `pinned`, the `if let` in `body`, and the call site's label). Add one sentence to the plan's Landed note: the band's parameter was renamed too, from review MNT-1, superseding STEP-5's "keeps its name".
Check: `rg -n 'pageSection' macos/App/Features/Conclusions/ConclusionDetailPage.swift` prints only `openPageSection` / `offeredPageSections` lines; scripts/check-pr.sh PASS.

**F-2 — Say what T-2 and T-4 cannot see while the first section reads nothing** (from CON-1)
Files: `macos/ProvenenciaTests/PlaceRegistryTests.swift` (`personDetailWarmsEveryReadOfTheOpenSection`), `macos/ProvenenciaTests/WorkspaceSessionTests.swift` (`applyPersonPageKeepsTheOpenSectionsReadsVisible`).
Failing test first: none — comments only.
Do: add a `///` doc comment on each test: while the first section (Vitals) reads nothing, the stored-nil case passes whether the place resolves `openPageSection` or the raw `pageSection`; the first PR that gives the first section a read must add that case's own key to this test's expectation, which then fails against the raw `pageSection`.
Check: scripts/check-pr.sh PASS.

## Round 1 — fix
Head `88744c49`. F-1 done (`55b241a8`): band parameter renamed, Landed note updated. F-2 done (`88744c49`): doc comments on T-2 and T-4. Read the fix commits: MNT-1 **Resolved**, CON-1 **Resolved**. No round 2 (no Fix-now blocker or major in round 1).

Gate FAIL ×2 (orchestrator), xcode-test: `WorkspaceDestinationHostTests` (`canvasToCanvasBuildsTheSecondGraphsModel`, `openGraphThatGoesFallsBackToTheSourcesList`, `goneGraphFallsBackToTheSourcesList`) — `waitUntil(timeout: 15s)` expectations fail. Reproduced in 4 of 4 full-suite runs at `88744c49`; the suite alone passes 3 of 3; `main` @ `16efd08f` full suite passes (1869 tests, 20s). The implementer's own first fix-round gate failed on other layout tests (`SourcePageOverviewLayoutTests`, `PVWidthClassTests`, `PVPageColumnsTests`) and passed on a rerun. Round 0's gate passed. Sent back as gate fix G-1.

**G-1 — Full-suite xcode-test fails in WorkspaceDestinationHostTests**
Find which of this PR's tests or changes interferes (likely shared state, a leaked task, or a session/navigation left alive by a new test; T-4 `applyPersonPageKeepsTheOpenSectionsReadsVisible` in `WorkspaceSessionTests` is a first suspect), and fix the cause. Do not loosen timeouts or disable tests.

G-1 **Not reproduced** by the implementer: full suite 5 of 5 PASS at `88744c49` (17–29s, one run with the CPU oversubscribed); no commit. Orchestrator reran `scripts/check-pr.sh` with no other xcodebuild running: PASS (xcode-test 72s). The new tests are synchronous and touch no shared state; the diff does not touch the graph canvas or the Sources-list fallback. Read as a load-sensitive flake in `WorkspaceDestinationHostTests`' 15s `waitUntil` window tests, not a regression from this PR. Logged as follow-up FU-1.

## Result
Gate PASS (`88744c49`) · CI green (run 42762770432) · open: none · follow-ups: FU-1 (`WorkspaceDestinationHostTests` graph-canvas waits time out under load in the full suite; failed 4 of 6 orchestrator full-suite runs at `88744c49`, 0 of 5 implementer runs and 0 of 3 alone) · to user: —
