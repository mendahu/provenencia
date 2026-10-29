# How a person's records connect

**Status:** pulled into Spike 13 on 2026-10-29 as slice 7 (S13-26 – S13-29). Decided in conversation with the user, 2026-10-29, after Identity Claim pins were dropped ([`conclusion-layer-data-model.md`](../../conclusion-layer-data-model.md) §5).

## The problem

Genealogy's core work is linking records: showing that a mention in one record and a mention in another are the same person. The genealogical proof standard asks for those links to be correlated and their conflicts resolved. Pins tried to record that correlation fact by fact on every Identity Claim, which made Promote busy work and stored only agreements. With pins gone, nothing shows a researcher how well a person's records hold together, which record stands alone, or whether two groups of records may be two different people.

## Shape

The Person page shows the person's **Evidence graphs** and the **links** between them.

- **A node is an Evidence graph** that holds an accepted member of the person: one record's reading of one group of people (a census household, an obituary, a birth certificate). Not a Source: one census Source can hold sixty household graphs.
- **A link between two graphs** compares the person's two members in them: every Property both record, through value compare and property compare (matching's layers 1 and 2), and every relationship whose neighbors are filed on the same handle (the father on both records is PER-X). It lists what agrees and what conflicts, each with the facts behind it.
- **Strength is read from that list**, not from points: how many facts agree, of which kinds, and whether anything conflicts. How it is expressed is open (below).
- **What the shape shows:** a graph that links to no other is a record with nothing tying it to the rest; two groups of graphs that do not connect may be two people folded into one; a link with a conflict is where a Reconciliation Claim or a rejected claim is due.

## What is stored

Nothing new. The links are computed on read from Observations, Identity Claims and the canonical graph, and kept current as records change. The researcher's judgment stays where it is: each Identity Claim's status and optional argument. Knowledge outside the records enters as a cited Source (a testimony), so the comparison can see it.

**Refused:** researcher-chosen links stored at Promote, and recording what led the researcher to each record. A research log (questions, searches with no result, what led where) is a separate idea, not proof.

## What it is not

- **Not Promote's score.** Graph alignment keeps ranking Promote's suggestions behind the scenes, unseen (working rule "The score ranks, unseen"). This comparison is new: it reads no alignment weight, band or accept bar.
- **Not the reconciler.** The reconciler concludes a value per Property; this compares records pairwise and keeps values that lost the vote.

## Scale

A person with 50 Evidence graphs has 1,225 pairs, each comparing two Subjects; most people have 5 – 20. A place or event with thousands of members is not on this page this spike; when Event and Place adopt it, the read bounds its pairs (for example, only graphs that share a fact or a neighbor).

## Open questions

1. **Which Properties count.** Every Property both members record, or a list per Subject type. Matching's lead list (name, sex, date, toponym, event type) misses occupation and military service, which are often the only thread (an obituary's "Robert, in the Air Force").
2. **How strength is expressed.** A count of agreeing facts by kind, named levels, or something else, and what makes a link strong enough to join two graphs into one group for the islands view.
3. **Family support.** Whether a link also shows the other people the two graphs share (a household that matches as a family is stronger than one person matching).
4. **Source quality and independence.** Whether credibility, transcription certainty, and two graphs from one Source or informant change a link's strength in the first version.

## Later

- Event and Place pages reuse the engine and view when they adopt sections.
- Promote could rank suggestions with this comparison instead of graph alignment's score, once it has proven itself here.
- The research log stays an idea.
