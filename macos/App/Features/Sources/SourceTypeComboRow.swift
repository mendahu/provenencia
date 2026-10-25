import SwiftUI

/// One look for a Source type in a picker: its mark, then the combo's plain row.
struct SourceTypeComboRow: View {
    /// The type behind `option`, nil when the pool has not loaded it.
    let type: CatalogSourceType?
    let option: PVComboBoxOption
    let query: String

    var body: some View {
        HStack(spacing: PVSpacing.space3) {
            if let type {
                PVMark(
                    PVMarkKey(catalogKey: type.iconKey),
                    size: .row,
                    decorative: true
                )
            }
            PVComboBoxPlainRow(option: option, query: query)
        }
    }

    /// The combo options for a type pool: id as value, label as text, key as subtext.
    static func options(for types: [CatalogSourceType]) -> [PVComboBoxOption] {
        types.map { PVComboBoxOption(value: $0.id, label: $0.label, subtext: $0.key) }
    }
}
