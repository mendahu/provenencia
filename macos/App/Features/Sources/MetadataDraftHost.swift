import Foundation
import Observation

/// A board over rows no write has saved: a form's copy details, sent with its
/// owner's one write as `draft`. Its writes change only its own rows.
@MainActor
@Observable
final class MetadataDraftHost: MetadataBoardHost {
    var entries: [CatalogMetadataEntry]
    @ObservationIgnored let fields: [CatalogMetadataField]
    /// The rows the host opened with: what `isChanged` compares against.
    @ObservationIgnored private let opened: [CatalogMetadataEntry]
    /// Suggestions the researcher dismissed, in order; each is sent as a dismissed layout row.
    private var dismissed: [String] = []

    @ObservationIgnored var confirmsClear: Bool { false }

    init(entries: [CatalogMetadataEntry], fields: [CatalogMetadataField]) {
        self.entries = entries
        self.fields = fields
        self.opened = entries
    }

    /// The starting rows of a board for an owner that holds no values and no layout yet:
    /// each of the type's suggestions, in its order, empty.
    static func unsavedRows(from suggestions: [CatalogTypeSuggestion]) -> [CatalogMetadataEntry] {
        suggestions.map {
            CatalogMetadataEntry(
                field: $0.field, valueText: "", hasValue: false, suggested: true,
                sortOrder: Int32($0.sortOrder)
            )
        }
    }

    // MARK: What the board shows against what it opened with

    /// True when what the board shows differs from what it opened with: a value, the order of the
    /// saved rows, the empty suggestions, or a dismissal. A value typed and put back, or a drag put
    /// back, is unchanged.
    var isChanged: Bool {
        Self.shown(entries) != Self.shown(opened) || !dismissed.isEmpty
    }

    /// What the owner's Save sends, or nil when nothing changed (a never-touched owner keeps no layout).
    /// `values` are the rows with a value, in row order; `layout` lists every row in row order and then
    /// each dismissed suggestion, because `Apply` appends no layout row for a new value.
    var draft: CatalogMetadataDraft? {
        guard isChanged else { return nil }
        var draft = CatalogMetadataDraft()
        draft.values = entries.filter(\.hasValue).map {
            CatalogMetadataDraft.Value(fieldID: $0.field.id, valueText: $0.valueText)
        }
        draft.layout = entries.enumerated().map { index, row in
            CatalogMetadataDraft.LayoutRow(fieldID: row.field.id, sortOrder: Int32(index), dismissed: false)
        }
        for (k, id) in dismissed.enumerated() {
            draft.layout.append(
                CatalogMetadataDraft.LayoutRow(
                    fieldID: id, sortOrder: Int32(entries.count + k), dismissed: true
                )
            )
        }
        return draft
    }

    private struct Shown: Equatable {
        var saved: [Saved]
        var suggestions: [String]
        struct Saved: Equatable {
            var fieldID: String
            var valueText: String
        }
    }

    /// The board draws saved rows then empty suggestions; `moveSaved` rebuilds the rows that way, so the
    /// two groups are compared apart.
    private static func shown(_ rows: [CatalogMetadataEntry]) -> Shown {
        Shown(
            saved: rows.filter(\.hasValue).map { Shown.Saved(fieldID: $0.field.id, valueText: $0.valueText) },
            suggestions: rows.filter { !$0.hasValue && $0.suggested }.map(\.field.id)
        )
    }

    // MARK: MetadataBoardHost

    func set(fieldID: String, valueText: String) -> ([CatalogMetadataEntry], CatalogWriteEffects) {
        if let index = entries.firstIndex(where: { $0.field.id == fieldID }) {
            let held = opened.first { $0.field.id == fieldID }
            entries[index].valueText = valueText
            entries[index].hasValue = true
            entries[index].openHref = (held?.hasValue == true && held?.valueText == valueText) ? held?.openHref ?? "" : ""
        } else if let field = fields.first(where: { $0.id == fieldID }) {
            entries.append(
                CatalogMetadataEntry(
                    field: field, valueText: valueText, hasValue: true, suggested: false,
                    sortOrder: Int32(entries.count)
                )
            )
        } else {
            return (entries, .empty)
        }
        dismissed.removeAll { $0 == fieldID }
        return (entries, .empty)
    }

    func clear(fieldID: String) -> ([CatalogMetadataEntry], CatalogWriteEffects) {
        guard let index = entries.firstIndex(where: { $0.field.id == fieldID }) else { return (entries, .empty) }
        if entries[index].suggested {
            entries[index].valueText = ""
            entries[index].hasValue = false
            entries[index].openHref = ""
        } else {
            entries.remove(at: index)
        }
        return (entries, .empty)
    }

    func dismissSuggestion(fieldID: String) -> ([CatalogMetadataEntry], CatalogWriteEffects) {
        guard let index = entries.firstIndex(where: { $0.field.id == fieldID }),
              entries[index].suggested, !entries[index].hasValue
        else { return (entries, .empty) }
        entries.remove(at: index)
        dismissed.append(fieldID)
        return (entries, .empty)
    }

    func reorder(fieldIDs: [String]) -> ([CatalogMetadataEntry], CatalogWriteEffects) {
        entries = entries.reordered(by: fieldIDs)
        return (entries, .empty)
    }

    func didWrite(_: CatalogWriteEffects) {}
    func clearFailure() {}
    func failureMessage(for error: Error) -> String? { L10n.Errors.message(for: error) }
    func showFailure(_: String) {}
}
