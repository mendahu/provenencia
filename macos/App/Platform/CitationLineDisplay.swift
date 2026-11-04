import Foundation

/// A record's citation line: "Source, Artifact, Citation, page" (S13-11). The
/// one place to replace when the standard-citation formatter of
/// `docs/deployment-plan/spike-15/citation-formats.md` lands.
enum CitationLineDisplay {
    /// A citation line in its two caption halves (S13-14 D-3).
    struct Parts: Equatable {
        /// The Source's title, else the Observation ref; not trimmed, as the old caption was not.
        var title: String
        /// `title` is the Observation ref (the Source has no title).
        var titleIsRef: Bool
        /// The Artifact label, Citation label and page, joined as in `line`; "" when none.
        var detail: String
        /// `title` then `detail`: the whole citation line.
        var line: String
    }

    /// The Source's title (else the Observation ref), the Artifact label, the
    /// Citation label and the page, each trimmed and dropped when empty. The
    /// page is the printed label ("p. 23") else the image number ("image 37").
    static func line(_ outcome: CatalogReconcilerOutcome) -> String {
        parts(outcome).line
    }

    /// The same citation line as `line(_:)`, split into the title and the rest.
    static func parts(_ outcome: CatalogReconcilerOutcome) -> Parts {
        let pageLabel = outcome.pageLabel.trimmingCharacters(in: .whitespacesAndNewlines)
        let page: String
        if !pageLabel.isEmpty {
            page = L10n.Conclusions.citationPagePrinted(pageLabel)
        } else if outcome.artifactPage > 0 {
            page = L10n.Conclusions.citationPageImage(outcome.artifactPage)
        } else {
            page = ""
        }
        let separator = L10n.string(L10n.Conclusions.citationLineSeparator)
        func joined(_ pieces: [String]) -> String {
            pieces
                .map { $0.trimmingCharacters(in: .whitespacesAndNewlines) }
                .filter { !$0.isEmpty }
                .joined(separator: separator)
        }
        let title = ReconciliationRecord.sourceTitle(for: outcome)
        return Parts(
            title: title,
            titleIsRef: outcome.sourceTitle.isEmpty,
            detail: joined([outcome.artifactLabel, outcome.citationLabel, page]),
            line: joined([title, outcome.artifactLabel, outcome.citationLabel, page])
        )
    }
}
