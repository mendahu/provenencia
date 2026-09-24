import SwiftUI

/// Shared DateValue create/edit form body (kind, qualifier, cascade, time, phrase).
struct DateValueEditorForm: View {
    @Binding var draft: DateValueDraft
    /// Prefix for control identifiers (default matches the Source page date dialog).
    var accessibilityIdentifierPrefix: String = "sources.page.date"

    var body: some View {
        VStack(alignment: .leading, spacing: PVSpacing.space7) {
            kindSection
            if draft.isPoint {
                qualifierSection
            }
            PVDivider()
            cascadeSection(start: true)
            if draft.isRange {
                PVDivider()
                cascadeSection(start: false)
            }
            PVDivider()
            advancedToggle
            if draft.showAdvanced {
                advancedSection
            }
        }
    }

    private var kindSection: some View {
        VStack(alignment: .leading, spacing: PVSpacing.space3) {
            Text(L10n.Sources.dateKindLabel)
                .font(PVFont.body(size: PVTypeScale.caption, weight: PVFontWeight.medium))
                .foregroundStyle(PVColor.textSecondary)
            PVChipGroup(style: .segmented) {
                PVChip(
                    L10n.Sources.dateKindPoint,
                    isSelected: draft.kind == "point",
                    expands: true,
                    selectionLift: true,
                    action: { draft.setKind("point") }
                )
                .accessibilityIdentifier("\(accessibilityIdentifierPrefix).kind.point")
                PVChip(
                    L10n.Sources.dateKindRange,
                    isSelected: draft.kind == "range",
                    expands: true,
                    selectionLift: true,
                    action: { draft.setKind("range") }
                )
                .accessibilityIdentifier("\(accessibilityIdentifierPrefix).kind.range")
            }
            if draft.isRange {
                Text(L10n.Sources.dateKindRangeHint)
                    .font(PVFont.body(size: PVTypeScale.micro, italic: true))
                    .foregroundStyle(PVColor.textMuted)
            }
        }
    }

    private var qualifierSection: some View {
        VStack(alignment: .leading, spacing: PVSpacing.space3) {
            Text(L10n.Sources.dateQualifierLabel)
                .font(PVFont.body(size: PVTypeScale.caption, weight: PVFontWeight.medium))
                .foregroundStyle(PVColor.textSecondary)
            PVChipGroup(style: .loose) {
                PVChip(
                    L10n.Sources.dateQualifierAsStated,
                    isSelected: draft.qualifier == "",
                    tone: .accent,
                    action: { draft.qualifier = "" }
                )
                .accessibilityIdentifier("\(accessibilityIdentifierPrefix).qualifier.exact")
                PVChip(
                    L10n.Sources.dateQualifierAbout,
                    isSelected: draft.qualifier == "ABT",
                    tone: .accent,
                    action: { draft.qualifier = "ABT" }
                )
                .accessibilityIdentifier("\(accessibilityIdentifierPrefix).qualifier.abt")
                PVChip(
                    L10n.Sources.dateQualifierBefore,
                    isSelected: draft.qualifier == "BEF",
                    tone: .accent,
                    action: { draft.qualifier = "BEF" }
                )
                .accessibilityIdentifier("\(accessibilityIdentifierPrefix).qualifier.bef")
                PVChip(
                    L10n.Sources.dateQualifierAfter,
                    isSelected: draft.qualifier == "AFT",
                    tone: .accent,
                    action: { draft.qualifier = "AFT" }
                )
                .accessibilityIdentifier("\(accessibilityIdentifierPrefix).qualifier.aft")
            }
        }
    }

