import SwiftUI

/// Multi-select control — the design system's `components/PVCheckbox.jsx`.
/// A 16pt box (accent when checked or indeterminate, with a check or minus)
/// beside an optional label and a muted description. The description pulls
/// the box to the top line.
///
/// `PVCheckboxMark` is the bare box for rows that are themselves the
/// control. `isLocked` is the seeded-binding lock the Properties board
/// draws in that same box: a filled square with a lock, not a check.
/// The row carries the accessibility; the mark is decoration.
struct PVCheckbox: View {
    private let label: Text?
    private let description: Text?
    @Binding private var isChecked: Bool
    private let isIndeterminate: Bool
    private let isDisabled: Bool
    private let accessibilityLabelText: Text?

    init(
        _ label: LocalizedStringResource,
        description: LocalizedStringResource? = nil,
        isChecked: Binding<Bool>,
        isIndeterminate: Bool = false,
        isDisabled: Bool = false
    ) {
        self.label = Text(label)
        self.description = description.map { Text($0) }
        self._isChecked = isChecked
        self.isIndeterminate = isIndeterminate
        self.isDisabled = isDisabled
        self.accessibilityLabelText = nil
    }

    init(
        verbatim label: String,
        description: String? = nil,
        isChecked: Binding<Bool>,
        isIndeterminate: Bool = false,
        isDisabled: Bool = false
    ) {
        self.label = Text(verbatim: label)
        self.description = description.map { Text(verbatim: $0) }
        self._isChecked = isChecked
        self.isIndeterminate = isIndeterminate
        self.isDisabled = isDisabled
        self.accessibilityLabelText = nil
    }

    /// Box only. Pass what VoiceOver should say; the column header is not enough.
    init(
        isChecked: Binding<Bool>,
        isIndeterminate: Bool = false,
        isDisabled: Bool = false,
        accessibilityLabel: LocalizedStringResource
    ) {
        self.label = nil
        self.description = nil
        self._isChecked = isChecked
        self.isIndeterminate = isIndeterminate
        self.isDisabled = isDisabled
        self.accessibilityLabelText = Text(accessibilityLabel)
    }

    /// Box only, with spoken copy the caller composed (an `L10n` format
    /// filled with what this box marks), when one fixed label can't say it.
    init(
        isChecked: Binding<Bool>,
        isIndeterminate: Bool = false,
        isDisabled: Bool = false,
        accessibilitySpokenLabel: String
    ) {
        self.label = nil
        self.description = nil
        self._isChecked = isChecked
        self.isIndeterminate = isIndeterminate
        self.isDisabled = isDisabled
        self.accessibilityLabelText = Text(verbatim: accessibilitySpokenLabel)
    }

    var body: some View {
        Button {
            isChecked.toggle()
        } label: {
            HStack(alignment: description == nil ? .center : .top, spacing: PVSpacing.space5) {
                PVCheckboxMark(isChecked: isChecked, isIndeterminate: isIndeterminate)
                    .padding(.top, description == nil ? 0 : 2)
                if label != nil || description != nil {
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
            }
            .contentShape(Rectangle())
        }
        .buttonStyle(.plain)
        .disabled(isDisabled)
        .opacity(isDisabled ? 0.5 : 1)
        .modifier(CheckboxAccessibility(
            spoken: label == nil ? accessibilityLabelText : nil,
            isOn: isChecked
        ))
    }
}

/// The checkbox box alone: 16pt, raised fill and inset shadow when clear,
/// accent fill and a check (or minus, when indeterminate) when marked.
/// `isLocked` swaps the check for a lock on a paper fill — a binding the
/// researcher cannot clear.
struct PVCheckboxMark: View {
    var isChecked: Bool
    var isIndeterminate: Bool = false
    var isLocked: Bool = false

    private var isOn: Bool { isChecked || isIndeterminate || isLocked }

    var body: some View {
        ZStack {
            RoundedRectangle(cornerRadius: PVRadius.xs, style: .continuous)
                .fill(fill)
                .pvInsetShadow(cornerRadius: PVRadius.xs, visible: !isOn)
            RoundedRectangle(cornerRadius: PVRadius.xs, style: .continuous)
                .strokeBorder(stroke, lineWidth: 1)
            PVIcon(glyph, size: 11)
                .foregroundStyle(PVColor.accentForeground)
                .opacity(isOn ? 1 : 0)
        }
        .frame(width: 16, height: 16)
        .accessibilityHidden(true)
    }

    private var fill: Color {
        if !isOn { return PVColor.surfaceRaised }
        if isLocked { return PVPalette.paper500 }
        return PVColor.accent
    }

    private var stroke: Color {
        if !isOn { return PVColor.borderStrong }
        if isLocked { return .clear }
        return PVColor.accent
    }

    private var glyph: PVSymbol {
        if isLocked { return .lock }
        if isIndeterminate { return .minus }
        return .check
    }
}

/// A labeled control keeps its visible text. A box-only control speaks `spoken`.
private struct CheckboxAccessibility: ViewModifier {
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

#Preview("PVCheckbox") {
    PVCheckboxPreview()
        .padding(PVSpacing.space8)
        .background(PVColor.surfacePage)
}

private struct PVCheckboxPreview: View {
    @State private var images = true
    @State private var index = false
    @State private var pin = true

    var body: some View {
        VStack(alignment: .leading, spacing: PVSpacing.space6) {
            PVCheckbox(verbatim: "Only records with images", description: "Excludes index-only entries", isChecked: $images)
            PVCheckbox(verbatim: "Include the index", isChecked: $index)
            PVCheckbox(verbatim: "Disabled", isChecked: .constant(true), isDisabled: true)
            HStack(spacing: PVSpacing.space4) {
                PVCheckbox(isChecked: $pin, accessibilitySpokenLabel: "Pin")
                PVCheckboxMark(isChecked: false)
                PVCheckboxMark(isChecked: false, isIndeterminate: true)
                PVCheckboxMark(isChecked: true, isLocked: true)
            }
        }
    }
}
