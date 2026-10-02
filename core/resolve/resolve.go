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
// v1 clusters by exact equality (names by normalized form). Later steps
// replace pieces without changing callers: name and date auto-reconcilers
// (S9-13, S9-21) swap the clustering for those value types, provenance ranking
// (S9-14) goes ahead of support in the order, and subject values map to their
// handles (S9-28). Callers pass positive-polarity candidates only until S9-14.
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
}

// Cluster is one distinct value and the Observations that support it.
type Cluster struct {
	// Value is the cluster's representative: the lowest-id member's value,
	// or the concluded value when this is the concluded cluster.
	Value          Value
	ObservationIDs [][]byte // ascending
	Support        int      // len(ObservationIDs); 0 for a concluded value no candidate carries
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
// clusters: support descending, then each cluster's lowest Observation id.
// The order does not depend on the input order. A non-nil concluded value
// always takes rank 1: it joins the cluster it equals, or stands alone with
// support 0 ahead of the rest.
func Resolve(valueType string, candidates []Candidate, concluded *Value) (Result, error) {
	if !knownValueType(valueType) {
		return Result{}, fmt.Errorf("%w: %q", ErrUnknownValueType, valueType)
	}

	sorted := make([]Candidate, len(candidates))
	copy(sorted, candidates)
	sort.SliceStable(sorted, func(i, j int) bool {
		return bytes.Compare(sorted[i].ObservationID, sorted[j].ObservationID) < 0
	})

	var clusters []Cluster
	var keys []string
	index := map[string]int{}
	for _, c := range sorted {
		k, err := key(valueType, c.Value)
		if err != nil {
			return Result{}, err
		}
		i, ok := index[k]
		if !ok {
			i = len(clusters)
			index[k] = i
			keys = append(keys, k)
			clusters = append(clusters, Cluster{Value: c.Value})
		}
		clusters[i].ObservationIDs = append(clusters[i].ObservationIDs, c.ObservationID)
		clusters[i].Support++
	}

	// Clusters were created in ascending first-member id, so a stable sort on
	// support alone leaves ties ordered by lowest Observation id.
	order := make([]int, len(clusters))
	for i := range order {
		order[i] = i
	}
	sort.SliceStable(order, func(a, b int) bool {
		return clusters[order[a]].Support > clusters[order[b]].Support
	})
	res := Result{Clusters: make([]Cluster, 0, len(clusters)+1)}
	orderedKeys := make([]string, 0, len(clusters))
	for _, i := range order {
		res.Clusters = append(res.Clusters, clusters[i])
		orderedKeys = append(orderedKeys, keys[i])
	}

	if concluded == nil {
		return res, nil
	}
	k, err := key(valueType, *concluded)
	if err != nil {
		return Result{}, err
	}
	top := Cluster{Value: *concluded}
	for i, ck := range orderedKeys {
		if ck == k {
			top.ObservationIDs = res.Clusters[i].ObservationIDs
			top.Support = res.Clusters[i].Support
			res.Clusters = append(res.Clusters[:i], res.Clusters[i+1:]...)
			break
		}
	}
	res.Clusters = append([]Cluster{top}, res.Clusters...)
	res.Concluded = true
	return res, nil
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

// key is the v1 clustering rule: values with equal keys are one cluster.
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
	case properties.ValueTypeName:
		if v.Name == nil {
			return "", ErrValueMismatch
		}
		return normalizeForm(v.Name.Form), nil
	case properties.ValueTypeDate:
		if v.Date == nil {
			return "", ErrValueMismatch
		}
		return dateKey(v.Date), nil
	}
	return "", ErrUnknownValueType
}

// normalizeForm ignores case, punctuation, and whitespace differences.
func normalizeForm(form string) string {
	var b strings.Builder
	space := false
	for _, r := range form {
		switch {
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
