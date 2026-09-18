import SwiftUI

/// One crumb in a `PVBreadcrumbs` trail. The last item is usually the current
/// page (non-interactive); earlier items are links.
struct PVBreadcrumbItem: Identifiable, Equatable {
    let id: String
    let label: String
    /// When non-nil, the crumb is a tappable link.
    var action: (() -> Void)?

    static func == (lhs: PVBreadcrumbItem, rhs: PVBreadcrumbItem) -> Bool {
        lhs.id == rhs.id && lhs.label == rhs.label && (lhs.action == nil) == (rhs.action == nil)
    }
}

/// Path trail for leaf destinations (e.g. Source page → Sources list).
/// Mirrors the design board breadcrumb: no standalone back button.
struct PVBreadcrumbs: View {
    let items: [PVBreadcrumbItem]

    var body: some View {
        HStack(spacing: PVSpacing.space3) {
            ForEach(Array(items.enumerated()), id: \.element.id) { index, item in
                if index > 0 {
                    Text("/")
                        .font(PVFont.body(size: PVTypeScale.caption))
                        .foregroundStyle(PVColor.textFaint)
                        .accessibilityHidden(true)
                }
                if let action = item.action {
                    Button(action: action) {
                        Text(item.label)
                            .font(PVFont.body(size: PVTypeScale.caption, weight: PVFontWeight.medium))
                            .foregroundStyle(PVColor.textSecondary)
                    }
                    .buttonStyle(.plain)
                    .onHover { hovering in
                        if hovering {
                            NSCursor.pointingHand.push()
                        } else {
                            NSCursor.pop()
                        }
                    }
                } else {
                    Text(item.label)
                        .font(PVFont.mono(size: PVTypeScale.caption))
                        .foregroundStyle(PVColor.textMuted)
                }
            }
        }
        .accessibilityElement(children: .contain)
    }
}

#Preview {
    PVBreadcrumbs(items: [
        PVBreadcrumbItem(id: "sources", label: "Sources", action: {}),
        PVBreadcrumbItem(id: "ref", label: "SRC-0412", action: nil),
    ])
    .padding(PVSpacing.space9)
    .background(PVColor.surfacePage)
}
