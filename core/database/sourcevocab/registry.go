package sourcevocab

import (
	"github.com/Jake-Robins-Studios/provenencia/core/database/sourcetypes"
	"github.com/Jake-Robins-Studios/provenencia/core/vocabkey"
)

// Declarative provenencia seed registry (create-time starter for new catalogs):
// the citation vocabulary of the Spike 15 map, section 6.8. The types and their
// formats live here; the fields are in registry_fields.go and the suggestions
// in registry_suggestions.go.

// seedType is one Source type. Formats name fields by key; Install stores them
// by id through sourcetypes.ToStored.
type seedType struct {
	Key, Label, Description, IconKey string
	Formats                          sourcetypes.Formats
}

// seedField is one metadata field. UnknownText is what a citation shows when a
// researcher marks the value unknown ("" for none).
type seedField struct {
	Key, Label, DataType, Description, UnknownText string
}

// seedSuggestion is one type's suggested fields at one level, in order; a
// field's position is its sort order.
type seedSuggestion struct {
	TypeKey   string
	Level     vocabkey.Level
	FieldKeys []string
}

// certificateFormats is the one pattern of the three civil certificates: word
// is the certificate in a note ("birth certificate") and plural opens its list
// entry ("Birth certificates").
func certificateFormats(word, plural string) sourcetypes.Formats {
	return sourcetypes.Formats{
		Full:  `{source.meta.jurisdiction}, [{source.meta.creator}, ]` + word + `[ no. {source.meta.document_number}][ ({source.date:full})][, {graph.meta.item|source.meta.item}][, vol. {source.meta.volume}][, p. {citation.page}][, entry {graph.meta.entry_number|source.meta.entry_number}][; {artifact.meta.format}, *{artifact.meta.provider}*][ ({artifact.meta.url}[ : accessed {artifact.meta.accessed_date}])][; {source.meta.repository}].`,
		Short: `{source.meta.jurisdiction}, [{source.meta.creator}, ]` + word + `[ no. {source.meta.document_number}][, {graph.meta.item|source.meta.item}].`,
		List:  `{source.meta.jurisdiction}.[ {source.meta.creator}.] ` + plural + `.[ {source.meta.repository}.]`,
	}
}

// journalArticleFormats is shared by the journal article and the broadcast
// (there published_in is the programme and publisher the network).
var journalArticleFormats = sourcetypes.Formats{
	Full:  `{source.meta.author}, "{source.title}," *{source.meta.published_in}*[ {source.meta.volume}][, no. {source.meta.issue}][ ({source.date:full})][: {source.meta.page_range}][, specifically {citation.page}].`,
	Short: `{source.meta.author}, "{source.meta.short_title|source.title}"[, {citation.page}].`,
	List:  `{source.meta.author_sort|source.meta.author}. "{source.title}." *{source.meta.published_in}*[ {source.meta.volume}][, no. {source.meta.issue}][ ({source.date:full})][: {source.meta.page_range}].`,
}

