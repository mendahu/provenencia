#if DEBUG
import Foundation

/// In-memory `GenealogyStore` for SwiftUI previews and tests. Not used in the shipped app.
/// Store methods run on the cooperative pool, so callers like the citation
/// composer can have several in flight at once. Every method body that touches
/// the stored properties below does so inside `withState`, which makes each
/// call's reads and writes atomic with respect to other calls. Tests set up and
/// inspect these properties directly, which is safe only while no store call is
/// in flight (i.e. before kicking work off or after awaiting it).
final class FakeStore: GenealogyStore, @unchecked Sendable {
    var identity: InstallIdentity?
    var activeProjectDir: String?
    var catalogUsers: [InstallIdentity]
    var projectInfos: [String: ProjectInfo] = [:]
    var sourcesByProject: [String: [CatalogSource]] = [:]
    var notesBySource: [String: [CatalogSourceNote]] = [:]
    var artifactsBySource: [String: [CatalogArtifact]] = [:]
    /// Every Evidence graph of a project. `createSource` and `copySource` add a
    /// Default graph, as the engine does; tests seed more directly, with ids that
    /// differ from their Source's.
    var evidenceGraphsByProject: [String: [CatalogEvidenceGraph]] = [:]
    /// What `getEvidenceGraphUnlinkImpact` returns, filtered to the Artifacts asked.
    var unlinkImpactBlocks: [CatalogEvidenceGraphUnlinkBlock] = []
    var credibilityBySource: [String: CatalogCredibilityAssessment] = [:]
    var credibilityGradesByProject: [String: [CatalogCredibilityGrade]] = [:]
    var sourceTypesByProject: [String: [CatalogSourceType]] = [:]
    var subjectTypesByProject: [String: [CatalogSubjectType]] = [:]
    /// Subjects by Evidence graph id.
    var subjectsByGraph: [String: [CatalogSubject]] = [:]
    /// Accepted Identity Claim per Subject id (Promote): the claim and its handle.
    var membershipBySubject: [String: CatalogSubjectMembership] = [:]
    /// Distinct handles that pin an Observation, for delete-impact confirms.
    var pinsByObservation: [String: [CatalogCanonicalEntity]] = [:]
    /// Queued graph-alignment proposals. The first is consumed; the last repeats.
    var promoteProposals: [CatalogPromoteGraphAlignmentProposal] = []
    /// Per-call delay before a proposal returns, consumed in call order, so
    /// tests can answer an earlier call after a later one.
    var promoteProposeDelays: [Duration] = []
    /// The decided rows the last proposal call held.
    var lastPromoteFixed: [CatalogPromoteGraphAlignmentFixed] = []
    /// Claim evidence per Subject id, as Go would compose it. A Subject with no
    /// entry gets empty evidence.
    var promoteEvidence: [String: CatalogClaimEvidence] = [:]
    /// Thrown by `getPromoteClaimEvidence` when set.
    var promoteEvidenceError: Error?
    /// The last Done the page filed, for tests.
    var lastPromoteBatch: [CatalogPromoteBatchRow]?
    /// Event and Place headers as Go would compose them, seeded by tests in
    /// list order. FakeStore does not reconcile, walk, or title them; Go owns
    /// those rules (`conclusionheaders`, `eventtitle`) and their tests.
    var seededEventHeaders: [CatalogEventHeader] = []
    var seededPlaceHeaders: [CatalogPlaceHeader] = []
    /// Event card titles Go would choose, per Evidence graph then Subject id.
    var eventTitlesByGraph: [String: [String: CatalogEventTitle]] = [:]
    /// A handle's detail as Go would compose it, seeded by tests. Persons
    /// without a seed fall back to the name mirror below.
    var seededConclusionDetails: [String: CatalogConclusionDetail] = [:]
    /// The Identity Claim Promote wrote per Subject id (confidence, argument).
    var claimBySubject: [String: CatalogIdentityClaim] = [:]
    var subjectPositionsBySubject: [String: CatalogSubjectPosition] = [:]
    /// Type↔field suggestion joins, keyed by source type id and held in the
    /// order they were assigned — the engine's `sort_order`.
    var suggestionsByType: [String: [CatalogTypeSuggestion]] = [:]
    var fieldsByProject: [String: [CatalogMetadataField]] = [:]
    var propertiesByProject: [String: [CatalogProperty]] = [:]
    var propertyTermsByProperty: [String: [CatalogPropertyTerm]] = [:]
    var subjectTypePropertiesByType: [String: [CatalogSubjectTypeProperty]] = [:]
    var subjectTypePresentations: [String: CatalogSubjectTypePresentation] = [:]
    var connectRules: [CatalogConnectRule] = CatalogConnectRule.productMatrix
    /// Observations by the Evidence graph of their Subject.
    var observationsByGraph: [String: [CatalogObservation]] = [:]
    var edgeObservationIDs: Set<String> = []
    var citationsByID: [String: CatalogCitation] = [:]
    var citationNotesByID: [String: [String]] = [:]
    var metadataBySource: [String: [CatalogMetadataEntry]] = [:]
    /// Project dir for which a catalog RPC has “held” a session (tests only).
    var heldCatalogProjectDir: String?
    /// Last `closeCatalogSession` argument (tests only).
    var lastClosedCatalogProjectDir: String?
    /// Ordered store calls for composer / graph tests.
    var recordedCalls: [String] = []
    /// Optional delay before each `createCitationWithObservations` (race tests).
    var createCitationDelayNanoseconds: UInt64 = 0
    /// Holds `promoteSubject` open this long, for tests of in-flight writes.
    var promoteSubjectDelayNanoseconds: UInt64 = 0
    /// When set, `listSources` throws instead of returning the in-memory list.
    var listSourcesError: Error?
    var listSubjectsCalls = 0
    var listSubjectTypesCalls = 0
    var getSourceWorkspaceCalls = 0
    var listEvidenceGraphProgressCalls = 0
    var getEvidenceGraphProgressCalls = 0
    var deleteSourceCalls = 0
    var deleteArtifactCalls = 0
    var deleteEvidenceGraphCalls = 0
    /// Optional Impact override keyed by entity id (tests).
    var deleteImpactByID: [String: CatalogDeleteImpact] = [:]
    /// Optional override for graph-progress rows by graph id (tests). Missing keys compute from subjects/observations.
    var graphProgressByGraph: [String: EvidenceGraphProgress] = [:]
    var listEvidenceGraphProgressError: Error?
    var listEvidenceGraphsError: Error?
    var getEvidenceGraphProgressError: Error?
    /// Monotonic stand-in for audit_transactions.revision (Sources Updated sort).
    private var nextAuditRevision: Int64 = 1
    /// When set, `searchCatalog` throws (omnibar error UI).
    var searchCatalogError: Error?
    /// When set, `workspaceNavCounts` throws instead of returning counts.
    var workspaceNavCountsError: Error?
    /// When set, `updateSource` throws (identity title/type/description saves).
    var updateSourceError: Error?
    /// When set, `addObservationsToCitation` throws after recording the call.
    var addObservationsToCitationError: Error?
    /// When set, `updateObservation` throws after recording the call.
    var updateObservationError: Error?
    private var depictionSequence = 0
    /// The temporary image path of every depiction input seen, in call order.
    private(set) var recordedDepictionPaths: [String] = []
    /// When set, `reorderSourceMetadata` throws (optimistic move should revert).
    var reorderSourceMetadataError: Error?
    /// When set, `unpromoteSubject` throws (the Unpromote confirm shows the error).
    var unpromoteSubjectError: Error?
    /// When set, `setPreferredLikeness` throws (the portrait model shows a danger toast).
    var setPreferredLikenessError: Error?
    /// When set, `setSubjectPosition` throws (Evidence graph drag should revert).
    var setSubjectPositionError: Error?
    /// When true, each `setSubjectPosition` suspends until `releaseSetSubjectPosition()`
    /// so generation-guard tests can order completions without sleeping.
    var holdsSetSubjectPosition = false
    /// Per-call errors assigned in call order; `nil` means that call succeeds.
    var setSubjectPositionErrors: [Error?] = []
    private var setSubjectPositionCallIndex = 0
    private var heldSetSubjectPositions: [CheckedContinuation<Void, Never>] = []
    var listSubjectsError: Error?
    var listConnectRulesError: Error?
    var createPropertyTermError: Error?
    /// When set, `setSourceMetadata` throws (field vs page error tests).
    var setSourceMetadataError: Error?
    /// When set, `addSourceNote` throws (`pageError` surfacing).
    var addSourceNoteError: Error?
    /// When set, `ingestArtifactFile` throws before mutating artifacts.
    var ingestArtifactFileError: Error?
    /// When set, `createSourceType` throws instead of the built-in duplicate/slug checks.
    var createSourceTypeError: Error?
    var lastResult = OnboardingResult(
        projectDir: "/tmp/robins-family.provenencia",
        userID: "00000000-0000-7000-8000-000000000001",
        displayName: "Jake Robins",
        ref: "USR-F4N2P",
        project: ProjectInfo(
            label: "Robins Family",
            folderName: "robins-family.provenencia",
            createdAt: "2026-01-01T00:00:00Z",
            updatedAt: "2026-01-01T00:00:00Z",
            updatedByUserID: "00000000-0000-7000-8000-000000000001",
            updatedByDisplayName: "Jake Robins",
            updatedByRef: "USR-F4N2P",
            uuid: "00000000-0000-7000-8000-0000000000aa"
        )
    )

    /// Guards all mutable state for the duration of each store call's
    /// synchronous work. Recursive so a helper that locks stays safe to call
    /// from inside another locked body.
    private let stateLock = NSRecursiveLock()

    init(
        identity: InstallIdentity? = nil,
        activeProjectDir: String? = nil,
        catalogUsers: [InstallIdentity] = [
            InstallIdentity(
                userID: "00000000-0000-7000-8000-000000000001",
                displayName: "Jake Robins",
                ref: "USR-F4N2P"
            )
        ]
    ) {
        self.identity = identity
        self.activeProjectDir = activeProjectDir
        self.catalogUsers = catalogUsers
    }

    func installIdentity(identityDir _: String) async throws -> InstallIdentity? {
        return withState {
            identity
        }
    }

    func completeOnboarding(
        identityDir _: String,
        parentDir: String,
        displayName: String,
        familyName: String
    ) async throws -> OnboardingResult {
        return withState {
            let folder = ProjectSlug.folderName(from: familyName)
            let projectDir = parentDir + "/" + folder
            let ref = identity?.ref ?? lastResult.ref
            let userID = identity?.userID ?? lastResult.userID
            identity = InstallIdentity(userID: userID, displayName: displayName, ref: ref)
            let now = ISO8601DateFormatter().string(from: Date())
            let info = ProjectInfo(
                label: familyName,
                folderName: folder,
                createdAt: now,
                updatedAt: now,
                updatedByUserID: userID,
                updatedByDisplayName: displayName,
                updatedByRef: ref,
                uuid: UUID().uuidString.lowercased()
            )
            let result = OnboardingResult(
                projectDir: projectDir,
                userID: userID,
                displayName: displayName,
                ref: ref,
                project: info
            )
            activeProjectDir = result.projectDir
            catalogUsers = [InstallIdentity(userID: userID, displayName: displayName, ref: ref)]
            projectInfos[projectDir] = info
            return result
        }
    }

    func activeProject(identityDir _: String) async throws -> String? {
        return withState {
            activeProjectDir
        }
    }

    func listProjectUsers(projectDir _: String) async throws -> [InstallIdentity] {
        return withState {
            catalogUsers
        }
    }

    func openProject(
        identityDir _: String,
        projectDir: String,
        displayName: String,
        adoptUserID: String
    ) async throws -> OnboardingResult {
        return withState {
            if !adoptUserID.isEmpty, let match = catalogUsers.first(where: { $0.userID == adoptUserID }) {
                identity = match
                activeProjectDir = projectDir
                let info = projectInfos[projectDir] ?? ProjectInfo(
                    label: ProjectSlug.labelFromFolder(projectDir),
                    folderName: URL(fileURLWithPath: projectDir).lastPathComponent,
                    createdAt: "",
                    updatedAt: "",
                    updatedByUserID: match.userID,
                    updatedByDisplayName: match.displayName,
                    updatedByRef: match.ref
                )
                return OnboardingResult(
                    projectDir: projectDir,
                    userID: match.userID,
                    displayName: match.displayName,
                    ref: match.ref,
                    project: info
                )
            }
            if identity == nil {
                identity = InstallIdentity(userID: lastResult.userID, displayName: displayName, ref: lastResult.ref)
            }
            let name = identity?.displayName ?? displayName
            let ref = identity?.ref ?? lastResult.ref
            let userID = identity?.userID ?? lastResult.userID
            let info = projectInfos[projectDir] ?? ProjectInfo(
                label: ProjectSlug.labelFromFolder(projectDir),
                folderName: URL(fileURLWithPath: projectDir).lastPathComponent,
                createdAt: "",
                updatedAt: "",
                updatedByUserID: userID,
                updatedByDisplayName: name,
                updatedByRef: ref
            )
            let result = OnboardingResult(
                projectDir: projectDir,
                userID: userID,
                displayName: name,
                ref: ref,
                project: info
            )
            activeProjectDir = projectDir
            projectInfos[projectDir] = info
            return result
        }
    }

    func removeActiveProject(identityDir _: String) async throws {
        withState {
            activeProjectDir = nil
        }
    }

    func signOut(identityDir _: String) async throws {
        withState {
            activeProjectDir = nil
            identity = nil
            heldCatalogProjectDir = nil
        }
    }

    func closeCatalogSession(projectDir: String) async throws {
        withState {
            lastClosedCatalogProjectDir = projectDir
            if heldCatalogProjectDir == projectDir {
                heldCatalogProjectDir = nil
            }
        }
    }

    func projectInfo(projectDir: String) async throws -> ProjectInfo {
        return withState {
            if let info = projectInfos[projectDir] {
                return info
            }
            return ProjectInfo(
                label: ProjectSlug.labelFromFolder(projectDir),
                folderName: URL(fileURLWithPath: projectDir).lastPathComponent,
                createdAt: "",
                updatedAt: "",
                updatedByUserID: "",
                updatedByDisplayName: "",
                updatedByRef: ""
            )
        }
    }

    func listSources(projectDir: String) async throws -> [CatalogSource] {
        return try withState {
            markCatalogSessionHeld(projectDir)
            if let listSourcesError { throw listSourcesError }
            // Match Go `sources.List`: newest-created-first (UUIDv7 / id DESC).
            let rows = (sourcesByProject[projectDir] ?? []).sorted { $0.id > $1.id }
            return rows.map { enrichCoverFields($0) }
        }
    }

    func getSourceWorkspace(projectDir: String, sourceID: String) async throws -> CatalogSourceWorkspace {
        return withState {
            getSourceWorkspaceCalls += 1
            markCatalogSessionHeld(projectDir)
            let raw = (sourcesByProject[projectDir] ?? []).first { $0.id == sourceID }
                ?? CatalogSource(id: sourceID, ref: "SRC-XXXXX", sourceTypeID: "", title: "", description: "")
            return CatalogSourceWorkspace(
                source: enrichCoverFields(raw),
                notes: notesBySource[sourceID] ?? [],
                metadata: metadataBySource[sourceID] ?? [],
                artifacts: artifactsBySource[sourceID] ?? [],
                credibility: credibilityBySource[sourceID]
            )
        }
    }

    func createSource(
        projectDir: String,
        userID _: String,
        sourceTypeID: String,
        title: String,
        description: String
    ) async throws -> (CatalogSource, CatalogWriteEffects) {
        return withState {
            let source = CatalogSource(
                id: UUID().uuidString.lowercased(),
                ref: "SRC-FAKE1",
                sourceTypeID: sourceTypeID,
                title: title,
                description: description
            )
            return insertSource(source, projectDir: projectDir)
        }
    }

    func copySource(
        projectDir: String,
        userID _: String,
        sourceID: String,
        title: String,
        sourceTypeID: String,
        copyDescription: Bool,
        copyMetadata: Bool,
        copyCredibility: Bool,
        artifactIDs: [String]
    ) async throws -> (CatalogSource, CatalogWriteEffects) {
        return try withState {
            guard let original = sourcesByProject[projectDir]?.first(where: { $0.id == sourceID }) else {
                throw StoreBoom.boom
            }
            let copyID = UUID().uuidString.lowercased()
            var copy = CatalogSource(
                id: copyID,
                ref: "SRC-FAKE1",
                sourceTypeID: sourceTypeID,
                title: title,
                description: copyDescription ? original.description : ""
            )
            if copyMetadata {
                // Entries copy as they are; a `suggested` flag can be stale under
                // another type in the fake.
                metadataBySource[copyID] = metadataBySource[sourceID]
            }
            if copyCredibility, var assessment = credibilityBySource[sourceID] {
                assessment.id = UUID().uuidString.lowercased()
                assessment.sourceID = copyID
                credibilityBySource[copyID] = assessment
            }
            let chosen = (artifactsBySource[sourceID] ?? []).filter { artifactIDs.contains($0.id) }
            if !chosen.isEmpty {
                artifactsBySource[copyID] = chosen.map { art in
                    var copied = art
                    copied.id = UUID().uuidString.lowercased()
                    copied.ref = "ART-FAKE1"
                    copied.sourceID = copyID
                    return copied
                }
                copy.hasArtifact = true
            }
            return insertSource(copy, projectDir: projectDir)
        }
    }

    func updateSource(
        projectDir: String,
        userID _: String,
        sourceID: String,
        sourceTypeID: String,
        title: String,
        description: String
    ) async throws -> CatalogSource {
        return try withState {
            if let updateSourceError { throw updateSourceError }
            var list = sourcesByProject[projectDir] ?? []
            guard let idx = list.firstIndex(where: { $0.id == sourceID }) else {
                throw StoreBoom.boom
            }
            list[idx].sourceTypeID = sourceTypeID
            list[idx].title = title
            list[idx].description = description
            list[idx].updated = CatalogActivity(revision: nextAuditRevision)
            nextAuditRevision += 1
            sourcesByProject[projectDir] = list
            return enrichCoverFields(list[idx])
        }
    }

