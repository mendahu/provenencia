package ffi

import (
	"strconv"

	"github.com/mendahu/provenencia/api/ffi/handlers"
	"github.com/mendahu/provenencia/api/proto/engine"
	"github.com/mendahu/provenencia/core/apperr"
)

const (
	MethodUnspecified                       = int32(engine.Method_METHOD_UNSPECIFIED)
	MethodPing                              = int32(engine.Method_METHOD_PING)
	MethodGetVersion                        = int32(engine.Method_METHOD_GET_VERSION)
	MethodGetInstallIdentity                = int32(engine.Method_METHOD_GET_INSTALL_IDENTITY)
	MethodCompleteOnboarding                = int32(engine.Method_METHOD_COMPLETE_ONBOARDING)
	MethodRemoveInstallIdentity             = int32(engine.Method_METHOD_REMOVE_INSTALL_IDENTITY)
	MethodGetActiveProject                  = int32(engine.Method_METHOD_GET_ACTIVE_PROJECT)
	MethodOpenProject                       = int32(engine.Method_METHOD_OPEN_PROJECT)
	MethodRemoveActiveProject               = int32(engine.Method_METHOD_REMOVE_ACTIVE_PROJECT)
	MethodListProjectUsers                  = int32(engine.Method_METHOD_LIST_PROJECT_USERS)
	MethodSignOut                           = int32(engine.Method_METHOD_SIGN_OUT)
	MethodGetProjectInfo                    = int32(engine.Method_METHOD_GET_PROJECT_INFO)
	MethodListSources                       = int32(engine.Method_METHOD_LIST_SOURCES)
	MethodGetSourceWorkspace                = int32(engine.Method_METHOD_GET_SOURCE_WORKSPACE)
	MethodCreateSource                      = int32(engine.Method_METHOD_CREATE_SOURCE)
	MethodUpdateSource                      = int32(engine.Method_METHOD_UPDATE_SOURCE)
	MethodAddSourceNote                     = int32(engine.Method_METHOD_ADD_SOURCE_NOTE)
	MethodUpdateSourceNote                  = int32(engine.Method_METHOD_UPDATE_SOURCE_NOTE)
	MethodDeleteSourceNote                  = int32(engine.Method_METHOD_DELETE_SOURCE_NOTE)
	MethodSetSourceMetadata                 = int32(engine.Method_METHOD_SET_SOURCE_METADATA)
	MethodClearSourceMetadata               = int32(engine.Method_METHOD_CLEAR_SOURCE_METADATA)
	MethodCreateArtifact                    = int32(engine.Method_METHOD_CREATE_ARTIFACT)
	MethodIngestArtifactFile                = int32(engine.Method_METHOD_INGEST_ARTIFACT_FILE)
	MethodListSourceTypes                   = int32(engine.Method_METHOD_LIST_SOURCE_TYPES)
	MethodCreateSourceType                  = int32(engine.Method_METHOD_CREATE_SOURCE_TYPE)
	MethodListMetadataFields                = int32(engine.Method_METHOD_LIST_METADATA_FIELDS)
	MethodCreateMetadataField               = int32(engine.Method_METHOD_CREATE_METADATA_FIELD)
	MethodUpdateMetadataField               = int32(engine.Method_METHOD_UPDATE_METADATA_FIELD)
	MethodDeleteSourceType                  = int32(engine.Method_METHOD_DELETE_SOURCE_TYPE)
	MethodDeleteMetadataField               = int32(engine.Method_METHOD_DELETE_METADATA_FIELD)
	MethodUpdateSourceType                  = int32(engine.Method_METHOD_UPDATE_SOURCE_TYPE)
	MethodListTypeSuggestions               = int32(engine.Method_METHOD_LIST_TYPE_SUGGESTIONS)
	MethodAssignTypeField                   = int32(engine.Method_METHOD_ASSIGN_TYPE_FIELD)
	MethodRemoveTypeField                   = int32(engine.Method_METHOD_REMOVE_TYPE_FIELD)
	MethodGetWorkspaceNavCounts             = int32(engine.Method_METHOD_GET_WORKSPACE_NAV_COUNTS)
	MethodUpdateArtifact                    = int32(engine.Method_METHOD_UPDATE_ARTIFACT)
	MethodListSourceCredibilityGrades       = int32(engine.Method_METHOD_LIST_SOURCE_CREDIBILITY_GRADES)
	MethodUpsertSourceCredibilityAssessment = int32(engine.Method_METHOD_UPSERT_SOURCE_CREDIBILITY_ASSESSMENT)
	MethodDismissSourceMetadataSuggestion   = int32(engine.Method_METHOD_DISMISS_SOURCE_METADATA_SUGGESTION)
	MethodReorderSourceMetadata             = int32(engine.Method_METHOD_REORDER_SOURCE_METADATA)
	MethodEnsureFileThumbnail               = int32(engine.Method_METHOD_ENSURE_FILE_THUMBNAIL)
	MethodCloseCatalogSession               = int32(engine.Method_METHOD_CLOSE_CATALOG_SESSION)
	MethodSetSourceCover                    = int32(engine.Method_METHOD_SET_SOURCE_COVER)
	MethodSearchCatalog                     = int32(engine.Method_METHOD_SEARCH_CATALOG)
	MethodListSubjectTypes                  = int32(engine.Method_METHOD_LIST_SUBJECT_TYPES)
	MethodCreateSubject                     = int32(engine.Method_METHOD_CREATE_SUBJECT)
	MethodUpdateSubject                     = int32(engine.Method_METHOD_UPDATE_SUBJECT)
	MethodDeleteSubject                     = int32(engine.Method_METHOD_DELETE_SUBJECT)
	MethodListSubjects                      = int32(engine.Method_METHOD_LIST_SUBJECTS)
	MethodSetSubjectPosition                = int32(engine.Method_METHOD_SET_SUBJECT_POSITION)
	MethodClearSubjectPosition              = int32(engine.Method_METHOD_CLEAR_SUBJECT_POSITION)
	MethodListSubjectPositions              = int32(engine.Method_METHOD_LIST_SUBJECT_POSITIONS)
	MethodListProperties                    = int32(engine.Method_METHOD_LIST_PROPERTIES)
	MethodCreateProperty                    = int32(engine.Method_METHOD_CREATE_PROPERTY)
	MethodUpdateProperty                    = int32(engine.Method_METHOD_UPDATE_PROPERTY)
	MethodDeleteProperty                    = int32(engine.Method_METHOD_DELETE_PROPERTY)
	MethodListSubjectTypeFields             = int32(engine.Method_METHOD_LIST_SUBJECT_TYPE_FIELDS)
	MethodAssignSubjectTypeField            = int32(engine.Method_METHOD_ASSIGN_SUBJECT_TYPE_FIELD)
	MethodRemoveSubjectTypeField            = int32(engine.Method_METHOD_REMOVE_SUBJECT_TYPE_FIELD)
	MethodListPlaceableSubjectTypes         = int32(engine.Method_METHOD_LIST_PLACEABLE_SUBJECT_TYPES)
	MethodGetSubjectTypePresentation        = int32(engine.Method_METHOD_GET_SUBJECT_TYPE_PRESENTATION)
	MethodListConnectRules                  = int32(engine.Method_METHOD_LIST_CONNECT_RULES)
	MethodListPropertyTerms                 = int32(engine.Method_METHOD_LIST_PROPERTY_TERMS)
	MethodCreatePropertyTerm                = int32(engine.Method_METHOD_CREATE_PROPERTY_TERM)
	MethodUpdatePropertyTerm                = int32(engine.Method_METHOD_UPDATE_PROPERTY_TERM)
	MethodDeletePropertyTerm                = int32(engine.Method_METHOD_DELETE_PROPERTY_TERM)
	MethodCreateCitationWithObservations    = int32(engine.Method_METHOD_CREATE_CITATION_WITH_OBSERVATIONS)
	MethodAddObservationsToCitation         = int32(engine.Method_METHOD_ADD_OBSERVATIONS_TO_CITATION)
	MethodListObservationsBySource          = int32(engine.Method_METHOD_LIST_OBSERVATIONS_BY_SOURCE)
	MethodGetCitation                       = int32(engine.Method_METHOD_GET_CITATION)
	MethodUpdateCitationWithObservations    = int32(engine.Method_METHOD_UPDATE_CITATION_WITH_OBSERVATIONS)
	MethodCreateCitedBridge                 = int32(engine.Method_METHOD_CREATE_CITED_BRIDGE)
	MethodCitationCountsBySource            = int32(engine.Method_METHOD_CITATION_COUNTS_BY_SOURCE)
	MethodListCitationsByArtifact           = int32(engine.Method_METHOD_LIST_CITATIONS_BY_ARTIFACT)
)

