import Foundation
import Testing
@testable import Provenencia

@Suite
struct PVDeleteImpactCopyTests {
    @Test func allowedConfirmNamesKindAndRef() {
        let copy = PVDeleteImpactCopy.confirmCopy(for: PVDeleteImpactPreviewData.citation)
        #expect(copy.title == "Delete citation CIT-7KD45?")
        #expect(copy.confirm == "Delete citation")
        #expect(copy.cancel == "Keep citation")
        #expect(copy.message.contains("erased"))
        #expect(PVDeleteImpactControls.showsDestructiveAction(PVDeleteImpactPreviewData.allowed))
    }

    @Test func promotedSubjectConfirmNamesTheHandleItLeaves() {
        let plain = PVDeleteImpactCopy.confirmCopy(
            for: PVDeleteImpactPreviewData.subjectG2,
            report: PVDeleteImpactPreviewData.allowed
        )
        #expect(!plain.message.contains("PER-"))
        let promoted = PVDeleteImpactCopy.confirmCopy(
            for: PVDeleteImpactPreviewData.subjectG2,
            report: PVDeleteImpactPreviewData.promotedSubject
        )
        #expect(promoted.message.contains("erased"))
        #expect(promoted.message.contains(L10n.DeleteImpact.leavesHandle(handleRefs: "PER-7KD45")))
        #expect(PVDeleteImpactControls.showsDestructiveAction(PVDeleteImpactPreviewData.promotedSubject))
    }

    @Test func pinnedObservationConfirmNamesTheEvidenceItLeaves() {
        let copy = PVDeleteImpactCopy.confirmCopy(
            for: PVDeleteImpactPreviewData.observation,
            report: PVDeleteImpactPreviewData.pinnedObservation
        )
        #expect(copy.message.contains(L10n.DeleteImpact.leavesEvidence(handleRefs: "PER-7KD45")))
        #expect(!copy.message.contains(L10n.DeleteImpact.leavesHandle(handleRefs: "PER-7KD45")))
        #expect(PVDeleteImpactControls.showsDestructiveAction(PVDeleteImpactPreviewData.pinnedObservation))
    }

    @Test func blockedCitationListsThreeObservations() {
        let report = PVDeleteImpactPreviewData.blockedCitation
        #expect(!PVDeleteImpactControls.showsDestructiveAction(report))
        #expect(report.groups.count == 1)
        #expect(report.groups[0].listed.map(\.ref) == ["OBS-4Q2PA", "OBS-4Q2PB", "OBS-9M1TR"])
        #expect(PVDeleteImpactCopy.heading(report.groups[0]) == "3 observations still belong to this")
    }

    @Test func overflowShowsRemainder() {
        let group = PVDeleteImpactPreviewData.overflow.groups[0]
        let rest = PVDeleteImpactControls.remainder(total: group.total, listedCount: group.listed.count)
        #expect(group.total == 27)
        #expect(group.listed.count == 20)
        #expect(rest == 7)
        #expect(PVDeleteImpactCopy.overflow(remainder: rest, kind: group.kind) == "And 7 more observations")
    }

    @Test func unknownViaStillListsRefs() {
        let group = PVDeleteImpactPreviewData.unknownVia.groups[0]
        #expect(!PVDeleteImpactCopy.isKnownVia(group.via))
        #expect(PVDeleteImpactCopy.heading(group) == "2 sameness claims reference this")
        #expect(group.listed.map(\.ref) == ["CLM-3JQ8", "CLM-3JQ9"])
    }

    @Test func extraGatesHaveNoDestructiveButtonAndNoInventedList() {
        for gate in [CatalogDeleteImpactGate.notFound, .edgeLocked, .originLocked, .infra] {
            let report = CatalogDeleteImpact(allowed: false, gate: gate, groups: [])
            #expect(!PVDeleteImpactControls.showsDestructiveAction(report))
            #expect(PVDeleteImpactControls.isExtraGate(gate))
            #expect(report.groups.isEmpty)
            #expect(PVDeleteImpactCopy.extraGate(for: gate, ref: "CIT-7KD45") != nil)
        }
        #expect(PVDeleteImpactCopy.extraGate(for: .ok, ref: "CIT-7KD45") == nil)
        #expect(PVDeleteImpactCopy.extraGate(for: .inbound, ref: "CIT-7KD45") == nil)
    }

    @Test func voiceOverNamesEraseVersusBlocked() {
        let allowed = PVDeleteImpactCopy.summary(
            PVDeleteImpactPreviewData.citation,
            PVDeleteImpactPreviewData.allowed
        )
        #expect(allowed.contains("Delete citation CIT-7KD45"))
        #expect(allowed.contains("erases"))

        let blocked = PVDeleteImpactCopy.summary(
            PVDeleteImpactPreviewData.subjectG2,
            PVDeleteImpactPreviewData.blockedG2
        )
        #expect(blocked.hasPrefix("Blocked."))
        #expect(blocked.contains("CPR-2HX8V"))
        #expect(blocked.contains("OBS-7C3LD"))
        #expect(blocked.contains("OBS-7C3LF"))
        #expect(!blocked.localizedCaseInsensitiveContains("uncited"))
    }

    @Test func g2CopyNeverSaysUncited() {
        let heading = PVDeleteImpactCopy.heading(PVDeleteImpactPreviewData.blockedG2.groups[0])
        #expect(heading == "2 observations use this as an endpoint")
        #expect(!heading.localizedCaseInsensitiveContains("uncited"))
        #expect(PVDeleteImpactPreviewData.blockedG2.groups.count == 1)
        #expect(PVDeleteImpactPreviewData.blockedG2.groups[0].via == "observations.value_subject_id")
    }

