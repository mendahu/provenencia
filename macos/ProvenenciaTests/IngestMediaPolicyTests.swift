import Foundation
import Testing
@testable import Provenencia

struct IngestMediaPolicyTests {
    @Test func acceptsCSVAndMarkdownExtensions() throws {
        let csv = try writeTemp(name: "people.csv", data: Data("a,b\n1,2\n".utf8))
        #expect(IngestMediaPolicy.validate(url: csv) == .ok)
        let md = try writeTemp(name: "notes.md", data: Data("# Hi\n".utf8))
        #expect(IngestMediaPolicy.validate(url: md) == .ok)
    }

    @Test func rejectsOfficeExtension() throws {
        let url = try writeTemp(name: "notes.docx", data: Data("PK\u{0003}\u{0004}xxxx".utf8))
        #expect(IngestMediaPolicy.validate(url: url) == .reject(.office))
    }

    @Test func rejectsEmpty() throws {
        let url = try writeTemp(name: "empty.txt", data: Data())
        #expect(IngestMediaPolicy.validate(url: url) == .reject(.empty))
    }

    @Test func rejectsUnknownExtension() throws {
        let url = try writeTemp(name: "blob.xyz", data: Data("hello".utf8))
        #expect(IngestMediaPolicy.validate(url: url) == .reject(.genericType))
    }

    @Test func calloutOfficeHasTitleAndHelp() {
        let callout = L10n.Errors.ingestCallout(reason: .office)
        #expect(String(localized: callout.title).contains("Office"))
        #expect(callout.message.contains("512 MB"))
    }

    @Test func calloutTooLargeInterpolatesSize() {
        let callout = L10n.Errors.ingestCallout(reason: .tooLarge, sizeLabel: "1.24 GB")
        #expect(callout.message.contains("1.24 GB"))
    }

    @Test func ffiCodeMapsToCallout() {
        let callout = L10n.Errors.ingestCallout(code: "ingest.unsupported_archive")
        #expect(callout != nil)
        #expect(String(localized: callout!.title).contains("Archives"))
    }

    private func writeTemp(name: String, data: Data) throws -> URL {
        let dir = FileManager.default.temporaryDirectory
            .appendingPathComponent("ingest-policy-\(UUID().uuidString)", isDirectory: true)
        try FileManager.default.createDirectory(at: dir, withIntermediateDirectories: true)
        let url = dir.appendingPathComponent(name)
        try data.write(to: url)
        return url
    }
}
