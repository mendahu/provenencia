import SwiftUI

/// Footer band treatment for ``PVPanel``.
///
/// - ``sunken``: `surfaceSunken` fill + top hairline — form/confirm sheets.
/// - ``plain``: top hairline only — when the footer should not sink.
enum PVPanelFooterChrome {
    case sunken
    case plain
}

/// Content-agnostic sheet **panel** primitive: title, optional subtitle, body
/// slot, optional footer. Draws **no** panel background, corner radius, shadow,
/// or scrim — the sheet window owns those. Composites (``PVFormDialogContent``,
/// confirm sheet) and one-off feature sheets compose this instead of
/// reimplementing header/footer chrome.
struct PVPanel<Body: View, Footer: View>: View {
    let title: Text
    let subtitle: Text?
    let width: CGFloat
    let footerChrome: PVPanelFooterChrome
    @ViewBuilder let panelBody: () -> Body
    @ViewBuilder let footer: () -> Footer

    /// Whether the footer band is rendered. `false` for the no-footer initializer.
    let showsFooter: Bool

    init(
        title: Text,
        subtitle: Text? = nil,
        width: CGFloat,
        footerChrome: PVPanelFooterChrome = .sunken,
        @ViewBuilder body: @escaping () -> Body,
        @ViewBuilder footer: @escaping () -> Footer
    ) {
        self.title = title
        self.subtitle = subtitle
        self.width = width
        self.footerChrome = footerChrome
        self.panelBody = body
        self.footer = footer
        self.showsFooter = true
    }

    var body: some View {
        VStack(spacing: 0) {
            headerAndBody
            if showsFooter {
                footerBand
            }
        }
        .frame(width: width)
    }

    private var headerAndBody: some View {
        VStack(alignment: .leading, spacing: PVSpacing.space7) {
            VStack(alignment: .leading, spacing: PVSpacing.space2) {
                title
                    .font(PVFont.display(size: PVTypeScale.h3))
                    .foregroundStyle(PVColor.textDisplay)
                    .fixedSize(horizontal: false, vertical: true)
                if let subtitle {
                    subtitle
                        .font(PVFont.body(size: PVTypeScale.bodySmall))
                        .foregroundStyle(PVColor.textSecondary)
                        .lineSpacing((PVLineHeight.relaxed - 1) * PVTypeScale.bodySmall)
                        .fixedSize(horizontal: false, vertical: true)
                }
            }
            panelBody()
        }
        .frame(maxWidth: .infinity, alignment: .leading)
        .padding(.horizontal, PVSpacing.space7)
        .padding(.vertical, PVSpacing.space9)
    }

    private var footerBand: some View {
        footer()
            .padding(.horizontal, PVSpacing.space8)
            .padding(.vertical, PVSpacing.space8)
            .frame(maxWidth: .infinity)
            .background(footerChrome == .sunken ? PVColor.surfaceSunken : Color.clear)
            .overlay(alignment: .top) {
                PVDivider()
            }
    }
}

extension PVPanel where Footer == EmptyView {
    /// Panel with no footer band.
    init(
        title: Text,
        subtitle: Text? = nil,
        width: CGFloat,
        footerChrome: PVPanelFooterChrome = .sunken,
        @ViewBuilder body: @escaping () -> Body
    ) {
        self.title = title
        self.subtitle = subtitle
        self.width = width
        self.footerChrome = footerChrome
        self.panelBody = body
        self.footer = { EmptyView() }
        self.showsFooter = false
    }
}

#Preview("Panel — sunken footer") {
    PVPanel(
        title: Text(verbatim: "Add source"),
        subtitle: Text(verbatim: "A thin record now — artifacts live on the Source page."),
        width: 480,
        footerChrome: .sunken
    ) {
        Text(verbatim: "Form body")
            .font(PVFont.body(size: PVTypeScale.bodySmall))
            .foregroundStyle(PVColor.textSecondary)
    } footer: {
        HStack {
            Spacer(minLength: PVSpacing.space8)
            Text(verbatim: "Cancel")
            Text(verbatim: "Create")
        }
    }
    .background(PVColor.surfaceCard)
}
