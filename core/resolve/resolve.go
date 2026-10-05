// Package resolve reconciles a handle's candidate values for one Property
// into what the Conclusion layer displays (conclusion-reconciliation.md).
//
// Every Property on a canonical entity is a list: any member Subject may carry
// it more than once, and a handle has many members. Resolve takes that list
// and runs it through one shared pipeline (pipeline.go): admit, deny, group,
// majority, confidence. Each value type plugs in a small module (modules.go)
// that says when two values are the same, what folds into what, and how a
// group becomes one value.
//
// Nothing is dropped. Every distinct value comes back as a Cluster with a
// Reason: ReasonKept when it is displayed, else why not (outvoted, weak,
// denied, provisional). Every candidate comes back as an Outcome. The state
// (single / merged / mixed / concluded) is read off the displayed values.
//
// Pure: no catalog access, no SQL, no writes. The output is display policy and
// lives only in the derived resolved-values cache — never a claim, DateValue,
// or NameValue row (seeded-vocabulary §5.3).
//
// Name, date and subject use interim exact-key modules until their own
// modules land (S9-13b, S9-21, S9-28).
package resolve

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unicode"

	"github.com/mendahu/provenencia/core/database/datevalues"
	"github.com/mendahu/provenencia/core/database/namevalues"
	"github.com/mendahu/provenencia/core/database/properties"
)

// State describes how a Property's candidates resolved.
type State string

const (
	StateEmpty     State = ""          // nothing displayed, no concluded value
	StateSingle    State = "single"    // one displayed value from one Source
	StateMerged    State = "merged"    // one displayed value from several Sources
	StateMixed     State = "mixed"     // several displayed values
	StateConcluded State = "concluded" // a concluded value (Reconciliation Claim, later)
)

// Reason says what happened to a candidate, or why a value is or isn't
// displayed (conclusion-reconciliation.md §6). Stored in the cache.
type Reason string

const (
	ReasonKept        Reason = "kept"        // displayed, or supports a displayed value as given
	ReasonFolded      Reason = "folded"      // merged into a fuller value
	ReasonOutvoted    Reason = "outvoted"    // crowded out by a majority of Sources
	ReasonWeak        Reason = "weak"        // only weak evidence, and stronger evidence disagreed
	ReasonDenied      Reason = "denied"      // eliminated by a stronger negative Observation
	ReasonProvisional Reason = "provisional" // from a provisional member; never displayed
	ReasonNoEvidence  Reason = "no_evidence" // nothing usable to compare
	ReasonAgainst     Reason = "against"     // a negative that counted against a value
)

var (
	// ErrUnknownValueType is returned for a value type outside the closed
	// properties.ValueType* set.
	ErrUnknownValueType = errors.New("resolve: unknown value type")
	// ErrValueMismatch is returned when a candidate or concluded value does
	// not carry the field its value type needs.
	ErrValueMismatch = errors.New("resolve: value does not match value type")
)

// Value is one Property value. Exactly the field for the Property's
// value_type is set.
type Value struct {
	Text       string
	HasText    bool
	Integer    int64
	HasInteger bool
	TermID     []byte
	TermKey    string // the term's key, when known; NeutralTermKeys are no evidence
	SubjectID  []byte
	Date       *datevalues.Value
	Name       *namevalues.Value
}

// Candidate is one member Observation's value for the Property, with the
// evidence it carries.
type Candidate struct {
	ObservationID []byte // UUIDv7; the stable tiebreak
	Value         Value
	// SourceID is the Source the Observation's Citation is under. Majority
	// counts distinct Sources; an empty SourceID counts as its own Source.
	SourceID []byte
	// Negative is a negative-polarity Observation: never displayed; it
	// denies weaker same-value candidates and counts against the rest.
	Negative bool
	// Provisional is a candidate from a provisional member: never displayed.
	Provisional bool
	Provenance  Provenance
}

// Provenance is how strong a candidate's evidence is, each part relative to
// its vocabulary's default grade: 0 is the default, below it negative, above
// it positive.
type Provenance struct {
	Credibility     int  // the Citation's Source credibility vs `standard`
	Uncertain       bool // the Citation's transcription is uncertain
	ClaimConfidence int  // the member's Identity Claim confidence vs `moderate`
}

// Weak is evidence below the default anywhere: a low-trust Source, an
// uncertain transcription, or a low-confidence claim.
func (p Provenance) Weak() bool {
	return p.Credibility < 0 || p.Uncertain || p.ClaimConfidence < 0
}

// Stronger reports whether p is strictly stronger than q, comparing Source
// credibility, then transcription certainty, then claim confidence.
func (p Provenance) Stronger(q Provenance) bool {
	if p.Credibility != q.Credibility {
		return p.Credibility > q.Credibility
	}
	if p.Uncertain != q.Uncertain {
		return !p.Uncertain
	}
	return p.ClaimConfidence > q.ClaimConfidence
}

// Cluster is one distinct value and the Observations behind it.
type Cluster struct {
	// Value is the displayed value: assembled by the value type's module
	// (for simple types, the best-ranked member's value), or the concluded
	// value when this is the concluded cluster.
	Value          Value
	ObservationIDs [][]byte // members, ascending
	Support        int      // distinct Sources among the members; 0 for a concluded value no candidate carries
	Against        int      // negative candidates that match this value
	// Reason is ReasonKept for a displayed value, else why it isn't displayed.
	Reason Reason
}

// Displayed reports whether the cluster is one of the displayed values.
func (c Cluster) Displayed() bool { return c.Reason == ReasonKept }

