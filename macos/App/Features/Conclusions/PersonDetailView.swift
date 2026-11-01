import SwiftUI

/// The Person page (S9-16, board S9-D5). A configuration of
/// `ConclusionDetailPage`: portrait, name, the b. / d. lines, and the sections under them. Life dates
/// and places come from the Person summary the detail carries.
struct PersonDetailView: View {
    let session: WorkspaceSession
    let entityId: String
    let userID: String

    var body: some View {
        ConclusionDetailPage(
            session: session,
            entityId: entityId,
            pageIdentifier: "persons.detail",
            errorIdentifier: "persons.detail.error",
            userID: userID,
            make: { PersonDetailContent(detail: $0, orders: session.nameFormat) },
            secondary: { PersonDetailLifeLines(lines: $0.lifeLines) },
            pageSections: { PersonDetailSections(content: $0, context: $1) }
        )
    }
}

/// The Person page below its header: the open section (Vitals, Gallery, or a section that
/// arrives later). The page owns which section is open and passes it in the context.
private struct PersonDetailSections: View {
    let content: PersonDetailContent
    let context: ConclusionDetailPageSectionsContext

    var body: some View {
        let section = context.pageSection ?? .vitals
        switch section {
        case .vitals:
            ConclusionDetailRowsSection(
                rows: content.rows,
                refreshing: context.refreshing,
                title: nil,
                provenance: .sheet
            )
        case .gallery:
            if content.gallery.tiles.isEmpty {
                PVEmptyState(
                    glyph: ConclusionPageSection.gallery.glyph,
                    title: L10n.Conclusions.galleryEmptyTitle,
                    message: L10n.string(L10n.Conclusions.galleryEmptyMessage)
                )
            } else {
                ConclusionDetailGallerySection(
                    projectDir: context.projectDir,
                    pageIdentifier: context.pageIdentifier,
                    gallery: content.gallery,
                    lead: context.lead
                )
            }
        case .family, .timeline, .sources:
            PVEmptyState(
                glyph: section.glyph,
                title: section.label,
                message: L10n.Conclusions.pageSectionLater(L10n.string(section.label))
            )
        }
    }
}

/// The two life lines under a Person's name.
private struct PersonDetailLifeLines: View {
    let lines: [PersonLifeDisplay.LifeLine]
    @Environment(\.conclusionDetailHeaderCompact) private var compact

    var body: some View {
        // Compact (the pinned header after a scroll) puts the two lines side by side.
        let layout = compact
            ? AnyLayout(HStackLayout(alignment: .firstTextBaseline, spacing: PVSpacing.space6))
            : AnyLayout(VStackLayout(alignment: .leading, spacing: PVSpacing.space1))
        layout {
            ForEach(lines, id: \.abbreviation) { line in
                ConclusionDatePlaceLine(
                    leading: line.abbreviation,
                    date: line.date,
                    place: line.place,
                    accessibilityLabel: line.accessibilityLabel
                )
            }
        }
    }
}
