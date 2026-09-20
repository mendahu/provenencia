import SwiftUI

/// The Provenencia logo mark — a seal roundel enclosing a stylized "P," in
/// the brand's iron-gall teal palette. Wraps `Assets.xcassets/LogoMark`
/// (a vector-preserving SVG image set with light/dark appearance variants,
/// same schema as `AccentColor.colorset`) so call sites never hardcode the
/// asset name directly — this is the first screen-facing use of the mark;
/// add new placements through this file, not through `Image("LogoMark")`.
///
/// See `DesignSystem/README.md` for brand-asset provenance (the mark also
/// backs `AppIcon.appiconset`, generated via `scripts/generate-app-icon.sh`)
/// and `PVButton.swift` for the component pattern this follows.
struct PVLogoMark: View {
    private let size: CGFloat

    init(size: CGFloat = 48) {
        self.size = size
    }

    var body: some View {
        Image("LogoMark")
            .resizable()
            .scaledToFit()
            .frame(width: size, height: size)
            .accessibilityHidden(true)
    }
}

#Preview {
    HStack(spacing: PVSpacing.space6) {
        PVLogoMark(size: 32)
        PVLogoMark(size: 48)
        PVLogoMark(size: 64)
    }
    .padding(PVSpacing.space9)
    .background(PVColor.surfacePage)
}
