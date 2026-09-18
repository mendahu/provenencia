@preconcurrency import AppKit
import SwiftUI

/// One row in a `PVComboBox` list. `subtext` is the secondary line the
/// default row renders under the label; it is matched by the filter too, so
/// a caller can put a machine key there and let the researcher type either.
struct PVComboBoxOption: Identifiable, Equatable {
    let value: String
    let label: String
    let subtext: String
    let isDisabled: Bool

    var id: String { value }

    init(value: String, label: String, subtext: String = "", isDisabled: Bool = false) {
        self.value = value
        self.label = label
        self.subtext = subtext
        self.isDisabled = isDisabled
    }
}

// MARK: - Pure helpers

/// Filtering and highlight rules, split out so they can be unit-tested
/// without a view — same split as `PVTableTypeSelectMatcher`.
enum PVComboBoxMatch {
    /// Decompose, drop combining marks, lowercase — so "Ó Riain" is reachable
    /// by typing "o ri" (mirrors `norm` in `ComboBox.jsx`).
    static func normalize(_ text: String) -> String {
        text.folding(options: [.diacriticInsensitive, .caseInsensitive], locale: nil)
    }

    /// A row matches when the normalized query appears in its label or its
    /// subtext. Substring, not prefix: a key like `publication-date` should
    /// be findable by typing `date`.
    static func matches(_ option: PVComboBoxOption, normalizedQuery: String) -> Bool {
        guard !normalizedQuery.isEmpty else { return true }
        return normalize(option.label).contains(normalizedQuery)
            || normalize(option.subtext).contains(normalizedQuery)
    }

    static func filter(_ options: [PVComboBoxOption], query: String) -> [PVComboBoxOption] {
        let q = normalize(query.trimmingCharacters(in: .whitespacesAndNewlines))
        guard !q.isEmpty else { return options }
        return options.filter { matches($0, normalizedQuery: q) }
    }

    /// Index of the first row that can be committed, or -1 when every row is
    /// disabled. This is the Spotlight rule's landing spot.
    static func firstEnabledIndex(in options: [PVComboBoxOption]) -> Int {
        options.firstIndex { !$0.isDisabled } ?? -1
    }

    /// Arrow-key movement. Unlike `PVTable`'s wrapping list this **clamps at
    /// the ends** — an `NSMenu` does not cycle — and steps over disabled
    /// rows in the direction of travel. Returns the current index unchanged
    /// when there is nowhere legal to go.
    ///
    /// `from` may be -1 (nothing active yet), in which case a downward move
    /// lands on the first row and an upward move on the last.
    static func move(
        from index: Int,
        delta: Int,
        in options: [PVComboBoxOption],
        to edge: Edge? = nil
    ) -> Int {
        let last = options.count - 1
        guard last >= 0 else { return -1 }
        var i: Int
        switch edge {
        case .first: i = 0
        case .last: i = last
        case nil: i = index < 0 ? (delta > 0 ? 0 : last) : index + delta
        }
        i = min(max(0, i), last)
        let step = (edge == .last || delta < 0) ? -1 : 1
        while options[i].isDisabled, i + step >= 0, i + step <= last {
            i += step
        }
        return options[i].isDisabled ? index : i
    }

    enum Edge {
        case first, last
    }
}

/// Where the popup window sits relative to the field.
enum PVComboBoxPlacement {
    /// Air between the field and its list.
    static let gap: CGFloat = 4

    /// Screen frame for the popup window. All rects are AppKit **screen**
    /// coordinates, so y grows upward and "below the field" means a smaller
    /// y than the field's `minY`.
    ///
    /// The popup is placed strictly outside the anchor — it can never cover
    /// the text you are typing into — and always inside `visibleFrame`, so it
    /// cannot run off the display. It prefers to open downward and flips
    /// above the field only when it does not fit below *and* there is more
    /// room up there; whichever side it lands on, the height is trimmed to
    /// the space actually available, leaving the list to scroll inside it.
    static func popupFrame(
        anchor: CGRect,
        contentHeight: CGFloat,
        visibleFrame: CGRect,
        gap: CGFloat = gap
    ) -> CGRect {
        let spaceBelow = max(0, anchor.minY - visibleFrame.minY - gap)
        let spaceAbove = max(0, visibleFrame.maxY - anchor.maxY - gap)
        let opensUp = contentHeight > spaceBelow && spaceAbove > spaceBelow
        let height = min(contentHeight, opensUp ? spaceAbove : spaceBelow)
        let y = opensUp ? anchor.maxY + gap : anchor.minY - gap - height
        let width = min(anchor.width, visibleFrame.width)
        let x = min(max(anchor.minX, visibleFrame.minX), visibleFrame.maxX - width)
        return CGRect(x: x, y: y, width: width, height: height)
    }
}

