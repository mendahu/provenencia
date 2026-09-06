import SwiftUI

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
                        model.chooseFolder()
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
                        model.chooseFolder()
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
