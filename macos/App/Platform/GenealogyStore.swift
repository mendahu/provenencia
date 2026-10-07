import Foundation

struct InstallIdentity: Sendable, Equatable {
    var userID: String
    var displayName: String
    var ref: String
}

struct ProjectInfo: Sendable, Equatable {
    var label: String
    var folderName: String
    var createdAt: String
    var updatedAt: String
    var updatedByUserID: String
    var updatedByDisplayName: String
    var updatedByRef: String
    /// Durable catalog project identity (UUIDv7 string). Empty only if absent.
    var uuid: String = ""
}

struct OnboardingResult: Sendable, Equatable {
    var projectDir: String
    var userID: String
    var displayName: String
    var ref: String
    var project: ProjectInfo
}

struct CatalogSource: Sendable, Equatable, Identifiable {
    var id: String
    var ref: String
    var sourceTypeID: String
    var title: String
    var description: String
    /// Thumbnail JPEG under `objects/…` when cover is a pinned raster Artifact.
    var thumbnailRelPath: String = ""
    /// Unused on Source (file-type glyphs stay on Artifact rows). Kept for wire compat.
    var thumbnailMediaType: String = ""
    var thumbnailOriginalFilename: String = ""
    /// `artifact` (pinned raster primary) or `type_icon`.
    var coverMode: String = "type_icon"
    /// Set when `coverMode` is `artifact`.
    var primaryArtifactID: String = ""
    /// True when this Source has at least one Artifact (fileless counts).
    /// Used by the Sources list Evidence graph gate — not cover presence.
    var hasArtifact: Bool = false
    /// Latest audit revision of any work scoped to this Source (its row, notes,
    /// metadata, artifacts, Evidence graph). Sources list "Updated" sort; zero
    /// means unset / unknown.
    var updatedRevision: Int64 = 0
}

struct CatalogSourceNote: Sendable, Equatable {
    var id: String
    var sourceID: String
    var body: String
    /// Create attribution from audit (not a domain column).
    var authorDisplayName: String
    /// RFC3339 UTC create time from audit.
    var createdAt: String
}

struct CatalogFileRef: Sendable, Equatable {
    var id: String
    var relPath: String
    var originalFilename: String
    var mediaType: String
    var byteSize: Int64
}

struct CatalogArtifact: Sendable, Equatable {
    var id: String
    var ref: String
    var sourceID: String
    var fileID: String
    var label: String
    var description: String
    var file: CatalogFileRef?
    /// Thumbnail JPEG under `objects/…` (empty when fileless / non-image / skipped).
    var thumbnailRelPath: String = ""
}

struct CatalogCredibilityGrade: Sendable, Equatable, Identifiable {
    var id: String
    var key: String
    var origin: String
    var label: String
    var sortOrder: Int
}

enum CatalogDeleteImpactGate: String, Sendable, Equatable {
    case ok
    case inbound
    case notFound = "not_found"
    case edgeLocked = "edge_locked"
    case infra
    case originLocked = "origin_locked"
}

struct CatalogDeleteImpact: Sendable, Equatable {
    var allowed: Bool
    var gate: CatalogDeleteImpactGate
    var groups: [CatalogDeleteImpactGroup]
    /// Non-blocking: rows that go with the target on erase (a Subject's handle membership).
    var cascades: [CatalogDeleteImpactGroup] = []
}

struct CatalogDeleteImpactGroup: Sendable, Equatable {
    var via: String
    var kind: String
    var total: Int
    var listed: [CatalogDeleteImpactListed]
}

struct CatalogDeleteImpactListed: Sendable, Equatable, Identifiable {
    var id: String
    var ref: String
    var title: String
    var location: WorkspaceLocation
}

/// One omnibar / SearchCatalog hit (stable kinds: source, source_type, metadata_field).
struct CatalogSearchHit: Sendable, Equatable, Identifiable {
    var kind: String
    var id: String
    var ref: String
    var title: String
    var subtitle: String
    /// Stable match field code from Go (`title`, `notes`, …) — not UI copy.
    var matchReason: String
    /// Optional raw snippet for body/rollup matches; Mac localizes the prefix.
    var matchSnippet: String = ""
    var location: WorkspaceLocation
    /// Source cover raster path when already derived; empty otherwise.
    var thumbnailRelPath: String = ""
    /// Source-type icon key (type hits and Source type fallback).
    var iconKey: String = ""
    /// Accepted members, for `person` / `event` / `place` hits.
    var memberCount: Int = 0
}

struct CatalogSubjectType: Sendable, Equatable, Identifiable {
    var id: String
    var key: String
    var origin: String
    var label: String
    var description: String
    var refPrefix: String
    var candidateRefPrefix: String
}

struct CatalogSubject: Sendable, Equatable, Identifiable {
    var id: String
    var ref: String
    var sourceID: String
    var subjectTypeID: String
    var label: String
    var description: String
}

/// A Conclusion handle (PER-…, EVT-…, PLC-…).
struct CatalogCanonicalEntity: Sendable, Equatable, Identifiable {
    var id: String
    var ref: String
    var subjectTypeID: String
    var label: String
}

