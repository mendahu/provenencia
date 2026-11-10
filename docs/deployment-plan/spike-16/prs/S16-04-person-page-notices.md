# S16-04 — The Person page says when a reload failed or the diagram is partial

**Status:** Blocked — on board S16-D1
**Spike / slice:** Spike 16 · Slice 1b — Current data and honest notices
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** [S16-D1](../design/S16-D1-person-page-notices.md), board not drawn yet (one brief for the Person page's notices and the record links note: PN-1 – PN-10, PN-W1, PN-W2)
**Board:** — (to be drawn on [Claude Design — Person Detail](https://claude.ai/artifact/N9V52ewQjxeNvh3Az3YkXQ), turn "S16-D1 · Notices"). The implementer reads it with the Artifact tool's `read` on that link. Until it is agreed, **Design → macOS** below describes each state in words from the brief
**Depends on:** S16-03 (landed [#769](https://github.com/mendahu/provenencia/pull/769)), S16-01 (landed [#756](https://github.com/mendahu/provenencia/pull/756)), S16-D1 (board not agreed)
**Blocks:** S16-09; Spike 15's S15-21 (the provenance sheet's key reads the same failure reading and `PVRetryCallout`)
**Branch:** `s16-04-person-page-notices`
**Version bump:** patch — a fix: a failed reload and a partial diagram now say so, a gone person is named a person, and a not-counted record links read sends no lines (old clients draw none); no catalog change: 0.6.0 → 0.6.1, or the next patch on `main` ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-workspace-place`](../../../../.agents/skills/add-workspace-place/SKILL.md) ("Stale-while-revalidate"), [`add-ui-component`](../../../../.agents/skills/add-ui-component/SKILL.md), [`add-localized-string`](../../../../.agents/skills/add-localized-string/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md)
**Rules that bind:** `.cursor/rules/swift-tests.mdc`, `.cursor/rules/go-tests.mdc`, `.cursor/rules/versioning.mdc`
**Context:** [audit](../person-page-audit.md) P-5, P-10, P-17; working rules "(Batch 1) A refresh failure is read beside its value", "(Batch 1) One open section, one list of reads", "(Batch 1) A person's records are records", "(Batch 1) The record links read is bounded and says so"; [`macos-client-patterns.md`](../../../macos-client-patterns.md) (the `QueryHandle` row, "One sectioned detail page", § 9); the Evidence graph's retry notice (`EvidenceGraphView.readErrorCallout`, S14-20), which this PR moves onto the new reading
**Planned:** 2026-11-10 against `main` @ `e70bba8`. The Swift steps need a Mac with Xcode (`xcodebuild`); the Go step and the doc checks run anywhere
**Review:** [`S16-04-person-page-notices.review.md`](S16-04-person-page-notices.review.md) — round 1 answered; round 2 pending

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

When a Person page's reload fails, the page keeps what it showed and says so, with Try again, in the place that read failed: the header band for the page's read, the top of a section for the section's read, the record links column for the diagram's read. A first load that fails offers Try again too, and VoiceOver hears each notice. A person deleted or merged elsewhere leaves with "This person no longer exists". The record links diagram says when its drawing is partial, and a read that did not compare sends and draws no link instead of drawing every link as weak (audit P-5, P-10, P-17).

**Why now.** S16-03 gave every section one list of reads and S16-01 put `notCounted` and `linesDropped` on the wire; both have landed. S16-09 and Spike 15's S15-21 read the failure reading this PR adds.

**What stays.** `QueryHandle.read` and `failure` for their callers; the lists' presentation; the page frame, the header, the "Refreshing" line, the sections' rows, folds and empty states; the diagram's circle, nodes and title; Promote's retry notice; Event and Place pages' gone line.

## 2. User stories

**US-1.** As a researcher, when a section's reload fails after a write, I see a notice with Try again above the rows that stayed, so that I know they are out of date and can read them again.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | Sources is open with two Source rows; `listPersonClaims` throws a generic error; a write names the person | The two rows stay; above them a danger notice "Sources could not be read", message "(the error's text) What is shown is from the last time it was read.", with Try again (PN-1, PN-4) |
| AC-1.2 | AC-1.1, then `listPersonClaims` answers again and Try again is pressed | The Button shows its loading state and the header shows "Refreshing"; the read succeeds and the notice goes (PN-6) |
| AC-1.3 | AC-1.1, Try again pressed with the keyboard, and the read fails again | The notice stays, the keyboard is on its Try again, and VoiceOver speaks the notice again (PN-6, PN-7) |
| AC-1.4 | The page's own read (`getConclusionDetail`) fails behind a shown page, Timeline open | A notice "This page could not be read" with Try again sits under the section switcher; the Timeline under it is unchanged (PN-3). On an Event page, the notice sits under the header, above Details |
| AC-1.5 | The record links read fails behind a shown diagram | A compact notice "The connections could not be read" with Try again sits at the top of the column, above the diagram's title; the Sources rows are unchanged (PN-5) |
| AC-1.6 | A reload fails while no Try again was pressed, or after a Try again that later succeeded | VoiceOver speaks the notice; the keyboard does not move (PN-7) |

**US-2.** As a researcher, when a first load fails, I can try again instead of being told to.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | Opening Family, `getPersonFamily` throws | In place of the tree: "Family could not be read", the error text only, Try again (PN-2) |
| AC-2.2 | AC-2.1, Try again pressed, the read still fails | The notice stays on screen while the read runs (its Button loading), then announces again with the keyboard on Try again (PN-6) |
| AC-2.3 | Opening a Person page, `getConclusionDetail` throws | The notice "This page could not be read" with Try again alone at the top of the page column (PN-3) |

**US-3.** As a researcher, when a person I am looking at was deleted or merged in another window, the app says the person no longer exists.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | Person page open; the detail read answers `conclusiondetails.not_found` (first load or behind a shown page) | The page leaves for the Persons list with the Toast "This person no longer exists. They may have been merged into another person." (PN-8) |
| AC-3.2 | Sources open; `listPersonClaims` answers `conclusiondetails.not_found` before the page's own read does | Sources shows a danger notice "Sources could not be read" with that line, no Try again, no rows and no column (PN-8, PN-10) |
| AC-3.3 | An Event or Place page whose entity is gone | The Toast keeps "This record no longer exists. It may have been merged into another." (PN-8) |

**US-4.** As a researcher, when the record links diagram is partial, a note says so, and a read that did not compare draws no link.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-4.1 | A person whose record links read passes `MaxReadWork` | Go sends the graphs, no pair, `not_counted` true, `lines_dropped` false; the app draws every node, no arc; each node is spoken as its ref, graph and Source only; the note "Too many records to compare. The links between them are not drawn." sits in place of the caption (PN-9) |
| AC-4.2 | A read with `linesDropped` | Every arc sent is drawn; the note "Too many records to draw them all. Some records and links are left out." (PN-9) |
| AC-4.3 | A read with 1,001 distinct graphs and neither flag | 1,000 nodes drawn; the same "left out" note (PN-9) |
| AC-4.4 | A read with `notCounted` and more than 1,000 graphs | The "did not compare" note (PN-9) |
| AC-4.5 | A full read (neither flag, at most 1,000 graphs) | No note; the caption as today |

## 3. Scope

| In | Covered by |
| --- | --- |
| `QueryHandle.reading` (loading, failed, loaded with an optional refresh failure), `failureRevision`, `isReading`; `read` and `failure` unchanged | STEP-1, AC-1.1, AC-1.3 |
| `ConclusionDetailPresentation` built from the reading: a refresh failure beside the content; `conclusiondetails.not_found` behind a value reads as gone | STEP-2, AC-1.1, AC-3.1 |
| PN-6, PN-7 in the kit: `PVRetryCallout` speaks again per failure revision and keeps the keyboard after a pressed Try again; the Evidence graph moves onto it and its own counter goes | STEP-4, AC-1.3, AC-1.6, AC-2.2 |
| PN-3: the page's notice in the header band and on a first-load failure; Event and Place pages too | STEP-5, AC-1.4, AC-2.3 |
| PN-1, PN-2, PN-4, PN-5, PN-10: the section host's band and column notices over stale content, over an empty state and on a first-load failure; Try again rereads the host's read | STEP-6, AC-1.1, AC-1.5, AC-2.1 |
| PN-8: the kind's gone line from one L10n function; a section's gone notice has no Try again | STEP-3, STEP-5, STEP-6, AC-3.1, AC-3.2, AC-3.3 |
| PN-9 engine half: a not-counted read sends no pair | STEP-8, AC-4.1 |
| PN-9 app half: `PersonRecordLinksContent.partial`, the uncounted spoken label, the column's note in place of the caption | STEP-9, AC-4.1, AC-4.2, AC-4.3, AC-4.4, AC-4.5 |
| PN-W1, PN-W2: notices keep the band's, the header's and the column's widths; state survives the Regular/Wide line | STEP-5, STEP-6, STEP-9, manual check |
| P-5's FakeStore test and the previews that show each failure state | STEP-7, T-5, T-6 |
| `add-workspace-place` and `macos-client-patterns.md` teach the reading | STEP-10 |

| Out | Home |
| --- | --- |
| The Conclusion lists' silent refresh failure | [idea](../../../ideas/person-page-followups.md) (its own brief) |
| `PromoteModel.loadAttempt` onto the kit's failure revision | not planned (Promote's read is a direct store call with no handle; it keeps its `.id` and its focus on appear) |
| Moving `SourcesModel` and `SourceTypesModel` off `QueryHandle.failure` | not planned (`failure` stays the primitive `reading` is built from; those lists show a failure in place of their state, not beside a value) |
| What Remove does on a record line under a stale notice (a second Remove is a silent no-op) | To user (review log B-2); recommended home [idea](../../../ideas/person-page-followups.md) |
| A notice for a failed Source types (supporting) read | not planned (the labels drop out, as today; brief §2.1) |
| Stale-while-refreshing of hidden keys, the 50-entity cap | S16-09 |
| Section-switch and Remove announcements; "records" in the header count and confirm | S16-06 |
| Back restoring a section's place (P-19) | [idea](../../../ideas/person-page-followups.md) |
| "Read again" on the Evidence graph's and the Source form's retry vs "Try again" | not planned (each feature keeps its own label) |
| Any data about a line, a new arc stroke style | not planned (Spike 13 rule "The view is a circular link diagram") |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Handle state | `macos/App/Features/Workspace/Session/QueryHandle.swift` `QueryRead`, `QueryHandle.read`, `failure`, `applyFailure(_:clearValue:)`, `beginLoading(keepingValue:)` | `read` folds a value first, even beside a failure. `failure` is the error while `status == .error`, even beside a value. `applyFailure` keeps a shown value (`clearValue: false`) and bumps `revision` only when it clears; it is called only from `WorkspaceSession.ensureQuery`. `beginLoading(keepingValue: false)` sets `.loading` and leaves `error` set, so during a retry after a first-load failure `read` is `.loading` |
| Session | `macos/App/Features/Workspace/Session/WorkspaceSession.swift` `ensureQuery` (line 422), `reread(_:)`, `isRefreshing(_:)` | A failed load keeps a value it had (`clearValue: !hadStaleValue`). `reread` invalidates then queries and keeps the stale value. `isRefreshing` is `hasValue && isFetching` |
| Page presentation | `macos/App/Features/Conclusions/ConclusionDetailPage.swift` `ConclusionDetailPresentation` (lines 6–31) | Built from `read`: `.loading`, `.gone` (engine code `conclusionNotFound`, only when no value), `.failed(String)`, `.content(_, refreshing:)`. A refresh failure is invisible; a not-found reload behind a value keeps the stale page |
| Page | same file, `ConclusionDetailPage`, the private `ConclusionDetailLoaded` (from line 237), `ConclusionDetailHeaderBand` (from line 330) | `.failed` draws a plain danger `PVCallout` with `errorIdentifier`; gone leaves with `L10n.Errors.conclusionDetailsNotFound` through `leavesForSectionRoot(when:notice:)`. The header band draws the header and, on a page with sections, the switcher and "Refreshing". `ConclusionDetailLoaded` reads `navigation.currentLocation.section` for the offered sections |
| Section host | `macos/App/Features/Conclusions/ConclusionPageSectionHost.swift` `ConclusionPageSectionHost` | Builds `ConclusionDetailPresentation` from `handle.read`; `.failed` and `.gone` draw `PVCallout(tone: .danger, title: errorTitle, …)` (compact in `.column`) with id `\(idRoot).error`; `.gone` uses the shared not-found line; content draws with no failure shown. It reads `WorkspaceSession` from the environment, not `WorkspaceNavigation` |
| Sources | `macos/App/Features/Conclusions/PersonSourcesSection.swift` `PersonSourcesSection`, `PersonSourcesList.diagram` | The record links host is built inside the rows host's content, so no column draws when the rows fail, are gone or are empty |
| Retry kit piece | `macos/App/DesignSystem/Components/Callout/PVRetryCallout.swift` `PVRetryCallout` | Danger `PVCallout(announces: true)` with one small secondary `PVButton` (loading state), `compact`, optional `retryFocused` binding. Callers: `PromoteView.notices` (`.id(model.loadAttempt)`, focus on every appear), `EvidenceGraphView.readErrorCallout` (`.id(model.graphReadAttempt)`), `SourceFormLoadArea` |
| Evidence graph | `macos/App/Features/EvidenceGraph/EvidenceGraphModel.swift` `graphReadError`, `isRereadingGraph`, `graphReadAttempt`, `readGraphAgain`, `graphReadFailedSettled` | A model counter bumped by a `Task` that waits on `session.readyValue` after each reread, so the callout remounts and speaks again. `isRereadingGraph` is the handle's `isFetching` (false during a retry after a first-load failure). Tests: `macos/ProvenenciaTests/EvidenceGraphModelTests.swift` (around line 974) |
| Announce | `macos/App/DesignSystem/Components/Announce/PVAnnounce.swift` | Speaks on appear and when the text changes; the same text again while mounted is not spoken |
| Focus | `macos/App/Platform/FocusClaim.swift` `FocusClaim.claim` | Writes a focus binding next turn and again a turn later |
| Record links, engine | `core/database/persondetail/recordlinks.go` `PersonRecordLinks`, its read (lines 55–71); `core/recordlinks/recordlinks.go` `Pairs` | When the read passes `MaxReadWork`, `Pairs` lists every pair with 0-fact (weak) lines, or none when the lines alone pass it; `persondetail` caps the lines (`CapLines`) and sends them with `NotCounted: true`. Test file: `core/database/persondetail/recordlinks_test.go` |
| Record links, wire | `api/proto/engine.proto` `GetPersonRecordLinksResponse` comment; `macos/App/Platform/GenealogyStore.swift` `CatalogPersonRecordLinks.notCounted`, `.linesDropped` | The comment says "every line weak, or no pair"; nothing in Swift reads the flags |
| Record links, app | `macos/App/Features/Conclusions/PersonRecordLinksContent.swift` `init(links:graphLocations:graphSourceRefs:)`, `maxNodes`; `PersonRecordLinksColumn.swift`; previews `PersonRecordLinksPreview` (diagram only) | Keeps the first 1,000 distinct graphs silently; draws every sent line, weak at `weakLineWidth`; a node with no neighbour is spoken with `recordLinksNodeApartAccessibility` ("…, shares no one"). The column draws title, caption (hidden for one graph), diagram, one-record line |
| Preview data | `macos/App/Features/Conclusions/PersonSourcesPreviewData.swift` `manyRecordLinks(graphs:)`, `householdRecordLinks(graphs:)`, `johannRecordLinks` | Build `CatalogPersonRecordLinks` values; `notCounted` and `linesDropped` are `var`s a preview can set |
| Strings | `macos/App/Platform/L10n.swift` `L10n.Conclusions.familyErrorTitle`, `timelineErrorTitle`, `sourcesErrorTitle`, `recordLinksErrorTitle`, `recordLinksCaption`, `recordLinksNodeApartAccessibility`; `L10n.Errors.conclusionDetailsNotFound`; `L10n.Unpromote.confirmMessageDeletes(kind:ref:)` (the per-kind shape) | Section titles stay; no Try again, stale, page title, gone-per-kind or partial-note keys exist |
| FakeStore | `macos/App/Platform/FakeStore.swift` `getPersonRecordLinksError`, `listPersonClaims`, `getPersonFamily`, `getConclusionDetail` | Only the record links read has an error hook; the others throw `conclusiondetails.not_found` for an unseeded id |
| Tests | `macos/ProvenenciaTests/WorkspaceSessionTests.swift` `queryReadAndProgressReads`, `applyPersonPageKeepsTheOpenSectionsReadsVisible`; `PersonDetailTests.swift` `presentationFollowsTheHandle`; `PersonRecordLinksContentTests.swift` `allLinesDrawAndNodesStopAtTheCap`, `aNodeThatSharesNoOneIsSpokenAsApart`, `goStoreCarriesThePartialFlags` | Cover `read`, the presentation's four cases, visible section keys, the node cap, the apart label and the flags' mapping. No UI test target exists |

**Precedent.** `EvidenceGraphView.readErrorCallout` (S14-20): a `PVRetryCallout` remounted with `.id(attempt)` so it speaks again, Try again through `session.reread`. This PR keeps the remount and the reread and moves the remount into the kit on the handle's `failureRevision`. For the kind-keyed copy, copy `L10n.Unpromote.confirmMessageDeletes(kind:ref:)`.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | `QueryReading<Value>` *(new)* in `QueryHandle.swift`: `.loading`, `.failed(Error)`, `.loaded(Value, refreshFailure: Error?)`. `QueryHandle.reading` *(new)*: a value → `.loaded(value, refreshFailure: failure)`; else an `error` → `.failed(error)`, also while a retry after a first-load failure runs; else `.loading`. `read` and `failure` are not changed; `reading` is built from `failure`, so there is one notion of "failed beside a value" | Working rule: one reading, `read` unchanged for its other callers. Holding `.failed` during the retry keeps the notice and its Try again on screen, so its loading state shows and the keyboard has somewhere to stay (PN-6) | A case added to `QueryRead`: every list and Evidence graph caller would have to handle it. Deleting `failure`: `SourcesModel` and `SourceTypesModel` read it for list states this PR does not touch (review M-1) |
| D-2 | `QueryHandle.failureRevision` *(new)*, `private(set) var Int`, bumps in every `applyFailure`. `QueryHandle.isReading` *(new)*: `status == .loading || isFetching` ("a first load or a refetch is running") | The one counter that makes a notice speak again; it replaces `EvidenceGraphModel.graphReadAttempt` and its wait `Task` (D-5). `isReading` is the retry Button's loading state in the first-load and the refresh case | A counter per model (the S14-20 shape): a second mechanism beside the handle's |
| D-3 | `ConclusionDetailPresentation` is built from `QueryReading`: `init(_ reading:isFetching:content:)`. `.content` gains `refreshFailure: String?` (the message from `L10n.Errors.message(for:)`). A `conclusiondetails.not_found` refresh failure behind a value is `.gone`, like a first-load one | The page and the host draw from one presentation (S9-24). A person gone behind a shown value must leave too; S16-09's scope reads "a not-found reload over a stale value reads as gone (S16-04's reading)" | A separate `.refreshFailed` case: the content still draws, so it is a property of `.content` |
| D-4 | `PVRetryCallout` gains `failureRevision: Int? = nil` *(new parameter)*. When it is set: the inner `PVCallout` carries `.id(failureRevision)` so each new failure speaks again; the callout keeps `@State retryPressed` and an internal `@FocusState` (used when the caller passes no `retryFocused`); the Button's action sets `retryPressed = true` then calls `onRetry`; on the inner callout's appear, when `retryPressed`, it calls `FocusClaim.claim` on the Try again Button and clears `retryPressed` in `then:`. When `failureRevision` is nil, the callout behaves as today. Callers: the page, the section host and `EvidenceGraphView.readErrorCallout` pass it; Promote and `SourceFormLoadArea` do not | Review M-2: the speak-again rule had two copies (`PromoteView`, `EvidenceGraphView`) and this PR would add a third; the kit owns it, and S15-21's provenance sheet (a recipe, which may not name a `Features/` type, `docs/design-system-layers.md`) gets it by passing one value. Clearing `retryPressed` after the claim means a later failure that arrives by a write does not take the keyboard (PN-7, review X-9) | A feature wrapper `ConclusionReadNotice` (round 1): a recipe could not use it. A model counter: see D-2 |
| D-5 | `EvidenceGraphView.readErrorCallout` passes `failureRevision` and `isRetrying` from the graph rows' handle (through `EvidenceGraphModel`: `isRereadingGraph` returns the handle's `isReading`; a new `graphReadFailureRevision` returns its `failureRevision`) and drops its `.id`. `EvidenceGraphModel.graphReadAttempt`, `graphReadFailedSettled` and the wait `Task` in `readGraphAgain` go; `readGraphAgain` is the `reread` alone | One mechanism (review M-1). The Evidence graph gains a loading Button during a first-load retry | Leaving the Evidence graph on its counter: two mechanisms after this PR |
| D-6 | The page's notice goes in `ConclusionDetailHeaderBand`, after the switcher (or after the header on a page without sections), from `.content`'s `refreshFailure`: `PVRetryCallout(title: L10n.Conclusions.pageReadErrorTitle, message: L10n.Conclusions.readNoticeStaleMessage(refreshFailure), retry: L10n.Conclusions.readNoticeTryAgain, isRetrying: handle.isReading, retryIdentifier: "\(pageIdentifier).refreshError.tryAgain", failureRevision: handle.failureRevision, onRetry:)`, id `\(pageIdentifier).refreshError`. `ConclusionDetailLoaded` builds `onRetry` as `{ let _: QueryHandle<CatalogConclusionDetail> = session.reread(key) }` and passes it, `isRetrying` and `failureRevision` to the band. `.failed` draws the same callout with the bare message in `PVPageColumns(primary:)`, id `errorIdentifier`, Try again id `\(errorIdentifier).tryAgain` | PN-3; the page's read covers the header, Vitals and the Gallery, so the band that holds the header holds its notice, on every section | A notice above Vitals only: the header would read stale under Family with no word |
| D-7 | The gone line is `L10n.Conclusions.goneNotice(section: WorkspaceSection)` *(new)*: `.persons` → `conclusions.person.gone`; every other section → `L10n.Errors.conclusionDetailsNotFound`. The page calls it with `navigation.currentLocation.section` (it already reads it for the offered sections); the host gains `@Environment(WorkspaceNavigation.self)` and calls it the same way | Working rule "A person's records are records": the kind supplies the line. Review M-5: one place to add a kind's line, the shape of `confirmMessageDeletes(kind:)`, and no parameter through six types | A `goneNotice` parameter through the page, the context, three sections, `PersonSourcesList` and the host (round 1) |
| D-8 | The host's states: `.loading` → today's spinner; `.failed` → `PVRetryCallout` with the bare message in place of the content, id `\(idRoot).error`; `.gone` → `PVCallout(tone: .danger, title: errorTitle, message: goneNotice)` with no button, in place of the content, id `\(idRoot).error`; `.content` with a `refreshFailure` → `PVRetryCallout` with the stale message above the content, or above the empty state in `.band`, id `\(idRoot).refreshError`. Every retry callout: `retry: L10n.Conclusions.readNoticeTryAgain`, `isRetrying: handle.isReading`, `failureRevision: handle.failureRevision`, `retryIdentifier: "\(id).tryAgain"`, `compact: true` in `.column`, `onRetry` = `session.reread(key)` with the host's own key (its role's entry in `ConclusionPageSection.reads`). In `.band` each callout takes `.pvKeepsPageBand()`; the stale stack is a `VStack(alignment: .leading)` with `PVSpacing.space7` in `.band` and `PVSpacing.space4` in `.column` (the column's own spacing) | PN-1, PN-2, PN-4, PN-5, PN-8, PN-10. A gone read would fail the same way on retry. `.refreshError` vs `.error` lets a test tell a stale notice from a failed first load (review X-10) | Try again rereading every read of the section: the band and the column fail apart and each notice reports one read (brief PN-6) |
| D-9 | Engine: in `persondetail` (`recordlinks.go`, the read at lines 55–71), when `counted` is false the pairs are dropped before `CapLines`, so a not-counted response holds the graphs, no pair, `NotCounted: true`, `LinesDropped: false`. The `PersonRecordLinks.NotCounted` comment and the `GetPersonRecordLinksResponse` comment in `api/proto/engine.proto` say "no pair is sent". No proto field changes | Review M-3: the engine owns what "not compared" means; a client decision keyed to an engine flag is a Known trap. The second not-counted shape already sends no pair, so both look the same | Dropping the arcs in Swift (round 1) |
| D-10 | `PersonRecordLinksContent.Partial` *(new)* enum `.notCounted`, `.leftOut`; `partial: Partial?` *(new)*. `.notCounted` when `links.notCounted`; else `.leftOut` when `links.linesDropped` or the graph loop met a graph id not yet kept while `graphs.count == maxNodes` (a distinct graph past the cap); else nil. With `links.notCounted` each node's accessibility label is `recordLinksNodeUncountedAccessibility(ref:graph:source:)` *(new)*, never the apart or connected label. The arcs are what Go sent (none when not counted, D-9) | PN-9; working rule: one note, whichever side cut it; the 1,000-node cut is Swift's, so Swift adds it | A flag per cut on the view: the rule asks for one note |
| D-11 | `PersonRecordLinksColumn` shows, when `content.partial != nil`, `PVCallout(tone: .warning, message: L10n.string(note), compact: true)` with no title, not announced, in place of the caption (also for a single graph), id `\(accessibilityIdentifierRoot).recordLinks.partial`; `note` is `recordLinksNotCounted` for `.notCounted` and `recordLinksPartial` for `.leftOut` | PN-9; brief inventory: Callout, warning, compact | A muted caption line: the brief picks the kit's hint piece |
| D-12 | Copy: the stale message is one format key, `conclusions.readNotice.staleMessage` "%@ What is shown is from the last time it was read."; Try again is `conclusions.readNotice.tryAgain` "Try again" | One sentence per language, word order free; a Conclusions key, not Promote's (`add-localized-string`: each feature owns its copy) | Concatenating two strings in Swift |
| D-13 | FakeStore gains `getConclusionDetailError`, `getPersonFamilyError` and `listPersonClaimsError` *(new)*, each `Error?`, each thrown at the top of its read before the seeded lookup, the shape of `getPersonRecordLinksError` | P-5's test needs a generic failure behind a seeded value; the previews (D-14) need a failing page, Family, Sources and record links read. Removing a seed throws not-found, which is gone, not a failure | One dictionary of failing reads: a second hook shape beside `getPersonRecordLinksError` |
| D-14 | `PersonDetailNoticePreviews.swift` *(new)*, `#if DEBUG`, in `macos/App/Features/Conclusions/`: `#Preview`s of `PersonDetailView` on a `WorkspaceSession(projectKey:store: FakeStore())` (the setup `ConclusionPictureViewer`'s preview uses) with `WorkspaceNavigation` in the environment, for: "Sources stale" (claims seeded, loaded, then `listPersonClaimsError` set and the key reread), "Record links failed", "Page failed", "Family failed", "Sources gone" (detail seeded, claims not seeded) and "Timeline, page stale". Each preview builds its state in a `.task` | The manual check cannot force a read to fail on a real project; previews show every failure state at Regular and Wide (review X-4) | A Debug launch flag for a FakeStore app: a new mechanism for a check |

**Rules this plan bends.** None.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `QueryReading` *(new)*, `QueryHandle.reading`, `.failureRevision`, `.isReading` | type, properties | `QueryRead`, `.read`, `.failure`, `.isFetching`, `EvidenceGraphModel.graphReadAttempt` | D-1, D-2: `reading` is built from `failure`; `failureRevision` replaces the Evidence graph's counter (D-5) |
| `PVRetryCallout(failureRevision:)` | parameter | `.id(attempt)` at `PromoteView` and `EvidenceGraphView` | Extended, D-4 |
| `EvidenceGraphModel.graphReadFailureRevision` | property | `graphReadAttempt` | Replaces it, D-5 |
| `ConclusionDetailPresentation.content(_, refreshing:, refreshFailure:)` | case change | today's `.content(_, refreshing:)` | Extended, D-3 |
| `L10n.Conclusions.goneNotice(section:)` | function | `L10n.Unpromote.confirmMessageDeletes(kind:ref:)` | Same shape, D-7 |
| `PersonRecordLinksContent.Partial`, `.partial` | type, property | `CatalogPersonRecordLinks.notCounted`, `.linesDropped` | Adds Swift's own 1,000-node cut (D-10) |
| `L10n.Conclusions.readNoticeTryAgain`, `readNoticeStaleMessage(_:)`, `pageReadErrorTitle`, key `conclusions.person.gone`, `recordLinksNotCounted`, `recordLinksPartial`, `recordLinksNodeUncountedAccessibility(ref:graph:source:)` | keys | `promote.error.tryAgain`, `error.conclusiondetails.not_found`, `recordLinksNodeApartAccessibility` | D-12; Conclusions copy lives in `L10n.Conclusions` |
| `FakeStore.getConclusionDetailError`, `getPersonFamilyError`, `listPersonClaimsError` | test hooks | `getPersonRecordLinksError` | Same shape, D-13 |
| `PersonDetailNoticePreviews.swift` *(new)* | previews | `ConclusionPictureViewer`'s `#Preview` session | D-14 |

**Builds on.** `PVRetryCallout` (kit); `session.reread`; `ConclusionPageSection.reads` (S16-03); `leavesForSectionRoot(when:notice:)`; `FocusClaim`; the S14-20 remount; `confirmMessageDeletes(kind:)`.
**Deletes.** `EvidenceGraphModel.graphReadAttempt`, `graphReadFailedSettled`, the wait `Task` in `readGraphAgain`, and `EvidenceGraphView.readErrorCallout`'s `.id(model.graphReadAttempt)` (D-5). The plain `PVCallout` failure branch of `ConclusionDetailLoaded` and of `ConclusionPageSectionHost.error(_:)` (D-6, D-8). The pairs a not-counted read used to send (D-9). `ConclusionPageSectionHost.Placement.column`'s doc line "a compact danger Callout" is reworded to the retry callout. No key goes: `L10n.Errors.conclusionDetailsNotFound` stays for Event, Place and the errors map.
**Next adds.** S16-09 reads D-3's "not-found behind a value is gone". S15-21's sheet key reads `QueryReading` and passes `failureRevision` to `PVRetryCallout`. The lists' later brief builds `ConclusionListPresentation` from `reading`. A section that gains a read adds a `reads` entry; its host draws the notice with no new code. A kind that gains its own gone line adds a case to `goneNotice(section:)`.

### Reach

| New path or state | Destination | Checked: resets on the new id · handles a missing referent · shows the state |
| --- | --- | --- |
| A page whose detail reload answers not-found behind a shown value now leaves | `ConclusionDetailLoaded` `.leavesForSectionRoot` | `onChange(of: gone, initial: true)` fires once on the change; the page is keyed by entity id, so a new page starts clean |
| A section shows content and a notice together | `ConclusionPageSectionHost` `.content` | Folds, focus and hover live in the section views, which stay mounted under the notice; the notice is a sibling above them |
| Content with a refresh failure and an empty value | host `.band` empty state | Notice above `PVEmptyState` (D-8) |
| A not-counted response with graphs and no pair | `PersonRecordLinksContent`, `CircularLinkDiagram` | The diagram already draws nodes with no arc (`aNodeThatSharesNoOneIsSpokenAsApart`); the label is the uncounted one (D-10) |
| An old app on a new engine | `PersonRecordLinksContent` before this PR | Draws nodes, no arc, no note: the same as today's second not-counted shape |

### Scale

At the map's Scale (a person in 50 Evidence graphs): no new read. A failure adds one notice per failed read; Try again issues one `reread` of one key. A not-counted record links read sends no pair, so it marshals less than today (up to 75,000 lines before). Nothing loops per row.

### Edges

- **Assumptions.** `applyFailure` is the only place `status` becomes `.error` (called only from `ensureQuery`), and every success clears `error` (`applySuccess`, `resetForInvalidation`), so `reading` never reports an old failure after a success (T-1). `beginLoading(keepingValue: false)` leaves `error` set: D-1 relies on it to keep `.failed` during a first-load retry; `resetForInvalidation` and eviction clear it, so a key evicted and queried again starts at `.loading` (T-1).
- **Neighbors.** A refresh failure that is not-found → `.gone`; any other code → content plus notice (T-2). A first-load not-found → `.gone` as today (T-2). The page read and a section read failing together → two notices, each with its own Try again (manual check 4). Both Sources reads failing → the band and the column notice, independent (AC-1.5). A counted read with no pair (records that share no one) keeps the apart label (T-8).
- **Removed inputs.** A person merged away while Sources is open: the section's read may say not-found first (gone notice, AC-3.2), then the page's read leaves (AC-3.1). A notice whose read later succeeds through a write goes with the failure (`applySuccess` clears `error`, T-1). A pressed Try again that succeeds unmounts the callout, so its `retryPressed` goes with it (D-4).
- **Real data.** None: no table or seed drives a rule here.
- **Untrusted size.** `links.graphs` is bounded by the existing `maxNodes` loop; `partial` is decided in the same loop with no extra set (D-10).

### Schema

None — no catalog change.

### Go core

| Package | Change |
| --- | --- |
| `core/database/persondetail` | The record links read drops the pairs when `counted` is false, before `CapLines` (D-9); `PersonRecordLinks.NotCounted`'s comment updated |

Errors: none new.

### FFI

None — no field changes. `api/proto/engine.proto` `GetPersonRecordLinksResponse`'s comment only (D-9). Run `./scripts/generate-proto.sh` so the generated files carry the new comment.

### macOS

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Read notice (PN-1 – PN-7) | `PVRetryCallout`, Extend | `macos/App/DesignSystem/Components/Callout/PVRetryCallout.swift` | D-4; full in the header band and a section band, `compact` in the column |
| Gone notice in a section (PN-8) | `PVCallout(tone: .danger)` | `macos/App/Features/Conclusions/ConclusionPageSectionHost.swift` | D-8; no button |
| Partial note (PN-9) | `PVCallout(tone: .warning, compact: true)` | `macos/App/Features/Conclusions/PersonRecordLinksColumn.swift` | D-11; not announced |
| Page frame | snowflake, Extend | `macos/App/Features/Conclusions/ConclusionDetailPage.swift` | D-6, D-7 |
| Section host | snowflake, Extend | `macos/App/Features/Conclusions/ConclusionPageSectionHost.swift` | D-7, D-8 |
| Leave Toast | `leavesForSectionRoot` | `macos/App/Features/Workspace/WorkspaceNavigation.swift` | Only the words change (D-7) |

The frames in words (from the brief, until the board is agreed):

- **Header band, refresh failure.** Header, switcher with "Refreshing" when it runs, then the notice at the header's width. Event and Place: header, then the notice, then Details.
- **Page, first-load failure.** No header; the notice alone at the top of the page column.
- **Section band.** The notice first, at the page band's width even when the section fills the pane at Wide (Timeline, Sources), then the unchanged rows, tiles, lifeline or empty state. First-load failure: the notice alone.
- **Diagram column.** Compact notice above the title; the column's top aligns with the rows' top. First-load failure: the compact notice is all the column shows. Partial: title, then the warning note in place of the caption, then the circle. When the rows failed their first load or are gone, there is no column.
- **Gone in a section.** Danger Callout, section title, the person's gone line, no button, no rows.

Accessibility: each notice speaks when it appears and on each new failure (danger priority); a notice takes the keyboard only when its own Try again was pressed and the read failed again; the partial note is read in order before the diagram.

### Strings

L10n keys *(new)*, all under `L10n.Conclusions`, in `macos/App/Resources/Localizable.xcstrings`:

- `conclusions.readNotice.tryAgain` = "Try again"
- `conclusions.readNotice.staleMessage` = "%@ What is shown is from the last time it was read." (argument: the error text)
- `conclusions.detail.errorTitle` = "This page could not be read"
- `conclusions.person.gone` = "This person no longer exists. They may have been merged into another person." (returned by `goneNotice(section: .persons)`)
- `conclusions.recordLinks.notCounted` = "Too many records to compare. The links between them are not drawn."
- `conclusions.recordLinks.partial` = "Too many records to draw them all. Some records and links are left out."
- `conclusions.recordLinks.nodeUncountedAccessibility` = "%1$@, %2$@, %3$@" (ref, graph, Source)

Removed: none. Old-scope copy: `error.conclusiondetails.not_found` ("This record no longer exists. It may have been merged into another.") stays for Event, Place and the engine errors map; the Person page stops showing it (D-7).

## 7. Steps

**STEP-1 — The handle's reading and failure revision**
Files: `macos/App/Features/Workspace/Session/QueryHandle.swift`, `macos/ProvenenciaTests/WorkspaceSessionTests.swift`. Skill: `add-swift-test`.
Do: add `QueryReading`, `reading`, `failureRevision` (bumped in `applyFailure`), `isReading` (D-1, D-2), each with a doc comment. Do not touch `read`, `failure` or any caller. Write T-1.
Check: `xcodebuild test -project macos/Provenencia.xcodeproj -scheme Provenencia -destination 'platform=macOS' -only-testing:ProvenenciaTests/WorkspaceSessionTests` passes, `queryReadAndProgressReads` unchanged.

**STEP-2 — The presentation reads it**
Files: `macos/App/Features/Conclusions/ConclusionDetailPage.swift` (`ConclusionDetailPresentation` and its use in `ConclusionDetailLoaded`), `macos/App/Features/Conclusions/ConclusionPageSectionHost.swift` (its `loaded`), `macos/ProvenenciaTests/PersonDetailTests.swift`.
Do: D-3. Both builders pass `handle.reading`. Views keep drawing what they draw today; `refreshFailure` is not drawn yet. Update `presentationFollowsTheHandle` to the new init and add T-2.
Check: `-only-testing:ProvenenciaTests/PersonDetailTests` passes; the app builds.

**STEP-3 — Copy**
Files: `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`, `macos/ProvenenciaTests/PersonDetailTests.swift`. Skill: `add-localized-string`.
Do: add the keys in **Strings** with comments, and `goneNotice(section:)` (D-7). Write T-3 and T-4.
Check: `python3 scripts/check-localizable-xcstrings.py` passes; `-only-testing:ProvenenciaTests/PersonDetailTests` passes.

**STEP-4 — The kit's retry callout speaks again and keeps the keyboard; the Evidence graph uses it**
Files: `macos/App/DesignSystem/Components/Callout/PVRetryCallout.swift`, `macos/App/Features/EvidenceGraph/EvidenceGraphView.swift`, `macos/App/Features/EvidenceGraph/EvidenceGraphModel.swift`, `macos/ProvenenciaTests/EvidenceGraphModelTests.swift`. Skill: `add-ui-component`.
Do: D-4, D-5. Add a `#Preview` that passes `failureRevision`. Update the `isRereadingGraph` tests around line 974 so they pass with `isReading` (T-10).
Check: `-only-testing:ProvenenciaTests/EvidenceGraphModelTests` passes; `rg "graphReadAttempt|graphReadFailedSettled" macos` finds nothing.

**STEP-5 — The page's notice and gone line**
Files: `macos/App/Features/Conclusions/ConclusionDetailPage.swift` (`ConclusionDetailLoaded`, `ConclusionDetailHeaderBand`).
Do: D-6, D-7. `ConclusionDetailHeaderBand` gains `refreshFailure: String?`, `isRetrying: Bool`, `failureRevision: Int`, `retry: () -> Void` and draws the callout last in its `VStack`. `.failed` draws the first-load callout. `leavesForSectionRoot` takes `L10n.string(L10n.Conclusions.goneNotice(section: navigation.currentLocation.section))`. No change to `ConclusionDetailPage`'s initializers or to the Person, Event and Place views.
Check: the app builds; `-only-testing:ProvenenciaTests/PersonDetailTests` passes.

**STEP-6 — The section host's notices**
Files: `macos/App/Features/Conclusions/ConclusionPageSectionHost.swift`.
Do: D-7, D-8. Add `@Environment(WorkspaceNavigation.self)`; replace `error(_:)` with the three states of D-8; draw the stale stack in `.content`. The section views do not change.
Check: the app builds; `-only-testing:ProvenenciaTests/PersonDetailTests` passes.

**STEP-7 — P-5's test and the failure previews**
Files: `macos/App/Platform/FakeStore.swift`, `macos/ProvenenciaTests/WorkspaceSessionTests.swift`, `macos/App/Features/Conclusions/PersonDetailNoticePreviews.swift` *(new)*.
Do: D-13, D-14; write T-5 and T-6.
Check: `-only-testing:ProvenenciaTests/WorkspaceSessionTests` passes; each preview renders its state in Xcode's canvas.

**STEP-8 — A not-counted read sends no lines**
Files: `core/database/persondetail/recordlinks.go`, `core/database/persondetail/recordlinks_test.go`, `api/proto/engine.proto` (comment). Rule: `.cursor/rules/go-tests.mdc`.
Do: D-9; write T-9; run `./scripts/generate-proto.sh`.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/persondetail/... ./api/...` passes.

**STEP-9 — The partial diagram**
Files: `macos/App/Features/Conclusions/PersonRecordLinksContent.swift`, `macos/App/Features/Conclusions/PersonRecordLinksColumn.swift`, `macos/App/Platform/GenealogyStore.swift` (the `notCounted` comment: "no pair is sent"), `macos/ProvenenciaTests/PersonRecordLinksContentTests.swift`.
Do: D-10, D-11. Update `allLinesDrawAndNodesStopAtTheCap` to also expect `.leftOut`; write T-7. Add four `#Preview`s of `PersonRecordLinksColumn` (not the diagram-only `PersonRecordLinksPreview`), each in a 400 pt wide frame with `.environment(\.pvPageViewportSize, CGSize(width: 1102, height: 900))`, a `CircularLinkHover<String>()` in `@State`, `focusedGraphID: nil`, `open: { _ in }` and the existing preview's `focus` closure: "Not counted" (`manyRecordLinks(graphs: 40)` with its `pairs` emptied and `notCounted = true`), "Lines dropped" (`manyRecordLinks(graphs: 50)` with `linesDropped = true`), "1,001 graphs" (`manyRecordLinks(graphs: 1_001)`), "Both" (`manyRecordLinks(graphs: 1_001)` with pairs emptied and `notCounted = true`).
Check: `-only-testing:ProvenenciaTests/PersonRecordLinksContentTests` passes.

**STEP-10 — Docs and version**
Files: `.agents/skills/add-workspace-place/SKILL.md` ("Stale-while-revalidate": read `reading` to show a refresh failure beside a value, and pass `failureRevision` and `isReading` to `PVRetryCallout`), `docs/macos-client-patterns.md` (the `QueryHandle` row: `read` for a reader that shows only a value or a failure; `reading` for one that shows a refresh failure beside its value; `failureRevision`; `isReading` = a first load or a refetch is running), `VERSION`, `core/version.go`, `MARKETING_VERSION` in `macos/Provenencia.xcodeproj/project.pbxproj` (both lines), `docs/deployment-plan/spike-16/completed.md`, `docs/deployment-plan/spike-16/deployment-plan.md` (map row), this plan, `docs/deployment-plan/spike-16/design/S16-D1-person-page-notices.md` → `docs/deployment-plan/spike-16/design/archive/`, `docs/deployment-plan/spike-16/design/README.md`.
Do: per the Definition of done.
Check: `scripts/check-pr.sh` prints `check-pr: PASS`; links resolve.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Show a Try again notice when a Person page section fails to reload.").

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Swift | `WorkspaceSessionTests.swift` `queryReadingCarriesARefreshFailure` *(new)* | new handle → `.loading`; `applyFailure(clearValue: true)` → `.failed`, `failureRevision` 1; `beginLoading(keepingValue: false)` → still `.failed`, `isReading`; `applySuccess(3)` → `.loaded(3, nil)`; `applyFailure(clearValue: false)` twice → `.loaded(3, refreshFailure: non-nil)`, `failureRevision` 3; `applySuccess(4)` → `.loaded(4, nil)`; `resetForInvalidation` → `.loading`; `read` unchanged at every step (`.loading` during the first-load retry) | AC-1.1, AC-1.3, AC-2.2, D-1, D-2 |
| T-2 | Swift | `PersonDetailTests.swift` `presentationFollowsTheHandle` (updated) + `presentationShowsARefreshFailureBesideTheContent` *(new)* | `.loaded(detail, refreshFailure: generic)` → `.content(_, refreshing:, refreshFailure: message)`; `.loaded(detail, refreshFailure: not_found)` → `.gone`; `.failed(not_found)` → `.gone`; `.loaded(detail, nil)` → `refreshFailure` nil | AC-1.1, AC-3.1, D-3 |
| T-3 | Swift | `PersonDetailTests.swift` `eachKindSuppliesItsGoneLine` *(new)* | `goneNotice(section: .persons)` resolves to "This person no longer exists. They may have been merged into another person."; `.events` and `.places` resolve to `L10n.Errors.conclusionDetailsNotFound` | AC-3.1, AC-3.3, D-7 |
| T-4 | Swift | `PersonDetailTests.swift` `readNoticeStaleMessageWrapsTheError` *(new)* | `readNoticeStaleMessage("Disk full.")` == "Disk full. What is shown is from the last time it was read." | AC-1.1, D-12 |
| T-5 | Swift | `WorkspaceSessionTests.swift` `aFailedSectionReloadIsReadBesideItsRows` *(new)* | FakeStore seeds claims for "per-1"; `apply(location:)` the Person page on Sources; query and wait; set `listPersonClaimsError`; `apply(CatalogWriteEffects(handles: ["per-1"]))`; `settleWrites` → the claims handle's `reading` is `.loaded(seeded, refreshFailure: non-nil)` and `failureRevision` is 1; clear the error, `reread`, wait → `.loaded(_, nil)` | AC-1.1, AC-1.2, P-5 |
| T-6 | Swift | `WorkspaceSessionTests.swift` `aSecondFailureBumpsTheFailureRevision` *(new)* | T-5's state, `reread` with the error still set, wait → `failureRevision` 2, value kept | AC-1.3, D-2 |
| T-7 | Swift | `PersonRecordLinksContentTests.swift` `aReadThatDidNotCountIsPartialAndSpokenWithoutLinks`, `droppedLinesAreLeftOut`, `droppedGraphsAreLeftOut`, `notCountedWinsOverLeftOut`, `aFullReadIsNotPartial` *(new)* | AC-4.1 (graphs, no pairs, `notCounted` → `.notCounted`, every label the uncounted one), AC-4.2, AC-4.3, AC-4.4, AC-4.5 inputs, one test each in that order | AC-4.1, AC-4.2, AC-4.3, AC-4.4, AC-4.5, D-10 |
| T-8 | Swift | `PersonRecordLinksContentTests.swift` `aNodeThatSharesNoOneIsSpokenAsApart` (unchanged) | A counted read with no pair still speaks "shares no one" | D-10 |
| T-9 | Go | `core/database/persondetail/recordlinks_test.go`, a new case in its table *(new)* | A read past `MaxReadWork` (lowered through the test's config) with pairs found → `Pairs` empty, `NotCounted` true, `LinesDropped` false, `Graphs` all listed | AC-4.1, D-9 |
| T-10 | Swift | `EvidenceGraphModelTests.swift` (the `isRereadingGraph` tests, updated) | After a first-load failure and `readGraphAgain`, `isRereadingGraph` is true until the read settles | D-5 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `PersonDetailTests`, `WorkspaceSessionTests` (`queryReadAndProgressReads` must still pass unchanged), `PersonRecordLinksContentTests`, `EvidenceGraphModelTests`, `core/database/persondetail` (`recordlinks_perf_test.go`), the string catalog check.

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left
- [ ] Sweep: each renamed or deleted name `rg`'d by its old, qualified name across `.agents/`, `.cursor/rules/`, `docs/*.md` and `L10n.swift` with no hit left (`graphReadAttempt`, `graphReadFailedSettled`); `scripts/check-guidance-names.py` passes (check-pr runs it)
- [ ] Every row of **Rules this plan bends** has the user's answer, and its rule is updated (none)
- [ ] UI matches the brief inventory and the agreed board; no new kit primitive
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together): 0.6.0 → 0.6.1, or the next patch on `main`
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated
- [ ] add-workspace-place "Stale-while-revalidate" and macos-client-patterns.md's QueryHandle row teach `reading`, `failureRevision` and `isReading`
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

Each line says where it runs: **Preview** (`PersonDetailNoticePreviews.swift` or the record links column previews, in Xcode's canvas at a 1102 pt and a 2296 pt wide frame) or **App** (a Debug build on a real project).

1. **Preview** "Sources stale": the two rows stay; "Sources could not be read" with the stale sentence above them; with VoiceOver on, it is spoken and the keyboard does not move (AC-1.1, AC-1.6).
2. **Preview** "Record links failed": the compact notice alone in the column, its top aligned with the rows; the rows are untouched (AC-1.5).
3. **Preview** "Sources stale": press Try again with the keyboard; the Button shows its loading state, the read fails again, the notice is spoken again and Try again has its focus ring (AC-1.3).
4. **Preview** "Timeline, page stale": the notice under the switcher, Timeline unchanged. **Preview** "Page failed": the notice alone at the top (AC-1.4, AC-2.3).
5. **Preview** "Family failed": the notice alone with Try again in place of the tree; Try again keeps it on screen with its Button loading (AC-2.1, AC-2.2).
6. **Preview** "Sources gone": the gone line, no Try again, no rows, no column (AC-3.2).
7. **App**: open a person's page, then in a second window unpromote that person's only record so the person is deleted: the first window leaves with "This person no longer exists. They may have been merged into another person." Do the same to an Event: "This record no longer exists. It may have been merged into another." (AC-3.1, AC-3.3). Then, with Sources open on a person with two records, unpromote one in the second window: the line goes, the header shows "Refreshing" while it reloads, and no notice appears (AC-1.2's success path).
8. **Preview** record links column "Not counted": 40 nodes, no arc, the "compare" note in place of the caption; "Lines dropped", "1,001 graphs" and "Both" show the "left out" or "compare" note (AC-4.1, AC-4.2, AC-4.3, AC-4.4). **App**: a person with records on two linked graphs shows the caption and no note (AC-4.5).
9. **Preview** at both widths for 1, 2 and 8: notices keep the band's and the column's widths; resizing the canvas across 1600 keeps them and the keyboard (PN-W1, PN-W2).

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| The inner `.id` remount drops the keyboard before `FocusClaim` writes it | `FocusClaim` writes twice across two main-queue turns; manual check 3 |
| A `.gone` behind a value now leaves a page that used to stay stale; a transient not-found could eject the researcher | Only the engine code `conclusiondetails.not_found` reads as gone, as for a first load today; any other error is a notice |
| `ConclusionPageSectionHost` and `ConclusionDetailPage.swift` change beside S16-06's edits | S16-06 rebases; the map lists the shared file |
| Promote's or the Source form's retry callout changes | They pass no `failureRevision`, so D-4 leaves them as today; Promote's load error checked once by hand |

**Stop and report if:**

- `QueryHandle.applyFailure` is not the only place `status` becomes `.error`, or a success path leaves `error` set (D-1's assumption).
- `session.reread` on a key in `.error` with a value does not keep the value while it reads (`ensureQuery` `hadStaleValue`).
- The agreed board places a notice somewhere D-6, D-8 or D-11 does not, or adds a control the brief's inventory does not list.
- `CircularLinkDiagram` cannot draw a dataset with nodes and no arcs at the circle's size.
- `queryReadAndProgressReads` needs a change to pass (that would mean `read` changed).
- A `PVRetryCallout` caller other than the three D-4 names would change behavior.

## 12. Open questions

- **Board S16-D1.** The plan is blocked until the board is agreed. Recommended: the brief as written; on agreement, copy any changed placement or copy into D-6, D-8, D-11 and **Strings**.
- **Remove on a stale line** (review B-2): leave it as today, with a line in the follow-ups idea, or decide it here. Recommended: leave it; the stale notice already says the rows are old, and Try again is one press away.

## 13. PR

**Title:** S16-04 — The Person page says when a reload failed or the diagram is partial
**Body:** objective in one sentence; what shipped against US-1 – US-4; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-16/prs/S16-04-person-page-notices.md`.
