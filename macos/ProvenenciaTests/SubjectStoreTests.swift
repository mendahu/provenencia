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

    private func storeWithPersonType() -> FakeStore {
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
        return store
    }

    @Test func promoteMintsHandleOnceThenRefuses() async throws {
        let store = storeWithPersonType()
        let subject = try await store.createSubject(
            projectDir: projectDir, userID: "user-1", sourceID: sourceID,
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

    @Test func promoteJoinsExistingHandleWithGradeAndArgument() async throws {
        let store = storeWithPersonType()
        store.subjectTypesByProject[projectDir]?.append(CatalogSubjectType(
            id: "type-event", key: "event", origin: "provenencia", label: "Event",
            description: "", refPrefix: "EVT", candidateRefPrefix: "CEV"
        ))
        func make(_ type: String) async throws -> CatalogSubject {
            try await store.createSubject(
                projectDir: projectDir, userID: "user-1", sourceID: sourceID,
                subjectTypeID: type, label: "", description: "", placement: nil
            )
        }
        let first = try await store.promoteSubject(projectDir: projectDir, userID: "user-1", subjectID: try await make(typeID).id)
        let second = try await make(typeID)
        let joined = try await store.promoteSubject(
            projectDir: projectDir, userID: "user-1", subjectID: second.id,
            entityID: first.entity.id, confidenceGradeID: "cg-high", argument: " Same name. "
        )
        #expect(joined.entity == first.entity)
        #expect(joined.claim.confidenceGradeID == "cg-high" && joined.claim.argument == "Same name.")
        #expect(store.membershipBySubject[second.id]?.entity.ref == first.entity.ref)

        let event = try await store.promoteSubject(projectDir: projectDir, userID: "user-1", subjectID: try await make("type-event").id)
        let james = try await make(typeID)
        await #expect(throws: CoreInvokeError.self) {
            _ = try await store.promoteSubject(
                projectDir: projectDir, userID: "user-1", subjectID: james.id,
                entityID: event.entity.id, confidenceGradeID: nil, argument: ""
            )
        }
        #expect(store.membershipBySubject[james.id] == nil)

        let grades = try await store.listClaimConfidenceGrades(projectDir: projectDir)
        #expect(grades.map(\.key) == ["low_confidence", "moderate", "high_confidence"])
    }

    @Test func promoteTargetSuggestionsRankExactThenSharedWord() async throws {
        let store = storeWithPersonType()
        var subjects: [String: String] = [:]
        var observations: [CatalogObservation] = []
        for (key, form) in [("exact", "James Robins"), ("shared", "Mary Robins"), ("other", "Ada Lovelace"), ("me", "james robins.")] {
            let s = try await store.createSubject(
                projectDir: projectDir, userID: "user-1", sourceID: sourceID,
                subjectTypeID: typeID, label: "", description: "", placement: nil
            )
            subjects[key] = s.id
            observations.append(CatalogObservation(
                id: "o-\(key)", ref: "OBS-\(key)", citationID: "c1", subjectID: s.id, propertyID: "p-name",
                polarity: "positive", valueText: form, valueInteger: nil, valueDateID: "",
                valueNameID: "n-\(key)", nameForm: form, valueSubjectID: "", valueTermID: "",
                propertyKey: "name", propertyLabel: "Name", propertyValueType: "name"
            ))
        }
        store.observationsBySource[sourceID] = observations
        var refs: [String: String] = [:]
        for key in ["exact", "shared", "other"] {
            refs[key] = try await store.promoteSubject(projectDir: projectDir, userID: "user-1", subjectID: subjects[key]!).entity.ref
        }

        let me = try #require(subjects["me"])
        let got = try await store.listPromoteTargetSuggestions(projectDir: projectDir, subjectID: me, limit: 0)
        #expect(got.map(\.entity.ref) == [refs["exact"], refs["shared"]])
        let one = try await store.listPromoteTargetSuggestions(projectDir: projectDir, subjectID: me, limit: 1)
        #expect(one.map(\.entity.ref) == [refs["exact"]])
    }

    @Test func deletingPromotedSubjectNamesHandleThenClearsMembership() async throws {
        let store = storeWithPersonType()
        let subject = try await store.createSubject(
            projectDir: projectDir, userID: "user-1", sourceID: sourceID,
            subjectTypeID: typeID, label: "James", description: "", placement: nil
        )
        let result = try await store.promoteSubject(projectDir: projectDir, userID: "user-1", subjectID: subject.id)

        let report = try await store.getDeleteImpact(projectDir: projectDir, kind: "subject", id: subject.id)
        #expect(report.allowed)
        #expect(report.cascades.first?.via == "identity_claims.subject_id")
        #expect(report.cascades.first?.listed.map(\.ref) == [result.entity.ref])

        #expect(try await store.listSubjectMemberships(projectDir: projectDir, sourceID: sourceID).map(\.entity.ref)
            == [result.entity.ref])

        try await store.deleteSubject(projectDir: projectDir, userID: "user-1", subjectID: subject.id)
        #expect(store.membershipBySubject[subject.id] == nil)
        #expect(try await store.listSubjectMemberships(projectDir: projectDir, sourceID: sourceID).isEmpty)
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