/// Renders `text` with the matched span wearing the highlighter mark.
enum PVComboBoxHighlight {
    static func attributed(_ text: String, query: String) -> AttributedString {
        var full = AttributedString(text)
        let q = query.trimmingCharacters(in: .whitespacesAndNewlines)
        guard !q.isEmpty else { return full }
        // Search the normalized forms so the highlight lands on the same span
        // the filter matched (accent- and case-insensitively).
        guard let range = text.range(
            of: q,
            options: [.diacriticInsensitive, .caseInsensitive]
        ), let attributed = Range(range, in: full) else {
            return full
        }
        // Typed `AttributeContainer` subscripts — not `.backgroundColor` /
        // `.foregroundColor` dynamic members — avoid Swift 6 key-path
        // Sendable warnings that otherwise spam every `PVComboBox` call site.
        var mark = AttributeContainer()
        mark[AttributeScopes.SwiftUIAttributes.BackgroundColorAttribute.self] =
            PVColor.markBackground
        mark[AttributeScopes.SwiftUIAttributes.ForegroundColorAttribute.self] =
            PVColor.markForeground
        full[attributed].mergeAttributes(mark)
        return full
    }
}

// MARK: - Default row

/// The `plain` row kind: label, and a muted second line when the option has
/// `subtext`. Both lines carry the match highlight.
struct PVComboBoxPlainRow: View {
    let option: PVComboBoxOption
    let query: String

    var body: some View {
        VStack(alignment: .leading, spacing: 1) {
            Text(PVComboBoxHighlight.attributed(option.label, query: query))
                .font(PVFont.body(size: PVTypeScale.bodySmall))
                .foregroundStyle(PVColor.textPrimary)
                .lineLimit(1)
            if !option.subtext.isEmpty {
                Text(PVComboBoxHighlight.attributed(option.subtext, query: query))
                    .font(PVFont.body(size: PVTypeScale.micro))
                    .foregroundStyle(PVColor.textMuted)
                    .lineLimit(1)
            }
        }
        .frame(maxWidth: .infinity, alignment: .leading)
    }
}

// MARK: - ComboBox

/// A searchable single-select field — the design system's
/// `components/forms/ComboBox.jsx`, ported for macOS 14.
///
/// It replaces a `PVSelect` wherever the pool is long enough that scanning a
/// popup menu is worse than typing: the field filters as you type, arrow keys
/// walk the list, and Return commits.
///
/// **Interaction contract** (all of it is load-bearing — the web component
/// documents why):
/// - The field's text is seeded from the committed option's label, so *having
///   text* is not the same as *having typed*. Only typed text filters, which
///   is why re-opening the field shows the whole list rather than just the row
///   already chosen.
/// - Spotlight rule: while there is a query the top hit is pre-highlighted so
///   Return takes it. With no query nothing is highlighted — it is a browse
///   list, and Return does nothing.
/// - Arrow keys **clamp** at the ends rather than cycling, matching `NSMenu`.
///   ⌘↑/⌘↓ and Home/End jump to the ends; Page Up/Down move by 8.
/// - Escape is two-stage: close the list, then abandon the typed text.
/// - Clicking anywhere outside the field or its list closes the list **and
///   blurs the field** — outside means done. Clicking the field itself
///   reopens a closed list.
/// - The pointer and the keyboard drive the *same* highlight, so there is no
///   separate hover state.
///
/// **The list is a child window, not an overlay.** A SwiftUI overlay is
/// clipped by any enclosing `ScrollView` and by the app window, which makes it
/// useless for a field near the bottom of a scrolling inspector. `NSMenu`,
/// `NSComboBox` and Spotlight all draw into their own borderless window for
/// exactly that reason, and so does this: `PVComboBoxPlacement` positions it
/// in screen coordinates against the field, so it clips to the *display*
/// rather than to any view, may extend past the app window's edge, and is
/// placed strictly outside the field so it can never cover what you type. The
/// window never becomes key, so the text field keeps focus and keeps handling
/// keys while the list is open.
///
/// **Not ported** (add here on demand, the way `PVSidebarNav` was): the
/// multi-select token field, the "use what you typed" create row, the
/// `person` row kind, and async `loading`. No call site needs them yet.
struct PVComboBox<Row: View>: View {
    /// The committed option's `value`; empty means nothing is chosen.
    @Binding var selection: String
    let options: [PVComboBoxOption]
    var size: PVControlSize = .md
    var placeholder: LocalizedStringResource?
    var emptyLabel: LocalizedStringResource
    var isInvalid: Bool = false
    var maxListHeight: CGFloat = 288
    /// Accessibility label for the field and its list — components never
    /// invent their own (`docs/macos-client-patterns.md` §5).
    let label: LocalizedStringResource
    var accessibilityIdentifierPrefix: String?
    /// When true, focus the field and open the list as soon as the control
    /// appears — for edit-in-place flows that swap a chip for a blank picker.
    /// Default stays false so sheet first-responder (e.g. Add Source) does not
    /// pop the list on open.
    var activateOnAppear: Bool = false
    @ViewBuilder var row: (PVComboBoxOption, String) -> Row

