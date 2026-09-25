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

    private weak var context: CitationComposerContext?

    var citationID: String? { context?.citationID }

    func attach(_ context: CitationComposerContext) {
        self.context = context
        context.fields = self
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
        guard !isSaving, let context else { return }
        isSaving = true
        defer { isSaving = false }
        error = nil
        guard let artifactID = context.artifactID else {
            error = String(localized: L10n.CitationComposer.needArtifact)
            return
        }
        do {
            if let citationID = context.citationID {
                _ = try await context.store.updateCitation(
                    projectDir: context.projectDir,
                    userID: context.userID,
                    citationID: citationID,
                    locatorJSON: locator.encodeJSON(),
                    transcription: transcription,
                    description: description,
                    transcriptionUncertain: transcriptionUncertain,
                    transcriptionNote: transcriptionNote
                )
                captureBaseline()
                context.applySavedCitation()
            } else {
                let created = try await context.store.createCitationWithObservations(
                    projectDir: context.projectDir,
                    userID: context.userID,
                    artifactID: artifactID,
                    locatorJSON: locator.encodeJSON(),
                    transcription: transcription,
                    description: description,
                    transcriptionUncertain: transcriptionUncertain,
                    transcriptionNote: transcriptionNote,
                    citationNotes: [],
                    observations: []
                )
                context.citationID = created.0.id
                captureBaseline()
                context.applySavedCitation()
                await context.reloadListedCitations()
            }
        } catch {
            self.error = L10n.Errors.message(for: error)
        }
    }
}
