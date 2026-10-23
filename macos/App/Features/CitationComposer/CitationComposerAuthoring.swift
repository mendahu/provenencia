import CoreGraphics
import Foundation

/// One authoring box: a property observation or a bridge, never both.
enum CitationAuthoring: Equatable {
    case rest
    case property(CitationPropertyDraft)
    case bridge(CitationBridgeDraft)

    var holdsLeave: Bool {
        switch self {
        case .rest: false
        case .property(let draft): draft.holdsLeave
        case .bridge(let draft): draft.holdsLeave
        }
    }

    var isSaving: Bool {
        switch self {
        case .rest: false
        case .property(let draft): draft.isSaving
        case .bridge(let draft): draft.isSaving
        }
    }
}

enum CitationAuthoringFocus: Hashable {
    case subject, property, value, endA, endB, term
}

/// The draft after a combo has written its choice. Field order reads this,
/// not a guess about what was on screen before the keystroke.
struct CitationAuthoringDraftAfterCommit: Equatable {
    var bridgeHasTerm = false
    /// False when the value area has no control to land on (a depiction's
    /// Draw button while it is disabled), so Tab goes on to the footer.
    var propertyValueTakesFocus = true
}

/// Field order for the observation dialogs. The combo reports the commit;
/// a forward Tab always moves to `next` when there is one.
enum CitationAuthoringNavigation {
    /// A real choice. The "create a subject" row opens another dialog instead.
    static func shouldAdvance(_ committed: String) -> Bool {
        let trimmed = committed.trimmingCharacters(in: .whitespacesAndNewlines)
        return !trimmed.isEmpty && !trimmed.hasPrefix(CitationComposerModel.newSubjectPrefix)
    }

    /// Nil is the last stop: the value is filled in place, or the bridge has no term.
    /// `draftAfterCommit` is the draft once this choice is written.
    static func next(
        after field: CitationAuthoringFocus,
        draftAfterCommit: CitationAuthoringDraftAfterCommit = CitationAuthoringDraftAfterCommit()
    ) -> CitationAuthoringFocus? {
        switch field {
        case .subject: .property
        case .property: draftAfterCommit.propertyValueTakesFocus ? .value : nil
        case .endA: .endB
        case .endB: draftAfterCommit.bridgeHasTerm ? .term : nil
        case .value, .term: nil
        }
    }

    /// Where a forward Tab goes after the choice is written. Click and Return
    /// stay on the field that committed. Nil lets Tab continue to the footer.
    static func tabDestination(
        from field: CitationAuthoringFocus,
        committed: String,
        via: PVComboCommitVia,
        draftAfterCommit: CitationAuthoringDraftAfterCommit
    ) -> CitationAuthoringFocus? {
        guard via == .tab, shouldAdvance(committed) else { return nil }
        return next(after: field, draftAfterCommit: draftAfterCommit)
    }

    /// A run of likenesses is one face per person, so a depiction opens on the person.
    static func initialPropertyFocus(
        subjectID: String,
        propertyID: String = "",
        isDepiction: Bool = false
    ) -> CitationAuthoringFocus {
        if subjectID.isEmpty { return .subject }
        if propertyID.isEmpty { return .property }
        return isDepiction ? .subject : .property
    }

    /// Property focus for a draft's values; the one place the depiction rule lives.
    static func initialPropertyFocus(
        values: ObservationRowValues,
        vocabulary: CitationComposerVocabulary
    ) -> CitationAuthoringFocus {
        initialPropertyFocus(
            subjectID: values.subjectID,
            propertyID: values.propertyID,
            isDepiction: vocabulary.isDepiction(propertyID: values.propertyID)
        )
    }

    static func initialBridgeFocus(endA: String, endB: String, hasTermField: Bool, termID: String) -> CitationAuthoringFocus {
        if hasTermField, termID.isEmpty { return .term }
        if endA.isEmpty { return .endA }
        if endB.isEmpty { return .endB }
        return hasTermField ? .term : .endB
    }
}

enum CitationBridgeOrigin: Equatable {
    case blank, graph, duplicate
}

enum CitationSubjectField: Equatable {
    case propertySubject, bridgeEndA, bridgeEndB
}

struct CitationPropertyDraft: Equatable {
    var values = ObservationRowValues()
    var baseline = ObservationRowValues()
    var isSaving = false
    var error: String?
    var note: String?
    var isDuplicate = false

