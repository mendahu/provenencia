# Review — S12-11 Source page Graphs section

Covers the plan and its brief, [S12-D2](../design/S12-D2-source-page-graphs.md).

**Verdict:** round 1 — Fits with changes / Not ready. Every finding Fixed; round 2 pending.
**Rounds:** 1. **Reviewers:** maintainability (opus), executability (sonnet)

## To user

None. The four product calls (Default verbatim, breadcrumb shape, Delete always enabled, placement above Artifacts) were answered by the user on 2026-10-25 before review.

## Writer's own calls

- The per-graph tooltip is new copy, "Add an artifact to open this evidence graph" (D-4), not the Sources list's "Add an artifact on the source page…".
- A toast after a graph delete, with Artifact delete's verb "erased" (D-13).
- No retry button on the section's error callout (D-8).
- The graph canvas and Promote places warm the graph list so a restored breadcrumb has its name (D-10).
- Breadcrumbs read the Source title from the Sources list as well as the graph name from the graph list (D-9, from M-6).
- `WorkspaceLocation.sourcePage` replaces three hand-built Source page locations in this PR (D-16, from M-5).
- A Promote button per graph is Out (not planned), though S12-09's notes pointed it at S12-11.
- The section count is localized like Artifacts ("3 evidence graphs"), not a bare number like Notes (D-15); the drift goes to the dogfood log.

## Round 1

### M-1 · blocker · The new graph reads bypass the session's graph-list owner
`WorkspaceSession+EvidenceGraphs.swift` is "the one place" graph-list reads live; the plan inlined a by-Source filter in the section and a by-id lookup in the toolbar. S12-13 will need an id index.
**Answer:** Fixed — D-2 adds `evidenceGraph(id:)` and `evidenceGraphs(sourceID:)`; `linkedArtifacts` uses the first; the section and toolbar read them (STEP-1, T-2).

### M-2 · blocker · The per-graph blocked reason is in the page section; S12-14 needs it on the list
**Answer:** Fixed — D-4 puts `graphBlockedReason(for:)` in `SourcesListNavigation` beside `graphLocation(for:graph:)`; key `sources.graph.needsArtifact`; the test moved into T-1.

### M-3 · major · STEP-4 leaves `graphsBySource(in:)` and `graphListLoaded(in:)` dead
Confirmed: single callers at `SourcePageModel.swift:70` and `:78`.
**Answer:** Fixed — D-11 and the STEP-4 `rg` check include both; the `graphBlockedReason` doc comment is reworded.

### M-4 · major · The S12-13 scope entry names `evidenceGraphLocation`, which this PR deletes
**Answer:** Fixed — STEP-7 rewrites that sentence onto the D-2 accessors and `SourcesModel.graphsBySourceCache`.

