import SwiftUI

/// The right pane of `SourceTypesView`: the add-type form, a selected
/// type's detail (editable for `user` / `provenencia`, locked read-only for
/// `plugin:…`) with the fields it suggests, or an empty prompt when nothing
/// is selected — S2-03 §3.2/3.3.
struct SourceTypesDetailPane: View {
    @Bindable var model: SourceTypesModel
    @State private var iconPickerOpen = false

    var body: some View {
        ScrollView {
            content
        }
        .background(PVColor.surfaceCard)
        .accessibilityIdentifier("sourceTypes.detail")
        .sheet(isPresented: $iconPickerOpen) {
            SourceTypeIconPickerSheet(
                selection: Binding(
                    get: {
                        model.draft?.iconKey ?? PVEvidenceIconKey.defaultTypeIcon.rawValue
                    },
                    set: { newValue in
                        guard var draft = model.draft else { return }
                        draft.iconKey = newValue
                        model.draft = draft
                    }
                ),
                onDone: { iconPickerOpen = false }
            )
        }
        .onChange(of: panelIdentity) { _, _ in
            iconPickerOpen = false
        }
    }

    @ViewBuilder
    private var content: some View {
        if model.isAdding {
            panel(isLocked: false)
        } else if model.selectedType != nil {
            panel(isLocked: model.isSelectedTypeLocked)
        } else {
            PVEmptyState(
                icon: .library,
                title: L10n.SourceTypes.panelEmptyTitle,
                message: String(localized: L10n.SourceTypes.panelEmptyBody),
                compact: true
            )
            .padding(PVSpacing.space9)
        }
    }

    private func panel(isLocked: Bool) -> some View {
        VStack(alignment: .leading, spacing: PVSpacing.space7) {
            panelHeader
            if isLocked {
                lockedNote
            } else {
                form
            }
            if model.isAdding {
                addSuggestionsNote
            } else {
                suggestedFields(isLocked: isLocked)
            }
        }
        .padding(PVSpacing.space8)
        .frame(maxWidth: .infinity, alignment: .leading)
        // Force a fresh subtree when switching add / view / edit so form
        // bindings are not reused across modes that no longer own a draft.
        .id(panelIdentity)
    }

    private var panelIdentity: String {
        switch model.mode {
        case .empty: "empty"
        case .adding: "adding"
        case .viewing(let id): "viewing-\(id)"
        case .editing(let id): "editing-\(id)"
        }
    }

    // MARK: Header (shared by add / locked / editable)

    private var panelHeader: some View {
        VocabularyPanelHeader(
            eyebrow: model.isAdding ? L10n.SourceTypes.detailEyebrowNewType : L10n.SourceTypes.detailEyebrowType,
            title: panelTitle,
            origin: model.isAdding ? CatalogOrigin.user : (model.selectedType?.origin ?? CatalogOrigin.user),
            keyText: panelKey,
            iconKey: panelIconKey,
            usageLine: model.isAdding ? nil : model.selectedType.map {
                String(localized: L10n.SourceTypes.usage(count: $0.usedBy))
            },
            keyHint: model.isAdding ? L10n.SourceTypes.keyHintAdd : L10n.SourceTypes.keyHintEdit,
            showsDelete: model.showsDelete,
            canDelete: model.canDeleteSelectedType,
            deleteTooltip: model.deleteTooltip,
            identifierPrefix: "sourceTypes",
            onDelete: { model.askDelete() }
        )
    }

    private var panelIconKey: String? {
        if model.isAdding {
            return model.draft?.iconKey
        }
        return model.selectedType?.iconKey ?? model.draft?.iconKey
    }

    private var panelTitle: String {
        if model.isAdding {
            let label = model.draft?.label.trimmingCharacters(in: .whitespacesAndNewlines) ?? ""
            return label.isEmpty ? String(localized: L10n.SourceTypes.detailEyebrowNewType) : label
        }
        return model.selectedType?.label ?? ""
    }

    private var panelKey: String {
        if model.isAdding {
            let key = model.draftKey
            return key.isEmpty ? "—" : key
        }
        return model.selectedType?.key ?? ""
    }

    // MARK: Locked (plugin) detail

