import SwiftUI

/// Citation composer: evidence column and a fixed observations column.
struct CitationComposerView: View {
    let entry: CitationComposerEntry
    let session: WorkspaceSession
    let store: any GenealogyStore
    let userID: String

    @Environment(WorkspaceNavigation.self) private var navigation
    @State private var model: CitationComposerModel

    init(
        entry: CitationComposerEntry,
        session: WorkspaceSession,
        store: any GenealogyStore,
        userID: String
    ) {
        self.entry = entry
        self.session = session
        self.store = store
        self.userID = userID
        _model = State(
            initialValue: CitationComposerModel(
                entry: entry,
                session: session,
                store: store,
                userID: userID
            )
        )
    }

    var body: some View {
        Group {
            switch model.phase {
            case .loading:
                ProgressView()
                    .frame(maxWidth: .infinity, maxHeight: .infinity)
            case .subjectMissing:
                Color.clear
                    .onAppear { navigation.go(to: model.graphLocation()) }
            case .loadFailed:
                loadFailedGate
            case .compose:
                if model.hasNoArtifacts {
                    noArtifactGate
                } else {
                    composeSplit
                }
            }
        }
        .frame(maxWidth: .infinity, maxHeight: .infinity)
        .background(PVColor.surfacePage)
        .onAppear {
            model.navigation = navigation
            navigation.leaveGuard = model
        }
        .onDisappear {
            if navigation.leaveGuard === model {
                navigation.leaveGuard = nil
            }
        }
        .accessibilityIdentifier("workspace.destination.citationComposer")
        .accessibilityElement(children: .contain)
        .accessibilityLabel(Text(L10n.CitationComposer.accessibilityTitle))
        .accessibilityHint(Text(verbatim: model.identityAnnouncement))
        .task(id: entry.identityKey) {
            await model.prepare()
        }
        .pvFormDialog(
            isPresented: editSheetBinding,
            copy: PVFormDialogCopy(
                title: editSheetTitle,
                verbatimSubtitle: editSheetSubtitle,
                confirm: editSheetConfirm,
                cancel: L10n.CitationComposer.cancel
            ),
            width: 520,
            isRunning: model.editSession?.isSaving == true,
            isDirty: model.editSession?.isDirty == true,
            confirmDisabled: !model.editCanConfirm,
            accessibilityIdentifierPrefix: "citationComposer.edit",
            onConfirm: { Task { await model.saveEdit() } }
        ) {
            CitationComposerEditSheet(model: model)
        }
        .pvFormDialog(
            isPresented: authoringSheetBinding,
            copy: PVFormDialogCopy(
                title: L10n.CitationComposer.newObservationTitle,
                subtitle: authoringSubtitle,
                confirm: authoringConfirm,
                cancel: L10n.CitationComposer.discardDraft
            ),
            width: 520,
            isRunning: model.authoring.isSaving,
            isDirty: model.authoring.holdsLeave,
            confirmDisabled: !model.authoringCanConfirm,
            accessibilityIdentifierPrefix: "citationComposer.authoring",
            onConfirm: { Task { await model.saveAuthoring() } }
        ) {
            CitationComposerAuthoringBox(model: model)
        }
        .pvFormDialog(
            isPresented: customTermDialogBinding,
            copy: PVFormDialogCopy(
                title: L10n.CitationComposer.addTermTitle,
                confirm: L10n.CitationComposer.addTermConfirm,
                cancel: L10n.CitationComposer.cancel
            ),
            isRunning: model.customTermDraft?.isSaving == true,
            isDirty: model.customTermDialogIsDirty,
            confirmDisabled: model.customTermDraft?.label.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty
                ?? true,
            accessibilityIdentifierPrefix: "citationComposer.term",
            onConfirm: { Task { await model.confirmCustomTerm() } }
        ) {
            PVField(label: L10n.CitationComposer.addTermLabel, error: model.customTermDraft?.error) {
                PVInput(text: customTermLabelBinding, size: .sm)
                    .accessibilityIdentifier("citationComposer.term.label")
            }
        }
        .pvFormDialog(
            isPresented: newSubjectBinding,
            copy: PVFormDialogCopy(
                title: L10n.CitationComposer.newSubjectTitle,
                confirm: L10n.CitationComposer.newSubjectConfirm,
                cancel: L10n.CitationComposer.cancel
            ),
            isRunning: model.newSubjectDraft?.isSaving == true,
            isDirty: model.newSubjectDialogIsDirty,
            confirmDisabled: (model.newSubjectDraft?.label.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty
                ?? true)
                || model.isGraphReloading,
            accessibilityIdentifierPrefix: "citationComposer.newSubject",
            onConfirm: { Task { await model.confirmNewSubject() } }
        ) {
            VStack(alignment: .leading, spacing: PVSpacing.space5) {
                PVField(label: L10n.CitationComposer.newSubjectLabel, error: model.newSubjectDraft?.error) {
                    PVInput(text: newSubjectLabelBinding, size: .sm)
                        .accessibilityIdentifier("citationComposer.newSubject.label")
                }
                PVField(label: L10n.CitationComposer.descriptionLabel) {
                    PVTextArea(text: newSubjectDescriptionBinding, lineLimit: 1...3)
                        .accessibilityIdentifier("citationComposer.newSubject.description")
                }
            }
        }
        .pvDeleteImpact(
            item: Binding(
                get: { model.deleteImpact.request },
                set: { model.applyDeleteImpactRequest($0) }
            ),
            isRunning: model.deleteImpact.isRunning,
            error: model.deleteImpact.error,
            accessibilityIdentifierPrefix: "citationComposer.deleteImpact",
            onConfirm: { Task { await model.confirmPendingImpact() } },
            onNavigate: { location in
                navigation.go(to: location)
            }
        )
        .pvConfirm(
            item: leaveBinding,
            copy: { _ in
                PVConfirmCopy(
                    title: L10n.string(L10n.CitationComposer.leaveTitle),
                    message: model.unsavedSummary,
                    confirm: L10n.CitationComposer.leaveDiscard,
                    cancel: L10n.CitationComposer.leaveKeepEditing
                )
            },
            tone: .danger,
            accessibilityIdentifierPrefix: "citationComposer.leave",
            onConfirm: { model.discardLeaveChanges() },
            detail: { _ in EmptyView() }
        )
        .pvConfirm(
            item: autoTranscribeConfirmBinding,
            copy: { pending in
                transcriptionConfirmCopy(pending)
            },
            tone: .irreversible,
            accessibilityIdentifierPrefix: "citationComposer.autoTranscribe.confirm",
            onConfirm: { model.confirmAutoTranscribe() },
            detail: { _ in EmptyView() }
        )
        .pvConfirm(
            item: pasteConfirmBinding,
            copy: { pending in
                transcriptionConfirmCopy(pending)
            },
            tone: .irreversible,
            accessibilityIdentifierPrefix: "citationComposer.pasteTranscription.confirm",
            onConfirm: { model.confirmPasteTranscription() },
            detail: { _ in EmptyView() }
        )
    }

