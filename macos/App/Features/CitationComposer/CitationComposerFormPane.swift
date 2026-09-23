import SwiftUI

/// Right sidebar: citation fields, Observation list, Cancel / Save footer.
struct CitationComposerFormPane: View {
    @Bindable var model: CitationComposerModel
    var inert: Bool

    @Environment(WorkspaceNavigation.self) private var navigation

    static let sidebarWidth: CGFloat = 400

    var body: some View {
        VStack(alignment: .leading, spacing: 0) {
            ScrollView {
                VStack(alignment: .leading, spacing: PVSpacing.space8) {
                    Text(L10n.CitationComposer.citationSection)
                        .font(PVFont.display(size: PVTypeScale.h2, weight: PVFontWeight.medium))
                        .foregroundStyle(PVColor.textDisplay)
                    citationFields
                    observationsSection
                }
                .padding(PVSpacing.space7)
            }
            .disabled(inert)
            .opacity(inert ? 0.55 : 1)
            PVDivider(color: PVColor.borderDefault)
            footer
        }
    }

    private var citationFields: some View {
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
            PVTextArea(text: $model.transcription, lineLimit: 2...6)
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
            HStack(alignment: .firstTextBaseline, spacing: PVSpacing.space3) {
                Text(L10n.CitationComposer.observationsSection)
                    .font(PVFont.body(size: PVTypeScale.body, weight: PVFontWeight.semibold))
                    .foregroundStyle(PVColor.textDisplay)
                if !model.observations.isEmpty {
                    Text(verbatim: "\(model.observations.count)")
                        .font(PVFont.mono(size: PVTypeScale.caption))
                        .foregroundStyle(PVColor.textMuted)
                }
                Spacer(minLength: 0)
            }

            if model.observations.isEmpty {
                PVCallout(
                    tone: .info,
                    message: String(localized: L10n.CitationComposer.noObservationsError),
                    compact: true
                )
            }

            ForEach(model.observations) { row in
                CitationComposerObservationRow(
                    row: row,
                    propertyLabel: model.catalogProperty(id: row.propertyID)?.label ?? "",
                    summary: model.observationSummary(for: row),
                    inert: inert,
                    onEdit: { model.beginEditObservation(row) },
                    onRemove: { model.removeObservation(id: row.id) }
                )
            }

            if !inert {
                PVButton(L10n.CitationComposer.addObservation, variant: .secondary, size: .sm) {
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
                PVButton(L10n.CitationComposer.cancel, variant: .secondary) {
                    navigation.go(to: model.graphLocation())
                }
                .disabled(model.isSubmitting)
                .accessibilityIdentifier("citationComposer.form.cancel")
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
}
