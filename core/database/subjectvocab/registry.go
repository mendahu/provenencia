package subjectvocab

import (
	"github.com/mendahu/provenencia/core/connectrules"
	"github.com/mendahu/provenencia/core/propertymeta"
)

// Declarative provenencia Interpretation subject vocabulary (create-time starter).
// Presentation and non-bridge locked bindings live here. The connect matrix
// (bridge endpoints + disambiguation) lives in core/connectrules; Install
// derives bridge bindings from connectrules.BridgeBindings().
// Term capabilities (birthday facets, tree-edge roles, …) are deferred until a later PR.

const (
	RoleRoot        = "root"
	RoleBridge      = "bridge"
	RoleReification = "reification"

	DisambiguationNone                  = connectrules.DisambiguationNone
	DisambiguationRole                  = connectrules.DisambiguationRole
	DisambiguationRelationshipType      = connectrules.DisambiguationRelationshipType
	DisambiguationPlaceRelationshipType = connectrules.DisambiguationPlaceRelationshipType
)

type presentation struct {
	L10nKey       string
	IconSymbol    string
	InkToken      string
	TintToken     string
	ChipToken     string
	LineToken     string
	EdgeFromToken string
	EdgeToToken   string
}

type seedType struct {
	Key, Label, Description, RefPrefix, CandidateRefPrefix string
	Role                                                   string
	Placeable                                              bool
	PaletteSort                                            int
	RequiresCitationAtCreate                               bool
	Presentation                                           presentation
}

type seedProperty struct {
	Key, Label, Description, ValueType, Cardinality string
}

type seedBinding struct {
	TypeKey, PropertyKey string
	SortOrder            int
	Locked               bool
}

type seedTerm struct {
	PropertyKey, Key, Label, Description string
	Directed                             bool
	// Inverse is the term that reads this relationship from the other end.
	Inverse string
	// Neutral terms are no evidence either way (an "unknown" sex at birth).
	Neutral bool
}

var seedTypes = []seedType{
	{
		Key: "person", Label: "Person",
		Description: "A person represented by interpreted evidence.",
		RefPrefix:   "PER", CandidateRefPrefix: "CPR",
		Role: RoleRoot, Placeable: true, PaletteSort: 0,
		Presentation: presentation{
			L10nKey: "subjectType.person", IconSymbol: "subject_person",
			InkToken: "subjectPersonInk", TintToken: "subjectPersonTint",
			ChipToken: "subjectPersonChip", LineToken: "subjectPersonLine",
		},
	},
	{
		Key: "event", Label: "Event",
		Description: "An occurrence represented by interpreted evidence.",
		RefPrefix:   "EVT", CandidateRefPrefix: "CEV",
		Role: RoleRoot, Placeable: true, PaletteSort: 1,
		Presentation: presentation{
			L10nKey: "subjectType.event", IconSymbol: "subject_event",
			InkToken: "subjectEventInk", TintToken: "subjectEventTint",
			ChipToken: "subjectEventChip", LineToken: "subjectEventLine",
		},
	},
	{
		Key: "place", Label: "Place",
		Description: "A geographic feature at one grain (town, township, colony, farm, …).",
		RefPrefix:   "PLC", CandidateRefPrefix: "CPL",
		Role: RoleRoot, Placeable: true, PaletteSort: 2,
		Presentation: presentation{
			L10nKey: "subjectType.place", IconSymbol: "subject_place",
			InkToken: "subjectPlaceInk", TintToken: "subjectPlaceTint",
			ChipToken: "subjectPlaceChip", LineToken: "subjectPlaceLine",
		},
	},
	{
		Key: "relationship", Label: "Relationship",
		Description: "A general association when evidence is not a more specific event/context structure.",
		RefPrefix:   "REL", CandidateRefPrefix: "CRL",
		Role: RoleBridge, RequiresCitationAtCreate: true, PaletteSort: 3,
		Presentation: presentation{
			L10nKey: "subjectType.relationship", IconSymbol: "subject_relationship",
			InkToken: "subjectPersonInk", TintToken: "subjectPersonTint",
			ChipToken: "subjectPersonChip", LineToken: "subjectPersonLine",
		},
	},
	{
		Key: "participation", Label: "Participation",
		Description: "Association between a person and an event, including role.",
		RefPrefix:   "PTN", CandidateRefPrefix: "CPA",
		Role: RoleBridge, RequiresCitationAtCreate: true, PaletteSort: 4,
		Presentation: presentation{
			L10nKey: "subjectType.participation", IconSymbol: "subject_participation",
			InkToken: "subjectEventInk", TintToken: "subjectEventTint",
			ChipToken: "subjectEventChip", LineToken: "subjectEventLine",
		},
	},
	{
		Key: "location", Label: "Location",
		Description: "Association between an event and a place.",
		RefPrefix:   "LOC", CandidateRefPrefix: "CLO",
		Role: RoleBridge, RequiresCitationAtCreate: true, PaletteSort: 5,
		Presentation: presentation{
			L10nKey: "subjectType.location", IconSymbol: "subject_location",
			InkToken: "subjectPlaceInk", TintToken: "subjectPlaceTint",
			ChipToken: "subjectPlaceChip", LineToken: "subjectPlaceLine",
		},
	},
	{
		Key: "place_relationship", Label: "Place relationship",
		Description: "Association between two places: part-of or succession.",
		RefPrefix:   "PLR", CandidateRefPrefix: "CLR",
		Role: RoleBridge, RequiresCitationAtCreate: true, PaletteSort: 6,
		Presentation: presentation{
			L10nKey: "subjectType.place_relationship", IconSymbol: "subject_place_relationship",
			InkToken: "subjectPlaceInk", TintToken: "subjectPlaceTint",
			ChipToken: "subjectPlaceChip", LineToken: "subjectPlaceLine",
		},
	},
	{
		Key: "source", Label: "Source",
		Description: "A Source reified so other evidence can refer to or comment on it.",
		RefPrefix:   "SRN", CandidateRefPrefix: "CSR",
		Role: RoleReification, PaletteSort: 7,
		Presentation: presentation{
			L10nKey: "subjectType.source", IconSymbol: "subject_source",
			InkToken: "subjectPersonInk", TintToken: "subjectPersonTint",
			ChipToken: "subjectPersonChip", LineToken: "subjectPersonLine",
		},
	},
}

