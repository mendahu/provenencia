import SwiftUI

/// One choice in a set — the design system's `components/PVRadio.jsx`,
/// ported for macOS 14. A 16pt ring (accent when chosen, with an 8pt dot)
/// beside an optional label and a muted description; the description pulls
/// the ring to the top line.
///
/// Group semantics are the caller's: put the radios in a container that
/// names the choice (`.accessibilityElement(children: .contain)` +
/// `.accessibilityLabel`), and keep exactly one `isSelected` true once the
/// researcher has chosen. Call sites own `.accessibilityIdentifier`.
///
/// `PVRadioMark` is the bare ring for rows that are themselves the
/// selectable control (a candidate row with a leading radio) — the row
/// carries the accessibility, the mark is decoration.
struct PVRadio: View {
    private let label: Text?
    private let description: Text?
    private let isSelected: Bool
    private let isDisabled: Bool
    private let action: () -> Void

    /// Fixed UI copy.
    init(
        _ label: LocalizedStringResource,
        description: LocalizedStringResource? = nil,
        isSelected: Bool,
        isDisabled: Bool = false,
        action: @escaping () -> Void
    ) {
        self.label = Text(label)
        self.description = description.map { Text($0) }
        self.isSelected = isSelected
        self.isDisabled = isDisabled
        self.action = action
    }

    /// Copy already localized (formatted with runtime values).
    init(
        verbatim label: String,
        description: String? = nil,
        isSelected: Bool,
        isDisabled: Bool = false,
        action: @escaping () -> Void
    ) {
        self.label = Text(verbatim: label)
        self.description = description.map { Text(verbatim: $0) }
        self.isSelected = isSelected
        self.isDisabled = isDisabled
        self.action = action
    }

    var body: some View {
        Button(action: action) {
            HStack(alignment: description == nil ? .center : .top, spacing: PVSpacing.space5) {
                PVRadioMark(isSelected: isSelected)
                    .padding(.top, description == nil ? 0 : 2)
                VStack(alignment: .leading, spacing: 1) {
                    label?
                        .font(PVFont.body(size: PVTypeScale.bodySmall))
                        .foregroundStyle(PVColor.textPrimary)
                    description?
                        .font(PVFont.body(size: PVTypeScale.micro))
                        .foregroundStyle(PVColor.textMuted)
                        .fixedSize(horizontal: false, vertical: true)
                }
            }
            .contentShape(Rectangle())
        }
        .buttonStyle(.plain)
        .disabled(isDisabled)
        .opacity(isDisabled ? 0.5 : 1)
        .accessibilityElement(children: .combine)
        .accessibilityAddTraits(isSelected ? [.isButton, .isSelected] : .isButton)
    }
}

/// The radio ring alone: 16pt, raised paper fill, accent ring and 8pt dot
/// when chosen, strong border and inset shadow when not. Decorative.
struct PVRadioMark: View {
    let isSelected: Bool

    var body: some View {
        ZStack {
            Circle()
                .fill(PVColor.surfaceRaised)
                .pvInsetShadow(cornerRadius: 8, visible: !isSelected)
            Circle()
                .strokeBorder(isSelected ? PVColor.accent : PVColor.borderStrong, lineWidth: 1)
            if isSelected {
                Circle()
                    .fill(PVColor.accent)
                    .frame(width: 8, height: 8)
            }
        }
        .frame(width: 16, height: 16)
        .accessibilityHidden(true)
    }
}

#Preview("PVRadio") {
    VStack(alignment: .leading, spacing: PVSpacing.space6) {
        PVRadio(verbatim: "New Person", description: "Start a new Person with James Robins as its first member", isSelected: true) {}
        PVRadio(verbatim: "Existing Person", description: "Join a Person you already have", isSelected: false) {}
        PVRadio(verbatim: "Disabled", isSelected: false, isDisabled: true) {}
        HStack(spacing: PVSpacing.space4) {
            PVRadioMark(isSelected: true)
            PVRadioMark(isSelected: false)
        }
    }
    .padding(PVSpacing.space8)
    .background(PVColor.surfacePage)
}
