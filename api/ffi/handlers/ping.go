package handlers

import (
	"fmt"

	"github.com/mendahu/provenencia/api/proto/engine"
	"google.golang.org/protobuf/proto"
)

func Ping(in []byte) ([]byte, error) {
	var req engine.PingRequest
	if err := proto.Unmarshal(in, &req); err != nil {
		return nil, fmt.Errorf("ping: unmarshal: %w", err)
	}
	return proto.Marshal(&engine.PingResponse{Message: req.GetMessage()})
}
