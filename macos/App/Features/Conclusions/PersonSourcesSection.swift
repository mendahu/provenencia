import SwiftUI

/// The Person page's Sources section (S13-16, board S13-D6): the Sources the person's
/// records come from, one row each, with a line for every record on that Source's Evidence
/// graphs. It declares its reads in `ConclusionPageSection.reads` (its own claims, the record
/// links as companion, the Source types as supporting), keeps the rows drawn while it reloads,
/// opens a Source on its page and a record on its graph, with that Subject selected.
struct PersonSourcesSection: View {
    let entityId: String
    let userID: String
    let pageIdentifier: String

    @Environment(WorkspaceSession.self) private var session
    @Environment(WorkspaceNavigation.self) private var navigation

    /// The pages of each Source's lines and of the Sources the researcher has unfolded; another
    /// person's page is another view, so it starts folded on every mount.
    @State private var folds = ConclusionFoldState<PersonSourcesContent.Fold>()
    /// The Unpromote confirm; another person's page is another view, so it starts fresh.
    @State private var unpromote = UnpromoteFlow()

    private var root: String { ConclusionPageSection.sources.accessibilityIdentifierRoot(pageIdentifier: pageIdentifier) }

    var body: some View {
        let orders = session.nameFormat
        let folds = folds
        let typesHandle: QueryHandle<[CatalogSourceType]>? = ConclusionPageSection.sources
            .read(.supporting, project: session.projectKey, entityId: entityId)
            .flatMap { session.queryHandle($0) }
        ConclusionPageSectionHost<[CatalogPersonClaimGroup], PersonSourcesContent, PersonSourcesList>(
            section: .sources,
            entityId: entityId,
            accessibilityIdentifierRoot: root,
            errorTitle: L10n.Conclusions.sourcesErrorTitle,
            emptyTitle: L10n.Conclusions.sourcesEmptyTitle,
            emptyMessage: L10n.string(L10n.Conclusions.sourcesEmptyMessage),
            make: {
                PersonSourcesContent(
                    groups: $0,
                    sourceTypes: typesHandle?.value ?? [],
                    orders: orders,
                    folds: folds
                )
            },
            isEmpty: \.isEmpty
        ) { content in
            PersonSourcesList(
                content: content,
                entityId: entityId,
                accessibilityIdentifierRoot: root,
                open: { navigation.go(to: $0) },
                expand: { self.folds.revealPage($0) },
                unpromote: unpromote,
                beginUnpromote: begin,
                confirmUnpromote: confirm
            )
        }
    }

    /// Opens the Unpromote confirm for a record's Subject.
    private func begin(_ line: PersonSourcesContent.Line) {
        Task {
            await unpromote.open(
                subjectID: line.subjectID,
                name: line.subjectName,
                kind: .person,
                handleID: entityId,
                graphID: line.graphID,
                session: session
            )
        }
    }

    /// Writes the Unpromote. The flow decides leave or stay: when the write deleted this person,
    /// the app leaves for the Persons list as the session applies the effects, so nothing reloads
    /// a handle that is gone; either way it says what happened.
    private func confirm() async -> Bool {
        await unpromote.confirm(session: session, userID: userID, leavingPage: navigation.fallbackToSectionRoot)
    }
}

/// What takes keyboard focus: a record's graph link, its Unpromote button, or a node of the
/// record links diagram.
enum PersonSourcesFocus: Hashable {
    case graph(claimID: String)
    case unpromote(claimID: String)
    case node(graphID: String)
}

/// The widths the header line and every Source row share: Citation (fills) · Type · Date ·
/// Ref, with one cell modifier per column so they cannot drift.
enum PersonSourcesColumns {
    static let spacing = PVProvenanceColumns.spacing
    static let typeWidth: CGFloat = 176
    static let dateWidth: CGFloat = 120
    static let refWidth: CGFloat = 104
    /// A graph line's indent under the title; the line's own 4pt padding makes the text 20pt in.
    static let lineIndent: CGFloat = 16
    /// The columns that never shrink (Type, Date, Ref and the three gaps between the four).
    static let fixedWidth: CGFloat = typeWidth + dateWidth + refWidth + 3 * spacing
    /// The narrowest the Citation column is drawn.
    static let citationMin: CGFloat = 200
    /// The narrowest a row is drawn; a pane with less room puts the diagram under the rows.
    static let minWidth: CGFloat = fixedWidth + citationMin
}

