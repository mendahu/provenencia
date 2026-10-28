import Foundation
import Testing
import UniformTypeIdentifiers
@testable import Provenencia

struct IngestMediaPolicyTests {
    @Test func acceptsCSVAndMarkdownExtensions() throws {
        let csv = try writeTemp(name: "people.csv", data: Data("a,b\n1,2\n".utf8))
        #expect(IngestMediaPolicy.validate(url: csv) == .ok)
        let md = try writeTemp(name: "notes.md", data: Data("# Hi\n".utf8))
        #expect(IngestMediaPolicy.validate(url: md) == .ok)
    }

    @Test func rejectsOfficeExtension() throws {
        let url = try writeTemp(name: "notes.xlsx", data: Data("PK\u{0003}\u{0004}xxxx".utf8))
        #expect(IngestMediaPolicy.validate(url: url) == .reject(.office))
    }

    @Test func acceptsWordExtensions() throws {
        let docx = try writeTemp(name: "letter.docx", data: Data("PK\u{0003}\u{0004}xxxx".utf8))
        #expect(IngestMediaPolicy.validate(url: docx) == .ok)
        let doc = try writeTemp(name: "letter.doc", data: Data(" helloid".utf8))
        #expect(IngestMediaPolicy.validate(url: doc) == .ok)
    }

    @Test func rejectsEmpty() throws {
        let url = try writeTemp(name: "empty.txt", data: Data())
        #expect(IngestMediaPolicy.validate(url: url) == .reject(.empty))
    }

    @Test func rejectsUnknownExtension() throws {
        let url = try writeTemp(name: "blob.xyz", data: Data("hello".utf8))
        #expect(IngestMediaPolicy.validate(url: url) == .reject(.genericType))
    }

    @Test func rejectsALinkFileAsAGenericType() throws {
        // A stored link or Terminal file is not a program, so ingest keeps the generic message.
        for name in ["x.webloc", "x.terminal"] {
            let url = try writeTemp(name: name, data: Data("hello".utf8))
            #expect(IngestMediaPolicy.validate(url: url) == .reject(.genericType), "\(name)")
        }
    }

    @Test func calloutOfficeHasTitleAndHelp() {
        let callout = L10n.Errors.ingestCallout(reason: .office)
        #expect(L10n.string(callout.title).contains("Spreadsheets"))
        #expect(callout.message.contains("Word"))
    }

    @Test func calloutTooLargeInterpolatesSize() {
        let callout = L10n.Errors.ingestCallout(reason: .tooLarge, sizeLabel: "1.24 GB")
        #expect(callout.message.contains("1.24 GB"))
    }

    @Test func ffiCodeMapsToCallout() {
        let callout = L10n.Errors.ingestCallout(code: "ingest.unsupported_archive")
        #expect(callout != nil)
        #expect(L10n.string(callout!.title).contains("Archives"))
    }

    @Test func unsafeObjectsCodeMapsToCallout() {
        let callout = L10n.Errors.ingestCallout(code: "ingest.unsafe_objects")
        #expect(callout != nil)
        #expect(L10n.string(callout!.title).contains("objects folder"))
        #expect(callout!.message.contains("link to another location"))
    }

    @Test func isExecutableTypeCoversScriptsAndApps() throws {
        #expect(IngestMediaPolicy.isExecutableType(.shellScript))
        #expect(IngestMediaPolicy.isExecutableType(.unixExecutable))
        #expect(IngestMediaPolicy.isExecutableType(.applicationBundle))
        #expect(!IngestMediaPolicy.isExecutableType(.pdf))
        // A stored link or Terminal file is not a program: only Open refuses it.
        for ext in ["webloc", "fileloc", "inetloc", "url", "terminal", "term"] {
            let type = try #require(UTType(filenameExtension: ext))
            #expect(!IngestMediaPolicy.isExecutableType(type), "\(ext)")
        }
    }

    @Test func opensAnotherItemCoversLinksAndTerminalFiles() throws {
        for ext in ["webloc", "fileloc", "inetloc", "url", "terminal", "term"] {
            let type = try #require(UTType(filenameExtension: ext))
            #expect(IngestMediaPolicy.opensAnotherItem(type), "\(ext)")
        }
        #expect(!IngestMediaPolicy.opensAnotherItem(.pdf))
        #expect(!IngestMediaPolicy.opensAnotherItem(.shellScript))
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
