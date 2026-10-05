# Possible values from provisional members

**Status:** idea, not scheduled. Raised while building the Person page (S9-16).

## The gap

A provisional member's records are reasoning only: the auto-reconciler never lets them count ([`conclusion-reconciliation.md`](../conclusion-reconciliation.md) §4, Membership status). That is stricter than weak evidence, which is still shown when nothing stronger disagrees. So when the **only** record for a field comes from a provisional member, the page says *No birth date recorded*. The record shows under *Why* as *provisional member*, but the page itself looks empty, even though the researcher has a lead.

## The idea

When nothing an accepted member recorded survives for a field, show the provisional member's value as a fallback, marked so it can't be mistaken for a conclusion: *possibly 14 May 1817*. *Why* explains it as usual (its Source, *provisional member*), so the researcher can see what the hint rests on and decide whether to accept the member.

## Open questions

- **Where the rule lives.** Either the auto-reconciler gives such a value its own reason (say `possible`) and caches it, so lists, search and the detail read agree, or the page derives it from the cached `provisional` rows. The cache is the safer home: one rule, every reader.
- **What else a possible value touches.** Should it reach the Persons list, search, matching, or only the detail page? Showing it in matching could pull a wrong Person together.
- **Other set-aside evidence.** The same fallback could apply to a value denied by a stronger negative, or a weak one that stronger evidence disagrees with. Those were judged and lost, so they probably shouldn't come back; provisional records were never judged.
- **Wording and mark.** *possibly X* in italics with a distinct badge? It should read without colour, like the other states (S9-D5).
- **State.** Is it a fifth field state beside single / merged / mixed / empty, or an empty field with a hint?
