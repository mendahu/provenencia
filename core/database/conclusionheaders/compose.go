package conclusionheaders

import (
	"bytes"
	"sort"
	"strings"

	"github.com/mendahu/provenencia/core/autoreconcile"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/canonicalentities"
	"github.com/mendahu/provenencia/core/database/datevalues"
	"github.com/mendahu/provenencia/core/database/graphcache"
	"github.com/mendahu/provenencia/core/database/namevalues"
	"github.com/mendahu/provenencia/core/database/properties"
	"github.com/mendahu/provenencia/core/valuecodec"
)

// vocab is the label map a header resolves at read time. Term labels and
// "today" are not stored on the row, so a rename or a new day does not
// throw the row away.
type vocab struct {
	propID    map[string][]byte
	termLabel map[string]string
	typeID    map[string][]byte
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
		propID:    map[string][]byte{},
		termLabel: map[string]string{},
		typeID:    map[string][]byte{},
	}
	if err := scanVocab(g, `SELECT id, key FROM properties WHERE origin = ?`, []any{properties.OriginProvenencia}, func(id []byte, key, _ string) {
		v.propID[key] = append([]byte(nil), id...)
	}); err != nil {
		return vocab{}, err
	}
	if err := scanVocab(g, `SELECT id, label FROM property_terms`, nil, func(id []byte, label, _ string) {
		v.termLabel[string(id)] = label
	}); err != nil {
		return vocab{}, err
	}
	if err := scanVocab(g, `SELECT id, key FROM subject_types WHERE origin = ?`, []any{properties.OriginProvenencia}, func(id []byte, key, _ string) {
		v.typeID[key] = append([]byte(nil), id...)
	}); err != nil {
		return vocab{}, err
	}
	return v, nil
}

func scanVocab(g *graphcache.Graph, query string, args []any, take func(id []byte, a, b string)) error {
	rows, err := g.Query(query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var id []byte
		var a string
		if err := rows.Scan(&id, &a); err != nil {
			return err
		}
		take(id, a, "")
	}
	return rows.Err()
}

func (v vocab) prop(key string) []byte { return v.propID[key] }

func entityOf(n *graphcache.Node) canonicalentities.Entity {
	return canonicalentities.Entity{
		ID:            append([]byte(nil), n.ID...),
		SubjectTypeID: append([]byte(nil), n.SubjectTypeID...),
		Ref:           n.Ref,
		Argument:      n.Argument,
		Label:         n.Label,
	}
}

func kept(n *graphcache.Node, propID []byte) []graphcache.Value {
	if n == nil || len(propID) != 16 {
		return nil
	}
	var out []graphcache.Value
	for _, row := range n.Values[string(propID)] {
		if row.Reason == "kept" {
			out = append(out, row)
		}
	}
	return out
}

func rank1(rows []graphcache.Value) (graphcache.Value, bool) {
	for _, row := range rows {
		if row.Rank == 1 {
			return row, true
		}
	}
	return graphcache.Value{}, false
}

func follow(n *graphcache.Node, bridge, term string, fromEnd bool) []graphcache.Link {
	if n == nil {
		return nil
	}
	var out []graphcache.Link
	for _, l := range n.Links {
		if l.BridgeTypeKey != bridge || l.FromEnd != fromEnd {
			continue
		}
		if term != "" && l.TermKey != term {
			continue
		}
		out = append(out, l)
	}
	return out
}

func keptName(n *graphcache.Node, propID []byte) (*namevalues.Value, int, error) {
	rows := kept(n, propID)
	row, ok := rank1(rows)
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
	rows := kept(n, propID)
	row, ok := rank1(rows)
	if !ok || !row.HasText {
		return "", len(rows)
	}
	return strings.TrimSpace(row.Text), len(rows)
}

func keptTerm(n *graphcache.Node, propID []byte) (id []byte, key string, nkept int) {
	rows := kept(n, propID)
	row, ok := rank1(rows)
	if !ok || row.TermKey == "" {
		return nil, "", len(rows)
	}
	return append([]byte(nil), row.TermID...), row.TermKey, len(rows)
}

func keptDateAt(n *graphcache.Node, propID []byte) (*datevalues.Value, int, string, bool, error) {
	rows := kept(n, propID)
	row, ok := rank1(rows)
	if !ok {
		return nil, len(rows), "", false, nil
	}
	date, err := unmarshalDate(row.Date)
	if err != nil || date == nil {
		return nil, len(rows), "", false, err
	}
	key, ok := autoreconcile.SortKey(properties.ValueTypeDate, autoreconcile.Value{Date: date})
	return date, len(rows), key, ok, nil
}

