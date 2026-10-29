import SwiftUI

/// The Promote place (S9-44): one page, laid out as the S9-D16 board.
struct PromoteView: View {
    @Environment(WorkspaceNavigation.self) private var navigation
    @State private var model: PromoteModel
    @Environment(\.pvWidthClass) private var widthClass

    init(
        entry: PromoteEntry,
        session: WorkspaceSession,
        store: any GenealogyStore,
        userID: String
    ) {
        _model = State(initialValue: PromoteModel(
            entry: entry,
            session: session,
            store: store,
            userID: userID
        ))
    }

    var body: some View {
        let model = model
        let connections = model.flow.connectionsByRow
        VStack(spacing: 0) {
            header
            ScrollView {
                VStack(alignment: .leading, spacing: PVSpacing.space7) {
                    notices
                    if model.flow.mappedRest && !model.entry.mapAll {
                        openedFromNote
                    }
                    ForEach(sections, id: \.id) { section in
                        sectionBlock(section, connections: connections)
                    }
                    if !model.flow.mappedRest && model.flow.restCount > 0 {
                        mapRest
                    }
                }
                .padding(.horizontal, PVSpacing.space10)
                .padding(.vertical, PVSpacing.space6)
                .frame(maxWidth: PromoteLayout.listMax(widthClass), alignment: .leading)
                .frame(maxWidth: .infinity, alignment: .center)
            }
            footer
        }
        .background(PVColor.surfacePage)
        .onAppear {
            model.navigation = navigation
            navigation.leaveGuard = model
        }
        .onDisappear {
            if navigation.leaveGuard === model {
                navigation.leaveGuard = nil
            }
        }
        .task {
            let _: QueryHandle<EvidenceGraphRows> = model.session.query(model.graphKey)
            let _: QueryHandle<PropertiesSnapshot> = model.session.query(model.propertiesKey)
            let _: QueryHandle<[CatalogConnectRule]> = model.session.query(model.rulesKey)
            let _: QueryHandle<[CatalogSource]> = model.session.query(model.sourcesKey)
            let _: QueryHandle<[CatalogSourceType]> = model.session.query(model.sourceTypesKey)
            await model.load()
        }
        .pvConfirm(
            item: leaveBinding,
            copy: { _ in
                PVConfirmCopy(
                    title: L10n.string(L10n.Promote.leaveTitle),
                    message: L10n.Promote.leaveDetail(rows: model.flow.rows.filter(\.decided).count),
                    confirm: L10n.Promote.leaveConfirm,
                    cancel: L10n.Promote.leaveCancel
                )
            },
            tone: .danger,
            accessibilityIdentifierPrefix: "promote.leave",
            onConfirm: { model.leave() },
            detail: { _ in EmptyView() }
        )
    }

    private var notices: some View {
        VStack(alignment: .leading, spacing: PVSpacing.space4) {
            if let error = model.loadError {
                PVCallout(tone: .danger, message: error)
            }
            if model.flow.staleNote != nil {
                PVCallout(
                    tone: .danger,
                    title: L10n.Promote.staleTitle,
                    message: L10n.string(L10n.Promote.stale)
                )
            }
            if let error = model.proposeError {
                PVCallout(tone: .danger, message: error)
            }
            if let error = model.saveError {
                PVCallout(tone: .danger, message: error)
            }
        }
    }

