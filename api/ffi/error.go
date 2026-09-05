package ffi

import (
	"github.com/mendahu/provenencia/api/proto/engine"
	"github.com/mendahu/provenencia/core/apperr"
	"google.golang.org/protobuf/proto"
)

// EncodeError maps err to a protobuf Error payload for provenencia_call status 1.
func EncodeError(err error) ([]byte, error) {
	ae := apperr.From(err)
	msg := &engine.Error{
		Code:   ae.Code(),
		Kind:   protoKind(ae.Kind()),
		Params: ae.Params(),
	}
	return proto.Marshal(msg)
}

func protoKind(k apperr.Kind) engine.ErrorKind {
	switch k {
	case apperr.KindUser:
		return engine.ErrorKind_ERROR_KIND_USER
	case apperr.KindConflict:
		return engine.ErrorKind_ERROR_KIND_CONFLICT
	case apperr.KindNotFound:
		return engine.ErrorKind_ERROR_KIND_NOT_FOUND
	default:
		return engine.ErrorKind_ERROR_KIND_INTERNAL
	}
}