    var holdsLeave: Bool { isDuplicate || values != baseline }
}

struct CitationBridgeDraft: Equatable {
    var endA = ""
    var endB = ""
    var termID = ""
    var baselineEndA = ""
    var baselineEndB = ""
    var baselineTermID = ""
    var isSaving = false
    var error: String?
    var note: String?
    var origin: CitationBridgeOrigin = .blank

    var holdsLeave: Bool {
        origin != .blank
            || endA != baselineEndA
            || endB != baselineEndB
            || termID != baselineTermID
    }
}

struct CitationEditSession: Equatable {
    enum Kind: Equatable {
        case property(values: ObservationRowValues, persistedID: String, persistedRef: String)
        case bridge(
            endA: String,
            endB: String,
            termID: String,
            bridgeID: String,
            bridgeRef: String,
            baselineTermID: String
        )
    }

    var kind: Kind
    var isSaving = false
    var error: String?

    var propertyBaseline = ObservationRowValues()
    /// Subject whose line opened the sheet, so the subtitle can name that end.
    var openedFromSubjectID = ""

    var isDirty: Bool {
        switch kind {
        case .property(let values, _, _):
            return values != propertyBaseline
        case .bridge(_, _, let termID, _, _, let baselineTermID):
            return termID != baselineTermID
        }
    }
}

enum CitationPair: Equatable {
    case incomplete(String)
    case refused(title: String, message: String)
    case ready(CitationReadyBridge)
}

struct CitationReadyBridge: Equatable {
    var rule: CatalogConnectRule
    var termProperty: CatalogProperty?
    var endATitle: String
    var endBTitle: String
    var canSwap: Bool
    var reading: String?
    var kindLabel: String
    var pairHint: String
    var markKey: PVMarkKey
}

enum CitationPreview: Equatable {
    case text(String)
    case empty
}

struct CitationSavedLine: Identifiable, Equatable {
    var id: String
    var sortKey: String
    var subjectID: String
    var observationID: String?
    var bridgeID: String?
    /// Property name, or the short phrase read from this subject ("Parent of", "Son in", "At").
    var label: String
    /// The value, or the other subject's name.
    var valueText: String
    /// Whole numbers and dates use the mono face.
    var mono: Bool
    var body: String
    var negative: Bool
    var markKey: PVMarkKey?
    /// `objects/…` path of a saved depiction's image; nil for every other line.
    var imageRelPath: String? = nil

    /// Label and value as one phrase, the way a duplicate note names the copied line.
    var phrase: String {
        [label, valueText]
            .map { $0.trimmingCharacters(in: .whitespacesAndNewlines) }
            .filter { !$0.isEmpty }
            .joined(separator: " ")
    }
}

struct CitationSubjectGroup: Identifiable, Equatable {
    var subjectID: String
    var name: String
    var ref: String
    var typeKey: String
    var lines: [CitationSavedLine]
    var id: String { subjectID }
}

enum CitationComposerAuthoring {
    static func preview(transcription: String, description: String) -> CitationPreview {
        let transcription = collapsed(transcription)
        if !transcription.isEmpty { return .text(transcription) }
        let description = collapsed(description)
        if !description.isEmpty { return .text(description) }
        return .empty
    }

    static func stepperOrder(_ listed: [CatalogListedCitation]) -> [CatalogListedCitation] {
        listed.sorted { lhs, rhs in
            let a = position(of: lhs.citation)
            let b = position(of: rhs.citation)
            if a.page != b.page { return a.page < b.page }
            if a.top != b.top { return a.top < b.top }
            return a.ref.localizedStandardCompare(b.ref) == .orderedAscending
        }
    }

