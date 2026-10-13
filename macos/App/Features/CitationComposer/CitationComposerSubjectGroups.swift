import SwiftUI

/// Observations column: header, authoring box, subject cards.
struct CitationComposerObservationsColumn: View {
    @Bindable var model: CitationComposerModel
    var onDone: () -> Void
    @State private var lineMenu = PVContextMenuState()
    @State private var menuLine: CitationSavedLine?
    @State private var hoveredLineID: String?

    var body: some View {
        VStack(alignment: .leading, spacing: 0) {
            header
                .padding(.horizontal, PVSpacing.space6)
                .padding(.top, PVSpacing.space6)
                .padding(.bottom, PVSpacing.space4)
            CitationComposerAuthoringBox(model: model)
                .padding(.horizontal, PVSpacing.space6)
                .padding(.bottom, PVSpacing.space4)
            ScrollView {
                groups
                    .padding(.horizontal, PVSpacing.space6)
                    .padding(.bottom, PVSpacing.space6)
            }
            .frame(maxWidth: .infinity, maxHeight: .infinity, alignment: .topLeading)
        }
        .background(PVColor.surfaceCard)
        .overlay(alignment: .bottom) {
            if let toast = model.bridgeToast {
                PVToast(tone: .success, title: toast.title, message: toast.message) {
                    model.bridgeToast = nil
                }
                .padding(PVSpacing.space5)
            }
        }
        .pvContextMenu(lineMenuBinding) {
            lineMenuContent
        }
        .accessibilityElement(children: .contain)
        .accessibilityLabel(Text(L10n.CitationComposer.observationsSection))
    }

    private var header: some View {
        PVSectionHeader(
            title: L10n.CitationComposer.observationsSection,
            meta: model.observationsMeta,
            aside: {
                if model.authoring.holdsLeave {
                    Text(L10n.CitationComposer.unsavedDraft)
                        .font(PVFont.body(size: PVTypeScale.caption, italic: true))
                        .foregroundStyle(PVColor.textMuted)
                }
            },
            actions: {
                PVButton(L10n.CitationComposer.done, variant: .primary, size: .sm) {
                    onDone()
                }
                .accessibilityIdentifier("citationComposer.done")
            }
        )
    }

    @ViewBuilder
    private var groups: some View {
        if model.subjectGroups.isEmpty {
            PVEmptyState(
                icon: .bookmark,
                title: L10n.CitationComposer.noObservations,
                message: L10n.string(L10n.CitationComposer.emptyObservationsMessage)
            ) {
                EmptyView()
            }
            .accessibilityIdentifier("citationComposer.observations.empty")
        } else {
            VStack(alignment: .leading, spacing: PVSpacing.space5) {
                ForEach(model.subjectGroups) { group in
                    subjectCard(group)
                }
            }
        }
    }

    private func subjectCard(_ group: CitationSubjectGroup) -> some View {
        PVCard(padding: PVSpacing.space4) {
            VStack(alignment: .leading, spacing: PVSpacing.space3) {
                HStack(spacing: PVSpacing.space3) {
                    PVMark(PVMarkKey(catalogKey: "subject_\(group.typeKey)"), size: 16, decorative: true)
                    Text(verbatim: group.name)
                        .font(PVFont.display(size: PVTypeScale.h4, weight: PVFontWeight.semibold))
                        .foregroundStyle(PVColor.textDisplay)
                        .lineLimit(1)
                    Spacer(minLength: 0)
                    PVIconButton(
                        .plus,
                        label: L10n.CitationComposer.addForSubject,
                        size: .sm
                    ) {
                        model.beginAddObservation(for: group.subjectID)
                    }
                    .accessibilityIdentifier("citationComposer.group.add.\(group.subjectID)")
                }
                ForEach(group.lines) { line in
                    lineRow(line)
                }
            }
        }
    }

