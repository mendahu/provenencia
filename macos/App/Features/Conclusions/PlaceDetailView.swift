import SwiftUI

/// The Place page (S9-27 / S9-40, board S9-D7). A configuration of
/// `ConclusionDetailPage`: names with Why, period, parent chain, and
/// Part of / Contains / Succession relationship rows.
struct PlaceDetailView: View {
    let session: WorkspaceSession
    let entityId: String

    var body: some View {
        ConclusionDetailPage<PlaceDetailContent, PlaceDetailChain>(
            session: session,
            entityId: entityId,
            mark: .subjectPlace,
            thumbnailLabel: L10n.Conclusions.placeNoImage,
            pageIdentifier: "places.detail",
            errorIdentifier: "places.detail.error",
            make: { PlaceDetailContent(detail: $0) },
            summary: { PlaceDetailChain(text: $0.chain, recorded: $0.chainIsRecorded) }
        )
    }
}

/// The parent chain under the title, or the sentence that none is recorded.
private struct PlaceDetailChain: View {
    let text: String
    let recorded: Bool

    var body: some View {
        Text(verbatim: text)
            .font(PVFont.body(size: PVTypeScale.bodySmall, italic: true))
            .foregroundStyle(recorded ? PVColor.textPrimary : PVColor.textMuted)
    }
}
