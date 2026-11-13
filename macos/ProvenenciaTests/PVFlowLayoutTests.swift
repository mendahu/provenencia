import SwiftUI
import Testing
@testable import Provenencia

@MainActor
struct PVFlowLayoutTests {
    /// Arranges items of these ideal widths (40pt tall); a wider-than-offered item is
    /// re-measured at the offered width. `calls` records every `fit` call.
    private func arrange(
        widths: [CGFloat], rowWidth: CGFloat, calls: inout [String]
    ) -> (frames: [CGRect], height: CGFloat) {
        var log: [String] = []
        let result = PVFlowLayout.arrange(
            width: rowWidth, spacing: 10, lineSpacing: 6, count: widths.count
        ) { index, offered in
            log.append("\(index):\(offered.map { String(Double($0)) } ?? "nil")")
            return CGSize(width: offered.map { min(widths[index], $0) } ?? widths[index], height: 40)
        }
        calls = log
        return result
    }

    @Test func wrapsAsBefore() {
        var calls: [String] = []
        let one = arrange(widths: [100, 120], rowWidth: 250, calls: &calls)
        #expect(one.frames.map(\.minX) == [0, 110])
        #expect(one.height == 40)
        let two = arrange(widths: [100, 200], rowWidth: 250, calls: &calls)
        #expect(two.frames[1].minX == 0 && two.frames[1].minY == 46)
        #expect(two.height == 86)
    }

    @Test func widerThanTheRowGetsTheRowWidth() {
        var calls: [String] = []
        let result = arrange(widths: [300], rowWidth: 250, calls: &calls)
        #expect(result.frames[0].width == 250 && result.frames[0].height == 40)
        #expect(result.height == 40)
        #expect(calls.contains("0:250.0"))
    }

    @Test func stillWiderThanTheRowIsPlacedAtTheRowWidth() {
        let result = PVFlowLayout.arrange(
            width: 250, spacing: 10, lineSpacing: 6, count: 2
        ) { index, offered in
            // The first item ignores the offer: ideal 300, 280 when offered 250.
            if index == 0 { return CGSize(width: offered == nil ? 300 : 280, height: 40) }
            return CGSize(width: 50, height: 40)
        }
        #expect(result.frames[0].width == 250)
        #expect(result.frames[1].minX == 0 && result.frames[1].minY == 46)
    }

    @Test func exactlyTheRowWidthStaysIdeal() {
        var calls: [String] = []
        let result = arrange(widths: [250], rowWidth: 250, calls: &calls)
        #expect(result.frames[0].width == 250)
        #expect(calls == ["0:nil"])
    }

    @Test func idealProbeReMeasuresNothing() {
        var calls: [String] = []
        let zero = arrange(widths: [300, 50], rowWidth: 0, calls: &calls)
        #expect(calls == ["0:nil", "1:nil"])
        #expect(zero.frames[1].minX == 0 && zero.frames[1].minY == 46)
        let infinite = arrange(widths: [300, 50], rowWidth: .infinity, calls: &calls)
        #expect(calls == ["0:nil", "1:nil"])
        #expect(infinite.frames[1].minX == 310 && infinite.frames[1].minY == 0)
        let none = arrange(widths: [], rowWidth: 250, calls: &calls)
        #expect(none.frames.isEmpty && none.height == 0)
    }
}
