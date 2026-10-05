import SwiftUI

/// The Promote place (S9-11, board S9-D9): a workspace place, not a sheet, so
/// the walk (S9-30) and compare (S9-19) get the page's width and the toolbar's
/// Back returns to the graph. The shell carries the subject being promoted,
/// step progress, Done, and the leave guard on every step.
struct PromoteView: View {
    @Environment(WorkspaceNavigation.self) private var navigation
    @State private var model: PromoteModel

    init(
        entry: PromoteEntry,
        session: WorkspaceSession,
        store: any GenealogyStore,
        userID: String,
        catalogCounts: CatalogCounts?
    ) {
        _model = State(initialValue: PromoteModel(
            entry: entry,
            session: session,
            store: store,
            userID: userID,
            catalogCounts: catalogCounts
        ))
    }

    var body: some View {
        let model = model
        VStack(spacing: 0) {
            PromoteHeader(
                subject: model.subject,
                sourceTitle: model.entry.sourceTitle,
                steps: model.steps,
                currentStep: model.currentStepIndex
            )
            ScrollView {
                stepScreen
                    .frame(maxWidth: 700, alignment: .leading)
                    .padding(.horizontal, PVSpacing.space10)
                    .padding(.top, 28)
                    .padding(.bottom, PVSpacing.space8)
                    .frame(maxWidth: .infinity, alignment: .leading)
            }
            footer
        }
        .background(PVColor.surfacePage)
        .background {
            if let graph: QueryHandle<SourceGraphRows> = model.session.queryHandle(model.graphKey) {
                PromoteSubjectWatch(handle: graph, model: model)
            }
        }
        .onAppear {
            model.navigation = navigation
            navigation.leaveGuard = model
        }
        .onDisappear {
            if navigation.leaveGuard === model {
                navigation.leaveGuard = nil
            }
        }
        .task {
            let _: QueryHandle<SourceGraphRows> = model.session.query(model.graphKey)
            let _: QueryHandle<[CatalogPromoteTargetSuggestion]> = model.session.query(model.suggestionsKey)
            let _: QueryHandle<[CatalogClaimConfidenceGrade]> = model.session.query(model.confidenceKey)
            let _: QueryHandle<PropertiesSnapshot> = model.session.query(model.propertiesKey)
        }
        .pvConfirm(
            item: leaveBinding,
            copy: { _ in
                PVConfirmCopy(
                    title: model.leaveTitle,
                    message: model.leaveMessage,
                    confirm: L10n.Promote.leaveConfirm,
                    cancel: L10n.Promote.leaveCancel
                )
            },
            tone: .irreversible,
            accessibilityIdentifierPrefix: "promote.leave",
            onConfirm: { model.leave() },
            detail: { _ in EmptyView() }
        )
        .accessibilityIdentifier("workspace.destination.promote")
    }

    /// The current step's screen; the flow decides which step that is.
    @ViewBuilder
    private var stepScreen: some View {
        switch model.flow.step {
        case .chooseTarget:
            PromoteTargetStep(
                model: model,
                suggestions: model.session.queryHandle(model.suggestionsKey)
            )
        case .claim:
            PromoteClaimStep(
                model: model,
                grades: model.session.queryHandle(model.confidenceKey),
                properties: model.session.queryHandle(model.propertiesKey)
            )
        case .compare:
            // Not built until S9-19; the flow never stops here.
            EmptyView()
        }
    }

    /// Back, the hint, Done and Next. Everything here reads `model.controls`
    /// (from the flow), never the step itself.
    private var footer: some View {
        let controls = model.controls
        return HStack(spacing: PVSpacing.space6) {
            if let backLabel = model.backLabel {
                PVButton(backLabel, variant: .ghost, icon: .arrowLeft) {
                    model.stepBack()
                }
                .disabled(!controls.canGoBack)
                .accessibilityIdentifier("promote.back")
            }
            Text(verbatim: model.hint)
                .font(PVFont.body(size: PVTypeScale.caption, italic: true))
                .foregroundStyle(PVColor.textMuted)
                .lineLimit(2)
                .accessibilityIdentifier("promote.hint")
            Spacer(minLength: PVSpacing.space6)
            PVButton(L10n.Promote.done, variant: .secondary) {
                model.done()
            }
            .disabled(model.isSaving)
            .accessibilityIdentifier("promote.done")
            PVButton(model.nextLabel, variant: .primary, iconRight: .arrowRight, loading: model.isSaving) {
                Task { await model.next() }
            }
            .disabled(!controls.canAdvance)
            .accessibilityIdentifier("promote.next")
        }
        .padding(.horizontal, PVSpacing.space10)
        .padding(.vertical, 14)
        .background(PVColor.surfaceCard)
        .overlay(alignment: .top) {
            Rectangle().fill(PVColor.borderSubtle).frame(height: 1)
        }
    }

    /// The flow owns the question; dismissing the sheet answers "keep promoting".
    private var leaveBinding: Binding<PromoteModel.PendingLeave?> {
        Binding(
            get: { model.pendingLeave },
            set: { newValue in
                if newValue == nil, model.pendingLeave != nil {
                    model.keepPromoting()
                }
            }
        )
    }
}

/// Tells the flow when the current subject is gone or was promoted elsewhere;
/// the flow sends the place back to the graph.
private struct PromoteSubjectWatch: View {
    @Bindable var handle: QueryHandle<SourceGraphRows>
    let model: PromoteModel

