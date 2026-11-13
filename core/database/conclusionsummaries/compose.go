package conclusionsummaries

import (
	"bytes"
	"fmt"
	"maps"
	"strings"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/canonicalentities"
	"github.com/mendahu/provenencia/core/database/effects"
	"github.com/mendahu/provenencia/core/database/gallery"
	"github.com/mendahu/provenencia/core/database/graphcache"
	"github.com/mendahu/provenencia/core/database/nameformats"
	"github.com/mendahu/provenencia/core/datevalue"
	"github.com/mendahu/provenencia/core/hops"
	"github.com/mendahu/provenencia/core/namevalue"
	"github.com/mendahu/provenencia/core/valuecodec"
	"github.com/mendahu/provenencia/core/valuecompare"
	"github.com/mendahu/provenencia/core/vocabkey"
)

// vocab is the label map a summary resolves at read time. Term and Property
// labels and "today" are not stored on the row, so a rename or a new day does
// not throw the row away.
type vocab struct {
	propID map[string][]byte
	// propLabel is each product Property's catalog label, keyed by Property key.
	propLabel map[string]string
	termLabel map[string]string
	// termOrigin is each term's origin: a key names a product term only with the product's.
	termOrigin map[string]string
	typeID     map[string][]byte
}

func vocabulary(g *graphcache.Graph) (vocab, error) {
	if v, ok := g.Labels(); ok {
		if got, is := v.(vocab); is {
			return got, nil
		}
	}
	loaded, err := loadVocab(g)
	if err != nil {
		return vocab{}, err
	}
	g.SetLabels(loaded)
	return loaded, nil
}

func loadVocab(g *graphcache.Graph) (vocab, error) {
	v := vocab{
		propID:     map[string][]byte{},
		propLabel:  map[string]string{},
		termLabel:  map[string]string{},
		termOrigin: map[string]string{},
		typeID:     map[string][]byte{},
	}
	if err := scanVocab(g, `SELECT id, key, label FROM properties WHERE origin = ?`, []any{vocabkey.OriginProvenencia}, func(id []byte, cols []string) {
		v.propID[cols[0]] = append([]byte(nil), id...)
		v.propLabel[cols[0]] = cols[1]
	}); err != nil {
		return vocab{}, err
	}
	if err := scanVocab(g, `SELECT id, label, origin FROM property_terms`, nil, func(id []byte, cols []string) {
		v.termLabel[string(id)] = cols[0]
		v.termOrigin[string(id)] = cols[1]
	}); err != nil {
		return vocab{}, err
	}
	if err := scanVocab(g, `SELECT id, key FROM subject_types WHERE origin = ?`, []any{vocabkey.OriginProvenencia}, func(id []byte, cols []string) {
		v.typeID[cols[0]] = append([]byte(nil), id...)
	}); err != nil {
		return vocab{}, err
	}
	return v, nil
}

