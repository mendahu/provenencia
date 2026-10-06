import Foundation
import Testing
@testable import Provenencia

@Suite
@MainActor
struct PropertiesModelTests {
    private let projectDir = "/tmp/properties.provenencia"
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
        fieldsByType: [String: [CatalogSubjectTypeProperty]] = [:]
    ) -> (PropertiesModel, WorkspaceSession, FakeStore) {
        store.propertiesByProject[projectDir] = properties
        store.subjectTypesByProject[projectDir] = types
        store.subjectTypePropertiesByType = fieldsByType
        let session = WorkspaceSession(projectKey: ProjectKey(projectDir: projectDir), store: store)
        let model = PropertiesModel(session: session, userID: userID, store: store)
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

    private func warm(_ model: PropertiesModel, session: WorkspaceSession) async {
        model.warmWorkspaceQuery()
        if let handle: QueryHandle<PropertiesSnapshot> = session.queryHandle(
            PropertiesModel.workspaceKey(for: session)
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
                person.id: [CatalogSubjectTypeProperty(property: nameProperty(), sortOrder: 0, locked: false)],
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
                person.id: [CatalogSubjectTypeProperty(property: nameProperty(), sortOrder: 0, locked: false)],
            ]
        )
        await warm(model, session: session)
        model.selectType("person")
        #expect(model.visibleProperties.map(\.key) == ["name"])
    }

    @Test func createPropertyUsesResearcherValueTypesOnly() async {
        let person = personType()
        let (model, session, store) = makeModel(types: [person])
        await warm(model, session: session)
        model.selectType("person")
        model.openCreate()
        model.draft = PropertiesModel.Draft(
            label: "Custom Fact",
            valueType: "text",
            description: "",
            bindTypeIDs: [person.id]
        )
        let location = await model.submitCreate()
        #expect(location?.section == .properties)
        #expect(location?.subjectTypeKey == "person")
        #expect(location?.propertyId == model.selectedPropertyID)
        #expect(location?.title == "Custom Fact")
        #expect(store.propertiesByProject[projectDir]?.contains { $0.key == "custom-fact" && $0.cardinality == "single" } == true)
        #expect(store.subjectTypePropertiesByType[person.id]?.contains { $0.property.key == "custom-fact" } == true)
        #expect(!(SubjectPropertyValueType.researcherCreatable.contains("term")))
    }

    @Test func createDefaultsToOneValueAndCanSubmitSeveral() async {
        let (model, session, store) = makeModel(types: [personType()])
        await warm(model, session: session)
        model.openCreate()
        #expect(model.draft?.cardinality == "single")
        model.draft?.cardinality = "multiple"
        model.draft?.label = "Languages spoken"
        let location = await model.submitCreate()
        #expect(location?.title == "Languages spoken")
        #expect(store.propertiesByProject[projectDir]?.first?.cardinality == "multiple")
    }

    @Test func userPropertyCardinalityChangeKeepsTheNoticeUntilSelectionMoves() async {
        let store = FakeStore()
        let property = userProperty()
        let other = nameProperty()
        let (model, session, _) = makeModel(store: store, properties: [property, other], types: [personType()])
        await warm(model, session: session)
        model.selectProperty(property.id)
        #expect(model.canEditCardinality)
        #expect(await model.setCardinality("single") == false)
        #expect(await model.setCardinality("multiple"))
        #expect(model.cardinalityNotice)
        #expect(store.propertiesByProject[projectDir]?.first { $0.id == property.id }?.cardinality == "multiple")
        model.selectProperty(other.id)
        #expect(!model.cardinalityNotice)
        #expect(!model.canEditCardinality)
        #expect(await model.setCardinality("single") == false)
        #expect(store.propertiesByProject[projectDir]?.first { $0.id == other.id }?.cardinality == "single")
    }

    @Test func typesFollowPaletteOrderNotAlphabetical() async {
        let (model, session, _) = makeModel(
            types: [eventType(), personType()]
        )
        await warm(model, session: session)
        #expect(model.types.map(\.key) == ["person", "event"])
    }

    @Test func lockedBindingShowsCalloutInsteadOfRemoving() async {
        let person = personType()
        let name = nameProperty()
        let (model, session, _) = makeModel(
            properties: [name],
            types: [person],
            fieldsByType: [
                person.id: [CatalogSubjectTypeProperty(property: name, sortOrder: 0, locked: true)],
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
        #expect(store.subjectTypePropertiesByType[person.id]?.contains { $0.property.id == custom.id } == true)
        await warm(model, session: session)
        #expect(model.isBound(propertyID: custom.id, typeID: person.id))
        model.selectProperty(custom.id)
        await model.toggleBinding(to: person)
        await warm(model, session: session)
        #expect(store.subjectTypePropertiesByType[person.id]?.contains { $0.property.id == custom.id } != true)
        #expect(!model.isBound(propertyID: custom.id, typeID: person.id))
    }

    private func userProperty(
        id: String = "prop-user",
        usedBy: Int = 0
    ) -> CatalogProperty {
        CatalogProperty(
            id: id,
            key: "burial_ground",
            origin: "user",
            label: "Burial ground",
            description: "Cemetery name",
            valueType: "text",
            usedBy: usedBy
        )
    }

    private func pluginProperty() -> CatalogProperty {
        CatalogProperty(
            id: "prop-plugin",
            key: "plugin_fact",
            origin: "plugin:acme",
            label: "Plugin fact",
            description: "",
            valueType: "text"
        )
    }

    private func observation(
        id: String,
        ref: String,
        propertyID: String,
        title: String
    ) -> CatalogObservation {
        CatalogObservation(
            id: id,
            ref: ref,
            citationID: "cit-1",
            subjectID: "sub-1",
            propertyID: propertyID,
            polarity: "positive",
            valueText: title,
            valueInteger: nil,
            valueDateID: "",
            valueNameID: "",
            valueSubjectID: "",
            valueTermID: "",
            propertyKey: "outcome",
            propertyLabel: "Outcome",
            propertyValueType: "text"
        )
    }

    @Test func trashIsOfferedWhenUsedBySeededOrPlugin() async {
        let inUse = userProperty(id: "prop-used", usedBy: 2)
        let (model, session, _) = makeModel(
            properties: [userProperty(), nameProperty(), pluginProperty(), inUse],
            types: [personType()]
        )
        await warm(model, session: session)

        model.selectProperty(userProperty().id)
        #expect(model.canDeleteSelected)
        #expect(model.showsDelete)
        #expect(model.canEditSelected)
        #expect(model.deleteTooltip == L10n.Properties.deleteProperty)
        #expect(
            model.deleteAccessibilityLabel
                == L10n.Properties.deletePropertyAccessibility(label: "Burial ground")
        )

        model.selectProperty(nameProperty().id)
        #expect(model.canDeleteSelected)
        #expect(model.canEditSelected)

        model.selectProperty(pluginProperty().id)
        #expect(model.canDeleteSelected)
        #expect(!model.canEditSelected)

        model.selectProperty(inUse.id)
        #expect(model.canDeleteSelected)
    }

    @Test func unusedUserPropertyConfirmDeletes() async {
        let store = FakeStore()
        let property = userProperty()
        let (model, session, _) = makeModel(store: store, properties: [property], types: [personType()])
        await warm(model, session: session)
        model.selectProperty(property.id)
        await model.askDelete()

        #expect(model.pendingImpact?.report.allowed == true)
        #expect(model.pendingImpact?.target.kind == "property")
        #expect(await model.confirmPendingImpact())
        #expect(store.propertiesByProject[projectDir]?.isEmpty == true)
        #expect(model.selectedPropertyID == nil)
    }

    @Test func boundOnlyPropertyConfirmDeletes() async {
        let store = FakeStore()
        let property = userProperty()
        let person = personType()
        let (model, session, _) = makeModel(
            store: store,
            properties: [property],
            types: [person],
            fieldsByType: [
                person.id: [CatalogSubjectTypeProperty(property: property, sortOrder: 0, locked: false)],
            ]
        )
        await warm(model, session: session)
        model.selectProperty(property.id)
        await model.askDelete()

        #expect(model.pendingImpact?.report.allowed == true)
        #expect(await model.confirmPendingImpact())
        #expect(store.propertiesByProject[projectDir]?.isEmpty == true)
        #expect(store.subjectTypePropertiesByType[person.id]?.isEmpty == true)
        #expect(model.selectedPropertyID == nil)
    }

    @Test func propertyUsedOnTwoObservationsShowsInboundNotice() async {
        let store = FakeStore()
        let property = userProperty(usedBy: 2)
        store.observationsBySource["src-1"] = [
            observation(id: "obs-1", ref: "OBS-AAAAA", propertyID: property.id, title: "accidental"),
            observation(id: "obs-2", ref: "OBS-BBBBB", propertyID: property.id, title: "natural"),
        ]
        let (model, session, _) = makeModel(store: store, properties: [property], types: [personType()])
        await warm(model, session: session)
        model.selectProperty(property.id)
        #expect(model.canDeleteSelected)
        await model.askDelete()

        let report = model.pendingImpact?.report
        #expect(report?.allowed == false)
        #expect(report?.gate == .inbound)
        #expect(report?.groups.first?.via == "observations.property_id")
        #expect(report?.groups.first?.total == 2)
        #expect(report?.groups.first?.total == model.selectedProperty?.usedBy)
        #expect(report?.groups.first?.listed.map(\.ref) == ["OBS-AAAAA", "OBS-BBBBB"])
        #expect(await model.confirmPendingImpact() == false)
        #expect(store.propertiesByProject[projectDir]?.map(\.id) == [property.id])
        #expect(model.selectedPropertyID == property.id)
    }

    @Test func unusedTermsShowInboundNotice() async {
        let store = FakeStore()
        let property = userProperty()
        store.propertyTermsByProperty[property.id] = [
            CatalogPropertyTerm(
                id: "term-1",
                propertyID: property.id,
                key: "lodger",
                origin: "user",
                label: "Lodger",
                description: ""
            ),
        ]
        let (model, session, _) = makeModel(store: store, properties: [property], types: [personType()])
        await warm(model, session: session)
        model.selectProperty(property.id)
        await model.askDelete()

        let report = model.pendingImpact?.report
        #expect(report?.allowed == false)
        #expect(report?.gate == .inbound)
        #expect(report?.groups.first?.via == "property_terms.property_id")
        #expect(report?.groups.first?.listed.map(\.ref) == ["lodger"])
        // Term rows deep-link to the parent property's inspector row, as the Go projector does.
        #expect(
            report?.groups.first?.listed.first?.location
                == WorkspaceLocation(section: .properties, propertyId: property.id)
        )
        #expect(await model.confirmPendingImpact() == false)
        #expect(store.propertiesByProject[projectDir]?.map(\.id) == [property.id])
    }

    @Test func seededPropertyOriginLockedTrashStillOffered() async {
        let store = FakeStore()
        let (model, session, _) = makeModel(store: store, properties: [nameProperty()], types: [personType()])
        await warm(model, session: session)
        model.selectProperty(nameProperty().id)
        #expect(model.canDeleteSelected)
        await model.askDelete()

        #expect(model.pendingImpact?.report.allowed == false)
        #expect(model.pendingImpact?.report.gate == .originLocked)
        #expect(await model.confirmPendingImpact() == false)
        #expect(store.propertiesByProject[projectDir]?.map(\.id) == [nameProperty().id])
    }

    @Test func pluginPropertyOriginLockedTrashStillOffered() async {
        let store = FakeStore()
        let (model, session, _) = makeModel(store: store, properties: [pluginProperty()], types: [personType()])
        await warm(model, session: session)
        model.selectProperty(pluginProperty().id)
        #expect(model.canDeleteSelected)
        await model.askDelete()

        #expect(model.pendingImpact?.report.allowed == false)
        #expect(model.pendingImpact?.report.gate == .originLocked)
        #expect(await model.confirmPendingImpact() == false)
        #expect(store.propertiesByProject[projectDir]?.map(\.id) == [pluginProperty().id])
    }

    @Test func isDirtyTracksUnsavedEditsAndMatchingValuesClearThem() async {
        let (model, session, _) = makeModel(properties: [userProperty()], types: [personType()])
        await warm(model, session: session)
        model.selectProperty(userProperty().id)
        model.beginEdit()
        #expect(!model.isEditDirty(label: "Burial ground", description: "Cemetery name"))
        #expect(model.isEditDirty(label: "Changed", description: "Cemetery name"))
        #expect(!model.isEditDirty(label: "Burial ground", description: "Cemetery name"))
        #expect(!model.canSubmitEdit(label: "Burial ground", description: "Cemetery name"))
    }

    @Test func saveEditPersistsLabelAndDescriptionAndKeepsKey() async {
        let store = FakeStore()
        let property = userProperty()
        let (model, session, _) = makeModel(store: store, properties: [property], types: [personType()])
        await warm(model, session: session)
        model.selectProperty(property.id)
        model.beginEdit()
        #expect(model.canSubmitEdit(label: "Cemetery", description: "Where they were buried"))
        #expect(await model.submitEdit(label: "Cemetery", description: "Where they were buried"))
        #expect(model.isEditingIdentity == false)
        await warm(model, session: session)
        #expect(store.propertiesByProject[projectDir]?.first?.label == "Cemetery")
        #expect(store.propertiesByProject[projectDir]?.first?.description == "Where they were buried")
        #expect(store.propertiesByProject[projectDir]?.first?.key == "burial_ground")
        #expect(store.propertiesByProject[projectDir]?.first?.valueType == "text")
        #expect(model.selectedPropertyID == property.id)
    }

    // MARK: History selection

    private func makeBoundModel() -> (PropertiesModel, WorkspaceSession) {
        let person = personType()
        let (model, session, _) = makeModel(
            properties: [nameProperty(), eventTypeProperty()],
            types: [person, eventType()],
            fieldsByType: [
                person.id: [CatalogSubjectTypeProperty(property: nameProperty(), sortOrder: 0, locked: true)],
            ]
        )
        return (model, session)
    }

    @Test func syncSelectionRestoresCategoryAndProperty() async {
        let (model, session) = makeBoundModel()
        await warm(model, session: session)

        let outcome = model.syncSelection(
            from: WorkspaceLocation(section: .properties, subjectTypeKey: "person", propertyId: "prop-name")
        )
        #expect(outcome == .applied)
        #expect(model.selectedTypeKey == "person")
        #expect(model.selectedPropertyID == "prop-name")
        #expect(model.selectedProperty?.key == "name")

        #expect(model.syncSelection(from: WorkspaceLocation(section: .properties, subjectTypeKey: "event")) == .applied)
        #expect(model.selectedTypeKey == "event")
        #expect(model.selectedPropertyID == nil)

        #expect(model.syncSelection(from: .sectionRoot(.properties)) == .applied)
        #expect(model.selectedTypeKey == nil)
        #expect(model.selectedPropertyID == nil)
    }

    @Test func syncSelectionIgnoresOtherSectionsAndOpenCreate() async {
        let (model, session) = makeBoundModel()
        await warm(model, session: session)
        model.selectType("person")
        model.selectProperty("prop-name")

        #expect(model.syncSelection(from: WorkspaceLocation(section: .metadata, fieldId: "fld-1")) == .ignored)
        #expect(model.selectedTypeKey == "person")
        #expect(model.selectedPropertyID == "prop-name")

        model.openCreate()
        #expect(model.syncSelection(from: .sectionRoot(.properties)) == .ignored)
        #expect(model.selectedPropertyID == "prop-name")
    }

    @Test func syncSelectionReportsMissingCategoryOrProperty() async {
        let (model, session) = makeBoundModel()
        await warm(model, session: session)

        #expect(
            model.syncSelection(
                from: WorkspaceLocation(section: .properties, subjectTypeKey: "person", propertyId: "gone")
            ) == .missingDeepId
        )
        #expect(model.selectedTypeKey == "person")
        #expect(model.selectedPropertyID == nil)

        #expect(
            model.syncSelection(
                from: WorkspaceLocation(section: .properties, subjectTypeKey: "ghost", propertyId: "prop-name")
            ) == .missingDeepId
        )
        #expect(model.selectedTypeKey == nil)
        #expect(model.selectedPropertyID == nil)
    }

    @Test func syncSelectionBeforeLoadIsIgnored() {
        let (model, _) = makeBoundModel()
        #expect(
            model.syncSelection(from: WorkspaceLocation(section: .properties, propertyId: "prop-name")) == .ignored
        )
        #expect(model.selectedPropertyID == nil)
    }

    @Test func reapplyingCurrentPlaceKeepsIdentityEdit() async {
        let (model, session, _) = makeModel(properties: [userProperty()], types: [personType()])
        await warm(model, session: session)
        model.selectProperty(userProperty().id)
        model.beginEdit()

        let place = WorkspaceLocation(section: .properties, propertyId: userProperty().id)
        #expect(model.syncSelection(from: place) == .applied)
        #expect(model.isEditingIdentity)

        #expect(model.syncSelection(from: .sectionRoot(.properties)) == .applied)
        #expect(!model.isEditingIdentity)
    }

    @Test func locationBuildersCarryBothLevelsAndTitle() async {
        let (model, session) = makeBoundModel()
        await warm(model, session: session)

        let category = model.location(afterPressingType: "person")
        #expect(category == WorkspaceLocation(section: .properties, subjectTypeKey: "person"))
        #expect(category.title == "Person")
        model.selectType("person")

        let row = model.location(selectingProperty: "prop-name")
        #expect(row == WorkspaceLocation(section: .properties, subjectTypeKey: "person", propertyId: "prop-name"))
        #expect(row.title == "Name")
        model.selectProperty("prop-name")

        // Pressing the focused type again clears the filter but keeps the inspector row.
        let cleared = model.location(afterPressingType: "person")
        #expect(cleared == WorkspaceLocation(section: .properties, propertyId: "prop-name"))
        #expect(cleared.title == "Name")
    }

    @Test func addPropertyFromComboReturnsRowPlace() async {
        let person = personType()
        let custom = CatalogProperty(
            id: "prop-custom",
            key: "custom",
            origin: "user",
            label: "Custom",
            description: "",
            valueType: "text"
        )
        let (model, session, store) = makeModel(properties: [custom], types: [person])
        await warm(model, session: session)
        model.selectType("person")
        model.addPropertySelection = custom.id

        let location = await model.addPropertyFromCombo()
        #expect(location == WorkspaceLocation(section: .properties, subjectTypeKey: "person", propertyId: custom.id))
        #expect(model.selectedPropertyID == custom.id)
        #expect(model.addPropertySelection.isEmpty)
        #expect(store.subjectTypePropertiesByType[person.id]?.contains { $0.property.id == custom.id } == true)
    }

    @Test func pluginHasNoEditControl() async {
        let (model, session, _) = makeModel(properties: [pluginProperty()], types: [personType()])
        await warm(model, session: session)
        model.selectProperty(pluginProperty().id)
        #expect(!model.canEditSelected)
        model.beginEdit()
        #expect(model.isEditingIdentity == false)
        #expect(!model.canSubmitEdit(label: "Plugin fact", description: ""))
    }
}
