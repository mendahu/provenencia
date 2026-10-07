package eventtitle_test

import (
	"testing"

	"github.com/mendahu/provenencia/core/database/namevalues"
	"github.com/mendahu/provenencia/core/eventtitle"
)

func named(forms ...string) []eventtitle.Subject {
	var out []eventtitle.Subject
	for _, f := range forms {
		if f == "" {
			out = append(out, eventtitle.Subject{})
			continue
		}
		out = append(out, eventtitle.Subject{Name: &namevalues.Value{Form: f}})
	}
	return out
}

func TestChoose(t *testing.T) {
	tests := []struct {
		name string
		in   eventtitle.Parts
		want eventtitle.Rule
	}{
		{name: "a recorded name wins over subjects", in: eventtitle.Parts{RecordedName: "The Great Fire", Subjects: named("James Robins")}, want: eventtitle.RuleRecordedName},
		{name: "a blank recorded name does not count", in: eventtitle.Parts{RecordedName: "  ", Ref: "EVT-1"}, want: eventtitle.RuleRef},
		{name: "one subject", in: eventtitle.Parts{TypeKey: "birth", Subjects: named("James Robins")}, want: eventtitle.RuleSubject},
		{name: "an unnamed subject is still a subject", in: eventtitle.Parts{TypeKey: "birth", Subjects: named("")}, want: eventtitle.RuleSubject},
		{name: "two subjects of a marriage are a couple", in: eventtitle.Parts{TypeKey: "marriage", Subjects: named("James", "Mary")}, want: eventtitle.RuleCouple},
		{name: "three subjects of a marriage are et al.", in: eventtitle.Parts{TypeKey: "marriage", Subjects: named("A", "B", "C")}, want: eventtitle.RuleSubjects},
		{name: "two subjects of a census are et al.", in: eventtitle.Parts{TypeKey: "census", Subjects: named("A", "B")}, want: eventtitle.RuleSubjects},
		{name: "subjects win over the label", in: eventtitle.Parts{Label: "Grandpa's birth", Subjects: named("A")}, want: eventtitle.RuleSubject},
		{name: "the label wins over type and place", in: eventtitle.Parts{Label: "Grandpa's house fire", TypeKey: "fire", Place: "York"}, want: eventtitle.RuleLabel},
		{name: "type at place", in: eventtitle.Parts{TypeKey: "fire", TypeLabel: "Fire", Place: "York"}, want: eventtitle.RuleTypeAtPlace},
		{name: "a place with no type is still at place", in: eventtitle.Parts{Place: "York"}, want: eventtitle.RuleTypeAtPlace},
		{name: "type only", in: eventtitle.Parts{TypeLabel: "Fire"}, want: eventtitle.RuleType},
		{name: "nothing but a ref", in: eventtitle.Parts{Ref: "EVT-9ZZ02"}, want: eventtitle.RuleRef},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := eventtitle.Choose(tt.in).Rule; got != tt.want {
				t.Fatalf("got %s want %s", got, tt.want)
			}
		})
	}
	t.Run("parts are trimmed", func(t *testing.T) {
		p := eventtitle.Choose(eventtitle.Parts{TypeLabel: " Fire ", Place: " York "}).Parts
		if p.TypeLabel != "Fire" || p.Place != "York" {
			t.Fatalf("%+v", p)
		}
	})
}
