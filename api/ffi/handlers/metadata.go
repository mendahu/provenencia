package handlers

import (
	"bytes"

	"github.com/mendahu/provenencia/api/proto/engine"
	"github.com/mendahu/provenencia/core/apperr"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/rowchange"
	"github.com/mendahu/provenencia/core/database/sourcemetadata"
	"github.com/mendahu/provenencia/core/vocabkey"
	"github.com/mendahu/provenencia/core/writes"
	"google.golang.org/protobuf/proto"
)

// parseOwner reads a metadata owner off the wire. An absent message is
// sourcemetadata.ErrInvalid; the id must be a UUIDv7; the level is copied as
// given and refused by sourcemetadata when it holds no metadata.
func parseOwner(o *engine.MetadataOwner) (sourcemetadata.Owner, error) {
	if o == nil {
		return sourcemetadata.Owner{}, sourcemetadata.ErrInvalid
	}
	id, err := parseID(o.GetId())
	if err != nil {
		return sourcemetadata.Owner{}, err
	}
	return sourcemetadata.Owner{Level: vocabkey.Level(o.GetLevel()), ID: id}, nil
}

// parseDraft reads a metadata draft off the wire. An absent message is nil
// (no metadata change); a present one, empty included, is the draft the owner's
// stored values become. A malformed field id is request.invalid_id before any
// write.
func parseDraft(d *engine.MetadataDraft) (*sourcemetadata.Draft, error) {
	if d == nil {
		return nil, nil
	}
	out := &sourcemetadata.Draft{}
	for _, v := range d.GetValues() {
		id, err := parseID(v.GetFieldId())
		if err != nil {
			return nil, err
		}
		out.Values = append(out.Values, sourcemetadata.Input{FieldID: id, ValueText: v.GetValueText()})
	}
	for _, l := range d.GetLayout() {
		id, err := parseID(l.GetFieldId())
		if err != nil {
			return nil, err
		}
		out.Layout = append(out.Layout, sourcemetadata.LayoutRow{FieldID: id, SortOrder: int(l.GetSortOrder()), Dismissed: l.GetDismissed()})
	}
	return out, nil
}

// applyDraft saves a draft on the owner inside the caller's write: no changes
// for an absent draft, else sourcemetadata.Apply.
func applyDraft(tx *database.Tx, userID []byte, owner sourcemetadata.Owner, d *sourcemetadata.Draft) ([]rowchange.Change, error) {
	if d == nil {
		return nil, nil
	}
	return sourcemetadata.Apply(tx, userID, owner, *d)
}

// artifactBoardsProto reads every Artifact's board in one batch, in the order
// of ids (the raw Artifact ids, as stored).
func artifactBoardsProto(c *database.Catalog, ids [][]byte) ([]*engine.MetadataBoard, error) {
	owners := make([]sourcemetadata.Owner, len(ids))
	for i, id := range ids {
		owners[i] = sourcemetadata.Owner{Level: vocabkey.LevelArtifact, ID: id}
	}
	lists, err := sourcemetadata.ListWorkspaces(c, owners)
	if err != nil {
		return nil, err
	}
	boards := make([]*engine.MetadataBoard, len(ids))
	for i, id := range ids {
		boards[i] = &engine.MetadataBoard{OwnerId: database.UUIDString(id), Entries: metadataEntriesProto(lists[i])}
	}
	return boards, nil
}

func SetMetadata(in []byte) ([]byte, error) {
	var req engine.SetMetadataRequest
	if err := proto.Unmarshal(in, &req); err != nil {
		return nil, unmarshalErr("set_metadata", err)
	}
	userID, err := parseUserID(req.GetUserId())
	if err != nil {
		return nil, err
	}
	owner, err := parseOwner(req.GetOwner())
	if err != nil {
		return nil, err
	}
	fieldID, err := parseID(req.GetFieldId())
	if err != nil {
		return nil, err
	}
	var out *engine.SetMetadataResponse
	err = withProjectCatalog(req.GetProjectDir(), func(c *database.Catalog) error {
		_, wrote, err := writes.Run(c, writes.Op{Action: "update_metadata", UserID: userID},
			func(tx *database.Tx) (sourcemetadata.Row, []rowchange.Change, error) {
				return sourcemetadata.Set(tx, userID, owner, sourcemetadata.Input{
					FieldID:   fieldID,
					ValueText: req.GetValueText(),
				})
			})
		if err != nil {
			return err
		}
		entries, err := sourcemetadata.ListWorkspace(c, owner)
		if err != nil {
			return err
		}
		for _, e := range entries {
			if bytes.Equal(e.Field.ID, fieldID) {
				out = &engine.SetMetadataResponse{
					Entry:   metadataEntryProto(e),
					Effects: writeEffectsProto(wrote),
				}
				return nil
			}
		}
		// Set succeeded, so the row is always in the workspace list.
		return apperr.New(apperr.CodeInternalUnknown, apperr.KindInternal)
	})
	if err != nil {
		return nil, err
	}
	return proto.Marshal(out)
}

