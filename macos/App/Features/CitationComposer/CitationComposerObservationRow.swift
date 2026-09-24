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
    var onSubject: (String) -> Void
    var onProperty: (String) -> Void
    var onText: (String) -> Void
    var onInteger: (String) -> Void
    var onTerm: (String) -> Void
    var onEditValue: () -> Void
    var onTogglePolarity: () -> Void
    var onRemove: () -> Void
    var onAddCustomTerm: () -> Void

    @State private var actionsMenu = PVContextMenuState()

    var body: some View {
        Group {
            if row.isConnectFixed {
                fixedRow
            } else {
                editableRow
            }
        }
        .pvContextMenu($actionsMenu, dismissOnClickAway: true) {
            PVContextMenuPanel(width: 180) {
                PVContextMenuItem(
                    row.polarity == "negative"
                        ? L10n.CitationComposer.polarityAsserts
                        : L10n.CitationComposer.polarityNegates,
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
                            action: { actionsMenu.present(at: .zero) }
                        )
                        .accessibilityIdentifier("citationComposer.observation.actions.\(row.id.uuidString)")
                    }
                }
                valueEditor
            }
            .padding(12)
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
            HStack(alignment: .center, spacing: 10) {
                PVIcon(.lock, size: 14)
                    .foregroundStyle(PVColor.textFaint)
                VStack(alignment: .leading, spacing: 2) {
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
            .padding(.horizontal, 12)
            .padding(.vertical, 10)
        }
        .accessibilityElement(children: .combine)
        .accessibilityLabel(
            Text(verbatim: "\(property?.label ?? ""), \(summary), \(String(localized: L10n.CitationComposer.connectSystemBadge))")
        )
    }

    @ViewBuilder
    private var valueEditor: some View {
        switch property?.valueType {
        case "text":
            PVInput(text: textBinding, size: .sm)
                .disabled(inert)
                .accessibilityIdentifier("citationComposer.observation.text.\(row.id.uuidString)")
        case "integer":
            PVInput(text: integerBinding, size: .sm, mono: true)
                .disabled(inert)
                .accessibilityIdentifier("citationComposer.observation.integer.\(row.id.uuidString)")
        case "term":
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
        case "name", "date":
            Button(action: onEditValue) {
                HStack {
                    if summary.isEmpty {
                        Text(L10n.CitationComposer.editValuePlaceholder)
                    } else {
                        Text(verbatim: summary)
                            .font(property?.valueType == "date" ? PVFont.mono(size: PVTypeScale.bodySmall) : PVFont.body(size: PVTypeScale.bodySmall))
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

    private var subjectBinding: Binding<String> {
        Binding(get: { row.subjectID }, set: onSubject)
    }

    private var propertyBinding: Binding<String> {
        Binding(get: { row.propertyID }, set: onProperty)
    }

    private var textBinding: Binding<String> {
        Binding(get: { row.valueText }, set: onText)
    }

    private var integerBinding: Binding<String> {
        Binding(get: { row.valueIntegerText }, set: onInteger)
    }

    private var termBinding: Binding<String> {
        Binding(get: { row.valueTermID }, set: onTerm)
    }
}
