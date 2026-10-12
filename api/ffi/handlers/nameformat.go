package handlers

import (
	"fmt"

	"github.com/mendahu/provenencia/api/proto/engine"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/nameformats"
	"google.golang.org/protobuf/proto"
)

func GetNameFormat(in []byte) ([]byte, error) {
	var req engine.GetNameFormatRequest
	if err := proto.Unmarshal(in, &req); err != nil {
		return nil, fmt.Errorf("get_name_format: unmarshal: %w", err)
	}
	var resp engine.GetNameFormatResponse
	err := withProjectCatalog(req.GetProjectDir(), func(c *database.Catalog) error {
		db, err := c.DB()
		if err != nil {
			return err
		}
		loaded, err := nameformats.Load(db)
		if err != nil {
			return err
		}
		if loaded.Builtin {
			return nil
		}
		resp.Natural = nameFormatParts(loaded.Profile.Natural)
		resp.Sorted = nameFormatParts(loaded.Profile.Sorted)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return proto.Marshal(&resp)
}

func nameFormatParts(rows []nameformats.PartOrder) []*engine.NameFormatPart {
	out := make([]*engine.NameFormatPart, 0, len(rows))
	for _, row := range rows {
		out = append(out, &engine.NameFormatPart{
			PartType:          row.Type,
			TrailingSeparator: row.TrailingSeparator,
		})
	}
	return out
}