    @State private var query = ""
    /// Whether the current text came from the keyboard rather than from
    /// mirroring the committed label. Only typed text filters.
    @State private var isDirty = false
    @State private var isOpen = false
    /// The highlighted row, or -1. The keyboard and the pointer both drive
    /// it — there is no separate hover state.
    @State private var activeIndex = -1
    /// Natural height of `rowsStack`, reported by `measuringRows`. The popup
    /// window is sized from this rather than from anything inside it.
    @State private var measuredRowsHeight: CGFloat = 0
    @FocusState private var isFocused: Bool
    @Environment(\.accessibilityReduceMotion) private var reduceMotion

    /// Gutter that holds the committed row's checkmark. Every row reserves it
    /// so labels stay aligned whether or not one is checked.
    private static var checkGutter: CGFloat { 22 }
    /// Air above and below the rows, inside the list's border.
    private static var listVerticalPadding: CGFloat { PVSpacing.space2 }

    /// How tall the popup wants to be: the rows plus the list's own padding.
    /// The coordinator clamps this to `maxListHeight` and to the room on
    /// screen, and the scroller takes up whatever it lands on.
    private var desiredPopupHeight: CGFloat {
        measuredRowsHeight + Self.listVerticalPadding * 2
    }

    private var selectedOption: PVComboBoxOption? {
        options.first { $0.value == selection }
    }

    private var selectedLabel: String {
        selectedOption?.label ?? ""
    }

    /// True only when the researcher typed something — see `isDirty`.
    private var hasTypedQuery: Bool {
        isDirty && !query.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty
    }

    private var visibleOptions: [PVComboBoxOption] {
        hasTypedQuery ? PVComboBoxMatch.filter(options, query: query) : options
    }

    /// The query the rows highlight against — empty unless the user typed.
    private var highlightQuery: String {
        hasTypedQuery ? query.trimmingCharacters(in: .whitespacesAndNewlines) : ""
    }

    var body: some View {
        field
            .background(measuringRows)
            .background(
                PVComboBoxPopupWindow(
                    isPresented: isOpen,
                    maxHeight: maxListHeight,
                    // Only things that change the window's *size* are passed
                    // as metrics. `activeIndex` changes on every mouse move,
                    // and re-framing the window that often tore down click
                    // handling inside it.
                    contentHeight: desiredPopupHeight,
                    optionCount: visibleOptions.count,
                    onDismiss: dismissAndBlur,
                    content: list
                )
            )
            .onChange(of: selectedLabel, initial: true) { _, newValue in
                // Mirror an externally-set value into the field, but not over
                // the top of someone mid-edit. The exception is the selection
                // being cleared or dropping out of `options` — the caller
                // does that after acting on the choice (assigning the field
                // removes it from the pool), and the field must not go on
                // showing a label that no longer names anything.
                guard !isFocused || selectedOption == nil else { return }
                query = newValue
                isDirty = false
            }
            .onChange(of: isFocused) { _, focused in
                // Losing focus closes the list and restores the field. Gaining
                // focus alone does **not** open — a sheet that lands first
                // responder on this field (Add Source) would otherwise pop the
                // list immediately. Open comes from a click, a keystroke, or
                // an arrow / Home / End / Page key. The popup's mouse monitor
                // still catches clicks on things that take no focus (a button,
                // a label, empty chrome) and resolves them via `dismissAndBlur`,
                // which resigns focus and lands back here.
                if !focused { closeAndRestore() }
            }
            .onChange(of: isOpen) { _, open in
                if !open { activeIndex = -1 }
            }
            .onAppear {
                guard activateOnAppear else { return }
                isFocused = true
                open()
            }
    }

