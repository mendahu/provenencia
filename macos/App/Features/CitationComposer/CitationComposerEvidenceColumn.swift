import SwiftUI

/// Evidence column: artifact and citation steppers, the reading, then the file.
struct CitationComposerEvidenceColumn: View {
    @Bindable var model: CitationComposerModel
    @FocusState private var stepperFocus: StepperFocus?

    private enum StepperFocus: Hashable {
        case artifact, citation
    }

    var body: some View {
        VStack(alignment: .leading, spacing: 0) {
            steppers
                .padding(.horizontal, PVSpacing.space6)
                .padding(.top, PVSpacing.space6)
                .padding(.bottom, PVSpacing.space4)
            reading
                .padding(.horizontal, PVSpacing.space6)
                .padding(.bottom, PVSpacing.space4)
            CitationComposerViewerPane(model: model)
        }
        .background(PVColor.surfaceSunken)
    }

    private var steppers: some View {
        HStack(alignment: .top, spacing: PVSpacing.space4) {
            artifactStepper
            citationStepper
        }
    }

    private var artifactStepper: some View {
        stepperCard(focused: stepperFocus == .artifact) {
            HStack(alignment: .top, spacing: PVSpacing.space4) {
                artifactThumbnail
                VStack(alignment: .leading, spacing: PVSpacing.space1) {
                    Text(verbatim: artifactTitle)
                        .font(PVFont.body(size: PVTypeScale.bodySmall, weight: PVFontWeight.medium))
                        .foregroundStyle(PVColor.textPrimary)
                        .lineLimit(2)
                    Text(verbatim: artifactMeta)
                        .font(PVFont.body(size: PVTypeScale.caption))
                        .foregroundStyle(PVColor.textMuted)
                        .lineLimit(1)
                }
                Spacer(minLength: 0)
            }
        } controls: {
            stepperControls(
                position: artifactPosition,
                canBack: model.canStepArtifactBackward && !model.isTranscribing,
                canForward: model.canStepArtifactForward && !model.isTranscribing,
                backLabel: L10n.CitationComposer.previousArtifact,
                forwardLabel: L10n.CitationComposer.nextArtifact,
                onBack: { model.stepArtifact(by: -1) },
                onForward: { model.stepArtifact(by: 1) }
            )
        }
        .focusable()
        .focused($stepperFocus, equals: .artifact)
        .onKeyPress(.leftArrow) {
            guard stepperFocus == .artifact else { return .ignored }
            model.stepArtifact(by: -1)
            return .handled
        }
        .onKeyPress(.rightArrow) {
            guard stepperFocus == .artifact else { return .ignored }
            model.stepArtifact(by: 1)
            return .handled
        }
        .accessibilityIdentifier("citationComposer.artifactStepper")
    }

    private var citationStepper: some View {
        VStack(alignment: .leading, spacing: PVSpacing.space3) {
            stepperCard(focused: stepperFocus == .citation) {
                VStack(alignment: .leading, spacing: PVSpacing.space2) {
                    HStack(spacing: PVSpacing.space3) {
                        Text(verbatim: citationTitle)
                            .font(PVFont.mono(size: PVTypeScale.caption, weight: PVFontWeight.semibold))
                            .foregroundStyle(PVColor.textPrimary)
                            .lineLimit(1)
                        Text(verbatim: citationCountLabel)
                            .font(PVFont.body(size: PVTypeScale.micro))
                            .foregroundStyle(PVColor.textMuted)
                            .lineLimit(1)
                        Spacer(minLength: 0)
                    }
                    if let names = citationNames {
                        Text(verbatim: names)
                            .font(PVFont.body(size: PVTypeScale.caption))
                            .foregroundStyle(PVColor.textSecondary)
                            .lineLimit(1)
                    } else if model.activeCitationID != nil {
                        Text(L10n.CitationComposer.noSubjectsYet)
                            .font(PVFont.body(size: PVTypeScale.caption, italic: true))
                            .foregroundStyle(PVColor.textMuted)
                            .lineLimit(1)
                    }
                    citationPreview
                        .lineLimit(2)
                }
            } controls: {
                stepperControls(
                    position: citationPosition,
                    canBack: model.canStepCitationBackward && !model.isTranscribing,
                    canForward: model.canStepCitationForward && !model.isTranscribing,
                    backLabel: L10n.CitationComposer.previousCitation,
                    forwardLabel: L10n.CitationComposer.nextCitation,
                    onBack: { model.stepCitation(by: -1) },
                    onForward: { model.stepCitation(by: 1) }
                )
            }
            PVButton(L10n.CitationComposer.newCitation, variant: .ghost, size: .sm, icon: .plus) {
                model.beginNewCitation()
            }
            .disabled(model.isTranscribing)
            .accessibilityIdentifier("citationComposer.citation.new")
        }
        .focusable()
        .focused($stepperFocus, equals: .citation)
        .onKeyPress(.leftArrow) {
            guard stepperFocus == .citation else { return .ignored }
            model.stepCitation(by: -1)
            return .handled
        }
        .onKeyPress(.rightArrow) {
            guard stepperFocus == .citation else { return .ignored }
            model.stepCitation(by: 1)
            return .handled
        }
        .accessibilityIdentifier("citationComposer.citationStepper")
    }

