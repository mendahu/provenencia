# Source structure: Copy Source and multiple Evidence graphs

Design for Spike 12. PR order and status are in the [map](deployment-plan.md). Captured from dogfooding on real Sources (2026-10-10) and settled with the product owner (2026-10-18).

## Problem

The schema assumes one flat Source with one Evidence graph. Two real cases strain that.

1. **One Source holds many unrelated pieces of evidence.** A nationwide census, a local history, a city directory or an encyclopedia mentions many families that have nothing to do with each other beyond the book. On one graph they never connect, and the graph keeps growing. A grandmother's obituary alone produced over 100 Subjects and over 200 Observations.
2. **Sources sit inside other Sources.** An obituary is in an issue of the Medicine Hat News. While entering obituaries it was unclear whether the Source type is "obituary" or "newspaper", and the publication's details had to be typed again for every obituary.

## The two tools

| Situation | Tool | Why |
| --- | --- | --- |
| The part can be cited on its own: its own author, title or date (an obituary, a signed encyclopedia article, an anthology chapter, a FamilySearch baptism record) | **Its own Source**, started with **Copy Source** | It gets its own credibility, and its Source type says what it is (obituary). The publication, issue, page, provider, collection, film and URL are metadata copied from a sibling. |
| The part is a region of one work (a census household, a directory entry, a family in a local history) | **Another Evidence graph** on the same Source | Same document, same credibility. Each family gets a small isolated graph that is promoted on its own. |

A graph can also be used however the researcher likes: a subgraph to keep things tidy, or a try-out graph that is never promoted. A graph with nothing promoted needs no special state.

## Copy Source

On the Source page, and from a Sources list row's context menu, **Copy…** opens a form:

- **Title** — a text field, prefilled with `Copy of <title>`.
- **Source type** — a dropdown, preselected to the original's type. It is required, so it is not a checkbox.
- **Checkboxes, all on:** description; metadata (values and the field layout); credibility (grade and argument); and **one checkbox per Artifact**.
- Not offered: notes. Never copied: graphs, Subjects, Citations, Observations.

Saving creates an ordinary new Source (new `SRC-` ref, audited create) and opens its page. Its own "Default" graph comes from Source creation. Each checked Artifact becomes a new Artifact row on the copy with the same label, description and file (`files` is unique by checksum, so nothing is stored twice). Whether the copy keeps the original's cover (`sources.primary_artifact_id`) is an open question for S12-01.

## Multiple Evidence graphs

### Data model

```sql
CREATE TABLE evidence_graphs (
	id          BLOB PRIMARY KEY,
	ref         TEXT UNIQUE NOT NULL,          -- GPH-XXXXX
	source_id   BLOB NOT NULL REFERENCES sources(id) ON DELETE CASCADE,
	name        TEXT NOT NULL,                 -- "Default" when minted with the Source
	description TEXT,

	UNIQUE (id, source_id)
) STRICT;

CREATE TABLE evidence_graph_artifacts (
	graph_id    BLOB NOT NULL REFERENCES evidence_graphs(id) ON DELETE CASCADE,
	artifact_id BLOB NOT NULL REFERENCES artifacts(id) ON DELETE CASCADE,

	PRIMARY KEY (graph_id, artifact_id)
) STRICT;

-- subjects gains graph_id; (graph_id, source_id) references evidence_graphs(id, source_id)
-- so a Subject's graph is always on its home Source.
```

The exact `ON DELETE` classes and `catalogmodel` buckets are set by the schema PR through [`add-catalog-migration`](../../../.agents/skills/add-catalog-migration/SKILL.md) and [`add-catalog-model`](../../../.agents/skills/add-catalog-model/SKILL.md): a graph is part of its Source's record (it goes when the Source goes, and a Source can only be deleted with no Artifacts and no Subjects, so its graphs are empty by then); a link row is part of both ends.

`subjects.source_id` stays. Code that is about the *document* (the auto-reconciler's agreement count, credibility, audit scope, delete impact) keeps reading it. Code that is about *the graph* reads `graph_id`.

### Rules

- **Every Source has at least one graph.** Creating a Source creates one named "Default". The last graph can't be deleted.
- **A graph has a `GPH-` ref, a name and an optional description.** Name and description are editable. Graphs are listed alphabetically by name.
- **A graph with Subjects can't be deleted.** An empty graph that is not the last one can.
- **Graphs are fully isolated.** A Subject belongs to exactly one graph and never moves. An Observation's Subject, the Subject it points at (relationships, participations), and its Citation's Artifact must all be in, or linked to, that one graph. Go refuses anything else at write time.
- **Artifacts and graphs are many-to-many.**
  - A graph created by the researcher needs at least one linked Artifact. Its form starts with every Artifact checked.
  - The Default graph may have none until the Source's first Artifact, which links to it automatically. "Add graph" is unavailable until the Source has an Artifact.
  - A new Artifact starts linked to every graph.
  - While a Source has one graph, linking is automatic and the checkboxes are hidden.
  - Unlinking is refused when that graph's Citations use the Artifact (the message lists them), and when it is the graph's last link.
  - Deleting an Artifact keeps today's rule (refused while Citations use it); it may leave a graph with no Artifacts.
- **The Citation composer offers only the graph's linked Artifacts.**
- **Promote is per graph.** A Promote layer is one graph's Subjects.
- **Agreement still counts per Source.** Two graphs of one census agreeing on a value are one vote for the auto-reconciler.
- **Credibility stays per Source.**

### Where graphs show

- **Source page:** a new Graphs section lists the Source's graphs, each with its name, ref, description and its own Open button. Create, edit (name, description, linked Artifacts) and delete live there. The header's jump to the Evidence graph goes; this section replaces it.
- **Source page, Artifacts section:** the add dialog and each Artifact card's edit show one checkbox per graph, when the Source has two or more.
- **Sources list:** every row expands like an accordion to list its graphs, each with its own progress counts and an Open link. A graph opens when it has at least one linked Artifact, as the single graph button does today.
- **Navigation:** a graph is a place. Its location carries the graph id, so Back / Forward and restore return to that exact graph, its Citation composer, and its Promote page. Breadcrumbs read Source › graph name.
- **Omnibar:** graphs are a searchable kind, by name, description and `GPH-` ref; a hit opens the graph.

### Migration

One `Default` graph per existing Source. Every Subject moves into its Source's Default graph, and every Artifact is linked to it. Positions are keyed by Subject, so the canvas looks the same. Existing projects look unchanged.

## Not this spike

These were weighed and parked. Their reasoning is in [`ideas/source-structure-later.md`](../../ideas/source-structure-later.md).

- Nested Sources (`parent_source_id`, metadata inheritance).
- Derivation links (`derived_from`: an index or abstract of an original) and the original / derivative / authored form.
- Named Source templates. Copy Source covers the case for now.
- Per-graph credibility.
- Layered citation output (Evidence Explained style "…; citing …").
- Canvas focus views (one connected component at a time, filter by Citation): [`ideas/graph-focus-views.md`](../../ideas/graph-focus-views.md).
- `source` Subjects, `mentions` and `remark`: [`ideas/source-to-source-relationships.md`](../../ideas/source-to-source-relationships.md). None exist yet; graph isolation will forbid cross-graph subject values when that idea is picked up, so it has to be re-thought then.
