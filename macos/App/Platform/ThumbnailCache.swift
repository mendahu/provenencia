import AppKit
import Foundation

/// Process-lifetime decoded thumbnail cache. Keys by project dir + relative
/// object path; loads off the main actor and coalesces in-flight requests.
@MainActor
final class ThumbnailCache {
    static let shared = ThumbnailCache()

    /// Bound entry count; oldest keys are dropped when exceeded.
    var maxEntries: Int = 100

    private struct Key: Hashable {
        let projectDir: String
        let relPath: String
    }

    private enum Entry {
        case image(NSImage)
        case missing
    }

    private var store: [Key: Entry] = [:]
    private var order: [Key] = []
    private var inflight: [Key: Task<NSImage?, Never>] = [:]

    init() {}

    /// Empty / whitespace `relPath` → nil without touching disk.
    func image(projectDir: String, relPath: String) async -> NSImage? {
        let trimmed = relPath.trimmingCharacters(in: .whitespacesAndNewlines)
        guard !trimmed.isEmpty else { return nil }

        let key = Key(projectDir: projectDir, relPath: trimmed)
        if let entry = store[key] {
            touch(key)
            switch entry {
            case .image(let img): return img
            case .missing: return nil
            }
        }

        if let existing = inflight[key] {
            return await existing.value
        }

        let task = Task<NSImage?, Never> {
            guard let url = ProjectFiles.objectURL(projectDir: projectDir, relPath: trimmed) else {
                return nil
            }
            return await Task.detached(priority: .userInitiated) {
                NSImage(contentsOf: url)
            }.value
        }
        inflight[key] = task
        let image = await task.value
        inflight[key] = nil

        if let image {
            insert(key, .image(image))
        } else {
            insert(key, .missing)
        }
        return image
    }

    func clear() {
        store.removeAll()
        order.removeAll()
        for (_, task) in inflight {
            task.cancel()
        }
        inflight.removeAll()
    }

    /// Test helper: seeded entry count after loads.
    var entryCount: Int { store.count }

    private func insert(_ key: Key, _ entry: Entry) {
        if store[key] == nil {
            order.append(key)
        } else {
            touch(key)
        }
        store[key] = entry
        while store.count > maxEntries, let victim = order.first {
            order.removeFirst()
            store[victim] = nil
        }
    }

    private func touch(_ key: Key) {
        if let idx = order.firstIndex(of: key) {
            order.remove(at: idx)
            order.append(key)
        }
    }
}