var seedProperties = []seedProperty{
	{Key: "name", Label: "Name", ValueType: propertymeta.ValueTypeName},
	{Key: "sex_at_birth", Label: "Sex at birth", Description: "Sex recorded or assigned at birth. Product term vocabulary.", ValueType: propertymeta.ValueTypeTerm},
	{Key: "gender", Label: "Gender", Description: "Gender identity. Separate from sex recorded or assigned at birth. Product term vocabulary.", ValueType: propertymeta.ValueTypeTerm},
	{Key: "event_type", Label: "Event type", Description: "Kind of event (birth, census, …). Product term vocabulary.", ValueType: propertymeta.ValueTypeTerm},
	{Key: "event_name", Label: "Event name", Description: "Recorded name of a historical event. Not a personal NameValue.", ValueType: propertymeta.ValueTypeText},
	{Key: "date", Label: "Date", Description: "Point-in-time when the event occurred (or the best single date when a span is unknown). Prefer this for births, deaths, and other one-day facts. Use start/end date instead when the event clearly lasts across a range.", ValueType: propertymeta.ValueTypeDate},
	{Key: "start_date", Label: "Start date", Description: "When a multi-day or open-ended event began (census day range, residence, military service, voyage). Leave empty for instantaneous events that only need Date.", ValueType: propertymeta.ValueTypeDate},
	{Key: "end_date", Label: "End date", Description: "When a spanned event ended or was last known. Pair with Start date; leave empty for instantaneous events that only need Date.", ValueType: propertymeta.ValueTypeDate},
	{Key: "role", Label: "Role", Description: "Participation role (subject, father, …). Product term vocabulary.", ValueType: propertymeta.ValueTypeTerm},
	{Key: "relationship_type", Label: "Relationship type", Description: "Directed kinship: person is this type of related_to. Spouse, sibling, and cousin are symmetric.", ValueType: propertymeta.ValueTypeTerm},
	{Key: "place_relationship_type", Label: "Place relationship type", Description: "Part of or succeeded by. Product-locked vocabulary.", ValueType: propertymeta.ValueTypeTerm},
	{Key: "from", Label: "From", Description: "Place relationship: the part, or the predecessor", ValueType: propertymeta.ValueTypeSubject},
	{Key: "to", Label: "To", Description: "Place relationship: the whole, or the successor", ValueType: propertymeta.ValueTypeSubject},
	{Key: "person", Label: "Person", Description: "Target hint: person (on a relationship: who is the typed relative)", ValueType: propertymeta.ValueTypeSubject},
	{Key: "event", Label: "Event", Description: "Target hint: event", ValueType: propertymeta.ValueTypeSubject},
	{Key: "place", Label: "Place", Description: "Target hint: place", ValueType: propertymeta.ValueTypeSubject},
	{Key: "related_to", Label: "Related to", Description: "Target hint: person (on a relationship: the other end — person is the X of related_to)", ValueType: propertymeta.ValueTypeSubject},
	{Key: "mentions", Label: "Mentions", Description: "Target hint: source", ValueType: propertymeta.ValueTypeSubject},
	{Key: "remark", Label: "Remark", Description: "Free-text commentary about a source subject", ValueType: propertymeta.ValueTypeText},
	{Key: "toponym", Label: "Toponym", Description: "Place name as interpreted from a Source", ValueType: propertymeta.ValueTypeText, Cardinality: propertymeta.CardinalityMultiple},
	{Key: "interests", Label: "Interests", Description: "Free-text interests. More than one value is allowed.", ValueType: propertymeta.ValueTypeText, Cardinality: propertymeta.CardinalityMultiple},
	{Key: "likeness", Label: "Likeness", Description: "A picture of this person: a region of a cited page and the image of it. More than one is allowed.", ValueType: propertymeta.ValueTypeDepiction, Cardinality: propertymeta.CardinalityMultiple},
	{Key: "religion", Label: "Religion", Description: "Religion or denomination as recorded. More than one value is allowed. Product term vocabulary; add your own terms.", ValueType: propertymeta.ValueTypeTerm, Cardinality: propertymeta.CardinalityMultiple},
	{Key: "languages_spoken", Label: "Languages spoken", Description: "Languages the person spoke. More than one value is allowed. Product term vocabulary; add your own terms.", ValueType: propertymeta.ValueTypeTerm, Cardinality: propertymeta.CardinalityMultiple},
}

