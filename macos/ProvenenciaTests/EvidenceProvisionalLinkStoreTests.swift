import Foundation
import Testing
@testable import Provenencia

@Suite
@MainActor
struct EvidenceProvisionalLinkStoreTests {
    @Test func memoryStoreRoundTripsUpsert() {
        let store = InMemoryEvidenceProvisionalLinkStore()
        let link = EvidenceProvisionalLink(
            bridgeSubjectID: "b1",
            endpointAID: "a",
            endpointBID: "c"
        )
        store.upsert(link, sourceID: "src-1")
        #expect(store.links(for: "src-1") == [link])
        store.upsert(
            EvidenceProvisionalLink(bridgeSubjectID: "b1", endpointAID: "a2", endpointBID: "c2"),
            sourceID: "src-1"
        )
        #expect(store.links(for: "src-1").first?.endpointAID == "a2")
        store.remove(bridgeSubjectID: "b1", sourceID: "src-1")
        #expect(store.links(for: "src-1").isEmpty)
    }

    @Test func fileStorePersistsAcrossReload() throws {
        let dir = FileManager.default.temporaryDirectory
            .appendingPathComponent("evidence-links-\(UUID().uuidString)", isDirectory: true)
        try FileManager.default.createDirectory(at: dir, withIntermediateDirectories: true)
        let file = dir.appendingPathComponent("links.json")
        let projectKey = "/tmp/demo.provenencia"

        let first = EvidenceProvisionalLinkStore(
            projectKey: projectKey,
            fileURL: file
        )
        first.upsert(
            EvidenceProvisionalLink(bridgeSubjectID: "bridge", endpointAID: "p1", endpointBID: "e1"),
            sourceID: "src"
        )

        let second = EvidenceProvisionalLinkStore(
            projectKey: projectKey,
            fileURL: file
        )
        #expect(second.links(for: "src").count == 1)
        #expect(second.links(for: "src").first?.endpointAID == "p1")
    }
}
