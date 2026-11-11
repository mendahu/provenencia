// Package valuerules builds a catalog's value-compare Rules: the registry's
// value section with the catalog's default name pattern and alias tables in
// it. Every caller that compares values against a catalog gets its rules here,
// so a new input to Rules is loaded in one place.
package valuerules

import (
	"sync"

	"github.com/mendahu/provenencia/core/database/graphcache"
	"github.com/mendahu/provenencia/core/database/namealiases"
	"github.com/mendahu/provenencia/core/database/nameformats"
	"github.com/mendahu/provenencia/core/matchregistry"
	"github.com/mendahu/provenencia/core/propertymeta"
	"github.com/mendahu/provenencia/core/subjectcompare"
	"github.com/mendahu/provenencia/core/valuecompare"
	"github.com/mendahu/provenencia/core/vocabkey"
)

// Load returns the catalog's rules and the name pattern key its name values
// carry. With no stored default name format, Names.Patterns stays empty and the
// key is western. A nil graph is database.ErrClosed.
func Load(g *graphcache.Graph) (rules valuecompare.Rules, namePattern string, err error) {
	loaded, err := nameformats.LoadFor(g)
	if err != nil {
		return valuecompare.Rules{}, "", err
	}
	rules, err = namealiases.Rules(g, loaded.Rules(matchregistry.Default().Value))
	if err != nil {
		return valuecompare.Rules{}, "", err
	}
	return rules, loaded.Key, nil
}

// Product returns the rules the reconciler cache reads: the shipped registry
// and the product alias seed, never a catalog's own rows or name formats. The
// cache then changes only when autoreconciler.CacheVersion does.
// It is built once; every call returns the same value, which callers must not mutate.
func Product() valuecompare.Rules { return product() }

var product = sync.OnceValue(func() valuecompare.Rules {
	rules := matchregistry.Default().Value
	rules.Names.Aliases = namealiases.ProductIndex()
	return rules
})

// Metas is the catalog's value type and cardinality of every product
// property.
func Metas(g *graphcache.Graph) (subjectcompare.Metas, error) {
	v, err := g.Vocabulary()
	if err != nil {
		return nil, err
	}
	out := subjectcompare.Metas{}
	for _, p := range v.Properties {
		if p.Origin != vocabkey.OriginProvenencia {
			continue
		}
		out[vocabkey.PropertyKey(p.Key)] = propertymeta.Meta{ValueType: p.ValueType, Cardinality: p.Cardinality}
	}
	return out, nil
}
