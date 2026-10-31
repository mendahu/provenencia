import Foundation

extension CitationComposerModel {
    var orderedCitations: [CatalogListedCitation] {
        CitationComposerAuthoring.stepperOrder(listedCitations)
    }

    var subjectGroups: [CitationSubjectGroup] {
        CitationComposerAuthoring.groups(
            observations: observationRows.rows,
            connections: connections.rows,
            vocabulary: vocabulary,
            nameOrders: session.nameFormat
        )
    }

    var savedLineCount: Int {
        let properties = observationRows.rows.filter { $0.persistedID != nil }.count
        let bridges = connections.rows.filter { !$0.isPending && !($0.bridgeID ?? "").isEmpty }.count
        return properties + bridges
    }

    var observationsMeta: String {
        "\(savedLineCount)"
    }

    var authoringCanConfirm: Bool {
        guard !isSwitchingCitation, !depictionImageBlocksSave else { return false }
        switch authoring {
        case .rest:
            return false
        case .property(let draft):
            return !draft.isSaving && Self.propertyValuesComplete(draft.values, vocabulary: vocabulary)
        case .bridge(let draft):
            let pair = CitationComposerAuthoring.resolvePair(
                endA: draft.endA,
                endB: draft.endB,
                termID: draft.termID,
                vocabulary: vocabulary
            )
            return !draft.isSaving && Self.bridgeReady(pair, termID: draft.termID)
        }
    }

    var editCanConfirm: Bool {
        guard !isSwitchingCitation, !depictionImageBlocksSave,
              let session = editSession, !session.isSaving
        else { return false }
        switch session.kind {
        case .property(let values, _, _):
            return Self.propertyValuesComplete(values, vocabulary: vocabulary)
        case .bridge(let endA, let endB, let termID, _, _, _):
            let pair = CitationComposerAuthoring.resolvePair(
                endA: endA,
                endB: endB,
                termID: termID,
                vocabulary: vocabulary
            )
            return Self.bridgeReady(pair, termID: termID)
        }
    }

    func citationStepLabel(for id: String?) -> String {
        guard let id, !id.isEmpty,
              let listed = orderedCitations.first(where: { $0.citation.id == id })
        else {
            return L10n.CitationComposer.citationClosed(
                L10n.string(L10n.CitationComposer.newCitation),
                L10n.string(L10n.CitationComposer.notSavedYet)
            )
        }
        let citation = listed.citation
        let label = citation.description.trimmingCharacters(in: .whitespacesAndNewlines)
        if !label.isEmpty {
            return L10n.CitationComposer.citationClosed(citation.ref, label)
        }
        let firstLine = citation.transcription
            .split(whereSeparator: \.isNewline)
            .map { String($0).trimmingCharacters(in: .whitespaces) }
            .first { !$0.isEmpty } ?? ""
        if !firstLine.isEmpty {
            return L10n.CitationComposer.citationClosed(citation.ref, firstLine)
        }
        return L10n.CitationComposer.citationClosed(
            citation.ref,
            L10n.string(L10n.CitationComposer.noTranscriptionLine)
        )
    }

    func showPropertyAuthoring() {
        guard case .bridge(let bridge) = authoring else { return }
        let note = bridge.note
        beginPropertyDraft(subjectID: bridge.endA)
        mutateProperty { draft in
            draft.note = note
            draft.isDuplicate = bridge.origin == .duplicate
        }
    }

    func showBridgeAuthoring() {
        guard case .property(let property) = authoring else { return }
        let note = property.note
        let origin: CitationBridgeOrigin = property.isDuplicate ? .duplicate : .blank
        beginBridgeDraft(
            endA: property.values.subjectID,
            endB: "",
            termID: "",
            origin: origin
        )
        if let note {
            mutateBridge { $0.note = note }
        }
    }

    func setAuthoringCitation(_ id: String) {
        selectCitation(id.isEmpty ? nil : id)
    }