    @ViewBuilder
    private var lockedNote: some View {
        if let type = model.selectedType {
            VStack(alignment: .leading, spacing: PVSpacing.space7) {
                PVCallout(
                    tone: .neutral,
                    icon: .lock,
                    message: L10n.SourceTypes.lockedNotePlugin(pluginID: CatalogOrigin.pluginID(from: type.origin)),
                    compact: true
                )
                VocabularyLabeledSection(label: L10n.SourceTypes.descriptionSectionLabel) {
                    Text(type.description.isEmpty
                        ? String(localized: L10n.SourceTypes.descriptionEmptyPlaceholder)
                        : type.description)
                        .font(PVFont.body(size: PVTypeScale.bodySmall))
                        .foregroundStyle(PVColor.textSecondary)
                }
            }
            .padding(.top, PVSpacing.space7)
            .overlay(alignment: .top) {
                PVDivider()
            }
        }
    }

    // MARK: Add / editable form

    /// `panel(isLocked: false)` is only reached in `.adding` / `.editing`,
    /// where `draft` is set, so the `Binding($model.draft)` unwrap always
    /// succeeds in practice.
    @ViewBuilder
    private var form: some View {
        if let draft = Binding($model.draft) {
            VStack(alignment: .leading, spacing: PVSpacing.space7) {
                PVField(label: L10n.SourceTypes.formLabel, error: model.formError, required: true) {
                    PVInput(
                        text: draft.label,
                        prompt: model.isAdding ? L10n.SourceTypes.formLabelPlaceholder : nil,
                        isInvalid: model.formError != nil
                    )
                    .accessibilityIdentifier("sourceTypes.form.label")
                }
                PVField(
                    label: L10n.SourceTypes.formIcon,
                    hint: L10n.SourceTypes.formIconHint,
                    required: true
                ) {
                    iconFieldButton(selection: draft.iconKey)
                }
                PVField(label: L10n.SourceTypes.formDescription, hint: L10n.SourceTypes.formDescriptionHint) {
                    PVInput(
                        text: draft.description,
                        prompt: model.isAdding ? L10n.SourceTypes.formDescriptionPlaceholder : nil
                    )
                    .accessibilityIdentifier("sourceTypes.form.description")
                }
                VocabularyFormActions(
                    primaryLabel: primaryLabel,
                    secondaryLabel: secondaryLabel,
                    isSaving: model.isSaving,
                    canSubmit: model.canSubmit,
                    isSecondaryDisabled: model.isSaving || (!model.isAdding && !model.isDirty),
                    identifierPrefix: "sourceTypes",
                    onPrimary: { Task { await model.submit() } },
                    onSecondary: { secondaryAction() }
                )
            }
            .padding(.top, PVSpacing.space7)
            .overlay(alignment: .top) {
                PVDivider()
            }
        }
    }

    private func iconFieldButton(selection: Binding<String>) -> some View {
        let key = PVEvidenceIconKey(catalogKey: selection.wrappedValue)
        return Button {
            iconPickerOpen = true
        } label: {
            HStack(spacing: PVSpacing.space5) {
                ZStack {
                    RoundedRectangle(cornerRadius: PVRadius.sm, style: .continuous)
                        .fill(PVColor.surfaceSunken)
                        .overlay(
                            RoundedRectangle(cornerRadius: PVRadius.sm, style: .continuous)
                                .stroke(PVColor.borderSubtle, lineWidth: 1)
                        )
                    PVEvidenceIcon(key, size: 28, decorative: true)
                }
                .frame(width: 44, height: 44)
                VStack(alignment: .leading, spacing: 1) {
                    Text(key.typePickerTitle)
                        .font(PVFont.body(size: PVTypeScale.bodySmall))
                        .foregroundStyle(PVColor.textPrimary)
                        .lineLimit(1)
                    Text(key.rawValue)
                        .font(PVFont.mono(size: PVTypeScale.micro))
                        .foregroundStyle(PVColor.textFaint)
                        .lineLimit(1)
                        .truncationMode(.middle)
                }
                .frame(maxWidth: .infinity, alignment: .leading)
                Text(L10n.SourceTypes.formIconChange)
                    .font(PVFont.body(size: PVTypeScale.micro))
                    .foregroundStyle(PVColor.textLink)
            }
            .padding(.horizontal, PVSpacing.space5)
            .padding(.vertical, PVSpacing.space4)
            .background(
                RoundedRectangle(cornerRadius: PVRadius.sm, style: .continuous)
                    .fill(PVColor.surfaceCard)
            )
            .overlay(
                RoundedRectangle(cornerRadius: PVRadius.sm, style: .continuous)
                    .stroke(PVColor.borderDefault, lineWidth: 1)
            )
            .pvInsetShadow(cornerRadius: PVRadius.sm)
        }
        .buttonStyle(.plain)
        .accessibilityLabel(
            L10n.SourceTypes.formIconChangeAccessibility(
                name: String(localized: key.typePickerTitle)
            )
        )
        .accessibilityIdentifier("sourceTypes.form.icon")
    }

