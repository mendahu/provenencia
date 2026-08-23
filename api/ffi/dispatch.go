package ffi

import (
	"fmt"

	"github.com/mendahu/provenance/api/proto/engine"
	"github.com/mendahu/provenance/core"
	"github.com/mendahu/provenance/core/database"
	"google.golang.org/protobuf/proto"
)

// Method values match provenance.engine.v1.Method in engine.proto.
const (
	MethodUnspecified int32 = 0
	MethodPing        int32 = 1
	MethodGetVersion  int32 = 2
	// TEMPORARY (Spike 1 PR3): remove with database.Probe / METHOD_SQLITE_PROBE.
	MethodSqliteProbe int32 = 3
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
	case MethodSqliteProbe:
		// TEMPORARY (Spike 1 PR3): temp-file SQLite round trip. Remove when PR4 lands.
		var req engine.SqliteProbeRequest
		if len(in) > 0 {
			if err := proto.Unmarshal(in, &req); err != nil {
				return nil, fmt.Errorf("sqlite_probe: unmarshal: %w", err)
			}
		}
		resp := &engine.SqliteProbeResponse{Ok: true, Detail: "ok"}
		if err := database.Probe(); err != nil {
			resp.Ok = false
			resp.Detail = err.Error()
		}
		return proto.Marshal(resp)
	default:
		return nil, fmt.Errorf("unknown method %d", method)
	}
}
