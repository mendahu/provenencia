import Foundation

/// Session-cached Subject fields destination payload (properties + types + bindings).
struct SubjectFieldsSnapshot: Sendable, Equatable {
    var properties: [CatalogProperty]
    var types: [CatalogSubjectType]
    /// Bindings keyed by subject type id.
    var fieldsByTypeID: [String: [CatalogSubjectTypeField]]

    static let empty = SubjectFieldsSnapshot(properties: [], types: [], fieldsByTypeID: [:])

    func boundTypeKeys(for propertyID: String) -> [String] {
        var keys: [String] = []
        for type in types {
            if fieldsByTypeID[type.id]?.contains(where: { $0.property.id == propertyID }) == true {
                keys.append(type.key)
            }
        }
        return keys
    }

    func binding(propertyID: String, typeID: String) -> CatalogSubjectTypeField? {
        fieldsByTypeID[typeID]?.first { $0.property.id == propertyID }
    }

    func propertyCount(forTypeID typeID: String) -> Int {
        fieldsByTypeID[typeID]?.count ?? 0
    }
}
