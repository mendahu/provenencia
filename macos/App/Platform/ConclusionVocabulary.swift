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

// A Property's value type is `PropertyValueType` (InterpretationValueKinds.swift).

/// Seeded Property keys the Conclusion pages lay out by name
/// (`docs/seeded-vocabulary.md`, `subjectvocab` seed). User Properties have
/// their own keys, so `propertyKey` stays a String.
enum SeededPropertyKey {
    static let name = "name"
    static let sexAtBirth = "sex_at_birth"
    static let toponym = "toponym"
    static let date = "date"
    static let startDate = "start_date"
    static let endDate = "end_date"
}

/// A DateValue qualifier (`datevalues.Qualifier*`).
enum DateQualifier {
    static let about = "ABT"
    static let before = "BEF"
    static let after = "AFT"
}
