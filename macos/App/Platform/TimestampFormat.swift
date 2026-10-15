import Foundation

/// Shared RFC 3339 → display formatting for catalog timestamps.
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

    /// Board stamp used on note rows: `04 Mar 2026 · 7:22 PM MST`.
    static func boardStamp(_ rfc3339: String) -> String {
        guard let date = parse(rfc3339) else { return rfc3339 }
        let datePart = date.formatted(.dateTime.day(.twoDigits).month(.abbreviated).year())
        let timePart = date.formatted(.dateTime.hour().minute().timeZone(.specificName(.short)))
        return "\(datePart) · \(timePart)"
    }

    /// Compact absolute date+time (onboarding project meta, etc.).
    static func abbreviatedDateTime(_ rfc3339: String) -> String {
        guard let date = parse(rfc3339) else { return rfc3339 }
        return date.formatted(date: .abbreviated, time: .shortened)
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