    private var primaryLabel: LocalizedStringResource {
        if model.isSaving { return L10n.SourceTypes.saveSaving }
        return model.isAdding ? L10n.SourceTypes.addType : L10n.SourceTypes.saveChanges
    }

    private var secondaryLabel: LocalizedStringResource {
        model.isAdding ? L10n.SourceTypes.cancel : L10n.SourceTypes.revert
    }

    private func secondaryAction() {
        if model.isAdding {
            model.cancelAdd()
        } else {
            model.revertEdit()
        }
    }

    /// Suggestions are deliberately not part of create (S2-03 T-15) — the
    /// callout says so, and the save lands on the detail that can assign.
    private var addSuggestionsNote: some View {
        PVCallout(
            tone: .neutral,
            icon: .info,
            message: String(localized: L10n.SourceTypes.addSuggestionsNote),
            compact: true
        )
        .padding(.top, PVSpacing.space7)
        .overlay(alignment: .top) {
            PVDivider()
        }
    }

    // MARK: Suggested fields

    private func suggestedFields(isLocked: Bool) -> some View {
        VStack(alignment: .leading, spacing: PVSpacing.space6) {
            HStack(alignment: .firstTextBaseline, spacing: PVSpacing.space5) {
                Text(L10n.SourceTypes.suggestedSectionLabel)
                    .pvMicroCaps()
                    .foregroundStyle(PVColor.textMuted)
                    .frame(maxWidth: .infinity, alignment: .leading)
                Text(L10n.SourceTypes.assignedCount(count: model.suggestions.count))
                    .font(PVFont.mono(size: PVTypeScale.micro))
                    .foregroundStyle(PVColor.textFaint)
                    .accessibilityIdentifier("sourceTypes.suggested.count")
            }
            if !isLocked {
                Text(L10n.SourceTypes.suggestedHint)
                    .font(PVFont.body(size: PVTypeScale.micro, italic: true))
                    .foregroundStyle(PVColor.textMuted)
            }
            if let suggestionError = model.suggestionError {
                PVCallout(tone: .danger, message: suggestionError)
            }
            if model.suggestions.isEmpty {
                emptySuggestions
            } else {
                suggestionRows(isLocked: isLocked)
            }
            if !isLocked {
                assignControl
            }
        }
        .padding(.top, PVSpacing.space7)
        .overlay(alignment: .top) {
            PVDivider()
        }
        .accessibilityIdentifier("sourceTypes.suggested")
    }

    @ViewBuilder
    private var emptySuggestions: some View {
        if model.isLoadingSuggestions {
            ProgressView()
                .tint(PVColor.accent)
                .frame(maxWidth: .infinity)
                .padding(PVSpacing.space7)
        } else {
            Text(L10n.SourceTypes.noAssignedBody)
                .font(PVFont.body(size: PVTypeScale.bodySmall))
                .foregroundStyle(PVColor.textSecondary)
                .frame(maxWidth: .infinity, alignment: .leading)
                .padding(PVSpacing.space6)
                .background(
                    RoundedRectangle(cornerRadius: PVRadius.sm, style: .continuous)
                        .fill(PVColor.surfaceSunken)
                )
                .overlay(
                    RoundedRectangle(cornerRadius: PVRadius.sm, style: .continuous)
                        .strokeBorder(PVColor.borderDefault, style: StrokeStyle(lineWidth: 1, dash: [4, 3]))
                )
                .accessibilityIdentifier("sourceTypes.suggested.empty")
        }
    }

