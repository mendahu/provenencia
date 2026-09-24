import SwiftUI

/// Citation composer place (S8-10): viewer | form, Layout A + 1500pt two-column.
struct CitationComposerView: View {
    let entry: CitationComposerEntry
    let session: WorkspaceSession
    let store: any GenealogyStore
    let userID: String

    @Environment(WorkspaceNavigation.self) private var navigation
    @State private var model: CitationComposerModel

    private static let wideBreakpoint: CGFloat = 1500
    private static let formWidthNarrow: CGFloat = 520
    private static let formWidthWideMin: CGFloat = 760
    private static let formWidthWideMax: CGFloat = 880

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
        .accessibilityIdentifier("workspace.destination.citationComposer")
        .accessibilityElement(children: .contain)
        .accessibilityLabel(Text(L10n.CitationComposer.accessibilityTitle))
        .accessibilityHint(Text(verbatim: model.identityAnnouncement))
        .task(id: entry.identityKey) {
            await model.prepare()
            if model.shouldFallbackToGraph {
                navigation.go(to: model.graphLocation())
            }
        }
        .pvFormDialog(
            isPresented: observationDialogBinding,
            copy: PVFormDialogCopy(
                title: L10n.CitationComposer.editValueTitle,
                confirm: L10n.CitationComposer.save,
                cancel: L10n.CitationComposer.cancel
            ),
            width: observationDialogWidth,
            isRunning: false,
            confirmDisabled: !model.canConfirmObservation,
            accessibilityIdentifierPrefix: "citationComposer.observation",
            onConfirm: { model.confirmObservationDialog() }
        ) {
            CitationComposerObservationDialogForm(model: model)
        }
        .pvFormDialog(
            isPresented: customTermDialogBinding,
            copy: PVFormDialogCopy(
                title: L10n.CitationComposer.addTermTitle,
                confirm: L10n.CitationComposer.addTermConfirm,
                cancel: L10n.CitationComposer.cancel
            ),
            isRunning: false,
            confirmDisabled: model.customTermLabel.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty,
            accessibilityIdentifierPrefix: "citationComposer.term",
            onConfirm: {
                Task {
                    guard let rowID = model.pendingCustomTermRowID,
                          let row = model.observations.first(where: { $0.id == rowID })
                    else { return }
                    guard let term = await model.createCustomTerm(
                        propertyID: row.propertyID,
                        label: model.customTermLabel
                    ) else { return }
                    model.updateObservationTerm(id: rowID, termID: term.id)
                    model.cancelCustomTermDialog()
                }
            }
        ) {
            PVField(label: L10n.CitationComposer.addTermLabel, error: model.termError) {
                PVInput(text: customTermLabelBinding, size: .sm)
                    .accessibilityIdentifier("citationComposer.term.label")
            }
        }
        .pvConfirm(
            item: abandonBinding,
            copy: { item in
                PVConfirmCopy(
                    title: L10n.CitationComposer.abandonTitle(ref: model.activeCitationRef),
                    message: L10n.CitationComposer.abandonMessage(
                        artifactTitle: model.artifacts.first(where: { $0.id == item.targetArtifactID })?.label
                            ?? item.targetArtifactID
                    ),
                    confirm: L10n.CitationComposer.abandonConfirm,
                    cancel: L10n.CitationComposer.abandonCancel
                )
            },
            tone: .danger,
            accessibilityIdentifierPrefix: "citationComposer.abandon",
            onConfirm: { model.confirmAbandonArtifact() },
            detail: { _ in EmptyView() }
        )
    }

    private var loadFailedGate: some View {
        VStack(alignment: .leading, spacing: PVSpacing.space6) {
            if let loadError = model.loadError {
                PVCallout(tone: .danger, message: loadError)
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
                message: String(localized: L10n.CitationComposer.noArtifactsCallout)
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
            let wide = geo.size.width >= Self.wideBreakpoint
            let formWidth = wide
                ? min(Self.formWidthWideMax, max(Self.formWidthWideMin, geo.size.width * 0.45))
                : Self.formWidthNarrow
            HStack(alignment: .top, spacing: 0) {
                CitationComposerViewerPane(model: model)
                    .frame(maxWidth: .infinity, maxHeight: .infinity)
                    .background(PVColor.surfaceSunken)
                    .accessibilityElement(children: .contain)
                    .accessibilityLabel(Text(L10n.CitationComposer.viewerGroup))
                PVDivider(axis: .vertical, color: PVColor.borderDefault)
                CitationComposerFormPane(
                    model: model,
                    inert: false,
                    wide: wide
                )
                .frame(width: formWidth)
                .frame(maxHeight: .infinity)
                .background(PVColor.surfaceCard)
                .accessibilityElement(children: .contain)
                .accessibilityLabel(Text(L10n.CitationComposer.formGroup))
            }
        }
    }

    private var observationDialogBinding: Binding<Bool> {
        Binding(
            get: { model.observationDialog != nil },
            set: { if !$0 { model.cancelObservationDialog() } }
        )
    }

    private var customTermDialogBinding: Binding<Bool> {
        Binding(
            get: { model.showCustomTermDialog },
            set: { if !$0 { model.cancelCustomTermDialog() } }
        )
    }

    private var customTermLabelBinding: Binding<String> {
        Binding(
            get: { model.customTermLabel },
            set: { model.customTermLabel = $0 }
        )
    }

    private var abandonBinding: Binding<CitationComposerModel.ArtifactAbandon?> {
        Binding(
            get: { model.pendingArtifactAbandon },
            set: { model.pendingArtifactAbandon = $0 }
        )
    }

    private var observationDialogWidth: CGFloat {
        let type = model.observationDialog.flatMap { model.catalogProperty(id: $0.propertyID)?.valueType }
        return type == "name" ? 560 : 480
    }
}
