import Foundation

struct GoStore: GenealogyStore {
    func installIdentity(identityDir: String) async throws -> InstallIdentity? {
        var req = Provenencia_Engine_V1_GetInstallIdentityRequest()
        req.identityDir = identityDir
        let resp: Provenencia_Engine_V1_GetInstallIdentityResponse = try await provenenciaCall(
            method: CoreMethod.getInstallIdentity,
            request: req
        )
        guard resp.found else {
            return nil
        }
        return InstallIdentity(userID: resp.userID, displayName: resp.displayName, ref: resp.ref)
    }

    func completeOnboarding(
        identityDir: String,
        parentDir: String,
        displayName: String,
        familyName: String
    ) async throws -> OnboardingResult {
        var req = Provenencia_Engine_V1_CompleteOnboardingRequest()
        req.identityDir = identityDir
        req.parentDir = parentDir
        req.displayName = displayName
        req.familyName = familyName
        let resp: Provenencia_Engine_V1_CompleteOnboardingResponse = try await provenenciaCall(
            method: CoreMethod.completeOnboarding,
            request: req
        )
        return OnboardingResult(
            projectDir: resp.projectDir,
            userID: resp.userID,
            displayName: resp.displayName,
            ref: resp.ref,
            project: Self.mapProject(resp.project)
        )
    }

    func activeProject(identityDir: String) async throws -> String? {
        var req = Provenencia_Engine_V1_GetActiveProjectRequest()
        req.identityDir = identityDir
        let resp: Provenencia_Engine_V1_GetActiveProjectResponse = try await provenenciaCall(
            method: CoreMethod.getActiveProject,
            request: req
        )
        guard resp.found else {
            return nil
        }
        return resp.projectDir
    }

    func openProject(
        identityDir: String,
        projectDir: String,
        displayName: String,
        adoptUserID: String
    ) async throws -> OnboardingResult {
        var req = Provenencia_Engine_V1_OpenProjectRequest()
        req.identityDir = identityDir
        req.projectDir = projectDir
        req.displayName = displayName
        req.adoptUserID = adoptUserID
        let resp: Provenencia_Engine_V1_OpenProjectResponse = try await provenenciaCall(
            method: CoreMethod.openProject,
            request: req
        )
        return OnboardingResult(
            projectDir: resp.projectDir,
            userID: resp.userID,
            displayName: resp.displayName,
            ref: resp.ref,
            project: Self.mapProject(resp.project)
        )
    }

    func listProjectUsers(projectDir: String) async throws -> [InstallIdentity] {
        var req = Provenencia_Engine_V1_ListProjectUsersRequest()
        req.projectDir = projectDir
        let resp: Provenencia_Engine_V1_ListProjectUsersResponse = try await provenenciaCall(
            method: CoreMethod.listProjectUsers,
            request: req
        )
        return resp.users.map {
            InstallIdentity(userID: $0.userID, displayName: $0.displayName, ref: $0.ref)
        }
    }

    func removeActiveProject(identityDir: String) async throws {
        var req = Provenencia_Engine_V1_RemoveActiveProjectRequest()
        req.identityDir = identityDir
        let _: Provenencia_Engine_V1_RemoveActiveProjectResponse = try await provenenciaCall(
            method: CoreMethod.removeActiveProject,
            request: req
        )
    }

    func signOut(identityDir: String) async throws {
        var req = Provenencia_Engine_V1_SignOutRequest()
        req.identityDir = identityDir
        let _: Provenencia_Engine_V1_SignOutResponse = try await provenenciaCall(
            method: CoreMethod.signOut,
            request: req
        )
    }

    func closeCatalogSession(projectDir: String) async throws {
        var req = Provenencia_Engine_V1_CloseCatalogSessionRequest()
        req.projectDir = projectDir
        let _: Provenencia_Engine_V1_CloseCatalogSessionResponse = try await provenenciaCall(
            method: CoreMethod.closeCatalogSession,
            request: req
        )
    }

    func projectInfo(projectDir: String) async throws -> ProjectInfo {
        var req = Provenencia_Engine_V1_GetProjectInfoRequest()
        req.projectDir = projectDir
        let resp: Provenencia_Engine_V1_GetProjectInfoResponse = try await provenenciaCall(
            method: CoreMethod.getProjectInfo,
            request: req
        )
        return Self.mapProject(resp.project)
    }

    func listSources(projectDir: String) async throws -> [CatalogSource] {
        var req = Provenencia_Engine_V1_ListSourcesRequest()
        req.projectDir = projectDir
        let resp: Provenencia_Engine_V1_ListSourcesResponse = try await provenenciaCall(
            method: CoreMethod.listSources,
            request: req
        )
        return resp.sources.map(Self.mapSource)
    }

    func getSourceWorkspace(projectDir: String, sourceID: String) async throws -> CatalogSourceWorkspace {
        var req = Provenencia_Engine_V1_GetSourceWorkspaceRequest()
        req.projectDir = projectDir
        req.sourceID = sourceID
        let resp: Provenencia_Engine_V1_GetSourceWorkspaceResponse = try await provenenciaCall(
            method: CoreMethod.getSourceWorkspace,
            request: req
        )
        return CatalogSourceWorkspace(
            source: Self.mapSource(resp.source),
            notes: resp.notes.map(Self.mapNote),
            metadata: resp.metadata.map(Self.mapMetadataEntry),
            artifacts: resp.artifacts.map(Self.mapArtifact),
            credibility: resp.hasCredibility ? Self.mapCredibility(resp.credibility) : nil
        )
    }

    func createSource(
        projectDir: String,
        userID: String,
        sourceTypeID: String,
        title: String,
        description: String
    ) async throws -> CatalogSource {
        var req = Provenencia_Engine_V1_CreateSourceRequest()
        req.projectDir = projectDir
        req.userID = userID
        req.sourceTypeID = sourceTypeID
        req.title = title
        req.description_p = description
        let resp: Provenencia_Engine_V1_CreateSourceResponse = try await provenenciaCall(
            method: CoreMethod.createSource,
            request: req
        )
        return Self.mapSource(resp.source)
    }

    func updateSource(
        projectDir: String,
        userID: String,
        sourceID: String,
        sourceTypeID: String,
        title: String,
        description: String
    ) async throws -> CatalogSource {
        var req = Provenencia_Engine_V1_UpdateSourceRequest()
        req.projectDir = projectDir
        req.userID = userID
        req.sourceID = sourceID
        req.sourceTypeID = sourceTypeID
        req.title = title
        req.description_p = description
        let resp: Provenencia_Engine_V1_UpdateSourceResponse = try await provenenciaCall(
            method: CoreMethod.updateSource,
            request: req
        )
        return Self.mapSource(resp.source)
    }

    func setSourceCover(
        projectDir: String,
        userID: String,
        sourceID: String,
        coverMode: String,
        primaryArtifactID: String
    ) async throws -> CatalogSource {
        var req = Provenencia_Engine_V1_SetSourceCoverRequest()
        req.projectDir = projectDir
        req.userID = userID
        req.sourceID = sourceID
        req.coverMode = coverMode
        req.primaryArtifactID = primaryArtifactID
        let resp: Provenencia_Engine_V1_SetSourceCoverResponse = try await provenenciaCall(
            method: CoreMethod.setSourceCover,
            request: req
        )
        return Self.mapSource(resp.source)
    }

