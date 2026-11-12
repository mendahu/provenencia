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
                            SourcePageMetadataSuggestionRow(
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
        HStack(alignment: .top, spacing: SourcePageLayout.metadataColumnSpacing) {
            PVReorderHandle()
                .padding(.top, 6)
            SourcePageMetadataLabel(
                text: entry.field.label,
                topPadding: 6,
                emphasized: true
            )

            SourcePageMetadataTextEditor(
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
private struct SourcePageMetadataSuggestionRow: View {
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
                HStack(spacing: SourcePageLayout.metadataColumnSpacing) {
                    SourcePageMetadataLabel(text: entry.field.label, alignWithReorderHandle: true)
                    PVInput(
                        text: $draft,
                        size: .sm,
                        mono: true,
                        prompt: L10n.Sources.metadataSuggestionPlaceholder,
                        isInvalid: error != nil
                    )
                    .onSubmit { onSave(draft) }
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

                    PVIconButton(.dismiss, label: L10n.Sources.dismissMetadataSuggestion, size: .sm) {
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
private struct SourcePageMetadataTextEditor: View {
    let accessibilityIdentifierPrefix: String
    let entry: CatalogMetadataEntry
    let isEditing: Bool
    let isSaving: Bool
    var error: String? = nil
    let onBeginEdit: () -> Void
    let onSave: (String) -> Void
    let onCancel: () -> Void
    let onClear: () -> Void

    @Environment(\.openURL) private var openURL
    @State private var draft = ""

    var body: some View {
        let fieldID = entry.field.id
        return PVInlineEdit(
            isEditing: isEditing,
            isSaving: isSaving,
            error: error,
            saveLabel: L10n.Sources.saveMetadataValue,
            cancelLabel: L10n.Sources.cancelEdit,
            editLabel: L10n.Sources.editMetadataValue,
            saveDisabled: draft.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty,
            axis: .horizontal,
            actionsStyle: .iconStack,
            accessibilityIdentifierPrefix: "\(accessibilityIdentifierPrefix).\(fieldID)",
            onEdit: {
                draft = entry.valueText
                onBeginEdit()
            },
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
                    label: L10n.Sources.deleteMetadataValue,
                    size: .sm,
                    tone: .danger,
                    action: onClear
                )
                .accessibilityIdentifier("\(accessibilityIdentifierPrefix).\(fieldID).delete")
            },
            editingTrailing: { EmptyView() }
        )
        .onChange(of: isEditing) { _, editing in
            if editing {
                draft = entry.valueText
            }
        }
    }

    @ViewBuilder
    private var displayValue: some View {
        let fieldID = entry.field.id
        if entry.field.dataType == CatalogFieldDataType.url, let url = entry.openURL {
            Button {
                openURL(url)
            } label: {
                Text(verbatim: entry.valueText)
                    .font(PVFont.mono(size: PVTypeScale.caption))
                    .foregroundStyle(PVColor.textLink)
                    .underline(true, color: PVColor.textLink)
                    .fixedSize(horizontal: false, vertical: true)
            }
            .buttonStyle(.plain)
            .sourcePageMetadataValueChrome()
            .accessibilityIdentifier("\(accessibilityIdentifierPrefix).\(fieldID).value")
            .help(L10n.string(L10n.Sources.openMetadataURL))
        } else {
            Text(verbatim: entry.valueText)
                .font(PVFont.mono(size: PVTypeScale.caption))
                .foregroundStyle(PVColor.textPrimary)
                .fixedSize(horizontal: false, vertical: true)
                .sourcePageMetadataValueChrome()
                .accessibilityIdentifier("\(accessibilityIdentifierPrefix).\(fieldID).value")
        }
    }
}
