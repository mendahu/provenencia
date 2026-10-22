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
        VStack(spacing: 0) {
            header
            ScrollView {
                VStack(alignment: .leading, spacing: PVSpacing.space7) {
                    notices
                    if model.flow.mappedRest {
                        mappedBar
                    }
                    ForEach(sections, id: \.id) { section in
                        sectionBlock(section)
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
            if bridgesOpen.wrappedValue {
                connectionsPanel
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
            let _: QueryHandle<SourceGraphRows> = model.session.query(model.graphKey)
            let _: QueryHandle<[CatalogClaimConfidenceGrade]> = model.session.query(model.confidenceKey)
            let _: QueryHandle<PropertiesSnapshot> = model.session.query(model.propertiesKey)
            let _: QueryHandle<[CatalogConnectRule]> = model.session.query(model.rulesKey)
            let _: QueryHandle<[CatalogSource]> = model.session.query(model.sourcesKey)
            let _: QueryHandle<[CatalogSourceType]> = model.session.query(model.sourceTypesKey)
            await model.load()
        }
        .sheet(item: sheetBinding) { row in
            PromoteEvidenceSheet(row: row, grades: model.grades, model: model)
        }
        .pvConfirm(
            item: leaveBinding,
            copy: { _ in
                PVConfirmCopy(
                    title: L10n.string(L10n.Promote.leaveTitle),
                    message: L10n.Promote.leaveDetail(
                        rows: model.flow.rows.filter(\.decided).count,
                        connections: model.flow.bridgeOff.count
                    ),
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

    private var mappedBar: some View {
        HStack(alignment: .center) {
            if !model.entry.mapAll {
                Text(verbatim: L10n.Promote.openedFrom(name: model.entry.subjectName))
                    .font(PVFont.body(size: PVTypeScale.bodySmall))
                    .italic()
                    .foregroundStyle(PVColor.textSecondary)
            }
            Spacer(minLength: PVSpacing.space4)
            PVSelect(
                selection: groupBinding,
                options: [
                    PVSelectOption(value: PromoteFlow.Grouping.kind.rawValue, label: L10n.string(L10n.Promote.groupKind)),
                    PVSelectOption(value: PromoteFlow.Grouping.assessment.rawValue, label: L10n.string(L10n.Promote.groupAssessment)),
                ],
                icon: .layers,
                displayLabel: model.flow.group == .assessment
                    ? L10n.string(L10n.Promote.groupedByAssessment)
                    : L10n.string(L10n.Promote.groupedByKind),
                fillsWidth: false,
                accessibilityLabel: L10n.Promote.groupLabel,
                accessibilityIdentifier: "promote.group"
            )
            .disabled(model.isSaving)
        }
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

    private func sectionBlock(_ section: Section) -> some View {
        VStack(alignment: .leading, spacing: 0) {
            PVSectionHeader(title: section.title, meta: section.meta)
            if !section.rows.isEmpty {
                columnHeaders
                ForEach(section.rows) { row in
                    PromoteAlignmentRow(row: row, model: model)
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

    private var connectionsPanel: some View {
        VStack(alignment: .leading, spacing: PVSpacing.space3) {
            HStack(alignment: .firstTextBaseline, spacing: PVSpacing.space4) {
                Text(L10n.Promote.connectionsTitle)
                    .font(PVFont.display(size: PVTypeScale.h4, weight: PVFontWeight.semibold))
                    .foregroundStyle(PVColor.textDisplay)
                Text(verbatim: connectionMeta)
                    .font(PVFont.mono(size: PVTypeScale.caption))
                    .foregroundStyle(PVColor.textMuted)
                Text(L10n.Promote.connectionsHint)
                    .font(PVFont.body(size: PVTypeScale.bodySmall))
                    .italic()
                    .foregroundStyle(PVColor.textSecondary)
            }
            ScrollView {
                VStack(spacing: 0) {
                    ForEach(model.flow.connectionLines()) { bridge in
                        connectionRow(bridge)
                    }
                }
            }
            .frame(maxHeight: 420)
        }
        .padding(.horizontal, PVSpacing.space10)
        .padding(.vertical, PVSpacing.space4)
        .background(PVColor.surfaceCard)
        .overlay(alignment: .top) { PVDivider() }
    }

    private func connectionRow(_ bridge: PromoteFlow.Bridge) -> some View {
        HStack(alignment: .center, spacing: PromoteLayout.columnGap) {
            PVMark(bridge.mark, size: 16, decorative: true)
                .foregroundStyle(PVColor.textMuted)
                .frame(width: 20)
            if bridge.endAName.isEmpty || bridge.endBName.isEmpty {
                Text(verbatim: bridge.sentence)
                    .font(PVFont.body(size: PVTypeScale.bodySmall))
                    .frame(maxWidth: .infinity, alignment: .leading)
            } else {
                HStack(spacing: PVSpacing.space2) {
                    Text(verbatim: bridge.endAName)
                    Text(verbatim: bridge.phrase)
                        .italic()
                        .foregroundStyle(PVColor.textSecondary)
                    Text(verbatim: bridge.endBName)
                }
                .font(PVFont.body(size: PVTypeScale.bodySmall))
                .frame(maxWidth: .infinity, alignment: .leading)
            }
            if bridge.state == .files || bridge.state == .off {
                PVSwitch(
                    L10n.Promote.connectionState(bridge.state),
                    isOn: bridgeOn(bridge.id),
                    isDisabled: model.isSaving
                )
                .accessibilityIdentifier("promote.connection.\(bridge.id)")
            } else {
                Text(L10n.Promote.connectionState(bridge.state))
                    .font(PVFont.body(size: PVTypeScale.caption))
                    .italic()
                    .foregroundStyle(PVColor.textMuted)
                    .multilineTextAlignment(.trailing)
            }
        }
        .padding(.vertical, PVSpacing.space2)
        .overlay(alignment: .top) { PVDivider() }
        // One element: the connection's sentence, then its switch or why it won't file.
        .accessibilityElement(children: .combine)
    }

    private var connectionMeta: String {
        let lines = model.flow.connectionLines()
        return L10n.Promote.connectionsMeta(
            filed: lines.filter(\.files).count,
            off: lines.filter { $0.state == .off }.count,
            unfiled: lines.filter { $0.state == .endSkipped || $0.state == .selfLink }.count
        )
    }

    private var footer: some View {
        HStack(spacing: PVSpacing.space4) {
            PVButton(
                L10n.Promote.connections(count: model.flow.connectionCount),
                variant: .ghost,
                size: .sm,
                iconRight: bridgesOpen.wrappedValue ? .chevronUp : .chevronDown
            ) {
                model.setBridgesOpen(!model.flow.bridgesOpen)
            }
            .disabled(model.isSaving)
            .accessibilityIdentifier("promote.connections.toggle")
            Spacer()
            Text(verbatim: L10n.Promote.doneSummary(
                file: model.flow.fileCount,
                skip: model.flow.skipCount,
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
        if model.flow.group == .assessment {
            return [PromoteFlow.Assessment.strong, .medium, .weak, .none].map { band in
                section(band.rawValue, L10n.Promote.assessment(band), working.filter { $0.assessment == band }, [])
            }.filter { !$0.rows.isEmpty }
        }
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
        return section(id, title, rows, filed)
    }

    private func section(
        _ id: String,
        _ title: LocalizedStringResource,
        _ rows: [PromoteFlow.Row],
        _ anchors: [PromoteFlow.Row]
    ) -> Section {
        let total = rows.count + anchors.count
        let meta = anchors.isEmpty
            ? L10n.Promote.sectionCount(total)
            : L10n.Promote.sectionCountFiled(rows: total, filed: anchors.count)
        return Section(id: id, title: title, meta: meta, rows: rows, anchors: anchors)
    }

    private var groupBinding: Binding<String> {
        Binding(
            get: { model.flow.group.rawValue },
            set: { model.setGroup(PromoteFlow.Grouping(rawValue: $0) ?? .kind) }
        )
    }

    private var bridgesOpen: Binding<Bool> {
        Binding(get: { model.flow.bridgesOpen }, set: { model.setBridgesOpen($0) })
    }

    private func bridgeOn(_ id: String) -> Binding<Bool> {
        Binding(
            get: { !model.flow.bridgeOff.contains(id) },
            set: { _ in model.toggleBridge(id) }
        )
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
                .frame(width: PromoteLayout.targetColumn)
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
                Text(verbatim: L10n.Promote.updatedDetail(model.reasonText(row.captionReason, on: row)))
                    .font(PVFont.body(size: PVTypeScale.caption))
                    .italic()
                    .foregroundStyle(PVColor.infoForeground)
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

    private var kindMark: some View {
        PromoteKindTile(kind: row.kind, size: PromoteLayout.kindTile, markSize: 16, background: PVColor.surfaceSunken)
    }

    private var targetOptions: [PVSelectOption] {
        var options = row.menu.map { alt in
            PVSelectOption(value: "handle:\(alt.id)", label: "\(alt.ref)  \(alt.title)")
        }
        options.append(PVSelectOption(value: "new", label: L10n.string(L10n.Promote.newOption(row.kind))))
        options.append(PVSelectOption(value: "skip", label: L10n.string(L10n.Promote.skip)))
        return options
    }

    private var targetBinding: Binding<String> {
        Binding(
            get: { row.target.token },
            set: { model.setTarget(subjectID: row.subjectID, token: $0) }
        )
    }
}

private struct PromoteTargetMenuRow: View {
    let option: PVSelectOption
    let row: PromoteFlow.Row

    var body: some View {
        if option.id == "new" || option.id == "skip" {
            Text(verbatim: option.label)
                .font(PVFont.body(size: PVTypeScale.bodySmall))
                .foregroundStyle(PVColor.textPrimary)
        } else if let alt = row.menu.first(where: { "handle:\($0.id)" == option.id }) {
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
