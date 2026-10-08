import SwiftUI

/// Immediate on/off control — the design system's `components/PVSwitch.jsx`.
/// A 32×18 track (accent when on, paper when off) and a 14pt thumb.
/// For a choice inside a form that is saved later, use ``PVCheckbox``.
struct PVSwitch: View {
    private let label: Text?
    private let accessibilityLabelText: Text?
    @Binding private var isOn: Bool
    private let isDisabled: Bool

    init(
        _ label: LocalizedStringResource,
        isOn: Binding<Bool>,
        isDisabled: Bool = false
    ) {
        self.label = Text(label)
        self.accessibilityLabelText = nil
        self._isOn = isOn
        self.isDisabled = isDisabled
    }

    init(
        verbatim label: String,
        isOn: Binding<Bool>,
        isDisabled: Bool = false
    ) {
        self.label = Text(verbatim: label)
        self.accessibilityLabelText = nil
        self._isOn = isOn
        self.isDisabled = isDisabled
    }

    /// Track only. Pass what VoiceOver should say.
    init(
        isOn: Binding<Bool>,
        isDisabled: Bool = false,
        accessibilityLabel: LocalizedStringResource
    ) {
        self.label = nil
        self.accessibilityLabelText = Text(accessibilityLabel)
        self._isOn = isOn
        self.isDisabled = isDisabled
    }

    var body: some View {
        Button {
            isOn.toggle()
        } label: {
            HStack(spacing: PVSpacing.space5) {
                track
                label?
                    .font(PVFont.body(size: PVTypeScale.bodySmall))
                    .foregroundStyle(PVColor.textPrimary)
            }
            .contentShape(Rectangle())
        }
        .buttonStyle(.plain)
        .disabled(isDisabled)
        .opacity(isDisabled ? 0.5 : 1)
        .modifier(SwitchAccessibility(
            spoken: label == nil ? accessibilityLabelText : nil,
            isOn: isOn
        ))
    }

    /// Off-track paper. Dark mode uses the same ink as a resting border so the
    /// thumb still reads on the page.
    private static let trackOff = Color.pvDynamic(light: PVPalette.paper300, dark: PVPalette.hex("#3C352C"))

    private var track: some View {
        ZStack(alignment: .leading) {
            Capsule(style: .continuous)
                .fill(isOn ? PVColor.accent : Self.trackOff)
            Circle()
                .fill(PVPalette.paper0)
                .pvShadow(PVElevation.sm)
                .frame(width: 14, height: 14)
                .padding(.leading, isOn ? 16 : 2)
        }
        .frame(width: 32, height: 18)
        .pvAnimation(PVMotion.easeStandard, value: isOn)
    }
}

/// A labeled switch keeps its visible text. A track-only switch speaks `spoken`.
private struct SwitchAccessibility: ViewModifier {
    var spoken: Text?
    var isOn: Bool

    func body(content: Content) -> some View {
        let marked = content.accessibilityAddTraits(isOn ? [.isButton, .isSelected] : .isButton)
        if let spoken {
            marked
                .accessibilityElement(children: .ignore)
                .accessibilityLabel(spoken)
        } else {
            marked.accessibilityElement(children: .combine)
        }
    }
}

#Preview("PVSwitch") {
    PVSwitchPreview()
        .padding(PVSpacing.space8)
        .background(PVColor.surfacePage)
}

private struct PVSwitchPreview: View {
    @State private var shown = true
    @State private var hidden = false

    var body: some View {
        VStack(alignment: .leading, spacing: PVSpacing.space6) {
            PVSwitch(verbatim: "Show inferred relationships", isOn: $shown)
            PVSwitch(verbatim: "Off", isOn: $hidden)
            PVSwitch(verbatim: "Disabled", isOn: .constant(true), isDisabled: true)
        }
    }
}
