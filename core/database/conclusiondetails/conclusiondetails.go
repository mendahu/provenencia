// Package conclusiondetails composes the detail of one canonical handle from
// the catalog graph. For every Property its members' records speak to, it
// returns the state, every auto-reconciled value (displayed or not, with its
// reason), and the auto-reconciler's outcome for each Observation it
// considered, joined to the record's Source: the "Why" a detail page shows.
//
// Field values are the node's auto-reconciler ranks. Term labels come from
// the graph vocabulary. The Why stays a SQL read each time the page opens.
// A Property no
// record mentions isn't returned. Pictures are listed apart from the
// fields, earliest first (gallery.ForEntities). Which empty fields a page still shows is
// the page's choice. Go returns structures; the app formats text.
// Reconciliation Claims are never in these tables: when they ship, a
// composer lays an accepted claim over the auto-reconciled value here.
// A Person's life rows read other handles' fields through PersonDerivedRows,
// composed by the same steps.
package conclusiondetails

import (
	"bytes"
	"database/sql"
	"sort"

	"github.com/mendahu/provenencia/core/apperr"
	"github.com/mendahu/provenencia/core/autoreconcile"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/autoreconciler"
	"github.com/mendahu/provenencia/core/database/canonicalentities"
	"github.com/mendahu/provenencia/core/database/conclusionheaders"
	"github.com/mendahu/provenencia/core/database/datevalues"
	"github.com/mendahu/provenencia/core/database/gallery"
	"github.com/mendahu/provenencia/core/database/graphcache"
	"github.com/mendahu/provenencia/core/database/namevalues"
	"github.com/mendahu/provenencia/core/database/sqlbatch"
	"github.com/mendahu/provenencia/core/datevalue"
	"github.com/mendahu/provenencia/core/locator"
	"github.com/mendahu/provenencia/core/namevalue"
	"github.com/mendahu/provenencia/core/propertymeta"
	"github.com/mendahu/provenencia/core/valuecodec"
	"github.com/mendahu/provenencia/core/vocabkey"
)

// ErrNotFound is returned for an unknown or merged handle.
var ErrNotFound = apperr.New(apperr.CodeConclusionDetailsNotFound, apperr.KindUser)

// Querier is *sql.Tx or *sql.DB.
type Querier interface {
	Query(query string, args ...any) (*sql.Rows, error)
}

// Value is one Property value. Exactly the field for the value type is set.
type Value struct {
	Text       string
	HasText    bool
	Integer    int64
	HasInteger bool
	TermID     []byte
	TermKey    string
	TermLabel  string
	// TermOrigin is the term's origin: its key is the product's only when this is vocabkey.OriginProvenencia.
	TermOrigin string
	Date       *datevalue.Value
	Name       *namevalue.Value
}

// ReconciledValue is one auto-reconciled value of a Property.
type ReconciledValue struct {
	Rank    int // 1 = first displayed
	Value   Value
	Support int    // distinct Sources
	Against int    // negative records that match it
	Reason  string // "kept" when displayed, else why not
}

// Displayed reports whether the value is one of the displayed values.
func (v ReconciledValue) Displayed() bool { return v.Reason == string(autoreconcile.ReasonKept) }

// Outcome is what the auto-reconciler did with one Observation, with the
// evidence it weighed.
type Outcome struct {
	ObservationID  []byte
	ObservationRef string
	Reason         string
	ValueRank      int                // the value it went into; 0 for none
	GroupRank      int                // the displayed value it counts under; 0 for none
	DeniedBy       []byte             // the negative Observation, for "denied"
	Vote           autoreconcile.Vote // the majority that beat it, for "outvoted"
	Recorded       Value              // what the record said
	SubjectID      []byte
	SubjectRef     string
	GraphID        []byte // the Evidence graph of the Observation's Subject
	CitationID     []byte
	ArtifactID     []byte
	SourceID       []byte
	SourceTitle    string
	ArtifactLabel  string // "" = none
	CitationLabel  string // "" = none
	ArtifactPage   int    // the Citation locator's page; 0 = none
	PageLabel      string // its printed label; "" = none
	CredibilityKey string // "" = no assessment (standard)
	// Provenance is the evidence as the auto-reconciler weighed it, relative
	// to the default grades; Provenance.Weak says whether it is weak, and
	// each part says why.
	Provenance autoreconcile.Provenance
}