    private var header: some View {
        HStack(alignment: .center, spacing: PVSpacing.space5) {
            RoundedRectangle(cornerRadius: PVRadius.sm, style: .continuous)
                .fill(PVColor.surfaceSunken)
                .overlay(RoundedRectangle(cornerRadius: PVRadius.sm, style: .continuous).strokeBorder(PVColor.borderSubtle, lineWidth: 1))
                .overlay(PVMark(model.sourceMark, size: 24, decorative: true).foregroundStyle(PVColor.textSecondary))
                .frame(width: 44, height: 44)
            VStack(alignment: .leading, spacing: PVSpacing.space2) {
                Text(L10n.Promote.pageTitle)
                    .pvMicroCaps()
                    .foregroundStyle(PVColor.textMuted)
                if let title = model.entry.sourceTitle {
                    Text(verbatim: title)
                        .font(PVFont.display(size: PVTypeScale.h3, weight: PVFontWeight.semibold))
                        .foregroundStyle(PVColor.textDisplay)
                }
                if !model.sourceRef.isEmpty {
                    Text(verbatim: model.sourceRef)
                        .font(PVFont.mono(size: PVTypeScale.caption))
                        .foregroundStyle(PVColor.textMuted)
                }
            }
            Spacer(minLength: PVSpacing.space6)
            VStack(alignment: .trailing, spacing: PVSpacing.space1) {
                Text(verbatim: promoting)
                    .font(PVFont.body(size: PVTypeScale.bodySmall, weight: PVFontWeight.medium))
                Text(verbatim: countLine)
                    .font(PVFont.mono(size: PVTypeScale.caption))
                    .foregroundStyle(PVColor.textMuted)
            }
        }
        .padding(.horizontal, PVSpacing.space10)
        .padding(.vertical, PVSpacing.space6)
        .background(PVColor.surfaceCard)
        .overlay(alignment: .bottom) { PVDivider() }
    }

    private var promoting: String {
        if model.flow.mappedRest {
            return L10n.Promote.promoting(count: model.flow.rows.count)
        }
        return model.entry.subjectName
    }

    private var countLine: String {
        if model.flow.mappedRest {
            return L10n.Promote.rowSummary(
                rows: model.flow.rows.count,
                filed: model.flow.rows.filter(\.anchor).count
            )
        }
        return L10n.Promote.singleSummary(more: model.flow.restCount)
    }

    private var openedFromNote: some View {
        Text(verbatim: L10n.Promote.openedFrom(name: model.entry.subjectName))
            .font(PVFont.body(size: PVTypeScale.bodySmall))
            .italic()
            .foregroundStyle(PVColor.textSecondary)
    }

    private var mapRest: some View {
        HStack(alignment: .center, spacing: PVSpacing.space4) {
            PVButton(L10n.Promote.mapRest(count: model.flow.restCount), variant: .secondary, icon: .network, action: model.mapRest)
                .disabled(model.isSaving)
                .accessibilityIdentifier("promote.mapRest")
            Text(L10n.Promote.mapRestHint)
                .font(PVFont.body(size: PVTypeScale.bodySmall))
                .italic()
                .foregroundStyle(PVColor.textSecondary)
        }
    }

    private func sectionBlock(
        _ section: Section,
        connections: [String: [PromoteFlow.RowConnection]]
    ) -> some View {
        VStack(alignment: .leading, spacing: 0) {
            PVSectionHeader(title: section.title, meta: section.meta)
            if !section.rows.isEmpty {
                columnHeaders
                ForEach(section.rows) { row in
                    PromoteAlignmentRow(
                        row: row,
                        model: model,
                        connections: connections[row.subjectID] ?? []
                    )
                    .disabled(model.isSaving)
                }
            }
            if !section.anchors.isEmpty {
                VStack(alignment: .leading, spacing: 0) {
                    Text(L10n.Promote.alreadyFiledHeading)
                        .pvMicroCaps()
                        .foregroundStyle(PVColor.textMuted)
                        .padding(.top, PVSpacing.space3)
                        .padding(.bottom, PVSpacing.space2)
                    ForEach(section.anchors) { row in
                        PromoteAnchorRow(row: row)
                    }
                }
                .padding(.horizontal, PVSpacing.space4)
                .background(PVColor.surfaceSunken)
            }
        }
    }

    private var columnHeaders: some View {
        PromoteGridRow(showsArrow: false) {
            Color.clear.frame(height: 1)
        } source: {
            columnLabel(L10n.Promote.columnSource)
        } target: { _ in
            columnLabel(L10n.Promote.columnTarget)
        } freed: {
            Color.clear.frame(height: 1)
        } badges: {
            Color.clear.frame(height: 1)
        }
        .padding(.top, PVSpacing.space5)
        .padding(.bottom, PVSpacing.space3)
    }

