import SwiftUI

/// What one person tile says (S13-12, board S13-D3 frame 11l): the title, the
/// *b.* and *d.* lines with a missing half left out, the likeness and the
/// gender colors, and the one sentence VoiceOver reads. Pure; no Family
/// vocabulary: the caller passes the role's label, or nil for none.
struct PersonTileContent: Equatable {
    /// The person's entity id.
    var id: String
    var title: ConclusionTitleSource
    var lines: [PersonLifeDisplay.LifeLine]
    var imageRelPath: String
    var colors: PVThumbnail.Colors?
    var role: String?
    var isFocus: Bool
    var spokenLabel: String

    init(
        summary: CatalogPersonSummary,
        role: String?,
        isFocus: Bool,
        locale: Locale = .autoupdatingCurrent,
        orders: NameFormatOrders = NameFormatOrders()
    ) {
        let title = PersonTitleDisplay.titleSource(summary, style: .natural, orders: orders)
        let lines = PersonLifeDisplay.lifeLines(summary, locale: locale, missing: .omitted)
        id = summary.entity.id
        self.title = title
        self.lines = lines
        imageRelPath = summary.pictureImageRelPath
        colors = PersonGenderTreatment.read(summary)?.thumbnailColors
        self.role = role
        self.isFocus = isFocus
        let join: (String, String) -> String = L10n.DesignSystem.spokenJoin
        let life = L10n.DesignSystem.spokenJoin(lines.map(\.accessibilityLabel))
        let named = life.map { join(title.spokenText, $0) } ?? title.spokenText
        if isFocus {
            spokenLabel = join(named, L10n.string(L10n.Conclusions.familyThisPersonSpoken))
        } else if let role {
            spokenLabel = join(role, named)
        } else {
            spokenLabel = named
        }
    }
}

/// One person as a card: likeness, name, and their *b.* and *d.* lines. A
/// relative is a button (`open`); the focus is a raised card, not a button.
struct PersonTile: View {
    let content: PersonTileContent
    let projectDir: String
    /// Opens the person; ignored for the focus.
    var open: () -> Void = {}

    static let height: CGFloat = 92

    var body: some View {
        if content.isFocus {
            card
                .accessibilityElement(children: .ignore)
                .accessibilityLabel(content.spokenLabel)
                .help(content.title.text)
        } else {
            Button(action: open) { card }
                .buttonStyle(PersonTileButtonStyle())
                .pvLiftHover(scale: 1.025)
                .accessibilityLabel(content.spokenLabel)
                .help(content.title.text)
        }
    }

    private var card: some View {
        PVCard(
            tone: content.isFocus ? .raised : .card,
            cornerRadius: PVRadius.sm,
            elevated: content.isFocus,
            padding: 10
        ) {
            HStack(spacing: 10) {
                CachedThumbnail(
                    projectDir: projectDir,
                    relPath: content.imageRelPath,
                    size: 70,
                    cornerRadius: PVRadius.sm,
                    fallbackMark: .subjectPerson,
                    fallbackLabel: L10n.Conclusions.personNoLikeness,
                    colors: content.colors
                )
                VStack(alignment: .leading, spacing: 3) {
                    HStack(spacing: PVSpacing.space4) {
                        ConclusionListRow.title(content.title, lineLimit: 2, emphasized: content.isFocus)
                            .font(PVListText.titleFont)
                            .foregroundStyle(PVColor.textPrimary)
                        if content.isFocus {
                            PVBadge(L10n.Conclusions.familyThisPerson)
                                .layoutPriority(1)
                        }
                    }
                    ForEach(content.lines, id: \.abbreviation) { line in
                        ConclusionDatePlaceLine(
                            leading: line.abbreviation,
                            date: line.date,
                            place: line.place,
                            accessibilityLabel: line.accessibilityLabel,
                            missing: .omitted
                        )
                    }
                }
                .frame(maxWidth: .infinity, alignment: .leading)
            }
            .frame(maxWidth: .infinity, minHeight: 70, maxHeight: 70, alignment: .leading)
        }
        .frame(height: Self.height)
        .clipped()
    }
}

/// A relative's tile as a button: the focus ring and nothing else; the hover
/// lift is `.pvLiftHover`.
private struct PersonTileButtonStyle: ButtonStyle {
    func makeBody(configuration: Configuration) -> some View {
        PersonTileButtonBody(configuration: configuration)
    }
}

private struct PersonTileButtonBody: View {
    let configuration: ButtonStyleConfiguration

    /// Read here, in the button's focus scope (see `PVButtonBody`).
    @Environment(\.isFocused) private var isFocused

    var body: some View {
        configuration.label
            .contentShape(Rectangle())
            .pvFocusRing(isFocused, cornerRadius: PVRadius.sm)
    }
}