    static func resolvePair(
        endA: String,
        endB: String,
        termID: String,
        vocabulary: CitationComposerVocabulary
    ) -> CitationPair {
        if endA.isEmpty || endB.isEmpty {
            let hint = endA.isEmpty && endB.isEmpty
                ? L10n.string(L10n.CitationComposer.pairPickBoth)
                : L10n.string(L10n.CitationComposer.pairPickSecond)
            return .incomplete(hint)
        }
        if endA == endB {
            return .refused(
                title: L10n.string(L10n.CitationComposer.sameSubjectTitle),
                message: L10n.string(L10n.CitationComposer.sameSubjectBody)
            )
        }
        guard let subjectA = vocabulary.subjectsByID[endA],
              let subjectB = vocabulary.subjectsByID[endB]
        else {
            return .incomplete(L10n.string(L10n.CitationComposer.pairPickBoth))
        }
        guard let rule = vocabulary.connectRules.first(where: {
            $0.fromTypeKey == subjectA.typeKey && $0.toTypeKey == subjectB.typeKey
        }) else {
            return .refused(
                title: L10n.string(L10n.CitationComposer.refusedPairTitle),
                message: L10n.string(L10n.CitationComposer.refusedPairBody)
            )
        }
        if rule.refuse {
            return refusedCopy(rule: rule, subjectA: subjectA, subjectB: subjectB)
        }
        let termProperty = termProperty(for: rule, vocabulary: vocabulary)
        return .ready(
            readyBridge(
                subjectA: subjectA,
                subjectB: subjectB,
                rule: rule,
                termID: termID,
                termProperty: termProperty,
                vocabulary: vocabulary
            )
        )
    }

    static func groups(
        observations: [ObservationRow],
        connections: [ConnectionRow],
        vocabulary: CitationComposerVocabulary,
        nameOrders: NameFormatOrders = NameFormatOrders()
    ) -> [CitationSubjectGroup] {
        var lines: [CitationSavedLine] = []
        for row in observations where row.persistedID != nil {
            guard let persistedID = row.persistedID, !row.subjectID.isEmpty else { continue }
            let property = vocabulary.property(id: row.propertyID)
            let termLabel = vocabulary.termLabel(termID: row.valueTermID, propertyID: row.propertyID)
            let valueText = CitationObservationValue.summary(
                valueType: property?.valueType ?? "",
                fields: row.valueFields(),
                termLabel: termLabel,
                orders: nameOrders
            )
            let label = property?.label ?? ""
            let mono = property?.valueType == PropertyValueType.integer.rawValue
                || property?.valueType == PropertyValueType.date.rawValue
            lines.append(
                CitationSavedLine(
                    id: "obs:\(persistedID)",
                    sortKey: persistedID,
                    subjectID: row.subjectID,
                    observationID: persistedID,
                    bridgeID: nil,
                    label: label,
                    valueText: valueText,
                    mono: mono,
                    body: [label, valueText].filter { !$0.isEmpty }.joined(separator: " "),
                    negative: row.isNegated,
                    markKey: nil,
                    imageRelPath: row.depictionDraft.imageRelPath.isEmpty
                        ? nil
                        : row.depictionDraft.imageRelPath
                )
            )
        }
        for connection in connections where !connection.isPending {
            guard let bridgeID = connection.bridgeID, !bridgeID.isEmpty else { continue }
            let rule = vocabulary.connectRule(bridgeTypeKey: connection.bridgeTypeKey)
            let termProperty = connection.termProperty
            for subjectID in [connection.fromSubjectID, connection.toSubjectID] where !subjectID.isEmpty {
                guard let viewer = vocabulary.subjectsByID[subjectID] else { continue }
                let otherID = subjectID == connection.fromSubjectID
                    ? connection.toSubjectID
                    : connection.fromSubjectID
                guard let other = vocabulary.subjectsByID[otherID] else { continue }
                let from = vocabulary.subjectsByID[connection.fromSubjectID]
                let to = vocabulary.subjectsByID[connection.toSubjectID]
                let phrase: (label: String, value: String)
                if let rule, let from, let to {
                    phrase = listedLine(
                        from: from,
                        to: to,
                        rule: rule,
                        termID: connection.roleTermID,
                        termProperty: termProperty,
                        vocabulary: vocabulary,
                        viewing: viewer.id
                    )
                } else {
                    phrase = (L10n.string(L10n.CitationComposer.bridgeKindRelationship), other.label)
                }
                lines.append(
                    CitationSavedLine(
                        id: "bridge:\(bridgeID):\(subjectID)",
                        sortKey: bridgeID,
                        subjectID: subjectID,
                        observationID: nil,
                        bridgeID: bridgeID,
                        label: phrase.label,
                        valueText: phrase.value,
                        mono: false,
                        body: [phrase.label, phrase.value].filter { !$0.isEmpty }.joined(separator: " "),
                        negative: false,
                        markKey: PVMarkKey(subjectTypeKey: connection.bridgeTypeKey) ?? .fallback
                    )
                )
            }
        }
        let subjectIDs = Set(lines.map(\.subjectID))
        var groups: [CitationSubjectGroup] = []
        for subjectID in subjectIDs {
            guard let subject = vocabulary.subjectsByID[subjectID] else { continue }
            let subjectLines = lines
                .filter { $0.subjectID == subjectID }
                .sorted { lhs, rhs in
                    if lhs.sortKey != rhs.sortKey { return lhs.sortKey < rhs.sortKey }
                    return lhs.id < rhs.id
                }
            groups.append(
                CitationSubjectGroup(
                    subjectID: subjectID,
                    name: subject.label,
                    ref: subject.ref,
                    typeKey: subject.typeKey,
                    lines: subjectLines
                )
            )
        }
        return groups.sorted { lhs, rhs in
            let rankA = typeRank(lhs.typeKey)
            let rankB = typeRank(rhs.typeKey)
            if rankA != rankB { return rankA < rankB }
            let firstA = lhs.lines.map(\.sortKey).min() ?? ""
            let firstB = rhs.lines.map(\.sortKey).min() ?? ""
            if firstA != firstB { return firstA < firstB }
            return lhs.name.localizedStandardCompare(rhs.name) == .orderedAscending
        }
    }