    private func lineRow(_ line: CitationSavedLine) -> some View {
        let revealed = hoveredLineID == line.id
        return HStack(alignment: .firstTextBaseline, spacing: PVSpacing.space3) {
            if let markKey = line.markKey {
                PVMark(PVMarkKey(catalogKey: markKey), size: 14, decorative: true)
            }
            Button {
                model.beginEdit(line: line)
            } label: {
                HStack(spacing: PVSpacing.space2) {
                    if line.negative {
                        Text(L10n.CitationComposer.notWord)
                            .italic()
                            .foregroundStyle(PVColor.textMuted)
                    }
                    Text(verbatim: line.body.isEmpty ? "—" : line.body)
                        .foregroundStyle(PVColor.textPrimary)
                        .multilineTextAlignment(.leading)
                    if line.negative {
                        PVBadge(L10n.CitationComposer.negatedBadge, tone: .danger)
                    }
                    if model.freshLineKey == line.id || model.freshLineKey == line.sortKey
                        || (line.bridgeID != nil && model.freshLineKey == "bridge:\(line.bridgeID ?? "")")
                    {
                        PVBadge(L10n.CitationComposer.savedBadge, tone: .success)
                    }
                }
                .font(PVFont.body(size: PVTypeScale.bodySmall))
                .frame(maxWidth: .infinity, alignment: .leading)
            }
            .buttonStyle(.plain)
            .accessibilityIdentifier("citationComposer.line.\(line.id)")
            if revealed {
                PVIconButton(.duplicate, label: L10n.CitationComposer.duplicate, size: .sm) {
                    model.duplicate(line: line)
                }
                PVIconButton(.ellipsis, label: L10n.CitationComposer.moreActions, size: .sm) {
                    menuLine = line
                    lineMenu.present(at: CGPoint(x: 0, y: PVSpacing.controlHeightSmall))
                }
            }
        }
        .padding(.vertical, PVSpacing.space2)
        .onHover { hovering in
            hoveredLineID = hovering ? line.id : (hoveredLineID == line.id ? nil : hoveredLineID)
        }
    }

    private var lineMenuBinding: Binding<PVContextMenuState> {
        $lineMenu
    }

    @ViewBuilder
    private var lineMenuContent: some View {
        if let line = menuLine {
            PVContextMenuItem(L10n.CitationComposer.editObservation, index: 0) {
                model.beginEdit(line: line)
            }
            PVContextMenuItem(L10n.CitationComposer.duplicate, index: 1) {
                model.duplicate(line: line)
            }
            if line.bridgeID == nil {
                PVContextMenuItem(
                    line.negative ? L10n.CitationComposer.affirmObservation : L10n.CitationComposer.negateObservation,
                    index: 2
                ) {
                    if let observationID = line.observationID {
                        Task { await model.negate(observationID: observationID) }
                    }
                }
                PVContextMenuItem(L10n.CitationComposer.deleteObservation, index: 3) {
                    Task { await model.askDeleteLine(line) }
                }
            } else {
                PVContextMenuItem(L10n.CitationComposer.showOnGraph, index: 2) {
                    if let bridgeID = line.bridgeID {
                        model.showOnGraph(bridgeID: bridgeID)
                    }
                }
                PVContextMenuItem(L10n.CitationComposer.deleteBridge, index: 3) {
                    Task { await model.askDeleteLine(line) }
                }
            }
        }
    }
}

/// Edit sheet for one saved property observation or one bridge.
struct CitationComposerEditSheet: View {
    @Bindable var model: CitationComposerModel

    var body: some View {
        if let session = model.editSession {
            switch session.kind {
            case .property(let values, _, _):
                propertyForm(values, error: session.error)
            case .bridge(let endA, let endB, _, _, _, _):
                bridgeForm(endA: endA, endB: endB, error: session.error)
            }
        }
    }