// Field is one Property of the handle.
type Field struct {
	PropertyID  []byte
	PropertyKey string
	Label       string
	ValueType   string
	State       autoreconcile.State // read off the displayed values
	cardinality string
	Values      []ReconciledValue // rank order
	Outcomes    []Outcome         // by group rank (none last), value rank (none last), then Observation id
}

// PictureTile is one picture a gallery shows: its image, and the outcome of
// its earliest gallery record, which the tile opens.
type PictureTile struct {
	ImageFileID []byte
	Record      Outcome
	Preferred   bool // the handle's chosen portrait or thumbnail (S10-39)
	Leads       bool // the picture the header leads with (gallery.Picture.Leads)
}

// Detail is one handle's detail.
type Detail struct {
	Entity      canonicalentities.Entity
	MemberCount int // accepted members
	Fields      []Field
	Pictures    []PictureTile // earliest recorded first
}

const (
	// Property ids this handle's outcomes name. Labels, value types and the
	// kind's binding order come from the graph vocabulary.
	sqlPropertyIDs = `SELECT DISTINCT ao.entity_id, ao.property_id FROM auto_reconciler_outcomes ao WHERE ao.entity_id IN (`

	sqlOutcomes = `SELECT ao.entity_id, ao.property_id, ao.observation_id, o.ref, ao.reason, ao.value_rank, ao.denied_by,
			COALESCE(ao.vote_support, 0), COALESCE(ao.vote_total, 0),
			o.value_text, o.value_integer, o.value_term_id, COALESCE(t.key, ''), COALESCE(t.label, ''), COALESCE(t.origin, ''),
			o.value_date_id, o.value_name_id,
			s.id, s.ref, s.graph_id, c.id, a.id, src.id, COALESCE(src.title, ''),
			COALESCE(a.label, ''), COALESCE(c.label, ''), c.locator_json,
			COALESCE(sg.key, ''), ` + autoreconciler.ProvenanceColumns + `
		FROM auto_reconciler_outcomes ao
		JOIN observations o ON o.id = ao.observation_id
		JOIN subjects s ON s.id = o.subject_id
		JOIN citations c ON c.id = o.citation_id
		JOIN artifacts a ON a.id = c.artifact_id
		JOIN sources src ON src.id = a.source_id
		LEFT JOIN property_terms t ON t.id = o.value_term_id
		` + autoreconciler.ProvenanceJoins + `
		WHERE ao.entity_id IN (`
)

// ForEntity composes the detail of one unmerged handle.
func ForEntity(c *database.Catalog, entityID []byte) (Detail, error) {
	if c == nil || c.Graph() == nil || len(entityID) != 16 {
		return Detail{}, ErrNotFound
	}
	n, err := c.Graph().Node(entityID)
	if err != nil {
		return Detail{}, err
	}
	if n == nil || n.Merged {
		return Detail{}, ErrNotFound
	}
	db, err := c.DB()
	if err != nil {
		return Detail{}, err
	}
	d := Detail{Entity: entityOf(n), MemberCount: acceptedMembers(n)}
	tiles, err := gallery.ForEntities(db, [][]byte{entityID})
	if err != nil {
		return Detail{}, err
	}
	leads := map[string]bool{}
	for _, l := range tiles[string(entityID)] {
		leads[string(l.ObservationIDs[0])] = true
	}
	fields, leadOutcomes, err := fieldsOf(c.Graph(), db, []*graphcache.Node{n}, leads)
	if err != nil {
		return Detail{}, err
	}
	d.Fields = fields[string(entityID)]
	for _, l := range tiles[string(entityID)] {
		if o, ok := leadOutcomes[string(l.ObservationIDs[0])]; ok {
			d.Pictures = append(d.Pictures, PictureTile{ImageFileID: l.ImageFileID, Record: o, Preferred: l.Preferred, Leads: l.Leads})
		}
	}
	return d, nil
}

