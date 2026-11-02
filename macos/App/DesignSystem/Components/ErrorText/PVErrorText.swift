import SwiftUI

/// A validation or failure line in the danger color. `.micro` is the field-level
/// line (under one control); `.caption` is the form-level line. Callers keep
/// their own `.padding`, `.fixedSize` and `.accessibilityIdentifier`.
struct PVErrorText: View {
    enum Size {
        case micro
        case caption

        fileprivate var points: CGFloat {
            switch self {
            case .micro: PVTypeScale.micro
            case .caption: PVTypeScale.caption
            }
        }
    }

    private let text: String
    private let size: Size
    private let announces: Bool

    /// `announces: true` speaks the line to VoiceOver when it appears or changes.
    init(_ text: String, size: Size = .micro, announces: Bool = false) {
        self.text = text
        self.size = size
        self.announces = announces
    }

    var body: some View {
        Text(verbatim: text)
            .font(PVFont.body(size: size.points))
            .foregroundStyle(PVColor.danger)
            .pvAnnounce(announces ? text : nil, priority: .high)
    }
}

#Preview {
    VStack(alignment: .leading, spacing: PVSpacing.space2) {
        PVErrorText("Enter a year.")
        PVErrorText("The date could not be saved.", size: .caption)
    }
    .padding()
}
