import Foundation
import Testing
@testable import Provenencia

@Suite(.serialized)
@MainActor
struct CitationComposerModelTests {
    private let projectDir = "/tmp/citation-composer-model.provenencia"
    private let sourceID = "src-1"
    private let subjectID = "sub-person-1"
    private let personTypeID = "type-person"
    private let participationTypeID = "type-participation"
    private let namePropertyID = "prop-name"
    private let occupationPropertyID = "prop-occupation"
    private let personEdgePropertyID = "prop-person-edge"
    private let eventEdgePropertyID = "prop-event-edge"

    private func makeStore() -> FakeStore {
        let store = FakeStore()
        store.subjectTypesByProject[projectDir] = [
            CatalogSubjectType(
                id: personTypeID,
                key: "person",
                origin: "provenencia",
                label: "Person",
                description: "",
                refPrefix: "PER",
                candidateRefPrefix: "CPR"
            ),
            CatalogSubjectType(
                id: participationTypeID,
                key: "participation",
                origin: "provenencia",
                label: "Participation",
                description: "",
                refPrefix: "PTN",
                candidateRefPrefix: "CPA"
            ),
        ]
        store.propertiesByProject[projectDir] = [
            CatalogProperty(
                id: namePropertyID,
                key: "name",
                origin: "provenencia",
                label: "Name",
                description: "",
                valueType: "name"
            ),
            CatalogProperty(
                id: occupationPropertyID,
                key: "occupation",
                origin: "provenencia",
                label: "Occupation",
                description: "",
                valueType: "text"
            ),
            CatalogProperty(
                id: personEdgePropertyID,
                key: "person",
                origin: "provenencia",
                label: "Person",
                description: "",
                valueType: "subject"
            ),
            CatalogProperty(
                id: eventEdgePropertyID,
                key: "event",
                origin: "provenencia",
                label: "Event",
                description: "",
                valueType: "subject"
            ),
            CatalogProperty(
                id: "prop-age",
                key: "age",
                origin: "user",
                label: "Age",
                description: "",
                valueType: "integer"
            ),
        ]
        store.subjectTypeFieldsByType[participationTypeID] = [
            CatalogSubjectTypeField(
                property: store.propertiesByProject[projectDir]![2],
                sortOrder: 0,
                locked: true
            ),
            CatalogSubjectTypeField(
                property: store.propertiesByProject[projectDir]![3],
                sortOrder: 1,
                locked: true
            ),
            CatalogSubjectTypeField(
                property: store.propertiesByProject[projectDir]![1],
                sortOrder: 2,
                locked: false
            ),
        ]
        store.subjectTypeFieldsByType[personTypeID] = [
            CatalogSubjectTypeField(
                property: store.propertiesByProject[projectDir]![0],
                sortOrder: 0,
                locked: false
            ),
            CatalogSubjectTypeField(
                property: store.propertiesByProject[projectDir]![1],
                sortOrder: 1,
                locked: false
            ),
            CatalogSubjectTypeField(
                property: store.propertiesByProject[projectDir]![4],
                sortOrder: 2,
                locked: false
            ),
        ]
        store.connectRules = [
            CatalogConnectRule(
                fromTypeKey: "person",
                toTypeKey: "event",
                bridgeTypeKey: "participation",
                edgePropertyKeys: ["person", "event"],
                disambiguation: "role",
                refuse: false
            ),
        ]
        store.sourcesByProject[projectDir] = [
            CatalogSource(
                id: sourceID,
                ref: "SRC-1",
                sourceTypeID: "stype",
                title: "Census",
                description: ""
            ),
        ]
        let person = CatalogSubject(
            id: subjectID,
            ref: "CPR-1",
            sourceID: sourceID,
            subjectTypeID: personTypeID,
            label: "Margt.",
            description: ""
        )
        store.subjectsBySource[sourceID] = [person]
        store.subjectPositionsBySubject[subjectID] = CatalogSubjectPosition(
            subjectID: subjectID,
            gridX: 0,
            gridY: 0
        )
        return store
    }

