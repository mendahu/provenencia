import SwiftUI

/// Resolves lead thumbnail + fills `PVOmnibarHitRow` slots for one hit.
struct CatalogSearchHitRow: View {
    let hit: CatalogSearchHit
    let projectDir: String
    var selected: Bool
    @Environment(WorkspaceSession.self) private var session

    var body: some View {
        PVOmnibarHitRow(
            lead: { leadView },
            title: CatalogSearchHitPresentation.title(for: hit, orders: session.nameFormat),
            kindLabel: CatalogSearchHitPresentation.kindLabel(for: hit.kind),
            secondary: CatalogSearchHitPresentation.secondary(for: hit),
            ref: hit.ref,
            refAccent: CatalogSearchHitPresentation.refAccent(for: hit),
            matchContext: CatalogSearchHitPresentation.matchContextText(
                field: hit.matchReason,
                snippet: hit.matchSnippet
            ),
            selected: selected
        )
    }

    @ViewBuilder
    private var leadView: some View {
        switch hit.kind {
        case "source":
            CachedThumbnail(
                projectDir: projectDir,
                relPath: hit.thumbnailRelPath,
                typeIconKey: hit.iconKey.isEmpty ? nil : hit.iconKey,
                size: 40,
                cornerRadius: PVRadius.sm
            )
        default:
            PVThumbnail(CatalogSearchHitPresentation.leadThumbnail(for: hit), size: 40)
        }
    }
}
