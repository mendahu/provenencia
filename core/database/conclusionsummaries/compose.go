package conclusionsummaries

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/mendahu/provenencia/core/autoreconcile"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/canonicalentities"
	"github.com/mendahu/provenencia/core/database/effects"
	"github.com/mendahu/provenencia/core/database/gallery"
	"github.com/mendahu/provenencia/core/database/graphcache"
	"github.com/mendahu/provenencia/core/database/nameformats"
	"github.com/mendahu/provenencia/core/datevalue"
	"github.com/mendahu/provenencia/core/hops"
	"github.com/mendahu/provenencia/core/namevalue"
	"github.com/mendahu/provenencia/core/propertymeta"
	"github.com/mendahu/provenencia/core/valuecodec"
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

func entityOf(n *graphcache.Node) canonicalentities.Entity {
	return canonicalentities.Entity{
		ID:            append([]byte(nil), n.ID...),
		SubjectTypeID: append([]byte(nil), n.SubjectTypeID...),
		Ref:           n.Ref,
		Argument:      n.Argument,
		Label:         n.Label,
	}
}

func follow(g *graphcache.Graph, id []byte, hop hops.Hop) ([]graphcache.Link, error) {
	if g == nil {
		return nil, nil
	}
	return g.Follow(id, hop)
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

func keptDateAt(n *graphcache.Node, propID []byte) (*datevalue.Value, int, string, bool, error) {
	rows := n.Kept(propID)
	row, ok := graphcache.Rank1(rows)
	if !ok {
		return nil, len(rows), "", false, nil
	}
	date, err := unmarshalDate(row.Date)
	if err != nil || date == nil {
		return nil, len(rows), "", false, err
	}
	key, ok := autoreconcile.SortKey(propertymeta.ValueTypeDate, autoreconcile.Value{Date: date})
	return date, len(rows), key, ok, nil
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
	out, err := persons(c.Graph(), nil, true, true)
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
	return persons(c.Graph(), ids, false, true)
}

func persons(g *graphcache.Graph, ids [][]byte, all, store bool) ([]PersonSummary, error) {
	if g == nil {
		return nil, nil
	}
	v, err := vocabulary(g)
	if err != nil {
		return nil, err
	}
	nodes, err := nodesOf(g, v.typeID[vocabkey.Person], ids, all)
	if err != nil {
		return nil, err
	}
	out, err := composePersons(g, v, nodes, store)
	if err != nil {
		return nil, err
	}
	loaded, err := nameformats.LoadFor(g, g)
	if err != nil {
		return nil, err
	}
	order := loaded.Profile.Sorted
	SortByTitle(out, func(h PersonSummary) string { return personSortKey(h, order) }, func(h PersonSummary) string { return h.Entity.Ref })
	return out, nil
}

// ListEvents returns every unmerged Event's summary in list order.
func ListEvents(c *database.Catalog) ([]EventSummary, error) {
	out, err := events(c.Graph(), nil, true, true)
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
	return events(c.Graph(), ids, false, true)
}

// RowsForSearch composes the three summary kinds on one graph. Search
// reprojection calls this so a mixed batch does not open three graphs.
func RowsForSearch(g *graphcache.Graph, w Want) (Summaries, error) {
	var ps []PersonSummary
	var es []EventSummary
	var pl []PlaceSummary
	var err error
	if len(w.Persons) > 0 {
		if ps, err = persons(g, w.Persons, false, false); err != nil {
			return Summaries{}, err
		}
	}
	if len(w.Events) > 0 {
		if es, err = events(g, w.Events, false, false); err != nil {
			return Summaries{}, err
		}
	}
	if len(w.Places) > 0 {
		if pl, err = places(g, w.Places, false, false); err != nil {
			return Summaries{}, err
		}
	}
	return NewSummaries(ps, es, pl), nil
}

func events(g *graphcache.Graph, ids [][]byte, all, store bool) ([]EventSummary, error) {
	if g == nil {
		return nil, nil
	}
	v, err := vocabulary(g)
	if err != nil {
		return nil, err
	}
	nodes, err := nodesOf(g, v.typeID[vocabkey.Event], ids, all)
	if err != nil {
		return nil, err
	}
	out, err := composeEvents(g, v, nodes, store)
	if err != nil {
		return nil, err
	}
	sortEvents(out)
	return out, nil
}

// ListPlaces returns every unmerged Place's summary in list order.
func ListPlaces(c *database.Catalog) ([]PlaceSummary, error) {
	out, err := places(c.Graph(), nil, true, true)
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
	return places(c.Graph(), ids, false, true)
}

func places(g *graphcache.Graph, ids [][]byte, all, store bool) ([]PlaceSummary, error) {
	if g == nil {
		return nil, nil
	}
	v, err := vocabulary(g)
	if err != nil {
		return nil, err
	}
	nodes, err := nodesOf(g, v.typeID[vocabkey.Place], ids, all)
	if err != nil {
		return nil, err
	}
	out, err := composePlaces(g, v, nodes, store)
	if err != nil {
		return nil, err
	}
	SortByTitle(out, placeTitle, func(h PlaceSummary) string { return h.Entity.Ref })
	return out, nil
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

func composePersons(g *graphcache.Graph, v vocab, nodes []*graphcache.Node, store bool) ([]PersonSummary, error) {
	var need [][]byte
	var eventIDs [][]byte
	for _, n := range needed[PersonSummary](g, nodes, store) {
		need = append(need, n.ID)
		links, err := follow(g, n.ID, hops.EventsOfSubject)
		if err != nil {
			return nil, err
		}
		for _, l := range links {
			eventIDs = append(eventIDs, l.Neighbor)
		}
	}
	if err := g.Nodes(eventIDs); err != nil {
		return nil, err
	}
	pg, err := eventPlaceGraph(g, v, eventIDs)
	if err != nil {
		return nil, err
	}
	byID := map[string]*graphcache.Node{}
	for _, n := range nodes {
		byID[string(n.ID)] = n
	}
	leads, err := leadPictures(g, need)
	if err != nil {
		return nil, err
	}
	fresh := map[string]PersonSummary{}
	for _, id := range need {
		h, err := onePerson(v, byID[string(id)], g, pg, leads[string(id)])
		if err != nil {
			return nil, err
		}
		fresh[string(id)] = h
		if store {
			g.SetDisplay(id, h)
		}
	}
	return displayed(g, nodes, fresh, func(h PersonSummary) PersonSummary { return finishPerson(v, h) }), nil
}

// finishPerson fills the read-time labels of a stored person row.
func finishPerson(v vocab, h PersonSummary) PersonSummary {
	h.Gender = v.finishFacts(h.Gender, vocabkey.PropertyGender)
	h.SexAtBirth = v.finishFacts(h.SexAtBirth, vocabkey.PropertySexAtBirth)
	return h
}

// leadPictures returns each handle's leading gallery image File id, keyed by
// string(entity id). A handle with no gallery picture is absent.
func leadPictures(g *graphcache.Graph, ids [][]byte) (map[string][]byte, error) {
	tiles, err := gallery.ForEntities(g, ids)
	if err != nil {
		return nil, err
	}
	out := map[string][]byte{}
	for _, id := range ids {
		if l, ok := gallery.Lead(tiles[string(id)]); ok {
			out[string(id)] = l.ImageFileID
		}
	}
	return out, nil
}

func onePerson(v vocab, n *graphcache.Node, g *graphcache.Graph, pg *placeGraph, lead []byte) (PersonSummary, error) {
	h := PersonSummary{Entity: entityOf(n), PictureImageFileID: lead}
	name, count, err := keptName(n, v.prop(vocabkey.PropertyName))
	if err != nil {
		return PersonSummary{}, err
	}
	h.Name, h.NameValueCount = name, count
	h.Gender = keptTermFacts(n, v.prop(vocabkey.PropertyGender))
	h.SexAtBirth = keptTermFacts(n, v.prop(vocabkey.PropertySexAtBirth))
	births, deaths, err := lifePicks(g, v, n)
	if err != nil {
		return PersonSummary{}, err
	}
	var errLife error
	if h.Birth, errLife = leadLife(g, v, pg, births); errLife != nil {
		return PersonSummary{}, errLife
	}
	if h.Death, errLife = leadLife(g, v, pg, deaths); errLife != nil {
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

// lifePicks returns the person's surviving birth and death events, in the
// order the person's links give them. It loads the events at once, so its
// graph queries do not grow with their number.
func lifePicks(g *graphcache.Graph, v vocab, n *graphcache.Node) (births, deaths []lifePick, err error) {
	lifeLinks, err := follow(g, n.ID, hops.EventsOfSubject)
	if err != nil {
		return nil, nil, err
	}
	eventIDs := make([][]byte, 0, len(lifeLinks))
	for _, l := range lifeLinks {
		eventIDs = append(eventIDs, l.Neighbor)
	}
	if err := g.Nodes(eventIDs); err != nil {
		return nil, nil, err
	}
	for _, l := range lifeLinks {
		ev, err := g.Node(l.Neighbor)
		if err != nil {
			return nil, nil, err
		}
		if ev == nil {
			continue
		}
		_, key, _ := keptTerm(ev, v.prop(vocabkey.PropertyEventType))
		if key != "birth" && key != "death" {
			continue
		}
		item, err := lifeFrom(v, ev)
		if err != nil {
			return nil, nil, err
		}
		if key == "birth" {
			births = append(births, item)
		} else {
			deaths = append(deaths, item)
		}
	}
	return births, deaths, nil
}

func lifeFrom(v vocab, ev *graphcache.Node) (lifePick, error) {
	date, dateCount, sortKey, dateOK, propID, err := eventDate(v, ev)
	if err != nil {
		return lifePick{}, err
	}
	return lifePick{
		event: ev, date: date, dateProp: propID, count: dateCount,
		key: dateKey{key: sortKey, ok: dateOK, ref: strings.ToLower(ev.Ref)},
	}, nil
}

// eventDate reads an Event's date: the kept date, else the kept start date.
// propID is the Property read; the date Property when neither has a value.
func eventDate(v vocab, ev *graphcache.Node) (date *datevalue.Value, count int, sortKey string, ok bool, propID []byte, err error) {
	propID = v.prop(vocabkey.PropertyDate)
	date, count, sortKey, ok, err = keptDateAt(ev, propID)
	if err != nil || date != nil {
		return date, count, sortKey, ok, propID, err
	}
	start := v.prop(vocabkey.PropertyStartDate)
	date, count, sortKey, ok, err = keptDateAt(ev, start)
	if err == nil && date != nil {
		propID = start
	}
	return date, count, sortKey, ok, propID, err
}

func leadLife(g *graphcache.Graph, v vocab, pg *placeGraph, items []lifePick) (LifeFacts, error) {
	if len(items) == 0 {
		return LifeFacts{}, nil
	}
	lead := items[0]
	for _, item := range items[1:] {
		if dateKeyLess(item.key, lead.key) {
			lead = item
		}
	}
	places, err := eventPlaces(g, v, pg, lead.event)
	if err != nil {
		return LifeFacts{}, err
	}
	ev := entityOf(lead.event)
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
	births, deaths, err := lifePicks(g, v, n)
	if err != nil {
		return LifeOwners{}, LifeOwners{}, err
	}
	var eventIDs [][]byte
	for _, pick := range append(append([]lifePick(nil), births...), deaths...) {
		eventIDs = append(eventIDs, pick.event.ID)
	}
	pg, err := eventPlaceGraph(g, v, eventIDs)
	if err != nil {
		return LifeOwners{}, LifeOwners{}, err
	}
	if birth, err = lifeOwners(g, v, pg, births); err != nil {
		return LifeOwners{}, LifeOwners{}, err
	}
	if death, err = lifeOwners(g, v, pg, deaths); err != nil {
		return LifeOwners{}, LifeOwners{}, err
	}
	return birth, death, nil
}

// eventPlaceGraph builds the placeGraph for every Place of the given events,
// skipping events with no node.
func eventPlaceGraph(g *graphcache.Graph, v vocab, eventIDs [][]byte) (*placeGraph, error) {
	var placeIDs [][]byte
	for _, id := range eventIDs {
		n, err := g.Node(id)
		if err != nil {
			return nil, err
		}
		if n == nil {
			continue
		}
		locations, err := follow(g, id, hops.PlacesOfEvent)
		if err != nil {
			return nil, err
		}
		for _, l := range locations {
			placeIDs = append(placeIDs, l.Neighbor)
		}
	}
	return buildPlaceGraph(g, v, placeIDs)
}

func lifeOwners(g *graphcache.Graph, v vocab, pg *placeGraph, picks []lifePick) (LifeOwners, error) {
	sortByDateKey(picks, func(p lifePick) dateKey { return p.key })
	var out LifeOwners
	listed := map[string]bool{}
	dated := map[string]bool{}
	// One load for every link's association, so the walk's queries do not
	// grow with the links. The links are kept for the walk below.
	var assocIDs [][]byte
	linksOf := map[string][]graphcache.Link{}
	for _, pick := range picks {
		links, err := follow(g, pick.event.ID, hops.PlacesOfEvent)
		if err != nil {
			return LifeOwners{}, err
		}
		linksOf[string(pick.event.ID)] = links
		for _, l := range links {
			assocIDs = append(assocIDs, l.Association)
		}
	}
	if err := g.Nodes(assocIDs); err != nil {
		return LifeOwners{}, err
	}
	for _, pick := range picks {
		if !dated[string(pick.event.ID)] {
			dated[string(pick.event.ID)] = true
			out.Dates = append(out.Dates, FieldOwner{Entity: entityOf(pick.event), PropertyID: pick.dateProp})
		}
		places, err := eventPlaces(g, v, pg, pick.event)
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
				assoc, err := g.Node(l.Association)
				if err != nil {
					return LifeOwners{}, err
				}
				if assoc == nil {
					return LifeOwners{}, fmt.Errorf("conclusionsummaries: association %x is missing", l.Association)
				}
				out.Places = append(out.Places, FieldOwner{
					Entity: entityOf(assoc), PropertyID: v.prop(hops.PlacesOfEvent.To()),
					Place: &ChainedPlace{Entity: p.Entity, Names: p.Names},
				})
			}
		}
	}
	return out, nil
}

func composeEvents(g *graphcache.Graph, v vocab, nodes []*graphcache.Node, store bool) ([]EventSummary, error) {
	need := needed[EventSummary](g, nodes, store)
	var personIDs, placeIDs [][]byte
	for _, n := range need {
		subjects, err := follow(g, n.ID, hops.SubjectsOfEvent)
		if err != nil {
			return nil, err
		}
		for _, l := range subjects {
			personIDs = append(personIDs, l.Neighbor)
		}
		locations, err := follow(g, n.ID, hops.PlacesOfEvent)
		if err != nil {
			return nil, err
		}
		for _, l := range locations {
			placeIDs = append(placeIDs, l.Neighbor)
		}
	}
	if err := g.Nodes(append(append([][]byte(nil), personIDs...), placeIDs...)); err != nil {
		return nil, err
	}
	pg, err := buildPlaceGraph(g, v, placeIDs)
	if err != nil {
		return nil, err
	}
	needIDs := make([][]byte, len(need))
	for i, n := range need {
		needIDs[i] = n.ID
	}
	leads, err := leadPictures(g, needIDs)
	if err != nil {
		return nil, err
	}
	fresh := map[string]EventSummary{}
	for _, n := range need {
		h, err := oneEvent(g, v, pg, n)
		if err != nil {
			return nil, err
		}
		h.PictureImageFileID = leads[string(n.ID)]
		fresh[string(n.ID)] = h
		if store {
			g.SetDisplay(n.ID, h)
		}
	}
	return displayed(g, nodes, fresh, func(h EventSummary) EventSummary { return finishEvent(v, h) }), nil
}

func oneEvent(g *graphcache.Graph, v vocab, pg *placeGraph, n *graphcache.Node) (EventSummary, error) {
	h := EventSummary{Entity: entityOf(n)}
	h.EventName, h.EventNameCount = keptText(n, v.prop("event_name"))
	h.Type = keptTermFacts(n, v.prop(vocabkey.PropertyEventType))
	var err error
	if h.Date, h.DateCount, _, _, err = keptDateAt(n, v.prop(vocabkey.PropertyDate)); err != nil {
		return EventSummary{}, err
	}
	if h.StartDate, h.StartDateCount, _, _, err = keptDateAt(n, v.prop(vocabkey.PropertyStartDate)); err != nil {
		return EventSummary{}, err
	}
	if h.EndDate, h.EndDateCount, _, _, err = keptDateAt(n, v.prop(vocabkey.PropertyEndDate)); err != nil {
		return EventSummary{}, err
	}
	h.Places, err = eventPlaces(g, v, pg, n)
	if err != nil {
		return EventSummary{}, err
	}
	seen := map[string]bool{}
	subjectLinks, err := follow(g, n.ID, hops.SubjectsOfEvent)
	if err != nil {
		return EventSummary{}, err
	}
	for _, l := range subjectLinks {
		if seen[string(l.Neighbor)] {
			continue
		}
		seen[string(l.Neighbor)] = true
		person, err := g.Node(l.Neighbor)
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
			Entity: entityOf(person), Name: name, NameValueCount: count,
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
	key string
	ok  bool
	ref string
}

func sortEvents(rows []EventSummary) {
	sortByDateKey(rows, eventKeyOf)
}

func dateKeyLess(a, b dateKey) bool {
	if a.ok != b.ok {
		return a.ok
	}
	if a.key != b.key {
		return a.key < b.key
	}
	return a.ref < b.ref
}

func eventKeyOf(h EventSummary) dateKey {
	if key, ok := dateSort(h.Date); ok {
		return dateKey{key: key, ok: true, ref: strings.ToLower(h.Entity.Ref)}
	}
	if key, ok := dateSort(h.StartDate); ok {
		return dateKey{key: key, ok: true, ref: strings.ToLower(h.Entity.Ref)}
	}
	return dateKey{ref: strings.ToLower(h.Entity.Ref)}
}

// birthKeyOf is a person's place in birth order: the birth date when there is
// one, else the ref alone.
func birthKeyOf(h PersonSummary) dateKey {
	ref := strings.ToLower(h.Entity.Ref)
	if key, ok := dateSort(h.Birth.Date); ok {
		return dateKey{key: key, ok: true, ref: ref}
	}
	return dateKey{ref: ref}
}

func dateSort(v *datevalue.Value) (string, bool) {
	if v == nil {
		return "", false
	}
	return autoreconcile.SortKey(propertymeta.ValueTypeDate, autoreconcile.Value{Date: v})
}

func eventPlaces(g *graphcache.Graph, v vocab, pg *placeGraph, ev *graphcache.Node) ([]ChainedPlace, error) {
	var raw []ChainedPlace
	seen := map[string]bool{}
	locations, err := follow(g, ev.ID, hops.PlacesOfEvent)
	if err != nil {
		return nil, err
	}
	for _, l := range locations {
		if seen[string(l.Neighbor)] {
			continue
		}
		seen[string(l.Neighbor)] = true
		n, err := g.Node(l.Neighbor)
		if err != nil {
			return nil, err
		}
		if n == nil {
			continue
		}
		raw = append(raw, ChainedPlace{Entity: entityOf(n), Names: toponyms(n, v.prop(vocabkey.PropertyToponym))})
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
	date, _, _, _, _, err := eventDate(v, ev)
	return date, err
}

func composePlaces(g *graphcache.Graph, v vocab, nodes []*graphcache.Node, store bool) ([]PlaceSummary, error) {
	need := needed[PlaceSummary](g, nodes, store)
	needIDs := make([][]byte, len(need))
	for i, n := range need {
		needIDs[i] = n.ID
	}
	leads, err := leadPictures(g, needIDs)
	if err != nil {
		return nil, err
	}
	fresh := map[string]PlaceSummary{}
	for _, n := range need {
		h, err := placeShell(n, v)
		if err != nil {
			return nil, err
		}
		h.PictureImageFileID = leads[string(n.ID)]
		fresh[string(n.ID)] = h
		if store {
			g.SetDisplay(n.ID, h)
		}
	}
	return finishPlaces(g, v, displayed(g, nodes, fresh, func(h PlaceSummary) PlaceSummary { return h }))
}

func placeShell(n *graphcache.Node, v vocab) (PlaceSummary, error) {
	h := PlaceSummary{Entity: entityOf(n), Names: toponyms(n, v.prop(vocabkey.PropertyToponym))}
	var err error
	if h.StartDate, _, _, _, err = keptDateAt(n, v.prop(vocabkey.PropertyStartDate)); err != nil {
		return PlaceSummary{}, err
	}
	if h.EndDate, _, _, _, err = keptDateAt(n, v.prop(vocabkey.PropertyEndDate)); err != nil {
		return PlaceSummary{}, err
	}
	return h, nil
}

func finishPlaces(g *graphcache.Graph, v vocab, summaries []PlaceSummary) ([]PlaceSummary, error) {
	if len(summaries) == 0 {
		return nil, nil
	}
	ids := make([][]byte, len(summaries))
	for i := range summaries {
		ids[i] = summaries[i].Entity.ID
	}
	pg, err := buildPlaceGraph(g, v, ids)
	if err != nil {
		return nil, err
	}
	today := TodayDate()
	for i := range summaries {
		pg.byID[string(summaries[i].Entity.ID)] = summaries[i]
		pg.period[string(summaries[i].Entity.ID)] = periodWindow(summaries[i].StartDate, summaries[i].EndDate)
		at := chainQueryDate(summaries[i], today)
		summaries[i].Parents, summaries[i].ParentsAreCandidates = pg.ParentChain(summaries[i].Entity.ID, &at)
	}
	return summaries, nil
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
	linked := map[string]bool{}
	for round := 0; round < summaryChainDepth && len(frontier) > 0; round++ {
		if err := g.Nodes(frontier); err != nil {
			return nil, err
		}
		var assoc, next [][]byte
		for _, id := range frontier {
			n, err := g.Node(id)
			if err != nil {
				return nil, err
			}
			if n == nil {
				continue
			}
			shell, err := placeShell(n, v)
			if err != nil {
				return nil, err
			}
			pg.byID[string(id)] = shell
			pg.period[string(id)] = periodWindow(shell.StartDate, shell.EndDate)
			links, err := follow(g, id, hops.ParentsOfPlace)
			if err != nil {
				return nil, err
			}
			for _, l := range links {
				assoc = append(assoc, l.Association)
				if !seen[string(l.Neighbor)] {
					seen[string(l.Neighbor)] = true
					next = append(next, l.Neighbor)
				}
			}
		}
		if err := g.Nodes(assoc); err != nil {
			return nil, err
		}
		for _, id := range frontier {
			n, err := g.Node(id)
			if err != nil {
				return nil, err
			}
			if n == nil {
				continue
			}
			links, err := follow(g, id, hops.ParentsOfPlace)
			if err != nil {
				return nil, err
			}
			for _, l := range links {
				if err := rememberPlaceLink(pg, g, v, id, l, true, linked); err != nil {
					return nil, err
				}
			}
		}
		frontier = next
	}
	return pg, nil
}

// placeStep follows each hop once from id and keeps those edges. A parent
// chain climbs; a parts or succession question does not.
func placeStep(c *database.Catalog, id []byte, hopList ...hops.Hop) (*placeGraph, error) {
	pg := newPlaceGraph()
	if c == nil || c.Graph() == nil || len(id) != 16 {
		return pg, nil
	}
	g := c.Graph()
	v, err := vocabulary(g)
	if err != nil {
		return nil, err
	}
	if _, err := g.Node(id); err != nil {
		return nil, err
	}
	if err := rememberShell(pg, g, v, id); err != nil {
		return nil, err
	}
	linked := map[string]bool{}
	var neighbors, assocs [][]byte
	type pending struct {
		id   []byte
		l    graphcache.Link
		part bool
	}
	var wait []pending
	for _, hop := range hopList {
		links, err := follow(g, id, hop)
		if err != nil {
			return nil, err
		}
		_, term, filtered := hop.Filter()
		part := filtered && term == "part_of"
		for _, l := range links {
			neighbors = append(neighbors, l.Neighbor)
			assocs = append(assocs, l.Association)
			wait = append(wait, pending{id: id, l: l, part: part})
		}
	}
	if err := g.Nodes(append(neighbors, assocs...)); err != nil {
		return nil, err
	}
	for _, n := range neighbors {
		if err := rememberShell(pg, g, v, n); err != nil {
			return nil, err
		}
	}
	for _, p := range wait {
		if err := rememberPlaceLink(pg, g, v, p.id, p.l, p.part, linked); err != nil {
			return nil, err
		}
	}
	return pg, nil
}

func rememberShell(pg *placeGraph, g *graphcache.Graph, v vocab, id []byte) error {
	n, err := g.Node(id)
	if err != nil {
		return err
	}
	if n == nil {
		return nil
	}
	shell, err := placeShell(n, v)
	if err != nil {
		return err
	}
	pg.byID[string(id)] = shell
	pg.period[string(id)] = periodWindow(shell.StartDate, shell.EndDate)
	return nil
}

func rememberPlaceLink(pg *placeGraph, g *graphcache.Graph, v vocab, id []byte, l graphcache.Link, part bool, linked map[string]bool) error {
	from, to := id, l.Neighbor
	if !l.FromEnd {
		from, to = l.Neighbor, id
	}
	key := string(from) + "\x00" + string(to) + "\x00" + string(l.Association)
	if linked[key] {
		return nil
	}
	linked[key] = true
	assoc, err := g.Node(l.Association)
	if err != nil {
		return err
	}
	if assoc == nil {
		return fmt.Errorf("conclusionsummaries: association %x is missing", l.Association)
	}
	d, err := assocDatesOf(v, assoc)
	if err != nil {
		return err
	}
	link := placeLink{
		from: append([]byte(nil), from...), to: append([]byte(nil), to...),
		association: append([]byte(nil), l.Association...), start: d.start, end: d.end,
	}
	if part {
		pg.partFrom[string(from)] = append(pg.partFrom[string(from)], link)
		pg.partTo[string(to)] = append(pg.partTo[string(to)], link)
		return nil
	}
	pg.succFrom[string(from)] = append(pg.succFrom[string(from)], link)
	pg.succTo[string(to)] = append(pg.succTo[string(to)], link)
	return nil
}

func assocDatesOf(v vocab, n *graphcache.Node) (assocDates, error) {
	start, _, _, _, err := keptDateAt(n, v.prop(vocabkey.PropertyStartDate))
	if err != nil {
		return assocDates{}, err
	}
	end, _, _, _, err := keptDateAt(n, v.prop(vocabkey.PropertyEndDate))
	if err != nil {
		return assocDates{}, err
	}
	return assocDates{start: start, end: end}, nil
}
