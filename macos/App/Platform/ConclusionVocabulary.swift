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

// A Property's value type is `PropertyValueType` (InterpretationValueKinds.swift).

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

/// A DateValue qualifier (`datevalue.Qualifier*`).
enum DateQualifier {
    static let about = "ABT"
    static let before = "BEF"
    static let after = "AFT"
}
