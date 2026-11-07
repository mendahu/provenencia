import Foundation
import Testing

@testable import Provenencia

@Suite struct HistoryJumpMenuModelTests {
    @Test(arguments: [
        (CGFloat(1102), CGFloat(12), CGFloat(600)),
        (CGFloat(436), CGFloat(12), CGFloat(408)),
        (CGFloat(0), CGFloat(12), CGFloat(600)),
        (CGFloat(628), CGFloat(12), CGFloat(600)),
        (CGFloat(627), CGFloat(12), CGFloat(599)),
        (CGFloat(436), CGFloat(32), CGFloat(388)),
        (CGFloat(436), CGFloat(82), CGFloat(338)),
    ])
    func panelWidthFitsTheColumn(contentWidth: CGFloat, anchorMinX: CGFloat, expected: CGFloat) {
        #expect(
            HistoryJumpMenuModel.fittedPanelWidth(contentWidth: contentWidth, anchorMinX: anchorMinX)
                == expected)
    }
}
