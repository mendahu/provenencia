import Foundation
import Testing
@testable import Provenencia

/// The graph list grouped by Source the way the session's index does.
private func graphsBySource(_ graphs: [CatalogEvidenceGraph]) -> [String: [CatalogEvidenceGraph]] {
    EvidenceGraphIndex(revision: 0, rows: graphs).bySource
}

@Suite
struct SourcesListNavigationTests {
    private func source(id: String = "s1", hasArtifact: Bool = true) -> CatalogSource {
        CatalogSource(
            id: id,
            ref: "SRC-AAAAA",
            sourceTypeID: "t1",
            title: "Passport",
            description: "",
            hasArtifact: hasArtifact
        )
    }

    @Test func pageLocationUsesPageSurface() {
        let location = SourcesListNavigation.pageLocation(for: source(hasArtifact: false))
        #expect(location.section == .sources)
        #expect(location.sourceId == "s1")
        #expect(location.sourceSurface == .page)
        #expect(location.ref == "SRC-AAAAA")
        #expect(location.title == "Passport")
    }

    @Test func graphLocationOpensTheFirstGraph() {
        // The graph list is sorted by name: Default comes before Jones.
        let graphs = [
            evidenceGraph("g-1", sourceID: "s1", name: "Default", artifactIDs: ["a1"]),
            evidenceGraph("g-2", sourceID: "s1", name: "Jones", artifactIDs: ["a2"]),
            evidenceGraph("g-9", sourceID: "s2", name: "Other Source's", artifactIDs: ["a1"]),
        ]
        let location = SourcesListNavigation.graphLocation(for: source(), graphs: graphsBySource(graphs))
        #expect(location?.sourceSurface == .graph)
        #expect(location?.sourceId == "s1")
        #expect(location?.graphId == "g-1")
        #expect(location?.ref == "SRC-AAAAA")
        #expect(location?.title == "Passport")
        #expect(SourcesListNavigation.entryGraph(sourceID: "s2", in: graphsBySource(graphs))?.id == "g-9")
    }

    @Test func graphBlockedReasonNeedsTheGraphList() {
        let reason = L10n.string(L10n.Sources.needsArtifactTooltip)
        let location = WorkspaceLocation.evidenceGraph(sourceId: "s1", graphId: "g-1", ref: "SRC-AAAAA", title: "Passport")
        #expect(SourcesListNavigation.graphBlockedReason(hasArtifact: false, graphListLoaded: false, location: nil) == reason)
        #expect(SourcesListNavigation.graphBlockedReason(hasArtifact: true, graphListLoaded: false, location: nil) == nil)
        #expect(SourcesListNavigation.graphBlockedReason(hasArtifact: true, graphListLoaded: true, location: nil) == reason)
        #expect(SourcesListNavigation.graphBlockedReason(hasArtifact: true, graphListLoaded: true, location: location) == nil)
    }

    @Test func graphLocationOpensTheGivenGraph() {
        let linked = evidenceGraph("g-2", sourceID: "s1", name: "Jones", artifactIDs: ["a2"])
        let location = SourcesListNavigation.graphLocation(for: source(id: "src-1"), graph: linked)
        #expect(location == .evidenceGraph(sourceId: "src-1", graphId: "g-2", ref: "SRC-AAAAA", title: "Passport"))
        #expect(location?.ref == "SRC-AAAAA")
        #expect(location?.title == "Passport")
        #expect(SourcesListNavigation.graphBlockedReason(for: linked) == nil)

        let bare = evidenceGraph("g-2", sourceID: "s1", name: "Jones")
        #expect(SourcesListNavigation.graphLocation(for: source(id: "src-1"), graph: bare) == nil)
        #expect(SourcesListNavigation.graphBlockedReason(for: bare)
            == L10n.string(L10n.Sources.graphNeedsArtifactTooltip))
    }

    @Test func graphLocationBlockedWithoutArtifact() {
        // The entry graph links no Artifact: Open stays blocked, though a later graph links one.
        let noLinks = [
            evidenceGraph("g-1", sourceID: "s1", name: "Default"),
            evidenceGraph("g-2", sourceID: "s1", name: "Jones", artifactIDs: ["a1"]),
        ]
        #expect(SourcesListNavigation.graphLocation(for: source(), graphs: graphsBySource(noLinks)) == nil)
        // The Source has no graph in the list (not loaded, or none).
        #expect(SourcesListNavigation.graphLocation(for: source(), graphs: [:]) == nil)
        let other = graphsBySource([
            evidenceGraph("g-1", sourceID: "s1", name: "Default", artifactIDs: ["a1"]),
        ])
        #expect(SourcesListNavigation.graphLocation(for: source(id: "s3"), graphs: other) == nil)
    }
}

@Suite
@MainActor
struct WorkspaceSessionGraphAccessorTests {
    private let projectDir = "/tmp/session-graphs.provenencia"

    @Test func sessionGraphAccessorsReadTheList() async {
        let store = FakeStore()
        store.evidenceGraphsByProject[projectDir] = [
            evidenceGraph("g-1", sourceID: "src-1", name: "Default"),
            evidenceGraph("g-2", sourceID: "src-1", name: "Jones"),
            evidenceGraph("g-3", sourceID: "src-2", name: "Default"),
        ]
        store.sourcesByProject[projectDir] = [
            CatalogSource(id: "src-2", ref: "SRC-BBBBB", sourceTypeID: "t1", title: "Other", description: ""),
        ]
        let session = WorkspaceSession(projectKey: ProjectKey(projectDir: projectDir), store: store)
        #expect(session.evidenceGraph(id: "g-2") == nil)
        #expect(session.evidenceGraphs(sourceID: "src-1") == nil)
        #expect(session.source(id: "src-2") == nil)

        let graphs: QueryHandle<[CatalogEvidenceGraph]> = session.query(session.evidenceGraphsKey)
        await waitForQuery(graphs)
        #expect(session.evidenceGraph(id: "g-2")?.name == "Jones")
        #expect(session.evidenceGraph(id: "missing") == nil)
        #expect(session.evidenceGraphs(sourceID: "src-1")?.map(\.id) == ["g-1", "g-2"])
        #expect(session.evidenceGraphs(sourceID: "src-9")?.isEmpty == true)

        #expect(session.source(id: "src-2") == nil)
        let sources: QueryHandle<[CatalogSource]> = session.query(.sourcesList(project: session.projectKey))
        await waitForQuery(sources)
        #expect(session.source(id: "src-2")?.title == "Other")
    }
}

@Suite
@MainActor
struct SourcesListArtifactGateTests {
    private let projectDir = "/tmp/sources-gate.provenencia"

    @Test func fakeStoreListSourcesReflectsArtifactPresence() async throws {
        let store = FakeStore()
        let bare = CatalogSource(
            id: "bare",
            ref: "SRC-BARE1",
            sourceTypeID: "t1",
            title: "Bare",
            description: ""
        )
        let withArt = CatalogSource(
            id: "full",
            ref: "SRC-FULL1",
            sourceTypeID: "t1",
            title: "Full",
            description: ""
        )
        store.sourcesByProject[projectDir] = [bare, withArt]
        store.artifactsBySource["full"] = [
            CatalogArtifact(
                id: "a1",
                ref: "ART-AAAAA",
                sourceID: "full",
                fileID: "",
                label: "Scan",
                description: "",
                file: nil
            )
        ]

        let listed = try await store.listSources(projectDir: projectDir)
        #expect(listed.first { $0.id == "bare" }?.hasArtifact == false)
        #expect(listed.first { $0.id == "full" }?.hasArtifact == true)
    }
}
