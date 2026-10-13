import SwiftUI

/// Observations column: header, authoring box, subject cards.
struct CitationComposerObservationsColumn: View {
    @Bindable var model: CitationComposerModel
    var onDone: () -> Void
    @State private var hoveredLineID: String?
    @State private var openMenuLineID: String?

    var body: some View {
        VStack(alignment: .leading, spacing: 0) {
            header
                .padding(.horizontal, PVSpacing.space6)
                .padding(.top, PVSpacing.space6)
                .padding(.bottom, PVSpacing.space4)
            CitationComposerAuthoringBox(model: model)
                .padding(.horizontal, PVSpacing.space6)
                .padding(.bottom, PVSpacing.space4)
            ScrollView {
                groups
                    .padding(.horizontal, PVSpacing.space6)
                    .padding(.bottom, PVSpacing.space6)
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
            aside: {
                if case .bridge(let draft) = model.authoring, draft.holdsLeave {
                    Text(L10n.CitationComposer.unsavedDraftBridge)
                        .font(PVFont.body(size: PVTypeScale.caption, italic: true))
                        .foregroundStyle(PVColor.textMuted)
                } else if case .property(let draft) = model.authoring, draft.holdsLeave {
                    Text(L10n.CitationComposer.unsavedDraftObservation)
                        .font(PVFont.body(size: PVTypeScale.caption, italic: true))
                        .foregroundStyle(PVColor.textMuted)
                }
            },
            actions: {
                PVButton(L10n.CitationComposer.done, variant: .primary, size: .sm) {
                    onDone()
                }
                .accessibilityIdentifier("citationComposer.done")
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
        let revealed = hoveredLineID == line.id || openMenuLineID == line.id
        let twin = line.bridgeID.map { bridgeID in
            hoveredLineID?.contains("bridge:\(bridgeID):") == true && hoveredLineID != line.id
        } ?? false
        let fresh = isFresh(line)
        return HStack(alignment: .center, spacing: PVSpacing.space4) {
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
                .frame(minWidth: 92, maxWidth: 124, alignment: .leading)
            HStack(alignment: .firstTextBaseline, spacing: PVSpacing.space2) {
                if line.negative {
                    Text(L10n.CitationComposer.notWord)
                        .italic()
                        .foregroundStyle(PVColor.textMuted)
                }
                Text(verbatim: line.valueText.isEmpty ? "—" : line.valueText)
                    .font(line.mono ? PVFont.mono(size: PVTypeScale.caption) : PVFont.body(size: PVTypeScale.bodySmall))
                    .foregroundStyle(PVColor.textPrimary)
            }
            .frame(maxWidth: .infinity, alignment: .leading)
            if line.negative {
                PVBadge(L10n.CitationComposer.negativeLabel, tone: .neutral, icon: .circleMinus, subtle: true)
            }
            if fresh {
                PVBadge(L10n.CitationComposer.savedBadge, tone: .success, icon: .check, subtle: true)
            }
            ObservationLineMenu(model: model, line: line, revealed: revealed) { presented in
                if presented {
                    openMenuLineID = line.id
                } else if openMenuLineID == line.id {
                    openMenuLineID = nil
                }
            }
        }
        .padding(.vertical, PVSpacing.space1)
        .padding(.leading, PVSpacing.space5)
        .padding(.trailing, PVSpacing.space3)
        .frame(minHeight: 32)
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

/// More-actions menu for one saved line. The panel supplies the opaque card;
/// the menu window itself is clear.
private struct ObservationLineMenu: View {
    @Bindable var model: CitationComposerModel
    let line: CitationSavedLine
    var revealed: Bool
    var onPresented: (Bool) -> Void
    @State private var menu = PVContextMenuState()
    @State private var keyboard = PVContextMenuKeyboard.inactive

    var body: some View {
        HStack(spacing: PVSpacing.space1) {
            if revealed || menu.isPresented {
                PVIconButton(
                    .duplicate,
                    label: line.bridgeID == nil
                        ? L10n.CitationComposer.duplicateObservation
                        : L10n.CitationComposer.duplicateBridge,
                    size: .sm
                ) {
                    model.duplicate(line: line)
                }
                PVIconButton(.ellipsis, label: L10n.CitationComposer.moreActions, size: .sm) {
                    toggle()
                }
                .accessibilityIdentifier("citationComposer.line.more.\(line.id)")
                .pvContextMenu($menu, keyboard: $keyboard, dismissOnClickAway: true) {
                    PVContextMenuPanel(width: 220, accessibilityIdentifier: "citationComposer.lineMenu") {
                        menuItems
                    }
                }
            }
        }
        .frame(width: 58, alignment: .trailing)
        .onChange(of: menu.isPresented) { _, presented in
            onPresented(presented)
        }
    }

    private func toggle() {
        if menu.isPresented {
            menu.dismiss()
            return
        }
        let titles = menuTitles.map { L10n.string($0) }
        keyboard = PVContextMenuKeyboard(itemCount: titles.count, activeIndex: -1, itemTitles: titles)
        menu.present(at: CGPoint(x: 0, y: PVSpacing.controlHeightSmall))
    }

    private var menuTitles: [LocalizedStringResource] {
        if line.bridgeID == nil {
            return [
                L10n.CitationComposer.editObservationMenu,
                L10n.CitationComposer.duplicate,
                line.negative ? L10n.CitationComposer.affirmObservation : L10n.CitationComposer.negateObservation,
                L10n.CitationComposer.deleteObservation,
            ]
        }
        return [
            L10n.CitationComposer.editBridgeMenu,
            L10n.CitationComposer.duplicate,
            L10n.CitationComposer.showOnGraph,
            L10n.CitationComposer.deleteBridge,
        ]
    }

    @ViewBuilder
    private var menuItems: some View {
        PVContextMenuItem(menuTitles[0], index: 0) {
            model.beginEdit(line: line)
        }
        PVContextMenuItem(menuTitles[1], index: 1) {
            model.duplicate(line: line)
        }
        if line.bridgeID == nil {
            PVContextMenuItem(menuTitles[2], index: 2) {
                if let observationID = line.observationID {
                    Task { await model.negate(observationID: observationID) }
                }
            }
            PVContextMenuItem(menuTitles[3], index: 3) {
                Task { await model.askDeleteLine(line) }
            }
        } else {
            PVContextMenuItem(menuTitles[2], index: 2) {
                if let bridgeID = line.bridgeID {
                    model.showOnGraph(bridgeID: bridgeID)
                }
            }
            PVContextMenuItem(menuTitles[3], index: 3) {
                Task { await model.askDeleteLine(line) }
            }
        }
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
            HStack(alignment: .top, spacing: PVSpacing.space5) {
                PVField(label: L10n.CitationComposer.subjectLabel) {
                    PVComboBox(
                        selection: editSubjectBinding,
                        options: model.subjectOptions,
                        size: .sm,
                        placeholder: L10n.CitationComposer.subjectPlaceholder,
                        emptyLabel: L10n.CitationComposer.subjectPlaceholder,
                        label: L10n.CitationComposer.subjectLabel,
                        accessibilityIdentifierPrefix: "citationComposer.edit.subject"
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
                        accessibilityIdentifierPrefix: "citationComposer.edit.property"
                    )
                }
                .frame(maxWidth: .infinity)
            }
            editValue(values)
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
    private func editValue(_ values: ObservationRowValues) -> some View {
        let property = model.vocabulary.property(id: values.propertyID)
        switch property?.valueType {
        case PropertyValueType.text.rawValue:
            PVInput(text: editTextBinding, size: .sm)
        case PropertyValueType.integer.rawValue:
            PVInput(text: editIntegerBinding, size: .sm, mono: true)
        case PropertyValueType.term.rawValue:
            HStack(spacing: PVSpacing.space3) {
                PVComboBox(
                    selection: editTermBinding,
                    options: model.vocabulary.termComboOptionsByPropertyID[values.propertyID] ?? [],
                    size: .sm,
                    placeholder: L10n.CitationComposer.termPlaceholder,
                    emptyLabel: L10n.CitationComposer.termEmpty,
                    label: L10n.CitationComposer.termLabel,
                    accessibilityIdentifierPrefix: "citationComposer.edit.term"
                )
                PVButton(L10n.CitationComposer.addCustomTerm, variant: .ghost, size: .sm) {
                    model.beginAddCustomTerm(propertyID: values.propertyID, applyToEdit: true)
                }
            }
        case PropertyValueType.name.rawValue:
            editStructuredValue(values, date: false)
        case PropertyValueType.date.rawValue:
            editStructuredValue(values, date: true)
        default:
            EmptyView()
        }
    }

    private func editStructuredValue(_ values: ObservationRowValues, date: Bool) -> some View {
        let property = model.vocabulary.property(id: values.propertyID)
        let summary = CitationObservationValue.summary(
            valueType: property?.valueType ?? "",
            fields: CitationObservationValue.Fields(
                valueText: values.valueText,
                valueIntegerText: values.valueIntegerText,
                valueTermID: values.valueTermID,
                valueSubjectID: values.valueSubjectID,
                dateDraft: values.dateDraft,
                nameDraft: values.nameDraft
            ),
            termLabel: nil
        )
        let empty = summary.isEmpty
        return HStack(spacing: PVSpacing.space5) {
            Text(verbatim: empty ? L10n.string(L10n.CitationComposer.notEnteredYet) : summary)
                .font(date ? PVFont.mono(size: PVTypeScale.caption) : PVFont.body(size: PVTypeScale.caption))
                .foregroundStyle(empty ? PVColor.textFaint : PVColor.textPrimary)
                .lineLimit(1)
                .frame(maxWidth: .infinity, alignment: .leading)
            PVButton(
                empty
                    ? (date ? L10n.CitationComposer.enterDate : L10n.CitationComposer.enterName)
                    : (date ? L10n.CitationComposer.editDateButton : L10n.CitationComposer.editName),
                variant: .secondary,
                size: .sm,
                icon: date ? .calendar : .signature
            ) {
                model.beginValueEditor(applyToEdit: true)
            }
        }
    }

    private func bridgeForm(endA: String, endB: String, error: String?) -> some View {
        let pair = CitationComposerAuthoring.resolvePair(
            endA: endA,
            endB: endB,
            termID: editBridgeTermBinding.wrappedValue,
            vocabulary: model.vocabulary
        )
        return VStack(alignment: .leading, spacing: PVSpacing.space4) {
            lockedEnd(id: endA, title: endTitle(pair, first: true))
            if case .ready(let ready) = pair {
                HStack(spacing: PVSpacing.space4) {
                    PVMark(PVMarkKey(catalogKey: ready.markKey), size: 16, decorative: true)
                    Text(verbatim: ready.kindLabel)
                        .font(PVFont.body(size: PVTypeScale.caption, weight: PVFontWeight.semibold))
                    Text(verbatim: ready.pairHint)
                        .font(PVFont.body(size: PVTypeScale.caption, italic: true))
                        .foregroundStyle(PVColor.textMuted)
                }
            }
            lockedEnd(id: endB, title: endTitle(pair, first: false))
            Text(L10n.CitationComposer.lockEndsNote)
                .font(PVFont.body(size: PVTypeScale.caption, italic: true))
                .foregroundStyle(PVColor.textMuted)
            if case .ready(let ready) = pair, let property = ready.termProperty {
                VStack(alignment: .leading, spacing: PVSpacing.space3) {
                Text(verbatim: ConnectionTermCopy.label(disambiguationKey: property.key))
                    .font(PVFont.body(size: PVTypeScale.caption, weight: PVFontWeight.medium))
                    .foregroundStyle(PVColor.textSecondary)
                HStack(spacing: PVSpacing.space4) {
                    PVComboBox(
                        selection: editBridgeTermBinding,
                        options: model.vocabulary.termComboOptionsByPropertyID[property.id] ?? [],
                        size: .sm,
                        placeholder: property.key == "role"
                            ? L10n.CitationComposer.searchRoles
                            : L10n.CitationComposer.searchTypes,
                        emptyLabel: L10n.CitationComposer.termEmpty,
                        label: ConnectionTermCopy.label(disambiguationKey: property.key),
                        accessibilityIdentifierPrefix: "citationComposer.edit.bridgeTerm"
                    )
                    PVButton(L10n.CitationComposer.addCustomTerm, variant: .ghost, size: .sm, icon: .plus) {
                        model.beginAddCustomTerm(propertyID: property.id, applyToEdit: true)
                    }
                }
                if let reading = ready.reading {
                    HStack(alignment: .firstTextBaseline, spacing: PVSpacing.space4) {
                        Text(L10n.CitationComposer.readsEyebrow)
                            .pvMicroCaps()
                            .foregroundStyle(PVColor.textMuted)
                        Text(verbatim: reading)
                            .font(PVFont.body(size: PVTypeScale.caption))
                    }
                    .padding(.horizontal, PVSpacing.space5)
                    .padding(.vertical, PVSpacing.space3)
                    .frame(maxWidth: .infinity, alignment: .leading)
                    .background(PVColor.surfaceSunken)
                    .clipShape(RoundedRectangle(cornerRadius: PVRadius.sm, style: .continuous))
                }
                }
            }
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

    private func endTitle(_ pair: CitationPair, first: Bool) -> String {
        if case .ready(let ready) = pair {
            return first ? ready.endATitle : ready.endBTitle
        }
        return L10n.string(first ? L10n.CitationComposer.endA : L10n.CitationComposer.endB)
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
