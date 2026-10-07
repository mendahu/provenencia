package handlers

import (
	"github.com/mendahu/provenencia/api/proto/engine"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/sourceeventtitles"
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

func ListSourceEventTitles(in []byte) ([]byte, error) {
	var req engine.ListSourceEventTitlesRequest
	if err := proto.Unmarshal(in, &req); err != nil {
		return nil, unmarshalErr("list_source_event_titles", err)
	}
	sourceID, err := parseID(req.GetSourceId())
	if err != nil {
		return nil, err
	}
	out := &engine.ListSourceEventTitlesResponse{}
	err = withProjectCatalog(req.GetProjectDir(), func(c *database.Catalog) error {
		db, err := c.DB()
		if err != nil {
			return err
		}
		titles, err := sourceeventtitles.ForSource(db, sourceID)
		if err != nil {
			return err
		}
		for _, t := range titles {
			out.Titles = append(out.Titles, &engine.SourceEventTitle{
				SubjectId: uuidString(t.SubjectID),
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
