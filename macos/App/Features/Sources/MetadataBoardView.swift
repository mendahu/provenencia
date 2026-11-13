import SwiftUI

/// One owner's metadata board: saved reorderable rows, type suggestions, and the
/// empty state. The page draws the section header and intro around it, and owns
/// the Add dialog and clear confirm (`MetadataBoardDialogs`) the board asks for.
struct MetadataBoardView: View {
    @Bindable var board: MetadataBoardModel
    /// Prefix of every accessibility identifier the board draws (`<prefix>.list`, `<prefix>.<fieldID>.value`, …).
    let accessibilityIdentifierPrefix: String
    let onAdd: () -> Void
    let onClear: (String) -> Void

    var body: some View {
        VStack(alignment: .leading, spacing: PVSpacing.space6) {
            if board.saved.isEmpty, board.suggested.isEmpty {
                PVEmptyState(
                    icon: .list,
                    title: L10n.Sources.metadataEmptyTitle,
                    message: L10n.string(L10n.Sources.metadataEmptyMessage),
                    compact: true
                ) {
                    PVButton(L10n.Sources.addMetadata, variant: .primary, size: .sm, icon: .plus) {
                        onAdd()
                    }
                    .accessibilityIdentifier("\(accessibilityIdentifierPrefix).empty.add")
                }
                .accessibilityIdentifier("\(accessibilityIdentifierPrefix).empty")
            } else {
                if !board.saved.isEmpty {
                    PVReorderableList(
                        items: board.saved,
                        freezesHeight: board.editingFieldID != nil,
                        onMove: { source, destination in
                            Task { await board.moveSaved(from: source, to: destination) }
                        }
                    ) { entry in
                        savedRow(entry)
                    }
                    .accessibilityIdentifier("\(accessibilityIdentifierPrefix).list")
                }

                if !board.suggested.isEmpty {
                    VStack(alignment: .leading, spacing: PVSpacing.space4) {
                        Text(L10n.Sources.metadataSuggestionsHeading)
                            .font(PVFont.body(size: PVTypeScale.micro, weight: PVFontWeight.semibold))
                            .tracking(PVTypeScale.micro * PVTracking.caps)
                            .textCase(.uppercase)
                            .foregroundStyle(PVColor.textFaint)
                            .padding(.top, board.saved.isEmpty ? 0 : PVSpacing.space7)

                        ForEach(board.suggested) { entry in
                            MetadataBoardSuggestionRow(
                                accessibilityIdentifierPrefix: accessibilityIdentifierPrefix,
                                entry: entry,
                                isSaving: board.savingFieldID == entry.field.id,
                                error: board.fieldErrorID == entry.field.id
                                    ? board.fieldError
                                    : nil,
                                onSave: { value in
                                    board.drafts[entry.field.id] = value
                                    Task { await board.save(fieldID: entry.field.id) }
                                },
                                onDismiss: {
                                    Task { await board.dismissSuggestion(fieldID: entry.field.id) }
                                }
                            )
                        }
                    }
                    .accessibilityIdentifier("\(accessibilityIdentifierPrefix).suggestions")
                }
            }
        }
    }

    private func savedRow(_ entry: CatalogMetadataEntry) -> some View {
        HStack(alignment: .top, spacing: PVSpacing.space5) {
            PVReorderHandle()
                .padding(.top, 6)
            MetadataBoardLabel(
                text: entry.field.label,
                topPadding: 6,
                emphasized: true
            )

            MetadataBoardTextEditor(
                accessibilityIdentifierPrefix: accessibilityIdentifierPrefix,
                entry: entry,
                isEditing: board.editingFieldID == entry.field.id,
                isSaving: board.savingFieldID == entry.field.id,
                error: board.fieldErrorID == entry.field.id
                    ? board.fieldError
                    : nil,
                onBeginEdit: { board.beginEdit(fieldID: entry.field.id) },
                onSave: { value in
                    board.drafts[entry.field.id] = value
                    Task { await board.save(fieldID: entry.field.id) }
                },
                onCancel: { board.cancelEdit() },
                onClear: { onClear(entry.field.id) }
            )
        }
        .padding(.vertical, PVSpacing.space3)
        .padding(.horizontal, PVSpacing.space4)
    }
}

/// Field + value form for the Add Metadata dialog (hosted by the page shell).
struct MetadataBoardAddForm: View {
    @Bindable var dialogs: MetadataBoardDialogs
    /// Prefix of the form's identifiers (`<prefix>.field`, `<prefix>.value`).
    let accessibilityIdentifierPrefix: String