// fieldsOf composes the fields of several handles at once, keyed by entity
// id: the one composition every page's fields come from. A nil or merged
// node, or a repeated id, is skipped. It also returns the outcomes of the
// Observations in leads (a gallery's earliest records), which belong to no
// field.
func fieldsOf(g *graphcache.Graph, q Querier, nodes []*graphcache.Node, leads map[string]bool) (map[string][]Field, map[string]Outcome, error) {
	var use []*graphcache.Node
	var ids [][]byte
	seen := map[string]bool{}
	for _, n := range nodes {
		if n == nil || n.Merged || seen[string(n.ID)] {
			continue
		}
		seen[string(n.ID)] = true
		use = append(use, n)
		ids = append(ids, n.ID)
	}
	out := map[string][]Field{}
	if len(use) == 0 {
		return out, map[string]Outcome{}, nil
	}
	ordered, byProp, err := loadProperties(g, q, use)
	if err != nil {
		return nil, nil, err
	}
	for _, n := range use {
		if err := attachValues(g, n, byProp[string(n.ID)]); err != nil {
			return nil, nil, err
		}
	}
	leadOutcomes, err := loadOutcomes(q, ids, byProp, leads)
	if err != nil {
		return nil, nil, err
	}
	for id, fs := range ordered {
		for _, f := range fs {
			finish(f)
			out[id] = append(out[id], *f)
		}
	}
	return out, leadOutcomes, nil
}

// finish reads a field's state off its values and orders its outcomes.
func finish(f *Field) {
	f.State = state(f.Values, f.cardinality)
	displayed := map[int]bool{}
	for _, v := range f.Values {
		if v.Displayed() {
			displayed[v.Rank] = true
		}
	}
	for i := range f.Outcomes {
		o := &f.Outcomes[i]
		o.GroupRank = 0
		counts := o.Reason == string(autoreconcile.ReasonKept) || o.Reason == string(autoreconcile.ReasonFolded)
		if counts && displayed[o.ValueRank] {
			o.GroupRank = o.ValueRank
		}
	}
	less := func(a, b int) (decided, before bool) {
		if (a == 0) != (b == 0) {
			return true, b == 0
		}
		if a != b {
			return true, a < b
		}
		return false, false
	}
	sort.SliceStable(f.Outcomes, func(i, j int) bool {
		if d, b := less(f.Outcomes[i].GroupRank, f.Outcomes[j].GroupRank); d {
			return b
		}
		if d, b := less(f.Outcomes[i].ValueRank, f.Outcomes[j].ValueRank); d {
			return b
		}
		return string(f.Outcomes[i].ObservationID) < string(f.Outcomes[j].ObservationID)
	})
}

func entityOf(n *graphcache.Node) canonicalentities.Entity {
	return canonicalentities.Entity{
		ID:            append([]byte(nil), n.ID...),
		SubjectTypeID: append([]byte(nil), n.SubjectTypeID...),
		Ref:           n.Ref,
		Argument:      n.Argument,
		Label:         n.Label,
	}
}

func acceptedMembers(n *graphcache.Node) int {
	var nAccepted int
	for _, m := range n.Members {
		if m.Accepted {
			nAccepted++
		}
	}
	return nAccepted
}

// state reads a field's state off its displayed values, by the same rule as
// autoreconcile.Result.State. Several displayed values on a multiple Property
// are all true, so the state is multiple rather than mixed.
func state(values []ReconciledValue, cardinality string) autoreconcile.State {
	var shown []ReconciledValue
	for _, v := range values {
		if v.Displayed() {
			shown = append(shown, v)
		}
	}
	switch {
	case len(shown) == 0:
		return autoreconcile.StateEmpty
	case len(shown) > 1:
		if cardinality == propertymeta.CardinalityMultiple {
			return autoreconcile.StateMultiple
		}
		return autoreconcile.StateMixed
	case shown[0].Support > 1:
		return autoreconcile.StateMerged
	default:
		return autoreconcile.StateSingle
	}
}

// loadProperties lists each node's fields in its kind's binding order, and
// the same fields by entity id then property id.
func loadProperties(g *graphcache.Graph, q Querier, nodes []*graphcache.Node) (map[string][]*Field, map[string]map[string]*Field, error) {
	v, err := g.Vocabulary()
	if err != nil {
		return nil, nil, err
	}
	props := v.PropertiesByID()
	ids := make([][]byte, 0, len(nodes))
	typeOf := map[string][]byte{}
	for _, n := range nodes {
		ids = append(ids, n.ID)
		typeOf[string(n.ID)] = n.SubjectTypeID
	}
	ordered := map[string][]*Field{}
	err = sqlbatch.Query(q, sqlPropertyIDs, ")", nil, ids, func(rows *sql.Rows) error {
		var entityID, id []byte
		if err := rows.Scan(&entityID, &id); err != nil {
			return err
		}
		p, ok := props[string(id)]
		if !ok || p.ValueType == propertymeta.ValueTypeSubject || p.ValueType == propertymeta.ValueTypeDepiction {
			return nil
		}
		ordered[string(entityID)] = append(ordered[string(entityID)], &Field{
			PropertyID: append([]byte(nil), p.ID...), PropertyKey: p.Key, Label: p.Label,
			ValueType: p.ValueType, cardinality: p.Cardinality,
		})
		return nil
	})
	if err != nil {
		return nil, nil, err
	}
	byProp := map[string]map[string]*Field{}
	for id, got := range ordered {
		order := v.TypeOrder(typeOf[id])
		sort.SliceStable(got, func(i, j int) bool {
			return vocabkey.CompareByOrder(order, vocabkey.PropertyKey(got[i].PropertyKey), vocabkey.PropertyKey(got[j].PropertyKey)) < 0
		})
		byProp[id] = map[string]*Field{}
		for _, f := range got {
			byProp[id][string(f.PropertyID)] = f
		}
	}
	return ordered, byProp, nil
}