/// Files one Subject onto one handle; accepted claims are members.
struct CatalogIdentityClaim: Sendable, Equatable, Identifiable {
    var id: String
    var subjectID: String
    var entityID: String
    var status: String
    /// `nil` when the claim carries no confidence grade.
    var confidenceGradeID: String? = nil
    var argument: String = ""
}

/// One Property's part in a match score (core/match): its best similarity,
/// 0…1, and the points it added (negative for a clear disagreement).
struct CatalogMatchReason: Sendable, Equatable {
    var propertyKey: String
    var propertyOrigin: String
    var similarity: Double
    var contribution: Double
}

/// One existing handle a Subject could join in Promote, best first. `person`
/// is the row header for Person handles; Events and Places carry the handle
/// alone until their header composers land.
struct CatalogPromoteTargetSuggestion: Sendable, Equatable, Identifiable {
    var entity: CatalogCanonicalEntity
    var score: Double
    var reasons: [CatalogMatchReason]
    var person: CatalogPersonHeader?
    /// Accepted members of the handle.
    var memberCount: Int = 0

    var id: String { entity.id }
}

/// One claim confidence grade (Low / Moderate / High): the scale on Identity
/// Claims. Not Source credibility, though the shape matches.
struct CatalogClaimConfidenceGrade: Sendable, Equatable, Identifiable {
    var id: String
    var key: String
    var origin: String
    var label: String
    var sortOrder: Int
}

/// A view of one Subject's **accepted Identity Claim**: the claim and the handle
/// it files the Subject onto. "Membership" is the data model's word for that
/// relation (§6); it is never stored on its own, and provisional / rejected
/// claims are not memberships. `kind` is the handle's subject type key
/// (person, event, place, …). Unpromoted Subjects have none.
struct CatalogSubjectMembership: Sendable, Equatable {
    var subjectID: String
    var claimID: String
    var entity: CatalogCanonicalEntity
    var kind: String
    /// The handle's displayed auto-reconciled name (S9-09); nil when it has none.
    var name: CatalogNameValue? = nil
}

/// One saved Promote step: the handle and the claim it wrote.
struct CatalogPromoteResult: Sendable, Equatable {
    var entity: CatalogCanonicalEntity
    var claim: CatalogIdentityClaim
}

struct CatalogSubjectPosition: Sendable, Equatable {
    var subjectID: String
    var gridX: Int64
    var gridY: Int64
}

struct CatalogCitation: Sendable, Equatable, Identifiable {
    var id: String
    var ref: String
    var artifactID: String
    var locatorJSON: String
    var transcription: String
    var description: String
    var transcriptionUncertain: Bool
    var transcriptionNote: String
}

/// Citation list row for the composer identity menu (ref + transcription + count).
struct CatalogListedCitation: Sendable, Equatable, Identifiable {
    var citation: CatalogCitation
    var observationCount: Int
    var id: String { citation.id }
}

/// One ordered NameValue part as stored on an Observation.
struct CatalogNameValuePart: Sendable, Equatable {
    var value: String
    var type: String
}

/// A structured NameValue: the full-form reading plus optional ordered parts.
struct CatalogNameValue: Sendable, Equatable {
    var form: String
    var parts: [CatalogNameValuePart] = []
}

/// One Place reached from a birth, death, or event. `names` are kept toponyms
/// in rank order. The parent chain is not here.
struct CatalogHeaderPlace: Sendable, Equatable {
    /// The Place; its page owns the names' Why.
    var entity: CatalogCanonicalEntity
    var names: [String] = []
}

/// A birth or a death composed from the canonical graph.
struct CatalogLifeFacts: Sendable, Equatable {
    /// The birth or death event read; `nil` when none is linked. Its page
    /// owns the date's Why.
    var event: CatalogCanonicalEntity?
    /// Surviving birth (or death) events; above 1 is a disagreement, and
    /// `event` is the earliest dated one.
    var eventCount: Int = 0
    var date: CatalogDateValueInput?
    var dateCount: Int = 0
    var places: [CatalogHeaderPlace] = []
}

/// One subject-role person on an event.
struct CatalogEventSubject: Sendable, Equatable {
    var entity: CatalogCanonicalEntity
    var name: CatalogNameValue?
    var nameValueCount: Int = 0
}

/// One Person as a row, composed by Go from the auto-reconciler cache (S9-07).
/// Structures only; `PersonHeaderDisplay` formats the title. Birth and death
/// are read by `PersonLifeDisplay`.
struct CatalogPersonHeader: Sendable, Equatable, Identifiable {
    var entity: CatalogCanonicalEntity
    /// Displayed auto-reconciled name; `nil` when no member names the Person.
    var name: CatalogNameValue?
    /// Displayed name values (names are one structure, so at most 1).
    var nameValueCount: Int
    var birth: CatalogLifeFacts = CatalogLifeFacts()
    var death: CatalogLifeFacts = CatalogLifeFacts()

    var id: String { entity.id }
}

