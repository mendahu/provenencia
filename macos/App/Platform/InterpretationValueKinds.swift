import Foundation

enum ObservationPolarity: String, Sendable {
    case positive
    case negative
}

enum PropertyValueType: String, Sendable {
    case text
    case integer
    case date
    case name
    case subject
    case term
}
