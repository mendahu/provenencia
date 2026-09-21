import SwiftUI

/// Floating Add Person / Event / Place + Connect tools over the Evidence graph (S6-D1 / S6-D2).
struct EvidenceGraphPalette: View {
    @Bindable var model: EvidenceGraphModel
    var focus: FocusState<EvidenceGraphFocus?>.Binding

    var body: some View {
        HStack(spacing: 2) {
            ForEach(EvidencePrimaryKind.allCases, id: \.self) { kind in
                toolButton(kind)
            }
            Rectangle()
                .fill(PVColor.borderSubtle)
                .frame(width: 1, height: 22)
                .padding(.horizontal, 4)
            connectButton
        }
        .padding(4)
        .background(
            RoundedRectangle(cornerRadius: PVRadius.md, style: .continuous)
                .fill(PVColor.surfaceCard)
                .overlay(
                    RoundedRectangle(cornerRadius: PVRadius.md, style: .continuous)
                        .strokeBorder(PVColor.borderSubtle, lineWidth: 1)
                )
                .shadow(color: Color.black.opacity(0.08), radius: 8, y: 2)
        )
        .opacity(model.canCite ? 1 : 0.55)
    }

    private func toolButton(_ kind: EvidencePrimaryKind) -> some View {
        let armed = model.armedKind == kind
        let style = EvidenceSubjectKindStyle.resolve(
            typeKey: kind.rawValue,
            presentation: model.presentation(for: kind.rawValue)
        )
        return Button {
            model.toggleArm(kind)
        } label: {
            HStack(spacing: 8) {
                PVMark(kind.markKey, size: 17)
                    .foregroundStyle(armed ? PVColor.accentForeground : style.ink)
                Text(model.toolName(for: kind))
                    .font(PVFont.body(size: PVTypeScale.caption, weight: PVFontWeight.medium))
            }
            .padding(.horizontal, 11)
            .padding(.vertical, 6)
            .foregroundStyle(
                armed
                    ? PVColor.accentForeground
                    : (model.canCite ? PVColor.textSecondary : PVColor.textFaint)
            )
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
            .pvFocusRing(focus.wrappedValue == .tool(kind), cornerRadius: 4)
        }
        .buttonStyle(.plain)
        .disabled(!model.canCite)
        .focused(focus, equals: .tool(kind))
        .onKeyPress(.escape) {
            model.disarm()
            focus.wrappedValue = nil
            return .handled
        }
        .accessibilityLabel(Text(verbatim: model.toolAccessibilityLabel(for: kind, armed: armed)))
        .accessibilityAddTraits(armed ? [.isSelected] : [])
        .accessibilityIdentifier("evidenceGraph.palette.\(kind.rawValue)")
    }

    private var connectButton: some View {
        let armed = model.armedConnect
        return Button {
            model.toggleConnect()
        } label: {
            HStack(spacing: 8) {
                Image(systemName: "point.topleft.down.to.point.bottomright.curvepath")
                    .font(.system(size: 13, weight: .medium))
                Text(L10n.EvidenceGraph.toolConnect)
                    .font(PVFont.body(size: PVTypeScale.caption, weight: PVFontWeight.medium))
            }
            .padding(.horizontal, 11)
            .padding(.vertical, 6)
            .foregroundStyle(
                armed
                    ? PVColor.accentForeground
                    : (model.canCite ? PVColor.textSecondary : PVColor.textFaint)
            )
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
            .pvFocusRing(focus.wrappedValue == .toolConnect, cornerRadius: 4)
        }
        .buttonStyle(.plain)
        .disabled(!model.canCite)
        .focused(focus, equals: .toolConnect)
        .onKeyPress(.escape) {
            model.disarm()
            focus.wrappedValue = nil
            return .handled
        }
        .accessibilityLabel(Text(verbatim: model.connectToolAccessibilityLabel(armed: armed)))
        .accessibilityAddTraits(armed ? [.isSelected] : [])
        .accessibilityIdentifier("evidenceGraph.palette.connect")
    }
}

/// Keyboard focus targets for the floating palette tools.
enum EvidenceGraphFocus: Hashable {
    case tool(EvidencePrimaryKind)
    case toolConnect
}
