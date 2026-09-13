import SwiftUI

/// Artifacts accordion: list, expand/collapse detail, and the Add dialog form.
struct SourcePageArtifactsView: View {
    @Bindable var model: SourcePageModel

    var body: some View {
        VStack(alignment: .leading, spacing: PVSpacing.space6) {
            PVSectionHeader(
                title: L10n.Sources.artifactsHeading,
                meta: model.artifacts.items.isEmpty
                    ? nil
                    : L10n.Sources.artifactsCount(model.artifacts.items.count),
                actions: {
                    PVButton(L10n.Sources.addArtifact, variant: .primary, size: .sm, icon: .plus) {
                        model.artifacts.openAdd()
                    }
                    .accessibilityIdentifier("sources.page.addArtifact")
                }
            )

            if model.artifacts.items.isEmpty {
                PVEmptyState(
                    icon: .photo,
                    title: L10n.Sources.artifactsEmptyTitle,
                    message: String(localized: L10n.Sources.artifactsEmptyMessage),
                    compact: true
                ) {
                    PVButton(L10n.Sources.addArtifact, variant: .primary, size: .sm, icon: .plus) {
                        model.artifacts.openAdd()
                    }
                    .accessibilityIdentifier("sources.page.artifacts.empty.add")
                }
                .accessibilityIdentifier("sources.page.artifacts.empty")
            } else {
                PVCard(elevated: true) {
                    VStack(spacing: 0) {
                        ForEach(model.artifacts.items, id: \.id) { art in
                            artifactRow(art)
                            if art.id != model.artifacts.items.last?.id {
                                PVDivider()
                            }
                        }
                    }
                }
                .accessibilityIdentifier("sources.page.artifacts.list")
            }
        }
        .padding(.top, PVSpacing.space11 - PVSpacing.space9)
    }

    /// Label / description / optional file form for the Add Artifact dialog.
    var addForm: some View {
        VStack(alignment: .leading, spacing: PVSpacing.space6) {
            PVField(
                label: L10n.Sources.artifactLabel,
                hint: L10n.Sources.artifactLabelHint,
                error: model.artifacts.draftLabelError,
                required: true
            ) {
                PVInput(
                    text: $model.artifacts.draft.label,
                    isInvalid: model.artifacts.draftLabelError != nil
                )
                .accessibilityIdentifier("sources.page.addArtifact.label")
                .onChange(of: model.artifacts.draft.label) { _, _ in
                    if model.artifacts.draftLabelError != nil {
                        model.artifacts.draftLabelError = nil
                    }
                }
            }
            PVField(
                label: L10n.Sources.artifactDescription,
                hint: L10n.Sources.formDescriptionHint
            ) {
                PVInput(text: $model.artifacts.draft.description)
                    .accessibilityIdentifier("sources.page.addArtifact.description")
            }
            PVField(label: L10n.Sources.optionalFile) {
                HStack(spacing: PVSpacing.space4) {
                    if let name = model.artifacts.draft.fileName {
                        Text(name)
                            .font(PVFont.mono(size: PVTypeScale.caption))
                            .foregroundStyle(PVColor.textPrimary)
                            .lineLimit(1)
                        PVButton(L10n.Sources.clearFile, variant: .ghost, size: .sm) {
                            model.artifacts.clearFile()
                        }
                    } else {
                        PVButton(L10n.Sources.chooseFile, variant: .secondary, size: .sm, icon: .fileUp) {
                            model.artifacts.pickFile()
                        }
                        .accessibilityIdentifier("sources.page.addArtifact.chooseFile")
                    }
                }
            }
        }
    }

    private func artifactTypeIconKey(for art: CatalogArtifact) -> String? {
        // File-bearing rows stay on raster → MIME; only fileless use type icon.
        guard art.file == nil || art.fileID.isEmpty else { return nil }
        return model.identity.types
            .first { $0.id == model.source?.sourceTypeID }?
            .iconKey
    }

