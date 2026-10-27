import Foundation

/// The evidence sheet's choices and words, made without a view: which boxes are
/// checked and the sentences the sheet reads.
enum PromoteEvidenceDisplay {
    /// How much of a line's or relationship's pins are checked.
    enum CheckState: Equatable {
        case checked
        case unchecked
        case partial
    }

    /// All pins in `pins` is checked, none is unchecked, some is partial. No pins is unchecked.
    static func checkState(pinIDs: [String], pins: Set<String>) -> CheckState {
        guard !pinIDs.isEmpty else { return .unchecked }
        let checked = pinIDs.filter(pins.contains).count
        if checked == 0 { return .unchecked }
        return checked == pinIDs.count ? .checked : .partial
    }

    /// How many of the offered pins are checked.
    static func checkedCount(offered: Set<String>, pins: Set<String>) -> Int {
        pins.intersection(offered).count
    }

    /// One record as VoiceOver speaks it: "Census 1851: J. Smith.", or "a record: J. Smith." when its
    /// Source has no title (``recordSource(title:)``). `value` is the record's worded value.
    static func spokenMember(source: String, value: String) -> String {
        L10n.Promote.spokenMember(source: source, value: value)
    }

    /// A line as VoiceOver speaks it: with its record count when several stand behind it, else with
    /// its one record.
    static func spokenLabel(property: String, sentence: String, here: String, count: Int, member: (source: String, value: String)?) -> String {
        if count > 1 {
            return L10n.Promote.spokenLineCount(property: property, sentence: sentence, count: count, here: here)
        }
        let line = L10n.Promote.spokenLine(property: property, sentence: sentence, here: here)
        guard let member else { return line }
        return L10n.Promote.spokenLineMember(line: line, member: spokenMember(source: member.source, value: member.value))
    }

    /// A relationship as VoiceOver speaks it: with its record count when it bundles several.
    static func spokenRelationship(header: String, outcome: EvidenceOutcome, count: Int) -> String {
        let word = L10n.string(L10n.Promote.outcome(outcome))
        return count > 1
            ? L10n.Promote.spokenRelationshipCount(sentence: header, outcome: word, count: count)
            : L10n.Promote.spokenRelationship(sentence: header, outcome: word)
    }

    /// A likeness line as VoiceOver speaks it: with its record count when it bundles several.
    static func spokenLikeness(source: String, sentence: String, count: Int) -> String {
        count > 1
            ? L10n.Promote.spokenLikenessCount(source: source, sentence: sentence, count: count)
            : L10n.Promote.spokenLikeness(source: source, sentence: sentence)
    }

    /// A record's Source title, or a word for it when the Source has none: the one rule, for every
    /// line, member and spoken label.
    static func recordSource(title: String?) -> String {
        guard let title, !title.isEmpty else { return L10n.string(L10n.Promote.untitledRecord) }
        return title
    }

    // MARK: Title, subtitle and heading

    /// What a styled run of the title, subtitle or heading is, as a role. The sheet maps a role to
    /// a font and a color; this file knows no view type.
    enum Role: String, Hashable, Codable, Sendable {
        /// A subject's or record's name.
        case name
        /// A Source's title.
        case source
        /// Connecting words set back from the names.
        case muted
        /// A ref.
        case ref
    }

    /// The attribute that carries a run's ``Role``.
    enum RoleAttribute: AttributedStringKey {
        typealias Value = Role
        static let name = "promoteEvidenceRole"
    }

    private static func styled(_ role: Role) -> AttributeContainer {
        var attributes = AttributeContainer()
        attributes[RoleAttribute.self] = role
        return attributes
    }

    /// "<name> joins <record title> <REF>": the connecting word muted, the ref a ref.
    static func title(name: String, recordTitle: String, ref: String) -> AttributedString {
        StyledFormat.attributed(L10n.Promote.titleHandle, [
            .init(text: name),
            .init(text: recordTitle),
            .init(text: ref, attributes: styled(.ref)),
        ], literal: styled(.muted))
    }

    /// The subtitle under the title: the claim in one sentence. A Source with no title is "this source".
    static func subtitle(name: String, sourceTitle: String?, kind: EvidencePrimaryKind, recordTitle: String) -> AttributedString {
        claim(L10n.Promote.subtitle(kind), untitled: L10n.Promote.subtitleUntitled(kind),
              name: name, sourceTitle: sourceTitle, recordTitle: recordTitle)
    }

    /// The evidence heading: what the lines below argue.
    static func evidenceHeading(name: String, sourceTitle: String?, kind: EvidencePrimaryKind, recordTitle: String) -> AttributedString {
        claim(L10n.Promote.evidenceHeading(kind), untitled: L10n.Promote.evidenceHeadingUntitled(kind),
              name: name, sourceTitle: sourceTitle, recordTitle: recordTitle)
    }

    /// A sentence about the row, its Source and the record: names as names, the Source title as a source.
    private static func claim(
        _ format: LocalizedStringResource,
        untitled: LocalizedStringResource,
        name: String,
        sourceTitle: String?,
        recordTitle: String
    ) -> AttributedString {
        let name = StyledFormat.Argument(text: name, attributes: styled(.name))
        let record = StyledFormat.Argument(text: recordTitle, attributes: styled(.name))
        guard let sourceTitle else {
            return StyledFormat.attributed(untitled, [name, record])
        }
        return StyledFormat.attributed(format, [
            name, .init(text: sourceTitle, attributes: styled(.source)), record,
        ])
    }