    /// The ordinal column makes `sort_order` legible without adding a
    /// reorder control the board deliberately left out.
    private func suggestionRows(isLocked: Bool) -> some View {
        VStack(spacing: 0) {
            ForEach(Array(model.suggestions.enumerated()), id: \.element.id) { index, suggestion in
                HStack(spacing: PVSpacing.space5) {
                    Text("\(index + 1)")
                        .font(PVFont.mono(size: PVTypeScale.micro))
                        .foregroundStyle(PVColor.textFaint)
                        .frame(width: 18, alignment: .leading)
                    VStack(alignment: .leading, spacing: 1) {
                        Text(suggestion.field.label)
                            .font(PVFont.body(size: PVTypeScale.bodySmall))
                            .foregroundStyle(PVColor.textPrimary)
                            .lineLimit(1)
                        Text(suggestion.field.key)
                            .font(PVFont.mono(size: PVTypeScale.micro))
                            .foregroundStyle(PVColor.textFaint)
                            .lineLimit(1)
                            .truncationMode(.middle)
                    }
                    .frame(maxWidth: .infinity, alignment: .leading)
                    CatalogFieldDataTypeBadge(dataType: suggestion.field.dataType)
                    if !isLocked {
                        PVIconButton(
                            .dismiss,
                            label: L10n.SourceTypes.removeSuggestion(label: suggestion.field.label),
                            size: .sm
                        ) {
                            Task { await model.removeSuggestion(fieldID: suggestion.field.id) }
                        }
                        .disabled(model.removingFieldID != nil)
                        .accessibilityIdentifier("sourceTypes.suggested.remove.\(suggestion.field.id)")
                    }
                }
                .padding(.horizontal, PVSpacing.space5)
                .padding(.vertical, PVSpacing.space4)
                .accessibilityElement(children: .combine)
                if index < model.suggestions.count - 1 {
                    PVDivider()
                }
            }
        }
        .background(PVColor.surfaceCard)
        .overlay(
            RoundedRectangle(cornerRadius: PVRadius.sm, style: .continuous)
                .strokeBorder(PVColor.borderSubtle, lineWidth: 1)
        )
        .clipShape(RoundedRectangle(cornerRadius: PVRadius.sm, style: .continuous))
    }

    /// The pool is the Source fields vocabulary and nothing else — no inline
    /// field creation (S2-03 T-9); the hint says where to add one instead.
    ///
    /// A combo box rather than a popup menu: the vocabulary grows without
    /// bound, and typing a label or a key beats scanning a long menu. The
    /// assign action sits to its right as an icon button, its tooltip
    /// carrying the reason when it is disabled — the same disabled-not-hidden
    /// pattern the delete affordance uses.
    private var assignControl: some View {
        // Resolved once per render, not once per row — the row builder runs
        // for every visible option and only needs the badge's data type.
        let dataTypes = Dictionary(uniqueKeysWithValues: model.assignPool.map { ($0.id, $0.dataType) })
        return VStack(alignment: .leading, spacing: PVSpacing.space4) {
            HStack(spacing: PVSpacing.space5) {
                PVComboBox(
                    selection: $model.assignPick,
                    options: poolOptions,
                    placeholder: L10n.SourceTypes.assignPlaceholder,
                    emptyLabel: L10n.SourceTypes.assignNoMatch,
                    maxListHeight: Self.assignListMaxHeight,
                    label: L10n.SourceTypes.assignFieldLabel,
                    accessibilityIdentifierPrefix: "sourceTypes.suggested.pool"
                ) { option, query in
                    poolRow(option, query: query, dataTypes: dataTypes)
                }
                .disabled(model.assignPool.isEmpty || model.isAssigning)
                PVIconButton(
                    .plus,
                    label: model.assignTooltip,
                    accessibilityLabel: model.assignAccessibilityLabel
                ) {
                    Task { await model.assignPickedField() }
                }
                .disabled(!model.canAssign)
                .accessibilityIdentifier("sourceTypes.suggested.assign")
            }
            Text(model.assignPool.isEmpty ? L10n.SourceTypes.poolHintEmpty : L10n.SourceTypes.poolHint)
                .font(PVFont.body(size: PVTypeScale.micro, italic: true))
                .foregroundStyle(PVColor.textMuted)
        }
    }

    /// The board's `max-height` for the assign list.
    private static let assignListMaxHeight: CGFloat = 220

    /// The pool row: field label over its mono key, with the data type badge
    /// trailing. Label and key both carry the match highlight, so typing
    /// either one shows you why the row matched.
    private func poolRow(_ option: PVComboBoxOption, query: String, dataTypes: [String: String]) -> some View {
        HStack(spacing: PVSpacing.space5) {
            VStack(alignment: .leading, spacing: 1) {
                Text(PVComboBoxHighlight.attributed(option.label, query: query))
                    .font(PVFont.body(size: PVTypeScale.bodySmall))
                    .foregroundStyle(PVColor.textPrimary)
                    .lineLimit(1)
                Text(PVComboBoxHighlight.attributed(option.subtext, query: query))
                    .font(PVFont.mono(size: PVTypeScale.micro))
                    .foregroundStyle(PVColor.textMuted)
                    .lineLimit(1)
                    .truncationMode(.middle)
            }
            .frame(maxWidth: .infinity, alignment: .leading)
            if let dataType = dataTypes[option.value] {
                CatalogFieldDataTypeBadge(dataType: dataType)
            }
        }
    }

    /// Label and key are both searchable, so the key rides along as the
    /// option's subtext rather than being flattened into the label.
    private var poolOptions: [PVComboBoxOption] {
        model.assignPool.map { PVComboBoxOption(value: $0.id, label: $0.label, subtext: $0.key) }
    }
}

