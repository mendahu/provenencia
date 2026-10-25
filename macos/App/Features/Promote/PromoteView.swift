import SwiftUI

/// The Promote place (S9-44): one page, laid out as the S9-D16 board.
struct PromoteView: View {
    @Environment(WorkspaceNavigation.self) private var navigation
    @State private var model: PromoteModel

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
                .frame(maxWidth: 1200, alignment: .leading)
                .frame(maxWidth: .infinity, alignment: .leading)
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
            let _: QueryHandle<[CatalogClaimConfidenceGrade]> = model.session.query(model.confidenceKey)
            let _: QueryHandle<PropertiesSnapshot> = model.session.query(model.propertiesKey)
            let _: QueryHandle<[CatalogConnectRule]> = model.session.query(model.rulesKey)
            let _: QueryHandle<[CatalogSource]> = model.session.query(model.sourcesKey)
            let _: QueryHandle<[CatalogSourceType]> = model.session.query(model.sourceTypesKey)
            await model.load()
        }
        .pvSheet(item: sheetBinding) { row in
            PromoteEvidenceSheet(row: row, grades: model.grades, model: model)
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
        HStack(spacing: PromoteLayout.columnGap) {
            Color.clear.frame(width: PromoteLayout.kindTile, height: 1)
            columnLabel(L10n.Promote.columnSource)
            Color.clear.frame(width: 14, height: 1)
            columnLabel(L10n.Promote.columnTarget).frame(width: PromoteLayout.targetColumn, alignment: .leading)
            columnLabel(L10n.Promote.columnAssessment)
            Color.clear.frame(width: PromoteLayout.badgeColumn, height: 1)
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

    private var sheetBinding: Binding<PromoteFlow.Row?> {
        Binding(
            get: { model.flow.row(model.flow.sheetSubjectID) },
            set: { model.setSheetSubject($0?.subjectID) }
        )
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

    var body: some View {
        VStack(alignment: .leading, spacing: PVSpacing.space2) {
            HStack(alignment: .center, spacing: PromoteLayout.columnGap) {
                kindMark
                VStack(alignment: .leading, spacing: 0) {
                    Text(verbatim: row.name)
                        .font(PVFont.body(size: PVTypeScale.bodySmall, weight: PVFontWeight.medium))
                        .lineLimit(1)
                    Text(verbatim: row.ref)
                        .font(PVFont.mono(size: PVTypeScale.caption))
                        .foregroundStyle(PVColor.textMuted)
                }
                .frame(maxWidth: .infinity, alignment: .leading)
                PVIcon(.arrowRight, size: 14)
                    .foregroundStyle(PVColor.textMuted)
                    .frame(width: 14)
                if model.recordSearch?.subjectID == row.subjectID {
                    PromoteRecordSearch(row: row, model: model)
                        .frame(width: PromoteLayout.targetColumn)
                } else {
                    targetSelect
                        .frame(width: PromoteLayout.targetColumn)
                }
                // A real stack, not a Group: Group forwards its frame to each
                // child, so an empty assessment column collapses and the
                // target menu slides into the gap.
                HStack(spacing: PVSpacing.space3) {
                    if let readout = row.selectionReadout {
                        Button { model.openSheet(row.subjectID) } label: {
                            HStack(spacing: PVSpacing.space3) {
                                PromoteReadoutBadges(readout: readout)
                                Text(verbatim: model.reasonText(readout.reason, on: row))
                                    .font(PVFont.body(size: PVTypeScale.caption))
                                    .italic()
                                    .foregroundStyle(PVColor.textSecondary)
                                    .lineLimit(1)
                                Spacer(minLength: 0)
                                PVIcon(.chevronForward, size: 14)
                                    .foregroundStyle(PVColor.textMuted)
                            }
                        }
                        .buttonStyle(.plain)
                        .frame(maxWidth: .infinity, alignment: .leading)
                        .accessibilityLabel(Text(verbatim: L10n.Promote.evidenceFor(
                            name: row.name,
                            assessment: model.assessmentText(readout),
                            reason: model.reasonText(readout.reason, on: row)
                        )))
                        .accessibilityIdentifier("promote.row.\(row.ref).evidence")
                    }
                }
                .frame(maxWidth: .infinity, alignment: .leading)
                HStack(spacing: PVSpacing.space2) {
                    if row.updated {
                        PVBadge(L10n.Promote.updated, tone: .info, icon: .refresh)
                    }
                    if row.decided {
                        PVBadge(L10n.Promote.decided, tone: .neutral, icon: .check, subtle: true)
                    }
                }
                .frame(width: PromoteLayout.badgeColumn, alignment: .trailing)
            }
            if row.updated {
                PromoteUpdatedCaption(reason: model.reasonText(row.captionReason, on: row))
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
    }

    private func duplicateTitle(_ note: PromoteFlow.DuplicateNote) -> String {
        switch note {
        case .sharesHandle(let other, let ref): L10n.Promote.duplicateTitle(name: other, ref: ref)
        case .alsoNew(let other): L10n.Promote.duplicateNewTitle(name: other)
        }
    }

    private var targetSelect: some View {
        PVSelect(
            selection: targetBinding,
            options: targetOptions,
            size: .sm,
            placeholder: row.target == .unset ? L10n.string(L10n.Promote.targetPlaceholder) : nil,
            menuWidth: PromoteLayout.targetColumn,
            fillsWidth: true,
            rowHeight: 46,
            isDisabled: model.isSaving,
            isInvalid: row.target == .unset,
            accessibilitySpokenLabel: L10n.Promote.filesOn(name: row.name),
            accessibilityIdentifier: "promote.row.\(row.ref).target"
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
            accessibilityIdentifier: "promote.row.\(row.ref).target.find"
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
        HStack(alignment: .center, spacing: PromoteLayout.columnGap) {
            PVMark(row.kind.markKey, size: 16, decorative: true)
                .foregroundStyle(PVColor.textMuted)
                .frame(width: PromoteLayout.kindTile, height: PromoteLayout.kindTile)
            VStack(alignment: .leading, spacing: 0) {
                Text(verbatim: row.name)
                    .font(PVFont.body(size: PVTypeScale.bodySmall))
                    .foregroundStyle(PVColor.textSecondary)
                    .lineLimit(1)
                Text(verbatim: row.ref)
                    .font(PVFont.mono(size: PVTypeScale.caption))
                    .foregroundStyle(PVColor.textMuted)
            }
            .frame(maxWidth: .infinity, alignment: .leading)
            PVIcon(.arrowRight, size: 14)
                .foregroundStyle(PVColor.textMuted)
                .frame(width: 14)
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
                .frame(width: PromoteLayout.targetColumn, alignment: .leading)
            }
            Text(L10n.Promote.filedEarlier)
                .font(PVFont.body(size: PVTypeScale.caption))
                .italic()
                .foregroundStyle(PVColor.textMuted)
                .frame(maxWidth: .infinity, alignment: .leading)
            Color.clear.frame(width: PromoteLayout.badgeColumn, height: 1)
        }
        .padding(.vertical, PVSpacing.space2)
        .overlay(alignment: .top) { PVDivider() }
    }
}

/// The board's column grid, shared by the header row and every row so the
/// columns line up.
private enum PromoteLayout {
    static let columnGap = PVSpacing.space5
    static let kindTile: CGFloat = 28
    static let targetColumn: CGFloat = 330
    static let badgeColumn: CGFloat = 112
    /// Notes under a row start where its name does.
    static let noteIndent = kindTile + columnGap
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
