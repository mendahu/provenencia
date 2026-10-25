import Foundation
import Testing
@testable import Provenencia

@Suite
struct EvidenceGraphStoreTests {
    private let projectDir = "/tmp/evidence-graphs.provenencia"
    private let userID = "user-1"
    private let sourceID = "src-1"

    private func seededStore() -> FakeStore {
        let store = FakeStore()
        store.sourcesByProject[projectDir] = [
            CatalogSource(id: sourceID, ref: "SRC-AAAAA", sourceTypeID: "t", title: "Census", description: ""),
        ]
        return store
    }

    @Test func createThenListSortsByName() async throws {
        let store = seededStore()
        _ = try await store.createEvidenceGraph(
            projectDir: projectDir, userID: userID, sourceID: sourceID,
            name: "Smith", description: "", artifactIDs: ["a1"]
        )
        let (jones, _) = try await store.createEvidenceGraph(
            projectDir: projectDir, userID: userID, sourceID: sourceID,
            name: " Jones ", description: "Household 12", artifactIDs: ["a1"]
        )
        #expect(jones.name == "Jones")
        let listed = try await store.listEvidenceGraphs(projectDir: projectDir)
        #expect(listed.map(\.name) == ["Jones", "Smith"])
        #expect(store.heldCatalogProjectDir == projectDir)
    }

    @Test func createArtifactLinksOnlyItsProjectsGraphs() async throws {
        let store = seededStore()
        let otherDir = "/tmp/other-evidence-graphs.provenencia"
        store.evidenceGraphsByProject[projectDir] = [evidenceGraph("g1", name: "Default", artifactIDs: [])]
        store.evidenceGraphsByProject[otherDir] = [evidenceGraph("g2", name: "Default", artifactIDs: [])]
        let (art, _) = try await store.createArtifact(
            projectDir: projectDir, userID: userID, sourceID: sourceID,
            fileID: "f1", label: "Page 1", description: "", graphIDs: []
        )
        #expect(store.evidenceGraphsByProject[projectDir]?.first?.artifactIDs == [art.id])
        #expect(store.evidenceGraphsByProject[otherDir]?.first?.artifactIDs == [])
    }

    @Test func writesReportTheGraphList() async throws {
        let store = seededStore()
        let expected: Set<CatalogList> = [.evidenceGraphs]
        let (made, created) = try await store.createEvidenceGraph(
            projectDir: projectDir, userID: userID, sourceID: sourceID,
            name: "Jones", description: "", artifactIDs: ["a1"]
        )
        let updated = try await store.updateEvidenceGraph(
            projectDir: projectDir, userID: userID, graphID: made.id,
            name: "Jones family", description: "", artifactIDs: ["a2"]
        )
        let linked = try await store.linkEvidenceGraphArtifact(
            projectDir: projectDir, userID: userID, graphID: made.id, artifactID: "a1"
        )
        let unlinked = try await store.unlinkEvidenceGraphArtifact(
            projectDir: projectDir, userID: userID, graphID: made.id, artifactID: "a1"
        )
        let deleted = try await store.deleteEvidenceGraph(projectDir: projectDir, userID: userID, graphID: made.id)
        for effects in [created, updated, linked, unlinked, deleted] {
            #expect(effects.lists == expected)
        }
        #expect(try await store.listEvidenceGraphs(projectDir: projectDir).isEmpty)
    }

    @Test func updateKeepsLinksWhenNoneGiven() async throws {
        let store = seededStore()
        store.evidenceGraphsByProject[projectDir] = [evidenceGraph("g1", name: "Default", artifactIDs: ["a1", "a2"])]
        _ = try await store.updateEvidenceGraph(
            projectDir: projectDir, userID: userID, graphID: "g1", name: "Household", description: "d", artifactIDs: []
        )
        let listed = try await store.listEvidenceGraphs(projectDir: projectDir)
        #expect(listed[0].name == "Household")
        #expect(listed[0].artifactIDs == ["a1", "a2"])
    }

