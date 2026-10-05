import SwiftUI

/// The Person page (S9-16, board S9-D5). Reads only the handle's
/// `.conclusionDetail` key, which the place warms on navigation; every field
/// explains itself with `ReconciledValueRow`. Read-only.
struct PersonDetailView: View {
    let session: WorkspaceSession
    let entityId: String

    private var key: CatalogQueryKey {
        .conclusionDetail(project: session.projectKey, entityId: entityId)
    }

    var body: some View {
        Group {
            if let handle: QueryHandle<CatalogConclusionDetail> = session.queryHandle(key) {
                PersonDetailPage(handle: handle)
            } else {
                PersonDetailPage.page { PersonDetailPage.loading }
            }
        }
        // A new Person starts with every disclosure closed.
        .id(entityId)
        .task(id: entityId) {
            let _: QueryHandle<CatalogConclusionDetail> = session.query(key)
        }
    }
}

private struct PersonDetailPage: View {
    @Bindable var handle: QueryHandle<CatalogConclusionDetail>

    private var presentation: PersonDetailPresentation {
        PersonDetailPresentation(
            value: handle.value, isFetching: handle.isFetching, status: handle.status, error: handle.error
        )
    }

    var body: some View {
        Self.page {
            switch presentation {
            case .loading:
                Self.loading
            case .failed(let message):
                PVCallout(tone: .danger, message: message)
                    .accessibilityIdentifier("person.detail.error")
            case .content(let content, let refreshing):
                PersonDetailHeader(content: content)
                VStack(alignment: .leading, spacing: 0) {
                    PVSectionHeader(title: L10n.Conclusions.personDetails) {
                        if refreshing {
                            ProgressView().controlSize(.small)
                        }
                    }
                    ForEach(content.rows) { row in
                        ReconciledValueRow(model: row)
                    }
                }
            }
        }
    }

    static var loading: some View {
        ProgressView()
            .controlSize(.small)
            .frame(maxWidth: .infinity, alignment: .leading)
    }

    /// Page chrome shared by every state, in the content column.
    static func page<Body: View>(@ViewBuilder body: () -> Body) -> some View {
        ScrollView {
            VStack(alignment: .leading, spacing: PVSpacing.space9) {
                body()
            }
            .frame(maxWidth: PVSpacing.widthContentMax, alignment: .leading)
            .padding(.horizontal, PVSpacing.gutterPage)
            .padding(.vertical, PVSpacing.space8)
            .frame(maxWidth: .infinity)
        }
        .background(PVColor.surfacePage)
        .accessibilityIdentifier("person.detail")
    }
}

/// Thumbnail slot · title and life shorthand · ref and member count.
private struct PersonDetailHeader: View {
    let content: PersonDetailContent

    var body: some View {
        HStack(alignment: .top, spacing: PVSpacing.space7) {
            PVThumbnail(.init(mark: .subjectPerson, label: L10n.Conclusions.personNoLikeness), size: 80)
            VStack(alignment: .leading, spacing: PVSpacing.space4) {
                title
                VStack(alignment: .leading, spacing: PVSpacing.space1) {
                    ForEach(content.vitals, id: \.abbreviation) { vital in
                        vitalLine(vital)
                    }
                }
            }
            .padding(.top, PVSpacing.space3)
            .frame(maxWidth: .infinity, alignment: .leading)
            VStack(alignment: .trailing, spacing: PVSpacing.space2) {
                if content.showsRef {
                    Text(content.ref)
                        .font(PVFont.mono(size: PVTypeScale.caption))
                        .foregroundStyle(PVColor.textSecondary)
                        .accessibilityLabel(L10n.Conclusions.a11yRef(content.ref))
                }
                Text(content.members)
                    .font(PVFont.mono(size: PVTypeScale.caption))
                    .foregroundStyle(PVColor.textMuted)
            }
            .padding(.top, PVSpacing.space6)
        }
    }

    @ViewBuilder
    private var title: some View {
        switch content.title {
        case .name(let text):
            Text(text)
                .font(PVFont.display(size: PVTypeScale.display3, weight: PVFontWeight.medium))
                .tracking(PVTypeScale.display3 * PVTracking.display)
                .foregroundStyle(PVColor.textDisplay)
                .accessibilityAddTraits(.isHeader)
        case .label(let text):
            Text(text)
                .font(PVFont.display(size: PVTypeScale.display3, weight: PVFontWeight.medium).italic())
                .foregroundStyle(PVColor.textDisplay)
                .accessibilityAddTraits(.isHeader)
        case .ref(let text):
            Text(text)
                .font(PVFont.mono(size: PVTypeScale.h1, weight: PVFontWeight.medium))
                .foregroundStyle(PVColor.textDisplay)
                .accessibilityAddTraits(.isHeader)
        }
    }

    private func vitalLine(_ vital: PersonDetailContent.Vital) -> some View {
        HStack(alignment: .firstTextBaseline, spacing: PVSpacing.space4) {
            Text(vital.abbreviation)
                .font(PVFont.body(size: PVTypeScale.bodySmall, italic: true))
                .foregroundStyle(PVColor.textMuted)
            if let date = vital.date {
                Text(date)
                    .font(PVFont.mono(size: PVTypeScale.caption))
                    .foregroundStyle(PVColor.textPrimary)
            } else {
                Text(L10n.Conclusions.personDateUnknown)
                    .font(PVFont.body(size: PVTypeScale.bodySmall, italic: true))
                    .foregroundStyle(PVColor.textMuted)
            }
            Text(verbatim: "·")
                .foregroundStyle(PVColor.textFaint)
            if let place = vital.place {
                Text(place)
                    .font(PVFont.body(size: PVTypeScale.bodySmall, italic: true))
                    .foregroundStyle(PVColor.textSecondary)
            } else {
                Text(L10n.Conclusions.personPlaceUnknown)
                    .font(PVFont.body(size: PVTypeScale.bodySmall, italic: true))
                    .foregroundStyle(PVColor.textMuted)
            }
        }
        .accessibilityElement(children: .ignore)
        .accessibilityLabel(vital.accessibilityLabel)
    }
}
