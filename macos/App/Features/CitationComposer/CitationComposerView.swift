import SwiftUI

/// Citation composer place (S7-08): board-aligned viewer | form + observation dialog.
struct CitationComposerView: View {
    let sourceID: String
    let subjectID: String
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
        session: WorkspaceSession,
        store: any GenealogyStore,
        userID: String
    ) {
        self.sourceID = sourceID
        self.subjectID = subjectID
        self.session = session
        self.store = store
        self.userID = userID
        _model = State(
            initialValue: CitationComposerModel(
                sourceID: sourceID,
                subjectID: subjectID,
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
                artifactPicker
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
        .task(id: "\(sourceID)-\(subjectID)") {
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
            observationDialogForm
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

    // MARK: - Frame 7 picker

    private var artifactPicker: some View {
        VStack(spacing: 0) {
            Spacer(minLength: PVSpacing.space9)
            VStack(alignment: .leading, spacing: PVSpacing.space7) {
                VStack(alignment: .leading, spacing: PVSpacing.space3) {
                    Text(L10n.CitationComposer.pickArtifactTitle)
                        .font(PVFont.display(size: PVTypeScale.h2, weight: PVFontWeight.medium))
                        .foregroundStyle(PVColor.textDisplay)
                    Text(verbatim: pickArtifactCaption)
                        .font(PVFont.body(size: PVTypeScale.body))
                        .foregroundStyle(PVColor.textSecondary)
                }
                LazyVGrid(
                    columns: [GridItem(.adaptive(minimum: 160), spacing: PVSpacing.space5)],
                    spacing: PVSpacing.space5
                ) {
                    ForEach(model.artifacts, id: \.id) { artifact in
                        artifactPickCard(artifact)
                    }
                }
            }
            .padding(PVSpacing.space9)
            .frame(maxWidth: 720)
            .background(
                RoundedRectangle(cornerRadius: PVRadius.lg, style: .continuous)
                    .fill(PVColor.surfaceRaised)
            )
            .overlay(
                RoundedRectangle(cornerRadius: PVRadius.lg, style: .continuous)
                    .strokeBorder(PVColor.borderSubtle, lineWidth: 1)
            )
            Spacer(minLength: PVSpacing.space9)
            HStack {
                PVButton(L10n.CitationComposer.cancel, variant: .secondary) {
                    navigation.go(to: model.graphLocation())
                }
                Spacer()
                PVButton(L10n.CitationComposer.continuePick, variant: .primary) {
                    model.confirmArtifactSelection()
                }
                .disabled(model.pendingArtifactID == nil)
                .accessibilityIdentifier("citationComposer.pick.continue")
            }
            .padding(PVSpacing.space7)
            .frame(maxWidth: 720)
        }
        .frame(maxWidth: .infinity, maxHeight: .infinity)
        .padding(.horizontal, PVSpacing.space9)
    }

    private var pickArtifactCaption: String {
        let title = model.sourceTitle.isEmpty ? "—" : model.sourceTitle
        return L10n.CitationComposer.pickArtifactCount(title: title, count: model.artifacts.count)
    }

    private func artifactPickCard(_ artifact: CatalogArtifact) -> some View {
        let selected = model.pendingArtifactID == artifact.id
        return Button {
            model.selectPendingArtifact(artifact.id)
        } label: {
            VStack(alignment: .leading, spacing: PVSpacing.space3) {
                RoundedRectangle(cornerRadius: PVRadius.sm, style: .continuous)
                    .fill(PVColor.surfaceSunken)
                    .frame(height: 100)
                    .overlay {
                        Text(verbatim: artifactDisplayName(artifact))
                            .font(PVFont.body(size: PVTypeScale.caption))
                            .foregroundStyle(PVColor.textMuted)
                            .multilineTextAlignment(.center)
                            .padding(PVSpacing.space4)
                    }
                    .overlay(
                        RoundedRectangle(cornerRadius: PVRadius.sm, style: .continuous)
                            .strokeBorder(
                                selected ? PVColor.accent : PVColor.borderDefault,
                                lineWidth: selected ? 2 : 1
                            )
                    )
                Text(verbatim: artifactDisplayName(artifact))
                    .font(PVFont.body(size: PVTypeScale.caption, weight: PVFontWeight.medium))
                    .foregroundStyle(PVColor.textPrimary)
                    .lineLimit(2)
                Text(verbatim: artifactMediaCaption(artifact))
                    .font(PVFont.body(size: PVTypeScale.micro))
                    .foregroundStyle(PVColor.textMuted)
            }
        }
        .buttonStyle(.plain)
        .accessibilityLabel(Text(verbatim: artifactDisplayName(artifact)))
        .accessibilityAddTraits(selected ? .isSelected : [])
    }

    // MARK: - Frame 8 no-artifact gate

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
            .frame(minWidth: 280, idealWidth: 420, maxWidth: 560, maxHeight: .infinity, alignment: .topLeading)
            PVDivider()
            formPane(inert: true)
                .frame(maxWidth: .infinity)
        }
    }

    // MARK: - Compose split

    private var composeSplit: some View {
        HStack(alignment: .top, spacing: 0) {
            viewerPane
                .frame(minWidth: 280, idealWidth: 440, maxWidth: 580)
            PVDivider()
            formPane(inert: false)
                .frame(maxWidth: .infinity)
        }
    }

    // MARK: - Viewer pane (Frames 1 / 2 / 12)

    private var viewerPane: some View {
        VStack(alignment: .leading, spacing: PVSpacing.space5) {
            viewerHeader
            viewerToolStrip
            if let locatorError = model.locatorError, model.submitAttempted {
                PVCallout(tone: .danger, message: locatorError, compact: true)
            } else if model.hasLocator {
                locatorCrumb
            }
            viewerCanvas
                .frame(maxWidth: .infinity, maxHeight: .infinity)
        }
        .padding(PVSpacing.space7)
    }

    private var viewerHeader: some View {
        HStack(alignment: .firstTextBaseline, spacing: PVSpacing.space3) {
            Text(verbatim: model.sourceTitle.isEmpty ? "—" : model.sourceTitle)
                .font(PVFont.body(size: PVTypeScale.body, weight: PVFontWeight.medium))
                .foregroundStyle(PVColor.textDisplay)
                .lineLimit(1)
            Text(verbatim: "·")
                .foregroundStyle(PVColor.textMuted)
            Text(verbatim: artifactIndexCaption)
                .font(PVFont.body(size: PVTypeScale.caption))
                .foregroundStyle(PVColor.textSecondary)
                .lineLimit(1)
            Spacer(minLength: PVSpacing.space3)
            if model.artifacts.count > 1 {
                Button {
                    model.changeArtifact()
                } label: {
                    Text(L10n.CitationComposer.changeArtifact)
                        .font(PVFont.body(size: PVTypeScale.caption, weight: PVFontWeight.medium))
                        .foregroundStyle(PVColor.accent)
                }
                .buttonStyle(.plain)
                .accessibilityIdentifier("citationComposer.changeArtifact")
            }
        }
    }

    private var artifactIndexCaption: String {
        guard let index = model.selectedArtifactIndex else {
            return String(localized: L10n.CitationComposer.artifactKindUnknown)
        }
        let kind = model.isPDFArtifact
            ? String(localized: L10n.CitationComposer.artifactKindPDF)
            : (model.isImageArtifact
                ? String(localized: L10n.CitationComposer.artifactKindImage)
                : String(localized: L10n.CitationComposer.artifactKindUnknown))
        return L10n.CitationComposer.artifactIndexOf(
            index: index + 1,
            total: model.artifacts.count,
            kind: kind
        )
    }

    private var viewerToolStrip: some View {
        HStack(spacing: PVSpacing.space4) {
            if model.isPDFArtifact {
                HStack(spacing: PVSpacing.space2) {
                    PVIconButton(.chevronBack, label: L10n.CitationComposer.previousPage, size: .sm) {
                        model.goToPreviousPage()
                    }
                    .disabled(model.viewerPage <= 1)
                    Text(verbatim: "\(model.viewerPage)")
                        .font(PVFont.mono(size: PVTypeScale.caption, weight: PVFontWeight.medium))
                        .foregroundStyle(PVColor.textPrimary)
                        .frame(minWidth: 20)
                    Text(verbatim: pageOfCaption)
                        .font(PVFont.body(size: PVTypeScale.caption))
                        .foregroundStyle(PVColor.textMuted)
                    PVIconButton(.chevronForward, label: L10n.CitationComposer.nextPage, size: .sm) {
                        model.goToNextPage()
                    }
                    .disabled(model.viewerPage >= model.viewerPageCount)
                }
                toolSep
            }

            HStack(spacing: PVSpacing.space2) {
                Text(verbatim: "100%")
                    .font(PVFont.body(size: PVTypeScale.caption))
                    .foregroundStyle(PVColor.textMuted)
            }
            toolSep

            PVButton(L10n.CitationComposer.drawRegion, variant: .secondary, size: .sm) {
                model.markWholeImageLocator()
            }
            .accessibilityIdentifier("citationComposer.drawRegion")

            Spacer(minLength: 0)

            if model.hasLocator {
                Button {
                    model.clearLocator()
                } label: {
                    Text(L10n.CitationComposer.clearLocator)
                        .font(PVFont.body(size: PVTypeScale.caption, weight: PVFontWeight.medium))
                        .foregroundStyle(PVColor.accent)
                }
                .buttonStyle(.plain)
                .accessibilityIdentifier("citationComposer.clearLocator")
            }
        }
    }

    private var toolSep: some View {
        Rectangle()
            .fill(PVColor.borderSubtle)
            .frame(width: 1, height: 16)
    }

    private var pageOfCaption: String {
        L10n.CitationComposer.pageOf(total: model.viewerPageCount)
    }

    private var locatorCrumb: some View {
        HStack(spacing: PVSpacing.space2) {
            if model.locatorIsWholeImage {
                Text(L10n.CitationComposer.locatorWholeImage)
                    .font(PVFont.body(size: PVTypeScale.caption, weight: PVFontWeight.medium))
                    .foregroundStyle(PVColor.textSecondary)
            } else if let page = model.locatorPage {
                Text(verbatim: L10n.CitationComposer.locatorPage(page))
                    .font(PVFont.body(size: PVTypeScale.caption, weight: PVFontWeight.medium))
                    .foregroundStyle(PVColor.textSecondary)
            }
        }
    }

    private var viewerCanvas: some View {
        RoundedRectangle(cornerRadius: PVRadius.md, style: .continuous)
            .strokeBorder(PVColor.borderDefault, style: StrokeStyle(lineWidth: 1, dash: [6, 4]))
            .background(
                RoundedRectangle(cornerRadius: PVRadius.md, style: .continuous)
                    .fill(PVColor.surfaceSunken)
            )
            .overlay {
                VStack(spacing: PVSpacing.space3) {
                    Text(verbatim: canvasCaption)
                        .font(PVFont.body(size: PVTypeScale.body, weight: PVFontWeight.medium))
                        .foregroundStyle(PVColor.textMuted)
                        .multilineTextAlignment(.center)
                    Text(L10n.CitationComposer.viewerPlaceholderMessage)
                        .font(PVFont.body(size: PVTypeScale.caption))
                        .foregroundStyle(PVColor.textFaint)
                        .multilineTextAlignment(.center)
                }
                .padding(PVSpacing.space7)
            }
    }

    private var canvasCaption: String {
        if !model.hasLocator {
            return String(localized: L10n.CitationComposer.canvasNothingSelected)
        }
        if let artifact = model.selectedArtifact {
            return artifactDisplayName(artifact)
        }
        return String(localized: L10n.CitationComposer.viewerPlaceholderTitle)
    }

    // MARK: - Form pane

    private func formPane(inert: Bool) -> some View {
        VStack(alignment: .leading, spacing: 0) {
            ScrollView {
                VStack(alignment: .leading, spacing: PVSpacing.space8) {
                    citationFields(inert: inert)
                    observationsSection(inert: inert)
                }
                .padding(PVSpacing.space7)
            }
            .disabled(inert)
            .opacity(inert ? 0.55 : 1)
            PVDivider()
            footer(inert: inert)
        }
    }

    private func citationFields(inert: Bool) -> some View {
        VStack(alignment: .leading, spacing: PVSpacing.space5) {
            PVField(label: L10n.CitationComposer.transcriptionLabel) {
                HStack(spacing: PVSpacing.space4) {
                    Toggle(isOn: Bindable(model).transcriptionUncertain) {
                        Text(L10n.CitationComposer.uncertainLabel)
                            .font(PVFont.body(size: PVTypeScale.caption))
                    }
                    .toggleStyle(.checkbox)
                    .disabled(inert)
                    .accessibilityIdentifier("citationComposer.uncertain")
                    Spacer(minLength: 0)
                }
            }
            PVTextArea(text: Bindable(model).transcription, lineLimit: 2...6)
                .disabled(inert)
                .accessibilityIdentifier("citationComposer.transcription")
            if model.transcriptionUncertain {
                PVField(label: L10n.CitationComposer.uncertainNoteLabel) {
                    PVTextArea(text: Bindable(model).transcriptionNote, lineLimit: 1...3)
                        .disabled(inert)
                        .accessibilityIdentifier("citationComposer.uncertainNote")
                }
            }
            PVField(label: L10n.CitationComposer.descriptionLabel) {
                PVTextArea(text: Bindable(model).citationDescription, lineLimit: 1...4)
                    .disabled(inert)
                    .accessibilityIdentifier("citationComposer.description")
            }
            if inert {
                Text(L10n.CitationComposer.fieldsDisabledHint)
                    .font(PVFont.body(size: PVTypeScale.caption))
                    .foregroundStyle(PVColor.textMuted)
            }
        }
    }

    private func observationsSection(inert: Bool) -> some View {
        VStack(alignment: .leading, spacing: PVSpacing.space5) {
            Text(L10n.CitationComposer.observationsSection)
                .font(PVFont.body(size: PVTypeScale.body, weight: PVFontWeight.semibold))
                .foregroundStyle(PVColor.textDisplay)

            if model.observations.isEmpty {
                PVCallout(
                    tone: .info,
                    message: String(localized: L10n.CitationComposer.noObservationsError),
                    compact: true
                )
            }

            ForEach(model.observations) { row in
                observationRow(row, inert: inert)
            }

            if !inert {
                PVButton(L10n.CitationComposer.addObservation, variant: .secondary, size: .sm) {
                    model.beginAddObservation()
                }
                .accessibilityIdentifier("citationComposer.addObservation")
            }
        }
    }

    private func observationRow(_ row: CitationComposerModel.ObservationRow, inert: Bool) -> some View {
        let property = model.catalogProperty(id: row.propertyID)
        return HStack(alignment: .top, spacing: PVSpacing.space4) {
            VStack(alignment: .leading, spacing: PVSpacing.space1) {
                Text(verbatim: property?.label ?? "—")
                    .font(PVFont.body(size: PVTypeScale.caption, weight: PVFontWeight.medium))
                    .foregroundStyle(PVColor.textSecondary)
                Text(verbatim: model.observationSummary(for: row))
                    .font(PVFont.body(size: PVTypeScale.body))
                    .foregroundStyle(PVColor.textPrimary)
            }
            Spacer(minLength: 0)
            HStack(spacing: PVSpacing.space2) {
                if row.polarity == "negative" {
                    Text(L10n.CitationComposer.polarityNegates)
                        .font(PVFont.body(size: PVTypeScale.micro, weight: PVFontWeight.medium))
                        .foregroundStyle(PVColor.textSecondary)
                        .padding(.horizontal, PVSpacing.space3)
                        .padding(.vertical, PVSpacing.space1)
                        .background(
                            Capsule(style: .continuous)
                                .fill(PVColor.surfaceSunken)
                        )
                }
                if !inert {
                    PVIconButton(.penLine, label: L10n.CitationComposer.editObservation, size: .sm) {
                        model.beginEditObservation(row)
                    }
                    PVIconButton(
                        .trash,
                        label: L10n.CitationComposer.removeObservation,
                        size: .sm,
                        tone: .danger
                    ) {
                        model.removeObservation(id: row.id)
                    }
                }
            }
        }
        .padding(PVSpacing.space5)
        .background(
            RoundedRectangle(cornerRadius: PVRadius.md, style: .continuous)
                .fill(PVColor.surfaceRaised)
        )
        .overlay(
            RoundedRectangle(cornerRadius: PVRadius.md, style: .continuous)
                .strokeBorder(PVColor.borderSubtle, lineWidth: 1)
        )
    }

    private func footer(inert: Bool) -> some View {
        VStack(alignment: .leading, spacing: PVSpacing.space4) {
            if let formError = model.formError, model.submitAttempted {
                PVCallout(tone: .danger, message: formError, compact: true)
            }
            HStack {
                PVButton(L10n.CitationComposer.cancel, variant: .secondary) {
                    navigation.go(to: model.graphLocation())
                }
                .disabled(model.isSubmitting)
                Spacer()
                PVButton(
                    L10n.CitationComposer.save,
                    variant: .primary,
                    loading: model.isSubmitting
                ) {
                    Task {
                        if let location = await model.submit() {
                            navigation.go(to: location)
                        }
                    }
                }
                .disabled(model.isSubmitting || inert)
                .accessibilityIdentifier("citationComposer.save")
            }
        }
        .padding(PVSpacing.space7)
    }

    // MARK: - Observation dialog (Frames 3–6, 13)

    @ViewBuilder
    private var observationDialogForm: some View {
        if let draft = model.observationDialog {
            VStack(alignment: .leading, spacing: PVSpacing.space6) {
                VStack(alignment: .leading, spacing: PVSpacing.space2) {
                    PVComboBox(
                        selection: dialogPropertyBinding,
                        options: model.propertyOptions,
                        size: .sm,
                        placeholder: L10n.CitationComposer.propertyPlaceholder,
                        emptyLabel: L10n.CitationComposer.propertyEmpty,
                        label: L10n.CitationComposer.propertyLabel,
                        accessibilityIdentifierPrefix: "citationComposer.dialog.property"
                    )
                    if let error = model.dialogPropertyError {
                        Text(verbatim: error)
                            .font(PVFont.body(size: PVTypeScale.caption))
                            .foregroundStyle(PVColor.danger)
                    }
                }

                VStack(alignment: .leading, spacing: PVSpacing.space2) {
                    Text(L10n.CitationComposer.polarityLabel)
                        .font(PVFont.body(size: PVTypeScale.caption, weight: PVFontWeight.medium))
                        .foregroundStyle(PVColor.textSecondary)
                    PVChipGroup(style: .segmented) {
                        PVChip(
                            L10n.CitationComposer.polarityAsserts,
                            isSelected: draft.polarity != "negative",
                            expands: true,
                            selectionLift: true,
                            action: {
                                var next = draft
                                next.polarity = "positive"
                                model.updateObservationDialog(next)
                            }
                        )
                        PVChip(
                            L10n.CitationComposer.polarityNegates,
                            isSelected: draft.polarity == "negative",
                            expands: true,
                            selectionLift: true,
                            action: {
                                var next = draft
                                next.polarity = "negative"
                                model.updateObservationDialog(next)
                            }
                        )
                    }
                }

                dialogValueEditor(draft)
            }
        }
    }

    @ViewBuilder
    private func dialogValueEditor(_ draft: CitationComposerModel.ObservationDialogState) -> some View {
        let property = model.catalogProperty(id: draft.propertyID)
        VStack(alignment: .leading, spacing: PVSpacing.space2) {
            Text(valueTypeLabel(property?.valueType))
                .font(PVFont.body(size: PVTypeScale.caption, weight: PVFontWeight.medium))
                .foregroundStyle(PVColor.textSecondary)

            if let property {
                switch property.valueType {
                case "text":
                    PVTextArea(text: dialogTextBinding, lineLimit: 1...4)
                        .accessibilityIdentifier("citationComposer.dialog.valueText")
                case "integer":
                    PVInput(text: dialogIntegerBinding, size: .sm)
                        .accessibilityIdentifier("citationComposer.dialog.valueInteger")
                case "term":
                    VStack(alignment: .leading, spacing: PVSpacing.space3) {
                        PVComboBox(
                            selection: dialogTermBinding,
                            options: model.termOptions(for: property.id),
                            size: .sm,
                            placeholder: L10n.CitationComposer.termPlaceholder,
                            emptyLabel: L10n.CitationComposer.termEmpty,
                            label: L10n.CitationComposer.termLabel,
                            accessibilityIdentifierPrefix: "citationComposer.dialog.term"
                        )
                        PVButton(L10n.CitationComposer.addCustomTerm, variant: .ghost, size: .sm) {
                            customTermLabel = ""
                            showCustomTermDialog = true
                        }
                    }
                case "date":
                    DateValueEditorForm(
                        draft: dialogDateBinding,
                        accessibilityIdentifierPrefix: "citationComposer.dialog.date"
                    )
                default:
                    Text(L10n.CitationComposer.unsupportedValueTypeError)
                        .font(PVFont.body(size: PVTypeScale.caption))
                        .foregroundStyle(PVColor.textMuted)
                }
            } else {
                Text(L10n.CitationComposer.dialogValuePickPropertyFirst)
                    .font(PVFont.body(size: PVTypeScale.body))
                    .foregroundStyle(PVColor.textMuted)
                    .frame(maxWidth: .infinity, alignment: .leading)
                    .padding(PVSpacing.space5)
                    .background(
                        RoundedRectangle(cornerRadius: PVRadius.sm, style: .continuous)
                            .fill(PVColor.surfaceSunken)
                    )
            }

            if let error = model.dialogValueError {
                Text(verbatim: error)
                    .font(PVFont.body(size: PVTypeScale.caption))
                    .foregroundStyle(PVColor.danger)
            }
        }
    }

    private func valueTypeLabel(_ type: String?) -> LocalizedStringResource {
        switch type {
        case "text": return L10n.CitationComposer.valueLabelText
        case "integer": return L10n.CitationComposer.valueLabelInteger
        case "term": return L10n.CitationComposer.valueLabelTerm
        case "date": return L10n.CitationComposer.valueLabelDate
        default: return L10n.CitationComposer.valueLabel
        }
    }

    // MARK: - Bindings

    private var observationDialogBinding: Binding<Bool> {
        Binding(
            get: { model.observationDialog != nil },
            set: { if !$0 { model.cancelObservationDialog() } }
        )
    }

    private var dialogPropertyBinding: Binding<String> {
        Binding(
            get: { model.observationDialog?.propertyID ?? "" },
            set: {
                guard var draft = model.observationDialog else { return }
                draft.propertyID = $0
                model.updateObservationDialog(draft)
            }
        )
    }

    private var dialogTextBinding: Binding<String> {
        Binding(
            get: { model.observationDialog?.valueText ?? "" },
            set: {
                guard var draft = model.observationDialog else { return }
                draft.valueText = $0
                model.updateObservationDialog(draft)
            }
        )
    }

    private var dialogIntegerBinding: Binding<String> {
        Binding(
            get: { model.observationDialog?.valueIntegerText ?? "" },
            set: {
                guard var draft = model.observationDialog else { return }
                draft.valueIntegerText = $0
                model.updateObservationDialog(draft)
            }
        )
    }

    private var dialogTermBinding: Binding<String> {
        Binding(
            get: { model.observationDialog?.valueTermID ?? "" },
            set: {
                guard var draft = model.observationDialog else { return }
                draft.valueTermID = $0
                model.updateObservationDialog(draft)
            }
        )
    }

    private var dialogDateBinding: Binding<DateValueDraft> {
        Binding(
            get: { model.observationDialog?.dateDraft ?? .empty() },
            set: {
                guard var draft = model.observationDialog else { return }
                draft.dateDraft = $0
                model.updateObservationDialog(draft)
            }
        )
    }

    // MARK: - Helpers

    private func artifactDisplayName(_ artifact: CatalogArtifact) -> String {
        artifact.label.isEmpty ? artifact.ref : artifact.label
    }

    private func artifactMediaCaption(_ artifact: CatalogArtifact) -> String {
        if CitationComposerModel.isPDF(artifact) {
            return String(localized: L10n.CitationComposer.mediaCaptionPDF)
        }
        if CitationComposerModel.isImage(artifact) {
            return String(localized: L10n.CitationComposer.mediaCaptionImage)
        }
        return artifact.ref
    }
}
