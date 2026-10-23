import SwiftUI

/// The Promote evidence sheet (S11-D2): why the row was suggested, in the
/// title, then the evidence that this row is its record, as sentences the
/// researcher checks or unchecks. Checked evidence is pinned on Done. Escape
/// and an outside click close it through `.pvSheet`; nothing is lost by closing,
/// because pins and the claim fields already live in the model.
struct PromoteEvidenceSheet: View {
    let row: PromoteFlow.Row
    let grades: [CatalogClaimConfidenceGrade]
    let model: PromoteModel

    /// Lines opened to their records, by a line's first pin id. Closing the sheet drops it.
    @State private var expanded: Set<String> = []

    var body: some View {
        PVPanel(
            title: titleText,
            subtitle: subtitleText,
            width: 880
        ) {
            ScrollView {
                LazyVStack(alignment: .leading, spacing: PVSpacing.space5) {
                    if let readout = row.selectionReadout {
                        EvidenceTitleLines(row: row, readout: readout, model: model)
                    }
                    if !hasRecord {
                        PVEmptyState(
                            icon: .searchEmpty,
                            title: L10n.Promote.pickRecordTitle,
                            message: L10n.string(L10n.Promote.pickRecordBody),
                            compact: true
                        )
                    }
                    if showsClaimFields {
                        claimFields
                    }
                    if let title = row.readingTitle {
                        evidenceArea(title: title)
                    }
                }
                .padding(PVSpacing.space5)
            }
        } footer: {
            PVButton(L10n.Promote.backToRows, variant: .primary) { model.closeSheet() }
                .accessibilityIdentifier("promote.sheet.back")
        } titleAccessory: {
            HStack(alignment: .top, spacing: PVSpacing.space4) {
                marks
                Spacer(minLength: PVSpacing.space4)
                PVIconButton(.dismiss, label: L10n.Promote.closeSheet, size: .sm) { model.closeSheet() }
                    .accessibilityIdentifier("promote.sheet.close")
            }
        }
    }

    // MARK: Title

    /// "<row name> joins <record title> <REF>" for a record; the row's name and its New option otherwise.
    private var titleText: Text {
        switch row.target {
        case .handle(_, let ref, _):
            let title = row.readingTitle ?? ""
            return Text(EvidenceTextStyle.painted(
                PromoteEvidenceDisplay.title(name: row.name, recordTitle: title, ref: ref)
            ))
        case .newKind:
            return Text(verbatim: L10n.Promote.sheetTitleNew(name: row.name, target: L10n.string(L10n.Promote.newOption(row.kind))))
        case .skip, .unset:
            return Text(verbatim: row.name)
        }
    }

    /// The claim in one sentence, under the title of a row filed on a record.
    private var subtitleText: Text? {
        guard let title = row.readingTitle else { return nil }
        return Text(EvidenceTextStyle.painted(
            PromoteEvidenceDisplay.subtitle(
                name: row.name, sourceTitle: model.entry.sourceTitle, kind: row.kind, recordTitle: title
            )
        ))
    }

    /// The band and Contested marks after the title. The reason is their tooltip;
    /// the pair is one element for VoiceOver.
    @ViewBuilder
    private var marks: some View {
        if let readout = row.selectionReadout {
            let reason = model.reasonText(readout.reason, on: row)
            PromoteReadoutBadges(readout: readout)
                .help(Text(verbatim: reason))
                .focusable()
                .accessibilityElement(children: .ignore)
                .accessibilityLabel(Text(verbatim: L10n.Promote.whySuggestedSpoken(
                    assessment: model.assessmentText(readout),
                    reason: reason
                )))
                .accessibilityIdentifier("promote.sheet.marks")
        }
    }

    private var hasRecord: Bool {
        if case .handle = row.target { return true }
        return false
    }

    // MARK: Claim fields

    private var showsClaimFields: Bool {
        switch row.target {
        case .handle, .newKind: true
        case .skip, .unset: false
        }
    }

    private var claimFields: some View {
        // The argument box is as tall as the Status and Confidence column beside it.
        HStack(alignment: .top, spacing: PVSpacing.space6) {
            VStack(alignment: .leading, spacing: PVSpacing.space4) {
                PVField(label: L10n.Promote.status) {
                    PVSelect(
                        selection: .constant("accepted"),
                        options: [PVSelectOption(value: "accepted", label: L10n.string(L10n.Promote.statusAccepted))]
                    )
                }
                PVField(label: L10n.Promote.confidence) {
                    PVSelect(
                        selection: confidenceBinding,
                        options: confidenceOptions,
                        accessibilityIdentifier: "promote.sheet.confidence"
                    )
                }
            }
            .frame(width: 200)
            PVField(label: L10n.Promote.argument, marker: .optional) {
                PVTextArea(text: argumentBinding, fillsHeight: true)
                    .accessibilityIdentifier("promote.sheet.argument")
            }
            .frame(maxWidth: .infinity, maxHeight: .infinity, alignment: .topLeading)
        }
        .fixedSize(horizontal: false, vertical: true)
    }

