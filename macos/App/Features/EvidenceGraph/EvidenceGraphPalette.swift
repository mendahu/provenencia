import SwiftUI

/// Floating Add Person / Event / Place + Connect + Select tools over the Evidence graph (S6-D1 / S6-D2 / S10-D6).
struct EvidenceGraphPalette: View {
    @Bindable var model: EvidenceGraphModel
    var focus: FocusState<EvidenceGraphFocus?>.Binding

    var body: some View {
        HStack(spacing: 2) {
            // Cite tools dim together when the Source has no Artifact. Select does not:
            // choosing cards writes nothing to the catalog.
            HStack(spacing: 2) {
                ForEach(EvidencePrimaryKind.allCases, id: \.self) { kind in
                    kindTool(kind)
                }
                hairline
                connectTool
            }
            .opacity(model.canCite ? 1 : 0.55)
            hairline
            selectTool
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
    }

    private var hairline: some View {
        Rectangle()
            .fill(PVColor.borderSubtle)
            .frame(width: 1, height: 22)
            .padding(.horizontal, 4)
    }

    private func kindTool(_ kind: EvidencePrimaryKind) -> some View {
        let style = EvidenceSubjectKindStyle.resolve(
            typeKey: kind.rawValue,
            presentation: model.presentation(for: kind.rawValue)
        )
        return paletteTool(
            title: model.toolName(for: kind),
            armed: model.armedKind == kind,
            enabled: model.canCite,
            focusCase: .tool(kind),
            identifier: "evidenceGraph.palette.\(kind.rawValue)",
            action: { model.toggleArm(kind) }
        ) { armed in
            PVMark(kind.markKey, size: 17)
                .foregroundStyle(armed ? PVColor.accentForeground : style.ink)
        }
    }

    private var connectTool: some View {
        paletteTool(
            title: L10n.EvidenceGraph.toolConnect,
            armed: model.armedConnect,
            enabled: model.canCite,
            focusCase: .toolConnect,
            identifier: "evidenceGraph.palette.connect",
            action: { model.toggleConnect() }
        ) { _ in
            Image(systemName: "point.topleft.down.to.point.bottomright.curvepath")
                .font(.system(size: 13, weight: .medium))
        }
    }

    private var selectTool: some View {
        paletteTool(
            title: L10n.EvidenceGraph.toolSelect,
            armed: model.inputMode == .selecting,
            enabled: true,
            help: L10n.EvidenceGraph.toolSelectHelp,
            focusCase: .toolSelect,
            identifier: "evidenceGraph.palette.select",
            action: { model.toggleSelect() }
        ) { _ in
            PVIcon(.regionRectangle, size: 17)
        }
    }

    /// One palette button: armed is the accent fill with the graph ring, a disabled
    /// tool fades its label. `icon` receives whether the tool is armed.
    private func paletteTool<Icon: View>(
        title: LocalizedStringResource,
        armed: Bool,
        enabled: Bool,
        help: LocalizedStringResource? = nil,
        focusCase: EvidenceGraphFocus,
        identifier: String,
        action: @escaping () -> Void,
        @ViewBuilder icon: (Bool) -> Icon
    ) -> some View {
        let glyph = icon(armed)
        return Button(action: action) {
            HStack(spacing: 8) {
                glyph
                Text(title)
                    .font(PVFont.body(size: PVTypeScale.caption, weight: PVFontWeight.medium))
            }
            .padding(.horizontal, 11)
            .padding(.vertical, 6)
            .foregroundStyle(
                armed
                    ? PVColor.accentForeground
                    : (enabled ? PVColor.textSecondary : PVColor.textFaint)
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
            .pvFocusRing(focus.wrappedValue == focusCase, cornerRadius: 4)
            .contentShape(Rectangle())
        }
        .buttonStyle(.plain)
        .disabled(!enabled)
        .focused(focus, equals: focusCase)
        .onKeyPress(.escape) {
            model.disarm()
            focus.wrappedValue = nil
            return .handled
        }
        .modifier(PaletteHelp(help: help))
        .accessibilityLabel(Text(verbatim: model.toolAccessibilityLabel(name: title, armed: armed)))
        .accessibilityAddTraits(armed ? [.isSelected] : [])
        .accessibilityIdentifier(identifier)
    }
}

/// Adds a tooltip only when a tool has one.
private struct PaletteHelp: ViewModifier {
    let help: LocalizedStringResource?

    @ViewBuilder
    func body(content: Content) -> some View {
        if let help {
            content.help(Text(help))
        } else {
            content
        }
    }
}

/// Keyboard focus targets for the floating palette tools.
enum EvidenceGraphFocus: Hashable {
    case tool(EvidencePrimaryKind)
    case toolConnect
    case toolSelect
}
