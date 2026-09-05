import CoreGraphics

/// Provenencia design-system corner-radius tokens.
///
/// Transcribed from `tokens/radii.css`: "Documents have crisp corners;
/// controls get a small, bookbinder's radius."
enum PVRadius {
    static let none: CGFloat = 0
    static let xs: CGFloat = 2
    static let sm: CGFloat = 3
    static let md: CGFloat = 5
    static let lg: CGFloat = 8
    static let xl: CGFloat = 12
    static let xxl: CGFloat = 18
    static let pill: CGFloat = 999

    // `--radius-seal` (50%) isn't a fixed point value — it means "fully
    // round," which in SwiftUI is `Circle()` / `.clipShape(Circle())`,
    // not a `CGFloat` radius. There's no `PVRadius.seal` constant on
    // purpose; use `Circle()` at the call site instead.
}
