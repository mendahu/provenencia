import Foundation
import Testing
@testable import Provenencia

@Suite(.serialized)
@MainActor
struct PromoteModelTests {
    private let projectDir = "/tmp/promote-model.provenencia"
    private let sourceID = "src-1"

    private func waitUntil(_ condition: () -> Bool) async -> Bool {
        for _ in 0..<200 {
            if condition() { return true }
            try? await Task.sleep(nanoseconds: 10_000_000)
        }
        return condition()
    }

    private func subject(_ id: String, ref: String, label: String, type: String = "type-person") -> CatalogSubject {
        CatalogSubject(id: id, ref: ref, sourceID: sourceID, subjectTypeID: type, label: label, description: "")
    }

    private func nameObservation(_ id: String, subjectID: String, form: String) -> CatalogObservation {
        CatalogObservation(
            id: id, ref: "OBS-\(id)", citationID: "c1", subjectID: subjectID, propertyID: "p-name",
            polarity: "positive", valueText: form, valueInteger: nil, valueDateID: "",
            valueNameID: "n-\(id)", nameForm: form, valueSubjectID: "", valueTermID: "",
            propertyKey: "name", propertyLabel: "Name", propertyValueType: "name"
        )
    }

    /// A Source with James (sub-1, unpromoted) and an already-filed James
    /// Robins (sub-2 on PER-…), so suggestions and search have a target.
    private func makeStore() -> FakeStore {
        let store = FakeStore()
        store.subjectTypesByProject[projectDir] = [
            CatalogSubjectType(
                id: "type-person", key: "person", origin: "provenencia", label: "Person",
                description: "", refPrefix: "PER", candidateRefPrefix: "CPR"
            ),
            CatalogSubjectType(
                id: "type-event", key: "event", origin: "provenencia", label: "Event",
                description: "", refPrefix: "EVT", candidateRefPrefix: "CEV"
            ),
        ]
        store.subjectsBySource[sourceID] = [
            subject("sub-1", ref: "CPR-1", label: "James Robins"),
            subject("sub-2", ref: "CPR-2", label: "James Robins"),
        ]
        store.observationsBySource[sourceID] = [
            nameObservation("o1", subjectID: "sub-1", form: "James Robins"),
            nameObservation("o2", subjectID: "sub-2", form: "James Robins"),
        ]
        return store
    }

    private func entry(
        subjectID: String = "sub-1",
        kind: EvidencePrimaryKind = .person,
        ref: String = "CPR-1",
        title: String = "James Robins"
    ) -> PromoteEntry {
        PromoteEntry(location: .promote(
            sourceId: sourceID, subjectId: subjectID, kind: kind, ref: ref, title: title, sourceTitle: "1851 census"
        ))!
    }

    private func makeModel(
        store: FakeStore,
        entry: PromoteEntry? = nil,
        counts: CatalogCounts? = nil
    ) -> (PromoteModel, WorkspaceNavigation) {
        let session = WorkspaceSession(projectKey: ProjectKey(projectDir: projectDir), store: store)
        let model = PromoteModel(
            entry: entry ?? self.entry(),
            session: session,
            store: store,
            userID: "user-1",
            catalogCounts: counts
        )
        model.searchDebounce = .zero
        let navigation = WorkspaceNavigation()
        navigation.go(to: model.entry.graphLocation)
        navigation.go(to: .promote(
            sourceId: sourceID, subjectId: model.entry.subjectID, kind: model.kind,
            ref: model.entry.subjectRef, title: model.entry.subjectName, sourceTitle: "1851 census"
        ))
        model.navigation = navigation
        navigation.leaveGuard = model
        return (model, navigation)
    }

    /// Files sub-2 so there is an existing Person to join.
    private func existingPerson(_ store: FakeStore) async throws -> CatalogPromoteResult {
        try await store.promoteSubject(projectDir: projectDir, userID: "user-1", subjectID: "sub-2")
    }

    // MARK: Entry