    private static func typeRank(_ key: String) -> Int {
        EvidencePrimaryKind(rawValue: key)
            .flatMap { EvidencePrimaryKind.allCases.firstIndex(of: $0) }
            ?? EvidencePrimaryKind.allCases.count
    }

    private static func position(of citation: CatalogCitation) -> (page: Int, top: Double, ref: String) {
        let locator = CitationLocatorDraft.decode(citation.locatorJSON)
        let page = locator.page ?? 0
        let top = locator.region?.points.map(\.y).min().map(Double.init) ?? -1
        return (page, top, citation.ref)
    }

    private static func collapsed(_ text: String) -> String {
        text
            .components(separatedBy: .newlines)
            .map { $0.trimmingCharacters(in: .whitespaces) }
            .filter { !$0.isEmpty }
            .joined(separator: " — ")
    }

    private static func termProperty(
        for rule: CatalogConnectRule,
        vocabulary: CitationComposerVocabulary
    ) -> CatalogProperty? {
        guard rule.disambiguation != "none", !rule.disambiguation.isEmpty else { return nil }
        return vocabulary.property(key: rule.disambiguation)
    }

    private static func readyBridge(
        subjectA: CitationComposerModel.GraphSubjectOption,
        subjectB: CitationComposerModel.GraphSubjectOption,
        rule: CatalogConnectRule,
        termID: String,
        termProperty: CatalogProperty?,
        vocabulary: CitationComposerVocabulary
    ) -> CitationReadyBridge {
        let term = storedTerm(termID: termID, property: termProperty, vocabulary: vocabulary)
        let ends = endTitles(
            subjectA: subjectA,
            subjectB: subjectB,
            rule: rule,
            term: term,
            vocabulary: vocabulary,
            property: termProperty
        )
        let chrome = kindChrome(rule)
        return CitationReadyBridge(
            rule: rule,
            termProperty: termProperty,
            endATitle: ends.a,
            endBTitle: ends.b,
            canSwap: ends.canSwap,
            reading: readingSentence(
                from: subjectA,
                to: subjectB,
                rule: rule,
                term: term
            ),
            kindLabel: chrome.label,
            pairHint: chrome.hint,
            markKey: PVMarkKey(subjectTypeKey: rule.bridgeTypeKey) ?? .fallback
        )
    }

    private static func kindChrome(_ rule: CatalogConnectRule) -> (label: String, hint: String) {
        switch rule.bridgeTypeKey {
        case EvidenceBridgeKind.relationship.rawValue:
            return (
                L10n.string(L10n.CitationComposer.bridgeKindRelationship),
                L10n.string(L10n.CitationComposer.pairPersonPerson)
            )
        case EvidenceBridgeKind.participation.rawValue:
            return (
                L10n.string(L10n.CitationComposer.bridgeKindParticipation),
                L10n.string(L10n.CitationComposer.pairPersonEvent)
            )
        case EvidenceBridgeKind.location.rawValue:
            return (
                L10n.string(L10n.CitationComposer.bridgeKindLocation),
                L10n.string(L10n.CitationComposer.pairEventPlace)
            )
        case EvidenceBridgeKind.placeRelationship.rawValue:
            return (
                L10n.string(L10n.CitationComposer.bridgeKindPlace),
                L10n.string(L10n.CitationComposer.pairPlacePlace)
            )
        default:
            return (
                L10n.string(L10n.CitationComposer.bridgeKindOther),
                ""
            )
        }
    }

