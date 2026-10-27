import SwiftUI

/// An Evidence graph's full count line: "12 subjects · 40 observations", or
/// "0 subjects · not started", a short rule while the counts load, or a danger
/// mark beside "Counts could not load" when they failed. The Source page's graph
/// cards draw it.
struct SourcesGraphCountLine: View {
    var progress: EvidenceGraphProgress?
    var countsLoading: Bool
    var countsError: Error?

    private var counts: SourcesGraphCounts {
        SourcesGraphCounts(progress: progress, countsLoading: countsLoading, countsError: countsError)
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
            PVSkeletonBar(width: nil, height: 1)
                .padding(.trailing, PVSpacing.space3)
                .padding(.vertical, 5)
        case .unavailable(let message):
            HStack(spacing: PVSpacing.space2) {
                PVIcon(.danger, size: 11)
                    .foregroundStyle(PVColor.danger)
                Text(verbatim: L10n.Sources.graphCounts(counts, style: .full))
                    .font(PVFont.body(size: PVTypeScale.micro))
                    .foregroundStyle(PVColor.textMuted)
                    .lineLimit(1)
            }
            .help(message)
        case .notStarted:
            // Split only so "not started" can be italic body beside mono counts;
            // the words come from the same strings `graphCounts(.notStarted, .full)` joins.
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
