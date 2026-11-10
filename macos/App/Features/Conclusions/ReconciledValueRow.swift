import SwiftUI

/// One field on a Conclusion detail page (board S9-D5): label · lead value ·
/// state badge, Source count and disagreement · disclosures for a mixed
/// field's other values and for the Why. State is always text, never colour
/// alone. Snowflake: Person, Event and Place pages share it from this folder.
struct ReconciledValueRow: View {
    /// How the row offers where its value came from. `.why` is the Why disclosure
    /// (the Event and Place pages). `.sheet` is the Person page's mode: the row
    /// ends with a 28pt slot holding the provenance control (empty when the row
    /// has no records) and draws no Why.
    enum Provenance {
        case why
        case sheet
    }

    let model: ReconciledValueRowModel
    var provenance: Provenance = .why
    /// Opens the provenance sheet on this row (`.sheet` mode only).
    var openProvenance: () -> Void = {}
    @Environment(WorkspaceNavigation.self) private var navigation
    @State private var showsOtherValues = false
    @State private var showsWhy = false

    /// The value column starts after the label column and its gap.
    static let labelWidth: CGFloat = 132
    static let disclosureInset: CGFloat = labelWidth + PVSpacing.space6

    var body: some View {
        VStack(alignment: .leading, spacing: 0) {
            PVDivider()
            HStack(alignment: .center, spacing: PVSpacing.space6) {
                summary
                disclosures
                if provenance == .sheet {
                    provenanceSlot
                }
            }
            .frame(minHeight: 52)
            .padding(.vertical, PVSpacing.space3)
            if showsOtherValues {
                otherValues
                    .padding(.leading, Self.disclosureInset)
                    .padding(.bottom, PVSpacing.space5)
            }
            if provenance == .why, showsWhy {
                ReconciliationReasoningView(model: model)
                .padding(.leading, Self.disclosureInset)
                .padding(.bottom, PVSpacing.space5)
            }
        }
        .accessibilityIdentifier("conclusion.detail.row.\(model.id)")
    }

    /// Label, value and state: one VoiceOver element with the row's spoken
    /// label; the disclosures stay separate controls.
    private var summary: some View {
        HStack(alignment: model.shown.listedValues.isEmpty ? .center : .firstTextBaseline, spacing: PVSpacing.space6) {
            Text(model.label)
                .pvMicroCaps()
                .foregroundStyle(PVColor.textMuted)
                .frame(width: Self.labelWidth, alignment: .leading)
            value
                .frame(maxWidth: .infinity, alignment: .leading)
            state
        }
        .accessibilityElement(children: .ignore)
        .accessibilityLabel(model.accessibilityLabel)
    }

    private var value: some View {
        PVProvenanceValueList(model.shown)
    }

    private var state: some View {
        HStack(spacing: PVSpacing.space4) {
            if let badge = model.shown.badge {
                PVProvenanceStateBadge(badge)
            }
            if let count = model.count {
                Text(count)
                    .font(PVFont.mono(size: PVTypeScale.caption))
                    .foregroundStyle(PVColor.textMuted)
            }
            if let against = model.against {
                HStack(spacing: PVSpacing.space2) {
                    PVIcon(.circleMinus, size: 12)
                    Text(against)
                        .font(PVFont.body(size: PVTypeScale.caption))
                }
                .foregroundStyle(PVColor.textSecondary)
            }
        }
    }

    private var disclosures: some View {
        HStack(spacing: PVSpacing.space2) {
            if let label = model.otherValuesLabel {
                PVDisclosureButton(label, isExpanded: $showsOtherValues)
                    .accessibilityIdentifier("conclusion.detail.row.\(model.id).others")
            }
            if provenance == .why, model.recordCount > 0 {
                PVDisclosureButton(L10n.Conclusions.whyButton, isExpanded: $showsWhy)
                    .accessibilityIdentifier("conclusion.detail.row.\(model.id).why")
            }
            if let opens = model.opens {
                PVIconButton(.arrowUpRight, label: opens.label, size: .sm) {
                    navigation.go(to: opens.location)
                }
                .accessibilityIdentifier("conclusion.detail.row.\(model.id).open")
            }
        }
        .frame(minWidth: 72, alignment: .trailing)
    }

