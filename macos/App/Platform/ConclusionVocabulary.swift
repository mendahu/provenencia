import Foundation

// Wire vocabulary Go owns and Swift reads. Each value here mirrors a Go
// constant; `api/ffi/swift_vocabulary_test.go` reads this file and fails
// when the two drift. Spell these values only here: views and models use
// the typed cases, never the strings.

/// A Conclusion field's state (`autoreconcile.State`).
enum ReconciledState: Sendable, Equatable {
    case empty
    case single
    case merged
    case mixed
    case multiple
    case concluded
    /// A state this build doesn't know; shown as no badge.
    case unrecognized(String)

    init(wire: String) {
        switch wire {
        case "": self = .empty
        case "single": self = .single
        case "merged": self = .merged
        case "mixed": self = .mixed
        case "multiple": self = .multiple
        case "concluded": self = .concluded
        default: self = .unrecognized(wire)
        }
    }
}

/// Why the auto-reconciler did or didn't display a value
/// (`autoreconcile.Reason`).
enum ReconcilerReason: Sendable, Equatable {
    case kept
    case folded
    case outvoted
    case weak
    case denied
    case provisional
    case noEvidence
    case against
    /// A reason this build doesn't know; never counted as displayed.
    case unrecognized(String)

    init(wire: String) {
        switch wire {
        case "kept": self = .kept
        case "folded": self = .folded
        case "outvoted": self = .outvoted
        case "weak": self = .weak
        case "denied": self = .denied
        case "provisional": self = .provisional
        case "no_evidence": self = .noEvidence
        case "against": self = .against
        default: self = .unrecognized(wire)
        }
    }
}

/// Why settling moved a Promote row (`graphalign.Move.Reason`). The page words the typed cases.
enum PromoteMoveReason: Sendable, Equatable {
    case neighborPlaced
    case neighborMoved
    case neighborReleased
    /// A code this build doesn't know.
    case unrecognized(String)

    init(wire: String) {
        switch wire {
        case "neighbor_placed": self = .neighborPlaced
        case "neighbor_moved": self = .neighborMoved
        case "neighbor_released": self = .neighborReleased
        default: self = .unrecognized(wire)
        }
    }
}

/// What a Promote proposal row targets (`graphalign.Target`): a record, a held decision, or nothing.
enum PromoteTargetCode: Sendable, Equatable {
    /// Promote suggests no record.
    case unplaced
    case handle
    case new
    case skip
    /// A code this build doesn't know.
    case unrecognized(String)

    init(wire: String) {
        switch wire {
        case "": self = .unplaced
        case "handle": self = .handle
        case "new": self = .new
        case "skip": self = .skip
        default: self = .unrecognized(wire)
        }
    }

    /// The code `init(wire:)` reads, for the send side.
    var wire: String {
        switch self {
        case .unplaced: ""
        case .handle: "handle"
        case .new: "new"
        case .skip: "skip"
        case .unrecognized(let code): code
        }
    }
}

// A Property's value type is `PropertyValueType` (InterpretationValueKinds.swift).

/// The `origin` namespace shared by every catalog vocabulary row
/// (`source_types`, `source_metadata_fields`) — mirrors the `Origin…`
/// constants in `core/database/sourcetypes` and `…/metadatafields`. The FFI
/// layer carries origin as a plain string, not an enum, and the vocabulary
/// is open: anything that is neither seeded nor researcher-authored belongs
/// to a plugin.
enum CatalogOrigin {
    static let provenencia = "provenencia"
    static let user = "user"

    private static let pluginPrefix: String = "plugin:"

    /// Whether a term or Property row is the product's: only then does its key
    /// name a seeded one. A researcher's own row may be keyed the same.
    static func isProduct(_ origin: String) -> Bool {
        origin == provenencia
    }

    /// Anything that is neither seeded nor researcher-authored is owned by a
    /// plugin — the same open-vocabulary stance the badges take.
    static func isPlugin(_ origin: String) -> Bool {
        origin != provenencia && origin != user
    }

    /// The id after `plugin:` (e.g. `"plugin:findagrave"` → `"findagrave"`),
    /// or the raw origin unchanged when it has no such prefix.
    static func pluginID(from origin: String) -> String {
        origin.hasPrefix(pluginPrefix) ? String(origin.dropFirst(pluginPrefix.count)) : origin
    }
}

