import Foundation

/// Browser-shaped navigation stack for one project window, persisted as JSON
/// under Application Support. See `docs/deployment-plan/spike-3/navigation-history.md`.
struct NavigationHistoryDocument: Codable, Equatable, Sendable {
    var v: Int
    var projectUuid: String
    var index: Int
    var entries: [WorkspaceLocation]

    static let formatVersion = 1
    static let maxEntries = 100
}

@MainActor
final class NavigationHistoryStore {
    private(set) var document: NavigationHistoryDocument
    private let fileURL: URL
    private let fileManager: FileManager

    var current: WorkspaceLocation {
        guard document.entries.indices.contains(document.index) else {
            return .sectionRoot(.sources)
        }
        return document.entries[document.index]
    }

    var canGoBack: Bool { document.index > 0 }
    var canGoForward: Bool { document.index + 1 < document.entries.count }

    /// Nearest entries before the current index (nearest first), capped at `limit`.
    func backJumpItems(limit: Int = 15) -> [(index: Int, location: WorkspaceLocation)] {
        guard document.index > 0, limit > 0 else { return [] }
        let start = max(0, document.index - limit)
        let range = start ..< document.index
        return range.reversed().map { (index: $0, location: document.entries[$0]) }
    }

    /// Entries after the current index in stack order, capped at `limit`.
    func forwardJumpItems(limit: Int = 15) -> [(index: Int, location: WorkspaceLocation)] {
        guard canGoForward, limit > 0 else { return [] }
        let end = min(document.entries.count, document.index + 1 + limit)
        let range = (document.index + 1) ..< end
        return range.map { (index: $0, location: document.entries[$0]) }
    }

    init(
        projectUuid: String,
        fileURL: URL,
        fileManager: FileManager = .default,
        seed: WorkspaceLocation = .sectionRoot(.sources)
    ) {
        self.fileURL = fileURL
        self.fileManager = fileManager
        if let loaded = Self.load(from: fileURL, fileManager: fileManager),
           loaded.projectUuid == projectUuid || loaded.projectUuid.isEmpty,
           !loaded.entries.isEmpty,
           loaded.entries.indices.contains(loaded.index)
        {
            var doc = loaded
            doc.projectUuid = projectUuid
            doc.v = NavigationHistoryDocument.formatVersion
            self.document = doc
        } else {
            self.document = NavigationHistoryDocument(
                v: NavigationHistoryDocument.formatVersion,
                projectUuid: projectUuid,
                index: 0,
                entries: [seed]
            )
            persist()
        }
    }

    /// Push a committed navigation (truncates forward). Coalesces when identical to current.
    @discardableResult
    func go(to location: WorkspaceLocation) -> WorkspaceLocation {
        if location == current {
            return current
        }
        if document.index < document.entries.count - 1 {
            document.entries = Array(document.entries.prefix(document.index + 1))
        }
        document.entries.append(location)
        while document.entries.count > NavigationHistoryDocument.maxEntries {
            document.entries.removeFirst()
        }
        document.index = document.entries.count - 1
        persist()
        return current
    }

    @discardableResult
    func goBack() -> WorkspaceLocation? {
        guard canGoBack else { return nil }
        document.index -= 1
        persist()
        return current
    }

    @discardableResult
    func goForward() -> WorkspaceLocation? {
        guard canGoForward else { return nil }
        document.index += 1
        persist()
        return current
    }

    /// Jump without truncating (history jump menu).
    @discardableResult
    func go(toIndex index: Int) -> WorkspaceLocation? {
        guard document.entries.indices.contains(index) else { return nil }
        document.index = index
        persist()
        return current
    }

    /// Replace the current entry in place (e.g. prune missing entity to section root).
    @discardableResult
    func replaceCurrent(with location: WorkspaceLocation) -> WorkspaceLocation {
        guard document.entries.indices.contains(document.index) else {
            return go(to: location)
        }
        if document.entries[document.index] == location {
            return current
        }
        document.entries[document.index] = location
        persist()
        return current
    }

    func persist() {
        do {
            let dir = fileURL.deletingLastPathComponent()
            try fileManager.createDirectory(at: dir, withIntermediateDirectories: true)
            let data = try JSONEncoder().encode(document)
            let temp = fileURL.appendingPathExtension("tmp")
            try data.write(to: temp, options: .atomic)
            if fileManager.fileExists(atPath: fileURL.path) {
                _ = try fileManager.replaceItemAt(fileURL, withItemAt: temp)
            } else {
                try fileManager.moveItem(at: temp, to: fileURL)
            }
        } catch {
            #if DEBUG
            assertionFailure("navigation history persist failed: \(error)")
            #endif
        }
    }

    private static func load(from url: URL, fileManager: FileManager) -> NavigationHistoryDocument? {
        guard fileManager.fileExists(atPath: url.path) else { return nil }
        guard let data = try? Data(contentsOf: url) else { return nil }
        return try? JSONDecoder().decode(NavigationHistoryDocument.self, from: data)
    }
}
