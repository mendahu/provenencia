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
/// Laid out from what Go gives, in Go's order, with no regrouping. Each list is
/// folded at `foldLimit` behind a "Show N more" row until its `Fold` is in
/// `expanded`; Go still returns every member, only the first paint is bounded.
/// Pure; the layout places it and the view draws it.
struct PersonFamilyContent: Equatable {
    /// How many tiles a list shows before it folds.
    static let foldLimit = 30

    /// A list that can be folded. `children` takes the group's index in Go's
    /// order (before the partner fold).
    enum Fold: Hashable {
        case parentSets, siblings, partners
        case children(Int)
    }

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
        /// How many children the fold hides.
        var childrenHidden: Int
        /// What VoiceOver calls the group.
        var label: String
    }

    var parentSets: [[Tile]]
    /// One caption per parent set.
    var parentCaptions: [String]
    /// How many parent sets the fold hides.
    var parentSetsHidden: Int
    var siblings: [Sibling]
    /// How many siblings the fold hides.
    var siblingsHidden: Int
    /// Over the siblings when there are no parent sets and at least one sibling
    /// besides the person; nil otherwise.
    var siblingsCaption: String?
    /// What VoiceOver calls the siblings group.
    var siblingsLabel: String
    var groups: [Group]
    /// How many partner groups the fold hides.
    var groupsHidden: Int
    /// The right column's one caption, over its first group; nil with no groups.
    var rightCaption: String?

    /// Every tile on the page, in reading order.
    var allTiles: [Tile] {
        parentSets.flatMap { $0 } + siblings.map(\.tile)
            + groups.flatMap { ($0.partner.map { [$0] } ?? []) + $0.children }
    }

    /// Nothing to draw but the person: no parent sets, no sibling besides the
    /// person, and no groups.
    var isEmpty: Bool {
        parentSets.isEmpty && !siblings.contains { !$0.tile.content.isFocus } && groups.isEmpty
    }

    init(
        family: CatalogPersonFamily,
        expanded: Set<Fold> = [],
        locale: Locale = .autoupdatingCurrent,
        orders: NameFormatOrders = NameFormatOrders()
    ) {
        let limit = Self.foldLimit
        /// The first `limit` of a list, or all of it once `fold` is expanded.
        func folded<T>(_ list: [T], _ fold: Fold) -> [T] {
            expanded.contains(fold) ? list : Array(list.prefix(limit))
        }
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
        let allParentUnits = family.units.filter { $0.kind == .parents }
        let parentUnits = folded(allParentUnits, .parentSets)
        parentSetsHidden = allParentUnits.count - parentUnits.count
        parentSets = parentUnits.enumerated().map { k, unit in
            unit.partners.map { tile($0, role: .parent, column: "l", group: "p\(k)") }
        }
        parentCaptions = parentUnits.indices.map { k in
            let plural = parentUnits[k].partners.count > 1 || (k == 0 && allParentUnits.count > 1)
            switch (k == 0, plural) {
            case (true, true): return L10n.string(L10n.Conclusions.familyCaptionParents)
            case (true, false): return L10n.string(L10n.Conclusions.familyCaptionParent)
            case (false, true): return L10n.string(L10n.Conclusions.familyCaptionOtherParents)
            case (false, false): return L10n.string(L10n.Conclusions.familyCaptionOtherParent)
            }
        }

        // Left: the siblings, once, in Go's order, the person among them. A
        // fold keeps the person when it falls after the first `limit`.
        var shownSiblings = folded(family.originSiblings, .siblings)
        if shownSiblings.count < family.originSiblings.count,
           let at = family.originSiblings.firstIndex(where: { $0.member.person.entity.id == focusID }),
           at >= limit {
            shownSiblings.append(family.originSiblings[at])
        }
        siblingsHidden = family.originSiblings.count - shownSiblings.count
        let visibleSets = parentUnits.count
        siblings = shownSiblings.map { origin in
            let summary = origin.member.person
            let sets = origin.parentSets.filter { $0 < visibleSets }
            if summary.entity.id == focusID {
                let focus = Tile(
                    id: "l-s-\(summary.entity.id)",
                    content: PersonTileContent(summary: summary, role: nil, isFocus: true, locale: locale, orders: orders),
                    location: nil
                )
                return Sibling(tile: focus, sets: sets)
            }
            return Sibling(tile: tile(origin.member, role: .sibling, column: "l", group: "s"), sets: sets)
        }
        let others = siblings.contains { !$0.tile.content.isFocus }
        siblingsCaption = allParentUnits.isEmpty && others ? L10n.string(L10n.Conclusions.familyCaptionSiblings) : nil
        siblingsLabel = others
            ? L10n.string(L10n.Conclusions.familyGroupThisPersonAndSiblings)
            : L10n.string(L10n.Conclusions.familyThisPerson)

        // Right: each partner with their children, then the children with no other parent.
        let partnerUnits = family.units.filter { $0.kind == .partner && !$0.partners.isEmpty }
        let oneParentUnits = family.units.filter { $0.kind == .oneParent }
        let allUnits = partnerUnits + oneParentUnits
        let shownUnits = folded(allUnits, .partners)
        func children(_ unit: CatalogFamilyUnit, group index: Int) -> (tiles: [Tile], hidden: Int) {
            let shown = folded(unit.children, .children(index))
            return (shown.map { tile($0, role: .child, column: "r", group: "\(index)") }, unit.children.count - shown.count)
        }
        var groups: [Group] = []
        for (index, unit) in shownUnits.enumerated() {
            let kids = children(unit, group: index)
            if unit.kind == .partner {
                let partner = tile(unit.partners[0], role: .partner, column: "r", group: "\(index)")
                let name = partner.content.title.spokenText
                groups.append(Group(
                    partner: partner,
                    children: kids.tiles,
                    childrenHidden: kids.hidden,
                    label: unit.children.isEmpty
                        ? L10n.Conclusions.familyGroupPartner(name)
                        : L10n.Conclusions.familyGroupPartnerAndChildren(name)
                ))
            } else {
                groups.append(Group(
                    partner: nil,
                    children: kids.tiles,
                    childrenHidden: kids.hidden,
                    label: L10n.string(L10n.Conclusions.familyCaptionChildren)
                ))
            }
        }
        self.groups = groups
        groupsHidden = allUnits.count - groups.count
        let anyChildren = allUnits.contains { !$0.children.isEmpty }
        if allUnits.isEmpty {
            rightCaption = nil
        } else {
            switch (partnerUnits.count, anyChildren) {
            case (0, _): rightCaption = L10n.string(L10n.Conclusions.familyCaptionChildren)
            case (1, true): rightCaption = L10n.string(L10n.Conclusions.familyCaptionPartnerAndChildren)
            case (1, false): rightCaption = L10n.string(L10n.Conclusions.familyCaptionPartner)
            case (_, true): rightCaption = L10n.string(L10n.Conclusions.familyCaptionPartnersAndChildren)
            case (_, false): rightCaption = L10n.string(L10n.Conclusions.familyCaptionPartners)
            }
        }
    }
}