    // MARK: Field

    private var field: some View {
        TextField("", text: typedText, prompt: placeholder.map { Text($0) })
            .textFieldStyle(.plain)
            .focused($isFocused)
            .onKeyPress(action: handleKey)
            .modifier(PVInputChrome(
                size: size,
                isFocused: isFocused,
                isInvalid: isInvalid,
                leadingIconInset: true
            ))
            .overlay(alignment: .leading) {
                PVIcon(.search, size: PVInputChrome.iconSize)
                    .foregroundStyle(PVColor.textFaint)
                    .padding(.leading, PVInputChrome.horizontalInset)
                    .allowsHitTesting(false)
            }
            // Clicking a focused field whose list is closed (Escape closed
            // it, or a commit did) reopens it — without this the component
            // is stuck until focus leaves and comes back. Simultaneous, so
            // the text field still places the caret.
            .simultaneousGesture(TapGesture().onEnded {
                if !isOpen { open() }
            })
            .accessibilityLabel(Text(label))
            .accessibilityIdentifier(accessibilityIdentifierPrefix ?? "")
    }

    /// The text binding the field edits. Only the *setter* marks the text as
    /// typed, which is what separates a keystroke from the code assigning
    /// `query` (mirroring a committed label, reverting on Escape, restoring
    /// on dismiss) — those assign `query` directly and leave `isDirty` alone.
    ///
    /// The same-value guard matters: `TextField` re-writes its binding when
    /// editing ends, and treating that write as typing reopened the list on
    /// blur. Only an actual change counts.
    private var typedText: Binding<String> {
        Binding(
            get: { query },
            set: { typed in
                guard typed != query else { return }
                query = typed
                isDirty = true
                isOpen = true
                resetActiveRow()
            }
        )
    }

    // MARK: List

    private var list: some View {
        scroller
            .padding(.vertical, Self.listVerticalPadding)
            .background(
                RoundedRectangle(cornerRadius: PVRadius.md, style: .continuous)
                    .fill(PVColor.surfaceCard)
            )
            .overlay(
                RoundedRectangle(cornerRadius: PVRadius.md, style: .continuous)
                    .strokeBorder(PVColor.borderDefault, lineWidth: 1)
            )
            .clipShape(RoundedRectangle(cornerRadius: PVRadius.md, style: .continuous))
            // No `.pvShadow` here: the popup is its own window, so AppKit
            // casts the shadow from the window's alpha — see
            // `DesignSystem/README.md` § "Platform deviations".
            .accessibilityLabel(Text(label))
            .accessibilityIdentifier(identifier("list") ?? "")
    }

    /// The rows at their natural height. Shared by the scroller and by the
    /// hidden measuring copy behind the field, so the two can never disagree.
    private var rowsStack: some View {
        VStack(spacing: 0) {
            ForEach(Array(visibleOptions.enumerated()), id: \.element.id) { index, option in
                rowView(option, index: index).id(option.id)
            }
            if visibleOptions.isEmpty {
                Text(emptyLabel)
                    .font(PVFont.body(size: PVTypeScale.bodySmall, italic: true))
                    .foregroundStyle(PVColor.textMuted)
                    .fixedSize(horizontal: false, vertical: true)
                    .frame(maxWidth: .infinity, alignment: .leading)
                    .padding(.leading, Self.checkGutter + PVSpacing.space5)
                    .padding(.trailing, PVSpacing.space5)
                    .padding(.vertical, PVSpacing.space4)
                    .accessibilityIdentifier(identifier("empty") ?? "")
            }
        }
    }

    /// The scroller simply **fills** the popup window — the window is the
    /// only thing that clamps its height. It deliberately does not also try
    /// to report an ideal height (an earlier version used
    /// `.frame(maxHeight:)` + `.fixedSize`): asking one view to both size the
    /// window and lay out inside it means any disagreement between the two
    /// clips a row, which is exactly what happened.
    private var scroller: some View {
        ScrollViewReader { proxy in
            ScrollView {
                rowsStack
            }
            .onChange(of: activeIndex) { _, index in
                guard visibleOptions.indices.contains(index) else { return }
                withAnimation(reduceMotion ? nil : PVMotion.fastStandard) {
                    proxy.scrollTo(visibleOptions[index].id)
                }
            }
            .onChange(of: visibleOptions.map(\.id)) { _, ids in
                // A narrowed list starts at the top. Without this the scroller
                // keeps whatever offset the longer list left behind, hiding
                // the first result of the query that just narrowed it.
                guard let first = ids.first else { return }
                proxy.scrollTo(first, anchor: .top)
            }
        }
    }

