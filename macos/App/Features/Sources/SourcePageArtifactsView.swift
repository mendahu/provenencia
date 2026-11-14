import SwiftUI

/// Artifact rows that open the Artifact sheet, and the Add dialog form.
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
                let lines = model.artifacts.graphLines
                LazyVStack(alignment: .leading, spacing: 0) {
                    PVDivider()
                    ForEach(model.artifacts.items, id: \.id) { art in
                        SourcePageArtifactRow(art: art, model: model, graphLine: lines[art.id])
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
}
