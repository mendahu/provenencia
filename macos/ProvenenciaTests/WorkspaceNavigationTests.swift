import Foundation
import Testing
@testable import Provenencia

@Suite
@MainActor
struct WorkspaceNavigationTests {
    private let projectUuid = "00000000-0000-7000-8000-0000000000aa"

    private func tempNavigationFile() throws -> URL {
        let dir = FileManager.default.temporaryDirectory
            .appendingPathComponent("nav-history-\(UUID().uuidString)", isDirectory: true)
        try FileManager.default.createDirectory(at: dir, withIntermediateDirectories: true)
        return dir.appendingPathComponent("history.json", isDirectory: false)
    }

    private func attachedNavigation(fileURL: URL? = nil) throws -> (WorkspaceNavigation, URL) {
        let url = try fileURL ?? tempNavigationFile()
        let navigation = WorkspaceNavigation()
        navigation.attachProject(uuid: projectUuid, fileURL: url)
        return (navigation, url)
    }

    @Test func defaultsToSources() {
        let navigation = WorkspaceNavigation()
        #expect(navigation.selectedSection == .sources)
        #expect(navigation.currentLocation == .sectionRoot(.sources))
    }

    @Test func selectingEachSectionUpdatesLocation() {
        let navigation = WorkspaceNavigation()
        for section in WorkspaceSection.allCases {
            navigation.go(to: .sectionRoot(section))
            #expect(navigation.selectedSection == section)
            #expect(navigation.currentLocation == .sectionRoot(section))
        }
    }

    @Test func goPushesAndCoalescesIdentical() throws {
        let (navigation, _) = try attachedNavigation()
        navigation.go(to: .sectionRoot(.sourceFields))
        #expect(navigation.canGoBack)
        navigation.go(to: .sectionRoot(.sourceFields))
        // Identical to current — no second push; still one step back to Sources.
        #expect(navigation.canGoBack)
        #expect(!navigation.canGoForward)

        navigation.go(to: WorkspaceLocation(section: .sources, sourceId: "src-1", title: "A"))
        #expect(navigation.currentLocation.sourceId == "src-1")

        navigation.go(to: WorkspaceLocation(section: .sources, sourceId: "src-1", title: "Renamed"))
        // Title fluff must not push a second identical place.
        navigation.goBack()
        #expect(navigation.currentLocation == .sectionRoot(.sourceFields))
    }

    @Test func goTruncatesForwardFromMiddle() throws {
        let (navigation, _) = try attachedNavigation()
        navigation.go(to: .sectionRoot(.sourceFields))
        navigation.go(to: .sectionRoot(.sourceTypes))
        navigation.go(to: .sectionRoot(.files))
        navigation.goBack()
        navigation.goBack()
        #expect(navigation.currentLocation == .sectionRoot(.sourceFields))
        #expect(navigation.canGoForward)

        navigation.go(to: WorkspaceLocation(section: .sources, sourceId: "src-9"))
        #expect(!navigation.canGoForward)
        #expect(navigation.currentLocation.sourceId == "src-9")
        navigation.goBack()
        #expect(navigation.currentLocation == .sectionRoot(.sourceFields))
    }

    @Test func backForwardAndJumpIndex() throws {
        let (navigation, _) = try attachedNavigation()
        navigation.go(to: .sectionRoot(.sourceFields))
        navigation.go(to: WorkspaceLocation(section: .sourceFields, fieldId: "fld-1", title: "Author"))
        navigation.go(to: .sectionRoot(.sourceTypes))

        #expect(navigation.canGoBack)
        navigation.goBack()
        #expect(navigation.currentLocation.fieldId == "fld-1")
        navigation.goForward()
        #expect(navigation.currentLocation == .sectionRoot(.sourceTypes))

        navigation.go(toIndex: 0)
        #expect(navigation.currentLocation == .sectionRoot(.sources))
        #expect(navigation.canGoForward)
        navigation.go(toIndex: 2)
        #expect(navigation.currentLocation.fieldId == "fld-1")
    }

    @Test func persistRoundTripRestoresIndex() throws {
        let (navigation, url) = try attachedNavigation()
        navigation.go(to: WorkspaceLocation(section: .sources, sourceId: "src-1", ref: "SRC-AAAA", title: "Deed"))
        navigation.go(to: WorkspaceLocation(section: .sourceTypes, typeId: "typ-1", title: "Deed book"))
        navigation.goBack()
        #expect(navigation.currentLocation.sourceId == "src-1")

        let reloaded = WorkspaceNavigation()
        reloaded.attachProject(uuid: projectUuid, fileURL: url)
        #expect(reloaded.currentLocation.sourceId == "src-1")
        #expect(reloaded.canGoBack)
        #expect(reloaded.canGoForward)
        reloaded.goForward()
        #expect(reloaded.currentLocation.typeId == "typ-1")
    }

    @Test func sourcesListAndPageRoundTrip() throws {
        let (navigation, _) = try attachedNavigation()
        navigation.go(to: WorkspaceLocation(section: .sources, sourceId: "src-42", ref: "SRC-4242", title: "Census"))
        #expect(navigation.selectedSection == .sources)
        #expect(navigation.currentLocation.sourceId == "src-42")

        navigation.go(to: .sectionRoot(.sources))
        #expect(navigation.currentLocation.sourceId == nil)
        navigation.goBack()
        #expect(navigation.currentLocation.sourceId == "src-42")
    }

