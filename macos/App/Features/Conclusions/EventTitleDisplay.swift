import Foundation

/// Event titles from the naming matrix (R4). Go chooses the rule and its
/// parts (`CatalogEventTitle`); this fills the rule's L10n template. It never
/// chooses a rule, so every surface titles an Event the same way.
enum EventTitleDisplay {
    static func titleSource(_ title: CatalogEventTitle, locale: Locale = .autoupdatingCurrent) -> ConclusionTitleSource {
        let type = typeWord(title)
        switch title.rule {
        case .recordedName:
            return .name(title.recordedName)
        case .subject:
            return .name(L10n.EventTitle.ofOne(type: type, subject: subjectName(title, at: 0), locale: locale))
        case .couple:
            return .name(L10n.EventTitle.marriage(
                a: subjectName(title, at: 0), b: subjectName(title, at: 1), locale: locale
            ))
        case .subjects:
            return .name(L10n.EventTitle.etAl(type: type, first: subjectName(title, at: 0), locale: locale))
        case .label:
            return .label(title.label)
        case .typeAtPlace:
            return .name(L10n.EventTitle.atPlace(type: type, place: title.place, locale: locale))
        case .type:
            return .name(L10n.EventTitle.unspecified(type: type.lowercased(with: locale), locale: locale))
        case .ref:
            return .ref(title.ref)
        }
    }

    static func title(_ title: CatalogEventTitle, locale: Locale = .autoupdatingCurrent) -> String {
        titleSource(title, locale: locale).text
    }

    /// An Event whose header has not loaded: its label, else its ref.
    static func untitled(_ entity: CatalogCanonicalEntity) -> ConclusionTitleSource {
        let label = entity.label.trimmingCharacters(in: .whitespacesAndNewlines)
        return label.isEmpty ? .ref(entity.ref) : .label(label)
    }

    /// A subject's displayed name, or "unnamed person".
    private static func subjectName(_ title: CatalogEventTitle, at index: Int) -> String {
        guard title.subjects.indices.contains(index), let name = title.subjects[index] else {
            return L10n.string(L10n.EventTitle.unnamedPerson)
        }
        let text = NameValueDisplay.string(for: name).trimmingCharacters(in: .whitespacesAndNewlines)
        return text.isEmpty ? L10n.string(L10n.EventTitle.unnamedPerson) : text
    }

    /// The type's label, else "Event".
    private static func typeWord(_ title: CatalogEventTitle) -> String {
        title.typeLabel.isEmpty ? L10n.string(L10n.EventTitle.fallbackType) : title.typeLabel
    }
}
