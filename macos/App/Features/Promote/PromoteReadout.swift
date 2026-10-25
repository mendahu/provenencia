import SwiftUI

/// What a Promote row says beside its name, drawn once for the Promote row and the evidence
/// sheet's title: its band and, when it applies, its Contested mark (what the proposal says about
/// the pair).
struct PromoteReadoutBadges: View {
    let readout: PromoteFlow.SelectionReadout

    var body: some View {
        HStack(spacing: PVSpacing.space3) {
            PVBadge(L10n.Promote.assessment(readout), tone: tone)
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

/// "Updated · …": why a suggested row moved on the last proposal. Drawn under the Promote row and
/// under the evidence sheet's title.
struct PromoteUpdatedCaption: View {
    /// Why it moved, as `PromoteModel.reasonText` words it.
    let reason: String

    var body: some View {
        Text(verbatim: L10n.Promote.updatedDetail(reason))
            .font(PVFont.body(size: PVTypeScale.caption))
            .italic()
            .foregroundStyle(PVColor.infoForeground)
    }
}

#Preview("Readout badges") {
    VStack(alignment: .leading, spacing: PVSpacing.space4) {
        PromoteReadoutBadges(readout: PromoteFlow.SelectionReadout(assessment: .strong, reason: .empty))
        PromoteReadoutBadges(readout: PromoteFlow.SelectionReadout(assessment: .none, reason: .decided, contested: true, chosen: true))
    }
    .padding(PVSpacing.space8)
    .background(PVColor.surfaceCard)
}
