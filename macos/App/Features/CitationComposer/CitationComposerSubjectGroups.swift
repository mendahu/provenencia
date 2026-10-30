import SwiftUI

/// Observations column: header, authoring box, subject cards.
struct CitationComposerObservationsColumn: View {
    @Bindable var model: CitationComposerModel
    var onDone: () -> Void
    @State private var hoveredLineID: String?
    @FocusState private var lineFocus: LineControl?

    /// The row and its Edit, Duplicate, and Delete buttons share one focus
    /// value so the buttons stay on screen while keyboard focus is inside the row.
    private enum LineControl: Hashable {
        case row(String), edit(String), duplicate(String), delete(String)
    }

    var body: some View {
        VStack(alignment: .leading, spacing: 0) {
            header
                .padding(.horizontal, PVSpacing.space7)
                .padding(.top, 14)
                .padding(.bottom, PVSpacing.space4)
            ScrollView {
                groups
                    .padding(.horizontal, PVSpacing.space7)
                    .padding(.top, PVSpacing.space1)
                    .padding(.bottom, PVSpacing.space8)
            }
            .frame(maxWidth: .infinity, maxHeight: .infinity, alignment: .topLeading)
        }
        .background(PVColor.surfacePage)
        .overlay(alignment: .bottom) {
            if let toast = model.bridgeToast {
                PVToast(tone: .success, title: toast.title, message: toast.message) {
                    model.bridgeToast = nil
                }
                .padding(PVSpacing.space5)
            }
        }
        .accessibilityElement(children: .contain)
        .accessibilityLabel(Text(L10n.CitationComposer.observationsSection))
    }

    private var header: some View {
        PVSectionHeader(
            title: L10n.CitationComposer.observationsSection,
            meta: model.observationsMeta,
            actions: {
                HStack(spacing: PVSpacing.space4) {
                    if case .bridge(let draft) = model.authoring, draft.holdsLeave {
                        Text(L10n.CitationComposer.unsavedDraftBridge)
                            .font(PVFont.body(size: 12, italic: true))
                            .foregroundStyle(PVColor.textMuted)
                    } else if case .property(let draft) = model.authoring, draft.holdsLeave {
                        Text(L10n.CitationComposer.unsavedDraftObservation)
                            .font(PVFont.body(size: 12, italic: true))
                            .foregroundStyle(PVColor.textMuted)
                    }
                    PVButton(L10n.CitationComposer.addObservation, variant: .secondary, size: .sm, icon: .plus) {
                        model.beginAddObservation()
                    }
                    .disabled(model.isDrawingDepiction)
                    .accessibilityIdentifier("citationComposer.addObservation")
                    PVButton(L10n.CitationComposer.done, variant: .primary, size: .sm) {
                        onDone()
                    }
                    .disabled(model.isDrawingDepiction)
                    .accessibilityIdentifier("citationComposer.done")
                }
            }
        )
    }

    @ViewBuilder
    private var groups: some View {
        if model.subjectGroups.isEmpty {
            PVEmptyState(
                icon: .listPlus,
                title: L10n.CitationComposer.noObservations,
                message: L10n.string(L10n.CitationComposer.emptyObservationsMessage),
                compact: true
            ) {
                EmptyView()
            }
            .accessibilityIdentifier("citationComposer.observations.empty")
        } else {
            VStack(alignment: .leading, spacing: PVSpacing.space5) {
                ForEach(model.subjectGroups) { group in
                    subjectCard(group)
                }
            }
        }
    }

