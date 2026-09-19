package subjectvocab

import (
	"github.com/mendahu/provenencia/core/database/properties"
)

// Declarative provenencia Interpretation subject vocabulary (create-time starter).
// Capabilities, presentation, locked bindings, and connect rules live here —
// Install writes only structural catalog rows (types, properties, bindings).

const (
	RoleRoot         = "root"
	RoleBridge       = "bridge"
	RoleReification  = "reification"

	DisambiguationNone               = "none"
	DisambiguationRole               = "role"
	DisambiguationRelationshipType   = "relationship_type"
	DisambiguationPersonPersonChoice = "person_person_choice"
)

type presentation struct {
	L10nKey      string
	IconSymbol   string
	InkToken     string
	TintToken    string
	ChipToken    string
	LineToken    string
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
	Key, Label, Description, ValueType string
}

type seedBinding struct {
	TypeKey, PropertyKey string
	SortOrder            int
	Locked               bool
}

type seedConnectRule struct {
	FromTypeKey, ToTypeKey string
	BridgeTypeKey          string
	EdgePropertyKeys       []string
	Disambiguation         string
	Refuse                 bool
}

var seedTypes = []seedType{
	{
		Key: "person", Label: "Person",
		Description: "A person represented by interpreted evidence.",
		RefPrefix:   "PER", CandidateRefPrefix: "CPR",
		Role: RoleRoot, Placeable: true, PaletteSort: 0,
		Presentation: presentation{
			L10nKey: "subjectType.person", IconSymbol: "person",
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
			L10nKey: "subjectType.event", IconSymbol: "event",
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
			L10nKey: "subjectType.place", IconSymbol: "place",
			InkToken: "subjectPlaceInk", TintToken: "subjectPlaceTint",
			ChipToken: "subjectPlaceChip", LineToken: "subjectPlaceLine",
		},
	},
	{
		Key: "relationship", Label: "Relationship",
		Description: "A general association when evidence is not a more specific event/context structure.",
		RefPrefix:   "REL", CandidateRefPrefix: "CRL",
		Role: RoleBridge, RequiresCitationAtCreate: true,
		Presentation: presentation{
			L10nKey: "subjectType.relationship", IconSymbol: "relationship",
			InkToken: "subjectPersonInk", TintToken: "subjectPersonTint",
			ChipToken: "subjectPersonChip", LineToken: "subjectPersonLine",
		},
	},
	{
		Key: "participation", Label: "Participation",
		Description: "Association between a person and an event, including role.",
		RefPrefix:   "PTN", CandidateRefPrefix: "CPA",
		Role: RoleBridge, RequiresCitationAtCreate: true,
		Presentation: presentation{
			L10nKey: "subjectType.participation", IconSymbol: "participation",
			InkToken: "subjectEventInk", TintToken: "subjectEventTint",
			ChipToken: "subjectEventChip", LineToken: "subjectEventLine",
		},
	},
	{
		Key: "location", Label: "Location",
		Description: "Association between an event and a place.",
		RefPrefix:   "LOC", CandidateRefPrefix: "CLO",
		Role: RoleBridge, RequiresCitationAtCreate: true,
		Presentation: presentation{
			L10nKey: "subjectType.location", IconSymbol: "location",
			InkToken: "subjectPlaceInk", TintToken: "subjectPlaceTint",
			ChipToken: "subjectPlaceChip", LineToken: "subjectPlaceLine",
		},
	},
	{
		Key: "source", Label: "Source",
		Description: "A Source reified so other evidence can refer to or comment on it.",
		RefPrefix:   "SRN", CandidateRefPrefix: "CSR",
		Role: RoleReification,
		Presentation: presentation{
			L10nKey: "subjectType.source", IconSymbol: "source",
			InkToken: "subjectPersonInk", TintToken: "subjectPersonTint",
			ChipToken: "subjectPersonChip", LineToken: "subjectPersonLine",
		},
	},
}

