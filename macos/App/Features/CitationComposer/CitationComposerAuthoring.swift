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

enum CitationAuthoringFocus: Equatable {
    case subject, property, value, endA, endB, term
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
    var focus: CitationAuthoringFocus = .subject
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
    var focus: CitationAuthoringFocus = .endA
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
    var body: String
    var negative: Bool
    var markKey: String?
}

struct CitationSubjectGroup: Identifiable, Equatable {
    var subjectID: String
    var name: String
    var typeKey: String
    var lines: [CitationSavedLine]
    var id: String { subjectID }
}

enum CitationComposerAuthoring {
    static let inverseTermKey: [String: String] = [
        "parent": "child",
        "child": "parent",
        "grandparent": "grandchild",
        "grandchild": "grandparent",
        "pibling": "nibling",
        "nibling": "pibling",
        "guardian": "ward",
        "ward": "guardian",
    ]

    static func preview(transcription: String, description: String) -> CitationPreview {
        let transcription = collapsed(transcription)
        if !transcription.isEmpty { return .text(transcription) }
        let description = collapsed(description)
        if !description.isEmpty { return .text(description) }
        return .empty
    }

    static func subjectNameSummary(_ names: [String]) -> String? {
        let names = names.map { $0.trimmingCharacters(in: .whitespacesAndNewlines) }.filter { !$0.isEmpty }
        switch names.count {
        case 0: return nil
        case 1: return names[0]
        case 2:
            return L10n.CitationComposer.joinedNames(names[0], names[1])
        default:
            let head = names.dropLast().joined(separator: ", ")
            return L10n.CitationComposer.joinedNames(head, names[names.count - 1])
        }
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
        let reading = bridgeReading(
            from: subjectA,
            to: subjectB,
            rule: rule,
            termID: termID,
            termProperty: termProperty,
            vocabulary: vocabulary,
            viewing: subjectA.id
        )
        return .ready(
            CitationReadyBridge(
                rule: rule,
                termProperty: termProperty,
                endATitle: endTitle(typeKey: subjectA.typeKey, first: true, peer: subjectB.typeKey),
                endBTitle: endTitle(typeKey: subjectB.typeKey, first: false, peer: subjectA.typeKey),
                canSwap: true,
                reading: reading
            )
        )
    }

