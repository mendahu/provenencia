import Foundation
import Observation

/// Credibility assessment (grade chips + argument) for the Source page.
/// Saved values derive from the workspace; the chips edit a local draft
/// until Save.
@MainActor
@Observable
final class SourceCredibilitySection {
    var draftKey = CatalogCredibility.standard
    var argumentDraft = ""
    private(set) var isSaving = false

    private let context: SourcePageContext

    init(context: SourcePageContext) {
        self.context = context
    }

    var grades: [CatalogCredibilityGrade] { context.grades }

    /// Saved grade key; defaults to Standard when no assessment row.
    var savedKey: String {
        context.workspace?.credibility?.gradeKey ?? CatalogCredibility.standard
    }

    var savedArgument: String { context.workspace?.credibility?.argument ?? "" }

    var hasSavedAssessment: Bool { context.workspace?.credibility != nil }

    /// Grade shown in the chips (draft selection, falling back to Standard).
    var displayedGrade: CatalogCredibilityGrade? {
        grades.first { $0.key == draftKey }
            ?? grades.first { $0.key == CatalogCredibility.standard }
            ?? grades.first
    }

    var isDirty: Bool {
        draftKey != savedKey
            || argumentDraft.trimmingCharacters(in: .whitespacesAndNewlines)
            != savedArgument.trimmingCharacters(in: .whitespacesAndNewlines)
    }

    func resetDrafts() {
        draftKey = savedKey
        argumentDraft = savedArgument
    }

    func selectDraft(key: String) {
        draftKey = key
    }

    func cancel() {
        guard !isSaving else { return }
        resetDrafts()
    }

    func save() async {
        guard !isSaving else { return }
        guard let grade = grades.first(where: { $0.key == draftKey })
            ?? displayedGrade
            ?? grades.first
        else {
            return
        }
        let argument = argumentDraft.trimmingCharacters(in: .whitespacesAndNewlines)
        // Missing row + Standard + empty argument: keep display-only (no write).
        if !hasSavedAssessment, grade.key == CatalogCredibility.standard, argument.isEmpty {
            draftKey = CatalogCredibility.standard
            argumentDraft = ""
            return
        }
        isSaving = true
        defer { isSaving = false }
        context.clearPageError()
        do {
            let assessment = try await context.store.upsertSourceCredibilityAssessment(
                projectDir: context.projectDir,
                userID: context.userID,
                sourceID: context.sourceID,
                gradeID: grade.id,
                argument: argument
            )
            context.workspace?.credibility = assessment
            context.notifyWorkspaceMutated()
            resetDrafts()
        } catch {
            context.pageError = L10n.Errors.message(for: error)
        }
    }
}
