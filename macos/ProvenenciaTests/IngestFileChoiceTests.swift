import Foundation
import Testing
@testable import Provenencia

@Suite
struct IngestFileChoiceTests {
    private struct OtherError: Error {}

    private func writeTempFile(named name: String, bytes: [UInt8]) throws -> URL {
        let dir = FileManager.default.temporaryDirectory
            .appendingPathComponent("ingest-choice-\(UUID().uuidString)", isDirectory: true)
        try FileManager.default.createDirectory(at: dir, withIntermediateDirectories: true)
        let url = dir.appendingPathComponent(name)
        try Data(bytes).write(to: url)
        return url
    }

    private func writeTempPNG() throws -> URL {
        try writeTempFile(named: "scan.png", bytes: [0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A])
    }

    @Test func chooseAcceptsAnImage() throws {
        let url = try writeTempPNG()
        var file = IngestFileChoice()
        file.choose(url)
        #expect(file.filePath == url.path)
        #expect(file.fileName == "scan.png")
        #expect(file.reject == nil)
    }

    @Test func chooseRefusesASpreadsheetWithoutTheEngine() throws {
        let url = try writeTempFile(named: "notes.xlsx", bytes: Array("PK\u{0003}\u{0004}".utf8))
        var file = IngestFileChoice()
        file.choose(url)
        #expect(file.filePath == nil)
        #expect(file.fileName == "notes.xlsx")
        #expect(file.reject != nil)
    }

    @Test func twoURLsAreRefused() throws {
        let first = try writeTempPNG()
        let second = try writeTempPNG()
        var file = IngestFileChoice()
        file.choose([first, second])
        #expect(file.filePath == nil)
        #expect(file.fileName == first.lastPathComponent)
        #expect(file.reject == L10n.Errors.ingestCallout(reason: .multiFile))
    }

    @Test func anIngestRefusalKeepsTheNameAndDropsThePath() throws {
        var file = IngestFileChoice()
        file.choose(try writeTempPNG())
        let error = CoreInvokeError.coded(status: 1, code: "ingest.invalid", kind: .user, params: [])
        let refused = file.refuse(ifIngestRefusal: error)
        #expect(refused)
        #expect(file.filePath == nil)
        #expect(file.fileName == "scan.png")
        #expect(file.reject != nil)
    }

    @Test func otherErrorsAreNotRefusals() throws {
        var file = IngestFileChoice()
        file.choose(try writeTempPNG())
        let before = file
        let refused = file.refuse(ifIngestRefusal: OtherError())
        #expect(!refused)
        #expect(file == before)
    }

    @Test func clearEmptiesTheChoice() throws {
        var file = IngestFileChoice()
        file.choose(try writeTempPNG())
        file.clear()
        #expect(file == IngestFileChoice())
    }
}
