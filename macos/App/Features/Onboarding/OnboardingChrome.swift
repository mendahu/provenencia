import AppKit
import SwiftUI

struct OnboardingFileChoice: View {
    let title: LocalizedStringResource
    let subtitle: LocalizedStringResource
    let icon: String
    let selected: Bool
    let action: () -> Void

    init(
        _ title: LocalizedStringResource,
        subtitle: LocalizedStringResource,
        icon: String,
        selected: Bool,
        action: @escaping () -> Void
    ) {
        self.title = title
        self.subtitle = subtitle
        self.icon = icon
        self.selected = selected
        self.action = action
    }

    var body: some View {
        Button(action: action) {
            VStack(alignment: .leading, spacing: PVSpacing.space4) {
                HStack(spacing: PVSpacing.space4) {
                    PVIcon(icon, size: 16)
                    titleText
                    if selected {
                        Spacer(minLength: 0)
                        checkIcon
                    }
                }
                Text(subtitle)
                    .font(PVFont.body(size: PVTypeScale.bodySmall))
                    .foregroundStyle(selected ? PVColor.accentSoftForeground : PVColor.textSecondary)
                    .multilineTextAlignment(.leading)
            }
            .frame(maxWidth: .infinity, alignment: .leading)
            .padding(PVSpacing.space6)
            .background(
                RoundedRectangle(cornerRadius: PVRadius.md, style: .continuous)
                    .fill(selected ? PVColor.surfaceSelected : PVColor.surfaceCard)
            )
            .overlay(
                RoundedRectangle(cornerRadius: PVRadius.md, style: .continuous)
                    .strokeBorder(selected ? PVColor.accent : PVColor.borderSubtle, lineWidth: selected ? 1.5 : 1)
            )
            .pvShadow(selected ? PVElevation.sm : [])
        }
        .buttonStyle(.plain)
    }

    private var titleText: some View {
        Text(title)
            .font(PVFont.body(size: PVTypeScale.h4, weight: PVFontWeight.semibold))
            .foregroundStyle(selected ? PVColor.accentSoftForeground : PVColor.textPrimary)
    }

    private var checkIcon: some View {
        PVIcon("check", size: 15)
            .foregroundStyle(PVColor.accent)
    }
}

struct OnboardingOpenPicker: View {
    @Bindable var model: OnboardingModel

    private var selectedProjectPath: Binding<String> {
        Binding(
            get: { model.selectedProject?.path ?? "" },
            set: { model.selectedProject = $0.isEmpty ? nil : URL(fileURLWithPath: $0) }
        )
    }

    private var projectOptions: [PVSelectOption] {
        [PVSelectOption(value: "", label: String(localized: L10n.Onboarding.selectProject))]
            + model.availableProjects.map { PVSelectOption(value: $0.path, label: $0.lastPathComponent) }
    }

    var body: some View {
        VStack(alignment: .leading, spacing: PVSpacing.space6) {
            Text(L10n.Onboarding.projectFolder)
                .font(PVFont.body(size: PVTypeScale.micro, weight: PVFontWeight.semibold))
                .tracking(PVTypeScale.micro * PVTracking.caps)
                .textCase(.uppercase)
                .foregroundStyle(PVColor.textMuted)
            if model.availableProjects.isEmpty {
                HStack(alignment: .center, spacing: PVSpacing.space5) {
                    Text(L10n.Onboarding.noProjectsInDocuments)
                        .font(PVFont.body(size: PVTypeScale.bodySmall))
                        .foregroundStyle(PVColor.textMuted)
                    Spacer(minLength: 0)
                    PVButton(L10n.Onboarding.chooseFolder, variant: .secondary, size: .sm) {
                        onboardingChooseFolder(model)
                    }
                    .accessibilityIdentifier("onboarding.chooseFolder")
                }
                .padding(PVSpacing.space5)
                .background(
                    RoundedRectangle(cornerRadius: PVRadius.md, style: .continuous)
                        .fill(PVColor.surfaceSunken)
                )
                .overlay(
                    RoundedRectangle(cornerRadius: PVRadius.md, style: .continuous)
                        .strokeBorder(PVColor.borderDefault, style: StrokeStyle(lineWidth: 1, dash: [4, 3]))
                )
            } else {
                HStack(alignment: .center, spacing: PVSpacing.space5) {
                    PVSelect(selection: selectedProjectPath, options: projectOptions)
                        .accessibilityIdentifier("onboarding.existingProject")
                    PVButton(L10n.Onboarding.chooseFolder, variant: .secondary, size: .sm) {
                        onboardingChooseFolder(model)
                    }
                    .accessibilityIdentifier("onboarding.chooseFolder")
                }
            }
            if model.selectedProject != nil, let project = model.project {
                OnboardingProjectMetaLines(
                    project,
                    showFolder: false,
                    accessibilityPrefix: "onboarding.open"
                )
                .padding(.top, PVSpacing.space2)
            }
        }
        .onChange(of: model.selectedProject) { _, _ in
            Task { await model.refreshSelectedProjectInfo() }
        }
    }
}

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

