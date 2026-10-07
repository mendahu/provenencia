import SwiftUI

/// What a Conclusion page shows for its handle's state. A stale page stays
/// readable while it refreshes. Person, Event, and Place each supply
/// their own content; the page does not know which kind it is.
enum ConclusionDetailPresentation<Content: Equatable>: Equatable {
    case loading
    case failed(String)
    case content(Content, refreshing: Bool)

    init(
        value: CatalogConclusionDetail?,
        isFetching: Bool,
        status: QueryStatus,
        error: Error?,
        content: (CatalogConclusionDetail) -> Content
    ) {
        if let value {
            self = .content(content(value), refreshing: isFetching)
        } else if status == .error, let error {
            self = .failed(L10n.Errors.message(for: error))
        } else {
            self = .loading
        }
    }
}

/// One section under Details. Empty until a later slice fills its rows.
/// The kind names it; the page draws it.
struct ConclusionDetailSection: Equatable, Identifiable {
    var id: String
    var title: LocalizedStringResource
    var aside: String
    var emptyText: String
}

/// The fields a Conclusion detail page lays out. Each kind owns how it
/// fills them (mark, title, the line under the title, and the rows). The
/// page only places them. Giving a kind its own page later does not need a
/// new query or history entry.
protocol ConclusionDetailBody: Equatable {
    var title: ConclusionTitleSource { get }
    var ref: String { get }
    var showsRef: Bool { get }
    var members: String { get }
    var rows: [ReconciledValueRowModel] { get }
    var sections: [ConclusionDetailSection] { get }
}

extension ConclusionDetailBody {
    var sections: [ConclusionDetailSection] { [] }
}

/// Shared Conclusion detail chrome (S9-24). A view, not a place: the caller
/// names the entity, and the place registry warms `conclusionDetail` for
/// that id. Person, Event, and Place stay separate history entries.
struct ConclusionDetailPage<Content: ConclusionDetailBody, Summary: View>: View {
    let session: WorkspaceSession
    let entityId: String
    let mark: PVMarkKey
    let noLikeness: LocalizedStringResource
    let pageIdentifier: String
    let errorIdentifier: String
    let make: (CatalogConclusionDetail) -> Content
    let summary: (Content) -> Summary

    private var key: CatalogQueryKey {
        .conclusionDetail(project: session.projectKey, entityId: entityId)
    }

    var body: some View {
        Group {
            if let handle: QueryHandle<CatalogConclusionDetail> = session.queryHandle(key) {
                ConclusionDetailLoaded(
                    handle: handle,
                    mark: mark,
                    noLikeness: noLikeness,
                    pageIdentifier: pageIdentifier,
                    errorIdentifier: errorIdentifier,
                    make: make,
                    summary: summary
                )
            } else {
                Self.page(identifier: pageIdentifier) { Self.loading }
            }
        }
        // A new handle starts with every disclosure closed.
        .id(entityId)
        .task(id: entityId) {
            let _: QueryHandle<CatalogConclusionDetail> = session.query(key)
        }
    }

    static var loading: some View {
        ProgressView()
            .controlSize(.small)
            .frame(maxWidth: .infinity, alignment: .leading)
    }

    /// Page chrome shared by every state, in the content column.
    static func page<Body: View>(identifier: String, @ViewBuilder body: () -> Body) -> some View {
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
        .accessibilityIdentifier(identifier)
    }
}

private struct ConclusionDetailLoaded<Content: ConclusionDetailBody, Summary: View>: View {
    @Bindable var handle: QueryHandle<CatalogConclusionDetail>
    let mark: PVMarkKey
    let noLikeness: LocalizedStringResource
    let pageIdentifier: String
    let errorIdentifier: String
    let make: (CatalogConclusionDetail) -> Content
    let summary: (Content) -> Summary

    private var presentation: ConclusionDetailPresentation<Content> {
        ConclusionDetailPresentation(
            value: handle.value,
            isFetching: handle.isFetching,
            status: handle.status,
            error: handle.error,
            content: make
        )
    }