    private func artifactRow(_ art: CatalogArtifact) -> some View {
        let expanded = model.artifacts.expandedIDs.contains(art.id)
        return VStack(alignment: .leading, spacing: 0) {
            Button {
                model.artifacts.toggleExpanded(art.id)
            } label: {
                HStack(spacing: PVSpacing.space6) {
                    PVIcon(.chevronForward, size: 15)
                        .foregroundStyle(PVColor.textMuted)
                        .rotationEffect(.degrees(expanded ? 90 : 0))
                        .frame(width: SourcePageLayout.artifactChevronWidth)
                    CachedThumbnail(
                        projectDir: model.pageProjectDir,
                        relPath: art.thumbnailRelPath,
                        mediaType: art.file?.mediaType ?? "",
                        originalFilename: art.file?.originalFilename ?? "",
                        typeIconKey: artifactTypeIconKey(for: art),
                        size: SourcePageLayout.artifactRowThumbnailSize
                    )
                    Text(art.label.isEmpty ? art.ref : art.label)
                        .font(PVFont.body(size: PVTypeScale.bodySmall))
                        .foregroundStyle(PVColor.textPrimary)
                        .lineLimit(1)
                        .frame(maxWidth: .infinity, alignment: .leading)
                    Text(art.ref)
                        .font(PVFont.mono(size: PVTypeScale.micro))
                        .foregroundStyle(PVColor.textMuted)
                }
                .padding(.horizontal, PVSpacing.space6)
                .padding(.vertical, PVSpacing.space5)
                .contentShape(Rectangle())
            }
            .buttonStyle(.plain)
            .accessibilityIdentifier("sources.page.artifact.\(art.id)")

            if expanded {
                artifactDetail(art)
            }
        }
    }

    private func artifactDetail(_ art: CatalogArtifact) -> some View {
        HStack(alignment: .top, spacing: PVSpacing.space8) {
            SourcePageArtifactFieldsEditor(
                artifact: art,
                isSaving: model.artifacts.savingID == art.id,
                fieldError: model.artifacts.fieldErrors[art.id],
                onSave: { label, description in
                    model.artifacts.labels[art.id] = label
                    model.artifacts.descriptions[art.id] = description
                    Task { await model.artifacts.saveFields(id: art.id) }
                },
                onCancel: {
                    model.artifacts.cancelFields(id: art.id)
                },
                onClearFieldError: {
                    if model.artifacts.fieldErrors[art.id] != nil {
                        model.artifacts.fieldErrors[art.id] = nil
                    }
                }
            )
            .frame(minWidth: 280, maxWidth: .infinity, alignment: .leading)

            artifactPrimaryFileColumn(art)
                .frame(minWidth: 300, maxWidth: .infinity, alignment: .leading)
        }
        .padding(.top, PVSpacing.space7)
        .padding(.trailing, PVSpacing.space8)
        .padding(.bottom, PVSpacing.space8)
        .padding(.leading, SourcePageLayout.artifactDetailLeading)
        .frame(maxWidth: .infinity, alignment: .leading)
        .background(PVColor.surfaceSunken)
        .overlay(alignment: .top) {
            Rectangle()
                .fill(PVColor.borderSubtle)
                .frame(height: 1)
        }
    }

    private func artifactPrimaryFileColumn(_ art: CatalogArtifact) -> some View {
        VStack(alignment: .leading, spacing: PVSpacing.space4) {
            Text(L10n.Sources.primaryFileHeading)
                .font(PVFont.body(size: PVTypeScale.micro, weight: PVFontWeight.semibold))
                .tracking(PVTypeScale.micro * PVTracking.caps)
                .textCase(.uppercase)
                .foregroundStyle(PVColor.textFaint)

            if let file = art.file, !art.fileID.isEmpty {
                PVCard(cornerRadius: PVRadius.sm, padding: PVSpacing.space6) {
                    HStack(spacing: PVSpacing.space6) {
                        CachedThumbnail(
                            projectDir: model.pageProjectDir,
                            relPath: art.thumbnailRelPath,
                            mediaType: file.mediaType,
                            originalFilename: file.originalFilename,
                            size: 56
                        )
                        VStack(alignment: .leading, spacing: PVSpacing.space2) {
                            Text(file.originalFilename)
                                .font(PVFont.mono(size: PVTypeScale.caption))
                                .foregroundStyle(PVColor.textPrimary)
                                .lineLimit(2)
                            Text(fileMetaLine(file))
                                .font(PVFont.mono(size: PVTypeScale.micro))
                                .foregroundStyle(PVColor.textMuted)
                        }
                        .frame(maxWidth: .infinity, alignment: .leading)
                        PVButton(L10n.Sources.openFile, variant: .secondary, size: .sm, icon: .externalLink) {
                            model.artifacts.open(art)
                        }
                        .accessibilityIdentifier("sources.page.artifact.\(art.id).open")
                    }
                }
                .contentShape(Rectangle())
                .onTapGesture { model.artifacts.open(art) }
            } else {
                PVCallout(
                    tone: .neutral,
                    message: String(localized: L10n.Sources.filelessHint),
                    compact: true
                )
                PVButton(L10n.Sources.addFile, variant: .primary, size: .sm, icon: .fileUp) {
                    Task { await model.artifacts.addFile(to: art.id) }
                }
                .accessibilityIdentifier("sources.page.artifact.\(art.id).addFile")
            }
        }
    }

