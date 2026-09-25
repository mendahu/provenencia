import SwiftUI

/// Inline Observation row: subject + property + value, with row-level Save.
struct CitationComposerObservationRow: View {
    let row: ObservationRow
    let property: CatalogProperty?
    let propertyOptions: [PVComboBoxOption]
    let subjectOptions: [PVComboBoxOption]
    let termOptions: [PVComboBoxOption]
    let summary: String
    var isFocused: Bool
    var inert: Bool
    var onSubject: @MainActor (String) -> Void
    var onProperty: @MainActor (String) -> Void
    var onText: @MainActor (String) -> Void
    var onInteger: @MainActor (String) -> Void
    var onTerm: @MainActor (String) -> Void
    var onEditValue: @MainActor () -> Void
    var onTogglePolarity: @MainActor () -> Void
    var onSave: @MainActor () -> Void
    var onRevert: @MainActor () -> Void
    var onRequestDelete: @MainActor () -> Void
    var onAddCustomTerm: @MainActor () -> Void

    @State private var actionsMenu = PVContextMenuState()
    @State private var actionsKeyboard = PVContextMenuKeyboard.inactive

    var body: some View {
        PVCard(tone: isFocused ? .raised : .card, padding: 0) {
            VStack(alignment: .leading, spacing: PVSpacing.space3) {
                HStack(alignment: .center, spacing: PVSpacing.space3) {
                    PVComboBox(
                        selection: subjectBinding,
                        options: subjectOptions,
                        size: .sm,
                        placeholder: L10n.CitationComposer.subjectPlaceholder,
                        emptyLabel: L10n.CitationComposer.subjectEmpty,
                        label: L10n.CitationComposer.subjectLabel,
                        accessibilityIdentifierPrefix: "citationComposer.observation.subject.\(row.id.uuidString)"
                    )
                    .disabled(inert || isSaving)
                    PVComboBox(
                        selection: propertyBinding,
                        options: propertyOptions,
                        size: .sm,
                        placeholder: L10n.CitationComposer.propertyPlaceholder,
                        emptyLabel: L10n.CitationComposer.propertyEmpty,
                        label: L10n.CitationComposer.propertyLabel,
                        accessibilityIdentifierPrefix: "citationComposer.observation.property.\(row.id.uuidString)"
                    )
                    .disabled(inert || isSaving)
                    statusBadge
                    if !inert {
                        PVIconButton(
                            .ellipsis,
                            label: L10n.CitationComposer.observationActions,
                            size: .sm,
                            action: toggleActionsMenu
                        )
                        .accessibilityIdentifier("citationComposer.observation.actions.\(row.id.uuidString)")
                        .pvContextMenu(
                            $actionsMenu,
                            keyboard: $actionsKeyboard,
                            dismissOnClickAway: true
                        ) {
                            actionsMenuPanel
                        }
                    }
                }
                if let error = row.propertyError {
                    Text(verbatim: error)
                        .font(PVFont.body(size: PVTypeScale.caption))
                        .foregroundStyle(PVColor.danger)
                }
                if case .error(let message) = row.state {
                    PVCallout(tone: .danger, message: message, compact: true)
                }
                valueEditor
                HStack {
                    Spacer(minLength: 0)
                    if canRevert {
                        PVButton(L10n.CitationComposer.revertRow, variant: .ghost, size: .sm, action: onRevert)
                            .disabled(inert || isSaving)
                    }
                    if row.canSave {
                        PVButton(
                            L10n.CitationComposer.saveRow,
                            variant: .secondary,
                            size: .sm,
                            loading: isSaving,
                            action: onSave
                        )
                        .disabled(inert)
                    }
                }
            }
            .padding(PVSpacing.space5)
        }
        .overlay {
            if isFocused {
                RoundedRectangle(cornerRadius: PVRadius.sm, style: .continuous)
                    .stroke(PVColor.accent, lineWidth: 1)
            }
        }
        .accessibilityIdentifier("citationComposer.observation.row.\(row.id.uuidString)")
        .accessibilityAddTraits(isFocused ? .isSelected : [])
    }

    private var isSaving: Bool {
        if case .saving = row.state { return true }
        return false
    }

    private var canRevert: Bool {
        switch row.state {
        case .draft, .edited, .error:
            return true
        case .saved, .saving:
            return false
        }
    }

