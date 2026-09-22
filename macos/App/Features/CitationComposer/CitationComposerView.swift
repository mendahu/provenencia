import SwiftUI

/// Citation composer place (S7-08 thin): viewer placeholder | form, Artifact pick, submit.
struct CitationComposerView: View {
    let sourceID: String
    let subjectID: String
    let session: WorkspaceSession
    let store: any GenealogyStore
    let userID: String

    @Environment(WorkspaceNavigation.self) private var navigation
    @State private var model: CitationComposerModel
    @State private var dateEditorRowID: UUID?
    @State private var customTermPropertyID: String?
    @State private var customTermLabel = ""

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
            case .noArtifacts:
                noArtifactsState
            case .pickArtifact:
                artifactPicker
            case .compose:
                composeSplit
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
            isPresented: dateEditorBinding,
            copy: PVFormDialogCopy(
                title: L10n.CitationComposer.dateDialogTitle,
                confirm: L10n.CitationComposer.dateDialogConfirm,
                cancel: L10n.CitationComposer.cancel
            ),
            isRunning: false,
            confirmDisabled: !dateDraftIsValid,
            accessibilityIdentifierPrefix: "citationComposer.date",
            onConfirm: { dateEditorRowID = nil }
        ) {
            if let id = dateEditorRowID,
               let index = model.observations.firstIndex(where: { $0.id == id })
            {
                DateValueEditorForm(
                    draft: dateDraftBinding(at: index),
                    accessibilityIdentifierPrefix: "citationComposer.date"
                )
            }
        }
        .pvFormDialog(
            isPresented: customTermBinding,
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
                    guard let propertyID = customTermPropertyID else { return }
                    if let term = await model.createCustomTerm(
                        propertyID: propertyID,
                        label: customTermLabel
                    ),
                       let index = model.observations.firstIndex(where: {
                           $0.propertyID == propertyID
                       })
                    {
                        var row = model.observations[index]
                        row.valueTermID = term.id
                        model.updateObservation(row)
                    }
                    customTermPropertyID = nil
                    customTermLabel = ""
                }
            }
        ) {
            PVField(label: L10n.CitationComposer.addTermLabel) {
                PVInput(text: $customTermLabel, size: .sm)
                    .accessibilityIdentifier("citationComposer.term.label")
            }
        }
    }

    // MARK: - States

    private var noArtifactsState: some View {
        VStack(alignment: .leading, spacing: PVSpacing.space7) {
            Text(L10n.CitationComposer.noArtifactsTitle)
                .font(PVFont.display(size: PVTypeScale.h2, weight: PVFontWeight.medium))
                .foregroundStyle(PVColor.textDisplay)
            Text(L10n.CitationComposer.noArtifactsMessage)
                .font(PVFont.body(size: PVTypeScale.body))
                .foregroundStyle(PVColor.textSecondary)
            PVButton(L10n.CitationComposer.backToGraph, variant: .secondary) {
                navigation.go(to: model.graphLocation())
            }
            Spacer(minLength: 0)
        }
        .padding(PVSpacing.space9)
        .frame(maxWidth: .infinity, maxHeight: .infinity, alignment: .topLeading)
    }

    private var artifactPicker: some View {
        VStack(alignment: .leading, spacing: PVSpacing.space7) {
            Text(L10n.CitationComposer.pickArtifactTitle)
                .font(PVFont.display(size: PVTypeScale.h2, weight: PVFontWeight.medium))
                .foregroundStyle(PVColor.textDisplay)
            Text(L10n.CitationComposer.pickArtifactMessage)
                .font(PVFont.body(size: PVTypeScale.body))
                .foregroundStyle(PVColor.textSecondary)
            LazyVGrid(
                columns: [GridItem(.adaptive(minimum: 180), spacing: PVSpacing.space5)],
                spacing: PVSpacing.space5
            ) {
                ForEach(model.artifacts, id: \.id) { artifact in
                    artifactTile(artifact)
                }
            }
            Spacer(minLength: 0)
        }
        .padding(PVSpacing.space9)
        .frame(maxWidth: .infinity, maxHeight: .infinity, alignment: .topLeading)
    }

    private func artifactTile(_ artifact: CatalogArtifact) -> some View {
        let selected = model.selectedArtifactID == artifact.id
        return Button {
            model.selectArtifact(artifact.id)
        } label: {
            VStack(alignment: .leading, spacing: PVSpacing.space3) {
                RoundedRectangle(cornerRadius: PVRadius.sm, style: .continuous)
                    .strokeBorder(
                        selected ? PVColor.accent : PVColor.borderDefault,
                        lineWidth: selected ? 2 : 1
                    )
                    .background(
                        RoundedRectangle(cornerRadius: PVRadius.sm, style: .continuous)
                            .fill(PVColor.surfaceSunken)
                    )
                    .frame(height: 120)
                    .overlay {
                        Text(verbatim: artifactMediaLabel(artifact))
                            .font(PVFont.body(size: PVTypeScale.caption))
                            .foregroundStyle(PVColor.textMuted)
                            .multilineTextAlignment(.center)
                            .padding(PVSpacing.space4)
                    }
                Text(verbatim: artifactDisplayName(artifact))
                    .font(PVFont.body(size: PVTypeScale.caption, weight: PVFontWeight.medium))
                    .foregroundStyle(PVColor.textPrimary)
                    .lineLimit(2)
            }
        }
        .buttonStyle(.plain)
        .accessibilityLabel(Text(verbatim: artifactDisplayName(artifact)))
    }

    private var composeSplit: some View {
        HStack(alignment: .top, spacing: 0) {
            viewerPane
                .frame(minWidth: 280, idealWidth: 420, maxWidth: 560)
            PVDivider()
            formPane
                .frame(maxWidth: .infinity)
        }
    }

    private var viewerPane: some View {
        VStack(alignment: .leading, spacing: PVSpacing.space5) {
            HStack {
                if let artifact = model.selectedArtifact {
                    Text(verbatim: artifactDisplayName(artifact))
                        .font(PVFont.body(size: PVTypeScale.body, weight: PVFontWeight.medium))
                        .foregroundStyle(PVColor.textDisplay)
                }
                Spacer()
                if model.artifacts.count > 1 {
                    PVButton(L10n.CitationComposer.changeArtifact, variant: .ghost, size: .sm) {
                        model.changeArtifact()
                    }
                }
            }
            RoundedRectangle(cornerRadius: PVRadius.md, style: .continuous)
                .strokeBorder(PVColor.borderDefault, style: StrokeStyle(lineWidth: 1, dash: [6, 4]))
                .background(
                    RoundedRectangle(cornerRadius: PVRadius.md, style: .continuous)
                        .fill(PVColor.surfaceSunken)
                )
                .overlay {
                    VStack(spacing: PVSpacing.space3) {
                        Text(L10n.CitationComposer.viewerPlaceholderTitle)
                            .font(PVFont.body(size: PVTypeScale.body, weight: PVFontWeight.medium))
                            .foregroundStyle(PVColor.textMuted)
                        Text(L10n.CitationComposer.viewerPlaceholderMessage)
                            .font(PVFont.body(size: PVTypeScale.caption))
                            .foregroundStyle(PVColor.textFaint)
                            .multilineTextAlignment(.center)
                    }
                    .padding(PVSpacing.space7)
                }
                .frame(maxWidth: .infinity, maxHeight: .infinity)
        }
        .padding(PVSpacing.space7)
    }

    private var formPane: some View {
        VStack(alignment: .leading, spacing: 0) {
            ScrollView {
                VStack(alignment: .leading, spacing: PVSpacing.space8) {
                    citationFields
                    observationsSection
                    if let formError = model.formError {
                        PVCallout(tone: .danger, message: formError)
                    }
                }
                .padding(PVSpacing.space7)
            }
            PVDivider()
            footer
        }
    }

    private var citationFields: some View {
        VStack(alignment: .leading, spacing: PVSpacing.space5) {
            Text(L10n.CitationComposer.citationSection)
                .font(PVFont.body(size: PVTypeScale.body, weight: PVFontWeight.semibold))
                .foregroundStyle(PVColor.textDisplay)
            PVField(label: L10n.CitationComposer.transcriptionLabel) {
                PVTextArea(text: Bindable(model).transcription, lineLimit: 2...6)
                    .accessibilityIdentifier("citationComposer.transcription")
            }
            Toggle(isOn: Bindable(model).transcriptionUncertain) {
                Text(L10n.CitationComposer.uncertainLabel)
                    .font(PVFont.body(size: PVTypeScale.body))
            }
            .toggleStyle(.checkbox)
            .accessibilityIdentifier("citationComposer.uncertain")
            if model.transcriptionUncertain {
                PVField(label: L10n.CitationComposer.uncertainNoteLabel) {
                    PVTextArea(text: Bindable(model).transcriptionNote, lineLimit: 1...3)
                        .accessibilityIdentifier("citationComposer.uncertainNote")
                }
            }
            PVField(label: L10n.CitationComposer.descriptionLabel) {
                PVTextArea(text: Bindable(model).citationDescription, lineLimit: 1...4)
                    .accessibilityIdentifier("citationComposer.description")
            }
        }
    }

    private var observationsSection: some View {
        VStack(alignment: .leading, spacing: PVSpacing.space5) {
            PVSectionHeader(
                title: L10n.CitationComposer.observationsSection,
                meta: "\(model.observations.count)"
            ) {
                EmptyView()
            } actions: {
                PVButton(L10n.CitationComposer.addObservation, variant: .secondary, size: .sm) {
                    model.addObservation()
                }
                .accessibilityIdentifier("citationComposer.addObservation")
            }
            if model.observations.isEmpty, model.submitAttempted {
                PVCallout(
                    tone: .danger,
                    message: String(localized: L10n.CitationComposer.noObservationsError)
                )
            }
            ForEach(model.observations) { row in
                observationCard(row)
            }
        }
    }

    private func observationCard(_ row: CitationComposerModel.ObservationRow) -> some View {
        let property = model.catalogProperty(id: row.propertyID)
        return VStack(alignment: .leading, spacing: PVSpacing.space4) {
            HStack(alignment: .top) {
                PVComboBox(
                    selection: propertyIDBinding(row.id),
                    options: model.propertyOptions,
                    size: .sm,
                    placeholder: L10n.CitationComposer.propertyPlaceholder,
                    emptyLabel: L10n.CitationComposer.propertyEmpty,
                    label: L10n.CitationComposer.propertyLabel,
                    accessibilityIdentifierPrefix: "citationComposer.property.\(row.id.uuidString)"
                )
                .frame(maxWidth: .infinity)
                PVButton(L10n.CitationComposer.removeObservation, variant: .ghost, size: .sm) {
                    model.removeObservation(id: row.id)
                }
            }
            polarityChips(row)
            if let property {
                valueEditor(row: row, property: property)
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

    private func polarityChips(_ row: CitationComposerModel.ObservationRow) -> some View {
        VStack(alignment: .leading, spacing: PVSpacing.space2) {
            Text(L10n.CitationComposer.polarityLabel)
                .font(PVFont.body(size: PVTypeScale.caption, weight: PVFontWeight.medium))
                .foregroundStyle(PVColor.textSecondary)
            PVChipGroup(style: .segmented) {
                PVChip(
                    L10n.CitationComposer.polarityAsserts,
                    isSelected: row.polarity != "negative",
                    expands: true,
                    selectionLift: true,
                    action: {
                        var next = row
                        next.polarity = "positive"
                        model.updateObservation(next)
                    }
                )
                PVChip(
                    L10n.CitationComposer.polarityNegates,
                    isSelected: row.polarity == "negative",
                    expands: true,
                    selectionLift: true,
                    action: {
                        var next = row
                        next.polarity = "negative"
                        model.updateObservation(next)
                    }
                )
            }
        }
    }

    @ViewBuilder
    private func valueEditor(
        row: CitationComposerModel.ObservationRow,
        property: CatalogProperty
    ) -> some View {
        switch property.valueType {
        case "text":
            PVField(label: L10n.CitationComposer.valueLabel) {
                PVTextArea(text: textBinding(row.id), lineLimit: 1...4)
                    .accessibilityIdentifier("citationComposer.value.\(row.id.uuidString)")
            }
        case "integer":
            PVField(label: L10n.CitationComposer.integerLabel) {
                PVInput(text: integerBinding(row.id), size: .sm)
                    .accessibilityIdentifier("citationComposer.integer.\(row.id.uuidString)")
            }
        case "term":
            VStack(alignment: .leading, spacing: PVSpacing.space3) {
                PVComboBox(
                    selection: termBinding(row.id),
                    options: model.termOptions(for: property.id),
                    size: .sm,
                    placeholder: L10n.CitationComposer.termPlaceholder,
                    emptyLabel: L10n.CitationComposer.termEmpty,
                    label: L10n.CitationComposer.termLabel,
                    accessibilityIdentifierPrefix: "citationComposer.term.\(row.id.uuidString)"
                )
                PVButton(L10n.CitationComposer.addCustomTerm, variant: .ghost, size: .sm) {
                    customTermPropertyID = property.id
                    customTermLabel = ""
                }
            }
        case "date":
            HStack {
                Text(verbatim: dateSummary(row.dateDraft))
                    .font(PVFont.body(size: PVTypeScale.body))
                    .foregroundStyle(PVColor.textPrimary)
                Spacer()
                PVButton(L10n.CitationComposer.editDate, variant: .secondary, size: .sm) {
                    dateEditorRowID = row.id
                }
            }
        default:
            Text(L10n.CitationComposer.unsupportedValueTypeError)
                .font(PVFont.body(size: PVTypeScale.caption))
                .foregroundStyle(PVColor.textMuted)
        }
    }

    private var footer: some View {
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
            .disabled(model.isSubmitting)
            .accessibilityIdentifier("citationComposer.save")
        }
        .padding(PVSpacing.space7)
    }

    // MARK: - Bindings

    private var dateEditorBinding: Binding<Bool> {
        Binding(
            get: { dateEditorRowID != nil },
            set: { if !$0 { dateEditorRowID = nil } }
        )
    }

    private var customTermBinding: Binding<Bool> {
        Binding(
            get: { customTermPropertyID != nil },
            set: {
                if !$0 {
                    customTermPropertyID = nil
                    customTermLabel = ""
                }
            }
        )
    }

    private var dateDraftIsValid: Bool {
        guard let id = dateEditorRowID,
              let row = model.observations.first(where: { $0.id == id })
        else { return false }
        return row.dateDraft.isValid
    }

    private func dateDraftBinding(at index: Int) -> Binding<DateValueDraft> {
        Binding(
            get: { model.observations[index].dateDraft },
            set: {
                var row = model.observations[index]
                row.dateDraft = $0
                model.updateObservation(row)
            }
        )
    }

    private func propertyIDBinding(_ id: UUID) -> Binding<String> {
        Binding(
            get: { model.observations.first(where: { $0.id == id })?.propertyID ?? "" },
            set: {
                guard var row = model.observations.first(where: { $0.id == id }) else { return }
                row.propertyID = $0
                model.updateObservation(row)
            }
        )
    }

    private func textBinding(_ id: UUID) -> Binding<String> {
        Binding(
            get: { model.observations.first(where: { $0.id == id })?.valueText ?? "" },
            set: {
                guard var row = model.observations.first(where: { $0.id == id }) else { return }
                row.valueText = $0
                model.updateObservation(row)
            }
        )
    }

    private func integerBinding(_ id: UUID) -> Binding<String> {
        Binding(
            get: { model.observations.first(where: { $0.id == id })?.valueIntegerText ?? "" },
            set: {
                guard var row = model.observations.first(where: { $0.id == id }) else { return }
                row.valueIntegerText = $0
                model.updateObservation(row)
            }
        )
    }

    private func termBinding(_ id: UUID) -> Binding<String> {
        Binding(
            get: { model.observations.first(where: { $0.id == id })?.valueTermID ?? "" },
            set: {
                guard var row = model.observations.first(where: { $0.id == id }) else { return }
                row.valueTermID = $0
                model.updateObservation(row)
            }
        )
    }

    private func artifactDisplayName(_ artifact: CatalogArtifact) -> String {
        artifact.label.isEmpty ? artifact.ref : artifact.label
    }

    private func artifactMediaLabel(_ artifact: CatalogArtifact) -> String {
        let media = artifact.file?.mediaType ?? ""
        if media.contains("pdf") { return "PDF" }
        if media.hasPrefix("image/") { return "Image" }
        return artifact.ref
    }

    private func dateSummary(_ draft: DateValueDraft) -> String {
        if !draft.isValid {
            return String(localized: L10n.CitationComposer.dateUnset)
        }
        let phrase = draft.phrase.trimmingCharacters(in: .whitespacesAndNewlines)
        if !phrase.isEmpty { return phrase }
        if let y = draft.startYear {
            if let m = draft.startMonth, let d = draft.startDay {
                return "\(y)-\(m)-\(d)"
            }
            if let m = draft.startMonth {
                return "\(y)-\(m)"
            }
            return "\(y)"
        }
        return String(localized: L10n.CitationComposer.dateUnset)
    }
}
