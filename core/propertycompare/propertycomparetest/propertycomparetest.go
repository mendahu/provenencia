// Package propertycomparetest builds property compare values for test fixtures.
package propertycomparetest

import (
	"bytes"
	"strings"

	"github.com/mendahu/provenencia/core/datevalue"
	"github.com/mendahu/provenencia/core/depictionvalue"
	"github.com/mendahu/provenencia/core/namevalue/namevaluetest"
	"github.com/mendahu/provenencia/core/propertycompare"
	"github.com/mendahu/provenencia/core/valuecompare"
)

// The builders leave Type empty: Compare sets it from the property.

// Term is a term value.
func Term(key string) propertycompare.Value {
	return propertycompare.Value{Value: valuecompare.Value{Term: key}}
}

// Text is a text value.
func Text(s string) propertycompare.Value {
	return propertycompare.Value{Value: valuecompare.Value{Text: s}}
}

// Name is a Western name value.
func Name(form string) propertycompare.Value {
	return propertycompare.Value{Value: valuecompare.Value{Name: namevaluetest.Western(form)}}
}

// Date is a date value.
func Date(d *datevalue.Value) propertycompare.Value {
	return propertycompare.Value{Value: valuecompare.Value{Date: d}}
}

// Depiction is a likeness value whose 16-byte image File id is id repeated.
func Depiction(id byte) propertycompare.Value {
	return propertycompare.Value{Value: valuecompare.Value{Depiction: &depictionvalue.Value{ImageFileID: bytes.Repeat([]byte{id}, 16)}}}
}

// Neg marks v negative.
func Neg(v propertycompare.Value) propertycompare.Value {
	v.Negative = true
	return v
}

// Neutral marks v neutral.
func Neutral(v propertycompare.Value) propertycompare.Value {
	v.Neutral = true
	return v
}

// Ref gives v a one-byte Observation ref.
func Ref(v propertycompare.Value, b byte) propertycompare.Value {
	v.Ref = []byte{b}
	return v
}

// Label names a value for a failure message and for an expected pair.
func Label(v propertycompare.Value) string {
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

// Labels names a side's values for a failure message.
func Labels(vs []propertycompare.Value) string {
	out := make([]string, len(vs))
	for i, v := range vs {
		out[i] = Label(v)
		if v.Negative {
			out[i] = "¬" + out[i]
		}
		if v.Neutral {
			out[i] += " (neutral)"
		}
	}
	return "{" + strings.Join(out, ", ") + "}"
}

// Dates builds a side from date fixtures.
func Dates(ds ...*datevalue.Value) []propertycompare.Value {
	out := make([]propertycompare.Value, len(ds))
	for i, d := range ds {
		out[i] = Date(d)
	}
	return out
}
