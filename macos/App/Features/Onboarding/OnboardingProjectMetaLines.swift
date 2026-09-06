import SwiftUI

/// Shared project bookkeeping lines for open-picker preview and home.
struct OnboardingProjectMetaLines: View {
    let project: ProjectInfo
    var showFolder: Bool = true
    var accessibilityPrefix: String = "onboarding.project"

    init(
        _ project: ProjectInfo,
        showFolder: Bool = true,
        accessibilityPrefix: String = "onboarding.project"
    ) {
        self.project = project
        self.showFolder = showFolder
        self.accessibilityPrefix = accessibilityPrefix
    }

    var body: some View {
        VStack(alignment: .leading, spacing: PVSpacing.space2) {
            if !project.label.isEmpty {
                Text(project.label)
                    .font(PVFont.body(size: PVTypeScale.h4, weight: PVFontWeight.semibold))
                    .foregroundStyle(PVColor.textPrimary)
                    .accessibilityIdentifier("\(accessibilityPrefix).projectLabel")
            }
            if showFolder, !project.folderName.isEmpty {
                Text(L10n.Onboarding.homeFolder(folderName: project.folderName))
                    .font(PVFont.mono(size: PVTypeScale.caption))
                    .foregroundStyle(PVColor.textMuted)
                    .accessibilityIdentifier("\(accessibilityPrefix).project")
            }
            if !project.createdAt.isEmpty {
                Text(L10n.Onboarding.homeCreated(date: Self.formatDate(project.createdAt)))
                    .font(PVFont.mono(size: PVTypeScale.caption))
                    .foregroundStyle(PVColor.textMuted)
                    .accessibilityIdentifier("\(accessibilityPrefix).created")
            }
            if !project.updatedAt.isEmpty {
                Text(L10n.Onboarding.homeUpdated(date: Self.formatDate(project.updatedAt)))
                    .font(PVFont.mono(size: PVTypeScale.caption))
                    .foregroundStyle(PVColor.textMuted)
                    .accessibilityIdentifier("\(accessibilityPrefix).updated")
            }
            if !project.updatedByDisplayName.isEmpty {
                Text(L10n.Onboarding.homeUpdatedBy(
                    displayName: project.updatedByDisplayName,
                    ref: project.updatedByRef
                ))
                    .font(PVFont.mono(size: PVTypeScale.caption))
                    .foregroundStyle(PVColor.textMuted)
                    .accessibilityIdentifier("\(accessibilityPrefix).updatedBy")
            }
        }
        .frame(maxWidth: .infinity, alignment: .leading)
        .padding(PVSpacing.space5)
        .background(
            RoundedRectangle(cornerRadius: PVRadius.md, style: .continuous)
                .fill(PVColor.surfaceCard)
        )
        .overlay(
            RoundedRectangle(cornerRadius: PVRadius.md, style: .continuous)
                .strokeBorder(PVColor.borderSubtle, lineWidth: 1)
        )
    }

    private static func formatDate(_ rfc3339: String) -> String {
        let parser = ISO8601DateFormatter()
        parser.formatOptions = [.withInternetDateTime, .withFractionalSeconds]
        if let date = parser.date(from: rfc3339) ?? ISO8601DateFormatter().date(from: rfc3339) {
            return date.formatted(date: .abbreviated, time: .shortened)
        }
        return rfc3339
    }
}
