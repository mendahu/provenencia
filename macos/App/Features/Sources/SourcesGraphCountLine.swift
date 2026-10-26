import SwiftUI

/// An Evidence graph's full count line: "12 subjects · 40 observations", or
/// "0 subjects · not started", or a short rule while the counts load. The
/// Sources list's split row and the Source page's graph cards draw it.
struct SourcesGraphCountLine: View {
    var progress: EvidenceGraphProgress?
    var countsLoading: Bool

    private var counts: SourcesGraphCounts {
        SourcesGraphCounts(progress: progress, countsLoading: countsLoading)
    }

    var body: some View {
        content
            .accessibilityElement(children: .ignore)
            .accessibilityLabel(Text(verbatim: L10n.Sources.graphCounts(counts, style: .spoken)))
    }

    @ViewBuilder
    private var content: some View {
        switch counts {
        case .loading:
            Rectangle()
                .fill(PVColor.borderSubtle)
                .frame(height: 1)
                .padding(.trailing, PVSpacing.space3)
                .padding(.vertical, 5)
        case .notStarted:
            HStack(alignment: .firstTextBaseline, spacing: PVSpacing.space2) {
                Text(verbatim: L10n.Sources.graphSubjectCount(0))
                    .font(PVFont.mono(size: PVTypeScale.micro))
                Text(verbatim: "·")
                    .font(PVFont.mono(size: PVTypeScale.micro))
                Text(L10n.Sources.graphNotStarted)
                    .font(PVFont.body(size: PVTypeScale.micro))
                    .italic()
            }
            .foregroundStyle(PVColor.textMuted)
            .lineLimit(1)
        case .counts:
            Text(verbatim: L10n.Sources.graphCounts(counts, style: .full))
                .font(PVFont.mono(size: PVTypeScale.micro))
                .foregroundStyle(PVColor.textMuted)
                .lineLimit(1)
        }
    }
}