    private func columnLabel(_ title: LocalizedStringResource) -> some View {
        Text(title)
            .pvMicroCaps()
            .foregroundStyle(PVColor.textMuted)
            .frame(maxWidth: .infinity, alignment: .leading)
    }

    private var footer: some View {
        HStack(spacing: PVSpacing.space4) {
            Spacer()
            Text(verbatim: L10n.Promote.doneSummary(
                file: model.flow.fileCount,
                skip: model.flow.skipCount,
                unselected: model.flow.unsetCount,
                connections: model.flow.connectionCount
            ))
            .font(PVFont.mono(size: PVTypeScale.caption))
            .foregroundStyle(PVColor.textSecondary)
            PVButton(L10n.Promote.cancel, variant: .secondary) { model.leave() }
                .disabled(model.isSaving)
                .accessibilityIdentifier("promote.cancel")
            PVButton(model.isSaving ? L10n.Promote.filing : L10n.Promote.done, variant: .primary, loading: model.isSaving) {
                model.done()
            }
            .disabled(!model.canFinish)
            .accessibilityIdentifier("promote.done")
        }
        .padding(.horizontal, PVSpacing.space10)
        .padding(.vertical, PVSpacing.space4)
        .background(PVColor.surfaceCard)
        .overlay(alignment: .top) { PVDivider() }
    }

    private struct Section {
        var id: String
        var title: LocalizedStringResource
        var meta: String
        var rows: [PromoteFlow.Row]
        var anchors: [PromoteFlow.Row]
    }

    private var sections: [Section] {
        let working = model.flow.visibleRows.filter { !$0.anchor }
        let anchors = model.flow.mappedRest ? model.flow.rows.filter(\.anchor) : []
        return [
            kindSection("person", L10n.Promote.kindPerson, .person, working, anchors),
            kindSection("event", L10n.Promote.kindEvent, .event, working, anchors),
            kindSection("place", L10n.Promote.kindPlace, .place, working, anchors),
        ].filter { !$0.rows.isEmpty || !$0.anchors.isEmpty }
    }

    private func kindSection(
        _ id: String,
        _ title: LocalizedStringResource,
        _ kind: EvidencePrimaryKind,
        _ working: [PromoteFlow.Row],
        _ anchors: [PromoteFlow.Row]
    ) -> Section {
        let rows = working.filter { $0.kind == kind }
        let filed = anchors.filter { $0.kind == kind }
        let total = rows.count + filed.count
        let meta = filed.isEmpty
            ? L10n.Promote.sectionCount(total)
            : L10n.Promote.sectionCountFiled(rows: total, filed: filed.count)
        return Section(id: id, title: title, meta: meta, rows: rows, anchors: filed)
    }

    private var leaveBinding: Binding<PromoteModel.PendingLeave?> {
        Binding(
            get: { model.pendingLeave },
            set: { if $0 == nil { model.keepPromoting() } }
        )
    }
}

private struct PromoteAlignmentRow: View {
    let row: PromoteFlow.Row
    let model: PromoteModel
    /// The connections shown under this row, in display order.
    let connections: [PromoteFlow.RowConnection]
    /// Lets the target menu take focus back when the record search closes.
    @FocusState private var targetFocus: Bool?

