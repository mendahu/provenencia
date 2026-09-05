import SwiftUI

/// Sizes a `PVInput`/`PVSelect` control. Shared so a `PVField` can mix the
/// two without their heights disagreeing.
enum PVControlSize {
    case sm, md, lg

    var height: CGFloat {
        switch self {
        case .sm: return PVSpacing.controlHeightSmall
        case .md: return PVSpacing.controlHeightMedium
        case .lg: return PVSpacing.controlHeightLarge
        }
    }

    var font: Font {
        self == .sm ? PVFont.body(size: PVTypeScale.caption) : PVFont.body(size: PVTypeScale.bodySmall)
    }
}

/// A styled single-line text field — mirrors `components/forms/Input.jsx`
/// (focus ring, inset shadow at rest, optional `mono` face for anything a
/// researcher would cite exactly). `isReadOnly` renders static text instead
/// of a `TextField`, matching the read-only fields the Onboarding Flow
/// board shows (e.g. a name carried over from a prior screen) — SwiftUI
/// has no direct "read-only but still looks editable" `TextField` mode.
///
/// See `DesignSystem/README.md` / `PVButton.swift` for the component
/// pattern this follows.
struct PVInput: View {
    @Binding private var text: String
    private let size: PVControlSize
    private let mono: Bool
    private let isReadOnly: Bool

    init(
        text: Binding<String>,
        size: PVControlSize = .md,
        mono: Bool = false,
        isReadOnly: Bool = false
    ) {
        self._text = text
        self.size = size
        self.mono = mono
        self.isReadOnly = isReadOnly
    }

    @FocusState private var isFocused: Bool

    var body: some View {
        Group {
            if isReadOnly {
                Text(text)
                    .frame(maxWidth: .infinity, alignment: .leading)
            } else {
                TextField("", text: $text)
                    .textFieldStyle(.plain)
                    .focused($isFocused)
            }
        }
        .font(mono ? PVFont.mono(size: size == .sm ? PVTypeScale.caption : PVTypeScale.bodySmall) : size.font)
        .foregroundStyle(PVColor.textPrimary)
        .padding(.horizontal, PVSpacing.space3 + PVSpacing.space1)
        .frame(height: size.height)
        .background(
            RoundedRectangle(cornerRadius: PVRadius.sm, style: .continuous)
                .fill(PVColor.surfaceRaised)
        )
        .overlay(
            RoundedRectangle(cornerRadius: PVRadius.sm, style: .continuous)
                .stroke(isFocused ? PVColor.borderFocus : PVColor.borderDefault, lineWidth: 1)
        )
        .modifier(PVInputRestingShadow(isFocused: isFocused))
        .animation(PVMotion.fastStandard, value: isFocused)
    }
}

/// Applies the focus ring when focused, or the resting inset shadow
/// otherwise — split out so `PVInput`'s `body` reads top-to-bottom like
/// the web `Input.jsx` styles object.
private struct PVInputRestingShadow: ViewModifier {
    let isFocused: Bool

    func body(content: Content) -> some View {
        if isFocused {
            content.pvFocusRing(cornerRadius: PVRadius.sm)
        } else {
            content.pvInsetShadow(cornerRadius: PVRadius.sm)
        }
    }
}

#Preview {
    VStack(spacing: PVSpacing.space5) {
        PVInput(text: .constant(""))
        PVInput(text: .constant("USR-A1B2C"), mono: true)
        PVInput(text: .constant("Jane Smith"), isReadOnly: true)
    }
    .padding(PVSpacing.space9)
    .frame(width: 320)
    .background(PVColor.surfacePage)
}