    @Test func entryIsFailClosed() {
        #expect(PromoteEntry(location: .sectionRoot(.sources)) == nil)
        #expect(PromoteEntry(location: WorkspaceLocation(
            section: .sources, sourceId: sourceID, subjectId: "sub-1", subjectTypeKey: "participation",
            sourceSurface: .promote
        )) == nil)
        let e = entry(ref: "CPR-1", title: "  ")
        #expect(e.subjectName == "CPR-1")
        #expect(e.graphLocation == WorkspaceLocation(section: .sources, sourceId: sourceID, sourceSurface: .graph))
    }

    // MARK: Choosing

    @Test func nothingIsChosenAtFirst() {
        let (model, _) = makeModel(store: makeStore())
        #expect(model.choice == .none)
        #expect(!model.canAdvance)
        #expect(model.steps.count == 2)
        #expect(model.stepText == L10n.Promote.stepOf(current: 1, total: 2))
        #expect(model.hint == L10n.string(L10n.Promote.hintChoose(.person)))
        #expect(!model.hasUnsavedChoice)
    }

    @Test func newIsACompleteChoice() {
        let (model, _) = makeModel(store: makeStore())
        model.choose(.new)
        #expect(model.canAdvance)
        #expect(model.steps.count == 2)
        #expect(model.hint == L10n.string(L10n.Promote.hintNew(.person)))
        #expect(model.hasUnsavedChoice)
    }

    @Test func existingAddsCompareAndWaitsForAHandle() {
        let (model, _) = makeModel(store: makeStore())
        model.choose(.existing)
        #expect(model.steps.count == 3)
        #expect(!model.canAdvance)
        #expect(model.hint == L10n.string(L10n.Promote.hintChoose(.person)))
        model.select(PromoteModel.Target(entityID: "e1", ref: "PER-1", title: "James Robins", memberCount: 2))
        #expect(model.canAdvance)
        #expect(model.hint == L10n.Promote.hintExisting(
            name: "James Robins", ref: "PER-1", members: L10n.Promote.memberOther(count: 2)
        ))
        // Back to New drops the handle.
        model.choose(.new)
        #expect(model.target == nil)
        #expect(model.steps.count == 2)
    }

    @Test func memberCountsReadNaturally() {
        #expect(PromoteModel.members(1) == L10n.Promote.memberOne(count: 1))
        #expect(PromoteModel.members(3) == L10n.Promote.memberOther(count: 3))
    }

    @Test func eventCopyNamesEvents() {
        let (model, _) = makeModel(store: makeStore(), entry: entry(subjectID: "sub-e", kind: .event, ref: "CEV-1", title: "Birth"))
        #expect(model.steps.first == L10n.Promote.chooseStep(.event))
        #expect(model.hint == L10n.string(L10n.Promote.hintChoose(.event)))
        #expect(L10n.string(L10n.Promote.chooseStep(.event)) == "Choose an Event")
    }

    // MARK: Suggestions and search

    @Test func suggestionsLoadThroughTheirKey() async throws {
        let store = makeStore()
        let filed = try await existingPerson(store)
        let (model, _) = makeModel(store: store)
        let handle: QueryHandle<[CatalogPromoteTargetSuggestion]> = model.session.query(model.suggestionsKey)
        #expect(await waitUntil { handle.value != nil })
        let suggestion = try #require(handle.value?.first)
        #expect(suggestion.entity.ref == filed.entity.ref)
        let target = PromoteModel.target(for: suggestion)
        #expect(target.title == "James Robins")
        #expect(target.memberCount == 1)
    }

    @Test func searchAsksForTheSubjectsKindAndChoosesAHit() async throws {
        let store = makeStore()
        let filed = try await existingPerson(store)
        let (model, _) = makeModel(store: store)
        model.choose(.existing)
        await model.runSearch("Robins")
        #expect(model.searchResults.map(\.ref) == [filed.entity.ref])
        #expect(!model.searchFailed)
        model.selectSearchResult(filed.entity.id)
        #expect(model.target?.ref == filed.entity.ref)
        #expect(model.target?.memberCount == 1)
        #expect(model.searchSelection == filed.entity.id)
        // The chosen handle stays an option when the results move on.
        await model.runSearch("Nobody")
        #expect(model.searchResults.isEmpty)
        #expect(model.searchOptions.map(\.value) == [filed.entity.id])
    }

    @Test func debouncedQueryRunsTheSearch() async throws {
        let store = makeStore()
        _ = try await existingPerson(store)
        let (model, _) = makeModel(store: store)
        model.updateQuery("James")
        #expect(await waitUntil { !model.searchResults.isEmpty })
        model.updateQuery("   ")
        #expect(model.searchResults.isEmpty)
    }

    @Test func aFailedSearchIsReported() async {
        let store = makeStore()
        store.searchCatalogError = CocoaError(.fileReadUnknown)
        let (model, _) = makeModel(store: store)
        await model.runSearch("Robins")
        #expect(model.searchFailed)
        #expect(model.searchResults.isEmpty)
    }

    // MARK: Saving

    @Test func nextMintsReturnsToTheGraphAndRecounts() async {
        let store = makeStore()
        let counts = CatalogCounts(projectDir: projectDir, store: store)
        let (model, navigation) = makeModel(store: store, counts: counts)
        await counts.refreshAll()
        #expect(counts.persons == 0)
        model.choose(.new)
        #expect(await model.next())
        #expect(store.recordedCalls.contains("promoteSubject id=sub-1"))
        #expect(store.membershipBySubject["sub-1"]?.entity.ref.hasPrefix("PER-") == true)
        #expect(counts.persons == 1)
        #expect(model.saved)
        #expect(!model.hasUnsavedChoice)
        #expect(navigation.currentLocation == model.entry.graphLocation)
        #expect(navigation.heldNavigation == nil)
    }

    @Test func nextJoinsTheChosenHandle() async throws {
        let store = makeStore()
        let filed = try await existingPerson(store)
        let (model, navigation) = makeModel(store: store)
        model.select(PromoteModel.Target(entityID: filed.entity.id, ref: filed.entity.ref, title: "James Robins", memberCount: 1))
        #expect(await model.next())
        #expect(store.recordedCalls.contains("promoteSubject id=sub-1 entity=\(filed.entity.id)"))
        #expect(store.membershipBySubject["sub-1"]?.entity.id == filed.entity.id)
        #expect(navigation.currentLocation == model.entry.graphLocation)
    }

    @Test func savingInvalidatesTheGraph() async {
        let store = makeStore()
        let (model, _) = makeModel(store: store)
        let graph: QueryHandle<SourceGraphRows> = model.session.query(model.graphKey)
        #expect(await waitUntil { graph.value != nil })
        model.choose(.new)
        #expect(await model.next())
        let reloaded: QueryHandle<SourceGraphRows> = model.session.query(model.graphKey)
        #expect(await waitUntil { reloaded.value?.memberships.contains { $0.subjectID == "sub-1" } == true })
    }

    @Test func aFailedSaveKeepsTheStep() async {
        let store = makeStore()
        store.subjectsBySource[sourceID] = []
        let (model, navigation) = makeModel(store: store)
        let before = navigation.currentLocation
        model.choose(.new)
        #expect(!(await model.next()))
        #expect(model.saveError?.isEmpty == false)
        #expect(!model.saved)
        #expect(model.choice == .new)
        #expect(navigation.currentLocation == before)
    }

    // MARK: Leave guard

    @Test func leavingWithoutAChoiceDoesNotAsk() {
        let (model, navigation) = makeModel(store: makeStore())
        model.done()
        #expect(navigation.heldNavigation == nil)
        #expect(navigation.currentLocation == model.entry.graphLocation)
    }

    @Test func leavingWithAChoiceAsksAndCanStay() {
        let (model, navigation) = makeModel(store: makeStore())
        let here = navigation.currentLocation
        model.choose(.new)
        navigation.go(to: .sectionRoot(.persons))
        #expect(navigation.heldNavigation == .location(.sectionRoot(.persons)))
        #expect(model.pendingLeave?.navigation == .location(.sectionRoot(.persons)))
        #expect(model.leaveTitle == L10n.Promote.leaveTitle(name: "James Robins"))
        #expect(model.leaveMessage == L10n.Promote.leaveMessageNew(.person, subjectRef: "CPR-1"))
        model.keepPromoting()
        #expect(model.pendingLeave == nil)
        #expect(navigation.heldNavigation == nil)
        #expect(navigation.currentLocation == here)
    }

    @Test func leavingWithAChoiceAsksAndCanLeave() {
        let (model, navigation) = makeModel(store: makeStore())
        model.select(PromoteModel.Target(entityID: "e1", ref: "PER-1", title: "James Robins", memberCount: 2))
        #expect(model.leaveMessage == L10n.Promote.leaveMessageExisting(target: "PER-1", subjectRef: "CPR-1"))
        navigation.go(to: .sectionRoot(.metadata))
        #expect(model.pendingLeave != nil)
        model.leave()
        #expect(navigation.currentLocation == .sectionRoot(.metadata))
    }

    @Test func existingWithoutAHandleStillAsks() {
        let (model, navigation) = makeModel(store: makeStore())
        model.choose(.existing)
        model.done()
        #expect(model.pendingLeave != nil)
        #expect(model.leaveMessage == L10n.Promote.leaveMessageUnchosen(subjectRef: "CPR-1"))
        model.keepPromoting()
        #expect(navigation.heldNavigation == nil)
    }

    // MARK: Subject status

    @Test func subjectStatusFollowsTheGraph() {
        let (model, _) = makeModel(store: makeStore())
        let james = subject("sub-1", ref: "CPR-1", label: "James Robins")
        #expect(model.subjectStatus(rows: nil) == .loading)
        #expect(model.subjectStatus(rows: SourceGraphRows(sourceId: sourceID)) == .missing)
        #expect(model.subjectStatus(rows: SourceGraphRows(sourceId: sourceID, subjects: [james])) == .promotable)
        let member = CatalogSubjectMembership(
            subjectID: "sub-1", claimID: "c1",
            entity: CatalogCanonicalEntity(id: "e1", ref: "PER-1", subjectTypeID: "type-person", label: ""),
            kind: "person"
        )
        let promoted = SourceGraphRows(sourceId: sourceID, subjects: [james], memberships: [member])
        #expect(model.subjectStatus(rows: promoted) == .alreadyPromoted)
    }

    @Test func aGoneSubjectReturnsToTheGraphWithoutAsking() {
        let (model, navigation) = makeModel(store: makeStore())
        model.choose(.new)
        model.returnToGraph()
        #expect(navigation.heldNavigation == nil)
        #expect(navigation.currentLocation == model.entry.graphLocation)
    }
}
