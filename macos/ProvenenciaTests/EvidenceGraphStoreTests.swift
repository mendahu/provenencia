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

    @Test func listReturnsSeedOrder() async throws {
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
        let listed = try await store.listEvidenceGraphs(projectDir: projectDir, sourceID: nil)
        #expect(listed.map(\.name) == ["Smith", "Jones"])
        #expect(store.heldCatalogProjectDir == projectDir)
    }

    @Test func listsOneSourcesGraphs() async throws {
        let store = seededStore()
        store.evidenceGraphsByProject[projectDir] = [
            evidenceGraph("g1", sourceID: "a", name: "Default"),
            evidenceGraph("g2", sourceID: "b", name: "Default"),
            evidenceGraph("g3", sourceID: "a", name: "Jones"),
        ]
        let ofA = try await store.listEvidenceGraphs(projectDir: projectDir, sourceID: "a")
        #expect(ofA.map(\.id) == ["g1", "g3"])
        let all = try await store.listEvidenceGraphs(projectDir: projectDir, sourceID: nil)
        #expect(all.map(\.id) == ["g1", "g2", "g3"])
        #expect(store.listEvidenceGraphsSourceIDs == ["a", nil])
    }

    @Test func writesReturnScriptedEffects() async throws {
        let store = seededStore()
        store.nextWriteEffects = CatalogWriteEffects(lists: [.evidenceGraphs])
        let expected: Set<CatalogList> = [.evidenceGraphs]
        let (made, created) = try await store.createEvidenceGraph(
            projectDir: projectDir, userID: userID, sourceID: sourceID,
            name: "Jones", description: "", artifactIDs: ["a1"]
        )
        let updated = try await store.updateEvidenceGraph(
            projectDir: projectDir, userID: userID, graphID: made.id,
            name: "Jones family", description: "", artifactIDs: ["a2"]
        )
        let deleted = try await store.deleteEvidenceGraph(projectDir: projectDir, userID: userID, graphID: made.id)
        for effects in [created, updated, deleted] {
            #expect(effects.lists == expected)
        }
        #expect(try await store.listEvidenceGraphs(projectDir: projectDir, sourceID: nil).isEmpty)
    }

    @Test func updateKeepsLinksWhenNoneGiven() async throws {
        let store = seededStore()
        store.evidenceGraphsByProject[projectDir] = [evidenceGraph("g1", name: "Default", artifactIDs: ["a1", "a2"])]
        _ = try await store.updateEvidenceGraph(
            projectDir: projectDir, userID: userID, graphID: "g1", name: "Household", description: "d", artifactIDs: []
        )
        let listed = try await store.listEvidenceGraphs(projectDir: projectDir, sourceID: nil)
        #expect(listed[0].name == "Household")
        #expect(listed[0].artifactIDs == ["a1", "a2"])
    }

    @Test func copySourceErrorIsScripted() async throws {
        let store = seededStore()
        struct Boom: Error {}
        store.copySourceError = Boom()
        await #expect(throws: (any Error).self) {
            _ = try await store.copySource(
                projectDir: projectDir, userID: userID, sourceID: sourceID, title: "Copy",
                sourceTypeID: "t", copyDescription: false, copyMetadata: false,
                copyCredibility: false, artifactIDs: []
            )
        }
        store.copySourceError = nil
        let (copy, _) = try await store.copySource(
            projectDir: projectDir, userID: userID, sourceID: sourceID, title: "",
            sourceTypeID: "t", copyDescription: false, copyMetadata: false,
            copyCredibility: false, artifactIDs: []
        )
        #expect(copy.title == "")
        #expect(copy.id != sourceID)
    }

    @Test func createEvidenceGraphErrorIsScripted() async {
        let store = seededStore()
        struct Boom: Error {}
        store.createEvidenceGraphError = Boom()
        await #expect(throws: (any Error).self) {
            _ = try await store.createEvidenceGraph(
                projectDir: projectDir, userID: userID, sourceID: sourceID,
                name: "Jones", description: "", artifactIDs: ["a1"]
            )
        }
        store.createEvidenceGraphError = nil
        _ = try? await store.createEvidenceGraph(
            projectDir: projectDir, userID: userID, sourceID: sourceID,
            name: "  ", description: "", artifactIDs: []
        )
        await #expect(throws: (any Error).self) {
            _ = try await store.deleteEvidenceGraph(projectDir: projectDir, userID: userID, graphID: "nope")
        }
    }

    @Test func createArtifactRecordsGraphIDsWithoutLinking() async throws {
        let store = seededStore()
        let otherDir = "/tmp/other-evidence-graphs.provenencia"
        store.evidenceGraphsByProject[projectDir] = [
            evidenceGraph("g1", name: "Default", artifactIDs: ["a0"]),
            evidenceGraph("g2", name: "Jones", artifactIDs: ["a0"]),
        ]
        store.evidenceGraphsByProject[otherDir] = [evidenceGraph("g2-other", name: "Default", artifactIDs: ["b0"])]
        let (nilLinks, _) = try await store.createArtifact(
            projectDir: projectDir, userID: userID, sourceID: sourceID,
            fileID: "f1", label: "Page 1", description: "", graphIDs: nil
        )
        #expect(store.createArtifactGraphIDs == [nil])
        #expect(store.evidenceGraphsByProject[projectDir]?.map(\.artifactIDs) == [["a0"], ["a0"]])
        let (chosen, _) = try await store.createArtifact(
            projectDir: projectDir, userID: userID, sourceID: sourceID,
            fileID: "", label: "Page 2", description: "", graphIDs: ["g2"]
        )
        #expect(store.createArtifactGraphIDs.last == ["g2"])
        #expect(chosen.id != nilLinks.id)
        #expect(store.evidenceGraphsByProject[projectDir]?.map(\.artifactIDs) == [["a0"], ["a0"]])
        let (empty, _) = try await store.createArtifact(
            projectDir: projectDir, userID: userID, sourceID: sourceID,
            fileID: "", label: "Page 3", description: "", graphIDs: []
        )
        #expect(store.createArtifactGraphIDs.last == [])
        #expect(store.evidenceGraphsByProject[projectDir]?.map(\.artifactIDs) == [["a0"], ["a0"]])
        #expect(store.evidenceGraphsByProject[otherDir]?.first?.artifactIDs == ["b0"])
        _ = empty
    }

    /// T-20 (D-14): a new Source has a Default graph; a copy's Default graph
    /// seeds the copied Artifact ids on that graph via `insertSource`.
    @Test func createSourceMakesADefaultGraph() async throws {
        let store = FakeStore()
        let (created, _) = try await store.createSource(
            projectDir: projectDir, userID: userID, sourceTypeID: "t", title: "Deed", description: ""
        )
        let graphs = try await store.listEvidenceGraphs(projectDir: projectDir, sourceID: nil)
        #expect(graphs.map(\.name) == ["Default"])
        #expect(graphs.first?.sourceID == created.id)
        #expect(graphs.first?.id != created.id)
        #expect(graphs.first?.artifactIDs == [])

        let (artifact, _) = try await store.createArtifact(
            projectDir: projectDir, userID: userID, sourceID: created.id,
            fileID: "", label: "Page 1", description: "", graphIDs: nil
        )
        #expect(store.createArtifactGraphIDs == [nil])
        let linked = try await store.listEvidenceGraphs(projectDir: projectDir, sourceID: nil)
        #expect(linked.first?.artifactIDs == [])

        let (copy, _) = try await store.copySource(
            projectDir: projectDir, userID: userID, sourceID: created.id, title: "Deed copy", sourceTypeID: "t",
            copyDescription: false, copyMetadata: false, copyCredibility: false, artifactIDs: [artifact.id]
        )
        let all = try await store.listEvidenceGraphs(projectDir: projectDir, sourceID: nil)
        let copied = try #require(all.first { $0.sourceID == copy.id })
        #expect(copied.name == "Default")
        #expect(copied.artifactIDs.count == 1)
        #expect(copied.artifactIDs != [artifact.id])
    }

    @Test func createPropertyAndTermKeysUseFieldSlug() async throws {
        let store = seededStore()
        store.propertiesByProject[projectDir] = [
            CatalogProperty(
                id: "p1", key: "custom-fact", origin: "user", label: "Custom Fact",
                description: "", valueType: "text", cardinality: "single"
            ),
        ]
        let property = try await store.createProperty(
            projectDir: projectDir, userID: userID, label: "Custom Fact",
            valueType: "text", description: "", cardinality: ""
        )
        #expect(property.key == "custom-fact")
        let term = try await store.createPropertyTerm(
            projectDir: projectDir, userID: userID, propertyID: "p1",
            label: "Grandma's album code", description: ""
        )
        #expect(term.key == "grandmas-album-code")
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
        let afterArtifact = try await store.listEvidenceGraphs(projectDir: projectDir, sourceID: nil)
        #expect(afterArtifact.map(\.artifactIDs) == [["a2"], []])

        try await store.deleteSource(projectDir: projectDir, userID: userID, sourceID: sourceID)
        #expect(try await store.listEvidenceGraphs(projectDir: projectDir, sourceID: nil).isEmpty)
    }

    /// T-13 (D-14): the fake answers with each scripted block whose pair was asked,
    /// in pair order, and records the pairs.
    @Test func unlinkImpactReturnsScriptedBlocksPerPair() async throws {
        let store = seededStore()
        store.unlinkImpactBlocks = [
            unlinkBlock("a1", graphID: "g1", citations: 1),
            unlinkBlock("a1", graphID: "g2", citations: 1),
        ]
        let pairs = [
            CatalogEvidenceGraphArtifactPair(graphID: "g2", artifactID: "a1"),
            CatalogEvidenceGraphArtifactPair(graphID: "g1", artifactID: "a1"),
            CatalogEvidenceGraphArtifactPair(graphID: "g2", artifactID: "a9"),
        ]
        let blocks = try await store.getEvidenceGraphUnlinkImpact(projectDir: projectDir, pairs: pairs)
        #expect(blocks.map(\.graphID) == ["g2", "g1"])
        #expect(store.getEvidenceGraphUnlinkImpactPairs == pairs)
    }
}
