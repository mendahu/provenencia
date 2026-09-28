import SwiftUI

/// Shared resource-delete presentation — counterpart to `recipes/PVDeleteImpact.jsx`.
///
/// The caller already has a `GetDeleteImpact` report. This recipe only
/// **renders** it: `allowed` composes the shipped ``pvConfirm(item:)``; `!allowed`
/// is a notice (groups + extra gates). The caller does not pick the branch.
///
/// Deliberate deviations from the web kit:
/// - All chrome copy goes through `L10n.DeleteImpact` (the JSX inlines English).
/// - Confirm uses ``pvConfirm(item:)``, not a Bool / system alert.
/// - The report is ``CatalogDeleteImpact`` (`listed`, no lock flag).
///
/// Lands as `DesignSystem/Recipes/DeleteImpact/PVDeleteImpact.swift`.

// MARK: - Request

struct PVDeleteImpactTarget: Hashable {
    let kind: String
    let id: String
    let ref: String
    let title: String
}

/// Snapshot ``pvDeleteImpact(item:)`` presents — never a Bool.
struct PVDeleteImpactRequest: Identifiable {
    let target: PVDeleteImpactTarget
    let report: CatalogDeleteImpact
    var id: String { "\(target.kind):\(target.id)" }
}

// MARK: - Copy helpers (testable)

enum PVDeleteImpactCopy {
    static func noun(_ kind: String, count: Int = 1) -> String {
        L10n.DeleteImpact.noun(kind, count: count)
    }

    static func heading(via: String, kind: String, total: Int) -> String {
        L10n.DeleteImpact.viaHeading(via: via, kind: kind, total: total)
    }

    static func heading(_ group: CatalogDeleteImpactGroup) -> String {
        heading(via: group.via, kind: group.kind, total: group.total)
    }

    static func overflow(remainder: Int, kind: String) -> String {
        L10n.DeleteImpact.overflow(count: remainder, kind: kind)
    }

    static func isKnownVia(_ via: String) -> Bool {
        L10n.DeleteImpact.isKnownVia(via)
    }

    static func confirmCopy(for target: PVDeleteImpactTarget) -> PVConfirmCopy {
        let noun = noun(target.kind)
        return PVConfirmCopy(
            title: L10n.DeleteImpact.confirmTitle(noun: noun, ref: target.ref),
            message: String(localized: L10n.DeleteImpact.confirmMessage),
            confirmLabel: L10n.DeleteImpact.deleteAction(noun: noun),
            cancelLabel: L10n.DeleteImpact.keepAction(noun: noun)
        )
    }

    static func noticeTitle(for target: PVDeleteImpactTarget) -> String {
        L10n.DeleteImpact.noticeTitle(noun: noun(target.kind))
    }

    static func extraGate(for gate: CatalogDeleteImpactGate, ref: String) -> (
        title: LocalizedStringResource,
        body: String
    )? {
        switch gate {
        case .ok, .inbound:
            return nil
        case .notFound:
            return (L10n.DeleteImpact.gateNotFoundTitle, L10n.DeleteImpact.gateNotFoundBody(ref: ref))
        case .edgeLocked:
            return (L10n.DeleteImpact.gateEdgeLockedTitle, String(localized: L10n.DeleteImpact.gateEdgeLockedBody))
        case .originLocked:
            return (L10n.DeleteImpact.gateOriginLockedTitle, String(localized: L10n.DeleteImpact.gateOriginLockedBody))
        case .infra:
            return (L10n.DeleteImpact.gateInfraTitle, String(localized: L10n.DeleteImpact.gateInfraBody))
        }
    }