    private func subjectCard(_ group: CitationSubjectGroup) -> some View {
        PVCard(padding: 0) {
            VStack(alignment: .leading, spacing: 0) {
                HStack(spacing: PVSpacing.space4) {
                    PVMark(PVMarkKey(subjectTypeKey: group.typeKey) ?? .fallback, size: 16, decorative: true)
                    Text(verbatim: group.name)
                        .font(PVFont.body(size: PVTypeScale.bodySmall, weight: PVFontWeight.semibold))
                        .foregroundStyle(PVColor.textPrimary)
                        .lineLimit(1)
                    if !group.ref.isEmpty {
                        Text(verbatim: group.ref)
                            .font(PVFont.mono(size: PVTypeScale.micro))
                            .foregroundStyle(PVColor.textMuted)
                            .lineLimit(1)
                    }
                    Spacer(minLength: 0)
                    Text(verbatim: "\(group.lines.count)")
                        .font(PVFont.mono(size: PVTypeScale.micro))
                        .foregroundStyle(PVColor.textMuted)
                    PVIconButton(
                        .plus,
                        label: L10n.CitationComposer.addForNamedSubject(group.name),
                        size: .sm
                    ) {
                        model.beginAddObservation(for: group.subjectID)
                    }
                    .disabled(model.isDrawingDepiction)
                    .accessibilityIdentifier("citationComposer.group.add.\(group.subjectID)")
                }
                .padding(.leading, PVSpacing.space5)
                .padding(.trailing, PVSpacing.space3)
                .frame(height: 38)
                .overlay(alignment: .bottom) {
                    PVDivider(axis: .horizontal, color: PVColor.borderSubtle)
                }
                ForEach(Array(group.lines.enumerated()), id: \.element.id) { index, line in
                    lineRow(line, rule: index > 0)
                }
            }
        }
    }

    private func lineRow(_ line: CitationSavedLine, rule: Bool) -> some View {
        let revealed = hoveredLineID == line.id || lineOwnsFocus(line.id)
        let twin = line.bridgeID.map { bridgeID in
            hoveredLineID?.contains("bridge:\(bridgeID):") == true && hoveredLineID != line.id
        } ?? false
        let fresh = isFresh(line)
        // A likeness line centres its label against the thumbnail; text lines share a baseline.
        let rowAlignment: VerticalAlignment = line.imageRelPath == nil ? .firstTextBaseline : .center
        return HStack(alignment: rowAlignment, spacing: PVSpacing.space4) {
            Group {
                if let markKey = line.markKey {
                    PVMark(markKey, size: 14, decorative: true)
                } else {
                    Color.clear
                }
            }
            .frame(width: 16, height: 16)
            Text(verbatim: line.label)
                .font(PVFont.body(size: 12))
                .foregroundStyle(PVColor.textSecondary)
                .lineLimit(2)
                .multilineTextAlignment(.leading)
                .fixedSize(horizontal: false, vertical: true)
                .frame(minWidth: 92, maxWidth: 124, alignment: .leading)
            HStack(alignment: rowAlignment, spacing: PVSpacing.space2) {
                if line.negative {
                    Text(L10n.CitationComposer.notWord)
                        .italic()
                        .foregroundStyle(PVColor.textMuted)
                        .fixedSize(horizontal: false, vertical: true)
                }
                if let relPath = line.imageRelPath {
                    CachedThumbnail(
                        projectDir: model.session.projectKey.projectDir,
                        relPath: relPath,
                        mediaType: "image/png",
                        size: 32
                    )
                    .accessibilityLabel(Text(verbatim: depictionThumbLabel(line) ?? line.label))
                } else {
                    Text(verbatim: line.valueText.isEmpty ? "—" : line.valueText)
                        .font(line.mono ? PVFont.mono(size: PVTypeScale.caption) : PVFont.body(size: PVTypeScale.bodySmall))
                        .foregroundStyle(PVColor.textPrimary)
                        .multilineTextAlignment(.leading)
                        .fixedSize(horizontal: false, vertical: true)
                }
            }
            .frame(maxWidth: .infinity, alignment: .leading)
            if line.negative {
                PVBadge(L10n.CitationComposer.negativeLabel, tone: .neutral, icon: .circleMinus, subtle: true)
            }
            lineActions(line, revealed: revealed, fresh: fresh)
        }
        .padding(.vertical, line.imageRelPath == nil ? PVSpacing.space3 : PVSpacing.space1)
        .padding(.leading, PVSpacing.space5)
        .padding(.trailing, PVSpacing.space3)
        .frame(minHeight: 32, alignment: .top)
        .background(lineBackground(fresh: fresh, revealed: revealed, twin: twin))
        .overlay(alignment: .top) {
            if rule {
                PVDivider(axis: .horizontal, color: PVColor.borderSubtle)
            }
        }
        .contentShape(Rectangle())
        .disabled(model.isDrawingDepiction)
        .focusable()
        .focused($lineFocus, equals: .row(line.id))
        .onTapGesture { model.beginEdit(line: line) }
        .onKeyPress(.return) {
            guard lineFocus == .row(line.id) else { return .ignored }
            model.beginEdit(line: line)
            return .handled
        }
        .onKeyPress(.space) {
            guard lineFocus == .row(line.id) else { return .ignored }
            model.beginEdit(line: line)
            return .handled
        }
        .onHover { hovering in
            hoveredLineID = hovering ? line.id : (hoveredLineID == line.id ? nil : hoveredLineID)
        }
        .accessibilityElement(children: .combine)
        .lineReading(depictionReading(line, fresh: fresh))
        .accessibilityAddTraits(.isButton)
        .accessibilityAction(.default) { model.beginEdit(line: line) }
        .accessibilityAction(named: Text(line.bridgeID == nil
            ? L10n.CitationComposer.editObservation
            : L10n.CitationComposer.editBridgeTitle)) {
            model.beginEdit(line: line)
        }
        .accessibilityAction(named: Text(line.bridgeID == nil
            ? L10n.CitationComposer.duplicateObservation
            : L10n.CitationComposer.duplicateBridge)) {
            model.duplicate(line: line)
        }
        .accessibilityAction(named: Text(line.bridgeID == nil
            ? L10n.CitationComposer.deleteObservation
            : L10n.CitationComposer.deleteBridge)) {
            Task { await model.askDeleteLine(line) }
        }
        .accessibilityIdentifier("citationComposer.line.\(line.id)")
    }

