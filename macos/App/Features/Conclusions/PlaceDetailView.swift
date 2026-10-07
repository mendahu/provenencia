import SwiftUI

/// The Place page (S9-27, board S9-D7 frame 1f). A configuration of
/// `ConclusionDetailPage`: every kept name and its Why, and period,
/// hierarchy, and succession drawn empty until S9-40.
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

/// The parent chain under the title. Empty until S9-40 walks it, so the
/// line is the stated sentence. A recorded chain uses the same line.
private struct PlaceDetailChain: View {
    let text: String
    let recorded: Bool

    var body: some View {
        Text(verbatim: text)
            .font(PVFont.body(size: PVTypeScale.bodySmall, italic: true))
            .foregroundStyle(recorded ? PVColor.textPrimary : PVColor.textMuted)
    }
}