### M-5 · major · A fourth hand-built Source page location
Confirmed: `SourcesListNavigation.pageLocation(for:)`, `EvidenceGraphModel.sourcePageLocation(title:ref:)`, `CitationComposerModel.sourcePageLocation()`.
**Answer:** Fixed — D-16 adds `WorkspaceLocation.sourcePage(sourceId:ref:title:)` and moves the three onto it (STEP-5; `EvidenceGraphModelTests` added to the step's check; a stop condition if a field would change).

### M-6 · major · Graph name from the list, Source title from the location
**Answer:** Fixed — D-9's `BreadcrumbNames` resolves both from the lists (all three graph places already warm `.sourcesList`), with the location fields as fallback. AC-3.4 now covers a rename; AC-3.5 covers the cold case.

### M-7 · minor · `confirmResourceDelete` treats any non-source kind as an Artifact
**Answer:** Fixed — D-5 makes the closure a `switch` with explicit `source`, `artifact`, `evidence_graph` cases and a no-write `default`; T-5 asserts `deleteArtifactCalls == 0`.

### M-8 · minor · Section count mechanism
**Answer:** Fixed — D-15 uses `L10n.Sources.graphsCount(_:)` like Artifacts; the Notes drift is a dogfood-log line in STEP-7 and an Out row.

### M-9 · minor · (S12-D2) GS-1 claims Artifacts and Notes count the same way
**Answer:** Fixed — GS-1 and the inventory row name Artifacts as the model, with examples in words.

### M-10 · minor · (S12-D2) "deleted" vs the Artifact toast's "erased"
**Answer:** Fixed — D-13, GS-6 and §2 of the brief use "was erased from this source."

### M-11 · minor · (S12-D2) The `evidence_graph` noun is new copy
**Answer:** Fixed — the inventory row says the noun is new; D-12 adds it.

### X-1 · blocker · Plan says Blocked yet tells the implementer to build from words
**Answer:** Fixed — the Board line says not to start until it holds the agreed link and Status is Planned; the planner refreshes the plan against the frames at agreement. The "board wins" sentence and the board stop condition are gone; §6 Frames is the build description the planner confirms.

### X-2 · major · Delete-impact row mislocated; confirm title false
Confirmed: `confirmTitle(noun:ref:)` is "Delete %1$@ %2$@?".
**Answer:** Fixed — the Current state row cites `core/database/deleteimpact/register.go`; AC-2.1, D-12, brief §2 and frame 4 read "Delete evidence graph GPH-…?".

### X-3 · major · Tests read the graph list without waiting for it
Confirmed: `warmFromSession` does not wait on the graph list; `settle(_:)` exists in `SourcePageModelTests`.
**Answer:** Fixed — D-14 and T-3 – T-5 settle `evidenceGraphsHandle`.

### X-4 · major · Row layout and spacing left open
**Answer:** Fixed — D-17 fixes padding, spacing tokens, the Spacer, button grouping, ref under the name, row and section spacing and narrow-width behavior; D-8 fixes the skeleton sizes; D-15 the count. The brief's GS-2 now places the ref under the name.

### X-5 · major · Helpers left dead by STEP-4
Same as M-3. **Answer:** Fixed — see M-3.

### X-6 · major · Where the place key goes; T-9 duplicates the table edit
**Answer:** Fixed — D-10 appends the key last, as the composer does; T-9 is the edit of the two `queryKeysMatchDesignTable` cases, not a new test.

### X-7 · minor · Breadcrumb ref, ids and cold render
**Answer:** Fixed — D-9: Source crumb id `source-<sourceId>`, `ref: nil`; graph crumb id `graph-<graphId>` on every surface; the cold render shows the location's Source title and "Evidence graph" until the lists land (AC-3.5).

### X-8 · minor · `confirmResourceDelete` shape and the pbxproj edit
**Answer:** Fixed — D-5 (switch, no `refreshCoverFromStore`), and §7's preamble names the four `project.pbxproj` entries to copy.

### X-9 · minor · Checks that cannot run
**Answer:** Fixed — STEP-3's check is a build plus running the app per README "Run the Mac app"; the xcodebuild commands match `scripts/check-pr.sh`, defined once as `XCT`.

### X-10 · minor · (S12-D2) Brief gaps
**Answer:** Fixed — ref position (GS-2), error callout keeps the header and its copy (GS-8), narrow window (GS-10, frame 9), confirm title (frame 4). The handoff is unchanged: the plan's Status line, not its Planned date, is what moves. The New-row lint warning stays: the row has no second call site by design and is a snowflake, not a kit primitive.

### X-11 · nit · Lint warnings and apostrophes
**Answer:** Fixed — STEP-7 says the "Copy…" warning is expected; AC-2.3 and the brief use the typographic apostrophe the L10n string has.

## Round 2

Executability (sonnet): every round-1 answer accepted; verdict **Ready with changes**.

### X-12 · blocker · STEP-4 `rg` check matches kept code
The bare `graphListLoaded` matches the kept `graphBlockedReason(hasArtifact:graphListLoaded:location:)` parameter, `SourcesModel.swift` and its tests.
**Answer:** Fixed — the pattern is `graphListLoaded\(in`.

### X-13 · minor · `sourcePage` parameter types; which title the graph crumb carries
**Answer:** Fixed — D-16 gives `sourcePage(sourceId: String, ref: String? = nil, title: String? = nil)`; D-9's graph crumb carries the Source crumb's resolved label.

### X-14 · nit · Header jump row used unlabelled `PVButton` arguments
**Answer:** Fixed.

### X-15 · nit · How to pick pbxproj ids
**Answer:** Fixed — new ids count up from the highest `AA…` id.
