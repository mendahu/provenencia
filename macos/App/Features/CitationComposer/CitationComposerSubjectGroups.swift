import SwiftUI

/// Observations column: header, authoring box, subject cards.
struct CitationComposerObservationsColumn: View {
    @Bindable var model: CitationComposerModel
    var onDone: () -> Void
    @State private var hoveredLineID: String?

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
                    .accessibilityIdentifier("citationComposer.addObservation")
                    PVButton(L10n.CitationComposer.done, variant: .primary, size: .sm) {
                        onDone()
                    }
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
                    PVMark(PVMarkKey(catalogKey: "subject_\(group.typeKey)"), size: 16, decorative: true)
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
        let revealed = hoveredLineID == line.id
        let twin = line.bridgeID.map { bridgeID in
            hoveredLineID?.contains("bridge:\(bridgeID):") == true && hoveredLineID != line.id
        } ?? false
        let fresh = isFresh(line)
        return HStack(alignment: .firstTextBaseline, spacing: PVSpacing.space4) {
            Group {
                if let markKey = line.markKey {
                    PVMark(PVMarkKey(catalogKey: markKey), size: 14, decorative: true)
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
            HStack(alignment: .firstTextBaseline, spacing: PVSpacing.space2) {
                if line.negative {
                    Text(L10n.CitationComposer.notWord)
                        .italic()
                        .foregroundStyle(PVColor.textMuted)
                        .fixedSize(horizontal: false, vertical: true)
                }
                Text(verbatim: line.valueText.isEmpty ? "—" : line.valueText)
                    .font(line.mono ? PVFont.mono(size: PVTypeScale.caption) : PVFont.body(size: PVTypeScale.bodySmall))
                    .foregroundStyle(PVColor.textPrimary)
                    .multilineTextAlignment(.leading)
                    .fixedSize(horizontal: false, vertical: true)
            }
            .frame(maxWidth: .infinity, alignment: .leading)
            if line.negative {
                PVBadge(L10n.CitationComposer.negativeLabel, tone: .neutral, icon: .circleMinus, subtle: true)
            }
            lineActions(line, revealed: revealed, fresh: fresh)
        }
        .padding(.vertical, PVSpacing.space3)
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
        .onTapGesture { model.beginEdit(line: line) }
        .onHover { hovering in
            hoveredLineID = hovering ? line.id : (hoveredLineID == line.id ? nil : hoveredLineID)
        }
        .accessibilityElement(children: .combine)
        .accessibilityIdentifier("citationComposer.line.\(line.id)")
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
                PVIconButton(
                    .duplicate,
                    label: line.bridgeID == nil
                        ? L10n.CitationComposer.duplicateObservation
                        : L10n.CitationComposer.duplicateBridge,
                    size: .sm
                ) {
                    model.duplicate(line: line)
                }
                PVIconButton(
                    .trash,
                    label: line.bridgeID == nil
                        ? L10n.CitationComposer.deleteObservation
                        : L10n.CitationComposer.deleteBridge,
                    size: .sm
                ) {
                    Task { await model.askDeleteLine(line) }
                }
            } else if fresh {
                PVBadge(L10n.CitationComposer.savedBadge, tone: .success, icon: .check, subtle: true)
            }
        }
        .frame(width: 88, alignment: .trailing)
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