func attachValues(g *graphcache.Graph, n *graphcache.Node, byProp map[string]*Field) error {
	vocab, err := g.Vocabulary()
	if err != nil {
		return err
	}
	terms := vocab.TermsByID()
	for prop, ranks := range n.Values {
		f, ok := byProp[prop]
		if !ok {
			continue // a subject-valued or depiction Property, or one the outcomes no longer name
		}
		for _, rank := range ranks {
			v := ReconciledValue{
				Rank: rank.Rank, Reason: rank.Reason,
				Support: rank.Support, Against: rank.Against,
			}
			v.Value.Text, v.Value.HasText = rank.Text, rank.HasText
			v.Value.Integer, v.Value.HasInteger = rank.Integer, rank.HasInteger
			v.Value.TermID = append([]byte(nil), rank.TermID...)
			if name, ok := terms[string(rank.TermID)]; ok {
				v.Value.TermKey, v.Value.TermLabel, v.Value.TermOrigin = name.Key, name.Label, name.Origin
			}
			if len(rank.Date) > 0 {
				d, err := valuecodec.UnmarshalDate(rank.Date)
				if err != nil {
					return err
				}
				v.Value.Date = &d
			}
			if len(rank.Name) > 0 {
				decoded, err := valuecodec.UnmarshalName(rank.Name)
				if err != nil {
					return err
				}
				v.Value.Name = &decoded
			}
			f.Values = append(f.Values, v)
		}
	}
	return nil
}

