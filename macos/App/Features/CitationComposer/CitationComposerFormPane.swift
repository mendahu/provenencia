import SwiftUI

/// Right sidebar: identity line, reading fields, Observation list, Cancel / Save.
struct CitationComposerFormPane: View {
    @Bindable var model: CitationComposerModel
    var inert: Bool
    var wide: Bool = false

    @Environment(WorkspaceNavigation.self) private var navigation

    static let sidebarWidth: CGFloat = 520

    var body: some View {
        VStack(alignment: .leading, spacing: 0) {
            if wide {
                wideBody
            } else {
                narrowBody
            }
            PVDivider(color: PVColor.borderDefault)
            footer
        }
    }

    private var narrowBody: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: PVSpacing.space8) {
                identityLine
                readingFields
                observationsSection
            }
            .padding(PVSpacing.space7)
        }
        .disabled(inert)
        .opacity(inert ? 0.55 : 1)
    }

    private var wideBody: some View {
        HStack(alignment: .top, spacing: 0) {
            ScrollView {
                VStack(alignment: .leading, spacing: PVSpacing.space6) {
                    identityLine
                    readingFields
                }
                .padding(PVSpacing.space7)
            }
            .frame(maxWidth: .infinity, maxHeight: .infinity, alignment: .top)
            PVDivider(axis: .vertical, color: PVColor.borderDefault)
            ScrollView {
                observationsSection
                    .padding(PVSpacing.space7)
            }
            .frame(maxWidth: .infinity, maxHeight: .infinity, alignment: .top)
        }
        .disabled(inert)
        .opacity(inert ? 0.55 : 1)
    }

    private var identityLine: some View {
        HStack(alignment: .center, spacing: PVSpacing.space3) {
            if model.showsArtifactSwitcher {
                artifactSelect
            }
            Spacer(minLength: 0)
            citationSelect
        }
        .accessibilityElement(children: .contain)
    }

    private var artifactSelect: some View {
        PVSelect(
            selection: artifactBinding,
            options: artifactOptions,
            size: .sm,
            icon: .file,
            displayLabel: model.selectedArtifact?.label,
            menuWidth: 320,
            fillsWidth: false,
            rowHeight: 52,
            isDisabled: inert,
            accessibilitySpokenLabel: L10n.CitationComposer.artifactMenuLabel(
                title: model.selectedArtifact?.label ?? ""
            ),
            accessibilityIdentifier: "citationComposer.artifact"
        ) { option in
            if let artifact = model.artifacts.first(where: { $0.id == option.id }) {
                CitationComposerArtifactMenuRow(
                    artifact: artifact,
                    citationCount: model.listedCount(for: artifact.id)
                )
            } else {
                PVSelectPlainRow(text: option.label)
            }
        }
    }

    private var citationSelect: some View {
        PVSelect(
            selection: citationBinding,
            options: citationOptions,
            size: .sm,
            displayLabel: model.activeCitationRef.isEmpty
                ? String(localized: L10n.CitationComposer.newCitation)
                : model.activeCitationRef,
            menuWidth: 400,
            fillsWidth: false,
            maxVisibleRows: 8,
            rowHeight: 68,
            isDisabled: inert,
            accessibilitySpokenLabel: model.activeCitationRef.isEmpty
                ? String(localized: L10n.CitationComposer.citationMenuNew)
                : L10n.CitationComposer.citationMenuRef(
                    ref: model.activeCitationRef,
                    count: model.observations.count
                ),
            accessibilityIdentifier: "citationComposer.citation"
        ) { option in
            if let listed = model.listedCitations.first(where: { $0.id == option.id }) {
                CitationComposerCitationMenuRow(listed: listed)
            } else {
                PVSelectPlainRow(text: option.label)
            }
        }
    }

    private var artifactBinding: Binding<String> {
        Binding(
            get: { model.selectedArtifactID ?? "" },
            set: { model.requestSelectArtifact($0) }
        )
    }

    private var citationBinding: Binding<String> {
        Binding(
            get: { model.activeCitationID ?? "" },
            set: { model.selectCitation($0.isEmpty ? nil : $0) }
        )
    }

    private var artifactOptions: [PVSelectOption] {
        model.artifacts.map { artifact in
            PVSelectOption(
                value: artifact.id,
                label: artifact.label,
                accessibilityIdentifier: "citationComposer.artifact.row.\(artifact.id)"
            )
        }
    }

    private var citationOptions: [PVSelectOption] {
        [PVSelectOption(
            value: "",
            label: String(localized: L10n.CitationComposer.newCitation),
            accessibilityIdentifier: "citationComposer.citation.new"
        )] + model.listedCitations.map { listed in
            PVSelectOption(
                value: listed.id,
                label: listed.citation.ref,
                accessibilityIdentifier: "citationComposer.citation.row.\(listed.id)"
            )
        }
    }

    private var readingFields: some View {
        VStack(alignment: .leading, spacing: PVSpacing.space5) {
            PVField(label: L10n.CitationComposer.transcriptionLabel) {
                HStack(spacing: PVSpacing.space4) {
                    Toggle(isOn: $model.transcriptionUncertain) {
                        Text(L10n.CitationComposer.uncertainLabel)
                            .font(PVFont.body(size: PVTypeScale.caption))
                    }
                    .toggleStyle(.checkbox)
                    .disabled(inert)
                    .accessibilityIdentifier("citationComposer.uncertain")
                    Spacer(minLength: 0)
                }
            }
            PVTextArea(text: $model.transcription, lineLimit: wide ? 6...8 : 2...6)
                .disabled(inert)
                .accessibilityIdentifier("citationComposer.transcription")
            if model.transcriptionUncertain {
                PVField(label: L10n.CitationComposer.uncertainNoteLabel) {
                    PVTextArea(text: $model.transcriptionNote, lineLimit: 1...3)
                        .disabled(inert)
                        .accessibilityIdentifier("citationComposer.uncertainNote")
                }
            }
            PVField(label: L10n.CitationComposer.descriptionLabel) {
                PVTextArea(text: $model.citationDescription, lineLimit: 1...4)
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

    private var observationsSection: some View {
        VStack(alignment: .leading, spacing: PVSpacing.space5) {
            PVSectionHeader(
                title: L10n.CitationComposer.observationsSection,
                meta: "\(model.observations.count)"
            )
            ForEach(model.observations) { row in
                CitationComposerObservationRow(
                    row: row,
                    property: model.catalogProperty(id: row.propertyID),
                    propertyOptions: model.propertyOptions(for: row.subjectID),
                    subjectOptions: model.subjectOptions,
                    termOptions: model.termOptions(for: row.propertyID),
                    summary: model.observationSummary(for: row),
                    isFocused: model.focusedObservationID == row.id,
                    inert: inert,
                    onSubject: { model.updateObservationSubject(id: row.id, subjectID: $0) },
                    onProperty: { model.updateObservationProperty(id: row.id, propertyID: $0) },
                    onText: { model.updateObservationText(id: row.id, text: $0) },
                    onInteger: { model.updateObservationInteger(id: row.id, text: $0) },
                    onTerm: { model.updateObservationTerm(id: row.id, termID: $0) },
                    onEditValue: { model.beginEditObservation(row) },
                    onTogglePolarity: { model.toggleObservationPolarity(id: row.id) },
                    onRemove: { model.removeObservation(id: row.id) },
                    onAddCustomTerm: { model.beginAddCustomTerm(rowID: row.id) }
                )
            }

            if !inert {
                PVButton(L10n.CitationComposer.addObservation, variant: .ghost, size: .sm, icon: .plus) {
                    model.beginAddObservation()
                }
                .accessibilityIdentifier("citationComposer.addObservation")
            }
        }
    }

    private var footer: some View {
        VStack(alignment: .leading, spacing: PVSpacing.space4) {
            if let formError = model.formError, model.submitAttempted {
                PVCallout(tone: .danger, message: formError, compact: true)
            }
            HStack {
                Spacer(minLength: 0)
                PVButton(L10n.CitationComposer.cancel, variant: .ghost, size: .sm) {
                    navigation.go(to: model.graphLocation())
                }
                .disabled(model.isSubmitting)
                .accessibilityIdentifier("citationComposer.form.cancel")
                PVButton(
                    L10n.CitationComposer.save,
                    variant: .primary,
                    size: .sm,
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
        .padding(.horizontal, PVSpacing.space7)
        .padding(.vertical, PVSpacing.space6)
    }
}