/// One Event as a row, composed by Go from the auto-reconciler cache (S9-22).
/// Structures only; `EventTitleDisplay` formats the title. A point `date`
/// wins over `startDate` / `endDate`.
struct CatalogEventHeader: Sendable, Equatable, Identifiable {
    var entity: CatalogCanonicalEntity
    var eventName: String = ""
    var eventNameCount: Int = 0
    var eventTypeKey: String = ""
    var eventTypeLabel: String = ""
    var eventTypeCount: Int = 0
    var date: CatalogDateValueInput?
    var dateCount: Int = 0
    var startDate: CatalogDateValueInput?
    var startDateCount: Int = 0
    var endDate: CatalogDateValueInput?
    var endDateCount: Int = 0
    /// Subject-role persons, then every location's kept names.
    var subjects: [CatalogEventSubject] = []
    var places: [CatalogHeaderPlace] = []
    /// The naming-matrix rule Go chose, and its parts.
    var title: CatalogEventTitle

    var id: String { entity.id }
}

/// An Event's title as Go chose it (Spike 9 R4): which row of the naming
/// matrix applies, and the parts that row reads. Canonical headers and
/// Evidence graph cards both carry one, so they can't disagree.
/// `EventTitleDisplay` fills the rule's L10n template; it does not choose.
struct CatalogEventTitle: Sendable, Equatable {
    enum Rule: Sendable, Equatable {
        case recordedName
        case subject
        case couple
        case subjects
        case label
        case typeAtPlace
        case type
        case ref
    }

    var rule: Rule
    var recordedName: String = ""
    var label: String = ""
    var ref: String = ""
    var typeKey: String = ""
    var typeLabel: String = ""
    /// Subject-role persons in stable order; `nil` reads "unnamed person".
    var subjects: [CatalogNameValue?] = []
    var place: String = ""
}

/// One Place as a row, composed by Go from the auto-reconciler cache (S9-25).
/// `names` are the kept toponyms in rank order. Period, kind, and parents
/// stay empty until S9-38 and S9-39. `PlaceTitleDisplay` formats the title.
struct CatalogPlaceHeader: Sendable, Equatable, Identifiable {
    var entity: CatalogCanonicalEntity
    var names: [String] = []
    var startDate: CatalogDateValueInput?
    var endDate: CatalogDateValueInput?
    var kind: String = ""
    var parents: [String] = []

    var id: String { entity.id }
}

/// One Property value on a Conclusion detail: exactly one case per value type.
enum CatalogConclusionValue: Sendable, Equatable {
    case none
    case text(String)
    case integer(Int64)
    case term(id: String, key: String, label: String)
    case date(CatalogDateValueInput)
    case name(CatalogNameValue)
}

/// One auto-reconciled value of a Property (S9-14). `.kept` is displayed;
/// any other reason says why it isn't.
struct CatalogReconciledValue: Sendable, Equatable, Identifiable {
    var rank: Int
    var reason: ReconcilerReason
    /// Distinct Sources behind it.
    var support: Int
    /// Negative records that match it.
    var against: Int
    var value: CatalogConclusionValue

    var id: Int { rank }
    var isDisplayed: Bool { reason == .kept }
}

/// What the auto-reconciler did with one Observation, with the record's own
/// value and the evidence it weighed. Empty keys are the scale defaults
/// (standard credibility, moderate confidence).
struct CatalogReconcilerOutcome: Sendable, Equatable, Identifiable {
    var observationID: String
    var observationRef: String
    var reason: ReconcilerReason
    /// The value it went into; nil for none.
    var valueRank: Int?
    /// The negative Observation that denied it; "" when not denied.
    var deniedByObservationID: String
    var recorded: CatalogConclusionValue
    var subjectID: String
    var subjectRef: String
    var citationID: String
    var artifactID: String = ""
    var sourceID: String
    var sourceTitle: String
    var credibilityKey: String
    var transcriptionUncertain: Bool
    var claimConfidenceKey: String
    /// Grade order relative to the default grade, as the auto-reconciler
    /// weighed it: below 0 is weak evidence.
    var credibilityOffset: Int = 0
    var claimConfidenceOffset: Int = 0
    /// For "outvoted": the winning value's Sources of every Source that
    /// voted on that unit. 0 otherwise.
    var voteSupport: Int = 0
    var voteTotal: Int = 0

    var id: String { observationID }
    var isLowTrustSource: Bool { credibilityOffset < 0 }
    var isLowConfidenceClaim: Bool { claimConfidenceOffset < 0 }
}

/// One Property of a handle's detail. Go computes `state`.
struct CatalogConclusionField: Sendable, Equatable, Identifiable {
    var propertyID: String
    var propertyKey: String
    var label: String
    var valueType: PropertyValueType
    var state: ReconciledState
    var values: [CatalogReconciledValue]
    var outcomes: [CatalogReconcilerOutcome]

    var id: String { propertyID }
    var displayedValues: [CatalogReconciledValue] { values.filter(\.isDisplayed) }
}

/// One handle's detail (S9-15), composed by Go from the auto-reconciler cache.
/// Generic over kind: Person, Event and Place pages all read it.
struct CatalogConclusionDetail: Sendable, Equatable {
    var entity: CatalogCanonicalEntity
    var fields: [CatalogConclusionField]
    /// Accepted members.
    var memberCount: Int = 0
    /// The handle's header for its kind, read in the same call, so a page
    /// has one load, one error, and one consistent read.
    var header: CatalogConclusionHeader?

