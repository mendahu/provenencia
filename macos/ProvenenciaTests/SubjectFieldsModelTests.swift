import Foundation
import Testing
@testable import Provenencia

@Suite
@MainActor
struct SubjectFieldsModelTests {
    private let projectDir = "/tmp/subject-fields.provenencia"
    private let userID = "00000000-0000-7000-8000-000000000001"

    private func personType() -> CatalogSubjectType {
        CatalogSubjectType(
            id: "type-person",
            key: "person",
            origin: "provenencia",
            label: "Person",
            description: "",
            refPrefix: "PER",
            candidateRefPrefix: "CPR"
        )
    }

    private func eventType() -> CatalogSubjectType {
        CatalogSubjectType(
            id: "type-event",
            key: "event",
            origin: "provenencia",
            label: "Event",
            description: "",
            refPrefix: "EVT",
            candidateRefPrefix: "CEV"
        )
    }

    private func nameProperty() -> CatalogProperty {
        CatalogProperty(
            id: "prop-name",
            key: "name",
            origin: "provenencia",
            label: "Name",
            description: "Display name",
            valueType: "name",
            usedBy: 1
        )
    }

    private func eventTypeProperty() -> CatalogProperty {
        CatalogProperty(
            id: "prop-event-type",
            key: "event_type",
            origin: "provenencia",
            label: "Event type",
            description: "",
            valueType: "term"
        )
    }

    private func makeModel(
        store: FakeStore = FakeStore(),
        properties: [CatalogProperty] = [],
        types: [CatalogSubjectType] = [],
        fieldsByType: [String: [CatalogSubjectTypeField]] = [:]
    ) -> (SubjectFieldsModel, WorkspaceSession, FakeStore) {
        store.propertiesByProject[projectDir] = properties
        store.subjectTypesByProject[projectDir] = types
        store.subjectTypeFieldsByType = fieldsByType
        let session = WorkspaceSession(projectKey: ProjectKey(projectDir: projectDir), store: store)
        let model = SubjectFieldsModel(session: session, userID: userID, store: store)
        return (model, session, store)
    }

    private func waitForQuery<Value>(_ handle: QueryHandle<Value>) async {
        var waited: UInt64 = 0
        let step: UInt64 = 10_000_000
        while waited < 2_000_000_000 {
            if !handle.isFetching, handle.status == .ready || handle.status == .error { return }
            await Task.yield()
            try? await Task.sleep(nanoseconds: step)
            waited += step
        }
    }

    private func warm(_ model: SubjectFieldsModel, session: WorkspaceSession) async {
        model.warmWorkspaceQuery()
        if let handle: QueryHandle<SubjectFieldsSnapshot> = session.queryHandle(
            SubjectFieldsModel.workspaceKey(for: session)
        ) {
            await waitForQuery(handle)
        }
    }

    @Test func loadPopulatesSnapshot() async {
        let person = personType()
        let (model, session, _) = makeModel(
            properties: [nameProperty(), eventTypeProperty()],
            types: [person, eventType()],
            fieldsByType: [
                person.id: [CatalogSubjectTypeField(property: nameProperty(), sortOrder: 0, locked: false)],
            ]
        )
        await warm(model, session: session)
        #expect(model.snapshot.properties.count == 2)
        #expect(model.types.count == 2)
    }

    @Test func typeFilterShowsOnlyBoundProperties() async {
        let person = personType()
        let (model, session, _) = makeModel(
            properties: [nameProperty(), eventTypeProperty()],
            types: [person, eventType()],
            fieldsByType: [
                person.id: [CatalogSubjectTypeField(property: nameProperty(), sortOrder: 0, locked: false)],
            ]
        )
        await warm(model, session: session)
        model.selectType("person")
        #expect(model.visibleProperties.map(\.key) == ["name"])
    }

    @Test func createPropertyUsesResearcherValueTypesOnly() async {
        let (model, session, store) = makeModel(types: [personType()])
        await warm(model, session: session)
        model.openCreate()
        model.draft = SubjectFieldsModel.Draft(label: "Custom Fact", valueType: "text", description: "")
        let ok = await model.submitCreate()
        #expect(ok)
        #expect(store.propertiesByProject[projectDir]?.contains { $0.key == "custom-fact" } == true)
        #expect(!(SubjectPropertyValueType.researcherCreatable.contains("term")))
    }

    @Test func lockedBindingShowsCalloutInsteadOfRemoving() async {
        let person = personType()
        let name = nameProperty()
        let (model, session, _) = makeModel(
            properties: [name],
            types: [person],
            fieldsByType: [
                person.id: [CatalogSubjectTypeField(property: name, sortOrder: 0, locked: true)],
            ]
        )
        await warm(model, session: session)
        model.selectProperty(name.id)
        await model.toggleBinding(to: person)
        #expect(model.lockedCallout != nil)
        #expect(model.isBound(propertyID: name.id, typeID: person.id))
    }

    @Test func assignAndRemoveUnlockedBinding() async {
        let person = personType()
        let custom = CatalogProperty(
            id: "prop-custom",
            key: "custom",
            origin: "user",
            label: "Custom",
            description: "",
            valueType: "text"
        )
        let (model, session, store) = makeModel(
            properties: [custom],
            types: [person]
        )
        await warm(model, session: session)
        model.selectProperty(custom.id)
        await model.toggleBinding(to: person)
        #expect(store.subjectTypeFieldsByType[person.id]?.contains { $0.property.id == custom.id } == true)
        await warm(model, session: session)
        #expect(model.isBound(propertyID: custom.id, typeID: person.id))
        model.selectProperty(custom.id)
        await model.toggleBinding(to: person)
        await warm(model, session: session)
        #expect(store.subjectTypeFieldsByType[person.id]?.contains { $0.property.id == custom.id } != true)
        #expect(!model.isBound(propertyID: custom.id, typeID: person.id))
    }

    @Test func deleteRefusesSeededProperty() async {
        let (model, session, _) = makeModel(properties: [nameProperty()], types: [personType()])
        await warm(model, session: session)
        model.selectProperty(nameProperty().id)
        #expect(!model.canDeleteSelected)
    }

    @Test func deleteRefusesInUseProperty() async {
        let inUse = CatalogProperty(
            id: "prop-in-use",
            key: "custom",
            origin: "user",
            label: "Custom",
            description: "",
            valueType: "text",
            usedBy: 2
        )
        let (model, session, _) = makeModel(properties: [inUse], types: [personType()])
        await warm(model, session: session)
        model.selectProperty(inUse.id)
        #expect(!model.canDeleteSelected)
    }
}