    /// The label of the Subject the line depicts; nil when it is unknown.
    private func depictionSubject(_ line: CitationSavedLine) -> String? {
        model.vocabulary.subjectsByID[line.subjectID]?.label
    }

    /// "Likeness of Ben Ford"; nil when the Subject is unknown.
    private func depictionThumbLabel(_ line: CitationSavedLine) -> String? {
        depictionSubject(line).map { L10n.CitationComposer.depictionThumbLabel(line.label, $0) }
    }

    /// What VoiceOver reads for a saved likeness: "Likeness, Ben Ford" (the Subject's label) with the
    /// badges the row shows (Negative, Saved) as its value. Nil for other lines,
    /// which keep their combined reading.
    private func depictionReading(_ line: CitationSavedLine, fresh: Bool) -> LineReading? {
        guard line.imageRelPath != nil else { return nil }
        let label = depictionSubject(line).map { L10n.CitationComposer.depictionLineLabel(line.label, $0) }
            ?? line.label
        var badges: [String] = []
        if line.negative { badges.append(L10n.string(L10n.CitationComposer.negativeLabel)) }
        if fresh { badges.append(L10n.string(L10n.CitationComposer.savedBadge)) }
        return LineReading(label: label, value: badges.joined(separator: ", "))
    }

    @ViewBuilder
    private func lineActions(_ line: CitationSavedLine, revealed: Bool, fresh: Bool) -> some View {
        HStack(spacing: PVSpacing.space2) {
            if revealed {
                PVIconButton(
                    .penLine,
                    label: line.bridgeID == nil
                        ? L10n.CitationComposer.editObservation
                        : L10n.CitationComposer.editBridgeTitle,
                    size: .sm
                ) {
                    model.beginEdit(line: line)
                }
                .focused($lineFocus, equals: .edit(line.id))
                PVIconButton(
                    .duplicate,
                    label: line.bridgeID == nil
                        ? L10n.CitationComposer.duplicateObservation
                        : L10n.CitationComposer.duplicateBridge,
                    size: .sm
                ) {
                    model.duplicate(line: line)
                }
                .focused($lineFocus, equals: .duplicate(line.id))
                PVIconButton(
                    .trash,
                    label: line.bridgeID == nil
                        ? L10n.CitationComposer.deleteObservation
                        : L10n.CitationComposer.deleteBridge,
                    size: .sm
                ) {
                    Task { await model.askDeleteLine(line) }
                }
                .focused($lineFocus, equals: .delete(line.id))
            } else if fresh {
                PVBadge(L10n.CitationComposer.savedBadge, tone: .success, icon: .check, subtle: true)
            }
        }
        .frame(width: 88, alignment: .trailing)
    }