    func addSourceNote(projectDir: String, userID: String, sourceID: String, body: String) async throws
        -> CatalogSourceNote
    {
        var req = Provenencia_Engine_V1_AddSourceNoteRequest()
        req.projectDir = projectDir
        req.userID = userID
        req.sourceID = sourceID
        req.body = body
        let resp: Provenencia_Engine_V1_AddSourceNoteResponse = try await provenenciaCall(
            method: CoreMethod.addSourceNote,
            request: req
        )
        return Self.mapNote(resp.note)
    }

    func updateSourceNote(projectDir: String, userID: String, noteID: String, body: String) async throws
        -> CatalogSourceNote
    {
        var req = Provenencia_Engine_V1_UpdateSourceNoteRequest()
        req.projectDir = projectDir
        req.userID = userID
        req.noteID = noteID
        req.body = body
        let resp: Provenencia_Engine_V1_UpdateSourceNoteResponse = try await provenenciaCall(
            method: CoreMethod.updateSourceNote,
            request: req
        )
        return Self.mapNote(resp.note)
    }

    func deleteSourceNote(projectDir: String, userID: String, noteID: String) async throws {
        var req = Provenencia_Engine_V1_DeleteSourceNoteRequest()
        req.projectDir = projectDir
        req.userID = userID
        req.noteID = noteID
        let _: Provenencia_Engine_V1_DeleteSourceNoteResponse = try await provenenciaCall(
            method: CoreMethod.deleteSourceNote,
            request: req
        )
    }

    func setSourceMetadata(
        projectDir: String,
        userID: String,
        sourceID: String,
        fieldID: String,
        valueText: String
    ) async throws -> CatalogMetadataEntry {
        var req = Provenencia_Engine_V1_SetSourceMetadataRequest()
        req.projectDir = projectDir
        req.userID = userID
        req.sourceID = sourceID
        req.fieldID = fieldID
        req.valueText = valueText
        let resp: Provenencia_Engine_V1_SetSourceMetadataResponse = try await provenenciaCall(
            method: CoreMethod.setSourceMetadata,
            request: req
        )
        return Self.mapMetadataEntry(resp.entry)
    }

    func clearSourceMetadata(projectDir: String, userID: String, sourceID: String, fieldID: String) async throws {
        var req = Provenencia_Engine_V1_ClearSourceMetadataRequest()
        req.projectDir = projectDir
        req.userID = userID
        req.sourceID = sourceID
        req.fieldID = fieldID
        let _: Provenencia_Engine_V1_ClearSourceMetadataResponse = try await provenenciaCall(
            method: CoreMethod.clearSourceMetadata,
            request: req
        )
    }

    func dismissSourceMetadataSuggestion(
        projectDir: String,
        userID: String,
        sourceID: String,
        fieldID: String
    ) async throws -> [CatalogMetadataEntry] {
        var req = Provenencia_Engine_V1_DismissSourceMetadataSuggestionRequest()
        req.projectDir = projectDir
        req.userID = userID
        req.sourceID = sourceID
        req.fieldID = fieldID
        let resp: Provenencia_Engine_V1_DismissSourceMetadataSuggestionResponse = try await provenenciaCall(
            method: CoreMethod.dismissSourceMetadataSuggestion,
            request: req
        )
        return resp.metadata.map(Self.mapMetadataEntry)
    }

    func reorderSourceMetadata(
        projectDir: String,
        userID: String,
        sourceID: String,
        fieldIDs: [String]
    ) async throws -> [CatalogMetadataEntry] {
        var req = Provenencia_Engine_V1_ReorderSourceMetadataRequest()
        req.projectDir = projectDir
        req.userID = userID
        req.sourceID = sourceID
        req.fieldIds = fieldIDs
        let resp: Provenencia_Engine_V1_ReorderSourceMetadataResponse = try await provenenciaCall(
            method: CoreMethod.reorderSourceMetadata,
            request: req
        )
        return resp.metadata.map(Self.mapMetadataEntry)
    }

    func createArtifact(
        projectDir: String,
        userID: String,
        sourceID: String,
        fileID: String,
        label: String,
        description: String
    ) async throws -> CatalogArtifact {
        var req = Provenencia_Engine_V1_CreateArtifactRequest()
        req.projectDir = projectDir
        req.userID = userID
        req.sourceID = sourceID
        req.fileID = fileID
        req.label = label
        req.description_p = description
        let resp: Provenencia_Engine_V1_CreateArtifactResponse = try await provenenciaCall(
            method: CoreMethod.createArtifact,
            request: req
        )
        return Self.mapArtifact(resp.artifact)
    }

    func updateArtifact(
        projectDir: String,
        userID: String,
        artifactID: String,
        label: String,
        description: String
    ) async throws -> CatalogArtifact {
        var req = Provenencia_Engine_V1_UpdateArtifactRequest()
        req.projectDir = projectDir
        req.userID = userID
        req.artifactID = artifactID
        req.label = label
        req.description_p = description
        let resp: Provenencia_Engine_V1_UpdateArtifactResponse = try await provenenciaCall(
            method: CoreMethod.updateArtifact,
            request: req
        )
        return Self.mapArtifact(resp.artifact)
    }

    func ingestArtifactFile(
        projectDir: String,
        userID: String,
        artifactID: String,
        path: String
    ) async throws -> (artifact: CatalogArtifact, file: CatalogFileRef, reused: Bool) {
        var req = Provenencia_Engine_V1_IngestArtifactFileRequest()
        req.projectDir = projectDir
        req.userID = userID
        req.artifactID = artifactID
        req.path = path
        let resp: Provenencia_Engine_V1_IngestArtifactFileResponse = try await provenenciaCall(
            method: CoreMethod.ingestArtifactFile,
            request: req
        )
        return (Self.mapArtifact(resp.artifact), Self.mapFile(resp.file), resp.reused)
    }

    func ensureFileThumbnail(
        projectDir: String,
        fileID: String
    ) async throws -> (relPath: String, skipped: Bool) {
        var req = Provenencia_Engine_V1_EnsureFileThumbnailRequest()
        req.projectDir = projectDir
        req.fileID = fileID
        let resp: Provenencia_Engine_V1_EnsureFileThumbnailResponse = try await provenenciaCall(
            method: CoreMethod.ensureFileThumbnail,
            request: req
        )
        return (resp.relPath, resp.skipped)
    }

    func listSourceCredibilityGrades(projectDir: String) async throws -> [CatalogCredibilityGrade] {
        var req = Provenencia_Engine_V1_ListSourceCredibilityGradesRequest()
        req.projectDir = projectDir
        let resp: Provenencia_Engine_V1_ListSourceCredibilityGradesResponse = try await provenenciaCall(
            method: CoreMethod.listSourceCredibilityGrades,
            request: req
        )
        return resp.grades.map(Self.mapCredibilityGrade)
    }

    func upsertSourceCredibilityAssessment(
        projectDir: String,
        userID: String,
        sourceID: String,
        gradeID: String,
        argument: String
    ) async throws -> CatalogCredibilityAssessment {
        var req = Provenencia_Engine_V1_UpsertSourceCredibilityAssessmentRequest()
        req.projectDir = projectDir
        req.userID = userID
        req.sourceID = sourceID
        req.gradeID = gradeID
        req.argument = argument
        let resp: Provenencia_Engine_V1_UpsertSourceCredibilityAssessmentResponse = try await provenenciaCall(
            method: CoreMethod.upsertSourceCredibilityAssessment,
            request: req
        )
        return Self.mapCredibility(resp.assessment)
    }

