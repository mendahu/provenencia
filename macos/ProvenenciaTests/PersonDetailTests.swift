import Foundation
import Testing
@testable import Provenencia

struct PersonDetailTests {
    private typealias D = ReconciledValueDisplayTests

    private func detail(
        name: CatalogConclusionField? = D.mergedName,
        label: String = "",
        members: Int = 4,
        extra: [CatalogConclusionField] = [],
        header: CatalogPersonHeader? = nil,
        derivedRows: [CatalogDerivedRow] = []
    ) -> CatalogConclusionDetail {
        let sex = D.field("", key: "sex_at_birth", label: "Sex at birth", valueType: "term")
        return CatalogConclusionDetail(
            entity: CatalogCanonicalEntity(id: "e1", ref: "PER-7KD45", subjectTypeID: "t", label: label),
            fields: (name.map { [$0] } ?? []) + [sex] + extra,
            memberCount: members,
            header: header.map { .person($0) },
            derivedRows: derivedRows
        )
    }

    @Test func presentationFollowsTheHandle() {
        #expect(ConclusionDetailPresentation<PersonDetailContent>(
            .loading, isFetching: true, content: { PersonDetailContent(detail: $0) }
        ) == .loading)
        let failed = ConclusionDetailPresentation<PersonDetailContent>(
            .failed(CoreInvokeError.coded(status: 1, code: "conclusiondetails.not_found", kind: .user, params: [])),
            isFetching: false,
            content: { PersonDetailContent(detail: $0) }
        )
        #expect(failed == .failed("This record no longer exists. It may have been merged into another."))
        let content = PersonDetailContent(detail: detail())
        #expect(ConclusionDetailPresentation<PersonDetailContent>(
            .loaded(detail()), isFetching: false, content: { PersonDetailContent(detail: $0) }
        ) == .content(content, refreshing: false))
        #expect(ConclusionDetailPresentation<PersonDetailContent>(
            .loaded(detail()), isFetching: true, content: { PersonDetailContent(detail: $0) }
        ) == .content(content, refreshing: true))
    }

    /// Name and the life rows always show; any other field only once a
    /// record speaks to it.
    @Test func rowsAreNameThenLifeThenRecordedFields() {
        let weak = D.outcome("w", "weak", rank: 1, recorded: .integer(8))
        let content = PersonDetailContent(detail: detail(extra: [
            D.field("", key: "occupation", label: "Occupation", valueType: "text", outcomes: [D.outcome("o", "no_evidence", rank: nil)]),
            D.field("single", key: "birth_weight", label: "Birth weight", valueType: "integer", [D.value(1, "weak", .integer(8))], outcomes: [weak]),
            D.field("", key: "height", label: "Height", valueType: "integer"),
        ]))
        // Sex at birth and Height have no records, so they don't show.
        #expect(content.rows.map(\.label) == [
            "Name", "Birth date", "Birth place", "Death date", "Death place", "Occupation", "Birth weight",
        ])
        #expect(content.rows[1...4].map(\.style) == [.date, .place, .date, .place])
        #expect(content.rows[1...4].map(\.emptyText) == [
            "No birth date recorded", "No birth place recorded", "No death date recorded", "No death place recorded",
        ])
        #expect(content.rows[1...4].allSatisfy { $0.lead == nil && $0.records.isEmpty && $0.count == nil })
        // A recorded field with nothing to show says so, and keeps its Why.
        #expect(content.rows[5].lead == nil && content.rows[5].emptyText == "Nothing recorded" && content.rows[5].records.count == 1)
    }

    @Test func treatmentRidesOnContentAndRowsStay() {
        let woman = CatalogConclusionValue.term(id: "t", key: "woman", label: "Woman", origin: CatalogOrigin.provenencia)
        let gender = D.field(
            "single", key: "gender", label: "Gender", valueType: "term",
            [D.value(1, "kept", woman)], outcomes: [D.outcome("g", "kept", recorded: woman)]
        )
        let content = PersonDetailContent(detail: detail(extra: [gender]))
        #expect(content.treatment?.family == .woman)
        #expect(content.thumbnail.colors != nil)
        #expect(content.badge?.text == "Woman")
        #expect(content.rows.map(\.label).contains("Gender"))
        let plain = PersonDetailContent(detail: detail())
        #expect(plain.treatment == nil && plain.badge == nil)
        #expect(plain.thumbnail == .init(mark: .subjectPerson, label: L10n.Conclusions.personNoLikeness))
        #expect(plain.rows.map(\.label) == ["Name", "Birth date", "Birth place", "Death date", "Death place"])
        let event = EventDetailContent(detail: detail())
        #expect(event.thumbnail == .init(mark: .subjectEvent, label: L10n.Conclusions.eventNoImage) && event.badge == nil)
        let place = PlaceDetailContent(detail: detail())
        #expect(place.thumbnail == .init(mark: .subjectPlace, label: L10n.Conclusions.placeNoImage) && place.badge == nil)
    }

    private func likeness(_ id: String, path: String, title: String) -> CatalogPicture {
        CatalogPicture(imageRelPath: path, record: D.outcome(id, "kept", title: title))
    }

    @Test func headerThumbIsTheLeadLikeness() {
        let woman = CatalogConclusionValue.term(id: "t", key: "woman", label: "Woman", origin: CatalogOrigin.provenencia)
        let gender = D.field(
            "single", key: "gender", label: "Gender", valueType: "term",
            [D.value(1, "kept", woman)], outcomes: [D.outcome("g", "kept", recorded: woman)]
        )
        var header = CatalogPersonHeader(
            entity: CatalogCanonicalEntity(id: "e1", ref: "PER-7KD45", subjectTypeID: "t", label: ""),
            name: nil, nameValueCount: 0
        )
        header.pictureImageRelPath = "objects/ab/cd.png"
        var source = detail(extra: [gender], header: header)
        source.pictures = [likeness("a", path: "objects/ab/cd.png", title: "Register")]
        let thumb = PersonDetailContent(detail: source).thumbnail
        #expect(thumb.imageRelPath == "objects/ab/cd.png")
        #expect(thumb.imageLabel == "Likeness from Register")
        #expect(thumb.colors != nil)
        #expect(thumb.mark == .subjectPerson)
        #expect(thumb.label == L10n.Conclusions.personLikenessThumb)
        header.pictureImageRelPath = ""
        let bare = PersonDetailContent(detail: detail(extra: [gender], header: header)).thumbnail
        #expect(bare.label == L10n.Conclusions.personNoLikeness)
    }

    @Test func likenessesBecomeTilesEarliestFirst() {
        var source = detail()
        var second = likeness("b", path: "objects/ef/01.png", title: "")
        second.record.observationRef = "OBS-2"
        source.pictures = [likeness("a", path: "objects/ab/cd.png", title: "Register"), second]
        let content = PersonDetailContent(detail: source)
        #expect(content.likenesses.map(\.id) == ["a", "b"])
        #expect(content.likenesses.map(\.sourceTitle) == ["Register", "OBS-2"])
        let first = content.likenesses[0].location
        #expect(first.citationId == "cit-a" && first.observationId == "a" && first.sourceSurface == .citationComposer)
        #expect(PersonDetailContent(detail: detail()).likenesses.isEmpty)
        #expect(EventDetailContent(detail: detail()).likenesses.isEmpty)
    }

    @Test func portraitFlagsFollowThePreferredLikeness() {
        func content(preferred: String?, lead: String) -> PersonDetailContent {
            var header = CatalogPersonHeader(
                entity: CatalogCanonicalEntity(id: "e1", ref: "PER-7KD45", subjectTypeID: "t", label: ""),
                name: nil, nameValueCount: 0
            )
            header.pictureImageRelPath = lead == "a" ? "objects/ab/cd.png" : "objects/ef/01.png"
            var a = likeness("a", path: "objects/ab/cd.png", title: "Register")
            var b = likeness("b", path: "objects/ef/01.png", title: "Census")
            a.preferred = preferred == "a"
            b.preferred = preferred == "b"
            a.leads = lead == "a"
            b.leads = lead == "b"
            var source = detail(header: header)
            source.pictures = [a, b]
            return PersonDetailContent(detail: source)
        }
        let neither = content(preferred: nil, lead: "a")
        #expect(neither.likenesses.map(\.isPortrait) == [true, false])
        #expect(!neither.thumbnail.offersEarliest)

        let chosen = content(preferred: "b", lead: "b")
        #expect(chosen.likenesses.map(\.isPortrait) == [false, true])
        #expect(chosen.thumbnail.offersEarliest)
        #expect(chosen.thumbnail.imageLabel == "Likeness from Census")

        let earliest = content(preferred: "a", lead: "a")
        #expect(earliest.likenesses.map(\.isPortrait) == [true, false])
        #expect(!earliest.thumbnail.offersEarliest)
        #expect(!EventDetailContent(detail: detail()).thumbnail.offersEarliest)
    }

    @MainActor
    private func portraitModel(_ store: FakeStore) -> ConclusionPortraitModel {
        let session = WorkspaceSession(projectKey: ProjectKey(projectDir: "/tmp/p.provenencia"), store: store)
        return ConclusionPortraitModel(session: session, userID: "user-1", entityId: "per-1")
    }

    @MainActor
    @Test func choosingAPortraitWritesAndToasts() async {
        let store = FakeStore()
        let model = portraitModel(store)
        let tile = ConclusionDetailLikeness(
            id: "obs-b", imageRelPath: "objects/ef/01.png", sourceTitle: "Census",
            location: ReconciliationRecord.composerLocation(for: D.outcome("obs-b", "kept", title: "Census"))
        )
        await model.choose(tile)
        #expect(store.recordedCalls.last == "setPreferredLikeness entity=per-1 observation=obs-b")
        #expect(model.toast == VocabularyToast(
            title: "Portrait updated", body: "Using the likeness from Census as the portrait.", tone: .success
        ))
        #expect(!model.isSaving)
    }

    @MainActor
    @Test func useEarliestClearsAndToasts() async {
        let store = FakeStore()
        let model = portraitModel(store)
        await model.useEarliest()
        #expect(store.recordedCalls.last == "setPreferredLikeness entity=per-1 observation=")
        #expect(model.toast == VocabularyToast(
            title: "Portrait updated", body: "Using the earliest likeness as the portrait.", tone: .success
        ))
    }

    @MainActor
    @Test func portraitFailureToastsTheError() async {
        let store = FakeStore()
        let failure = CoreInvokeError.coded(status: 1, code: "canonicalentities.invalid", kind: .user, params: [])
        store.setPreferredLikenessError = failure
        let model = portraitModel(store)
        await model.useEarliest()
        #expect(model.toast == VocabularyToast(title: L10n.Errors.message(for: failure), body: "", tone: .danger))
        #expect(!model.isSaving)
    }

    @Test func aMergedNameRowCarriesItsStateAndWhy() {
        let row = PersonDetailContent(detail: detail()).rows[0]
        #expect(row.lead == "James Robins" && row.style == .text)
        #expect(row.badge == .merged && row.count == "2 Sources" && row.against == nil)
        #expect(row.otherValuesLabel == nil && row.otherValues.isEmpty)
        #expect(row.whyTitle == "Why “James Robins”")
        #expect(row.records.map(\.phrase) == [
            "kept", "folded into James Robins", "outvoted (2 of 3 Sources)", "weak · low-trust Source",
        ])
        #expect(row.records.map(\.sourceTitle).first == "Census of Canada West, 1851")
        #expect(row.accessibilityLabel == "Name, James Robins, merged from 2 Sources")
    }

    @Test func aMixedRowDisclosesItsOtherValues() {
        let date = CatalogDateValueInput(kind: "point", startYear: 1878)
        let mixed = D.field("mixed", key: "x", label: "Occupation", valueType: "text", [
            D.value(1, "kept", support: 2, .text("farmer")), D.value(2, "kept", .text("labourer")),
        ])
        let row = ReconciledValueRowModel(field: mixed)
        #expect(row.lead == "farmer" && row.badge == .mixed)
        #expect(row.otherValuesLabel == "1 other value")
        #expect(row.otherValues == [.init(rank: 2, text: "labourer", support: "1 Source")])
        let dated = ReconciledValueRowModel(field: D.field("single", key: "d", label: "Date", valueType: "date", [D.value(1, "kept", .date(date))]))
        #expect(dated.style == .date)
    }

    @Test func headerFallsBackNameLabelRef() {
        let named = PersonDetailContent(detail: detail())
        #expect(named.title == .name("James Robins") && named.showsRef && named.members == "4 members")
        let labelled = PersonDetailContent(detail: detail(name: D.field("", [], outcomes: []), label: "Grandpa", members: 1))
        #expect(labelled.title == .label("Grandpa") && labelled.showsRef && labelled.members == "1 member")
        let bare = PersonDetailContent(detail: detail(name: nil, members: 1))
        #expect(bare.title == .ref("PER-7KD45") && !bare.showsRef)
        // No name field at all: the Name row still shows, empty.
        #expect(bare.rows.first?.label == "Name" && bare.rows.first?.lead == nil)
        #expect(bare.rows.first?.emptyText == "No name recorded")
        #expect(bare.rows.first?.accessibilityLabel == "Name, empty")
    }

    @Test func lifeLineReadsBirthAndDeath() {
        let en = Locale(identifier: "en_US")
        let birthEvent = CatalogCanonicalEntity(id: "ev1", ref: "EVT-1", subjectTypeID: "t", label: "")
        let york = CatalogCanonicalEntity(id: "pl1", ref: "PLC-1", subjectTypeID: "t", label: "")
        let toronto = CatalogCanonicalEntity(id: "pl2", ref: "PLC-2", subjectTypeID: "t", label: "")
        let james = CatalogPersonHeader(
            entity: CatalogCanonicalEntity(id: "e1", ref: "PER-7KD45", subjectTypeID: "t", label: ""),
            name: CatalogNameValue(form: "James Robins"),
            nameValueCount: 1,
            birth: CatalogLifeFacts(
                event: birthEvent,
                date: CatalogDateValueInput(kind: "point", startYear: 1817),
                dateCount: 1,
                places: [CatalogHeaderPlace(entity: york, names: ["York"])]
            ),
            death: CatalogLifeFacts(
                date: CatalogDateValueInput(kind: "point", startYear: 1880),
                dateCount: 1,
                places: [CatalogHeaderPlace(entity: toronto, names: ["Toronto"])]
            )
        )
        #expect(PersonLifeDisplay.line(james, locale: en).text == "1817 – 1880 · York → Toronto")
        #expect(PersonLifeDisplay.line(james, locale: en).extraPlaces == 0)

        let page = PersonDetailContent(detail: detail(header: james), locale: en)
        #expect(page.lifeLines.map(\.date) == ["1817", "1880"])
        #expect(page.lifeLines.map(\.place) == ["York", "Toronto"])
        #expect(page.rows[1].lead == "1817" && page.rows[1].badge == nil && page.rows[1].records.isEmpty)
        #expect(page.rows[2].lead == "York" && page.rows[2].style == .place)
        #expect(page.rows[4].lead == "Toronto")
        // Derived rows open the Event or Place that owns their Why.
        #expect(page.rows[1].opens?.location == .eventDetail(entityId: "ev1", ref: "EVT-1", title: nil))
        #expect(page.rows[2].opens?.location == .placeDetail(entityId: "pl1", ref: "PLC-1", title: "York"))
        #expect(page.rows[2].opens?.label == L10n.Conclusions.openPlace("York"))
        #expect(page.rows[3].opens == nil, "a death with no linked event opens nothing")
        #expect(page.rows[4].opens?.location == .placeDetail(entityId: "pl2", ref: "PLC-2", title: "Toronto"))

        var loneBirth = james
        loneBirth.death = CatalogLifeFacts()
        #expect(PersonLifeDisplay.line(loneBirth, locale: en).text == "1817 – · York")

        // A death with no birth is marked as a death, never read as a birth.
        var loneDeath = james
        loneDeath.birth = CatalogLifeFacts()
        #expect(PersonLifeDisplay.line(loneDeath, locale: en).text == "– 1880 · → Toronto")

        // Two names of one Place are one place: no +N, not mixed.
        var renamed = james
        renamed.birth.places = [CatalogHeaderPlace(entity: york, names: ["York", "Tkaronto"])]
        #expect(PersonLifeDisplay.line(renamed, locale: en).text == "1817 – 1880 · York → Toronto")
        #expect(PersonLifeDisplay.line(renamed, locale: en).extraPlaces == 0)
        let oneplace = PersonDetailContent(detail: detail(header: renamed), locale: en)
        #expect(oneplace.rows[2].lead == "York" && oneplace.rows[2].badge == nil)

        // Two Places disagree: +1 and mixed, with no Why on this row.
        var twoPlaces = james
        let upper = CatalogCanonicalEntity(id: "pl3", ref: "PLC-3", subjectTypeID: "t", label: "")
        twoPlaces.birth.places = [
            CatalogHeaderPlace(entity: york, names: ["York"]),
            CatalogHeaderPlace(entity: upper, names: ["Kingston"]),
        ]
        #expect(PersonLifeDisplay.line(twoPlaces, locale: en).extraPlaces == 1)
        let mixed = PersonDetailContent(detail: detail(header: twoPlaces), locale: en)
        #expect(mixed.rows[2].lead == "York" && mixed.rows[2].badge == .mixed && mixed.rows[2].records.isEmpty)
        #expect(mixed.rows[1].badge == nil, "one birth event with one date is not mixed")

        // Two birth events disagree: both birth rows are mixed.
        var twoBirths = james
        twoBirths.birth.eventCount = 2
        let competing = PersonDetailContent(detail: detail(header: twoBirths), locale: en)
        #expect(competing.rows[1].badge == .mixed && competing.rows[2].badge == .mixed)
        #expect(competing.rows[3].badge == nil && competing.rows[4].badge == nil)
    }

    @Test func lifeLinesSayUnknownUntilLifeEventsArrive() {
        let lines = PersonDetailContent(detail: detail()).lifeLines
        #expect(lines.map(\.abbreviation) == ["b.", "d."])
        #expect(lines.allSatisfy { $0.date == nil && $0.place == nil })
        #expect(lines.map(\.accessibilityLabel) == ["Born date unknown, place unknown", "Died date unknown, place unknown"])
    }

    @Test func aSourceOpensTheRecordsCitationInTheComposer() throws {
        let outcome = D.outcome("bible", "outvoted", rank: 2, source: "s-bible", title: "Family Bible")
        let location = ReconciliationRecord.composerLocation(for: outcome)
        #expect(location.section == .sources && location.sourceSurface == .citationComposer)
        let entry = try #require(CitationComposerEntry(location: location))
        guard case let .edit(sourceID, graphID, subjectID, citationID, artifactID, observationID) = entry else {
            Issue.record("expected an edit entry, got \(entry)")
            return
        }
        #expect(sourceID == "s-bible" && graphID == "g-bible" && subjectID == "sub-bible" && citationID == "cit-bible")
        #expect(location.graphId == "g-bible")
        #expect(artifactID == "art-bible" && observationID == "bible")
    }

    // MARK: Life rows carry their owners' records (S13-06)

    private let en = Locale(identifier: "en_US")

    private func entity(_ id: String, _ ref: String, label: String = "") -> CatalogCanonicalEntity {
        CatalogCanonicalEntity(id: id, ref: ref, subjectTypeID: "t", label: label)
    }

    private func owner(
        _ entity: CatalogCanonicalEntity,
        key: String,
        outcomes: [CatalogReconcilerOutcome],
        values: [CatalogReconciledValue] = [],
        event: CatalogEventHeader? = nil
    ) -> CatalogOwnerField {
        CatalogOwnerField(
            entity: entity, event: event,
            field: D.field(values.isEmpty ? "" : "single", key: key, label: key, valueType: key == "date" ? "date" : "text", values, outcomes: outcomes)
        )
    }

    private func lifeHeader(birth: CatalogLifeFacts = CatalogLifeFacts(), death: CatalogLifeFacts = CatalogLifeFacts()) -> CatalogPersonHeader {
        CatalogPersonHeader(entity: entity("e1", "PER-7KD45"), name: CatalogNameValue(form: "Ada"), nameValueCount: 1, birth: birth, death: death)
    }

    @Test func lifeRowsCarryTheirOwnersRecords() {
        let event = entity("ev1", "EVT-1")
        let date = CatalogDateValueInput(kind: "point", startYear: 1817, startMonth: 5, startDay: 14)
        let header = lifeHeader(birth: CatalogLifeFacts(event: event, eventCount: 1, date: date, dateCount: 1))
        let kept = [D.outcome("reg", "kept", title: "Register"), D.outcome("cen", "kept", title: "Census")]
        let rows = [CatalogDerivedRow(kind: .birthDate, owners: [
            owner(event, key: "date", outcomes: kept, values: [D.value(1, "kept", support: 2, .date(date))]),
        ])]
        let page = PersonDetailContent(detail: detail(header: header, derivedRows: rows), locale: en)
        let birth = page.rows[1]
        #expect(birth.recordGroups.count == 1 && birth.recordGroups[0].owner == nil)
        #expect(birth.records.count == 2 && birth.records.allSatisfy { $0.counted })
        #expect(birth.whyTitle == "Why “\(birth.lead ?? "")”" && birth.lead != nil)
        // No death event: no Why.
        #expect(page.rows[3].records.isEmpty && page.rows[3].whyTitle == "")

        // A date whose only record is denied: stated empty, with its Why.
        let denied = [CatalogDerivedRow(kind: .birthDate, owners: [
            owner(event, key: "date", outcomes: [D.outcome("den", "denied", rank: nil)]),
        ])]
        let empty = PersonDetailContent(detail: detail(header: lifeHeader(), derivedRows: denied), locale: en).rows[1]
        #expect(empty.lead == nil && empty.records.count == 1)
        #expect(empty.whyTitle == "Why “Birth date”")
    }

    @Test func competingOwnersAreGroupedAndSecondary() {
        let first = entity("ev1", "EVT-4KX2M"), second = entity("ev2", "EVT-9RW7P")
        let date = CatalogDateValueInput(kind: "point", startYear: 1822)
        let header = lifeHeader(birth: CatalogLifeFacts(event: second, eventCount: 2, date: date, dateCount: 1))
        func eventHeader(_ entity: CatalogCanonicalEntity) -> CatalogEventHeader {
            CatalogEventHeader(
                entity: entity, eventTypeKey: "birth", eventTypeOrigin: CatalogOrigin.provenencia,
                title: CatalogEventTitle(
                    rule: .subject, ref: entity.ref, typeKey: "birth", typeOrigin: CatalogOrigin.provenencia,
                    subjects: [CatalogNameValue(form: "Ada Lovelace")]
                )
            )
        }
        let rows = [CatalogDerivedRow(kind: .birthDate, owners: [
            owner(first, key: "date", outcomes: [D.outcome("a", "kept")], event: eventHeader(first)),
            owner(second, key: "date", outcomes: [D.outcome("b", "kept"), D.outcome("c", "kept")], event: eventHeader(second)),
        ])]
        let birth = PersonDetailContent(detail: detail(header: header, derivedRows: rows), locale: en).rows[1]
        // The header's lead event (the second) is shown, so it comes first.
        #expect(birth.recordGroups.map(\.id) == ["ev2", "ev1"])
        #expect(birth.recordGroups[0].records.allSatisfy { $0.counted })
        #expect(birth.recordGroups[1].records.allSatisfy { !$0.counted })
        #expect(birth.recordGroups[0].owner?.title == "Birth of Ada Lovelace")
        #expect(birth.recordGroups[0].owner?.ref == "EVT-9RW7P")
        #expect(birth.recordGroups[0].owner?.spoken == "Birth of Ada Lovelace, EVT-9RW7P")

        // An owner with no header and no label is titled by its ref, once.
        let bare = [CatalogDerivedRow(kind: .birthDate, owners: [
            owner(first, key: "date", outcomes: [D.outcome("a", "kept")]),
            owner(second, key: "date", outcomes: [D.outcome("b", "kept")]),
        ])]
        let line = PersonDetailContent(detail: detail(header: header, derivedRows: bare), locale: en).rows[1].recordGroups[1].owner
        #expect(line?.title == "EVT-4KX2M" && line?.ref == nil && line?.spoken == "EVT-4KX2M")

        // Two Places are titled as the Places list titles them.
        let toronto = entity("pl1", "PLC-1"), yorkville = entity("pl2", "PLC-2")
        let name = { (text: String) in [D.value(1, "kept", .text(text))] }
        let places = [CatalogDerivedRow(kind: .deathPlace, owners: [
            owner(toronto, key: "toponym", outcomes: [D.outcome("t", "kept")], values: name("Toronto")),
            owner(yorkville, key: "toponym", outcomes: [D.outcome("y", "kept")], values: name("Yorkville")),
        ])]
        let death = lifeHeader(death: CatalogLifeFacts(places: [
            CatalogHeaderPlace(entity: toronto, names: ["Toronto"]), CatalogHeaderPlace(entity: yorkville, names: ["Yorkville"]),
        ]))
        let row = PersonDetailContent(detail: detail(header: death, derivedRows: places), locale: en).rows[4]
        #expect(row.recordGroups.compactMap { $0.owner?.title } == ["Toronto", "Yorkville"])
        #expect(row.recordGroups[1].records.allSatisfy { !$0.counted })

        // A nameless first Place is not the shown one: the second leads the Why.
        let nameless = CatalogHeaderPlace(entity: toronto, names: [])
        let paris = CatalogHeaderPlace(entity: yorkville, names: ["Paris"])
        let lopsided = lifeHeader(death: CatalogLifeFacts(places: [nameless, paris]))
        let shown = PersonDetailContent(detail: detail(header: lopsided, derivedRows: places), locale: en).rows[4]
        #expect(shown.lead == "Paris" && shown.recordGroups.first?.owner?.id == "pl2")
        #expect(shown.opens?.location == .placeDetail(entityId: "pl2", ref: "PLC-2", title: "Paris"))
    }
}
