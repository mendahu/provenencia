import Foundation

/// A localized format with one `%@` that wraps a decorated name word, and whether the word
/// is emphasized where the title can draw it.
struct NamePartDecoration {
    let format: LocalizedStringResource
    let emphasized: Bool
}

/// Product NameValue part-type registry. Keys match `namevalue.PartTypes()`.
enum NamePartType: String, CaseIterable, Sendable {
    case prefix = "prefix"
    case given = "given"
    case nick = "nick"
    case surnamePrefix = "surname_prefix"
    case surname = "surname"
    case formerSurname = "former_surname"
    case suffix = "suffix"
    case undetermined = "undetermined"

    var label: LocalizedStringResource {
        switch self {
        case .prefix: L10n.NameValue.partTypePrefix
        case .given: L10n.NameValue.partTypeGiven
        case .nick: L10n.NameValue.partTypeNick
        case .surnamePrefix: L10n.NameValue.partTypeSurnamePrefix
        case .surname: L10n.NameValue.partTypeSurname
        case .formerSurname: L10n.NameValue.partTypeFormerSurname
        case .suffix: L10n.NameValue.partTypeSuffix
        case .undetermined: L10n.NameValue.partTypeUndetermined
        }
    }

    /// How a displayed name decorates a word of this type; nil shows it as typed.
    /// Display only: Go's `nameformats.Format` sort key never decorates.
    var decoration: NamePartDecoration? {
        switch self {
        case .nick: NamePartDecoration(format: L10n.NameValue.displayNick, emphasized: false)
        case .formerSurname: NamePartDecoration(format: L10n.NameValue.displayFormerSurname, emphasized: true)
        case .prefix, .given, .surnamePrefix, .surname, .suffix, .undetermined: nil
        }
    }

    /// Empty is untyped (allowed). Non-empty must be a registry key.
    static func isAllowed(_ raw: String) -> Bool {
        let key = raw.trimmingCharacters(in: .whitespacesAndNewlines)
        return key.isEmpty || NamePartType(rawValue: key) != nil
    }

    static var selectOptions: [PVSelectOption] {
        [PVSelectOption(value: "", label: L10n.string(L10n.NameValue.partTypeNone))]
            + NamePartType.allCases.map {
                PVSelectOption(value: $0.rawValue, label: L10n.string($0.label))
            }
    }

    static func label(forRaw raw: String) -> LocalizedStringResource {
        let key = raw.trimmingCharacters(in: .whitespacesAndNewlines)
        if key.isEmpty { return L10n.NameValue.partTypeNone }
        return NamePartType(rawValue: key)?.label ?? L10n.NameValue.partTypeNone
    }
}