    var body: some View {
        VStack(alignment: .leading, spacing: PVSpacing.space2) {
            PromoteGridRow {
                kindMark
            } source: {
                VStack(alignment: .leading, spacing: 0) {
                    Text(verbatim: row.name)
                        .font(PVFont.body(size: PVTypeScale.bodySmall, weight: PVFontWeight.medium))
                        .lineLimit(1)
                    Text(verbatim: row.ref)
                        .font(PVFont.mono(size: PVTypeScale.caption))
                        .foregroundStyle(PVColor.textMuted)
                }
            } target: { width in
                if model.recordSearch?.subjectID == row.subjectID {
                    PromoteRecordSearch(row: row, model: model)
                } else {
                    targetSelect(width: width)
                }
            } freed: {
                // Empty on a working row. S13-30 puts the claim's status control here.
                Color.clear.frame(height: 1)
            } badges: {
                HStack(spacing: PVSpacing.space2) {
                    if row.updated {
                        PVBadge(L10n.Promote.updated, tone: .info, icon: .refresh)
                    }
                    if row.decided {
                        PVBadge(L10n.Promote.decided, tone: .neutral, icon: .check, subtle: true)
                    }
                }
            }
            if row.updated, let caption = row.caption {
                PromoteUpdatedCaption(reason: model.reasonText(caption, on: row))
                    .padding(.leading, PromoteLayout.noteIndent)
            }
            if let ref = row.conflictNote {
                PVCallout(
                    tone: .warning,
                    title: L10n.Promote.conflictTitle(ref: ref),
                    message: L10n.string(L10n.Promote.conflictBody),
                    compact: true
                )
                .padding(.leading, PromoteLayout.noteIndent)
            }
            if let note = row.duplicateNote {
                PVCallout(
                    tone: .warning,
                    title: duplicateTitle(note),
                    message: L10n.string(L10n.Promote.duplicateBody),
                    compact: true
                )
                .padding(.leading, PromoteLayout.noteIndent)
            }
            if !connections.isEmpty {
                PromoteConnectionSubrows(rowName: row.name, rowRef: row.ref, connections: connections)
            }
        }
        .padding(.vertical, PVSpacing.space3)
        .overlay(alignment: .top) { PVDivider() }
        .onChange(of: model.recordSearch?.subjectID) { old, new in
            // Pick, Cancel and Escape all end here: the search unmounts, so hand focus
            // back to the menu that replaces it once it is on screen.
            if old == row.subjectID, new != row.subjectID {
                DispatchQueue.main.async { targetFocus = true }
            }
        }
    }

    private func duplicateTitle(_ note: PromoteFlow.DuplicateNote) -> String {
        switch note {
        case .sharesHandle(let other, let ref): L10n.Promote.duplicateTitle(name: other, ref: ref)
        case .alsoNew(let other): L10n.Promote.duplicateNewTitle(name: other)
        }
    }

    private func targetSelect(width: CGFloat) -> some View {
        PVSelect(
            selection: targetBinding,
            options: targetOptions,
            size: .sm,
            placeholder: row.target == .unset ? L10n.string(L10n.Promote.targetPlaceholder) : nil,
            menuWidth: width,
            fillsWidth: true,
            rowHeight: 46,
            isDisabled: model.isSaving,
            isInvalid: row.target == .unset,
            accessibilitySpokenLabel: L10n.Promote.filesOn(name: row.name),
            accessibilityIdentifier: "promote.row.\(row.ref).target",
            focus: PVSelectFocus($targetFocus, equals: true)
        ) { option in
            PromoteTargetMenuRow(option: option, row: row)
        }
    }

    private var kindMark: some View {
        PromoteKindTile(kind: row.kind, size: PromoteLayout.kindTile, markSize: 16, background: PVColor.surfaceSunken)
    }

    private var targetOptions: [PVSelectOption] {
        var options = row.menu.map { alt in
            PVSelectOption(value: PromoteFlow.Target.handleToken(alt.id), label: "\(alt.ref)  \(alt.title)")
        }
        options.append(PVSelectOption(value: PromoteFlow.Target.newKind.token, label: L10n.string(L10n.Promote.newOption(row.kind))))
        options.append(PVSelectOption(value: PromoteFlow.Target.skip.token, label: L10n.string(L10n.Promote.skip)))
        options.append(PVSelectOption(
            value: PromoteFlow.Target.findToken,
            label: L10n.string(L10n.Promote.findOption),
            accessibilityIdentifier: "promote.row.\(row.ref).target.find",
            isMenuOnly: true
        ))
        return options
    }

    private var targetBinding: Binding<String> {
        Binding(
            get: { row.target.token },
            set: { model.setTarget(subjectID: row.subjectID, token: $0) }
        )
    }
}

