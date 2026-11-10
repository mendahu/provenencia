import SwiftUI

/// The host of a Person page section that reads from its list
/// (`ConclusionPageSection.reads`): it gets the handle of the section's `read` role
/// (its own read by default), starts the read, and draws the spinner, the read-failure
/// notice and the empty state around the section's `content`. Only the content differs between sections.
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
        /// height the column draws at, so it does not jump), a compact retry callout, and no
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
    @Environment(WorkspaceNavigation.self) private var navigation

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

    private var isColumn: Bool {
        if case .column = placement { return true }
        return false
    }

    private var idRoot: String {
        accessibilityIdentifierRoot + (isColumn ? ".column" : "")
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

    /// A notice keeps the page band's width in the band, and fills the column in a column.
    @ViewBuilder
    private func placed(_ notice: some View) -> some View {
        if isColumn { notice } else { notice.pvKeepsPageBand() }
    }

    /// A failed read's notice with Try again, which reads this host's own key again.
    private func retryNotice(
        _ message: String, id: String, handle: QueryHandle<Value>
    ) -> some View {
        placed(
            PVRetryCallout(
                title: errorTitle,
                message: message,
                handle: handle,
                compact: isColumn,
                id: id,
                onRetry: { let _: QueryHandle<Value> = session.reread(key) }
            )
            .accessibilityIdentifier(id)
        )
    }

    /// The read is gone (deleted or merged): the page's kind says so, with nothing to try again.
    private var goneNotice: some View {
        placed(
            PVCallout(
                tone: .danger,
                title: errorTitle,
                message: L10n.string(L10n.Conclusions.goneNotice(section: navigation.currentLocation.section)),
                compact: isColumn,
                announces: true
            )
            .accessibilityIdentifier("\(idRoot).error")
        )
    }

    @ViewBuilder
    private func loaded(_ handle: QueryHandle<Value>) -> some View {
        let presentation = ConclusionDetailPresentation<Content>(
            handle.reading,
            isFetching: handle.isFetching,
            content: make
        )
        switch presentation {
        case .loading:
            loading
        case .gone:
            // The page above leaves when its own read reports the handle gone; a section whose
            // read fails first, under a page that still holds its value, shows why.
            goneNotice
        case .failed(let message):
            retryNotice(message, id: "\(idRoot).error", handle: handle)
        case .content(let value, _, let refreshFailure):
            // One stack either way, so the section stays mounted when a notice comes and goes.
            VStack(alignment: .leading, spacing: isColumn ? PVSpacing.space4 : PVSpacing.space8) {
                if let refreshFailure {
                    retryNotice(
                        L10n.Conclusions.readNoticeStaleMessage(refreshFailure),
                        id: "\(idRoot).refreshError",
                        handle: handle
                    )
                }
                if !isColumn, let emptyTitle, isEmpty(value) {
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
}

extension ConclusionPageSection {
    /// The root of the identifiers inside this section's view: the page's identifier, then the
    /// section. The section computes it once and passes it to every view that builds an identifier.
    func accessibilityIdentifierRoot(pageIdentifier: String) -> String {
        "\(pageIdentifier).\(rawValue)"
    }
}
