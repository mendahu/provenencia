import Foundation
import Testing
@testable import Provenencia

@Suite
struct TimestampFormatTests {
    @Test func parseAcceptsFractionalAndPlainRFC3339() {
        #expect(TimestampFormat.parse("2026-03-04T19:22:00.123Z") != nil)
        #expect(TimestampFormat.parse("2026-03-04T19:22:00Z") != nil)
        #expect(TimestampFormat.parse("not-a-date") == nil)
    }

    private static let edmonton: Calendar = {
        var calendar = Calendar(identifier: .gregorian)
        calendar.timeZone = TimeZone(identifier: "America/Edmonton")!
        return calendar
    }()

    /// 2026-10-14 18:00 in Edmonton.
    private static let now = TimestampFormat.parse("2026-10-15T00:00:00Z")!

    /// Newer macOS puts a narrow no-break space before AM/PM.
    private func plain(_ text: String) -> String {
        text.replacingOccurrences(of: "\u{202F}", with: " ")
    }

    private func ctx(_ locale: String = "en_US") -> TimestampFormat.Context {
        TimestampFormat.Context(locale: Locale(identifier: locale), calendar: Self.edmonton, now: Self.now)
    }

    /// Every style over fixed inputs. `T` is 2026-10-12 09:00 in Edmonton.
    @Test(arguments: [
        ("2026-10-12T15:00:00Z", TimestampFormat.Style.shortDate, "en_US", "Oct 12, 2026"),
        ("2026-10-12T15:00:00Z", .shortDate, "en_GB", "12 Oct 2026"),
        ("2026-10-12T15:00:00Z", .longDate, "en_US", "October 12, 2026"),
        ("2026-10-12T15:00:00Z", .dateTime, "en_US", "Oct 12, 2026 at 9:00 AM"),
        ("2026-10-12T15:00:00Z", .stamp, "en_US", "Oct 12, 2026 · 9:00 AM MDT"),
        ("2026-10-12T15:00:00Z", .stamp, "en_GB", "12 Oct 2026 · 09:00 GMT-6"),
        ("2026-10-14T21:15:00Z", .relative, "en_US", "Today, 3:15 PM"),
        ("2026-10-13T22:40:00Z", .relative, "en_US", "Yesterday"),
        ("2026-10-12T15:00:00Z", .relative, "en_US", "Oct 12, 2026"),
        ("2026-10-12T15:00:00Z", .relative, "en_GB", "12 Oct 2026"),
        ("2026-10-14T21:15:00Z", .spoken, "en_US", "Today at 3:15 PM"),
        ("2026-10-13T22:40:00Z", .spoken, "en_US", "Yesterday"),
        ("2026-10-12T15:00:00Z", .spoken, "en_US", "October 12, 2026"),
    ])
    func stylesFormatOneLine(raw: String, style: TimestampFormat.Style, locale: String, expected: String) {
        #expect(plain(TimestampFormat.string(raw, style, in: ctx(locale)) ?? "nil") == expected)
    }

    @Test(arguments: ["", "garbage"])
    func unparseableIsNilInEveryStyle(raw: String) {
        for style in [TimestampFormat.Style.shortDate, .longDate, .dateTime, .stamp, .relative, .spoken] {
            #expect(TimestampFormat.string(raw, style, in: ctx()) == nil)
        }
    }
}
