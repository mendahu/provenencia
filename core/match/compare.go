package match

import "github.com/mendahu/provenencia/core/valuecompare"

// Compare is a and b, values of valueType, under rules: the similarity value
// compare gives, or not comparable when either value is neutral, either is
// not a value of valueType, or the relation is Unknown. The neutral check
// moves to property compare (S11-09).
func Compare(rules valuecompare.Rules, valueType string, a, b Value) (similarity float64, comparable bool) {
	if a.Neutral || b.Neutral || a.valueType() != valueType || b.valueType() != valueType {
		return 0, false
	}
	res, _ := valuecompare.Compare(a.compareValue(valueType), b.compareValue(valueType), rules)
	if res.Relation == valuecompare.Unknown {
		return 0, false
	}
	return res.Similarity, true
}

// compareValue is the value in the shape value compare reads.
func (v Value) compareValue(valueType string) valuecompare.Value {
	return valuecompare.Value{
		Type: valueType, Text: v.Text, Term: v.Term, Integer: v.Integer,
		Name: v.Name, NamePattern: v.NamePattern, Date: v.Date,
	}
}