    @Test func knownViaHeadingsMatchEngineEdges() {
        let cases: [(String, String, Int, String)] = [
            ("observations.citation_id", "observation", 1, "1 observation still belongs to this"),
            ("observations.subject_id", "observation", 1, "This subject has 1 observation"),
            ("observations.value_subject_id", "observation", 2, "2 observations use this as an endpoint"),
            ("artifacts.source_id", "artifact", 1, "1 artifact belongs to this source"),
            ("subjects.source_id", "subject", 2, "2 subjects belong to this source"),
            ("citations.artifact_id", "citation", 3, "3 citations are drawn from this artifact"),
            ("sources.source_type_id", "source", 1, "1 source has this type"),
            ("source_metadata.field_id", "source", 2, "2 sources record a value for this field"),
            ("source_credibility_assessments.credibility_grade_id", "source", 1, "1 source uses this grade"),
            ("subjects.subject_type_id", "subject", 4, "4 subjects have this type"),
            ("observations.property_id", "observation", 2, "2 observations use this property"),
            ("property_terms.property_id", "property_term", 2, "2 terms belong to this property"),
            ("observations.value_term_id", "observation", 1, "1 observation uses this term"),
            ("artifacts.file_id", "artifact", 1, "1 artifact uses this file"),
        ]
        for item in cases {
            #expect(PVDeleteImpactCopy.isKnownVia(item.0))
            #expect(PVDeleteImpactCopy.heading(via: item.0, kind: item.1, total: item.2) == item.3)
        }
    }

    @Test @MainActor func activateNavigatesThenDismisses() {
        var navigated: WorkspaceLocation?
        var dismissed = false
        let row = PVDeleteImpactPreviewData.listed(ref: "OBS-4Q2PA", title: "Name: Ellen Hartley")
        PVDeleteImpactControls.activate(
            row,
            onNavigate: { navigated = $0 },
            dismiss: { dismissed = true },
            then: { $0() }
        )
        #expect(dismissed)
        #expect(navigated == row.location)
    }
}

@Suite
struct PVDeleteImpactRequestTests {
    @Test func identityCombinesKindAndId() {
        let request = PVDeleteImpactRequest(
            target: PVDeleteImpactPreviewData.citation,
            report: PVDeleteImpactPreviewData.allowed
        )
        #expect(request.id == "citation:cit-preview")
        #expect(request.target.id == "cit-preview")
    }
}

@Suite
@MainActor
struct DeleteImpactFlowTests {
    @Test func askStoresCatalogIdOnTheRequest() async {
        let flow = DeleteImpactFlow()
        await flow.ask(kind: "citation", id: "cit-1", ref: "CIT-1", title: "Census") {
            CatalogDeleteImpact(allowed: true, gate: .ok, groups: [])
        }
        #expect(flow.request?.target.id == "cit-1")
        #expect(flow.request?.id == "citation:cit-1")
        #expect(flow.error == nil)
    }

    @Test func confirmPerformsAgainstTheAskedId() async {
        let flow = DeleteImpactFlow()
        await flow.ask(kind: "observation", id: "obs-1", ref: "OBS-1", title: "Name") {
            CatalogDeleteImpact(allowed: true, gate: .ok, groups: [])
        }
        var seen: String?
        let ok = await flow.confirm { target in
            seen = target.id
        }
        #expect(ok)
        #expect(seen == "obs-1")
        #expect(flow.request == nil)
    }

    @Test func inUseRefetchTurnsConfirmIntoNotice() async {
        let flow = DeleteImpactFlow()
        var fetches = 0
        await flow.ask(kind: "citation", id: "cit-1", ref: "CIT-1", title: "Census") {
            fetches += 1
            if fetches == 1 {
                return CatalogDeleteImpact(allowed: true, gate: .ok, groups: [])
            }
            return CatalogDeleteImpact(
                allowed: false,
                gate: .inbound,
                groups: [
                    CatalogDeleteImpactGroup(
                        via: "observations.citation_id",
                        kind: "observation",
                        total: 1,
                        listed: []
                    ),
                ]
            )
        }
        let ok = await flow.confirm { _ in
            throw CoreInvokeError.coded(
                status: 1,
                code: "citations.in_use",
                kind: .conflict,
                params: []
            )
        }
        #expect(!ok)
        #expect(fetches == 2)
        #expect(flow.request?.report.allowed == false)
        #expect(flow.request?.report.gate == .inbound)
        #expect(flow.error == nil)
    }

    @Test func otherConfirmErrorStaysOnTheSheet() async {
        let flow = DeleteImpactFlow()
        await flow.ask(kind: "citation", id: "cit-1", ref: "CIT-1", title: "Census") {
            CatalogDeleteImpact(allowed: true, gate: .ok, groups: [])
        }
        let ok = await flow.confirm { _ in
            throw CoreInvokeError.coded(
                status: 1,
                code: "internal",
                kind: .internal,
                params: []
            )
        }
        #expect(!ok)
        #expect(flow.request?.report.allowed == true)
        #expect(flow.error != nil)
    }
}

@Suite
struct PVConfirmCopyLabelTests {
    @Test func stringLabelsStoreVerbatim() {
        let copy = PVConfirmCopy(
            title: "Delete citation CIT-7KD45?",
            message: "It’s erased from the catalog.",
            confirmLabel: "Delete citation",
            cancelLabel: "Keep citation"
        )
        #expect(copy.confirm == "Delete citation")
        #expect(copy.cancel == "Keep citation")
    }
}
