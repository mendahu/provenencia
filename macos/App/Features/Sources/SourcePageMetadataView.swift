import SwiftUI

/// The Source page's Metadata section: header and its board. The header's help
/// icon carries the intro as a tooltip; its Add field opens the board's inline row.
struct SourcePageMetadataView: View {
    @Bindable var model: SourcePageModel

    var body: some View {
        VStack(alignment: .leading, spacing: PVSpacing.space6) {
            PVSectionHeader(
                title: L10n.Sources.metadataHeading,
                aside: {
                    PVIcon(.circleHelp, size: 15)
                        .foregroundStyle(PVColor.textMuted)
                        .help(L10n.string(L10n.Sources.metadataIntro))
                        .accessibilityElement()
                        .accessibilityLabel(L10n.string(L10n.Sources.metadataHelpLabel))
                        .accessibilityHint(L10n.string(L10n.Sources.metadataIntro))
                        .accessibilityIdentifier("sources.page.metadataHelp")
                },
                actions: {
                    PVButton(L10n.Sources.addMetadata, variant: .ghost, size: .sm, icon: .plus) {
                        model.metadata.openAddRow()
                    }
                    .accessibilityIdentifier("sources.page.addMetadata")
                }
            )

            MetadataBoardView(
                board: model.metadata,
                placement: .page,
                accessibilityIdentifierPrefix: "sources.page.metadata",
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
