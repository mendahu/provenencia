import Foundation
import Testing
@testable import Provenencia

@Suite
struct SearchCatalogStoreTests {
    private let projectDir = "/tmp/search.provenencia"

    @Test func emptyQueryReturnsNoHits() async throws {
        let store = FakeStore()
        store.sourcesByProject[projectDir] = [
            CatalogSource(id: "1", ref: "SRC-AAAAA", sourceTypeID: "t1", title: "Ilminster", description: ""),
        ]
        let hits = try await store.searchCatalog(
            projectDir: projectDir,
            query: "   ",
            location: .sectionRoot(.sources)
        )
        #expect(hits.isEmpty)
        #expect(store.heldCatalogProjectDir == projectDir)
    }

    @Test func findsSourceByTitleAndMapsLocation() async throws {
        let store = FakeStore()
        store.sourcesByProject[projectDir] = [
            CatalogSource(
                id: "src-1",
                ref: "SRC-ILMIN",
                sourceTypeID: "t1",
                title: "Ilminster parish register",
                description: "notes"
            ),
            CatalogSource(
                id: "src-2",
                ref: "SRC-OTHER",
                sourceTypeID: "t1",
                title: "Unrelated census",
                description: ""
            ),
        ]
        store.sourceTypesByProject[projectDir] = [
            CatalogSourceType(id: "t1", key: "parish", origin: "provenencia", label: "Parish register", description: ""),
        ]

        let hits = try await store.searchCatalog(
            projectDir: projectDir,
            query: "Ilminster",
            location: .sectionRoot(.sources)
        )
        #expect(hits.count == 1)
        let hit = try #require(hits.first)
        #expect(hit.kind == "source")
        #expect(hit.id == "src-1")
        #expect(hit.ref == "SRC-ILMIN")
        #expect(hit.location.section == .sources)
        #expect(hit.location.sourceId == "src-1")
        #expect(hit.location.ref == "SRC-ILMIN")
        #expect(hit.subtitle == "Parish register")
        #expect(store.heldCatalogProjectDir == projectDir)
    }

    @Test func findsTypeAndFieldByLabel() async throws {
        let store = FakeStore()
        store.sourceTypesByProject[projectDir] = [
            CatalogSourceType(
                id: "t1",
                key: "book",
                origin: "provenencia",
                label: "Published book",
                description: "",
                iconKey: "type_book"
            ),
        ]
        store.fieldsByProject[projectDir] = [
            CatalogMetadataField(
                id: "f1",
                key: "publication-date",
                origin: "provenencia",
                label: "Publication date",
                dataType: "date",
                description: ""
            ),
        ]

        let typeHits = try await store.searchCatalog(
            projectDir: projectDir,
            query: "Published book",
            location: .sectionRoot(.sourceTypes)
        )
        #expect(typeHits.contains {
            $0.kind == "source_type" && $0.location.typeId == "t1" && $0.iconKey == "type_book"
        })

        let fieldHits = try await store.searchCatalog(
            projectDir: projectDir,
            query: "Publication",
            location: .sectionRoot(.metadata)
        )
        #expect(fieldHits.contains { $0.kind == "metadata_field" && $0.location.fieldId == "f1" })
    }

    @Test func findsSourceByRef() async throws {
        let store = FakeStore()
        store.sourcesByProject[projectDir] = [
            CatalogSource(id: "src-1", ref: "SRC-ZZ9K2", sourceTypeID: "", title: "Quiet title", description: ""),
            CatalogSource(id: "src-2", ref: "SRC-OTHER", sourceTypeID: "", title: "SRC-ZZ9K2 in title", description: ""),
        ]
        let hits = try await store.searchCatalog(
            projectDir: projectDir,
            query: "SRC-ZZ9K2",
            location: .sectionRoot(.sources)
        )
        #expect(hits.first?.id == "src-1")
        #expect(hits.first?.matchReason == "ref")
    }

    @Test func prefixRefPromotesSource() async throws {
        let store = FakeStore()
        store.sourcesByProject[projectDir] = [
            CatalogSource(id: "src-1", ref: "SRC-ZZ9K2", sourceTypeID: "", title: "Quiet", description: ""),
            CatalogSource(id: "src-2", ref: "SRC-AAAAA", sourceTypeID: "", title: "Noise", description: ""),
        ]
        let hits = try await store.searchCatalog(
            projectDir: projectDir,
            query: "SRC-ZZ",
            location: .sectionRoot(.sources)
        )
        #expect(hits.first?.id == "src-1")
        #expect(hits.first?.matchReason == "ref")
    }

    @Test func personsRankAboveSourcesInTheDefaultOmnibar() async throws {
        let store = FakeStore()
        store.subjectTypesByProject[projectDir] = [
            CatalogSubjectType(
                id: "type-person", key: "person", origin: "provenencia", label: "Person",
                description: "", refPrefix: "PER", candidateRefPrefix: "CPR"
            ),
        ]
        store.sourcesByProject[projectDir] = [
            CatalogSource(id: "src-1", ref: "SRC-1", sourceTypeID: "", title: "Robins register", description: ""),
        ]
        store.subjectsBySource["s1"] = [
            CatalogSubject(id: "sub-1", ref: "CPR-1", sourceID: "s1", subjectTypeID: "type-person", label: "", description: ""),
            CatalogSubject(id: "sub-2", ref: "CPR-2", sourceID: "s1", subjectTypeID: "type-person", label: "", description: ""),
        ]
        store.observationsBySource["s1"] = [
            CatalogObservation(
                id: "o1", ref: "OBS-1", citationID: "c1", subjectID: "sub-1", propertyID: "p-name",
                polarity: "positive", valueText: "James Robins", valueInteger: nil, valueDateID: "",
                valueNameID: "n-1", nameForm: "James Robins", valueSubjectID: "", valueTermID: "",
                propertyKey: "name", propertyLabel: "Name", propertyValueType: "name"
            ),
        ]
        let james = try await store.promoteSubject(projectDir: projectDir, userID: "u", subjectID: "sub-1")
        _ = try await store.promoteSubject(
            projectDir: projectDir, userID: "u", subjectID: "sub-2",
            entityID: james.entity.id, confidenceGradeID: nil, argument: ""
        )

        let omnibar = try await store.searchCatalog(projectDir: projectDir, query: "Robins", location: .sectionRoot(.sources))
        #expect(omnibar.map(\.kind) == ["person", "source"])

        let hits = try await store.searchCatalog(
            projectDir: projectDir, query: "Robins", location: .sectionRoot(.sources), kinds: ["person"]
        )
        #expect(hits.map(\.ref) == [james.entity.ref])
        #expect(hits.first?.title == "James Robins")
        #expect(hits.first?.memberCount == 2)
        #expect(hits.first?.location == WorkspaceLocation(section: .persons, entityId: james.entity.id))
    }

    @Test func entityIdRoundTripsThroughTheEngineLocation() {
        let location = WorkspaceLocation(section: .persons, entityId: "e-1", ref: "PER-1", title: "James")
        var proto = Provenencia_Engine_V1_WorkspaceLocation()
        proto.section = "persons"
        proto.entityID = "e-1"
        #expect(GoStore.mapWorkspaceLocationFromProto(proto) == location)
    }
}
