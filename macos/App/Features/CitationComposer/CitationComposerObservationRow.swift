import SwiftUI

/// Inline Observation row: subject + property + value, or sunken Connect lock.
struct CitationComposerObservationRow: View {
    let row: CitationComposerModel.ObservationRow
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
    var onRemove: @MainActor () -> Void
    var onAddCustomTerm: @MainActor () -> Void

    @State private var actionsMenu = PVContextMenuState()
    @State private var actionsKeyboard = PVContextMenuKeyboard.inactive

    var body: some View {
        Group {
            if row.isConnectFixed {
                fixedRow
            } else {
                editableRow
            }
        }
    }

    private var editableRow: some View {
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
                    .disabled(inert)
                    PVComboBox(
                        selection: propertyBinding,
                        options: propertyOptions,
                        size: .sm,
                        placeholder: L10n.CitationComposer.propertyPlaceholder,
                        emptyLabel: L10n.CitationComposer.propertyEmpty,
                        label: L10n.CitationComposer.propertyLabel,
                        accessibilityIdentifierPrefix: "citationComposer.observation.property.\(row.id.uuidString)"
                    )
                    .disabled(inert)
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
                valueEditor
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

    private var fixedRow: some View {
        PVCard(tone: .sunken, cornerRadius: PVRadius.sm) {
            HStack(alignment: .center, spacing: PVSpacing.space5) {
                PVIcon(.lock, size: 14)
                    .foregroundStyle(PVColor.textFaint)
                VStack(alignment: .leading, spacing: PVSpacing.space1) {
                    Text(verbatim: property?.label ?? "—")
                        .font(PVFont.body(size: PVTypeScale.caption, weight: PVFontWeight.medium))
                        .foregroundStyle(PVColor.textSecondary)
                    Text(verbatim: summary)
                        .font(PVFont.body(size: PVTypeScale.bodySmall))
                        .foregroundStyle(PVColor.textPrimary)
                }
                .frame(maxWidth: .infinity, alignment: .leading)
                PVBadge(L10n.CitationComposer.connectSystemBadge, tone: .neutral)
            }
            .padding(.horizontal, PVSpacing.space5)
            .padding(.vertical, PVSpacing.space5)
        }
        .accessibilityElement(children: .combine)
        .accessibilityLabel(
            Text(verbatim: "\(property?.label ?? ""), \(summary), \(String(localized: L10n.CitationComposer.connectSystemBadge))")
        )
    }

    @ViewBuilder
    private var valueEditor: some View {
        switch property?.valueType {
        case PropertyValueType.text.rawValue:
            PVInput(text: textBinding, size: .sm)
                .disabled(inert)
                .accessibilityIdentifier("citationComposer.observation.text.\(row.id.uuidString)")
        case PropertyValueType.integer.rawValue:
            PVInput(text: integerBinding, size: .sm, mono: true)
                .disabled(inert)
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
                .disabled(inert)
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
            .disabled(inert)
            .accessibilityLabel(Text(L10n.CitationComposer.editObservation))
            .accessibilityIdentifier("citationComposer.observation.edit.\(row.id.uuidString)")
        default:
            EmptyView()
        }
    }

    private var polarityActionTitle: LocalizedStringResource {
        row.polarity == ObservationPolarity.negative.rawValue
            ? L10n.CitationComposer.polarityAsserts
            : L10n.CitationComposer.polarityNegates
    }

    private var actionsMenuPanel: some View {
        PVContextMenuPanel(width: 180) {
            PVContextMenuItem(
                polarityActionTitle,
                index: 0,
                accessibilityIdentifier: "citationComposer.observation.polarity.\(row.id.uuidString)"
            ) {
                onTogglePolarity()
            }
            PVContextMenuItem(
                L10n.CitationComposer.removeObservation,
                index: 1,
                accessibilityIdentifier: "citationComposer.observation.remove.\(row.id.uuidString)"
            ) {
                onRemove()
            }
        }
    }

    private func toggleActionsMenu() {
        if actionsMenu.isPresented {
            actionsMenu.dismiss()
            return
        }
        actionsKeyboard = PVContextMenuKeyboard(
            itemCount: 2,
            activeIndex: -1,
            itemTitles: [
                String(localized: polarityActionTitle),
                String(localized: L10n.CitationComposer.removeObservation),
            ]
        )
        actionsMenu.present(at: CGPoint(x: 0, y: PVSpacing.controlHeightSmall))
    }

    private var subjectBinding: Binding<String> {
        Binding(
            get: { row.subjectID },
            set: { newValue in onSubject(newValue) }
        )
    }

    private var propertyBinding: Binding<String> {
        Binding(
            get: { row.propertyID },
            set: { newValue in onProperty(newValue) }
        )
    }

    private var textBinding: Binding<String> {
        Binding(
            get: { row.valueText },
            set: { newValue in onText(newValue) }
        )
    }

    private var integerBinding: Binding<String> {
        Binding(
            get: { row.valueIntegerText },
            set: { newValue in onInteger(newValue) }
        )
    }

    private var termBinding: Binding<String> {
        Binding(
            get: { row.valueTermID },
            set: { newValue in onTerm(newValue) }
        )
    }
}
