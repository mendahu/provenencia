import SwiftUI
import AppKit

/// Provenencia design-system type tokens.
///
/// Transcribed from `tokens/fonts.css` and `tokens/typography.css`. All
/// three brand families are serif: **Newsreader** for display/headings,
/// **Spectral** for body/UI, **IBM Plex Mono** for anything a researcher
/// would cite exactly (dates, ids, folio references, counts). The actual
/// `.ttf` files are bundled under `macos/App/Resources/Fonts/` and
/// registered at launch — see `PVFontRegistration` below and
/// `DesignSystem/README.md` for the full weight list and provenance.
enum PVFontFamily {
    static let display = "Newsreader"
    static let body = "Spectral"
    static let mono = "IBM Plex Mono"
}

/// Font weights, matching `--weight-*`. Raw values are the CSS numeric
/// weights (used when resolving the installed custom font face), not
/// `Font.Weight`, since we resolve against `NSFont` — see `PVFont`.
enum PVFontWeight {
    static let light: CGFloat = 300
    static let regular: CGFloat = 400
    static let medium: CGFloat = 500
    static let semibold: CGFloat = 600
    static let bold: CGFloat = 700
}

/// Type scale, matching `--text-*` (all values in points; CSS px maps 1:1
/// to SwiftUI points).
enum PVTypeScale {
    static let display1: CGFloat = 56
    static let display2: CGFloat = 44
    static let display3: CGFloat = 34
    static let h1: CGFloat = 27
    static let h2: CGFloat = 22
    static let h3: CGFloat = 18
    static let h4: CGFloat = 16
    static let body: CGFloat = 16
    static let bodySmall: CGFloat = 14
    static let caption: CGFloat = 13
    static let micro: CGFloat = 11
}

/// Line heights, matching `--lh-*` (multipliers of font size).
enum PVLineHeight {
    static let tight: CGFloat = 1.08
    static let snug: CGFloat = 1.25
    static let normal: CGFloat = 1.5
    static let relaxed: CGFloat = 1.65
}

/// Letter spacing (tracking), matching `--tracking-*`, in points of
/// `kerning`/`tracking` per the SwiftUI `Text` modifiers of the same name.
/// CSS values are `em`-relative; these are pre-multiplied against
/// `PVTypeScale` at the two call sites that use non-zero tracking
/// (`--tracking-display` on display type, `--tracking-caps` on eyebrows) —
/// see usage notes in `DesignSystem/README.md`.
enum PVTracking {
    static let display: CGFloat = -0.02
    static let normal: CGFloat = 0
    static let wide: CGFloat = 0.02
    static let caps: CGFloat = 0.11
}

/// Resolves Provenencia's brand fonts by family name + numeric weight
/// against whatever weight faces are registered (see
/// `PVFontRegistration.registerBundledFontsIfNeeded()`), falling back to
/// the closest system serif/monospace face if a bundled font is somehow
/// unavailable (e.g. running in a context that skipped registration) —
/// never crashes.
enum PVFont {
    static func display(size: CGFloat = PVTypeScale.h2, weight: CGFloat = PVFontWeight.semibold) -> Font {
        resolved(family: PVFontFamily.display, size: size, weight: weight, italic: false, fallbackDesign: .serif)
    }

    static func body(size: CGFloat = PVTypeScale.body, weight: CGFloat = PVFontWeight.regular, italic: Bool = false) -> Font {
        resolved(family: PVFontFamily.body, size: size, weight: weight, italic: italic, fallbackDesign: .serif)
    }

    static func mono(size: CGFloat = PVTypeScale.caption, weight: CGFloat = PVFontWeight.regular) -> Font {
        resolved(family: PVFontFamily.mono, size: size, weight: weight, italic: false, fallbackDesign: .monospaced)
    }

    private static func resolved(
        family: String,
        size: CGFloat,
        weight: CGFloat,
        italic: Bool,
        fallbackDesign: Font.Design
    ) -> Font {
        var traits: [NSFontDescriptor.TraitKey: Any] = [.weight: nsWeight(for: weight)]
        if italic {
            traits[.symbolic] = NSFontDescriptor.SymbolicTraits.italic.rawValue
        }
        let descriptor = NSFontDescriptor(fontAttributes: [.family: family, .traits: traits])
        if let nsFont = NSFont(descriptor: descriptor, size: size) {
            return Font(nsFont)
        }
        return .system(size: size, weight: systemWeight(for: weight), design: fallbackDesign)
    }

    /// CSS numeric weight (300-700) -> the roughly-equivalent `NSFont.Weight`
    /// raw value, so `NSFontDescriptor` resolution asks for the same weight
    /// tier the family actually ships faces for.
    private static func nsWeight(for weight: CGFloat) -> CGFloat {
        switch weight {
        case ..<PVFontWeight.regular: return NSFont.Weight.light.rawValue
        case PVFontWeight.regular..<PVFontWeight.medium: return NSFont.Weight.regular.rawValue
        case PVFontWeight.medium..<PVFontWeight.semibold: return NSFont.Weight.medium.rawValue
        case PVFontWeight.semibold..<PVFontWeight.bold: return NSFont.Weight.semibold.rawValue
        default: return NSFont.Weight.bold.rawValue
        }
    }

    private static func systemWeight(for weight: CGFloat) -> Font.Weight {
        switch weight {
        case ..<PVFontWeight.regular: return .light
        case PVFontWeight.regular..<PVFontWeight.medium: return .regular
        case PVFontWeight.medium..<PVFontWeight.semibold: return .medium
        case PVFontWeight.semibold..<PVFontWeight.bold: return .semibold
        default: return .bold
        }
    }
}

/// Registers the bundled brand font files with Core Text at process
/// launch. Call once, early (e.g. from `ProvenenciaApp.init`) — see
/// `DesignSystem/README.md` for why this is used instead of
/// `ATSApplicationFontsPath`.
enum PVFontRegistration {
    private static var didRegister = false

    static func registerBundledFontsIfNeeded() {
        guard !didRegister else { return }
        didRegister = true
        guard let urls = Bundle.main.urls(forResourcesWithExtension: "ttf", subdirectory: "Fonts") else { return }
        for url in urls {
            // Ignore per-file errors: "already registered" is common (e.g.
            // repeated calls, or the font also installed system-wide) and
            // isn't actionable here.
            CTFontManagerRegisterFontsForURL(url as CFURL, .process, nil)
        }
    }
}
