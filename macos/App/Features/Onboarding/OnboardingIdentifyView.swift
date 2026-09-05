import SwiftUI

struct OnboardingIdentifyView: View {
    @Bindable var model: OnboardingModel

    var body: some View {
        VStack(alignment: .leading, spacing: 0) {
            VStack(alignment: .leading, spacing: PVSpacing.space7) {
                if model.mode == .create {
                    Text(L10n.Onboarding.nameResearchTitle)
                        .font(PVFont.display(size: PVTypeScale.h2))
                        .foregroundStyle(PVColor.textDisplay)
                    Text(L10n.Onboarding.nameResearchBody)
                        .font(PVFont.body(size: PVTypeScale.bodySmall))
                        .foregroundStyle(PVColor.textSecondary)
                    PVField(label: L10n.Onboarding.researcherName) {
                        PVInput(
                            text: $model.displayName,
                            isReadOnly: model.researcherLocked,
                            prompt: L10n.Onboarding.researcherNamePrompt
                        )
                        .accessibilityIdentifier("onboarding.displayName")
                    }
                    VStack(alignment: .leading, spacing: PVSpacing.space3) {
                        PVField(label: L10n.Onboarding.familyName) {
                            PVInput(
                                text: $model.familyName,
                                prompt: L10n.Onboarding.familyNamePrompt
                            )
                            .accessibilityIdentifier("onboarding.familyName")
                        }
                        if !model.folderNamePreview.isEmpty {
                            Text(L10n.Onboarding.folderNamePreview(folderName: model.folderNamePreview))
                                .font(PVFont.mono(size: PVTypeScale.caption))
                                .foregroundStyle(PVColor.textMuted)
                                .padding(.horizontal, PVSpacing.space4)
                                .padding(.vertical, PVSpacing.space3)
                                .background(
                                    RoundedRectangle(cornerRadius: PVRadius.sm, style: .continuous)
                                        .fill(PVColor.surfaceSunken)
                                )
                                .accessibilityIdentifier("onboarding.folderPreview")
                        }
                    }
                } else {
                    Text(L10n.Onboarding.whoAreYouTitle)
                        .font(PVFont.display(size: PVTypeScale.h2))
                        .foregroundStyle(PVColor.textDisplay)
                    Text(
                        L10n.Onboarding.contributorsBody(
                            projectName: model.selectedProject?.lastPathComponent
                                ?? String(localized: L10n.Onboarding.thisProject)
                        )
                    )
                        .font(PVFont.body(size: PVTypeScale.bodySmall))
                        .foregroundStyle(PVColor.textSecondary)
                    VStack(alignment: .leading, spacing: 0) {
                        ForEach(Array(model.catalogUsers.enumerated()), id: \.element.userID) { index, user in
                            OnboardingContributorRow(
                                title: user.displayName,
                                subtitle: user.ref.isEmpty ? nil : user.ref,
                                selected: model.selectedContributorID == user.userID,
                                divider: index == 0 ? .none : .solid
                            ) {
                                model.selectedContributorID = user.userID
                            }
                            .accessibilityLabel(L10n.Onboarding.contributorOption(
                                displayName: user.displayName,
                                ref: user.ref
                            ))
                        }
                        OnboardingContributorRow(
                            title: String(localized: L10n.Onboarding.notListed),
                            selected: model.selectedContributorID == OnboardingModel.newContributorID,
                            divider: .dashed
                        ) {
                            model.selectedContributorID = OnboardingModel.newContributorID
                        }
                    }
                    .background(PVColor.surfaceCard)
                    .clipShape(RoundedRectangle(cornerRadius: PVRadius.md, style: .continuous))
                    .overlay(
                        RoundedRectangle(cornerRadius: PVRadius.md, style: .continuous)
                            .strokeBorder(PVColor.borderSubtle, lineWidth: 1)
                    )
                    .accessibilityElement(children: .contain)
                    .accessibilityIdentifier("onboarding.contributor")
                    if model.selectedContributorID == OnboardingModel.newContributorID {
                        VStack(alignment: .leading, spacing: PVSpacing.space3) {
                            PVField(label: L10n.Onboarding.researcherName) {
                                PVInput(
                                    text: $model.displayName,
                                    isReadOnly: model.researcherLocked,
                                    prompt: L10n.Onboarding.researcherNamePrompt
                                )
                                .accessibilityIdentifier("onboarding.displayName")
                            }
                            Text(L10n.Onboarding.newContributorIDNote)
                                .font(PVFont.mono(size: PVTypeScale.micro))
                                .foregroundStyle(PVColor.textMuted)
                        }
                        .padding(PVSpacing.space5)
                        .background(
                            RoundedRectangle(cornerRadius: PVRadius.md, style: .continuous)
                                .fill(PVColor.surfaceSunken)
                        )
                    }
                }
                if let errorText = model.errorText {
                    Text(errorText)
                        .font(PVFont.body(size: PVTypeScale.bodySmall))
                        .foregroundStyle(PVColor.danger)
                        .accessibilityIdentifier("onboarding.error")
                }
            }
            Spacer(minLength: 24)
            OnboardingFooter(
                model: model,
                leading: {
                    PVButton(L10n.Onboarding.back, variant: .ghost) {
                        model.goBack()
                    }
                    .accessibilityIdentifier("onboarding.back")
                },
                trailing: {
                    PVButton(L10n.Onboarding.continueAction, variant: .primary) {
                        Task { await model.submit() }
                    }
                    .keyboardShortcut(.defaultAction)
                    .disabled(!model.canContinue)
                    .accessibilityIdentifier("onboarding.continue")
                }
            )
        }
        .padding(32)
        .frame(maxWidth: 560, alignment: .leading)
        .frame(maxWidth: .infinity, maxHeight: .infinity, alignment: .center)
    }
}
