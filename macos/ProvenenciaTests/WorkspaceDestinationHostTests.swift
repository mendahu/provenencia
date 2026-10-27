import AppKit
import Foundation
import SwiftUI
import Testing
@testable import Provenencia

@MainActor @Suite
struct WorkspaceDestinationHostTests {
    private let projectDir = "/tmp/workspace-destination-host.provenencia"

    private func presentation(for location: WorkspaceLocation) -> WorkspacePresentationID {
        WorkspaceDestinationHost.place(for: location, projectDir: projectDir)?.presentation ?? .sourcesList
    }

    /// Two census Sources, each with one graph linked to an Artifact; the canvas reads this store.
    private func makeCanvasStore() -> FakeStore {
        let store = FakeStore()
        store.subjectTypesByProject[projectDir] = [
            CatalogSubjectType(
                id: "type-person", key: "person", origin: "provenencia", label: "Person",
                description: "", refPrefix: "PER", candidateRefPrefix: "CPR"
            ),
        ]
        store.propertiesByProject[projectDir] = [
            CatalogProperty(
                id: "prop-role", key: "role", origin: "provenencia", label: "Role",
                description: "", valueType: "term"
            ),
        ]
        store.sourcesByProject[projectDir] = [
            CatalogSource(id: "src-1", ref: "SRC-1", sourceTypeID: "", title: "Jones census", description: ""),
            CatalogSource(id: "src-2", ref: "SRC-2", sourceTypeID: "", title: "Smith census", description: ""),
        ]
        store.evidenceGraphsByProject[projectDir] = [
            evidenceGraph("g-a", sourceID: "src-1", artifactIDs: ["art-1"]),
            evidenceGraph("g-b", sourceID: "src-2", artifactIDs: ["art-2"]),
        ]
        return store
    }

    /// Mounts the host in a borderless window so the canvas's `.task` runs.
    private func mountHost(
        navigation: WorkspaceNavigation,
        session: WorkspaceSession,
        store: FakeStore
    ) -> (NSWindow, NSHostingView<some View>) {
        let frame = NSRect(x: 0, y: 0, width: 1200, height: 800)
        let host = NSHostingView(
            rootView: WorkspaceDestinationHost(
                userID: "user-1",
                sessionDisplayName: "",
                store: store,
                catalogCounts: CatalogCounts(projectDir: projectDir, store: store)
            )
            .environment(navigation)
            .environment(session)
        )
        host.frame = frame
        let window = NSWindow(contentRect: frame, styleMask: .borderless, backing: .buffered, defer: false)
        window.contentView = host
        window.orderFrontRegardless()
        return (window, host)
    }

    @Test func canvasToCanvasBuildsTheSecondGraphsModel() async {
        let store = makeCanvasStore()
        let session = WorkspaceSession(projectKey: ProjectKey(projectDir: projectDir), store: store)
        let navigation = WorkspaceNavigation()
        navigation.go(to: .evidenceGraph(sourceId: "src-1", graphId: "g-a", title: "Jones"))
        let (window, host) = mountHost(navigation: navigation, session: session, store: store)

        let graphA = CatalogQueryKey.evidenceGraph(project: session.projectKey, graphId: "g-a")
        let loadedA = await waitUntil {
            (session.queryHandle(graphA) as QueryHandle<EvidenceGraphRows>?)?.isSettled == true
        }
        #expect(loadedA)

        navigation.go(to: .evidenceGraph(sourceId: "src-2", graphId: "g-b", title: "Smith"))
        let graphB = CatalogQueryKey.evidenceGraph(project: session.projectKey, graphId: "g-b")
        let workspaceB = CatalogQueryKey.sourceWorkspace(project: session.projectKey, sourceId: "src-2")
        let loadedB = await waitUntil {
            (session.queryHandle(graphB) as QueryHandle<EvidenceGraphRows>?)?.isSettled == true
                && (session.queryHandle(workspaceB) as QueryHandle<CatalogSourceWorkspace>?)?.isSettled == true
        }
        #expect(loadedB)
        withExtendedLifetime((window, host)) {}
    }

    @Test func goneGraphFallsBackToTheSourcesList() async throws {
        let store = makeCanvasStore()
        let session = WorkspaceSession(projectKey: ProjectKey(projectDir: projectDir), store: store)
        let directory = FileManager.default.temporaryDirectory
            .appendingPathComponent("host-history-\(UUID().uuidString)", isDirectory: true)
        try FileManager.default.createDirectory(at: directory, withIntermediateDirectories: true)
        let navigation = WorkspaceNavigation()
        navigation.attachProject(uuid: "project-1", fileURL: directory.appendingPathComponent("history.json"))
        navigation.go(to: .sectionRoot(.sources))
        navigation.go(to: .evidenceGraph(sourceId: "src-1", graphId: "g-x", title: "Gone"))
        let (window, host) = mountHost(navigation: navigation, session: session, store: store)

        let fellBack = await waitUntil { navigation.currentLocation == .sectionRoot(.sources) }
        #expect(fellBack)
        let remembered = navigation.backJumpItems(limit: 100) + navigation.forwardJumpItems(limit: 100)
        #expect(!remembered.contains { $0.location.graphId == "g-x" })
        withExtendedLifetime((window, host)) {}
    }

    @Test func presentationSourcesRoot() {
        #expect(presentation(for: .sectionRoot(.sources)) == .sourcesList)
    }

