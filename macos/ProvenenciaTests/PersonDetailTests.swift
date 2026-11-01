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
        summary: CatalogPersonSummary? = nil,
        derivedRows: [CatalogDerivedRow] = []
    ) -> CatalogConclusionDetail {
        let sex = D.field("", key: "sex_at_birth", label: "Sex at birth", valueType: "term")
        return CatalogConclusionDetail(
            entity: CatalogCanonicalEntity(id: "e1", ref: "PER-7KD45", subjectTypeID: "t", label: label),
            fields: (name.map { [$0] } ?? []) + [sex] + extra,
            memberCount: members,
            summary: summary.map { .person($0) },
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
        var personSummary = CatalogPersonSummary(
            entity: CatalogCanonicalEntity(id: "e1", ref: "PER-7KD45", subjectTypeID: "t", label: ""),
            name: nil, nameValueCount: 0
        )
        personSummary.gender = CatalogTermFacts(term: woman, count: 1, propertyLabel: "Gender")
        let content = PersonDetailContent(detail: detail(extra: [gender], summary: personSummary))
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

    @Test func treatmentReadsOnlyTheSummary() {
        let woman = CatalogConclusionValue.term(id: "t", key: "woman", label: "Woman", origin: CatalogOrigin.provenencia)
        let gender = D.field(
            "single", key: "gender", label: "Gender", valueType: "term",
            [D.value(1, "kept", woman)], outcomes: [D.outcome("g", "kept", recorded: woman)]
        )
        let content = PersonDetailContent(detail: detail(extra: [gender]))
        #expect(content.treatment == nil && content.badge == nil)
        #expect(content.rows.map(\.label).contains("Gender"))
    }

    @Test func headerThumbIsTheLeadLikeness() {
        let woman = CatalogConclusionValue.term(id: "t", key: "woman", label: "Woman", origin: CatalogOrigin.provenencia)
        let gender = D.field(
            "single", key: "gender", label: "Gender", valueType: "term",
            [D.value(1, "kept", woman)], outcomes: [D.outcome("g", "kept", recorded: woman)]
        )
        var summary = CatalogPersonSummary(
            entity: CatalogCanonicalEntity(id: "e1", ref: "PER-7KD45", subjectTypeID: "t", label: ""),
            name: nil, nameValueCount: 0
        )
        summary.pictureImageRelPath = "objects/ab/cd.png"
        summary.gender = CatalogTermFacts(term: woman, count: 1, propertyLabel: "Gender")
        var source = detail(extra: [gender], summary: summary)
        var lead = D.picture("a", path: "objects/ab/cd.png", title: "Register")
        lead.leads = true
        source.pictures = [lead]
        let thumb = PersonDetailContent(detail: source).thumbnail
        #expect(thumb.imageRelPath == "objects/ab/cd.png")
        #expect(thumb.imageLabel == "Likeness from Register")
        #expect(thumb.colors != nil)
        #expect(thumb.mark == .subjectPerson)
        #expect(thumb.label == L10n.Conclusions.personLikenessThumb)
        summary.pictureImageRelPath = ""
        let bare = PersonDetailContent(detail: detail(extra: [gender], summary: summary)).thumbnail
        #expect(bare.label == L10n.Conclusions.personNoLikeness)
    }

    @Test func likenessesBecomeTilesEarliestFirst() {
        var source = detail()
        var second = D.picture("b", path: "objects/ef/01.png", title: "")
        second.record.observationRef = "OBS-2"
        source.pictures = [D.picture("a", path: "objects/ab/cd.png", title: "Register"), second]
        let content = PersonDetailContent(detail: source)
        let gallery = content.gallery
        #expect(gallery.tiles.map(\.id) == ["a", "b"])
        #expect(gallery.tiles.map(\.sourceTitle) == ["Register", "OBS-2"])
        #expect(gallery.title == L10n.Conclusions.likenessesSection)
        #expect(gallery.tiles.map(\.accessibilityLabel) == [
            L10n.Conclusions.likenessTileLabel("Register"), L10n.Conclusions.likenessTileLabel("OBS-2"),
        ])
        #expect(gallery.choose == .personPortrait)
        let first = gallery.tiles[0].location
        #expect(first.citationId == "cit-a" && first.observationId == "a" && first.sourceSurface == .citationComposer)
        #expect(PersonDetailContent(detail: detail()).gallery.tiles.isEmpty)
        #expect(EventDetailContent(detail: detail()).gallery.tiles.isEmpty)
    }

    @Test func portraitFlagsFollowThePreferredLikeness() {
        func content(preferred: String?, lead: String) -> PersonDetailContent {
            var summary = CatalogPersonSummary(
                entity: CatalogCanonicalEntity(id: "e1", ref: "PER-7KD45", subjectTypeID: "t", label: ""),
                name: nil, nameValueCount: 0
            )
            summary.pictureImageRelPath = lead == "a" ? "objects/ab/cd.png" : "objects/ef/01.png"
            var a = D.picture("a", path: "objects/ab/cd.png", title: "Register")
            var b = D.picture("b", path: "objects/ef/01.png", title: "Census")
            a.preferred = preferred == "a"
            b.preferred = preferred == "b"
            a.leads = lead == "a"
            b.leads = lead == "b"
            var source = detail(summary: summary)
            source.pictures = [a, b]
            return PersonDetailContent(detail: source)
        }
        let neither = content(preferred: nil, lead: "a")
        #expect(neither.gallery.tiles.map(\.isLead) == [true, false])
        #expect(!neither.gallery.offersEarliest)

        let chosen = content(preferred: "b", lead: "b")
        #expect(chosen.gallery.tiles.map(\.isLead) == [false, true])
        #expect(chosen.gallery.offersEarliest)
        #expect(chosen.thumbnail.imageRelPath == "objects/ef/01.png")
        #expect(chosen.thumbnail.label == L10n.Conclusions.personLikenessThumb)
        #expect(chosen.thumbnail.imageLabel == L10n.Conclusions.likenessLabel("Census"))

        let earliest = content(preferred: "a", lead: "a")
        #expect(earliest.gallery.tiles.map(\.isLead) == [true, false])
        #expect(!earliest.gallery.offersEarliest)
        #expect(!EventDetailContent(detail: detail()).gallery.offersEarliest)
    }

    @MainActor
    private func leadModel(_ store: FakeStore) -> ConclusionLeadModel {
        let session = WorkspaceSession(projectKey: ProjectKey(projectDir: "/tmp/p.provenencia"), store: store)
        return ConclusionLeadModel(session: session, userID: "user-1", entityId: "per-1")
    }

    private func censusTile() -> ConclusionDetailPicture {
        ConclusionDetailGallery.tiles(
            from: {
                var source = detail()
                source.pictures = [D.picture("obs-b", path: "objects/ef/01.png", title: "Census")]
                return source
            }(),
            tileLabel: L10n.Conclusions.likenessTileLabel,
            useAsLabel: L10n.Conclusions.useAsPortraitLabel,
            chosenBody: L10n.Conclusions.toastPortraitChosenBody
        )[0]
    }

    @MainActor
    @Test func choosingAPortraitWritesAndToasts() async {
        let store = FakeStore()
        let model = leadModel(store)
        await model.choose(censusTile(), copy: .personPortrait)
        #expect(store.recordedCalls.last == "setPreferredLikeness entity=per-1 observation=obs-b")
        #expect(model.toast == VocabularyToast(
            title: "Portrait updated", body: "Using the likeness from Census as the portrait.", tone: .success
        ))
        #expect(!model.isSaving)
    }

    @MainActor
    @Test func useEarliestClearsAndToasts() async {
        let store = FakeStore()
        let model = leadModel(store)
        await model.useEarliest(copy: .personPortrait)
        #expect(store.recordedCalls.last == "setPreferredLikeness entity=per-1 observation=")
        #expect(model.toast == VocabularyToast(
            title: "Portrait updated", body: "Using the earliest likeness as the portrait.", tone: .success
        ))
    }

    @MainActor
    @Test func theLeadCopyIsReadAtCallTime() async {
        let store = FakeStore()
        let model = leadModel(store)
        var tile = censusTile()
        tile.chosenToastBody = L10n.Conclusions.toastThumbnailChosenBody("Census")
        await model.choose(tile, copy: .eventThumbnail)
        #expect(store.recordedCalls.last == "setPreferredLikeness entity=per-1 observation=obs-b")
        #expect(model.toast == VocabularyToast(
            title: "Thumbnail updated", body: "Using the picture from Census as the thumbnail.", tone: .success
        ))
        await model.useEarliest(copy: .eventThumbnail)
        #expect(store.recordedCalls.last == "setPreferredLikeness entity=per-1 observation=")
        #expect(model.toast == VocabularyToast(
            title: "Thumbnail updated", body: "Using the earliest picture as the thumbnail.", tone: .success
        ))
    }

    @MainActor
    @Test func portraitFailureToastsTheError() async {
        let store = FakeStore()
        let failure = CoreInvokeError.coded(status: 1, code: "canonicalentities.invalid", kind: .user, params: [])
        store.setPreferredLikenessError = failure
        let model = leadModel(store)
        await model.useEarliest(copy: .personPortrait)
        #expect(model.toast == VocabularyToast(title: L10n.Errors.message(for: failure), body: "", tone: .danger))
        #expect(!model.isSaving)
    }

    @Test func aMergedNameRowCarriesItsStateAndWhy() {
        let row = PersonDetailContent(detail: detail()).rows[0]
        #expect(row.lead == "James Robins" && row.style == .text)
        #expect(row.badge == .merged && row.count == "2 Sources" && row.against == nil)
        #expect(row.otherValuesLabel == nil && row.otherValues.isEmpty)
        #expect(row.whyTitle == "Why James Robins")
        let nick = CatalogNameValue(form: "unused", parts: [
            CatalogNameValuePart(value: "James", type: "given"),
            CatalogNameValuePart(value: "Jake", type: "nick"),
            CatalogNameValuePart(value: "Robins", type: "surname"),
        ])
        let nicked = PersonDetailContent(detail: detail(name: D.field("single", [D.value(1, "kept", .name(nick))],
            outcomes: [D.outcome("c", "kept", recorded: .name(nick))]))).rows[0]
        #expect(nicked.whyTitle == "Why James “Jake” Robins")
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
        let james = CatalogPersonSummary(
            entity: CatalogCanonicalEntity(id: "e1", ref: "PER-7KD45", subjectTypeID: "t", label: ""),
            name: CatalogNameValue(form: "James Robins"),
            nameValueCount: 1,
            birth: CatalogLifeFacts(
                event: birthEvent,
                date: CatalogDateValueInput(kind: "point", startYear: 1817),
                dateCount: 1,
                places: [CatalogChainedPlace(entity: york, names: ["York"])]
            ),
            death: CatalogLifeFacts(
                date: CatalogDateValueInput(kind: "point", startYear: 1880),
                dateCount: 1,
                places: [CatalogChainedPlace(entity: toronto, names: ["Toronto"])]
            )
        )
        #expect(PersonLifeDisplay.line(james, locale: en).text == "1817 – 1880 · York → Toronto")
        #expect(PersonLifeDisplay.line(james, locale: en).extraPlaces == 0)

        let page = PersonDetailContent(detail: detail(summary: james), locale: en)
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
        renamed.birth.places = [CatalogChainedPlace(entity: york, names: ["York", "Tkaronto"])]
        #expect(PersonLifeDisplay.line(renamed, locale: en).text == "1817 – 1880 · York → Toronto")
        #expect(PersonLifeDisplay.line(renamed, locale: en).extraPlaces == 0)
        let oneplace = PersonDetailContent(detail: detail(summary: renamed), locale: en)
        #expect(oneplace.rows[2].lead == "York" && oneplace.rows[2].badge == nil)

        // Two Places disagree: +1 and mixed, with no Why on this row.
        var twoPlaces = james
        let upper = CatalogCanonicalEntity(id: "pl3", ref: "PLC-3", subjectTypeID: "t", label: "")
        twoPlaces.birth.places = [
            CatalogChainedPlace(entity: york, names: ["York"]),
            CatalogChainedPlace(entity: upper, names: ["Kingston"]),
        ]
        #expect(PersonLifeDisplay.line(twoPlaces, locale: en).extraPlaces == 1)
        let mixed = PersonDetailContent(detail: detail(summary: twoPlaces), locale: en)
        #expect(mixed.rows[2].lead == "York" && mixed.rows[2].badge == .mixed && mixed.rows[2].records.isEmpty)
        #expect(mixed.rows[1].badge == nil, "one birth event with one date is not mixed")

        // Two birth events disagree: both birth rows are mixed.
        var twoBirths = james
        twoBirths.birth.eventCount = 2
        let competing = PersonDetailContent(detail: detail(summary: twoBirths), locale: en)
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
        event: CatalogEventSummary? = nil,
        place: CatalogChainedPlace? = nil
    ) -> CatalogOwnerField {
        CatalogOwnerField(
            entity: entity, event: event, place: place,
            field: D.field(values.isEmpty ? "" : "single", key: key, label: key, valueType: key == "date" ? "date" : "text", values, outcomes: outcomes)
        )
    }

    private func lifeSummary(birth: CatalogLifeFacts = CatalogLifeFacts(), death: CatalogLifeFacts = CatalogLifeFacts()) -> CatalogPersonSummary {
        CatalogPersonSummary(entity: entity("e1", "PER-7KD45"), name: CatalogNameValue(form: "Ada"), nameValueCount: 1, birth: birth, death: death)
    }

    @Test func lifeRowsCarryTheirOwnersRecords() {
        let event = entity("ev1", "EVT-1")
        let date = CatalogDateValueInput(kind: "point", startYear: 1817, startMonth: 5, startDay: 14)
        let summary = lifeSummary(birth: CatalogLifeFacts(event: event, eventCount: 1, date: date, dateCount: 1))
        let kept = [D.outcome("reg", "kept", title: "Register"), D.outcome("cen", "kept", title: "Census")]
        let rows = [CatalogDerivedRow(kind: .birthDate, owners: [
            owner(event, key: "date", outcomes: kept, values: [D.value(1, "kept", support: 2, .date(date))]),
        ])]
        let page = PersonDetailContent(detail: detail(summary: summary, derivedRows: rows), locale: en)
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
        let empty = PersonDetailContent(detail: detail(summary: lifeSummary(), derivedRows: denied), locale: en).rows[1]
        #expect(empty.lead == nil && empty.records.count == 1)
        #expect(empty.whyTitle == "Why “Birth date”")
    }

    @Test func anOwnerThatIsNotTheShownOneIsNamed() {
        // The lead event has no Place; another event is at York. The one
        // group is not the shown value, so it is named and not in primary ink.
        let york = entity("pl1", "PLC-1")
        let summary = lifeSummary(birth: CatalogLifeFacts(eventCount: 2))
        let rows = [CatalogDerivedRow(kind: .birthPlace, owners: [
            owner(entity("lk1", "LOC-1"), key: "place", outcomes: [D.outcome("y", "kept")], values: [D.value(1, "kept", .text("York"))],
                  place: CatalogChainedPlace(entity: york, names: ["York"])),
        ])]
        let birth = PersonDetailContent(detail: detail(summary: summary, derivedRows: rows), locale: en).rows[2]
        #expect(birth.recordGroups.count == 1)
        #expect(birth.recordGroups[0].owner?.title == "York")
        #expect(birth.recordGroups[0].records.allSatisfy { !$0.counted })
    }

    @Test func competingOwnersAreGroupedAndSecondary() {
        let first = entity("ev1", "EVT-4KX2M"), second = entity("ev2", "EVT-9RW7P")
        let date = CatalogDateValueInput(kind: "point", startYear: 1822)
        let summary = lifeSummary(birth: CatalogLifeFacts(event: second, eventCount: 2, date: date, dateCount: 1))
        func eventSummary(_ entity: CatalogCanonicalEntity) -> CatalogEventSummary {
            CatalogEventSummary(
                entity: entity, eventTypeKey: "birth", eventTypeOrigin: CatalogOrigin.provenencia,
                title: CatalogEventTitle(
                    rule: .subject, ref: entity.ref, typeKey: "birth", typeOrigin: CatalogOrigin.provenencia,
                    subjects: [CatalogNameValue(form: "Ada Lovelace")]
                )
            )
        }
        let rows = [CatalogDerivedRow(kind: .birthDate, owners: [
            owner(first, key: "date", outcomes: [D.outcome("a", "kept")], event: eventSummary(first)),
            owner(second, key: "date", outcomes: [D.outcome("b", "kept"), D.outcome("c", "kept")], event: eventSummary(second)),
        ])]
        let birth = PersonDetailContent(detail: detail(summary: summary, derivedRows: rows), locale: en).rows[1]
        // The summary's lead event (the second) is shown, so it comes first.
        #expect(birth.recordGroups.map(\.id) == ["ev2", "ev1"])
        #expect(birth.recordGroups[0].records.allSatisfy { $0.counted })
        #expect(birth.recordGroups[1].records.allSatisfy { !$0.counted })
        #expect(birth.recordGroups[0].owner?.title == "Birth of Ada Lovelace")
        #expect(birth.recordGroups[0].owner?.ref == "EVT-9RW7P")
        #expect(birth.recordGroups[0].owner?.spoken == "Birth of Ada Lovelace, EVT-9RW7P, shown value")

        // An owner with no summary and no label is titled by its ref, once.
        let bare = [CatalogDerivedRow(kind: .birthDate, owners: [
            owner(first, key: "date", outcomes: [D.outcome("a", "kept")]),
            owner(second, key: "date", outcomes: [D.outcome("b", "kept")]),
        ])]
        let line = PersonDetailContent(detail: detail(summary: summary, derivedRows: bare), locale: en).rows[1].recordGroups[1].owner
        #expect(line?.title == "EVT-4KX2M" && line?.ref == nil && line?.spoken == "EVT-4KX2M")

        // Two Places are titled as the Places list titles them.
        let toronto = entity("pl1", "PLC-1"), yorkville = entity("pl2", "PLC-2")
        let name = { (text: String) in [D.value(1, "kept", .text(text))] }
        let places = [CatalogDerivedRow(kind: .deathPlace, owners: [
            owner(entity("lk1", "LOC-1"), key: "place", outcomes: [D.outcome("t", "kept")], values: name("Toronto"),
                  place: CatalogChainedPlace(entity: toronto, names: ["Toronto"])),
            owner(entity("lk2", "LOC-2"), key: "place", outcomes: [D.outcome("y", "kept")], values: name("Yorkville"),
                  place: CatalogChainedPlace(entity: yorkville, names: ["Yorkville"])),
        ])]
        let death = lifeSummary(death: CatalogLifeFacts(places: [
            CatalogChainedPlace(entity: toronto, names: ["Toronto"]), CatalogChainedPlace(entity: yorkville, names: ["Yorkville"]),
        ]))
        let row = PersonDetailContent(detail: detail(summary: death, derivedRows: places), locale: en).rows[4]
        #expect(row.recordGroups.compactMap { $0.owner?.title } == ["Toronto", "Yorkville"])
        #expect(row.recordGroups[1].records.allSatisfy { !$0.counted })

        // A nameless first Place is not the shown one: the second leads the Why.
        let nameless = CatalogChainedPlace(entity: toronto, names: [])
        let paris = CatalogChainedPlace(entity: yorkville, names: ["Paris"])
        let lopsided = lifeSummary(death: CatalogLifeFacts(places: [nameless, paris]))
        let shown = PersonDetailContent(detail: detail(summary: lopsided, derivedRows: places), locale: en).rows[4]
        #expect(shown.lead == "Paris" && shown.recordGroups.first?.owner?.id == "pl2")
        #expect(shown.opens?.location == .placeDetail(entityId: "pl2", ref: "PLC-2", title: "Paris"))
    }

    // MARK: The provenance sheet (S13-11)

    @Test func aRecordOpensItsSourceCitationAndObservation() {
        let outcome = D.outcome("bible", "kept", source: "S1", title: "Family Bible")
        let record = ReconciliationRecord(outcome: outcome, in: D.field("single"), locale: en)
        #expect(record.sourceLocation == .sourcePage(sourceId: "S1", ref: nil, title: "Family Bible"))
        #expect(record.citationLocation.sourceSurface == .citationComposer)
        #expect(record.citationLocation.citationId == "cit-bible" && record.citationLocation.observationId == nil)
        #expect(record.location.citationId == "cit-bible" && record.location.observationId == "bible")
        #expect(record.observationRef == "OBS-bible" && record.subjectRef == "CPR-bible")
    }

    @Test func theShownOwnerIsSpokenAsShown() {
        let first = entity("ev1", "EVT-4KX2M"), second = entity("ev2", "EVT-9PQ2R")
        let summary = lifeSummary(birth: CatalogLifeFacts(event: first, eventCount: 2))
        func eventSummary(_ entity: CatalogCanonicalEntity) -> CatalogEventSummary {
            CatalogEventSummary(
                entity: entity, eventTypeKey: "birth", eventTypeOrigin: CatalogOrigin.provenencia,
                title: CatalogEventTitle(
                    rule: .subject, ref: entity.ref, typeKey: "birth", typeOrigin: CatalogOrigin.provenencia,
                    subjects: [CatalogNameValue(form: "Ada Lovelace")]
                )
            )
        }
        let rows = [CatalogDerivedRow(kind: .birthDate, owners: [
            owner(second, key: "date", outcomes: [D.outcome("b", "kept")], event: eventSummary(second)),
            owner(first, key: "date", outcomes: [D.outcome("a", "kept")], event: eventSummary(first)),
        ])]
        let birth = PersonDetailContent(detail: detail(summary: summary, derivedRows: rows), locale: en).rows[1]
        #expect(birth.recordGroups.map(\.id) == ["ev1", "ev2"])
        #expect(birth.recordGroups[0].owner?.shown == true)
        #expect(birth.recordGroups[0].owner?.spoken == "Birth of Ada Lovelace, EVT-4KX2M, shown value")
        #expect(birth.recordGroups[1].owner?.shown == false)
        #expect(birth.recordGroups[1].owner?.spoken == "Birth of Ada Lovelace, EVT-9PQ2R")
    }

    @Test func oneOwnerHasNoGroupLine() {
        let event = entity("ev1", "EVT-1")
        let summary = lifeSummary(birth: CatalogLifeFacts(event: event, eventCount: 1))
        let rows = [CatalogDerivedRow(kind: .birthDate, owners: [
            owner(event, key: "date", outcomes: [D.outcome("a", "kept")]),
        ])]
        let birth = PersonDetailContent(detail: detail(summary: summary, derivedRows: rows), locale: en).rows[1]
        #expect(birth.recordGroups.count == 1 && birth.recordGroups[0].owner == nil)

        // The shown owner has no records; the one other owner is not the shown one.
        let other = entity("ev2", "EVT-2")
        let lopsided = lifeSummary(birth: CatalogLifeFacts(event: event, eventCount: 2))
        let others = [CatalogDerivedRow(kind: .birthDate, owners: [
            owner(other, key: "date", outcomes: [D.outcome("b", "kept")]),
        ])]
        let row = PersonDetailContent(detail: detail(summary: lopsided, derivedRows: others), locale: en).rows[1]
        #expect(row.recordGroups.count == 1 && row.recordGroups[0].owner?.shown == false)
    }

    @Test func onlyRowsWithRecordsOfferProvenance() {
        let page = PersonDetailContent(detail: detail(summary: lifeSummary()), locale: en)
        #expect(page.rows[1].records.isEmpty)

        let event = entity("ev1", "EVT-1")
        let denied = [CatalogDerivedRow(kind: .birthDate, owners: [
            owner(event, key: "date", outcomes: [D.outcome("den", "denied", rank: nil)]),
        ])]
        let birth = PersonDetailContent(detail: detail(summary: lifeSummary(), derivedRows: denied), locale: en).rows[1]
        #expect(birth.lead == nil && birth.records.count == 1)
        #expect(ProvenanceSheetItem(id: birth.id).row(in: [birth]) == birth)
    }

    @Test func aRecordNamesItsCitation() {
        var outcome = D.outcome("reg", "kept", title: "Register of Baptisms")
        outcome.artifactLabel = "Vol. 3"
        outcome.citationLabel = "Baptism of Ada"
        outcome.artifactPage = 37
        #expect(CitationLineDisplay.line(outcome) == "Register of Baptisms, Vol. 3, Baptism of Ada, image 37")
        outcome.pageLabel = " 23 "
        #expect(CitationLineDisplay.line(outcome) == "Register of Baptisms, Vol. 3, Baptism of Ada, p. 23")
        let bare = D.outcome("reg", "kept", title: "Register of Baptisms")
        #expect(CitationLineDisplay.line(bare) == "Register of Baptisms")
        let untitled = D.outcome("reg", "kept", title: "")
        #expect(CitationLineDisplay.line(untitled) == "OBS-reg")
        var labelled = untitled
        labelled.artifactLabel = "Vol. 3"
        #expect(CitationLineDisplay.line(labelled) == "OBS-reg, Vol. 3")
    }
}
