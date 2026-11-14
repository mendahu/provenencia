import SwiftUI

/// Artifact rows that open the Artifact sheet, and the Add file dialog form.
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
                        Task { await model.artifacts.openAdd() }
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
                        Task { await model.artifacts.openAdd() }
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
                file: $model.artifacts.attachDraft.file,
                isBusy: model.artifacts.isSavingAttach,
                idleHint: L10n.Sources.ingestDropIdleAttach,
                rejectAccessibilityIdentifier: "sources.page.addFile.ingestReject"
            )
        }
    }
}
