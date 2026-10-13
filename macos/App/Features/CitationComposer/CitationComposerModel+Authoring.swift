import Foundation

extension CitationComposerModel {
    var orderedCitations: [CatalogListedCitation] {
        CitationComposerAuthoring.stepperOrder(listedCitations)
    }

    var subjectGroups: [CitationSubjectGroup] {
        CitationComposerAuthoring.groups(
            observations: observationRows.rows,
            connections: connections.rows,
            vocabulary: vocabulary
        )
    }

    var savedLineCount: Int {
        let properties = observationRows.rows.filter { $0.persistedID != nil }.count
        let bridges = connections.rows.filter { !$0.isPending && !($0.bridgeID ?? "").isEmpty }.count
        return properties + bridges
    }

    var observationsMeta: String {
        let ref = activeCitationRef
        if savedLineCount == 0 {
            return ref.isEmpty
                ? L10n.string(L10n.CitationComposer.noneYet)
                : L10n.CitationComposer.noneOnCitation(ref: ref)
        }
        return L10n.CitationComposer.observationsOnCitation(count: savedLineCount, ref: ref)
    }

    func openAuthoring(_ work: @escaping @MainActor () -> Void) {
        if authoring.holdsLeave || (editSession?.isDirty ?? false) {
            stepContinuation = work
            pendingLeave = PendingLeave(
                citationDirty: fields.isDirty,
                observationCount: 1,
                connectionTouched: false
            )
            return
        }
        work()
    }

    func beginPropertyDraft(subjectID: String) {
        authoringFocusNonce += 1
        var values = ObservationRowValues()
        values.subjectID = subjectID
        authoring = .property(
            CitationPropertyDraft(
                values: values,
                baseline: values,
                focus: subjectID.isEmpty ? .subject : .property
            )
        )
    }

