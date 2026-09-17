import SwiftUI

/// Split-row Sources list cell: filing zone | Evidence graph zone (S5-D2).
struct SourcesSplitRow: View {
    static let graphZoneWidth: CGFloat = 210

    let source: CatalogSource
    let typeLabel: String
    let typeIconKey: String?
    let projectDir: String
    var onOpenPage: () -> Void
    var onOpenGraph: () -> Void

    var body: some View {
        HStack(spacing: 0) {
            pageZone
            graphZone
                .frame(width: Self.graphZoneWidth)
                .overlay(alignment: .leading) {
                    Rectangle()
                        .fill(PVColor.borderSubtle)
                        .frame(width: 1)
                }
        }
        .frame(maxWidth: .infinity, alignment: .leading)
        .overlay(alignment: .bottom) {
            PVDivider()
        }
        .accessibilityElement(children: .contain)
        .accessibilityIdentifier("sources.row.\(source.id)")
    }

    private var pageZone: some View {
        Button(action: onOpenPage) {
            HStack(spacing: PVSpacing.space5) {
                CachedThumbnail(
                    projectDir: projectDir,
                    relPath: source.thumbnailRelPath,
                    typeIconKey: typeIconKey
                )
                VStack(alignment: .leading, spacing: PVSpacing.space1) {
                    Text(source.title)
                        .font(PVFont.body(size: PVTypeScale.bodySmall, weight: PVFontWeight.medium))
                        .foregroundStyle(PVColor.textPrimary)
                        .lineLimit(1)
                    if !typeLabel.isEmpty {
                        Text(typeLabel)
                            .font(PVFont.body(size: PVTypeScale.caption))
                            .foregroundStyle(PVColor.textMuted)
                            .lineLimit(1)
                    }
                }
                .frame(maxWidth: .infinity, alignment: .leading)
                Text(source.ref)
                    .font(PVFont.mono(size: PVTypeScale.micro))
                    .foregroundStyle(PVColor.textMuted)
                    .lineLimit(1)
            }
            .padding(.horizontal, PVSpacing.space6)
            .padding(.vertical, PVSpacing.space5 + PVSpacing.spacePx)
            .frame(maxWidth: .infinity, alignment: .leading)
            .contentShape(Rectangle())
        }
        .buttonStyle(SourcesSplitZoneButtonStyle())
        .accessibilityLabel(Text(L10n.Sources.openSourcePage))
        .accessibilityValue("\(source.title), \(source.ref)")
        .accessibilityIdentifier("sources.row.\(source.id).page")
    }

    @ViewBuilder
    private var graphZone: some View {
        if source.hasArtifact {
            Button(action: onOpenGraph) {
                HStack(spacing: PVSpacing.space3) {
                    PVIcon(.network, size: 14)
                    Text(L10n.Sources.openGraph)
                        .font(PVFont.body(size: PVTypeScale.caption, weight: PVFontWeight.medium))
                        .lineLimit(1)
                    Spacer(minLength: 0)
                }
                .foregroundStyle(PVColor.textPrimary)
                .padding(.horizontal, PVSpacing.space6)
                .padding(.vertical, PVSpacing.space5 + PVSpacing.spacePx)
                .frame(maxWidth: .infinity, maxHeight: .infinity, alignment: .leading)
                .contentShape(Rectangle())
            }
            .buttonStyle(SourcesSplitZoneButtonStyle())
            .accessibilityLabel(Text(L10n.Sources.openGraph))
            .accessibilityIdentifier("sources.row.\(source.id).graph")
        } else {
            HStack(spacing: PVSpacing.space3) {
                Text(L10n.Sources.needsArtifact)
                    .font(PVFont.body(size: PVTypeScale.caption))
                    .italic()
                    .foregroundStyle(PVColor.textFaint)
                    .lineLimit(1)
                Spacer(minLength: 0)
            }
            .padding(.horizontal, PVSpacing.space6)
            .padding(.vertical, PVSpacing.space5 + PVSpacing.spacePx)
            .frame(maxWidth: .infinity, maxHeight: .infinity, alignment: .leading)
            .background(PVColor.surfaceSunken)
            .help(String(localized: L10n.Sources.needsArtifactTooltip))
            .accessibilityElement(children: .ignore)
            .accessibilityLabel(Text(L10n.Sources.needsArtifact))
            .accessibilityHint(Text(L10n.Sources.needsArtifactTooltip))
            .accessibilityIdentifier("sources.row.\(source.id).graphBlocked")
        }
    }
}

private struct SourcesSplitZoneButtonStyle: ButtonStyle {
    @Environment(\.accessibilityReduceMotion) private var reduceMotion

    func makeBody(configuration: Configuration) -> some View {
        PVHoverEffect(isPressed: configuration.isPressed, hoverAnimation: PVMotion.fastStandard) { showHover in
            configuration.label
                .background(
                    showHover || configuration.isPressed
                        ? PVColor.surfaceHover
                        : Color.clear
                )
        }
        .animation(reduceMotion ? nil : PVMotion.fastStandard, value: configuration.isPressed)
    }
}
