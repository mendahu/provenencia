import SwiftUI

/// The host of a Person page section that reads its own query
/// (`ConclusionPageSection.queryKey`): it gets the section's handle, starts
/// the read, and draws the spinner, the error Callout and the empty state
/// around the section's `content`. Only the content differs between sections.
///
/// The id root is `persons.detail.<section>`: `.error`, `.empty`, and the root
/// itself on the content. A second host over a section's companion read
/// (`ConclusionPageSection.companionQueryKey`) draws in its own region: it passes
/// `queryKey` and `placement: .column`, and its id root gains `.column`.
struct ConclusionPageSectionHost<Value, Content: Equatable, Body: View>: View {
    /// Where the host draws its states.
    enum Placement {
        /// The section's own region: states keep the page band, an empty read has an EmptyState.
        case band
        /// A column beside the section's rows: the spinner centred in the column's height, a
        /// compact danger Callout, and no empty state.
        case column
    }

    /// The height the column's spinner is centred in (the diagram's smallest drawing).
    static var columnLoadingHeight: CGFloat { 232 }

    let section: ConclusionPageSection
    let entityId: String
    let queryKey: CatalogQueryKey?
    let placement: Placement
    let errorTitle: LocalizedStringResource
    let emptyTitle: LocalizedStringResource?
    let emptyMessage: String
    let make: (Value) -> Content
    let isEmpty: (Content) -> Bool
    let content: (Content) -> Body

    @Environment(WorkspaceSession.self) private var session

    init(
        section: ConclusionPageSection,
        entityId: String,
        queryKey: CatalogQueryKey? = nil,
        placement: Placement = .band,
        errorTitle: LocalizedStringResource,
        emptyTitle: LocalizedStringResource? = nil,
        emptyMessage: String = "",
        make: @escaping (Value) -> Content,
        isEmpty: @escaping (Content) -> Bool = { _ in false },
        @ViewBuilder body: @escaping (Content) -> Body
    ) {
        self.section = section
        self.entityId = entityId
        self.queryKey = queryKey
        self.placement = placement
        self.errorTitle = errorTitle
        self.emptyTitle = emptyTitle
        self.emptyMessage = emptyMessage
        self.make = make
        self.isEmpty = isEmpty
        self.content = body
    }

    private var key: CatalogQueryKey {
        guard let key = queryKey ?? section.queryKey(project: session.projectKey, entityId: entityId) else {
            preconditionFailure("\(section.rawValue) reads no query of its own")
        }
        return key
    }

    private var idRoot: String {
        "persons.detail.\(section.rawValue)" + (placement == .column ? ".column" : "")
    }

    var body: some View {
        Group {
            if let handle: QueryHandle<Value> = session.queryHandle(key) {
                loaded(handle)
            } else {
                loading
            }
        }
        .task {
            let _: QueryHandle<Value> = session.query(key)
        }
    }

    @ViewBuilder
    private var loading: some View {
        switch placement {
        case .band:
            ProgressView()
                .frame(maxWidth: .infinity)
                .pvKeepsPageBand()
                .padding(.top, 96)
        case .column:
            ProgressView()
                .frame(maxWidth: .infinity)
                .frame(height: Self.columnLoadingHeight)
        }
    }

    @ViewBuilder
    private func error(_ message: String) -> some View {
        switch placement {
        case .band:
            PVCallout(tone: .danger, title: errorTitle, message: message)
                .pvKeepsPageBand()
                .accessibilityIdentifier("\(idRoot).error")
        case .column:
            PVCallout(tone: .danger, title: errorTitle, message: message, compact: true)
                .accessibilityIdentifier("\(idRoot).error")
        }
    }

    @ViewBuilder
    private func loaded(_ handle: QueryHandle<Value>) -> some View {
        let presentation = ConclusionDetailPresentation<Content>(
            handle.read,
            isFetching: handle.isFetching,
            content: make
        )
        switch presentation {
        case .loading:
            loading
        case .gone:
            // The page above leaves when its own read reports the handle gone; a section whose
            // read fails first, under a page that still holds its value, shows why.
            error(L10n.string(L10n.Errors.conclusionDetailsNotFound))
        case .failed(let message):
            error(message)
        case .content(let value, _):
            if placement == .band, let emptyTitle, isEmpty(value) {
                PVEmptyState(glyph: section.glyph, title: emptyTitle, message: emptyMessage)
                    .pvKeepsPageBand()
                    .accessibilityIdentifier("\(idRoot).empty")
            } else {
                content(value)
                    .accessibilityIdentifier(idRoot)
            }
        }
    }
}