    /// The 28pt slot at the row's end: the provenance control when the row has
    /// records, else empty so the values line up (VoiceOver skips the empty slot).
    @ViewBuilder
    private var provenanceSlot: some View {
        if model.recordCount == 0 {
            Color.clear
                .frame(width: 28, height: 28)
                .accessibilityHidden(true)
        } else {
            PVIconButton(.info, label: L10n.Conclusions.provenanceOpen(model.label), size: .sm, action: openProvenance)
                .frame(width: 28, height: 28)
                .accessibilityIdentifier("conclusion.detail.row.\(model.id).provenance")
        }
    }

    private var otherValues: some View {
        PVCard(tone: .sunken, cornerRadius: PVRadius.sm) {
            VStack(alignment: .leading, spacing: 0) {
                ForEach(model.otherValues) { other in
                    HStack(alignment: .firstTextBaseline, spacing: PVSpacing.space5) {
                        Text(verbatim: "\(other.rank)")
                            .font(PVFont.mono(size: PVTypeScale.caption))
                            .foregroundStyle(PVColor.textFaint)
                            .frame(width: 24, alignment: .leading)
                        Text(other.text)
                            .font(model.shown.face.valueFont)
                            .foregroundStyle(PVColor.textPrimary)
                            .frame(maxWidth: .infinity, alignment: .leading)
                        Text(other.support)
                            .font(PVFont.mono(size: PVTypeScale.caption))
                            .foregroundStyle(PVColor.textSecondary)
                    }
                    .padding(.horizontal, PVSpacing.space5)
                    .padding(.vertical, PVSpacing.space4)
                    .accessibilityElement(children: .combine)
                }
            }
            .padding(.vertical, PVSpacing.space2)
        }
        .accessibilityElement(children: .contain)
        .accessibilityLabel(L10n.Conclusions.otherValuesList(model.label))
    }
}

/// A page's field rows, under a heading with the refreshing spinner when `title` is given.
/// A page that draws neither (the Person page's Vitals, which shows refreshing in its
/// section switcher) passes nil.
struct ConclusionDetailRowsSection: View {
    let rows: [ReconciledValueRowModel]
    let refreshing: Bool
    var title: LocalizedStringResource? = nil
    var provenance: ReconciledValueRow.Provenance = .why

    @Environment(WorkspaceNavigation.self) private var navigation
    @Environment(\.pvWidthClass) private var widthClass
    /// The row whose provenance sheet is open (`.sheet` mode); one presenter for all rows.
    @State private var provenanceItem: ProvenanceSheetItem?

    var body: some View {
        VStack(alignment: .leading, spacing: 0) {
            if let title {
                PVSectionHeader(title: title, aside: {
                    if refreshing {
                        ProgressView().controlSize(.small)
                    }
                })
            }
            ForEach(rows) { row in
                ReconciledValueRow(
                    model: row, provenance: provenance,
                    openProvenance: { provenanceItem = ProvenanceSheetItem(id: row.id) }
                )
            }
        }
        .pvSheet(item: $provenanceItem) { item in
            if let row = item.row(in: rows) {
                PVProvenanceSheet(
                    content: row.provenanceContent(),
                    columns: PVProvenanceColumns(widthClass: widthClass),
                    idPrefix: "conclusion.detail.provenance",
                    open: { navigation.go(to: .provenance($0, $1)) }
                )
                .accessibilityIdentifier("conclusion.detail.provenance.\(row.id)")
            }
        }
        // A row that leaves the page closes its sheet.
        .onChange(of: rows) {
            if let item = provenanceItem, item.row(in: rows) == nil {
                provenanceItem = nil
            }
        }
    }
}