    /// Both citation menus show the citation open in the composer.
    var authoringCitationID: String {
        activeCitationID ?? ""
    }

    private static func propertyValuesComplete(
        _ values: ObservationRowValues,
        vocabulary: CitationComposerVocabulary
    ) -> Bool {
        guard !values.subjectID.isEmpty, !values.propertyID.isEmpty else { return false }
        guard let property = vocabulary.property(id: values.propertyID) else { return false }
        return CitationObservationValue.isDialogValid(
            valueType: property.valueType,
            fields: CitationObservationValue.Fields(
                valueText: values.valueText,
                valueIntegerText: values.valueIntegerText,
                valueTermID: values.valueTermID,
                valueSubjectID: values.valueSubjectID,
                dateDraft: values.dateDraft,
                nameDraft: values.nameDraft,
                depictionDraft: values.depictionDraft
            )
        )
    }

    private static func bridgeReady(_ pair: CitationPair, termID: String) -> Bool {
        guard case .ready(let ready) = pair else { return false }
        if ready.termProperty != nil { return !termID.isEmpty }
        return true
    }

    func openAuthoring(_ work: @escaping @MainActor () -> Void) {
        guard !isDrawingDepiction else { return }
        if authoring.holdsLeave || (editSession?.isDirty ?? false) {
            stepContinuation = work
            pendingLeave = PendingLeave()
            return
        }
        work()
    }

    func beginPropertyDraft(subjectID: String) {
        var values = ObservationRowValues()
        values.subjectID = subjectID
        authoring = .property(
            CitationPropertyDraft(
                values: values,
                baseline: values
            )
        )
        showAuthoringDialog()
    }

    func beginBridgeDraft(
        endA: String,
        endB: String,
        termID: String,
        origin: CitationBridgeOrigin
    ) {
        let note: String?
        switch origin {
        case .graph: note = L10n.string(L10n.CitationComposer.fromGraph)
        case .duplicate: note = L10n.string(L10n.CitationComposer.duplicatedNote)
        case .blank: note = nil
        }
        authoring = .bridge(
            CitationBridgeDraft(
                endA: endA,
                endB: endB,
                termID: termID,
                baselineEndA: endA,
                note: note,
                origin: origin
            )
        )
        showAuthoringDialog()
    }

    func beginAddObservation(for subjectID: String) {
        openAuthoring { self.beginPropertyDraft(subjectID: subjectID) }
    }

    func discardAuthoring() {
        authoring = .rest
        if case .authoring = presented { show(nil) }
    }

    /// Subject, property, and value writes for both observation dialogs.
    /// A new-subject row opens that dialog and leaves the values unchanged.
    func applyObservationSubject(
        _ raw: String,
        field: CitationSubjectField,
        values: inout ObservationRowValues
    ) {
        if raw.hasPrefix(Self.newSubjectPrefix) {
            let key = String(raw.dropFirst(Self.newSubjectPrefix.count))
            beginNewSubject(typeKey: key, field: field)
            return
        }
        values.subjectID = raw
        // A Property the new subject also offers keeps its value, so a run of
        // likenesses keeps Likeness while the person changes.
        if !values.propertyID.isEmpty {
            let offered = vocabulary.propertyOptions(
                forSubjectID: raw,
                capabilities: depictionCapabilities
            )
            if !offered.contains(where: { $0.id == values.propertyID }) {
                values.propertyID = ""
                values.clearValue()
            }
        }
    }

    func applyObservationProperty(_ propertyID: String, values: inout ObservationRowValues) {
        values.propertyID = propertyID
        values.clearValue()
    }

