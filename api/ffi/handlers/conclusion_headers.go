package handlers

import (
	"github.com/mendahu/provenencia/api/proto/engine"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/conclusionheaders"
	"github.com/mendahu/provenencia/core/valuecodec"
	"google.golang.org/protobuf/proto"
)

func ListPersonHeaders(in []byte) ([]byte, error) {
	var req engine.ListPersonHeadersRequest
	if err := proto.Unmarshal(in, &req); err != nil {
		return nil, unmarshalErr("list_person_headers", err)
	}
	out := &engine.ListPersonHeadersResponse{}
	err := withProjectCatalog(req.GetProjectDir(), func(c *database.Catalog) error {
		db, err := c.DB()
		if err != nil {
			return err
		}
		headers, err := conclusionheaders.ListPersons(db)
		if err != nil {
			return err
		}
		for _, h := range headers {
			ph := &engine.PersonHeader{
				Entity:           canonicalEntityProto(h.Entity),
				NameClusterCount: int32(h.NameClusterCount),
			}
			if h.Name != nil {
				ph.Name = valuecodec.NameToProto(*h.Name)
			}
			out.Headers = append(out.Headers, ph)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return proto.Marshal(out)
}
