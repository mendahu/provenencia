import SwiftUI

/// Resolves lead thumbnail + fills `PVOmnibarHitRow` slots for one hit.
struct CatalogSearchHitRow: View {
    let hit: CatalogSearchHit
    let projectDir: String
    @Environment(WorkspaceSession.self) private var session

    var body: some View {
        PVOmnibarHitRow(
            lead: { leadView },
            title: CatalogSearchHitPresentation.title(for: hit, orders: session.nameFormat),
            kindLabel: CatalogSearchHitPresentation.kindLabel(for: hit.searchKind),
            secondary: CatalogSearchHitPresentation.secondary(for: hit),
            ref: hit.ref,
            refAccent: CatalogSearchHitPresentation.refAccent(for: hit),
            matchContext: CatalogSearchHitPresentation.matchContextText(
                field: hit.matchReason,
                snippet: hit.matchSnippet
            )
        )
        .accessibilityLabel(CatalogSearchHitPresentation.accessibilityLabel(for: hit, orders: session.nameFormat))
    }

    @ViewBuilder
    private var leadView: some View {
        if let lead = CatalogSearchHitPresentation.leadThumbnail(for: hit) {
            PVThumbnail(lead, size: 40)
        } else {
            CachedThumbnail(
                projectDir: projectDir,
                relPath: CatalogSearchHitPresentation.leadRelPath(for: hit),
                typeIconKey: hit.iconKey.isEmpty ? nil : hit.iconKey,
                size: 40,
                cornerRadius: PVRadius.sm,
                fallbackMark: CatalogSearchHitPresentation.fallbackMark(for: hit)
            )
        }
    }
}