    var personHeader: CatalogPersonHeader? {
        if case .person(let header) = header { header } else { nil }
    }

    var eventHeader: CatalogEventHeader? {
        if case .event(let header) = header { header } else { nil }
    }

    var placeHeader: CatalogPlaceHeader? {
        if case .place(let header) = header { header } else { nil }
    }
}

/// One handle's row for its kind, as the lists show it.
enum CatalogConclusionHeader: Sendable, Equatable {
    case person(CatalogPersonHeader)
    case event(CatalogEventHeader)
    case place(CatalogPlaceHeader)
}

/// One Observation row with Property summary (graph / card payloads).
struct CatalogObservation: Sendable, Equatable, Identifiable {
    var id: String
    var ref: String
    var citationID: String
    var subjectID: String
    var propertyID: String
    var polarity: String
    var valueText: String
    var valueInteger: Int64?
    var valueDateID: String
    /// Structured DateValue when listed from the catalog (locale-aware display).
    var date: CatalogDateValueInput? = nil
    var valueNameID: String
    /// name_values.form when listed (denormalized into valueText as well).
    var nameForm: String = ""
    /// Ordered name_value_parts when listed (empty when form-only).
    var nameParts: [CatalogNameValuePart] = []
    var valueSubjectID: String
    var valueTermID: String
    var propertyKey: String
    var propertyLabel: String
    var propertyValueType: String
    /// Product/user term key when `valueTermID` is set (empty when unknown).
    var valueTermKey: String = ""
}

struct CatalogProperty: Sendable, Equatable, Identifiable {
    var id: String
    var key: String
    var origin: String
    var label: String
    var description: String
    /// text | integer | date | name | subject | term
    var valueType: String
    /// single | multiple. Empty input is treated as single.
    var cardinality: String = "single"
    /// Observation count (`observations.property_id`). Display only; not a delete gate.
    var usedBy: Int = 0
}

struct CatalogPropertyTerm: Sendable, Equatable, Identifiable {
    var id: String
    var propertyID: String
    var key: String
    var origin: String
    var label: String
    var description: String
}

struct CatalogSubjectTypeProperty: Sendable, Equatable, Identifiable {
    var property: CatalogProperty
    var sortOrder: Int
    var locked: Bool

    var id: String { property.id }
}

struct CatalogSubjectTypePresentation: Sendable, Equatable, Identifiable {
    var typeKey: String
    var l10nKey: String
    var iconSymbol: String
    var inkToken: String
    var tintToken: String
    var chipToken: String
    var lineToken: String
    var edgeFromToken: String
    var edgeToToken: String
    var role: String
    var placeable: Bool
    var paletteSort: Int
    var requiresCitationAtCreate: Bool
    var label: String

    var id: String { typeKey }
}

struct CatalogGridCell: Sendable, Equatable {
    var gridX: Int64
    var gridY: Int64
}

struct CatalogConnectEdge: Sendable, Equatable {
    var propertyKey: String
    var endpointTypeKey: String
}

struct CatalogConnectRule: Sendable, Equatable {
    var fromTypeKey: String
    var toTypeKey: String
    var bridgeTypeKey: String
    var edgePropertyKeys: [String]
    var disambiguation: String
    var refuse: Bool
    var edges: [CatalogConnectEdge] = []