    func selectAuthoringSubject(_ raw: String, field: CitationSubjectField) {
        switch field {
        case .propertySubject:
            mutateProperty { draft in
                self.applyObservationSubject(raw, field: field, values: &draft.values)
            }
        case .bridgeEndA:
            if raw.hasPrefix(Self.newSubjectPrefix) {
                let key = String(raw.dropFirst(Self.newSubjectPrefix.count))
                beginNewSubject(typeKey: key, field: field)
                return
            }
            mutateBridge { $0.endA = raw }
        case .bridgeEndB:
            if raw.hasPrefix(Self.newSubjectPrefix) {
                let key = String(raw.dropFirst(Self.newSubjectPrefix.count))
                beginNewSubject(typeKey: key, field: field)
                return
            }
            mutateBridge { $0.endB = raw }
        }
    }

    func assignCreatedSubject(_ subjectID: String, field: CitationSubjectField) {
        if case .newSubject(_, let over) = presented, over == .edit {
            mutateEdit { session in
                guard case .property(var values, let persistedID, let persistedRef) = session.kind else { return }
                self.applyObservationSubject(subjectID, field: field, values: &values)
                session.kind = .property(
                    values: values,
                    persistedID: persistedID,
                    persistedRef: persistedRef
                )
            }
            return
        }
        selectAuthoringSubject(subjectID, field: field)
    }

    /// Creates the citation when the composer does not have one yet, and only
    /// then treats the current field values as the saved baseline.
    func ensuredCitationID() async throws -> String {
        let creating = context.citationID?.isEmpty != false
        let id = try await context.ensureCitationID(fields.values)
        if creating { fields.captureBaseline() }
        return id
    }

    func saveAuthoring() async {
        await awaitIdentitySwitch()
        switch authoring {
        case .rest: return
        case .property(var draft):
            await savePropertyDraft(&draft)
        case .bridge(var draft):
            await saveBridgeDraft(&draft)
        }
    }

    func duplicate(line: CitationSavedLine) {
        guard !isDrawingDepiction else { return }
        openAuthoring {
            if let observationID = line.observationID,
               let row = self.observationRows.rows.first(where: { $0.persistedID == observationID })
            {
                // The copy owns its own image: it keeps the region, not the saved value.
                var values = row.values
                values.depictionDraft.valueDepictionID = ""
                values.depictionDraft.imageRelPath = ""
                self.authoring = .property(
                    CitationPropertyDraft(
                        values: values,
                        baseline: ObservationRowValues(),
                        note: self.duplicateOrigin(line: line),
                        isDuplicate: true
                    )
                )
                self.showAuthoringDialog()
                return
            }
            if let bridgeID = line.bridgeID,
               let connection = self.connections.rows.first(where: { $0.bridgeID == bridgeID })
            {
                self.beginBridgeDraft(
                    endA: connection.fromSubjectID,
                    endB: connection.toSubjectID,
                    termID: connection.roleTermID,
                    origin: .duplicate
                )
                if case .bridge(var draft) = self.authoring {
                    draft.note = self.duplicateOrigin(line: line)
                    self.authoring = .bridge(draft)
                }
            }
        }
    }

    func beginEdit(line: CitationSavedLine) {
        guard !isDrawingDepiction else { return }
        if let observationID = line.observationID,
           beginEdit(observationID: observationID, subjectID: line.subjectID)
        {
            return
        }
        if let bridgeID = line.bridgeID,
           let connection = connections.rows.first(where: { $0.bridgeID == bridgeID })
        {
            editSession = CitationEditSession(
                kind: .bridge(
                    endA: connection.fromSubjectID,
                    endB: connection.toSubjectID,
                    termID: connection.roleTermID,
                    bridgeID: bridgeID,
                    bridgeRef: connection.bridgeRef ?? bridgeID,
                    baselineTermID: connection.roleTermID
                ),
                openedFromSubjectID: line.subjectID
            )
            showEditDialog()
        }
    }

