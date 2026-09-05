import SwiftUI
import AppKit

/// Provenencia design-system color tokens.
///
/// Values are transcribed from the design system's `tokens/colors.css`
/// (Claude Design project "Genealogy app onboarding flow",
/// `b51790c9-7f65-4e91-a7ef-a900095c5870`). See `DesignSystem/README.md`
/// for the full provenance and re-sync process.
///
/// Two layers, matching the CSS:
/// - `PVPalette` — the raw pigment ramps (paper/iron/madder/verdigris/ochre/
///   lapis/plum/copper), each 100/300/500/700/900. Rarely referenced directly.
/// - `PVColor` — semantic aliases (`text-primary`, `surface-card`, `accent`,
///   evidence grades, record types, lineage lines) that resolve to a palette
///   color and automatically flip between the light and dark value CSS
///   defines under `[data-theme="dark"]`. Views should always reach for a
///   `PVColor.*` alias, never a `PVPalette.*` ramp color or a raw hex.
enum PVColor {
    // MARK: Text
    static let textPrimary = Color.pvDynamic(light: PVPalette.paper900, dark: PVPalette.paper100)
    static let textSecondary = Color.pvDynamic(light: PVPalette.paper600, dark: PVPalette.paper300)
    static let textMuted = Color.pvDynamic(light: PVPalette.paper500, dark: PVPalette.paper400)
    static let textFaint = Color.pvDynamic(light: PVPalette.paper400, dark: PVPalette.paper500)
    static let textInverse = Color.pvDynamic(light: PVPalette.paper0, dark: PVPalette.paper950)
    static let textLink = Color.pvDynamic(light: PVPalette.lapis700, dark: PVPalette.lapis300)
    static let textLinkHover = Color.pvDynamic(light: PVPalette.lapis500, dark: PVPalette.hex("#C3CDF0"))
    static let textDisplay = Color.pvDynamic(light: PVPalette.paper950, dark: PVPalette.hex("#FBF7F0"))

    // MARK: Surfaces
    static let surfacePage = Color.pvDynamic(light: PVPalette.paper50, dark: PVPalette.paper950)
    static let surfaceCard = Color.pvDynamic(light: PVPalette.hex("#FFFDF9"), dark: PVPalette.hex("#191510"))
    static let surfaceRaised = Color.pvDynamic(light: PVPalette.paper0, dark: PVPalette.hex("#221D17"))
    static let surfaceSunken = Color.pvDynamic(light: PVPalette.paper100, dark: PVPalette.hex("#0B0907"))
    static let surfaceInset = Color.pvDynamic(light: PVPalette.paper200, dark: PVPalette.hex("#2C2721"))
    static let surfaceHover = Color.pvDynamic(
        light: Color(.sRGB, red: 62.0 / 255, green: 110.0 / 255, blue: 133.0 / 255, opacity: 0.06),
        dark: Color(.sRGB, red: 220.0 / 255, green: 231.0 / 255, blue: 236.0 / 255, opacity: 0.06)
    )
    static let surfaceActive = Color.pvDynamic(
        light: Color(.sRGB, red: 62.0 / 255, green: 110.0 / 255, blue: 133.0 / 255, opacity: 0.12),
        dark: Color(.sRGB, red: 220.0 / 255, green: 231.0 / 255, blue: 236.0 / 255, opacity: 0.12)
    )
    static let surfaceSelected = Color.pvDynamic(light: PVPalette.iron100, dark: PVPalette.hex("#1E2E37"))
    static let surfaceOverlay = Color.pvDynamic(
        light: Color(.sRGB, red: 27.0 / 255, green: 23.0 / 255, blue: 18.0 / 255, opacity: 0.44),
        dark: Color(.sRGB, red: 6.0 / 255, green: 5.0 / 255, blue: 4.0 / 255, opacity: 0.6)
    )

    // MARK: Borders
    static let borderSubtle = Color.pvDynamic(light: PVPalette.paper200, dark: PVPalette.hex("#2B2620"))
    static let borderDefault = Color.pvDynamic(light: PVPalette.paper300, dark: PVPalette.hex("#3C352C"))
    static let borderStrong = Color.pvDynamic(light: PVPalette.paper400, dark: PVPalette.hex("#544B3F"))
    static let borderInked = Color.pvDynamic(light: PVPalette.paper700, dark: PVPalette.paper300)
    static let borderFocus = Color.pvDynamic(light: PVPalette.iron500, dark: PVPalette.iron300)

