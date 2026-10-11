import SwiftUI

/// Zoomed-out evidence card, framed to the full card's rectangle.
///
/// Shared by primary cards and bridges so the treatment is not copied twice.
/// It does not report a layout: the frame is the last measured card size, and
/// a new measurement would move the edges.
struct EvidenceGraphOverviewFace: View {
    var size: CGSize
    var label: String
    var labelSize: CGFloat
    var horizontalPadding: CGFloat
    var columnSpacing: CGFloat
    var fill: Color
    /// Kind wash laid over `fill` for an uncited primary. Bridges pass nil.
    var wash: Color?
    var borderColor: Color
    /// Document-point width of the card stroke. Selection keeps its own width.
    var borderLineWidth: CGFloat = 1
    var chip: Chip?
    var showsSelection: Bool
    var isDragging: Bool

    struct Chip: Equatable {
        var fill: Color
        var border: Color
        var mark: PVMarkKey
        var ink: Color
    }

    var body: some View {
        VStack(spacing: columnSpacing) {
            if let chip {
                ZStack {
                    RoundedRectangle(cornerRadius: 10, style: .continuous)
                        .fill(chip.fill)
                    RoundedRectangle(cornerRadius: 10, style: .continuous)
                        .strokeBorder(chip.border, lineWidth: 1)
                    PVMark(chip.mark, size: 46)
                        .foregroundStyle(chip.ink)
                }
                .frame(width: 72, height: 72)
            }
            Text(verbatim: label)
                .font(PVFont.display(size: labelSize, weight: PVFontWeight.medium))
                .foregroundStyle(PVColor.textPrimary)
                .lineLimit(1)
                .truncationMode(.tail)
                .multilineTextAlignment(.center)
                .frame(maxWidth: .infinity)
                .help(Text(verbatim: label))
        }
        .padding(.horizontal, horizontalPadding)
        .frame(width: size.width, height: size.height)
        .background {
            RoundedRectangle(cornerRadius: PVRadius.md, style: .continuous)
                .fill(fill)
                .overlay {
                    if let wash {
                        RoundedRectangle(cornerRadius: PVRadius.md, style: .continuous)
                            .fill(wash)
                    }
                }
        }
        .overlay {
            RoundedRectangle(cornerRadius: PVRadius.md, style: .continuous)
                .strokeBorder(borderColor, lineWidth: showsSelection ? 1.5 : borderLineWidth)
        }
        .clipShape(RoundedRectangle(cornerRadius: PVRadius.md, style: .continuous))
        .shadow(
            color: Color.black.opacity(isDragging ? 0.1 : 0.06),
            radius: isDragging ? 10 : 6,
            y: isDragging ? 4 : 2
        )
        .background { selectionRings }
        .accessibilityHidden(true)
    }

    /// 4pt accent outline inside a 16pt ring, in document points, so the
    /// selection still reads at the zoom floor. The reading face keeps its
    /// own 3pt ring.
    @ViewBuilder
    private var selectionRings: some View {
        if showsSelection {
            ZStack {
                RoundedRectangle(cornerRadius: PVRadius.md + 16, style: .continuous)
                    .fill(PVColor.graphRing)
                    .padding(-16)
                RoundedRectangle(cornerRadius: PVRadius.md + 4, style: .continuous)
                    .fill(PVColor.accent)
                    .padding(-4)
            }
        }
    }

    static func subject(
        placed: SourceGraphPlacedSubject,
        presentation: CatalogSubjectTypePresentation?,
        size: CGSize,
        showsSelection: Bool,
        isDragging: Bool
    ) -> EvidenceGraphOverviewFace {
        let style = EvidenceSubjectKindStyle.resolve(
            typeKey: placed.kind.rawValue,
            presentation: presentation
        )
        return EvidenceGraphOverviewFace(
            size: size,
            label: subjectLabel(placed),
            labelSize: 28,
            horizontalPadding: 20,
            columnSpacing: 16,
            fill: placed.isCited ? style.tint : PVColor.surfaceCard,
            wash: placed.isCited ? nil : style.tint.opacity(0.38),
            borderColor: showsSelection ? PVColor.accent : style.line,
            chip: Chip(fill: style.chip, border: style.line, mark: placed.kind.markKey, ink: style.ink),
            showsSelection: showsSelection,
            isDragging: isDragging
        )
    }

    static func bridge(
        summary: String,
        size: CGSize,
        showsSelection: Bool,
        isDragging: Bool
    ) -> EvidenceGraphOverviewFace {
        EvidenceGraphOverviewFace(
            size: size,
            label: summary,
            labelSize: 22,
            horizontalPadding: 16,
            columnSpacing: 12,
            fill: PVColor.surfaceCard,
            wash: nil,
            borderColor: showsSelection ? PVColor.accent : PVColor.borderStrong,
            borderLineWidth: bridgeBorderWidth,
            chip: nil,
            showsSelection: showsSelection,
            isDragging: isDragging
        )
    }

    /// Document points. At 0.25× this is still about a point on screen.
    private static let bridgeBorderWidth: CGFloat = 3

    private static func subjectLabel(_ placed: SourceGraphPlacedSubject) -> String {
        let trimmed = placed.subject.label.trimmingCharacters(in: .whitespacesAndNewlines)
        return trimmed.isEmpty ? placed.typeLabel : trimmed
    }
}
