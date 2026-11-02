package handlers

import (
	"github.com/mendahu/provenencia/core/catalogsession"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/derivatives"
	"github.com/mendahu/provenencia/core/ingest"
)

// thumbnailBatch is how many thumbnails one session commits, and so how many
// temp files a call holds open at once.
const thumbnailBatch = 16

// thumbnailPass makes the thumbnails a response carries without decoding
// inside a catalog session. A handler builds its response in its first
// session, calling need for each File that shows a thumbnail; need fills the
// ones that exist and records a job for the rest. After that session the
// handler calls run: it decodes each job outside the session and commits in
// small sessions, filling the response as it goes. Every response that
// carries a thumbnail path goes through one pass.
type thumbnailPass struct {
	jobs   []*thumbnailJob
	byFile map[string]*thumbnailJob
}

// thumbnailFill sets the response field a thumbnail fills: the path, or skipped
// with an empty path when the File shows none.
type thumbnailFill func(relPath string, skipped bool)

type thumbnailJob struct {
	job   *derivatives.Job
	fills []thumbnailFill
}

func (j *thumbnailJob) fill(relPath string, skipped bool) {
	for _, f := range j.fills {
		f(relPath, skipped)
	}
}

// need runs inside a session. A File that already has its thumbnail, or never
// will, is filled at once; a File that needs one is recorded for run, once per
// File id, with every fill for it. fill may be nil.
func (p *thumbnailPass) need(c *database.Catalog, fileID []byte, fill thumbnailFill) error {
	if fill == nil {
		fill = func(string, bool) {}
	}
	if len(fileID) != 16 {
		fill("", true)
		return nil
	}
	if j, ok := p.byFile[string(fileID)]; ok {
		j.fills = append(j.fills, fill)
		return nil
	}
	res, job, err := derivatives.Plan(c, fileID, derivatives.ThumbnailSpec())
	if err != nil {
		if derivatives.Refused(err) {
			fill("", true)
			return nil
		}
		return err
	}
	if job == nil {
		return fillLink(c, res, fill)
	}
	if p.byFile == nil {
		p.byFile = map[string]*thumbnailJob{}
	}
	j := &thumbnailJob{job: job, fills: []thumbnailFill{fill}}
	p.byFile[string(fileID)] = j
	p.jobs = append(p.jobs, j)
	return nil
}

// fillLink fills from a finished Result: skipped, or the derived File's path.
func fillLink(c *database.Catalog, res derivatives.Result, fill thumbnailFill) error {
	if res.Skipped || len(res.Link.DerivedFileID) != 16 {
		fill("", true)
		return nil
	}
	rel, skipped, err := derivedRelPath(c, res.Link)
	if err != nil {
		return err
	}
	fill(rel, skipped)
	return nil
}

// run makes the recorded thumbnails, in batches of thumbnailBatch: each batch
// decodes its jobs outside the session (withStaged), then commits them in one
// session. A job whose decode or commit is refused fills as skipped. then, when
// not nil, runs inside the last batch's session, or in one session of its own
// when no job was recorded, so a write that needs the thumbnails shares their
// session. With no job and no then, run opens no session.
func (p *thumbnailPass) run(projectDir string, then func(*database.Catalog) error) error {
	if len(p.jobs) == 0 {
		if then == nil {
			return nil
		}
		return withProjectCatalog(projectDir, then)
	}
	dir, err := catalogsession.Key(projectDir)
	if err != nil {
		return err
	}
	for start := 0; start < len(p.jobs); start += thumbnailBatch {
		end := min(start+thumbnailBatch, len(p.jobs))
		batch := p.jobs[start:end]
		var staged []*thumbnailJob
		last := end == len(p.jobs)
		var stages ingest.Stages
		for _, j := range batch {
			s, err := j.job.Stage(dir)
			if err != nil {
				if derivatives.Refused(err) {
					j.fill("", true)
					continue
				}
				stages.Discard()
				return err
			}
			stages = append(stages, s)
			staged = append(staged, j)
		}
		if len(staged) == 0 && !(last && then != nil) {
			continue // every job was refused at the stage: nothing to commit
		}
		err := withStaged(projectDir, func(string) (ingest.Stages, error) {
			return stages, nil
		}, func(c *database.Catalog) error {
			for _, j := range staged {
				res, err := derivatives.Commit(c, j.job)
				if err != nil {
					if derivatives.Refused(err) {
						j.fill("", true)
						continue
					}
					return err
				}
				if err := fillLink(c, res, j.fill); err != nil {
					return err
				}
			}
			if last && then != nil {
				return then(c)
			}
			return nil
		})
		if err != nil {
			return err
		}
	}
	return nil
}
