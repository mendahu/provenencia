import SwiftUI

/// Sticky identity header: thumbnail, title edit, type chip.
struct SourcePageIdentityHeader: View {
    @Bindable var model: SourcePageModel
    /// Local draft so title keystrokes don't invalidate the whole Source page
    /// observation graph on every character.
    @State private var titleDraft = ""
    @State private var coverMenu = PVContextMenuState()
    @State private var coverKeyboard = PVContextMenuKeyboard.inactive

    private var coverMenuItemCount: Int {
        model.source?.coverMode == "type_icon" ? 0 : 1
    }

    var body: some View {
        VStack(alignment: .leading, spacing: PVSpacing.space5) {
            if model.workspace != nil {
                HStack(alignment: .top, spacing: PVSpacing.space7) {
                    coverThumbnail
                    titleCluster
                        .frame(maxWidth: .infinity, alignment: .leading)
                }
                // Menu on the row (not the 72pt thumb) so overflow hits register.
                .pvContextMenu(
                    $coverMenu,
                    keyboard: coverMenuItemCount > 0 ? $coverKeyboard : nil
                ) {
                    coverMenuPanel
                }
                .frame(maxWidth: PVSpacing.widthContentMax, alignment: .leading)
            }
        }
        .padding(.top, PVSpacing.space5)
        .padding(.horizontal, PVSpacing.gutterPage)
        .padding(.bottom, PVSpacing.space6)
        .frame(maxWidth: .infinity, alignment: .leading)
        .background(PVColor.surfaceCard)
        .overlay(alignment: .bottom) {
            Rectangle()
                .fill(PVColor.borderSubtle)
                .frame(height: 1)
        }
    }

    private var coverThumbnail: some View {
        CachedThumbnail(
            projectDir: model.pageProjectDir,
            relPath: identityCover.relPath,
            mediaType: identityCover.mediaType,
            originalFilename: identityCover.originalFilename,
            typeIconKey: identityCover.typeIconKey,
            size: 72
        )
        .accessibilityHint(String(localized: L10n.Sources.thumbnailMenuHint))
        .accessibilityIdentifier("sources.page.cover")
        .pvContextMenuTrigger($coverMenu)
        .onChange(of: coverMenu.isPresented) { _, open in
            if open {
                coverKeyboard = PVContextMenuKeyboard(
                    itemCount: coverMenuItemCount,
                    activeIndex: -1
                )
            }
        }
        .help(String(localized: L10n.Sources.thumbnailMenuHint))
    }

    private var coverMenuPanel: some View {
        PVContextMenuPanel(
            title: L10n.Sources.thumbnailMenuTitle,
            accessibilityIdentifier: "sources.page.cover.menu"
        ) {
            if model.source?.coverMode == "type_icon" {
                PVContextMenuItem(L10n.Sources.thumbnailDefaultInUse, isEnabled: false)
            } else {
                PVContextMenuItem(
                    L10n.Sources.thumbnailRevertToDefault,
                    index: 0,
                    accessibilityIdentifier: "sources.page.cover.revert"
                ) {
                    Task { await model.artifacts.revertCoverToTypeIcon() }
                }
            }
        }
    }

    private var identityCover: (
        relPath: String,
        mediaType: String,
        originalFilename: String,
        typeIconKey: String?
    ) {
        let typeIcon = model.identity.types
            .first { $0.id == model.identity.sourceTypeID }?
            .iconKey
        // Source cover is type icon or an explicit raster pin — never MIME.
        if let source = model.source,
           source.coverMode == "artifact",
           !source.thumbnailRelPath.isEmpty
        {
            return (source.thumbnailRelPath, "", "", typeIcon)
        }
        return ("", "", "", typeIcon)
    }

