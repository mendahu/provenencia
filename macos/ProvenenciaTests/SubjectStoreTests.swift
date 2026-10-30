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

    @Test func setPositionPersistsAcrossSessionCloseThenClear() async throws {
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

        try await store.clearSubjectPosition(projectDir: projectDir, subjectID: subject.id)
        let cleared = try await store.listSubjectPositions(projectDir: projectDir, graphID: graphID)
        #expect(cleared.isEmpty)
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

    @Test func promoteMintsHandleOnceThenRefuses() async throws {
        let store = storeWithPersonType()
        let (subject, _) = try await store.createSubject(
            projectDir: projectDir, userID: "user-1", graphID: graphID,
            subjectTypeID: typeID, label: "James", description: "", placement: nil
        )
        let result = try await store.promoteSubject(projectDir: projectDir, userID: "user-1", subjectID: subject.id)
        #expect(result.entity.ref.hasPrefix("PER-"))
        #expect(result.claim.status == "accepted")
        #expect(result.claim.entityID == result.entity.id)
        #expect(store.membershipBySubject[subject.id]?.entity == result.entity)
        #expect(store.membershipBySubject[subject.id]?.claimID == result.claim.id)

        await #expect(throws: CoreInvokeError.self) {
            _ = try await store.promoteSubject(projectDir: projectDir, userID: "user-1", subjectID: subject.id)
        }
    }

    @Test func promoteJoinsExistingHandle() async throws {
        let store = storeWithPersonType()
        store.subjectTypesByProject[projectDir]?.append(CatalogSubjectType(
            id: "type-event", key: "event", origin: "provenencia", label: "Event",
            description: "", refPrefix: "EVT", candidateRefPrefix: "CEV"
        ))
        func make(_ type: String) async throws -> CatalogSubject {
            try await store.createSubject(
                projectDir: projectDir, userID: "user-1", graphID: graphID,
                subjectTypeID: type, label: "", description: "", placement: nil
            ).0
        }
        let first = try await store.promoteSubject(projectDir: projectDir, userID: "user-1", subjectID: try await make(typeID).id)
        let second = try await make(typeID)
        let joined = try await store.promoteSubject(
            projectDir: projectDir, userID: "user-1", subjectID: second.id,
            entityID: first.entity.id
        )
        #expect(joined.entity == first.entity)
        #expect(joined.claim.status == "accepted")
        #expect(store.membershipBySubject[second.id]?.entity.ref == first.entity.ref)

        let event = try await store.promoteSubject(projectDir: projectDir, userID: "user-1", subjectID: try await make("type-event").id)
        let james = try await make(typeID)
        await #expect(throws: CoreInvokeError.self) {
            _ = try await store.promoteSubject(
                projectDir: projectDir, userID: "user-1", subjectID: james.id,
                entityID: event.entity.id
            )
        }
        #expect(store.membershipBySubject[james.id] == nil)
    }

    @Test func deletingPromotedSubjectNamesHandleThenClearsMembership() async throws {
        let store = storeWithPersonType()
        let (subject, _) = try await store.createSubject(
            projectDir: projectDir, userID: "user-1", graphID: graphID,
            subjectTypeID: typeID, label: "James", description: "", placement: nil
        )
        let result = try await store.promoteSubject(projectDir: projectDir, userID: "user-1", subjectID: subject.id)

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

    @Test func listSubjectTypesReturnsSeededRows() async throws {
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
        let types = try await store.listSubjectTypes(projectDir: projectDir)
        #expect(types.count == 1)
        #expect(types[0].key == "person")
        #expect(store.heldCatalogProjectDir == projectDir)
    }
}