func toponyms(n *graphcache.Node, propID []byte) []string {
	var names []string
	for _, row := range kept(n, propID) {
		name := strings.TrimSpace(row.Text)
		if row.HasText && name != "" {
			names = append(names, name)
		}
	}
	return names
}

// ListPersons returns every unmerged Person's header in list order.
func ListPersons(c *database.Catalog) ([]PersonHeader, error) {
	return persons(c.Graph(), nil, true, true)
}

// PersonsByIDs returns the headers of the given unmerged Persons in list
// order. Unknown, merged, and non-Person ids are absent.
func PersonsByIDs(c *database.Catalog, ids [][]byte) ([]PersonHeader, error) {
	if len(database.UniqueBlobIDs(ids)) == 0 {
		return nil, nil
	}
	return persons(c.Graph(), ids, false, true)
}

// PersonsFromQuerier composes person rows from q without storing them.
// Search calls this inside the write transaction.
func PersonsFromQuerier(q Querier, ids [][]byte) ([]PersonHeader, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	return persons(graphcache.Over(q), ids, false, false)
}

func persons(g *graphcache.Graph, ids [][]byte, all, store bool) ([]PersonHeader, error) {
	if g == nil {
		return nil, nil
	}
	v, err := vocabulary(g)
	if err != nil {
		return nil, err
	}
	nodes, err := nodesOf(g, v.typeID["person"], ids, all)
	if err != nil {
		return nil, err
	}
	out, err := composePersons(g, v, nodes, store)
	if err != nil {
		return nil, err
	}
	sortByTitle(out, personTitle, func(h PersonHeader) string { return h.Entity.Ref })
	return out, nil
}

// ListEvents returns every unmerged Event's header in list order.
func ListEvents(c *database.Catalog) ([]EventHeader, error) {
	return events(c.Graph(), nil, true, true)
}

// EventsByIDs returns the headers of the given unmerged Events in list order.
func EventsByIDs(c *database.Catalog, ids [][]byte) ([]EventHeader, error) {
	if len(database.UniqueBlobIDs(ids)) == 0 {
		return nil, nil
	}
	return events(c.Graph(), ids, false, true)
}

// EventsFromQuerier composes event rows from q without storing them.
func EventsFromQuerier(q Querier, ids [][]byte) ([]EventHeader, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	return events(graphcache.Over(q), ids, false, false)
}

func events(g *graphcache.Graph, ids [][]byte, all, store bool) ([]EventHeader, error) {
	if g == nil {
		return nil, nil
	}
	v, err := vocabulary(g)
	if err != nil {
		return nil, err
	}
	nodes, err := nodesOf(g, v.typeID["event"], ids, all)
	if err != nil {
		return nil, err
	}
	out, err := composeEvents(g, v, nodes, store)
	if err != nil {
		return nil, err
	}
	sort.SliceStable(out, func(i, j int) bool { return eventLess(out[i], out[j]) })
	return out, nil
}

// ListPlaces returns every unmerged Place's header in list order.
func ListPlaces(c *database.Catalog) ([]PlaceHeader, error) {
	return places(c.Graph(), nil, true, true)
}

// PlacesByIDs returns the headers of the given unmerged Places in list order.
func PlacesByIDs(c *database.Catalog, ids [][]byte) ([]PlaceHeader, error) {
	if len(database.UniqueBlobIDs(ids)) == 0 {
		return nil, nil
	}
	return places(c.Graph(), ids, false, true)
}

// PlacesFromQuerier composes place rows from q without storing them.
func PlacesFromQuerier(q Querier, ids [][]byte) ([]PlaceHeader, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	return places(graphcache.Over(q), ids, false, false)
}

func places(g *graphcache.Graph, ids [][]byte, all, store bool) ([]PlaceHeader, error) {
	if g == nil {
		return nil, nil
	}
	v, err := vocabulary(g)
	if err != nil {
		return nil, err
	}
	nodes, err := nodesOf(g, v.typeID["place"], ids, all)
	if err != nil {
		return nil, err
	}
	out, err := composePlaces(g, v, nodes, store)
	if err != nil {
		return nil, err
	}
	sortByTitle(out, placeTitle, func(h PlaceHeader) string { return h.Entity.Ref })
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
		if n == nil || n.Merged || !bytes.Equal(n.SubjectTypeID, typeID) {
			continue
		}
		out = append(out, n)
	}
	return out, nil
}

