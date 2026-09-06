import CoreGraphics

/// Provenencia design-system spacing tokens.
///
/// Transcribed from `tokens/spacing.css` — a 2px base scale, non-linear
/// above 32px. CSS px maps 1:1 to SwiftUI points.
enum PVSpacing {
    static let space0: CGFloat = 0
    static let spacePx: CGFloat = 1
    static let space1: CGFloat = 2
    static let space2: CGFloat = 4
    static let space3: CGFloat = 6
    static let space4: CGFloat = 8
    static let space5: CGFloat = 12
    static let space6: CGFloat = 16
    static let space7: CGFloat = 20
    static let space8: CGFloat = 24
    static let space9: CGFloat = 32
    static let space10: CGFloat = 40
    static let space11: CGFloat = 48
    static let space12: CGFloat = 64
    static let space13: CGFloat = 80
    static let space14: CGFloat = 96
    static let space15: CGFloat = 128

    static let gutterPage: CGFloat = 32
    static let gutterPageWide: CGFloat = 56

    /// `--measure-prose` / `--measure-narrow` are `ch`-relative (character
    /// widths), which has no direct SwiftUI equivalent — approximate with
    /// `.frame(maxWidth:)` using these point values (66/46 characters at
    /// `PVTypeScale.body` in Spectral is roughly 480pt / 340pt).
    static let measureProse: CGFloat = 480
    static let measureNarrow: CGFloat = 340
    /// Onboarding form column width (choose-file / identify pages).
    static let measureForm: CGFloat = 560

    static let widthSidebar: CGFloat = 264
    static let widthInspector: CGFloat = 340
    static let widthContentMax: CGFloat = 1180
    /// Minimum onboarding window width.
    static let widthWindowMin: CGFloat = 700

    /// Control heights. **Deliberately not the web `--control-h-*` pixel
    /// values (28/34/42)** — the source design system's own readme
    /// documents AppKit metrics as the intended macOS platform deviation.
    /// See `DesignSystem/README.md` "Platform deviations".
    static let controlHeightSmall: CGFloat = 22
    static let controlHeightMedium: CGFloat = 28
    static let controlHeightLarge: CGFloat = 36

    static let hitMin: CGFloat = 44
}
