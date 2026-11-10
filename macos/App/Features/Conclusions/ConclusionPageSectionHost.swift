import SwiftUI

/// The host of a Person page section that reads from its list
/// (`ConclusionPageSection.reads`): it gets the handle of the section's `read` role
/// (its own read by default), starts the read, and draws the spinner, the error Callout
/// and the empty state around the section's `content`. Only the content differs between sections.
///
/// The id root is the section's `accessibilityIdentifierRoot(pageIdentifier:)`: `.error`,
/// `.empty`, and the root itself on the content. A second host over a section's companion
/// read draws in its own region: it passes `read: .companion` and
/// `placement: .column(loadingHeight:)`, and its id root gains `.column`.
struct ConclusionPageSectionHost<Value, Content: Equatable, Body: View>: View {
    /// Where the host draws its states.
    enum Placement {
        /// The section's own region: states keep the page band, an empty read has an EmptyState.
        case band
        /// A column beside the section's rows: the spinner centred in `loadingHeight` (the
        /// height the column draws at, so it does not jump), a compact danger Callout, and no
        /// empty state.
        case column(loadingHeight: CGFloat)
    }

    let section: ConclusionPageSection
    let entityId: String
    let accessibilityIdentifierRoot: String
    let read: ConclusionSectionRead.Role
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
        accessibilityIdentifierRoot: String,
        read: ConclusionSectionRead.Role = .own,
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
        self.accessibilityIdentifierRoot = accessibilityIdentifierRoot
        self.read = read
        self.placement = placement
        self.errorTitle = errorTitle
        self.emptyTitle = emptyTitle
        self.emptyMessage = emptyMessage
        self.make = make
        self.isEmpty = isEmpty
        self.content = body
    }

    private var key: CatalogQueryKey {
        guard let key = section.read(read, project: session.projectKey, entityId: entityId) else {
            preconditionFailure("\(section.rawValue) has no \(read) read")
        }
        return key
    }

    private var columnLoadingHeight: CGFloat? {
        if case .column(let height) = placement { return height }
        return nil
    }

    private var idRoot: String {
        accessibilityIdentifierRoot + (columnLoadingHeight == nil ? "" : ".column")
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
        case .column(let height):
            ProgressView()
                .frame(maxWidth: .infinity)
                .frame(height: height)
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
            if case .band = placement, let emptyTitle, isEmpty(value) {
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

extension ConclusionPageSection {
    /// The root of the identifiers inside this section's view: the page's identifier, then the
    /// section. The section computes it once and passes it to every view that builds an identifier.
    func accessibilityIdentifierRoot(pageIdentifier: String) -> String {
        "\(pageIdentifier).\(rawValue)"
    }
}