    static func summary(_ target: PVDeleteImpactTarget, _ report: CatalogDeleteImpact, firstN: Int = 3) -> String {
        let noun = noun(target.kind)
        if report.allowed {
            return L10n.DeleteImpact.summaryAllowed(noun: noun, ref: target.ref, title: target.title)
        }
        var parts = [L10n.DeleteImpact.summaryBlocked(noun: noun, ref: target.ref)]
        if let gate = extraGate(for: report.gate, ref: target.ref) {
            parts.append(String(localized: gate.title) + ".")
        }
        for group in report.groups where group.total > 0 {
            let refs = group.listed.prefix(firstN).map(\.ref)
            let rest = group.total - refs.count
            var named = refs.joined(separator: ", ")
            if rest > 0 {
                named += L10n.DeleteImpact.summaryMore(count: rest)
            }
            parts.append(L10n.DeleteImpact.summaryGroup(heading: heading(group), refs: named))
        }
        return parts.joined(separator: " ")
    }
}

enum PVDeleteImpactControls {
    static func remainder(total: Int, listedCount: Int) -> Int {
        max(0, total - listedCount)
    }

    static func showsDestructiveAction(_ report: CatalogDeleteImpact) -> Bool {
        report.allowed
    }

    static func isExtraGate(_ gate: CatalogDeleteImpactGate) -> Bool {
        switch gate {
        case .ok, .inbound: false
        case .notFound, .edgeLocked, .infra, .originLocked: true
        }
    }

    @MainActor
    static func activate(
        _ listed: CatalogDeleteImpactListed,
        onNavigate: @escaping (WorkspaceLocation) -> Void,
        dismiss: () -> Void,
        then: ((@escaping () -> Void) -> Void)? = nil
    ) {
        dismiss()
        let location = listed.location
        if let then {
            then { onNavigate(location) }
            return
        }
        Task { @MainActor in
            onNavigate(location)
        }
    }
}

enum PVDeleteImpactLayout {
    static let panelWidth: CGFloat = 440
    static let listMaxHeight: CGFloat = 320
}

enum PVDeleteImpactAccessibility {
    static let noticePrefix = "deleteImpact"
}

// MARK: - Notice

struct PVDeleteImpactNotice: View {
    let target: PVDeleteImpactTarget
    let report: CatalogDeleteImpact
    let onNavigate: (WorkspaceLocation) -> Void
    let onDismiss: () -> Void

    var body: some View {
        let groups = report.groups.filter { $0.total > 0 }
        let gate = PVDeleteImpactCopy.extraGate(for: report.gate, ref: target.ref)
        PVPanel(
            title: Text(verbatim: PVDeleteImpactCopy.noticeTitle(for: target)),
            subtitle: groups.isEmpty ? nil : Text(L10n.DeleteImpact.noticeSubtitle),
            width: PVDeleteImpactLayout.panelWidth,
            footerChrome: .plain
        ) {
            VStack(alignment: .leading, spacing: PVSpacing.space7) {
                targetLine
                if let gate {
                    PVCallout(tone: .neutral, title: gate.title, message: gate.body, compact: true)
                }
                if !groups.isEmpty {
                    ScrollView {
                        VStack(alignment: .leading, spacing: PVSpacing.space7) {
                            ForEach(Array(groups.enumerated()), id: \.element.via) { _, group in
                                groupView(group)
                            }
                        }
                    }
                    .frame(maxHeight: PVDeleteImpactLayout.listMaxHeight)
                }
            }
            .accessibilityElement(children: .contain)
            .accessibilityLabel(Text(verbatim: PVDeleteImpactCopy.summary(target, report)))
        } footer: {
            HStack {
                Spacer(minLength: PVSpacing.space8)
                PVButton(L10n.DeleteImpact.done, variant: .primary, size: .lg, action: onDismiss)
                    .keyboardShortcut(.defaultAction)
                    .accessibilityIdentifier("deleteImpact.done")
            }
        }
        .accessibilityIdentifier("deleteImpact.notice")
    }

    private var targetLine: some View {
        HStack(alignment: .firstTextBaseline, spacing: PVSpacing.space4) {
            Text(verbatim: target.ref)
                .font(PVFont.mono(size: PVTypeScale.micro))
                .foregroundStyle(PVColor.textPrimary)
                .padding(.horizontal, PVSpacing.space3)
                .padding(.vertical, PVSpacing.space1)
                .background(
                    PVColor.surfaceSunken,
                    in: RoundedRectangle(cornerRadius: PVRadius.xs, style: .continuous)
                )
            Text(verbatim: target.title)
                .font(PVFont.body(size: PVTypeScale.bodySmall))
                .foregroundStyle(PVColor.textPrimary)
                .fixedSize(horizontal: false, vertical: true)
        }
    }