private extension View {
    /// The record line's highlight: a node of the record links diagram is pointed at or focused.
    /// A Show more standing for folded lines wears it too (S16-D2 FH-1).
    func personSourcesHighlight(_ on: Bool) -> some View {
        background(
            on ? PVColor.surfaceSelected : Color.clear,
            in: RoundedRectangle(cornerRadius: PVRadius.sm, style: .continuous)
        )
    }

    func personSourcesCitationCell() -> some View {
        frame(maxWidth: .infinity, alignment: .leading)
    }

    func personSourcesTypeCell() -> some View {
        frame(width: PersonSourcesColumns.typeWidth, alignment: .leading)
    }

    func personSourcesDateCell() -> some View {
        frame(width: PersonSourcesColumns.dateWidth, alignment: .leading)
    }

    func personSourcesRefCell() -> some View {
        frame(width: PersonSourcesColumns.refWidth, alignment: .leading)
    }
}

/// The header line and the Source rows, with the record links diagram beside them, pinned
/// while they scroll (`ConclusionPinnedSideColumn`). It owns the interaction state of both:
/// the one keyboard focus and the one hovered graph, so a node and a graph line highlight
/// each other.
struct PersonSourcesList: View {
    let content: PersonSourcesContent
    let entityId: String
    let accessibilityIdentifierRoot: String
    let open: (WorkspaceLocation) -> Void
    let expand: (PersonSourcesContent.Fold) -> Void
    let unpromote: UnpromoteFlow
    let beginUnpromote: (PersonSourcesContent.Line) -> Void
    let confirmUnpromote: () async -> Bool

    /// The graph link or the Unpromote button with the keyboard.
    @FocusState private var focus: PersonSourcesFocus?
    /// The line to focus once the Unpromote's reload drops the removed one: captured before the
    /// write, so a reload that beats the write's return cannot miss it.
    @State private var pendingFocus: (subjectID: String, target: String?)?
    /// The Unpromote button that had the keyboard when its confirm opened, so Keep gives it back.
    /// Nil after a mouse click, which draws no focus ring.
    @State private var returnFocus: PersonSourcesFocus?
    /// The graph the pointer is on, on its node or on any part of its line. Read by the rows, the
    /// nodes and the highlighted arcs and not by this view, so a hover rebuilds none of the diagram's
    /// content or layout.
    @State private var hover = CircularLinkHover<String>()
    /// The line a keyboard Show more just revealed, to scroll into view.
    @State private var scrollTarget: String?
    @State private var scrollRequest = 0

    @Environment(WorkspaceSession.self) private var session

    /// The graph of what has the keyboard; the hovered graph, when there is one, wins over it.
    private var focusedGraphID: String? {
        switch focus {
        case .graph(let claimID), .unpromote(let claimID): return content.claimGraphs[claimID]
        case .node(let graphID): return graphID
        case nil: return nil
        }
    }

    var body: some View {
        ConclusionPinnedSideColumn(mainMin: PersonSourcesColumns.minWidth) {
            rows
        } side: {
            diagram
        }
        .unpromoteConfirm(
            flow: unpromote,
            accessibilityIdentifierPrefix: "\(accessibilityIdentifierRoot).unpromote",
            onConfirm: confirmAndTrackFocus
        )
        .onChange(of: content) { _, new in
            applyPendingFocus(in: new)
            // A line that left or moved never reports the pointer gone; clear what it held.
            if let id = hover.id, !new.claimGraphs.values.contains(id) { hover.id = nil }
        }
        .onChange(of: unpromote.pending?.id) { _, id in
            if id == nil { restoreFocusAfterKeep() }
        }
    }

