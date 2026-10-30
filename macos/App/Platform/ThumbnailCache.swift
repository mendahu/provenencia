import AppKit
import Foundation
import ImageIO
import UniformTypeIdentifiers

/// Process-lifetime decoded thumbnail cache. Keys by project dir + relative
/// object path; loads off the main actor and coalesces in-flight requests.
/// Images decode eagerly at no more than `maxPixelSize` on the longest edge,
/// so a full-resolution crop never sits decoded in the cache. Decodes only what
/// `ThumbnailDecodePolicy` accepts, at most `maxConcurrentDecodes` at a time, and checks the object path inside the
/// slot, right before it opens the file. A queued decode with no caller left is dropped.
@MainActor
final class ThumbnailCache {
    static let shared = ThumbnailCache()

    /// Longest-edge pixel cap for a decoded thumbnail (the 120pt header thumb
    /// at 2x, with margin).
    nonisolated static let maxPixelSize = 512

    /// Bound entry count; oldest keys are dropped when exceeded.
    var maxEntries: Int = 100

    /// Most decodes running at once; later requests wait in arrival order.
    var maxConcurrentDecodes: Int = 4

    /// Test helper: the most decodes that ran at once.
    private(set) var peakConcurrentDecodes = 0
    private var runningDecodes = 0
    private var decodeWaiters: [CheckedContinuation<Void, Never>] = []

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
    private var inflight: [Key: Task<Void, Never>] = [:]
    /// Callers awaiting each inflight key; a load with none left is dropped.
    private var waiters: [Key: [UUID: CheckedContinuation<NSImage?, Never>]] = [:]

    private enum DecodeOutcome: Sendable {
        case image(NSImage)
        case refused
        case abandoned
    }

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

        if inflight[key] == nil {
            // Each caller registers in `waiters` before this task first runs.
            inflight[key] = Task { @MainActor in
                let outcome: DecodeOutcome = await self.withDecodeSlot {
                    // A queued decode nobody still wants costs nothing.
                    if Task.isCancelled || self.waiters[key]?.isEmpty ?? true { return .abandoned }
                    guard let url = ProjectFiles.objectURL(projectDir: projectDir, relPath: trimmed) else {
                        return .refused
                    }
                    let image = await Task.detached(priority: .userInitiated) {
                        guard let src = CGImageSourceCreateWithURL(url as CFURL, nil),
                              ThumbnailDecodePolicy.accepts(src),
                              let cg = ImageHeaderDecode.decode(
                                  src, maxPixelSize: ThumbnailCache.maxPixelSize
                              )
                        else { return nil as NSImage? }
                        return NSImage(cgImage: cg, size: NSSize(width: cg.width, height: cg.height))
                    }.value
                    return image.map(DecodeOutcome.image) ?? .refused
                }
                self.finish(key, outcome)
            }
        }
        return await awaitLoad(key)
    }

    /// Waits for the inflight load of `key`. A cancelled caller leaves the wait
    /// at once with nil, and the load drops if no caller is left.
    private func awaitLoad(_ key: Key) async -> NSImage? {
        let id = UUID()
        return await withTaskCancellationHandler {
            await withCheckedContinuation { (cont: CheckedContinuation<NSImage?, Never>) in
                if Task.isCancelled {
                    cont.resume(returning: nil)
                } else {
                    waiters[key, default: [:]][id] = cont
                }
            }
        } onCancel: {
            Task { @MainActor in
                if let cont = self.waiters[key]?.removeValue(forKey: id) { cont.resume(returning: nil) }
            }
        }
    }

    /// Stores a real result and hands it to the callers still waiting. An
    /// abandoned load stores nothing.
    private func finish(_ key: Key, _ outcome: DecodeOutcome) {
        let image: NSImage?
        switch outcome {
        case .image(let img): image = img
        case .refused, .abandoned: image = nil
        }
        inflight[key] = nil
        if case .abandoned = outcome {} else { insert(key, image.map(Entry.image) ?? .missing) }
        for (_, cont) in waiters.removeValue(forKey: key) ?? [:] {
            cont.resume(returning: image)
        }
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

    /// Test helper: seeded entry count after loads.
    var entryCount: Int { store.count }

    /// Runs `body` holding one decode slot, like Go's `raster.Slot`.
    private func withDecodeSlot<T: Sendable>(_ body: () async -> T) async -> T {
        if runningDecodes < max(1, maxConcurrentDecodes) {
            runningDecodes += 1
        } else {
            await withCheckedContinuation { decodeWaiters.append($0) } // slot handed over
        }
        peakConcurrentDecodes = max(peakConcurrentDecodes, runningDecodes)
        defer {
            if decodeWaiters.isEmpty {
                runningDecodes -= 1
            } else {
                decodeWaiters.removeFirst().resume()
            }
        }
        return await body()
    }

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

/// What `ThumbnailCache` decodes: an image Go's `raster` package could have
/// written. Matches `raster.MaxEdgeLimit`.
enum ThumbnailDecodePolicy {
    static let maxDeclaredEdge = 4096
    static let allowedTypes: Set<String> = [UTType.png.identifier, UTType.jpeg.identifier]

    /// True for a PNG or JPEG whose first image declares both edges within
    /// `maxDeclaredEdge`. Reads the header only; no pixels decode.
    static func accepts(_ source: CGImageSource) -> Bool {
        guard let type = CGImageSourceGetType(source) as String?, allowedTypes.contains(type) else {
            return false
        }
        return ImageHeaderDecode.declaredImage(source)?
            .fits(maxEdge: maxDeclaredEdge, maxPixels: maxDeclaredEdge * maxDeclaredEdge) == true
    }
}