    func listSourceTypes(projectDir: String) async throws -> [CatalogSourceType] {
        var req = Provenencia_Engine_V1_ListSourceTypesRequest()
        req.projectDir = projectDir
        let resp: Provenencia_Engine_V1_ListSourceTypesResponse = try await provenenciaCall(
            method: CoreMethod.listSourceTypes,
            request: req
        )
        return resp.types.map(Self.mapSourceType)
    }

    func createSourceType(
        projectDir: String,
        userID: String,
        label: String,
        description: String,
        iconKey: String
    ) async throws -> CatalogSourceType {
        var req = Provenencia_Engine_V1_CreateSourceTypeRequest()
        req.projectDir = projectDir
        req.userID = userID
        req.label = label
        req.description_p = description
        req.iconKey = iconKey
        let resp: Provenencia_Engine_V1_CreateSourceTypeResponse = try await provenenciaCall(
            method: CoreMethod.createSourceType,
            request: req
        )
        return Self.mapSourceType(resp.type)
    }

    func updateSourceType(
        projectDir: String,
        userID: String,
        typeID: String,
        label: String,
        description: String,
        iconKey: String
    ) async throws -> CatalogSourceType {
        var req = Provenencia_Engine_V1_UpdateSourceTypeRequest()
        req.projectDir = projectDir
        req.userID = userID
        req.typeID = typeID
        req.label = label
        req.description_p = description
        req.iconKey = iconKey
        let resp: Provenencia_Engine_V1_UpdateSourceTypeResponse = try await provenenciaCall(
            method: CoreMethod.updateSourceType,
            request: req
        )
        return Self.mapSourceType(resp.type)
    }

    func deleteSourceType(
        projectDir: String,
        userID: String,
        typeID: String
    ) async throws {
        var req = Provenencia_Engine_V1_DeleteSourceTypeRequest()
        req.projectDir = projectDir
        req.userID = userID
        req.typeID = typeID
        let _: Provenencia_Engine_V1_DeleteSourceTypeResponse = try await provenenciaCall(
            method: CoreMethod.deleteSourceType,
            request: req
        )
    }

    func listTypeSuggestions(projectDir: String, typeID: String) async throws -> [CatalogTypeSuggestion] {
        var req = Provenencia_Engine_V1_ListTypeSuggestionsRequest()
        req.projectDir = projectDir
        req.typeID = typeID
        let resp: Provenencia_Engine_V1_ListTypeSuggestionsResponse = try await provenenciaCall(
            method: CoreMethod.listTypeSuggestions,
            request: req
        )
        return resp.suggestions.map(Self.mapTypeSuggestion)
    }

    func assignTypeField(
        projectDir: String,
        userID: String,
        typeID: String,
        fieldID: String
    ) async throws -> [CatalogTypeSuggestion] {
        var req = Provenencia_Engine_V1_AssignTypeFieldRequest()
        req.projectDir = projectDir
        req.userID = userID
        req.typeID = typeID
        req.fieldID = fieldID
        let resp: Provenencia_Engine_V1_AssignTypeFieldResponse = try await provenenciaCall(
            method: CoreMethod.assignTypeField,
            request: req
        )
        return resp.suggestions.map(Self.mapTypeSuggestion)
    }

    func removeTypeField(
        projectDir: String,
        userID: String,
        typeID: String,
        fieldID: String
    ) async throws -> [CatalogTypeSuggestion] {
        var req = Provenencia_Engine_V1_RemoveTypeFieldRequest()
        req.projectDir = projectDir
        req.userID = userID
        req.typeID = typeID
        req.fieldID = fieldID
        let resp: Provenencia_Engine_V1_RemoveTypeFieldResponse = try await provenenciaCall(
            method: CoreMethod.removeTypeField,
            request: req
        )
        return resp.suggestions.map(Self.mapTypeSuggestion)
    }

    func listMetadataFields(projectDir: String) async throws -> [CatalogMetadataField] {
        var req = Provenencia_Engine_V1_ListMetadataFieldsRequest()
        req.projectDir = projectDir
        let resp: Provenencia_Engine_V1_ListMetadataFieldsResponse = try await provenenciaCall(
            method: CoreMethod.listMetadataFields,
            request: req
        )
        return resp.fields.map(Self.mapMetadataField)
    }

    func createMetadataField(
        projectDir: String,
        userID: String,
        label: String,
        dataType: String,
        description: String
    ) async throws -> CatalogMetadataField {
        var req = Provenencia_Engine_V1_CreateMetadataFieldRequest()
        req.projectDir = projectDir
        req.userID = userID
        req.label = label
        req.dataType = dataType
        req.description_p = description
        let resp: Provenencia_Engine_V1_CreateMetadataFieldResponse = try await provenenciaCall(
            method: CoreMethod.createMetadataField,
            request: req
        )
        return Self.mapMetadataField(resp.field)
    }

    func updateMetadataField(
        projectDir: String,
        userID: String,
        fieldID: String,
        label: String,
        dataType: String,
        description: String
    ) async throws -> CatalogMetadataField {
        var req = Provenencia_Engine_V1_UpdateMetadataFieldRequest()
        req.projectDir = projectDir
        req.userID = userID
        req.fieldID = fieldID
        req.label = label
        req.dataType = dataType
        req.description_p = description
        let resp: Provenencia_Engine_V1_UpdateMetadataFieldResponse = try await provenenciaCall(
            method: CoreMethod.updateMetadataField,
            request: req
        )
        return Self.mapMetadataField(resp.field)
    }

    func deleteMetadataField(
        projectDir: String,
        userID: String,
        fieldID: String
    ) async throws {
        var req = Provenencia_Engine_V1_DeleteMetadataFieldRequest()
        req.projectDir = projectDir
        req.userID = userID
        req.fieldID = fieldID
        let _: Provenencia_Engine_V1_DeleteMetadataFieldResponse = try await provenenciaCall(
            method: CoreMethod.deleteMetadataField,
            request: req
        )
    }

    func workspaceNavCounts(projectDir: String) async throws -> WorkspaceNavCounts {
        var req = Provenencia_Engine_V1_GetWorkspaceNavCountsRequest()
        req.projectDir = projectDir
        let resp: Provenencia_Engine_V1_GetWorkspaceNavCountsResponse = try await provenenciaCall(
            method: CoreMethod.getWorkspaceNavCounts,
            request: req
        )
        return WorkspaceNavCounts(
            sources: Int(resp.sources),
            sourceTypes: Self.mapOriginCounts(resp.sourceTypes),
            sourceFields: Self.mapOriginCounts(resp.sourceFields)
        )
    }

    func searchCatalog(
        projectDir: String,
        query: String,
        location: WorkspaceLocation
    ) async throws -> [CatalogSearchHit] {
        var req = Provenencia_Engine_V1_SearchCatalogRequest()
        req.projectDir = projectDir
        req.query = query
        req.location = Self.mapWorkspaceLocationToProto(location)
        let resp: Provenencia_Engine_V1_SearchCatalogResponse = try await provenenciaCall(
            method: CoreMethod.searchCatalog,
            request: req
        )
        return resp.hits.map(Self.mapSearchHit)
    }

