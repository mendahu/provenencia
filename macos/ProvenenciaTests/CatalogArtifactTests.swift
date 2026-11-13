import Foundation
import Testing
@testable import Provenencia

@Suite
struct CatalogArtifactTests {
    @Test func displayTitleAndIcon() {
        let labelled = catalogArtifact("a1", ref: "ART-0001", label: " Scan ")
        #expect(labelled.displayTitle == "Scan")
        #expect(labelled.hasDisplayLabel)

        let blank = catalogArtifact("a1", ref: "ART-0001", label: "  ")
        #expect(blank.displayTitle == "ART-0001")
        #expect(!blank.hasDisplayLabel)

        #expect(blank.thumbnailTypeIconKey(sourceTypeIconKey: "k") == "k")

        let file = CatalogFileRef(
            id: "f1", relPath: "objects/a", originalFilename: "a.png", mediaType: "image/png", byteSize: 1
        )
        let withFile = catalogArtifact("a1", ref: "ART-0001", label: "", fileID: "f1", file: file)
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
        #expect(
            catalogArtifact("a1", ref: "ART-0001", label: "Scan", fileID: "f1", file: file).fileDisplayName == "scan.png"
        )
        #expect(catalogArtifact("a1", ref: "ART-0001", label: "Scan").fileDisplayName == "ART-0001")
    }
}
