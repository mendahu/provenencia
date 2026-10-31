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

    @Test func formerlyInterpolatedCopyFormatsItsArguments() {
        #expect(L10n.NameValue.partMoveUp(position: 3) == "Move part 3 up")
        #expect(L10n.CitationComposer.pasteHintSelected(lines: 4, page: 12) == "Pastes the 4 lines selected on page 12")
        #expect(L10n.SourceTypes.assignFieldNamed(field: "Author", type: "Book") == "Assign field Author to Book")
        #expect(L10n.Workspace.personCount(1) == "1 person")
        #expect(L10n.Workspace.personCount(3) == "3 persons")
        #expect(L10n.SourceTypes.usage(count: 0) == "no sources yet")
        #expect(L10n.SourceTypes.usage(count: 7) == "in use on 7 sources")
    }

    @Test func unpromoteCopyResolves() {
        #expect(L10n.string(L10n.EvidenceGraph.unpromote) == "Unpromote")
        #expect(L10n.string(L10n.EvidenceGraph.unpromoteKeep) == "Keep promotion")
        #expect(L10n.EvidenceGraph.unpromoteConfirmTitle(name: "James Robins") == "Unpromote James Robins?")
        #expect(
            L10n.string(L10n.EvidenceGraph.unpromoteConfirmMessage)
                == "This promotion is removed. The cards stay on the graph, and the handle stays."
        )
        #expect(L10n.string(L10n.EvidenceGraph.unpromoteConfirmBridgesHeader) == "Also unpromoted")
        #expect(L10n.string(L10n.EvidenceGraph.unpromoteConfirmBridgesNote) == "The other end of each stays promoted.")
        #expect(
            L10n.Promote.duplicateBody(.event)
                == "These may be the same event. Combine them on the Evidence graph, or pick another target for one row."
        )
        #expect(L10n.Promote.openedFrom(name: "Gracie") == "Opened from Gracie")
    }

    @Test func promoteCountsUseSingularForms() {
        #expect(L10n.Promote.promotedTitle(promoted: 1, connections: 0) == "Promoted 1 subject and 0 connections")
        #expect(L10n.Promote.promotedTitle(promoted: 2, connections: 1) == "Promoted 2 subjects and 1 connection")
        #expect(L10n.Promote.promotedBody(skipped: 1) == "1 subject is still on the graph, not promoted yet")
        #expect(L10n.Promote.promotedBody(skipped: 3) == "3 subjects are still on the graph, not promoted yet")
        #expect(L10n.Promote.rowSummary(rows: 1, promoted: 1) == "1 row · 1 already promoted")
        #expect(L10n.Promote.rowSummary(rows: 4, promoted: 1) == "4 rows · 1 already promoted")
    }

    @Test func promoteDuplicateWarningsSayWhatIsSelected() {
        #expect(L10n.Promote.duplicateTitle(name: "Gracie", ref: "PER-1") == "Gracie has also been selected to be promoted to PER-1")
        #expect(L10n.Promote.duplicateNewTitle(name: "Gracie") == "Gracie has also been selected to be promoted as New and looks the same")
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

    @Test func errorCodesHaveCopy() {
        let unknown = L10n.Errors.message(code: "internal.unknown")
        let expected = [
            "evidencegraphs.invalid": "That evidence graph isn’t valid.",
            "evidencegraphs.artifact_in_use": "Citations in this evidence graph still use an artifact you removed, so it stays linked.",
            "evidencegraphs.last_artifact": "An evidence graph needs at least one artifact.",
            "evidencegraphs.in_use": "Subjects still need this evidence graph.",
            "evidencegraphs.last_graph": "A source keeps at least one evidence graph.",
            "observations.other_graph": "That subject or artifact is in another evidence graph.",
        ]
        for (code, sentence) in expected {
            let message = L10n.Errors.message(code: code)
            #expect(message == sentence)
            #expect(message != unknown)
        }
    }

    @Test func depictionSaveErrorsHaveCopy() {
        let unknown = L10n.Errors.message(code: "internal.unknown")
        let image = L10n.Errors.message(code: "ingest.unsupported_image")
        let invalid = L10n.Errors.message(code: "depictionvalues.invalid")
        #expect(image == "The picture of that region couldn’t be read. Draw the region again, or try a smaller one.")
        #expect(invalid == "Invalid depiction value.")
        #expect(image != unknown)
        #expect(invalid != unknown)
    }
}
