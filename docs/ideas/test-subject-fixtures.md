# Test Subject fixtures

**Status:** idea only — not roadmapped. The shared create fixture (`subjectstest.On` / `OfType` / `At`) shipped in S12-18; this is the part left over, found by the S12-18 plan review (M-4).

## One placement write

These test helpers create a Subject with `subjects.Create` and no placement, then run a second write with `subjectpositions.Set`: promotealign, canonicalgraph, the graphcache walk, the promote batch, autoreconciler, grapheventtitles, the conclusionheaders walks and search handles. `subjectstest.At` takes a `subjects.Placement`, so each can create and place the Subject in one write; moving them onto it is the work left.
