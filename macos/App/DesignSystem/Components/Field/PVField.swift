import SwiftUI

/// A label (+ optional hint/error) wrapper around arbitrary field content —
/// mirrors `components/forms/Field.jsx`. Generic over its content, the same
/// shape `OnboardingFooter` already uses in
/// `Features/Onboarding/` screens, so this isn't a new
/// pattern for the codebase, just this file's version of it.
///
/// See `DesignSystem/README.md` / `PVButton.swift` for the component
/// pattern this follows.
struct PVField<Content: View>: View {
    private let label: LocalizedStringResource?
    private let hint: LocalizedStringResource?
    /// Plain `String`, not `LocalizedStringResource` — unlike `label`/`hint`,
    /// an error is always dynamic content (a validation message, a mapped
    /// FFI error code), already resolved via `String(localized:)` or
    /// `L10n.Errors.message` before it reaches here. Same reasoning as
    /// `PVToast`'s `title`/`message`.
    private let error: String?
    private let required: Bool
    private let content: Content

    init(
        label: LocalizedStringResource? = nil,
        hint: LocalizedStringResource? = nil,
        error: String? = nil,
        required: Bool = false,
        @ViewBuilder content: () -> Content
    ) {
        self.label = label
        self.hint = hint
        self.error = error
        self.required = required
        self.content = content()
    }

    var body: some View {
        VStack(alignment: .leading, spacing: PVSpacing.space3) {
            if let label {
                HStack(spacing: PVSpacing.space1) {
                    Text(label)
                    if required {
                        Text(L10n.DesignSystem.requiredMarker).foregroundStyle(PVColor.danger)
                    }
                }
                .font(PVFont.body(size: PVTypeScale.caption, weight: PVFontWeight.medium))
                .foregroundStyle(PVColor.textSecondary)
            }

            content

            if let error {
                Text(error)
                    .font(PVFont.body(size: PVTypeScale.micro))
                    .foregroundStyle(PVColor.danger)
            } else if let hint {
                Text(hint)
                    .font(PVFont.body(size: PVTypeScale.micro, italic: true))
                    .foregroundStyle(PVColor.textMuted)
            }
        }
    }
}

#Preview {
    PVField(label: "Researcher name", hint: "Enter your name as it should appear on every fact you assert") {
        PVInput(text: .constant("Jane Smith"))
    }
    .padding(PVSpacing.space9)
    .frame(width: 320)
    .background(PVColor.surfacePage)
}
