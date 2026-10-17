import Foundation
import Testing
@testable import Provenencia

@Suite
struct ObservationDepictionMappingTests {
    @Test func mapsSavedDepiction() {
        var proto = Provenencia_Engine_V1_Observation()
        proto.valueDepictionID = "v"
        proto.depiction.regionJson = "r"
        proto.depiction.imageRelPath = "objects/a.png"

        let mapped = GoStore.mapObservation(proto)
        #expect(mapped.valueDepictionID == "v")
        #expect(mapped.depiction == CatalogDepictionValue(regionJSON: "r", imageRelPath: "objects/a.png"))

        #expect(GoStore.mapObservation(Provenencia_Engine_V1_Observation()).depiction == nil)
    }

    @Test func draftSendsDepiction() {
        let draft = CatalogObservationDraft(
            subjectID: "s",
            propertyID: "p",
            depiction: CatalogDepictionValueInput(regionJSON: "r", imagePath: "/tmp/c.png")
        )
        let proto = GoStore.mapObservationDraft(draft)
        #expect(proto.hasDepiction)
        #expect(proto.depiction.regionJson == "r")
        #expect(proto.depiction.imagePath == "/tmp/c.png")
    }

    @Test func updateSendsInputBesideTheRow() {
        var row = CatalogObservation(
            id: "o", ref: "OBS-1", citationID: "c", subjectID: "s", propertyID: "p", polarity: "positive",
            valueText: "", valueInteger: nil, valueDateID: "", valueNameID: "", valueSubjectID: "",
            valueTermID: "", propertyKey: "likeness", propertyLabel: "Likeness", propertyValueType: "depiction"
        )
        row.valueDepictionID = "v"
        row.depiction = CatalogDepictionValue(regionJSON: "r", imageRelPath: "objects/a.png")

        let req = GoStore.updateObservationRequest(
            projectDir: "/tmp/p",
            userID: "u",
            observation: row,
            depictionInput: CatalogDepictionValueInput(regionJSON: "r2", imagePath: "/tmp/c.png")
        )
        #expect(req.hasDepictionInput)
        #expect(req.depictionInput.regionJson == "r2")
        #expect(req.depictionInput.imagePath == "/tmp/c.png")
        #expect(!req.observation.hasDepiction)
        #expect(req.observation.valueDepictionID == "v")
    }
}