struct OnboardingFooter<Leading: View, Trailing: View>: View {
    @Bindable var model: OnboardingModel
    @ViewBuilder var leading: Leading
    @ViewBuilder var trailing: Trailing

    var body: some View {
        HStack(spacing: PVSpacing.space5) {
            leading
            Spacer()
            if model.isBusy {
                ProgressView()
                    .controlSize(.small)
                    .tint(PVColor.accent)
            }
            trailing
        }
    }
}

/// Whether a divider is drawn above an `OnboardingContributorRow`, and in
/// what style — solid between ordinary rows, dashed above the trailing
/// "not listed" row (the design system's "absence of evidence" convention).
enum OnboardingContributorRowDivider {
    case none, solid, dashed
}

/// A single selectable row in the "who are you in this file?" contributor
/// list. Bespoke to Onboarding, not a shared `DesignSystem/Components` file
/// — the source design system's own mockup hand-styles these rows as
/// page-specific markup too (it doesn't reach for its own `Radio`
/// component here), and `DesignSystem/README.md` lists `Radio` as
/// "add on demand." Only reaches for `PV*` tokens.
struct OnboardingContributorRow: View {
    let title: String
    var subtitle: String?
    let selected: Bool
    var divider: OnboardingContributorRowDivider = .none
    let action: () -> Void

    var body: some View {
        Button(action: action) {
            HStack(spacing: PVSpacing.space5) {
                Circle()
                    .strokeBorder(selected ? PVColor.accent : PVColor.borderStrong, lineWidth: selected ? 4 : 1)
                    .background(Circle().fill(selected ? PVColor.surfaceCard : PVColor.surfaceRaised))
                    .frame(width: 13, height: 13)
                Text(title)
                    .font(PVFont.body(size: PVTypeScale.bodySmall, weight: selected ? PVFontWeight.semibold : PVFontWeight.regular))
                    .foregroundStyle(PVColor.textPrimary)
                if let subtitle {
                    Text(subtitle)
                        .font(PVFont.mono(size: PVTypeScale.micro))
                        .foregroundStyle(PVColor.textMuted)
                }
                Spacer(minLength: 0)
            }
            .padding(.horizontal, PVSpacing.space5)
            .padding(.vertical, PVSpacing.space4)
            .frame(maxWidth: .infinity, alignment: .leading)
            .background(selected ? PVColor.surfaceSelected : Color.clear)
            .overlay(alignment: .leading) {
                if selected {
                    Rectangle().fill(PVColor.accent).frame(width: 2)
                }
            }
            .overlay(alignment: .top) {
                switch divider {
                case .none:
                    EmptyView()
                case .solid:
                    Rectangle().fill(PVColor.borderSubtle).frame(height: 1)
                case .dashed:
                    OnboardingHorizontalLine()
                        .stroke(PVColor.borderDefault, style: StrokeStyle(lineWidth: 1, dash: [4, 3]))
                        .frame(height: 1)
                }
            }
        }
        .buttonStyle(.plain)
    }
}

/// A single horizontal line spanning its frame's width — used with a dashed
/// `StrokeStyle` for the "not listed" row's divider (dashed = absence of
/// evidence, per the design system's border-weight convention).
private struct OnboardingHorizontalLine: Shape {
    func path(in rect: CGRect) -> Path {
        var path = Path()
        path.move(to: CGPoint(x: rect.minX, y: rect.midY))
        path.addLine(to: CGPoint(x: rect.maxX, y: rect.midY))
        return path
    }
}

@MainActor
func onboardingChooseFolder(_ model: OnboardingModel) {
    let panel = NSOpenPanel()
    panel.canChooseFiles = false
    panel.canChooseDirectories = true
    panel.allowsMultipleSelection = false
    panel.treatsFilePackagesAsDirectories = true
    panel.prompt = String(localized: L10n.Onboarding.openPanelPrompt)
    panel.message = String(localized: L10n.Onboarding.openPanelMessage)
    guard panel.runModal() == .OK, let url = panel.url else {
        return
    }
    Task { await model.choose(url) }
}