    func listSubjectTypes(projectDir: String) async throws -> [CatalogSubjectType] {
        var req = Provenencia_Engine_V1_ListSubjectTypesRequest()
        req.projectDir = projectDir
        let resp: Provenencia_Engine_V1_ListSubjectTypesResponse = try await provenenciaCall(
            method: CoreMethod.listSubjectTypes,
            request: req
        )
        return resp.types.map(Self.mapSubjectType)
    }

    func createSubject(
        projectDir: String,
        userID: String,
        sourceID: String,
        subjectTypeID: String,
        label: String,
        description: String,
        placement: CatalogGridCell?
    ) async throws -> CatalogSubject {
        var req = Provenencia_Engine_V1_CreateSubjectRequest()
        req.projectDir = projectDir
        req.userID = userID
        req.sourceID = sourceID
        req.subjectTypeID = subjectTypeID
        req.label = label
        req.description_p = description
        if let placement {
            req.hasPlacement_p = true
            req.gridX = placement.gridX
            req.gridY = placement.gridY
        }
        let resp: Provenencia_Engine_V1_CreateSubjectResponse = try await provenenciaCall(
            method: CoreMethod.createSubject,
            request: req
        )
        return Self.mapSubject(resp.subject)
    }

    func updateSubject(
        projectDir: String,
        userID: String,
        subjectID: String,
        label: String,
        description: String
    ) async throws -> CatalogSubject {
        var req = Provenencia_Engine_V1_UpdateSubjectRequest()
        req.projectDir = projectDir
        req.userID = userID
        req.subjectID = subjectID
        req.label = label
        req.description_p = description
        let resp: Provenencia_Engine_V1_UpdateSubjectResponse = try await provenenciaCall(
            method: CoreMethod.updateSubject,
            request: req
        )
        return Self.mapSubject(resp.subject)
    }

    func deleteSubject(projectDir: String, userID: String, subjectID: String) async throws {
        var req = Provenencia_Engine_V1_DeleteSubjectRequest()
        req.projectDir = projectDir
        req.userID = userID
        req.subjectID = subjectID
        let _: Provenencia_Engine_V1_DeleteSubjectResponse = try await provenenciaCall(
            method: CoreMethod.deleteSubject,
            request: req
        )
    }

    func listSubjects(projectDir: String, sourceID: String) async throws -> [CatalogSubject] {
        var req = Provenencia_Engine_V1_ListSubjectsRequest()
        req.projectDir = projectDir
        req.sourceID = sourceID
        let resp: Provenencia_Engine_V1_ListSubjectsResponse = try await provenenciaCall(
            method: CoreMethod.listSubjects,
            request: req
        )
        return resp.subjects.map(Self.mapSubject)
    }

    func setSubjectPosition(
        projectDir: String,
        subjectID: String,
        gridX: Int64,
        gridY: Int64
    ) async throws -> CatalogSubjectPosition {
        var req = Provenencia_Engine_V1_SetSubjectPositionRequest()
        req.projectDir = projectDir
        req.subjectID = subjectID
        req.gridX = gridX
        req.gridY = gridY
        let resp: Provenencia_Engine_V1_SetSubjectPositionResponse = try await provenenciaCall(
            method: CoreMethod.setSubjectPosition,
            request: req
        )
        return Self.mapSubjectPosition(resp.position)
    }

    func clearSubjectPosition(projectDir: String, subjectID: String) async throws {
        var req = Provenencia_Engine_V1_ClearSubjectPositionRequest()
        req.projectDir = projectDir
        req.subjectID = subjectID
        let _: Provenencia_Engine_V1_ClearSubjectPositionResponse = try await provenenciaCall(
            method: CoreMethod.clearSubjectPosition,
            request: req
        )
    }

    func listSubjectPositions(projectDir: String, sourceID: String) async throws -> [CatalogSubjectPosition] {
        var req = Provenencia_Engine_V1_ListSubjectPositionsRequest()
        req.projectDir = projectDir
        req.sourceID = sourceID
        let resp: Provenencia_Engine_V1_ListSubjectPositionsResponse = try await provenenciaCall(
            method: CoreMethod.listSubjectPositions,
            request: req
        )
        return resp.positions.map(Self.mapSubjectPosition)
    }

    func listProperties(projectDir: String) async throws -> [CatalogProperty] {
        var req = Provenencia_Engine_V1_ListPropertiesRequest()
        req.projectDir = projectDir
        let resp: Provenencia_Engine_V1_ListPropertiesResponse = try await provenenciaCall(
            method: CoreMethod.listProperties,
            request: req
        )
        return resp.properties.map(Self.mapProperty)
    }

    func createProperty(
        projectDir: String,
        userID: String,
        label: String,
        valueType: String,
        description: String
    ) async throws -> CatalogProperty {
        var req = Provenencia_Engine_V1_CreatePropertyRequest()
        req.projectDir = projectDir
        req.userID = userID
        req.label = label
        req.valueType = valueType
        req.description_p = description
        let resp: Provenencia_Engine_V1_CreatePropertyResponse = try await provenenciaCall(
            method: CoreMethod.createProperty,
            request: req
        )
        return Self.mapProperty(resp.property)
    }

    func updateProperty(
        projectDir: String,
        userID: String,
        propertyID: String,
        label: String,
        valueType: String,
        description: String
    ) async throws -> CatalogProperty {
        var req = Provenencia_Engine_V1_UpdatePropertyRequest()
        req.projectDir = projectDir
        req.userID = userID
        req.propertyID = propertyID
        req.label = label
        req.valueType = valueType
        req.description_p = description
        let resp: Provenencia_Engine_V1_UpdatePropertyResponse = try await provenenciaCall(
            method: CoreMethod.updateProperty,
            request: req
        )
        return Self.mapProperty(resp.property)
    }

    func deleteProperty(projectDir: String, userID: String, propertyID: String) async throws {
        var req = Provenencia_Engine_V1_DeletePropertyRequest()
        req.projectDir = projectDir
        req.userID = userID
        req.propertyID = propertyID
        let _: Provenencia_Engine_V1_DeletePropertyResponse = try await provenenciaCall(
            method: CoreMethod.deleteProperty,
            request: req
        )
    }

    func listPropertyTerms(projectDir: String, propertyID: String) async throws -> [CatalogPropertyTerm] {
        var req = Provenencia_Engine_V1_ListPropertyTermsRequest()
        req.projectDir = projectDir
        req.propertyID = propertyID
        let resp: Provenencia_Engine_V1_ListPropertyTermsResponse = try await provenenciaCall(
            method: CoreMethod.listPropertyTerms,
            request: req
        )
        return resp.terms.map(Self.mapPropertyTerm)
    }

    func createPropertyTerm(
        projectDir: String,
        userID: String,
        propertyID: String,
        label: String,
        description: String
    ) async throws -> CatalogPropertyTerm {
        var req = Provenencia_Engine_V1_CreatePropertyTermRequest()
        req.projectDir = projectDir
        req.userID = userID
        req.propertyID = propertyID
        req.label = label
        req.description_p = description
        let resp: Provenencia_Engine_V1_CreatePropertyTermResponse = try await provenenciaCall(
            method: CoreMethod.createPropertyTerm,
            request: req
        )
        return Self.mapPropertyTerm(resp.term)
    }