// loadOutcomes attaches each field's outcomes to it. It also returns the
// outcomes of the Observations in leads (a gallery's earliest records), which
// belong to no field; their Recorded value has no Date or Name.
func loadOutcomes(q Querier, entityIDs [][]byte, byProp map[string]map[string]*Field, leads map[string]bool) (map[string]Outcome, error) {
	leadOutcomes := map[string]Outcome{}
	type pending struct {
		f              *Field
		o              Outcome
		dateID, nameID []byte
	}
	var all []pending
	var dateIDs, nameIDs [][]byte
	err := sqlbatch.Query(q, sqlOutcomes, ")", nil, entityIDs, func(rows *sql.Rows) error {
		var (
			entityID, propertyID, termID, dateID, nameID []byte
			rank                                         sql.NullInt64
			text                                         sql.NullString
			integer                                      sql.NullInt64
			o                                            Outcome
			locatorJSON                                  string
		)
		dest := append([]any{&entityID, &propertyID, &o.ObservationID, &o.ObservationRef, &o.Reason, &rank, &o.DeniedBy,
			&o.Vote.Support, &o.Vote.Of,
			&text, &integer, &termID, &o.Recorded.TermKey, &o.Recorded.TermLabel, &o.Recorded.TermOrigin, &dateID, &nameID,
			&o.SubjectID, &o.SubjectRef, &o.GraphID, &o.CitationID, &o.ArtifactID, &o.SourceID, &o.SourceTitle,
			&o.ArtifactLabel, &o.CitationLabel, &locatorJSON,
			&o.CredibilityKey}, autoreconciler.ProvenanceDest(&o.Provenance)...)
		if err := rows.Scan(dest...); err != nil {
			return err
		}
		o.ValueRank = int(rank.Int64)
		o.ArtifactPage, o.PageLabel = locator.Page(locatorJSON)
		o.Recorded.Text, o.Recorded.HasText = text.String, text.Valid
		o.Recorded.Integer, o.Recorded.HasInteger = integer.Int64, integer.Valid
		o.Recorded.TermID = termID
		f, ok := byProp[string(entityID)][string(propertyID)]
		if !ok {
			if leads[string(o.ObservationID)] {
				leadOutcomes[string(o.ObservationID)] = o
			}
			return nil
		}
		all = append(all, pending{f: f, o: o, dateID: dateID, nameID: nameID})
		if len(dateID) > 0 {
			dateIDs = append(dateIDs, dateID)
		}
		if len(nameID) > 0 {
			nameIDs = append(nameIDs, nameID)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	dates, err := datevalues.LookupManyTx(q, dateIDs)
	if err != nil {
		return nil, err
	}
	names, err := namevalues.LookupManyTx(q, nameIDs)
	if err != nil {
		return nil, err
	}
	for _, p := range all {
		if d, ok := dates[string(p.dateID)]; ok {
			p.o.Recorded.Date = &d
		}
		if n, ok := names[string(p.nameID)]; ok {
			p.o.Recorded.Name = &n
		}
		p.f.Outcomes = append(p.f.Outcomes, p.o)
	}
	return leadOutcomes, nil
}

// DerivedRowKind names a Person life row.
type DerivedRowKind string

const (
	DerivedRowBirthDate  DerivedRowKind = "birth_date"
	DerivedRowBirthPlace DerivedRowKind = "birth_place"
	DerivedRowDeathDate  DerivedRowKind = "death_date"
	DerivedRowDeathPlace DerivedRowKind = "death_place"
)

// OwnerField is the field a derived row reads off the entity that owns it.
// Event is the owner's header when the owner is an Event.
type OwnerField struct {
	Entity canonicalentities.Entity
	Event  *conclusionheaders.EventHeader
	Field  Field
}

// DerivedRow is one life row's records: one entry per owner, in the order
// the row lists them (the lead event first; Places in the header's order).
type DerivedRow struct {
	Kind   DerivedRowKind
	Owners []OwnerField
}

// PersonDerivedRows returns the records behind a Person's birth date, birth
// place, death date and death place rows, in that order. A row with no
// owners is not returned. An owner merged or gone, or without the field the
// row reads, is dropped.
func PersonDerivedRows(c *database.Catalog, personID []byte) ([]DerivedRow, error) {
	birth, death, err := conclusionheaders.PersonLifeOwners(c, personID)
	if err != nil {
		return nil, err
	}
	lists := []struct {
		kind   DerivedRowKind
		owners []conclusionheaders.FieldOwner
		events bool // the owners are Events, so each carries its header
	}{
		{DerivedRowBirthDate, birth.Dates, true}, {DerivedRowBirthPlace, birth.Places, false},
		{DerivedRowDeathDate, death.Dates, true}, {DerivedRowDeathPlace, death.Places, false},
	}
	var all, eventIDs [][]byte
	for _, l := range lists {
		for _, o := range l.owners {
			all = append(all, o.Entity.ID)
			if l.events {
				eventIDs = append(eventIDs, o.Entity.ID)
			}
		}
	}
	ids := database.UniqueBlobIDs(all)
	if len(ids) == 0 {
		return nil, nil
	}
	g := c.Graph()
	if err := g.Nodes(ids); err != nil {
		return nil, err
	}
	nodes := make([]*graphcache.Node, 0, len(ids))
	for _, id := range ids {
		n, err := g.Node(id)
		if err != nil {
			return nil, err
		}
		nodes = append(nodes, n)
	}
	db, err := c.DB()
	if err != nil {
		return nil, err
	}
	fields, _, err := fieldsOf(g, db, nodes, nil)
	if err != nil {
		return nil, err
	}
	headers, err := conclusionheaders.EventsByIDs(c, database.UniqueBlobIDs(eventIDs))
	if err != nil {
		return nil, err
	}
	eventOf := map[string]*conclusionheaders.EventHeader{}
	for i := range headers {
		eventOf[string(headers[i].Entity.ID)] = &headers[i]
	}
	var out []DerivedRow
	for _, l := range lists {
		row := DerivedRow{Kind: l.kind}
		for _, o := range l.owners {
			for _, f := range fields[string(o.Entity.ID)] {
				if !bytes.Equal(f.PropertyID, o.PropertyID) {
					continue
				}
				owner := OwnerField{Entity: o.Entity, Field: f}
				if l.events {
					owner.Event = eventOf[string(o.Entity.ID)]
				}
				row.Owners = append(row.Owners, owner)
				break
			}
		}
		if len(row.Owners) > 0 {
			out = append(out, row)
		}
	}
	return out, nil
}