    func setSourceCover(
        projectDir: String,
        userID _: String,
        sourceID: String,
        coverMode: String,
        primaryArtifactID: String
    ) async throws -> CatalogSource {
        return try withState {
            var list = sourcesByProject[projectDir] ?? []
            guard let idx = list.firstIndex(where: { $0.id == sourceID }) else {
                throw StoreBoom.boom
            }
            switch coverMode {
            case "type_icon":
                list[idx].coverMode = "type_icon"
                list[idx].primaryArtifactID = ""
            case "artifact":
                let arts = artifactsBySource[sourceID] ?? []
                guard let art = arts.first(where: { $0.id == primaryArtifactID }),
                      !art.thumbnailRelPath.isEmpty
                else {
                    throw StoreBoom.boom
                }
                list[idx].coverMode = "artifact"
                list[idx].primaryArtifactID = primaryArtifactID
            default:
                throw StoreBoom.boom
            }
            list[idx].updated = CatalogActivity(revision: nextAuditRevision)
            nextAuditRevision += 1
            sourcesByProject[projectDir] = list
            return enrichCoverFields(list[idx])
        }
    }

    func addSourceNote(projectDir _: String, userID: String, sourceID: String, body: String) async throws
        -> (CatalogSourceNote, CatalogWriteEffects)
    {
        return try withState {
            if let addSourceNoteError { throw addSourceNoteError }
            let author = catalogUsers.first { $0.userID == userID }?.displayName
                ?? identity?.displayName
                ?? ""
            let formatter = ISO8601DateFormatter()
            formatter.formatOptions = [.withInternetDateTime]
            let note = CatalogSourceNote(
                id: UUID().uuidString.lowercased(),
                sourceID: sourceID,
                body: body,
                authorDisplayName: author,
                createdAt: formatter.string(from: Date())
            )
            notesBySource[sourceID, default: []].append(note)
            return (note, bumpSource(sourceID, graphs: []))
        }
    }

    func updateSourceNote(projectDir _: String, userID _: String, noteID: String, body: String) async throws
        -> (CatalogSourceNote, CatalogWriteEffects)
    {
        return try withState {
            for (sourceID, notes) in notesBySource {
                if let idx = notes.firstIndex(where: { $0.id == noteID }) {
                    var copy = notes
                    copy[idx].body = body
                    notesBySource[sourceID] = copy
                    return (copy[idx], bumpSource(sourceID, graphs: []))
                }
            }
            throw StoreBoom.boom
        }
    }

    func deleteSource(projectDir: String, userID _: String, sourceID: String) async throws {
        withState {
            markCatalogSessionHeld(projectDir)
            deleteSourceCalls += 1
            sourcesByProject[projectDir] = (sourcesByProject[projectDir] ?? []).filter { $0.id != sourceID }
            notesBySource[sourceID] = nil
            artifactsBySource[sourceID] = nil
            for graphID in graphIDsOfSource(sourceID) {
                subjectsByGraph[graphID] = nil
                observationsByGraph[graphID] = nil
                eventTitlesByGraph[graphID] = nil
                graphProgressByGraph[graphID] = nil
            }
            for (project, graphs) in evidenceGraphsByProject {
                evidenceGraphsByProject[project] = graphs.filter { $0.sourceID != sourceID }
            }
            metadataBySource[sourceID] = nil
            credibilityBySource[sourceID] = nil
        }
    }

    func deleteSourceNote(projectDir _: String, userID _: String, noteID: String) async throws -> CatalogWriteEffects {
        withState {
            var effects = CatalogWriteEffects.empty
            for (sourceID, notes) in notesBySource where notes.contains(where: { $0.id == noteID }) {
                notesBySource[sourceID] = notes.filter { $0.id != noteID }
                effects = bumpSource(sourceID, graphs: [])
            }
            return effects
        }
    }

    func setSourceMetadata(
        projectDir: String,
        userID _: String,
        sourceID: String,
        fieldID: String,
        valueText: String
    ) async throws -> (CatalogMetadataEntry, CatalogWriteEffects) {
        return try withState {
            if let setSourceMetadataError { throw setSourceMetadataError }
            var list = metadataBySource[sourceID] ?? []
            let entry: CatalogMetadataEntry
            if let idx = list.firstIndex(where: { $0.field.id == fieldID }) {
                list[idx].valueText = valueText
                list[idx].hasValue = true
                entry = list[idx]
            } else {
                let field = (fieldsByProject[projectDir] ?? []).first { $0.id == fieldID }
                    ?? CatalogMetadataField(
                        id: fieldID, key: "field", origin: "user", label: "Field", dataType: "text", description: ""
                    )
                let order = Int32(list.count)
                entry = CatalogMetadataEntry(
                    field: field,
                    valueText: valueText,
                    hasValue: true,
                    suggested: false,
                    sortOrder: order
                )
                list.append(entry)
            }
            metadataBySource[sourceID] = list
            return (entry, bumpSource(sourceID, graphs: []))
        }
    }

    func clearSourceMetadata(
        projectDir _: String, userID _: String, sourceID: String, fieldID: String
    ) async throws -> CatalogWriteEffects {
        withState {
            var list = metadataBySource[sourceID] ?? []
            guard let idx = list.firstIndex(where: { $0.field.id == fieldID }) else { return .empty }
            if list[idx].suggested {
                list[idx].valueText = ""
                list[idx].hasValue = false
            } else {
                list.remove(at: idx)
            }
            metadataBySource[sourceID] = list
            return bumpSource(sourceID, graphs: [])
        }
    }

    func dismissSourceMetadataSuggestion(
        projectDir _: String,
        userID _: String,
        sourceID: String,
        fieldID: String
    ) async throws -> ([CatalogMetadataEntry], CatalogWriteEffects) {
        return withState {
            var list = metadataBySource[sourceID] ?? []
            list.removeAll { $0.field.id == fieldID && !$0.hasValue }
            metadataBySource[sourceID] = list
            return (list, bumpSource(sourceID, graphs: []))
        }
    }

    func reorderSourceMetadata(
        projectDir _: String,
        userID _: String,
        sourceID: String,
        fieldIDs: [String]
    ) async throws -> ([CatalogMetadataEntry], CatalogWriteEffects) {
        return try withState {
            if let reorderSourceMetadataError { throw reorderSourceMetadataError }
            let current = metadataBySource[sourceID] ?? []
            var byID = Dictionary(uniqueKeysWithValues: current.map { ($0.field.id, $0) })
            var next: [CatalogMetadataEntry] = []
            for (i, id) in fieldIDs.enumerated() {
                guard var entry = byID.removeValue(forKey: id) else { continue }
                entry.sortOrder = Int32(i)
                next.append(entry)
            }
            for (_, leftover) in byID {
                var entry = leftover
                entry.sortOrder = Int32(next.count)
                next.append(entry)
            }
            metadataBySource[sourceID] = next
            return (next, bumpSource(sourceID, graphs: []))
        }
    }

    func createArtifact(
        projectDir: String,
        userID _: String,
        sourceID: String,
        fileID: String,
        label: String,
        description: String,
        graphIDs: [String]
    ) async throws -> (CatalogArtifact, CatalogWriteEffects) {
        return withState {
            let art = CatalogArtifact(
                id: UUID().uuidString.lowercased(),
                ref: "ART-FAKE1",
                sourceID: sourceID,
                fileID: fileID,
                label: label,
                description: description,
                file: nil
            )
            artifactsBySource[sourceID, default: []].append(art)
            var graphs = evidenceGraphsByProject[projectDir] ?? []
            for i in graphs.indices where graphs[i].sourceID == sourceID
                && (graphIDs.isEmpty || graphIDs.contains(graphs[i].id))
            {
                graphs[i].artifactIDs.append(art.id)
            }
            evidenceGraphsByProject[projectDir] = graphs
            var effects = bumpSource(sourceID, graphs: graphIDsOfSource(sourceID))
            effects.lists.insert(.evidenceGraphs)
            return (art, effects)
        }
    }

    func updateArtifact(
        projectDir _: String,
        userID _: String,
        artifactID: String,
        label: String,
        description: String
    ) async throws -> (CatalogArtifact, CatalogWriteEffects) {
        return try withState {
            for (sourceID, arts) in artifactsBySource {
                if let idx = arts.firstIndex(where: { $0.id == artifactID }) {
                    var copy = arts
                    copy[idx].label = label
                    copy[idx].description = description
                    artifactsBySource[sourceID] = copy
                    return (copy[idx], bumpSource(sourceID, graphs: graphIDsOfSource(sourceID)))
                }
            }
            throw StoreBoom.boom
        }
    }

    func deleteArtifact(
        projectDir: String, userID _: String, artifactID: String
    ) async throws -> CatalogWriteEffects {
        withState {
            markCatalogSessionHeld(projectDir)
            deleteArtifactCalls += 1
            var effects = CatalogWriteEffects.empty
            for (sourceID, arts) in artifactsBySource where arts.contains(where: { $0.id == artifactID }) {
                artifactsBySource[sourceID] = arts.filter { $0.id != artifactID }
                effects = bumpSource(sourceID, graphs: graphIDsOfSource(sourceID))
            }
            for (project, var graphs) in evidenceGraphsByProject {
                for i in graphs.indices {
                    graphs[i].artifactIDs.removeAll { $0 == artifactID }
                }
                evidenceGraphsByProject[project] = graphs
            }
            for (project, var sources) in sourcesByProject {
                for i in sources.indices where sources[i].primaryArtifactID == artifactID {
                    sources[i].primaryArtifactID = ""
                    sources[i].coverMode = "type_icon"
                    sources[i].thumbnailRelPath = ""
                }
                sourcesByProject[project] = sources
            }
            citationsByID = citationsByID.filter { $0.value.artifactID != artifactID }
            return effects
        }
    }

    func listEvidenceGraphs(projectDir: String) async throws -> [CatalogEvidenceGraph] {
        try withState {
            if let listEvidenceGraphsError { throw listEvidenceGraphsError }
            markCatalogSessionHeld(projectDir)
            return (evidenceGraphsByProject[projectDir] ?? []).sorted(by: Self.graphListOrder)
        }
    }

    func createEvidenceGraph(
        projectDir: String,
        userID _: String,
        sourceID: String,
        name: String,
        description: String,
        artifactIDs: [String]
    ) async throws -> (CatalogEvidenceGraph, CatalogWriteEffects) {
        try withState {
            markCatalogSessionHeld(projectDir)
            let trimmed = name.trimmingCharacters(in: .whitespacesAndNewlines)
            guard !trimmed.isEmpty, !artifactIDs.isEmpty else { throw StoreBoom.boom }
            let graph = appendGraph(
                projectDir: projectDir, sourceID: sourceID, name: trimmed, description: description,
                artifactIDs: artifactIDs
            )
            return (graph, Self.graphEffects(sourceID: sourceID))
        }
    }

    func updateEvidenceGraph(
        projectDir: String,
        userID _: String,
        graphID: String,
        name: String,
        description: String,
        artifactIDs: [String]
    ) async throws -> CatalogWriteEffects {
        try withState {
            markCatalogSessionHeld(projectDir)
            guard let idx = evidenceGraphsByProject[projectDir]?.firstIndex(where: { $0.id == graphID }) else {
                throw StoreBoom.boom
            }
            evidenceGraphsByProject[projectDir]?[idx].name = name.trimmingCharacters(in: .whitespacesAndNewlines)
            evidenceGraphsByProject[projectDir]?[idx].description = description
            if !artifactIDs.isEmpty {
                evidenceGraphsByProject[projectDir]?[idx].artifactIDs = artifactIDs
            }
            return Self.graphEffects(sourceID: evidenceGraphsByProject[projectDir]?[idx].sourceID)
        }
    }

    func deleteEvidenceGraph(projectDir: String, userID _: String, graphID: String) async throws -> CatalogWriteEffects {
        try withState {
            markCatalogSessionHeld(projectDir)
            deleteEvidenceGraphCalls += 1
            guard evidenceGraphsByProject[projectDir]?.contains(where: { $0.id == graphID }) == true else {
                throw StoreBoom.boom
            }
            evidenceGraphsByProject[projectDir]?.removeAll { $0.id == graphID }
            return Self.graphEffects(sourceID: nil)
        }
    }

    func linkEvidenceGraphArtifact(
        projectDir: String,
        userID _: String,
        graphID: String,
        artifactID: String
    ) async throws -> CatalogWriteEffects {
        try withState {
            markCatalogSessionHeld(projectDir)
            guard let idx = evidenceGraphsByProject[projectDir]?.firstIndex(where: { $0.id == graphID }) else {
                throw StoreBoom.boom
            }
            if evidenceGraphsByProject[projectDir]?[idx].artifactIDs.contains(artifactID) == false {
                evidenceGraphsByProject[projectDir]?[idx].artifactIDs.append(artifactID)
            }
            return Self.graphEffects(sourceID: evidenceGraphsByProject[projectDir]?[idx].sourceID)
        }
    }

    func unlinkEvidenceGraphArtifact(
        projectDir: String,
        userID _: String,
        graphID: String,
        artifactID: String
    ) async throws -> CatalogWriteEffects {
        try withState {
            markCatalogSessionHeld(projectDir)
            guard let idx = evidenceGraphsByProject[projectDir]?.firstIndex(where: { $0.id == graphID }) else {
                throw StoreBoom.boom
            }
            evidenceGraphsByProject[projectDir]?[idx].artifactIDs.removeAll { $0 == artifactID }
            return Self.graphEffects(sourceID: evidenceGraphsByProject[projectDir]?[idx].sourceID)
        }
    }

    func getEvidenceGraphUnlinkImpact(
        projectDir: String,
        graphID _: String,
        artifactIDs: [String]
    ) async throws -> [CatalogEvidenceGraphUnlinkBlock] {
        withState {
            markCatalogSessionHeld(projectDir)
            return artifactIDs.compactMap { id in unlinkImpactBlocks.first { $0.artifactID == id } }
        }
    }

    /// Every graph write reports the graph list, as Go does. A create, link,
    /// unlink or edit also reports its Source for the citation reads (graph
    /// delete has no Source to report: the link rows go with it).
    private static func graphEffects(sourceID: String?) -> CatalogWriteEffects {
        CatalogWriteEffects(citationSources: sourceID.map { [$0] } ?? [], lists: [.evidenceGraphs])
    }

    func ingestArtifactFile(
        projectDir: String,
        userID _: String,
        artifactID: String,
        path: String
    ) async throws -> (artifact: CatalogArtifact, file: CatalogFileRef, reused: Bool, effects: CatalogWriteEffects) {
        return try withState {
            if let ingestArtifactFileError { throw ingestArtifactFileError }
            let filename = URL(fileURLWithPath: path).lastPathComponent
            let ext = URL(fileURLWithPath: path).pathExtension.lowercased()
            let mediaType: String = switch ext {
            case "png": "image/png"
            case "jpg", "jpeg": "image/jpeg"
            case "gif": "image/gif"
            case "webp": "image/webp"
            case "tif", "tiff": "image/tiff"
            case "pdf": "application/pdf"
            case "mp4": "video/mp4"
            case "mov": "video/quicktime"
            case "mp3": "audio/mpeg"
            case "wav": "audio/wav"
            case "txt": "text/plain"
            case "csv": "text/csv"
            case "doc", "docx": "application/msword"
            default: "application/octet-stream"
            }
            let file = CatalogFileRef(
                id: UUID().uuidString.lowercased(),
                relPath: "objects/aa/bb/aabb",
                originalFilename: filename,
                mediaType: mediaType,
                byteSize: 0
            )
            for (sourceID, arts) in artifactsBySource {
                if let idx = arts.firstIndex(where: { $0.id == artifactID }) {
                    if !arts[idx].fileID.isEmpty {
                        throw CoreInvokeError.coded(
                            status: 1,
                            code: "artifacts.file_already_attached",
                            kind: .conflict,
                            params: []
                        )
                    }
                    var copy = arts
                    copy[idx].fileID = file.id
                    copy[idx].file = file
                    if ["png", "jpg", "jpeg", "gif", "webp", "tif", "tiff"].contains(ext) {
                        copy[idx].thumbnailRelPath = "objects/aa/bb/thumb-\(file.id.prefix(8))"
                    }
                    artifactsBySource[sourceID] = copy
                    return (copy[idx], file, false, bumpSource(sourceID, graphs: graphIDsOfSource(sourceID)))
                }
            }
            throw StoreBoom.boom
        }
    }

    func listSourceCredibilityGrades(projectDir: String) async throws -> [CatalogCredibilityGrade] {
        return withState {
            if let grades = credibilityGradesByProject[projectDir], !grades.isEmpty {
                return grades
            }
            return [
                CatalogCredibilityGrade(id: "g-low", key: "low_trust", origin: CatalogOrigin.provenencia, label: "Low trust", sortOrder: 1),
                CatalogCredibilityGrade(id: "g-std", key: "standard", origin: CatalogOrigin.provenencia, label: "Standard", sortOrder: 2),
                CatalogCredibilityGrade(id: "g-high", key: "high_trust", origin: CatalogOrigin.provenencia, label: "High trust", sortOrder: 3),
            ]
        }
    }