    /// Opens the edit dialog for one saved observation. A property row edits in
    /// place. A bridge role edits that bridge. Returns false when this citation
    /// has no such observation.
    @discardableResult
    func beginEdit(observationID: String, subjectID: String) -> Bool {
        guard !isDrawingDepiction else { return false }
        if let row = observationRows.rows.first(where: { $0.persistedID == observationID }) {
            editSession = CitationEditSession(
                kind: .property(
                    values: row.values,
                    persistedID: observationID,
                    persistedRef: row.persistedRef ?? observationID
                ),
                propertyBaseline: row.values,
                openedFromSubjectID: subjectID
            )
            showEditDialog()
            return true
        }
        if let connection = connections.rows.first(where: { $0.rolePersistedID == observationID }),
           let bridgeID = connection.bridgeID
        {
            editSession = CitationEditSession(
                kind: .bridge(
                    endA: connection.fromSubjectID,
                    endB: connection.toSubjectID,
                    termID: connection.roleTermID,
                    bridgeID: bridgeID,
                    bridgeRef: connection.bridgeRef ?? bridgeID,
                    baselineTermID: connection.roleTermID
                ),
                openedFromSubjectID: subjectID
            )
            showEditDialog()
            return true
        }
        return false
    }

    func mutateEdit(_ body: (inout CitationEditSession) -> Void) {
        guard var session = editSession else { return }
        body(&session)
        editSession = session
    }

    func discardEdit() {
        closeEditDialog()
    }

    /// Drops the edit draft and the sheet together. Clearing only the draft
    /// leaves the sheet presented with a title and buttons and no fields.
    private func closeEditDialog() {
        editSession = nil
        if case .edit = presented {
            if case .rest = authoring { show(nil) } else { show(.authoring) }
        }
    }

    func saveEdit() async {
        await awaitIdentitySwitch()
        guard var session = editSession, !session.isSaving else { return }
        session.isSaving = true
        session.error = nil
        editSession = session
        switch session.kind {
        case .property(let values, let persistedID, _):
            guard let row = observationRows.rows.first(where: { $0.persistedID == persistedID }) else {
                closeEditDialog()
                return
            }
            observationRows.replaceValues(rowID: row.id, values: values)
            let creating = context.citationID?.isEmpty != false
            do {
                if isDepictionRedrawn {
                    try await withDepictionInput(for: values) { input in
                        await self.observationRows.commit(
                            rowID: row.id,
                            citation: self.fields.values,
                            depictionInput: input
                        )
                    }
                } else {
                    await observationRows.commit(rowID: row.id, citation: fields.values)
                }
            } catch {
                session.isSaving = false
                session.error = L10n.string(L10n.CitationComposer.depictionRenderFailed)
                editSession = session
                return
            }
            if creating, context.citationID?.isEmpty == false {
                fields.captureBaseline()
            }
            if let saved = observationRows.rows.first(where: { $0.id == row.id }),
               case .error(let message) = saved.state
            {
                session.isSaving = false
                session.error = message
                editSession = session
                return
            }
            freshLineKey = "obs:\(persistedID)"
            closeEditDialog()
            await refreshCitationLines()
        case .bridge(_, _, let termID, let bridgeID, _, _):
            guard let connection = connections.rows.first(where: { $0.bridgeID == bridgeID }) else {
                closeEditDialog()
                return
            }
            if let index = connections.rows.firstIndex(where: { $0.id == connection.id }) {
                connections.rows[index].roleTermID = termID
                connections.rows[index].termTouched = termID != (connections.rows[index].roleBaselineTermID ?? "")
            }
            if connections.rows.first(where: { $0.id == connection.id })?.canCommitRole == true {
                await connections.commitRole(connectionID: connection.id)
            }
            if let saved = connections.rows.first(where: { $0.id == connection.id }), saved.error != nil {
                session.isSaving = false
                session.error = saved.error
                editSession = session
                return
            }
            freshLineKey = "bridge:\(bridgeID)"
            closeEditDialog()
            await refreshCitationLines()
        }
    }

