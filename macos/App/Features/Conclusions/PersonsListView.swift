import SwiftUI

/// The Persons list (S9-09, board S9-D2). Rows are kit `PVList` with
/// `ConclusionListRow` slots; a row opens the Person's page. Reads only the
/// `.personsList` key, which the place warms on navigation.
struct PersonsListView: View {
    let session: WorkspaceSession

    var body: some View {
        Group {
            if let handle: QueryHandle<[CatalogPersonHeader]> = session.queryHandle(
                .personsList(project: session.projectKey)
            ) {
                PersonsListContent(handle: handle)
            } else {
                PersonsListContent.page(meta: nil) { PVListSkeleton() }
            }
        }
        .task {
            let _: QueryHandle<[CatalogPersonHeader]> = session.query(.personsList(project: session.projectKey))
        }
    }
}

/// What the Persons page shows for a handle's state (S9-D2 frames 03, 05–07).
enum PersonsListPresentation: Equatable {
    /// First load: skeleton, no count, rows inert.
    case loading
    /// Loaded with nothing to list: no count, Promote explanation.
    case empty
    /// Rows, with the header meta; stale rows stay live while refreshing.
    case rows([CatalogPersonHeader], refreshing: Bool)
    /// The read failed with nothing to show.
    case failed(String)

    init(value: [CatalogPersonHeader]?, isFetching: Bool, status: QueryStatus, error: Error?) {
        if let value {
            self = value.isEmpty && !isFetching ? .empty : .rows(value, refreshing: isFetching)
        } else if status == .error, let error {
            self = .failed(L10n.Errors.message(for: error))
        } else {
            self = .loading
        }
    }

    /// Section header meta: "N persons", "N persons · refreshing", or none.
    var meta: String? {
        guard case .rows(let rows, let refreshing) = self, !rows.isEmpty else { return nil }
        return refreshing
            ? L10n.Workspace.personCountRefreshing(rows.count)
            : L10n.Workspace.personCount(rows.count)
    }
}

private struct PersonsListContent: View {
    @Bindable var handle: QueryHandle<[CatalogPersonHeader]>
    @Environment(WorkspaceNavigation.self) private var navigation

    private var presentation: PersonsListPresentation {
        PersonsListPresentation(
            value: handle.value, isFetching: handle.isFetching, status: handle.status, error: handle.error
        )
    }

    var body: some View {
        Self.page(meta: presentation.meta) {
            switch presentation {
            case .loading:
                PVListSkeleton()
            case .empty:
                PVEmptyState(
                    icon: .person,
                    title: L10n.Workspace.personsEmptyTitle,
                    message: L10n.string(L10n.Workspace.personsEmptyMessage)
                )
                .accessibilityIdentifier("persons.empty")
            case .failed(let message):
                PVCallout(tone: .danger, message: message)
            case .rows(let headers, _):
                PVList(
                    items: headers,
                    thumbnail: { _ in ConclusionListRow.thumbnail(mark: .subjectPerson) },
                    meta: { $0.entity.ref },
                    label: L10n.Workspace.personsTitle,
                    itemAccessibilityLabel: { header in
                        ConclusionListRow.accessibilityLabel(
                            PersonHeaderDisplay.titleSource(header), ref: header.entity.ref
                        )
                    },
                    itemAccessibilityIdentifier: { "persons.row.\($0.entity.ref)" },
                    onActivate: { navigation.go(to: Self.location(for: $0)) },
                    primary: { ConclusionListRow.title(PersonHeaderDisplay.titleSource($0)) }
                )
            }
        }
    }

    static func location(for header: CatalogPersonHeader) -> WorkspaceLocation {
        .personDetail(entityId: header.entity.id, ref: header.entity.ref, title: PersonHeaderDisplay.title(header))
    }

    /// Page chrome shared by every state: the section header over the body,
    /// in the content column.
    static func page<Body: View>(meta: String?, @ViewBuilder body: () -> Body) -> some View {
        ScrollView {
            VStack(alignment: .leading, spacing: PVSpacing.space6) {
                PVSectionHeader(title: L10n.Workspace.personsTitle, meta: meta)
                body()
            }
            .frame(maxWidth: PVSpacing.widthContentMax, alignment: .leading)
            .padding(.horizontal, PVSpacing.gutterPage)
            .padding(.vertical, PVSpacing.space8)
            .frame(maxWidth: .infinity)
        }
        .accessibilityIdentifier("persons.list")
    }
}