    // MARK: Accent (iron gall — the house ink)
    static let accent = Color.pvDynamic(light: PVPalette.iron700, dark: PVPalette.iron300)
    static let accentHover = Color.pvDynamic(light: PVPalette.iron900, dark: PVPalette.hex("#A9C4D0"))
    static let accentActive = Color.pvDynamic(light: PVPalette.paper950, dark: PVPalette.hex("#C6DAE3"))
    static let accentForeground = Color.pvDynamic(light: PVPalette.paper0, dark: PVPalette.paper950)
    static let accentSoft = Color.pvDynamic(light: PVPalette.iron100, dark: PVPalette.hex("#1E2E37"))
    static let accentSoftForeground = Color.pvDynamic(light: PVPalette.iron900, dark: PVPalette.hex("#BDD5DF"))
    static let accentLine = Color.pvDynamic(light: PVPalette.iron500, dark: PVPalette.iron300)

    // MARK: Feedback
    static let success = Color.pvDynamic(light: PVPalette.verdigris700, dark: PVPalette.hex("#6FB394"))
    static let successSoft = Color.pvDynamic(light: PVPalette.verdigris100, dark: PVPalette.hex("#152C23"))
    static let successForeground = Color.pvDynamic(light: PVPalette.verdigris900, dark: PVPalette.hex("#B6DCC9"))
    static let warning = Color.pvDynamic(light: PVPalette.ochre700, dark: PVPalette.hex("#DDB35C"))
    static let warningSoft = Color.pvDynamic(light: PVPalette.ochre100, dark: PVPalette.hex("#2E2410"))
    static let warningForeground = Color.pvDynamic(light: PVPalette.ochre900, dark: PVPalette.hex("#F0D79A"))
    static let danger = Color.pvDynamic(light: PVPalette.madder500, dark: PVPalette.hex("#DE8271"))
    static let dangerSoft = Color.pvDynamic(light: PVPalette.madder100, dark: PVPalette.hex("#331611"))
    static let dangerForeground = Color.pvDynamic(light: PVPalette.madder900, dark: PVPalette.hex("#F2BCB0"))
    static let info = Color.pvDynamic(light: PVPalette.lapis500, dark: PVPalette.hex("#8E9DDA"))
    static let infoSoft = Color.pvDynamic(light: PVPalette.lapis100, dark: PVPalette.hex("#181E38"))
    static let infoForeground = Color.pvDynamic(light: PVPalette.lapis900, dark: PVPalette.hex("#C3CDF0"))

    // MARK: Evidence grades (the spine of the product)
    static let evidenceProven = Color.pvDynamic(light: PVPalette.verdigris700, dark: PVPalette.hex("#6FB394"))
    static let evidenceProvenSoft = Color.pvDynamic(light: PVPalette.verdigris100, dark: PVPalette.hex("#152C23"))
    static let evidenceProbable = Color.pvDynamic(light: PVPalette.lapis500, dark: PVPalette.hex("#8E9DDA"))
    static let evidenceProbableSoft = Color.pvDynamic(light: PVPalette.lapis100, dark: PVPalette.hex("#181E38"))
    static let evidencePossible = Color.pvDynamic(light: PVPalette.ochre700, dark: PVPalette.hex("#DDB35C"))
    static let evidencePossibleSoft = Color.pvDynamic(light: PVPalette.ochre100, dark: PVPalette.hex("#2E2410"))
    static let evidenceDisputed = Color.pvDynamic(light: PVPalette.madder500, dark: PVPalette.hex("#DE8271"))
    static let evidenceDisputedSoft = Color.pvDynamic(light: PVPalette.madder100, dark: PVPalette.hex("#331611"))
    static let evidenceUndocumented = Color.pvDynamic(light: PVPalette.paper500, dark: PVPalette.paper400)
    static let evidenceUndocumentedSoft = Color.pvDynamic(light: PVPalette.paper100, dark: PVPalette.hex("#241F19"))

    // MARK: Record types (categorical, deliberately non-sequential)
    static let recordBirth = Color.pvDynamic(light: PVPalette.verdigris500, dark: PVPalette.hex("#6FB394"))
    static let recordMarriage = Color.pvDynamic(light: PVPalette.plum500, dark: PVPalette.hex("#C08FB2"))
    static let recordDeath = Color.pvDynamic(light: PVPalette.iron900, dark: PVPalette.iron300)
    static let recordCensus = Color.pvDynamic(light: PVPalette.lapis500, dark: PVPalette.hex("#8E9DDA"))
    static let recordMigration = Color.pvDynamic(light: PVPalette.copper500, dark: PVPalette.hex("#DB9764"))
    static let recordMilitary = Color.pvDynamic(light: PVPalette.ochre700, dark: PVPalette.hex("#DDB35C"))
    static let recordProbate = Color.pvDynamic(light: PVPalette.paper600, dark: PVPalette.paper300)
    static let recordDNA = Color.pvDynamic(light: PVPalette.madder500, dark: PVPalette.hex("#DE8271"))

