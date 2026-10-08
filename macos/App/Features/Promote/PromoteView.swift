import SwiftUI

/// The Promote place (S9-44): one page for a Source's graph alignment.
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
                VStack(alignment: .leading, spacing: PVSpacing.space6) {
                    if let error = model.loadError {
                        PVCallout(tone: .danger, message: error)
                    }
                    if let stale = model.flow.staleNote {
                        PVCallout(tone: .warning, message: stale)
                    }
                    if let error = model.saveError {
                        PVCallout(tone: .danger, message: error)
                    }
                    if !model.flow.mappedRest && model.flow.restCount > 0 {
                        PVButton(L10n.Promote.mapRest(count: model.flow.restCount), variant: .secondary, action: model.mapRest)
                            .disabled(model.isSaving)
                    }
                    ForEach(sections, id: \.id) { section in
                        PVSectionHeader(title: section.title)
                        ForEach(section.rows) { row in
                            PromoteAlignmentRow(row: row, model: model)
                        }
                    }
                    connections
                }
                .padding(.horizontal, PVSpacing.space10)
                .padding(.vertical, PVSpacing.space8)
                .frame(maxWidth: 960, alignment: .leading)
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
            let _: QueryHandle<SourceGraphRows> = model.session.query(model.graphKey)
            let _: QueryHandle<[CatalogClaimConfidenceGrade]> = model.session.query(model.confidenceKey)
            let _: QueryHandle<PropertiesSnapshot> = model.session.query(model.propertiesKey)
            let _: QueryHandle<[CatalogConnectRule]> = model.session.query(model.rulesKey)
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
                    message: L10n.string(L10n.Promote.leaveMessage),
                    confirm: L10n.Promote.leaveConfirm,
                    cancel: L10n.Promote.leaveCancel
                )
            },
            tone: .irreversible,
            accessibilityIdentifierPrefix: "promote.leave",
            onConfirm: { model.leave() },
            detail: { _ in EmptyView() }
        )
        .disabled(model.isSaving)
    }

    private var header: some View {
        VStack(alignment: .leading, spacing: PVSpacing.space2) {
            Text(L10n.Promote.pageTitle)
                .font(PVFont.display(size: PVTypeScale.h2, weight: PVFontWeight.semibold))
            if let title = model.entry.sourceTitle {
                Text(verbatim: title)
                    .font(PVFont.body(size: PVTypeScale.bodySmall))
                    .foregroundStyle(PVColor.textSecondary)
            }
            Text(verbatim: L10n.Promote.rowCount(count: model.flow.visibleRows.count))
                .font(PVFont.mono(size: PVTypeScale.caption))
                .foregroundStyle(PVColor.textMuted)
            PVSelect(
                selection: groupBinding,
                options: [
                    PVSelectOption(value: "kind", label: L10n.string(L10n.Promote.groupKind)),
                    PVSelectOption(value: "assess", label: L10n.string(L10n.Promote.groupAssessment)),
                ],
                fillsWidth: false,
                accessibilityLabel: L10n.Promote.groupLabel
            )
        }
        .frame(maxWidth: .infinity, alignment: .leading)
        .padding(.horizontal, PVSpacing.space10)
        .padding(.vertical, PVSpacing.space6)
    }

    private var footer: some View {
        HStack {
            Text(verbatim: L10n.Promote.doneSummary(
                file: model.flow.fileCount,
                skip: model.flow.skipCount,
                connections: model.flow.connectionCount
            ))
            .font(PVFont.body(size: PVTypeScale.bodySmall))
            .foregroundStyle(PVColor.textSecondary)
            Spacer()
            PVButton(L10n.Promote.cancel, variant: .secondary) { model.leave() }
                .disabled(model.isSaving)
            PVButton(model.isSaving ? L10n.Promote.filing : L10n.Promote.done, variant: .primary) { model.done() }
                .disabled(model.isSaving)
        }
        .padding(.horizontal, PVSpacing.space10)
        .padding(.vertical, PVSpacing.space5)
        .background(PVColor.surfaceCard)
    }

    private var connections: some View {
        VStack(alignment: .leading, spacing: PVSpacing.space3) {
            PVDisclosureButton(L10n.Promote.connections(count: model.flow.connectionCount), isExpanded: bridgesOpen)
            if bridgesOpen.wrappedValue {
                ForEach(model.flow.connectionLines()) { bridge in
                    HStack(alignment: .firstTextBaseline) {
                        Text(verbatim: bridge.sentence)
                            .font(PVFont.body(size: PVTypeScale.bodySmall))
                        Spacer()
                        if bridge.files {
                            Toggle(L10n.string(L10n.Promote.connectionFiled), isOn: bridgeOn(bridge.id))
                                .toggleStyle(.switch)
                                .labelsHidden()
                        } else {
                            Text(L10n.Promote.connectionWhy(bridge.why))
                                .font(PVFont.body(size: PVTypeScale.caption))
                                .foregroundStyle(PVColor.textMuted)
                        }
                    }
                }
            }
        }
    }

    private var sections: [(id: String, title: LocalizedStringResource, rows: [PromoteFlow.Row])] {
        let rows = model.flow.visibleRows
        if model.flow.group == "assess" {
            return [
                ("strong", L10n.Promote.assessmentStrong, rows.filter { $0.assessment == "strong" }),
                ("weak", L10n.Promote.assessmentWeak, rows.filter { $0.assessment == "weak" }),
                ("none", L10n.Promote.assessmentNone, rows.filter { $0.assessment == "none" }),
            ].filter { !$0.rows.isEmpty }
        }
        return [
            ("person", L10n.Promote.kindPerson, rows.filter { $0.kind == .person }),
            ("event", L10n.Promote.kindEvent, rows.filter { $0.kind == .event }),
            ("place", L10n.Promote.kindPlace, rows.filter { $0.kind == .place }),
        ].filter { !$0.rows.isEmpty }
    }

    private var groupBinding: Binding<String> {
        Binding(get: { model.flow.group }, set: { model.setGroup($0) })
    }

    private var bridgesOpen: Binding<Bool> {
        Binding(
            get: { model.flow.bridgesOpen },
            set: { model.setBridgesOpen($0) }
        )
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
        PVCard {
            VStack(alignment: .leading, spacing: PVSpacing.space3) {
                HStack(alignment: .center, spacing: PVSpacing.space4) {
                    PromoteKindTile(kind: row.kind, size: 28, markSize: 14, background: EvidenceSubjectKindStyle.forKind(row.kind).tint)
                    VStack(alignment: .leading, spacing: 0) {
                        Text(verbatim: row.name)
                            .font(PVFont.display(size: PVTypeScale.h4, weight: PVFontWeight.medium))
                        Text(verbatim: row.ref)
                            .font(PVFont.mono(size: PVTypeScale.caption))
                            .foregroundStyle(PVColor.textMuted)
                    }
                    Spacer()
                    if row.anchor, case .handle(_, let ref, let title) = row.target {
                        Text(verbatim: "\(ref) \(title)")
                            .font(PVFont.mono(size: PVTypeScale.caption))
                        PVBadge(L10n.Promote.anchorBadge, tone: .neutral, subtle: true)
                    } else if !row.anchor {
                        PVSelect(
                            selection: targetBinding,
                            options: targetOptions,
                            fillsWidth: false,
                            accessibilitySpokenLabel: L10n.Promote.targetLabel(name: row.name)
                        )
                    }
                    Button { model.openSheet(row.subjectID) } label: {
                        PVBadge(assessmentTitle(row.assessment), tone: badgeTone)
                    }
                    .buttonStyle(.plain)
                    .disabled(row.anchor)
                    if row.decided {
                        PVBadge(L10n.Promote.decided, tone: .info, subtle: true)
                    }
                    if row.updatedNote != nil {
                        PVBadge(L10n.Promote.updated, tone: .accent, subtle: true)
                    }
                }
                Text(verbatim: model.reasonText(for: row))
                    .font(PVFont.body(size: PVTypeScale.caption))
                    .foregroundStyle(PVColor.textSecondary)
                if let note = row.updatedNote {
                    Text(verbatim: L10n.Promote.updatedDetail(note))
                        .font(PVFont.body(size: PVTypeScale.caption))
                        .foregroundStyle(PVColor.textSecondary)
                }
                if let ref = row.conflictNote {
                    PVCallout(tone: .warning, message: L10n.Promote.conflict(ref: ref))
                }
                if let note = row.duplicateNote {
                    let parts = note.split(separator: "|", maxSplits: 1).map(String.init)
                    if parts.count == 2 {
                        PVCallout(tone: .warning, message: L10n.Promote.duplicate(name: parts[0], ref: parts[1]))
                    }
                }
            }
            .padding(PVSpacing.space4)
        }
        .opacity(row.anchor ? 0.72 : 1)
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
            PVSelectOption(value: "handle:\(alt.id)", label: "\(alt.ref) \(alt.title)")
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

private struct PromoteEvidenceSheet: View {
    let row: PromoteFlow.Row
    let grades: [CatalogClaimConfidenceGrade]
    let model: PromoteModel

    var body: some View {
        PVPanel(title: Text(L10n.Promote.evidenceTitle), width: 640) {
            ScrollView {
                VStack(alignment: .leading, spacing: PVSpacing.space5) {
                    ForEach(groups, id: \.label) { group in
                        Text(verbatim: group.label)
                            .font(PVFont.display(size: PVTypeScale.h4, weight: PVFontWeight.semibold))
                        ForEach(group.lines) { line in
                            HStack(alignment: .top) {
                                VStack(alignment: .leading, spacing: PVSpacing.space1) {
                                    Text(verbatim: line.propertyKey)
                                        .font(PVFont.body(size: PVTypeScale.caption))
                                        .foregroundStyle(PVColor.textMuted)
                                    Text(verbatim: line.incomingDisplay)
                                    Text(verbatim: line.memberDisplay)
                                        .foregroundStyle(PVColor.textSecondary)
                                    Text(L10n.Promote.outcome(line.outcome))
                                        .font(PVFont.body(size: PVTypeScale.caption))
                                }
                                Spacer()
                                Toggle(L10n.string(L10n.Promote.pin), isOn: pin(line.id))
                                    .toggleStyle(.switch)
                                    .disabled(line.incomingObservationID.isEmpty)
                            }
                        }
                    }
                    PVField(label: L10n.Promote.status) {
                        PVSelect(
                            selection: .constant("accepted"),
                            options: [PVSelectOption(value: "accepted", label: L10n.string(L10n.Promote.statusAccepted))],
                            isDisabled: true
                        )
                    }
                    PVField(label: L10n.Promote.confidence) {
                        PVSelect(selection: confidenceBinding, options: confidenceOptions)
                    }
                    PVField(label: L10n.Promote.argument) {
                        PVTextArea(text: argumentBinding)
                    }
                }
                .padding(PVSpacing.space5)
            }
        } footer: {
            PVButton(L10n.Promote.backToRows, variant: .primary) { model.closeSheet() }
        }
    }

    private var groups: [(label: String, lines: [CatalogPromoteGraphAlignmentComparison])] {
        let labels = Array(Set(row.comparisons.map(\.groupLabel))).sorted()
        return labels.map { label in
            (label.isEmpty ? L10n.string(L10n.Promote.ownRecords) : label, row.comparisons.filter { $0.groupLabel == label })
        }
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
