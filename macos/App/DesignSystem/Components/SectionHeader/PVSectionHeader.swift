import SwiftUI

/// Detail-page section chrome: display title, optional mono meta, leading
/// aside, trailing actions, and a bottom rule. Extracted from the Source
/// page during structural cleanup — not a direct port of a web-kit
/// component (the kit has no SectionHeader). Use on any detail surface
/// that stacks titled sections (Sources, later People / Facts / etc.).
/// Call sites own `.accessibilityIdentifier`.
struct PVSectionHeader<Aside: View, Actions: View>: View {
    private let title: LocalizedStringResource
    private let meta: String?
    private let aside: Aside
    private let actions: Actions

    init(
        title: LocalizedStringResource,
        meta: String? = nil,
        @ViewBuilder aside: () -> Aside = { EmptyView() },
        @ViewBuilder actions: () -> Actions = { EmptyView() }
    ) {
        self.title = title
        self.meta = meta
        self.aside = aside()
        self.actions = actions()
    }

    var body: some View {
        HStack(alignment: .firstTextBaseline, spacing: PVSpacing.space5) {
            Text(title)
                .font(PVFont.display(size: PVTypeScale.h3, weight: PVFontWeight.semibold))
                .foregroundStyle(PVColor.textDisplay)
            if let meta {
                Text(meta)
                    .font(PVFont.mono(size: PVTypeScale.caption))
                    .foregroundStyle(PVColor.textMuted)
            }
            aside
            Spacer(minLength: 0)
            actions
        }
        .padding(.bottom, PVSpacing.space4)
        .overlay(alignment: .bottom) {
            Rectangle()
                .fill(PVColor.borderDefault)
                .frame(height: 1)
        }
    }
}

#Preview("Title only") {
    PVSectionHeader(title: "Description")
        .padding(PVSpacing.space9)
        .background(PVColor.surfacePage)
}

#Preview("Meta + actions") {
    PVSectionHeader(title: "Artifacts", meta: "3 attached") {
        EmptyView()
    } actions: {
        PVButton("Add artifact", variant: .primary, size: .sm, icon: .plus) {}
    }
    .padding(PVSpacing.space9)
    .background(PVColor.surfacePage)
}

#Preview("Aside") {
    PVSectionHeader(title: "Credibility", aside: {
        Text("How much weight this record should carry.")
            .font(PVFont.body(size: PVTypeScale.caption, italic: true))
            .foregroundStyle(PVColor.textMuted)
            .lineLimit(2)
    })
    .padding(PVSpacing.space9)
    .background(PVColor.surfacePage)
}
