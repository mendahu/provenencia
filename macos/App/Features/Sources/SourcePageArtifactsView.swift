import SwiftUI

/// Artifact cards: wrapping rows that open the Artifact sheet, and the Add dialog form.
struct SourcePageArtifactsView: View {
    /// Narrowest card. At `PVSpacing.widthContentMax` this is four across.
    private static let cardMinWidth: CGFloat = 264

    @Bindable var model: SourcePageModel
    @FocusState private var focusedPreviewID: String?

    private var cardColumns: [GridItem] {
        [GridItem(.adaptive(minimum: Self.cardMinWidth), spacing: PVSpacing.space6, alignment: .top)]
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
                let inNoGraph = model.artifacts.inNoGraph
                LazyVGrid(columns: cardColumns, spacing: PVSpacing.space6) {
                    ForEach(model.artifacts.items, id: \.id) { art in
                        artifactCard(art, noGraph: inNoGraph.contains(art.id))
                            .frame(maxWidth: .infinity, alignment: .top)
                    }
                }
                .accessibilityIdentifier("sources.page.artifacts.list")
            }
        }
        .sourcePageSectionTopPadding()
    }

    /// Label / description / optional file form for the Add Artifact dialog.
    var addForm: some View {
        VStack(alignment: .leading, spacing: PVSpacing.space6) {
            PVField(
                label: L10n.Sources.artifactLabel,
                hint: L10n.Sources.artifactLabelHint,
                error: model.artifacts.draftLabelError,
                marker: .required
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
            if let graphs = model.artifacts.visibleGraphs {
                PVField(
                    label: L10n.Sources.artifactGraphsLabel,
                    hint: L10n.Sources.artifactGraphsHint
                ) {
                    SourceArtifactGraphLinks(
                        graphs: graphs,
                        excluded: $model.artifacts.draft.excludedGraphIDs,
                        isDisabled: model.artifacts.draftGraphsLocked,
                        accessibilityIdentifierPrefix: "sources.page.addArtifact.graphs"
                    )
                    if model.artifacts.draftInNoGraph {
                        SourceArtifactNoGraphLine(accessibilityIdentifier: "sources.page.addArtifact.graphs.none")
                    }
                }
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

    private func artifactCard(_ art: CatalogArtifact, noGraph: Bool) -> some View {
        PVCard(elevated: false, padding: 0) {
            VStack(alignment: .leading, spacing: 0) {
                previewWell(art)
                VStack(alignment: .leading, spacing: PVSpacing.space4) {
                    metaRow(art)
                    restingText(art, noGraph: noGraph)
                }
                .padding(.horizontal, PVSpacing.space6)
                .padding(.top, PVSpacing.space6)
                .padding(.bottom, PVSpacing.space5)
                .frame(maxWidth: .infinity, alignment: .topLeading)
            }
            .frame(maxWidth: .infinity, alignment: .top)
        }
        .pvLiftHover(scale: 1.02, resting: PVElevation.sm)
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
                typeIconKey: art.thumbnailTypeIconKey(sourceTypeIconKey: sourceTypeIconKey),
                layout: .well
            )
        }
        .fixedSize(horizontal: false, vertical: true)
        .background(PVColor.surfaceInset)
        .clipped()
        .overlay(alignment: .bottom) {
            PVDivider()
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
            .buttonStyle(PVFocusRingButtonStyle(cornerRadius: 0))
            .focused($focusedPreviewID, equals: art.id)
            .help(Text(art.hasFile ? L10n.Sources.artifactOpenFileTip : L10n.Sources.artifactAddFileTip))
            .accessibilityLabel(previewLabel(art))
            .accessibilityIdentifier(
                art.hasFile
                    ? "sources.page.artifact.\(art.id).open"
                    : "sources.page.artifact.\(art.id).addFile"
            )
        }
    }

    private func metaRow(_ art: CatalogArtifact) -> some View {
        HStack(spacing: PVSpacing.space4) {
            Text(verbatim: art.ref)
                .font(PVFont.mono(size: PVTypeScale.caption))
                .foregroundStyle(PVColor.textMuted)
                .fixedSize(horizontal: true, vertical: false)
            if model.artifacts.isCover(art) {
                PVBadge(L10n.Sources.coverBadge, tone: .accent, icon: .photo)
                    .accessibilityIdentifier("sources.page.artifact.\(art.id).coverBadge")
            }
            if model.artifacts.canUseAsThumbnail(art) {
                PVButton(L10n.Sources.useAsThumbnail, variant: .ghost, size: .sm, icon: .imageUp) {
                    Task { await model.artifacts.useAsThumbnail(art) }
                }
                .fixedSize(horizontal: true, vertical: false)
                .layoutPriority(1)
                .accessibilityIdentifier("sources.page.artifact.\(art.id).useAsThumbnail")
            }
            Spacer(minLength: 0)
            HStack(spacing: PVSpacing.space2) {
                PVIconButton(
                    .penLine,
                    label: L10n.Sources.editArtifactLabel(ref: art.ref),
                    size: .sm
                ) {
                    model.artifacts.openSheet(artifactID: art.id)
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
            .fixedSize(horizontal: true, vertical: false)
        }
        .frame(minHeight: 28)
    }

    private func restingText(_ art: CatalogArtifact, noGraph: Bool) -> some View {
        VStack(alignment: .leading, spacing: PVSpacing.space3) {
            Text(verbatim: art.displayTitle)
                .font(PVFont.body(size: PVTypeScale.body, weight: PVFontWeight.semibold))
                .foregroundStyle(PVColor.textPrimary)
                .lineLimit(2)
            descriptionLine(art)
            ArtifactCopyDetailsView(
                rows: model.artifacts.copyDetails(for: art.id),
                ref: art.ref,
                artifactID: art.id,
                onMore: { model.artifacts.openSheet(artifactID: art.id, at: .copyDetails) }
            )
            if noGraph {
                noGraphRow(art)
            }
        }
    }

    /// The badge and **Link to a graph** on a card in no graph: one row with the
    /// button at the end, else the button on its own line at the end. A card
    /// with the row is taller than its neighbours.
    private func noGraphRow(_ art: CatalogArtifact) -> some View {
        let badge = PVBadge(L10n.Sources.artifactNotInGraph, tone: .neutral, icon: .link)
            .fixedSize()
            .accessibilityIdentifier("sources.page.artifact.\(art.id).noGraph")
        let button = PVButton(L10n.Sources.artifactLinkToGraph, variant: .ghost, size: .sm) {
            model.artifacts.openSheet(artifactID: art.id, at: .graphs)
        }
        .fixedSize()
        .accessibilityLabel(L10n.Sources.artifactLinkToGraphLabel(ref: art.ref))
        .accessibilityIdentifier("sources.page.artifact.\(art.id).linkToGraph")
        return ViewThatFits(in: .horizontal) {
            HStack(spacing: PVSpacing.space3) {
                badge
                Spacer(minLength: 0)
                button
            }
            VStack(alignment: .leading, spacing: PVSpacing.space3) {
                badge
                button.frame(maxWidth: .infinity, alignment: .trailing)
            }
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

    private func activatePreview(_ art: CatalogArtifact) {
        if art.hasFile {
            Task { await model.artifacts.open(art) }
        } else {
            model.artifacts.openAttach(to: art.id)
        }
    }

    private func previewLabel(_ art: CatalogArtifact) -> String {
        if art.hasFile {
            return L10n.Sources.openArtifactFile(name: art.fileDisplayName)
        }
        return L10n.Sources.addFileToArtifact(ref: art.ref)
    }

    private var sourceTypeIconKey: String? {
        model.identity.types.iconKey(forTypeID: model.source?.sourceTypeID)
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