var seedProperties = []seedProperty{
	{Key: "event_type", Label: "Event type", ValueType: properties.ValueTypeText},
	{Key: "date", Label: "Date", Description: "Event date", ValueType: properties.ValueTypeDate},
	{Key: "role", Label: "Role", ValueType: properties.ValueTypeText},
	{Key: "relationship_type", Label: "Relationship type", ValueType: properties.ValueTypeText},
	{Key: "person", Label: "Person", Description: "Target hint: person", ValueType: properties.ValueTypeSubject},
	{Key: "event", Label: "Event", Description: "Target hint: event", ValueType: properties.ValueTypeSubject},
	{Key: "place", Label: "Place", Description: "Target hint: place", ValueType: properties.ValueTypeSubject},
	{Key: "participant", Label: "Participant", Description: "Target hint: person", ValueType: properties.ValueTypeSubject},
	{Key: "mentions", Label: "Mentions", Description: "Target hint: source", ValueType: properties.ValueTypeSubject},
	{Key: "remark", Label: "Remark", Description: "Free-text commentary about a source subject", ValueType: properties.ValueTypeText},
	{Key: "toponym", Label: "Toponym", Description: "Place name as interpreted from a Source", ValueType: properties.ValueTypeText},
}

// Bindings from docs/seeded-vocabulary.md §3.3. Locked = required for connect macros.
// person has no create-time Property bindings (researchers bind NameValue fields in S7-05).
var seedBindings = []seedBinding{
	{TypeKey: "event", PropertyKey: "event_type", SortOrder: 0},
	{TypeKey: "event", PropertyKey: "date", SortOrder: 1},

	{TypeKey: "place", PropertyKey: "toponym", SortOrder: 0},

	{TypeKey: "participation", PropertyKey: "person", SortOrder: 0, Locked: true},
	{TypeKey: "participation", PropertyKey: "event", SortOrder: 1, Locked: true},
	{TypeKey: "participation", PropertyKey: "role", SortOrder: 2},

	{TypeKey: "location", PropertyKey: "event", SortOrder: 0, Locked: true},
	{TypeKey: "location", PropertyKey: "place", SortOrder: 1, Locked: true},

	{TypeKey: "relationship", PropertyKey: "participant", SortOrder: 0, Locked: true},
	{TypeKey: "relationship", PropertyKey: "relationship_type", SortOrder: 1},

	{TypeKey: "source", PropertyKey: "mentions", SortOrder: 0},
	{TypeKey: "source", PropertyKey: "remark", SortOrder: 1},
}

// Connect matrix from interpretation-graph-ui.md §3.2. Omitted pairs refuse by default.
var seedConnect = []seedConnectRule{
	{
		FromTypeKey: "person", ToTypeKey: "event",
		BridgeTypeKey: "participation",
		EdgePropertyKeys: []string{"person", "event"},
		Disambiguation: DisambiguationRole,
	},
	{
		FromTypeKey: "event", ToTypeKey: "person",
		BridgeTypeKey: "participation",
		EdgePropertyKeys: []string{"person", "event"},
		Disambiguation: DisambiguationRole,
	},
	{
		FromTypeKey: "person", ToTypeKey: "person",
		BridgeTypeKey: "relationship",
		EdgePropertyKeys: []string{"participant", "participant"},
		Disambiguation: DisambiguationPersonPersonChoice,
	},
	{
		FromTypeKey: "event", ToTypeKey: "place",
		BridgeTypeKey: "location",
		EdgePropertyKeys: []string{"event", "place"},
		Disambiguation: DisambiguationNone,
	},
	{
		FromTypeKey: "place", ToTypeKey: "event",
		BridgeTypeKey: "location",
		EdgePropertyKeys: []string{"event", "place"},
		Disambiguation: DisambiguationNone,
	},
	{FromTypeKey: "person", ToTypeKey: "place", Refuse: true},
	{FromTypeKey: "place", ToTypeKey: "person", Refuse: true},
	{FromTypeKey: "event", ToTypeKey: "event", Refuse: true},
	{FromTypeKey: "place", ToTypeKey: "place", Refuse: true},
}
