// Package conclusiondetails composes the detail of one canonical handle from
// the catalog graph. For every Property its members' records speak to, it
// returns the state, every auto-reconciled value (displayed or not, with its
// reason), and the auto-reconciler's outcome for each Observation it
// considered, joined to the record's Source: the "Why" a detail page shows.
//
// Field values are the node's auto-reconciler ranks, and the Why's reconciler
// half (reason, value rank, denied by, vote) is the node's outcome rows. Term
// labels come from the graph vocabulary. The record shells (Observation,
// Subject, Citation, Artifact, Source words) are one SQL read keyed by the
// rows' Subjects until the records tier holds them. A Property no
// record mentions isn't returned. Pictures are listed apart from the
// fields, earliest first (gallery.Of over the node). Which empty fields a page
// still shows is the page's choice. Go returns structures; the app formats text.
// Reconciliation Claims are never in these tables: when they ship, a
// composer lays an accepted claim over the auto-reconciled value here.
// A Person's life rows read other handles' fields through PersonDerivedRows,
// composed by the same steps. A place row reads the event–place link's place
// end, each record as its Place's toponym.
package conclusiondetails

import (
	"bytes"
	"database/sql"
	"fmt"
	"sort"

	"github.com/mendahu/provenencia/core/apperr"
	"github.com/mendahu/provenencia/core/autoreconcile"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/canonicalentities"
	"github.com/mendahu/provenencia/core/database/conclusionsummaries"
	"github.com/mendahu/provenencia/core/database/datevalues"
	"github.com/mendahu/provenencia/core/database/gallery"
	"github.com/mendahu/provenencia/core/database/graphcache"
	"github.com/mendahu/provenencia/core/database/namevalues"
	"github.com/mendahu/provenencia/core/database/observations"
	"github.com/mendahu/provenencia/core/database/scalarcols"
	"github.com/mendahu/provenencia/core/database/sources"
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
	propertymeta.Scalar
	TermID    []byte
	TermKey   string
	TermLabel string
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
	// A Subject-valued record reads as its Subject's toponym: every positive
	// toponym Observation of the Subjects in the IN list. INDEXED BY: there are
	// no planner stats, and property_id would scan every toponym.
	sqlToponyms = `SELECT o.subject_id, o.ref, o.polarity, COALESCE(o.value_text, '') FROM observations o INDEXED BY observations_subject_id_idx
		WHERE o.property_id = ? AND o.polarity = 'positive' AND o.subject_id IN (`

	// The record shells of the Observations of the Subjects in the IN list:
	// what each record said and where it came from. The reconciler half of the
	// Why is the node's outcome row. INDEXED BY as sqlToponyms.
	sqlRecords = `SELECT o.id, o.ref,
			o.value_text, o.value_integer, o.value_boolean, o.value_term_id, COALESCE(t.key, ''), COALESCE(t.label, ''), COALESCE(t.origin, ''),
			o.value_date_id, o.value_name_id, o.value_subject_id,
			s.id, s.ref, s.graph_id, c.id, c.artifact_id,
			COALESCE(c.label, ''), c.locator_json, c.transcription_uncertain
		FROM observations o INDEXED BY observations_subject_id_idx
		JOIN subjects s ON s.id = o.subject_id
		JOIN citations c ON c.id = o.citation_id
		LEFT JOIN property_terms t ON t.id = o.value_term_id
		WHERE o.subject_id IN (`
)

// Handle returns the graph node of one unmerged handle. A nil catalog or
// graph, a malformed id, and an unknown or merged handle are ErrNotFound.
func Handle(c *database.Catalog, entityID []byte) (*graphcache.Node, error) {
	if c == nil || c.Graph() == nil || len(entityID) != 16 {
		return nil, ErrNotFound
	}
	n, err := c.Graph().Node(entityID)
	if err != nil {
		return nil, err
	}
	if n == nil || n.Merged {
		return nil, ErrNotFound
	}
	return n, nil
}