/// The connections that join a row, under its notes: a rule at the note indent and one line each,
/// with a second line for the other end or a self-link's warning.
private struct PromoteConnectionSubrows: View {
    let rowName: String
    let rowRef: String
    let connections: [PromoteFlow.RowConnection]

    var body: some View {
        VStack(alignment: .leading, spacing: 0) {
            ForEach(connections) { connection in
                subrow(connection)
            }
        }
        .padding(.leading, PVSpacing.space4)
        .overlay(alignment: .leading) {
            Rectangle().fill(PVColor.borderSubtle).frame(width: 1)
        }
        .padding(.leading, PromoteLayout.noteIndent)
        .accessibilityElement(children: .contain)
        .accessibilityLabel(Text(verbatim: L10n.Promote.connectionsOf(name: rowName)))
    }

    private func subrow(_ connection: PromoteFlow.RowConnection) -> some View {
        let bridge = connection.bridge
        let endName = bridge.name(of: bridge.state.waitingEndID)
        let files = bridge.state == .files
        return VStack(alignment: .leading, spacing: PVSpacing.space1) {
            HStack(spacing: PVSpacing.space3) {
                PVMark(bridge.mark, size: 14, decorative: true)
                    .foregroundStyle(PVColor.textMuted)
                Text(verbatim: bridge.sentence)
                    .font(PVFont.body(size: PVTypeScale.bodySmall))
                    .lineLimit(1)
                    .truncationMode(.tail)
                    .help(Text(verbatim: bridge.sentence))
                    .foregroundStyle(files ? PVColor.textPrimary : PVColor.textMuted)
                Spacer(minLength: PVSpacing.space4)
                if let text = L10n.Promote.connectionState(bridge.state, endName: endName) {
                    Text(verbatim: text)
                        .font(PVFont.body(size: PVTypeScale.caption))
                        .italic()
                        .foregroundStyle(files ? PVColor.textSecondary : PVColor.textMuted)
                }
            }
            if let other = connection.otherEnd {
                Text(verbatim: L10n.Promote.connectionOtherEnd(name: other.name, filed: other.filed))
                    .font(PVFont.body(size: PVTypeScale.caption))
                    .foregroundStyle(PVColor.textMuted)
                    .padding(.leading, 14 + PVSpacing.space3)
            }
            if case .selfLink(let ref) = bridge.state {
                PVCallout(tone: .warning, message: L10n.Promote.connectionSelfLink(ref: ref), compact: true)
            }
        }
        .padding(.vertical, PVSpacing.space1)
        .accessibilityElement(children: .ignore)
        .accessibilityLabel(Text(verbatim: L10n.Promote.connectionSpoken(sentence: bridge.sentence, state: bridge.state, endName: endName)))
        .accessibilityIdentifier("promote.row.\(rowRef).connection.\(bridge.id)")
    }
}

private struct PromoteTargetMenuRow: View {
    let option: PVSelectOption
    let row: PromoteFlow.Row

    var body: some View {
        if option.id == PromoteFlow.Target.newKind.token || option.id == PromoteFlow.Target.skip.token {
            Text(verbatim: option.label)
                .font(PVFont.body(size: PVTypeScale.bodySmall))
                .foregroundStyle(PVColor.textPrimary)
        } else if option.id == PromoteFlow.Target.findToken {
            HStack(spacing: PVSpacing.space3) {
                PVIcon(.search, size: 14)
                Text(verbatim: option.label)
                    .font(PVFont.body(size: PVTypeScale.bodySmall))
            }
            .foregroundStyle(PVColor.textSecondary)
        } else if let alt = row.menuItem(forToken: option.id) {
            HStack(spacing: PVSpacing.space3) {
                PVMark(row.kind.markKey, size: 16, decorative: true)
                    .foregroundStyle(PVColor.textSecondary)
                VStack(alignment: .leading, spacing: 0) {
                    Text(verbatim: alt.title)
                        .font(PVFont.body(size: PVTypeScale.bodySmall, weight: PVFontWeight.medium))
                        .lineLimit(1)
                    Text(verbatim: alt.subtitle.isEmpty ? alt.ref : "\(alt.ref) · \(alt.subtitle)")
                        .font(PVFont.mono(size: PVTypeScale.caption))
                        .foregroundStyle(PVColor.textMuted)
                        .lineLimit(1)
                }
            }
        }
    }
}

