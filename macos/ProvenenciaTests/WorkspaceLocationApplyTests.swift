import Foundation
import Testing
@testable import Provenencia

@Suite
@MainActor
struct WorkspaceLocationApplyTests {
    private let projectDir = "/tmp/location-apply.provenencia"
    private let userID = "00000000-0000-7000-8000-000000000001"
    private let projectUuid = "00000000-0000-7000-8000-0000000000bb"

    private func tempNavigationFile() throws -> URL {
        let dir = FileManager.default.temporaryDirectory
            .appendingPathComponent("nav-apply-\(UUID().uuidString)", isDirectory: true)
        try FileManager.default.createDirectory(at: dir, withIntermediateDirectories: true)
        return dir.appendingPathComponent("history.json", isDirectory: false)
    }

    private func attachedNavigation(at location: WorkspaceLocation) throws -> WorkspaceNavigation {
        let url = try tempNavigationFile()
        let navigation = WorkspaceNavigation()
        navigation.attachProject(uuid: projectUuid, fileURL: url)
        navigation.go(to: location)
        return navigation
    }

    // MARK: - Omnibar activate

    @Test func activateOmnibarSourceHitCommitsLocationAndClearsResults() throws {
        let hit = CatalogSearchHit(
            kind: "source",
            id: "s1",
            ref: "SRC-AAAAA",
            title: "Deed",
            subtitle: "Photograph",
            matchReason: "title",
            location: WorkspaceLocation(section: .sources, sourceId: "s1", ref: "SRC-AAAAA", title: "Deed")
        )
        let navigation = try attachedNavigation(at: .sectionRoot(.sources))
        let results = OmnibarResultsModel()
        results.query = "Deed"
        results.hits = [hit]
        results.hasSearched = true

        WorkspaceLocationApply.activateOmnibarHit(hit, navigation: navigation, results: results)

        #expect(navigation.currentLocation.sourceId == "s1")
        #expect(navigation.selectedSection == .sources)
        #expect(results.query.isEmpty)
        #expect(results.hits.isEmpty)
        #expect(!results.hasSearched)
    }

    @Test func activateOmnibarFieldHitFromFakeStoreSearch() async throws {
        let store = FakeStore()
        store.fieldsByProject[projectDir] = [
            CatalogMetadataField(
                id: "f9", key: "author", origin: "provenencia",
                label: "Author", dataType: "text", description: ""
            ),
        ]
        let results = OmnibarResultsModel()
        results.query = "Author"
        results.scheduleSearch(
            projectDir: projectDir,
            location: .sectionRoot(.sourceFields),
            store: store
        )
        for _ in 0..<80 {
            if !results.isLoading, results.hasSearched || results.searchError != nil {
                break
            }
            try? await Task.sleep(nanoseconds: 25_000_000)
        }
        let hit = try #require(results.hits.first { $0.id == "f9" })

        let navigation = try attachedNavigation(at: .sectionRoot(.sourceFields))
        WorkspaceLocationApply.activateOmnibarHit(hit, navigation: navigation, results: results)

        #expect(navigation.currentLocation.fieldId == "f9")
        #expect(navigation.selectedSection == .sourceFields)
        #expect(results.query.isEmpty)
    }

    // MARK: - Sources restore

    @Test func sourcesPresentDeepIdOpensAfterLoad() async throws {
        let store = FakeStore()
        store.sourcesByProject[projectDir] = [
            CatalogSource(id: "s1", ref: "SRC-11111", sourceTypeID: "t1", title: "Census", description: ""),
        ]
        store.sourceTypesByProject[projectDir] = [
            CatalogSourceType(id: "t1", key: "census", origin: "provenencia", label: "Census", description: ""),
        ]
        let model = SourcesModel(projectDir: projectDir, userID: userID, store: store)
        await model.load()
        #expect(model.hasCompletedInitialLoad)

        let navigation = try attachedNavigation(
            at: WorkspaceLocation(section: .sources, sourceId: "s1", title: "Census")
        )
        WorkspaceLocationApply.applySources(
            location: navigation.currentLocation,
            model: model,
            navigation: navigation
        )
        #expect(model.openedSourceID == "s1")
        #expect(navigation.currentLocation.sourceId == "s1")
    }

