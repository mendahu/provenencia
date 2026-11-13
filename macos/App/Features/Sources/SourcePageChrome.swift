import SwiftUI

/// Shared layout metrics for Source page section rows (the notes byline column).
enum SourcePageLayout {
    /// Author / timestamp column on note rows and the composer byline.
    static let notesBylineWidth: CGFloat = 150
}

/// Gap above a Source page section, beyond the page stack's own spacing.
struct SourcePageSectionTopPadding: ViewModifier {
    func body(content: Content) -> some View {
        content.padding(.top, PVSpacing.space11 - PVSpacing.space9)
    }
}

extension View {
    func sourcePageSectionTopPadding() -> some View {
        modifier(SourcePageSectionTopPadding())
    }
}