    @ViewBuilder
    private func groupView(_ group: CatalogDeleteImpactGroup) -> some View {
        let rest = PVDeleteImpactControls.remainder(total: group.total, listedCount: group.listed.count)
        VStack(alignment: .leading, spacing: PVSpacing.space3) {
            HStack(alignment: .firstTextBaseline) {
                Text(verbatim: PVDeleteImpactCopy.heading(group))
                    .font(PVFont.body(size: PVTypeScale.bodySmall, weight: PVFontWeight.semibold))
                    .foregroundStyle(PVColor.textPrimary)
                    .fixedSize(horizontal: false, vertical: true)
                Spacer(minLength: PVSpacing.space5)
                if !PVDeleteImpactCopy.isKnownVia(group.via) {
                    Text(verbatim: group.via)
                        .font(PVFont.mono(size: PVTypeScale.micro))
                        .foregroundStyle(PVColor.textMuted)
                }
            }
            VStack(spacing: 0) {
                ForEach(Array(group.listed.enumerated()), id: \.element.id) { index, row in
                    if index > 0 { Divider() }
                    Button {
                        PVDeleteImpactControls.activate(row, onNavigate: onNavigate, dismiss: onDismiss)
                    } label: {
                        HStack(alignment: .firstTextBaseline, spacing: PVSpacing.space5) {
                            Text(verbatim: row.ref)
                                .font(PVFont.mono(size: PVTypeScale.micro))
                                .foregroundStyle(PVColor.textLink)
                                .fixedSize()
                            Text(verbatim: row.title)
                                .font(PVFont.body(size: PVTypeScale.bodySmall))
                                .foregroundStyle(PVColor.textPrimary)
                                .frame(maxWidth: .infinity, alignment: .leading)
                                .lineLimit(1)
                                .truncationMode(.tail)
                            PVIcon(.arrowUpRight, size: 14)
                                .foregroundStyle(PVColor.textFaint)
                        }
                        .padding(.horizontal, PVSpacing.space5)
                        .padding(.vertical, PVSpacing.space4)
                        .contentShape(Rectangle())
                    }
                    .buttonStyle(.plain)
                    .accessibilityLabel(Text(verbatim: L10n.DeleteImpact.rowAccessibility(ref: row.ref, title: row.title)))
                }
                if rest > 0 {
                    Divider()
                    Text(verbatim: PVDeleteImpactCopy.overflow(remainder: rest, kind: group.kind))
                        .font(PVFont.body(size: PVTypeScale.caption))
                        .foregroundStyle(PVColor.textMuted)
                        .frame(maxWidth: .infinity, alignment: .leading)
                        .padding(.horizontal, PVSpacing.space5)
                        .padding(.vertical, PVSpacing.space4)
                        .background(PVColor.surfaceSunken)
                }
            }
            .background(
                PVColor.surfaceRaised,
                in: RoundedRectangle(cornerRadius: PVRadius.md, style: .continuous)
            )
            .overlay(
                RoundedRectangle(cornerRadius: PVRadius.md, style: .continuous)
                    .strokeBorder(PVColor.borderSubtle)
            )
        }
    }
}

// MARK: - Presenter