    private func stepperCard<Content: View, Controls: View>(
        focused: Bool,
        @ViewBuilder content: () -> Content,
        @ViewBuilder controls: () -> Controls
    ) -> some View {
        PVCard(tone: .raised, padding: PVSpacing.space4) {
            VStack(alignment: .leading, spacing: PVSpacing.space3) {
                content()
                controls()
            }
        }
        .overlay(
            RoundedRectangle(cornerRadius: PVRadius.md, style: .continuous)
                .stroke(focused ? PVColor.borderFocus : Color.clear, lineWidth: 1)
        )
    }

    private func stepperControls(
        position: String,
        canBack: Bool,
        canForward: Bool,
        backLabel: LocalizedStringResource,
        forwardLabel: LocalizedStringResource,
        onBack: @escaping () -> Void,
        onForward: @escaping () -> Void
    ) -> some View {
        HStack(spacing: PVSpacing.space2) {
            PVIconButton(.chevronBack, label: backLabel, size: .sm, action: onBack)
                .disabled(!canBack)
            Text(verbatim: position)
                .font(PVFont.mono(size: PVTypeScale.micro))
                .foregroundStyle(PVColor.textMuted)
                .frame(maxWidth: .infinity)
            PVIconButton(.chevronForward, label: forwardLabel, size: .sm, action: onForward)
                .disabled(!canForward)
        }
    }

    private var artifactThumbnail: some View {
        let content: PVThumbnail.Content
        if let image = model.artifactViewer.displayImage, model.artifactViewer.kind == .image {
            content = PVThumbnail.Content(image: Image(nsImage: image))
        } else {
            content = PVThumbnail.Content(mark: artifactMark)
        }
        return PVThumbnail(content, size: 44)
    }

    private var artifactMark: PVMarkKey? {
        PVFileTypeGlyph.key(
            mediaType: model.selectedArtifact?.file?.mediaType,
            originalFilename: model.selectedArtifact?.file?.originalFilename
        )
    }

    private var artifactTitle: String {
        let label = model.selectedArtifact?.label.trimmingCharacters(in: .whitespacesAndNewlines) ?? ""
        if !label.isEmpty { return label }
        return model.selectedArtifact?.file?.originalFilename ?? ""
    }

    private var artifactMeta: String {
        let kind = artifactKindLabel
        let detail = artifactDetail
        let count = model.selectedArtifact.map { model.listedCount(for: $0.id) } ?? 0
        return L10n.CitationComposer.artifactStepperMeta(kind: kind, detail: detail, count: count)
    }

    private var artifactKindLabel: String {
        switch model.artifactViewer.kind {
        case .pdf: L10n.string(L10n.CitationComposer.artifactKindPDF)
        case .image: L10n.string(L10n.CitationComposer.artifactKindImage)
        default: L10n.string(L10n.CitationComposer.artifactKindUnknown)
        }
    }

    private var artifactDetail: String {
        if model.artifactViewer.kind == .pdf, model.artifactViewer.pageCount > 0 {
            return L10n.CitationComposer.pageCount(count: model.artifactViewer.pageCount)
        }
        let bytes = model.selectedArtifact?.file?.byteSize ?? 0
        return ByteCountFormatter.string(fromByteCount: bytes, countStyle: .file)
    }

    private var artifactPosition: String {
        guard let index = model.selectedArtifactIndex else { return "" }
        return L10n.CitationComposer.stepperPosition(index: index + 1, total: model.artifacts.count)
    }