    func upsertSourceCredibilityAssessment(
        projectDir: String,
        userID _: String,
        sourceID: String,
        gradeID: String,
        argument: String
    ) async throws -> (CatalogCredibilityAssessment, CatalogWriteEffects) {
        let grades = try await listSourceCredibilityGrades(projectDir: projectDir)
        let grade = grades.first { $0.id == gradeID } ?? grades[1]
        return withState {
            let assessment = CatalogCredibilityAssessment(
                id: credibilityBySource[sourceID]?.id ?? UUID().uuidString.lowercased(),
                sourceID: sourceID,
                gradeID: grade.id,
                gradeKey: grade.key,
                gradeLabel: grade.label,
                argument: argument
            )
            credibilityBySource[sourceID] = assessment
            return (assessment, bumpSource(sourceID, graphs: []))
        }
    }

    func listSourceTypes(projectDir: String) async throws -> [CatalogSourceType] {
        return withState {
            markCatalogSessionHeld(projectDir)
            return (sourceTypesByProject[projectDir] ?? []).map(withSuggestedFieldCount)
        }
    }

    /// The engine derives this column in its list query, so the fake keeps it
    /// in step with `suggestionsByType` rather than making tests set it.
    private func withSuggestedFieldCount(_ type: CatalogSourceType) -> CatalogSourceType {
        var copy = type
        copy.suggestedFieldCount = (suggestionsByType[type.id] ?? []).count
        return copy
    }

    func createSourceType(
        projectDir: String,
        userID _: String,
        label: String,
        description: String,
        iconKey: String
    ) async throws -> CatalogSourceType {
        return try withState {
            if let createSourceTypeError { throw createSourceTypeError }
            let key = FieldSlug.kebab(label)
            if key.isEmpty {
                throw StoreBoom.boom
            }
            if (sourceTypesByProject[projectDir] ?? []).contains(where: { $0.origin == CatalogOrigin.user && $0.key == key }) {
                throw CoreInvokeError.coded(
                    status: 1,
                    code: "sourcetypes.duplicate_key",
                    kind: .conflict,
                    params: [key]
                )
            }
            let resolvedIcon = iconKey.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty
                ? PVMarkKey.defaultTypeMark.rawValue
                : iconKey
            let type = CatalogSourceType(
                id: UUID().uuidString.lowercased(),
                key: key,
                origin: "user",
                label: label,
                description: description,
                iconKey: resolvedIcon
            )
            sourceTypesByProject[projectDir, default: []].append(type)
            return type
        }
    }

    func updateSourceType(
        projectDir: String,
        userID _: String,
        typeID: String,
        label: String,
        description: String,
        iconKey: String
    ) async throws -> CatalogSourceType {
        return try withState {
            var list = sourceTypesByProject[projectDir] ?? []
            guard let idx = list.firstIndex(where: { $0.id == typeID }) else {
                throw StoreBoom.boom
            }
            guard list[idx].origin == CatalogOrigin.user || list[idx].origin == CatalogOrigin.provenencia else {
                throw StoreBoom.boom
            }
            list[idx].label = label
            list[idx].description = description
            let trimmed = iconKey.trimmingCharacters(in: .whitespacesAndNewlines)
            list[idx].iconKey = trimmed.isEmpty ? PVMarkKey.defaultTypeMark.rawValue : trimmed
            sourceTypesByProject[projectDir] = list
            return withSuggestedFieldCount(list[idx])
        }
    }

    func deleteSourceType(
        projectDir: String,
        userID _: String,
        typeID: String
    ) async throws {
        try withState {
            var list = sourceTypesByProject[projectDir] ?? []
            guard let idx = list.firstIndex(where: { $0.id == typeID }) else {
                throw StoreBoom.boom
            }
            let type = list[idx]
            if CatalogOrigin.isPlugin(type.origin) {
                throw CoreInvokeError.coded(
                    status: 1,
                    code: "sourcetypes.origin_locked",
                    kind: .conflict,
                    params: []
                )
            }
            let inbound = (sourcesByProject[projectDir] ?? []).filter { $0.sourceTypeID == typeID }
            if !inbound.isEmpty || type.usedBy > 0 {
                throw CoreInvokeError.coded(
                    status: 1,
                    code: "sourcetypes.in_use",
                    kind: .conflict,
                    params: []
                )
            }
            list.remove(at: idx)
            sourceTypesByProject[projectDir] = list
            // Suggestion joins cascade; the fields they named do not.
            suggestionsByType[typeID] = nil
        }
    }

    func listTypeSuggestions(projectDir _: String, typeID: String) async throws -> [CatalogTypeSuggestion] {
        return withState {
            suggestionsByType[typeID] ?? []
        }
    }

    func assignTypeField(
        projectDir: String,
        userID _: String,
        typeID: String,
        fieldID: String
    ) async throws -> [CatalogTypeSuggestion] {
        return try withState {
            guard let field = (fieldsByProject[projectDir] ?? []).first(where: { $0.id == fieldID }) else {
                throw StoreBoom.boom
            }
            var list = suggestionsByType[typeID] ?? []
            // Assigning a field the type already suggests leaves its place alone.
            if !list.contains(where: { $0.field.id == fieldID }) {
                list.append(CatalogTypeSuggestion(field: field, sortOrder: (list.last?.sortOrder ?? -1) + 1))
                suggestionsByType[typeID] = list
            }
            return list
        }
    }

    func removeTypeField(
        projectDir _: String,
        userID _: String,
        typeID: String,
        fieldID: String
    ) async throws -> [CatalogTypeSuggestion] {
        return withState {
            let list = (suggestionsByType[typeID] ?? []).filter { $0.field.id != fieldID }
            suggestionsByType[typeID] = list
            return list
        }
    }

    func listMetadataFields(projectDir: String) async throws -> [CatalogMetadataField] {
        return withState {
            markCatalogSessionHeld(projectDir)
            return fieldsByProject[projectDir] ?? []
        }
    }

    func createMetadataField(
        projectDir: String,
        userID _: String,
        label: String,
        dataType: String,
        description: String
    ) async throws -> CatalogMetadataField {
        return try withState {
            let key = FieldSlug.kebab(label)
            if key.isEmpty {
                throw StoreBoom.boom
            }
            if (fieldsByProject[projectDir] ?? []).contains(where: { $0.origin == CatalogOrigin.user && $0.key == key }) {
                throw CoreInvokeError.coded(
                    status: 1,
                    code: "metadatafields.duplicate_key",
                    kind: .conflict,
                    params: [key]
                )
            }
            let field = CatalogMetadataField(
                id: UUID().uuidString.lowercased(),
                key: key,
                origin: "user",
                label: label,
                dataType: dataType,
                description: description
            )
            fieldsByProject[projectDir, default: []].append(field)
            return field
        }
    }

    func updateMetadataField(
        projectDir: String,
        userID _: String,
        fieldID: String,
        label: String,
        dataType: String,
        description: String
    ) async throws -> CatalogMetadataField {
        return try withState {
            var list = fieldsByProject[projectDir] ?? []
            guard let idx = list.firstIndex(where: { $0.id == fieldID }) else {
                throw StoreBoom.boom
            }
            guard list[idx].origin == CatalogOrigin.user || list[idx].origin == CatalogOrigin.provenencia else {
                throw StoreBoom.boom
            }
            guard list[idx].dataType == dataType else {
                throw StoreBoom.boom
            }
            list[idx].label = label
            list[idx].description = description
            fieldsByProject[projectDir] = list
            return list[idx]
        }
    }

    func deleteMetadataField(
        projectDir: String,
        userID _: String,
        fieldID: String
    ) async throws {
        try withState {
            var list = fieldsByProject[projectDir] ?? []
            guard let idx = list.firstIndex(where: { $0.id == fieldID }) else {
                throw StoreBoom.boom
            }
            let field = list[idx]
            if CatalogOrigin.isPlugin(field.origin) {
                throw CoreInvokeError.coded(
                    status: 1,
                    code: "metadatafields.origin_locked",
                    kind: .conflict,
                    params: []
                )
            }
            let inbound = sourcesHoldingField(projectDir: projectDir, fieldID: fieldID)
            if !inbound.isEmpty || field.usedBy > 0 {
                throw CoreInvokeError.coded(
                    status: 1,
                    code: "metadatafields.in_use",
                    kind: .conflict,
                    params: []
                )
            }
            list.remove(at: idx)
            fieldsByProject[projectDir] = list
        }
    }

    func workspaceNavCounts(projectDir: String) async throws -> WorkspaceNavCounts {
        return try withState {
            markCatalogSessionHeld(projectDir)
            if let workspaceNavCountsError { throw workspaceNavCountsError }
            let types = sourceTypesByProject[projectDir] ?? []
            let fields = fieldsByProject[projectDir] ?? []
            return WorkspaceNavCounts(
                sources: (sourcesByProject[projectDir] ?? []).count,
                sourceTypes: Self.originCounts(from: types.map(\.origin)),
                metadataFields: Self.originCounts(from: fields.map(\.origin)),
                persons: Self.handleCount(membershipBySubject, kind: "person"),
                events: Self.handleCount(membershipBySubject, kind: "event"),
                places: Self.handleCount(membershipBySubject, kind: "place")
            )
        }
    }

    func nameFormat(projectDir: String) async throws -> NameFormatOrders {
        withState {
            markCatalogSessionHeld(projectDir)
            return NameFormatOrders()
        }
    }

    /// Handle kinds: Persons by auto-reconciled name or ref, Events and Places by ref or
    /// label (FakeStore has no event / place values). Other kinds: the vocabulary
    /// fake below, filtered to the requested kinds.
    func searchCatalog(
        projectDir: String,
        query: String,
        location: WorkspaceLocation?,
        kinds: [String]
    ) async throws -> [CatalogSearchHit] {
        let handleKinds = Self.primaryTypeKeys
        let wanted = Set(kinds)
        var scored: [(hit: CatalogSearchHit, score: Double)] = []
        if wanted.isEmpty || !wanted.isDisjoint(with: handleKinds) {
            let handleWanted = wanted.isEmpty ? handleKinds : wanted.intersection(handleKinds)
            scored += try withState {
                if let searchCatalogError { throw searchCatalogError }
                markCatalogSessionHeld(projectDir)
                return handleSearchHits(query: query, kinds: handleWanted)
            }
        }
        if wanted.isEmpty || !wanted.isSubset(of: handleKinds) {
            let omnibar = try await vocabularySearchHits(projectDir: projectDir, query: query, location: location)
            scored += wanted.isEmpty ? omnibar : omnibar.filter { wanted.contains($0.hit.kind) }
        }
        return Array(Self.orderedByMixedScore(scored).prefix(50).map(\.hit))
    }

    /// Kind share mirrored from `KindSpec.Priority` and `ScoreMix` (90% text, 10% kind).
    private static let searchPriority: [String: Double] = [
        "person": 4,
        "event": 3,
        "place": 2,
        "source": 1,
    ]

    private static func orderedByMixedScore(
        _ scored: [(hit: CatalogSearchHit, score: Double)]
    ) -> [(hit: CatalogSearchHit, score: Double)] {
        let maxPriority = 4.0
        return scored.enumerated().map { item in
            let fraction = (searchPriority[item.element.hit.kind] ?? 0) / maxPriority
            let mixed = item.element.score * (0.9 + 0.1 * fraction)
            return (offset: item.offset, hit: item.element.hit, score: mixed)
        }.sorted { a, b in
            if a.score != b.score { return a.score > b.score }
            if a.hit.kind != b.hit.kind { return a.hit.kind < b.hit.kind }
            if a.hit.title != b.hit.title { return a.hit.title < b.hit.title }
            return a.offset < b.offset
        }.map { (hit: $0.hit, score: $0.score) }
    }

    /// Call inside `withState`.
    private func handleSearchHits(query: String, kinds: Set<String>) -> [(hit: CatalogSearchHit, score: Double)] {
        let q = TextMatch.Query(query, locale: TextMatch.posix)
        guard !q.isEmpty else { return [] }
        let names = Dictionary(personHeaders().map { ($0.entity.id, $0.name?.form ?? "") }, uniquingKeysWith: { a, _ in a })
        var seen = Set<String>()
        var out: [(hit: CatalogSearchHit, score: Double)] = []
        for membership in membershipBySubject.values.sorted(by: { $0.entity.ref < $1.entity.ref })
        where kinds.contains(membership.kind) && seen.insert(membership.entity.id).inserted {
            let entity = membership.entity
            let header = handleSearchHeader(kind: membership.kind, id: entity.id)
            let name = names[entity.id] ?? ""
            let title = !name.isEmpty ? name : (!entity.label.isEmpty ? entity.label : entity.ref)
            guard q.matches(title, entity.ref, entity.label, handleSearchMatchText(header)) else { continue }
            let section: WorkspaceSection = membership.kind == "person" ? .persons : (membership.kind == "event" ? .events : .places)
            let refMatch = q.matches(entity.ref)
            out.append((
                CatalogSearchHit(
                    kind: membership.kind, id: entity.id, ref: entity.ref, title: title, subtitle: "",
                    matchReason: refMatch ? "ref" : "title",
                    location: WorkspaceLocation(section: section, entityId: entity.id, ref: entity.ref, title: title),
                    memberCount: membershipBySubject.values.filter { $0.entity.id == entity.id }.count,
                    header: header
                ),
                refMatch ? 12 : 10
            ))
        }
        return out
    }

    /// The list header for a handle hit, when the fake has one.
    private func handleSearchHeader(kind: String, id: String) -> CatalogConclusionHeader? {
        switch kind {
        case "person":
            if let header = personHeaders().first(where: { $0.entity.id == id }) {
                return .person(header)
            }
        case "event":
            if let header = seededEventHeaders.first(where: { $0.entity.id == id }) {
                return .event(header)
            }
        case "place":
            if let header = seededPlaceHeaders.first(where: { $0.entity.id == id }) {
                return .place(header)
            }
        default:
            break
        }
        return nil
    }

    /// Match text the lists would show, so a query for an event title or a
    /// place chain finds the handle the way the index does.
    private func handleSearchMatchText(_ header: CatalogConclusionHeader?) -> String {
        switch header {
        case .person(let header):
            return PersonLifeDisplay.line(header).text
        case .event(let header):
            return EventTitleDisplay.title(header) + " " + EventSecondaryDisplay.line(header)
        case .place(let header):
            return (header.names + header.parents).joined(separator: " ")
        case nil:
            return ""
        }
    }

    private func vocabularySearchHits(
        projectDir: String,
        query: String,
        location: WorkspaceLocation?
    ) async throws -> [(hit: CatalogSearchHit, score: Double)] {
        return try withState {
            markCatalogSessionHeld(projectDir)
            if let searchCatalogError { throw searchCatalogError }
            let raw = query.trimmingCharacters(in: .whitespacesAndNewlines)
            let tokens = Self.tokenizeSearch(raw)
            let refKey = Self.classifyRefQuery(raw)
            guard !tokens.isEmpty || refKey != nil else { return [] }

            var scored: [(hit: CatalogSearchHit, score: Double)] = []
            let types = sourceTypesByProject[projectDir] ?? []
            let typeLabelByID = Dictionary(uniqueKeysWithValues: types.map { ($0.id, $0.label) })
            let typeIconByID = Dictionary(uniqueKeysWithValues: types.map { ($0.id, $0.iconKey) })

            for source in sourcesByProject[projectDir] ?? [] {
                var description = source.description
                if let typeLabel = typeLabelByID[source.sourceTypeID], !typeLabel.isEmpty {
                    description = "\(description) \(typeLabel)".trimmingCharacters(in: .whitespaces)
                }
                let values = [
                    "title": source.title,
                    "ref": source.ref,
                    "description": description,
                ]
                var (score, reason) = Self.scoreSearchFields(
                    fields: [("title", 10), ("ref", 12), ("description", 3)],
                    values: values,
                    tokens: tokens.isEmpty ? [raw.lowercased()] : tokens
                )
                var refBoost = 1.0
                if let refKey {
                    let srcRef = source.ref.uppercased()
                    if refKey.exact && srcRef == refKey.key {
                        score = max(score, 12)
                        reason = "ref"
                        refBoost = 8
                    } else if !refKey.exact && srcRef.hasPrefix(refKey.key) {
                        score = max(score, 12)
                        reason = "ref"
                        refBoost = 3
                    } else if score <= 0 {
                        continue
                    }
                } else {
                    guard score > 0 else { continue }
                }
                let boost = location?.section == .sources ? 1.1 : 1.0
                scored.append((
                    CatalogSearchHit(
                        kind: "source",
                        id: source.id,
                        ref: source.ref,
                        title: source.title,
                        subtitle: typeLabelByID[source.sourceTypeID] ?? "",
                        matchReason: reason,
                        location: WorkspaceLocation(
                            section: .sources,
                            sourceId: source.id,
                            ref: source.ref,
                            title: source.title
                        ),
                        thumbnailRelPath: source.thumbnailRelPath,
                        iconKey: typeIconByID[source.sourceTypeID] ?? ""
                    ),
                    score * boost * refBoost
                ))
            }

            // Ref-shaped queries only promote Sources (vocab has no SRC-… refs).
            if refKey == nil {
                for type in types {
                    let (score, reason) = Self.scoreSearchFields(
                        fields: [("label", 10), ("key", 8), ("description", 3)],
                        values: [
                            "label": type.label,
                            "key": type.key,
                            "description": type.description,
                        ],
                        tokens: tokens
                    )
                    guard score > 0 else { continue }
                    let boost = location?.section == .sourceTypes ? 1.1 : 1.0
                    scored.append((
                        CatalogSearchHit(
                            kind: "source_type",
                            id: type.id,
                            ref: "",
                            title: type.label,
                            subtitle: type.key,
                            matchReason: reason,
                            location: WorkspaceLocation(
                                section: .sourceTypes,
                                typeId: type.id,
                                title: type.label
                            ),
                            iconKey: type.iconKey
                        ),
                        score * boost
                    ))
                }

                for field in fieldsByProject[projectDir] ?? [] {
                    let (score, reason) = Self.scoreSearchFields(
                        fields: [("label", 10), ("key", 8), ("description", 3)],
                        values: [
                            "label": field.label,
                            "key": field.key,
                            "description": field.description,
                        ],
                        tokens: tokens
                    )
                    guard score > 0 else { continue }
                    let boost = location?.section == .metadata ? 1.1 : 1.0
                    scored.append((
                        CatalogSearchHit(
                            kind: "metadata_field",
                            id: field.id,
                            ref: "",
                            title: field.label,
                            subtitle: field.key,
                            matchReason: reason,
                            location: WorkspaceLocation(
                                section: .metadata,
                                fieldId: field.id,
                                title: field.label
                            )
                        ),
                        score * boost
                    ))
                }
            }

            return scored
        }
    }

