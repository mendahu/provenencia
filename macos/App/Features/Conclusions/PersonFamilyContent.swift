import Foundation

/// What a relative is to the person on the Family section (spoken before the
/// tile). Family's own; the tile takes the label.
enum PersonFamilyRole: Equatable {
    case parent, sibling, partner, child

    var label: String {
        switch self {
        case .parent: L10n.string(L10n.Conclusions.familyRoleParent)
        case .sibling: L10n.string(L10n.Conclusions.familyRoleSibling)
        case .partner: L10n.string(L10n.Conclusions.familyRolePartner)
        case .child: L10n.string(L10n.Conclusions.familyRoleChild)
        }
    }
}

/// The Family section's two columns (S13-12, board S13-D3): the origin on the
/// left (the parent sets, then the siblings with the person among them) and the
/// family the person formed on the right (each partner with their children).
/// Laid out from what Go gives, in Go's order, with no regrouping. Pure; the
/// layout places it and the view draws it.
struct PersonFamilyContent: Equatable {
    /// One person on the page. `location` opens the relative on Family; nil for
    /// the person.
    struct Tile: Equatable, Identifiable {
        /// "\(column)-\(group)-\(entityId)", unique on the page.
        var id: String
        var content: PersonTileContent
        var location: WorkspaceLocation?
    }

    /// A sibling, once, with the parent sets it is a child of (indexes into
    /// `parentSets`; empty when recorded only).
    struct Sibling: Equatable {
        var tile: Tile
        var sets: [Int]
    }

    /// One partner and the children with them, or the children with no other
    /// known parent (`partner` nil).
    struct Group: Equatable {
        var partner: Tile?
        var children: [Tile]
        /// What VoiceOver calls the group.
        var label: String
    }

    var parentSets: [[Tile]]
    /// One caption per parent set.
    var parentCaptions: [String]
    var siblings: [Sibling]
    /// True when there are no parent sets and at least one sibling besides the person.
    var siblingsCaption: Bool
    /// What VoiceOver calls the siblings group.
    var siblingsLabel: String
    var groups: [Group]
    /// The right column's one caption, over its first group; nil with no groups.
    var rightCaption: String?

    /// Nothing to draw but the person: no parent sets, no sibling besides the
    /// person, and no groups.
    var isEmpty: Bool {
        parentSets.isEmpty && !siblings.contains { !$0.tile.content.isFocus } && groups.isEmpty
    }

    init(family: CatalogPersonFamily, locale: Locale = .autoupdatingCurrent, orders: NameFormatOrders = NameFormatOrders()) {
        let focusID = family.person?.entity.id
        func tile(_ member: CatalogFamilyMember, role: PersonFamilyRole, column: String, group: String) -> Tile {
            let summary = member.person
            return Tile(
                id: "\(column)-\(group)-\(summary.entity.id)",
                content: PersonTileContent(summary: summary, role: role.label, isFocus: false, locale: locale, orders: orders),
                location: PersonsList.location(summary, orders: orders).withPageSection(.family)
            )
        }

        // Left: the parent sets, as Go grouped them.
        let parentUnits = family.units.filter { $0.kind == .parents }
        parentSets = parentUnits.enumerated().map { k, unit in
            unit.partners.map { tile($0, role: .parent, column: "l", group: "p\(k)") }
        }
        parentCaptions = parentUnits.enumerated().map { k, unit in
            let plural = unit.partners.count > 1 || (k == 0 && parentUnits.count > 1)
            switch (k == 0, plural) {
            case (true, true): return L10n.string(L10n.Conclusions.familyCaptionParents)
            case (true, false): return L10n.string(L10n.Conclusions.familyCaptionParent)
            case (false, true): return L10n.string(L10n.Conclusions.familyCaptionOtherParents)
            case (false, false): return L10n.string(L10n.Conclusions.familyCaptionOtherParent)
            }
        }

        // Left: the siblings, once, in Go's order, the person among them.
        siblings = family.originSiblings.map { origin in
            let summary = origin.member.person
            if summary.entity.id == focusID {
                let focus = Tile(
                    id: "l-s-\(summary.entity.id)",
                    content: PersonTileContent(summary: summary, role: nil, isFocus: true, locale: locale, orders: orders),
                    location: nil
                )
                return Sibling(tile: focus, sets: origin.parentSets)
            }
            return Sibling(tile: tile(origin.member, role: .sibling, column: "l", group: "s"), sets: origin.parentSets)
        }
        let others = siblings.contains { !$0.tile.content.isFocus }
        siblingsCaption = parentUnits.isEmpty && others
        siblingsLabel = others
            ? L10n.string(L10n.Conclusions.familyGroupThisPersonAndSiblings)
            : L10n.string(L10n.Conclusions.familyThisPerson)

        // Right: each partner with their children, then the children with no other parent.
        var groups: [Group] = []
        for unit in family.units where unit.kind == .partner {
            guard let member = unit.partners.first else { continue }
            let index = groups.count
            let partner = tile(member, role: .partner, column: "r", group: "\(index)")
            let name = partner.content.title.spokenText
            groups.append(Group(
                partner: partner,
                children: unit.children.map { tile($0, role: .child, column: "r", group: "\(index)") },
                label: unit.children.isEmpty
                    ? L10n.Conclusions.familyGroupPartner(name)
                    : L10n.Conclusions.familyGroupPartnerAndChildren(name)
            ))
        }
        for unit in family.units where unit.kind == .oneParent {
            let index = groups.count
            groups.append(Group(
                partner: nil,
                children: unit.children.map { tile($0, role: .child, column: "r", group: "\(index)") },
                label: L10n.string(L10n.Conclusions.familyCaptionChildren)
            ))
        }
        self.groups = groups
        let partners = groups.filter { $0.partner != nil }.count
        let anyChildren = groups.contains { !$0.children.isEmpty }
        if groups.isEmpty {
            rightCaption = nil
        } else {
            switch (partners, anyChildren) {
            case (0, _): rightCaption = L10n.string(L10n.Conclusions.familyCaptionChildren)
            case (1, true): rightCaption = L10n.string(L10n.Conclusions.familyCaptionPartnerAndChildren)
            case (1, false): rightCaption = L10n.string(L10n.Conclusions.familyCaptionPartner)
            case (_, true): rightCaption = L10n.string(L10n.Conclusions.familyCaptionPartnersAndChildren)
            case (_, false): rightCaption = L10n.string(L10n.Conclusions.familyCaptionPartners)
            }
        }
    }
}
