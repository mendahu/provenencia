import Foundation
import Testing
@testable import Provenencia

@Suite
struct SubjectStoreTests {
    private let projectDir = "/tmp/subjects.provenencia"
    private let sourceID = "src-1"
    private let graphID = "g-1"
    private let typeID = "type-person"

    /// A fake whose project holds the Source's Default graph, under an id that is not the Source's.
    private func makeStore() -> FakeStore {
        let store = FakeStore()
        store.evidenceGraphsByProject[projectDir] = [
            CatalogEvidenceGraph(
                id: graphID, ref: "GPH-1", sourceID: sourceID, name: "Default", description: "", artifactIDs: []
            ),
        ]
        return store
    }

    @Test func createSubjectThenList() async throws {
        let store = makeStore()
        store.subjectTypesByProject[projectDir] = [
            CatalogSubjectType(
                id: typeID,
                key: "person",
                origin: "provenencia",
                label: "Person",
                description: "",
                refPrefix: "PER",
                candidateRefPrefix: "CPR"
            ),
        ]

        let (created, _) = try await store.createSubject(
            projectDir: projectDir,
            userID: "user-1",
            graphID: graphID,
            subjectTypeID: typeID,
            label: "Alice",
            description: "daughter",
            placement: nil
        )
        #expect(created.label == "Alice")
        #expect(created.sourceID == sourceID)
        #expect(store.heldCatalogProjectDir == projectDir)

        let listed = try await store.listSubjects(projectDir: projectDir, graphID: graphID)
        #expect(listed.count == 1)
        #expect(listed[0].id == created.id)
        #expect(listed[0].ref.hasPrefix("CPR-"))
    }

    @Test func setPositionPersistsAcrossSessionClose() async throws {
        let store = makeStore()
        let (subject, _) = try await store.createSubject(
            projectDir: projectDir,
            userID: "user-1",
            graphID: graphID,
            subjectTypeID: typeID,
            label: "Bob",
            description: "",
            placement: nil
        )

        let position = try await store.setSubjectPosition(
            projectDir: projectDir,
            subjectID: subject.id,
            gridX: 2,
            gridY: 4
        )
        #expect(position.gridX == 2)
        #expect(position.gridY == 4)

        try await store.closeCatalogSession(projectDir: projectDir)
        #expect(store.heldCatalogProjectDir == nil)
        #expect(store.lastClosedCatalogProjectDir == projectDir)

        let afterReopen = try await store.listSubjectPositions(projectDir: projectDir, graphID: graphID)
        #expect(afterReopen.count == 1)
        #expect(afterReopen[0].subjectID == subject.id)
        #expect(afterReopen[0].gridX == 2)
        #expect(afterReopen[0].gridY == 4)
        #expect(store.heldCatalogProjectDir == projectDir)
    }

    private func storeWithPersonType() -> FakeStore {
        let store = makeStore()
        store.subjectTypesByProject[projectDir] = [
            CatalogSubjectType(
                id: typeID,
                key: "person",
                origin: "provenencia",
                label: "Person",
                description: "",
                refPrefix: "PER",
                candidateRefPrefix: "CPR"
            ),
        ]
        return store
    }

    @Test func deletingPromotedSubjectNamesHandleThenClearsMembership() async throws {
        let store = storeWithPersonType()
        let (subject, _) = try await store.createSubject(
            projectDir: projectDir, userID: "user-1", graphID: graphID,
            subjectTypeID: typeID, label: "James", description: "", placement: nil
        )
        let result = store.seedPromotion(projectDir: projectDir, subjectID: subject.id)

        let report = try await store.getDeleteImpact(projectDir: projectDir, kind: "subject", id: subject.id)
        #expect(report.allowed)
        #expect(report.cascades.first?.via == "identity_claims.subject_id")
        #expect(report.cascades.first?.listed.map(\.ref) == [result.entity.ref])

        #expect(try await store.listSubjectMemberships(projectDir: projectDir, graphID: graphID).map(\.entity.ref)
            == [result.entity.ref])

        _ = try await store.deleteSubject(projectDir: projectDir, userID: "user-1", subjectID: subject.id)
        #expect(store.membershipBySubject[subject.id] == nil)
        #expect(try await store.listSubjectMemberships(projectDir: projectDir, graphID: graphID).isEmpty)
    }
}
