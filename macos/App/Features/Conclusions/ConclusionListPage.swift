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
}

/// One Conclude list: what `ConclusionListPage` needs to know about a kind.
/// Persons, Events, and Places each conform once; the page owns the chrome,
/// the query handle, and navigation. Each kind stays its own place: its
/// query key and the location a row opens.
protocol ConclusionListKind {
    associatedtype Row: Identifiable & Equatable
    associatedtype Secondary: View

    static func key(project: ProjectKey) -> CatalogQueryKey
    static var title: LocalizedStringResource { get }
    /// Stable identifier root: `<root>.list`, `<root>.empty`, `<root>.row.<ref>`.
    static var identifierRoot: String { get }
    static var emptyIcon: PVSymbol { get }
    static var emptyTitle: LocalizedStringResource { get }
    static var emptyMessage: LocalizedStringResource { get }
    static func countMeta(_ count: Int) -> String
    static func refreshingMeta(_ count: Int) -> String
    static var mark: PVMarkKey { get }

    static func ref(_ row: Row) -> String
    static func titleSource(_ row: Row) -> ConclusionTitleSource
    /// Names beyond the one shown. Zero omits the +N badge.
    static func extraCount(_ row: Row) -> Int
    static func accessibilityLabel(_ row: Row) -> String
    static func location(_ row: Row) -> WorkspaceLocation
    @MainActor static func secondary(_ row: Row) -> Secondary
}

extension ConclusionListKind {
    static func extraCount(_: Row) -> Int { 0 }
}

/// The shared Conclude list page for one kind. This view does not choose a
/// section or push history.
struct ConclusionListPage<Kind: ConclusionListKind>: View {
    let session: WorkspaceSession

    private var key: CatalogQueryKey { Kind.key(project: session.projectKey) }

    var body: some View {
        Group {
            if let handle: QueryHandle<[Kind.Row]> = session.queryHandle(key) {
                ConclusionListBody<Kind>(handle: handle)
            } else {
                Self.page(meta: nil) {
                    PVListSkeleton()
                }
            }
        }
        .task {
            let _: QueryHandle<[Kind.Row]> = session.query(key)
        }
    }

    /// Page chrome shared by every state: the section header over the body,
    /// in the content column.
    static func page<Body: View>(meta: String?, @ViewBuilder body: () -> Body) -> some View {
        ScrollView {
            VStack(alignment: .leading, spacing: PVSpacing.space6) {
                PVSectionHeader(title: Kind.title, meta: meta)
                body()
            }
            .frame(maxWidth: PVSpacing.widthContentMax, alignment: .leading)
            .padding(.horizontal, PVSpacing.gutterPage)
            .padding(.vertical, PVSpacing.space8)
            .frame(maxWidth: .infinity)
        }
        .accessibilityIdentifier("\(Kind.identifierRoot).list")
    }
}

private struct ConclusionListBody<Kind: ConclusionListKind>: View {
    @Bindable var handle: QueryHandle<[Kind.Row]>
    @Environment(WorkspaceNavigation.self) private var navigation

    private var presentation: ConclusionListPresentation<Kind.Row> {
        ConclusionListPresentation(
            value: handle.value, isFetching: handle.isFetching, status: handle.status, error: handle.error
        )
    }

    private var headerMeta: String? {
        guard case .rows(let rows, let isRefreshing) = presentation, !rows.isEmpty else { return nil }
        return isRefreshing
            ? Kind.refreshingMeta(rows.count)
            : Kind.countMeta(rows.count)
    }

    var body: some View {
        ConclusionListPage<Kind>.page(meta: headerMeta) {
            switch presentation {
            case .loading:
                PVListSkeleton()
            case .empty:
                PVEmptyState(
                    icon: Kind.emptyIcon,
                    title: Kind.emptyTitle,
                    message: L10n.string(Kind.emptyMessage)
                )
                .accessibilityIdentifier("\(Kind.identifierRoot).empty")
            case .failed(let message):
                PVCallout(tone: .danger, message: message)
            case .rows(let rows, _):
                PVList(
                    // Go owns list order (conclusionheaders.sortByTitle); rows show as given.
                    items: rows,
                    thumbnail: { _ in ConclusionListRow.thumbnail(mark: Kind.mark) },
                    meta: { Kind.ref($0) },
                    label: Kind.title,
                    itemAccessibilityLabel: { Kind.accessibilityLabel($0) },
                    itemAccessibilityIdentifier: { "\(Kind.identifierRoot).row.\(Kind.ref($0))" },
                    onActivate: { navigation.go(to: Kind.location($0)) },
                    primary: { ConclusionListRow.title(Kind.titleSource($0), extraCount: Kind.extraCount($0)) },
                    secondary: { Kind.secondary($0) }
                )
            }
        }
    }
}