// Non-bridge bindings from docs/seeded-vocabulary.md §3.3.
// Locked = the Properties page must not unbind it: Conclusion ordering
// (event dates), or a slot the screens key into (event_type, name, toponym).
// Bridge type bindings come from connectrules.BridgeBindings() at Install time.
var seedBindings = []seedBinding{
	{TypeKey: "person", PropertyKey: "name", SortOrder: 0, Locked: true},
	{TypeKey: "person", PropertyKey: "likeness", SortOrder: 1, Locked: true},
	{TypeKey: "person", PropertyKey: "sex_at_birth", SortOrder: 2},
	{TypeKey: "person", PropertyKey: "gender", SortOrder: 3},
	{TypeKey: "person", PropertyKey: "interests", SortOrder: 4},
	{TypeKey: "person", PropertyKey: "religion", SortOrder: 5},
	{TypeKey: "person", PropertyKey: "languages_spoken", SortOrder: 6},

	{TypeKey: "event", PropertyKey: "event_type", SortOrder: 0, Locked: true},
	{TypeKey: "event", PropertyKey: "event_name", SortOrder: 1},
	{TypeKey: "event", PropertyKey: "date", SortOrder: 2, Locked: true},
	{TypeKey: "event", PropertyKey: "start_date", SortOrder: 3, Locked: true},
	{TypeKey: "event", PropertyKey: "end_date", SortOrder: 4, Locked: true},

	{TypeKey: "place", PropertyKey: "toponym", SortOrder: 0, Locked: true},
	{TypeKey: "place", PropertyKey: "start_date", SortOrder: 1, Locked: true},
	{TypeKey: "place", PropertyKey: "end_date", SortOrder: 2, Locked: true},

	// Bridge endpoints + type come from connectrules; membership span is here.
	{TypeKey: "place_relationship", PropertyKey: "start_date", SortOrder: 3, Locked: true},
	{TypeKey: "place_relationship", PropertyKey: "end_date", SortOrder: 4, Locked: true},

	{TypeKey: "source", PropertyKey: "mentions", SortOrder: 0},
	{TypeKey: "source", PropertyKey: "remark", SortOrder: 1},
}

