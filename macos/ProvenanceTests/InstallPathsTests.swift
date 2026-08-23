import Foundation
import Testing
@testable import Provenance

struct InstallPathsTests {
    @Test func provenanceProjectsIgnoresMissingDirFilesAndNonProvenanceFolders() throws {
        let root = FileManager.default.temporaryDirectory.appendingPathComponent("install-paths-\(UUID().uuidString)", isDirectory: true)
        let docs = root.appendingPathComponent("Documents", isDirectory: true)
        try FileManager.default.createDirectory(at: docs, withIntermediateDirectories: true)

        let keep = docs.appendingPathComponent("Smith Family.provenance", isDirectory: true)
        try FileManager.default.createDirectory(at: keep, withIntermediateDirectories: true)

        let notADir = docs.appendingPathComponent("notes.provenance")
        try Data("file".utf8).write(to: notADir)

        try FileManager.default.createDirectory(
            at: docs.appendingPathComponent("NotAProject", isDirectory: true),
            withIntermediateDirectories: true
        )

        let found = try InstallPaths.provenanceProjects(in: docs)
        #expect(found.map(\.lastPathComponent).sorted() == ["Smith Family.provenance"])

        let missing = root.appendingPathComponent("no-such-docs", isDirectory: true)
        #expect(try InstallPaths.provenanceProjects(in: missing).isEmpty)
    }

    @Test func isProjectDirectoryTrueForDirFalseForMissing() throws {
        let dir = FileManager.default.temporaryDirectory.appendingPathComponent("proj-dir-\(UUID().uuidString)", isDirectory: true)
        try FileManager.default.createDirectory(at: dir, withIntermediateDirectories: true)
        #expect(InstallPaths.isProjectDirectory(dir.path))
        #expect(!InstallPaths.isProjectDirectory(dir.appendingPathComponent("gone").path))
    }
}
