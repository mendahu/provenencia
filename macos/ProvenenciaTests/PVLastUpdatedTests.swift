import Foundation
import Testing
@testable import Provenencia

@MainActor
struct PVLastUpdatedTests {
    private static let edmonton: TimestampFormat.Context = {
        var calendar = Calendar(identifier: .gregorian)
        calendar.timeZone = TimeZone(identifier: "America/Edmonton")!
        // 2026-10-14 18:00 in Edmonton.
        return TimestampFormat.Context(
            locale: Locale(identifier: "en_US"), calendar: calendar, now: TimestampFormat.parse("2026-10-15T00:00:00Z")!
        )
    }()

    @Test func rowLabelEndsWithUpdated() {
        let base = "Robins, James, PER-7KD45, b. 1817"
        func label(_ updatedAt: String) -> String {
            PVLastUpdated.label(base, updatedAt: updatedAt, in: Self.edmonton)
        }
        #expect(label("2026-10-13T22:40:00Z") == "\(base), updated Yesterday")
        #expect(label("2026-10-12T15:00:00Z") == "\(base), updated October 12, 2026")
        #expect(label("") == "\(base), last updated unknown")
    }

    @Test func updatedDetailComposesTheCell() {
        let known = PVLastUpdated.detail("2026-10-14T21:15:00Z", in: Self.edmonton)
        #expect(known.text.replacingOccurrences(of: "\u{202F}", with: " ") == "Today, 3:15 PM")
        #expect(known.help?.hasPrefix("Last updated ") == true)
        #expect(!known.faint)
        for raw in ["", "garbage"] {
            let unknown = PVLastUpdated.detail(raw, in: Self.edmonton)
            #expect(unknown.text == "—")
            #expect(unknown.help == nil)
            #expect(unknown.faint)
        }
    }

    @Test func sourceRowLabelEndsWithUpdated() {
        let base = "1851 census, SRC-7KQ2D"
        #expect(
            PVLastUpdated.label(base, updatedAt: "2026-10-13T22:40:00Z", in: Self.edmonton)
                == "1851 census, SRC-7KQ2D, updated Yesterday"
        )
        #expect(
            PVLastUpdated.label(base, updatedAt: "", in: Self.edmonton)
                == "1851 census, SRC-7KQ2D, last updated unknown"
        )
    }
}