extension View {
    /// Present with a snapshot. The report decides confirm vs notice.
    func pvDeleteImpact(
        item: Binding<PVDeleteImpactRequest?>,
        isRunning: Bool = false,
        error: String? = nil,
        accessibilityIdentifierPrefix: String? = PVDeleteImpactAccessibility.noticePrefix,
        onConfirm: @escaping () -> Void,
        onNavigate: @escaping (WorkspaceLocation) -> Void
    ) -> some View {
        let confirmItem = Binding<PVDeleteImpactRequest?>(
            get: { item.wrappedValue.flatMap { $0.report.allowed ? $0 : nil } },
            set: { item.wrappedValue = $0 }
        )
        let noticeItem = Binding<PVDeleteImpactRequest?>(
            get: { item.wrappedValue.flatMap { $0.report.allowed ? nil : $0 } },
            set: { item.wrappedValue = $0 }
        )
        return self
            .pvConfirm(
                item: confirmItem,
                copy: { PVDeleteImpactCopy.confirmCopy(for: $0.target) },
                isRunning: isRunning,
                accessibilityIdentifierPrefix: accessibilityIdentifierPrefix,
                onConfirm: onConfirm
            ) { request in
                VStack(alignment: .leading, spacing: PVSpacing.space4) {
                    HStack(alignment: .firstTextBaseline, spacing: PVSpacing.space4) {
                        Text(verbatim: request.target.ref)
                            .font(PVFont.mono(size: PVTypeScale.micro))
                            .foregroundStyle(PVColor.textPrimary)
                            .padding(.horizontal, PVSpacing.space3)
                            .padding(.vertical, PVSpacing.space1)
                            .background(
                                PVColor.surfaceSunken,
                                in: RoundedRectangle(cornerRadius: PVRadius.xs, style: .continuous)
                            )
                        Text(verbatim: request.target.title)
                            .font(PVFont.body(size: PVTypeScale.bodySmall))
                            .foregroundStyle(PVColor.textPrimary)
                            .fixedSize(horizontal: false, vertical: true)
                    }
                    if let error, !error.isEmpty {
                        PVCallout(tone: .danger, message: error)
                            .accessibilityIdentifier("deleteImpact.confirmError")
                    }
                    Text(L10n.DeleteImpact.nothingReferences)
                        .font(PVFont.body(size: PVTypeScale.caption))
                        .foregroundStyle(PVColor.textMuted)
                }
                .accessibilityLabel(Text(verbatim: PVDeleteImpactCopy.summary(request.target, request.report)))
            }
            .sheet(item: noticeItem) { request in
                PVDeleteImpactNotice(
                    target: request.target,
                    report: request.report,
                    onNavigate: onNavigate,
                    onDismiss: { item.wrappedValue = nil }
                )
            }
    }

    func pvDeleteImpact(
        flow: DeleteImpactFlow,
        accessibilityIdentifierPrefix: String? = PVDeleteImpactAccessibility.noticePrefix,
        onConfirm: @escaping () -> Void,
        onNavigate: @escaping (WorkspaceLocation) -> Void
    ) -> some View {
        pvDeleteImpact(
            item: Binding(
                get: { flow.request },
                set: { flow.applyRequest($0) }
            ),
            isRunning: flow.isRunning,
            error: flow.error,
            accessibilityIdentifierPrefix: accessibilityIdentifierPrefix,
            onConfirm: onConfirm,
            onNavigate: onNavigate
        )
    }
}

// MARK: - Preview fixtures

enum PVDeleteImpactPreviewData {
    static let citation = PVDeleteImpactTarget(
        kind: "citation",
        id: "cit-preview",
        ref: "CIT-7KD45",
        title: "1881 census, Leeds — RG11/4543 f.62 p.18"
    )
    static let subjectG2 = PVDeleteImpactTarget(
        kind: "subject",
        id: "sub-preview-g2",
        ref: "CPR-2HX8V",
        title: "Thomas Hartley (b. c.1872)"
    )
    static let bridge = PVDeleteImpactTarget(
        kind: "subject",
        id: "sub-preview-bridge",
        ref: "CPR-0PW4K",
        title: "Ellen Hartley → Thomas Hartley (son)"
    )
    static let observation = PVDeleteImpactTarget(
        kind: "observation",
        id: "obs-preview",
        ref: "OBS-9M1TR",
        title: "Residence: 14 Back Nile Street, Leeds"
    )
    static let property = PVDeleteImpactTarget(
        kind: "property",
        id: "prop-preview",
        ref: "PRP-EVTYP",
        title: "Event type"
    )
    static let user = PVDeleteImpactTarget(
        kind: "user",
        id: "usr-preview",
        ref: "USR-4N2P0",
        title: "Jake Robins"
    )