    /// The diagram column: a second host over the section's companion read, so the rows stay
    /// drawn while it loads, fails or reloads.
    private var diagram: some View {
        let focused = focusedGraphID
        let hover = hover
        return ConclusionPageSectionHost<CatalogPersonRecordLinks, PersonRecordLinksContent, PersonRecordLinksColumn>(
            section: .sources,
            entityId: entityId,
            accessibilityIdentifierRoot: accessibilityIdentifierRoot,
            read: .companion,
            placement: .column(
                loadingHeight: CircularLinkDiagramStyle.standard.minimumHeight
            ),
            errorTitle: L10n.Conclusions.recordLinksErrorTitle,
            make: {
                PersonRecordLinksContent(
                    links: $0,
                    graphLocations: content.graphLocations,
                    graphSourceRefs: content.graphSourceRefs
                )
            }
        ) { links in
            PersonRecordLinksColumn(
                content: links,
                accessibilityIdentifierRoot: accessibilityIdentifierRoot,
                hover: hover,
                focusedGraphID: focused,
                open: { id in
                    if let location = links.locations[id] { open(location) }
                },
                focus: { PVFieldFocus($focus, equals: .node(graphID: $0)) }
            )
        }
    }

    private var rows: some View {
        VStack(alignment: .leading, spacing: 0) {
            header
            PVDivider(color: PVColor.borderStrong)
            ForEach(Array(content.rows.enumerated()), id: \.element.id) { index, row in
                if index > 0 { PVDivider() }
                let fold = PersonSourcesContent.Fold.lines(sourceID: row.id)
                PersonSourcesRowView(
                    row: row, accessibilityIdentifierRoot: accessibilityIdentifierRoot, open: open, expand: { unfold(fold) }, focused: $focus,
                    revealed: content.firstRevealed(by: fold).map { .graph(claimID: $0) },
                    foldedGraphIDs: content.foldedGraphIDs(by: fold),
                    hover: hover,
                    focusedGraphID: focusedGraphID,
                    unpromoteDisabled: unpromote.isBusy,
                    beginUnpromote: { line in
                        returnFocus = focus == .unpromote(claimID: line.id) ? focus : nil
                        beginUnpromote(line)
                    }
                )
            }
            if content.hiddenRowCount > 0 {
                PersonSourcesRowsShowMore(
                    hidden: content.hiddenRowCount,
                    focus: $focus,
                    revealed: content.firstRevealed(by: .rows).map { .graph(claimID: $0) },
                    foldedGraphIDs: content.foldedGraphIDs(by: .rows),
                    hover: hover,
                    focusedGraphID: focusedGraphID
                ) { unfold(.rows) }
                .padding(.vertical, PVSpacing.space3)
                .accessibilityIdentifier("\(accessibilityIdentifierRoot).more")
            }
        }
        .pvScrollsRowIntoView(scrollTarget, whenRequested: scrollRequest)
    }

    /// Notes where focus goes after the removed line, then writes.
    private func confirmAndTrackFocus() {
        guard !unpromote.isRunning, let subjectID = unpromote.pending?.id else { return }
        pendingFocus = (subjectID, content.lineAfterRemoving(subjectID: subjectID))
        Task { if await !confirmUnpromote() { pendingFocus = nil } }
    }

    /// The confirm closed: after Keep promotion, focus returns to the button that opened it. After a
    /// write, `applyPendingFocus` places it.
    private func restoreFocusAfterKeep() {
        defer { returnFocus = nil }
        guard pendingFocus == nil, let returnFocus,
              case .unpromote(let claimID) = returnFocus,
              content.rows.contains(where: { $0.lines.contains { $0.id == claimID } })
        else { return }
        FocusClaim.claim($focus, returnFocus)
    }

    /// Moves focus to the next line's Unpromote once the content no longer holds the removed line.
    private func applyPendingFocus(in new: PersonSourcesContent) {
        guard let pending = pendingFocus,
              !new.rows.contains(where: { $0.lines.contains { $0.subjectID == pending.subjectID } })
        else { return }
        pendingFocus = nil
        if let target = pending.target {
            FocusClaim.claim($focus, .unpromote(claimID: target))
        }
    }

