package deleteimpact_test

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/mendahu/provenencia/core/database/catalogmodel"

	"github.com/google/uuid"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/artifacts"
	"github.com/mendahu/provenencia/core/database/autoreconciler"
	"github.com/mendahu/provenencia/core/database/citations"
	"github.com/mendahu/provenencia/core/database/deleteimpact"
	"github.com/mendahu/provenencia/core/database/identityclaims"
	"github.com/mendahu/provenencia/core/database/observations"
	"github.com/mendahu/provenencia/core/database/promote"
	"github.com/mendahu/provenencia/core/database/properties"
	"github.com/mendahu/provenencia/core/database/rowchange"
	"github.com/mendahu/provenencia/core/database/sources"
	"github.com/mendahu/provenencia/core/database/sourcetypes"
	"github.com/mendahu/provenencia/core/database/subjects"
	"github.com/mendahu/provenencia/core/database/subjecttypes"
	"github.com/mendahu/provenencia/core/database/subjectvocab"
	"github.com/mendahu/provenencia/core/writes"
)

// pinnedHandle is one Place handle filed through Promote: A grounds it, B
// joins with two confirmed pairs (each Subject's toponym and alternate name),
// so all four Observations are pinned on both claims.
type pinnedHandle struct {
	c                   *database.Catalog
	userID              []byte
	a, b                subjects.Subject
	aName, aAlt         []byte // A's Observations
	bName, bAlt         []byte // B's Observations
	entityID, entityRef string
	ca, cb              []byte
}

func newPinnedHandle(t *testing.T) pinnedHandle {
	t.Helper()
	c, userID := testCatalog(t)
	must := func(err error) {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
	}
	must(subjectvocab.Install(c))
	typeID, err := sourcetypes.Upsert(c, sourcetypes.Type{Key: "book", Origin: sourcetypes.OriginProvenencia, Label: "Book"})
	must(err)
	src, _, err := writes.Run(c, writes.Op{Action: "create_source", UserID: userID}, func(tx *database.Tx) (sources.Source, []rowchange.Change, error) {
		return sources.Create(tx, userID, sources.CreateInput{SourceTypeID: typeID, Title: "Gazetteer"})
	})
	must(err)
	art, err := runArtifactCreate(c, userID, artifacts.CreateInput{SourceID: src.ID, Label: "Scan"})
	must(err)
	place, err := subjecttypes.Lookup(c, "place", subjecttypes.OriginProvenencia)
	must(err)
	toponym, err := properties.Lookup(c, "toponym", properties.OriginProvenencia)
	must(err)
	mk := func(label string) subjects.Subject {
		t.Helper()
		s, err := writes.Call(c, writes.Op{Action: "create_subject", UserID: userID}, func(tx *database.Tx) (subjects.Subject, []rowchange.Change, error) {
			return subjects.Create(tx, userID, subjects.CreateInput{SourceID: src.ID, SubjectTypeID: place.ID, Label: label}, nil)
		})
		must(err)
		return s
	}
	a, b := mk("York"), mk("York (U.C.)")
	res, err := writes.Call(c, writes.Op{Action: "create_citation_with_observations", UserID: userID}, func(tx *database.Tx) (citations.CreateResult, []rowchange.Change, error) {
		return citations.CreateWithObservations(tx, userID, citations.CreateInput{
			ArtifactID: art.ID, LocatorJSON: testLocator, Transcription: "York, otherwise Toronto",
		}, []observations.Input{
			{SubjectID: a.ID, PropertyID: toponym.ID, ValueText: "York", HasText: true},
			{SubjectID: a.ID, PropertyID: toponym.ID, ValueText: "Toronto", HasText: true},
			{SubjectID: b.ID, PropertyID: toponym.ID, ValueText: "York", HasText: true},
			{SubjectID: b.ID, PropertyID: toponym.ID, ValueText: "Toronto", HasText: true},
		})
	})
	must(err)
	o := res.Observations
	h := pinnedHandle{c: c, userID: userID, a: a, b: b,
		aName: o[0].ID, aAlt: o[1].ID, bName: o[2].ID, bAlt: o[3].ID}

	grounding, err := writes.Call(c, writes.Op{Action: "promote_subject", UserID: userID}, func(tx *database.Tx) (promote.Result, []rowchange.Change, error) {
		return promote.Save(tx, userID, promote.Input{SubjectID: a.ID, Argument: "The gazetteer entry."})
	})
	must(err)
	joined, err := writes.Call(c, writes.Op{Action: "promote_subject", UserID: userID}, func(tx *database.Tx) (promote.Result, []rowchange.Change, error) {
		return promote.Save(tx, userID, promote.Input{SubjectID: b.ID, EntityID: grounding.Entity.ID, Pairs: []promote.Pair{
			{IncomingObservationID: h.bName, MemberObservationID: h.aName},
			{IncomingObservationID: h.bAlt, MemberObservationID: h.aAlt},
		}})
	})
	must(err)
	h.entityID, h.entityRef = string(grounding.Entity.ID), grounding.Entity.Ref
	h.ca, h.cb = grounding.Claim.ID, joined.Claim.ID
	return h
}

