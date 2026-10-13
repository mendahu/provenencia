import SwiftUI

/// Evidence column: citation and artifact menus, the reading, then the file.
struct CitationComposerEvidenceColumn: View {
    @Bindable var model: CitationComposerModel

    var body: some View {
        VStack(alignment: .leading, spacing: 0) {
            VStack(alignment: .leading, spacing: PVSpacing.space5) {
                picker
                reading
            }
            .padding(.horizontal, PVSpacing.space6)
            .padding(.top, PVSpacing.space5)
            .padding(.bottom, 14)
            .frame(maxWidth: .infinity, alignment: .leading)
            .background(PVColor.surfaceCard)
            .overlay(alignment: .bottom) {
                PVDivider(axis: .horizontal, color: PVColor.borderSubtle)
            }
            CitationComposerViewerPane(model: model)
        }
        .background(PVColor.surfaceSunken)
    }

    /// Panel picker: citation menu, artifact menu, then the locator summary.
    private var picker: some View {
        HStack(alignment: .center, spacing: PVSpacing.space4) {
            citationSelect
            Text(L10n.CitationComposer.pickerIn)
                .font(PVFont.body(size: 12))
                .foregroundStyle(PVColor.textMuted)
            artifactSelect
            Text(L10n.CitationComposer.pickerAt)
                .font(PVFont.body(size: 12))
                .foregroundStyle(PVColor.textMuted)
            locatorSummary
            Spacer(minLength: 0)
        }
    }

