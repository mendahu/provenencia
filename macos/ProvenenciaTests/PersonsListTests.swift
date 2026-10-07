import Foundation
import Testing
@testable import Provenencia

@MainActor
struct PersonsListTests {
    private func header(_ ref: String, name: String? = nil, label: String = "") -> CatalogPersonHeader {
        CatalogPersonHeader(
            entity: CatalogCanonicalEntity(id: "id-\(ref)", ref: ref, subjectTypeID: "t", label: label),
            name: name.map { CatalogNameValue(form: $0) },
            nameValueCount: name == nil ? 0 : 1
        )
    }

    @Test func titleSourceFallsBackNameThenLabelThenRef() {
        #expect(PersonHeaderDisplay.titleSource(header("PER-1", name: "James Robins", label: "Grandpa")) == .name("James Robins"))
        #expect(PersonHeaderDisplay.titleSource(header("PER-2", label: " Unknown father ")) == .label("Unknown father"))
        #expect(PersonHeaderDisplay.titleSource(header("PER-3")) == .ref("PER-3"))
        #expect(PersonHeaderDisplay.title(header("PER-3")) == "PER-3")
    }

    @Test func rowAccessibilityLabelReadsTitleThenRef() {
        let label = ConclusionListRow.accessibilityLabel(.name("James Robins"), ref: "PER-7KD45")
        #expect(label == "James Robins, PER-7KD45")
    }

    @Test func presentationFollowsTheHandleState() {
        let rows = [header("PER-1", name: "A"), header("PER-2", name: "B")]
        let none: [CatalogPersonHeader]? = nil
        let empty: [CatalogPersonHeader] = []
        #expect(ConclusionListPresentation(value: none, isFetching: false, status: .loading, error: nil) == .loading)
        #expect(ConclusionListPresentation(value: empty, isFetching: false, status: .ready, error: nil) == .empty)
        #expect(ConclusionListPresentation(value: rows, isFetching: false, status: .ready, error: nil) == .rows(rows, refreshing: false))
        #expect(ConclusionListPresentation(value: rows, isFetching: true, status: .ready, error: nil) == .rows(rows, refreshing: true))
    }

    @Test func headerMetaCountsAndShowsRefreshing() {
        let one = [header("PER-1", name: "A")]
        let two = one + [header("PER-2")]
        #expect(ConclusionListPresentation.rows(one, refreshing: false).meta(
            count: L10n.Workspace.personCount, refreshing: L10n.Workspace.personCountRefreshing
        ) == "1 person")
        #expect(ConclusionListPresentation.rows(two, refreshing: false).meta(
            count: L10n.Workspace.personCount, refreshing: L10n.Workspace.personCountRefreshing
        ) == "2 persons")
        #expect(ConclusionListPresentation.rows(two, refreshing: true).meta(
            count: L10n.Workspace.personCount, refreshing: L10n.Workspace.personCountRefreshing
        ) == "2 persons · refreshing")
        #expect(ConclusionListPresentation<[CatalogPersonHeader]>.empty.meta(
            count: L10n.Workspace.personCount, refreshing: L10n.Workspace.personCountRefreshing
        ) == nil)
        #expect(ConclusionListPresentation<[CatalogPersonHeader]>.loading.meta(
            count: L10n.Workspace.personCount, refreshing: L10n.Workspace.personCountRefreshing
        ) == nil)
    }

    @Test func eventsHeaderMetaNamesTheDateSort() {
        let one = [CatalogEventHeader(entity: CatalogCanonicalEntity(id: "e1", ref: "EVT-1", subjectTypeID: "t", label: ""), title: CatalogEventTitle(rule: .ref, ref: "EVT-1"))]
        let two = one + [CatalogEventHeader(entity: CatalogCanonicalEntity(id: "e2", ref: "EVT-2", subjectTypeID: "t", label: ""), title: CatalogEventTitle(rule: .ref, ref: "EVT-2"))]
        #expect(ConclusionListPresentation.rows(one, refreshing: false).meta(
            count: L10n.Workspace.eventCount, refreshing: L10n.Workspace.eventCountRefreshing
        ) == "1 event · by date")
        #expect(ConclusionListPresentation.rows(two, refreshing: false).meta(
            count: L10n.Workspace.eventCount, refreshing: L10n.Workspace.eventCountRefreshing
        ) == "2 events · by date")
        #expect(ConclusionListPresentation.rows(two, refreshing: true).meta(
            count: L10n.Workspace.eventCount, refreshing: L10n.Workspace.eventCountRefreshing
        ) == "2 events · by date · refreshing")
    }

    @Test func rowOpensThePersonDetailPlace() {
        let location = WorkspaceLocation.personDetail(entityId: "id-PER-1", ref: "PER-1", title: "James Robins")
        #expect(location.section == .persons)
        #expect(location.entityId == "id-PER-1")
        let place = PlaceRegistry.standard.resolve(location, project: ProjectKey(projectDir: "/tmp/p.provenencia"))
        #expect(place?.placeID == .personDetail)
        #expect(place?.queryKeys == [
            .conclusionDetail(project: ProjectKey(projectDir: "/tmp/p.provenencia"), entityId: "id-PER-1"),
            .personHeader(project: ProjectKey(projectDir: "/tmp/p.provenencia"), entityId: "id-PER-1"),
        ])
        #expect(place?.deepId == "id-PER-1")
        let list = PlaceRegistry.standard.resolve(.sectionRoot(.persons), project: ProjectKey(projectDir: "/tmp/p.provenencia"))
        #expect(list?.placeID == .personsList)
        #expect(list?.queryKeys == [.personsList(project: ProjectKey(projectDir: "/tmp/p.provenencia"))])
    }

    @Test func personDetailRoundTripsThroughHistoryAndOldHistoryStillDecodes() throws {
        let location = WorkspaceLocation.personDetail(entityId: "id-PER-1", ref: "PER-1", title: "James Robins")
        let data = try JSONEncoder().encode(location)
        let decoded = try JSONDecoder().decode(WorkspaceLocation.self, from: data)
        #expect(decoded == location)
        #expect(decoded.entityId == "id-PER-1")
        let old = try JSONDecoder().decode(WorkspaceLocation.self, from: Data(#"{"section":"persons"}"#.utf8))
        #expect(old.entityId == nil)
        #expect(old != location)
    }
}