    /// Unfolds a page; the Show more row moves the keyboard to the first line it adds, and the
    /// list scrolls that line just into view.
    private func unfold(_ fold: PersonSourcesContent.Fold) {
        let revealed = content.firstRevealed(by: fold)
        expand(fold)
        scrollTarget = revealed
        scrollRequest += 1
    }

    private var header: some View {
        HStack(alignment: .firstTextBaseline, spacing: PersonSourcesColumns.spacing) {
            Text(L10n.Conclusions.sourcesColumnSource).personSourcesCitationCell()
            Text(L10n.Conclusions.sourcesColumnType).personSourcesTypeCell()
            Text(L10n.Conclusions.sourcesColumnDate).personSourcesDateCell()
            Text(L10n.Conclusions.sourcesColumnRef).personSourcesRefCell()
        }
        .pvMicroCaps()
        .foregroundStyle(PVColor.textMuted)
        .padding(.vertical, PVSpacing.space2)
        .accessibilityHidden(true)
    }
}

/// The "Show N more" row under the Sources, which takes the highlight when the diagram's
/// highlighted graph is on a Source behind it (S16-D2 FH-2). Its own view so only it reads the
/// hover: a pointer move rebuilds the list's body for nothing else.
private struct PersonSourcesRowsShowMore: View {
    let hidden: Int
    var focus: FocusState<PersonSourcesFocus?>.Binding
    let revealed: PersonSourcesFocus?
    /// The graphs of the Sources the fold leaves out.
    let foldedGraphIDs: Set<String>
    let hover: CircularLinkHover<String>
    let focusedGraphID: String?
    let expand: () -> Void

    var body: some View {
        let highlighted = hover.highlighted(focused: focusedGraphID).map(foldedGraphIDs.contains) ?? false
        ConclusionShowMoreRow(
            hidden: hidden,
            fold: PersonSourcesContent.rowFold,
            focus: focus,
            revealed: revealed,
            expand: expand
        )
        .frame(maxWidth: .infinity, alignment: .leading)
        .personSourcesHighlight(highlighted)
    }
}

/// One Source: its title as a link, its type, date and ref, and under the title a line per record.
private struct PersonSourcesRowView: View {
    let row: PersonSourcesContent.Row
    let accessibilityIdentifierRoot: String
    let open: (WorkspaceLocation) -> Void
    let expand: () -> Void
    var focused: FocusState<PersonSourcesFocus?>.Binding
    /// The graph link the next page of this Source's lines starts with; nil when it folds nothing.
    let revealed: PersonSourcesFocus?
    /// The graphs whose lines are all behind this Source's Show more.
    let foldedGraphIDs: Set<String>
    let hover: CircularLinkHover<String>
    let focusedGraphID: String?
    let unpromoteDisabled: Bool
    let beginUnpromote: (PersonSourcesContent.Line) -> Void

    var body: some View {
        VStack(alignment: .leading, spacing: PVSpacing.space2) {
            columns
            lines
        }
        // The row as one element; the title link and the lines follow as controls.
        .overlay(alignment: .topLeading) {
            Color.clear.frame(width: 1, height: 1)
                .accessibilityElement(children: .ignore)
                .accessibilityLabel(row.accessibilityLabel)
        }
        .padding(.vertical, PVSpacing.space3)
        .accessibilityElement(children: .contain)
        .accessibilityIdentifier("\(accessibilityIdentifierRoot).row.\(row.id)")
    }

