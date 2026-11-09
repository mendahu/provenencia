package deleteimpact_test

import (
	"testing"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/datevalues"
	"github.com/mendahu/provenencia/core/database/deleteimpact"
	"github.com/mendahu/provenencia/core/database/rowchange"
	"github.com/mendahu/provenencia/core/database/sourcetypes"
	"github.com/mendahu/provenencia/core/datevalue"
	"github.com/mendahu/provenencia/core/vocabkey"
	"github.com/mendahu/provenencia/core/writes"
)

// ReleaseReplaced returns the audited delete of a row child it released, and
// nothing when there is nothing to release.
func TestReleaseReplacedAudits(t *testing.T) {
	c, userID := testCatalog(t)
	typeID, err := sourcetypes.Upsert(c, sourcetypes.Type{
		Key: "book", Origin: vocabkey.OriginProvenencia, Label: "Book",
	})
	if err != nil {
		t.Fatal(err)
	}
	db, err := c.DB()
	if err != nil {
		t.Fatal(err)
	}
	sqlTx, err := db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	year := 1880
	dateID, err := datevalues.InsertTx(sqlTx, datevalue.Value{Kind: datevalue.KindPoint, StartYear: &year})
	if err != nil {
		t.Fatal(err)
	}
	srcID := []byte{9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9}
	if _, err := sqlTx.Exec(
		`INSERT INTO sources (id, ref, source_type_id, title, date_value_id) VALUES (?, 'SRC-AAAAA', ?, 'Deed', ?)`,
		srcID, typeID, dateID,
	); err != nil {
		t.Fatal(err)
	}
	if err := sqlTx.Commit(); err != nil {
		t.Fatal(err)
	}

	_, err = writes.Call(c, writes.Op{Action: "t", UserID: userID},
		func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
			rowMap := func() (map[string]any, error) {
				v, err := datevalues.LookupTx(tx.Tx, dateID)
				return datevalues.RowMap(dateID, v), err
			}
			t.Run("malformed prev id is a no op", func(t *testing.T) {
				called := false
				ch, err := deleteimpact.ReleaseReplaced(tx, "sources.date_value_id", "date_value", []byte{1, 2, 3}, nil,
					func() (map[string]any, error) { called = true; return nil, nil })
				if ch != nil || err != nil || called {
					t.Fatalf("ch=%+v err=%v rowMap called=%v", ch, err, called)
				}
			})
			if ch, err := deleteimpact.ReleaseReplaced(tx, "sources.date_value_id", "date_value", nil, nil, rowMap); err != nil || ch != nil {
				t.Fatalf("empty prev: ch=%+v err=%v", ch, err)
			}
			if ch, err := deleteimpact.ReleaseReplaced(tx, "sources.date_value_id", "date_value", dateID, dateID, rowMap); err != nil || ch != nil {
				t.Fatalf("prev equals new: ch=%+v err=%v", ch, err)
			}
			var n int
			if err := tx.QueryRow(`SELECT COUNT(*) FROM date_values WHERE id = ?`, dateID).Scan(&n); err != nil || n != 1 {
				t.Fatalf("row released early: n=%d err=%v", n, err)
			}
			if _, err := tx.Exec(`UPDATE sources SET date_value_id = NULL WHERE id = ?`, srcID); err != nil {
				t.Fatal(err)
			}
			ch, err := deleteimpact.ReleaseReplaced(tx, "sources.date_value_id", "date_value", dateID, nil, rowMap)
			if err != nil {
				t.Fatal(err)
			}
			if ch == nil || ch.EntityType != "date_value" || ch.Action != rowchange.ActionDelete ||
				ch.Fields["start_year"].Old != 1880 || ch.Fields["kind"].Old != datevalue.KindPoint || ch.Fields["kind"].New != nil {
				t.Fatalf("delete change %+v", ch)
			}
			return struct{}{}, []rowchange.Change{*ch}, nil
		})
	if err != nil {
		t.Fatal(err)
	}
	var n int
	if err := db.QueryRow(`SELECT COUNT(*) FROM date_values WHERE id = ?`, dateID).Scan(&n); err != nil || n != 0 {
		t.Fatalf("row left: n=%d err=%v", n, err)
	}
}