    /// A hidden, non-interactive copy of the rows, laid out behind the field
    /// at the field's width purely to measure them.
    ///
    /// The popup window has to know how tall to be *before* it is shown, and
    /// a `ScrollView` inside that window cannot tell it — nor can
    /// `NSHostingView.fittingSize`, reliably. Laying the rows out once more
    /// here is the honest way to get the number: a background never
    /// influences its parent's size, so this costs a layout pass and nothing
    /// else. The pool is a vocabulary and the rows are single-line, so that
    /// pass is cheap.
    private var measuringRows: some View {
        rowsStack
            .fixedSize(horizontal: false, vertical: true)
            .hidden()
            .allowsHitTesting(false)
            .accessibilityHidden(true)
            .background(
                GeometryReader { proxy in
                    Color.clear
                        .onChange(of: proxy.size.height, initial: true) { _, height in
                            measuredRowsHeight = height
                        }
                }
            )
    }

    /// The row shell — highlight fill, check gutter, hit area — is owned here
    /// so every row stays aligned whatever the caller draws inside it.
    private func rowView(_ option: PVComboBoxOption, index: Int) -> some View {
        HStack(spacing: PVSpacing.space5) {
            PVIcon(.check, size: 13)
                .foregroundStyle(PVColor.accent)
                .opacity(option.value == selection ? 1 : 0)
                .frame(width: Self.checkGutter)
                .accessibilityHidden(true)
            row(option, highlightQuery)
        }
        .padding(.trailing, PVSpacing.space5)
        .padding(.vertical, PVSpacing.space3)
        .frame(maxWidth: .infinity, alignment: .leading)
        .background(index == activeIndex ? PVColor.surfaceSelected : .clear)
        .opacity(option.isDisabled ? 0.42 : 1)
        .contentShape(Rectangle())
        // Pointer and keyboard share one highlight, so moving the mouse makes
        // a row active rather than painting a second, competing hover state.
        .onHover { inside in
            if inside, !option.isDisabled {
                activeIndex = index
            }
        }
        // The sole commit path for a click. The popup window never becomes
        // key, but mouse events are still delivered to the window under the
        // pointer, so the gesture fires; the mouse monitor passes panel
        // clicks through untouched.
        .onTapGesture { commit(option) }
        .accessibilityAddTraits(option.value == selection ? [.isSelected] : [])
        .accessibilityIdentifier(identifier("option.\(option.value)") ?? "")
    }

    // MARK: Behaviour

    private func open() {
        isOpen = true
        resetActiveRow()
    }

    /// Spotlight rule: a query pre-selects the top hit so Return takes it;
    /// with no query the list is a browse list and nothing is highlighted.
    private func resetActiveRow() {
        activeIndex = hasTypedQuery ? PVComboBoxMatch.firstEnabledIndex(in: visibleOptions) : -1
    }

    private func commit(_ option: PVComboBoxOption) {
        guard !option.isDisabled else { return }
        selection = option.value
        query = option.label
        isDirty = false
        isOpen = false
    }

    /// Dismissing abandons whatever was typed and puts the committed label
    /// back — the field is a picker, not free text.
    private func closeAndRestore() {
        isOpen = false
        query = selectedLabel
        isDirty = false
    }

    /// The popup's dismissal (click away, window resigning key). On top of
    /// closing, it resigns focus: clicking anywhere outside the field or its
    /// list means "I'm done here", and a field that kept its focus ring after
    /// that read as a bug.
    private func dismissAndBlur() {
        closeAndRestore()
        isFocused = false
    }

