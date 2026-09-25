import Foundation

/// Snapshot of catalog vocabulary the composer can read without hitting the store.
struct CitationComposerVocabulary: Equatable {
    static let supportedValueTypes: Set<String> = [
        PropertyValueType.text.rawValue,
        PropertyValueType.integer.rawValue,
        PropertyValueType.date.rawValue,
        PropertyValueType.term.rawValue,
        PropertyValueType.name.rawValue,
    ]

    var propertiesByID: [String: CatalogProperty]
    var propertiesByTypeID: [String: [CatalogProperty]]
    var graphSubjects: [CitationComposerModel.GraphSubjectOption]
    var termsByPropertyID: [String: [CatalogPropertyTerm]]
    var connectRules: [CatalogConnectRule]
    var typesByID: [String: CatalogSubjectType]

    static let empty = CitationComposerVocabulary(
        propertiesByID: [:],
        propertiesByTypeID: [:],
        graphSubjects: [],
        termsByPropertyID: [:],
        connectRules: [],
        typesByID: [:]
    )

    static func make(
        fields: SubjectFieldsSnapshot,
        snapshot: SourceGraphSnapshot,
        rules: [CatalogConnectRule],
        termsByPropertyID: [String: [CatalogPropertyTerm]]
    ) -> CitationComposerVocabulary {
        var propertiesByID = Dictionary(uniqueKeysWithValues: fields.properties.map { ($0.id, $0) })
        var propertiesByTypeID: [String: [CatalogProperty]] = [:]
        for (typeID, typeFields) in fields.fieldsByTypeID {
            propertiesByTypeID[typeID] = typeFields.map(\.property)
            for field in typeFields {
                propertiesByID[field.property.id] = field.property
            }
        }
        let typesByID = Dictionary(uniqueKeysWithValues: fields.types.map { ($0.id, $0) })
        var graphSubjects: [CitationComposerModel.GraphSubjectOption] = []
        for placed in snapshot.subjects {
            graphSubjects.append(
                CitationComposerModel.GraphSubjectOption(
                    id: placed.id,
                    label: placed.subject.label,
                    ref: placed.subject.ref,
                    typeID: placed.subject.subjectTypeID,
                    typeKey: placed.kind.rawValue
                )
            )
        }
        for bridge in snapshot.bridges {
            let sentence = EvidenceBridgeEdgeSummary.sentence(for: bridge, in: snapshot)
            graphSubjects.append(
                CitationComposerModel.GraphSubjectOption(
                    id: bridge.id,
                    label: sentence,
                    ref: bridge.subject.ref,
                    typeID: bridge.subject.subjectTypeID,
                    typeKey: bridge.kind.rawValue
                )
            )
        }
        return CitationComposerVocabulary(
            propertiesByID: propertiesByID,
            propertiesByTypeID: propertiesByTypeID,
            graphSubjects: graphSubjects,
            termsByPropertyID: termsByPropertyID,
            connectRules: rules,
            typesByID: typesByID
        )
    }

    func isEdge(subjectID: String, propertyID: String) -> Bool {
        guard let subject = graphSubjects.first(where: { $0.id == subjectID }) else { return false }
        let propertyKey = propertiesByID[propertyID]?.key ?? ""
        return connectRules.contains { rule in
            !rule.refuse
                && rule.bridgeTypeKey == subject.typeKey
                && rule.edges.contains { $0.propertyKey == propertyKey }
        }
    }

    func propertyOptions(forSubjectID subjectID: String) -> [CatalogProperty] {
        guard let subject = graphSubjects.first(where: { $0.id == subjectID }) else { return [] }
        let excluded = Set(
            connectRules
                .filter { $0.bridgeTypeKey == subject.typeKey && !$0.refuse }
                .flatMap { $0.edges.map(\.propertyKey) }
        )
        return (propertiesByTypeID[subject.typeID] ?? [])
            .filter {
                CitationComposerVocabulary.supportedValueTypes.contains($0.valueType)
                    && !excluded.contains($0.key)
            }
            .sorted { $0.label.localizedCaseInsensitiveCompare($1.label) == .orderedAscending }
    }

    func connectRule(bridgeTypeKey: String) -> CatalogConnectRule? {
        connectRules.first { $0.bridgeTypeKey == bridgeTypeKey && !$0.refuse }
    }

    func property(id: String) -> CatalogProperty? {
        propertiesByID[id]
    }

    func property(key: String) -> CatalogProperty? {
        propertiesByID.values.first { $0.key == key }
    }

    func termLabel(termID: String, propertyID: String) -> String? {
        guard let term = termsByPropertyID[propertyID]?.first(where: { $0.id == termID }) else {
            return nil
        }
        let propertyKey = propertiesByID[propertyID]?.key ?? ""
        return PropertyTermDisplay.name(term: term, propertyKey: propertyKey)
    }
}
