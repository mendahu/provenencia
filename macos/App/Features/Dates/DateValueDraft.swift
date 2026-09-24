import Foundation

/// Editable draft of a genealogical DateValue for the shared Structure/Edit date modal.
/// Validation mirrors `core/database/datevalues` (gaps allowed; at least one civil field or a phrase).
struct DateValueDraft: Equatable, Sendable {
    var kind: String
    var qualifier: String
    var calendar: String
    var phrase: String

    var startYear: Int32?
    var startMonth: Int32?
    var startDay: Int32?
    var startHour: Int32?
    var startMinute: Int32?
    var startSecond: Int32?
    var startMillisecond: Int32?
    var startTZ: String

    var endYear: Int32?
    var endMonth: Int32?
    var endDay: Int32?
    var endHour: Int32?
    var endMinute: Int32?
    var endSecond: Int32?
    var endMillisecond: Int32?
    var endTZ: String

    /// UI: clock fields revealed after “Add time”.
    var showStartTime: Bool
    var showEndTime: Bool
    /// UI: calendar + phrase section.
    var showAdvanced: Bool

    static func empty() -> DateValueDraft {
        DateValueDraft(
            kind: "point",
            qualifier: "",
            calendar: "gregorian",
            phrase: "",
            startYear: nil,
            startMonth: nil,
            startDay: nil,
            startHour: nil,
            startMinute: nil,
            startSecond: nil,
            startMillisecond: nil,
            startTZ: "",
            endYear: nil,
            endMonth: nil,
            endDay: nil,
            endHour: nil,
            endMinute: nil,
            endSecond: nil,
            endMillisecond: nil,
            endTZ: "",
            showStartTime: false,
            showEndTime: false,
            showAdvanced: false
        )
    }

    // MARK: Cascade UI helpers

    var canSetStartMonth: Bool { startYear != nil }
    var canSetStartDay: Bool { startMonth != nil }
    var canSetStartHour: Bool { startDay != nil }
    var canSetStartMinute: Bool { startHour != nil }
    var canSetStartSecond: Bool { startMinute != nil }
    var canSetStartMillisecond: Bool { startSecond != nil }

    var canSetEndMonth: Bool { endYear != nil }
    var canSetEndDay: Bool { endMonth != nil }
    var canSetEndHour: Bool { endDay != nil }
    var canSetEndMinute: Bool { endHour != nil }
    var canSetEndSecond: Bool { endMinute != nil }
    var canSetEndMillisecond: Bool { endSecond != nil }

    var isPoint: Bool { kind == "point" }
    var isRange: Bool { kind == "range" }

    /// Day is set on the start cascade — time fields may be revealed.
    var hasFullStartDay: Bool { startYear != nil && startMonth != nil && startDay != nil }
    /// Day is set on the end cascade — time fields may be revealed.
    var hasFullEndDay: Bool { endYear != nil && endMonth != nil && endDay != nil }

    /// Bindings still call this after each edit; components are independent
    /// so there is nothing to wipe.
    mutating func applyStartCascade() {}

    mutating func applyEndCascade() {}

    /// Switching to range clears qualifier (Between is the hedge). Switching to point clears end_*.
    mutating func setKind(_ newKind: String) {
        kind = newKind
        if newKind == "range" {
            qualifier = ""
        } else {
            endYear = nil
            endMonth = nil
            endDay = nil
            endHour = nil
            endMinute = nil
            endSecond = nil
            endMillisecond = nil
            endTZ = ""
            showEndTime = false
        }
    }

    // MARK: Validation

    /// Civil component that can show an out-of-range field error.
    enum CascadeComponent: Equatable, Sendable {
        case year, month, day, hour, minute, second, millisecond
    }

    /// Inline range-order message when latest falls before earliest; otherwise nil.
    var endError: String? {
        guard kind == "range" else { return nil }
        guard startSideHasYear, endSideHasYear else { return nil }
        guard !Self.sideLessOrEqual(startSide, endSide) else { return nil }
        return String(localized: L10n.Sources.dateRangeOrderError)
    }

