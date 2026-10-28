import Foundation
import Testing
@testable import Provenencia

@Suite
struct CatalogArtifactTests {
    private func artifact(label: String, file: CatalogFileRef? = nil, fileID: String = "") -> CatalogArtifact {
        CatalogArtifact(
            id: "a1", ref: "ART-0001", sourceID: "s1", fileID: fileID,
            label: label, description: "", file: file
        )
    }

    @Test func displayTitleAndIcon() {
        let labelled = artifact(label: " Scan ")
        #expect(labelled.displayTitle == "Scan")
        #expect(labelled.hasDisplayLabel)

        let blank = artifact(label: "  ")
        #expect(blank.displayTitle == "ART-0001")
        #expect(!blank.hasDisplayLabel)

        #expect(blank.thumbnailTypeIconKey(sourceTypeIconKey: "k") == "k")

        let file = CatalogFileRef(
            id: "f1", relPath: "objects/a", originalFilename: "a.png", mediaType: "image/png", byteSize: 1
        )
        let withFile = artifact(label: "", file: file, fileID: "f1")
        #expect(withFile.hasFile)
        #expect(withFile.thumbnailTypeIconKey(sourceTypeIconKey: "k") == nil)
    }

    @Test(arguments: [
        ("scan.png", "objects/aa/bb/aabb.png", "scan.png"),
        ("  scan.png ", "objects/aa/bb/aabb.png", "scan.png"),
        ("  ", "objects/aa/bb/aabb.png", "aabb.png"),
        ("", "objects/aa/bb/aabb.png", "aabb.png"),
        ("  ", "", "ART-0001"),
        ("", "", "ART-0001"),
    ])
    func fileDisplayNameNamesTheFile(originalFilename: String, relPath: String, expected: String) {
        let name = CatalogArtifact.fileDisplayName(
            originalFilename: originalFilename, relPath: relPath, ref: "ART-0001"
        )
        #expect(name == expected)
    }

    @Test func fileDisplayNameReadsTheArtifactsOwnFile() {
        let file = CatalogFileRef(
            id: "f1", relPath: "objects/aa/bb/aabb.png", originalFilename: "scan.png", mediaType: "image/png", byteSize: 1
        )
        #expect(artifact(label: "Scan", file: file, fileID: "f1").fileDisplayName == "scan.png")
        #expect(artifact(label: "Scan").fileDisplayName == "ART-0001")
    }
}
