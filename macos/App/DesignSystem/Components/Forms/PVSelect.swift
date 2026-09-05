import SwiftUI

/// One option in a `PVSelect` — `value` is the stable identifier (what a
/// binding stores), `label` is what's shown. Plain `String`, not
/// `LocalizedStringResource`: options are almost always data (a project
/// filename, a person's name), not static UI copy — see
/// `docs/macos-client-patterns.md` §6.
struct PVSelectOption: Identifiable {
    let id: String
    let label: String

    init(value: String, label: String) {
        self.id = value
        self.label = label
    }
}

/// A styled dropdown — mirrors `components/forms/Select.jsx` (chevron via
/// `PVIcon`, same rest border as `PVInput`). Built on `Menu`, not `Picker`,
/// so the label and chrome can be fully restyled with `PV*` tokens rather
/// than inheriting `NSPopUpButton`'s default appearance.
///
/// See `DesignSystem/README.md` / `PVButton.swift` for the component
/// pattern this follows.
struct PVSelect: View {
    @Binding private var selection: String
    private let options: [PVSelectOption]
    private let size: PVControlSize

    init(selection: Binding<String>, options: [PVSelectOption], size: PVControlSize = .md) {
        self._selection = selection
        self.options = options
        self.size = size
    }

    private var currentLabel: String {
        options.first(where: { $0.id == selection })?.label ?? ""
    }

    var body: some View {
        Menu {
            ForEach(options) { option in
                Button(option.label) { selection = option.id }
            }
        } label: {
            HStack(spacing: PVSpacing.space3) {
                Text(currentLabel)
                    .foregroundStyle(PVColor.textPrimary)
                    .lineLimit(1)
                Spacer(minLength: 0)
                PVIcon("chevron-down", size: 14)
                    .foregroundStyle(PVColor.textFaint)
            }
            .font(size.font)
            .padding(.horizontal, PVSpacing.space3 + PVSpacing.space1)
            .frame(height: size.height)
            .frame(maxWidth: .infinity)
            .background(
                RoundedRectangle(cornerRadius: PVRadius.sm, style: .continuous)
                    .fill(PVColor.surfaceRaised)
            )
            .overlay(
                RoundedRectangle(cornerRadius: PVRadius.sm, style: .continuous)
                    .stroke(PVColor.borderDefault, lineWidth: 1)
            )
            .pvInsetShadow(cornerRadius: PVRadius.sm)
        }
        .menuStyle(.borderlessButton)
        .buttonStyle(.plain)
    }
}

#Preview {
    PVSelect(
        selection: .constant("smith"),
        options: [
            PVSelectOption(value: "smith", label: "smith-family.provenencia"),
            PVSelectOption(value: "halvorsen", label: "halvorsen-line.provenencia"),
            PVSelectOption(value: "baca", label: "baca-oaxaca.provenencia")
        ]
    )
    .padding(PVSpacing.space9)
    .frame(width: 320)
    .background(PVColor.surfacePage)
}