    static func loc(_ observationId: String) -> WorkspaceLocation {
        WorkspaceLocation(
            section: .sources,
            sourceId: "preview-source",
            observationId: observationId,
            sourceSurface: .citationComposer
        )
    }

    static func listed(ref: String, title: String, id: String? = nil) -> CatalogDeleteImpactListed {
        CatalogDeleteImpactListed(id: id ?? ref, ref: ref, title: title, location: loc(ref))
    }

    static let allowed = CatalogDeleteImpact(allowed: true, gate: .ok, groups: [])

    static let blockedCitation = CatalogDeleteImpact(
        allowed: false,
        gate: .inbound,
        groups: [
            CatalogDeleteImpactGroup(
                via: "observations.citation_id",
                kind: "observation",
                total: 3,
                listed: [
                    listed(ref: "OBS-4Q2PA", title: "Name: Ellen Hartley"),
                    listed(ref: "OBS-4Q2PB", title: "Occupation: worsted weaver"),
                    listed(ref: "OBS-9M1TR", title: "Residence: 14 Back Nile Street, Leeds"),
                ]
            ),
        ]
    )

    static let blockedG2 = CatalogDeleteImpact(
        allowed: false,
        gate: .inbound,
        groups: [
            CatalogDeleteImpactGroup(
                via: "observations.value_subject_id",
                kind: "observation",
                total: 2,
                listed: [
                    listed(ref: "OBS-7C3LD", title: "Son: Thomas Hartley"),
                    listed(ref: "OBS-7C3LF", title: "Witness: Thomas Hartley"),
                ]
            ),
        ]
    )

    static let blockedG4 = CatalogDeleteImpact(
        allowed: false,
        gate: .inbound,
        groups: [
            CatalogDeleteImpactGroup(
                via: "observations.subject_id",
                kind: "observation",
                total: 1,
                listed: [listed(ref: "OBS-4Q2PA", title: "Name: Thomas Hartley")]
            ),
            CatalogDeleteImpactGroup(
                via: "observations.value_subject_id",
                kind: "observation",
                total: 2,
                listed: [
                    listed(ref: "OBS-7C3LD", title: "Son: Thomas Hartley"),
                    listed(ref: "OBS-7C3LF", title: "Witness: Thomas Hartley"),
                ]
            ),
        ]
    )

    static let overflow: CatalogDeleteImpact = {
        let titles = [
            "Head of household", "Occupation recorded", "Birthplace: Leeds, Yorkshire",
            "Residence on census night", "Relationship to head",
        ]
        let listed = (0..<20).map { index in
            Self.listed(
                ref: String(format: "OBS-%05d", index + 1),
                title: titles[index % titles.count],
                id: "overflow-\(index)"
            )
        }
        return CatalogDeleteImpact(
            allowed: false,
            gate: .inbound,
            groups: [
                CatalogDeleteImpactGroup(
                    via: "observations.citation_id",
                    kind: "observation",
                    total: 27,
                    listed: listed
                ),
            ]
        )
    }()

    static let unknownVia = CatalogDeleteImpact(
        allowed: false,
        gate: .inbound,
        groups: [
            CatalogDeleteImpactGroup(
                via: "sameness_claim_evidence.observation_id",
                kind: "sameness_claim",
                total: 2,
                listed: [
                    listed(ref: "CLM-3JQ8", title: "Same as Thomas Hartley of Leeds"),
                    listed(ref: "CLM-3JQ9", title: "Same as Thos. Hartley, 1881 census"),
                ]
            ),
        ]
    )

    static let g6Extras = CatalogDeleteImpact(
        allowed: false,
        gate: .inbound,
        groups: [
            CatalogDeleteImpactGroup(
                via: "observations.subject_id",
                kind: "observation",
                total: 1,
                listed: [listed(ref: "OBS-8K2NN", title: "Notes: mentioned as grandson")]
            ),
        ]
    )