    @ViewBuilder
    private func cascadeSection(start: Bool) -> some View {
        let heading: LocalizedStringResource = start
            ? (draft.isRange ? L10n.Sources.dateEarliestHeading : L10n.Sources.datePointHeading)
            : L10n.Sources.dateLatestHeading
        VStack(alignment: .leading, spacing: PVSpacing.space5) {
            HStack(alignment: .firstTextBaseline) {
                Text(heading)
                    .font(PVFont.body(size: PVTypeScale.micro, weight: PVFontWeight.semibold))
                    .tracking(PVTypeScale.micro * PVTracking.caps)
                    .textCase(.uppercase)
                    .foregroundStyle(PVColor.textMuted)
                Spacer()
                if start {
                    Text(L10n.Sources.dateLeaveEmptyHint)
                        .font(PVFont.body(size: PVTypeScale.micro, italic: true))
                        .foregroundStyle(PVColor.textFaint)
                }
            }
            HStack(alignment: .bottom, spacing: PVSpacing.space4) {
                cascadeField(
                    L10n.Sources.dateYear,
                    text: yearBinding(start),
                    width: 92,
                    mono: true,
                    isInvalid: draft.isFieldInvalid(.year, start: start),
                    accessibilityIdentifier: "\(accessibilityIdentifierPrefix).\(start ? "start" : "end").year"
                )
                monthSelect(start: start)
                cascadeField(
                    L10n.Sources.dateDay,
                    text: dayBinding(start),
                    width: 76,
                    mono: true,
                    isInvalid: draft.isFieldInvalid(.day, start: start),
                    accessibilityIdentifier: "\(accessibilityIdentifierPrefix).\(start ? "start" : "end").day"
                )
            }
            if let fieldErr = draft.cascadeFieldError(start: start) {
                Text(fieldErr)
                    .font(PVFont.body(size: PVTypeScale.micro))
                    .foregroundStyle(PVColor.danger)
                    .accessibilityIdentifier(start ? "\(accessibilityIdentifierPrefix).start.fieldError" : "\(accessibilityIdentifierPrefix).end.fieldError")
            } else if !start, let err = draft.endError {
                Text(err)
                    .font(PVFont.body(size: PVTypeScale.micro))
                    .foregroundStyle(PVColor.danger)
            }
            if start ? draft.showStartTime : draft.showEndTime {
                timeBlock(start: start)
            }
            Button {
                if start {
                    draft.showStartTime.toggle()
                } else {
                    draft.showEndTime.toggle()
                }
            } label: {
                Text(start
                    ? (draft.showStartTime ? L10n.Sources.dateHideTime : L10n.Sources.dateAddTime)
                    : (draft.showEndTime ? L10n.Sources.dateHideTime : L10n.Sources.dateAddTime)
                )
                .font(PVFont.body(size: PVTypeScale.caption))
                .foregroundStyle(PVColor.textLink)
            }
            .buttonStyle(.plain)
            .accessibilityIdentifier("\(accessibilityIdentifierPrefix).\(start ? "start" : "end").timeToggle")
        }
    }

    private func timeBlock(start: Bool) -> some View {
        let side = start ? "start" : "end"
        return VStack(alignment: .leading, spacing: PVSpacing.space4) {
            HStack(alignment: .bottom, spacing: PVSpacing.space3) {
                cascadeField(
                    L10n.Sources.dateHour,
                    text: hourBinding(start),
                    width: 62,
                    mono: true,
                    isInvalid: draft.isFieldInvalid(.hour, start: start),
                    accessibilityIdentifier: "\(accessibilityIdentifierPrefix).\(side).hour"
                )
                cascadeField(
                    L10n.Sources.dateMinute,
                    text: minuteBinding(start),
                    width: 62,
                    mono: true,
                    isInvalid: draft.isFieldInvalid(.minute, start: start),
                    accessibilityIdentifier: "\(accessibilityIdentifierPrefix).\(side).minute"
                )
                if start {
                    cascadeField(
                        L10n.Sources.dateSecond,
                        text: secondBinding(start: true),
                        width: 62,
                        mono: true,
                        isInvalid: draft.isFieldInvalid(.second, start: true),
                        accessibilityIdentifier: "\(accessibilityIdentifierPrefix).start.second"
                    )
                    cascadeField(
                        L10n.Sources.dateMillisecond,
                        text: millisecondBinding(start: true),
                        width: 70,
                        mono: true,
                        isInvalid: draft.isFieldInvalid(.millisecond, start: true),
                        accessibilityIdentifier: "\(accessibilityIdentifierPrefix).start.millisecond"
                    )
                }
                cascadeField(
                    L10n.Sources.dateTimeZone,
                    text: tzBinding(start),
                    width: nil,
                    mono: false,
                    accessibilityIdentifier: "\(accessibilityIdentifierPrefix).\(side).timezone"
                )
            }
        }
        .padding(PVSpacing.space5)
        .background(PVColor.surfaceSunken)
        .clipShape(RoundedRectangle(cornerRadius: PVRadius.sm, style: .continuous))
        .overlay(
            RoundedRectangle(cornerRadius: PVRadius.sm, style: .continuous)
                .stroke(PVColor.borderSubtle, lineWidth: 1)
        )
    }

