import SwiftUI

/// The Person page's Family section (S13-12, board S13-D3): the person's
/// parents, siblings, partners and children as person tiles joined by lines.
/// It reads its own query (`ConclusionPageSection.queryKey`), keeps the diagram
/// drawn while it reloads, and opens a relative on their Family.
struct PersonFamilySection: View {
    let entityId: String

    @Environment(WorkspaceSession.self) private var session
    @Environment(WorkspaceNavigation.self) private var navigation
    @Environment(\.pvWidthClass) private var widthClass

    /// The pages of each list the researcher has unfolded. Reset when the page is another person.
    @State private var folds = ConclusionFoldState<PersonFamilyContent.Fold>()

    var body: some View {
        let orders = session.nameFormat
        let folds = folds
        ConclusionPageSectionHost<CatalogPersonFamily, PersonFamilyContent, PersonFamilyDiagram>(
            section: .family,
            entityId: entityId,
            errorTitle: L10n.Conclusions.familyErrorTitle,
            emptyTitle: L10n.Conclusions.familyEmptyTitle,
            emptyMessage: L10n.string(L10n.Conclusions.familyEmptyMessage),
            make: { PersonFamilyContent(family: $0, folds: folds, locale: .autoupdatingCurrent, orders: orders) },
            isEmpty: \.isEmpty
        ) { content in
            PersonFamilyDiagram(
                content: content,
                projectDir: session.projectKey.projectDir,
                widthClass: widthClass,
                open: { navigation.go(to: $0) },
                expand: { self.folds.revealPage($0) }
            )
        }
    }
}

/// The two columns laid out by `PersonFamilyLayout`: lines, captions, then one
/// container per group in reading order, so Tab runs the left column top to
/// bottom, then the right column group by group.
struct PersonFamilyDiagram: View {
    let content: PersonFamilyContent
    let projectDir: String
    let widthClass: PVWidthClass
    let open: (WorkspaceLocation) -> Void
    var expand: (PersonFamilyContent.Fold) -> Void = { _ in }

    @State private var width: CGFloat = 0
    /// The tile with the keyboard: a tile's id.
    @FocusState private var focused: String?
    /// The tile a keyboard Show more just revealed, to scroll into view.
    @State private var scrollTarget: String?
    @State private var scrollRequest = 0

    var body: some View {
        Group {
            if width > 0 {
                diagram(PersonFamilyLayout(content: content, width: width, widthClass: widthClass))
            } else {
                Color.clear.frame(height: 1)
            }
        }
        .frame(maxWidth: .infinity, alignment: .topLeading)
        .onGeometryChange(for: CGFloat.self) { $0.size.width } action: { width = $0 }
        .pvScrollsRowIntoView(scrollTarget, whenRequested: scrollRequest)
    }

    private func diagram(_ layout: PersonFamilyLayout) -> some View {
        let tiles = Dictionary(content.allTiles.map { ($0.id, $0) }, uniquingKeysWith: { first, _ in first })
        return ZStack(alignment: .topLeading) {
            Path { path in
                for segment in layout.segments { path.addRect(segment) }
            }
            .fill(PVColor.borderStrong)
            .accessibilityHidden(true)
            ForEach(Array(layout.captions.enumerated()), id: \.offset) { _, caption in
                Text(verbatim: caption.text)
                    .pvMicroCaps()
                    .foregroundStyle(PVColor.textMuted)
                    .fixedSize()
                    .offset(x: caption.origin.x, y: caption.origin.y)
                    .accessibilityHidden(true)
            }
            ForEach(Array(layout.groups.enumerated()), id: \.offset) { _, group in
                ZStack(alignment: .topLeading) {
                    ForEach(group.tiles, id: \.id) { placed in
                        if let tile = tiles[placed.id] {
                            PVPersonCard(content: tile.content, projectDir: projectDir) {
                                if let location = tile.location { open(location) }
                            }
                            .frame(width: placed.frame.width, height: placed.frame.height)
                            .focused($focused, equals: placed.id)
                            .id(placed.id)
                            .offset(x: placed.frame.minX, y: placed.frame.minY)
                            .accessibilityIdentifier("persons.detail.family.tile.\(placed.id)")
                        }
                    }
                    ForEach(group.more, id: \.fold) { more in
                        let revealed = content.firstRevealed(by: more.fold)
                        ConclusionShowMoreRow(
                            hidden: more.count,
                            fold: PersonFamilyContent.fold,
                            focus: $focused,
                            revealed: revealed
                        ) {
                            expand(more.fold)
                            scrollTarget = revealed
                            scrollRequest += 1
                        }
                            .frame(width: more.frame.width, height: more.frame.height, alignment: .leading)
                            .offset(x: more.frame.minX, y: more.frame.minY)
                            .accessibilityIdentifier("persons.detail.family.more")
                    }
                }
                .frame(width: group.frame.width, height: group.frame.height, alignment: .topLeading)
                .offset(x: group.frame.minX, y: group.frame.minY)
                .accessibilityElement(children: .contain)
                .accessibilityLabel(group.label)
            }
        }
        .frame(width: width, height: layout.height, alignment: .topLeading)
    }
}

#if DEBUG
#Preview("11a Full family") { PersonFamilyPreviewData.diagram(PersonFamilyPreviewData.full) }
#Preview("11c Conflicting parents") { PersonFamilyPreviewData.diagram(PersonFamilyPreviewData.conflicting) }
#Preview("11d Recorded siblings only") { PersonFamilyPreviewData.diagram(PersonFamilyPreviewData.siblingsOnly) }
#Preview("11h After clicking Anna") { PersonFamilyPreviewData.diagram(PersonFamilyPreviewData.annaOnly) }
#endif