    func listSubjectTypes(projectDir: String) async throws -> [CatalogSubjectType] {
        return withState {
            listSubjectTypesCalls += 1
            markCatalogSessionHeld(projectDir)
            return subjectTypesByProject[projectDir] ?? []
        }
    }

    func createSubject(
        projectDir: String,
        userID _: String,
        graphID: String,
        subjectTypeID: String,
        label: String,
        description: String,
        placement: CatalogGridCell?
    ) async throws -> (CatalogSubject, CatalogWriteEffects) {
        return try withState {
            recordedCalls.append("createSubject typeID=\(subjectTypeID) placement=\(placement.map { "\($0.gridX),\($0.gridY)" } ?? "nil")")
            markCatalogSessionHeld(projectDir)
            guard let sourceID = sourceOfGraph(graphID) else {
                throw CoreInvokeError.coded(status: 1, code: "subjects.invalid", kind: .user, params: [])
            }
            let subject = CatalogSubject(
                id: UUID().uuidString.lowercased(),
                ref: "CPR-FAKE1",
                sourceID: sourceID,
                subjectTypeID: subjectTypeID,
                label: label,
                description: description
            )
            subjectsByGraph[graphID, default: []].append(subject)
            let effects = bumpSource(sourceID, graphs: [graphID])
            if let placement {
                subjectPositionsBySubject[subject.id] = CatalogSubjectPosition(
                    subjectID: subject.id,
                    gridX: placement.gridX,
                    gridY: placement.gridY
                )
            }
            return (subject, effects)
        }
    }

    func updateSubject(
        projectDir: String,
        userID _: String,
        subjectID: String,
        label: String,
        description: String
    ) async throws -> (CatalogSubject, CatalogWriteEffects) {
        return try withState {
            recordedCalls.append("updateSubject id=\(subjectID) label=\(label)")
            markCatalogSessionHeld(projectDir)
            for (graphID, var list) in subjectsByGraph {
                guard let idx = list.firstIndex(where: { $0.id == subjectID }) else { continue }
                list[idx].label = label
                list[idx].description = description
                subjectsByGraph[graphID] = list
                return (list[idx], bumpSource(sourceOfGraph(graphID), graphs: [graphID]))
            }
            throw StoreBoom.boom
        }
    }

    func promoteSubject(
        projectDir: String,
        userID _: String,
        subjectID: String,
        entityID: String?,
        confidenceGradeID: String?,
        argument: String
    ) async throws -> CatalogPromoteResult {
        let delay = withState { promoteSubjectDelayNanoseconds }
        if delay > 0 {
            try? await Task.sleep(nanoseconds: delay)
        }
        return try withState {
            markCatalogSessionHeld(projectDir)
            recordedCalls.append("promoteSubject id=\(subjectID)" + (entityID.map { " entity=\($0)" } ?? ""))
            guard let subject = subjectsByGraph.values.flatMap({ $0 }).first(where: { $0.id == subjectID }),
                  let type = subjectTypesByProject[projectDir]?.first(where: { $0.id == subject.subjectTypeID })
            else {
                throw CoreInvokeError.coded(status: 1, code: "promote.invalid", kind: .user, params: [])
            }
            guard Self.primaryTypeKeys.contains(type.key) else {
                throw CoreInvokeError.coded(status: 1, code: "promote.unsupported_type", kind: .user, params: [])
            }
            if membershipBySubject[subjectID] != nil {
                throw CoreInvokeError.coded(status: 1, code: "identityclaims.already_member", kind: .conflict, params: [])
            }
            let entity: CatalogCanonicalEntity
            if let entityID {
                guard let existing = membershipBySubject.values.first(where: { $0.entity.id == entityID })?.entity else {
                    throw CoreInvokeError.coded(status: 1, code: "promote.invalid", kind: .user, params: [])
                }
                guard existing.subjectTypeID == type.id else {
                    throw CoreInvokeError.coded(status: 1, code: "identityclaims.type_mismatch", kind: .user, params: [])
                }
                entity = existing
            } else {
                entity = CatalogCanonicalEntity(
                    id: UUID().uuidString.lowercased(),
                    ref: "\(type.refPrefix)-FAKE\(membershipBySubject.count + 1)",
                    subjectTypeID: type.id,
                    label: ""
                )
            }
            let claim = CatalogIdentityClaim(
                id: UUID().uuidString.lowercased(),
                subjectID: subjectID,
                entityID: entity.id,
                status: "accepted",
                confidenceGradeID: confidenceGradeID,
                argument: argument.trimmingCharacters(in: .whitespacesAndNewlines)
            )
            membershipBySubject[subjectID] = CatalogSubjectMembership(
                subjectID: subjectID,
                claimID: claim.id,
                entity: entity,
                kind: type.key
            )
            claimBySubject[subjectID] = claim
            return CatalogPromoteResult(entity: entity, claim: claim)
        }
    }

    /// Removes the Subject's filing, and the filing of every Subject on its Source that has a membership
    /// and an Observation naming it as its value (the engine's rule for bridge ends). Handles stay.
    func unpromoteSubject(
        projectDir: String,
        userID _: String,
        subjectID: String
    ) async throws -> CatalogWriteEffects {
        try withState {
            markCatalogSessionHeld(projectDir)
            recordedCalls.append("unpromoteSubject id=\(subjectID)")
            if let unpromoteSubjectError { throw unpromoteSubjectError }
            guard let graphID = graphOfSubject(subjectID) else {
                throw CoreInvokeError.coded(status: 1, code: "promote.invalid", kind: .user, params: [])
            }
            guard membershipBySubject[subjectID] != nil else { return .empty }
            let ends = (observationsByGraph[graphID] ?? [])
                .filter { $0.valueSubjectID == subjectID }
                .map(\.subjectID)
            var removed: [String] = []
            for id in [subjectID] + ends {
                guard let membership = membershipBySubject[id] else { continue }
                removed.append(membership.entity.id)
                membershipBySubject[id] = nil
                claimBySubject[id] = nil
            }
            return CatalogWriteEffects(handles: removed, evidenceGraphs: [graphID], handleScopes: removed)
        }
    }

    /// Records the call and reports effects only; Go owns the lead rule, so seeded details do not change.
    func setPreferredLikeness(
        projectDir: String,
        userID _: String,
        entityID: String,
        observationID: String
    ) async throws -> CatalogWriteEffects {
        try withState {
            markCatalogSessionHeld(projectDir)
            recordedCalls.append("setPreferredLikeness entity=\(entityID) observation=\(observationID)")
            if let setPreferredLikenessError { throw setPreferredLikenessError }
            return CatalogWriteEffects(handles: [entityID], handleScopes: [entityID])
        }
    }

    /// Stub graph-alignment proposal: every unpromoted primary Subject is Skip;
    /// already-promoted Subjects are Target handle with their membership.
    func proposePromoteGraphAlignment(
        projectDir: String,
        graphID: String,
        fixed: [CatalogPromoteGraphAlignmentFixed]
    ) async throws -> CatalogPromoteGraphAlignmentProposal {
        // Delay and scripted rows are one claim. Concurrent calls must not
        // pair this call's wait with the next call's proposal.
        let claimed: (delay: Duration?, proposal: CatalogPromoteGraphAlignmentProposal?) = withState {
            markCatalogSessionHeld(projectDir)
            lastPromoteFixed = fixed
            recordedCalls.append("proposePromoteGraphAlignment graph=\(graphID)")
            let delay: Duration? = promoteProposeDelays.isEmpty ? nil : promoteProposeDelays.removeFirst()
            guard !promoteProposals.isEmpty else { return (delay, nil) }
            let next = promoteProposals.count > 1 ? promoteProposals.removeFirst() : promoteProposals[0]
            let proposal = CatalogPromoteGraphAlignmentProposal(revision: nextAuditRevision - 1, rows: next.rows)
            return (delay, proposal)
        }
        if let delay = claimed.delay { try? await Task.sleep(for: delay) }
        if let proposal = claimed.proposal { return proposal }
        return buildPromoteProposal(projectDir: projectDir, graphID: graphID, fixed: fixed)
    }

    private func buildPromoteProposal(
        projectDir: String,
        graphID: String,
        fixed: [CatalogPromoteGraphAlignmentFixed]
    ) -> CatalogPromoteGraphAlignmentProposal {
        withState {
            markCatalogSessionHeld(projectDir)
            if !promoteProposals.isEmpty {
                let next = promoteProposals.count > 1 ? promoteProposals.removeFirst() : promoteProposals[0]
                return CatalogPromoteGraphAlignmentProposal(revision: nextAuditRevision - 1, rows: next.rows)
            }
            let fixedBySubject = Dictionary(uniqueKeysWithValues: fixed.map { ($0.subjectID, $0.handleID) })
            let subjects = subjectsByGraph[graphID] ?? []
            let types = Dictionary(uniqueKeysWithValues: (subjectTypesByProject[projectDir] ?? []).map { ($0.id, $0) })
            let rows = subjects.compactMap { subject -> CatalogPromoteGraphAlignmentRow? in
                guard let type = types[subject.subjectTypeID],
                      Self.primaryTypeKeys.contains(type.key)
                else { return nil }
                if let membership = membershipBySubject[subject.id] {
                    return CatalogPromoteGraphAlignmentRow(
                        subjectID: subject.id,
                        kind: type.key,
                        target: "handle",
                        handleID: membership.entity.id,
                        handleRef: membership.entity.ref,
                        score: 0,
                        assessment: "strong",
                        reason: .decided,
                        alternatives: [],
                        conflictWithFixed: false,
                        possibleDuplicate: false
                    )
                }
                if let handleID = fixedBySubject[subject.id],
                   let entity = membershipBySubject.values.first(where: { $0.entity.id == handleID })?.entity
                {
                    return CatalogPromoteGraphAlignmentRow(
                        subjectID: subject.id,
                        kind: type.key,
                        target: "handle",
                        handleID: entity.id,
                        handleRef: entity.ref,
                        score: 0,
                        assessment: "strong",
                        reason: .decided,
                        alternatives: [],
                        conflictWithFixed: false,
                        possibleDuplicate: false
                    )
                }
                return CatalogPromoteGraphAlignmentRow(
                    subjectID: subject.id,
                    kind: type.key,
                    target: "skip",
                    handleID: "",
                    handleRef: "",
                    score: 0,
                    assessment: "none",
                    reason: .empty,
                    alternatives: [],
                    conflictWithFixed: false,
                    possibleDuplicate: false
                )
            }
            return CatalogPromoteGraphAlignmentProposal(revision: nextAuditRevision - 1, rows: rows)
        }
    }

    func getPromoteClaimEvidence(
        projectDir: String,
        graphID: String,
        subjectID: String,
        handleID: String,
        neighbors: [CatalogPromoteGraphAlignmentFixed]
    ) async throws -> CatalogClaimEvidence {
        try withState {
            markCatalogSessionHeld(projectDir)
            recordedCalls.append("getPromoteClaimEvidence graph=\(graphID) subject=\(subjectID) handle=\(handleID) neighbors=\(neighbors.count)")
            if let promoteEvidenceError { throw promoteEvidenceError }
            return promoteEvidence[subjectID] ?? CatalogClaimEvidence()
        }
    }

    /// Files claims. Which Observations a pin names is Go's (`promotealign.PinOffers`); the fake pins none.
    func applyPromoteGraphAlignment(
        projectDir: String,
        userID _: String,
        graphID: String,
        seenRevision: Int64,
        rows: [CatalogPromoteBatchRow]
    ) async throws -> CatalogPromoteBatchResult {
        try withState {
            markCatalogSessionHeld(projectDir)
            recordedCalls.append("applyPromoteGraphAlignment graph=\(graphID) rows=\(rows.count)")
            lastPromoteBatch = rows
            let current = nextAuditRevision - 1
            guard seenRevision == current else {
                throw CoreInvokeError.coded(status: 1, code: "promote.stale", kind: .conflict, params: [])
            }
            var written: [CatalogPromoteBatchWritten] = []
            for row in rows {
                // Go's applyRow: a Skip writes nothing and carries nothing; a New has no pins.
                switch row.target {
                case "skip":
                    guard row.entityID == nil, row.pinIDs.isEmpty, row.confidenceGradeID == nil, row.argument.isEmpty else {
                        throw CoreInvokeError.coded(status: 1, code: "promote.invalid", kind: .user, params: [])
                    }
                    continue
                case "new":
                    guard row.entityID == nil, row.pinIDs.isEmpty else {
                        throw CoreInvokeError.coded(status: 1, code: "promote.invalid", kind: .user, params: [])
                    }
                case "handle":
                    break
                default:
                    throw CoreInvokeError.coded(status: 1, code: "promote.invalid", kind: .user, params: [])
                }
                guard let subject = (subjectsByGraph[graphID] ?? []).first(where: { $0.id == row.subjectID }),
                      let type = subjectTypesByProject[projectDir]?.first(where: { $0.id == subject.subjectTypeID }),
                      Self.primaryTypeKeys.contains(type.key)
                else {
                    throw CoreInvokeError.coded(status: 1, code: "promote.invalid", kind: .user, params: [])
                }
                if let existing = membershipBySubject[subject.id] {
                    if row.target == "handle", existing.entity.id == row.entityID, row.pinIDs.isEmpty {
                        continue
                    }
                    throw CoreInvokeError.coded(status: 1, code: "identityclaims.already_member", kind: .conflict, params: [])
                }
                let entity: CatalogCanonicalEntity
                if row.target == "handle" {
                    guard let entityID = row.entityID,
                          let existing = membershipBySubject.values.first(where: { $0.entity.id == entityID })?.entity,
                          existing.subjectTypeID == type.id
                    else {
                        throw CoreInvokeError.coded(status: 1, code: "promote.invalid", kind: .user, params: [])
                    }
                    entity = existing
                } else {
                    entity = CatalogCanonicalEntity(
                        id: UUID().uuidString.lowercased(),
                        ref: "\(type.refPrefix)-FAKE\(membershipBySubject.count + 1)",
                        subjectTypeID: type.id,
                        label: ""
                    )
                }
                let claim = CatalogIdentityClaim(
                    id: UUID().uuidString.lowercased(),
                    subjectID: subject.id,
                    entityID: entity.id,
                    status: "accepted",
                    confidenceGradeID: row.confidenceGradeID,
                    argument: row.argument.trimmingCharacters(in: .whitespacesAndNewlines)
                )
                membershipBySubject[subject.id] = CatalogSubjectMembership(
                    subjectID: subject.id,
                    claimID: claim.id,
                    entity: entity,
                    kind: type.key
                )
                claimBySubject[subject.id] = claim
                written.append(CatalogPromoteBatchWritten(entity: entity, claim: claim, pins: row.pinIDs.count))
            }
            let revision: Int64
            if written.isEmpty {
                revision = seenRevision
            } else {
                revision = nextAuditRevision
                nextAuditRevision += 1
            }
            return CatalogPromoteBatchResult(revision: revision, written: written)
        }
    }

    func listClaimConfidenceGrades(projectDir: String) async throws -> [CatalogClaimConfidenceGrade] {
        return withState {
            markCatalogSessionHeld(projectDir)
            return [
                CatalogClaimConfidenceGrade(id: "cg-low", key: "low_confidence", origin: CatalogOrigin.provenencia, label: "Low confidence", sortOrder: 1),
                CatalogClaimConfidenceGrade(id: "cg-mod", key: "moderate", origin: CatalogOrigin.provenencia, label: "Moderate", sortOrder: 2),
                CatalogClaimConfidenceGrade(id: "cg-high", key: "high_confidence", origin: CatalogOrigin.provenencia, label: "High confidence", sortOrder: 3),
            ]
        }
    }

    func listEvidenceGraphEventTitles(projectDir: String, graphID: String) async throws -> [String: CatalogEventTitle] {
        withState {
            markCatalogSessionHeld(projectDir)
            return eventTitlesByGraph[graphID] ?? [:]
        }
    }

    func listSubjectMemberships(projectDir: String, graphID: String) async throws -> [CatalogSubjectMembership] {
        // Reads do not append to `recordedCalls`: graph reads run as parallel
        // `async let`s beside position writes that log their calls.
        return withState {
            markCatalogSessionHeld(projectDir)
            let headers = Dictionary(
                personHeaders().map { ($0.entity.id, $0) },
                uniquingKeysWith: { first, _ in first }
            )
            return (subjectsByGraph[graphID] ?? []).compactMap { subject -> CatalogSubjectMembership? in
                guard var membership = membershipBySubject[subject.id] else { return nil }
                membership.name = headers[membership.entity.id]?.name
                return membership
            }
        }
    }

    private static func handleCount(_ memberships: [String: CatalogSubjectMembership], kind: String) -> Int {
        Set(memberships.values.filter { $0.kind == kind }.map(\.entity.id)).count
    }

    func listPersonHeaders(projectDir: String) async throws -> [CatalogPersonHeader] {
        // A read: no `recordedCalls`.
        return withState {
            markCatalogSessionHeld(projectDir)
            return personHeaders()
        }
    }

    func listEventHeaders(projectDir: String) async throws -> [CatalogEventHeader] {
        withState {
            markCatalogSessionHeld(projectDir)
            return seededEventHeaders
        }
    }

