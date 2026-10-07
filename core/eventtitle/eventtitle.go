// Package eventtitle chooses which rule of the Event naming matrix (Spike 9
// R4, Q10) titles an Event, and the parts that rule uses. It does not make
// text: the app fills an L10n template per rule, so the matrix stays
// localizable.
//
// Every surface that titles an Event chooses through Choose: canonical
// headers, Evidence graph cards, and later search documents. So a card,
// a list row, a page, and a search hit can't title the same Event
// differently.
package eventtitle

import (
	"strings"

	"github.com/mendahu/provenencia/core/database/namevalues"
)

// Rule is one row of the naming matrix, in precedence order.
type Rule string

const (
	// RuleRecordedName: the event_name Property (The Great Fire of 1849).
	RuleRecordedName Rule = "recorded_name"
	// RuleSubject: {Type} of {subject}.
	RuleSubject Rule = "subject"
	// RuleCouple: Marriage of {a} and {b}, for two subjects of a marriage.
	RuleCouple Rule = "couple"
	// RuleSubjects: {Type} of {first} et al.
	RuleSubjects Rule = "subjects"
	// RuleLabel: the researcher's working label.
	RuleLabel Rule = "label"
	// RuleTypeAtPlace: {Type} at {place}; Event at {place} with no type.
	RuleTypeAtPlace Rule = "type_at_place"
	// RuleType: Unspecified {type}.
	RuleType Rule = "type"
	// RuleRef: the handle or Subject ref.
	RuleRef Rule = "ref"
)

// MarriageTypeKey is the event_type whose two subjects are a couple.
const MarriageTypeKey = "marriage"

// Subject is one subject-role person, in stable order. A nil Name reads as
// "unnamed person".
type Subject struct {
	Name *namevalues.Value
}

// Parts are what an Event's title can be chosen from.
type Parts struct {
	RecordedName string
	Label        string
	Ref          string
	TypeKey      string
	TypeLabel    string
	Subjects     []Subject
	// Place is the first named place; chains are not part of a title.
	Place string
}

// Plan is the chosen rule and the parts, trimmed. The rule says which
// parts the template reads.
type Plan struct {
	Rule  Rule
	Parts Parts
}

// Choose applies the precedence: recorded name, then subjects, then label,
// then type and place, then ref.
func Choose(p Parts) Plan {
	p.RecordedName = strings.TrimSpace(p.RecordedName)
	p.Label = strings.TrimSpace(p.Label)
	p.Ref = strings.TrimSpace(p.Ref)
	p.TypeKey = strings.TrimSpace(p.TypeKey)
	p.TypeLabel = strings.TrimSpace(p.TypeLabel)
	p.Place = strings.TrimSpace(p.Place)
	switch {
	case p.RecordedName != "":
		return Plan{Rule: RuleRecordedName, Parts: p}
	case len(p.Subjects) == 1:
		return Plan{Rule: RuleSubject, Parts: p}
	case len(p.Subjects) == 2 && p.TypeKey == MarriageTypeKey:
		return Plan{Rule: RuleCouple, Parts: p}
	case len(p.Subjects) > 1:
		return Plan{Rule: RuleSubjects, Parts: p}
	case p.Label != "":
		return Plan{Rule: RuleLabel, Parts: p}
	case p.Place != "":
		return Plan{Rule: RuleTypeAtPlace, Parts: p}
	case p.TypeKey != "" || p.TypeLabel != "":
		return Plan{Rule: RuleType, Parts: p}
	default:
		return Plan{Rule: RuleRef, Parts: p}
	}
}
