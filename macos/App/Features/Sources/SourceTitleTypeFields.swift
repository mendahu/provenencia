import SwiftUI

/// The Title and Source type fields shared by Add Source and Copy Source. It
/// owns the one display rule: nothing shows on open; after the first edit by
/// the researcher to either field, both errors follow `draft.validation`. The
/// app's own changes to the draft (a sole type filled in, the Copy prefill)
/// change no error.
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
                    text: Binding(
                        get: { draft.title },
                        set: { new in
                            guard new != draft.title else { return }
                            draft.title = new
                            showErrors()
                        }
                    ),
                    isInvalid: titleError != nil,
                    activateOnAppear: true
                )
                .accessibilityIdentifier("\(accessibilityIdentifierPrefix).title")
            }
            PVField(
                label: L10n.Sources.formType,
                error: typeError,
                required: true
            ) {
                PVComboBox(
                    selection: Binding(
                        get: { draft.sourceTypeID },
                        set: { new in
                            guard new != draft.sourceTypeID else { return }
                            draft.sourceTypeID = new
                            showErrors()
                        }
                    ),
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
            }
        }
    }

    private func showErrors() {
        titleError = draft.validation.titleError
        typeError = draft.validation.typeError
    }
}
