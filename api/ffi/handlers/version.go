package handlers

import (
	"fmt"

	"github.com/mendahu/provenance/api/proto/engine"
	"github.com/mendahu/provenance/core"
	"google.golang.org/protobuf/proto"
)

func GetVersion(in []byte) ([]byte, error) {
	var req engine.GetVersionRequest
	if err := proto.Unmarshal(in, &req); err != nil {
		return nil, fmt.Errorf("get_version: unmarshal: %w", err)
	}
	return proto.Marshal(&engine.GetVersionResponse{Version: core.Version})
}