    private var advancedToggle: some View {
        Button {
            draft.showAdvanced.toggle()
        } label: {
            Text(draft.showAdvanced ? L10n.Sources.dateHideAdvanced : L10n.Sources.dateShowAdvanced)
                .font(PVFont.body(size: PVTypeScale.caption))
                .foregroundStyle(PVColor.textLink)
        }
        .buttonStyle(.plain)
        .accessibilityIdentifier("\(accessibilityIdentifierPrefix).advancedToggle")
    }

    private var advancedSection: some View {
        VStack(alignment: .leading, spacing: PVSpacing.space5) {
            VStack(alignment: .leading, spacing: PVSpacing.space2) {
                Text(L10n.Sources.dateCalendar)
                    .font(PVFont.body(size: PVTypeScale.micro))
                    .foregroundStyle(PVColor.textSecondary)
                PVSelect(
                    selection: $draft.calendar,
                    options: DateValueSelectOptions.calendars,
                    menuWidth: 220,
                    fillsWidth: false,
                    accessibilityLabel: L10n.Sources.dateCalendar,
                    accessibilityIdentifier: "\(accessibilityIdentifierPrefix).calendar"
                )
                .frame(width: 220)
            }
            VStack(alignment: .leading, spacing: PVSpacing.space2) {
                Text(L10n.Sources.datePhrase)
                    .font(PVFont.body(size: PVTypeScale.micro))
                    .foregroundStyle(PVColor.textSecondary)
                PVInput(text: $draft.phrase, size: .sm, prompt: L10n.Sources.datePhrasePrompt)
                .accessibilityIdentifier("\(accessibilityIdentifierPrefix).phrase")
                Text(L10n.Sources.datePhraseHint)
                    .font(PVFont.body(size: PVTypeScale.micro, italic: true))
                    .foregroundStyle(PVColor.textMuted)
            }
        }
    }

    private func cascadeField(
        _ label: LocalizedStringResource,
        text: Binding<String>,
        width: CGFloat?,
        mono: Bool,
        disabled: Bool = false,
        isInvalid: Bool = false,
        accessibilityIdentifier: String? = nil
    ) -> some View {
        VStack(alignment: .leading, spacing: PVSpacing.space2) {
            Text(label)
                .font(PVFont.body(size: PVTypeScale.micro))
                .foregroundStyle(isInvalid ? PVColor.danger : PVColor.textSecondary)
            PVInput(text: text, size: .sm, mono: mono, isInvalid: isInvalid)
                .frame(width: width)
                .disabled(disabled)
                .opacity(disabled ? 0.42 : 1)
                .modifier(DateOptionalAccessibilityIdentifier(identifier: accessibilityIdentifier))
        }
        .frame(maxWidth: width == nil ? .infinity : nil, alignment: .leading)
    }

    private func monthSelect(start: Bool) -> some View {
        VStack(alignment: .leading, spacing: PVSpacing.space2) {
            Text(L10n.Sources.dateMonth)
                .font(PVFont.body(size: PVTypeScale.micro))
                .foregroundStyle(PVColor.textSecondary)
            PVSelect(
                selection: monthBinding(start),
                options: DateValueSelectOptions.months,
                fillsWidth: false,
                accessibilityLabel: L10n.Sources.dateMonth,
                accessibilityIdentifier: "\(accessibilityIdentifierPrefix).\(start ? "start" : "end").month"
            )
            .frame(width: 136)
        }
    }

    private func yearBinding(_ start: Bool) -> Binding<String> {
        intFieldBinding(
            get: { start ? draft.startYear : draft.endYear },
            set: { value in
                if start {
                    draft.startYear = value
                    draft.applyStartCascade()
                } else {
                    draft.endYear = value
                    draft.applyEndCascade()
                }
            }
        )
    }

    private func monthBinding(_ start: Bool) -> Binding<String> {
        Binding(
            get: {
                let value = start ? draft.startMonth : draft.endMonth
                return value.map(String.init) ?? ""
            },
            set: { raw in
                let value = Int32(raw)
                if start {
                    draft.startMonth = value
                    draft.applyStartCascade()
                } else {
                    draft.endMonth = value
                    draft.applyEndCascade()
                }
            }
        )
    }

