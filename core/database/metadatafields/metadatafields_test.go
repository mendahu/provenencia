package metadatafields

import (
	"bytes"
	"database/sql"
	"encoding/binary"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/artifacts"
	"github.com/mendahu/provenencia/core/database/catalogmodel"
	"github.com/mendahu/provenencia/core/database/dbtest"
	"github.com/mendahu/provenencia/core/database/deleteimpact"
	"github.com/mendahu/provenencia/core/database/rowchange"
	"github.com/mendahu/provenencia/core/database/sources"
	"github.com/mendahu/provenencia/core/database/sourcetypes"
	"github.com/mendahu/provenencia/core/database/sqlbatch"
	"github.com/mendahu/provenencia/core/database/users"
	"github.com/mendahu/provenencia/core/ref"
	"github.com/mendahu/provenencia/core/vocabkey"
	"github.com/mendahu/provenencia/core/writes"
)

func TestUpsertLookupList(t *testing.T) {
	tests := []struct {
		name string
		run  func(t *testing.T, c *database.Catalog)
	}{
		{
			name: "upsert lookup refresh data type",
			run: func(t *testing.T, c *database.Catalog) {
				id, err := Upsert(c, Field{
					Key: "author", Origin: vocabkey.OriginProvenencia, Label: "Author", DataType: DataTypeText,
				})
				if err != nil {
					t.Fatal(err)
				}
				got, err := Lookup(c, "author", vocabkey.OriginProvenencia)
				if err != nil {
					t.Fatal(err)
				}
				if string(got.ID) != string(id) || got.DataType != DataTypeText {
					t.Fatalf("got %+v", got)
				}
				if _, err := Upsert(c, Field{
					Key: "author", Origin: vocabkey.OriginProvenencia, Label: "Author", DataType: DataTypeURL,
				}); err != nil {
					t.Fatal(err)
				}
				got, err = Lookup(c, "author", vocabkey.OriginProvenencia)
				if err != nil {
					t.Fatal(err)
				}
				if got.DataType != DataTypeURL {
					t.Fatalf("data_type %q", got.DataType)
				}
			},
		},
		{
			name: "rejects bad data type",
			run: func(t *testing.T, c *database.Catalog) {
				if _, err := Upsert(c, Field{
					Key: "x", Origin: vocabkey.OriginUser, Label: "X", DataType: "number",
				}); !errors.Is(err, ErrInvalid) {
					t.Fatalf("got %v", err)
				}
			},
		},
		{
			name: "upsert accepts url data type",
			run: func(t *testing.T, c *database.Catalog) {
				id, err := Upsert(c, Field{
					Key: "landing-page", Origin: vocabkey.OriginUser, Label: "Landing page", DataType: DataTypeURL,
				})
				if err != nil {
					t.Fatal(err)
				}
				got, err := Lookup(c, "landing-page", vocabkey.OriginUser)
				if err != nil {
					t.Fatal(err)
				}
				if string(got.ID) != string(id) || got.DataType != DataTypeURL {
					t.Fatalf("got %+v", got)
				}
			},
		},
		{
			name: "lookup missing",
			run: func(t *testing.T, c *database.Catalog) {
				_, err := Lookup(c, "nope", vocabkey.OriginProvenencia)
				if !errors.Is(err, sql.ErrNoRows) {
					t.Fatalf("got %v", err)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := database.Create(t.TempDir(), "t.provenencia")
			if err != nil {
				t.Fatal(err)
			}
			defer c.Close()
			tt.run(t, c)
		})
	}
}

func TestDelete(t *testing.T) {
	userID := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}

	mustUser := func(t *testing.T, c *database.Catalog) {
		t.Helper()
		r, err := ref.Mint(ref.PrefixUser)
		if err != nil {
			t.Fatal(err)
		}
		if err := users.Upsert(c, userID, "Jake", r); err != nil {
			t.Fatal(err)
		}
	}
	latestAction := func(t *testing.T, c *database.Catalog) string {
		t.Helper()
		db, err := c.DB()
		if err != nil {
			t.Fatal(err)
		}
		var actionType string
		if err := db.QueryRow(
			`SELECT action_type FROM audit_transactions ORDER BY revision DESC LIMIT 1`,
		).Scan(&actionType); err != nil {
			t.Fatal(err)
		}
		return actionType
	}

	tests := []struct {
		name string
		run  func(t *testing.T, c *database.Catalog)
	}{
		{
			name: "unused provenencia ok",
			run: func(t *testing.T, c *database.Catalog) {
				mustUser(t, c)
				id, err := Upsert(c, Field{
					Key: "author", Origin: vocabkey.OriginProvenencia, Label: "Author", DataType: DataTypeText,
				})
				if err != nil {
					t.Fatal(err)
				}
				if err := runDelete(c, userID, id); err != nil {
					t.Fatal(err)
				}
				if latestAction(t, c) != "delete_metadata_field" {
					t.Fatalf("action %q", latestAction(t, c))
				}
				_, err = Lookup(c, "author", vocabkey.OriginProvenencia)
				if !errors.Is(err, sql.ErrNoRows) {
					t.Fatalf("got %v", err)
				}
			},
		},
		{
			name: "in use refuses",
			run: func(t *testing.T, c *database.Catalog) {
				mustUser(t, c)
				typeID, err := sourcetypes.Upsert(c, sourcetypes.Type{
					Key: "book", Origin: vocabkey.OriginUser, Label: "Book",
				})
				if err != nil {
					t.Fatal(err)
				}
				src, _, err := writes.Run(c, writes.Op{Action: "create_source", UserID: userID}, func(tx *database.Tx) (sources.Source, []rowchange.Change, error) {
					return sources.Create(tx, userID, sources.CreateInput{SourceTypeID: typeID, Title: "T"})
				})
				if err != nil {
					t.Fatal(err)
				}
				field, err := runCreate(c, "Folio", DataTypeText, "")
				if err != nil {
					t.Fatal(err)
				}
				db, err := c.DB()
				if err != nil {
					t.Fatal(err)
				}
				metaID, err := uuid.NewV7()
				if err != nil {
					t.Fatal(err)
				}
				if _, err := db.Exec(
					`INSERT INTO source_metadata (id, source_id, field_id, value_text) VALUES (?, ?, ?, ?)`,
					metaID[:], src.ID, field.ID, "12",
				); err != nil {
					t.Fatal(err)
				}
				if err := runDelete(c, userID, field.ID); !errors.Is(err, ErrInUse) {
					t.Fatalf("got %v", err)
				}
				if _, err := GetByID(c, field.ID); err != nil {
					t.Fatalf("field gone: %v", err)
				}
			},
		},
		{
			name: "a field held only on a graph is in use",
			run: func(t *testing.T, c *database.Catalog) {
				mustUser(t, c)
				src, _, _ := fieldHolders(t, c, userID)
				graphID := insertGraph(t, c, src.ID, "Household")
				field, err := runCreate(c, "Line", DataTypeText, "")
				if err != nil {
					t.Fatal(err)
				}
				db := dbtest.MustDB(t, c)
				id, err := uuid.NewV7()
				if err != nil {
					t.Fatal(err)
				}
				if _, err := db.Exec(`INSERT INTO evidence_graph_metadata (id, graph_id, field_id, value_text) VALUES (?, ?, ?, '14')`,
					id[:], graphID, field.ID); err != nil {
					t.Fatal(err)
				}
				if err := runDelete(c, userID, field.ID); !errors.Is(err, ErrInUse) {
					t.Fatalf("got %v, want ErrInUse", err)
				}
				if _, err := GetByID(c, field.ID); err != nil {
					t.Fatalf("field gone: %v", err)
				}
			},
		},
		{
			name: "a field only dismissed on an Artifact is released audited",
			run: func(t *testing.T, c *database.Catalog) {
				mustUser(t, c)
				_, art, typeID := fieldHolders(t, c, userID)
				field, err := runCreate(c, "Provider", DataTypeText, "")
				if err != nil {
					t.Fatal(err)
				}
				db := dbtest.MustDB(t, c)
				if _, err := db.Exec(`INSERT INTO source_type_metadata_fields (source_type_id, field_id, level, sort_order) VALUES (?, ?, 'artifact', 0)`,
					typeID, field.ID); err != nil {
					t.Fatal(err)
				}
				if _, err := db.Exec(`INSERT INTO artifact_metadata_layout (artifact_id, field_id, sort_order, dismissed) VALUES (?, ?, 0, 1)`,
					art, field.ID); err != nil {
					t.Fatal(err)
				}
				if n, err := deleteimpact.UsedBy(c, catalogmodel.KindMetadataField, field.ID); err != nil || n != 0 {
					t.Fatalf("UsedBy = %d, %v; want 0", n, err)
				}
				var changes []rowchange.Change
				if _, _, err := writes.Run(c, writes.Op{Action: "delete_metadata_field", UserID: userID},
					func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
						var err error
						changes, err = Delete(tx, userID, field.ID)
						return struct{}{}, changes, err
					}); err != nil {
					t.Fatal(err)
				}
				var layout []rowchange.Change
				for _, ch := range changes {
					if ch.EntityType == "artifact_metadata_layout" {
						layout = append(layout, ch)
					}
				}
				if len(layout) != 1 || layout[0].Action != rowchange.ActionDelete ||
					layout[0].Fields["artifact_id"].Old != database.UUIDString(art) ||
					layout[0].Fields["field_id"].Old != database.UUIDString(field.ID) {
					t.Fatalf("layout changes = %+v", layout)
				}
			},
		},
		{
			name: "UsedBy and List count Sources, Artifacts and graphs alike",
			run: func(t *testing.T, c *database.Catalog) {
				mustUser(t, c)
				s1, art, _ := fieldHolders(t, c, userID)
				s2, _, _ := fieldHolders(t, c, userID)
				graphID := insertGraph(t, c, s1.ID, "Household")
				field, err := runCreate(c, "Shelf mark", DataTypeText, "")
				if err != nil {
					t.Fatal(err)
				}
				db := dbtest.MustDB(t, c)
				for _, ins := range []struct {
					query string
					owner []byte
				}{
					{`INSERT INTO source_metadata (id, source_id, field_id, value_text) VALUES (?, ?, ?, 'a')`, s1.ID},
					{`INSERT INTO source_metadata (id, source_id, field_id, value_text) VALUES (?, ?, ?, 'b')`, s2.ID},
					{`INSERT INTO artifact_metadata (id, artifact_id, field_id, value_text) VALUES (?, ?, ?, 'c')`, art},
					{`INSERT INTO evidence_graph_metadata (id, graph_id, field_id, value_text) VALUES (?, ?, ?, 'd')`, graphID},
				} {
					id, err := uuid.NewV7()
					if err != nil {
						t.Fatal(err)
					}
					if _, err := db.Exec(ins.query, id[:], ins.owner, field.ID); err != nil {
						t.Fatal(err)
					}
				}
				if n, err := deleteimpact.UsedBy(c, catalogmodel.KindMetadataField, field.ID); err != nil || n != 4 {
					t.Fatalf("UsedBy = %d, %v; want 4", n, err)
				}
				all, err := List(c)
				if err != nil {
					t.Fatal(err)
				}
				for _, f := range all {
					if bytes.Equal(f.ID, field.ID) && f.UsedBy != 4 {
						t.Fatalf("List UsedBy = %d, want 4", f.UsedBy)
					}
				}
			},
		},
		{
			name: "unused plugin is origin locked",
			run: func(t *testing.T, c *database.Catalog) {
				mustUser(t, c)
				id, err := Upsert(c, Field{
					Key: "memorial_id", Origin: "plugin:findagrave", Label: "Memorial id", DataType: DataTypeText,
				})
				if err != nil {
					t.Fatal(err)
				}
				if err := runDelete(c, userID, id); !errors.Is(err, ErrOriginLocked) {
					t.Fatalf("got %v", err)
				}
				if _, err := GetByID(c, id); err != nil {
					t.Fatalf("field gone: %v", err)
				}
			},
		},
		{
			name: "missing is invalid",
			run: func(t *testing.T, c *database.Catalog) {
				mustUser(t, c)
				missing := make([]byte, 16)
				missing[15] = 9
				if err := runDelete(c, userID, missing); !errors.Is(err, ErrInvalid) {
					t.Fatalf("got %v", err)
				}
			},
		},
		{
			name: "unused field with suggestion and layout still erases",
			run: func(t *testing.T, c *database.Catalog) {
				mustUser(t, c)
				typeID, err := sourcetypes.Upsert(c, sourcetypes.Type{
					Key: "book", Origin: vocabkey.OriginUser, Label: "Book",
				})
				if err != nil {
					t.Fatal(err)
				}
				src, _, err := writes.Run(c, writes.Op{Action: "create_source", UserID: userID}, func(tx *database.Tx) (sources.Source, []rowchange.Change, error) {
					return sources.Create(tx, userID, sources.CreateInput{SourceTypeID: typeID, Title: "T"})
				})
				if err != nil {
					t.Fatal(err)
				}
				id, err := Upsert(c, Field{
					Key: "author", Origin: vocabkey.OriginUser, Label: "Author", DataType: DataTypeText,
				})
				if err != nil {
					t.Fatal(err)
				}
				db, err := c.DB()
				if err != nil {
					t.Fatal(err)
				}
				if _, err := db.Exec(
					`INSERT INTO source_type_metadata_fields (source_type_id, field_id, level, sort_order) VALUES (?, ?, 'source', 0)`,
					typeID, id,
				); err != nil {
					t.Fatal(err)
				}
				if _, err := db.Exec(
					`INSERT INTO source_metadata_layout (source_id, field_id, sort_order) VALUES (?, ?, 0)`,
					src.ID, id,
				); err != nil {
					t.Fatal(err)
				}
				if err := runDelete(c, userID, id); err != nil {
					t.Fatal(err)
				}
				if _, err := GetByID(c, id); !errors.Is(err, sql.ErrNoRows) {
					t.Fatalf("field still there: %v", err)
				}
				var joins int
				if err := db.QueryRow(
					`SELECT COUNT(*) FROM source_type_metadata_fields WHERE field_id = ?`, id,
				).Scan(&joins); err != nil || joins != 0 {
					t.Fatalf("suggestion leftover %d %v", joins, err)
				}
				if err := db.QueryRow(
					`SELECT COUNT(*) FROM source_metadata_layout WHERE field_id = ?`, id,
				).Scan(&joins); err != nil || joins != 0 {
					t.Fatalf("layout leftover %d %v", joins, err)
				}
				if _, err := sourcetypes.GetByID(c, typeID); err != nil {
					t.Fatalf("type should survive: %v", err)
				}
			},
		},
		{
			name: "bad id",
			run: func(t *testing.T, c *database.Catalog) {
				if err := runDelete(c, userID, []byte{1}); !errors.Is(err, ErrInvalid) {
					t.Fatalf("got %v", err)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := database.Create(t.TempDir(), "t.provenencia")
			if err != nil {
				t.Fatal(err)
			}
			defer c.Close()
			tt.run(t, c)
		})
	}
}