    var body: some View {
        if let session = model.editSession {
            switch session.kind {
            case .property(let values, _, _):
                propertyForm(values, error: session.error)
            case .bridge(let endA, let endB, _, _, _, _):
                bridgeForm(endA: endA, endB: endB, error: session.error)
            }
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
            HStack(alignment: .top, spacing: PVSpacing.space5) {
                PVField(label: L10n.CitationComposer.subjectLabel) {
                    PVComboBox(
                        selection: editSubjectBinding,
                        options: model.subjectOptions,
                        size: .sm,
                        placeholder: L10n.CitationComposer.subjectPlaceholder,
                        emptyLabel: L10n.CitationComposer.subjectPlaceholder,
                        label: L10n.CitationComposer.subjectLabel,
                        accessibilityIdentifierPrefix: "citationComposer.edit.subject",
                        activateOnAppear: model.editSession?.focus == .subject,
                        focusNonce: model.editSession?.focus == .subject ? model.authoringFocusNonce : 0,
                        onCommit: { model.advanceEditFocus(from: .subject, committed: $0) }
                    )
                }
                .frame(maxWidth: .infinity)
                PVField(label: L10n.CitationComposer.propertyLabel) {
                    PVComboBox(
                        selection: editPropertyBinding,
                        options: model.vocabulary.propertyOptions(forSubjectID: values.subjectID).map {
                            PVComboBoxOption(value: $0.id, label: $0.label, subtext: $0.key)
                        },
                        size: .sm,
                        placeholder: L10n.CitationComposer.propertyPlaceholder,
                        emptyLabel: L10n.CitationComposer.propertyPlaceholder,
                        label: L10n.CitationComposer.propertyLabel,
                        accessibilityIdentifierPrefix: "citationComposer.edit.property",
                        activateOnAppear: model.editSession?.focus == .property,
                        focusNonce: model.editSession?.focus == .property ? model.authoringFocusNonce : 0,
                        onCommit: { model.advanceEditFocus(from: .property, committed: $0) }
                    )
                }
                .frame(maxWidth: .infinity)
            }
            editValue(values, focusValue: model.editSession?.focus == .value)
            PVCheckbox(
                L10n.CitationComposer.negativeLabel,
                description: L10n.CitationComposer.negativeDescription,
                isChecked: editNegativeBinding
            )
            if let error {
                PVCallout(tone: .danger, message: error)
            }
        }
    }

    @ViewBuilder
    private func editValue(_ values: ObservationRowValues, focusValue: Bool) -> some View {
        let property = model.vocabulary.property(id: values.propertyID)
        let nonce = focusValue ? model.authoringFocusNonce : 0
        switch property?.valueType {
        case PropertyValueType.text.rawValue:
            PVInput(
                text: editTextBinding,
                size: .sm,
                activateOnAppear: focusValue,
                focusNonce: nonce
            )
        case PropertyValueType.integer.rawValue:
            PVInput(
                text: editIntegerBinding,
                size: .sm,
                mono: true,
                activateOnAppear: focusValue,
                focusNonce: nonce
            )
        case PropertyValueType.term.rawValue:
            HStack(spacing: PVSpacing.space3) {
                PVComboBox(
                    selection: editTermBinding,
                    options: model.vocabulary.termComboOptionsByPropertyID[values.propertyID] ?? [],
                    size: .sm,
                    placeholder: L10n.CitationComposer.termPlaceholder,
                    emptyLabel: L10n.CitationComposer.termEmpty,
                    label: L10n.CitationComposer.termLabel,
                    accessibilityIdentifierPrefix: "citationComposer.edit.term",
                    activateOnAppear: focusValue,
                    focusNonce: nonce
                )
                PVButton(L10n.CitationComposer.addCustomTerm, variant: .ghost, size: .sm) {
                    model.beginAddCustomTerm(propertyID: values.propertyID, applyToEdit: true)
                }
            }
        case PropertyValueType.name.rawValue:
            editInlineValue {
                NameValueEditorForm(
                    draft: editNameBinding,
                    accessibilityIdentifierPrefix: "citationComposer.edit.name",
                    focusNonce: nonce
                )
            }
        case PropertyValueType.date.rawValue:
            editInlineValue {
                DateValueEditorForm(
                    draft: editDateBinding,
                    accessibilityIdentifierPrefix: "citationComposer.edit.date",
                    focusNonce: nonce
                )
            }
        default:
            EmptyView()
        }
    }

