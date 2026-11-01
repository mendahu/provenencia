import AppKit
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
/// These are form fields, so Tab and Shift-Tab move to the next control.
/// Option-Tab still inserts a tab character. The text view underneath
/// would otherwise take Tab as a character.
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
    /// When true, the field takes the height its container offers (never less than `lineLimit`'s
    /// lower bound) instead of sizing to its text; the text scrolls inside it.
    private let fillsHeight: Bool

    init(
        text: Binding<String>,
        lineLimit: ClosedRange<Int> = 2...8,
        prompt: LocalizedStringResource? = nil,
        typography: Typography = .body,
        isInvalid: Bool = false,
        activateOnAppear: Bool = false,
        fillsHeight: Bool = false
    ) {
        self._text = text
        self.lineLimit = lineLimit
        self.prompt = prompt
        self.typography = typography
        self.isInvalid = isInvalid
        self.activateOnAppear = activateOnAppear
        self.fillsHeight = fillsHeight
    }

    @FocusState private var isFocused: Bool
    @State private var lineHeight: CGFloat = 0
    @State private var contentHeight: CGFloat = 0

    /// Field body height: at least `lineLimit`'s lower bound, growing with the
    /// text until the upper bound, then the editor scrolls.
    private var minEditorHeight: CGFloat {
        (lineHeight > 0 ? lineHeight : 18) * CGFloat(lineLimit.lowerBound)
    }

    private var editorHeight: CGFloat {
        let line = lineHeight > 0 ? lineHeight : 18
        let minHeight = line * CGFloat(lineLimit.lowerBound)
        let maxHeight = line * CGFloat(lineLimit.upperBound)
        let content = text.isEmpty ? minHeight : max(contentHeight, line)
        return min(max(content, minHeight), maxHeight)
    }

    var body: some View {
        TextEditor(text: $text)
            .scrollContentBackground(.hidden)
            .focusEffectDisabled()
            .focused($isFocused)
            .frame(
                minHeight: fillsHeight ? minEditorHeight : editorHeight,
                maxHeight: fillsHeight ? .infinity : editorHeight
            )
            .background(alignment: .topLeading) {
                if text.isEmpty, let prompt {
                    Text(prompt)
                        .font(typography.font)
                        .foregroundStyle(PVColor.textMuted)
                        .allowsHitTesting(false)
                        .accessibilityHidden(true)
                }
                PVTextAreaMeasure(text: text, font: typography.font)
                PVTextEditorInsetClearer()
            }
            .modifier(
                PVTextAreaChrome(
                    typography: typography,
                    isFocused: isFocused,
                    isInvalid: isInvalid
                )
            )
            .onPreferenceChange(PVTextAreaLineHeightKey.self) { lineHeight = $0 }
            .onPreferenceChange(PVTextAreaContentHeightKey.self) { contentHeight = $0 }
        .onAppear {
            guard activateOnAppear else { return }
            // Defer past the resting→editing view swap; same-cycle focus is dropped.
            DispatchQueue.main.async {
                isFocused = true
            }
        }
        .pvFieldNamed()
    }
}

/// Measures one line and the full string in the editor's font so the field
/// can grow up to `lineLimit` without leaving every area at its max height.
private struct PVTextAreaMeasure: View {
    var text: String
    var font: Font

    var body: some View {
        GeometryReader { geo in
            ZStack(alignment: .topLeading) {
                Text(verbatim: "Ag")
                    .font(font)
                    .fixedSize()
                    .hidden()
                    .background {
                        GeometryReader { lineGeo in
                            Color.clear.preference(
                                key: PVTextAreaLineHeightKey.self,
                                value: lineGeo.size.height
                            )
                        }
                    }
                Text(verbatim: text.isEmpty ? " " : text)
                    .font(font)
                    .frame(width: geo.size.width, alignment: .leading)
                    .fixedSize(horizontal: false, vertical: true)
                    .hidden()
                    .background {
                        GeometryReader { contentGeo in
                            Color.clear.preference(
                                key: PVTextAreaContentHeightKey.self,
                                value: contentGeo.size.height
                            )
                        }
                    }
            }
        }
        .allowsHitTesting(false)
        .accessibilityHidden(true)
    }
}

private struct PVTextAreaLineHeightKey: PreferenceKey {
    static let defaultValue: CGFloat = 0
    static func reduce(value: inout CGFloat, nextValue: () -> CGFloat) {
        value = max(value, nextValue())
    }
}

