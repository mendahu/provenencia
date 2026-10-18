# Find another record in the Promote target menu

**Status:** idea only — not roadmapped. Found in S11-05 (D-16).

## What happens

Promote's target menu lists the records `Rank` offers. Since matching moved onto value compare, a record whose first given name or surname differs from the Subject's is `Different`, so `Rank` offers no candidate. The row opens New (reason `no_match`) with no alternatives; before, it opened Skip (reason `weak`) with alternatives. A researcher who knows the Subject and a differently named record are the same person (a nickname not in the aliases, a recorded variant) has no way to link them from Promote.

## Suggested shape

- A "Find another record…" entry at the end of the target menu that searches the catalog's records and lets the researcher pick one `Rank` did not offer.
- The pick is a decision like any other row choice: held as given, never rescored.

## Open questions

- Whether the search reuses the omnibar's record search or a smaller picker.
- Whether a picked record that the comparison calls `Different` shows a warning.