    private var columns: some View {
        HStack(alignment: .firstTextBaseline, spacing: PersonSourcesColumns.spacing) {
            title.personSourcesCitationCell()
            typeCell.personSourcesTypeCell().accessibilityHidden(true)
            Text(L10n.Conclusions.sourcesDateNone)
                .font(PVFont.body(size: PVTypeScale.caption))
                .foregroundStyle(PVColor.textMuted)
                .personSourcesDateCell()
                .accessibilityHidden(true)
            Text(verbatim: row.ref)
                .font(PVFont.mono(size: PVTypeScale.caption))
                .foregroundStyle(PVColor.textSecondary)
                .textSelection(.enabled)
                .personSourcesRefCell()
                .accessibilityHidden(true)
        }
    }

    private var title: some View {
        PVInkLink(ink: .primary, action: { open(row.location) }) {
            PVTitleText(row.title, style: .rowLink, lineLimit: 2)
        }
        .accessibilityLabel(Text(verbatim: L10n.Conclusions.sourcesOpenSource(row.helpTitle)))
        .accessibilityIdentifier("\(accessibilityIdentifierRoot).title.\(row.id)")
        .help(row.helpTitle)
    }

    /// The record lines span the whole row less their indent.
    private var lines: some View {
        let highlightedGraphID = hover.highlighted(focused: focusedGraphID)
        return VStack(alignment: .leading, spacing: PVSpacing.space2) {
            ForEach(row.lines) { line in
                PersonSourcesLineView(
                    line: line, accessibilityIdentifierRoot: accessibilityIdentifierRoot, open: open, focused: focused,
                    highlighted: line.graphID == highlightedGraphID, hover: { hover.id = $0 },
                    unpromoteDisabled: unpromoteDisabled, beginUnpromote: beginUnpromote
                )
                .id(line.id)
            }
            if row.hiddenCount > 0 {
                ConclusionShowMoreRow(
                    hidden: row.hiddenCount,
                    fold: PersonSourcesContent.lineFold,
                    focus: focused,
                    revealed: revealed,
                    expand: expand
                )
                .frame(maxWidth: .infinity, alignment: .leading)
                .personSourcesHighlight(highlightedGraphID.map(foldedGraphIDs.contains) ?? false)
                .accessibilityIdentifier("\(accessibilityIdentifierRoot).more.\(row.id)")
            }
        }
        .padding(.leading, PersonSourcesColumns.lineIndent)
    }

    @ViewBuilder
    private var typeCell: some View {
        if let label = row.typeLabel {
            CatalogSourceTypePill(label: label, iconKey: row.typeIconKey)
        } else {
            Color.clear.frame(height: 1)
        }
    }
}

/// One record: the graph's name as a link, then "· graph ref · Subject name · Subject ref".
/// Every part stays on one line; a line too narrow cuts the Subject's name first.
private struct PersonSourcesLineView: View {
    let line: PersonSourcesContent.Line
    let accessibilityIdentifierRoot: String
    let open: (WorkspaceLocation) -> Void
    var focused: FocusState<PersonSourcesFocus?>.Binding
    /// The line's graph is the one highlighted, from a node or from the line itself.
    let highlighted: Bool
    let hover: (String?) -> Void
    let unpromoteDisabled: Bool
    let beginUnpromote: (PersonSourcesContent.Line) -> Void