// pinned is the claim's pinned Observations, as "obs" names in a fixed order.
func (h pinnedHandle) pinned(t *testing.T, claimID []byte) string {
	t.Helper()
	db, err := h.c.DB()
	if err != nil {
		t.Fatal(err)
	}
	ids, err := identityclaims.PinnedObservations(db, claimID)
	if err != nil {
		t.Fatal(err)
	}
	have := map[string]bool{}
	for _, id := range ids {
		have[string(id)] = true
	}
	var out []string
	for _, n := range []struct {
		name string
		id   []byte
	}{{"aName", h.aName}, {"aAlt", h.aAlt}, {"bName", h.bName}, {"bAlt", h.bAlt}} {
		if have[string(n.id)] {
			out = append(out, n.name)
		}
	}
	return strings.Join(out, ",")
}

// exhibitHistory replays (identity_claim_evidence, claim id) from the audit in
// revision order: "action_type:create|delete:obs name" per change. This is the
// per-claim history the review alert (model §5.2) reads.
func (h pinnedHandle) exhibitHistory(t *testing.T, claimID []byte) []string {
	t.Helper()
	db, err := h.c.DB()
	if err != nil {
		t.Fatal(err)
	}
	rows, err := db.Query(`SELECT tx.action_type, ch.action, ch.changes_json
		FROM audit_changes ch JOIN audit_transactions tx ON tx.id = ch.audit_transaction_id
		WHERE ch.entity_type = 'identity_claim_evidence' AND ch.entity_id = ?
		ORDER BY tx.revision, ch.rowid`, claimID)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	names := map[string]string{}
	for name, id := range map[string][]byte{"aName": h.aName, "aAlt": h.aAlt, "bName": h.bName, "bAlt": h.bAlt} {
		names[uuid.UUID(id).String()] = name
	}
	claim := uuid.UUID(claimID).String()
	var out []string
	for rows.Next() {
		var actionType, action, js string
		if err := rows.Scan(&actionType, &action, &js); err != nil {
			t.Fatal(err)
		}
		var diff map[string]struct{ Old, New any }
		if err := json.Unmarshal([]byte(js), &diff); err != nil {
			t.Fatal(err)
		}
		side := func(f string) any {
			if action == "delete" {
				return diff[f].Old
			}
			return diff[f].New
		}
		if side("identity_claim_id") != claim {
			t.Fatalf("change filed under the wrong claim: %s", js)
		}
		out = append(out, fmt.Sprintf("%s:%s:%s", actionType, action, names[fmt.Sprint(side("observation_id"))]))
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	return out
}

// cacheMatchesRebuild: the auto-reconciler cache upkeep left equals a rebuild
// from truth tables (rolled back, so it changes nothing).
func (h pinnedHandle) cacheMatchesRebuild(t *testing.T) {
	t.Helper()
	db, err := h.c.DB()
	if err != nil {
		t.Fatal(err)
	}
	snapshot := func(tx *sql.Tx) string {
		t.Helper()
		var b strings.Builder
		for _, table := range []string{"auto_reconciler_values", "auto_reconciler_outcomes"} {
			rows, err := tx.Query(`SELECT * FROM ` + table + ` ORDER BY 1, 2, 3`)
			if err != nil {
				t.Fatal(err)
			}
			cols, _ := rows.Columns()
			for rows.Next() {
				vals := make([]any, len(cols))
				ptrs := make([]any, len(cols))
				for i := range vals {
					ptrs[i] = &vals[i]
				}
				if err := rows.Scan(ptrs...); err != nil {
					t.Fatal(err)
				}
				fmt.Fprintf(&b, "%s %v\n", table, vals)
			}
			_ = rows.Close()
		}
		return b.String()
	}
	tx, err := db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback() }()
	before := snapshot(tx)
	if err := autoreconciler.Rebuild(tx); err != nil {
		t.Fatal(err)
	}
	if after := snapshot(tx); after != before {
		t.Fatalf("cache upkeep differs from a rebuild\nupkeep:\n%s\nrebuild:\n%s", before, after)
	}
}