    func askDeleteLine(_ line: CitationSavedLine) async {
        guard !isDrawingDepiction else { return }
        if let observationID = line.observationID,
           let row = observationRows.rows.first(where: { $0.persistedID == observationID })
        {
            await askDeleteObservation(rowID: row.id)
            return
        }
        if let bridgeID = line.bridgeID,
           let connection = connections.rows.first(where: { $0.bridgeID == bridgeID })
        {
            let ref = connection.bridgeRef ?? bridgeID
            await presentImpact(kind: "subject", id: bridgeID, ref: ref, title: ref)
        }
    }

    func swapBridgeEnds() {
        mutateBridge { draft in
            let endA = draft.endA
            draft.endA = draft.endB
            draft.endB = endA
        }
    }

    private func savePropertyDraft(_ draft: inout CitationPropertyDraft) async {
        guard let property = vocabulary.property(id: draft.values.propertyID) else {
            draft.error = L10n.string(L10n.CitationComposer.missingPropertyError)
            authoring = .property(draft)
            return
        }
        let row = ObservationRow(
            id: UUID(),
            persistedID: nil,
            persistedRef: nil,
            values: draft.values,
            baseline: nil,
            state: .draft,
            propertyError: nil
        )
        let (draftOrNil, draftError) = row.asDraft(property: property)
        if let draftError {
            draft.error = draftError
            authoring = .property(draft)
            return
        }
        guard let observation = draftOrNil else { return }
        draft.isSaving = true
        draft.error = nil
        authoring = .property(draft)
        let values = draft.values
        let isDepiction = vocabulary.isDepiction(propertyID: values.propertyID)
        do {
            let written: ([CatalogObservation], CatalogWriteEffects)
            if isDepiction {
                written = try await withDepictionInput(for: values) { input in
                    var withImage = observation
                    withImage.depiction = input
                    return try await self.addObservation(withImage)
                }
            } else {
                written = try await addObservation(observation)
            }
            context.applyWrite(written.1)
            let savedID = written.0.first?.id
            await refreshCitationLines()
            // A run of faces is one per person: Likeness stays, the person goes.
            var kept = ObservationRowValues()
            if isDepiction {
                kept.propertyID = values.propertyID
            } else {
                kept.subjectID = values.subjectID
            }
            authoring = .property(CitationPropertyDraft(values: kept, baseline: kept))
            showAuthoringDialog()
            if let savedID {
                freshLineKey = "obs:\(savedID)"
            }
        } catch is CitationComposerNeedArtifact {
            draft.isSaving = false
            draft.error = L10n.string(L10n.CitationComposer.needArtifact)
            authoring = .property(draft)
        } catch is DepictionRenderFailed {
            draft.isSaving = false
            draft.error = L10n.string(L10n.CitationComposer.depictionRenderFailed)
            authoring = .property(draft)
        } catch {
            draft.isSaving = false
            draft.error = L10n.Errors.message(for: error)
            authoring = .property(draft)
        }
    }

    private func addObservation(
        _ observation: CatalogObservationDraft
    ) async throws -> ([CatalogObservation], CatalogWriteEffects) {
        let citationID = try await ensuredCitationID()
        return try await context.store.addObservationsToCitation(
            projectDir: context.projectDir,
            userID: context.userID,
            citationID: citationID,
            observations: [observation]
        )
    }

