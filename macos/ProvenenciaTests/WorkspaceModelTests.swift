import Foundation
import Testing
@testable import Provenencia

@Suite
@MainActor
struct WorkspaceModelTests {
    private let projectUuid = "00000000-0000-7000-8000-0000000000aa"

    private func makeDefaults() -> UserDefaults {
        let suiteName = "WorkspaceModelTests.\(UUID().uuidString)"
        let defaults = UserDefaults(suiteName: suiteName)!
        defaults.removePersistentDomain(forName: suiteName)
        return defaults
    }

    private func makeModel(defaults: UserDefaults? = nil) -> WorkspaceModel {
        WorkspaceModel(defaults: defaults ?? makeDefaults())
    }

    private func tempNavigationFile() throws -> URL {
        let dir = FileManager.default.temporaryDirectory
            .appendingPathComponent("nav-history-\(UUID().uuidString)", isDirectory: true)
        try FileManager.default.createDirectory(at: dir, withIntermediateDirectories: true)
        return dir.appendingPathComponent("history.json", isDirectory: false)
    }

    private func attachedModel(fileURL: URL? = nil) throws -> (WorkspaceModel, URL) {
        let url = try fileURL ?? tempNavigationFile()
        let model = makeModel()
        model.attachProject(uuid: projectUuid, fileURL: url)
        return (model, url)
    }

    @Test func defaultsToSourcesExpanded() {
        let model = makeModel()
        #expect(model.selectedSection == .sources)
        #expect(model.currentLocation == .sectionRoot(.sources))
        #expect(!model.isSidebarCollapsed)
    }

    @Test func toggleCollapsesAndPersists() {
        let defaults = makeDefaults()
        let model = makeModel(defaults: defaults)
        model.toggleSidebarCollapsed()
        #expect(model.isSidebarCollapsed)

        let reloaded = makeModel(defaults: defaults)
        #expect(reloaded.isSidebarCollapsed)
    }

    @Test func toggleTwiceReturnsToExpanded() {
        let model = makeModel()
        model.toggleSidebarCollapsed()
        model.toggleSidebarCollapsed()
        #expect(!model.isSidebarCollapsed)
    }

    @Test func selectingEachSectionUpdatesLocation() {
        let model = makeModel()
        for section in WorkspaceModel.Section.allCases {
            model.go(to: .sectionRoot(section))
            #expect(model.selectedSection == section)
            #expect(model.currentLocation == .sectionRoot(section))
        }
    }

    @Test func goPushesAndCoalescesIdentical() throws {
        let (model, _) = try attachedModel()
        model.go(to: .sectionRoot(.sourceFields))
        #expect(model.canGoBack)
        model.go(to: .sectionRoot(.sourceFields))
        // Identical to current — no second push; still one step back to Sources.
        #expect(model.canGoBack)
        #expect(!model.canGoForward)

        model.go(to: WorkspaceLocation(section: .sources, sourceId: "src-1", title: "A"))
        #expect(model.currentLocation.sourceId == "src-1")

        model.go(to: WorkspaceLocation(section: .sources, sourceId: "src-1", title: "Renamed"))
        // Title fluff must not push a second identical place.
        model.goBack()
        #expect(model.currentLocation == .sectionRoot(.sourceFields))
    }

    @Test func goTruncatesForwardFromMiddle() throws {
        let (model, _) = try attachedModel()
        model.go(to: .sectionRoot(.sourceFields))
        model.go(to: .sectionRoot(.sourceTypes))
        model.go(to: .sectionRoot(.files))
        model.goBack()
        model.goBack()
        #expect(model.currentLocation == .sectionRoot(.sourceFields))
        #expect(model.canGoForward)

        model.go(to: WorkspaceLocation(section: .sources, sourceId: "src-9"))
        #expect(!model.canGoForward)
        #expect(model.currentLocation.sourceId == "src-9")
        model.goBack()
        #expect(model.currentLocation == .sectionRoot(.sourceFields))
    }

