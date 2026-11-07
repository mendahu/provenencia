# How a person's records connect

**Status:** requirements, captured 2026-10-29 (user); the first version decided 2026-11-07 (user, [below](#first-version-2026-11-07)). Spike 13 slice 7 (S13-26 – S13-29) runs after the Promote cleanup, the pin removal and the Person page.

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

## First version (2026-11-07)

Start as simple as possible, build it, and look at many people in the dogfood projects before designing anything richer. The decisions are working rules in the [map](deployment-plan.md#slice-7--how-a-persons-records-connect): "Slice 7 starts with a simple count" and "The view is a circular link diagram". In short:

- **This is not Promote's job.** Promote's suggestions guess which Subjects might match; here the researcher has already joined both members to one entity, and the question is how much evidence the two records share. So the derived lifespan, a suggestion helper, is not counted, and nothing reads graph alignment's scoring.
- **One line per shared person.** Two graphs get a line for each other person both name. Events, places and the page's own person are not lines.
- **A line counts facts.** Each Property both members record that agrees or partly agrees (a name with at least one word in common), and each link both record to the same canonical entity (a relationship, or the same birth, death or marriage in the same role), counts once.
- **Two weights.** One fact is thin (weak); two or more is thick (strong). Conflicts are ignored. The number of arcs is the data, so two graphs sharing three people get three arcs.
- **A circular link diagram.** Graphs sit around a circle, labelled by the graph ref the Sources list shows; hovering a graph on the circle or in the list highlights it in both, so the researcher knows which graph to go and look at.
- **No explanation yet.** Nothing about a line is shown beyond its arc and weight: no label, tooltip or detail. Requirement 1's "every result explains itself" waits for a later version.

What this version leaves out is in [Open](#open), still open for a later version.

## Scale

A person with 50 Evidence graphs has 1,225 graph pairs; each compares only the Subjects the two graphs share (a household's few people, an obituary's few dozen at most). Most people have 5 – 20 graphs. A place or event with thousands of member records needs bounding when Event and Place pages adopt this.

## Refused

Researcher-chosen links stored at Promote, and recording what led the researcher to each record. A research log (questions, searches with no result, what led where) is a separate idea, not proof.

## Open

Left out of the first version, to revisit after dogfooding it:

- Showing what a line counted, and its conflicts.
- Whether conflicts and contested values should weaken or break a line.
- A richer score than two weights, and what should make a connection strong enough to join two graphs into one group (today any line does).
- Whether Source quality (credibility, transcription certainty) and independence (two graphs from one Source or informant) weigh in.
- Whether shared events and places, not only persons, should draw lines.

## Later

- Event and Place pages reuse the scoring and the view when they adopt sections.
- Promote could rank suggestions with this scoring instead of graph alignment's, once it has proven itself.
- The research log stays an idea.
