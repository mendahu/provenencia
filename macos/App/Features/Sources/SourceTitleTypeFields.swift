import SwiftUI

/// The Title and Source type fields shared by Add Source and Copy Source. It
/// owns the one display rule: a field's error follows `draft.validation` once
/// that field is edited, so errors show after an edit, not on open.
struct SourceTitleTypeFields: View {
    @Binding var draft: SourcesModel.Draft
    @Binding var titleError: String?
    @Binding var typeError: String?
    let types: [CatalogSourceType]
    let accessibilityIdentifierPrefix: String

    var body: some View {
        let options = SourceTypeComboRow.options(for: types)
        VStack(alignment: .leading, spacing: PVSpacing.space6) {
            PVField(
                label: L10n.Sources.formTitle,
                error: titleError,
                required: true
            ) {
                PVInput(
                    text: $draft.title,
                    isInvalid: titleError != nil,
                    activateOnAppear: true
                )
                .accessibilityIdentifier("\(accessibilityIdentifierPrefix).title")
                .onChange(of: draft.title) { _, _ in
                    titleError = draft.validation.titleError
                }
            }
            PVField(
                label: L10n.Sources.formType,
                error: typeError,
                required: true
            ) {
                PVComboBox(
                    selection: $draft.sourceTypeID,
                    options: options,
                    placeholder: L10n.Sources.typePlaceholder,
                    emptyLabel: options.isEmpty
                        ? L10n.Sources.typePoolEmpty
                        : L10n.Sources.typeNoMatch,
                    isInvalid: typeError != nil,
                    label: L10n.Sources.formType,
                    accessibilityIdentifierPrefix: "\(accessibilityIdentifierPrefix).type"
                ) { option, query in
                    SourceTypeComboRow(
                        type: types.first { $0.id == option.value },
                        option: option,
                        query: query
                    )
                }
                .onChange(of: draft.sourceTypeID) { _, _ in
                    typeError = draft.validation.typeError
                }
            }
        }
    }
}