    func updatePropertyTerm(
        projectDir: String,
        userID: String,
        termID: String,
        label: String,
        description: String
    ) async throws -> CatalogPropertyTerm {
        var req = Provenencia_Engine_V1_UpdatePropertyTermRequest()
        req.projectDir = projectDir
        req.userID = userID
        req.termID = termID
        req.label = label
        req.description_p = description
        let resp: Provenencia_Engine_V1_UpdatePropertyTermResponse = try await provenenciaCall(
            method: CoreMethod.updatePropertyTerm,
            request: req
        )
        return Self.mapPropertyTerm(resp.term)
    }

    func deletePropertyTerm(projectDir: String, userID: String, termID: String) async throws {
        var req = Provenencia_Engine_V1_DeletePropertyTermRequest()
        req.projectDir = projectDir
        req.userID = userID
        req.termID = termID
        let _: Provenencia_Engine_V1_DeletePropertyTermResponse = try await provenenciaCall(
            method: CoreMethod.deletePropertyTerm,
            request: req
        )
    }

    func listSubjectTypeFields(projectDir: String, subjectTypeID: String) async throws -> [CatalogSubjectTypeField] {
        var req = Provenencia_Engine_V1_ListSubjectTypeFieldsRequest()
        req.projectDir = projectDir
        req.subjectTypeID = subjectTypeID
        let resp: Provenencia_Engine_V1_ListSubjectTypeFieldsResponse = try await provenenciaCall(
            method: CoreMethod.listSubjectTypeFields,
            request: req
        )
        return resp.fields.map(Self.mapSubjectTypeField)
    }

    func assignSubjectTypeField(
        projectDir: String,
        userID: String,
        subjectTypeID: String,
        propertyID: String
    ) async throws {
        var req = Provenencia_Engine_V1_AssignSubjectTypeFieldRequest()
        req.projectDir = projectDir
        req.userID = userID
        req.subjectTypeID = subjectTypeID
        req.propertyID = propertyID
        let _: Provenencia_Engine_V1_AssignSubjectTypeFieldResponse = try await provenenciaCall(
            method: CoreMethod.assignSubjectTypeField,
            request: req
        )
    }

    func removeSubjectTypeField(
        projectDir: String,
        userID: String,
        subjectTypeID: String,
        propertyID: String
    ) async throws {
        var req = Provenencia_Engine_V1_RemoveSubjectTypeFieldRequest()
        req.projectDir = projectDir
        req.userID = userID
        req.subjectTypeID = subjectTypeID
        req.propertyID = propertyID
        let _: Provenencia_Engine_V1_RemoveSubjectTypeFieldResponse = try await provenenciaCall(
            method: CoreMethod.removeSubjectTypeField,
            request: req
        )
    }

    func listPlaceableSubjectTypes() async throws -> [CatalogSubjectTypePresentation] {
        let req = Provenencia_Engine_V1_ListPlaceableSubjectTypesRequest()
        let resp: Provenencia_Engine_V1_ListPlaceableSubjectTypesResponse = try await provenenciaCall(
            method: CoreMethod.listPlaceableSubjectTypes,
            request: req
        )
        return resp.types.map(Self.mapSubjectTypePresentation)
    }

    func getSubjectTypePresentation(typeKey: String) async throws -> CatalogSubjectTypePresentation {
        var req = Provenencia_Engine_V1_GetSubjectTypePresentationRequest()
        req.typeKey = typeKey
        let resp: Provenencia_Engine_V1_GetSubjectTypePresentationResponse = try await provenenciaCall(
            method: CoreMethod.getSubjectTypePresentation,
            request: req
        )
        return Self.mapSubjectTypePresentation(resp.presentation)
    }

    func listConnectRules() async throws -> [CatalogConnectRule] {
        let req = Provenencia_Engine_V1_ListConnectRulesRequest()
        let resp: Provenencia_Engine_V1_ListConnectRulesResponse = try await provenenciaCall(
            method: CoreMethod.listConnectRules,
            request: req
        )
        return resp.rules.map(Self.mapConnectRule)
    }

    func createCitedBridge(
        projectDir: String,
        userID: String,
        sourceID: String,
        fromSubjectID: String,
        toSubjectID: String,
        bridgeTypeKey: String,
        description: String,
        artifactID: String,
        locatorJSON: String,
        transcription: String,
        citationDescription: String,
        transcriptionUncertain: Bool,
        transcriptionNote: String,
        citationNotes: [String],
        observations: [CatalogObservationDraft],
        citationID: String?
    ) async throws -> (CatalogSubject, CatalogCitation, [CatalogObservation]) {
        var req = Provenencia_Engine_V1_CreateCitedBridgeRequest()
        req.projectDir = projectDir
        req.userID = userID
        req.sourceID = sourceID
        req.fromSubjectID = fromSubjectID
        req.toSubjectID = toSubjectID
        req.bridgeTypeKey = bridgeTypeKey
        req.description_p = description
        req.artifactID = artifactID
        req.locatorJson = locatorJSON
        req.transcription = transcription
        req.citationDescription = citationDescription
        req.transcriptionUncertain = transcriptionUncertain
        req.transcriptionNote = transcriptionNote
        req.citationNotes = citationNotes
        req.observations = observations.map(Self.mapObservationDraft)
        req.citationID = citationID ?? ""
        let resp: Provenencia_Engine_V1_CreateCitedBridgeResponse = try await provenenciaCall(
            method: CoreMethod.createCitedBridge,
            request: req
        )
        return (
            Self.mapSubject(resp.subject),
            Self.mapCitation(resp.citation),
            resp.observations.map(Self.mapObservation)
        )
    }

    func createCitationWithObservations(
        projectDir: String,
        userID: String,
        artifactID: String,
        locatorJSON: String,
        transcription: String,
        description: String,
        transcriptionUncertain: Bool,
        transcriptionNote: String,
        citationNotes: [String],
        observations: [CatalogObservationDraft]
    ) async throws -> (CatalogCitation, [CatalogObservation]) {
        var req = Provenencia_Engine_V1_CreateCitationWithObservationsRequest()
        req.projectDir = projectDir
        req.userID = userID
        req.artifactID = artifactID
        req.locatorJson = locatorJSON
        req.transcription = transcription
        req.description_p = description
        req.transcriptionUncertain = transcriptionUncertain
        req.transcriptionNote = transcriptionNote
        req.citationNotes = citationNotes
        req.observations = observations.map(Self.mapObservationDraft)
        let resp: Provenencia_Engine_V1_CreateCitationWithObservationsResponse = try await provenenciaCall(
            method: CoreMethod.createCitationWithObservations,
            request: req
        )
        return (Self.mapCitation(resp.citation), resp.observations.map(Self.mapObservation))
    }

    func getCitation(
        projectDir: String,
        citationID: String
    ) async throws -> (CatalogCitation, [String], [CatalogObservation]) {
        var req = Provenencia_Engine_V1_GetCitationRequest()
        req.projectDir = projectDir
        req.citationID = citationID
        let resp: Provenencia_Engine_V1_GetCitationResponse = try await provenenciaCall(
            method: CoreMethod.getCitation,
            request: req
        )
        return (
            Self.mapCitation(resp.citation),
            resp.notes,
            resp.observations.map(Self.mapObservation)
        )
    }

    func addObservationsToCitation(
        projectDir: String,
        userID: String,
        citationID: String,
        observations: [CatalogObservationDraft]
    ) async throws -> [CatalogObservation] {
        var req = Provenencia_Engine_V1_AddObservationsToCitationRequest()
        req.projectDir = projectDir
        req.userID = userID
        req.citationID = citationID
        req.observations = observations.map(Self.mapObservationDraft)
        let resp: Provenencia_Engine_V1_AddObservationsToCitationResponse = try await provenenciaCall(
            method: CoreMethod.addObservationsToCitation,
            request: req
        )
        return resp.observations.map(Self.mapObservation)
    }