    private func saveBridgeDraft(_ draft: inout CitationBridgeDraft) async {
        let pair = CitationComposerAuthoring.resolvePair(
            endA: draft.endA,
            endB: draft.endB,
            termID: draft.termID,
            vocabulary: vocabulary
        )
        guard case .ready(let ready) = pair else {
            draft.error = L10n.string(L10n.CitationComposer.refusedPairBody)
            authoring = .bridge(draft)
            return
        }
        if ready.termProperty != nil, draft.termID.isEmpty {
            draft.error = L10n.string(L10n.CitationComposer.connectionTermNotChosen)
            authoring = .bridge(draft)
            return
        }
        let pending = ConnectionRow(
            id: UUID(),
            isPending: true,
            fromSubjectID: draft.endA,
            toSubjectID: draft.endB,
            fromLabel: vocabulary.subjectsByID[draft.endA]?.label ?? "",
            toLabel: vocabulary.subjectsByID[draft.endB]?.label ?? "",
            bridgeTypeKey: ready.rule.bridgeTypeKey,
            sentence: "",
            termProperty: ready.termProperty,
            roleTermID: draft.termID,
            termTouched: !draft.termID.isEmpty,
            isSaving: false
        )
        guard let observations = connections.edgeDrafts(for: pending, vocabulary: vocabulary) else {
            draft.error = L10n.string(L10n.CitationComposer.refusedPairBody)
            authoring = .bridge(draft)
            return
        }
        draft.isSaving = true
        draft.error = nil
        authoring = .bridge(draft)
        let endA = draft.endA
        let endALabel = vocabulary.subjectsByID[endA]?.label ?? ""
        let endBLabel = vocabulary.subjectsByID[draft.endB]?.label ?? ""
        do {
            let citationID = try await ensuredCitationID()
            let created = try await context.store.createCitedBridge(
                projectDir: context.projectDir,
                userID: context.userID,
                fromSubjectID: draft.endA,
                toSubjectID: draft.endB,
                bridgeTypeKey: ready.rule.bridgeTypeKey,
                description: "",
                artifactID: "",
                locatorJSON: "",
                transcription: "",
                citationDescription: "",
                transcriptionUncertain: false,
                transcriptionNote: "",
                citationNotes: [],
                observations: observations,
                citationID: citationID
            )
            context.applyWrite(created.3)
            await refreshCitationLines()
            authoring = .bridge(
                CitationBridgeDraft(
                    endA: endA,
                    baselineEndA: endA
                )
            )
            showAuthoringDialog()
            freshLineKey = "bridge:\(created.0.id)"
            let sentence = ready.reading ?? L10n.CitationComposer.bridgeEndsReading(
                endA: endALabel,
                endB: endBLabel
            )
            let ref = created.0.ref
            let message = draft.origin == .duplicate
                ? L10n.CitationComposer.duplicateSavedBody(reading: sentence, ref: ref)
                : L10n.CitationComposer.bridgeSavedBody(reading: sentence, ref: ref)
            bridgeToast = (
                title: L10n.string(L10n.CitationComposer.bridgeSavedTitle),
                message: message
            )
        } catch is CitationComposerNeedArtifact {
            draft.isSaving = false
            draft.error = L10n.string(L10n.CitationComposer.needArtifact)
            authoring = .bridge(draft)
        } catch {
            draft.isSaving = false
            draft.error = L10n.Errors.message(for: error)
            authoring = .bridge(draft)
        }
    }

    func mutateProperty(_ body: (inout CitationPropertyDraft) -> Void) {
        guard case .property(var draft) = authoring else { return }
        body(&draft)
        authoring = .property(draft)
    }

    func duplicateOrigin(line: CitationSavedLine) -> String {
        let name = vocabulary.subjectsByID[line.subjectID]?.label
            .trimmingCharacters(in: .whitespacesAndNewlines) ?? ""
        let isDepiction = observationRows.rows
            .first { $0.persistedID == line.observationID && line.observationID != nil }
            .map { vocabulary.isDepiction(propertyID: $0.propertyID) } ?? false
        if isDepiction, !name.isEmpty {
            return L10n.CitationComposer.depictionDuplicatedFrom(name, line.label)
        }
        let phrase = line.phrase.trimmingCharacters(in: .whitespacesAndNewlines)
        guard !name.isEmpty, !phrase.isEmpty else {
            return L10n.string(L10n.CitationComposer.duplicatedNote)
        }
        return L10n.CitationComposer.duplicatedFrom(name, phrase)
    }

    func mutateBridge(_ body: (inout CitationBridgeDraft) -> Void) {
        guard case .bridge(var draft) = authoring else { return }
        body(&draft)
        authoring = .bridge(draft)
    }
}
