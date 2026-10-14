import SwiftUI

enum CitationComposerColumnLayout {
    /// Citation form in its own column beside the file.
    case side
    /// Citation form stacked above the file.
    case panel
}

/// Citation column: artifact and citation menus, then the reading.
struct CitationComposerEvidenceColumn: View {
    @Bindable var model: CitationComposerModel
    var layout: CitationComposerColumnLayout

    var body: some View {
        switch layout {
        case .side:
            VStack(alignment: .leading, spacing: 0) {
                PVSectionHeader(
                    title: L10n.CitationComposer.citationEyebrow,
                    meta: "\(model.orderedCitations.count)"
                )
                .padding(.horizontal, PVSpacing.space7)
                .padding(.top, 14)
                .padding(.bottom, PVSpacing.space4)
                ScrollView {
                    sideForm
                        .padding(.horizontal, PVSpacing.space7)
                        .padding(.top, PVSpacing.space2)
                        .padding(.bottom, PVSpacing.space7)
                }
            }
            .frame(maxHeight: .infinity, alignment: .topLeading)
            .background(PVColor.surfaceCard)
        case .panel:
            ScrollView {
                panelForm
                    .padding(.horizontal, PVSpacing.space6)
                    .padding(.top, PVSpacing.space5)
                    .padding(.bottom, 14)
            }
            .frame(maxWidth: .infinity, maxHeight: .infinity, alignment: .topLeading)
            .background(PVColor.surfaceCard)
            .overlay(alignment: .bottom) {
                PVDivider(axis: .horizontal, color: PVColor.borderSubtle)
            }
        }
    }

    private var sideForm: some View {
        VStack(alignment: .leading, spacing: PVSpacing.space7) {
            VStack(alignment: .leading, spacing: PVSpacing.space5) {
                artifactStepper
                citationStepper
            }
            PVDivider(axis: .horizontal, color: PVColor.borderSubtle)
            labelField
            transcriptionColumn(rows: 8...40)
            footer
        }
    }

    private var panelForm: some View {
        VStack(alignment: .leading, spacing: 10) {
            HStack(alignment: .bottom, spacing: PVSpacing.space5) {
                artifactStepper
                    .frame(maxWidth: 300)
                citationStepper
                    .frame(maxWidth: .infinity)
            }
            WeightedColumns(weights: [1.5, 1], spacing: PVSpacing.space6) {
                VStack(alignment: .leading, spacing: 10) {
                    transcriptionColumn(rows: 6...40)
                    footer
                }
                labelField
            }
        }
    }

    private var artifactStepper: some View {
        stepper(
            eyebrow: L10n.CitationComposer.artifactEyebrow,
            position: artifactPosition
        ) {
            artifactSelect
        }
    }

    private var citationStepper: some View {
        stepper(
            eyebrow: L10n.CitationComposer.citationEyebrow,
            position: citationPosition
        ) {
            citationSelect
        }
    }

    private func stepper<Control: View>(
        eyebrow: LocalizedStringResource,
        position: String,
        @ViewBuilder control: () -> Control
    ) -> some View {
        VStack(alignment: .leading, spacing: PVSpacing.space2) {
            HStack(spacing: PVSpacing.space4) {
                Text(eyebrow)
                    .pvMicroCaps()
                    .foregroundStyle(PVColor.textSecondary)
                Spacer(minLength: 0)
                Text(verbatim: position)
                    .font(PVFont.mono(size: 11))
                    .foregroundStyle(PVColor.textMuted)
            }
            .frame(height: 20)
            control()
        }
    }

    private var citationSelect: some View {
        PVSelect(
            selection: citationSelection,
            options: citationOptions,
            size: .md,
            displayLabel: model.citationStepLabel(for: model.activeCitationID),
            menuWidth: 480,
            maxVisibleRows: 6,
            rowHeight: 84,
            isDisabled: model.isTranscribing,
            accessibilityLabel: L10n.CitationComposer.citationEyebrow,
            accessibilityIdentifier: "citationComposer.citationSelect"
        ) { option in
            if option.id.isEmpty {
                HStack(spacing: PVSpacing.space4) {
                    PVIcon(.plus, size: 14)
                    Text(L10n.CitationComposer.newCitation)
                        .font(PVFont.body(size: PVTypeScale.caption))
                    Text(L10n.CitationComposer.onThisArtifact)
                        .font(PVFont.body(size: 12, italic: true))
                        .foregroundStyle(PVColor.textMuted)
                }
            } else if let listed = model.orderedCitations.first(where: { $0.citation.id == option.id }) {
                citationMenuRow(listed)
            } else {
                Text(verbatim: option.label)
            }
        }
    }