    static let propertyTerms = CatalogDeleteImpact(
        allowed: false,
        gate: .inbound,
        groups: [
            CatalogDeleteImpactGroup(
                via: "property_terms.property_id",
                kind: "property_term",
                total: 2,
                listed: [
                    listed(ref: "TRM-MARR", title: "marriage"),
                    listed(ref: "TRM-BIRT", title: "birth"),
                ]
            ),
        ]
    )
}

#Preview("Clear citation") {
    PVConfirmContent(
        copy: PVDeleteImpactCopy.confirmCopy(for: PVDeleteImpactPreviewData.citation),
        onConfirm: {},
        onCancel: {}
    ) {
        Text(verbatim: PVDeleteImpactPreviewData.citation.title)
    }
    .background(PVColor.surfaceCard)
}

#Preview("G6 allowed") {
    PVConfirmContent(
        copy: PVDeleteImpactCopy.confirmCopy(for: PVDeleteImpactPreviewData.bridge),
        onConfirm: {},
        onCancel: {}
    ) {
        Text(verbatim: PVDeleteImpactPreviewData.bridge.title)
    }
    .background(PVColor.surfaceCard)
}

#Preview("Blocked citation") {
    PVDeleteImpactNotice(
        target: PVDeleteImpactPreviewData.citation,
        report: PVDeleteImpactPreviewData.blockedCitation,
        onNavigate: { _ in },
        onDismiss: {}
    )
}

#Preview("Blocked subject G2") {
    PVDeleteImpactNotice(
        target: PVDeleteImpactPreviewData.subjectG2,
        report: PVDeleteImpactPreviewData.blockedG2,
        onNavigate: { _ in },
        onDismiss: {}
    )
}

#Preview("Blocked subject G4") {
    PVDeleteImpactNotice(
        target: PVDeleteImpactPreviewData.subjectG2,
        report: PVDeleteImpactPreviewData.blockedG4,
        onNavigate: { _ in },
        onDismiss: {}
    )
}

#Preview("Overflow") {
    PVDeleteImpactNotice(
        target: PVDeleteImpactPreviewData.citation,
        report: PVDeleteImpactPreviewData.overflow,
        onNavigate: { _ in },
        onDismiss: {}
    )
}

#Preview("Unknown via") {
    PVDeleteImpactNotice(
        target: PVDeleteImpactPreviewData.observation,
        report: PVDeleteImpactPreviewData.unknownVia,
        onNavigate: { _ in },
        onDismiss: {}
    )
}

#Preview("Gate edge locked") {
    PVDeleteImpactNotice(
        target: PVDeleteImpactPreviewData.observation,
        report: CatalogDeleteImpact(allowed: false, gate: .edgeLocked, groups: []),
        onNavigate: { _ in },
        onDismiss: {}
    )
}

#Preview("Gate not found") {
    PVDeleteImpactNotice(
        target: PVDeleteImpactPreviewData.citation,
        report: CatalogDeleteImpact(allowed: false, gate: .notFound, groups: []),
        onNavigate: { _ in },
        onDismiss: {}
    )
}

#Preview("Gate origin locked") {
    PVDeleteImpactNotice(
        target: PVDeleteImpactPreviewData.property,
        report: CatalogDeleteImpact(allowed: false, gate: .originLocked, groups: []),
        onNavigate: { _ in },
        onDismiss: {}
    )
}

#Preview("Gate infra") {
    PVDeleteImpactNotice(
        target: PVDeleteImpactPreviewData.user,
        report: CatalogDeleteImpact(allowed: false, gate: .infra, groups: []),
        onNavigate: { _ in },
        onDismiss: {}
    )
}

#Preview("G6 extras") {
    PVDeleteImpactNotice(
        target: PVDeleteImpactPreviewData.bridge,
        report: PVDeleteImpactPreviewData.g6Extras,
        onNavigate: { _ in },
        onDismiss: {}
    )
}

#Preview("Property terms") {
    PVDeleteImpactNotice(
        target: PVDeleteImpactPreviewData.property,
        report: PVDeleteImpactPreviewData.propertyTerms,
        onNavigate: { _ in },
        onDismiss: {}
    )
}
