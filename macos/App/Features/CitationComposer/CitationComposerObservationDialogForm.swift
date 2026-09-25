import SwiftUI

/// Name / date editor body for `.pvFormDialog` (inline types stay on the row).
struct CitationComposerObservationDialogForm: View {
    @Bindable var model: CitationComposerModel

    var body: some View {
        if let draft = model.observationDialog {
            let property = model.catalogProperty(id: draft.propertyID)
            PVField(error: model.dialogValueError) {
                VStack(alignment: .leading, spacing: PVSpacing.space6) {
                    if property?.valueType == PropertyValueType.name.rawValue {
                        NameValueEditorForm(
                            draft: nameBinding,
                            accessibilityIdentifierPrefix: "citationComposer.dialog.name"
                        )
                    } else if property?.valueType == PropertyValueType.date.rawValue {
                        DateValueEditorForm(
                            draft: dateBinding,
                            accessibilityIdentifierPrefix: "citationComposer.dialog.date"
                        )
                    }
                }
            }
        }
    }

    private var dateBinding: Binding<DateValueDraft> {
        Binding(
            get: { model.observationDialog?.dateDraft ?? .empty() },
            set: {
                guard var draft = model.observationDialog else { return }
                draft.dateDraft = $0
                model.updateObservationDialog(draft)
            }
        )
    }

    private var nameBinding: Binding<NameValueDraft> {
        Binding(
            get: { model.observationDialog?.nameDraft ?? .empty() },
            set: {
                guard var draft = model.observationDialog else { return }
                draft.nameDraft = $0
                model.updateObservationDialog(draft)
            }
        )
    }
}
