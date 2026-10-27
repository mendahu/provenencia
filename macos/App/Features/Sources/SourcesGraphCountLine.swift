import SwiftUI

/// An Evidence graph's counts, the one view that draws a `SourcesGraphCounts`:
/// "12 subjects · 40 observations" with the digits in mono, "0 subjects · not
/// started" (the card) or "0 subjects · 0 observations" (the button, whose badge
/// says "not started"), a short rule while the counts load, or a mark beside
/// "Counts could not load" when they failed. The Source page's graph cards draw
/// it as `.card`; a Sources list graph button as `.button`.
struct SourcesGraphCountLine: View {
    enum Style {
        case card
        case button
    }

    let counts: SourcesGraphCounts
    let style: Style

    private var notStarted: L10n.Sources.NotStartedReading {
        style == .button ? .zeroCounts : .words
    }

    var body: some View {
        content
            .accessibilityElement(children: .ignore)
            .accessibilityLabel(Text(verbatim: L10n.Sources.graphCounts(counts, style: .spoken, notStarted: notStarted)))
    }

    @ViewBuilder
    private var content: some View {
        switch counts {
        case .loading:
            switch style {
            case .card:
                PVSkeletonBar(width: nil, height: 1)
                    .padding(.trailing, PVSpacing.space3)
                    .padding(.vertical, 5)
            case .button:
                PVSkeletonBar(width: 132, height: 6)
                    .padding(.vertical, PVSpacing.space1)
            }
        case .unavailable(let message):
            HStack(spacing: PVSpacing.space2) {
                PVIcon(.circleAlert, size: 11)
                    .foregroundStyle(PVColor.danger)
                text
            }
            .help(message)
        case .notStarted, .counts:
            text
        }
    }

    private var text: some View {
        Text(Self.countsText(counts, style: style))
            .foregroundStyle(PVColor.textMuted)
            .lineLimit(1)
            .truncationMode(.tail)
    }

    /// The count line as styled text: micro body throughout, mono on each run of
    /// ASCII digits, italic on "not started". The words come from the one
    /// localized line, so a translation keeps its order.
    static func countsText(_ counts: SourcesGraphCounts, style: Style) -> AttributedString {
        let line = L10n.Sources.graphCounts(counts, style: .full, notStarted: style == .button ? .zeroCounts : .words)
        let body = PVFont.body(size: PVTypeScale.micro)
        let mono = PVFont.mono(size: PVTypeScale.micro)
        var out = AttributedString()
        var run = ""
        var runIsDigits = false
        func flush() {
            guard !run.isEmpty else { return }
            var piece = AttributedString(run)
            piece.font = runIsDigits ? mono : body
            out.append(piece)
            run = ""
        }
        for character in line {
            let isDigit = character.isASCII && character.isNumber
            if isDigit != runIsDigits { flush() }
            runIsDigits = isDigit
            run.append(character)
        }
        flush()
        let notStartedWords = L10n.string(L10n.Sources.graphNotStarted)
        if let range = out.range(of: notStartedWords) {
            out[range].font = body.italic()
        }
        return out
    }
}