    private func lineOwnsFocus(_ id: String) -> Bool {
        switch lineFocus {
        case .row(let lineID), .edit(let lineID), .duplicate(let lineID), .delete(let lineID):
            lineID == id
        case nil:
            false
        }
    }

    private func isFresh(_ line: CitationSavedLine) -> Bool {
        if model.freshLineKey == line.id || model.freshLineKey == line.sortKey { return true }
        if let bridgeID = line.bridgeID, model.freshLineKey == "bridge:\(bridgeID)" { return true }
        if let observationID = line.observationID, model.freshLineKey == "obs:\(observationID)" { return true }
        return false
    }

    private func lineBackground(fresh: Bool, revealed: Bool, twin: Bool) -> Color {
        if fresh { return PVColor.successSoft }
        if revealed || twin { return PVColor.surfaceActive }
        return .clear
    }

}

/// Edit sheet for one saved property observation or one bridge.
struct CitationComposerEditSheet: View {
    @Bindable var model: CitationComposerModel
    @FocusState private var focus: CitationAuthoringFocus?
    /// Placed once. `.defaultFocus` would reapply this when the value editor changes.
    @State private var didPlaceOpeningFocus = false

    var body: some View {
        editBody
            .onAppear { placeOpeningFocusIfNeeded() }
    }

    private func placeOpeningFocusIfNeeded() {
        guard !didPlaceOpeningFocus else { return }
        didPlaceOpeningFocus = true
        let initial = openingFocus
        DispatchQueue.main.async { focus = initial }
    }

    @ViewBuilder
    private var editBody: some View {
        if let session = model.editSession {
            switch session.kind {
            case .property(let values, _, _):
                propertyForm(values, error: session.error)
            case .bridge(let endA, let endB, _, _, _, _):
                bridgeForm(endA: endA, endB: endB, error: session.error)
            }
        }
    }

    private var openingFocus: CitationAuthoringFocus {
        guard let session = model.editSession else { return .subject }
        switch session.kind {
        case .property(let values, _, _):
            return CitationAuthoringNavigation.initialPropertyFocus(values: values, vocabulary: model.vocabulary)
        case .bridge(let endA, let endB, let termID, _, _, _):
            let pair = CitationComposerAuthoring.resolvePair(
                endA: endA,
                endB: endB,
                termID: termID,
                vocabulary: model.vocabulary
            )
            let hasTerm = if case .ready(let ready) = pair { ready.termProperty != nil } else { false }
            return CitationAuthoringNavigation.initialBridgeFocus(
                endA: endA,
                endB: endB,
                hasTermField: hasTerm,
                termID: termID
            )
        }
    }

    private func propertyForm(_ values: ObservationRowValues, error: String?) -> some View {
        VStack(alignment: .leading, spacing: PVSpacing.space5) {
            CitationComposerModalLead(
                model: model,
                isProperty: true,
                lockKind: true,
                citationSelection: editCitationBinding,
                citationEnabled: false,
                onProperty: {},
                onBridge: {}
            )
            CitationPropertyFields(
                model: model,
                values: editValuesBinding,
                subjectField: .propertySubject,
                error: error,
                identifierPrefix: "citationComposer.edit",
                focus: $focus,
                applyTermToEdit: true
            )
        }
    }
    private func bridgeForm(endA: String, endB: String, error: String?) -> some View {
        let pair = CitationComposerAuthoring.resolvePair(
            endA: endA,
            endB: endB,
            termID: editBridgeTermBinding.wrappedValue,
            vocabulary: model.vocabulary
        )
        return VStack(alignment: .leading, spacing: PVSpacing.space5) {
            CitationComposerModalLead(
                model: model,
                isProperty: false,
                lockKind: true,
                citationSelection: editCitationBinding,
                citationEnabled: false,
                onProperty: {},
                onBridge: {}
            )
            HStack(alignment: .top, spacing: PVSpacing.space5) {
                lockedEnd(id: endA, title: L10n.string(L10n.CitationComposer.endA))
                lockedEnd(id: endB, title: L10n.string(L10n.CitationComposer.endB))
            }
            switch pair {
            case .incomplete(let hint):
                Text(verbatim: hint)
                    .font(PVFont.body(size: 12, italic: true))
                    .foregroundStyle(PVColor.textMuted)
            case .refused(let title, let message):
                Text(L10n.CitationComposer.noBridgeJoins)
                    .font(PVFont.body(size: 12, italic: true))
                    .foregroundStyle(PVColor.textMuted)
                PVCallout(tone: .warning, title: title, message: message)
            case .ready(let ready):
                CitationBridgeSentence(
                    ready: ready,
                    subjectA: model.vocabulary.subjectsByID[endA],
                    subjectB: model.vocabulary.subjectsByID[endB],
                    terms: model.vocabulary.termsByPropertyID[ready.termProperty?.id ?? ""] ?? [],
                    termID: editBridgeTermBinding,
                    canSwap: false,
                    onSwap: {},
                    onAddTerm: {
                        if let property = ready.termProperty {
                            model.beginAddCustomTerm(propertyID: property.id, applyToEdit: true)
                        }
                    },
                    termFocus: PVComboFocus($focus, equals: .term),
                    termIdentifierPrefix: "citationComposer.edit.bridgeTerm"
                )
            }
            Text(L10n.CitationComposer.lockEndsNote)
                .font(PVFont.body(size: 12, italic: true))
                .foregroundStyle(PVColor.textMuted)
                .fixedSize(horizontal: false, vertical: true)
            if let error {
                PVCallout(tone: .danger, title: L10n.CitationComposer.rowStateError, message: error)
            }
        }
    }

