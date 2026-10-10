package conclusiondetails

import (
	"fmt"
	"reflect"

	"github.com/mendahu/provenencia/core/database/graphcache"
)

func init() {
	graphcache.SetDetailCheck(checkDetail)
}

// valueMemo is one handle's auto-reconciler ranks. Labels, state, and
// outcomes are applied when the page is read.
type valueMemo struct {
	ranks map[string][]storedRank
}

type storedRank struct {
	Rank       int
	Reason     string
	Support    int
	Against    int
	Text       string
	HasText    bool
	Integer    int64
	HasInteger bool
	TermID     []byte
	Date       []byte
	Name       []byte
}

func valueMemoOf(g *graphcache.Graph, n *graphcache.Node) valueMemo {
	if stored, ok := g.Detail(n.ID); ok {
		if memo, is := stored.(valueMemo); is {
			return memo
		}
	}
	memo := ranksFrom(n)
	g.SetDetail(n.ID, memo)
	return memo
}

func ranksFrom(n *graphcache.Node) valueMemo {
	memo := valueMemo{ranks: map[string][]storedRank{}}
	for prop, values := range n.Values {
		ranks := make([]storedRank, 0, len(values))
		for _, v := range values {
			ranks = append(ranks, storedRank{
				Rank: v.Rank, Reason: v.Reason,
				Support: v.Support, Against: v.Against,
				Text: v.Text, HasText: v.HasText,
				Integer: v.Integer, HasInteger: v.HasInteger,
				TermID: cloneBytes(v.TermID),
				Date:   cloneBytes(v.Date),
				Name:   cloneBytes(v.Name),
			})
		}
		memo.ranks[prop] = ranks
	}
	return memo
}

func (m valueMemo) termIDs() [][]byte {
	var ids [][]byte
	for _, ranks := range m.ranks {
		for _, rank := range ranks {
			if len(rank.TermID) > 0 {
				ids = append(ids, rank.TermID)
			}
		}
	}
	return ids
}

func cloneBytes(b []byte) []byte {
	if len(b) == 0 {
		return nil
	}
	return append([]byte(nil), b...)
}

func checkDetail(g *graphcache.Graph, id []byte, stored any) error {
	memo, ok := stored.(valueMemo)
	if !ok {
		return fmt.Errorf("conclusiondetails: stored detail %x has type %T", id, stored)
	}
	n, err := g.Node(id)
	if err != nil {
		return err
	}
	if n == nil || !reflect.DeepEqual(memo, ranksFrom(n)) {
		return fmt.Errorf("conclusiondetails: stored values %x do not match the node", id)
	}
	return nil
}
