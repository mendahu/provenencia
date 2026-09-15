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

/// Load/persist problems for navigation history. Workspace chrome surfaces these
/// as a non-blocking toast; navigation itself keeps working from an in-memory stack.
enum NavigationHistoryIssue: Equatable, Sendable {
    /// File existed but was unreadable, undecodable, or failed validation.
    case loadFailed
    /// Could not create directories or write the history JSON.
    case persistFailed
}

@MainActor
final class NavigationHistoryStore {
    private(set) var document: NavigationHistoryDocument
    private(set) var lastIssue: NavigationHistoryIssue?
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
        switch Self.loadOutcome(from: fileURL, fileManager: fileManager) {
        case .missing:
            self.document = Self.seedDocument(projectUuid: projectUuid, seed: seed)
            persist()
        case .decoded(let loaded) where Self.isUsable(loaded, projectUuid: projectUuid):
            var doc = loaded
            doc.projectUuid = projectUuid
            doc.v = NavigationHistoryDocument.formatVersion
            self.document = doc
        case .decoded, .unreadable:
            lastIssue = .loadFailed
            self.document = Self.seedDocument(projectUuid: projectUuid, seed: seed)
            persist()
        }
    }

    func clearLastIssue() {
        lastIssue = nil
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
            lastIssue = .persistFailed
        }
    }

    private static func seedDocument(
        projectUuid: String,
        seed: WorkspaceLocation
    ) -> NavigationHistoryDocument {
        NavigationHistoryDocument(
            v: NavigationHistoryDocument.formatVersion,
            projectUuid: projectUuid,
            index: 0,
            entries: [seed]
        )
    }

    private static func isUsable(_ doc: NavigationHistoryDocument, projectUuid: String) -> Bool {
        (doc.projectUuid == projectUuid || doc.projectUuid.isEmpty)
            && !doc.entries.isEmpty
            && doc.entries.indices.contains(doc.index)
    }

    private enum LoadOutcome {
        case missing
        case decoded(NavigationHistoryDocument)
        case unreadable
    }

    private static func loadOutcome(from url: URL, fileManager: FileManager) -> LoadOutcome {
        guard fileManager.fileExists(atPath: url.path) else { return .missing }
        guard let data = try? Data(contentsOf: url) else { return .unreadable }
        guard let doc = try? JSONDecoder().decode(NavigationHistoryDocument.self, from: data) else {
            return .unreadable
        }
        return .decoded(doc)
    }
}