    /// FakeStore and unit-test double of `connectrules` product bridges / `All()`.
    /// Live connect reads `listConnectRules` only. When the Go registry changes, update this table in the same change.
    static let productMatrix: [CatalogConnectRule] = [
        CatalogConnectRule(
            fromTypeKey: "person", toTypeKey: "event",
            bridgeTypeKey: "participation", edgePropertyKeys: ["person", "event"],
            disambiguation: "role", refuse: false,
            edges: [
                CatalogConnectEdge(propertyKey: "person", endpointTypeKey: "person"),
                CatalogConnectEdge(propertyKey: "event", endpointTypeKey: "event"),
            ]
        ),
        CatalogConnectRule(
            fromTypeKey: "event", toTypeKey: "person",
            bridgeTypeKey: "participation", edgePropertyKeys: ["person", "event"],
            disambiguation: "role", refuse: false,
            edges: [
                CatalogConnectEdge(propertyKey: "person", endpointTypeKey: "person"),
                CatalogConnectEdge(propertyKey: "event", endpointTypeKey: "event"),
            ]
        ),
        CatalogConnectRule(
            fromTypeKey: "person", toTypeKey: "person",
            bridgeTypeKey: "relationship", edgePropertyKeys: ["person", "related_to"],
            disambiguation: "relationship_type", refuse: false,
            edges: [
                CatalogConnectEdge(propertyKey: "person", endpointTypeKey: "person"),
                CatalogConnectEdge(propertyKey: "related_to", endpointTypeKey: "person"),
            ]
        ),
        CatalogConnectRule(
            fromTypeKey: "event", toTypeKey: "place",
            bridgeTypeKey: "location", edgePropertyKeys: ["event", "place"],
            disambiguation: "none", refuse: false,
            edges: [
                CatalogConnectEdge(propertyKey: "event", endpointTypeKey: "event"),
                CatalogConnectEdge(propertyKey: "place", endpointTypeKey: "place"),
            ]
        ),
        CatalogConnectRule(
            fromTypeKey: "place", toTypeKey: "event",
            bridgeTypeKey: "location", edgePropertyKeys: ["event", "place"],
            disambiguation: "none", refuse: false,
            edges: [
                CatalogConnectEdge(propertyKey: "event", endpointTypeKey: "event"),
                CatalogConnectEdge(propertyKey: "place", endpointTypeKey: "place"),
            ]
        ),
        CatalogConnectRule(
            fromTypeKey: "place", toTypeKey: "place",
            bridgeTypeKey: "place_relationship", edgePropertyKeys: ["from", "to"],
            disambiguation: "place_relationship_type", refuse: false,
            edges: [
                CatalogConnectEdge(propertyKey: "from", endpointTypeKey: "place"),
                CatalogConnectEdge(propertyKey: "to", endpointTypeKey: "place"),
            ]
        ),
        CatalogConnectRule(
            fromTypeKey: "person", toTypeKey: "place",
            bridgeTypeKey: "", edgePropertyKeys: [], disambiguation: "", refuse: true
        ),
        CatalogConnectRule(
            fromTypeKey: "place", toTypeKey: "person",
            bridgeTypeKey: "", edgePropertyKeys: [], disambiguation: "", refuse: true
        ),
        CatalogConnectRule(
            fromTypeKey: "event", toTypeKey: "event",
            bridgeTypeKey: "", edgePropertyKeys: [], disambiguation: "", refuse: true
        ),
    ]

    static func match(from fromTypeKey: String, to toTypeKey: String, in rules: [CatalogConnectRule]) -> CatalogConnectRule {
        rules.first(where: { $0.fromTypeKey == fromTypeKey && $0.toTypeKey == toTypeKey })
            ?? CatalogConnectRule(
                fromTypeKey: fromTypeKey,
                toTypeKey: toTypeKey,
                bridgeTypeKey: "",
                edgePropertyKeys: [],
                disambiguation: "",
                refuse: true
            )
    }
}

struct CatalogCredibilityAssessment: Sendable, Equatable {
    var id: String
    var sourceID: String
    var gradeID: String
    var gradeKey: String
    var gradeLabel: String
    var argument: String
}

struct CatalogSourceType: Sendable, Equatable, Identifiable {
    var id: String
    var key: String
    var origin: String
    var label: String
    var description: String
    /// Closed design-system `type_*` key for evidence representation.
    var iconKey: String = PVMarkKey.defaultTypeMark.rawValue
    /// How many sources classify as this type. Deleting is only allowed at
    /// 0 — the engine refuses otherwise (`sourcetypes.in_use`). Only
    /// `listSourceTypes` and `updateSourceType` populate it.
    var usedBy: Int = 0
    /// How many metadata fields this type suggests — the list's third
    /// column, so browsing never fetches every type's join rows. Only
    /// `listSourceTypes` and `updateSourceType` populate it.
    var suggestedFieldCount: Int = 0
}

/// One `source_type_metadata_fields` join row: a field a type suggests, in
/// its stored order. Suggestions are not a schema — a source of the type may
/// leave any of them blank.
struct CatalogTypeSuggestion: Sendable, Equatable, Identifiable {
    var field: CatalogMetadataField
    var sortOrder: Int

    var id: String { field.id }
}

struct CatalogMetadataField: Sendable, Equatable, Identifiable {
    var id: String
    var key: String
    var origin: String
    var label: String
    var dataType: String
    var description: String
    /// How many sources already carry a value for this field. Deleting is
    /// only allowed at 0 — the engine refuses otherwise
    /// (`metadatafields.in_use`). Only `listMetadataFields` and
    /// `updateMetadataField` populate it.
    var usedBy: Int = 0
}

struct CatalogMetadataEntry: Sendable, Equatable, Identifiable {
    var id: String { field.id }
    var field: CatalogMetadataField
    var valueText: String
    var hasValue: Bool
    var suggested: Bool
    var sortOrder: Int32
}

struct CatalogSourceWorkspace: Sendable, Equatable {
    var source: CatalogSource
    var notes: [CatalogSourceNote]
    var metadata: [CatalogMetadataEntry]
    var artifacts: [CatalogArtifact]
    /// Nil when no assessment row (UI may display Standard without a row).
    var credibility: CatalogCredibilityAssessment?
}

struct CatalogDateValueInput: Sendable, Equatable {
    var kind: String
    var qualifier: String = ""
    var calendar: String = "gregorian"
    var startYear: Int32?
    var startMonth: Int32?
    var startDay: Int32?
    var startHour: Int32?
    var startMinute: Int32?
    var startSecond: Int32?
    var startMillisecond: Int32?
    var startTZ: String = ""
    var endYear: Int32?
    var endMonth: Int32?
    var endDay: Int32?
    var endHour: Int32?
    var endMinute: Int32?
    var endSecond: Int32?
    var endMillisecond: Int32?
    var endTZ: String = ""
    var phrase: String = ""
}

