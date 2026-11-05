# Place bridge card reads "Is part of of" with no endpoint

- **Date:** 2026-11-05
- **Type:** Bug
- **Where:** Evidence graph / place relationship card
- **Annoyance:** With a missing endpoint and no stored label, a place bridge reads "Is part of of · <ref>" or "Is succeeded by of · <ref>". It is ungrammatical, and it now says "is" where the sentence says "was".
- **Wanted:** A copy fix; the user has not decided the wording.

## Code notes

- `bridgeSummaryPlaceRelationshipTypeOnly` ("Is %@ of") and `wholeNameFallback` build the phrase for a product term.
- The `Sentence` rows added for place relationships leave `phrase` nil, so the registry does not override it.