    // MARK: Lineage lines
    static let linePaternal = Color.pvDynamic(light: PVPalette.iron500, dark: PVPalette.iron300)
    static let lineMaternal = Color.pvDynamic(light: PVPalette.plum500, dark: PVPalette.hex("#C08FB2"))
    static let lineInferred = Color.pvDynamic(light: PVPalette.paper400, dark: PVPalette.paper600)
}

/// Raw pigment ramps. Prefer `PVColor` semantic aliases at call sites; reach
/// for `PVPalette` only when introducing a new semantic alias.
enum PVPalette {
    // Parchment → ink neutrals (warm, never blue-grey)
    static let paper0 = hex("#FDFBF7")
    static let paper50 = hex("#F8F4ED")
    static let paper100 = hex("#F0EAE0")
    static let paper200 = hex("#E3DACB")
    static let paper300 = hex("#CDC0AC")
    static let paper400 = hex("#A5967F")
    static let paper500 = hex("#7D705D")
    static let paper600 = hex("#5D5346")
    static let paper700 = hex("#433C32")
    static let paper800 = hex("#2C2721")
    static let paper900 = hex("#1B1712")
    static let paper950 = hex("#100E0A")

    // Iron gall — the primary ink
    static let iron100 = hex("#DCE7EC")
    static let iron300 = hex("#8FB1C0")
    static let iron500 = hex("#3E6E85")
    static let iron700 = hex("#27495B")
    static let iron900 = hex("#152F3B")

    // Madder — alarm, contradiction, deletion
    static let madder100 = hex("#F8DFDB")
    static let madder300 = hex("#E0A196")
    static let madder500 = hex("#B23F2D")
    static let madder700 = hex("#892D1F")
    static let madder900 = hex("#591C13")

    // Verdigris — confirmation, verified records
    static let verdigris100 = hex("#D9EBE2")
    static let verdigris300 = hex("#8CC1AC")
    static let verdigris500 = hex("#3F8C6E")
    static let verdigris700 = hex("#2B6550")
    static let verdigris900 = hex("#1A4032")

    // Ochre — inference, highlight, marginalia
    static let ochre100 = hex("#F8EBCE")
    static let ochre300 = hex("#E4C178")
    static let ochre500 = hex("#BF8927")
    static let ochre700 = hex("#95671A")
    static let ochre900 = hex("#5E4110")

    // Lapis — sources, citations, external links
    static let lapis100 = hex("#DEE3F5")
    static let lapis300 = hex("#9AA9E0")
    static let lapis500 = hex("#4257A8")
    static let lapis700 = hex("#2E3D7B")
    static let lapis900 = hex("#1C2550")

    // Plum — unions, marriages, relationship edges
    static let plum100 = hex("#EFDDEA")
    static let plum300 = hex("#C79BBB")
    static let plum500 = hex("#8B4A78")
    static let plum700 = hex("#6A3459")
    static let plum900 = hex("#432038")

    // Copper — migration, movement, in-progress work
    static let copper100 = hex("#FAE3D5")
    static let copper300 = hex("#E8AE86")
    static let copper500 = hex("#C1662B")
    static let copper700 = hex("#95491C")
    static let copper900 = hex("#5D2D10")

    /// Parses a `#RRGGBB` string into a `Color`. Traps on a malformed literal
    /// (all call sites above are compile-time constants copied from
    /// `tokens/colors.css`, so a bad value is a transcription bug, not
    /// something to recover from at runtime).
    static func hex(_ string: String) -> Color {
        var value: UInt64 = 0
        let scanner = Scanner(string: String(string.dropFirst()))
        guard scanner.scanHexInt64(&value) else {
            preconditionFailure("PVPalette.hex: invalid literal \(string)")
        }
        let r = Double((value & 0xFF0000) >> 16) / 255
        let g = Double((value & 0x00FF00) >> 8) / 255
        let b = Double(value & 0x0000FF) / 255
        return Color(.sRGB, red: r, green: g, blue: b, opacity: 1)
    }
}

extension Color {
    /// A color that resolves to `light` or `dark` based on the window's
    /// current appearance, independent of SwiftUI's `colorScheme`
    /// environment (so it works in contexts that don't have one, like
    /// `NSColor`-consuming AppKit call sites). This is the pattern
    /// `docs/macos-client-patterns.md` §4 anticipates when it calls
    /// `Color(nsColor:)` in the blank window "the smallest scale" of the
    /// AppKit escape hatch — `PVColor` is that escape hatch's next step.
    static func pvDynamic(light: Color, dark: Color) -> Color {
        Color(nsColor: NSColor(name: nil) { appearance in
            let isDark = appearance.bestMatch(from: [.aqua, .darkAqua]) == .darkAqua
            return NSColor(isDark ? dark : light)
        })
    }
}
