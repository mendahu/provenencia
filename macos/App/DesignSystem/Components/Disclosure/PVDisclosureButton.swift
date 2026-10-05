import SwiftUI

/// A button that shows and hides a section the caller lays out.
///
/// Expanding changes three things: the open state, which chevron the trigger
/// shows, and what VoiceOver says. The button owns all three: it flips
/// `isExpanded`, picks the trailing chevron for a plain `PVButton`, and
/// reports expanded / collapsed with `pvExpandedState`. The disclosed content
/// stays with the caller (`if isExpanded { … }`), so a trigger can sit in one
/// grid column and open a panel that spans the row.
///
/// The web kit has no Disclosure component; its boards compose `Button.jsx`
/// with `iconRight` and `aria-expanded`, which this mirrors.
struct PVDisclosureButton<Title: PVCopy>: View {
    private let title: Title
    @Binding private var isExpanded: Bool
    private let variant: PVButtonVariant
    private let size: PVControlSize

    init(
        _ title: Title,
        isExpanded: Binding<Bool>,
        variant: PVButtonVariant = .ghost,
        size: PVControlSize = .sm
    ) {
        self.title = title
        self._isExpanded = isExpanded
        self.variant = variant
        self.size = size
    }

    /// Opens a closed section, closes an open one.
    func toggle() {
        isExpanded.toggle()
    }

    var body: some View {
        PVButton(title, variant: variant, size: size, iconRight: PVExpandedState.chevron(isExpanded: isExpanded), action: toggle)
            .pvExpandedState(isExpanded)
    }
}

/// A disclosure's state as drawn and as spoken.
enum PVExpandedState {
    /// The trailing chevron: down to open, up to close.
    static func chevron(isExpanded: Bool) -> PVSymbol {
        isExpanded ? .chevronUp : .chevronDown
    }

    static func spoken(isExpanded: Bool) -> LocalizedStringResource {
        isExpanded ? L10n.DesignSystem.disclosureExpanded : L10n.DesignSystem.disclosureCollapsed
    }
}

extension View {
    /// Reports expanded / collapsed to VoiceOver as custom content (not in
    /// the value), for any control that shows or hides something: a
    /// `PVDisclosureButton`, a `PVSelect` trigger, a whole-row toggle.
    func pvExpandedState(_ isExpanded: Bool) -> some View {
        accessibilityCustomContent(
            Text(L10n.DesignSystem.disclosureState),
            Text(PVExpandedState.spoken(isExpanded: isExpanded))
        )
    }
}

#Preview {
    VStack(alignment: .leading, spacing: PVSpacing.space4) {
        PVDisclosureButton("2 other values", isExpanded: .constant(false))
        PVDisclosureButton("Why", isExpanded: .constant(true))
        PVCard(tone: .sunken, cornerRadius: PVRadius.sm, padding: PVSpacing.space5) {
            Text(verbatim: "Disclosed content")
        }
    }
    .padding()
}