    /// Out-of-range message for one cascade field, if its stored value is invalid.
    func fieldError(_ component: CascadeComponent, start: Bool) -> String? {
        let side = start ? startSide : endSide
        switch component {
        case .year:
            guard let y = side.year else { return nil }
            return (1...9999).contains(y) ? nil : String(localized: L10n.Sources.dateYearOutOfRange)
        case .month:
            guard let m = side.month else { return nil }
            return (1...12).contains(m) ? nil : String(localized: L10n.Sources.dateMonthOutOfRange)
        case .day:
            guard let d = side.day else { return nil }
            return (1...31).contains(d) ? nil : String(localized: L10n.Sources.dateDayOutOfRange)
        case .hour:
            guard let h = side.hour else { return nil }
            return (0...23).contains(h) ? nil : String(localized: L10n.Sources.dateHourOutOfRange)
        case .minute:
            guard let mi = side.minute else { return nil }
            return (0...59).contains(mi) ? nil : String(localized: L10n.Sources.dateMinuteOutOfRange)
        case .second:
            guard let s = side.second else { return nil }
            return (0...59).contains(s) ? nil : String(localized: L10n.Sources.dateSecondOutOfRange)
        case .millisecond:
            guard let ms = side.millisecond else { return nil }
            return (0...999).contains(ms) ? nil : String(localized: L10n.Sources.dateMillisecondOutOfRange)
        }
    }

    func isFieldInvalid(_ component: CascadeComponent, start: Bool) -> Bool {
        fieldError(component, start: start) != nil
    }

    /// First out-of-range message on a side (Y/M/D then time), for group-level display.
    func cascadeFieldError(start: Bool) -> String? {
        let order: [CascadeComponent] = [.year, .month, .day, .hour, .minute, .second, .millisecond]
        for component in order {
            if let message = fieldError(component, start: start) {
                return message
            }
        }
        return nil
    }

    var isValid: Bool {
        let q = qualifier.trimmingCharacters(in: .whitespacesAndNewlines)
        let p = phrase.trimmingCharacters(in: .whitespacesAndNewlines)
        let startTZTrim = startTZ.trimmingCharacters(in: .whitespacesAndNewlines)
        let endTZTrim = endTZ.trimmingCharacters(in: .whitespacesAndNewlines)

        switch kind {
        case "point":
            guard Self.pointQualifierOK(q) else { return false }
            if !endSide.empty || !endTZTrim.isEmpty { return false }
            if !startSide.hasCivil {
                // Phrase-only point: no year/month/day/time; timezone alone is not enough.
                return !p.isEmpty && startTZTrim.isEmpty && startMillisecond == nil
            }
            return Self.validateComponentRanges(startSide)
        case "range":
            guard q.isEmpty else { return false }
            guard Self.validateComponentRanges(startSide) else { return false }
            guard Self.validateComponentRanges(endSide) else { return false }
            return Self.sideLessOrEqual(startSide, endSide)
        default:
            return false
        }
    }

    func toInput() -> CatalogDateValueInput {
        CatalogDateValueInput(
            kind: kind,
            qualifier: kind == "range" ? "" : qualifier.trimmingCharacters(in: .whitespacesAndNewlines),
            calendar: calendar.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty
                ? "gregorian"
                : calendar.trimmingCharacters(in: .whitespacesAndNewlines),
            startYear: startYear,
            startMonth: startMonth,
            startDay: startDay,
            startHour: startHour,
            startMinute: startMinute,
            startSecond: startSecond,
            startMillisecond: startMillisecond,
            startTZ: startTZ.trimmingCharacters(in: .whitespacesAndNewlines),
            endYear: kind == "range" ? endYear : nil,
            endMonth: kind == "range" ? endMonth : nil,
            endDay: kind == "range" ? endDay : nil,
            endHour: kind == "range" ? endHour : nil,
            endMinute: kind == "range" ? endMinute : nil,
            endSecond: kind == "range" ? endSecond : nil,
            endMillisecond: kind == "range" ? endMillisecond : nil,
            endTZ: kind == "range" ? endTZ.trimmingCharacters(in: .whitespacesAndNewlines) : "",
            phrase: phrase.trimmingCharacters(in: .whitespacesAndNewlines)
        )
    }

