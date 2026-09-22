import SwiftUI

/// Citation composer place (S7-08): board-aligned viewer | form + observation dialog.
struct CitationComposerView: View {
    let sourceID: String
    let subjectID: String
    let citationID: String?
    let session: WorkspaceSession
    let store: any GenealogyStore
    let userID: String

    @Environment(WorkspaceNavigation.self) private var navigation
    @State private var model: CitationComposerModel
    @State private var customTermLabel = ""
    @State private var showCustomTermDialog = false

    init(
        sourceID: String,
        subjectID: String,
        citationID: String? = nil,
        session: WorkspaceSession,
        store: any GenealogyStore,
        userID: String
    ) {
        self.sourceID = sourceID
        self.subjectID = subjectID
        self.citationID = citationID
        self.session = session
        self.store = store
        self.userID = userID
        _model = State(
            initialValue: CitationComposerModel(
                sourceID: sourceID,
                subjectID: subjectID,
                citationID: citationID,
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
            case .pickArtifact:
                CitationComposerArtifactPicker(model: model) {
                    navigation.go(to: model.graphLocation())
                }
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
        .accessibilityIdentifier("workspace.destination.citationComposer")
        .accessibilityElement(children: .contain)
        .accessibilityLabel(Text(L10n.CitationComposer.accessibilityTitle))
        .task(id: "\(sourceID)-\(subjectID)-\(citationID ?? "")") {
            await model.prepare()
            if model.shouldFallbackToGraph {
                navigation.go(to: model.graphLocation())
            }
        }
        .pvFormDialog(
            isPresented: observationDialogBinding,
            copy: PVFormDialogCopy(
                title: L10n.CitationComposer.addObservationTitle,
                subtitle: L10n.CitationComposer.addObservationSubtitle,
                confirm: L10n.CitationComposer.addObservation,
                cancel: L10n.CitationComposer.cancel
            ),
            isRunning: false,
            confirmDisabled: !model.canConfirmObservation,
            accessibilityIdentifierPrefix: "citationComposer.observation",
            onConfirm: { model.confirmObservationDialog() }
        ) {
            CitationComposerObservationDialogForm(model: model) {
                customTermLabel = ""
                showCustomTermDialog = true
            }
        }
        .pvFormDialog(
            isPresented: $showCustomTermDialog,
            copy: PVFormDialogCopy(
                title: L10n.CitationComposer.addTermTitle,
                confirm: L10n.CitationComposer.addTermConfirm,
                cancel: L10n.CitationComposer.cancel
            ),
            isRunning: false,
            confirmDisabled: customTermLabel.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty,
            accessibilityIdentifierPrefix: "citationComposer.term",
            onConfirm: {
                Task {
                    guard let propertyID = model.observationDialog?.propertyID else { return }
                    if let term = await model.createCustomTerm(
                        propertyID: propertyID,
                        label: customTermLabel
                    ),
                       var draft = model.observationDialog
                    {
                        draft.valueTermID = term.id
                        model.updateObservationDialog(draft)
                    }
                    customTermLabel = ""
                    showCustomTermDialog = false
                }
            }
        ) {
            PVField(label: L10n.CitationComposer.addTermLabel) {
                PVInput(text: $customTermLabel, size: .sm)
                    .accessibilityIdentifier("citationComposer.term.label")
            }
        }
    }

    // MARK: - Frame 8 / compose layout

    private var noArtifactGate: some View {
        HStack(alignment: .top, spacing: 0) {
            VStack(alignment: .leading, spacing: PVSpacing.space7) {
                PVCallout(
                    tone: .warning,
                    message: String(localized: L10n.CitationComposer.noArtifactsCallout)
                ) {
                    PVButton(L10n.CitationComposer.goToSourcePage, variant: .secondary, size: .sm) {
                        navigation.go(to: model.sourcePageLocation())
                    }
                }
                Spacer(minLength: 0)
            }
            .padding(PVSpacing.space7)
            .frame(maxWidth: .infinity, maxHeight: .infinity, alignment: .topLeading)
            .background(PVColor.surfaceSunken)
            PVDivider(axis: .vertical, color: PVColor.borderDefault)
            CitationComposerFormPane(model: model, inert: true)
                .frame(width: CitationComposerFormPane.sidebarWidth)
                .frame(maxHeight: .infinity)
                .background(PVColor.surfaceCard)
        }
    }

    private var composeSplit: some View {
        HStack(alignment: .top, spacing: 0) {
            CitationComposerViewerPane(model: model)
                .frame(maxWidth: .infinity, maxHeight: .infinity)
                .background(PVColor.surfaceSunken)
            PVDivider(axis: .vertical, color: PVColor.borderDefault)
            CitationComposerFormPane(model: model, inert: false)
                .frame(width: CitationComposerFormPane.sidebarWidth)
                .frame(maxHeight: .infinity)
                .background(PVColor.surfaceCard)
        }
    }

    private var observationDialogBinding: Binding<Bool> {
        Binding(
            get: { model.observationDialog != nil },
            set: { if !$0 { model.cancelObservationDialog() } }
        )
    }
}
