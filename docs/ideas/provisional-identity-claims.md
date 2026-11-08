# Provisional members — Promote status and possible values

**Status:** idea, not scheduled. Promote status planned for Spike 13 as S13-30 (2026-10-29) and descoped on 2026-10-30 before any of it was built. Promote files every row Accepted today. Possible values raised while building the Person page (S9-16).

## What stays in place

The schema already holds provisional claims, and it stays as it is: `identity_claims.status` is `provisional`, `accepted` or `rejected` ([`conclusion-layer-data-model.md`](../conclusion-layer-data-model.md) §4, [`seeded-vocabulary.md`](../seeded-vocabulary.md)). The engine already treats a provisional claim as a non-member:

- bridge filing reads accepted claims only (`core/database/identityclaims/bridgefile.go`), so an end with only a provisional claim stays unfiled (conclusion §5.4);
- the auto-reconciler passes provisional members in and always eliminates them (reason `provisional`; [`conclusion-reconciliation.md`](../conclusion-reconciliation.md) §4);
- matching's profile leaves provisional values out ([`matching-model.md`](../matching-model.md));
- a merge keeps an accepted claim over a provisional one for the same pair (conclusion §2.3).

What is missing is a way to write one: no screen offers Provisional, and Promote writes `accepted` (`core/database/promote/batch.go`).

## The idea

Each promoted row on the Promote page gets a **status**, **Accepted** (the default) or **Provisional**. Rejected is not offered at Promote. A provisional claim files nothing else and only preselects the row the next time Promote opens on that graph; accepting it then files its connections.

### Model rules (as agreed 2026-10-29)

- **Provisional writes a provisional Identity Claim and nothing else.** The Subject still counts as not promoted: on its Evidence graph card, in the graph's status and progress, and among Promote's anchors. Its bridges stay unfiled, and the auto-reconciler ignores the claim, so the handle's values do not change.
- **It comes back preselected.** The next time Promote runs on that graph, the row comes back on the provisional handle, with status Provisional, and not Decided.
- **Accepting upgrades the same claim.** Accepting it turns the same `(subject, entity)` claim accepted, an audited update of that row, and files its waiting bridges in the same transaction.
- **One claim per Subject from Promote.** Choosing another handle or New, or promoting Provisional on another handle, replaces the Subject's earlier provisional claim from that graph. Skip leaves it as it is.
- **Done** still sends no pins, no confidence and no argument: status is the row's only claim field.

### Engine shape (from the S13-30 scope)

- **Apply:** `ApplyPromoteGraphAlignmentRow` gains `status` (`accepted` or `provisional`; empty reads as accepted; anything else is `promote.invalid`), and `promote.BatchRow` carries it. Filing runs on the status change as well as on create (`identityclaims.Create` / `AppendFiling`).
- **Reads:** a provisional-only Subject counts as unpromoted wherever "promoted" is read: the Evidence graph card's membership, the graph's status and progress counts, and Promote's anchors.
- **Proposal:** it returns a provisional Subject's record as the row's target with its status, so the page preselects it. After S13-24 this is one more field on `graphalign.RowStart` (`Status`) and one more branch in `Row.Start`.
- **Swift:** the row's status control, Accepted by default; Done sends the status; a provisional end is a connection's "waits for a named end" state; the proposal's provisional preselect maps onto the row.
- **Ordering with Spike 14:** S14-16 (landed) rewrote `ApplyPromoteGraphAlignment`'s handler and Swift mapping; this work carries the `status` field through the rewritten code.

**Check:** a row filed Provisional leaves its Subject unpromoted on the card and in the graph's counts, files none of its connections, and comes back preselected and Provisional on the next Promote. Accepting it then files its connections, and the record's auto-reconciled values are unchanged by the provisional claim.

### Design (board agreed 2026-10-29)

Drawn on the [Promote flow board](https://claude.ai/design/p/bc84685e-bbc3-4053-a5c9-0f5ac7a13ccd?via=share), frames 30, 30b, 30W and 31, from brief S13-D11 §4.1 (in git with #633). Those frames are marked descoped, not deleted.

| ID | Requirement |
| --- | --- |
| PS-1 | A working row whose target is a record shows a **status** control, **Accepted** or **Provisional**, starting on Accepted: a 140 pt small Select in the row's empty slot (PA-2 option (a)), lined up down the page like the target menu. |
| PS-2 | A row set to Skip, or still unset, shows no status control and keeps its row height. |
| PS-3 | Choosing Provisional marks the row Decided. Every connection subrow with that row as an end shows the waits state, naming it: "Waits for Robert Frickleton", spoken "…, who is provisional." Switching back to Accepted restores the subrows' states. |
| PS-4 | A row that comes back provisional opens on its record with status Provisional and is not Decided. Nothing else on the row marks it; its connections still wait. |
| PS-5 | The footer counts a Provisional row as promoted, and the board adds an optional "1 provisional" part to the summary. |
| PS-6 | VoiceOver reads the control as "Status for Mary Smith, Accepted" (or Provisional), and a waiting subrow as today. |

The slot PS-1 uses is still on every Promote row: empty on a working row, holding "Promoted from this Source earlier" on an Already promoted row (`macos/App/Features/Promote/PromoteView.swift`, the `freed:` slot). S14-22 (landed) left the slot as it was; if it is reclaimed first, the board's placement needs redrawing.

## Open questions

- **Can a New row be Provisional?** It would mint a record whose only claim is provisional. Recommendation: no, Provisional only on an existing record (frame 30b draws New without the control).
- **Who reads provisional claims?** The Person page and entity pages read none today. Showing a handle's candidates (a distinct stroke, badge or filter, as conclusion §4 asks) is part of making this useful; see also possible values below and [`matching-test-suite.md`](matching-test-suite.md) § Remember researcher decisions.
- **Rejected.** Not offered at Promote, and graph alignment reads no rejected claims, so a rule that stops suggesting a rejected record is not planned either.

## Possible values from provisional members

A provisional member's records are reasoning only: the auto-reconciler never lets them count ([`conclusion-reconciliation.md`](../conclusion-reconciliation.md) §4, Membership status). That is stricter than weak evidence, which is still shown when nothing stronger disagrees. So when the **only** record for a field comes from a provisional member, the page says *No birth date recorded*. The record shows under *Why* as *provisional member*, but the page itself looks empty, even though the researcher has a lead.

**Shape:** When nothing an accepted member recorded survives for a field, show the provisional member's value as a fallback, marked so it can't be mistaken for a conclusion: *possibly 14 May 1817*. *Why* explains it as usual (its Source, *provisional member*).

**Open questions:**

- **Where the rule lives.** Either the auto-reconciler gives such a value its own reason (say `possible`) and caches it, so lists, search and the detail read agree, or the page derives it from the cached `provisional` rows. The cache is the safer home: one rule, every reader.
- **What else a possible value touches.** Should it reach the Persons list, search, or only the detail page? **Not matching:** Spike 11's entity profile is strict and leaves provisional values out ([`comparison-layers.md`](../deployment-plan/archive/spike-11/comparison-layers.md#entities-compare-against-the-conclusion)).
- **Other set-aside evidence.** The same fallback could apply to a value denied by a stronger negative, or a weak one that stronger evidence disagrees with. Those were judged and lost, so they probably shouldn't come back; provisional records were never judged.
- **Wording and mark.** *possibly X* in italics with a distinct badge? It should read without colour, like the other states (S9-D5).
- **State.** Is it a fifth field state beside single / merged / mixed / empty, or an empty field with a hint?
