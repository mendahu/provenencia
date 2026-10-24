# Test Subject fixtures

**Status:** idea only — not roadmapped. The shared create fixture (`subjectstest.On` / `OfType`) shipped in S12-18; this is the part left over, found by the S12-18 plan review (M-4).

## One placement write

Several of those helpers call `subjects.Create` with no placement and then run a second write with `subjectpositions.Set`. `subjects.Create` takes a `*subjects.Placement`; the shared fixture can take one and the second write goes.