    private func seedArtifact(_ store: FakeStore, count: Int = 1) {
        store.artifactsBySource[sourceID] = (0..<count).map { i in
            CatalogArtifact(
                id: "art-\(i)",
                ref: "ART-\(i)",
                sourceID: sourceID,
                fileID: "file-\(i)",
                label: "Scan \(i + 1)",
                description: "",
                file: CatalogFileRef(
                    id: "file-\(i)",
                    relPath: "objects/file-\(i).pdf",
                    originalFilename: "scan\(i).pdf",
                    mediaType: "application/pdf",
                    byteSize: 10
                )
            )
        }
    }

    private func makeModel(store: FakeStore, subjectID: String? = nil) -> CitationComposerModel {
        let session = WorkspaceSession(
            projectKey: ProjectKey(projectDir: projectDir),
            store: store
        )
        return CitationComposerModel(
            sourceID: sourceID,
            subjectID: subjectID ?? self.subjectID,
            session: session,
            store: store,
            userID: "user-1"
        )
    }

    @Test func excludedEdgePropertyKeysForBridgeType() {
        let keys = CitationComposerModel.excludedEdgePropertyKeys(
            typeKey: "participation",
            rules: [
                CatalogConnectRule(
                    fromTypeKey: "person",
                    toTypeKey: "event",
                    bridgeTypeKey: "participation",
                    edgePropertyKeys: ["person", "event"],
                    disambiguation: "role",
                    refuse: false
                ),
            ]
        )
        #expect(keys == Set(["person", "event"]))
    }

    @Test func prepareAutoSelectsSingleArtifactAndFiltersNameType() async {
        let store = makeStore()
        seedArtifact(store, count: 1)
        let model = makeModel(store: store)
        await model.prepare()
        #expect(model.phase == .compose)
        #expect(model.selectedArtifactID == "art-0")
        #expect(model.availableProperties.map(\.id).sorted() == ["prop-age", occupationPropertyID].sorted())
        #expect(!model.availableProperties.contains(where: { $0.valueType == "name" }))
    }

    @Test func prepareShowsPickerWhenMultipleArtifacts() async {
        let store = makeStore()
        seedArtifact(store, count: 2)
        let model = makeModel(store: store)
        await model.prepare()
        #expect(model.phase == .pickArtifact)
        model.selectArtifact("art-1")
        #expect(model.phase == .compose)
        #expect(model.selectedArtifactID == "art-1")
    }

    @Test func prepareExcludesConnectEdgePropertiesOnBridge() async {
        let store = makeStore()
        seedArtifact(store)
        let bridgeID = "sub-bridge-1"
        store.subjectsBySource[sourceID]?.append(
            CatalogSubject(
                id: bridgeID,
                ref: "CPA-1",
                sourceID: sourceID,
                subjectTypeID: participationTypeID,
                label: "Participation",
                description: ""
            )
        )
        store.subjectPositionsBySubject[bridgeID] = CatalogSubjectPosition(
            subjectID: bridgeID,
            gridX: 1,
            gridY: 0
        )
        let model = makeModel(store: store, subjectID: bridgeID)
        await model.prepare()
        #expect(model.phase == .compose)
        #expect(model.availableProperties.map(\.key) == ["occupation"])
        #expect(!model.availableProperties.contains(where: { $0.key == "person" }))
    }

    @Test func submitRequiresObservationValueThenWrites() async throws {
        let store = makeStore()
        seedArtifact(store)
        let model = makeModel(store: store)
        await model.prepare()
        #expect(model.phase == .compose)

        let empty = await model.submit()
        #expect(empty == nil)
        #expect(model.formError != nil)

        var row = model.observations[0]
        row.propertyID = occupationPropertyID
        row.valueText = "Farmer"
        model.updateObservation(row)

        let location = await model.submit()
        #expect(location?.sourceSurface == .graph)
        #expect(location?.sourceId == sourceID)
        let listed = try await store.listObservationsBySource(
            projectDir: projectDir,
            sourceID: sourceID
        )
        #expect(listed.count == 1)
        #expect(listed[0].valueText == "Farmer")
        #expect(listed[0].propertyID == occupationPropertyID)
        #expect(listed[0].subjectID == subjectID)
    }

    @Test func missingSubjectFallsBack() async {
        let store = makeStore()
        seedArtifact(store)
        let model = makeModel(store: store, subjectID: "missing")
        await model.prepare()
        #expect(model.phase == .subjectMissing)
        #expect(model.shouldFallbackToGraph)
    }
}
