package sourcevocab

import (
	"github.com/mendahu/provenencia/core/database/sourcefields"
)

// Declarative provenencia seed registry (create-time starter for new catalogs).

type seedType struct {
	Key, Label, Description, IconKey string
}

type seedField struct {
	Key, Label, DataType, Description string
}

type seedSuggestion struct {
	TypeKey   string
	FieldKey  string
	SortOrder int
}

var seedTypes = []seedType{
	{
		Key: "birth_certificate", Label: "Birth certificate",
		Description: "Civil or parish record of a birth (certificate, register entry, abstract, etc.).",
		IconKey:     "type_certificate",
	},
}

var seedFields = []seedField{
	{Key: "document_number", Label: "Document number", DataType: sourcefields.DataTypeText},
	{Key: "record_date", Label: "Record date", DataType: sourcefields.DataTypeDate},
	{Key: "issue_date", Label: "Issue date", DataType: sourcefields.DataTypeDate},
}

var seedSuggestions = []seedSuggestion{
	{TypeKey: "birth_certificate", FieldKey: "document_number", SortOrder: 0},
	{TypeKey: "birth_certificate", FieldKey: "record_date", SortOrder: 1},
	{TypeKey: "birth_certificate", FieldKey: "issue_date", SortOrder: 2},
}
