import Foundation
import Testing
@testable import Provenencia

@Suite
struct SourcesListSettingsTests {
    private let uuid = "00000000-0000-7000-8000-0000000000aa"

    private func tempFile() throws -> URL {
        let dir = FileManager.default.temporaryDirectory
            .appendingPathComponent("sources-list-settings-\(UUID().uuidString)", isDirectory: true)
        try FileManager.default.createDirectory(at: dir, withIntermediateDirectories: true)
        return dir.appendingPathComponent("settings.json", isDirectory: false)
    }

    private func store(
        _ projectUuid: String, _ userID: String, _ url: URL
    ) -> SourcesListSettingsStore? {
        SourcesListSettingsStore(projectUuid: projectUuid, userID: userID, fileURL: url)
    }

    private func settings(
        sort: SourcesModel.Sort, query: String
    ) -> SourcesListSettings {
        var value = SourcesListSettings()
        value.sort = sort
        value.typeFilterID = "t1"
        value.status = .inProgress
        value.query = query
        return value
    }

    @Test func roundTrip() throws {
        let url = try tempFile()
        let saved = settings(sort: .az, query: "smith")
        try #require(store(uuid, "u1", url)).save(saved)
        #expect(try #require(store(uuid, "u1", url)).load() == saved)
    }

    @Test func perProjectAndUser() throws {
        let url = try tempFile()
        let other = try tempFile()
        let a = settings(sort: .az, query: "a")
        let b = settings(sort: .za, query: "b")
        try #require(store(uuid, "u1", url)).save(a)
        try #require(store(uuid, "u2", url)).save(b)
        #expect(try #require(store(uuid, "u1", url)).load() == a)
        #expect(try #require(store(uuid, "u2", url)).load() == b)
        // A second project keys a different file.
        #expect(try #require(store("00000000-0000-7000-8000-0000000000bb", "u1", other)).load() == SourcesListSettings())
        let first = try InstallPaths.sourcesListSettingsFile(projectUuid: uuid)
        let second = try InstallPaths.sourcesListSettingsFile(projectUuid: "00000000-0000-7000-8000-0000000000bb")
        #expect(first != second)
    }

    @Test func badDataReadsDefaults() throws {
        let url = try tempFile()
        try Data("not json".utf8).write(to: url)
        #expect(try #require(store(uuid, "u1", url)).load() == SourcesListSettings())
    }

    @Test func unknownFieldValuesFallBack() throws {
        let url = try tempFile()
        let json = #"{"u1":{"sort":"bogus","status":"x","query":"smith"}}"#
        try Data(json.utf8).write(to: url)
        let loaded = try #require(store(uuid, "u1", url)).load()
        #expect(loaded.sort == .added)
        #expect(loaded.status == nil)
        #expect(loaded.query == "smith")
    }

    @Test func emptyKeysMakeNoStore() throws {
        let url = try tempFile()
        #expect(store("", "u1", url) == nil)
        #expect(store("  ", "u1", url) == nil)
        #expect(store("{}", "u1", url) == nil)
        #expect(store(uuid, "", url) == nil)
        #expect(store(uuid, "  ", url) == nil)
    }

    @Test func projectKeyMatchesNavigationFileName() {
        #expect(
            InstallPaths.navigationFileName(projectUuid: "{00000000-0000-7000-8000-0000000000AA}")
                == InstallPaths.projectKey(projectUuid: "{00000000-0000-7000-8000-0000000000AA}") + ".json"
        )
        #expect(InstallPaths.projectKey(projectUuid: uuid) == "000000000000700080000000000000aa")
    }
}
