import SwiftUI

/// Metadata board: saved reorderable rows, type suggestions, and the Add dialog form.
struct SourcePageMetadataView: View {
    @Bindable var model: SourcePageModel

    var body: some View {
        VStack(alignment: .leading, spacing: PVSpacing.space6) {
            PVSectionHeader(
                title: L10n.Sources.metadataHeading,
                meta: model.metadata.saved.isEmpty
                    ? nil
                    : L10n.Sources.metadataFieldCount(model.metadata.saved.count),
                actions: {
                    PVButton(L10n.Sources.addMetadata, variant: .primary, size: .sm, icon: .plus) {
                        model.metadata.openAdd()
                    }
                    .accessibilityIdentifier("sources.page.addMetadata")
                }
            )

            Text(L10n.Sources.metadataIntro)
                .font(PVFont.body(size: PVTypeScale.caption))
                .foregroundStyle(PVColor.textMuted)
                .frame(maxWidth: PVSpacing.measureProse, alignment: .leading)

            if model.metadata.saved.isEmpty, model.metadata.suggested.isEmpty {
                PVEmptyState(
                    icon: .list,
                    title: L10n.Sources.metadataEmptyTitle,
                    message: String(localized: L10n.Sources.metadataEmptyMessage),
                    compact: true
                ) {
                    PVButton(L10n.Sources.addMetadata, variant: .primary, size: .sm, icon: .plus) {
                        model.metadata.openAdd()
                    }
                    .accessibilityIdentifier("sources.page.metadata.empty.add")
                }
                .accessibilityIdentifier("sources.page.metadata.empty")
            } else {
                if !model.metadata.saved.isEmpty {
                    PVReorderableList(
                        items: model.metadata.saved,
                        freezesHeight: model.metadata.editingFieldID != nil,
                        onMove: { source, destination in
                            Task { await model.metadata.moveSaved(from: source, to: destination) }
                        }
                    ) { entry in
                        savedRow(entry)
                    }
                    .accessibilityIdentifier("sources.page.metadata.list")
                }

                if !model.metadata.suggested.isEmpty {
                    VStack(alignment: .leading, spacing: PVSpacing.space4) {
                        Text(L10n.Sources.metadataSuggestionsHeading)
                            .font(PVFont.body(size: PVTypeScale.micro, weight: PVFontWeight.semibold))
                            .tracking(PVTypeScale.micro * PVTracking.caps)
                            .textCase(.uppercase)
                            .foregroundStyle(PVColor.textFaint)
                            .padding(.top, model.metadata.saved.isEmpty ? 0 : PVSpacing.space7)

                        ForEach(model.metadata.suggested) { entry in
                            SourcePageMetadataSuggestionRow(
                                entry: entry,
                                isSaving: model.metadata.savingFieldID == entry.field.id,
                                error: model.metadata.fieldErrorID == entry.field.id
                                    ? model.metadata.fieldError
                                    : nil,
                                onSave: { value in
                                    model.metadata.drafts[entry.field.id] = value
                                    Task { await model.metadata.save(fieldID: entry.field.id) }
                                },
                                onDismiss: {
                                    Task { await model.metadata.dismissSuggestion(fieldID: entry.field.id) }
                                }
                            )
                        }
                    }
                    .accessibilityIdentifier("sources.page.metadata.suggestions")
                }
            }
        }
    }