func TestPinnedObservationDeleteEndToEnd(t *testing.T) {
	h := newPinnedHandle(t)
	if got := h.pinned(t, h.ca); got != "aName,aAlt,bName,bAlt" {
		t.Fatalf("setup: A's claim pins %s", got)
	}
	if got := h.pinned(t, h.cb); got != "aName,aAlt,bName,bAlt" {
		t.Fatalf("setup: B's claim pins %s", got)
	}

	// The confirm: allowed, and it names the handle once though two claims pin it.
	report := mustImpact(t, h.c, catalogmodel.KindObservation, h.aName)
	if !report.Allowed || report.Gate != deleteimpact.GateOK || len(report.Groups) != 0 || len(report.Cascades) != 1 {
		t.Fatalf("%+v", report)
	}
	g := report.Cascades[0]
	if g.Via != "identity_claim_evidence.observation_id" || g.Kind != catalogmodel.KindCanonicalEntity ||
		g.Total != 1 || len(g.Listed) != 1 || g.Listed[0].Ref != h.entityRef {
		t.Fatalf("%+v", g)
	}

	// The delete: both claims lose that Observation and keep every other pin.
	if _, err := writes.Call(h.c, writes.Op{Action: "delete_observation", UserID: h.userID}, func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
		changes, err := observations.Delete(tx, h.userID, h.aName)
		return struct{}{}, changes, err
	}); err != nil {
		t.Fatal(err)
	}
	action, types := lastRevision(t, h.c)
	if action != "delete_observation" || strings.Join(types, ",") != "identity_claim_evidence,identity_claim_evidence,observation" {
		t.Fatalf("%s %v", action, types)
	}
	for _, claimID := range [][]byte{h.ca, h.cb} {
		if got := h.pinned(t, claimID); got != "aAlt,bName,bAlt" {
			t.Fatalf("pins after delete: %s", got)
		}
	}
	member, err := identityclaims.Get(h.c, h.ca)
	if err != nil || member.Argument != "The gazetteer entry." {
		t.Fatalf("claim should stay as written: %v %+v", err, member)
	}

	// The audit reads back per claim: what Promote pinned, then what the delete took.
	// Promote pins each pair's incoming record, then the member's, on each claim.
	want := []string{
		"promote_subject:create:bName", "promote_subject:create:aName",
		"promote_subject:create:bAlt", "promote_subject:create:aAlt",
		"delete_observation:delete:aName",
	}
	for name, claimID := range map[string][]byte{"A": h.ca, "B": h.cb} {
		if got := h.exhibitHistory(t, claimID); strings.Join(got, " ") != strings.Join(want, " ") {
			t.Fatalf("claim %s history\n got %v\nwant %v", name, got, want)
		}
	}
	h.cacheMatchesRebuild(t)
}

func TestPinnedMemberDeleteEndToEnd(t *testing.T) {
	h := newPinnedHandle(t)

	// B's own Observations block deleting B; each goes first, off both claims.
	for _, id := range [][]byte{h.bName, h.bAlt} {
		report := mustImpact(t, h.c, catalogmodel.KindObservation, id)
		if !report.Allowed || len(report.Cascades) != 1 || report.Cascades[0].Listed[0].Ref != h.entityRef {
			t.Fatalf("%+v", report)
		}
		if _, err := writes.Call(h.c, writes.Op{Action: "delete_observation", UserID: h.userID}, func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
			changes, err := observations.Delete(tx, h.userID, id)
			return struct{}{}, changes, err
		}); err != nil {
			t.Fatal(err)
		}
	}
	if got := h.pinned(t, h.ca); got != "aName,aAlt" {
		t.Fatalf("A's claim after B's records went: %s", got)
	}

	// B is empty; its confirm names the handle it leaves, and its claim still
	// pins A's records (the backfill), which go with it, audited.
	report := mustImpact(t, h.c, catalogmodel.KindSubject, h.b.ID)
	if !report.Allowed {
		t.Fatalf("%+v", report)
	}
	assertLeaves(t, report, h.entityRef)
	if _, err := writes.Call(h.c, writes.Op{Action: "delete_subject", UserID: h.userID}, func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
		changes, err := subjects.Delete(tx, h.userID, h.b.ID)
		return struct{}{}, changes, err
	}); err != nil {
		t.Fatal(err)
	}
	if got := h.pinned(t, h.cb); got != "" {
		t.Fatalf("B's claim pins left: %s", got)
	}
	if got := h.pinned(t, h.ca); got != "aName,aAlt" {
		t.Fatalf("A's claim lost its own pins: %s", got)
	}
	if hist := h.exhibitHistory(t, h.cb); strings.Join(hist[len(hist)-2:], " ") !=
		"delete_subject:delete:aName delete_subject:delete:aAlt" {
		t.Fatalf("B's claim history %v", hist)
	}
	members, err := identityclaims.AcceptedMembers(h.c, []byte(h.entityID))
	if err != nil || len(members) != 1 || string(members[0].ID) != string(h.ca) {
		t.Fatalf("%v %+v", err, members)
	}
	h.cacheMatchesRebuild(t)
}
