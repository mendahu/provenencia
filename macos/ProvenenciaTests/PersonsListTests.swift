import Foundation
import Testing
@testable import Provenencia

@MainActor
struct PersonsListTests {
    private func summary(_ ref: String, name: String? = nil, label: String = "") -> CatalogPersonSummary {
        CatalogPersonSummary(
            entity: CatalogCanonicalEntity(id: "id-\(ref)", ref: ref, subjectTypeID: "t", label: label),
            name: name.map { CatalogNameValue(form: $0) },
            nameValueCount: name == nil ? 0 : 1
        )
    }

    @Test func listThumbIsTheLikenessPath() {
        var withLikeness = summary("PER-1", name: "A")
        withLikeness.pictureImageRelPath = "objects/ab/cd.png"
        #expect(PersonsList.thumbnailRelPath(withLikeness) == "objects/ab/cd.png")
        #expect(PersonsList.thumbnailRelPath(summary("PER-2")) == "")
        let event = CatalogEventSummary(entity: CatalogCanonicalEntity(id: "e", ref: "EVT-1", subjectTypeID: "t", label: ""), title: CatalogEventTitle(rule: .ref, ref: "EVT-1"))
        #expect(EventsList.thumbnailRelPath(event) == "")
        var pictured = event
        pictured.pictureImageRelPath = "objects/12/34.png"
        #expect(EventsList.thumbnailRelPath(pictured) == "objects/12/34.png")
        let place = CatalogPlaceSummary(entity: CatalogCanonicalEntity(id: "p", ref: "PLC-1", subjectTypeID: "t", label: ""))
        #expect(PlacesList.thumbnailRelPath(place) == "")
        var picturedPlace = place
        picturedPlace.pictureImageRelPath = "objects/56/78.png"
        #expect(PlacesList.thumbnailRelPath(picturedPlace) == "objects/56/78.png")
    }

    private func listThumb<K: ConclusionListKind>(_: K.Type, _ row: K.Row) -> String {
        K.thumbnailRelPath(row)
    }

    @Test func listThumbDispatchesThroughTheKind() {
        var withLikeness = summary("PER-1", name: "A")
        withLikeness.pictureImageRelPath = "objects/ab/cd.png"
        #expect(listThumb(PersonsList.self, withLikeness) == "objects/ab/cd.png")
        let event = CatalogEventSummary(entity: CatalogCanonicalEntity(id: "e", ref: "EVT-1", subjectTypeID: "t", label: ""), title: CatalogEventTitle(rule: .ref, ref: "EVT-1"))
        #expect(listThumb(EventsList.self, event) == "")
    }

    @Test func titleSourceFallsBackNameThenLabelThenRef() {
        #expect(PersonTitleDisplay.titleSource(summary("PER-1", name: "James Robins", label: "Grandpa")) == .name("James Robins"))
        #expect(PersonTitleDisplay.titleSource(summary("PER-2", label: " Unknown father ")) == .label("Unknown father"))
        #expect(PersonTitleDisplay.titleSource(summary("PER-3")) == .ref("PER-3"))
        #expect(PersonTitleDisplay.title(summary("PER-3")) == "PER-3")
    }

    @Test func rowAccessibilityLabelReadsTitleThenRef() {
        let label = ConclusionListRow.accessibilityLabel(.name("James Robins"), ref: "PER-7KD45")
        #expect(label == "James Robins, PER-7KD45")
    }

    @Test func presentationFollowsTheHandleState() {
        let rows = [summary("PER-1", name: "A"), summary("PER-2", name: "B")]
        let none: QueryRead<[CatalogPersonSummary]> = .loading
        let empty: QueryRead<[CatalogPersonSummary]> = .loaded([])
        #expect(ConclusionListPresentation(none, isFetching: false) == .loading)
        #expect(ConclusionListPresentation(empty, isFetching: false) == .empty)
        #expect(ConclusionListPresentation(.loaded(rows), isFetching: false) == .rows(rows, refreshing: false))
        #expect(ConclusionListPresentation(.loaded(rows), isFetching: true) == .rows(rows, refreshing: true))
    }