    private var artifactSelect: some View {
        PVSelect(
            selection: artifactSelection,
            options: artifactOptions,
            size: .md,
            menuWidth: 360,
            maxVisibleRows: 6,
            rowHeight: 64,
            isDisabled: model.isTranscribing || model.artifacts.isEmpty,
            accessibilityLabel: L10n.CitationComposer.artifactEyebrow,
            accessibilityIdentifier: "citationComposer.artifactSelect"
        ) { option in
            if let artifact = model.artifacts.first(where: { $0.id == option.id }) {
                artifactMenuRow(artifact)
            } else {
                Text(verbatim: option.label)
            }
        }
    }

    private var citationSelection: Binding<String> {
        Binding(
            get: { model.activeCitationID ?? "" },
            set: { newValue in
                if newValue.isEmpty {
                    if model.activeCitationID != nil { model.selectCitation(nil) }
                    return
                }
                guard newValue != model.activeCitationID else { return }
                model.selectCitation(newValue)
            }
        )
    }

    private var artifactSelection: Binding<String> {
        Binding(
            get: { model.selectedArtifactID ?? "" },
            set: { newValue in
                guard !newValue.isEmpty, newValue != model.selectedArtifactID else { return }
                model.requestSelectArtifact(newValue)
            }
        )
    }

    private var citationOptions: [PVSelectOption] {
        [PVSelectOption(value: "", label: L10n.string(L10n.CitationComposer.newCitation))]
            + model.orderedCitations.map {
                PVSelectOption(value: $0.citation.id, label: $0.citation.ref)
            }
    }

    private var artifactOptions: [PVSelectOption] {
        model.artifacts.map { artifact in
            PVSelectOption(value: artifact.id, label: artifactTitle(artifact))
        }
    }

    private var citationPosition: String {
        guard let id = model.activeCitationID,
              let index = model.orderedCitations.firstIndex(where: { $0.citation.id == id })
        else {
            return L10n.string(L10n.CitationComposer.newPosition)
        }
        return L10n.DesignSystem.selectOptionPosition(
            current: index + 1,
            count: model.orderedCitations.count
        )
    }

    private var artifactPosition: String {
        guard let id = model.selectedArtifactID,
              let index = model.artifacts.firstIndex(where: { $0.id == id }),
              !model.artifacts.isEmpty
        else { return "" }
        return L10n.DesignSystem.selectOptionPosition(current: index + 1, count: model.artifacts.count)
    }

    private func citationMenuRow(_ listed: CatalogListedCitation) -> some View {
        CitationStepMenuRow(listed: listed, locator: menuLocator(listed.citation))
    }

    private func artifactMenuRow(_ artifact: CatalogArtifact) -> some View {
        HStack(alignment: .center, spacing: 10) {
            artifactThumbnail(artifact)
            VStack(alignment: .leading, spacing: 1) {
                Text(verbatim: artifactTitle(artifact))
                    .font(PVFont.body(size: PVTypeScale.caption))
                    .foregroundStyle(PVColor.textPrimary)
                    .lineLimit(1)
                Text(verbatim: artifactMeta(artifact))
                    .font(PVFont.mono(size: PVTypeScale.micro))
                    .foregroundStyle(PVColor.textMuted)
                    .lineLimit(1)
                Text(verbatim: L10n.CitationComposer.artifactCitationCount(count: model.listedCount(for: artifact.id)))
                    .font(PVFont.mono(size: PVTypeScale.micro))
                    .foregroundStyle(PVColor.textMuted)
                    .lineLimit(1)
            }
        }
        .padding(.vertical, PVSpacing.space2)
    }

    private func artifactThumbnail(_ artifact: CatalogArtifact) -> some View {
        let content: PVThumbnail.Content
        if artifact.id == model.selectedArtifactID,
           let image = model.artifactViewer.displayImage,
           model.artifactViewer.kind == .image
        {
            content = PVThumbnail.Content(image: Image(nsImage: image))
        } else {
            content = PVThumbnail.Content(glyph: artifactMark(artifact).map(PVGlyph.mark))
        }
        return PVThumbnail(content, size: 48)
    }

    private var labelField: some View {
        PVField(label: L10n.CitationComposer.descriptionLabel) {
            PVInput(
                text: $model.citationDescription,
                size: .sm,
                prompt: L10n.CitationComposer.descriptionPrompt
            )
            .accessibilityIdentifier("citationComposer.label")
            .accessibilityLabel(Text(L10n.CitationComposer.descriptionLabel))
        }
    }

