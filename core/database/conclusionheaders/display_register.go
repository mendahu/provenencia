package conclusionheaders

import (
	"fmt"
	"reflect"

	"github.com/mendahu/provenencia/core/database/graphcache"
)

func init() {
	graphcache.SetDisplayDeps(func(q graphcache.Querier, ids [][]byte) ([][]byte, error) {
		return Dependents(q, ids)
	})
	graphcache.SetDisplayCheck(checkDisplay)
}

func checkDisplay(g *graphcache.Graph, id []byte, stored any) error {
	v, err := vocabulary(g)
	if err != nil {
		return err
	}
	switch row := stored.(type) {
	case PersonHeader:
		got, err := persons(g, [][]byte{id}, false, false)
		if err != nil {
			return err
		}
		if len(got) != 1 || !reflect.DeepEqual(got[0], row) {
			return fmt.Errorf("conclusionheaders: stored person %x does not match a fresh row", id)
		}
	case EventHeader:
		got, err := events(g, [][]byte{id}, false, false)
		if err != nil {
			return err
		}
		if len(got) != 1 || !reflect.DeepEqual(got[0], finishEvent(v, row)) {
			return fmt.Errorf("conclusionheaders: stored event %x does not match a fresh row", id)
		}
	case PlaceHeader:
		got, err := places(g, [][]byte{id}, false, false)
		if err != nil {
			return err
		}
		if len(got) != 1 {
			return fmt.Errorf("conclusionheaders: stored place %x is missing", id)
		}
		fresh := got[0]
		fresh.Parents, fresh.ParentsAreCandidates = nil, false
		row.Parents, row.ParentsAreCandidates = nil, false
		if !reflect.DeepEqual(fresh, row) {
			return fmt.Errorf("conclusionheaders: stored place %x does not match a fresh row", id)
		}
	default:
		return fmt.Errorf("conclusionheaders: stored row %x has type %T", id, stored)
	}
	return nil
}
