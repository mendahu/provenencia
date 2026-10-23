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

    private func entryCount(_ navigation: WorkspaceNavigation) -> Int {
        navigation.backJumpItems(limit: 100).count + navigation.forwardJumpItems(limit: 100).count + 1
    }

    private func attachedNavigation(fileURL: URL? = nil) throws -> (WorkspaceNavigation, URL) {
        let url = try fileURL ?? tempNavigationFile()
        let navigation = WorkspaceNavigation()
        navigation.attachProject(uuid: projectUuid, fileURL: url)
        return (navigation, url)
    }

    @Test func eventsListAndDetailStayOffThePersonsPlace() throws {
        let (navigation, _) = try attachedNavigation()
        navigation.go(to: .sectionRoot(.events))
        navigation.go(to: .eventDetail(entityId: "evt-1", ref: "EVT-1", title: "Fire at York"))
        #expect(navigation.currentLocation.section == .events)
        #expect(navigation.currentLocation.entityId == "evt-1")
        navigation.goBack()
        #expect(navigation.currentLocation == .sectionRoot(.events))
        #expect(navigation.currentLocation.section != .persons)
        navigation.goForward()
        #expect(navigation.currentLocation.entityId == "evt-1")
        #expect(navigation.currentLocation.section == .events)
        navigation.goBack()
        navigation.goBack()
        #expect(navigation.currentLocation == .sectionRoot(.sources))
    }

    @Test func attachProjectSetsUuidAndCommitsSynchronously() throws {
        let (seed, url) = try attachedNavigation()
        seed.go(to: WorkspaceLocation.sectionRoot(.sources).withListScrollAnchor("s3"))

        let navigation = WorkspaceNavigation()
        var recorded: [WorkspaceLocation] = []
        navigation.onLocationCommit = { recorded.append($0) }
        navigation.attachProject(uuid: projectUuid, fileURL: url)

        #expect(navigation.projectUuid == projectUuid)
        #expect(navigation.currentLocation == .sectionRoot(.sources))
        #expect(navigation.currentLocation.listScrollAnchor == "s3")
        #expect(recorded.count == 1)
        #expect(recorded[0].listScrollAnchor == "s3")

        navigation.attachProject(uuid: projectUuid, fileURL: url)
        #expect(recorded.count == 1)
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
        navigation.go(to: .sectionRoot(.metadata))
        #expect(navigation.canGoBack)
        navigation.go(to: .sectionRoot(.metadata))
        // Identical to current — no second push; still one step back to Sources.
        #expect(navigation.canGoBack)
        #expect(!navigation.canGoForward)

        navigation.go(to: WorkspaceLocation(section: .sources, sourceId: "src-1", title: "A"))
        #expect(navigation.currentLocation.sourceId == "src-1")

        navigation.go(to: WorkspaceLocation(section: .sources, sourceId: "src-1", title: "Renamed"))
        // Title fluff must not push a second identical place.
        navigation.goBack()
        #expect(navigation.currentLocation == .sectionRoot(.metadata))
    }

    @Test func goTruncatesForwardFromMiddle() throws {
        let (navigation, _) = try attachedNavigation()
        navigation.go(to: .sectionRoot(.metadata))
        navigation.go(to: .sectionRoot(.sourceTypes))
        navigation.go(to: .sectionRoot(.sources))
        navigation.goBack()
        navigation.goBack()
        #expect(navigation.currentLocation == .sectionRoot(.metadata))
        #expect(navigation.canGoForward)

        navigation.go(to: WorkspaceLocation(section: .sources, sourceId: "src-9"))
        #expect(!navigation.canGoForward)
        #expect(navigation.currentLocation.sourceId == "src-9")
        navigation.goBack()
        #expect(navigation.currentLocation == .sectionRoot(.metadata))
    }

    @Test func backForwardAndJumpIndex() throws {
        let (navigation, _) = try attachedNavigation()
        navigation.go(to: .sectionRoot(.metadata))
        navigation.go(to: WorkspaceLocation(section: .metadata, fieldId: "fld-1", title: "Author"))
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
        navigation.go(to: .sectionRoot(.metadata))
        navigation.go(to: WorkspaceLocation(section: .metadata, fieldId: "fld-9", title: "Place"))
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

    @Test func propertiesTwoLevelSelectionPushAndRestore() throws {
        let (navigation, url) = try attachedNavigation()
        navigation.go(to: .sectionRoot(.properties))
        navigation.go(to: WorkspaceLocation(section: .properties, subjectTypeKey: "person", title: "Person"))
        navigation.go(
            to: WorkspaceLocation(
                section: .properties,
                subjectTypeKey: "person",
                propertyId: "prop-name",
                title: "Name"
            )
        )
        navigation.go(to: .sectionRoot(.sources))

        navigation.goBack()
        #expect(navigation.selectedSection == .properties)
        #expect(navigation.currentLocation.subjectTypeKey == "person")
        #expect(navigation.currentLocation.propertyId == "prop-name")
        navigation.goBack()
        #expect(navigation.currentLocation.subjectTypeKey == "person")
        #expect(navigation.currentLocation.propertyId == nil)
        navigation.goBack()
        #expect(navigation.currentLocation == .sectionRoot(.properties))
        navigation.goForward()
        navigation.goForward()
        #expect(navigation.currentLocation.propertyId == "prop-name")

        // Category and property are both identity, so a strip change is its own entry
        // and title fluff still coalesces.
        #expect(
            WorkspaceLocation(section: .properties, subjectTypeKey: "person", propertyId: "prop-name")
                != WorkspaceLocation(section: .properties, propertyId: "prop-name")
        )
        #expect(
            WorkspaceLocation(section: .properties, subjectTypeKey: "person", propertyId: "prop-name", title: "A")
                == WorkspaceLocation(section: .properties, subjectTypeKey: "person", propertyId: "prop-name", title: "B")
        )

        let reloaded = WorkspaceNavigation()
        reloaded.attachProject(uuid: projectUuid, fileURL: url)
        #expect(reloaded.currentLocation.subjectTypeKey == "person")
        #expect(reloaded.currentLocation.propertyId == "prop-name")
    }

    @Test(arguments: [
        ("files", WorkspaceSection.sources),
        ("source-fields", WorkspaceSection.metadata),
        ("subject-fields", WorkspaceSection.properties),
    ])
    func retiredSectionIdsDecodeToTheirSection(id: String, section: WorkspaceSection) throws {
        let json = Data("""
        {"section":"\(id)"}
        """.utf8)
        let location = try JSONDecoder().decode(WorkspaceLocation.self, from: json)
        #expect(location == .sectionRoot(section))
        #expect(location.subjectTypeKey == nil)
        #expect(location.propertyId == nil)
        #expect(WorkspaceSection(id: id) == section)
    }

    @Test func currentSectionIdsRoundTrip() throws {
        for section in WorkspaceSection.allCases {
            #expect(WorkspaceSection(id: section.rawValue) == section)
            let data = try JSONEncoder().encode(WorkspaceLocation.sectionRoot(section))
            #expect(try JSONDecoder().decode(WorkspaceLocation.self, from: data) == .sectionRoot(section))
        }
        #expect(WorkspaceSection(id: "nope") == nil)
    }

    /// A history file saved before S9-07b keeps its entries and raises no
    /// "history couldn't load" reset.
    @Test func historyWithRetiredSectionIdsLoadsIntact() throws {
        let url = try tempNavigationFile()
        let json = """
        {"v":1,"projectUuid":"\(projectUuid)","index":1,"entries":[{"section":"source-fields"},{"section":"subject-fields"}]}
        """
        try Data(json.utf8).write(to: url)
        let (navigation, _) = try attachedNavigation(fileURL: url)
        #expect(navigation.lastHistoryIssue == nil)
        #expect(navigation.currentLocation == .sectionRoot(.properties))
        navigation.goBack()
        #expect(navigation.currentLocation == .sectionRoot(.metadata))
    }

    @Test func engineLocationWithRetiredSectionIdResolves() {
        var proto = Provenencia_Engine_V1_WorkspaceLocation()
        proto.section = "source-fields"
        proto.fieldID = "fld-1"
        let location = GoStore.mapWorkspaceLocationFromProto(proto)
        #expect(location.section == .metadata)
        #expect(location.fieldId == "fld-1")
        proto.section = "metadata"
        #expect(GoStore.mapWorkspaceLocationFromProto(proto).section == .metadata)
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
        var scrolled = WorkspaceLocation.sectionRoot(.persons)
        scrolled.listScrollAnchor = "person-1"
        #expect(scrolled == .sectionRoot(.persons))
        #expect(a != WorkspaceLocation(section: .sources, sourceId: "2"))
        #expect(
            WorkspaceLocation(section: .sources, sourceId: "1", sourceSurface: .page)
                != WorkspaceLocation(section: .sources, sourceId: "1", graphId: "g-1", sourceSurface: .graph)
        )
    }

    @Test func legacyHistoryMissingSourceSurfaceDecodesAsPage() throws {
        let json = Data("""
        {"section":"sources","sourceId":"src-legacy","title":"Deed"}
        """.utf8)
        let location = try JSONDecoder().decode(WorkspaceLocation.self, from: json)
        #expect(location.sourceId == "src-legacy")
        #expect(location.sourceSurface == .page)
        #expect(location.title == "Deed")
        #expect(location.listScrollAnchor == nil)
    }

    /// T-12 (AC-1.5): canvas, composer and Promote locations keep their graph
    /// through the history file.
    @Test func graphLocationsRoundTripTheGraph() throws {
        let (navigation, url) = try attachedNavigation()
        navigation.go(to: .evidenceGraph(sourceId: "src-1", graphId: "g-1", title: "Deed"))
        navigation.go(to: .citationComposer(sourceId: "src-1", graphId: "g-1", subjectId: "sub-1"))
        navigation.go(to: .promote(
            sourceId: "src-1", graphId: "g-1", subjectId: "sub-1", kind: .person,
            ref: "CPR-1", title: "James", sourceTitle: "Deed"
        ))
        let reloaded = WorkspaceNavigation()
        reloaded.attachProject(uuid: projectUuid, fileURL: url)
        for surface in [SourceSurface.promote, .citationComposer, .graph] {
            #expect(reloaded.currentLocation.sourceSurface == surface)
            #expect(reloaded.currentLocation.graphId == "g-1")
            #expect(reloaded.currentLocation.sourceId == "src-1")
            reloaded.goBack()
        }
    }

    /// T-12 (AC-1.6): history written before graph ids named a Source with a
    /// graph surface and no graph; it decodes as that Source's page.
    @Test func legacyGraphSurfaceWithoutGraphDecodesAsPage() throws {
        for surface in ["graph", "citationComposer", "promote"] {
            let json = Data("""
            {"section":"sources","sourceId":"src-legacy","subjectId":"sub-1","sourceSurface":"\(surface)"}
            """.utf8)
            let location = try JSONDecoder().decode(WorkspaceLocation.self, from: json)
            #expect(location.sourceSurface == .page, "\(surface)")
            #expect(location.sourceId == "src-legacy")
            #expect(location.graphId == nil)
        }
        let withGraph = try JSONDecoder().decode(WorkspaceLocation.self, from: Data("""
        {"section":"sources","sourceId":"src-1","graphId":"g-1","sourceSurface":"graph"}
        """.utf8))
        #expect(withGraph.sourceSurface == .graph)
    }

    /// T-23 (AC-1.6, D-11): every way a location is built applies the same rule.
    @Test func graphSurfaceWithoutGraphBecomesThePage() {
        for surface in [SourceSurface.graph, .citationComposer, .promote] {
            for graphId in [nil, ""] as [String?] {
                let location = WorkspaceLocation(
                    section: .sources, sourceId: "src-1", graphId: graphId, sourceSurface: surface
                )
                #expect(location.sourceSurface == .page)
                #expect(location.sourceId == "src-1")
            }
            #expect(
                WorkspaceLocation(section: .sources, sourceId: "src-1", graphId: "g-1", sourceSurface: surface)
                    .sourceSurface == surface
            )
        }
        var proto = Provenencia_Engine_V1_WorkspaceLocation()
        proto.section = "sources"
        proto.sourceID = "src-1"
        proto.sourceSurface = "graph"
        let withoutGraph = GoStore.mapWorkspaceLocationFromProto(proto)
        #expect(withoutGraph.sourceSurface == .page)
        #expect(withoutGraph.sourceId == "src-1")
        proto.graphID = "g-1"
        let withGraph = GoStore.mapWorkspaceLocationFromProto(proto)
        #expect(withGraph.sourceSurface == .graph)
        #expect(withGraph.graphId == "g-1")
    }

    @Test func legacyHistoryMissingListScrollAnchorDecodesAsNil() throws {
        let json = Data("""
        {"v":1,"projectUuid":"\(projectUuid)","index":0,"entries":[{"section":"persons"}]}
        """.utf8)
        let document = try JSONDecoder().decode(NavigationHistoryDocument.self, from: json)
        #expect(document.entries.count == 1)
        #expect(document.entries[0].listScrollAnchor == nil)
        #expect(document.entries[0] == .sectionRoot(.persons))
    }

    @Test func listScrollAnchorRoundTrips() throws {
        var location = WorkspaceLocation.sectionRoot(.sources)
        location.listScrollAnchor = "src-1"
        let decoded = try JSONDecoder().decode(
            WorkspaceLocation.self,
            from: JSONEncoder().encode(location)
        )
        #expect(decoded.listScrollAnchor == "src-1")
        #expect(decoded == location)
    }

    @Test func notingListScrollDoesNotPushHistory() throws {
        let (navigation, _) = try attachedNavigation()
        navigation.go(to: .sectionRoot(.persons))
        let count = entryCount(navigation)
        navigation.noteListScroll("person-1")
        navigation.noteListScroll("person-2")
        #expect(entryCount(navigation) == count)
        #expect(!navigation.canGoForward)
        #expect(navigation.currentLocation.listScrollAnchor == nil)
    }

    @Test func leavingAListStampsTheOutgoingEntry() throws {
        let (navigation, url) = try attachedNavigation()
        navigation.go(to: .sectionRoot(.persons))
        navigation.noteListScroll("person-1")
        navigation.go(to: .personDetail(entityId: "p1", ref: "PER-1", title: "Ada"))
        #expect(navigation.currentLocation.listScrollAnchor == nil)

        navigation.goBack()
        #expect(navigation.currentLocation.section == .persons)
        #expect(navigation.currentLocation.listScrollAnchor == "person-1")

        navigation.goForward()
        #expect(navigation.currentLocation.entityId == "p1")
        navigation.goBack()
        #expect(navigation.currentLocation.listScrollAnchor == "person-1")

        let reloaded = WorkspaceNavigation()
        reloaded.attachProject(uuid: projectUuid, fileURL: url)
        #expect(reloaded.currentLocation.listScrollAnchor == "person-1")
    }

    @Test func forwardAndJumpRestoreTheStampedAnchor() throws {
        let (navigation, _) = try attachedNavigation()
        navigation.go(to: .sectionRoot(.persons))
        navigation.noteListScroll("person-1")
        navigation.goBack()
        navigation.goForward()
        #expect(navigation.currentLocation == .sectionRoot(.persons))
        #expect(navigation.currentLocation.listScrollAnchor == "person-1")

        navigation.noteListScroll("person-4")
        navigation.go(to: .sectionRoot(.places))
        let persons = try #require(
            navigation.backJumpItems(limit: 15).first { $0.location.section == .persons }
        )
        navigation.go(toIndex: persons.index)
        #expect(navigation.currentLocation == .sectionRoot(.persons))
        #expect(navigation.currentLocation.listScrollAnchor == "person-4")
    }

    @Test func revisitingAListWithoutAnAnchorClearsItAndDoesNotPush() throws {
        let (navigation, _) = try attachedNavigation()
        navigation.go(to: .sectionRoot(.persons))
        navigation.noteListScroll("person-1")
        navigation.go(to: .personDetail(entityId: "p1", ref: "PER-1", title: "Ada"))
        navigation.goBack()
        #expect(navigation.currentLocation.listScrollAnchor == "person-1")
        let count = entryCount(navigation)

        navigation.go(to: .sectionRoot(.persons))
        #expect(navigation.currentLocation.listScrollAnchor == nil)
        #expect(entryCount(navigation) == count)
        #expect(navigation.canGoForward)

        navigation.goBack()
        navigation.goForward()
        #expect(navigation.currentLocation == .sectionRoot(.persons))
        #expect(navigation.currentLocation.listScrollAnchor == nil)
    }

    @Test func goToGraphIsDistinctPlaceFromPage() throws {
        let (navigation, _) = try attachedNavigation()
        navigation.go(to: WorkspaceLocation(section: .sources, sourceId: "src-1", sourceSurface: .page))
        navigation.go(to: WorkspaceLocation(section: .sources, sourceId: "src-1", graphId: "g-1", sourceSurface: .graph))
        #expect(navigation.currentLocation.sourceSurface == .graph)
        #expect(navigation.selectedSection == .sources)
        #expect(navigation.canGoBack)
        navigation.goBack()
        #expect(navigation.currentLocation.sourceSurface == .page)
        #expect(navigation.currentLocation.sourceId == "src-1")
    }

    @Test func goToCitationComposerRestoresSubjectId() throws {
        let (navigation, _) = try attachedNavigation()
        navigation.go(to: WorkspaceLocation(section: .sources, sourceId: "src-1", graphId: "g-1", sourceSurface: .graph))
        navigation.go(
            to: WorkspaceLocation(
                section: .sources,
                sourceId: "src-1",
                graphId: "g-1",
                subjectId: "sub-9",
                citationId: "cit-1",
                sourceSurface: .citationComposer,
                title: "Margt."
            )
        )
        #expect(navigation.currentLocation.sourceSurface == .citationComposer)
        #expect(navigation.currentLocation.subjectId == "sub-9")
        #expect(navigation.currentLocation.citationId == "cit-1")
        navigation.goBack()
        #expect(navigation.currentLocation.sourceSurface == .graph)
        #expect(navigation.currentLocation.subjectId == nil)
        #expect(
            WorkspaceLocation(
                section: .sources,
                sourceId: "src-1",
                graphId: "g-1",
                subjectId: "sub-9",
                sourceSurface: .citationComposer
            ) != WorkspaceLocation(
                section: .sources,
                sourceId: "src-1",
                graphId: "g-1",
                subjectId: "sub-9",
                citationId: "cit-1",
                sourceSurface: .citationComposer
            )
        )
    }

    @Test func goToCitationComposerRestoresObservationId() throws {
        let (navigation, _) = try attachedNavigation()
        navigation.go(to: WorkspaceLocation(section: .sources, sourceId: "src-1", graphId: "g-1", sourceSurface: .graph))
        navigation.go(
            to: WorkspaceLocation(
                section: .sources,
                sourceId: "src-1",
                graphId: "g-1",
                subjectId: "sub-9",
                citationId: "cit-1",
                artifactId: "art-0",
                observationId: "obs-row",
                sourceSurface: .citationComposer
            )
        )
        #expect(navigation.currentLocation.observationId == "obs-row")
        #expect(navigation.currentLocation.artifactId == "art-0")
        navigation.goBack()
        #expect(navigation.currentLocation.observationId == nil)
        #expect(
            WorkspaceLocation(
                section: .sources,
                sourceId: "src-1",
                graphId: "g-1",
                subjectId: "sub-9",
                citationId: "cit-1",
                observationId: "obs-a",
                sourceSurface: .citationComposer
            ) != WorkspaceLocation(
                section: .sources,
                sourceId: "src-1",
                graphId: "g-1",
                subjectId: "sub-9",
                citationId: "cit-1",
                observationId: "obs-b",
                sourceSurface: .citationComposer
            )
        )
    }

    @Test func navigationFileNameStripsDashesAndBraces() {
        #expect(
            InstallPaths.navigationFileName(projectUuid: "{00000000-0000-7000-8000-0000000000AA}")
                == "000000000000700080000000000000aa.json"
        )
    }

    @Test func backJumpItemsNearestFirstAndCapped() throws {
        let (navigation, _) = try attachedNavigation()
        // indices: 0 sources, 1 fields, 2 types, 3 source detail, 4 deep source — current at 4
        navigation.go(to: .sectionRoot(.metadata))
        navigation.go(to: .sectionRoot(.sourceTypes))
        navigation.go(to: WorkspaceLocation(section: .sources, sourceId: "src-0"))
        navigation.go(to: WorkspaceLocation(section: .sources, sourceId: "src-1", title: "Deep"))

        let back = navigation.backJumpItems(limit: 2)
        #expect(back.map(\.index) == [3, 2])
        #expect(back[0].location.sourceId == "src-0")
        #expect(back[1].location == .sectionRoot(.sourceTypes))

        let allBack = navigation.backJumpItems(limit: 15)
        #expect(allBack.map(\.index) == [3, 2, 1, 0])
    }

    @Test func forwardJumpItemsStackOrderAndCapped() throws {
        let (navigation, _) = try attachedNavigation()
        navigation.go(to: .sectionRoot(.metadata))
        navigation.go(to: .sectionRoot(.sourceTypes))
        navigation.go(to: WorkspaceLocation(section: .sources, sourceId: "src-0"))
        navigation.goBack()
        navigation.goBack()
        // current at metadata (index 1); forward: types (2), source detail (3)
        #expect(navigation.currentLocation == .sectionRoot(.metadata))

        let forward = navigation.forwardJumpItems(limit: 1)
        #expect(forward.map(\.index) == [2])
        #expect(forward[0].location == .sectionRoot(.sourceTypes))

        let allForward = navigation.forwardJumpItems(limit: 15)
        #expect(allForward.map(\.index) == [2, 3])
        #expect(allForward[1].location.sourceId == "src-0")
    }

    @Test func jumpItemsEmptyAtStackEnds() throws {
        let (navigation, _) = try attachedNavigation()
        #expect(navigation.backJumpItems().isEmpty)
        #expect(navigation.forwardJumpItems().isEmpty)

        navigation.go(to: .sectionRoot(.metadata))
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
        navigation.go(to: .sectionRoot(.metadata))
        #expect(navigation.currentLocation == .sectionRoot(.metadata))
        #expect(navigation.lastHistoryIssue == .persistFailed)
    }

    @Test func missingHistoryFileDoesNotReportLoadIssue() throws {
        let (navigation, _) = try attachedNavigation()
        #expect(navigation.lastHistoryIssue == nil)
        #expect(navigation.currentLocation == .sectionRoot(.sources))
    }

    @Test func locationCommitOnGo() {
        let navigation = WorkspaceNavigation()
        var committed: [WorkspaceLocation] = []
        navigation.onLocationCommit = { committed.append($0) }

        navigation.go(to: .sectionRoot(.metadata))
        #expect(committed == [.sectionRoot(.metadata)])
    }

    @Test func locationCommitOnAttachProjectRestore() throws {
        let (navigation, url) = try attachedNavigation()
        navigation.go(to: WorkspaceLocation(section: .sources, sourceId: "src-1", title: "Deed"))
        navigation.go(to: WorkspaceLocation(section: .sourceTypes, typeId: "typ-1", title: "Book"))
        navigation.goBack()

        var committed: [WorkspaceLocation] = []
        let reloaded = WorkspaceNavigation()
        reloaded.onLocationCommit = { committed.append($0) }
        reloaded.attachProject(uuid: projectUuid, fileURL: url)

        #expect(committed.count == 1)
        #expect(committed[0].sourceId == "src-1")
    }

    @Test func locationCommitOnBackForward() throws {
        let (navigation, _) = try attachedNavigation()
        var committed: [WorkspaceLocation] = []
        navigation.onLocationCommit = { committed.append($0) }

        navigation.go(to: .sectionRoot(.metadata))
        navigation.go(to: .sectionRoot(.sourceTypes))
        committed.removeAll()

        navigation.goBack()
        #expect(committed.last == .sectionRoot(.metadata))

        navigation.goForward()
        #expect(committed.last == .sectionRoot(.sourceTypes))

        committed.removeAll()
        navigation.go(toIndex: 0)
        #expect(committed.last == .sectionRoot(.sources))
    }

    @Test func leaveGuardHoldsAndResumes() throws {
        let (navigation, _) = try attachedNavigation()
        let guardBox = HoldGuard()
        navigation.leaveGuard = guardBox
        navigation.go(to: .sectionRoot(.metadata))
        #expect(navigation.currentLocation == .sectionRoot(.sources))
        #expect(navigation.heldNavigation == .location(.sectionRoot(.metadata)))
        navigation.resumeHeldNavigation()
        #expect(navigation.currentLocation == .sectionRoot(.metadata))
        #expect(navigation.heldNavigation == nil)
    }

    @Test func leaveGuardCancelKeepsPlace() throws {
        let (navigation, _) = try attachedNavigation()
        navigation.go(to: .sectionRoot(.metadata))
        let guardBox = HoldGuard()
        navigation.leaveGuard = guardBox
        navigation.goBack()
        #expect(navigation.currentLocation == .sectionRoot(.metadata))
        #expect(navigation.heldNavigation == .back)
        navigation.cancelHeldNavigation()
        #expect(navigation.currentLocation == .sectionRoot(.metadata))
        #expect(navigation.heldNavigation == nil)
    }

    @Test func leaveGuardDoesNotHoldIdenticalLocation() throws {
        let (navigation, _) = try attachedNavigation()
        let guardBox = HoldGuard()
        navigation.leaveGuard = guardBox
        navigation.go(to: .sectionRoot(.sources))
        #expect(navigation.heldNavigation == nil)
        #expect(navigation.currentLocation == .sectionRoot(.sources))
    }
}

@MainActor
private final class HoldGuard: WorkspaceLeaveGuard {
    func shouldHoldNavigation(_ pending: PendingNavigation) -> Bool {
        true
    }
}
