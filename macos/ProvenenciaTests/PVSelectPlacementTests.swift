import CoreGraphics
import Testing
@testable import Provenencia

@Suite
struct PVSelectPlacementTests {
    private let screen = CGRect(x: 0, y: 0, width: 1440, height: 875)

    private func anchor(y: CGFloat, width: CGFloat = 136) -> CGRect {
        CGRect(x: 100, y: y, width: width, height: 28)
    }

    @Test func popupHangsBelowWhenThereIsRoom() {
        let field = anchor(y: 600)
        let height = PVSelectPlacement.contentHeight(optionCount: 4)
        let frame = PVSelectPlacement.popupFrame(
            anchor: field, contentHeight: height, visibleFrame: screen
        )
        #expect(frame.maxY == field.minY - PVSelectPlacement.gap)
        #expect(frame.height == height)
        #expect(frame.minX == field.minX)
        #expect(frame.width == PVSelectPlacement.defaultMenuWidth)
        #expect(!frame.intersects(field))
    }

    @Test func popupFlipsAboveWhenItWillNotFitBelow() {
        let field = anchor(y: 60)
        let height = PVSelectPlacement.contentHeight(optionCount: 8)
        let frame = PVSelectPlacement.popupFrame(
            anchor: field, contentHeight: height, visibleFrame: screen
        )
        #expect(frame.minY == field.maxY + PVSelectPlacement.gap)
        #expect(frame.height == height)
    }

    @Test func popupNeverOverlapsTheTrigger() {
        let height = PVSelectPlacement.contentHeight(optionCount: 6)
        for y in stride(from: CGFloat(0), through: 860, by: 20) {
            let field = anchor(y: y)
            let frame = PVSelectPlacement.popupFrame(
                anchor: field, contentHeight: height, visibleFrame: screen
            )
            #expect(!frame.intersects(field), "overlapped the trigger at y=\(y)")
        }
    }

    @Test func popupStaysInsideTheVisibleFrame() {
        let height = PVSelectPlacement.contentHeight(optionCount: 6)
        for y in stride(from: CGFloat(0), through: 860, by: 20) {
            let frame = PVSelectPlacement.popupFrame(
                anchor: anchor(y: y), contentHeight: height, visibleFrame: screen
            )
            #expect(screen.contains(frame) || frame.height == 0, "escaped the screen at y=\(y)")
        }
    }

    @Test func heightTrimsToAvailableSpaceAndMaxVisibleRows() {
        let cramped = CGRect(x: 0, y: 0, width: 1440, height: 300)
        let field = anchor(y: 100)
        let wanted = PVSelectPlacement.contentHeight(optionCount: 20, maxVisibleRows: 10)
        let frame = PVSelectPlacement.popupFrame(
            anchor: field, contentHeight: wanted, visibleFrame: cramped
        )
        #expect(frame.height < wanted)
        #expect(cramped.contains(frame))
        #expect(PVSelectPlacement.contentHeight(optionCount: 20) < 20 * PVSelectPlacement.rowHeight + 20)
        #expect(PVSelectPlacement.contentHeight(optionCount: 20, maxVisibleRows: 10)
            == 10 * PVSelectPlacement.rowHeight + PVSelectPlacement.panelPadding)
    }

    @Test func widthIsTheGreaterOfTriggerAndMenuWidthThenClamped() {
        let narrow = CGRect(x: 100, y: 600, width: 40, height: 28)
        let wide = PVSelectPlacement.popupFrame(
            anchor: narrow,
            contentHeight: 80,
            visibleFrame: screen,
            menuWidth: 210
        )
        #expect(wide.width == 210)

        let field = CGRect(x: 100, y: 600, width: 300, height: 28)
        let follows = PVSelectPlacement.popupFrame(
            anchor: field,
            contentHeight: 80,
            visibleFrame: screen,
            menuWidth: 210
        )
        #expect(follows.width == 300)

        let offRight = CGRect(x: 1300, y: 600, width: 300, height: 28)
        let clamped = PVSelectPlacement.popupFrame(
            anchor: offRight,
            contentHeight: 80,
            visibleFrame: screen,
            menuWidth: 210
        )
        #expect(clamped.maxX <= screen.maxX)
        #expect(clamped.minX >= screen.minX)
    }

    @Test func menuOriginMapsABelowFrameBackThroughContextPlacement() {
        let field = CGRect(x: 200, y: 400, width: 180, height: 28)
        let frame = PVSelectPlacement.popupFrame(
            anchor: field, contentHeight: 96, visibleFrame: screen
        )
        let origin = PVSelectPlacement.menuOrigin(anchor: field, frame: frame)
        let restored = PVContextMenuPlacement.panelFrame(
            anchorOnScreen: field,
            origin: origin,
            size: frame.size
        )
        #expect(restored.minX == frame.minX)
        #expect(restored.maxY == frame.maxY)
        #expect(restored.width == frame.width)
        #expect(restored.height == frame.height)
    }

    // MARK: Trigger sizing (#181)

    private let chrome = CGSize(width: 136, height: 28)

    @Test func triggerKeepsChromeHeightWhateverTheStackProposes() {
        for stretches in [true, false] {
            let size = PVSelectPlacement.triggerSize(
                proposedWidth: 400,
                fitting: chrome,
                stretchesHorizontally: stretches
            )
            #expect(size.height == chrome.height)
        }
    }

    @Test func fullWidthFieldFollowsTheProposedWidth() {
        let wide = PVSelectPlacement.triggerSize(
            proposedWidth: 400, fitting: chrome, stretchesHorizontally: true
        )
        #expect(wide.width == 400)

        let infinite = PVSelectPlacement.triggerSize(
            proposedWidth: .infinity, fitting: chrome, stretchesHorizontally: true
        )
        #expect(infinite.width == .infinity)

        let narrow = PVSelectPlacement.triggerSize(
            proposedWidth: 0, fitting: chrome, stretchesHorizontally: true
        )
        #expect(narrow.width == chrome.width)

        let unspecified = PVSelectPlacement.triggerSize(
            proposedWidth: nil, fitting: chrome, stretchesHorizontally: true
        )
        #expect(unspecified.width == chrome.width)
    }

    @Test func chipIgnoresTheProposedWidth() {
        for proposed in [CGFloat?.none, 0, 400, .infinity] {
            let size = PVSelectPlacement.triggerSize(
                proposedWidth: proposed, fitting: chrome, stretchesHorizontally: false
            )
            #expect(size.width == chrome.width)
            #expect(size.height == chrome.height)
        }
    }

    @Test func triggerNeverReportsZeroSize() {
        let size = PVSelectPlacement.triggerSize(
            proposedWidth: nil, fitting: .zero, stretchesHorizontally: true
        )
        #expect(size.width >= 1)
        #expect(size.height >= 1)
    }
}
