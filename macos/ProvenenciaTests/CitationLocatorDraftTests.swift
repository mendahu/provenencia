import CoreGraphics
import Testing
@testable import Provenencia

@Suite
struct CitationLocatorDraftTests {
    @Test func pdfRegionStampsPageWhenMissing() {
        var draft = CitationLocatorDraft.artifactOnly()
        let stamped = draft.setRegion(Self.rectangle, capabilities: Self.pdf, viewerPage: 3)
        #expect(stamped)
        #expect(draft.page == 3)
        #expect(draft.hasRegion)
    }

    @Test func pdfRegionRestampsPageAfterClearingRegion() {
        var draft = CitationLocatorDraft.artifactOnly()
        let first = draft.setRegion(Self.rectangle, capabilities: Self.pdf, viewerPage: 3)
        #expect(first)
        draft.clearRegion()
        #expect(draft.page == 3)
        #expect(!draft.hasRegion)

        let restamped = draft.setRegion(Self.rectangle, capabilities: Self.pdf, viewerPage: 5)
        #expect(restamped)
        #expect(draft.page == 5)
        #expect(draft.hasRegion)
    }

    @Test func imageRegionDoesNotWritePage() {
        var draft = CitationLocatorDraft.artifactOnly()
        let ok = draft.setRegion(Self.rectangle, capabilities: Self.image, viewerPage: 2)
        #expect(ok)
        #expect(draft.page == nil)
        #expect(draft.hasRegion)
    }

    private static let pdf = ArtifactViewerKind.pdf.locatorCapabilities
    private static let image = ArtifactViewerKind.image.locatorCapabilities
    private static let rectangle = ArtifactRegionDraft(
        kind: .rectangle,
        points: [
            CGPoint(x: 0.1, y: 0.1),
            CGPoint(x: 0.4, y: 0.1),
            CGPoint(x: 0.4, y: 0.4),
            CGPoint(x: 0.1, y: 0.4),
        ]
    )
}
