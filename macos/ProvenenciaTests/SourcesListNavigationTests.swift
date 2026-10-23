import Foundation
import Testing
@testable import Provenencia

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

    private func graph(
        _ id: String, name: String, sourceID: String = "s1", artifactIDs: [String] = ["a1"]
    ) -> CatalogEvidenceGraph {
        CatalogEvidenceGraph(
            id: id, ref: "GPH-\(id)", sourceID: sourceID, name: name, description: "", artifactIDs: artifactIDs
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
            graph("g-1", name: "Default"),
            graph("g-2", name: "Jones", artifactIDs: ["a2"]),
            graph("g-9", name: "Other Source's", sourceID: "s2"),
        ]
        let location = SourcesListNavigation.graphLocation(for: source(), graphs: graphs)
        #expect(location?.sourceSurface == .graph)
        #expect(location?.sourceId == "s1")
        #expect(location?.graphId == "g-1")
        #expect(location?.ref == "SRC-AAAAA")
        #expect(location?.title == "Passport")
        #expect(SourcesListNavigation.entryGraph(sourceID: "s2", in: graphs)?.id == "g-9")
    }

    @Test func graphLocationBlockedWithoutArtifact() {
        // The entry graph links no Artifact: Open stays blocked, though a later graph links one.
        let noLinks = [graph("g-1", name: "Default", artifactIDs: []), graph("g-2", name: "Jones")]
        #expect(SourcesListNavigation.graphLocation(for: source(), graphs: noLinks) == nil)
        // The Source has no graph in the list (not loaded, or none).
        #expect(SourcesListNavigation.graphLocation(for: source(), graphs: []) == nil)
        #expect(SourcesListNavigation.graphLocation(for: source(id: "s3"), graphs: [graph("g-1", name: "Default")]) == nil)
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