func TestCreateUpdateGetByID(t *testing.T) {
	tests := []struct {
		name string
		run  func(t *testing.T, c *database.Catalog)
	}{
		{
			name: "create mints slug key",
			run: func(t *testing.T, c *database.Catalog) {
				f, err := runCreate(c, "Grandma's album code", DataTypeText, "")
				if err != nil {
					t.Fatal(err)
				}
				if f.Key != "grandmas-album-code" || f.Origin != vocabkey.OriginUser {
					t.Fatalf("got %+v", f)
				}
			},
		},
		{
			name: "create rejects unslugifiable label",
			run: func(t *testing.T, c *database.Catalog) {
				if _, err := runCreate(c, "...", DataTypeText, ""); !errors.Is(err, ErrInvalid) {
					t.Fatalf("got %v", err)
				}
			},
		},
		{
			name: "create rejects duplicate key under user origin",
			run: func(t *testing.T, c *database.Catalog) {
				if _, err := runCreate(c, "Album code", DataTypeText, ""); err != nil {
					t.Fatal(err)
				}
				_, err := runCreate(c, "Album code", DataTypeText, "")
				if !errors.Is(err, ErrDuplicateKey) {
					t.Fatalf("got %v", err)
				}
			},
		},
		{
			name: "same key different origin does not collide",
			run: func(t *testing.T, c *database.Catalog) {
				if _, err := Upsert(c, Field{
					Key: "photographer", Origin: vocabkey.OriginProvenencia, Label: "Photographer", DataType: DataTypeText,
				}); err != nil {
					t.Fatal(err)
				}
				f, err := runCreate(c, "Photographer", DataTypeText, "")
				if err != nil {
					t.Fatal(err)
				}
				if f.Key != "photographer" || f.Origin != vocabkey.OriginUser {
					t.Fatalf("got %+v", f)
				}
			},
		},
		{
			name: "update patches label and description but not key or data type",
			run: func(t *testing.T, c *database.Catalog) {
				created, err := runCreate(c, "Album code", DataTypeText, "old")
				if err != nil {
					t.Fatal(err)
				}
				updated, err := runUpdate(c, created.ID, "Album Code", DataTypeText, "new")
				if err != nil {
					t.Fatal(err)
				}
				if updated.Key != created.Key || updated.Label != "Album Code" ||
					updated.DataType != DataTypeText || updated.Description != "new" {
					t.Fatalf("got %+v", updated)
				}
				got, err := GetByID(c, created.ID)
				if err != nil {
					t.Fatal(err)
				}
				if got.Key != updated.Key || got.Label != updated.Label {
					t.Fatalf("got %+v want %+v", got, updated)
				}
			},
		},
		{
			name: "update rejects data type change",
			run: func(t *testing.T, c *database.Catalog) {
				created, err := runCreate(c, "Album code", DataTypeText, "")
				if err != nil {
					t.Fatal(err)
				}
				if _, err := runUpdate(c, created.ID, "Album code", DataTypeURL, ""); !errors.Is(err, ErrInvalid) {
					t.Fatalf("got %v", err)
				}
			},
		},
		{
			name: "update provenencia field",
			run: func(t *testing.T, c *database.Catalog) {
				id, err := Upsert(c, Field{
					Key: "author", Origin: vocabkey.OriginProvenencia, Label: "Author", DataType: DataTypeText,
				})
				if err != nil {
					t.Fatal(err)
				}
				updated, err := runUpdate(c, id, "Author renamed", DataTypeText, "edited")
				if err != nil {
					t.Fatal(err)
				}
				if updated.Key != "author" || updated.Origin != vocabkey.OriginProvenencia ||
					updated.Label != "Author renamed" || updated.DataType != DataTypeText ||
					updated.Description != "edited" {
					t.Fatalf("got %+v", updated)
				}
			},
		},
		{
			name: "update rejects plugin field",
			run: func(t *testing.T, c *database.Catalog) {
				id, err := Upsert(c, Field{
					Key: "memorial_id", Origin: "plugin:findagrave", Label: "Memorial id", DataType: DataTypeText,
				})
				if err != nil {
					t.Fatal(err)
				}
				if _, err := runUpdate(c, id, "Memorial", DataTypeText, ""); !errors.Is(err, ErrLocked) {
					t.Fatalf("got %v", err)
				}
			},
		},
		{
			name: "get by id missing",
			run: func(t *testing.T, c *database.Catalog) {
				if _, err := GetByID(c, make([]byte, 16)); !errors.Is(err, sql.ErrNoRows) {
					t.Fatalf("got %v", err)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := database.Create(t.TempDir(), "t.provenencia")
			if err != nil {
				t.Fatal(err)
			}
			defer c.Close()
			tt.run(t, c)
		})
	}
}

func TestListByIDsTx(t *testing.T) {
	c, err := database.Create(t.TempDir(), "t.provenencia")
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	a, err := runCreate(c, "Author", DataTypeText, "")
	if err != nil {
		t.Fatal(err)
	}
	b, err := runCreate(c, "Link", DataTypeURL, "")
	if err != nil {
		t.Fatal(err)
	}
	db, err := c.DB()
	if err != nil {
		t.Fatal(err)
	}
	tx, err := db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback() }()

	got, err := ListByIDsTx(tx, [][]byte{a.ID, b.ID, a.ID, make([]byte, 16)})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("got %d fields, want 2: %+v", len(got), got)
	}
	byID := map[string]Field{}
	for _, f := range got {
		byID[string(f.ID)] = f
	}
	if byID[string(a.ID)].Key != "author" || byID[string(b.ID)].DataType != DataTypeURL {
		t.Fatalf("got %+v", got)
	}
	none, err := ListByIDsTx(tx, nil)
	if err != nil || none != nil {
		t.Fatalf("nil ids = %+v, %v", none, err)
	}

	t.Run("reads past one batch", func(t *testing.T) {
		ids := [][]byte{a.ID}
		for i := 0; i < sqlbatch.Size*2+1; i++ {
			unknown := make([]byte, 16)
			binary.BigEndian.PutUint64(unknown[8:], uint64(i+1))
			ids = append(ids, unknown)
		}
		ids = append(ids, b.ID)
		got, err := ListByIDsTx(tx, ids)
		if err != nil {
			t.Fatal(err)
		}
		if len(got) != 2 {
			t.Fatalf("got %d fields, want the 2 known: %+v", len(got), got)
		}
	})
}

