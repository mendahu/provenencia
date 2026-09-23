import SwiftUI

/// Shared chrome for single-line inputs (focus ring, inset rest shadow).
/// Focus only changes colors/opacities on layers that are always present —
/// never branch the view tree on focus (see `DesignSystem/README.md`
/// § "Interaction state").
struct PVInputChrome: ViewModifier {
    /// Horizontal inset when there is no leading icon (matches trailing padding).
    static let horizontalInset: CGFloat = 10
    /// Leading overlay glyph size — must match `PVInput`'s icon overlay.
    static let iconSize: CGFloat = 14
    /// Gap between the leading icon and the text.
    static let iconTextGap: CGFloat = 7

    var size: PVControlSize = .md
    var mono: Bool = false
    var isFocused: Bool = false
    var isInvalid: Bool = false
    /// Reserves room on the left for `PVInput`'s icon overlay.
    var leadingIconInset: Bool = false
    /// Reserves room on the right for an optional suffix overlay (e.g. ⌘K).
    var trailingSuffixWidth: CGFloat = 0

    private var leadingPadding: CGFloat {
        if leadingIconInset {
            return Self.horizontalInset + Self.iconSize + Self.iconTextGap
        }
        return Self.horizontalInset
    }

    private var trailingPadding: CGFloat {
        if trailingSuffixWidth > 0 {
            return Self.horizontalInset + trailingSuffixWidth + Self.iconTextGap
        }
        return Self.horizontalInset
    }

    func body(content: Content) -> some View {
        content
            .font(mono ? PVFont.mono(size: size == .sm ? PVTypeScale.caption : PVTypeScale.bodySmall) : size.font)
            .foregroundStyle(PVColor.textPrimary)
            .padding(.leading, leadingPadding)
            .padding(.trailing, trailingPadding)
            .frame(height: size.height)
            .background(
                RoundedRectangle(cornerRadius: PVRadius.sm, style: .continuous)
                    .fill(PVColor.surfaceRaised)
            )
            .overlay(border)
            .pvInsetShadow(cornerRadius: PVRadius.sm, visible: !isFocused)
            .pvFocusRing(isFocused, cornerRadius: PVRadius.sm)
    }

    private var border: some View {
        RoundedRectangle(cornerRadius: PVRadius.sm, style: .continuous)
            .stroke(borderColor, lineWidth: 1)
            .allowsHitTesting(false)
            .pvAnimation(PVMotion.fastStandard, value: isFocused)
            .pvAnimation(PVMotion.fastStandard, value: isInvalid)
    }

    private var borderColor: Color {
        if isInvalid { return PVColor.danger }
        return isFocused ? PVColor.borderFocus : PVColor.borderDefault
    }
}

/// Convenience field with optional leading icon, trailing suffix, read-only
/// rendering, and owned or externally driven focus state. `activateOnAppear`
/// focuses the field on the next turn so an inserting action (and its Space
/// key) can finish before the caret moves.
struct PVInput: View {
    @Binding private var text: String
    private let size: PVControlSize
    private let mono: Bool
    private let isReadOnly: Bool
    private let prompt: LocalizedStringResource?
    private let icon: PVSymbol?
    private let suffix: String?
    private let isInvalid: Bool
    private let activateOnAppear: Bool
    private let externalFocus: FocusState<Bool>.Binding?

    @FocusState private var ownedFocus: Bool

    init(
        text: Binding<String>,
        size: PVControlSize = .md,
        mono: Bool = false,
        isReadOnly: Bool = false,
        prompt: LocalizedStringResource? = nil,
        icon: PVSymbol? = nil,
        suffix: String? = nil,
        isInvalid: Bool = false,
        focused: FocusState<Bool>.Binding? = nil,
        activateOnAppear: Bool = false
    ) {
        self._text = text
        self.size = size
        self.mono = mono
        self.isReadOnly = isReadOnly
        self.prompt = prompt
        self.icon = icon
        self.suffix = Self.nilIfEmpty(suffix)
        self.isInvalid = isInvalid
        self.activateOnAppear = activateOnAppear
        self.externalFocus = focused
    }

    var body: some View {
        Group {
            if isReadOnly {
                Text(text)
                    .frame(maxWidth: .infinity, alignment: .leading)
                    .modifier(chrome(isFocused: false))
            } else if let externalFocus {
                TextField(text: $text, prompt: prompt.map { Text($0) }) {
                    EmptyView()
                }
                    .textFieldStyle(.plain)
                    .focused(externalFocus)
                    .modifier(chrome(isFocused: externalFocus.wrappedValue))
            } else {
                TextField(text: $text, prompt: prompt.map { Text($0) }) {
                    EmptyView()
                }
                    .textFieldStyle(.plain)
                    .focused($ownedFocus)
                    .modifier(chrome(isFocused: ownedFocus))
            }
        }
        .overlay(alignment: .leading) {
            if let icon {
                PVIcon(icon, size: PVInputChrome.iconSize)
                    .foregroundStyle(PVColor.textFaint)
                    .padding(.leading, PVInputChrome.horizontalInset)
                    .allowsHitTesting(false)
            }
        }
        .overlay(alignment: .trailing) {
            if let suffix {
                Text(suffix)
                    .font(PVFont.mono(size: PVTypeScale.micro))
                    .foregroundStyle(PVColor.textSecondary)
                    .padding(.trailing, PVInputChrome.horizontalInset)
                    .allowsHitTesting(false)
                    .accessibilityHidden(true)
            }
        }
        .onAppear {
            guard activateOnAppear, !isReadOnly else { return }
            // Same-cycle focus is dropped while the field is inserted; the
            // next turn also lets a Space-activated control finish so the
            // key is not typed into the new value.
            DispatchQueue.main.async {
                if let externalFocus {
                    externalFocus.wrappedValue = true
                } else {
                    ownedFocus = true
                }
            }
        }
    }

    private func chrome(isFocused: Bool) -> PVInputChrome {
        PVInputChrome(
            size: size,
            mono: mono,
            isFocused: isFocused,
            isInvalid: isInvalid,
            leadingIconInset: icon != nil,
            trailingSuffixWidth: suffix.map { Self.estimatedSuffixWidth($0) } ?? 0
        )
    }

    private static func nilIfEmpty(_ value: String?) -> String? {
        guard let value else { return nil }
        let trimmed = value.trimmingCharacters(in: .whitespacesAndNewlines)
        return trimmed.isEmpty ? nil : trimmed
    }

    private static func estimatedSuffixWidth(_ suffix: String) -> CGFloat {
        // Mono micro shortcut badges (⌘K, ⌘[) are short; reserve a stable inset.
        max(22, CGFloat(suffix.count) * 7)
    }
}

#Preview {
    VStack(spacing: PVSpacing.space5) {
        PVInput(text: .constant(""), prompt: "Search fields")
        PVInput(text: .constant(""), prompt: "Search fields", icon: .search)
        PVInput(text: .constant(""), size: .sm, prompt: "Search everything", icon: .search, suffix: "⌘K")
        PVInput(text: .constant("USR-A1B2C"), mono: true)
        PVInput(text: .constant("Jane Smith"), isReadOnly: true)
        PVInput(text: .constant(""), prompt: "Required", isInvalid: true)
    }
    .padding(PVSpacing.space9)
    .frame(width: 320)
    .background(PVColor.surfacePage)
}