    @Test func presentationSourceDetail() {
        #expect(
            presentation(for: WorkspaceLocation(section: .sources, sourceId: "src-1", title: "Deed"))
                == .sourcePage
        )
    }

    @Test func presentationEvidenceGraph() {
        #expect(
            presentation(
                for: WorkspaceLocation(section: .sources, sourceId: "src-1", graphId: "g-1", sourceSurface: .graph)
            ) == .sourceGraph
        )
    }

    @Test func presentationPromote() {
        let location = WorkspaceLocation.promote(
            sourceId: "src-1", graphId: "g-1", subjectId: "sub-1", kind: .event, ref: "CEV-1", title: nil, sourceTitle: nil
        )
        #expect(presentation(for: location) == .sourcePromote)
        #expect(WorkspaceDestinationHost.destinationKind(for: .sourcePromote) == .sources)
    }

    @Test func presentationCitationComposer() {
        #expect(
            presentation(
                for: WorkspaceLocation(
                    section: .sources,
                    sourceId: "src-1",
                    graphId: "g-1",
                    subjectId: "sub-1",
                    sourceSurface: .citationComposer
                )
            ) == .sourceCitationComposer
        )
    }

    @Test func presentationCitationJumpComposer() {
        #expect(
            presentation(
                for: WorkspaceLocation(
                    section: .sources,
                    sourceId: "src-1",
                    graphId: "g-1",
                    citationId: "cit-1",
                    artifactId: "art-0",
                    sourceSurface: .citationComposer
                )
            ) == .sourceCitationComposer
        )
    }

    @Test func presentationConnectPrefillComposer() {
        #expect(
            presentation(
                for: WorkspaceLocation(
                    section: .sources,
                    sourceId: "src-1",
                    graphId: "g-1",
                    connectFromSubjectId: "p1",
                    connectToSubjectId: "e1",
                    connectBridgeTypeKey: "participation",
                    sourceSurface: .citationComposer
                )
            ) == .sourceCitationComposer
        )
    }

    @Test func presentationMetadataRootAndRow() {
        #expect(presentation(for: .sectionRoot(.metadata)) == .metadata)
        #expect(
            presentation(for: WorkspaceLocation(section: .metadata, fieldId: "fld-1", title: "Author"))
                == .metadata
        )
    }

    @Test func presentationSourceTypesRootAndRow() {
        #expect(presentation(for: .sectionRoot(.sourceTypes)) == .sourceTypes)
        #expect(
            presentation(for: WorkspaceLocation(section: .sourceTypes, typeId: "typ-1", title: "Book"))
                == .sourceTypes
        )
    }

    @Test func presentationProperties() {
        #expect(presentation(for: .sectionRoot(.properties)) == .properties)
    }

    @Test func sourceFamilyPresentationsUseSourcesDestination() {
        #expect(WorkspaceDestinationHost.destinationKind(for: .sourcesList) == .sources)
        #expect(WorkspaceDestinationHost.destinationKind(for: .sourcePage) == .sources)
        #expect(WorkspaceDestinationHost.destinationKind(for: .sourceGraph) == .sources)
        #expect(WorkspaceDestinationHost.destinationKind(for: .sourceCitationComposer) == .sources)
    }

    @Test func conclusionPresentationsHaveOwnKinds() {
        #expect(presentation(for: .sectionRoot(.persons)) == .personsList)
        #expect(presentation(for: .sectionRoot(.events)) == .eventsList)
        #expect(presentation(for: .sectionRoot(.places)) == .placesList)
        #expect(WorkspaceDestinationHost.destinationKind(for: .personsList) == .persons)
        #expect(presentation(for: .personDetail(entityId: "e1", ref: "PER-1", title: nil)) == .personDetail)
        #expect(WorkspaceDestinationHost.destinationKind(for: .personDetail) == .persons)
        #expect(WorkspaceDestinationHost.destinationKind(for: .eventsList) == .events)
        #expect(presentation(for: .eventDetail(entityId: "e1", ref: "EVT-1", title: "Fire")) == .eventDetail)
        #expect(WorkspaceDestinationHost.destinationKind(for: .eventDetail) == .events)
        #expect(WorkspaceDestinationHost.destinationKind(for: .placesList) == .places)
        #expect(presentation(for: .placeDetail(entityId: "e1", ref: "PLC-1", title: "York")) == .placeDetail)
        #expect(WorkspaceDestinationHost.destinationKind(for: .placeDetail) == .places)
    }

    @Test func propertiesPresentationHasOwnKind() {
        #expect(WorkspaceDestinationHost.destinationKind(for: .properties) == .properties)
    }

    @Test func registryPresentationsAreKnown() {
        let project = ProjectKey(projectDir: projectDir)
        let registry = PlaceRegistry.standard
        let locations: [WorkspaceLocation] = [
            .sectionRoot(.sources),
            WorkspaceLocation(section: .sources, sourceId: "s1"),
            WorkspaceLocation(section: .sources, sourceId: "s1", graphId: "g-1", sourceSurface: .graph),
            WorkspaceLocation(
                section: .sources,
                sourceId: "s1",
                graphId: "g-1",
                subjectId: "sub-1",
                sourceSurface: .citationComposer
            ),
            .sectionRoot(.metadata),
            WorkspaceLocation(section: .metadata, fieldId: "f1"),
            .sectionRoot(.sourceTypes),
            WorkspaceLocation(section: .sourceTypes, typeId: "t1"),
            .sectionRoot(.properties),
            .sectionRoot(.persons),
            .sectionRoot(.events),
            .sectionRoot(.places),
            .personDetail(entityId: "e1", ref: "PER-1", title: nil),
            .eventDetail(entityId: "e1", ref: "EVT-1", title: "Fire at York"),
            .placeDetail(entityId: "e1", ref: "PLC-1", title: "York"),
        ]
        let known = Set(WorkspacePresentationID.allCases)
        for location in locations {
            guard let place = registry.resolve(location, project: project) else {
                Issue.record("Expected place for \(location)")
                continue
            }
            #expect(known.contains(place.presentation))
        }
        #expect(known.count == 14)
    }
}