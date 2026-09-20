import SwiftUI

/// A 1pt hairline in `PVColor.borderSubtle` (override with `color`).
/// Horizontal fills width; vertical fills height.
struct PVDivider: View {
    enum Axis {
        case horizontal, vertical
    }

    var axis: Axis = .horizontal
    var color: Color = PVColor.borderSubtle

    var body: some View {
        Rectangle()
            .fill(color)
            .frame(
                width: axis == .vertical ? 1 : nil,
                height: axis == .horizontal ? 1 : nil
            )
    }
}

#Preview {
    VStack(spacing: PVSpacing.space6) {
        Text("Above")
        PVDivider()
        HStack {
            Text("Left")
            PVDivider(axis: .vertical)
            Text("Right")
        }
        .frame(height: 40)
        PVDivider(color: PVColor.borderDefault)
    }
    .padding(PVSpacing.space9)
    .background(PVColor.surfacePage)
}