private struct PVTextAreaContentHeightKey: PreferenceKey {
    static let defaultValue: CGFloat = 0
    static func reduce(value: inout CGFloat, nextValue: () -> CGFloat) {
        value = max(value, nextValue())
    }
}

/// `TextEditor` insets its glyphs past the chrome padding, so a Source title
/// would jump between resting text and edit mode. SwiftUI has no inset
/// modifier for that padding; this clears the text container only.
private struct PVTextEditorInsetClearer: NSViewRepresentable {
    func makeNSView(context: Context) -> PVTextEditorInsetView {
        PVTextEditorInsetView()
    }

    func updateNSView(_ nsView: PVTextEditorInsetView, context: Context) {
        nsView.clearInsets()
    }
}

private final class PVTextEditorInsetView: NSView {
    private let tabRouter = PVTextAreaTabRouter()

    override func viewDidMoveToSuperview() {
        super.viewDidMoveToSuperview()
        clearInsets()
    }

    override func layout() {
        super.layout()
        clearInsets()
    }

    func clearInsets() {
        guard let textView = nearestTextView() else { return }
        textView.textContainerInset = .zero
        textView.textContainer?.lineFragmentPadding = 0
        textView.drawsBackground = false
        textView.backgroundColor = .clear
        if let scroll = textView.enclosingScrollView {
            scroll.drawsBackground = false
            scroll.backgroundColor = .clear
            scroll.scrollerStyle = .overlay
            scroll.autohidesScrollers = true
        }
        let alreadyRouting = textView.delegate === tabRouter
        tabRouter.attach(to: textView)
        if !alreadyRouting { scheduleTabRouterRecheck() }
    }

    /// SwiftUI may install its own delegate after this layout pass. Look
    /// again on the next turn, and stop after a couple of tries.
    private func scheduleTabRouterRecheck() {
        guard tabRouterRechecks < 2 else { return }
        tabRouterRechecks += 1
        DispatchQueue.main.async { [weak self] in
            self?.clearInsets()
        }
    }

    private var tabRouterRechecks = 0

    /// The clearer is a background of this editor, so the text view is a
    /// nearby sibling. Stop after a few ancestors so a second area on the
    /// same screen is left alone.
    private func nearestTextView() -> NSTextView? {
        var node: NSView? = superview
        var hops = 0
        while let current = node, hops < 4 {
            if let found = current.firstDescendantTextView() {
                return found
            }
            node = current.superview
            hops += 1
        }
        return nil
    }
}

/// Tab and Shift-Tab leave a form text area. Option-Tab uses a different
/// command and still inserts a tab character.
enum PVTextAreaTabRouting {
    static func movesFocus(_ selector: Selector) -> Bool {
        selector == #selector(NSResponder.insertTab(_:))
            || selector == #selector(NSResponder.insertBacktab(_:))
    }
}

/// Sits in front of the text view's existing delegate so SwiftUI still hears
/// edits. Tab moves through the key loop; every other command is passed on.
final class PVTextAreaTabRouter: NSObject, NSTextViewDelegate {
    private weak var wrapped: NSTextViewDelegate?

    @MainActor
    func attach(to textView: NSTextView) {
        if textView.delegate === self { return }
        wrapped = textView.delegate
        textView.delegate = self
    }

    func textView(_ textView: NSTextView, doCommandBy commandSelector: Selector) -> Bool {
        guard PVTextAreaTabRouting.movesFocus(commandSelector) else {
            return wrapped?.textView?(textView, doCommandBy: commandSelector) ?? false
        }
        if commandSelector == #selector(NSResponder.insertBacktab(_:)) {
            textView.window?.selectPreviousKeyView(nil)
        } else {
            textView.window?.selectNextKeyView(nil)
        }
        return true
    }

    override func responds(to aSelector: Selector!) -> Bool {
        if super.responds(to: aSelector) { return true }
        return wrapped?.responds(to: aSelector) ?? false
    }

    override func forwardingTarget(for aSelector: Selector!) -> Any? {
        if let wrapped, wrapped.responds(to: aSelector) { return wrapped }
        return super.forwardingTarget(for: aSelector)
    }
}

private extension NSView {
    func firstDescendantTextView() -> NSTextView? {
        if let textView = self as? NSTextView { return textView }
        for subview in subviews {
            if let found = subview.firstDescendantTextView() { return found }
        }
        return nil
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
