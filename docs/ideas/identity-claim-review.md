# Identity Claim review (weak-claim alert)

**Status:** parked for claim management (Spike 10). Rules live in [`conclusion-layer-data-model.md`](../conclusion-layer-data-model.md) §5.2; this note is the UI shape.

## Why

Deletes never refuse because of Conclusion pins (Spike 9, S9-02). Deleting a pinned Observation, or a member whose Observations other claims pinned, leaves those claims standing with **less evidence**. That is valid — a claim with an empty exhibit is allowed — but the researcher should find out and decide, instead of it going unnoticed.

## What counts as weak

- **Unpinned with siblings:** accepted, zero pins, and its handle has more than one accepted member. A handle's only (grounding) claim is not flagged.
- **Lost its comparison:** every pinned comparison Observation belonged to a member that has since left (§5.2).
- Later, maybe: a pinned Observation's value changed materially after it was confirmed.

All are queries over `identity_claims`, `identity_claim_evidence`, and members — no stored flag. The audit history of `identity_claim_evidence` (removals recorded under the claim's id) explains *why*: never pinned vs. evidence removed by which delete, when.

## Affordances (rough)

- **Badge** on the Person / Event / Place page and its list row when any member's claim is weak.
- **Member list** on the detail page marks the weak claim and shows what was removed ("lost OBS-… when it was deleted on …").
- **Review queue** (sidebar or page section): every weak claim across the project, newest first.
- **Actions** per claim: re-pin against a remaining member (reuses Promote's evidence sheet and graph alignment, S9-44), accept as-is (dismiss until something else changes), or reject the claim. The app never auto-rejects or evicts.
- **At delete time** the confirm already names the handles affected (S9-02); a link from that sentence to the review queue would close the loop.

## Open questions

- Is "accept as-is" a stored dismissal, or just the absence of an alert until the next change?
- Do provisional claims get reviewed too, or only accepted ones?
- Does the same alert cover Reconciliation Claims' evidence when they ship?
