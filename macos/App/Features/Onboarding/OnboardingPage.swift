import SwiftUI

/// Shared choose-file / identify page shell: form column, padding, footer slot.
struct OnboardingPage<Content: View, Footer: View>: View {
    @ViewBuilder var content: Content
    @ViewBuilder var footer: Footer

    var body: some View {
        VStack(alignment: .leading, spacing: 0) {
            VStack(alignment: .leading, spacing: PVSpacing.space7) {
                content
            }
            Spacer(minLength: PVSpacing.space8)
            footer
        }
        .padding(PVSpacing.space9)
        .frame(maxWidth: PVSpacing.measureForm, alignment: .leading)
        .frame(maxWidth: .infinity, maxHeight: .infinity, alignment: .center)
    }
}

/// Title + supporting body used at the top of onboarding form pages.
struct OnboardingHeader: View {
    let title: LocalizedStringResource
    private let bodyText: String

    init(_ title: LocalizedStringResource, body: String) {
        self.title = title
        self.bodyText = body
    }

    init(_ title: LocalizedStringResource, body: LocalizedStringResource) {
        self.title = title
        self.bodyText = String(localized: body)
    }

    var body: some View {
        Group {
            Text(title)
                .font(PVFont.display(size: PVTypeScale.h2))
                .foregroundStyle(PVColor.textDisplay)
            Text(bodyText)
                .font(PVFont.body(size: PVTypeScale.bodySmall))
                .foregroundStyle(PVColor.textSecondary)
                .fixedSize(horizontal: false, vertical: true)
        }
    }
}
