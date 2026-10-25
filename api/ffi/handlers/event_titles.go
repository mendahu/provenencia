package handlers

import (
	"github.com/mendahu/provenencia/api/proto/engine"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/grapheventtitles"
	"github.com/mendahu/provenencia/core/eventtitle"
	"github.com/mendahu/provenencia/core/valuecodec"
	"google.golang.org/protobuf/proto"
)

var eventTitleRules = map[eventtitle.Rule]engine.EventTitleRule{
	eventtitle.RuleRecordedName: engine.EventTitleRule_EVENT_TITLE_RULE_RECORDED_NAME,
	eventtitle.RuleSubject:      engine.EventTitleRule_EVENT_TITLE_RULE_SUBJECT,
	eventtitle.RuleCouple:       engine.EventTitleRule_EVENT_TITLE_RULE_COUPLE,
	eventtitle.RuleSubjects:     engine.EventTitleRule_EVENT_TITLE_RULE_SUBJECTS,
	eventtitle.RuleLabel:        engine.EventTitleRule_EVENT_TITLE_RULE_LABEL,
	eventtitle.RuleTypeAtPlace:  engine.EventTitleRule_EVENT_TITLE_RULE_TYPE_AT_PLACE,
	eventtitle.RuleType:         engine.EventTitleRule_EVENT_TITLE_RULE_TYPE,
	eventtitle.RuleRef:          engine.EventTitleRule_EVENT_TITLE_RULE_REF,
}

func eventTitleProto(p eventtitle.Plan) *engine.EventTitle {
	out := &engine.EventTitle{
		Rule:         eventTitleRules[p.Rule],
		RecordedName: p.Parts.RecordedName,
		Label:        p.Parts.Label,
		Ref:          p.Parts.Ref,
		TypeKey:      p.Parts.TypeKey,
		TypeOrigin:   p.Parts.TypeOrigin,
		TypeLabel:    p.Parts.TypeLabel,
		Place:        p.Parts.Place,
	}
	for _, s := range p.Parts.Subjects {
		sub := &engine.EventTitleSubject{}
		if s.Name != nil {
			sub.Name = valuecodec.NameToProto(*s.Name)
		}
		out.Subjects = append(out.Subjects, sub)
	}
	return out
}

func ListEvidenceGraphEventTitles(in []byte) ([]byte, error) {
	var req engine.ListEvidenceGraphEventTitlesRequest
	if err := proto.Unmarshal(in, &req); err != nil {
		return nil, unmarshalErr("list_evidence_graph_event_titles", err)
	}
	graphID, err := parseID(req.GetGraphId())
	if err != nil {
		return nil, err
	}
	out := &engine.ListEvidenceGraphEventTitlesResponse{}
	err = withProjectCatalog(req.GetProjectDir(), func(c *database.Catalog) error {
		eg, err := c.Graph().EvidenceGraph(graphID)
		if err != nil {
			return err
		}
		props, err := loadPropertyInfo(c)
		if err != nil {
			return err
		}
		vocab, err := c.Graph().Vocabulary()
		if err != nil {
			return err
		}
		keys := map[string]string{}
		for id, info := range props {
			keys[id] = info.Key
		}
		termLookup := map[string]struct{ Key, Origin, Label string }{}
		for id, info := range vocab.TermsByID() {
			termLookup[id] = struct{ Key, Origin, Label string }{Key: info.Key, Origin: info.Origin, Label: info.Label}
		}
		for _, t := range grapheventtitles.FromStored(eg, keys, termLookup) {
			out.Titles = append(out.Titles, &engine.SourceEventTitle{
				SubjectId: database.UUIDString(t.SubjectID),
				Title:     eventTitleProto(t.Plan),
			})
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return proto.Marshal(out)
}