func ClearMetadata(in []byte) ([]byte, error) {
	var req engine.ClearMetadataRequest
	if err := proto.Unmarshal(in, &req); err != nil {
		return nil, unmarshalErr("clear_metadata", err)
	}
	userID, err := parseUserID(req.GetUserId())
	if err != nil {
		return nil, err
	}
	owner, err := parseOwner(req.GetOwner())
	if err != nil {
		return nil, err
	}
	fieldID, err := parseID(req.GetFieldId())
	if err != nil {
		return nil, err
	}
	var out *engine.ClearMetadataResponse
	err = withProjectCatalog(req.GetProjectDir(), func(c *database.Catalog) error {
		_, wrote, err := writes.Run(c, writes.Op{Action: "update_metadata", UserID: userID},
			func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
				changes, err := sourcemetadata.Clear(tx, userID, owner, fieldID)
				return struct{}{}, changes, err
			})
		if err != nil {
			return err
		}
		out = &engine.ClearMetadataResponse{Effects: writeEffectsProto(wrote)}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return proto.Marshal(out)
}

func DismissMetadataSuggestion(in []byte) ([]byte, error) {
	var req engine.DismissMetadataSuggestionRequest
	if err := proto.Unmarshal(in, &req); err != nil {
		return nil, unmarshalErr("dismiss_metadata_suggestion", err)
	}
	userID, err := parseUserID(req.GetUserId())
	if err != nil {
		return nil, err
	}
	owner, err := parseOwner(req.GetOwner())
	if err != nil {
		return nil, err
	}
	fieldID, err := parseID(req.GetFieldId())
	if err != nil {
		return nil, err
	}
	var out *engine.DismissMetadataSuggestionResponse
	err = withProjectCatalog(req.GetProjectDir(), func(c *database.Catalog) error {
		_, wrote, err := writes.Run(c, writes.Op{Action: "dismiss_metadata_suggestion", UserID: userID},
			func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
				changes, err := sourcemetadata.DismissSuggestion(tx, userID, owner, fieldID)
				return struct{}{}, changes, err
			})
		if err != nil {
			return err
		}
		entries, err := sourcemetadata.ListWorkspace(c, owner)
		if err != nil {
			return err
		}
		out = &engine.DismissMetadataSuggestionResponse{Effects: writeEffectsProto(wrote), Metadata: metadataEntriesProto(entries)}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return proto.Marshal(out)
}

func ReorderMetadata(in []byte) ([]byte, error) {
	var req engine.ReorderMetadataRequest
	if err := proto.Unmarshal(in, &req); err != nil {
		return nil, unmarshalErr("reorder_metadata", err)
	}
	userID, err := parseUserID(req.GetUserId())
	if err != nil {
		return nil, err
	}
	owner, err := parseOwner(req.GetOwner())
	if err != nil {
		return nil, err
	}
	fieldIDs, err := parseIDs(req.GetFieldIds())
	if err != nil {
		return nil, err
	}
	var out *engine.ReorderMetadataResponse
	err = withProjectCatalog(req.GetProjectDir(), func(c *database.Catalog) error {
		_, wrote, err := writes.Run(c, writes.Op{Action: "reorder_metadata", UserID: userID},
			func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
				changes, err := sourcemetadata.Reorder(tx, userID, owner, fieldIDs)
				return struct{}{}, changes, err
			})
		if err != nil {
			return err
		}
		entries, err := sourcemetadata.ListWorkspace(c, owner)
		if err != nil {
			return err
		}
		out = &engine.ReorderMetadataResponse{Effects: writeEffectsProto(wrote), Metadata: metadataEntriesProto(entries)}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return proto.Marshal(out)
}

// GetMetadataBoard reads one owner's board: its type's suggestions at its
// level plus extras, by its layout. A missing owner is sourcemetadata.invalid.
func GetMetadataBoard(in []byte) ([]byte, error) {
	var req engine.GetMetadataBoardRequest
	if err := proto.Unmarshal(in, &req); err != nil {
		return nil, unmarshalErr("get_metadata_board", err)
	}
	owner, err := parseOwner(req.GetOwner())
	if err != nil {
		return nil, err
	}
	var out *engine.GetMetadataBoardResponse
	err = withProjectCatalog(req.GetProjectDir(), func(c *database.Catalog) error {
		entries, err := sourcemetadata.ListWorkspace(c, owner)
		if err != nil {
			return err
		}
		out = &engine.GetMetadataBoardResponse{Board: &engine.MetadataBoard{
			OwnerId: database.UUIDString(owner.ID),
			Entries: metadataEntriesProto(entries),
		}}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return proto.Marshal(out)
}

// metadataEntriesProto is metadataEntryProto for a whole board, nil when empty.
func metadataEntriesProto(entries []sourcemetadata.WorkspaceEntry) []*engine.MetadataWorkspaceEntry {
	var out []*engine.MetadataWorkspaceEntry
	for _, e := range entries {
		out = append(out, metadataEntryProto(e))
	}
	return out
}

func metadataEntryProto(e sourcemetadata.WorkspaceEntry) *engine.MetadataWorkspaceEntry {
	out := &engine.MetadataWorkspaceEntry{
		Field: &engine.MetadataField{
			Id:          database.UUIDString(e.Field.ID),
			Key:         e.Field.Key,
			Origin:      e.Field.Origin,
			Label:       e.Field.Label,
			DataType:    e.Field.DataType,
			Description: e.Field.Description,
		},
		Suggested: e.Suggested,
		SortOrder: int32(e.SortOrder),
		OpenHref:  e.OpenHref(),
	}
	if e.Value != nil {
		out.HasValue = true
		out.ValueText = e.Value.ValueText
	}
	return out
}