    private func handleKey(_ press: KeyPress) -> KeyPress.Result {
        switch press.key {
        case .downArrow, .upArrow:
            let delta = press.key == .downArrow ? 1 : -1
            let edge: PVComboBoxMatch.Edge? = press.modifiers.contains(.command)
                ? (delta > 0 ? .last : .first)
                : nil
            moveActive(delta: delta, to: edge)
            return .handled
        case .home:
            moveActive(delta: 1, to: .first)
            return .handled
        case .end:
            moveActive(delta: -1, to: .last)
            return .handled
        case .pageDown, .pageUp:
            moveActive(delta: press.key == .pageDown ? 8 : -8, to: nil)
            return .handled
        case .return:
            guard visibleOptions.indices.contains(activeIndex) else { return .ignored }
            commit(visibleOptions[activeIndex])
            return .handled
        case .escape:
            // Two-stage: close the list first, then abandon the typed text.
            // With neither left to undo, let Escape through — an enclosing
            // sheet or dialog still has to be dismissable from this field.
            if isOpen {
                isOpen = false
                return .handled
            }
            if query != selectedLabel {
                query = selectedLabel
                isDirty = false
                return .handled
            }
            return .ignored
        case .tab:
            // macOS accepts the highlight on tab-out, then lets focus leave.
            if isOpen, visibleOptions.indices.contains(activeIndex) {
                commit(visibleOptions[activeIndex])
            }
            isOpen = false
            return .ignored
        default:
            return .ignored
        }
    }

    private func moveActive(delta: Int, to edge: PVComboBoxMatch.Edge?) {
        if !isOpen { open() }
        activeIndex = PVComboBoxMatch.move(from: activeIndex, delta: delta, in: visibleOptions, to: edge)
    }

    private func identifier(_ suffix: String) -> String? {
        accessibilityIdentifierPrefix.map { "\($0).\(suffix)" }
    }
}

extension PVComboBox where Row == PVComboBoxPlainRow {
    /// The `plain` row kind — label plus optional muted subtext.
    init(
        selection: Binding<String>,
        options: [PVComboBoxOption],
        size: PVControlSize = .md,
        placeholder: LocalizedStringResource? = nil,
        emptyLabel: LocalizedStringResource,
        isInvalid: Bool = false,
        maxListHeight: CGFloat = 288,
        label: LocalizedStringResource,
        accessibilityIdentifierPrefix: String? = nil,
        activateOnAppear: Bool = false
    ) {
        self.init(
            selection: selection,
            options: options,
            size: size,
            placeholder: placeholder,
            emptyLabel: emptyLabel,
            isInvalid: isInvalid,
            maxListHeight: maxListHeight,
            label: label,
            accessibilityIdentifierPrefix: accessibilityIdentifierPrefix,
            activateOnAppear: activateOnAppear
        ) { option, query in
            PVComboBoxPlainRow(option: option, query: query)
        }
    }
}

// MARK: - Popup window

/// Hosts `content` in a borderless child window anchored to the view this is
/// a background of. It sits in the view tree only to get hold of an `NSView`
/// to measure against — it draws nothing itself.
private struct PVComboBoxPopupWindow<Content: View>: NSViewRepresentable {
    let isPresented: Bool
    let maxHeight: CGFloat
    /// The height the content wants, measured by the caller (see
    /// `PVComboBox.measuringRows`).
    let contentHeight: CGFloat
    /// How many rows the list is showing. Together with `contentHeight` this
    /// is what decides whether the window needs re-framing — SwiftUI hands us
    /// a fresh `content` on every update, so the values are compared instead.
    let optionCount: Int
    let onDismiss: () -> Void
    let content: Content

    func makeNSView(context: Context) -> NSView {
        let view = NSView(frame: .zero)
        context.coordinator.anchor = view
        return view
    }

    func updateNSView(_: NSView, context: Context) {
        let coordinator = context.coordinator
        coordinator.onDismiss = onDismiss
        coordinator.maxHeight = maxHeight
        coordinator.contentHeight = contentHeight
        coordinator.optionCount = optionCount
        coordinator.wantsPresented = isPresented
        coordinator.setContent(content)
        // Showing or moving a window mutates AppKit state, which must not
        // happen inside SwiftUI's layout pass — hence the hop. The closure
        // deliberately captures no values: `applyPresentation` reads the
        // coordinator's *current* fields, so a closure queued before a
        // dismissal cannot re-show the panel with stale state.
        DispatchQueue.main.async {
            coordinator.applyPresentation()
        }
    }

    static func dismantleNSView(_: NSView, coordinator: Coordinator) {
        coordinator.tearDown()
    }

    func makeCoordinator() -> Coordinator {
        Coordinator()
    }

    @MainActor
    final class Coordinator {
        weak var anchor: NSView?
        var onDismiss: () -> Void = {}
        var maxHeight: CGFloat = 288
        var contentHeight: CGFloat = 0
        var optionCount = 0
        /// The single source of truth for whether the panel should be up,
        /// written synchronously by every `updateNSView` and by `dismiss()`.
        /// `applyPresentation` mirrors it into AppKit.
        var wantsPresented = false

