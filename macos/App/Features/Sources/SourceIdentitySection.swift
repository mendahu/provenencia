import Foundation
import Observation

/// Title, type, and description edit flows for the Source page. Each field
/// carries its own draft, edit flag, and error channel; all three share one
/// `updateSource` write.
@MainActor
@Observable
final class SourceIdentitySection {
    var editingTitle = false
    var titleDraft = ""
    var titleError: String?

    var editingType = false
    var typeDraftID = ""
    var typeError: String?

    var editingDescription = false
    var descriptionDraft = ""
    var descriptionError: String?

    private(set) var isSaving = false

    private let context: SourcePageContext

    init(context: SourcePageContext) {
        self.context = context
    }

    // MARK: Saved values (derived from the workspace)

    var title: String { context.workspace?.source.title ?? "" }

    var description: String { context.workspace?.source.description ?? "" }

    var sourceTypeID: String { context.workspace?.source.sourceTypeID ?? "" }

    var types: [CatalogSourceType] { context.workspace?.types ?? [] }

    /// Resting type label from committed `sourceTypeID`.
    var typeLabel: String {
        types.first { $0.id == sourceTypeID }?.label ?? ""
    }

    var typeComboOptions: [PVComboBoxOption] {
        types.map { PVComboBoxOption(value: $0.id, label: $0.label, subtext: $0.key) }
    }

    /// Discards all edit sessions after a workspace (re)load.
    func reset() {
        editingTitle = false
        titleDraft = ""
        titleError = nil
        editingType = false
        typeDraftID = sourceTypeID
        typeError = nil
        editingDescription = false
        descriptionDraft = ""
        descriptionError = nil
    }

    // MARK: Title

    func beginEditTitle() {
        titleDraft = title
        titleError = nil
        editingTitle = true
    }

    func cancelEditTitle() {
        guard !isSaving else { return }
        editingTitle = false
        titleDraft = ""
        titleError = nil
    }

    func saveTitle() async {
        guard !isSaving else { return }
        let trimmed = titleDraft.trimmingCharacters(in: .whitespacesAndNewlines)
        if trimmed.isEmpty {
            titleError = String(localized: L10n.Sources.pageTitleRequired)
            return
        }
        titleError = await save(
            title: trimmed,
            description: description.trimmingCharacters(in: .whitespacesAndNewlines),
            sourceTypeID: sourceTypeID
        )
        if titleError == nil {
            editingTitle = false
            titleDraft = ""
        }
    }

    // MARK: Type

    func beginEditType() {
        // Blank the picker so the researcher can type immediately; Cancel
        // restores `sourceTypeID`.
        typeDraftID = ""
        typeError = nil
        editingType = true
    }

    func cancelEditType() {
        guard !isSaving else { return }
        typeDraftID = sourceTypeID
        typeError = nil
        editingType = false
    }

    func saveType() async {
        guard !isSaving else { return }
        let typeID = typeDraftID.isEmpty ? sourceTypeID : typeDraftID
        typeError = await save(
            title: title.trimmingCharacters(in: .whitespacesAndNewlines),
            description: description.trimmingCharacters(in: .whitespacesAndNewlines),
            sourceTypeID: typeID
        )
        if typeError == nil {
            editingType = false
        }
    }

    // MARK: Description

    func beginEditDescription() {
        descriptionDraft = description
        descriptionError = nil
        editingDescription = true
    }

    func cancelEditDescription() {
        guard !isSaving else { return }
        editingDescription = false
        descriptionDraft = ""
        descriptionError = nil
    }

    func saveDescription() async {
        guard !isSaving else { return }
        descriptionError = await save(
            title: title.trimmingCharacters(in: .whitespacesAndNewlines),
            description: descriptionDraft.trimmingCharacters(in: .whitespacesAndNewlines),
            sourceTypeID: sourceTypeID
        )
        if descriptionError == nil {
            editingDescription = false
            descriptionDraft = ""
        }
    }

    /// Shared `updateSource` write used by title / type / description saves.
    /// Returns a localized error message on failure, nil on success (or no-op).
    private func save(title: String, description: String, sourceTypeID: String) async -> String? {
        if title.isEmpty {
            return String(localized: L10n.Sources.pageTitleRequired)
        }
        guard let current = context.workspace?.source else { return nil }
        if title == current.title,
           description == current.description,
           sourceTypeID == current.sourceTypeID
        {
            return nil
        }
        isSaving = true
        defer { isSaving = false }
        do {
            let updated = try await context.store.updateSource(
                projectDir: context.projectDir,
                userID: context.userID,
                sourceID: context.sourceID,
                sourceTypeID: sourceTypeID,
                title: title,
                description: description
            )
            context.workspace?.source = updated
            typeDraftID = updated.sourceTypeID
            context.applySource(updated)
            return nil
        } catch {
            return L10n.Errors.message(for: error)
        }
    }
}