/// The target cell while its row searches every record of its kind: a search field over
/// `CatalogSearchModel` and a button that gives the menu back. A pick holds the record through
/// `PromoteModel.pickSearchHit`; nothing here decides a row.
private struct PromoteRecordSearch: View {
    let row: PromoteFlow.Row
    let model: PromoteModel
    @FocusState private var focus: Field?

    private enum Field { case query }

    var body: some View {
        if let search = model.recordSearch?.search {
            HStack(spacing: PVSpacing.space2) {
                PVComboBox(
                    selection: .constant(""),
                    options: search.hits.map {
                        PVComboBoxOption(value: $0.id, label: CatalogSearchHitPresentation.title(for: $0, orders: model.session.nameFormat))
                    },
                    size: .sm,
                    placeholder: L10n.Promote.searchPlaceholder(row.kind),
                    label: L10n.Promote.searchLabel(name: row.name),
                    accessibilityIdentifierPrefix: "promote.row.\(row.ref).search",
                    focus: PVComboFocus($focus, equals: .query),
                    onCommit: { id, _ in
                        model.pickSearchHit(id: id)
                        return false
                    },
                    onQueryChange: model.recordSearchQueryChanged,
                    isLoading: search.isLoading,
                    row: { option, _ in
                        if let hit = search.hits.first(where: { $0.id == option.value }) {
                            CatalogSearchHitRow(hit: hit, projectDir: model.session.projectKey.projectDir)
                        } else {
                            EmptyView()
                        }
                    },
                    empty: { _ in
                        Text(verbatim: emptyLine(search))
                    }
                )
                .onAppear { DispatchQueue.main.async { focus = .query } }
                PVIconButton(.dismiss, label: L10n.Promote.searchCancel, size: .sm) {
                    model.closeRecordSearch()
                }
                .accessibilityIdentifier("promote.row.\(row.ref).search.cancel")
            }
            .onExitCommand { model.closeRecordSearch() }
        }
    }

    /// What the list says when it has no hits: the failure, the two-letter floor, or no match.
    private func emptyLine(_ search: CatalogSearchModel) -> String {
        if search.showsErrorState, let error = search.searchError { return error }
        if search.isBelowMinimumQuery { return L10n.string(L10n.Promote.searchHint) }
        return L10n.Promote.searchNoMatches(row.kind, query: search.trimmedQuery)
    }
}

private struct PromoteAnchorRow: View {
    let row: PromoteFlow.Row

    var body: some View {
        PromoteGridRow {
            PVMark(row.kind.markKey, size: 16, decorative: true)
                .foregroundStyle(PVColor.textMuted)
        } source: {
            VStack(alignment: .leading, spacing: 0) {
                Text(verbatim: row.name)
                    .font(PVFont.body(size: PVTypeScale.bodySmall))
                    .foregroundStyle(PVColor.textSecondary)
                    .lineLimit(1)
                Text(verbatim: row.ref)
                    .font(PVFont.mono(size: PVTypeScale.caption))
                    .foregroundStyle(PVColor.textMuted)
            }
        } target: { _ in
            if case .handle(_, let ref, let title) = row.target {
                VStack(alignment: .leading, spacing: 0) {
                    Text(verbatim: title)
                        .font(PVFont.body(size: PVTypeScale.bodySmall))
                        .foregroundStyle(PVColor.textSecondary)
                        .lineLimit(1)
                    Text(verbatim: ref)
                        .font(PVFont.mono(size: PVTypeScale.caption))
                        .foregroundStyle(PVColor.textMuted)
                }
            } else {
                Color.clear.frame(height: 1)
            }
        } freed: {
            Text(L10n.Promote.filedEarlier)
                .font(PVFont.body(size: PVTypeScale.caption))
                .italic()
                .foregroundStyle(PVColor.textMuted)
        } badges: {
            Color.clear.frame(height: 1)
        }
        .padding(.vertical, PVSpacing.space2)
        .overlay(alignment: .top) { PVDivider() }
    }
}