func runCreate(c *database.Catalog, label, dataType, description string) (Field, error) {
	got, _, err := writes.Run(c, writes.Op{Action: "create_metadata_field"},
		func(tx *database.Tx) (Field, []rowchange.Change, error) {
			return Create(tx, label, dataType, description)
		})
	return got, err
}

func runUpdate(c *database.Catalog, id []byte, label, dataType, description string) (Field, error) {
	got, _, err := writes.Run(c, writes.Op{Action: "update_metadata_field"},
		func(tx *database.Tx) (Field, []rowchange.Change, error) {
			return Update(tx, id, label, dataType, description)
		})
	return got, err
}

func runDelete(c *database.Catalog, userID, id []byte) error {
	_, _, err := writes.Run(c, writes.Op{Action: "delete_metadata_field", UserID: userID},
		func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
			changes, err := Delete(tx, userID, id)
			return struct{}{}, changes, err
		})
	return err
}

// fieldHolders makes a Source of a new type with an Artifact and returns the
// Source, the Artifact id and the type id.
func fieldHolders(t *testing.T, c *database.Catalog, userID []byte) (src sources.Source, artifactID, typeID []byte) {
	t.Helper()
	typeID, err := sourcetypes.Upsert(c, sourcetypes.Type{Key: "ledger", Origin: vocabkey.OriginUser, Label: "Ledger"})
	if err != nil {
		t.Fatal(err)
	}
	src, _, err = writes.Run(c, writes.Op{Action: "create_source", UserID: userID}, func(tx *database.Tx) (sources.Source, []rowchange.Change, error) {
		return sources.Create(tx, userID, sources.CreateInput{SourceTypeID: typeID, Title: "T"})
	})
	if err != nil {
		t.Fatal(err)
	}
	art, _, err := writes.Run(c, writes.Op{Action: "create_artifact", UserID: userID}, func(tx *database.Tx) (artifacts.Artifact, []rowchange.Change, error) {
		return artifacts.Create(tx, userID, artifacts.CreateInput{SourceID: src.ID, Label: "Page"})
	})
	if err != nil {
		t.Fatal(err)
	}
	return src, art.ID, typeID
}

// insertGraph adds a graph to the Source by SQL and returns its id.
func insertGraph(t *testing.T, c *database.Catalog, sourceID []byte, name string) []byte {
	t.Helper()
	id, err := uuid.NewV7()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := dbtest.MustDB(t, c).Exec(`INSERT INTO evidence_graphs (id, source_id, name) VALUES (?, ?, ?)`, id[:], sourceID, name); err != nil {
		t.Fatal(err)
	}
	return id[:]
}