    var body: some View {
        HStack(alignment: .firstTextBaseline, spacing: 0) {
            PVButton(line.graphTitle, variant: .link, size: .sm) { open(line.location) }
                .lineLimit(1)
                .truncationMode(.tail)
                .help(line.graphTitle)
                .focused(focused, equals: .graph(claimID: line.id))
                .pvActivatesOnReturn(when: focused.wrappedValue == .graph(claimID: line.id)) { open(line.location) }
                .accessibilityLabel(line.accessibilityLabel)
                .accessibilityIdentifier("\(accessibilityIdentifierRoot).graph.\(line.id)")
            Group {
                if let graphRef = line.graphRef {
                    Text(verbatim: " · ")
                    Text(verbatim: graphRef).font(PVFont.mono(size: PVTypeScale.caption))
                        .lineLimit(1).fixedSize(horizontal: true, vertical: false)
                }
                Text(verbatim: " · ")
                Text(verbatim: line.subjectName)
                    .lineLimit(1)
                    .truncationMode(.tail)
                    .layoutPriority(-1)
                    .help(line.subjectName)
                Text(verbatim: " · ")
                Text(verbatim: line.subjectRef).font(PVFont.mono(size: PVTypeScale.caption))
                    .lineLimit(1).fixedSize(horizontal: true, vertical: false)
            }
            .font(PVFont.body(size: PVTypeScale.caption))
            .foregroundStyle(PVColor.textSecondary)
            .accessibilityHidden(true)
            PVButton(L10n.Unpromote.button, variant: .ghost, size: .sm, icon: .unpromote) {
                beginUnpromote(line)
            }
            .fixedSize()
            .padding(.leading, PVSpacing.space3)
            .disabled(unpromoteDisabled)
            .focused(focused, equals: .unpromote(claimID: line.id))
            .pvActivatesOnReturn(
                when: focused.wrappedValue == .unpromote(claimID: line.id) && !unpromoteDisabled
            ) { beginUnpromote(line) }
            .accessibilityLabel(line.unpromoteAccessibilityLabel)
            .accessibilityIdentifier("\(accessibilityIdentifierRoot).unpromote.\(line.id)")
        }
        .frame(maxWidth: .infinity, alignment: .leading)
        .padding(.vertical, 2)
        .padding(.horizontal, PVSpacing.space2)
        .personSourcesHighlight(highlighted)
        .contentShape(Rectangle())
        .onHover { hover($0 ? line.graphID : nil) }
        .accessibilityElement(children: .contain)
        .accessibilityIdentifier("\(accessibilityIdentifierRoot).line.\(line.id)")
    }
}

#if DEBUG
/// The Sources section over a fixture, with its folds working and its reads served by a
/// preview session. With `claimsFailure` or `claimsGone` it draws the whole section, so the
/// rows' own read shows its notice.
private struct PersonSourcesPreview: View {
    static let entityId = "preview-person"

    let groups: [CatalogPersonClaimGroup]
    var claimsFailure: Bool
    var claimsGone: Bool
    var width: CGFloat
    var foldsAtStart: ConclusionFoldState<PersonSourcesContent.Fold>

    @State private var unpromote = UnpromoteFlow()
    @State private var folds = ConclusionFoldState<PersonSourcesContent.Fold>()
    @State private var session: WorkspaceSession
    @State private var navigation = PersonSourcesPreview.makeNavigation()

    init(
        groups: [CatalogPersonClaimGroup],
        recordLinks: CatalogPersonRecordLinks? = nil,
        failure: Bool = false,
        claimsFailure: Bool = false,
        claimsGone: Bool = false,
        width: CGFloat = 1102,
        foldsAtStart: ConclusionFoldState<PersonSourcesContent.Fold> = ConclusionFoldState<PersonSourcesContent.Fold>()
    ) {
        self.groups = groups
        self.claimsFailure = claimsFailure
        self.claimsGone = claimsGone
        self.width = width
        self.foldsAtStart = foldsAtStart
        _session = State(initialValue: PersonSourcesPreviewData.previewSession(
            entityId: Self.entityId,
            recordLinks: recordLinks,
            failure: failure,
            claims: claimsFailure ? groups : nil
        ))
    }

    /// The Person location the section's host reads for its kind's words.
    private static func makeNavigation() -> WorkspaceNavigation {
        let navigation = WorkspaceNavigation()
        navigation.go(to: WorkspaceLocation.personDetail(
            entityId: entityId, ref: "PER-7KD45", title: nil
        ).withPageSection(.sources))
        return navigation
    }