/// Origin split returned by `GetWorkspaceNavCounts` for vocabulary
/// destinations. `total` is always `seeded + user + plugin`.
struct WorkspaceNavOriginCounts: Sendable, Equatable {
    var total: Int
    var seeded: Int
    var user: Int
    var plugin: Int

    static let zero = WorkspaceNavOriginCounts(total: 0, seeded: 0, user: 0, plugin: 0)
}

/// Aggregate workspace chrome counts from one catalog open.
struct WorkspaceNavCounts: Sendable, Equatable {
    var sources: Int
    var sourceTypes: WorkspaceNavOriginCounts
    var metadataFields: WorkspaceNavOriginCounts
    /// Unmerged Person, Event, and Place handles.
    var persons: Int = 0
    var events: Int = 0
    var places: Int = 0
}

protocol GenealogyStore: Sendable {
    func installIdentity(identityDir: String) async throws -> InstallIdentity?
    func completeOnboarding(
        identityDir: String,
        parentDir: String,
        displayName: String,
        familyName: String
    ) async throws -> OnboardingResult
    func signOut(identityDir: String) async throws
    func activeProject(identityDir: String) async throws -> String?
    func listProjectUsers(projectDir: String) async throws -> [InstallIdentity]
    func openProject(
        identityDir: String,
        projectDir: String,
        displayName: String,
        adoptUserID: String
    ) async throws -> OnboardingResult
    func removeActiveProject(identityDir: String) async throws
    /// Releases the held exclusive catalog session for `projectDir` (workspace leave).
    func closeCatalogSession(projectDir: String) async throws
    func projectInfo(projectDir: String) async throws -> ProjectInfo

    func listSources(projectDir: String) async throws -> [CatalogSource]
    func getSourceWorkspace(projectDir: String, sourceID: String) async throws -> CatalogSourceWorkspace
    func createSource(
        projectDir: String,
        userID: String,
        sourceTypeID: String,
        title: String,
        description: String
    ) async throws -> CatalogSource
    func deleteSource(projectDir: String, userID: String, sourceID: String) async throws
    func updateSource(
        projectDir: String,
        userID: String,
        sourceID: String,
        sourceTypeID: String,
        title: String,
        description: String
    ) async throws -> CatalogSource
    /// Pin a raster Artifact as cover, or revert to the Source type icon (`type_icon`).
    /// Non-image Files (PDF, …) cannot be cover.
    func setSourceCover(
        projectDir: String,
        userID: String,
        sourceID: String,
        coverMode: String,
        primaryArtifactID: String
    ) async throws -> CatalogSource
    func addSourceNote(projectDir: String, userID: String, sourceID: String, body: String) async throws
        -> CatalogSourceNote
    func updateSourceNote(projectDir: String, userID: String, noteID: String, body: String) async throws
        -> CatalogSourceNote
    func deleteSourceNote(projectDir: String, userID: String, noteID: String) async throws
    /// Returns the refreshed workspace entry so callers can patch without refetching.
    func setSourceMetadata(
        projectDir: String,
        userID: String,
        sourceID: String,
        fieldID: String,
        valueText: String
    ) async throws -> CatalogMetadataEntry
    func clearSourceMetadata(projectDir: String, userID: String, sourceID: String, fieldID: String) async throws
    /// Permanently dismiss an unfilled type suggestion for this Source.
    /// Returns the updated workspace metadata list.
    func dismissSourceMetadataSuggestion(
        projectDir: String,
        userID: String,
        sourceID: String,
        fieldID: String
    ) async throws -> [CatalogMetadataEntry]
    /// Persist display order for visible metadata rows (field ids in order).
    func reorderSourceMetadata(
        projectDir: String,
        userID: String,
        sourceID: String,
        fieldIDs: [String]
    ) async throws -> [CatalogMetadataEntry]
    func createArtifact(
        projectDir: String,
        userID: String,
        sourceID: String,
        fileID: String,
        label: String,
        description: String
    ) async throws -> CatalogArtifact
    func deleteArtifact(projectDir: String, userID: String, artifactID: String) async throws
    func updateArtifact(
        projectDir: String,
        userID: String,
        artifactID: String,
        label: String,
        description: String
    ) async throws -> CatalogArtifact
    func ingestArtifactFile(
        projectDir: String,
        userID: String,
        artifactID: String,
        path: String
    ) async throws -> (artifact: CatalogArtifact, file: CatalogFileRef, reused: Bool)
    func listSourceCredibilityGrades(projectDir: String) async throws -> [CatalogCredibilityGrade]
    func upsertSourceCredibilityAssessment(
        projectDir: String,
        userID: String,
        sourceID: String,
        gradeID: String,
        argument: String
    ) async throws -> CatalogCredibilityAssessment
    func listSourceTypes(projectDir: String) async throws -> [CatalogSourceType]
    /// `key` is never accepted from the caller — the engine mints it as a
    /// kebab-case slug of `label`, the same rule `createMetadataField` uses.
    func createSourceType(
        projectDir: String,
        userID: String,
        label: String,
        description: String,
        iconKey: String
    ) async throws -> CatalogSourceType
    /// Patches label, description, and icon for a `user` or `provenencia` type.
    /// The key never changes here, so a rename keeps existing sources
    /// attached. Fails for `plugin:…` rows.
    func updateSourceType(
        projectDir: String,
        userID: String,
        typeID: String,
        label: String,
        description: String,
        iconKey: String
    ) async throws -> CatalogSourceType
    /// Deletes a type no source refers to. Suggestion joins cascade; the
    /// fields they named stay in the vocabulary.
    func deleteSourceType(
        projectDir: String,
        userID: String,
        typeID: String
    ) async throws
    func listTypeSuggestions(projectDir: String, typeID: String) async throws -> [CatalogTypeSuggestion]
    /// Attaches an existing field to a type at the end of its order.
    /// Assigning a field the type already suggests is a no-op. Returns the
    /// type's whole suggestion list so the caller never re-derives order.
    func assignTypeField(
        projectDir: String,
        userID: String,
        typeID: String,
        fieldID: String
    ) async throws -> [CatalogTypeSuggestion]
    /// Detaches a field from a type — the join only. The field stays in the
    /// vocabulary and sources already carrying a value for it keep it.
    func removeTypeField(
        projectDir: String,
        userID: String,
        typeID: String,
        fieldID: String
    ) async throws -> [CatalogTypeSuggestion]
    func listMetadataFields(projectDir: String) async throws -> [CatalogMetadataField]
    /// `key` is never accepted from the caller — the engine mints it as a
    /// kebab-case slug of `label` (see `FieldSlug.kebab` for the client-side
    /// preview mirror).
    func createMetadataField(
        projectDir: String,
        userID: String,
        label: String,
        dataType: String,
        description: String
    ) async throws -> CatalogMetadataField
    /// Patches label, data type, and description for a `user`-origin field.
    /// The key never changes here. Fails for `provenencia`/`plugin:…` rows.
    func updateMetadataField(
        projectDir: String,
        userID: String,
        fieldID: String,
        label: String,
        dataType: String,
        description: String
    ) async throws -> CatalogMetadataField

