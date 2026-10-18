import AppKit
import Foundation
import Testing
@testable import Provenencia

@Suite
@MainActor
struct ThumbnailCacheTests {
    @Test func emptyRelPathReturnsNilWithoutCaching() async {
        let cache = ThumbnailCache()
        let image = await cache.image(
            projectDir: "/tmp/x.provenencia",
            relPath: "  "
        )
        #expect(image == nil)
        #expect(cache.entryCount == 0)
    }

    @Test func missingFileCachesSentinel() async {
        let cache = ThumbnailCache()
        let dir = FileManager.default.temporaryDirectory
            .appendingPathComponent("thumb-cache-\(UUID().uuidString)", isDirectory: true)
        try? FileManager.default.createDirectory(at: dir, withIntermediateDirectories: true)

        let first = await cache.image(
            projectDir: dir.path,
            relPath: "objects/ab/cd/missing"
        )
        #expect(first == nil)
        #expect(cache.entryCount == 1)

        let second = await cache.image(
            projectDir: dir.path,
            relPath: "objects/ab/cd/missing"
        )
        #expect(second == nil)
        #expect(cache.entryCount == 1)
    }

    @Test func symlinkedObjectCachesSentinel() async throws {
        let fm = FileManager.default
        let cache = ThumbnailCache()
        let project = ArtifactFileFixtures.uniqueProjectDir()
        let rel = "objects/ab/cd/x.png"
        let outside = try ArtifactFileFixtures.linkToOutsidePNG(at: project + "/" + rel)
        defer {
            try? fm.removeItem(atPath: project)
            try? fm.removeItem(atPath: outside)
        }
        let image = await cache.image(projectDir: project, relPath: rel)
        #expect(image == nil)
        #expect(cache.entryCount == 1)
    }

    @Test func loadsAndHitsCache() async throws {
        let cache = ThumbnailCache()
        let dir = FileManager.default.temporaryDirectory
            .appendingPathComponent("thumb-cache-\(UUID().uuidString)", isDirectory: true)
        let objectDir = dir.appendingPathComponent("objects/aa/bb", isDirectory: true)
        try FileManager.default.createDirectory(at: objectDir, withIntermediateDirectories: true)

        let fileURL = objectDir.appendingPathComponent("thumb.jpg")
        // Minimal 1×1 PNG
        let png = Data(base64Encoded:
            "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8z8BQDwAEhQGAhKmMIQAAAABJRU5ErkJggg=="
        )!
        try png.write(to: fileURL)

        let first = await cache.image(
            projectDir: dir.path,
            relPath: "objects/aa/bb/thumb.jpg"
        )
        #expect(first != nil)
        #expect(cache.entryCount == 1)

        let second = await cache.image(
            projectDir: dir.path,
            relPath: "objects/aa/bb/thumb.jpg"
        )
        #expect(second != nil)
        #expect(cache.entryCount == 1)
    }

    @Test func decodesAtABoundedPixelSize() async throws {
        let cache = ThumbnailCache()
        let dir = FileManager.default.temporaryDirectory
            .appendingPathComponent("thumb-cache-\(UUID().uuidString)", isDirectory: true)
        let objectDir = dir.appendingPathComponent("objects/aa/bb", isDirectory: true)
        try FileManager.default.createDirectory(at: objectDir, withIntermediateDirectories: true)

        let rep = NSBitmapImageRep(
            bitmapDataPlanes: nil, pixelsWide: 2000, pixelsHigh: 1000,
            bitsPerSample: 8, samplesPerPixel: 4, hasAlpha: true, isPlanar: false,
            colorSpaceName: .deviceRGB, bytesPerRow: 0, bitsPerPixel: 0
        )!
        let png = try #require(rep.representation(using: .png, properties: [:]))
        try png.write(to: objectDir.appendingPathComponent("big.png"))

        let image = try #require(await cache.image(
            projectDir: dir.path, relPath: "objects/aa/bb/big.png"))
        let cg = try #require(image.cgImage(forProposedRect: nil, context: nil, hints: nil))
        #expect(max(cg.width, cg.height) <= ThumbnailCache.maxPixelSize)
    }

    @Test func cachedImageIsAMissUntilLoaded() async throws {
        let cache = ThumbnailCache()
        let dir = FileManager.default.temporaryDirectory
            .appendingPathComponent("thumb-cache-\(UUID().uuidString)", isDirectory: true)
        let objectDir = dir.appendingPathComponent("objects/aa/bb", isDirectory: true)
        try FileManager.default.createDirectory(at: objectDir, withIntermediateDirectories: true)
        let png = Data(base64Encoded:
            "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8z8BQDwAEhQGAhKmMIQAAAABJRU5ErkJggg=="
        )!
        try png.write(to: objectDir.appendingPathComponent("thumb.png"))

        #expect(cache.cachedImage(projectDir: dir.path, relPath: "objects/aa/bb/thumb.png") == nil)
        #expect(cache.entryCount == 0)
        let loaded = await cache.image(projectDir: dir.path, relPath: "objects/aa/bb/thumb.png")
        let hit = cache.cachedImage(projectDir: dir.path, relPath: "objects/aa/bb/thumb.png")
        #expect(hit != nil)
        #expect(hit === loaded)
    }

    @Test func coalescesInflightLoads() async throws {
        let cache = ThumbnailCache()
        let dir = FileManager.default.temporaryDirectory
            .appendingPathComponent("thumb-cache-\(UUID().uuidString)", isDirectory: true)
        let objectDir = dir.appendingPathComponent("objects/cc/dd", isDirectory: true)
        try FileManager.default.createDirectory(at: objectDir, withIntermediateDirectories: true)

        let fileURL = objectDir.appendingPathComponent("shared.jpg")
        let png = Data(base64Encoded:
            "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8z8BQDwAEhQGAhKmMIQAAAABJRU5ErkJggg=="
        )!
        try png.write(to: fileURL)

        async let a = cache.image(
            projectDir: dir.path,
            relPath: "objects/cc/dd/shared.jpg"
        )
        async let b = cache.image(
            projectDir: dir.path,
            relPath: "objects/cc/dd/shared.jpg"
        )
        let (imgA, imgB) = await (a, b)
        #expect(imgA != nil)
        #expect(imgB != nil)
        #expect(cache.entryCount == 1)
    }

    @Test func cachedThumbnailFallsBackToTheGivenMark() {
        let given = CachedThumbnail.fallbackContent(
            fallbackMark: .subjectPerson, fallbackLabel: nil,
            typeIconKey: "type_book", mediaType: "", originalFilename: ""
        )
        guard case .mark(let key)? = given.glyph else {
            Issue.record("expected a mark, got \(String(describing: given.glyph))")
            return
        }
        #expect(key == .subjectPerson)

        let typed = CachedThumbnail.fallbackContent(
            fallbackMark: nil, fallbackLabel: nil,
            typeIconKey: "type_book", mediaType: "", originalFilename: ""
        )
        guard case .mark(let typeKey)? = typed.glyph else {
            Issue.record("expected the type mark, got \(String(describing: typed.glyph))")
            return
        }
        #expect(typeKey == PVMarkKey(catalogKey: "type_book"))

        let bare = CachedThumbnail.fallbackContent(
            fallbackMark: nil, fallbackLabel: nil,
            typeIconKey: nil, mediaType: "", originalFilename: ""
        )
        #expect(bare.glyph == nil && bare.image == nil && !bare.loading)
    }
}