    private func transcriptionColumn(rows: ClosedRange<Int>) -> some View {
        VStack(alignment: .leading, spacing: PVSpacing.space4) {
            PVField(label: L10n.CitationComposer.transcriptionLabel, error: model.fields.error) {
                PVTextArea(
                    text: $model.transcription,
                    lineLimit: rows,
                    prompt: L10n.CitationComposer.transcriptionPrompt
                )
                .accessibilityIdentifier("citationComposer.transcription")
            }
            .overlay(alignment: .topTrailing) {
                readingAction
                    .offset(y: -5)
            }
            PVCheckbox(
                L10n.CitationComposer.uncertainLabel,
                isChecked: $model.transcriptionUncertain
            )
            if let message = model.transcriptionOCRMessage, !message.isEmpty {
                PVCallout(tone: .warning, message: message)
            }
        }
    }

    @ViewBuilder
    private var readingAction: some View {
        if model.artifactViewer.kind == .pdf {
            PVButton(
                L10n.CitationComposer.pasteTranscription,
                variant: .ghost,
                size: .sm,
                icon: .clipboardPaste
            ) {
                model.requestPasteTranscription()
            }
            .disabled(!model.canPasteTranscription || model.isTranscribing)
        } else if model.artifactViewer.kind == .image {
            PVButton(
                L10n.CitationComposer.autoTranscribe,
                variant: .primary,
                size: .sm,
                icon: .scanText,
                loading: model.isTranscribing
            ) {
                model.requestAutoTranscribe()
            }
            .disabled(!model.canAutoTranscribe || model.isTranscribing)
        }
    }

    private var footer: some View {
        HStack(spacing: PVSpacing.space4) {
            Spacer(minLength: 0)
            if model.activeCitationID != nil {
                PVButton(
                    L10n.CitationComposer.deleteReading,
                    variant: .ghost,
                    size: .sm,
                    icon: .trash
                ) {
                    Task { await model.askDeleteCitation() }
                }
                .disabled(model.isTranscribing)
                .accessibilityLabel(Text(verbatim: L10n.CitationComposer.deleteCitationAccessibility(ref: model.activeCitationRef)))
            }
            PVButton(
                L10n.CitationComposer.save,
                variant: .secondary,
                size: .sm,
                loading: model.fields.isSaving
            ) {
                Task { await model.fields.saveCitation() }
            }
            .disabled(!canSaveCitation)
            .accessibilityIdentifier("citationComposer.saveCitation")
        }
    }

    private var canSaveCitation: Bool {
        !model.fields.isSaving && !model.isTranscribing
            && (model.fields.isDirty || model.activeCitationID == nil)
    }

    private func menuLocator(_ citation: CatalogCitation) -> String {
        let locator = CitationLocatorDraft.decode(citation.locatorJSON)
        var parts: [String] = []
        if let page = locator.page {
            parts.append(L10n.CitationComposer.menuPage(page))
        }
        if locator.region != nil {
            parts.append(L10n.string(L10n.CitationComposer.menuRegion))
        }
        return parts.joined(separator: " · ")
    }

    private func artifactTitle(_ artifact: CatalogArtifact) -> String {
        let label = artifact.label.trimmingCharacters(in: .whitespacesAndNewlines)
        if !label.isEmpty { return label }
        return artifact.file?.originalFilename ?? ""
    }

    private func artifactMeta(_ artifact: CatalogArtifact) -> String {
        L10n.CitationComposer.artifactKindDetail(
            kind: artifactKindLabel(artifact),
            detail: artifactDetail(artifact)
        )
    }

    private func artifactKindLabel(_ artifact: CatalogArtifact) -> String {
        let media = artifact.file?.mediaType ?? ""
        if media == "application/pdf" {
            return L10n.string(L10n.CitationComposer.artifactKindPDF)
        }
        if media.hasPrefix("image/") {
            return L10n.string(L10n.CitationComposer.artifactKindImage)
        }
        return L10n.string(L10n.CitationComposer.artifactKindUnknown)
    }

    private func artifactDetail(_ artifact: CatalogArtifact) -> String {
        if artifact.id == model.selectedArtifactID {
            if model.artifactViewer.kind == .pdf, model.artifactViewer.pageCount > 0 {
                return L10n.CitationComposer.pageCount(count: model.artifactViewer.pageCount)
            }
            if model.artifactViewer.kind == .image,
               let image = model.artifactViewer.displayImage,
               let size = pixelSize(image)
            {
                return L10n.CitationComposer.imagePixels(width: size.width, height: size.height)
            }
        }
        return ByteCountFormatter.string(fromByteCount: artifact.file?.byteSize ?? 0, countStyle: .file)
    }