    // MARK: Private sides

    private var startSide: Side {
        Side(
            year: startYear.map(Int.init),
            month: startMonth.map(Int.init),
            day: startDay.map(Int.init),
            hour: startHour.map(Int.init),
            minute: startMinute.map(Int.init),
            second: startSecond.map(Int.init),
            millisecond: startMillisecond.map(Int.init)
        )
    }

    private var endSide: Side {
        Side(
            year: endYear.map(Int.init),
            month: endMonth.map(Int.init),
            day: endDay.map(Int.init),
            hour: endHour.map(Int.init),
            minute: endMinute.map(Int.init),
            second: endSecond.map(Int.init),
            millisecond: endMillisecond.map(Int.init)
        )
    }

    private var startSideHasYear: Bool { startYear != nil }
    private var endSideHasYear: Bool { endYear != nil }

    private struct Side {
        var year, month, day, hour, minute, second, millisecond: Int?

        var empty: Bool {
            year == nil && month == nil && day == nil
                && hour == nil && minute == nil && second == nil && millisecond == nil
        }

        var hasCivil: Bool {
            year != nil || month != nil || day != nil
                || hour != nil || minute != nil || second != nil
        }
    }

    private static func pointQualifierOK(_ qual: String) -> Bool {
        switch qual {
        case "", "ABT", "BEF", "AFT": return true
        default: return false
        }
    }

    private static func validateComponentRanges(_ s: Side) -> Bool {
        guard s.hasCivil else { return false }
        if let m = s.month, !(1...12).contains(m) { return false }
        if let d = s.day, !(1...31).contains(d) { return false }
        if let y = s.year, !(1...9999).contains(y) { return false }
        if let h = s.hour, !(0...23).contains(h) { return false }
        if let mi = s.minute, !(0...59).contains(mi) { return false }
        if let sec = s.second, !(0...59).contains(sec) { return false }
        if let ms = s.millisecond, !(0...999).contains(ms) { return false }
        return true
    }

    private static func sideLessOrEqual(_ a: Side, _ b: Side) -> Bool {
        let levelsA: [Int?] = [a.year, a.month, a.day, a.hour, a.minute, a.second, a.millisecond]
        let levelsB: [Int?] = [b.year, b.month, b.day, b.hour, b.minute, b.second, b.millisecond]
        for i in levelsA.indices {
            guard let va = levelsA[i], let vb = levelsB[i] else { return true }
            if va < vb { return true }
            if va > vb { return false }
        }
        return true
    }
}

extension DateValueDraft {
    /// Rebuild a draft from a previously saved `CatalogDateValueInput` (session cache / FakeStore).
    init(from input: CatalogDateValueInput) {
        self.init(
            kind: input.kind.isEmpty ? "point" : input.kind,
            qualifier: input.qualifier,
            calendar: input.calendar.isEmpty ? "gregorian" : input.calendar,
            phrase: input.phrase,
            startYear: input.startYear,
            startMonth: input.startMonth,
            startDay: input.startDay,
            startHour: input.startHour,
            startMinute: input.startMinute,
            startSecond: input.startSecond,
            startMillisecond: input.startMillisecond,
            startTZ: input.startTZ,
            endYear: input.endYear,
            endMonth: input.endMonth,
            endDay: input.endDay,
            endHour: input.endHour,
            endMinute: input.endMinute,
            endSecond: input.endSecond,
            endMillisecond: input.endMillisecond,
            endTZ: input.endTZ,
            showStartTime: input.startHour != nil,
            showEndTime: input.endHour != nil,
            showAdvanced: !input.phrase.isEmpty
                || (!input.calendar.isEmpty && input.calendar != "gregorian")
        )
    }
}
