package ffi

import (
	"fmt"

	"github.com/mendahu/provenance/api/proto/engine"
	"github.com/mendahu/provenance/core"
	"google.golang.org/protobuf/proto"
)

// Method values match provenance.engine.v1.Method in engine.proto.
const (
	MethodPing       int32 = 1
	MethodGetVersion int32 = 2
)

// Call dispatches one coarse FFI operation. in/out are protobuf bytes.
func Call(method int32, in []byte) ([]byte, error) {
	switch method {
	case MethodPing:
		var req engine.PingRequest
		if err := proto.Unmarshal(in, &req); err != nil {
			return nil, fmt.Errorf("ping: unmarshal: %w", err)
		}
		return proto.Marshal(&engine.PingResponse{Message: req.GetMessage()})
	case MethodGetVersion:
		var req engine.GetVersionRequest
		if len(in) > 0 {
			if err := proto.Unmarshal(in, &req); err != nil {
				return nil, fmt.Errorf("get_version: unmarshal: %w", err)
			}
		}
		return proto.Marshal(&engine.GetVersionResponse{Version: core.Version})
	default:
		return nil, fmt.Errorf("unknown method %d", method)
	}
}
