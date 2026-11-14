import Foundation
import Testing
import UniformTypeIdentifiers
@testable import Provenencia

struct IngestMediaPolicyTests {
    @Test func acceptsCSVAndMarkdownExtensions() throws {
        let csv = try ArtifactFileFixtures.writeTemp(Data("a,b\n1,2\n".utf8), named: "people.csv")
        #expect(IngestMediaPolicy.validate(url: csv) == .ok)
        let md = try ArtifactFileFixtures.writeTemp(Data("# Hi\n".utf8), named: "notes.md")
        #expect(IngestMediaPolicy.validate(url: md) == .ok)
    }

    @Test func rejectsOfficeExtension() throws {
        let url = try ArtifactFileFixtures.writeTemp(Data("PK\u{0003}\u{0004}xxxx".utf8), named: "notes.xlsx")
        #expect(IngestMediaPolicy.validate(url: url) == .reject(.office))
    }

    @Test func acceptsWordExtensions() throws {
        let docx = try ArtifactFileFixtures.writeTemp(Data("PK\u{0003}\u{0004}xxxx".utf8), named: "letter.docx")
        #expect(IngestMediaPolicy.validate(url: docx) == .ok)
        let doc = try ArtifactFileFixtures.writeTemp(Data(" helloid".utf8), named: "letter.doc")
        #expect(IngestMediaPolicy.validate(url: doc) == .ok)
    }

    @Test func rejectsEmpty() throws {
        let url = try ArtifactFileFixtures.writeTemp(Data(), named: "empty.txt")
        #expect(IngestMediaPolicy.validate(url: url) == .reject(.empty))
    }

    @Test func rejectsUnknownExtension() throws {
        let url = try ArtifactFileFixtures.writeTemp(Data("hello".utf8), named: "blob.xyz")
        #expect(IngestMediaPolicy.validate(url: url) == .reject(.genericType))
    }

    @Test func rejectsALinkFileAsAGenericType() throws {
        // A stored link or Terminal file is not a program, so ingest keeps the generic message.
        for name in ["x.webloc", "x.terminal"] {
            let url = try ArtifactFileFixtures.writeTemp(Data("hello".utf8), named: name)
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
}