// ForEntity composes the detail of one unmerged handle.
func ForEntity(c *database.Catalog, entityID []byte) (Detail, error) {
	n, err := Handle(c, entityID)
	if err != nil {
		return Detail{}, err
	}
	db, err := c.DB()
	if err != nil {
		return Detail{}, err
	}
	d := Detail{Entity: entityOf(n), MemberCount: len(n.AcceptedMembers())}
	vocab, err := c.Graph().Vocabulary()
	if err != nil {
		return Detail{}, err
	}
	galleryProps := gallery.PropertyIDs(vocab)
	tiles := gallery.Of(n, galleryProps)
	leads := map[string]bool{}
	for _, l := range tiles {
		leads[string(l.ObservationIDs[0])] = true
	}
	var leadRows []graphcache.Outcome
	for _, prop := range galleryProps {
		for _, row := range n.Outcomes[string(prop)] {
			if leads[string(row.ObservationID)] {
				leadRows = append(leadRows, row)
			}
		}
	}
	fields, leadOutcomes, err := fieldsOf(c.Graph(), db, vocab, []*graphcache.Node{n}, leadRows, nil)
	if err != nil {
		return Detail{}, err
	}
	d.Fields = fields[string(entityID)]
	for _, l := range tiles {
		if o, ok := leadOutcomes[string(l.ObservationIDs[0])]; ok {
			d.Pictures = append(d.Pictures, PictureTile{ImageFileID: l.ImageFileID, Record: o, Preferred: l.Preferred, Leads: l.Leads})
		}
	}
	return d, nil
}

