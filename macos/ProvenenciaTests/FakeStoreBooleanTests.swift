import Foundation
import Testing
@testable import Provenencia

@Suite
struct FakeStoreBooleanTests {
    /// The fake keeps a recorded No as No, not as absent, the way the engine does.
    @Test func fakeStoreKeepsABooleanDraft() async throws {
        let store = FakeStore()
        store.subjectsByGraph["g-1"] = [
            CatalogSubject(
                id: "sub-1", ref: "CPR-1", sourceID: "src-1", subjectTypeID: "type-person",
                label: "Ada", description: ""
            ),
        ]
        store.citationsByID["cit-1"] = CatalogCitation(
            id: "cit-1", ref: "CIT-1", artifactID: "art-0", locatorJSON: "{}", transcription: "",
            description: "", transcriptionUncertain: false, transcriptionNote: ""
        )
        var draft = CatalogObservationDraft(subjectID: "sub-1", propertyID: "prop-literate")
        draft.valueBoolean = false
        let written = try await store.addObservationsToCitation(
            projectDir: "/p", userID: "user-1", citationID: "cit-1", observations: [draft]
        )
        let saved = try #require(written.0.first)
        #expect(saved.valueBoolean == false)
    }
}