    private func editInlineValue<Content: View>(@ViewBuilder content: () -> Content) -> some View {
        content()
            .padding(.horizontal, 14)
            .padding(.vertical, PVSpacing.space5)
            .frame(maxWidth: .infinity, alignment: .leading)
            .background(PVColor.surfaceSunken)
            .clipShape(RoundedRectangle(cornerRadius: PVRadius.md, style: .continuous))
            .overlay(
                RoundedRectangle(cornerRadius: PVRadius.md, style: .continuous)
                    .stroke(PVColor.borderSubtle, lineWidth: 1)
            )
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
                    termIdentifierPrefix: "citationComposer.edit.bridgeTerm"
                )
            }
            Text(L10n.CitationComposer.lockEndsNote)
                .font(PVFont.body(size: 12, italic: true))
                .foregroundStyle(PVColor.textMuted)
                .fixedSize(horizontal: false, vertical: true)
            if let error {
                PVCallout(tone: .danger, message: error)
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
                    PVMark(PVMarkKey(catalogKey: "subject_\(typeKey)"), size: 14, decorative: true)
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

    private var editDateBinding: Binding<DateValueDraft> {
        Binding(
            get: {
                if case .property(let values, _, _) = model.editSession?.kind { return values.dateDraft }
                return .empty()
            },
            set: { newValue in
                model.mutateEdit { session in
                    guard case .property(var values, let id, let ref) = session.kind else { return }
                    values.dateDraft = newValue
                    session.kind = .property(values: values, persistedID: id, persistedRef: ref)
                }
            }
        )
    }

    private var editNameBinding: Binding<NameValueDraft> {
        Binding(
            get: {
                if case .property(let values, _, _) = model.editSession?.kind { return values.nameDraft }
                return .empty()
            },
            set: { newValue in
                model.mutateEdit { session in
                    guard case .property(var values, let id, let ref) = session.kind else { return }
                    values.nameDraft = newValue
                    session.kind = .property(values: values, persistedID: id, persistedRef: ref)
                }
            }
        )
    }

    private var editSubjectBinding: Binding<String> {
        Binding(
            get: {
                if case .property(let values, _, _) = model.editSession?.kind { return values.subjectID }
                return ""
            },
            set: { newValue in
                model.mutateEdit { session in
                    guard case .property(var values, let id, let ref) = session.kind else { return }
                    values.subjectID = newValue
                    session.kind = .property(values: values, persistedID: id, persistedRef: ref)
                }
            }
        )
    }

    private var editPropertyBinding: Binding<String> {
        Binding(
            get: {
                if case .property(let values, _, _) = model.editSession?.kind { return values.propertyID }
                return ""
            },
            set: { newValue in
                model.mutateEdit { session in
                    guard case .property(var values, let id, let ref) = session.kind else { return }
                    values.propertyID = newValue
                    session.kind = .property(values: values, persistedID: id, persistedRef: ref)
                }
            }
        )
    }

    private var editTextBinding: Binding<String> {
        Binding(
            get: {
                if case .property(let values, _, _) = model.editSession?.kind { return values.valueText }
                return ""
            },
            set: { newValue in
                model.mutateEdit { session in
                    guard case .property(var values, let id, let ref) = session.kind else { return }
                    values.valueText = newValue
                    session.kind = .property(values: values, persistedID: id, persistedRef: ref)
                }
            }
        )
    }

    private var editIntegerBinding: Binding<String> {
        Binding(
            get: {
                if case .property(let values, _, _) = model.editSession?.kind { return values.valueIntegerText }
                return ""
            },
            set: { newValue in
                model.mutateEdit { session in
                    guard case .property(var values, let id, let ref) = session.kind else { return }
                    values.valueIntegerText = newValue
                    session.kind = .property(values: values, persistedID: id, persistedRef: ref)
                }
            }
        )
    }

    private var editTermBinding: Binding<String> {
        Binding(
            get: {
                if case .property(let values, _, _) = model.editSession?.kind { return values.valueTermID }
                return ""
            },
            set: { newValue in
                model.mutateEdit { session in
                    guard case .property(var values, let id, let ref) = session.kind else { return }
                    values.valueTermID = newValue
                    session.kind = .property(values: values, persistedID: id, persistedRef: ref)
                }
            }
        )
    }

    private var editNegativeBinding: Binding<Bool> {
        Binding(
            get: {
                if case .property(let values, _, _) = model.editSession?.kind {
                    return values.polarity == ObservationPolarity.negative.rawValue
                }
                return false
            },
            set: { checked in
                model.mutateEdit { session in
                    guard case .property(var values, let id, let ref) = session.kind else { return }
                    values.polarity = checked
                        ? ObservationPolarity.negative.rawValue
                        : ObservationPolarity.positive.rawValue
                    session.kind = .property(values: values, persistedID: id, persistedRef: ref)
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
