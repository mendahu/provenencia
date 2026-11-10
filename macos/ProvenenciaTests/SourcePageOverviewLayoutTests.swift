import AppKit
import SwiftUI
import Testing
@testable import Provenencia

@MainActor
private final class OverviewFrames {
    var error: CGRect?
    var body: CGRect?
    var side: CGRect?
    var rest: CGRect?
}

private struct OverviewProbe: View {
    let frames: OverviewFrames
    let widthClass: PVWidthClass
    let errorHeight: CGFloat

    var body: some View {
        SourcePageOverviewLayout(widthClass: widthClass) {
            Color.clear.frame(height: errorHeight)
                .onGeometryChange(for: CGRect.self) { $0.frame(in: .named("overview")) } action: { frames.error = $0 }
            Color.clear.frame(minHeight: 100)
                .onGeometryChange(for: CGRect.self) { $0.frame(in: .named("overview")) } action: { frames.body = $0 }
            Color.clear.frame(height: 140)
                .onGeometryChange(for: CGRect.self) { $0.frame(in: .named("overview")) } action: { frames.side = $0 }
            Color.clear.frame(height: 80)
                .onGeometryChange(for: CGRect.self) { $0.frame(in: .named("overview")) } action: { frames.rest = $0 }
        }
        .frame(maxWidth: .infinity, maxHeight: .infinity, alignment: .topLeading)
        .coordinateSpace(name: "overview")
    }
}

@MainActor @Suite
struct SourcePageOverviewLayoutTests {
    private func host(
        width: CGFloat, widthClass: PVWidthClass, errorHeight: CGFloat
    ) async -> (OverviewFrames, NSWindow) {
        let frames = OverviewFrames()
        let (window, _) = hostInBorderlessWindow(
            OverviewProbe(frames: frames, widthClass: widthClass, errorHeight: errorHeight),
            windowSize: CGSize(width: width, height: 800)
        )
        _ = await waitUntil { frames.error != nil && frames.body != nil && frames.side != nil && frames.rest != nil }
        return (frames, window)
    }

    @Test func regularPlacesDescriptionBesideMetadataThenTheRest() async throws {
        let (frames, window) = await host(width: 1038, widthClass: .regular, errorHeight: 30)
        let error = try #require(frames.error), body = try #require(frames.body)
        let side = try #require(frames.side), rest = try #require(frames.rest)
        let half: CGFloat = (1038 - PVSpacing.space11) / 2
        #expect(error.minX == 0 && error.minY == 0 && error.width == 1038)
        #expect(body.width == half && side.width == half)
        #expect(body.minY == side.minY && body.minY == 30 + PVSpacing.space9)
        #expect(side.minX == half + PVSpacing.space11)
        // Both are as tall as the taller, so Credibility sits at the bottom of its column.
        #expect(body.height == 140 && side.height == 140)
        #expect(rest.minX == 0 && rest.width == 1038 && rest.minY == body.maxY + PVSpacing.space9)
        withExtendedLifetime(window) {}
    }

    @Test func anEmptyErrorAddsNoGap() async throws {
        let (frames, window) = await host(width: 1038, widthClass: .regular, errorHeight: 0)
        let body = try #require(frames.body)
        #expect(body.minY == 0)
        withExtendedLifetime(window) {}
    }

    @Test func widePlacesMetadataInASideColumn() async throws {
        let group = PVPageColumnsMetrics.split(
            available: 2296 - 2 * PVSpacing.gutterPage, sideMax: PVSpacing.widthSecondaryMax
        )
        let (frames, window) = await host(width: group.group, widthClass: .wide, errorHeight: 30)
        let error = try #require(frames.error), body = try #require(frames.body)
        let side = try #require(frames.side), rest = try #require(frames.rest)
        #expect(group.main == 1180 && group.side == 480)
        #expect(error.width == 1180 && body.width == 1180 && rest.width == 1180)
        #expect(error.minX == 0 && body.minX == 0 && rest.minX == 0)
        #expect(error.minY == 0)
        #expect(body.minY == 30 + PVSpacing.space9)
        #expect(rest.minY == body.maxY + PVSpacing.space9)
        #expect(side.minX == 1180 + PVSpacing.space9 && side.minY == 0 && side.width == 480)
        withExtendedLifetime(window) {}
    }
}