    func beginBridgeDraft(
        endA: String,
        endB: String,
        termID: String,
        origin: CitationBridgeOrigin
    ) {
        authoringFocusNonce += 1
        let pair = CitationComposerAuthoring.resolvePair(
            endA: endA,
            endB: endB,
            termID: termID,
            vocabulary: vocabulary
        )
        let focus: CitationAuthoringFocus
        if case .ready(let ready) = pair, ready.termProperty != nil, termID.isEmpty {
            focus = .term
        } else if endA.isEmpty {
            focus = .endA
        } else if endB.isEmpty {
            focus = .endB
        } else {
            focus = .term
        }
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
                focus: focus,
                note: note,
                origin: origin
            )
        )
    }

    func beginAddBridge() {
        openAuthoring { self.beginBridgeDraft(endA: "", endB: "", termID: "", origin: .blank) }
    }

    func beginAddObservation(for subjectID: String) {
        openAuthoring { self.beginPropertyDraft(subjectID: subjectID) }
    }

    func discardAuthoring() {
        authoring = .rest
    }

    func selectAuthoringSubject(_ raw: String, field: CitationSubjectField) {
        if raw.hasPrefix(Self.newSubjectPrefix) {
            let key = String(raw.dropFirst(Self.newSubjectPrefix.count))
            beginNewSubject(typeKey: key, field: field)
            return
        }
        switch field {
        case .propertySubject:
            mutateProperty { $0.values.subjectID = raw }
        case .bridgeEndA:
            mutateBridge { $0.endA = raw }
        case .bridgeEndB:
            mutateBridge { $0.endB = raw }
        }
    }

    func assignCreatedSubject(_ subjectID: String, field: CitationSubjectField) {
        selectAuthoringSubject(subjectID, field: field)
    }

    func saveAuthoring() async {
        switch authoring {
        case .rest: return
        case .property(var draft):
            await savePropertyDraft(&draft)
        case .bridge(var draft):
            await saveBridgeDraft(&draft)
        }
    }

    func duplicate(line: CitationSavedLine) {
        openAuthoring {
            if let observationID = line.observationID,
               let row = self.observationRows.rows.first(where: { $0.persistedID == observationID })
            {
                self.authoringFocusNonce += 1
                self.authoring = .property(
                    CitationPropertyDraft(
                        values: row.values,
                        baseline: ObservationRowValues(),
                        focus: .property,
                        note: L10n.string(L10n.CitationComposer.duplicatedNote),
                        isDuplicate: true
                    )
                )
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
            }
        }
    }

    func beginEdit(line: CitationSavedLine) {
        if let observationID = line.observationID,
           let row = observationRows.rows.first(where: { $0.persistedID == observationID })
        {
            editSession = CitationEditSession(
                kind: .property(
                    values: row.values,
                    persistedID: observationID,
                    persistedRef: row.persistedRef ?? observationID
                ),
                propertyBaseline: row.values
            )
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
                )
            )
        }
    }

    func mutateEdit(_ body: (inout CitationEditSession) -> Void) {
        guard var session = editSession else { return }
        body(&session)
        editSession = session
    }

    func discardEdit() {
        editSession = nil
        editSheetSuspended = false
    }

    func saveEdit() async {
        guard var session = editSession, !session.isSaving else { return }
        session.isSaving = true
        session.error = nil
        editSession = session
        switch session.kind {
        case .property(let values, let persistedID, _):
            guard let row = observationRows.rows.first(where: { $0.persistedID == persistedID }) else {
                editSession = nil
                return
            }
            observationRows.replaceValues(rowID: row.id, values: values)
            await observationRows.commit(rowID: row.id)
            if let saved = observationRows.rows.first(where: { $0.id == row.id }),
               case .error(let message) = saved.state
            {
                session.isSaving = false
                session.error = message
                editSession = session
                return
            }
            freshLineKey = "obs:\(persistedID)"
            editSession = nil
            await refreshOpenCitation()
        case .bridge(_, _, let termID, let bridgeID, _, _):
            guard let connection = connections.rows.first(where: { $0.bridgeID == bridgeID }) else {
                editSession = nil
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
            editSession = nil
            await refreshOpenCitation()
        }
    }

    func negate(observationID: String) async {
        guard let row = observationRows.rows.first(where: { $0.persistedID == observationID }) else { return }
        observationRows.togglePolarity(rowID: row.id)
        await observationRows.commit(rowID: row.id)
        freshLineKey = "obs:\(observationID)"
    }

    func askDeleteLine(_ line: CitationSavedLine) async {
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

    func showOnGraph(bridgeID: String) {
        requestStep {
            self.session.setPendingGraphSelection(subjectID: bridgeID, sourceID: self.sourceID)
            self.navigation?.go(to: self.graphLocation())
        }
    }

    func beginValueEditor(applyToEdit: Bool) {
        let values: ObservationRowValues
        let propertyID: String
        if applyToEdit, let session = editSession, case .property(let current, _, _) = session.kind {
            values = current
            propertyID = current.propertyID
            editSheetSuspended = true
        } else if case .property(let draft) = authoring {
            values = draft.values
            propertyID = draft.values.propertyID
        } else {
            return
        }
        var state = ObservationDialogState(
            editingID: nil,
            propertyID: propertyID,
            polarity: values.polarity,
            valueText: values.valueText,
            valueIntegerText: values.valueIntegerText,
            valueTermID: values.valueTermID,
            dateDraft: values.dateDraft,
            nameDraft: values.nameDraft,
            showValidation: false
        )
        if catalogProperty(id: propertyID)?.valueType == PropertyValueType.date.rawValue,
           state.dateDraft.isBlank,
           let match = DateValueTranscriptionMatch.first(in: fields.transcription, locale: .current)
        {
            state.dateDraft.startYear = match.year
            state.dateDraft.startMonth = match.month
            state.dateDraft.startDay = match.day
        }
        observationDialogBaseline = state
        observationDialog = state
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
        let subjectID = draft.values.subjectID
        do {
            let citationID = try await context.ensureCitationID()
            let written = try await context.store.addObservationsToCitation(
                projectDir: context.projectDir,
                userID: context.userID,
                citationID: citationID,
                observations: [observation]
            )
            context.applyWrite(written.1)
            let savedID = written.0.first?.id
            await refreshOpenCitation()
            authoringFocusNonce += 1
            var kept = ObservationRowValues()
            kept.subjectID = subjectID
            authoring = .property(
                CitationPropertyDraft(values: kept, baseline: kept, focus: .property)
            )
            if let savedID {
                freshLineKey = "obs:\(savedID)"
            }
        } catch is CitationComposerNeedArtifact {
            draft.isSaving = false
            draft.error = L10n.string(L10n.CitationComposer.needArtifact)
            authoring = .property(draft)
        } catch {
            draft.isSaving = false
            draft.error = L10n.Errors.message(for: error)
            authoring = .property(draft)
        }
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
        if let termProperty = ready.termProperty, draft.termID.isEmpty {
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
            let citationID = try await context.ensureCitationID()
            let created = try await context.store.createCitedBridge(
                projectDir: context.projectDir,
                userID: context.userID,
                sourceID: context.sourceID,
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
            await refreshOpenCitation()
            authoringFocusNonce += 1
            authoring = .bridge(
                CitationBridgeDraft(
                    endA: endA,
                    baselineEndA: endA,
                    focus: .endB
                )
            )
            freshLineKey = "bridge:\(created.0.id)"
            bridgeToast = (
                title: L10n.string(L10n.CitationComposer.bridgeSavedTitle),
                message: L10n.CitationComposer.bridgeSavedBody(endA: endALabel, endB: endBLabel)
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

    func mutateBridge(_ body: (inout CitationBridgeDraft) -> Void) {
        guard case .bridge(var draft) = authoring else { return }
        body(&draft)
        authoring = .bridge(draft)
    }
}
