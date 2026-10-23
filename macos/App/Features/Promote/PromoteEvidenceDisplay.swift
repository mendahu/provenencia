import Foundation

/// The evidence sheet's choices and words, made without a view: which boxes are
/// checked, which bridge a relationship names, and the sentences the sheet reads.
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

    /// How many of the evidence's offered pins are checked.
    static func checkedCount(_ evidence: CatalogClaimEvidence, pins: Set<String>) -> Int {
        pins.intersection(evidence.offeredPinIDs).count
    }

    /// Loaded, and nothing to compare: no line and no relationship.
    static func hasNothingToCompare(_ evidence: CatalogClaimEvidence) -> Bool {
        evidence.sections.allSatisfy { $0.lines.isEmpty } && evidence.relationships.isEmpty
    }

    /// A record behind a line: its Source's title when it has one, and its worded value.
    struct Member: Equatable {
        var source: String?
        var value: String

        var title: String? {
            guard let source, !source.isEmpty else { return nil }
            return source
        }
    }

    /// One record as VoiceOver speaks it; the value alone when its Source has no title.
    static func spokenMember(_ member: Member) -> String {
        guard let title = member.title else { return L10n.Promote.spokenMemberUntitled(value: member.value) }
        return L10n.Promote.spokenMember(source: title, value: member.value)
    }

    /// A line as VoiceOver speaks it: with its record count when several stand behind it, else with
    /// its one record.
    static func spokenLabel(property: String, sentence: String, here: String, count: Int, member: Member?) -> String {
        if count > 1 {
            return L10n.Promote.spokenLineCount(property: property, sentence: sentence, count: count, here: here)
        }
        let line = L10n.Promote.spokenLine(property: property, sentence: sentence, here: here)
        guard let member else { return line }
        return L10n.Promote.spokenLineMember(line: line, member: spokenMember(member))
    }

    /// A record's Source title, or a word for it when the Source has none: on every line and member row.
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

    /// "<name> joins <record title> <REF>": "joins" muted, the ref a ref.
    static func title(name: String, recordTitle: String, ref: String) -> AttributedString {
        StyledFormat.attributed(L10n.Promote.titleHandle, [
            .init(text: name),
            .init(text: L10n.string(L10n.Promote.titleJoins), attributes: styled(.muted)),
            .init(text: recordTitle),
            .init(text: ref, attributes: styled(.ref)),
        ])
    }

    /// The subtitle under the title: the claim in one sentence. A Source with no title is "this source".
    static func subtitle(name: String, sourceTitle: String?, kind: EvidencePrimaryKind, recordTitle: String) -> AttributedString {
        claim(L10n.Promote.subtitle, untitled: L10n.Promote.subtitleUntitled,
              name: name, sourceTitle: sourceTitle, kind: kind, recordTitle: recordTitle)
    }

    /// The evidence heading: what the lines below argue.
    static func evidenceHeading(name: String, sourceTitle: String?, kind: EvidencePrimaryKind, recordTitle: String) -> AttributedString {
        claim(L10n.Promote.evidenceHeading, untitled: L10n.Promote.evidenceHeadingUntitled,
              name: name, sourceTitle: sourceTitle, kind: kind, recordTitle: recordTitle)
    }

    /// A sentence about the row, its Source and the record: names as names, the Source title as a source.
    private static func claim(
        _ format: LocalizedStringResource,
        untitled: LocalizedStringResource,
        name: String,
        sourceTitle: String?,
        kind: EvidencePrimaryKind,
        recordTitle: String
    ) -> AttributedString {
        let kindWord = StyledFormat.Argument(text: L10n.Promote.kindWord(kind))
        let name = StyledFormat.Argument(text: name, attributes: styled(.name))
        let record = StyledFormat.Argument(text: recordTitle, attributes: styled(.name))
        guard let sourceTitle else {
            return StyledFormat.attributed(untitled, [name, kindWord, record])
        }
        return StyledFormat.attributed(format, [
            name, .init(text: sourceTitle, attributes: styled(.source)), kindWord, record,
        ])
    }

    // MARK: Cards

    /// "…they have a similarity in name": the property's label in lowercase for the current locale.
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

    /// A Card header as VoiceOver speaks it: without the ellipsis that opens it on screen.
    static func spokenHeader(_ header: String) -> String {
        header.hasPrefix("…") ? String(header.dropFirst()) : header
    }

    /// "…they share a spouse (CEV-3B5RT and PER-7QH3D)": the connected row's own ref, then its
    /// record's. Only the refs that are known and differ are named.
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
}
