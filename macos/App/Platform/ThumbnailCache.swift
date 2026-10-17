import AppKit
import Foundation
import ImageIO

/// Process-lifetime decoded thumbnail cache. Keys by project dir + relative
/// object path; loads off the main actor and coalesces in-flight requests.
/// Images decode eagerly at no more than `maxPixelSize` on the longest edge,
/// so a full-resolution crop never sits decoded in the cache.
@MainActor
final class ThumbnailCache {
    static let shared = ThumbnailCache()

    /// Longest-edge pixel cap for a decoded thumbnail (the 120pt header thumb
    /// at 2x, with margin).
    static let maxPixelSize = 512

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
                guard let src = CGImageSourceCreateWithURL(url as CFURL, nil),
                      let cg = CGImageSourceCreateThumbnailAtIndex(src, 0, [
                          kCGImageSourceCreateThumbnailFromImageAlways: true,
                          kCGImageSourceCreateThumbnailWithTransform: true,
                          kCGImageSourceShouldCacheImmediately: true,
                          kCGImageSourceThumbnailMaxPixelSize: ThumbnailCache.maxPixelSize,
                      ] as CFDictionary)
                else { return nil }
                return NSImage(cgImage: cg, size: NSSize(width: cg.width, height: cg.height))
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

    /// A synchronous hit: the decoded image when it is already cached, else
    /// nil. Never reads the disk; a hit moves the key to the newest LRU slot.
    func cachedImage(projectDir: String, relPath: String) -> NSImage? {
        let trimmed = relPath.trimmingCharacters(in: .whitespacesAndNewlines)
        guard !trimmed.isEmpty else { return nil }
        let key = Key(projectDir: projectDir, relPath: trimmed)
        guard case .image(let img) = store[key] else { return nil }
        touch(key)
        return img
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
