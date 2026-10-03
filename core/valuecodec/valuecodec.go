// Package valuecodec converts structured DateValues and NameValues to and
// from their protobuf messages (DateValueInput / NameValueInput in
// engine.proto). The FFI handlers use it at the client boundary; the
// resolved-values cache uses the Marshal / Unmarshal pairs to store whole
// values without decoding them in SQL (deployment-plan Spike 9, Q12).
package valuecodec

import (
	"strings"

	"google.golang.org/protobuf/proto"

	"github.com/mendahu/provenencia/api/proto/engine"
	"github.com/mendahu/provenencia/core/database/datevalues"
	"github.com/mendahu/provenencia/core/database/namevalues"
)

// DateToProto carries a DateValue's components into its message.
func DateToProto(v datevalues.Value) *engine.DateValueInput {
	return &engine.DateValueInput{
		Kind:             v.Kind,
		Qualifier:        v.Qualifier,
		Calendar:         v.Calendar,
		StartYear:        toInt32(v.StartYear),
		StartMonth:       toInt32(v.StartMonth),
		StartDay:         toInt32(v.StartDay),
		StartHour:        toInt32(v.StartHour),
		StartMinute:      toInt32(v.StartMinute),
		StartSecond:      toInt32(v.StartSecond),
		StartMillisecond: toInt32(v.StartMillisecond),
		StartTz:          v.StartTZ,
		EndYear:          toInt32(v.EndYear),
		EndMonth:         toInt32(v.EndMonth),
		EndDay:           toInt32(v.EndDay),
		EndHour:          toInt32(v.EndHour),
		EndMinute:        toInt32(v.EndMinute),
		EndSecond:        toInt32(v.EndSecond),
		EndMillisecond:   toInt32(v.EndMillisecond),
		EndTz:            v.EndTZ,
		Phrase:           v.Phrase,
	}
}

// DateFromProto is the inverse of DateToProto. Unset components stay nil
// (unknown, never zero).
func DateFromProto(d *engine.DateValueInput) datevalues.Value {
	return datevalues.Value{
		Kind:             d.GetKind(),
		Qualifier:        d.GetQualifier(),
		Calendar:         d.GetCalendar(),
		StartYear:        toInt(d.StartYear),
		StartMonth:       toInt(d.StartMonth),
		StartDay:         toInt(d.StartDay),
		StartHour:        toInt(d.StartHour),
		StartMinute:      toInt(d.StartMinute),
		StartSecond:      toInt(d.StartSecond),
		StartMillisecond: toInt(d.StartMillisecond),
		StartTZ:          d.GetStartTz(),
		EndYear:          toInt(d.EndYear),
		EndMonth:         toInt(d.EndMonth),
		EndDay:           toInt(d.EndDay),
		EndHour:          toInt(d.EndHour),
		EndMinute:        toInt(d.EndMinute),
		EndSecond:        toInt(d.EndSecond),
		EndMillisecond:   toInt(d.EndMillisecond),
		EndTZ:            d.GetEndTz(),
		Phrase:           d.GetPhrase(),
	}
}

// NameToProto carries a NameValue's form and ordered parts into its message.
func NameToProto(v namevalues.Value) *engine.NameValueInput {
	out := &engine.NameValueInput{Form: v.Form}
	for _, p := range v.Parts {
		out.Parts = append(out.Parts, &engine.NameValuePartInput{Value: p.Value, Type: p.Type})
	}
	return out
}

// NameFromProto is the inverse of NameToProto. Text is trimmed and parts are
// indexed in message order.
func NameFromProto(n *engine.NameValueInput) namevalues.Value {
	v := namevalues.Value{Form: strings.TrimSpace(n.GetForm())}
	for i, p := range n.GetParts() {
		v.Parts = append(v.Parts, namevalues.Part{
			Idx:   i,
			Value: strings.TrimSpace(p.GetValue()),
			Type:  strings.TrimSpace(p.GetType()),
		})
	}
	return v
}

// MarshalDate encodes a DateValue as its protobuf message bytes.
func MarshalDate(v datevalues.Value) ([]byte, error) { return proto.Marshal(DateToProto(v)) }

// UnmarshalDate decodes MarshalDate output.
func UnmarshalDate(b []byte) (datevalues.Value, error) {
	var d engine.DateValueInput
	if err := proto.Unmarshal(b, &d); err != nil {
		return datevalues.Value{}, err
	}
	return DateFromProto(&d), nil
}

// MarshalName encodes a NameValue as its protobuf message bytes.
func MarshalName(v namevalues.Value) ([]byte, error) { return proto.Marshal(NameToProto(v)) }

// UnmarshalName decodes MarshalName output.
func UnmarshalName(b []byte) (namevalues.Value, error) {
	var n engine.NameValueInput
	if err := proto.Unmarshal(b, &n); err != nil {
		return namevalues.Value{}, err
	}
	return NameFromProto(&n), nil
}

func toInt32(p *int) *int32 {
	if p == nil {
		return nil
	}
	n := int32(*p)
	return &n
}

func toInt(p *int32) *int {
	if p == nil {
		return nil
	}
	n := int(*p)
	return &n
}