    private var citationSelect: some View {
        PVSelect(
            selection: citationSelection,
            options: citationOptions,
            size: .sm,
            icon: .quote,
            displayLabel: citationClosedLabel,
            menuWidth: 460,
            fillsWidth: false,
            rowHeight: 76,
            isDisabled: model.isTranscribing,
            accessibilityLabel: L10n.CitationComposer.citationEyebrow,
            accessibilityIdentifier: "citationComposer.citationSelect"
        ) { option in
            if option.id.isEmpty {
                Text(L10n.CitationComposer.newCitation)
                    .font(PVFont.body(size: PVTypeScale.caption))
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
            size: .sm,
            icon: .file,
            menuWidth: 360,
            fillsWidth: false,
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

    private var citationClosedLabel: String {
        guard let listed = currentListed else {
            return L10n.string(L10n.CitationComposer.newCitation)
        }
        let detail = CitationComposerAuthoring.subjectNameSummary(listed.subjectNames)
            ?? collapsedDetail(listed.citation.description)
            ?? L10n.string(L10n.CitationComposer.closedNoTranscription)
        return L10n.CitationComposer.citationClosed(listed.citation.ref, detail)
    }

    private func citationMenuRow(_ listed: CatalogListedCitation) -> some View {
        let citation = listed.citation
        let names = CitationComposerAuthoring.subjectNameSummary(listed.subjectNames)
        let preview = CitationComposerAuthoring.preview(
            transcription: citation.transcription,
            description: citation.description
        )
        let locator = menuLocator(citation)
        return HStack(alignment: .center, spacing: 10) {
            PVThumbnail(PVThumbnail.Content(icon: .scrollText), size: 52)
            VStack(alignment: .leading, spacing: 1) {
                HStack(alignment: .firstTextBaseline, spacing: PVSpacing.space4) {
                    Text(verbatim: citation.ref)
                        .font(PVFont.mono(size: 12, weight: PVFontWeight.semibold))
                        .foregroundStyle(PVColor.textPrimary)
                        .lineLimit(1)
                    Text(verbatim: L10n.CitationComposer.citationStepperCount(count: listed.observationCount))
                        .font(PVFont.mono(size: PVTypeScale.micro))
                        .foregroundStyle(PVColor.textMuted)
                        .lineLimit(1)
                    Spacer(minLength: PVSpacing.space4)
                    if !locator.isEmpty {
                        Text(verbatim: locator)
                            .font(PVFont.mono(size: PVTypeScale.micro))
                            .foregroundStyle(PVColor.textFaint)
                            .lineLimit(1)
                    }
                }
                if let names {
                    Text(verbatim: names)
                        .font(PVFont.body(size: 12))
                        .foregroundStyle(PVColor.textPrimary)
                        .lineLimit(1)
                } else {
                    Text(L10n.CitationComposer.noSubjectsYet)
                        .font(PVFont.body(size: 12, italic: true))
                        .foregroundStyle(PVColor.textMuted)
                        .lineLimit(1)
                }
                switch preview {
                case .text(let text):
                    Text(verbatim: text)
                        .font(PVFont.body(size: 12))
                        .foregroundStyle(PVColor.textSecondary)
                        .lineLimit(1)
                case .empty:
                    Text(L10n.CitationComposer.noTranscriptionYet)
                        .font(PVFont.body(size: 12, italic: true))
                        .foregroundStyle(PVColor.textMuted)
                        .lineLimit(1)
                }
            }
        }
        .padding(.vertical, PVSpacing.space2)
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
            content = PVThumbnail.Content(mark: artifactMark(artifact))
        }
        return PVThumbnail(content, size: 40)
    }

    /// One sunken chip: page, then region. Drawing on the file changes it.
    private var locatorSummary: some View {
        let page = model.locatorCapabilities.supportsPageLocator ? model.locator.page : nil
        let region = model.locatorCapabilities.supportsRegionLocator ? model.locator.region : nil
        return HStack(spacing: PVSpacing.space3) {
            if let page {
                PVIcon(.file, size: 13)
                Text(verbatim: L10n.CitationComposer.locatorPage(page))
                    .lineLimit(1)
            } else if region == nil {
                PVIcon(.scrollText, size: 13)
                Text(L10n.CitationComposer.locatorEntireArtifact)
                    .lineLimit(1)
            }
            if let region {
                if page != nil {
                    PVIcon(.chevronForward, size: 12)
                        .foregroundStyle(PVColor.textFaint)
                }
                PVIcon(.crop, size: 13)
                Text(verbatim: regionNoun(region.kind))
                    .lineLimit(1)
            }
        }
        .font(PVFont.body(size: 12))
        .foregroundStyle(PVColor.textSecondary)
        .padding(.horizontal, 10)
        .frame(height: PVSpacing.controlHeightSmall)
        .background(PVColor.surfaceSunken)
        .clipShape(RoundedRectangle(cornerRadius: PVRadius.sm, style: .continuous))
        .overlay(
            RoundedRectangle(cornerRadius: PVRadius.sm, style: .continuous)
                .stroke(PVColor.borderSubtle, lineWidth: 1)
        )
        .help(L10n.string(L10n.CitationComposer.locatorChipHint))
        .accessibilityElement(children: .combine)
        .accessibilityIdentifier("citationComposer.locatorSummary")
    }

    private var reading: some View {
        VStack(alignment: .leading, spacing: PVSpacing.space5) {
            ReadingColumns(spacing: PVSpacing.space6) {
                VStack(alignment: .leading, spacing: PVSpacing.space5) {
                    PVField(label: L10n.CitationComposer.transcriptionLabel, error: model.fields.error) {
                        PVTextArea(
                            text: $model.transcription,
                            lineLimit: 4...4,
                            prompt: L10n.CitationComposer.transcriptionPrompt
                        )
                        .accessibilityIdentifier("citationComposer.transcription")
                    }
                    HStack(spacing: PVSpacing.space6) {
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
                            .disabled(model.isTranscribing)
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
                    }
                }
                VStack(alignment: .leading, spacing: PVSpacing.space5) {
                    PVField(label: L10n.CitationComposer.descriptionLabel) {
                        PVTextArea(
                            text: $model.citationDescription,
                            lineLimit: 4...4,
                            prompt: L10n.CitationComposer.descriptionPrompt
                        )
                        .accessibilityIdentifier("citationComposer.description")
                    }
                    HStack(spacing: PVSpacing.space4) {
                        Spacer(minLength: 0)
                        if model.activeCitationID != nil {
                            PVIconButton(
                                .trash,
                                label: L10n.CitationComposer.deleteCitationAccessibility(ref: model.activeCitationRef),
                                size: .sm
                            ) {
                                Task { await model.askDeleteCitation() }
                            }
                            .disabled(model.isTranscribing)
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

    private var canSaveCitation: Bool {
        !model.fields.isSaving && !model.isTranscribing
            && (model.fields.isDirty || model.activeCitationID == nil)
    }

    private var currentListed: CatalogListedCitation? {
        model.orderedCitations.first { $0.citation.id == model.activeCitationID }
    }

    private func collapsedDetail(_ text: String) -> String? {
        let collapsed = text
            .components(separatedBy: .newlines)
            .map { $0.trimmingCharacters(in: .whitespaces) }
            .filter { !$0.isEmpty }
            .joined(separator: " — ")
        return collapsed.isEmpty ? nil : collapsed
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

    private func regionNoun(_ kind: ArtifactRegionKind) -> String {
        switch kind {
        case .rectangle:
            L10n.string(L10n.CitationComposer.locatorRectangle)
        case .lOpenTopRight, .lOpenTopLeft, .lOpenBottomRight, .lOpenBottomLeft:
            L10n.string(L10n.CitationComposer.locatorLShape)
        case .circle:
            L10n.string(L10n.CitationComposer.locatorCircle)
        case .freeform:
            L10n.string(L10n.CitationComposer.locatorPolygon)
        }
    }
}

/// Transcription takes 1.6 of the reading row; description takes 1.
private struct ReadingColumns: Layout {
    var spacing: CGFloat = 16

    func sizeThatFits(proposal: ProposedViewSize, subviews: Subviews, cache: inout ()) -> CGSize {
        let widths = columnWidths(total: proposal.width ?? 0, count: subviews.count)
        let height = zip(subviews, widths).map { view, width in
            view.sizeThatFits(ProposedViewSize(width: width, height: proposal.height)).height
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
        guard count == 2 else {
            let each = count == 0 ? 0 : max(0, (total - spacing * CGFloat(max(count - 1, 0))) / CGFloat(max(count, 1)))
            return Array(repeating: each, count: count)
        }
        let available = max(0, total - spacing)
        let leading = available * 1.6 / 2.6
        return [leading, available - leading]
    }
}
