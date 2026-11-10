import SwiftUI

/// The side column of the Person page's Sources section (S13-28): the title and caption (or,
/// when the drawing is partial, the note that says so in its place), then the record links diagram at the column's width. The circle shrinks to fit the part of
/// the page in view, to its smallest, below which `ConclusionPinnedSideColumn` lets the column
/// scroll with the page. Draws nothing for a person with no records on the read.
struct PersonRecordLinksColumn: View {
    let content: PersonRecordLinksContent
    let accessibilityIdentifierRoot: String
    let hover: CircularLinkHover<String>
    let focusedGraphID: String?
    let open: (String) -> Void
    let focus: (String) -> PVFieldFocus

    @Environment(\.pvPageViewportSize) private var viewport
    /// Zero until the first layout pass has measured the column; the diagram is not drawn before.
    @State private var width: CGFloat = 0
    @State private var headerHeight: CGFloat = 0

    private var single: Bool { content.dataset.nodes.count == 1 }

    /// The height left for the drawing under the title block while the page is in view.
    private var fitting: CGFloat? {
        guard viewport.height > 0 else { return nil }
        return viewport.height - 2 * PVPageColumnsMetrics.verticalPadding - headerHeight - PVSpacing.space4
    }

    var body: some View {
        if !content.dataset.nodes.isEmpty {
            VStack(alignment: .leading, spacing: PVSpacing.space4) {
                VStack(alignment: .leading, spacing: PVSpacing.space2) {
                    Text(L10n.Conclusions.recordLinksTitle)
                        .pvMicroCaps()
                        .foregroundStyle(PVColor.textMuted)
                    if let partial = content.partial {
                        PVCallout(
                            tone: .warning,
                            message: L10n.string(partial == .notCounted
                                ? L10n.Conclusions.recordLinksNotCounted
                                : L10n.Conclusions.recordLinksPartial),
                            compact: true
                        )
                        .accessibilityIdentifier("\(accessibilityIdentifierRoot).recordLinks.partial")
                    } else if !single {
                        Text(L10n.Conclusions.recordLinksCaption)
                            .font(PVFont.body(size: PVTypeScale.caption))
                            .foregroundStyle(PVColor.textMuted)
                            .lineLimit(1)
                    }
                }
                .onGeometryChange(for: CGFloat.self) { $0.size.height } action: { headerHeight = $0 }
                if width > 0 {
                    CircularLinkDiagram(
                        dataset: content.dataset,
                        width: width,
                        fittingHeight: fitting,
                        hover: hover,
                        focusedID: focusedGraphID,
                        accessibilityLabel: L10n.Conclusions.recordLinksGroupAccessibility(count: content.dataset.nodes.count),
                        accessibilityIdentifierRoot: accessibilityIdentifierRoot,
                        onOpen: open,
                        focus: focus
                    )
                    if single {
                        Text(L10n.Conclusions.recordLinksOneRecord)
                            .font(PVFont.body(size: PVTypeScale.caption))
                            .italic()
                            .foregroundStyle(PVColor.textMuted)
                            .frame(maxWidth: .infinity)
                    }
                }
            }
            .frame(maxWidth: .infinity, alignment: .topLeading)
            .onGeometryChange(for: CGFloat.self) { $0.size.width } action: { width = $0 }
        }
    }
}

#if DEBUG
private struct PersonRecordLinksPreview: View {
    let links: CatalogPersonRecordLinks
    var width: CGFloat = 340
    @State private var hover = CircularLinkHover<String>()

    var body: some View {
        let content = PersonRecordLinksContent(links: links, graphLocations: [:])
        CircularLinkDiagram(
            dataset: content.dataset,
            width: width,
            hover: hover,
            accessibilityLabel: L10n.Conclusions.recordLinksGroupAccessibility(count: content.dataset.nodes.count),
            accessibilityIdentifierRoot: "preview",
            focus: { _ in PVFieldFocus(isActive: { false }, focus: {}, resign: {}, apply: { $0 }) }
        )
        .padding()
        .background(PVColor.surfacePage)
    }
}

#Preview("Johann") {
    PersonRecordLinksPreview(links: PersonSourcesPreviewData.johannRecordLinks)
}

#Preview("One graph") {
    PersonRecordLinksPreview(links: CatalogPersonRecordLinks(
        graphs: [PersonSourcesPreviewData.johannRecordLinks.graphs[0]], pairs: []
    ))
}

#Preview("50 household graphs, Wide") {
    PersonRecordLinksPreview(links: PersonSourcesPreviewData.householdRecordLinks(graphs: 50), width: 720)
}

#Preview("50 graphs, Wide") {
    PersonRecordLinksPreview(links: PersonSourcesPreviewData.manyRecordLinks(graphs: 50), width: 720)
}
#endif