    private var leaveBinding: Binding<CitationComposerModel.PendingLeave?> {
        Binding(
            get: { model.pendingLeave },
            set: { newValue in
                if newValue == nil {
                    model.keepEditingAfterLeave()
                } else {
                    model.pendingLeave = newValue
                }
            }
        )
    }

    private var transcriptionConfirmBinding: Binding<CitationComposerModel.PendingTranscriptionConfirm?> {
        Binding(
            get: { model.pendingTranscriptionConfirm },
            set: { newValue in
                if newValue == nil {
                    model.cancelAutoTranscribeConfirm()
                } else {
                    model.pendingTranscriptionConfirm = newValue
                }
            }
        )
    }

    private var autoTranscribeConfirmBinding: Binding<CitationComposerModel.PendingTranscriptionConfirm?> {
        Binding(
            get: {
                guard let pending = model.pendingTranscriptionConfirm, !pending.isPaste else { return nil }
                return pending
            },
            set: { transcriptionConfirmBinding.wrappedValue = $0 }
        )
    }

    private var pasteConfirmBinding: Binding<CitationComposerModel.PendingTranscriptionConfirm?> {
        Binding(
            get: {
                guard let pending = model.pendingTranscriptionConfirm, pending.isPaste else { return nil }
                return pending
            },
            set: { transcriptionConfirmBinding.wrappedValue = $0 }
        )
    }