    private func propertyForm(_ values: ObservationRowValues, error: String?) -> some View {
        VStack(alignment: .leading, spacing: PVSpacing.space4) {
            PVComboBox(
                selection: editSubjectBinding,
                options: model.subjectOptions,
                size: .sm,
                placeholder: L10n.CitationComposer.subjectPlaceholder,
                emptyLabel: L10n.CitationComposer.subjectPlaceholder,
                label: L10n.CitationComposer.subjectLabel,
                accessibilityIdentifierPrefix: "citationComposer.edit.subject"
            )
            PVComboBox(
                selection: editPropertyBinding,
                options: model.vocabulary.propertyOptions(forSubjectID: values.subjectID).map {
                    PVComboBoxOption(value: $0.id, label: $0.label, subtext: $0.key)
                },
                size: .sm,
                placeholder: L10n.CitationComposer.propertyPlaceholder,
                emptyLabel: L10n.CitationComposer.propertyPlaceholder,
                label: L10n.CitationComposer.propertyLabel,
                accessibilityIdentifierPrefix: "citationComposer.edit.property"
            )
            editValue(values)
            PVCheckbox(
                L10n.CitationComposer.negativeLabel,
                description: L10n.CitationComposer.negativeDescription,
                isChecked: editNegativeBinding
            )
            if let error {
                PVCallout(tone: .danger, message: error)
            }
        }
    }

    @ViewBuilder
    private func editValue(_ values: ObservationRowValues) -> some View {
        let property = model.vocabulary.property(id: values.propertyID)
        switch property?.valueType {
        case PropertyValueType.text.rawValue:
            PVInput(text: editTextBinding, size: .sm)
        case PropertyValueType.integer.rawValue:
            PVInput(text: editIntegerBinding, size: .sm, mono: true)
        case PropertyValueType.term.rawValue:
            HStack(spacing: PVSpacing.space3) {
                PVComboBox(
                    selection: editTermBinding,
                    options: model.vocabulary.termComboOptionsByPropertyID[values.propertyID] ?? [],
                    size: .sm,
                    placeholder: L10n.CitationComposer.termPlaceholder,
                    emptyLabel: L10n.CitationComposer.termEmpty,
                    label: L10n.CitationComposer.termLabel,
                    accessibilityIdentifierPrefix: "citationComposer.edit.term"
                )
                PVButton(L10n.CitationComposer.addCustomTerm, variant: .ghost, size: .sm) {
                    model.beginAddCustomTerm(propertyID: values.propertyID, applyToEdit: true)
                }
            }
        case PropertyValueType.name.rawValue, PropertyValueType.date.rawValue:
            Button {
                model.beginValueEditor(applyToEdit: true)
            } label: {
                Text(L10n.CitationComposer.editObservation)
            }
            .buttonStyle(.pv(.secondary, size: .sm))
        default:
            EmptyView()
        }
    }

    private func bridgeForm(endA: String, endB: String, error: String?) -> some View {
        let endAName = model.vocabulary.subjectsByID[endA]?.label ?? endA
        let endBName = model.vocabulary.subjectsByID[endB]?.label ?? endB
        let property = bridgeTermProperty()
        return VStack(alignment: .leading, spacing: PVSpacing.space4) {
            Text(L10n.CitationComposer.lockEndsNote)
                .font(PVFont.body(size: PVTypeScale.caption, italic: true))
                .foregroundStyle(PVColor.textMuted)
            labeledReadOnly(L10n.CitationComposer.endA, value: endAName)
            labeledReadOnly(L10n.CitationComposer.endB, value: endBName)
            if let property {
                HStack(spacing: PVSpacing.space3) {
                    PVComboBox(
                        selection: editBridgeTermBinding,
                        options: model.vocabulary.termComboOptionsByPropertyID[property.id] ?? [],
                        size: .sm,
                        placeholder: L10n.CitationComposer.termPlaceholder,
                        emptyLabel: L10n.CitationComposer.termEmpty,
                        label: ConnectionTermCopy.label(disambiguationKey: property.key),
                        accessibilityIdentifierPrefix: "citationComposer.edit.bridgeTerm"
                    )
                    PVButton(L10n.CitationComposer.addCustomTerm, variant: .ghost, size: .sm) {
                        model.beginAddCustomTerm(propertyID: property.id, applyToEdit: true)
                    }
                }
            }
            if let error {
                PVCallout(tone: .danger, message: error)
            }
        }
    }