    @ViewBuilder
    private var statusBadge: some View {
        switch row.state {
        case .draft:
            PVBadge(L10n.CitationComposer.rowStateNew, tone: .info)
        case .edited:
            PVBadge(L10n.CitationComposer.rowStateEdited, tone: .warning)
        case .saving:
            PVBadge(L10n.CitationComposer.rowStateSaving, tone: .neutral)
        case .error:
            PVBadge(L10n.CitationComposer.rowStateError, tone: .danger)
        case .saved:
            if let ref = row.persistedRef, !ref.isEmpty {
                Text(verbatim: ref)
                    .font(PVFont.mono(size: PVTypeScale.caption))
                    .foregroundStyle(PVColor.textMuted)
            }
        }
    }

    @ViewBuilder
    private var valueEditor: some View {
        switch property?.valueType {
        case PropertyValueType.text.rawValue:
            PVInput(text: textBinding, size: .sm)
                .disabled(inert || isSaving)
                .accessibilityIdentifier("citationComposer.observation.text.\(row.id.uuidString)")
        case PropertyValueType.integer.rawValue:
            PVInput(text: integerBinding, size: .sm, mono: true)
                .disabled(inert || isSaving)
                .accessibilityIdentifier("citationComposer.observation.integer.\(row.id.uuidString)")
        case PropertyValueType.term.rawValue:
            HStack(spacing: PVSpacing.space3) {
                PVComboBox(
                    selection: termBinding,
                    options: termOptions,
                    size: .sm,
                    placeholder: L10n.CitationComposer.termPlaceholder,
                    emptyLabel: L10n.CitationComposer.termEmpty,
                    label: L10n.CitationComposer.termLabel,
                    accessibilityIdentifierPrefix: "citationComposer.observation.term.\(row.id.uuidString)"
                )
                .disabled(inert || isSaving)
                if !inert {
                    PVButton(L10n.CitationComposer.addCustomTerm, variant: .ghost, size: .sm, action: onAddCustomTerm)
                }
            }
        case PropertyValueType.name.rawValue, PropertyValueType.date.rawValue:
            Button(action: onEditValue) {
                HStack {
                    if summary.isEmpty {
                        Text(L10n.CitationComposer.editValuePlaceholder)
                    } else {
                        Text(verbatim: summary)
                            .font(property?.valueType == PropertyValueType.date.rawValue ? PVFont.mono(size: PVTypeScale.bodySmall) : PVFont.body(size: PVTypeScale.bodySmall))
                    }
                    Spacer(minLength: 0)
                    PVIcon(.penLine, size: 12)
                }
            }
            .buttonStyle(.pv(.secondary, size: .sm))
            .disabled(inert || isSaving)
            .accessibilityLabel(Text(L10n.CitationComposer.editObservation))
            .accessibilityIdentifier("citationComposer.observation.edit.\(row.id.uuidString)")
        default:
            EmptyView()
        }
    }

    private var polarityActionTitle: LocalizedStringResource {
        row.polarity == ObservationPolarity.negative.rawValue
            ? L10n.CitationComposer.affirmObservation
            : L10n.CitationComposer.negateObservation
    }

    private var actionsMenuPanel: some View {
        PVContextMenuPanel(width: 220) {
            PVContextMenuItem(
                polarityActionTitle,
                index: 0,
                accessibilityIdentifier: "citationComposer.observation.polarity.\(row.id.uuidString)"
            ) {
                onTogglePolarity()
            }
            if row.persistedID != nil {
                PVContextMenuItem(
                    L10n.CitationComposer.removeObservation,
                    index: 1,
                    accessibilityIdentifier: "citationComposer.observation.remove.\(row.id.uuidString)"
                ) {
                    onRequestDelete()
                }
            }
        }
    }

    private func toggleActionsMenu() {
        if actionsMenu.isPresented {
            actionsMenu.dismiss()
            return
        }
        var titles = [String(localized: polarityActionTitle)]
        if row.persistedID != nil {
            titles.append(String(localized: L10n.CitationComposer.removeObservation))
        }
        actionsKeyboard = PVContextMenuKeyboard(
            itemCount: titles.count,
            activeIndex: -1,
            itemTitles: titles
        )
        actionsMenu.present(at: CGPoint(x: 0, y: PVSpacing.controlHeightSmall))
    }

    private var subjectBinding: Binding<String> {
        Binding(get: { row.subjectID }, set: { onSubject($0) })
    }

    private var propertyBinding: Binding<String> {
        Binding(get: { row.propertyID }, set: { onProperty($0) })
    }

    private var textBinding: Binding<String> {
        Binding(get: { row.valueText }, set: { onText($0) })
    }

    private var integerBinding: Binding<String> {
        Binding(get: { row.valueIntegerText }, set: { onInteger($0) })
    }

    private var termBinding: Binding<String> {
        Binding(get: { row.valueTermID }, set: { onTerm($0) })
    }
}