    var body: some View {
        VStack(alignment: .leading, spacing: PVSpacing.space6) {
            PVField(
                label: L10n.Sources.metadataField,
                hint: L10n.Sources.metadataFieldHint,
                error: dialogs.addFieldError,
                marker: .required
            ) {
                PVComboBox(
                    selection: $dialogs.addFieldID,
                    options: dialogs.fieldComboOptions,
                    placeholder: L10n.Sources.metadataField,
                    emptyLabel: L10n.Sources.typeNoMatch,
                    isInvalid: dialogs.addFieldError != nil,
                    label: L10n.Sources.metadataField,
                    accessibilityIdentifierPrefix: "\(accessibilityIdentifierPrefix).field"
                )
                .onChange(of: dialogs.addFieldID) { _, newValue in
                    if !newValue.isEmpty { dialogs.addFieldError = nil }
                }
            }
            PVField(
                label: L10n.Sources.metadataValue,
                hint: L10n.Sources.metadataValueHint,
                error: dialogs.addValueError,
                marker: .required
            ) {
                PVInput(
                    text: $dialogs.addValue,
                    isInvalid: dialogs.addValueError != nil
                )
                .accessibilityIdentifier("\(accessibilityIdentifierPrefix).value")
                .onChange(of: dialogs.addValue) { _, _ in
                    if dialogs.addValueError != nil {
                        dialogs.addValueError = nil
                    }
                }
            }
        }
    }
}

/// Suggestion quick-add row with a local draft.
private struct MetadataBoardSuggestionRow: View {
    let accessibilityIdentifierPrefix: String
    let entry: CatalogMetadataEntry
    let isSaving: Bool
    var error: String? = nil
    let onSave: (String) -> Void
    let onDismiss: () -> Void

    @State private var draft = ""

    var body: some View {
        let fieldID = entry.field.id
        return PVCard(border: .dashed, cornerRadius: PVRadius.sm) {
            VStack(alignment: .leading, spacing: PVSpacing.space2) {
                HStack(spacing: PVSpacing.space5) {
                    MetadataBoardLabel(
                        text: entry.field.label,
                        leadingInset: PVReorderHandle.width
                    )
                    PVInput(
                        text: $draft,
                        size: .sm,
                        mono: true,
                        prompt: L10n.Sources.metadataSuggestionPlaceholder,
                        isInvalid: error != nil
                    )
                    .onSubmit { onSave(draft) }
                    .accessibilityLabel(entry.field.label)
                    .accessibilityIdentifier("\(accessibilityIdentifierPrefix).\(fieldID).value")

                    PVButton(
                        L10n.Sources.saveMetadataSuggestion,
                        variant: .ghost,
                        size: .sm,
                        loading: isSaving
                    ) {
                        onSave(draft)
                    }
                    .disabled(draft.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty)
                    .accessibilityIdentifier("\(accessibilityIdentifierPrefix).\(fieldID).save")

                    PVIconButton(
                        .dismiss,
                        label: L10n.Sources.dismissMetadataSuggestionLabel(field: entry.field.label),
                        size: .sm
                    ) {
                        onDismiss()
                    }
                    .accessibilityIdentifier("\(accessibilityIdentifierPrefix).\(fieldID).dismiss")
                }
                if let error, !error.isEmpty {
                    PVErrorText(error, size: .caption)
                }
            }
            .padding(.vertical, PVSpacing.space4)
            .padding(.horizontal, PVSpacing.space5)
        }
    }
}

/// Text metadata row editor with a **local** draft so keystrokes do not rewrite
/// `MetadataBoardModel.drafts` (and rebuild the reorderable `List`) on every
/// character — that was dropping TextField focus and feeling multi-second laggy.
private struct MetadataBoardTextEditor: View {
    let accessibilityIdentifierPrefix: String
    let entry: CatalogMetadataEntry
    let isEditing: Bool
    let isSaving: Bool
    var error: String? = nil
    let onBeginEdit: () -> Void
    let onSave: (String) -> Void
    let onCancel: () -> Void
    let onClear: () -> Void

    @State private var draft = ""

    private func beginEdit() {
        draft = entry.valueText
        onBeginEdit()
    }

