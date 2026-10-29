# How a person's records connect

**Status:** pulled into Spike 13 on 2026-10-29 as slice 7 (S13-26 – S13-29). Decided in conversation with the user, 2026-10-29, after Identity Claim pins were dropped ([`conclusion-layer-data-model.md`](../../conclusion-layer-data-model.md) §5).

## The problem

Genealogy's core work is linking records: showing that a mention in one record and a mention in another are the same person. The genealogical proof standard asks for those links to be correlated and their conflicts resolved. Pins tried to record that correlation fact by fact on every Identity Claim, which made Promote busy work and stored only agreements. With pins gone, nothing shows a researcher how well a person's records hold together, which record stands alone, or whether two groups of records may be two different people.

## Shape

The Person page shows the person's **Evidence graphs** and the **links** between them.

- **A node is an Evidence graph** that holds an accepted member of the person: one record's reading of one group of people (a census household, an obituary, a birth certificate). Not a Source: one census Source can hold sixty household graphs.
- **A link between two graphs compares the graphs, not one person** (user, 2026-10-29). Two graphs are linked by the Subjects they have in common: each pair of Subjects, one in each graph, filed on the same canonical entity. A census household and an obituary that share five people share five pairs. The link carries that count and, for each shared entity, how its two Subjects compare. Shared relatives are the family support; no Property or relationship is named in code.
- **Each shared pair is compared as Subject compare does** (user, 2026-10-29): its per-Property router (leads, pools, periods, lifespan) and value and property compare under it (matching's layers 1 – 3), as the starting point. Which Properties count follows the registry's subject section, so a kind's list is one registry entry.
- **Layered and driven by the registry, as matching is.** One layer compares two Subjects; the layer above compares two graphs from their shared pairs. Each is a pure package with its own section of `core/matchregistry` for its knobs. The layers read nothing of graph alignment's weights, band or accept bar.
- **Strength is read from what the layers return:** how many entities the graphs share, and for each shared pair what agrees and what conflicts. How it is expressed is open (below).
- **What the shape shows:** a graph that links to no other is a record with nothing tying it to the rest; two groups of graphs that do not connect may be two people folded into one; a link with a conflict is where a Reconciliation Claim or a rejected claim is due.

## What is stored

Nothing new. The links are computed on read from Observations, Identity Claims and the canonical graph, and kept current as records change. The researcher's judgment stays where it is: each Identity Claim's status and optional argument. Knowledge outside the records enters as a cited Source (a testimony), so the comparison can see it.

**Refused:** researcher-chosen links stored at Promote, and recording what led the researcher to each record. A research log (questions, searches with no result, what led where) is a separate idea, not proof.

## What it is not

- **Not Promote's score.** Graph alignment keeps ranking Promote's suggestions behind the scenes, unseen (working rule "The score ranks, unseen"). This comparison is new: it reads no alignment weight, band or accept bar.
- **Not the reconciler.** The reconciler concludes a value per Property; this compares records pairwise and keeps values that lost the vote.

## Scale

A person with 50 Evidence graphs has 1,225 graph pairs; each compares only the Subjects the two graphs share (a household's few people, an obituary's few dozen at most); most people have 5 – 20 graphs. A place or event with thousands of members is not on this page this spike; when Event and Place adopt it, the read bounds its pairs (for example, only graphs that share a fact or a neighbor).

## Open questions

1. ~~**Which Properties count.**~~ Answered 2026-10-29: start from Subject compare's subject section; the graph layer counts shared entities and names no Property. Whether a kind's list widens (occupation, military service: an obituary's "Robert, in the Air Force") is a registry entry later.
2. **How strength is expressed.** A count of agreeing facts by kind, named levels, or something else, and what makes a link strong enough to join two graphs into one group for the islands view. On the Person page every pair of the person's graphs shares at least the person, so a link always exists: a record stands apart, and two groups separate, only by this strength rule (a conflicting or too-thin link does not join).
3. ~~**Family support.**~~ Answered 2026-10-29: it is the link itself; two graphs are linked by every entity they share.
4. **Source quality and independence.** Whether credibility, transcription certainty, and two graphs from one Source or informant change a link's strength in the first version.

## Later

- Event and Place pages reuse the engine and view when they adopt sections.
- Promote could rank suggestions with this comparison instead of graph alignment's score, once it has proven itself here.
- The research log stays an idea.
