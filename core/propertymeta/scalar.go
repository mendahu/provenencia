package propertymeta

// Scalar is the text, integer and boolean halves of one Property value, each
// with a flag for whether it is set. Every struct that carries a value
// embeds it, so a new scalar type adds its fields here once.
type Scalar struct {
	Text       string
	HasText    bool
	Integer    int64
	HasInteger bool
	Boolean    bool
	HasBoolean bool
}

// TextScalar is a Scalar holding the text s.
func TextScalar(s string) Scalar { return Scalar{Text: s, HasText: true} }

// IntegerScalar is a Scalar holding the integer n.
func IntegerScalar(n int64) Scalar { return Scalar{Integer: n, HasInteger: true} }

// BooleanScalar is a Scalar holding the boolean b.
func BooleanScalar(b bool) Scalar { return Scalar{Boolean: b, HasBoolean: true} }
