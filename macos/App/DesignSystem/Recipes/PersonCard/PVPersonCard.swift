import SwiftUI

/// What one person card says (S13-12, board S13-D3 frame 11l): the title, the
/// *b.* and *d.* lines with a missing half left out, the life span in years,
/// the likeness and the gender colors, and the one sentence VoiceOver reads.
/// Pure; no Family vocabulary: the caller passes the role's label, or nil for
/// none.
struct PVPersonCardContent: Equatable {
    /// The person's entity id.
    var id: String
    var title: PVTitleSource
    var lines: [PersonLifeDisplay.LifeLine]
    /// The life span at year precision, for the compact size ("1790 – 1851");
    /// empty when there is none.
    var years: String
    var imageRelPath: String
    var colors: PVThumbnail.Colors?
    var role: String?
    var isFocus: Bool
    var spokenLabel: String
    /// The compact card's tooltip: the title, then each life line in full.
    var helpText: String

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
        years = PersonLifeDisplay.years(summary, locale: locale)
        helpText = ([title.text] + lines.map(\.text) + [role].compactMap { $0 }).joined(separator: "\n")
        imageRelPath = summary.pictureImageRelPath
        colors = PVGenderTreatment.read(summary)?.thumbnailColors
        self.role = role
        self.isFocus = isFocus
        let join: (String, String) -> String = L10n.DesignSystem.spokenJoin
        let life = L10n.DesignSystem.spokenJoin(lines.map(\.accessibilityLabel))
        let named = life.map { join(title.spokenText, $0) } ?? title.spokenText
        if isFocus {
            spokenLabel = join(named, L10n.string(L10n.PersonCard.thisPersonSpoken))
        } else if let role {
            spokenLabel = join(role, named)
        } else {
            spokenLabel = named
        }
    }
}

/// One person as a card: likeness, name, and their *b.* and *d.* lines (the
/// life span in years and the role at the compact size). A relative is a button (`open`);
/// the focus is a raised card, not a button.
struct PVPersonCard: View {
    let content: PVPersonCardContent
    let projectDir: String
    var size: PVTileCardSize = .regular
    /// Opens the person; ignored for the focus.
    var open: () -> Void = {}

    /// The AppKit face of the title at `size`, so a measure matches the paint. Regular
    /// weight at both sizes: Spectral ships an italic at that weight, which a former
    /// surname's italic run needs.
    static func nsTitleFont(_ size: PVTileCardSize) -> NSFont {
        PVFont.nsBody(size: size == .compact ? PVTypeScale.caption : PVTypeScale.body)
    }

    static func titleFont(_ size: PVTileCardSize) -> Font { Font(nsTitleFont(size)) }

    /// The AppKit face of the compact size's life span.
    static let nsCompactYearsFont = PVFont.nsMono(size: PVTypeScale.micro)

    var body: some View {
        if content.isFocus && size == .regular {
            PVTileCard(
                size: size,
                isRaised: content.isFocus,
                accessibilityLabel: content.spokenLabel,
                help: help,
                action: nil,
                thumbnail: { thumbnail },
                title: { title },
                lines: { lines },
                trailing: { PVBadge(L10n.PersonCard.thisPerson) }
            )
        } else {
            PVTileCard(
                size: size,
                isRaised: content.isFocus,
                accessibilityLabel: content.spokenLabel,
                help: help,
                action: content.isFocus ? nil : open,
                thumbnail: { thumbnail },
                title: { title },
                lines: { lines }
            )
        }
    }

    private var help: String {
        size == .compact ? content.helpText : content.title.text
    }

    private var thumbnail: some View {
        CachedThumbnail(
            projectDir: projectDir,
            relPath: content.imageRelPath,
            size: size.thumbnail,
            cornerRadius: PVRadius.sm,
            fallbackMark: .subjectPerson,
            fallbackLabel: L10n.PersonCard.noLikeness,
            colors: content.colors
        )
    }

    private var title: some View {
        PVTitleText(content.title, lineLimit: size.titleLineLimit, emphasized: content.isFocus)
            .font(Self.titleFont(size))
            .foregroundStyle(PVColor.textPrimary)
    }

    @ViewBuilder
    private var lines: some View {
        if size == .compact {
            HStack(spacing: 12) {
                if !content.years.isEmpty {
                    Text(verbatim: content.years)
                        .font(Font(Self.nsCompactYearsFont))
                        .foregroundStyle(PVColor.textMuted)
                        .lineLimit(size.linesLineLimit)
                }
                if let role = content.role {
                    Spacer(minLength: 0)
                    Text(verbatim: role)
                        .pvMicroCaps()
                        .foregroundStyle(PVColor.textMuted)
                        .lineLimit(1)
                }
            }
        } else {
            ForEach(content.lines, id: \.abbreviation) { line in
                PVDatePlaceLine(
                    leading: line.abbreviation,
                    date: line.date,
                    place: line.place,
                    accessibilityLabel: line.accessibilityLabel,
                    missing: .omitted
                )
            }
        }
    }
}

#if DEBUG
private enum PVPersonCardPreviewData {
    static func place(_ name: String) -> CatalogChainedPlace {
        CatalogChainedPlace(
            entity: CatalogCanonicalEntity(id: "pl-\(name)", ref: "PLC-\(name)", subjectTypeID: "t", label: ""),
            names: [name]
        )
    }

    static var anna: CatalogPersonSummary {
        var summary = CatalogPersonSummary(
            entity: CatalogCanonicalEntity(id: "anna", ref: "PER-7KD45", subjectTypeID: "t", label: ""),
            name: CatalogNameValue(form: "Anna Weber"),
            nameValueCount: 1,
            birth: CatalogLifeFacts(
                date: CatalogDateValueInput(kind: "point", startYear: 1790, startMonth: 5, startDay: 3),
                dateCount: 1,
                places: [place("Bremen")]
            ),
            death: CatalogLifeFacts(
                date: CatalogDateValueInput(kind: "point", startYear: 1851),
                dateCount: 1
            )
        )
        summary.gender = CatalogTermFacts(
            term: .term(id: "t-woman", key: "woman", label: "Woman", origin: CatalogOrigin.provenencia),
            count: 1, propertyLabel: "Gender"
        )
        return summary
    }
}

#Preview("Regular: relative and focus") {
    VStack(spacing: 12) {
        PVPersonCard(
            content: PVPersonCardContent(summary: PVPersonCardPreviewData.anna, role: "Mother", isFocus: false),
            projectDir: ""
        )
        PVPersonCard(
            content: PVPersonCardContent(summary: PVPersonCardPreviewData.anna, role: nil, isFocus: true),
            projectDir: ""
        )
    }
    .frame(width: 260)
    .padding()
}

#Preview("Compact: relative and focus") {
    VStack(spacing: 12) {
        PVPersonCard(
            content: PVPersonCardContent(summary: PVPersonCardPreviewData.anna, role: "Mother", isFocus: false),
            projectDir: "",
            size: .compact
        )
        PVPersonCard(
            content: PVPersonCardContent(summary: PVPersonCardPreviewData.anna, role: nil, isFocus: true),
            projectDir: "",
            size: .compact
        )
    }
    .frame(width: 260)
    .padding()
}
#endif