    @Test func refusalsThrowForUnknownGraphsAndEmptyInput() async {
        let store = seededStore()
        await #expect(throws: (any Error).self) {
            _ = try await store.createEvidenceGraph(
                projectDir: projectDir, userID: userID, sourceID: sourceID,
                name: "  ", description: "", artifactIDs: ["a1"]
            )
        }
        await #expect(throws: (any Error).self) {
            _ = try await store.createEvidenceGraph(
                projectDir: projectDir, userID: userID, sourceID: sourceID,
                name: "Jones", description: "", artifactIDs: []
            )
        }
        await #expect(throws: (any Error).self) {
            _ = try await store.deleteEvidenceGraph(projectDir: projectDir, userID: userID, graphID: "nope")
        }
    }

    @Test func createArtifactLinksEveryGraphWhenNoneChosen() async throws {
        let store = seededStore()
        store.evidenceGraphsByProject[projectDir] = [
            evidenceGraph("g1", name: "Default", artifactIDs: ["a0"]),
            evidenceGraph("g2", name: "Jones", artifactIDs: ["a0"]),
            evidenceGraph("g3", sourceID: "src-2", name: "Other", artifactIDs: ["b0"]),
        ]
        let (every, effects) = try await store.createArtifact(
            projectDir: projectDir, userID: userID, sourceID: sourceID,
            fileID: "", label: "Page 2", description: "", graphIDs: []
        )
        #expect(effects.lists.contains(.evidenceGraphs))
        let (chosen, _) = try await store.createArtifact(
            projectDir: projectDir, userID: userID, sourceID: sourceID,
            fileID: "", label: "Page 3", description: "", graphIDs: ["g2"]
        )
        let byID = Dictionary(
            uniqueKeysWithValues: try await store.listEvidenceGraphs(projectDir: projectDir).map { ($0.id, $0) }
        )
        #expect(byID["g1"]?.artifactIDs == ["a0", every.id])
        #expect(byID["g2"]?.artifactIDs == ["a0", every.id, chosen.id])
        #expect(byID["g3"]?.artifactIDs == ["b0"])
    }

    /// T-20 (D-14): a new Source has a Default graph, like the engine's, and its
    /// first Artifact links to it; a copy's graph links the copied Artifacts.
    @Test func createSourceMakesADefaultGraph() async throws {
        let store = FakeStore()
        let (created, createdEffects) = try await store.createSource(
            projectDir: projectDir, userID: userID, sourceTypeID: "t", title: "Deed", description: ""
        )
        let graphs = try await store.listEvidenceGraphs(projectDir: projectDir)
        #expect(createdEffects.evidenceGraphs == graphs.map(\.id))
        #expect(graphs.map(\.name) == ["Default"])
        #expect(graphs.first?.sourceID == created.id)
        #expect(graphs.first?.id != created.id)
        #expect(graphs.first?.artifactIDs == [])

        let (artifact, _) = try await store.createArtifact(
            projectDir: projectDir, userID: userID, sourceID: created.id,
            fileID: "", label: "Page 1", description: "", graphIDs: []
        )
        let linked = try await store.listEvidenceGraphs(projectDir: projectDir)
        #expect(linked.first?.artifactIDs == [artifact.id])

        let (copy, _) = try await store.copySource(
            projectDir: projectDir, userID: userID, sourceID: created.id, title: "Deed copy", sourceTypeID: "t",
            copyDescription: false, copyMetadata: false, copyCredibility: false, artifactIDs: [artifact.id]
        )
        let all = try await store.listEvidenceGraphs(projectDir: projectDir)
        let copied = try #require(all.first { $0.sourceID == copy.id })
        #expect(copied.name == "Default")
        #expect(copied.artifactIDs.count == 1)
        #expect(copied.artifactIDs != [artifact.id])
    }

    @Test func createSubjectOnAnUnknownGraphThrows() async {
        let store = seededStore()
        await #expect(throws: (any Error).self) {
            _ = try await store.createSubject(
                projectDir: projectDir, userID: userID, graphID: sourceID,
                subjectTypeID: "t", label: "Ann", description: "", placement: nil
            )
        }
    }

    @Test func deletesDropLinksAndGraphs() async throws {
        let store = seededStore()
        store.artifactsBySource[sourceID] = [
            CatalogArtifact(id: "a1", ref: "ART-1", sourceID: sourceID, fileID: "", label: "P1", description: "", file: nil),
        ]
        store.evidenceGraphsByProject[projectDir] = [
            evidenceGraph("g1", name: "Default", artifactIDs: ["a1", "a2"]),
            evidenceGraph("g2", name: "Jones", artifactIDs: ["a1"]),
        ]
        _ = try await store.deleteArtifact(projectDir: projectDir, userID: userID, artifactID: "a1")
        let afterArtifact = try await store.listEvidenceGraphs(projectDir: projectDir)
        #expect(afterArtifact.map(\.artifactIDs) == [["a2"], []])

        try await store.deleteSource(projectDir: projectDir, userID: userID, sourceID: sourceID)
        #expect(try await store.listEvidenceGraphs(projectDir: projectDir).isEmpty)
    }

    @Test func unlinkImpactFiltersToAskedArtifacts() async throws {
        let store = seededStore()
        let group = { (via: String) in
            CatalogDeleteImpactGroup(via: via, kind: "citation", total: 1, listed: [])
        }
        store.unlinkImpactBlocks = [
            CatalogEvidenceGraphUnlinkBlock(artifactID: "a1", citations: group("evidencegraphs.artifact_in_use")),
            CatalogEvidenceGraphUnlinkBlock(artifactID: "a2", citations: group("evidencegraphs.artifact_in_use")),
        ]
        let blocks = try await store.getEvidenceGraphUnlinkImpact(
            projectDir: projectDir, graphID: "g1", artifactIDs: ["a2"]
        )
        #expect(blocks.map(\.artifactID) == ["a2"])
    }
}