    func listPlaceHeaders(projectDir: String) async throws -> [CatalogPlaceHeader] {
        withState {
            markCatalogSessionHeld(projectDir)
            return seededPlaceHeaders
        }
    }

    /// Mirrors the Go composer closely enough for UI tests: members' name
    /// Observations group by parts (`nameField`), the most-supported value
    /// (then the earliest) is rank 1, and named Persons sort by name before
    /// unnamed ones by ref. Call inside `withState`.
    private func personHeaders() -> [CatalogPersonHeader] {
        let members = membershipBySubject.values.filter { $0.kind == "person" }
        let byEntity = Dictionary(grouping: members, by: \.entity.id)
        let headers = byEntity.values.compactMap { group -> CatalogPersonHeader? in
            guard let entity = group.first?.entity else { return nil }
            let field = nameField(memberIDs: Set(group.map(\.subjectID)), sources: [])
            var top: CatalogNameValue?
            if case let .name(name)? = field.values.first?.value { top = name }
            return CatalogPersonHeader(entity: entity, name: top, nameValueCount: field.values.count)
        }
        return headers.sorted { a, b in
            switch (a.name, b.name) {
            case let (x?, y?) where x.form.lowercased() != y.form.lowercased():
                return x.form.lowercased() < y.form.lowercased()
            case (.some, .none): return true
            case (.none, .some): return false
            default: return a.entity.ref.localizedCaseInsensitiveCompare(b.entity.ref) == .orderedAscending
            }
        }
    }

    func getConclusionDetail(projectDir: String, entityID: String) async throws -> CatalogConclusionDetail {
        // A read: no `recordedCalls`.
        try withState {
            markCatalogSessionHeld(projectDir)
            if let seeded = seededConclusionDetails[entityID] {
                return seeded
            }
            guard var detail = conclusionDetail(projectDir: projectDir, entityID: entityID) else {
                throw CoreInvokeError.coded(status: 1, code: "conclusiondetails.not_found", kind: .user, params: [])
            }
            if let person = personHeaders().first(where: { $0.entity.id == entityID }) {
                detail.header = .person(person)
            }
            return detail
        }
    }

    /// A Person's conclusion detail: the name field only. Call inside `withState`.
    private func conclusionDetail(projectDir: String, entityID: String) -> CatalogConclusionDetail? {
        let group = membershipBySubject.values.filter { $0.entity.id == entityID }
        guard let entity = group.first?.entity else { return nil }
        let field = nameField(memberIDs: Set(group.map(\.subjectID)), sources: sourcesByProject[projectDir] ?? [])
        return CatalogConclusionDetail(entity: entity, fields: [field], memberCount: group.count)
    }

    /// A Person's name field, grouped by parts (`nameKey`): one value per
    /// group, supported by distinct Sources; rank 1 displayed (`kept`) and the
    /// rest `outvoted`, each Observation an outcome. A name with no part word
    /// is an outcome with no value (`noEvidence`). Other Properties are left
    /// out. Call inside `withState`.
    private func nameField(memberIDs: Set<String>, sources: [CatalogSource]) -> CatalogConclusionField {
        var values: [CatalogReconciledValue] = []
        var outcomes: [CatalogReconcilerOutcome] = []
        var keys: [String] = []
        var supportingSources: [Set<String>] = []
        var propertyID = ""
        for (graphID, list) in observationsByGraph.sorted(by: { $0.key < $1.key }) {
            let sourceID = sourceOfGraph(graphID) ?? ""
            for o in list where o.propertyKey == "name" && memberIDs.contains(o.subjectID) && !o.nameForm.isEmpty {
                propertyID = o.propertyID
                let name = CatalogNameValue(form: o.nameForm, parts: o.nameParts)
                let key = Self.nameKey(o.nameParts)
                if let key {
                    if let i = keys.firstIndex(of: key) {
                        supportingSources[i].insert(sourceID)
                        values[i].support = supportingSources[i].count
                    } else {
                        keys.append(key)
                        supportingSources.append([sourceID])
                        values.append(CatalogReconciledValue(rank: 0, reason: .noEvidence, support: 1, against: 0, value: .name(name)))
                    }
                }
                outcomes.append(CatalogReconcilerOutcome(
                    observationID: o.id,
                    observationRef: o.ref,
                    reason: .noEvidence,
                    valueRank: key.flatMap { keys.firstIndex(of: $0) },
                    deniedByObservationID: "",
                    recorded: .name(name),
                    subjectID: o.subjectID,
                    subjectRef: subjectsByGraph[graphID]?.first { $0.id == o.subjectID }?.ref ?? "",
                    graphID: graphID,
                    citationID: o.citationID,
                    sourceID: sourceID,
                    sourceTitle: sources.first { $0.id == sourceID }?.title ?? "",
                    credibilityKey: "",
                    transcriptionUncertain: false,
                    claimConfidenceKey: ""
                ))
            }
        }
        // Rank by support, then first seen; outcomes point at the new ranks.
        let order = values.indices.sorted { values[$0].support != values[$1].support ? values[$0].support > values[$1].support : $0 < $1 }
        var rankOf: [Int: Int] = [:]
        var ranked: [CatalogReconciledValue] = []
        for (n, i) in order.enumerated() {
            rankOf[i] = n + 1
            var v = values[i]
            v.rank = n + 1
            v.reason = n == 0 ? .kept : .outvoted
            ranked.append(v)
        }
        // Every Source that voted, for the vote an outvoted record lost.
        let voters = Set(supportingSources.flatMap { $0 }).count
        for i in outcomes.indices {
            // A name with no value keeps `noEvidence` and zero votes.
            guard let rank = outcomes[i].valueRank.flatMap({ rankOf[$0] }) else { continue }
            outcomes[i].valueRank = rank
            outcomes[i].reason = rank == 1 ? .kept : .outvoted
            if rank != 1 {
                outcomes[i].voteSupport = ranked.first?.support ?? 0
                outcomes[i].voteTotal = voters
            }
        }
        let state: ReconciledState
        if ranked.isEmpty {
            state = .empty
        } else if (ranked.first?.support ?? 0) > 1 {
            state = .merged
        } else {
            state = .single
        }
        return CatalogConclusionField(
            propertyID: propertyID,
            propertyKey: SeededPropertyKey.name,
            label: "Name",
            valueType: .name,
            state: state,
            values: ranked,
            outcomes: outcomes.sorted { ($0.valueRank ?? .max, $0.observationID) < ($1.valueRank ?? .max, $1.observationID) }
        )
    }

    /// Go `autoreconcile` `nameModule.split`, simplified: by part type, untyped
    /// words at `Same`, so case and punctuation never matter and accents do
    /// (S11-06 policy); a name with no part word is not a value. No
    /// generation, initials, folds or outvoting.
    private static func nameKey(_ parts: [CatalogNameValuePart]) -> String? {
        var wordsByType: [String: [String]] = [:]
        for part in parts {
            let type = part.type.trimmingCharacters(in: .whitespaces)
            wordsByType[type, default: []].append(contentsOf: TextMatch.words(part.value, at: .same, locale: TextMatch.posix))
        }
        let groups = wordsByType.filter { !$0.value.isEmpty }
            .sorted { $0.key < $1.key }
            .map { "\($0.key)=\($0.value.sorted().joined(separator: " "))" }
        return groups.isEmpty ? nil : groups.joined(separator: "|")
    }

    func deleteSubject(
        projectDir: String, userID _: String, subjectID: String
    ) async throws -> CatalogWriteEffects {
        try withState {
            markCatalogSessionHeld(projectDir)
            recordedCalls.append("deleteSubject id=\(subjectID)")
            let report = subjectDeleteImpact(projectDir: projectDir, id: subjectID)
            if !report.allowed {
                if report.gate == .notFound {
                    throw StoreBoom.boom
                }
                throw CoreInvokeError.coded(
                    status: 1,
                    code: "subjects.in_use",
                    kind: .conflict,
                    params: []
                )
            }
            for (graphID, list) in observationsByGraph {
                let kept = list.filter { observation in
                    guard observation.subjectID == subjectID else { return true }
                    return !isConnectionFacet(observation, projectDir: projectDir)
                }
                for observation in list where !kept.contains(where: { $0.id == observation.id }) {
                    edgeObservationIDs.remove(observation.id)
                }
                observationsByGraph[graphID] = kept
            }
            for (graphID, var list) in subjectsByGraph {
                guard let idx = list.firstIndex(where: { $0.id == subjectID }) else { continue }
                list.remove(at: idx)
                subjectsByGraph[graphID] = list
                let effects = bumpSource(sourceOfGraph(graphID), graphs: [graphID])
                subjectPositionsBySubject[subjectID] = nil
                membershipBySubject[subjectID] = nil
                return effects
            }
            throw StoreBoom.boom
        }
    }

    func listSubjects(projectDir: String, graphID: String) async throws -> [CatalogSubject] {
        return try withState {
            listSubjectsCalls += 1
            markCatalogSessionHeld(projectDir)
            if let listSubjectsError {
                throw listSubjectsError
            }
            return subjectsByGraph[graphID] ?? []
        }
    }

    func setSubjectPosition(
        projectDir: String,
        subjectID: String,
        gridX: Int64,
        gridY: Int64
    ) async throws -> CatalogSubjectPosition {
        let (error, holds): (Error?, Bool) = withState {
            markCatalogSessionHeld(projectDir)
            recordedCalls.append("setSubjectPosition subjectID=\(subjectID) \(gridX),\(gridY)")
            let error: Error?
            if setSubjectPositionCallIndex < setSubjectPositionErrors.count {
                error = setSubjectPositionErrors[setSubjectPositionCallIndex]
            } else {
                error = setSubjectPositionError
            }
            setSubjectPositionCallIndex += 1
            return (error, holdsSetSubjectPosition)
        }
        if holds {
            await withCheckedContinuation { continuation in
                withState {
                    heldSetSubjectPositions.append(continuation)
                }
            }
        }
        if let error { throw error }
        let position = CatalogSubjectPosition(subjectID: subjectID, gridX: gridX, gridY: gridY)
        withState {
            subjectPositionsBySubject[subjectID] = position
        }
        return position
    }

    /// Number of `setSubjectPosition` calls suspended by `holdsSetSubjectPosition`.
    var heldSetSubjectPositionCount: Int {
        withState { heldSetSubjectPositions.count }
    }

    /// Resumes the oldest held `setSubjectPosition` call.
    func releaseSetSubjectPosition() {
        let continuation = withState {
            heldSetSubjectPositions.isEmpty ? nil : heldSetSubjectPositions.removeFirst()
        }
        continuation?.resume()
    }

    func clearSubjectPosition(projectDir: String, subjectID: String) async throws {
        withState {
            markCatalogSessionHeld(projectDir)
            subjectPositionsBySubject[subjectID] = nil
        }
    }

    func listSubjectPositions(projectDir: String, graphID: String) async throws -> [CatalogSubjectPosition] {
        return withState {
            markCatalogSessionHeld(projectDir)
            let subjects = subjectsByGraph[graphID] ?? []
            return subjects.compactMap { subjectPositionsBySubject[$0.id] }
                .sorted {
                    if $0.gridY != $1.gridY { return $0.gridY < $1.gridY }
                    if $0.gridX != $1.gridX { return $0.gridX < $1.gridX }
                    return $0.subjectID < $1.subjectID
                }
        }
    }

    func createProperty(
        projectDir: String,
        userID _: String,
        label: String,
        valueType: String,
        description: String,
        cardinality: String
    ) async throws -> CatalogProperty {
        return try withState {
            markCatalogSessionHeld(projectDir)
            let key = label
                .lowercased()
                .replacingOccurrences(of: " ", with: "-")
                .filter { $0.isLetter || $0.isNumber || $0 == "-" }
            let stored = cardinality.isEmpty ? "single" : cardinality
            guard stored == "single" || stored == "multiple" else {
                throw CoreInvokeError.coded(status: 1, code: "properties.invalid", kind: .user, params: [])
            }
            let property = CatalogProperty(
                id: UUID().uuidString.lowercased(),
                key: key,
                origin: "user",
                label: label,
                description: description,
                valueType: valueType,
                cardinality: stored
            )
            propertiesByProject[projectDir, default: []].append(property)
            return property
        }
    }

    func updateProperty(
        projectDir: String,
        userID _: String,
        propertyID: String,
        label: String,
        valueType _: String,
        description: String,
        cardinality: String
    ) async throws -> CatalogProperty {
        return try withState {
            markCatalogSessionHeld(projectDir)
            guard var list = propertiesByProject[projectDir],
                  let idx = list.firstIndex(where: { $0.id == propertyID })
            else {
                throw CoreInvokeError.coded(status: 1, code: "properties.invalid", kind: .user, params: [])
            }
            if CatalogOrigin.isPlugin(list[idx].origin) {
                throw CoreInvokeError.coded(status: 1, code: "properties.invalid", kind: .user, params: [])
            }
            if list[idx].origin != "user" && cardinality != list[idx].cardinality {
                throw CoreInvokeError.coded(status: 1, code: "properties.invalid", kind: .user, params: [])
            }
            list[idx].label = label
            list[idx].description = description
            if !cardinality.isEmpty {
                list[idx].cardinality = cardinality
            }
            propertiesByProject[projectDir] = list
            return list[idx]
        }
    }

    func deleteProperty(projectDir: String, userID _: String, propertyID: String) async throws {
        try withState {
            markCatalogSessionHeld(projectDir)
            guard let property = (propertiesByProject[projectDir] ?? []).first(where: { $0.id == propertyID }) else {
                throw CoreInvokeError.coded(status: 1, code: "properties.invalid", kind: .user, params: [])
            }
            if CatalogOrigin.isPlugin(property.origin) || property.origin == CatalogOrigin.provenencia {
                throw CoreInvokeError.coded(
                    status: 1,
                    code: "properties.origin_locked",
                    kind: .conflict,
                    params: []
                )
            }
            let observations = observationsHoldingProperty(propertyID: propertyID)
            let terms = propertyTermsByProperty[propertyID] ?? []
            if !observations.isEmpty || !terms.isEmpty || property.usedBy > 0 {
                throw CoreInvokeError.coded(
                    status: 1,
                    code: "properties.in_use",
                    kind: .conflict,
                    params: []
                )
            }
            propertiesByProject[projectDir]?.removeAll { $0.id == propertyID }
            propertyTermsByProperty[propertyID] = nil
            for (typeID, list) in subjectTypePropertiesByType {
                subjectTypePropertiesByType[typeID] = list.filter { $0.property.id != propertyID }
            }
        }
    }

    func listPropertyTerms(projectDir: String, propertyID: String) async throws -> [CatalogPropertyTerm] {
        return withState {
            markCatalogSessionHeld(projectDir)
            return propertyTermsByProperty[propertyID] ?? []
        }
    }

    func createPropertyTerm(
        projectDir: String,
        userID _: String,
        propertyID: String,
        label: String,
        description: String
    ) async throws -> CatalogPropertyTerm {
        return try withState {
            markCatalogSessionHeld(projectDir)
            if let createPropertyTermError {
                throw createPropertyTermError
            }
            let key = label
                .lowercased()
                .replacingOccurrences(of: " ", with: "-")
                .filter { $0.isLetter || $0.isNumber || $0 == "-" }
            if let property = (propertiesByProject[projectDir] ?? []).first(where: { $0.id == propertyID }),
               property.key == "place_relationship_type",
               property.origin == CatalogOrigin.provenencia
            {
                throw CoreInvokeError.coded(
                    status: 1, code: "propertyterms.locked", kind: .user, params: []
                )
            }
            let term = CatalogPropertyTerm(
                id: UUID().uuidString.lowercased(),
                propertyID: propertyID,
                key: key,
                origin: "user",
                label: label,
                description: description
            )
            propertyTermsByProperty[propertyID, default: []].append(term)
            return term
        }
    }

    func assignSubjectTypeProperty(
        projectDir: String,
        userID _: String,
        subjectTypeID: String,
        propertyID: String
    ) async throws {
        withState {
            markCatalogSessionHeld(projectDir)
            let property = (propertiesByProject[projectDir] ?? []).first { $0.id == propertyID }
                ?? CatalogProperty(
                    id: propertyID, key: "prop", origin: "user", label: "Prop",
                    description: "", valueType: "text"
                )
            var list = subjectTypePropertiesByType[subjectTypeID] ?? []
            guard !list.contains(where: { $0.property.id == propertyID }) else { return }
            list.append(CatalogSubjectTypeProperty(property: property, sortOrder: list.count, locked: false))
            subjectTypePropertiesByType[subjectTypeID] = list
        }
    }

    func removeSubjectTypeProperty(
        projectDir: String,
        userID _: String,
        subjectTypeID: String,
        propertyID: String
    ) async throws {
        try withState {
            markCatalogSessionHeld(projectDir)
            if let field = subjectTypePropertiesByType[subjectTypeID]?.first(where: { $0.property.id == propertyID }),
               field.locked
            {
                throw CoreInvokeError.coded(status: 1, code: "subjectvocab.locked", kind: .conflict, params: [])
            }
            subjectTypePropertiesByType[subjectTypeID]?.removeAll { $0.property.id == propertyID }
        }
    }