    @Test func sourcesMissingDeepIdFallsBackAfterLoad() async throws {
        let store = FakeStore()
        store.sourcesByProject[projectDir] = [
            CatalogSource(id: "s1", ref: "SRC-11111", sourceTypeID: "", title: "Kept", description: ""),
        ]
        let model = SourcesModel(projectDir: projectDir, userID: userID, store: store)
        await model.load()

        let navigation = try attachedNavigation(
            at: WorkspaceLocation(section: .sources, sourceId: "gone", title: "Deleted")
        )
        WorkspaceLocationApply.applySources(
            location: navigation.currentLocation,
            model: model,
            navigation: navigation
        )
        #expect(model.openedSourceID == nil)
        #expect(navigation.currentLocation == .sectionRoot(.sources))
    }

    @Test func sourcesNilDeepIdClosesOpenedSource() async throws {
        let store = FakeStore()
        store.sourcesByProject[projectDir] = [
            CatalogSource(id: "s1", ref: "SRC-11111", sourceTypeID: "", title: "Census", description: ""),
        ]
        let model = SourcesModel(projectDir: projectDir, userID: userID, store: store)
        await model.load()
        model.openSource(id: "s1")

        let navigation = try attachedNavigation(at: .sectionRoot(.sources))
        WorkspaceLocationApply.applySources(
            location: navigation.currentLocation,
            model: model,
            navigation: navigation
        )
        #expect(model.openedSourceID == nil)
        #expect(navigation.currentLocation == .sectionRoot(.sources))
    }

    // MARK: - Source fields restore

    @Test func fieldsPresentDeepIdSelectsAfterLoad() async throws {
        let store = FakeStore()
        store.fieldsByProject[projectDir] = [
            CatalogMetadataField(
                id: "f1", key: "author", origin: "provenencia",
                label: "Author", dataType: "text", description: ""
            ),
        ]
        let model = SourceFieldsModel(projectDir: projectDir, userID: userID, store: store)
        await model.load()

        let navigation = try attachedNavigation(
            at: WorkspaceLocation(section: .sourceFields, fieldId: "f1", title: "Author")
        )
        WorkspaceLocationApply.applySourceFields(
            location: navigation.currentLocation,
            model: model,
            navigation: navigation
        )
        #expect(model.selectedField?.id == "f1")
    }

    @Test func fieldsMissingDeepIdFallsBackAfterLoad() async throws {
        let store = FakeStore()
        store.fieldsByProject[projectDir] = [
            CatalogMetadataField(
                id: "f1", key: "author", origin: "provenencia",
                label: "Author", dataType: "text", description: ""
            ),
        ]
        let model = SourceFieldsModel(projectDir: projectDir, userID: userID, store: store)
        await model.load()

        let navigation = try attachedNavigation(
            at: WorkspaceLocation(section: .sourceFields, fieldId: "gone", title: "Missing")
        )
        WorkspaceLocationApply.applySourceFields(
            location: navigation.currentLocation,
            model: model,
            navigation: navigation
        )
        #expect(model.selectedField == nil)
        #expect(navigation.currentLocation == .sectionRoot(.sourceFields))
    }

    @Test func fieldsNilDeepIdClearsSelection() async throws {
        let store = FakeStore()
        store.fieldsByProject[projectDir] = [
            CatalogMetadataField(
                id: "f1", key: "author", origin: "provenencia",
                label: "Author", dataType: "text", description: ""
            ),
        ]
        let model = SourceFieldsModel(projectDir: projectDir, userID: userID, store: store)
        await model.load()
        model.select("f1")

        let navigation = try attachedNavigation(at: .sectionRoot(.sourceFields))
        WorkspaceLocationApply.applySourceFields(
            location: navigation.currentLocation,
            model: model,
            navigation: navigation
        )
        #expect(model.selectedField == nil)
    }

