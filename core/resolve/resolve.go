// Package resolve turns a handle's candidate values for one Property into
// ranked clusters (deployment-plan R2).
//
// Every Property on a canonical entity is a list: any member Subject may carry
// it more than once, and a handle has many members. Resolve takes that list —
// zero, one, or many candidates — and returns zero, one, or many clusters in
// display order. The state (single / merged / mixed / concluded) is read off
// the shape of the result, never stored beside it.
//
// Pure: no catalog access, no SQL, no writes. The output is display policy and
// lives only in the derived resolved-values cache — never a claim, DateValue,
// or NameValue row (seeded-vocabulary §5.3).
//
// Names go through the name auto-reconciler (names.go): structured parts by
// type, never form, with weak and denied candidates eliminated by their
// provenance. Every other value type clusters by exact equality and is not
// reconciled; its negative candidates only count against the cluster they
// match. Later steps replace pieces without changing callers: the date
// auto-reconciler (S9-21), a term reconciler, and subject values mapped to
// their handles (S9-28).
package resolve

import (
	"bytes"
	"errors"
	"fmt"
	"sort"
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
	StateEmpty     State = ""          // no candidates, no concluded value
	StateSingle    State = "single"    // one candidate
	StateMerged    State = "merged"    // several candidates, one cluster
	StateMixed     State = "mixed"     // several clusters
	StateConcluded State = "concluded" // a concluded value (Reconciliation Claim, later)
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
	SubjectID  []byte
	Date       *datevalues.Value
	Name       *namevalues.Value
}