    private func transcriptionConfirmCopy(
        _ pending: CitationComposerModel.PendingTranscriptionConfirm
    ) -> PVConfirmCopy {
        if pending.isPaste {
            let page = pending.pastePage ?? 1
            let lines = pending.pasteLineCount ?? 1
            return PVConfirmCopy(
                title: L10n.string(L10n.CitationComposer.autoTranscribeReplaceTitle),
                message: L10n.CitationComposer.pasteReplaceMessage(selectedLines: lines, page: page),
                confirm: L10n.CitationComposer.pasteReplaceConfirm,
                cancel: L10n.CitationComposer.autoTranscribeKeep
            )
        }
        let replacing = !pending.existingText.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty
        let title: LocalizedStringResource
        let message: LocalizedStringResource
        if replacing && pending.includeWholePage {
            title = L10n.CitationComposer.autoTranscribeReplaceTitle
            message = L10n.CitationComposer.autoTranscribeReplaceAndWholePageMessage
        } else if replacing {
            title = L10n.CitationComposer.autoTranscribeReplaceTitle
            message = L10n.CitationComposer.autoTranscribeReplaceMessage
        } else {
            title = L10n.CitationComposer.autoTranscribeWholePageTitle
            message = L10n.CitationComposer.autoTranscribeWholePageMessage
        }
        return PVConfirmCopy(
            title: L10n.string(title),
            message: L10n.string(message),
            confirm: L10n.CitationComposer.autoTranscribeConfirm,
            cancel: replacing
                ? L10n.CitationComposer.autoTranscribeKeep
                : L10n.CitationComposer.cancel
        )
    }

    private var loadFailedGate: some View {
        VStack(alignment: .leading, spacing: PVSpacing.space6) {
            if let loadError = model.loadError {
                PVCallout(
                    tone: model.loadFailureIsMissingCitation ? .neutral : .danger,
                    message: loadError
                )
                    .accessibilityIdentifier("citationComposer.loadFailed")
            }
            PVButton(L10n.CitationComposer.loadFailedBack, variant: .secondary) {
                navigation.go(to: model.graphLocation())
            }
            .accessibilityIdentifier("citationComposer.loadFailed.back")
            Spacer(minLength: 0)
        }
        .padding(PVSpacing.space7)
        .frame(maxWidth: .infinity, maxHeight: .infinity, alignment: .topLeading)
    }

    private var noArtifactGate: some View {
        HStack {
            Spacer(minLength: 0)
            PVCallout(
                tone: .neutral,
                title: L10n.CitationComposer.noArtifactsTitle,
                message: L10n.string(L10n.CitationComposer.noArtifactsCallout)
            ) {
                PVButton(L10n.CitationComposer.goToSourcePage, variant: .secondary, size: .sm) {
                    navigation.go(to: model.sourcePageLocation())
                }
            }
            .frame(maxWidth: 480)
            .accessibilityIdentifier("citationComposer.noArtifacts")
            Spacer(minLength: 0)
        }
        .padding(PVSpacing.space7)
        .frame(maxWidth: .infinity, maxHeight: .infinity)
    }

    private var composeSplit: some View {
        GeometryReader { geo in
            let threeUp = geo.size.width >= 1400
            HStack(alignment: .top, spacing: 0) {
                evidence(threeUp: threeUp, height: geo.size.height)
                CitationComposerObservationsColumn(model: model) {
                    navigation.go(to: model.graphLocation())
                }
                .frame(width: 520)
                .frame(maxHeight: .infinity)
            }
        }
    }

