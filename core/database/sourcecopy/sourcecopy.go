// Package sourcecopy copies a Source in one write: a new Source with a chosen
// title and type, and, each only when chosen, the original's description,
// published date (as a date row of the copy's own), metadata with its layout,
// credibility and any subset of its Artifacts. It composes the owner packages
// inside the caller's writes.Run transaction and never calls writes.Run
// itself.
//
// Notes, Subjects, Citations, Observations, graphs and the cover are never
// copied.
package sourcecopy

import (
	"database/sql"
	"errors"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/artifacts"
	"github.com/mendahu/provenencia/core/database/rowchange"
	"github.com/mendahu/provenencia/core/database/sourcecredibility"
	"github.com/mendahu/provenencia/core/database/sourcemetadata"
	"github.com/mendahu/provenencia/core/database/sources"
	"github.com/mendahu/provenencia/core/datevalue"
	"github.com/mendahu/provenencia/core/vocabkey"
)

// Draft is everything a copy can carry from a Source. Every SourceID is nil
// until CreateFromDraft sets it to the new Source.
type Draft struct {
	Source      sources.CreateInput
	Date        *datevalue.Value // nil when the Source is undated; its ID is always nil
	Metadata    sourcemetadata.Draft
	Credibility *sourcecredibility.UpsertInput // nil when the Source has none
	Artifacts   []DraftArtifact                // the original's ref order
}

// DraftArtifact is one Artifact to create; FromID is the Artifact it was read from.
type DraftArtifact struct {
	FromID []byte
	In     artifacts.CreateInput
}

// CopyInput is what the researcher chose on the Copy form.
type CopyInput struct {
	SourceID     []byte
	Title        string
	SourceTypeID []byte
	Description  bool
	Date         bool
	Metadata     bool
	Credibility  bool
	ArtifactIDs  [][]byte
}

// ReadDraft reads everything a copy can carry from a Source on tx. A missing
// Source is sources.ErrNotFound.
func ReadDraft(tx *sql.Tx, sourceID []byte) (Draft, error) {
	src, err := sources.GetTx(tx, sourceID)
	if err != nil {
		return Draft{}, err
	}
	d := Draft{Source: sources.CreateInput{
		SourceTypeID: src.SourceTypeID,
		Title:        src.Title,
		Description:  src.Description,
	}}

	if src.Date != nil {
		date := *src.Date
		date.ID = nil
		d.Date = &date
	}

	if d.Metadata, err = sourcemetadata.ReadDraftTx(tx, sourcemetadata.Owner{Level: vocabkey.LevelSource, ID: sourceID}); err != nil {
		return Draft{}, err
	}

	cred, err := sourcecredibility.GetBySourceTx(tx, sourceID)
	switch {
	case err == nil:
		d.Credibility = &sourcecredibility.UpsertInput{
			CredibilityGradeID: cred.CredibilityGradeID,
			Argument:           cred.Argument,
		}
	case !errors.Is(err, sql.ErrNoRows):
		return Draft{}, err
	}

	arts, err := artifacts.ListBySourceTx(tx, sourceID)
	if err != nil {
		return Draft{}, err
	}
	for _, a := range arts {
		d.Artifacts = append(d.Artifacts, DraftArtifact{
			FromID: a.ID,
			In:     artifacts.CreateInput{FileID: a.FileID, Label: a.Label, Description: a.Description},
		})
	}
	return d, nil
}

// CreateFromDraft writes the draft as a new Source and returns it with every
// change, in write order: Source, its date, metadata values, layout rows,
// credibility, then Artifacts. The returned Source carries its date. The new
// Source keeps the default cover. The caller records the changes.
func CreateFromDraft(tx *database.Tx, userID []byte, d Draft) (sources.Source, []rowchange.Change, error) {
	src, changes, err := sources.Create(tx, userID, d.Source)
	if err != nil {
		return sources.Source{}, nil, err
	}

	if d.Date != nil {
		var more []rowchange.Change
		src, more, err = sources.SetDate(tx, userID, src.ID, d.Date)
		if err != nil {
			return sources.Source{}, nil, err
		}
		changes = append(changes, more...)
	}

	more, err := sourcemetadata.CreateAll(tx, userID, sourcemetadata.Owner{Level: vocabkey.LevelSource, ID: src.ID}, d.Metadata)
	if err != nil {
		return sources.Source{}, nil, err
	}
	changes = append(changes, more...)

	if d.Credibility != nil {
		in := *d.Credibility
		in.SourceID = src.ID
		_, more, err := sourcecredibility.Upsert(tx, userID, in)
		if err != nil {
			return sources.Source{}, nil, err
		}
		changes = append(changes, more...)
	}

	for _, a := range d.Artifacts {
		in := a.In
		in.SourceID = src.ID
		_, more, err := artifacts.Create(tx, userID, in)
		if err != nil {
			return sources.Source{}, nil, err
		}
		changes = append(changes, more...)
	}
	return src, changes, nil
}

// Copy reads the original, keeps only the parts in, and creates the copy.
// Title and type always come from in; sources.Create trims and validates them.
// A missing original, or an Artifact id that is not among the original's or is
// repeated, is sources.ErrInvalid before anything is written. The caller
// records the changes.
func Copy(tx *database.Tx, userID []byte, in CopyInput) (sources.Source, []rowchange.Change, error) {
	if tx == nil {
		return sources.Source{}, nil, sources.ErrInvalid
	}
	if err := database.RequireUserID(userID, sources.ErrInvalid); err != nil {
		return sources.Source{}, nil, err
	}
	d, err := ReadDraft(tx.Tx, in.SourceID)
	if err != nil {
		return sources.Source{}, nil, err
	}
	if d, err = apply(d, in); err != nil {
		return sources.Source{}, nil, err
	}
	return CreateFromDraft(tx, userID, d)
}

// apply narrows the draft to what in chose.
func apply(d Draft, in CopyInput) (Draft, error) {
	d.Source.Title = in.Title
	d.Source.SourceTypeID = in.SourceTypeID
	if !in.Description {
		d.Source.Description = ""
	}
	if !in.Date {
		d.Date = nil
	}
	if !in.Metadata {
		d.Metadata = sourcemetadata.Draft{}
	}
	if !in.Credibility {
		d.Credibility = nil
	}

	chosen := make(map[string]struct{}, len(in.ArtifactIDs))
	for _, id := range in.ArtifactIDs {
		if _, dup := chosen[string(id)]; dup {
			return Draft{}, sources.ErrInvalid
		}
		chosen[string(id)] = struct{}{}
	}
	var kept []DraftArtifact
	for _, a := range d.Artifacts {
		if _, ok := chosen[string(a.FromID)]; ok {
			kept = append(kept, a)
			delete(chosen, string(a.FromID))
		}
	}
	if len(chosen) != 0 {
		return Draft{}, sources.ErrInvalid
	}
	d.Artifacts = kept
	return d, nil
}