    // MARK: Evidence

    /// The heading, the evidence's state, then one Card per property and per relationship. These are
    /// siblings of the sheet's lazy stack, so a Card is built when it scrolls into view.
    @ViewBuilder
    private func evidenceArea(title: String) -> some View {
        let board = row.board
        PVDivider()
        PVSectionHeader(
            title: EvidenceTextStyle.painted(
                PromoteEvidenceDisplay.evidenceHeading(
                    name: row.name, sourceTitle: model.entry.sourceTitle, kind: row.kind, recordTitle: title
                )
            ),
            meta: board.map { L10n.Promote.checkedCount($0.checkedCount(pins: row.pins)) }
        )
        .accessibilityAddTraits(.isHeader)
        .accessibilityIdentifier("promote.sheet.heading")
        if let board, board.hasNothingToCompare {
            PVEmptyState(
                icon: .searchEmpty,
                title: L10n.Promote.nothingToCompareTitle,
                message: L10n.string(L10n.Promote.nothingToCompareBody),
                compact: true
            )
        }
        if case .failed(let message) = model.sheetEvidence {
            PVCallout(tone: .danger, title: L10n.Promote.evidenceUnavailable, message: message, compact: true) {
                PVButton(L10n.Promote.evidenceRetry, variant: .secondary, size: .sm) { model.reloadEvidence() }
                    .accessibilityIdentifier("promote.sheet.retry")
            }
        }
        if let board {
            cards(board)
        } else if model.sheetEvidence == .loading {
            VStack(alignment: .leading, spacing: PVSpacing.space3) {
                PVSkeletonBar(width: nil, height: 12)
                PVSkeletonBar(width: 320, height: 12)
                PVSkeletonBar(width: 240, height: 12)
            }
            // One element: VoiceOver says "Loading" once, not once per bar.
            .accessibilityElement(children: .ignore)
            .accessibilityLabel(Text(L10n.DesignSystem.listLoading))
            .accessibilityIdentifier("promote.sheet.loading")
        }
    }

    /// The Subject's own Cards in Go's order, then one Card per relationship.
    @ViewBuilder
    private func cards(_ board: EvidenceBoard) -> some View {
        let env = EvidenceEnv(
            projectDir: model.session.projectKey.projectDir,
            toggle: { model.togglePin(subjectID: row.subjectID, pinIDs: $0) },
            setOpen: { key, open in
                if open { expanded.insert(key) } else { expanded.remove(key) }
            }
        )
        ForEach(board.cards) { card in
            EvidenceSectionCard(group: card, pins: row.pins, expanded: expanded, env: env)
        }
        ForEach(board.relationships) { relationship in
            EvidenceRelationshipCard(relationship: relationship, pins: row.pins, expanded: expanded, env: env)
        }
    }

    // MARK: Bindings

    private var argumentBinding: Binding<String> {
        Binding(get: { row.argument }, set: { model.setArgument(subjectID: row.subjectID, argument: $0) })
    }

    private var confidenceBinding: Binding<String> {
        Binding(
            get: { row.confidenceGradeID ?? "" },
            set: { model.setConfidence(subjectID: row.subjectID, gradeID: $0.isEmpty ? nil : $0) }
        )
    }

    private var confidenceOptions: [PVSelectOption] {
        [PVSelectOption(value: "", label: L10n.string(L10n.Promote.confidenceUnset))]
            + grades.map { PVSelectOption(value: $0.id, label: $0.label) }
    }
}

// MARK: - Lines under the title

/// What the row shows beside its band, kept off the badges: the Updated caption and the contested line.
private struct EvidenceTitleLines: View {
    let row: PromoteFlow.Row
    let readout: PromoteFlow.SelectionReadout
    let model: PromoteModel

    var body: some View {
        if row.updated || readout.contested {
            VStack(alignment: .leading, spacing: PVSpacing.space2) {
                if row.updated {
                    PromoteUpdatedCaption(reason: model.reasonText(row.captionReason, on: row))
                }
                if readout.contested {
                    Text(L10n.Promote.somethingDisagrees)
                        .font(PVFont.body(size: PVTypeScale.bodySmall))
                        .foregroundStyle(PVColor.textPrimary)
                }
            }
            .frame(maxWidth: .infinity, alignment: .leading)
            .accessibilityElement(children: .contain)
            .accessibilityLabel(Text(L10n.Promote.whySuggested))
        }
    }
}
