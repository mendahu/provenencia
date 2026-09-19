import SwiftUI

/// Floating Add Person / Event / Place tools over the Evidence graph pane (S6-D1).
struct EvidenceGraphPalette: View {
    @Bindable var model: EvidenceGraphModel

    var body: some View {
        HStack(spacing: 2) {
            ForEach(EvidencePrimaryKind.allCases, id: \.self) { kind in
                toolButton(kind)
            }
        }
        .padding(4)
        .background(PVColor.surfaceCard)
        .overlay(
            RoundedRectangle(cornerRadius: PVRadius.md, style: .continuous)
                .strokeBorder(PVColor.borderSubtle, lineWidth: 1)
        )
        .clipShape(RoundedRectangle(cornerRadius: PVRadius.md, style: .continuous))
        .shadow(color: Color.black.opacity(0.08), radius: 8, y: 2)
        .accessibilityElement(children: .contain)
    }

    private func toolButton(_ kind: EvidencePrimaryKind) -> some View {
        let armed = model.armedKind == kind
        let style = EvidenceSubjectKindStyle.forKind(kind)
        return Button {
            model.toggleArm(kind)
        } label: {
            HStack(spacing: 8) {
                PVSubjectIcon(kind: iconKind(kind), size: 17)
                    .foregroundStyle(armed ? PVColor.accentForeground : style.ink)
                Text(model.toolName(for: kind))
                    .font(PVFont.body(size: PVTypeScale.caption, weight: PVFontWeight.medium))
            }
            .padding(.horizontal, 11)
            .padding(.vertical, 6)
            .foregroundStyle(armed ? PVColor.accentForeground : PVColor.textSecondary)
            .background(
                RoundedRectangle(cornerRadius: 4, style: .continuous)
                    .fill(armed ? PVColor.accent : Color.clear)
            )
            .background {
                if armed {
                    RoundedRectangle(cornerRadius: 4, style: .continuous)
                        .fill(PVColor.graphRing)
                        .padding(-3)
                }
            }
        }
        .buttonStyle(.plain)
        .accessibilityLabel(Text(verbatim: model.toolAccessibilityLabel(for: kind, armed: armed)))
        .accessibilityAddTraits(armed ? [.isSelected] : [])
        .accessibilityIdentifier("evidenceGraph.palette.\(kind.rawValue)")
    }

    private func iconKind(_ kind: EvidencePrimaryKind) -> PVSubjectIconKind {
        switch kind {
        case .person: .person
        case .event: .event
        case .place: .place
        }
    }
}
