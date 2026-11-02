# Spike 14 — Completed

One short paragraph per landed PR, newest first. Written by the PR that lands it.

## S14-14 — Copy and VoiceOver on graph work

**Landed 2026-11-01** in [#675](https://github.com/mendahu/provenencia/pull/675). Return cancels a danger confirm (`PVConfirmTone.returnConfirms`), and a dirty graph form or Artifact card asks "Discard changes?" before a Kept-linked link or another card takes it, through one `PVDiscardGate` the sheet chrome (`pvSheetClose`), the Artifact cards and the composer share. Toasts, error callouts and error lines (`announces: true`) speak to VoiceOver through `pvAnnounce`, section titles are headings, and `PVField` gives its label to the `PVInput`, `PVTextArea` or `PVSelect` inside it. `WorkspaceLocationLabels` names the open Evidence graph in the crumbs, the Back / Forward rows and the canvas's VoiceOver label. Four counts and the 18 delete-impact kind nouns are plural keys, five strings name the Evidence graph, and `check-localizable-xcstrings.py` fails on an uninflected `%lld <word>`. A failed Citation load names the Citation under the picker, a file macOS fails to open says so (`ObjectRefusal.notOpened`), and the Source type editor returns focus to the Type pencil. Version 0.2.6.

## S14-12 — File work before the catalog session

**Landed 2026-11-01** in [#664](https://github.com/mendahu/provenencia/pull/664). Attaching a scan, saving a depiction, setting a Source cover and opening a Source page with missing thumbnails no longer copy, hash, decode or encode while holding the catalog session. One `ingest.Staged` type (`StageFile`, `StageImage`, `StageBytes`) stages into a temp in `objects/` outside `catalogsession.Do`; `withStaged` / `withStagedDepictions` stage first, and the session only looks up, inserts and installs (`Settle` purges what a write did not use). `derivatives.Plan` / `Job.Stage` / `Commit` split a thumbnail, and `thumbnailPass` commits them in batches of 16; `sources.NormalizeCover` is lookup-only. `catalogsessiontest.Hold` holds the session in tests. Version 0.2.5. Audit A-24.

## S14-11 — Reload only what moved, only when it is on screen

**Landed 2026-11-01** in [#657](https://github.com/mendahu/provenencia/pull/657). A graph, link or Artifact write reports a keyed `evidence_graphs` read for its Source (`effects.ReadEvidenceGraphs`, replacing the project-wide list), `ListEvidenceGraphs` takes an optional `source_id`, and the graph list re-reads only the named Sources and merges them (`CatalogQueryRefresh.merged`, `WorkspaceSession.refreshCached`; graph progress uses the same factory). Add and Copy Source insert the returned row through `.createdSource(source)` and `CatalogQueryPatch.sourceRow()`. `QueryHandle.isStale` replaces the session's `invalidatedKeys`: a stale key is not settled, so a returning list publishes its sidebar count after its reload. A visited graph canvas and Source page go stale while hidden and reload when shown. `SourcesModel.visibleSources` builds once per input change. Version 0.2.4. Audit A-9, A-10, A-25.

## S14-23 — Oversized JPEG thumbnails

**Landed 2026-10-31** in [#654](https://github.com/mendahu/provenencia/pull/654). A JPEG over 64 megapixels gets its thumbnail: `raster.EncodeJPEG` decodes it at 1/8 scale from its DC coefficients with a pure-Go decoder (`core/raster/jpegdc.go`, baseline and progressive Huffman, fuzzed), then scales and encodes as before, so the tile, Source cover and omnibar show it. `core/raster` also refuses input over `MaxDecodeBytes` (128 MiB, which `derivatives` now uses) and every decode failure is `ErrUndecodable`, mapped by `derivatives` to the glyph instead of a failed Artifact response. `core/raster/rastertest` holds the shared over-budget fixtures; `go-tests.mdc` has a "Fuzz targets" paragraph.

## S14-09 — Quote nicknames and add former surnames

**Landed 2026-10-31** in [#649](https://github.com/mendahu/provenencia/pull/649). A nickname part shows in curly quotes and the new `former_surname` product part type (role `family`, after the nickname in both Western orders, seeded in new catalogs only) shows in parentheses, in every name display. `NameValueDisplay.display` returns the text and, when a word is emphasized, a styled copy; the Persons list row and Person page title draw the former surname italic through `ConclusionTitleSource.name(_, styled:)` and `ConclusionListRow.italicizingEmphasis`. The two titles also carry a spoken copy (`NameDisplay.spoken`: "Jake, nickname", "former surname Smith") for VoiceOver, and the Why heading of a Name value has no outer quotes. Go's `nameformats.Format` sort key stays undecorated. `NamePartType` spells every raw value and has a "name part types" row in `TestSwiftVocabularyMatchesGo`; `add-name-value` step 6 lists every place a part type lives.

## S14-10 — A Citation belongs to one Evidence graph

**Landed 2026-10-30** in [#645](https://github.com/mendahu/provenencia/pull/645). `citations.graph_id` (migration `000059.sql`, NOT NULL, an audited CASCADE facet of its graph) stores the graph a Citation was made in, required on create and linked to its Artifact, and never changes. A graph lists and counts only its own Citations, an Observation must name a Subject in its Citation's graph, and the Citation `citations` read is its own `graph_id`. An unlink is refused while the graph has any Citation on the Artifact; `citations_need_graph` is gone. A graph delete releases its Citations through `ReleaseCitation` (the removal `citations.Delete` shares) and its confirm names them. The composer passes its graph (`CreateCitationWithObservationsRequest.graph_id`).

## S14-02 — Pictures in the app

**Landed 2026-10-30** in [#641](https://github.com/mendahu/provenencia/pull/641). Event and Place pages show a Pictures section, lead their header with the engine's picture and choose a thumbnail as a Person chooses a portrait; the Events and Places lists lead with it. `ConclusionDetailBody.gallery` is one slot for all three kinds and `ConclusionLeadModel`, owned by the page, chooses any kind's lead; the Person's copy is unchanged and its ids and caption wrapping follow the shared rule. The Properties list reads the key and bound types, so `event_picture` and `place_picture` sound different. Briefs S14-D2 – S14-D5 are archived.

## S14-03 — Rectangle-only depiction drawing

**Landed 2026-10-31** in [#652](https://github.com/mendahu/provenencia/pull/652). A depiction is drawn with the rectangle tool alone (`CitationComposerModel.viewerRegionTools`), a Citation's region keeps all seven, and "Use the citation's region" sets the region's bounding rectangle (`ArtifactRegionDraft.boundingRectangle`). While any region is drawn, Set page (now a `PVChip`), the locator chip ×, the Clear menu and Find are disabled and the locator hint hides (`isDrawingRegion`). Draw and the region buttons wait for a file `ArtifactRegionCrop.snapshot` can cut; a failed cut shows in the caption (announced, pointing at the file) with Save disabled and retries on the next set, driven by `depictionPendingCutKey`; the needs-page Callout saves the viewer's page to the Citation, shows a failed save beneath the button and returns focus to Draw. Captions name the region buttons, an unchanged edit reads "Saved image", and VoiceOver hears the freehand cap. Brief S14-D6 is archived; product version is 0.2.3.

## S14-13 — A crafted project can't crash the app

**Landed 2026-10-29** in [#632](https://github.com/mendahu/provenencia/pull/632). `subjectpositions.MaxCell` (±1,000,000) is refused on write and clamped on every read, and Swift placement and moves use saturating cells (`EvidenceGraphPlacement.cellLimit`, held to Go's by `TestSwiftVocabularyMatchesGo`). The Artifact viewer decodes once through `ArtifactImageDecode`: the sniffed type must conform to one of the type identifiers Go's stored-type table sends with the file (`SourceFileRef.type_identifiers`), a declaration over 256 MP is refused with its own "Image too large to show" state, both budgets count bytes (divided by the bytes per color component), and the decode is held to `raster.MaxDecodePixels` (64 MP) for display, crop and OCR, with `displayImage` keeping the original point size; `ImageHeaderDecode` now owns the shared header read and decode. `ffi.Call` recovers a panic into `internal.unknown` (the switch is `route`), and a panicking `catalogsession.Do` callback drops its session so the next call reopens, and a researcher catalog that panics while opening is closed. The PDF cut reopens its file through `ProjectFiles.objectURL`. Version 0.2.2. Audit A-26 to A-28.

## S14-01 — Pictures engine

**Landed 2026-10-28** in [#615](https://github.com/mendahu/provenencia/pull/615). New catalogs seed locked `event_picture` / `place_picture` (Label Picture). Package `core/database/gallery` admits `GalleryKeys()`, reads each tile from `depiction_values.image_file_id` (SEC-4), and `leadPictures` fills `PictureImageFileID` on Person, Event and Place headers. Proto/Swift wire rename to `PictureDetail` / `pictures` / `pictureImageRelPath`; preferred column and `SetPreferredLikeness` stay. View-layer gallery chrome remains S14-02.

## S14-08 — Other-link support calibration

**Landed 2026-10-28** in [#613](https://github.com/mendahu/provenencia/pull/613). `segment.OtherSupport` is 0.75 (`OtherCredit` stays 0.25), so a Source person with only a short given name plus a lined-up grandparent (or other non-birth/death link) clears Weak on Promote for the Jake + Carol dogfood shape; inventory and `promote-graph-alignment.md` quote the new total, `TestPartialNameWithGrandparentClearsWeak` locks the band, and product version is 0.2.1.

## S14-24 — Workspace width class

**Landed 2026-10-29** in [#628](https://github.com/mendahu/provenencia/pull/628). `PVSpacing.widthWideBreakpoint` (1600) and `PVWidthClass` with the `pvWidthClass` environment value; `WorkspaceDestinationHost` fills the detail area and measures the pane once, outside the place's `.id`, so any page reads `.regular` or `.wide` without measuring itself. No page reads it yet and nothing on screen changes; a shared `hostInBorderlessWindow` test helper replaces three hand-written window mounts.

## S14-D1 — Wide layout in the design system (board)

**Agreed 2026-10-29** in the Provenencia [design-system project](https://claude.ai/design/p/0f6c1f29-8408-4735-b114-5ae4c4c18445); brief archived in [`design/archive/`](design/archive/S14-D1-wide-layout.md). Per the brief, the kit gains `--width-wide-breakpoint` (1600) and `--width-secondary-max` (480), a Width classes README section with the six Wide rules, and the Page columns component: a centered primary column at Regular, primary plus a 340–480 secondary column at Wide, the group centered past both maximums, an optional header slot that scrolls or pins. No PR of its own: Spike 13's S13-10 builds `PVPageColumns` from it with the Person page frame (S14-25 descoped).

## S14-15 — Xcode synchronized folders

**Landed 2026-11-01** in [#668](https://github.com/mendahu/provenencia/pull/668). `App/App`, `App/Platform`, `App/Features`, `App/DesignSystem` and `ProvenenciaTests` are Xcode synchronized folders (`objectVersion = 77`, Xcode 16 or later): a Swift file saved there is built with no `project.pbxproj` edit, and the 425 hand-written Swift entries are gone (audit A-34). `App/Resources` stays a classic group so `Fonts/` keeps its bundle subfolder; membership exceptions keep the three `.md` files out of the app, and `check-guidance-names.py` fails on a non-Swift file that is not an exception or a new folder under `macos/App` the project does not build. The compiled files and bundled resources are identical to before.

## S14-18 — Name folds never crash

**Landed 2026-11-01** in [#677](https://github.com/mendahu/provenencia/pull/677). The reconciler resolves each name value through the fold graph to the best supported, then best ranked, value that folds into nothing, so a person recorded as B., Bill and William (or Ann, Anna and Hannah) files, reopens and shows the formal name instead of writing to a nil map (audit 2 B-1). A fold cycle with no way out keeps each value. Which pairs fold is unchanged; `CacheVersion` is 22, so every catalog rebuilds once on open.