    private func fileMetaLine(_ file: CatalogFileRef) -> String {
        let size = ByteCountFormatter.string(fromByteCount: file.byteSize, countStyle: .file)
        return "\(file.mediaType) · \(size)"
    }
}

/// Expanded artifact label/description editor with local drafts.
private struct SourcePageArtifactFieldsEditor: View {
    let artifact: CatalogArtifact
    let isSaving: Bool
    let fieldError: String?
    let onSave: (_ label: String, _ description: String) -> Void
    let onCancel: () -> Void
    let onClearFieldError: () -> Void

    @State private var labelDraft = ""
    @State private var descriptionDraft = ""

    var body: some View {
        VStack(alignment: .leading, spacing: PVSpacing.space6) {
            PVField(
                label: L10n.Sources.artifactLabel,
                hint: L10n.Sources.artifactLabelHint,
                error: fieldError,
                required: true
            ) {
                PVInput(
                    text: $labelDraft,
                    size: .sm,
                    isInvalid: fieldError != nil
                )
                .onChange(of: labelDraft) { _, _ in onClearFieldError() }
                .onSubmit { onSave(labelDraft, descriptionDraft) }
            }

            PVField(
                label: L10n.Sources.artifactDescription,
                hint: L10n.Sources.artifactDescriptionHint
            ) {
                PVTextArea(
                    text: $descriptionDraft,
                    lineLimit: 3...8
                )
                .accessibilityIdentifier("sources.page.artifact.\(artifact.id).description")
                .onSubmit { onSave(labelDraft, descriptionDraft) }
            }

            HStack(spacing: PVSpacing.space4) {
                PVInlineEditActions(
                    isSaving: isSaving,
                    saveLabel: L10n.Sources.saveArtifact,
                    cancelLabel: L10n.Sources.cancelEdit,
                    saveDisabled: !isDirty || isSaving || labelDraft.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty,
                    showsCancel: isDirty,
                    accessibilityIdentifierPrefix: "sources.page.artifact.\(artifact.id)",
                    onSave: { onSave(labelDraft, descriptionDraft) },
                    onCancel: {
                        labelDraft = artifact.label
                        descriptionDraft = artifact.description
                        onCancel()
                    }
                )

                if !isDirty {
                    Text(L10n.Sources.noUnsavedChanges)
                        .font(PVFont.body(size: PVTypeScale.caption, italic: true))
                        .foregroundStyle(PVColor.textFaint)
                }

                Spacer(minLength: 0)
            }
        }
        .onAppear { seedFromArtifact() }
        .onChange(of: artifact.label) { _, _ in
            if !isDirty { seedFromArtifact() }
        }
        .onChange(of: artifact.description) { _, _ in
            if !isDirty { seedFromArtifact() }
        }
    }

    private var isDirty: Bool {
        labelDraft.trimmingCharacters(in: .whitespacesAndNewlines)
            != artifact.label.trimmingCharacters(in: .whitespacesAndNewlines)
            || descriptionDraft.trimmingCharacters(in: .whitespacesAndNewlines)
            != artifact.description.trimmingCharacters(in: .whitespacesAndNewlines)
    }

    private func seedFromArtifact() {
        labelDraft = artifact.label
        descriptionDraft = artifact.description
    }
}
