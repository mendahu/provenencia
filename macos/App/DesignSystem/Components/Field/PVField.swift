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
    /// What the label wears beside it: nothing, the required star, or a muted "(optional)".
    enum Marker {
        case none
        case required
        case optional
    }

    private let label: LocalizedStringResource?
    private let hint: Text?
    /// Plain `String`, not `LocalizedStringResource` — unlike `label`/`hint`,
    /// an error is always dynamic content (a validation message, a mapped
    /// FFI error code), already resolved via `String(localized:)` or
    /// `L10n.Errors.message` before it reaches here. Same reasoning as
    /// `PVToast`'s `title`/`message`.
    private let error: String?
    private let marker: Marker
    private let content: Content

    init(
        label: LocalizedStringResource? = nil,
        hint: (any PVCopy)? = nil,
        error: String? = nil,
        marker: Marker = .none,
        @ViewBuilder content: () -> Content
    ) {
        self.label = label
        self.hint = hint?.pvText
        self.error = error
        self.marker = marker
        self.content = content()
    }

    var body: some View {
        VStack(alignment: .leading, spacing: PVSpacing.space3) {
            if let label {
                HStack(spacing: PVSpacing.space1) {
                    Text(label)
                    switch marker {
                    case .none:
                        EmptyView()
                    case .required:
                        Text(L10n.DesignSystem.requiredMarker).foregroundStyle(PVColor.danger)
                    case .optional:
                        Text(L10n.DesignSystem.optionalMarker).foregroundStyle(PVColor.textMuted)
                    }
                }
                .font(PVFont.body(size: PVTypeScale.caption, weight: PVFontWeight.medium))
                .foregroundStyle(PVColor.textSecondary)
            }

            content

            if let error {
                PVErrorText(error)
            } else if let hint {
                hint
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

#Preview("Optional marker") {
    PVField(label: "Additional comments", marker: .optional) {
        PVInput(text: .constant(""))
    }
    .padding(PVSpacing.space9)
    .frame(width: 320)
    .background(PVColor.surfacePage)
}
