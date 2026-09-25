import Foundation
import Testing
@testable import Provenencia

@Suite
struct SubjectStoreTests {
    private let projectDir = "/tmp/subjects.provenencia"
    private let sourceID = "src-1"
    private let typeID = "type-person"

    @Test func createSubjectThenList() async throws {
        let store = FakeStore()
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

        let created = try await store.createSubject(
            projectDir: projectDir,
            userID: "user-1",
            sourceID: sourceID,
            subjectTypeID: typeID,
            label: "Alice",
            description: "daughter",
            placement: nil
        )
        #expect(created.label == "Alice")
        #expect(created.sourceID == sourceID)
        #expect(store.heldCatalogProjectDir == projectDir)

        let listed = try await store.listSubjects(projectDir: projectDir, sourceID: sourceID)
        #expect(listed.count == 1)
        #expect(listed[0].id == created.id)
        #expect(listed[0].ref.hasPrefix("CPR-"))
    }

    @Test func setPositionPersistsAcrossSessionCloseThenClear() async throws {
        let store = FakeStore()
        let subject = try await store.createSubject(
            projectDir: projectDir,
            userID: "user-1",
            sourceID: sourceID,
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

        let afterReopen = try await store.listSubjectPositions(projectDir: projectDir, sourceID: sourceID)
        #expect(afterReopen.count == 1)
        #expect(afterReopen[0].subjectID == subject.id)
        #expect(afterReopen[0].gridX == 2)
        #expect(afterReopen[0].gridY == 4)
        #expect(store.heldCatalogProjectDir == projectDir)

        try await store.clearSubjectPosition(projectDir: projectDir, subjectID: subject.id)
        let cleared = try await store.listSubjectPositions(projectDir: projectDir, sourceID: sourceID)
        #expect(cleared.isEmpty)
    }

    @Test func listSubjectTypesReturnsSeededRows() async throws {
        let store = FakeStore()
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
