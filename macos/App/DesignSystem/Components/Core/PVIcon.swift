import SwiftUI

/// Wraps SF Symbols behind the same semantic names the web design system
/// uses for its Lucide icons — mirrors `components/core/Icon.jsx`, whose
/// own doc comment (in the source design system's readme) calls it "a
/// wrapper over the substituted icon set, so swapping icon systems later
/// is a one-file change." On macOS, Lucide (a web-only substitution, per
/// that readme) becomes SF Symbols instead; `PVIconMap` below is the
/// one file to extend when a screen needs an icon name that isn't mapped
/// yet.
///
/// Only the names actually consumed so far (the Onboarding Flow board's
/// icons, plus `PVToast`'s tone icons and dismiss glyph) are mapped. See
/// `DesignSystem/README.md` for the full component/token pattern this file
/// follows (canonical example: `PVButton.swift`).
struct PVIcon: View {
    private let name: String
    private let size: CGFloat

    init(_ name: String, size: CGFloat = 16) {
        self.name = name
        self.size = size
    }

    var body: some View {
        Image(systemName: PVIconMap.sfSymbol(for: name))
            .resizable()
            .scaledToFit()
            .frame(width: size, height: size)
            .accessibilityHidden(true)
    }
}

/// Lucide name -> SF Symbol name. Add an entry here when a new icon name
/// is needed; falls back to `question.mark` (visibly wrong, easy to spot)
/// rather than crashing on an unmapped name.
enum PVIconMap {
    private static let symbols: [String: String] = [
        "folder-plus": "folder.badge.plus",
        "folder-open": "folder",
        "check": "checkmark",
        "info": "info.circle",
        "chevron-down": "chevron.down",
        "check-check": "checkmark.circle.fill",
        "triangle-alert": "exclamationmark.triangle.fill",
        "octagon-alert": "exclamationmark.octagon.fill",
        "x": "xmark"
    ]

    static func sfSymbol(for lucideName: String) -> String {
        symbols[lucideName] ?? "questionmark"
    }
}

#Preview {
    HStack(spacing: PVSpacing.space6) {
        PVIcon("folder-plus")
        PVIcon("folder-open")
        PVIcon("check")
        PVIcon("info")
        PVIcon("chevron-down")
    }
    .foregroundStyle(PVColor.accent)
    .padding(PVSpacing.space9)
    .background(PVColor.surfacePage)
}
