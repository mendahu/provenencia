package valuecompare

// ComparerTypes are the value types that have a comparer, sorted.
var ComparerTypes = comparerTypes

// WordOptions builds the options a name passes to the word table.
func WordOptions(initials bool, aliases AliasIndex, rx, ry NameRole) wordOptions {
	return wordOptions{initials: initials, aliases: aliases, roleX: rx, roleY: ry}
}

// WordSimilarity is the word table, folding both words with the rules' folds.
func WordSimilarity(x, y string, w WordRules, o wordOptions) (float64, Relation, Cause) {
	return wordSimilarity(x, y, FoldWord(x, w.Folds), FoldWord(y, w.Folds), w, o)
}

// WordOptionsType names the unexported options type for table tests.
type WordOptionsType = wordOptions
