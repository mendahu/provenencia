import SwiftUI

/// A row's band and, when it applies, its Contested mark: what the proposal
/// says about the pair, drawn once for the Promote row and the evidence
/// sheet's title.
struct PromoteReadoutBadges: View {
    let readout: PromoteFlow.SelectionReadout

    var body: some View {
        HStack(spacing: PVSpacing.space3) {
            PVBadge(L10n.Promote.assessment(readout.assessment), tone: tone)
            if readout.contested {
                PVBadge(L10n.Promote.contested, tone: .danger, icon: .warning)
            }
        }
    }

    private var tone: PVBadgeTone {
        switch readout.assessment {
        case .strong: .success
        case .medium: .info
        case .weak: .warning
        case .none: .neutral
        }
    }
}

#Preview("Readout badges") {
    VStack(alignment: .leading, spacing: PVSpacing.space4) {
        PromoteReadoutBadges(readout: PromoteFlow.SelectionReadout(assessment: .strong, reason: .empty))
        PromoteReadoutBadges(readout: PromoteFlow.SelectionReadout(assessment: .none, reason: .decided, contested: true))
    }
    .padding(PVSpacing.space8)
    .background(PVColor.surfaceCard)
}