    private func dayBinding(_ start: Bool) -> Binding<String> {
        intFieldBinding(
            get: { start ? draft.startDay : draft.endDay },
            set: { value in
                if start {
                    draft.startDay = value
                    draft.applyStartCascade()
                } else {
                    draft.endDay = value
                    draft.applyEndCascade()
                }
            }
        )
    }

    private func hourBinding(_ start: Bool) -> Binding<String> {
        intFieldBinding(
            get: { start ? draft.startHour : draft.endHour },
            set: { value in
                if start {
                    draft.startHour = value
                    draft.applyStartCascade()
                } else {
                    draft.endHour = value
                    draft.applyEndCascade()
                }
            }
        )
    }

    private func minuteBinding(_ start: Bool) -> Binding<String> {
        intFieldBinding(
            get: { start ? draft.startMinute : draft.endMinute },
            set: { value in
                if start {
                    draft.startMinute = value
                    draft.applyStartCascade()
                } else {
                    draft.endMinute = value
                    draft.applyEndCascade()
                }
            }
        )
    }

    private func secondBinding(start: Bool) -> Binding<String> {
        intFieldBinding(
            get: { start ? draft.startSecond : draft.endSecond },
            set: { value in
                if start {
                    draft.startSecond = value
                    draft.applyStartCascade()
                } else {
                    draft.endSecond = value
                    draft.applyEndCascade()
                }
            }
        )
    }

    private func millisecondBinding(start: Bool) -> Binding<String> {
        intFieldBinding(
            get: { start ? draft.startMillisecond : draft.endMillisecond },
            set: { value in
                if start {
                    draft.startMillisecond = value
                    draft.applyStartCascade()
                } else {
                    draft.endMillisecond = value
                    draft.applyEndCascade()
                }
            }
        )
    }

    private func tzBinding(_ start: Bool) -> Binding<String> {
        start ? $draft.startTZ : $draft.endTZ
    }

    private func intFieldBinding(
        get: @escaping () -> Int32?,
        set: @escaping (Int32?) -> Void
    ) -> Binding<String> {
        Binding(
            get: { get().map(String.init) ?? "" },
            set: { raw in
                let trimmed = raw.trimmingCharacters(in: .whitespacesAndNewlines)
                if trimmed.isEmpty {
                    set(nil)
                } else if let value = Int32(trimmed) {
                    set(value)
                }
            }
        )
    }
}

/// Calendar and month option lists for DateValue `PVSelect` remounts.
enum DateValueSelectOptions {
    static var calendars: [PVSelectOption] {
        [
            PVSelectOption(value: "gregorian", label: String(localized: L10n.Sources.dateCalendarGregorian)),
            PVSelectOption(value: "julian", label: String(localized: L10n.Sources.dateCalendarJulian)),
            PVSelectOption(
                value: "french-republican",
                label: String(localized: L10n.Sources.dateCalendarFrenchRepublican)
            ),
            PVSelectOption(value: "hebrew", label: String(localized: L10n.Sources.dateCalendarHebrew)),
        ]
    }

    static var months: [PVSelectOption] {
        [
            ("", L10n.Sources.dateMonthNone),
            ("1", L10n.Sources.dateMonthJanuary),
            ("2", L10n.Sources.dateMonthFebruary),
            ("3", L10n.Sources.dateMonthMarch),
            ("4", L10n.Sources.dateMonthApril),
            ("5", L10n.Sources.dateMonthMay),
            ("6", L10n.Sources.dateMonthJune),
            ("7", L10n.Sources.dateMonthJuly),
            ("8", L10n.Sources.dateMonthAugust),
            ("9", L10n.Sources.dateMonthSeptember),
            ("10", L10n.Sources.dateMonthOctober),
            ("11", L10n.Sources.dateMonthNovember),
            ("12", L10n.Sources.dateMonthDecember),
        ].map { PVSelectOption(value: $0.0, label: String(localized: $0.1)) }
    }
}

private struct DateOptionalAccessibilityIdentifier: ViewModifier {
    let identifier: String?

    func body(content: Content) -> some View {
        if let identifier {
            content.accessibilityIdentifier(identifier)
        } else {
            content
        }
    }
}
