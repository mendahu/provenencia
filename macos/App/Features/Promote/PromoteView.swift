import SwiftUI

/// The Promote place (S9-44): one page, laid out as the S9-D16 board.
struct PromoteView: View {
    @Environment(WorkspaceNavigation.self) private var navigation
    @State private var model: PromoteModel

    init(
        entry: PromoteEntry,
        session: WorkspaceSession,
        store: any GenealogyStore,
        userID: String,
        catalogCounts: CatalogCounts?
    ) {
        _model = State(initialValue: PromoteModel(
            entry: entry,
            session: session,
            store: store,
            userID: userID,
            catalogCounts: catalogCounts
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
            if let error = model.saveError {
                PVCallout(tone: .danger, message: error)
            }
        }
    }

    private var header: some View {
        HStack(alignment: .center, spacing: PVSpacing.space5) {
            RoundedRectangle(cornerRadius: 4, style: .continuous)
                .fill(PVColor.surfaceSunken)
                .overlay(RoundedRectangle(cornerRadius: 4, style: .continuous).strokeBorder(PVColor.borderSubtle, lineWidth: 1))
                .overlay(PVMark(model.sourceMark, size: 24, decorative: true).foregroundStyle(PVColor.textSecondary))
                .frame(width: 44, height: 44)
            VStack(alignment: .leading, spacing: PVSpacing.space2) {
                Text(L10n.Promote.pageTitle)
                    .font(PVFont.body(size: 11, weight: PVFontWeight.semibold))
                    .tracking(1.2)
                    .textCase(.uppercase)
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
            Text(verbatim: L10n.Promote.openedFrom(name: model.entry.subjectName))
                .font(PVFont.body(size: PVTypeScale.bodySmall))
                .italic()
                .foregroundStyle(PVColor.textSecondary)
            Spacer(minLength: PVSpacing.space4)
            PVSelect(
                selection: groupBinding,
                options: [
                    PVSelectOption(value: "kind", label: L10n.string(L10n.Promote.groupKind)),
                    PVSelectOption(value: "assess", label: L10n.string(L10n.Promote.groupAssessment)),
                ],
                icon: .layers,
                displayLabel: model.flow.group == "assess"
                    ? L10n.string(L10n.Promote.groupedByAssessment)
                    : L10n.string(L10n.Promote.groupedByKind),
                fillsWidth: false,
                accessibilityLabel: L10n.Promote.groupLabel
            )
            .disabled(model.isSaving)
        }
    }

    private var mapRest: some View {
        HStack(alignment: .center, spacing: PVSpacing.space4) {
            PVButton(L10n.Promote.mapRest(count: model.flow.restCount), variant: .secondary, icon: .network, action: model.mapRest)
                .disabled(model.isSaving)
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
                        .font(PVFont.body(size: 11, weight: PVFontWeight.semibold))
                        .tracking(1.2)
                        .textCase(.uppercase)
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
        HStack(spacing: 14) {
            Color.clear.frame(width: 28, height: 1)
            columnLabel(L10n.Promote.columnSource)
            Color.clear.frame(width: 14, height: 1)
            columnLabel(L10n.Promote.columnTarget).frame(width: 330, alignment: .leading)
            columnLabel(L10n.Promote.columnAssessment)
            Color.clear.frame(width: 112, height: 1)
        }
        .padding(.top, 10)
        .padding(.bottom, PVSpacing.space3)
    }

    private func columnLabel(_ title: LocalizedStringResource) -> some View {
        Text(title)
            .font(PVFont.body(size: 11, weight: PVFontWeight.semibold))
            .tracking(1.2)
            .textCase(.uppercase)
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
        HStack(alignment: .center, spacing: 14) {
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
            if bridge.files || bridge.why == "off" {
                PVSwitch(
                    model.flow.bridgeOff.contains(bridge.id) ? L10n.Promote.connectionWhy("off") : L10n.Promote.connectionFiled,
                    isOn: bridgeOn(bridge.id),
                    isDisabled: model.isSaving
                )
            } else {
                Text(L10n.Promote.connectionWhy(bridge.why))
                    .font(PVFont.body(size: PVTypeScale.caption))
                    .italic()
                    .foregroundStyle(PVColor.textMuted)
                    .multilineTextAlignment(.trailing)
            }
        }
        .padding(.vertical, PVSpacing.space2)
        .overlay(alignment: .top) { PVDivider() }
    }

    private var connectionMeta: String {
        let lines = model.flow.connectionLines()
        return L10n.Promote.connectionsMeta(
            filed: lines.filter(\.files).count,
            off: lines.filter { $0.why == "off" }.count,
            unfiled: lines.filter { !$0.files && $0.why != "off" }.count
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
            PVButton(model.isSaving ? L10n.Promote.filing : L10n.Promote.done, variant: .primary, loading: model.isSaving) {
                model.done()
            }
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
        if model.flow.group == "assess" {
            return [
                section("strong", L10n.Promote.assessmentStrong, working.filter { $0.assessment == "strong" }, []),
                section("weak", L10n.Promote.assessmentWeak, working.filter { $0.assessment == "weak" }, []),
                section("none", L10n.Promote.assessmentNone, working.filter { $0.assessment == "none" }, []),
            ].filter { !$0.rows.isEmpty }
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
        Binding(get: { model.flow.group }, set: { model.setGroup($0) })
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
            get: { model.flow.rows.first { $0.subjectID == model.flow.sheetSubjectID } },
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
            HStack(alignment: .center, spacing: 14) {
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
                    menuWidth: 330,
                    fillsWidth: true,
                    rowHeight: 46,
                    isDisabled: model.isSaving,
                    accessibilitySpokenLabel: L10n.Promote.filesOn(name: row.name)
                ) { option in
                    PromoteTargetMenuRow(option: option, row: row)
                }
                .frame(width: 330)
                Button { model.openSheet(row.subjectID) } label: {
                    HStack(spacing: PVSpacing.space3) {
                        PVBadge(assessmentTitle(row.assessment), tone: badgeTone)
                        Text(verbatim: model.reasonText(for: row))
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
                HStack(spacing: PVSpacing.space2) {
                    if row.updatedNote != nil {
                        PVBadge(L10n.Promote.updated, tone: .info, icon: .refresh)
                    }
                    if row.decided {
                        PVBadge(L10n.Promote.decided, tone: .neutral, icon: .check, subtle: true)
                    }
                }
                .frame(width: 112, alignment: .trailing)
            }
            if let note = row.updatedNote {
                Text(verbatim: L10n.Promote.updatedDetail(note))
                    .font(PVFont.body(size: PVTypeScale.caption))
                    .italic()
                    .foregroundStyle(PVColor.textSecondary)
                    .padding(.leading, 42)
            }
            if let ref = row.conflictNote {
                PVCallout(
                    tone: .warning,
                    title: L10n.Promote.conflictTitle(ref: ref),
                    message: L10n.string(L10n.Promote.conflictBody),
                    compact: true
                )
                .padding(.leading, 42)
            }
            if let note = row.duplicateNote {
                let parts = note.split(separator: "|", maxSplits: 1).map(String.init)
                if parts.count == 2 {
                    PVCallout(
                        tone: .warning,
                        title: L10n.Promote.duplicateTitle(name: parts[0], ref: parts[1]),
                        message: L10n.string(L10n.Promote.duplicateBody),
                        compact: true
                    )
                    .padding(.leading, 42)
                }
            }
        }
        .padding(.vertical, PVSpacing.space3)
        .overlay(alignment: .top) { PVDivider() }
    }

    private var kindMark: some View {
        PromoteKindTile(kind: row.kind, size: 28, markSize: 16, background: PVColor.surfaceSunken)
    }

    private func assessmentTitle(_ assessment: String) -> LocalizedStringResource {
        switch assessment {
        case "strong": L10n.Promote.assessmentStrong
        case "weak": L10n.Promote.assessmentWeak
        default: L10n.Promote.assessmentNone
        }
    }

    private var badgeTone: PVBadgeTone {
        switch row.assessment {
        case "strong": .success
        case "weak": .warning
        default: .neutral
        }
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
        HStack(alignment: .center, spacing: 14) {
            PVMark(row.kind.markKey, size: 16, decorative: true)
                .foregroundStyle(PVColor.textMuted)
                .frame(width: 28, height: 28)
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
                .frame(width: 330, alignment: .leading)
            }
            Text(L10n.Promote.filedEarlier)
                .font(PVFont.body(size: PVTypeScale.caption))
                .italic()
                .foregroundStyle(PVColor.textMuted)
                .frame(maxWidth: .infinity, alignment: .leading)
            Color.clear.frame(width: 100, height: 1)
        }
        .padding(.vertical, PVSpacing.space2)
        .overlay(alignment: .top) { PVDivider() }
    }
}

private struct PromoteEvidenceSheet: View {
    let row: PromoteFlow.Row
    let grades: [CatalogClaimConfidenceGrade]
    let model: PromoteModel

    var body: some View {
        PVPanel(title: Text(verbatim: sheetTitle), subtitle: Text(verbatim: sheetSubtitle), width: 880) {
            ScrollView {
                VStack(alignment: .leading, spacing: PVSpacing.space5) {
                    if row.comparisons.contains(where: { $0.outcome == "conflict" }) {
                        PVCallout(
                            tone: .warning,
                            title: L10n.Promote.conflictCalloutTitle,
                            message: L10n.string(L10n.Promote.conflictCalloutBody),
                            compact: true
                        )
                    }
                    sheetColumns
                    ForEach(groups) { group in
                        PVCard {
                            VStack(alignment: .leading, spacing: 0) {
                                HStack(alignment: .firstTextBaseline, spacing: PVSpacing.space3) {
                                    Text(verbatim: group.title)
                                        .font(PVFont.display(size: PVTypeScale.h4, weight: PVFontWeight.medium))
                                        .foregroundStyle(PVColor.textDisplay)
                                    if !group.ref.isEmpty {
                                        Text(verbatim: group.ref)
                                            .font(PVFont.mono(size: PVTypeScale.caption))
                                            .foregroundStyle(PVColor.textMuted)
                                    }
                                }
                                .padding(.horizontal, PVSpacing.space4)
                                .padding(.vertical, PVSpacing.space3)
                                .frame(maxWidth: .infinity, alignment: .leading)
                                .background(PVColor.surfaceSunken)
                                ForEach(group.lines) { line in
                                    comparisonLine(line)
                                }
                            }
                        }
                    }
                    VStack(alignment: .leading, spacing: PVSpacing.space4) {
                        Text(L10n.Promote.claimHeading)
                            .font(PVFont.display(size: PVTypeScale.h4, weight: PVFontWeight.semibold))
                            .foregroundStyle(PVColor.textDisplay)
                        HStack(alignment: .top, spacing: PVSpacing.space5) {
                            PVField(label: L10n.Promote.status, hint: L10n.Promote.sheetStatusHint) {
                                PVSelect(
                                    selection: .constant("accepted"),
                                    options: [PVSelectOption(value: "accepted", label: L10n.string(L10n.Promote.statusAccepted))]
                                )
                            }
                            PVField(label: L10n.Promote.confidence) {
                                PVSelect(selection: confidenceBinding, options: confidenceOptions)
                            }
                        }
                        PVField(label: L10n.Promote.argument, hint: L10n.Promote.argumentHint(pinned: row.pins.count)) {
                            PVTextArea(text: argumentBinding)
                        }
                    }
                }
                .padding(PVSpacing.space5)
            }
        } footer: {
            PVButton(L10n.Promote.backToRows, variant: .primary) { model.closeSheet() }
        }
    }

    private var sheetTitle: String {
        switch row.target {
        case .handle(_, let ref, let title):
            return "\(row.name) → \(ref) \(title)"
        case .newKind:
            return "\(row.name) → \(L10n.string(L10n.Promote.newOption(row.kind)))"
        case .skip:
            return row.name
        }
    }

    private var sheetSubtitle: String {
        L10n.Promote.sheetSubtitle(assessment: L10n.string(assessmentTitle(row.assessment)), reason: model.reasonText(for: row))
    }

    private func assessmentTitle(_ assessment: String) -> LocalizedStringResource {
        switch assessment {
        case "strong": L10n.Promote.assessmentStrong
        case "weak": L10n.Promote.assessmentWeak
        default: L10n.Promote.assessmentNone
        }
    }

    private var sheetColumns: some View {
        HStack(spacing: 14) {
            Text(L10n.Promote.sheetPin).frame(width: 22, alignment: .leading)
            Text(L10n.Promote.sheetCompared).frame(width: 150, alignment: .leading)
            Text(L10n.Promote.sheetHere).frame(maxWidth: .infinity, alignment: .leading)
            Text(verbatim: targetColumn).frame(maxWidth: .infinity, alignment: .leading)
            Text(L10n.Promote.sheetResult).frame(width: 92, alignment: .leading)
            Text(L10n.Promote.sheetWeight).frame(width: 48, alignment: .trailing)
        }
        .font(PVFont.body(size: 11, weight: PVFontWeight.semibold))
        .tracking(1.2)
        .textCase(.uppercase)
        .foregroundStyle(PVColor.textMuted)
    }

    private var targetColumn: String {
        switch row.target {
        case .handle(_, let ref, _): ref
        case .newKind: L10n.string(L10n.Promote.newOption(row.kind))
        case .skip: L10n.string(L10n.Promote.skip)
        }
    }

    private func comparisonLine(_ line: CatalogPromoteGraphAlignmentComparison) -> some View {
        HStack(alignment: .top, spacing: 14) {
            PVCheckbox(
                isChecked: pin(line.id),
                isDisabled: line.incomingObservationID.isEmpty || line.outcome == "unknown",
                accessibilityLabel: L10n.Promote.pin
            )
            .frame(width: 22)
            Text(verbatim: compared(line))
                .font(PVFont.body(size: PVTypeScale.caption, weight: PVFontWeight.medium))
                .frame(width: 150, alignment: .leading)
            Text(verbatim: line.incomingDisplay)
                .font(PVFont.body(size: PVTypeScale.caption))
                .frame(maxWidth: .infinity, alignment: .leading)
            VStack(alignment: .leading, spacing: 2) {
                Text(verbatim: line.memberDisplay)
                    .font(PVFont.body(size: PVTypeScale.caption))
                if !line.memberSource.isEmpty {
                    Text(verbatim: line.memberSource)
                        .font(PVFont.body(size: PVTypeScale.caption))
                        .italic()
                        .foregroundStyle(PVColor.textMuted)
                }
            }
            .frame(maxWidth: .infinity, alignment: .leading)
            PVBadge(L10n.Promote.outcome(line.outcome), tone: outcomeTone(line.outcome), subtle: true)
                .frame(width: 92, alignment: .leading)
            Text(verbatim: weightText(line.weight))
                .font(PVFont.mono(size: PVTypeScale.caption))
                .foregroundStyle(PVColor.textSecondary)
                .frame(width: 48, alignment: .trailing)
        }
        .padding(.horizontal, PVSpacing.space4)
        .padding(.vertical, PVSpacing.space3)
        .overlay(alignment: .top) { PVDivider() }
    }

    private struct ExhibitGroup: Identifiable {
        var id: String
        var title: String
        var ref: String
        var lines: [CatalogPromoteGraphAlignmentComparison]
    }

    private var groups: [ExhibitGroup] {
        var order: [String] = []
        var lines: [String: [CatalogPromoteGraphAlignmentComparison]] = [:]
        for line in row.comparisons {
            let key = neighbor(line.groupLabel)
            if lines[key] == nil { order.append(key) }
            lines[key, default: []].append(line)
        }
        return order.map { key in
            let title = key.isEmpty ? L10n.string(L10n.Promote.ownRecords) : L10n.Promote.through(key)
            let ref = key.isEmpty ? row.ref : ""
            return ExhibitGroup(id: key.isEmpty ? "own" : key, title: title, ref: ref, lines: lines[key] ?? [])
        }
    }

    private func neighbor(_ label: String) -> String {
        guard let split = label.range(of: " · ", options: .backwards) else { return label }
        return String(label[..<split.lowerBound])
    }

    private func compared(_ line: CatalogPromoteGraphAlignmentComparison) -> String {
        guard let split = line.groupLabel.range(of: " · ", options: .backwards) else {
            return line.propertyKey.replacingOccurrences(of: "_", with: " ")
        }
        return String(line.groupLabel[split.upperBound...])
    }

    private func outcomeTone(_ outcome: String) -> PVBadgeTone {
        switch outcome {
        case "agree": .success
        case "conflict": .danger
        default: .neutral
        }
    }

    private func weightText(_ weight: Double) -> String {
        if weight == 0 { return "0" }
        let sign = weight > 0 ? "+" : "−"
        return sign + String(format: "%.2f", abs(weight))
    }

    private func pin(_ id: String) -> Binding<Bool> {
        Binding(
            get: { row.pins.contains(id) },
            set: { _ in model.togglePin(subjectID: row.subjectID, comparisonID: id) }
        )
    }

    private var argumentBinding: Binding<String> {
        Binding(get: { row.argument }, set: { model.setArgument(subjectID: row.subjectID, argument: $0) })
    }

    private var confidenceBinding: Binding<String> {
        Binding(
            get: { row.confidenceGradeID ?? "" },
            set: { model.setConfidence(subjectID: row.subjectID, gradeID: $0.isEmpty ? nil : $0) }
        )
    }

    private var confidenceOptions: [PVSelectOption] {
        [PVSelectOption(value: "", label: L10n.string(L10n.Promote.confidenceUnset))]
            + grades.map { PVSelectOption(value: $0.id, label: $0.label) }
    }
}

private struct PromoteKindTile: View {
    let kind: EvidencePrimaryKind
    let size: CGFloat
    let markSize: CGFloat
    let background: Color

    var body: some View {
        let style = EvidenceSubjectKindStyle.forKind(kind)
        RoundedRectangle(cornerRadius: 4, style: .continuous)
            .fill(background)
            .overlay(RoundedRectangle(cornerRadius: 4, style: .continuous).strokeBorder(style.line, lineWidth: 1))
            .overlay(PVMark(kind.markKey, size: markSize, decorative: true).foregroundStyle(style.ink))
            .frame(width: size, height: size)
            .accessibilityHidden(true)
    }
}
