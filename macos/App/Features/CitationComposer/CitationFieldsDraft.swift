import Foundation
import Observation

/// Citation-row fields and the citation-only Save command.
@MainActor
@Observable
final class CitationFieldsDraft {
    struct Values: Equatable {
        var locator: CitationLocatorDraft = .artifactOnly()
        var transcription: String = ""
        var transcriptionUncertain: Bool = false
        var transcriptionNote: String = ""
        var description: String = ""
    }

    var values = Values()
    var baseline = Values()
    var error: String?
    var isSaving = false

    var locator: CitationLocatorDraft {
        get { values.locator }
        set { values.locator = newValue }
    }

    var transcription: String {
        get { values.transcription }
        set { values.transcription = newValue }
    }

    var transcriptionUncertain: Bool {
        get { values.transcriptionUncertain }
        set { values.transcriptionUncertain = newValue }
    }

    var transcriptionNote: String {
        get { values.transcriptionNote }
        set { values.transcriptionNote = newValue }
    }

    var description: String {
        get { values.description }
        set { values.description = newValue }
    }

    var isDirty: Bool { values != baseline }

    var statusText: String {
        if citationID == nil {
            return String(localized: L10n.CitationComposer.citationStatusNew)
        }
        if isDirty {
            return String(localized: L10n.CitationComposer.citationStatusDirty)
        }
        return String(localized: L10n.CitationComposer.citationStatusSaved)
    }

    private weak var owner: CitationComposerModel?

    var citationID: String? { owner?.activeCitationID }

    func attach(_ owner: CitationComposerModel) {
        self.owner = owner
    }

    func replace(from citation: CatalogCitation) {
        values = Values(
            locator: CitationLocatorDraft.decode(citation.locatorJSON),
            transcription: citation.transcription,
            transcriptionUncertain: citation.transcriptionUncertain,
            transcriptionNote: citation.transcriptionNote,
            description: citation.description
        )
        baseline = values
        error = nil
    }

    func resetBlank(keepLocator: Bool = false) {
        let locator = keepLocator ? values.locator : CitationLocatorDraft.artifactOnly()
        values = Values(locator: locator)
        baseline = values
        error = nil
    }

    func captureBaseline() {
        baseline = values
        error = nil
    }

    func saveCitation() async {
        guard !isSaving, let owner else { return }
        isSaving = true
        defer { isSaving = false }
        error = nil
        await owner.performSaveCitation()
    }
}