    func deleteMetadataField(
        projectDir: String,
        userID: String,
        fieldID: String
    ) async throws
    /// One catalog open: sidebar / vocabulary-header totals for sources,
    /// types, and fields.
    func workspaceNavCounts(projectDir: String) async throws -> WorkspaceNavCounts
    /// Catalog search (S3-07+). Empty / whitespace query → empty hits. `kinds`
    /// restricts hits to those kinds; empty is the omnibar's default set, which
    /// leaves out `person` / `event` / `place` until S9-35.
    func searchCatalog(
        projectDir: String,
        query: String,
        location: WorkspaceLocation,
        kinds: [String]
    ) async throws -> [CatalogSearchHit]

    func listSubjectTypes(projectDir: String) async throws -> [CatalogSubjectType]
    func createSubject(
        projectDir: String,
        userID: String,
        sourceID: String,
        subjectTypeID: String,
        label: String,
        description: String,
        placement: CatalogGridCell?
    ) async throws -> CatalogSubject
    func updateSubject(
        projectDir: String,
        userID: String,
        subjectID: String,
        label: String,
        description: String
    ) async throws -> CatalogSubject
    func deleteSubject(projectDir: String, userID: String, subjectID: String) async throws
    /// File an accepted claim for the Subject (one transaction): onto a new handle of its type when
    /// `entityID` is `nil`, else onto that existing handle (same type, unmerged; claim only).
    func promoteSubject(
        projectDir: String,
        userID: String,
        subjectID: String,
        entityID: String?,
        confidenceGradeID: String?,
        argument: String
    ) async throws -> CatalogPromoteResult
    /// Existing handles the Subject could join, best first: same type, scored by the type's match profile.
    func listPromoteTargetSuggestions(
        projectDir: String,
        subjectID: String,
        limit: Int
    ) async throws -> [CatalogPromoteTargetSuggestion]
    /// The claim confidence scale, in order.
    func listClaimConfidenceGrades(projectDir: String) async throws -> [CatalogClaimConfidenceGrade]
    /// Accepted handle of every promoted Subject on one Source's Evidence graph.
    func listSubjectMemberships(projectDir: String, sourceID: String) async throws -> [CatalogSubjectMembership]
    /// The title of every Event Subject on one Source's Evidence graph, keyed by Subject id.
    func listSourceEventTitles(projectDir: String, sourceID: String) async throws -> [String: CatalogEventTitle]
    /// Every unmerged Person as a row header, in Go's list order (by shown title, then ref-only by ref).
    func listPersonHeaders(projectDir: String) async throws -> [CatalogPersonHeader]
    func listEventHeaders(projectDir: String) async throws -> [CatalogEventHeader]
    func listPlaceHeaders(projectDir: String) async throws -> [CatalogPlaceHeader]
    /// One handle's fields, auto-reconciled values and outcomes. Throws
    /// `conclusiondetails.not_found` for an unknown or merged handle.
    func getConclusionDetail(projectDir: String, entityID: String) async throws -> CatalogConclusionDetail
    func listSubjects(projectDir: String, sourceID: String) async throws -> [CatalogSubject]
    func setSubjectPosition(
        projectDir: String,
        subjectID: String,
        gridX: Int64,
        gridY: Int64
    ) async throws -> CatalogSubjectPosition
    func clearSubjectPosition(projectDir: String, subjectID: String) async throws
    func listSubjectPositions(projectDir: String, sourceID: String) async throws -> [CatalogSubjectPosition]