    func listObservationsBySource(projectDir: String, sourceID: String) async throws -> [CatalogObservation] {
        var req = Provenencia_Engine_V1_ListObservationsBySourceRequest()
        req.projectDir = projectDir
        req.sourceID = sourceID
        let resp: Provenencia_Engine_V1_ListObservationsBySourceResponse = try await provenenciaCall(
            method: CoreMethod.listObservationsBySource,
            request: req
        )
        return resp.observations.map(Self.mapObservation)
    }

    func citationCountsBySource(projectDir: String, sourceID: String) async throws -> [String: Int] {
        var req = Provenencia_Engine_V1_CitationCountsBySourceRequest()
        req.projectDir = projectDir
        req.sourceID = sourceID
        let resp: Provenencia_Engine_V1_CitationCountsBySourceResponse = try await provenenciaCall(
            method: CoreMethod.citationCountsBySource,
            request: req
        )
        var counts: [String: Int] = [:]
        for row in resp.counts {
            counts[row.artifactID] = Int(row.count)
        }
        return counts
    }

    func listSourceGraphProgress(projectDir: String) async throws -> [SourceGraphProgress] {
        var req = Provenencia_Engine_V1_ListSourceGraphProgressRequest()
        req.projectDir = projectDir
        let resp: Provenencia_Engine_V1_ListSourceGraphProgressResponse = try await provenenciaCall(
            method: CoreMethod.listSourceGraphProgress,
            request: req
        )
        return resp.rows.map(Self.mapSourceGraphProgress)
    }

    func getSourceGraphProgress(projectDir: String, sourceID: String) async throws -> SourceGraphProgress {
        var req = Provenencia_Engine_V1_GetSourceGraphProgressRequest()
        req.projectDir = projectDir
        req.sourceID = sourceID
        let resp: Provenencia_Engine_V1_GetSourceGraphProgressResponse = try await provenenciaCall(
            method: CoreMethod.getSourceGraphProgress,
            request: req
        )
        if resp.hasProgress {
            return Self.mapSourceGraphProgress(resp.progress)
        }
        return .zeros(sourceId: sourceID)
    }

    private static func mapSourceGraphProgress(_ row: Provenencia_Engine_V1_SourceGraphProgress) -> SourceGraphProgress {
        SourceGraphProgress(
            sourceId: row.sourceID,
            subjectCount: Int(row.subjectCount),
            observationCount: Int(row.observationCount)
        )
    }

    func updateCitation(
        projectDir: String,
        userID: String,
        citationID: String,
        locatorJSON: String,
        transcription: String,
        description: String,
        transcriptionUncertain: Bool,
        transcriptionNote: String
    ) async throws -> CatalogCitation {
        var req = Provenencia_Engine_V1_UpdateCitationRequest()
        req.projectDir = projectDir
        req.userID = userID
        req.citationID = citationID
        req.locatorJson = locatorJSON
        req.transcription = transcription
        req.description_p = description
        req.transcriptionUncertain = transcriptionUncertain
        req.transcriptionNote = transcriptionNote
        let resp: Provenencia_Engine_V1_UpdateCitationResponse = try await provenenciaCall(
            method: CoreMethod.updateCitation,
            request: req
        )
        return Self.mapCitation(resp.citation)
    }

    func updateObservation(
        projectDir: String,
        userID: String,
        observation: CatalogObservation
    ) async throws -> CatalogObservation {
        var req = Provenencia_Engine_V1_UpdateObservationRequest()
        req.projectDir = projectDir
        req.userID = userID
        req.observation = Self.mapCatalogObservation(observation)
        let resp: Provenencia_Engine_V1_UpdateObservationResponse = try await provenenciaCall(
            method: CoreMethod.updateObservation,
            request: req
        )
        return Self.mapObservation(resp.observation)
    }

    func deleteObservation(projectDir: String, userID: String, observationID: String) async throws {
        var req = Provenencia_Engine_V1_DeleteObservationRequest()
        req.projectDir = projectDir
        req.userID = userID
        req.observationID = observationID
        let _: Provenencia_Engine_V1_DeleteObservationResponse = try await provenenciaCall(
            method: CoreMethod.deleteObservation,
            request: req
        )
    }

    func getSubjectFieldsWorkspace(projectDir: String) async throws -> SubjectFieldsSnapshot {
        var req = Provenencia_Engine_V1_GetSubjectFieldsWorkspaceRequest()
        req.projectDir = projectDir
        let resp: Provenencia_Engine_V1_GetSubjectFieldsWorkspaceResponse = try await provenenciaCall(
            method: CoreMethod.getSubjectFieldsWorkspace,
            request: req
        )
        var fieldsByTypeID: [String: [CatalogSubjectTypeField]] = [:]
        var presentationsByKey: [String: CatalogSubjectTypePresentation] = [:]
        for group in resp.groups {
            fieldsByTypeID[group.subjectTypeID] = group.fields.map(Self.mapSubjectTypeField)
            if group.hasPresentation {
                let presentation = Self.mapSubjectTypePresentation(group.presentation)
                presentationsByKey[presentation.typeKey] = presentation
            }
        }
        return SubjectFieldsSnapshot(
            properties: resp.properties.map(Self.mapProperty),
            types: resp.types.map(Self.mapSubjectType),
            fieldsByTypeID: fieldsByTypeID,
            presentationsByKey: presentationsByKey
        )
    }

    func listCitationsByArtifact(projectDir: String, artifactID: String) async throws -> [CatalogListedCitation] {
        var req = Provenencia_Engine_V1_ListCitationsByArtifactRequest()
        req.projectDir = projectDir
        req.artifactID = artifactID
        let resp: Provenencia_Engine_V1_ListCitationsByArtifactResponse = try await provenenciaCall(
            method: CoreMethod.listCitationsByArtifact,
            request: req
        )
        return resp.citations.map { row in
            CatalogListedCitation(
                citation: Self.mapCitation(row.citation),
                observationCount: Int(row.observationCount)
            )
        }
    }

    private static func mapOriginCounts(
        _ c: Provenencia_Engine_V1_VocabularyOriginCounts
    ) -> WorkspaceNavOriginCounts {
        WorkspaceNavOriginCounts(
            total: Int(c.total),
            seeded: Int(c.seeded),
            user: Int(c.user),
            plugin: Int(c.plugin)
        )
    }

    private static func mapWorkspaceLocationToProto(
        _ loc: WorkspaceLocation
    ) -> Provenencia_Engine_V1_WorkspaceLocation {
        var p = Provenencia_Engine_V1_WorkspaceLocation()
        p.section = loc.section.rawValue
        p.sourceID = loc.sourceId ?? ""
        p.fieldID = loc.fieldId ?? ""
        p.typeID = loc.typeId ?? ""
        p.ref = loc.ref ?? ""
        p.title = loc.title ?? ""
        return p
    }

    private static func mapWorkspaceLocationFromProto(
        _ p: Provenencia_Engine_V1_WorkspaceLocation
    ) -> WorkspaceLocation {
        WorkspaceLocation(
            section: WorkspaceSection(rawValue: p.section) ?? .sources,
            sourceId: p.sourceID,
            fieldId: p.fieldID,
            typeId: p.typeID,
            ref: p.ref,
            title: p.title
        )
    }

