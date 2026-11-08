import SwiftUI

/// The side column of the Person page's Sources section (S13-28): the title and caption,
/// then the record links diagram at the column's width. The circle shrinks to fit the part of
/// the page in view, to its smallest, below which `ConclusionPinnedSideColumn` lets the column
/// scroll with the page. Draws nothing for a person with no records on the read.
struct PersonRecordLinksColumn: View {
    let content: PersonRecordLinksContent
    let highlightedGraphID: String?
    let onHover: (String?) -> Void
    let open: (String) -> Void
    let focus: (String) -> PVFieldFocus

    @Environment(\.pvPageViewportSize) private var viewport
    /// Zero until the first layout pass has measured the column; the diagram is not drawn before.
    @State private var width: CGFloat = 0
    @State private var headerHeight: CGFloat = 0

    private var single: Bool { content.nodes.count == 1 }

    /// The height left for the drawing under the title block while the page is in view.
    private var fitting: CGFloat? {
        guard viewport.height > 0 else { return nil }
        return viewport.height - 2 * PVPageColumnsMetrics.verticalPadding - headerHeight - PVSpacing.space4
    }

    var body: some View {
        if !content.nodes.isEmpty {
            VStack(alignment: .leading, spacing: PVSpacing.space4) {
                VStack(alignment: .leading, spacing: PVSpacing.space2) {
                    Text(L10n.Conclusions.recordLinksTitle)
                        .pvMicroCaps()
                        .foregroundStyle(PVColor.textMuted)
                    if !single {
                        Text(L10n.Conclusions.recordLinksCaption)
                            .font(PVFont.body(size: PVTypeScale.caption))
                            .foregroundStyle(PVColor.textMuted)
                            .lineLimit(1)
                    }
                }
                .onGeometryChange(for: CGFloat.self) { $0.size.height } action: { headerHeight = $0 }
                if width > 0 {
                    RecordLinksDiagram(
                        nodes: content.nodes,
                        arcs: content.arcs,
                        highlightedGraphID: highlightedGraphID,
                        width: width,
                        diameter: single ? nil : RecordLinksLayout.diameter(width: width, fitting: fitting),
                        onHover: onHover,
                        open: open,
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
