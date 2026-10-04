import Foundation
import Testing
@testable import Provenencia

@Suite
struct L10nResolutionTests {
    @Test func stringMatchesFoundationResolution() {
        let resources = [
            L10n.DesignSystem.toastDismiss,
            L10n.DesignSystem.requiredMarker,
            L10n.EvidenceGraph.addProperty,
        ]
        for resource in resources {
            #expect(L10n.string(resource) == String(localized: resource))
        }
    }

    @Test func formatFillsCatalogFormat() {
        #expect(
            L10n.EvidenceGraph.deleteAccessibility(kind: "Person", label: "Ada Lovelace", ref: "PER-1")
                == "Delete Person Ada Lovelace, PER-1"
        )
        #expect(L10n.DesignSystem.selectOptionPosition(current: 2, count: 5) == "2 of 5")
    }

    @Test func formatWithExplicitLocaleUsesIt() {
        #expect(
            L10n.Dates.displayBetween(start: "1850", end: "1860", locale: Locale(identifier: "en_US"))
                == "Between 1850 and 1860"
        )
    }

    @Test func formattingLocaleKeepsCurrentWhenLanguagesMatch() {
        let current = Locale(identifier: "en_CA")
        let locale = L10n.formattingLocale(current: current, resolvedLocalization: "en")
        #expect(locale.identifier == current.identifier)
    }

    @Test func formattingLocaleTakesResolvedLanguageAndKeepsRegion() {
        let locale = L10n.formattingLocale(current: Locale(identifier: "en_US"), resolvedLocalization: "fr")
        #expect(locale.language.languageCode == "fr")
        #expect(locale.region == "US")
    }

    @Test func formattingLocaleWithoutResolvedLocalizationIsCurrent() {
        let current = Locale(identifier: "de_DE")
        #expect(L10n.formattingLocale(current: current, resolvedLocalization: nil).identifier == current.identifier)
    }

    /// Guards against copy resolving through `String(localized: LocalizedStringResource)`,
    /// which re-parses the strings table on every call (~1.5 ms; 1,000 calls ≈ 1.5 s).
    /// The cached path takes a few milliseconds for the same work.
    @Test func formatDoesNotReparseTheStringsTable() {
        let start = ContinuousClock.now
        for index in 0..<1_000 {
            _ = L10n.EvidenceGraph.deleteAccessibility(kind: "Person", label: "Ada", ref: "PER-\(index)")
        }
        #expect(ContinuousClock.now - start < .milliseconds(250))
    }
}
