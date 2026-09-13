import SwiftUI

/// Sticky identity header: breadcrumbs, thumbnail, title edit, type chip.
struct SourcePageIdentityHeader: View {
    @Bindable var model: SourcePageModel
    let onBackToList: () -> Void
    /// Local draft so title keystrokes don't invalidate the whole Source page
    /// observation graph on every character.
    @State private var titleDraft = ""

    var body: some View {
        VStack(alignment: .leading, spacing: PVSpacing.space5) {
            breadcrumbs
            if model.workspace != nil {
                HStack(alignment: .top, spacing: PVSpacing.space7) {
                    CachedThumbnail(
                        projectDir: model.pageProjectDir,
                        relPath: identityCover.relPath,
                        mediaType: identityCover.mediaType,
                        originalFilename: identityCover.originalFilename,
                        typeIconKey: identityCover.typeIconKey,
                        size: 72
                    )
                    .accessibilityHint(L10n.Sources.thumbnailMenuHint)
                    .contextMenu {
                        Section {
                            Text(L10n.Sources.thumbnailMenuTitle)
                        }
                        if model.source?.coverMode == "type_icon" {
                            Button(String(localized: L10n.Sources.thumbnailDefaultInUse)) {}
                                .disabled(true)
                        } else {
                            Button(String(localized: L10n.Sources.thumbnailRevertToDefault)) {
                                Task { await model.artifacts.revertCoverToTypeIcon() }
                            }
                        }
                    }
                    .accessibilityIdentifier("sources.page.cover")
                    titleCluster
                        .frame(maxWidth: .infinity, alignment: .leading)
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

    private var identityCover: (
        relPath: String,
        mediaType: String,
        originalFilename: String,
        typeIconKey: String?
    ) {
        let typeIcon = model.identity.types
            .first { $0.id == model.identity.sourceTypeID }?
            .iconKey
        // Prefer persisted / enriched Source cover; fall through only when
        // workspace source has not been enriched yet (load race).
        if let source = model.source {
            if source.coverMode == "type_icon"
                || !source.thumbnailRelPath.isEmpty
                || !source.thumbnailMediaType.isEmpty
                || !source.thumbnailOriginalFilename.isEmpty
            {
                return (
                    source.thumbnailRelPath,
                    source.thumbnailMediaType,
                    source.thumbnailOriginalFilename,
                    typeIcon
                )
            }
            if source.coverMode == "artifact",
               let art = model.artifacts.items.first(where: { $0.id == source.primaryArtifactID })
            {
                if !art.thumbnailRelPath.isEmpty {
                    return (art.thumbnailRelPath, "", "", typeIcon)
                }
                if let file = art.file {
                    return ("", file.mediaType, file.originalFilename, typeIcon)
                }
            }
        }
        return ("", "", "", typeIcon)
    }

    private var breadcrumbs: some View {
        PVBreadcrumbs(items: [
            PVBreadcrumbItem(
                id: "sources",
                label: String(localized: L10n.Sources.breadcrumbSources),
                action: onBackToList
            ),
            PVBreadcrumbItem(
                id: "ref",
                label: model.source?.ref ?? "…",
                action: nil
            ),
        ])
        .accessibilityIdentifier("sources.page.breadcrumbs")
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

            Text("·")
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