func composePersons(g *graphcache.Graph, v vocab, nodes []*graphcache.Node, store bool) ([]PersonHeader, error) {
	var need [][]byte
	var eventIDs [][]byte
	for _, n := range nodes {
		if store {
			if row, ok := g.Display(n.ID); ok {
				if _, is := row.(PersonHeader); is {
					continue
				}
			}
		}
		need = append(need, n.ID)
		for _, l := range follow(n, "participation", "subject", true) {
			eventIDs = append(eventIDs, l.Neighbor)
		}
	}
	if err := g.Nodes(eventIDs); err != nil {
		return nil, err
	}
	var placeIDs [][]byte
	for _, id := range eventIDs {
		n, err := g.Node(id)
		if err != nil {
			return nil, err
		}
		for _, l := range follow(n, "location", "", true) {
			placeIDs = append(placeIDs, l.Neighbor)
		}
	}
	pg, err := buildPlaceGraph(g, v, placeIDs)
	if err != nil {
		return nil, err
	}
	byID := map[string]*graphcache.Node{}
	for _, n := range nodes {
		byID[string(n.ID)] = n
	}
	fresh := map[string]PersonHeader{}
	for _, id := range need {
		h, err := onePerson(v, byID[string(id)], g, pg)
		if err != nil {
			return nil, err
		}
		fresh[string(id)] = h
		if store {
			g.SetDisplay(id, h)
		}
	}
	out := make([]PersonHeader, 0, len(nodes))
	for _, n := range nodes {
		if h, ok := fresh[string(n.ID)]; ok {
			out = append(out, h)
			continue
		}
		row, _ := g.Display(n.ID)
		out = append(out, row.(PersonHeader))
	}
	return out, nil
}

func onePerson(v vocab, n *graphcache.Node, g *graphcache.Graph, pg *placeGraph) (PersonHeader, error) {
	h := PersonHeader{Entity: entityOf(n)}
	name, count, err := keptName(n, v.prop("name"))
	if err != nil {
		return PersonHeader{}, err
	}
	h.Name, h.NameValueCount = name, count
	var births, deaths []lifePick
	for _, l := range follow(n, "participation", "subject", true) {
		ev, err := g.Node(l.Neighbor)
		if err != nil {
			return PersonHeader{}, err
		}
		if ev == nil {
			continue
		}
		_, key, _ := keptTerm(ev, v.prop("event_type"))
		if key != "birth" && key != "death" {
			continue
		}
		item, err := lifeFrom(v, ev)
		if err != nil {
			return PersonHeader{}, err
		}
		if key == "birth" {
			births = append(births, item)
		} else {
			deaths = append(deaths, item)
		}
	}
	var errLife error
	if h.Birth, errLife = leadLife(g, v, pg, births); errLife != nil {
		return PersonHeader{}, errLife
	}
	if h.Death, errLife = leadLife(g, v, pg, deaths); errLife != nil {
		return PersonHeader{}, errLife
	}
	return h, nil
}

type lifePick struct {
	event *graphcache.Node
	date  *datevalues.Value
	count int
	sort  string
	ok    bool
}

func lifeFrom(v vocab, ev *graphcache.Node) (lifePick, error) {
	date, dateCount, dateKey, dateOK, err := keptDateAt(ev, v.prop("date"))
	if err != nil {
		return lifePick{}, err
	}
	if date == nil {
		date, dateCount, dateKey, dateOK, err = keptDateAt(ev, v.prop("start_date"))
		if err != nil {
			return lifePick{}, err
		}
	}
	return lifePick{event: ev, date: date, count: dateCount, sort: dateKey, ok: dateOK}, nil
}

