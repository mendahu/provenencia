import SwiftUI

/// The Source page's Metadata section: header, intro and its board.
struct SourcePageMetadataView: View {
    @Bindable var model: SourcePageModel

    var body: some View {
        VStack(alignment: .leading, spacing: PVSpacing.space6) {
            PVSectionHeader(
                title: L10n.Sources.metadataHeading,
                meta: model.metadata.saved.isEmpty
                    ? nil
                    : L10n.Sources.metadataFieldCount(model.metadata.saved.count),
                actions: {
                    PVButton(L10n.Sources.addMetadata, variant: .primary, size: .sm, icon: .plus) {
                        model.metadataDialogs.openAdd(on: model.metadata)
                    }
                    .accessibilityIdentifier("sources.page.addMetadata")
                }
            )

            Text(L10n.Sources.metadataIntro)
                .font(PVFont.body(size: PVTypeScale.caption))
                .foregroundStyle(PVColor.textMuted)
                .frame(maxWidth: PVSpacing.measureProse, alignment: .leading)

            MetadataBoardView(
                board: model.metadata,
                accessibilityIdentifierPrefix: "sources.page.metadata",
                onAdd: { model.metadataDialogs.openAdd(on: model.metadata) },
                onClear: { id in
                    Task {
                        if await model.metadata.requestClear(fieldID: id) {
                            model.metadataDialogs.askClear(on: model.metadata, fieldID: id)
                        }
                    }
                }
            )
        }
    }
}