    private static func syntheticPresentation(typeKey: String) -> CatalogSubjectTypePresentation {
        let role: String
        let sort: Int
        switch typeKey {
        case "person": role = "root"; sort = 0
        case "event": role = "root"; sort = 1
        case "place": role = "root"; sort = 2
        case "relationship": role = "bridge"; sort = 3
        case "participation": role = "bridge"; sort = 4
        case "location": role = "bridge"; sort = 5
        case "place_relationship": role = "bridge"; sort = 6
        case "source": role = "reification"; sort = 7
        default: role = "root"; sort = 99
        }
        let ink: String
        switch typeKey {
        case "event", "participation": ink = "subjectEventInk"
        case "place", "location", "place_relationship": ink = "subjectPlaceInk"
        default: ink = "subjectPersonInk"
        }
        let tint = ink.replacingOccurrences(of: "Ink", with: "Tint")
        let chip = ink.replacingOccurrences(of: "Ink", with: "Chip")
        let line = ink.replacingOccurrences(of: "Ink", with: "Line")
        return CatalogSubjectTypePresentation(
            typeKey: typeKey,
            l10nKey: "subjectType.\(typeKey)",
            iconSymbol: typeKey,
            inkToken: ink,
            tintToken: tint,
            chipToken: chip,
            lineToken: line,
            edgeFromToken: "",
            edgeToToken: "",
            role: role,
            placeable: role == "root",
            paletteSort: sort,
            requiresCitationAtCreate: role == "bridge",
            label: typeKey.replacingOccurrences(of: "_", with: " ").capitalized
        )
    }

    func listConnectRules() async throws -> [CatalogConnectRule] {
        return try withState {
            if let listConnectRulesError {
                throw listConnectRulesError
            }
            return connectRules
        }
    }

    func createCitedBridge(
        projectDir: String,
        userID: String,
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
        observations drafts: [CatalogObservationDraft],
        citationID: String?
    ) async throws -> (CatalogSubject, CatalogCitation, [CatalogObservation], CatalogWriteEffects) {
        // Validate and pick the bridge type under the lock; `createSubject` and
        // the citation write below take it themselves.
        let plan: (rule: CatalogConnectRule, typeID: String, cell: CatalogGridCell, graphID: String, sourceID: String) = try withState {
            recordedCalls.append(
                "createCitedBridge citationID=\(citationID ?? "nil") observations=\(drafts.count)"
            )
            markCatalogSessionHeld(projectDir)
            let all = subjectsByGraph.values.flatMap { $0 }
            let from = all.first(where: { $0.id == fromSubjectID })
            let to = all.first(where: { $0.id == toSubjectID })
            guard let from, let to, from.id != to.id else {
                throw CoreInvokeError.coded(status: 1, code: "connect.invalid", kind: .user, params: [])
            }
            let fromKey = subjectTypesByProject[projectDir]?.first(where: { $0.id == from.subjectTypeID })?.key ?? ""
            let toKey = subjectTypesByProject[projectDir]?.first(where: { $0.id == to.subjectTypeID })?.key ?? ""
            let rule = CatalogConnectRule.match(from: fromKey, to: toKey, in: connectRules)
            if rule.refuse {
                throw CoreInvokeError.coded(status: 1, code: "connect.refused", kind: .user, params: [])
            }
            if !bridgeTypeKey.isEmpty, bridgeTypeKey != rule.bridgeTypeKey {
                throw CoreInvokeError.coded(status: 1, code: "connect.invalid", kind: .user, params: [])
            }
            if drafts.contains(where: { !$0.subjectID.isEmpty }) {
                throw CoreInvokeError.coded(status: 1, code: "connect.invalid", kind: .user, params: [])
            }
            var expectedKeys = Set(rule.edges.map(\.propertyKey))
            if rule.disambiguation != "none", !rule.disambiguation.isEmpty {
                expectedKeys.insert(rule.disambiguation)
            }
            let properties = propertiesByProject[projectDir] ?? []
            let draftKeys = Set(drafts.compactMap { draft in
                properties.first { $0.id == draft.propertyID }?.key
            })
            if draftKeys != expectedKeys || drafts.count != expectedKeys.count {
                throw CoreInvokeError.coded(status: 1, code: "connect.invalid", kind: .user, params: [])
            }
            if let citationID, !citationID.isEmpty {
                let fieldsUsed = !locatorJSON.isEmpty || !transcription.isEmpty
                    || !citationDescription.isEmpty || transcriptionUncertain || !transcriptionNote.isEmpty
                    || !citationNotes.isEmpty || !artifactID.isEmpty
                if fieldsUsed {
                    throw CoreInvokeError.coded(status: 1, code: "connect.invalid", kind: .user, params: [])
                }
            }
            guard let fromPos = subjectPositionsBySubject[from.id],
                  let toPos = subjectPositionsBySubject[to.id]
            else {
                throw CoreInvokeError.coded(status: 1, code: "connect.invalid", kind: .user, params: [])
            }
            guard let type = subjectTypesByProject[projectDir]?.first(where: { $0.key == rule.bridgeTypeKey }) else {
                throw StoreBoom.boom
            }
            let midX = Self.floorDiv(fromPos.gridX + toPos.gridX + 1, 2)
            let midY = Self.floorDiv(fromPos.gridY + toPos.gridY + 1, 2)
            // The bridge lands on its from-end's graph.
            let graphID = graphOfSubject(from.id) ?? ""
            return (rule, type.id, CatalogGridCell(gridX: midX, gridY: midY), graphID, from.sourceID)
        }
        let (subject, _) = try await createSubject(
            projectDir: projectDir,
            userID: userID,
            graphID: plan.graphID,
            subjectTypeID: plan.typeID,
            label: "",
            description: description,
            placement: plan.cell
        )
        var stamped = drafts
        for index in stamped.indices {
            stamped[index].subjectID = subject.id
        }
        let citation: CatalogCitation
        let observations: [CatalogObservation]
        if let citationID, !citationID.isEmpty {
            guard let existing = withState({ citationsByID[citationID] }) else {
                throw CoreInvokeError.coded(status: 1, code: "connect.invalid", kind: .user, params: [])
            }
            citation = existing
            observations = try await addObservationsToCitation(
                projectDir: projectDir,
                userID: userID,
                citationID: citationID,
                observations: stamped,
                allowEdgeRows: true
            )
        } else {
            (citation, observations, _) = try await createCitationWithObservations(
                projectDir: projectDir,
                userID: userID,
                artifactID: artifactID,
                locatorJSON: locatorJSON,
                transcription: transcription,
                description: citationDescription,
                transcriptionUncertain: transcriptionUncertain,
                transcriptionNote: transcriptionNote,
                citationNotes: citationNotes,
                observations: stamped
            )
        }
        withState {
            let catalogProperties = propertiesByProject[projectDir] ?? []
            for obs in observations {
                let propertyKey = catalogProperties.first(where: { $0.id == obs.propertyID })?.key ?? obs.propertyKey
                if plan.rule.edges.contains(where: { $0.propertyKey == propertyKey }) {
                    edgeObservationIDs.insert(obs.id)
                }
            }
        }
        return (subject, citation, observations, CatalogWriteEffects(evidenceGraphs: [plan.graphID], citationSources: [plan.sourceID], lists: [.propertyUsage]))
    }

    private static func floorDiv(_ a: Int64, _ b: Int64) -> Int64 {
        guard b != 0 else { return 0 }
        var q = a / b
        if (a ^ b) < 0 && a % b != 0 {
            q -= 1
        }
        return q
    }

    private func isEdgeLocked(observation: CatalogObservation, projectDir: String) -> Bool {
        if edgeObservationIDs.contains(observation.id) { return true }
        let subject = subjectsByGraph.values.flatMap { $0 }.first(where: { $0.id == observation.subjectID })
        let typeKey = subject.flatMap { subject in
            subjectTypesByProject[projectDir]?.first(where: { $0.id == subject.subjectTypeID })?.key
        } ?? ""
        let propertyKey = observation.propertyKey.isEmpty
            ? (propertiesByProject[projectDir]?.first(where: { $0.id == observation.propertyID })?.key ?? "")
            : observation.propertyKey
        return connectRules.contains { rule in
            !rule.refuse && rule.bridgeTypeKey == typeKey && rule.edges.contains { $0.propertyKey == propertyKey }
        }
    }

    private func edgeLockedError() -> CoreInvokeError {
        CoreInvokeError.coded(status: 1, code: "observations.edge_locked", kind: .conflict, params: [])
    }

    func createCitationWithObservations(
        projectDir: String,
        userID _: String,
        artifactID: String,
        locatorJSON: String,
        transcription: String,
        description: String,
        transcriptionUncertain: Bool,
        transcriptionNote: String,
        citationNotes: [String],
        observations drafts: [CatalogObservationDraft]
    ) async throws -> (CatalogCitation, [CatalogObservation], CatalogWriteEffects) {
        let delay: UInt64 = withState {
            recordedCalls.append("createCitationWithObservations observations=\(drafts.count)")
            return createCitationDelayNanoseconds
        }
        if delay > 0 {
            try await Task.sleep(nanoseconds: delay)
        }
        return try withState {
            markCatalogSessionHeld(projectDir)
            let citation = CatalogCitation(
                id: UUID().uuidString.lowercased(),
                ref: "CIT-FAKE1",
                artifactID: artifactID,
                locatorJSON: locatorJSON,
                transcription: transcription,
                description: description,
                transcriptionUncertain: transcriptionUncertain,
                transcriptionNote: transcriptionNote
            )
            citationsByID[citation.id] = citation
            citationNotesByID[citation.id] = citationNotes
            let created = try appendFakeObservations(
                projectDir: projectDir,
                citationID: citation.id,
                drafts: drafts
            )
            let owner = owningSource(ofCitation: citation.id)
            let graphs = owner.map { graphIDsOfSource($0) } ?? []
            bumpSource(owner, graphs: graphs)
            return (citation, created, CatalogWriteEffects(evidenceGraphs: graphs, citationSources: owner.map { [$0] } ?? []))
        }
    }

    func getCitation(
        projectDir: String,
        graphID: String,
        citationID: String
    ) async throws -> (CatalogCitation, [String], [CatalogObservation]) {
        return try withState {
            markCatalogSessionHeld(projectDir)
            recordedCalls.append("getCitation \(citationID) graphID=\(graphID)")
            guard let citation = citationsByID[citationID] else { throw StoreBoom.boom }
            let notes = citationNotesByID[citationID] ?? []
            let observations = (observationsByGraph[graphID] ?? [])
                .filter { $0.citationID == citationID }
            return (citation, notes, observations)
        }
    }

    func addObservationsToCitation(
        projectDir: String,
        userID: String,
        citationID: String,
        observations drafts: [CatalogObservationDraft]
    ) async throws -> ([CatalogObservation], CatalogWriteEffects) {
        let written = try await addObservationsToCitation(
            projectDir: projectDir,
            userID: userID,
            citationID: citationID,
            observations: drafts,
            allowEdgeRows: false
        )
        return (written, {
            withState {
                let owner = owningSource(ofCitation: citationID).map { [$0] } ?? []
                // The Observations land on their Subjects' graphs.
                let graphs = written.compactMap { graphOfSubject($0.subjectID) }
                return CatalogWriteEffects(
                    evidenceGraphs: Self.unique(graphs),
                    citationSources: owner,
                    lists: [.propertyUsage]
                )
            }
        }())
    }

    private func addObservationsToCitation(
        projectDir: String,
        userID _: String,
        citationID: String,
        observations drafts: [CatalogObservationDraft],
        allowEdgeRows: Bool
    ) async throws -> [CatalogObservation] {
        return try withState {
            recordedCalls.append(
                "addObservationsToCitation citationID=\(citationID) observations=\(drafts.count)"
                    + drafts.compactMap { depictionRecord($0.depiction) }.joined()
            )
            markCatalogSessionHeld(projectDir)
            if let addObservationsToCitationError { throw addObservationsToCitationError }
            let written = try appendFakeObservations(
                projectDir: projectDir,
                citationID: citationID,
                drafts: drafts
            )
            if !allowEdgeRows {
                for obs in written where isEdgeLocked(observation: obs, projectDir: projectDir) {
                    throw edgeLockedError()
                }
            }
            let owner = owningSource(ofCitation: citationID)
            bumpSource(owner, graphs: owner.map { graphIDsOfSource($0) } ?? [])
            return written
        }
    }

    func updateCitation(
        projectDir: String,
        userID _: String,
        citationID: String,
        locatorJSON: String,
        transcription: String,
        description: String,
        transcriptionUncertain: Bool,
        transcriptionNote: String
    ) async throws -> (CatalogCitation, CatalogWriteEffects) {
        return try withState {
            recordedCalls.append("updateCitation citationID=\(citationID)")
            markCatalogSessionHeld(projectDir)
            guard var citation = citationsByID[citationID] else { throw StoreBoom.boom }
            citation.locatorJSON = locatorJSON
            citation.transcription = transcription
            citation.description = description
            citation.transcriptionUncertain = transcriptionUncertain
            citation.transcriptionNote = transcriptionNote
            citationsByID[citationID] = citation
            let owner = owningSource(ofCitation: citationID)
            let graphs = owner.map { graphIDsOfSource($0) } ?? []
            bumpSource(owner, graphs: graphs)
            return (citation, CatalogWriteEffects(evidenceGraphs: graphs, citationSources: owner.map { [$0] } ?? []))
        }
    }

    func updateObservation(
        projectDir: String,
        userID _: String,
        observation: CatalogObservation,
        depictionInput: CatalogDepictionValueInput?
    ) async throws -> (CatalogObservation, CatalogWriteEffects) {
        return try withState {
            recordedCalls.append(
                "updateObservation id=\(observation.id)"
                    + (observation.valueDepictionID.isEmpty ? "" : " valueDepictionID=\(observation.valueDepictionID)")
                    + (depictionRecord(depictionInput) ?? "")
            )
            markCatalogSessionHeld(projectDir)
            if let updateObservationError { throw updateObservationError }
            let existing = observationsByGraph.values.flatMap { $0 }.first(where: { $0.id == observation.id })
            guard let existing else { throw StoreBoom.boom }
            if isEdgeLocked(observation: existing, projectDir: projectDir)
                || isEdgeLocked(observation: observation, projectDir: projectDir)
            {
                throw edgeLockedError()
            }
            var next = try fakeObservation(
                citationID: existing.citationID,
                id: existing.id,
                ref: existing.ref,
                row: observation
            )
            if let depictionInput {
                let stored = nextFakeDepiction(depictionInput)
                next.valueDepictionID = stored.0
                next.depiction = stored.1
            } else {
                next.valueDepictionID = existing.valueDepictionID
                next.depiction = existing.depiction
            }
            var owners: [String] = []
            var graphs: [String] = []
            for (graphID, list) in observationsByGraph where list.contains(where: { $0.id == next.id }) {
                observationsByGraph[graphID] = list.map { $0.id == next.id ? next : $0 }
                bumpSource(sourceOfGraph(graphID), graphs: [graphID])
                graphs.append(graphID)
                owners.append(contentsOf: sourceOfGraph(graphID).map { [$0] } ?? [])
            }
            return (next, CatalogWriteEffects(evidenceGraphs: graphs, citationSources: Self.unique(owners)))
        }
    }

    func deleteObservation(projectDir: String, userID _: String, observationID: String) async throws -> CatalogWriteEffects {
        return try withState {
            recordedCalls.append("deleteObservation id=\(observationID)")
            markCatalogSessionHeld(projectDir)
            let existing = observationsByGraph.values.flatMap { $0 }.first(where: { $0.id == observationID })
            guard let existing else { throw StoreBoom.boom }
            if isEdgeLocked(observation: existing, projectDir: projectDir) {
                throw edgeLockedError()
            }
            var owners: [String] = []
            var graphs: [String] = []
            for (graphID, list) in observationsByGraph where list.contains(where: { $0.id == observationID }) {
                graphs.append(graphID)
                owners.append(contentsOf: sourceOfGraph(graphID).map { [$0] } ?? [])
                observationsByGraph[graphID] = list.filter { $0.id != observationID }
                bumpSource(sourceOfGraph(graphID), graphs: [graphID])
            }
            pinsByObservation[observationID] = nil
            return CatalogWriteEffects(evidenceGraphs: graphs, citationSources: Self.unique(owners), lists: [.propertyUsage])
        }
    }

    func deleteCitation(projectDir: String, userID _: String, citationID: String) async throws -> CatalogWriteEffects {
        return try withState {
            recordedCalls.append("deleteCitation id=\(citationID)")
            markCatalogSessionHeld(projectDir)
            let report = citationDeleteImpact(projectDir: projectDir, id: citationID)
            if !report.allowed {
                if report.gate == .notFound {
                    throw StoreBoom.boom
                }
                throw CoreInvokeError.coded(
                    status: 1,
                    code: "citations.in_use",
                    kind: .conflict,
                    params: []
                )
            }
            let owner = owningSource(ofCitation: citationID)
            citationsByID.removeValue(forKey: citationID)
            citationNotesByID.removeValue(forKey: citationID)
            let graphs = owner.map { graphIDsOfSource($0) } ?? []
            bumpSource(owner, graphs: graphs)
            return CatalogWriteEffects(evidenceGraphs: graphs, citationSources: owner.map { [$0] } ?? [])
        }
    }

    func getPropertiesWorkspace(projectDir: String) async throws -> PropertiesSnapshot {
        return withState {
            markCatalogSessionHeld(projectDir)
            let types = subjectTypesByProject[projectDir] ?? []
            var propertiesByTypeID: [String: [CatalogSubjectTypeProperty]] = [:]
            var presentationsByKey: [String: CatalogSubjectTypePresentation] = [:]
            for type in types {
                propertiesByTypeID[type.id] = subjectTypePropertiesByType[type.id] ?? []
                presentationsByKey[type.key] = subjectTypePresentations[type.key]
                    ?? Self.syntheticPresentation(typeKey: type.key)
            }
            return PropertiesSnapshot(
                properties: propertiesByProject[projectDir] ?? [],
                types: types,
                propertiesByTypeID: propertiesByTypeID,
                presentationsByKey: presentationsByKey
            )
        }
    }

    func listObservationsByEvidenceGraph(projectDir: String, graphID: String) async throws -> [CatalogObservation] {
        return withState {
            markCatalogSessionHeld(projectDir)
            return observationsByGraph[graphID] ?? []
        }
    }