    func createProperty(
        projectDir: String,
        userID: String,
        label: String,
        valueType: String,
        description: String,
        cardinality: String
    ) async throws -> CatalogProperty
    func updateProperty(
        projectDir: String,
        userID: String,
        propertyID: String,
        label: String,
        valueType: String,
        description: String,
        cardinality: String
    ) async throws -> CatalogProperty
    func deleteProperty(projectDir: String, userID: String, propertyID: String) async throws
    func listPropertyTerms(projectDir: String, propertyID: String) async throws -> [CatalogPropertyTerm]
    func createPropertyTerm(
        projectDir: String,
        userID: String,
        propertyID: String,
        label: String,
        description: String
    ) async throws -> CatalogPropertyTerm
    func assignSubjectTypeProperty(
        projectDir: String,
        userID: String,
        subjectTypeID: String,
        propertyID: String
    ) async throws
    func removeSubjectTypeProperty(
        projectDir: String,
        userID: String,
        subjectTypeID: String,
        propertyID: String
    ) async throws
    func listConnectRules() async throws -> [CatalogConnectRule]

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
    ) async throws -> (CatalogSubject, CatalogCitation, [CatalogObservation])

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
    ) async throws -> (CatalogCitation, [CatalogObservation])

    func getCitation(
        projectDir: String,
        citationID: String
    ) async throws -> (CatalogCitation, [String], [CatalogObservation])

    func updateCitation(
        projectDir: String,
        userID: String,
        citationID: String,
        locatorJSON: String,
        transcription: String,
        description: String,
        transcriptionUncertain: Bool,
        transcriptionNote: String
    ) async throws -> CatalogCitation

    func updateObservation(
        projectDir: String,
        userID: String,
        observation: CatalogObservation
    ) async throws -> CatalogObservation

    func deleteObservation(projectDir: String, userID: String, observationID: String) async throws
    func deleteCitation(projectDir: String, userID: String, citationID: String) async throws

    func getPropertiesWorkspace(projectDir: String) async throws -> PropertiesSnapshot

    func addObservationsToCitation(
        projectDir: String,
        userID: String,
        citationID: String,
        observations: [CatalogObservationDraft]
    ) async throws -> [CatalogObservation]

    func listObservationsBySource(projectDir: String, sourceID: String) async throws -> [CatalogObservation]

    func citationCountsBySource(projectDir: String, sourceID: String) async throws -> [String: Int]

    func listCitationsByArtifact(projectDir: String, artifactID: String) async throws -> [CatalogListedCitation]

    func listSourceGraphProgress(projectDir: String) async throws -> [SourceGraphProgress]

    func getSourceGraphProgress(projectDir: String, sourceID: String) async throws -> SourceGraphProgress

    func getDeleteImpact(projectDir: String, kind: String, id: String) async throws -> CatalogDeleteImpact
}

/// Per-Source Evidence-graph counts for the Sources list (S8-08).
struct SourceGraphProgress: Sendable, Equatable, Identifiable {
    var id: String { sourceId }
    var sourceId: String
    var subjectCount: Int
    var observationCount: Int

    static func zeros(sourceId: String) -> SourceGraphProgress {
        SourceGraphProgress(sourceId: sourceId, subjectCount: 0, observationCount: 0)
    }

    var isZero: Bool {
        subjectCount == 0 && observationCount == 0
    }
}

/// Draft payload for one Observation insert (FFI ObservationDraft).
struct CatalogObservationDraft: Sendable {
    var subjectID: String
    var propertyID: String
    var polarity: String = ""
    var valueText: String = ""
    var valueInteger: Int64?
    var date: CatalogDateValueInput?
    var valueDateID: String = ""
    var nameForm: String = ""
    var nameParts: [CatalogNameValuePart] = []
    var valueNameID: String = ""
    var valueSubjectID: String = ""
    var valueTermID: String = ""
    var notes: [String] = []
}

extension GenealogyStore {
    /// Promote onto a new handle with no grade or argument (the graph card's Confirm).
    func promoteSubject(projectDir: String, userID: String, subjectID: String) async throws -> CatalogPromoteResult {
        try await promoteSubject(
            projectDir: projectDir,
            userID: userID,
            subjectID: subjectID,
            entityID: nil,
            confidenceGradeID: nil,
            argument: ""
        )
    }
}

extension GenealogyStore {
    /// Omnibar search: the default kind set.
    func searchCatalog(projectDir: String, query: String, location: WorkspaceLocation) async throws -> [CatalogSearchHit] {
        try await searchCatalog(projectDir: projectDir, query: query, location: location, kinds: [])
    }
}