    @Test func fieldsSkipApplyWhileAdding() async throws {
        let store = FakeStore()
        store.fieldsByProject[projectDir] = [
            CatalogMetadataField(
                id: "f1", key: "author", origin: "provenencia",
                label: "Author", dataType: "text", description: ""
            ),
        ]
        let model = SourceFieldsModel(projectDir: projectDir, userID: userID, store: store)
        await model.load()
        model.openAdd()
        #expect(model.isAdding)

        let navigation = try attachedNavigation(
            at: WorkspaceLocation(section: .sourceFields, fieldId: "f1", title: "Author")
        )
        WorkspaceLocationApply.applySourceFields(
            location: navigation.currentLocation,
            model: model,
            navigation: navigation
        )
        #expect(model.isAdding)
        #expect(model.selectedField == nil)
        #expect(navigation.currentLocation.fieldId == "f1")
    }

    // MARK: - Source types restore

    @Test func typesPresentDeepIdSelectsAfterLoad() async throws {
        let store = FakeStore()
        store.sourceTypesByProject[projectDir] = [
            CatalogSourceType(
                id: "t1", key: "photograph", origin: "provenencia",
                label: "Photograph", description: ""
            ),
        ]
        let model = SourceTypesModel(projectDir: projectDir, userID: userID, store: store)
        await model.load()

        let navigation = try attachedNavigation(
            at: WorkspaceLocation(section: .sourceTypes, typeId: "t1", title: "Photograph")
        )
        WorkspaceLocationApply.applySourceTypes(
            location: navigation.currentLocation,
            model: model,
            navigation: navigation
        )
        #expect(model.selectedType?.id == "t1")
    }

    @Test func typesMissingDeepIdFallsBackAfterLoad() async throws {
        let store = FakeStore()
        store.sourceTypesByProject[projectDir] = [
            CatalogSourceType(
                id: "t1", key: "photograph", origin: "provenencia",
                label: "Photograph", description: ""
            ),
        ]
        let model = SourceTypesModel(projectDir: projectDir, userID: userID, store: store)
        await model.load()

        let navigation = try attachedNavigation(
            at: WorkspaceLocation(section: .sourceTypes, typeId: "gone", title: "Missing")
        )
        WorkspaceLocationApply.applySourceTypes(
            location: navigation.currentLocation,
            model: model,
            navigation: navigation
        )
        #expect(model.selectedType == nil)
        #expect(navigation.currentLocation == .sectionRoot(.sourceTypes))
    }

    @Test func typesNilDeepIdClearsSelection() async throws {
        let store = FakeStore()
        store.sourceTypesByProject[projectDir] = [
            CatalogSourceType(
                id: "t1", key: "photograph", origin: "provenencia",
                label: "Photograph", description: ""
            ),
        ]
        let model = SourceTypesModel(projectDir: projectDir, userID: userID, store: store)
        await model.load()
        model.select("t1")

        let navigation = try attachedNavigation(at: .sectionRoot(.sourceTypes))
        WorkspaceLocationApply.applySourceTypes(
            location: navigation.currentLocation,
            model: model,
            navigation: navigation
        )
        #expect(model.selectedType == nil)
    }

    @Test func typesSkipApplyWhileAdding() async throws {
        let store = FakeStore()
        store.sourceTypesByProject[projectDir] = [
            CatalogSourceType(
                id: "t1", key: "photograph", origin: "provenencia",
                label: "Photograph", description: ""
            ),
        ]
        let model = SourceTypesModel(projectDir: projectDir, userID: userID, store: store)
        await model.load()
        model.openAdd()
        #expect(model.isAdding)

        let navigation = try attachedNavigation(
            at: WorkspaceLocation(section: .sourceTypes, typeId: "t1", title: "Photograph")
        )
        WorkspaceLocationApply.applySourceTypes(
            location: navigation.currentLocation,
            model: model,
            navigation: navigation
        )
        #expect(model.isAdding)
        #expect(model.selectedType == nil)
        #expect(navigation.currentLocation.typeId == "t1")
    }
}
