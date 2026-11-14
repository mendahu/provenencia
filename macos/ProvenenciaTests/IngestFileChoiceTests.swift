import Foundation
import Testing
@testable import Provenencia

@Suite
struct IngestFileChoiceTests {
    private struct OtherError: Error {}

    @Test func chooseAcceptsAnImage() throws {
        let url = try ArtifactFileFixtures.writeStubPNG()
        var file = IngestFileChoice()
        file.choose(url)
        #expect(file.filePath == url.path)
        #expect(file.fileName == "scan.png")
        #expect(file.reject == nil)
    }

    @Test func chooseRefusesASpreadsheetWithoutTheEngine() throws {
        let url = try ArtifactFileFixtures.writeTemp(Data("PK\u{0003}\u{0004}".utf8), named: "notes.xlsx")
        var file = IngestFileChoice()
        file.choose(url)
        #expect(file.filePath == nil)
        #expect(file.fileName == "notes.xlsx")
        #expect(file.reject != nil)
    }

    @Test func twoURLsAreRefused() throws {
        let first = try ArtifactFileFixtures.writeStubPNG()
        let second = try ArtifactFileFixtures.writeStubPNG()
        var file = IngestFileChoice()
        file.choose([first, second])
        #expect(file.filePath == nil)
        #expect(file.fileName == first.lastPathComponent)
        #expect(file.reject == L10n.Errors.ingestCallout(reason: .multiFile))
    }

    @Test func anIngestRefusalKeepsTheNameAndDropsThePath() throws {
        var file = IngestFileChoice()
        file.choose(try ArtifactFileFixtures.writeStubPNG())
        let error = CoreInvokeError.coded(status: 1, code: "ingest.invalid", kind: .user, params: [])
        let refused = file.refuse(ifIngestRefusal: error)
        #expect(refused)
        #expect(file.filePath == nil)
        #expect(file.fileName == "scan.png")
        #expect(file.reject != nil)
    }

    @Test func aFailureAfterTheCreateKeepsThePath() throws {
        let url = try ArtifactFileFixtures.writeStubPNG()
        var file = IngestFileChoice()
        file.choose(url)
        file.failAfterCreate(message: "Already attached.")
        #expect(file.filePath == url.path)
        #expect(file.fileName == "scan.png")
        #expect(file.reject?.title == L10n.Sources.artifactFileNotAttachedTitle)
        #expect(file.reject?.message.contains("Already attached.") == true)
        #expect(!file.isRefused)
    }

    @Test func otherErrorsAreNotRefusals() throws {
        var file = IngestFileChoice()
        file.choose(try ArtifactFileFixtures.writeStubPNG())
        let before = file
        let refused = file.refuse(ifIngestRefusal: OtherError())
        #expect(!refused)
        #expect(file == before)
    }

    @Test func clearEmptiesTheChoice() throws {
        var file = IngestFileChoice()
        file.choose(try ArtifactFileFixtures.writeStubPNG())
        file.clear()
        #expect(file == IngestFileChoice())
    }
}