    @Test func headerMetaCountsAndShowsRefreshing() {
        let one = [summary("PER-1", name: "A")]
        let two = one + [summary("PER-2")]
        #expect(ConclusionListPresentation.rows(one, refreshing: false).meta(
            count: L10n.Workspace.personCount, refreshing: L10n.Workspace.personCountRefreshing
        ) == "1 person")
        #expect(ConclusionListPresentation.rows(two, refreshing: false).meta(
            count: L10n.Workspace.personCount, refreshing: L10n.Workspace.personCountRefreshing
        ) == "2 persons")
        #expect(ConclusionListPresentation.rows(two, refreshing: true).meta(
            count: L10n.Workspace.personCount, refreshing: L10n.Workspace.personCountRefreshing
        ) == "2 persons · refreshing")
        #expect(ConclusionListPresentation<[CatalogPersonSummary]>.empty.meta(
            count: L10n.Workspace.personCount, refreshing: L10n.Workspace.personCountRefreshing
        ) == nil)
        #expect(ConclusionListPresentation<[CatalogPersonSummary]>.loading.meta(
            count: L10n.Workspace.personCount, refreshing: L10n.Workspace.personCountRefreshing
        ) == nil)
    }

    @Test func eventsHeaderMetaNamesTheDateSort() {
        let one = [CatalogEventSummary(entity: CatalogCanonicalEntity(id: "e1", ref: "EVT-1", subjectTypeID: "t", label: ""), title: CatalogEventTitle(rule: .ref, ref: "EVT-1"))]
        let two = one + [CatalogEventSummary(entity: CatalogCanonicalEntity(id: "e2", ref: "EVT-2", subjectTypeID: "t", label: ""), title: CatalogEventTitle(rule: .ref, ref: "EVT-2"))]
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
        ])
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

/// Each kind's list contract: identifier roots and query keys stay per kind.
struct ConclusionListKindTests {
    private let project = ProjectKey(projectDir: "/tmp/p.provenencia")

    @Test func eachKindKeepsItsOwnPlaceAndIdentifiers() {
        #expect(PersonsList.key(project: project) == .personsList(project: project))
        #expect(EventsList.key(project: project) == .eventsList(project: project))
        #expect(PlacesList.key(project: project) == .placesList(project: project))
        #expect([PersonsList.identifierRoot, EventsList.identifierRoot, PlacesList.identifierRoot] == ["persons", "events", "places"])
    }

    @Test func onlyPlacesBadgeNames() {
        let entity = CatalogCanonicalEntity(id: "e1", ref: "PER-1", subjectTypeID: "t", label: "")
        let person = CatalogPersonSummary(entity: entity, name: nil, nameValueCount: 0)
        #expect(PersonsList.extraCount(person) == 0)
        let place = CatalogPlaceSummary(entity: entity, names: ["Montréal", "Montreal"])
        #expect(PlacesList.extraCount(place) == 1)
    }

    @Test func rowLabelsSpeakThePlacesTheBadgeCounts() {
        let york = CatalogCanonicalEntity(id: "pl1", ref: "PLC-1", subjectTypeID: "t", label: "")
        let kingston = CatalogCanonicalEntity(id: "pl2", ref: "PLC-2", subjectTypeID: "t", label: "")
        let places = [
            CatalogChainedPlace(entity: york, names: ["York"]),
            CatalogChainedPlace(entity: kingston, names: ["Kingston"]),
        ]
        var person = CatalogPersonSummary(
            entity: CatalogCanonicalEntity(id: "e1", ref: "PER-1", subjectTypeID: "t", label: ""),
            name: CatalogNameValue(form: "James Robins"), nameValueCount: 1
        )
        person.birth.places = places
        #expect(PersonsList.accessibilityLabel(person).hasSuffix(L10n.Conclusions.morePlaces(count: 1)))
        person.birth.places = [places[0]]
        #expect(!PersonsList.accessibilityLabel(person).contains(L10n.Conclusions.morePlaces(count: 1)))

        let event = CatalogEventSummary(
            entity: CatalogCanonicalEntity(id: "ev1", ref: "EVT-1", subjectTypeID: "t", label: ""),
            places: places,
            title: CatalogEventTitle(rule: .typeAtPlace, ref: "EVT-1", typeLabel: "Fire", place: "York")
        )
        #expect(EventsList.accessibilityLabel(event).hasSuffix(L10n.Conclusions.morePlaces(count: 1)))
        #expect(L10n.Conclusions.morePlaces(count: 1) == "1 more place")
        #expect(L10n.Conclusions.morePlaces(count: 2) == "2 more places")
    }
}