// Property terms from docs/seeded-vocabulary.md §3.4–3.11.
var seedTerms = []seedTerm{
	{PropertyKey: "sex_at_birth", Key: "female", Label: "Female"},
	{PropertyKey: "sex_at_birth", Key: "male", Label: "Male"},
	{PropertyKey: "sex_at_birth", Key: "intersex", Label: "Intersex"},
	{PropertyKey: "sex_at_birth", Key: "indeterminate", Label: "Indeterminate", Neutral: true},
	{PropertyKey: "sex_at_birth", Key: "unknown", Label: "Unknown", Neutral: true},

	{PropertyKey: "gender", Key: "man", Label: "Man"},
	{PropertyKey: "gender", Key: "woman", Label: "Woman"},
	{PropertyKey: "gender", Key: "non_binary", Label: "Non-binary"},

	{PropertyKey: "event_type", Key: "birth", Label: "Birth"},
	{PropertyKey: "event_type", Key: "death", Label: "Death"},
	{PropertyKey: "event_type", Key: "marriage", Label: "Marriage"},
	{PropertyKey: "event_type", Key: "engagement", Label: "Engagement"},
	{PropertyKey: "event_type", Key: "baptism", Label: "Baptism"},
	{PropertyKey: "event_type", Key: "burial", Label: "Burial"},
	{PropertyKey: "event_type", Key: "funeral", Label: "Funeral"},
	{PropertyKey: "event_type", Key: "census", Label: "Census Enumeration"},
	{PropertyKey: "event_type", Key: "residence", Label: "Residence"},
	{PropertyKey: "event_type", Key: "migration", Label: "Migration"},
	{PropertyKey: "event_type", Key: "travel", Label: "Travel"},

	{PropertyKey: "role", Key: "subject", Label: "Subject"},
	{PropertyKey: "role", Key: "father", Label: "Father"},
	{PropertyKey: "role", Key: "mother", Label: "Mother"},
	{PropertyKey: "role", Key: "child", Label: "Child"},
	{PropertyKey: "role", Key: "witness", Label: "Witness"},
	{PropertyKey: "role", Key: "informant", Label: "Informant"},
	{PropertyKey: "role", Key: "pallbearer", Label: "Pallbearer"},

	{PropertyKey: "relationship_type", Key: "spouse", Label: "Spouse"},
	{PropertyKey: "relationship_type", Key: "sibling", Label: "Sibling"},
	{PropertyKey: "relationship_type", Key: "cousin", Label: "Cousin"},
	{PropertyKey: "relationship_type", Key: "parent", Label: "Parent", Directed: true, Inverse: "child"},
	{PropertyKey: "relationship_type", Key: "child", Label: "Child", Directed: true, Inverse: "parent"},
	{PropertyKey: "relationship_type", Key: "grandparent", Label: "Grandparent", Directed: true, Inverse: "grandchild"},
	{PropertyKey: "relationship_type", Key: "grandchild", Label: "Grandchild", Directed: true, Inverse: "grandparent"},
	{PropertyKey: "relationship_type", Key: "pibling", Label: "Aunt / uncle", Directed: true, Inverse: "nibling"},
	{PropertyKey: "relationship_type", Key: "nibling", Label: "Niece / nephew", Directed: true, Inverse: "pibling"},
	{PropertyKey: "relationship_type", Key: "guardian", Label: "Guardian", Directed: true, Inverse: "ward"},
	{PropertyKey: "relationship_type", Key: "ward", Label: "Ward", Directed: true, Inverse: "guardian"},

	{PropertyKey: "place_relationship_type", Key: "part_of", Label: "Part of",
		Description: "From is part of to. Builds display chains.", Directed: true},
	{PropertyKey: "place_relationship_type", Key: "succeeded_by", Label: "Succeeded by",
		Description: "From was succeeded by to. Lineage; never a chain.", Directed: true},

	{PropertyKey: "religion", Key: "catholic", Label: "Catholic"},

	{PropertyKey: "languages_spoken", Key: "english", Label: "English"},
}
