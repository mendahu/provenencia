// Package evrun is the test helper that runs an evidence-layer write through
// writes.Run. Production handlers call writes.Run themselves.
package evrun

import (
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/citations"
	"github.com/mendahu/provenencia/core/database/connect"
	"github.com/mendahu/provenencia/core/database/observations"
	"github.com/mendahu/provenencia/core/database/rowchange"
	"github.com/mendahu/provenencia/core/database/subjectpositions"
	"github.com/mendahu/provenencia/core/database/subjects"
	"github.com/mendahu/provenencia/core/writes"
)

func CreateSubject(c *database.Catalog, userID []byte, in subjects.CreateInput, placement *subjects.Placement) (subjects.Subject, error) {
	s, _, err := writes.Run(c, writes.Op{Action: "create_subject", UserID: userID},
		func(tx *database.Tx) (subjects.Subject, []rowchange.Change, error) {
			return subjects.Create(tx, userID, in, placement)
		})
	return s, err
}

func UpdateSubject(c *database.Catalog, userID, id []byte, label, description string) error {
	_, _, err := writes.Run(c, writes.Op{Action: "update_subject", UserID: userID},
		func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
			changes, err := subjects.Update(tx, userID, id, label, description)
			return struct{}{}, changes, err
		})
	return err
}

func DeleteSubject(c *database.Catalog, userID, id []byte) error {
	_, _, err := writes.Run(c, writes.Op{Action: "delete_subject", UserID: userID},
		func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
			changes, err := subjects.Delete(tx, userID, id)
			return struct{}{}, changes, err
		})
	return err
}

func CreateCitation(c *database.Catalog, userID []byte, in citations.CreateInput, obs []observations.Input) (citations.CreateResult, error) {
	res, _, err := writes.Run(c, writes.Op{Action: "create_citation_with_observations", UserID: userID},
		func(tx *database.Tx) (citations.CreateResult, []rowchange.Change, error) {
			return citations.CreateWithObservations(tx, userID, in, obs)
		})
	return res, err
}

func UpdateCitation(c *database.Catalog, userID, citationID []byte, in citations.CitationFieldsInput) (citations.Citation, error) {
	cit, _, err := writes.Run(c, writes.Op{Action: "update_citation", UserID: userID},
		func(tx *database.Tx) (citations.Citation, []rowchange.Change, error) {
			return citations.Update(tx, userID, citationID, in)
		})
	return cit, err
}

func DeleteCitation(c *database.Catalog, userID, id []byte) error {
	_, _, err := writes.Run(c, writes.Op{Action: "delete_citation", UserID: userID},
		func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
			changes, err := citations.Delete(tx, userID, id)
			return struct{}{}, changes, err
		})
	return err
}

func UpdateObservation(c *database.Catalog, userID []byte, in observations.Input) (observations.Listed, error) {
	row, _, err := writes.Run(c, writes.Op{Action: "update_observation", UserID: userID},
		func(tx *database.Tx) (observations.Listed, []rowchange.Change, error) {
			return observations.Update(tx, userID, in)
		})
	return row, err
}

func DeleteObservation(c *database.Catalog, userID, id []byte) error {
	_, _, err := writes.Run(c, writes.Op{Action: "delete_observation", UserID: userID},
		func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
			changes, err := observations.Delete(tx, userID, id)
			return struct{}{}, changes, err
		})
	return err
}

func AddObservations(c *database.Catalog, userID, citationID []byte, inputs []observations.Input) ([]observations.Observation, error) {
	rows, _, err := writes.Run(c, writes.Op{Action: "add_observations", UserID: userID},
		func(tx *database.Tx) ([]observations.Observation, []rowchange.Change, error) {
			return observations.AddToCitation(tx, userID, citationID, inputs)
		})
	return rows, err
}

func CreateBridge(c *database.Catalog, userID []byte, in connect.CreateInput) (connect.Result, error) {
	res, _, err := writes.Run(c, writes.Op{Action: "create_cited_bridge", UserID: userID},
		func(tx *database.Tx) (connect.Result, []rowchange.Change, error) {
			return connect.CreateCitedBridge(tx, userID, in)
		})
	return res, err
}

func SetPosition(c *database.Catalog, subjectID []byte, gridX, gridY int64) (subjectpositions.Position, error) {
	p, _, err := writes.Run(c, writes.Op{},
		func(tx *database.Tx) (subjectpositions.Position, []rowchange.Change, error) {
			return subjectpositions.Set(tx, subjectID, gridX, gridY)
		})
	return p, err
}

func ClearPosition(c *database.Catalog, subjectID []byte) error {
	_, _, err := writes.Run(c, writes.Op{},
		func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
			changes, err := subjectpositions.Clear(tx, subjectID)
			return struct{}{}, changes, err
		})
	return err
}