    private static func mapSearchHit(_ h: Provenencia_Engine_V1_SearchHit) -> CatalogSearchHit {
        CatalogSearchHit(
            kind: h.kind,
            id: h.id,
            ref: h.ref,
            title: h.title,
            subtitle: h.subtitle,
            matchReason: h.matchReason,
            matchSnippet: h.matchSnippet,
            location: Self.mapWorkspaceLocationFromProto(h.location),
            thumbnailRelPath: h.thumbnailRelPath,
            iconKey: h.iconKey
        )
    }

    private static func mapProject(_ p: Provenencia_Engine_V1_ProjectInfo) -> ProjectInfo {
        ProjectInfo(
            label: p.label,
            folderName: p.folderName,
            createdAt: p.createdAt,
            updatedAt: p.updatedAt,
            updatedByUserID: p.updatedByUserID,
            updatedByDisplayName: p.updatedByDisplayName,
            updatedByRef: p.updatedByRef,
            uuid: p.uuid
        )
    }

    private static func mapSource(_ s: Provenencia_Engine_V1_Source) -> CatalogSource {
        CatalogSource(
            id: s.id,
            ref: s.ref,
            sourceTypeID: s.sourceTypeID,
            title: s.title,
            description: s.description_p,
            thumbnailRelPath: s.thumbnailRelPath,
            thumbnailMediaType: s.thumbnailMediaType,
            thumbnailOriginalFilename: s.thumbnailOriginalFilename,
            coverMode: s.coverMode.isEmpty ? "type_icon" : s.coverMode,
            primaryArtifactID: s.primaryArtifactID,
            hasArtifact: s.hasArtifact_p,
            updatedRevision: s.updatedRevision
        )
    }

    private static func mapNote(_ n: Provenencia_Engine_V1_SourceNote) -> CatalogSourceNote {
        CatalogSourceNote(
            id: n.id,
            sourceID: n.sourceID,
            body: n.body,
            authorDisplayName: n.authorDisplayName,
            createdAt: n.createdAt
        )
    }

    private static func mapFile(_ f: Provenencia_Engine_V1_SourceFileRef) -> CatalogFileRef {
        CatalogFileRef(
            id: f.id,
            relPath: f.relPath,
            originalFilename: f.originalFilename,
            mediaType: f.mediaType,
            byteSize: f.byteSize
        )
    }

    private static func mapArtifact(_ a: Provenencia_Engine_V1_Artifact) -> CatalogArtifact {
        CatalogArtifact(
            id: a.id,
            ref: a.ref,
            sourceID: a.sourceID,
            fileID: a.fileID,
            label: a.label,
            description: a.description_p,
            file: a.hasFile ? Self.mapFile(a.file) : nil,
            thumbnailRelPath: a.thumbnailRelPath
        )
    }

    private static func mapCredibilityGrade(_ g: Provenencia_Engine_V1_SourceCredibilityGrade) -> CatalogCredibilityGrade {
        CatalogCredibilityGrade(
            id: g.id,
            key: g.key,
            origin: g.origin,
            label: g.label,
            sortOrder: Int(g.sortOrder)
        )
    }

    private static func mapCredibility(_ a: Provenencia_Engine_V1_SourceCredibilityAssessment) -> CatalogCredibilityAssessment {
        CatalogCredibilityAssessment(
            id: a.id,
            sourceID: a.sourceID,
            gradeID: a.gradeID,
            gradeKey: a.gradeKey,
            gradeLabel: a.gradeLabel,
            argument: a.argument
        )
    }

    private static func mapSourceType(_ t: Provenencia_Engine_V1_SourceType) -> CatalogSourceType {
        CatalogSourceType(
            id: t.id,
            key: t.key,
            origin: t.origin,
            label: t.label,
            description: t.description_p,
            iconKey: t.iconKey.isEmpty ? PVMarkKey.defaultTypeMark.rawValue : t.iconKey,
            usedBy: Int(t.usedBy),
            suggestedFieldCount: Int(t.suggestedFieldCount)
        )
    }

    private static func mapSubjectType(_ t: Provenencia_Engine_V1_SubjectType) -> CatalogSubjectType {
        CatalogSubjectType(
            id: t.id,
            key: t.key,
            origin: t.origin,
            label: t.label,
            description: t.description_p,
            refPrefix: t.refPrefix,
            candidateRefPrefix: t.candidateRefPrefix
        )
    }

    private static func mapSubject(_ s: Provenencia_Engine_V1_Subject) -> CatalogSubject {
        CatalogSubject(
            id: s.id,
            ref: s.ref,
            sourceID: s.sourceID,
            subjectTypeID: s.subjectTypeID,
            label: s.label,
            description: s.description_p
        )
    }

    private static func mapSubjectPosition(_ p: Provenencia_Engine_V1_SubjectPosition) -> CatalogSubjectPosition {
        CatalogSubjectPosition(subjectID: p.subjectID, gridX: p.gridX, gridY: p.gridY)
    }

    private static func mapProperty(_ p: Provenencia_Engine_V1_Property) -> CatalogProperty {
        CatalogProperty(
            id: p.id,
            key: p.key,
            origin: p.origin,
            label: p.label,
            description: p.description_p,
            valueType: p.valueType,
            usedBy: Int(p.usedBy)
        )
    }

    private static func mapPropertyTerm(_ t: Provenencia_Engine_V1_PropertyTerm) -> CatalogPropertyTerm {
        CatalogPropertyTerm(
            id: t.id,
            propertyID: t.propertyID,
            key: t.key,
            origin: t.origin,
            label: t.label,
            description: t.description_p
        )
    }

    private static func mapSubjectTypeField(_ f: Provenencia_Engine_V1_SubjectTypeField) -> CatalogSubjectTypeField {
        CatalogSubjectTypeField(
            property: Self.mapProperty(f.property),
            sortOrder: Int(f.sortOrder),
            locked: f.locked
        )
    }

    private static func mapSubjectTypePresentation(
        _ p: Provenencia_Engine_V1_SubjectTypePresentation
    ) -> CatalogSubjectTypePresentation {
        CatalogSubjectTypePresentation(
            typeKey: p.typeKey,
            l10nKey: p.l10NKey,
            iconSymbol: p.iconSymbol,
            inkToken: p.inkToken,
            tintToken: p.tintToken,
            chipToken: p.chipToken,
            lineToken: p.lineToken,
            edgeFromToken: p.edgeFromToken,
            edgeToToken: p.edgeToToken,
            role: p.role,
            placeable: p.placeable,
            paletteSort: Int(p.paletteSort),
            requiresCitationAtCreate: p.requiresCitationAtCreate,
            label: p.label
        )
    }

    private static func mapConnectRule(_ r: Provenencia_Engine_V1_ConnectRule) -> CatalogConnectRule {
        CatalogConnectRule(
            fromTypeKey: r.fromTypeKey,
            toTypeKey: r.toTypeKey,
            bridgeTypeKey: r.bridgeTypeKey,
            edgePropertyKeys: r.edgePropertyKeys,
            disambiguation: r.disambiguation,
            refuse: r.refuse,
            edges: r.edges.map { CatalogConnectEdge(propertyKey: $0.propertyKey, endpointTypeKey: $0.endpointTypeKey) }
        )
    }