    private var citationTitle: String {
        if model.activeCitationID == nil {
            return L10n.string(L10n.CitationComposer.newCitation)
        }
        return model.activeCitationRef
    }

    private var citationCountLabel: String {
        guard let listed = currentListed else {
            return L10n.string(L10n.CitationComposer.noneYet)
        }
        return L10n.CitationComposer.citationMenuCount(count: listed.observationCount)
    }

    private var citationNames: String? {
        guard let listed = currentListed else { return nil }
        return CitationComposerAuthoring.subjectNameSummary(listed.subjectNames)
    }

    @ViewBuilder
    private var citationPreview: some View {
        switch CitationComposerAuthoring.preview(
            transcription: model.transcription,
            description: model.citationDescription
        ) {
        case .text(let text):
            Text(verbatim: text)
                .font(PVFont.body(size: PVTypeScale.caption))
                .foregroundStyle(PVColor.textSecondary)
        case .empty:
            Text(L10n.CitationComposer.noTranscriptionYet)
                .font(PVFont.body(size: PVTypeScale.caption, italic: true))
                .foregroundStyle(PVColor.textMuted)
        }
    }

    private var citationPosition: String {
        if model.activeCitationID == nil {
            return L10n.string(L10n.CitationComposer.newCitation)
        }
        let ordered = model.orderedCitations
        guard let index = ordered.firstIndex(where: { $0.citation.id == model.activeCitationID }) else {
            return ""
        }
        return L10n.CitationComposer.stepperPosition(index: index + 1, total: ordered.count)
    }

    private var currentListed: CatalogListedCitation? {
        model.orderedCitations.first { $0.citation.id == model.activeCitationID }
    }

    private var reading: some View {
        VStack(alignment: .leading, spacing: PVSpacing.space4) {
            HStack(alignment: .top, spacing: PVSpacing.space4) {
                PVField(label: L10n.CitationComposer.transcriptionLabel, error: model.fields.error) {
                    PVTextArea(text: $model.transcription, lineLimit: 4...4)
                        .accessibilityIdentifier("citationComposer.transcription")
                }
                PVField(label: L10n.CitationComposer.descriptionLabel) {
                    PVTextArea(text: $model.citationDescription, lineLimit: 4...4)
                        .accessibilityIdentifier("citationComposer.description")
                }
            }
            HStack(spacing: PVSpacing.space4) {
                if model.canAutoTranscribe {
                    PVButton(
                        L10n.CitationComposer.autoTranscribe,
                        variant: .secondary,
                        size: .sm,
                        icon: .scanText,
                        loading: model.isTranscribing
                    ) {
                        Task { await model.runAutoTranscribe() }
                    }
                    .disabled(!model.canAutoTranscribe || model.isTranscribing)
                } else if model.canPasteTranscription {
                    PVButton(
                        L10n.CitationComposer.pasteTranscription,
                        variant: .secondary,
                        size: .sm,
                        icon: .clipboardPaste
                    ) {
                        model.requestPasteTranscription()
                    }
                    .disabled(model.isTranscribing)
                }
                PVCheckbox(
                    L10n.CitationComposer.uncertainLabel,
                    isChecked: $model.transcriptionUncertain
                )
                Spacer(minLength: 0)
                if model.activeCitationID != nil {
                    PVButton(L10n.CitationComposer.deleteCitation, variant: .ghost, size: .sm) {
                        Task { await model.askDeleteCitation() }
                    }
                    .disabled(model.isTranscribing)
                }
                PVButton(L10n.CitationComposer.save, variant: .primary, size: .sm, loading: model.fields.isSaving) {
                    Task { await model.fields.saveCitation() }
                }
                .disabled(model.fields.isSaving || model.isTranscribing)
                .accessibilityIdentifier("citationComposer.saveCitation")
            }
            if model.transcriptionUncertain {
                PVField(label: L10n.CitationComposer.uncertainNoteLabel) {
                    PVTextArea(text: $model.transcriptionNote, lineLimit: 2...3)
                        .accessibilityIdentifier("citationComposer.transcriptionNote")
                }
            }
            if let message = model.transcriptionOCRMessage, !message.isEmpty {
                PVCallout(tone: .warning, message: message)
            }
        }
    }
}
