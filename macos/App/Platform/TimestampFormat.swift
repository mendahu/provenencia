import Foundation

/// Shared RFC 3339 → display formatting for catalog timestamps: parse once,
/// then `string(_:_:in:)` with a named `Style`. Text only; callers add styling.
///
/// Formatters are cached and lock-guarded: `ISO8601DateFormatter` is not
/// `Sendable`, so the static caches are `nonisolated(unsafe)` and all access
/// goes through `lock`.
enum TimestampFormat {
    private static let lock = NSLock()
    nonisolated(unsafe) private static let withFraction: ISO8601DateFormatter = {
        let formatter = ISO8601DateFormatter()
        formatter.formatOptions = [.withInternetDateTime, .withFractionalSeconds]
        return formatter
    }()
    nonisolated(unsafe) private static let plain: ISO8601DateFormatter = {
        let formatter = ISO8601DateFormatter()
        formatter.formatOptions = [.withInternetDateTime]
        return formatter
    }()

    static func parse(_ rfc3339: String) -> Date? {
        lock.lock()
        defer { lock.unlock() }
        return withFraction.date(from: rfc3339) ?? plain.date(from: rfc3339)
    }

    /// How a timestamp reads. Every style is one line of plain text.
    enum Style: Sendable, Equatable {
        /// `Oct 12, 2026`
        case shortDate
        /// `October 12, 2026`
        case longDate
        /// `Oct 12, 2026 at 9:00 AM`
        case dateTime
        /// `Oct 12, 2026 · 9:00 AM MDT`
        case stamp
        /// `Today, 3:15 PM`, `Yesterday`, else `shortDate`
        case relative
        /// `Today at 3:15 PM`, `Yesterday`, else `longDate` (VoiceOver)
        case spoken
    }

    /// What a style needs besides the timestamp. `calendar.timeZone` is the zone.
    struct Context: Sendable {
        var locale: Locale
        var calendar: Calendar
        var now: Date

        static var current: Context {
            Context(locale: .current, calendar: .current, now: Date())
        }
    }

    /// The one entry point for a catalog timestamp. `nil` when `rfc3339` is
    /// empty or unparseable, so the caller picks its own fallback.
    static func string(_ rfc3339: String, _ style: Style, in context: Context = .current) -> String? {
        parse(rfc3339).map { string($0, style, in: context) }
    }

    static func string(_ date: Date, _ style: Style, in context: Context = .current) -> String {
        let base = Date.FormatStyle(
            locale: context.locale, calendar: context.calendar, timeZone: context.calendar.timeZone
        )
        let shortDate = date.formatted(base.day().month(.abbreviated).year())
        let longDate = date.formatted(base.day().month(.wide).year())
        switch style {
        case .shortDate:
            return shortDate
        case .longDate:
            return longDate
        case .dateTime:
            return date.formatted(Date.FormatStyle(
                date: .abbreviated, time: .shortened,
                locale: context.locale, calendar: context.calendar, timeZone: context.calendar.timeZone
            ))
        case .stamp:
            return L10n.Timestamp.stamp(
                date: date.formatted(base.day(.twoDigits).month(.abbreviated).year()),
                time: date.formatted(base.hour().minute().timeZone(.specificName(.short)))
            )
        case .relative:
            switch day(of: date, in: context) {
            case .today: return L10n.Timestamp.today(time: date.formatted(base.hour().minute()))
            case .yesterday: return L10n.Timestamp.yesterday
            case .earlier: return shortDate
            }
        case .spoken:
            switch day(of: date, in: context) {
            case .today: return L10n.Timestamp.todayAt(time: date.formatted(base.hour().minute()))
            case .yesterday: return L10n.Timestamp.yesterday
            case .earlier: return longDate
            }
        }
    }

    /// Board stamp used on note rows: `04 Mar 2026 · 7:22 PM MST`.
    /// Removed with `LastUpdatedDisplay` in the next step.
    static func boardStamp(_ rfc3339: String) -> String {
        guard let date = parse(rfc3339) else { return rfc3339 }
        let datePart = date.formatted(.dateTime.day(.twoDigits).month(.abbreviated).year())
        let timePart = date.formatted(.dateTime.hour().minute().timeZone(.specificName(.short)))
        return "\(datePart) · \(timePart)"
    }

    private enum Day { case today, yesterday, earlier }

    private static func day(of date: Date, in context: Context) -> Day {
        let calendar = context.calendar
        if calendar.isDate(date, inSameDayAs: context.now) { return .today }
        if let yesterday = calendar.date(byAdding: .day, value: -1, to: context.now),
           calendar.isDate(date, inSameDayAs: yesterday) {
            return .yesterday
        }
        return .earlier
    }
}


/// One last-updated cell: the text a list row shows, its hover help, and its
/// spoken form. Day words for today and yesterday, a short date otherwise, and
/// a dash when the time is empty or unparseable.
struct LastUpdatedDisplay: Equatable {
    /// Width of the list column that holds the cell, icon included.
    static let columnWidth: CGFloat = 136

    let text: String
    /// The full stamp; `nil` for the dash.
    let help: String?
    let spoken: String
    /// `false` for the dash (no time recorded).
    let isKnown: Bool

    init(_ rfc3339: String, now: Date, calendar: Calendar = .current, locale: Locale = .current) {
        guard let date = TimestampFormat.parse(rfc3339) else {
            text = "—"
            help = nil
            spoken = L10n.LastUpdated.a11yUnknown
            isKnown = false
            return
        }
        let style = Date.FormatStyle(locale: locale, calendar: calendar, timeZone: calendar.timeZone)
        if calendar.isDate(date, inSameDayAs: now) {
            text = L10n.LastUpdated.today(time: date.formatted(style.hour().minute()))
        } else if let yesterday = calendar.date(byAdding: .day, value: -1, to: now),
                  calendar.isDate(date, inSameDayAs: yesterday) {
            text = L10n.LastUpdated.yesterday
        } else {
            text = date.formatted(style.day().month(.abbreviated).year())
        }
        help = L10n.LastUpdated.help(stamp: TimestampFormat.boardStamp(rfc3339))
        spoken = L10n.LastUpdated.a11yUpdated(text)
        isKnown = true
    }
}
