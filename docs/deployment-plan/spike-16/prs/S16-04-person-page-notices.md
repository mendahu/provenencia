# S16-04 — The Person page says when a reload failed or the diagram is partial

**Status:** Blocked — on board S16-D1
**Spike / slice:** Spike 16 · Slice 1b — Current data and honest notices
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** [S16-D1](../design/S16-D1-person-page-notices.md), board not drawn yet (one brief for the Person page's notices and the record links note: PN-1 – PN-9, PN-W1, PN-W2)
**Board:** — (to be drawn on [Claude Design — Person Detail](https://claude.ai/artifact/N9V52ewQjxeNvh3Az3YkXQ), turn "S16-D1 · Notices"). The implementer reads it with the Artifact tool's `read` on that link. Until it is agreed, **Design → macOS** below describes each state in words from the brief
**Depends on:** S16-03 (landed [#769](https://github.com/mendahu/provenencia/pull/769)), S16-01 (landed [#756](https://github.com/mendahu/provenencia/pull/756)), S16-D1 (board not agreed)
**Blocks:** S16-09; Spike 15's S15-21 (the provenance sheet's key reads the same failure reading)
**Branch:** `s16-04-person-page-notices`
**Version bump:** patch — a Mac UI fix (a failed reload and a partial diagram now say so; a gone person is named a person) with no wire or catalog change: 0.6.0 → 0.6.1, or the next patch on `main` ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-workspace-place`](../../../../.agents/skills/add-workspace-place/SKILL.md) ("Stale-while-revalidate"), [`add-ui-component`](../../../../.agents/skills/add-ui-component/SKILL.md), [`add-localized-string`](../../../../.agents/skills/add-localized-string/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md)
**Rules that bind:** `.cursor/rules/swift-tests.mdc`, `.cursor/rules/versioning.mdc`
**Context:** [audit](../person-page-audit.md) P-5, P-10, P-17; working rules "(Batch 1) A refresh failure is read beside its value", "(Batch 1) One open section, one list of reads", "(Batch 1) A person's records are records", "(Batch 1) The record links read is bounded and says so"; [`macos-client-patterns.md`](../../../macos-client-patterns.md) (the `QueryHandle` row, "One sectioned detail page", § 9); the Evidence graph's and Promote's retry notices (`EvidenceGraphView.readErrorCallout`, `PromoteView.notices`) for `PVRetryCallout`
**Planned:** 2026-11-10 against `main` @ `e70bba8`
**Review:** [`S16-04-person-page-notices.review.md`](S16-04-person-page-notices.review.md)

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

When a Person page's reload fails, the page keeps what it showed and says so, with Try again, in the place that read failed: the header band for the page's read, the top of a section for the section's read, the record links column for the diagram's read. A first load that fails offers Try again too, and VoiceOver hears each notice. A person deleted or merged elsewhere leaves with "This person no longer exists". The record links diagram says when its drawing is partial and stops drawing links it did not compare as weak (audit P-5, P-10, P-17).

**Why now.** S16-03 gave every section one list of reads and S16-01 put `notCounted` and `linesDropped` on the wire; both have landed. S16-09 and Spike 15's S15-21 read the failure reading this PR adds.

**What stays.** `QueryHandle.read` and every caller of it; the lists' presentation; the page frame, the header, the "Refreshing" line, the sections' rows, folds and empty states; the diagram's circle, nodes and title; the wire; Event and Place pages' gone line.

## 2. User stories

**US-1.** As a researcher, when a section's reload fails after a write, I see a notice with Try again above the rows that stayed, so that I know they are out of date and can read them again.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | Sources is open with two Source rows; `listPersonClaims` throws a generic error; a write names the person | The two rows stay; above them a danger notice "Sources could not be read", message "(the error's text) What is shown is from the last time it was read.", with Try again (PN-1, PN-4) |
| AC-1.2 | AC-1.1, then `listPersonClaims` answers again and Try again is pressed | The Button shows its loading state and the header shows "Refreshing"; the read succeeds and the notice goes (PN-6) |
| AC-1.3 | AC-1.1, Try again pressed with the keyboard, and the read fails again | The notice stays, the keyboard is on its Try again, and VoiceOver speaks the notice again (PN-6, PN-7) |
| AC-1.4 | The page's own read (`getConclusionDetail`) fails behind a shown page, Timeline open | A notice "This page could not be read" with Try again sits under the section switcher; the Timeline under it is unchanged (PN-3). On an Event page, the notice sits under the header, above Details |
| AC-1.5 | The record links read fails behind a shown diagram | A compact notice "The connections could not be read" with Try again sits at the top of the column, above the diagram's title; the Sources rows are unchanged (PN-5) |
| AC-1.6 | A reload fails while no Try again was pressed | VoiceOver speaks the notice; the keyboard does not move (PN-7) |

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
| AC-3.2 | Sources open; `listPersonClaims` answers `conclusiondetails.not_found` before the page's own read does | Sources shows a danger notice "Sources could not be read" with that line and no Try again (PN-8) |
| AC-3.3 | An Event page whose event is gone | The Toast keeps "This record no longer exists. It may have been merged into another." (PN-8) |

**US-4.** As a researcher, when the record links diagram is partial, a note says so, and a read that did not compare draws no link.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-4.1 | A read with `notCounted` and 40 graphs whose pairs each have one weak line | 40 nodes, no arc; no node is spoken as sharing no one; the note "Too many records to compare. The links between them are not drawn." in place of the caption (PN-9) |
| AC-4.2 | A read with `linesDropped` | Every arc sent is drawn; the note "Too many records to draw them all. The strongest links are drawn." (PN-9) |
| AC-4.3 | A read with 1,001 distinct graphs and neither flag | 1,000 nodes drawn; the same "draw them all" note (PN-9) |
| AC-4.4 | A read with both `notCounted` and `linesDropped` | The "did not compare" note and no arc (PN-9) |
| AC-4.5 | A full read (neither flag, at most 1,000 graphs) | No note; the caption as today |

## 3. Scope

| In | Covered by |
| --- | --- |
| `QueryHandle.reading` (loading, failed, loaded with an optional refresh failure) and `failureRevision`; `read` unchanged | STEP-1, AC-1.1, AC-1.3 |
| `ConclusionDetailPresentation` built from the reading: a refresh failure beside the content; `conclusiondetails.not_found` behind a value reads as gone | STEP-2, AC-1.1, AC-3.1 |
| PN-1, PN-2, PN-6, PN-7: one notice snowflake over `PVRetryCallout` that re-announces per failure and keeps the keyboard after a pressed Try again | STEP-4, AC-1.2, AC-1.3, AC-1.6, AC-2.2 |
| PN-3: the page's notice in the header band and on a first-load failure; Event and Place pages too | STEP-5, AC-1.4, AC-2.3 |
| PN-4, PN-5: the section host's band and column notices above stale content and on a first-load failure; Try again rereads the host's read | STEP-6, AC-1.1, AC-1.5, AC-2.1 |
| PN-8: the kind supplies its gone line; Person's is "This person no longer exists. They may have been merged into another person."; a section's gone notice has no Try again | STEP-5, STEP-6, AC-3.1 – AC-3.3 |
| PN-9: `PersonRecordLinksContent.partial`; no arc when not counted; the column's note in place of the caption | STEP-8, AC-4.1 – AC-4.5 |
| PN-W1, PN-W2: notices keep the band's, the header's and the column's widths; state survives the Regular/Wide line | STEP-5, STEP-6, STEP-8, manual check |
| P-5's FakeStore test | STEP-7, T-5 |
| `add-workspace-place` and `macos-client-patterns.md` teach the reading | STEP-9 |

| Out | Home |
| --- | --- |
| The Conclusion lists' silent refresh failure | [idea](../../../ideas/person-page-followups.md) (its own brief) |
| `EvidenceGraphModel`'s and `PromoteModel`'s attempt counters onto `failureRevision` | not planned (the working rule says "may follow"; no PR names it) |
| A notice for a failed Source types (supporting) read | not planned (the labels drop out, as today; brief §2.1) |
| Stale-while-refreshing of hidden keys, the 50-entity cap | S16-09 |
| Section-switch and Remove announcements; "records" in the header count and confirm | S16-06 |
| Back restoring a section's place (P-19) | [idea](../../../ideas/person-page-followups.md) |
| Any data about a line, a new arc stroke style | not planned (Spike 13 rule "The view is a circular link diagram") |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Handle state | `macos/App/Features/Workspace/Session/QueryHandle.swift` `QueryRead`, `QueryHandle.read`, `failure`, `applyFailure(_:clearValue:)`, `beginLoading(keepingValue:)` | `read` folds a value first, even beside a failure, so a reader cannot see a refresh failure. `failure` is the error while `status == .error`. `applyFailure` keeps a shown value (`clearValue: false`) and bumps `revision` only when it clears. `beginLoading(keepingValue: false)` sets `.loading` and leaves `error` set, so during a retry after a first-load failure `read` is `.loading` |
| Session | `macos/App/Features/Workspace/Session/WorkspaceSession.swift` `ensureQuery` (≈ line 418), `reread(_:)`, `isRefreshing(_:)` | A failed load keeps a value it had (`clearValue: !hadStaleValue`). `reread` invalidates then queries and keeps the stale value. `isRefreshing` is `hasValue && isFetching` |
| Page presentation | `macos/App/Features/Conclusions/ConclusionDetailPage.swift` `ConclusionDetailPresentation` (lines 6–31) | Built from `read`: `.loading`, `.gone` (engine code `conclusionNotFound`, only when no value), `.failed(String)`, `.content(_, refreshing:)`. A refresh failure is invisible; a not-found reload behind a value keeps the stale page |
| Page | same file, `ConclusionDetailLoaded` (≈ 250–326), `ConclusionDetailHeaderBand` (≈ 330–395) | `.failed` draws a plain danger `PVCallout` with `errorIdentifier`; gone leaves with `L10n.Errors.conclusionDetailsNotFound` through `leavesForSectionRoot(when:notice:)`. The header band draws the header and, on a page with sections, the switcher and "Refreshing" |
| Page sections context | same file, `ConclusionDetailPageSectionsContext` | `projectDir`, `refreshing`, `pageIdentifier`, `lead`, `openPageSection` |
| Section host | `macos/App/Features/Conclusions/ConclusionPageSectionHost.swift` `ConclusionPageSectionHost` | Builds `ConclusionDetailPresentation` from `handle.read`; `.failed` and `.gone` draw `PVCallout(tone: .danger, title: errorTitle, …)` (compact in `.column`) with id `\(idRoot).error`; `.gone` uses the shared not-found line; content draws with no failure shown |
| Sections | `PersonFamilySection.swift`, `PersonTimelineSection.swift`, `PersonSourcesSection.swift` (`body` and `diagram`) | Each builds a host with `section`, `entityId`, `accessibilityIdentifierRoot`, `errorTitle`; the Sources `diagram` passes `read: .companion, placement: .column(…)` |
| Person fill | `PersonDetailView.swift` `PersonDetailView`, `PersonDetailSections` | Passes `pageIdentifier: "persons.detail"`, `errorIdentifier: "persons.detail.error"`; the sections get `entityId` and `pageIdentifier` |
| Event, Place | `EventDetailView.swift`, `PlaceDetailView.swift` | Use `ConclusionDetailPage` with the default page sections |
| Section reads | `macos/App/Features/Workspace/ConclusionPageSection.swift` `reads(project:entityId:)`, `read(_:project:entityId:)` | One list per section (S16-03); the host takes its role's key from it |
| Record links | `PersonRecordLinksContent.swift` `PersonRecordLinksContent.init(links:graphLocations:graphSourceRefs:)`, `maxNodes` | Keeps the first 1,000 distinct graphs silently; draws every sent line, a weak one at `weakLineWidth`, so a not-counted read draws every arc thin; a node with no neighbour is spoken with `recordLinksNodeApartAccessibility` ("…, shares no one") |
| Wire flags | `macos/App/Platform/GenealogyStore.swift` `CatalogPersonRecordLinks.notCounted`, `.linesDropped` | Mapped from the wire by S16-01; nothing reads them |
| Column | `PersonRecordLinksColumn.swift` `PersonRecordLinksColumn` | Title, caption (`recordLinksCaption`, hidden for a single graph), the diagram, the one-record line |
| Retry kit piece | `macos/App/DesignSystem/Components/Callout/PVRetryCallout.swift` `PVRetryCallout` | Danger `PVCallout(announces: true)` with one small secondary `PVButton` (loading state), `compact`, `retryFocused` binding |
| Announce | `DesignSystem/Components/Announce/PVAnnounce.swift` | Speaks on appear and when the text changes; the same text again while mounted is not spoken |
| Focus | `macos/App/Platform/FocusClaim.swift` `FocusClaim.claim` | Writes a focus binding next turn and again a turn later |
| Strings | `macos/App/Platform/L10n.swift` `L10n.Conclusions.familyErrorTitle`, `timelineErrorTitle`, `sourcesErrorTitle`, `recordLinksErrorTitle`, `recordLinksCaption`, `recordLinksNodeApartAccessibility`; `L10n.Errors.conclusionDetailsNotFound` | Section titles stay; no Try again, stale, page title, person-gone or partial-note keys exist |
| FakeStore | `macos/App/Platform/FakeStore.swift` `getPersonRecordLinksError`, `listPersonClaims`, `getPersonFamily`, `getConclusionDetail` | Only the record links read has an error hook; the others throw `conclusiondetails.not_found` for an unseeded id |
| Tests | `macos/ProvenenciaTests/WorkspaceSessionTests.swift` `queryReadAndProgressReads`, `applyPersonPageKeepsTheOpenSectionsReadsVisible`, `aKeyInvalidatedAndRefreshedRevalidatesOnce`; `PersonDetailTests.swift` `presentationFollowsTheHandle`; `PersonRecordLinksContentTests.swift` `allLinesDrawAndNodesStopAtTheCap`, `goStoreCarriesThePartialFlags` | Cover `read`, the presentation's four cases, visible section keys, the node cap and the flags' mapping |

**Precedent.** `EvidenceGraphView.readErrorCallout` and `EvidenceGraphModel.readGraphAgain` (S14-20): a `PVRetryCallout` remounted with `.id(attempt)` so it announces again, Try again through `session.reread`. Copy the remount and the reread. Do not copy the model-held attempt counter (the handle's `failureRevision` replaces it here) or `PromoteView`'s focus on every appear (this PR focuses only after a pressed Try again, PN-7).

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | `QueryReading<Value>` *(new)* in `QueryHandle.swift`: `.loading`, `.failed(Error)`, `.loaded(Value, refreshFailure: Error?)`. `QueryHandle.reading` *(new)*: a value → `.loaded(value, refreshFailure: failure)`; else an `error` → `.failed(error)`, also while a retry after a first-load failure runs; else `.loading`. `read` is not changed | Working rule: one reading, `read` unchanged for its other callers. Holding `.failed` during the retry keeps the notice and its Try again on screen, so its loading state shows and the keyboard has somewhere to stay (PN-6) | Adding a case to `QueryRead` — every list and Evidence graph caller would have to handle it; deriving `read` from `reading` — `read` reports `.loading` during that retry today, and the rule keeps it |
| D-2 | `QueryHandle.failureRevision` *(new)*, `private(set) var Int`, bumps in every `applyFailure`. `QueryHandle.isReading` *(new)*: `status == .loading || isFetching` | A notice remounts on it (`.id`) so a repeated failure with the same text is spoken again (PVAnnounce speaks a mounted text once). `isReading` is the retry Button's loading state in both the first-load and the refresh case | A counter on each view or model (the S14-20 shape) — four call sites would each keep one |
| D-3 | `ConclusionDetailPresentation` is built from `QueryReading`: `init(_ reading:isFetching:content:)`. `.content` gains `refreshFailure: String?` (the message from `L10n.Errors.message(for:)`). A `conclusiondetails.not_found` refresh failure behind a value is `.gone`, like a first-load one | The page and the host draw from one presentation (S9-24). A person gone behind a shown value must leave too; S16-09's scope reads "a not-found reload over a stale value reads as gone (S16-04's reading)" | A separate `.refreshFailed` case — the content still draws, so it is a property of `.content`, not another state |
| D-4 | One snowflake `ConclusionReadNotice` *(new)*, `macos/App/Features/Conclusions/ConclusionReadNotice.swift`, wraps `PVRetryCallout` for every failed-read notice on the page: title, message, `stale: Bool` (adds the stale sentence), `compact`, `isRetrying`, `failureRevision`, `identifier`, `retry`. It owns `@FocusState retryFocused` and `@State retryPressed`; the inner `PVRetryCallout` carries `.id(failureRevision)`, the wrapper does not, so its state survives the remount. On the inner callout's appear, when `retryPressed`, `FocusClaim.claim($retryFocused, true)`; the Try again action sets `retryPressed = true` and calls `retry` | Three places draw the same notice (page band and page column, section band, diagram column); one owner of the focus and re-announce rule. A feature snowflake, not kit: `PVRetryCallout` is the kit piece, unchanged (brief inventory) | `PVRetryCallout` inline at each site — three copies of the focus rule; a kit change to `PVRetryCallout` — Promote and the Evidence graph would change with it |
| D-5 | The page's notice goes in `ConclusionDetailHeaderBand`, after the switcher (or after the header on a page without sections), from `.content`'s `refreshFailure`; title `L10n.Conclusions.pageReadErrorTitle` *(new)*; Try again calls `session.reread(key)` for the page's `.conclusionDetail` key. `.failed` draws the same notice (not stale) in `PVPageColumns(primary:)` with id `errorIdentifier` | PN-3; the page's read covers the header, Vitals and the Gallery, so the band that holds the header holds its notice, on every section | A notice above Vitals only — the header would read stale under Family with no word |
| D-6 | The kind supplies its gone line: `ConclusionDetailPage` takes `goneNotice: LocalizedStringResource` *(new parameter)*, default `L10n.Errors.conclusionDetailsNotFound`; `PersonDetailView` passes `L10n.Conclusions.personGone` *(new)*. `ConclusionDetailPageSectionsContext.goneNotice` *(new)* carries it; `PersonDetailSections` passes it to `PersonFamilySection`, `PersonTimelineSection` and `PersonSourcesSection` (a new stored `goneNotice` each), which pass it to their hosts (`ConclusionPageSectionHost` gains `goneNotice:`) | Working rule "A person's records are records": the kind supplies the line; Event and Place keep the shared one with no change at their call sites | A switch on `WorkspaceSection` — copy for a deleted entity does not belong on the sidebar's enum; an environment value — hides the dependency |
| D-7 | The host's states: `.loading` → today's spinner; `.failed` → `ConclusionReadNotice` (not stale) in place of the content; `.gone` → `PVCallout(tone: .danger, title: errorTitle, message: goneNotice)` with no button (compact in `.column`); `.content` with a `refreshFailure` → `ConclusionReadNotice` (stale) above the content, or above the empty state in `.band`. All keep today's id `\(idRoot).error`; Try again's id is `\(idRoot).error.tryAgain`. Try again calls `session.reread(key)` with the host's own key (its role's entry in `ConclusionPageSection.reads`). In `.band` the notice takes `.pvKeepsPageBand()` | PN-1, PN-2, PN-4, PN-5, PN-8; working rule "One open section, one list of reads": the host already takes its key from the list. A gone read would fail the same way on retry | Try again rereading every read of the section — the band and the column fail apart and each notice reports one read (brief PN-6) |
| D-8 | Notice ids: page first load `errorIdentifier` (`persons.detail.error`, `events.detail.error`, `places.detail.error`), its Try again `\(errorIdentifier).tryAgain`; page refresh notice `\(pageIdentifier).refreshError`, its Try again `\(pageIdentifier).refreshError.tryAgain`; the diagram's note `\(accessibilityIdentifierRoot).recordLinks.partial` | Existing ids stay; the band notice lives beside a shown page so it cannot reuse the first-load id that UI tests read as "the page failed" | — |
| D-9 | `PersonRecordLinksContent.Partial` *(new)* enum `.notCounted`, `.leftOut`; `partial: Partial?` *(new)*. `.notCounted` when `links.notCounted`; else `.leftOut` when `links.linesDropped` or the read had more than `maxNodes` distinct graph ids; else nil. With `.notCounted` the dataset has no arcs, and each node's accessibility label is `recordLinksNodeUncountedAccessibility` *(new)* in place of the apart or connected label | PN-9; a link that was not compared cannot be thick or thin, and Go's other not-counted shape (no pair at all) already draws none, so both look the same. Speaking "shares no one" would state what was not counted | One "uncounted" arc style — a new stroke on the diagram snowflake for a crafted-data case (brief non-goal) |
| D-10 | `PersonRecordLinksColumn` shows, when `content.partial != nil`, `PVCallout(tone: .warning, message: L10n.string(note), compact: true)` (`note` is `recordLinksNotCounted` for `.notCounted`, `recordLinksPartial` for `.leftOut`), with no title, not announced, in place of the caption (also for a single graph), with id per D-8 | PN-9; brief inventory: Callout, warning, compact | A muted caption line — the brief picks the kit's hint piece |
| D-11 | The stale message is one format key, `conclusions.readNotice.staleMessage` *(new)* "%@ What is shown is from the last time it was read.", the argument the error text; Try again is `conclusions.readNotice.tryAgain` *(new)* "Try again" | One sentence per language, word order free; a Conclusions key, not Promote's `promote.error.tryAgain` (each feature owns its copy, `add-localized-string`) | Concatenating two strings in Swift |
| D-12 | FakeStore gains `listPersonClaimsError: Error?` *(new)*, read in `listPersonClaims` before the seeded lookup, the shape of `getPersonRecordLinksError` | P-5's test needs a generic failure behind a seeded value; removing the seed would throw not-found, which is gone, not a failure | A throwing closure hook — no other FakeStore read uses one |

**Rules this plan bends.** None.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `QueryReading` *(new)*, `QueryHandle.reading`, `.failureRevision`, `.isReading` | type, properties | `QueryRead`, `QueryHandle.read`, `.failure`, `.isFetching` | D-1, D-2; searched `rg "failure\b|attempt" macos/App/Features` — the attempt counters are per model (`EvidenceGraphModel.graphReadAttempt`, `PromoteModel.loadAttempt`) |
| `ConclusionReadNotice` *(new)* | view (snowflake) | `PVRetryCallout`; `SourceFormLoadArea`'s retry branch | D-4: it composes `PVRetryCallout`; `SourceFormLoadArea` is a form's loading area with skeleton bars, not a notice over content |
| `ConclusionDetailPresentation.content(_, refreshing:, refreshFailure:)` | case change | today's `.content(_, refreshing:)` | Extended, D-3 |
| `goneNotice` on `ConclusionDetailPage`, the context, three sections and the host | parameter | `errorIdentifier`, `errorTitle` parameters | Same shape: the kind's copy passed down, D-6 |
| `PersonRecordLinksContent.Partial`, `.partial` | type, property | `notCounted`, `linesDropped` on `CatalogPersonRecordLinks` | The content adds its own 1,000-node cut, which the wire does not know (working rule "The record links read is bounded and says so": one note, whichever side cut it) |
| `L10n.Conclusions.readNoticeTryAgain`, `readNoticeStaleMessage(_:)`, `pageReadErrorTitle`, `personGone`, `recordLinksNotCounted`, `recordLinksPartial`, `recordLinksNodeUncountedAccessibility(ref:graph:source:)` | keys | `promote.error.tryAgain`, `error.conclusiondetails.not_found`, `recordLinksNodeApartAccessibility` | D-11; Conclusions copy lives in `L10n.Conclusions` |
| `FakeStore.listPersonClaimsError` | test hook | `getPersonRecordLinksError` | Same shape, D-12 |

**Builds on.** `PVRetryCallout` (kit); `session.reread`; `ConclusionPageSection.reads` (S16-03); `leavesForSectionRoot(when:notice:)`; `FocusClaim`; the S14-20 remount.
**Deletes.** The plain `PVCallout` failure branch of `ConclusionDetailLoaded` and of `ConclusionPageSectionHost.error(_:)`, replaced by D-5 and D-7. `ConclusionPageSectionHost.Placement.column`'s doc line "a compact danger Callout" is reworded to the notice. No key goes: `L10n.Errors.conclusionDetailsNotFound` stays for Event, Place and the errors map.
**Next adds.** S16-09 reads D-3's "not-found behind a value is gone" for its stale keys. S15-21's provenance sheet key reads `QueryReading` and draws `ConclusionReadNotice`. The lists' later brief builds `ConclusionListPresentation` from `reading` the same way. A section that gains a read adds a `reads` entry; its host draws the notice with no new code.

### Reach

| New path or state | Destination | Checked: resets on the new id · handles a missing referent · shows the state |
| --- | --- | --- |
| A page whose detail reload answers not-found behind a shown value now leaves | `ConclusionDetailLoaded` `.leavesForSectionRoot` | `onChange(of: gone, initial: true)` fires once on the change; the page is keyed by entity id, so a new page starts clean |
| A section shows content and a notice together | `ConclusionPageSectionHost` `.content` | Folds, focus and hover live in the section views, which stay mounted under the notice; the notice is a sibling above them |
| Content with a refresh failure and an empty value | host `.band` empty state | Notice above `PVEmptyState` (D-7) |
| The diagram with no arcs (not counted) | `CircularLinkDiagram` | It already draws a dataset with no arc (a person whose records share no one, `aNodeThatSharesNoOneIsSpokenAsApart`) |

### Scale

At the map's Scale (a person in 50 Evidence graphs): no new read. A failure adds one notice per failed read; Try again issues one `reread` of one key. `PersonRecordLinksContent` does less work when not counted (no arcs). Nothing loops per row.

### Edges

- **Assumptions.** `applyFailure` is the only place `status` becomes `.error`, and every success clears `error` (`applySuccess`, `resetForInvalidation`), so `reading` never reports an old failure after a success (T-1). `beginLoading(keepingValue: false)` leaves `error` set: D-1 relies on it to keep `.failed` during a first-load retry; `resetForInvalidation` and eviction clear it, so a key evicted and queried again starts at `.loading` (T-1).
- **Neighbors.** A refresh failure that is not-found → `.gone`; any other code → content plus notice (T-2). A first-load not-found → `.gone` as today (T-2). Both the page read and a section read failing → two notices, each with its own Try again (manual check 4). Both Sources reads failing → the band and the column notice, independent (AC-1.5, manual check 2).
- **Removed inputs.** A person merged away while Sources is open: the section's read may say not-found first (gone notice, AC-3.2), then the page's read leaves (AC-3.1). A notice whose read later succeeds through a write (not Try again) goes with the failure (`applySuccess` clears `error`, T-1).
- **Real data.** None — no table or seed drives a rule here.
- **Untrusted size.** `links.graphs` is bounded by the existing `maxNodes` loop; `partial` is decided in the same loop (one pass, counts distinct ids past the cap with no extra set).

### Schema

None — no catalog change.

### Go core

None — S16-01 already sends `not_counted` and `lines_dropped`.

### FFI

None — no wire change.

### macOS

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Read notice (PN-1 – PN-6) | `PVRetryCallout` inside `ConclusionReadNotice` *(new)* | `macos/App/Features/Conclusions/ConclusionReadNotice.swift` | D-4. Full in the header band and a section band, `compact` in the column |
| Gone notice in a section (PN-8) | `PVCallout(tone: .danger)` | `ConclusionPageSectionHost.swift` | D-7; no button |
| Partial note (PN-9) | `PVCallout(tone: .warning, compact: true)` | `PersonRecordLinksColumn.swift` | D-10; not announced |
| Page frame | snowflake, Extend | `ConclusionDetailPage.swift` | D-5, D-6 |
| Section host | snowflake, Extend | `ConclusionPageSectionHost.swift` | D-7 |
| Leave Toast | `leavesForSectionRoot` | `WorkspaceNavigation.swift` | Only the words change (D-6) |

The frames in words (from the brief, until the board is agreed):

- **Header band, refresh failure.** Header, switcher with "Refreshing" when it runs, then the notice at the header's width. Event and Place: header, then the notice, then Details.
- **Page, first-load failure.** No header; the notice alone at the top of the page column.
- **Section band.** The notice first, at the page band's width even when the section fills the pane at Wide (Timeline, Sources), then the unchanged rows, tiles or lifeline. First-load failure: the notice alone.
- **Diagram column.** Compact notice above the title. First-load failure: the compact notice is all the column shows. Partial: title, then the warning note in place of the caption, then the circle.
- **Gone in a section.** Danger Callout, section title, the person's gone line, no button.

States: loading as today; error per above; empty with a refresh failure shows the notice over the empty state. Accessibility: each notice announces when it appears and on each new failure (danger priority); a notice never takes the keyboard unless its Try again was pressed; the partial note is read in order before the diagram.

### Strings

L10n keys *(new)*, all under `L10n.Conclusions`, in `Localizable.xcstrings`:

- `conclusions.readNotice.tryAgain` = "Try again"
- `conclusions.readNotice.staleMessage` = "%@ What is shown is from the last time it was read." (argument: the error text)
- `conclusions.detail.errorTitle` = "This page could not be read"
- `conclusions.person.gone` = "This person no longer exists. They may have been merged into another person."
- `conclusions.recordLinks.notCounted` = "Too many records to compare. The links between them are not drawn."
- `conclusions.recordLinks.partial` = "Too many records to draw them all. The strongest links are drawn."
- `conclusions.recordLinks.nodeUncountedAccessibility` = "%1$@, %2$@, %3$@" (ref, graph, Source; the connected and apart labels' first three arguments, with no claim about links)

Removed: none. Old-scope copy: `error.conclusiondetails.not_found` ("This record no longer exists. It may have been merged into another.") stays for Event, Place and the engine errors map; the Person page stops showing it (D-6).

## 7. Steps

**STEP-1 — The handle's reading and failure revision**
Files: `macos/App/Features/Workspace/Session/QueryHandle.swift`, `macos/ProvenenciaTests/WorkspaceSessionTests.swift`. Skill: `add-swift-test`.
Do: add `QueryReading`, `reading`, `failureRevision` (bumped in `applyFailure`), `isReading` (D-1, D-2), each with a doc comment. Do not touch `read`, `failure` or any caller. Write T-1.
Check: `xcodebuild test -project macos/Provenencia.xcodeproj -scheme Provenencia -destination 'platform=macOS' -only-testing:ProvenenciaTests/WorkspaceSessionTests` passes.

**STEP-2 — The presentation reads it**
Files: `macos/App/Features/Conclusions/ConclusionDetailPage.swift` (`ConclusionDetailPresentation` and its use in `ConclusionDetailLoaded`), `ConclusionPageSectionHost.swift` (its `loaded`), `macos/ProvenenciaTests/PersonDetailTests.swift`.
Do: D-3. Both builders pass `handle.reading`. Views keep drawing what they draw today; the new `refreshFailure` is not drawn yet. Update `presentationFollowsTheHandle` to the new init and add T-2.
Check: `-only-testing:ProvenenciaTests/PersonDetailTests` passes; the app builds.

**STEP-3 — Copy**
Files: `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`. Skill: `add-localized-string`.
Do: add the seven keys in **Strings** with comments.
Check: `python3 scripts/check-localizable-xcstrings.py` passes; the app builds.

**STEP-4 — The notice snowflake**
Files: `macos/App/Features/Conclusions/ConclusionReadNotice.swift` *(new)*. Skill: `add-ui-component`.
Do: D-4. Parameters: `title: LocalizedStringResource`, `error: String`, `stale: Bool`, `compact: Bool`, `isRetrying: Bool`, `failureRevision: Int`, `identifier: String`, `retry: () -> Void`. Message: `stale ? L10n.Conclusions.readNoticeStaleMessage(error) : error`. Retry copy `L10n.Conclusions.readNoticeTryAgain`, retry id `\(identifier).tryAgain`. Add a `#Preview` with a stale and a first-load state.
Check: the app builds; the preview renders.

**STEP-5 — The page's notice and gone line**
Files: `ConclusionDetailPage.swift`, `PersonDetailView.swift`, `EventDetailView.swift` and `PlaceDetailView.swift` (unchanged unless the build needs it).
Do: D-5, D-6, D-8. `ConclusionDetailHeaderBand` gains the notice's inputs (refresh failure message, `isRetrying`, `failureRevision`, retry action) and draws `ConclusionReadNotice(stale: true)` last in its `VStack`. `.failed` draws `ConclusionReadNotice(stale: false)` with id `errorIdentifier`. `leavesForSectionRoot` takes `L10n.string(goneNotice)`. `PersonDetailView` passes `goneNotice: L10n.Conclusions.personGone`; the context carries it.
Check: the app builds; T-3 passes.

**STEP-6 — The section host's notices**
Files: `ConclusionPageSectionHost.swift`, `PersonFamilySection.swift`, `PersonTimelineSection.swift`, `PersonSourcesSection.swift`, `PersonDetailView.swift` (`PersonDetailSections`).
Do: D-6, D-7. The host gains `goneNotice:`; its `error(_:)` becomes the three states of D-7; `.content` draws the stale notice above the content in a `VStack(alignment: .leading, spacing: PVSpacing.space7)`. The three sections take a stored `goneNotice` from `PersonDetailSections` (`context.goneNotice`) and pass it to each host (Sources passes it to both).
Check: the app builds; `-only-testing:ProvenenciaTests/PersonDetailTests` passes (identifier tests unchanged).

**STEP-7 — P-5's test**
Files: `macos/App/Platform/FakeStore.swift`, `macos/ProvenenciaTests/WorkspaceSessionTests.swift`.
Do: D-12; write T-5 and T-6.
Check: `-only-testing:ProvenenciaTests/WorkspaceSessionTests` passes.

**STEP-8 — The partial diagram**
Files: `PersonRecordLinksContent.swift`, `PersonRecordLinksColumn.swift`, `macos/ProvenenciaTests/PersonRecordLinksContentTests.swift`.
Do: D-9, D-10. Update `allLinesDrawAndNodesStopAtTheCap` to also expect `.leftOut`; write T-7. Add `#Preview("Not counted")`, `#Preview("Lines dropped")`, `#Preview("1,001 graphs")` and `#Preview("Both")` beside the others.
Check: `-only-testing:ProvenenciaTests/PersonRecordLinksContentTests` passes.

**STEP-9 — Docs and version**
Files: `.agents/skills/add-workspace-place/SKILL.md` ("Stale-while-revalidate": a destination reads `reading` to show a refresh failure beside its value, and draws `ConclusionReadNotice` on a Conclusion page), `docs/macos-client-patterns.md` (the `QueryHandle` row: `reading`, `failureRevision`; `read` stays for readers that do not show refresh failures), `VERSION`, `core/version.go`, `MARKETING_VERSION`, `completed.md`, `deployment-plan.md` (map row), this plan, `design/S16-D1-person-page-notices.md` → `design/archive/`, `design/README.md`.
Do: per the Definition of done.
Check: `scripts/check-pr.sh` prints `check-pr: PASS`; links resolve.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Show a Try again notice when a Person page section fails to reload.").

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Swift | `WorkspaceSessionTests.swift` `queryReadingCarriesARefreshFailure` *(new)* | new handle → `.loading`; `applyFailure(clearValue: true)` → `.failed`, `failureRevision` 1; `beginLoading(keepingValue: false)` → still `.failed`, `isReading`; `applySuccess(3)` → `.loaded(3, nil)`; `applyFailure(clearValue: false)` twice → `.loaded(3, refreshFailure: non-nil)`, `failureRevision` 3; `applySuccess(4)` → `.loaded(4, nil)`; `resetForInvalidation` → `.loading`; `read` unchanged at every step (`.loading` during the first-load retry) | AC-1.1, AC-1.3, AC-2.2, D-1, D-2 |
| T-2 | Swift | `PersonDetailTests.swift` `presentationFollowsTheHandle` (updated) + `presentationShowsARefreshFailureBesideTheContent` *(new)* | `.loaded(detail, refreshFailure: generic)` → `.content(_, refreshing:, refreshFailure: message)`; `.loaded(detail, refreshFailure: not_found)` → `.gone`; `.failed(not_found)` → `.gone`; `.loaded(detail, nil)` → `refreshFailure` nil | AC-1.1, AC-3.1, D-3 |
| T-3 | Swift | `PersonDetailTests.swift` `personPageSuppliesItsGoneLine` *(new)* | `L10n.string(L10n.Conclusions.personGone)` == "This person no longer exists. They may have been merged into another person."; `L10n.string(L10n.Errors.conclusionDetailsNotFound)` unchanged | AC-3.1, AC-3.3, D-6 |
| T-4 | Swift | `PersonDetailTests.swift` `readNoticeStaleMessageWrapsTheError` *(new)* | `readNoticeStaleMessage("Disk full.")` == "Disk full. What is shown is from the last time it was read." | AC-1.1, D-11 |
| T-5 | Swift | `WorkspaceSessionTests.swift` `aFailedSectionReloadIsReadBesideItsRows` *(new)* | FakeStore seeds claims for "per-1"; `apply(location:)` the Person page on Sources; query and wait; set `listPersonClaimsError`; `apply(CatalogWriteEffects(handles: ["per-1"]))`; `settleWrites` → the claims handle's `reading` is `.loaded(seeded, refreshFailure: non-nil)` and `failureRevision` is 1; clear the error, `reread`, wait → `.loaded(_, nil)` | AC-1.1, AC-1.2, P-5 |
| T-6 | Swift | `WorkspaceSessionTests.swift` `aSecondFailureBumpsTheFailureRevision` *(new)* | T-5's state, `reread` with the error still set, wait → `failureRevision` 2, value kept | AC-1.3, D-2 |
| T-7 | Swift | `PersonRecordLinksContentTests.swift` `aReadThatDidNotCountDrawsNoArc`, `droppedLinesAreLeftOut`, `droppedGraphsAreLeftOut`, `notCountedWinsOverLeftOut`, `aFullReadIsNotPartial` *(new)* | AC-4.1, AC-4.2, AC-4.3, AC-4.4, AC-4.5 inputs, one test each in that order; for not counted, no node's label equals the apart label | AC-4.1, AC-4.2, AC-4.3, AC-4.4, AC-4.5, D-9 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `PersonDetailTests` (the presentation's case shape), `WorkspaceSessionTests` (`queryReadAndProgressReads` must still pass unchanged), `PersonRecordLinksContentTests`, the string catalog check.

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left
- [ ] Sweep: each renamed or deleted name `rg`'d by its old, qualified name across `.agents/`, `.cursor/rules/`, `docs/*.md` and `L10n.swift` with no hit left; `scripts/check-guidance-names.py` passes (check-pr runs it)
- [ ] Every row of **Rules this plan bends** has the user's answer, and its rule is updated (none)
- [ ] UI matches the brief inventory and the agreed board; no new kit primitive
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together): 0.6.0 → 0.6.1, or the next patch on `main`
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated
- [ ] add-workspace-place "Stale-while-revalidate" and macos-client-patterns.md's QueryHandle row teach `reading` and `failureRevision`
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

Build a Debug app. A read cannot be made to fail on demand on a real project, so the failure states (steps 1–4) are checked in `ConclusionReadNotice`'s previews and in a Debug run against a FakeStore session with `listPersonClaimsError` or `getPersonRecordLinksError` set (the way `WorkspaceSessionTests` builds one); the PR says which each step used. Steps 5–7 use a real project.

1. Person with records on four Sources, Sources open. Force the claims read to fail, make a write that names the person (unpromote a record in another window). The rows stay; the notice "Sources could not be read" with the stale sentence sits above them; VoiceOver speaks it. Restore, press Try again: the Button loads, "Refreshing" shows, the notice goes (AC-1.1, AC-1.2).
2. Same page, force only the record links read to fail: the compact notice sits above the diagram's title; the rows are untouched (AC-1.5).
3. Press Try again with the keyboard while the read still fails: the notice announces again and Try again keeps its focus ring (AC-1.3).
4. Force the page's read to fail with Timeline open: the notice under the switcher; Timeline unchanged. Open an Event and do the same: the notice under the header (AC-1.4).
5. Delete the person in another window: the page leaves with "This person no longer exists. They may have been merged into another person." Delete an Event the same way: "This record no longer exists. It may have been merged into another." (AC-3.1, AC-3.3). With the FakeStore session, make `listPersonClaims` answer not-found while the detail still answers: Sources shows the gone line with no Try again (AC-3.2).
6. With the FakeStore session, open Family with its read failing: the notice alone with Try again in place of the tree (AC-2.1); open a Person whose detail read fails: the notice alone at the top of the page (AC-2.3). In step 1, the notice that appears after the write does not move the keyboard (AC-1.6).
8. Open a crafted person whose record links read did not count (S16-01's handler fixture past `MaxReadWork`, or the `#Preview("Not counted")`): nodes, no arc, the note in place of the caption (AC-4.1); the previews "Lines dropped", "1,001 graphs" and "Both" show the other note (AC-4.2, AC-4.3, AC-4.4).
9. Steps 1, 2 and 8 at Regular and at Wide (pane ≥ 1600 pt): notices keep the band's and the column's widths; resizing across 1600 keeps them and the keyboard (PN-W1, PN-W2).

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| The inner `.id` remount drops the keyboard before `FocusClaim` writes it | `FocusClaim` writes twice across two main-queue turns; manual check 3 |
| A `.gone` behind a value now leaves a page that used to stay stale; a transient not-found could eject the researcher | Only the engine code `conclusiondetails.not_found` reads as gone, as for a first load today; any other error is a notice |
| `PersonDetailSections` and the three sections change signature beside S16-06's edits | S16-06 rebases; the map lists the shared file |

**Stop and report if:**

- `QueryHandle.applyFailure` is not the only place `status` becomes `.error`, or a success path leaves `error` set (D-1's assumption).
- `session.reread` on a key in `.error` with a value does not keep the value while it reads (`ensureQuery` `hadStaleValue`).
- The agreed board places a notice somewhere D-5, D-7 or D-10 does not, or adds a control the brief's inventory does not list.
- `CircularLinkDiagram` cannot draw a dataset with nodes and no arcs at the circle's size.
- `queryReadAndProgressReads` needs a change to pass (that would mean `read` changed).

## 12. Open questions

- **Board S16-D1.** The plan is blocked until the board is agreed. Recommended: the brief as written; on agreement, copy any changed placement or copy into D-5, D-7, D-10 and **Strings**.

## 13. PR

**Title:** S16-04 — The Person page says when a reload failed or the diagram is partial
**Body:** objective in one sentence; what shipped against US-1 – US-4; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-16/prs/S16-04-person-page-notices.md`.