    private static func endTitles(
        subjectA: CitationComposerModel.GraphSubjectOption,
        subjectB: CitationComposerModel.GraphSubjectOption,
        rule: CatalogConnectRule,
        term: CatalogPropertyTerm?,
        vocabulary: CitationComposerVocabulary,
        property: CatalogProperty?
    ) -> (a: String, b: String, canSwap: Bool) {
        switch rule.bridgeTypeKey {
        case EvidenceBridgeKind.relationship.rawValue:
            if let term {
                let inverse = inverseTerm(of: term, property: property, vocabulary: vocabulary)
                return (term.label, inverse?.label ?? term.label, true)
            }
            return (
                L10n.string(L10n.CitationComposer.firstPerson),
                L10n.string(L10n.CitationComposer.secondPerson),
                true
            )
        case EvidenceBridgeKind.placeRelationship.rawValue:
            switch term?.productKey {
            case SeededPlaceRelationshipTypeTerm.partOf:
                return (
                    L10n.string(L10n.CitationComposer.partEnd),
                    L10n.string(L10n.CitationComposer.wholeEnd),
                    true
                )
            case SeededPlaceRelationshipTypeTerm.succeededBy:
                return (
                    L10n.string(L10n.CitationComposer.earlierPlace),
                    L10n.string(L10n.CitationComposer.laterPlace),
                    true
                )
            default:
                if let term { return (term.label, term.label, true) }
                return (
                    L10n.string(L10n.CitationComposer.firstPlace),
                    L10n.string(L10n.CitationComposer.secondPlace),
                    true
                )
            }
        case EvidenceBridgeKind.participation.rawValue:
            if subjectA.typeKey == "person" {
                return (
                    L10n.string(L10n.CitationComposer.participantEnd),
                    L10n.string(L10n.CitationComposer.eventEnd),
                    false
                )
            }
            return (
                L10n.string(L10n.CitationComposer.eventEnd),
                L10n.string(L10n.CitationComposer.participantEnd),
                false
            )
        case EvidenceBridgeKind.location.rawValue:
            if subjectA.typeKey == "event" {
                return (
                    L10n.string(L10n.CitationComposer.eventEnd),
                    L10n.string(L10n.CitationComposer.placeEnd),
                    false
                )
            }
            return (
                L10n.string(L10n.CitationComposer.placeEnd),
                L10n.string(L10n.CitationComposer.eventEnd),
                false
            )
        default:
            _ = subjectB
            return (
                L10n.string(L10n.CitationComposer.endA),
                L10n.string(L10n.CitationComposer.endB),
                false
            )
        }
    }

    private static func listedLine(
        from: CitationComposerModel.GraphSubjectOption,
        to: CitationComposerModel.GraphSubjectOption,
        rule: CatalogConnectRule,
        termID: String,
        termProperty: CatalogProperty?,
        vocabulary: CitationComposerVocabulary,
        viewing: String
    ) -> (label: String, value: String) {
        let other = viewing == from.id ? to : from
        let term = storedTerm(termID: termID, property: termProperty, vocabulary: vocabulary)
        switch rule.bridgeTypeKey {
        case EvidenceBridgeKind.relationship.rawValue:
            let spoken = viewing == to.id
                ? inverseTerm(of: term, property: termProperty, vocabulary: vocabulary) ?? term
                : term
            let label = spoken.map { L10n.CitationComposer.termOf($0.label) }
                ?? L10n.string(L10n.CitationComposer.bridgeKindRelationship)
            return (label, other.label)
        case EvidenceBridgeKind.participation.rawValue:
            let person = from.typeKey == "person" ? from : to
            let event = from.typeKey == "event" ? from : to
            if viewing == person.id {
                let label = term.map { L10n.CitationComposer.roleIn($0.label) }
                    ?? L10n.string(L10n.CitationComposer.takesPart)
                return (label, event.label)
            }
            return (term?.label ?? L10n.string(L10n.CitationComposer.takesPart), person.label)
        case EvidenceBridgeKind.location.rawValue:
            let event = from.typeKey == "event" ? from : to
            let place = from.typeKey == "place" ? from : to
            if viewing == event.id {
                return (L10n.string(L10n.CitationComposer.atWord), place.label)
            }
            return (L10n.string(L10n.CitationComposer.siteOfWord), event.label)
        case EvidenceBridgeKind.placeRelationship.rawValue:
            let label = term?.label ?? L10n.string(L10n.CitationComposer.bridgeKindPlace)
            return (label, other.label)
        default:
            return ("", other.label)
        }
    }