    /// Field + value form for the Add Metadata dialog (hosted by the page shell).
    var addForm: some View {
        VStack(alignment: .leading, spacing: PVSpacing.space6) {
            PVField(
                label: L10n.Sources.metadataField,
                hint: L10n.Sources.metadataFieldHint,
                error: model.metadata.addFieldError,
                required: true
            ) {
                PVComboBox(
                    selection: $model.metadata.addFieldID,
                    options: model.metadata.fieldComboOptions,
                    placeholder: L10n.Sources.metadataField,
                    emptyLabel: L10n.Sources.typeNoMatch,
                    isInvalid: model.metadata.addFieldError != nil,
                    label: L10n.Sources.metadataField,
                    accessibilityIdentifierPrefix: "sources.page.addMetadata.field"
                )
                .onChange(of: model.metadata.addFieldID) { _, newValue in
                    if !newValue.isEmpty { model.metadata.addFieldError = nil }
                }
            }
            PVField(
                label: L10n.Sources.metadataValue,
                hint: L10n.Sources.metadataValueHint,
                error: model.metadata.addValueError,
                required: true
            ) {
            PVInput(
                text: $model.metadata.addValue,
                isInvalid: model.metadata.addValueError != nil
            )
            .accessibilityIdentifier("sources.page.addMetadata.value")
            .onChange(of: model.metadata.addValue) { _, _ in
                if model.metadata.addValueError != nil {
                    model.metadata.addValueError = nil
                }
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
                entry: entry,
                isEditing: model.metadata.editingFieldID == entry.field.id,
                isSaving: model.metadata.savingFieldID == entry.field.id,
                error: model.metadata.fieldErrorID == entry.field.id
                    ? model.metadata.fieldError
                    : nil,
                onBeginEdit: { model.metadata.beginEdit(fieldID: entry.field.id) },
                onSave: { value in
                    model.metadata.drafts[entry.field.id] = value
                    Task { await model.metadata.save(fieldID: entry.field.id) }
                },
                onCancel: { model.metadata.cancelEdit() },
                onClear: { model.metadata.askClear(fieldID: entry.field.id) }
            )
        }
        .padding(.vertical, PVSpacing.space3)
        .padding(.horizontal, PVSpacing.space4)
    }
}

/// Suggestion quick-add row with a local draft.
private struct SourcePageMetadataSuggestionRow: View {
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
                    .accessibilityIdentifier("sources.page.metadata.\(fieldID).value")

                    PVButton(
                        L10n.Sources.saveMetadataSuggestion,
                        variant: .ghost,
                        size: .sm,
                        loading: isSaving
                    ) {
                        onSave(draft)
                    }
                    .disabled(draft.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty)
                    .accessibilityIdentifier("sources.page.metadata.\(fieldID).save")

                    PVIconButton(.dismiss, label: L10n.Sources.dismissMetadataSuggestion, size: .sm) {
                        onDismiss()
                    }
                    .accessibilityIdentifier("sources.page.metadata.\(fieldID).dismiss")
                }
                if let error, !error.isEmpty {
                    Text(error)
                        .font(PVFont.body(size: PVTypeScale.caption))
                        .foregroundStyle(PVColor.danger)
                }
            }
            .padding(.vertical, PVSpacing.space4)
            .padding(.horizontal, PVSpacing.space5)
        }
    }
}

/// Text metadata row editor with a **local** draft so keystrokes do not rewrite
/// `SourceMetadataSection.drafts` (and rebuild the reorderable `List`) on every
/// character — that was dropping TextField focus and feeling multi-second laggy.
private struct SourcePageMetadataTextEditor: View {
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
            accessibilityIdentifierPrefix: "sources.page.metadata.\(fieldID)",
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
                .accessibilityIdentifier("sources.page.metadata.\(fieldID).value")
            },
            restingTrailing: {
                PVIconButton(
                    .trash,
                    label: L10n.Sources.deleteMetadataValue,
                    size: .sm,
                    tone: .danger,
                    action: onClear
                )
                .accessibilityIdentifier("sources.page.metadata.\(fieldID).delete")
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
        if entry.field.dataType == CatalogFieldDataType.url {
            Button {
                openSavedURL(entry.valueText)
            } label: {
                Text(verbatim: entry.valueText)
                    .font(PVFont.mono(size: PVTypeScale.caption))
                    .foregroundStyle(PVColor.textLink)
                    .underline(true, color: PVColor.textLink)
                    .fixedSize(horizontal: false, vertical: true)
            }
            .buttonStyle(.plain)
            .sourcePageMetadataValueChrome()
            .accessibilityIdentifier("sources.page.metadata.\(fieldID).value")
            .help(String(localized: L10n.Sources.openMetadataURL))
        } else {
            Text(verbatim: entry.valueText)
                .font(PVFont.mono(size: PVTypeScale.caption))
                .foregroundStyle(PVColor.textPrimary)
                .fixedSize(horizontal: false, vertical: true)
                .sourcePageMetadataValueChrome()
                .accessibilityIdentifier("sources.page.metadata.\(fieldID).value")
        }
    }

    private func openSavedURL(_ typed: String) {
        var s = typed.trimmingCharacters(in: .whitespacesAndNewlines)
        if !s.contains("://") {
            s = "https://" + s
        }
        if let url = URL(string: s) {
            openURL(url)
        }
    }
}