    var body: some View {
        Color.clear
            .onChange(of: model.subjectStatus(rows: handle.value), initial: true) { _, status in
                model.subjectStatusChanged(status)
            }
    }
}

/// Subject header band: the kind's wash, a tile with its mark, the eyebrow,
/// name and ref, the Source, and the step row on the right.
struct PromoteHeader: View {
    let subject: PromoteFlow.Subject
    let sourceTitle: String?
    let steps: [LocalizedStringResource]
    let currentStep: Int

    private var style: EvidenceSubjectKindStyle { .forKind(subject.kind) }

    var body: some View {
        HStack(alignment: .center, spacing: PVSpacing.space6) {
            PromoteKindTile(kind: subject.kind, size: 40, markSize: 20, background: style.chip)
            VStack(alignment: .leading, spacing: PVSpacing.space2) {
                Text(L10n.Promote.eyebrow(subject.kind))
                    .font(PVFont.body(size: PVTypeScale.micro, weight: PVFontWeight.semibold))
                    .tracking(PVTypeScale.micro * PVTracking.caps)
                    .textCase(.uppercase)
                    .foregroundStyle(style.ink)
                HStack(alignment: .firstTextBaseline, spacing: 10) {
                    Text(verbatim: subject.name)
                        .font(PVFont.display(size: 24, weight: PVFontWeight.medium))
                        .tracking(24 * PVTracking.display)
                        .foregroundStyle(PVColor.textDisplay)
                        .lineLimit(1)
                    Text(verbatim: subject.ref)
                        .font(PVFont.mono(size: 12))
                        .foregroundStyle(style.ink)
                }
                if let source = sourceTitle {
                    Text(verbatim: source)
                        .font(PVFont.body(size: PVTypeScale.caption, italic: true))
                        .foregroundStyle(PVColor.textSecondary)
                        .lineLimit(1)
                }
            }
            .accessibilityElement(children: .combine)
            Spacer(minLength: PVSpacing.space6)
            PromoteStepRow(steps: steps, current: currentStep, ink: style.ink, line: style.line)
        }
        .padding(.horizontal, PVSpacing.space10)
        .padding(.vertical, 18)
        .background(style.tint)
        .overlay(alignment: .bottom) {
            Rectangle().fill(style.line).frame(height: 1)
        }
    }
}

/// Step progress composed from text and one icon — no stepper component;
/// this flow is its only user (S9-D9).
struct PromoteStepRow: View {
    let steps: [LocalizedStringResource]
    let current: Int
    let ink: Color
    let line: Color

    var body: some View {
        HStack(spacing: 10) {
            ForEach(Array(steps.enumerated()), id: \.offset) { index, label in
                if index > 0 {
                    PVIcon(.chevronForward, size: 12)
                        .foregroundStyle(PVColor.textFaint)
                        .accessibilityHidden(true)
                }
                step(index: index, label: label)
            }
        }
        .accessibilityElement(children: .contain)
        .accessibilityLabel(Text(L10n.Promote.steps))
    }

    @ViewBuilder
    private func step(index: Int, label: LocalizedStringResource) -> some View {
        let number = Text(verbatim: "\(index + 1)")
        if index == current {
            HStack(spacing: 7) {
                number
                    .font(PVFont.mono(size: PVTypeScale.micro, weight: PVFontWeight.medium))
                    .foregroundStyle(ink)
                Text(label)
                    .font(PVFont.body(size: PVTypeScale.caption, weight: PVFontWeight.medium))
                    .foregroundStyle(PVColor.textPrimary)
                    .fixedSize()
            }
            .padding(.vertical, 5)
            .padding(.horizontal, 10)
            .background(
                RoundedRectangle(cornerRadius: PVRadius.sm, style: .continuous).fill(PVColor.surfaceCard)
            )
            .overlay(
                RoundedRectangle(cornerRadius: PVRadius.sm, style: .continuous).strokeBorder(line, lineWidth: 1)
            )
            .accessibilityElement(children: .combine)
            .accessibilityAddTraits(.isSelected)
        } else {
            HStack(spacing: 7) {
                number
                    .font(PVFont.mono(size: PVTypeScale.micro))
                Text(label)
                    .font(PVFont.body(size: PVTypeScale.caption))
                    .fixedSize()
            }
            .foregroundStyle(PVColor.textMuted)
            .padding(.vertical, 5)
            .accessibilityElement(children: .combine)
        }
    }
}

/// A kind's mark on a small tile in the kind's colours.
struct PromoteKindTile: View {
    let kind: EvidencePrimaryKind
    let size: CGFloat
    let markSize: CGFloat
    let background: Color

    var body: some View {
        let style = EvidenceSubjectKindStyle.forKind(kind)
        RoundedRectangle(cornerRadius: size >= 40 ? 5 : 4, style: .continuous)
            .fill(background)
            .overlay(
                RoundedRectangle(cornerRadius: size >= 40 ? 5 : 4, style: .continuous)
                    .strokeBorder(style.line, lineWidth: 1)
            )
            .overlay(
                PVMark(kind.markKey, size: markSize, decorative: true)
                    .foregroundStyle(style.ink)
            )
            .frame(width: size, height: size)
            .accessibilityHidden(true)
    }
}