    private static func mapCitation(_ c: Provenencia_Engine_V1_Citation) -> CatalogCitation {
        CatalogCitation(
            id: c.id,
            ref: c.ref,
            artifactID: c.artifactID,
            locatorJSON: c.locatorJson,
            transcription: c.transcription,
            description: c.description_p,
            transcriptionUncertain: c.transcriptionUncertain,
            transcriptionNote: c.transcriptionNote
        )
    }

    private static func mapObservation(_ o: Provenencia_Engine_V1_Observation) -> CatalogObservation {
        CatalogObservation(
            id: o.id,
            ref: o.ref,
            citationID: o.citationID,
            subjectID: o.subjectID,
            propertyID: o.propertyID,
            polarity: o.polarity,
            valueText: o.valueText,
            valueInteger: o.hasValueInteger ? o.valueInteger : nil,
            valueDateID: o.valueDateID,
            date: o.hasDate ? Self.mapDateValue(o.date) : nil,
            valueNameID: o.valueNameID,
            nameForm: o.hasName ? o.name.form : "",
            nameParts: o.hasName ? o.name.parts.map { CatalogNameValuePart(value: $0.value, type: $0.type) } : [],
            valueSubjectID: o.valueSubjectID,
            valueTermID: o.valueTermID,
            propertyKey: o.propertyKey,
            propertyLabel: o.propertyLabel,
            propertyValueType: o.propertyValueType
        )
    }

    private static func mapCatalogObservation(_ o: CatalogObservation) -> Provenencia_Engine_V1_Observation {
        var msg = Provenencia_Engine_V1_Observation()
        msg.id = o.id
        msg.ref = o.ref
        msg.citationID = o.citationID
        msg.subjectID = o.subjectID
        msg.propertyID = o.propertyID
        msg.polarity = o.polarity
        msg.valueText = o.valueText
        if let valueInteger = o.valueInteger {
            msg.valueInteger = valueInteger
        }
        if let date = o.date {
            msg.date = Self.mapDateValueToProto(date)
        }
        msg.valueDateID = o.valueDateID
        if !o.nameForm.isEmpty {
            var name = Provenencia_Engine_V1_NameValueInput()
            name.form = o.nameForm
            for part in o.nameParts {
                var p = Provenencia_Engine_V1_NameValuePartInput()
                p.value = part.value
                p.type = part.type
                name.parts.append(p)
            }
            msg.name = name
        }
        msg.valueNameID = o.valueNameID
        msg.valueSubjectID = o.valueSubjectID
        msg.valueTermID = o.valueTermID
        msg.propertyKey = o.propertyKey
        msg.propertyLabel = o.propertyLabel
        msg.propertyValueType = o.propertyValueType
        return msg
    }

    private static func mapObservationDraft(_ d: CatalogObservationDraft) -> Provenencia_Engine_V1_ObservationDraft {
        var o = Provenencia_Engine_V1_ObservationDraft()
        o.subjectID = d.subjectID
        o.propertyID = d.propertyID
        o.polarity = d.polarity
        o.valueText = d.valueText
        if let valueInteger = d.valueInteger {
            o.valueInteger = valueInteger
        }
        if let date = d.date {
            o.date = Self.mapDateValueToProto(date)
        }
        o.valueDateID = d.valueDateID
        if !d.nameForm.isEmpty {
            var name = Provenencia_Engine_V1_NameValueInput()
            name.form = d.nameForm
            for part in d.nameParts {
                var p = Provenencia_Engine_V1_NameValuePartInput()
                p.value = part.value
                p.type = part.type
                name.parts.append(p)
            }
            o.name = name
        }
        o.valueNameID = d.valueNameID
        o.valueSubjectID = d.valueSubjectID
        o.valueTermID = d.valueTermID
        o.notes = d.notes
        return o
    }

    private static func mapDateValueToProto(_ d: CatalogDateValueInput) -> Provenencia_Engine_V1_DateValueInput {
        var p = Provenencia_Engine_V1_DateValueInput()
        p.kind = d.kind
        p.qualifier = d.qualifier
        p.calendar = d.calendar
        if let startYear = d.startYear { p.startYear = startYear }
        if let startMonth = d.startMonth { p.startMonth = startMonth }
        if let startDay = d.startDay { p.startDay = startDay }
        if let startHour = d.startHour { p.startHour = startHour }
        if let startMinute = d.startMinute { p.startMinute = startMinute }
        if let startSecond = d.startSecond { p.startSecond = startSecond }
        if let startMillisecond = d.startMillisecond { p.startMillisecond = startMillisecond }
        p.startTz = d.startTZ
        if let endYear = d.endYear { p.endYear = endYear }
        if let endMonth = d.endMonth { p.endMonth = endMonth }
        if let endDay = d.endDay { p.endDay = endDay }
        if let endHour = d.endHour { p.endHour = endHour }
        if let endMinute = d.endMinute { p.endMinute = endMinute }
        if let endSecond = d.endSecond { p.endSecond = endSecond }
        if let endMillisecond = d.endMillisecond { p.endMillisecond = endMillisecond }
        p.endTz = d.endTZ
        p.phrase = d.phrase
        return p
    }

    private static func mapTypeSuggestion(_ s: Provenencia_Engine_V1_TypeSuggestion) -> CatalogTypeSuggestion {
        CatalogTypeSuggestion(field: Self.mapMetadataField(s.field), sortOrder: Int(s.sortOrder))
    }

    private static func mapMetadataField(_ f: Provenencia_Engine_V1_MetadataField) -> CatalogMetadataField {
        CatalogMetadataField(
            id: f.id,
            key: f.key,
            origin: f.origin,
            label: f.label,
            dataType: f.dataType,
            description: f.description_p,
            usedBy: Int(f.usedBy)
        )
    }

    private static func mapMetadataEntry(_ e: Provenencia_Engine_V1_MetadataWorkspaceEntry) -> CatalogMetadataEntry {
        CatalogMetadataEntry(
            field: Self.mapMetadataField(e.field),
            valueText: e.valueText,
            hasValue: e.hasValue_p,
            suggested: e.suggested,
            sortOrder: e.sortOrder
        )
    }

    private static func mapDateValue(_ d: Provenencia_Engine_V1_DateValueInput) -> CatalogDateValueInput {
        CatalogDateValueInput(
            kind: d.kind,
            qualifier: d.qualifier,
            calendar: d.calendar,
            startYear: d.hasStartYear ? d.startYear : nil,
            startMonth: d.hasStartMonth ? d.startMonth : nil,
            startDay: d.hasStartDay ? d.startDay : nil,
            startHour: d.hasStartHour ? d.startHour : nil,
            startMinute: d.hasStartMinute ? d.startMinute : nil,
            startSecond: d.hasStartSecond ? d.startSecond : nil,
            startMillisecond: d.hasStartMillisecond ? d.startMillisecond : nil,
            startTZ: d.startTz,
            endYear: d.hasEndYear ? d.endYear : nil,
            endMonth: d.hasEndMonth ? d.endMonth : nil,
            endDay: d.hasEndDay ? d.endDay : nil,
            endHour: d.hasEndHour ? d.endHour : nil,
            endMinute: d.hasEndMinute ? d.endMinute : nil,
            endSecond: d.hasEndSecond ? d.endSecond : nil,
            endMillisecond: d.hasEndMillisecond ? d.endMillisecond : nil,
            endTZ: d.endTz,
            phrase: d.phrase
        )
    }
}
