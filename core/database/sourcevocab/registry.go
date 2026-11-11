package sourcevocab

import (
	"github.com/mendahu/provenencia/core/database/metadatafields"
	"github.com/mendahu/provenencia/core/vocabkey"
)

// Declarative provenencia seed registry (create-time starter for new catalogs).

type seedType struct {
	Key, Label, Description, IconKey string
}

type seedField struct {
	Key, Label, DataType, Description string
}

// seedSuggestion is one type's suggested fields at one level, in order; a
// field's position is its sort order.
type seedSuggestion struct {
	TypeKey   string
	Level     vocabkey.Level
	FieldKeys []string
}

var seedTypes = []seedType{
	{
		Key: "birth_certificate", Label: "Birth certificate",
		Description: "Civil or parish record of a birth (certificate, register entry, abstract, etc.).",
		IconKey:     "type_certificate",
	},
}

var seedFields = []seedField{
	{Key: "document_number", Label: "Document number", DataType: metadatafields.DataTypeText},
	{Key: "record_date", Label: "Record date", DataType: metadatafields.DataTypeText},
	{Key: "issue_date", Label: "Issue date", DataType: metadatafields.DataTypeText},
}

var seedSuggestions = []seedSuggestion{
	{TypeKey: "birth_certificate", Level: vocabkey.LevelSource, FieldKeys: []string{"document_number", "record_date", "issue_date"}},
}
