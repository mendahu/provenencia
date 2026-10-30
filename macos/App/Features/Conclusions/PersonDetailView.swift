import SwiftUI

/// The Person page (S9-16, board S9-D5). A configuration of
/// `ConclusionDetailPage`: portrait, name, the b. / d. lines, and the sections under them. Life dates
/// and places come from the Person header the detail carries.
struct PersonDetailView: View {
    let session: WorkspaceSession
    let entityId: String
    @State private var portrait: ConclusionPortraitModel

    init(session: WorkspaceSession, entityId: String, userID: String) {
        self.session = session
        self.entityId = entityId
        _portrait = State(
            initialValue: ConclusionPortraitModel(session: session, userID: userID, entityId: entityId)
        )
    }

    var body: some View {
        ConclusionDetailPage(
            session: session,
            entityId: entityId,
            pageIdentifier: "persons.detail",
            errorIdentifier: "persons.detail.error",
            make: { PersonDetailContent(detail: $0, orders: session.nameFormat) },
            summary: { PersonDetailLifeLines(lines: $0.lifeLines) },
            portrait: portrait,
            pageSections: { PersonDetailSections(content: $0, context: $1) }
        )
        .vocabularyToastOverlay($portrait.toast, identifier: "persons.detail.toast")
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
            if content.likenesses.isEmpty {
                PVEmptyState(
                    glyph: ConclusionPageSection.gallery.glyph,
                    title: L10n.Conclusions.galleryEmptyTitle,
                    message: L10n.string(L10n.Conclusions.galleryEmptyMessage)
                )
            } else {
                ConclusionDetailLikenessSection(
                    projectDir: context.projectDir,
                    likenesses: content.likenesses,
                    portrait: context.portrait
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
    let lines: [PersonDetailContent.LifeLine]
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
