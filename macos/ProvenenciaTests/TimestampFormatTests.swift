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

    @Test func boardStampFallsBackToRawWhenUnparseable() {
        #expect(TimestampFormat.boardStamp("garbage") == "garbage")
    }

    @Test func boardStampIncludesDateTimeAndTimezoneSeparator() {
        let stamp = TimestampFormat.boardStamp("2026-03-04T19:22:00Z")
        #expect(stamp.contains("·"))
        #expect(stamp.contains("2026"))
    }

    @Test func abbreviatedDateTimeFallsBackToRawWhenUnparseable() {
        #expect(TimestampFormat.abbreviatedDateTime("nope") == "nope")
    }

    private static let edmonton: Calendar = {
        var calendar = Calendar(identifier: .gregorian)
        calendar.timeZone = TimeZone(identifier: "America/Edmonton")!
        return calendar
    }()

    /// 2026-10-14 18:00 in Edmonton.
    private static let now = TimestampFormat.parse("2026-10-15T00:00:00Z")!

    private func display(_ rfc3339: String, locale: String = "en_US") -> LastUpdatedDisplay {
        LastUpdatedDisplay(rfc3339, now: Self.now, calendar: Self.edmonton, locale: Locale(identifier: locale))
    }

    /// Newer macOS puts a narrow no-break space before AM/PM.
    private func plain(_ text: String) -> String {
        text.replacingOccurrences(of: "\u{202F}", with: " ")
    }

    @Test func lastUpdatedNamesTodayYesterdayThenDate() {
        #expect(plain(display("2026-10-14T21:15:00Z").text) == "Today, 3:15 PM")
        #expect(display("2026-10-13T22:40:00Z").text == "Yesterday")
        #expect(display("2026-10-12T15:00:00Z").text == "Oct 12, 2026")
        #expect(display("2026-10-12T15:00:00Z", locale: "en_GB").text == "12 Oct 2026")
    }

    @Test func lastUpdatedUnknownIsDashWithoutHelp() {
        for raw in ["", "garbage"] {
            let unknown = display(raw)
            #expect(unknown.text == "—")
            #expect(unknown.help == nil)
            #expect(unknown.spoken == "last updated unknown")
            #expect(!unknown.isKnown)
        }
        let known = display("2026-10-13T22:40:00Z")
        #expect(known.isKnown)
        #expect(known.help?.hasPrefix("Last updated ") == true)
        #expect(known.spoken == "updated Yesterday")
    }
}