/// Seeded Property keys the Conclusion pages lay out by name
/// (`docs/seeded-vocabulary.md`, `subjectvocab` seed). User Properties have
/// their own keys, so `propertyKey` stays a String.
enum SeededPropertyKey {
    static let name = "name"
    static let sexAtBirth = "sex_at_birth"
    static let gender = "gender"
    static let toponym = "toponym"
    static let date = "date"
    static let startDate = "start_date"
    static let endDate = "end_date"
    static let role = "role"
    static let eventType = "event_type"
    static let relationshipType = "relationship_type"
    static let placeRelationshipType = "place_relationship_type"
    /// The endpoint Properties of the four bridge types: where a neighbor sits on a link.
    static let person = "person"
    static let relatedTo = "related_to"
    static let event = "event"
    static let place = "place"
    static let from = "from"
    static let to = "to"
}

/// Seeded `gender` terms the Person header colours by (`subjectvocab` seed).
enum SeededGenderTerm {
    static let man = "man"
    static let woman = "woman"
    static let nonBinary = "non_binary"
}

/// Seeded `sex_at_birth` terms (`subjectvocab` seed); the Person header colours by female and male.
enum SeededSexAtBirthTerm {
    static let female = "female"
    static let male = "male"
    static let intersex = "intersex"
    static let indeterminate = "indeterminate"
    static let unknown = "unknown"
}

/// Seeded `role` terms (`subjectvocab` seed).
enum SeededRoleTerm {
    static let subject = "subject"
    static let father = "father"
    static let mother = "mother"
    static let child = "child"
    static let witness = "witness"
    static let informant = "informant"
    static let pallbearer = "pallbearer"
}

/// Seeded `event_type` terms.
enum SeededEventTypeTerm {
    static let birth = "birth"
    static let death = "death"
    static let marriage = "marriage"
    static let engagement = "engagement"
    static let baptism = "baptism"
    static let burial = "burial"
    static let funeral = "funeral"
    static let census = "census"
    static let residence = "residence"
    static let migration = "migration"
    static let travel = "travel"
}

/// Seeded `relationship_type` terms.
enum SeededRelationshipTypeTerm {
    static let spouse = "spouse"
    static let sibling = "sibling"
    static let cousin = "cousin"
    static let parent = "parent"
    static let child = "child"
    static let grandparent = "grandparent"
    static let grandchild = "grandchild"
    static let pibling = "pibling"
    static let nibling = "nibling"
    static let guardian = "guardian"
    static let ward = "ward"

    /// The term read from the other end of the link: a parent's child is a child's parent. A
    /// symmetric term is its own inverse. The one inverse map.
    static let inverse: [String: String] = [
        spouse: spouse,
        sibling: sibling,
        cousin: cousin,
        parent: child,
        child: parent,
        grandparent: grandchild,
        grandchild: grandparent,
        pibling: nibling,
        nibling: pibling,
        guardian: ward,
        ward: guardian,
    ]
}

/// Seeded `place_relationship_type` terms.
enum SeededPlaceRelationshipTypeTerm {
    static let partOf = "part_of"
    static let succeededBy = "succeeded_by"
}

/// A DateValue qualifier (`datevalue.Qualifier*`).
enum DateQualifier {
    static let about = "ABT"
    static let before = "BEF"
    static let after = "AFT"
}

/// Why the engine refuses to unlink an Artifact from a graph (the code of the
/// refusal: `evidencegraphs.artifact_in_use`, `.last_artifact`).
enum EvidenceGraphUnlinkReason: Sendable, Equatable {
    /// The graph has Citations on the Artifact.
    case artifactInUse
    /// The graph keeps at least one Artifact.
    case lastArtifact
    /// A reason this build doesn't know.
    case unrecognized(String)

    init(wire: String) {
        switch wire {
        case "evidencegraphs.artifact_in_use": self = .artifactInUse
        case "evidencegraphs.last_artifact": self = .lastArtifact
        default: self = .unrecognized(wire)
        }
    }

    var wire: String {
        switch self {
        case .artifactInUse: "evidencegraphs.artifact_in_use"
        case .lastArtifact: "evidencegraphs.last_artifact"
        case .unrecognized(let code): code
        }
    }
}