func leadLife(g *graphcache.Graph, v vocab, pg *placeGraph, items []lifePick) (LifeFacts, error) {
	if len(items) == 0 {
		return LifeFacts{}, nil
	}
	lead := items[0]
	for _, item := range items[1:] {
		if lifeBefore(item, lead) {
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

func lifeBefore(a, b lifePick) bool {
	if a.ok != b.ok {
		return a.ok
	}
	if a.sort != b.sort {
		return a.sort < b.sort
	}
	return a.event.Ref < b.event.Ref
}

func composeEvents(g *graphcache.Graph, v vocab, nodes []*graphcache.Node, store bool) ([]EventHeader, error) {
	var need []*graphcache.Node
	var personIDs, placeIDs [][]byte
	for _, n := range nodes {
		if store {
			if row, ok := g.Display(n.ID); ok {
				if _, is := row.(EventHeader); is {
					continue
				}
			}
		}
		need = append(need, n)
		for _, l := range follow(n, "participation", "subject", false) {
			personIDs = append(personIDs, l.Neighbor)
		}
		for _, l := range follow(n, "location", "", true) {
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
	fresh := map[string]EventHeader{}
	for _, n := range need {
		h, err := oneEvent(g, v, pg, n)
		if err != nil {
			return nil, err
		}
		fresh[string(n.ID)] = h
		if store {
			g.SetDisplay(n.ID, h)
		}
	}
	out := make([]EventHeader, 0, len(nodes))
	for _, n := range nodes {
		h, ok := fresh[string(n.ID)]
		if !ok {
			row, _ := g.Display(n.ID)
			h = row.(EventHeader)
		}
		out = append(out, finishEvent(v, h))
	}
	return out, nil
}

func oneEvent(g *graphcache.Graph, v vocab, pg *placeGraph, n *graphcache.Node) (EventHeader, error) {
	h := EventHeader{Entity: entityOf(n)}
	h.EventName, h.EventNameCount = keptText(n, v.prop("event_name"))
	if id, key, count := keptTerm(n, v.prop("event_type")); key != "" {
		h.EventType = &EventType{ID: id, Key: key}
		h.EventTypeCount = count
	}
	var err error
	if h.Date, h.DateCount, _, _, err = keptDateAt(n, v.prop("date")); err != nil {
		return EventHeader{}, err
	}
	if h.StartDate, h.StartDateCount, _, _, err = keptDateAt(n, v.prop("start_date")); err != nil {
		return EventHeader{}, err
	}
	if h.EndDate, h.EndDateCount, _, _, err = keptDateAt(n, v.prop("end_date")); err != nil {
		return EventHeader{}, err
	}
	h.Places, err = eventPlaces(g, v, pg, n)
	if err != nil {
		return EventHeader{}, err
	}
	seen := map[string]bool{}
	for _, l := range follow(n, "participation", "subject", false) {
		if seen[string(l.Neighbor)] {
			continue
		}
		seen[string(l.Neighbor)] = true
		person, err := g.Node(l.Neighbor)
		if err != nil {
			return EventHeader{}, err
		}
		if person == nil {
			continue
		}
		name, count, err := keptName(person, v.prop("name"))
		if err != nil {
			return EventHeader{}, err
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

func finishEvent(v vocab, h EventHeader) EventHeader {
	if h.EventType != nil {
		et := *h.EventType
		et.Label = v.termLabel[string(et.ID)]
		h.EventType = &et
	}
	h.Title = eventTitle(h)
	return h
}

type eventKey struct {
	key string
	ok  bool
	ref string
}

func eventLess(a, b EventHeader) bool {
	ak, bk := eventKeyOf(a), eventKeyOf(b)
	if ak.ok != bk.ok {
		return ak.ok
	}
	if ak.key != bk.key {
		return ak.key < bk.key
	}
	return ak.ref < bk.ref
}

func eventKeyOf(h EventHeader) eventKey {
	if key, ok := dateSort(h.Date); ok {
		return eventKey{key: key, ok: true, ref: strings.ToLower(h.Entity.Ref)}
	}
	if key, ok := dateSort(h.StartDate); ok {
		return eventKey{key: key, ok: true, ref: strings.ToLower(h.Entity.Ref)}
	}
	return eventKey{ref: strings.ToLower(h.Entity.Ref)}
}

func dateSort(v *datevalues.Value) (string, bool) {
	if v == nil {
		return "", false
	}
	return autoreconcile.SortKey(properties.ValueTypeDate, autoreconcile.Value{Date: v})
}

func eventPlaces(g *graphcache.Graph, v vocab, pg *placeGraph, ev *graphcache.Node) ([]HeaderPlace, error) {
	var raw []HeaderPlace
	seen := map[string]bool{}
	for _, l := range follow(ev, "location", "", true) {
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
		raw = append(raw, HeaderPlace{Entity: entityOf(n), Names: toponyms(n, v.prop("toponym"))})
	}
	if len(raw) == 0 {
		return nil, nil
	}
	at := eventWhen(v, ev)
	return pg.FoldLocations(raw, at), nil
}

func eventWhen(v vocab, ev *graphcache.Node) *datevalues.Value {
	if date, _, _, _, err := keptDateAt(ev, v.prop("date")); err == nil && date != nil {
		return date
	}
	date, _, _, _, _ := keptDateAt(ev, v.prop("start_date"))
	return date
}

func composePlaces(g *graphcache.Graph, v vocab, nodes []*graphcache.Node, store bool) ([]PlaceHeader, error) {
	var need []*graphcache.Node
	for _, n := range nodes {
		if store {
			if row, ok := g.Display(n.ID); ok {
				if _, is := row.(PlaceHeader); is {
					continue
				}
			}
		}
		need = append(need, n)
	}
	fresh := map[string]PlaceHeader{}
	for _, n := range need {
		h := placeShell(n, v)
		fresh[string(n.ID)] = h
		if store {
			g.SetDisplay(n.ID, h)
		}
	}
	out := make([]PlaceHeader, 0, len(nodes))
	for _, n := range nodes {
		h, ok := fresh[string(n.ID)]
		if !ok {
			row, _ := g.Display(n.ID)
			h = row.(PlaceHeader)
		}
		out = append(out, h)
	}
	return finishPlaces(g, v, out)
}

func placeShell(n *graphcache.Node, v vocab) PlaceHeader {
	h := PlaceHeader{Entity: entityOf(n), Names: toponyms(n, v.prop("toponym"))}
	h.StartDate, _, _, _, _ = keptDateAt(n, v.prop("start_date"))
	h.EndDate, _, _, _, _ = keptDateAt(n, v.prop("end_date"))
	return h
}

func finishPlaces(g *graphcache.Graph, v vocab, headers []PlaceHeader) ([]PlaceHeader, error) {
	if len(headers) == 0 {
		return nil, nil
	}
	ids := make([][]byte, len(headers))
	for i := range headers {
		ids[i] = headers[i].Entity.ID
	}
	pg, err := buildPlaceGraph(g, v, ids)
	if err != nil {
		return nil, err
	}
	today := TodayDate()
	for i := range headers {
		pg.byID[string(headers[i].Entity.ID)] = headers[i]
		pg.period[string(headers[i].Entity.ID)] = autoreconcile.PeriodWindow(headers[i].StartDate, headers[i].EndDate)
		at := chainQueryDate(headers[i], today)
		headers[i].Parents, headers[i].ParentsAreCandidates = pg.ParentChain(headers[i].Entity.ID, &at)
	}
	return headers, nil
}

func buildPlaceGraph(g *graphcache.Graph, v vocab, seeds [][]byte) (*placeGraph, error) {
	pg := &placeGraph{byID: map[string]PlaceHeader{}, period: map[string]autoreconcile.Window{}}
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
	for round := 0; round < headerChainDepth && len(frontier) > 0; round++ {
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
			shell := placeShell(n, v)
			pg.byID[string(id)] = shell
			pg.period[string(id)] = autoreconcile.PeriodWindow(shell.StartDate, shell.EndDate)
			for _, l := range n.Links {
				if l.BridgeTypeKey != "place_relationship" || (l.TermKey != "part_of" && l.TermKey != "succeeded_by") {
					continue
				}
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
			for _, l := range n.Links {
				if l.BridgeTypeKey != "place_relationship" {
					continue
				}
				from, to := id, l.Neighbor
				if !l.FromEnd {
					from, to = l.Neighbor, id
				}
				key := string(from) + "\x00" + string(to) + "\x00" + string(l.Association)
				if linked[key] {
					continue
				}
				linked[key] = true
				d := assocDatesOf(v, mustNode(g, l.Association))
				link := placeLink{from: append([]byte(nil), from...), to: append([]byte(nil), to...), association: append([]byte(nil), l.Association...), start: d.start, end: d.end}
				switch l.TermKey {
				case "part_of":
					pg.partOf = append(pg.partOf, link)
				case "succeeded_by":
					pg.succ = append(pg.succ, link)
				}
			}
		}
		frontier = next
	}
	return pg, nil
}

func mustNode(g *graphcache.Graph, id []byte) *graphcache.Node {
	n, err := g.Node(id)
	if err != nil {
		return nil
	}
	return n
}

func assocDatesOf(v vocab, n *graphcache.Node) assocDates {
	if n == nil {
		return assocDates{}
	}
	start, _, _, _, _ := keptDateAt(n, v.prop("start_date"))
	end, _, _, _, _ := keptDateAt(n, v.prop("end_date"))
	return assocDates{start: start, end: end}
}
