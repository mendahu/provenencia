import Foundation

/// A restoreable workspace place: sidebar destination plus optional deep
/// location (Source page, vocabulary row, …). Persisted in navigation history.
/// See `docs/deployment-plan/spike-3/navigation-history.md`.
struct WorkspaceLocation: Codable, Equatable, Sendable {
    var section: WorkspaceSection
    var sourceId: String?
    var fieldId: String?
    var typeId: String?
    /// Denormalized jump-menu cache; ignored for navigation identity.
    var ref: String?
    /// Denormalized jump-menu cache; ignored for navigation identity.
    var title: String?

    init(
        section: WorkspaceSection,
        sourceId: String? = nil,
        fieldId: String? = nil,
        typeId: String? = nil,
        ref: String? = nil,
        title: String? = nil
    ) {
        self.section = section
        self.sourceId = Self.nilIfEmpty(sourceId)
        self.fieldId = Self.nilIfEmpty(fieldId)
        self.typeId = Self.nilIfEmpty(typeId)
        self.ref = Self.nilIfEmpty(ref)
        self.title = Self.nilIfEmpty(title)
    }

    /// Section list root (no deep id).
    static func sectionRoot(_ section: WorkspaceSection) -> WorkspaceLocation {
        WorkspaceLocation(section: section)
    }

    /// Identity used for coalesce / equality of navigation — deep ids only.
    static func == (lhs: WorkspaceLocation, rhs: WorkspaceLocation) -> Bool {
        lhs.section == rhs.section
            && lhs.sourceId == rhs.sourceId
            && lhs.fieldId == rhs.fieldId
            && lhs.typeId == rhs.typeId
    }

    private static func nilIfEmpty(_ value: String?) -> String? {
        guard let value else { return nil }
        let trimmed = value.trimmingCharacters(in: .whitespacesAndNewlines)
        return trimmed.isEmpty ? nil : trimmed
    }
}
