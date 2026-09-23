import SwiftUI

/// Collapsed NameValue summary + Add/Edit control for any host form row (NV-5).
struct NameValueHostControl: View {
    let draft: NameValueDraft
    var accessibilityIdentifierPrefix: String = "nameValue.host"
    let onEdit: () -> Void

    private var hasForm: Bool { !draft.trimmedForm.isEmpty }

    var body: some View {
        HStack(alignment: .center, spacing: PVSpacing.space4) {
            VStack(alignment: .leading, spacing: PVSpacing.space1) {
                if hasForm {
                    Text(verbatim: draft.trimmedForm)
                        .font(PVFont.body(size: PVTypeScale.body))
                        .foregroundStyle(PVColor.textPrimary)
                    if !draft.storedPartsLine.isEmpty {
                        Text(verbatim: draft.storedPartsLine)
                            .font(PVFont.mono(size: PVTypeScale.micro))
                            .foregroundStyle(PVColor.textMuted)
                    }
                } else {
                    Text(L10n.NameValue.summaryNotNormalized)
                        .font(PVFont.body(size: PVTypeScale.body, italic: true))
                        .foregroundStyle(PVColor.textMuted)
                }
            }
            .frame(maxWidth: .infinity, alignment: .leading)
            PVButton(
                hasForm ? L10n.NameValue.summaryEdit : L10n.NameValue.summaryAdd,
                variant: hasForm ? .secondary : .ghost,
                size: .sm,
                action: onEdit
            )
            .accessibilityIdentifier("\(accessibilityIdentifierPrefix).action")
        }
        .padding(PVSpacing.space5)
        .frame(maxWidth: .infinity, alignment: .leading)
        .background(
            RoundedRectangle(cornerRadius: PVRadius.sm, style: .continuous)
                .fill(PVColor.surfaceSunken)
        )
        .accessibilityIdentifier(accessibilityIdentifierPrefix)
    }
}
