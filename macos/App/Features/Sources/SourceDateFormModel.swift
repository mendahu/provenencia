import Foundation
import Observation

/// State of the Source date dialog: the shared date draft, whether Save will
/// clear the date, and the one `setSourceDate` write. The Source page presents
/// it through `View.sourceDateDialog`.
@MainActor
@Observable
final class SourceDateFormModel {
    /// The Source's date when the dialog opened; nil for an undated Source.
    let initial: CatalogDateValueInput?
    private var storedDraft: DateValueDraft
    private let prefill: DateValueDraft
    /// Save writes no date. Set by `clear()`; typing a date after it takes it back.
    private(set) var clears = false
    private(set) var isSaving = false
    var saveError: String?

    private let context: SourcePageContext

    init(context: SourcePageContext, initial: CatalogDateValueInput?) {
        self.context = context
        self.initial = initial
        let start = initial.map(DateValueDraft.init(from:)) ?? .empty()
        prefill = start
        storedDraft = start
    }

    var draft: DateValueDraft {
        get { storedDraft }
        set {
            storedDraft = newValue
            if !newValue.isBlank { clears = false }
        }
    }

    var isDirty: Bool { clears || storedDraft != prefill }

    var canSave: Bool {
        !isSaving && (clears || (!storedDraft.isBlank && storedDraft.isValid))
    }

    /// Empties the form so that Save removes the Source's date.
    func clear() {
        storedDraft = .empty()
        clears = true
        saveError = nil
    }

    /// One write. True when the date was saved, or when the Source turned out to be deleted
    /// elsewhere (the page is leaving); false keeps the dialog open with its draft and error.
    func save() async -> Bool {
        guard canSave else { return false }
        isSaving = true
        defer { isSaving = false }
        saveError = nil
        do {
            let (updated, effects) = try await context.store.setSourceDate(
                projectDir: context.projectDir,
                userID: context.userID,
                sourceID: context.sourceID,
                date: clears ? nil : storedDraft.toInput()
            )
            context.applySource(updated, effects: effects)
            return true
        } catch {
            guard let message = context.fail(error) else { return true }
            saveError = message
            return false
        }
    }
}
