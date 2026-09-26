import SwiftUI

/// Shared column track for the Sources list caption band and split rows (S5-D2 / S8-D5):
/// `minmax(0,1fr) | 248px`, so the hairline stacks straight down.
enum SourcesSplitLayout {
    static let graphZoneWidth: CGFloat = 248

    static func columns<Page: View, Graph: View>(
        @ViewBuilder page: () -> Page,
        @ViewBuilder graph: () -> Graph
    ) -> some View {
        HStack(spacing: 0) {
            page()
                .frame(maxWidth: .infinity, maxHeight: .infinity, alignment: .leading)
            graph()
                .frame(width: graphZoneWidth, alignment: .leading)
                .frame(maxHeight: .infinity)
                .overlay(alignment: .leading) {
                    Rectangle()
                        .fill(PVColor.borderSubtle)
                        .frame(width: 1)
                }
        }
        .frame(maxWidth: .infinity, alignment: .leading)
    }
}

/// Split-row Sources list cell: filing zone | Evidence graph zone (S5-D2).
struct SourcesSplitRow: View {
    let source: CatalogSource
    let typeLabel: String
    let typeIconKey: String?
    let projectDir: String
    var progress: SourceGraphProgress?
    var countsLoading: Bool = false
    var onOpenPage: () -> Void
    var onOpenGraph: () -> Void

    var body: some View {
        SourcesSplitLayout.columns {
            pageZone
        } graph: {
            graphZone
        }
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
            // Design: 13px 16px 13px 32px — page gutter on the leading edge.
            .padding(.leading, PVSpacing.gutterPage)
            .padding(.trailing, PVSpacing.space6)
            .padding(.vertical, PVSpacing.space5 + PVSpacing.spacePx)
            .frame(maxWidth: .infinity, maxHeight: .infinity, alignment: .leading)
            .contentShape(Rectangle())
        }
        .buttonStyle(SourcesSplitZoneButtonStyle(emphasizeOnHover: false))
        .accessibilityLabel(Text(L10n.Sources.openSourcePage))
        .accessibilityValue(Text(verbatim: "\(source.title), \(source.ref)"))
        .accessibilityIdentifier("sources.row.\(source.id).page")
    }

    @ViewBuilder
    private var graphZone: some View {
        if source.hasArtifact {
            Button(action: onOpenGraph) {
                HStack(alignment: .top, spacing: PVSpacing.space3 + PVSpacing.spacePx) {
                    PVIcon(.network, size: 16)
                    VStack(alignment: .leading, spacing: PVSpacing.space1) {
                        Text(L10n.Sources.openGraph)
                            .font(PVFont.body(size: PVTypeScale.bodySmall))
                            .lineLimit(1)
                        graphCountLine
                    }
                }
                .padding(.horizontal, PVSpacing.space6)
                .padding(.vertical, PVSpacing.space5 + PVSpacing.spacePx)
                .frame(maxWidth: .infinity, maxHeight: .infinity, alignment: .leading)
                .contentShape(Rectangle())
            }
            .buttonStyle(SourcesSplitZoneButtonStyle(emphasizeOnHover: true))
            .accessibilityLabel(Text(verbatim: L10n.Sources.graphZoneAccessibility(
                progress: progress,
                countsLoading: countsLoading
            )))
            .accessibilityIdentifier("sources.row.\(source.id).graph")
        } else {
            HStack(spacing: PVSpacing.space3 + PVSpacing.spacePx) {
                PVIcon(.network, size: 16)
                Text(L10n.Sources.needsArtifact)
                    .font(PVFont.body(size: PVTypeScale.caption))
                    .italic()
                    .foregroundStyle(PVColor.textMuted)
                    .lineLimit(1)
                Spacer(minLength: 0)
            }
            .foregroundStyle(PVColor.textFaint)
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

    @ViewBuilder
    private var graphCountLine: some View {
        if countsLoading && progress == nil {
            Rectangle()
                .fill(PVColor.borderSubtle)
                .frame(height: 1)
                .padding(.trailing, PVSpacing.space3)
                .padding(.vertical, 5)
                .accessibilityHidden(true)
        } else if (progress?.subjectCount ?? 0) == 0 {
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
        } else if let progress {
            Text(verbatim: L10n.Sources.graphCountLine(
                subjects: progress.subjectCount,
                observations: progress.observationCount
            ))
            .font(PVFont.mono(size: PVTypeScale.micro))
            .foregroundStyle(PVColor.textMuted)
            .lineLimit(1)
        }
    }
}

/// Hover fill for a split zone. Graph open zone also lifts label color on hover
/// (`text-secondary` → `text-primary`), matching the board.
private struct SourcesSplitZoneButtonStyle: ButtonStyle {
    var emphasizeOnHover: Bool
    @Environment(\.accessibilityReduceMotion) private var reduceMotion

    func makeBody(configuration: Configuration) -> some View {
        PVHoverEffect(isPressed: configuration.isPressed, hoverAnimation: PVMotion.fastStandard) { showHover in
            configuration.label
                .foregroundStyle(
                    emphasizeOnHover
                        ? (showHover || configuration.isPressed ? PVColor.textPrimary : PVColor.textMuted)
                        : PVColor.textPrimary
                )
                .background(
                    showHover || configuration.isPressed
                        ? PVColor.surfaceHover
                        : Color.clear
                )
        }
        .animation(reduceMotion ? nil : PVMotion.fastStandard, value: configuration.isPressed)
    }
}
