import AppKit
import CoreGraphics
import Foundation
import Testing
@testable import Provenencia

@Suite(.serialized)
@MainActor
struct CitationAutoTranscribeTests {
    private let sourceID = "src-1"
    private let subjectID = "sub-person-1"
    private let personTypeID = "type-person"
    private let eventTypeID = "type-event"
    private let placeTypeID = "type-place"
    private let namePropertyID = "prop-name"

    @Test func pdfArtifactDoesNotSendToOCR() async {
        let projectDir = uniqueProjectDir()
        let store = makeStore(projectDir: projectDir)
        seedArtifact(store, mediaType: "application/pdf", relPath: "objects/file-0.pdf")
        let engine = FakeOCREngine()
        let model = makeModel(store: store, projectDir: projectDir, engine: engine)
        await model.prepare()
        #expect(!model.canAutoTranscribe)
        #expect(model.isPDFArtifact)
        model.requestAutoTranscribe()
        #expect(engine.recognizeCount == 0)
        #expect(String(localized: model.autoTranscribeHint) == String(localized: L10n.CitationComposer.autoTranscribeHintPDF))
    }

    @Test func missingImageFileDisablesButton() async {
        let projectDir = uniqueProjectDir()
        let store = makeStore(projectDir: projectDir)
        seedArtifact(store, mediaType: "image/png", relPath: "objects/ab/cd/missing.png")
        let engine = FakeOCREngine()
        let model = makeModel(store: store, projectDir: projectDir, engine: engine)
        await model.prepare()
        #expect(model.isImageArtifact)
        #expect(model.imageRaster == nil)
        #expect(!model.canAutoTranscribe)
        model.requestAutoTranscribe()
        #expect(engine.recognizeCount == 0)
        #expect(
            String(localized: model.autoTranscribeHint)
                == String(localized: L10n.CitationComposer.autoTranscribeHintMissingFile)
        )
    }

    @Test func audioArtifactDisablesButton() async {
        let projectDir = uniqueProjectDir()
        let store = makeStore(projectDir: projectDir)
        seedArtifact(store, mediaType: "audio/mpeg", relPath: "objects/file-0.mp3")
        let engine = FakeOCREngine()
        let model = makeModel(store: store, projectDir: projectDir, engine: engine)
        await model.prepare()
        #expect(!model.canAutoTranscribe)
        model.requestAutoTranscribe()
        #expect(engine.recognizeCount == 0)
        #expect(
            String(localized: model.autoTranscribeHint)
                == String(localized: L10n.CitationComposer.autoTranscribeHintAudio)
        )
    }

    @Test func fillWritesTranscriptionAndDirtiesFields() async throws {
        let projectDir = uniqueProjectDir()
        let store = makeStore(projectDir: projectDir)
        try writePNG(projectDir: projectDir, relPath: "objects/ab/cd/sample.png")
        seedArtifact(store, mediaType: "image/png", relPath: "objects/ab/cd/sample.png")
        let engine = FakeOCREngine()
        engine.result = .success("Alderwick")
        let model = makeModel(store: store, projectDir: projectDir, engine: engine)
        await model.prepare()
        #expect(model.canAutoTranscribe)
        #expect(!model.fields.isDirty)
        await model.runAutoTranscribe()
        #expect(model.transcription == "Alderwick")
        #expect(model.fields.isDirty)
        #expect(model.shouldHoldLeave)
        #expect(model.identityMenusDisabled)
        #expect(model.transcriptionOCRMessage == nil)
        #expect(!model.transcriptionUncertain)
        #expect(engine.recognizeCount == 1)
    }

    @Test func emptyResultKeepsPriorText() async throws {
        let projectDir = uniqueProjectDir()
        let store = makeStore(projectDir: projectDir)
        try writePNG(projectDir: projectDir, relPath: "objects/ab/cd/sample.png")
        seedArtifact(store, mediaType: "image/png", relPath: "objects/ab/cd/sample.png")
        let engine = FakeOCREngine()
        engine.result = .success("   ")
        let model = makeModel(store: store, projectDir: projectDir, engine: engine)
        await model.prepare()
        model.transcription = "kept"
        model.fields.captureBaseline()
        await model.runAutoTranscribe()
        #expect(model.transcription == "kept")
        #expect(
            model.transcriptionOCRMessage
                == String(localized: L10n.CitationComposer.autoTranscribeNothingFound)
        )
    }

    @Test func failureKeepsPriorText() async throws {
        let projectDir = uniqueProjectDir()
        let store = makeStore(projectDir: projectDir)
        try writePNG(projectDir: projectDir, relPath: "objects/ab/cd/sample.png")
        seedArtifact(store, mediaType: "image/png", relPath: "objects/ab/cd/sample.png")
        let engine = FakeOCREngine()
        engine.result = .failure(TestOCRError.failed)
        let model = makeModel(store: store, projectDir: projectDir, engine: engine)
        await model.prepare()
        model.transcription = "kept"
        model.fields.captureBaseline()
        await model.runAutoTranscribe()
        #expect(model.transcription == "kept")
        #expect(
            model.transcriptionOCRMessage
                == String(localized: L10n.CitationComposer.autoTranscribeFailed)
        )
    }

    @Test func replaceDoesNotOverwriteWithoutConfirm() async throws {
        let projectDir = uniqueProjectDir()
        let store = makeStore(projectDir: projectDir)
        try writePNG(projectDir: projectDir, relPath: "objects/ab/cd/sample.png")
        seedArtifact(store, mediaType: "image/png", relPath: "objects/ab/cd/sample.png")
        let engine = FakeOCREngine()
        engine.result = .success("new")
        let model = makeModel(store: store, projectDir: projectDir, engine: engine)
        await model.prepare()
        model.transcription = "existing"
        model.requestAutoTranscribe()
        #expect(engine.recognizeCount == 0)
        #expect(model.transcription == "existing")
        #expect(model.pendingTranscriptionConfirm?.existingText == "existing")
        #expect(model.pendingTranscriptionConfirm?.includeWholePage == false)
        model.pendingTranscriptionConfirm = nil
        await model.runAutoTranscribe()
        #expect(model.transcription == "new")
    }

    @Test func wholePageWarnsOnlyWhenOversizedAndNoRegion() async throws {
        let projectDir = uniqueProjectDir()
        let store = makeStore(projectDir: projectDir)
        try writePNG(projectDir: projectDir, relPath: "objects/ab/cd/sample.png")
        seedArtifact(store, mediaType: "image/png", relPath: "objects/ab/cd/sample.png")
        let engine = FakeOCREngine()
        let model = makeModel(store: store, projectDir: projectDir, engine: engine)
        await model.prepare()
        #expect(!model.needsWholePageWarning)
        await model.runAutoTranscribe()
        #expect(engine.recognizeCount == 1)

        model.artifactViewer.installImageRasterForTesting(
            NSImage(size: NSSize(width: 6200, height: 8400))
        )
        engine.resetCount()
        model.requestAutoTranscribe()
        #expect(model.needsWholePageWarning)
        #expect(model.pendingTranscriptionConfirm?.includeWholePage == true)
        #expect(engine.recognizeCount == 0)

        let points = ArtifactRegionGeometry.rectanglePoints(
            from: CGPoint(x: 0.1, y: 0.1),
            to: CGPoint(x: 0.4, y: 0.4)
        )
        model.artifactViewer.installImageRasterForTesting(
            try #require(pixelExactImage(width: 8, height: 8))
        )
        #expect(model.setRegion(ArtifactRegionDraft(kind: .rectangle, points: points)))
        model.transcription = ""
        model.pendingTranscriptionConfirm = nil
        #expect(!model.needsWholePageWarning)
        await model.runAutoTranscribe()
        #expect(engine.recognizeCount == 1)
    }

    @Test func regionCropIsSmallerThanFullImage() async throws {
        let projectDir = uniqueProjectDir()
        let store = makeStore(projectDir: projectDir)
        try writePNG(projectDir: projectDir, relPath: "objects/ab/cd/sample.png")
        seedArtifact(store, mediaType: "image/png", relPath: "objects/ab/cd/sample.png")
        let engine = FakeOCREngine()
        let model = makeModel(store: store, projectDir: projectDir, engine: engine)
        await model.prepare()
        let raster = try #require(pixelExactImage(width: 8, height: 8))
        model.artifactViewer.installImageRasterForTesting(raster)
        let points = ArtifactRegionGeometry.rectanglePoints(
            from: CGPoint(x: 0.25, y: 0.25),
            to: CGPoint(x: 0.75, y: 0.75)
        )
        #expect(model.setRegion(ArtifactRegionDraft(kind: .rectangle, points: points)))
        await model.runAutoTranscribe()
        #expect(engine.lastSize == CGSize(width: 4, height: 4))
    }

    @Test func disabledWhileRunningAndLocatorHeld() async throws {
        let projectDir = uniqueProjectDir()
        let store = makeStore(projectDir: projectDir)
        try writePNG(projectDir: projectDir, relPath: "objects/ab/cd/sample.png")
        seedArtifact(store, mediaType: "image/png", relPath: "objects/ab/cd/sample.png")
        let engine = FakeOCREngine()
        engine.shouldHold = true
        let model = makeModel(store: store, projectDir: projectDir, engine: engine)
        await model.prepare()
        let task = Task { await model.runAutoTranscribe() }
        #expect(await waitUntil { model.isTranscribing })
        #expect(!model.canAutoTranscribe)
        #expect(model.identityMenusDisabled)
        #expect(model.shouldHoldLeave)
        let points = ArtifactRegionGeometry.rectanglePoints(
            from: CGPoint(x: 0.1, y: 0.1),
            to: CGPoint(x: 0.2, y: 0.2)
        )
        #expect(!model.setRegion(ArtifactRegionDraft(kind: .rectangle, points: points)))
        engine.release()
        await task.value
        #expect(!model.isTranscribing)
        #expect(model.canAutoTranscribe)
    }

    @Test func editingFieldClearsCallout() async throws {
        let projectDir = uniqueProjectDir()
        let store = makeStore(projectDir: projectDir)
        try writePNG(projectDir: projectDir, relPath: "objects/ab/cd/sample.png")
        seedArtifact(store, mediaType: "image/png", relPath: "objects/ab/cd/sample.png")
        let engine = FakeOCREngine()
        engine.result = .failure(TestOCRError.failed)
        let model = makeModel(store: store, projectDir: projectDir, engine: engine)
        await model.prepare()
        await model.runAutoTranscribe()
        #expect(model.transcriptionOCRMessage != nil)
        model.transcription = "typed"
        #expect(model.transcriptionOCRMessage == nil)
    }

    private func uniqueProjectDir() -> String {
        FileManager.default.temporaryDirectory
            .appendingPathComponent("pv-auto-transcribe-\(UUID().uuidString).provenencia", isDirectory: true)
            .path
    }

    private func writePNG(projectDir: String, relPath: String, width: Int = 8, height: Int = 8) throws {
        let image = NSImage(size: NSSize(width: width, height: height))
        image.lockFocus()
        NSColor.red.setFill()
        NSBezierPath(rect: NSRect(x: 0, y: 0, width: width, height: height)).fill()
        image.unlockFocus()
        guard let tiff = image.tiffRepresentation,
              let rep = NSBitmapImageRep(data: tiff),
              let data = rep.representation(using: .png, properties: [:])
        else {
            throw TestOCRError.failed
        }
        let url = URL(fileURLWithPath: projectDir).appendingPathComponent(relPath)
        try FileManager.default.createDirectory(
            at: url.deletingLastPathComponent(),
            withIntermediateDirectories: true
        )
        try data.write(to: url)
    }

    private func makeStore(projectDir: String) -> FakeStore {
        let store = FakeStore()
        store.subjectTypesByProject[projectDir] = [
            CatalogSubjectType(
                id: personTypeID, key: "person", origin: "provenencia", label: "Person",
                description: "", refPrefix: "PER", candidateRefPrefix: "CPER"
            ),
            CatalogSubjectType(
                id: eventTypeID, key: "event", origin: "provenencia", label: "Event",
                description: "", refPrefix: "EVT", candidateRefPrefix: "CEVT"
            ),
            CatalogSubjectType(
                id: placeTypeID, key: "place", origin: "provenencia", label: "Place",
                description: "", refPrefix: "PLC", candidateRefPrefix: "CPLC"
            ),
        ]
        store.propertiesByProject[projectDir] = [
            CatalogProperty(
                id: namePropertyID, key: "name", origin: "provenencia", label: "Name",
                description: "", valueType: "name"
            ),
        ]
        store.subjectTypeFieldsByType[personTypeID] = [
            CatalogSubjectTypeField(
                property: store.propertiesByProject[projectDir]![0],
                sortOrder: 0,
                locked: false
            ),
        ]
        store.connectRules = CatalogConnectRule.productMatrix
        store.sourcesByProject[projectDir] = [
            CatalogSource(
                id: sourceID, ref: "SRC-1", sourceTypeID: "stype", title: "Census", description: ""
            ),
        ]
        store.subjectsBySource[sourceID] = [
            CatalogSubject(
                id: subjectID, ref: "CPR-1", sourceID: sourceID, subjectTypeID: personTypeID,
                label: "Margt.", description: ""
            ),
        ]
        store.subjectPositionsBySubject[subjectID] = CatalogSubjectPosition(
            subjectID: subjectID, gridX: 0, gridY: 0
        )
        return store
    }

    private func seedArtifact(_ store: FakeStore, mediaType: String, relPath: String) {
        store.artifactsBySource[sourceID] = [
            CatalogArtifact(
                id: "art-0",
                ref: "ART-0",
                sourceID: sourceID,
                fileID: "file-0",
                label: "Scan 1",
                description: "",
                file: CatalogFileRef(
                    id: "file-0",
                    relPath: relPath,
                    originalFilename: URL(fileURLWithPath: relPath).lastPathComponent,
                    mediaType: mediaType,
                    byteSize: 10
                )
            ),
        ]
    }

    private func makeModel(
        store: FakeStore,
        projectDir: String,
        engine: FakeOCREngine
    ) -> CitationComposerModel {
        CitationComposerModel(
            entry: .addProperty(sourceID: sourceID, subjectID: subjectID, artifactID: "art-0"),
            session: WorkspaceSession(projectKey: ProjectKey(projectDir: projectDir), store: store),
            store: store,
            userID: "user-1",
            ocrEngine: engine
        )
    }

    private func pixelExactImage(width: Int, height: Int) -> NSImage? {
        let colorSpace = CGColorSpaceCreateDeviceRGB()
        guard let context = CGContext(
            data: nil,
            width: width,
            height: height,
            bitsPerComponent: 8,
            bytesPerRow: 0,
            space: colorSpace,
            bitmapInfo: CGImageAlphaInfo.premultipliedLast.rawValue
        ), let cgImage = context.makeImage() else { return nil }
        return NSImage(cgImage: cgImage, size: NSSize(width: width, height: height))
    }

    private func waitUntil(
        timeout: Duration = .seconds(2),
        _ predicate: () -> Bool
    ) async -> Bool {
        let deadline = ContinuousClock.now + timeout
        while ContinuousClock.now < deadline {
            if predicate() { return true }
            try? await Task.sleep(for: .milliseconds(10))
        }
        return predicate()
    }
}

private enum TestOCRError: Error {
    case failed
}

private final class FakeOCREngine: OCREngine, @unchecked Sendable {
    private let lock = NSLock()
    var result: Result<String, Error> = .success("recognized")
    var shouldHold = false
    private var hold: CheckedContinuation<Void, Never>?
    private(set) var lastSize: CGSize?
    private(set) var recognizeCount = 0

    func recognizeText(in image: CGImage) async throws -> String {
        lock.lock()
        recognizeCount += 1
        lastSize = CGSize(width: image.width, height: image.height)
        let shouldHold = self.shouldHold
        let result = self.result
        lock.unlock()
        if shouldHold {
            await withCheckedContinuation { (continuation: CheckedContinuation<Void, Never>) in
                lock.lock()
                hold = continuation
                lock.unlock()
            }
        }
        return try result.get()
    }

    func release() {
        lock.lock()
        let continuation = hold
        hold = nil
        lock.unlock()
        continuation?.resume()
    }

    func resetCount() {
        lock.lock()
        recognizeCount = 0
        lock.unlock()
    }
}