    @Test func backForwardAndJumpIndex() throws {
        let (model, _) = try attachedModel()
        model.go(to: .sectionRoot(.sourceFields))
        model.go(to: WorkspaceLocation(section: .sourceFields, fieldId: "fld-1", title: "Author"))
        model.go(to: .sectionRoot(.sourceTypes))

        #expect(model.canGoBack)
        model.goBack()
        #expect(model.currentLocation.fieldId == "fld-1")
        model.goForward()
        #expect(model.currentLocation == .sectionRoot(.sourceTypes))

        model.go(toIndex: 0)
        #expect(model.currentLocation == .sectionRoot(.sources))
        #expect(model.canGoForward)
        model.go(toIndex: 2)
        #expect(model.currentLocation.fieldId == "fld-1")
    }

    @Test func persistRoundTripRestoresIndex() throws {
        let (model, url) = try attachedModel()
        model.go(to: WorkspaceLocation(section: .sources, sourceId: "src-1", ref: "SRC-AAAA", title: "Deed"))
        model.go(to: WorkspaceLocation(section: .sourceTypes, typeId: "typ-1", title: "Deed book"))
        model.goBack()
        #expect(model.currentLocation.sourceId == "src-1")

        let reloaded = makeModel()
        reloaded.attachProject(uuid: projectUuid, fileURL: url)
        #expect(reloaded.currentLocation.sourceId == "src-1")
        #expect(reloaded.canGoBack)
        #expect(reloaded.canGoForward)
        reloaded.goForward()
        #expect(reloaded.currentLocation.typeId == "typ-1")
    }

    @Test func sourcesListAndPageRoundTrip() throws {
        let (model, _) = try attachedModel()
        model.go(to: WorkspaceLocation(section: .sources, sourceId: "src-42", ref: "SRC-4242", title: "Census"))
        #expect(model.selectedSection == .sources)
        #expect(model.currentLocation.sourceId == "src-42")

        model.go(to: .sectionRoot(.sources))
        #expect(model.currentLocation.sourceId == nil)
        model.goBack()
        #expect(model.currentLocation.sourceId == "src-42")
    }

    @Test func vocabularySelectionPushAndRestore() throws {
        let (model, _) = try attachedModel()
        model.go(to: .sectionRoot(.sourceFields))
        model.go(to: WorkspaceLocation(section: .sourceFields, fieldId: "fld-9", title: "Place"))
        model.go(to: .sectionRoot(.sourceTypes))
        model.go(to: WorkspaceLocation(section: .sourceTypes, typeId: "typ-9", title: "Will"))

        model.goBack()
        #expect(model.currentLocation == .sectionRoot(.sourceTypes))
        model.goBack()
        #expect(model.currentLocation.fieldId == "fld-9")
        model.goForward()
        model.goForward()
        #expect(model.currentLocation.typeId == "typ-9")
    }

    @Test func missingEntityPruneRewritesCurrentEntry() throws {
        let (model, url) = try attachedModel()
        model.go(to: WorkspaceLocation(section: .sources, sourceId: "gone", title: "Deleted"))
        model.fallbackToSectionRoot()
        #expect(model.currentLocation == .sectionRoot(.sources))

        let reloaded = makeModel()
        reloaded.attachProject(uuid: projectUuid, fileURL: url)
        #expect(reloaded.currentLocation == .sectionRoot(.sources))
        #expect(reloaded.currentLocation.sourceId == nil)
    }

    @Test func locationEqualityIgnoresTitleAndRef() {
        let a = WorkspaceLocation(section: .sources, sourceId: "1", ref: "SRC-1", title: "Old")
        let b = WorkspaceLocation(section: .sources, sourceId: "1", ref: "SRC-2", title: "New")
        #expect(a == b)
        #expect(a != WorkspaceLocation(section: .sources, sourceId: "2"))
    }

    @Test func navigationFileNameStripsDashesAndBraces() {
        #expect(
            InstallPaths.navigationFileName(projectUuid: "{00000000-0000-7000-8000-0000000000AA}")
                == "000000000000700080000000000000aa.json"
        )
    }
}