// fieldsOf composes the fields of several handles at once, keyed by entity
// id: the one composition every page's fields come from. A nil or merged
// node, or a repeated id, is skipped. It also returns the outcomes of the
// Observations of leadRows (a gallery's earliest records, the node's own
// outcome rows), which belong to no field. only nil composes each node's
// page fields: every Property but depiction and subject-valued ones. Non-nil,
// it names the Properties to compose (any value type but depiction) under
// each node's id; a node it does not name has no field.
func fieldsOf(g *graphcache.Graph, q Querier, v graphcache.Vocabulary, nodes []*graphcache.Node, leadRows []graphcache.Outcome, only map[string]map[string]bool) (map[string][]Field, map[string]Outcome, error) {
	var use []*graphcache.Node
	seen := map[string]bool{}
	for _, n := range nodes {
		if n == nil || n.Merged || seen[string(n.ID)] {
			continue
		}
		seen[string(n.ID)] = true
		use = append(use, n)
	}
	out := map[string][]Field{}
	if len(use) == 0 {
		return out, map[string]Outcome{}, nil
	}
	ordered, byProp, err := loadProperties(g, use, only)
	if err != nil {
		return nil, nil, err
	}
	for _, n := range use {
		if err := attachValues(g, n, byProp[string(n.ID)]); err != nil {
			return nil, nil, err
		}
	}
	toponym := v.PropertyID(vocabkey.PropertyToponym, vocabkey.OriginProvenencia)
	recs, err := sources.RecordsOf(g)
	if err != nil {
		return nil, nil, err
	}
	leadOutcomes, err := loadRecords(q, recs, toponym, use, byProp, leadRows)
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
		counts := autoreconcile.Reason(o.Reason).IntoValue()
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
// the same fields by entity id then property id. A node's fields are the
// Properties its outcome rows name, narrowed by only (see fieldsOf).
func loadProperties(g *graphcache.Graph, nodes []*graphcache.Node, only map[string]map[string]bool) (map[string][]*Field, map[string]map[string]*Field, error) {
	v, err := g.Vocabulary()
	if err != nil {
		return nil, nil, err
	}
	props := v.PropertiesByID()
	ordered := map[string][]*Field{}
	byProp := map[string]map[string]*Field{}
	for _, n := range nodes {
		ids := make([]string, 0, len(n.Outcomes))
		for id := range n.Outcomes {
			if only != nil && !only[string(n.ID)][id] {
				continue
			}
			ids = append(ids, id)
		}
		// Ties in the type order keep a fixed order: the Properties' id bytes.
		sort.Strings(ids)
		var got []*Field
		for _, id := range ids {
			p, ok := props[id]
			if !ok || p.ValueType == propertymeta.ValueTypeDepiction || (only == nil && p.ValueType == propertymeta.ValueTypeSubject) {
				continue
			}
			got = append(got, &Field{
				PropertyID: append([]byte(nil), p.ID...), PropertyKey: p.Key, Label: p.Label,
				ValueType: p.ValueType, cardinality: p.Cardinality,
			})
		}
		if len(got) == 0 {
			continue
		}
		order := v.TypeOrder(n.SubjectTypeID)
		sort.SliceStable(got, func(i, j int) bool {
			return vocabkey.CompareByOrder(order, vocabkey.PropertyKey(got[i].PropertyKey), vocabkey.PropertyKey(got[j].PropertyKey)) < 0
		})
		ordered[string(n.ID)] = got
		byProp[string(n.ID)] = map[string]*Field{}
		for _, f := range got {
			byProp[string(n.ID)][string(f.PropertyID)] = f
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
			continue // a depiction Property, or one the outcomes no longer name
		}
		for _, rank := range ranks {
			v := ReconciledValue{
				Rank: rank.Rank, Reason: rank.Reason,
				Support: rank.Support, Against: rank.Against,
			}
			v.Value.Scalar = rank.Scalar
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

// loadRecords attaches each field's outcomes to it: the node's outcome row
// with the record shell read by the rows' Subjects. It also returns the
// outcomes of the Observations in leads (a gallery's earliest records, found
// among the node's rows of the gallery Properties), which belong to no field;
// their Recorded value has no Date or Name. A row with no record is dropped.
func loadRecords(q Querier, recs *sources.RecordIndex, toponym []byte, nodes []*graphcache.Node, byProp map[string]map[string]*Field, leadRows []graphcache.Outcome) (map[string]Outcome, error) {
	leadOutcomes := map[string]Outcome{}
	type want struct {
		f   *Field // nil for a lead
		row graphcache.Outcome
	}
	wanted := map[string][]want{}
	var subjectIDs [][]byte
	add := func(f *Field, row graphcache.Outcome) {
		if len(row.SubjectID) == 0 {
			return
		}
		wanted[string(row.ObservationID)] = append(wanted[string(row.ObservationID)], want{f: f, row: row})
		subjectIDs = append(subjectIDs, row.SubjectID)
	}
	for _, n := range nodes {
		for prop, f := range byProp[string(n.ID)] {
			for _, row := range n.Outcomes[prop] {
				add(f, row)
			}
		}
	}
	for _, row := range leadRows {
		add(nil, row)
	}
	type pending struct {
		f                         *Field
		o                         Outcome
		dateID, nameID, subjectID []byte
	}
	var all []pending
	var dateIDs, nameIDs [][]byte
	var valueSubjectIDs [][]byte
	err := sqlbatch.Query(q, sqlRecords, ")", nil, database.UniqueBlobIDs(subjectIDs), func(rows *sql.Rows) error {
		var (
			observationID, termID, dateID, nameID, valueSubjectID []byte
			scalar                                                scalarcols.Columns
			o                                                     Outcome
			locatorJSON                                           string
		)
		dest := []any{&observationID, &o.ObservationRef,
			&scalar.Text, &scalar.Integer, &scalar.Boolean, &termID, &o.Recorded.TermKey, &o.Recorded.TermLabel, &o.Recorded.TermOrigin, &dateID, &nameID, &valueSubjectID,
			&o.SubjectID, &o.SubjectRef, &o.GraphID, &o.CitationID, &o.ArtifactID,
			&o.CitationLabel, &locatorJSON, &o.Provenance.Uncertain}
		if err := rows.Scan(dest...); err != nil {
			return err
		}
		targets, ok := wanted[string(observationID)]
		if !ok {
			return nil
		}
		src, label, found := recs.Artifact(o.ArtifactID)
		if !found {
			return fmt.Errorf("conclusiondetails: Artifact %x has no Source record", o.ArtifactID)
		}
		o.SourceID, o.SourceTitle, o.ArtifactLabel = src.ID, src.Title, label
		o.CredibilityKey, o.Provenance.Credibility = src.CredibilityKey, src.Credibility
		o.ArtifactPage, o.PageLabel = locator.Page(locatorJSON)
		o.Recorded.Scalar = scalar.Scalar()
		o.Recorded.TermID = termID
		for _, t := range targets {
			rec := o
			rec.ObservationID = append([]byte(nil), t.row.ObservationID...)
			rec.Reason = t.row.Reason
			rec.ValueRank = t.row.ValueRank
			rec.DeniedBy = t.row.DeniedBy // the node's rows are cloned at load
			rec.Vote = autoreconcile.Vote{Support: t.row.VoteSupport, Of: t.row.VoteOf}
			if t.f == nil {
				leadOutcomes[string(rec.ObservationID)] = rec
				continue
			}
			all = append(all, pending{f: t.f, o: rec, dateID: dateID, nameID: nameID, subjectID: valueSubjectID})
			if len(valueSubjectID) > 0 {
				valueSubjectIDs = append(valueSubjectIDs, valueSubjectID)
			}
			if len(dateID) > 0 {
				dateIDs = append(dateIDs, dateID)
			}
			if len(nameID) > 0 {
				nameIDs = append(nameIDs, nameID)
			}
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
	toponyms, err := toponymsOf(q, toponym, valueSubjectIDs)
	if err != nil {
		return nil, err
	}
	for _, p := range all {
		if len(p.subjectID) > 0 {
			// A record on an event–place link's place end holds a Place: it
			// reads as that Place's first toponym (D-21).
			p.o.Recorded.Text = toponyms[string(p.subjectID)]
			p.o.Recorded.HasText = p.o.Recorded.Text != ""
		}
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

// toponymsOf reads each Subject as its first toponym (observations.FirstText),
// from one batched load. A Subject with none is absent.
func toponymsOf(q Querier, toponym []byte, subjectIDs [][]byte) (map[string]string, error) {
	subjectIDs = database.UniqueBlobIDs(subjectIDs)
	out := map[string]string{}
	if len(subjectIDs) == 0 {
		return out, nil
	}
	records := map[string][]observations.TextRecord{}
	err := sqlbatch.Query(q, sqlToponyms, ")", []any{toponym}, subjectIDs, func(rows *sql.Rows) error {
		var id []byte
		var r observations.TextRecord
		if err := rows.Scan(&id, &r.Ref, &r.Polarity, &r.Text); err != nil {
			return err
		}
		records[string(id)] = append(records[string(id)], r)
		return nil
	})
	if err != nil {
		return nil, err
	}
	for id, rs := range records {
		out[id] = observations.FirstText(rs)
	}
	return out, nil
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
// Event is the owner's summary when the owner is an Event. Place is the Place
// a place row's owner (an event–place link's association) holds on its place
// end, with its own toponyms; nil for every other owner.
type OwnerField struct {
	Entity canonicalentities.Entity
	Event  *conclusionsummaries.EventSummary
	Place  *conclusionsummaries.ChainedPlace
	Field  Field
}

// DerivedRow is one life row's records: one entry per owner, in the order
// the row lists them (the lead event first; place links in the summary's order).
type DerivedRow struct {
	Kind   DerivedRowKind
	Owners []OwnerField
}

// PersonDerivedRows returns the records behind a Person's birth date, birth
// place, death date and death place rows, in that order. A row with no
// owners is not returned. An owner merged or gone, or without the field the
// row reads, is dropped.
func PersonDerivedRows(c *database.Catalog, personID []byte) ([]DerivedRow, error) {
	birth, death, err := conclusionsummaries.PersonLifeOwners(c, personID)
	if err != nil {
		return nil, err
	}
	lists := []struct {
		kind   DerivedRowKind
		owners []conclusionsummaries.FieldOwner
		events bool // the owners are Events, so each carries its summary
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
	v, err := g.Vocabulary()
	if err != nil {
		return nil, err
	}
	only := map[string]map[string]bool{}
	for _, l := range lists {
		for _, o := range l.owners {
			if only[string(o.Entity.ID)] == nil {
				only[string(o.Entity.ID)] = map[string]bool{}
			}
			only[string(o.Entity.ID)][string(o.PropertyID)] = true
		}
	}
	fields, _, err := fieldsOf(g, db, v, nodes, nil, only)
	if err != nil {
		return nil, err
	}
	summaries, err := conclusionsummaries.EventsByIDs(c, database.UniqueBlobIDs(eventIDs))
	if err != nil {
		return nil, err
	}
	eventOf := map[string]*conclusionsummaries.EventSummary{}
	for i := range summaries {
		eventOf[string(summaries[i].Entity.ID)] = &summaries[i]
	}
	var out []DerivedRow
	for _, l := range lists {
		row := DerivedRow{Kind: l.kind}
		for _, o := range l.owners {
			for _, f := range fields[string(o.Entity.ID)] {
				if !bytes.Equal(f.PropertyID, o.PropertyID) {
					continue
				}
				owner := OwnerField{Entity: o.Entity, Place: o.Place, Field: f}
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
