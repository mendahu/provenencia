import SwiftUI

/// Right sidebar: identity line, citation fields, connections, observations, Done.
struct CitationComposerFormPane: View {
    @Bindable var model: CitationComposerModel
    var wide: Bool = false

    @Environment(WorkspaceNavigation.self) private var navigation

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
        CitationComposerObservationScroll(rows: model.observations, row: observationRow) {
            VStack(alignment: .leading, spacing: PVSpacing.space5) {
                VStack(alignment: .leading, spacing: PVSpacing.space8) {
                    identityLine
                    citationFields
                    connectionsSection
                }
                .padding(.bottom, PVSpacing.space3)
                observationHeader
            }
        } footer: {
            addObservationButton
        }
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
            CitationComposerObservationScroll(rows: model.observations, row: observationRow) {
                VStack(alignment: .leading, spacing: PVSpacing.space5) {
                    connectionsSection
                        .padding(.bottom, PVSpacing.space3)
                    observationHeader
                }
            } footer: {
                addObservationButton
            }
            .frame(maxWidth: .infinity, maxHeight: .infinity, alignment: .top)
        }
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
            isDisabled: model.identityMenusDisabled,
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
                ? L10n.string(L10n.CitationComposer.newCitation)
                : model.activeCitationRef,
            menuWidth: 400,
            fillsWidth: false,
            maxVisibleRows: 8,
            rowHeight: 68,
            isDisabled: model.identityMenusDisabled,
            accessibilitySpokenLabel: model.activeCitationRef.isEmpty
                ? L10n.string(L10n.CitationComposer.citationMenuNew)
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
            label: L10n.string(L10n.CitationComposer.newCitation),
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
                hint: model.transcriptionActionHint
            ) {
                HStack(spacing: PVSpacing.space4) {
                    PVCheckbox(
                        L10n.CitationComposer.uncertainLabel,
                        isChecked: $model.transcriptionUncertain,
                        isDisabled: model.isTranscribing
                    )
                    .accessibilityIdentifier("citationComposer.uncertain")
                    Spacer(minLength: 0)
                    if model.isPDFArtifact {
                        pasteTranscriptionButton
                    } else {
                        autoTranscribeButton
                    }
                }
            }
            PVTextArea(text: $model.transcription, lineLimit: wide ? 6...8 : 2...6)
                .disabled(model.isTranscribing)
                .accessibilityIdentifier("citationComposer.transcription")
            if model.transcriptionUncertain {
                PVField(label: L10n.CitationComposer.uncertainNoteLabel) {
                    PVTextArea(text: $model.transcriptionNote, lineLimit: 1...3)
                        .accessibilityIdentifier("citationComposer.uncertainNote")
                }
            }
            PVField(label: L10n.CitationComposer.descriptionLabel) {
                PVTextArea(text: $model.citationDescription, lineLimit: 1...4)
                    .accessibilityIdentifier("citationComposer.description")
            }
            if let error = model.fields.error {
                PVCallout(tone: .danger, message: error, compact: true)
                    .accessibilityIdentifier("citationComposer.saveError")
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
                if model.activeCitationID != nil {
                    PVButton(
                        L10n.CitationComposer.deleteCitation,
                        variant: .ghost,
                        size: .sm
                    ) {
                        Task { await model.askDeleteCitation() }
                    }
                    .disabled(model.isDeletingResource)
                    .accessibilityLabel(
                        Text(verbatim: L10n.CitationComposer.deleteCitationAccessibility(
                            ref: model.activeCitationRef
                        ))
                    )
                    .accessibilityIdentifier("citationComposer.deleteCitation")
                }
                PVButton(
                    L10n.CitationComposer.save,
                    variant: .secondary,
                    size: .sm,
                    loading: model.fields.isSaving
                ) {
                    Task { await model.fields.saveCitation() }
                }
                .disabled(model.fields.isSaving || model.isTranscribing)
                .accessibilityIdentifier("citationComposer.saveCitation")
            }
        }
    }

    private var autoTranscribeButton: some View {
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
        .disabled(!model.canAutoTranscribe)
        .accessibilityIdentifier("citationComposer.autoTranscribe")
        .accessibilityLabel(
            model.locator.hasRegion
                ? Text(L10n.CitationComposer.autoTranscribeDrawnRegion)
                : Text(L10n.CitationComposer.autoTranscribe)
        )
        .accessibilityHint(Text(model.autoTranscribeHint))
    }

    private var pasteTranscriptionButton: some View {
        let enabled = model.canPasteTranscription
        return PVButton(
            L10n.CitationComposer.pasteTranscription,
            variant: .secondary,
            size: .sm,
            icon: .clipboardPaste
        ) {
            model.requestPasteTranscription()
        }
        .disabled(!enabled)
        .accessibilityIdentifier("citationComposer.pasteTranscription")
        .accessibilityLabel(
            enabled
                ? Text(L10n.CitationComposer.pasteTranscription)
                : Text(verbatim: L10n.CitationComposer.pasteUnavailable(
                    hint: model.transcriptionActionHint
                ))
        )
    }

    private var connectionsSection: some View {
        VStack(alignment: .leading, spacing: PVSpacing.space5) {
            ForEach(model.connections.rows) { row in
                CitationComposerConnectionRow(
                    row: row,
                    termOptions: row.termProperty.map { model.termOptions(for: $0.id) } ?? [],
                    onTerm: { model.connections.applyTerm(connectionID: row.id, termID: $0) },
                    onSave: { Task { await model.connections.saveConnection() } },
                    onDiscard: { model.connections.discard() },
                    onCommitRole: { Task { await model.connections.commitRole(connectionID: row.id) } },
                    onRevertRole: { model.connections.revertRole(connectionID: row.id) }
                )
            }
        }
    }

    private var observationHeader: some View {
        PVSectionHeader(
            title: L10n.CitationComposer.observationsSection,
            meta: "\(model.observations.count)"
        )
    }

    private func observationRow(_ row: ObservationRow) -> some View {
        CitationComposerObservationRow(
            row: row,
            property: model.catalogProperty(id: row.propertyID),
            propertyOptions: model.propertyOptions(for: row.subjectID),
            subjectOptions: model.subjectOptions,
            termOptions: model.termOptions(for: row.propertyID),
            summary: model.observationSummary(for: row),
            isFocused: model.focusedObservationID == row.id,
            onSubject: { model.updateObservationSubject(id: row.id, subjectID: $0) },
            onProperty: { model.updateObservationProperty(id: row.id, propertyID: $0) },
            onText: { model.updateObservationText(id: row.id, text: $0) },
            onInteger: { model.updateObservationInteger(id: row.id, text: $0) },
            onTerm: { model.updateObservationTerm(id: row.id, termID: $0) },
            onEditValue: { model.beginEditObservation(row) },
            onTogglePolarity: { model.toggleObservationPolarity(id: row.id) },
            onSave: { Task { await model.observationRows.commit(rowID: row.id) } },
            onRevert: { model.observationRows.revert(rowID: row.id) },
            onRequestDelete: { Task { await model.askDeleteObservation(rowID: row.id) } },
            onAddCustomTerm: { model.beginAddCustomTerm(rowID: row.id) }
        )
    }

    private var addObservationButton: some View {
        PVButton(L10n.CitationComposer.addObservation, variant: .ghost, size: .sm, icon: .plus) {
            model.beginAddObservation()
        }
        .accessibilityIdentifier("citationComposer.addObservation")
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

/// Which observation rows to build, plus the spacers that stand in for the rest.
enum CitationComposerObservationWindowing {
    struct Layout: Equatable {
        var start: Int
        var end: Int
        var topSpacer: CGFloat
        var bottomSpacer: CGFloat
    }

    static let estimatedRowHeight: CGFloat = 104
    static let overscan: CGFloat = 480
    /// A short list is one stack. Past this, only the rows near the viewport are built.
    static let fullCount = 16

    static func layout(
        rowHeights: [CGFloat],
        spacing: CGFloat,
        blockMinY: CGFloat,
        viewportHeight: CGFloat,
        overscan: CGFloat
    ) -> Layout {
        let count = rowHeights.count
        if count == 0 {
            return Layout(start: 0, end: 0, topSpacer: 0, bottomSpacer: 0)
        }
        if count <= fullCount {
            return Layout(start: 0, end: count, topSpacer: 0, bottomSpacer: 0)
        }
        if viewportHeight <= 0 {
            let end = min(count, fullCount)
            return Layout(
                start: 0,
                end: end,
                topSpacer: 0,
                bottomSpacer: stackHeight(Array(rowHeights[end...]), spacing: spacing)
            )
        }
        let visibleTop = -blockMinY - overscan
        let visibleBottom = -blockMinY + viewportHeight + overscan
        var start: Int?
        var end = count
        var y: CGFloat = 0
        for index in 0..<count {
            let rowTop = y
            let rowBottom = y + rowHeights[index]
            let intersects = rowBottom > visibleTop && rowTop < visibleBottom
            if intersects {
                if start == nil { start = index }
            } else if start != nil {
                end = index
                break
            }
            y = rowBottom + spacing
        }
        guard let start else {
            let total = stackHeight(rowHeights, spacing: spacing)
            if visibleBottom <= 0 {
                return Layout(start: 0, end: 0, topSpacer: 0, bottomSpacer: total)
            }
            return Layout(start: count, end: count, topSpacer: total, bottomSpacer: 0)
        }
        return Layout(
            start: start,
            end: end,
            topSpacer: stackHeight(Array(rowHeights[..<start]), spacing: spacing),
            bottomSpacer: stackHeight(Array(rowHeights[end...]), spacing: spacing)
        )
    }

    /// Height of rows with `spacing` between them and no trailing gap.
    static func stackHeight(_ heights: [CGFloat], spacing: CGFloat) -> CGFloat {
        guard !heights.isEmpty else { return 0 }
        return heights.reduce(0, +) + spacing * CGFloat(heights.count - 1)
    }
}

/// Scrolls the citation column and builds only the observation rows near the viewport.
///
/// A `LazyVStack` inserts those rows, and the text fields inside them, during the
/// scroll view's constraint pass. AppKit then aborts: the window asks for more
/// constraint passes than it has views. This stack is complete before that pass.
private struct CitationComposerObservationScroll<Above: View, Row: View, Footer: View>: View {
    private let rows: [ObservationRow]
    private let above: Above
    private let row: (ObservationRow) -> Row
    private let footer: Footer
    private let spacing = PVSpacing.space5

    @State private var metrics = Metrics()
    @State private var layout = CitationComposerObservationWindowing.Layout(
        start: 0, end: 0, topSpacer: 0, bottomSpacer: 0
    )
    @State private var hasMetrics = false

    init(
        rows: [ObservationRow],
        row: @escaping (ObservationRow) -> Row,
        @ViewBuilder above: () -> Above,
        @ViewBuilder footer: () -> Footer
    ) {
        self.rows = rows
        self.row = row
        self.above = above()
        self.footer = footer()
    }

    var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: spacing) {
                above
                observationRows
                footer
            }
            .padding(PVSpacing.space7)
        }
        .coordinateSpace(name: citationComposerObservationSpace)
        .background { viewportReader }
        .onPreferenceChange(BlockMinYKey.self) { minY in
            guard minY != metrics.blockMinY else { return }
            metrics.blockMinY = minY
            hasMetrics = true
            publishLayout()
        }
        .onPreferenceChange(ViewportHeightKey.self) { height in
            guard height != metrics.viewportHeight else { return }
            metrics.viewportHeight = height
            hasMetrics = true
            publishLayout()
        }
        .onPreferenceChange(RowHeightKey.self) { reported in
            var changed = false
            for (id, height) in reported where abs((metrics.measured[id] ?? -1) - height) > 0.5 {
                metrics.measured[id] = height
                changed = true
            }
            if changed { publishLayout() }
        }
        .onChange(of: rows.map(\.id)) { _, _ in
            publishLayout()
        }
    }

    private var shown: CitationComposerObservationWindowing.Layout {
        if hasMetrics { return layout }
        return CitationComposerObservationWindowing.layout(
            rowHeights: resolvedHeights,
            spacing: spacing,
            blockMinY: 0,
            viewportHeight: 0,
            overscan: CitationComposerObservationWindowing.overscan
        )
    }

    @ViewBuilder
    private var observationRows: some View {
        let window = clamped(shown)
        VStack(alignment: .leading, spacing: spacing) {
            if window.topSpacer > 0 {
                Color.clear
                    .frame(height: window.topSpacer)
                    .accessibilityHidden(true)
            }
            ForEach(rows[window.start..<window.end]) { item in
                row(item)
                    .background { heightReader(item.id) }
            }
            if window.bottomSpacer > 0 {
                Color.clear
                    .frame(height: window.bottomSpacer)
                    .accessibilityHidden(true)
            }
        }
        .background { blockOriginReader }
    }

    private func clamped(_ layout: CitationComposerObservationWindowing.Layout) -> CitationComposerObservationWindowing.Layout {
        let start = min(max(layout.start, 0), rows.count)
        let end = min(max(layout.end, start), rows.count)
        return CitationComposerObservationWindowing.Layout(
            start: start,
            end: end,
            topSpacer: layout.topSpacer,
            bottomSpacer: layout.bottomSpacer
        )
    }

    private var resolvedHeights: [CGFloat] {
        rows.map { metrics.measured[$0.id] ?? CitationComposerObservationWindowing.estimatedRowHeight }
    }

    private func publishLayout() {
        let next = CitationComposerObservationWindowing.layout(
            rowHeights: resolvedHeights,
            spacing: spacing,
            blockMinY: metrics.blockMinY,
            viewportHeight: metrics.viewportHeight,
            overscan: CitationComposerObservationWindowing.overscan
        )
        if next != layout {
            layout = next
        }
    }

    private var blockOriginReader: some View {
        GeometryReader { proxy in
            Color.clear.preference(
                key: BlockMinYKey.self,
                value: proxy.frame(in: .named(citationComposerObservationSpace)).minY
            )
        }
    }

    private var viewportReader: some View {
        GeometryReader { proxy in
            Color.clear.preference(key: ViewportHeightKey.self, value: proxy.size.height)
        }
    }

    private func heightReader(_ id: UUID) -> some View {
        GeometryReader { proxy in
            Color.clear.preference(key: RowHeightKey.self, value: [id: proxy.size.height])
        }
    }

    private final class Metrics {
        var blockMinY: CGFloat = 0
        var viewportHeight: CGFloat = 0
        var measured: [UUID: CGFloat] = [:]
    }
}

private let citationComposerObservationSpace = "citationComposerObservations"

private struct BlockMinYKey: PreferenceKey {
    static var defaultValue: CGFloat = 0
    static func reduce(value: inout CGFloat, nextValue: () -> CGFloat) {
        value = nextValue()
    }
}

private struct ViewportHeightKey: PreferenceKey {
    static var defaultValue: CGFloat = 0
    static func reduce(value: inout CGFloat, nextValue: () -> CGFloat) {
        value = nextValue()
    }
}

private struct RowHeightKey: PreferenceKey {
    static var defaultValue: [UUID: CGFloat] = [:]
    static func reduce(value: inout [UUID: CGFloat], nextValue: () -> [UUID: CGFloat]) {
        value.merge(nextValue(), uniquingKeysWith: { _, new in new })
    }
}