        private var panel: PVComboBoxPanel?
        private var hosting: PVComboBoxHostingView?
        /// Kept unwidened so `position` can re-wrap it at the field's width.
        private var content: AnyView = AnyView(EmptyView())
        private var clickMonitor: Any?
        private var observers: [NSObjectProtocol] = []
        /// The content metrics the panel was last framed for; `nil` while it
        /// is down. Comparing against these is what keeps re-framing (which
        /// is disruptive mid-interaction) to actual size changes.
        private var shownHeight: CGFloat?
        private var shownCount: Int?

        func setContent(_ content: some View) {
            self.content = AnyView(content)
            if let hosting, let panel, panel.parent != nil {
                hosting.rootView = widened(to: panel.frame.width)
            }
        }

        /// The popup is exactly as wide as the field, and measuring it at that
        /// width is what makes `fittingSize` trustworthy — an unconstrained
        /// `NSHostingView` reports the height its content would take at its
        /// *ideal* width, which is not the width we are about to give it.
        private func widened(to width: CGFloat) -> AnyView {
            AnyView(content.frame(width: width))
        }

        /// Brings the panel in line with `wantsPresented` and the current
        /// content metrics. Idempotent, and always reads current state — it
        /// is safe to call from a stale queued closure.
        func applyPresentation() {
            guard wantsPresented else {
                hide()
                return
            }
            if panel?.parent == nil {
                show()
                shownHeight = contentHeight
                shownCount = optionCount
            } else if shownHeight != contentHeight || shownCount != optionCount {
                // Re-frame when the rows changed, not on every SwiftUI update.
                shownHeight = contentHeight
                shownCount = optionCount
                position()
            }
        }

        // MARK: Show / hide

        private func show() {
            guard let anchor, let parent = anchor.window else { return }
            if hosting == nil {
                hosting = PVComboBoxHostingView(rootView: widened(to: anchor.bounds.width))
            }
            guard let hosting else { return }
            if panel == nil {
                let created = PVComboBoxPanel(
                    contentRect: NSRect(x: 0, y: 0, width: 100, height: 100),
                    styleMask: [.borderless, .nonactivatingPanel],
                    backing: .buffered,
                    defer: false
                )
                created.isOpaque = false
                created.backgroundColor = .clear
                // AppKit derives the shadow from the window's alpha, so the
                // rounded content gets a correctly-shaped native popup shadow.
                created.hasShadow = true
                created.level = .popUpMenu
                created.animationBehavior = .none
                created.contentView = hosting
                panel = created
            }
            guard let panel, panel.parent == nil else { return }
            position()
            parent.addChildWindow(panel, ordered: .above)
            startWatching(parent: parent)
        }

        private func hide() {
            shownHeight = nil
            shownCount = nil
            guard let panel, panel.parent != nil else { return }
            stopWatching()
            panel.parent?.removeChildWindow(panel)
            panel.orderOut(nil)
        }

        func tearDown() {
            stopWatching()
            if let panel {
                panel.parent?.removeChildWindow(panel)
                panel.orderOut(nil)
            }
            panel = nil
            hosting = nil
        }

        // MARK: Geometry

        /// Sizes the panel to its content (clamped by `maxHeight` and by the
        /// room on screen) and puts it just outside the field.
        private func position() {
            guard let panel, let anchor, let window = anchor.window, let hosting else { return }
            let anchorOnScreen = window.convertToScreen(anchor.convert(anchor.bounds, to: nil))

            hosting.rootView = widened(to: anchorOnScreen.width)
            let wanted = min(contentHeight, maxHeight)

            let visible = (window.screen ?? NSScreen.main)?.visibleFrame ?? anchorOnScreen
            panel.setFrame(
                PVComboBoxPlacement.popupFrame(
                    anchor: anchorOnScreen,
                    contentHeight: wanted,
                    visibleFrame: visible
                ),
                display: true
            )
            panel.invalidateShadow()
        }

        // MARK: Dismissal