// scanVocab runs query and hands each row to take: the id, then every further
// column as a string.
func scanVocab(g *graphcache.Graph, query string, args []any, take func(id []byte, cols []string)) error {
	rows, err := g.Query(query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()
	names, err := rows.Columns()
	if err != nil {
		return err
	}
	for rows.Next() {
		var id []byte
		cols := make([]string, len(names)-1)
		dest := make([]any, len(names))
		dest[0] = &id
		for i := range cols {
			dest[i+1] = &cols[i]
		}
		if err := rows.Scan(dest...); err != nil {
			return err
		}
		take(id, cols)
	}
	return rows.Err()
}

// TermLabeler returns a function that builds the summary-wire Term of a term
// id and key, with its label and origin filled as every summary term is. It
// reports false when the id is not a term of the catalog's vocabulary.
func TermLabeler(c *database.Catalog) (func(id []byte, key string) (Term, bool), error) {
	v, err := vocabulary(c.Graph())
	if err != nil {
		return nil, err
	}
	return func(id []byte, key string) (Term, bool) {
		if _, ok := v.termLabel[string(id)]; !ok {
			return Term{}, false
		}
		return *v.finishTerm(&Term{ID: id, Key: key}), true
	}, nil
}

func (v vocab) prop(key string) []byte { return v.propID[key] }

// propLabelOf is the catalog label of the product Property with key.
func (v vocab) propLabelOf(key string) string { return v.propLabel[key] }

// finishTerm returns a copy of t with its label and origin filled from the
// vocabulary; it never writes through t, which may be a stored row's.
func (v vocab) finishTerm(t *Term) *Term {
	if t == nil {
		return nil
	}
	out := *t
	out.Label = v.termLabel[string(out.ID)]
	out.Origin = v.termOrigin[string(out.ID)]
	return &out
}

// finishFacts fills the read-time labels of a term-valued Property's facts.
func (v vocab) finishFacts(f TermFacts, propertyKey string) TermFacts {
	f.Term = v.finishTerm(f.Term)
	f.PropertyLabel = v.propLabelOf(propertyKey)
	return f
}

// follow reads a hop through a scope, so a row's inputs record it. A nil scope
// follows nothing.
func follow(in *graphcache.InputScope, id []byte, hop hops.Hop) ([]graphcache.Link, error) {
	if in == nil {
		return nil, nil
	}
	return in.Follow(id, hop)
}

func keptName(n *graphcache.Node, propID []byte) (*namevalue.Value, int, error) {
	rows := n.Kept(propID)
	row, ok := graphcache.Rank1(rows)
	if !ok || len(row.Name) == 0 {
		return nil, len(rows), nil
	}
	name, err := valuecodec.UnmarshalName(row.Name)
	if err != nil {
		return nil, 0, err
	}
	return &name, len(rows), nil
}

func keptText(n *graphcache.Node, propID []byte) (string, int) {
	rows := n.Kept(propID)
	row, ok := graphcache.Rank1(rows)
	if !ok || !row.HasText {
		return "", len(rows)
	}
	return strings.TrimSpace(row.Text), len(rows)
}

func keptTerm(n *graphcache.Node, propID []byte) (id []byte, key string, nkept int) {
	rows := n.Kept(propID)
	row, ok := graphcache.Rank1(rows)
	if !ok || row.TermKey == "" {
		return nil, "", len(rows)
	}
	return append([]byte(nil), row.TermID...), row.TermKey, len(rows)
}

// keptTermFacts is the rank-1 kept term of a Property and the kept count.
func keptTermFacts(n *graphcache.Node, propID []byte) TermFacts {
	id, key, count := keptTerm(n, propID)
	f := TermFacts{Count: count}
	if key != "" {
		f.Term = &Term{ID: id, Key: key}
	}
	return f
}

func keptDateAt(n *graphcache.Node, propID []byte) (*datevalue.Value, int, error) {
	rows := n.Kept(propID)
	row, ok := graphcache.Rank1(rows)
	if !ok {
		return nil, len(rows), nil
	}
	date, err := unmarshalDate(row.Date)
	if err != nil || date == nil {
		return nil, len(rows), err
	}
	return date, len(rows), nil
}

func toponyms(n *graphcache.Node, propID []byte) []string {
	var names []string
	for _, row := range n.Kept(propID) {
		name := strings.TrimSpace(row.Text)
		if row.HasText && name != "" {
			names = append(names, name)
		}
	}
	return names
}

// ListPersons returns every unmerged Person's summary in list order.
func ListPersons(c *database.Catalog) ([]PersonSummary, error) {
	out, _, err := persons(c.Graph(), nil, true, true, nil)
	if err != nil {
		return nil, err
	}
	for i := range out {
		if out[i].Updated, err = handleActivity(c, out[i].Entity.ID); err != nil {
			return nil, err
		}
	}
	return out, nil
}

// handleActivity is a handle's newest canonical_entity revision, from the
// display cache's last-activity map. The lists fill it onto the rows they
// return, after composing; the finished rows never hold it, because a note edit
// moves the scope without dropping the cached row, and search composes the same
// rows over a throwaway graph during a write.
func handleActivity(c *database.Catalog, id []byte) (effects.Activity, error) {
	return c.Graph().Activity(effects.ScopeCanonicalEntity, id)
}

// Summaries is the summaries ByIDs read, found by the entity's raw id.
type Summaries struct {
	persons map[string]PersonSummary
	events  map[string]EventSummary
	places  map[string]PlaceSummary
	// inputs are the keys each composed row was read from, by the entity's
	// raw id. Only RowsForSearch fills them.
	inputs map[string][]graphcache.InputKey
}

// Inputs are the handles, links and associations the row with id was composed
// from, without the row's own values. A row that was not composed has none.
func (h Summaries) Inputs(id []byte) []graphcache.InputKey {
	return h.inputs[string(id)]
}

// Person is the summary of the Person with id, or false when it was not read.
func (h Summaries) Person(id []byte) (PersonSummary, bool) {
	v, ok := h.persons[string(id)]
	return v, ok
}

// Event is the summary of the Event with id, or false when it was not read.
func (h Summaries) Event(id []byte) (EventSummary, bool) { v, ok := h.events[string(id)]; return v, ok }

// Place is the summary of the Place with id, or false when it was not read.
func (h Summaries) Place(id []byte) (PlaceSummary, bool) { v, ok := h.places[string(id)]; return v, ok }

// PictureImageFileIDs is the image File of every summary that leads with a
// picture, for one files.RelPaths call.
func (h Summaries) PictureImageFileIDs() [][]byte {
	var out [][]byte
	for _, p := range h.persons {
		if len(p.PictureImageFileID) > 0 {
			out = append(out, p.PictureImageFileID)
		}
	}
	for _, e := range h.events {
		if len(e.PictureImageFileID) > 0 {
			out = append(out, e.PictureImageFileID)
		}
	}
	for _, p := range h.places {
		if len(p.PictureImageFileID) > 0 {
			out = append(out, p.PictureImageFileID)
		}
	}
	return out
}

// Want is the Persons, Events and Places a read asks for, split by kind.
type Want struct {
	Persons, Events, Places [][]byte
}

// Add keeps a 16-byte id of kind vocabkey.Person, Event or Place and ignores
// any other kind or id.
func (w *Want) Add(kind string, id []byte) {
	if len(id) != 16 {
		return
	}
	switch kind {
	case vocabkey.Person:
		w.Persons = append(w.Persons, id)
	case vocabkey.Event:
		w.Events = append(w.Events, id)
	case vocabkey.Place:
		w.Places = append(w.Places, id)
	}
}

// NewSummaries indexes the given rows by the entity's raw id.
func NewSummaries(ps []PersonSummary, es []EventSummary, pl []PlaceSummary) Summaries {
	out := Summaries{
		persons: make(map[string]PersonSummary, len(ps)),
		events:  make(map[string]EventSummary, len(es)),
		places:  make(map[string]PlaceSummary, len(pl)),
	}
	for _, h := range ps {
		out.persons[string(h.Entity.ID)] = h
	}
	for _, h := range es {
		out.events[string(h.Entity.ID)] = h
	}
	for _, h := range pl {
		out.places[string(h.Entity.ID)] = h
	}
	return out
}

// ByIDs reads the summaries of the wanted unmerged Persons, Events and Places,
// each kind in one call. A kind with no ids is not read. An id that is unknown,
// merged, or of another kind is absent from its kind's summaries.
func ByIDs(c *database.Catalog, w Want) (Summaries, error) {
	ps, err := PersonsByIDs(c, w.Persons)
	if err != nil {
		return Summaries{}, err
	}
	es, err := EventsByIDs(c, w.Events)
	if err != nil {
		return Summaries{}, err
	}
	pl, err := PlacesByIDs(c, w.Places)
	if err != nil {
		return Summaries{}, err
	}
	return NewSummaries(ps, es, pl), nil
}

// PersonsByIDs returns the summaries of the given unmerged Persons in list
// order. Unknown, merged, and non-Person ids are absent.
func PersonsByIDs(c *database.Catalog, ids [][]byte) ([]PersonSummary, error) {
	if len(database.UniqueBlobIDs(ids)) == 0 {
		return nil, nil
	}
	out, _, err := persons(c.Graph(), ids, false, true, nil)
	return out, err
}

func persons(g *graphcache.Graph, ids [][]byte, all, store bool, rec *inputBudget) ([]PersonSummary, map[string][]graphcache.InputKey, error) {
	if g == nil {
		return nil, nil, nil
	}
	v, err := vocabulary(g)
	if err != nil {
		return nil, nil, err
	}
	nodes, err := nodesOf(g, v.typeID[vocabkey.Person], ids, all)
	if err != nil {
		return nil, nil, err
	}
	out, inputs, err := composePersons(g, v, nodes, store, rec)
	if err != nil {
		return nil, nil, err
	}
	loaded, err := nameformats.LoadFor(g)
	if err != nil {
		return nil, nil, err
	}
	sortPersons(out, loaded.Profile)
	return out, inputs, nil
}

// ListEvents returns every unmerged Event's summary in list order.
func ListEvents(c *database.Catalog) ([]EventSummary, error) {
	out, _, err := events(c.Graph(), nil, true, true, nil)
	if err != nil {
		return nil, err
	}
	for i := range out {
		if out[i].Updated, err = handleActivity(c, out[i].Entity.ID); err != nil {
			return nil, err
		}
	}
	return out, nil
}

// EventsByIDs returns the summaries of the given unmerged Events in list order.
func EventsByIDs(c *database.Catalog, ids [][]byte) ([]EventSummary, error) {
	if len(database.UniqueBlobIDs(ids)) == 0 {
		return nil, nil
	}
	out, _, err := events(c.Graph(), ids, false, true, nil)
	return out, err
}

// RowsForSearch composes the three summary kinds on one graph. Search
// reprojection calls this so a mixed batch does not open three graphs.
func RowsForSearch(g *graphcache.Graph, w Want) (Summaries, error) {
	var ps []PersonSummary
	var es []EventSummary
	var pl []PlaceSummary
	inputs := map[string][]graphcache.InputKey{}
	rec := &inputBudget{}
	var err error
	var more map[string][]graphcache.InputKey
	if len(w.Persons) > 0 {
		if ps, more, err = persons(g, w.Persons, false, false, rec); err != nil {
			return Summaries{}, err
		}
		maps.Copy(inputs, more)
	}
	if len(w.Events) > 0 {
		if es, more, err = events(g, w.Events, false, false, rec); err != nil {
			return Summaries{}, err
		}
		maps.Copy(inputs, more)
	}
	if len(w.Places) > 0 {
		if pl, more, err = places(g, w.Places, false, false, rec); err != nil {
			return Summaries{}, err
		}
		maps.Copy(inputs, more)
	}
	s := NewSummaries(ps, es, pl)
	if !rec.over {
		s.inputs = inputs
	}
	return s, nil
}

func events(g *graphcache.Graph, ids [][]byte, all, store bool, rec *inputBudget) ([]EventSummary, map[string][]graphcache.InputKey, error) {
	if g == nil {
		return nil, nil, nil
	}
	v, err := vocabulary(g)
	if err != nil {
		return nil, nil, err
	}
	nodes, err := nodesOf(g, v.typeID[vocabkey.Event], ids, all)
	if err != nil {
		return nil, nil, err
	}
	out, inputs, err := composeEvents(g, v, nodes, store, rec)
	if err != nil {
		return nil, nil, err
	}
	sortEvents(out)
	return out, inputs, nil
}

// ListPlaces returns every unmerged Place's summary in list order.
func ListPlaces(c *database.Catalog) ([]PlaceSummary, error) {
	out, _, err := places(c.Graph(), nil, true, true, nil)
	if err != nil {
		return nil, err
	}
	for i := range out {
		if out[i].Updated, err = handleActivity(c, out[i].Entity.ID); err != nil {
			return nil, err
		}
	}
	return out, nil
}

// PlacesByIDs returns the summaries of the given unmerged Places in list order.
func PlacesByIDs(c *database.Catalog, ids [][]byte) ([]PlaceSummary, error) {
	if len(database.UniqueBlobIDs(ids)) == 0 {
		return nil, nil
	}
	out, _, err := places(c.Graph(), ids, false, true, nil)
	return out, err
}

func places(g *graphcache.Graph, ids [][]byte, all, store bool, rec *inputBudget) ([]PlaceSummary, map[string][]graphcache.InputKey, error) {
	if g == nil {
		return nil, nil, nil
	}
	v, err := vocabulary(g)
	if err != nil {
		return nil, nil, err
	}
	nodes, err := nodesOf(g, v.typeID[vocabkey.Place], ids, all)
	if err != nil {
		return nil, nil, err
	}
	out, inputs, err := composePlaces(g, v, nodes, store, rec)
	if err != nil {
		return nil, nil, err
	}
	sortPlaces(out)
	return out, inputs, nil
}

func nodesOf(g *graphcache.Graph, typeID []byte, ids [][]byte, all bool) ([]*graphcache.Node, error) {
	if all {
		return g.Kind(typeID)
	}
	ids = database.UniqueBlobIDs(ids)
	if err := g.Nodes(ids); err != nil {
		return nil, err
	}
	var out []*graphcache.Node
	for _, id := range ids {
		n, err := g.Node(id)
		if err != nil {
			return nil, err
		}
		if !n.Is(typeID) {
			continue
		}
		out = append(out, n)
	}
	return out, nil
}

// needed returns the nodes with no stored H display row when store is on, else
// all of them: the ones a compose must build fresh.
func needed[H any](g *graphcache.Graph, nodes []*graphcache.Node, store bool) []*graphcache.Node {
	if !store {
		return nodes
	}
	var need []*graphcache.Node
	for _, n := range nodes {
		if row, ok := g.Display(n.ID); ok {
			if _, is := row.(H); is {
				continue
			}
		}
		need = append(need, n)
	}
	return need
}

// displayed returns one H per node in order: the fresh row, else the stored
// one, each through finish (a stored row is composed but not finished).
func displayed[H any](g *graphcache.Graph, nodes []*graphcache.Node, fresh map[string]H, finish func(H) H) []H {
	out := make([]H, 0, len(nodes))
	for _, n := range nodes {
		h, ok := fresh[string(n.ID)]
		if !ok {
			row, _ := g.Display(n.ID)
			h = row.(H)
		}
		out = append(out, finish(h))
	}
	return out
}

func composePersons(g *graphcache.Graph, v vocab, nodes []*graphcache.Node, store bool, rec *inputBudget) ([]PersonSummary, map[string][]graphcache.InputKey, error) {
	var need [][]byte
	var eventIDs [][]byte
	byID := map[string]*graphcache.Node{}
	for _, n := range nodes {
		byID[string(n.ID)] = n
	}
	scopes := map[string]*graphcache.InputScope{}
	collect := g.NoInputs()
	for _, n := range needed[PersonSummary](g, nodes, store) {
		need = append(need, n.ID)
		scopes[string(n.ID)] = rec.scope(g, n.ID)
		links, err := follow(collect, n.ID, hops.EventsOfSubject)
		if err != nil {
			return nil, nil, err
		}
		for _, l := range links {
			eventIDs = append(eventIDs, l.Neighbor)
		}
	}
	if err := g.Nodes(eventIDs); err != nil {
		return nil, nil, err
	}
	// A person's places are those of its birth and death events only.
	type life struct{ births, deaths []lifePick }
	lives := map[string]life{}
	var pickIDs [][]byte
	for _, id := range need {
		births, deaths, err := lifePicks(scopes[string(id)], v, byID[string(id)])
		if err != nil {
			return nil, nil, err
		}
		lives[string(id)] = life{births, deaths}
		pickIDs = append(pickIDs, lifeEvents(births, deaths)...)
	}
	seeds, err := placesOfEvents(g.NoInputs(), pickIDs)
	if err != nil {
		return nil, nil, err
	}
	pg, err := buildPlaceGraph(g, v, seeds)
	if err != nil {
		return nil, nil, err
	}
	needNodes := make([]*graphcache.Node, len(need))
	for i, id := range need {
		needNodes[i] = byID[string(id)]
	}
	leads, err := leadPictures(g, needNodes)
	if err != nil {
		return nil, nil, err
	}
	fresh := map[string]PersonSummary{}
	inputs := map[string][]graphcache.InputKey{}
	for _, id := range need {
		in, l := scopes[string(id)], lives[string(id)]
		h, err := onePerson(v, byID[string(id)], in, pg, leads[string(id)], l.births, l.deaths)
		if err != nil {
			return nil, nil, err
		}
		if err := pg.addChains(in, rec, lifeEvents(l.births, l.deaths)); err != nil {
			return nil, nil, err
		}
		rec.keep(inputs, id, in)
		fresh[string(id)] = h
		if store {
			g.SetDisplay(id, h)
		}
	}
	return displayed(g, nodes, fresh, func(h PersonSummary) PersonSummary { return finishPerson(v, h) }), inputs, nil
}

// finishPerson fills the read-time labels of a stored person row.
func finishPerson(v vocab, h PersonSummary) PersonSummary {
	h.Gender = v.finishFacts(h.Gender, vocabkey.PropertyGender)
	h.SexAtBirth = v.finishFacts(h.SexAtBirth, vocabkey.PropertySexAtBirth)
	return h
}

// leadPictures returns each node's leading gallery image File id, keyed by
// string(entity id). A node with no gallery picture is absent.
func leadPictures(g *graphcache.Graph, nodes []*graphcache.Node) (map[string][]byte, error) {
	v, err := g.Vocabulary()
	if err != nil {
		return nil, err
	}
	props := gallery.PropertyIDs(v)
	out := map[string][]byte{}
	for _, n := range nodes {
		if l, ok := gallery.Lead(gallery.Of(n, props)); ok {
			out[string(n.ID)] = l.ImageFileID
		}
	}
	return out, nil
}

func onePerson(v vocab, n *graphcache.Node, in *graphcache.InputScope, pg *placeGraph, lead []byte, births, deaths []lifePick) (PersonSummary, error) {
	h := PersonSummary{Entity: canonicalentities.FromNode(n), PictureImageFileID: lead}
	name, count, err := keptName(n, v.prop(vocabkey.PropertyName))
	if err != nil {
		return PersonSummary{}, err
	}
	h.Name, h.NameValueCount = name, count
	h.Gender = keptTermFacts(n, v.prop(vocabkey.PropertyGender))
	h.SexAtBirth = keptTermFacts(n, v.prop(vocabkey.PropertySexAtBirth))
	var errLife error
	if h.Birth, errLife = leadLife(in, v, pg, births); errLife != nil {
		return PersonSummary{}, errLife
	}
	if h.Death, errLife = leadLife(in, v, pg, deaths); errLife != nil {
		return PersonSummary{}, errLife
	}
	return h, nil
}

type lifePick struct {
	event *graphcache.Node
	date  *datevalue.Value
	// dateProp is the Property the date was read from (eventDate).
	dateProp []byte
	count    int
	// key is the pick's place in date order.
	key dateKey
}

// lifePicks returns the person's canonical birth and death events (the
// product's own types), in the
// order the person's links give them. It loads the events at once, so its
// graph queries do not grow with their number.
func lifePicks(in *graphcache.InputScope, v vocab, n *graphcache.Node) (births, deaths []lifePick, err error) {
	lifeLinks, err := follow(in, n.ID, hops.EventsOfSubject)
	if err != nil {
		return nil, nil, err
	}
	eventIDs := make([][]byte, 0, len(lifeLinks))
	for _, l := range lifeLinks {
		eventIDs = append(eventIDs, l.Neighbor)
	}
	if err := in.Nodes(eventIDs); err != nil {
		return nil, nil, err
	}
	for _, l := range lifeLinks {
		ev, err := in.Node(l.Neighbor)
		if err != nil {
			return nil, nil, err
		}
		if ev == nil {
			continue
		}
		// A provisional-only event is not the person's birth or death.
		if !ev.Canonical(v.typeID[vocabkey.Event]) {
			continue
		}
		typeID, kept, _ := keptTerm(ev, v.prop(vocabkey.PropertyEventType))
		key := v.finishTerm(&Term{ID: typeID, Key: kept}).ProductKey()
		if key != string(vocabkey.EventBirth) && key != string(vocabkey.EventDeath) {
			continue
		}
		item, err := lifeFrom(v, ev)
		if err != nil {
			return nil, nil, err
		}
		if key == string(vocabkey.EventBirth) {
			births = append(births, item)
		} else {
			deaths = append(deaths, item)
		}
	}
	return births, deaths, nil
}

func lifeFrom(v vocab, ev *graphcache.Node) (lifePick, error) {
	dating, dateCount, propID, err := eventDate(v, ev)
	if err != nil {
		return lifePick{}, err
	}
	return lifePick{
		event: ev, date: dating.value, dateProp: propID, count: dateCount,
		key: dating.key.withRef(ev.Ref),
	}, nil
}

// eventDate reads an Event's date through eventDating: the kept date, else the
// kept start date, whichever has a place, with its place. propID is the
// Property read; the date Property when neither has a value.
func eventDate(v vocab, ev *graphcache.Node) (dating eventDated, count int, propID []byte, err error) {
	propID = v.prop(vocabkey.PropertyDate)
	date, count, err := keptDateAt(ev, propID)
	if err != nil {
		return eventDated{}, 0, propID, err
	}
	start := v.prop(vocabkey.PropertyStartDate)
	startDate, startCount, err := keptDateAt(ev, start)
	if err != nil {
		return eventDated{}, 0, propID, err
	}
	dating = eventDating(date, startDate)
	if dating.value != nil && dating.isStart {
		return dating, startCount, start, nil
	}
	return dating, count, propID, nil
}

func leadLife(in *graphcache.InputScope, v vocab, pg *placeGraph, items []lifePick) (LifeFacts, error) {
	if len(items) == 0 {
		return LifeFacts{}, nil
	}
	lead := items[0]
	for _, item := range items[1:] {
		if dateKeyLess(item.key, lead.key) {
			lead = item
		}
	}
	locations, err := eventLocations(in, lead.event.ID)
	if err != nil {
		return LifeFacts{}, err
	}
	places, err := eventPlaces(v, pg, lead.event, locations)
	if err != nil {
		return LifeFacts{}, err
	}
	ev := canonicalentities.FromNode(lead.event)
	return LifeFacts{
		Event: &ev, EventCount: len(items),
		Date: lead.date, DateCount: lead.count,
		Places: places,
	}, nil
}

// PersonLifeOwners names, for a Person's birth and death rows, the entity
// and Property each row reads: one date owner per surviving birth (or death)
// event, lead first, and one owner per event–place link of those events, in
// the order the events give their Places (folded at each event's date), each
// holding its link's place end and that Place. A nil or merged
// Person gives none.
func PersonLifeOwners(c *database.Catalog, personID []byte) (birth, death LifeOwners, err error) {
	if c == nil || c.Graph() == nil || len(personID) != 16 {
		return LifeOwners{}, LifeOwners{}, nil
	}
	g := c.Graph()
	n, err := g.Node(personID)
	if err != nil || n == nil || n.Merged {
		return LifeOwners{}, LifeOwners{}, err
	}
	v, err := vocabulary(g)
	if err != nil {
		return LifeOwners{}, LifeOwners{}, err
	}
	in := g.NoInputs()
	births, deaths, err := lifePicks(in, v, n)
	if err != nil {
		return LifeOwners{}, LifeOwners{}, err
	}
	seeds, err := placesOfEvents(g.NoInputs(), lifeEvents(births, deaths))
	if err != nil {
		return LifeOwners{}, LifeOwners{}, err
	}
	pg, err := buildPlaceGraph(g, v, seeds)
	if err != nil {
		return LifeOwners{}, LifeOwners{}, err
	}
	if birth, err = lifeOwners(in, v, pg, births); err != nil {
		return LifeOwners{}, LifeOwners{}, err
	}
	if death, err = lifeOwners(in, v, pg, deaths); err != nil {
		return LifeOwners{}, LifeOwners{}, err
	}
	return birth, death, nil
}

// lifeEvents are the ids of the births' and then the deaths' events.
func lifeEvents(births, deaths []lifePick) [][]byte {
	out := make([][]byte, 0, len(births)+len(deaths))
	for _, pick := range append(append([]lifePick(nil), births...), deaths...) {
		out = append(out, pick.event.ID)
	}
	return out
}

// maxEventPlaces is the most places one event names to the page; the rest are
// not read.
// decided: user, 2026-11-08 (Spike 16 rule "Crafted fan-out is capped, not engineered"): a real event names one to three places
const maxEventPlaces = 20

// eventLocations follows an event's locations through in and keeps every link
// whose place is among the first maxEventPlaces distinct places (by
// Link.Neighbor, whether or not that place's node loads later), in link order.
// It is the one read of an event's places: the place graph, the folded places
// and the life owners all agree on it.
func eventLocations(in *graphcache.InputScope, eventID []byte) ([]graphcache.Link, error) {
	links, err := follow(in, eventID, hops.PlacesOfEvent)
	if err != nil {
		return nil, err
	}
	places := map[string]bool{}
	kept := links[:0:0]
	for _, l := range links {
		if !places[string(l.Neighbor)] {
			if len(places) >= maxEventPlaces {
				continue
			}
			places[string(l.Neighbor)] = true
		}
		kept = append(kept, l)
	}
	return kept, nil
}

// placesOfEvents follows each event's locations through in and returns the
// places at their far ends.
func placesOfEvents(in *graphcache.InputScope, eventIDs [][]byte) ([][]byte, error) {
	var out [][]byte
	for _, id := range eventIDs {
		locations, err := eventLocations(in, id)
		if err != nil {
			return nil, err
		}
		out = append(out, neighborsOf(locations)...)
	}
	return out, nil
}

func neighborsOf(links []graphcache.Link) [][]byte {
	out := make([][]byte, len(links))
	for i, l := range links {
		out[i] = l.Neighbor
	}
	return out
}

func lifeOwners(in *graphcache.InputScope, v vocab, pg *placeGraph, picks []lifePick) (LifeOwners, error) {
	sortByDateKey(picks, func(p lifePick) dateKey { return p.key })
	var out LifeOwners
	listed := map[string]bool{}
	dated := map[string]bool{}
	// One load for every link's association, so the walk's queries do not
	// grow with the links. The links are kept for the walk below.
	var assocIDs [][]byte
	linksOf := map[string][]graphcache.Link{}
	for _, pick := range picks {
		links, err := eventLocations(in, pick.event.ID)
		if err != nil {
			return LifeOwners{}, err
		}
		linksOf[string(pick.event.ID)] = links
		for _, l := range links {
			assocIDs = append(assocIDs, l.Association)
		}
	}
	if err := in.Nodes(assocIDs); err != nil {
		return LifeOwners{}, err
	}
	for _, pick := range picks {
		if !dated[string(pick.event.ID)] {
			dated[string(pick.event.ID)] = true
			out.Dates = append(out.Dates, FieldOwner{Entity: canonicalentities.FromNode(pick.event), PropertyID: pick.dateProp})
		}
		places, err := eventPlaces(v, pg, pick.event, linksOf[string(pick.event.ID)])
		if err != nil {
			return LifeOwners{}, err
		}
		for _, p := range places {
			for _, l := range linksOf[string(pick.event.ID)] {
				if !bytes.Equal(l.Neighbor, p.Entity.ID) || listed[string(l.Association)] {
					continue
				}
				if len(l.Association) == 0 {
					return LifeOwners{}, fmt.Errorf("conclusionsummaries: location link of event %x to place %x has no association", pick.event.ID, l.Neighbor)
				}
				listed[string(l.Association)] = true
				assoc, err := in.Node(l.Association)
				if err != nil {
					return LifeOwners{}, err
				}
				if assoc == nil {
					return LifeOwners{}, fmt.Errorf("conclusionsummaries: association %x is missing", l.Association)
				}
				out.Places = append(out.Places, FieldOwner{
					Entity: canonicalentities.FromNode(assoc), PropertyID: v.prop(hops.PlacesOfEvent.To()),
					Place: &ChainedPlace{Entity: p.Entity, Names: p.Names},
				})
			}
		}
	}
	return out, nil
}

func composeEvents(g *graphcache.Graph, v vocab, nodes []*graphcache.Node, store bool, rec *inputBudget) ([]EventSummary, map[string][]graphcache.InputKey, error) {
	need := needed[EventSummary](g, nodes, store)
	var personIDs, placeIDs [][]byte
	scopes := map[string]*graphcache.InputScope{}
	locationsOf := map[string][]graphcache.Link{}
	collect := g.NoInputs()
	for _, n := range need {
		scopes[string(n.ID)] = rec.scope(g, n.ID)
		subjects, err := follow(collect, n.ID, hops.SubjectsOfEvent)
		if err != nil {
			return nil, nil, err
		}
		for _, l := range subjects {
			personIDs = append(personIDs, l.Neighbor)
		}
		locations, err := eventLocations(scopes[string(n.ID)], n.ID)
		if err != nil {
			return nil, nil, err
		}
		locationsOf[string(n.ID)] = locations
		placeIDs = append(placeIDs, neighborsOf(locations)...)
	}
	if err := g.Nodes(personIDs); err != nil {
		return nil, nil, err
	}
	pg, err := buildPlaceGraph(g, v, placeIDs)
	if err != nil {
		return nil, nil, err
	}
	leads, err := leadPictures(g, need)
	if err != nil {
		return nil, nil, err
	}
	fresh := map[string]EventSummary{}
	inputs := map[string][]graphcache.InputKey{}
	for _, n := range need {
		in := scopes[string(n.ID)]
		locations := locationsOf[string(n.ID)]
		h, err := oneEvent(in, v, pg, n, locations)
		if err != nil {
			return nil, nil, err
		}
		rec.addClosure(in, pg, neighborsOf(locations))
		rec.keep(inputs, n.ID, in)
		h.PictureImageFileID = leads[string(n.ID)]
		fresh[string(n.ID)] = h
		if store {
			g.SetDisplay(n.ID, h)
		}
	}
	return displayed(g, nodes, fresh, func(h EventSummary) EventSummary { return finishEvent(v, h) }), inputs, nil
}

func oneEvent(in *graphcache.InputScope, v vocab, pg *placeGraph, n *graphcache.Node, locations []graphcache.Link) (EventSummary, error) {
	h := EventSummary{Entity: canonicalentities.FromNode(n)}
	h.EventName, h.EventNameCount = keptText(n, v.prop("event_name"))
	h.Type = keptTermFacts(n, v.prop(vocabkey.PropertyEventType))
	var err error
	if h.Date, h.DateCount, err = keptDateAt(n, v.prop(vocabkey.PropertyDate)); err != nil {
		return EventSummary{}, err
	}
	if h.StartDate, h.StartDateCount, err = keptDateAt(n, v.prop(vocabkey.PropertyStartDate)); err != nil {
		return EventSummary{}, err
	}
	if h.EndDate, h.EndDateCount, err = keptDateAt(n, v.prop(vocabkey.PropertyEndDate)); err != nil {
		return EventSummary{}, err
	}
	h.Places, err = eventPlaces(v, pg, n, locations)
	if err != nil {
		return EventSummary{}, err
	}
	seen := map[string]bool{}
	subjectLinks, err := follow(in, n.ID, hops.SubjectsOfEvent)
	if err != nil {
		return EventSummary{}, err
	}
	for _, l := range subjectLinks {
		if seen[string(l.Neighbor)] {
			continue
		}
		seen[string(l.Neighbor)] = true
		person, err := in.Node(l.Neighbor)
		if err != nil {
			return EventSummary{}, err
		}
		if person == nil {
			continue
		}
		name, count, err := keptName(person, v.prop(vocabkey.PropertyName))
		if err != nil {
			return EventSummary{}, err
		}
		h.Subjects = append(h.Subjects, EventSubject{
			Entity: canonicalentities.FromNode(person), Name: name, NameValueCount: count,
		})
	}
	if h.Date != nil {
		h.StartDate = nil
		h.EndDate = nil
	}
	return h, nil
}

func finishEvent(v vocab, h EventSummary) EventSummary {
	h.Type = v.finishFacts(h.Type, vocabkey.PropertyEventType)
	h.Title = eventTitle(h)
	return h
}

type dateKey struct {
	// at is the date's place in date order (valuecompare.DateKeyOf).
	at  valuecompare.DateKey
	ok  bool
	ref string
}

// dateKeyOf is v's place in date order, with ref breaking ties.
func dateKeyOf(v *datevalue.Value, ref string) dateKey {
	p, ok := datePlace(v)
	return dateKey{at: p.at, ok: ok, ref: strings.ToLower(ref)}
}

// withRef is k with ref as its tie-break.
func (k dateKey) withRef(ref string) dateKey {
	k.ref = strings.ToLower(ref)
	return k
}

func sortEvents(rows []EventSummary) {
	sortByDateKey(rows, eventKeyOf)
}

func dateKeyLess(a, b dateKey) bool {
	if a.ok != b.ok {
		return a.ok
	}
	if c := a.at.Compare(b.at); c != 0 {
		return c < 0
	}
	return a.ref < b.ref
}

// AgeRange is the whole years of age a person had on a day: Lo to Hi, equal
// when the dates leave no doubt. A side the dates leave open (a BEF or AFT
// birth or event) is flagged and its bound is 0: only the other side is known.
type AgeRange struct {
	Lo, Hi         int
	OpenLo, OpenHi bool
}

// TimelineSpan is where a date sits on a line of years: From to To, with the
// side the date leaves open flagged. A date open on one side (BEF, AFT, a range
// with one year) spans its known end, so it can be drawn there.
type TimelineSpan struct {
	From, To  float64
	OpenStart bool
	OpenEnd   bool
}

// placedDate is where a date sits: its place in date order, its span and the
// two ends they come from.
type placedDate struct {
	at          valuecompare.DateKey
	span        TimelineSpan
	first, last valuecompare.DateEnd
}

// datePlace is where a date sits: valuecompare.DateKeyOf's key and ends, and
// the span drawn from them. ok is false when v is nil or DateKeyOf is not ok.
func datePlace(v *datevalue.Value) (placedDate, bool) {
	if v == nil {
		return placedDate{}, false
	}
	at, first, last, ok := valuecompare.DateKeyOf(*v)
	if !ok {
		return placedDate{}, false
	}
	p := placedDate{at: at, first: first, last: last}
	switch at.Side {
	case -1:
		p.span = TimelineSpan{From: last.From, To: last.To, OpenStart: true}
	case 1:
		p.span = TimelineSpan{From: first.From, To: first.To, OpenEnd: true}
	default:
		p.span = TimelineSpan{From: first.From, To: last.To}
	}
	return p, true
}

// eventDated is the date an event is placed by, with its place.
type eventDated struct {
	value *datevalue.Value
	// isStart is true when value is the start date.
	isStart bool
	// key is value's order key with no ref; key.ok is false when value has no
	// place.
	key  dateKey
	span TimelineSpan
	// first and last are value's ends when it has a place.
	first, last valuecompare.DateEnd
}

// eventDating chooses the date an event is placed by: date when it has a place,
// else start when it has one, else date if present, else start. It places each
// date once.
func eventDating(date, start *datevalue.Value) eventDated {
	if p, ok := datePlace(date); ok {
		return eventDated{value: date, key: dateKey{at: p.at, ok: true}, span: p.span, first: p.first, last: p.last}
	}
	if p, ok := datePlace(start); ok {
		return eventDated{value: start, isStart: true, key: dateKey{at: p.at, ok: true}, span: p.span, first: p.first, last: p.last}
	}
	if date != nil {
		return eventDated{value: date}
	}
	return eventDated{value: start, isStart: start != nil}
}

func eventKeyOf(h EventSummary) dateKey {
	return eventDating(h.Date, h.StartDate).key.withRef(h.Entity.Ref)
}

// birthKeyOf is a person's place in birth order: the birth date when there is
// one, else the ref alone.
func birthKeyOf(h PersonSummary) dateKey {
	return dateKeyOf(h.Birth.Date, h.Entity.Ref)
}

// eventPlaces are the event's Places, folded at its date; locations are the
// links the caller followed from the event. Each place's names come from its
// entry in pg, so no place node is read.
func eventPlaces(v vocab, pg *placeGraph, ev *graphcache.Node, locations []graphcache.Link) ([]ChainedPlace, error) {
	var raw []ChainedPlace
	seen := map[string]bool{}
	for _, l := range locations {
		if seen[string(l.Neighbor)] {
			continue
		}
		seen[string(l.Neighbor)] = true
		shell, ok := pg.shell(l.Neighbor)
		if !ok {
			continue
		}
		raw = append(raw, ChainedPlace{Entity: shell.Entity, Names: shell.Names})
	}
	if len(raw) == 0 {
		return nil, nil
	}
	at, err := eventWhen(v, ev)
	if err != nil {
		return nil, err
	}
	return pg.FoldLocations(raw, at), nil
}

func eventWhen(v vocab, ev *graphcache.Node) (*datevalue.Value, error) {
	dating, _, _, err := eventDate(v, ev)
	return dating.value, err
}

func composePlaces(g *graphcache.Graph, v vocab, nodes []*graphcache.Node, store bool, rec *inputBudget) ([]PlaceSummary, map[string][]graphcache.InputKey, error) {
	need := needed[PlaceSummary](g, nodes, store)
	leads, err := leadPictures(g, need)
	if err != nil {
		return nil, nil, err
	}
	fresh := map[string]PlaceSummary{}
	for _, n := range need {
		h, err := placeShell(n, v)
		if err != nil {
			return nil, nil, err
		}
		h.PictureImageFileID = leads[string(n.ID)]
		fresh[string(n.ID)] = h
		if store {
			g.SetDisplay(n.ID, h)
		}
	}
	return finishPlaces(g, v, rec, displayed(g, nodes, fresh, func(h PlaceSummary) PlaceSummary { return h }))
}

func placeShell(n *graphcache.Node, v vocab) (PlaceSummary, error) {
	h := PlaceSummary{Entity: canonicalentities.FromNode(n), Names: toponyms(n, v.prop(vocabkey.PropertyToponym))}
	var err error
	if h.StartDate, _, err = keptDateAt(n, v.prop(vocabkey.PropertyStartDate)); err != nil {
		return PlaceSummary{}, err
	}
	if h.EndDate, _, err = keptDateAt(n, v.prop(vocabkey.PropertyEndDate)); err != nil {
		return PlaceSummary{}, err
	}
	return h, nil
}

func finishPlaces(g *graphcache.Graph, v vocab, rec *inputBudget, summaries []PlaceSummary) ([]PlaceSummary, map[string][]graphcache.InputKey, error) {
	if len(summaries) == 0 {
		return nil, nil, nil
	}
	ids := make([][]byte, len(summaries))
	for i := range summaries {
		ids[i] = summaries[i].Entity.ID
	}
	pg, err := buildPlaceGraph(g, v, ids)
	if err != nil {
		return nil, nil, err
	}
	today := TodayDate()
	inputs := make(map[string][]graphcache.InputKey, len(summaries))
	for i := range summaries {
		id := summaries[i].Entity.ID
		at := chainQueryDate(summaries[i], today)
		summaries[i].Parents, summaries[i].ParentsAreCandidates = pg.ParentChain(id, &at)
		in := rec.scope(g, id)
		rec.addClosure(in, pg, [][]byte{id})
		rec.keep(inputs, id, in)
	}
	return summaries, inputs, nil
}

func buildPlaceGraph(g *graphcache.Graph, v vocab, seeds [][]byte) (*placeGraph, error) {
	pg := newPlaceGraph()
	seeds = database.UniqueBlobIDs(seeds)
	if len(seeds) == 0 {
		return pg, nil
	}
	frontier := append([][]byte(nil), seeds...)
	seen := map[string]bool{}
	for _, id := range seeds {
		seen[string(id)] = true
	}
	for round := 0; round < summaryChainDepth && len(frontier) > 0; round++ {
		entries, err := placeEntriesOf(g, v, frontier)
		if err != nil {
			return nil, err
		}
		var next [][]byte
		for _, id := range frontier {
			e, ok := entries[string(id)]
			if !ok {
				continue
			}
			pg.places[string(id)] = e
			for _, l := range e.partFrom {
				if !seen[string(l.to)] {
					seen[string(l.to)] = true
					next = append(next, l.to)
				}
			}
		}
		frontier = next
	}
	return pg, nil
}

// placeStep reads the entry of id and of each place it links to, which is what
// the place page and the parts and succession questions need.
func placeStep(c *database.Catalog, id []byte) (*placeGraph, error) {
	pg := newPlaceGraph()
	if c == nil || c.Graph() == nil || len(id) != 16 {
		return pg, nil
	}
	g := c.Graph()
	v, err := vocabulary(g)
	if err != nil {
		return nil, err
	}
	pages, err := placePagesOf(g, v, [][]byte{id})
	if err != nil {
		return nil, err
	}
	page, ok := pages[string(id)]
	if !ok {
		return pg, nil
	}
	pg.pages[string(id)] = page
	neighbors := [][]byte{id}
	for _, links := range [][]placeLink{page.partTo, page.succFrom, page.succTo} {
		for _, l := range links {
			neighbors = append(neighbors, l.from, l.to)
		}
	}
	own, err := placeEntriesOf(g, v, [][]byte{id})
	if err != nil {
		return nil, err
	}
	for _, l := range own[string(id)].partFrom {
		neighbors = append(neighbors, l.to)
	}
	entries, err := placeEntriesOf(g, v, neighbors)
	if err != nil {
		return nil, err
	}
	for k, e := range entries {
		pg.places[k] = e
	}
	return pg, nil
}

func assocDatesOf(v vocab, n *graphcache.Node) (assocDates, error) {
	start, _, err := keptDateAt(n, v.prop(vocabkey.PropertyStartDate))
	if err != nil {
		return assocDates{}, err
	}
	end, _, err := keptDateAt(n, v.prop(vocabkey.PropertyEndDate))
	if err != nil {
		return assocDates{}, err
	}
	return assocDates{start: start, end: end}, nil
}
