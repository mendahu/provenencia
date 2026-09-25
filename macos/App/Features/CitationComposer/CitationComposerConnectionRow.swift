import SwiftUI

/// Compact saved or pending connection: sentence plus optional role/type.
struct CitationComposerConnectionRow: View {
    let row: ConnectionRow
    let termOptions: [PVComboBoxOption]
    var inert: Bool
    var onTerm: @MainActor (String) -> Void
    var onSave: @MainActor () -> Void
    var onDiscard: @MainActor () -> Void
    var onCommitRole: @MainActor () -> Void
    var onRevertRole: @MainActor () -> Void

    var body: some View {
        PVCard(tone: .card, padding: 0) {
            VStack(alignment: .leading, spacing: PVSpacing.space3) {
                HStack(alignment: .center, spacing: PVSpacing.space3) {
                    Text(verbatim: row.sentence)
                        .font(PVFont.body(size: PVTypeScale.bodySmall))
                        .foregroundStyle(PVColor.textPrimary)
                        .frame(maxWidth: .infinity, alignment: .leading)
                    statusBadge
                }
                if row.isLocation {
                    Text(L10n.CitationComposer.connectionNoRole)
                        .font(PVFont.body(size: PVTypeScale.caption))
                        .foregroundStyle(PVColor.textMuted)
                } else if let _ = row.termProperty {
                    PVComboBox(
                        selection: termBinding,
                        options: termOptions,
                        size: .sm,
                        placeholder: L10n.CitationComposer.termPlaceholder,
                        emptyLabel: L10n.CitationComposer.termEmpty,
                        label: termLabelResource,
                        accessibilityIdentifierPrefix: "citationComposer.connection.term.\(row.id.uuidString)"
                    )
                    .disabled(inert)
                }
                if let error = row.error {
                    PVCallout(tone: .danger, message: error, compact: true)
                }
                HStack {
                    Spacer(minLength: 0)
                    if row.isPending {
                        PVButton(L10n.CitationComposer.discardConnection, variant: .ghost, size: .sm, action: onDiscard)
                            .disabled(inert)
                        PVButton(
                            L10n.CitationComposer.saveConnection,
                            variant: .primary,
                            size: .sm,
                            loading: row.isSaving,
                            action: onSave
                        )
                        .disabled(inert || !row.canSave)
                    } else if row.termProperty != nil {
                        PVButton(L10n.CitationComposer.revertRole, variant: .ghost, size: .sm, action: onRevertRole)
                            .disabled(inert || !row.isTouched)
                        PVButton(
                            L10n.CitationComposer.saveRole,
                            variant: .secondary,
                            size: .sm,
                            loading: row.isSaving,
                            action: onCommitRole
                        )
                        .disabled(inert || !row.canCommitRole)
                    }
                }
            }
            .padding(PVSpacing.space5)
        }
        .accessibilityElement(children: .contain)
        .accessibilityLabel(Text(verbatim: row.accessibilityLabel))
        .accessibilityIdentifier("citationComposer.connection.row.\(row.id.uuidString)")
    }

    @ViewBuilder
    private var statusBadge: some View {
        if row.isPending {
            PVBadge(L10n.CitationComposer.rowStateNew, tone: .info)
        } else if row.isTouched {
            PVBadge(L10n.CitationComposer.rowStateEdited, tone: .warning)
        } else if let ref = row.bridgeRef, !ref.isEmpty {
            Text(verbatim: ref)
                .font(PVFont.mono(size: PVTypeScale.caption))
                .foregroundStyle(PVColor.textMuted)
        }
    }

    private var termLabelResource: LocalizedStringResource {
        switch row.bridgeTypeKey {
        case "relationship":
            return L10n.CitationComposer.connectionRelationship
        case "participation":
            return L10n.CitationComposer.connectionRole
        default:
            return L10n.CitationComposer.connectionNoRole
        }
    }

    private var termBinding: Binding<String> {
        Binding(
            get: { row.roleTermID },
            set: { onTerm($0) }
        )
    }
}