// Call routes one coarse FFI operation to api/ffi/handlers.
func Call(method int32, in []byte) ([]byte, error) {
	switch method {
	case MethodPing:
		return handlers.Ping(in)
	case MethodGetVersion:
		return handlers.GetVersion(in)
	case MethodGetInstallIdentity:
		return handlers.GetInstallIdentity(in)
	case MethodCompleteOnboarding:
		return handlers.CompleteOnboarding(in)
	case MethodRemoveInstallIdentity:
		return handlers.RemoveInstallIdentity(in)
	case MethodGetActiveProject:
		return handlers.GetActiveProject(in)
	case MethodOpenProject:
		return handlers.OpenProject(in)
	case MethodRemoveActiveProject:
		return handlers.RemoveActiveProject(in)
	case MethodListProjectUsers:
		return handlers.ListProjectUsers(in)
	case MethodSignOut:
		return handlers.SignOut(in)
	case MethodGetProjectInfo:
		return handlers.GetProjectInfo(in)
	case MethodListSources:
		return handlers.ListSources(in)
	case MethodGetSourceWorkspace:
		return handlers.GetSourceWorkspace(in)
	case MethodCreateSource:
		return handlers.CreateSource(in)
	case MethodUpdateSource:
		return handlers.UpdateSource(in)
	case MethodAddSourceNote:
		return handlers.AddSourceNote(in)
	case MethodUpdateSourceNote:
		return handlers.UpdateSourceNote(in)
	case MethodDeleteSourceNote:
		return handlers.DeleteSourceNote(in)
	case MethodSetSourceMetadata:
		return handlers.SetSourceMetadata(in)
	case MethodClearSourceMetadata:
		return handlers.ClearSourceMetadata(in)
	case MethodCreateArtifact:
		return handlers.CreateArtifact(in)
	case MethodIngestArtifactFile:
		return handlers.IngestArtifactFile(in)
	case MethodListSourceTypes:
		return handlers.ListSourceTypes(in)
	case MethodCreateSourceType:
		return handlers.CreateSourceType(in)
	case MethodListMetadataFields:
		return handlers.ListMetadataFields(in)
	case MethodCreateMetadataField:
		return handlers.CreateMetadataField(in)
	case MethodUpdateMetadataField:
		return handlers.UpdateMetadataField(in)
	case MethodDeleteSourceType:
		return handlers.DeleteSourceType(in)
	case MethodDeleteMetadataField:
		return handlers.DeleteMetadataField(in)
	case MethodUpdateSourceType:
		return handlers.UpdateSourceType(in)
	case MethodListTypeSuggestions:
		return handlers.ListTypeSuggestions(in)
	case MethodAssignTypeField:
		return handlers.AssignTypeField(in)
	case MethodRemoveTypeField:
		return handlers.RemoveTypeField(in)
	case MethodGetWorkspaceNavCounts:
		return handlers.GetWorkspaceNavCounts(in)
	case MethodUpdateArtifact:
		return handlers.UpdateArtifact(in)
	case MethodListSourceCredibilityGrades:
		return handlers.ListSourceCredibilityGrades(in)
	case MethodUpsertSourceCredibilityAssessment:
		return handlers.UpsertSourceCredibilityAssessment(in)
	case MethodDismissSourceMetadataSuggestion:
		return handlers.DismissSourceMetadataSuggestion(in)
	case MethodReorderSourceMetadata:
		return handlers.ReorderSourceMetadata(in)
	case MethodEnsureFileThumbnail:
		return handlers.EnsureFileThumbnail(in)
	case MethodCloseCatalogSession:
		return handlers.CloseCatalogSession(in)
	case MethodSetSourceCover:
		return handlers.SetSourceCover(in)
	case MethodSearchCatalog:
		return handlers.SearchCatalog(in)
	case MethodListSubjectTypes:
		return handlers.ListSubjectTypes(in)
	case MethodCreateSubject:
		return handlers.CreateSubject(in)
	case MethodUpdateSubject:
		return handlers.UpdateSubject(in)
	case MethodDeleteSubject:
		return handlers.DeleteSubject(in)
	case MethodListSubjects:
		return handlers.ListSubjects(in)
	case MethodSetSubjectPosition:
		return handlers.SetSubjectPosition(in)
	case MethodClearSubjectPosition:
		return handlers.ClearSubjectPosition(in)
	case MethodListSubjectPositions:
		return handlers.ListSubjectPositions(in)
	case MethodListProperties:
		return handlers.ListProperties(in)
	case MethodCreateProperty:
		return handlers.CreateProperty(in)
	case MethodUpdateProperty:
		return handlers.UpdateProperty(in)
	case MethodDeleteProperty:
		return handlers.DeleteProperty(in)
	case MethodListSubjectTypeFields:
		return handlers.ListSubjectTypeFields(in)
	case MethodAssignSubjectTypeField:
		return handlers.AssignSubjectTypeField(in)
	case MethodRemoveSubjectTypeField:
		return handlers.RemoveSubjectTypeField(in)
	case MethodListPlaceableSubjectTypes:
		return handlers.ListPlaceableSubjectTypes(in)
	case MethodGetSubjectTypePresentation:
		return handlers.GetSubjectTypePresentation(in)
	case MethodListConnectRules:
		return handlers.ListConnectRules(in)
	case MethodListPropertyTerms:
		return handlers.ListPropertyTerms(in)
	case MethodCreatePropertyTerm:
		return handlers.CreatePropertyTerm(in)
	case MethodUpdatePropertyTerm:
		return handlers.UpdatePropertyTerm(in)
	case MethodDeletePropertyTerm:
		return handlers.DeletePropertyTerm(in)
	case MethodCreateCitationWithObservations:
		return handlers.CreateCitationWithObservations(in)
	case MethodAddObservationsToCitation:
		return handlers.AddObservationsToCitation(in)
	case MethodListObservationsBySource:
		return handlers.ListObservationsBySource(in)
	case MethodGetCitation:
		return handlers.GetCitation(in)
	case MethodUpdateCitationWithObservations:
		return handlers.UpdateCitationWithObservations(in)
	case MethodCreateCitedBridge:
		return handlers.CreateCitedBridge(in)
	case MethodCitationCountsBySource:
		return handlers.CitationCountsBySource(in)
	case MethodListCitationsByArtifact:
		return handlers.ListCitationsByArtifact(in)
	default:
		return nil, apperr.New(apperr.CodeInternalUnknownMethod, apperr.KindInternal, strconv.Itoa(int(method)))
	}
}
