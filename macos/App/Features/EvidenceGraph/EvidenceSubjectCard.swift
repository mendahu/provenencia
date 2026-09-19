import SwiftUI

/// Drag-enabled primary subject card (S6-03).
struct EvidenceSubjectCard: View {
    let placed: SourceGraphPlacedSubject
    var isSelected: Bool
    var isDragging: Bool = false
    var dragEnabled: Bool = true
    var onSelect: () -> Void
    var onDragChanged: ((CGSize) -> Void)?
    var onDragEnded: (() -> Void)?

    private var style: EvidenceSubjectKindStyle {
        EvidenceSubjectKindStyle.forKind(placed.kind)
    }

    private var iconKind: PVSubjectIconKind {
        switch placed.kind {
        case .person: .person
        case .event: .event
        case .place: .place
        }
    }

    var body: some View {
        cardChrome
            .opacity(ghostOpacity)
            .gesture(dragGesture)
            .onTapGesture(perform: onSelect)
            .accessibilityHidden(true)
    }

    /// Non-interactive placement preview while a tool is armed.
    static func ghost(placed: SourceGraphPlacedSubject) -> some View {
        EvidenceSubjectCard(
            placed: placed,
            isSelected: false,
            isDragging: false,
            dragEnabled: false,
            onSelect: {}
        )
        .opacity(0.62)
        .allowsHitTesting(false)
    }

    private var ghostOpacity: Double {
        if isDragging { return 0.92 }
        return placed.isCited ? 1 : 0.76
    }

    private var cardChrome: some View {
        VStack(alignment: .leading, spacing: 10) {
            HStack(alignment: .top, spacing: 10) {
                iconChip
                VStack(alignment: .leading, spacing: 2) {
                    Text(verbatim: displayLabel)
                        .font(PVFont.display(size: 16, weight: PVFontWeight.medium))
                        .foregroundStyle(placed.isCited ? PVColor.textDisplay : PVColor.textSecondary)
                        .lineLimit(2)
                        .multilineTextAlignment(.leading)
                    Text(verbatim: placed.typeLabel)
                        .font(PVFont.mono(size: 10, weight: PVFontWeight.medium))
                        .tracking(1)
                        .textCase(.uppercase)
                        .foregroundStyle(style.ink)
                }
                .frame(maxWidth: .infinity, alignment: .leading)
                citationMark
            }
            if let description = nonEmptyDescription {
                Text(verbatim: description)
                    .font(PVFont.body(size: PVTypeScale.bodySmall))
                    .foregroundStyle(PVColor.textMuted)
                    .lineLimit(3)
                    .multilineTextAlignment(.leading)
            }
        }
        .padding(.horizontal, 13)
        .padding(.vertical, 12)
        .frame(width: 236, alignment: .leading)
        .background(cardBackground)
        .overlay(cardBorder)
        .clipShape(RoundedRectangle(cornerRadius: PVRadius.md, style: .continuous))
        .shadow(
            color: (placed.isCited || isDragging) && !isSelected
                ? Color.black.opacity(0.1)
                : .clear,
            radius: isDragging ? 10 : 6,
            y: isDragging ? 4 : 2
        )
        .background(selectionHalo)
    }

    private var dragGesture: some Gesture {
        DragGesture(minimumDistance: dragEnabled ? 4 : 10_000)
            .onChanged { value in
                guard dragEnabled else { return }
                onDragChanged?(value.translation)
            }
            .onEnded { _ in
                guard dragEnabled else { return }
                onDragEnded?()
            }
    }

    private var displayLabel: String {
        let trimmed = placed.subject.label.trimmingCharacters(in: .whitespacesAndNewlines)
        return trimmed.isEmpty ? placed.typeLabel : trimmed
    }

    private var nonEmptyDescription: String? {
        let trimmed = placed.subject.description.trimmingCharacters(in: .whitespacesAndNewlines)
        return trimmed.isEmpty ? nil : trimmed
    }

    private var iconChip: some View {
        ZStack {
            RoundedRectangle(cornerRadius: 4, style: .continuous)
                .fill(style.chip)
            RoundedRectangle(cornerRadius: 4, style: .continuous)
                .strokeBorder(
                    placed.isCited ? style.line : PVColor.borderDefault,
                    style: StrokeStyle(
                        lineWidth: 1,
                        dash: placed.isCited ? [] : [3, 2]
                    )
                )
            PVSubjectIcon(kind: iconKind, size: 15)
                .foregroundStyle(style.ink)
        }
        .frame(width: 28, height: 28)
    }

    @ViewBuilder
    private var citationMark: some View {
        if placed.isCited {
            Image(systemName: "checkmark")
                .font(.system(size: 11, weight: .semibold))
                .foregroundStyle(style.ink)
                .accessibilityHidden(true)
        } else {
            Image(systemName: "circle.dashed")
                .font(.system(size: 13, weight: .regular))
                .foregroundStyle(PVColor.evidenceUndocumented)
                .accessibilityHidden(true)
        }
    }

    private var cardBackground: some View {
        RoundedRectangle(cornerRadius: PVRadius.md, style: .continuous)
            .fill(PVColor.surfaceCard)
            .overlay(
                RoundedRectangle(cornerRadius: PVRadius.md, style: .continuous)
                    .fill(style.tint.opacity(placed.isCited ? 0.45 : 0.38))
            )
    }

    private var cardBorder: some View {
        RoundedRectangle(cornerRadius: PVRadius.md, style: .continuous)
            .strokeBorder(
                borderColor,
                style: StrokeStyle(
                    lineWidth: (isSelected || isDragging) ? 1.5 : 1,
                    dash: placed.isCited ? [] : [4, 3]
                )
            )
    }

    private var borderColor: Color {
        if isSelected || isDragging { return PVColor.accent }
        return placed.isCited ? style.line : PVColor.borderDefault
    }

    @ViewBuilder
    private var selectionHalo: some View {
        if isSelected || isDragging {
            RoundedRectangle(cornerRadius: PVRadius.md + 2, style: .continuous)
                .fill(PVColor.graphRing)
                .padding(-3)
        }
    }
}

extension EvidenceSubjectCard {
    /// VoiceOver label for the accessibility representation (not the drawn card).
    static func accessibilityLabel(for placed: SourceGraphPlacedSubject) -> String {
        let name = {
            let trimmed = placed.subject.label.trimmingCharacters(in: .whitespacesAndNewlines)
            return trimmed.isEmpty ? placed.typeLabel : trimmed
        }()
        let citation = placed.isCited
            ? String(localized: L10n.EvidenceGraph.citedAccessibility)
            : String(localized: L10n.EvidenceGraph.uncitedAccessibility)
        return "\(placed.typeLabel), \(name), \(citation)"
    }
}