    @ViewBuilder
    private func evidence(threeUp: Bool, height: CGFloat) -> some View {
        if threeUp {
            HStack(alignment: .top, spacing: 0) {
                CitationComposerViewerPane(model: model)
                    .frame(maxWidth: .infinity, maxHeight: .infinity)
                    .accessibilityElement(children: .contain)
                    .accessibilityLabel(Text(L10n.CitationComposer.viewerGroup))
                    .overlay(alignment: .trailing) {
                        PVDivider(axis: .vertical, color: PVColor.borderSubtle)
                    }
                CitationComposerEvidenceColumn(model: model, layout: .side)
                    .frame(width: 440, alignment: .topLeading)
                    .overlay(alignment: .trailing) {
                        PVDivider(axis: .vertical, color: PVColor.borderSubtle)
                    }
            }
            .frame(maxWidth: .infinity, maxHeight: .infinity)
        } else {
            VStack(spacing: 0) {
                CitationComposerEvidenceColumn(model: model, layout: .panel)
                    .frame(maxHeight: height * 0.55, alignment: .top)
                CitationComposerViewerPane(model: model)
                    .frame(maxWidth: .infinity, maxHeight: .infinity)
            }
            .frame(maxWidth: .infinity, maxHeight: .infinity)
            .accessibilityElement(children: .contain)
            .accessibilityLabel(Text(L10n.CitationComposer.viewerGroup))
            .overlay(alignment: .trailing) {
                PVDivider(axis: .vertical, color: PVColor.borderSubtle)
            }
        }
    }

    private var authoringSheetBinding: Binding<Bool> {
        Binding(
            get: {
                if case .authoring = model.presented { return true }
                return false
            },
            set: { if !$0 { model.discardAuthoring() } }
        )
    }

    private var authoringSubtitle: LocalizedStringResource {
        if case .bridge = model.authoring {
            return L10n.CitationComposer.newBridgeSubtitle
        }
        return L10n.CitationComposer.newObservationSubtitle
    }

    private var authoringConfirm: LocalizedStringResource {
        if case .bridge = model.authoring {
            return L10n.CitationComposer.saveBridge
        }
        return L10n.CitationComposer.saveObservation
    }

    private var editSheetBinding: Binding<Bool> {
        Binding(
            get: {
                if case .edit = model.presented { return true }
                return false
            },
            set: { if !$0 { model.discardEdit() } }
        )
    }

    private var editSheetTitle: LocalizedStringResource {
        if case .bridge = model.editSession?.kind {
            return L10n.CitationComposer.editBridgeTitle
        }
        return L10n.CitationComposer.editObservation
    }

    private var editSheetConfirm: LocalizedStringResource {
        if case .bridge = model.editSession?.kind {
            return L10n.CitationComposer.saveBridge
        }
        return L10n.CitationComposer.saveObservation
    }

    private var editSheetSubtitle: String {
        guard let session = model.editSession else { return "" }
        switch session.kind {
        case .property(_, let id, let ref):
            let shown = ref.isEmpty ? id : ref
            return L10n.CitationComposer.editObservationSubtitle(ref: shown)
        case .bridge(let endA, let endB, _, _, let bridgeRef, _):
            let opened = model.vocabulary.subjectsByID[session.openedFromSubjectID]?.label
                ?? model.vocabulary.subjectsByID[endA]?.label
                ?? ""
            let nameA = model.vocabulary.subjectsByID[endA]?.label ?? ""
            let nameB = model.vocabulary.subjectsByID[endB]?.label ?? ""
            return L10n.CitationComposer.editBridgeSubtitle(
                ref: bridgeRef,
                openedFrom: opened,
                endA: nameA,
                endB: nameB
            )
        }
    }

    private var customTermDialogBinding: Binding<Bool> {
        Binding(
            get: { model.customTermDraft != nil },
            set: { if !$0 { model.cancelCustomTermDialog() } }
        )
    }

    private var customTermLabelBinding: Binding<String> {
        Binding(
            get: { model.customTermDraft?.label ?? "" },
            set: { model.customTermDraft?.label = $0 }
        )
    }

    private var newSubjectBinding: Binding<Bool> {
        Binding(
            get: { model.newSubjectDraft != nil },
            set: { if !$0 { model.cancelNewSubject() } }
        )
    }

    private var newSubjectLabelBinding: Binding<String> {
        Binding(
            get: { model.newSubjectDraft?.label ?? "" },
            set: { model.newSubjectDraft?.label = $0 }
        )
    }

    private var newSubjectDescriptionBinding: Binding<String> {
        Binding(
            get: { model.newSubjectDraft?.description ?? "" },
            set: { model.newSubjectDraft?.description = $0 }
        )
    }
}