        /// SwiftUI does not resign a `TextField`'s focus when you click inert
        /// content, so "clicked away" has to be observed at the event level —
        /// the same job the web component's `mousedown` listener does.
        private func startWatching(parent: NSWindow) {
            clickMonitor = NSEvent.addLocalMonitorForEvents(
                matching: [.leftMouseDown, .rightMouseDown, .otherMouseDown]
            ) { [weak self] event in
                self?.handleMouseDown(event)
                return event
            }
            // No global monitor: a click in another app deactivates this one,
            // and the `didResignKeyNotification` observer below already
            // dismisses on that.

            let center = NotificationCenter.default
            // Follow the field when the window moves or resizes, and when the
            // pane it lives in scrolls — a detached popup is worse than none.
            for name in [NSWindow.didMoveNotification, NSWindow.didResizeNotification] {
                observers.append(center.addObserver(forName: name, object: parent, queue: .main) { [weak self] _ in
                    MainActor.assumeIsolated { self?.position() }
                })
            }
            observers.append(
                center.addObserver(forName: NSWindow.didResignKeyNotification, object: parent, queue: .main) { [weak self] _ in
                    MainActor.assumeIsolated { self?.dismiss() }
                }
            )
            if let clip = anchor?.enclosingScrollView?.contentView {
                clip.postsBoundsChangedNotifications = true
                observers.append(
                    center.addObserver(forName: NSView.boundsDidChangeNotification, object: clip, queue: .main) { [weak self] _ in
                        MainActor.assumeIsolated { self?.position() }
                    }
                )
            }
        }

        private func stopWatching() {
            if let clickMonitor {
                NSEvent.removeMonitor(clickMonitor)
            }
            clickMonitor = nil
            observers.forEach(NotificationCenter.default.removeObserver)
            observers.removeAll()
        }

        /// Every mouse-down in the app while the popup is open lands here.
        /// There are exactly two things it can be: a click on a row — which
        /// the monitor passes through untouched, the row's own tap gesture
        /// commits it — or a click away. Anything that is not in the popup
        /// and not on the field itself dismisses — no exceptions, whether or
        /// not the thing clicked takes focus.
        private func handleMouseDown(_ event: NSEvent) {
            if let panel, event.window === panel { return }
            if let anchor, event.window === anchor.window {
                let point = anchor.convert(event.locationInWindow, from: nil)
                // Clicking the field it belongs to leaves the list up — the
                // caret moves, the list stays.
                if anchor.bounds.contains(point) { return }
            }
            dismiss()
        }

        /// A dismissal decided on the AppKit side (click away, window resigns
        /// key). Writing `wantsPresented` first means any queued
        /// `applyPresentation` agrees with the hide instead of undoing it;
        /// `onDismiss` then brings the SwiftUI state along.
        private func dismiss() {
            wantsPresented = false
            hide()
            onDismiss()
        }
    }
}

/// Borderless popup that never takes key focus, so the text field behind it
/// keeps handling arrows, Return and Escape while the list is open.
private final class PVComboBoxPanel: NSPanel {
    override var canBecomeKey: Bool { false }
    override var canBecomeMain: Bool { false }
}

/// Hosting view for the popup's content. Rows commit on the first click, so a
/// click arriving while the app is inactive still lands on the row instead of
/// being swallowed to activate the app.
private final class PVComboBoxHostingView: NSHostingView<AnyView> {
    override func acceptsFirstMouse(for event: NSEvent?) -> Bool { true }

    @available(*, unavailable)
    required init(coder: NSCoder) {
        fatalError("init(coder:) is not used — the popup is built in code")
    }

    @MainActor
    required init(rootView: AnyView) {
        super.init(rootView: rootView)
    }
}

private struct PVComboBoxPreview: View {
    @State private var selection = ""

    var body: some View {
        VStack(alignment: .leading, spacing: PVSpacing.space7) {
            PVComboBox(
                selection: $selection,
                options: [
                    PVComboBoxOption(value: "f1", label: "Author", subtext: "author"),
                    PVComboBoxOption(value: "f2", label: "Publication date", subtext: "publication-date"),
                    PVComboBoxOption(value: "f3", label: "Publisher", subtext: "publisher"),
                    PVComboBoxOption(value: "f4", label: "Repository", subtext: "repository"),
                ],
                placeholder: L10n.SourceTypes.assignPlaceholder,
                emptyLabel: L10n.SourceTypes.assignNoMatch,
                label: L10n.SourceTypes.assignFieldLabel
            )
            Text(verbatim: selection.isEmpty ? "nothing chosen" : selection)
                .font(PVFont.mono(size: PVTypeScale.micro))
                .foregroundStyle(PVColor.textMuted)
            Spacer()
        }
        .padding(PVSpacing.space9)
        .frame(width: 380, height: 420)
        .background(PVColor.surfaceCard)
    }
}

#Preview {
    PVComboBoxPreview()
}
