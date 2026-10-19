import SwiftUI

/// One static grey bar standing in for a line of text while content loads.
/// `width` nil fills the proposed width. Content-agnostic: lists and forms
/// stack these under their own loading label.
struct PVSkeletonBar: View {
    var width: CGFloat?
    var height: CGFloat

    var body: some View {
        RoundedRectangle(cornerRadius: 2, style: .continuous)
            .fill(PVColor.borderSubtle)
            .frame(width: width, height: height)
    }
}

#Preview {
    VStack(alignment: .leading, spacing: PVSpacing.space5) {
        PVSkeletonBar(width: 120, height: 12)
        PVSkeletonBar(width: nil, height: 12)
    }
    .padding(PVSpacing.space6)
    .frame(width: 240)
    .background(PVColor.surfaceCard)
}