    var body: some View {
        ConclusionDetailPage<Content, Summary>.page(identifier: pageIdentifier) {
            switch presentation {
            case .loading:
                ConclusionDetailPage<Content, Summary>.loading
            case .failed(let message):
                PVCallout(tone: .danger, message: message)
                    .accessibilityIdentifier(errorIdentifier)
            case .content(let content, let refreshing):
                ConclusionDetailHeader(
                    content: content,
                    mark: mark,
                    noLikeness: noLikeness,
                    summary: summary
                )
                VStack(alignment: .leading, spacing: 0) {
                    PVSectionHeader(title: L10n.Conclusions.personDetails, aside: {
                        if refreshing {
                            ProgressView().controlSize(.small)
                        }
                    })
                    ForEach(content.rows) { row in
                        ReconciledValueRow(model: row)
                    }
                }
                ForEach(content.sections) { section in
                    ConclusionDetailEmptySection(section: section)
                }
            }
        }
    }
}

/// A relationship section with its header and an empty sentence. S9-40 fills the rows.
private struct ConclusionDetailEmptySection: View {
    let section: ConclusionDetailSection

    var body: some View {
        VStack(alignment: .leading, spacing: 0) {
            PVSectionHeader(title: section.title, aside: {
                Text(verbatim: section.aside)
                    .font(PVFont.body(size: PVTypeScale.caption, italic: true))
                    .foregroundStyle(PVColor.textMuted)
                    .lineLimit(2)
            })
            Text(verbatim: section.emptyText)
                .font(PVFont.body(size: PVTypeScale.bodySmall, italic: true))
                .foregroundStyle(PVColor.textMuted)
                .frame(maxWidth: .infinity, alignment: .leading)
                .padding(.vertical, PVSpacing.space5)
        }
        .accessibilityElement(children: .combine)
        .accessibilityLabel(L10n.Conclusions.a11yList(L10n.string(section.title), rest: section.emptyText))
    }
}

/// Thumbnail slot · title and the kind's summary · ref and member count.
private struct ConclusionDetailHeader<Content: ConclusionDetailBody, Summary: View>: View {
    let content: Content
    let mark: PVMarkKey
    let noLikeness: LocalizedStringResource
    let summary: (Content) -> Summary

    var body: some View {
        HStack(alignment: .top, spacing: PVSpacing.space7) {
            PVThumbnail(.init(mark: mark, label: noLikeness), size: 80)
            VStack(alignment: .leading, spacing: PVSpacing.space4) {
                title
                summary(content)
            }
            .padding(.top, PVSpacing.space3)
            .frame(maxWidth: .infinity, alignment: .leading)
            VStack(alignment: .trailing, spacing: PVSpacing.space2) {
                if content.showsRef {
                    Text(verbatim: content.ref)
                        .font(PVFont.mono(size: PVTypeScale.caption))
                        .foregroundStyle(PVColor.textSecondary)
                        .accessibilityLabel(L10n.Conclusions.a11yRef(content.ref))
                }
                Text(verbatim: content.members)
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
            Text(verbatim: text)
                .font(PVFont.display(size: PVTypeScale.display3, weight: PVFontWeight.medium))
                .tracking(PVTypeScale.display3 * PVTracking.display)
                .foregroundStyle(PVColor.textDisplay)
                .accessibilityAddTraits(.isHeader)
        case .label(let text):
            Text(verbatim: text)
                .font(PVFont.display(size: PVTypeScale.display3, weight: PVFontWeight.medium).italic())
                .foregroundStyle(PVColor.textDisplay)
                .accessibilityAddTraits(.isHeader)
        case .ref(let text):
            Text(verbatim: text)
                .font(PVFont.mono(size: PVTypeScale.h1, weight: PVFontWeight.medium))
                .foregroundStyle(PVColor.textDisplay)
                .accessibilityAddTraits(.isHeader)
        }
    }
}