// Outcome is what happened to one candidate.
type Outcome struct {
	ObservationID []byte
	Reason        Reason
	Cluster       int    // index into Result.Clusters; -1 for no_evidence and negatives
	DeniedBy      []byte // the negative that denied it, for ReasonDenied
}

// Result is the reconciler's output for one (handle, Property).
type Result struct {
	// Clusters holds every distinct value: displayed ones first, then the
	// rest, each group by support descending, then its best-ranked member.
	// Index 0 is rank 1.
	Clusters   []Cluster
	Candidates []Outcome // one per input candidate, ascending Observation id
	Concluded  bool      // Clusters[0] is the concluded value
}

// State reads the state off the displayed values.
func (r Result) State() State {
	if r.Concluded {
		return StateConcluded
	}
	var shown []Cluster
	for _, c := range r.Clusters {
		if c.Displayed() {
			shown = append(shown, c)
		}
	}
	switch {
	case len(shown) == 0:
		return StateEmpty
	case len(shown) > 1:
		return StateMixed
	case shown[0].Support > 1:
		return StateMerged
	default:
		return StateSingle
	}
}

// Resolve reconciles candidates for a Property of valueType. The result does
// not depend on the input order. A non-nil concluded value always takes rank
// 1: it joins the cluster with the same value, or stands alone with support 0
// ahead of the rest.
func Resolve(valueType string, candidates []Candidate, concluded *Value) (Result, error) {
	if !knownValueType(valueType) {
		return Result{}, fmt.Errorf("%w: %q", ErrUnknownValueType, valueType)
	}
	for _, c := range candidates {
		if !carries(valueType, c.Value) {
			return Result{}, ErrValueMismatch
		}
	}
	if concluded != nil && !carries(valueType, *concluded) {
		return Result{}, ErrValueMismatch
	}
	return reconcile(moduleFor(valueType), candidates, concluded, cardinalitySingle), nil
}

func knownValueType(vt string) bool {
	switch vt {
	case properties.ValueTypeText, properties.ValueTypeInteger, properties.ValueTypeDate,
		properties.ValueTypeName, properties.ValueTypeSubject, properties.ValueTypeTerm:
		return true
	default:
		return false
	}
}

// carries reports whether v sets the field its value type needs. An empty
// value of the right kind (blank text, a name with no parts) is carried and
// becomes no evidence in the pipeline.
func carries(valueType string, v Value) bool {
	switch valueType {
	case properties.ValueTypeText:
		return v.HasText
	case properties.ValueTypeInteger:
		return v.HasInteger
	case properties.ValueTypeTerm:
		return len(v.TermID) > 0
	case properties.ValueTypeSubject:
		return len(v.SubjectID) > 0
	case properties.ValueTypeName:
		return v.Name != nil
	case properties.ValueTypeDate:
		return v.Date != nil
	}
	return false
}

// SortKey is the text a cluster sorts by within its Property (the cache's
// sort_key): normalized form for names, case-folded text, and an
// order-preserving encoding for integers. Other value types have no sort key
// here (ok = false): dates sort by their window (S9-21), terms by label.
func SortKey(valueType string, v Value) (key string, ok bool) {
	switch valueType {
	case properties.ValueTypeName:
		if v.Name == nil {
			return "", false
		}
		return NormalizeForm(v.Name.Form), true
	case properties.ValueTypeText:
		if !v.HasText {
			return "", false
		}
		return strings.ToLower(strings.TrimSpace(v.Text)), true
	case properties.ValueTypeInteger:
		if !v.HasInteger {
			return "", false
		}
		// Offset into uint64 so byte order matches numeric order, negatives included.
		return fmt.Sprintf("%020d", uint64(v.Integer)^(1<<63)), true
	}
	return "", false
}

// NormalizeForm ignores case, punctuation, and whitespace differences: the
// name cluster key and sort key, and the text core/match tokenizes. Dashes
// and slashes separate words ("Smith-Jones" reads as "smith jones"); other
// punctuation is dropped, so an apostrophe joins ("O'Brien" reads as
// "obrien", the same as "OBrien").
func NormalizeForm(form string) string {
	var b strings.Builder
	space := false
	for _, r := range form {
		switch {
		case unicode.Is(unicode.Pd, r) || r == '/':
			space = b.Len() > 0
			continue
		case unicode.IsPunct(r):
			continue
		case unicode.IsSpace(r):
			space = b.Len() > 0
			continue
		}
		if space {
			b.WriteByte(' ')
			space = false
		}
		b.WriteRune(unicode.ToLower(r))
	}
	return b.String()
}

// dateKey matches on every structured field; containment and overlap are the
// date reconciler's job (S9-21).
func dateKey(d *datevalues.Value) string {
	parts := []string{
		d.Kind, d.Qualifier, d.Calendar,
		optInt(d.StartYear), optInt(d.StartMonth), optInt(d.StartDay),
		optInt(d.StartHour), optInt(d.StartMinute), optInt(d.StartSecond), optInt(d.StartMillisecond),
		d.StartTZ,
		optInt(d.EndYear), optInt(d.EndMonth), optInt(d.EndDay),
		optInt(d.EndHour), optInt(d.EndMinute), optInt(d.EndSecond), optInt(d.EndMillisecond),
		d.EndTZ,
		d.Phrase,
	}
	for i, p := range parts {
		parts[i] = strconv.Quote(p)
	}
	return strings.Join(parts, "|")
}

func optInt(p *int) string {
	if p == nil {
		return "-"
	}
	return strconv.Itoa(*p)
}
