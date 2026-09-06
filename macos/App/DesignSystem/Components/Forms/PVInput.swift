import SwiftUI

/// Shared chrome for single-line inputs (focus ring, inset rest shadow).
struct PVInputChrome: ViewModifier {
    var size: PVControlSize = .md
    var mono: Bool = false
    var isFocused: Bool = false

    func body(content: Content) -> some View {
        content
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
            .pvAnimation(PVMotion.fastStandard, value: isFocused)
    }
}

/// TextField style that applies `PVInputChrome`. Pass `isFocused` from a
/// `@FocusState` at the call site when you need the focus ring.
struct PVTextFieldStyle: TextFieldStyle {
    var size: PVControlSize = .md
    var mono: Bool = false
    var isFocused: Bool = false

    func _body(configuration: TextField<Self._Label>) -> some View {
        configuration
            .modifier(PVInputChrome(size: size, mono: mono, isFocused: isFocused))
    }
}

extension TextFieldStyle where Self == PVTextFieldStyle {
    static func pv(size: PVControlSize = .md, mono: Bool = false, isFocused: Bool = false) -> PVTextFieldStyle {
        PVTextFieldStyle(size: size, mono: mono, isFocused: isFocused)
    }
}

/// Convenience field with optional read-only rendering and owned focus state.
struct PVInput: View {
    @Binding private var text: String
    private let size: PVControlSize
    private let mono: Bool
    private let isReadOnly: Bool
    private let prompt: LocalizedStringResource?

    init(
        text: Binding<String>,
        size: PVControlSize = .md,
        mono: Bool = false,
        isReadOnly: Bool = false,
        prompt: LocalizedStringResource? = nil
    ) {
        self._text = text
        self.size = size
        self.mono = mono
        self.isReadOnly = isReadOnly
        self.prompt = prompt
    }

    @FocusState private var isFocused: Bool

    var body: some View {
        Group {
            if isReadOnly {
                Text(text)
                    .frame(maxWidth: .infinity, alignment: .leading)
                    .modifier(PVInputChrome(size: size, mono: mono, isFocused: false))
            } else if let prompt {
                TextField("", text: $text, prompt: Text(prompt))
                    .textFieldStyle(.pv(size: size, mono: mono, isFocused: isFocused))
                    .focused($isFocused)
            } else {
                TextField("", text: $text)
                    .textFieldStyle(.pv(size: size, mono: mono, isFocused: isFocused))
                    .focused($isFocused)
            }
        }
    }
}

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