    var body: some View {
        let fieldID = entry.field.id
        return PVInlineEdit(
            isEditing: isEditing,
            isSaving: isSaving,
            error: error,
            saveLabel: L10n.Sources.saveMetadataValue,
            cancelLabel: L10n.Sources.cancelEdit,
            editLabel: L10n.Sources.editMetadataValueLabel(field: entry.field.label),
            saveDisabled: draft.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty,
            axis: .horizontal,
            actionsStyle: .iconStack,
            accessibilityIdentifierPrefix: "\(accessibilityIdentifierPrefix).\(fieldID)",
            onEdit: beginEdit,
            onSave: { onSave(draft) },
            onCancel: {
                draft = entry.valueText
                onCancel()
            },
            display: { displayValue },
            editor: {
                PVTextArea(
                    text: $draft,
                    lineLimit: 1...6,
                    typography: .mono,
                    isInvalid: error != nil,
                    activateOnAppear: true
                )
                .onSubmit { onSave(draft) }
                .disabled(isSaving)
                .accessibilityIdentifier("\(accessibilityIdentifierPrefix).\(fieldID).value")
            },
            restingTrailing: {
                PVIconButton(
                    .trash,
                    label: L10n.Sources.deleteMetadataValueLabel(field: entry.field.label),
                    size: .sm,
                    tone: .danger,
                    action: onClear
                )
                .accessibilityIdentifier("\(accessibilityIdentifierPrefix).\(fieldID).delete")
            },
            editingTrailing: { EmptyView() }
        )
        .onChange(of: isEditing, initial: true) { _, editing in
            if editing {
                draft = entry.valueText
            }
        }
    }

    private var displayValue: some View {
        MetadataValueText(
            entry: entry,
            accessibilityIdentifier: "\(accessibilityIdentifierPrefix).\(entry.field.id).value",
            lineLimit: nil,
            chrome: true
        )
    }
}

/// One metadata value, drawn the same on a board row and in a card's read-only
/// copy details: mono caption, and a URL value with a safe link as a plain
/// underlined link button. `chrome` adds the board row's resting inset so
/// entering edit does not shift glyphs; `lineLimit` nil wraps, 1 truncates the tail.
struct MetadataValueText: View {
    let entry: CatalogMetadataEntry
    let accessibilityIdentifier: String
    var lineLimit: Int?
    var chrome: Bool

    @Environment(\.openURL) private var openURL

    var body: some View {
        if entry.field.dataType == CatalogFieldDataType.url, let url = entry.openURL {
            Button {
                openURL(url)
            } label: {
                valueText
                    .foregroundStyle(PVColor.textLink)
                    .underline(true, color: PVColor.textLink)
            }
            .buttonStyle(.plain)
            .modifier(MetadataValueChromeIf(chrome: chrome))
            .accessibilityIdentifier(accessibilityIdentifier)
            .help(L10n.string(L10n.Sources.openMetadataURL))
        } else {
            valueText
                .foregroundStyle(PVColor.textPrimary)
                .modifier(MetadataValueChromeIf(chrome: chrome))
                .accessibilityIdentifier(accessibilityIdentifier)
        }
    }

    private var valueText: some View {
        Text(verbatim: entry.valueText)
            .font(PVFont.mono(size: PVTypeScale.caption))
            .lineLimit(lineLimit)
            .truncationMode(.tail)
            .fixedSize(horizontal: false, vertical: lineLimit == nil)
    }
}

private struct MetadataValueChromeIf: ViewModifier {
    let chrome: Bool

    @ViewBuilder
    func body(content: Content) -> some View {
        if chrome { content.metadataBoardValueChrome() } else { content }
    }
}

/// Resting metadata value chrome that matches `PVTextArea` `.mono` insets so
/// entering inline edit does not shift the glyphs when the border appears.
private struct MetadataBoardValueChrome: ViewModifier {
    func body(content: Content) -> some View {
        content
            .padding(.horizontal, PVInputChrome.horizontalInset)
            .padding(.vertical, PVTextArea.Typography.mono.verticalPadding)
            .overlay(
                RoundedRectangle(cornerRadius: PVRadius.sm, style: .continuous)
                    .stroke(Color.clear, lineWidth: 1)
            )
    }
}

private extension View {
    func metadataBoardValueChrome() -> some View {
        modifier(MetadataBoardValueChrome())
    }
}

/// Label column used by metadata saved / suggestion rows.
private struct MetadataBoardLabel: View {
    static let width: CGFloat = 132

    let text: String
    /// Leading inset so suggestion labels align with saved-row labels.
    var leadingInset: CGFloat = 0
    var topPadding: CGFloat = 0
    /// Saved rows use a stronger field name; suggestions stay a muted caption.
    var emphasized: Bool = false

    var body: some View {
        Text(verbatim: text)
            .font(
                emphasized
                    ? PVFont.body(size: PVTypeScale.bodySmall, weight: PVFontWeight.semibold)
                    : PVFont.body(size: PVTypeScale.caption)
            )
            .foregroundStyle(emphasized ? PVColor.textPrimary : PVColor.textMuted)
            .fixedSize(horizontal: false, vertical: true)
            .frame(width: Self.width, alignment: .leading)
            .padding(.leading, leadingInset)
            .padding(.top, topPadding)
    }
}
