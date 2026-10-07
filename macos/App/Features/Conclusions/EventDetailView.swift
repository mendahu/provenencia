import SwiftUI

/// The Event page (S9-24, board S9-D6 frame 2f). A configuration of
/// `ConclusionDetailPage`: the event mark, a date · place line, and the
/// Date and Place rows. Subjects and places come from the Event header.
struct EventDetailView: View {
    let session: WorkspaceSession
    let entityId: String

    private var headerKey: CatalogQueryKey {
        .eventHeader(project: session.projectKey, entityId: entityId)
    }

    var body: some View {
        let header: CatalogEventHeader? = session.queryHandle(headerKey)?.value
        ConclusionDetailPage<EventDetailContent, EventDetailSummary>(
            session: session,
            entityId: entityId,
            mark: .subjectEvent,
            noLikeness: L10n.Conclusions.personNoLikeness,
            pageIdentifier: "events.detail",
            errorIdentifier: "events.detail.error",
            make: { EventDetailContent(detail: $0, header: header) },
            summary: { EventDetailSummary(date: $0.summaryDate, place: $0.summaryPlace) }
        )
        .task(id: entityId) {
            let _: QueryHandle<CatalogEventHeader> = session.query(headerKey)
        }
    }
}

/// Compact date, then the place, under the Event title.
private struct EventDetailSummary: View {
    let date: String?
    let place: String?

    var body: some View {
        HStack(alignment: .firstTextBaseline, spacing: PVSpacing.space4) {
            if let date {
                Text(verbatim: date)
                    .font(PVFont.mono(size: PVTypeScale.caption))
                    .foregroundStyle(PVColor.textPrimary)
            } else {
                Text(L10n.Conclusions.personDateUnknown)
                    .font(PVFont.body(size: PVTypeScale.bodySmall, italic: true))
                    .foregroundStyle(PVColor.textMuted)
            }
            Text(verbatim: "·")
                .foregroundStyle(PVColor.textFaint)
            if let place {
                Text(verbatim: place)
                    .font(PVFont.body(size: PVTypeScale.bodySmall, italic: true))
                    .foregroundStyle(PVColor.textSecondary)
            } else {
                Text(L10n.Conclusions.personPlaceUnknown)
                    .font(PVFont.body(size: PVTypeScale.bodySmall, italic: true))
                    .foregroundStyle(PVColor.textMuted)
            }
        }
        .accessibilityElement(children: .ignore)
        .accessibilityLabel(L10n.Conclusions.a11yList(
            date ?? L10n.string(L10n.Conclusions.personDateUnknown),
            rest: place ?? L10n.string(L10n.Conclusions.personPlaceUnknown)
        ))
    }
}
