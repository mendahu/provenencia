package propertycompare_test

import (
	"strings"

	"github.com/mendahu/provenencia/core/datevalue"
	"github.com/mendahu/provenencia/core/datevalue/datevaluetest"
	"github.com/mendahu/provenencia/core/namevalue/namevaluetest"
	"github.com/mendahu/provenencia/core/propertycompare"
	"github.com/mendahu/provenencia/core/valuecompare"
)

// The builders leave Type empty: Compare sets it from the property.

func term(key string) propertycompare.Value {
	return propertycompare.Value{Value: valuecompare.Value{Term: key}}
}

func text(s string) propertycompare.Value {
	return propertycompare.Value{Value: valuecompare.Value{Text: s}}
}

func name(form string) propertycompare.Value {
	return propertycompare.Value{Value: valuecompare.Value{Name: namevaluetest.Western(form)}}
}

func date(d *datevalue.Value) propertycompare.Value {
	return propertycompare.Value{Value: valuecompare.Value{Date: d}}
}

func neg(v propertycompare.Value) propertycompare.Value {
	v.Negative = true
	return v
}

func neutral(v propertycompare.Value) propertycompare.Value {
	v.Neutral = true
	return v
}

func ref(v propertycompare.Value, b byte) propertycompare.Value {
	v.Ref = []byte{b}
	return v
}

// label names a value for a failure message and for an expected pair.
func label(v propertycompare.Value) string {
	switch {
	case v.Term != "":
		return v.Term
	case v.Text != "":
		return v.Text
	case v.Name != nil:
		return v.Name.Form
	case v.Date != nil:
		if v.Date.Phrase != "" {
			return v.Date.Phrase
		}
		return datevalue.CompactDisplay(*v.Date)
	}
	return ""
}

func labels(vs []propertycompare.Value) string {
	out := make([]string, len(vs))
	for i, v := range vs {
		out[i] = label(v)
		if v.Negative {
			out[i] = "¬" + out[i]
		}
		if v.Neutral {
			out[i] += " (neutral)"
		}
	}
	return "{" + strings.Join(out, ", ") + "}"
}

// dates builds a side from date fixtures.
func dates(ds ...*datevalue.Value) []propertycompare.Value {
	out := make([]propertycompare.Value, len(ds))
	for i, d := range ds {
		out[i] = date(d)
	}
	return out
}

var (
	y1817     = datevaluetest.YMD(1817, 0, 0)
	y1818     = datevaluetest.YMD(1818, 0, 0)
	y1830     = datevaluetest.YMD(1830, 0, 0)
	about1817 = datevaluetest.About(1817, 0, 0)
	spring    = datevaluetest.Phrase("spring")
)
