import CoreGraphics

/// Screen-space placement for a Select menu. AppKit coordinates (y grows up).
/// Width follows the trigger with `menuWidth` as a minimum — not ComboBox's
/// "never wider than the field."
enum PVSelectPlacement {
    static let gap: CGFloat = 4
    static let defaultMenuWidth: CGFloat = 210
    static let defaultMaxVisibleRows = 10
    /// Matches `PVContextMenuItem` body + vertical padding.
    static let rowHeight: CGFloat = 32
    /// `PVContextMenuPanel` padding (`space2` × 2).
    static let panelPadding: CGFloat = 8

    static func contentHeight(
        optionCount: Int,
        maxVisibleRows: Int = defaultMaxVisibleRows,
        rowHeight: CGFloat = rowHeight
    ) -> CGFloat {
        let rows = min(max(optionCount, 0), max(maxVisibleRows, 0))
        guard rows > 0 else { return 0 }
        return CGFloat(rows) * rowHeight + panelPadding
    }

    /// Screen frame for the menu window.
    static func popupFrame(
        anchor: CGRect,
        contentHeight: CGFloat,
        visibleFrame: CGRect,
        menuWidth: CGFloat = defaultMenuWidth,
        gap: CGFloat = gap
    ) -> CGRect {
        let spaceBelow = max(0, anchor.minY - visibleFrame.minY - gap)
        let spaceAbove = max(0, visibleFrame.maxY - anchor.maxY - gap)
        let opensUp = contentHeight > spaceBelow && spaceAbove > spaceBelow
        let height = min(contentHeight, opensUp ? spaceAbove : spaceBelow)
        let y = opensUp ? anchor.maxY + gap : anchor.minY - gap - height
        let width = min(max(anchor.width, menuWidth, 1), visibleFrame.width)
        let x = min(max(anchor.minX, visibleFrame.minX), visibleFrame.maxX - width)
        return CGRect(x: x, y: y, width: width, height: height)
    }

    /// SwiftUI-downward origin inside the trigger that `PVContextMenuPlacement`
    /// maps back to `frame`.
    static func menuOrigin(anchor: CGRect, frame: CGRect) -> CGPoint {
        CGPoint(x: frame.minX - anchor.minX, y: anchor.maxY - frame.maxY)
    }

    /// Layout size of the trigger for a SwiftUI proposal. The trigger is a
    /// control, not a container: it always takes its own chrome height, and
    /// only a full-width field follows the proposed width (like
    /// `.frame(maxWidth: .infinity)`). Accepting the proposal outright made
    /// stacks hand the select all remaining space (#181).
    static func triggerSize(
        proposedWidth: CGFloat?,
        fitting: CGSize,
        stretchesHorizontally: Bool
    ) -> CGSize {
        var width = fitting.width
        if stretchesHorizontally, let proposedWidth {
            width = max(proposedWidth, fitting.width)
        }
        return CGSize(width: max(width, 1), height: max(fitting.height, 1))
    }
}
