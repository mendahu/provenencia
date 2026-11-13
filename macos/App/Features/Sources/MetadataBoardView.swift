import SwiftUI

/// Where a board sits, which decides how a researcher adds a field to it.
enum MetadataBoardPlacement {
    /// On the Source page: the section header's Add field button opens the row at the foot.
    case page
    /// In a form: the board draws its own ghost Add field button at the foot.
    case form
}

/// One owner's metadata board: saved reorderable rows, type suggestions, the
/// empty state and the inline Add field row. The page draws the section header
/// around it, and owns the clear confirm (`MetadataBoardDialogs`) the board asks for.
struct MetadataBoardView: View {
    @Bindable var board: MetadataBoardModel
    let placement: MetadataBoardPlacement
    /// Prefix of every accessibility identifier the board draws (`<prefix>.list`, `<prefix>.<fieldID>.value`, …).
    let accessibilityIdentifierPrefix: String
    let onClear: (String) -> Void

    var body: some View {
        VStack(alignment: .leading, spacing: PVSpacing.space6) {
            if placement == .page, board.saved.isEmpty, board.suggested.isEmpty {
                PVEmptyState(
                    icon: .list,
                    title: L10n.Sources.metadataEmptyTitle,
                    message: L10n.string(L10n.Sources.metadataEmptyMessage),
                    compact: true
                )
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
                                onTyped: { board.setTyped($0, fieldID: entry.field.id) },
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

            if board.addRow != nil {
                MetadataBoardAddRowView(
                    board: board,
                    comboWidth: placement == .form ? 168 : 150,
                    accessibilityIdentifierPrefix: accessibilityIdentifierPrefix
                )
            } else if placement == .form {
                HStack {
                    PVButton(L10n.Sources.addMetadata, variant: .ghost, size: .sm, icon: .plus) {
                        board.openAddRow()
                    }
                    .accessibilityIdentifier("\(accessibilityIdentifierPrefix).add")
                    Spacer(minLength: 0)
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
                onTyped: { board.setTyped($0, fieldID: entry.field.id) },
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

/// The inline Add field row at the foot of a board: a field picker, a value
/// input, ✓ and ✕, and under them the refusal or the "taken" notice. ✓ saves at
/// once; Escape closes the row. While drawn it is the sheet's one inner edit.
private struct MetadataBoardAddRowView: View {
    @Bindable var board: MetadataBoardModel
    let comboWidth: CGFloat
    let accessibilityIdentifierPrefix: String

    private enum Field { case picker }
    @FocusState private var focus: Field?

    var body: some View {
        let prefix = "\(accessibilityIdentifierPrefix).addRow"
        VStack(alignment: .leading, spacing: PVSpacing.space2) {
            PVCard(cornerRadius: PVRadius.sm) {
                HStack(spacing: PVSpacing.space4) {
                    PVComboBox(
                        selection: Binding(
                            get: { board.addRow?.fieldID ?? "" },
                            set: { board.pickAddRowField($0) }
                        ),
                        options: board.fieldComboOptions,
                        size: .sm,
                        placeholder: L10n.Sources.metadataField,
                        emptyLabel: L10n.Sources.typeNoMatch,
                        isInvalid: false,
                        label: L10n.Sources.metadataField,
                        accessibilityIdentifierPrefix: "\(prefix).field",
                        focus: PVFieldFocus($focus, equals: .picker)
                    )
                    .onAppear { DispatchQueue.main.async { focus = .picker } }
                    .frame(width: comboWidth)
                    PVInput(
                        text: Binding(
                            get: { board.addRow?.value ?? "" },
                            set: { board.setAddRowValue($0) }
                        ),
                        size: .sm,
                        mono: true,
                        prompt: L10n.Sources.metadataValue,
                        isInvalid: board.addRow?.error != nil
                    )
                    .onSubmit { Task { await board.confirmAddRow() } }
                    .accessibilityLabel(L10n.string(L10n.Sources.metadataValue))
                    .accessibilityIdentifier("\(prefix).value")
                    PVInlineEditActions(
                        isSaving: board.isSavingAddRow,
                        saveLabel: L10n.Sources.metadataAddRowConfirmLabel,
                        cancelLabel: L10n.Sources.metadataAddRowCancelLabel,
                        saveDisabled: !board.canConfirmAddRow,
                        style: .iconRow,
                        accessibilityIdentifierPrefix: prefix,
                        onSave: { Task { await board.confirmAddRow() } },
                        onCancel: { board.closeAddRow() }
                    )
                }
                .padding(.vertical, PVSpacing.space4)
                .padding(.horizontal, PVSpacing.space5)
            }
            if let error = board.addRow?.error, !error.isEmpty {
                PVErrorText(error, size: .caption, announces: true)
            } else if let taken = board.addRow?.takenFieldLabel {
                Text(L10n.Sources.metadataAddRowFieldTaken(field: taken))
                    .font(PVFont.body(size: PVTypeScale.caption))
                    .foregroundStyle(PVColor.textMuted)
                    .pvAnnounce(L10n.Sources.metadataAddRowFieldTaken(field: taken))
            }
        }
        // The ✕ takes Escape; the enclosing sheet's own Escape stands down while the row is open.
        .pvSheetInnerEdit(true)
    }
}

/// Suggestion quick-add row with a local draft.
private struct MetadataBoardSuggestionRow: View {
    let accessibilityIdentifierPrefix: String
    let entry: CatalogMetadataEntry
    let isSaving: Bool
    var error: String? = nil
    let onTyped: (Bool) -> Void
    let onSave: (String) -> Void
    let onDismiss: () -> Void

    @State private var draft = ""
    @FocusState private var isFocused: Bool

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
                        isInvalid: error != nil,
                        focused: $isFocused
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
        // Typed text in the focused input is an inner edit: Escape clears it before the sheet reacts.
        .pvSheetInnerEdit(isFocused && !draft.isEmpty)
        .onExitCommand { draft = "" }
        .onChange(of: draft) { _, text in onTyped(!text.isEmpty) }
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
    let onTyped: (Bool) -> Void
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
            } else {
                onTyped(false)
            }
        }
        .onChange(of: draft) { _, text in onTyped(isEditing && text != entry.valueText) }
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
            .help(lineLimit == nil
                ? L10n.string(L10n.Sources.openMetadataURL)
                : "\(L10n.string(L10n.Sources.openMetadataURL))\n\(entry.valueText)")
        } else {
            valueText
                .foregroundStyle(PVColor.textPrimary)
                .modifier(MetadataValueChromeIf(chrome: chrome))
                .accessibilityIdentifier(accessibilityIdentifier)
                // A truncated value shows whole on hover.
                .help(lineLimit == nil ? "" : entry.valueText)
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
