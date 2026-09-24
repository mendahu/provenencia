import SwiftUI

/// Right sidebar: identity line, reading fields, Observation list, Cancel / Save.
struct CitationComposerFormPane: View {
    @Bindable var model: CitationComposerModel
    var inert: Bool
    var wide: Bool = false
    var onAddCustomTerm: (UUID) -> Void = { _ in }

    @Environment(WorkspaceNavigation.self) private var navigation
    @State private var artifactMenu = PVContextMenuState()
    @State private var citationMenu = PVContextMenuState()

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
        .pvContextMenu($artifactMenu, dismissOnClickAway: true) {
            artifactMenuPanel
        }
        .pvContextMenu($citationMenu, dismissOnClickAway: true) {
            citationMenuPanel
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
                artifactTrigger
            }
            Spacer(minLength: 0)
            citationTrigger
        }
        .accessibilityElement(children: .contain)
    }

    private var artifactTrigger: some View {
        Button {
            artifactMenu.present(at: .zero)
        } label: {
            HStack(spacing: 6) {
                PVMark(artifactMarkKey, size: 16, decorative: true)
                if let title = model.selectedArtifact?.label {
                    Text(verbatim: title)
                        .lineLimit(1)
                }
                PVIcon(.sortUnsorted, size: 12)
            }
        }
        .buttonStyle(.pv(.ghost, size: .sm))
        .disabled(inert)
        .accessibilityLabel(Text(verbatim: L10n.CitationComposer.artifactMenuLabel(title: model.selectedArtifact?.label ?? "")))
        .accessibilityIdentifier("citationComposer.artifact")
    }

    private var citationTrigger: some View {
        Button {
            citationMenu.present(at: .zero)
        } label: {
            HStack(spacing: 6) {
                if model.activeCitationRef.isEmpty {
                    Text(L10n.CitationComposer.newCitation)
                } else {
                    Text(verbatim: model.activeCitationRef)
                        .font(PVFont.mono(size: PVTypeScale.caption))
                }
                PVIcon(.chevronDown, size: 12)
            }
        }
        .buttonStyle(.pv(.secondary, size: .sm))
        .disabled(inert)
        .accessibilityLabel(
            Text(verbatim: model.activeCitationRef.isEmpty
                ? String(localized: L10n.CitationComposer.citationMenuNew)
                : L10n.CitationComposer.citationMenuRef(ref: model.activeCitationRef, count: model.observations.count))
        )
        .accessibilityIdentifier("citationComposer.citation")
    }

    private var artifactMarkKey: PVMarkKey {
        PVFileTypeGlyph.key(
            mediaType: model.selectedArtifact?.file?.mediaType,
            originalFilename: model.selectedArtifact?.file?.originalFilename
        )
    }

    private var artifactMenuPanel: some View {
        PVContextMenuPanel(
            title: L10n.CitationComposer.artifactMenuTitle,
            width: 320,
            accessibilityIdentifier: "citationComposer.artifact.menu"
        ) {
            ForEach(Array(model.artifacts.enumerated()), id: \.element.id) { index, artifact in
                CitationComposerArtifactMenuRow(
                    artifact: artifact,
                    citationCount: model.listedCount(for: artifact.id),
                    isSelected: artifact.id == model.selectedArtifactID,
                    sourceTypeIconKey: model.sourceTypeIconKey,
                    index: index
                ) {
                    model.requestSelectArtifact(artifact.id)
                }
            }
        }
    }

    private var citationMenuPanel: some View {
        PVContextMenuPanel(
            title: L10n.CitationComposer.citationMenuTitle,
            width: 400,
            maxHeight: 380,
            accessibilityIdentifier: "citationComposer.citation.menu"
        ) {
            PVContextMenuItem(
                L10n.CitationComposer.newCitation,
                index: 0,
                isSelected: model.activeCitationID == nil,
                showsSelectionMark: true,
                accessibilityIdentifier: "citationComposer.citation.new"
            ) {
                model.selectCitation(nil)
            }
            ForEach(Array(model.listedCitations.enumerated()), id: \.element.id) { index, listed in
                CitationComposerCitationMenuRow(
                    listed: listed,
                    isSelected: listed.id == model.activeCitationID,
                    index: index + 1
                ) {
                    model.selectCitation(listed.id)
                }
            }
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
                    onAddCustomTerm: { onAddCustomTerm(row.id) }
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
        .padding(.vertical, 14)
    }
}