    private func lockedEnd(id: String, title: String) -> some View {
        let subject = model.vocabulary.subjectsByID[id]
        return VStack(alignment: .leading, spacing: PVSpacing.space2) {
            Text(verbatim: title)
                .font(PVFont.body(size: PVTypeScale.caption))
                .foregroundStyle(PVColor.textMuted)
            HStack(spacing: PVSpacing.space4) {
                if let typeKey = subject?.typeKey {
                    PVMark(PVMarkKey(subjectTypeKey: typeKey) ?? .fallback, size: 14, decorative: true)
                }
                Text(verbatim: subject?.label ?? id)
                    .font(PVFont.body(size: PVTypeScale.bodySmall))
                    .foregroundStyle(PVColor.textPrimary)
                if let ref = subject?.ref, !ref.isEmpty {
                    Text(verbatim: ref)
                        .font(PVFont.mono(size: PVTypeScale.micro))
                        .foregroundStyle(PVColor.textMuted)
                }
            }
            .frame(minHeight: PVSpacing.controlHeightSmall, alignment: .leading)
        }
    }

    private var editCitationBinding: Binding<String> {
        Binding(
            get: { model.activeCitationID ?? "" },
            set: { _ in }
        )
    }

    private var editValuesBinding: Binding<ObservationRowValues> {
        Binding(
            get: {
                if case .property(let values, _, _) = model.editSession?.kind { return values }
                return ObservationRowValues()
            },
            set: { newValue in
                model.mutateEdit { session in
                    guard case .property(_, let id, let ref) = session.kind else { return }
                    session.kind = .property(values: newValue, persistedID: id, persistedRef: ref)
                }
            }
        )
    }

    private var editBridgeTermBinding: Binding<String> {
        Binding(
            get: {
                if case .bridge(_, _, let termID, _, _, _) = model.editSession?.kind { return termID }
                return ""
            },
            set: { newValue in
                model.mutateEdit { session in
                    guard case .bridge(let endA, let endB, _, let bridgeID, let bridgeRef, let baseline) = session.kind
                    else { return }
                    session.kind = .bridge(
                        endA: endA,
                        endB: endB,
                        termID: newValue,
                        bridgeID: bridgeID,
                        bridgeRef: bridgeRef,
                        baselineTermID: baseline
                    )
                }
            }
        )
    }
}

/// A line's VoiceOver label and value when the combined reading would drop a badge.
private struct LineReading {
    var label: String
    var value: String
}

private extension View {
    /// Overrides the combined reading when the line has its own; nil keeps it.
    @ViewBuilder
    func lineReading(_ reading: LineReading?) -> some View {
        if let reading {
            accessibilityLabel(Text(verbatim: reading.label))
                .accessibilityValue(Text(verbatim: reading.value))
        } else {
            self
        }
    }
}
