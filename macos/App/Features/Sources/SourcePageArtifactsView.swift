import SwiftUI

/// Artifact cards: wrapping rows, on-card edit, and the Add dialog form.
struct SourcePageArtifactsView: View {
    /// Board hover on the card wrapper (`transform: scale(1.02)`).
    private static let cardHoverScale: CGFloat = 1.02

    @Bindable var model: SourcePageModel
    @FocusState private var focusedPreviewID: String?
    @State private var hoveredCardID: String?

    private var cardColumns: [GridItem] {
        [GridItem(.adaptive(minimum: 300), spacing: PVSpacing.space7, alignment: .top)]
    }

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
                    message: L10n.string(L10n.Sources.artifactsEmptyMessage),
                    compact: true
                ) {
                    PVButton(L10n.Sources.addArtifact, variant: .primary, size: .sm, icon: .plus) {
                        model.artifacts.openAdd()
                    }
                    .accessibilityIdentifier("sources.page.artifacts.empty.add")
                }
                .accessibilityIdentifier("sources.page.artifacts.empty")
            } else {
                LazyVGrid(columns: cardColumns, spacing: PVSpacing.space7) {
                    ForEach(model.artifacts.items, id: \.id) { art in
                        artifactCard(art)
                            .frame(maxWidth: .infinity, alignment: .top)
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
                IngestFileDropRow(
                    fileName: model.artifacts.draft.fileName,
                    isRejected: model.artifacts.draft.reject != nil,
                    isBusy: model.artifacts.isSavingDraft,
                    idleHint: L10n.Sources.ingestDropIdleCreate,
                    onChoose: { model.artifacts.pickFile() },
                    onClear: { model.artifacts.clearFile() },
                    onDropURLs: { model.artifacts.applyDroppedURLs($0, into: .create) }
                )
                if let reject = model.artifacts.draft.reject {
                    PVCallout(tone: .danger, title: reject.title, message: reject.message)
                        .accessibilityIdentifier("sources.page.addArtifact.ingestReject")
                }
            }
        }
    }

    /// Form for the first-attach Add File dialog.
    var addFileForm: some View {
        VStack(alignment: .leading, spacing: PVSpacing.space6) {
            PVCallout(
                tone: .info,
                message: L10n.string(L10n.Sources.addFileFirstAttachHint),
                compact: true
            )
            if !model.artifacts.attachDraft.artifactRef.isEmpty {
                Text(model.artifacts.attachDraft.artifactRef)
                    .font(PVFont.mono(size: PVTypeScale.caption))
                    .foregroundStyle(PVColor.textMuted)
            }
            IngestFileDropRow(
                fileName: model.artifacts.attachDraft.fileName,
                isRejected: model.artifacts.attachDraft.reject != nil,
                isBusy: model.artifacts.isSavingAttach,
                idleHint: L10n.Sources.ingestDropIdleAttach,
                onChoose: { model.artifacts.pickAttachFile() },
                onClear: { model.artifacts.clearAttachFile() },
                onDropURLs: { model.artifacts.applyDroppedURLs($0, into: .attach) }
            )
            if let reject = model.artifacts.attachDraft.reject {
                PVCallout(tone: .danger, title: reject.title, message: reject.message)
                    .accessibilityIdentifier("sources.page.addFile.ingestReject")
            }
        }
    }

    private func artifactCard(_ art: CatalogArtifact) -> some View {
        let editing = model.artifacts.editingArtifactID == art.id
        let hovering = hoveredCardID == art.id
        return PVCard(elevated: false, padding: 0) {
            VStack(alignment: .leading, spacing: 0) {
                previewWell(art)
                VStack(alignment: .leading, spacing: PVSpacing.space4) {
                    metaRow(art, editing: editing)
                    if editing {
                        editFields(art)
                    } else {
                        restingText(art)
                    }
                }
                .padding(.horizontal, PVSpacing.space6)
                .padding(.top, PVSpacing.space6)
                .padding(.bottom, PVSpacing.space5)
                .frame(maxWidth: .infinity, alignment: .topLeading)
            }
            .frame(maxWidth: .infinity, alignment: .top)
        }
        .pvShadow(hovering ? PVElevation.md : PVElevation.sm)
        .offset(y: hovering ? PVMotion.liftHover : 0)
        .scaleEffect(hovering ? Self.cardHoverScale : 1)
        .zIndex(hovering ? 1 : 0)
        .onHover { inside in
            if inside {
                hoveredCardID = art.id
            } else if hoveredCardID == art.id {
                hoveredCardID = nil
            }
        }
        .pvAnimation(PVMotion.easeStandard, value: hovering)
        .accessibilityIdentifier("sources.page.artifact.\(art.id)")
    }

    /// 1:1 well sized from the card width. Extra card height never reaches it.
    private func previewWell(_ art: CatalogArtifact) -> some View {
        ArtifactSquarePreview {
            CachedThumbnail(
                projectDir: model.pageProjectDir,
                relPath: art.thumbnailRelPath,
                mediaType: art.file?.mediaType ?? "",
                originalFilename: art.file?.originalFilename ?? "",
                typeIconKey: artifactTypeIconKey(for: art),
                layout: .well
            )
        }
        .fixedSize(horizontal: false, vertical: true)
        .background(PVColor.surfaceInset)
        .clipped()
        .overlay(alignment: .bottom) {
            Rectangle()
                .fill(PVColor.borderSubtle)
                .frame(height: 1)
                .allowsHitTesting(false)
        }
        .overlay {
            Button {
                activatePreview(art)
            } label: {
                Color.clear
                    .frame(maxWidth: .infinity, maxHeight: .infinity)
                    .contentShape(Rectangle())
            }
            .buttonStyle(.plain)
            .focused($focusedPreviewID, equals: art.id)
            .pvFocusRing(focusedPreviewID == art.id, cornerRadius: 0)
            .help(Text(hasFile(art) ? L10n.Sources.artifactOpenFileTip : L10n.Sources.artifactAddFileTip))
            .accessibilityLabel(previewLabel(art))
            .accessibilityIdentifier(
                hasFile(art)
                    ? "sources.page.artifact.\(art.id).open"
                    : "sources.page.artifact.\(art.id).addFile"
            )
        }
    }

    private func metaRow(_ art: CatalogArtifact, editing: Bool) -> some View {
        HStack(spacing: PVSpacing.space4) {
            Text(verbatim: art.ref)
                .font(PVFont.mono(size: PVTypeScale.caption))
                .foregroundStyle(PVColor.textMuted)
            if model.artifacts.isCover(art) {
                PVBadge(L10n.Sources.coverBadge, tone: .accent, icon: .photo)
                    .accessibilityIdentifier("sources.page.artifact.\(art.id).coverBadge")
            }
            if !editing, model.artifacts.canUseAsThumbnail(art) {
                PVButton(L10n.Sources.useAsThumbnail, variant: .ghost, size: .sm, icon: .imageUp) {
                    Task { await model.artifacts.useAsThumbnail(art) }
                }
                .accessibilityIdentifier("sources.page.artifact.\(art.id).useAsThumbnail")
            }
            Spacer(minLength: 0)
            if !editing {
                HStack(spacing: PVSpacing.space2) {
                    PVIconButton(
                        .penLine,
                        label: L10n.Sources.editArtifactLabel(ref: art.ref),
                        size: .sm
                    ) {
                        model.artifacts.beginEdit(art.id)
                    }
                    .accessibilityIdentifier("sources.page.artifact.\(art.id).edit")
                    PVIconButton(
                        .trash,
                        label: L10n.Sources.deleteArtifact,
                        size: .sm,
                        tone: .danger
                    ) {
                        Task { await model.askDeleteArtifact(id: art.id) }
                    }
                    .accessibilityIdentifier("sources.page.artifact.\(art.id).delete")
                }
            }
        }
        .frame(minHeight: 28)
    }

    private func restingText(_ art: CatalogArtifact) -> some View {
        VStack(alignment: .leading, spacing: PVSpacing.space3) {
            Text(verbatim: art.label.isEmpty ? art.ref : art.label)
                .font(PVFont.body(size: PVTypeScale.body, weight: PVFontWeight.semibold))
                .foregroundStyle(PVColor.textPrimary)
                .lineLimit(2)
            descriptionLine(art)
        }
    }

    private func descriptionLine(_ art: CatalogArtifact) -> some View {
        let minHeight = PVTypeScale.caption * PVLineHeight.normal * 3
        let trimmed = art.description.trimmingCharacters(in: .whitespacesAndNewlines)
        return Group {
            if trimmed.isEmpty {
                Text(L10n.Sources.artifactNoDescription)
                    .font(PVFont.body(size: PVTypeScale.caption, italic: true))
                    .foregroundStyle(PVColor.textFaint)
            } else {
                Text(verbatim: art.description)
                    .font(PVFont.body(size: PVTypeScale.caption))
                    .foregroundStyle(PVColor.textSecondary)
                    .lineLimit(3)
            }
        }
        .frame(maxWidth: .infinity, minHeight: minHeight, alignment: .topLeading)
    }

    private func editFields(_ art: CatalogArtifact) -> some View {
        VStack(alignment: .leading, spacing: PVSpacing.space5) {
            PVField(
                label: L10n.Sources.artifactLabel,
                error: model.artifacts.cardLabelError(id: art.id),
                required: true
            ) {
                PVInput(
                    text: labelBinding(art),
                    size: .sm,
                    isInvalid: model.artifacts.cardLabelError(id: art.id) != nil,
                    activateOnAppear: true
                )
                .accessibilityIdentifier("sources.page.artifact.\(art.id).label")
                .onSubmit { commitEdit(art) }
            }
            PVField(
                label: L10n.Sources.artifactDescription,
                hint: L10n.Sources.artifactDescriptionHint
            ) {
                PVTextArea(
                    text: descriptionBinding(art),
                    lineLimit: 3...8
                )
                .accessibilityIdentifier("sources.page.artifact.\(art.id).description")
            }
            HStack(spacing: PVSpacing.space3) {
                PVButton(
                    L10n.Sources.saveArtifact,
                    variant: .primary,
                    size: .sm,
                    loading: model.artifacts.savingID == art.id
                ) {
                    commitEdit(art)
                }
                .disabled(!model.artifacts.canSaveFields(art.id))
                .accessibilityIdentifier("sources.page.artifact.\(art.id).save")
                PVButton(L10n.Sources.cancelEdit, variant: .ghost, size: .sm) {
                    model.artifacts.cancelFields(id: art.id)
                }
                .accessibilityIdentifier("sources.page.artifact.\(art.id).cancel")
            }
        }
        .onExitCommand { model.artifacts.cancelFields(id: art.id) }
        .background {
            Button {
                commitEdit(art)
            } label: {
                EmptyView()
            }
            .keyboardShortcut(.return, modifiers: .command)
            .disabled(!model.artifacts.canSaveFields(art.id))
            .accessibilityHidden(true)
        }
    }

    private func labelBinding(_ art: CatalogArtifact) -> Binding<String> {
        Binding(
            get: { model.artifacts.labels[art.id] ?? art.label },
            set: { model.artifacts.labels[art.id] = $0 }
        )
    }

    private func descriptionBinding(_ art: CatalogArtifact) -> Binding<String> {
        Binding(
            get: { model.artifacts.descriptions[art.id] ?? art.description },
            set: { model.artifacts.descriptions[art.id] = $0 }
        )
    }

    private func commitEdit(_ art: CatalogArtifact) {
        guard model.artifacts.canSaveFields(art.id) else { return }
        Task { await model.artifacts.saveFields(id: art.id) }
    }

    private func activatePreview(_ art: CatalogArtifact) {
        if hasFile(art) {
            model.artifacts.open(art)
        } else {
            model.artifacts.openAttach(to: art.id)
        }
    }

    private func previewLabel(_ art: CatalogArtifact) -> String {
        if hasFile(art) {
            let name = art.file?.originalFilename ?? ""
            return L10n.Sources.openArtifactFile(name: name.isEmpty ? art.ref : name)
        }
        return L10n.Sources.addFileToArtifact(ref: art.ref)
    }

    private func hasFile(_ art: CatalogArtifact) -> Bool {
        art.file != nil && !art.fileID.isEmpty
    }

    private func artifactTypeIconKey(for art: CatalogArtifact) -> String? {
        guard art.file == nil || art.fileID.isEmpty else { return nil }
        return model.identity.types
            .first { $0.id == model.source?.sourceTypeID }?
            .iconKey
    }
}

/// Thumbnail frame whose height is always its width. A taller card does not stretch it.
private struct ArtifactSquarePreview<Content: View>: View {
    @ViewBuilder var content: () -> Content

    var body: some View {
        SquareLayout { content() }
    }

    private struct SquareLayout: Layout {
        func sizeThatFits(proposal: ProposedViewSize, subviews: Subviews, cache: inout ()) -> CGSize {
            let proposed = proposal.width ?? 0
            let width = proposed.isFinite ? proposed : 0
            return CGSize(width: width, height: width)
        }

        func placeSubviews(in bounds: CGRect, proposal: ProposedViewSize, subviews: Subviews, cache: inout ()) {
            let side = bounds.width
            let child = ProposedViewSize(width: side, height: side)
            for subview in subviews {
                subview.place(at: bounds.origin, anchor: .topLeading, proposal: child)
            }
        }
    }
}
