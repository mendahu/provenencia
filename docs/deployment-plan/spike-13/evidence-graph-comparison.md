# How a person's records connect

**Status:** requirements, captured 2026-10-29 (user). Spike 13 slice 7 (S13-26 – S13-29) is blocked on the scoring design (open question 11 in the [map](deployment-plan.md#open-questions)) and runs after the Promote cleanup, the pin removal and the Person page. The implementation is deliberately not decided yet.

## The problem

Genealogy's core work is linking records: showing that a mention in one record and a mention in another are the same person, event or place. The genealogical proof standard (GPS) asks for those links to be correlated and their conflicts resolved. Identity Claim pins tried to record that fact by fact and were dropped ([`conclusion-layer-data-model.md`](../../conclusion-layer-data-model.md) §5). Nothing yet shows a researcher how well a handle's records hold together, which record stands alone, or whether two groups of records may be two different people.

The work splits into two concerns, built and reviewed separately: **scoring** a connection between two records, and **visualizing** the connections.

## Requirement 1 — Score the connection between two Evidence graphs

A function takes two Evidence graphs and says how strongly they connect, in a way the GPS can stand behind: every result explains itself in terms a researcher can check (what agrees, what conflicts, on which records).

- **The unit is an Evidence graph**, one record's reading of one group of people, not a Source: one census Source can hold sixty household graphs.
- **Quantity of shared Subjects.** Two graphs connect through the Subjects they have in common: pairs of Subjects, one in each graph, filed on the same canonical entity. Primary entities (people, events, places) and relationship entities (participations, locations, relationships, place relationships) both count.
- **Strength of each shared Subject.** A shared Subject is not binary. A pair that agrees on a single first name is weaker than one that agrees on a name, a gender and a religion.
- **Conflicts and contested values.** A pair that disagrees on a fact, or whose values are contested across the handle's records, weakens or breaks the connection, and the result says where.
- **Derived structure.** The same fact recorded in different shapes must compare across: one record gives two people a spouse relationship, the other gives them a shared marriage event (the implied links matching already derives, [`matching-model.md`](../../matching-model.md)).
- **The page's own entity.** On a handle's page every pair of its graphs shares that handle, so the connection between two graphs is read from what they share other than it, while the handle's own pair stays visible as the thing being proved.
- **Constraints:** pure, layered and driven by the match registry like matching, so its knobs are tuned in one place; Subject compare is the starting point to draw on; it reads none of graph alignment's weights, band or accept bar (Promote keeps ranking with those, unseen).
- **Computed, never stored.** Nothing is added to the catalog. Each Identity Claim keeps its status; it carries no argument (2026-10-29). Knowledge outside the records enters as a cited Source.

## Requirement 2 — Visualize the connections

Given Requirement 1's results for a handle, show its Evidence graphs and their connections so a researcher can see weak links, a record that stands apart, and groups of records that do not connect. The form is open: a node-to-node web, a circle with arcs between its records (a chord diagram), or another. The view depends only on Requirement 1's output, not on how it is computed. Its brief is S13-D10.

## Scale

A person with 50 Evidence graphs has 1,225 graph pairs; each compares only the Subjects the two graphs share (a household's few people, an obituary's few dozen at most). Most people have 5 – 20 graphs. A place or event with thousands of member records needs bounding when Event and Place pages adopt this.

## Refused

Researcher-chosen links stored at Promote, and recording what led the researcher to each record. A research log (questions, searches with no result, what led where) is a separate idea, not proof.

## Open

- How the score is built and expressed: named levels per pair, a numeric score explained in the view, or something else; and what makes a connection strong enough to join two graphs into one group.
- Whether Source quality (credibility, transcription certainty) and independence (two graphs from one Source or informant) weigh in.
- The visualization's form.

## Later

- Event and Place pages reuse the scoring and the view when they adopt sections.
- Promote could rank suggestions with this scoring instead of graph alignment's, once it has proven itself.
- The research log stays an idea.
