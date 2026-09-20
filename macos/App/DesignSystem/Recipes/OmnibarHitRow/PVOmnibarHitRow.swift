import SwiftUI

/// Shared omnibar result row skeleton — one layout for every hit kind.
/// Callers fill slots via a lead view + copy; the row is not kind-aware.
///
/// Board: Spike 3 Omnibar Results / Omnibar Hit Row → `PVOmnibarHitRow`.
struct PVOmnibarHitRow<Lead: View>: View {
    let lead: Lead
    let title: String
    let kindLabel: LocalizedStringResource
    let secondary: String
    let ref: String
    /// When true, mono ref uses accent (exact catalog-ref match).
    var refAccent: Bool = false
    let matchContext: String
    var selected: Bool = false

    init(
        @ViewBuilder lead: () -> Lead,
        title: String,
        kindLabel: LocalizedStringResource,
        secondary: String,
        ref: String,
        refAccent: Bool = false,
        matchContext: String,
        selected: Bool = false
    ) {
        self.lead = lead()
        self.title = title
        self.kindLabel = kindLabel
        self.secondary = secondary
        self.ref = ref
        self.refAccent = refAccent
        self.matchContext = matchContext
        self.selected = selected
    }

    var body: some View {
        HStack(alignment: .top, spacing: PVSpacing.space5) {
            lead
                .frame(width: 40, height: 40)
                .accessibilityHidden(true)

            VStack(alignment: .leading, spacing: 2) {
                HStack(alignment: .firstTextBaseline, spacing: PVSpacing.space4) {
                    Text(title)
                        .font(PVFont.body(size: PVTypeScale.bodySmall, weight: PVFontWeight.semibold))
                        .foregroundStyle(PVColor.textPrimary)
                        .lineLimit(1)
                        .truncationMode(.tail)
                        .frame(maxWidth: .infinity, alignment: .leading)
                    PVBadge(kindLabel, tone: .neutral, subtle: true)
                }

                HStack(alignment: .firstTextBaseline, spacing: PVSpacing.space4) {
                    if !secondary.isEmpty {
                        Text(secondary)
                            .font(PVFont.body(size: PVTypeScale.caption))
                            .foregroundStyle(PVColor.textMuted)
                            .lineLimit(1)
                            .truncationMode(.tail)
                            .frame(maxWidth: .infinity, alignment: .leading)
                    } else {
                        Spacer(minLength: 0)
                    }
                    if !ref.isEmpty {
                        Text(ref)
                            .font(PVFont.mono(size: PVTypeScale.micro))
                            .foregroundStyle(refAccent ? PVColor.accent : PVColor.textFaint)
                            .fixedSize(horizontal: true, vertical: false)
                    }
                }

                if !matchContext.isEmpty {
                    Text(matchContext)
                        .font(PVFont.body(size: PVTypeScale.caption))
                        .foregroundStyle(PVColor.textFaint)
                        .lineLimit(1)
                        .truncationMode(.tail)
                }
            }
        }
        .padding(.horizontal, 9)
        .padding(.vertical, 7)
        .frame(maxWidth: .infinity, alignment: .leading)
        .contentShape(Rectangle())
        .background(selected ? PVColor.surfaceHover : Color.clear)
        .clipShape(RoundedRectangle(cornerRadius: PVRadius.xs, style: .continuous))
        .accessibilityElement(children: .combine)
    }
}

#Preview {
    VStack(spacing: 0) {
        PVOmnibarHitRow(
            lead: { PVThumbnail(PVThumbnail.Content(icon: .scrollText), size: 40) },
            title: "Ilminster parish register, 1841–1852",
            kindLabel: L10n.Workspace.omnibarKindSource,
            secondary: "Parish register",
            ref: "SRC-3K9M2",
            refAccent: true,
            matchContext: "Exact match — opens on return",
            selected: true
        )
        PVOmnibarHitRow(
            lead: { PVThumbnail(PVThumbnail.Content(icon: .library), size: 40) },
            title: "Parish register",
            kindLabel: L10n.Workspace.omnibarKindType,
            secondary: "parish-register",
            ref: "",
            matchContext: "",
            selected: false
        )
    }
    .padding(PVSpacing.space2)
    .frame(width: 640)
    .background(PVColor.surfaceCard)
}
