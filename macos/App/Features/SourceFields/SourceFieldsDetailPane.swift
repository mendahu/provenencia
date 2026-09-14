import SwiftUI

/// The right pane of `SourceFieldsView`: the add-field form, a selected
/// field's detail (editable for `user` / `provenencia`, locked read-only for
/// `plugin:…`), or an empty prompt when nothing is selected — S2-02 §3.2/3.3.
struct SourceFieldsDetailPane: View {
    @Environment(WorkspaceNavigation.self) private var navigation
    @Bindable var model: SourceFieldsModel

    var body: some View {
        ScrollView {
            content
        }
        .background(PVColor.surfaceCard)
        .accessibilityIdentifier("sourceFields.detail")
    }

    @ViewBuilder
    private var content: some View {
        if model.isAdding {
            panel(isLocked: false)
        } else if model.selectedField != nil {
            panel(isLocked: model.isSelectedFieldLocked)
        } else {
            PVEmptyState(
                icon: .tag,
                title: L10n.SourceFields.panelEmptyTitle,
                message: String(localized: L10n.SourceFields.panelEmptyBody),
                compact: true
            )
            .padding(PVSpacing.space9)
        }
    }

    private func panel(isLocked: Bool) -> some View {
        VStack(alignment: .leading, spacing: PVSpacing.space7) {
            panelHeader
            if isLocked {
                lockedBody
            } else {
                form
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
            eyebrow: model.isAdding ? L10n.SourceFields.detailEyebrowNewField : L10n.SourceFields.detailEyebrowField,
            title: panelTitle,
            origin: model.isAdding ? CatalogOrigin.user : (model.selectedField?.origin ?? CatalogOrigin.user),
            keyText: panelKey,
            keyHint: model.isAdding ? L10n.SourceFields.keyHintAdd : L10n.SourceFields.keyHintEdit,
            showsDelete: model.showsDelete,
            canDelete: model.canDeleteSelectedField,
            deleteTooltip: model.deleteTooltip,
            identifierPrefix: "sourceFields",
            onDelete: { model.askDelete() }
        )
    }

    private var panelTitle: String {
        if model.isAdding {
            let label = model.draft?.label.trimmingCharacters(in: .whitespacesAndNewlines) ?? ""
            return label.isEmpty ? String(localized: L10n.SourceFields.detailEyebrowNewField) : label
        }
        return model.selectedField?.label ?? ""
    }

    private var panelKey: String {
        if model.isAdding {
            let key = model.draftKey
            return key.isEmpty ? "—" : key
        }
        return model.selectedField?.key ?? ""
    }

    // MARK: Locked (plugin) detail

    @ViewBuilder
    private var lockedBody: some View {
        if let field = model.selectedField {
            VStack(alignment: .leading, spacing: PVSpacing.space7) {
                PVCallout(tone: .neutral, icon: .lock, message: lockedNote(for: field), compact: true)
                VocabularyLabeledSection(label: L10n.SourceFields.dataTypeSectionLabel) {
                    Text(CatalogFieldDataType.label(for: field.dataType))
                        .font(PVFont.body(size: PVTypeScale.bodySmall))
                        .foregroundStyle(PVColor.textPrimary)
                }
                VocabularyLabeledSection(label: L10n.SourceFields.descriptionSectionLabel) {
                    Text(field.description.isEmpty ? String(localized: L10n.SourceFields.descriptionEmptyPlaceholder) : field.description)
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

    private func lockedNote(for field: CatalogMetadataField) -> String {
        L10n.SourceFields.lockedNotePlugin(pluginID: CatalogOrigin.pluginID(from: field.origin))
    }

    // MARK: Add / editable form

    private var dataTypeOptions: [PVSelectOption] {
        [
            PVSelectOption(value: CatalogFieldDataType.text, label: String(localized: L10n.SourceFields.dataTypeText)),
            PVSelectOption(value: CatalogFieldDataType.date, label: String(localized: L10n.SourceFields.dataTypeDate)),
        ]
    }

    /// `panel(isLocked: false)` is only reached in `.adding` / `.editing`,
    /// where `draft` is set, so the `Binding($model.draft)` unwrap always
    /// succeeds in practice.
    @ViewBuilder
    private var form: some View {
        if let draft = Binding($model.draft) {
            VStack(alignment: .leading, spacing: PVSpacing.space7) {
                PVField(label: L10n.SourceFields.formLabel, error: model.formError, required: true) {
                    PVInput(
                        text: draft.label,
                        prompt: model.isAdding ? L10n.SourceFields.formLabelPlaceholder : nil,
                        isInvalid: model.formError != nil
                    )
                    .accessibilityIdentifier("sourceFields.form.label")
                }
                PVField(label: L10n.SourceFields.formDataType, hint: model.isAdding ? L10n.SourceFields.formDataTypeHint : L10n.SourceFields.formDataTypeImmutableHint) {
                    if model.isAdding {
                        PVSelect(selection: draft.dataType, options: dataTypeOptions)
                            .accessibilityIdentifier("sourceFields.form.dataType")
                    } else {
                        Text(CatalogFieldDataType.label(for: draft.wrappedValue.dataType))
                            .font(PVFont.body(size: PVTypeScale.body))
                            .foregroundStyle(PVColor.textPrimary)
                            .accessibilityIdentifier("sourceFields.form.dataType.readonly")
                    }
                }
                PVField(label: L10n.SourceFields.formDescription, hint: L10n.SourceFields.formDescriptionHint) {
                    PVInput(text: draft.description, prompt: model.isAdding ? L10n.SourceFields.formDescriptionPlaceholder : nil)
                        .accessibilityIdentifier("sourceFields.form.description")
                }
                VocabularyFormActions(
                    primaryLabel: primaryLabel,
                    secondaryLabel: secondaryLabel,
                    isSaving: model.isSaving,
                    canSubmit: model.canSubmit,
                    isSecondaryDisabled: model.isSaving || (!model.isAdding && !model.isDirty),
                    identifierPrefix: "sourceFields",
                    onPrimary: {
                        Task {
                            if let location = await model.submit() {
                                navigation.go(to: location)
                            }
                        }
                    },
                    onSecondary: { secondaryAction() }
                )
            }
            .padding(.top, PVSpacing.space7)
            .overlay(alignment: .top) {
                PVDivider()
            }
        }
    }

    private var primaryLabel: LocalizedStringResource {
        if model.isSaving { return L10n.SourceFields.saveSaving }
        return model.isAdding ? L10n.SourceFields.addField : L10n.SourceFields.saveChanges
    }

    private var secondaryLabel: LocalizedStringResource {
        model.isAdding ? L10n.SourceFields.cancel : L10n.SourceFields.revert
    }

    private func secondaryAction() {
        if model.isAdding {
            model.cancelAdd()
        } else {
            model.revertEdit()
        }
    }
}