var seedTypes = []seedType{
	{
		Key: "census", Label: "Census",
		Description: "A population, agricultural or other schedule from a government census.",
		IconKey:     "type_census",
		Formats: sourcetypes.Formats{
			Full:  `{source.date:year} [{source.meta.creator} ]census, {source.meta.jurisdiction}[, {source.meta.schedule}][, ED {source.meta.enumeration_district}][, piece {source.meta.piece}][, fol. {graph.meta.folio|source.meta.folio}][, p. {citation.page}][, schedule {graph.meta.schedule_number|source.meta.schedule_number}][, dwelling {graph.meta.dwelling_number|source.meta.dwelling_number}][, family {graph.meta.family_number|source.meta.family_number}][, line {graph.meta.line_number|source.meta.line_number}][, {graph.meta.item|source.meta.item}][; {artifact.meta.format}, *{artifact.meta.provider}*][ ({artifact.meta.url}[ : accessed {artifact.meta.accessed_date}])][; citing {artifact.meta.citing}].`,
			Short: `{source.date:year} [{source.meta.creator} ]census, {source.meta.jurisdiction}[, ED {source.meta.enumeration_district}][, fol. {graph.meta.folio|source.meta.folio}][, p. {citation.page}][, dwell. {graph.meta.dwelling_number|source.meta.dwelling_number}][, fam. {graph.meta.family_number|source.meta.family_number}][, {graph.meta.item|source.meta.item}].`,
			List:  `{source.meta.jurisdiction}. {source.date:year} [{source.meta.creator} ]census[, {source.meta.schedule}].`,
		},
	},
	{
		Key: "birth_certificate", Label: "Birth certificate",
		Description: "A civil certificate or register entry of a birth.",
		IconKey:     "type_certificate",
		Formats:     certificateFormats("birth certificate", "Birth certificates"),
	},
	{
		Key: "marriage_certificate", Label: "Marriage certificate",
		Description: "A civil certificate or register entry of a marriage.",
		IconKey:     "type_certificate",
		Formats:     certificateFormats("marriage certificate", "Marriage certificates"),
	},
	{
		Key: "death_certificate", Label: "Death certificate",
		Description: "A civil certificate or register entry of a death.",
		IconKey:     "type_certificate",
		Formats:     certificateFormats("death certificate", "Death certificates"),
	},
	{
		Key: "church_register", Label: "Church register",
		Description: "A parish register, church book or bishop's transcript of baptisms, marriages and burials.",
		IconKey:     "type_scroll",
		Formats: sourcetypes.Formats{
			Full:  `{source.meta.creator}[ ({source.meta.church_location})], *{source.title}*[, vol. {source.meta.volume}][, p. {citation.page}][, fol. {graph.meta.folio|source.meta.folio}][, no. {graph.meta.entry_number|source.meta.entry_number}][, {graph.meta.item|source.meta.item}][; {artifact.meta.format}, *{artifact.meta.provider}*][ ({artifact.meta.url}[ : accessed {artifact.meta.accessed_date}])][; film {artifact.meta.film_number}[, item {artifact.meta.item_number}]][; {source.meta.repository}].`,
			Short: `{source.meta.creator}, *{source.meta.short_title|source.title}*[, vol. {source.meta.volume}][, p. {citation.page}][, no. {graph.meta.entry_number|source.meta.entry_number}].`,
			List:  `{source.meta.creator}[ ({source.meta.church_location})]. *{source.title}*[, {source.meta.coverage}].[ {source.meta.repository}.]`,
		},
	},
	{
		Key: "record_book", Label: "Record book",
		Description: "A deed, will, court, marriage or tax book kept by a public office.",
		IconKey:     "type_book",
		Formats: sourcetypes.Formats{
			Full:  `{source.meta.jurisdiction}, [{source.meta.creator}, ]{source.meta.book|source.title}[: {citation.page}][, entry {graph.meta.entry_number|source.meta.entry_number}][, {graph.meta.item|source.meta.item}][; {artifact.meta.format}, *{artifact.meta.provider}*][ ({artifact.meta.url}[ : accessed {artifact.meta.accessed_date}])][; film {artifact.meta.film_number}][; {source.meta.repository}].`,
			Short: `{source.meta.jurisdiction}, {source.meta.book|source.title}[: {citation.page}].`,
			List:  `{source.meta.jurisdiction}.[ {source.meta.creator}.] {source.title}[, {source.meta.coverage}].[ {source.meta.repository}.]`,
		},
	},
	{
		Key: "passenger_list", Label: "Passenger list",
		Description: "A ship's manifest or passenger list.",
		IconKey:     "type_document",
		Formats: sourcetypes.Formats{
			Full:  `{source.title}[, p. {citation.page}][, line {graph.meta.line_number|source.meta.line_number}][, {graph.meta.item|source.meta.item}][; {artifact.meta.format}, *{artifact.meta.provider}*][ ({artifact.meta.url}[ : accessed {artifact.meta.accessed_date}])][; citing {artifact.meta.citing}][; {artifact.meta.film_publication}[, roll {artifact.meta.roll}]].`,
			Short: `{source.meta.short_title|source.title}[, p. {citation.page}][, line {graph.meta.line_number|source.meta.line_number}].`,
			List:  `[{source.meta.creator}. ]{source.meta.collection|source.title}.[ {source.meta.repository}.]`,
		},
	},
	{
		Key: "case_file", Label: "Case file",
		Description: "The loose papers of one court, probate, naturalization or land case.",
		IconKey:     "type_folder_archive",
		Formats: sourcetypes.Formats{
			Full:  `{source.meta.jurisdiction}, [{source.meta.creator}, ]{source.title}[ {source.meta.case_number}][, {graph.meta.item|source.meta.item}][; {artifact.meta.format}, *{artifact.meta.provider}*][ ({artifact.meta.url}[ : accessed {artifact.meta.accessed_date}])][; {source.meta.repository}].`,
			Short: `{source.meta.jurisdiction}, {source.meta.short_title|source.title}[ {source.meta.case_number}][, {graph.meta.item|source.meta.item}].`,
			List:  `{source.meta.jurisdiction}.[ {source.meta.creator}.] {source.meta.collection|source.title}.[ {source.meta.repository}.]`,
		},
	},
	{
		Key: "manuscript", Label: "Archived manuscript",
		Description: "Letters, diaries, business or institutional records held by an archive.",
		IconKey:     "type_folder_archive",
		Formats: sourcetypes.Formats{
			Full:  `[{source.meta.creator}, ]{source.title}[ ({source.date:full})][, {graph.meta.item|source.meta.item}][; {source.meta.collection}][, {source.meta.call_number}][, box {source.meta.box}][, folder {source.meta.folder}][; {artifact.meta.format}, *{artifact.meta.provider}*][ ({artifact.meta.url}[ : accessed {artifact.meta.accessed_date}])]; {source.meta.repository}.`,
			Short: `[{source.meta.creator}, ]{source.meta.short_title|source.title}[, {source.meta.collection}][, box {source.meta.box}][, folder {source.meta.folder}].`,
			List:  `{source.meta.collection|source.title}[, {source.meta.coverage}].[ {source.meta.call_number}.] {source.meta.repository}.`,
		},
	},
	{
		Key: "government_file", Label: "Government file",
		Description: "A military, pension, passport or agency file held by a national archive.",
		IconKey:     "type_folder_archive",
		Formats: sourcetypes.Formats{
			Full:  `{source.title}[ ({source.date:full})][, {graph.meta.item|source.meta.item}][; {source.meta.collection}][; Record Group {source.meta.record_group}][; {artifact.meta.format}, *{artifact.meta.provider}*][ ({artifact.meta.url}[ : accessed {artifact.meta.accessed_date}])][; {artifact.meta.film_publication}[, roll {artifact.meta.roll}]]; {source.meta.repository}.`,
			Short: `{source.meta.short_title|source.title}[, {graph.meta.item|source.meta.item}].`,
			List:  `{source.meta.repository}.[ Record Group {source.meta.record_group}.] {source.meta.collection|source.title}.`,
		},
	},
	{
		Key: "cemetery_marker", Label: "Grave marker",
		Description: "A headstone or monument.",
		IconKey:     "type_grave",
		Formats: sourcetypes.Formats{
			Full:  `{source.meta.creator}[ ({source.meta.cemetery_location})], {graph.meta.item|source.meta.item|source.title}[, section {source.meta.section}][, lot {source.meta.lot}][; photographed by {source.meta.photographer}[, {artifact.meta.accessed_date}]].`,
			Short: `{source.meta.creator}[ ({source.meta.cemetery_location})], {graph.meta.item|source.meta.item|source.title}.`,
			List:  `{source.meta.cemetery_location}. {source.meta.creator}. Grave markers.`,
		},
	},
	{
		Key: "family_bible", Label: "Family Bible",
		Description: "The family record pages of a Bible.",
		IconKey:     "type_book",
		Formats: sourcetypes.Formats{
			Full:  `{source.title}[, *{source.meta.published_in}*][ ([{source.meta.publication_place}: ]{source.meta.publisher}[, {source.date:year}])][, {graph.meta.item|source.meta.item}][; held by {source.meta.holder}].`,
			Short: `{source.meta.short_title|source.title}[, {graph.meta.item|source.meta.item}].`,
			List:  `{source.title}[, {source.meta.coverage}].[ *{source.meta.published_in}*.][ Held by {source.meta.holder}.]`,
		},
	},
	{
		Key: "photograph", Label: "Photograph",
		Description: "A photograph or portrait.",
		IconKey:     "type_photograph",
		Formats: sourcetypes.Formats{
			Full:  `{source.title}[, {source.meta.medium}][, {source.date:full}][, {source.meta.photographer|source.meta.studio}][; held by {source.meta.holder}].`,
			Short: `{source.meta.short_title|source.title}[, {source.date:year}].`,
			List:  `{source.title}.[ {source.meta.medium}.][ Held by {source.meta.holder}.]`,
		},
	},
	{
		Key: "private_papers", Label: "Private papers",
		Description: "Letters, diaries, research notes or a compiled family file in private hands.",
		IconKey:     "type_document",
		Formats: sourcetypes.Formats{
			Full:  `[{source.meta.creator}, ]{source.title}[, {source.date:full}][, {graph.meta.item|source.meta.item}][; held by {source.meta.holder}].`,
			Short: `[{source.meta.creator}, ]{source.meta.short_title|source.title}[, {graph.meta.item|source.meta.item}].`,
			List:  `[{source.meta.creator}. ]{source.title}[, {source.meta.coverage}].[ Held by {source.meta.holder}.]`,
		},
	},
	{
		Key: "dna_match_report", Label: "DNA match report",
		Description: "A DNA test result or match list.",
		IconKey:     "type_dna",
		Formats: sourcetypes.Formats{
			Full:  `{source.meta.testing_company}, {source.title}[, kit {source.meta.kit_number}][, match {source.meta.match_identifier}][, {graph.meta.item|source.meta.item}][, report dated {source.date:full}][; held by {source.meta.holder}].`,
			Short: `{source.meta.testing_company}, {source.meta.short_title|source.title}[, match {source.meta.match_identifier}].`,
			List:  `{source.meta.testing_company}. {source.title}.[ Held by {source.meta.holder}.]`,
		},
	},
	{
		Key: "gedcom_file", Label: "GEDCOM file",
		Description: "A family file exported from genealogy software.",
		IconKey:     "type_gedcom",
		Formats: sourcetypes.Formats{
			Full:  `[{source.meta.creator}, ]{source.title}[, GEDCOM file exported from {source.meta.software}][, {source.date:full}][, {graph.meta.item|source.meta.item}][; held by {source.meta.holder}].`,
			Short: `[{source.meta.creator}, ]{source.meta.short_title|source.title}[, {graph.meta.item|source.meta.item}].`,
			List:  `[{source.meta.creator}. ]{source.title}.[ Held by {source.meta.holder}.]`,
		},
	},
	{
		Key: "interview", Label: "Interview",
		Description: "A recorded or published interview.",
		IconKey:     "type_oral_history",
		Formats: sourcetypes.Formats{
			Full:  `{source.meta.interviewee}[ ({source.meta.interview_place})], interview by {source.meta.interviewer}, {source.date:full}[, {graph.meta.item|source.meta.item}][; held by {source.meta.holder}].`,
			Short: `{source.meta.interviewee}, interview, {source.date:full}.`,
			List:  `{source.meta.interviewee}. Interview by {source.meta.interviewer}.[ {source.meta.interview_place},] {source.date:full}.[ Held by {source.meta.holder}.]`,
		},
	},
	{
		Key: "oral_testimony", Label: "Oral testimony",
		Description: "Informal family testimony, told to the researcher or someone else.",
		IconKey:     "type_oral_history",
		Formats: sourcetypes.Formats{
			Full:  `{source.meta.interviewee}[ ({source.meta.interview_place})][, interview by {source.meta.interviewer}], {source.date:full}[, {graph.meta.item|source.meta.item}][; held by {source.meta.holder}].`,
			Short: `{source.meta.interviewee}, oral testimony, {source.date:full}.`,
			List:  `{source.meta.interviewee}.[ Interview by {source.meta.interviewer}.][ {source.meta.interview_place},] {source.date:full}.[ Held by {source.meta.holder}.]`,
		},
	},
	{
		Key: "correspondence", Label: "Correspondence",
		Description: "A letter or email written to the researcher.",
		IconKey:     "type_email",
		Formats: sourcetypes.Formats{
			Full:  `{source.meta.creator}[ ({source.meta.written_from})] to {source.meta.recipient}, {source.title}, {source.date:full}[; held by {source.meta.holder}].`,
			Short: `{source.meta.creator} to {source.meta.recipient}, {source.date:full}.`,
			List:  `{source.meta.creator}. Correspondence to {source.meta.recipient}[, {source.meta.coverage}].[ Held by {source.meta.holder}.]`,
		},
	},
	{
		Key: "book", Label: "Book",
		Description: "A book, chapter, map, CD or film.",
		IconKey:     "type_book",
		Formats: sourcetypes.Formats{
			Full:  `{source.meta.author}[, ed. {source.meta.editor}][, comp. {source.meta.compiler}][, trans. {source.meta.translator}], *{source.title}*[, {source.meta.edition} ed.][, vol. {source.meta.volume}][, {source.meta.series}][ ([{source.meta.publication_place}: ]{source.meta.publisher}[, {source.date:year}])][, {citation.page}][; {artifact.meta.format}, *{artifact.meta.provider}*][ ({artifact.meta.url}[ : accessed {artifact.meta.accessed_date}])].`,
			Short: `{source.meta.author}, *{source.meta.short_title|source.title}*[, {citation.page}].`,
			List:  `{source.meta.author_sort|source.meta.author}. *{source.title}*.[ {source.meta.edition} ed.][ [{source.meta.publication_place}: ]{source.meta.publisher}[, {source.date:year}].]`,
		},
	},
	{
		Key: "journal_article", Label: "Journal article",
		Description: "An article in a journal, magazine or newsletter.",
		IconKey:     "type_newspaper",
		Formats:     journalArticleFormats,
	},
	{
		Key: "broadcast", Label: "Broadcast",
		Description: "A radio or television programme, or a podcast.",
		IconKey:     "type_oral_history",
		Formats:     journalArticleFormats,
	},
	{
		Key: "newspaper_item", Label: "Newspaper item",
		Description: "An article, notice or obituary in a newspaper.",
		IconKey:     "type_newspaper",
		Formats: sourcetypes.Formats{
			Full:  `"{source.title}," *{source.meta.published_in}*[ ({source.meta.publication_place})], {source.date:full}[, p. {citation.page}][, col. {graph.meta.column|source.meta.column}][; {artifact.meta.format}, *{artifact.meta.provider}*][ ({artifact.meta.url}[ : accessed {artifact.meta.accessed_date}])].`,
			Short: `"{source.meta.short_title|source.title}," *{source.meta.published_in}*, {source.date:full}.`,
			List:  `*{source.meta.published_in}*[ ({source.meta.publication_place})].`,
		},
	},
	{
		Key: "website", Label: "Website or online database",
		Description: "A web page, online database, online tree or memorial site.",
		IconKey:     "type_website",
		Formats: sourcetypes.Formats{
			Full:  `[{source.meta.creator}, ]"{source.title}," [{artifact.meta.format}, ]*{source.meta.site_name}*[ ({source.meta.url}[ : accessed {source.meta.accessed_date}])][, {graph.meta.item|source.meta.item}][; citing {artifact.meta.citing}].`,
			Short: `"{source.meta.short_title|source.title}"[, {graph.meta.item|source.meta.item}].`,
			List:  `[{source.meta.creator}. ]"{source.title}." *{source.meta.site_name}*.[ {source.meta.url}.]`,
		},
	},
	{
		Key: "statute", Label: "Statute",
		Description: "A statute, session law or code.",
		IconKey:     "type_document",
		Formats: sourcetypes.Formats{
			Full:  `{source.meta.jurisdiction}, *{source.meta.published_in}*[ {source.date:year}][, vol. {source.meta.volume}][, chap. {source.meta.chapter}][, sec. {source.meta.section}][, p. {citation.page}].`,
			Short: `{source.meta.jurisdiction}, *{source.meta.published_in}*[ {source.date:year}][, chap. {source.meta.chapter}].`,
			List:  `{source.meta.jurisdiction}. *{source.meta.published_in}*[, {source.date:year}].`,
		},
	},
	{
		Key: "court_report", Label: "Court report",
		Description: "A published report of a court case.",
		IconKey:     "type_book",
		Formats: sourcetypes.Formats{
			Full:  `*{source.title}*, [{source.meta.volume} ]{source.meta.published_in} {citation.page|source.meta.page_range}[ ({source.date:year})].`,
			Short: `*{source.meta.short_title|source.title}*[, {citation.page}].`,
			List:  `*{source.title}*. [{source.meta.volume} ]{source.meta.published_in}[ ({source.date:year})].`,
		},
	},
}
