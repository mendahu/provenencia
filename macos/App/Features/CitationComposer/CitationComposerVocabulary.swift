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
    var subjectsByID: [String: CitationComposerModel.GraphSubjectOption]
    var termsByPropertyID: [String: [CatalogPropertyTerm]]
    var connectRules: [CatalogConnectRule]
    var typesByID: [String: CatalogSubjectType]
    /// One subject menu for every observation row, including "new subject".
    var subjectComboOptions: [PVComboBoxOption]
    /// Property menus shared by every subject of a type.
    var propertyComboOptionsByTypeID: [String: [PVComboBoxOption]]
    /// Term menus shared by every observation of a property.
    var termComboOptionsByPropertyID: [String: [PVComboBoxOption]]

    static let empty = CitationComposerVocabulary(
        propertiesByID: [:],
        propertiesByTypeID: [:],
        graphSubjects: [],
        subjectsByID: [:],
        termsByPropertyID: [:],
        connectRules: [],
        typesByID: [:],
        subjectComboOptions: [],
        propertyComboOptionsByTypeID: [:],
        termComboOptionsByPropertyID: [:]
    )

    static func make(
        fields: PropertiesSnapshot,
        snapshot: SourceGraphSnapshot,
        rules: [CatalogConnectRule],
        termsByPropertyID: [String: [CatalogPropertyTerm]]
    ) -> CitationComposerVocabulary {
        var propertiesByID = Dictionary(uniqueKeysWithValues: fields.properties.map { ($0.id, $0) })
        var propertiesByTypeID: [String: [CatalogProperty]] = [:]
        for (typeID, typeFields) in fields.propertiesByTypeID {
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
        var subjectsByID: [String: CitationComposerModel.GraphSubjectOption] = [:]
        subjectsByID.reserveCapacity(graphSubjects.count)
        var typeKeyByTypeID: [String: String] = [:]
        for subject in graphSubjects {
            if subjectsByID[subject.id] == nil {
                subjectsByID[subject.id] = subject
            }
            if typeKeyByTypeID[subject.typeID] == nil {
                typeKeyByTypeID[subject.typeID] = subject.typeKey
            }
        }
        for (typeID, type) in typesByID {
            typeKeyByTypeID[typeID] = type.key
        }
        var propertyComboOptionsByTypeID: [String: [PVComboBoxOption]] = [:]
        for (typeID, properties) in propertiesByTypeID {
            propertyComboOptionsByTypeID[typeID] = Self.listedProperties(
                typeKey: typeKeyByTypeID[typeID] ?? "",
                properties: properties,
                rules: rules
            ).map { PVComboBoxOption(value: $0.id, label: $0.label, subtext: $0.key) }
        }
        var termComboOptionsByPropertyID: [String: [PVComboBoxOption]] = [:]
        for (propertyID, terms) in termsByPropertyID {
            let propertyKey = propertiesByID[propertyID]?.key ?? ""
            termComboOptionsByPropertyID[propertyID] = terms.map { term in
                PVComboBoxOption(
                    value: term.id,
                    label: PropertyTermDisplay.name(term: term, propertyKey: propertyKey),
                    subtext: term.key
                )
            }
        }
        let subjectComboOptions = graphSubjects.map {
            PVComboBoxOption(value: $0.id, label: $0.label, subtext: $0.ref)
        } + fields.typesInPaletteOrder.compactMap { type -> PVComboBoxOption? in
            guard fields.presentationsByKey[type.key]?.placeable == true else { return nil }
            return PVComboBoxOption(
                value: CitationComposerModel.newSubjectPrefix + type.key,
                label: L10n.CitationComposer.newSubject(typeKey: type.key)
            )
        }
        return CitationComposerVocabulary(
            propertiesByID: propertiesByID,
            propertiesByTypeID: propertiesByTypeID,
            graphSubjects: graphSubjects,
            subjectsByID: subjectsByID,
            termsByPropertyID: termsByPropertyID,
            connectRules: rules,
            typesByID: typesByID,
            subjectComboOptions: subjectComboOptions,
            propertyComboOptionsByTypeID: propertyComboOptionsByTypeID,
            termComboOptionsByPropertyID: termComboOptionsByPropertyID
        )
    }

    func isEdge(subjectID: String, propertyID: String) -> Bool {
        guard let subject = subjectsByID[subjectID] else { return false }
        let propertyKey = propertiesByID[propertyID]?.key ?? ""
        return connectRules.contains { rule in
            !rule.refuse
                && rule.bridgeTypeKey == subject.typeKey
                && rule.edges.contains { $0.propertyKey == propertyKey }
        }
    }

    func propertyOptions(forSubjectID subjectID: String) -> [CatalogProperty] {
        guard let subject = subjectsByID[subjectID] else { return [] }
        return Self.listedProperties(
            typeKey: subject.typeKey,
            properties: propertiesByTypeID[subject.typeID] ?? [],
            rules: connectRules
        )
    }

    func propertyComboOptions(forSubjectID subjectID: String) -> [PVComboBoxOption] {
        guard let subject = subjectsByID[subjectID] else { return [] }
        return propertyComboOptionsByTypeID[subject.typeID] ?? []
    }

    private static func listedProperties(
        typeKey: String,
        properties: [CatalogProperty],
        rules: [CatalogConnectRule]
    ) -> [CatalogProperty] {
        let excluded = Set(
            rules
                .filter { $0.bridgeTypeKey == typeKey && !$0.refuse }
                .flatMap { $0.edges.map(\.propertyKey) }
        )
        return properties
            .filter {
                supportedValueTypes.contains($0.valueType) && !excluded.contains($0.key)
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