    private var titleCluster: some View {
        VStack(alignment: .leading, spacing: PVSpacing.space4) {
            PVInlineEdit(
                isEditing: model.identity.editingTitle,
                isSaving: model.identity.isSaving,
                error: model.identity.titleError,
                saveLabel: L10n.Sources.saveAction,
                cancelLabel: L10n.Sources.cancelEdit,
                editLabel: L10n.Sources.editTitle,
                saveDisabled: titleDraft.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty,
                expandsContent: false,
                axis: .horizontal,
                accessibilityIdentifierPrefix: "sources.page.title",
                onEdit: {
                    titleDraft = model.identity.title
                    model.identity.beginEditTitle()
                },
                onSave: {
                    model.identity.titleDraft = titleDraft
                    Task { await model.identity.saveTitle() }
                },
                onCancel: {
                    model.identity.cancelEditTitle()
                    titleDraft = ""
                }
            ) {
                // Transparent chrome matching `PVTextArea` display so edit
                // mode does not shift the glyphs.
                Text(model.identity.title)
                    .font(PVFont.display(size: PVTypeScale.h1, weight: PVFontWeight.semibold))
                    .foregroundStyle(PVColor.textDisplay)
                    .lineLimit(4)
                    .fixedSize(horizontal: false, vertical: true)
                    .padding(.horizontal, PVInputChrome.horizontalInset)
                    .padding(.vertical, PVTextArea.Typography.display.verticalPadding)
                    .overlay(
                        RoundedRectangle(cornerRadius: PVRadius.sm, style: .continuous)
                            .stroke(Color.clear, lineWidth: 1)
                    )
                    .accessibilityIdentifier("sources.page.title")
            } editor: {
                PVTextArea(
                    text: $titleDraft,
                    lineLimit: 1...4,
                    prompt: L10n.Sources.formTitle,
                    typography: .display,
                    isInvalid: model.identity.titleError != nil,
                    activateOnAppear: true
                )
                .accessibilityIdentifier("sources.page.title")
                .onChange(of: titleDraft) { _, _ in
                    if model.identity.titleError != nil {
                        model.identity.titleError = nil
                    }
                }
                .onSubmit {
                    model.identity.titleDraft = titleDraft
                    Task { await model.identity.saveTitle() }
                }
                .disabled(model.identity.isSaving)
                .frame(minWidth: 240, maxWidth: 720, alignment: .leading)
            }

            metaRow
        }
    }

    private var metaRow: some View {
        HStack(alignment: .center, spacing: PVSpacing.space5) {
            if let ref = model.source?.ref {
                Text(ref)
                    .font(PVFont.mono(size: PVTypeScale.caption))
                    .foregroundStyle(PVColor.textMuted)
                    .accessibilityIdentifier("sources.page.ref")
            }

            Text(verbatim: "·")
                .foregroundStyle(PVColor.borderDefault)

            if model.identity.editingType {
                VStack(alignment: .leading, spacing: PVSpacing.space2) {
                    HStack(spacing: PVSpacing.space3) {
                        PVComboBox(
                            selection: $model.identity.typeDraftID,
                            options: model.identity.typeComboOptions,
                            placeholder: L10n.Sources.typePlaceholder,
                            emptyLabel: L10n.Sources.typeNoMatch,
                            label: L10n.Sources.pageFormType,
                            accessibilityIdentifierPrefix: "sources.page.type",
                            activateOnAppear: true
                        ) { option, query in
                            HStack(spacing: PVSpacing.space3) {
                                if let type = model.identity.types.first(where: { $0.id == option.value }) {
                                    PVEvidenceIcon(
                                        PVEvidenceIconKey(catalogKey: type.iconKey),
                                        size: .row,
                                        decorative: true
                                    )
                                }
                                PVComboBoxPlainRow(option: option, query: query)
                            }
                        }
                        .frame(width: 210)
                        .disabled(model.identity.isSaving)
                        .onChange(of: model.identity.typeDraftID) { _, newValue in
                            guard !newValue.isEmpty, newValue != model.identity.sourceTypeID else { return }
                            Task { await model.identity.saveType() }
                        }

                        PVIconButton(.dismiss, label: L10n.Sources.cancelEdit, size: .sm) {
                            model.identity.cancelEditType()
                        }
                        .disabled(model.identity.isSaving)
                        .accessibilityIdentifier("sources.page.type.cancel")
                    }

                    if let typeError = model.identity.typeError {
                        Text(typeError)
                            .font(PVFont.body(size: PVTypeScale.caption))
                            .foregroundStyle(PVColor.danger)
                    }
                }
            } else {
                HStack(spacing: PVSpacing.space3) {
                    CatalogSourceTypePill(
                        label: model.identity.typeLabel,
                        iconKey: model.identity.types
                            .first { $0.id == model.identity.sourceTypeID }?
                            .iconKey
                    )
                        .accessibilityIdentifier("sources.page.type")

                    PVIconButton(.penLine, label: L10n.Sources.editType, size: .sm) {
                        model.identity.beginEditType()
                    }
                    .accessibilityIdentifier("sources.page.type.edit")
                }
            }
        }
    }
}