// Candidate is one member Observation's value for the Property.
type Candidate struct {
	ObservationID []byte // UUIDv7; the stable tiebreak
	Value         Value
	// Negative is a negative-polarity Observation: it is never displayed,
	// and counts against the values it matches.
	Negative   bool
	Provenance Provenance
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

// Cluster is one distinct value and the Observations that support it.
type Cluster struct {
	// Value is the cluster's representative: the lowest-id member's value;
	// for names, the reconciled name (a member's own value when one carries
	// exactly the reconciled parts); or the concluded value when this is the
	// concluded cluster.
	Value          Value
	ObservationIDs [][]byte // ascending
	Support        int      // len(ObservationIDs); 0 for a concluded value no candidate carries
	Against        int      // negative candidates that match this value
}

// Result is the resolver's output for one (handle, Property).
type Result struct {
	Clusters  []Cluster // display order; index 0 is rank 1
	Concluded bool      // Clusters[0] is the concluded value
}

// State reads the state off the result's shape.
func (r Result) State() State {
	switch {
	case len(r.Clusters) == 0:
		return StateEmpty
	case r.Concluded:
		return StateConcluded
	case len(r.Clusters) > 1:
		return StateMixed
	case r.Clusters[0].Support > 1:
		return StateMerged
	default:
		return StateSingle
	}
}

// Resolve clusters candidates for a Property of valueType and orders the
// clusters: support descending, then each cluster's lowest Observation id
// (names: surviving clusters first, see names.go). The order does not depend
// on the input order. A non-nil concluded value always takes rank 1: it joins
// the cluster it equals, or stands alone with support 0 ahead of the rest.
func Resolve(valueType string, candidates []Candidate, concluded *Value) (Result, error) {
	if !knownValueType(valueType) {
		return Result{}, fmt.Errorf("%w: %q", ErrUnknownValueType, valueType)
	}

	sorted := make([]Candidate, len(candidates))
	copy(sorted, candidates)
	sort.SliceStable(sorted, func(i, j int) bool {
		return bytes.Compare(sorted[i].ObservationID, sorted[j].ObservationID) < 0
	})

	var positives, negatives []Candidate
	for _, c := range sorted {
		if c.Negative {
			negatives = append(negatives, c)
		} else {
			positives = append(positives, c)
		}
	}

	var clusters []Cluster
	if valueType == properties.ValueTypeName {
		for _, c := range sorted {
			if c.Value.Name == nil {
				return Result{}, ErrValueMismatch
			}
		}
		// Rank order: strongest provenance first, then id.
		sort.SliceStable(positives, func(i, j int) bool {
			return positives[i].Provenance.Stronger(positives[j].Provenance)
		})
		clusters = reconcileNames(positives, negatives)
	} else {
		var err error
		if clusters, err = clusterByKey(valueType, positives); err != nil {
			return Result{}, err
		}
		if err := countAgainst(valueType, clusters, negatives); err != nil {
			return Result{}, err
		}
	}
	res := Result{Clusters: clusters}

	if concluded == nil {
		return res, nil
	}
	same, err := matcher(valueType, *concluded)
	if err != nil {
		return Result{}, err
	}
	top := Cluster{Value: *concluded}
	for i, cl := range res.Clusters {
		if same(cl.Value) {
			top.ObservationIDs = cl.ObservationIDs
			top.Support = cl.Support
			top.Against = cl.Against
			res.Clusters = append(res.Clusters[:i:i], res.Clusters[i+1:]...)
			break
		}
	}
	res.Clusters = append([]Cluster{top}, res.Clusters...)
	res.Concluded = true
	return res, nil
}

// clusterByKey groups equal keys and orders the clusters by support, then
// lowest Observation id.
func clusterByKey(valueType string, sorted []Candidate) ([]Cluster, error) {
	var clusters []Cluster
	index := map[string]int{}
	for _, c := range sorted {
		k, err := key(valueType, c.Value)
		if err != nil {
			return nil, err
		}
		i, ok := index[k]
		if !ok {
			i = len(clusters)
			index[k] = i
			clusters = append(clusters, Cluster{Value: c.Value})
		}
		clusters[i].ObservationIDs = append(clusters[i].ObservationIDs, c.ObservationID)
		clusters[i].Support++
	}
	// Clusters were created in ascending first-member id, so a stable sort on
	// support alone leaves ties ordered by lowest Observation id.
	sort.SliceStable(clusters, func(a, b int) bool {
		return clusters[a].Support > clusters[b].Support
	})
	return clusters, nil
}

// countAgainst adds each negative to the cluster with its key. Values no
// positive candidate carries have no cluster to count against.
func countAgainst(valueType string, clusters []Cluster, negatives []Candidate) error {
	for _, n := range negatives {
		k, err := key(valueType, n.Value)
		if err != nil {
			return err
		}
		for i := range clusters {
			if ck, err := key(valueType, clusters[i].Value); err == nil && ck == k {
				clusters[i].Against++
				break
			}
		}
	}
	return nil
}

// matcher reports which cluster a concluded value belongs to: the same key,
// or for names the same parts by type.
func matcher(valueType string, concluded Value) (func(Value) bool, error) {
	if valueType == properties.ValueTypeName {
		if concluded.Name == nil {
			return nil, ErrValueMismatch
		}
		return func(v Value) bool { return sameName(concluded.Name, v.Name) }, nil
	}
	k, err := key(valueType, concluded)
	if err != nil {
		return nil, err
	}
	return func(v Value) bool {
		ck, err := key(valueType, v)
		return err == nil && ck == k
	}, nil
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

// key is the clustering rule for every value type but names: values with
// equal keys are one cluster.
func key(valueType string, v Value) (string, error) {
	switch valueType {
	case properties.ValueTypeText:
		if !v.HasText {
			return "", ErrValueMismatch
		}
		return strings.TrimSpace(v.Text), nil
	case properties.ValueTypeInteger:
		if !v.HasInteger {
			return "", ErrValueMismatch
		}
		return strconv.FormatInt(v.Integer, 10), nil
	case properties.ValueTypeTerm:
		if len(v.TermID) == 0 {
			return "", ErrValueMismatch
		}
		return string(v.TermID), nil
	case properties.ValueTypeSubject:
		if len(v.SubjectID) == 0 {
			return "", ErrValueMismatch
		}
		return string(v.SubjectID), nil
	case properties.ValueTypeDate:
		if v.Date == nil {
			return "", ErrValueMismatch
		}
		return dateKey(v.Date), nil
	}
	return "", ErrUnknownValueType
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

// NormalizeForm ignores case, punctuation, and whitespace differences: how
// the name reconciler compares one part, the name sort key, and the text
// core/match tokenizes. Dashes
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