/// The board's column grid (frame 27a), widths and all.
private enum PromoteLayout {
    static let columnGap = PVSpacing.space5
    static let kindTile: CGFloat = 28
    static let arrowColumn: CGFloat = 14
    static let badgeColumn: CGFloat = 112
    /// Notes under a row start where its name does.
    static let noteIndent = kindTile + columnGap
    static let listMaxRegular: CGFloat = 1200
    static let listMaxWide: CGFloat = 1600
    static let targetColumnRegular: CGFloat = 330
    static let targetColumnWide: CGFloat = 460

    static func listMax(_ widthClass: PVWidthClass) -> CGFloat {
        widthClass == .wide ? listMaxWide : listMaxRegular
    }

    static func targetColumn(_ widthClass: PVWidthClass) -> CGFloat {
        widthClass == .wide ? targetColumnWide : targetColumnRegular
    }
}

/// The one grid every Promote row spells: kind tile, Source, an arrow, the target, a slot that
/// stays free of copy on a working row, and the badges. The Source and the free slot share the
/// width the fixed columns leave. It reads the width class once and hands the target column's
/// width to its target slot.
private struct PromoteGridRow<Tile: View, Source: View, Target: View, Freed: View, Badges: View>: View {
    @Environment(\.pvWidthClass) private var widthClass
    let showsArrow: Bool
    let tile: Tile
    let source: Source
    let target: (CGFloat) -> Target
    let freed: Freed
    let badges: Badges

    init(
        showsArrow: Bool = true,
        @ViewBuilder tile: () -> Tile,
        @ViewBuilder source: () -> Source,
        @ViewBuilder target: @escaping (CGFloat) -> Target,
        @ViewBuilder freed: () -> Freed,
        @ViewBuilder badges: () -> Badges
    ) {
        self.showsArrow = showsArrow
        self.tile = tile()
        self.source = source()
        self.target = target
        self.freed = freed()
        self.badges = badges()
    }

    var body: some View {
        let targetWidth = PromoteLayout.targetColumn(widthClass)
        HStack(alignment: .center, spacing: PromoteLayout.columnGap) {
            tile.frame(width: PromoteLayout.kindTile, height: PromoteLayout.kindTile)
            source.frame(maxWidth: .infinity, alignment: .leading)
            Group {
                if showsArrow {
                    PVIcon(.arrowRight, size: PromoteLayout.arrowColumn)
                        .foregroundStyle(PVColor.textMuted)
                } else {
                    Color.clear
                }
            }
            .frame(width: PromoteLayout.arrowColumn)
            target(targetWidth).frame(width: targetWidth, alignment: .leading)
            freed.frame(maxWidth: .infinity, alignment: .leading)
            badges.frame(width: PromoteLayout.badgeColumn, alignment: .trailing)
        }
    }
}

/// "Updated · …": why a suggested row moved on the last proposal.
private struct PromoteUpdatedCaption: View {
    /// Why it moved, as `PromoteModel.reasonText` words it.
    let reason: String

    var body: some View {
        Text(verbatim: L10n.Promote.updatedDetail(reason))
            .font(PVFont.body(size: PVTypeScale.caption))
            .italic()
            .foregroundStyle(PVColor.infoForeground)
    }
}

private struct PromoteKindTile: View {
    let kind: EvidencePrimaryKind
    let size: CGFloat
    let markSize: CGFloat
    let background: Color

    var body: some View {
        let style = EvidenceSubjectKindStyle.forKind(kind)
        RoundedRectangle(cornerRadius: PVRadius.sm, style: .continuous)
            .fill(background)
            .overlay(RoundedRectangle(cornerRadius: PVRadius.sm, style: .continuous).strokeBorder(style.line, lineWidth: 1))
            .overlay(PVMark(kind.markKey, size: markSize, decorative: true).foregroundStyle(style.ink))
            .frame(width: size, height: size)
            .accessibilityHidden(true)
    }
}