/// Modal grid of closed `type_*` marks — S2-03 board §08 / design pane
/// "Choose an icon". Selection updates the draft immediately; Done dismisses.
private struct SourceTypeIconPickerSheet: View {
    @Binding var selection: String
    let onDone: () -> Void

    private let columns = [GridItem(.adaptive(minimum: 108), spacing: PVSpacing.space4)]

    var body: some View {
        VStack(spacing: 0) {
            VStack(alignment: .leading, spacing: PVSpacing.space7) {
                VStack(alignment: .leading, spacing: PVSpacing.space2) {
                    Text(L10n.SourceTypes.iconPickerTitle)
                        .font(PVFont.display(size: PVTypeScale.h3))
                        .foregroundStyle(PVColor.textDisplay)
                    Text(L10n.SourceTypes.iconPickerSubtitle)
                        .font(PVFont.body(size: PVTypeScale.bodySmall))
                        .foregroundStyle(PVColor.textSecondary)
                        .lineSpacing((PVLineHeight.relaxed - 1) * PVTypeScale.bodySmall)
                        .fixedSize(horizontal: false, vertical: true)
                }
                LazyVGrid(columns: columns, spacing: PVSpacing.space4) {
                    ForEach(PVEvidenceIconKey.typeKeys, id: \.rawValue) { key in
                        iconCell(key)
                    }
                }
                .accessibilityElement(children: .contain)
                .accessibilityLabel(Text(L10n.SourceTypes.iconPickerGroupLabel))
            }
            .padding(.horizontal, PVSpacing.space7)
            .padding(.vertical, PVSpacing.space9)

            HStack(alignment: .center, spacing: PVSpacing.space5) {
                Text(selectedKey.typeMetaphor)
                    .font(PVFont.body(size: PVTypeScale.micro, italic: true))
                    .foregroundStyle(PVColor.textMuted)
                    .frame(maxWidth: .infinity, alignment: .leading)
                    .fixedSize(horizontal: false, vertical: true)
                Button(String(localized: L10n.SourceTypes.iconPickerDone)) {
                    onDone()
                }
                .buttonStyle(.pv(.ghost, size: .lg))
                .keyboardShortcut(.defaultAction)
                .accessibilityIdentifier("sourceTypes.iconPicker.done")
            }
            .padding(.horizontal, PVSpacing.space8)
            .padding(.vertical, PVSpacing.space8)
            .overlay(alignment: .top) {
                PVDivider()
            }
        }
        .frame(width: 660)
        .background(PVColor.surfaceCard)
        .accessibilityIdentifier("sourceTypes.iconPicker")
    }

    private var selectedKey: PVEvidenceIconKey {
        PVEvidenceIconKey(catalogKey: selection)
    }

    private func iconCell(_ key: PVEvidenceIconKey) -> some View {
        let selected = selection == key.rawValue
        return Button {
            selection = key.rawValue
        } label: {
            VStack(spacing: PVSpacing.space4) {
                PVEvidenceIcon(key, size: 40, decorative: true)
                Text(key.typePickerTitle)
                    .font(PVFont.body(size: PVTypeScale.micro))
                    .multilineTextAlignment(.center)
                    .lineLimit(2)
                    .fixedSize(horizontal: false, vertical: true)
            }
            .frame(maxWidth: .infinity)
            .padding(.horizontal, PVSpacing.space4)
            .padding(.vertical, PVSpacing.space5)
            .foregroundStyle(selected ? PVColor.accent : PVColor.textSecondary)
            .background(
                RoundedRectangle(cornerRadius: PVRadius.sm, style: .continuous)
                    .fill(selected ? PVColor.surfaceActive : PVColor.surfaceCard)
            )
            .overlay(
                RoundedRectangle(cornerRadius: PVRadius.sm, style: .continuous)
                    .stroke(selected ? PVColor.accent : PVColor.borderSubtle, lineWidth: 1)
            )
            .modifier(SelectedIconShadow(selected: selected))
        }
        .buttonStyle(.plain)
        .accessibilityLabel(Text(key.typePickerTitle))
        .accessibilityAddTraits(selected ? [.isSelected, .isButton] : .isButton)
        .accessibilityIdentifier("sourceTypes.iconPicker.\(key.rawValue)")
    }
}

private struct SelectedIconShadow: ViewModifier {
    let selected: Bool

    func body(content: Content) -> some View {
        if selected {
            content.pvShadow(PVElevation.sm)
        } else {
            content
        }
    }
}
