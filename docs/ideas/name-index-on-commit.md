# Name index kept current on commit

**Status:** idea, 2026-10-26. Not scheduled. From the S11-39 ship loop (#567) and the slice 9 re-cut.

## The problem

Promote's name index (`promotealign`'s shortlist, built over every person, event and place) and every Source's ranked candidates are thrown away by any handle commit (`graphcache.OnCommit` clears the match index; the candidates memo is keyed on its generation). The next proposal rebuilds the index (about 70 ms on an 18,000-handle catalog) and ranks every Subject again (about 200 ms). On that catalog a proposal after a handle commit took about 0.54 s locally and 1.75 s on CI against S11-21's 1.5 s budget.

S11-50 moves the first build behind the open spinner. It does not help after a commit.

## A shape

- The commit updates the index in place: remove the committed handles' old keys, add their new ones, keep the generation stable for handles it did not touch.
- The candidates memo re-ranks only the Subjects whose shortlist keys meet a changed handle's old or new keys, and keeps the rest.
- Graph verify mode compares the kept index with a fresh build, as it does for nodes.

## Open questions

- Whether the shortlist index (`core/shortlist`) can remove a handle cheaply, or needs per-handle key lists.
- How a merge (two handles become one) reads: remove both, add the survivor.
