# Deployment plan — Spike 14

Spike 14 is an odds-and-ends spike (working rule "Even spikes are odds and ends"). It takes small features, UX refinements and fixes in batches; each batch is a slice with its own Check. The first batch pulls the "other keys" item and the transparent-polygon item from [`ideas/depictions-and-likenesses.md`](../../ideas/depictions-and-likenesses.md), and adds thumbnail parity and rectangle-only depiction drawing. It also takes four Spike 10 likeness-audit leftovers that edit the same code: the gallery's image source (S14-01) and the composer's drawing feedback, captions and cap announcement (S14-03) (user, 2026-10-27, while planning Spike 12's audit fixes). Batch 1 is cut by layer (engine, app, drawing) because almost all of it reuses Person behavior that already exists (working rule "Batch 1 is cut by layer"). Batches 4 and 5 are what remained of Spike 12 when it closed (user, 2026-10-29): the last Evidence graph work, and the app-health audit fixes that were not descoped. The two audits moved here with them: [`app-health-audit.md`](app-health-audit.md), [`app-health-audit-2.md`](app-health-audit-2.md).

**A brief covers one view.** A PR that changes chrome waits on a brief for each view it touches. Briefs and PR plans are written just before the PR ([`add-design-brief`](../../../.agents/skills/add-design-brief/SKILL.md), [`plan-pr`](../../../.agents/skills/plan-pr/SKILL.md)). A bugfix or a behavior change with no new chrome is not gated.

## Status

**Open.** Per-PR status is the [PR map](#pr-map).

> **Goal of this spike:** a run of small, independent improvements, each batch landing on its own.

## Goal (dogfood bar)

**By spike close**, each of these is true on a real project. A batch adds its outcomes here when it is cut.

**Batch 5 — Audit fixes** (moved from Spike 12, 2026-10-29):

15. **A shared project is safe to open.** A crafted image or engine fault does not crash the app; the Artifact viewer decodes a crafted image within a bound; a PDF crop re-checks its file.
16. **One place for each job.** Every audited write returns its effects, pinned by a test; FakeStore re-implements no engine rule, search included; a new Swift file needs no hand-written project entry.
17. **Keyboard and VoiceOver are safe on graph work.** Return in a delete confirm cancels; a dirty graph form asks before a link inside it closes it; VoiceOver hears a graph save or its failure, can jump to the Evidence graphs section by heading, names every form field and names the Evidence graph on the canvas and in history; counts read "1 subject", not "1 subjects".
18. **Names fold to the formal name and never crash.** A person recorded as B., Bill and William files, reopens and shows William, and one recorded as Ann, Anna and Hannah files and reopens without a crash; Charlotte with Chick shows Charlotte, and John with Ivan shows John.
19. **The Promote page recovers, speaks, and files what it shows.** A Subject deleted in another window names its row and the page proposes again without losing other decisions; errors come into view when they appear and VoiceOver hears them; the footer's connection count is what Done files; a merge leaves no stale link in the graph cache.
20. **Search finds every spelling.** Search finds Søby from "Soby", and "Tom" finds a person shown as Thomas.

**Batch 4 — Evidence graphs finished** (moved from Spike 12, 2026-10-29):

13. **A Citation belongs to one Evidence graph.** A Citation made in the Jones graph shows only there; two graphs that need one reading each make their own Citation; deleting a graph deletes its Citations.
14. **Graph work stays fast on a split census.** A Citation save on a 60-graph census reloads only the graphs its Observations are on, and a hidden graph or Source page waits until it is shown; attaching a large scan does not stall the other panes.

**Batch 3 — Name display** (moved from Spike 11, 2026-10-29):

11. **A name shows its nickname and former surname.** A Person recorded as James with the nickname Jake reads *James “Jake” Robins*; one with the former surname Smith reads *Jones, Mary (Smith)* on the Persons list, with *Smith* italic there and on the Person page title.

**Batch 2 — Matching calibration:**

10. **Thin name plus kinship is Weak.** Promoting a Source person recorded only as a short given name who shares a lined-up grandparent (or other non-birth/death link) with a matching neighbor reaches at least the Weak band — not "Little evidence" for that shape. (Since 2026-10-29 the page shows no band; the calibration test checks this.)

**Batch 1 — Event and Place pictures** (new catalogs; working rule "New seeds reach new catalogs only"):

1. **A place has pictures.** On a Citation of a photo of St Mary's Church, the composer offers Picture for the Place Subject; the saved line reads "Picture of St Mary's Church".
2. **An event has pictures.** The same for a wedding photo filed on the Event.
3. **The pages show them.** The Place and Event pages show a Pictures section with the crop, captioned with its Source; a tile opens its record, as a likeness tile does on the Person page.
4. **A depiction is a rectangle.** Drawing a likeness or a picture shows the rectangle tool alone; a Citation's own region still offers all seven tools. "Use the citation's region" on a circled entry sets the depiction to the circle's bounding rectangle.
5. **An event or place has a thumbnail.** The Event and Place page headers lead with the earliest picture; choosing another tile as the thumbnail (or going back to the earliest) works as a Person's portrait does, and the Events and Places lists lead with the same picture.
6. ~~**Promote says picture.**~~ Dropped 2026-10-29: the evidence sheet it worded is removed by Spike 13's [S13-23](../spike-13/deployment-plan.md#s13-23--promote-stops-asking-for-pins).
7. **Likenesses are unchanged.** A Person's likeness, gallery and portrait work as before.
8. **A tile shows its own picture.** Every gallery tile, and the header or list thumbnail it leads, shows the image stored with its own record; an edited reconciler cache row cannot show another picture.
12. **Every JPEG gets a thumbnail.** An 11,000 × 18,000 scan, over the 64-megapixel decode budget, shows a thumbnail instead of the missing-image glyph and can be used as a cover, through Go code that a Windows build reuses. (Moved from Spike 12, 2026-10-29.)
9. **Drawing guides and reports.** While a region or a depiction is drawn, Set page, the locator chip × buttons and Find are disabled and the draw hint hides; Draw is blocked on a file the viewer cannot cut, and a failed cut shows with Save disabled; a PDF Citation's page can be set from the needs-page callout; the captions name the region buttons, not only Draw; an edit that saves no image does not say "Image to save"; VoiceOver hears the freehand tool's point cap.

**Scale** (Batches 4 and 5, carried from Spike 12): a census Source with 60 household graphs and 300 Citations on one page; a project of 3,000 Sources, 37,000 Subjects and 92,000 Observations, with 200,000 audit rows; 100 new people on one Source; 6,000 events on one place.

## Where the notes went

| Dogfood note or idea | PR | Brief |
| --- | --- | --- |
| Spike 12's open PRs when it closed (user, 2026-10-29): S12-52, S12-26, S12-27 (Batch 4); S12-28, S12-29 + S12-30 + S12-45, S12-31, S12-32, S12-33, S12-40, S12-43's stale-link fix, S12-44, S12-47 + S12-50, S12-48 (Batch 5); S12-39 (Batch 1). The rest were descoped ([Spike 12's note](../archive/spike-12/README.md)) | S14-10 – S14-23 | — |
| Spike 11's S11-29, planned and never built, moved here when Spike 11 closes (user, 2026-10-29) | S14-09 | — |
| Promote Jake + grandparent read "Little evidence" under Weak (user, 2026-10-28); raise `segment.OtherSupport` | S14-08 | — |
| [`depictions-and-likenesses.md`](../../ideas/depictions-and-likenesses.md) "Other keys of the same value type" (place and event) | S14-01, S14-02 | — |
| Thumbnail parity for Event and Place (user, 2026-10-27) | S14-01, S14-02 | — |
| Promote wording for a non-Person depiction (map review K-4) | Descoped (2026-10-29): the evidence sheet goes in Spike 13's S13-23 | — |
| Rename the gallery layer away from "likeness" (map review K-5) | S14-01, S14-02 | — |
| Rectangle-only depiction drawing (user, 2026-10-27) | S14-03 | — |
| [`depictions-and-likenesses.md`](../../ideas/depictions-and-likenesses.md) "Transparent polygon" | Refused (working rule "A depiction is drawn as a rectangle") | — |
| [`depictions-and-likenesses.md`](../../ideas/depictions-and-likenesses.md) "Gallery image from the record" (likeness audit SEC-4) | S14-01 | — |
| [`depictions-and-likenesses.md`](../../ideas/depictions-and-likenesses.md) composer feedback (UX-5, UX-6, UX-7), copy (UX-12, UX-13) and cap feedback, moved from Spike 12's audit planning (user, 2026-10-27) | S14-03 | — |

**Back in the ideas folder** — [`ideas/depictions-and-likenesses.md`](../../ideas/depictions-and-likenesses.md): signature and the Spike 10 audit leftovers not taken here.

## Working rules

These are the calls for this spike. PR plans treat them as binding. A rule names its batch.

- **Even spikes are odds and ends.** From Spike 13, odd-numbered spikes are large new features and even-numbered spikes collect small features, UX refinements and fixes. An odds-and-ends spike adds work in batches: each batch is a slice with its own Check and bar outcomes, cut with [`plan-spike`](../../../.agents/skills/plan-spike/SKILL.md) when it is pulled in.
- **(Batch 1) Batch 1 is cut by layer.** One engine PR, one app PR, one drawing PR. The app PR touches the Event and Place pages and the Events and Places lists (the Promote evidence sheet left its scope on 2026-10-29), more than one view, because each piece reuses a Person behavior with no new chrome and no engine path. If its plan runs past about 10 steps, the Promote wording splits out as its own PR. Decided 2026-10-27, replacing a seven-PR cut.
- **(Batch 1) Event and Place each get their own picture Property, labelled "Picture".** `event_picture` bound to `event` and `place_picture` bound to `place`, value type `depiction`, cardinality `multiple`, locked, each with its own description. The data model is right first and the UI adapts to it: one key per kind keeps each description true to its kind, though a Property can bind to several kinds (`start_date`). Cost accepted: a composer run that saves an Event's Picture and then picks a Place clears the Property, and the researcher picks Picture again (`applyObservationSubject`). `likeness` stays the Person key. No Property is named `depiction`. [Design](event-place-pictures.md#two-properties-one-label). Binds S14-01.
- **(Batch 1) The gallery layer is named gallery, not likeness.** "Gallery" for the layer and "picture" for one item in it, in Go, proto and Swift (for example `core/database/likenesses` → `gallery`, `LikenessTile`, `LikenessDetail`, `CatalogLikeness`, `ConclusionDetailLikeness`, `LikenessTileView`, `CachedThumbnail.likenessCrop`; Promote's `isLikeness` goes with the evidence sheet in Spike 13's S13-23). Proto field numbers are kept, so the wire stays compatible. "Likeness" stays only for the Person key, the Person page's copy and the portrait. The column `canonical_entities.preferred_likeness_observation_id` keeps its name: a column rename is a migration for a name only, and `catalogmodel` and the docs say what it holds. S14-01 renames Go, proto and the Swift wire mapping; S14-02 the Swift view layer. Binds S14-01, S14-02.
- **(Batch 1) The gallery reads one list of gallery keys.** The gallery query filters on a list in `core/vocabkey` (`likeness`, `event_picture`, `place_picture`), not on `PropertyLikeness` alone, and a `subjectvocab` test holds each to a seeded `depiction` Property. The next depiction Property that belongs in a gallery is one entry. Binds S14-01.
- **(Batch 1) New seeds reach new catalogs only.** As in Spike 10: the two Properties come from Install, and no migration adds them to an existing catalog. Binds S14-01.
- **(Batch 1) An Event or Place chooses its thumbnail as a Person chooses its portrait.** Feature parity: the earliest picture leads unless one is chosen; the header, the gallery's chosen badge, set and go-back-to-earliest, and the list row all work as on a Person. One mechanism serves all three kinds: the preferred pointer, `canonicalentities.SetPreferredLikeness` (it checks gallery membership, not kind), the lead rule and `ConclusionPortraitModel`. No per-kind column or RPC. A Person's copy keeps "portrait"; an Event's or Place's says "thumbnail". S14-01 updates `conclusion-layer-data-model.md` (the pointer is the handle's, not the Person's). Binds S14-01, S14-02.
- **(Batch 1) The gallery is one slot on the detail body.** `ConclusionDetailBody` carries one gallery value (title, tile wording, chosen-badge and set wording, fallback mark, tiles) built by one tile mapping, in place of `likenesses` and the Person-only strings, ids and mark in `ConclusionDetailLikenessSection`. Precedent: `ConclusionDetailThumbnail`. Spike 13's Gallery section (S13-14) is built from the same slot, so the old and new galleries can live side by side and retiring the old one does not drop Event and Place galleries. Binds S14-02.
- **(Batch 1) A depiction is worded as a picture outside the Person's own copy.** The Promote evidence sheet words a depiction line by value type ("Same picture", "Compare these pictures") for every kind, and takes its fallback mark from the line's kind. Binds S14-02. **Descoped for Promote (2026-10-29):** Identity Claims no longer carry pins and Spike 13's S13-23 removes the evidence sheet; S14-02 drops its Promote wording and the relationship `kind` wire field. The rule still holds for the Event and Place pages.
- **(Batch 1) A depiction is drawn as a rectangle.** The composer arms the viewer with the rectangle tool alone for a depiction; a Citation's own region keeps all seven tools. "Use the citation's region" takes a non-rectangle region's bounding rectangle. Go keeps accepting any valid depiction region, so a depiction saved as a polygon loads, displays and saves as it is until redrawn. The transparent-polygon crop is refused. Binds S14-03.
- **(Batch 1) Oversized JPEGs get a thumbnail from a reduced decode of our own.** A JPEG over `maxDecodePixels` is decoded at 1/8 scale from its DC coefficients only, then scaled and encoded as today; JPEGs within the budget keep the standard library, and other formats over it stay refused. It is pure Go, with no C library and no platform API, so a Windows build reuses it (user, 2026-10-27). It reads untrusted bytes, so it is bounded and fuzzed ([design](oversized-jpeg-thumbnails.md)). Binds S14-23.
- **(Batch 4) A Citation belongs to one Evidence graph** (user, 2026-10-29). It is made in one graph, stored as `citations.graph_id`, and never moves; a graph shows only its own Citations, and two graphs that need the same reading each make their own Citation. A graph delete deletes its Citations. Binds S14-10.
- **(Batch 4) A write names only the graphs whose loaded rows it moves, and the partial reads it moves by key.** Narrows "Effects are graph-keyed, in one step": a Citation names the graphs of its Observations, an Artifact names none, and a Source title edit names its graphs, which hold the title. Partial reads ride one keyed effect in the shape of `Set.Search`: `Effect.Reads []ReadPath{Read, Path}` → `Set.Reads` → `WriteEffects.reads` (a read name and its ids) → one Swift line per read name. It replaces `citation_sources` with citation reads per graph, carries graph progress only on writes that change counts (Subjects, Observations, Identity Claims), and carries the graph list by Source. Which write moves which read stays in Go's registry (map review K-2). Binds S12-25, S14-11.
- **(Batch 4) A hidden graph or Source page goes stale; it does not reload.** `.evidenceGraph` and `.sourceWorkspace` take `whenHidden: .markStale`, as the lists do. A stale key is not settled: `QueryHandle.isStale`, set by `markStale` and cleared by `beginLoading`, keeps `onSettled` from publishing a cached count before the reload ([idea](../../ideas/archive/settled-list-count-publish.md), user 2026-10-27). Binds S14-11.
- **(Batch 4) File work and decoding happen before the catalog session.** One staged-object type in `ingest` stages bytes from the project folder outside `catalogsession.Do` (stream, hash and sniff, and for an image decode and encode), inserts its `files` row inside the write, and discards what the write did not use. `ingest.File`, `StageImage` (through `StageDepictions`) and the thumbnail derivative all use it, and handlers stage before `withProjectCatalog` (map review K-4). Binds S14-12.
- **(Batch 5) What the app hands to macOS to open is checked first, and Go decides the type.** The `files` read returns an object path only for a type on the ingest allowlist, so Swift checks only that the object has an extension and is not a type that could run a program or open another item, and opens it with the app for its type; a refusal shows an error naming the file. Swift keeps no second allowlist. The viewer checks only that the sniffed type matches the media type Go returned, and a pixel budget; `ThumbnailDecodePolicy` gains that budget check as one shared function and keeps PNG and JPEG as the thumbnail cache's own list (map review K-8, K-23). A metadata link opens only through `open_href`, which Go computes with `urlshape.OpenHref` (web links only). Binds S12-23, S14-13.
- **(Batch 5) Kit components announce themselves.** `PVToast` posts its own announcement, and `PVCallout` takes an opt-in announcement for an error, which the graph form and the Artifact card use; no call site posts its own (map review K-11). Binds S14-14.
- **(Batch 5) A danger confirm's default action is Cancel.** Return cancels; the destructive button has no Return shortcut. The kit README changes with it. Binds S14-14.
- **(Batch 5) A dirty form asks before a link inside it navigates away.** A Kept-linked Citation in the graph form asks "Discard changes?" first. Overrides S12-D3 GF-8. Binds S14-14.
- **(Batch 5) No graph name in the canvas header.** Refused: the breadcrumb names the graph, and S14-14 adds it to the canvas's VoiceOver label and the Back / Forward history rows through the breadcrumbs' own name lookup (map review K-16). Binds S14-14.
- **(Batch 5) Every audited write returns its effects, and a test holds them to it.** A test in `api/ffi` holds every request with a `user_id` to a response with a `WriteEffects` field: S14-16 adds it with an allow-list of S14-17's writes, and S14-17 empties the list (map review K-10). Sidebar counts become an effects list, `counts`, set by `sources` and `canonical_entities` creates and deletes, so the recount follows `WriteEffects.lists` and `navCountTriggers` goes (map review K-5). Binds S14-16, S14-17.
- **(Batch 5) FakeStore is a seeded fake, not a second engine.** It returns seeded rows and scripted effects and errors (`nextWriteEffects`, its `…Error` hooks, and impact maps). S12-35 removed graph list order, Citation membership, WriteEffects composition, and graph / Copy Source refusals from the fake. A store method a PR adds follows that scripting shape and adds none of those rules; leftover synthesizers, vocabulary `CoreInvokeError`s, and local sorts outside that set are not the pattern new methods copy. Engine rules stay tested in Go (map review K-9, user 2026-10-27, option a). Binds S12-22, S12-25, S12-35 and every PR that adds a store method.
- **(Batch 5) Proposals are liberal; the researcher decides.** Promote surfaces a candidate on forgiving evidence: a prefix pair of two full names (Joseph / Josephine), a namesake inside the lifespan slack, an exact tie placed on the lowest ref. What files is the researcher's choice. No PR in slices 5b and 6b narrows what the comparers accept. Spike 11's dogfood bar 4 still holds: a row whose comparisons include a conflict shows Contested, whatever its band, which S12-41 restores for decided and held neighbors. Also a Spike 11 working rule.
- **(Batch 5) A name that heads its own nickname set never folds into another set's head** unless `nicknameDirections` names the pair (audit 2 B-3, option a). Correct folds not yet reviewed, such as Jack → John (`jack` heads `{jack, jackie}`), stop until a Direction is added; until then the Person page shows both names. Binds S14-18.
- **(Batch 5) Audit items keep their IDs.** Each PR names its audit items (`A-n` from [`app-health-audit.md`](app-health-audit.md), `B-n` from [`app-health-audit-2.md`](app-health-audit-2.md)); `plan-pr` reads them there and re-checks each against the code, because the audits' line numbers are from Spike 12. Binds S14-13 – S14-22.
- **(Batch 3) S14-09 is Spike 11's S11-29, refreshed.** Its plan and review moved with it and keep their decisions (the user chose the treatment on 2026-10-19: curly quotes, parentheses, italic surname on the two titles; no migration of existing catalogs). The plan is refreshed against `main` with `plan-pr` because S11-28's alias fold and S11-44 changed name display after it was written. Binds S14-09.
- **(Batch 2) Other-link support is 0.75.** `segment.OtherSupport` is 0.75 (`OtherCredit` stays 0.25): a partial given name plus one lined-up non-birth/death link clears Weak on the 2026-10-28 Jake + grandparent dogfood shape, without a kinship-only tier or a name-curve change. Decided user 2026-10-28 (option A). Binds S14-08.

## PR map

| PR | Title | Slice | Scope | Brief | Depends on | Plan | Status |
| --- | --- | --- | --- | --- | --- | --- | --- |
| S14-01 | Pictures engine | 1 | Rename the gallery layer (Go, proto, Swift wire mapping); seed `event_picture` and `place_picture`; the gallery reads the gallery keys and each picture's own image; Event and Place headers carry the lead picture | — | — | [S14-01-pictures-engine.md](prs/S14-01-pictures-engine.md) | Landed #615 |
| S14-02 | Pictures in the app | 1 | The gallery slot and the Swift view-layer rename; Event and Place pages show Pictures, a header thumbnail and choosing a thumbnail; Events and Places lists lead with it (Promote wording dropped 2026-10-29) | — | S14-01, S13-23 | [S14-02-pictures-in-the-app.md](prs/S14-02-pictures-in-the-app.md) | Blocked — re-plan |
| S14-03 | Rectangle-only depiction drawing | 1 | Depiction drawing offers only the rectangle tool; the Citation's region as a depiction uses its bounding rectangle; the composer's drawing feedback, captions and cap announcement | — | S14-13 | — | Scoped |
| S14-04 | Promote wording for pictures | — | Folded into S14-02 (working rule "Batch 1 is cut by layer") | — | — | — | Descoped (2026-10-27) → S14-02 |
| S14-05 | Event and Place headers lead with a picture | — | Folded into S14-01 | — | — | — | Descoped (2026-10-27) → S14-01 |
| S14-06 | Choose an Event or Place thumbnail | — | Folded into S14-02 | — | — | — | Descoped (2026-10-27) → S14-02 |
| S14-07 | Events and Places lists lead with a picture | — | Folded into S14-02 | — | — | — | Descoped (2026-10-27) → S14-02 |
| S14-08 | Other-link support calibration | 2 | Raise `segment.OtherSupport` 0.25 → 0.75 so a thin name plus kinship clears Weak | — | — | [S14-08-other-support-calibration.md](prs/S14-08-other-support-calibration.md) | Landed #613 |
| S14-09 | Quote nicknames and add former surnames | 3 | Nickname parts display in curly quotes; a `former_surname` product part type displays in parentheses, italic on the Persons list row and Person page title (was S11-29) | — | — | [S14-09-name-display.md](prs/S14-09-name-display.md) | Blocked — re-plan |
| S14-10 | A Citation belongs to one Evidence graph | 4 | `citations.graph_id` (migration and backfill); a Citation is created in one linked graph and never moves; graph reads, Observation isolation and unlink go by it; a graph delete deletes its Citations (was S12-52) | — | — | [S14-10-citation-in-one-graph.md](prs/S14-10-citation-in-one-graph.md) | Planned |
| S14-11 | Reload only what moved, only when it is on screen | 4 | The graph list re-read per Source through a `reads` entry; Add Source inserts its row; hidden graph and Source page keys go stale until shown; the Sources list sorts once per change (was S12-26) | — | — | — | Scoped |
| S14-12 | File work before the catalog session | 4 | One staged-object type stages ingest, depiction images and thumbnails outside `catalogsession.Do`; Go only (was S12-27) | — | — | — | Scoped |
| S14-13 | A crafted project can't crash the app | 5 | Bounded graph positions; a bounded Artifact viewer decode; the PDF crop re-checks its file; panic recovery at the FFI boundary (was S12-28) | — | — | [S14-13-crafted-project-safe.md](prs/S14-13-crafted-project-safe.md) | Planned |
| S14-14 | Copy and VoiceOver on graph work | 5 | The graph named in VoiceOver and history; plural counts; Cancel is a danger confirm's default; a dirty graph form asks first; kit toasts and callouts announce; headings; `PVField` names its content (was S12-29, S12-30, S12-45) | — | — | — | Scoped |
| S14-15 | Xcode synchronized folders | 5 | `project.pbxproj` on synchronized folders; the pbxproj steps leave the skills and rule (was S12-31) | — | — | — | Scoped |
| S14-16 | Promote and Source writes return their effects | 5 | Promote and Source handlers return `WriteEffects`; sidebar counts become a `counts` effects list; a test pins effects on every audited write; the dead `promoteSubject` path goes (was S12-32) | — | S13-24 | — | Scoped |
| S14-17 | Vocabulary writes return their effects; dead RPCs go | 5 | Vocabulary handlers return `WriteEffects` and the allow-list empties; unclaimed RPCs and test-only store methods deleted (was S12-33) | — | S14-16 | — | Scoped |
| S14-18 | Name folds keep the formal name and never crash | 5 | The reconciler resolves values through the fold graph without panics; nickname heads fold only through `nicknameDirections`; `CacheVersion` bump; Go only (was S12-40) | — | — | — | Scoped |
| S14-19 | A merge leaves no stale link in the graph cache | 5 | `PROVENENCIA_GRAPH_VERIFY=1` passes `TestMemoFollowsCommit/a_handle_merged`: a handle merge drops the link the catalog no longer has; Go only (was the bug half of S12-43) | — | — | — | Scoped |
| S14-20 | The Promote page recovers and speaks | 5 | One serial, latest-wins proposal runner; Go releases a decided pair whose Subject or record went away and names it; errors scrolled into view and announced; the "start as Skip" line goes (was S12-44) | — | S14-14, S13-24 | — | Scoped |
| S14-21 | Search finds every spelling | 5 | The search document carries folded spellings and nickname expansions; a join reprojects only what changed; FakeStore's search returns seeded hits in seeded order (was S12-47, S12-50) | — | — | — | Scoped |
| S14-22 | Promote's wire says what Go decided | 5 | The proposal says whether each connection will file and why not through one pure cycle rule; name part types guarded (was S12-48) | — | S14-09, S13-23 | — | Scoped |
| S14-23 | Thumbnails for oversized JPEGs | 1 | A pure-Go reduced decoder in `core/raster` gives a JPEG over the 64-megapixel budget a 1/8-scale thumbnail; bounded and fuzzed; Go only (was S12-39) | — | — | — | Scoped |
| S14-99 | Dogfood close / docs | — | Bar pass, [`archive-docs`](../../../.agents/skills/archive-docs/SKILL.md) | — | all | — | Scoped |

## Design track

No briefs. S14-02 reuses the Person page's gallery section, header thumb, portrait controls and list thumbnail with new headings and wording; S14-03 shows fewer radios on an existing strip.

## Slices

**✎ = design brief**, just before the PR it gates. **Check** = what you can verify in the app when the slice lands.

```text
SLICE 1 — Event and Place pictures (batch 1)
  S14-01  Pictures engine
  S14-02  Pictures in the app
  S14-03  Rectangle-only depiction drawing
  S14-23  Thumbnails for oversized JPEGs
  Check: bar outcomes 1–9 and 12 on a new project.

SLICE 2 — Matching calibration (batch 2)
  S14-08  Other-link support calibration
  Check: bar outcome 10 on the Jake + grandparent dogfood Source (or the calibration test's shape).

SLICE 4 — Evidence graphs finished (batch 4)
  S14-10  A Citation belongs to one Evidence graph
  S14-11  Reload only what moved, only when it is on screen
  S14-12  File work before the catalog session
  Check: bar outcomes 13 – 14 on the 60-graph census at the Scale.

SLICE 5 — Audit fixes (batch 5)
  S14-15  Xcode synchronized folders   (first, before PRs that add or move Swift files)
  S14-13  A crafted project can't crash the app
  S14-14  Copy and VoiceOver on graph work
  S14-18  Name folds keep the formal name and never crash
  S14-19  A merge leaves no stale link in the graph cache
  S14-21  Search finds every spelling
  S14-16  Promote and Source writes return their effects   (after S13-24)
  S14-17  Vocabulary writes return their effects
  S14-20  The Promote page recovers and speaks   (after S14-14, S13-24)
  S14-22  Promote's wire says what Go decided   (after S14-09, S13-23)
  Check: bar outcomes 15 – 20.

SLICE 3 — Name display (batch 3)
  S14-09  Quote nicknames and add former surnames
  Check: bar outcome 11 on a new project.

S14-99 — Dogfood close / docs
```

S14-01 and S14-03 can run in parallel; S14-02 follows S14-01. S14-03 waits on S14-13 (was Spike 12's S12-28), which changes in place the one predicate that says whether the viewer can cut a file; S14-03 gates Draw on it. S14-08 is independent of batch 1. S14-09 is independent of the other batches; S14-22 waits on it.

### Dependencies at a glance

```text
S14-01 ──▶ S14-02
S13-23 ──▶ S14-02   (Spike 13: deletes the Promote files the old plan renamed)
S14-13 ──▶ S14-03
S14-16 ──▶ S14-17
S14-14 ──▶ S14-20
S13-24 ──▶ S14-16, S14-20      (Spike 13)
S13-23 ──▶ S14-22             (Spike 13)
S14-09 ──▶ S14-22
S14-08  (independent)
```

---

## PR scopes

Input to [`plan-pr`](../../../.agents/skills/plan-pr/SKILL.md). Two to six lines each. Link, do not restate.

### Slice 1 — Event and Place pictures

#### S14-01 — Pictures engine

**Ships.** The gallery-layer rename in Go and proto, with the Swift wire mapping (`GoStore`, `CatalogLikeness`) renamed to compile. `event_picture` and `place_picture` in `subjectvocab/registry.go` with `vocabkey` constants; the gallery reads the gallery keys, so `conclusiondetails.ForEntity` returns tiles for an Event or Place. `EventHeader` and `PlaceHeader` carry the lead image from the lead rule (Go, proto, `GoStore`), as `PersonHeader` does, and a chosen thumbnail refreshes the cached header. The gallery reads each picture's image File from its own `depiction_values` row, not the copy in `auto_reconciler_values.value_image_file_id` (likeness audit SEC-4). `observationstest.SaveDepiction(…, key, …)`, with `SaveLikeness` a wrapper. Updates the docs in [design § Domain docs](event-place-pictures.md#domain-docs-the-prs-update).
**Not this PR.** Anything the app shows (S14-02); drawing (S14-03); renaming the preferred-pointer column.
**Check.** `TestSeedProductValues` cases for both keys and the gallery-key test; a gallery test where a Place's picture is in its gallery and a Person's likeness is unchanged; an Event's and a Place's header carry the earliest picture, then the chosen one after `SetPreferredLikeness`; a handler test for the detail's header path; a gallery test where an edited `auto_reconciler_values.value_image_file_id` does not change a tile's image; the app builds and its tests pass with the renamed mapping.
**Context.** Working rules "Event and Place each get their own picture Property", "The gallery layer is named gallery", "The gallery reads one list of gallery keys", "New seeds reach new catalogs only", "An Event or Place chooses its thumbnail"; [design § What changes](event-place-pictures.md#what-changes); `compose.go` `onePerson`, `conclusion_detail.go` `attachHeader`; [`add-seeded-vocabulary`](../../../.agents/skills/add-seeded-vocabulary/SKILL.md), [`add-depiction-value`](../../../.agents/skills/add-depiction-value/SKILL.md), [`add-ffi-handler`](../../../.agents/skills/add-ffi-handler/SKILL.md).

#### S14-02 — Pictures in the app

**Ships.** The gallery slot, with the Swift view-layer rename (`ConclusionDetailLikeness`, `ConclusionDetailLikenessSection`, `LikenessTileView`, `CachedThumbnail.likenessCrop`). The Event and Place pages show Pictures, lead their header thumbnail with the picture, and take a `ConclusionPortraitModel` (set as thumbnail, chosen badge, use the earliest). `EventsListView` and `PlacesListView` rows lead with it as `PersonsListView.thumbnailRelPath` does. The Promote sheet's wording left this PR on 2026-10-29: the sheet is removed by Spike 13's S13-23. The composer's depiction translator comments and `depictionPerson` say Subject. Strings through `add-localized-string`. Properties list VoiceOver distinguishes two Properties that share the label Picture (include the machine key and/or bound kind labels in `rowAccessibilityLabel` — S14-01 ship UX-1).
**Not this PR.** Spike 13's redesigned Gallery section (S13-14).
**Check.** On a new project: a Place picture shows under Pictures and opens its record; choosing a second picture leads the header and the Places list with it and badges it; use the earliest goes back; the same on an Event; a Person's page and list are unchanged; VoiceOver on the Properties list can tell `event_picture` from `place_picture`. Swift tests beside `PersonDetailTests` and `PersonsListTests`.
**Context.** Working rules "The gallery is one slot on the detail body", "An Event or Place chooses its thumbnail", "A depiction is worded as a picture", "Batch 1 is cut by layer"; [design § What already works](event-place-pictures.md#what-already-works); `ConclusionDetailPage.swift`, `ConclusionDetailThumbnail`, `PersonDetailContent`, `PersonDetailView`, `ConclusionListRow.swift`, `PropertiesView.swift` `rowAccessibilityLabel`.

#### S14-03 — Rectangle-only depiction drawing

**Ships.** The viewer's region strip takes the tools to offer as `[ArtifactRegionTool]` composer state beside `viewerArmedRegionTool` (the viewer stays unaware of depictions); a depiction draw offers `.rectangle` alone, a Citation region all of `ArtifactRegionTool`. `ArtifactRegionDraft.boundingRectangle`, beside `.whole` and built from `ArtifactRegionGeometry.boundingRect(of:)` and `rectanglePoints`, is what "Use the citation's region" sets on the depiction. While a region or a depiction is drawn, Set page, the locator chip × buttons and Find are disabled and the draw hint hides; Draw and the region buttons are blocked on a file the viewer cannot cut, and a failed cut shows with Save disabled; the needs-page callout sets a PDF Citation's page (UX-5, UX-6, UX-7). The empty, needs-page and region-hidden captions name the region buttons, not only Draw, and an edit that saves no image drops "Image to save" (UX-12, UX-13). The freehand tool's point cap (`ArtifactRegionOverlay`) is announced to VoiceOver, not only beeped. Draw's block and the failed-cut state read one "can cut" predicate, `ArtifactRegionCrop.snapshot`, as S12-28 leaves it.
**Not this PR.** Removing shapes from Citation regions; Go-side rectangle enforcement; rewriting saved polygon depictions.
**Check.** Draw on a likeness shows one radio; a Citation region shows seven; Swift tests for the strip's tool list and the bounding-rectangle conversion; an old polygon depiction still opens and saves; bar outcome 9, with Swift tests for the disabled controls, the blocked Draw and the failed-cut state.
**Context.** Working rule "A depiction is drawn as a rectangle"; `ArtifactViewer.swift` `regionTools`, `CitationComposerModel+Depiction.swift` `startDepictionDrawing`, `depictionCitationRegion`, `useDepictionRegion`; the audit items in [`depictions-and-likenesses.md`](../../ideas/depictions-and-likenesses.md) (2026-10-18 likeness audit). If the plan runs past about ten steps, the composer feedback splits out as its own PR.

#### S14-23 — Thumbnails for oversized JPEGs

**Ships.** A reduced JPEG decoder in `core/raster` that reads only DC coefficients and returns the image at 1/8 scale: baseline and progressive Huffman, 8-bit, one to three components with any sampling factors, restart intervals; arithmetic coding, 12-bit, lossless and four-component JPEGs return `ErrTooLarge`. `EncodeJPEG` uses it when the format is JPEG and the declared size is over `maxDecodePixels`, then scales and encodes as today, so `derivatives` stores the usual thumbnail row and the next Source open shows it. Every count read from the file is checked before use, and the 1/8 image counts against the budget. `core/raster` also takes a byte limit of its own, so every decode is bounded by bytes as well as pixels instead of trusting each caller's cap (likeness audit SEC-3, [idea](../../ideas/depictions-and-likenesses.md)). Go native fuzzing over the decoder. Go only.
**Not this PR.** EXIF orientation; a larger cover derivative; other formats over the budget; the Mac viewer (S14-13).
**Check.** Bar outcome 12. A JPEG just over 64 MP, built in the test from a lazy image with no full buffer, thumbnails to a 256 px long edge; a header-only JPEG declaring a huge frame is refused without allocating it; small baseline and progressive JPEGs decode within tolerance of the standard library's decode scaled by 8; the fuzz seeds pass; PNG and GIF header bombs still return `ErrTooLarge`; an input over the byte limit is refused before any decode.
**Context.** Working rule "Oversized JPEGs get a thumbnail from a reduced decode of our own"; [design](oversized-jpeg-thumbnails.md); `core/raster` (`EncodeJPEG`, `decode`, `maxDecodePixels`; the decoder runs under its caller's `Slot` and takes none itself, map review K-24), `core/derivatives/ensure.go`, `api/ffi/handlers/thumbnails.go`; ITU-T T.81. Preferred after S14-12. If the plan runs past about ten steps, progressive support splits into its own PR.

### Slice 3 — Name display

#### S14-09 — Quote nicknames and add former surnames

**Ships.** Nickname parts display in curly quotes. A `former_surname` product part type (role `family`, after the nickname in both Western orders) displays in parentheses, with the surname italic on the Persons list row and the Person page title. Was Spike 11's S11-29; the plan is [S14-09-name-display.md](prs/S14-09-name-display.md), refreshed before it ships.
**Not this PR.** Migrating existing catalogs; italics on other surfaces; a Go↔Swift part-type parity check (S14-22).
**Check.** Bar outcome 11: *James “Jake” Robins*; *Jones, Mary (Smith)* with *Smith* italic.
**Context.** Working rule "(Batch 3) S14-09 is Spike 11's S11-29, refreshed"; [`seeded-vocabulary.md`](../../seeded-vocabulary.md) §4; [`structured-name-model.md`](../../structured-name-model.md) §4.2a, §4.5; [`ideas/international-names.md`](../../ideas/international-names.md); `add-name-value`, `add-localized-string`. Spike 12's S12-48 waits on this PR.

### Slice 4 — Evidence graphs finished

#### S14-10 — A Citation belongs to one Evidence graph

**Ships.** `citations.graph_id` (NOT NULL, an audited CASCADE facet of its graph), backfilled from the Citation's earliest Observation, else its Artifact's first linked graph by name. `citations.CreateInput.GraphID` and `CreateCitationWithObservationsRequest.graph_id` are required, and the composer passes its graph. A graph's Citations are those with its `graph_id`: `ListByArtifact`, `CountByEvidenceGraph`, `requireSameGraph` and the unlink check read it. A graph delete releases its Citations through one `ReleaseCitation`, which `citations.Delete` also calls. `evidencegraphs.citations_need_graph`, its probe and strings go, and S12-25's `prefer` and `citationArtifactGraphs` paths go.
**Not this PR.** A Citation's graph on the wire to Swift; moving a Citation between graphs.
**Check.** A Citation made in one graph shows in no other graph's counts or stepper, before or after its first Observation; a graph delete with Citations and no Subjects lists them and deletes them.
**Context.** S12-25 ship log CON-1; working rule "A Citation belongs to one Evidence graph" (user, 2026-10-29); precedent S12-18; `add-catalog-migration`, `add-catalog-delete`.

#### S14-11 — Reload only what moved, only when it is on screen

**Ships.** A `reads` entry for the graph list by Source (graph and link rows, Source create, copy and delete) that replaces the catalog-wide `evidence_graphs` list key: `effects.ListEvidenceGraphs`, the `evidence_graphs` wire list key, `CatalogList.evidenceGraphs` with its invalidation case and its `swift_vocabulary_test.go` row go. `ListEvidenceGraphsRequest` takes an optional Source, and the session re-reads only the named Sources' graphs and merges them into the cached list through the cache layer. Add and Copy Source insert the row `CreateSource` returns into the Sources list through the same layer, not a feature write-through. `.evidenceGraph` and `.sourceWorkspace` go stale while hidden, and `QueryHandle.isStale` (set by `markStale`, cleared by `beginLoading`) keeps `isSettled` false until a stale key reloads. `SourcesModel.visibleSources` is memoized against the revisions it reads. Updates `macos-client-patterns.md` rule 4 (list membership patched from a write's own row) and `source-layer-data-model.md` (the graph list's effect).
**Not this PR.** The descriptor itself (S12-34): the graph-list merge and the Sources-list insert are its `refresh` and `patch` entries.
**Check.** A link write re-reads one Source's graphs; an off-screen graph reloads only when revisited; a list that reappears after its key went stale publishes its count only after the reload; typing in Add Source does not re-sort the list (model tests).
**Context.** Audit A-9, A-10, A-25; [`ideas/archive/settled-list-count-publish.md`](../../ideas/archive/settled-list-count-publish.md) (its open question); working rules "One graph list", "A hidden graph or Source page goes stale; it does not reload"; S12-16 (why the hand-written insert went); `add-workspace-place`, `add-ffi-handler`. After S12-25, whose keyed effect it adds a read to.

#### S14-12 — File work before the catalog session

**Ships.** One staged-object type in `ingest`: it stages bytes from the project folder outside `catalogsession.Do` (stream to an `objectstore` temp file, hash and sniff, and for an image decode and encode), inserts its `files` row inside the write (`LookupByChecksum`, `Install`), and discards what the write did not use. `ingest.File`, `StageImage` (through `StageDepictions` in the citation, observation and connect handlers, one stage-then-write entry point) and the thumbnail derivative (`derivatives.Ensure`) all use it, and handlers stage before `withProjectCatalog`; the thumbnail made today inside the session through `thumbnailRelPath` splits into read, decode outside, and write (map review round 2). The stored-image media type becomes one constant in `core/depictionvalue`, used by `stage.go` (three literals today) and `ReadableImage` ([idea](../../ideas/depictions-and-likenesses.md)). Go only. Updates `use-catalog-session` and `source-layer-data-model.md`.
**Not this PR.** Oversized-JPEG thumbnails (S14-23); storing crops below 2,048 px ([idea](../../ideas/depictions-and-likenesses.md)).
**Check.** Bar outcome 14, last clause. A Go test holds a slow reader in the stage step while another session call completes; a depiction save decodes before the session; the ingest, dedupe, purge and depiction tests pass unchanged.
**Context.** Audit A-24; working rule "File work and decoding happen before the catalog session"; map review K-4; `core/ingest/stage.go` (`StagedImage`, the split to generalize); `derivatives.Ensure`; [`source-layer-data-model.md`](../../source-layer-data-model.md) (a purge and an import never interleave); `use-catalog-session`. **Spike 12 ship-log leftover (2026-10-29):** `api/ffi/handlers/thumbnails_test.go` still runs a raw `DELETE FROM file_derivatives` where the harness's `damageCatalog` exists.

### Slice 5 — Audit fixes

#### S14-13 — A crafted project can't crash the app

**Ships.** `subjectpositions` refuses a position outside one cell range on write and clamps it on read; the Swift placement code uses checked arithmetic. The Artifact viewer checks, before it decodes, that the sniffed type matches the media type Go returned and that the declared pixel size is within a budget, through one shared check in `ThumbnailDecodePolicy` (no type list of its own; the policy keeps PNG and JPEG for thumbnails); it shows a screen-sized decode and crops from a bounded one. Whether a file can be cut stays one predicate (`ArtifactRegionCrop.snapshot`), changed in place, which Spike 14's S14-03 gates Draw on. `ffi.Call` recovers a panic into `internal.unknown`, and `catalogsession.Do` defers its unlock. The PDF crop (`ArtifactRegionCrop`) reopens its file through `ProjectFiles.objectURL`: the off-main cut carries the project folder and the object's path instead of a bare URL ([idea](../../ideas/depictions-and-likenesses.md)).
**Not this PR.** Length caps on catalog text (audit L-10); the foreign-catalog open order (audit L-7).
**Check.** Bar outcome 15, last clause. Go tests for the position bounds and for a recovered panic that leaves the session usable; a Swift test feeds the placement extreme positions.
**Context.** Audit A-26, A-27, A-28; working rule "What the app hands to macOS to open is checked first, and Go decides the type"; map review K-8; `core/raster` for the bounds to reuse; `add-ffi-handler`. After S12-23, whose `files` filter the viewer relies on. Spike 14's S14-03 waits on this PR (map review K-25).

#### S14-14 — Copy and VoiceOver on graph work

**Ships.** Three Spike 12 PRs in one (user, 2026-10-29), each part as its scope said:
- **Name the Evidence graph; counts take plurals** (A-12, A-14, A-20): One function gives a location's label from the session's names; the breadcrumbs, the Back / Forward history rows (graph, composer and Promote) and the canvas's VoiceOver label all read it. The A-14 strings name the Evidence graph. The A-20 strings get plural variations, `DeleteImpact.noun` reads plural keys, and `check-localizable-xcstrings.py` fails on a `%lld` noun with no variations.
- **Dialogs and VoiceOver behave safely** (A-13, A-17, A-18, A-21): In `PVConfirm` the danger tone makes Cancel the default action and drops Return from the destructive button. The graph form's Kept-linked Citation rows ask "Discard changes?" when the form is dirty and are disabled while it saves. `PVToast` posts its own announcement, and `PVCallout` takes an opt-in announcement for an error, which the graph form's and the Artifact card's save errors use, and so do the Source page error (`sources.page.error`) and the canvas file-open callout (`evidenceGraph.fileOpen.callout`), which show S12-23's Open refusal (S12-23 review UX-3). `PVSectionHeader` titles are headings. Updates `macos/App/DesignSystem/README.md`.
- **Form fields are named** (B-18): `PVField` ties its label to its content in the kit, which gives every form's fields their spoken names, and the hand-made label in `CitationComposerEvidenceColumn.labelField` goes (B-18). Updates `macos/App/DesignSystem/README.md` (the Field row). Swift only.
- **Focus after the Source type editor** (a Spike 11 leftover): closing it returns keyboard focus to the Type control, as Promote's record search does with `PVComboFocus`.

If the plan runs past about ten steps, it splits along these parts.
**Not this PR.** A graph name in the canvas header (refused); tooltips on truncated names; holding toasts longer while VoiceOver runs.
**Check.** Bar outcome 17; the plurals script fails on a planted `%lld subjects` key.
**Context.** Audit A-12, A-13, A-14, A-17, A-18, A-20, A-21, audit 2 B-18; working rules "A danger confirm's default action is Cancel", "A dirty form asks before a link inside it navigates away", "Kit components announce themselves", "No graph name in the canvas header"; `WorkspaceToolbar.swift`, `macos/App/DesignSystem/README.md`; `add-ui-component`, `add-localized-string`, `add-swift-test`. **Spike 12 ship-log leftovers (2026-10-29):** a failed Citation load in the composer shows its error in the Transcription field's slot and does not name the Citation (`CitationComposerModel` `fields.error`); `ProjectFiles.openObject` ignores `NSWorkspace.open`'s result, so a failure after an app is found shows nothing; and "Link to a graph" or the pencil on one Artifact card drops another card's unsaved edit without asking, which this PR's dirty-form rule extends to.

#### S14-15 — Xcode synchronized folders

**Ships.** The App and test groups of `project.pbxproj` become synchronized folders; `add-swift-test`, `add-ui-component` and `.cursor/rules/swift-tests.mdc` drop their pbxproj steps.
**Not this PR.** Any source change.
**Check.** `xcodebuild` build and test pass; a new Swift file builds with no pbxproj edit.
**Context.** Audit A-34. Needs a Mac; lands best before the other slice 5–7 PRs add Swift files.

#### S14-16 — Promote and Source writes return their effects

**Ships.** `ApplyPromoteGraphAlignment`, `UpdateSource`, `DeleteSource` and `SetSourceCover` return `WriteEffects`, and their callers apply them. Sidebar counts become an effects list, `counts`, set by `sources` and `canonical_entities` creates and deletes; the recount follows `WriteEffects.lists` and `navCountTriggers` goes. `.promotedSubject` goes when `conclusionTriggers` no longer needs it, and the other `CatalogMutation` cases the effects now cover go with it. The dead `promoteSubject` path goes (handler, proto method reserved, store methods, FakeStore). A test in `api/ffi` holds every request with a `user_id` to a response with `WriteEffects`, with an allow-list of S14-17's writes.
**Not this PR.** Vocabulary writes (S14-17).
**Depends on.** Spike 13's S13-24, which first removes Apply's pin and claim-field inputs from `promote_batch.go` and `GoStore.applyPromoteGraphAlignment`, the handler and mapping this PR rewrites (2026-10-29).
**Check.** Handler tests assert each write's effects; the pin test passes with its allow-list; a Promote Done refreshes the canvas, progress and the sidebar counts through effects alone; a Source rename refreshes Promote's Source titles.
**Context.** Audit A-29, A-35; working rule "Every audited write returns its effects, and a test holds them to it"; map review K-5, K-10; [`macos-client-patterns.md`](../../macos-client-patterns.md) cache contract rule 2 (its exceptions shrink) and `add-workspace-place` (it teaches `navCountTriggers`), both updated; S12-16 (`navCountTriggers`); `add-ffi-handler`. **Spike 12 ship-log leftover (2026-10-29):** the plan confirms the newest recount wins when two finish out of order (today `CatalogCounts.refreshAll` is last-writer-wins).

#### S14-17 — Vocabulary writes return their effects; dead RPCs go

**Ships.** The Source type, metadata field, Property, binding and term writes return `WriteEffects`, their callers apply them, and the `CatalogMutation` cases that go dead are deleted. `EnsureFileThumbnail`, `GetSubjectTypePresentation`, `ListPlaceableSubjectTypes`, `ListProperties`, `ListSubjectTypeProperties` and `RemoveInstallIdentity` go with their handlers, tests and FakeStore code (method numbers reserved), as do the test-only `listSubjectTypes` and `clearSubjectPosition`.
**Not this PR.** Deleting `UpdatePropertyTerm`, `DeletePropertyTerm` or `ReorderSubjectTypeProperties` (named by idea files): they stay and return `WriteEffects` like the rest, so the allow-list empties; the audit's test-net items (L-16).
**Check.** Handler tests assert effects; the pin test's allow-list is empty; no RPC is left without a Swift caller except the three kept; `macos-client-patterns.md` rule 2 lists only exceptions effects cannot express.
**Context.** Audit A-29, A-35; working rule "Every audited write returns its effects, and a test holds them to it"; `add-ffi-handler`. After S14-16, whose pin test it empties. The plan confirms `RemoveInstallIdentity` is not planned before deleting it.

#### S14-18 — Name folds keep the formal name and never crash

**Ships.** `foldValues` resolves each value through the fold graph to a maximal value it can reach (best supported, then best ranked); a value with no reachable target, or on a cycle, is its own root (B-1). `AliasIndex.Expands` treats a name that heads any set as formal: it folds only where a Direction names the pair (B-3). `TestReconcileNamesInvariants` runs with the product alias tables over a pool of nicknames, initials, prefixes and accented forms in one- to three-word units, asserting no panic and no root with a target; table rows for the B-1 chains and the B-3 examples, with Jack / John as the known cost. `CacheVersion` bump and its pin. `conclusion-reconciliation.md`, `seeded-vocabulary.md` and `add-seeded-vocabulary` say what a head is. Go only.
**Not this PR.** The FFI panic recover (S14-13); accented alias lookups (descoped with S12-46 to [`ideas/promote-ranking.md`](../../ideas/promote-ranking.md)); adding unreviewed folds to `nicknameDirections` beyond the cases above (the user, as dogfood finds them).
**Check.** Bar outcome 18. A project that holds B., Bill and William opens after the bump and shows William; Ann, Anna and Hannah file and reopen without a crash; Charlotte, Charlotte, Chick shows Charlotte; Jack and John show both names.
**Context.** Audit 2 B-1, B-3; working rules "A name that heads its own nickname set never folds into another set's head", "Proposals are liberal; the researcher decides"; Spike 11's S11-28 plan (D-1a, Appendix A) and S11-45's fold; `add-match-rule`, `add-seeded-vocabulary`. First in slice 5b.

#### S14-19 — A merge leaves no stale link in the graph cache

**Ships.** With `PROVENENCIA_GRAPH_VERIFY=1`, `promotealign` `TestMemoFollowsCommit/a_handle_merged` fails today: after a handle merge the graph cache keeps a link the catalog no longer has. The merge's effects drop that link, and the verify-mode test passes. Go only.
**Not this PR.** S12-43's performance work (the shortlist index kept on commit, profile memos, budgets at Scale), descoped 2026-10-29 to [`ideas/promote-ranking.md`](../../ideas/promote-ranking.md).
**Check.** `PROVENENCIA_GRAPH_VERIFY=1 go test ./core/database/promotealign/...` passes.
**Context.** Bar outcome 19; `core/database/graphcache` (memos and `Deps`), the merge path in `canonicalentities`; `use-catalog-session`.

#### S14-20 — The Promote page recovers and speaks

**Ships.** One serial, latest-wins runner *(new)* that runs a call, keeps only the newest pending input and sends it when the call returns; proposals use it in place of `proposeGeneration`, so Go runs one proposal at a time (B-12). Propose releases a fixed pair that `checkFixed` finds invalid (its Subject gone from the graph, its record gone or merged, its kind changed) instead of failing, and the proposal returns each released pair with a reason, a typed wire vocabulary with its `swift_vocabulary_test.go` row; Swift sends what the researcher decided and words a released row with the row-note pattern (`conflictNote`, `duplicateNote`), so the other decisions stay (B-13). Apply checks the revision before filing, and a `promote.invalid` at Done is a real race that proposes again once, as stale does, with the stale note cleared when that proposal fails. Stale, save and proposal errors scroll into view when they appear and are announced through the kit's callout announcement; a failed load offers Try again; the page's existing callout, so no new chrome (B-17). The "start as Skip" line goes or is reworded (B-16). FakeStore scripts the new errors and released pairs. `promote-graph-alignment.md` says how a decided pair whose Subject or record went away is released.
**Not this PR.** The proposal's own cost (descoped 2026-10-29 to [`ideas/promote-ranking.md`](../../ideas/promote-ranking.md)); Promote Done's effects (S14-16).
**Check.** Bar outcome Spike 12 19: five quick picks run one proposal at a time. Bar outcome 19, recovery: B-13's manual steps end with Mary's row released and named and the other decisions kept; VoiceOver hears a save failure; the page header no longer mentions Skip.
**Context.** Audit 2 B-12, B-13, B-16, B-17 (map review K-30, K-39); working rules "Kit components announce themselves", "FakeStore is a seeded fake, not a second engine"; the S11-35 ship log (SEC-out-1); `core/database/promotealign` `checkFixed`; `add-localized-string`, `add-ffi-handler`, `add-swift-test`. After S14-14, whose callout announcement it uses, and S13-24, which removes the evidence read and `ResolvePins`.

#### S14-21 — Search finds every spelling

**Ships.** The search document carries each value's `Equivalent`-folded form where FTS retrieves on it, so "Soby" finds Søby with or without an alternate, and `distinctAlternates` drops no spelling FTS cannot fold (B-21). A folded alternate in the document's secondary text, tagged as toponyms are today, needs no new column; the plan adds a column and a migration only if it finds one needed. `ProjectionVersion` bump. `RecomputeReported` reprojects a root's dependents only when the root's header-visible row changed, and composes unchanged dependents from the warm graph (B-10). Go only.
**Not this PR.** The search registry's columns and projector table (descoped with S12-36); ranking changes.
**Check.** Bar outcome 20, search: B-21's probe finds Ærøskøbing, Søby, Łask and Œuilly from their ASCII spellings. The join: joining a place Subject to a 6,000-event place reprojects only what changed, held by a budget subtest.
**Context.** Audit 2 B-10, B-21 (map review K-45); Spike 11 dogfood bar 6; S11-07 D-6; `core/database/searchindex` (`matchText`, `distinctAlternates`), `conclusionheaders.RowsForSearch`; `add-searchable-kind`, `add-catalog-migration` if a column is added. **Spike 11 leftover (2026-10-29):** since S11-28 a nickname folds into its formal name, so searching "Tom" no longer finds a person shown as Thomas; the search document carries the alias-table expansions of a person's given names.
**Also ships (was S12-50, Swift):** FakeStore's search returns seeded hits in seeded order, filtered only by the requested kinds, as `add-searchable-kind` and `.cursor/rules/searchable-kinds.mdc` now say (B-7): `searchPriority`, `orderedByMixedScore`, the scoring in `handleSearchHits`, `scoreRefRecord`, `scoreSearchFields`, `classifyRefQuery` and `tokenizeSearch` go, and `vocabularySearchHits` returns seeded locations instead of mirroring Go's `locationFor`. The `SearchCatalogStoreTests` cases that test the fake's own ranking go (`fakeSearchFoldsLikeGo`, `kindPriorityIsATenthAndSectionNudgeStillMatters`, `prefixRefPromotesSource`); a test that needs an order scripts it. Ranking stays tested in Go (`core/search`). `add-searchable-kind` drops its `searchPriority` line. Swift only.

#### S14-22 — Promote's wire says what Go decided

**Ships.** The proposal reports, per connection, whether it will file at Done and why not (a self-link, a `part_of` that would close a cycle). The cycle rule becomes one pure function over edges: the filer in `core/database/identityclaims/bridgefile.go` feeds it from SQL, and the proposal from the canon edges it already holds plus the page's fixed connections, so a proposal runs no catalog scan (its work at Scale is the place edges in memory). The reasons are a typed enum in `Platform/ConclusionVocabulary.swift` with a `swift_vocabulary_test.go` row. Swift's subrows and footer count read the result, a refused connection is drawn and worded as the self-link subrow is today (no new chrome), and Swift's own filing rule goes (B-23). `NamePartType` stays beside its owner, as `.cursor/rules/macos-client.mdc` allows; it gets a guard row, and `add-name-value` step 6 names the match-registry role, the format seed roles and the Swift case (B-24). `promote-graph-alignment.md` §9.1 says where the refusal is reported.
**Not this PR.** Swift's weak-row draft (B-26, moved to Spike 13's S13-24 on 2026-10-29); unread wire fields (B-37).
**Check.** Bar outcome 19, footer: a page whose `part_of` connection would close a cycle shows it as not filing, and Done files what the footer counted; the guard row fails on a new part type.
**Context.** Audit 2 B-23, B-24 (map review K-38, K-40); `.cursor/rules/macos-client.mdc` (wire enums with guard rows); S11-17 (reasons typed), S11-47 D-13; `add-ffi-handler`, `add-name-value`. After S14-09 (Spike 11's S11-29, moved to Spike 14 on 2026-10-29), which adds the part type the guard row covers (its plan's Out row points here), and S13-23, which changes the same Promote rows. The typed band (B-24's first half) is dropped: the band is deprecated (Spike 13 working rule "The band stays, deprecated"). **Spike 11 leftover (2026-10-29):** the Swift-vocabulary guard tests hand-write their expected lists (`api/ffi/swift_vocabulary_test.go`), so a new Go reason passes unnoticed; this PR's typed enums read the list from Go.

### Slice 2 — Matching calibration

#### S14-08 — Other-link support calibration

**Ships.** `segment.OtherSupport` 0.25 → 0.75 in `core/matchregistry/segment.go`; inventory and `promote-graph-alignment.md` quote the new number; `segmentcompare` AC-2.5 / AC-2.8 expects; a graphalign calibration test for Jake given-only + Carol grandparent clearing Weak. Patch VERSION.
**Not this PR.** Name-curve relaxation; Weak-bar change; `OtherCredit`; kinship-only tier; multi-Source corroboration.
**Check.** `TestPartialNameWithGrandparentClearsWeak`; `TestCompare` AC-2.5 / AC-2.8; dogfood Promote row for Jake reads Weak (or stronger), not Little evidence.
**Context.** Working rule "Other-link support is 0.75"; [`matching-model.md`](../../matching-model.md) `segment.OtherSupport`; [`add-match-rule`](../../../.agents/skills/add-match-rule/SKILL.md).
