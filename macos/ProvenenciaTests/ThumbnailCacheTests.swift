import AppKit
import Foundation
import ImageIO
import Testing
@testable import Provenencia

fileprivate struct DecodePolicyCase: Sendable {
    let label: String
    let fileType: NSBitmapImageRep.FileType?
    let width: Int
    let height: Int
    let expected: Bool
}

fileprivate let decodePolicyCases: [DecodePolicyCase] = [
    DecodePolicyCase(label: "png 8x8", fileType: .png, width: 8, height: 8, expected: true),
    DecodePolicyCase(label: "jpeg 8x8", fileType: .jpeg, width: 8, height: 8, expected: true),
    DecodePolicyCase(label: "tiff 8x8", fileType: .tiff, width: 8, height: 8, expected: false),
    DecodePolicyCase(label: "png 4096x8", fileType: .png, width: 4096, height: 8, expected: true),
    DecodePolicyCase(label: "png 4097x8", fileType: .png, width: 4097, height: 8, expected: false),
    DecodePolicyCase(label: "png 8x4097", fileType: .png, width: 8, height: 4097, expected: false),
    DecodePolicyCase(label: "random bytes", fileType: nil, width: 0, height: 0, expected: false),
]

@Suite
@MainActor
struct ThumbnailCacheTests {
    @Test(arguments: decodePolicyCases)
    fileprivate func decodePolicyAcceptsOnlyGoImages(_ c: DecodePolicyCase) throws {
        let project = ArtifactFileFixtures.uniqueProjectDir()
        defer { try? FileManager.default.removeItem(atPath: project) }
        let rel = "objects/aa/bb/x.png"
        if let fileType = c.fileType {
            try ArtifactFileFixtures.writeImage(
                projectDir: project, relPath: rel, width: c.width, height: c.height, fileType: fileType)
        } else {
            let bytes = Data((0..<64).map { UInt8(truncatingIfNeeded: $0 &* 37) })
            try ArtifactFileFixtures.write(bytes, projectDir: project, relPath: rel)
        }
        let url = URL(fileURLWithPath: project).appendingPathComponent(rel)
        let accepted = CGImageSourceCreateWithURL(url as CFURL, nil).map(ThumbnailDecodePolicy.accepts) ?? false
        #expect(accepted == c.expected)
    }

    @Test func loadsAGoJPEG() async throws {
        let project = ArtifactFileFixtures.uniqueProjectDir()
        defer { try? FileManager.default.removeItem(atPath: project) }
        let rel = "objects/aa/bb/x.jpg"
        try ArtifactFileFixtures.writeImage(projectDir: project, relPath: rel, fileType: .jpeg)
        #expect(await ThumbnailCache().image(projectDir: project, relPath: rel) != nil)
    }

    @Test func refusesAnotherFormat() async throws {
        let project = ArtifactFileFixtures.uniqueProjectDir()
        defer { try? FileManager.default.removeItem(atPath: project) }
        let cache = ThumbnailCache()
        try ArtifactFileFixtures.writeImage(projectDir: project, relPath: "objects/aa/bb/x.png", fileType: .tiff)
        #expect(await cache.image(projectDir: project, relPath: "objects/aa/bb/x.png") == nil)
        #expect(cache.entryCount == 1)
        #expect(cache.cachedImage(projectDir: project, relPath: "objects/aa/bb/x.png") == nil)
        try ArtifactFileFixtures.writePNG(projectDir: project, relPath: "objects/aa/bb/y.png")
        #expect(await cache.image(projectDir: project, relPath: "objects/aa/bb/y.png") != nil)
    }

    @Test func refusesAnOversizedDeclaredEdge() async throws {
        let project = ArtifactFileFixtures.uniqueProjectDir()
        defer { try? FileManager.default.removeItem(atPath: project) }
        let cache = ThumbnailCache()
        try ArtifactFileFixtures.writePNG(projectDir: project, relPath: "objects/aa/bb/big.png", width: 4097, height: 8)
        try ArtifactFileFixtures.writePNG(projectDir: project, relPath: "objects/aa/bb/edge.png", width: 4096, height: 8)
        #expect(await cache.image(projectDir: project, relPath: "objects/aa/bb/big.png") == nil)
        #expect(await cache.image(projectDir: project, relPath: "objects/aa/bb/edge.png") != nil)
    }

    @Test(.timeLimit(.minutes(1))) func runsAtMostTheCappedDecodes() async throws {
        let project = ArtifactFileFixtures.uniqueProjectDir()
        defer { try? FileManager.default.removeItem(atPath: project) }
        let cache = ThumbnailCache()
        cache.maxConcurrentDecodes = 2
        for i in 0..<8 {
            try ArtifactFileFixtures.writePNG(
                projectDir: project, relPath: "objects/aa/bb/k\(i).png", width: 64, height: 64)
        }
        let decodes = (0..<8).map { i in
            Task { await cache.image(projectDir: project, relPath: "objects/aa/bb/k\(i).png") != nil }
        }
        var results: [Bool] = []
        for decode in decodes { results.append(await decode.value) }
        #expect(results.count == 8 && results.allSatisfy { $0 })
        #expect((1...2).contains(cache.peakConcurrentDecodes))
    }

    @Test(.timeLimit(.minutes(1))) func dropsAQueuedDecodeWhoseCallersHaveGone() async throws {
        let project = ArtifactFileFixtures.uniqueProjectDir()
        defer { try? FileManager.default.removeItem(atPath: project) }
        let cache = ThumbnailCache()
        cache.maxConcurrentDecodes = 1
        for i in 0..<3 {
            try ArtifactFileFixtures.writePNG(
                projectDir: project, relPath: "objects/aa/bb/k\(i).png", width: 64, height: 64)
        }
        let tasks = (0..<3).map { i in
            Task { @MainActor in
                await cache.image(projectDir: project, relPath: "objects/aa/bb/k\(i).png")
            }
        }
        tasks[1].cancel()
        let first = await tasks[0].value
        let second = await tasks[1].value
        let third = await tasks[2].value
        #expect(first != nil)
        #expect(second == nil)
        #expect(third != nil)
        // The dropped key is not stored as missing, so a fresh request decodes it.
        #expect(cache.entryCount == 2)
        #expect(await cache.image(projectDir: project, relPath: "objects/aa/bb/k1.png") != nil)
    }

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

        // Control: a real file at the same path loads (a fresh cache, since the
        // old one keeps its sentinel).
        try fm.removeItem(atPath: project + "/" + rel)
        try ArtifactFileFixtures.writePNG(projectDir: project, relPath: rel)
        #expect(await ThumbnailCache().image(projectDir: project, relPath: rel) != nil)
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
