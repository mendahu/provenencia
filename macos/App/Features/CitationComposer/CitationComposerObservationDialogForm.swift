import SwiftUI

/// Name / date editor body for `.pvFormDialog` (inline types stay on the row).
struct CitationComposerObservationDialogForm: View {
    @Bindable var model: CitationComposerModel

    var body: some View {
        if let draft = model.observationDialog {
            let property = model.catalogProperty(id: draft.propertyID)
            VStack(alignment: .leading, spacing: PVSpacing.space6) {
                if property?.valueType == "name" {
                    NameValueEditorForm(
                        draft: nameBinding,
                        accessibilityIdentifierPrefix: "citationComposer.dialog.name"
                    )
                } else if property?.valueType == "date" {
                    DateValueEditorForm(
                        draft: dateBinding,
                        accessibilityIdentifierPrefix: "citationComposer.dialog.date"
                    )
                }
                if let error = model.dialogValueError {
                    Text(verbatim: error)
                        .font(PVFont.body(size: PVTypeScale.caption))
                        .foregroundStyle(PVColor.danger)
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