    func citationCountsByEvidenceGraph(projectDir: String, graphID: String) async throws -> [String: Int] {
        return withState {
            markCatalogSessionHeld(projectDir)
            guard let graph = evidenceGraphsByProject[projectDir]?.first(where: { $0.id == graphID }) else {
                return [:]
            }
            let artifactIDs = Set(graph.artifactIDs)
            var counts: [String: Int] = [:]
            for citation in citationsByID.values
            where artifactIDs.contains(citation.artifactID) && citationIsInGraph(citation.id, graphID: graphID) {
                counts[citation.artifactID, default: 0] += 1
            }
            return counts
        }
    }

    func listEvidenceGraphProgress(projectDir: String) async throws -> [EvidenceGraphProgress] {
        return try withState {
            markCatalogSessionHeld(projectDir)
            listEvidenceGraphProgressCalls += 1
            if let listEvidenceGraphProgressError { throw listEvidenceGraphProgressError }
            let ids = Set(
                (evidenceGraphsByProject[projectDir] ?? []).map(\.id) + subjectsByGraph.keys + graphProgressByGraph.keys
            )
            return ids.compactMap { id in
                let row = graphProgress(for: id, projectDir: projectDir)
                return row.isZero ? nil : row
            }
        }
    }

    func getEvidenceGraphProgress(projectDir: String, graphID: String) async throws -> EvidenceGraphProgress {
        return try withState {
            markCatalogSessionHeld(projectDir)
            getEvidenceGraphProgressCalls += 1
            if let getEvidenceGraphProgressError { throw getEvidenceGraphProgressError }
            return graphProgress(for: graphID, projectDir: projectDir)
        }
    }

    func getDeleteImpact(projectDir: String, kind: String, id: String) async throws -> CatalogDeleteImpact {
        return withState {
            markCatalogSessionHeld(projectDir)
            if let override = deleteImpactByID[id] {
                return override
            }
            if kind == "source" {
                let arts = artifactsBySource[id] ?? []
                let subs = subjectsOfSource(id)
                var groups: [CatalogDeleteImpactGroup] = []
                if !arts.isEmpty {
                    groups.append(
                        CatalogDeleteImpactGroup(
                            via: "artifacts.source_id",
                            kind: "artifact",
                            total: arts.count,
                            listed: arts.prefix(20).map {
                                CatalogDeleteImpactListed(
                                    id: $0.id,
                                    ref: $0.ref,
                                    title: $0.label,
                                    location: WorkspaceLocation(
                                        section: .sources,
                                        sourceId: id,
                                        artifactId: $0.id,
                                        sourceSurface: .page,
                                        ref: $0.ref,
                                        title: $0.label
                                    )
                                )
                            }
                        )
                    )
                }
                if !subs.isEmpty {
                    groups.append(
                        CatalogDeleteImpactGroup(
                            via: "subjects.source_id",
                            kind: "subject",
                            total: subs.count,
                            listed: subs.prefix(20).map { subject in
                                let location = WorkspaceLocation.evidenceGraph(
                                    sourceId: id,
                                    graphId: graphOfSubject(subject.id) ?? "",
                                    subjectId: subject.id,
                                    ref: subject.ref,
                                    title: subject.label
                                )
                                return CatalogDeleteImpactListed(
                                    id: subject.id,
                                    ref: subject.ref,
                                    title: subject.label,
                                    location: location
                                )
                            }
                        )
                    )
                }
                if !groups.isEmpty {
                    return CatalogDeleteImpact(allowed: false, gate: .inbound, groups: groups)
                }
            }
            if kind == "artifact" {
                let cites = citationsByID.values.filter { $0.artifactID == id }
                if !cites.isEmpty {
                    let sourceID = artifactsBySource.first { _, arts in
                        arts.contains { $0.id == id }
                    }?.key
                    return CatalogDeleteImpact(
                        allowed: false,
                        gate: .inbound,
                        groups: [
                            CatalogDeleteImpactGroup(
                                via: "citations.artifact_id",
                                kind: "citation",
                                total: cites.count,
                                listed: cites.prefix(20).map {
                                    CatalogDeleteImpactListed(
                                        id: $0.id,
                                        ref: $0.ref,
                                        title: $0.transcription,
                                        location: WorkspaceLocation.citationComposer(
                                            sourceId: sourceID ?? "",
                                            graphId: firstGraphID(ofArtifact: $0.artifactID) ?? "",
                                            citationId: $0.id,
                                            artifactId: $0.artifactID,
                                            ref: $0.ref,
                                            title: $0.ref
                                        )
                                    )
                                }
                            ),
                        ]
                    )
                }
            }
            if kind == "subject" {
                return subjectDeleteImpact(projectDir: projectDir, id: id)
            }
            if kind == "citation" {
                return citationDeleteImpact(projectDir: projectDir, id: id)
            }
            if kind == "observation" {
                return observationDeleteImpact(projectDir: projectDir, id: id)
            }
            if kind == "source_type" {
                return sourceTypeDeleteImpact(projectDir: projectDir, id: id)
            }
            if kind == "metadata_field" {
                return metadataFieldDeleteImpact(projectDir: projectDir, id: id)
            }
            if kind == "property" {
                return propertyDeleteImpact(projectDir: projectDir, id: id)
            }
            if fixtureContains(projectDir: projectDir, id: id) {
                return CatalogDeleteImpact(allowed: true, gate: .ok, groups: [])
            }
            return CatalogDeleteImpact(allowed: false, gate: .notFound, groups: [])
        }
    }

    private func observationsHoldingProperty(propertyID: String) -> [CatalogObservation] {
        observationsByGraph.values.flatMap { $0 }.filter { $0.propertyID == propertyID }
    }

    private func propertyDeleteImpact(projectDir: String, id: String) -> CatalogDeleteImpact {
        guard let property = (propertiesByProject[projectDir] ?? []).first(where: { $0.id == id }) else {
            return CatalogDeleteImpact(allowed: false, gate: .notFound, groups: [])
        }
        if CatalogOrigin.isPlugin(property.origin) || property.origin == CatalogOrigin.provenencia {
            return CatalogDeleteImpact(allowed: false, gate: .originLocked, groups: [])
        }
        let observations = observationsHoldingProperty(propertyID: id)
        let terms = propertyTermsByProperty[id] ?? []
        let observationTotal = observations.isEmpty ? property.usedBy : observations.count
        var groups: [CatalogDeleteImpactGroup] = []
        if observationTotal > 0 {
            if observations.isEmpty {
                groups.append(
                    CatalogDeleteImpactGroup(
                        via: "observations.property_id",
                        kind: "observation",
                        total: observationTotal,
                        listed: []
                    )
                )
            } else {
                groups.append(
                    observationImpactGroup(via: "observations.property_id", observations: observations)
                )
            }
        }
        if !terms.isEmpty {
            groups.append(
                CatalogDeleteImpactGroup(
                    via: "property_terms.property_id",
                    kind: "property_term",
                    total: terms.count,
                    listed: terms.prefix(20).map {
                        CatalogDeleteImpactListed(
                            id: $0.id,
                            ref: $0.key,
                            title: $0.label.isEmpty ? $0.key : $0.label,
                            location: WorkspaceLocation(
                                section: .properties,
                                propertyId: id,
                                ref: $0.key,
                                title: $0.label.isEmpty ? $0.key : $0.label
                            )
                        )
                    }
                )
            )
        }
        if groups.isEmpty {
            return CatalogDeleteImpact(allowed: true, gate: .ok, groups: [])
        }
        return CatalogDeleteImpact(allowed: false, gate: .inbound, groups: groups)
    }

    private func sourcesHoldingField(projectDir: String, fieldID: String) -> [CatalogSource] {
        (sourcesByProject[projectDir] ?? []).filter { source in
            (metadataBySource[source.id] ?? []).contains { $0.field.id == fieldID && $0.hasValue }
        }
    }

    private func metadataFieldDeleteImpact(projectDir: String, id: String) -> CatalogDeleteImpact {
        guard let field = (fieldsByProject[projectDir] ?? []).first(where: { $0.id == id }) else {
            return CatalogDeleteImpact(allowed: false, gate: .notFound, groups: [])
        }
        if CatalogOrigin.isPlugin(field.origin) {
            return CatalogDeleteImpact(allowed: false, gate: .originLocked, groups: [])
        }
        let inbound = sourcesHoldingField(projectDir: projectDir, fieldID: id)
        let total = inbound.isEmpty ? field.usedBy : inbound.count
        if total == 0 {
            return CatalogDeleteImpact(allowed: true, gate: .ok, groups: [])
        }
        return CatalogDeleteImpact(
            allowed: false,
            gate: .inbound,
            groups: [
                CatalogDeleteImpactGroup(
                    via: "source_metadata.field_id",
                    kind: "source",
                    total: total,
                    listed: inbound.prefix(20).map {
                        CatalogDeleteImpactListed(
                            id: $0.id,
                            ref: $0.ref,
                            title: $0.title.isEmpty ? $0.ref : $0.title,
                            location: WorkspaceLocation.sourcePage(
                                sourceId: $0.id,
                                ref: $0.ref,
                                title: $0.title.isEmpty ? $0.ref : $0.title
                            )
                        )
                    }
                ),
            ]
        )
    }

    private func sourceTypeDeleteImpact(projectDir: String, id: String) -> CatalogDeleteImpact {
        guard let type = (sourceTypesByProject[projectDir] ?? []).first(where: { $0.id == id }) else {
            return CatalogDeleteImpact(allowed: false, gate: .notFound, groups: [])
        }
        if CatalogOrigin.isPlugin(type.origin) {
            return CatalogDeleteImpact(allowed: false, gate: .originLocked, groups: [])
        }
        let inbound = (sourcesByProject[projectDir] ?? []).filter { $0.sourceTypeID == id }
        let total = inbound.isEmpty ? type.usedBy : inbound.count
        if total == 0 {
            return CatalogDeleteImpact(allowed: true, gate: .ok, groups: [])
        }
        return CatalogDeleteImpact(
            allowed: false,
            gate: .inbound,
            groups: [
                CatalogDeleteImpactGroup(
                    via: "sources.source_type_id",
                    kind: "source",
                    total: total,
                    listed: inbound.prefix(20).map {
                        CatalogDeleteImpactListed(
                            id: $0.id,
                            ref: $0.ref,
                            title: $0.title.isEmpty ? $0.ref : $0.title,
                            location: WorkspaceLocation.sourcePage(
                                sourceId: $0.id,
                                ref: $0.ref,
                                title: $0.title.isEmpty ? $0.ref : $0.title
                            )
                        )
                    }
                ),
            ]
        )
    }

    private func citationDeleteImpact(projectDir: String, id: String) -> CatalogDeleteImpact {
        guard citationsByID[id] != nil else {
            return CatalogDeleteImpact(allowed: false, gate: .notFound, groups: [])
        }
        let observations = observationsByGraph.values.flatMap { $0 }.filter { $0.citationID == id }
        if observations.isEmpty {
            return CatalogDeleteImpact(allowed: true, gate: .ok, groups: [])
        }
        return CatalogDeleteImpact(
            allowed: false,
            gate: .inbound,
            groups: [observationImpactGroup(via: "observations.citation_id", observations: observations)]
        )
    }

    private func observationDeleteImpact(projectDir: String, id: String) -> CatalogDeleteImpact {
        let existing = observationsByGraph.values.flatMap { $0 }.first { $0.id == id }
        guard let existing else {
            return CatalogDeleteImpact(allowed: false, gate: .notFound, groups: [])
        }
        if isEdgeLocked(observation: existing, projectDir: projectDir) {
            return CatalogDeleteImpact(allowed: false, gate: .edgeLocked, groups: [])
        }
        let handles = pinsByObservation[id] ?? []
        guard !handles.isEmpty else {
            return CatalogDeleteImpact(allowed: true, gate: .ok, groups: [])
        }
        return CatalogDeleteImpact(
            allowed: true,
            gate: .ok,
            groups: [],
            cascades: [
                CatalogDeleteImpactGroup(
                    via: "identity_claim_evidence.observation_id",
                    kind: "canonical_entity",
                    total: handles.count,
                    listed: handles.map { entity in
                        CatalogDeleteImpactListed(
                            id: entity.id,
                            ref: entity.ref,
                            title: entity.ref,
                            location: WorkspaceLocation(section: .sources, ref: entity.ref, title: entity.ref)
                        )
                    }
                ),
            ]
        )
    }

    private func subjectDeleteImpact(projectDir: String, id: String) -> CatalogDeleteImpact {
        guard subjectsByGraph.values.contains(where: { $0.contains(where: { $0.id == id }) }) else {
            return CatalogDeleteImpact(allowed: false, gate: .notFound, groups: [])
        }
        let observations = observationsByGraph.values.flatMap { $0 }
        let asEndpoint = observations.filter { $0.valueSubjectID == id }
        let asOwner = observations.filter {
            $0.subjectID == id && !isConnectionFacet($0, projectDir: projectDir)
        }
        var groups: [CatalogDeleteImpactGroup] = []
        if !asOwner.isEmpty {
            groups.append(observationImpactGroup(via: "observations.subject_id", observations: asOwner))
        }
        if !asEndpoint.isEmpty {
            groups.append(observationImpactGroup(via: "observations.value_subject_id", observations: asEndpoint))
        }
        var cascades: [CatalogDeleteImpactGroup] = []
        if let entity = membershipBySubject[id]?.entity {
            cascades.append(
                CatalogDeleteImpactGroup(
                    via: "identity_claims.subject_id",
                    kind: "canonical_entity",
                    total: 1,
                    listed: [
                        CatalogDeleteImpactListed(
                            id: entity.id,
                            ref: entity.ref,
                            title: entity.ref,
                            location: WorkspaceLocation(section: .sources, ref: entity.ref, title: entity.ref)
                        ),
                    ]
                )
            )
        }
        if groups.isEmpty {
            return CatalogDeleteImpact(allowed: true, gate: .ok, groups: [], cascades: cascades)
        }
        return CatalogDeleteImpact(allowed: false, gate: .inbound, groups: groups, cascades: cascades)
    }

    private func observationImpactGroup(
        via: String,
        observations: [CatalogObservation]
    ) -> CatalogDeleteImpactGroup {
        CatalogDeleteImpactGroup(
            via: via,
            kind: "observation",
            total: observations.count,
            listed: observations.prefix(20).map { observation in
                let owner = subjectsByGraph.values.flatMap { $0 }.first(where: { $0.id == observation.subjectID })
                var title = observation.propertyLabel
                if !observation.valueText.isEmpty {
                    title = title.isEmpty
                        ? observation.valueText
                        : "\(title): \(observation.valueText)"
                }
                if title.isEmpty {
                    title = observation.ref
                }
                return CatalogDeleteImpactListed(
                    id: observation.id,
                    ref: observation.ref,
                    title: title,
                    location: WorkspaceLocation.citationComposer(
                        sourceId: owner?.sourceID ?? "",
                        graphId: graphOfSubject(observation.subjectID) ?? "",
                        subjectId: observation.subjectID,
                        citationId: observation.citationID,
                        observationId: observation.id,
                        ref: observation.ref,
                        title: owner?.label.isEmpty == false ? owner?.label : observation.ref
                    )
                )
            }
        )
    }

    private func isConnectionFacet(_ observation: CatalogObservation, projectDir: String) -> Bool {
        if isEdgeLocked(observation: observation, projectDir: projectDir) {
            return true
        }
        let propertyKey = observation.propertyKey.isEmpty
            ? (propertiesByProject[projectDir]?.first(where: { $0.id == observation.propertyID })?.key ?? "")
            : observation.propertyKey
        return propertyKey == SeededPropertyKey.role || propertyKey == SeededPropertyKey.relationshipType
    }

    private func fixtureContains(projectDir: String, id: String) -> Bool {
        if (sourcesByProject[projectDir] ?? []).contains(where: { $0.id == id }) { return true }
        if artifactsBySource.values.contains(where: { $0.contains(where: { $0.id == id }) }) { return true }
        if (subjectTypesByProject[projectDir] ?? []).contains(where: { $0.id == id }) { return true }
        if subjectsByGraph.values.contains(where: { $0.contains(where: { $0.id == id }) }) { return true }
        if (sourceTypesByProject[projectDir] ?? []).contains(where: { $0.id == id }) { return true }
        if (fieldsByProject[projectDir] ?? []).contains(where: { $0.id == id }) { return true }
        if (propertiesByProject[projectDir] ?? []).contains(where: { $0.id == id }) { return true }
        if propertyTermsByProperty.values.contains(where: { $0.contains(where: { $0.id == id }) }) { return true }
        if (credibilityGradesByProject[projectDir] ?? []).contains(where: { $0.id == id }) { return true }
        if citationsByID[id] != nil { return true }
        if observationsByGraph.values.contains(where: { $0.contains(where: { $0.id == id }) }) { return true }
        return false
    }

    /// Subject type keys Promote starts from; the fake has no origin to check.
    private static let primaryTypeKeys = Set(EvidencePrimaryKind.allCases.map(\.rawValue))

    private func graphProgress(for graphID: String, projectDir: String) -> EvidenceGraphProgress {
        if let override = graphProgressByGraph[graphID] {
            return override
        }
        let sourceTypeIDs = Set(
            (subjectTypesByProject[projectDir] ?? []).filter { $0.key == "source" }.map(\.id)
        )
        let subjects = (subjectsByGraph[graphID] ?? []).filter { !sourceTypeIDs.contains($0.subjectTypeID) }
        let observations = observationsByGraph[graphID] ?? []
        let primaryTypeIDs = Set(
            (subjectTypesByProject[projectDir] ?? []).filter { Self.primaryTypeKeys.contains($0.key) }.map(\.id)
        )
        return EvidenceGraphProgress(
            graphId: graphID,
            sourceId: sourceOfGraph(graphID) ?? "",
            subjectCount: subjects.count,
            observationCount: observations.count,
            primaryCount: subjects.filter { primaryTypeIDs.contains($0.subjectTypeID) }.count,
            filedCount: 0
        )
    }