    private func pixelSize(_ image: NSImage) -> (width: Int, height: Int)? {
        if let rep = image.representations.first, rep.pixelsWide > 0, rep.pixelsHigh > 0 {
            return (rep.pixelsWide, rep.pixelsHigh)
        }
        let width = Int(image.size.width.rounded())
        let height = Int(image.size.height.rounded())
        guard width > 0, height > 0 else { return nil }
        return (width, height)
    }

    private func artifactMark(_ artifact: CatalogArtifact) -> PVMarkKey? {
        PVFileTypeGlyph.key(
            mediaType: artifact.file?.mediaType,
            originalFilename: artifact.file?.originalFilename
        )
    }

}

/// Citation menu row for the step picker and the observation dialog.
struct CitationStepMenuRow: View {
    let listed: CatalogListedCitation
    var locator: String

    var body: some View {
        let citation = listed.citation
        let label = citation.description.trimmingCharacters(in: .whitespacesAndNewlines)
        let transcription = citation.transcription
            .components(separatedBy: .newlines)
            .map { $0.trimmingCharacters(in: .whitespaces) }
            .filter { !$0.isEmpty }
            .joined(separator: " — ")
        return VStack(alignment: .leading, spacing: 2) {
            HStack(alignment: .firstTextBaseline, spacing: PVSpacing.space4) {
                Text(verbatim: citation.ref)
                    .font(PVFont.mono(size: 12, weight: PVFontWeight.semibold))
                    .foregroundStyle(PVColor.textPrimary)
                    .lineLimit(1)
                Text(verbatim: L10n.CitationComposer.citationStepperCount(count: listed.observationCount))
                    .font(PVFont.mono(size: 11))
                    .foregroundStyle(PVColor.textMuted)
                    .lineLimit(1)
                Spacer(minLength: PVSpacing.space4)
                if !locator.isEmpty {
                    Text(verbatim: locator)
                        .font(PVFont.mono(size: 11))
                        .foregroundStyle(PVColor.textFaint)
                        .lineLimit(1)
                }
            }
            if !label.isEmpty {
                Text(verbatim: label)
                    .font(PVFont.body(size: 13))
                    .foregroundStyle(PVColor.textPrimary)
                    .lineLimit(1)
            }
            if transcription.isEmpty {
                Text(L10n.CitationComposer.noTranscriptionLine)
                    .font(PVFont.body(size: 12, italic: true))
                    .foregroundStyle(PVColor.textMuted)
                    .lineLimit(2)
            } else {
                Text(verbatim: transcription)
                    .font(PVFont.body(size: 12))
                    .foregroundStyle(PVColor.textSecondary)
                    .lineLimit(2)
            }
        }
        .padding(.vertical, PVSpacing.space3)
        .frame(maxWidth: .infinity, alignment: .leading)
    }
}

/// Places children in weighted columns. Height follows the tallest child.
private struct WeightedColumns: Layout {
    var weights: [CGFloat]
    var spacing: CGFloat

    func sizeThatFits(proposal: ProposedViewSize, subviews: Subviews, cache: inout ()) -> CGSize {
        let widths = columnWidths(total: proposal.width ?? 0, count: subviews.count)
        let height = zip(subviews, widths).map { view, width in
            view.sizeThatFits(ProposedViewSize(width: width, height: nil)).height
        }.max() ?? 0
        return CGSize(width: proposal.width ?? 0, height: height)
    }

    func placeSubviews(in bounds: CGRect, proposal: ProposedViewSize, subviews: Subviews, cache: inout ()) {
        let widths = columnWidths(total: bounds.width, count: subviews.count)
        var x = bounds.minX
        for (view, width) in zip(subviews, widths) {
            view.place(
                at: CGPoint(x: x, y: bounds.minY),
                anchor: .topLeading,
                proposal: ProposedViewSize(width: width, height: nil)
            )
            x += width + spacing
        }
    }

    private func columnWidths(total: CGFloat, count: Int) -> [CGFloat] {
        let weightCount = min(count, weights.count)
        guard weightCount > 0 else { return [] }
        let used = Array(weights.prefix(weightCount))
        let sum = used.reduce(0, +)
        let available = max(0, total - spacing * CGFloat(max(count - 1, 0)))
        guard sum > 0 else {
            let each = count == 0 ? 0 : available / CGFloat(count)
            return Array(repeating: each, count: count)
        }
        return used.map { available * $0 / sum }
    }
}