    @Test func vocabularySelectionPushAndRestore() throws {
        let (navigation, _) = try attachedNavigation()
        navigation.go(to: .sectionRoot(.sourceFields))
        navigation.go(to: WorkspaceLocation(section: .sourceFields, fieldId: "fld-9", title: "Place"))
        navigation.go(to: .sectionRoot(.sourceTypes))
        navigation.go(to: WorkspaceLocation(section: .sourceTypes, typeId: "typ-9", title: "Will"))

        navigation.goBack()
        #expect(navigation.currentLocation == .sectionRoot(.sourceTypes))
        navigation.goBack()
        #expect(navigation.currentLocation.fieldId == "fld-9")
        navigation.goForward()
        navigation.goForward()
        #expect(navigation.currentLocation.typeId == "typ-9")
    }

    @Test func missingEntityPruneRewritesCurrentEntry() throws {
        let (navigation, url) = try attachedNavigation()
        navigation.go(to: WorkspaceLocation(section: .sources, sourceId: "gone", title: "Deleted"))
        navigation.fallbackToSectionRoot()
        #expect(navigation.currentLocation == .sectionRoot(.sources))

        let reloaded = WorkspaceNavigation()
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

    @Test func backJumpItemsNearestFirstAndCapped() throws {
        let (navigation, _) = try attachedNavigation()
        // indices: 0 sources, 1 fields, 2 types, 3 files, 4 deep source — current at 4
        navigation.go(to: .sectionRoot(.sourceFields))
        navigation.go(to: .sectionRoot(.sourceTypes))
        navigation.go(to: .sectionRoot(.files))
        navigation.go(to: WorkspaceLocation(section: .sources, sourceId: "src-1", title: "Deep"))

        let back = navigation.backJumpItems(limit: 2)
        #expect(back.map(\.index) == [3, 2])
        #expect(back[0].location == .sectionRoot(.files))
        #expect(back[1].location == .sectionRoot(.sourceTypes))

        let allBack = navigation.backJumpItems(limit: 15)
        #expect(allBack.map(\.index) == [3, 2, 1, 0])
    }

    @Test func forwardJumpItemsStackOrderAndCapped() throws {
        let (navigation, _) = try attachedNavigation()
        navigation.go(to: .sectionRoot(.sourceFields))
        navigation.go(to: .sectionRoot(.sourceTypes))
        navigation.go(to: .sectionRoot(.files))
        navigation.goBack()
        navigation.goBack()
        // current at sourceFields (index 1); forward: types (2), files (3)
        #expect(navigation.currentLocation == .sectionRoot(.sourceFields))

        let forward = navigation.forwardJumpItems(limit: 1)
        #expect(forward.map(\.index) == [2])
        #expect(forward[0].location == .sectionRoot(.sourceTypes))

        let allForward = navigation.forwardJumpItems(limit: 15)
        #expect(allForward.map(\.index) == [2, 3])
        #expect(allForward[1].location == .sectionRoot(.files))
    }

    @Test func jumpItemsEmptyAtStackEnds() throws {
        let (navigation, _) = try attachedNavigation()
        #expect(navigation.backJumpItems().isEmpty)
        #expect(navigation.forwardJumpItems().isEmpty)

        navigation.go(to: .sectionRoot(.sourceFields))
        #expect(navigation.forwardJumpItems().isEmpty)
        #expect(navigation.backJumpItems().map(\.index) == [0])
    }

    @Test func corruptJSONFallsBackAndReportsLoadIssue() throws {
        let url = try tempNavigationFile()
        try Data("{not-json".utf8).write(to: url)

        let (navigation, _) = try attachedNavigation(fileURL: url)
        #expect(navigation.currentLocation == .sectionRoot(.sources))
        #expect(!navigation.canGoBack)
        #expect(navigation.lastHistoryIssue == .loadFailed)

        navigation.acknowledgeHistoryIssue()
        #expect(navigation.lastHistoryIssue == nil)
    }

    @Test func unwritablePathReportsPersistIssue() throws {
        let dir = FileManager.default.temporaryDirectory
            .appendingPathComponent("nav-history-ro-\(UUID().uuidString)", isDirectory: true)
        try FileManager.default.createDirectory(at: dir, withIntermediateDirectories: true)
        try FileManager.default.setAttributes(
            [.posixPermissions: 0o555],
            ofItemAtPath: dir.path
        )
        defer {
            try? FileManager.default.setAttributes(
                [.posixPermissions: 0o755],
                ofItemAtPath: dir.path
            )
            try? FileManager.default.removeItem(at: dir)
        }

        let url = dir.appendingPathComponent("history.json", isDirectory: false)
        let navigation = WorkspaceNavigation()
        navigation.attachProject(uuid: projectUuid, fileURL: url)
        #expect(navigation.currentLocation == .sectionRoot(.sources))
        #expect(navigation.lastHistoryIssue == .persistFailed)

        navigation.acknowledgeHistoryIssue()
        #expect(navigation.lastHistoryIssue == nil)

        // In-memory navigation still works; further writes keep reporting.
        navigation.go(to: .sectionRoot(.sourceFields))
        #expect(navigation.currentLocation == .sectionRoot(.sourceFields))
        #expect(navigation.lastHistoryIssue == .persistFailed)
    }

    @Test func missingHistoryFileDoesNotReportLoadIssue() throws {
        let (navigation, _) = try attachedNavigation()
        #expect(navigation.lastHistoryIssue == nil)
        #expect(navigation.currentLocation == .sectionRoot(.sources))
    }
}
