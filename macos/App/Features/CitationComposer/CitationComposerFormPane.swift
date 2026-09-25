import SwiftUI

/// Right sidebar: identity line, citation fields, connections, observations, Done.
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
                citationFields
                connectionsSection
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
                    citationFields
                }
                .padding(PVSpacing.space7)
            }
            .frame(maxWidth: .infinity, maxHeight: .infinity, alignment: .top)
            PVDivider(axis: .vertical, color: PVColor.borderDefault)
            ScrollView {
                VStack(alignment: .leading, spacing: PVSpacing.space8) {
                    connectionsSection
                    observationsSection
                }
                .padding(PVSpacing.space7)
            }
            .frame(maxWidth: .infinity, maxHeight: .infinity, alignment: .top)
        }
        .disabled(inert)
        .opacity(inert ? 0.55 : 1)
    }

    private var identityLine: some View {
        VStack(alignment: .leading, spacing: PVSpacing.space2) {
            HStack(alignment: .center, spacing: PVSpacing.space3) {
                if model.showsArtifactSwitcher {
                    artifactSelect
                }
                Spacer(minLength: 0)
                citationSelect
            }
            if model.identityMenusDisabled {
                Text(L10n.CitationComposer.identityMenusDisabledHint)
                    .font(PVFont.body(size: PVTypeScale.caption))
                    .foregroundStyle(PVColor.textMuted)
            }
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
            isDisabled: inert || model.identityMenusDisabled,
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
            isDisabled: inert || model.identityMenusDisabled,
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

    private var citationFields: some View {
        VStack(alignment: .leading, spacing: PVSpacing.space5) {
            PVField(
                label: L10n.CitationComposer.transcriptionLabel,
                hint: model.autoTranscribeHint
            ) {
                HStack(spacing: PVSpacing.space4) {
                    Toggle(isOn: $model.transcriptionUncertain) {
                        Text(L10n.CitationComposer.uncertainLabel)
                            .font(PVFont.body(size: PVTypeScale.caption))
                    }
                    .toggleStyle(.checkbox)
                    .disabled(inert || model.isTranscribing)
                    .accessibilityIdentifier("citationComposer.uncertain")
                    Spacer(minLength: 0)
                    PVButton(
                        model.isTranscribing
                            ? L10n.CitationComposer.autoTranscribeReading
                            : L10n.CitationComposer.autoTranscribe,
                        variant: .secondary,
                        size: .sm,
                        icon: .scanText,
                        loading: model.isTranscribing
                    ) {
                        model.requestAutoTranscribe()
                    }
                    .disabled(inert || !model.canAutoTranscribe)
                    .accessibilityIdentifier("citationComposer.autoTranscribe")
                    .accessibilityLabel(
                        model.locator.hasRegion
                            ? Text(L10n.CitationComposer.autoTranscribeDrawnRegion)
                            : Text(L10n.CitationComposer.autoTranscribe)
                    )
                    .accessibilityHint(Text(model.autoTranscribeHint))
                }
            }
            PVTextArea(text: $model.transcription, lineLimit: wide ? 6...8 : 2...6)
                .disabled(inert || model.isTranscribing)
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
            if let error = model.fields.error {
                PVCallout(tone: .danger, message: error, compact: true)
            }
            if let message = model.transcriptionOCRMessage {
                PVCallout(tone: .warning, message: message, compact: true) {
                    PVButton(
                        L10n.CitationComposer.autoTranscribeDismiss,
                        variant: .ghost,
                        size: .sm
                    ) {
                        model.dismissTranscriptionOCRMessage()
                    }
                    .accessibilityIdentifier("citationComposer.autoTranscribe.dismiss")
                }
                .accessibilityIdentifier("citationComposer.autoTranscribe.callout")
            }
            HStack {
                Text(verbatim: model.fields.statusText)
                    .font(PVFont.body(size: PVTypeScale.caption))
                    .foregroundStyle(PVColor.textMuted)
                Spacer(minLength: 0)
                PVButton(
                    L10n.CitationComposer.save,
                    variant: .secondary,
                    size: .sm,
                    loading: model.fields.isSaving
                ) {
                    Task { await model.fields.saveCitation() }
                }
                .disabled(inert || model.fields.isSaving || model.isTranscribing)
                .accessibilityIdentifier("citationComposer.saveCitation")
            }
        }
    }

    private var connectionsSection: some View {
        VStack(alignment: .leading, spacing: PVSpacing.space5) {
            ForEach(model.connections.rows) { row in
                CitationComposerConnectionRow(
                    row: row,
                    termOptions: row.termProperty.map { model.termOptions(for: $0.id) } ?? [],
                    inert: inert,
                    onTerm: { model.connections.applyTerm(connectionID: row.id, termID: $0) },
                    onSave: { Task { await model.connections.saveConnection() } },
                    onDiscard: { model.connections.discard() },
                    onCommitRole: { Task { await model.connections.commitRole(connectionID: row.id) } },
                    onRevertRole: { model.connections.revertRole(connectionID: row.id) }
                )
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
                    onSave: { Task { await model.observationRows.commit(rowID: row.id) } },
                    onRevert: { model.observationRows.revert(rowID: row.id) },
                    onRequestDelete: { model.observationRows.requestDelete(rowID: row.id) },
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
            if model.shouldHoldLeave {
                Text(verbatim: model.unsavedSummary)
                    .font(PVFont.body(size: PVTypeScale.caption))
                    .foregroundStyle(PVColor.textMuted)
                    .accessibilityIdentifier("citationComposer.unsavedSummary")
            }
            HStack {
                Spacer(minLength: 0)
                PVButton(L10n.CitationComposer.done, variant: .primary, size: .sm) {
                    navigation.go(to: model.graphLocation())
                }
                .accessibilityIdentifier("citationComposer.done")
            }
        }
        .padding(.horizontal, PVSpacing.space7)
        .padding(.vertical, PVSpacing.space6)
    }
}
