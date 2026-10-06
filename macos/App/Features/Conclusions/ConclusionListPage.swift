import SwiftUI

/// What a Conclusion list shows for its query handle. Persons, Events, and
/// Places share this; the count wording stays with the kind.
enum ConclusionListPresentation<Row: Equatable>: Equatable {
    /// First load: skeleton, no count, rows inert.
    case loading
    /// Loaded with nothing to list: no count, Promote explanation.
    case empty
    /// Rows, with the header meta; stale rows stay live while refreshing.
    case rows([Row], refreshing: Bool)
    /// The read failed with nothing to show.
    case failed(String)

    init(value: [Row]?, isFetching: Bool, status: QueryStatus, error: Error?) {
        if let value {
            self = value.isEmpty && !isFetching ? .empty : .rows(value, refreshing: isFetching)
        } else if status == .error, let error {
            self = .failed(L10n.Errors.message(for: error))
        } else {
            self = .loading
        }
    }

    /// Header meta from the kind's count copy, or none when there are no rows.
    func meta(count: (Int) -> String, refreshing: (Int) -> String) -> String? {
        guard case .rows(let rows, let isRefreshing) = self, !rows.isEmpty else { return nil }
        return isRefreshing ? refreshing(rows.count) : count(rows.count)
    }
}

/// The shared Conclude list page. Each kind stays its own place: the caller
/// passes that place's query key and the location a row opens. This view
/// does not choose a section or push history.
struct ConclusionListPage<Row: Identifiable & Equatable, Secondary: View>: View {
    let session: WorkspaceSession
    let key: CatalogQueryKey
    let title: LocalizedStringResource
    let pageAccessibilityIdentifier: String
    let emptyIcon: PVSymbol
    let emptyTitle: LocalizedStringResource
    let emptyMessage: LocalizedStringResource
    let emptyAccessibilityIdentifier: String
    let countMeta: (Int) -> String
    let refreshingMeta: (Int) -> String
    let mark: PVMarkKey
    let ref: (Row) -> String
    let rowAccessibilityIdentifier: (Row) -> String
    let titleSource: (Row) -> ConclusionTitleSource
    let accessibilityLabel: (Row) -> String
    let location: (Row) -> WorkspaceLocation
    let secondary: (Row) -> Secondary

    var body: some View {
        Group {
            if let handle: QueryHandle<[Row]> = session.queryHandle(key) {
                ConclusionListBody(
                    handle: handle,
                    title: title,
                    pageAccessibilityIdentifier: pageAccessibilityIdentifier,
                    emptyIcon: emptyIcon,
                    emptyTitle: emptyTitle,
                    emptyMessage: emptyMessage,
                    emptyAccessibilityIdentifier: emptyAccessibilityIdentifier,
                    countMeta: countMeta,
                    refreshingMeta: refreshingMeta,
                    mark: mark,
                    ref: ref,
                    rowAccessibilityIdentifier: rowAccessibilityIdentifier,
                    titleSource: titleSource,
                    accessibilityLabel: accessibilityLabel,
                    location: location,
                    secondary: secondary
                )
            } else {
                Self.page(title: title, meta: nil, accessibilityIdentifier: pageAccessibilityIdentifier) {
                    PVListSkeleton()
                }
            }
        }
        .task {
            let _: QueryHandle<[Row]> = session.query(key)
        }
    }

    /// Page chrome shared by every state: the section header over the body,
    /// in the content column.
    static func page<Body: View>(
        title: LocalizedStringResource,
        meta: String?,
        accessibilityIdentifier: String,
        @ViewBuilder body: () -> Body
    ) -> some View {
        ScrollView {
            VStack(alignment: .leading, spacing: PVSpacing.space6) {
                PVSectionHeader(title: title, meta: meta)
                body()
            }
            .frame(maxWidth: PVSpacing.widthContentMax, alignment: .leading)
            .padding(.horizontal, PVSpacing.gutterPage)
            .padding(.vertical, PVSpacing.space8)
            .frame(maxWidth: .infinity)
        }
        .accessibilityIdentifier(accessibilityIdentifier)
    }
}

private struct ConclusionListBody<Row: Identifiable & Equatable, Secondary: View>: View {
    @Bindable var handle: QueryHandle<[Row]>
    @Environment(WorkspaceNavigation.self) private var navigation

    let title: LocalizedStringResource
    let pageAccessibilityIdentifier: String
    let emptyIcon: PVSymbol
    let emptyTitle: LocalizedStringResource
    let emptyMessage: LocalizedStringResource
    let emptyAccessibilityIdentifier: String
    let countMeta: (Int) -> String
    let refreshingMeta: (Int) -> String
    let mark: PVMarkKey
    let ref: (Row) -> String
    let rowAccessibilityIdentifier: (Row) -> String
    let titleSource: (Row) -> ConclusionTitleSource
    let accessibilityLabel: (Row) -> String
    let location: (Row) -> WorkspaceLocation
    let secondary: (Row) -> Secondary

    private var presentation: ConclusionListPresentation<Row> {
        ConclusionListPresentation(
            value: handle.value, isFetching: handle.isFetching, status: handle.status, error: handle.error
        )
    }

    var body: some View {
        ConclusionListPage<Row, Secondary>.page(
            title: title,
            meta: presentation.meta(count: countMeta, refreshing: refreshingMeta),
            accessibilityIdentifier: pageAccessibilityIdentifier
        ) {
            switch presentation {
            case .loading:
                PVListSkeleton()
            case .empty:
                PVEmptyState(
                    icon: emptyIcon,
                    title: emptyTitle,
                    message: L10n.string(emptyMessage)
                )
                .accessibilityIdentifier(emptyAccessibilityIdentifier)
            case .failed(let message):
                PVCallout(tone: .danger, message: message)
            case .rows(let headers, _):
                PVList(
                    items: headers,
                    thumbnail: { _ in ConclusionListRow.thumbnail(mark: mark) },
                    meta: ref,
                    label: title,
                    itemAccessibilityLabel: accessibilityLabel,
                    itemAccessibilityIdentifier: rowAccessibilityIdentifier,
                    onActivate: { navigation.go(to: location($0)) },
                    primary: { ConclusionListRow.title(titleSource($0)) },
                    secondary: secondary
                )
            }
        }
    }
}
