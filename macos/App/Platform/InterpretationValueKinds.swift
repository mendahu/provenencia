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
    case depiction

    var label: LocalizedStringResource {
        switch self {
        case .text: return L10n.Properties.valueTypeText
        case .integer: return L10n.Properties.valueTypeInteger
        case .date: return L10n.Properties.valueTypeDate
        case .name: return L10n.Properties.valueTypeName
        case .subject: return L10n.Properties.valueTypeSubject
        case .term: return L10n.Properties.valueTypeTerm
        case .depiction: return L10n.Properties.valueTypeDepiction
        }
    }
}