    var body: some View {
        PVPageColumns(primaryFill: true) {
            if claimsFailure || claimsGone {
                PersonSourcesSection(entityId: Self.entityId, userID: "preview-user", pageIdentifier: "preview")
            } else {
                PersonSourcesList(
                    content: PersonSourcesContent(
                        groups: groups,
                        sourceTypes: PersonSourcesPreviewData.types,
                        orders: NameFormatOrders(),
                        folds: folds
                    ),
                    entityId: Self.entityId,
                    accessibilityIdentifierRoot: "preview",
                    open: { _ in },
                    expand: { folds.revealPage($0) },
                    unpromote: unpromote,
                    beginUnpromote: { _ in },
                    confirmUnpromote: { false }
                )
            }
        }
        .environment(session)
        .environment(navigation)
        .environment(\.pvWidthClass, width >= PVSpacing.widthWideBreakpoint ? .wide : .regular)
        .frame(width: width, height: 700)
        .background(PVColor.surfacePage)
        .onAppear { folds = foldsAtStart }
        .task { await failClaimsAfterTheFirstAnswer() }
    }

    /// "Sources stale": the rows are read once, then the next read fails and is asked for.
    private func failClaimsAfterTheFirstAnswer() async {
        guard claimsFailure else { return }
        let key = CatalogQueryKey.personClaims(project: session.projectKey, entityId: Self.entityId)
        await PersonSourcesPreviewData.failAfterFirstAnswer(
            key, as: [CatalogPersonClaimGroup].self, session: session
        ) { $0.listPersonClaimsError = PersonSourcesPreviewData.previewFailure }
    }
}

#Preview("Sources") { PersonSourcesPreview(groups: PersonSourcesPreviewData.regular) }
#Preview("Johann") {
    PersonSourcesPreview(
        groups: PersonSourcesPreviewData.johannClaims, recordLinks: PersonSourcesPreviewData.johannRecordLinks
    )
}
#Preview("Johann, Wide") {
    PersonSourcesPreview(
        groups: PersonSourcesPreviewData.johannClaims, recordLinks: PersonSourcesPreviewData.johannRecordLinks,
        width: 2296
    )
}
#Preview("Many") {
    PersonSourcesPreview(
        groups: PersonSourcesPreviewData.manyClaims(graphs: 50),
        recordLinks: PersonSourcesPreviewData.manyRecordLinks(graphs: 50)
    )
}
#Preview("Diagram loading") { PersonSourcesPreview(groups: PersonSourcesPreviewData.regular) }
#Preview("Diagram error") { PersonSourcesPreview(groups: PersonSourcesPreviewData.regular, failure: true) }
#Preview("Not counted") {
    PersonSourcesPreview(
        groups: PersonSourcesPreviewData.manyClaims(graphs: 40),
        recordLinks: PersonSourcesPreviewData.notCounted(PersonSourcesPreviewData.manyRecordLinks(graphs: 40))
    )
}
#Preview("Lines dropped") {
    PersonSourcesPreview(
        groups: PersonSourcesPreviewData.manyClaims(graphs: 50),
        recordLinks: PersonSourcesPreviewData.linesDropped(PersonSourcesPreviewData.manyRecordLinks(graphs: 50))
    )
}
#Preview("1,001 graphs") {
    PersonSourcesPreview(
        groups: PersonSourcesPreviewData.crowdClaims(graphs: 1_001),
        recordLinks: PersonSourcesPreviewData.crowdRecordLinks(graphs: 1_001)
    )
}
#Preview("Both") {
    PersonSourcesPreview(
        groups: PersonSourcesPreviewData.crowdClaims(graphs: 1_001),
        recordLinks: PersonSourcesPreviewData.notCounted(PersonSourcesPreviewData.crowdRecordLinks(graphs: 1_001))
    )
}
#Preview("Sources stale") {
    PersonSourcesPreview(
        groups: PersonSourcesPreviewData.regular, recordLinks: PersonSourcesPreviewData.johannRecordLinks,
        claimsFailure: true
    )
}
#Preview("Sources gone") {
    PersonSourcesPreview(groups: PersonSourcesPreviewData.regular, claimsGone: true)
}
#Preview("Untitled Source") { PersonSourcesPreview(groups: PersonSourcesPreviewData.untitled) }
#Preview("Folded") { PersonSourcesPreview(groups: PersonSourcesPreviewData.folded) }
#Preview("Many Sources") { PersonSourcesPreview(groups: PersonSourcesPreviewData.manySources) }
#endif
