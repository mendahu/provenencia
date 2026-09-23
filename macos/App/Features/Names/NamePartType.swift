import Foundation

/// Product NameValue part-type registry. Keys match `namevalues.PartTypes()`.
enum NamePartType: String, CaseIterable, Sendable {
    case prefix
    case given
    case initial
    case nick
    case surnamePrefix = "surname_prefix"
    case surname
    case suffix
    case undetermined

    var label: LocalizedStringResource {
        switch self {
        case .prefix: L10n.NameValue.partTypePrefix
        case .given: L10n.NameValue.partTypeGiven
        case .initial: L10n.NameValue.partTypeInitial
        case .nick: L10n.NameValue.partTypeNick
        case .surnamePrefix: L10n.NameValue.partTypeSurnamePrefix
        case .surname: L10n.NameValue.partTypeSurname
        case .suffix: L10n.NameValue.partTypeSuffix
        case .undetermined: L10n.NameValue.partTypeUndetermined
        }
    }

    /// Empty is untyped (allowed). Non-empty must be a registry key.
    static func isAllowed(_ raw: String) -> Bool {
        let key = raw.trimmingCharacters(in: .whitespacesAndNewlines)
        return key.isEmpty || NamePartType(rawValue: key) != nil
    }

    static func label(forRaw raw: String) -> LocalizedStringResource {
        let key = raw.trimmingCharacters(in: .whitespacesAndNewlines)
        if key.isEmpty { return L10n.NameValue.partTypeNone }
        return NamePartType(rawValue: key)?.label ?? L10n.NameValue.partTypeNone
    }
}