    /// The engine's rule for a graph's Citations: one with an Observation in the
    /// graph, or one with no Observation anywhere yet.
    private func citationIsInGraph(_ citationID: String, graphID: String) -> Bool {
        if (observationsByGraph[graphID] ?? []).contains(where: { $0.citationID == citationID }) {
            return true
        }
        return !observationsByGraph.values.contains { $0.contains { $0.citationID == citationID } }
    }

    private func personNames(for citationID: String, graphID: String, projectDir: String) -> [String] {
        let personTypeIDs = Set(
            (subjectTypesByProject[projectDir] ?? []).filter { $0.key == "person" }.map(\.id)
        )
        let subjects = subjectsByGraph[graphID] ?? []
        func label(_ id: String) -> String? {
            guard let subject = subjects.first(where: { $0.id == id }),
                  personTypeIDs.contains(subject.subjectTypeID)
            else { return nil }
            return subject.trimmedLabel
        }
        var firstID: [String: String] = [:]
        var labels: [String: String] = [:]
        let observations = (observationsByGraph[graphID] ?? [])
            .filter { $0.citationID == citationID }
            .sorted { $0.id < $1.id }
        for observation in observations {
            if let name = label(observation.subjectID), firstID[observation.subjectID] == nil {
                firstID[observation.subjectID] = observation.id
                labels[observation.subjectID] = name
            }
            if let name = label(observation.valueSubjectID), firstID[observation.valueSubjectID] == nil {
                firstID[observation.valueSubjectID] = observation.id
                labels[observation.valueSubjectID] = name
            }
        }
        return firstID
            .sorted { $0.value < $1.value }
            .prefix(6)
            .compactMap { labels[$0.key] }
    }

    func listCitationsByArtifact(
        projectDir: String,
        graphID: String,
        artifactID: String
    ) async throws -> [CatalogListedCitation] {
        return withState {
            markCatalogSessionHeld(projectDir)
            let listed = citationsByID.values
                .filter { $0.artifactID == artifactID && citationIsInGraph($0.id, graphID: graphID) }
                .sorted { $0.ref.localizedCaseInsensitiveCompare($1.ref) == .orderedAscending }
            return listed.map { citation in
                let count = (observationsByGraph[graphID] ?? [])
                    .filter { $0.citationID == citation.id }
                    .count
                return CatalogListedCitation(
                    citation: citation,
                    observationCount: count,
                    subjectNames: personNames(for: citation.id, graphID: graphID, projectDir: projectDir)
                )
            }
        }
    }

    private func fakeObservation(
        citationID: String,
        id: String,
        ref: String,
        row: CatalogObservation
    ) throws -> CatalogObservation {
        guard subjectsByGraph.values.flatMap({ $0 }).contains(where: { $0.id == row.subjectID }) else {
            throw StoreBoom.boom
        }
        let property = propertiesByProject.values.flatMap { $0 }.first(where: { $0.id == row.propertyID })
        var displayText = row.valueText
        if displayText.isEmpty, !row.valueTermID.isEmpty {
            displayText = propertyTermsByProperty[row.propertyID]?
                .first(where: { $0.id == row.valueTermID })?
                .label ?? ""
        }
        if displayText.isEmpty, !row.nameForm.isEmpty {
            displayText = row.nameForm
        }
        if displayText.isEmpty, let date = row.date {
            displayText = DateValueDisplay.string(for: date)
        }
        if displayText.isEmpty, !row.valueSubjectID.isEmpty {
            displayText = subjectsByGraph.values
                .flatMap { $0 }
                .first(where: { $0.id == row.valueSubjectID })?
                .label ?? ""
        }
        return CatalogObservation(
            id: id,
            ref: ref,
            citationID: citationID,
            subjectID: row.subjectID,
            propertyID: row.propertyID,
            polarity: row.polarity.isEmpty ? "positive" : row.polarity,
            valueText: displayText,
            valueInteger: row.valueInteger,
            valueDateID: row.valueDateID,
            date: row.date,
            valueNameID: row.valueNameID,
            nameForm: row.nameForm,
            nameParts: row.nameParts,
            valueSubjectID: row.valueSubjectID,
            valueTermID: row.valueTermID,
            propertyKey: row.propertyKey.isEmpty ? (property?.key ?? "") : row.propertyKey,
            propertyLabel: row.propertyLabel.isEmpty ? (property?.label ?? "") : row.propertyLabel,
            propertyValueType: row.propertyValueType.isEmpty ? (property?.valueType ?? "") : row.propertyValueType,
            valueTermKey: row.valueTermKey.isEmpty
                ? (propertyTermsByProperty[row.propertyID]?.first(where: { $0.id == row.valueTermID })?.key ?? "")
                : row.valueTermKey,
            valueTermOrigin: row.valueTermOrigin.isEmpty
                ? (propertyTermsByProperty[row.propertyID]?.first(where: { $0.id == row.valueTermID })?.origin ?? "")
                : row.valueTermOrigin
        )
    }

    /// ` depiction=1 depictionFile=present|missing` for a depiction input, else nil.
    private func depictionRecord(_ input: CatalogDepictionValueInput?) -> String? {
        guard let input else { return nil }
        recordedDepictionPaths.append(input.imagePath)
        let present = FileManager.default.fileExists(atPath: input.imagePath)
        return " depiction=1 depictionFile=\(present ? "present" : "missing")"
    }

    private func nextFakeDepiction(_ input: CatalogDepictionValueInput) -> (String, CatalogDepictionValue) {
        depictionSequence += 1
        let id = "dep-\(depictionSequence)"
        return (id, CatalogDepictionValue(regionJSON: input.regionJSON, imageRelPath: "objects/fake/\(id).png"))
    }

    private func appendFakeObservations(
        projectDir: String,
        citationID: String,
        drafts: [CatalogObservationDraft]
    ) throws -> [CatalogObservation] {
        if drafts.isEmpty {
            return []
        }
        var created: [CatalogObservation] = []
        for draft in drafts {
            guard let graphID = graphOfSubject(draft.subjectID) else {
                throw StoreBoom.boom
            }
            let property = propertiesByProject[projectDir]?.first(where: { $0.id == draft.propertyID })
            var displayText = draft.valueText
            if displayText.isEmpty, !draft.valueTermID.isEmpty {
                displayText = propertyTermsByProperty[draft.propertyID]?
                    .first(where: { $0.id == draft.valueTermID })?
                    .label ?? ""
            }
            if displayText.isEmpty, !draft.nameForm.isEmpty {
                displayText = draft.nameForm
            }
            if displayText.isEmpty, let date = draft.date {
                displayText = DateValueDisplay.string(for: date)
            }
            if displayText.isEmpty, !draft.valueSubjectID.isEmpty {
                displayText = subjectsByGraph.values
                    .flatMap { $0 }
                    .first(where: { $0.id == draft.valueSubjectID })?
                    .label ?? ""
            }
            var obs = CatalogObservation(
                id: UUID().uuidString.lowercased(),
                ref: "OBS-FAKE1",
                citationID: citationID,
                subjectID: draft.subjectID,
                propertyID: draft.propertyID,
                polarity: draft.polarity.isEmpty ? "positive" : draft.polarity,
                valueText: displayText,
                valueInteger: draft.valueInteger,
                valueDateID: draft.valueDateID,
                date: draft.date,
                valueNameID: draft.valueNameID,
                nameForm: draft.nameForm,
                nameParts: draft.nameParts,
                valueSubjectID: draft.valueSubjectID,
                valueTermID: draft.valueTermID,
                propertyKey: property?.key ?? "",
                propertyLabel: property?.label ?? "",
                propertyValueType: property?.valueType ?? "",
                valueTermKey: propertyTermsByProperty[draft.propertyID]?
                    .first(where: { $0.id == draft.valueTermID })?
                    .key ?? "",
                valueTermOrigin: propertyTermsByProperty[draft.propertyID]?
                    .first(where: { $0.id == draft.valueTermID })?
                    .origin ?? ""
            )
            if let input = draft.depiction {
                let stored = nextFakeDepiction(input)
                obs.valueDepictionID = stored.0
                obs.depiction = stored.1
            }
            observationsByGraph[graphID, default: []].append(obs)
            created.append(obs)
        }
        return created
    }

    /// Appends a new Source with its Default graph, which links the Source's
    /// Artifacts (a copy's), and bumps it like the engine's `sources` insert,
    /// reporting the Default graph. Call inside `withState`.
    private func insertSource(
        _ source: CatalogSource,
        projectDir: String
    ) -> (CatalogSource, CatalogWriteEffects) {
        sourcesByProject[projectDir, default: []].append(source)
        let graph = appendGraph(
            projectDir: projectDir, sourceID: source.id, name: "Default", description: "",
            artifactIDs: (artifactsBySource[source.id] ?? []).map(\.id)
        )
        let effects = bumpSource(source.id, graphs: [graph.id])
        return (sourcesByProject[projectDir]?.first { $0.id == source.id } ?? source, effects)
    }

    /// Appends an Evidence graph to the project's list. Call inside `withState`.
    private func appendGraph(
        projectDir: String,
        sourceID: String,
        name: String,
        description: String,
        artifactIDs: [String]
    ) -> CatalogEvidenceGraph {
        let graph = CatalogEvidenceGraph(
            id: UUID().uuidString.lowercased(),
            ref: "GPH-FAKE1",
            sourceID: sourceID,
            name: name,
            description: description,
            artifactIDs: artifactIDs
        )
        evidenceGraphsByProject[projectDir, default: []].append(graph)
        return graph
    }

    /// The Source of an Evidence graph, over every project's graph list. Call inside `withState`.
    private func sourceOfGraph(_ graphID: String) -> String? {
        for graphs in evidenceGraphsByProject.values {
            if let graph = graphs.first(where: { $0.id == graphID }) { return graph.sourceID }
        }
        return nil
    }

    /// The ids of every Evidence graph of a Source. Call inside `withState`.
    private func graphIDsOfSource(_ sourceID: String) -> [String] {
        var ids = evidenceGraphsByProject.values.flatMap { $0 }.filter { $0.sourceID == sourceID }.map(\.id)
        for (graphID, list) in subjectsByGraph where !ids.contains(graphID) && list.first?.sourceID == sourceID {
            ids.append(graphID)
        }
        return ids
    }

    /// Every Subject of a Source, over its graphs. Call inside `withState`.
    private func subjectsOfSource(_ sourceID: String) -> [CatalogSubject] {
        graphIDsOfSource(sourceID).flatMap { subjectsByGraph[$0] ?? [] }
    }

    /// The Evidence graph a Subject is on. Call inside `withState`.
    private func graphOfSubject(_ subjectID: String) -> String? {
        subjectsByGraph.first(where: { $0.value.contains { $0.id == subjectID } })?.key
    }

    /// The graph list order, as `evidencegraphs.sqlListOrder`: name without
    /// case, then ref, then id.
    private static func graphListOrder(_ a: CatalogEvidenceGraph, _ b: CatalogEvidenceGraph) -> Bool {
        let order = a.name.localizedCaseInsensitiveCompare(b.name)
        if order != .orderedSame { return order == .orderedAscending }
        return a.ref == b.ref ? a.id < b.id : a.ref < b.ref
    }

    /// The first graph, by name, an Artifact is linked to. Call inside `withState`.
    private func firstGraphID(ofArtifact artifactID: String) -> String? {
        evidenceGraphsByProject.values.flatMap { $0 }
            .filter { $0.artifactIDs.contains(artifactID) }
            .sorted(by: Self.graphListOrder)
            .first?.id
    }

    private static func unique(_ ids: [String]) -> [String] {
        var seen = Set<String>()
        return ids.filter { seen.insert($0).inserted }
    }

    /// Runs `body` under `stateLock`. Never spans an `await`: methods that
    /// suspend take the lock separately for each synchronous stretch.
    /// Mirrors the engine's audit scopes: any write under a Source moves its
    /// "Updated" revision, not just edits to the Source row.
    /// Returns the effects the engine would report for that write: the Source
    /// scope and the new revision, plus the Evidence graphs whose rows it
    /// changes (`graphs`: a Subject write names its Subject's graph, an
    /// Artifact or Citation write every graph of the Source) and the Source in
    /// `citationSources` when `graphs` is not empty. A Source in no list
    /// reports its graphs and no activity; no Source reports nothing.
    @discardableResult
    private func bumpSource(_ sourceID: String?, graphs: [String]) -> CatalogWriteEffects {
        guard let sourceID else { return .empty }
        for (project, var list) in sourcesByProject {
            guard let idx = list.firstIndex(where: { $0.id == sourceID }) else { continue }
            let activity = CatalogActivity(revision: nextAuditRevision, at: "2026-10-15T16:00:00Z")
            list[idx].updated = activity
            nextAuditRevision += 1
            sourcesByProject[project] = list
            return CatalogWriteEffects(
                evidenceGraphs: graphs,
                citationSources: graphs.isEmpty ? [] : [sourceID],
                sourceScopes: [sourceID],
                activity: activity
            )
        }
        // A Source no list holds has no activity to move, but the write still
        // reports the graphs it changed.
        return CatalogWriteEffects(evidenceGraphs: graphs, citationSources: graphs.isEmpty ? [] : [sourceID])
    }

    private func owningSource(ofCitation citationID: String) -> String? {
        if let artifactID = citationsByID[citationID]?.artifactID,
           let owner = artifactsBySource.first(where: { entry in entry.value.contains { $0.id == artifactID } })?.key
        {
            return owner
        }
        return observationsByGraph.first(where: { entry in
            entry.value.contains { $0.citationID == citationID }
        }).flatMap { sourceOfGraph($0.key) }
    }

    private func withState<T>(_ body: () throws -> T) rethrows -> T {
        try stateLock.withLock(body)
    }

    private func markCatalogSessionHeld(_ projectDir: String) {
        heldCatalogProjectDir = projectDir
    }

    /// Exact catalog ref or PREFIX-token prefix (mirrors core/search classifyRefQuery).
    private static func classifyRefQuery(_ query: String) -> (key: String, exact: Bool)? {
        let s = query.trimmingCharacters(in: .whitespacesAndNewlines).uppercased()
        guard !s.isEmpty else { return nil }
        let exactPattern = /^[A-Z]{3}-[0-9A-HJKMNP-TV-Z]{5}$/
        if s.wholeMatch(of: exactPattern) != nil {
            return (s, true)
        }
        let prefixPattern = /^[A-Z]{3}-[0-9A-HJKMNP-TV-Z]{1,4}$/
        if s.wholeMatch(of: prefixPattern) != nil {
            return (s, false)
        }
        return nil
    }

    private static func tokenizeSearch(_ query: String) -> [String] {
        let trimmed = query.trimmingCharacters(in: .whitespacesAndNewlines).lowercased()
        guard !trimmed.isEmpty else { return [] }
        return trimmed.split(whereSeparator: \.isWhitespace).compactMap { part in
            let cleaned = part.trimmingCharacters(in: CharacterSet(charactersIn: "\"'.,;:!?()[]{}"))
            return cleaned.isEmpty ? nil : String(cleaned)
        }
    }

    private static func scoreSearchFields(
        fields: [(name: String, weight: Double)],
        values: [String: String],
        tokens: [String]
    ) -> (score: Double, reason: String) {
        let tokenKeys = tokens.map { TextMatch.key($0, locale: TextMatch.posix) }.filter { !$0.isEmpty }
        guard !tokenKeys.isEmpty else { return (0, "") }
        let valueKeys = Dictionary(fields.map {
            ($0.name, TextMatch.key(values[$0.name] ?? "", locale: TextMatch.posix))
        }, uniquingKeysWith: { first, _ in first })
        var score = 0.0
        var bestField = ""
        var bestWeight = 0.0
        var matchedTokens = 0
        for tok in tokenKeys {
            var tokHit = false
            for field in fields {
                let value = valueKeys[field.name] ?? ""
                guard !value.isEmpty, value.contains(tok) else { continue }
                tokHit = true
                score += field.weight
                if field.weight > bestWeight || (field.weight == bestWeight && bestField.isEmpty) {
                    bestWeight = field.weight
                    bestField = field.name
                }
            }
            if tokHit { matchedTokens += 1 }
        }
        guard matchedTokens > 0 else { return (0, "") }
        score *= Double(matchedTokens) / Double(tokenKeys.count)
        return (score, bestField)
    }

    /// Resolves list/identity paint fields from persisted cover mode.
    /// Source cover is type icon or a raster path — never Artifact file-type MIME.
    /// Also sets `hasArtifact` from the Artifact bag (fileless counts).
    private func enrichCoverFields(_ source: CatalogSource) -> CatalogSource {
        var copy = source
        let arts = artifactsBySource[copy.id] ?? []
        copy.hasArtifact = !arts.isEmpty
        copy.thumbnailRelPath = ""
        copy.thumbnailMediaType = ""
        copy.thumbnailOriginalFilename = ""
        guard copy.coverMode == "artifact", !copy.primaryArtifactID.isEmpty else {
            return copy
        }
        guard let art = arts.first(where: { $0.id == copy.primaryArtifactID }),
              !art.thumbnailRelPath.isEmpty
        else {
            return copy
        }
        copy.thumbnailRelPath = art.thumbnailRelPath
        return copy
    }

    private static func originCounts(from origins: [String]) -> WorkspaceNavOriginCounts {
        var seeded = 0
        var user = 0
        var plugin = 0
        for origin in origins {
            switch origin {
            case CatalogOrigin.provenencia: seeded += 1
            case CatalogOrigin.user: user += 1
            default: plugin += 1
            }
        }
        return WorkspaceNavOriginCounts(
            total: origins.count,
            seeded: seeded,
            user: user,
            plugin: plugin
        )
    }
}

private enum StoreBoom: Error { case boom }
#endif
