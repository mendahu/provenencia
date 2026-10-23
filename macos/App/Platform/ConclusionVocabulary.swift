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

/// Why a Promote row reads as it does, and why settling moved it
/// (`graphalign.Reason`). The page words the typed cases.
enum PromoteReasonCode: Sendable, Equatable {
    case via
    case decided
    case agrees
    case weak
    case taken
    case noMatch
    case empty
    case neighborPlaced
    case neighborMoved
    case neighborReleased
    /// A code this build doesn't know.
    case unrecognized(String)

    init(wire: String) {
        switch wire {
        case "via": self = .via
        case "decided": self = .decided
        case "agrees": self = .agrees
        case "weak": self = .weak
        case "taken": self = .taken
        case "no_match": self = .noMatch
        case "empty": self = .empty
        case "neighbor_placed": self = .neighborPlaced
        case "neighbor_moved": self = .neighborMoved
        case "neighbor_released": self = .neighborReleased
        default: self = .unrecognized(wire)
        }
    }
}

/// How a claim evidence line or relationship stands (`propertycompare.Outcome`).
enum EvidenceOutcome: Sendable, Equatable {
    case agree
    case partial
    case conflict
    case unknown
    /// A code this build doesn't know.
    case unrecognized(String)

    init(wire: String) {
        switch wire {
        case "agree": self = .agree
        case "partial": self = .partial
        case "conflict": self = .conflict
        case "unknown": self = .unknown
        default: self = .unrecognized(wire)
        }
    }
}

/// Why a line's outcome came out as it did (`propertycompare.Cause`).
enum EvidencePropertyCause: Sendable, Equatable {
    /// No cause: the code is empty.
    case none
    case relation
    case negative
    case noValues
    case cardinality
    /// A code this build doesn't know.
    case unrecognized(String)

    init(wire: String) {
        switch wire {
        case "": self = .none
        case "relation": self = .relation
        case "negative": self = .negative
        case "no_values": self = .noValues
        case "cardinality": self = .cardinality
        default: self = .unrecognized(wire)
        }
    }
}

/// How two values relate (`valuecompare.Relation`).
enum ValueRelation: Sendable, Equatable {
    /// No relation: the code is empty (no pair spoke).
    case none
    case same
    case equivalent
    case contains
    case overlaps
    case near
    case different
    case unknown
    /// A code this build doesn't know.
    case unrecognized(String)

    init(wire: String) {
        switch wire {
        case "": self = .none
        case "same": self = .same
        case "equivalent": self = .equivalent
        case "contains": self = .contains
        case "overlaps": self = .overlaps
        case "near": self = .near
        case "different": self = .different
        case "unknown": self = .unknown
        default: self = .unrecognized(wire)
        }
    }
}

/// Why two values relate as they do (`valuecompare.Cause`).
enum ValueCause: Sendable, Equatable {
    /// No cause: the code is empty.
    case none
    case exact
    case letterCase
    case fold
    case alias
    case abbreviation
    case initial
    case spelling
    case precision
    case missingPart
    case overlap
    case gap
    case beyondTolerance
    case generationConflict
    case empty
    case phraseOnly
    case belowMinimum
    case noMatch
    case tooLong
    case order
    case role
    case unsupported
    /// Two values the layer cannot relate: the researcher judges them.
    case unjudged
    /// A code this build doesn't know.
    case unrecognized(String)

    init(wire: String) {
        switch wire {
        case "": self = .none
        case "exact": self = .exact
        case "case": self = .letterCase
        case "fold": self = .fold
        case "alias": self = .alias
        case "abbreviation": self = .abbreviation
        case "initial": self = .initial
        case "spelling": self = .spelling
        case "precision": self = .precision
        case "missing_part": self = .missingPart
        case "overlap": self = .overlap
        case "gap": self = .gap
        case "beyond_tolerance": self = .beyondTolerance
        case "generation_conflict": self = .generationConflict
        case "empty": self = .empty
        case "phrase_only": self = .phraseOnly
        case "below_minimum": self = .belowMinimum
        case "no_match": self = .noMatch
        case "too_long": self = .tooLong
        case "order": self = .order
        case "role": self = .role
        case "unsupported": self = .unsupported
        case "unjudged": self = .unjudged
        default: self = .unrecognized(wire)
        }
    }
}

/// Why a relationship adds nothing or subtracts (`segmentcompare.Cause`).
enum SegmentCause: Sendable, Equatable {
    /// No cause: the code is empty.
    case none
    case notCorresponding
    case gated
    case placedApart
    case unplacedDiffers
    case unplacedUnsure
    /// A code this build doesn't know.
    case unrecognized(String)

    init(wire: String) {
        switch wire {
        case "": self = .none
        case "not_corresponding": self = .notCorresponding
        case "gated": self = .gated
        case "placed_apart": self = .placedApart
        case "unplaced_differs": self = .unplacedDiffers
        case "unplaced_unsure": self = .unplacedUnsure
        default: self = .unrecognized(wire)
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
}

/// Seeded `gender` terms the Person header colours by (`subjectvocab` seed).
enum SeededGenderTerm {
    static let man = "man"
    static let woman = "woman"
    static let nonBinary = "non_binary"
}

/// Seeded `sex_at_birth` terms the Person header colours by.
enum SeededSexAtBirthTerm {
    static let female = "female"
    static let male = "male"
}

/// Seeded `role` terms (`subjectvocab` seed).
enum SeededRoleTerm {
    static let subject = "subject"
    static let father = "father"
    static let mother = "mother"
    static let spouse = "spouse"
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