    private func labeledReadOnly(_ label: LocalizedStringResource, value: String) -> some View {
        VStack(alignment: .leading, spacing: PVSpacing.space1) {
            Text(label)
                .font(PVFont.body(size: PVTypeScale.caption))
                .foregroundStyle(PVColor.textMuted)
            Text(verbatim: value)
                .font(PVFont.body(size: PVTypeScale.bodySmall))
                .foregroundStyle(PVColor.textPrimary)
        }
    }

    private func bridgeTermProperty() -> CatalogProperty? {
        guard let session = model.editSession,
              case .bridge(_, _, _, let bridgeID, _, _) = session.kind,
              let connection = model.connections.rows.first(where: { $0.bridgeID == bridgeID })
        else { return nil }
        return connection.termProperty
    }

    private var editSubjectBinding: Binding<String> {
        Binding(
            get: {
                if case .property(let values, _, _) = model.editSession?.kind { return values.subjectID }
                return ""
            },
            set: { newValue in
                model.mutateEdit { session in
                    guard case .property(var values, let id, let ref) = session.kind else { return }
                    values.subjectID = newValue
                    session.kind = .property(values: values, persistedID: id, persistedRef: ref)
                }
            }
        )
    }

    private var editPropertyBinding: Binding<String> {
        Binding(
            get: {
                if case .property(let values, _, _) = model.editSession?.kind { return values.propertyID }
                return ""
            },
            set: { newValue in
                model.mutateEdit { session in
                    guard case .property(var values, let id, let ref) = session.kind else { return }
                    values.propertyID = newValue
                    session.kind = .property(values: values, persistedID: id, persistedRef: ref)
                }
            }
        )
    }

    private var editTextBinding: Binding<String> {
        Binding(
            get: {
                if case .property(let values, _, _) = model.editSession?.kind { return values.valueText }
                return ""
            },
            set: { newValue in
                model.mutateEdit { session in
                    guard case .property(var values, let id, let ref) = session.kind else { return }
                    values.valueText = newValue
                    session.kind = .property(values: values, persistedID: id, persistedRef: ref)
                }
            }
        )
    }

    private var editIntegerBinding: Binding<String> {
        Binding(
            get: {
                if case .property(let values, _, _) = model.editSession?.kind { return values.valueIntegerText }
                return ""
            },
            set: { newValue in
                model.mutateEdit { session in
                    guard case .property(var values, let id, let ref) = session.kind else { return }
                    values.valueIntegerText = newValue
                    session.kind = .property(values: values, persistedID: id, persistedRef: ref)
                }
            }
        )
    }

    private var editTermBinding: Binding<String> {
        Binding(
            get: {
                if case .property(let values, _, _) = model.editSession?.kind { return values.valueTermID }
                return ""
            },
            set: { newValue in
                model.mutateEdit { session in
                    guard case .property(var values, let id, let ref) = session.kind else { return }
                    values.valueTermID = newValue
                    session.kind = .property(values: values, persistedID: id, persistedRef: ref)
                }
            }
        )
    }

    private var editNegativeBinding: Binding<Bool> {
        Binding(
            get: {
                if case .property(let values, _, _) = model.editSession?.kind {
                    return values.polarity == ObservationPolarity.negative.rawValue
                }
                return false
            },
            set: { checked in
                model.mutateEdit { session in
                    guard case .property(var values, let id, let ref) = session.kind else { return }
                    values.polarity = checked
                        ? ObservationPolarity.negative.rawValue
                        : ObservationPolarity.positive.rawValue
                    session.kind = .property(values: values, persistedID: id, persistedRef: ref)
                }
            }
        )
    }

    private var editBridgeTermBinding: Binding<String> {
        Binding(
            get: {
                if case .bridge(_, _, let termID, _, _, _) = model.editSession?.kind { return termID }
                return ""
            },
            set: { newValue in
                model.mutateEdit { session in
                    guard case .bridge(let endA, let endB, _, let bridgeID, let bridgeRef, let baseline) = session.kind
                    else { return }
                    session.kind = .bridge(
                        endA: endA,
                        endB: endB,
                        termID: newValue,
                        bridgeID: bridgeID,
                        bridgeRef: bridgeRef,
                        baselineTermID: baseline
                    )
                }
            }
        )
    }
}
