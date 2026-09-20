import SwiftUI

/// Shared chrome for multiline text areas (focus ring, inset rest shadow).
/// Mirrors `PVInputChrome` without a fixed control height so the field can
/// grow with `lineLimit`. Focus only retints always-present layers — never
/// branch the view tree on focus (see `DesignSystem/README.md`
/// § "Interaction state").
struct PVTextAreaChrome: ViewModifier {
    var typography: PVTextArea.Typography = .body
    var isFocused: Bool = false
    var isInvalid: Bool = false

    func body(content: Content) -> some View {
        content
            .font(typography.font)
            .foregroundStyle(typography.foreground)
            .padding(.horizontal, PVInputChrome.horizontalInset)
            .padding(.vertical, typography.verticalPadding)
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

/// Multiline sibling of `PVInput`. Replaces hand-rolled styled `TextField`s
/// (title / description / metadata / artifact description / notes).
///
/// Not a direct port of a web-kit component — extracted from Provenencia
/// Source-page chrome during structural cleanup. Call sites own
/// `.accessibilityIdentifier` (`docs/macos-client-patterns.md` §5).
struct PVTextArea: View {
    /// Type ramp for the field contents.
    enum Typography {
        /// Default body-small (artifact description, note composer / editor).
        case body
        /// Full body size (Source description).
        case bodyLarge
        /// Mono caption (metadata values).
        case mono
        /// Display h1 (Source title).
        case display

        var font: Font {
            switch self {
            case .body:
                return PVFont.body(size: PVTypeScale.bodySmall, weight: PVFontWeight.regular)
            case .bodyLarge:
                return PVFont.body(size: PVTypeScale.body, weight: PVFontWeight.regular)
            case .mono:
                return PVFont.mono(size: PVTypeScale.caption)
            case .display:
                return PVFont.display(size: PVTypeScale.h1, weight: PVFontWeight.semibold)
            }
        }

        var foreground: Color {
            switch self {
            case .display: return PVColor.textDisplay
            case .body, .bodyLarge, .mono: return PVColor.textPrimary
            }
        }

        var verticalPadding: CGFloat {
            switch self {
            case .mono: return PVSpacing.space3
            case .body, .bodyLarge, .display: return PVSpacing.space4
            }
        }
    }

    @Binding private var text: String
    private let lineLimit: ClosedRange<Int>
    private let prompt: LocalizedStringResource?
    private let typography: Typography
    private let isInvalid: Bool
    /// When true, claims focus once the field appears (inline edit / type picker).
    private let activateOnAppear: Bool

    init(
        text: Binding<String>,
        lineLimit: ClosedRange<Int> = 2...8,
        prompt: LocalizedStringResource? = nil,
        typography: Typography = .body,
        isInvalid: Bool = false,
        activateOnAppear: Bool = false
    ) {
        self._text = text
        self.lineLimit = lineLimit
        self.prompt = prompt
        self.typography = typography
        self.isInvalid = isInvalid
        self.activateOnAppear = activateOnAppear
    }

    @FocusState private var isFocused: Bool

    var body: some View {
        TextField(
            text: $text,
            prompt: prompt.map { Text($0) },
            axis: .vertical
        ) {
            EmptyView()
        }
        .textFieldStyle(.plain)
        .lineLimit(lineLimit)
        .focused($isFocused)
        .modifier(
            PVTextAreaChrome(
                typography: typography,
                isFocused: isFocused,
                isInvalid: isInvalid
            )
        )
        .onAppear {
            guard activateOnAppear else { return }
            // Defer past the resting→editing view swap; same-cycle focus is dropped.
            DispatchQueue.main.async {
                isFocused = true
            }
        }
    }
}

#Preview {
    VStack(alignment: .leading, spacing: PVSpacing.space5) {
        PVTextArea(
            text: .constant("Family album, Norfolk"),
            lineLimit: 1...4,
            prompt: "Title",
            typography: .display
        )
        PVTextArea(
            text: .constant("Held by Mary; consulted at the NRO."),
            lineLimit: 3...8,
            prompt: "Description",
            typography: .bodyLarge
        )
        PVTextArea(
            text: .constant("Mary Robins"),
            lineLimit: 1...4,
            typography: .mono
        )
        PVTextArea(
            text: .constant(""),
            lineLimit: 2...6,
            prompt: "Add a note",
            isInvalid: true
        )
    }
    .padding(PVSpacing.space9)
    .frame(width: 420)
    .background(PVColor.surfacePage)
}
