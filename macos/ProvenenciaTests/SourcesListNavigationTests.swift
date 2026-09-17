import Foundation
import Testing
@testable import Provenencia

@Suite
struct SourcesListNavigationTests {
    private func source(id: String = "s1", hasArtifact: Bool) -> CatalogSource {
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

    @Test func graphLocationWhenArtifactPresent() {
        let location = SourcesListNavigation.graphLocation(for: source(hasArtifact: true))
        #expect(location?.sourceSurface == .graph)
        #expect(location?.sourceId == "s1")
    }

    @Test func graphLocationBlockedWithoutArtifact() {
        #expect(SourcesListNavigation.graphLocation(for: source(hasArtifact: false)) == nil)
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