    // MARK: Cards

    /// "they have a similarity in name": the property's label in lowercase for the current locale.
    /// The Card draws it after an ellipsis (`L10n.Promote.cardHeaderLead`); VoiceOver reads it as is.
    static func similarityHeader(property: String) -> String {
        L10n.Promote.similarityIn(property: property.lowercased(with: .autoupdatingCurrent))
    }

    /// Where a Card says what this Source has: once in its header when every line agrees on it,
    /// else under each line's sentence. A Card with a likeness says it nowhere.
    enum ThisSourcePlacement: Equatable {
        case nowhere
        case header(String)
        case lines

        /// What the Card's header says, when it says it.
        var headerValue: String? {
            if case .header(let value) = self { return value }
            return nil
        }
    }

    /// `values` are the lines' incoming values as worded; nil for a likeness.
    static func thisSourcePlacement(values: [String?]) -> ThisSourcePlacement {
        let worded = values.compactMap { $0 }
        guard let first = worded.first, worded.count == values.count else { return .nowhere }
        return worded.allSatisfy { $0 == first } ? .header(first) : .lines
    }

    /// What the row shares with a connected row, after "they share": the engine's reading of the
    /// link from the neighbor's side says it; with none the sheet can word, "a connection".
    static func sharePhrase(relationship: CatalogClaimEvidenceRelationship, neighborName: String) -> String {
        ConnectionSentenceRegistry.share(
            ConnectionSentenceRegistry.ShareInput(relationship, neighborName: neighborName),
            similar: relationship.outcome == .partial
        )
    }

    /// "they share a spouse (CEV-3B5RT and PER-7QH3D)": the connected row's own ref, then its
    /// record's. Only the refs that are known and differ are named. The Card draws it after an
    /// ellipsis; VoiceOver reads it as is.
    static func shareHeader(phrase: String, ownRef: String?, recordRef: String?) -> String {
        var refs: [String] = []
        for ref in [ownRef, recordRef] {
            if let ref, !ref.isEmpty, !refs.contains(ref) { refs.append(ref) }
        }
        switch refs.count {
        case 0: return L10n.Promote.shareHeaderBare(phrase: phrase)
        case 1: return L10n.Promote.shareHeaderOne(phrase: phrase, ref: refs[0])
        default: return L10n.Promote.shareHeader(phrase: phrase, first: refs[0], second: refs[1])
        }
    }

    // MARK: How each side recorded a link

    /// How a Source and the record each say the link: the Card's caption pair, "This source" first.
    /// Nil when both record it directly. A side that records it through a marriage, a birth or a
    /// parent says so; one that records it directly says it in the link's own term.
    static func recorded(
        layer: KinshipPattern,
        layerPivot: EvidenceBoard.Pivot?,
        canon: KinshipPattern,
        canonPivot: EvidenceBoard.Pivot?,
        term: String?
    ) -> (source: String, record: String)? {
        guard layer != .direct || canon != .direct else { return nil }
        func side(_ pattern: KinshipPattern, _ pivot: EvidenceBoard.Pivot?) -> String {
            wording(pattern).map { implied($0, pivot) } ?? direct(term: term)
        }
        return (side(layer, layerPivot), side(canon, canonPivot))
    }

    /// What kind of record a pattern's pivot is.
    private enum PivotShape {
        case event
        case person
    }

    /// The words of one kinship pattern: its pivot, a phrase that names the pivot, and one that cannot.
    private struct PatternWording {
        var pivot: PivotShape
        var named: (String) -> String
        var none: LocalizedStringResource
    }

    /// The one place a kinship pattern gets its words; nil for a direct side, which is worded by its term
    /// (``direct(term:)``). Exhaustive, so a new pattern does not compile
    /// without them.
    private static func wording(_ pattern: KinshipPattern) -> PatternWording? {
        switch pattern {
        case .marriageSpouses:
            PatternWording(pivot: .event, named: L10n.Promote.recordedMarriage, none: L10n.Promote.recordedMarriageNone)
        case .birthParents:
            PatternWording(pivot: .event, named: L10n.Promote.recordedBirth, none: L10n.Promote.recordedBirthNone)
        case .parentOfParent:
            PatternWording(pivot: .person, named: L10n.Promote.recordedChain, none: L10n.Promote.recordedChainNone)
        case .unrecognized:
            PatternWording(pivot: .person, named: L10n.Promote.recordedOther, none: L10n.Promote.recordedOtherNone)
        case .direct:
            nil
        }
    }

    /// An implied side: its pivot named by date (an event) or name (a person), else by ref, else not at all.
    private static func implied(_ wording: PatternWording, _ pivot: EvidenceBoard.Pivot?) -> String {
        let label: String
        switch wording.pivot {
        case .event: label = pivot.map { $0.date.isEmpty ? "" : L10n.Promote.recordedOn($0.date) } ?? ""
        case .person: label = pivot?.name ?? ""
        }
        if !label.isEmpty { return wording.named(label) }
        if let ref = pivot?.ref, !ref.isEmpty { return wording.named(ref) }
        return L10n.string(wording.none)
    }

    /// A direct side, in the term of the link both sides record.
    private static func direct(term: String?) -> String {
        let words = term.flatMap {
            L10n.PropertyTerm.wording(propertyKey: SeededPropertyKey.relationshipType, termKey: $0)?.recordedDirect
        }
        return L10n.string(words ?? L10n.Promote.recordedDirectOther)
    }
}
