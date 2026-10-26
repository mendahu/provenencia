import SwiftUI

/// Provenencia design-system stroke tokens.
enum PVStroke {
    /// The dash pattern of every provisional or placeholder outline.
    static let dashPattern: [CGFloat] = [4, 3]

    /// A 1 pt dashed outline: provisional cards, empty states, blocked tokens.
    static let dashed = StrokeStyle(lineWidth: 1, dash: dashPattern)
}
