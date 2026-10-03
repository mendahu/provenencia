package valuecodec

import (
	"reflect"
	"testing"

	"github.com/mendahu/provenencia/core/database/datevalues"
	"github.com/mendahu/provenencia/core/database/namevalues"
)

func ip(n int) *int { return &n }

func TestDateRoundTrip(t *testing.T) {
	for _, v := range []datevalues.Value{
		{Kind: datevalues.KindPoint, Calendar: "gregorian", StartYear: ip(1985), StartMonth: ip(5)},
		{Kind: datevalues.KindPoint, Qualifier: "ABT", Calendar: "gregorian", StartYear: ip(1817), StartMonth: ip(0), StartDay: ip(14)},
		{Kind: datevalues.KindRange, Calendar: "gregorian", StartYear: ip(1850), EndYear: ip(1860), EndMonth: ip(12),
			StartHour: ip(9), StartMinute: ip(30), StartSecond: ip(1), StartMillisecond: ip(250), StartTZ: "EST", EndTZ: "UTC"},
		{Kind: datevalues.KindPoint, Phrase: "the winter after the fire"},
	} {
		b, err := MarshalDate(v)
		if err != nil {
			t.Fatal(err)
		}
		got, err := UnmarshalDate(b)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(got, v) {
			t.Fatalf("round trip\n got %+v\nwant %+v", got, v)
		}
	}
}

func TestNameRoundTrip(t *testing.T) {
	v := namevalues.Value{Form: "James Robins", Parts: []namevalues.Part{
		{Idx: 0, Value: "James", Type: namevalues.PartTypeGiven},
		{Idx: 1, Value: "Robins", Type: namevalues.PartTypeSurname},
		{Idx: 2, Value: "Jr", Type: ""},
	}}
	b, err := MarshalName(v)
	if err != nil {
		t.Fatal(err)
	}
	got, err := UnmarshalName(b)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, v) {
		t.Fatalf("round trip\n got %+v\nwant %+v", got, v)
	}
	if _, err := UnmarshalName([]byte{0xff, 0xff}); err == nil {
		t.Fatal("garbage decoded")
	}
}
