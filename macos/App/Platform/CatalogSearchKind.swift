import Foundation

/// The omnibar's hit kinds (`SearchHit.kind`), one case per `core/search`
/// `Registry` entry. `api/ffi/swift_vocabulary_test.go` fails if a case and
/// the registry drift. Raw values are explicit so the test reads every case.
enum CatalogSearchKind: String, CaseIterable, Sendable {
    case source = "source"
    case sourceType = "source_type"
    case metadataField = "metadata_field"
    case person = "person"
    case event = "event"
    case place = "place"
    case evidenceGraph = "evidence_graph"

    /// Go's handle kinds are the primary kinds (`KindPerson = vocabkey.Person`).
    init(_ kind: EvidencePrimaryKind) {
        switch kind {
        case .person: self = .person
        case .event: self = .event
        case .place: self = .place
        }
    }
}

extension CatalogSearchHit {
    /// The hit's kind, or nil for a kind this build doesn't know.
    var searchKind: CatalogSearchKind? { CatalogSearchKind(rawValue: kind) }
}
