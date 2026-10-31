import Foundation

/// A record's citation line: "Source, Artifact, Citation, page" (S13-11). The
/// one place to replace when the standard-citation formatter of
/// `docs/ideas/citation-formats.md` lands.
enum CitationLineDisplay {
    /// The Source's title (else the Observation ref), the Artifact label, the
    /// Citation label and the page, each trimmed and dropped when empty. The
    /// page is the printed label ("p. 23") else the image number ("image 37").
    static func line(_ outcome: CatalogReconcilerOutcome) -> String {
        let pageLabel = outcome.pageLabel.trimmingCharacters(in: .whitespacesAndNewlines)
        let page: String
        if !pageLabel.isEmpty {
            page = L10n.Conclusions.citationPagePrinted(pageLabel)
        } else if outcome.artifactPage > 0 {
            page = L10n.Conclusions.citationPageImage(outcome.artifactPage)
        } else {
            page = ""
        }
        return [
            ReconciliationRecord.sourceTitle(for: outcome),
            outcome.artifactLabel,
            outcome.citationLabel,
            page,
        ]
        .map { $0.trimmingCharacters(in: .whitespacesAndNewlines) }
        .filter { !$0.isEmpty }
        .joined(separator: L10n.string(L10n.Conclusions.citationLineSeparator))
    }
}