    private static func readingSentence(
        from: CitationComposerModel.GraphSubjectOption,
        to: CitationComposerModel.GraphSubjectOption,
        rule: CatalogConnectRule,
        term: CatalogPropertyTerm?
    ) -> String? {
        switch rule.bridgeTypeKey {
        case EvidenceBridgeKind.relationship.rawValue:
            guard let term else { return nil }
            if term.productKey == "spouse" {
                return L10n.CitationComposer.areSpouses(from.label, to.label)
            }
            if term.productKey == "sibling" {
                return L10n.CitationComposer.areSiblings(from.label, to.label)
            }
            return L10n.CitationComposer.isATermOf(from.label, term.label.lowercased(), to.label)
        case EvidenceBridgeKind.participation.rawValue:
            guard let term else { return nil }
            let person = from.typeKey == "person" ? from : to
            let event = from.typeKey == "event" ? from : to
            return L10n.CitationComposer.roleInTheEvent(person.label, term.label, event.label)
        case EvidenceBridgeKind.location.rawValue:
            let event = from.typeKey == "event" ? from : to
            let place = from.typeKey == "place" ? from : to
            return L10n.CitationComposer.tookPlaceAt(event.label, place.label)
        case EvidenceBridgeKind.placeRelationship.rawValue:
            guard let term else { return nil }
            if term.productKey == "part_of" {
                return L10n.CitationComposer.isPartOf(from.label, to.label)
            }
            if term.productKey == "succeeded_by" {
                return L10n.CitationComposer.wasSucceededBy(from.label, to.label)
            }
            return nil
        default:
            return nil
        }
    }

    private static func refusedCopy(
        rule: CatalogConnectRule,
        subjectA: CitationComposerModel.GraphSubjectOption,
        subjectB: CitationComposerModel.GraphSubjectOption
    ) -> CitationPair {
        if (subjectA.typeKey == "person" && subjectB.typeKey == "place")
            || (subjectA.typeKey == "place" && subjectB.typeKey == "person")
        {
            let person = subjectA.typeKey == "person" ? subjectA.label : subjectB.label
            let place = subjectA.typeKey == "place" ? subjectA.label : subjectB.label
            return .refused(
                title: L10n.string(L10n.CitationComposer.personPlaceTitle),
                message: L10n.CitationComposer.personPlaceBody(person: person, place: place)
            )
        }
        if subjectA.typeKey == "event", subjectB.typeKey == "event" {
            return .refused(
                title: L10n.string(L10n.CitationComposer.twoEventsTitle),
                message: L10n.string(L10n.CitationComposer.twoEventsBody)
            )
        }
        _ = rule
        return .refused(
            title: L10n.string(L10n.CitationComposer.refusedPairTitle),
            message: L10n.string(L10n.CitationComposer.refusedPairBody)
        )
    }

    private static func storedTerm(
        termID: String,
        property: CatalogProperty?,
        vocabulary: CitationComposerVocabulary
    ) -> CatalogPropertyTerm? {
        guard let property, !termID.isEmpty else { return nil }
        return vocabulary.termsByPropertyID[property.id]?.first { $0.id == termID }
    }

    private static func inverseTerm(
        of term: CatalogPropertyTerm?,
        property: CatalogProperty?,
        vocabulary: CitationComposerVocabulary
    ) -> CatalogPropertyTerm? {
        guard let term, let property, let inverseKey = term.productKey.flatMap({ SeededRelationshipTypeTerm.inverse[$0] }) else { return nil }
        return vocabulary.termsByPropertyID[property.id]?.first { $0.productKey == inverseKey }
    }
}
