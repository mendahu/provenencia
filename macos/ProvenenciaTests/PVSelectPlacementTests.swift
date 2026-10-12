import AppKit
import CoreGraphics
import SwiftUI
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
        #expect(PVSelectPlacement.contentHeight(optionCount: 3, maxVisibleRows: 10, rowHeight: 68)
            == 3 * 68 + PVSelectPlacement.panelPadding)
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

        let tighterThanTheLabel = PVSelectPlacement.triggerSize(
            proposedWidth: 80, fitting: chrome, stretchesHorizontally: true
        )
        #expect(tighterThanTheLabel.width == 80)
    }

    @Test func aTightFrameTruncatesInsteadOfGrowing() {
        let longLabel = CGSize(width: 800, height: 28)
        let framed = PVSelectPlacement.triggerSize(
            proposedWidth: 330, fitting: longLabel, stretchesHorizontally: false
        )
        #expect(framed.width == 330)

        let roomToSpare = PVSelectPlacement.triggerSize(
            proposedWidth: 400, fitting: chrome, stretchesHorizontally: false
        )
        #expect(roomToSpare.width == chrome.width)
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

    @Test @MainActor func aLongClosedLabelStaysInsideTheCallerWidth() {
        let column: CGFloat = 330
        let label = "EVT-RSKRW Residence of James Kenneth Jim Robins"
        let host = NSHostingView(rootView:
            HStack(spacing: 8) {
                PVSelect(
                    selection: .constant("short"),
                    options: [PVSelectOption(value: "short", label: "New Event")],
                    fillsWidth: true
                )
                .frame(width: column)
                PVSelect(
                    selection: .constant("long"),
                    options: [PVSelectOption(value: "long", label: label)],
                    fillsWidth: true
                )
                .frame(width: column)
                Text(verbatim: "STRONG")
            }
            .frame(width: 780, alignment: .leading)
        )
        let window = NSWindow(
            contentRect: NSRect(x: 0, y: 0, width: 900, height: 200),
            styleMask: [.borderless],
            backing: .buffered,
            defer: false
        )
        window.contentView = host
        host.frame = NSRect(x: 0, y: 0, width: 900, height: 80)
        host.layoutSubtreeIfNeeded()
        let frames = PVSelectLayoutProbe.collect(in: host)
        let selects = frames.filter { $0.width > 200 && $0.width <= column + 1 && $0.height >= 20 }
        #expect(selects.count >= 2, "frames \(frames)")
    }

    @Test @MainActor func aShortClosedLabelHugsItsText() {
        let host = NSHostingView(rootView:
            PVSelect(
                selection: .constant("a"),
                options: [PVSelectOption(value: "a", label: "May")],
                fillsWidth: false
            )
        )
        let window = NSWindow(
            contentRect: NSRect(x: 0, y: 0, width: 900, height: 200),
            styleMask: [.borderless],
            backing: .buffered,
            defer: false
        )
        window.contentView = host
        host.frame = NSRect(x: 0, y: 0, width: 900, height: 80)
        host.layoutSubtreeIfNeeded()
        let frames = PVSelectLayoutProbe.collect(in: host)
        #expect(frames.contains { $0.width > 36 && $0.width < 160 && $0.height >= 20 }, "frames \(frames)")
    }

    @Test func triggerNeverReportsZeroSize() {
        let size = PVSelectPlacement.triggerSize(
            proposedWidth: nil, fitting: .zero, stretchesHorizontally: true
        )
        #expect(size.width >= 1)
        #expect(size.height >= 1)
    }
}

private enum PVSelectLayoutProbe {
    @MainActor
    static func collect(in view: NSView) -> [CGRect] {
        var found = [view.bounds]
        for subview in view.subviews {
            found.append(contentsOf: collect(in: subview))
        }
        return found
    }
}
