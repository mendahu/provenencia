import SwiftUI

/// Artifact cards: wrapping rows, on-card edit, and the Add dialog form.
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
                LazyVGrid(columns: cardColumns, spacing: PVSpacing.space6) {
                    ForEach(model.artifacts.items, id: \.id) { art in
                        artifactCard(art)
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
                        accessibilityIdentifierPrefix: "sources.page.addArtifact.graphs"
                    )
                    if model.artifacts.draftNeedsGraph {
                        Text(L10n.Sources.artifactGraphsNeedsOne)
                            .font(PVFont.body(size: PVTypeScale.caption, italic: true))
                            .foregroundStyle(PVColor.textMuted)
                            .accessibilityIdentifier("sources.page.addArtifact.graphs.needsGraph")
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

    private func artifactCard(_ art: CatalogArtifact) -> some View {
        let editing = model.artifacts.editingArtifactID == art.id
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
        .modifier(ArtifactCardHover())
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
            .help(Text(art.hasFile ? L10n.Sources.artifactOpenFileTip : L10n.Sources.artifactAddFileTip))
            .accessibilityLabel(previewLabel(art))
            .accessibilityIdentifier(
                art.hasFile
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
                .fixedSize(horizontal: true, vertical: false)
            if model.artifacts.isCover(art) {
                PVBadge(L10n.Sources.coverBadge, tone: .accent, icon: .photo)
                    .accessibilityIdentifier("sources.page.artifact.\(art.id).coverBadge")
            }
            if !editing, model.artifacts.canUseAsThumbnail(art) {
                PVButton(L10n.Sources.useAsThumbnail, variant: .ghost, size: .sm, icon: .imageUp) {
                    Task { await model.artifacts.useAsThumbnail(art) }
                }
                .fixedSize(horizontal: true, vertical: false)
                .layoutPriority(1)
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
                .fixedSize(horizontal: true, vertical: false)
            }
        }
        .frame(minHeight: 28)
    }

    private func restingText(_ art: CatalogArtifact) -> some View {
        VStack(alignment: .leading, spacing: PVSpacing.space3) {
            Text(verbatim: art.displayTitle)
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

    /// The Evidence graphs the card in edit links the Artifact to; absent on a
    /// Source with one graph.
    @ViewBuilder
    private func graphLinks(_ art: CatalogArtifact) -> some View {
        if let edit = model.artifacts.linkEdit, edit.artifactID == art.id {
            let prefix = "sources.page.artifact.\(art.id).graphs"
            PVField(
                label: L10n.Sources.artifactGraphsLabel,
                hint: L10n.Sources.artifactGraphsHint
            ) {
                SourceFormLoadArea(
                    state: edit.loadState,
                    loadingText: L10n.Sources.artifactGraphsLoading,
                    errorTitle: L10n.Sources.artifactGraphsLoadErrorTitle,
                    errorMessage: L10n.Sources.artifactGraphsLoadErrorMessage,
                    onReadAgain: { Task { await model.artifacts.loadLinkLocks() } },
                    compact: true,
                    accessibilityIdentifierPrefix: prefix
                ) {
                    SourceArtifactGraphLinks(
                        graphs: model.artifacts.linkEditGraphs,
                        excluded: Binding(
                            get: { model.artifacts.linkEdit?.excluded ?? [] },
                            set: { model.artifacts.linkEdit?.excluded = $0 }
                        ),
                        locked: edit.locked,
                        accessibilityIdentifierPrefix: prefix
                    )
                    .disabled(model.artifacts.savingID == art.id)
                }
            }
        }
    }

    private func editFields(_ art: CatalogArtifact) -> some View {
        VStack(alignment: .leading, spacing: PVSpacing.space5) {
            if let saveError = model.artifacts.cardSaveError {
                PVCallout(
                    tone: .danger,
                    title: L10n.Sources.artifactCardSaveErrorTitle,
                    message: saveError,
                    compact: true
                )
                .accessibilityIdentifier("sources.page.artifact.\(art.id).saveError")
            }
            PVField(
                label: L10n.Sources.artifactLabel,
                error: model.artifacts.cardLabelError(id: art.id),
                marker: .required
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
            graphLinks(art)
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
        if art.hasFile {
            model.artifacts.open(art)
        } else {
            model.artifacts.openAttach(to: art.id)
        }
    }

    private func previewLabel(_ art: CatalogArtifact) -> String {
        if art.hasFile {
            let name = art.file?.originalFilename ?? ""
            return L10n.Sources.openArtifactFile(name: name.isEmpty ? art.ref : name)
        }
        return L10n.Sources.addFileToArtifact(ref: art.ref)
    }

    private var sourceTypeIconKey: String? {
        model.identity.types.iconKey(forTypeID: model.source?.sourceTypeID)
    }
}

/// Grows the card to the board's 1.02 scale. Hover state stays on this card.
/// The hit shape is always the grown size, so the shrinking edge never
/// crosses the pointer and retriggers the hover. The card is one layer
/// before it scales, so the glyphs are not redrawn on every frame.
private struct ArtifactCardHover: ViewModifier {
    private static let hoverScale: CGFloat = 1.02

    @State private var isHovering = false
    @Environment(\.accessibilityReduceMotion) private var reduceMotion

    func body(content: Content) -> some View {
        let resting = PVElevation.sm[0]
        let lifted = PVElevation.md[1]
        let grow = isHovering && !reduceMotion
        content
            .compositingGroup()
            .shadow(color: resting.color, radius: resting.radius, x: resting.x, y: resting.y)
            .shadow(
                color: lifted.color.opacity(isHovering ? 1 : 0),
                radius: lifted.radius,
                x: lifted.x,
                y: lifted.y
            )
            .scaleEffect(grow ? Self.hoverScale : 1, anchor: .center)
            .animation(reduceMotion ? nil : .easeOut(duration: PVMotion.durationFast), value: isHovering)
            .contentShape(GrownHitShape(scale: Self.hoverScale))
            .onHover { hovering in
                guard hovering != isHovering else { return }
                isHovering = hovering
            }
            .zIndex(isHovering ? 1 : 0)
    }
}

/// Layout frame expanded to the hovered scale. It does not change with the
/// animation, so pointer exit happens once.
private struct GrownHitShape: Shape {
    var scale: CGFloat

    func path(in rect: CGRect) -> Path {
        let dx = rect.width * (scale - 1) / 2
        let dy = rect.height * (scale - 1) / 2
        return Path(rect.insetBy(dx: -dx, dy: -dy))
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