    static func groups(
        observations: [ObservationRow],
        connections: [ConnectionRow],
        vocabulary: CitationComposerVocabulary
    ) -> [CitationSubjectGroup] {
        var lines: [CitationSavedLine] = []
        for row in observations where row.persistedID != nil {
            guard let persistedID = row.persistedID, !row.subjectID.isEmpty else { continue }
            let property = vocabulary.property(id: row.propertyID)
            let termLabel = vocabulary.termLabel(termID: row.valueTermID, propertyID: row.propertyID)
            let body = CitationObservationValue.summary(
                valueType: property?.valueType ?? "",
                fields: row.valueFields(),
                termLabel: termLabel
            )
            lines.append(
                CitationSavedLine(
                    id: "obs:\(persistedID)",
                    sortKey: persistedID,
                    subjectID: row.subjectID,
                    observationID: persistedID,
                    bridgeID: nil,
                    body: body,
                    negative: row.isNegated,
                    markKey: nil
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
                let body: String
                if let rule, let from, let to {
                    body = bridgeReading(
                        from: from,
                        to: to,
                        rule: rule,
                        termID: connection.roleTermID,
                        termProperty: termProperty,
                        vocabulary: vocabulary,
                        viewing: viewer.id
                    ) ?? other.label
                } else {
                    body = other.label
                }
                lines.append(
                    CitationSavedLine(
                        id: "bridge:\(bridgeID):\(subjectID)",
                        sortKey: bridgeID,
                        subjectID: subjectID,
                        observationID: nil,
                        bridgeID: bridgeID,
                        body: body,
                        negative: false,
                        markKey: "subject_\(connection.bridgeTypeKey)"
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
        switch key {
        case "person": 0
        case "event": 1
        case "place": 2
        default: 3
        }
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

    private static func endTitle(typeKey: String, first: Bool, peer: String) -> String {
        if typeKey == peer {
            switch typeKey {
            case "person":
                return L10n.string(first ? L10n.CitationComposer.firstPerson : L10n.CitationComposer.secondPerson)
            case "place":
                return L10n.string(first ? L10n.CitationComposer.firstPlace : L10n.CitationComposer.secondPlace)
            default:
                break
            }
        }
        switch typeKey {
        case "person": return L10n.string(L10n.CitationComposer.participantEnd)
        case "event": return L10n.string(L10n.CitationComposer.eventEnd)
        case "place": return L10n.string(L10n.CitationComposer.placeEnd)
        default: return L10n.string(first ? L10n.CitationComposer.endA : L10n.CitationComposer.endB)
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

    private static func bridgeReading(
        from: CitationComposerModel.GraphSubjectOption,
        to: CitationComposerModel.GraphSubjectOption,
        rule: CatalogConnectRule,
        termID: String,
        termProperty: CatalogProperty?,
        vocabulary: CitationComposerVocabulary,
        viewing: String
    ) -> String? {
        let other = viewing == from.id ? to : from
        let term = termLabel(termID: termID, property: termProperty, vocabulary: vocabulary)
        switch rule.bridgeTypeKey {
        case EvidenceBridgeKind.relationship.rawValue:
            let spoken = viewing == to.id
                ? inverseLabel(termID: termID, property: termProperty, vocabulary: vocabulary) ?? term
                : term
            if let spoken, !spoken.isEmpty {
                return L10n.CitationComposer.termOfName(term: spoken, name: other.label)
            }
            return L10n.CitationComposer.relationshipWith(name: other.label)
        case EvidenceBridgeKind.participation.rawValue:
            let person = from.typeKey == "person" ? from : to
            let event = from.typeKey == "event" ? from : to
            if viewing == person.id {
                if let term, !term.isEmpty {
                    return L10n.CitationComposer.roleInEvent(role: term, event: event.label)
                }
                return L10n.CitationComposer.takesPartIn(event: event.label)
            }
            if let term, !term.isEmpty {
                return L10n.CitationComposer.roleOnPerson(role: term, person: person.label)
            }
            return L10n.CitationComposer.personTakesPart(person: person.label)
        case EvidenceBridgeKind.location.rawValue:
            let event = from.typeKey == "event" ? from : to
            let place = from.typeKey == "place" ? from : to
            if viewing == event.id {
                return L10n.CitationComposer.atPlace(place: place.label)
            }
            return L10n.CitationComposer.siteOf(event: event.label)
        case EvidenceBridgeKind.placeRelationship.rawValue:
            if let term, !term.isEmpty {
                return L10n.CitationComposer.termOfName(term: term, name: other.label)
            }
            return L10n.CitationComposer.placeRelationshipWith(name: other.label)
        default:
            return other.label
        }
    }

    private static func termLabel(
        termID: String,
        property: CatalogProperty?,
        vocabulary: CitationComposerVocabulary
    ) -> String? {
        guard let property, !termID.isEmpty else { return nil }
        return vocabulary.termLabel(termID: termID, propertyID: property.id)
    }

    private static func inverseLabel(
        termID: String,
        property: CatalogProperty?,
        vocabulary: CitationComposerVocabulary
    ) -> String? {
        guard let property, !termID.isEmpty,
              let term = vocabulary.termsByPropertyID[property.id]?.first(where: { $0.id == termID }),
              let inverseKey = inverseTermKey[term.key],
              let inverse = vocabulary.termsByPropertyID[property.id]?.first(where: { $0.key == inverseKey })
        else { return nil }
        return vocabulary.termLabel(termID: inverse.id, propertyID: property.id)
    }
}
